// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import {BunkerRegistry} from "../src/BunkerRegistry.sol";

/// @title DeployRegistry
/// @notice Standalone deployment of BunkerRegistry to an existing testnet.
///         All other contracts (Token, Staking, etc.) are already deployed.
///
/// Required env vars:
///   DEPLOYER_PK    - Deployer private key
///   TREASURY       - Treasury wallet address (same as used for other contracts)
///
/// The token and staking addresses are hardcoded to the existing Base Sepolia deployment.
///
/// Usage:
///   cd contracts
///   forge script script/DeployRegistry.s.sol \
///     --rpc-url https://sepolia.base.org \
///     --broadcast --verify
contract DeployRegistry is Script {
    // Existing BunkerToken on Base Sepolia (deployed 2026-02-13)
    address constant TOKEN = 0x4cc3F5C0d2Ecb4118e214980906eFe5c880a6ceA;

    // Existing BunkerStaking on Base Sepolia (deployed 2026-02-13)
    address constant STAKING = 0xDC76d972a827D2a19867EF9aBD335014d5Cf7D6a;

    // Registration fee: 1,000,000 BUNKER (with 18 decimals)
    uint256 constant REGISTRATION_FEE = 1_000_000 * 1e18;

    function run() external {
        uint256 deployerPk = vm.envUint("DEPLOYER_PK");
        address deployer = vm.addr(deployerPk);
        address treasury = vm.envAddress("TREASURY");

        console.log("=== Deploying BunkerRegistry v2.0.0 ===");
        console.log("Deployer:", deployer);
        console.log("Treasury:", treasury);
        console.log("Token:   ", TOKEN);
        console.log("Staking: ", STAKING);
        console.log("Fee:      1,000,000 BUNKER");

        vm.startBroadcast(deployerPk);

        BunkerRegistry registry = new BunkerRegistry(
            TOKEN,
            treasury,
            REGISTRATION_FEE,
            deployer,
            STAKING
        );

        vm.stopBroadcast();

        console.log("");
        console.log("=== BunkerRegistry v2.0.0 Deployed ===");
        console.log("REGISTRY_ADDRESS=%s", address(registry));
        console.log("");
        console.log("Next steps:");
        console.log("  1. Update configs/daemon.yaml -> subdomain_registry_address: <address>");
        console.log("  2. Update web-admin/.env.local -> VITE_REGISTRY_ADDRESS=<address>");
    }
}
