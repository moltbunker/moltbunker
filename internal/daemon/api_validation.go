package daemon

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// Security constants
const (
	// MaxRequestSize is the maximum allowed request size (1MB)
	MaxRequestSize = 1 * 1024 * 1024

	// Resource limits bounds
	maxMemoryLimit = 64 * 1024 * 1024 * 1024   // 64GB max
	maxDiskLimit   = 1024 * 1024 * 1024 * 1024 // 1TB max
	maxCPUQuota    = 10000000                  // 10 seconds in microseconds
	maxPIDLimit    = 10000
	maxNetworkBW   = 10 * 1024 * 1024 * 1024 // 10GB/s max

	// MaxLogTailLines is the maximum number of lines that can be requested via Tail parameter
	MaxLogTailLines = 10000
)

// Validation errors
var (
	ErrInvalidImageName     = errors.New("invalid container image name")
	ErrInvalidContainerID   = errors.New("invalid container ID format")
	ErrInvalidResourceLimit = errors.New("invalid resource limit")
	ErrInvalidTailValue     = errors.New("invalid tail value")
	ErrRequestTooLarge      = errors.New("request exceeds maximum size")
)

// Regex patterns for validation
var (
	// containerImageRegex allows alphanumeric, colons, slashes, dots, hyphens, and underscores
	// Examples: nginx:latest, docker.io/library/nginx:1.21, my-registry.com/my-image:v1.0.0
	containerImageRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._\-/:]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)

	// containerIDRegex allows hex characters (SHA256 prefix) or alphanumeric with hyphens
	containerIDRegex = regexp.MustCompile(`^[a-f0-9]{8,64}$|^[a-zA-Z0-9][a-zA-Z0-9\-]{0,62}[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)
)

// validateContainerImage validates a container image name
func validateContainerImage(image string) error {
	if image == "" {
		return fmt.Errorf("%w: image name cannot be empty", ErrInvalidImageName)
	}
	if len(image) > 256 {
		return fmt.Errorf("%w: image name exceeds maximum length of 256 characters", ErrInvalidImageName)
	}
	if !containerImageRegex.MatchString(image) {
		return fmt.Errorf("%w: image name contains invalid characters (allowed: alphanumeric, colons, slashes, dots, hyphens, underscores)", ErrInvalidImageName)
	}
	return nil
}

// validateContainerID validates a container ID format
func validateContainerID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: container ID cannot be empty", ErrInvalidContainerID)
	}
	if len(id) > 64 {
		return fmt.Errorf("%w: container ID exceeds maximum length of 64 characters", ErrInvalidContainerID)
	}
	if !containerIDRegex.MatchString(id) {
		return fmt.Errorf("%w: container ID must be hex characters or alphanumeric with hyphens", ErrInvalidContainerID)
	}
	return nil
}

// validateResourceLimits validates resource limits are positive and within reasonable bounds
func validateResourceLimits(r types.ResourceLimits) error {
	// CPU quota validation (if set, must be positive and within bounds)
	if r.CPUQuota < 0 {
		return fmt.Errorf("%w: CPU quota cannot be negative", ErrInvalidResourceLimit)
	}
	if r.CPUQuota > maxCPUQuota {
		return fmt.Errorf("%w: CPU quota exceeds maximum of %d", ErrInvalidResourceLimit, maxCPUQuota)
	}

	// Memory limit validation
	if r.MemoryLimit < 0 {
		return fmt.Errorf("%w: memory limit cannot be negative", ErrInvalidResourceLimit)
	}
	if r.MemoryLimit > maxMemoryLimit {
		return fmt.Errorf("%w: memory limit exceeds maximum of %d bytes", ErrInvalidResourceLimit, maxMemoryLimit)
	}

	// Disk limit validation
	if r.DiskLimit < 0 {
		return fmt.Errorf("%w: disk limit cannot be negative", ErrInvalidResourceLimit)
	}
	if r.DiskLimit > maxDiskLimit {
		return fmt.Errorf("%w: disk limit exceeds maximum of %d bytes", ErrInvalidResourceLimit, maxDiskLimit)
	}

	// Network bandwidth validation
	if r.NetworkBW < 0 {
		return fmt.Errorf("%w: network bandwidth cannot be negative", ErrInvalidResourceLimit)
	}
	if r.NetworkBW > maxNetworkBW {
		return fmt.Errorf("%w: network bandwidth exceeds maximum of %d bytes/sec", ErrInvalidResourceLimit, maxNetworkBW)
	}

	// PID limit validation
	if r.PIDLimit < 0 {
		return fmt.Errorf("%w: PID limit cannot be negative", ErrInvalidResourceLimit)
	}
	if r.PIDLimit > maxPIDLimit {
		return fmt.Errorf("%w: PID limit exceeds maximum of %d", ErrInvalidResourceLimit, maxPIDLimit)
	}

	return nil
}

// validateDeployRequest validates a deployment request
func validateDeployRequest(req *DeployRequest) error {
	// Validate image name
	if err := validateContainerImage(req.Image); err != nil {
		return err
	}

	// Validate resource limits
	if err := validateResourceLimits(req.Resources); err != nil {
		return err
	}

	// Validate onion port if specified
	if req.OnionPort < 0 || req.OnionPort > 65535 {
		return fmt.Errorf("invalid onion port: must be between 0 and 65535")
	}

	return nil
}
