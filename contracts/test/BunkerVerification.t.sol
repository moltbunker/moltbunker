// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";

/// @title BunkerVerificationTest
/// @notice Tests for the provider verification system built on BunkerStaking.
///         Verification combines staking attestation, slash proposals, appeals,
///         and the ability to suspend/reinstate providers.
///
/// @dev Provider verification in the current system is achieved through:
///        - isActiveProvider() — baseline verification of staked status
///        - Slash proposals — mechanism for challenging provider behavior
///        - Appeal system — dispute resolution for challenged providers
///        - Slash execution — suspension (deactivation) for confirmed violations
///        - Re-staking — reinstatement by meeting minimum requirements
///
///      A dedicated BunkerVerification contract with attestation scheduling,
///      rate limiting, and automated suspension is planned for V2.
contract BunkerVerificationTest is Test {
    BunkerToken public token;
    BunkerStaking public staking;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public provider1 = makeAddr("provider1");
    address public provider2 = makeAddr("provider2");
    address public slasher = makeAddr("slasher");
    address public verifier = makeAddr("verifier");

    uint256 constant STARTER_MIN = 1_000_000e18;
    uint256 constant BRONZE_MIN = 5_000_000e18;

    function setUp() public {
        vm.startPrank(owner);
        token = new BunkerToken(owner);
        staking = new BunkerStaking(address(token), treasury, owner);
        staking.grantRole(staking.SLASHER_ROLE(), slasher);

        // Grant slasher role to verifier (verifier acts as challenger).
        staking.grantRole(staking.SLASHER_ROLE(), verifier);
        staking.setSlashingEnabled(true);

        token.mint(provider1, 10_000_000e18);
        token.mint(provider2, 10_000_000e18);
        vm.stopPrank();

        vm.prank(provider1);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider2);
        token.approve(address(staking), type(uint256).max);
    }

    // ================================================================
    //  1. SUBMIT ATTESTATION (provider proves active staking)
    // ================================================================

    function test_submitAttestation_providerIsActive() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Provider "attests" by maintaining active staking status.
        assertTrue(staking.isActiveProvider(provider1));
    }

    function test_submitAttestation_providerInfoAvailable() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertTrue(info.active);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN));
        assertGt(info.registeredAt, 0);
    }

    function test_submitAttestation_slashableBalanceIsTotal() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        uint256 slashable = staking.getSlashableBalance(provider1);
        assertEq(slashable, BRONZE_MIN);
    }

    function test_submitAttestation_multipleProvidersIndependent() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(provider2);
        staking.stake(BRONZE_MIN);

        assertTrue(staking.isActiveProvider(provider1));
        assertTrue(staking.isActiveProvider(provider2));

        // Different tiers.
        assertEq(uint8(staking.getTier(provider1)), uint8(BunkerStaking.Tier.Starter));
        assertEq(uint8(staking.getTier(provider2)), uint8(BunkerStaking.Tier.Bronze));
    }

    // ================================================================
    //  2. ATTESTATION RATE LIMITING
    // ================================================================
    // NOTE: Rate limiting on attestation submission is a planned V2 feature.
    // These tests verify the existing proposal system as a foundation.

    function test_attestationTooEarly_proposalCreationHasNoRateLimit() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Multiple proposals can be created rapidly (no rate limit currently).
        vm.startPrank(slasher);
        staking.proposeSlash(provider1, 100e18, "check 1");
        staking.proposeSlash(provider1, 100e18, "check 2");
        staking.proposeSlash(provider1, 100e18, "check 3");
        vm.stopPrank();

        assertEq(staking.slashProposalCount(), 3);
    }

    function test_attestationTooEarly_appealWindowIsRateLimit() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(slasher);
        uint256 proposalId = staking.proposeSlash(provider1, 100e18, "check");

        // Cannot execute before 48-hour window.
        vm.prank(slasher);
        vm.expectRevert();
        staking.executeSlash(proposalId);

        // Must wait for the window.
        vm.warp(block.timestamp + 48 hours);

        vm.prank(slasher);
        staking.executeSlash(proposalId);
    }

    // ================================================================
    //  3. MISSED ATTESTATIONS (suspension after violations)
    // ================================================================

    function test_missedAttestations_slashDeactivatesProvider() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Three slashes deactivate the provider.
        vm.startPrank(slasher);
        staking.slash(provider1, 30_000e18);
        staking.slash(provider1, 30_000e18);
        staking.slash(provider1, 30_000e18);
        vm.stopPrank();

        // Below minimum -> deactivated.
        assertFalse(staking.isActiveProvider(provider1));
    }

    function test_missedAttestations_multipleSlashesReduceStake() public {
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

    function test_missedAttestations_viaProposalSystem() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Create proposal for "missed attestation".
        vm.prank(verifier);
        uint256 proposalId = staking.proposeSlash(provider1, STARTER_MIN, "missed 3 attestations");

        // Provider does not appeal. Verifier executes after window.
        vm.warp(block.timestamp + 48 hours);

        vm.prank(verifier);
        staking.executeSlash(proposalId);

        assertFalse(staking.isActiveProvider(provider1));
    }

    // ================================================================
    //  4. REINSTATE (owner can reinstate via re-staking)
    // ================================================================

    function test_reinstate_providerCanRestake() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Slash to deactivate.
        vm.prank(slasher);
        staking.slash(provider1, STARTER_MIN);
        assertFalse(staking.isActiveProvider(provider1));

        // Reinstate by re-staking.
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        assertTrue(staking.isActiveProvider(provider1));
    }

    function test_reinstate_requiresMinimumStake() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(slasher);
        staking.slash(provider1, STARTER_MIN);

        // Insufficient stake.
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

    function test_reinstate_reinstatedProviderCanEarnRewards() public {
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Slash.
        vm.prank(slasher);
        staking.slash(provider1, STARTER_MIN);

        // Reinstate.
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // Set up rewards.
        vm.startPrank(owner);
        token.mint(owner, 84_000e18);
        token.approve(address(staking), 84_000e18);
        token.transfer(address(staking), 84_000e18);
        staking.notifyRewardAmount(70_000e18);
        vm.stopPrank();

        vm.warp(block.timestamp + 7 days);

        uint256 earned = staking.earned(provider1);
        assertGt(earned, 0, "Reinstated provider should earn rewards");
    }

    // ================================================================
    //  5. CHALLENGE (verifier can challenge via proposal)
    // ================================================================

    function test_challenge_verifierCanPropose() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(verifier);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "verification challenge");

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertEq(proposal.provider, provider1);
        assertEq(proposal.amount, 500e18);
        assertFalse(proposal.executed);
    }

    function test_challenge_providerCanAppealChallenge() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(verifier);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "challenge");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertTrue(proposal.appealed);
    }

    function test_challenge_appealedChallengeResolved() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(verifier);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "challenge");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        // Owner resolves: dismiss the challenge.
        vm.prank(owner);
        staking.resolveAppeal(proposalId, false);

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertTrue(proposal.resolved);
        assertFalse(proposal.executed);

        // Provider stake unchanged.
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN));
    }

    function test_challenge_upheldChallengeSlashesProvider() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(verifier);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "challenge");

        vm.prank(provider1);
        staking.appealSlash(proposalId);

        // Owner upholds.
        vm.prank(owner);
        staking.resolveAppeal(proposalId, true);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 500e18));
    }

    function test_challenge_unappealed_executedAfterWindow() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.prank(verifier);
        uint256 proposalId = staking.proposeSlash(provider1, 500e18, "challenge");

        // Provider does not appeal.
        vm.warp(block.timestamp + 48 hours);

        vm.prank(verifier);
        staking.executeSlash(proposalId);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 500e18));
    }

    function test_challenge_nonSlasherCannotChallenge() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        address nobody = makeAddr("nobody");
        vm.prank(nobody);
        vm.expectRevert();
        staking.proposeSlash(provider1, 500e18, "unauthorized challenge");
    }

    // ================================================================
    //  6. FULL VERIFICATION LIFECYCLE
    // ================================================================

    function test_lifecycle_registerVerifyChallengeResolve() public {
        // 1. Provider registers.
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);
        assertTrue(staking.isActiveProvider(provider1));

        // 2. Verifier challenges.
        vm.prank(verifier);
        uint256 proposalId = staking.proposeSlash(provider1, 200e18, "memory check failed");

        // 3. Provider appeals.
        vm.prank(provider1);
        staking.appealSlash(proposalId);

        // 4. Owner resolves (false alarm).
        vm.prank(owner);
        staking.resolveAppeal(proposalId, false);

        // Provider unaffected.
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN));
        assertTrue(info.active);
    }

    function test_lifecycle_registerFailVerificationSuspendReinstate() public {
        // 1. Provider registers.
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        // 2. Verifier detects violation.
        vm.prank(verifier);
        staking.slash(provider1, STARTER_MIN);

        // 3. Provider suspended (deactivated).
        assertFalse(staking.isActiveProvider(provider1));

        // 4. Provider reinstates.
        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        assertTrue(staking.isActiveProvider(provider1));
    }

    // ================================================================
    //  7. MULTIPLE VERIFICATION ROUNDS
    // ================================================================

    function test_multipleRounds_sequentialChallenges() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Round 1: challenge and execute.
        vm.prank(verifier);
        uint256 id1 = staking.proposeSlash(provider1, 100e18, "round 1");
        vm.warp(block.timestamp + 48 hours);
        vm.prank(verifier);
        staking.executeSlash(id1);

        // Round 2: challenge and execute.
        vm.prank(verifier);
        uint256 id2 = staking.proposeSlash(provider1, 100e18, "round 2");
        vm.warp(block.timestamp + 48 hours);
        vm.prank(verifier);
        staking.executeSlash(id2);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 200e18));
    }

    function test_multipleRounds_mixOfImmediateAndProposal() public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        // Immediate slash.
        vm.prank(slasher);
        staking.slashImmediate(provider1, 100e18);

        // Proposal-based slash.
        vm.prank(verifier);
        uint256 proposalId = staking.proposeSlash(provider1, 100e18, "verification");
        vm.warp(block.timestamp + 48 hours);
        vm.prank(verifier);
        staking.executeSlash(proposalId);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        assertEq(info.stakedAmount, uint128(BRONZE_MIN - 200e18));
    }

    // ================================================================
    //  8. FUZZ TESTS
    // ================================================================

    function testFuzz_challengeAmount(uint128 challengeAmount) public {
        vm.prank(provider1);
        staking.stake(BRONZE_MIN);

        vm.assume(challengeAmount > 0 && uint256(challengeAmount) <= BRONZE_MIN);

        vm.prank(verifier);
        uint256 proposalId = staking.proposeSlash(provider1, challengeAmount, "fuzz");

        BunkerStaking.SlashProposal memory proposal = staking.getSlashProposal(proposalId);
        assertEq(proposal.amount, uint256(challengeAmount));
        assertEq(proposal.provider, provider1);
    }

    function testFuzz_reinstateAfterSlash(uint128 slashAmount) public {
        vm.assume(slashAmount > 0 && uint256(slashAmount) <= STARTER_MIN);

        vm.prank(provider1);
        staking.stake(STARTER_MIN);

        vm.prank(slasher);
        staking.slash(provider1, slashAmount);

        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(provider1);
        uint256 remaining = info.stakedAmount;

        if (remaining < STARTER_MIN) {
            // Provider deactivated. Re-staking needed.
            assertFalse(info.active);
            uint256 needed = STARTER_MIN - remaining;

            vm.prank(provider1);
            staking.stake(needed);

            assertTrue(staking.isActiveProvider(provider1));
        } else {
            // Provider still active.
            assertTrue(info.active);
        }
    }
}
