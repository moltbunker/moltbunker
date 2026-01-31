package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// CgroupManager manages cgroup v2 resource limits
type CgroupManager struct {
	cgroupRoot string
}

// NewCgroupManager creates a new cgroup manager
func NewCgroupManager(cgroupRoot string) *CgroupManager {
	if cgroupRoot == "" {
		cgroupRoot = "/sys/fs/cgroup"
	}

	return &CgroupManager{
		cgroupRoot: cgroupRoot,
	}
}

// CreateCgroup creates a cgroup for a container
func (cm *CgroupManager) CreateCgroup(containerID string) error {
	cgroupPath := filepath.Join(cm.cgroupRoot, "moltbunker", containerID)

	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return fmt.Errorf("failed to create cgroup: %w", err)
	}

	return nil
}

// SetCPULimit sets CPU quota and period
func (cm *CgroupManager) SetCPULimit(containerID string, quota int64, period uint64) error {
	cgroupPath := filepath.Join(cm.cgroupRoot, "moltbunker", containerID)

	// Set CPU quota
	quotaPath := filepath.Join(cgroupPath, "cpu.max")
	quotaStr := fmt.Sprintf("%d %d", quota, period)
	if err := os.WriteFile(quotaPath, []byte(quotaStr), 0644); err != nil {
		return fmt.Errorf("failed to set CPU quota: %w", err)
	}

	return nil
}

// SetMemoryLimit sets memory limit
func (cm *CgroupManager) SetMemoryLimit(containerID string, limit int64) error {
	cgroupPath := filepath.Join(cm.cgroupRoot, "moltbunker", containerID)

	// Set memory limit
	memoryPath := filepath.Join(cgroupPath, "memory.max")
	limitStr := strconv.FormatInt(limit, 10)
	if err := os.WriteFile(memoryPath, []byte(limitStr), 0644); err != nil {
		return fmt.Errorf("failed to set memory limit: %w", err)
	}

	return nil
}

// SetPIDLimit sets PID limit
func (cm *CgroupManager) SetPIDLimit(containerID string, limit int) error {
	cgroupPath := filepath.Join(cm.cgroupRoot, "moltbunker", containerID)

	// Set PID limit
	pidPath := filepath.Join(cgroupPath, "pids.max")
	limitStr := strconv.Itoa(limit)
	if err := os.WriteFile(pidPath, []byte(limitStr), 0644); err != nil {
		return fmt.Errorf("failed to set PID limit: %w", err)
	}

	return nil
}

// ApplyResourceLimits applies all resource limits to a cgroup
func (cm *CgroupManager) ApplyResourceLimits(containerID string, resources types.ResourceLimits) error {
	if err := cm.CreateCgroup(containerID); err != nil {
		return err
	}

	if err := cm.SetCPULimit(containerID, resources.CPUQuota, resources.CPUPeriod); err != nil {
		return err
	}

	if err := cm.SetMemoryLimit(containerID, resources.MemoryLimit); err != nil {
		return err
	}

	if err := cm.SetPIDLimit(containerID, resources.PIDLimit); err != nil {
		return err
	}

	return nil
}

// DeleteCgroup deletes a cgroup
func (cm *CgroupManager) DeleteCgroup(containerID string) error {
	cgroupPath := filepath.Join(cm.cgroupRoot, "moltbunker", containerID)

	if err := os.RemoveAll(cgroupPath); err != nil {
		return fmt.Errorf("failed to delete cgroup: %w", err)
	}

	return nil
}
