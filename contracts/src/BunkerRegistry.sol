// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/Ownable2Step.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "./interfaces/IBurnable.sol";

/// @title BunkerRegistry
/// @author Moltbunker
/// @notice On-chain registry for vanity subdomain names (<name>.moltbunker.dev).
/// @dev Names are mapped to deployment IDs. Registration requires a BUNKER fee
///      (80% burned, 20% treasury). Auto-assigned 8-char prefixes from deployment
///      IDs are free and don't go through this contract.
contract BunkerRegistry is Ownable2Step, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;

    // ─── State ───────────────────────────────────────────────────────────

    /// @notice BUNKER token contract.
    IERC20 public immutable bunkerToken;

    /// @notice Treasury address receiving 20% of registration fees.
    address public treasury;

    /// @notice Registration fee in BUNKER tokens (default: 10,000 * 1e18).
    uint256 public registrationFee;

    /// @notice Burn share in basis points (default: 8000 = 80%).
    uint256 public constant BURN_BPS = 8000;

    /// @notice BPS denominator.
    uint256 public constant BPS_DENOMINATOR = 10000;

    /// @notice Maximum number of names a single address can own.
    uint256 public constant MAX_NAMES_PER_OWNER = 100;

    /// @notice Minimum registration fee (1,000 BUNKER) to prevent fee removal attacks.
    uint256 public constant MIN_REGISTRATION_FEE = 1000 * 1e18;

    /// @notice Contract version.
    string public constant VERSION = "1.0.0";

    struct SubdomainRecord {
        address owner;
        bytes32 deploymentID;
        uint256 registeredAt;
    }

    /// @notice Name hash → subdomain record.
    mapping(bytes32 => SubdomainRecord) public subdomains;

    /// @notice Name hash → original name string (for enumeration).
    mapping(bytes32 => string) public nameOf;

    /// @notice Owner → list of name hashes they own.
    mapping(address => bytes32[]) private _ownedNames;

    // ─── Events ──────────────────────────────────────────────────────────

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

    event RegistrationFeeUpdated(uint256 oldFee, uint256 newFee);
    event TreasuryUpdated(address oldTreasury, address newTreasury);

    // ─── Errors ──────────────────────────────────────────────────────────

    error NameAlreadyRegistered(string name);
    error NameNotRegistered(string name);
    error NotNameOwner(string name, address caller);
    error InvalidName(string name);
    error InvalidAddress();
    error InvalidDeploymentID();
    error FeeTransferFailed();
    error TooManyNames(address owner, uint256 max);
    error FeeBelowMinimum(uint256 fee, uint256 minimum);

    // ─── Constructor ─────────────────────────────────────────────────────

    /// @param _token BUNKER token address.
    /// @param _treasury Treasury address for fee collection.
    /// @param _registrationFee Initial registration fee in BUNKER (with 18 decimals).
    /// @param _owner Contract owner (typically Timelock).
    constructor(
        address _token,
        address _treasury,
        uint256 _registrationFee,
        address _owner
    ) Ownable(_owner) {
        if (_token == address(0) || _treasury == address(0)) revert InvalidAddress();
        bunkerToken = IERC20(_token);
        treasury = _treasury;
        registrationFee = _registrationFee;
    }

    // ─── Core Operations ─────────────────────────────────────────────────

    /// @notice Register a vanity subdomain name.
    /// @param name The subdomain name (3-32 chars, lowercase alphanumeric + hyphens).
    /// @param deploymentID The deployment to point this name to.
    function register(string calldata name, bytes32 deploymentID)
        external
        nonReentrant
        whenNotPaused
    {
        _validateName(name);
        if (deploymentID == bytes32(0)) revert InvalidDeploymentID();
        if (_ownedNames[msg.sender].length >= MAX_NAMES_PER_OWNER) {
            revert TooManyNames(msg.sender, MAX_NAMES_PER_OWNER);
        }

        bytes32 nameHash = keccak256(abi.encodePacked(name));
        if (subdomains[nameHash].owner != address(0)) revert NameAlreadyRegistered(name);

        // Collect fee: 80% burned, 20% treasury
        if (registrationFee > 0) {
            uint256 burnAmount = (registrationFee * BURN_BPS) / BPS_DENOMINATOR;
            uint256 treasuryAmount = registrationFee - burnAmount;

            bunkerToken.safeTransferFrom(msg.sender, address(this), registrationFee);

            // Burn
            if (burnAmount > 0) {
                IBurnable(address(bunkerToken)).burn(burnAmount);
            }
            // Treasury
            if (treasuryAmount > 0) {
                bunkerToken.safeTransfer(treasury, treasuryAmount);
            }
        }

        subdomains[nameHash] = SubdomainRecord({
            owner: msg.sender,
            deploymentID: deploymentID,
            registeredAt: block.timestamp
        });
        nameOf[nameHash] = name;
        _ownedNames[msg.sender].push(nameHash);

        emit SubdomainRegistered(name, name, msg.sender, deploymentID, registrationFee);
    }

    /// @notice Release a subdomain name. No refund.
    /// @param name The subdomain name to release.
    function release(string calldata name) external nonReentrant {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotRegistered(name);
        if (record.owner != msg.sender) revert NotNameOwner(name, msg.sender);

        address owner = record.owner;
        delete subdomains[nameHash];
        delete nameOf[nameHash];
        _removeOwnedName(owner, nameHash);

        emit SubdomainReleased(name, name, owner);
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

        bytes32 nameHash = keccak256(abi.encodePacked(name));
        SubdomainRecord storage record = subdomains[nameHash];
        if (record.owner == address(0)) revert NameNotRegistered(name);
        if (record.owner != msg.sender) revert NotNameOwner(name, msg.sender);

        address oldOwner = record.owner;
        record.owner = newOwner;
        _removeOwnedName(oldOwner, nameHash);
        _ownedNames[newOwner].push(nameHash);

        emit SubdomainTransferred(name, name, oldOwner, newOwner);
    }

    /// @notice Update the deployment ID a name points to.
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

        bytes32 oldID = record.deploymentID;
        record.deploymentID = newDeploymentID;

        emit SubdomainUpdated(name, name, oldID, newDeploymentID);
    }

    // ─── View Functions ──────────────────────────────────────────────────

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
        return (record.owner, record.deploymentID, record.registeredAt);
    }

    /// @notice Check if a name is available for registration.
    function isAvailable(string calldata name) external view returns (bool) {
        bytes32 nameHash = keccak256(abi.encodePacked(name));
        return subdomains[nameHash].owner == address(0);
    }

    /// @notice Get the number of names owned by an address.
    function nameCount(address owner) external view returns (uint256) {
        return _ownedNames[owner].length;
    }

    /// @notice Get the name hash at a given index for an owner.
    function ownedNameAt(address owner, uint256 index) external view returns (bytes32) {
        return _ownedNames[owner][index];
    }

    // ─── Admin Functions ─────────────────────────────────────────────────

    /// @notice Update the registration fee. Owner-only (via Timelock).
    /// @dev Fee must be 0 (free) or >= MIN_REGISTRATION_FEE to prevent trivial name squatting.
    function setRegistrationFee(uint256 newFee) external onlyOwner {
        if (newFee > 0 && newFee < MIN_REGISTRATION_FEE) {
            revert FeeBelowMinimum(newFee, MIN_REGISTRATION_FEE);
        }
        uint256 oldFee = registrationFee;
        registrationFee = newFee;
        emit RegistrationFeeUpdated(oldFee, newFee);
    }

    /// @notice Update the treasury address. Owner-only (via Timelock).
    function setTreasury(address newTreasury) external onlyOwner {
        if (newTreasury == address(0)) revert InvalidAddress();
        address oldTreasury = treasury;
        treasury = newTreasury;
        emit TreasuryUpdated(oldTreasury, newTreasury);
    }

    /// @notice Pause the contract. Owner-only.
    function pause() external onlyOwner {
        _pause();
    }

    /// @notice Unpause the contract. Owner-only.
    function unpause() external onlyOwner {
        _unpause();
    }

    // ─── Internal ────────────────────────────────────────────────────────

    /// @dev Validate a subdomain name: 3-32 chars, [a-z0-9-], no leading/trailing hyphens.
    function _validateName(string calldata name) internal pure {
        bytes memory b = bytes(name);
        if (b.length < 3 || b.length > 32) revert InvalidName(name);

        // First and last char must be alphanumeric
        if (!_isAlphanumeric(b[0]) || !_isAlphanumeric(b[b.length - 1])) {
            revert InvalidName(name);
        }

        // Middle chars: alphanumeric or hyphen
        for (uint256 i = 1; i < b.length - 1; i++) {
            if (!_isAlphanumeric(b[i]) && b[i] != 0x2D) { // 0x2D = '-'
                revert InvalidName(name);
            }
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
