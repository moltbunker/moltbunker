// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEscrow.sol";
import "../src/BunkerPricing.sol";
import "../src/BunkerTimelock.sol";

/// @title DeployProtocol
/// @notice Step 2: Deploy Staking, Escrow, Pricing using an already-deployed BunkerToken.
///
/// Required env vars:
///   DEPLOYER_PK    - Deployer private key
///   BUNKER_TOKEN   - Deployed BunkerToken address (from DeployToken step)
///   TREASURY       - Treasury wallet address
///   OPERATOR       - Operator address (gets OPERATOR_ROLE on Escrow)
///   SLASHER        - Slasher address (gets SLASHER_ROLE on Staking)
///   GUARDIAN       - Guardian address (gets GUARDIAN_ROLE on Timelock for emergency pause)
///
/// Usage (local):
///   forge script script/DeployProtocol.s.sol --rpc-url http://127.0.0.1:8545 --broadcast
///
/// Usage (Base mainnet):
///   forge script script/DeployProtocol.s.sol --rpc-url $BASE_RPC_URL --broadcast --verify
contract DeployProtocol is Script {
    function run() external {
        uint256 deployerPk = vm.envUint("DEPLOYER_PK");
        address deployer = vm.addr(deployerPk);
        address tokenAddr = vm.envAddress("BUNKER_TOKEN");
        address treasury = vm.envAddress("TREASURY");
        address operator = vm.envAddress("OPERATOR");
        address slasher = vm.envAddress("SLASHER");
        address guardian = vm.envAddress("GUARDIAN");

        // Verify token is live
        BunkerToken token = BunkerToken(tokenAddr);
        require(bytes(token.name()).length > 0, "BUNKER_TOKEN not a valid ERC-20");
        console.log("Using BunkerToken at:", tokenAddr);

        vm.startBroadcast(deployerPk);

        // 1. Deploy Staking
        BunkerStaking staking = new BunkerStaking(tokenAddr, treasury, deployer);
        console.log("BunkerStaking deployed at:", address(staking));

        // 2. Deploy Escrow
        BunkerEscrow escrow = new BunkerEscrow(tokenAddr, treasury, deployer);
        console.log("BunkerEscrow deployed at: ", address(escrow));

        // 3. Deploy Pricing
        BunkerPricing pricing = new BunkerPricing(deployer);
        console.log("BunkerPricing deployed at:", address(pricing));

        // 4. Deploy Timelock
        address[] memory proposers = new address[](1);
        proposers[0] = deployer;
        address[] memory executors = new address[](1);
        executors[0] = deployer;
        BunkerTimelock timelock = new BunkerTimelock(
            24 hours,    // minDelay
            proposers,
            executors,
            deployer,    // admin
            guardian     // guardian for emergency pause
        );
        console.log("BunkerTimelock deployed at:", address(timelock));

        // 5. Grant roles
        escrow.grantRole(escrow.OPERATOR_ROLE(), operator);
        staking.grantRole(staking.SLASHER_ROLE(), slasher);
        console.log("OPERATOR_ROLE granted to: ", operator);
        console.log("SLASHER_ROLE granted to:  ", slasher);

        // NOTE: After deployment verification, manually transfer ownership of
        // BunkerStaking, BunkerEscrow, and BunkerPricing to the BunkerTimelock
        // address using their transferOwnership() functions. Then accept from
        // the timelock via a scheduled operation.

        vm.stopBroadcast();

        console.log("");
        console.log("=== Deployment Summary ===");
        console.log("Token:   ", tokenAddr);
        console.log("Staking: ", address(staking));
        console.log("Escrow:  ", address(escrow));
        console.log("Pricing: ", address(pricing));
        console.log("Timelock:", address(timelock));
        console.log("Treasury:", treasury);
        console.log("Guardian:", guardian);
        console.log("Owner:   ", deployer);
    }
}
