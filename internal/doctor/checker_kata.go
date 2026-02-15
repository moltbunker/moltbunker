//go:build linux

package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// KataRuntimeChecker checks Kata Containers availability and provider tier classification.
//
// Checks performed:
//  1. KVM device (/dev/kvm) — hardware virtualization support
//  2. containerd-shim-kata-v2 binary — Kata runtime shim
//  3. QEMU hypervisor (qemu-system-x86_64) — Kata backend
//  4. Kata configuration file — runtime config
//  5. AMD SEV-SNP — confidential computing (Tier 1 vs Tier 2)
//  6. Provider tier classification — overall assessment
type KataRuntimeChecker struct{}

func NewKataRuntimeChecker() *KataRuntimeChecker {
	return &KataRuntimeChecker{}
}

func (c *KataRuntimeChecker) Name() string       { return "Kata Containers" }
func (c *KataRuntimeChecker) Category() Category { return CategoryRuntime }
func (c *KataRuntimeChecker) CanFix() bool       { return false }

func (c *KataRuntimeChecker) Check(ctx context.Context) CheckResult {
	result := CheckResult{
		Name:     c.Name(),
		Category: c.Category(),
		Fixable:  false,
	}

	// Step 1: Check KVM support (hardware requirement for Kata)
	kvmInfo, err := os.Stat("/dev/kvm")
	if err != nil {
		result.Status = StatusError
		result.Message = "Kata Containers: /dev/kvm not available"
		result.Details = "KVM is required for Kata VM isolation. " +
			"Ensure CPU has VT-x/AMD-V and kvm modules are loaded: " +
			"modprobe kvm && modprobe kvm_amd (or kvm_intel)"
		result.FixCommand = "modprobe kvm && modprobe kvm_amd"
		return result
	}

	// Check KVM is accessible (not just present)
	if kvmInfo.Mode()&0060 == 0 {
		// No group read/write — might not be accessible
		result.Status = StatusWarning
		result.Message = "Kata Containers: /dev/kvm exists but may not be accessible"
		result.Details = "Ensure your user is in the 'kvm' group: usermod -aG kvm $USER"
		result.FixCommand = "usermod -aG kvm $USER"
	}

	// Step 2: Check containerd-shim-kata-v2
	kataAvailable := runtime.IsKataAvailable()
	if !kataAvailable {
		result.Status = StatusWarning
		result.Message = "Kata Containers: Not installed (running as runc-only provider)"
		result.Details = buildKataInstallGuide()
		result.FixCommand = "apt install kata-containers"
		return result
	}

	// Step 3: Check QEMU (Kata's hypervisor backend)
	qemuPath, qemuErr := exec.LookPath("qemu-system-x86_64")
	qemuNote := ""
	if qemuErr != nil {
		// Also check aarch64 for ARM servers
		if _, err := exec.LookPath("qemu-system-aarch64"); err != nil {
			qemuNote = " (QEMU not found — Kata may fail to start VMs)"
		}
	} else {
		// Get QEMU version for details
		if out, err := exec.CommandContext(ctx, qemuPath, "--version").Output(); err == nil {
			lines := strings.SplitN(string(out), "\n", 2)
			if len(lines) > 0 {
				qemuNote = fmt.Sprintf(" (QEMU: %s)", strings.TrimSpace(lines[0]))
			}
		}
	}

	// Step 4: Check Kata config file
	kataConfigPaths := []string{
		"/etc/kata-containers/configuration.toml",
		"/opt/kata/share/defaults/kata-containers/configuration.toml",
		"/usr/share/kata-containers/defaults/configuration.toml",
	}
	configFound := ""
	for _, path := range kataConfigPaths {
		if _, err := os.Stat(path); err == nil {
			configFound = path
			break
		}
	}

	// Step 5: Detect full runtime capabilities (reuse existing detection)
	caps := runtime.DetectRuntime(runtime.RuntimeAuto)

	// Build result
	var parts []string
	parts = append(parts, fmt.Sprintf("Runtime: %s", caps.RuntimeName))
	parts = append(parts, fmt.Sprintf("Tier: %s (rank %d/3)", caps.ProviderTier, caps.ProviderTier.TierRank()))
	parts = append(parts, fmt.Sprintf("VM isolation: %v", caps.VMIsolation))
	parts = append(parts, fmt.Sprintf("SEV-SNP: %v", caps.SEVSNPActive))
	if configFound != "" {
		parts = append(parts, fmt.Sprintf("Config: %s", configFound))
	} else {
		parts = append(parts, "Config: not found (using defaults)")
	}

	result.Status = StatusOK
	result.Message = formatKataMessage(caps)
	result.Details = strings.Join(parts, "\n") + qemuNote

	return result
}

func (c *KataRuntimeChecker) Fix(ctx context.Context, pm PackageManager) error {
	return fmt.Errorf("Kata Containers installation requires manual setup: see https://katacontainers.io/docs/")
}

// formatKataMessage returns a concise status line based on detected capabilities.
func formatKataMessage(caps types.RuntimeCapabilities) string {
	switch {
	case caps.SEVSNPActive && caps.KataAvailable:
		return "Kata Containers: Tier 1 Confidential (Kata + SEV-SNP)"
	case caps.KataAvailable:
		return "Kata Containers: Tier 2 Standard (Kata VM isolation)"
	default:
		return "Kata Containers: Available (runc fallback)"
	}
}

// buildKataInstallGuide returns installation instructions for Kata Containers.
func buildKataInstallGuide() string {
	return "Install Kata Containers for VM-based container isolation:\n" +
		"  Ubuntu/Debian: https://github.com/kata-containers/kata-containers/blob/main/docs/install/ubuntu-installation-guide.md\n" +
		"  Or via kata-deploy: kubectl apply -f kata-deploy.yaml\n" +
		"  Then configure containerd: add [plugins.\"io.containerd.grpc.v1.cri\".containerd.runtimes.kata] to /etc/containerd/config.toml\n" +
		"  Provider will operate as Tier 2 (Standard) without Kata"
}
