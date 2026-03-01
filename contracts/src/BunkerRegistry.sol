// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/Ownable2Step.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "./interfaces/IBurnable.sol";

/// @notice Interface for querying staking tiers from BunkerStaking.
interface IBunkerStakingTier {
    function getTier(address provider) external view returns (uint8);
}

/// @title BunkerRegistry
/// @author Moltbunker
/// @notice On-chain registry for vanity subdomain names (<name>.moltbunker.dev).
/// @dev v2.0.0 — Adds expiration, renewal, premium pricing, staking discounts,
///      referral, bulk operations, name reservation, squatting protection,
///      reverse resolution, and metadata storage.
contract BunkerRegistry is Ownable2Step, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;

    // ─── Constants ──────────────────────────────────────────────────────

    /// @notice BUNKER token contract.
    IERC20 public immutable bunkerToken;

    /// @notice Burn share in basis points (80%).
    uint256 public constant BURN_BPS = 8000;

    /// @notice BPS denominator.
    uint256 public constant BPS_DENOMINATOR = 10000;

    /// @notice Maximum number of names a single address can own.
    uint256 public constant MAX_NAMES_PER_OWNER = 100;

    /// @notice Minimum registration fee (1,000 BUNKER).
    uint256 public constant MIN_REGISTRATION_FEE = 1000 * 1e18;

    /// @notice Maximum names per bulk operation.
    uint256 public constant MAX_BULK_SIZE = 20;

    /// @notice Premium multiplier for 1-char names (100x).
    uint256 public constant PREMIUM_1_CHAR_MULTIPLIER = 100;

    /// @notice Premium multiplier for 2-char names (50x).
    uint256 public constant PREMIUM_2_CHAR_MULTIPLIER = 50;

    /// @notice Premium multiplier for 3-char names (10x).
    uint256 public constant PREMIUM_3_CHAR_MULTIPLIER = 10;

    /// @notice Premium multiplier for 4-char names (5x).
    uint256 public constant PREMIUM_4_CHAR_MULTIPLIER = 5;

    /// @notice Maximum description length for metadata.
    uint256 public constant MAX_DESCRIPTION_LENGTH = 160;

    /// @notice Maximum avatar URL length for metadata.
    uint256 public constant MAX_AVATAR_URL_LENGTH = 256;

    /// @notice Contract version.
    string public constant VERSION = "2.0.0";

    // ─── State ──────────────────────────────────────────────────────────

    /// @notice Treasury address receiving 20% of registration fees.
    address public treasury;

    /// @notice Registration fee in BUNKER tokens.
    uint256 public registrationFee;

    /// @notice Staking contract for tier discount lookups.
    IBunkerStakingTier public stakingContract;

    /// @notice Name lifetime in seconds (default: 365 days).
    uint256 public expirationPeriod = 365 days;

    /// @notice Post-expiry renewal window (default: 30 days).
    uint256 public gracePeriod = 30 days;

    /// @notice Reservation hold time (default: 48 hours).
    uint256 public reservationPeriod = 48 hours;

    /// @notice Fee for updateDeployment and setMetadata (default: 10K BUNKER).
    uint256 public changeFee = 10_000 * 1e18;

    /// @notice Referral discount in BPS (default: 1000 = 10%).
    uint256 public referralDiscountBps = 1000;

    /// @notice Referral reward in BPS of original fee (default: 500 = 5%).
    uint256 public referralRewardBps = 500;

    /// @notice Grace period to set a deployment after registration (default: 7 days).
    uint256 public squattingGracePeriod = 7 days;

    /// @notice Whether 3-character names can be registered (default: false, admin-gated).
    bool public shortNamesEnabled;

    /// @notice Reserved name hashes that cannot be registered by users.
    mapping(bytes32 => bool) public reservedNames;

    /// @notice Staking tier discount BPS: None=0, Starter=0, Bronze=500, Silver=1000, Gold=1500, Platinum=2000.
    uint256[6] internal _tierDiscountBps = [0, 0, 500, 1000, 1500, 2000];

    struct SubdomainRecord {
        address owner;
        bytes32 deploymentID;
        uint48  registeredAt;
        uint48  expiresAt;
        uint48  reservedUntil;   // 0 = not reserved
        address referrer;
    }

    struct Metadata {
        string description;  // max 160 chars
        string avatarURL;    // max 256 chars
    }

    /// @notice Name hash → subdomain record.
    mapping(bytes32 => SubdomainRecord) public subdomains;

    /// @notice Name hash → original name string (for enumeration).
    mapping(bytes32 => string) public nameOf;

    /// @notice Owner → list of name hashes they own.
    mapping(address => bytes32[]) private _ownedNames;

    /// @notice Name hash → metadata.
    mapping(bytes32 => Metadata) public metadata;

    /// @notice Deployment ID → primary name hash (reverse resolution).
    mapping(bytes32 => bytes32) public primaryName;

    /// @notice Name hash → original name for reverse resolution lookups.
    // (reuses nameOf mapping — no separate storage needed)

    // ─── Events ─────────────────────────────────────────────────────────

    event SubdomainRegistered(
        string indexed nameIndexed,
        string name,
        address indexed owner,
        bytes32 deploymentID,
        uint256 fee
    );

    event SubdomainReleased(
        string indexed nameIndexed,
        string name,
        address indexed owner
    );

    event SubdomainTransferred(
        string indexed nameIndexed,
        string name,
        address indexed from,
        address indexed to
    );

    event SubdomainUpdated(
        string indexed nameIndexed,
        string name,
        bytes32 oldDeploymentID,
        bytes32 newDeploymentID
    );

    event SubdomainRenewed(
        string indexed nameIndexed,
        string name,
        address indexed owner,
        uint48 newExpiry,
        uint256 fee
    );

    event SubdomainReserved(
        string indexed nameIndexed,
        string name,
        address indexed reserver,
        uint48 reservedUntil,
        uint256 fee
    );

    event ReservationClaimed(
        string indexed nameIndexed,
        string name,
        address indexed owner,
        bytes32 deploymentID
    );

    event ReservationCancelled(
        string indexed nameIndexed,
        string name,
        address indexed owner
    );

    event SquattedNameReclaimed(
        string indexed nameIndexed,
        string name,
        address indexed reclaimer
    );

    event MetadataUpdated(
        string indexed nameIndexed,
        string name,
        address indexed owner
    );

    event PrimaryNameSet(
        bytes32 indexed deploymentID,
        string name,
        address indexed owner
    );

    event RegistrationFeeUpdated(uint256 oldFee, uint256 newFee);
    event TreasuryUpdated(address oldTreasury, address newTreasury);
    event ChangeFeeUpdated(uint256 oldFee, uint256 newFee);
    event ExpirationPeriodUpdated(uint256 oldPeriod, uint256 newPeriod);
    event GracePeriodUpdated(uint256 oldPeriod, uint256 newPeriod);
    event ReservationPeriodUpdated(uint256 oldPeriod, uint256 newPeriod);
    event SquattingGracePeriodUpdated(uint256 oldPeriod, uint256 newPeriod);
    event StakingContractUpdated(address oldAddr, address newAddr);
    event ReferralDiscountUpdated(uint256 oldBps, uint256 newBps);
    event ReferralRewardUpdated(uint256 oldBps, uint256 newBps);
    event ShortNamesEnabledUpdated(bool enabled);
    event ReservedNameUpdated(string name, bool reserved);

    // ─── Errors ─────────────────────────────────────────────────────────

    error NameAlreadyRegistered(string name);
    error NameNotRegistered(string name);
    error NotNameOwner(string name, address caller);
    error InvalidName(string name);
    error InvalidAddress();
    error InvalidDeploymentID();
    error FeeTransferFailed();
    error TooManyNames(address owner, uint256 max);
    error FeeBelowMinimum(uint256 fee, uint256 minimum);
    error CannotTransferToSelf();
    error NameExpired(string name);
    error NameInGracePeriod(string name);
    error NameNotExpired(string name);
    error NameReserved(string name);
    error ReservationExpired(string name);
    error NotReservationOwner(string name, address caller);
    error InvalidReferrer(address referrer);
    error MetadataDescriptionTooLong(uint256 length, uint256 max);
    error MetadataAvatarURLTooLong(uint256 length, uint256 max);
    error ArrayLengthMismatch();
    error ArrayTooLarge(uint256 length, uint256 max);
    error NameNotSquatted(string name);
    error DeploymentNotOwned(bytes32 deploymentID);
    error InvalidPeriod();
    error ShortNamesDisabled();

    // ─── Constructor ────────────────────────────────────────────────────

    /// @param _token BUNKER token address.
    /// @param _treasury Treasury address for fee collection.
    /// @param _registrationFee Initial registration fee in BUNKER (with 18 decimals).
    /// @param _owner Contract owner (typically Timelock).
    /// @param _stakingContract Staking contract for tier discount lookups.
    constructor(
        address _token,
        address _treasury,
        uint256 _registrationFee,
        address _owner,
        address _stakingContract
    ) Ownable(_owner) {
        if (_token == address(0) || _treasury == address(0)) revert InvalidAddress();
        if (_registrationFee > 0 && _registrationFee < MIN_REGISTRATION_FEE) {
            revert FeeBelowMinimum(_registrationFee, MIN_REGISTRATION_FEE);
        }
        bunkerToken = IERC20(_token);
        treasury = _treasury;
        registrationFee = _registrationFee;
        // staking contract is optional (can be address(0) if no discount needed)
        stakingContract = IBunkerStakingTier(_stakingContract);
    }

    // ─── Core Operations ────────────────────────────────────────────────

    /// @notice Register a vanity subdomain name.
    /// @param name The subdomain name (3-32 chars, lowercase alphanumeric + hyphens).
    /// @param deploymentID The deployment to point this name to.
    function register(string calldata name, bytes32 deploymentID)
        external
        nonReentrant
        whenNotPaused
    {
        _registerInternal(name, deploymentID, address(0));
    }

    /// @notice Register with a referral for a 10% discount.
    /// @param name The subdomain name.
    /// @param deploymentID The deployment to point this name to.
    /// @param referrer Address of an existing user who referred you.
    function registerWithReferral(
        string calldata name,
        bytes32 deploymentID,
        address referrer
    )
        external
        nonReentrant
        whenNotPaused
    {
        if (referrer == address(0)) revert InvalidReferrer(referrer);
        if (referrer == msg.sender) revert InvalidReferrer(referrer);
        _registerInternal(name, deploymentID, referrer);
    }

    /// @notice Register multiple names in a single transaction.
    /// @param names Array of subdomain names.
    /// @param deploymentIDs Array of deployment IDs.
    function bulkRegister(
        string[] calldata names,
        bytes32[] calldata deploymentIDs
    )
        external
        nonReentrant
        whenNotPaused
    {
        if (names.length != deploymentIDs.length) revert ArrayLengthMismatch();
        if (names.length > MAX_BULK_SIZE) revert ArrayTooLarge(names.length, MAX_BULK_SIZE);

        for (uint256 i = 0; i < names.length; i++) {
            _registerInternal(names[i], deploymentIDs[i], address(0));
        }
    }

    /// @notice Renew a name, extending its expiration.
    /// @param name The subdomain name to renew.
    function renew(string calldata name)
        external
        nonReentrant
        whenNotPaused
    {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotRegistered(name);

        // Only the owner can renew. If in grace period, also only the owner.
        if (record.owner != msg.sender) revert NotNameOwner(name, msg.sender);

        // If fully expired (past grace period), must re-register instead
        if (_isFullyExpired(nameHash)) revert NameExpired(name);

        uint256 price = _calculatePrice(name);
        price = _applyStakingDiscount(price, msg.sender);
        _collectFee(msg.sender, price, address(0), 0);

        // Extend from current expiry, or from now if in grace period
        uint48 baseTime = record.expiresAt;
        if (block.timestamp > baseTime) {
            baseTime = uint48(block.timestamp);
        }
        record.expiresAt = baseTime + uint48(expirationPeriod);

        emit SubdomainRenewed(name, name, msg.sender, record.expiresAt, price);
    }

    /// @notice Renew multiple names in a single transaction.
    /// @param names Array of subdomain names to renew.
    function bulkRenew(string[] calldata names)
        external
        nonReentrant
        whenNotPaused
    {
        if (names.length > MAX_BULK_SIZE) revert ArrayTooLarge(names.length, MAX_BULK_SIZE);

        for (uint256 i = 0; i < names.length; i++) {
            bytes32 nameHash = keccak256(abi.encodePacked(names[i]));
            SubdomainRecord storage record = subdomains[nameHash];
            if (record.owner == address(0)) revert NameNotRegistered(names[i]);
            if (record.owner != msg.sender) revert NotNameOwner(names[i], msg.sender);
            if (_isFullyExpired(nameHash)) revert NameExpired(names[i]);

            uint256 price = _calculatePrice(names[i]);
            price = _applyStakingDiscount(price, msg.sender);
            _collectFee(msg.sender, price, address(0), 0);

            uint48 baseTime = record.expiresAt;
            if (baseTime == 0 || block.timestamp > baseTime) {
                baseTime = uint48(block.timestamp);
            }
            record.expiresAt = baseTime + uint48(expirationPeriod);

            emit SubdomainRenewed(names[i], names[i], msg.sender, record.expiresAt, price);
        }
    }

    /// @notice Reserve a name before deploying (holds for reservationPeriod).
    /// @param name The subdomain name to reserve.
    function reserve(string calldata name)
        external
        nonReentrant
        whenNotPaused
    {
        _validateName(name);
        if (bytes(name).length <= 3 && !shortNamesEnabled) revert ShortNamesDisabled();
        if (reservedNames[keccak256(abi.encodePacked(name))]) revert NameReserved(name);
        if (_ownedNames[msg.sender].length >= MAX_NAMES_PER_OWNER) {
            revert TooManyNames(msg.sender, MAX_NAMES_PER_OWNER);
        }

        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];

        // Allow if: unregistered, or fully expired (past grace), or expired reservation
        if (record.owner != address(0)) {
            if (!_isFullyExpired(nameHash)) {
                // Check if it's an expired reservation (no deployment set, reservation expired)
                if (record.reservedUntil == 0 || block.timestamp <= record.reservedUntil) {
                    revert NameAlreadyRegistered(name);
                }
                // Expired reservation — clean up the old record
                _removeOwnedName(record.owner, nameHash);
            } else {
                // Fully expired — clean up
                _removeOwnedName(record.owner, nameHash);
            }
        }

        uint256 price = _calculatePrice(name);
        price = _applyStakingDiscount(price, msg.sender);
        _collectFee(msg.sender, price, address(0), 0);

        uint48 resUntil = uint48(block.timestamp + reservationPeriod);
        subdomains[nameHash] = SubdomainRecord({
            owner: msg.sender,
            deploymentID: bytes32(0),
            registeredAt: uint48(block.timestamp),
            expiresAt: uint48(block.timestamp + expirationPeriod),
            reservedUntil: resUntil,
            referrer: address(0)
        });
        nameOf[nameHash] = name;
        _ownedNames[msg.sender].push(nameHash);

        emit SubdomainReserved(name, name, msg.sender, resUntil, price);
    }

    /// @notice Claim a reservation by setting the deployment ID.
    /// @param name The reserved subdomain name.
    /// @param deploymentID The deployment to point this name to.
    function claimReservation(string calldata name, bytes32 deploymentID)
        external
        nonReentrant
        whenNotPaused
    {
        if (deploymentID == bytes32(0)) revert InvalidDeploymentID();

        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotRegistered(name);
        if (record.owner != msg.sender) revert NotReservationOwner(name, msg.sender);
        if (record.reservedUntil == 0) revert NameNotRegistered(name); // not a reservation
        if (block.timestamp > record.reservedUntil) revert ReservationExpired(name);

        record.deploymentID = deploymentID;
        record.reservedUntil = 0; // clear reservation flag

        emit ReservationClaimed(name, name, msg.sender, deploymentID);
    }

    /// @notice Cancel a reservation. No refund.
    /// @param name The reserved subdomain name.
    function cancelReservation(string calldata name)
        external
        nonReentrant
        whenNotPaused
    {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotRegistered(name);
        if (record.owner != msg.sender) revert NotReservationOwner(name, msg.sender);
        if (record.reservedUntil == 0) revert NameNotRegistered(name); // not a reservation

        address recordOwner = record.owner;
        delete subdomains[nameHash];
        delete nameOf[nameHash];
        _removeOwnedName(recordOwner, nameHash);

        emit ReservationCancelled(name, name, recordOwner);
    }

    /// @notice Release a subdomain name. No refund.
    /// @param name The subdomain name to release.
    function release(string calldata name) external nonReentrant whenNotPaused {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotRegistered(name);
        if (record.owner != msg.sender) revert NotNameOwner(name, msg.sender);

        address recordOwner = record.owner;
        // Clear primary name if this name was the primary for its deployment
        if (record.deploymentID != bytes32(0) && primaryName[record.deploymentID] == nameHash) {
            delete primaryName[record.deploymentID];
        }
        delete subdomains[nameHash];
        delete nameOf[nameHash];
        delete metadata[nameHash];
        _removeOwnedName(recordOwner, nameHash);

        emit SubdomainReleased(name, name, recordOwner);
    }

    /// @notice Transfer a subdomain to a new owner.
    /// @param name The subdomain name to transfer.
    /// @param newOwner The new owner address.
    function transfer(string calldata name, address newOwner)
        external
        nonReentrant
        whenNotPaused
    {
        if (newOwner == address(0)) revert InvalidAddress();
        if (newOwner == msg.sender) revert CannotTransferToSelf();
        if (_ownedNames[newOwner].length >= MAX_NAMES_PER_OWNER) {
            revert TooManyNames(newOwner, MAX_NAMES_PER_OWNER);
        }

        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotRegistered(name);
        if (record.owner != msg.sender) revert NotNameOwner(name, msg.sender);

        // Cannot transfer expired names
        if (_isExpired(nameHash)) revert NameExpired(name);

        address oldOwner = record.owner;
        record.owner = newOwner;
        _removeOwnedName(oldOwner, nameHash);
        _ownedNames[newOwner].push(nameHash);

        emit SubdomainTransferred(name, name, oldOwner, newOwner);
    }

    /// @notice Update the deployment ID a name points to. Charges changeFee.
    /// @param name The subdomain name.
    /// @param newDeploymentID The new deployment ID.
    function updateDeployment(string calldata name, bytes32 newDeploymentID)
        external
        nonReentrant
        whenNotPaused
    {
        if (newDeploymentID == bytes32(0)) revert InvalidDeploymentID();

        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotRegistered(name);
        if (record.owner != msg.sender) revert NotNameOwner(name, msg.sender);
        if (_isExpired(nameHash)) revert NameExpired(name);

        // Charge change fee
        if (changeFee > 0) {
            _collectFee(msg.sender, changeFee, address(0), 0);
        }

        // Clear old primary name mapping
        bytes32 oldID = record.deploymentID;
        if (oldID != bytes32(0) && primaryName[oldID] == nameHash) {
            delete primaryName[oldID];
        }

        record.deploymentID = newDeploymentID;

        emit SubdomainUpdated(name, name, oldID, newDeploymentID);
    }

    /// @notice Set metadata (description + avatar URL) for a name. Charges changeFee.
    /// @param name The subdomain name.
    /// @param description Description string (max 160 chars).
    /// @param avatarURL Avatar URL string (max 256 chars).
    function setMetadata(
        string calldata name,
        string calldata description,
        string calldata avatarURL
    )
        external
        nonReentrant
        whenNotPaused
    {
        if (bytes(description).length > MAX_DESCRIPTION_LENGTH) {
            revert MetadataDescriptionTooLong(bytes(description).length, MAX_DESCRIPTION_LENGTH);
        }
        if (bytes(avatarURL).length > MAX_AVATAR_URL_LENGTH) {
            revert MetadataAvatarURLTooLong(bytes(avatarURL).length, MAX_AVATAR_URL_LENGTH);
        }

        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotRegistered(name);
        if (record.owner != msg.sender) revert NotNameOwner(name, msg.sender);
        if (_isExpired(nameHash)) revert NameExpired(name);

        // Charge change fee
        if (changeFee > 0) {
            _collectFee(msg.sender, changeFee, address(0), 0);
        }

        metadata[nameHash] = Metadata({
            description: description,
            avatarURL: avatarURL
        });

        emit MetadataUpdated(name, name, msg.sender);
    }

    /// @notice Reclaim a squatted name (no deployment set past grace period, or expired reservation).
    /// @param name The subdomain name to reclaim.
    function reclaimSquatted(string calldata name)
        external
        nonReentrant
        whenNotPaused
    {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotSquatted(name);

        bool isSquatted = false;

        // Case 1: Reservation expired without being claimed
        if (record.reservedUntil > 0 && block.timestamp > record.reservedUntil) {
            isSquatted = true;
        }
        // Case 2: No deployment set and past squatting grace period
        else if (
            record.deploymentID == bytes32(0) &&
            record.reservedUntil == 0 &&
            block.timestamp > uint256(record.registeredAt) + squattingGracePeriod
        ) {
            isSquatted = true;
        }

        if (!isSquatted) revert NameNotSquatted(name);

        address oldOwner = record.owner;
        delete subdomains[nameHash];
        delete nameOf[nameHash];
        delete metadata[nameHash];
        _removeOwnedName(oldOwner, nameHash);

        emit SquattedNameReclaimed(name, name, msg.sender);
    }

    /// @notice Set the primary name for a deployment (for reverse resolution).
    /// @param name The subdomain name to set as primary.
    function setPrimaryName(string calldata name)
        external
        nonReentrant
        whenNotPaused
    {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotRegistered(name);
        if (record.owner != msg.sender) revert NotNameOwner(name, msg.sender);
        if (_isExpired(nameHash)) revert NameExpired(name);
        if (record.deploymentID == bytes32(0)) revert InvalidDeploymentID();

        primaryName[record.deploymentID] = nameHash;

        emit PrimaryNameSet(record.deploymentID, name, msg.sender);
    }

    // ─── View Functions ─────────────────────────────────────────────────

    /// @notice Resolve a subdomain name.
    /// @return owner The name owner.
    /// @return deploymentID The deployment ID it points to.
    /// @return registeredAt When the name was registered.
    function resolve(string calldata name)
        external
        view
        returns (address owner, bytes32 deploymentID, uint256 registeredAt)
    {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        return (record.owner, record.deploymentID, uint256(record.registeredAt));
    }

    /// @notice Check if a name is available for registration.
    function isAvailable(string calldata name) external view returns (bool) {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) return true;
        // Available if fully expired (past grace period)
        if (_isFullyExpired(nameHash)) return true;
        // Available if reservation expired
        if (record.reservedUntil > 0 && block.timestamp > record.reservedUntil) return true;
        return false;
    }

    /// @notice Check if a name is expired.
    function isExpired(string calldata name) external view returns (bool) {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        return _isExpired(nameHash);
    }

    /// @notice Check if a name is in its grace period (expired but owner can still renew).
    function isInGracePeriod(string calldata name) external view returns (bool) {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        return _isInGracePeriod(nameHash);
    }

    /// @notice Reverse resolve a deployment ID to its primary name.
    /// @param deploymentID The deployment ID to look up.
    /// @return name The primary name string, empty if none set.
    function reverseResolve(bytes32 deploymentID) external view returns (string memory name) {
        bytes32 nameHash = primaryName[deploymentID];
        if (nameHash == bytes32(0)) return "";
        return nameOf[nameHash];
    }

    /// @notice Get the number of names owned by an address.
    function nameCount(address owner) external view returns (uint256) {
        return _ownedNames[owner].length;
    }

    /// @notice Get the name hash at a given index for an owner.
    function ownedNameAt(address owner, uint256 index) external view returns (bytes32) {
        return _ownedNames[owner][index];
    }

    /// @notice Calculate the registration price for a name (with premium + staking discount).
    /// @param name The subdomain name.
    /// @param user The user address (for staking discount).
    /// @return price The final price after discounts.
    function calculatePrice(string calldata name, address user) external view returns (uint256 price) {
        price = _calculatePrice(name);
        price = _applyStakingDiscount(price, user);
    }

    // ─── Admin Functions ────────────────────────────────────────────────

    /// @notice Update the registration fee. Owner-only.
    function setRegistrationFee(uint256 newFee) external onlyOwner {
        if (newFee > 0 && newFee < MIN_REGISTRATION_FEE) {
            revert FeeBelowMinimum(newFee, MIN_REGISTRATION_FEE);
        }
        uint256 oldFee = registrationFee;
        registrationFee = newFee;
        emit RegistrationFeeUpdated(oldFee, newFee);
    }

    /// @notice Update the treasury address. Owner-only.
    function setTreasury(address newTreasury) external onlyOwner {
        if (newTreasury == address(0)) revert InvalidAddress();
        address oldTreasury = treasury;
        treasury = newTreasury;
        emit TreasuryUpdated(oldTreasury, newTreasury);
    }

    /// @notice Update the change fee. Owner-only.
    function setChangeFee(uint256 fee) external onlyOwner {
        if (fee > 0 && fee < MIN_REGISTRATION_FEE) {
            revert FeeBelowMinimum(fee, MIN_REGISTRATION_FEE);
        }
        uint256 oldFee = changeFee;
        changeFee = fee;
        emit ChangeFeeUpdated(oldFee, fee);
    }

    /// @notice Update the expiration period. Owner-only.
    function setExpirationPeriod(uint256 period) external onlyOwner {
        if (period < 30 days || period > 3650 days) revert InvalidPeriod();
        uint256 oldPeriod = expirationPeriod;
        expirationPeriod = period;
        emit ExpirationPeriodUpdated(oldPeriod, period);
    }

    /// @notice Update the grace period. Owner-only.
    function setGracePeriod(uint256 period) external onlyOwner {
        if (period < 7 days || period > 90 days) revert InvalidPeriod();
        uint256 oldPeriod = gracePeriod;
        gracePeriod = period;
        emit GracePeriodUpdated(oldPeriod, period);
    }

    /// @notice Update the reservation period. Owner-only.
    function setReservationPeriod(uint256 period) external onlyOwner {
        if (period < 1 hours || period > 7 days) revert InvalidPeriod();
        uint256 oldPeriod = reservationPeriod;
        reservationPeriod = period;
        emit ReservationPeriodUpdated(oldPeriod, period);
    }

    /// @notice Update the squatting grace period. Owner-only.
    function setSquattingGracePeriod(uint256 period) external onlyOwner {
        if (period < 1 days || period > 30 days) revert InvalidPeriod();
        uint256 oldPeriod = squattingGracePeriod;
        squattingGracePeriod = period;
        emit SquattingGracePeriodUpdated(oldPeriod, period);
    }

    /// @notice Update the referral discount BPS. Owner-only.
    function setReferralDiscountBps(uint256 bps) external onlyOwner {
        if (bps > 2000) revert InvalidPeriod(); // max 20%
        uint256 oldBps = referralDiscountBps;
        referralDiscountBps = bps;
        emit ReferralDiscountUpdated(oldBps, bps);
    }

    /// @notice Update the referral reward BPS. Owner-only.
    function setReferralRewardBps(uint256 bps) external onlyOwner {
        if (bps > 1000) revert InvalidPeriod(); // max 10%
        uint256 oldBps = referralRewardBps;
        referralRewardBps = bps;
        emit ReferralRewardUpdated(oldBps, bps);
    }

    /// @notice Update the staking contract address. Owner-only.
    function setStakingContract(address addr) external onlyOwner {
        address oldAddr = address(stakingContract);
        stakingContract = IBunkerStakingTier(addr);
        emit StakingContractUpdated(oldAddr, addr);
    }

    /// @notice Enable or disable 3-character name registration. Owner-only.
    function setShortNamesEnabled(bool enabled) external onlyOwner {
        shortNamesEnabled = enabled;
        emit ShortNamesEnabledUpdated(enabled);
    }

    /// @notice Mark a name as reserved or unreserved. Owner-only.
    /// @param name The subdomain name to reserve/unreserve.
    /// @param reserved Whether the name should be reserved.
    function setReservedName(string calldata name, bool reserved) external onlyOwner {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        reservedNames[nameHash] = reserved;
        emit ReservedNameUpdated(name, reserved);
    }

    /// @notice Batch-reserve multiple names. Owner-only.
    /// @param names Array of subdomain names to reserve.
    function batchReserveNames(string[] calldata names) external onlyOwner {
        for (uint256 i = 0; i < names.length; i++) {
            bytes32 nameHash = keccak256(abi.encodePacked(names[i]));
            reservedNames[nameHash] = true;
            emit ReservedNameUpdated(names[i], true);
        }
    }

    /// @notice Pause the contract. Owner-only.
    function pause() external onlyOwner {
        _pause();
    }

    /// @notice Unpause the contract. Owner-only.
    function unpause() external onlyOwner {
        _unpause();
    }

    // ─── Internal ───────────────────────────────────────────────────────

    /// @dev Core registration logic shared by register, registerWithReferral, and bulkRegister.
    function _registerInternal(
        string calldata name,
        bytes32 deploymentID,
        address referrer
    ) internal {
        _validateName(name);
        if (bytes(name).length <= 3 && !shortNamesEnabled) revert ShortNamesDisabled();
        if (reservedNames[keccak256(abi.encodePacked(name))]) revert NameReserved(name);
        if (deploymentID == bytes32(0)) revert InvalidDeploymentID();
        if (_ownedNames[msg.sender].length >= MAX_NAMES_PER_OWNER) {
            revert TooManyNames(msg.sender, MAX_NAMES_PER_OWNER);
        }

        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];

        if (record.owner != address(0)) {
            // Allow re-registration if fully expired (past grace period)
            if (_isFullyExpired(nameHash)) {
                // Clean up old record
                _removeOwnedName(record.owner, nameHash);
                if (record.deploymentID != bytes32(0) && primaryName[record.deploymentID] == nameHash) {
                    delete primaryName[record.deploymentID];
                }
                delete metadata[nameHash];
            }
            // Allow re-registration if reservation expired
            else if (record.reservedUntil > 0 && block.timestamp > record.reservedUntil) {
                _removeOwnedName(record.owner, nameHash);
                delete metadata[nameHash];
            }
            else {
                revert NameAlreadyRegistered(name);
            }
        }

        // Calculate fee with premium + staking discount + referral
        uint256 price = _calculatePrice(name);
        price = _applyStakingDiscount(price, msg.sender);
        if (referrer != address(0)) {
            (uint256 discounted, uint256 referrerReward) = _applyReferralDiscount(price);
            _collectFee(msg.sender, discounted, referrer, referrerReward);
        } else {
            _collectFee(msg.sender, price, address(0), 0);
        }

        subdomains[nameHash] = SubdomainRecord({
            owner: msg.sender,
            deploymentID: deploymentID,
            registeredAt: uint48(block.timestamp),
            expiresAt: uint48(block.timestamp + expirationPeriod),
            reservedUntil: 0,
            referrer: referrer
        });
        nameOf[nameHash] = name;
        _ownedNames[msg.sender].push(nameHash);

        emit SubdomainRegistered(name, name, msg.sender, deploymentID, price);
    }

    /// @dev Calculate the base price for a name (with premium multiplier for short names).
    function _calculatePrice(string calldata name) internal view returns (uint256) {
        uint256 price = registrationFee;
        uint256 nameLen = bytes(name).length;
        if (nameLen == 1) {
            price = price * PREMIUM_1_CHAR_MULTIPLIER;
        } else if (nameLen == 2) {
            price = price * PREMIUM_2_CHAR_MULTIPLIER;
        } else if (nameLen == 3) {
            price = price * PREMIUM_3_CHAR_MULTIPLIER;
        } else if (nameLen == 4) {
            price = price * PREMIUM_4_CHAR_MULTIPLIER;
        }
        return price;
    }

    /// @dev Apply staking tier discount to a price.
    function _applyStakingDiscount(uint256 price, address user) internal view returns (uint256) {
        if (address(stakingContract) == address(0)) return price;
        try stakingContract.getTier(user) returns (uint8 tier) {
            if (tier > 5) tier = 5;
            uint256 discount = _tierDiscountBps[tier];
            if (discount > 0) {
                price = price - (price * discount / BPS_DENOMINATOR);
            }
        } catch {
            // If staking contract call fails, no discount
        }
        return price;
    }

    /// @dev Apply referral discount to a price.
    /// @return discounted The discounted price the user pays.
    /// @return referrerReward The reward amount for the referrer.
    function _applyReferralDiscount(uint256 price)
        internal
        view
        returns (uint256 discounted, uint256 referrerReward)
    {
        discounted = price - (price * referralDiscountBps / BPS_DENOMINATOR);
        // Referrer reward is based on the original (pre-referral-discount) price
        referrerReward = price * referralRewardBps / BPS_DENOMINATOR;
        return (discounted, referrerReward);
    }

    /// @dev Collect fee: pull tokens, burn 80%, treasury 20%, handle referral reward.
    /// @param referrerReward Pre-computed referrer reward (comes from treasury portion).
    function _collectFee(address payer, uint256 amount, address referrer, uint256 referrerReward) internal {
        if (amount == 0) return;

        bunkerToken.safeTransferFrom(payer, address(this), amount);

        uint256 burnAmount = (amount * BURN_BPS) / BPS_DENOMINATOR;
        uint256 treasuryAmount = amount - burnAmount - referrerReward;

        if (burnAmount > 0) {
            IBurnable(address(bunkerToken)).burn(burnAmount);
        }
        if (treasuryAmount > 0) {
            bunkerToken.safeTransfer(treasury, treasuryAmount);
        }
        if (referrerReward > 0) {
            bunkerToken.safeTransfer(referrer, referrerReward);
        }
    }

    /// @dev Check if a name is expired (expiresAt > 0 and past expiry).
    function _isExpired(bytes32 nameHash) internal view returns (bool) {
        uint48 exp = subdomains[nameHash].expiresAt;
        return exp > 0 && block.timestamp > exp;
    }

    /// @dev Check if a name is in its grace period (expired but within grace window).
    function _isInGracePeriod(bytes32 nameHash) internal view returns (bool) {
        uint48 exp = subdomains[nameHash].expiresAt;
        return exp > 0 && block.timestamp > exp && block.timestamp <= uint256(exp) + gracePeriod;
    }

    /// @dev Check if a name is fully expired (past grace period).
    function _isFullyExpired(bytes32 nameHash) internal view returns (bool) {
        uint48 exp = subdomains[nameHash].expiresAt;
        return exp > 0 && block.timestamp > uint256(exp) + gracePeriod;
    }

    /// @dev Validate a subdomain name: 1-32 chars, [a-z0-9-], no leading/trailing hyphens,
    ///      no consecutive hyphens.
    function _validateName(string calldata name) internal pure {
        bytes memory b = bytes(name);
        if (b.length < 1 || b.length > 32) revert InvalidName(name);

        // First and last char must be alphanumeric (no hyphens)
        if (!_isAlphanumeric(b[0]) || !_isAlphanumeric(b[b.length - 1])) {
            revert InvalidName(name);
        }

        // Middle chars: alphanumeric or hyphen, no consecutive hyphens
        bool prevHyphen = false;
        for (uint256 i = 1; i < b.length - 1; i++) {
            bool isHyphen = b[i] == 0x2D;
            if (!_isAlphanumeric(b[i]) && !isHyphen) {
                revert InvalidName(name);
            }
            if (isHyphen && prevHyphen) {
                revert InvalidName(name);
            }
            prevHyphen = isHyphen;
        }
    }

    /// @dev Check if a byte is lowercase letter or digit.
    function _isAlphanumeric(bytes1 c) internal pure returns (bool) {
        return (c >= 0x30 && c <= 0x39) || // 0-9
               (c >= 0x61 && c <= 0x7A);   // a-z
    }

    /// @dev Remove a name hash from an owner's list. Swap-and-pop for O(1).
    function _removeOwnedName(address owner, bytes32 nameHash) internal {
        bytes32[] storage names = _ownedNames[owner];
        for (uint256 i = 0; i < names.length; i++) {
            if (names[i] == nameHash) {
                names[i] = names[names.length - 1];
                names.pop();
                return;
            }
        }
    }
}
