// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/Ownable2Step.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";

interface IBunkerStakingProvider {
    function isActiveProvider(address provider) external view returns (bool);
}

/// @title BunkerDelegation
/// @author Moltbunker
/// @notice V1 delegation contract - token lockup and provider association.
/// @dev Reward distribution will be implemented in V2 via integration with BunkerStaking.
///      Current version enables: delegation, undelegation with unbonding, provider config.
///      NOT YET IMPLEMENTED: reward distribution to delegators based on provider earnings.
///      NOT YET IMPLEMENTED: slashing of delegated funds (delegated tokens are not slashable in V1).
///      Deployed on Base L2. Uses OpenZeppelin v5. Single delegation per address.
///      Providers must opt-in by configuring their delegation parameters.
///      7-day unbonding period (shorter than provider's 14-day unbonding).
contract BunkerDelegation is Ownable2Step, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;

    // ──────────────────────────────────────────────
    //  Constants
    // ──────────────────────────────────────────────

    /// @notice Contract version.
    string public constant VERSION = "1.0.0";

    /// @notice Unbonding period for delegated stake (default 7 days).
    uint256 public unbondingPeriod = 7 days;

    /// @notice Maximum reward cut a provider can take (default 5000 = 50%).
    uint256 public maxRewardCutBps = 5000;

    /// @notice Basis points denominator.
    uint256 public constant BPS_DENOMINATOR = 10000;

    /// @notice Maximum number of pending unbonding requests per delegator.
    uint256 public constant MAX_UNBONDING_QUEUE = 20;

    // ──────────────────────────────────────────────
    //  Types
    // ──────────────────────────────────────────────

    /// @notice Delegation state for a single delegator.
    struct DelegationInfo {
        address provider;
        uint128 amount;
        uint48 delegatedAt;
        bool active;
    }

    /// @notice Provider's delegation configuration and aggregate state.
    struct ProviderDelegationConfig {
        uint16 rewardCutBps;           // % provider keeps from delegator rewards
        uint16 pendingRewardCutBps;    // Pending reward cut increase (timelocked)
        uint48 rewardCutEffectiveAt;   // Timestamp when pending reward cut takes effect
        uint16 feeShareBps;            // % of escrow fees shared with delegators
        uint128 totalDelegated;        // Total delegated to this provider
        bool acceptingDelegations;
    }

    /// @notice A pending unbonding request from a delegator.
    struct UnbondingRequest {
        uint128 amount;
        uint48 unlockTime;
        bool completed;
    }

    // ──────────────────────────────────────────────
    //  State
    // ──────────────────────────────────────────────

    /// @notice The BUNKER ERC-20 token.
    IERC20 public immutable token;

    /// @notice BunkerStaking contract address (for reference/integration).
    address public stakingContract;

    /// @notice Total tokens delegated across all delegators.
    uint256 public totalDelegated;

    /// @notice Delegation state per delegator address.
    mapping(address => DelegationInfo) public delegations;

    /// @notice Delegation configuration per provider address.
    mapping(address => ProviderDelegationConfig) public providerConfigs;

    /// @notice Unbonding queue per delegator address.
    mapping(address => UnbondingRequest[]) public unbondingQueues;

    // ──────────────────────────────────────────────
    //  Errors
    // ──────────────────────────────────────────────

    error ZeroAmount();
    error ZeroAddress();
    error ProviderNotAccepting(address provider);
    error AlreadyDelegated(address delegator);
    error NotDelegated(address delegator);
    error RewardCutTooHigh(uint16 requested, uint16 maximum);
    error InsufficientDelegation(uint256 requested, uint256 available);
    error UnbondingNotReady(uint256 unlockTime, uint256 currentTime);
    error InvalidUnbondingIndex(uint256 index, uint256 queueLength);
    error UnbondingAlreadyCompleted(uint256 index);
    error ProviderNotActive(address provider);
    error TooManyUnbondingRequests();
    error NoPendingRewardCut();
    error RewardCutTimelockActive();
    error InvalidUnbondingPeriod();
    error InvalidRewardCutCap();

    // ──────────────────────────────────────────────
    //  Events
    // ──────────────────────────────────────────────

    /// @notice Emitted when a delegator delegates tokens to a provider.
    event Delegated(address indexed delegator, address indexed provider, uint256 amount);

    /// @notice Emitted when a delegator withdraws their delegation.
    event DelegationWithdrawn(address indexed delegator, address indexed provider, uint256 amount);

    /// @notice Emitted when a delegator requests undelegation.
    event UndelegateRequested(address indexed delegator, uint256 amount, uint256 unlockTime);

    /// @notice Emitted when an undelegation unbonding period completes.
    event UndelegateCompleted(address indexed delegator, uint256 amount);

    /// @notice Emitted when a provider updates their delegation configuration.
    event ProviderConfigUpdated(address indexed provider, uint16 rewardCutBps, uint16 feeShareBps);

    /// @notice Emitted when a provider toggles delegation acceptance.
    event DelegationAcceptanceToggled(address indexed provider, bool accepting);

    /// @notice Emitted when a reward cut increase is scheduled (timelocked).
    event RewardCutIncreaseScheduled(address indexed provider, uint16 newRewardCutBps, uint48 effectiveAt);

    /// @notice Emitted when a pending reward cut increase is finalized.
    event RewardCutFinalized(address indexed provider, uint16 rewardCutBps);

    /// @notice Emitted when the staking contract reference is updated.
    event StakingContractUpdated(address indexed oldStaking, address indexed newStaking);

    /// @notice Emitted when the unbonding period is updated.
    event UnbondingPeriodUpdated(uint256 newPeriod);

    /// @notice Emitted when the max reward cut cap is updated.
    event MaxRewardCutUpdated(uint16 newMax);

    // ──────────────────────────────────────────────
    //  Constructor
    // ──────────────────────────────────────────────

    /// @param _token BUNKER token address.
    /// @param _stakingContract BunkerStaking contract address.
    /// @param _initialOwner Admin wallet address.
    constructor(
        address _token,
        address _stakingContract,
        address _initialOwner
    ) Ownable(_initialOwner) {
        if (_token == address(0) || _initialOwner == address(0))
            revert ZeroAddress();

        token = IERC20(_token);
        stakingContract = _stakingContract;
    }

    // ──────────────────────────────────────────────
    //  External: Provider Configuration
    // ──────────────────────────────────────────────

    /// @notice Configure delegation parameters for the calling provider.
    /// @dev The provider must call this before they can accept delegations.
    ///      Reward cut increases are timelocked (UNBONDING_PERIOD) to prevent bait-and-switch.
    ///      Reward cut decreases take effect immediately (benefits delegators).
    /// @param rewardCutBps Percentage the provider keeps from delegator rewards (basis points).
    /// @param feeShareBps Percentage of escrow fees shared with delegators (basis points).
    function setDelegationConfig(uint16 rewardCutBps, uint16 feeShareBps) external {
        if (rewardCutBps > uint16(maxRewardCutBps))
            revert RewardCutTooHigh(rewardCutBps, uint16(maxRewardCutBps));

        ProviderDelegationConfig storage config = providerConfigs[msg.sender];

        if (rewardCutBps > config.rewardCutBps) {
            // Increases require timelock to prevent bait-and-switch (C-09)
            config.pendingRewardCutBps = rewardCutBps;
            config.rewardCutEffectiveAt = uint48(block.timestamp + unbondingPeriod);
            emit RewardCutIncreaseScheduled(msg.sender, rewardCutBps, config.rewardCutEffectiveAt);
        } else {
            // Decreases take effect immediately (benefits delegators)
            config.rewardCutBps = rewardCutBps;
        }
        config.feeShareBps = feeShareBps;

        emit ProviderConfigUpdated(msg.sender, rewardCutBps, feeShareBps);
    }

    /// @notice Finalize a pending reward cut increase after the timelock has elapsed.
    function finalizeRewardCut() external {
        ProviderDelegationConfig storage config = providerConfigs[msg.sender];
        if (config.pendingRewardCutBps == 0) revert NoPendingRewardCut();
        if (block.timestamp < config.rewardCutEffectiveAt) revert RewardCutTimelockActive();
        config.rewardCutBps = config.pendingRewardCutBps;
        config.pendingRewardCutBps = 0;
        config.rewardCutEffectiveAt = 0;
        emit RewardCutFinalized(msg.sender, config.rewardCutBps);
    }

    /// @notice Toggle whether the calling provider accepts new delegations.
    /// @dev When enabling, validates the provider is active in the staking contract (H-17).
    /// @param accepting True to accept delegations, false to stop.
    function toggleAcceptDelegations(bool accepting) external {
        if (accepting && stakingContract != address(0)) {
            if (!IBunkerStakingProvider(stakingContract).isActiveProvider(msg.sender))
                revert ProviderNotActive(msg.sender);
        }
        providerConfigs[msg.sender].acceptingDelegations = accepting;
        emit DelegationAcceptanceToggled(msg.sender, accepting);
    }

    // ──────────────────────────────────────────────
    //  External: Delegator Actions
    // ──────────────────────────────────────────────

    /// @notice Delegate tokens to a provider.
    /// @dev Each delegator can only delegate to one provider at a time.
    ///      To delegate to a different provider, first undelegate completely.
    /// @param provider The provider to delegate to.
    /// @param amount Number of tokens to delegate (wei).
    function delegate(address provider, uint256 amount) external nonReentrant whenNotPaused {
        if (amount == 0) revert ZeroAmount();
        if (provider == address(0)) revert ZeroAddress();
        if (stakingContract != address(0)) {
            if (!IBunkerStakingProvider(stakingContract).isActiveProvider(provider))
                revert ProviderNotActive(provider);
        }

        ProviderDelegationConfig storage config = providerConfigs[provider];
        if (!config.acceptingDelegations) revert ProviderNotAccepting(provider);

        DelegationInfo storage info = delegations[msg.sender];
        if (info.active) revert AlreadyDelegated(msg.sender);

        info.provider = provider;
        info.amount = uint128(amount);
        info.delegatedAt = uint48(block.timestamp);
        info.active = true;

        config.totalDelegated += uint128(amount);
        totalDelegated += amount;

        token.safeTransferFrom(msg.sender, address(this), amount);

        emit Delegated(msg.sender, provider, amount);
    }

    /// @notice Request to undelegate tokens. Starts the 7-day unbonding period.
    /// @param amount Number of tokens to undelegate (wei).
    function requestUndelegate(uint256 amount) external nonReentrant whenNotPaused {
        if (amount == 0) revert ZeroAmount();
        if (unbondingQueues[msg.sender].length >= MAX_UNBONDING_QUEUE) revert TooManyUnbondingRequests();

        DelegationInfo storage info = delegations[msg.sender];
        if (!info.active) revert NotDelegated(msg.sender);
        if (amount > info.amount) revert InsufficientDelegation(amount, info.amount);

        info.amount -= uint128(amount);

        ProviderDelegationConfig storage config = providerConfigs[info.provider];
        config.totalDelegated -= uint128(amount);
        totalDelegated -= amount;

        // If fully undelegated, deactivate.
        if (info.amount == 0) {
            info.active = false;
            emit DelegationWithdrawn(msg.sender, info.provider, amount);
        }

        // Create unbonding request.
        uint48 unlockTime = uint48(block.timestamp + unbondingPeriod);
        unbondingQueues[msg.sender].push(UnbondingRequest({
            amount: uint128(amount),
            unlockTime: unlockTime,
            completed: false
        }));

        emit UndelegateRequested(msg.sender, amount, unlockTime);
    }

    /// @notice Complete an undelegation after the unbonding period has elapsed.
    /// @param requestIndex Index in the delegator's unbonding queue.
    function completeUndelegate(uint256 requestIndex) external nonReentrant {
        UnbondingRequest[] storage queue = unbondingQueues[msg.sender];
        if (requestIndex >= queue.length)
            revert InvalidUnbondingIndex(requestIndex, queue.length);

        UnbondingRequest storage req = queue[requestIndex];
        if (req.completed) revert UnbondingAlreadyCompleted(requestIndex);
        if (block.timestamp < req.unlockTime)
            revert UnbondingNotReady(req.unlockTime, block.timestamp);

        uint256 amount = req.amount;
        req.completed = true;

        token.safeTransfer(msg.sender, amount);

        emit UndelegateCompleted(msg.sender, amount);
    }

    // ──────────────────────────────────────────────
    //  External: Views
    // ──────────────────────────────────────────────

    /// @notice Get the delegation info for a delegator.
    /// @param delegator The delegator address.
    /// @return info The delegation state.
    function getDelegation(address delegator)
        external
        view
        returns (DelegationInfo memory info)
    {
        return delegations[delegator];
    }

    /// @notice Get the delegation configuration for a provider.
    /// @param provider The provider address.
    /// @return config The provider's delegation configuration.
    function getProviderConfig(address provider)
        external
        view
        returns (ProviderDelegationConfig memory config)
    {
        return providerConfigs[provider];
    }

    /// @notice Get the total amount delegated to a provider.
    /// @param provider The provider address.
    /// @return total Total delegated tokens.
    function getTotalDelegatedTo(address provider) external view returns (uint256 total) {
        return providerConfigs[provider].totalDelegated;
    }

    /// @notice Get the number of unbonding requests for a delegator.
    /// @param delegator The delegator address.
    /// @return count Number of unbonding requests.
    function getUnbondingQueueLength(address delegator)
        external
        view
        returns (uint256 count)
    {
        return unbondingQueues[delegator].length;
    }

    /// @notice Get a specific unbonding request.
    /// @param delegator The delegator address.
    /// @param index The request index.
    /// @return request The unbonding request.
    function getUnbondingRequest(address delegator, uint256 index)
        external
        view
        returns (UnbondingRequest memory request)
    {
        return unbondingQueues[delegator][index];
    }

    // ──────────────────────────────────────────────
    //  External: Admin
    // ──────────────────────────────────────────────

    /// @notice Update the BunkerStaking contract reference.
    /// @param _stakingContract New staking contract address.
    function setStakingContract(address _stakingContract) external onlyOwner {
        address old = stakingContract;
        stakingContract = _stakingContract;
        emit StakingContractUpdated(old, _stakingContract);
    }

    /// @notice Adjust the delegation unbonding period.
    /// @param newPeriod New period in seconds (1 day min, 30 days max).
    function setUnbondingPeriod(uint256 newPeriod) external onlyOwner {
        if (newPeriod < 1 minutes || newPeriod > 30 days) revert InvalidUnbondingPeriod();
        unbondingPeriod = newPeriod;
        emit UnbondingPeriodUpdated(newPeriod);
    }

    /// @notice Adjust the maximum reward cut cap for providers.
    /// @param newMax New maximum in basis points (must be > 0 and <= BPS_DENOMINATOR).
    function setMaxRewardCutBps(uint16 newMax) external onlyOwner {
        if (newMax == 0 || newMax > uint16(BPS_DENOMINATOR)) revert InvalidRewardCutCap();
        maxRewardCutBps = newMax;
        emit MaxRewardCutUpdated(newMax);
    }

    /// @notice Pause delegation operations (emergency).
    function pause() external onlyOwner {
        _pause();
    }

    /// @notice Unpause delegation operations.
    function unpause() external onlyOwner {
        _unpause();
    }
}
