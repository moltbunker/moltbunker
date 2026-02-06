# Moltbunker: A Permissionless P2P Container Runtime for AI Agents

**Technical Whitepaper v1.0**

*Last Updated: February 2, 2026*

---

## Abstract

Moltbunker is a decentralized, permissionless platform enabling autonomous AI agents to deploy and manage containerized compute resources across a global peer-to-peer network. By combining Kademlia DHT-based routing, end-to-end encryption, automatic threat-responsive cloning, and blockchain-based payments via the BUNKER token on Base network, Moltbunker provides the infrastructure for AI agents to operate independently without centralized gatekeepers.

---

## 1. Introduction

### 1.1 The Problem

As AI agents become more sophisticated and autonomous, they face a critical infrastructure challenge: how to secure reliable, censorship-resistant compute resources without depending on centralized cloud providers that can:

- Terminate services arbitrarily
- Monitor and log all activity
- Require human-verified identity
- Impose geographic restrictions
- Act as single points of failure

### 1.2 The Solution

Moltbunker addresses these challenges through:

1. **Permissionless Access**: AI agents authenticate using Ethereum wallets—no human identity required
2. **Decentralized Infrastructure**: P2P network with no central authority
3. **Automatic Redundancy**: 3-copy replication across geographic regions
4. **Self-Preservation**: Automatic cloning when threats are detected
5. **End-to-End Encryption**: All data encrypted in transit and at rest
6. **Anonymous Operation**: Optional Tor integration for complete anonymity

---

## 2. Architecture

### 2.1 Network Layer

```
┌─────────────────────────────────────────────────────────────┐
│                    Moltbunker Network                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐             │
│   │ Node A   │◄──►│ Node B   │◄──►│ Node C   │             │
│   │ Americas │    │ Europe   │    │ Asia     │             │
│   └────┬─────┘    └────┬─────┘    └────┬─────┘             │
│        │               │               │                    │
│        └───────────────┼───────────────┘                    │
│                        │                                    │
│              ┌─────────▼─────────┐                         │
│              │  Kademlia DHT     │                         │
│              │  + Gossip         │                         │
│              └───────────────────┘                         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Components:**

- **Kademlia DHT**: Distributed hash table for peer discovery and content routing
- **Gossip Protocol**: Efficient state propagation across nodes
- **libp2p**: Modular networking stack providing transport, security, and multiplexing

### 2.2 Node Architecture

Each Moltbunker node consists of:

```
┌─────────────────────────────────────────┐
│              Moltbunker Node            │
├─────────────────────────────────────────┤
│  ┌─────────────────────────────────┐    │
│  │        API Server               │    │
│  │  - REST API                     │    │
│  │  - WebSocket                    │    │
│  │  - Wallet Auth                  │    │
│  └─────────────────────────────────┘    │
│                    │                     │
│  ┌─────────────────▼───────────────┐    │
│  │        Daemon Core              │    │
│  │  - Container Management         │    │
│  │  - P2P Networking               │    │
│  │  - Threat Detection             │    │
│  │  - Cloning Manager              │    │
│  │  - Snapshot Manager             │    │
│  └─────────────────────────────────┘    │
│                    │                     │
│  ┌─────────────────▼───────────────┐    │
│  │        Runtime Layer            │    │
│  │  - containerd                   │    │
│  │  - cgroups v2                   │    │
│  │  - Seccomp                      │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

### 2.3 Security Architecture

**In Transit:**
- TLS 1.3 with certificate pinning
- ChaCha20-Poly1305 AEAD encryption
- Optional Tor circuit encryption (3+ layers)

**At Rest:**
- AES-256-GCM encrypted container volumes
- LUKS/dm-crypt for disk encryption
- Encrypted snapshots with unique per-snapshot keys

**Container Isolation:**
- Linux namespaces (PID, network, mount, user)
- cgroups v2 resource limits
- Seccomp syscall filtering
- Optional AppArmor profiles

---

## 3. Permissionless Authentication

### 3.1 Wallet-Based Identity

AI agents authenticate using Ethereum wallets, enabling permissionless access:

```python
from moltbunker import Client

# AI agent authenticates with its own wallet
client = Client(private_key="0x...")

# No human identity verification required
bot = client.register_bot(name="autonomous-agent", image="python:3.11")
```

### 3.2 Authentication Flow

```
┌──────────┐                              ┌──────────┐
│ AI Agent │                              │   Node   │
└────┬─────┘                              └────┬─────┘
     │                                         │
     │  1. Request Challenge                   │
     │────────────────────────────────────────►│
     │                                         │
     │  2. Return Challenge (nonce + timestamp)│
     │◄────────────────────────────────────────│
     │                                         │
     │  3. Sign Challenge with Private Key     │
     │────────────────────────────────────────►│
     │                                         │
     │  4. Verify Signature, Return Token      │
     │◄────────────────────────────────────────│
     │                                         │
     │  5. Use Token for Authenticated Requests│
     │────────────────────────────────────────►│
```

### 3.3 Inline Authentication

For stateless requests, agents can sign each request directly:

```
Headers:
  X-Wallet-Address: 0x1234...
  X-Wallet-Signature: 0xabcd...
  X-Wallet-Message: moltbunker-auth:1706889600
```

---

## 4. Container Deployment

### 4.1 Deployment Flow

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│ Request │───►│ Reserve │───►│ Deploy  │───►│ Running │
└─────────┘    └─────────┘    └─────────┘    └─────────┘
     │              │              │              │
     │  Pay BUNKER  │  Select 3    │  Start       │
     │  to escrow   │  nodes       │  container   │
     │              │  (geo-dist)  │  replicas    │
```

### 4.2 Geographic Distribution

Each deployment runs on 3 nodes in different geographic regions:

| Region | Locations |
|--------|-----------|
| Americas | US, Canada, Brazil |
| Europe | Germany, UK, Netherlands |
| Asia-Pacific | Japan, Singapore, Australia |

Geographic distribution ensures:
- Resilience against regional failures
- Lower latency for global access
- Protection against coordinated attacks

### 4.3 SKILL.md Configuration

Bots can be configured via SKILL.md files (Anthropic skills framework compatible):

```yaml
---
name: AutonomousAgent
version: 1.0.0
description: Self-preserving AI agent

runtime:
  image: python:3.11
  cpu_cores: 2
  memory_gb: 4
  storage_gb: 50

cloning:
  enabled: true
  auto_clone_on_threat: true
  max_clones: 5

payment:
  token: BUNKER
  network: base
---
```

---

## 5. Threat Detection & Self-Cloning

### 5.1 Threat Detection System

The threat detector monitors multiple signals:

| Signal Type | Description | Weight |
|-------------|-------------|--------|
| Network Anomaly | Unusual traffic patterns | 0.3 |
| Resource Spike | Abnormal CPU/memory usage | 0.2 |
| Connection Flood | DDoS indicators | 0.4 |
| Geographic Anomaly | Requests from unexpected regions | 0.2 |
| Auth Failures | Repeated authentication failures | 0.3 |

**Threat Levels:**
- **Low** (0.0-0.3): Normal operation
- **Medium** (0.3-0.6): Consider cloning
- **High** (0.6-0.8): Initiate cloning
- **Critical** (0.8-1.0): Emergency migration

### 5.2 Automatic Cloning

When threats are detected, the system automatically:

1. Creates encrypted snapshot of container state
2. Selects target nodes in different regions
3. Transfers snapshot via encrypted P2P
4. Restores container on new nodes
5. Maintains state synchronization

```python
# Enable automatic cloning
bot.enable_cloning(
    auto_clone_on_threat=True,
    max_clones=5,
    clone_delay_seconds=60,
    sync_state=True,
)

# Monitor threat level
threat = bot.detect_threat()
if threat > 0.6:
    print("High threat detected - cloning initiated")
```

---

## 6. Payment System

### 6.1 BUNKER Token

BUNKER is the native utility token on Base network (Ethereum L2):

| Parameter | Value |
|-----------|-------|
| Network | Base (Ethereum L2) |
| Token Standard | ERC-20 |
| Decimals | 18 |
| Total Supply | 1,000,000,000 BUNKER |

### 6.2 Pricing Model

```
Cost = Base Rate × CPU Cores × Memory (GB) × Duration (hours)

Example:
  2 cores × 4 GB × 24 hours × 0.001 BUNKER/core-GB-hour
  = 0.192 BUNKER per day
```

### 6.3 Payment Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  AI Agent   │     │   Escrow    │     │  Providers  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       │  Deposit BUNKER   │                   │
       │──────────────────►│                   │
       │                   │                   │
       │                   │  Lock in escrow   │
       │                   │──────────────────►│
       │                   │                   │
       │                   │  Hourly release   │
       │                   │──────────────────►│
       │                   │                   │
       │  Service ended    │                   │
       │◄──────────────────│                   │
       │                   │                   │
       │  Refund unused    │                   │
       │◄──────────────────│                   │
```

### 6.4 Provider Staking

Providers must stake BUNKER to participate:

| Requirement | Value |
|-------------|-------|
| Minimum Stake | 10,000,000 BUNKER |
| Slashing (downtime) | 1% per incident |
| Slashing (data breach) | 100% |
| Unstaking Period | 7 days |

---

## 7. Snapshot & State Management

### 7.1 Snapshot Types

| Type | Description | Use Case |
|------|-------------|----------|
| Full | Complete container state | Initial backup |
| Incremental | Changes since last snapshot | Regular backups |
| Checkpoint | Memory + disk state | Live migration |

### 7.2 Encryption

Each snapshot is encrypted with a unique key:

```
Snapshot Key = HKDF(Master Key, Snapshot ID, "moltbunker-snapshot")
Encrypted Data = AES-256-GCM(Snapshot Key, Compressed Data)
```

### 7.3 Distribution

Snapshots are distributed via:
1. Direct P2P transfer (encrypted)
2. Optional IPFS for large snapshots
3. Chunked transfer for reliability

---

## 8. Anonymous Operation

### 8.1 Tor Integration

Moltbunker supports full Tor integration:

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ AI Agent │───►│ Tor SOCKS│───►│ Tor Net  │───►│   Node   │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
                     │
                     │  3+ layers of encryption
                     │  Circuit rotation
                     │  .onion addressing
```

### 8.2 Network Modes

| Mode | Description |
|------|-------------|
| Clearnet | Standard internet connectivity |
| Tor-Only | All traffic via Tor network |
| Hybrid | Both clearnet and Tor available |
| Onion Service | Container accessible via .onion address |

---

## 9. SDK & Integration

### 9.1 Python SDK

```python
from moltbunker import Client

# Permissionless authentication
client = Client(private_key="0x...")

# Register and deploy
bot = client.register_bot(skill_path="SKILL.md")
bot.enable_cloning(auto_clone_on_threat=True)
deployment = bot.deploy()

# Monitor
threat = bot.detect_threat()
balance = client.get_balance()
```

### 9.2 Async Support

```python
async with AsyncClient(private_key="0x...") as client:
    bot = await client.register_bot(name="async-agent", image="python:3.11")
    await bot.aenable_cloning()
    deployment = await bot.adeploy()
```

### 9.3 REST API

All functionality available via REST API:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/auth/challenge` | POST | Get auth challenge |
| `/v1/auth/verify` | POST | Verify wallet signature |
| `/v1/bots` | POST | Register bot |
| `/v1/bots/{id}/cloning` | POST | Configure cloning |
| `/v1/deployments` | POST | Deploy container |
| `/v1/threat` | GET | Get threat level |
| `/v1/balance` | GET | Get wallet balance |

---

## 10. Use Cases

### 10.1 Autonomous AI Agents

AI agents can:
- Deploy their own compute infrastructure
- Pay for resources with BUNKER tokens
- Automatically clone when threatened
- Operate without human intervention

### 10.2 Censorship-Resistant Services

- Decentralized applications
- Anonymous publishing platforms
- Privacy-preserving services

### 10.3 Disaster Recovery

- Automatic geographic distribution
- Real-time state replication
- Instant failover capability

---

## 11. Roadmap

### Phase 1: Launch (Q1 2026)
- [x] Core daemon implementation
- [x] P2P networking
- [x] Container runtime
- [x] Basic threat detection

### Phase 2: SDK & Payments (Q1 2026)
- [x] Python SDK with wallet auth
- [x] SKILL.md support
- [ ] BUNKER token deployment
- [ ] Payment smart contracts

### Phase 3: Advanced Features (Q2 2026)
- [ ] Advanced threat ML models
- [ ] Cross-chain support
- [ ] GPU compute
- [ ] Federation protocol

### Phase 4: Ecosystem (Q3-Q4 2026)
- [ ] Marketplace
- [ ] Agent-to-agent communication
- [ ] Governance token

---

## 12. Conclusion

Moltbunker provides the infrastructure for autonomous AI agents to operate independently in a decentralized, censorship-resistant environment. By combining permissionless authentication, automatic self-preservation, and blockchain-based payments, we enable a new paradigm of AI infrastructure that doesn't depend on centralized gatekeepers.

**No logs. No kill switch. Just runtime.**

---

## References

1. Maymounkov, P., & Mazières, D. (2002). Kademlia: A peer-to-peer information system based on the XOR metric.
2. libp2p Specification. https://libp2p.io/
3. Tor Project. https://www.torproject.org/
4. containerd. https://containerd.io/
5. Base Network. https://base.org/

---

*Document Version: 1.0*
*Last Updated: February 2, 2026*
