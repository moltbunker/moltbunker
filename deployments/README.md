# Contract address manifest

`addresses.json` is the **single source of truth** for moltbunker contract
addresses across every consumer: the Go daemon, the web dapp, the web-admin
panel, and the Python SDK.

Contract addresses are **public on-chain facts**, not secrets. They are
committed to source so consumers can resolve them at build/runtime without
environment-variable injection. **Never** put private keys, mnemonics, or API
secrets in this file — those live in gitignored keystores / `.env.local`.

## Schema

```jsonc
{
  "chains": {
    "<chainId>": {                  // stringified EVM chain id, e.g. "84532"
      "chainName": "Base Sepolia",
      "rpcUrl": "https://sepolia.base.org",
      "contracts": {                // all 10 names required, every chain
        "token": "0x…",
        "staking": "0x…",
        "escrow": "0x…",
        "pricing": "0x…",
        "timelock": "0x…",          // = daemon governance_address
        "delegation": "0x…",
        "reputation": "0x…",
        "verification": "0x…",
        "registry": "0x…",          // = daemon subdomain_registry_address
        "slashing": "0x…"
      },
      "deployedAt": "2026-02-26T00:00:00Z",
      "note": "freeform"
    }
  }
}
```

The zero address (`0x000…0`) is permitted and means "not yet deployed". The
daemon's own config validation rejects a zero address when `mock_payments:
false`, so a not-deployed chain fails fast at startup rather than silently.

## Consumers

| Consumer | How it reads the manifest |
|---|---|
| **Go daemon** | `internal/deployment` embeds `addresses.json` (via the root `deployments` package `//go:embed`) and `internal/config` fills empty `economics.*_address` fields from it, keyed by `chain_id`. Operator YAML overrides always win. |
| **web / web-admin** | `tools/gen-addresses` emits a typed `generated-addresses.ts` module (`CHAIN_CONFIGS`, `getContracts()`). Wiring those files into the `web/` and `web-admin/` repos is done in **their** PRs — this repo only owns the generator. |
| **Python SDK** | A small stdlib generator emits `chains.py`. Lives in the SDK repo; this repo only owns the manifest + Go/TS generator. |

The TS / env / Python emitters are **scaffolded** here: `make gen-addresses`
only writes the in-repo `configs/addresses-fragment.yaml` by default. To emit the
cross-repo TS / env files, point the output flags at a sibling checkout:

```sh
make gen-addresses \
  ADDR_OUT_WEB=../web/src/lib/generated-addresses.ts \
  ADDR_OUT_ADMIN=../web-admin/src/lib/generated-addresses.ts \
  ADDR_OUT_ADMIN_ENV=../web-admin/.env.example
```

## Mainnet cutover (the whole point)

1. Edit the `"8453"` (Base Mainnet) block in `addresses.json` with the real
   deployed addresses.
2. Run `make gen-addresses` (regenerates `configs/addresses-fragment.yaml`; add
   the cross-repo flags above to refresh web/web-admin in their checkouts).
3. Update `TestContractSetZeroOnMainnet` in
   `internal/deployment/addresses_test.go` (it documents the all-zero invariant).
4. Commit. One PR propagates the new addresses to every consumer.

A daemon operator targeting mainnet then needs only `chain_id: 8453` +
`mock_payments: false`; addresses resolve automatically from the embedded
manifest.
