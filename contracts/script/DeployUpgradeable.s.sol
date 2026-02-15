// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEscrow.sol";
import "../src/BunkerPricing.sol";
import "../src/BunkerTimelock.sol";
import "../src/proxy/BunkerProxyAdmin.sol";

/// @title DeployUpgradeable
/// @author Moltbunker
/// @notice Template for proxy-based deployment of the Moltbunker protocol.
///
/// @dev IMPORTANT: This script is a TEMPLATE for the future upgradeable deployment.
///      Current contracts (BunkerStaking, BunkerEscrow) use immutable variables set
///      in constructors, which are stored in bytecode rather than storage. This means
///      they CANNOT be used behind proxies as-is.
///
///      To enable full upgradeability, each contract needs these changes:
///
///      BunkerStaking:
///        - TODO: Replace `IERC20 public immutable token` with `IERC20 public token`
///        - TODO: Replace `IERC20 public immutable rewardsToken` with storage variable
///        - TODO: Replace constructor with `function initialize(address _token, address _treasury, address _initialOwner) external initializer`
///        - TODO: Add `Initializable` from OpenZeppelin
///        - TODO: Add `UUPSUpgradeable` with `_authorizeUpgrade(address) onlyOwner`
///        - TODO: Replace `Ownable(_initialOwner)` with `__Ownable_init(_initialOwner)` etc.
///
///      BunkerEscrow:
///        - TODO: Replace `IERC20 public immutable token` with `IERC20 public token`
///        - TODO: Replace constructor with `function initialize(address _token, address _treasury, address _initialOwner) external initializer`
///        - TODO: Add `Initializable` and `UUPSUpgradeable`
///        - TODO: Replace inheritance chain with upgradeable variants
///
///      BunkerPricing:
///        - TODO: Replace constructor with `function initialize(address _initialOwner) external initializer`
///        - TODO: Add `Initializable` and `UUPSUpgradeable`
///
///      Until these changes are made, use DeployProtocol.s.sol for direct deployment.
///
/// Required env vars:
///   DEPLOYER_PK    - Deployer private key
///   BUNKER_TOKEN   - Deployed BunkerToken address (from DeployToken step)
///   TREASURY       - Treasury wallet address
///   OPERATOR       - Operator address (gets OPERATOR_ROLE on Escrow)
///   SLASHER        - Slasher address (gets SLASHER_ROLE on Staking)
///   GUARDIAN       - Guardian address (gets GUARDIAN_ROLE on Timelock)
///
/// Usage (when contracts are converted to initializable):
///   forge script script/DeployUpgradeable.s.sol --rpc-url $BASE_RPC_URL --broadcast --verify
contract DeployUpgradeable is Script {
    function run() external {
        uint256 deployerPk = vm.envUint("DEPLOYER_PK");
        address deployer = vm.addr(deployerPk);
        address tokenAddr = vm.envAddress("BUNKER_TOKEN");
        address treasury = vm.envAddress("TREASURY");
        address operator = vm.envAddress("OPERATOR");
        address slasher = vm.envAddress("SLASHER");
        address guardian = vm.envAddress("GUARDIAN");

        // Verify token is live.
        BunkerToken token = BunkerToken(tokenAddr);
        require(bytes(token.name()).length > 0, "BUNKER_TOKEN not a valid ERC-20");
        console.log("Using BunkerToken at:", tokenAddr);

        vm.startBroadcast(deployerPk);

        // ═══════════════════════════════════════════════
        //  Step 1: Deploy BunkerToken directly (immutable, no proxy)
        // ═══════════════════════════════════════════════
        // Token is already deployed via DeployToken.s.sol.
        // Tokens SHOULD be immutable - no upgrade path needed.
        console.log("BunkerToken (immutable):", tokenAddr);

        // ═══════════════════════════════════════════════
        //  Step 2: Deploy ProxyAdmin (manages all proxies)
        // ═══════════════════════════════════════════════
        // TODO: Once contracts are converted to initializable, deploy ProxyAdmin.
        // BunkerProxyAdmin proxyAdmin = new BunkerProxyAdmin(deployer);
        // console.log("ProxyAdmin deployed at:", address(proxyAdmin));

        // ═══════════════════════════════════════════════
        //  Step 3: Deploy BunkerStaking via proxy
        // ═══════════════════════════════════════════════
        // TODO: Deploy implementation, then proxy.
        //
        // BunkerStaking stakingImpl = new BunkerStaking();  // No constructor args
        //
        // bytes memory stakingInit = abi.encodeWithSelector(
        //     BunkerStaking.initialize.selector,
        //     tokenAddr,
        //     treasury,
        //     deployer
        // );
        //
        // TransparentUpgradeableProxy stakingProxy = new TransparentUpgradeableProxy(
        //     address(stakingImpl),
        //     address(proxyAdmin),
        //     stakingInit
        // );
        //
        // BunkerStaking staking = BunkerStaking(address(stakingProxy));
        // console.log("BunkerStaking implementation:", address(stakingImpl));
        // console.log("BunkerStaking proxy:         ", address(stakingProxy));

        // For now, deploy directly (non-upgradeable).
        BunkerStaking staking = new BunkerStaking(tokenAddr, treasury, deployer);
        console.log("BunkerStaking (direct):", address(staking));

        // ═══════════════════════════════════════════════
        //  Step 4: Deploy BunkerEscrow via proxy
        // ═══════════════════════════════════════════════
        // TODO: Deploy implementation, then proxy.
        //
        // BunkerEscrow escrowImpl = new BunkerEscrow();  // No constructor args
        //
        // bytes memory escrowInit = abi.encodeWithSelector(
        //     BunkerEscrow.initialize.selector,
        //     tokenAddr,
        //     treasury,
        //     deployer
        // );
        //
        // TransparentUpgradeableProxy escrowProxy = new TransparentUpgradeableProxy(
        //     address(escrowImpl),
        //     address(proxyAdmin),
        //     escrowInit
        // );
        //
        // BunkerEscrow escrow = BunkerEscrow(address(escrowProxy));
        // console.log("BunkerEscrow implementation:", address(escrowImpl));
        // console.log("BunkerEscrow proxy:         ", address(escrowProxy));

        // For now, deploy directly (non-upgradeable).
        BunkerEscrow escrow = new BunkerEscrow(tokenAddr, treasury, deployer);
        console.log("BunkerEscrow (direct): ", address(escrow));

        // ═══════════════════════════════════════════════
        //  Step 5: Deploy BunkerPricing via proxy
        // ═══════════════════════════════════════════════
        // TODO: Deploy implementation, then proxy.
        //
        // BunkerPricing pricingImpl = new BunkerPricing();  // No constructor args
        //
        // bytes memory pricingInit = abi.encodeWithSelector(
        //     BunkerPricing.initialize.selector,
        //     deployer
        // );
        //
        // TransparentUpgradeableProxy pricingProxy = new TransparentUpgradeableProxy(
        //     address(pricingImpl),
        //     address(proxyAdmin),
        //     pricingInit
        // );
        //
        // BunkerPricing pricing = BunkerPricing(address(pricingProxy));
        // console.log("BunkerPricing implementation:", address(pricingImpl));
        // console.log("BunkerPricing proxy:         ", address(pricingProxy));

        // For now, deploy directly (non-upgradeable).
        BunkerPricing pricing = new BunkerPricing(deployer);
        console.log("BunkerPricing (direct):", address(pricing));

        // ═══════════════════════════════════════════════
        //  Step 6: Deploy Timelock and grant roles
        // ═══════════════════════════════════════════════
        address[] memory proposers = new address[](1);
        proposers[0] = deployer;
        address[] memory executors = new address[](1);
        executors[0] = deployer;
        BunkerTimelock timelock = new BunkerTimelock(
            24 hours,
            proposers,
            executors,
            deployer,
            guardian
        );
        console.log("BunkerTimelock deployed:", address(timelock));

        // Grant roles.
        escrow.grantRole(escrow.OPERATOR_ROLE(), operator);
        staking.grantRole(staking.SLASHER_ROLE(), slasher);
        console.log("OPERATOR_ROLE granted to:", operator);
        console.log("SLASHER_ROLE granted to: ", slasher);

        // ═══════════════════════════════════════════════
        //  Step 7: Transfer ownership to Timelock
        // ═══════════════════════════════════════════════
        // TODO: After deployment verification, transfer ownership:
        //   staking.transferOwnership(address(timelock));
        //   escrow.transferOwnership(address(timelock));
        //   pricing.transferOwnership(address(timelock));
        //   proxyAdmin.transferOwnership(address(timelock));
        //
        // Then accept ownership from timelock via scheduled operations.

        vm.stopBroadcast();

        // ═══════════════════════════════════════════════
        //  Summary
        // ═══════════════════════════════════════════════
        console.log("");
        console.log("=== Deployment Summary ===");
        console.log("Token (immutable):", tokenAddr);
        console.log("Staking:          ", address(staking));
        console.log("Escrow:           ", address(escrow));
        console.log("Pricing:          ", address(pricing));
        console.log("Timelock:         ", address(timelock));
        // console.log("ProxyAdmin:       ", address(proxyAdmin));
        console.log("Treasury:         ", treasury);
        console.log("Guardian:         ", guardian);
        console.log("Owner:            ", deployer);
        console.log("");
        console.log("NOTE: Contracts deployed directly (non-upgradeable).");
        console.log("      See TODO comments in this script for proxy migration path.");
        console.log("");
        console.log("Upgrade checklist:");
        console.log("  1. Convert contracts to use initialize() instead of constructors");
        console.log("  2. Replace immutable variables with storage variables");
        console.log("  3. Add Initializable + UUPSUpgradeable base contracts");
        console.log("  4. Uncomment proxy deployment code in this script");
        console.log("  5. Transfer ProxyAdmin ownership to Timelock");
    }
}
