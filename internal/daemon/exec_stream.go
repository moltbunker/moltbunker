package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// ExecStream bridges a remote exec request to a local container PTY session.
// It manages the bidirectional relay between P2P messages and the interactive session.
type ExecStream struct {
	sessionID   string
	containerID string
	fromNode    types.NodeID
	session     *runtime.InteractiveSession
	router      interface{ SendMessage(ctx context.Context, to types.NodeID, msg *types.Message) error }
	localNodeID types.NodeID

	cancel    context.CancelFunc
	done      chan struct{}
	closeOnce sync.Once
}

// newExecStream creates a new exec stream bridging P2P to container PTY
func newExecStream(
	ctx context.Context,
	sessionID string,
	containerID string,
	fromNode types.NodeID,
	localNodeID types.NodeID,
	session *runtime.InteractiveSession,
	router interface{ SendMessage(ctx context.Context, to types.NodeID, msg *types.Message) error },
) *ExecStream {
	streamCtx, cancel := context.WithCancel(ctx)

	es := &ExecStream{
		sessionID:   sessionID,
		containerID: containerID,
		fromNode:    fromNode,
		session:     session,
		router:      router,
		localNodeID: localNodeID,
		cancel:      cancel,
		done:        make(chan struct{}),
	}

	// Start reading container stdout and forwarding as P2P messages
	go es.readLoop(streamCtx)

	return es
}

// readLoop reads from the container's stdout and sends data back to the requester
func (es *ExecStream) readLoop(ctx context.Context) {
	defer es.Close()

	// Close stdout when context is cancelled to unblock Read()
	// This prevents goroutine leak when PTY hangs (P2-6 fix)
	go func() {
		select {
		case <-ctx.Done():
		case <-es.session.Done():
		}
		es.session.Stdout.Close()
	}()

	buf := make([]byte, 4096)
	for {
		n, err := es.session.Stdout.Read(buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])

			payload := types.ExecDataPayload{
				SessionID: es.sessionID,
				Data:      data,
			}
			payloadBytes, _ := json.Marshal(payload)

			sendErr := es.router.SendMessage(ctx, es.fromNode, &types.Message{
				Type:      types.MessageTypeExecData,
				From:      es.localNodeID,
				To:        es.fromNode,
				Payload:   payloadBytes,
				Timestamp: time.Now(),
			})
			if sendErr != nil {
				logging.Debug("exec stream send failed",
					"session_id", es.sessionID,
					"error", sendErr.Error(),
					logging.Component("exec_stream"))
				return
			}
		}
		if err != nil {
			return
		}
	}
}

// WriteData writes incoming data to the container's stdin
func (es *ExecStream) WriteData(data []byte) error {
	_, err := es.session.Stdin.Write(data)
	return err
}

// Resize changes the PTY dimensions
func (es *ExecStream) Resize(cols, rows uint16) error {
	return es.session.Resize(cols, rows)
}

// Close terminates the exec stream and cleans up
func (es *ExecStream) Close() {
	es.closeOnce.Do(func() {
		es.cancel()
		es.session.Close()

		// Send close message to the requester
		payload := types.ExecClosePayload{
			SessionID: es.sessionID,
			Reason:    "session_ended",
		}
		payloadBytes, _ := json.Marshal(payload)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = es.router.SendMessage(ctx, es.fromNode, &types.Message{
			Type:      types.MessageTypeExecClose,
			From:      es.localNodeID,
			To:        es.fromNode,
			Payload:   payloadBytes,
			Timestamp: time.Now(),
		})

		close(es.done)
	})
}

// Done returns a channel that's closed when the stream ends
func (es *ExecStream) Done() <-chan struct{} {
	return es.done
}

// ExecStreamManager tracks active exec streams on a provider node
type ExecStreamManager struct {
	streams map[string]*ExecStream // sessionID → stream
	mu      sync.RWMutex

	// Limits
	maxStreamsPerContainer int
	maxTotalStreams        int
}

// NewExecStreamManager creates a new exec stream manager
func NewExecStreamManager() *ExecStreamManager {
	return &ExecStreamManager{
		streams:               make(map[string]*ExecStream),
		maxStreamsPerContainer: 3,
		maxTotalStreams:        20,
	}
}

// Add registers a new exec stream
func (m *ExecStreamManager) Add(stream *ExecStream) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.streams) >= m.maxTotalStreams {
		return fmt.Errorf("max total exec streams (%d) reached", m.maxTotalStreams)
	}

	// Count streams for this container
	count := 0
	for _, s := range m.streams {
		if s.containerID == stream.containerID {
			count++
		}
	}
	if count >= m.maxStreamsPerContainer {
		return fmt.Errorf("max exec streams (%d) for container %s reached", m.maxStreamsPerContainer, stream.containerID)
	}

	m.streams[stream.sessionID] = stream

	// Clean up when stream ends
	go func() {
		<-stream.Done()
		m.Remove(stream.sessionID)
	}()

	return nil
}

// Get returns an exec stream by session ID
func (m *ExecStreamManager) Get(sessionID string) (*ExecStream, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.streams[sessionID]
	return s, ok
}

// Remove removes an exec stream
func (m *ExecStreamManager) Remove(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.streams, sessionID)
}

// CloseAllForContainer closes all active exec streams for a specific container
func (m *ExecStreamManager) CloseAllForContainer(containerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, s := range m.streams {
		if s.containerID == containerID {
			s.Close()
			delete(m.streams, id)
		}
	}
}

// CloseAll closes all active exec streams
func (m *ExecStreamManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, s := range m.streams {
		s.Close()
		delete(m.streams, id)
	}
}

// Count returns the number of active streams
func (m *ExecStreamManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.streams)
}

// ExecRelay represents a requester-side exec session that relays P2P data to a WebSocket.
// When the provider sends ExecData back, the relay forwards it to the connected browser.
type ExecRelay struct {
	SessionID   string
	ContainerID string
	ProviderID  types.NodeID
	OnData      func(data []byte)   // Called when container output arrives
	OnClose     func(reason string) // Called when session ends
}

// RegisterExecRelay registers a requester-side exec relay for P2P → WebSocket forwarding
func (cm *ContainerManager) RegisterExecRelay(relay *ExecRelay) {
	cm.execRelaysMu.Lock()
	defer cm.execRelaysMu.Unlock()
	cm.execRelays[relay.SessionID] = relay
}

// RemoveExecRelay removes a requester-side exec relay
func (cm *ContainerManager) RemoveExecRelay(sessionID string) {
	cm.execRelaysMu.Lock()
	defer cm.execRelaysMu.Unlock()
	delete(cm.execRelays, sessionID)
}

// getExecRelay returns a relay by session ID
func (cm *ContainerManager) getExecRelay(sessionID string) (*ExecRelay, bool) {
	cm.execRelaysMu.RLock()
	defer cm.execRelaysMu.RUnlock()
	r, ok := cm.execRelays[sessionID]
	return r, ok
}

// SendExecMessage sends an exec P2P message to a target node
func (cm *ContainerManager) SendExecMessage(ctx context.Context, to types.NodeID, msg *types.Message) error {
	return cm.router.SendMessage(ctx, to, msg)
}

// LocalNodeID returns this node's ID
func (cm *ContainerManager) LocalNodeID() types.NodeID {
	return cm.node.nodeInfo.ID
}

// GetDeploymentOwnerWallet returns the wallet address that created a deployment.
// For exec authorization: only the deployer can exec into their container.
func (cm *ContainerManager) GetDeploymentOwnerWallet(containerID string) (string, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	d, exists := cm.deployments[containerID]
	if !exists {
		return "", false
	}
	if d.Owner != "" {
		return d.Owner, true
	}
	return "", false
}

// ExecLocal opens a direct PTY session on the local container runtime.
// Used when the API server and provider are the same node, bypassing P2P.
// walletAddress is verified against the deployment owner for authorization.
func (cm *ContainerManager) ExecLocal(ctx context.Context, containerID string, walletAddress string, cols, rows uint16) (*runtime.InteractiveSession, error) {
	if cm.containerd == nil {
		return nil, fmt.Errorf("container runtime not available")
	}
	cm.mu.RLock()
	deployment, exists := cm.deployments[containerID]
	cm.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("container not found")
	}
	if deployment.Status != types.ContainerStatusRunning {
		return nil, fmt.Errorf("container not running")
	}
	// Ownership verification: wallet must match deployment owner
	if deployment.Owner != "" && !strings.EqualFold(deployment.Owner, walletAddress) {
		return nil, fmt.Errorf("forbidden: wallet %s does not own container", walletAddress)
	}
	if !cm.containerd.CanExec(containerID) {
		return nil, fmt.Errorf("exec disabled for container")
	}
	return cm.containerd.ExecInteractive(ctx, containerID, cols, rows)
}

// GetContainerProviderNode returns the provider node ID for a container
func (cm *ContainerManager) GetContainerProviderNode(containerID string) (types.NodeID, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	d, exists := cm.deployments[containerID]
	if !exists {
		var zero types.NodeID
		return zero, false
	}
	if d.ReplicaSet != nil && len(d.ReplicaSet.Replicas) > 0 {
		// Return the primary replica's node
		if d.ReplicaSet.Replicas[0] != nil {
			return d.ReplicaSet.Replicas[0].NodeID, true
		}
	}
	// If no replica set, container runs locally
	return cm.node.nodeInfo.ID, true
}
