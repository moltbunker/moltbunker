// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerReputation.sol";
import "../src/BunkerDelegation.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerVerification.sol";

// ================================================================
//  BunkerReputation Admin Setter Tests
// ================================================================

contract BunkerReputationSetterTest is Test {
    BunkerReputation public rep;

    address public owner = makeAddr("owner");
    address public reporter = makeAddr("reporter");
    address public provider1 = makeAddr("provider1");

    function setUp() public {
        vm.startPrank(owner);
        rep = new BunkerReputation(owner);
        rep.grantRole(rep.REPORTER_ROLE(), reporter);
        vm.stopPrank();
    }

    // -- setTierThresholds --

    function test_setTierThresholds_succeeds() public {
        vm.prank(owner);
        rep.setTierThresholds(200, 450, 700, 850);
        assertEq(rep.tierProbation(), 200);
        assertEq(rep.tierStandard(), 450);
        assertEq(rep.tierTrusted(), 700);
        assertEq(rep.tierElite(), 850);
    }

    function test_setTierThresholds_revertsNotAscending() public {
        vm.prank(owner);
        vm.expectRevert(BunkerReputation.ThresholdsNotAscending.selector);
        rep.setTierThresholds(500, 500, 750, 900);
    }

    function test_setTierThresholds_revertsEliteExceedsMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerReputation.ThresholdExceedsMax.selector);
        rep.setTierThresholds(250, 500, 750, 1001);
    }

    function test_setTierThresholds_revertsNonOwner() public {
        vm.prank(reporter);
        vm.expectRevert();
        rep.setTierThresholds(200, 450, 700, 850);
    }

    function test_setTierThresholds_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerReputation.TierThresholdsUpdated(200, 450, 700, 850);
        rep.setTierThresholds(200, 450, 700, 850);
    }

    function test_setTierThresholds_affectsGetTier() public {
        // Register provider
        vm.prank(reporter);
        rep.registerProvider(provider1);

        // Default: score=500 is Standard (threshold 500)
        assertEq(uint8(rep.getTier(provider1)), uint8(BunkerReputation.ReputationTier.Standard));

        // Lower the Standard threshold to 600 — provider at 500 drops to Probation
        vm.prank(owner);
        rep.setTierThresholds(200, 600, 750, 900);
        assertEq(uint8(rep.getTier(provider1)), uint8(BunkerReputation.ReputationTier.Probation));
    }

    // -- setDecayParams --

    function test_setDecayParams_succeeds() public {
        vm.prank(owner);
        rep.setDecayParams(5, 200);
        assertEq(rep.decayRate(), 5);
        assertEq(rep.decayFloor(), 200);
    }

    function test_setDecayParams_revertsFloorTooHigh() public {
        vm.prank(owner);
        vm.expectRevert(BunkerReputation.FloorTooHigh.selector);
        rep.setDecayParams(1, 501);
    }

    function test_setDecayParams_revertsNonOwner() public {
        vm.prank(reporter);
        vm.expectRevert();
        rep.setDecayParams(5, 200);
    }

    function test_setDecayParams_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerReputation.DecayParamsUpdated(3, 150);
        rep.setDecayParams(3, 150);
    }

    // -- setMinScoreForJobs --

    function test_setMinScoreForJobs_succeeds() public {
        vm.prank(owner);
        rep.setMinScoreForJobs(300);
        assertEq(rep.minScoreForJobs(), 300);
    }

    function test_setMinScoreForJobs_revertsExceedsMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerReputation.ScoreExceedsMax.selector);
        rep.setMinScoreForJobs(1001);
    }

    function test_setMinScoreForJobs_revertsNonOwner() public {
        vm.prank(reporter);
        vm.expectRevert();
        rep.setMinScoreForJobs(300);
    }

    function test_setMinScoreForJobs_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerReputation.MinScoreForJobsUpdated(300);
        rep.setMinScoreForJobs(300);
    }

    function test_setMinScoreForJobs_affectsEligibility() public {
        vm.prank(reporter);
        rep.registerProvider(provider1);

        // Default: score=500, min=250 → eligible
        assertTrue(rep.isEligibleForJobs(provider1));

        // Raise min to 600 → not eligible
        vm.prank(owner);
        rep.setMinScoreForJobs(600);
        assertFalse(rep.isEligibleForJobs(provider1));
    }

    // -- setMaxCustomDelta --

    function test_setMaxCustomDelta_succeeds() public {
        vm.prank(owner);
        rep.setMaxCustomDelta(300);
        assertEq(rep.maxCustomDelta(), 300);
        assertEq(rep.minCustomDelta(), -300);
    }

    function test_setMaxCustomDelta_revertsZero() public {
        vm.prank(owner);
        vm.expectRevert(BunkerReputation.InvalidDelta.selector);
        rep.setMaxCustomDelta(0);
    }

    function test_setMaxCustomDelta_revertsNegative() public {
        vm.prank(owner);
        vm.expectRevert(BunkerReputation.InvalidDelta.selector);
        rep.setMaxCustomDelta(-1);
    }

    function test_setMaxCustomDelta_revertsExceedsHalfMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerReputation.InvalidDelta.selector);
        rep.setMaxCustomDelta(501);
    }

    function test_setMaxCustomDelta_revertsNonOwner() public {
        vm.prank(reporter);
        vm.expectRevert();
        rep.setMaxCustomDelta(300);
    }

    function test_setMaxCustomDelta_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerReputation.MaxCustomDeltaUpdated(300);
        rep.setMaxCustomDelta(300);
    }
}

// ================================================================
//  BunkerDelegation Admin Setter Tests
// ================================================================

contract BunkerDelegationSetterTest is Test {
    BunkerToken public token;
    BunkerDelegation public delegation;
    BunkerStaking public staking;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public provider1 = makeAddr("provider1");
    address public delegator1 = makeAddr("delegator1");

    function setUp() public {
        vm.startPrank(owner);
        token = new BunkerToken(owner);
        staking = new BunkerStaking(address(token), treasury, owner);
        delegation = new BunkerDelegation(address(token), address(staking), owner);
        token.mint(provider1, 10_000_000e18);
        token.mint(delegator1, 10_000_000e18);
        vm.stopPrank();

        vm.prank(provider1);
        token.approve(address(staking), type(uint256).max);
        vm.prank(delegator1);
        token.approve(address(delegation), type(uint256).max);
    }

    // -- setUnbondingPeriod --

    function test_setUnbondingPeriod_succeeds() public {
        vm.prank(owner);
        delegation.setUnbondingPeriod(14 days);
        assertEq(delegation.unbondingPeriod(), 14 days);
    }

    function test_setUnbondingPeriod_revertsBelowMin() public {
        vm.prank(owner);
        vm.expectRevert(BunkerDelegation.InvalidUnbondingPeriod.selector);
        delegation.setUnbondingPeriod(30);
    }

    function test_setUnbondingPeriod_revertsAboveMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerDelegation.InvalidUnbondingPeriod.selector);
        delegation.setUnbondingPeriod(31 days);
    }

    function test_setUnbondingPeriod_revertsNonOwner() public {
        vm.prank(delegator1);
        vm.expectRevert();
        delegation.setUnbondingPeriod(14 days);
    }

    function test_setUnbondingPeriod_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerDelegation.UnbondingPeriodUpdated(10 days);
        delegation.setUnbondingPeriod(10 days);
    }

    // -- setMaxRewardCutBps --

    function test_setMaxRewardCutBps_succeeds() public {
        vm.prank(owner);
        delegation.setMaxRewardCutBps(8000);
        assertEq(delegation.maxRewardCutBps(), 8000);
    }

    function test_setMaxRewardCutBps_revertsZero() public {
        vm.prank(owner);
        vm.expectRevert(BunkerDelegation.InvalidRewardCutCap.selector);
        delegation.setMaxRewardCutBps(0);
    }

    function test_setMaxRewardCutBps_revertsAboveMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerDelegation.InvalidRewardCutCap.selector);
        delegation.setMaxRewardCutBps(10001);
    }

    function test_setMaxRewardCutBps_revertsNonOwner() public {
        vm.prank(delegator1);
        vm.expectRevert();
        delegation.setMaxRewardCutBps(8000);
    }

    function test_setMaxRewardCutBps_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerDelegation.MaxRewardCutUpdated(8000);
        delegation.setMaxRewardCutBps(8000);
    }

    function test_setMaxRewardCutBps_allowsFullRange() public {
        vm.prank(owner);
        delegation.setMaxRewardCutBps(10000);
        assertEq(delegation.maxRewardCutBps(), 10000);
    }
}

// ================================================================
//  BunkerVerification Admin Setter Tests
// ================================================================

contract BunkerVerificationSetterTest is Test {
    BunkerVerification public verification;

    address public owner = makeAddr("owner");
    address public verifier = makeAddr("verifier");
    address public provider1 = makeAddr("provider1");

    function setUp() public {
        vm.startPrank(owner);
        verification = new BunkerVerification(owner);
        verification.grantRole(verification.VERIFIER_ROLE(), verifier);
        vm.stopPrank();
    }

    // -- setReinstatementCooldown --

    function test_setReinstatementCooldown_succeeds() public {
        vm.prank(owner);
        verification.setReinstatementCooldown(14 days);
        assertEq(verification.reinstatementCooldown(), 14 days);
    }

    function test_setReinstatementCooldown_revertsBelowMin() public {
        vm.prank(owner);
        vm.expectRevert(BunkerVerification.InvalidCooldown.selector);
        verification.setReinstatementCooldown(30);
    }

    function test_setReinstatementCooldown_revertsAboveMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerVerification.InvalidCooldown.selector);
        verification.setReinstatementCooldown(31 days);
    }

    function test_setReinstatementCooldown_revertsNonOwner() public {
        vm.prank(verifier);
        vm.expectRevert();
        verification.setReinstatementCooldown(14 days);
    }

    function test_setReinstatementCooldown_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerVerification.ReinstatementCooldownUpdated(14 days);
        verification.setReinstatementCooldown(14 days);
    }

    function test_setReinstatementCooldown_affectsReinstatement() public {
        // Submit first attestation
        vm.prank(provider1);
        verification.submitAttestation(keccak256("metrics1"));

        // Suspend provider via challenge
        vm.warp(block.timestamp + 25 hours);
        vm.prank(verifier);
        verification.challengeAttestation(provider1, hex"01");

        // Set cooldown to 1 day
        vm.prank(owner);
        verification.setReinstatementCooldown(1 days);

        // Try to reinstate after 1 day — should succeed
        vm.warp(block.timestamp + 1 days);
        vm.prank(owner);
        verification.reinstateProvider(provider1);

        BunkerVerification.AttestationRecord memory rec = verification.getAttestation(provider1);
        assertFalse(rec.suspended);
    }
}
