# Moltbunker Deployment Guide

*Last Updated: February 2, 2026*

This guide covers deploying Moltbunker in various environments.

## Table of Contents

- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Daemon](#running-the-daemon)
- [Docker Deployment](#docker-deployment)
- [Production Considerations](#production-considerations)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## Quick Start

```bash
# Install
moltbunker install

# Start daemon
moltbunker start

# Check status
moltbunker status

# Deploy a container
moltbunker deploy nginx:latest --onion-service
```

## Prerequisites

### System Requirements

- **OS**: Linux (recommended), macOS, Windows
- **CPU**: 2+ cores
- **RAM**: 4GB minimum, 8GB recommended
- **Disk**: 50GB+ available space
- **Network**: Public IP or NAT traversal capable

### Software Dependencies

- **containerd** (1.7+): Container runtime
- **Tor** (optional): For anonymous networking
- **IPFS** (optional): For container image distribution

### Install Dependencies (Ubuntu/Debian)

```bash
# Install containerd
sudo apt-get update
sudo apt-get install containerd

# Install Tor (optional)
sudo apt-get install tor

# Install IPFS (optional)
wget https://dist.ipfs.tech/kubo/v0.27.0/kubo_v0.27.0_linux-amd64.tar.gz
tar xvfz kubo_v0.27.0_linux-amd64.tar.gz
sudo ./kubo/install.sh
ipfs init
```

### Install Dependencies (macOS)

```bash
# Using Homebrew
brew install containerd tor ipfs
```

## Installation

### From Binary

```bash
# Download latest release
curl -L https://github.com/moltbunker/moltbunker/releases/latest/download/moltbunker_linux_amd64.tar.gz | tar xz

# Install binaries
sudo mv moltbunker /usr/local/bin/
sudo mv moltbunker-daemon /usr/local/bin/

# Initialize
moltbunker install
```

### From Source

```bash
# Clone repository
git clone https://github.com/moltbunker/moltbunker.git
cd moltbunker

# Build
make build

# Install
sudo make install

# Initialize
moltbunker install
```

### Using Homebrew (macOS)

```bash
brew tap moltbunker/tap
brew install moltbunker
moltbunker install
```

## Configuration

Configuration file location: `~/.moltbunker/config.yaml`

### Basic Configuration

```yaml
daemon:
  port: 9000
  data_dir: ~/.moltbunker
  log_level: info

p2p:
  bootstrap_nodes: []
  network_mode: hybrid  # clearnet, tor_only, hybrid

tor:
  enabled: false
  socks_port: 9050
  control_port: 9051

container:
  default_cpu_quota: 100000
  default_memory_limit: 536870912  # 512MB
  default_disk_limit: 10737418240  # 10GB

redundancy:
  replica_count: 3
  health_check_interval: 30s
```

### Environment Variables

See `.env.example` for all available environment variables.

```bash
export MOLTBUNKER_PORT=9000
export MOLTBUNKER_TOR_ENABLED=true
export MOLTBUNKER_LOG_LEVEL=debug
```

## Running the Daemon

### Direct Start

```bash
# Start in foreground
moltbunker start --foreground

# Start in background
moltbunker start
```

### Using systemd (Linux)

```bash
# Install service
sudo moltbunker install --systemd

# Start service
sudo systemctl start moltbunker
sudo systemctl enable moltbunker

# Check status
sudo systemctl status moltbunker
```

### Using launchd (macOS)

```bash
# Install service
moltbunker install --launchd

# Start service
launchctl start com.moltbunker.daemon

# Check status
launchctl list | grep moltbunker
```

## Docker Deployment

### Using Docker Compose

```bash
# Production deployment
docker-compose up -d

# View logs
docker-compose logs -f moltbunker-daemon

# Check status
docker exec moltbunker-daemon moltbunker status
```

### Kubernetes (Helm)

```bash
# Add Helm repository
helm repo add moltbunker https://charts.moltbunker.com

# Install
helm install moltbunker moltbunker/moltbunker \
  --set tor.enabled=true \
  --set replicas=3
```

## Production Considerations

### Security

1. **Firewall**: Open only port 9000 (P2P) and optionally 9050 (Tor SOCKS)
2. **TLS**: All P2P communication uses TLS 1.3
3. **Certificate Pinning**: Enabled by default for MITM protection
4. **Container Isolation**: Uses namespaces, cgroups, and seccomp

### High Availability

- Run 3+ nodes across different geographic regions
- Automatic failover via consensus mechanism
- 3-copy redundancy for all containers

### Backup

```bash
# Backup node keys (critical!)
cp -r ~/.moltbunker/keys /secure/backup/

# Backup configuration
cp ~/.moltbunker/config.yaml /secure/backup/
```

### Scaling

- Horizontal scaling: Add more nodes to the network
- Vertical scaling: Increase resources per node
- Load balancing: Automatic via DHT-based routing

## Monitoring

### Built-in Monitor

```bash
# Real-time dashboard
moltbunker monitor

# JSON status
moltbunker status --json
```

### Prometheus Metrics

Metrics endpoint: `http://localhost:9000/metrics`

Key metrics:
- `moltbunker_peers_total`: Number of connected peers
- `moltbunker_containers_running`: Running container count
- `moltbunker_bandwidth_bytes`: Network bandwidth usage
- `moltbunker_tor_circuits`: Active Tor circuits

### Grafana Dashboard

Import dashboard ID: `xxxxx` (coming soon)

## Troubleshooting

### Daemon Won't Start

```bash
# Check logs
tail -f ~/.moltbunker/logs/daemon.log

# Verify port availability
lsof -i :9000

# Check containerd
sudo systemctl status containerd
```

### No Peers Found

```bash
# Check bootstrap nodes
moltbunker config get p2p.bootstrap_nodes

# Verify network connectivity
curl -I https://api.ipify.org

# Check firewall
sudo ufw status
```

### Tor Not Working

```bash
# Check Tor status
moltbunker tor status

# Verify Tor service
sudo systemctl status tor

# Test SOCKS proxy
curl --socks5 localhost:9050 https://check.torproject.org/api/ip
```

### Container Deployment Fails

```bash
# Check containerd
sudo ctr --address /run/containerd/containerd.sock version

# View deployment logs
moltbunker logs <container-id>

# Check IPFS
ipfs id
```

## Support

- GitHub Issues: https://github.com/moltbunker/moltbunker/issues
- Documentation: https://docs.moltbunker.com
- Discord: https://discord.gg/moltbunker
