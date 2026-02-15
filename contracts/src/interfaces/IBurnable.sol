// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @dev Minimal interface for ERC20Burnable.burn(uint256).
interface IBurnable {
    function burn(uint256 amount) external;
}
