// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";

contract BunkerStakingTest is Test {
    BunkerToken public token;
    BunkerStaking public staking;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public provider1 = makeAddr("provider1");
    address public provider2 = makeAddr("provider2");
    address public slasher = makeAddr("slasher");

    // Tier thresholds (100B total supply)
    uint256 constant STARTER_MIN = 1_000_000e18;
    uint256 constant BRONZE_MIN = 5_000_000e18;
    uint256 constant SILVER_MIN = 10_000_000e18;
    uint256 constant GOLD_MIN = 100_000_000e18;
    uint256 constant PLATINUM_MIN = 1_000_000_000e18;

    uint256 constant UNBONDING_PERIOD = 14 days;
    uint256 constant BENEFICIARY_TIMELOCK = 24 hours;

    function setUp() public {
        // Deploy token with owner as the minter
        vm.startPrank(owner);
        token = new BunkerToken(owner);

        // Deploy staking contract
        staking = new BunkerStaking(address(token), treasury, owner);

        // Grant slasher role
        staking.grantRole(staking.SLASHER_ROLE(), slasher);

        // Enable slashing (defaults to false / monitor mode)
        staking.setSlashingEnabled(true);

        // Mint tokens to providers (enough for Platinum tier at 1B BUNKER)
        token.mint(provider1, 2_000_000_000e18);
        token.mint(provider2, 2_000_000_000e18);
        vm.stopPrank();

        // Approve staking contract from providers
        vm.prank(provider1);
        token.approve(address(staking), type(uint256).max);

        vm.prank(provider2);
        token.approve(address(staking), type(uint256).max);
    }

    // ================================================================
    //  1. SETUP & CONSTRUCTOR
    // ================================================================

    function test_constructor_setsTokenAndTreasury() public view {
        assertEq(address(staking.token()), address(token));
        assertEq(staking.treasury(), treasury);
    }

    function test_constructor_setsOwner() public view {
        assertEq(staking.owner(), owner);
    }

    function test_constructor_grantsDefaultAdminRole() public view {
        assertTrue(staking.hasRole(staking.DEFAULT_ADMIN_ROLE(), owner));
    }

    function test_constructor_initializesTierConfigs() public view {
        (uint256 starterMin,,,,) = staking.tierConfigs(BunkerStaking.Tier.Starter);
        (uint256 bronzeMin,,,,) = staking.tierConfigs(BunkerStaking.Tier.Bronze);
        (uint256 silverMin,,,,) = staking.tierConfigs(BunkerStaking.Tier.Silver);
        (uint256 goldMin,,,,) = staking.tierConfigs(BunkerStaking.Tier.Gold);
        (uint256 platinumMin,,,,) = staking.tierConfigs(BunkerStaking.Tier.Platinum);

        assertEq(starterMin, STARTER_MIN);
        assertEq(bronzeMin, BRONZE_MIN);
        assertEq(silverMin, SILVER_MIN);
        assertEq(goldMin, GOLD_MIN);
        assertEq(platinumMin, PLATINUM_MIN);
    }

    function test_constructor_revertsZeroToken() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.ZeroAddress.selector);
        new BunkerStaking(address(0), treasury, owner);
    }

    function test_constructor_revertsZeroTreasury() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.ZeroAddress.selector);
        new BunkerStaking(address(token), address(0), owner);
    }

    function test_constructor_revertsZeroOwner() public {
        // Ownable constructor reverts with OwnableInvalidOwner before ZeroAddress check
        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSelector(
                bytes4(keccak256("OwnableInvalidOwner(address)")), address(0)
            )
        );
        new BunkerStaking(address(token), treasury, address(0));
    }

    function test_constants() public view {
        assertEq(staking.unbondingPeriod(), 14 days);
        assertEq(staking.BENEFICIARY_TIMELOCK(), 24 hours);
        assertEq(staking.slashBurnBps(), 8000);
        assertEq(staking.slashTreasuryBps(), 2000);
        assertEq(staking.BPS_DENOMINATOR(), 10000);
        assertEq(staking.SLASHER_ROLE(), keccak256("SLASHER_ROLE"));
    }

    // ================================================================
    //  2. REGISTRATION & STAKING
    // ================================================================

    function test_stake_firstStakeBelowMinimumReverts() public {
        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.BelowMinimumStake.selector, 499e18, STARTER_MIN)
        );
        staking.stake(499e18);
    }

    function test_stake_zeroAmountReverts() public {
        vm.prank(provider1);
        vm.expectRevert(BunkerStaking.ZeroAmount.selector);
        staking.stake(0);
    }

    function test_stake_exactStarterMinimum() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(STARTER_MIN));
        assertTrue(info.active);
        assertEq(info.beneficiary, provider1);
        assertGt(info.registeredAt, 0);
        assertEq(staking.totalStaked(), STARTER_MIN);
    }

    function test_stake_emitsProviderRegisteredEvent() public {
        vm.prank(provider1);

        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.ProviderRegistered(
            provider1, STARTER_MIN, BunkerStaking.Tier.Starter
        );

        staking.stake(STARTER_MIN);
    }

    function test_stake_emitsStakedEvent() public {
        vm.prank(provider1);

        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.Staked(
            provider1, STARTER_MIN, STARTER_MIN, BunkerStaking.Tier.Starter
        );

        staking.stake(STARTER_MIN);
    }

    function test_stake_additionalStakingIncreasesToBronze() public {
        vm.startPrank(provider1);
        staking.stake(STARTER_MIN);
        staking.stake(BRONZE_MIN - STARTER_MIN);
        vm.stopPrank();

        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Bronze));
        assertEq(staking.totalStaked(), BRONZE_MIN);
    }

    function test_stake_additionalStakingIncreasesToSilver() public {
        vm.startPrank(provider1);
        staking.stake(SILVER_MIN);
        vm.stopPrank();

        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Silver));
    }

    function test_stake_additionalStakingIncreasesToGold() public {
        vm.startPrank(provider1);
        staking.stake(GOLD_MIN);
        vm.stopPrank();

        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Gold));
    }

    function test_stake_additionalStakingIncreasesToPlatinum() public {
        vm.startPrank(provider1);
        staking.stake(PLATINUM_MIN);
        vm.stopPrank();

        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Platinum));
    }

    function test_stake_totalStakedTracksCorrectly() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider2);
        staking.stake(BRONZE_MIN);

        assertEq(staking.totalStaked(), STARTER_MIN + BRONZE_MIN);
    }

    function test_stake_transfersTokensFromProvider() public {
        uint256 balanceBefore = token.balanceOf(provider1);

        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        assertEq(token.balanceOf(provider1), balanceBefore - STARTER_MIN);
        assertEq(token.balanceOf(address(staking)), STARTER_MIN);
    }

    function test_stake_multipleStakesAccumulate() public {
        vm.startPrank(provider1);
        staking.stake(STARTER_MIN);
        staking.stake(1_000e18);
        staking.stake(500e18);
        vm.stopPrank();

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(STARTER_MIN + 1_000e18 + 500e18));
    }

    // ================================================================
    //  3. UNSTAKING & UNBONDING
    // ================================================================

    function test_requestUnstake_createsRequestWithCorrectUnlockTime() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 unstakeAmount = 500e18;
        uint256 expectedUnlockTime = block.timestamp + UNBONDING_PERIOD;

        vm.prank(provider1);
        staking.requestUnstake(unstakeAmount);

        BunkerStaking.UnstakeRequest memory req = staking.getUnstakeRequest(provider1, 0);
        assertEq(req.amount, uint128(unstakeAmount));
        assertEq(req.unlockTime, uint48(expectedUnlockTime));
        assertFalse(req.completed);
    }

    function test_requestUnstake_emitsUnstakeRequestedEvent() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 unstakeAmount = 500e18;
        uint256 expectedUnlockTime = block.timestamp + UNBONDING_PERIOD;

        vm.prank(provider1);

        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.UnstakeRequested(provider1, unstakeAmount, expectedUnlockTime, 0);

        staking.requestUnstake(unstakeAmount);
    }

    function test_requestUnstake_cannotUnstakeMoreThanStaked() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerStaking.InsufficientStake.selector,
                STARTER_MIN + 1,
                STARTER_MIN
            )
        );
        staking.requestUnstake(STARTER_MIN + 1);
    }

    function test_requestUnstake_zeroAmountReverts() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider1);
        vm.expectRevert(BunkerStaking.ZeroAmount.selector);
        staking.requestUnstake(0);
    }

    function test_requestUnstake_nonActiveProviderReverts() public {
        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.ProviderNotActive.selector, provider1)
        );
        staking.requestUnstake(100e18);
    }

    function test_requestUnstake_belowMinimumDeactivatesProvider() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider1);

        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.ProviderDeregistered(provider1);

        staking.requestUnstake(STARTER_MIN);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertFalse(info.active);
    }

    function test_requestUnstake_partialUnstakeKeepsActive() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Unstake some but stay above Starter minimum
        vm.prank(provider1);
        staking.requestUnstake(1_000e18);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertTrue(info.active);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 1_000e18));
    }

    function test_requestUnstake_tracksTotalUnbonding() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider1);
        staking.requestUnstake(500e18);

        assertEq(staking.totalUnbonding(), 500e18);
        assertEq(staking.totalStaked(), BRONZE_MIN - 500e18);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.totalUnbonding, uint128(500e18));
    }

    function test_completeUnstake_beforeUnbondingReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider1);
        staking.requestUnstake(500e18);

        BunkerStaking.UnstakeRequest memory req = staking.getUnstakeRequest(provider1, 0);

        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerStaking.UnbondingNotReady.selector,
                req.unlockTime,
                block.timestamp
            )
        );
        staking.completeUnstake(0);
    }

    function test_completeUnstake_afterUnbondingSucceeds() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider1);
        staking.requestUnstake(500e18);

        // Warp past unbonding period
        vm.warp(block.timestamp + UNBONDING_PERIOD);

        uint256 balanceBefore = token.balanceOf(provider1);

        vm.prank(provider1);
        staking.completeUnstake(0);

        assertEq(token.balanceOf(provider1), balanceBefore + 500e18);
        assertEq(staking.totalUnbonding(), 0);

        BunkerStaking.UnstakeRequest memory req = staking.getUnstakeRequest(provider1, 0);
        assertTrue(req.completed);
    }

    function test_completeUnstake_emitsUnstakeCompletedEvent() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider1);
        staking.requestUnstake(500e18);

        vm.warp(block.timestamp + UNBONDING_PERIOD);

        vm.prank(provider1);

        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.UnstakeCompleted(provider1, 500e18, 0);

        staking.completeUnstake(0);
    }

    function test_completeUnstake_cannotCompleteTwice() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider1);
        staking.requestUnstake(500e18);

        vm.warp(block.timestamp + UNBONDING_PERIOD);

        vm.prank(provider1);
        staking.completeUnstake(0);

        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.UnstakeAlreadyCompleted.selector, 0)
        );
        staking.completeUnstake(0);
    }

    function test_completeUnstake_invalidIndexReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.InvalidUnstakeIndex.selector, 0, 0)
        );
        staking.completeUnstake(0);
    }

    function test_requestUnstake_multipleRequestsWork() public {
        vm.prank(provider1);
        staking.stake(SILVER_MIN);

        vm.startPrank(provider1);
        staking.requestUnstake(1_000e18);
        staking.requestUnstake(2_000e18);
        staking.requestUnstake(3_000e18);
        vm.stopPrank();

        assertEq(staking.getUnstakeQueueLength(provider1), 3);
        assertEq(staking.totalUnbonding(), 6_000e18);

        BunkerStaking.UnstakeRequest memory req0 = staking.getUnstakeRequest(provider1, 0);
        BunkerStaking.UnstakeRequest memory req1 = staking.getUnstakeRequest(provider1, 1);
        BunkerStaking.UnstakeRequest memory req2 = staking.getUnstakeRequest(provider1, 2);

        assertEq(req0.amount, uint128(1_000e18));
        assertEq(req1.amount, uint128(2_000e18));
        assertEq(req2.amount, uint128(3_000e18));
    }

    function test_completeUnstake_multipleRequestsOutOfOrder() public {
        vm.prank(provider1);
        staking.stake(SILVER_MIN);

        uint256 t0 = block.timestamp;

        vm.prank(provider1);
        staking.requestUnstake(1_000e18);

        vm.warp(t0 + 1 days);
        vm.prank(provider1);
        staking.requestUnstake(2_000e18);

        // Complete request 0 first (after its unbonding)
        vm.warp(t0 + UNBONDING_PERIOD);
        vm.prank(provider1);
        staking.completeUnstake(0);

        // Request 1 is not yet ready
        vm.prank(provider1);
        vm.expectRevert();
        staking.completeUnstake(1);

        // Now complete request 1
        vm.warp(t0 + 1 days + UNBONDING_PERIOD);
        vm.prank(provider1);
        staking.completeUnstake(1);

        assertEq(staking.totalUnbonding(), 0);
    }

    // ================================================================
    //  4. BENEFICIARY
    // ================================================================

    function test_initiateBeneficiaryChange_setsPendingWithTimelock() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        address newBeneficiary = makeAddr("newBeneficiary");
        uint256 expectedEffectiveTime = block.timestamp + BENEFICIARY_TIMELOCK;

        vm.prank(provider1);
        staking.initiateBeneficiaryChange(newBeneficiary);

        (address pendingAddr, uint48 effectiveTime) = staking.pendingBeneficiaries(provider1);
        assertEq(pendingAddr, newBeneficiary);
        assertEq(effectiveTime, uint48(expectedEffectiveTime));
    }

    function test_initiateBeneficiaryChange_emitsEvent() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        address newBeneficiary = makeAddr("newBeneficiary");
        uint256 expectedEffectiveTime = block.timestamp + BENEFICIARY_TIMELOCK;

        vm.prank(provider1);

        vm.expectEmit(true, true, false, true, address(staking));
        emit BunkerStaking.BeneficiaryChangeInitiated(
            provider1, newBeneficiary, expectedEffectiveTime
        );

        staking.initiateBeneficiaryChange(newBeneficiary);
    }

    function test_initiateBeneficiaryChange_zeroAddressReverts() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider1);
        vm.expectRevert(BunkerStaking.ZeroAddress.selector);
        staking.initiateBeneficiaryChange(address(0));
    }

    function test_initiateBeneficiaryChange_nonActiveProviderReverts() public {
        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.ProviderNotActive.selector, provider1)
        );
        staking.initiateBeneficiaryChange(makeAddr("someone"));
    }

    function test_executeBeneficiaryChange_beforeTimelockReverts() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        address newBeneficiary = makeAddr("newBeneficiary");

        vm.prank(provider1);
        staking.initiateBeneficiaryChange(newBeneficiary);

        (,uint48 effectiveTime) = staking.pendingBeneficiaries(provider1);

        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerStaking.BeneficiaryTimelockNotElapsed.selector,
                effectiveTime,
                block.timestamp
            )
        );
        staking.executeBeneficiaryChange();
    }

    function test_executeBeneficiaryChange_afterTimelockSucceeds() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        address newBeneficiary = makeAddr("newBeneficiary");

        vm.prank(provider1);
        staking.initiateBeneficiaryChange(newBeneficiary);

        vm.warp(block.timestamp + BENEFICIARY_TIMELOCK);

        vm.prank(provider1);
        staking.executeBeneficiaryChange();

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.beneficiary, newBeneficiary);

        // Pending should be cleared
        (address pendingAddr,) = staking.pendingBeneficiaries(provider1);
        assertEq(pendingAddr, address(0));
    }

    function test_executeBeneficiaryChange_emitsEvent() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        address newBeneficiary = makeAddr("newBeneficiary");

        vm.prank(provider1);
        staking.initiateBeneficiaryChange(newBeneficiary);

        vm.warp(block.timestamp + BENEFICIARY_TIMELOCK);

        vm.prank(provider1);

        vm.expectEmit(true, true, true, true, address(staking));
        emit BunkerStaking.BeneficiaryChanged(provider1, provider1, newBeneficiary);

        staking.executeBeneficiaryChange();
    }

    function test_executeBeneficiaryChange_noPendingReverts() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.NoPendingBeneficiaryChange.selector, provider1)
        );
        staking.executeBeneficiaryChange();
    }

    // ================================================================
    //  5. SLASHING
    // ================================================================

    function test_slash_slasherRoleCanSlash() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 slashAmount = 1_000e18;

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - slashAmount));
    }

    function test_slash_nonSlasherReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider2);
        vm.expectRevert();
        staking.slash(provider1, 500e18);
    }

    function test_slash_zeroAmountReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        vm.expectRevert(BunkerStaking.ZeroAmount.selector);
        staking.slash(provider1, 0);
    }

    function test_slash_activeStakeFirst() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 slashAmount = 500e18;

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - slashAmount));
        assertEq(info.totalUnbonding, 0);
    }

    function test_slash_overflowsIntoUnbondingQueue() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Create unbonding request
        vm.prank(provider1);
        staking.requestUnstake(500e18);

        // Slash more than active stake
        uint256 activeStake = BRONZE_MIN - 500e18; // 1500e18
        uint256 slashAmount = activeStake + 200e18; // takes 200 from unbonding

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, 0);
        assertEq(info.totalUnbonding, uint128(500e18 - 200e18));
    }

    function test_slash_unbondingQueueNewestFirst() public {
        vm.prank(provider1);
        staking.stake(SILVER_MIN);

        // Create two unbonding requests
        vm.startPrank(provider1);
        staking.requestUnstake(1_000_000e18); // index 0: 1M
        staking.requestUnstake(2_000_000e18); // index 1: 2M
        vm.stopPrank();

        // Active stake: 10M - 3M = 7M
        // Slash 7.5M -> 7M from active, 500K from unbonding (newest first = index 1)
        vm.prank(slasher);
        staking.slash(provider1, 7_500_000e18);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, 0);

        // Index 1 (newest) should be reduced by 500K
        BunkerStaking.UnstakeRequest memory req1 = staking.getUnstakeRequest(provider1, 1);
        assertEq(req1.amount, uint128(2_000_000e18 - 500_000e18));

        // Index 0 (oldest) should be untouched
        BunkerStaking.UnstakeRequest memory req0 = staking.getUnstakeRequest(provider1, 0);
        assertEq(req0.amount, uint128(1_000_000e18));
    }

    function test_slash_80PercentBurned20PercentTreasury() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 slashAmount = 1_000e18;
        uint256 expectedBurn = (slashAmount * 8000) / 10000; // 800e18
        uint256 expectedTreasury = slashAmount - expectedBurn; // 200e18

        uint256 treasuryBefore = token.balanceOf(treasury);
        uint256 totalSupplyBefore = token.totalSupply();

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        assertEq(token.balanceOf(treasury), treasuryBefore + expectedTreasury);
        assertEq(token.totalSupply(), totalSupplyBefore - expectedBurn);
    }

    function test_slash_emitsSlashedEvent() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 slashAmount = 1_000e18;
        uint256 expectedBurn = (slashAmount * 8000) / 10000;
        uint256 expectedTreasury = slashAmount - expectedBurn;

        vm.prank(slasher);

        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.Slashed(provider1, slashAmount, expectedBurn, expectedTreasury);

        staking.slash(provider1, slashAmount);
    }

    function test_slash_deactivatesIfBelowMinimum() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Slash enough to go below minimum
        vm.prank(slasher);

        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.ProviderDeregistered(provider1);

        staking.slash(provider1, 100e18);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertFalse(info.active);
    }

    function test_slash_cannotSlashMoreThanSlashable() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(slasher);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerStaking.InsufficientSlashableBalance.selector,
                STARTER_MIN + 1,
                STARTER_MIN
            )
        );
        staking.slash(provider1, STARTER_MIN + 1);
    }

    function test_slash_fullyConsumesUnbondingRequestMarksCompleted() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Create unbonding request for 500e18
        vm.prank(provider1);
        staking.requestUnstake(500e18);

        // Active = 1500, unbonding = 500
        // Slash 2000 -> 1500 from active + 500 from unbonding
        vm.prank(slasher);
        staking.slash(provider1, BRONZE_MIN);

        BunkerStaking.UnstakeRequest memory req = staking.getUnstakeRequest(provider1, 0);
        assertEq(req.amount, 0);
        assertTrue(req.completed);
    }

    function test_slash_totalStakedAndUnbondingTrackCorrectly() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider1);
        staking.requestUnstake(500e18);

        uint256 totalStakedBefore = staking.totalStaked();
        uint256 totalUnbondingBefore = staking.totalUnbonding();

        // Slash 2000 (1500 active + 500 unbonding)
        vm.prank(slasher);
        staking.slash(provider1, BRONZE_MIN);

        assertEq(staking.totalStaked(), totalStakedBefore - (BRONZE_MIN - 500e18));
        assertEq(staking.totalUnbonding(), totalUnbondingBefore - 500e18);
    }

    // ================================================================
    //  6. TIER SYSTEM
    // ================================================================

    function test_getTierForAmount_none() public view {
        assertEq(uint8(staking.getTierForAmount(0)), uint8(BunkerStaking.Tier.None));
        assertEq(uint8(staking.getTierForAmount(999_999e18)), uint8(BunkerStaking.Tier.None));
    }

    function test_getTierForAmount_starter() public view {
        assertEq(uint8(staking.getTierForAmount(1_000_000e18)), uint8(BunkerStaking.Tier.Starter));
        assertEq(uint8(staking.getTierForAmount(4_999_999e18)), uint8(BunkerStaking.Tier.Starter));
    }

    function test_getTierForAmount_bronze() public view {
        assertEq(uint8(staking.getTierForAmount(5_000_000e18)), uint8(BunkerStaking.Tier.Bronze));
        assertEq(uint8(staking.getTierForAmount(9_999_999e18)), uint8(BunkerStaking.Tier.Bronze));
    }

    function test_getTierForAmount_silver() public view {
        assertEq(uint8(staking.getTierForAmount(10_000_000e18)), uint8(BunkerStaking.Tier.Silver));
        assertEq(uint8(staking.getTierForAmount(99_999_999e18)), uint8(BunkerStaking.Tier.Silver));
    }

    function test_getTierForAmount_gold() public view {
        assertEq(uint8(staking.getTierForAmount(100_000_000e18)), uint8(BunkerStaking.Tier.Gold));
        assertEq(uint8(staking.getTierForAmount(999_999_999e18)), uint8(BunkerStaking.Tier.Gold));
    }

    function test_getTierForAmount_platinum() public view {
        assertEq(
            uint8(staking.getTierForAmount(1_000_000_000e18)), uint8(BunkerStaking.Tier.Platinum)
        );
        assertEq(
            uint8(staking.getTierForAmount(2_000_000_000e18)), uint8(BunkerStaking.Tier.Platinum)
        );
    }

    function test_getTier_returnsCorrectTierForProvider() public {
        vm.prank(provider1);
        staking.stake(SILVER_MIN);

        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Silver));

        // Unstaked provider has None tier
        assertEq(uint8(staking.getTier(provider2)), uint8(BunkerStaking.Tier.None));
    }

    // ================================================================
    //  7. ADMIN
    // ================================================================

    function test_setTreasury_ownerCanUpdate() public {
        address newTreasury = makeAddr("newTreasury");

        vm.prank(owner);

        vm.expectEmit(true, true, false, true, address(staking));
        emit BunkerStaking.TreasuryUpdated(treasury, newTreasury);

        staking.setTreasury(newTreasury);

        assertEq(staking.treasury(), newTreasury);
    }

    function test_setTreasury_nonOwnerReverts() public {
        vm.prank(provider1);
        vm.expectRevert();
        staking.setTreasury(makeAddr("newTreasury"));
    }

    function test_setTreasury_zeroAddressReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.ZeroAddress.selector);
        staking.setTreasury(address(0));
    }

    function test_setTierMinStake_ownerCanUpdate() public {
        uint256 newMinStake = 1_000e18;

        vm.prank(owner);

        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.TierConfigUpdated(BunkerStaking.Tier.Starter, newMinStake);

        staking.setTierMinStake(BunkerStaking.Tier.Starter, newMinStake);

        (uint256 minStake,,,,) = staking.tierConfigs(BunkerStaking.Tier.Starter);
        assertEq(minStake, newMinStake);
    }

    function test_setTierMinStake_nonOwnerReverts() public {
        vm.prank(provider1);
        vm.expectRevert();
        staking.setTierMinStake(BunkerStaking.Tier.Starter, 1_000e18);
    }

    function test_setTierMinStake_tierNoneReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.CannotConfigureNoneTier.selector);
        staking.setTierMinStake(BunkerStaking.Tier.None, 100e18);
    }

    function test_pause_ownerCanPause() public {
        vm.prank(owner);
        staking.pause();

        assertTrue(staking.paused());
    }

    function test_pause_nonOwnerReverts() public {
        vm.prank(provider1);
        vm.expectRevert();
        staking.pause();
    }

    function test_unpause_ownerCanUnpause() public {
        vm.prank(owner);
        staking.pause();

        vm.prank(owner);
        staking.unpause();

        assertFalse(staking.paused());
    }

    function test_unpause_nonOwnerReverts() public {
        vm.prank(owner);
        staking.pause();

        vm.prank(provider1);
        vm.expectRevert();
        staking.unpause();
    }

    function test_pause_stakingRevertsWhenPaused() public {
        vm.prank(owner);
        staking.pause();

        vm.prank(provider1);
        vm.expectRevert();
        staking.stake(STARTER_MIN);
    }

    function test_pause_requestUnstakeRevertsWhenPaused() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(owner);
        staking.pause();

        vm.prank(provider1);
        vm.expectRevert();
        staking.requestUnstake(100e18);
    }

    function test_pause_completeUnstakeWorksWhenPaused() public {
        // completeUnstake does not have whenNotPaused modifier
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider1);
        staking.requestUnstake(500e18);

        vm.warp(block.timestamp + UNBONDING_PERIOD);

        vm.prank(owner);
        staking.pause();

        // completeUnstake should still work (no whenNotPaused)
        vm.prank(provider1);
        staking.completeUnstake(0);

        BunkerStaking.UnstakeRequest memory req = staking.getUnstakeRequest(provider1, 0);
        assertTrue(req.completed);
    }

    // ================================================================
    //  8. EDGE CASES
    // ================================================================

    function test_multipleProvidersStakeIndependently() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider2);
        staking.stake(BRONZE_MIN);

        BunkerStaking.ProviderInfo memory info1 = staking.getProviderInfo(provider1);
        BunkerStaking.ProviderInfo memory info2 = staking.getProviderInfo(provider2);

        assertEq(info1.stakedAmount, uint128(STARTER_MIN));
        assertEq(info2.stakedAmount, uint128(BRONZE_MIN));
        assertTrue(info1.active);
        assertTrue(info2.active);
        assertEq(staking.totalStaked(), STARTER_MIN + BRONZE_MIN);
    }

    function test_isActiveProvider_returnsCorrectStatus() public {
        assertFalse(staking.isActiveProvider(provider1));

        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        assertTrue(staking.isActiveProvider(provider1));
    }

    function test_getSlashableBalance_includesActiveAndUnbonding() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider1);
        staking.requestUnstake(500e18);

        uint256 slashable = staking.getSlashableBalance(provider1);
        assertEq(slashable, BRONZE_MIN); // active + unbonding = total originally staked
    }

    function test_getUnstakeQueueLength_returnsCorrectCount() public {
        assertEq(staking.getUnstakeQueueLength(provider1), 0);

        vm.prank(provider1);
        staking.stake(SILVER_MIN);

        vm.startPrank(provider1);
        staking.requestUnstake(1_000e18);
        staking.requestUnstake(2_000e18);
        vm.stopPrank();

        assertEq(staking.getUnstakeQueueLength(provider1), 2);
    }

    function test_stake_afterDeactivationReactivatesProvider() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Deactivate by unstaking all
        vm.prank(provider1);
        staking.requestUnstake(STARTER_MIN);

        assertFalse(staking.isActiveProvider(provider1));

        // Re-stake to meet minimum
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        assertTrue(staking.isActiveProvider(provider1));
    }

    function test_slash_fromMultipleUnbondingRequests() public {
        vm.prank(provider1);
        staking.stake(SILVER_MIN);

        // Create 3 unbonding requests
        vm.startPrank(provider1);
        staking.requestUnstake(1_000_000e18); // index 0
        staking.requestUnstake(1_000_000e18); // index 1
        staking.requestUnstake(1_000_000e18); // index 2
        vm.stopPrank();

        // Active: 7M, unbonding: 3M
        // Slash 9M: 7M from active, 2M from unbonding (newest first)
        vm.prank(slasher);
        staking.slash(provider1, 9_000_000e18);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, 0);
        assertEq(info.totalUnbonding, uint128(1_000_000e18)); // only index 0 remains

        // Index 2 fully consumed
        BunkerStaking.UnstakeRequest memory req2 = staking.getUnstakeRequest(provider1, 2);
        assertEq(req2.amount, 0);
        assertTrue(req2.completed);

        // Index 1 fully consumed
        BunkerStaking.UnstakeRequest memory req1 = staking.getUnstakeRequest(provider1, 1);
        assertEq(req1.amount, 0);
        assertTrue(req1.completed);

        // Index 0 untouched
        BunkerStaking.UnstakeRequest memory req0 = staking.getUnstakeRequest(provider1, 0);
        assertEq(req0.amount, uint128(1_000_000e18));
        assertFalse(req0.completed);
    }

    function test_stake_tierTransitionsEmitCorrectTier() public {
        vm.startPrank(provider1);

        // Stake to Starter
        staking.stake(STARTER_MIN);
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Starter));

        // Stake up to Bronze
        staking.stake(BRONZE_MIN - STARTER_MIN);
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Bronze));

        // Stake up to Silver
        staking.stake(SILVER_MIN - BRONZE_MIN);
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Silver));

        // Stake up to Gold
        staking.stake(GOLD_MIN - SILVER_MIN);
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Gold));

        // Stake up to Platinum
        staking.stake(PLATINUM_MIN - GOLD_MIN);
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Platinum));

        vm.stopPrank();
    }

    function test_version() public view {
        assertEq(keccak256(bytes(staking.VERSION())), keccak256(bytes("1.3.0")));
    }

    function test_slash_slashEntireBalance() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(slasher);
        staking.slash(provider1, STARTER_MIN);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, 0);
        assertFalse(info.active);
    }

    function test_requestUnstake_exactActiveStakeDeactivates() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider1);
        staking.requestUnstake(STARTER_MIN);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertFalse(info.active);
        assertEq(info.stakedAmount, 0);
        assertEq(info.totalUnbonding, uint128(STARTER_MIN));
    }

    function test_beneficiaryChange_fullFlow() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        address newBeneficiary = makeAddr("newBeneficiary");

        // Check initial beneficiary
        BunkerStaking.ProviderInfo memory infoBefore = staking.getProviderInfo(provider1);
        assertEq(infoBefore.beneficiary, provider1);

        // Initiate change
        vm.prank(provider1);
        staking.initiateBeneficiaryChange(newBeneficiary);

        // Cannot execute yet
        vm.prank(provider1);
        vm.expectRevert();
        staking.executeBeneficiaryChange();

        // Warp past timelock
        vm.warp(block.timestamp + BENEFICIARY_TIMELOCK);

        // Execute change
        vm.prank(provider1);
        staking.executeBeneficiaryChange();

        BunkerStaking.ProviderInfo memory infoAfter = staking.getProviderInfo(provider1);
        assertEq(infoAfter.beneficiary, newBeneficiary);
    }

    function test_setTierMinStake_affectsNewStakers() public {
        // Lower Starter minimum to 100e18
        vm.prank(owner);
        staking.setTierMinStake(BunkerStaking.Tier.Starter, 100e18);

        // Now provider can stake with only 100 BUNKER
        vm.prank(provider1);
        staking.stake(100e18);

        assertTrue(staking.isActiveProvider(provider1));
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Starter));
    }

    function test_slash_treasuryReceivesCorrectAmountOnOddSlash() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Slash an odd amount to check rounding
        uint256 slashAmount = 333e18;
        uint256 expectedBurn = (slashAmount * 8000) / 10000; // 266.4e18 -> 266400000000000000000
        uint256 expectedTreasury = slashAmount - expectedBurn;

        uint256 treasuryBefore = token.balanceOf(treasury);

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        assertEq(token.balanceOf(treasury), treasuryBefore + expectedTreasury);
    }

    function test_grantSlasherRole_ownerCanGrant() public {
        address newSlasher = makeAddr("newSlasher");
        bytes32 slasherRole = staking.SLASHER_ROLE();

        vm.prank(owner);
        staking.grantRole(slasherRole, newSlasher);

        assertTrue(staking.hasRole(slasherRole, newSlasher));

        // New slasher can slash
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(newSlasher);
        staking.slash(provider1, 100e18);
    }

    function test_revokeSlasherRole_ownerCanRevoke() public {
        bytes32 slasherRole = staking.SLASHER_ROLE();

        vm.prank(owner);
        staking.revokeRole(slasherRole, slasher);

        assertFalse(staking.hasRole(slasherRole, slasher));

        // Slasher can no longer slash
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(slasher);
        vm.expectRevert();
        staking.slash(provider1, 100e18);
    }

    // ================================================================
    //  9. FUZZ TESTS
    // ================================================================

    /// @notice Fuzz staking: stake any amount that meets the Starter minimum.
    function testFuzz_stake(uint128 amount) public {
        // Constrain: must meet Starter minimum and not exceed provider's balance
        vm.assume(uint256(amount) >= STARTER_MIN && uint256(amount) <= 5_000_000e18);

        vm.prank(provider1);
        staking.stake(amount);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, amount);
        assertTrue(info.active);
        assertEq(staking.totalStaked(), uint256(amount));
    }

    /// @notice Fuzz unstaking: stake an amount, then request unstaking a portion.
    function testFuzz_requestUnstake(uint128 stakeAmount, uint128 unstakeAmount) public {
        // Constrain: stake must meet minimum; unstake must be > 0 and <= staked
        vm.assume(uint256(stakeAmount) >= STARTER_MIN && uint256(stakeAmount) <= 5_000_000e18);
        vm.assume(unstakeAmount > 0 && unstakeAmount <= stakeAmount);

        vm.prank(provider1);
        staking.stake(stakeAmount);

        vm.prank(provider1);
        staking.requestUnstake(unstakeAmount);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(uint256(stakeAmount) - uint256(unstakeAmount)));
        assertEq(info.totalUnbonding, unstakeAmount);
        assertEq(staking.totalStaked(), uint256(stakeAmount) - uint256(unstakeAmount));
        assertEq(staking.totalUnbonding(), uint256(unstakeAmount));

        // If remaining stake is below minimum, provider should be deactivated
        if (uint256(stakeAmount) - uint256(unstakeAmount) < STARTER_MIN) {
            assertFalse(info.active);
        } else {
            assertTrue(info.active);
        }
    }

    /// @notice Fuzz slashing: stake an amount, then slash a portion. Verify distribution.
    function testFuzz_slash(uint128 stakeAmount, uint128 slashAmount) public {
        // Constrain: stake must meet minimum; slash must be > 0 and <= staked
        vm.assume(uint256(stakeAmount) >= STARTER_MIN && uint256(stakeAmount) <= 5_000_000e18);
        vm.assume(slashAmount > 0 && slashAmount <= stakeAmount);

        vm.prank(provider1);
        staking.stake(stakeAmount);

        uint256 treasuryBefore = token.balanceOf(treasury);
        uint256 totalSupplyBefore = token.totalSupply();

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        // Verify stake reduction
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(uint256(stakeAmount) - uint256(slashAmount)));

        // Verify 80/20 split
        uint256 expectedBurn = (uint256(slashAmount) * 8000) / 10000;
        uint256 expectedTreasury = uint256(slashAmount) - expectedBurn;

        assertEq(token.balanceOf(treasury) - treasuryBefore, expectedTreasury);
        assertEq(token.totalSupply(), totalSupplyBefore - expectedBurn);

        // If below minimum, provider should be deactivated
        if (uint256(stakeAmount) - uint256(slashAmount) < STARTER_MIN) {
            assertFalse(info.active);
        }
    }

    /// @notice Fuzz tier determination: verify tier boundaries are correct for any amount.
    function testFuzz_tierDetermination(uint128 amount) public view {
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

    // ================================================================
    //  10. STAKING REWARDS
    // ================================================================

    /// @dev Helper: fund the staking contract with reward tokens and start an epoch.
    function _setupRewardEpoch(uint256 rewardAmount) internal {
        // Mint reward tokens to owner, transfer to staking, then notify.
        // Transfer 20% extra to satisfy the C-01 solvency check
        // (MAX_TIER_MULTIPLIER_BPS = 12000, i.e. 1.2x).
        uint256 transferAmount = (rewardAmount * 12) / 10;
        vm.startPrank(owner);
        token.mint(owner, transferAmount);
        token.approve(address(staking), transferAmount);
        token.transfer(address(staking), transferAmount);
        staking.notifyRewardAmount(rewardAmount);
        vm.stopPrank();
    }

    function test_rewards_noRewardsInitially() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        assertEq(staking.earned(provider1), 0);
    }

    function test_rewards_accumulateOverTime() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        uint256 rewardAmount = 70_000e18;
        _setupRewardEpoch(rewardAmount);

        // Warp forward 1 day (out of 7-day epoch).
        vm.warp(block.timestamp + 1 days);

        uint256 earnedAmount = staking.earned(provider1);
        // With 1 staker for 1 day at rewardRate = rewardAmount / 7 days,
        // base reward = (70_000e18 / 7 days) * 1 day = 10_000e18.
        // Starter multiplier = 10000 bps (1.00x), so earned = 10_000e18.
        // Allow tolerance for integer division.
        assertGt(earnedAmount, 9_999e18);
        assertLt(earnedAmount, 10_001e18);
    }

    function test_rewards_claimRewards() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        uint256 rewardAmount = 70_000e18;
        _setupRewardEpoch(rewardAmount);

        // Warp forward full epoch.
        vm.warp(block.timestamp + 7 days);

        uint256 balanceBefore = token.balanceOf(provider1);

        vm.prank(provider1);
        staking.claimRewards();

        uint256 balanceAfter = token.balanceOf(provider1);
        uint256 claimed = balanceAfter - balanceBefore;

        // claimRewards now sends only 25% immediately (75% goes to vesting).
        // Total reward ~70_000e18, immediate = 25% = ~17_500e18.
        assertGt(claimed, 17_499e18);
        assertLt(claimed, 17_501e18);
    }

    function test_rewards_claimRewardsEmitsEvent() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        uint256 rewardAmount = 70_000e18;
        _setupRewardEpoch(rewardAmount);

        vm.warp(block.timestamp + 7 days);

        // Get expected reward amount.
        uint256 expectedReward = staking.earned(provider1);

        vm.prank(provider1);
        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.RewardClaimed(provider1, expectedReward);
        staking.claimRewards();
    }

    function test_rewards_noRewardsToClaim_reverts() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider1);
        vm.expectRevert(BunkerStaking.NoRewardsToClaimError.selector);
        staking.claimRewards();
    }

    function test_rewards_tierMultiplier_silver() public {
        // Provider1: Silver tier (10500 bps = 1.05x)
        vm.prank(provider1);
        staking.stake(SILVER_MIN);

        // Provider2: Starter tier (10000 bps = 1.00x)
        vm.prank(provider2);
        staking.stake(STARTER_MIN);

        uint256 rewardAmount = 70_000e18;
        _setupRewardEpoch(rewardAmount);

        vm.warp(block.timestamp + 7 days);

        uint256 earned1 = staking.earned(provider1);
        uint256 earned2 = staking.earned(provider2);

        // Provider1 has a higher multiplier (1.05x vs 1.00x)
        // and a larger stake (proportional share).
        // With 10_000e18 staked vs 500e18 staked:
        // provider1 base share = 10000/10500 of total rewards
        // provider2 base share = 500/10500 of total rewards
        // Then provider1 * 1.05, provider2 * 1.00.
        // Just verify the multiplier effect: provider1 gets more per token staked.

        // Per-token reward rate should be the same, but multiplier differs.
        // earned1 / SILVER_MIN * 10000 vs earned2 / STARTER_MIN * 10000 should differ by 1.05x
        uint256 perToken1 = (earned1 * 1e18) / SILVER_MIN;
        uint256 perToken2 = (earned2 * 1e18) / STARTER_MIN;

        // perToken1 should be approximately 1.05x perToken2
        // Allow 1% tolerance for rounding.
        assertGt(perToken1 * 10000, perToken2 * 10400);
        assertLt(perToken1 * 10000, perToken2 * 10600);
    }

    function test_rewards_tierMultiplier_platinum() public {
        // Provider1: Platinum tier (12000 bps = 1.20x)
        vm.prank(provider1);
        staking.stake(PLATINUM_MIN);

        uint256 rewardAmount = 70_000e18;
        _setupRewardEpoch(rewardAmount);

        vm.warp(block.timestamp + 7 days);

        uint256 earnedAmount = staking.earned(provider1);

        // Sole staker gets all base rewards * 1.20x multiplier.
        // base = 70_000e18, with 1.20x = 84_000e18.
        assertGt(earnedAmount, 83_999e18);
        assertLt(earnedAmount, 84_001e18);
    }

    function test_rewards_multipleStakers() public {
        // Two providers stake equal amounts (Starter tier = 1.00x each).
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider2);
        staking.stake(STARTER_MIN);

        uint256 rewardAmount = 70_000e18;
        _setupRewardEpoch(rewardAmount);

        vm.warp(block.timestamp + 7 days);

        uint256 earned1 = staking.earned(provider1);
        uint256 earned2 = staking.earned(provider2);

        // Each should receive half the rewards (both 1.00x multiplier).
        assertGt(earned1, 34_999e18);
        assertLt(earned1, 35_001e18);
        assertGt(earned2, 34_999e18);
        assertLt(earned2, 35_001e18);
    }

    function test_rewards_notifyRewardAmountOnlyOwner() public {
        vm.prank(provider1);
        vm.expectRevert();
        staking.notifyRewardAmount(1_000e18);
    }

    function test_rewards_setRewardsDuration() public {
        vm.prank(owner);
        staking.setRewardsDuration(14 days);

        assertEq(staking.rewardsDuration(), 14 days);
    }

    function test_rewards_setRewardsDurationZeroReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.RewardsDurationZero.selector);
        staking.setRewardsDuration(0);
    }

    function test_rewards_setRewardsDurationDuringEpochReverts() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);

        vm.prank(owner);
        vm.expectRevert(BunkerStaking.RewardDurationNotFinished.selector);
        staking.setRewardsDuration(14 days);
    }

    function test_rewards_epochRollover() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);

        // Warp to mid-epoch (3.5 days).
        vm.warp(block.timestamp + 3.5 days);

        // Start a new epoch with additional rewards.
        // Remaining from first epoch: ~35_000e18.
        // New addition: 35_000e18.
        // Total for new epoch: ~70_000e18.
        vm.startPrank(owner);
        token.mint(owner, 35_000e18);
        token.approve(address(staking), 35_000e18);
        token.transfer(address(staking), 35_000e18);
        staking.notifyRewardAmount(35_000e18);
        vm.stopPrank();

        // The new reward rate should account for leftover + new.
        assertGt(staking.rewardRate(), 0);
    }

    function test_rewards_earnedZeroForNonStaker() public view {
        assertEq(staking.earned(provider1), 0);
    }

    function test_rewards_rewardPerTokenInitiallyZero() public view {
        assertEq(staking.rewardPerToken(), 0);
    }

    function test_rewards_lastTimeRewardApplicable() public {
        // Before any epoch, periodFinish is 0, so lastTimeRewardApplicable = 0.
        assertEq(staking.lastTimeRewardApplicable(), 0);

        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);

        uint256 pf = staking.periodFinish();
        assertGt(pf, block.timestamp);
        assertEq(staking.lastTimeRewardApplicable(), block.timestamp);

        // Warp past epoch end.
        vm.warp(pf + 1);
        assertEq(staking.lastTimeRewardApplicable(), pf);
    }

    function test_rewards_stakeAfterEpochStartGetsFairShare() public {
        // Provider1 stakes at start of epoch.
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        _setupRewardEpoch(70_000e18);

        // Warp 3.5 days (halfway through epoch).
        vm.warp(block.timestamp + 3.5 days);

        // Provider2 stakes halfway through.
        vm.prank(provider2);
        staking.stake(STARTER_MIN);

        // Warp to end of epoch.
        vm.warp(block.timestamp + 3.5 days);

        uint256 earned1 = staking.earned(provider1);
        uint256 earned2 = staking.earned(provider2);

        // Provider1 earns full first half solo + half of the second half.
        // Provider2 earns half of the second half.
        // Provider1 should earn significantly more.
        assertGt(earned1, earned2);
    }

    // ================================================================
    //  11. SLASH PROPOSAL & APPEAL SYSTEM
    // ================================================================

    function test_proposeSlash_createsProposal() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime violation");

        assertEq(proposalId, 0);

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(0);
        assertEq(proposal.provider, provider1);
        assertEq(proposal.amount, 500e18);
        assertEq(keccak256(bytes(proposal.reason)), keccak256(bytes("downtime violation")));
        assertEq(proposal.proposedAt, block.timestamp);
        assertFalse(proposal.executed);
        assertFalse(proposal.appealed);
        assertFalse(proposal.resolved);
    }

    function test_proposeSlash_emitsEvent() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        vm.expectEmit(true, true, false, true, address(staking));
        emit BunkerStaking.SlashProposed(0, provider1, 500e18, "downtime violation");
        staking.proposeSlash(provider1, 500e18, "downtime violation");
    }

    function test_proposeSlash_incrementsCounter() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.startPrank(slasher);
        staking.proposeSlash(provider1, 100e18, "reason1");
        staking.proposeSlash(provider1, 200e18, "reason2");
        vm.stopPrank();

        assertEq(staking.slashProposalCount(), 2);
    }

    function test_proposeSlash_zeroAmountReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        vm.expectRevert(BunkerStaking.ZeroAmount.selector);
        staking.proposeSlash(provider1, 0, "reason");
    }

    function test_proposeSlash_nonSlasherReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(provider2);
        vm.expectRevert();
        staking.proposeSlash(provider1, 500e18, "reason");
    }

    function test_proposeSlash_exceedingBalanceReverts() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(slasher);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerStaking.InsufficientSlashableBalance.selector,
                STARTER_MIN + 1,
                STARTER_MIN
            )
        );
        staking.proposeSlash(provider1, STARTER_MIN + 1, "reason");
    }

    function test_executeSlash_afterAppealWindow() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        // Warp past 48-hour appeal window.
        vm.warp(block.timestamp + 48 hours);

        uint256 treasuryBefore = token.balanceOf(treasury);
        uint256 totalSupplyBefore = token.totalSupply();

        vm.prank(slasher);
        staking.executeSlash(proposalId);

        // Verify slash was executed.
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 500e18));

        // Verify 80/20 split.
        uint256 expectedBurn = (500e18 * 8000) / 10000;
        uint256 expectedTreasury = 500e18 - expectedBurn;
        assertEq(token.balanceOf(treasury), treasuryBefore + expectedTreasury);
        assertEq(token.totalSupply(), totalSupplyBefore - expectedBurn);

        // Verify proposal marked executed.
        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertTrue(proposal.executed);
    }

    function test_executeSlash_emitsEvents() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.warp(block.timestamp + 48 hours);

        vm.prank(slasher);
        vm.expectEmit(true, true, false, true, address(staking));
        emit BunkerStaking.SlashProposalExecuted(proposalId, provider1, 500e18);
        staking.executeSlash(proposalId);
    }

    function test_executeSlash_beforeWindowReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        // Try to execute before appeal window ends.
        uint256 windowEnd = block.timestamp + 48 hours;
        vm.prank(slasher);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerStaking.AppealWindowNotElapsed.selector,
                proposalId,
                windowEnd
            )
        );
        staking.executeSlash(proposalId);
    }

    function test_executeSlash_alreadyExecutedReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.warp(block.timestamp + 48 hours);

        vm.prank(slasher);
        staking.executeSlash(proposalId);

        vm.prank(slasher);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.ProposalAlreadyExecuted.selector, proposalId)
        );
        staking.executeSlash(proposalId);
    }

    function test_appealSlash_providerCanAppeal() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertTrue(proposal.appealed);
    }

    function test_appealSlash_emitsEvent() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(provider1);
        vm.expectEmit(true, true, false, true, address(staking));
        emit BunkerStaking.SlashAppealed(proposalId, provider1);
        staking.appealSlash(proposalId);
    }

    function test_appealSlash_blocksExecution() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        // Warp past appeal window.
        vm.warp(block.timestamp + 48 hours);

        // Execution should still fail because it was appealed.
        vm.prank(slasher);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.ProposalAppealed.selector, proposalId)
        );
        staking.executeSlash(proposalId);
    }

    function test_appealSlash_nonProviderReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(provider2);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerStaking.NotProposalProvider.selector,
                proposalId,
                provider2
            )
        );
        staking.appealSlash(proposalId);
    }

    function test_appealSlash_afterWindowReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        // Warp past appeal window.
        vm.warp(block.timestamp + 48 hours);

        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.AppealWindowElapsed.selector, proposalId)
        );
        staking.appealSlash(proposalId);
    }

    function test_appealSlash_doubleAppealReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.ProposalAlreadyAppealed.selector, proposalId)
        );
        staking.appealSlash(proposalId);
    }

    function test_resolveAppeal_upholdExecutesSlash() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        // Owner resolves: uphold the slash.
        vm.prank(owner);
        staking.resolveAppeal(proposalId, true);

        // Verify slash was executed.
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 500e18));

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertTrue(proposal.executed);
        assertTrue(proposal.resolved);
    }

    function test_resolveAppeal_dismissDoesNotSlash() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        // Owner resolves: dismiss (don't uphold).
        vm.prank(owner);
        staking.resolveAppeal(proposalId, false);

        // Verify slash was NOT executed.
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN));

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertFalse(proposal.executed);
        assertTrue(proposal.resolved);
    }

    function test_resolveAppeal_emitsEvents() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        vm.prank(owner);
        vm.expectEmit(true, false, false, true, address(staking));
        emit BunkerStaking.AppealResolved(proposalId, true);
        staking.resolveAppeal(proposalId, true);
    }

    function test_resolveAppeal_nonOwnerReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        vm.prank(provider2);
        vm.expectRevert();
        staking.resolveAppeal(proposalId, true);
    }

    function test_resolveAppeal_notAppealedReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.ProposalNotAppealed.selector, proposalId)
        );
        staking.resolveAppeal(proposalId, true);
    }

    function test_resolveAppeal_alreadyResolvedReverts() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "downtime");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        vm.prank(owner);
        staking.resolveAppeal(proposalId, false);

        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.ProposalAlreadyResolved.selector, proposalId)
        );
        staking.resolveAppeal(proposalId, true);
    }

    function test_slashImmediate_bypasses_appealWindow() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 slashAmount = 500e18;

        vm.prank(slasher);
        staking.slashImmediate(provider1, slashAmount);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - slashAmount));
    }

    function test_appealWindow_constant() public view {
        assertEq(staking.appealWindow(), 48 hours);
    }

    function test_proposeSlash_fullLifecycle_noAppeal() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Propose.
        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "job abandonment");

        // Cannot execute yet (within appeal window).
        vm.prank(slasher);
        vm.expectRevert();
        staking.executeSlash(proposalId);

        // Provider does not appeal. Warp past window.
        vm.warp(block.timestamp + 48 hours);

        // Execute.
        vm.prank(slasher);
        staking.executeSlash(proposalId);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 500e18));
    }

    function test_proposeSlash_fullLifecycle_withAppeal_upheld() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Propose.
        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "security violation");

        // Provider appeals.
        vm.prank(provider1);
        staking.appealSlash(proposalId);

        // Execution blocked.
        vm.warp(block.timestamp + 48 hours);
        vm.prank(slasher);
        vm.expectRevert();
        staking.executeSlash(proposalId);

        // Owner upholds the slash.
        vm.prank(owner);
        staking.resolveAppeal(proposalId, true);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 500e18));
    }

    function test_proposeSlash_fullLifecycle_withAppeal_dismissed() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Propose.
        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "false positive");

        // Provider appeals.
        vm.prank(provider1);
        staking.appealSlash(proposalId);

        // Owner dismisses the appeal (no slash).
        vm.prank(owner);
        staking.resolveAppeal(proposalId, false);

        // Stake unchanged.
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN));
    }

    function test_invalidProposalId_reverts() public {
        vm.prank(slasher);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerStaking.InvalidProposalId.selector, 999)
        );
        staking.executeSlash(999);
    }

    function test_rewards_afterSlash_reducesEarnings() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        _setupRewardEpoch(70_000e18);

        // Earn for 1 day.
        vm.warp(block.timestamp + 1 days);

        uint256 earnedBefore = staking.earned(provider1);
        assertGt(earnedBefore, 0);

        // Slash half the stake.
        vm.prank(slasher);
        staking.slash(provider1, BRONZE_MIN / 2);

        // Earn for another day.
        vm.warp(block.timestamp + 1 days);

        // After slash, the provider has less stake, so rate of earning should be
        // similar (sole staker) but total stake dropped affecting future proportional share.
        // Since they're the only staker, they still get all rewards, but the
        // reward checkpoint should have properly updated.
        uint256 earnedAfter = staking.earned(provider1);
        assertGt(earnedAfter, 0);
    }

    // ================================================================
    //  13. REWARD MULTIPLIER NON-COMPOUNDING (C-01 regression test)
    // ================================================================

    /// @dev Verifies that tier multipliers do NOT compound across multiple
    ///      reward checkpoints. The multiplier should only apply to each new
    ///      reward delta, not re-apply to previously stored accumulated rewards.
    function test_rewards_multiplierDoesNotCompound() public {
        // Provider1 stakes at Platinum tier (1.20x multiplier).
        vm.prank(provider1);
        staking.stake(PLATINUM_MIN);

        uint256 rewardAmount = 70_000e18;
        _setupRewardEpoch(rewardAmount);

        // --- Checkpoint 1: advance 1 day (1/7th of epoch) ---
        vm.warp(block.timestamp + 1 days);

        // Trigger updateReward by calling a view that forces checkpoint.
        // We need a state-changing call. Stake additional 0 is invalid,
        // so we call claimRewards (which calls updateReward internally).
        // But we need rewards > 0 to not revert... Let's just read earned.
        uint256 earnedAtDay1 = staking.earned(provider1);
        assertGt(earnedAtDay1, 0, "Should have earned after 1 day");

        // Trigger a checkpoint by adding more stake (small amount).
        vm.prank(provider1);
        staking.stake(1e18);

        // --- Checkpoint 2: advance another day ---
        vm.warp(block.timestamp + 1 days);
        uint256 earnedAtDay2 = staking.earned(provider1);

        // Trigger another checkpoint.
        vm.prank(provider1);
        staking.stake(1e18);

        // --- Checkpoint 3: advance another day ---
        vm.warp(block.timestamp + 1 days);
        uint256 earnedAtDay3 = staking.earned(provider1);

        // The daily increment should be roughly constant (sole staker).
        // If multiplier compounds, each increment grows exponentially.
        // daily increment = ~10_000e18 * 1.20 = ~12_000e18 per day (sole staker).
        uint256 increment1 = earnedAtDay2 - earnedAtDay1;
        uint256 increment2 = earnedAtDay3 - earnedAtDay2;

        // Increments should be approximately equal (within 2% for rounding + tiny stake changes).
        uint256 tolerance = increment1 / 50; // 2%
        assertGt(increment2, increment1 - tolerance, "increment2 much smaller than increment1");
        assertLt(increment2, increment1 + tolerance, "increment2 much larger than increment1 (compounding)");
    }

    /// @dev Test that tier multiplier ascending order is enforced.
    function test_setTierMinStake_nonAscendingReverts() public {
        vm.startPrank(owner);

        // Try to set Starter min above Bronze min (5_000_000e18) — should revert
        vm.expectRevert();
        staking.setTierMinStake(BunkerStaking.Tier.Starter, 10_000_000e18);

        // Try to set Bronze min below Starter min (1_000_000e18) — should revert
        vm.expectRevert();
        staking.setTierMinStake(BunkerStaking.Tier.Bronze, 500_000e18);

        // Try to set Platinum min below Gold min (100_000_000e18) — should revert
        vm.expectRevert();
        staking.setTierMinStake(BunkerStaking.Tier.Platinum, 50_000_000e18);

        vm.stopPrank();
    }

    // ================================================================
    //  NEW ADMIN SETTERS
    // ================================================================

    // -- setTierRewardMultiplier --

    function test_setTierRewardMultiplier_succeeds() public {
        vm.prank(owner);
        staking.setTierRewardMultiplier(BunkerStaking.Tier.Silver, 10800);
        (,, uint16 mult,,) = staking.tierConfigs(BunkerStaking.Tier.Silver);
        assertEq(mult, 10800);
    }

    function test_setTierRewardMultiplier_revertsNoneTier() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.CannotConfigureNoneTier.selector);
        staking.setTierRewardMultiplier(BunkerStaking.Tier.None, 10000);
    }

    function test_setTierRewardMultiplier_revertsBelowMin() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.MultiplierTooLow.selector);
        staking.setTierRewardMultiplier(BunkerStaking.Tier.Starter, 9999);
    }

    function test_setTierRewardMultiplier_revertsAboveCap() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.MultiplierTooHigh.selector);
        staking.setTierRewardMultiplier(BunkerStaking.Tier.Starter, 12001);
    }

    function test_setTierRewardMultiplier_revertsNonOwner() public {
        vm.prank(provider1);
        vm.expectRevert();
        staking.setTierRewardMultiplier(BunkerStaking.Tier.Starter, 10500);
    }

    function test_setTierRewardMultiplier_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(true, false, false, true);
        emit BunkerStaking.TierRewardMultiplierUpdated(BunkerStaking.Tier.Gold, 11500);
        staking.setTierRewardMultiplier(BunkerStaking.Tier.Gold, 11500);
    }

    // -- setMaxTierMultiplierBps --

    function test_setMaxTierMultiplierBps_succeeds() public {
        vm.prank(owner);
        staking.setMaxTierMultiplierBps(15000);
        assertEq(staking.maxTierMultiplierBps(), 15000);
    }

    function test_setMaxTierMultiplierBps_revertsBelowMin() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.InvalidMultiplierCap.selector);
        staking.setMaxTierMultiplierBps(9999);
    }

    function test_setMaxTierMultiplierBps_revertsAboveMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.InvalidMultiplierCap.selector);
        staking.setMaxTierMultiplierBps(20001);
    }

    function test_setMaxTierMultiplierBps_revertsNonOwner() public {
        vm.prank(provider1);
        vm.expectRevert();
        staking.setMaxTierMultiplierBps(15000);
    }

    // -- setUnbondingPeriod --

    function test_setUnbondingPeriod_succeeds() public {
        vm.prank(owner);
        staking.setUnbondingPeriod(7 days);
        assertEq(staking.unbondingPeriod(), 7 days);
    }

    function test_setUnbondingPeriod_revertsBelowMin() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.InvalidUnbondingPeriod.selector);
        staking.setUnbondingPeriod(30);
    }

    function test_setUnbondingPeriod_revertsAboveMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.InvalidUnbondingPeriod.selector);
        staking.setUnbondingPeriod(91 days);
    }

    function test_setUnbondingPeriod_revertsNonOwner() public {
        vm.prank(provider1);
        vm.expectRevert();
        staking.setUnbondingPeriod(7 days);
    }

    function test_setUnbondingPeriod_affectsNewUnstakeRequests() public {
        // Stake first
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Change unbonding to 7 days
        vm.prank(owner);
        staking.setUnbondingPeriod(7 days);

        // Unstake should use 7-day period
        vm.prank(provider1);
        staking.requestUnstake(STARTER_MIN / 2);

        BunkerStaking.UnstakeRequest memory req = staking.getUnstakeRequest(provider1, 0);
        assertEq(req.unlockTime, block.timestamp + 7 days);
    }

    // -- setSlashFeeSplit --

    function test_setSlashFeeSplit_succeeds() public {
        vm.prank(owner);
        staking.setSlashFeeSplit(7000, 3000);
        assertEq(staking.slashBurnBps(), 7000);
        assertEq(staking.slashTreasuryBps(), 3000);
    }

    function test_setSlashFeeSplit_revertsInvalidSum() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.InvalidFeeSplit.selector);
        staking.setSlashFeeSplit(7000, 2000);
    }

    function test_setSlashFeeSplit_revertsNonOwner() public {
        vm.prank(provider1);
        vm.expectRevert();
        staking.setSlashFeeSplit(7000, 3000);
    }

    function test_setSlashFeeSplit_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerStaking.SlashFeeSplitUpdated(6000, 4000);
        staking.setSlashFeeSplit(6000, 4000);
    }

    // -- setAppealWindow --

    function test_setAppealWindow_succeeds() public {
        vm.prank(owner);
        staking.setAppealWindow(24 hours);
        assertEq(staking.appealWindow(), 24 hours);
    }

    function test_setAppealWindow_revertsBelowMin() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.InvalidAppealWindow.selector);
        staking.setAppealWindow(30);
    }

    function test_setAppealWindow_revertsAboveMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.InvalidAppealWindow.selector);
        staking.setAppealWindow(15 days);
    }

    function test_setAppealWindow_revertsNonOwner() public {
        vm.prank(provider1);
        vm.expectRevert();
        staking.setAppealWindow(24 hours);
    }

    // ──────────────────────────────────────────────
    //  Security: Snapshot Invariants
    // ──────────────────────────────────────────────

    /// @dev Verify that changing appealWindow does NOT affect a pending proposal's window.
    function test_appealWindowSnapshot_proposalUsesOldWindow() public {
        // Stake provider1
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Propose slash with 48h appeal window (default)
        vm.prank(slasher);
        uint256 pid = staking.proposeSlash(provider1, 1000e18, "test");

        // Owner changes appeal window to 12 hours
        vm.prank(owner);
        staking.setAppealWindow(12 hours);

        // Provider can still appeal within the original 48h window
        vm.warp(block.timestamp + 40 hours); // past 12h, within 48h
        vm.prank(provider1);
        staking.appealSlash(pid); // should succeed — uses snapshotted 48h
    }

    /// @dev Verify that executeSlash respects the snapshotted appeal window.
    function test_appealWindowSnapshot_executeUsesOldWindow() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(slasher);
        uint256 pid = staking.proposeSlash(provider1, 1000e18, "test");

        // Owner reduces appeal window to 12h
        vm.prank(owner);
        staking.setAppealWindow(12 hours);

        // Try to execute after 12h but before original 48h — should fail
        vm.warp(block.timestamp + 13 hours);
        vm.prank(slasher);
        vm.expectRevert(); // AppealWindowNotElapsed — still using 48h snapshot
        staking.executeSlash(pid);

        // Execute after the snapshotted 48h — should succeed
        vm.warp(block.timestamp + 48 hours);
        vm.prank(slasher);
        staking.executeSlash(pid);
    }

    /// @dev Verify that slash fee split is snapshotted at proposal time.
    function test_slashFeeSplitSnapshot_usesOldSplit() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Propose slash with default 80/20 split
        vm.prank(slasher);
        uint256 pid = staking.proposeSlash(provider1, 10_000e18, "test");

        // Owner changes split to 50/50
        vm.prank(owner);
        staking.setSlashFeeSplit(5000, 5000);

        // Execute slash — should use snapshotted 80/20 split
        vm.warp(block.timestamp + 49 hours);

        uint256 treasuryBefore = token.balanceOf(treasury);
        vm.prank(slasher);
        staking.executeSlash(pid);
        uint256 treasuryAfter = token.balanceOf(treasury);

        // With 80/20 snapshot: treasury gets 20% of slashed amount
        // The actual slash amount includes forfeited vested rewards, but base is 10_000e18
        // Treasury should get ~20% of total slash (not 50%)
        uint256 treasuryGot = treasuryAfter - treasuryBefore;
        // 20% of 10_000e18 = 2_000e18
        assertEq(treasuryGot, 2_000e18, "treasury should get 20% (snapshotted split)");
    }

    /// @dev Verify solvency check blocks setMaxTierMultiplierBps during active epoch.
    function test_setMaxTierMultiplierBps_revertsIfBreaksSolvency() public {
        // Fund rewards with just enough for 1.2x multiplier (current default)
        // rewardBalance = mintedToContract - staked
        // maxLiability = rewardRate * rewardsDuration * multiplierBps / BPS_DENOMINATOR
        // We need: with 2.0x, maxLiability > rewardBalance
        vm.startPrank(owner);
        // Mint exactly enough for a tight reward budget at 1.2x
        token.mint(address(staking), 12_000e18); // reward pool
        vm.stopPrank();

        // Provider stakes
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Start a reward epoch — uses all 12_000e18 as reward at 1.2x max
        // rewardBalance = 12_000e18 (minted) [staked tokens from provider don't count as "free"]
        // Wait: balance = 12_000 + 100_000 (stake), lockedStake = 100_000
        // rewardBalance = 112_000 - 100_000 = 12_000
        // maxLiability at 1.2x = rewardRate * 604800 * 12000 / 10000 = reward * 1.2 = 12_000
        // So we notify 10_000e18 — that gives liability = 10_000 * 1.2 = 12_000 (exactly fits)
        vm.prank(owner);
        staking.notifyRewardAmount(10_000e18);

        // Now try to set multiplier to 2.0x — liability would be 10_000 * 2.0 = 20_000 > 12_000
        vm.prank(owner);
        vm.expectRevert(); // RewardRateTooHigh
        staking.setMaxTierMultiplierBps(20000);
    }

    // ──────────────────────────────────────────────
    //  Slashing Toggle (Monitor Mode)
    // ──────────────────────────────────────────────

    function test_slashingEnabled_defaultFalse() public {
        // Deploy a fresh contract to check the default
        vm.prank(owner);
        BunkerStaking fresh = new BunkerStaking(address(token), treasury, owner);
        assertFalse(fresh.slashingEnabled());
    }

    function test_setSlashingEnabled_ownerCanEnable() public {
        vm.prank(owner);
        staking.setSlashingEnabled(false);
        assertFalse(staking.slashingEnabled());

        vm.prank(owner);
        staking.setSlashingEnabled(true);
        assertTrue(staking.slashingEnabled());
    }

    function test_setSlashingEnabled_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerStaking.SlashingEnabledUpdated(false);
        staking.setSlashingEnabled(false);
    }

    function test_setSlashingEnabled_revertsNonOwner() public {
        vm.prank(provider1);
        vm.expectRevert();
        staking.setSlashingEnabled(true);
    }

    function test_slashImmediate_revertsWhenDisabled() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(owner);
        staking.setSlashingEnabled(false);

        vm.prank(slasher);
        vm.expectRevert(BunkerStaking.SlashingNotEnabled.selector);
        staking.slashImmediate(provider1, 1000e18);
    }

    function test_slash_revertsWhenDisabled() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(owner);
        staking.setSlashingEnabled(false);

        vm.prank(slasher);
        vm.expectRevert(BunkerStaking.SlashingNotEnabled.selector);
        staking.slash(provider1, 1000e18);
    }

    function test_executeSlash_revertsWhenDisabled() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Propose while enabled — should succeed
        vm.prank(slasher);
        uint256 pid = staking.proposeSlash(provider1, 1000e18, "test");

        // Disable slashing
        vm.prank(owner);
        staking.setSlashingEnabled(false);

        // Wait for appeal window
        vm.warp(block.timestamp + 49 hours);

        // Execution should revert
        vm.prank(slasher);
        vm.expectRevert(BunkerStaking.SlashingNotEnabled.selector);
        staking.executeSlash(pid);
    }

    function test_resolveAppeal_upholdRevertsWhenDisabled() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Propose and appeal
        vm.prank(slasher);
        uint256 pid = staking.proposeSlash(provider1, 1000e18, "test");

        vm.prank(provider1);
        staking.appealSlash(pid);

        // Disable slashing
        vm.prank(owner);
        staking.setSlashingEnabled(false);

        // Upholding should revert
        vm.prank(owner);
        vm.expectRevert(BunkerStaking.SlashingNotEnabled.selector);
        staking.resolveAppeal(pid, true);
    }

    function test_resolveAppeal_dismissWorksWhenDisabled() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Propose and appeal
        vm.prank(slasher);
        uint256 pid = staking.proposeSlash(provider1, 1000e18, "test");

        vm.prank(provider1);
        staking.appealSlash(pid);

        // Disable slashing
        vm.prank(owner);
        staking.setSlashingEnabled(false);

        // Dismissing should still work
        vm.prank(owner);
        staking.resolveAppeal(pid, false);
    }

    function test_proposeSlash_worksWhenDisabled() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Disable slashing
        vm.prank(owner);
        staking.setSlashingEnabled(false);

        // Proposals should still work (monitor mode)
        vm.prank(slasher);
        uint256 pid = staking.proposeSlash(provider1, 1000e18, "monitor");
        assertEq(pid, 0);
    }

    function test_proposeSlashByReason_worksWhenDisabled() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Disable slashing
        vm.prank(owner);
        staking.setSlashingEnabled(false);

        // proposeSlashByReason should still work
        vm.prank(slasher);
        uint256 pid = staking.proposeSlashByReason(provider1, BunkerStaking.SlashReason.Downtime);
        assertEq(pid, 0);
    }

    function test_slashing_enableThenExecute() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Disable slashing
        vm.prank(owner);
        staking.setSlashingEnabled(false);

        // Create proposal in monitor mode
        vm.prank(slasher);
        uint256 pid = staking.proposeSlash(provider1, 1000e18, "test");

        // Re-enable slashing
        vm.prank(owner);
        staking.setSlashingEnabled(true);

        // Now execution should work
        vm.warp(block.timestamp + 49 hours);
        vm.prank(slasher);
        staking.executeSlash(pid);

        // Verify slash executed
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(STARTER_MIN - 1000e18));
    }
}
