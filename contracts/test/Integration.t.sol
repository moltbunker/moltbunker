// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEscrow.sol";
import "../src/BunkerPricing.sol";

/// @title IntegrationTest
/// @notice End-to-end tests combining all Moltbunker contracts.
/// @dev Simulates real-world flows: staking → pricing → escrow → payment.
contract IntegrationTest is Test {
    BunkerToken token;
    BunkerStaking staking;
    BunkerEscrow escrow;
    BunkerPricing pricing;

    address owner = makeAddr("owner");
    address treasury = makeAddr("treasury");
    address operator = makeAddr("operator");
    address slasher = makeAddr("slasher");
    address requester = makeAddr("requester");
    address provider0 = makeAddr("provider0");
    address provider1 = makeAddr("provider1");
    address provider2 = makeAddr("provider2");

    uint256 constant MINT_AMOUNT = 200_000_000e18;

    function setUp() public {
        vm.startPrank(owner);

        // Deploy all contracts
        token = new BunkerToken(owner);
        staking = new BunkerStaking(address(token), treasury, owner);
        escrow = new BunkerEscrow(address(token), treasury, owner);
        pricing = new BunkerPricing(owner);

        // Grant roles
        escrow.grantRole(escrow.OPERATOR_ROLE(), operator);
        staking.grantRole(staking.SLASHER_ROLE(), slasher);
        staking.setSlashingEnabled(true);

        // Mint tokens
        token.mint(requester, MINT_AMOUNT);
        token.mint(provider0, MINT_AMOUNT);
        token.mint(provider1, MINT_AMOUNT);
        token.mint(provider2, MINT_AMOUNT);

        vm.stopPrank();

        // Providers approve staking contract
        vm.prank(provider0);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider1);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider2);
        token.approve(address(staking), type(uint256).max);

        // Requester approves escrow contract
        vm.prank(requester);
        token.approve(address(escrow), type(uint256).max);
    }

    // ──────────────────────────────────────────────
    //  Full Lifecycle: Stake → Price → Escrow → Pay
    // ──────────────────────────────────────────────

    function test_FullDeploymentLifecycle() public {
        // 1. Providers stake to become eligible
        vm.prank(provider0);
        staking.stake(10_000_000e18); // Silver tier

        vm.prank(provider1);
        staking.stake(10_000_000e18);

        vm.prank(provider2);
        staking.stake(10_000_000e18);

        // Verify all providers are active at Silver tier
        assertTrue(staking.isActiveProvider(provider0));
        assertTrue(staking.isActiveProvider(provider1));
        assertTrue(staking.isActiveProvider(provider2));
        assertEq(uint256(staking.getTier(provider0)), uint256(BunkerStaking.Tier.Silver));

        // 2. Calculate deployment cost using pricing contract
        BunkerPricing.ResourceRequest memory req = BunkerPricing.ResourceRequest({
            cpuCores: 2000,      // 2 cores
            memoryGB: 4000,      // 4 GB
            storageGB: 20,       // 20 GB
            networkGB: 10,       // 10 GB
            durationHours: 720,  // 30 days
            useGPUBasic: false,
            useGPUPremium: false,
            useTor: false,
            usePremiumSLA: false,
            useSpot: false
        });
        uint256 cost = pricing.calculateCost(req);
        assertGt(cost, 0, "Cost should be > 0");

        // 3. Requester creates escrow reservation
        vm.prank(requester);
        uint256 reservationId = escrow.createReservation(cost, 720 hours);
        assertEq(reservationId, 1);

        // Verify tokens moved to escrow
        uint256 escrowBalance = token.balanceOf(address(escrow));
        assertEq(escrowBalance, cost);

        // 4. Operator selects providers
        vm.prank(operator);
        escrow.selectProviders(reservationId, [provider0, provider1, provider2]);

        // Verify reservation is Active
        BunkerEscrow.Reservation memory res = escrow.getReservation(reservationId);
        assertEq(uint256(res.status), uint256(BunkerEscrow.Status.Active));

        // 5. Progressive payment: release after 360 hours (50%)
        uint256 provider0BalanceBefore = token.balanceOf(provider0);
        uint256 treasuryBalanceBefore = token.balanceOf(treasury);

        vm.prank(operator);
        escrow.releasePayment(reservationId, 360 hours);

        // Verify providers received payment (minus 5% fee)
        uint256 halfPayment = cost / 2;
        uint256 fee = (halfPayment * 500) / 10000; // 5%
        uint256 netToProviders = halfPayment - fee;
        uint256 perProvider = netToProviders / 3;

        // Provider0 gets perProvider + dust
        uint256 dust = netToProviders - (perProvider * 3);
        assertEq(
            token.balanceOf(provider0) - provider0BalanceBefore,
            perProvider + dust,
            "Provider0 should get share + dust"
        );

        // Treasury gets 20% of fee
        uint256 burnAmount = (fee * 8000) / 10000;
        uint256 treasuryAmount = fee - burnAmount;
        assertEq(
            token.balanceOf(treasury) - treasuryBalanceBefore,
            treasuryAmount,
            "Treasury should get 20% of fee"
        );

        // 6. Finalize reservation (release remaining 50%)
        vm.prank(operator);
        escrow.finalizeReservation(reservationId);

        // Verify status is Completed
        res = escrow.getReservation(reservationId);
        assertEq(uint256(res.status), uint256(BunkerEscrow.Status.Completed));

        // Verify escrow balance is zero (all distributed)
        assertEq(token.balanceOf(address(escrow)), 0, "Escrow should be empty");
    }

    // ──────────────────────────────────────────────
    //  Staking + Slashing + Escrow Interaction
    // ──────────────────────────────────────────────

    function test_SlashingDuringActiveDeployment() public {
        // Provider stakes
        vm.prank(provider0);
        staking.stake(100_000_000e18); // Gold tier
        vm.prank(provider1);
        staking.stake(10_000_000e18);
        vm.prank(provider2);
        staking.stake(10_000_000e18);

        // Requester creates and starts deployment
        vm.prank(requester);
        uint256 reservationId = escrow.createReservation(1000e18, 24 hours);

        vm.prank(operator);
        escrow.selectProviders(reservationId, [provider0, provider1, provider2]);

        // Provider0 misbehaves — slash 10% of stake
        uint256 slashAmount = 10_000_000e18;
        uint256 treasuryBefore = token.balanceOf(treasury);

        vm.prank(slasher);
        staking.slash(provider0, slashAmount);

        // Verify provider0 is still active but dropped to Silver tier (90M < 100M Gold threshold)
        assertTrue(staking.isActiveProvider(provider0));
        assertEq(uint256(staking.getTier(provider0)), uint256(BunkerStaking.Tier.Silver));

        // Verify slash distribution: 80% burned, 20% to treasury
        uint256 expectedTreasury = (slashAmount * 2000) / 10000; // 20%
        assertEq(
            token.balanceOf(treasury) - treasuryBefore,
            expectedTreasury,
            "Treasury should receive 20% of slash"
        );

        // Deployment can still be finalized
        vm.prank(operator);
        escrow.finalizeReservation(reservationId);

        BunkerEscrow.Reservation memory res = escrow.getReservation(reservationId);
        assertEq(uint256(res.status), uint256(BunkerEscrow.Status.Completed));
    }

    // ──────────────────────────────────────────────
    //  Refund Flow
    // ──────────────────────────────────────────────

    function test_RefundBeforeProviderSelection() public {
        uint256 requesterBalanceBefore = token.balanceOf(requester);

        // Create reservation
        vm.prank(requester);
        uint256 reservationId = escrow.createReservation(5000e18, 48 hours);

        // Requester changes mind, refunds before providers are selected
        vm.prank(requester);
        escrow.refund(reservationId);

        // Full amount returned
        assertEq(token.balanceOf(requester), requesterBalanceBefore, "Full refund expected");

        BunkerEscrow.Reservation memory res = escrow.getReservation(reservationId);
        assertEq(uint256(res.status), uint256(BunkerEscrow.Status.Refunded));
    }

    function test_PartialRefundAfterUsage() public {
        // Create and start deployment
        vm.prank(requester);
        uint256 reservationId = escrow.createReservation(1000e18, 100 hours);

        // Stake providers first
        vm.prank(provider0);
        staking.stake(1_000_000e18);
        vm.prank(provider1);
        staking.stake(1_000_000e18);
        vm.prank(provider2);
        staking.stake(1_000_000e18);

        vm.prank(operator);
        escrow.selectProviders(reservationId, [provider0, provider1, provider2]);

        // Release 30% of the deployment
        vm.prank(operator);
        escrow.releasePayment(reservationId, 30 hours);

        // Then requester requests early termination/refund
        uint256 requesterBalanceBefore = token.balanceOf(requester);

        vm.prank(requester);
        escrow.refund(reservationId);

        // Requester gets back ~70% (minus what was released)
        BunkerEscrow.Reservation memory res = escrow.getReservation(reservationId);
        assertEq(uint256(res.status), uint256(BunkerEscrow.Status.Refunded));

        // Remaining tokens returned to requester
        uint256 released = (1000e18 * 30) / 100; // 300e18
        uint256 expectedRefund = 1000e18 - released;
        assertEq(
            token.balanceOf(requester) - requesterBalanceBefore,
            expectedRefund,
            "Should refund unreleased amount"
        );
    }

    // ──────────────────────────────────────────────
    //  Dispute Resolution
    // ──────────────────────────────────────────────

    function test_DisputeSettlement5050() public {
        // Stake providers
        vm.prank(provider0);
        staking.stake(1_000_000e18);
        vm.prank(provider1);
        staking.stake(1_000_000e18);
        vm.prank(provider2);
        staking.stake(1_000_000e18);

        // Create and activate reservation
        vm.prank(requester);
        uint256 reservationId = escrow.createReservation(1000e18, 24 hours);

        vm.prank(operator);
        escrow.selectProviders(reservationId, [provider0, provider1, provider2]);

        // Dispute: 50/50 split of remaining
        uint256 requesterBefore = token.balanceOf(requester);

        vm.prank(operator);
        escrow.settleDispute(reservationId, 500e18, 500e18);

        // Requester gets 500
        assertEq(token.balanceOf(requester) - requesterBefore, 500e18);

        // Providers split 500 (minus fee)
        BunkerEscrow.Reservation memory res = escrow.getReservation(reservationId);
        assertEq(uint256(res.status), uint256(BunkerEscrow.Status.Disputed));
    }

    // ──────────────────────────────────────────────
    //  Pricing → Escrow Integration
    // ──────────────────────────────────────────────

    function test_PricingCalculationFeedsIntoEscrow() public {
        // Admin updates pricing
        vm.prank(owner);
        pricing.setCPUPrice(1e18); // Double default CPU price

        // Calculate cost for a workload
        BunkerPricing.ResourceRequest memory req = BunkerPricing.ResourceRequest({
            cpuCores: 4000,       // 4 cores
            memoryGB: 8000,       // 8 GB
            storageGB: 50,
            networkGB: 100,
            durationHours: 24,
            useGPUBasic: false,
            useGPUPremium: false,
            useTor: true,
            usePremiumSLA: false,
            useSpot: false
        });
        uint256 cost = pricing.calculateCost(req);

        // Use that cost to create an escrow reservation
        vm.prank(requester);
        uint256 reservationId = escrow.createReservation(cost, 24 hours);

        BunkerEscrow.Reservation memory res = escrow.getReservation(reservationId);
        assertEq(res.totalAmount, uint128(cost));
        assertEq(res.requester, requester);
    }

    // ──────────────────────────────────────────────
    //  Multi-Reservation Concurrent Test
    // ──────────────────────────────────────────────

    function test_MultipleReservationsConcurrently() public {
        // Stake providers
        vm.prank(provider0);
        staking.stake(1_000_000e18);
        vm.prank(provider1);
        staking.stake(1_000_000e18);
        vm.prank(provider2);
        staking.stake(1_000_000e18);

        // Create 3 reservations
        vm.startPrank(requester);
        uint256 res1 = escrow.createReservation(1000e18, 24 hours);
        uint256 res2 = escrow.createReservation(2000e18, 48 hours);
        uint256 res3 = escrow.createReservation(3000e18, 72 hours);
        vm.stopPrank();

        assertEq(res1, 1);
        assertEq(res2, 2);
        assertEq(res3, 3);

        // Activate all 3
        vm.startPrank(operator);
        escrow.selectProviders(res1, [provider0, provider1, provider2]);
        escrow.selectProviders(res2, [provider0, provider1, provider2]);
        escrow.selectProviders(res3, [provider0, provider1, provider2]);

        // Release payments for each at different rates
        escrow.releasePayment(res1, 12 hours);  // 50%
        escrow.releasePayment(res2, 24 hours);  // 50%
        escrow.releasePayment(res3, 24 hours);  // 33%

        // Finalize res1, refund res2, dispute res3
        escrow.finalizeReservation(res1);
        escrow.refund(res2);
        escrow.settleDispute(res3, 1000e18, 1000e18);
        vm.stopPrank();

        // Verify final statuses
        assertEq(uint256(escrow.getReservation(res1).status), uint256(BunkerEscrow.Status.Completed));
        assertEq(uint256(escrow.getReservation(res2).status), uint256(BunkerEscrow.Status.Refunded));
        assertEq(uint256(escrow.getReservation(res3).status), uint256(BunkerEscrow.Status.Disputed));
    }

    // ──────────────────────────────────────────────
    //  Unstaking Flow with Timelock
    // ──────────────────────────────────────────────

    function test_FullUnstakingLifecycle() public {
        // Provider stakes Gold tier
        vm.prank(provider0);
        staking.stake(100_000_000e18);

        assertEq(uint256(staking.getTier(provider0)), uint256(BunkerStaking.Tier.Gold));

        // Request partial unstake
        vm.prank(provider0);
        staking.requestUnstake(50_000_000e18);

        // Still active at Silver tier (50M remaining)
        assertTrue(staking.isActiveProvider(provider0));

        // Try to complete before 14 days — should fail
        vm.prank(provider0);
        vm.expectRevert();
        staking.completeUnstake(0);

        // Warp 14 days
        vm.warp(block.timestamp + 14 days);

        // Complete unstake
        uint256 balanceBefore = token.balanceOf(provider0);
        vm.prank(provider0);
        staking.completeUnstake(0);

        assertEq(token.balanceOf(provider0) - balanceBefore, 50_000_000e18);
    }

    // ──────────────────────────────────────────────
    //  Token Supply Integrity
    // ──────────────────────────────────────────────

    function test_TokenSupplyAfterBurns() public {
        uint256 initialSupply = token.totalSupply();

        // Stake provider
        vm.prank(provider0);
        staking.stake(1_000_000e18);

        // Slash provider (burns 80% of 100_000)
        vm.prank(slasher);
        staking.slash(provider0, 100_000e18);

        uint256 burned = (100_000e18 * 8000) / 10000; // 80_000e18
        assertEq(token.totalSupply(), initialSupply - burned, "Supply should decrease by burned amount");

        // Escrow payment also burns
        vm.prank(provider0);
        staking.stake(1_000_000e18); // re-stake to stay active
        vm.prank(provider1);
        staking.stake(1_000_000e18);
        vm.prank(provider2);
        staking.stake(1_000_000e18);

        vm.prank(requester);
        uint256 resId = escrow.createReservation(10_000e18, 24 hours);

        vm.prank(operator);
        escrow.selectProviders(resId, [provider0, provider1, provider2]);

        uint256 supplyBefore = token.totalSupply();
        vm.prank(operator);
        escrow.finalizeReservation(resId);

        // 5% fee on 10000 = 500, 80% burned = 400
        uint256 escrowBurn = (10_000e18 * 500 / 10000) * 8000 / 10000;
        assertEq(token.totalSupply(), supplyBefore - escrowBurn, "Supply should decrease by escrow burn");
    }
}
