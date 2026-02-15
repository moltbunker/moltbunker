// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/governance/TimelockController.sol";

/// @title BunkerTimelock
/// @author Moltbunker
/// @notice Timelock controller for admin parameter changes across the Moltbunker
///         protocol. Enforces a 24-hour minimum delay on sensitive operations
///         (setProtocolFee, setTreasury, setPrices, setTierMinStake, etc.) while
///         allowing immediate pause/unpause for emergencies.
/// @dev Extends OpenZeppelin's TimelockController v5. Intended to be set as the
///      owner of BunkerStaking, BunkerEscrow, and BunkerPricing. The GUARDIAN_ROLE
///      allows immediate emergency pause/unpause without going through the timelock
///      delay. All other admin calls must be scheduled, delayed, and then executed.
contract BunkerTimelock is TimelockController {

    // ──────────────────────────────────────────────
    //  Constants
    // ──────────────────────────────────────────────

    /// @notice Contract version.
    string public constant VERSION = "1.0.0";

    /// @notice Minimum delay: 24 hours. Cannot be reduced below this.
    uint256 public constant MIN_DELAY_FLOOR = 24 hours;

    /// @notice Role for guardians who can trigger emergency pause/unpause.
    bytes32 public constant GUARDIAN_ROLE = keccak256("GUARDIAN_ROLE");

    /// @dev Function selector for pause().
    bytes4 private constant PAUSE_SELECTOR = bytes4(keccak256("pause()"));

    /// @dev Function selector for unpause().
    bytes4 private constant UNPAUSE_SELECTOR = bytes4(keccak256("unpause()"));

    // ──────────────────────────────────────────────
    //  Errors
    // ──────────────────────────────────────────────

    error DelayBelowFloor(uint256 requested, uint256 floor);
    error DelayChangeBlocked();
    error ZeroAddress();
    error TargetNotContract(address target);
    error EmergencyCallFailed(address target, bytes data);
    error PauseCooldownActive(address target);

    // ──────────────────────────────────────────────
    //  Events
    // ──────────────────────────────────────────────

    /// @notice Emitted when an emergency pause or unpause is executed.
    event EmergencyAction(
        address indexed guardian,
        address indexed target,
        bytes4 indexed selector
    );

    // ──────────────────────────────────────────────
    //  State
    // ──────────────────────────────────────────────

    /// @notice Timestamp when a target was paused (for cooldown enforcement).
    mapping(address => uint256) public pausedAt;

    // ──────────────────────────────────────────────
    //  Constructor
    // ──────────────────────────────────────────────

    /// @param _minDelay    Initial timelock delay (must be >= 24 hours).
    /// @param _proposers   Addresses granted PROPOSER_ROLE (and CANCELLER_ROLE).
    /// @param _executors   Addresses granted EXECUTOR_ROLE.
    /// @param _admin       Optional admin for initial setup (should be renounced later).
    /// @param _guardian    Address granted GUARDIAN_ROLE for emergency actions.
    constructor(
        uint256 _minDelay,
        address[] memory _proposers,
        address[] memory _executors,
        address _admin,
        address _guardian
    ) TimelockController(_minDelay, _proposers, _executors, _admin) {
        if (_minDelay < MIN_DELAY_FLOOR)
            revert DelayBelowFloor(_minDelay, MIN_DELAY_FLOOR);

        /// @dev The timelock grants itself PROPOSER_ROLE to enable convenience scheduling helpers.
        /// This means executed operations can schedule further operations as a side effect.
        /// All operations still require the full delay before execution.
        _grantRole(PROPOSER_ROLE, address(this));

        // Grant GUARDIAN_ROLE. The admin can manage this role via DEFAULT_ADMIN_ROLE.
        if (_guardian != address(0)) {
            _grantRole(GUARDIAN_ROLE, _guardian);
        }
    }

    // ──────────────────────────────────────────────
    //  Override: Block delay changes (C-03)
    // ──────────────────────────────────────────────

    /// @notice Delay changes are blocked to prevent bypass of MIN_DELAY_FLOOR.
    /// @dev The initial delay is set in the constructor and cannot be changed.
    function updateDelay(uint256) external pure override {
        revert DelayChangeBlocked();
    }

    // ──────────────────────────────────────────────
    //  Emergency: Pause / Unpause (no delay)
    // ──────────────────────────────────────────────

    /// @notice Emergency pause on a target contract. No timelock delay.
    /// @dev Calls `pause()` on the target. The target must be a Pausable contract
    ///      whose owner is this timelock. Records pause timestamp for cooldown.
    /// @param target Address of the contract to pause.
    function emergencyPause(address target) external onlyRole(GUARDIAN_ROLE) {
        if (target == address(0)) revert ZeroAddress();
        if (target.code.length == 0) revert TargetNotContract(target);

        pausedAt[target] = block.timestamp;

        bytes memory data = abi.encodeWithSignature("pause()");
        (bool success,) = target.call(data);
        if (!success) revert EmergencyCallFailed(target, data);

        emit EmergencyAction(msg.sender, target, PAUSE_SELECTOR);
    }

    /// @notice Emergency unpause on a target contract. No timelock delay.
    /// @dev Calls `unpause()` on the target. The target must be a Pausable contract
    ///      whose owner is this timelock. Requires at least 1 hour since pause (M-10).
    /// @param target Address of the contract to unpause.
    function emergencyUnpause(address target) external onlyRole(GUARDIAN_ROLE) {
        if (target == address(0)) revert ZeroAddress();
        if (target.code.length == 0) revert TargetNotContract(target);
        if (block.timestamp < pausedAt[target] + 1 hours) revert PauseCooldownActive(target);

        bytes memory data = abi.encodeWithSignature("unpause()");
        (bool success,) = target.call(data);
        if (!success) revert EmergencyCallFailed(target, data);

        emit EmergencyAction(msg.sender, target, UNPAUSE_SELECTOR);
    }

    // ──────────────────────────────────────────────
    //  Convenience: Schedule Helpers
    // ──────────────────────────────────────────────

    /// @notice Schedule a setTreasury call on a target contract.
    /// @param target     The contract address (BunkerStaking or BunkerEscrow).
    /// @param newTreasury The new treasury address.
    /// @param salt       Unique salt for the operation.
    function scheduleSetTreasury(
        address target,
        address newTreasury,
        bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setTreasury(address)", newTreasury
        );
        uint256 delay = getMinDelay();
        this.schedule(target, 0, data, bytes32(0), salt, delay);
    }

    /// @notice Schedule a setProtocolFee call on BunkerEscrow.
    /// @param target    The BunkerEscrow contract address.
    /// @param newFeeBps The new protocol fee in basis points.
    /// @param salt      Unique salt for the operation.
    function scheduleSetProtocolFee(
        address target,
        uint256 newFeeBps,
        bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setProtocolFee(uint256)", newFeeBps
        );
        uint256 delay = getMinDelay();
        this.schedule(target, 0, data, bytes32(0), salt, delay);
    }

    /// @notice Schedule a setTierMinStake call on BunkerStaking.
    /// @param target   The BunkerStaking contract address.
    /// @param tier     The tier enum value (1=Starter..5=Platinum).
    /// @param minStake The new minimum stake for the tier.
    /// @param salt     Unique salt for the operation.
    function scheduleSetTierMinStake(
        address target,
        uint8 tier,
        uint256 minStake,
        bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setTierMinStake(uint8,uint256)", tier, minStake
        );
        uint256 delay = getMinDelay();
        this.schedule(target, 0, data, bytes32(0), salt, delay);
    }

    /// @notice Schedule a setTierRewardMultiplier call on BunkerStaking.
    function scheduleSetTierRewardMultiplier(
        address target, uint8 tier, uint16 multiplierBps, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setTierRewardMultiplier(uint8,uint16)", tier, multiplierBps
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    /// @notice Schedule a setMaxTierMultiplierBps call on BunkerStaking.
    function scheduleSetMaxTierMultiplierBps(
        address target, uint16 newMax, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setMaxTierMultiplierBps(uint16)", newMax
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    /// @notice Schedule a setUnbondingPeriod call on BunkerStaking or BunkerDelegation.
    function scheduleSetUnbondingPeriod(
        address target, uint256 newPeriod, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setUnbondingPeriod(uint256)", newPeriod
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    /// @notice Schedule a setSlashFeeSplit call on BunkerStaking.
    function scheduleSetSlashFeeSplit(
        address target, uint16 burnBps, uint16 treasuryBps, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setSlashFeeSplit(uint16,uint16)", burnBps, treasuryBps
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    /// @notice Schedule a setAppealWindow call on BunkerStaking.
    function scheduleSetAppealWindow(
        address target, uint256 newWindow, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setAppealWindow(uint256)", newWindow
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    /// @notice Schedule a setFeeSplit call on BunkerEscrow.
    function scheduleSetFeeSplit(
        address target, uint16 burnBps, uint16 treasuryBps, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setFeeSplit(uint16,uint16)", burnBps, treasuryBps
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    /// @notice Schedule a setMaxRewardCutBps call on BunkerDelegation.
    function scheduleSetMaxRewardCutBps(
        address target, uint16 newMax, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setMaxRewardCutBps(uint16)", newMax
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    /// @notice Schedule a setReinstatementCooldown call on BunkerVerification.
    function scheduleSetReinstatementCooldown(
        address target, uint256 newCooldown, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setReinstatementCooldown(uint256)", newCooldown
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    /// @notice Schedule a setMaxMultiplierBps call on BunkerPricing.
    function scheduleSetMaxMultiplierBps(
        address target, uint256 newMax, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setMaxMultiplierBps(uint256)", newMax
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    /// @notice Schedule a setStaleThresholdBounds call on BunkerPricing.
    function scheduleSetStaleThresholdBounds(
        address target, uint256 newMin, uint256 newMax, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setStaleThresholdBounds(uint256,uint256)", newMin, newMax
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    /// @notice Schedule a setSlashingEnabled call on BunkerStaking.
    function scheduleSetSlashingEnabled(
        address target, bool enabled, bytes32 salt
    ) external onlyRole(PROPOSER_ROLE) {
        bytes memory data = abi.encodeWithSignature(
            "setSlashingEnabled(bool)", enabled
        );
        this.schedule(target, 0, data, bytes32(0), salt, getMinDelay());
    }

    // ──────────────────────────────────────────────
    //  Views
    // ──────────────────────────────────────────────

    /// @notice Get the operation hash for a scheduled call (convenience view).
    /// @dev Delegates to the inherited hashOperation.
    function getOperationHash(
        address target,
        uint256 value,
        bytes calldata data,
        bytes32 predecessor,
        bytes32 salt
    ) external pure returns (bytes32) {
        return hashOperation(target, value, data, predecessor, salt);
    }
}
