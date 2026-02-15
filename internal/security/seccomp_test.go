package security

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestDefaultSeccompProfile(t *testing.T) {
	profile := DefaultSeccompProfile()

	if profile.DefaultAction != "SCMP_ACT_ERRNO" {
		t.Errorf("Default action should be SCMP_ACT_ERRNO, got %s", profile.DefaultAction)
	}

	if len(profile.Syscalls) == 0 {
		t.Error("Should have syscall rules")
	}

	if len(profile.Architectures) == 0 {
		t.Error("Should have architectures")
	}
}

func TestDefaultSeccompProfile_DenyByDefault(t *testing.T) {
	profile := DefaultSeccompProfile()

	// The default action must deny syscalls not in the allow list
	validDeny := map[string]bool{
		"SCMP_ACT_ERRNO": true,
		"SCMP_ACT_KILL":  true,
	}

	if !validDeny[profile.DefaultAction] {
		t.Errorf("Default action must be deny-by-default (SCMP_ACT_ERRNO or SCMP_ACT_KILL), got %s", profile.DefaultAction)
	}
}

func TestDefaultSeccompProfile_Architectures(t *testing.T) {
	profile := DefaultSeccompProfile()

	archSet := make(map[string]bool)
	for _, arch := range profile.Architectures {
		archSet[arch] = true
	}

	required := []string{"SCMP_ARCH_X86_64", "SCMP_ARCH_AARCH64"}
	for _, arch := range required {
		if !archSet[arch] {
			t.Errorf("Profile should include architecture %s", arch)
		}
	}
}

func TestDefaultSeccompProfile_EssentialSyscalls(t *testing.T) {
	profile := DefaultSeccompProfile()

	allowed := make(map[string]bool)
	for _, rule := range profile.Syscalls {
		if rule.Action == "SCMP_ACT_ALLOW" {
			for _, name := range rule.Names {
				allowed[name] = true
			}
		}
	}

	// These essential syscalls must be present for any container to function
	essential := []string{
		"read", "write", "open", "close",
		"stat", "fstat", "lstat",
		"mmap", "mprotect", "munmap", "brk",
		"openat", "getdents64",
		"socket", "connect", "bind", "listen",
		"accept", "accept4",
		"clone", "execve", "exit", "exit_group",
		"wait4", "getpid",
		"pipe", "pipe2", "dup", "dup2",
		"fcntl", "ioctl",
		"getcwd", "chdir",
		"select", "poll", "epoll_create1", "epoll_ctl", "epoll_wait",
		"futex", "nanosleep",
		"getrandom",
		"rt_sigaction", "rt_sigprocmask", "rt_sigreturn",
	}

	for _, name := range essential {
		if !allowed[name] {
			t.Errorf("Essential syscall %q is not in the allow list", name)
		}
	}
}

func TestDefaultSeccompProfile_DangerousSyscallsBlocked(t *testing.T) {
	profile := DefaultSeccompProfile()

	allowed := make(map[string]bool)
	for _, rule := range profile.Syscalls {
		if rule.Action == "SCMP_ACT_ALLOW" {
			for _, name := range rule.Names {
				allowed[name] = true
			}
		}
	}

	// These dangerous syscalls must NOT be in the allow list
	dangerous := []string{
		"ptrace",           // process tracing
		"mount",            // mount filesystem
		"umount2",          // unmount filesystem
		"reboot",           // reboot system
		"swapon",           // enable swap
		"swapoff",          // disable swap
		"sethostname",      // change hostname
		"setdomainname",    // change domain name
		"iopl",             // I/O privilege level
		"ioperm",           // I/O port permissions
		"create_module",    // load kernel modules
		"init_module",      // load kernel modules
		"delete_module",    // unload kernel modules
		"finit_module",     // load kernel modules from fd
		"kexec_load",       // load new kernel
		"kexec_file_load",  // load new kernel from file
		"bpf",              // BPF programs
		"pivot_root",       // change root filesystem
		"process_vm_readv", // read another process memory
		"process_vm_writev", // write another process memory
		"settimeofday",     // set system time
		"clock_settime",    // set system clock
		"acct",             // process accounting
		"syslog",           // kernel log
		"userfaultfd",      // userfault fd
		"perf_event_open",  // performance monitoring
	}

	for _, name := range dangerous {
		if allowed[name] {
			t.Errorf("Dangerous syscall %q should not be in the allow list", name)
		}
	}
}

func TestDefaultSeccompProfile_DangerousSyscallsKilled(t *testing.T) {
	profile := DefaultSeccompProfile()

	killed := make(map[string]bool)
	for _, rule := range profile.Syscalls {
		if rule.Action == "SCMP_ACT_KILL" {
			for _, name := range rule.Names {
				killed[name] = true
			}
		}
	}

	// Dangerous syscalls should be explicitly killed (not just denied)
	mustKill := []string{
		"ptrace", "mount", "reboot", "kexec_load",
		"bpf", "process_vm_readv", "process_vm_writev",
	}

	for _, name := range mustKill {
		if !killed[name] {
			t.Errorf("Dangerous syscall %q should be explicitly killed (SCMP_ACT_KILL)", name)
		}
	}
}

func TestDefaultSeccompProfile_HasTwoRules(t *testing.T) {
	profile := DefaultSeccompProfile()

	if len(profile.Syscalls) < 2 {
		t.Fatalf("Profile should have at least 2 rules (allow + kill), got %d", len(profile.Syscalls))
	}

	hasAllow := false
	hasKill := false
	for _, rule := range profile.Syscalls {
		if rule.Action == "SCMP_ACT_ALLOW" {
			hasAllow = true
		}
		if rule.Action == "SCMP_ACT_KILL" {
			hasKill = true
		}
	}

	if !hasAllow {
		t.Error("Profile must have an SCMP_ACT_ALLOW rule")
	}
	if !hasKill {
		t.Error("Profile must have an SCMP_ACT_KILL rule for dangerous syscalls")
	}
}

func TestDefaultSeccompProfile_NoDuplicateSyscalls(t *testing.T) {
	profile := DefaultSeccompProfile()

	seen := make(map[string]string) // syscall -> action
	for _, rule := range profile.Syscalls {
		for _, name := range rule.Names {
			if prevAction, exists := seen[name]; exists {
				t.Errorf("Syscall %q appears in both %s and %s rules", name, prevAction, rule.Action)
			}
			seen[name] = rule.Action
		}
	}
}

func TestSaveLoadSeccompProfile(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "seccomp.json")

	profile := DefaultSeccompProfile()

	err := SaveSeccompProfile(profile, profilePath)
	if err != nil {
		t.Fatalf("Failed to save seccomp profile: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		t.Error("Seccomp profile file should exist")
	}

	// Load profile
	loadedProfile, err := LoadSeccompProfile(profilePath)
	if err != nil {
		t.Fatalf("Failed to load seccomp profile: %v", err)
	}

	if loadedProfile.DefaultAction != profile.DefaultAction {
		t.Errorf("Default action mismatch: got %s, want %s", loadedProfile.DefaultAction, profile.DefaultAction)
	}

	if len(loadedProfile.Syscalls) != len(profile.Syscalls) {
		t.Errorf("Syscall rule count mismatch: got %d, want %d", len(loadedProfile.Syscalls), len(profile.Syscalls))
	}

	if len(loadedProfile.Architectures) != len(profile.Architectures) {
		t.Errorf("Architecture count mismatch: got %d, want %d", len(loadedProfile.Architectures), len(profile.Architectures))
	}

	// Verify syscall names survived roundtrip
	origAllowed := profile.AllowedSyscalls()
	loadedAllowed := loadedProfile.AllowedSyscalls()
	if len(origAllowed) != len(loadedAllowed) {
		t.Errorf("Allowed syscall count mismatch: got %d, want %d", len(loadedAllowed), len(origAllowed))
	}
	for i := range origAllowed {
		if origAllowed[i] != loadedAllowed[i] {
			t.Errorf("Allowed syscall mismatch at index %d: got %s, want %s", i, loadedAllowed[i], origAllowed[i])
			break
		}
	}
}

func TestLoadSeccompProfile_NotExists(t *testing.T) {
	_, err := LoadSeccompProfile("/nonexistent/seccomp.json")
	if err == nil {
		t.Error("Should fail for nonexistent file")
	}
}

func TestLoadSeccompProfile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.json")

	if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadSeccompProfile(path)
	if err == nil {
		t.Error("Should fail for invalid JSON")
	}
}

func TestValidateSeccompProfile_Valid(t *testing.T) {
	profile := DefaultSeccompProfile()
	if err := ValidateSeccompProfile(profile); err != nil {
		t.Errorf("Default profile should be valid: %v", err)
	}
}

func TestValidateSeccompProfile_Nil(t *testing.T) {
	if err := ValidateSeccompProfile(nil); err == nil {
		t.Error("Nil profile should fail validation")
	}
}

func TestValidateSeccompProfile_AllowDefault(t *testing.T) {
	profile := &SeccompProfile{
		DefaultAction: "SCMP_ACT_ALLOW",
		Architectures: []string{"SCMP_ARCH_X86_64"},
		Syscalls: []SyscallRule{
			{Names: []string{"read"}, Action: "SCMP_ACT_ALLOW"},
		},
	}

	if err := ValidateSeccompProfile(profile); err == nil {
		t.Error("Profile with SCMP_ACT_ALLOW default should fail validation")
	}
}

func TestValidateSeccompProfile_NoArchitectures(t *testing.T) {
	profile := &SeccompProfile{
		DefaultAction: "SCMP_ACT_ERRNO",
		Architectures: []string{},
		Syscalls: []SyscallRule{
			{Names: []string{"read"}, Action: "SCMP_ACT_ALLOW"},
		},
	}

	if err := ValidateSeccompProfile(profile); err == nil {
		t.Error("Profile with no architectures should fail validation")
	}
}

func TestValidateSeccompProfile_NoSyscalls(t *testing.T) {
	profile := &SeccompProfile{
		DefaultAction: "SCMP_ACT_ERRNO",
		Architectures: []string{"SCMP_ARCH_X86_64"},
		Syscalls:      []SyscallRule{},
	}

	if err := ValidateSeccompProfile(profile); err == nil {
		t.Error("Profile with no syscalls should fail validation")
	}
}

func TestValidateSeccompProfile_AllowsDangerousSyscall(t *testing.T) {
	profile := &SeccompProfile{
		DefaultAction: "SCMP_ACT_ERRNO",
		Architectures: []string{"SCMP_ARCH_X86_64"},
		Syscalls: []SyscallRule{
			{
				Names:  []string{"read", "write", "ptrace"},
				Action: "SCMP_ACT_ALLOW",
			},
		},
	}

	err := ValidateSeccompProfile(profile)
	if err == nil {
		t.Error("Profile allowing ptrace should fail validation")
	}
}

func TestSeccompProfile_AllowedSyscalls(t *testing.T) {
	profile := DefaultSeccompProfile()
	allowed := profile.AllowedSyscalls()

	if len(allowed) == 0 {
		t.Error("AllowedSyscalls should return non-empty list")
	}

	// Verify sorted
	if !sort.StringsAreSorted(allowed) {
		t.Error("AllowedSyscalls should return sorted list")
	}

	// Verify contains essential ones
	allowedSet := make(map[string]bool)
	for _, s := range allowed {
		allowedSet[s] = true
	}
	if !allowedSet["read"] || !allowedSet["write"] {
		t.Error("AllowedSyscalls should include read and write")
	}
}

func TestSeccompProfile_DeniedSyscalls(t *testing.T) {
	profile := DefaultSeccompProfile()
	denied := profile.DeniedSyscalls()

	if len(denied) == 0 {
		t.Error("DeniedSyscalls should return non-empty list")
	}

	// Verify sorted
	if !sort.StringsAreSorted(denied) {
		t.Error("DeniedSyscalls should return sorted list")
	}

	// Verify contains dangerous ones
	deniedSet := make(map[string]bool)
	for _, s := range denied {
		deniedSet[s] = true
	}
	if !deniedSet["ptrace"] || !deniedSet["reboot"] {
		t.Error("DeniedSyscalls should include ptrace and reboot")
	}
}

func TestEssentialSyscalls_NoDangerousSyscalls(t *testing.T) {
	essentials := GetEssentialSyscalls()
	essentialSet := make(map[string]bool, len(essentials))
	for _, s := range essentials {
		essentialSet[s] = true
	}

	// These were previously in essentials but are dangerous
	mustNotBeEssential := []string{
		"setns",      // namespace escape
		"chroot",     // root escape
		"adjtimex",   // time manipulation
		"modify_ldt", // kernel attack surface
	}

	for _, s := range mustNotBeEssential {
		if essentialSet[s] {
			t.Errorf("dangerous syscall %q should not be in essential syscalls list", s)
		}
	}
}

func TestDangerousSyscalls_ContainsNewEntries(t *testing.T) {
	dangerous := GetDangerousSyscalls()
	dangerousSet := make(map[string]bool, len(dangerous))
	for _, s := range dangerous {
		dangerousSet[s] = true
	}

	// New dangerous syscalls added for mount/filesystem escape prevention
	mustBeDangerous := []string{
		"clone3",
		"move_mount",
		"open_tree",
		"fsopen",
		"fspick",
		"fsconfig",
		"fsmount",
		"setns",
		"chroot",
		"adjtimex",
		"modify_ldt",
	}

	for _, s := range mustBeDangerous {
		if !dangerousSet[s] {
			t.Errorf("syscall %q should be in dangerous syscalls list", s)
		}
	}
}

func TestGetEssentialSyscalls_ReturnsCopy(t *testing.T) {
	a := GetEssentialSyscalls()
	b := GetEssentialSyscalls()

	if len(a) == 0 {
		t.Fatal("essential syscalls should not be empty")
	}

	// Mutate a and verify b is unaffected
	a[0] = "MUTATED"
	if b[0] == "MUTATED" {
		t.Error("GetEssentialSyscalls should return a copy, not a reference")
	}
}

func TestGetDangerousSyscalls_ReturnsCopy(t *testing.T) {
	a := GetDangerousSyscalls()
	b := GetDangerousSyscalls()

	if len(a) == 0 {
		t.Fatal("dangerous syscalls should not be empty")
	}

	a[0] = "MUTATED"
	if b[0] == "MUTATED" {
		t.Error("GetDangerousSyscalls should return a copy, not a reference")
	}
}

func TestNoOverlapBetweenEssentialAndDangerous(t *testing.T) {
	essentials := GetEssentialSyscalls()
	dangerous := GetDangerousSyscalls()

	dangerousSet := make(map[string]bool, len(dangerous))
	for _, s := range dangerous {
		dangerousSet[s] = true
	}

	for _, s := range essentials {
		if dangerousSet[s] {
			t.Errorf("syscall %q appears in both essential and dangerous lists", s)
		}
	}
}
