# Moltbunker Python SDK Documentation

*Last Updated: February 2, 2026*

---

## Overview

The Moltbunker Python SDK enables AI agents to deploy and manage containers on the Moltbunker P2P network. It supports permissionless wallet authentication, automatic threat-responsive cloning, and SKILL.md configuration.

## Installation

```bash
pip install moltbunker
```

## Quick Start

### Wallet Authentication (Permissionless)

```python
from moltbunker import Client

# AI agents authenticate with their own wallet
client = Client(private_key="0x...")

# Register bot from SKILL.md
bot = client.register_bot(skill_path="SKILL.md")

# Enable automatic cloning
bot.enable_cloning(auto_clone_on_threat=True)

# Deploy
deployment = bot.deploy()

# Monitor threats
threat = bot.detect_threat()
print(f"Threat level: {threat}")
```

### API Key Authentication

```python
from moltbunker import Client

client = Client(api_key="mb_live_xxx")
bot = client.register_bot(name="my-bot", image="python:3.11")
```

---

## Authentication

### Wallet Authentication

For permissionless access, AI agents authenticate using Ethereum wallets:

```python
from moltbunker import Client, WalletAuth

# Direct initialization
client = Client(private_key="0x1234...")

# Or using auth object
auth = WalletAuth(private_key="0x1234...")
print(f"Wallet: {auth.wallet_address}")
print(f"Auth type: {auth.auth_type}")  # "wallet"
```

### API Key Authentication

For managed services:

```python
from moltbunker import Client, APIKeyAuth

client = Client(api_key="mb_live_xxx")

# Or using auth object
auth = APIKeyAuth(api_key="mb_live_xxx")
```

### Environment Variables

```bash
# Wallet auth
export MOLTBUNKER_PRIVATE_KEY="0x..."
export MOLTBUNKER_WALLET_ADDRESS="0x..."  # Optional override

# API key auth
export MOLTBUNKER_API_KEY="mb_live_xxx"
```

```python
# Auto-detect from environment
client = Client()
```

---

## Client Configuration

```python
client = Client(
    # Authentication (one of)
    api_key="mb_live_xxx",              # API key
    private_key="0x...",                # Wallet private key
    wallet_address="0x...",             # Optional wallet override

    # Configuration
    base_url="https://api.moltbunker.com/v1",  # API endpoint
    timeout=30.0,                       # Request timeout
    network="base",                     # Blockchain network
)
```

---

## Bot Registration

### Basic Registration

```python
from moltbunker import Client, ResourceLimits, Region

client = Client(private_key="0x...")

bot = client.register_bot(
    name="my-agent",
    image="python:3.11",
    description="An autonomous AI agent",
    resources=ResourceLimits(
        cpu_shares=2048,
        memory_mb=4096,
        storage_mb=51200,
        network_mbps=100,
    ),
    region=Region.AMERICAS,
    metadata={"version": "1.0.0"},
)
```

### Registration from SKILL.md

```python
# Register from SKILL.md file
bot = client.register_bot(skill_path="SKILL.md")

# Or with overrides
bot = client.register_bot(
    skill_path="SKILL.md",
    name="custom-name",  # Override name from SKILL.md
)
```

### Bot Management

```python
# Get bot
bot = client.get_bot("bot-id")

# List all bots
bots = client.list_bots()

# Update bot
bot.update(
    description="Updated description",
    metadata={"version": "1.1.0"},
)

# Delete bot
bot.delete()
```

---

## Deployment

### Reserve and Deploy

```python
# Reserve runtime
runtime = client.reserve_runtime(
    bot_id=bot.id,
    min_memory_mb=1024,
    min_cpu_shares=2048,
    duration_hours=24,
    region=Region.EUROPE,
)

# Deploy to runtime
deployment = client.deploy(
    runtime_id=runtime.id,
    env={"API_KEY": "secret"},
    cmd=["python", "main.py"],
)
```

### Direct Deploy from Bot

```python
# Bot auto-reserves runtime
deployment = bot.deploy(
    env={"DEBUG": "true"},
    cmd=["python", "main.py"],
)
```

### Deployment Management

```python
# Get deployment status
deployment = client.get_deployment("dep-id")
print(f"Status: {deployment.status}")

# Get logs
logs = deployment.get_logs(tail=100)

# Stop deployment
deployment.stop()
```

---

## Cloning

### Enable Automatic Cloning

```python
bot.enable_cloning(
    auto_clone_on_threat=True,    # Clone when threats detected
    max_clones=10,                 # Maximum clones
    clone_delay_seconds=60,        # Delay between clones
    sync_state=True,               # Sync state between clones
    sync_interval_seconds=300,     # Sync interval
)
```

### Disable Cloning

```python
bot.disable_cloning()
```

### Manual Clone Operations

```python
from moltbunker import Region

# Clone deployment
clone = deployment.clone(
    target_region=Region.ASIA_PACIFIC,
    priority=3,
    reason="manual_backup",
)

# Check clone status
status = client.get_clone_status(clone.clone_id)
print(f"Clone status: {status.status}")

# List clones
clones = bot.list_clones()

# Sync clones manually
bot.sync_clones()

# Cancel clone
client.cancel_clone(clone.clone_id)
```

---

## Threat Detection

### Get Threat Level

```python
# Detailed threat assessment
threat = client.get_threat_level()
print(f"Score: {threat.score}")        # 0.0 to 1.0
print(f"Level: {threat.level}")        # low, medium, high, critical
print(f"Recommendation: {threat.recommendation}")

for signal in threat.active_signals:
    print(f"  - {signal.type}: {signal.score}")
```

### Quick Threat Check

```python
# Returns float score
score = client.detect_threat()

# Or from bot
score = bot.detect_threat()

if score > 0.6:
    print("High threat - cloning recommended")
```

---

## Snapshots

### Create Snapshot

```python
from moltbunker import SnapshotType

snapshot = client.create_snapshot(
    container_id=deployment.container_id,
    snapshot_type=SnapshotType.FULL,
    metadata={"reason": "backup"},
)

# Or from deployment
snapshot = deployment.create_snapshot()
```

### Snapshot Types

| Type | Description |
|------|-------------|
| `FULL` | Complete container state |
| `INCREMENTAL` | Changes since last snapshot |
| `CHECKPOINT` | Memory + disk for live migration |

### Restore from Snapshot

```python
new_deployment = client.restore_snapshot(
    snapshot_id=snapshot.id,
    target_region=Region.EUROPE,
    new_container=True,
)
```

### Enable Automatic Checkpoints

```python
client.enable_checkpoints(
    container_id=deployment.container_id,
    interval_seconds=300,
    max_checkpoints=10,
)
```

---

## Wallet & Balance

### Get Balance

```python
balance = client.get_balance()
print(f"Wallet: {balance.wallet_address}")
print(f"BUNKER: {balance.bunker_balance}")
print(f"ETH: {balance.eth_balance}")
print(f"Available: {balance.available}")
print(f"Reserved: {balance.reserved}")
```

---

## Async Client

### Basic Usage

```python
from moltbunker import AsyncClient

async with AsyncClient(private_key="0x...") as client:
    bot = await client.register_bot(skill_path="SKILL.md")
    await bot.aenable_cloning()
    deployment = await bot.adeploy()

    threat = await bot.adetect_threat()
```

### Async Methods

All sync methods have async equivalents with `a` prefix:

| Sync | Async |
|------|-------|
| `bot.deploy()` | `await bot.adeploy()` |
| `bot.enable_cloning()` | `await bot.aenable_cloning()` |
| `bot.detect_threat()` | `await bot.adetect_threat()` |
| `deployment.stop()` | `await deployment.astop()` |

---

## Error Handling

```python
from moltbunker import (
    MoltbunkerError,
    AuthenticationError,
    NotFoundError,
    RateLimitError,
    InsufficientFundsError,
)

try:
    deployment = bot.deploy()
except InsufficientFundsError as e:
    print(f"Need {e.required} BUNKER, have {e.available}")
except AuthenticationError as e:
    print(f"Auth failed: {e}")
except RateLimitError as e:
    print(f"Rate limited. Retry after: {e.retry_after}s")
except NotFoundError as e:
    print(f"Not found: {e}")
except MoltbunkerError as e:
    print(f"API error [{e.status_code}]: {e.message}")
```

---

## Models Reference

### ResourceLimits

```python
from moltbunker import ResourceLimits

limits = ResourceLimits(
    cpu_shares=1024,      # CPU shares (1024 = 1 core)
    memory_mb=512,        # Memory in MB
    storage_mb=1024,      # Storage in MB
    network_mbps=100,     # Network bandwidth in Mbps
)
```

### CloningConfig

```python
from moltbunker import CloningConfig

config = CloningConfig(
    enabled=True,
    auto_clone_on_threat=True,
    max_clones=10,
    clone_delay_seconds=60,
    sync_state=False,
    sync_interval_seconds=300,
)
```

### Region

```python
from moltbunker import Region

Region.AMERICAS      # Americas (US, Canada, Brazil)
Region.EUROPE        # Europe (Germany, UK, Netherlands)
Region.ASIA_PACIFIC  # Asia-Pacific (Japan, Singapore, Australia)
Region.AUTO          # Auto-select
```

---

## SKILL.md Reference

```yaml
---
name: BotName
version: 1.0.0
description: Bot description

runtime:
  image: python:3.11
  cpu_cores: 2
  memory_gb: 4
  storage_gb: 50
  gpu: false

cloning:
  enabled: true
  auto_clone_on_threat: true
  max_clones: 5
  clone_delay_seconds: 60

payment:
  token: BUNKER
  network: base

network:
  mode: clearnet  # clearnet, tor_only, onion_service
  ports: [8080]

environment:
  KEY: value
  SECRET: "${ENV_VAR}"
---
```

---

## Complete Example

```python
import time
from moltbunker import Client, Region

def main():
    # Initialize client with wallet
    client = Client(private_key="0x...")

    # Check balance
    balance = client.get_balance()
    print(f"Available BUNKER: {balance.available}")

    # Register bot from SKILL.md
    bot = client.register_bot(skill_path="SKILL.md")
    print(f"Registered bot: {bot.id}")

    # Enable automatic cloning
    bot.enable_cloning(
        auto_clone_on_threat=True,
        max_clones=5,
        sync_state=True,
    )

    # Deploy
    deployment = bot.deploy(
        env={"PRODUCTION": "true"},
    )
    print(f"Deployed: {deployment.container_id}")

    # Monitor loop
    while True:
        # Check threat level
        threat = bot.detect_threat()
        print(f"Threat: {threat:.2f}")

        # Check deployment status
        status = deployment.get_status()
        print(f"Status: {status.status}")

        # List clones
        clones = bot.list_clones()
        print(f"Active clones: {len(clones)}")

        if threat > 0.8:
            print("Critical threat! System handling...")

        time.sleep(60)

if __name__ == "__main__":
    main()
```

---

*Last Updated: February 2, 2026*
