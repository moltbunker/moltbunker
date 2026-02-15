// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/access/IAccessControl.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerEscrow.sol";

contract BunkerEscrowTest is Test {
    BunkerToken public token;
    BunkerEscrow public escrow;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public operator = makeAddr("operator");
    address public requester = makeAddr("requester");
    address public provider0 = makeAddr("provider0");
    address public provider1 = makeAddr("provider1");
    address public provider2 = makeAddr("provider2");

    uint256 public constant RESERVATION_AMOUNT = 1000 ether;
    uint256 public constant RESERVATION_DURATION = 3600; // 1 hour

    function setUp() public {
        // Deploy token with owner
        vm.startPrank(owner);
        token = new BunkerToken(owner);
        // Deploy escrow
        escrow = new BunkerEscrow(address(token), treasury, owner);
        // Grant OPERATOR_ROLE to operator
        escrow.grantRole(escrow.OPERATOR_ROLE(), operator);
        // Mint tokens to requester
        token.mint(requester, 100_000 ether);
        vm.stopPrank();

        // Requester approves escrow to spend tokens
        vm.prank(requester);
        token.approve(address(escrow), type(uint256).max);
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

    function _defaultProviders() internal view returns (address[3] memory) {
        return [provider0, provider1, provider2];
    }

    // ──────────────────────────────────────────────
    //  Constructor Tests
    // ──────────────────────────────────────────────

    function test_constructor_setsState() public view {
        assertEq(address(escrow.token()), address(token));
        assertEq(escrow.treasury(), treasury);
        assertEq(escrow.protocolFeeBps(), 500);
        assertEq(escrow.nextReservationId(), 1);
        assertEq(escrow.totalBurned(), 0);
        assertEq(escrow.totalTreasuryFees(), 0);
    }

    function test_constructor_grantsAdminRole() public view {
        assertTrue(escrow.hasRole(escrow.DEFAULT_ADMIN_ROLE(), owner));
    }

    function test_constructor_ownerIsSet() public view {
        assertEq(escrow.owner(), owner);
    }

    function test_constructor_revertsZeroToken() public {
        vm.expectRevert(BunkerEscrow.ZeroAddress.selector);
        new BunkerEscrow(address(0), treasury, owner);
    }

    function test_constructor_revertsZeroTreasury() public {
        vm.expectRevert(BunkerEscrow.ZeroAddress.selector);
        new BunkerEscrow(address(token), address(0), owner);
    }

    function test_constructor_revertsZeroOwner() public {
        // OpenZeppelin's Ownable constructor reverts before the ZeroAddress check
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableInvalidOwner.selector, address(0)));
        new BunkerEscrow(address(token), treasury, address(0));
    }

    // ──────────────────────────────────────────────
    //  Reservation Creation Tests
    // ──────────────────────────────────────────────

    function test_createReservation_succeeds() public {
        uint256 balanceBefore = token.balanceOf(requester);

        vm.prank(requester);
        uint256 id = escrow.createReservation(RESERVATION_AMOUNT, RESERVATION_DURATION);

        assertEq(id, 1);
        assertEq(token.balanceOf(requester), balanceBefore - RESERVATION_AMOUNT);
        assertEq(token.balanceOf(address(escrow)), RESERVATION_AMOUNT);
    }

    function test_createReservation_structPopulated() public {
        uint256 id = _createDefaultReservation();

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.requester, requester);
        assertEq(res.totalAmount, uint128(RESERVATION_AMOUNT));
        assertEq(res.releasedAmount, 0);
        assertEq(res.duration, uint48(RESERVATION_DURATION));
        assertEq(res.startTime, 0);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Created));
        assertEq(res.providers[0], address(0));
        assertEq(res.providers[1], address(0));
        assertEq(res.providers[2], address(0));
    }

    function test_createReservation_sequentialIds() public {
        vm.startPrank(requester);
        uint256 id1 = escrow.createReservation(100 ether, 100);
        uint256 id2 = escrow.createReservation(200 ether, 200);
        uint256 id3 = escrow.createReservation(300 ether, 300);
        vm.stopPrank();

        assertEq(id1, 1);
        assertEq(id2, 2);
        assertEq(id3, 3);
        assertEq(escrow.nextReservationId(), 4);
    }

    function test_createReservation_emitsEvent() public {
        vm.expectEmit(true, true, false, true);
        emit BunkerEscrow.ReservationCreated(1, requester, RESERVATION_AMOUNT, RESERVATION_DURATION);

        vm.prank(requester);
        escrow.createReservation(RESERVATION_AMOUNT, RESERVATION_DURATION);
    }

    function test_createReservation_revertsZeroAmount() public {
        vm.prank(requester);
        vm.expectRevert(BunkerEscrow.ZeroAmount.selector);
        escrow.createReservation(0, RESERVATION_DURATION);
    }

    function test_createReservation_revertsZeroDuration() public {
        vm.prank(requester);
        vm.expectRevert(BunkerEscrow.ZeroDuration.selector);
        escrow.createReservation(RESERVATION_AMOUNT, 0);
    }

    function test_createReservation_transfersTokensFromRequester() public {
        uint256 balanceBefore = token.balanceOf(requester);

        _createDefaultReservation();

        assertEq(token.balanceOf(requester), balanceBefore - RESERVATION_AMOUNT);
        assertEq(token.balanceOf(address(escrow)), RESERVATION_AMOUNT);
    }

    function test_createReservation_revertsWhenPaused() public {
        vm.prank(owner);
        escrow.pause();

        vm.prank(requester);
        vm.expectRevert(abi.encodeWithSignature("EnforcedPause()"));
        escrow.createReservation(RESERVATION_AMOUNT, RESERVATION_DURATION);
    }

    // ──────────────────────────────────────────────
    //  Provider Selection Tests
    // ──────────────────────────────────────────────

    function test_selectProviders_succeeds() public {
        uint256 id = _createDefaultReservation();

        vm.prank(operator);
        escrow.selectProviders(id, [provider0, provider1, provider2]);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Active));
        assertEq(res.startTime, uint48(block.timestamp));
        assertEq(res.providers[0], provider0);
        assertEq(res.providers[1], provider1);
        assertEq(res.providers[2], provider2);
    }

    function test_selectProviders_emitsEvent() public {
        uint256 id = _createDefaultReservation();

        vm.expectEmit(true, false, false, true);
        emit BunkerEscrow.ProvidersSelected(id, provider0, provider1, provider2);

        vm.prank(operator);
        escrow.selectProviders(id, [provider0, provider1, provider2]);
    }

    function test_selectProviders_revertsNonOperator() public {
        uint256 id = _createDefaultReservation();

        vm.prank(requester);
        vm.expectRevert();
        escrow.selectProviders(id, [provider0, provider1, provider2]);
    }

    function test_selectProviders_revertsZeroAddressProvider() public {
        uint256 id = _createDefaultReservation();

        vm.prank(operator);
        vm.expectRevert(BunkerEscrow.ZeroAddress.selector);
        escrow.selectProviders(id, [address(0), provider1, provider2]);
    }

    function test_selectProviders_revertsZeroAddressSecondProvider() public {
        uint256 id = _createDefaultReservation();

        vm.prank(operator);
        vm.expectRevert(BunkerEscrow.ZeroAddress.selector);
        escrow.selectProviders(id, [provider0, address(0), provider2]);
    }

    function test_selectProviders_revertsZeroAddressThirdProvider() public {
        uint256 id = _createDefaultReservation();

        vm.prank(operator);
        vm.expectRevert(BunkerEscrow.ZeroAddress.selector);
        escrow.selectProviders(id, [provider0, provider1, address(0)]);
    }

    function test_selectProviders_revertsNonCreatedStatus() public {
        uint256 id = _createAndActivateReservation();

        // Already Active, cannot select again
        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidStatus.selector,
                id,
                BunkerEscrow.Status.Created,
                BunkerEscrow.Status.Active
            )
        );
        escrow.selectProviders(id, [provider0, provider1, provider2]);
    }

    function test_selectProviders_revertsWhenPaused() public {
        uint256 id = _createDefaultReservation();

        vm.prank(owner);
        escrow.pause();

        vm.prank(operator);
        vm.expectRevert(abi.encodeWithSignature("EnforcedPause()"));
        escrow.selectProviders(id, [provider0, provider1, provider2]);
    }

    function test_selectProviders_getProvidersView() public {
        uint256 id = _createAndActivateReservation();

        address[3] memory providers = escrow.getProviders(id);
        assertEq(providers[0], provider0);
        assertEq(providers[1], provider1);
        assertEq(providers[2], provider2);
    }

    // ──────────────────────────────────────────────
    //  Progressive Payment Release Tests
    // ──────────────────────────────────────────────

    function test_releasePayment_halfDuration() public {
        uint256 id = _createAndActivateReservation();

        // Release for 50% of duration
        uint256 settledDuration = RESERVATION_DURATION / 2;
        uint256 grossRelease = (RESERVATION_AMOUNT * settledDuration) / RESERVATION_DURATION;
        // grossRelease = 500 ether

        vm.prank(operator);
        escrow.releasePayment(id, settledDuration);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.releasedAmount, uint128(grossRelease));
    }

    function test_releasePayment_fullDuration() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.releasedAmount, uint128(RESERVATION_AMOUNT));
    }

    function test_releasePayment_progressiveRelease() public {
        uint256 id = _createAndActivateReservation();

        // First release: 25%
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION / 4);

        BunkerEscrow.Reservation memory res1 = escrow.getReservation(id);
        assertEq(res1.releasedAmount, uint128(RESERVATION_AMOUNT / 4));

        // Second release: up to 75%
        vm.prank(operator);
        escrow.releasePayment(id, (RESERVATION_DURATION * 3) / 4);

        BunkerEscrow.Reservation memory res2 = escrow.getReservation(id);
        assertEq(res2.releasedAmount, uint128((RESERVATION_AMOUNT * 3) / 4));
    }

    function test_releasePayment_revertsSettledDurationExceedsTotal() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.SettledDurationExceedsTotal.selector,
                RESERVATION_DURATION + 1,
                RESERVATION_DURATION
            )
        );
        escrow.releasePayment(id, RESERVATION_DURATION + 1);
    }

    function test_releasePayment_revertsNothingToRelease() public {
        uint256 id = _createAndActivateReservation();

        // Release for 50%
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION / 2);

        // Try to release same amount again (settled = 50% again, proportionalTotal <= releasedAmount)
        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.NothingToRelease.selector, id)
        );
        escrow.releasePayment(id, RESERVATION_DURATION / 2);
    }

    function test_releasePayment_revertsNonOperator() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(requester);
        vm.expectRevert();
        escrow.releasePayment(id, RESERVATION_DURATION / 2);
    }

    function test_releasePayment_revertsNonActiveStatus() public {
        uint256 id = _createDefaultReservation();

        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidStatus.selector,
                id,
                BunkerEscrow.Status.Active,
                BunkerEscrow.Status.Created
            )
        );
        escrow.releasePayment(id, RESERVATION_DURATION / 2);
    }

    function test_releasePayment_revertsWhenPaused() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(owner);
        escrow.pause();

        vm.prank(operator);
        vm.expectRevert(abi.encodeWithSignature("EnforcedPause()"));
        escrow.releasePayment(id, RESERVATION_DURATION / 2);
    }

    function test_releasePayment_emitsEvent() public {
        uint256 id = _createAndActivateReservation();
        uint256 grossRelease = RESERVATION_AMOUNT;
        uint256 fee = (grossRelease * 500) / 10000; // 50 ether
        uint256 net = grossRelease - fee; // 950 ether
        uint256 burnAmount = (fee * 8000) / 10000; // 40 ether
        uint256 treasuryAmount = fee - burnAmount; // 10 ether

        vm.expectEmit(true, false, false, true);
        emit BunkerEscrow.PaymentReleased(id, grossRelease, net, fee, burnAmount, treasuryAmount);

        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION);
    }

    // ──────────────────────────────────────────────
    //  Fee Math Verification Tests
    // ──────────────────────────────────────────────

    function test_feeMath_1000BunkerReservation() public {
        uint256 id = _createAndActivateReservation();

        uint256 treasuryBefore = token.balanceOf(treasury);
        uint256 p0Before = token.balanceOf(provider0);
        uint256 p1Before = token.balanceOf(provider1);
        uint256 p2Before = token.balanceOf(provider2);

        // Release full amount (1000 ether)
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION);

        // Fee: 5% of 1000 = 50 ether
        uint256 fee = 50 ether;
        // Burn: 80% of 50 = 40 ether
        uint256 burned = 40 ether;
        // Treasury: 20% of 50 = 10 ether
        uint256 treasuryFee = 10 ether;
        // Net to providers: 950 ether
        uint256 net = 950 ether;
        // Per provider: 950 / 3 = 316.666...666 ether
        uint256 perProvider = net / 3;
        // Dust: 950 - (316.666...666 * 3) = 2 wei
        uint256 dust = net - (perProvider * 3);

        assertEq(escrow.totalBurned(), burned);
        assertEq(escrow.totalTreasuryFees(), treasuryFee);
        assertEq(token.balanceOf(treasury) - treasuryBefore, treasuryFee);
        assertEq(token.balanceOf(provider0) - p0Before, perProvider + dust);
        assertEq(token.balanceOf(provider1) - p1Before, perProvider);
        assertEq(token.balanceOf(provider2) - p2Before, perProvider);

        // Verify exact values
        assertEq(fee, 50 ether);
        assertEq(burned, 40 ether);
        assertEq(treasuryFee, 10 ether);
        assertEq(net, 950 ether);
        assertEq(dust, 2); // 2 wei dust to first provider

        // Verify total distributed == totalAmount
        uint256 totalDistributed = burned + treasuryFee + (perProvider + dust) + perProvider + perProvider;
        assertEq(totalDistributed, RESERVATION_AMOUNT);
    }

    function test_feeMath_protocolFeeCalculation() public view {
        uint256 fee = escrow.calculateProtocolFee(1000 ether);
        assertEq(fee, 50 ether);
    }

    function test_feeMath_zeroProtocolFee() public {
        vm.prank(owner);
        escrow.setProtocolFee(0);

        uint256 id = _createAndActivateReservation();

        uint256 p0Before = token.balanceOf(provider0);
        uint256 p1Before = token.balanceOf(provider1);
        uint256 p2Before = token.balanceOf(provider2);

        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION);

        // No fee: full 1000 ether to providers
        uint256 perProvider = RESERVATION_AMOUNT / 3;
        uint256 dust = RESERVATION_AMOUNT - (perProvider * 3);

        assertEq(escrow.totalBurned(), 0);
        assertEq(escrow.totalTreasuryFees(), 0);
        assertEq(token.balanceOf(provider0) - p0Before, perProvider + dust);
        assertEq(token.balanceOf(provider1) - p1Before, perProvider);
        assertEq(token.balanceOf(provider2) - p2Before, perProvider);
    }

    function test_feeMath_maxProtocolFee() public {
        vm.prank(owner);
        escrow.setProtocolFee(2000); // 20%

        uint256 id = _createAndActivateReservation();

        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION);

        // Fee: 20% of 1000 = 200 ether
        uint256 fee = 200 ether;
        uint256 burned = (fee * 8000) / 10000; // 160 ether
        uint256 treasuryFee = fee - burned; // 40 ether

        assertEq(escrow.totalBurned(), burned);
        assertEq(escrow.totalTreasuryFees(), treasuryFee);
    }

    function test_feeMath_totalBurnedAccumulatesAcrossReservations() public {
        // First reservation
        uint256 id1 = _createAndActivateReservation();
        vm.prank(operator);
        escrow.releasePayment(id1, RESERVATION_DURATION);
        uint256 burnedAfterFirst = escrow.totalBurned();

        // Second reservation
        vm.prank(requester);
        uint256 id2 = escrow.createReservation(RESERVATION_AMOUNT, RESERVATION_DURATION);
        vm.prank(operator);
        escrow.selectProviders(id2, [provider0, provider1, provider2]);
        vm.prank(operator);
        escrow.releasePayment(id2, RESERVATION_DURATION);

        assertEq(escrow.totalBurned(), burnedAfterFirst * 2);
    }

    // ──────────────────────────────────────────────
    //  Finalization Tests
    // ──────────────────────────────────────────────

    function test_finalizeReservation_releasesRemainingAndCompletes() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(operator);
        escrow.finalizeReservation(id);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Completed));
        assertEq(res.releasedAmount, res.totalAmount);
    }

    function test_finalizeReservation_afterPartialRelease() public {
        uint256 id = _createAndActivateReservation();

        // Release 50%
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION / 2);

        uint256 p0Before = token.balanceOf(provider0);

        // Finalize releases remaining 50%
        vm.prank(operator);
        escrow.finalizeReservation(id);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Completed));
        assertEq(res.releasedAmount, res.totalAmount);
        // Provider0 should have received more tokens
        assertGt(token.balanceOf(provider0), p0Before);
    }

    function test_finalizeReservation_emitsEvent() public {
        uint256 id = _createAndActivateReservation();

        vm.expectEmit(true, false, false, true);
        emit BunkerEscrow.ReservationFinalized(id);

        vm.prank(operator);
        escrow.finalizeReservation(id);
    }

    function test_finalizeReservation_revertsNonActiveStatus() public {
        uint256 id = _createDefaultReservation();

        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidStatus.selector,
                id,
                BunkerEscrow.Status.Active,
                BunkerEscrow.Status.Created
            )
        );
        escrow.finalizeReservation(id);
    }

    function test_finalizeReservation_revertsNonOperator() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(requester);
        vm.expectRevert();
        escrow.finalizeReservation(id);
    }

    function test_finalizeReservation_cannotFinalizeCompleted() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(operator);
        escrow.finalizeReservation(id);

        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidStatus.selector,
                id,
                BunkerEscrow.Status.Active,
                BunkerEscrow.Status.Completed
            )
        );
        escrow.finalizeReservation(id);
    }

    // ──────────────────────────────────────────────
    //  Refund Tests
    // ──────────────────────────────────────────────

    function test_refund_requesterCanRefundCreated() public {
        uint256 id = _createDefaultReservation();
        uint256 balanceBefore = token.balanceOf(requester);

        vm.prank(requester);
        escrow.refund(id);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Refunded));
        assertEq(token.balanceOf(requester), balanceBefore + RESERVATION_AMOUNT);
    }

    function test_refund_requesterCanRefundActive() public {
        uint256 id = _createAndActivateReservation();
        uint256 balanceBefore = token.balanceOf(requester);

        vm.prank(requester);
        escrow.refund(id);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Refunded));
        assertEq(token.balanceOf(requester), balanceBefore + RESERVATION_AMOUNT);
    }

    function test_refund_operatorCanRefund() public {
        uint256 id = _createDefaultReservation();
        uint256 balanceBefore = token.balanceOf(requester);

        vm.prank(operator);
        escrow.refund(id);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Refunded));
        assertEq(token.balanceOf(requester), balanceBefore + RESERVATION_AMOUNT);
    }

    function test_refund_afterPartialRelease() public {
        uint256 id = _createAndActivateReservation();

        // Release 50%
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION / 2);

        uint256 released = escrow.getReservation(id).releasedAmount;
        uint256 remaining = RESERVATION_AMOUNT - released;
        uint256 balanceBefore = token.balanceOf(requester);

        vm.prank(requester);
        escrow.refund(id);

        assertEq(token.balanceOf(requester), balanceBefore + remaining);
    }

    function test_refund_emitsEvent() public {
        uint256 id = _createDefaultReservation();

        vm.expectEmit(true, true, false, true);
        emit BunkerEscrow.Refunded(id, requester, RESERVATION_AMOUNT);

        vm.prank(requester);
        escrow.refund(id);
    }

    function test_refund_revertsNonRequesterNonOperator() public {
        uint256 id = _createDefaultReservation();

        address stranger = makeAddr("stranger");
        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.NotRequesterOrOperator.selector,
                stranger,
                id
            )
        );
        escrow.refund(id);
    }

    function test_refund_revertsCompletedReservation() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(operator);
        escrow.finalizeReservation(id);

        vm.prank(requester);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidStatus.selector,
                id,
                BunkerEscrow.Status.Active,
                BunkerEscrow.Status.Completed
            )
        );
        escrow.refund(id);
    }

    function test_refund_revertsRefundedReservation() public {
        uint256 id = _createDefaultReservation();

        vm.prank(requester);
        escrow.refund(id);

        vm.prank(requester);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidStatus.selector,
                id,
                BunkerEscrow.Status.Active,
                BunkerEscrow.Status.Refunded
            )
        );
        escrow.refund(id);
    }

    // ──────────────────────────────────────────────
    //  Dispute Settlement Tests
    // ──────────────────────────────────────────────

    function test_settleDispute_validSplit() public {
        uint256 id = _createAndActivateReservation();

        uint256 requesterAmount = 400 ether;
        uint256 providerAmount = 600 ether;

        uint256 requesterBefore = token.balanceOf(requester);

        vm.prank(operator);
        escrow.settleDispute(id, requesterAmount, providerAmount);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Disputed));
        assertEq(res.releasedAmount, res.totalAmount);
        assertEq(token.balanceOf(requester) - requesterBefore, requesterAmount);
    }

    function test_settleDispute_allToRequester() public {
        uint256 id = _createAndActivateReservation();

        uint256 balanceBefore = token.balanceOf(requester);

        vm.prank(operator);
        escrow.settleDispute(id, RESERVATION_AMOUNT, 0);

        assertEq(token.balanceOf(requester) - balanceBefore, RESERVATION_AMOUNT);
    }

    function test_settleDispute_allToProviders() public {
        uint256 id = _createAndActivateReservation();

        uint256 p0Before = token.balanceOf(provider0);

        vm.prank(operator);
        escrow.settleDispute(id, 0, RESERVATION_AMOUNT);

        // Provider0 should have received some tokens (with fee deducted)
        assertGt(token.balanceOf(provider0), p0Before);
    }

    function test_settleDispute_partialSplitReverts() public {
        uint256 id = _createAndActivateReservation();

        // Partial split (300+200=500 < 1000 remaining) must now revert
        // to prevent permanent token lockup.
        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidDisputeAmounts.selector,
                300 ether,
                200 ether,
                RESERVATION_AMOUNT
            )
        );
        escrow.settleDispute(id, 300 ether, 200 ether);
    }

    function test_settleDispute_revertsExceedsRemaining() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidDisputeAmounts.selector,
                600 ether,
                500 ether,
                RESERVATION_AMOUNT
            )
        );
        escrow.settleDispute(id, 600 ether, 500 ether);
    }

    function test_settleDispute_revertsAfterPartialRelease() public {
        uint256 id = _createAndActivateReservation();

        // Release 50%
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION / 2);

        uint256 remaining = RESERVATION_AMOUNT - RESERVATION_AMOUNT / 2;

        // Dispute split exceeds remaining
        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidDisputeAmounts.selector,
                remaining + 1,
                0,
                remaining
            )
        );
        escrow.settleDispute(id, remaining + 1, 0);
    }

    function test_settleDispute_revertsNonOperator() public {
        uint256 id = _createAndActivateReservation();

        vm.prank(requester);
        vm.expectRevert();
        escrow.settleDispute(id, 500 ether, 500 ether);
    }

    function test_settleDispute_revertsNonActiveStatus() public {
        uint256 id = _createDefaultReservation();

        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidStatus.selector,
                id,
                BunkerEscrow.Status.Active,
                BunkerEscrow.Status.Created
            )
        );
        escrow.settleDispute(id, 500 ether, 500 ether);
    }

    function test_settleDispute_emitsEvent() public {
        uint256 id = _createAndActivateReservation();

        vm.expectEmit(true, false, false, true);
        emit BunkerEscrow.DisputeSettled(id, 400 ether, 600 ether);

        vm.prank(operator);
        escrow.settleDispute(id, 400 ether, 600 ether);
    }

    function test_settleDispute_providerAmountDistributed() public {
        uint256 id = _createAndActivateReservation();
        uint256 providerAmount = 600 ether;

        uint256 p0Before = token.balanceOf(provider0);
        uint256 p1Before = token.balanceOf(provider1);
        uint256 p2Before = token.balanceOf(provider2);
        uint256 treasuryBefore = token.balanceOf(treasury);

        vm.prank(operator);
        escrow.settleDispute(id, 400 ether, providerAmount);

        // Fee: 5% of 600 = 30 ether
        uint256 fee = 30 ether;
        uint256 burned = (fee * 8000) / 10000; // 24 ether
        uint256 treasuryFee = fee - burned; // 6 ether
        uint256 net = providerAmount - fee; // 570 ether
        uint256 perProvider = net / 3;
        uint256 dust = net - (perProvider * 3);

        assertEq(token.balanceOf(treasury) - treasuryBefore, treasuryFee);
        assertEq(token.balanceOf(provider0) - p0Before, perProvider + dust);
        assertEq(token.balanceOf(provider1) - p1Before, perProvider);
        assertEq(token.balanceOf(provider2) - p2Before, perProvider);
        assertEq(escrow.totalBurned(), burned);
    }

    // ──────────────────────────────────────────────
    //  Admin Tests
    // ──────────────────────────────────────────────

    function test_setProtocolFee_succeeds() public {
        vm.prank(owner);
        escrow.setProtocolFee(1000);

        assertEq(escrow.protocolFeeBps(), 1000);
    }

    function test_setProtocolFee_emitsEvent() public {
        vm.expectEmit(false, false, false, true);
        emit BunkerEscrow.ProtocolFeeUpdated(500, 1000);

        vm.prank(owner);
        escrow.setProtocolFee(1000);
    }

    function test_setProtocolFee_zeroAllowed() public {
        vm.prank(owner);
        escrow.setProtocolFee(0);

        assertEq(escrow.protocolFeeBps(), 0);
    }

    function test_setProtocolFee_maxAllowed() public {
        vm.prank(owner);
        escrow.setProtocolFee(2000);

        assertEq(escrow.protocolFeeBps(), 2000);
    }

    function test_setProtocolFee_revertsAboveMax() public {
        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.ProtocolFeeTooHigh.selector, 2001, 2000)
        );
        escrow.setProtocolFee(2001);
    }

    function test_setProtocolFee_revertsNonOwner() public {
        vm.prank(operator);
        vm.expectRevert();
        escrow.setProtocolFee(1000);
    }

    function test_setTreasury_succeeds() public {
        address newTreasury = makeAddr("newTreasury");

        vm.prank(owner);
        escrow.setTreasury(newTreasury);

        assertEq(escrow.treasury(), newTreasury);
    }

    function test_setTreasury_emitsEvent() public {
        address newTreasury = makeAddr("newTreasury");

        vm.expectEmit(true, true, false, true);
        emit BunkerEscrow.TreasuryUpdated(treasury, newTreasury);

        vm.prank(owner);
        escrow.setTreasury(newTreasury);
    }

    function test_setTreasury_revertsZeroAddress() public {
        vm.prank(owner);
        vm.expectRevert(BunkerEscrow.ZeroAddress.selector);
        escrow.setTreasury(address(0));
    }

    function test_setTreasury_revertsNonOwner() public {
        vm.prank(operator);
        vm.expectRevert();
        escrow.setTreasury(makeAddr("newTreasury"));
    }

    function test_pause_succeeds() public {
        vm.prank(owner);
        escrow.pause();

        assertTrue(escrow.paused());
    }

    function test_unpause_succeeds() public {
        vm.prank(owner);
        escrow.pause();

        vm.prank(owner);
        escrow.unpause();

        assertFalse(escrow.paused());
    }

    function test_pause_revertsNonOwner() public {
        vm.prank(operator);
        vm.expectRevert();
        escrow.pause();
    }

    function test_unpause_revertsNonOwner() public {
        vm.prank(owner);
        escrow.pause();

        vm.prank(operator);
        vm.expectRevert();
        escrow.unpause();
    }

    // ──────────────────────────────────────────────
    //  View Function Tests
    // ──────────────────────────────────────────────

    function test_getReservation_nonExistentReturnsDefaults() public view {
        BunkerEscrow.Reservation memory res = escrow.getReservation(999);
        assertEq(res.requester, address(0));
        assertEq(res.totalAmount, 0);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.None));
    }

    function test_version() public view {
        assertEq(keccak256(bytes(escrow.VERSION())), keccak256(bytes("1.2.0")));
    }

    function test_constants() public view {
        assertEq(escrow.PROVIDERS_PER_RESERVATION(), 3);
        assertEq(escrow.MAX_PROTOCOL_FEE_BPS(), 2000);
        assertEq(escrow.feeBurnBps(), 8000);
        assertEq(escrow.feeTreasuryBps(), 2000);
        assertEq(escrow.BPS_DENOMINATOR(), 10000);
    }

    // ──────────────────────────────────────────────
    //  End-to-End Lifecycle Tests
    // ──────────────────────────────────────────────

    function test_lifecycle_createSelectReleaseFinalize() public {
        // Create
        uint256 id = _createDefaultReservation();
        assertEq(uint8(escrow.getReservation(id).status), uint8(BunkerEscrow.Status.Created));

        // Select providers
        vm.prank(operator);
        escrow.selectProviders(id, [provider0, provider1, provider2]);
        assertEq(uint8(escrow.getReservation(id).status), uint8(BunkerEscrow.Status.Active));

        // Progressive releases
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION / 4);

        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION / 2);

        vm.prank(operator);
        escrow.releasePayment(id, (RESERVATION_DURATION * 3) / 4);

        // Finalize
        vm.prank(operator);
        escrow.finalizeReservation(id);
        assertEq(uint8(escrow.getReservation(id).status), uint8(BunkerEscrow.Status.Completed));

        // All tokens distributed
        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.releasedAmount, res.totalAmount);
    }

    function test_lifecycle_createSelectRefund() public {
        uint256 id = _createAndActivateReservation();

        // Partial release
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION / 2);

        uint256 balanceBefore = token.balanceOf(requester);
        uint256 remaining = RESERVATION_AMOUNT - escrow.getReservation(id).releasedAmount;

        // Refund remainder
        vm.prank(requester);
        escrow.refund(id);

        assertEq(token.balanceOf(requester) - balanceBefore, remaining);
        assertEq(uint8(escrow.getReservation(id).status), uint8(BunkerEscrow.Status.Refunded));
    }

    function test_lifecycle_createSelectDispute() public {
        uint256 id = _createAndActivateReservation();

        // Release 25%
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION / 4);

        uint256 remaining = RESERVATION_AMOUNT - escrow.getReservation(id).releasedAmount;

        // Settle dispute: exact split of remaining (handle odd remaining)
        uint256 half = remaining / 2;
        vm.prank(operator);
        escrow.settleDispute(id, half, remaining - half);

        assertEq(uint8(escrow.getReservation(id).status), uint8(BunkerEscrow.Status.Disputed));
    }

    // ──────────────────────────────────────────────
    //  Multiple Reservations Test
    // ──────────────────────────────────────────────

    function test_multipleReservations_independent() public {
        uint256 id1 = _createAndActivateReservation();

        vm.prank(requester);
        uint256 id2 = escrow.createReservation(500 ether, 7200);
        address otherProvider0 = makeAddr("otherP0");
        address otherProvider1 = makeAddr("otherP1");
        address otherProvider2 = makeAddr("otherP2");
        vm.prank(operator);
        escrow.selectProviders(id2, [otherProvider0, otherProvider1, otherProvider2]);

        // Finalize first
        vm.prank(operator);
        escrow.finalizeReservation(id1);

        // Refund second
        vm.prank(operator);
        escrow.refund(id2);

        assertEq(uint8(escrow.getReservation(id1).status), uint8(BunkerEscrow.Status.Completed));
        assertEq(uint8(escrow.getReservation(id2).status), uint8(BunkerEscrow.Status.Refunded));
    }

    // ──────────────────────────────────────────────
    //  Edge Case Tests
    // ──────────────────────────────────────────────

    function test_edge_smallAmountReservation() public {
        // Very small amount: 3 wei (minimum that splits among 3 providers)
        vm.prank(requester);
        uint256 id = escrow.createReservation(3, 100);

        vm.prank(operator);
        escrow.selectProviders(id, [provider0, provider1, provider2]);

        // With 5% fee on 3 wei: fee = 0 (integer division)
        // All goes to providers: 1 wei each
        vm.prank(operator);
        escrow.finalizeReservation(id);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Completed));
    }

    function test_edge_releaseZeroDurationSettled() public {
        uint256 id = _createAndActivateReservation();

        // settledDuration = 0 means proportionalTotal = 0 <= releasedAmount (0)
        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.NothingToRelease.selector, id)
        );
        escrow.releasePayment(id, 0);
    }

    function test_edge_disputeZeroAmounts_reverts() public {
        uint256 id = _createAndActivateReservation();

        // Both zero (0+0 != remaining) must revert to prevent lockup.
        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEscrow.InvalidDisputeAmounts.selector,
                0,
                0,
                RESERVATION_AMOUNT
            )
        );
        escrow.settleDispute(id, 0, 0);
    }

    function test_edge_refundAfterFullRelease() public {
        uint256 id = _createAndActivateReservation();

        // Release everything
        vm.prank(operator);
        escrow.releasePayment(id, RESERVATION_DURATION);

        // Refund: remaining is 0, but still changes status
        vm.prank(requester);
        escrow.refund(id);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Refunded));
    }

    // ──────────────────────────────────────────────
    //  Access Control Tests
    // ──────────────────────────────────────────────

    function test_operatorRole_grantAndRevoke() public {
        address newOperator = makeAddr("newOperator");

        vm.startPrank(owner);
        escrow.grantRole(escrow.OPERATOR_ROLE(), newOperator);
        assertTrue(escrow.hasRole(escrow.OPERATOR_ROLE(), newOperator));

        escrow.revokeRole(escrow.OPERATOR_ROLE(), newOperator);
        assertFalse(escrow.hasRole(escrow.OPERATOR_ROLE(), newOperator));
        vm.stopPrank();
    }

    function test_operatorRole_cannotGrantWithoutAdmin() public {
        address newOp = makeAddr("newOp");
        bytes32 operatorRole = escrow.OPERATOR_ROLE();
        bytes32 adminRole = escrow.DEFAULT_ADMIN_ROLE();
        vm.prank(operator);
        vm.expectRevert(
            abi.encodeWithSelector(
                IAccessControl.AccessControlUnauthorizedAccount.selector,
                operator,
                adminRole
            )
        );
        escrow.grantRole(operatorRole, newOp);
    }

    // ──────────────────────────────────────────────
    //  Fuzz Tests
    // ──────────────────────────────────────────────

    function testFuzz_createReservation(uint128 amount, uint48 duration) public {
        vm.assume(amount > 0 && amount <= 100_000 ether);
        vm.assume(duration > 0);

        vm.prank(requester);
        uint256 id = escrow.createReservation(amount, duration);

        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.requester, requester);
        assertEq(res.totalAmount, amount);
        assertEq(res.duration, duration);
    }

    function testFuzz_releasePayment_proportional(uint48 settledDuration) public {
        vm.assume(settledDuration > 0 && settledDuration <= RESERVATION_DURATION);

        uint256 id = _createAndActivateReservation();

        vm.prank(operator);
        escrow.releasePayment(id, settledDuration);

        uint256 expectedRelease = (RESERVATION_AMOUNT * settledDuration) / RESERVATION_DURATION;
        BunkerEscrow.Reservation memory res = escrow.getReservation(id);
        assertEq(res.releasedAmount, uint128(expectedRelease));
    }

    function testFuzz_setProtocolFee_withinRange(uint256 feeBps) public {
        vm.assume(feeBps <= 2000);

        vm.prank(owner);
        escrow.setProtocolFee(feeBps);

        assertEq(escrow.protocolFeeBps(), feeBps);
    }

    function testFuzz_setProtocolFee_aboveMaxReverts(uint256 feeBps) public {
        vm.assume(feeBps > 2000);

        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEscrow.ProtocolFeeTooHigh.selector, feeBps, 2000)
        );
        escrow.setProtocolFee(feeBps);
    }

    function testFuzz_disputeSettlement(uint128 requesterAmt) public {
        vm.assume(uint256(requesterAmt) <= RESERVATION_AMOUNT);
        uint256 providerAmt = RESERVATION_AMOUNT - uint256(requesterAmt);

        uint256 id = _createAndActivateReservation();

        vm.prank(operator);
        escrow.settleDispute(id, requesterAmt, providerAmt);

        assertEq(uint8(escrow.getReservation(id).status), uint8(BunkerEscrow.Status.Disputed));
    }

    // ──────────────────────────────────────────────
    //  setFeeSplit Tests
    // ──────────────────────────────────────────────

    function test_setFeeSplit_succeeds() public {
        vm.prank(owner);
        escrow.setFeeSplit(7000, 3000);
        assertEq(escrow.feeBurnBps(), 7000);
        assertEq(escrow.feeTreasuryBps(), 3000);
    }

    function test_setFeeSplit_revertsInvalidSum() public {
        vm.prank(owner);
        vm.expectRevert(BunkerEscrow.InvalidFeeSplit.selector);
        escrow.setFeeSplit(7000, 2000);
    }

    function test_setFeeSplit_revertsNonOwner() public {
        vm.prank(requester);
        vm.expectRevert();
        escrow.setFeeSplit(7000, 3000);
    }

    function test_setFeeSplit_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerEscrow.FeeSplitUpdated(6000, 4000);
        escrow.setFeeSplit(6000, 4000);
    }

    function test_setFeeSplit_allBurn() public {
        vm.prank(owner);
        escrow.setFeeSplit(10000, 0);
        assertEq(escrow.feeBurnBps(), 10000);
        assertEq(escrow.feeTreasuryBps(), 0);
    }

    function test_setFeeSplit_allTreasury() public {
        vm.prank(owner);
        escrow.setFeeSplit(0, 10000);
        assertEq(escrow.feeBurnBps(), 0);
        assertEq(escrow.feeTreasuryBps(), 10000);
    }
}
