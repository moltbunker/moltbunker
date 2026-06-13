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
/// Optional env vars:
///   SAFE_MULTISIG  - Gnosis Safe (or any second governance address). If set, it is
///                    granted PROPOSER_ROLE and CANCELLER_ROLE on the Timelock so a
///                    multisig can co-propose / veto operations. Public address — safe to log.
///
/// Usage (local):
///   forge script script/DeployProtocol.s.sol --rpc-url http://127.0.0.1:8545 --broadcast
///
/// Usage (Base mainnet):
///   forge script script/DeployProtocol.s.sol --rpc-url $BASE_RPC_URL --broadcast --verify
///
/// @dev Governance handoff: at the end of the deploy this script transfers ownership of
///      Staking, Escrow, and Pricing to the Timelock (step 1 of Ownable2Step) and schedules
///      a single Timelock batch that calls acceptOwnership() on all three (step 2). The
///      deployer EOA owns the contracts only for the minDelay window; after the delay the
///      operator must run executeBatch with the logged parameters to finalise the handoff.
///
/// @dev POST-DEPLOY OPERATIONAL NOTE — residual single-EOA timelock control:
///      This script renounces the deployer's DEFAULT_ADMIN_ROLE on the Timelock, but the
///      deployer is still seeded as the sole PROPOSER_ROLE and EXECUTOR_ROLE member (see the
///      proposers/executors arrays in _deployContracts). Until a Safe/second proposer is
///      wired, that single EOA can unilaterally propose AND execute any timelock operation —
///      so the timelock's separation-of-powers guarantee does not yet hold.
///      ACTION REQUIRED once SAFE_MULTISIG (or another independent proposer/executor) is live:
///        1. Grant PROPOSER_ROLE + EXECUTOR_ROLE to the Safe/second proposer (set SAFE_MULTISIG
///           on a redeploy, or schedule grantRole ops through the Timelock).
///        2. Schedule a Timelock op (the only path now that DEFAULT_ADMIN_ROLE is renounced)
///           that revokes PROPOSER_ROLE and EXECUTOR_ROLE from the deployer EOA, then execute
///           it after minDelay.
///      Track this as an open governance item until both deployer roles are revoked.
contract DeployProtocol is Script {
    /// @dev acceptOwnership() selector for the scheduled Timelock batch.
    bytes private constant ACCEPT_OWNERSHIP_CALLDATA = abi.encodeWithSignature("acceptOwnership()");

    // Deployed addresses stored in storage to avoid stack-too-deep in run().
    address public tokenAddr;
    address public stakingAddr;
    address public escrowAddr;
    address public pricingAddr;
    address public timelockAddr;
    address public treasuryAddr;
    address public guardianAddr;
    bytes32 public governanceOperationId;

    function run() external {
        uint256 deployerPk = vm.envUint("DEPLOYER_PK");
        address deployer = vm.addr(deployerPk);

        tokenAddr = vm.envAddress("BUNKER_TOKEN");
        treasuryAddr = vm.envAddress("TREASURY");
        guardianAddr = vm.envAddress("GUARDIAN");

        // Verify token is live
        require(bytes(BunkerToken(tokenAddr).name()).length > 0, "BUNKER_TOKEN not a valid ERC-20");
        console.log("Using BunkerToken at:", tokenAddr);

        vm.startBroadcast(deployerPk);

        _deployContracts(deployer, treasuryAddr, guardianAddr);
        _grantRoles();

        // Hand governance to the Timelock (transferOwnership now, scheduled
        // acceptOwnership batch, optional Safe co-proposer, deployer admin revoke).
        governanceOperationId =
            _wireGovernance(deployer, timelockAddr, stakingAddr, escrowAddr, pricingAddr);

        vm.stopBroadcast();

        _printSummary(deployer);
    }

    /// @dev Deploy Staking, Escrow, Pricing, and Timelock. Stores addresses in storage.
    function _deployContracts(address deployer, address treasury, address guardian) internal {
        // 1. Deploy Staking
        stakingAddr = address(new BunkerStaking(tokenAddr, treasury, deployer));
        console.log("BunkerStaking deployed at:", stakingAddr);

        // 2. Deploy Escrow
        escrowAddr = address(new BunkerEscrow(tokenAddr, treasury, deployer));
        console.log("BunkerEscrow deployed at: ", escrowAddr);

        // 3. Deploy Pricing
        pricingAddr = address(new BunkerPricing(deployer));
        console.log("BunkerPricing deployed at:", pricingAddr);

        // 4. Deploy Timelock
        address[] memory proposers = new address[](1);
        proposers[0] = deployer;
        address[] memory executors = new address[](1);
        executors[0] = deployer;
        timelockAddr = address(
            new BunkerTimelock(
                24 hours,    // minDelay
                proposers,
                executors,
                deployer,    // admin
                guardian     // guardian for emergency pause
            )
        );
        console.log("BunkerTimelock deployed at:", timelockAddr);
    }

    /// @dev Grant OPERATOR_ROLE (Escrow) and SLASHER_ROLE (Staking).
    function _grantRoles() internal {
        address operator = vm.envAddress("OPERATOR");
        address slasher = vm.envAddress("SLASHER");

        BunkerEscrow escrow = BunkerEscrow(escrowAddr);
        BunkerStaking staking = BunkerStaking(stakingAddr);
        escrow.grantRole(escrow.OPERATOR_ROLE(), operator);
        staking.grantRole(staking.SLASHER_ROLE(), slasher);
        console.log("OPERATOR_ROLE granted to: ", operator);
        console.log("SLASHER_ROLE granted to:  ", slasher);
    }

    /// @dev Print deployment + governance summary.
    function _printSummary(address deployer) internal view {
        console.log("");
        console.log("=== Deployment Summary ===");
        console.log("Token:   ", tokenAddr);
        console.log("Staking: ", stakingAddr);
        console.log("Escrow:  ", escrowAddr);
        console.log("Pricing: ", pricingAddr);
        console.log("Timelock:", timelockAddr);
        console.log("Treasury:", treasuryAddr);
        console.log("Guardian:", guardianAddr);
        console.log("Owner:   ", deployer);
        console.log("");
        console.log("=== Governance Handoff ===");
        console.log("acceptOwnership batch scheduled on Timelock.");
        console.log("Batch operationId:");
        console.logBytes32(governanceOperationId);
        console.log("After minDelay, run executeBatch on the Timelock to finalise (see _wireGovernance docs).");
        console.log("");
        console.log("=== OPEN GOVERNANCE ITEM ===");
        console.log("Deployer still holds sole PROPOSER_ROLE + EXECUTOR_ROLE on the Timelock.");
        console.log("Once a Safe/second proposer is wired, schedule a Timelock op to revoke");
        console.log("both roles from the deployer EOA (see DeployProtocol POST-DEPLOY NOTE).");
    }

    /// @notice Transfer ownership of the three Ownable2Step contracts to the Timelock
    ///         and schedule the matching acceptOwnership() batch through the Timelock.
    /// @dev Must be called inside an active broadcast (deployer holds PROPOSER_ROLE,
    ///      EXECUTOR_ROLE and DEFAULT_ADMIN_ROLE on the freshly deployed Timelock).
    ///      Ownable2Step requires the new owner (the Timelock) to call acceptOwnership();
    ///      we cannot call it inline because the Timelock only acts via scheduled ops, so
    ///      we schedule the batch and the operator executes it after the delay:
    ///
    ///        cast send <TIMELOCK> "executeBatch(address[],uint256[],bytes[],bytes32,bytes32)" \
    ///          "[<STAKING>,<ESCROW>,<PRICING>]" "[0,0,0]" \
    ///          "[0x79ba5097,0x79ba5097,0x79ba5097]" \
    ///          0x0000...0000 <SALT>
    ///
    ///      (0x79ba5097 == acceptOwnership() selector.)
    /// @return operationId The hashOperationBatch id the operator must execute after minDelay.
    function _wireGovernance(
        address deployer,
        address timelock,
        address staking,
        address escrow,
        address pricing
    ) internal returns (bytes32 operationId) {
        // Step 1 of Ownable2Step: nominate the Timelock as pending owner.
        BunkerStaking(staking).transferOwnership(timelock);
        BunkerEscrow(escrow).transferOwnership(timelock);
        BunkerPricing(pricing).transferOwnership(timelock);
        console.log("Ownership transfer (step 1) -> Timelock for Staking/Escrow/Pricing.");

        // Step 2 of Ownable2Step: schedule a Timelock batch that accepts all three.
        operationId = _scheduleAcceptBatch(timelock, staking, escrow, pricing);

        // Optional: add a Gnosis Safe (or any second governance address) as a
        // co-proposer / canceller so it can propose and veto operations.
        _maybeAddSafe(timelock);

        // Drop the deployer's super-admin power: from here only the Timelock itself
        // (via a scheduled op) can manage roles. This is the mainnet-safe posture.
        //
        // NOTE: this renounces DEFAULT_ADMIN_ROLE only. The deployer is still the sole
        // PROPOSER_ROLE + EXECUTOR_ROLE holder, so it retains unilateral propose+execute
        // control of the Timelock. That residual single-EOA control MUST be revoked via a
        // scheduled Timelock op once a Safe/second proposer is wired (see the contract-level
        // "POST-DEPLOY OPERATIONAL NOTE"). It is intentionally not revoked here because doing
        // so before a replacement proposer/executor exists would brick governance.
        BunkerTimelock tl = BunkerTimelock(payable(timelock));
        tl.renounceRole(tl.DEFAULT_ADMIN_ROLE(), deployer);
        console.log("Deployer DEFAULT_ADMIN_ROLE renounced on Timelock.");
    }

    /// @dev Build and schedule the acceptOwnership() batch through the Timelock.
    ///      Returns the hashOperationBatch id the operator executes after minDelay.
    function _scheduleAcceptBatch(
        address timelock,
        address staking,
        address escrow,
        address pricing
    ) internal returns (bytes32 operationId) {
        BunkerTimelock tl = BunkerTimelock(payable(timelock));

        address[] memory targets = new address[](3);
        targets[0] = staking;
        targets[1] = escrow;
        targets[2] = pricing;

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
    function _maybeAddSafe(address timelock) internal {
        address safe = _envOrZero("SAFE_MULTISIG");
        if (safe == address(0)) return;
        BunkerTimelock tl = BunkerTimelock(payable(timelock));
        tl.grantRole(tl.PROPOSER_ROLE(), safe);
        tl.grantRole(tl.CANCELLER_ROLE(), safe);
        console.log("SAFE_MULTISIG granted PROPOSER_ROLE + CANCELLER_ROLE:", safe);
    }

    /// @dev Read an optional address env var, returning address(0) when unset/empty.
    function _envOrZero(string memory key) internal view returns (address) {
        try vm.envAddress(key) returns (address val) {
            return val;
        } catch {
            return address(0);
        }
    }
}
