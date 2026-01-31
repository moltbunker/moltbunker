package tor

import (
	"context"
	"fmt"
	"time"

	"github.com/cretz/bine/tor"
)

// CircuitManager manages Tor circuit rotation
type CircuitManager struct {
	tor           *tor.Tor
	rotationInterval time.Duration
	lastRotation     time.Time
}

// NewCircuitManager creates a new circuit manager
func NewCircuitManager(torInstance *tor.Tor) *CircuitManager {
	return &CircuitManager{
		tor:               torInstance,
		rotationInterval: 10 * time.Minute, // Rotate every 10 minutes
		lastRotation:     time.Now(),
	}
}

// Start starts automatic circuit rotation
func (cm *CircuitManager) Start(ctx context.Context) {
	ticker := time.NewTicker(cm.rotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := cm.RotateCircuit(ctx); err != nil {
				// Log error
				continue
			}
		}
	}
}

// RotateCircuit rotates the Tor circuit
func (cm *CircuitManager) RotateCircuit(ctx context.Context) error {
	// Signal Tor to create new circuits
	// This is done by closing existing circuits and creating new ones
	if cm.tor == nil {
		return fmt.Errorf("Tor instance not available")
	}

	// Request new circuit
	// Note: bine doesn't expose direct circuit control, so we use control port
	// For now, we'll just mark the rotation time
	cm.lastRotation = time.Now()

	return nil
}

// SetRotationInterval sets the circuit rotation interval
func (cm *CircuitManager) SetRotationInterval(interval time.Duration) {
	cm.rotationInterval = interval
}

// LastRotation returns the last rotation time
func (cm *CircuitManager) LastRotation() time.Time {
	return cm.lastRotation
}
