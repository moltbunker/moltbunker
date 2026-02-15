// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/Ownable2Step.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "./interfaces/IBurnable.sol";

/// @dev Minimal interface for BunkerStaking active-provider check.
interface IBunkerStaking {
    function isActiveProvider(address provider) external view returns (bool);
}

/// @title BunkerEscrow
/// @author Moltbunker
/// @notice Manages per-deployment payment escrow with protocol fees, progressive
///         release, multi-provider splitting, and dispute resolution.
/// @dev Deployed on Base L2. Uses OpenZeppelin v5. Protocol fee: 5% default
///      (80% burn, 20% treasury). All token amounts in wei (18 decimals).
contract BunkerEscrow is Ownable2Step, AccessControl, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;

    // ──────────────────────────────────────────────
    //  Constants
    // ──────────────────────────────────────────────

    string public constant VERSION = "1.2.0";

    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

    uint256 public constant PROVIDERS_PER_RESERVATION = 3;

    uint256 public constant MAX_PROTOCOL_FEE_BPS = 2000;

    uint256 public feeBurnBps = 8000;

    uint256 public feeTreasuryBps = 2000;

    uint256 public constant BPS_DENOMINATOR = 10000;

    // ──────────────────────────────────────────────
    //  Types
    // ──────────────────────────────────────────────

    enum Status { None, Created, Active, Completed, Refunded, Disputed }

    struct Reservation {
        address requester;
        uint128 totalAmount;
        uint128 releasedAmount;
        uint48 duration;
        uint48 startTime;
        Status status;
        address[3] providers;
    }

    // ──────────────────────────────────────────────
    //  State
    // ──────────────────────────────────────────────

    IERC20 public immutable token;

    address public treasury;

    uint256 public protocolFeeBps;

    uint256 public nextReservationId;

    uint256 public totalBurned;

    uint256 public totalTreasuryFees;

    /// @notice Low-balance warning threshold in basis points (default 2000 = 20%).
    uint16 public lowBalanceThresholdBps = 2000;

    /// @notice Address of the BunkerStaking contract for provider verification.
    address public stakingContract;

    mapping(uint256 => Reservation) internal _reservations;

    /// @notice Protocol fee snapshot per reservation (set at creation time).
    mapping(uint256 => uint16) public reservationFeeBps;

    // ──────────────────────────────────────────────
    //  Errors
    // ──────────────────────────────────────────────

    error ZeroAmount();
    error ZeroAddress();
    error ZeroDuration();
    error AmountOverflow(uint256 amount);
    error DurationOverflow(uint256 duration);
    error InvalidReservation(uint256 reservationId);
    error InvalidStatus(uint256 reservationId, Status expected, Status actual);
    error SettledDurationExceedsTotal(uint256 settled, uint256 total);
    error NothingToRelease(uint256 reservationId);
    error ProtocolFeeTooHigh(uint256 requested, uint256 maximum);
    error NotRequesterOrOperator(address caller, uint256 reservationId);
    error InvalidDisputeAmounts(uint256 requesterAmt, uint256 providerAmt, uint256 available);
    error NotRequester(address caller, uint256 reservationId);
    error ThresholdTooHigh(uint16 requested, uint16 maximum);
    error NotProvider(address caller, uint256 reservationId);
    error ProviderNotStaked(address provider);
    error DuplicateProvider();
    error CannotDisableStakingVerification();
    error InvalidFeeSplit();

    // ──────────────────────────────────────────────
    //  Events
    // ──────────────────────────────────────────────

    event ReservationCreated(
        uint256 indexed reservationId,
        address indexed requester,
        uint256 amount,
        uint256 duration
    );

    event ProvidersSelected(
        uint256 indexed reservationId,
        address provider0,
        address provider1,
        address provider2
    );

    event PaymentReleased(
        uint256 indexed reservationId,
        uint256 grossAmount,
        uint256 netToProviders,
        uint256 protocolFee,
        uint256 burnedAmount,
        uint256 treasuryAmount
    );

    event ReservationFinalized(uint256 indexed reservationId);

    event Refunded(
        uint256 indexed reservationId,
        address indexed requester,
        uint256 refundAmount
    );

    event DisputeSettled(
        uint256 indexed reservationId,
        uint256 requesterAmount,
        uint256 providerAmount
    );

    event ProtocolFeeUpdated(uint256 oldFeeBps, uint256 newFeeBps);

    event TreasuryUpdated(address indexed oldTreasury, address indexed newTreasury);

    event DepositIncreased(
        uint256 indexed reservationId,
        address indexed requester,
        uint256 additionalAmount,
        uint256 newTotal
    );

    event LowBalance(
        uint256 indexed reservationId,
        uint256 remaining,
        uint256 threshold
    );

    event StakingContractUpdated(
        address indexed oldStaking,
        address indexed newStaking
    );

    event FeeSplitUpdated(uint16 burnBps, uint16 treasuryBps);

    // ──────────────────────────────────────────────
    //  Constructor
    // ──────────────────────────────────────────────

    constructor(
        address _token,
        address _treasury,
        address _initialOwner
    ) Ownable(_initialOwner) {
        if (_token == address(0) || _treasury == address(0) || _initialOwner == address(0))
            revert ZeroAddress();

        token = IERC20(_token);
        treasury = _treasury;
        protocolFeeBps = 500; // 5% default
        nextReservationId = 1;

        _grantRole(DEFAULT_ADMIN_ROLE, _initialOwner);
    }

    // ──────────────────────────────────────────────
    //  External: Requester Functions
    // ──────────────────────────────────────────────

    /// @notice Create a new deployment reservation and deposit tokens into escrow.
    function createReservation(uint256 amount, uint256 duration)
        external
        nonReentrant
        whenNotPaused
        returns (uint256 reservationId)
    {
        if (amount == 0) revert ZeroAmount();
        if (duration == 0) revert ZeroDuration();
        if (amount > type(uint128).max) revert AmountOverflow(amount);
        if (duration > type(uint48).max) revert DurationOverflow(duration);

        reservationId = nextReservationId;
        unchecked { ++nextReservationId; }

        _reservations[reservationId] = Reservation({
            requester: msg.sender,
            totalAmount: uint128(amount),
            releasedAmount: 0,
            duration: uint48(duration),
            startTime: 0,
            status: Status.Created,
            providers: [address(0), address(0), address(0)]
        });

        // Snapshot the protocol fee at reservation creation time (H-09).
        reservationFeeBps[reservationId] = uint16(protocolFeeBps);

        token.safeTransferFrom(msg.sender, address(this), amount);

        emit ReservationCreated(reservationId, msg.sender, amount, duration);
    }

    /// @notice Increase the escrow deposit for an existing reservation.
    /// @param reservationId The reservation to top up.
    /// @param amount Additional BUNKER tokens to deposit.
    /// @dev M-07: The top-up applies retroactively to the full elapsed period.
    ///      This is by design — the total distributed across all releases never
    ///      exceeds totalAmount, so providers simply earn a higher rate going forward.
    function increaseDeposit(uint256 reservationId, uint256 amount)
        external
        nonReentrant
        whenNotPaused
    {
        if (amount == 0) revert ZeroAmount();

        Reservation storage res = _reservations[reservationId];
        if (res.requester == address(0))
            revert InvalidReservation(reservationId);
        if (msg.sender != res.requester)
            revert NotRequester(msg.sender, reservationId);
        if (res.status != Status.Created && res.status != Status.Active)
            revert InvalidStatus(reservationId, Status.Created, res.status);

        uint256 newTotal = uint256(res.totalAmount) + amount;
        if (newTotal > type(uint128).max) revert AmountOverflow(newTotal);

        res.totalAmount = uint128(newTotal);

        token.safeTransferFrom(msg.sender, address(this), amount);

        emit DepositIncreased(reservationId, msg.sender, amount, newTotal);
    }

    // ──────────────────────────────────────────────
    //  External: Provider Functions
    // ──────────────────────────────────────────────

    /// @notice Allows a provider to claim their proportional share based on elapsed time.
    /// @param reservationId The active reservation to claim from.
    function claim(uint256 reservationId)
        external
        nonReentrant
        whenNotPaused
    {
        Reservation storage res = _reservations[reservationId];
        if (res.status != Status.Active)
            revert InvalidStatus(reservationId, Status.Active, res.status);

        // Verify caller is a provider.
        bool isProvider = false;
        for (uint256 i = 0; i < PROVIDERS_PER_RESERVATION;) {
            if (res.providers[i] == msg.sender) { isProvider = true; break; }
            unchecked { ++i; }
        }
        if (!isProvider) revert NotProvider(msg.sender, reservationId);

        // Calculate proportional release based on elapsed time.
        uint256 elapsed = block.timestamp - res.startTime;
        if (elapsed > res.duration) elapsed = res.duration;

        uint256 proportionalTotal = (uint256(res.totalAmount) * elapsed) / uint256(res.duration);
        if (proportionalTotal <= res.releasedAmount)
            revert NothingToRelease(reservationId);

        uint256 grossRelease = proportionalTotal - uint256(res.releasedAmount);
        res.releasedAmount += uint128(grossRelease);

        (uint256 fee, uint256 burnAmount, uint256 treasuryAmount) =
            _distributePayment(reservationId, res, grossRelease);

        // Check low balance.
        _checkLowBalance(reservationId, res);

        emit PaymentReleased(
            reservationId,
            grossRelease,
            grossRelease - fee,
            fee,
            burnAmount,
            treasuryAmount
        );
    }

    // ──────────────────────────────────────────────
    //  External: Operator Functions (OPERATOR_ROLE)
    // ──────────────────────────────────────────────

    /// @notice Select 3 providers for a reservation and start the deployment.
    function selectProviders(uint256 reservationId, address[3] calldata providerAddrs)
        external
        onlyRole(OPERATOR_ROLE)
        whenNotPaused
    {
        Reservation storage res = _reservations[reservationId];
        if (res.status != Status.Created)
            revert InvalidStatus(reservationId, Status.Created, res.status);

        for (uint256 i = 0; i < 3;) {
            if (providerAddrs[i] == address(0)) revert ZeroAddress();

            // Verify provider is staked if staking contract is configured.
            if (stakingContract != address(0)) {
                if (!IBunkerStaking(stakingContract).isActiveProvider(providerAddrs[i]))
                    revert ProviderNotStaked(providerAddrs[i]);
            }

            res.providers[i] = providerAddrs[i];
            unchecked { ++i; }
        }

        // C-04: Reject duplicate provider addresses.
        if (providerAddrs[0] == providerAddrs[1] ||
            providerAddrs[0] == providerAddrs[2] ||
            providerAddrs[1] == providerAddrs[2]) revert DuplicateProvider();

        res.startTime = uint48(block.timestamp);
        res.status = Status.Active;

        emit ProvidersSelected(
            reservationId,
            providerAddrs[0],
            providerAddrs[1],
            providerAddrs[2]
        );
    }

    /// @notice Release payment for a settled period. Progressive settlement.
    function releasePayment(uint256 reservationId, uint256 settledDuration)
        external
        onlyRole(OPERATOR_ROLE)
        nonReentrant
        whenNotPaused
    {
        Reservation storage res = _reservations[reservationId];
        if (res.status != Status.Active)
            revert InvalidStatus(reservationId, Status.Active, res.status);
        if (settledDuration > res.duration)
            revert SettledDurationExceedsTotal(settledDuration, res.duration);

        uint256 proportionalTotal = (uint256(res.totalAmount) * settledDuration) / uint256(res.duration);
        if (proportionalTotal <= res.releasedAmount)
            revert NothingToRelease(reservationId);

        uint256 grossRelease = proportionalTotal - uint256(res.releasedAmount);
        res.releasedAmount += uint128(grossRelease);

        (uint256 fee, uint256 burnAmount, uint256 treasuryAmount) =
            _distributePayment(reservationId, res, grossRelease);

        // Check low balance after payment distribution.
        _checkLowBalance(reservationId, res);

        emit PaymentReleased(
            reservationId,
            grossRelease,
            grossRelease - fee,
            fee,
            burnAmount,
            treasuryAmount
        );
    }

    /// @notice Finalize a completed reservation.
    function finalizeReservation(uint256 reservationId)
        external
        onlyRole(OPERATOR_ROLE)
        nonReentrant
    {
        Reservation storage res = _reservations[reservationId];
        if (res.status != Status.Active)
            revert InvalidStatus(reservationId, Status.Active, res.status);

        uint256 remaining = uint256(res.totalAmount) - uint256(res.releasedAmount);
        res.releasedAmount = res.totalAmount;
        res.status = Status.Completed;

        if (remaining > 0) {
            _distributePayment(reservationId, res, remaining);
        }

        // Check low balance (will be zero after finalization, always triggers if threshold > 0).
        _checkLowBalance(reservationId, res);

        emit ReservationFinalized(reservationId);
    }

    /// @notice Refund unused escrow on early termination.
    function refund(uint256 reservationId)
        external
        nonReentrant
    {
        Reservation storage res = _reservations[reservationId];
        if (res.status != Status.Active && res.status != Status.Created)
            revert InvalidStatus(reservationId, Status.Active, res.status);

        if (msg.sender != res.requester && !hasRole(OPERATOR_ROLE, msg.sender))
            revert NotRequesterOrOperator(msg.sender, reservationId);

        // H-08: For Active reservations, release proportional earned amount to
        // providers before refunding the remainder to the requester.
        if (res.status == Status.Active) {
            uint256 elapsed = block.timestamp - uint256(res.startTime);
            if (elapsed > uint256(res.duration)) elapsed = uint256(res.duration);
            uint256 proportionalTotal = (uint256(res.totalAmount) * elapsed) / uint256(res.duration);
            if (proportionalTotal > uint256(res.releasedAmount)) {
                uint256 earnedUnreleased = proportionalTotal - uint256(res.releasedAmount);
                _distributePayment(reservationId, res, earnedUnreleased);
                res.releasedAmount += uint128(earnedUnreleased);
            }
        }

        uint256 remaining = uint256(res.totalAmount) - uint256(res.releasedAmount);
        res.releasedAmount = res.totalAmount;
        res.status = Status.Refunded;

        if (remaining > 0) {
            token.safeTransfer(res.requester, remaining);
        }

        emit Refunded(reservationId, res.requester, remaining);
    }

    /// @notice Settle a dispute with custom split.
    function settleDispute(
        uint256 reservationId,
        uint256 requesterAmount,
        uint256 providerAmount
    )
        external
        onlyRole(OPERATOR_ROLE)
        nonReentrant
    {
        Reservation storage res = _reservations[reservationId];
        if (res.status != Status.Active)
            revert InvalidStatus(reservationId, Status.Active, res.status);

        uint256 remaining = uint256(res.totalAmount) - uint256(res.releasedAmount);
        if (requesterAmount + providerAmount != remaining)
            revert InvalidDisputeAmounts(requesterAmount, providerAmount, remaining);

        res.releasedAmount = res.totalAmount;
        res.status = Status.Disputed;

        if (requesterAmount > 0) {
            token.safeTransfer(res.requester, requesterAmount);
        }

        // M-08: Note — the protocol fee is intentionally applied to the provider
        // portion during disputes. This is by design to prevent fee circumvention.
        if (providerAmount > 0) {
            _distributePayment(reservationId, res, providerAmount);
        }

        emit DisputeSettled(reservationId, requesterAmount, providerAmount);
    }

    // ──────────────────────────────────────────────
    //  External: Views
    // ──────────────────────────────────────────────

    function getReservation(uint256 reservationId)
        external
        view
        returns (Reservation memory)
    {
        return _reservations[reservationId];
    }

    function getProviders(uint256 reservationId)
        external
        view
        returns (address[3] memory)
    {
        return _reservations[reservationId].providers;
    }

    function calculateProtocolFee(uint256 amount)
        external
        view
        returns (uint256 fee)
    {
        return (amount * protocolFeeBps) / BPS_DENOMINATOR;
    }

    // ──────────────────────────────────────────────
    //  External: Admin
    // ──────────────────────────────────────────────

    function setProtocolFee(uint256 newFeeBps) external onlyOwner {
        if (newFeeBps > MAX_PROTOCOL_FEE_BPS)
            revert ProtocolFeeTooHigh(newFeeBps, MAX_PROTOCOL_FEE_BPS);
        uint256 old = protocolFeeBps;
        protocolFeeBps = newFeeBps;
        emit ProtocolFeeUpdated(old, newFeeBps);
    }

    function setTreasury(address newTreasury) external onlyOwner {
        if (newTreasury == address(0)) revert ZeroAddress();
        address old = treasury;
        treasury = newTreasury;
        emit TreasuryUpdated(old, newTreasury);
    }

    /// @notice Update the low-balance warning threshold.
    /// @param newThresholdBps New threshold in basis points (max 5000 = 50%).
    function setLowBalanceThreshold(uint16 newThresholdBps) external onlyOwner {
        if (newThresholdBps > 5000)
            revert ThresholdTooHigh(newThresholdBps, 5000);
        lowBalanceThresholdBps = newThresholdBps;
    }

    /// @notice Set the BunkerStaking contract address for provider verification.
    /// @param _stakingContract Address of the BunkerStaking contract.
    /// @dev H-10: Once set to a non-zero address, cannot be reset to address(0).
    function setStakingContract(address _stakingContract) external onlyOwner {
        if (stakingContract != address(0) && _stakingContract == address(0))
            revert CannotDisableStakingVerification();
        address old = stakingContract;
        stakingContract = _stakingContract;
        emit StakingContractUpdated(old, _stakingContract);
    }

    /// @notice Adjust the fee distribution split between burn and treasury.
    /// @param burnBps Burn portion in basis points.
    /// @param treasuryBps Treasury portion in basis points.
    function setFeeSplit(uint16 burnBps, uint16 treasuryBps) external onlyOwner {
        if (uint256(burnBps) + uint256(treasuryBps) != BPS_DENOMINATOR) revert InvalidFeeSplit();
        feeBurnBps = burnBps;
        feeTreasuryBps = treasuryBps;
        emit FeeSplitUpdated(burnBps, treasuryBps);
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

    // ──────────────────────────────────────────────
    //  Internal
    // ──────────────────────────────────────────────

    /// @dev Distribute payment to providers after deducting protocol fee.
    /// @dev H-09: Uses snapshotted reservationFeeBps instead of mutable global protocolFeeBps.
    /// @dev M-09: Returns fee breakdown to avoid recalculation in callers.
    function _distributePayment(
        uint256 reservationId,
        Reservation storage res,
        uint256 grossAmount
    ) internal returns (uint256 fee, uint256 burnAmount, uint256 treasuryAmount) {
        fee = (grossAmount * uint256(reservationFeeBps[reservationId])) / BPS_DENOMINATOR;
        uint256 netToProviders = grossAmount - fee;

        burnAmount = (fee * feeBurnBps) / BPS_DENOMINATOR;
        treasuryAmount = fee - burnAmount;

        totalBurned += burnAmount;
        totalTreasuryFees += treasuryAmount;

        if (burnAmount > 0) {
            _burnTokens(burnAmount);
        }

        if (treasuryAmount > 0) {
            token.safeTransfer(treasury, treasuryAmount);
        }

        uint256 perProvider = netToProviders / PROVIDERS_PER_RESERVATION;
        uint256 dust = netToProviders - (perProvider * PROVIDERS_PER_RESERVATION);

        for (uint256 i = 0; i < PROVIDERS_PER_RESERVATION;) {
            uint256 amt = perProvider;
            if (i == 0) amt += dust;
            if (amt > 0) {
                token.safeTransfer(res.providers[i], amt);
            }
            unchecked { ++i; }
        }
    }

    /// @dev Check if remaining balance is below the low-balance threshold and emit event.
    /// @param reservationId The reservation ID (for the event).
    /// @param res The reservation storage reference.
    function _checkLowBalance(uint256 reservationId, Reservation storage res) internal {
        uint256 remaining = uint256(res.totalAmount) - uint256(res.releasedAmount);
        uint256 threshold = (uint256(res.totalAmount) * uint256(lowBalanceThresholdBps)) / BPS_DENOMINATOR;
        if (remaining < threshold) {
            emit LowBalance(reservationId, remaining, threshold);
        }
    }

    /// @dev Burn tokens held by this contract via ERC20Burnable.burn().
    function _burnTokens(uint256 amount) internal {
        IBurnable(address(token)).burn(amount);
    }
}
