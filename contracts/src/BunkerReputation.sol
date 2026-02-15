// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable2Step.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

/// @title BunkerReputation
/// @author Moltbunker
/// @notice On-chain reputation scoring for compute providers. Reporters (trusted
///         off-chain nodes) record events that adjust a provider's score. The score
///         decays over time toward a floor to incentivize continued participation.
/// @dev Deployed on Base L2. Uses OpenZeppelin v5. Score range: 0-1000.
///      Initial score: 500 (Standard tier). Decay: 1 point/week toward floor of 100.
///      Tier thresholds: Elite >= 900, Trusted >= 750, Standard >= 500,
///      Probation >= 250, Restricted < 250.
contract BunkerReputation is Ownable2Step, AccessControl {

    // ──────────────────────────────────────────────
    //  Constants
    // ──────────────────────────────────────────────

    /// @notice Contract version.
    string public constant VERSION = "1.0.0";

    /// @notice Role identifier for authorized reputation reporters.
    bytes32 public constant REPORTER_ROLE = keccak256("REPORTER_ROLE");

    /// @notice Maximum reputation score.
    uint256 public constant MAX_SCORE = 1000;

    /// @notice Initial reputation score for newly registered providers.
    uint256 public constant INITIAL_SCORE = 500;

    /// @notice One week in seconds (for decay calculation).
    uint256 public constant WEEK = 7 days;

    // ──────────────────────────────────────────────
    //  Configurable Parameters (converted from constants)
    // ──────────────────────────────────────────────

    /// @notice Minimum score required to be eligible for job assignments.
    uint256 public minScoreForJobs = 250;

    /// @notice Decay rate: points per week.
    uint256 public decayRate = 1;

    /// @notice Decay floor: score will not decay below this value.
    uint256 public decayFloor = 100;

    /// @notice Maximum positive delta for custom events.
    int16 public maxCustomDelta = 200;

    /// @notice Minimum (most negative) delta for custom events.
    int16 public minCustomDelta = -200;

    /// @notice Tier threshold: Elite (default 900).
    uint16 public tierElite = 900;

    /// @notice Tier threshold: Trusted (default 750).
    uint16 public tierTrusted = 750;

    /// @notice Tier threshold: Standard (default 500).
    uint16 public tierStandard = 500;

    /// @notice Tier threshold: Probation (default 250).
    uint16 public tierProbation = 250;

    // ──────────────────────────────────────────────
    //  Types
    // ──────────────────────────────────────────────

    /// @notice Full reputation data for a provider.
    struct ReputationData {
        uint32 score;
        uint32 jobsCompleted;
        uint32 jobsFailed;
        uint32 slashCount;
        uint48 lastUpdated;
        uint48 registeredAt;
    }

    /// @notice Reputation tier classification.
    enum ReputationTier { Restricted, Probation, Standard, Trusted, Elite }

    // ──────────────────────────────────────────────
    //  State
    // ──────────────────────────────────────────────

    /// @notice Reputation data per provider address.
    mapping(address => ReputationData) public reputations;

    /// @notice Score adjustment for a completed job.
    int16 public jobCompletedDelta = 5;

    /// @notice Score adjustment for early job completion (bonus).
    int16 public jobEarlyDelta = 10;

    /// @notice Score adjustment for perfect uptime over a period.
    int16 public perfectUptimeDelta = 20;

    /// @notice Score adjustment for a job timeout.
    int16 public jobTimeoutDelta = -10;

    /// @notice Score adjustment for a health check failure.
    int16 public healthFailDelta = -25;

    /// @notice Score adjustment for a replica mismatch.
    int16 public replicaMismatchDelta = -50;

    /// @notice Score adjustment for a slash event.
    int16 public slashEventDelta = -100;

    /// @notice Score adjustment for a security violation.
    int16 public securityViolationDelta = -200;

    // ──────────────────────────────────────────────
    //  Errors
    // ──────────────────────────────────────────────

    error ProviderNotRegistered(address provider);
    error ProviderAlreadyRegistered(address provider);
    error ScoreOverflow();
    error DeltaOutOfBounds(int16 delta);
    error PositiveDeltaRequired();
    error NegativeDeltaRequired();
    error ThresholdsNotAscending();
    error ThresholdExceedsMax();
    error FloorTooHigh();
    error ScoreExceedsMax();
    error InvalidDelta();

    // ──────────────────────────────────────────────
    //  Events
    // ──────────────────────────────────────────────

    /// @notice Emitted when a provider's score is updated.
    event ScoreUpdated(
        address indexed provider,
        uint32 oldScore,
        uint32 newScore,
        string reason
    );

    /// @notice Emitted when a provider is registered in the reputation system.
    event ProviderRegistered(address indexed provider);

    /// @notice Emitted when score delta parameters are updated by the owner.
    event DeltaParametersUpdated();

    /// @notice Emitted when tier thresholds are updated.
    event TierThresholdsUpdated(uint16 probation, uint16 standard, uint16 trusted, uint16 elite);

    /// @notice Emitted when decay parameters are updated.
    event DecayParamsUpdated(uint256 rate, uint256 floor);

    /// @notice Emitted when the minimum score for jobs is updated.
    event MinScoreForJobsUpdated(uint256 minScore);

    /// @notice Emitted when the max custom delta range is updated.
    event MaxCustomDeltaUpdated(int16 maxDelta);

    // ──────────────────────────────────────────────
    //  Constructor
    // ──────────────────────────────────────────────

    /// @param _initialOwner Admin wallet address.
    constructor(address _initialOwner) Ownable(_initialOwner) {
        _grantRole(DEFAULT_ADMIN_ROLE, _initialOwner);
    }

    // ──────────────────────────────────────────────
    //  External: Reporter Functions
    // ──────────────────────────────────────────────

    /// @notice Register a new provider in the reputation system.
    /// @dev Sets the initial score to INITIAL_SCORE (500).
    /// @param provider The provider address to register.
    function registerProvider(address provider) external onlyRole(REPORTER_ROLE) {
        if (reputations[provider].registeredAt != 0)
            revert ProviderAlreadyRegistered(provider);

        reputations[provider] = ReputationData({
            score: uint32(INITIAL_SCORE),
            jobsCompleted: 0,
            jobsFailed: 0,
            slashCount: 0,
            lastUpdated: uint48(block.timestamp),
            registeredAt: uint48(block.timestamp)
        });

        emit ProviderRegistered(provider);
    }

    /// @notice Record a completed job for a provider.
    /// @param provider The provider address.
    function recordJobCompleted(address provider) external onlyRole(REPORTER_ROLE) {
        _ensureRegistered(provider);
        _applyDecayInternal(provider);

        ReputationData storage rep = reputations[provider];
        uint32 oldScore = rep.score;
        rep.jobsCompleted += 1;
        rep.score = _applyDelta(rep.score, jobCompletedDelta);
        rep.lastUpdated = uint48(block.timestamp);

        emit ScoreUpdated(provider, oldScore, rep.score, "job_completed");
    }

    /// @notice Record a failed job for a provider.
    /// @param provider The provider address.
    function recordJobFailed(address provider) external onlyRole(REPORTER_ROLE) {
        _ensureRegistered(provider);
        _applyDecayInternal(provider);

        ReputationData storage rep = reputations[provider];
        uint32 oldScore = rep.score;
        rep.jobsFailed += 1;
        rep.score = _applyDelta(rep.score, jobTimeoutDelta);
        rep.lastUpdated = uint48(block.timestamp);

        emit ScoreUpdated(provider, oldScore, rep.score, "job_failed");
    }

    /// @notice Record a slash event for a provider.
    /// @param provider The provider address.
    function recordSlashEvent(address provider) external onlyRole(REPORTER_ROLE) {
        _ensureRegistered(provider);
        _applyDecayInternal(provider);

        ReputationData storage rep = reputations[provider];
        uint32 oldScore = rep.score;
        rep.slashCount += 1;
        rep.score = _applyDelta(rep.score, slashEventDelta);
        rep.lastUpdated = uint48(block.timestamp);

        emit ScoreUpdated(provider, oldScore, rep.score, "slash_event");
    }

    /// @notice Record a custom event with an arbitrary score delta and reason.
    /// @param provider The provider address.
    /// @param delta The score change (positive or negative).
    /// @param reason Human-readable reason string.
    function recordEvent(
        address provider,
        int16 delta,
        string calldata reason
    ) external onlyRole(REPORTER_ROLE) {
        if (delta > maxCustomDelta || delta < minCustomDelta) revert DeltaOutOfBounds(delta);
        _ensureRegistered(provider);
        _applyDecayInternal(provider);

        ReputationData storage rep = reputations[provider];
        uint32 oldScore = rep.score;
        rep.score = _applyDelta(rep.score, delta);
        rep.lastUpdated = uint48(block.timestamp);

        emit ScoreUpdated(provider, oldScore, rep.score, reason);
    }

    // ──────────────────────────────────────────────
    //  External: Decay (Permissionless)
    // ──────────────────────────────────────────────

    /// @notice Apply time-based decay to a provider's score.
    /// @dev Anyone can call this to keep scores up-to-date. Decay is computed
    ///      as DECAY_RATE points per week since lastUpdated, down to DECAY_FLOOR.
    /// @param provider The provider address.
    function applyDecay(address provider) external {
        _ensureRegistered(provider);

        ReputationData storage rep = reputations[provider];
        uint32 oldScore = rep.score;

        _applyDecayInternal(provider);

        if (rep.score != oldScore) {
            emit ScoreUpdated(provider, oldScore, rep.score, "decay");
        }
    }

    // ──────────────────────────────────────────────
    //  External: Views
    // ──────────────────────────────────────────────

    /// @notice Get the current score for a provider (with pending decay applied).
    /// @param provider The provider address.
    /// @return score The provider's effective score including pending decay.
    function getScore(address provider) external view returns (uint32) {
        _ensureRegistered(provider);
        return _getDecayedScore(provider);
    }

    /// @notice Get the reputation tier for a provider (with pending decay applied).
    /// @param provider The provider address.
    /// @return tier The provider's current tier.
    function getTier(address provider) external view returns (ReputationTier) {
        _ensureRegistered(provider);
        return _scoreTier(_getDecayedScore(provider));
    }

    /// @notice Check if a provider is eligible for job assignments (with pending decay applied).
    /// @param provider The provider address.
    /// @return eligible True if the provider's score meets the minimum.
    function isEligibleForJobs(address provider) external view returns (bool eligible) {
        if (reputations[provider].registeredAt == 0) return false;
        return _getDecayedScore(provider) >= uint32(minScoreForJobs);
    }

    /// @notice Get the full reputation data for a provider.
    /// @param provider The provider address.
    /// @return data The provider's reputation data.
    function getReputation(address provider)
        external
        view
        returns (ReputationData memory data)
    {
        return reputations[provider];
    }

    // ──────────────────────────────────────────────
    //  External: Admin
    // ──────────────────────────────────────────────

    /// @notice Update the score delta parameters.
    /// @dev Only owner can adjust the scoring deltas for different event types.
    function setDeltaParameters(
        int16 _jobCompletedDelta,
        int16 _jobEarlyDelta,
        int16 _perfectUptimeDelta,
        int16 _jobTimeoutDelta,
        int16 _healthFailDelta,
        int16 _replicaMismatchDelta,
        int16 _slashEventDelta,
        int16 _securityViolationDelta
    ) external onlyOwner {
        // M-14: Validate positive deltas are non-negative
        if (_jobCompletedDelta < 0 || _jobEarlyDelta < 0 || _perfectUptimeDelta < 0)
            revert PositiveDeltaRequired();
        // M-14: Validate negative deltas are non-positive
        if (_jobTimeoutDelta > 0 || _healthFailDelta > 0 || _replicaMismatchDelta > 0 ||
            _slashEventDelta > 0 || _securityViolationDelta > 0)
            revert NegativeDeltaRequired();

        jobCompletedDelta = _jobCompletedDelta;
        jobEarlyDelta = _jobEarlyDelta;
        perfectUptimeDelta = _perfectUptimeDelta;
        jobTimeoutDelta = _jobTimeoutDelta;
        healthFailDelta = _healthFailDelta;
        replicaMismatchDelta = _replicaMismatchDelta;
        slashEventDelta = _slashEventDelta;
        securityViolationDelta = _securityViolationDelta;

        emit DeltaParametersUpdated();
    }

    /// @notice Adjust reputation tier thresholds. Must be strictly ascending.
    /// @param probation Probation threshold.
    /// @param standard Standard threshold.
    /// @param trusted Trusted threshold.
    /// @param elite Elite threshold.
    function setTierThresholds(
        uint16 probation, uint16 standard, uint16 trusted, uint16 elite
    ) external onlyOwner {
        if (probation >= standard || standard >= trusted || trusted >= elite)
            revert ThresholdsNotAscending();
        if (elite > MAX_SCORE) revert ThresholdExceedsMax();
        tierElite = elite;
        tierTrusted = trusted;
        tierStandard = standard;
        tierProbation = probation;
        emit TierThresholdsUpdated(probation, standard, trusted, elite);
    }

    /// @notice Adjust decay rate and floor.
    /// @param rate New decay rate (points per week).
    /// @param floor New decay floor.
    function setDecayParams(uint256 rate, uint256 floor) external onlyOwner {
        if (floor > MAX_SCORE / 2) revert FloorTooHigh();
        decayRate = rate;
        decayFloor = floor;
        emit DecayParamsUpdated(rate, floor);
    }

    /// @notice Adjust minimum score for job eligibility.
    /// @param _minScore New minimum score.
    function setMinScoreForJobs(uint256 _minScore) external onlyOwner {
        if (_minScore > MAX_SCORE) revert ScoreExceedsMax();
        minScoreForJobs = _minScore;
        emit MinScoreForJobsUpdated(_minScore);
    }

    /// @notice Adjust the max custom delta range (symmetric: +maxDelta / -maxDelta).
    /// @param maxDelta New maximum positive delta.
    function setMaxCustomDelta(int16 maxDelta) external onlyOwner {
        if (maxDelta <= 0 || maxDelta > int16(int256(MAX_SCORE) / 2)) revert InvalidDelta();
        maxCustomDelta = maxDelta;
        minCustomDelta = -maxDelta;
        emit MaxCustomDeltaUpdated(maxDelta);
    }

    // ──────────────────────────────────────────────
    //  Internal
    // ──────────────────────────────────────────────

    /// @dev Revert if the provider is not registered.
    function _ensureRegistered(address provider) internal view {
        if (reputations[provider].registeredAt == 0)
            revert ProviderNotRegistered(provider);
    }

    /// @dev Apply a signed delta to a score, clamping to [0, MAX_SCORE].
    function _applyDelta(uint32 currentScore, int16 delta) internal pure returns (uint32) {
        int256 newScore = int256(uint256(currentScore)) + int256(delta);
        if (newScore < 0) return 0;
        if (newScore > int256(MAX_SCORE)) return uint32(MAX_SCORE);
        return uint32(uint256(newScore));
    }

    /// @dev Compute the decayed score without writing to storage (for view functions).
    function _getDecayedScore(address provider) internal view returns (uint32) {
        ReputationData storage rep = reputations[provider];
        if (rep.registeredAt == 0) return 0;
        if (rep.score <= uint32(decayFloor)) return rep.score;

        uint256 elapsed = block.timestamp - uint256(rep.lastUpdated);
        uint256 weeksPassed = elapsed / WEEK;
        if (weeksPassed == 0) return rep.score;

        uint256 decayAmount = weeksPassed * decayRate;
        uint256 currentScore = uint256(rep.score);
        if (currentScore <= decayFloor + decayAmount) return uint32(decayFloor);
        return uint32(currentScore - decayAmount);
    }

    /// @dev Apply time-based decay to a provider's score.
    ///      M-13: Advances lastUpdated by whole weeks (not reset to now) to prevent
    ///      partial-week remainder gaming.
    function _applyDecayInternal(address provider) internal {
        ReputationData storage rep = reputations[provider];

        uint256 elapsed = block.timestamp - uint256(rep.lastUpdated);
        uint256 weeksPassed = elapsed / WEEK;

        if (weeksPassed == 0) return;

        // M-13: Advance lastUpdated by whole weeks (not reset to block.timestamp)
        // to prevent partial-week remainder gaming.
        rep.lastUpdated = uint48(uint256(rep.lastUpdated) + (weeksPassed * WEEK));

        if (rep.score <= uint32(decayFloor)) {
            return;
        }

        uint256 decay = weeksPassed * decayRate;
        uint256 currentScore = uint256(rep.score);

        if (currentScore - decay < decayFloor) {
            rep.score = uint32(decayFloor);
        } else {
            rep.score = uint32(currentScore - decay);
        }
    }

    /// @dev Map a score to a reputation tier.
    function _scoreTier(uint32 score) internal view returns (ReputationTier) {
        if (score >= tierElite) return ReputationTier.Elite;
        if (score >= tierTrusted) return ReputationTier.Trusted;
        if (score >= tierStandard) return ReputationTier.Standard;
        if (score >= tierProbation) return ReputationTier.Probation;
        return ReputationTier.Restricted;
    }
}
