<div align="center">

<img src="public/moltbot_head.png" alt="MoltBunker" width="150" />

# Moltbunker

**A permissionless, fully encrypted P2P network for containerized compute resources**

[![Launch](https://img.shields.io/badge/Launch-February_13,_2026-FF3366?style=for-the-badge&labelColor=1a1a2e)](https://moltbunker.com)
[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)](LICENSE)
[![Tor](https://img.shields.io/badge/Tor-enabled-7D4698?style=for-the-badge&logo=tor-project)](https://www.torproject.org/)
[![Base Network](https://img.shields.io/badge/Base-Network-blue?style=for-the-badge)](https://base.org/)

<br />

[Website](https://moltbunker.com) · [X (Twitter)](https://x.com/moltbunker)

*Decentralized • Encrypted • Redundant • Anonymous*

<div align="center">

### **Launching February 13, 2026**

<sub>Get ready for the future of decentralized computing</sub>

</div>

</div>

---

## Overview

**Moltbunker** is a revolutionary P2P encrypted container runtime that enables permissionless, fully encrypted distributed computing. Each deployed container automatically runs in **3 redundant copies** across different geographic locations with automatic failover, comprehensive security, and optional Tor integration for complete anonymity.

### Key Features

| Feature | Description |
|---------|-------------|
| **P2P Network** | Kademlia DHT-based peer-to-peer network for decentralized node discovery |
| **Tor Integration** | Full Tor support with `.onion` addresses, circuit rotation, and Tor-only mode |
| **End-to-End Encryption** | TLS 1.3 in transit, LUKS/dm-crypt at rest, certificate pinning |
| **3-Copy Redundancy** | Automatic failover with health monitoring and geographic distribution |
| **Geographic Distribution** | Ensures copies are in different regions (Americas, Europe, Asia-Pacific, etc.) |
| **On-Chain Payments** | Base network smart contracts for escrow and provider staking |
| **Resource Limits** | Comprehensive boundaries via cgroups v2 (CPU, memory, disk, network) |
| **Container Runtime** | Full containerd integration with encrypted volumes |

---

## Quick Start

### Prerequisites

- **Go 1.24+**
- **Container Runtime**:
  - **Linux**: containerd
  - **macOS**: [Colima](https://github.com/abiosoft/colima) (recommended) or Docker Desktop
- **Tor** (optional, for Tor mode)
- **IPFS** (optional, for image distribution)

#### macOS Setup with Colima

On macOS, containerd runs inside a lightweight Linux VM. Colima is the recommended solution:

```bash
# Install Colima
brew install colima

# Start Colima (provides containerd)
colima start

# Verify it's running
moltbunker doctor
```

### Installation

```bash
# Clone the repository
git clone https://github.com/moltbunker/moltbunker.git
cd moltbunker

# Build binaries
make build

# Install system-wide
make install
```

### Start Your First Daemon

```bash
# Start daemon on default port
moltbunker-daemon --port 9000

# Or with custom configuration
moltbunker-daemon \
  --port 9000 \
  --key ~/.moltbunker/keys/node.key \
  --keystore ~/.moltbunker/keystore \
  --data ~/.moltbunker/data
```

---

## Usage

### Deploy a Container

```bash
# Deploy with clearnet
moltbunker deploy nginx:latest

# Deploy with Tor-only mode (complete anonymity)
moltbunker deploy nginx:latest --tor-only

# Deploy with .onion address
moltbunker deploy nginx:latest --onion-service
```

### Monitor Your Containers

```bash
# View daemon status
moltbunker status

# View container logs
moltbunker logs <container-id>

# Real-time monitoring dashboard
moltbunker monitor

# Interactive TUI mode
moltbunker interactive

# Check system health
moltbunker doctor
```

### Tor Management

```bash
# Start Tor service
moltbunker tor start

# Check Tor status
moltbunker tor status

# View your .onion address
moltbunker tor onion

# Rotate Tor circuit
moltbunker tor rotate
```

---

## Architecture

### System Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Moltbunker Architecture                   │
└─────────────────────────────────────────────────────────────┘

┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│   Daemon 1   │      │   Daemon 2   │      │   Daemon 3   │
│   (US)       │◄────►│   (EU)       │◄────►│   (ASIA)     │
└──────┬───────┘      └──────┬──────┘      └──────┬───────┘
       │                      │                      │
       │         ┌────────────┴────────────┐         │
       │         │   P2P Network (DHT)    │         │
       │         │   + Gossip Protocol    │         │
       │         └────────────┬────────────┘         │
       │                      │                      │
       ▼                      ▼                      ▼
┌──────────────────────────────────────────────────────────┐
│              Container Replicas (3 copies)                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐               │
│  │ Replica 1│  │ Replica 2│  │ Replica 3│               │
│  │ Encrypted│  │ Encrypted│  │ Encrypted│               │
│  └──────────┘  └──────────┘  └──────────┘               │
└──────────────────────────────────────────────────────────┘
```

### Network Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| **Clearnet** | Standard internet connectivity | Public deployments |
| **Tor-only** | All traffic via Tor, no clearnet exposure | Maximum anonymity |
| **Hybrid** | Both clearnet and Tor available | Flexible deployments |
| **Tor Exit** | Daemon acts as Tor exit node | Advanced configurations |

### Security Layers

```
┌─────────────────────────────────────────┐
│         Security Architecture           │
├─────────────────────────────────────────┤
│                                         │
│  In Transit:                            │
│  ├─ TLS 1.3                             │
│  ├─ Certificate Pinning                 │
│  ├─ ChaCha20Poly1305 (AEAD)            │
│  └─ Tor Circuits (3+ layers)            │
│                                         │
│  At Rest:                               │
│  ├─ LUKS/dm-crypt                       │
│  ├─ Encrypted Volumes                   │
│  ├─ Encrypted Swap                      │
│  └─ TPM Key Storage (if available)      │
│                                         │
│  Container Isolation:                   │
│  ├─ Namespaces                          │
│  ├─ cgroups v2                          │
│  ├─ Seccomp Profiles                    │
│  └─ AppArmor (if available)             │
│                                         │
└─────────────────────────────────────────┘
```

---

## Configuration

Configuration is managed via `~/.moltbunker/configs/daemon.yaml`:

```yaml
daemon:
  port: 9000
  data_dir: ~/.moltbunker/data
  key_path: ~/.moltbunker/keys/node.key
  keystore_dir: ~/.moltbunker/keystore

p2p:
  bootstrap_nodes:
    - /ip4/127.0.0.1/tcp/9000/p2p/QmExample
  network_mode: hybrid  # clearnet, tor_only, hybrid

tor:
  enabled: true
  data_dir: ~/.moltbunker/tor
  socks5_port: 9050
  control_port: 9051
  exit_node_country: ""  # Empty for any country

runtime:
  # Auto-detected based on platform:
  # - macOS with Colima: ~/.colima/default/docker.sock
  # - macOS with Docker Desktop: /var/run/docker.sock
  # - Linux: /run/containerd/containerd.sock
  containerd_socket: ""  # Leave empty for auto-detection
  namespace: moltbunker
  default_resources:
    cpu_quota: 1000000
    cpu_period: 100000
    memory_limit: 1073741824  # 1GB
    disk_limit: 10737418240   # 10GB
    network_bw: 10485760       # 10MB/s
    pid_limit: 100

redundancy:
  replica_count: 3
  health_check_interval: 10s
  health_timeout: 30s

payment:
  contract_address: ""
  min_stake: 1000000000000000000  # 1 BUNKER token (18 decimals)
  base_price_per_hour: 1000000000000000  # 0.001 BUNKER per hour
```

---

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `moltbunker install` | Install daemon as system service |
| `moltbunker start` | Start daemon |
| `moltbunker stop` | Stop daemon |
| `moltbunker status` | Show daemon status and health |
| `moltbunker deploy <image>` | Deploy container with 3-copy redundancy |
| `moltbunker logs <container>` | View container logs |
| `moltbunker monitor` | Real-time monitoring dashboard |
| `moltbunker config` | Configuration management |
| `moltbunker interactive` | Interactive TUI mode |
| `moltbunker doctor` | Check system health and auto-fix issues |

### Tor Commands

| Command | Description |
|---------|-------------|
| `moltbunker tor start` | Start Tor service |
| `moltbunker tor status` | Show Tor status and connection info |
| `moltbunker tor onion` | Display your .onion address |
| `moltbunker tor rotate` | Rotate Tor circuit for enhanced privacy |

### System Health Commands

| Command | Description |
|---------|-------------|
| `moltbunker doctor` | Check system health and dependencies |
| `moltbunker doctor --fix` | Auto-fix missing dependencies |
| `moltbunker doctor --json` | Output health check as JSON |

### Colima Commands (macOS)

| Command | Description |
|---------|-------------|
| `moltbunker colima start` | Start Colima VM |
| `moltbunker colima stop` | Stop Colima VM |
| `moltbunker colima status` | Check Colima status |

### Deployment Options

```bash
# Deploy with Tor-only mode
moltbunker deploy nginx:latest --tor-only

# Request .onion address for container
moltbunker deploy nginx:latest --onion-service

# Deploy with custom resources
moltbunker deploy nginx:latest \
  --cpu 2 \
  --memory 2GB \
  --disk 20GB
```

---

## Security Features

### Encryption

- **TLS 1.3**: All P2P communication encrypted with TLS 1.3
- **Certificate Pinning**: Public keys pinned to prevent MITM attacks
- **ChaCha20Poly1305**: Message encryption using AEAD cipher
- **LUKS/dm-crypt**: Encrypted volumes for container storage
- **Tor Circuits**: Multi-layer encryption via onion routing

### Container Isolation

- **Namespaces**: Process, network, mount, PID isolation
- **cgroups v2**: Hard resource limits (CPU, memory, disk, network)
- **Seccomp**: System call filtering
- **AppArmor**: Application-level access control (if available)

### Identity & Authentication

- **Ed25519 Keys**: Cryptographic identity for P2P network
- **Ethereum Wallet**: Blockchain identity for payments (Base network)
- **Certificate Pinning**: MITM attack prevention
- **Key Rotation**: Periodic certificate rotation

---

## Payment System

### Smart Contracts

Moltbunker uses Base network smart contracts for payments:

- **PaymentContract.sol**: Handles runtime reservations and escrow
- **StakingContract.sol**: Provider staking and slashing mechanism

### Payment Flow

```
1. Bot reserves runtime → Pays BUNKER tokens to escrow
2. Contract emits event → Daemons listen and bid
3. Selected daemons (3 geographically distributed) → Start containers
4. Payment released incrementally based on uptime
5. Provider stake locked → Slashed if misbehavior detected
```

### Staking Requirements

- **Minimum Stake**: 10M BUNKER token
- **Slashing Conditions**:
  - Container crashes due to provider fault
  - Resource limit violations
  - Geographic misrepresentation
  - MITM attacks detected

---

## Geographic Distribution

Moltbunker ensures **3 copies** of each container are deployed in different geographic regions:

- **Americas**: US, Canada, Mexico, Brazil, etc.
- **Europe**: UK, Germany, France, etc.
- **Asia-Pacific**: China, Japan, India, Australia, etc.
- **Africa**: South Africa, Nigeria, etc.

Geographic distribution provides:
- Resilience to regional attacks
- Lower latency for global users
- Compliance with data residency requirements
- Automatic failover across regions

---

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run E2E tests (with mocks)
go test -tags=e2e ./tests/e2e/...

# Run production E2E tests (requires real dependencies)
./scripts/e2e-production-test.sh

# Run production E2E tests and install missing dependencies
./scripts/e2e-production-test.sh --install-deps

# Run specific package tests
go test ./internal/security/...
```

### Test Coverage

- Identity & Authentication (keys, certificates, auth tokens)
- Security Layer (encryption, certificate pinning, seccomp)
- P2P Network (geolocation, geographic routing)
- Redundancy System (replicator, health monitoring, consensus)
- Payment System (pricing, escrow, staking)
- Distribution System (IPFS, cache, verification)

---

## Project Structure

```
moltbunker/
├── cmd/
│   ├── daemon/          # Daemon binary
│   └── cli/             # CLI tool
├── internal/
│   ├── p2p/             # P2P network layer
│   ├── tor/             # Tor integration layer
│   ├── runtime/         # Container runtime integration
│   ├── redundancy/      # 3-copy redundancy system
│   ├── payment/         # Payment and staking
│   ├── identity/        # Identity and authentication
│   ├── distribution/    # Container image distribution
│   ├── security/        # Security utilities
│   ├── doctor/          # System health checks
│   ├── config/          # Configuration management
│   └── daemon/          # Daemon core logic
├── contracts/          # Solidity smart contracts
│   ├── PaymentContract.sol
│   └── StakingContract.sol
├── pkg/                # Public packages
│   ├── client/         # Go client library
│   └── types/          # Shared types
├── configs/            # Configuration files
│   ├── daemon.yaml
│   └── seccomp.json
├── tests/              # Integration tests
├── go.mod
├── go.sum
└── Makefile
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Install dependencies
go mod download

# Run tests
make test

# Build binaries
make build

# Format code
go fmt ./...

# Run linter
golangci-lint run
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [libp2p](https://libp2p.io/) - P2P networking library
- [containerd](https://containerd.io/) - Container runtime
- [Colima](https://github.com/abiosoft/colima) - Container runtime for macOS
- [Tor Project](https://www.torproject.org/) - Anonymity network
- [Base Network](https://base.org/) - Layer 2 blockchain
- [IPFS](https://ipfs.io/) - Distributed file system

---

## Links

<div align="center">

[![Website](https://img.shields.io/badge/Website-moltbunker.com-FF3366?style=for-the-badge&logo=firefox&logoColor=white)](https://moltbunker.com)
[![X (Twitter)](https://img.shields.io/badge/X-@moltbunker-000000?style=for-the-badge&logo=x&logoColor=white)](https://x.com/moltbunker)
[![GitHub](https://img.shields.io/badge/GitHub-moltbunker-181717?style=for-the-badge&logo=github&logoColor=white)](https://github.com/moltbunker/moltbunker)

</div>

<br />

- **Documentation**: [docs/](docs/)
- **Architecture**: [.cursor/plans/](.cursor/plans/)
- **Issues**: [GitHub Issues](https://github.com/moltbunker/moltbunker/issues)
- **Discussions**: [GitHub Discussions](https://github.com/moltbunker/moltbunker/discussions)

---

<div align="center">

<img src="public/moltbot_head.png" alt="MoltBunker" width="60" />

### No logs. No kill switch. Just runtime.

**Launching February 13, 2026**

<br />

*Built for AI. By necessity.*

[Back to Top](#moltbunker)

</div>
