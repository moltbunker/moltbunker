// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";
import "@openzeppelin/contracts/access/Ownable2Step.sol";

/// @title BunkerToken
/// @author Moltbunker
/// @notice ERC-20 token for the Moltbunker decentralized compute network.
/// @dev Capped at 100 billion tokens. Owner-controlled minting. EIP-2612 Permit.
///      Burnable by any holder. Deployed on Base L2.
contract BunkerToken is ERC20, ERC20Burnable, ERC20Permit, Ownable2Step {
    /// @notice Maximum token supply: 100 billion BUNKER (18 decimals).
    uint256 public constant SUPPLY_CAP = 100_000_000_000 * 1e18;

    /// @notice Contract version for daemon compatibility checks.
    string public constant VERSION = "1.0.0";

    /// @notice Thrown when minting would exceed the supply cap.
    error SupplyCapExceeded(uint256 requested, uint256 available);

    /// @notice Thrown when minting zero tokens.
    error MintAmountZero();

    /// @param initialOwner Address that will own the contract and control minting.
    constructor(address initialOwner)
        ERC20("Bunker Token", "BUNKER")
        ERC20Permit("Bunker Token")
        Ownable(initialOwner)
    {}

    /// @notice Mint new tokens to a specified address.
    /// @param to Recipient address.
    /// @param amount Number of tokens to mint (in wei, 18 decimals).
    function mint(address to, uint256 amount) external onlyOwner {
        if (amount == 0) revert MintAmountZero();
        uint256 available = SUPPLY_CAP - totalSupply();
        if (amount > available) revert SupplyCapExceeded(amount, available);
        _mint(to, amount);
    }

    /// @notice Returns the maximum number of tokens that can still be minted.
    function mintableSupply() external view returns (uint256 remaining) {
        return SUPPLY_CAP - totalSupply();
    }
}
