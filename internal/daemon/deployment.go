package daemon

import (
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// mutexType is an alias for sync.Mutex used in pendingDeployment
type mutexType = sync.Mutex

// Deployment represents a deployed container with its metadata
type Deployment struct {
	ID              string                 `json:"id"`
	Image           string                 `json:"image"`
	Status          types.ContainerStatus  `json:"status"`
	Resources       types.ResourceLimits   `json:"resources"`
	CreatedAt       time.Time              `json:"created_at"`
	StartedAt       time.Time              `json:"started_at,omitempty"`
	Encrypted       bool                   `json:"encrypted"`
	EncryptedVolume string                 `json:"encrypted_volume,omitempty"`
	OnionService    bool                   `json:"onion_service"`
	OnionAddress    string                 `json:"onion_address,omitempty"`
	OnionPort       int                    `json:"onion_port,omitempty"` // Port exposed via Tor
	TorOnly         bool                   `json:"tor_only"`
	ReplicaSet      *types.ReplicaSet      `json:"replica_set,omitempty"`
	LocalReplica    int                    `json:"local_replica"`
	Regions         []string               `json:"regions"`
	Locations       []ReplicaLocation      `json:"locations,omitempty"` // Detailed location per replica
	OriginatorID    types.NodeID           `json:"originator_id"`       // Node that originated the deployment
	Owner           string                 `json:"owner,omitempty"`     // Wallet address of the deployer
	ExecAgentEnabled bool                  `json:"exec_agent_enabled,omitempty"` // True if exec-agent binary is injected (E2E encrypted exec)
	ExecKeyPath      string               `json:"exec_key_path,omitempty"`      // Host path to exec_key file (for cleanup)
	DeployNonce      string               `json:"deploy_nonce,omitempty"`       // Hex-encoded deploy nonce for exec_key re-derivation
	ExposedPorts     []ExposedPort         `json:"exposed_ports,omitempty"`      // Ports exposed publicly via ingress
	PublicURLs       []string              `json:"public_urls,omitempty"`        // Generated public URLs
	MinProviderTier  types.ProviderTier    `json:"min_provider_tier,omitempty"`  // Minimum provider tier required
	StoppedAt        time.Time             `json:"stopped_at,omitempty"`         // When container was stopped
	VolumeExpiresAt  time.Time             `json:"volume_expires_at,omitempty"`  // When volume will be auto-deleted
}

// pendingDeployment tracks replica acknowledgments for a deployment
type pendingDeployment struct {
	containerID      string
	ackChan          chan replicaAck
	ackCount         int
	successCount     int
	acks             []replicaAck
	mu               mutexType
	created          time.Time
	closeOnce        sync.Once // Ensures ackChan is closed exactly once
	closed           bool      // tracks if ackChan has been closed (for readers)
	escrowActivated  bool      // true once SelectProviders has been called
}

// replicaAck represents an acknowledgment from a replica node
type replicaAck struct {
	NodeID  string
	Region  string
	Success bool
	Error   string
}

// DeployResult contains the result of a deployment including replica count
type DeployResult struct {
	Deployment   *Deployment
	ReplicaCount int
}

// close safely closes the ack channel exactly once.
// Both the closed flag and channel close happen under the lock to prevent
// a race with handleDeployAck's non-blocking send.
func (p *pendingDeployment) close() {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		p.closed = true
		close(p.ackChan)
		p.mu.Unlock()
	})
}

// isClosed returns whether the channel has been closed
func (p *pendingDeployment) isClosed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closed
}

// ContainerManagerConfig contains configuration for the container manager
type ContainerManagerConfig struct {
	DataDir          string
	ContainerdSocket string
	RuntimeName      string // "auto" or explicit OCI runtime name
	KataConfig       *runtime.KataConfig
	TorDataDir       string
	EnableEncryption bool
	PaymentService   *payment.PaymentService
}
