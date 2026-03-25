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

### 🧪 Public Testnet Trials (Free BUNKER)

Before mainnet, developers can run end-to-end reservation flows on Base Sepolia using free testnet BUNKER.

- **Goal:** validate onboarding, reservation UX, and token economics with real users
- **Cost:** free (testnet-only tokens)
- **Network:** Base Sepolia
- **Feedback loop:** automated collection from trial runs to surface friction and bugs early

Start at the docs: [https://moltbunker.com/docs](https://moltbunker.com/docs)

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
├─ Replay protection: 24-byte nonc
```