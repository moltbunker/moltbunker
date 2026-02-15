// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEscrow.sol";
import "../src/BunkerPricing.sol";
import "../src/BunkerTimelock.sol";
import "../src/BunkerDelegation.sol";
import "../src/BunkerReputation.sol";
import "../src/BunkerVerification.sol";

/// @title DeployTestnet
/// @notice Single-step deployment of all 8 Moltbunker contracts for testnet.
///
/// Required env vars:
///   DEPLOYER_PK    - Deployer private key
///   TREASURY       - Treasury wallet address
///
/// Optional env vars (defaults to deployer if not set):
///   OPERATOR       - Gets OPERATOR_ROLE on Escrow
///   SLASHER        - Gets SLASHER_ROLE on Staking
///   GUARDIAN       - Gets GUARDIAN_ROLE on Timelock
///   REPORTER       - Gets REPORTER_ROLE on Reputation
///   VERIFIER       - Gets VERIFIER_ROLE on Verification
///
/// Usage:
///   forge script script/DeployTestnet.s.sol --rpc-url $RPC_URL --broadcast --verify
contract DeployTestnet is Script {
    // Store deployed addresses in storage to avoid stack-too-deep
    address public tokenAddr;
    address public stakingAddr;
    address public escrowAddr;
    address public pricingAddr;
    address public timelockAddr;
    address public delegationAddr;
    address public reputationAddr;
    address public verificationAddr;

    function run() external {
        uint256 deployerPk = vm.envUint("DEPLOYER_PK");
        address deployer = vm.addr(deployerPk);
        address treasury = vm.envAddress("TREASURY");

        console.log("=== Deploying Moltbunker Testnet ===");
        console.log("Deployer:", deployer);
        console.log("Treasury:", treasury);

        vm.startBroadcast(deployerPk);

        _deployContracts(deployer, treasury);
        _grantRoles(deployer);

        vm.stopBroadcast();

        _printSummary();
    }

    function _deployContracts(address deployer, address treasury) internal {
        // 1. Token
        BunkerToken token = new BunkerToken(deployer);
        tokenAddr = address(token);

        // 2. Staking (needs token)
        BunkerStaking staking = new BunkerStaking(tokenAddr, treasury, deployer);
        stakingAddr = address(staking);

        // 3. Escrow (needs token)
        BunkerEscrow escrow = new BunkerEscrow(tokenAddr, treasury, deployer);
        escrowAddr = address(escrow);

        // 4. Pricing
        pricingAddr = address(new BunkerPricing(deployer));

        // 5. Timelock
        address[] memory proposers = new address[](1);
        proposers[0] = deployer;
        address[] memory executors = new address[](1);
        executors[0] = deployer;
        timelockAddr = address(
            new BunkerTimelock(24 hours, proposers, executors, deployer, deployer)
        );

        // 6. Delegation (needs token + staking)
        delegationAddr = address(
            new BunkerDelegation(tokenAddr, stakingAddr, deployer)
        );

        // 7. Reputation
        reputationAddr = address(new BunkerReputation(deployer));

        // 8. Verification
        verificationAddr = address(new BunkerVerification(deployer));

        // Wire escrow → staking
        escrow.setStakingContract(stakingAddr);

        // Testnet: short waiting times — guarded to prevent accidental mainnet use
        require(block.chainid == 84532, "DeployTestnet: Base Sepolia only (chainid 84532)");
        staking.setUnbondingPeriod(2 minutes);   // prod: 14 days
        staking.setAppealWindow(2 minutes);       // prod: 48 hours
        staking.setRewardsDuration(3 minutes);    // prod: 7 days
        staking.setVestingParams(0, 10000);       // disabled (100% immediate), prod: 30 days
        BunkerDelegation(delegationAddr).setUnbondingPeriod(2 minutes);   // prod: 7 days
        BunkerVerification(verificationAddr).setReinstatementCooldown(2 minutes); // prod: 7 days
    }

    function _grantRoles(address deployer) internal {
        address operator = _envOrDefault("OPERATOR", deployer);
        address slasher  = _envOrDefault("SLASHER", deployer);
        address guardian = _envOrDefault("GUARDIAN", deployer);
        address reporter = _envOrDefault("REPORTER", deployer);
        address verifier = _envOrDefault("VERIFIER", deployer);

        BunkerStaking(stakingAddr).grantRole(
            BunkerStaking(stakingAddr).SLASHER_ROLE(), slasher
        );
        BunkerEscrow(escrowAddr).grantRole(
            BunkerEscrow(escrowAddr).OPERATOR_ROLE(), operator
        );
        BunkerReputation(reputationAddr).grantRole(
            BunkerReputation(reputationAddr).REPORTER_ROLE(), reporter
        );
        BunkerVerification(verificationAddr).grantRole(
            BunkerVerification(verificationAddr).VERIFIER_ROLE(), verifier
        );

        // If guardian != deployer and we used deployer as timelock guardian,
        // grant guardian role on timelock
        if (guardian != deployer) {
            BunkerTimelock tl = BunkerTimelock(payable(timelockAddr));
            tl.grantRole(tl.GUARDIAN_ROLE(), guardian);
        }

        console.log("SLASHER_ROLE  ->", slasher);
        console.log("OPERATOR_ROLE ->", operator);
        console.log("REPORTER_ROLE ->", reporter);
        console.log("VERIFIER_ROLE ->", verifier);
        console.log("GUARDIAN_ROLE ->", guardian);
    }

    function _printSummary() internal view {
        console.log("");
        console.log("=== Deployed Addresses ===");
        console.log("VITE_TOKEN_ADDRESS=%s", tokenAddr);
        console.log("VITE_STAKING_ADDRESS=%s", stakingAddr);
        console.log("VITE_ESCROW_ADDRESS=%s", escrowAddr);
        console.log("VITE_PRICING_ADDRESS=%s", pricingAddr);
        console.log("VITE_TIMELOCK_ADDRESS=%s", timelockAddr);
        console.log("VITE_DELEGATION_ADDRESS=%s", delegationAddr);
        console.log("VITE_REPUTATION_ADDRESS=%s", reputationAddr);
        console.log("VITE_VERIFICATION_ADDRESS=%s", verificationAddr);
    }

    function _envOrDefault(string memory key, address fallback_) internal view returns (address) {
        try vm.envAddress(key) returns (address val) {
            return val;
        } catch {
            return fallback_;
        }
    }
}
