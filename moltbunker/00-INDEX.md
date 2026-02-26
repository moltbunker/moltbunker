# Moltbunker — Production Plan

*Complete blueprint for building a production-ready decentralized containerized compute network.*

*Generated: February 9, 2026 | Updated: February 12, 2026*

---

## Implementation Status

| Section | Plan | Code | Status |
|---------|------|------|--------|
| 01 Architecture | DONE | DONE | System architecture fully implemented |
| 02 Gap Analysis | DONE | ~85% fixed | Most critical/high gaps addressed |
| 03 Smart Contracts | DONE | DONE | 9 contracts, 880 Foundry tests, deployment scripts |
| 04 Security | DONE | DONE | TLS 1.3, cert pinning, seccomp, encryption, key rotation |
| 05 Testing | DONE | DONE | 150+ test files, E2E, fuzz, localnet, Colima |
| 06 Pricing & Economics | DONE | DONE | Pricing calculator, staking tiers, profitability |
| 07 Platform | DONE | DONE | Linux setup, macOS Colima, systemd, logrotate |
| 08 Production Roadmap | DONE | Phase 1 done | Foundation complete, testnet next |
| 09 Network Strategy | DONE | DONE | Provider tiers, bootstrap, infrastructure |
| 10 Deployment Infra | DONE | TODO | Domain, Cloudflare, API serving, dashboard |
| 11 Wallet Integration | DONE | TODO | OnchainKit, Base chain, SIWE auth, contract interactions |
| 12 Web Terminal | DONE | TODO | E2E encrypted browser-to-container exec, wallet-derived keys |
| 13 Code Audit | DONE | 12 fixed | 12 bugs fixed across 4 phases, 8 remaining |
| 14 Molts (WASM) | DONE | DONE | Dual-engine (containers + Molts via wazero + Deno JS/TS), macOS first-class |
| 15 P0 Services | DONE | DONE | Object Storage, Proxy, Crawl, Agent — wired into daemon, 177 tests |

---

## Quick Start

1. Read [Architecture Overview](01-architecture/overview.md) to understand the system
2. Review [Gap Analysis](02-gap-analysis/overview.md) to see what needs fixing (75 gaps identified)
3. Follow the [Production Roadmap](08-production-roadmap/overview.md) for phased execution
4. Track progress with the [Production Checklist](08-production-roadmap/checklist.md) (110 items)

---

## Document Map

### 01 — Architecture (`93K`) — DONE

System design — how Moltbunker works end-to-end.

| Document | Description |
|----------|-------------|
| [overview.md](01-architecture/overview.md) | Full system architecture: components, data flow, node roles, container lifecycle, 3-region redundancy, state management |
| [p2p-network.md](01-architecture/p2p-network.md) | P2P layer: libp2p production config, Kademlia DHT, GossipSub v1.1, NAT traversal, peer scoring, QUIC transport |
| [container-runtime.md](01-architecture/container-runtime.md) | Container layer: containerd integration, IPFS image distribution, imgcrypt encryption, cgroups v2, health monitoring, Linux vs macOS |

**Implementation:** All components implemented in `internal/daemon/`, `internal/p2p/`, `internal/runtime/`, `internal/redundancy/`.

### 02 — Gap Analysis (`30K`) — 85% RESOLVED

What's broken, what's missing, what needs hardening.

| Document | Description |
|----------|-------------|
| [overview.md](02-gap-analysis/overview.md) | 75 gaps identified (17 critical, 25 high, 22 medium, 11 low) with priority matrix, effort estimates, and 5-phase implementation plan |

**Implementation:** Most critical and high-priority gaps resolved:
- Key encryption at rest
- Message signing + versioning
- Graceful shutdown
- Rate limiting + connection limits
- Peer banning + address book persistence
- Certificate pinning store
- Container health checks + cleanup + image GC
- Prometheus metrics + Grafana dashboards
- Payment service wiring (escrow gating, staking checks)
- TLS cert rotation + upgrade mechanism
- Remaining: State persistence (bbolt), SEV-SNP attestation, formal security audit

### 03 — Smart Contracts (`140K`) — DONE

Full contract system for Base L2 — BUNKER token, staking, escrow, pricing.

| Document | Description |
|----------|-------------|
| [overview.md](03-smart-contracts/overview.md) | Contract architecture, interaction diagram, admin wallet config, multi-sig plan, gas optimization, deployment order |
| [bunker-token.md](03-smart-contracts/bunker-token.md) | ERC-20 with Permit (EIP-2612), burn mechanism, 1B supply cap, full Solidity implementation |
| [staking.md](03-smart-contracts/staking.md) | 5-tier staking (500–250K BUNKER), 14-day unbonding, beneficiary support, slashing (80% burn / 20% treasury), full Solidity |
| [escrow.md](03-smart-contracts/escrow.md) | Secure escrow with access control, 5% protocol fee, progressive release, disputes, refunds, full Solidity |
| [pricing.md](03-smart-contracts/pricing.md) | Resource pricing (CPU/memory/storage/network/GPU), admin-configurable, multipliers, oracle-ready, full Solidity |
| [foundry-setup.md](03-smart-contracts/foundry-setup.md) | Complete Foundry project: foundry.toml, 6 test suites, fuzz tests, invariant tests, deployment scripts, Anvil multi-account testing |

**Implementation:** 5 Solidity contracts in `contracts/src/`, 7 test files in `contracts/test/`, deployment scripts in `contracts/script/`. Go ABI bindings aligned with Solidity interfaces. Payment service wired into daemon with escrow gating and staking checks. BunkerTimelock added for governance.

### 04 — Security (`179K`) — DONE

Defense-in-depth security architecture — transport, identity, containers, threats.

| Document | Description |
|----------|-------------|
| [overview.md](04-security/overview.md) | Security layers, threat model summary, trust boundaries, attack surfaces, security principles, audit checklist |
| [transport.md](04-security/transport.md) | TLS 1.3 mTLS, certificate pinning (TOFU + DHT), libp2p Noise, QUIC security, Tor hidden services, rate limiting, connection gating |
| [identity.md](04-security/identity.md) | Ed25519 keys, key encryption at rest, X25519 key exchange, wallet management, key rotation, peer ID derivation |
| [container-security.md](04-security/container-security.md) | Seccomp, AppArmor/SELinux, namespaces, cgroups v2, encrypted images (imgcrypt), encrypted volumes, rootless mode, image signing |
| [threat-model.md](04-security/threat-model.md) | STRIDE analysis (30 threats), Sybil/Eclipse/DHT/replay/MITM prevention, container escape, smart contract vectors, Tor risks, DoS vectors, risk matrix, incident response |

**Implementation:** Full security stack in `internal/security/` (encryption, cert pinning, seccomp, deployment encryption, MITM protection), `internal/identity/` (key management, key encryption at rest, cert rotation, wallet), `internal/p2p/` (TLS 1.3 transport, peer banning, rate limiting, resource manager).

### 05 — Testing (`188K`) — DONE

Full testing strategy — unit tests to chaos testing, local network, Foundry.

| Document | Description |
|----------|-------------|
| [overview.md](05-testing/overview.md) | Testing pyramid, coverage targets, CI/CD pipeline, mock strategy, performance benchmarks |
| [unit-tests.md](05-testing/unit-tests.md) | Per-package test plans with Go code: payment, security, identity, p2p, daemon, redundancy — mock interfaces, fuzz targets |
| [e2e-tests.md](05-testing/e2e-tests.md) | 10 full E2E scenarios with Go code: network formation, staking, deployment, payment, failover, consensus, threat cloning, snapshots, encrypted deployment |
| [local-network.md](05-testing/local-network.md) | 5-node local network (2 providers, 2 requesters, 1 hybrid), per-node configs, Anvil blockchain, shell scripts, Docker Compose, macOS/Linux setup |
| [foundry-tests.md](05-testing/foundry-tests.md) | Smart contract tests: Token/Staking/Escrow/Pricing suites, integration tests, fuzz tests, invariant tests, gas reports, fork testing |
| [colima-e2e.md](05-testing/colima-e2e.md) | Colima E2E tests: real container testing on macOS, ContainerRuntime interface prerequisite, 8 test scenarios (lifecycle, resources, health, deployment flow), test harness |

**Implementation:** 150+ test files across codebase. E2E scenarios in `tests/e2e/scenarios/` (deploy lifecycle, staking, escrow, gossip convergence, replica failure, wallet). Colima tests in `tests/e2e/colima/`. Fuzz tests for P2P messages and encryption. Localnet tests in `tests/localnet/`. Smoke tests in `tests/e2e/smoke/`. Integration tests in `tests/integration/`. Mock infrastructure in `tests/mocks/`.

### 06 — Pricing & Economics (`49K`) — DONE

BUNKER token economics, resource pricing, staking tiers.

| Document | Description |
|----------|-------------|
| [overview.md](06-pricing-economics/overview.md) | Token utility, 1B supply, distribution (40% community, 18% team, 15% investors, 15% treasury), emission schedule, burn mechanics, death spiral prevention |
| [pricing-model.md](06-pricing-economics/pricing-model.md) | Resource pricing table, calculation formula, 5 worked examples, AWS/Akash/Flux comparison, admin wallet config with timelock, price bounds |
| [staking-tiers.md](06-pricing-economics/staking-tiers.md) | 5 tiers with full benefits, 14-day unbonding, delegation, Synthetix-style rewards, anti-gaming measures, ROI estimates |

**Implementation:** `internal/payment/` — PricingCalculator, StakingManager (5 tiers, beneficiary, cooldown, rewards), EscrowManager, PaymentService (wired into daemon). Profitability calculator in `profitability.go`. Smart contract ABIs aligned with Solidity. Mock and real modes for all payment contracts.

### 07 — Platform (`34K`) — DONE

Linux production setup and macOS development + Tier 2 provider environment.

| Document | Description |
|----------|-------------|
| [linux.md](07-platform/linux.md) | Ubuntu/Debian/Fedora setup, containerd, cgroups v2, firewall, Tor, IPFS, systemd service files, AppArmor, monitoring, log rotation |
| [macos.md](07-platform/macos.md) | Colima for containerd, **Tier 2 provider support**, Homebrew deps, Tor/IPFS setup, developer workflow, known limitations vs Linux |

**Implementation:** `scripts/setup-linux.sh` (full production setup), systemd services in `configs/` (moltbunkerd.service, ipfs.service), logrotate and journald configs. Colima E2E tests validate macOS provider flow. Cross-platform containerd socket detection in config.

### 08 — Production Roadmap (`33K`) — Phase 1 DONE

Phased plan from prototype to mainnet.

| Document | Description |
|----------|-------------|
| [overview.md](08-production-roadmap/overview.md) | 4 phases over 18–26 weeks: Foundation → Testnet → Beta → Mainnet, team requirements, risk register, dependency graph |
| [checklist.md](08-production-roadmap/checklist.md) | 110-item checklist across 9 categories (networking, security, contracts, containers, testing, monitoring, ops, economics, docs) with priority and owner |

**Phase 1 (Foundation) status:**
- [x] Smart contracts: 9 contracts, 880 Foundry tests, Anvil-tested, Timelock governance
- [x] Deployment scripts: DeployToken + DeployProtocol + DeployLocal (2-step)
- [x] Security hardening: Key encryption, cert pinning, seccomp, message signing, rate limiting, peer banning
- [x] Payment wiring: Escrow gating on deploy, staking checks on providers, payment release on stop
- [x] Test coverage: 150+ test files, E2E/fuzz/localnet/integration
- [x] Operations: systemd, logrotate, Prometheus, Grafana, backup scripts
- [x] Documentation: 13 docs (whitepaper, architecture, operator guide, incident response, etc.)
- [ ] State persistence (bbolt) — not yet implemented
- [ ] Provider tier auto-detection — planned, not implemented

### 09 — Network Strategy — DONE

Infrastructure, provider tiers, memory integrity, and deployment plan.

| Document | Description |
|----------|-------------|
| [overview.md](09-network-strategy/overview.md) | Master network plan: 3-tier provider system, 3-replica model, encryption layers, Hetzner infrastructure, Tor integration, cost analysis |
| [provider-tiers.md](09-network-strategy/provider-tiers.md) | Tier 1 (Linux SEV-SNP), Tier 2 (Linux/macOS with canaries), Tier 3 (dev/mock), macOS provider guide, economic incentives |
| [memory-integrity.md](09-network-strategy/memory-integrity.md) | 5-layer memory detection: canaries, spot checks, software attestation, memguard, economic deterrent. Slashing for violations |
| [infrastructure.md](09-network-strategy/infrastructure.md) | Hetzner deployment: AX162-S (Tier 1), CX43 (bootstrap), CX23 (Tor relay). Setup scripts, monitoring, backup, scaling plan |

**Implementation:** P2P bootstrap with DNS discovery in `internal/p2p/bootstrap.go`. Address book persistence. Peer exchange protocol. Geographic node selection for replication. Connection manager with watermarks. Resource manager for DoS protection. NAT traversal (UPnP, hole punching, relay).

### 10 — Deployment Infrastructure — NEW

Domain architecture, Cloudflare, public API serving, and authentication.

| Document | Description |
|----------|-------------|
| [overview.md](10-deployment-infrastructure/overview.md) | 3 subdomains (moltbunker.com for landing+app, api, status), Cloudflare proxy config, nginx origin, 3 auth methods (wallet SIWE, API keys, inline signatures), bootstrap DNS, cost analysis. Docs at /docs, app at /app |
| [ovh-server-setup.md](10-deployment-infrastructure/ovh-server-setup.md) | OVH main node setup: systemd services, containerd config, socket permissions, root daemon, 6 issues fixed, E2E test results (17/17), deploy commands |

**Implementation status:**
- [x] HTTP API server (`cmd/api/`, `internal/api/`)
- [x] Auth: API keys (`mb_live_*`) + wallet signatures + session tokens
- [x] Rate limiting, CORS, WebSocket support
- [x] Daemon bridge (API → daemon via Unix socket)
- [x] Localnet script with Anvil contract deployment + config generation
- [x] Landing page + docs (deployed on Cloudflare Pages at `bunker/web`)
- [x] OVH main node deployed — daemon + API + containerd running, E2E verified
- [x] Container lifecycle working: deploy, run, logs, stop, delete (17/17 E2E pass)
- [ ] /app routes (wallet connect, dashboard) — not yet built
- [ ] Cloudflare DNS records — not yet configured
- [ ] nginx origin config — not yet deployed

### 11 — Wallet Integration (Base Chain) — NEW

OnchainKit wallet connection, SIWE authentication, and contract interactions.

| Document | Description |
|----------|-------------|
| [overview.md](11-wallet-integration/overview.md) | OnchainKit (Coinbase official), wagmi/viem setup, Base chain config, SIWE auth flow, Smart Wallet + EOA support, contract reads/writes (staking, escrow, token), gas sponsorship, /app route structure |

**Implementation status:**
- [ ] Install OnchainKit + wagmi + viem + tanstack-query
- [ ] OnchainProvider setup (Vite, not Next.js)
- [ ] WalletButton in Header
- [ ] SIWE auth flow (challenge → sign → verify)
- [ ] /app routes with auth guard
- [ ] Contract interactions (staking, escrow, token balance)

### 13 — Code Audit (Phases 10-13) — IN PROGRESS

Systematic codebase audit for bugs, races, leaks, and logic errors.

| Document | Description |
|----------|-------------|
| [overview.md](13-code-audit/overview.md) | 12 fixes across 4 phases, remaining findings, rejected false positives |

**Status:** 12 bugs fixed (4 Critical, 5 High, 3 Medium). 8 remaining candidates identified for future phases.

### 14 — Molts (Serverless Functions) — NEW

**Molts** are Moltbunker's serverless functions — lightweight, encrypted, per-invocation compute powered by WebAssembly (wazero).

| Document | Description |
|----------|-------------|
| [overview.md](14-wasm-runtime/overview.md) | Full feasibility analysis: security, runtime selection (wazero), industry precedent, architecture, provider tiers, subdomain routing, pricing, language support, limitations, WASI roadmap, implementation plan |

**Key decisions:**
- **Branding**: "Molts" — from **molt**bunker (a quick, lightweight transformation)
- **Runtime**: wazero (pure Go, zero CGO, ARM64 macOS native)
- **Architecture**: Dual-engine — containers for services/GPU, Molts for serverless functions
- **macOS nodes**: Tier 2 Standard for Molt workloads (no Colima required)
- **Ingress**: Zero changes — Molts expose local HTTP listener, same gossip discovery
- **Pricing**: Per-invocation (100ms minimum) vs per-hour for containers
- **CLI**: `moltbunker molt deploy`, `molt list`, `molt logs`, `molt stop`
- **Existing readiness**: `WorkloadTypeFunction`, `AcceptFunctions`, `TriggerConfig`, `ScaleToZero`, `MinimumFunctionMillis` already defined (unwired)

**Implementation status:**
- [x] Research: security analysis, runtime comparison, industry survey
- [x] Subdomain registry: BunkerRegistry.sol for `<name>.moltbunker.dev` routing
- [x] `internal/molt/` package (runtime, module cache, HTTP handler, host functions, metrics)
- [x] WASM host functions: 12 functions (result_size/read, error_message, random_bytes, time_now_ms, http_request, storage_put/get/delete/list, crawl_page)
- [x] HostServices layer: shared between WASM and Deno runtimes (host_services.go, host_http.go, host_storage.go, host_crawl.go)
- [x] Deno JS/TS runtime: `internal/jsruntime/` — worker pool, stdio JSON-RPC, embedded bindings.ts
- [x] Config + runtime detection (JSRuntimeConfig, capability flags)
- [x] CLI: `moltbunker molt deploy/list/logs/stop/invoke`
- [x] Per-invocation pricing + prepaid credits (MoltCreditManager)
- [x] 73 tests (62 molt + 11 jsruntime)

### 15 — P0 Services — NEW

4 platform services wired into the daemon, advertised via P2P announce, exposed through REST API.

| Document | Description |
|----------|-------------|
| [overview.md](15-p0-services/overview.md) | Service architecture, daemon wiring, REST routes, config, P2P capability advertisement, security, pricing |

**Key decisions:**
- Services are **capabilities**, not runtimes — no new tiers needed
- Advertised via `AnnouncePayload` fields (same pattern as Molt)
- Conditional initialization: each service gated on `cfg.<Service>.Enabled`
- Proxy is the only service with its own network listeners (SOCKS5 + HTTP)

**Implementation status:**
- [x] Object Storage: StorageEngine + REST handler + bbolt metadata
- [x] Proxy: SOCKS5 + HTTP proxy + session tracking + Tor support
- [x] Web Crawling: Scheduler + robots.txt + job management
- [x] Agent Runtime: Framework dispatch + memory store
- [x] Daemon wiring: all 4 services initialized in `cmd/daemon/main.go`
- [x] P2P announce: capability flags in `AnnouncePayload`
- [x] REST routes: registered via `Set<X>Handler` pattern
- [x] Types: `ServiceCapabilities` + per-service types in `service_types.go`

### 12 — Web Terminal — NEW

E2E encrypted browser-to-container exec with wallet-derived keys.

| Document | Description |
|----------|-------------|
| [overview.md](12-web-terminal/overview.md) | Architecture: xterm.js → WebSocket → Cloudflare → API → P2P → containerd exec. Exec agent, API endpoints, UX flow, component map |
| [security.md](12-web-terminal/security.md) | 10-threat analysis: session theft, CSWSH, replay, provider eavesdrop, MITM, container escape. Two-factor auth (session + wallet sign). Risk matrix |
| [encryption.md](12-web-terminal/encryption.md) | E2E encryption protocol: wallet sign → HKDF → master_kek → per-container exec_key → per-session AES-256-GCM. Deploy-time key seeding, RFC 6979, Web Crypto API |
| [implementation.md](12-web-terminal/implementation.md) | 4 new message types, 6 new Go files, 5 new TS files, modified files, pseudocode for handler + frontend, testing plan, build sequence |

**Key design decisions:**
- No SSH daemon in containers — uses containerd exec (no extra attack surface)
- Wallet signature required per exec session (prevents session token theft → shell access)
- Pre-shared secret (no runtime key exchange → no MITM)
- RFC 6979 deterministic signatures → same wallet = same keys on any device
- No exec for API keys (`mb_live_*`) — interactive shell is a human-only operation

**Implementation status:**
- [ ] P2P exec message types (ExecOpen/Data/Resize/Close)
- [ ] Interactive PTY in containerd (extend container_exec.go)
- [ ] P2P bidirectional exec stream
- [ ] WebSocket exec endpoint + session manager
- [ ] Exec agent binary (in-container encryption relay)
- [ ] Frontend: xterm.js + Web Crypto AES-256-GCM
- [ ] Deploy-time exec key seeding
- [ ] E2E tests

---

## Research Library

Deep research on competitors and technical patterns — referenced throughout.

| File | Location | Description |
|------|----------|-------------|
| Competitive Landscape | [research/competitive-landscape.md](../research/competitive-landscape.md) | 24 projects surveyed with GitHub URLs and status |
| Deep Dive Analysis | [research/decentralized-computing-landscape-analysis.md](../research/decentralized-computing-landscape-analysis.md) | NKN, SONM, Golem, Akash, Flux, Bacalhau technical analysis |
| P2P Networking | [research/01-p2p-networking.md](../research/01-p2p-networking.md) | libp2p, DHT, GossipSub, NAT traversal patterns |
| Security & Encryption | [research/02-security-encryption.md](../research/02-security-encryption.md) | TLS, keys, zero-trust, Sybil prevention, real-world incidents |
| Container Orchestration | [research/03-container-orchestration.md](../research/03-container-orchestration.md) | containerd, scheduling, failover, IPFS distribution |
| Token Economics | [research/04-token-economics-staking.md](../research/04-token-economics-staking.md) | Escrow, pricing, incentives, Solidity patterns |
| Production Operations | [research/05-production-operations.md](../research/05-production-operations.md) | Bootstrap, testing, monitoring, scaling, disaster recovery |

---

## Key Numbers

| Metric | Value |
|--------|-------|
| Total plan documentation | 900K+ across 43 files |
| Gaps identified | 75 (17 critical) — ~85% resolved |
| Production checklist items | 110 |
| Smart contracts | 9 (Token, Staking, Escrow, Pricing, Timelock, Delegation, Reputation, Verification, Registry) — **880 Foundry tests** |
| Go test files | 150+ (1,469 test functions) |
| E2E test scenarios | 16+ (deploy, staking, escrow, gossip, replica, wallet, Colima) |
| Staking tiers | 5 (500–250,000 BUNKER) |
| P0 services | 4 (Storage, Proxy, Crawl, Agent) — 177 tests, wired into daemon |
| Provider tiers | 3 (Confidential, Standard, Development) |
| Subdomains | 3 (moltbunker.com, api, status) — app + docs served at /app and /docs |
| Testnet infrastructure | 3–4 Hetzner nodes (~€15–221/mo) |

---

## Critical Path

```
Phase 1: Foundation (Weeks 1-6) ████████████████████ 95% DONE
├── [DONE] Smart contracts: 9 contracts, 880 Foundry tests, Anvil-tested
├── [DONE] Deployment scripts: DeployToken + DeployProtocol + DeployLocal
├── [DONE] Security: key encryption, cert pinning, seccomp, message signing
├── [DONE] Payment wiring: escrow gating, staking checks, payment release
├── [DONE] Testing: 150+ test files, E2E, fuzz, localnet, integration
├── [DONE] Operations: systemd, logrotate, Prometheus, Grafana, backup
├── [DONE] Documentation: 13 docs + 37 plan files
├── [TODO] State persistence (bbolt)
└── [TODO] Provider tier auto-detection

Phase 2: Testnet (Weeks 5-10)
├── Deploy 3 Hetzner bootstrap nodes (EU/US/AP)
├── Configure Cloudflare DNS + proxy for api.moltbunker.com
├── Deploy API server with nginx + Cloudflare origin cert
├── 3-replica model: 1 active (Tier 1) + 2 warm (Tier 2)
├── Smart contracts on Base Sepolia
├── Build /app routes (OnchainKit wallet, dashboard pages)
├── Web Terminal: E2E encrypted browser exec (xterm.js + wallet-derived keys)
└── 10 E2E scenarios passing on real network

Phase 3: Beta (Weeks 9-16)
├── Public testnet (20+ external nodes)
├── macOS Tier 2 providers (Colima setup guide)
├── SEV-SNP attestation verification
├── Professional security audit
├── Bug bounty program
└── Production monitoring (Prometheus/Grafana dashboards)

Phase 4: Mainnet (Weeks 15-20)
├── Base mainnet deployment
├── 5 bootstrap nodes across 5 regions
├── Landing page + docs site live
├── macOS provider onboarding campaign
└── Incident response & on-call
```

---

## How to Use This Plan

1. **Start with gaps**: Read `02-gap-analysis/overview.md` — most critical gaps are now resolved
2. **Smart contracts**: Follow `03-smart-contracts/foundry-setup.md` to set up Foundry, deploy locally with Anvil
3. **Run tests**: `go test ./...` (unit) and `go test -tags e2e ./tests/e2e/...` (E2E)
4. **Local network**: Use `scripts/localnet.sh start` — deploys contracts to Anvil, starts daemon + API with real payments
5. **Deploy infrastructure**: Follow `10-deployment-infrastructure/overview.md` for Cloudflare + Hetzner setup
6. **Track with checklist**: Check off items in `08-production-roadmap/checklist.md`
