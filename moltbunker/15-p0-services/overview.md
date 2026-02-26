# P0 Services Architecture

*Object Storage, Proxy, Web Crawling, AI Agent Runtime*

*Created: February 26, 2026*

---

## Overview

Moltbunker ships 4 P0 services that extend the base container + Molt compute with platform capabilities. Each service runs as an optional module inside the daemon, enabled via config, advertised via the P2P announce protocol, and exposed through the REST API.

```
moltbunkerd
  ├── Container Engine (containerd)     ← core compute
  ├── Molt Runtime (wazero + Deno)      ← serverless compute
  ├── Object Storage                    ← encrypted S3-compatible store
  ├── Proxy                             ← SOCKS5/HTTP proxy (optional Tor)
  ├── Web Crawling                      ← distributed page scraping
  └── Agent Runtime                     ← AI agent orchestration
```

---

## Service Summary

| Service | Package | REST Routes | Config Key | Tests |
|---------|---------|-------------|------------|-------|
| Object Storage | `internal/storage/` | `/v1/storage/buckets`, `/v1/storage/objects/`, `/v1/storage/usage` | `storage.enabled` | 50+ |
| Proxy | `internal/proxy/` | `/v1/proxy/sessions`, `/v1/proxy/usage`, `/v1/proxy/status` | `proxy.enabled` | 40+ |
| Web Crawling | `internal/crawl/` | `/v1/crawl/jobs`, `/v1/crawl/pages`, `/v1/crawl/stats` | `crawl.enabled` | 45+ |
| Agent Runtime | `internal/agent/` | `/v1/agents`, `/v1/agents/{id}/invoke`, `/v1/agents/{id}/memory` | `agent.enabled` | 42+ |

---

## Daemon Wiring

All 4 services follow the same integration pattern:

1. **Config gate** — each service checks `cfg.<Service>.Enabled` before initialization
2. **Constructor** — creates the service core (engine, server, scheduler, runtime)
3. **REST handler** — `NewRESTHandler(core)` wraps the service in HTTP endpoints
4. **Server registration** — `httpAPIServer.Set<X>Handler(handler)` wires routes
5. **Announce** — capability flags are advertised to peers via `AnnouncePayload`

```
cmd/daemon/main.go
  │
  ├─ if cfg.Storage.Enabled:
  │    StorageEngine(dataDir, stateStore, EngineConfig)
  │    → httpAPIServer.SetStorageHandler(RESTHandler)
  │
  ├─ if cfg.Proxy.Enabled:
  │    proxy.NewServer(Config, DirectDialer, AllowAllAuth)
  │    → proxyServer.Start(ctx)
  │    → httpAPIServer.SetProxyHandler(RESTHandler)
  │    → defer proxyServer.Stop()
  │
  ├─ if cfg.Crawl.Enabled:
  │    crawl.NewScheduler(SchedulerConfig)
  │    → httpAPIServer.SetCrawlHandler(RESTHandler, RobotsChecker)
  │
  └─ if cfg.Agent.Enabled:
       agent.NewAgentRuntime(RuntimeConfig)
       → httpAPIServer.SetAgentHandler(RESTHandler, MemoryStore)
```

### Lifecycle

- **Storage, Crawl, Agent**: Stateless REST handlers — no background goroutines, no shutdown hook needed.
- **Proxy**: Runs SOCKS5 + HTTP listeners — requires `Start(ctx)` and `Stop()` on shutdown.

---

## P2P Capability Advertisement

Each P0 service's availability is advertised in the `AnnouncePayload` during the P2P handshake:

```go
type AnnouncePayload struct {
    // ... existing fields ...
    StorageAvailable bool `json:"storage_available,omitempty"`
    ProxyAvailable   bool `json:"proxy_available,omitempty"`
    CrawlAvailable   bool `json:"crawl_available,omitempty"`
    AgentAvailable   bool `json:"agent_available,omitempty"`
}
```

These fields are also stored in `NodeCapabilities` on the peer record, enabling service-aware routing (e.g., route crawl tasks only to peers with `CrawlAvailable=true`).

The full service capability struct is defined in `pkg/types/service_types.go` as `ServiceCapabilities`, which extends the announce fields with additional metadata (storage capacity, proxy Tor support, agent frameworks).

---

## 1. Object Storage

**Package**: `internal/storage/`

Encrypted object storage with per-wallet bucket isolation. Objects are encrypted client-side (DEK wrapped with X25519), stored locally, and replicated via IPFS CID.

### Architecture

```
Client → REST API → StorageEngine → local disk (encrypted blobs)
                                   → bbolt metadata (via StateStore)
                                   → IPFS for CID-based replication
```

### Key Components

| Component | File | Purpose |
|-----------|------|---------|
| `StorageEngine` | `engine.go` | Core storage operations, bucket management |
| `RESTHandler` | `rest.go` | HTTP endpoints for buckets, objects, usage |
| `EngineConfig` | `types.go` | MaxBuckets, MaxObjectSize |

### REST API

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | `/v1/storage/buckets` | write | Create bucket |
| GET | `/v1/storage/buckets` | read | List buckets |
| GET/HEAD/DELETE | `/v1/storage/buckets/{name}` | read/write | Bucket operations |
| PUT/GET/HEAD/DELETE | `/v1/storage/objects/{bucket}/{key}` | read/write | Object CRUD |
| GET | `/v1/storage/usage` | read | Quota usage |

### Config

```yaml
storage:
  enabled: false          # Enable object storage
  data_dir: ""            # Blob directory (default: <data_dir>/storage)
  max_buckets: 100        # Per-wallet limit
  max_object_size: 5368709120  # 5GB
```

---

## 2. Decentralized Proxy

**Package**: `internal/proxy/`

SOCKS5 and HTTP proxy with optional Tor routing. Per-session metering for bandwidth billing.

### Architecture

```
Client → SOCKS5/HTTP → proxy.Server → DirectDialer (clearnet)
                                     → TorDialer (optional)
       REST API → RESTHandler → session tracking + bandwidth stats
```

### Key Components

| Component | File | Purpose |
|-----------|------|---------|
| `Server` | `server.go` | SOCKS5 + HTTP proxy listeners |
| `RESTHandler` | `rest.go` | Session management, usage stats |
| `DirectDialer` | `dialer.go` | Clearnet TCP dialer |
| `AllowAllAuth` | `auth.go` | Default authenticator |

### REST API

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/v1/proxy/sessions` | read | List active sessions |
| GET/DELETE | `/v1/proxy/sessions/{id}` | read/write | Session details, terminate |
| GET | `/v1/proxy/usage` | read | Bandwidth report |
| GET | `/v1/proxy/status` | read | Proxy server status |

### Config

```yaml
proxy:
  enabled: false
  socks5_addr: ":1080"
  http_addr: ":8118"
  use_tor: false
  max_sessions: 1000
```

---

## 3. Web Crawling

**Package**: `internal/crawl/`

Multi-page web crawling with robots.txt compliance, job scheduling, and distributed task assignment.

### Architecture

```
Client → REST API → Scheduler → crawler workers (concurrent)
                   → RobotsChecker → per-domain compliance
       P2P → CrawlTask/CrawlResult messages → distributed crawling
```

### Key Components

| Component | File | Purpose |
|-----------|------|---------|
| `Scheduler` | `scheduler.go` | Job lifecycle, worker pool |
| `RESTHandler` | `rest.go` | Job CRUD, single-page crawl, stats |
| `RobotsChecker` | `robots.go` | robots.txt parsing and caching |

### REST API

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | `/v1/crawl/jobs` | write | Submit crawl job |
| GET | `/v1/crawl/jobs` | read | List jobs |
| GET | `/v1/crawl/jobs/{id}` | read | Job status |
| GET | `/v1/crawl/jobs/{id}/results` | read | Job results |
| POST | `/v1/crawl/jobs/{id}/cancel` | write | Cancel job |
| POST | `/v1/crawl/pages` | write | Single-page crawl |
| GET | `/v1/crawl/stats` | read | Crawl statistics |

### Config

```yaml
crawl:
  enabled: false
  max_depth: 3
  max_pages: 1000
  max_concurrent: 10
  respect_robots: true
  default_delay_ms: 1000
```

---

## 4. AI Agent Runtime

**Package**: `internal/agent/`

Orchestration for AI agents (LangGraph, CrewAI, AutoGen, custom). Agents run in containers with MCP tool access and persistent memory via object storage.

### Architecture

```
Client → REST API → AgentRuntime → container deployment
                   → MemoryStore → in-memory agent state
                   → Object Storage → persistent memory checkpoints
```

### Key Components

| Component | File | Purpose |
|-----------|------|---------|
| `AgentRuntime` | `runtime.go` | Agent lifecycle, framework dispatch |
| `RESTHandler` | `rest.go` | Agent CRUD, invocation, memory |
| `MemoryStore` | `memory.go` | In-memory agent state |

### REST API

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| POST | `/v1/agents` | write | Deploy agent |
| GET | `/v1/agents` | read | List agents |
| GET/DELETE | `/v1/agents/{id}` | read/write | Agent details, remove |
| POST | `/v1/agents/{id}/invoke` | write | Invoke agent |
| GET | `/v1/agents/{id}/memory` | read | Agent memory state |
| POST | `/v1/agents/{id}/stop` | write | Stop agent |

### Config

```yaml
agent:
  enabled: false
  frameworks: [langgraph, crewai, autogen, custom]
  default_memory_mb: 2048
  max_agents_per_wallet: 10
  sync_interval_secs: 60
```

---

## Security

All P0 services inherit the API server's authentication stack:

- **API keys** (`mb_live_*`): Scoped read/write permissions
- **Wallet auth** (SIWE): Full access, per-wallet resource isolation
- **Rate limiting**: Per-IP token bucket

### Per-Service Security

| Service | Isolation | Encryption |
|---------|-----------|------------|
| Storage | Per-wallet bucket scoping | Client-side DEK + X25519 |
| Proxy | Per-session wallet binding | TLS to destination (HTTPS) |
| Crawl | Per-wallet job limits | Results encrypted at rest |
| Agent | Per-wallet agent limits | Container-level encryption |

---

## Pricing

P0 services are metered and billed in BUNKER tokens:

| Service | Billing Unit | Rate |
|---------|-------------|------|
| Storage | GB-month | 2,000 BUNKER/GB-mo |
| Proxy | GB transferred | 1,000 BUNKER/GB |
| Crawl | Per page | 50 BUNKER/page |
| Agent | Token budget | Pass-through + 5% fee |

Pricing integrates with the existing escrow system (`BunkerEscrow.sol`). Mol credit-style prepaid deposits are planned for high-frequency crawl/agent usage.

---

## Types

Shared types for all P0 services are defined in `pkg/types/service_types.go`:

- `ServiceCapabilities` — P2P capability advertisement
- `StorageBucket`, `StorageObject`, `StorageQuota`, `MultipartUpload` — storage
- `ProxySession`, `BandwidthReport` — proxy
- `AgentSpec`, `AgentDeployment`, `MCPToolDef` — agents
- `CrawlJob`, `CrawlTarget`, `CrawlResult` — crawling

---

## Future Work

- [ ] Storage replication via IPFS (object CIDs in gossip state)
- [ ] Proxy relay through P2P network (multi-hop anonymity)
- [ ] Distributed crawl task assignment via `CrawlTask` P2P messages
- [ ] Agent memory persistence to object storage buckets
- [ ] Per-service metering integration with escrow
- [ ] S3-compatible endpoint for storage (config: `storage.enable_s3`)
