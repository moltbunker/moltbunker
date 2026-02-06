package snapshot

import (
	"context"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

// CheckpointConfig configures automatic checkpointing
type CheckpointConfig struct {
	Enabled          bool `yaml:"enabled"`
	IntervalSeconds  int  `yaml:"interval_seconds"`   // Default: 300 (5 minutes)
	MaxCheckpoints   int  `yaml:"max_checkpoints"`    // Per container, default: 10
	RetainOnShutdown bool `yaml:"retain_on_shutdown"` // Keep checkpoints when container stops
}

// DefaultCheckpointConfig returns the default checkpoint configuration
func DefaultCheckpointConfig() *CheckpointConfig {
	return &CheckpointConfig{
		Enabled:          true,
		IntervalSeconds:  300,
		MaxCheckpoints:   10,
		RetainOnShutdown: true,
	}
}

// StateProvider is an interface for components that can provide state data
type StateProvider interface {
	// GetState returns the current state data for checkpointing
	GetState() ([]byte, error)
	// GetContainerID returns the container ID
	GetContainerID() string
}

// Checkpointer handles automatic periodic checkpoints
type Checkpointer struct {
	config   *CheckpointConfig
	manager  *Manager
	mu       sync.RWMutex
	running  bool
	stopCh   chan struct{}

	// Tracked containers
	containers map[string]StateProvider
	schedules  map[string]*checkpointSchedule
}

type checkpointSchedule struct {
	containerID    string
	provider       StateProvider
	lastCheckpoint time.Time
	nextCheckpoint time.Time
	ticker         *time.Ticker
	stopCh         chan struct{}
}

// NewCheckpointer creates a new checkpointer
func NewCheckpointer(manager *Manager, config *CheckpointConfig) *Checkpointer {
	if config == nil {
		config = DefaultCheckpointConfig()
	}

	return &Checkpointer{
		config:     config,
		manager:    manager,
		containers: make(map[string]StateProvider),
		schedules:  make(map[string]*checkpointSchedule),
		stopCh:     make(chan struct{}),
	}
}

// Start begins the checkpointing service
func (c *Checkpointer) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	if !c.config.Enabled {
		logging.Info("checkpointing disabled", logging.Component("checkpoint"))
		return nil
	}

	logging.Info("checkpointer started",
		"interval", c.config.IntervalSeconds,
		"max_checkpoints", c.config.MaxCheckpoints,
		logging.Component("checkpoint"))

	// Start cleanup routine
	util.SafeGoWithName("checkpoint-cleanup", func() {
		c.cleanupLoop(ctx)
	})

	return nil
}

// Stop halts the checkpointing service
func (c *Checkpointer) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}
	c.running = false
	close(c.stopCh)

	// Stop all schedules
	for _, sched := range c.schedules {
		if sched.ticker != nil {
			sched.ticker.Stop()
		}
		close(sched.stopCh)
	}
	c.schedules = make(map[string]*checkpointSchedule)

	logging.Info("checkpointer stopped", logging.Component("checkpoint"))
}

// RegisterContainer adds a container for automatic checkpointing
func (c *Checkpointer) RegisterContainer(provider StateProvider) error {
	if !c.config.Enabled {
		return nil
	}

	containerID := provider.GetContainerID()

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.containers[containerID]; exists {
		return nil // Already registered
	}

	c.containers[containerID] = provider

	// Create schedule
	interval := time.Duration(c.config.IntervalSeconds) * time.Second
	sched := &checkpointSchedule{
		containerID:    containerID,
		provider:       provider,
		lastCheckpoint: time.Time{},
		nextCheckpoint: time.Now().Add(interval),
		ticker:         time.NewTicker(interval),
		stopCh:         make(chan struct{}),
	}
	c.schedules[containerID] = sched

	// Start checkpoint goroutine
	util.SafeGoWithName("checkpoint-"+containerID, func() {
		c.checkpointLoop(sched)
	})

	logging.Info("container registered for checkpointing",
		"container_id", containerID,
		"interval", c.config.IntervalSeconds,
		logging.Component("checkpoint"))

	return nil
}

// UnregisterContainer removes a container from automatic checkpointing
func (c *Checkpointer) UnregisterContainer(containerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.containers, containerID)

	if sched, exists := c.schedules[containerID]; exists {
		if sched.ticker != nil {
			sched.ticker.Stop()
		}
		close(sched.stopCh)
		delete(c.schedules, containerID)
	}

	// Optionally delete checkpoints
	if !c.config.RetainOnShutdown {
		c.manager.DeleteContainerSnapshots(containerID)
	}

	logging.Info("container unregistered from checkpointing",
		"container_id", containerID,
		logging.Component("checkpoint"))
}

// checkpointLoop runs periodic checkpoints for a container
func (c *Checkpointer) checkpointLoop(sched *checkpointSchedule) {
	for {
		select {
		case <-sched.stopCh:
			return
		case <-c.stopCh:
			return
		case <-sched.ticker.C:
			c.createCheckpoint(sched)
		}
	}
}

// createCheckpoint creates a checkpoint for a scheduled container
func (c *Checkpointer) createCheckpoint(sched *checkpointSchedule) {
	c.mu.RLock()
	provider, exists := c.containers[sched.containerID]
	c.mu.RUnlock()

	if !exists {
		return
	}

	// Get state from provider
	data, err := provider.GetState()
	if err != nil {
		logging.Warn("failed to get container state for checkpoint",
			"container_id", sched.containerID,
			"error", err.Error(),
			logging.Component("checkpoint"))
		return
	}

	// Create checkpoint snapshot
	metadata := map[string]string{
		"checkpoint": "true",
		"automated":  "true",
	}

	snapshot, err := c.manager.CreateSnapshot(sched.containerID, data, SnapshotTypeCheckpoint, metadata)
	if err != nil {
		logging.Warn("failed to create checkpoint",
			"container_id", sched.containerID,
			"error", err.Error(),
			logging.Component("checkpoint"))
		return
	}

	sched.lastCheckpoint = time.Now()
	sched.nextCheckpoint = sched.lastCheckpoint.Add(time.Duration(c.config.IntervalSeconds) * time.Second)

	logging.Debug("checkpoint created",
		"container_id", sched.containerID,
		"snapshot_id", snapshot.ID,
		"size", snapshot.Size,
		logging.Component("checkpoint"))
}

// cleanupLoop periodically cleans up old checkpoints
func (c *Checkpointer) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.manager.CleanupExpired()
		}
	}
}

// TriggerCheckpoint manually triggers a checkpoint for a container
func (c *Checkpointer) TriggerCheckpoint(containerID string) (*Snapshot, error) {
	c.mu.RLock()
	provider, exists := c.containers[containerID]
	c.mu.RUnlock()

	if !exists {
		return nil, nil // Container not registered
	}

	// Get state
	data, err := provider.GetState()
	if err != nil {
		return nil, err
	}

	// Create manual checkpoint
	metadata := map[string]string{
		"checkpoint": "true",
		"automated":  "false",
		"manual":     "true",
	}

	return c.manager.CreateSnapshot(containerID, data, SnapshotTypeCheckpoint, metadata)
}

// GetLastCheckpoint returns the time of the last checkpoint for a container
func (c *Checkpointer) GetLastCheckpoint(containerID string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if sched, exists := c.schedules[containerID]; exists {
		return sched.lastCheckpoint, !sched.lastCheckpoint.IsZero()
	}
	return time.Time{}, false
}

// GetNextCheckpoint returns when the next checkpoint is scheduled
func (c *Checkpointer) GetNextCheckpoint(containerID string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if sched, exists := c.schedules[containerID]; exists {
		return sched.nextCheckpoint, true
	}
	return time.Time{}, false
}

// SetInterval changes the checkpoint interval for a container
func (c *Checkpointer) SetInterval(containerID string, seconds int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if sched, exists := c.schedules[containerID]; exists {
		if sched.ticker != nil {
			sched.ticker.Stop()
		}
		interval := time.Duration(seconds) * time.Second
		sched.ticker = time.NewTicker(interval)
		sched.nextCheckpoint = time.Now().Add(interval)

		logging.Info("checkpoint interval changed",
			"container_id", containerID,
			"interval", seconds,
			logging.Component("checkpoint"))
	}
}

// IsEnabled returns whether checkpointing is enabled
func (c *Checkpointer) IsEnabled() bool {
	return c.config.Enabled
}

// GetRegisteredContainers returns the list of containers with checkpointing
func (c *Checkpointer) GetRegisteredContainers() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]string, 0, len(c.containers))
	for id := range c.containers {
		ids = append(ids, id)
	}
	return ids
}
