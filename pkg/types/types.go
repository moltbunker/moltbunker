package types

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// NodeID is a SHA-256 hash of Ed25519 public key
type NodeID [32]byte

// String returns the NodeID as a hex string
func (n NodeID) String() string {
	return hex.EncodeToString(n[:])
}

// NodeLocation contains detailed geographic information for a node
type NodeLocation struct {
	Region      string  `json:"region"`                 // Broad region: Americas, Europe, Asia-Pacific, Middle-East, Africa
	Country     string  `json:"country"`                // ISO 3166-1 alpha-2 code (e.g., "FR")
	CountryName string  `json:"country_name,omitempty"` // Full country name (e.g., "France")
	City        string  `json:"city,omitempty"`          // City name (e.g., "Strasbourg")
	Lat         float64 `json:"lat,omitempty"`           // Latitude
	Lon         float64 `json:"lon,omitempty"`           // Longitude
}

// Node represents a peer in the P2P network
type Node struct {
	ID            NodeID
	PublicKey     ed25519.PublicKey
	Address       string         // Clearnet address or .onion address
	Port          int
	Region        string         // Geographic region (kept for backward compat)
	Country       string         // ISO 3166-1 alpha-2 (kept for backward compat)
	Location      NodeLocation   // Detailed location info
	OnionAddr     string         // .onion address if available
	LastSeen      time.Time
	WalletAddress common.Address // Ethereum wallet address (set after announce)
	ProviderTier  ProviderTier  // Provider isolation tier (set after announce)
	Capabilities  NodeCapabilities
}

// NodeCapabilities describes what a node can do
type NodeCapabilities struct {
	ContainerRuntime bool
	TorSupport       bool
	StakingAmount    uint64 // BUNKER tokens staked
}

// Container represents a running container instance
type Container struct {
	ID          string
	ImageCID    string // IPFS CID of container image
	NodeID      NodeID
	Status      ContainerStatus
	Resources   ResourceLimits
	CreatedAt   time.Time
	Health      HealthStatus
	Encrypted   bool
	OnionAddr   string // .onion address if Tor-enabled
}

// ContainerStatus represents container state
type ContainerStatus string

const (
	ContainerStatusPending     ContainerStatus = "pending"
	ContainerStatusCreated     ContainerStatus = "created"
	ContainerStatusRunning     ContainerStatus = "running"
	ContainerStatusPaused      ContainerStatus = "paused"
	ContainerStatusStopped     ContainerStatus = "stopped"
	ContainerStatusFailed      ContainerStatus = "failed"
	ContainerStatusReplicating ContainerStatus = "replicating"
)

// ResourceLimits defines resource boundaries
type ResourceLimits struct {
	CPUQuota    int64  // CPU quota in microseconds
	CPUPeriod   uint64 // CPU period in microseconds
	MemoryLimit int64  // Memory limit in bytes
	DiskLimit   int64  // Disk limit in bytes
	NetworkBW   int64  // Network bandwidth in bytes/sec
	PIDLimit    int    // Max process count
}

// HealthStatus contains health metrics
type HealthStatus struct {
	CPUUsage    float64
	MemoryUsage int64
	DiskUsage   int64
	NetworkIn   int64
	NetworkOut  int64
	LastUpdate  time.Time
	Healthy     bool
}

// ReplicaSet represents 1-3 copies of a container across regions
type ReplicaSet struct {
	ContainerID string
	Replicas    []*Container
	Regions     []string
	CreatedAt   time.Time
}

// NetworkMode represents network connectivity mode
type NetworkMode string

const (
	NetworkModeClearnet NetworkMode = "clearnet"
	NetworkModeTorOnly  NetworkMode = "tor_only"
	NetworkModeHybrid   NetworkMode = "hybrid"
	NetworkModeTorExit  NetworkMode = "tor_exit"
)

// ProtocolVersion is the current protocol version for P2P messages
const ProtocolVersion uint32 = 1

// MinSupportedVersion is the minimum protocol version accepted from peers
const MinSupportedVersion uint32 = 1

// Message types for P2P communication
type MessageType string

const (
	MessageTypePing            MessageType = "ping"
	MessageTypePong            MessageType = "pong"
	MessageTypeFindNode        MessageType = "find_node"
	MessageTypeNodes           MessageType = "nodes"
	MessageTypeHealth          MessageType = "health"
	MessageTypeDeploy          MessageType = "deploy"
	MessageTypeDeployAck       MessageType = "deploy_ack"  // Acknowledgment of deployment
	MessageTypeBid             MessageType = "bid"
	MessageTypeGossip          MessageType = "gossip"
	MessageTypeContainerStatus MessageType = "container_status"
	MessageTypeReplicaSync     MessageType = "replica_sync"
	MessageTypeLogs            MessageType = "logs"
	MessageTypeStop            MessageType = "stop"
	MessageTypeDelete          MessageType = "delete"

	// Exec message types for interactive terminal sessions
	MessageTypeExecOpen   MessageType = "exec_open"   // Request to open exec stream
	MessageTypeExecData   MessageType = "exec_data"   // Bidirectional encrypted terminal data
	MessageTypeExecResize MessageType = "exec_resize" // Terminal resize event
	MessageTypeExecClose  MessageType = "exec_close"  // Close exec stream

	// Identity exchange
	MessageTypeAnnounce MessageType = "announce" // Wallet ownership proof after TLS handshake
)

// Message represents a P2P message
type Message struct {
	Type      MessageType `json:"type"`
	From      NodeID      `json:"from"`
	To        NodeID      `json:"to"`
	Payload   []byte      `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
	Nonce     [24]byte    `json:"nonce"`
	Signature []byte      `json:"signature,omitempty"`
	Version   uint32      `json:"version"`
}

// SignableBytes returns a deterministic byte representation of the message
// fields used for signing. The Signature field itself is excluded to avoid
// circular dependency during sign/verify.
func (m *Message) SignableBytes() []byte {
	signable := struct {
		Type      MessageType `json:"type"`
		From      NodeID      `json:"from"`
		To        NodeID      `json:"to"`
		Payload   []byte      `json:"payload"`
		Timestamp time.Time   `json:"timestamp"`
		Nonce     [24]byte    `json:"nonce"`
		Version   uint32      `json:"version"`
	}{
		Type:      m.Type,
		From:      m.From,
		To:        m.To,
		Payload:   m.Payload,
		Timestamp: m.Timestamp,
		Nonce:     m.Nonce,
		Version:   m.Version,
	}

	data, err := json.Marshal(signable)
	if err != nil {
		// This should never fail for these types, but return nil to signal error
		return nil
	}
	return data
}

// DeploymentRequest represents a container deployment request
type DeploymentRequest struct {
	ImageCID     string
	Resources    ResourceLimits
	NetworkMode  NetworkMode
	TorOnly      bool
	OnionService bool
	PaymentHash  string // Transaction hash for payment

	// Minimum provider tier required for this deployment
	MinProviderTier ProviderTier `json:"min_provider_tier,omitempty"`

	// Exec encryption (optional — omit to disable exec for this container)
	DeployNonce      string `json:"deploy_nonce,omitempty"`
	EncryptedExecKey []byte `json:"encrypted_exec_key,omitempty"`
	ExecKeyNonce     []byte `json:"exec_key_nonce,omitempty"`
}

// ExecOpenPayload is sent from API node to provider to request an exec stream
type ExecOpenPayload struct {
	ContainerID   string `json:"container_id"`
	SessionID     string `json:"session_id"`
	WalletAddress string `json:"wallet_address"`
	Cols          int    `json:"cols"`
	Rows          int    `json:"rows"`
}

// ExecDataPayload carries encrypted terminal I/O bytes (opaque to relays)
type ExecDataPayload struct {
	SessionID string `json:"session_id"`
	Data      []byte `json:"data"`
}

// ExecResizePayload carries terminal dimension changes
type ExecResizePayload struct {
	SessionID string `json:"session_id"`
	Cols      int    `json:"cols"`
	Rows      int    `json:"rows"`
}

// ExecClosePayload signals exec session termination
type ExecClosePayload struct {
	SessionID string `json:"session_id"`
	Reason    string `json:"reason"`
}

// AnnouncePayload is sent after TLS handshake to prove wallet ownership.
// The sender signs a message binding their NodeID to their wallet address.
type AnnouncePayload struct {
	WalletAddress string `json:"wallet_address"` // 0x-prefixed checksummed Ethereum address
	NodeID        string `json:"node_id"`         // Hex-encoded node ID (binds wallet to node)
	Timestamp     int64  `json:"timestamp"`       // Unix timestamp (must be within 5 min)
	Nonce         string `json:"nonce"`            // Random 32-byte hex string
	EthSignature  string       `json:"eth_signature"`              // EIP-191 personal_sign (65 bytes, hex)
	ProviderTier  ProviderTier `json:"provider_tier,omitempty"`    // Self-reported provider tier
}

// Bid represents a daemon's bid for hosting a container
type Bid struct {
	NodeID    NodeID
	Region    string
	Price     uint64 // Price in BUNKER tokens
	Resources ResourceLimits
	Stake     uint64 // Staked amount
}
