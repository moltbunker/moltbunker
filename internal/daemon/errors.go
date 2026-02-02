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

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// DeploymentError represents a deployment-related error
type DeploymentError struct {
	ContainerID string
	Reason      string
	Err         error
}

func (e DeploymentError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("deployment error for %s: %s: %v", e.ContainerID, e.Reason, e.Err)
	}
	return fmt.Sprintf("deployment error for %s: %s", e.ContainerID, e.Reason)
}

func (e DeploymentError) Unwrap() error {
	return e.Err
}

// ReplicationError represents a replication-related error
type ReplicationError struct {
	ContainerID string
	Region      string
	Reason      string
	Err         error
}

func (e ReplicationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("replication error for %s in %s: %s: %v", e.ContainerID, e.Region, e.Reason, e.Err)
	}
	return fmt.Sprintf("replication error for %s in %s: %s", e.ContainerID, e.Region, e.Reason)
}

func (e ReplicationError) Unwrap() error {
	return e.Err
}

// TorError represents a Tor-related error
type TorError struct {
	Operation string
	Err       error
}

func (e TorError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("tor error during %s: %v", e.Operation, e.Err)
	}
	return fmt.Sprintf("tor error during %s", e.Operation)
}

func (e TorError) Unwrap() error {
	return e.Err
}
