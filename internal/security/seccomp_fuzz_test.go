package security

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzSeccompLoadCorrupt fuzzes the seccomp profile loader with arbitrary data.
// It verifies that LoadSeccompProfile never panics on malformed input.
func FuzzSeccompLoadCorrupt(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"defaultAction":"SCMP_ACT_ERRNO","architectures":["SCMP_ARCH_X86_64"],"syscalls":[{"names":["read"],"action":"SCMP_ACT_ALLOW"}]}`))
	f.Add([]byte(`not json`))
	f.Add([]byte{0xff, 0xfe, 0x00})
	f.Add([]byte(`{"defaultAction":"","syscalls":[]}`))
	f.Add([]byte(`{"defaultAction":"SCMP_ACT_ALLOW","architectures":["x"],"syscalls":[{"names":["ptrace"],"action":"SCMP_ACT_ALLOW"}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "seccomp.json")

		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Skip("cannot write test file")
		}

		// Must not panic — errors are acceptable
		profile, err := LoadSeccompProfile(path)
		if err != nil {
			return
		}

		// If it loaded, validate must not panic either
		_ = ValidateSeccompProfile(profile)

		// AllowedSyscalls/DeniedSyscalls must not panic
		_ = profile.AllowedSyscalls()
		_ = profile.DeniedSyscalls()
	})
}

// FuzzSeccompValidate fuzzes the ValidateSeccompProfile function with
// constructed profiles using arbitrary syscall names and actions.
func FuzzSeccompValidate(f *testing.F) {
	f.Add("SCMP_ACT_ERRNO", "SCMP_ARCH_X86_64", "read", "SCMP_ACT_ALLOW")
	f.Add("SCMP_ACT_KILL", "SCMP_ARCH_AARCH64", "ptrace", "SCMP_ACT_KILL")
	f.Add("SCMP_ACT_ALLOW", "", "", "")
	f.Add("", "x", "mount", "SCMP_ACT_ALLOW")

	f.Fuzz(func(t *testing.T, defaultAction, arch, syscallName, ruleAction string) {
		var archs []string
		if arch != "" {
			archs = []string{arch}
		}

		var rules []SyscallRule
		if syscallName != "" {
			rules = []SyscallRule{
				{Names: []string{syscallName}, Action: ruleAction},
			}
		}

		profile := &SeccompProfile{
			DefaultAction: defaultAction,
			Architectures: archs,
			Syscalls:      rules,
		}

		// Must not panic
		_ = ValidateSeccompProfile(profile)
		_ = profile.AllowedSyscalls()
		_ = profile.DeniedSyscalls()
	})
}
