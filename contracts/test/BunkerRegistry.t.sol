// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerRegistry.sol";

/// @dev Minimal mock staking contract for tier discount tests.
contract MockStaking {
    mapping(address => uint8) public tiers;

    function setTier(address user, uint8 tier) external {
        tiers[user] = tier;
    }

    function getTier(address provider) external view returns (uint8) {
        return tiers[provider];
    }
}

/// @dev Mock staking that always reverts (for graceful fallback test).
contract RevertingStaking {
    function getTier(address) external pure returns (uint8) {
        revert("staking down");
    }
}

contract BunkerRegistryTest is Test {
    BunkerToken public token;
    BunkerRegistry public registry;
    MockStaking public staking;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public alice = makeAddr("alice");
    address public bob = makeAddr("bob");
    address public charlie = makeAddr("charlie");

    uint256 public constant REG_FEE = 1_000_000 * 1e18; // 1M BUNKER
    uint256 public constant CHANGE_FEE = 10_000 * 1e18;  // 10K BUNKER

    function setUp() public {
        token = new BunkerToken(owner);
        staking = new MockStaking();
        registry = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(staking)
        );

        // Mint tokens for testing
        vm.startPrank(owner);
        token.mint(alice, 500_000_000 * 1e18);
        token.mint(bob, 500_000_000 * 1e18);
        token.mint(charlie, 500_000_000 * 1e18);
        vm.stopPrank();

        // Enable 3-char names for tests that need them
        vm.prank(owner);
        registry.setShortNamesEnabled(true);

        // Approve registry for spending
        vm.prank(alice);
        token.approve(address(registry), type(uint256).max);
        vm.prank(bob);
        token.approve(address(registry), type(uint256).max);
        vm.prank(charlie);
        token.approve(address(registry), type(uint256).max);
    }

    // =====================================================================
    //  1. Deployment Tests
    // =====================================================================

    function test_Deployment_Version() public view {
        assertEq(registry.VERSION(), "2.0.0");
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

    function test_Deployment_StakingContract() public view {
        assertEq(address(registry.stakingContract()), address(staking));
    }

    function test_Deployment_DefaultPeriods() public view {
        assertEq(registry.expirationPeriod(), 365 days);
        assertEq(registry.gracePeriod(), 30 days);
        assertEq(registry.reservationPeriod(), 48 hours);
        assertEq(registry.changeFee(), CHANGE_FEE);
        assertEq(registry.squattingGracePeriod(), 7 days);
    }

    function test_Deployment_RevertZeroToken() public {
        vm.expectRevert(BunkerRegistry.InvalidAddress.selector);
        new BunkerRegistry(address(0), treasury, REG_FEE, owner, address(staking));
    }

    function test_Deployment_RevertZeroTreasury() public {
        vm.expectRevert(BunkerRegistry.InvalidAddress.selector);
        new BunkerRegistry(address(token), address(0), REG_FEE, owner, address(staking));
    }

    function test_Deployment_AllowZeroStaking() public {
        // Zero staking address is allowed (no discounts)
        BunkerRegistry r = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(0)
        );
        assertEq(address(r.stakingContract()), address(0));
    }

    function test_Constructor_RevertFeeBelowMinimum() public {
        uint256 minFee = 1000 * 1e18;
        vm.expectRevert(abi.encodeWithSelector(
            BunkerRegistry.FeeBelowMinimum.selector, 1, minFee));
        new BunkerRegistry(address(token), treasury, 1, owner, address(staking));
    }

    function test_Constructor_AllowZeroFee() public {
        BunkerRegistry r = new BunkerRegistry(
            address(token), treasury, 0, owner, address(staking)
        );
        assertEq(r.registrationFee(), 0);
    }

    // =====================================================================
    //  2. Registration Tests
    // =====================================================================

    function test_Register_Success() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);

        (address regOwner, bytes32 regDep, uint256 regAt) = registry.resolve("my-app");
        assertEq(regOwner, alice);
        assertEq(regDep, depID);
        assertGt(regAt, 0);
    }

    function test_Register_SetsExpiration() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        registry.register("my-app", depID);

        bytes32 nameHash = keccak256(abi.encodePacked("my-app"));
        (,,, uint48 expiresAt,,) = registry.subdomains(nameHash);
        assertEq(expiresAt, uint48(block.timestamp + 365 days));
    }

    function test_Register_FeeDistribution() public {
        bytes32 depID = bytes32(uint256(1));
        uint256 aliceBefore = token.balanceOf(alice);
        uint256 treasuryBefore = token.balanceOf(treasury);
        uint256 supplyBefore = token.totalSupply();

        vm.prank(alice);
        registry.register("my-app", depID);

        assertEq(token.balanceOf(alice), aliceBefore - REG_FEE);
        uint256 treasuryAmount = REG_FEE * 2000 / 10000;
        assertEq(token.balanceOf(treasury), treasuryBefore + treasuryAmount);
        uint256 burnAmount = REG_FEE * 8000 / 10000;
        assertEq(token.totalSupply(), supplyBefore - burnAmount);
    }

    function test_Register_RevertDuplicateName() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameAlreadyRegistered.selector, "my-app"));
        registry.register("my-app", bytes32(uint256(2)));
    }

    function test_Register_RevertZeroDeploymentID() public {
        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.InvalidDeploymentID.selector);
        registry.register("my-app", bytes32(0));
    }

    function test_Register_RevertEmptyName() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidName.selector, ""));
        registry.register("", bytes32(uint256(1)));
    }

    function test_Register_RevertLongName() public {
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

    function test_Register_RevertConsecutiveHyphens() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidName.selector, "my--app"));
        registry.register("my--app", bytes32(uint256(1)));
    }

    function test_Register_RevertUppercase() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidName.selector, "My-App"));
        registry.register("My-App", bytes32(uint256(1)));
    }

    function test_Register_WhenPaused_Reverts() public {
        vm.prank(owner);
        registry.pause();

        vm.prank(alice);
        vm.expectRevert();
        registry.register("my-app", bytes32(uint256(1)));
    }

    function test_Register_1CharName() public {
        vm.prank(alice);
        registry.register("x", bytes32(uint256(1)));
        (address regOwner,,) = registry.resolve("x");
        assertEq(regOwner, alice);
    }

    function test_Register_2CharName() public {
        vm.prank(alice);
        registry.register("xy", bytes32(uint256(1)));
        (address regOwner,,) = registry.resolve("xy");
        assertEq(regOwner, alice);
    }

    function test_Register_MaxLengthName() public {
        string memory maxName = "abcdefghijklmnopqrstuvwxyz123456";
        assertEq(bytes(maxName).length, 32);
        vm.prank(alice);
        registry.register(maxName, bytes32(uint256(1)));
        (address regOwner,,) = registry.resolve(maxName);
        assertEq(regOwner, alice);
    }

    // =====================================================================
    //  3. Premium Pricing Tests
    // =====================================================================

    function test_PremiumPricing_1CharName_100x() public {
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("a", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE * 100);
    }

    function test_PremiumPricing_2CharName_50x() public {
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("ab", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE * 50);
    }

    function test_PremiumPricing_3CharName_10x() public {
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("abc", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE * 10);
    }

    function test_PremiumPricing_4CharName_5x() public {
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("abcd", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE * 5);
    }

    function test_PremiumPricing_5CharName_1x() public {
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("abcde", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE);
    }

    function test_CalculatePrice_View() public view {
        assertEq(registry.calculatePrice("a", alice), REG_FEE * 100);
        assertEq(registry.calculatePrice("ab", alice), REG_FEE * 50);
        assertEq(registry.calculatePrice("abc", alice), REG_FEE * 10);
        assertEq(registry.calculatePrice("abcd", alice), REG_FEE * 5);
        assertEq(registry.calculatePrice("abcde", alice), REG_FEE);
    }

    // =====================================================================
    //  4. Staking Tier Discount Tests
    // =====================================================================

    function test_StakingDiscount_Bronze_5Pct() public {
        staking.setTier(alice, 2); // Bronze
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        uint256 expected = REG_FEE - (REG_FEE * 500 / 10000); // 5% off
        assertEq(paid, expected);
    }

    function test_StakingDiscount_Silver_10Pct() public {
        staking.setTier(alice, 3); // Silver
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        uint256 expected = REG_FEE - (REG_FEE * 1000 / 10000); // 10% off
        assertEq(paid, expected);
    }

    function test_StakingDiscount_Gold_15Pct() public {
        staking.setTier(alice, 4); // Gold
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        uint256 expected = REG_FEE - (REG_FEE * 1500 / 10000); // 15% off
        assertEq(paid, expected);
    }

    function test_StakingDiscount_Platinum_20Pct() public {
        staking.setTier(alice, 5); // Platinum
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        uint256 expected = REG_FEE - (REG_FEE * 2000 / 10000); // 20% off
        assertEq(paid, expected);
    }

    function test_StakingDiscount_Starter_NoDiscount() public {
        staking.setTier(alice, 1); // Starter — no discount
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE);
    }

    function test_StakingDiscount_NoStakingContract() public {
        // Deploy registry without staking contract
        BunkerRegistry noStakeReg = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(0)
        );
        vm.prank(alice);
        token.approve(address(noStakeReg), type(uint256).max);

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        noStakeReg.register("my-app", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE); // full price
    }

    function test_StakingDiscount_RevertingStaking_GracefulFallback() public {
        RevertingStaking revStaking = new RevertingStaking();
        BunkerRegistry revReg = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(revStaking)
        );
        vm.prank(alice);
        token.approve(address(revReg), type(uint256).max);

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        revReg.register("my-app", bytes32(uint256(1)));
        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE); // full price, no revert
    }

    function test_StakingDiscount_CombinedWithPremium() public {
        staking.setTier(alice, 5); // Platinum 20%
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("abc", bytes32(uint256(1))); // 3-char = 10x
        uint256 paid = aliceBefore - token.balanceOf(alice);
        uint256 premiumPrice = REG_FEE * 10;
        uint256 expected = premiumPrice - (premiumPrice * 2000 / 10000);
        assertEq(paid, expected);
    }

    function test_CalculatePrice_WithStakingDiscount() public {
        staking.setTier(alice, 4); // Gold 15%
        uint256 price = registry.calculatePrice("my-app", alice);
        uint256 expected = REG_FEE - (REG_FEE * 1500 / 10000);
        assertEq(price, expected);
    }

    // =====================================================================
    //  5. Referral Discount Tests
    // =====================================================================

    function test_Referral_10PctDiscount() public {
        // Bob is the referrer
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.registerWithReferral("my-app", bytes32(uint256(1)), bob);
        uint256 paid = aliceBefore - token.balanceOf(alice);
        uint256 discounted = REG_FEE - (REG_FEE * 1000 / 10000); // 10% off
        assertEq(paid, discounted);
    }

    function test_Referral_ReferrerGets5Pct() public {
        uint256 bobBefore = token.balanceOf(bob);
        vm.prank(alice);
        registry.registerWithReferral("my-app", bytes32(uint256(1)), bob);
        uint256 bobGot = token.balanceOf(bob) - bobBefore;
        uint256 expectedReward = REG_FEE * 500 / 10000; // 5% of original fee
        // Reward comes from what the user pays (the discounted amount), so from treasury portion
        assertEq(bobGot, expectedReward);
    }

    function test_Referral_RevertSelfReferral() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidReferrer.selector, alice));
        registry.registerWithReferral("my-app", bytes32(uint256(1)), alice);
    }

    function test_Referral_RevertZeroReferrer() public {
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.InvalidReferrer.selector, address(0)));
        registry.registerWithReferral("my-app", bytes32(uint256(1)), address(0));
    }

    function test_Referral_CombinedWithStakingDiscount() public {
        staking.setTier(alice, 5); // Platinum 20%
        uint256 aliceBefore = token.balanceOf(alice);
        uint256 bobBefore = token.balanceOf(bob);

        vm.prank(alice);
        registry.registerWithReferral("my-app", bytes32(uint256(1)), bob);

        // Staking discount applied first
        uint256 afterStaking = REG_FEE - (REG_FEE * 2000 / 10000); // 800K
        // Then referral discount
        uint256 discounted = afterStaking - (afterStaking * 1000 / 10000);
        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, discounted);

        // Referrer reward is 5% of post-staking price
        uint256 referrerReward = afterStaking * 500 / 10000;
        assertEq(token.balanceOf(bob) - bobBefore, referrerReward);
    }

    // =====================================================================
    //  6. Expiration & Renewal Tests
    // =====================================================================

    function test_Expiration_NameExpiresAfterPeriod() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        assertFalse(registry.isExpired("my-app"));

        // Warp past expiration
        vm.warp(block.timestamp + 365 days + 1);
        assertTrue(registry.isExpired("my-app"));
    }

    function test_Expiration_CannotTransferExpired() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.warp(block.timestamp + 365 days + 1);

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameExpired.selector, "my-app"));
        registry.transfer("my-app", bob);
    }

    function test_Expiration_CannotUpdateExpired() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.warp(block.timestamp + 365 days + 1);

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameExpired.selector, "my-app"));
        registry.updateDeployment("my-app", bytes32(uint256(2)));
    }

    function test_Renew_ExtendsExpiration() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        bytes32 nameHash = keccak256(abi.encodePacked("my-app"));
        (,,, uint48 oldExpiry,,) = registry.subdomains(nameHash);

        // Renew before expiry
        vm.warp(block.timestamp + 300 days);
        vm.prank(alice);
        registry.renew("my-app");

        (,,, uint48 newExpiry,,) = registry.subdomains(nameHash);
        assertEq(newExpiry, oldExpiry + uint48(365 days));
    }

    function test_Renew_ChargesFee() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.renew("my-app");

        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE);
    }

    function test_Renew_DuringGracePeriod() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        // Warp into grace period
        vm.warp(block.timestamp + 365 days + 10 days);
        assertTrue(registry.isExpired("my-app"));
        assertTrue(registry.isInGracePeriod("my-app"));

        // Owner can still renew during grace
        vm.prank(alice);
        registry.renew("my-app");

        assertFalse(registry.isExpired("my-app"));
    }

    function test_Renew_RevertAfterGracePeriod() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        // Warp past grace period
        vm.warp(block.timestamp + 365 days + 30 days + 1);

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameExpired.selector, "my-app"));
        registry.renew("my-app");
    }

    function test_Renew_RevertNotOwner() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NotNameOwner.selector, "my-app", bob));
        registry.renew("my-app");
    }

    function test_Renew_PremiumNameChargesMultiplier() public {
        vm.prank(alice);
        registry.register("abc", bytes32(uint256(1)));

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.renew("abc");

        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE * 10); // 3-char premium
    }

    // =====================================================================
    //  7. Grace Period Tests
    // =====================================================================

    function test_GracePeriod_Expired_NotInGrace() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        // Before expiry
        assertFalse(registry.isInGracePeriod("my-app"));
    }

    function test_GracePeriod_InGrace() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));
        vm.warp(block.timestamp + 365 days + 15 days);
        assertTrue(registry.isInGracePeriod("my-app"));
    }

    function test_GracePeriod_PastGrace() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));
        vm.warp(block.timestamp + 365 days + 30 days + 1);
        assertFalse(registry.isInGracePeriod("my-app"));
    }

    function test_GracePeriod_OthersCannotRegisterDuringGrace() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.warp(block.timestamp + 365 days + 10 days);

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameAlreadyRegistered.selector, "my-app"));
        registry.register("my-app", bytes32(uint256(2)));
    }

    function test_GracePeriod_OthersCanRegisterAfterGrace() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.warp(block.timestamp + 365 days + 30 days + 1);

        vm.prank(bob);
        registry.register("my-app", bytes32(uint256(2)));

        (address regOwner,,) = registry.resolve("my-app");
        assertEq(regOwner, bob);
    }

    function test_IsAvailable_TrueAfterFullExpiry() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.warp(block.timestamp + 365 days + 30 days + 1);
        assertTrue(registry.isAvailable("my-app"));
    }

    function test_IsAvailable_FalseDuringGrace() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.warp(block.timestamp + 365 days + 10 days);
        assertFalse(registry.isAvailable("my-app"));
    }

    // =====================================================================
    //  8. Name Reservation Tests
    // =====================================================================

    function test_Reserve_Success() public {
        vm.prank(alice);
        registry.reserve("my-app");

        bytes32 nameHash = keccak256(abi.encodePacked("my-app"));
        (address recOwner,, uint48 regAt,, uint48 resUntil,) = registry.subdomains(nameHash);
        assertEq(recOwner, alice);
        assertGt(resUntil, 0);
        assertEq(resUntil, uint48(regAt + 48 hours));
    }

    function test_Reserve_ChargesFullFee() public {
        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.reserve("my-app");
        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, REG_FEE);
    }

    function test_Reserve_DeploymentIDIsZero() public {
        vm.prank(alice);
        registry.reserve("my-app");

        (, bytes32 depID,) = registry.resolve("my-app");
        assertEq(depID, bytes32(0));
    }

    function test_ClaimReservation_Success() public {
        vm.prank(alice);
        registry.reserve("my-app");

        bytes32 depID = bytes32(uint256(42));
        vm.prank(alice);
        registry.claimReservation("my-app", depID);

        (, bytes32 regDep,) = registry.resolve("my-app");
        assertEq(regDep, depID);

        // reservedUntil should be cleared
        bytes32 nameHash = keccak256(abi.encodePacked("my-app"));
        (,,,, uint48 resUntil,) = registry.subdomains(nameHash);
        assertEq(resUntil, 0);
    }

    function test_ClaimReservation_RevertExpired() public {
        vm.prank(alice);
        registry.reserve("my-app");

        vm.warp(block.timestamp + 48 hours + 1);

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.ReservationExpired.selector, "my-app"));
        registry.claimReservation("my-app", bytes32(uint256(42)));
    }

    function test_ClaimReservation_RevertNotOwner() public {
        vm.prank(alice);
        registry.reserve("my-app");

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NotReservationOwner.selector, "my-app", bob));
        registry.claimReservation("my-app", bytes32(uint256(42)));
    }

    function test_ClaimReservation_RevertZeroDeploymentID() public {
        vm.prank(alice);
        registry.reserve("my-app");

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.InvalidDeploymentID.selector);
        registry.claimReservation("my-app", bytes32(0));
    }

    function test_CancelReservation_Success() public {
        vm.prank(alice);
        registry.reserve("my-app");

        vm.prank(alice);
        registry.cancelReservation("my-app");

        assertTrue(registry.isAvailable("my-app"));
        assertEq(registry.nameCount(alice), 0);
    }

    function test_CancelReservation_RevertNotOwner() public {
        vm.prank(alice);
        registry.reserve("my-app");

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NotReservationOwner.selector, "my-app", bob));
        registry.cancelReservation("my-app");
    }

    function test_Reserve_ExpiredReservation_OtherCanRegister() public {
        vm.prank(alice);
        registry.reserve("my-app");

        vm.warp(block.timestamp + 48 hours + 1);

        // Bob can now register because reservation expired
        vm.prank(bob);
        registry.register("my-app", bytes32(uint256(99)));

        (address regOwner,,) = registry.resolve("my-app");
        assertEq(regOwner, bob);
    }

    function test_IsAvailable_TrueAfterReservationExpiry() public {
        vm.prank(alice);
        registry.reserve("my-app");

        vm.warp(block.timestamp + 48 hours + 1);
        assertTrue(registry.isAvailable("my-app"));
    }

    // =====================================================================
    //  9. Squatting Protection Tests
    // =====================================================================

    function test_ReclaimSquatted_NoDeploymentPastGrace() public {
        // Register with deploymentID but then update to 0 won't work since updateDeployment rejects 0
        // So we test via reservation that was claimed then... actually squatting protection works
        // for registrations where deploymentID stays bytes32(0), which can happen via reserve
        vm.prank(alice);
        registry.reserve("squat-me");

        // Warp past squatting grace (7 days) but still within reservation period wouldn't trigger
        // Need reservation to expire first, or not be a reservation at all
        // Actually, reservation has reservedUntil > 0, so case 2 won't trigger
        // Let's test case 1: expired reservation
        vm.warp(block.timestamp + 48 hours + 1);

        vm.prank(bob);
        registry.reclaimSquatted("squat-me");

        assertTrue(registry.isAvailable("squat-me"));
    }

    function test_ReclaimSquatted_RevertNameNotSquatted() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameNotSquatted.selector, "my-app"));
        registry.reclaimSquatted("my-app");
    }

    function test_ReclaimSquatted_RevertUnregistered() public {
        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameNotSquatted.selector, "no-exist"));
        registry.reclaimSquatted("no-exist");
    }

    function test_ReclaimSquatted_RevertWithinReservationPeriod() public {
        vm.prank(alice);
        registry.reserve("my-app");

        // Still within reservation period
        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameNotSquatted.selector, "my-app"));
        registry.reclaimSquatted("my-app");
    }

    // =====================================================================
    //  10. Bulk Operations Tests
    // =====================================================================

    function test_BulkRegister_Success() public {
        string[] memory names = new string[](3);
        names[0] = "app1";
        names[1] = "app2";
        names[2] = "app3";
        bytes32[] memory deps = new bytes32[](3);
        deps[0] = bytes32(uint256(1));
        deps[1] = bytes32(uint256(2));
        deps[2] = bytes32(uint256(3));

        vm.prank(alice);
        registry.bulkRegister(names, deps);

        assertEq(registry.nameCount(alice), 3);
        for (uint256 i = 0; i < 3; i++) {
            (address regOwner,,) = registry.resolve(names[i]);
            assertEq(regOwner, alice);
        }
    }

    function test_BulkRegister_RevertLengthMismatch() public {
        string[] memory names = new string[](2);
        names[0] = "app1";
        names[1] = "app2";
        bytes32[] memory deps = new bytes32[](1);
        deps[0] = bytes32(uint256(1));

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.ArrayLengthMismatch.selector);
        registry.bulkRegister(names, deps);
    }

    function test_BulkRegister_RevertTooLarge() public {
        string[] memory names = new string[](21);
        bytes32[] memory deps = new bytes32[](21);
        for (uint256 i = 0; i < 21; i++) {
            bytes memory nameBytes = new bytes(5);
            nameBytes[0] = "a";
            nameBytes[1] = "p";
            nameBytes[2] = "p";
            nameBytes[3] = bytes1(uint8(0x30 + (i / 10)));
            nameBytes[4] = bytes1(uint8(0x30 + (i % 10)));
            names[i] = string(nameBytes);
            deps[i] = bytes32(uint256(i + 1));
        }

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.ArrayTooLarge.selector, 21, 20));
        registry.bulkRegister(names, deps);
    }

    function test_BulkRenew_Success() public {
        vm.startPrank(alice);
        registry.register("app1", bytes32(uint256(1)));
        registry.register("app2", bytes32(uint256(2)));
        vm.stopPrank();

        string[] memory names = new string[](2);
        names[0] = "app1";
        names[1] = "app2";

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.bulkRenew(names);

        uint256 paid = aliceBefore - token.balanceOf(alice);
        // Both are 4-char names, so 5x each
        assertEq(paid, REG_FEE * 5 * 2);
    }

    function test_BulkRenew_RevertNotOwner() public {
        vm.prank(alice);
        registry.register("app1", bytes32(uint256(1)));

        string[] memory names = new string[](1);
        names[0] = "app1";

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NotNameOwner.selector, "app1", bob));
        registry.bulkRenew(names);
    }

    // =====================================================================
    //  11. Update Deployment (with change fee) Tests
    // =====================================================================

    function test_UpdateDeployment_ChargesChangeFee() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.updateDeployment("my-app", bytes32(uint256(2)));

        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, CHANGE_FEE);
    }

    function test_UpdateDeployment_UpdatesRecord() public {
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

    function test_UpdateDeployment_RevertExpired() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.warp(block.timestamp + 365 days + 1);

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameExpired.selector, "my-app"));
        registry.updateDeployment("my-app", bytes32(uint256(2)));
    }

    function test_UpdateDeployment_WhenPaused_Reverts() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(owner);
        registry.pause();

        vm.prank(alice);
        vm.expectRevert();
        registry.updateDeployment("my-app", bytes32(uint256(2)));
    }

    // =====================================================================
    //  12. Metadata Tests
    // =====================================================================

    function test_SetMetadata_Success() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(alice);
        registry.setMetadata("my-app", "My cool app", "https://example.com/avatar.png");

        bytes32 nameHash = keccak256(abi.encodePacked("my-app"));
        (string memory desc, string memory avatar) = registry.metadata(nameHash);
        assertEq(desc, "My cool app");
        assertEq(avatar, "https://example.com/avatar.png");
    }

    function test_SetMetadata_ChargesChangeFee() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.setMetadata("my-app", "desc", "url");

        uint256 paid = aliceBefore - token.balanceOf(alice);
        assertEq(paid, CHANGE_FEE);
    }

    function test_SetMetadata_RevertNotOwner() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NotNameOwner.selector, "my-app", bob));
        registry.setMetadata("my-app", "desc", "url");
    }

    function test_SetMetadata_RevertDescriptionTooLong() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        // 161 chars
        bytes memory longDesc = new bytes(161);
        for (uint256 i = 0; i < 161; i++) longDesc[i] = "a";

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(
            BunkerRegistry.MetadataDescriptionTooLong.selector, 161, 160));
        registry.setMetadata("my-app", string(longDesc), "url");
    }

    function test_SetMetadata_RevertAvatarTooLong() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        // 257 chars
        bytes memory longURL = new bytes(257);
        for (uint256 i = 0; i < 257; i++) longURL[i] = "a";

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(
            BunkerRegistry.MetadataAvatarURLTooLong.selector, 257, 256));
        registry.setMetadata("my-app", "desc", string(longURL));
    }

    function test_SetMetadata_RevertExpired() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.warp(block.timestamp + 365 days + 1);

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameExpired.selector, "my-app"));
        registry.setMetadata("my-app", "desc", "url");
    }

    function test_SetMetadata_ClearedOnRelease() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));
        vm.prank(alice);
        registry.setMetadata("my-app", "desc", "url");
        vm.prank(alice);
        registry.release("my-app");

        bytes32 nameHash = keccak256(abi.encodePacked("my-app"));
        (string memory desc, string memory avatar) = registry.metadata(nameHash);
        assertEq(desc, "");
        assertEq(avatar, "");
    }

    // =====================================================================
    //  13. Reverse Resolution Tests
    // =====================================================================

    function test_SetPrimaryName_Success() public {
        bytes32 depID = bytes32(uint256(42));
        vm.prank(alice);
        registry.register("my-app", depID);

        vm.prank(alice);
        registry.setPrimaryName("my-app");

        string memory result = registry.reverseResolve(depID);
        assertEq(result, "my-app");
    }

    function test_ReverseResolve_EmptyIfNotSet() public view {
        string memory result = registry.reverseResolve(bytes32(uint256(999)));
        assertEq(result, "");
    }

    function test_SetPrimaryName_RevertNotOwner() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(42)));

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NotNameOwner.selector, "my-app", bob));
        registry.setPrimaryName("my-app");
    }

    function test_SetPrimaryName_RevertNoDeployment() public {
        vm.prank(alice);
        registry.reserve("my-app");

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.InvalidDeploymentID.selector);
        registry.setPrimaryName("my-app");
    }

    function test_SetPrimaryName_RevertExpired() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(42)));

        vm.warp(block.timestamp + 365 days + 1);

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NameExpired.selector, "my-app"));
        registry.setPrimaryName("my-app");
    }

    function test_PrimaryName_ClearedOnRelease() public {
        bytes32 depID = bytes32(uint256(42));
        vm.prank(alice);
        registry.register("my-app", depID);
        vm.prank(alice);
        registry.setPrimaryName("my-app");

        vm.prank(alice);
        registry.release("my-app");

        string memory result = registry.reverseResolve(depID);
        assertEq(result, "");
    }

    function test_PrimaryName_ClearedOnUpdateDeployment() public {
        bytes32 depID1 = bytes32(uint256(42));
        bytes32 depID2 = bytes32(uint256(99));

        vm.prank(alice);
        registry.register("my-app", depID1);
        vm.prank(alice);
        registry.setPrimaryName("my-app");

        vm.prank(alice);
        registry.updateDeployment("my-app", depID2);

        // Old deployment should have no primary name
        string memory result = registry.reverseResolve(depID1);
        assertEq(result, "");
    }

    // =====================================================================
    //  14. Release Tests
    // =====================================================================

    function test_Release_Success() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(alice);
        registry.release("my-app");

        assertTrue(registry.isAvailable("my-app"));
    }

    function test_Release_RevertNotOwner() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

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
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));
        vm.prank(alice);
        registry.release("my-app");

        vm.prank(bob);
        registry.register("my-app", bytes32(uint256(2)));

        (address regOwner,,) = registry.resolve("my-app");
        assertEq(regOwner, bob);
    }

    function test_Release_WhenPaused_Reverts() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(owner);
        registry.pause();

        vm.prank(alice);
        vm.expectRevert();
        registry.release("my-app");
    }

    function test_Release_ClearsNameOf() public {
        bytes32 nameHash = keccak256(abi.encodePacked("cleanup"));
        vm.prank(alice);
        registry.register("cleanup", bytes32(uint256(1)));
        assertEq(registry.nameOf(nameHash), "cleanup");

        vm.prank(alice);
        registry.release("cleanup");
        assertEq(registry.nameOf(nameHash), "");
    }

    // =====================================================================
    //  15. Transfer Tests
    // =====================================================================

    function test_Transfer_Success() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(alice);
        registry.transfer("my-app", bob);

        (address regOwner,,) = registry.resolve("my-app");
        assertEq(regOwner, bob);
    }

    function test_Transfer_UpdatesOwnershipLists() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(alice);
        registry.transfer("my-app", bob);

        assertEq(registry.nameCount(alice), 0);
        assertEq(registry.nameCount(bob), 1);
    }

    function test_Transfer_RevertNotOwner() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(BunkerRegistry.NotNameOwner.selector, "my-app", bob));
        registry.transfer("my-app", charlie);
    }

    function test_Transfer_RevertZeroAddress() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.InvalidAddress.selector);
        registry.transfer("my-app", address(0));
    }

    function test_Transfer_RevertToSelf() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.CannotTransferToSelf.selector);
        registry.transfer("my-app", alice);
    }

    function test_Transfer_WhenPaused_Reverts() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(owner);
        registry.pause();

        vm.prank(alice);
        vm.expectRevert();
        registry.transfer("my-app", bob);
    }

    // =====================================================================
    //  16. Admin Functions Tests
    // =====================================================================

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

    function test_SetRegistrationFee_RevertBelowMinimum() public {
        uint256 minFee = registry.MIN_REGISTRATION_FEE();
        vm.prank(owner);
        vm.expectRevert(abi.encodeWithSelector(
            BunkerRegistry.FeeBelowMinimum.selector, minFee - 1, minFee));
        registry.setRegistrationFee(minFee - 1);
    }

    function test_SetRegistrationFee_AllowZero() public {
        vm.prank(owner);
        registry.setRegistrationFee(0);
        assertEq(registry.registrationFee(), 0);
    }

    function test_SetTreasury() public {
        address newTreasury = makeAddr("newTreasury");
        vm.prank(owner);
        registry.setTreasury(newTreasury);
        assertEq(registry.treasury(), newTreasury);
    }

    function test_SetTreasury_RevertZero() public {
        vm.prank(owner);
        vm.expectRevert(BunkerRegistry.InvalidAddress.selector);
        registry.setTreasury(address(0));
    }

    function test_SetChangeFee() public {
        vm.prank(owner);
        registry.setChangeFee(20_000 * 1e18);
        assertEq(registry.changeFee(), 20_000 * 1e18);
    }

    function test_SetChangeFee_AllowZero() public {
        vm.prank(owner);
        registry.setChangeFee(0);
        assertEq(registry.changeFee(), 0);
    }

    function test_SetExpirationPeriod() public {
        vm.prank(owner);
        registry.setExpirationPeriod(180 days);
        assertEq(registry.expirationPeriod(), 180 days);
    }

    function test_SetExpirationPeriod_RevertTooShort() public {
        vm.prank(owner);
        vm.expectRevert(BunkerRegistry.InvalidPeriod.selector);
        registry.setExpirationPeriod(29 days);
    }

    function test_SetExpirationPeriod_RevertTooLong() public {
        vm.prank(owner);
        vm.expectRevert(BunkerRegistry.InvalidPeriod.selector);
        registry.setExpirationPeriod(3651 days);
    }

    function test_SetGracePeriod() public {
        vm.prank(owner);
        registry.setGracePeriod(14 days);
        assertEq(registry.gracePeriod(), 14 days);
    }

    function test_SetGracePeriod_RevertTooShort() public {
        vm.prank(owner);
        vm.expectRevert(BunkerRegistry.InvalidPeriod.selector);
        registry.setGracePeriod(6 days);
    }

    function test_SetGracePeriod_RevertTooLong() public {
        vm.prank(owner);
        vm.expectRevert(BunkerRegistry.InvalidPeriod.selector);
        registry.setGracePeriod(91 days);
    }

    function test_SetReservationPeriod() public {
        vm.prank(owner);
        registry.setReservationPeriod(24 hours);
        assertEq(registry.reservationPeriod(), 24 hours);
    }

    function test_SetReservationPeriod_RevertTooShort() public {
        vm.prank(owner);
        vm.expectRevert(BunkerRegistry.InvalidPeriod.selector);
        registry.setReservationPeriod(30 minutes);
    }

    function test_SetSquattingGracePeriod() public {
        vm.prank(owner);
        registry.setSquattingGracePeriod(14 days);
        assertEq(registry.squattingGracePeriod(), 14 days);
    }

    function test_SetSquattingGracePeriod_RevertTooShort() public {
        vm.prank(owner);
        vm.expectRevert(BunkerRegistry.InvalidPeriod.selector);
        registry.setSquattingGracePeriod(12 hours);
    }

    function test_SetReferralDiscountBps() public {
        vm.prank(owner);
        registry.setReferralDiscountBps(1500);
        assertEq(registry.referralDiscountBps(), 1500);
    }

    function test_SetReferralDiscountBps_RevertTooHigh() public {
        vm.prank(owner);
        vm.expectRevert(BunkerRegistry.InvalidPeriod.selector);
        registry.setReferralDiscountBps(2001);
    }

    function test_SetReferralRewardBps() public {
        vm.prank(owner);
        registry.setReferralRewardBps(800);
        assertEq(registry.referralRewardBps(), 800);
    }

    function test_SetReferralRewardBps_RevertTooHigh() public {
        vm.prank(owner);
        vm.expectRevert(BunkerRegistry.InvalidPeriod.selector);
        registry.setReferralRewardBps(1001);
    }

    function test_SetStakingContract() public {
        address newStaking = makeAddr("newStaking");
        vm.prank(owner);
        registry.setStakingContract(newStaking);
        assertEq(address(registry.stakingContract()), newStaking);
    }

    function test_Pause_Unpause() public {
        vm.startPrank(owner);
        registry.pause();
        assertTrue(registry.paused());
        registry.unpause();
        assertFalse(registry.paused());
        vm.stopPrank();
    }

    // =====================================================================
    //  17. Event Tests
    // =====================================================================

    function test_EmitSubdomainRegistered() public {
        bytes32 depID = bytes32(uint256(1));
        vm.prank(alice);
        vm.expectEmit(false, true, false, true);
        emit BunkerRegistry.SubdomainRegistered("my-app", "my-app", alice, depID, REG_FEE);
        registry.register("my-app", depID);
    }

    function test_EmitSubdomainReleased() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(alice);
        vm.expectEmit(false, true, false, true);
        emit BunkerRegistry.SubdomainReleased("my-app", "my-app", alice);
        registry.release("my-app");
    }

    function test_EmitSubdomainTransferred() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(alice);
        vm.expectEmit(false, true, true, true);
        emit BunkerRegistry.SubdomainTransferred("my-app", "my-app", alice, bob);
        registry.transfer("my-app", bob);
    }

    function test_EmitSubdomainRenewed() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        bytes32 nameHash = keccak256(abi.encodePacked("my-app"));
        (,,, uint48 currentExpiry,,) = registry.subdomains(nameHash);
        uint48 expectedNewExpiry = currentExpiry + uint48(365 days);

        vm.prank(alice);
        vm.expectEmit(false, true, false, true);
        emit BunkerRegistry.SubdomainRenewed("my-app", "my-app", alice, expectedNewExpiry, REG_FEE);
        registry.renew("my-app");
    }

    function test_EmitSubdomainReserved() public {
        uint48 expectedReservedUntil = uint48(block.timestamp + 48 hours);

        vm.prank(alice);
        vm.expectEmit(false, true, false, true);
        emit BunkerRegistry.SubdomainReserved("my-app", "my-app", alice, expectedReservedUntil, REG_FEE);
        registry.reserve("my-app");
    }

    function test_EmitReservationClaimed() public {
        vm.prank(alice);
        registry.reserve("my-app");

        bytes32 depID = bytes32(uint256(42));
        vm.prank(alice);
        vm.expectEmit(false, true, false, true);
        emit BunkerRegistry.ReservationClaimed("my-app", "my-app", alice, depID);
        registry.claimReservation("my-app", depID);
    }

    function test_EmitMetadataUpdated() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(alice);
        vm.expectEmit(false, true, false, true);
        emit BunkerRegistry.MetadataUpdated("my-app", "my-app", alice);
        registry.setMetadata("my-app", "desc", "url");
    }

    function test_EmitPrimaryNameSet() public {
        bytes32 depID = bytes32(uint256(42));
        vm.prank(alice);
        registry.register("my-app", depID);

        vm.prank(alice);
        vm.expectEmit(true, false, true, true);
        emit BunkerRegistry.PrimaryNameSet(depID, "my-app", alice);
        registry.setPrimaryName("my-app");
    }

    // =====================================================================
    //  18. Max Names Per Owner Tests
    // =====================================================================

    function test_Register_RevertTooManyNames() public {
        uint256 maxNames = registry.MAX_NAMES_PER_OWNER();
        vm.startPrank(alice);
        for (uint256 i = 0; i < maxNames; i++) {
            // Use 5-char names to avoid premium pricing (3-char = 10x, 4-char = 5x)
            bytes memory nameBytes = new bytes(5);
            nameBytes[0] = bytes1(uint8(0x61 + (i / 100))); // a-z
            nameBytes[1] = bytes1(uint8(0x30 + ((i / 10) % 10))); // 0-9
            nameBytes[2] = bytes1(uint8(0x30 + (i % 10))); // 0-9
            nameBytes[3] = "x";
            nameBytes[4] = "y";
            registry.register(string(nameBytes), bytes32(uint256(i + 1)));
        }
        vm.stopPrank();

        assertEq(registry.nameCount(alice), maxNames);

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(
            BunkerRegistry.TooManyNames.selector, alice, maxNames));
        registry.register("overflow", bytes32(uint256(999)));
    }

    function test_Transfer_RevertRecipientTooManyNames() public {
        uint256 maxNames = registry.MAX_NAMES_PER_OWNER();
        vm.startPrank(bob);
        for (uint256 i = 0; i < maxNames; i++) {
            bytes memory nameBytes = new bytes(5);
            nameBytes[0] = bytes1(uint8(0x61 + (i / 100)));
            nameBytes[1] = bytes1(uint8(0x30 + ((i / 10) % 10)));
            nameBytes[2] = bytes1(uint8(0x30 + (i % 10)));
            nameBytes[3] = "x";
            nameBytes[4] = "y";
            registry.register(string(nameBytes), bytes32(uint256(i + 1)));
        }
        vm.stopPrank();

        vm.prank(alice);
        registry.register("overflow", bytes32(uint256(999)));

        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(
            BunkerRegistry.TooManyNames.selector, bob, maxNames));
        registry.transfer("overflow", bob);
    }

    // =====================================================================
    //  19. Short Names Gate Tests
    // =====================================================================

    function test_ShortNames_DisabledByDefault() public {
        // Deploy a fresh registry without enabling short names
        BunkerRegistry freshReg = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(staking)
        );
        assertFalse(freshReg.shortNamesEnabled());
    }

    function test_ShortNames_1Char_RevertWhenDisabled() public {
        BunkerRegistry freshReg = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(staking)
        );
        vm.prank(alice);
        token.approve(address(freshReg), type(uint256).max);

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.ShortNamesDisabled.selector);
        freshReg.register("a", bytes32(uint256(1)));
    }

    function test_ShortNames_2Char_RevertWhenDisabled() public {
        BunkerRegistry freshReg = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(staking)
        );
        vm.prank(alice);
        token.approve(address(freshReg), type(uint256).max);

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.ShortNamesDisabled.selector);
        freshReg.register("ab", bytes32(uint256(1)));
    }

    function test_ShortNames_3Char_RevertWhenDisabled() public {
        BunkerRegistry freshReg = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(staking)
        );
        vm.prank(alice);
        token.approve(address(freshReg), type(uint256).max);

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.ShortNamesDisabled.selector);
        freshReg.register("abc", bytes32(uint256(1)));
    }

    function test_ShortNames_ReserveRevertWhenDisabled() public {
        BunkerRegistry freshReg = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(staking)
        );
        vm.prank(alice);
        token.approve(address(freshReg), type(uint256).max);

        vm.prank(alice);
        vm.expectRevert(BunkerRegistry.ShortNamesDisabled.selector);
        freshReg.reserve("ab");
    }

    function test_ShortNames_4CharAlwaysAllowed() public {
        // 4-char names work even when short names disabled
        BunkerRegistry freshReg = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(staking)
        );
        vm.prank(alice);
        token.approve(address(freshReg), type(uint256).max);

        vm.prank(alice);
        freshReg.register("abcd", bytes32(uint256(1)));
        (address regOwner,,) = freshReg.resolve("abcd");
        assertEq(regOwner, alice);
    }

    function test_ShortNames_EnabledByAdmin_AllLengths() public {
        BunkerRegistry freshReg = new BunkerRegistry(
            address(token), treasury, REG_FEE, owner, address(staking)
        );
        vm.prank(alice);
        token.approve(address(freshReg), type(uint256).max);

        vm.prank(owner);
        freshReg.setShortNamesEnabled(true);

        // 1-char
        vm.prank(alice);
        freshReg.register("a", bytes32(uint256(1)));
        (address o1,,) = freshReg.resolve("a");
        assertEq(o1, alice);

        // 2-char
        vm.prank(alice);
        freshReg.register("ab", bytes32(uint256(2)));
        (address o2,,) = freshReg.resolve("ab");
        assertEq(o2, alice);

        // 3-char
        vm.prank(alice);
        freshReg.register("abc", bytes32(uint256(3)));
        (address o3,,) = freshReg.resolve("abc");
        assertEq(o3, alice);
    }

    function test_ShortNames_SetRevertNotOwner() public {
        vm.prank(alice);
        vm.expectRevert();
        registry.setShortNamesEnabled(false);
    }

    // =====================================================================
    //  20. Zero Fee Registration Tests
    // =====================================================================

    function test_ZeroFee_Registration() public {
        vm.prank(owner);
        registry.setRegistrationFee(0);

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.register("free-app", bytes32(uint256(1)));
        assertEq(token.balanceOf(alice), aliceBefore);

        (address regOwner,,) = registry.resolve("free-app");
        assertEq(regOwner, alice);
    }

    function test_ZeroChangeFee_Update() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(owner);
        registry.setChangeFee(0);

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.updateDeployment("my-app", bytes32(uint256(2)));
        assertEq(token.balanceOf(alice), aliceBefore);
    }

    function test_ZeroChangeFee_Metadata() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        vm.prank(owner);
        registry.setChangeFee(0);

        uint256 aliceBefore = token.balanceOf(alice);
        vm.prank(alice);
        registry.setMetadata("my-app", "desc", "url");
        assertEq(token.balanceOf(alice), aliceBefore);
    }

    // =====================================================================
    //  21. Edge Cases
    // =====================================================================

    function test_Register_AfterFullExpiry_CleansUpOldRecord() public {
        vm.prank(alice);
        registry.register("recycle", bytes32(uint256(1)));
        assertEq(registry.nameCount(alice), 1);

        vm.warp(block.timestamp + 365 days + 30 days + 1);

        vm.prank(bob);
        registry.register("recycle", bytes32(uint256(2)));

        // Alice's count should be decremented
        assertEq(registry.nameCount(alice), 0);
        assertEq(registry.nameCount(bob), 1);

        (address regOwner,,) = registry.resolve("recycle");
        assertEq(regOwner, bob);
    }

    function test_RenewDuringGrace_ExtendsFromNow() public {
        vm.prank(alice);
        registry.register("my-app", bytes32(uint256(1)));

        // Warp into grace period
        uint256 graceTime = block.timestamp + 365 days + 15 days;
        vm.warp(graceTime);

        vm.prank(alice);
        registry.renew("my-app");

        bytes32 nameHash = keccak256(abi.encodePacked("my-app"));
        (,,, uint48 newExpiry,,) = registry.subdomains(nameHash);
        // Should extend from now since we're past the original expiry
        assertEq(newExpiry, uint48(graceTime + 365 days));
    }

    function test_MultipleNonConsecutiveHyphensAllowed() public {
        vm.prank(alice);
        registry.register("my-cool-app", bytes32(uint256(1)));
        (address regOwner,,) = registry.resolve("my-cool-app");
        assertEq(regOwner, alice);
    }

    function test_ReserveClaimTransfer_FullFlow() public {
        // Alice reserves
        vm.prank(alice);
        registry.reserve("my-app");

        // Alice claims with deployment
        bytes32 depID = bytes32(uint256(42));
        vm.prank(alice);
        registry.claimReservation("my-app", depID);

        // Alice sets primary name
        vm.prank(alice);
        registry.setPrimaryName("my-app");

        // Alice transfers to bob
        vm.prank(alice);
        registry.transfer("my-app", bob);

        (address regOwner,,) = registry.resolve("my-app");
        assertEq(regOwner, bob);

        // Reverse resolution still works
        string memory name = registry.reverseResolve(depID);
        assertEq(name, "my-app");
    }
}
