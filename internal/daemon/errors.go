package daemon

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrContainerdNotAvailable indicates containerd is not available
	ErrContainerdNotAvailable = errors.New("containerd not available")
)

// ErrDeploymentNotFound is returned when a deployment is not found
type ErrDeploymentNotFound struct {
	ContainerID string
}

func (e ErrDeploymentNotFound) Error() string {
	return fmt.Sprintf("deployment not found: %s", e.ContainerID)
}
