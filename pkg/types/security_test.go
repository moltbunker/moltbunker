package types

import (
	"testing"
)

func TestDeploymentSecurityProfile(t *testing.T) {
	profile := DeploymentSecurityProfile()
	if profile == nil {
		t.Fatal("DeploymentSecurityProfile() returned nil")
	}

	// Interactive access should be enabled
	if profile.DisableExec {
		t.Error("exec should be enabled for deployment profile")
	}
	if profile.DisableAttach {
		t.Error("attach should be enabled for deployment profile")
	}
	if profile.DisableShell {
		t.Error("shell should be enabled for deployment profile")
	}
}

func TestDeploymentSecurityProfile_Capabilities(t *testing.T) {
	profile := DeploymentSecurityProfile()

	if !profile.DropAllCapabilities {
		t.Error("should drop all capabilities")
	}

	// Docker default safe set
	expectedCaps := map[string]bool{
		"CAP_CHOWN":            true,
		"CAP_DAC_OVERRIDE":     true,
		"CAP_FSETID":           true,
		"CAP_FOWNER":           true,
		"CAP_MKNOD":            true,
		"CAP_NET_RAW":          true,
		"CAP_SETGID":           true,
		"CAP_SETUID":           true,
		"CAP_SETFCAP":          true,
		"CAP_SETPCAP":          true,
		"CAP_NET_BIND_SERVICE": true,
		"CAP_SYS_CHROOT":      true,
		"CAP_KILL":             true,
		"CAP_AUDIT_WRITE":      true,
	}

	if len(profile.AddCapabilities) != len(expectedCaps) {
		t.Errorf("expected %d capabilities, got %d", len(expectedCaps), len(profile.AddCapabilities))
	}

	for _, cap := range profile.AddCapabilities {
		if !expectedCaps[cap] {
			t.Errorf("unexpected capability: %s", cap)
		}
	}

	// Dangerous capabilities must NOT be in the list
	dangerousCaps := []string{
		"CAP_SYS_ADMIN",
		"CAP_SYS_PTRACE",
		"CAP_SYS_MODULE",
		"CAP_SYS_RAWIO",
		"CAP_SYS_BOOT",
		"CAP_SYS_TIME",
		"CAP_NET_ADMIN",
		"CAP_LINUX_IMMUTABLE",
	}
	capSet := make(map[string]bool)
	for _, c := range profile.AddCapabilities {
		capSet[c] = true
	}
	for _, cap := range dangerousCaps {
		if capSet[cap] {
			t.Errorf("dangerous capability should not be in deployment profile: %s", cap)
		}
	}
}

func TestDeploymentSecurityProfile_Namespaces(t *testing.T) {
	profile := DeploymentSecurityProfile()

	if profile.UserNamespace {
		t.Error("user namespace should be disabled (breaks most images)")
	}
	if !profile.PIDNamespace {
		t.Error("PID namespace should be enabled")
	}
	if !profile.NetworkNamespace {
		t.Error("network namespace should be enabled")
	}
	if !profile.IPCNamespace {
		t.Error("IPC namespace should be enabled")
	}
}

func TestDeploymentSecurityProfile_Filesystem(t *testing.T) {
	profile := DeploymentSecurityProfile()

	if profile.ReadOnlyRoot {
		t.Error("read-only root should be disabled (many images need writable root)")
	}
	if !profile.NoNewPrivileges {
		t.Error("NoNewPrivileges should be enabled")
	}

	// Check masked paths
	maskedPathSet := make(map[string]bool)
	for _, p := range profile.MaskPaths {
		maskedPathSet[p] = true
	}
	requiredMasked := []string{
		"/proc/kcore", "/proc/kmem", "/proc/mem", "/proc/keys",
		"/sys/firmware", "/sys/fs/selinux",
	}
	for _, p := range requiredMasked {
		if !maskedPathSet[p] {
			t.Errorf("expected masked path: %s", p)
		}
	}

	// Check read-only paths
	roPathSet := make(map[string]bool)
	for _, p := range profile.ReadOnlyPaths {
		roPathSet[p] = true
	}
	requiredRO := []string{"/proc/bus", "/proc/fs", "/proc/irq", "/proc/sys", "/proc/sysrq-trigger"}
	for _, p := range requiredRO {
		if !roPathSet[p] {
			t.Errorf("expected read-only path: %s", p)
		}
	}
}

func TestDeploymentSecurityProfile_Seccomp(t *testing.T) {
	profile := DeploymentSecurityProfile()

	if profile.SeccompProfile != "default" {
		t.Errorf("expected seccomp profile 'default', got %q", profile.SeccompProfile)
	}

	if len(profile.BlockedSyscalls) == 0 {
		t.Error("blocked syscalls should not be empty")
	}

	// Verify critical dangerous syscalls are blocked
	blockedSet := make(map[string]bool)
	for _, s := range profile.BlockedSyscalls {
		blockedSet[s] = true
	}
	mustBlock := []string{
		"ptrace", "mount", "umount2", "reboot", "kexec_load",
		"bpf", "setns", "unshare", "clone3", "move_mount",
		"fsopen", "fsmount",
	}
	for _, s := range mustBlock {
		if !blockedSet[s] {
			t.Errorf("dangerous syscall %q should be in blocked list", s)
		}
	}
}

func TestDeploymentSecurityProfile_Ulimits(t *testing.T) {
	profile := DeploymentSecurityProfile()

	if profile.Ulimits.NoFile != 65536 {
		t.Errorf("expected NoFile=65536, got %d", profile.Ulimits.NoFile)
	}
	if profile.Ulimits.NProc != 4096 {
		t.Errorf("expected NProc=4096, got %d", profile.Ulimits.NProc)
	}
	if profile.Ulimits.MemLock != 0 {
		t.Errorf("expected MemLock=0, got %d", profile.Ulimits.MemLock)
	}
	if profile.Ulimits.Core != 0 {
		t.Errorf("expected Core=0, got %d", profile.Ulimits.Core)
	}
	if profile.Ulimits.Stack != 8388608 {
		t.Errorf("expected Stack=8388608, got %d", profile.Ulimits.Stack)
	}
}

func TestDeploymentSecurityProfile_NoAppArmor(t *testing.T) {
	profile := DeploymentSecurityProfile()

	if profile.AppArmorProfile != "" {
		t.Errorf("deployment profile should have empty AppArmor profile, got %q", profile.AppArmorProfile)
	}
}

func TestDeploymentSecurityProfile_DiffersFromDefault(t *testing.T) {
	deploy := DeploymentSecurityProfile()
	def := DefaultContainerSecurityProfile()

	// Deployment profile should be more permissive in key areas
	if deploy.DisableExec == def.DisableExec {
		t.Error("deployment profile should differ from default on exec")
	}
	if deploy.ReadOnlyRoot == def.ReadOnlyRoot {
		t.Error("deployment profile should differ from default on read-only root")
	}
	if deploy.UserNamespace == def.UserNamespace {
		t.Error("deployment profile should differ from default on user namespace")
	}
	if deploy.SeccompProfile == def.SeccompProfile {
		t.Error("deployment profile should use different seccomp mode than default")
	}
}
