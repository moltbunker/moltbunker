package daemon

import (
	"context"
	"fmt"
)

// StartTor starts the Tor service
func (cm *ContainerManager) StartTor(ctx context.Context) error {
	if cm.torService == nil {
		return fmt.Errorf("Tor service not configured")
	}
	return cm.torService.Start(ctx)
}

// StopTor stops the Tor service
func (cm *ContainerManager) StopTor() error {
	if cm.torService == nil {
		return nil
	}
	return cm.torService.Stop()
}

// GetTorStatus returns Tor service status
func (cm *ContainerManager) GetTorStatus() (bool, string) {
	if cm.torService == nil {
		return false, ""
	}
	return cm.torService.IsRunning(), cm.torService.GetOnionAddress()
}

// RotateTorCircuit rotates the Tor circuit
func (cm *ContainerManager) RotateTorCircuit(ctx context.Context) error {
	if cm.torService == nil {
		return fmt.Errorf("Tor service not configured")
	}
	return cm.torService.RotateCircuit(ctx)
}
