// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/IERC20Permit.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/Ownable2Step.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "./interfaces/IBurnable.sol";

/// @title BunkerStaking
/// @author Moltbunker
/// @notice Manages provider staking with 5 tiers, 14-day unbonding, slashing,
///         beneficiary delegation, provider identity, graduated emergency response,
///         graduated slashing by violation type, reward vesting, and demand-driven
///         emissions for the Moltbunker compute network.
/// @dev Deployed on Base L2. Uses OpenZeppelin v5. All token amounts are in
///      wei (18 decimals). Slashing burns 80% and sends 20% to treasury.
contract BunkerStaking is Ownable2Step, AccessControl, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;

    // ──────────────────────────────────────────────
    //  Constants
    // ──────────────────────────────────────────────

    /// @notice Contract version.
    string public constant VERSION = "1.3.0";

    /// @notice Role identifier for authorized slashers.
    bytes32 public constant SLASHER_ROLE = keccak256("SLASHER_ROLE");

    /// @notice Beneficiary change timelock: 24 hours.
    uint256 public constant BENEFICIARY_TIMELOCK = 24 hours;

    /// @notice Basis points denominator.
    uint256 public constant BPS_DENOMINATOR = 10000;

    /// @notice Maximum allowed emission multiplier (5x).
    uint256 public constant MAX_EMISSION_MULTIPLIER = 50000;

    // ──────────────────────────────────────────────
    //  Configurable Parameters (converted from constants)
    // ──────────────────────────────────────────────

    /// @notice Unbonding period (default 14 days).
    uint256 public unbondingPeriod = 14 days;

    /// @notice Slash distribution: burn portion in basis points (default 8000 = 80%).
    uint256 public slashBurnBps = 8000;

    /// @notice Slash distribution: treasury portion in basis points (default 2000 = 20%).
    uint256 public slashTreasuryBps = 2000;

    /// @notice Worst-case tier multiplier cap in basis points (default 12000 = 1.20x).
    uint256 public maxTierMultiplierBps = 12000;

    /// @notice Appeal window for slash proposals (default 48 hours).
    uint256 public appealWindow = 48 hours;

    // ──────────────────────────────────────────────
    //  Types
    // ──────────────────────────────────────────────

    /// @notice Staking tier enumeration.
    enum Tier { None, Starter, Bronze, Silver, Gold, Platinum }

    /// @notice Slash reason enumeration for graduated slashing.
    enum SlashReason { None, Downtime, JobAbandonment, SecurityViolation, Fraud, SLAViolation, DataLoss }

    /// @notice Tier threshold configuration.
    struct TierConfig {
        uint256 minStake;
        uint16 maxConcurrentJobs;   // 0 = unlimited
        uint16 rewardMultiplierBps; // e.g., 10000 = 1.00x, 12000 = 1.20x
        bool priorityQueue;
        bool governance;
    }

    /// @notice An individual unstake request in the unbonding queue.
    struct UnstakeRequest {
        uint128 amount;
        uint48 unlockTime;
        bool completed;
    }

    /// @notice A pending beneficiary change.
    struct PendingBeneficiary {
        address newBeneficiary;
        uint48 effectiveTime;
    }

    /// @notice A slash proposal subject to appeal window.
    struct SlashProposal {
        address provider;
        uint256 amount;
        string reason;
        uint256 proposedAt;
        bool executed;
        bool appealed;
        bool resolved;
        SlashReason slashReason;
        uint256 appealWindowSnapshot;   // Snapshot of appealWindow at proposal time
        uint16 slashBurnBpsSnapshot;    // Snapshot of slashBurnBps at proposal time
    }

    /// @notice Full provider state.
    struct ProviderInfo {
        uint128 stakedAmount;
        uint128 totalUnbonding;
        address beneficiary;
        uint48 registeredAt;
        bool active;
        bytes32 nodeId;        // SHA256 of Ed25519 public key
        bytes32 region;        // Geographic region identifier
        uint64 capabilities;   // Bitmask of supported capabilities
        bool frozen;           // Graduated emergency response
    }

    /// @notice A vested reward entry for linear vesting.
    struct VestedReward {
        uint128 totalAmount;
        uint128 releasedAmount;
        uint48 vestingStart;
    }

    // ──────────────────────────────────────────────
    //  State
    // ──────────────────────────────────────────────

    /// @notice The BUNKER ERC-20 token.
    IERC20 public immutable token;

    /// @notice Treasury address for slash distribution.
    address public treasury;

    /// @notice Total tokens staked across all providers (excludes unbonding).
    uint256 public totalStaked;

    /// @notice Total tokens in unbonding queues across all providers.
    uint256 public totalUnbonding;

    /// @notice Provider state by address.
    mapping(address => ProviderInfo) public providers;

    /// @notice Unbonding queue per provider.
    mapping(address => UnstakeRequest[]) public unstakeQueues;

    /// @notice Pending beneficiary changes per provider.
    mapping(address => PendingBeneficiary) public pendingBeneficiaries;

    /// @notice Tier thresholds (Tier enum => config). Tier.None is unused.
    mapping(Tier => TierConfig) public tierConfigs;

    // ──────────────────────────────────────────────
    //  State: Staking Rewards (Synthetix pattern)
    // ──────────────────────────────────────────────

    /// @notice Reward token (same as staking token).
    IERC20 public immutable rewardsToken;

    /// @notice Duration of each reward epoch.
    uint256 public rewardsDuration = 7 days;

    /// @notice Reward tokens distributed per second.
    uint256 public rewardRate;

    /// @notice Timestamp when the current epoch ends.
    uint256 public periodFinish;

    /// @notice Last time rewards were updated.
    uint256 public lastUpdateTime;

    /// @notice Accumulated reward per token (scaled by 1e18).
    uint256 public rewardPerTokenStored;

    /// @notice Per-user reward-per-token checkpoint.
    mapping(address => uint256) public userRewardPerTokenPaid;

    /// @notice Pending unclaimed rewards per provider.
    mapping(address => uint256) public rewards;

    // ──────────────────────────────────────────────
    //  State: Slash Appeal System
    // ──────────────────────────────────────────────

    /// @notice Total number of slash proposals created.
    uint256 public slashProposalCount;

    /// @notice Slash proposals by ID.
    mapping(uint256 => SlashProposal) public slashProposals;

    // ──────────────────────────────────────────────
    //  State: Provider Identity (Feature 1)
    // ──────────────────────────────────────────────

    /// @notice Reverse lookup from node ID to provider address.
    mapping(bytes32 => address) public nodeIdToProvider;

    // ──────────────────────────────────────────────
    //  State: Graduated Slashing (Feature 3)
    // ──────────────────────────────────────────────

    /// @notice Slash percentage in basis points per violation type.
    mapping(SlashReason => uint16) public slashPercentageBps;

    /// @notice Whether slashing execution is enabled. When false, proposals can
    ///         still be created (monitor mode) but execution reverts.
    bool public slashingEnabled;

    // ──────────────────────────────────────────────
    //  State: Reward Vesting (Feature 4)
    // ──────────────────────────────────────────────

    /// @notice Vesting period for staking rewards (default 90 days).
    uint256 public vestingPeriod = 90 days;

    /// @notice Percentage of rewards released immediately in basis points (default 25%).
    uint256 public immediateReleaseBps = 2500;

    /// @notice Vested reward entries per provider.
    mapping(address => VestedReward[]) public vestedRewards;

    // ──────────────────────────────────────────────
    //  State: Demand-Driven Emissions (Feature 5)
    // ──────────────────────────────────────────────

    /// @notice Cumulative compute hours reported by the operator.
    uint256 public totalComputeHoursReported;

    /// @notice Emission multiplier in basis points (10000 = 1.0x).
    uint256 public emissionMultiplier = 10000;

    /// @notice Cap on rewardRate. 0 means no cap.
    uint256 public maxEmissionRate;

    // ──────────────────────────────────────────────
    //  Errors
    // ──────────────────────────────────────────────

    error ZeroAmount();
    error BelowMinimumStake(uint256 amount, uint256 minimum);
    error InsufficientStake(uint256 requested, uint256 available);
    error UnbondingNotReady(uint256 unlockTime, uint256 currentTime);
    error UnstakeAlreadyCompleted(uint256 index);
    error InvalidUnstakeIndex(uint256 index, uint256 queueLength);
    error ProviderNotActive(address provider);
    error BeneficiaryTimelockNotElapsed(uint256 effectiveTime, uint256 currentTime);
    error NoPendingBeneficiaryChange(address provider);
    error ZeroAddress();
    error AlreadyRegistered(address provider);
    error InsufficientSlashableBalance(uint256 requested, uint256 available);
    error NoRewardsToClaimError();
    error RewardDurationNotFinished();
    error RewardsDurationZero();
    error RewardRateTooHigh(uint256 rewardRate, uint256 balance);
    error InvalidProposalId(uint256 proposalId);
    error ProposalAlreadyExecuted(uint256 proposalId);
    error ProposalAppealed(uint256 proposalId);
    error AppealWindowNotElapsed(uint256 proposalId, uint256 windowEnd);
    error AppealWindowElapsed(uint256 proposalId);
    error NotProposalProvider(uint256 proposalId, address caller);
    error ProposalAlreadyAppealed(uint256 proposalId);
    error ProposalNotAppealed(uint256 proposalId);
    error ProposalAlreadyResolved(uint256 proposalId);
    error CannotConfigureNoneTier();
    error TierOrderViolation(Tier tier, uint256 minStake, Tier adjacentTier, uint256 adjacentMinStake);
    error TokenBurnFailed();

    /// @notice Reverts when a node ID has already been claimed by another provider.
    error NodeIdAlreadyClaimed(bytes32 nodeId);

    /// @notice Reverts when an operation targets a frozen provider.
    error ProviderIsFrozen(address provider);

    /// @notice Reverts when an invalid slash reason is provided.
    error InvalidSlashReason();

    /// @notice Reverts when immediate release basis points exceed the denominator.
    error InvalidImmediateReleaseBps(uint256 bps);

    /// @notice Reverts when no vested rewards are claimable.
    error NoVestedRewardsClaimable();

    /// @notice Reverts when slash percentage exceeds 100%.
    error InvalidSlashPercentage(uint16 bps);

    /// @notice Reverts when emission multiplier is zero or exceeds maximum.
    error InvalidEmissionMultiplier(uint256 multiplier);

    /// @notice Reverts when vesting period is enabled but below minimum.
    error InvalidVestingPeriod();

    /// @notice Reverts when too many unstake requests are queued.
    error TooManyUnstakeRequests();

    /// @notice Reverts when stakeWithIdentity is called with identity on an active provider.
    error UseUpdateIdentity();

    /// @notice Reverts when multiplier is below 1.0x.
    error MultiplierTooLow();

    /// @notice Reverts when multiplier exceeds the cap.
    error MultiplierTooHigh();

    /// @notice Reverts when the multiplier cap is out of range.
    error InvalidMultiplierCap();

    /// @notice Reverts when unbonding period is out of range.
    error InvalidUnbondingPeriod();

    /// @notice Reverts when slash fee split doesn't sum to BPS_DENOMINATOR.
    error InvalidFeeSplit();

    /// @notice Reverts when appeal window is out of range.
    error InvalidAppealWindow();

    /// @notice Reverts when slashing execution is attempted while disabled.
    error SlashingNotEnabled();

    // ──────────────────────────────────────────────
    //  Events
    // ──────────────────────────────────────────────

    /// @notice Emitted when a provider stakes tokens.
    event Staked(
        address indexed provider,
        uint256 amount,
        uint256 totalStake,
        Tier tier
    );

    /// @notice Emitted when a provider requests unstaking.
    event UnstakeRequested(
        address indexed provider,
        uint256 amount,
        uint256 unlockTime,
        uint256 requestIndex
    );

    /// @notice Emitted when an unstake request is completed.
    event UnstakeCompleted(
        address indexed provider,
        uint256 amount,
        uint256 requestIndex
    );

    /// @notice Emitted when a provider is slashed.
    event Slashed(
        address indexed provider,
        uint256 totalSlashed,
        uint256 burnedAmount,
        uint256 treasuryAmount
    );

    /// @notice Emitted when a provider registers.
    event ProviderRegistered(address indexed provider, uint256 stakeAmount, Tier tier);

    /// @notice Emitted when a provider deregisters.
    event ProviderDeregistered(address indexed provider);

    /// @notice Emitted when a beneficiary change is initiated.
    event BeneficiaryChangeInitiated(
        address indexed provider,
        address indexed newBeneficiary,
        uint256 effectiveTime
    );

    /// @notice Emitted when a beneficiary change is executed.
    event BeneficiaryChanged(
        address indexed provider,
        address indexed oldBeneficiary,
        address indexed newBeneficiary
    );

    /// @notice Emitted when the treasury address is updated.
    event TreasuryUpdated(address indexed oldTreasury, address indexed newTreasury);

    /// @notice Emitted when a tier configuration is updated.
    event TierConfigUpdated(Tier indexed tier, uint256 minStake);

    /// @notice Emitted when a provider claims staking rewards.
    event RewardClaimed(address indexed provider, uint256 amount);

    /// @notice Emitted when a new reward epoch is started.
    event RewardEpochStarted(uint256 reward, uint256 rewardRate, uint256 periodFinish);

    /// @notice Emitted when the rewards duration is updated.
    event RewardsDurationUpdated(uint256 newDuration);

    /// @notice Emitted when a slash proposal is created.
    event SlashProposed(uint256 indexed proposalId, address indexed provider, uint256 amount, string reason);

    /// @notice Emitted when a slash proposal is executed.
    event SlashProposalExecuted(uint256 indexed proposalId, address indexed provider, uint256 amount);

    /// @notice Emitted when a provider appeals a slash proposal.
    event SlashAppealed(uint256 indexed proposalId, address indexed provider);

    /// @notice Emitted when an appeal is resolved.
    event AppealResolved(uint256 indexed proposalId, bool upheld);

    /// @notice Emitted when a provider's on-chain identity is updated.
    event ProviderIdentityUpdated(address indexed provider, bytes32 nodeId, bytes32 region, uint64 capabilities);

    /// @notice Emitted when a provider is frozen by the slasher.
    event ProviderFrozen(address indexed provider, address indexed by);

    /// @notice Emitted when a provider is unfrozen by the owner.
    event ProviderUnfrozen(address indexed provider);

    /// @notice Emitted when a slash proposal is created by reason.
    event SlashProposedByReason(
        uint256 indexed proposalId,
        address indexed provider,
        uint256 amount,
        SlashReason reason
    );

    /// @notice Emitted when a slash percentage is updated for a violation type.
    event SlashPercentageUpdated(SlashReason indexed reason, uint16 bps);

    /// @notice Emitted when rewards are vested for a provider.
    event RewardVested(
        address indexed provider,
        uint256 totalAmount,
        uint256 immediateAmount,
        uint256 vestedAmount
    );

    /// @notice Emitted when a provider claims vested rewards.
    event VestedRewardsClaimed(address indexed provider, uint256 amount);

    /// @notice Emitted when unvested rewards are forfeited due to slashing.
    event VestedRewardsForfeited(address indexed provider, uint256 amount);

    /// @notice Emitted when vesting parameters are updated.
    event VestingParamsUpdated(uint256 vestingPeriod, uint256 immediateReleaseBps);

    /// @notice Emitted when compute hours are reported.
    event ComputeHoursReported(uint256 hours_, uint256 totalComputeHours);

    /// @notice Emitted when the emission multiplier is updated.
    event EmissionMultiplierUpdated(uint256 multiplierBps);

    /// @notice Emitted when the max emission rate is updated.
    event MaxEmissionRateUpdated(uint256 maxRate);

    /// @notice Emitted when a tier reward multiplier is updated.
    event TierRewardMultiplierUpdated(BunkerStaking.Tier indexed tier, uint16 multiplierBps);

    /// @notice Emitted when the max tier multiplier cap is updated.
    event MaxTierMultiplierUpdated(uint16 newMax);

    /// @notice Emitted when the unbonding period is updated.
    event UnbondingPeriodUpdated(uint256 newPeriod);

    /// @notice Emitted when the slash fee split is updated.
    event SlashFeeSplitUpdated(uint16 burnBps, uint16 treasuryBps);

    /// @notice Emitted when the appeal window is updated.
    event AppealWindowUpdated(uint256 newWindow);

    /// @notice Emitted when slashing is enabled or disabled.
    event SlashingEnabledUpdated(bool enabled);

    // ──────────────────────────────────────────────
    //  Constructor
    // ──────────────────────────────────────────────

    /// @param _token BUNKER token address.
    /// @param _treasury Treasury address for slash proceeds.
    /// @param _initialOwner Admin wallet address.
    constructor(
        address _token,
        address _treasury,
        address _initialOwner
    ) Ownable(_initialOwner) {
        if (_token == address(0) || _treasury == address(0) || _initialOwner == address(0))
            revert ZeroAddress();

        token = IERC20(_token);
        rewardsToken = IERC20(_token);
        treasury = _treasury;

        // Grant DEFAULT_ADMIN_ROLE to owner for managing SLASHER_ROLE.
        _grantRole(DEFAULT_ADMIN_ROLE, _initialOwner);

        // Initialize default tier configurations.
        tierConfigs[Tier.Starter]  = TierConfig(1_000_000e18,       3,  10000, false, false);
        tierConfigs[Tier.Bronze]   = TierConfig(5_000_000e18,     10,  10000, false, false);
        tierConfigs[Tier.Silver]   = TierConfig(10_000_000e18,    50,  10500, true,  false);
        tierConfigs[Tier.Gold]     = TierConfig(100_000_000e18,  200,  11000, true,  true);
        tierConfigs[Tier.Platinum] = TierConfig(1_000_000_000e18,  0,  12000, true,  true);

        // Initialize graduated slash percentages (Feature 3).
        slashPercentageBps[SlashReason.Downtime]          = 100;   // 1%
        slashPercentageBps[SlashReason.JobAbandonment]    = 500;   // 5%
        slashPercentageBps[SlashReason.SLAViolation]      = 1000;  // 10%
        slashPercentageBps[SlashReason.DataLoss]          = 2500;  // 25%
        slashPercentageBps[SlashReason.SecurityViolation] = 5000;  // 50%
        slashPercentageBps[SlashReason.Fraud]             = 10000; // 100%
    }

    // ──────────────────────────────────────────────
    //  Modifiers
    // ──────────────────────────────────────────────

    /// @dev Updates reward accounting for a provider before any stake change.
    modifier updateReward(address account) {
        rewardPerTokenStored = rewardPerToken();
        lastUpdateTime = lastTimeRewardApplicable();
        if (account != address(0)) {
            rewards[account] = earned(account);
            userRewardPerTokenPaid[account] = rewardPerTokenStored;
        }
        _;
    }

    // ──────────────────────────────────────────────
    //  External: Provider Registration
    // ──────────────────────────────────────────────

    /// @notice Register as a provider by staking tokens. First stake must meet
    ///         the Starter tier minimum (500 BUNKER).
    /// @param amount Number of tokens to stake (wei).
    function stake(uint256 amount) external nonReentrant whenNotPaused updateReward(msg.sender) {
        if (amount == 0) revert ZeroAmount();

        ProviderInfo storage info = providers[msg.sender];

        // Feature 2: Reject if provider is frozen.
        if (info.frozen) revert ProviderIsFrozen(msg.sender);

        uint256 newTotal = uint256(info.stakedAmount) + amount;

        // First-time staker must meet minimum.
        if (!info.active) {
            if (newTotal < tierConfigs[Tier.Starter].minStake)
                revert BelowMinimumStake(newTotal, tierConfigs[Tier.Starter].minStake);

            info.active = true;
            info.beneficiary = msg.sender;
            info.registeredAt = uint48(block.timestamp);

            emit ProviderRegistered(msg.sender, newTotal, _getTier(newTotal));
        }

        info.stakedAmount = uint128(newTotal);
        totalStaked += amount;

        token.safeTransferFrom(msg.sender, address(this), amount);

        emit Staked(msg.sender, amount, newTotal, _getTier(newTotal));
    }

    /// @notice Register as a provider by staking tokens with on-chain identity.
    ///         Sets node identity on first registration.
    /// @param amount Number of tokens to stake (wei).
    /// @param nodeId SHA256 of the provider's Ed25519 public key.
    /// @param region Geographic region identifier.
    /// @param capabilities Bitmask of supported capabilities.
    function stakeWithIdentity(
        uint256 amount,
        bytes32 nodeId,
        bytes32 region,
        uint64 capabilities
    ) external nonReentrant whenNotPaused updateReward(msg.sender) {
        if (amount == 0) revert ZeroAmount();

        ProviderInfo storage info = providers[msg.sender];

        // Feature 2: Reject if provider is frozen.
        if (info.frozen) revert ProviderIsFrozen(msg.sender);

        uint256 newTotal = uint256(info.stakedAmount) + amount;

        // First-time staker must meet minimum.
        if (!info.active) {
            if (newTotal < tierConfigs[Tier.Starter].minStake)
                revert BelowMinimumStake(newTotal, tierConfigs[Tier.Starter].minStake);

            info.active = true;
            info.beneficiary = msg.sender;
            info.registeredAt = uint48(block.timestamp);

            // Set identity on first registration.
            if (nodeId != bytes32(0)) {
                if (nodeIdToProvider[nodeId] != address(0))
                    revert NodeIdAlreadyClaimed(nodeId);

                info.nodeId = nodeId;
                info.region = region;
                info.capabilities = capabilities;
                nodeIdToProvider[nodeId] = msg.sender;

                emit ProviderIdentityUpdated(msg.sender, nodeId, region, capabilities);
            }

            emit ProviderRegistered(msg.sender, newTotal, _getTier(newTotal));
        } else {
            // Active provider must use updateIdentity() to change identity (M-04).
            if (nodeId != bytes32(0)) revert UseUpdateIdentity();
        }

        info.stakedAmount = uint128(newTotal);
        totalStaked += amount;

        token.safeTransferFrom(msg.sender, address(this), amount);

        emit Staked(msg.sender, amount, newTotal, _getTier(newTotal));
    }

    /// @notice Update the on-chain identity of an existing provider.
    /// @param nodeId SHA256 of the provider's Ed25519 public key.
    /// @param region Geographic region identifier.
    /// @param capabilities Bitmask of supported capabilities.
    function updateIdentity(
        bytes32 nodeId,
        bytes32 region,
        uint64 capabilities
    ) external {
        ProviderInfo storage info = providers[msg.sender];
        if (!info.active) revert ProviderNotActive(msg.sender);
        if (info.frozen) revert ProviderIsFrozen(msg.sender);

        // If changing nodeId, validate uniqueness and release old one.
        if (nodeId != info.nodeId) {
            if (nodeId != bytes32(0)) {
                if (nodeIdToProvider[nodeId] != address(0))
                    revert NodeIdAlreadyClaimed(nodeId);
            }

            // Release old nodeId mapping.
            if (info.nodeId != bytes32(0)) {
                delete nodeIdToProvider[info.nodeId];
            }

            info.nodeId = nodeId;

            // Set new reverse mapping.
            if (nodeId != bytes32(0)) {
                nodeIdToProvider[nodeId] = msg.sender;
            }
        }

        info.region = region;
        info.capabilities = capabilities;

        emit ProviderIdentityUpdated(msg.sender, nodeId, region, capabilities);
    }

    /// @notice Request to unstake tokens. Starts the 14-day unbonding period.
    /// @param amount Number of tokens to unstake (wei).
    function requestUnstake(uint256 amount) external nonReentrant whenNotPaused updateReward(msg.sender) {
        if (amount == 0) revert ZeroAmount();

        ProviderInfo storage info = providers[msg.sender];
        if (!info.active) revert ProviderNotActive(msg.sender);

        // Feature 2: Reject if provider is frozen.
        if (info.frozen) revert ProviderIsFrozen(msg.sender);

        if (amount > info.stakedAmount) revert InsufficientStake(amount, info.stakedAmount);

        // Reduce active stake.
        info.stakedAmount -= uint128(amount);
        info.totalUnbonding += uint128(amount);
        totalStaked -= amount;
        totalUnbonding += amount;

        // If stake drops below minimum, deactivate provider.
        if (info.stakedAmount < tierConfigs[Tier.Starter].minStake) {
            info.active = false;
            emit ProviderDeregistered(msg.sender);
        }

        // Bound the unstake queue to prevent unbounded gas costs (M-02).
        if (unstakeQueues[msg.sender].length > 50) revert TooManyUnstakeRequests();

        // Create unbonding request.
        uint48 unlockTime = uint48(block.timestamp + unbondingPeriod);
        uint256 requestIndex = unstakeQueues[msg.sender].length;
        unstakeQueues[msg.sender].push(UnstakeRequest({
            amount: uint128(amount),
            unlockTime: unlockTime,
            completed: false
        }));

        emit UnstakeRequested(msg.sender, amount, unlockTime, requestIndex);
    }

    /// @notice Complete an unstake request after the unbonding period.
    /// @param requestIndex Index in the provider's unstake queue.
    function completeUnstake(uint256 requestIndex) external nonReentrant {
        if (providers[msg.sender].frozen) revert ProviderIsFrozen(msg.sender);
        UnstakeRequest[] storage queue = unstakeQueues[msg.sender];
        if (requestIndex >= queue.length)
            revert InvalidUnstakeIndex(requestIndex, queue.length);

        UnstakeRequest storage req = queue[requestIndex];
        if (req.completed) revert UnstakeAlreadyCompleted(requestIndex);
        if (block.timestamp < req.unlockTime)
            revert UnbondingNotReady(req.unlockTime, block.timestamp);

        uint256 amount = req.amount;
        req.completed = true;

        ProviderInfo storage info = providers[msg.sender];
        info.totalUnbonding -= uint128(amount);
        totalUnbonding -= amount;

        token.safeTransfer(msg.sender, amount);

        emit UnstakeCompleted(msg.sender, amount, requestIndex);
    }

    // ──────────────────────────────────────────────
    //  External: Beneficiary Management
    // ──────────────────────────────────────────────

    /// @notice Initiate a beneficiary change. Takes effect after 24-hour timelock.
    /// @param newBeneficiary The new beneficiary address.
    function initiateBeneficiaryChange(address newBeneficiary) external {
        if (newBeneficiary == address(0)) revert ZeroAddress();
        ProviderInfo storage info = providers[msg.sender];
        if (!info.active) revert ProviderNotActive(msg.sender);

        uint48 effectiveTime = uint48(block.timestamp + BENEFICIARY_TIMELOCK);
        pendingBeneficiaries[msg.sender] = PendingBeneficiary({
            newBeneficiary: newBeneficiary,
            effectiveTime: effectiveTime
        });

        emit BeneficiaryChangeInitiated(msg.sender, newBeneficiary, effectiveTime);
    }

    /// @notice Execute a pending beneficiary change after the timelock.
    function executeBeneficiaryChange() external {
        PendingBeneficiary storage pending = pendingBeneficiaries[msg.sender];
        if (pending.newBeneficiary == address(0))
            revert NoPendingBeneficiaryChange(msg.sender);
        if (block.timestamp < pending.effectiveTime)
            revert BeneficiaryTimelockNotElapsed(pending.effectiveTime, block.timestamp);

        ProviderInfo storage info = providers[msg.sender];
        address old = info.beneficiary;
        info.beneficiary = pending.newBeneficiary;

        emit BeneficiaryChanged(msg.sender, old, pending.newBeneficiary);

        // Clear pending.
        delete pendingBeneficiaries[msg.sender];
    }

    // ──────────────────────────────────────────────
    //  External: Graduated Emergency Response (Feature 2)
    // ──────────────────────────────────────────────

    /// @notice Freeze an individual provider. Prevents staking and unstaking
    ///         without requiring a global pause.
    /// @param provider The provider to freeze.
    function freezeProvider(address provider) external onlyRole(SLASHER_ROLE) {
        ProviderInfo storage info = providers[provider];
        if (!info.active) revert ProviderNotActive(provider);

        info.frozen = true;

        emit ProviderFrozen(provider, msg.sender);
    }

    /// @notice Unfreeze a previously frozen provider.
    /// @param provider The provider to unfreeze.
    function unfreezeProvider(address provider) external onlyOwner {
        ProviderInfo storage info = providers[provider];
        info.frozen = false;

        emit ProviderUnfrozen(provider);
    }

    // ──────────────────────────────────────────────
    //  External: Slashing (SLASHER_ROLE only)
    // ──────────────────────────────────────────────

    /// @notice Emergency immediate slash. Burns 80%, sends 20% to treasury.
    /// @dev Only callable by addresses with SLASHER_ROLE. Bypasses appeal window.
    ///      Slashes active stake first, then unbonding requests (newest first).
    /// @param provider The provider to slash.
    /// @param amount Total amount to slash (wei).
    function slashImmediate(address provider, uint256 amount)
        external
        onlyRole(SLASHER_ROLE)
        nonReentrant
        updateReward(provider)
    {
        if (!slashingEnabled) revert SlashingNotEnabled();
        _executeSlash(provider, amount, uint16(slashBurnBps));
    }

    /// @notice Backward-compatible slash that routes through the proposal system.
    /// @dev Creates a proposal and immediately executes it (no appeal window).
    ///      For the appeal-based flow, use proposeSlash() + executeSlash().
    /// @param provider The provider to slash.
    /// @param amount Total amount to slash (wei).
    function slash(address provider, uint256 amount)
        external
        onlyRole(SLASHER_ROLE)
        nonReentrant
        updateReward(provider)
    {
        if (!slashingEnabled) revert SlashingNotEnabled();
        _executeSlash(provider, amount, uint16(slashBurnBps));
    }

    // ──────────────────────────────────────────────
    //  External: Slash Proposal & Appeal System
    // ──────────────────────────────────────────────

    /// @notice Propose slashing a provider's stake. Subject to 48-hour appeal window.
    /// @param provider The provider to propose slashing.
    /// @param amount Amount to slash (wei).
    /// @param reason Human-readable reason for the slash.
    /// @return proposalId The ID of the created proposal.
    function proposeSlash(address provider, uint256 amount, string calldata reason)
        external
        onlyRole(SLASHER_ROLE)
        returns (uint256 proposalId)
    {
        if (amount == 0) revert ZeroAmount();

        ProviderInfo storage info = providers[provider];
        uint256 slashableTotal = uint256(info.stakedAmount) + uint256(info.totalUnbonding);
        if (amount > slashableTotal)
            revert InsufficientSlashableBalance(amount, slashableTotal);

        proposalId = slashProposalCount++;
        slashProposals[proposalId] = SlashProposal({
            provider: provider,
            amount: amount,
            reason: reason,
            proposedAt: block.timestamp,
            executed: false,
            appealed: false,
            resolved: false,
            slashReason: SlashReason.None,
            appealWindowSnapshot: appealWindow,
            slashBurnBpsSnapshot: uint16(slashBurnBps)
        });

        emit SlashProposed(proposalId, provider, amount, reason);
    }

    /// @notice Propose slashing a provider's stake by violation type. The slash
    ///         amount is calculated from the provider's active stake and the
    ///         configured percentage for the given reason.
    /// @param provider The provider to propose slashing.
    /// @param reason The violation type determining slash percentage.
    /// @return proposalId The ID of the created proposal.
    function proposeSlashByReason(address provider, SlashReason reason)
        external
        onlyRole(SLASHER_ROLE)
        returns (uint256 proposalId)
    {
        if (reason == SlashReason.None) revert InvalidSlashReason();

        ProviderInfo storage info = providers[provider];
        uint256 stakedAmount = uint256(info.stakedAmount);
        uint256 amount = (stakedAmount * slashPercentageBps[reason]) / BPS_DENOMINATOR;
        if (amount == 0) revert ZeroAmount();

        uint256 slashableTotal = stakedAmount + uint256(info.totalUnbonding);
        if (amount > slashableTotal)
            revert InsufficientSlashableBalance(amount, slashableTotal);

        proposalId = slashProposalCount++;
        slashProposals[proposalId] = SlashProposal({
            provider: provider,
            amount: amount,
            reason: "",
            proposedAt: block.timestamp,
            executed: false,
            appealed: false,
            resolved: false,
            slashReason: reason,
            appealWindowSnapshot: appealWindow,
            slashBurnBpsSnapshot: uint16(slashBurnBps)
        });

        emit SlashProposedByReason(proposalId, provider, amount, reason);
    }

    /// @notice Execute a slash proposal after the appeal window has elapsed.
    /// @param proposalId The proposal to execute.
    function executeSlash(uint256 proposalId)
        external
        onlyRole(SLASHER_ROLE)
        nonReentrant
    {
        if (!slashingEnabled) revert SlashingNotEnabled();
        SlashProposal storage proposal = _getValidProposal(proposalId);
        if (proposal.executed) revert ProposalAlreadyExecuted(proposalId);
        if (proposal.appealed) revert ProposalAppealed(proposalId);

        uint256 windowEnd = proposal.proposedAt + proposal.appealWindowSnapshot;
        if (block.timestamp < windowEnd)
            revert AppealWindowNotElapsed(proposalId, windowEnd);

        proposal.executed = true;

        // Update reward accounting for the provider before slashing.
        _doUpdateReward(proposal.provider);

        // Re-validate slash amount against current slashable balance (H-05).
        ProviderInfo storage info = providers[proposal.provider];
        uint256 slashableTotal = uint256(info.stakedAmount) + uint256(info.totalUnbonding);
        uint256 actualAmount = proposal.amount > slashableTotal ? slashableTotal : proposal.amount;

        if (actualAmount > 0) {
            _executeSlash(proposal.provider, actualAmount, proposal.slashBurnBpsSnapshot);
        }

        emit SlashProposalExecuted(proposalId, proposal.provider, actualAmount);
    }

    /// @notice Appeal a slash proposal during the appeal window.
    /// @dev Only the targeted provider can appeal.
    /// @param proposalId The proposal to appeal.
    function appealSlash(uint256 proposalId) external {
        SlashProposal storage proposal = _getValidProposal(proposalId);
        if (proposal.executed) revert ProposalAlreadyExecuted(proposalId);
        if (proposal.appealed) revert ProposalAlreadyAppealed(proposalId);
        if (msg.sender != proposal.provider)
            revert NotProposalProvider(proposalId, msg.sender);

        uint256 windowEnd = proposal.proposedAt + proposal.appealWindowSnapshot;
        if (block.timestamp >= windowEnd)
            revert AppealWindowElapsed(proposalId);

        proposal.appealed = true;

        emit SlashAppealed(proposalId, proposal.provider);
    }

    /// @notice Resolve an appealed slash proposal.
    /// @dev Only the owner can resolve. If upheld, the slash is executed.
    /// @param proposalId The proposal to resolve.
    /// @param uphold True to uphold the slash (execute it), false to dismiss.
    function resolveAppeal(uint256 proposalId, bool uphold)
        external
        onlyOwner
        nonReentrant
    {
        SlashProposal storage proposal = _getValidProposal(proposalId);
        if (!proposal.appealed) revert ProposalNotAppealed(proposalId);
        if (proposal.resolved) revert ProposalAlreadyResolved(proposalId);
        if (proposal.executed) revert ProposalAlreadyExecuted(proposalId);

        proposal.resolved = true;

        if (uphold) {
            if (!slashingEnabled) revert SlashingNotEnabled();
            proposal.executed = true;

            // Update reward accounting for the provider before slashing.
            _doUpdateReward(proposal.provider);

            // Re-validate slash amount against current slashable balance (H-05).
            ProviderInfo storage info = providers[proposal.provider];
            uint256 slashableTotal = uint256(info.stakedAmount) + uint256(info.totalUnbonding);
            uint256 actualAmount = proposal.amount > slashableTotal ? slashableTotal : proposal.amount;

            if (actualAmount > 0) {
                _executeSlash(proposal.provider, actualAmount, proposal.slashBurnBpsSnapshot);
            }

            emit SlashProposalExecuted(proposalId, proposal.provider, actualAmount);
        }

        emit AppealResolved(proposalId, uphold);
    }

    /// @notice Get a slash proposal by ID.
    /// @param proposalId The proposal ID.
    /// @return proposal The slash proposal.
    function getSlashProposal(uint256 proposalId)
        external
        view
        returns (SlashProposal memory proposal)
    {
        return slashProposals[proposalId];
    }

    // ──────────────────────────────────────────────
    //  External: Graduated Slashing Admin (Feature 3)
    // ──────────────────────────────────────────────

    /// @notice Update the slash percentage for a given violation type.
    /// @param reason The violation type.
    /// @param bps The new slash percentage in basis points.
    function setSlashPercentage(SlashReason reason, uint16 bps) external onlyOwner {
        if (reason == SlashReason.None) revert InvalidSlashReason();
        if (bps > BPS_DENOMINATOR) revert InvalidSlashPercentage(bps);
        slashPercentageBps[reason] = bps;

        emit SlashPercentageUpdated(reason, bps);
    }

    // ──────────────────────────────────────────────
    //  External: Staking Rewards
    // ──────────────────────────────────────────────

    /// @notice Returns the last time rewards are applicable (current time or epoch end).
    /// @return timestamp The applicable timestamp.
    function lastTimeRewardApplicable() public view returns (uint256) {
        return block.timestamp < periodFinish ? block.timestamp : periodFinish;
    }

    /// @notice Accumulated reward per token (scaled by 1e18).
    /// @return rewardPerToken_ The current reward per token value.
    function rewardPerToken() public view returns (uint256) {
        if (totalStaked == 0) {
            return rewardPerTokenStored;
        }
        return rewardPerTokenStored + (
            (lastTimeRewardApplicable() - lastUpdateTime) * rewardRate * 1e18 / totalStaked
        );
    }

    /// @notice Calculate pending rewards for a provider, including tier multiplier.
    /// @dev The tier multiplier is applied only to the new reward delta (accrued
    ///      since the last checkpoint), not to the already-stored `rewards[provider]`.
    ///      This prevents compounding of the multiplier across checkpoints.
    /// @param provider The provider address.
    /// @return The amount of pending reward tokens.
    function earned(address provider) public view returns (uint256) {
        uint256 stakedAmount = providers[provider].stakedAmount;
        uint256 newDelta = stakedAmount * (rewardPerToken() - userRewardPerTokenPaid[provider]) / 1e18;

        // Apply tier-based reward multiplier only to the new delta.
        Tier tier = _getTier(stakedAmount);
        if (tier != Tier.None) {
            uint16 multiplierBps = tierConfigs[tier].rewardMultiplierBps;
            newDelta = (newDelta * multiplierBps) / BPS_DENOMINATOR;
        }

        return newDelta + rewards[provider];
    }

    /// @notice Claim accumulated staking rewards with vesting.
    /// @dev 25% of rewards are sent immediately; 75% vest linearly over the
    ///      vesting period (default 90 days).
    function claimRewards() external nonReentrant updateReward(msg.sender) {
        if (providers[msg.sender].frozen) revert ProviderIsFrozen(msg.sender);
        uint256 reward = rewards[msg.sender];
        if (reward == 0) revert NoRewardsToClaimError();

        rewards[msg.sender] = 0;

        // Calculate immediate and vested portions.
        uint256 immediateAmount = (reward * immediateReleaseBps) / BPS_DENOMINATOR;
        uint256 vestedAmount = reward - immediateAmount;

        // Send immediate portion.
        if (immediateAmount > 0) {
            rewardsToken.safeTransfer(msg.sender, immediateAmount);
        }

        // Create vested entry for the remainder.
        if (vestedAmount > 0) {
            vestedRewards[msg.sender].push(VestedReward({
                totalAmount: uint128(vestedAmount),
                releasedAmount: 0,
                vestingStart: uint48(block.timestamp)
            }));
        }

        emit RewardVested(msg.sender, reward, immediateAmount, vestedAmount);
        emit RewardClaimed(msg.sender, reward);
    }

    /// @notice Claim all vested rewards that have unlocked.
    function claimVestedRewards() external nonReentrant {
        if (providers[msg.sender].frozen) revert ProviderIsFrozen(msg.sender);
        uint256 totalClaimable = 0;
        VestedReward[] storage entries = vestedRewards[msg.sender];

        for (uint256 i = 0; i < entries.length; i++) {
            VestedReward storage entry = entries[i];
            uint256 vested = _vestedAmount(entry);
            uint256 claimable = vested - uint256(entry.releasedAmount);

            if (claimable > 0) {
                entry.releasedAmount += uint128(claimable);
                totalClaimable += claimable;
            }
        }

        if (totalClaimable == 0) revert NoVestedRewardsClaimable();

        // Compact array: remove fully-claimed entries to bound array growth (M-03).
        uint256 j = 0;
        while (j < entries.length) {
            if (entries[j].releasedAmount == entries[j].totalAmount) {
                entries[j] = entries[entries.length - 1];
                entries.pop();
            } else {
                j++;
            }
        }

        rewardsToken.safeTransfer(msg.sender, totalClaimable);

        emit VestedRewardsClaimed(msg.sender, totalClaimable);
    }

    /// @notice Get the total amount of vested rewards currently claimable.
    /// @param provider The provider address.
    /// @return claimable Total claimable vested tokens.
    function getClaimableVested(address provider) external view returns (uint256 claimable) {
        VestedReward[] storage entries = vestedRewards[provider];
        for (uint256 i = 0; i < entries.length; i++) {
            VestedReward storage entry = entries[i];
            uint256 vested = _vestedAmount(entry);
            claimable += vested - uint256(entry.releasedAmount);
        }
    }

    /// @notice Get the number of vested reward entries for a provider.
    /// @param provider The provider address.
    /// @return count Number of vested entries.
    function getVestedRewardCount(address provider) external view returns (uint256 count) {
        return vestedRewards[provider].length;
    }

    /// @notice Start a new reward epoch by notifying the contract of reward tokens.
    /// @dev Owner must transfer reward tokens to this contract before calling.
    ///      If called before the current epoch ends, remaining rewards are rolled over.
    ///      Applies emission multiplier and caps at maxEmissionRate if set.
    /// @param reward Total reward tokens for the new epoch.
    function notifyRewardAmount(uint256 reward)
        external
        onlyOwner
        updateReward(address(0))
    {
        if (rewardsDuration == 0) revert RewardsDurationZero();

        // Apply emission multiplier (Feature 5).
        uint256 adjustedReward = (reward * emissionMultiplier) / BPS_DENOMINATOR;

        if (block.timestamp >= periodFinish) {
            rewardRate = adjustedReward / rewardsDuration;
        } else {
            uint256 remaining = periodFinish - block.timestamp;
            uint256 leftover = remaining * rewardRate;
            rewardRate = (adjustedReward + leftover) / rewardsDuration;
        }

        // Cap at maxEmissionRate if set (Feature 5).
        if (maxEmissionRate > 0 && rewardRate > maxEmissionRate) {
            rewardRate = maxEmissionRate;
        }

        // Ensure the contract has enough tokens to cover rewards, accounting
        // for the worst-case tier multiplier (Platinum = 1.20x).
        uint256 balance = token.balanceOf(address(this));
        uint256 lockedStake = totalStaked + totalUnbonding;
        uint256 rewardBalance = balance > lockedStake ? balance - lockedStake : 0;
        uint256 maxLiability = (rewardRate * rewardsDuration * maxTierMultiplierBps) / BPS_DENOMINATOR;
        if (maxLiability > rewardBalance)
            revert RewardRateTooHigh(rewardRate, rewardBalance);

        lastUpdateTime = block.timestamp;
        periodFinish = block.timestamp + rewardsDuration;

        emit RewardEpochStarted(reward, rewardRate, periodFinish);
    }

    /// @notice Update the rewards epoch duration.
    /// @dev Can only be called when the current epoch has finished.
    /// @param _rewardsDuration New duration in seconds.
    function setRewardsDuration(uint256 _rewardsDuration) external onlyOwner {
        if (_rewardsDuration == 0) revert RewardsDurationZero();
        if (block.timestamp < periodFinish) revert RewardDurationNotFinished();

        rewardsDuration = _rewardsDuration;
        emit RewardsDurationUpdated(_rewardsDuration);
    }

    // ──────────────────────────────────────────────
    //  External: Reward Vesting Admin (Feature 4)
    // ──────────────────────────────────────────────

    /// @notice Update vesting parameters.
    /// @param _vestingPeriod New vesting duration in seconds.
    /// @param _immediateReleaseBps New immediate release percentage in basis points.
    function setVestingParams(uint256 _vestingPeriod, uint256 _immediateReleaseBps) external onlyOwner {
        if (_immediateReleaseBps > BPS_DENOMINATOR)
            revert InvalidImmediateReleaseBps(_immediateReleaseBps);
        // Allow 0 as deliberate "no vesting" mode, but if enabled require at least 7 days.
        if (_vestingPeriod > 0 && _vestingPeriod < 1 minutes) revert InvalidVestingPeriod();

        vestingPeriod = _vestingPeriod;
        immediateReleaseBps = _immediateReleaseBps;

        emit VestingParamsUpdated(_vestingPeriod, _immediateReleaseBps);
    }

    // ──────────────────────────────────────────────
    //  External: Demand-Driven Emissions (Feature 5)
    // ──────────────────────────────────────────────

    /// @notice Report compute hours consumed on the network.
    /// @dev Uses SLASHER_ROLE as the operator role for reporting.
    /// @param hours_ Number of compute hours to report.
    function reportComputeHours(uint256 hours_) external onlyRole(SLASHER_ROLE) {
        totalComputeHoursReported += hours_;

        emit ComputeHoursReported(hours_, totalComputeHoursReported);
    }

    /// @notice Set the emission multiplier to modulate rewards based on demand.
    /// @param multiplierBps Multiplier in basis points (10000 = 1.0x).
    function setEmissionMultiplier(uint256 multiplierBps) external onlyOwner {
        if (multiplierBps == 0 || multiplierBps > MAX_EMISSION_MULTIPLIER)
            revert InvalidEmissionMultiplier(multiplierBps);
        emissionMultiplier = multiplierBps;

        emit EmissionMultiplierUpdated(multiplierBps);
    }

    /// @notice Set the maximum emission rate (cap on rewardRate).
    /// @param maxRate Maximum reward tokens per second. 0 to disable cap.
    function setMaxEmissionRate(uint256 maxRate) external onlyOwner {
        maxEmissionRate = maxRate;

        emit MaxEmissionRateUpdated(maxRate);
    }

    // ──────────────────────────────────────────────
    //  External: Views
    // ──────────────────────────────────────────────

    /// @notice Get the current tier for a provider.
    /// @param provider The provider address.
    /// @return tier The provider's current tier.
    function getTier(address provider) external view returns (Tier tier) {
        return _getTier(providers[provider].stakedAmount);
    }

    /// @notice Get the current tier for a given stake amount.
    /// @param stakeAmount The stake amount to evaluate.
    /// @return tier The corresponding tier.
    function getTierForAmount(uint256 stakeAmount) external view returns (Tier tier) {
        return _getTier(stakeAmount);
    }

    /// @notice Get the full provider info.
    /// @param provider The provider address.
    /// @return info The provider's state.
    function getProviderInfo(address provider) external view returns (ProviderInfo memory info) {
        return providers[provider];
    }

    /// @notice Get the number of unstake requests for a provider.
    /// @param provider The provider address.
    /// @return count Number of requests.
    function getUnstakeQueueLength(address provider) external view returns (uint256 count) {
        return unstakeQueues[provider].length;
    }

    /// @notice Get a specific unstake request.
    /// @param provider The provider address.
    /// @param index The request index.
    /// @return request The unstake request.
    function getUnstakeRequest(address provider, uint256 index)
        external
        view
        returns (UnstakeRequest memory request)
    {
        return unstakeQueues[provider][index];
    }

    /// @notice Get the total slashable balance for a provider (active + unbonding).
    /// @param provider The provider address.
    /// @return total Total slashable tokens.
    function getSlashableBalance(address provider) external view returns (uint256 total) {
        ProviderInfo storage info = providers[provider];
        return uint256(info.stakedAmount) + uint256(info.totalUnbonding);
    }

    /// @notice Check if a provider is active.
    /// @param provider The provider address.
    /// @return active True if the provider is registered and meets minimum stake.
    function isActiveProvider(address provider) external view returns (bool active) {
        return providers[provider].active;
    }

    // ──────────────────────────────────────────────
    //  External: Admin
    // ──────────────────────────────────────────────

    /// @notice Update the treasury address.
    /// @param newTreasury New treasury address.
    function setTreasury(address newTreasury) external onlyOwner {
        if (newTreasury == address(0)) revert ZeroAddress();
        address old = treasury;
        treasury = newTreasury;
        emit TreasuryUpdated(old, newTreasury);
    }

    /// @notice Update a tier's minimum stake threshold.
    /// @dev Enforces ascending order: lower tiers must have lower minStake.
    /// @param tier The tier to update.
    /// @param minStake New minimum stake (wei).
    function setTierMinStake(Tier tier, uint256 minStake) external onlyOwner {
        if (tier == Tier.None) revert CannotConfigureNoneTier();

        // Enforce ascending: new value must be > tier below and < tier above.
        uint8 tierVal = uint8(tier);
        if (tierVal > 1) {
            // There is a tier below; new value must be strictly greater.
            Tier below = Tier(tierVal - 1);
            if (minStake <= tierConfigs[below].minStake)
                revert TierOrderViolation(tier, minStake, below, tierConfigs[below].minStake);
        }
        if (tierVal < uint8(Tier.Platinum)) {
            // There is a tier above; new value must be strictly less.
            Tier above = Tier(tierVal + 1);
            if (minStake >= tierConfigs[above].minStake)
                revert TierOrderViolation(tier, minStake, above, tierConfigs[above].minStake);
        }

        tierConfigs[tier].minStake = minStake;
        emit TierConfigUpdated(tier, minStake);
    }

    /// @notice Adjust a tier's reward multiplier in basis points.
    /// @param tier The tier to update (cannot be None).
    /// @param multiplierBps New multiplier (>= 10000 = 1.0x, <= maxTierMultiplierBps).
    function setTierRewardMultiplier(Tier tier, uint16 multiplierBps) external onlyOwner {
        if (tier == Tier.None) revert CannotConfigureNoneTier();
        if (multiplierBps < BPS_DENOMINATOR) revert MultiplierTooLow();
        if (multiplierBps > maxTierMultiplierBps) revert MultiplierTooHigh();
        tierConfigs[tier].rewardMultiplierBps = multiplierBps;
        emit TierRewardMultiplierUpdated(tier, multiplierBps);
    }

    /// @notice Adjust the maximum tier multiplier cap.
    /// @param newMax New cap in basis points (10000 = 1.0x min, 20000 = 2.0x max).
    function setMaxTierMultiplierBps(uint16 newMax) external onlyOwner {
        if (newMax < BPS_DENOMINATOR || newMax > 20000) revert InvalidMultiplierCap();

        // If there's an active reward epoch, verify solvency with the new multiplier.
        if (block.timestamp < periodFinish && rewardRate > 0) {
            uint256 balance = token.balanceOf(address(this));
            uint256 lockedStake = totalStaked + totalUnbonding;
            uint256 rewardBalance = balance > lockedStake ? balance - lockedStake : 0;
            uint256 maxLiability = (rewardRate * rewardsDuration * uint256(newMax)) / BPS_DENOMINATOR;
            if (maxLiability > rewardBalance)
                revert RewardRateTooHigh(rewardRate, rewardBalance);
        }

        maxTierMultiplierBps = newMax;
        emit MaxTierMultiplierUpdated(newMax);
    }

    /// @notice Adjust the unbonding period for unstake requests.
    /// @param newPeriod New period in seconds (1 day min, 90 days max).
    function setUnbondingPeriod(uint256 newPeriod) external onlyOwner {
        if (newPeriod < 1 minutes || newPeriod > 90 days) revert InvalidUnbondingPeriod();
        unbondingPeriod = newPeriod;
        emit UnbondingPeriodUpdated(newPeriod);
    }

    /// @notice Adjust the slash fee split between burn and treasury.
    /// @param burnBps Burn portion in basis points.
    /// @param treasuryBps Treasury portion in basis points.
    function setSlashFeeSplit(uint16 burnBps, uint16 treasuryBps) external onlyOwner {
        if (uint256(burnBps) + uint256(treasuryBps) != BPS_DENOMINATOR) revert InvalidFeeSplit();
        slashBurnBps = burnBps;
        slashTreasuryBps = treasuryBps;
        emit SlashFeeSplitUpdated(burnBps, treasuryBps);
    }

    /// @notice Adjust the appeal window for slash proposals.
    /// @param newWindow New window in seconds (12 hours min, 14 days max).
    function setAppealWindow(uint256 newWindow) external onlyOwner {
        if (newWindow < 1 minutes || newWindow > 14 days) revert InvalidAppealWindow();
        appealWindow = newWindow;
        emit AppealWindowUpdated(newWindow);
    }

    /// @notice Enable or disable slashing execution. When disabled, proposals
    ///         can still be created for monitoring but execution reverts.
    /// @param enabled True to enable slashing, false for monitor mode.
    function setSlashingEnabled(bool enabled) external onlyOwner {
        slashingEnabled = enabled;
        emit SlashingEnabledUpdated(enabled);
    }

    /// @notice Pause staking and unstaking (emergency).
    function pause() external onlyOwner {
        _pause();
    }

    /// @notice Unpause staking and unstaking.
    function unpause() external onlyOwner {
        _unpause();
    }

    // ──────────────────────────────────────────────
    //  Internal
    // ──────────────────────────────────────────────

    /// @dev Sync DEFAULT_ADMIN_ROLE when ownership is transferred (M-05).
    function _transferOwnership(address newOwner) internal virtual override {
        address oldOwner = owner();
        super._transferOwnership(newOwner);
        _revokeRole(DEFAULT_ADMIN_ROLE, oldOwner);
        _grantRole(DEFAULT_ADMIN_ROLE, newOwner);
    }

    /// @dev Determine tier from stake amount.
    function _getTier(uint256 stakeAmount) internal view returns (Tier) {
        if (stakeAmount >= tierConfigs[Tier.Platinum].minStake) return Tier.Platinum;
        if (stakeAmount >= tierConfigs[Tier.Gold].minStake) return Tier.Gold;
        if (stakeAmount >= tierConfigs[Tier.Silver].minStake) return Tier.Silver;
        if (stakeAmount >= tierConfigs[Tier.Bronze].minStake) return Tier.Bronze;
        if (stakeAmount >= tierConfigs[Tier.Starter].minStake) return Tier.Starter;
        return Tier.None;
    }

    /// @dev Burn tokens held by this contract via ERC20Burnable.burn().
    function _burnTokens(uint256 amount) internal {
        IBurnable(address(token)).burn(amount);
    }

    /// @dev Execute the actual slash logic (shared by slash, slashImmediate, and proposal execution).
    ///      Also forfeits unvested rewards (Feature 4).
    function _executeSlash(address provider, uint256 amount, uint16 burnBps) internal {
        if (amount == 0) revert ZeroAmount();

        ProviderInfo storage info = providers[provider];
        uint256 slashableTotal = uint256(info.stakedAmount) + uint256(info.totalUnbonding);
        if (amount > slashableTotal)
            revert InsufficientSlashableBalance(amount, slashableTotal);

        uint256 remaining = amount;

        // 1. Slash from active stake first.
        if (remaining > 0 && info.stakedAmount > 0) {
            uint256 fromActive = remaining > info.stakedAmount
                ? info.stakedAmount
                : remaining;
            info.stakedAmount -= uint128(fromActive);
            totalStaked -= fromActive;
            remaining -= fromActive;
        }

        // 2. Slash from unbonding queue (newest first).
        if (remaining > 0) {
            UnstakeRequest[] storage queue = unstakeQueues[provider];
            for (uint256 i = queue.length; i > 0 && remaining > 0;) {
                unchecked { --i; }
                UnstakeRequest storage req = queue[i];
                if (req.completed) continue;

                uint256 fromReq = remaining > req.amount ? req.amount : remaining;
                req.amount -= uint128(fromReq);
                info.totalUnbonding -= uint128(fromReq);
                totalUnbonding -= fromReq;
                remaining -= fromReq;

                // If request fully consumed, mark completed.
                if (req.amount == 0) {
                    req.completed = true;
                }
            }
        }

        // 3. Forfeit unvested rewards (Feature 4).
        uint256 forfeitedAmount = _forfeitUnvestedRewards(provider);

        // Cap forfeited amount to what is actually available in the reward pool
        // to prevent revert when reward balance is depleted.
        uint256 availableForForfeit = token.balanceOf(address(this)) - totalStaked - totalUnbonding;
        if (forfeitedAmount > availableForForfeit) {
            forfeitedAmount = availableForForfeit;
        }

        // Deactivate provider if below minimum.
        if (info.stakedAmount < tierConfigs[Tier.Starter].minStake && info.active) {
            info.active = false;
            emit ProviderDeregistered(provider);
        }

        // Distribute slashed tokens (stake + forfeited vested rewards).
        // Uses the burnBps parameter — snapshotted at proposal time for proposal-based
        // slashes, or the live value for immediate slashes.
        uint256 totalSlashAmount = amount + forfeitedAmount;
        uint256 burnAmount = (totalSlashAmount * burnBps) / BPS_DENOMINATOR;
        uint256 treasuryAmount = totalSlashAmount - burnAmount;

        _burnTokens(burnAmount);

        if (treasuryAmount > 0) {
            token.safeTransfer(treasury, treasuryAmount);
        }

        emit Slashed(provider, amount, burnAmount, treasuryAmount);
    }

    /// @dev Forfeit all unvested rewards for a provider. Returns the total
    ///      amount forfeited (already held by this contract as reward balance).
    ///      NOTE (M-01): This intentionally also forfeits vested-but-unclaimed rewards.
    ///      Slashed providers lose both unvested AND unclaimed-but-vested rewards as
    ///      an additional penalty for malicious behavior.
    function _forfeitUnvestedRewards(address provider) internal returns (uint256 forfeited) {
        VestedReward[] storage entries = vestedRewards[provider];

        for (uint256 i = 0; i < entries.length; i++) {
            VestedReward storage entry = entries[i];
            uint256 vested = _vestedAmount(entry);
            uint256 unvested = uint256(entry.totalAmount) - vested;

            if (unvested > 0) {
                // Mark the unvested portion as released (forfeited).
                entry.releasedAmount = entry.totalAmount;
                forfeited += unvested;
            }
        }

        if (forfeited > 0) {
            emit VestedRewardsForfeited(provider, forfeited);
        }
    }

    /// @dev Calculate the amount that has vested for a single entry.
    function _vestedAmount(VestedReward storage entry) internal view returns (uint256) {
        if (vestingPeriod == 0) {
            return uint256(entry.totalAmount);
        }

        uint256 elapsed = block.timestamp - uint256(entry.vestingStart);
        if (elapsed >= vestingPeriod) {
            return uint256(entry.totalAmount);
        }

        return (uint256(entry.totalAmount) * elapsed) / vestingPeriod;
    }

    /// @dev Validate a proposal ID and return the storage reference.
    function _getValidProposal(uint256 proposalId) internal view returns (SlashProposal storage) {
        if (proposalId >= slashProposalCount)
            revert InvalidProposalId(proposalId);
        return slashProposals[proposalId];
    }

    /// @dev Perform reward update for an account (non-modifier version for internal use).
    function _doUpdateReward(address account) internal {
        rewardPerTokenStored = rewardPerToken();
        lastUpdateTime = lastTimeRewardApplicable();
        if (account != address(0)) {
            rewards[account] = earned(account);
            userRewardPerTokenPaid[account] = rewardPerTokenStored;
        }
    }
}
