# Moltbunker Proxy Upgrade Strategy

## Current State

The Moltbunker protocol contracts are deployed as **non-upgradeable** direct deployments. This directory contains proxy infrastructure prepared for a future migration to upgradeable contracts.

### Why Not Upgradeable Yet

The current contracts use `immutable` variables in their constructors:

| Contract | Immutable Variables |
|----------|-------------------|
| `BunkerStaking` | `IERC20 public immutable token`, `IERC20 public immutable rewardsToken` |
| `BunkerEscrow` | `IERC20 public immutable token` |
| `BunkerPricing` | None (already compatible) |
| `BunkerToken` | N/A (tokens should never be upgradeable) |

Immutable variables are stored in the **bytecode** of the implementation contract, not in storage. When a proxy uses `delegatecall`, it shares the proxy's storage but executes the implementation's bytecode. This means immutable values work correctly behind proxies, but they cannot be changed during an upgrade since they are embedded in the new implementation's bytecode at deploy time.

The real blocker is that current contracts use **constructors** instead of **initializer functions**. Proxies require `initialize()` functions because constructors only run once during implementation deployment and do not execute in the proxy's context.

## Upgrade Path

### Phase 1: Refactor Contracts (Required)

Each contract needs these changes:

**BunkerStaking:**
```solidity
// Before (constructor):
constructor(address _token, address _treasury, address _initialOwner) Ownable(_initialOwner) {
    token = IERC20(_token);
    ...
}

// After (initializer):
function initialize(address _token, address _treasury, address _initialOwner) external initializer {
    __Ownable_init(_initialOwner);
    __AccessControl_init();
    __ReentrancyGuard_init();
    token = IERC20(_token);      // Change from immutable to storage
    rewardsToken = IERC20(_token); // Change from immutable to storage
    ...
}
```

**BunkerEscrow:**
```solidity
function initialize(address _token, address _treasury, address _initialOwner) external initializer {
    __Ownable_init(_initialOwner);
    __AccessControl_init();
    __ReentrancyGuard_init();
    token = IERC20(_token);      // Change from immutable to storage
    ...
}
```

**BunkerPricing:**
```solidity
function initialize(address _initialOwner) external initializer {
    __Ownable_init(_initialOwner);
    // Set default prices in initializer instead of constructor
    ...
}
```

### Phase 2: Deploy Behind Proxies

Use the `TransparentUpgradeableProxy` pattern (OpenZeppelin v5):

1. Deploy each implementation contract (no constructor args)
2. Deploy `BunkerProxyAdmin` (owner = deployer initially)
3. Deploy `TransparentUpgradeableProxy` for each contract, pointing to its implementation
4. Call `initialize()` via the proxy constructor's `data` parameter
5. Transfer `ProxyAdmin` ownership to `BunkerTimelock`

See `DeployUpgradeable.s.sol` for the full deployment template with commented-out proxy code.

### Phase 3: Production Migration

For existing deployments, state migration is required:

1. Deploy new initializable implementations
2. Deploy proxies with migrated state
3. Update all contract references (escrow -> staking, etc.)
4. Verify state integrity
5. Transfer admin ownership to timelock

## Architecture

```
User/Requester
      |
      v
TransparentUpgradeableProxy (BunkerStaking)  --delegatecall-->  BunkerStaking Implementation V1
TransparentUpgradeableProxy (BunkerEscrow)   --delegatecall-->  BunkerEscrow Implementation V1
TransparentUpgradeableProxy (BunkerPricing)  --delegatecall-->  BunkerPricing Implementation V1
      |
      v (admin calls only)
BunkerProxyAdmin  <--owned by-->  BunkerTimelock
```

### Upgrade Flow

1. Develop and audit new implementation (V2)
2. Propose upgrade via `BunkerTimelock` (24-hour minimum delay)
3. Guardian can cancel malicious proposals
4. After timelock delay, execute: `proxyAdmin.upgradeAndCall(proxy, newImpl, data)`
5. Proxy now delegates to V2 while preserving all storage

## Files

| File | Purpose |
|------|---------|
| `BunkerProxy.sol` | ERC1967 proxy wrapper (for direct proxy deployments) |
| `BunkerProxyAdmin.sol` | Admin contract managing all protocol proxies |
| `../../script/DeployUpgradeable.s.sol` | Deployment template with proxy migration TODOs |

## Design Decisions

- **TransparentProxy over UUPS**: Admin logic is separated from implementation, reducing implementation surface area and upgrade bugs. The `ProxyAdmin` contract is the only address that can call upgrade functions.
- **Single ProxyAdmin**: One admin manages all protocol proxies, simplifying governance.
- **Timelock ownership**: ProxyAdmin is owned by BunkerTimelock, ensuring upgrades go through governance with a delay.
- **Token stays non-upgradeable**: `BunkerToken` (ERC-20) remains immutable. Token contracts should not be upgradeable to maintain trust.
- **Storage gap convention**: When converting, each contract should include `uint256[50] private __gap` for future storage expansion.
