package api

import (
	"fmt"
	"sync"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/internal/logging"
)

// DaemonBridge provides a connection pool to the daemon for the HTTP API
type DaemonBridge struct {
	socketPath string
	pool       chan *client.DaemonClient
	poolSize   int
	mu         sync.RWMutex
}

// NewDaemonBridge creates a new daemon bridge
func NewDaemonBridge(socketPath string, poolSize int) *DaemonBridge {
	if poolSize <= 0 {
		poolSize = 10
	}

	return &DaemonBridge{
		socketPath: socketPath,
		pool:       make(chan *client.DaemonClient, poolSize),
		poolSize:   poolSize,
	}
}

// getClient gets a client from the pool or creates a new one
func (b *DaemonBridge) getClient() (*client.DaemonClient, error) {
	select {
	case c := <-b.pool:
		return c, nil
	default:
		// Pool empty, create new client
		c := client.NewDaemonClient(b.socketPath)
		if err := c.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect to daemon: %w", err)
		}
		return c, nil
	}
}

// putClient returns a client to the pool
func (b *DaemonBridge) putClient(c *client.DaemonClient) {
	select {
	case b.pool <- c:
		// Returned to pool
	default:
		// Pool full, close client
		c.Close()
	}
}

// withClient executes a function with a client from the pool
func (b *DaemonBridge) withClient(fn func(*client.DaemonClient) error) error {
	c, err := b.getClient()
	if err != nil {
		return err
	}

	err = fn(c)
	if err != nil {
		// On error, close and don't return to pool
		c.Close()
		return err
	}

	b.putClient(c)
	return nil
}

// Status forwards to daemon
func (b *DaemonBridge) Status() (*client.StatusResponse, error) {
	var result *client.StatusResponse
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.Status()
		return err
	})
	return result, err
}

// Deploy forwards to daemon
func (b *DaemonBridge) Deploy(req *client.DeployRequest) (*client.DeployResponse, error) {
	var result *client.DeployResponse
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.Deploy(req)
		return err
	})
	return result, err
}

// List forwards to daemon
func (b *DaemonBridge) List() ([]client.ContainerInfo, error) {
	var result []client.ContainerInfo
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.List()
		return err
	})
	return result, err
}

// Stop forwards to daemon
func (b *DaemonBridge) Stop(containerID string) error {
	return b.withClient(func(c *client.DaemonClient) error {
		return c.Stop(containerID)
	})
}

// Start forwards to daemon
func (b *DaemonBridge) Start(containerID string) error {
	return b.withClient(func(c *client.DaemonClient) error {
		return c.Start(containerID)
	})
}

// Delete forwards to daemon
func (b *DaemonBridge) Delete(containerID string) error {
	return b.withClient(func(c *client.DaemonClient) error {
		return c.Delete(containerID)
	})
}

// ThreatLevel forwards to daemon
func (b *DaemonBridge) ThreatLevel() (*client.ThreatLevelResponse, error) {
	var result *client.ThreatLevelResponse
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.ThreatLevel()
		return err
	})
	return result, err
}

// Clone forwards to daemon
func (b *DaemonBridge) Clone(req *client.CloneRequest) (*client.CloneResponse, error) {
	var result *client.CloneResponse
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.Clone(req)
		return err
	})
	return result, err
}

// CloneStatus forwards to daemon
func (b *DaemonBridge) CloneStatus(cloneID string) (*client.CloneResponse, error) {
	var result *client.CloneResponse
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.CloneStatus(cloneID)
		return err
	})
	return result, err
}

// CloneList forwards to daemon
func (b *DaemonBridge) CloneList(active bool, limit int) ([]client.CloneResponse, error) {
	var result []client.CloneResponse
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.CloneList(active, limit)
		return err
	})
	return result, err
}

// SnapshotCreate forwards to daemon
func (b *DaemonBridge) SnapshotCreate(req *client.SnapshotRequest) (*client.SnapshotResponse, error) {
	var result *client.SnapshotResponse
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.SnapshotCreate(req)
		return err
	})
	return result, err
}

// SnapshotGet forwards to daemon
func (b *DaemonBridge) SnapshotGet(snapshotID string) (*client.SnapshotResponse, error) {
	var result *client.SnapshotResponse
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.SnapshotGet(snapshotID)
		return err
	})
	return result, err
}

// SnapshotList forwards to daemon
func (b *DaemonBridge) SnapshotList(containerID string) ([]client.SnapshotResponse, error) {
	var result []client.SnapshotResponse
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.SnapshotList(containerID)
		return err
	})
	return result, err
}

// SnapshotDelete forwards to daemon
func (b *DaemonBridge) SnapshotDelete(snapshotID string) error {
	return b.withClient(func(c *client.DaemonClient) error {
		return c.SnapshotDelete(snapshotID)
	})
}

// SnapshotRestore forwards to daemon
func (b *DaemonBridge) SnapshotRestore(snapshotID, targetRegion string, newContainer bool) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.SnapshotRestore(snapshotID, targetRegion, newContainer)
		return err
	})
	return result, err
}

// RequesterBalance forwards to daemon
func (b *DaemonBridge) RequesterBalance() (map[string]interface{}, error) {
	var result map[string]interface{}
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.RequesterBalanceMap()
		return err
	})
	return result, err
}

// GetLogs retrieves container logs from the daemon
func (b *DaemonBridge) GetLogs(containerID string, follow bool, tail int) (string, error) {
	var result string
	err := b.withClient(func(c *client.DaemonClient) error {
		var err error
		result, err = c.GetLogs(containerID, follow, tail)
		return err
	})
	return result, err
}

// Close closes all connections in the pool
func (b *DaemonBridge) Close() {
	close(b.pool)
	for c := range b.pool {
		c.Close()
	}
	logging.Info("daemon bridge closed", logging.Component("api"))
}

// IsConnected checks if daemon is reachable
func (b *DaemonBridge) IsConnected() bool {
	c, err := b.getClient()
	if err != nil {
		return false
	}
	defer b.putClient(c)
	return c.IsDaemonRunning()
}
