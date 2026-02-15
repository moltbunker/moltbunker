// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEscrow.sol";

/// @title BunkerEscrowV2Test
/// @notice Tests for BunkerEscrow V2 features: auto top-up (increaseDeposit),
///         provider claims, cross-contract staking verification, and low-balance events.
contract BunkerEscrowV2Test is Test {
    BunkerToken public token;
    BunkerStaking public staking;
    BunkerEscrow public escrow;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public operator = makeAddr("operator");
    address public requester = makeAddr("requester");
    address public provider0 = makeAddr("provider0");
    address public provider1 = makeAddr("provider1");
    address public provider2 = makeAddr("provider2");
    address public slasher = makeAddr("slasher");

    uint256 public constant RESERVATION_AMOUNT = 1000 ether;
    uint256 public constant RESERVATION_DURATION = 3600; // 1 hour

    // Staking tier threshold (Starter minimum).
    uint256 constant STARTER_MIN = 1_000_000e18;

    function setUp() public {
        // Deploy token.
        vm.startPrank(owner);
        token = new BunkerToken(owner);

        // Deploy staking.
        staking = new BunkerStaking(address(token), treasury, owner);
        staking.grantRole(staking.SLASHER_ROLE(), slasher);

        // Deploy escrow with staking contract reference.
        escrow = new BunkerEscrow(address(token), treasury, owner);
        escrow.grantRole(escrow.OPERATOR_ROLE(), operator);
        escrow.setStakingContract(address(staking));

        // Mint tokens to requester and providers.
        token.mint(requester, 100_000 ether);
        token.mint(provider0, 1_000_000e18);
        token.mint(provider1, 1_000_000e18);
        token.mint(provider2, 1_000_000e18);
        vm.stopPrank();

        // Approve escrow from requester.
        vm.prank(requester);
        token.approve(address(escrow), type(uint256).max);

        // Stake providers so they pass verification.
        vm.prank(provider0);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider0);
        staking.stake(STARTER_MIN);

        vm.prank(provider1);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider2);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider2);
        staking.stake(STARTER_MIN);
    }

    // ──────────────────────────────────────────────
    //  Helpers
    // ──────────────────────────────────────────────

    function _createDefaultReservation() internal returns (uint256) {
        vm.prank(requester);
        return escrow.createReservation(RESERVATION_AMOUNT, RESERVATION_DURATION);
    }

    function _createAndActivateReservation() internal returns (uint256) {
        uint256 id = _createDefaultReservation();
        vm.prank(operator);
        escrow.selectProviders(id, [provider0, provider1, provider2]);
        return id;
    }

    // ================================================================
    //  1. INCREASE DEPOSIT (AUTO TOP-UP)
    // ================================================================

    function test_increaseDeposit_requesterCanAdd() public {
        uint256 id = _createDefaultReservation();

        uint256 additionalAmount = 500 ether;
        uint256 balanceBefore = token.balanceOf(requester);

        vm.prank(requester);
        escrow.increaseDeposit(id, additionalAmount);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.totalAmount, uint128(RESERVATION_AMOUNT + additionalAmount));
        assertEq(token.balanceOf(requester), balanceBefore - additionalAmount);
    }

    function test_increaseDeposit_emitsDepositIncreasedEvent() public {
        uint256 id = _createDefaultReservation();

        vm.prank(requester);
        vm.expectEmit(true, true, false, true, address(escrow));
        emit BunkerEscrow.DepositIncreased(
            id,
            requester,
            500 ether,
            RESERVATION_AMOUNT + 500 ether
        );
        escrow.increaseDeposit(id, 500 ether);
    }

    function test_increaseDeposit_worksOnActiveReservation() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(requester);
        escrow.increaseDeposit(id, 200 ether);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.totalAmount, uint128(RESERVATION_AMOUNT + 200 ether));
    }

    function test_increaseDeposit_notRequesterReverts() public {
        uint256 id = _createDefaultReservation();

        vm.prank(provider0);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.NotRequester.selector, provider0, id)
        );
        escrow.increaseDeposit(id, 100 ether);
    }

    function test_increaseDeposit_wrongStatusReverts() public {
        uint256 id = _createAndActivateReservation();

        // Finalize reservation to Completed status.
        vm.prank(operator);
        escrow.finalizeReservation(id);

        vm.prank(requester);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidStatus.selector,
                id,
                BunkerEscrow.Status.Created,
                BunkerEscrow.Status.Completed
            )
        );
        escrow.increaseDeposit(id, 100 ether);
    }

    function test_increaseDeposit_refundedStatusReverts() public {
        uint256 id = _createDefaultReservation();

        // Refund.
        vm.prank(requester);
        escrow.refund(id);

        vm.prank(requester);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidStatus.selector,
                id,
                BunkerEscrow.Status.Created,
                BunkerEscrow.Status.Refunded
            )
        );
        escrow.increaseDeposit(id, 100 ether);
    }

    function test_increaseDeposit_zeroAmountReverts() public {
        uint256 id = _createDefaultReservation();

        vm.prank(requester);
        vm.expectRevert(BunkerEscrow.ZeroAmount.selector);
        escrow.increaseDeposit(id, 0);
    }

    function test_increaseDeposit_invalidReservationReverts() public {
        vm.prank(requester);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.InvalidReservation.selector, 999)
        );
        escrow.increaseDeposit(999, 100 ether);
    }

    function test_increaseDeposit_multipleIncreasesAccumulate() public {
        uint256 id = _createDefaultReservation();

        vm.startPrank(requester);
        escrow.increaseDeposit(id, 100 ether);
        escrow.increaseDeposit(id, 200 ether);
        escrow.increaseDeposit(id, 300 ether);
        vm.stopPrank();

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.totalAmount, uint128(RESERVATION_AMOUNT + 600 ether));
    }

    function test_increaseDeposit_transfersTokens() public {
        uint256 id = _createDefaultReservation();

        uint256 escrowBalanceBefore = token.balanceOf(address(escrow));

        vm.prank(requester);
        escrow.increaseDeposit(id, 500 ether);

        assertEq(token.balanceOf(address(escrow)), escrowBalanceBefore + 500 ether);
    }

    // ================================================================
    //  2. LOW BALANCE EVENT
    // ================================================================

    function test_lowBalanceEvent_emittedWhenBelowThreshold() public {
        uint256 id = _createAndActivateReservation();

        // Default threshold is 2000 bps (20%).
        // Release 90% of duration -> remaining is 10% < 20% threshold.
        uint256 settledDuration = (RESERVATION_DURATION * 9) / 10;
        uint256 proportionalTotal = (RESERVATION_AMOUNT * settledDuration) / RESERVATION_DURATION;
        uint256 remaining = RESERVATION_AMOUNT - proportionalTotal;
        uint256 threshold = (RESERVATION_AMOUNT * 2000) / 10000;

        // remaining = 100 ether, threshold = 200 ether -> triggers LowBalance.
        assertTrue(remaining < threshold, "remaining should be below threshold");

        vm.prank(operator);
        vm.expectEmit(true, false, false, true, address(escrow));
        emit BunkerEscrow.LowBalance(id, remaining, threshold);
        escrow.releasePayment(id, settledDuration);
    }

    function test_lowBalanceEvent_notEmittedWhenAboveThreshold() public {
        uint256 id = _createAndActivateReservation();

        // Release only 10% of duration -> remaining is 90% > 20% threshold.
        uint256 settledDuration = RESERVATION_DURATION / 10;

        // Record logs to verify LowBalance is NOT emitted.
        vm.recordLogs();

        vm.prank(operator);
        escrow.releasePayment(id, settledDuration);

        Vm.Log[] memory entries = vm.getRecordedLogs();
        bytes32 lowBalanceSig = keccak256("LowBalance(uint256,uint256,uint256)");
        for (uint256 i = 0; i < entries.length; i++) {
            assertTrue(
                entries[i].topics[0] != lowBalanceSig,
                "LowBalance event should not be emitted when above threshold"
            );
        }
    }

    function test_lowBalanceEvent_emittedOnFinalization() public {
        uint256 id = _createAndActivateReservation();

        // Finalization releases everything -> remaining = 0 < threshold.
        uint256 threshold = (RESERVATION_AMOUNT * 2000) / 10000;

        vm.prank(operator);
        vm.expectEmit(true, false, false, true, address(escrow));
        emit BunkerEscrow.LowBalance(id, 0, threshold);
        escrow.finalizeReservation(id);
    }

    function test_lowBalanceEvent_emittedOnProviderClaim() public {
        uint256 id = _createAndActivateReservation();

        // Warp to 95% through the reservation.
        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        vm.warp(res.startTime + (RESERVATION_DURATION * 95) / 100);

        uint256 elapsed = block.timestamp - res.startTime;
        uint256 proportionalTotal = (RESERVATION_AMOUNT * elapsed) / RESERVATION_DURATION;
        uint256 remaining = RESERVATION_AMOUNT - proportionalTotal;
        uint256 threshold = (RESERVATION_AMOUNT * 2000) / 10000;

        assertTrue(remaining < threshold, "remaining should trigger low balance");

        vm.prank(provider0);
        vm.expectEmit(true, false, false, true, address(escrow));
        emit BunkerEscrow.LowBalance(id, remaining, threshold);
        escrow.claim(id);
    }

    // ================================================================
    //  3. PROVIDER CLAIM
    // ================================================================

    function test_providerClaim_basedOnElapsedTime() public {
        uint256 id = _createAndActivateReservation();

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        // Warp to 50% through the reservation.
        vm.warp(res.startTime + RESERVATION_DURATION / 2);

        uint256 p0Before = token.balanceOf(provider0);

        vm.prank(provider0);
        escrow.claim(id);

        // Provider should have received tokens.
        assertGt(token.balanceOf(provider0), p0Before);

        // Check released amount.
        BunkerEscrow.Reservation memory resAfter = escrow.getReservation(id);
        uint256 expectedRelease = (RESERVATION_AMOUNT * (RESERVATION_DURATION / 2)) / RESERVATION_DURATION;
        assertEq(resAfter.releasedAmount, uint128(expectedRelease));
    }

    function test_providerClaim_notProviderReverts() public {
        uint256 id = _createAndActivateReservation();

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        vm.warp(res.startTime + RESERVATION_DURATION / 2);

        address stranger = makeAddr("stranger");
        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.NotProvider.selector, stranger, id)
        );
        escrow.claim(id);
    }

    function test_providerClaim_noFundsReverts() public {
        uint256 id = _createAndActivateReservation();

        // Claim immediately (no time elapsed) -> nothing to release.
        vm.prank(provider0);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.NothingToRelease.selector, id)
        );
        escrow.claim(id);
    }

    function test_providerClaim_nonActiveReverts() public {
        uint256 id = _createDefaultReservation();

        vm.prank(provider0);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidStatus.selector,
                id,
                BunkerEscrow.Status.Active,
                BunkerEscrow.Status.Created
            )
        );
        escrow.claim(id);
    }

    function test_providerClaim_multipleClaimsByDifferentProviders() public {
        uint256 id = _createAndActivateReservation();

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        // Warp to 50% through.
        vm.warp(res.startTime + RESERVATION_DURATION / 2);

        // Provider0 claims.
        vm.prank(provider0);
        escrow.claim(id);

        // Warp to 75% through.
        vm.warp(res.startTime + (RESERVATION_DURATION * 3) / 4);

        // Provider1 claims (should release the incremental amount).
        uint256 p1Before = token.balanceOf(provider1);
        vm.prank(provider1);
        escrow.claim(id);

        assertGt(token.balanceOf(provider1), p1Before);
    }

    function test_providerClaim_capsAtDuration() public {
        uint256 id = _createAndActivateReservation();

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        // Warp well past the reservation duration.
        vm.warp(res.startTime + RESERVATION_DURATION * 2);

        vm.prank(provider0);
        escrow.claim(id);

        // Should release exactly the total amount.
        BunkerEscrow.Reservation memory resAfter = escrow.getReservation(id);
        assertEq(resAfter.releasedAmount, resAfter.totalAmount);
    }

    function test_providerClaim_secondClaimAfterFullReleaseReverts() public {
        uint256 id = _createAndActivateReservation();

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        vm.warp(res.startTime + RESERVATION_DURATION);

        // First claim releases everything.
        vm.prank(provider0);
        escrow.claim(id);

        // Second claim reverts.
        vm.prank(provider1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.NothingToRelease.selector, id)
        );
        escrow.claim(id);
    }

    function test_providerClaim_emitsPaymentReleasedEvent() public {
        uint256 id = _createAndActivateReservation();

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        vm.warp(res.startTime + RESERVATION_DURATION / 2);

        uint256 grossRelease = (RESERVATION_AMOUNT * (RESERVATION_DURATION / 2)) / RESERVATION_DURATION;
        uint256 fee = (grossRelease * 500) / 10000;
        uint256 burnAmount = (fee * 8000) / 10000;

        vm.prank(provider0);
        vm.expectEmit(true, false, false, true, address(escrow));
        emit BunkerEscrow.PaymentReleased(
            id,
            grossRelease,
            grossRelease - fee,
            fee,
            burnAmount,
            fee - burnAmount
        );
        escrow.claim(id);
    }

    // ================================================================
    //  4. STAKING VERIFICATION
    // ================================================================

    function test_stakingVerification_activeProviderAccepted() public {
        uint256 id = _createDefaultReservation();

        // Providers are already staked in setUp().
        vm.prank(operator);
        escrow.selectProviders(id, [provider0, provider1, provider2]);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Active));
    }

    function test_stakingVerification_unstakedProviderReverts() public {
        uint256 id = _createDefaultReservation();

        address unstakedProvider = makeAddr("unstaked");

        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.ProviderNotStaked.selector, unstakedProvider)
        );
        escrow.selectProviders(id, [unstakedProvider, provider1, provider2]);
    }

    function test_stakingVerification_secondProviderUnstakedReverts() public {
        uint256 id = _createDefaultReservation();

        address unstakedProvider = makeAddr("unstaked2");

        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.ProviderNotStaked.selector, unstakedProvider)
        );
        escrow.selectProviders(id, [provider0, unstakedProvider, provider2]);
    }

    function test_stakingVerification_thirdProviderUnstakedReverts() public {
        uint256 id = _createDefaultReservation();

        address unstakedProvider = makeAddr("unstaked3");

        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.ProviderNotStaked.selector, unstakedProvider)
        );
        escrow.selectProviders(id, [provider0, provider1, unstakedProvider]);
    }

    function test_stakingVerification_skippedWhenNoStakingContract() public {
        // Deploy escrow without staking contract reference.
        vm.startPrank(owner);
        BunkerEscrow escrowNoStaking = new BunkerEscrow(address(token), treasury, owner);
        escrowNoStaking.grantRole(escrowNoStaking.OPERATOR_ROLE(), operator);
        vm.stopPrank();

        // Approve from requester.
        vm.prank(requester);
        token.approve(address(escrowNoStaking), type(uint256).max);

        vm.prank(requester);
        uint256 id = escrowNoStaking.createReservation(RESERVATION_AMOUNT, RESERVATION_DURATION);

        // Should succeed even with unstaked providers (no staking contract set).
        address unstakedProvider = makeAddr("unstaked4");
        vm.prank(operator);
        escrowNoStaking.selectProviders(id, [unstakedProvider, provider1, provider2]);

        BunkerEscrow.Reservation memory res = escrowNoStaking.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Active));
    }

    function test_stakingVerification_deactivatedProviderReverts() public {
        uint256 id = _createDefaultReservation();

        // Deactivate provider0 by unstaking everything.
        vm.prank(provider0);
        staking.requestUnstake(STARTER_MIN);

        // provider0 is now inactive.
        assertFalse(staking.isActiveProvider(provider0));

        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.ProviderNotStaked.selector, provider0)
        );
        escrow.selectProviders(id, [provider0, provider1, provider2]);
    }

    // ================================================================
    //  5. ADMIN: LOW BALANCE THRESHOLD
    // ================================================================

    function test_setLowBalanceThreshold_ownerCanUpdate() public {
        vm.prank(owner);
        escrow.setLowBalanceThreshold(3000);

        assertEq(escrow.lowBalanceThresholdBps(), 3000);
    }

    function test_setLowBalanceThreshold_maxIs5000() public {
        vm.prank(owner);
        escrow.setLowBalanceThreshold(5000);

        assertEq(escrow.lowBalanceThresholdBps(), 5000);
    }

    function test_setLowBalanceThreshold_aboveMaxReverts() public {
        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.ThresholdTooHigh.selector, 5001, 5000)
        );
        escrow.setLowBalanceThreshold(5001);
    }

    function test_setLowBalanceThreshold_nonOwnerReverts() public {
        vm.prank(operator);
        vm.expectRevert();
        escrow.setLowBalanceThreshold(3000);
    }

    function test_setLowBalanceThreshold_zeroAllowed() public {
        vm.prank(owner);
        escrow.setLowBalanceThreshold(0);

        assertEq(escrow.lowBalanceThresholdBps(), 0);
    }

    function test_setLowBalanceThreshold_zeroSuppressesEvent() public {
        vm.prank(owner);
        escrow.setLowBalanceThreshold(0);

        uint256 id = _createAndActivateReservation();

        // Release 90% -> remaining = 10%. With threshold = 0, no LowBalance event.
        vm.recordLogs();

        vm.prank(operator);
        escrow.releasePayment(id, (RESERVATION_DURATION * 9) / 10);

        Vm.Log[] memory entries = vm.getRecordedLogs();
        bytes32 lowBalanceSig = keccak256("LowBalance(uint256,uint256,uint256)");
        for (uint256 i = 0; i < entries.length; i++) {
            assertTrue(
                entries[i].topics[0] != lowBalanceSig,
                "LowBalance should not be emitted with zero threshold"
            );
        }
    }

    // ================================================================
    //  6. ADMIN: SET STAKING CONTRACT
    // ================================================================

    function test_setStakingContract_ownerCanSet() public {
        address newStaking = makeAddr("newStaking");

        vm.prank(owner);
        vm.expectEmit(true, true, false, true, address(escrow));
        emit BunkerEscrow.StakingContractUpdated(address(staking), newStaking);
        escrow.setStakingContract(newStaking);

        assertEq(escrow.stakingContract(), newStaking);
    }

    function test_setStakingContract_nonOwnerReverts() public {
        vm.prank(operator);
        vm.expectRevert();
        escrow.setStakingContract(makeAddr("newStaking"));
    }

    function test_setStakingContract_zeroAddressDisablesVerification() public {
        // After the security audit fix, setting staking contract to zero address
        // is no longer allowed once it has been set to a non-zero address.
        vm.prank(owner);
        vm.expectRevert(BunkerEscrow.CannotDisableStakingVerification.selector);
        escrow.setStakingContract(address(0));
    }

    // ================================================================
    //  7. INCREASE DEPOSIT AFTER PARTIAL RELEASE
    // ================================================================

    function test_increaseDeposit_afterPartialRelease() public {
        uint256 id = _createAndActivateReservation();

        // Release 50%.
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION / 2);

        // Top up with additional tokens.
        vm.prank(requester);
        escrow.increaseDeposit(id, 500 ether);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.totalAmount, uint128(RESERVATION_AMOUNT + 500 ether));
        assertEq(res.releasedAmount, uint128(RESERVATION_AMOUNT / 2));
    }

    // ================================================================
    //  8. FUZZ TESTS
    // ================================================================

    function testFuzz_increaseDeposit(uint128 additional) public {
        vm.assume(additional > 0 && uint256(additional) <= 50_000 ether);

        uint256 id = _createDefaultReservation();

        vm.prank(requester);
        escrow.increaseDeposit(id, additional);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.totalAmount, uint128(RESERVATION_AMOUNT + uint256(additional)));
    }

    function testFuzz_providerClaim_atVariousTimestamps(uint48 elapsedFraction) public {
        // elapsedFraction represents a fraction of RESERVATION_DURATION in basis points.
        // We need elapsed >= 1 second. With RESERVATION_DURATION=3600, elapsedFraction
        // must be >= ceil(10000/3600) = 3 to guarantee elapsed >= 1 second.
        vm.assume(elapsedFraction >= 3 && elapsedFraction <= 10000);

        uint256 id = _createAndActivateReservation();

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        uint256 elapsed = (RESERVATION_DURATION * uint256(elapsedFraction)) / 10000;
        require(elapsed >= 1, "elapsed must be at least 1 second");
        vm.warp(res.startTime + elapsed);

        vm.prank(provider0);
        escrow.claim(id);

        BunkerEscrow.Reservation memory resAfter = escrow.getReservation(id);
        assertGt(resAfter.releasedAmount, 0);
    }
}
