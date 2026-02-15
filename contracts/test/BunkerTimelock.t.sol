// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEscrow.sol";
import "../src/BunkerPricing.sol";
import "../src/BunkerTimelock.sol";

contract BunkerTimelockTest is Test {
    BunkerToken public token;
    BunkerStaking public staking;
    BunkerEscrow public escrow;
    BunkerPricing public pricing;
    BunkerTimelock public timelock;

    address public admin = makeAddr("admin");
    address public proposer = makeAddr("proposer");
    address public executor = makeAddr("executor");
    address public guardian = makeAddr("guardian");
    address public treasury = makeAddr("treasury");
    address public newTreasury = makeAddr("newTreasury");
    address public attacker = makeAddr("attacker");

    uint256 constant MIN_DELAY = 24 hours;
    bytes32 constant ZERO_PREDECESSOR = bytes32(0);

    function setUp() public {
        // Deploy timelock
        address[] memory proposers = new address[](1);
        proposers[0] = proposer;
        address[] memory executors = new address[](1);
        executors[0] = executor;

        vm.startPrank(admin);

        timelock = new BunkerTimelock(
            MIN_DELAY,
            proposers,
            executors,
            admin,
            guardian
        );

        // Deploy protocol contracts with the timelock as owner
        token = new BunkerToken(admin);
        staking = new BunkerStaking(address(token), treasury, address(timelock));
        escrow = new BunkerEscrow(address(token), treasury, address(timelock));
        pricing = new BunkerPricing(address(timelock));

        vm.stopPrank();
    }

    // ================================================================
    //  1. CONSTRUCTOR & SETUP
    // ================================================================

    function test_constructor_setsMinDelay() public view {
        assertEq(timelock.getMinDelay(), MIN_DELAY);
    }

    function test_constructor_setsVersion() public view {
        assertEq(
            keccak256(bytes(timelock.VERSION())),
            keccak256(bytes("1.0.0"))
        );
    }

    function test_constructor_grantsProposerRole() public view {
        assertTrue(timelock.hasRole(timelock.PROPOSER_ROLE(), proposer));
    }

    function test_constructor_grantsCancellerRoleToProposer() public view {
        assertTrue(timelock.hasRole(timelock.CANCELLER_ROLE(), proposer));
    }

    function test_constructor_grantsExecutorRole() public view {
        assertTrue(timelock.hasRole(timelock.EXECUTOR_ROLE(), executor));
    }

    function test_constructor_grantsGuardianRole() public view {
        assertTrue(timelock.hasRole(timelock.GUARDIAN_ROLE(), guardian));
    }

    function test_constructor_grantsAdminRole() public view {
        assertTrue(timelock.hasRole(timelock.DEFAULT_ADMIN_ROLE(), admin));
    }

    function test_constructor_revertsBelowMinDelayFloor() public {
        address[] memory proposers = new address[](1);
        proposers[0] = proposer;
        address[] memory executors = new address[](1);
        executors[0] = executor;

        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerTimelock.DelayBelowFloor.selector,
                12 hours,
                MIN_DELAY
            )
        );
        new BunkerTimelock(12 hours, proposers, executors, admin, guardian);
    }

    function test_constructor_zeroGuardianAllowed() public {
        address[] memory proposers = new address[](1);
        proposers[0] = proposer;
        address[] memory executors = new address[](1);
        executors[0] = executor;

        // Should not revert with zero guardian
        BunkerTimelock tl = new BunkerTimelock(
            MIN_DELAY, proposers, executors, admin, address(0)
        );
        assertFalse(tl.hasRole(tl.GUARDIAN_ROLE(), address(0)));
    }

    // ================================================================
    //  2. SCHEDULE + EXECUTE AFTER DELAY (setTreasury on Staking)
    // ================================================================

    function test_scheduleAndExecute_setTreasuryOnStaking() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("setTreasury-staking-1");
        bytes32 opId = timelock.hashOperation(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );

        // Schedule
        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        // Verify operation is pending
        assertTrue(timelock.isOperationPending(opId));
        assertFalse(timelock.isOperationReady(opId));

        // Warp past delay
        vm.warp(block.timestamp + MIN_DELAY);

        // Now it should be ready
        assertTrue(timelock.isOperationReady(opId));

        // Execute
        vm.prank(executor);
        timelock.execute(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );

        // Verify the treasury was updated
        assertEq(staking.treasury(), newTreasury);
        assertTrue(timelock.isOperationDone(opId));
    }

    // ================================================================
    //  3. EXECUTION BEFORE DELAY REVERTS
    // ================================================================

    function test_executeBeforeDelay_reverts() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("setTreasury-early");

        // Schedule
        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        // Try to execute immediately (before delay)
        vm.prank(executor);
        vm.expectRevert(); // TimelockUnexpectedOperationState
        timelock.execute(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );
    }

    function test_executeHalfwayThroughDelay_reverts() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("setTreasury-halfway");

        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        // Warp halfway
        vm.warp(block.timestamp + MIN_DELAY / 2);

        vm.prank(executor);
        vm.expectRevert(); // TimelockUnexpectedOperationState
        timelock.execute(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );
    }

    // ================================================================
    //  4. CANCELLATION
    // ================================================================

    function test_cancel_pendingOperation() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("setTreasury-cancel");
        bytes32 opId = timelock.hashOperation(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );

        // Schedule
        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );
        assertTrue(timelock.isOperationPending(opId));

        // Cancel (proposer also has CANCELLER_ROLE by default)
        vm.prank(proposer);
        timelock.cancel(opId);

        // Verify cancelled (no longer an operation)
        assertFalse(timelock.isOperation(opId));
        assertFalse(timelock.isOperationPending(opId));
    }

    function test_cancel_readyOperation() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("setTreasury-cancel-ready");
        bytes32 opId = timelock.hashOperation(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );

        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        // Warp past delay so it becomes ready
        vm.warp(block.timestamp + MIN_DELAY);
        assertTrue(timelock.isOperationReady(opId));

        // Cancel the ready operation
        vm.prank(proposer);
        timelock.cancel(opId);

        assertFalse(timelock.isOperation(opId));
    }

    function test_executeAfterCancel_reverts() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("setTreasury-exec-after-cancel");
        bytes32 opId = timelock.hashOperation(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );

        // Schedule
        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        // Cancel
        vm.prank(proposer);
        timelock.cancel(opId);

        // Warp past delay
        vm.warp(block.timestamp + MIN_DELAY);

        // Try to execute -- should revert
        vm.prank(executor);
        vm.expectRevert(); // TimelockUnexpectedOperationState
        timelock.execute(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );
    }

    // ================================================================
    //  5. EMERGENCY PAUSE / UNPAUSE (no delay)
    // ================================================================

    function test_emergencyPause_staking() public {
        vm.prank(guardian);
        timelock.emergencyPause(address(staking));

        assertTrue(staking.paused());
    }

    function test_emergencyUnpause_staking() public {
        // First pause
        vm.prank(guardian);
        timelock.emergencyPause(address(staking));
        assertTrue(staking.paused());

        // Wait for M-10 pause cooldown (1 hour)
        vm.warp(block.timestamp + 1 hours + 1);

        // Then unpause
        vm.prank(guardian);
        timelock.emergencyUnpause(address(staking));
        assertFalse(staking.paused());
    }

    function test_emergencyPause_escrow() public {
        vm.prank(guardian);
        timelock.emergencyPause(address(escrow));

        assertTrue(escrow.paused());
    }

    function test_emergencyUnpause_escrow() public {
        vm.prank(guardian);
        timelock.emergencyPause(address(escrow));

        // Wait for M-10 pause cooldown (1 hour)
        vm.warp(block.timestamp + 1 hours + 1);

        vm.prank(guardian);
        timelock.emergencyUnpause(address(escrow));
        assertFalse(escrow.paused());
    }

    function test_emergencyPause_emitsEvent() public {
        vm.prank(guardian);
        vm.expectEmit(true, true, true, true);
        emit BunkerTimelock.EmergencyAction(
            guardian,
            address(staking),
            bytes4(keccak256("pause()"))
        );
        timelock.emergencyPause(address(staking));
    }

    function test_emergencyUnpause_emitsEvent() public {
        vm.prank(guardian);
        timelock.emergencyPause(address(staking));

        // Wait for M-10 pause cooldown (1 hour)
        vm.warp(block.timestamp + 1 hours + 1);

        vm.prank(guardian);
        vm.expectEmit(true, true, true, true);
        emit BunkerTimelock.EmergencyAction(
            guardian,
            address(staking),
            bytes4(keccak256("unpause()"))
        );
        timelock.emergencyUnpause(address(staking));
    }

    function test_emergencyPause_zeroAddressReverts() public {
        vm.prank(guardian);
        vm.expectRevert(BunkerTimelock.ZeroAddress.selector);
        timelock.emergencyPause(address(0));
    }

    function test_emergencyUnpause_zeroAddressReverts() public {
        vm.prank(guardian);
        vm.expectRevert(BunkerTimelock.ZeroAddress.selector);
        timelock.emergencyUnpause(address(0));
    }

    // ================================================================
    //  6. ACCESS CONTROL: NON-PROPOSER CANNOT SCHEDULE
    // ================================================================

    function test_schedule_nonProposerReverts() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("unauthorized");

        vm.prank(attacker);
        vm.expectRevert(); // AccessControlUnauthorizedAccount
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );
    }

    function test_schedule_executorCannotSchedule() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("exec-schedule");

        vm.prank(executor);
        vm.expectRevert(); // AccessControlUnauthorizedAccount
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );
    }

    function test_schedule_guardianCannotSchedule() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("guardian-schedule");

        vm.prank(guardian);
        vm.expectRevert(); // AccessControlUnauthorizedAccount
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );
    }

    // ================================================================
    //  7. ACCESS CONTROL: NON-EXECUTOR CANNOT EXECUTE
    // ================================================================

    function test_execute_nonExecutorReverts() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("non-executor");

        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        vm.warp(block.timestamp + MIN_DELAY);

        vm.prank(attacker);
        vm.expectRevert(); // AccessControlUnauthorizedAccount
        timelock.execute(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );
    }

    // ================================================================
    //  8. ACCESS CONTROL: NON-GUARDIAN CANNOT EMERGENCY PAUSE
    // ================================================================

    function test_emergencyPause_nonGuardianReverts() public {
        vm.prank(attacker);
        vm.expectRevert(); // AccessControlUnauthorizedAccount
        timelock.emergencyPause(address(staking));
    }

    function test_emergencyUnpause_nonGuardianReverts() public {
        vm.prank(attacker);
        vm.expectRevert(); // AccessControlUnauthorizedAccount
        timelock.emergencyUnpause(address(staking));
    }

    function test_emergencyPause_proposerCannotPause() public {
        vm.prank(proposer);
        vm.expectRevert(); // AccessControlUnauthorizedAccount
        timelock.emergencyPause(address(staking));
    }

    // ================================================================
    //  9. ACCESS CONTROL: NON-CANCELLER CANNOT CANCEL
    // ================================================================

    function test_cancel_nonCancellerReverts() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("cancel-unauth");
        bytes32 opId = timelock.hashOperation(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );

        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        vm.prank(attacker);
        vm.expectRevert(); // AccessControlUnauthorizedAccount
        timelock.cancel(opId);
    }

    // ================================================================
    //  10. SETPROTOCOLFEE VIA TIMELOCK
    // ================================================================

    function test_scheduleAndExecute_setProtocolFee() public {
        uint256 newFee = 300; // 3%
        bytes memory data = abi.encodeWithSignature(
            "setProtocolFee(uint256)", newFee
        );
        bytes32 salt = keccak256("setProtocolFee-1");

        vm.prank(proposer);
        timelock.schedule(
            address(escrow), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        vm.warp(block.timestamp + MIN_DELAY);

        vm.prank(executor);
        timelock.execute(
            address(escrow), 0, data, ZERO_PREDECESSOR, salt
        );

        assertEq(escrow.protocolFeeBps(), newFee);
    }

    // ================================================================
    //  11. SETTIERMIN STAKE VIA TIMELOCK
    // ================================================================

    function test_scheduleAndExecute_setTierMinStake() public {
        uint256 newMin = 1_000e18;
        bytes memory data = abi.encodeWithSignature(
            "setTierMinStake(uint8,uint256)",
            uint8(BunkerStaking.Tier.Starter),
            newMin
        );
        bytes32 salt = keccak256("setTierMinStake-1");

        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        vm.warp(block.timestamp + MIN_DELAY);

        vm.prank(executor);
        timelock.execute(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );

        (uint256 minStake,,,,) = staking.tierConfigs(BunkerStaking.Tier.Starter);
        assertEq(minStake, newMin);
    }

    // ================================================================
    //  12. SETPRICES VIA TIMELOCK (BunkerPricing)
    // ================================================================

    function test_scheduleAndExecute_setPrices() public {
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour:    1e18,    // 1.00 BUNKER
            memoryPerGBHour:   200000000000000000,  // 0.20 BUNKER
            storagePerGBMonth: 100000000000000000,  // 0.10 BUNKER
            networkPerGB:      50000000000000000,   // 0.05 BUNKER
            gpuBasicPerHour:   10e18,   // 10.00 BUNKER
            gpuPremiumPerHour: 30e18    // 30.00 BUNKER
        });

        bytes memory data = abi.encodeWithSignature(
            "setPrices((uint256,uint256,uint256,uint256,uint256,uint256))",
            newPrices
        );
        bytes32 salt = keccak256("setPrices-1");

        vm.prank(proposer);
        timelock.schedule(
            address(pricing), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        vm.warp(block.timestamp + MIN_DELAY);

        vm.prank(executor);
        timelock.execute(
            address(pricing), 0, data, ZERO_PREDECESSOR, salt
        );

        BunkerPricing.ResourcePrices memory fetched = pricing.getPrices();
        assertEq(fetched.cpuPerCoreHour, 1e18);
        assertEq(fetched.memoryPerGBHour, 200000000000000000);
        assertEq(fetched.gpuPremiumPerHour, 30e18);
    }

    // ================================================================
    //  13. SCHEDULE WITH DELAY BELOW MINIMUM REVERTS
    // ================================================================

    function test_schedule_belowMinDelayReverts() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("below-min-delay");

        vm.prank(proposer);
        vm.expectRevert(); // TimelockInsufficientDelay
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, 1 hours
        );
    }

    // ================================================================
    //  14. DUPLICATE SCHEDULE REVERTS
    // ================================================================

    function test_schedule_duplicateReverts() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("duplicate");

        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        // Same operation again
        vm.prank(proposer);
        vm.expectRevert(); // TimelockUnexpectedOperationState (already exists)
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );
    }

    // ================================================================
    //  15. GET OPERATION HASH HELPER
    // ================================================================

    function test_getOperationHash() public view {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("hash-test");

        bytes32 hash1 = timelock.getOperationHash(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );
        bytes32 hash2 = timelock.hashOperation(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );

        assertEq(hash1, hash2);
    }

    // ================================================================
    //  16. SETTREASURY ON ESCROW VIA TIMELOCK
    // ================================================================

    function test_scheduleAndExecute_setTreasuryOnEscrow() public {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("setTreasury-escrow-1");

        vm.prank(proposer);
        timelock.schedule(
            address(escrow), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        vm.warp(block.timestamp + MIN_DELAY);

        vm.prank(executor);
        timelock.execute(
            address(escrow), 0, data, ZERO_PREDECESSOR, salt
        );

        assertEq(escrow.treasury(), newTreasury);
    }

    // ================================================================
    //  17. DIRECT ADMIN CALLS REVERT (timelock is the owner)
    // ================================================================

    function test_directSetTreasury_revertsNotOwner() public {
        vm.prank(admin);
        vm.expectRevert(); // OwnableUnauthorizedAccount
        staking.setTreasury(newTreasury);
    }

    function test_directSetProtocolFee_revertsNotOwner() public {
        vm.prank(admin);
        vm.expectRevert(); // OwnableUnauthorizedAccount
        escrow.setProtocolFee(300);
    }

    function test_directSetPrices_revertsNotOwner() public {
        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        vm.prank(admin);
        vm.expectRevert(); // OwnableUnauthorizedAccount
        pricing.setPrices(p);
    }

    function test_directPause_revertsNotOwner() public {
        vm.prank(admin);
        vm.expectRevert(); // OwnableUnauthorizedAccount
        staking.pause();
    }

    // ================================================================
    //  18. LONGER DELAY WORKS
    // ================================================================

    function test_scheduleWithLongerDelay() public {
        uint256 longerDelay = 48 hours;
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        bytes32 salt = keccak256("longer-delay");
        bytes32 opId = timelock.hashOperation(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );

        vm.prank(proposer);
        timelock.schedule(
            address(staking), 0, data, ZERO_PREDECESSOR, salt, longerDelay
        );

        // Not ready at MIN_DELAY
        vm.warp(block.timestamp + MIN_DELAY);
        assertFalse(timelock.isOperationReady(opId));

        // Ready at longerDelay
        vm.warp(block.timestamp + longerDelay);
        assertTrue(timelock.isOperationReady(opId));

        vm.prank(executor);
        timelock.execute(
            address(staking), 0, data, ZERO_PREDECESSOR, salt
        );
        assertEq(staking.treasury(), newTreasury);
    }

    // ================================================================
    //  19. SETMULTIPLIERS VIA TIMELOCK (BunkerPricing)
    // ================================================================

    function test_scheduleAndExecute_setMultipliers() public {
        BunkerPricing.Multipliers memory newMults = BunkerPricing.Multipliers({
            redundancy: 20000,  // 2.00x
            tor:        15000,  // 1.50x
            premiumSLA: 20000,  // 2.00x
            spot:        3000   // 0.30x
        });

        bytes memory data = abi.encodeWithSignature(
            "setMultipliers((uint256,uint256,uint256,uint256))",
            newMults
        );
        bytes32 salt = keccak256("setMultipliers-1");

        vm.prank(proposer);
        timelock.schedule(
            address(pricing), 0, data, ZERO_PREDECESSOR, salt, MIN_DELAY
        );

        vm.warp(block.timestamp + MIN_DELAY);

        vm.prank(executor);
        timelock.execute(
            address(pricing), 0, data, ZERO_PREDECESSOR, salt
        );

        BunkerPricing.Multipliers memory fetched = pricing.getMultipliers();
        assertEq(fetched.redundancy, 20000);
        assertEq(fetched.tor, 15000);
        assertEq(fetched.premiumSLA, 20000);
        assertEq(fetched.spot, 3000);
    }

    // ================================================================
    //  20. EMERGENCY CALL FAILED REVERTS
    // ================================================================

    function test_emergencyPause_failsOnNonPausable() public {
        // Calling pause() on the token (which has no pause function) should fail
        vm.prank(guardian);
        vm.expectRevert(); // EmergencyCallFailed
        timelock.emergencyPause(address(token));
    }
}
