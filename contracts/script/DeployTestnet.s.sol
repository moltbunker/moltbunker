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
import "../src/BunkerRegistry.sol";

/// @title DeployTestnet
/// @notice Single-step deployment of all 9 Moltbunker contracts for testnet.
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
///   SAFE_MULTISIG  - If set, granted PROPOSER_ROLE + CANCELLER_ROLE on the Timelock
///                    (public address, safe to log).
///
/// Usage:
///   forge script script/DeployTestnet.s.sol --rpc-url $RPC_URL --broadcast --verify
///
/// @dev After deployment the script transfers ownership of Staking, Escrow, and Pricing
///      to the Timelock (Ownable2Step step 1) and schedules a Timelock batch that calls
///      acceptOwnership() on all three (step 2). On Base Sepolia (chainid 84532) the
///      Timelock minDelay is 24h, so the operator must run executeBatch with the logged
///      operationId after the delay to finalise the handoff.
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
    address public registryAddr;

    /// @notice acceptOwnership() batch id the operator must execute after minDelay.
    bytes32 public governanceOperationId;

    /// @dev acceptOwnership() selector for the scheduled Timelock batch.
    bytes private constant ACCEPT_OWNERSHIP_CALLDATA = abi.encodeWithSignature("acceptOwnership()");

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

        // Hand governance to the Timelock: transferOwnership now, schedule the
        // acceptOwnership batch, add optional Safe co-proposer, renounce deployer admin.
        governanceOperationId = _wireGovernance(deployer);

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

        // 9. Registry (needs token, treasury, staking)
        registryAddr = address(
            new BunkerRegistry(tokenAddr, treasury, 1_000_000 * 1e18, deployer, stakingAddr)
        );

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

    /// @notice Transfer ownership of Staking, Escrow, and Pricing to the Timelock and
    ///         schedule the matching acceptOwnership() batch.
    /// @dev Must run inside the deployer broadcast — the deployer holds PROPOSER_ROLE,
    ///      EXECUTOR_ROLE and DEFAULT_ADMIN_ROLE on the freshly deployed Timelock.
    ///      Ownable2Step needs the new owner (the Timelock) to call acceptOwnership(),
    ///      which only happens via a scheduled op, so the operator runs executeBatch
    ///      after the 24h delay:
    ///
    ///        cast send <TIMELOCK> "executeBatch(address[],uint256[],bytes[],bytes32,bytes32)" \
    ///          "[<STAKING>,<ESCROW>,<PRICING>]" "[0,0,0]" \
    ///          "[0x79ba5097,0x79ba5097,0x79ba5097]" \
    ///          0x0000...0000 <SALT>
    ///
    ///      (0x79ba5097 == acceptOwnership() selector.)
    /// @return operationId The hashOperationBatch id to execute after minDelay (24h on testnet).
    function _wireGovernance(address deployer) internal returns (bytes32 operationId) {
        // Step 1 of Ownable2Step: nominate the Timelock as pending owner.
        BunkerStaking(stakingAddr).transferOwnership(timelockAddr);
        BunkerEscrow(escrowAddr).transferOwnership(timelockAddr);
        BunkerPricing(pricingAddr).transferOwnership(timelockAddr);
        console.log("Ownership transfer (step 1) -> Timelock for Staking/Escrow/Pricing.");

        // Step 2 of Ownable2Step: schedule the Timelock batch that accepts all three.
        operationId = _scheduleAcceptBatch();

        // Optional: add a Gnosis Safe (or any second governance address) as a
        // co-proposer / canceller so a multisig can co-propose and veto.
        _maybeAddSafe();

        // Drop the deployer's super-admin power on the Timelock. Uses getMinDelay()
        // (24h on testnet) for the scheduled handoff; role management now flows only
        // through scheduled Timelock operations.
        BunkerTimelock tl = BunkerTimelock(payable(timelockAddr));
        tl.renounceRole(tl.DEFAULT_ADMIN_ROLE(), deployer);
        console.log("Deployer DEFAULT_ADMIN_ROLE renounced on Timelock.");
    }

    /// @dev Build and schedule the acceptOwnership() batch through the Timelock.
    function _scheduleAcceptBatch() internal returns (bytes32 operationId) {
        BunkerTimelock tl = BunkerTimelock(payable(timelockAddr));

        address[] memory targets = new address[](3);
        targets[0] = stakingAddr;
        targets[1] = escrowAddr;
        targets[2] = pricingAddr;

        uint256[] memory values = new uint256[](3);

        bytes[] memory payloads = new bytes[](3);
        payloads[0] = ACCEPT_OWNERSHIP_CALLDATA;
        payloads[1] = ACCEPT_OWNERSHIP_CALLDATA;
        payloads[2] = ACCEPT_OWNERSHIP_CALLDATA;

        bytes32 salt = keccak256("BUNKER_GOVERNANCE_HANDOFF_V1");
        uint256 delay = tl.getMinDelay();

        operationId = tl.hashOperationBatch(targets, values, payloads, bytes32(0), salt);
        tl.scheduleBatch(targets, values, payloads, bytes32(0), salt, delay);
    }

    /// @dev Grant PROPOSER_ROLE + CANCELLER_ROLE to SAFE_MULTISIG if the env var is set.
    function _maybeAddSafe() internal {
        address safe = _envOrDefault("SAFE_MULTISIG", address(0));
        if (safe == address(0)) return;
        BunkerTimelock tl = BunkerTimelock(payable(timelockAddr));
        tl.grantRole(tl.PROPOSER_ROLE(), safe);
        tl.grantRole(tl.CANCELLER_ROLE(), safe);
        console.log("SAFE_MULTISIG granted PROPOSER_ROLE + CANCELLER_ROLE:", safe);
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
        console.log("VITE_REGISTRY_ADDRESS=%s", registryAddr);
        console.log("");
        console.log("=== Governance Handoff ===");
        console.log("acceptOwnership batch scheduled on Timelock. operationId:");
        console.logBytes32(governanceOperationId);
        console.log("Run executeBatch on the Timelock after minDelay (24h) to finalise.");
    }

    function _envOrDefault(string memory key, address fallback_) internal view returns (address) {
        try vm.envAddress(key) returns (address val) {
            return val;
        } catch {
            return fallback_;
        }
    }
}
