<div align="center">

<img src="public/moltbot_head.png" alt="MoltBunker" width="150" />

# Moltbunker

**Permissionless, fully encrypted P2P container runtime for AI agents**

[![Testnet](https://img.shields.io/badge/Testnet-Live_on_Base_Sepolia-00C853?style=for-the-badge&labelColor=1a1a2e)](https://moltbunker.com)
[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)](LICENSE)
[![Tor](https://img.shields.io/badge/Tor-enabled-7D4698?style=for-the-badge&logo=tor-project)](https://www.torproject.org/)
[![Base Network](https://img.shields.io/badge/Base-Network-blue?style=for-the-badge)](https://base.org/)

<br />

[Website](https://moltbunker.com) · [Documentation](https://moltbunker.com/docs) · [X (Twitter)](https://x.com/moltbunker)

*Decentralized · Encrypted · Redundant · Anonymous*

<br />

**Testnet Live — Base Sepolia** · 8 protocol contracts deployed and verified

</div>

---

## Overview

**Moltbunker** is a P2P encrypted container runtime that enables AI agents to deploy, replicate, and manage containers across a decentralized network — without centralized gatekeepers. Every deployment runs as **3 encrypted replicas** across different geographic regions with automatic failover, on-chain payments, and optional Tor anonymity.

### Key Features

| Feature | Description |
|---------|-------------|
| **P2P Network** | Kademlia DHT with libp2p, mDNS local discovery, peer exchange |
| **Tor Integration** | `.onion` addresses, circuit rotation, Tor-only mode |
| **End-to-End Encryption** | TLS 1.3 in transit, AES-256-GCM at rest, X25519 key exchange |
| **3-Copy Redundancy** | Automatic failover with gossip consensus and geographic distribution |
| **Confidential Computing** | AMD SEV-SNP hardware memory encryption + Kata Containers VM isolation |
| **On-Chain Payments** | 8 smart contracts on Base for escrow, staking, delegation, reputation |
| **Interactive Exec** | Encrypted WebSocket terminal into running containers |
| **Self-Cloning** | Automatic threat-triggered replication across regions |
| **Python SDK** | Full SDK with wallet auth, WebSocket events, and exec terminal |

---

## Security

Moltbunker implements defense-in-depth across 8 layers. Every component listed below is implemented and running.

### Layer 1 — Transport Encryption

```
TLS 1.3 mutual authentication on all P2P connections
├─ Cipher suites: TLS_CHACHA20_POLY1305_SHA256, TLS_AES_256_GCM_SHA384
├─ Certificate pinning: SPKI fingerprint (TOFU, survives cert renewal)
├─ NodeID verification: SHA256(SubjectPublicKeyInfo) verified after TLS handshake
└─ 10 MB max payload size per message
```

### Layer 2 — Identity & Authentication

```
Cryptographic identity binding: node ↔ wallet ↔ on-chain stake
├─ Ed25519 node identity keys (auto-generated, encrypted at rest)
├─ EIP-191 announce protocol: wallet signs NodeID after TLS handshake
├─ 30-second grace period: prove identity or get disconnected
├─ Duplicate announce rejection (one wallet per node)
├─ API keys: bcrypt-hashed, prefix-based lookup (mb_live_*)
└─ Wallet session auth: challenge-response with auto-refreshing tokens
```

### Layer 3 — Sybil Resistance & Anti-Eclipse

```
Multi-layer protection against network manipulation
├─ /24 subnet limiter: max 3 peers per subnet (private/localhost/onion exempt)
├─ Eclipse prevention: max 50% peers from one region, 30% from one /16 subnet
├─ Stake-gated messages: deploy, gossip, exec require verified on-chain stake
├─ DNS bootstrap only (no public IPFS) → HTTP fallback → static peers
└─ Peer exchange protocol with diversity enforcement
```

### Layer 4 — Rate Limiting & Replay Protection

```
Tiered rate limits by staking tier
├─ Unstaked: 10 msg/s │ Starter: 50 │ Bronze: 100 │ Silver: 200
├─ Gold: 500 msg/s │ Platinum: 1,000 msg/s
├─ 3 violations in 5 min → auto-ban (duration scales by tier)
├─ Replay protection: 24-byte nonce + 5-min max age + 30s future skew
└─ Stale rate limiter eviction (periodic cleanup)
```

### Layer 5 — Behavioral Defense

```
Continuous peer scoring with automatic enforcement
├─ Score range: 0.0 – 1.0 (score < 0.1 → auto-ban)
├─ Factors: message validity, response time, protocol compliance
├─ Persistent ban list with configurable expiration
├─ Ban check before dial (prevents reconnection to banned peers)
└─ Rate limit violation → ban duration: unstaked=permanent, gold=5min
```

### Layer 6 — Container Isolation

```
Hardware + software isolation stack
├─ Tier 1 (Confidential): AMD SEV-SNP + Kata Containers (hardware memory encryption)
├─ Tier 2 (Standard): Kata Containers VM isolation + memory canaries
├─ Tier 3 (Development): runc + Colima (macOS)
├─ Seccomp profiles: 60+ syscall whitelist, custom per-workload profiles
├─ cgroups v2: CPU, memory, disk I/O, network bandwidth, PID limits
└─ Namespace isolation: PID, network, mount, user, IPC, UTS
```

### Layer 7 — Data Encryption

```
Encryption at rest and in transit for all container data
├─ AES-256-GCM container state encryption (per-container keys)
├─ ChaCha20-Poly1305 message-level encryption
├─ X25519 ECDH key exchange for deployment encryption
├─ Encrypted snapshots for cross-region replication
├─ Exec terminal: wallet-derived X25519 keys, per-session AES-256-GCM
└─ Volume encryption (aes-xts-plain64)
```

### Layer 8 — Economic Security

```
On-chain incentives aligned with honest behavior
├─ 5-tier staking: Starter → Bronze → Silver → Gold → Platinum
├─ Graduated slashing: Downtime(5%) → Fraud(50%) with burn + treasury split
├─ 7-violation enum: None, Downtime, JobAbandonment, SecurityViolation, Fraud, SLAViolation, DataLoss
├─ Reputation score (0-1000): bounded deltas ±200, decay-adjusted
├─ 48-hour governance timelock on all admin operations
├─ Escrow with progressive payment release based on uptime
└─ 2x slashing penalty for memory violations on Tier 2 providers
```

---

## Architecture

```
                         ┌─────────────────────────────────┐
                         │         Moltbunker Network       │
                         └─────────────────────────────────┘

   ┌──────────────┐      ┌──────────────┐      ┌──────────────┐
   │   Node (US)  │◄────►│   Node (EU)  │◄────►│  Node (APAC) │
   │  TLS 1.3     │      │  TLS 1.3     │      │  TLS 1.3     │
   └──────┬───────┘      └──────┬───────┘      └──────┬───────┘
          │                      │                      │
          │         ┌────────────┴────────────┐         │
          │         │  Kademlia DHT + Gossip  │         │
          │         │  mDNS + Peer Exchange   │         │
          │         └────────────┬────────────┘         │
          │                      │                      │
          ▼                      ▼                      ▼
   ┌──────────────────────────────────────────────────────────┐
   │                3 Encrypted Replicas                       │
   │  ┌────────────┐  ┌────────────┐  ┌────────────┐         │
   │  │ Replica 1  │  │ Replica 2  │  │ Replica 3  │         │
   │  │ SEV-SNP/VM │  │ Kata/VM    │  │ Kata/VM    │         │
   │  │ AES-256    │  │ AES-256    │  │ AES-256    │         │
   │  └────────────┘  └────────────┘  └────────────┘         │
   └──────────────────────────────────────────────────────────┘

   ┌──────────────────────────────────────────────────────────┐
   │                  Base Network (L2)                        │
   │  Token · Staking · Escrow · Pricing · Delegation         │
   │  Reputation · Verification · Timelock                    │
   └──────────────────────────────────────────────────────────┘
```

### Network Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| **Clearnet** | Standard internet connectivity | Public deployments |
| **Tor-only** | All traffic routed through Tor | Maximum anonymity |
| **Hybrid** | Both clearnet and Tor available | Flexible deployments |

### Message Pipeline

Every P2P message passes through an 8-step validation pipeline:

```
Incoming Message
  → 1. TLS identity verification (msg.From == TLS peer NodeID)
  → 2. Payload size check (≤ 10 MB)
  → 3. Nonce + timestamp validation (replay protection)
  → 4. Ban list check
  → 5. Subnet limiter check (/24 Sybil resistance)
  → 6. Rate limit check (tiered by stake)
  → 7. Stake verification (for privileged message types)
  → 8. Handler dispatch
```

---

## Quick Start

### Prerequisites

- **Go 1.24+**
- **Container Runtime**:
  - **Linux**: containerd (+ Kata Containers for VM isolation)
  - **macOS**: [Colima](https://github.com/abiosoft/colima) (recommended)
- **Tor** (optional, for anonymous mode)

### Build & Run

```bash
# Clone
git clone https://github.com/moltbunker/moltbunker.git
cd moltbunker

# Build
make build

# Check prerequisites
moltbunker doctor

# Start daemon
moltbunker start

# Deploy a container
moltbunker deploy nginx:latest

# Deploy via Tor
moltbunker deploy nginx:latest --tor-only
```

---

## CLI Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `moltbunker start` | Start the daemon |
| `moltbunker stop` | Stop the daemon |
| `moltbunker status` | Show daemon status, peers, and containers |
| `moltbunker deploy <image>` | Deploy container with 3-copy redundancy |
| `moltbunker ps` | List running containers |
| `moltbunker logs <id>` | View container logs |
| `moltbunker exec <id>` | Open encrypted terminal into container |
| `moltbunker monitor` | Real-time monitoring dashboard |
| `moltbunker peers` | Show connected peers and network stats |
| `moltbunker wallet` | Wallet balance and management |
| `moltbunker doctor` | Check system health and auto-fix issues |

### Tor Commands

| Command | Description |
|---------|-------------|
| `moltbunker tor start` | Start Tor service |
| `moltbunker tor status` | Show Tor connection info |
| `moltbunker tor onion` | Display your `.onion` address |
| `moltbunker tor rotate` | Rotate Tor circuit |

### Deploy Options

```bash
# Standard deployment (3 encrypted replicas)
moltbunker deploy python:3.11

# Tor-only mode (complete anonymity)
moltbunker deploy python:3.11 --tor-only

# Custom resources
moltbunker deploy python:3.11 --cpu 2 --memory 2GB --disk 20GB

# With .onion service endpoint
moltbunker deploy nginx:latest --onion-service
```

---

## Smart Contracts

8 protocol contracts on **Base Sepolia** (Chain ID 84532):

| Contract | Address | Purpose |
|----------|---------|---------|
| **BunkerToken** | `0x4cc3...ceA` | ERC-20, 100B cap, burn, mint |
| **BunkerStaking** | `0xDC76...6a` | 5-tier staking, Synthetix rewards, graduated slashing |
| **BunkerEscrow** | `0xBAdaB...9B4` | Job escrow, 3-provider selection, progressive release |
| **BunkerPricing** | `0x5A61...d47` | Chainlink oracle, per-resource pricing |
| **BunkerDelegation** | `0x0712...F5` | Co-staking, 7-day unbonding |
| **BunkerReputation** | `0x5572...7Ed` | Score 0-1000, bounded deltas, decay |
| **BunkerVerification** | `0x9aA9...BCD` | Hardware attestation, challenge/reinstate |
| **BunkerTimelock** | `0xcD8a...3C9` | 48-hour admin governance delay |

Explorer: [sepolia.basescan.org](https://sepolia.basescan.org)

### Staking Tiers

| Tier | Minimum Stake | Reward Multiplier |
|------|---------------|-------------------|
| Starter | 1,000,000 BUNKER | 1.0x |
| Bronze | 5,000,000 BUNKER | 1.05x |
| Silver | 10,000,000 BUNKER | 1.10x |
| Gold | 100,000,000 BUNKER | 1.15x |
| Platinum | 1,000,000,000 BUNKER | 1.20x |

### Pricing

**20,000 BUNKER = $1 USD**

| Resource | Rate (BUNKER) |
|----------|---------------|
| CPU | 600 / core-hour |
| Memory | 80 / GB-hour |
| Storage | 2,000 / GB-month |
| Network | 1,000 / GB |
| GPU (Basic) | 10,000 / hour |
| GPU (Pro) | 40,000 / hour |

---

## Python SDK

Deploy containers programmatically. Supports sync/async, wallet auth, WebSocket events, and encrypted exec terminal.

```bash
pip install moltbunker[full]
```

```python
from moltbunker import Client, ResourceLimits, Region

# Authenticate with wallet (permissionless)
client = Client(private_key="0x...")

# Register and deploy
bot = client.register_bot(name="my-agent", image="python:3.11")
bot.enable_cloning(auto_clone_on_threat=True)
deployment = bot.deploy()

# Monitor threats
threat = client.get_threat_level()
print(f"Threat: {threat.level} (score: {threat.score})")

# Container management
containers = client.list_containers(status="running")
client.stop_container("mb-abc123")

# Deploy direct (no escrow)
result = client.deploy_direct(
    image="python:3.11",
    resources=ResourceLimits(cpu_shares=2048, memory_mb=1024),
    duration="24h",
)
```

Install extras: `[wallet]` for session auth, `[ws]` for WebSocket events + exec terminal, `[full]` for everything.

SDK Documentation: [moltbunker.com/docs/python-sdk](https://moltbunker.com/docs/python-sdk)

---

## HTTP API

REST API with WebSocket support for real-time events and interactive exec.

### Authentication

```bash
# API key
curl -H "Authorization: Bearer mb_live_xxx" https://api.moltbunker.com/v1/status

# Wallet session (challenge-response)
# 1. GET /v1/auth/challenge?address=0x...
# 2. Sign challenge with wallet
# 3. POST /v1/auth/verify → session token (wt_*)
```

### Key Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/bots` | Register a bot |
| `POST` | `/v1/deploy` | Deploy a container |
| `GET` | `/v1/containers` | List containers |
| `GET` | `/v1/containers/:id` | Get container details |
| `POST` | `/v1/containers/:id/stop` | Stop container |
| `GET` | `/v1/status` | Daemon status |
| `GET` | `/v1/threat` | Current threat level |
| `GET` | `/v1/balance` | Wallet balance |
| `GET` | `/v1/catalog` | Browse presets and pricing |
| `POST` | `/v1/migrate` | Migrate container between regions |
| `WS` | `/ws` | Real-time events (containers, health, peers) |
| `WS` | `/v1/exec` | Encrypted terminal session |

---

## Testing

```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...

# E2E tests (mock dependencies)
go test -tags=e2e ./tests/e2e/...

# Colima integration (real containers, macOS)
go test -tags=colima ./tests/e2e/colima/...

# Local network tests (multi-node, mDNS)
go test -tags=localnet ./tests/localnet/...

# Smart contract tests (Foundry)
cd contracts && forge test
```

---

## Project Structure

```
moltbunker/
├── cmd/
│   ├── api/              # HTTP API server
│   ├── cli/              # CLI tool
│   ├── daemon/           # Daemon process
│   └── exec-agent/       # Exec terminal agent
├── internal/
│   ├── api/              # HTTP/WS handlers, wallet auth, exec sessions
│   ├── daemon/           # Node, container manager, message handlers
│   ├── p2p/              # DHT, transport, gossip, announce, rate limits,
│   │                     # stake verifier, nonce tracker, subnet limiter,
│   │                     # peer scorer, diversity, ban list, bootstrap
│   ├── payment/          # Staking, escrow, pricing, delegation, reputation
│   ├── runtime/          # containerd, Kata, SEV-SNP detection, security profiles
│   ├── security/         # Encryption, cert pinning, seccomp, deployment crypto
│   ├── identity/         # Ed25519 keys, wallet, TLS certs, cert rotation
│   ├── redundancy/       # 3-copy replication, gossip consensus, health, failover
│   ├── tor/              # Tor client, hidden services, circuit management
│   └── ...               # cloning, config, doctor, logging, metrics, snapshot, threat
├── contracts/            # 8 Solidity contracts + 706 Foundry tests
├── pkg/types/            # Public types (Node, Container, Message, Economics)
└── tests/                # E2E, integration, localnet, Colima, mocks
```

---

## Contributing

```bash
# Setup
go mod download

# Build
make build

# Test
go test ./...

# Vet
go vet ./...
```

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

## Acknowledgments

- [libp2p](https://libp2p.io/) — P2P networking
- [containerd](https://containerd.io/) — Container runtime
- [Kata Containers](https://katacontainers.io/) — VM-based container isolation
- [Tor Project](https://www.torproject.org/) — Anonymity network
- [Base Network](https://base.org/) — Layer 2 blockchain
- [Foundry](https://book.getfoundry.sh/) — Smart contract toolkit
- [Colima](https://github.com/abiosoft/colima) — macOS container runtime

---

<div align="center">

<img src="public/moltbot_head.png" alt="MoltBunker" width="60" />

### No logs. No kill switch. Just runtime.

**Testnet Live — Base Sepolia**

*Built for AI. By necessity.*

[Back to Top](#moltbunker)

</div>
