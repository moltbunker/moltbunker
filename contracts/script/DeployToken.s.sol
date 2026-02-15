// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import "../src/BunkerToken.sol";

/// @title DeployToken
/// @notice Step 1: Deploy BunkerToken. Run this first, verify, then run DeployProtocol.
///
/// Usage (local):
///   forge script script/DeployToken.s.sol --rpc-url http://127.0.0.1:8545 --broadcast
///
/// Usage (Base mainnet):
///   forge script script/DeployToken.s.sol --rpc-url $BASE_RPC_URL --broadcast --verify
contract DeployToken is Script {
    function run() external {
        uint256 deployerPk = vm.envUint("DEPLOYER_PK");
        address deployer = vm.addr(deployerPk);

        vm.startBroadcast(deployerPk);

        BunkerToken token = new BunkerToken(deployer);

        console.log("=== BunkerToken Deployed ===");
        console.log("Token:    ", address(token));
        console.log("Owner:    ", deployer);
        console.log("Supply Cap:", token.SUPPLY_CAP());
        console.log("");
        console.log("Next step: set BUNKER_TOKEN env var and run DeployProtocol.s.sol");
        console.log("  export BUNKER_TOKEN=%s", address(token));

        vm.stopBroadcast();
    }
}
