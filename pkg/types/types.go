package types

import (
	"crypto/ed25519"
	"encoding/hex"
	"time"
)

// NodeID is a SHA-256 hash of Ed25519 public key
type NodeID [32]byte

// String returns the NodeID as a hex string
func (n NodeID) String() string {
	return hex.EncodeToString(n[:])
}

// Node represents a peer in the P2P network
type Node struct {
	ID          NodeID
	PublicKey   ed25519.PublicKey
	Address     string // Clearnet address or .onion address
	Port        int
	Region      string // Geographic region
	Country     string
	OnionAddr   string // .onion address if available
	LastSeen    time.Time
	Capabilities NodeCapabilities
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

// ReplicaSet represents 3 copies of a container
type ReplicaSet struct {
	ContainerID string
	Replicas    [3]*Container
	Region1     string
	Region2     string
	Region3     string
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
)

// Message represents a P2P message
type Message struct {
	Type      MessageType
	From      NodeID
	To        NodeID
	Payload   []byte
	Timestamp time.Time
	Nonce     [24]byte // For ChaCha20Poly1305
}

// DeploymentRequest represents a container deployment request
type DeploymentRequest struct {
	ImageCID    string
	Resources   ResourceLimits
	NetworkMode NetworkMode
	TorOnly     bool
	OnionService bool
	PaymentHash string // Transaction hash for payment
}

// Bid represents a daemon's bid for hosting a container
type Bid struct {
	NodeID    NodeID
	Region    string
	Price     uint64 // Price in BUNKER tokens
	Resources ResourceLimits
	Stake     uint64 // Staked amount
}
