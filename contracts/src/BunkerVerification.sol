// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable2Step.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

/// @title BunkerVerification
/// @author Moltbunker
/// @notice On-chain verification of compute capacity through periodic attestations.
///         Providers submit hashes of their container metrics at regular intervals.
///         Missed attestations trigger warnings and eventual suspension. Verifiers
///         can challenge fraudulent attestations with fraud proofs.
/// @dev Deployed on Base L2. Uses OpenZeppelin v5. Default attestation interval is
///      24 hours. Providers are suspended after 3 consecutive missed attestations.
///      Reinstatement requires owner approval.
contract BunkerVerification is Ownable2Step, AccessControl {

    // ──────────────────────────────────────────────
    //  Constants
    // ──────────────────────────────────────────────

    /// @notice Contract version.
    string public constant VERSION = "1.0.0";

    /// @notice Role identifier for authorized verifiers.
    bytes32 public constant VERIFIER_ROLE = keccak256("VERIFIER_ROLE");

    // ──────────────────────────────────────────────
    //  Types
    // ──────────────────────────────────────────────

    /// @notice Attestation state for a provider.
    struct AttestationRecord {
        bytes32 lastAttestationHash; // Hash of container metrics
        uint48 lastAttestationTime;
        uint32 totalAttestations;
        uint32 missedAttestations;
        uint32 consecutiveMissed;
        bool suspended;
    }

    // ──────────────────────────────────────────────
    //  State
    // ──────────────────────────────────────────────

    /// @notice Required interval between attestations.
    uint256 public attestationInterval = 24 hours;

    /// @notice Maximum consecutive missed attestations before suspension.
    uint256 public maxMissedAttestations = 3;

    /// @notice Reinstatement cooldown after suspension (default 7 days).
    uint256 public reinstatementCooldown = 7 days;

    /// @notice Attestation records per provider address.
    mapping(address => AttestationRecord) public attestations;

    /// @notice Timestamp when a provider was suspended (for reinstatement cooldown).
    mapping(address => uint256) public suspendedAt;

    // ──────────────────────────────────────────────
    //  Errors
    // ──────────────────────────────────────────────

    error AttestationTooEarly(uint256 nextAllowed, uint256 current);
    error ProviderSuspendedError(address provider);
    error InvalidAttestation();
    error ProviderNotAttesting(address provider);
    error InvalidInterval(uint256 interval);
    error InvalidMaxMissed(uint256 maxMissed);
    error ReinstatementCooldownActive();
    error InvalidCooldown();

    // ──────────────────────────────────────────────
    //  Events
    // ──────────────────────────────────────────────

    /// @notice Emitted when a provider submits an attestation.
    event AttestationSubmitted(
        address indexed provider,
        bytes32 attestationHash,
        uint32 totalAttestations
    );

    /// @notice Emitted when a provider misses an attestation check.
    event AttestationMissed(address indexed provider, uint32 consecutiveMissed);

    /// @notice Emitted when a provider is suspended for missed attestations.
    event ProviderSuspended(address indexed provider, uint32 missedCount);

    /// @notice Emitted when a suspended provider is reinstated by the owner.
    event ProviderReinstated(address indexed provider);

    /// @notice Emitted when a verifier challenges an attestation.
    event AttestationChallenged(
        address indexed challenger,
        address indexed provider,
        bytes32 attestationHash
    );

    /// @notice Emitted when the attestation interval is updated.
    event AttestationIntervalUpdated(uint256 oldInterval, uint256 newInterval);

    /// @notice Emitted when the max missed attestations is updated.
    event MaxMissedAttestationsUpdated(uint256 oldMax, uint256 newMax);

    /// @notice Emitted when the reinstatement cooldown is updated.
    event ReinstatementCooldownUpdated(uint256 newCooldown);

    // ──────────────────────────────────────────────
    //  Constructor
    // ──────────────────────────────────────────────

    /// @param _initialOwner Admin wallet address.
    constructor(address _initialOwner) Ownable(_initialOwner) {
        _grantRole(DEFAULT_ADMIN_ROLE, _initialOwner);
    }

    // ──────────────────────────────────────────────
    //  External: Provider Functions
    // ──────────────────────────────────────────────

    /// @notice Submit a periodic attestation of compute capacity.
    /// @dev The attestation hash should be a keccak256 digest of the provider's
    ///      container metrics (CPU usage, memory, running containers, etc.).
    ///      First attestation initializes the record. Subsequent attestations must
    ///      wait at least `attestationInterval` since the last one.
    /// @param attestationHash Hash of the provider's container metrics.
    function submitAttestation(bytes32 attestationHash) external {
        if (attestationHash == bytes32(0)) revert InvalidAttestation();

        AttestationRecord storage record = attestations[msg.sender];
        if (record.suspended) revert ProviderSuspendedError(msg.sender);

        // First attestation: initialize the record.
        if (record.lastAttestationTime == 0) {
            record.lastAttestationHash = attestationHash;
            record.lastAttestationTime = uint48(block.timestamp);
            record.totalAttestations = 1;
            record.missedAttestations = 0;
            record.consecutiveMissed = 0;
            record.suspended = false;

            emit AttestationSubmitted(msg.sender, attestationHash, 1);
            return;
        }

        // Subsequent attestations: enforce interval.
        uint256 nextAllowed = uint256(record.lastAttestationTime) + attestationInterval;
        if (block.timestamp < nextAllowed)
            revert AttestationTooEarly(nextAllowed, block.timestamp);

        record.lastAttestationHash = attestationHash;
        record.lastAttestationTime = uint48(block.timestamp);
        record.totalAttestations += 1;
        record.consecutiveMissed = 0; // Reset consecutive missed on successful attestation.

        emit AttestationSubmitted(
            msg.sender,
            attestationHash,
            record.totalAttestations
        );
    }

    // ──────────────────────────────────────────────
    //  External: Verifier Functions
    // ──────────────────────────────────────────────

    /// @notice Check if a provider has missed their attestation window and record it.
    /// @dev Only callable by addresses with VERIFIER_ROLE. If the provider has not
    ///      attested within the required interval, their missed count is incremented.
    ///      Suspension is triggered after `maxMissedAttestations` consecutive misses.
    /// @param provider The provider address to check.
    function checkMissedAttestations(address provider) external onlyRole(VERIFIER_ROLE) {
        AttestationRecord storage record = attestations[provider];
        if (record.lastAttestationTime == 0) revert ProviderNotAttesting(provider);
        if (record.suspended) revert ProviderSuspendedError(provider);

        uint256 deadline = uint256(record.lastAttestationTime) + attestationInterval;
        if (block.timestamp <= deadline) {
            // Provider is still within their attestation window; nothing to do.
            return;
        }

        // Calculate how many intervals have been missed.
        uint256 elapsed = block.timestamp - uint256(record.lastAttestationTime);
        uint256 intervalsMissed = elapsed / attestationInterval;
        if (intervalsMissed == 0) return;

        record.missedAttestations += uint32(intervalsMissed);
        record.consecutiveMissed += uint32(intervalsMissed);

        // C-10: Advance lastAttestationTime by missed intervals to prevent double-counting
        record.lastAttestationTime = uint48(
            uint256(record.lastAttestationTime) + (intervalsMissed * attestationInterval)
        );

        emit AttestationMissed(provider, record.consecutiveMissed);

        // Suspend if consecutive missed exceeds threshold.
        if (record.consecutiveMissed >= uint32(maxMissedAttestations)) {
            record.suspended = true;
            suspendedAt[provider] = block.timestamp;
            emit ProviderSuspended(provider, record.consecutiveMissed);
        }
    }

    /// @notice Challenge a provider's attestation with a fraud proof.
    /// @dev Only callable by addresses with VERIFIER_ROLE. The fraud proof is
    ///      logged on-chain as an event. Off-chain systems handle the actual
    ///      verification and may trigger slashing via the staking contract.
    /// @param provider The provider whose attestation is being challenged.
    /// @param fraudProof Encoded fraud proof data.
    function challengeAttestation(
        address provider,
        bytes calldata fraudProof
    ) external onlyRole(VERIFIER_ROLE) {
        AttestationRecord storage record = attestations[provider];
        if (record.lastAttestationTime == 0) revert ProviderNotAttesting(provider);
        if (fraudProof.length == 0) revert InvalidAttestation();

        // H-20: Suspend provider on successful challenge
        record.suspended = true;
        suspendedAt[provider] = block.timestamp;

        emit AttestationChallenged(
            msg.sender,
            provider,
            record.lastAttestationHash
        );
        emit ProviderSuspended(provider, record.consecutiveMissed);
    }

    // ──────────────────────────────────────────────
    //  External: Owner Functions
    // ──────────────────────────────────────────────

    /// @notice Reinstate a suspended provider.
    /// @dev Resets the consecutive missed counter and lifts the suspension.
    /// @param provider The provider address to reinstate.
    function reinstateProvider(address provider) external onlyOwner {
        AttestationRecord storage record = attestations[provider];
        if (record.lastAttestationTime == 0) revert ProviderNotAttesting(provider);
        // M-12: Enforce reinstatement cooldown
        if (block.timestamp < suspendedAt[provider] + reinstatementCooldown)
            revert ReinstatementCooldownActive();

        record.suspended = false;
        record.consecutiveMissed = 0;
        record.lastAttestationTime = uint48(block.timestamp);

        emit ProviderReinstated(provider);
    }

    /// @notice Update the required attestation interval.
    /// @param interval New interval in seconds.
    function setAttestationInterval(uint256 interval) external onlyOwner {
        if (interval == 0) revert InvalidInterval(interval);
        uint256 old = attestationInterval;
        attestationInterval = interval;
        emit AttestationIntervalUpdated(old, interval);
    }

    /// @notice Update the maximum allowed consecutive missed attestations.
    /// @param maxMissed New maximum before suspension.
    function setMaxMissedAttestations(uint256 maxMissed) external onlyOwner {
        if (maxMissed == 0) revert InvalidMaxMissed(maxMissed);
        uint256 old = maxMissedAttestations;
        maxMissedAttestations = maxMissed;
        emit MaxMissedAttestationsUpdated(old, maxMissed);
    }

    /// @notice Adjust the reinstatement cooldown after suspension.
    /// @param newCooldown New cooldown in seconds (1 day min, 30 days max).
    function setReinstatementCooldown(uint256 newCooldown) external onlyOwner {
        if (newCooldown < 1 minutes || newCooldown > 30 days) revert InvalidCooldown();
        reinstatementCooldown = newCooldown;
        emit ReinstatementCooldownUpdated(newCooldown);
    }

    // ──────────────────────────────────────────────
    //  External: Views
    // ──────────────────────────────────────────────

    /// @notice Get the full attestation record for a provider.
    /// @param provider The provider address.
    /// @return record The attestation record.
    function getAttestation(address provider)
        external
        view
        returns (AttestationRecord memory record)
    {
        return attestations[provider];
    }

    /// @notice Check if a provider's attestation is current (within the interval).
    /// @param provider The provider address.
    /// @return current True if the provider has attested within the required interval.
    function isAttestationCurrent(address provider) external view returns (bool current) {
        AttestationRecord storage record = attestations[provider];
        if (record.lastAttestationTime == 0) return false;
        if (record.suspended) return false;
        return (block.timestamp - uint256(record.lastAttestationTime)) <= attestationInterval;
    }
}
