package daemon

import (
	"sync"
	"time"

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
}

// pendingDeployment tracks replica acknowledgments for a deployment
type pendingDeployment struct {
	containerID  string
	ackChan      chan replicaAck
	ackCount     int
	successCount int
	acks         []replicaAck
	mu           mutexType
	created      time.Time
	closeOnce    sync.Once // Ensures ackChan is closed exactly once
	closed       bool      // tracks if ackChan has been closed (for readers)
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

// close safely closes the ack channel exactly once
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
	TorDataDir       string
	EnableEncryption bool
}
