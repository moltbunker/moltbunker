package runtime

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/user"
	goruntime "runtime"
	"strings"

	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// SecurityEnforcer enforces container security policies
type SecurityEnforcer struct {
	profile *types.ContainerSecurityProfile
}

// NewSecurityEnforcer creates a new security enforcer with the given profile
func NewSecurityEnforcer(profile *types.ContainerSecurityProfile) *SecurityEnforcer {
	if profile == nil {
		profile = types.DefaultContainerSecurityProfile()
	}
	return &SecurityEnforcer{profile: profile}
}

// GetProfile returns the security profile
func (se *SecurityEnforcer) GetProfile() *types.ContainerSecurityProfile {
	return se.profile
}

// CanExec returns true if exec is allowed
func (se *SecurityEnforcer) CanExec() bool {
	return !se.profile.DisableExec
}

// CanAttach returns true if attach is allowed
func (se *SecurityEnforcer) CanAttach() bool {
	return !se.profile.DisableAttach
}

// CanShell returns true if shell is allowed
func (se *SecurityEnforcer) CanShell() bool {
	return !se.profile.DisableShell
}

// ValidateExecCommand checks if a command is allowed to be executed
// Returns an error if the command is blocked (e.g., shell commands when shells are disabled)
func (se *SecurityEnforcer) ValidateExecCommand(cmd []string) error {
	if !se.CanExec() {
		return ErrExecDisabled
	}

	if len(cmd) == 0 {
		return fmt.Errorf("empty command")
	}

	// Check for shell commands if shell is disabled
	if !se.CanShell() {
		shellCommands := []string{
			"sh", "bash", "zsh", "ash", "dash", "csh", "tcsh", "ksh", "fish",
			"/bin/sh", "/bin/bash", "/bin/zsh", "/bin/ash", "/bin/dash",
			"/bin/csh", "/bin/tcsh", "/bin/ksh", "/bin/fish",
			"/usr/bin/sh", "/usr/bin/bash", "/usr/bin/zsh", "/usr/bin/fish",
		}
		cmdLower := strings.ToLower(cmd[0])
		for _, shell := range shellCommands {
			if cmdLower == shell {
				return ErrShellDisabled
			}
		}
	}

	return nil
}

// BuildOCISpecOpts converts the security profile into OCI spec options
func (se *SecurityEnforcer) BuildOCISpecOpts() []oci.SpecOpts {
	var opts []oci.SpecOpts

	// Capability handling
	if se.profile.DropAllCapabilities {
		opts = append(opts, oci.WithCapabilities(se.profile.AddCapabilities))
	}

	// Filesystem restrictions
	if se.profile.ReadOnlyRoot {
		opts = append(opts, oci.WithRootFSReadonly())
	}

	// No new privileges
	if se.profile.NoNewPrivileges {
		opts = append(opts, WithNoNewPrivileges())
	}

	// Masked paths
	if len(se.profile.MaskPaths) > 0 {
		opts = append(opts, WithMaskedPaths(se.profile.MaskPaths))
	}

	// Read-only paths
	if len(se.profile.ReadOnlyPaths) > 0 {
		opts = append(opts, WithReadonlyPaths(se.profile.ReadOnlyPaths))
	}

	// Seccomp profile
	if se.profile.SeccompProfile != "" && se.profile.SeccompProfile != "unconfined" {
		opts = append(opts, WithSeccompProfile(se.profile.SeccompProfile, se.profile.BlockedSyscalls, se.profile.AllowedSyscalls))
	}

	// AppArmor profile — only apply if running on Linux and profile is loaded.
	// R9: when a profile is requested but NOT present in the kernel, log a Debug
	// line so an operator can see the confinement gap (and run `moltbunker doctor`
	// to load it) instead of the gate silently no-op'ing.
	if se.profile.AppArmorProfile != "" {
		if isAppArmorProfileLoaded(se.profile.AppArmorProfile) {
			opts = append(opts, WithAppArmorProfile(se.profile.AppArmorProfile))
		} else {
			logging.Debug("AppArmor profile not loaded; container will run without AppArmor confinement — run `moltbunker node doctor` to load it",
				"profile", se.profile.AppArmorProfile,
				logging.Component("apparmor"))
		}
	}

	// SELinux label
	if se.profile.SELinuxLabel != "" {
		opts = append(opts, WithSELinuxLabel(se.profile.SELinuxLabel))
	}

	// Ulimits
	opts = append(opts, WithUlimits(se.profile.Ulimits))

	// Namespace isolation
	if se.profile.UserNamespace {
		opts = append(opts, WithUserNamespace())
	}

	return opts
}

// WithNoNewPrivileges sets the no_new_privs flag
func WithNoNewPrivileges() oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		if s.Process == nil {
			s.Process = &specs.Process{}
		}
		s.Process.NoNewPrivileges = true
		return nil
	}
}

// WithMaskedPaths adds paths to mask in the container
func WithMaskedPaths(paths []string) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		if s.Linux == nil {
			s.Linux = &specs.Linux{}
		}
		s.Linux.MaskedPaths = append(s.Linux.MaskedPaths, paths...)
		return nil
	}
}

// WithReadonlyPaths adds read-only paths in the container
func WithReadonlyPaths(paths []string) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		if s.Linux == nil {
			s.Linux = &specs.Linux{}
		}
		s.Linux.ReadonlyPaths = append(s.Linux.ReadonlyPaths, paths...)
		return nil
	}
}

// WithSeccompProfile sets the seccomp profile
func WithSeccompProfile(profile string, blockedSyscalls, allowedSyscalls []string) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		if s.Linux == nil {
			s.Linux = &specs.Linux{}
		}

		// Build seccomp configuration based on profile type
		var seccomp *specs.LinuxSeccomp

		switch profile {
		case "strict":
			// Strict mode: deny by default, only allow specific syscalls
			seccomp = &specs.LinuxSeccomp{
				DefaultAction: specs.ActErrno,
			}
			// Use provided allowlist, or fall back to essential syscalls
			allowed := allowedSyscalls
			if len(allowed) == 0 {
				allowed = security.GetEssentialSyscalls()
			}
			for _, syscall := range allowed {
				seccomp.Syscalls = append(seccomp.Syscalls, specs.LinuxSyscall{
					Names:  []string{syscall},
					Action: specs.ActAllow,
				})
			}
			// Always block dangerous syscalls (overrides any allow)
			for _, syscall := range blockedSyscalls {
				seccomp.Syscalls = append(seccomp.Syscalls, specs.LinuxSyscall{
					Names:  []string{syscall},
					Action: specs.ActErrno,
				})
			}
		case "default":
			// Default mode: allow by default, block specific syscalls
			seccomp = &specs.LinuxSeccomp{
				DefaultAction: specs.ActAllow,
			}
			for _, syscall := range blockedSyscalls {
				seccomp.Syscalls = append(seccomp.Syscalls, specs.LinuxSyscall{
					Names:  []string{syscall},
					Action: specs.ActErrno,
				})
			}
		default:
			return nil
		}

		s.Linux.Seccomp = seccomp
		return nil
	}
}

// WithAppArmorProfile sets the AppArmor profile
func WithAppArmorProfile(profile string) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		if s.Process == nil {
			s.Process = &specs.Process{}
		}
		s.Process.ApparmorProfile = profile
		return nil
	}
}

// WithSELinuxLabel sets the SELinux label
func WithSELinuxLabel(label string) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		if s.Process == nil {
			s.Process = &specs.Process{}
		}
		s.Process.SelinuxLabel = label
		return nil
	}
}

// WithUlimits sets resource limits
func WithUlimits(ulimits types.UlimitConfig) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		if s.Process == nil {
			s.Process = &specs.Process{}
		}

		// Set ulimits
		if ulimits.NoFile > 0 {
			s.Process.Rlimits = append(s.Process.Rlimits, specs.POSIXRlimit{
				Type: "RLIMIT_NOFILE",
				Hard: uint64(ulimits.NoFile),
				Soft: uint64(ulimits.NoFile),
			})
		}

		if ulimits.NProc > 0 {
			s.Process.Rlimits = append(s.Process.Rlimits, specs.POSIXRlimit{
				Type: "RLIMIT_NPROC",
				Hard: uint64(ulimits.NProc),
				Soft: uint64(ulimits.NProc),
			})
		}

		// MemLock of 0 means no locked memory
		s.Process.Rlimits = append(s.Process.Rlimits, specs.POSIXRlimit{
			Type: "RLIMIT_MEMLOCK",
			Hard: uint64(ulimits.MemLock), // #nosec G115 -- ulimit config value, non-negative by contract (0 = none)
			Soft: uint64(ulimits.MemLock), // #nosec G115 -- ulimit config value, non-negative by contract (0 = none)
		})

		// Core of 0 means no core dumps
		s.Process.Rlimits = append(s.Process.Rlimits, specs.POSIXRlimit{
			Type: "RLIMIT_CORE",
			Hard: uint64(ulimits.Core), // #nosec G115 -- ulimit config value, non-negative by contract (0 = no core dumps)
			Soft: uint64(ulimits.Core), // #nosec G115 -- ulimit config value, non-negative by contract (0 = no core dumps)
		})

		if ulimits.Stack > 0 {
			s.Process.Rlimits = append(s.Process.Rlimits, specs.POSIXRlimit{
				Type: "RLIMIT_STACK",
				Hard: uint64(ulimits.Stack),
				Soft: uint64(ulimits.Stack),
			})
		}

		return nil
	}
}

// WithUserNamespace enables user-namespace isolation for tenant workloads.
//
// OPT-IN ONLY: DeploymentSecurityProfile leaves UserNamespace=false by default
// (see pkg/types/security.go). This option emits the OCI UID/GID mappings but the
// rootfs snapshot is NOT remapped — containers are created with a plain overlay
// snapshot owned by host UID 0. With a userns active, container-UID-0 maps to host
// UID <subStart>, so the host-UID-0 rootfs appears as nobody/overflow inside the
// namespace and most real images fail to start. Enabling userns for real therefore
// requires a REMAPPED snapshot — containerd's WithRemapperLabels
// (containerd.io/snapshot/uidmapping + gidmapping) or an idmapped-mount-capable
// snapshotter that ID-shifts the rootfs into the host subordinate range to match
// these mappings. That snapshot-remap wiring is the R11-gated follow-up (needs real
// Linux container-startup CI). Do NOT flip the profile default to ON until it lands.
//
// R12: This is gated by CheckUserNSCompat. The previous behavior mapped only a
// single ID (container root -> host nobody 65534), which is too narrow to host a
// real container (anything that runs as a non-zero UID inside the container had no
// host UID to land on). When the host supports unprivileged user namespaces and the
// daemon's user has a subordinate ID range, this emits the standard 65536-entry
// mapping (container 0..65535 -> host <subStart>..<subStart+65535>), matching what
// Docker/containerd's rootless support uses.
//
// On a host that disables unprivileged user namespaces (e.g. sysctl
// kernel.unprivileged_userns_clone=0, or no /etc/subuid range), it degrades
// gracefully: it logs a structured Warning and returns without touching the spec,
// so the container still starts (without userns) rather than failing the deploy.
// On non-Linux platforms CheckUserNSCompat always reports unsupported, so this is
// a no-op there.
//
// NOTE(R16): when userns is active, XFS disk-quota project IDs are assigned to the
// host UID range (hostStart), not to container UID 0 — quotactl/xfs_quota reports
// bytes against the host subordinate user, so reports must be correlated back to
// the container by project ID, not by the in-container UID.
func WithUserNamespace() oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		compat := CheckUserNSCompat()
		if !compat.Supported {
			logging.Warn("user namespace requested but not enabled on this host; container will run in the host user namespace",
				"reason", compat.Reason,
				logging.Component("userns"))
			return nil
		}

		uname := "root"
		if u, err := user.Current(); err == nil && u.Username != "" {
			uname = u.Username
		}
		hostStart, size, err := ResolveSubUIDRange(uname)
		if err != nil {
			logging.Warn("user namespace requested but no usable subordinate ID range; running without userns",
				"reason", err.Error(),
				logging.Component("userns"))
			return nil
		}

		if s.Linux == nil {
			s.Linux = &specs.Linux{}
		}

		s.Linux.Namespaces = append(s.Linux.Namespaces, specs.LinuxNamespace{
			Type: specs.UserNamespace,
		})

		// Map the full container UID/GID space (0..65535) onto the host
		// subordinate range so processes running as any in-container UID land on
		// an unprivileged host UID.
		s.Linux.UIDMappings = []specs.LinuxIDMapping{
			{ContainerID: 0, HostID: hostStart, Size: size},
		}
		s.Linux.GIDMappings = []specs.LinuxIDMapping{
			{ContainerID: 0, HostID: hostStart, Size: size},
		}

		return nil
	}
}

// SecurityProfileError represents a security policy violation
type SecurityProfileError struct {
	Operation string
	Reason    string
}

func (e *SecurityProfileError) Error() string {
	return fmt.Sprintf("security policy violation: %s - %s", e.Operation, e.Reason)
}

// Common security errors
var (
	ErrExecDisabled = &SecurityProfileError{
		Operation: "exec",
		Reason:    "container exec is disabled by security policy",
	}
	ErrAttachDisabled = &SecurityProfileError{
		Operation: "attach",
		Reason:    "container attach is disabled by security policy",
	}
	ErrShellDisabled = &SecurityProfileError{
		Operation: "shell",
		Reason:    "shell access is disabled by security policy",
	}
)

// IsSecurityError checks if an error is a security policy error
func IsSecurityError(err error) bool {
	_, ok := err.(*SecurityProfileError)
	return ok
}

// isAppArmorProfileLoaded checks if an AppArmor profile is loaded in the kernel.
// Returns false on non-Linux platforms or if the profiles file cannot be read.
func isAppArmorProfileLoaded(profile string) bool {
	if goruntime.GOOS != "linux" {
		return false
	}
	f, err := os.Open("/sys/kernel/security/apparmor/profiles")
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Lines look like: "profile_name (enforce)" or "profile_name (complain)"
		line := scanner.Text()
		if strings.HasPrefix(line, profile+" ") || line == profile {
			return true
		}
	}
	return false
}
