// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";

/// @title BunkerProxyAdmin
/// @author Moltbunker
/// @notice Admin contract for managing proxy upgrades across the Moltbunker protocol.
/// @dev Deployed once and manages all protocol proxies (BunkerStaking, BunkerEscrow,
///      BunkerPricing). Ownership should be transferred to BunkerTimelock for
///      governance-controlled upgrades.
///
///      Upgrade flow:
///        1. Deploy new implementation contract
///        2. ProxyAdmin.upgradeAndCall(proxy, newImpl, data) via timelock
///        3. Proxy delegates to new implementation transparently
///
///      NOTE: Current contracts use immutable variables in constructors and are
///      not yet compatible with the proxy pattern. This admin is deployed as
///      infrastructure for the future upgrade path. See DeployUpgradeable.s.sol.
contract BunkerProxyAdmin is ProxyAdmin {
    /// @param initialOwner The initial admin (should be BunkerTimelock address).
    constructor(address initialOwner) ProxyAdmin(initialOwner) {}
}
