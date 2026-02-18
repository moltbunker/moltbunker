// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerRegistry.sol";

contract BunkerRegistryTest is Test {
    BunkerToken public token;
    BunkerRegistry public registry;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public alice = makeAddr("alice");
    address public bob = makeAddr("bob");

    uint256 public constant REG_FEE = 10_000 * 1e18; // 10K BUNKER

    function setUp() public {
        token = new BunkerToken(owner);
        registry = new BunkerRegistry(address(token), treasury, REG_FEE, owner);

        // Mint tokens for testing
        vm.startPrank(owner);
        token.mint(alice, 1_000_000 * 1e18);
        token.mint(bob, 1_000_000 * 1e18);
        vm.stopPrank();

        // Approve registry for spending
        vm.prank(alice);
        token.approve(address(registry), type(uint256).max);
        vm.prank(bob);
        token.approve(address(registry), type(uint256).max);
    }

    // -----------------------------------------------------------------------
    //  1. Deployment Tests
    // -----------------------------------------------------------------------

    function test_Deployment_Version() public view {
        assertEq(registry.VERSION(), "1.0.0");
    }

    function test_Deployment_TokenAddress() public view {
        assertEq(address(registry.bunkerToken()), address(token));
    }

    function test_Deployment_Treasury() public view {
        assertEq(registry.treasury(), treasury);
    }

    function test_Deployment_RegistrationFee() public view {
        assertEq(registry.registrationFee(), REG_FEE);
    }

    function test_Deployment_OwnerIsCorrect() public view {
        assertEq(registry.owner(), owner);
    }

    function test_Deployment_RevertZeroToken() public {
        vm.expectRevert(BunkerRegistry.InvalidAddress.selector);
        new BunkerRegistry(address(0), treasury, REG_FEE, owner);
    }

    function test_Deployment_RevertZeroTreasury() public {
        vm.expectRevert(BunkerRegistry.InvalidAddress.selector);
        new BunkerRegistry(address(token), address(0), REG_FEE, owner);
    }

    // -----------------------------------------------------------------------
    //  2. Registration Tests
    // -----------------------------------------------------------------------

    function test_Register_Success() public {
        bytes32 depID = bytes32(uint256(1));

        vm.prank(alice);
        registry.register("my-app", depID);

        (address regOwner, bytes32 regDep, uint256 regAt) = registry.resolve("my-app");
        assertEq(regOwner, alice);
        assertEq(regDep, depID);
        assertGt(regAt, 0);
    }

    function test_Register_FeeDistribution() public {
        bytes32 depID = bytes32(uint256(1));
        uint256 aliceBefore = token.balanceOf(alice);
        uint256 treasuryBefore = token.balanceOf(treasury);
        uint256 supplyBefore = token.totalSupply();

        vm.prank(alice);
        registry.register("my-app", depID);

        // Alice pays full fee
        assertEq(token.balanceOf(alice), aliceBefore - REG_FEE);

        // Treasury gets 20%
        uint256 treasuryAmount = REG_FEE * 2000 / 10000;
        assertEq(token.balanceOf(treasury), treasuryBefore + treasuryAmount);

        // 80% burned (supply decreased)
        uint256 burnAmount = REG_FEE * 8000 / 10000;
        assertEq(token.totalSupply(), supplyBefore - burnAmount);
    }

    function test_Register_RevertDuplicateName() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameAlreadyRegistered.selector, "my-app"));
        registry.register("my-app", bytes32(uint256(2)));
    }

    function test_Register_RevertZeroDeploymentID() public {
        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.InvalidDeploymentID.selector);
        registry.register("my-app", bytes32(0));
    }

    function test_Register_RevertShortName() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidName.selector, "ab"));
        registry.register("ab", bytes32(uint256(1)));
    }

    function test_Register_RevertLongName() public {
        // 33 chars
        string memory longName = "abcdefghijklmnopqrstuvwxyz1234567";
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidName.selector, longName));
        registry.register(longName, bytes32(uint256(1)));
    }

    function test_Register_RevertLeadingHyphen() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidName.selector, "-my-app"));
        registry.register("-my-app", bytes32(uint256(1)));
    }

    function test_Register_RevertTrailingHyphen() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidName.selector, "my-app-"));
        registry.register("my-app-", bytes32(uint256(1)));
    }

    function test_Register_RevertUppercase() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidName.selector, "My-App"));
        registry.register("My-App", bytes32(uint256(1)));
    }

    function test_Register_RevertSpecialChars() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidName.selector, "my_app"));
        registry.register("my_app", bytes32(uint256(1)));
    }

    function test_Register_MinLengthName() public {
        vm.prank(alice);
        registry.register("abc", bytes32(uint256(1)));

        (address regOwner,,) = registry.resolve("abc");
        assertEq(regOwner, alice);
    }

    function test_Register_MaxLengthName() public {
        // Exactly 32 chars
        string memory maxName = "abcdefghijklmnopqrstuvwxyz123456";
        assertEq(bytes(maxName).length, 32);

        vm.prank(alice);
        registry.register(maxName, bytes32(uint256(1)));

        (address regOwner,,) = registry.resolve(maxName);
        assertEq(regOwner, alice);
    }

    function test_Register_WhenPaused_Reverts() public {
        vm.prank(owner);
        registry.pause();

        vm.prank(alice);
        vm.expectRevert();
        registry.register("my-app", bytes32(uint256(1)));
    }

    // -----------------------------------------------------------------------
    //  3. Release Tests
    // -----------------------------------------------------------------------

    function test_Release_Success() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);

        vm.prank(alice);
        registry.release("my-app");

        assertTrue(registry.isAvailable("my-app"));
    }

    function test_Release_RevertNotOwner() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NotNameOwner.selector, "my-app", bob));
        registry.release("my-app");
    }

    function test_Release_RevertNotRegistered() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameNotRegistered.selector, "no-exist"));
        registry.release("no-exist");
    }

    function test_Release_AllowsReregistration() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);
        vm.prank(alice);
        registry.release("my-app");

        // Bob can now register the same name
        vm.prank(bob);
        registry.register("my-app", bytes32(uint256(2)));

        (address regOwner,,) = registry.resolve("my-app");
        assertEq(regOwner, bob);
    }

    // -----------------------------------------------------------------------
    //  4. Transfer Tests
    // -----------------------------------------------------------------------

    function test_Transfer_Success() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);

        vm.prank(alice);
        registry.transfer("my-app", bob);

        (address regOwner,,) = registry.resolve("my-app");
        assertEq(regOwner, bob);
    }

    function test_Transfer_UpdatesOwnershipLists() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);
        assertEq(registry.nameCount(alice), 1);

        vm.prank(alice);
        registry.transfer("my-app", bob);

        assertEq(registry.nameCount(alice), 0);
        assertEq(registry.nameCount(bob), 1);
    }

    function test_Transfer_RevertNotOwner() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NotNameOwner.selector, "my-app", bob));
        registry.transfer("my-app", bob);
    }

    function test_Transfer_RevertZeroAddress() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.InvalidAddress.selector);
        registry.transfer("my-app", address(0));
    }

    // -----------------------------------------------------------------------
    //  5. Update Deployment Tests
    // -----------------------------------------------------------------------

    function test_UpdateDeployment_Success() public {
        bytes32 depID1 = bytes32(uint256(1));
        bytes32 depID2 = bytes32(uint256(2));

        vm.prank(alice);
        registry.register("my-app", depID1);

        vm.prank(alice);
        registry.updateDeployment("my-app", depID2);

        (, bytes32 regDep,) = registry.resolve("my-app");
        assertEq(regDep, depID2);
    }

    function test_UpdateDeployment_RevertNotOwner() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NotNameOwner.selector, "my-app", bob));
        registry.updateDeployment("my-app", bytes32(uint256(2)));
    }

    function test_UpdateDeployment_RevertZeroID() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.InvalidDeploymentID.selector);
        registry.updateDeployment("my-app", bytes32(0));
    }

    // -----------------------------------------------------------------------
    //  6. View Functions
    // -----------------------------------------------------------------------

    function test_IsAvailable_True() public view {
        assertTrue(registry.isAvailable("my-app"));
    }

    function test_IsAvailable_False() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));
        assertFalse(registry.isAvailable("my-app"));
    }

    function test_NameCount_Multiple() public {
        vm.startPrank(alice);
        registry.register("app1", bytes32(uint256(1)));
        registry.register("app2", bytes32(uint256(2)));
        registry.register("app3", bytes32(uint256(3)));
        vm.stopPrank();

        assertEq(registry.nameCount(alice), 3);
    }

    // -----------------------------------------------------------------------
    //  7. Admin Functions
    // -----------------------------------------------------------------------

    function test_SetRegistrationFee() public {
        uint256 newFee = 50_000 * 1e18;

        vm.prank(owner);
        registry.setRegistrationFee(newFee);
        assertEq(registry.registrationFee(), newFee);
    }

    function test_SetRegistrationFee_RevertNotOwner() public {
        vm.prank(alice);
        vm.expectRevert();
        registry.setRegistrationFee(0);
    }

    function test_SetTreasury() public {
        address newTreasury = makeAddr("newTreasury");

        vm.prank(owner);
        registry.setTreasury(newTreasury);
        assertEq(registry.treasury(), newTreasury);
    }

    function test_SetTreasury_RevertZeroAddress() public {
        vm.prank(owner);
        vm.expectRevert(BunkerRegistry.InvalidAddress.selector);
        registry.setTreasury(address(0));
    }

    function test_ZeroFeeRegistration() public {
        // Set fee to zero
        vm.prank(owner);
        registry.setRegistrationFee(0);

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("free-app", bytes32(uint256(1)));

        // No tokens deducted
        assertEq(token.balanceOf(alice), aliceBefore);

        (address regOwner,,) = registry.resolve("free-app");
        assertEq(regOwner, alice);
    }

    function test_Pause_Unpause() public {
        vm.startPrank(owner);
        registry.pause();
        assertTrue(registry.paused());
        registry.unpause();
        assertFalse(registry.paused());
        vm.stopPrank();
    }

    // -----------------------------------------------------------------------
    //  8. Event Tests
    // -----------------------------------------------------------------------

    function test_EmitSubdomainRegistered() public {
        bytes32 depID = bytes32(uint256(1));

        vm.prank(alice);
        vm.expectEmit(false, true, false, true);
        emit BunkerRegistry.SubdomainRegistered("my-app", "my-app", alice, depID, REG_FEE);
        registry.register("my-app", depID);
    }

    function test_EmitSubdomainReleased() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);

        vm.prank(alice);
        vm.expectEmit(false, true, false, true);
        emit BunkerRegistry.SubdomainReleased("my-app", "my-app", alice);
        registry.release("my-app");
    }

    function test_EmitSubdomainTransferred() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);

        vm.prank(alice);
        vm.expectEmit(false, true, true, true);
        emit BunkerRegistry.SubdomainTransferred("my-app", "my-app", alice, bob);
        registry.transfer("my-app", bob);
    }

    function test_EmitSubdomainUpdated() public {
        bytes32 depID1 = bytes32(uint256(1));
        bytes32 depID2 = bytes32(uint256(2));
        vm.prank(alice);
        registry.register("my-app", depID1);

        vm.prank(alice);
        vm.expectEmit(false, false, false, true);
        emit BunkerRegistry.SubdomainUpdated("my-app", "my-app", depID1, depID2);
        registry.updateDeployment("my-app", depID2);
    }

    // -----------------------------------------------------------------------
    //  Pause guards on transfer and updateDeployment (C7 security fix)
    // -----------------------------------------------------------------------

    function test_Transfer_WhenPaused_Reverts() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("pausetest", depID);

        vm.prank(owner);
        registry.pause();

        vm.prank(alice);
        vm.expectRevert();
        registry.transfer("pausetest", bob);
    }

    function test_UpdateDeployment_WhenPaused_Reverts() public {
        bytes32 depID1 = bytes32(uint256(1));
        bytes32 depID2 = bytes32(uint256(2));
        vm.prank(alice);
        registry.register("pausetest2", depID1);

        vm.prank(owner);
        registry.pause();

        vm.prank(alice);
        vm.expectRevert();
        registry.updateDeployment("pausetest2", depID2);
    }

    function test_Transfer_AfterUnpause_Succeeds() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("unpausetest", depID);

        vm.startPrank(owner);
        registry.pause();
        registry.unpause();
        vm.stopPrank();

        vm.prank(alice);
        registry.transfer("unpausetest", bob);
        (address newOwner,,) = registry.resolve("unpausetest");
        assertEq(newOwner, bob);
    }

    // -----------------------------------------------------------------------
    //  I7: Max names per owner cap
    // -----------------------------------------------------------------------

    function test_Register_RevertTooManyNames() public {
        // Register MAX_NAMES_PER_OWNER names
        uint256 maxNames = registry.MAX_NAMES_PER_OWNER();
        vm.startPrank(alice);
        for (uint256 i = 0; i < maxNames; i++) {
            // Generate unique 3-char names: "a00", "a01", ..., "a99", "b00", ...
            bytes memory nameBytes = new bytes(3);
            nameBytes[0] = bytes1(uint8(0x61 + (i / 100))); // a-z
            nameBytes[1] = bytes1(uint8(0x30 + ((i / 10) % 10))); // 0-9
            nameBytes[2] = bytes1(uint8(0x30 + (i % 10))); // 0-9
            registry.register(string(nameBytes), bytes32(uint256(i + 1)));
        }
        vm.stopPrank();

        assertEq(registry.nameCount(alice), maxNames);

        // 101st name should revert
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(
            BunkerRegistry.TooManyNames.selector, alice, maxNames));
        registry.register("overflow", bytes32(uint256(999)));
    }

    // -----------------------------------------------------------------------
    //  I8: Minimum registration fee floor
    // -----------------------------------------------------------------------

    function test_SetRegistrationFee_RevertBelowMinimum() public {
        uint256 minFee = registry.MIN_REGISTRATION_FEE();

        vm.prank(owner);
        vm.expectRevert(abi.encodeWithSelector(
            BunkerRegistry.FeeBelowMinimum.selector, minFee - 1, minFee));
        registry.setRegistrationFee(minFee - 1);
    }

    function test_SetRegistrationFee_AllowZero() public {
        // Zero fee is explicitly allowed (free registrations)
        vm.prank(owner);
        registry.setRegistrationFee(0);
        assertEq(registry.registrationFee(), 0);
    }

    function test_SetRegistrationFee_AllowMinimum() public {
        uint256 minFee = registry.MIN_REGISTRATION_FEE();
        vm.prank(owner);
        registry.setRegistrationFee(minFee);
        assertEq(registry.registrationFee(), minFee);
    }
}
