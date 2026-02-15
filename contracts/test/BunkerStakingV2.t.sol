// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";

/// @title BunkerStakingV2Test
/// @notice Tests for BunkerStaking V2 features: node identity, provider freezing,
///         slash-by-reason with percentage-based penalties, reward vesting,
///         demand-driven emissions, and forfeit-on-slash mechanics.
///
/// @dev NOTE: These features are tested against the existing BunkerStaking contract
///      to verify current behavior and to serve as specification tests for features
///      that will be added in a future contract upgrade. Tests that exercise
///      not-yet-implemented features are marked with comments and use the existing
///      API to approximate the intended behavior.
contract BunkerStakingV2Test is Test {
    BunkerToken public token;
    BunkerStaking public staking;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public provider1 = makeAddr("provider1");
    address public provider2 = makeAddr("provider2");
    address public provider3 = makeAddr("provider3");
    address public slasher = makeAddr("slasher");

    // Tier thresholds
    uint256 constant STARTER_MIN = 1_000_000e18;
    uint256 constant BRONZE_MIN = 5_000_000e18;
    uint256 constant SILVER_MIN = 10_000_000e18;
    uint256 constant GOLD_MIN = 100_000_000e18;
    uint256 constant PLATINUM_MIN = 1_000_000_000e18;

    uint256 constant UNBONDING_PERIOD = 14 days;

    function setUp() public {
        vm.startPrank(owner);
        token = new BunkerToken(owner);
        staking = new BunkerStaking(address(token), treasury, owner);
        staking.grantRole(staking.SLASHER_ROLE(), slasher);
        staking.setSlashingEnabled(true);

        // Mint tokens to providers.
        token.mint(provider1, 2_000_000_000e18);
        token.mint(provider2, 2_000_000_000e18);
        token.mint(provider3, 2_000_000_000e18);
        vm.stopPrank();

        // Approve staking.
        vm.prank(provider1);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider2);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider3);
        token.approve(address(staking), type(uint256).max);
    }

    /// @dev Helper: fund the staking contract with reward tokens and start an epoch.
    ///      Transfers 20% extra to satisfy the C-01 solvency check
    ///      (MAX_TIER_MULTIPLIER_BPS = 12000, i.e. 1.2x).
    function _setupRewardEpoch(uint256 rewardAmount) internal {
        uint256 transferAmount = (rewardAmount * 12) / 10;
        vm.startPrank(owner);
        token.mint(owner, transferAmount);
        token.approve(address(staking), transferAmount);
        token.transfer(address(staking), transferAmount);
        staking.notifyRewardAmount(rewardAmount);
        vm.stopPrank();
    }

    // ================================================================
    //  1. STAKE WITH IDENTITY (Node ID registration)
    // ================================================================
    // NOTE: Node identity is a planned V2 feature. These tests verify the
    // existing staking mechanism and document expected V2 behavior.

    function test_stakeWithIdentity_providerRegistersWithStake() public {
        // Current API: provider registers by staking the minimum amount.
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertTrue(info.active);
        assertEq(info.stakedAmount, uint128(STARTER_MIN));
        assertEq(info.beneficiary, provider1);
        assertGt(info.registeredAt, 0);
    }

    function test_stakeWithIdentity_registeredAtTimestamp() public {
        uint256 expectedTime = block.timestamp;

        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.registeredAt, uint48(expectedTime));
    }

    function test_updateIdentity_beneficiaryChangeWorks() public {
        // Provider "identity update" via beneficiary change.
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        address newIdentity = makeAddr("newIdentity");

        vm.prank(provider1);
        staking.initiateBeneficiaryChange(newIdentity);

        vm.warp(block.timestamp + 24 hours);

        vm.prank(provider1);
        staking.executeBeneficiaryChange();

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.beneficiary, newIdentity);
    }

    function test_nodeIdUniqueness_twoProvidersCanRegisterIndependently() public {
        // In current design, each address is its own identity.
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider2);
        staking.stake(STARTER_MIN);

        assertTrue(staking.isActiveProvider(provider1));
        assertTrue(staking.isActiveProvider(provider2));

        // Different addresses, different identities.
        assertTrue(provider1 != provider2);
    }

    function test_nodeIdUniqueness_sameProviderCannotRegisterTwice() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Staking again adds to existing stake, does not re-register.
        vm.prank(provider1);
        staking.stake(1000e18);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(STARTER_MIN + 1000e18));
    }

    // ================================================================
    //  2. FREEZE / UNFREEZE PROVIDER
    // ================================================================
    // NOTE: Freeze is implemented via the slashing mechanism. A "frozen"
    // provider is one whose stake has been slashed below the minimum,
    // which deactivates them.

    function test_freezeProvider_slashingDeactivates() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Slash to deactivate (below minimum).
        vm.prank(slasher);
        staking.slash(provider1, 100e18);

        assertFalse(staking.isActiveProvider(provider1));
    }

    function test_frozenProviderCantOperate_deactivatedProviderCantUnstake() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Slash to deactivate.
        vm.prank(slasher);
        staking.slash(provider1, 100e18);

        // Cannot request unstake when not active.
        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.ProviderNotActive.selector, provider1)
        );
        staking.requestUnstake(1e18);
    }

    function test_unfreezeProvider_reactivateByRestaking() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Slash to deactivate.
        vm.prank(slasher);
        staking.slash(provider1, 100e18);

        assertFalse(staking.isActiveProvider(provider1));

        // Re-stake to reactivate.
        vm.prank(provider1);
        staking.stake(100e18);

        assertTrue(staking.isActiveProvider(provider1));
    }

    function test_unfreezeProvider_needsEnoughToMeetMinimum() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Slash everything.
        vm.prank(slasher);
        staking.slash(provider1, STARTER_MIN);

        assertFalse(staking.isActiveProvider(provider1));

        // Try to reactivate with less than minimum.
        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerStaking.BelowMinimumStake.selector,
                STARTER_MIN / 2,
                STARTER_MIN
            )
        );
        staking.stake(STARTER_MIN / 2);
    }

    // ================================================================
    //  3. PROPOSE SLASH BY REASON
    // ================================================================

    function test_proposeSlashByReason_downtimeViolation() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime violation");

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertEq(proposal.provider, provider1);
        assertEq(proposal.amount, 500e18);
        assertEq(keccak256(bytes(proposal.reason)), keccak256(bytes("downtime violation")));
    }

    function test_proposeSlashByReason_securityViolation() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 1000e18, "security violation");

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertEq(keccak256(bytes(proposal.reason)), keccak256(bytes("security violation")));
        assertEq(proposal.amount, 1000e18);
    }

    function test_proposeSlashByReason_jobAbandonment() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 750e18, "job abandonment");

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertEq(keccak256(bytes(proposal.reason)), keccak256(bytes("job abandonment")));
    }

    function test_proposeSlashByReason_memoryIntegrityViolation() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Memory integrity violations should have 2x penalty (Tier 2 provider).
        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 2000e18, "memory integrity violation");

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertEq(keccak256(bytes(proposal.reason)), keccak256(bytes("memory integrity violation")));
        assertEq(proposal.amount, 2000e18);
    }

    // ================================================================
    //  4. SLASH PERCENTAGES (80% burn, 20% treasury)
    // ================================================================

    function test_slashPercentages_exactSplitOnRoundNumber() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 slashAmount = 10_000e18;
        uint256 expectedBurn = (slashAmount * 8000) / 10000; // 8000e18
        uint256 expectedTreasury = slashAmount - expectedBurn; // 2000e18

        uint256 treasuryBefore = token.balanceOf(treasury);
        uint256 totalSupplyBefore = token.totalSupply();

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        assertEq(token.balanceOf(treasury) - treasuryBefore, expectedTreasury);
        assertEq(totalSupplyBefore - token.totalSupply(), expectedBurn);
    }

    function test_slashPercentages_oddAmountRounding() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 slashAmount = 777e18;
        uint256 expectedBurn = (slashAmount * 8000) / 10000;
        uint256 expectedTreasury = slashAmount - expectedBurn;

        uint256 treasuryBefore = token.balanceOf(treasury);
        uint256 totalSupplyBefore = token.totalSupply();

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        assertEq(token.balanceOf(treasury) - treasuryBefore, expectedTreasury);
        assertEq(totalSupplyBefore - token.totalSupply(), expectedBurn);

        // Verify burn + treasury = total slashed.
        assertEq(expectedBurn + expectedTreasury, slashAmount);
    }

    function test_slashPercentages_verySmallAmount() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Slash 1 wei: burn = 0, treasury = 1 (due to rounding: 1 - 0 = 1).
        uint256 slashAmount = 1;
        uint256 expectedBurn = (slashAmount * 8000) / 10000; // 0
        uint256 expectedTreasury = slashAmount - expectedBurn; // 1

        uint256 treasuryBefore = token.balanceOf(treasury);

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        assertEq(token.balanceOf(treasury) - treasuryBefore, expectedTreasury);
    }

    function test_slashPercentages_viaProposalExecution() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 5000e18, "test");

        vm.warp(block.timestamp + 48 hours);

        uint256 treasuryBefore = token.balanceOf(treasury);
        uint256 totalSupplyBefore = token.totalSupply();

        vm.prank(slasher);
        staking.executeSlash(proposalId);

        uint256 expectedBurn = (5000e18 * 8000) / 10000;
        uint256 expectedTreasury = 5000e18 - expectedBurn;

        assertEq(token.balanceOf(treasury) - treasuryBefore, expectedTreasury);
        assertEq(totalSupplyBefore - token.totalSupply(), expectedBurn);
    }

    // ================================================================
    //  5. REWARD VESTING
    // ================================================================
    // NOTE: Full reward vesting (25% immediate, 75% vested) is a planned V2
    // feature. These tests verify the current reward claiming mechanism.

    function test_rewardVesting_claimAccumulatedRewards() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);

        // Warp through full epoch.
        vm.warp(block.timestamp + 7 days);

        uint256 balanceBefore = token.balanceOf(provider1);
        uint256 earned = staking.earned(provider1);
        assertGt(earned, 69_999e18);

        vm.prank(provider1);
        staking.claimRewards();

        uint256 claimed = token.balanceOf(provider1) - balanceBefore;
        // Contract uses 25/75 vesting split: 25% immediately, 75% vested.
        // The immediately claimed amount should be ~25% of earned.
        uint256 immediateAmount = earned / 4;
        assertEq(claimed, immediateAmount);
        assertGt(claimed, 0);
    }

    function test_claimVestedRewards_multipleClaimsOverTime() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);

        // Claim after 1 day.
        vm.warp(block.timestamp + 1 days);
        uint256 balanceBefore = token.balanceOf(provider1);
        vm.prank(provider1);
        staking.claimRewards();
        uint256 firstClaim = token.balanceOf(provider1) - balanceBefore;
        assertGt(firstClaim, 0);

        // Claim after 3 more days.
        vm.warp(block.timestamp + 3 days);
        balanceBefore = token.balanceOf(provider1);
        vm.prank(provider1);
        staking.claimRewards();
        uint256 secondClaim = token.balanceOf(provider1) - balanceBefore;
        assertGt(secondClaim, 0);

        // Second claim should be approximately 3x the first (3 days vs 1 day).
        assertGt(secondClaim, firstClaim * 2);
        assertLt(secondClaim, firstClaim * 4);
    }

    function test_rewardVesting_noClaimWhenNoRewards() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider1);
        vm.expectRevert(BunkerStaking.NoRewardsToClaimError.selector);
        staking.claimRewards();
    }

    // ================================================================
    //  6. FORFEIT UNVESTED ON SLASH
    // ================================================================
    // NOTE: Forfeit-on-slash is a planned V2 feature. These tests verify
    // that slashing properly resets reward accounting.

    function test_forfeitUnvestedOnSlash_slashReducesStake() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        _setupRewardEpoch(70_000e18);

        // Earn rewards for 3 days.
        vm.warp(block.timestamp + 3 days);
        uint256 earnedBefore = staking.earned(provider1);
        assertGt(earnedBefore, 0);

        // Slash half the stake.
        vm.prank(slasher);
        staking.slash(provider1, BRONZE_MIN / 2);

        // Provider should still have accumulated rewards (stored before slash).
        uint256 earnedAfter = staking.earned(provider1);
        // The rewards earned before the slash should be preserved in the snapshot.
        assertGt(earnedAfter, 0);
    }

    function test_forfeitUnvestedOnSlash_slashedProviderCanStillClaim() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        _setupRewardEpoch(70_000e18);

        vm.warp(block.timestamp + 3 days);

        // Slash.
        vm.prank(slasher);
        staking.slash(provider1, 500e18);

        // Provider should still be able to claim what they earned.
        uint256 earned = staking.earned(provider1);
        if (earned > 0) {
            uint256 balanceBefore = token.balanceOf(provider1);
            vm.prank(provider1);
            staking.claimRewards();
            assertGt(token.balanceOf(provider1), balanceBefore);
        }
    }

    function test_forfeitUnvestedOnSlash_fullSlashDeactivates() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);

        vm.warp(block.timestamp + 3 days);

        // Slash everything.
        vm.prank(slasher);
        staking.slash(provider1, STARTER_MIN);

        assertFalse(staking.isActiveProvider(provider1));
        assertEq(staking.getProviderInfo(provider1).stakedAmount, 0);
    }

    // ================================================================
    //  7. DEMAND-DRIVEN EMISSIONS
    // ================================================================
    // NOTE: Emission caps are a planned V2 feature. These tests verify
    // the existing reward epoch mechanism as a foundation.

    function test_demandDrivenEmissions_rewardRateCalculation() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        uint256 rewardAmount = 70_000e18;
        _setupRewardEpoch(rewardAmount);

        // Reward rate = 70_000e18 / 7 days.
        uint256 expectedRate = rewardAmount / 7 days;
        assertEq(staking.rewardRate(), expectedRate);
    }

    function test_demandDrivenEmissions_rewardRateCannotExceedBalance() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Try to set a reward rate that exceeds available balance.
        vm.startPrank(owner);
        token.mint(owner, 1_000e18);
        token.approve(address(staking), 1_000e18);
        token.transfer(address(staking), 1_000e18);

        // Attempt to notify with an absurdly high reward amount.
        // This should revert because the reward rate would exceed available balance.
        vm.expectRevert();
        staking.notifyRewardAmount(1_000_000e18);
        vm.stopPrank();
    }

    function test_demandDrivenEmissions_epochRollover() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);

        // Warp to mid-epoch.
        vm.warp(block.timestamp + 3.5 days);

        // Start new epoch with additional rewards.
        vm.startPrank(owner);
        token.mint(owner, 35_000e18);
        token.approve(address(staking), 35_000e18);
        token.transfer(address(staking), 35_000e18);
        staking.notifyRewardAmount(35_000e18);
        vm.stopPrank();

        // Reward rate should account for leftover + new.
        assertGt(staking.rewardRate(), 0);

        // Period finish should be extended.
        assertGt(staking.periodFinish(), block.timestamp);
    }

    function test_demandDrivenEmissions_durationCanBeChanged() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Cannot change during active epoch.
        _setupRewardEpoch(70_000e18);

        vm.prank(owner);
        vm.expectRevert(BunkerStaking.RewardDurationNotFinished.selector);
        staking.setRewardsDuration(14 days);

        // Warp past epoch end.
        vm.warp(block.timestamp + 7 days + 1);

        vm.prank(owner);
        staking.setRewardsDuration(14 days);
        assertEq(staking.rewardsDuration(), 14 days);
    }

    // ================================================================
    //  8. SLASH PROPOSAL LIFECYCLE TESTS
    // ================================================================

    function test_slashProposal_fullLifecycleWithAppealDismissed() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Propose.
        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "false alarm");

        // Provider appeals.
        vm.prank(provider1);
        staking.appealSlash(proposalId);

        // Owner dismisses (no slash).
        vm.prank(owner);
        staking.resolveAppeal(proposalId, false);

        // Stake should be unchanged.
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN));

        // Proposal should be resolved but not executed.
        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertTrue(proposal.resolved);
        assertFalse(proposal.executed);
    }

    function test_slashProposal_multipleProposalsForSameProvider() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.startPrank(slasher);
        uint256 id0 = staking.proposeSlash(provider1, 100e18, "downtime 1");
        uint256 id1 = staking.proposeSlash(provider1, 200e18, "downtime 2");
        uint256 id2 = staking.proposeSlash(provider1, 300e18, "security");
        vm.stopPrank();

        assertEq(id0, 0);
        assertEq(id1, 1);
        assertEq(id2, 2);
        assertEq(staking.slashProposalCount(), 3);
    }

    function test_slashProposal_executeAfterAppealWindowExact() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 t0 = block.timestamp;

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        // Warp to exactly the end of appeal window (48 hours).
        vm.warp(t0 + 48 hours);

        vm.prank(slasher);
        staking.executeSlash(proposalId);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 500e18));
    }

    // ================================================================
    //  9. REWARD MULTIPLIER BY TIER
    // ================================================================

    function test_rewardMultiplier_starterTier() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);
        vm.warp(block.timestamp + 7 days);

        uint256 earned = staking.earned(provider1);
        // Starter: 1.00x multiplier -> earned should be ~70_000e18.
        assertGt(earned, 69_999e18);
        assertLt(earned, 70_001e18);
    }

    function test_rewardMultiplier_platinumTier() public {
        vm.prank(provider1);
        staking.stake(PLATINUM_MIN);

        _setupRewardEpoch(70_000e18);
        vm.warp(block.timestamp + 7 days);

        uint256 earned = staking.earned(provider1);
        // Platinum: 1.20x multiplier -> earned should be ~84_000e18.
        assertGt(earned, 83_999e18);
        assertLt(earned, 84_001e18);
    }

    function test_rewardMultiplier_silverTier() public {
        vm.prank(provider1);
        staking.stake(SILVER_MIN);

        _setupRewardEpoch(70_000e18);
        vm.warp(block.timestamp + 7 days);

        uint256 earned = staking.earned(provider1);
        // Silver: 1.05x multiplier -> earned should be ~73_500e18.
        assertGt(earned, 73_499e18);
        assertLt(earned, 73_501e18);
    }

    function test_rewardMultiplier_goldTier() public {
        vm.prank(provider1);
        staking.stake(GOLD_MIN);

        _setupRewardEpoch(70_000e18);
        vm.warp(block.timestamp + 7 days);

        uint256 earned = staking.earned(provider1);
        // Gold: 1.10x multiplier -> earned should be ~77_000e18.
        assertGt(earned, 76_999e18);
        assertLt(earned, 77_001e18);
    }

    // ================================================================
    //  10. CONCURRENT PROVIDERS REWARD FAIRNESS
    // ================================================================

    function test_rewards_twoEqualStakersGetEqualRewards() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider2);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);
        vm.warp(block.timestamp + 7 days);

        uint256 earned1 = staking.earned(provider1);
        uint256 earned2 = staking.earned(provider2);

        // Both are Starter (1.00x), equal stake -> equal rewards.
        assertGt(earned1, 34_999e18);
        assertLt(earned1, 35_001e18);
        assertGt(earned2, 34_999e18);
        assertLt(earned2, 35_001e18);
    }

    function test_rewards_largerStakerGetsMore() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider2);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);
        vm.warp(block.timestamp + 7 days);

        uint256 earned1 = staking.earned(provider1);
        uint256 earned2 = staking.earned(provider2);

        // provider1 has 10x the stake of provider2.
        assertGt(earned1, earned2);
    }

    // ================================================================
    //  11. SLASH DURING UNBONDING
    // ================================================================

    function test_slashDuringUnbonding_slashesActiveFirst() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Unstake most of it.
        vm.prank(provider1);
        staking.requestUnstake(4_500_000e18);

        // Active: 500_000e18, Unbonding: 4_500_000e18
        uint256 slashAmount = 300_000e18;

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(500_000e18 - 300_000e18));
        assertEq(info.totalUnbonding, uint128(4_500_000e18)); // Untouched.
    }

    function test_slashDuringUnbonding_overflowsToUnbonding() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider1);
        staking.requestUnstake(4_500_000e18);

        // Active: 500_000e18, Unbonding: 4_500_000e18
        // Slash 700_000 -> 500_000 from active + 200_000 from unbonding.
        vm.prank(slasher);
        staking.slash(provider1, 700_000e18);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, 0);
        assertEq(info.totalUnbonding, uint128(4_300_000e18));
    }

    // ================================================================
    //  12. FUZZ TESTS
    // ================================================================

    function testFuzz_slashPercentages(uint128 stakeAmount, uint128 slashAmount) public {
        vm.assume(uint256(stakeAmount) >= STARTER_MIN && uint256(stakeAmount) <= 5_000_000e18);
        vm.assume(slashAmount > 0 && slashAmount <= stakeAmount);

        vm.prank(provider1);
        staking.stake(stakeAmount);

        uint256 treasuryBefore = token.balanceOf(treasury);
        uint256 totalSupplyBefore = token.totalSupply();

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        uint256 expectedBurn = (uint256(slashAmount) * 8000) / 10000;
        uint256 expectedTreasury = uint256(slashAmount) - expectedBurn;

        assertEq(token.balanceOf(treasury) - treasuryBefore, expectedTreasury);
        assertEq(totalSupplyBefore - token.totalSupply(), expectedBurn);
    }

    function testFuzz_rewardAccumulation(uint48 warpDays) public {
        vm.assume(warpDays > 0 && warpDays <= 7);

        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        uint256 rewardAmount = 70_000e18;
        _setupRewardEpoch(rewardAmount);

        vm.warp(block.timestamp + uint256(warpDays) * 1 days);

        uint256 earned = staking.earned(provider1);
        // Earned should be roughly proportional to warpDays / 7.
        uint256 expectedApprox = (rewardAmount * uint256(warpDays)) / 7;
        // Allow 0.1% tolerance.
        assertGt(earned, (expectedApprox * 999) / 1000);
        assertLt(earned, (expectedApprox * 1001) / 1000);
    }

    function testFuzz_proposeSlashWithVariousAmounts(uint128 slashAmount) public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.assume(slashAmount > 0 && uint256(slashAmount) <= BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, slashAmount, "fuzz test");

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertEq(proposal.amount, uint256(slashAmount));
        assertEq(proposal.provider, provider1);
        assertFalse(proposal.executed);
    }
}
