package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd/cio"
)

// ExecInContainer executes a command in a running container
// NOTE: This will fail if the container's security profile disables exec
func (cc *ContainerdClient) ExecInContainer(ctx context.Context, id string, cmd []string) ([]byte, error) {
	ctx = cc.WithNamespace(ctx)

	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("container not found: %s", id)
	}

	managed.mu.RLock()
	task := managed.Task
	securityEnforcer := managed.SecurityEnforcer
	managed.mu.RUnlock()

	// Check security policy before allowing exec
	if securityEnforcer != nil {
		if err := securityEnforcer.ValidateExecCommand(cmd); err != nil {
			return nil, err
		}
	}

	if task == nil {
		return nil, fmt.Errorf("container not running: %s", id)
	}

	// Create exec process
	spec, err := managed.Container.Spec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container spec: %w", err)
	}

	pspec := spec.Process
	pspec.Args = cmd

	execID := fmt.Sprintf("exec-%d", time.Now().UnixNano())

	process, err := task.Exec(ctx, execID, pspec, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}
	defer process.Delete(ctx)

	if err := process.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start exec: %w", err)
	}

	exitCh, err := process.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for exec: %w", err)
	}

	<-exitCh

	return nil, nil
}

// CanExec checks if exec is allowed for the container
func (cc *ContainerdClient) CanExec(id string) bool {
	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return false
	}

	managed.mu.RLock()
	defer managed.mu.RUnlock()

	if managed.SecurityEnforcer == nil {
		return true // No security enforcer means exec is allowed
	}
	return managed.SecurityEnforcer.CanExec()
}

// CanAttach checks if attach is allowed for the container
func (cc *ContainerdClient) CanAttach(id string) bool {
	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return false
	}

	managed.mu.RLock()
	defer managed.mu.RUnlock()

	if managed.SecurityEnforcer == nil {
		return true
	}
	return managed.SecurityEnforcer.CanAttach()
}

// CanShell checks if shell access is allowed for the container
func (cc *ContainerdClient) CanShell(id string) bool {
	cc.mu.RLock()
	managed, exists := cc.containers[id]
	cc.mu.RUnlock()

	if !exists {
		return false
	}

	managed.mu.RLock()
	defer managed.mu.RUnlock()

	if managed.SecurityEnforcer == nil {
		return true
	}
	return managed.SecurityEnforcer.CanShell()
}
