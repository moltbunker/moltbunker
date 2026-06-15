// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "@openzeppelin/contracts/governance/TimelockController.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEscrow.sol";
import "../src/BunkerPricing.sol";
import "../src/BunkerTimelock.sol";

/// @title GovernanceHandoffTest
/// @notice Verifies the ownership-handoff-to-Timelock pattern that DeployProtocol /
///         DeployTestnet automate: transferOwnership (Ownable2Step step 1) → schedule
///         the acceptOwnership batch → execute after minDelay (step 2). Also verifies
///         the deployer's DEFAULT_ADMIN_ROLE is dropped and an optional Safe co-proposer
///         can be granted PROPOSER_ROLE + CANCELLER_ROLE.
/// @dev This is a regression test for the script governance wiring. It replicates the
///      exact call sequence the scripts perform so a future change that breaks the
///      handoff (wrong selector, wrong batch shape, premature renounce) fails here.
contract GovernanceHandoffTest is Test {
    BunkerToken internal token;
    BunkerStaking internal staking;
    BunkerEscrow internal escrow;
    BunkerPricing internal pricing;
    BunkerTimelock internal timelock;

    address internal deployer = makeAddr("deployer");
    address internal treasury = makeAddr("treasury");
    address internal guardian = makeAddr("guardian");
    address internal safe = makeAddr("safeMultisig");

    bytes internal constant ACCEPT = abi.encodeWithSignature("acceptOwnership()");
    bytes32 internal constant SALT = keccak256("BUNKER_GOVERNANCE_HANDOFF_V1");

    function setUp() public {
        vm.startPrank(deployer);
        token = new BunkerToken(deployer);
        staking = new BunkerStaking(address(token), treasury, deployer);
        escrow = new BunkerEscrow(address(token), treasury, deployer);
        pricing = new BunkerPricing(deployer);

        address[] memory proposers = new address[](1);
        proposers[0] = deployer;
        address[] memory executors = new address[](1);
        executors[0] = deployer;
        timelock = new BunkerTimelock(24 hours, proposers, executors, deployer, guardian);
        vm.stopPrank();
    }

    /// @dev Replicate the script wiring: transfer ownership, schedule the accept batch,
    ///      optionally add a Safe, then renounce deployer admin.
    function _wire(bool withSafe) internal returns (bytes32 opId) {
        vm.startPrank(deployer);

        staking.transferOwnership(address(timelock));
        escrow.transferOwnership(address(timelock));
        pricing.transferOwnership(address(timelock));

        address[] memory targets = new address[](3);
        targets[0] = address(staking);
        targets[1] = address(escrow);
        targets[2] = address(pricing);
        uint256[] memory values = new uint256[](3);
        bytes[] memory payloads = new bytes[](3);
        payloads[0] = ACCEPT;
        payloads[1] = ACCEPT;
        payloads[2] = ACCEPT;

        opId = timelock.hashOperationBatch(targets, values, payloads, bytes32(0), SALT);
        timelock.scheduleBatch(targets, values, payloads, bytes32(0), SALT, timelock.getMinDelay());

        if (withSafe) {
            timelock.grantRole(timelock.PROPOSER_ROLE(), safe);
            timelock.grantRole(timelock.CANCELLER_ROLE(), safe);
        }

        timelock.renounceRole(timelock.DEFAULT_ADMIN_ROLE(), deployer);
        vm.stopPrank();
    }

    function _executeAccept() internal {
        address[] memory targets = new address[](3);
        targets[0] = address(staking);
        targets[1] = address(escrow);
        targets[2] = address(pricing);
        uint256[] memory values = new uint256[](3);
        bytes[] memory payloads = new bytes[](3);
        payloads[0] = ACCEPT;
        payloads[1] = ACCEPT;
        payloads[2] = ACCEPT;

        vm.prank(deployer); // deployer retains EXECUTOR_ROLE after admin renounce
        timelock.executeBatch(targets, values, payloads, bytes32(0), SALT);
    }

    // ──────────────────────────────────────────────

    function test_afterWiring_timelockIsPendingOwner() public {
        _wire(false);
        assertEq(staking.owner(), deployer, "owner not changed before accept");
        assertEq(staking.pendingOwner(), address(timelock));
        assertEq(escrow.pendingOwner(), address(timelock));
        assertEq(pricing.pendingOwner(), address(timelock));
    }

    function test_batchCannotExecuteBeforeDelay() public {
        _wire(false);
        // Not yet ready: TimelockController reverts on premature execute.
        vm.expectRevert();
        _executeAccept();
        // Ownership unchanged.
        assertEq(staking.owner(), deployer);
    }

    function test_fullHandoff_timelockOwnsAllThree() public {
        _wire(false);

        vm.warp(block.timestamp + 24 hours + 1);
        _executeAccept();

        assertEq(staking.owner(), address(timelock), "staking not owned by timelock");
        assertEq(escrow.owner(), address(timelock), "escrow not owned by timelock");
        assertEq(pricing.owner(), address(timelock), "pricing not owned by timelock");
    }

    function test_deployerAdminRoleRenounced() public {
        _wire(false);
        assertFalse(
            timelock.hasRole(timelock.DEFAULT_ADMIN_ROLE(), deployer),
            "deployer should not retain admin role"
        );
    }

    function test_deployerRetainsExecutorRoleForFinalisation() public {
        _wire(false);
        // EXECUTOR_ROLE is independent of DEFAULT_ADMIN_ROLE; deployer keeps it so the
        // accept batch can be finalised after the delay.
        assertTrue(timelock.hasRole(timelock.EXECUTOR_ROLE(), deployer));
    }

    function test_safeMultisigGrantedProposerAndCanceller() public {
        _wire(true);
        assertTrue(timelock.hasRole(timelock.PROPOSER_ROLE(), safe), "safe not proposer");
        assertTrue(timelock.hasRole(timelock.CANCELLER_ROLE(), safe), "safe not canceller");
    }

    function test_safeNotGrantedWhenAbsent() public {
        _wire(false);
        assertFalse(timelock.hasRole(timelock.PROPOSER_ROLE(), safe));
        assertFalse(timelock.hasRole(timelock.CANCELLER_ROLE(), safe));
    }

    /// @dev After handoff, owner-only admin functions can only be driven via a scheduled
    ///      Timelock operation — the deployer can no longer call them directly.
    function test_postHandoff_deployerCannotCallOwnerFunctions() public {
        _wire(false);
        vm.warp(block.timestamp + 24 hours + 1);
        _executeAccept();

        vm.prank(deployer);
        vm.expectRevert();
        escrow.setProtocolFee(100);
    }

    /// @dev After handoff, the Timelock can still drive owner functions through a
    ///      scheduled operation, proving governance is functional (not bricked).
    function test_postHandoff_timelockCanGovernViaSchedule() public {
        _wire(false);
        vm.warp(block.timestamp + 24 hours + 1);
        _executeAccept();

        uint256 newFee = 100;
        bytes memory data = abi.encodeWithSignature("setProtocolFee(uint256)", newFee);
        bytes32 salt = keccak256("set-fee");
        // Resolve the view call before pranking — an external call in the argument
        // list would otherwise consume the single-shot prank.
        uint256 delay = timelock.getMinDelay();

        vm.prank(deployer); // deployer retains PROPOSER_ROLE
        timelock.schedule(address(escrow), 0, data, bytes32(0), salt, delay);

        vm.warp(block.timestamp + delay + 1);
        vm.prank(deployer); // retains EXECUTOR_ROLE
        timelock.execute(address(escrow), 0, data, bytes32(0), salt);

        assertEq(escrow.protocolFeeBps(), newFee, "timelock-driven fee change failed");
    }
}
