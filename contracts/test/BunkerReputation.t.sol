// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";

/// @title BunkerReputationTest
/// @notice Tests for the reputation system built on top of BunkerStaking.
///         Reputation is derived from staking tier, provider activity, slashing
///         history, and reward claiming patterns.
///
/// @dev The on-chain reputation system is a planned V2 feature. These tests
///      verify the existing staking infrastructure that serves as the foundation
///      for reputation scoring:
///        - Registration timestamp (longevity)
///        - Staking tier (capital commitment)
///        - Slashing history (reliability)
///        - Active/inactive status (availability)
///
///      The reputation score (0-1000) and decay mechanics will be implemented
///      in a dedicated BunkerReputation contract. These tests document the
///      expected behavior and validate the staking contract's suitability as
///      a data source for reputation.
contract BunkerReputationTest is Test {
    BunkerToken public token;
    BunkerStaking public staking;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public provider1 = makeAddr("provider1");
    address public provider2 = makeAddr("provider2");
    address public provider3 = makeAddr("provider3");
    address public slasher = makeAddr("slasher");

    uint256 constant STARTER_MIN = 1_000_000e18;
    uint256 constant BRONZE_MIN = 5_000_000e18;
    uint256 constant SILVER_MIN = 10_000_000e18;
    uint256 constant GOLD_MIN = 100_000_000e18;
    uint256 constant PLATINUM_MIN = 1_000_000_000e18;

    function setUp() public {
        vm.startPrank(owner);
        token = new BunkerToken(owner);
        staking = new BunkerStaking(address(token), treasury, owner);
        staking.grantRole(staking.SLASHER_ROLE(), slasher);
        staking.setSlashingEnabled(true);

        token.mint(provider1, 2_000_000_000e18);
        token.mint(provider2, 2_000_000_000e18);
        token.mint(provider3, 2_000_000_000e18);
        vm.stopPrank();

        vm.prank(provider1);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider2);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider3);
        token.approve(address(staking), type(uint256).max);
    }

    // ================================================================
    //  1. INITIAL SCORE (Registration)
    // ================================================================

    function test_initialScore_providerRegistersAsActive() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        assertTrue(staking.isActiveProvider(provider1));

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertGt(info.registeredAt, 0);
    }

    function test_initialScore_registrationTimestampRecorded() public {
        uint256 expectedTime = block.timestamp;

        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.registeredAt, uint48(expectedTime));
    }

    function test_initialScore_unregisteredProviderHasNoInfo() public view {
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertFalse(info.active);
        assertEq(info.stakedAmount, 0);
        assertEq(info.registeredAt, 0);
    }

    function test_initialScore_higherTierIndicatesHigherReputation() public {
        vm.prank(provider1);
        staking.stake(PLATINUM_MIN);

        vm.prank(provider2);
        staking.stake(STARTER_MIN);

        // Platinum provider should have higher implicit reputation.
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Platinum));
        assertEq(uint8(staking.getTier(provider2)), uint8(BunkerStaking.Tier.Starter));
    }

    // ================================================================
    //  2. RECORD JOB COMPLETED (score increases via reward claims)
    // ================================================================

    function test_recordJobCompleted_providerEarnsRewards() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Set up rewards as proxy for job completion.
        vm.startPrank(owner);
        token.mint(owner, 84_000e18);
        token.approve(address(staking), 84_000e18);
        token.transfer(address(staking), 84_000e18);
        staking.notifyRewardAmount(70_000e18);
        vm.stopPrank();

        vm.warp(block.timestamp + 7 days);

        uint256 earned = staking.earned(provider1);
        assertGt(earned, 0, "Provider should have earned rewards after completing epoch");
    }

    function test_recordJobCompleted_rewardClaimIncreasesBalance() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.startPrank(owner);
        token.mint(owner, 84_000e18);
        token.approve(address(staking), 84_000e18);
        token.transfer(address(staking), 84_000e18);
        staking.notifyRewardAmount(70_000e18);
        vm.stopPrank();

        vm.warp(block.timestamp + 7 days);

        uint256 balanceBefore = token.balanceOf(provider1);

        vm.prank(provider1);
        staking.claimRewards();

        assertGt(token.balanceOf(provider1), balanceBefore);
    }

    // ================================================================
    //  3. RECORD SLASH EVENT (score decreases)
    // ================================================================

    function test_recordSlashEvent_slashReducesStake() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 stakeBefore = staking.getProviderInfo(provider1).stakedAmount;

        vm.prank(slasher);
        staking.slash(provider1, 100_000e18);

        uint256 stakeAfter = staking.getProviderInfo(provider1).stakedAmount;
        assertEq(stakeAfter, stakeBefore - 100_000e18);
    }

    function test_recordSlashEvent_emitsSlashedEvent() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 slashAmount = 100_000e18;
        uint256 expectedBurn = (slashAmount * 8000) / 10000;
        uint256 expectedTreasury = slashAmount - expectedBurn;

        vm.prank(slasher);
        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.Slashed(provider1, slashAmount, expectedBurn, expectedTreasury);
        staking.slash(provider1, slashAmount);
    }

    function test_recordSlashEvent_multipleLashesAccumulate() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.startPrank(slasher);
        staking.slash(provider1, 100_000e18);
        staking.slash(provider1, 100_000e18);
        staking.slash(provider1, 100_000e18);
        vm.stopPrank();

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 300_000e18));
    }

    // ================================================================
    //  4. SCORE CAPPED (max 1000, min 0 via tier system)
    // ================================================================

    function test_scoreCapped_maxTierIsPlatinum() public {
        vm.prank(provider1);
        staking.stake(PLATINUM_MIN);

        // Staking more than Platinum threshold still returns Platinum.
        vm.prank(provider1);
        staking.stake(100_000e18);

        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Platinum));
    }

    function test_scoreCapped_minTierIsNone() public view {
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.None));
        assertEq(uint8(staking.getTierForAmount(0)), uint8(BunkerStaking.Tier.None));
        assertEq(uint8(staking.getTierForAmount(1)), uint8(BunkerStaking.Tier.None));
    }

    function test_scoreCapped_slashToZeroStillWorks() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(slasher);
        staking.slash(provider1, STARTER_MIN);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, 0);
        assertFalse(info.active);
    }

    // ================================================================
    //  5. TIER CALCULATION
    // ================================================================

    function test_tierCalculation_allTiersCorrect() public view {
        assertEq(uint8(staking.getTierForAmount(0)), uint8(BunkerStaking.Tier.None));
        assertEq(uint8(staking.getTierForAmount(STARTER_MIN)), uint8(BunkerStaking.Tier.Starter));
        assertEq(uint8(staking.getTierForAmount(BRONZE_MIN)), uint8(BunkerStaking.Tier.Bronze));
        assertEq(uint8(staking.getTierForAmount(SILVER_MIN)), uint8(BunkerStaking.Tier.Silver));
        assertEq(uint8(staking.getTierForAmount(GOLD_MIN)), uint8(BunkerStaking.Tier.Gold));
        assertEq(uint8(staking.getTierForAmount(PLATINUM_MIN)), uint8(BunkerStaking.Tier.Platinum));
    }

    function test_tierCalculation_boundaryValues() public view {
        assertEq(uint8(staking.getTierForAmount(STARTER_MIN - 1)), uint8(BunkerStaking.Tier.None));
        assertEq(uint8(staking.getTierForAmount(BRONZE_MIN - 1)), uint8(BunkerStaking.Tier.Starter));
        assertEq(uint8(staking.getTierForAmount(SILVER_MIN - 1)), uint8(BunkerStaking.Tier.Bronze));
        assertEq(uint8(staking.getTierForAmount(GOLD_MIN - 1)), uint8(BunkerStaking.Tier.Silver));
        assertEq(uint8(staking.getTierForAmount(PLATINUM_MIN - 1)), uint8(BunkerStaking.Tier.Gold));
    }

    function test_tierCalculation_slashingDemotesTier() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Bronze));

        // Slash down to Starter level.
        vm.prank(slasher);
        staking.slash(provider1, BRONZE_MIN - STARTER_MIN);

        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Starter));
    }

    function test_tierCalculation_stakingPromotesTier() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Starter));

        vm.prank(provider1);
        staking.stake(BRONZE_MIN - STARTER_MIN);
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Bronze));
    }

    // ================================================================
    //  6. DECAY (longevity simulation)
    // ================================================================

    function test_decay_rewardRateDecreasesPastEpochEnd() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.startPrank(owner);
        token.mint(owner, 84_000e18);
        token.approve(address(staking), 84_000e18);
        token.transfer(address(staking), 84_000e18);
        staking.notifyRewardAmount(70_000e18);
        vm.stopPrank();

        // Earn during epoch.
        vm.warp(block.timestamp + 7 days);
        uint256 earnedAtEnd = staking.earned(provider1);

        // Earn past epoch end (no more rewards).
        vm.warp(block.timestamp + 7 days);
        uint256 earnedPastEnd = staking.earned(provider1);

        // No additional rewards after epoch ends.
        assertEq(earnedAtEnd, earnedPastEnd);
    }

    function test_decay_noNewRewardsAfterEpoch() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.startPrank(owner);
        token.mint(owner, 84_000e18);
        token.approve(address(staking), 84_000e18);
        token.transfer(address(staking), 84_000e18);
        staking.notifyRewardAmount(70_000e18);
        vm.stopPrank();

        vm.warp(block.timestamp + 14 days); // 2 epochs past

        uint256 earned = staking.earned(provider1);

        // Should be capped at the full epoch reward (7 days worth).
        assertGt(earned, 69_999e18);
        assertLt(earned, 70_001e18);
    }

    // ================================================================
    //  7. DECAY FLOOR (minimum viable score)
    // ================================================================

    function test_decayFloor_providerRemainsRegisteredAfterEpoch() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Warp far into the future.
        vm.warp(block.timestamp + 365 days);

        // Provider is still active (stake hasn't changed).
        assertTrue(staking.isActiveProvider(provider1));
        assertEq(staking.getProviderInfo(provider1).stakedAmount, uint128(STARTER_MIN));
    }

    function test_decayFloor_slashedProviderCanRecover() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Slash to just above minimum.
        vm.prank(slasher);
        staking.slash(provider1, BRONZE_MIN - STARTER_MIN);

        assertTrue(staking.isActiveProvider(provider1));
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Starter));

        // Can recover by staking more.
        vm.prank(provider1);
        staking.stake(BRONZE_MIN - STARTER_MIN);
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Bronze));
    }

    // ================================================================
    //  8. JOB ELIGIBILITY (minimum tier required)
    // ================================================================

    function test_jobEligibility_activeProviderIsEligible() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        assertTrue(staking.isActiveProvider(provider1));
    }

    function test_jobEligibility_inactiveProviderNotEligible() public view {
        assertFalse(staking.isActiveProvider(provider1));
    }

    function test_jobEligibility_deactivatedAfterSlashNotEligible() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);
        assertTrue(staking.isActiveProvider(provider1));

        // Slash below minimum.
        vm.prank(slasher);
        staking.slash(provider1, 1e18);

        assertFalse(staking.isActiveProvider(provider1));
    }

    function test_jobEligibility_deactivatedAfterFullUnstakeNotEligible() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider1);
        staking.requestUnstake(STARTER_MIN);

        assertFalse(staking.isActiveProvider(provider1));
    }

    function test_jobEligibility_higherTierMoreCapabilities() public {
        // Verify tier configs: higher tiers have more concurrent jobs.
        (,uint16 starterJobs,,,) = staking.tierConfigs(BunkerStaking.Tier.Starter);
        (,uint16 bronzeJobs,,,) = staking.tierConfigs(BunkerStaking.Tier.Bronze);
        (,uint16 silverJobs,,,) = staking.tierConfigs(BunkerStaking.Tier.Silver);
        (,uint16 goldJobs,,,) = staking.tierConfigs(BunkerStaking.Tier.Gold);
        (,uint16 platinumJobs,,,) = staking.tierConfigs(BunkerStaking.Tier.Platinum);

        assertEq(starterJobs, 3);
        assertEq(bronzeJobs, 10);
        assertEq(silverJobs, 50);
        assertEq(goldJobs, 200);
        assertEq(platinumJobs, 0); // 0 = unlimited
    }

    function test_jobEligibility_priorityQueueForHigherTiers() public {
        // Silver, Gold, Platinum have priority queue access.
        (,,, bool starterPriority,) = staking.tierConfigs(BunkerStaking.Tier.Starter);
        (,,, bool silverPriority,) = staking.tierConfigs(BunkerStaking.Tier.Silver);
        (,,, bool goldPriority,) = staking.tierConfigs(BunkerStaking.Tier.Gold);
        (,,, bool platinumPriority,) = staking.tierConfigs(BunkerStaking.Tier.Platinum);

        assertFalse(starterPriority);
        assertTrue(silverPriority);
        assertTrue(goldPriority);
        assertTrue(platinumPriority);
    }

    function test_jobEligibility_governanceForGoldAndAbove() public {
        (,,,, bool starterGov) = staking.tierConfigs(BunkerStaking.Tier.Starter);
        (,,,, bool bronzeGov) = staking.tierConfigs(BunkerStaking.Tier.Bronze);
        (,,,, bool goldGov) = staking.tierConfigs(BunkerStaking.Tier.Gold);
        (,,,, bool platinumGov) = staking.tierConfigs(BunkerStaking.Tier.Platinum);

        assertFalse(starterGov);
        assertFalse(bronzeGov);
        assertTrue(goldGov);
        assertTrue(platinumGov);
    }

    // ================================================================
    //  9. FUZZ TESTS
    // ================================================================

    function testFuzz_tierCalculation(uint128 amount) public view {
        BunkerStaking.Tier tier = staking.getTierForAmount(uint256(amount));

        if (uint256(amount) >= PLATINUM_MIN) {
            assertEq(uint8(tier), uint8(BunkerStaking.Tier.Platinum));
        } else if (uint256(amount) >= GOLD_MIN) {
            assertEq(uint8(tier), uint8(BunkerStaking.Tier.Gold));
        } else if (uint256(amount) >= SILVER_MIN) {
            assertEq(uint8(tier), uint8(BunkerStaking.Tier.Silver));
        } else if (uint256(amount) >= BRONZE_MIN) {
            assertEq(uint8(tier), uint8(BunkerStaking.Tier.Bronze));
        } else if (uint256(amount) >= STARTER_MIN) {
            assertEq(uint8(tier), uint8(BunkerStaking.Tier.Starter));
        } else {
            assertEq(uint8(tier), uint8(BunkerStaking.Tier.None));
        }
    }

    function testFuzz_slashAndTierDemotion(uint128 stakeAmount, uint128 slashAmount) public {
        vm.assume(uint256(stakeAmount) >= STARTER_MIN && uint256(stakeAmount) <= 5_000_000e18);
        vm.assume(slashAmount > 0 && slashAmount <= stakeAmount);

        vm.prank(provider1);
        staking.stake(stakeAmount);

        BunkerStaking.Tier tierBefore = staking.getTier(provider1);

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        BunkerStaking.Tier tierAfter = staking.getTier(provider1);

        // Tier should either stay the same or decrease.
        assertTrue(uint8(tierAfter) <= uint8(tierBefore));
    }
}
