// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable2Step.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";

/// @notice Minimal read interface to BunkerStaking. Only `stakedAmount` is
///         decoded from `getProviderInfo`; the remaining tuple fields are
///         ignored so this contract never couples to the full ProviderInfo
///         struct ABI. `slash` routes edge-provider penalties to the existing
///         stake (no duplicate token custody).
interface IBunkerStaking {
    /// @dev Returns the full provider tuple; callers destructure only the
    ///      leading `stakedAmount`. The trailing fields are intentionally
    ///      decoded as their concrete types so the ABI selector matches.
    function getProviderInfo(address provider)
        external
        view
        returns (
            uint128 stakedAmount,
            uint128 totalUnbonding,
            address beneficiary,
            uint48 registeredAt,
            bool active,
            bytes32 nodeId,
            bytes32 region,
            uint64 capabilities,
            bool frozen
        );

    /// @dev Slash a provider's stake. Requires SLASHER_ROLE on BunkerStaking,
    ///      which must be granted to this contract during deployment setup.
    function slash(address provider, uint256 amount) external;
}

/// @title BunkerEdgeRegistry
/// @author Moltbunker
/// @notice On-chain source of truth for edge-provider registration under
///         Approach A (tiered edge nodes terminate TLS + run the L7 WAF).
///         Edge providers must already hold a minimum BUNKER stake in the
///         existing BunkerStaking contract, then register their static edge
///         metadata here. They are slashing-eligible for SLA violations; the
///         penalty is routed back to BunkerStaking.slash so it hits the same
///         stake — no duplicate token custody.
/// @dev Additive contract. Zero changes to BunkerStaking, BunkerRegistry, or any
///      other deployed contract. Reuses the BunkerStaking conventions verbatim:
///      OpenZeppelin v5 base set, SLASHER_ROLE, snapshotted appeal window
///      (H-21/H-22 pattern), and the `slashingEnabled` monitor-mode gate.
contract BunkerEdgeRegistry is Ownable2Step, AccessControl, ReentrancyGuard, Pausable {
    // ──────────────────────────────────────────────
    //  Constants
    // ──────────────────────────────────────────────

    /// @notice Contract version.
    string public constant VERSION = "1.0.0";

    /// @notice Role identifier for authorized slashers.
    bytes32 public constant SLASHER_ROLE = keccak256("SLASHER_ROLE");

    /// @notice Lower bound for minStakeForEdge (BunkerStaking Starter minimum, 1 M BUNKER).
    uint256 public constant MIN_EDGE_STAKE_FLOOR = 1_000_000e18;

    /// @notice Upper bound for minStakeForEdge (BunkerStaking Platinum minimum, 1 B BUNKER).
    uint256 public constant MIN_EDGE_STAKE_CEILING = 1_000_000_000e18;

    /// @notice Lower bound for the appeal window (12 hours).
    uint256 public constant APPEAL_WINDOW_FLOOR = 12 hours;

    /// @notice Upper bound for the appeal window (14 days).
    uint256 public constant APPEAL_WINDOW_CEILING = 14 days;

    // ──────────────────────────────────────────────
    //  Configurable Parameters
    // ──────────────────────────────────────────────

    /// @notice Minimum BUNKER stake (in BunkerStaking) required to register as an
    ///         edge provider. Default 50 M BUNKER — between Silver (10 M) and Gold
    ///         (100 M) to create a meaningful edge-only capital hurdle.
    uint256 public minStakeForEdge = 50_000_000e18;

    /// @notice Appeal window for edge-slash proposals (default 48 hours).
    uint256 public appealWindow = 48 hours;

    /// @notice Whether slashing execution is enabled. When false, proposals can
    ///         still be created (monitor mode) but execution reverts.
    bool public slashingEnabled;

    // ──────────────────────────────────────────────
    //  Types
    // ──────────────────────────────────────────────

    /// @notice Static, on-chain edge-provider metadata. The fixed-size leading
    ///         fields pack into two slots (32 + 32 + 6 + 1 + 1 bytes).
    struct EdgeProviderInfo {
        bytes32 nodeId; // SHA256 of the provider's Ed25519 public key
        bytes32 region; // Geographic region identifier
        uint48 registeredAt; // Block timestamp at registration
        bool active; // Whether the provider is currently registered
        bool frozen; // Whether the provider is frozen (emergency response)
        string endpointURL; // Public TLS-terminating edge endpoint
        bytes tlsPubkeyHash; // Hash of the edge node's TLS public key
    }

    /// @notice An edge-slash proposal subject to an appeal window.
    struct EdgeSlashProposal {
        address provider;
        uint128 amount;
        uint48 proposedAt;
        bool executed;
        bool appealed;
        bool resolved;
        uint256 appealWindowSnapshot; // Snapshot of appealWindow at proposal time
        string reason;
    }

    // ──────────────────────────────────────────────
    //  State
    // ──────────────────────────────────────────────

    /// @notice Read-only reference to the BunkerStaking contract (stake source of truth).
    IBunkerStaking public immutable stakingContract;

    /// @notice Edge-provider state by address.
    mapping(address => EdgeProviderInfo) public edgeProviders;

    /// @notice Reverse lookup from edge node ID to edge-provider address.
    mapping(bytes32 => address) public nodeIdToEdgeProvider;

    /// @notice Total number of edge-slash proposals created.
    uint256 public slashProposalCount;

    /// @notice Edge-slash proposals by ID.
    mapping(uint256 => EdgeSlashProposal) public slashProposals;

    // ──────────────────────────────────────────────
    //  Errors
    // ──────────────────────────────────────────────

    error ZeroAddress();
    error ZeroAmount();
    error InsufficientStake(uint256 actual, uint256 required);
    error InsufficientSlashableBalance(uint256 amount, uint256 slashable);
    error AlreadyRegistered(address provider);
    error NotActive(address provider);
    error NodeIdAlreadyClaimed(bytes32 nodeId);
    error ProviderIsFrozen(address provider);
    error SlashingNotEnabled();
    error InvalidProposalId(uint256 proposalId);
    error ProposalAlreadyExecuted(uint256 proposalId);
    error AppealWindowNotElapsed(uint256 proposalId);
    error AppealWindowElapsed(uint256 proposalId);
    error ProposalAppealed(uint256 proposalId);
    error ProposalNotAppealed(uint256 proposalId);
    error ProposalAlreadyResolved(uint256 proposalId);
    error NotProposalProvider(uint256 proposalId, address caller);
    error InvalidMinStake(uint256 minStake);
    error InvalidAppealWindow();

    // ──────────────────────────────────────────────
    //  Events
    // ──────────────────────────────────────────────

    /// @notice Emitted when an edge provider registers.
    event EdgeProviderRegistered(address indexed provider, bytes32 nodeId, bytes32 region);

    /// @notice Emitted when an edge provider deregisters.
    event EdgeProviderDeregistered(address indexed provider);

    /// @notice Emitted when an edge provider is frozen by the slasher.
    event EdgeProviderFrozen(address indexed provider, address indexed by);

    /// @notice Emitted when an edge provider is unfrozen by the owner.
    event EdgeProviderUnfrozen(address indexed provider);

    /// @notice Emitted when an edge provider updates its endpoint metadata.
    event EdgeEndpointUpdated(address indexed provider, string endpointURL);

    /// @notice Emitted when an edge-slash proposal is created.
    event EdgeSlashProposed(
        uint256 indexed proposalId, address indexed provider, uint256 amount, string reason
    );

    /// @notice Emitted when an edge-slash proposal is executed.
    event EdgeSlashExecuted(uint256 indexed proposalId, address indexed provider, uint256 amount);

    /// @notice Emitted when an edge provider appeals a slash proposal.
    event EdgeSlashAppealed(uint256 indexed proposalId, address indexed provider);

    /// @notice Emitted when an edge-slash appeal is resolved.
    event EdgeAppealResolved(uint256 indexed proposalId, bool upheld);

    /// @notice Emitted when the minimum edge stake threshold is updated.
    event MinStakeUpdated(uint256 oldMinStake, uint256 newMinStake);

    /// @notice Emitted when the appeal window is updated.
    event AppealWindowUpdated(uint256 newWindow);

    /// @notice Emitted when slashing is enabled or disabled.
    event SlashingEnabledUpdated(bool enabled);

    // ──────────────────────────────────────────────
    //  Constructor
    // ──────────────────────────────────────────────

    /// @param _stakingContract Address of the deployed BunkerStaking contract.
    /// @param _initialOwner Admin wallet address.
    constructor(address _stakingContract, address _initialOwner) Ownable(_initialOwner) {
        if (_stakingContract == address(0) || _initialOwner == address(0)) revert ZeroAddress();

        stakingContract = IBunkerStaking(_stakingContract);

        // Grant DEFAULT_ADMIN_ROLE to owner for managing SLASHER_ROLE.
        _grantRole(DEFAULT_ADMIN_ROLE, _initialOwner);
    }

    // ──────────────────────────────────────────────
    //  External: Edge-Provider Registration
    // ──────────────────────────────────────────────

    /// @notice Register as an edge provider. Requires an active BunkerStaking
    ///         stake at least equal to `minStakeForEdge`.
    /// @param nodeId SHA256 of the edge node's Ed25519 public key (must be unique).
    /// @param region Geographic region identifier.
    /// @param endpointURL Public TLS-terminating edge endpoint.
    /// @param tlsPubkeyHash Hash of the edge node's TLS public key.
    function registerEdgeProvider(
        bytes32 nodeId,
        bytes32 region,
        string calldata endpointURL,
        bytes calldata tlsPubkeyHash
    ) external nonReentrant whenNotPaused {
        EdgeProviderInfo storage info = edgeProviders[msg.sender];
        if (info.active) revert AlreadyRegistered(msg.sender);
        if (info.frozen) revert ProviderIsFrozen(msg.sender);
        if (nodeId != bytes32(0) && nodeIdToEdgeProvider[nodeId] != address(0)) {
            revert NodeIdAlreadyClaimed(nodeId);
        }

        _requireMinStake(msg.sender);

        info.nodeId = nodeId;
        info.region = region;
        info.endpointURL = endpointURL;
        info.tlsPubkeyHash = tlsPubkeyHash;
        info.registeredAt = uint48(block.timestamp);
        info.active = true;

        if (nodeId != bytes32(0)) {
            nodeIdToEdgeProvider[nodeId] = msg.sender;
        }

        emit EdgeProviderRegistered(msg.sender, nodeId, region);
    }

    /// @notice Self-deregister as an edge provider. Sets active=false, clears the
    ///         nodeId mapping so the node ID can be re-claimed, and wipes the
    ///         endpoint metadata so getEdgeProviderInfo never returns stale
    ///         routing data for a deregistered provider. `frozen` is intentionally
    ///         preserved so a frozen provider cannot escape the freeze by
    ///         deregistering and re-registering.
    function deregisterEdgeProvider() external nonReentrant {
        EdgeProviderInfo storage info = edgeProviders[msg.sender];
        if (!info.active) revert NotActive(msg.sender);

        info.active = false;
        if (info.nodeId != bytes32(0)) {
            delete nodeIdToEdgeProvider[info.nodeId];
            info.nodeId = bytes32(0);
        }
        info.region = bytes32(0);
        delete info.endpointURL;
        delete info.tlsPubkeyHash;

        emit EdgeProviderDeregistered(msg.sender);
    }

    /// @notice Update the endpoint metadata for an active edge provider.
    /// @param endpointURL New public edge endpoint.
    /// @param tlsPubkeyHash New TLS public key hash.
    function updateEndpoint(string calldata endpointURL, bytes calldata tlsPubkeyHash) external {
        EdgeProviderInfo storage info = edgeProviders[msg.sender];
        if (!info.active) revert NotActive(msg.sender);
        if (info.frozen) revert ProviderIsFrozen(msg.sender);

        info.endpointURL = endpointURL;
        info.tlsPubkeyHash = tlsPubkeyHash;

        emit EdgeEndpointUpdated(msg.sender, endpointURL);
    }

    // ──────────────────────────────────────────────
    //  External: Graduated Emergency Response
    // ──────────────────────────────────────────────

    /// @notice Freeze an edge provider. A frozen provider cannot update its
    ///         endpoint and is reported as inactive to consumers.
    /// @param provider The edge provider to freeze.
    function freezeEdgeProvider(address provider) external onlyRole(SLASHER_ROLE) {
        EdgeProviderInfo storage info = edgeProviders[provider];
        if (!info.active) revert NotActive(provider);

        info.frozen = true;

        emit EdgeProviderFrozen(provider, msg.sender);
    }

    /// @notice Unfreeze a previously frozen edge provider.
    /// @param provider The edge provider to unfreeze.
    function unfreezeEdgeProvider(address provider) external onlyOwner {
        EdgeProviderInfo storage info = edgeProviders[provider];
        info.frozen = false;

        emit EdgeProviderUnfrozen(provider);
    }

    // ──────────────────────────────────────────────
    //  External: Edge-Slash Proposal & Appeal System
    // ──────────────────────────────────────────────

    /// @notice Propose slashing an edge provider's stake. Subject to the appeal
    ///         window snapshotted at proposal time.
    /// @dev Least-privilege: the registry holds SLASHER_ROLE on BunkerStaking, so
    ///      it must only ever propose slashes against providers it actually
    ///      governs — i.e. registered, active edge providers. The active check
    ///      prevents this contract from being used to slash arbitrary BunkerStaking
    ///      stakers who never opted into the edge role. As a defensive pre-check
    ///      (mirroring BunkerStaking.proposeSlash, H-05) the amount is also
    ///      validated against the provider's current slashable balance
    ///      (stakedAmount + totalUnbonding) at propose time; this is informational
    ///      only — executeEdgeSlash relies on BunkerStaking's own re-validation at
    ///      execution time for the binding check.
    /// @param provider The edge provider to propose slashing.
    /// @param amount Amount to slash (wei).
    /// @param reason Human-readable reason for the slash.
    /// @return proposalId The ID of the created proposal.
    function proposeEdgeSlash(address provider, uint128 amount, string calldata reason)
        external
        onlyRole(SLASHER_ROLE)
        returns (uint256 proposalId)
    {
        if (provider == address(0)) revert ZeroAddress();
        if (amount == 0) revert ZeroAmount();
        if (!edgeProviders[provider].active) revert NotActive(provider);

        // Defensive pre-check: reject proposals that exceed the provider's
        // current slashable balance in BunkerStaking. Decodes stakedAmount +
        // totalUnbonding from the existing getProviderInfo tuple (no new
        // BunkerStaking surface coupled).
        (uint128 stakedAmount, uint128 totalUnbonding,,,,,,,) =
            stakingContract.getProviderInfo(provider);
        uint256 slashable = uint256(stakedAmount) + uint256(totalUnbonding);
        if (uint256(amount) > slashable) {
            revert InsufficientSlashableBalance(uint256(amount), slashable);
        }

        proposalId = slashProposalCount++;
        slashProposals[proposalId] = EdgeSlashProposal({
            provider: provider,
            amount: amount,
            proposedAt: uint48(block.timestamp),
            executed: false,
            appealed: false,
            resolved: false,
            appealWindowSnapshot: appealWindow,
            reason: reason
        });

        emit EdgeSlashProposed(proposalId, provider, amount, reason);
    }

    /// @notice Execute an edge-slash proposal after the appeal window has elapsed.
    /// @dev Routes the slash to BunkerStaking.slash, which requires this contract
    ///      to hold SLASHER_ROLE on BunkerStaking (granted at deployment setup).
    /// @param proposalId The proposal to execute.
    function executeEdgeSlash(uint256 proposalId) external onlyRole(SLASHER_ROLE) nonReentrant {
        if (!slashingEnabled) revert SlashingNotEnabled();
        EdgeSlashProposal storage proposal = _getValidProposal(proposalId);
        if (proposal.executed) revert ProposalAlreadyExecuted(proposalId);
        if (proposal.appealed) revert ProposalAppealed(proposalId);

        uint256 windowEnd = uint256(proposal.proposedAt) + proposal.appealWindowSnapshot;
        if (block.timestamp < windowEnd) revert AppealWindowNotElapsed(proposalId);

        proposal.executed = true;

        stakingContract.slash(proposal.provider, proposal.amount);

        emit EdgeSlashExecuted(proposalId, proposal.provider, proposal.amount);
    }

    /// @notice Appeal an edge-slash proposal during the appeal window.
    /// @dev Only the targeted provider can appeal.
    /// @param proposalId The proposal to appeal.
    function appealEdgeSlash(uint256 proposalId) external {
        EdgeSlashProposal storage proposal = _getValidProposal(proposalId);
        if (proposal.executed) revert ProposalAlreadyExecuted(proposalId);
        if (proposal.appealed) revert ProposalAppealed(proposalId);
        if (msg.sender != proposal.provider) {
            revert NotProposalProvider(proposalId, msg.sender);
        }

        uint256 windowEnd = uint256(proposal.proposedAt) + proposal.appealWindowSnapshot;
        if (block.timestamp >= windowEnd) revert AppealWindowElapsed(proposalId);

        proposal.appealed = true;

        emit EdgeSlashAppealed(proposalId, proposal.provider);
    }

    /// @notice Resolve an appealed edge-slash proposal.
    /// @dev Only the owner can resolve. If upheld, the slash is executed.
    /// @param proposalId The proposal to resolve.
    /// @param uphold True to uphold the slash (execute it), false to dismiss.
    function resolveEdgeSlashAppeal(uint256 proposalId, bool uphold)
        external
        onlyOwner
        nonReentrant
    {
        EdgeSlashProposal storage proposal = _getValidProposal(proposalId);
        if (!proposal.appealed) revert ProposalNotAppealed(proposalId);
        if (proposal.resolved) revert ProposalAlreadyResolved(proposalId);
        if (proposal.executed) revert ProposalAlreadyExecuted(proposalId);

        proposal.resolved = true;

        if (uphold) {
            if (!slashingEnabled) revert SlashingNotEnabled();
            proposal.executed = true;

            stakingContract.slash(proposal.provider, proposal.amount);

            emit EdgeSlashExecuted(proposalId, proposal.provider, proposal.amount);
        }

        emit EdgeAppealResolved(proposalId, uphold);
    }

    // ──────────────────────────────────────────────
    //  External: Admin
    // ──────────────────────────────────────────────

    /// @notice Update the minimum BUNKER stake required to register as an edge
    ///         provider. Bounded by the BunkerStaking Starter and Platinum minima.
    /// @param newMinStake New minimum stake (wei).
    function setMinStakeForEdge(uint256 newMinStake) external onlyOwner {
        if (newMinStake < MIN_EDGE_STAKE_FLOOR || newMinStake > MIN_EDGE_STAKE_CEILING) {
            revert InvalidMinStake(newMinStake);
        }
        uint256 old = minStakeForEdge;
        minStakeForEdge = newMinStake;

        emit MinStakeUpdated(old, newMinStake);
    }

    /// @notice Adjust the appeal window for edge-slash proposals.
    /// @param newWindow New window in seconds (12 hours min, 14 days max).
    function setAppealWindow(uint256 newWindow) external onlyOwner {
        if (newWindow < APPEAL_WINDOW_FLOOR || newWindow > APPEAL_WINDOW_CEILING) {
            revert InvalidAppealWindow();
        }
        appealWindow = newWindow;

        emit AppealWindowUpdated(newWindow);
    }

    /// @notice Enable or disable slashing execution. When disabled, proposals can
    ///         still be created for monitoring but execution reverts.
    /// @param enabled True to enable slashing, false for monitor mode.
    function setSlashingEnabled(bool enabled) external onlyOwner {
        slashingEnabled = enabled;

        emit SlashingEnabledUpdated(enabled);
    }

    /// @notice Pause edge-provider registration and updates (emergency).
    function pause() external onlyOwner {
        _pause();
    }

    /// @notice Unpause edge-provider registration and updates.
    function unpause() external onlyOwner {
        _unpause();
    }

    // ──────────────────────────────────────────────
    //  External: Views
    // ──────────────────────────────────────────────

    /// @notice Returns true if the provider is registered, active, and not frozen.
    /// @param provider The edge-provider address to check.
    /// @return active True when the address can serve edge traffic.
    function isActiveEdgeProvider(address provider) external view returns (bool active) {
        EdgeProviderInfo storage info = edgeProviders[provider];
        return info.active && !info.frozen;
    }

    /// @notice Returns the full edge-provider metadata struct.
    /// @param provider The edge-provider address to look up.
    /// @return info The provider's registration metadata.
    function getEdgeProviderInfo(address provider)
        external
        view
        returns (EdgeProviderInfo memory info)
    {
        return edgeProviders[provider];
    }

    /// @notice Returns an edge-slash proposal by ID.
    /// @param proposalId The proposal ID.
    /// @return proposal The edge-slash proposal.
    function getEdgeSlashProposal(uint256 proposalId)
        external
        view
        returns (EdgeSlashProposal memory proposal)
    {
        return slashProposals[proposalId];
    }

    // ──────────────────────────────────────────────
    //  Internal
    // ──────────────────────────────────────────────

    /// @dev Sync DEFAULT_ADMIN_ROLE when ownership is transferred (mirrors M-05
    ///      in BunkerStaking).
    function _transferOwnership(address newOwner) internal virtual override {
        address oldOwner = owner();
        super._transferOwnership(newOwner);
        _revokeRole(DEFAULT_ADMIN_ROLE, oldOwner);
        _grantRole(DEFAULT_ADMIN_ROLE, newOwner);
    }

    /// @dev Reverts unless the address holds at least `minStakeForEdge` staked in
    ///      BunkerStaking. Reads storage (not pause-guarded), so the check still
    ///      works while BunkerStaking is paused.
    function _requireMinStake(address provider) internal view {
        (uint128 stakedAmount,,,,,,,,) = stakingContract.getProviderInfo(provider);
        if (uint256(stakedAmount) < minStakeForEdge) {
            revert InsufficientStake(uint256(stakedAmount), minStakeForEdge);
        }
    }

    /// @dev Validate a proposal ID and return the storage reference.
    function _getValidProposal(uint256 proposalId)
        internal
        view
        returns (EdgeSlashProposal storage)
    {
        if (proposalId >= slashProposalCount) revert InvalidProposalId(proposalId);
        return slashProposals[proposalId];
    }
}
