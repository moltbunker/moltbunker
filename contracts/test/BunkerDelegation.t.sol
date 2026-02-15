// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";

/// @title BunkerDelegationTest
/// @notice Tests for delegation features built on top of the existing BunkerStaking
///         contract. Delegation is modeled using the beneficiary mechanism and the
///         staking system. These tests verify the foundation for a future dedicated
///         delegation contract.
///
/// @dev Delegation concept: A delegator stakes tokens to a provider's pool.
///      The provider earns rewards and shares a configurable cut with delegators.
///      In the current contract, delegation is approximated via:
///        - Staking tokens (each address stakes independently)
///        - Beneficiary changes (to route rewards)
///        - Unstaking with unbonding period (undelegation)
contract BunkerDelegationTest is Test {
    BunkerToken public token;
    BunkerStaking public staking;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public provider = makeAddr("provider");
    address public delegator1 = makeAddr("delegator1");
    address public delegator2 = makeAddr("delegator2");
    address public delegator3 = makeAddr("delegator3");
    address public slasher = makeAddr("slasher");

    uint256 constant STARTER_MIN = 1_000_000e18;
    uint256 constant BRONZE_MIN = 5_000_000e18;
    uint256 constant UNBONDING_PERIOD = 14 days;
    uint256 constant BENEFICIARY_TIMELOCK = 24 hours;

    function setUp() public {
        vm.startPrank(owner);
        token = new BunkerToken(owner);
        staking = new BunkerStaking(address(token), treasury, owner);
        staking.grantRole(staking.SLASHER_ROLE(), slasher);
        staking.setSlashingEnabled(true);

        // Mint tokens.
        token.mint(provider, 10_000_000e18);
        token.mint(delegator1, 10_000_000e18);
        token.mint(delegator2, 10_000_000e18);
        token.mint(delegator3, 10_000_000e18);
        vm.stopPrank();

        // Approve staking.
        vm.prank(provider);
        token.approve(address(staking), type(uint256).max);
        vm.prank(delegator1);
        token.approve(address(staking), type(uint256).max);
        vm.prank(delegator2);
        token.approve(address(staking), type(uint256).max);
        vm.prank(delegator3);
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
    //  1. DELEGATE (stake to provider pool)
    // ================================================================

    function test_delegate_delegatorCanStakeIndependently() public {
        // Provider stakes first.
        vm.prank(provider);
        staking.stake(BRONZE_MIN);

        // Delegator stakes independently.
        vm.prank(delegator1);
        staking.stake(STARTER_MIN);

        assertTrue(staking.isActiveProvider(provider));
        assertTrue(staking.isActiveProvider(delegator1));

        // Both contribute to totalStaked.
        assertEq(staking.totalStaked(), BRONZE_MIN + STARTER_MIN);
    }

    function test_delegate_multipleDelegatorsCanStake() public {
        vm.prank(provider);
        staking.stake(BRONZE_MIN);

        vm.prank(delegator1);
        staking.stake(STARTER_MIN);

        vm.prank(delegator2);
        staking.stake(STARTER_MIN);

        vm.prank(delegator3);
        staking.stake(STARTER_MIN);

        assertEq(staking.totalStaked(), BRONZE_MIN + (STARTER_MIN * 3));
    }

    function test_delegate_delegatorEarnsRewardsProportionally() public {
        // Provider with Bronze tier.
        vm.prank(provider);
        staking.stake(BRONZE_MIN);

        // Delegator with Starter tier.
        vm.prank(delegator1);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);
        vm.warp(block.timestamp + 7 days);

        uint256 providerEarned = staking.earned(provider);
        uint256 delegatorEarned = staking.earned(delegator1);

        // Provider has 10x the stake -> should earn proportionally more.
        assertGt(providerEarned, delegatorEarned);
    }

    // ================================================================
    //  2. REQUEST UNDELEGATE (start unbonding)
    // ================================================================

    function test_requestUndelegate_startsUnbonding() public {
        vm.prank(delegator1);
        staking.stake(BRONZE_MIN);

        vm.prank(delegator1);
        staking.requestUnstake(500_000e18);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(delegator1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 500_000e18));
        assertEq(info.totalUnbonding, uint128(500_000e18));
    }

    function test_requestUndelegate_emitsUnstakeRequestedEvent() public {
        vm.prank(delegator1);
        staking.stake(BRONZE_MIN);

        uint256 expectedUnlockTime = block.timestamp + UNBONDING_PERIOD;

        vm.prank(delegator1);
        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.UnstakeRequested(delegator1, 500_000e18, expectedUnlockTime, 0);
        staking.requestUnstake(500_000e18);
    }

    function test_requestUndelegate_fullUndelegationDeactivates() public {
        vm.prank(delegator1);
        staking.stake(STARTER_MIN);

        vm.prank(delegator1);
        staking.requestUnstake(STARTER_MIN);

        assertFalse(staking.isActiveProvider(delegator1));
    }

    // ================================================================
    //  3. COMPLETE UNDELEGATE (after unbonding period)
    // ================================================================

    function test_completeUndelegate_afterPeriodSucceeds() public {
        vm.prank(delegator1);
        staking.stake(BRONZE_MIN);

        vm.prank(delegator1);
        staking.requestUnstake(500_000e18);

        vm.warp(block.timestamp + UNBONDING_PERIOD);

        uint256 balanceBefore = token.balanceOf(delegator1);

        vm.prank(delegator1);
        staking.completeUnstake(0);

        assertEq(token.balanceOf(delegator1), balanceBefore + 500_000e18);
    }

    function test_completeUndelegate_beforePeriodReverts() public {
        vm.prank(delegator1);
        staking.stake(BRONZE_MIN);

        vm.prank(delegator1);
        staking.requestUnstake(500_000e18);

        // Try before unbonding period.
        vm.prank(delegator1);
        vm.expectRevert();
        staking.completeUnstake(0);
    }

    function test_completeUndelegate_fullFlow() public {
        // 1. Delegate (stake).
        vm.prank(delegator1);
        staking.stake(BRONZE_MIN);

        assertTrue(staking.isActiveProvider(delegator1));

        // 2. Request undelegate.
        vm.prank(delegator1);
        staking.requestUnstake(BRONZE_MIN);

        assertFalse(staking.isActiveProvider(delegator1));

        // 3. Wait for unbonding.
        vm.warp(block.timestamp + UNBONDING_PERIOD);

        // 4. Complete.
        uint256 balanceBefore = token.balanceOf(delegator1);
        vm.prank(delegator1);
        staking.completeUnstake(0);

        assertEq(token.balanceOf(delegator1), balanceBefore + BRONZE_MIN);
        assertEq(staking.totalUnbonding(), 0);
    }

    // ================================================================
    //  4. PROVIDER CONFIG (beneficiary as reward routing)
    // ================================================================

    function test_providerConfig_setBeneficiaryAsRewardRoute() public {
        vm.prank(delegator1);
        staking.stake(STARTER_MIN);

        // Set beneficiary to route rewards to a different address.
        address rewardWallet = makeAddr("rewardWallet");

        vm.prank(delegator1);
        staking.initiateBeneficiaryChange(rewardWallet);

        vm.warp(block.timestamp + BENEFICIARY_TIMELOCK);

        vm.prank(delegator1);
        staking.executeBeneficiaryChange();

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(delegator1);
        assertEq(info.beneficiary, rewardWallet);
    }

    function test_providerConfig_beneficiaryTimelockProtection() public {
        vm.prank(delegator1);
        staking.stake(STARTER_MIN);

        address rewardWallet = makeAddr("rewardWallet");

        vm.prank(delegator1);
        staking.initiateBeneficiaryChange(rewardWallet);

        // Cannot execute before timelock.
        vm.prank(delegator1);
        vm.expectRevert();
        staking.executeBeneficiaryChange();
    }

    // ================================================================
    //  5. REWARD CUT VALIDATION
    // ================================================================
    // NOTE: Reward cut configuration (provider sets % to keep) is a planned
    // V2 feature. These tests verify the existing tier multiplier system
    // as a foundation for per-provider reward configuration.

    function test_rewardCutTooHigh_tierMultiplierCannotExceedDesign() public {
        // Tier configs are set by admin. Verify Platinum (1.20x) is the max.
        (,, uint16 platinumMultiplier,,) = staking.tierConfigs(BunkerStaking.Tier.Platinum);
        assertEq(platinumMultiplier, 12000); // 1.20x

        // Starter multiplier is 1.00x.
        (,, uint16 starterMultiplier,,) = staking.tierConfigs(BunkerStaking.Tier.Starter);
        assertEq(starterMultiplier, 10000); // 1.00x
    }

    // ================================================================
    //  6. DELEGATE TO NON-ACCEPTING (inactive provider concept)
    // ================================================================

    function test_delegateToNonAccepting_deactivatedProviderCanReactivate() public {
        vm.prank(delegator1);
        staking.stake(STARTER_MIN);

        // Deactivate.
        vm.prank(delegator1);
        staking.requestUnstake(STARTER_MIN);
        assertFalse(staking.isActiveProvider(delegator1));

        // Reactivate.
        vm.prank(delegator1);
        staking.stake(STARTER_MIN);
        assertTrue(staking.isActiveProvider(delegator1));
    }

    function test_delegateToNonAccepting_belowMinimumReverts() public {
        // Try to stake below Starter minimum.
        vm.prank(delegator1);
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
    //  7. TOTAL DELEGATED TRACKING
    // ================================================================

    function test_totalDelegated_tracksCorrectlyWithMultipleStakers() public {
        vm.prank(provider);
        staking.stake(BRONZE_MIN);

        vm.prank(delegator1);
        staking.stake(STARTER_MIN);

        vm.prank(delegator2);
        staking.stake(STARTER_MIN);

        assertEq(staking.totalStaked(), BRONZE_MIN + STARTER_MIN * 2);

        // Delegator1 unstakes partially.
        vm.prank(delegator1);
        staking.requestUnstake(50_000e18);

        assertEq(staking.totalStaked(), BRONZE_MIN + STARTER_MIN * 2 - 50_000e18);
        assertEq(staking.totalUnbonding(), 50_000e18);
    }

    function test_totalDelegated_decreasesOnUnstake() public {
        vm.prank(delegator1);
        staking.stake(BRONZE_MIN);

        uint256 totalBefore = staking.totalStaked();

        vm.prank(delegator1);
        staking.requestUnstake(200_000e18);

        assertEq(staking.totalStaked(), totalBefore - 200_000e18);
    }

    function test_totalDelegated_decreasesOnSlash() public {
        vm.prank(delegator1);
        staking.stake(BRONZE_MIN);

        uint256 totalBefore = staking.totalStaked();

        vm.prank(slasher);
        staking.slash(delegator1, 100_000e18);

        assertEq(staking.totalStaked(), totalBefore - 100_000e18);
    }

    function test_totalDelegated_correctAfterCompleteUnstake() public {
        vm.prank(delegator1);
        staking.stake(BRONZE_MIN);

        vm.prank(delegator1);
        staking.requestUnstake(200_000e18);

        vm.warp(block.timestamp + UNBONDING_PERIOD);

        vm.prank(delegator1);
        staking.completeUnstake(0);

        assertEq(staking.totalStaked(), BRONZE_MIN - 200_000e18);
        assertEq(staking.totalUnbonding(), 0);
    }

    // ================================================================
    //  8. SLASHING AFFECTS DELEGATORS
    // ================================================================

    function test_slashAffectsDelegator_independentlySlashable() public {
        vm.prank(provider);
        staking.stake(BRONZE_MIN);

        vm.prank(delegator1);
        staking.stake(STARTER_MIN);

        // Slash only the delegator.
        vm.prank(slasher);
        staking.slash(delegator1, 50_000e18);

        // Provider is unaffected.
        BunkerStaking.ProviderInfo memory providerInfo = staking.getProviderInfo(provider);
        assertEq(providerInfo.stakedAmount, uint128(BRONZE_MIN));

        // Delegator is slashed.
        BunkerStaking.ProviderInfo memory delegatorInfo = staking.getProviderInfo(delegator1);
        assertEq(delegatorInfo.stakedAmount, uint128(STARTER_MIN - 50_000e18));
    }

    // ================================================================
    //  9. FUZZ TESTS
    // ================================================================

    function testFuzz_delegateAndUndelegate(uint128 stakeAmount, uint128 unstakeAmount) public {
        vm.assume(uint256(stakeAmount) >= STARTER_MIN && uint256(stakeAmount) <= 5_000_000e18);
        vm.assume(unstakeAmount > 0 && unstakeAmount <= stakeAmount);

        vm.prank(delegator1);
        staking.stake(stakeAmount);

        vm.prank(delegator1);
        staking.requestUnstake(unstakeAmount);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(delegator1);
        assertEq(info.stakedAmount, uint128(uint256(stakeAmount) - uint256(unstakeAmount)));
        assertEq(info.totalUnbonding, unstakeAmount);
    }

    function testFuzz_totalDelegatedConsistency(
        uint256 stake1,
        uint256 stake2,
        uint256 stake3
    ) public {
        stake1 = bound(stake1, STARTER_MIN, 3_000_000e18);
        stake2 = bound(stake2, STARTER_MIN, 3_000_000e18);
        stake3 = bound(stake3, STARTER_MIN, 3_000_000e18);

        vm.prank(delegator1);
        staking.stake(stake1);

        vm.prank(delegator2);
        staking.stake(stake2);

        vm.prank(delegator3);
        staking.stake(stake3);

        assertEq(
            staking.totalStaked(),
            uint256(stake1) + uint256(stake2) + uint256(stake3)
        );
    }
}
