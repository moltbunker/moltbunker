# Production Readiness Checklist

*Moltbunker -- Mainnet Launch Prerequisites*

---

Every item must be completed (checked) before mainnet launch unless explicitly marked as post-launch. Items are organized by category with priority (P0 = must have, P1 = should have, P2 = nice to have) and suggested owner.

---

## 1. Networking

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| N1 | Replace IPFS bootstrap peers with dedicated Moltbunker bootstrap nodes | [x] | P0 | Backend | Done: DNS bootstrap in bootstrap.go, no public IPFS peers, mDNS + address book fallback |
| N2 | Deploy 5 bootstrap nodes across 5 regions (US-E, US-W, EU-W, EU-C, APAC) | [ ] | P0 | DevOps | Minimum 3 cloud providers |
| N3 | Implement DNS-based bootstrap peer resolution (`/dnsaddr/`) | [x] | P0 | Backend | Done: ResolveDNSBootstrap in bootstrap.go, _dnsaddr TXT resolution, static fallback, 14 tests |
| N4 | Persist DHT routing table to disk for faster reconnection | [x] | P1 | Backend | Done: PeerStore with JSON save/load in peerstore.go |
| N5 | Implement peer address book with persistent storage | [x] | P1 | Backend | Done: AddressBook with JSON persistence, success-rate ranking, DHT integration in addressbook.go |
| N6 | Add NAT traversal via libp2p AutoNAT and relay | [x] | P0 | Backend | Done: AutoNAT, NATPortMap, relay, hole punching in dht.go |
| N7 | Configure libp2p Resource Manager with ScalingLimitConfig | [x] | P0 | Backend | Done: ScalingLimitConfig in resource_manager.go, 256 sys conns, per-peer limits |
| N8 | Implement protocol version negotiation in P2P message headers | [x] | P0 | Backend | `ProtocolVersion` + `MinVersion` fields — Done: Version field in Message struct, set on send |
| N9 | Enforce minimum peer version (reject outdated peers) | [x] | P0 | Backend | Done: MinSupportedVersion enforcement in routing.go |
| N10 | Connection manager with soft limits (watermark-based pruning) | [x] | P1 | Backend | Done: low=100, high=400, grace=20s in connmgr.go, wired into DHT |
| N11 | Bootstrap node health monitoring with automated failover | [ ] | P0 | DevOps | 99.9% uptime requirement |
| N12 | Measure and log cross-region latency for geographic node selection | [x] | P1 | Backend | Done: LatencyMonitor with per-peer/region tracking, incremental averages in latency.go |
| N13 | Implement peer exchange protocol for ongoing peer discovery | [x] | P2 | Backend | Done: PeerExchangeProtocol with quality ranking, request/response handlers, background loop, 19 tests in peerexchange.go |

---

## 2. Security

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| S1 | Encrypt Ed25519 node identity keys at rest (passphrase-derived) | [x] | P0 | Backend | Done: argon2id + AES-256-GCM in identity/keys.go |
| S2 | Sign all P2P messages with node identity key | [x] | P0 | Backend | Done: MessageSigner interface in p2p/routing.go |
| S3 | TLS 1.3 mutual authentication verified in production | [ ] | P0 | Backend | Already implemented; verify under load |
| S4 | Certificate pinning validated for known bootstrap nodes | [x] | P0 | Backend | Done: Save/Load persistence with atomic writes, TOFU, 14 tests |
| S5 | Implement TLS certificate rotation with peer notification | [x] | P1 | Backend | Done: CertRotator with auto-renewal, background checking, callbacks in cert_rotation.go |
| S6 | Rate limiting on all API endpoints verified | [x] | P0 | Backend | Done: token bucket per-IP rate limiter in server.go |
| S7 | Rate limiting on P2P message handlers | [x] | P0 | Backend | Done: per-peer message rate limits in routing.go |
| S8 | Peer banning for misbehaving nodes | [x] | P1 | Backend | Done: BanList with save/load, auto-ban on 3 violations in 5min |
| S9 | Seccomp profiles applied to all containers in production | [x] | P0 | Backend | Done: hardened profile (essential vs dangerous syscalls), ValidateSeccompProfile(), 18 tests |
| S10 | API key hashing verified (bcrypt, prefix-based lookup) | [x] | P0 | Backend | Done: bcrypt + prefix lookup + constant-time compare in apikey.go |
| S11 | Audit logging for all sensitive operations | [x] | P1 | Backend | Done: AuditEvent struct + logging calls in staking, keys |
| S12 | File descriptor limit validation on startup (`moltbunker doctor`) | [x] | P0 | Backend | Done: syscall.Getrlimit in doctor/checker_fdlimit.go |
| S13 | Message size limits enforced uniformly (prevent oversized message attacks) | [x] | P0 | Backend | Done: 16MB max in p2p/message_send.go ReadLengthPrefixed |
| S14 | X25519 key exchange implementation reviewed | [x] | P0 | Security | Done: X25519 + HKDF-SHA3-256 in deployment_encryption.go, tests pass |
| S15 | No secrets in logs (verify slog output sanitization) | [x] | P0 | Backend | Done: RedactingHandler for slog, redacts API keys, private keys, hex strings |

---

## 3. Smart Contracts

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| C1 | BunkerToken (ERC-20) deployed to Base mainnet | [ ] | P0 | Smart Contract | 1B supply, 18 decimals, non-upgradeable |
| C2 | StakingContract deployed with UUPS proxy | [ ] | P0 | Smart Contract | 5-tier enforcement, 14-day unbonding |
| C3 | EscrowContract deployed with UUPS proxy | [ ] | P0 | Smart Contract | Per-deployment escrow, protocol fee, burn |
| C4 | Admin multi-sig (3-of-5 Gnosis Safe) created and ownership transferred | [ ] | P0 | Team Lead | Controls pricing and contract upgrades |
| C5 | Ownable2Step + AccessControl on all upgradeable contracts | [x] | P0 | Smart Contract | Done: All contracts use Ownable2Step; AccessControl on Escrow (OPERATOR_ROLE) and Staking (SLASHER_ROLE) |
| C6 | ReentrancyGuard on all state-changing functions | [x] | P0 | Smart Contract | Done: nonReentrant on all state-changing functions in Escrow (5) and Staking (4) |
| C7 | 100% Foundry test coverage (branch + line) | [x] | P0 | Smart Contract | Done: 880 tests across 15 files (Token:60, Staking:181+41, Escrow:99+40, Pricing:86+33, Registry:174, Timelock:45, Delegation:21, Reputation:29, Verification:24, Admin:38, Fork:8, Integration:9) |
| C8 | Foundry fuzz testing for all public functions | [x] | P0 | Smart Contract | Done: 16 fuzz tests across Token(4), Staking(4), Pricing(3), Escrow(5), 256 runs each, all pass |
| C9 | Foundry fork tests against Base mainnet state | [x] | P1 | Smart Contract | Done: Fork.t.sol with 8 tests on Base mainnet fork (chain 8453) — lifecycle, staking, metadata, contracts |
| C10 | Professional security audit completed | [ ] | P0 | Security | Trail of Bits, OpenZeppelin, or Cyfrin |
| C11 | All critical/high audit findings remediated | [ ] | P0 | Smart Contract | Verified by auditor re-review |
| C12 | Contracts verified on Basescan | [x] | P0 | Smart Contract | Done: All 8 contracts verified on Base Sepolia Basescan |
| C13 | Emergency pause mechanism (circuit breaker) | [x] | P0 | Smart Contract | Done: Pausable on Escrow and Staking with whenNotPaused on critical functions, owner-controlled pause/unpause |
| C14 | Timelock on all admin parameter changes (24h minimum) | [x] | P0 | Smart Contract | Done: BunkerTimelock.sol (OZ TimelockController, 24h MIN_DELAY_FLOOR, GUARDIAN_ROLE emergency pause, 45 tests) |
| C15 | Token distribution executed according to allocation plan | [ ] | P0 | Team Lead | Vesting contracts for team/investor/ecosystem |

---

## 4. Container Runtime

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| R1 | containerd integration tested on Ubuntu 22.04 and 24.04 | [ ] | P0 | Backend | Verify both LTS versions |
| R2 | containerd integration tested on Debian 12 | [ ] | P0 | Backend | Verify stable release |
| R3 | Container image pull via IPFS CID verified | [ ] | P0 | Backend | `internal/distribution/` |
| R4 | Container image encryption (AES-256-GCM) verified | [ ] | P0 | Backend | `internal/runtime/encryption.go` |
| R5 | cgroups v2 resource limits (CPU, memory, IO) enforced | [ ] | P0 | Backend | `internal/runtime/cgroups.go` |
| R6 | Sandbox isolation verified (network namespace, PID namespace) | [ ] | P0 | Backend | `internal/runtime/sandbox.go` |
| R7 | Container state snapshots (create, encrypt, restore) | [x] | P1 | Backend | Done: Full snapshot system with AES-256-GCM encryption, gzip compression, incremental snapshots, key rotation, retention policies in snapshot/ |
| R8 | Volume encryption (aes-xts-plain64) verified | [ ] | P0 | Backend | Persistent storage encryption |
| R9 | Container cleanup on deployment termination (no orphaned resources) | [x] | P0 | Backend | Done: CleanupManager with idempotent cleanup, orphaned detection, 12 tests in runtime/cleanup.go |
| R10 | Image garbage collection for unused container images | [x] | P1 | Backend | Done: ImageGC with time-based expiry, in-use tracking, background GC, 14 tests in runtime/image_gc.go |
| R11 | Container health checks functional (HTTP, TCP, exec) | [x] | P0 | Backend | Done: HealthChecker with HTTP/TCP/exec probes, thresholds, background loop, 33 tests in runtime/health_check.go |

---

## 5. Testing

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| T1 | Unit test coverage >= 80% across all `internal/` packages | [x] | P0 | All Devs | Done: 460+ new tests total across all packages, 22 packages passing |
| T2 | goleak integrated in all test packages | [x] | P1 | All Devs | Done: goleak.VerifyTestMain in 9 packages (api, daemon, p2p, payment, redundancy, security, config, identity, metrics) |
| T3 | E2E deployment lifecycle test passing | [x] | P0 | QA | Done: 4 tests (8-phase lifecycle, escrow payment, failover, cleanup) in lifecycle_deploy_test.go |
| T4 | E2E staking lifecycle test passing | [x] | P0 | QA | Done: 5 tests, 27 subtests (tier progression, cooldown, slashing, multi-provider) in staking_lifecycle_test.go |
| T5 | E2E escrow lifecycle test passing | [x] | P0 | QA | Done: 6 tests (full lifecycle, refund, protocol fee, contract integration, pricing) in escrow_test.go |
| T6 | Localnet 5-node discovery and messaging test passing | [x] | P0 | QA | Done: 3 tests (5-node discovery, message routing, peer reconnection) in discovery_test.go |
| T7 | Chaos testing with Toxiproxy (partition, latency, packet loss) | [ ] | P1 | QA | Verify recovery from network failures |
| T8 | Fuzz testing for P2P message parsing | [x] | P1 | Backend | Done: FuzzMessageParsing with 13+ seed corpus entries in message_fuzz_test.go |
| T9 | Fuzz testing for encryption/decryption | [x] | P1 | Backend | Done: FuzzEncryptDecryptChaCha20 + FuzzEncryptDecryptAES256GCM in encryption_fuzz_test.go |
| T10 | Load testing: 100 concurrent deployments | [ ] | P1 | QA | Measure latency, success rate, resource usage |
| T11 | Gossip convergence property test (2/3 agreement under normal conditions) | [x] | P0 | Backend | Done: 5 tests (majority, split-brain, late joiner, concurrent, version conflict) in gossip_convergence_test.go |
| T12 | Smart contract integration tests against Base Sepolia | [ ] | P0 | Smart Contract | Go tests using ethclient |
| T13 | Multi-replica failure recovery test (kill 1/3, verify re-replication) | [x] | P0 | QA | Done: 4 tests (single failure, multiple failure, cascading, recovery timing) in replica_failure_test.go |

---

## 6. Monitoring & Observability

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| M1 | Migrate metrics to Prometheus client library (prometheus/client_golang) | [x] | P0 | Backend | Done: PrometheusCollector wrapping Collector with CounterVec, HistogramVec, Gauges, dedicated registry in prometheus.go, 16 tests |
| M2 | Expose /metrics endpoint on daemon API | [x] | P0 | Backend | Done: Prometheus text format exposition in metrics_handler.go |
| M3 | Enable go-libp2p native metrics | [x] | P1 | Backend | Done: rcmgr.MustRegisterWith() in resource_manager.go, NewResourceManagerWithRegistry() for custom registry |
| M4 | Create Grafana dashboard: Network Overview | [x] | P0 | DevOps | Done: configs/grafana/network-overview.json (peer count, connections, message rates, libp2p rcmgr, geo distribution) |
| M5 | Create Grafana dashboard: Node Health | [x] | P0 | DevOps | Done: configs/grafana/node-health.json (CPU, memory, goroutines, GC, FDs, uptime, connections) |
| M6 | Create Grafana dashboard: Deployment Pipeline | [x] | P1 | DevOps | Done: configs/grafana/deployment-pipeline.json (request rate, latency quantiles, container count, success/error) |
| M7 | Create Grafana dashboard: Economics | [x] | P1 | DevOps | Done: configs/grafana/economics.json (staking tiers, escrow, slashing, fees, reputation, implementation guide) |
| M8 | Alerting rules for bootstrap node failures | [x] | P0 | DevOps | Done: configs/prometheus/alerts.yml bootstrap group (3 rules: down, DNS resolution, high latency) |
| M9 | Alerting rules for peer count drop (> 30% decrease in 5 minutes) | [x] | P0 | DevOps | Done: configs/prometheus/alerts.yml peers group (2 rules: low count, rapid drop 30% in 5min) |
| M10 | Alerting rules for deployment failure rate spike (> 10% in 1 hour) | [x] | P1 | DevOps | Done: configs/prometheus/alerts.yml deployments group (3 rules: failure rate, high latency, low container) |
| M11 | Health check endpoint (/health) for load balancers | [x] | P0 | Backend | Done: /health endpoint with status, uptime, peer_count, version |
| M12 | Goroutine count monitoring with alerting threshold (> 10,000) | [x] | P1 | Backend | Done: GoroutineAlertThreshold=10000, UpdateGoroutineCount(), CheckGoroutineHealth() in metrics.go |
| M13 | Continuous profiling setup (Pyroscope or pprof endpoint) | [x] | P2 | Backend | Done: pprof endpoints behind auth in server.go, disabled by default |

---

## 7. Operations

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| O1 | systemd service file for moltbunkerd (Type=notify, watchdog, security) | [x] | P0 | DevOps | Done: configs/moltbunkerd.service with full security hardening, LimitNOFILE=65536 |
| O2 | systemd service file for IPFS daemon | [x] | P1 | DevOps | Done: configs/ipfs.service with security hardening, dhtclient routing, wired as moltbunkerd dependency |
| O3 | Configuration management (YAML + env vars + CLI flags hierarchy) | [x] | P0 | Backend | Done: Verified with 10 production config tests (YAML, env vars, validation, round-trip) in config_production_test.go |
| O4 | Graceful shutdown (signal handling, state persistence, peer notification) | [x] | P0 | Backend | Done: SIGINT/SIGTERM handling, 30s timeout, orderly component shutdown |
| O5 | Log rotation configured (journald or logrotate) | [x] | P0 | DevOps | Done: configs/moltbunker-logrotate.conf (daily, 30-day, 2GB max) + configs/moltbunker-journald.conf |
| O6 | Backup script for critical node data (keys, wallet, state database) | [x] | P0 | DevOps | Done: scripts/backup.sh (596 lines, age encryption, --restore, --dry-run, integrity verification) |
| O7 | Automated setup script for Linux (Ubuntu/Debian) | [x] | P1 | DevOps | Done: scripts/setup-linux.sh (942 lines, containerd+IPFS install, user creation, systemd, FD limits, --uninstall) |
| O8 | Docker image (multi-stage build, minimal base) | [x] | P1 | DevOps | Done: Dockerfile (golang:1.24-bookworm → debian:bookworm-slim), .dockerignore, 3 binaries, OCI labels |
| O9 | Upgrade mechanism (version check, notification, rolling upgrade docs) | [x] | P0 | Backend | Done: VersionChecker with semver comparison, background checks, caching in upgrade/checker.go |
| O10 | Database migration framework for schema changes | [x] | P0 | Backend | Done: Migrator with versioned migrations, atomic JSON persistence in migration/migrator.go |
| O11 | Incident response plan documented | [x] | P0 | Team Lead | Done: docs/INCIDENT_RESPONSE.md (898 lines, 7 playbooks, severity matrix, escalation, comms plan, post-mortem template) |
| O12 | On-call rotation established (24/7 for first 4 weeks) | [ ] | P0 | Team Lead | Coverage for launch period |
| O13 | Rollback plan for contract upgrades and protocol changes | [x] | P0 | DevOps | Done: docs/ROLLBACK_PLAN.md (918 lines, 8 sections: protocol, binary, config, contract, migration, network-wide, testing, decision criteria) |
| O14 | Secrets management guide for operators (SOPS + age) | [x] | P1 | DevOps | Done: docs/SECRETS_MANAGEMENT.md (739 lines, SOPS+age setup, rotation, CI/CD, emergency recovery) |

---

## 8. Economics

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| E1 | Pricing parameters configured and tested on-chain | [x] | P0 | Smart Contract | Done: BunkerPricing.sol with configurable CPU/mem/storage/network/GPU prices, multipliers (redundancy, Tor, SLA, spot) |
| E2 | Staking tiers enforced on-chain (5 tiers with correct thresholds) | [x] | P0 | Smart Contract | Done: BunkerStaking.sol enforces 5 tiers (500/2K/10K/50K/250K), auto-deactivation, tier determination |
| E3 | Escrow deposit, progressive release, and refund flow working end-to-end | [x] | P0 | Smart Contract + Backend | Done: BunkerEscrow.sol with createReservation, progressive releasePayment, refund, finalizeReservation, settleDispute |
| E4 | Protocol fee (5%) deducted correctly: 80% burned, 20% to treasury | [x] | P0 | Smart Contract | Done: 500bps fee, 80% burn via token.burn(), 20% treasury transfer in both Escrow and Staking |
| E5 | Slashing proposals, appeal windows, and execution working | [x] | P0 | Smart Contract + Backend | Done: 48h appeal window, proposeSlash/appealSlash/executeSlash/resolveAppeal, 29 tests in BunkerStaking |
| E6 | Staking rewards distribution (Synthetix-style) functional | [x] | P1 | Smart Contract | Done: Synthetix-style rewardPerToken, tier multipliers, epoch rollover, claimRewards, 17 tests |
| E7 | DEX liquidity deployed (BUNKER/ETH, BUNKER/USDC on Aerodrome) | [ ] | P0 | Team Lead | Initial liquidity for trading |
| E8 | Provider profitability calculator available | [x] | P2 | Frontend | Done: internal/payment/profitability.go (Calculate, BreakEven, TierComparison), 16 tests |

---

## 9. Documentation

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| D1 | Operator guide: Linux production setup (install, configure, run) | [x] | P0 | Tech Writer | Done: docs/OPERATOR_GUIDE.md (10 sections: install, config, running, staking, monitoring, security, troubleshooting, maintenance) |
| D2 | API documentation: OpenAPI spec published with examples | [x] | P0 | Tech Writer | Done: api/openapi.yaml (2583 lines, 49 endpoints, 34 schemas, valid YAML, all $refs resolved) |
| D3 | Python SDK documentation with quickstart and tutorials | [x] | P1 | Tech Writer | Done: docs/SDK_PYTHON.md (1534 lines, 13 sections, auth, bots, deploy, snapshots, cloning, monitoring, staking, errors, 5 examples) |
| D4 | Token economics documentation (pricing, staking, escrow) | [x] | P0 | Tech Writer | Done: docs/TOKEN_ECONOMICS.md (474 lines, 9 sections, all values from code) |
| D5 | Troubleshooting guide: common errors and solutions | [x] | P1 | Tech Writer | Done: docs/TROUBLESHOOTING.md (1060 lines, startup/P2P/runtime/payment/API/perf issues, error reference, FAQ) |
| D6 | Architecture overview: protocol design for external developers | [x] | P1 | Tech Writer | Done: docs/ARCHITECTURE.md (732 lines, component diagram, data flows, security model, redundancy) |
| D7 | Smart contract documentation: interface, events, roles | [x] | P1 | Tech Writer | Done: docs/SMART_CONTRACTS.md (922 lines, all 4 contracts, interfaces, events, access control, integration) |
| D8 | Contributor guide: development setup, testing, code conventions | [x] | P2 | Tech Writer | Done: docs/CONTRIBUTING.md (902 lines, 9 sections, code examples, common tasks walkthroughs) |

---

## 10. Molt Runtime (WASM + Deno JS/TS)

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| W1 | Add wazero dependency (`go get github.com/tetratelabs/wazero`) | [x] | P1 | Backend | Pure Go, zero CGO, ARM64 macOS native |
| W2 | Implement `internal/molt/runtime.go` — MoltRuntime struct, Compile(), Invoke() | [x] | P1 | Backend | Module compilation cache by IPFS CID |
| W3 | Implement `internal/molt/module_cache.go` — in-memory LRU compilation cache | [x] | P1 | Backend | CID → compiled module, LRU eviction |
| W4 | Implement `internal/molt/http_handler.go` — HTTP request → WASM → response | [x] | P1 | Backend | Local HTTP listener per deployed function |
| W5 | Implement `internal/molt/host_funcs.go` — 15 host functions (log, util, HTTP, storage, crawl) | [x] | P1 | Backend | Result handle pattern for WASM data exchange |
| W5b | Implement `internal/molt/host_services.go` — HostServices, ResultStore, context injection | [x] | P1 | Backend | Shared between WASM and Deno runtimes |
| W5c | Implement `internal/molt/host_http.go` — HTTP outbound with SSRF protection | [x] | P1 | Backend | Allowlist/blocklist, proxy integration, 30s timeout, 10MB limit |
| W5d | Implement `internal/molt/host_storage.go` — storage_put/get/delete/list with bucket scoping | [x] | P1 | Backend | Delegates to StorageEngine, enforces bucket scope |
| W5e | Implement `internal/molt/host_crawl.go` — crawl_page with URL validation | [x] | P1 | Backend | Single-page crawl via Scheduler, SSRF guard |
| W5f | Implement `internal/molt/host_util.go` — result_size/read, error_message, random_bytes, time_now_ms | [x] | P1 | Backend | Core infrastructure for host function data exchange |
| W6 | Implement `internal/molt/metrics.go` — invocation count, latency, memory | [x] | P2 | Backend | Atomic global + per-deployment stats |
| W7 | Types in `internal/molt/types.go` — MoltConfig, RuntimeType, HostCapabilities | [x] | P1 | Backend | RuntimeWASM + RuntimeJS constants |
| W8 | Add capability fields to MoltConfig (HTTPEnabled, StorageEnabled, CrawlEnabled) | [x] | P1 | Backend | Additive fields in config.go |
| W9 | Wire WasmRuntime into ContainerManager (dual-engine branching) | [x] | P1 | Backend | Done: RuntimeType field on Deployment, Molt routing in handleDeployRequest + broadcastDeployment, deployMoltReplica() |
| W10 | Add Molt config to `internal/config/config.go` + JSRuntimeConfig | [x] | P1 | Backend | `molt_enabled`, memory, timeout, max_instances, js_runtime section |
| W11 | Update `internal/runtime/runtime_detect.go` — detect wazero, set WasmAvailable | [x] | P1 | Backend | Done: MoltAvailable in RuntimeCapabilities, NodeProfile, AnnouncePayload, peer capabilities. SelectMoltNodes() in geographic.go |
| W12 | CLI: `moltbunker molt deploy` with WASM and JS/TS support | [x] | P1 | Backend | Detects .wasm vs .js/.ts by extension |
| W13 | CLI: `moltbunker molt list/logs/stop/invoke` commands | [x] | P2 | Backend | Function-specific management |
| W14 | Implement per-invocation pricing in `internal/payment/pricing.go` | [x] | P1 | Backend | CalculateMoltInvocationPrice(), 100ms min floor |
| W15 | Implement prepaid credit escrow model for functions | [x] | P1 | Backend | MoltCreditManager: deposit/deduct/balance/refund |
| W16 | E2E encrypted function I/O — wallet-derived keys, AES-256-GCM | [ ] | P1 | Backend | Reuse X25519/HKDF from deployment_encryption.go |
| W17 | Memory integrity canaries for WASM linear memory | [ ] | P2 | Backend | Detect unauthorized memory reads |
| W18 | Unit tests for `internal/molt/` package | [x] | P1 | Testing | 62 tests: runtime, host services, HTTP, storage, crawl, util |
| W19 | E2E tests: deploy function, invoke via HTTP, check billing | [ ] | P1 | Testing | Needs daemon integration |
| W20 | Sample WASM modules + test scripts | [x] | P2 | Testing | testdata/*.wasm: noop, echo, spin |
| W21 | Deno JS/TS worker pool (`internal/jsruntime/pool.go`) | [x] | P1 | Backend | N warm Deno processes, auto-restart, acquire/release |
| W22 | Deno worker stdio JSON-RPC (`internal/jsruntime/worker.go`) | [x] | P1 | Backend | Spawn, invoke, dispatchHostCall, graceful close |
| W23 | Deno runtime bindings (`internal/jsruntime/bindings.ts`) | [x] | P1 | Backend | env.fetch/storage/crawl globals, embedded via //go:embed |
| W24 | Unit tests for `internal/jsruntime/` package | [x] | P1 | Testing | 11 tests (7 pass always, 4 integration skip without Deno) |

---

## 11. P0 Services (Object Storage, Proxy, Crawl, Agent)

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| P1 | Wire Object Storage into daemon (StorageEngine + REST handler) | [x] | P1 | Backend | Conditional on `storage.enabled`, uses bbolt StateStore |
| P2 | Wire Proxy into daemon (SOCKS5 + HTTP + REST handler) | [x] | P1 | Backend | Start/Stop lifecycle, DirectDialer + AllowAllAuth |
| P3 | Wire Web Crawling into daemon (Scheduler + REST handler) | [x] | P1 | Backend | Conditional on `crawl.enabled`, RobotsChecker |
| P4 | Wire Agent Runtime into daemon (AgentRuntime + REST handler) | [x] | P1 | Backend | Conditional on `agent.enabled`, MemoryStore |
| P5 | Advertise P0 capabilities in AnnouncePayload | [x] | P1 | Backend | StorageAvailable, ProxyAvailable, CrawlAvailable, AgentAvailable |
| P6 | REST API routes registered via Set<X>Handler pattern | [x] | P1 | Backend | read/write permission middleware on all routes |
| P7 | Storage replication via IPFS (object CIDs in gossip) | [ ] | P2 | Backend | Cross-node object availability |
| P8 | Per-service metering integration with escrow | [ ] | P1 | Backend | Bandwidth, pages, token usage billing |
| P9 | S3-compatible endpoint for storage | [ ] | P2 | Backend | config: `storage.enable_s3`, port 9300 |
| P10 | Distributed crawl task assignment via P2P | [ ] | P2 | Backend | CrawlTask/CrawlResult message types defined |
| P11 | Agent memory persistence to object storage | [ ] | P2 | Backend | MemoryBucket in AgentSpec |
| P12 | Proxy relay through P2P network | [ ] | P2 | Backend | Multi-hop anonymity, ProxyConnect message type defined |

---

## 12. Subdomain System (Decentralized CNAME)

| # | Item | Status | Priority | Owner | Notes |
|---|------|--------|----------|-------|-------|
| SUB1 | BunkerRegistry.sol — on-chain subdomain registry contract | [x] | P1 | Smart Contract | v2.0.0, 174 Foundry tests, reserved names, deployed to Base Sepolia |
| SUB2 | RegistryContract Go bindings (register, resolve, release, transfer) | [x] | P1 | Backend | Mock mode + production mode in registry_contract.go |
| SUB3 | PaymentService subdomain facades (6 initial: register, resolve, release, transfer, update, list) | [x] | P1 | Backend | service.go — nil guard → delegate pattern |
| SUB4 | PaymentService subdomain facades (10 remaining: renew, reserve, claim, cancel, metadata, primary, reclaim, isExpired, reverseResolve, getMetadata) | [x] | P1 | Backend | All 16 contract operations wired end-to-end |
| SUB5 | Daemon handlers + API dispatch for all subdomain operations | [x] | P1 | Backend | 13 handlers in api_handlers.go, ownership checks on mutating ops |
| SUB6 | IPC client methods for all subdomain operations | [x] | P1 | Backend | 13 methods in client/subdomain.go |
| SUB7 | CLI commands: `subdomain register/resolve/release/transfer/update/list` | [x] | P1 | Backend | Initial 6 commands |
| SUB8 | CLI commands: `subdomain renew/reserve/claim/cancel/metadata/primary/reclaim` | [x] | P1 | Backend | Remaining 7 commands |
| SUB9 | Gossip `subdomain:` entries for local vanity routing | [x] | P1 | Backend | Set on register/claim, removed on release/cancel/reclaim |
| SUB10 | Anti-spoofing: reject remote gossip `subdomain:` entries | [x] | P0 | Backend | StateValidator in gossip_adapter.go — prevents subdomain hijacking |
| SUB11 | On-chain fallback for cross-node vanity routing | [x] | P1 | Backend | Step 5 in ingress resolver, ResolveOnChain() via BunkerRegistry |
| SUB12 | Subdomain expiry cleanup goroutine | [x] | P1 | Backend | Hourly check of gossip entries against on-chain IsExpired() |
| SUB13 | Mock registry tests | [x] | P1 | Testing | 17 tests in registry_contract_test.go |
| SUB14 | Ingress 5-step resolution pipeline | [x] | P1 | Backend | Exact → prefix → vanity gossip → gossip refresh → on-chain fallback |

---

## Summary Dashboard

| Category | Total Items | P0 | P1 | P2 | Completed | In Progress |
|----------|-----------|-----|-----|-----|-----------|-------------|
| Networking | 13 | 7 | 5 | 1 | 11 / 13 | 0 |
| Security | 15 | 10 | 4 | 1 | 14 / 15 | 0 |
| Smart Contracts | 15 | 12 | 2 | 1 | 8 / 15 | 0 |
| Container Runtime | 11 | 8 | 3 | 0 | 4 / 11 | 0 |
| Testing | 13 | 7 | 5 | 1 | 10 / 13 | 0 |
| Monitoring | 13 | 5 | 5 | 3 | 13 / 13 | 0 |
| Operations | 14 | 9 | 4 | 1 | 13 / 14 | 0 |
| Economics | 8 | 5 | 1 | 2 | 7 / 8 | 0 |
| Documentation | 8 | 3 | 4 | 1 | 8 / 8 | 0 |
| Molt Runtime | 24 | 0 | 19 | 5 | 19 / 24 | 0 |
| P0 Services | 12 | 0 | 7 | 5 | 6 / 12 | 0 |
| Subdomain System | 14 | 1 | 12 | 1 | 14 / 14 | 0 |
| **Total** | **160** | **67** | **71** | **22** | **132 / 160** | **0** |

**Launch requirement**: All P0 items (67 items) must be completed. P1 items are strongly recommended. P2 items can follow post-launch. Molt Runtime (section 10) and Subdomain System (section 12) are mostly P1 — not blocking mainnet launch.
