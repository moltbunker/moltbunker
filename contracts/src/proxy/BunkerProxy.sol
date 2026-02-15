// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

/// @title BunkerProxy
/// @author Moltbunker
/// @notice ERC1967 proxy for upgradeable Moltbunker contracts.
/// @dev Used for BunkerStaking, BunkerEscrow, and BunkerPricing.
///      BunkerToken remains non-upgradeable (tokens should be immutable).
///
///      Current contracts use constructors with immutable state variables
///      (e.g., `IERC20 public immutable token`), which are stored in bytecode
///      rather than storage. This means they are NOT directly compatible with
///      proxy patterns. To enable full upgradeability, the contracts need to
///      be refactored to:
///        1. Replace constructors with `initialize()` functions
///        2. Replace `immutable` variables with regular storage variables
///        3. Add `_authorizeUpgrade()` for UUPS, or use TransparentProxy
///        4. Use OpenZeppelin's `Initializable` base contract
///
///      This proxy is deployed as infrastructure for when the contracts are
///      converted. See DeployUpgradeable.s.sol for the deployment template.
contract BunkerProxy is ERC1967Proxy {
    /// @param implementation Address of the implementation contract.
    /// @param data Encoded initializer call (abi.encodeWithSelector).
    constructor(address implementation, bytes memory data)
        ERC1967Proxy(implementation, data)
    {}
}
