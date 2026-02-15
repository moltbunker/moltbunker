package runtime

import (
	"context"
	"fmt"
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestNewSecurityEnforcer(t *testing.T) {
	t.Run("with nil profile uses defaults", func(t *testing.T) {
		se := NewSecurityEnforcer(nil)
		if se == nil {
			t.Fatal("expected security enforcer, got nil")
		}
		if se.profile == nil {
			t.Fatal("expected profile, got nil")
		}
		// Default profile should disable exec
		if se.CanExec() {
			t.Error("default profile should disable exec")
		}
	})

	t.Run("with custom profile", func(t *testing.T) {
		profile := &types.ContainerSecurityProfile{
			DisableExec:   false,
			DisableAttach: true,
			DisableShell:  true,
		}
		se := NewSecurityEnforcer(profile)
		if !se.CanExec() {
			t.Error("exec should be enabled")
		}
		if se.CanAttach() {
			t.Error("attach should be disabled")
		}
		if se.CanShell() {
			t.Error("shell should be disabled")
		}
	})
}

func TestSecurityEnforcerCanExec(t *testing.T) {
	tests := []struct {
		name     string
		disabled bool
		expected bool
	}{
		{"exec enabled", false, true},
		{"exec disabled", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			se := NewSecurityEnforcer(&types.ContainerSecurityProfile{
				DisableExec: tt.disabled,
			})
			if se.CanExec() != tt.expected {
				t.Errorf("CanExec() = %v, want %v", se.CanExec(), tt.expected)
			}
		})
	}
}

func TestSecurityEnforcerCanAttach(t *testing.T) {
	tests := []struct {
		name     string
		disabled bool
		expected bool
	}{
		{"attach enabled", false, true},
		{"attach disabled", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			se := NewSecurityEnforcer(&types.ContainerSecurityProfile{
				DisableAttach: tt.disabled,
			})
			if se.CanAttach() != tt.expected {
				t.Errorf("CanAttach() = %v, want %v", se.CanAttach(), tt.expected)
			}
		})
	}
}

func TestSecurityEnforcerCanShell(t *testing.T) {
	tests := []struct {
		name     string
		disabled bool
		expected bool
	}{
		{"shell enabled", false, true},
		{"shell disabled", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			se := NewSecurityEnforcer(&types.ContainerSecurityProfile{
				DisableShell: tt.disabled,
			})
			if se.CanShell() != tt.expected {
				t.Errorf("CanShell() = %v, want %v", se.CanShell(), tt.expected)
			}
		})
	}
}

func TestValidateExecCommand(t *testing.T) {
	tests := []struct {
		name         string
		execDisabled bool
		shellDisabled bool
		cmd          []string
		expectError  bool
		errorType    error
	}{
		{
			name:        "exec disabled blocks all commands",
			execDisabled: true,
			cmd:         []string{"ls", "-la"},
			expectError: true,
			errorType:   ErrExecDisabled,
		},
		{
			name:         "exec enabled allows non-shell commands",
			execDisabled: false,
			shellDisabled: true,
			cmd:          []string{"ls", "-la"},
			expectError:  false,
		},
		{
			name:         "shell disabled blocks /bin/sh",
			execDisabled: false,
			shellDisabled: true,
			cmd:          []string{"/bin/sh"},
			expectError:  true,
			errorType:    ErrShellDisabled,
		},
		{
			name:         "shell disabled blocks /bin/bash",
			execDisabled: false,
			shellDisabled: true,
			cmd:          []string{"/bin/bash"},
			expectError:  true,
			errorType:    ErrShellDisabled,
		},
		{
			name:         "shell disabled blocks sh",
			execDisabled: false,
			shellDisabled: true,
			cmd:          []string{"sh"},
			expectError:  true,
			errorType:    ErrShellDisabled,
		},
		{
			name:         "shell disabled blocks bash",
			execDisabled: false,
			shellDisabled: true,
			cmd:          []string{"bash"},
			expectError:  true,
			errorType:    ErrShellDisabled,
		},
		{
			name:         "shell disabled blocks zsh",
			execDisabled: false,
			shellDisabled: true,
			cmd:          []string{"zsh"},
			expectError:  true,
			errorType:    ErrShellDisabled,
		},
		{
			name:         "shell enabled allows shells",
			execDisabled: false,
			shellDisabled: false,
			cmd:          []string{"/bin/bash"},
			expectError:  false,
		},
		{
			name:        "empty command returns error",
			execDisabled: false,
			cmd:         []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			se := NewSecurityEnforcer(&types.ContainerSecurityProfile{
				DisableExec:  tt.execDisabled,
				DisableShell: tt.shellDisabled,
			})
			err := se.ValidateExecCommand(tt.cmd)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBuildOCISpecOpts_DeploymentProfile(t *testing.T) {
	profile := types.DeploymentSecurityProfile()
	se := NewSecurityEnforcer(profile)
	opts := se.BuildOCISpecOpts()

	if len(opts) == 0 {
		t.Error("deployment profile should generate OCI spec options")
	}

	// Deployment profile enables: capabilities, NoNewPrivileges, masked paths,
	// read-only paths, seccomp, ulimits — at minimum 6 option functions
	if len(opts) < 5 {
		t.Errorf("expected at least 5 OCI spec options for deployment profile, got %d", len(opts))
	}
}

func TestBuildOCISpecOpts(t *testing.T) {
	t.Run("default profile generates options", func(t *testing.T) {
		profile := types.DefaultContainerSecurityProfile()
		se := NewSecurityEnforcer(profile)
		opts := se.BuildOCISpecOpts()

		// Default profile has many security features enabled
		// so we expect multiple OCI spec options
		if len(opts) == 0 {
			t.Error("expected OCI spec options, got none")
		}
	})

	t.Run("minimal profile generates fewer options", func(t *testing.T) {
		profile := &types.ContainerSecurityProfile{
			// Minimal profile with no security features
		}
		se := NewSecurityEnforcer(profile)
		opts := se.BuildOCISpecOpts()

		// Minimal profile should still have ulimits (default)
		// but not as many options as the default profile
		defaultSe := NewSecurityEnforcer(types.DefaultContainerSecurityProfile())
		defaultOpts := defaultSe.BuildOCISpecOpts()

		if len(opts) >= len(defaultOpts) {
			t.Errorf("minimal profile should have fewer options than default: minimal=%d, default=%d",
				len(opts), len(defaultOpts))
		}
	})
}

func TestIsSecurityError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ErrExecDisabled is security error",
			err:      ErrExecDisabled,
			expected: true,
		},
		{
			name:     "ErrAttachDisabled is security error",
			err:      ErrAttachDisabled,
			expected: true,
		},
		{
			name:     "ErrShellDisabled is security error",
			err:      ErrShellDisabled,
			expected: true,
		},
		{
			name:     "regular error is not security error",
			err:      fmt.Errorf("some error"),
			expected: false,
		},
		{
			name:     "nil is not security error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsSecurityError(tt.err) != tt.expected {
				t.Errorf("IsSecurityError(%v) = %v, want %v", tt.err, IsSecurityError(tt.err), tt.expected)
			}
		})
	}
}

func TestWithSeccompProfile_StrictEmptyAllowlist(t *testing.T) {
	// This tests the bug fix: "strict" mode with empty allowedSyscalls
	// must still set DefaultAction to ActErrno (not ActAllow).
	opt := WithSeccompProfile("strict", []string{"ptrace"}, []string{})

	// Create a minimal spec and apply the option
	s := &specs.Spec{
		Linux: &specs.Linux{},
	}
	if err := opt(context.Background(), nil, nil, s); err != nil {
		t.Fatalf("WithSeccompProfile failed: %v", err)
	}

	if s.Linux.Seccomp == nil {
		t.Fatal("seccomp profile should be set")
	}

	if s.Linux.Seccomp.DefaultAction != specs.ActErrno {
		t.Errorf("strict mode should always set DefaultAction to ActErrno, got %v", s.Linux.Seccomp.DefaultAction)
	}

	// With empty allowlist, should fall back to essential syscalls
	hasAllowRules := false
	for _, rule := range s.Linux.Seccomp.Syscalls {
		if rule.Action == specs.ActAllow {
			hasAllowRules = true
			break
		}
	}
	if !hasAllowRules {
		t.Error("strict mode with empty allowlist should fall back to essential syscalls")
	}
}

func TestWithSeccompProfile_StrictWithAllowlist(t *testing.T) {
	allowed := []string{"read", "write", "open"}
	opt := WithSeccompProfile("strict", []string{"ptrace"}, allowed)

	s := &specs.Spec{
		Linux: &specs.Linux{},
	}
	if err := opt(context.Background(), nil, nil, s); err != nil {
		t.Fatalf("WithSeccompProfile failed: %v", err)
	}

	if s.Linux.Seccomp.DefaultAction != specs.ActErrno {
		t.Errorf("strict mode should set DefaultAction to ActErrno, got %v", s.Linux.Seccomp.DefaultAction)
	}

	// Should use the provided allowlist, not the fallback
	allowCount := 0
	for _, rule := range s.Linux.Seccomp.Syscalls {
		if rule.Action == specs.ActAllow {
			allowCount += len(rule.Names)
		}
	}
	if allowCount != len(allowed) {
		t.Errorf("expected %d allowed syscalls, got %d", len(allowed), allowCount)
	}
}

func TestWithSeccompProfile_DefaultMode(t *testing.T) {
	blocked := []string{"ptrace", "mount"}
	opt := WithSeccompProfile("default", blocked, []string{})

	s := &specs.Spec{
		Linux: &specs.Linux{},
	}
	if err := opt(context.Background(), nil, nil, s); err != nil {
		t.Fatalf("WithSeccompProfile failed: %v", err)
	}

	if s.Linux.Seccomp.DefaultAction != specs.ActAllow {
		t.Errorf("default mode should set DefaultAction to ActAllow, got %v", s.Linux.Seccomp.DefaultAction)
	}

	// All blocked syscalls should have ActErrno rules
	errnoCount := 0
	for _, rule := range s.Linux.Seccomp.Syscalls {
		if rule.Action == specs.ActErrno {
			errnoCount += len(rule.Names)
		}
	}
	if errnoCount != len(blocked) {
		t.Errorf("expected %d blocked syscalls, got %d", len(blocked), errnoCount)
	}
}

func TestSecurityProfileError(t *testing.T) {
	err := &SecurityProfileError{
		Operation: "exec",
		Reason:    "disabled by policy",
	}

	expected := "security policy violation: exec - disabled by policy"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
