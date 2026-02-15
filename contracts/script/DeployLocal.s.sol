// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEscrow.sol";
import "../src/BunkerPricing.sol";
import "../src/BunkerTimelock.sol";

/// @title DeployLocal
/// @notice Deploys all Moltbunker contracts to a local Anvil instance.
/// @dev Uses Anvil's default accounts:
///      Account 0 (deployer/owner): 0xf39Fd6...  PK: 0xac0974...
///      Account 1 (treasury):       0x70997970...
///      Account 2 (operator):       0x3C44Cd...
///      Account 3 (slasher):        0x90F79b...
///      Account 4-9 (providers/requesters)
///
/// Usage:
///   anvil &
///   forge script script/DeployLocal.s.sol --rpc-url http://127.0.0.1:8545 --broadcast
contract DeployLocal is Script {
    // Anvil default private keys
    uint256 constant DEPLOYER_PK = 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80;

    // Anvil default addresses
    address constant DEPLOYER  = 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266;
    address constant TREASURY  = 0x70997970C51812dc3A010C7d01b50e0d17dc79C8;
    address constant OPERATOR  = 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC;
    address constant SLASHER   = 0x90F79bf6EB2c4f870365E785982E1f101E93b906;
    address constant PROVIDER1 = 0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65;
    address constant PROVIDER2 = 0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc;
    address constant PROVIDER3 = 0x976EA74026E726554dB657fA54763abd0C3a0aa9;
    address constant REQUESTER = 0x14dC79964da2C08b23698B3D3cc7Ca32193d9955;
    address constant GUARDIAN  = 0xa0Ee7A142d267C1f36714E4a8F75612F20a79720;

    function run() external {
        vm.startBroadcast(DEPLOYER_PK);

        // 1. Deploy BunkerToken
        BunkerToken token = new BunkerToken(DEPLOYER);
        console.log("BunkerToken deployed at:", address(token));

        // 2. Deploy BunkerStaking
        BunkerStaking staking = new BunkerStaking(
            address(token),
            TREASURY,
            DEPLOYER
        );
        console.log("BunkerStaking deployed at:", address(staking));

        // 3. Deploy BunkerEscrow
        BunkerEscrow escrow = new BunkerEscrow(
            address(token),
            TREASURY,
            DEPLOYER
        );
        console.log("BunkerEscrow deployed at:", address(escrow));

        // 4. Deploy BunkerPricing
        BunkerPricing pricing = new BunkerPricing(DEPLOYER);
        console.log("BunkerPricing deployed at:", address(pricing));

        // 4b. Deploy BunkerTimelock
        address[] memory proposers = new address[](1);
        proposers[0] = DEPLOYER;
        address[] memory executors = new address[](1);
        executors[0] = DEPLOYER;
        BunkerTimelock timelock = new BunkerTimelock(
            24 hours,    // minDelay (MIN_DELAY_FLOOR enforced by contract)
            proposers,
            executors,
            DEPLOYER,
            GUARDIAN
        );
        console.log("BunkerTimelock deployed at:", address(timelock));

        // 5. Grant roles
        escrow.grantRole(escrow.OPERATOR_ROLE(), OPERATOR);
        staking.grantRole(staking.SLASHER_ROLE(), SLASHER);
        console.log("Roles granted: OPERATOR ->", OPERATOR);
        console.log("Roles granted: SLASHER ->", SLASHER);

        // 6. Mint tokens for testing
        uint256 mintAmount = 10_000_000e18; // 10M BUNKER each

        token.mint(PROVIDER1, mintAmount);
        token.mint(PROVIDER2, mintAmount);
        token.mint(PROVIDER3, mintAmount);
        token.mint(REQUESTER, mintAmount);
        token.mint(OPERATOR, mintAmount);
        console.log("Minted 10M BUNKER to each test account");

        console.log("");
        console.log("=== Deployment Summary ===");
        console.log("Token:   ", address(token));
        console.log("Staking: ", address(staking));
        console.log("Escrow:  ", address(escrow));
        console.log("Pricing: ", address(pricing));
        console.log("Timelock:", address(timelock));
        console.log("Treasury:", TREASURY);
        console.log("Guardian:", GUARDIAN);
        console.log("Operator:", OPERATOR);
        console.log("Slasher: ", SLASHER);

        vm.stopBroadcast();
    }
}
