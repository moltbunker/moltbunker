package runtime

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/containerd/containerd/cio"
)

// InteractiveSession represents an active PTY exec session in a container
type InteractiveSession struct {
	SessionID string
	Stdin     io.WriteCloser
	Stdout    io.ReadCloser
	done      chan struct{}
	closeOnce sync.Once
	cancel    context.CancelFunc

	// resizeFn is called to resize the PTY
	resizeFn func(cols, rows uint16) error
}

// Resize changes the terminal dimensions
func (s *InteractiveSession) Resize(cols, rows uint16) error {
	if s.resizeFn != nil {
		return s.resizeFn(cols, rows)
	}
	return nil
}

// Done returns a channel that's closed when the session ends
func (s *InteractiveSession) Done() <-chan struct{} {
	return s.done
}

// Close terminates the exec session and cleans up resources
func (s *InteractiveSession) Close() {
	s.closeOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		if s.Stdin != nil {
			s.Stdin.Close()
		}
		close(s.done)
	})
}

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

	// Wait for exit or context cancellation. Without the select, a cancelled
	// context leaves this goroutine blocked forever on the channel read.
	select {
	case <-exitCh:
		// Normal exit
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return nil, nil
}

// ExecInteractive opens an interactive PTY session in a container.
// Returns an InteractiveSession with bidirectional I/O and resize support.
func (cc *ContainerdClient) ExecInteractive(ctx context.Context, id string, cols, rows uint16) (*InteractiveSession, error) {
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

	// Check security policy
	if securityEnforcer != nil {
		if err := securityEnforcer.ValidateExecCommand([]string{"/bin/sh"}); err != nil {
			return nil, err
		}
	}

	if task == nil {
		return nil, fmt.Errorf("container not running: %s", id)
	}

	// Get container spec for process defaults
	spec, err := managed.Container.Spec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container spec: %w", err)
	}

	pspec := spec.Process
	pspec.Args = []string{"/bin/sh"}
	pspec.Terminal = true

	// Create pipes for bidirectional I/O
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	execID := fmt.Sprintf("exec-pty-%d", time.Now().UnixNano())
	execCtx, execCancel := context.WithCancel(ctx)

	// Terminal mode: only stdin+stdout FIFOs (PTY merges stderr into stdout).
	// cio.WithTerminal prevents creating a stderr FIFO that the shim never opens,
	// which would block the entire I/O setup in copyIO's wg.Wait().
	process, err := task.Exec(execCtx, execID, pspec,
		cio.NewCreator(cio.WithTerminal, cio.WithStreams(stdinR, stdoutW, nil)))
	if err != nil {
		execCancel()
		stdinR.Close()
		stdinW.Close()
		stdoutR.Close()
		stdoutW.Close()
		return nil, fmt.Errorf("failed to create interactive exec: %w", err)
	}

	if err := process.Start(execCtx); err != nil {
		process.Delete(execCtx)
		execCancel()
		stdinR.Close()
		stdinW.Close()
		stdoutR.Close()
		stdoutW.Close()
		return nil, fmt.Errorf("failed to start interactive exec: %w", err)
	}

	session := &InteractiveSession{
		SessionID: execID,
		Stdin:     stdinW,
		Stdout:    stdoutR,
		done:      make(chan struct{}),
		cancel:    execCancel,
		resizeFn: func(c, r uint16) error {
			return process.Resize(execCtx, uint32(c), uint32(r))
		},
	}

	// Wait for process exit in background and clean up
	go func() {
		defer func() {
			stdoutW.Close()
			stdinR.Close()
			process.Delete(context.Background())
			session.Close()
		}()

		exitCh, err := process.Wait(execCtx)
		if err != nil {
			return
		}
		<-exitCh
	}()

	return session, nil
}

// ExecRaw opens a raw (non-PTY) exec session in a container.
// Unlike ExecInteractive, this uses plain stdin/stdout pipes without terminal processing.
// Designed for running the exec-agent binary which manages its own PTY internally.
func (cc *ContainerdClient) ExecRaw(ctx context.Context, id string, cmd []string) (*InteractiveSession, error) {
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

	if securityEnforcer != nil {
		if err := securityEnforcer.ValidateExecCommand(cmd); err != nil {
			return nil, err
		}
	}

	if task == nil {
		return nil, fmt.Errorf("container not running: %s", id)
	}

	spec, err := managed.Container.Spec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container spec: %w", err)
	}

	pspec := spec.Process
	pspec.Args = cmd
	pspec.Terminal = false // raw pipes, no PTY

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	execID := fmt.Sprintf("exec-raw-%d", time.Now().UnixNano())
	execCtx, execCancel := context.WithCancel(ctx)

	// Non-terminal mode: plain stdin+stdout+stderr pipes.
	// Stderr is merged into stdout (stdoutW) so exec-agent output is unified.
	process, err := task.Exec(execCtx, execID, pspec,
		cio.NewCreator(cio.WithStreams(stdinR, stdoutW, stdoutW)))
	if err != nil {
		execCancel()
		stdinR.Close()
		stdinW.Close()
		stdoutR.Close()
		stdoutW.Close()
		return nil, fmt.Errorf("failed to create raw exec: %w", err)
	}

	if err := process.Start(execCtx); err != nil {
		process.Delete(execCtx)
		execCancel()
		stdinR.Close()
		stdinW.Close()
		stdoutR.Close()
		stdoutW.Close()
		return nil, fmt.Errorf("failed to start raw exec: %w", err)
	}

	session := &InteractiveSession{
		SessionID: execID,
		Stdin:     stdinW,
		Stdout:    stdoutR,
		done:      make(chan struct{}),
		cancel:    execCancel,
		// resizeFn left nil — exec-agent handles resize internally via frames
	}

	go func() {
		defer func() {
			stdoutW.Close()
			stdinR.Close()
			process.Delete(context.Background())
			session.Close()
		}()

		exitCh, err := process.Wait(execCtx)
		if err != nil {
			return
		}
		<-exitCh
	}()

	return session, nil
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
