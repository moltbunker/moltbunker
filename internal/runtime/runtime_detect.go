package runtime

import (
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"

	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// RuntimeRuncV2 is the standard runc OCI runtime.
	RuntimeRuncV2 = "io.containerd.runc.v2"

	// RuntimeKataV2 is the Kata Containers OCI runtime (VM isolation).
	RuntimeKataV2 = "io.containerd.kata.v2"

	// RuntimeKataQemuSNP is Kata with QEMU backend and SEV-SNP support.
	RuntimeKataQemuSNP = "io.containerd.kata-qemu-snp.v2"

	// RuntimeAuto lets the system auto-detect the best available runtime.
	RuntimeAuto = "auto"
)

// DetectRuntime selects the best OCI runtime based on preference and system capabilities.
//
// Logic:
//
//	preference="auto":
//	  macOS                                     → runc.v2, tier=dev
//	  Linux + kata shim + SEV-SNP               → kata-qemu-snp.v2, tier=confidential
//	  Linux + kata shim, no SNP                 → kata.v2, tier=standard
//	  Linux, no kata                            → runc.v2, tier=standard
//
//	preference=explicit runtime name            → use as-is, detect tier
func DetectRuntime(preference string) types.RuntimeCapabilities {
	if preference != RuntimeAuto && preference != "" {
		// Explicit runtime override — detect tier based on name
		tier := classifyTier(preference, isSEVSNPActive())
		kataAvail := IsKataAvailable()
		return types.RuntimeCapabilities{
			RuntimeName:  preference,
			ProviderTier: tier,
			KataAvailable: kataAvail,
			SEVSNPActive: isSEVSNPActive(),
			VMIsolation:  isKataRuntime(preference),
		}
	}

	// Auto-detect
	if goruntime.GOOS != "linux" {
		return types.RuntimeCapabilities{
			RuntimeName:  RuntimeRuncV2,
			ProviderTier: types.ProviderTierDev,
			KataAvailable: false,
			SEVSNPActive: false,
			VMIsolation:  false,
		}
	}

	kataAvail := IsKataAvailable()
	sevActive := isSEVSNPActive()

	if kataAvail && sevActive {
		return types.RuntimeCapabilities{
			RuntimeName:  RuntimeKataQemuSNP,
			ProviderTier: types.ProviderTierConfidential,
			KataAvailable: true,
			SEVSNPActive: true,
			VMIsolation:  true,
		}
	}

	if kataAvail {
		return types.RuntimeCapabilities{
			RuntimeName:  RuntimeKataV2,
			ProviderTier: types.ProviderTierStandard,
			KataAvailable: true,
			SEVSNPActive: false,
			VMIsolation:  true,
		}
	}

	return types.RuntimeCapabilities{
		RuntimeName:  RuntimeRuncV2,
		ProviderTier: types.ProviderTierStandard,
		KataAvailable: false,
		SEVSNPActive: false,
		VMIsolation:  false,
	}
}

// IsKataAvailable checks whether the Kata Containers runtime shim is installed.
func IsKataAvailable() bool {
	if goruntime.GOOS != "linux" {
		return false
	}
	_, err := exec.LookPath("containerd-shim-kata-v2")
	return err == nil
}

// isSEVSNPActive checks whether AMD SEV-SNP is active on the host.
//
// SNP requires SEV-ES, which in turn requires SEV. On kernels that don't
// actually support SNP (e.g. Ubuntu 24.04 kernel 6.8), the boot parameter
// kvm_amd.sev_snp=1 may be silently ignored while the sysfs parameter still
// reads "Y" — a false positive. We guard against this by also verifying that
// SEV-ES is enabled (sev_es sysfs param == "Y"), since SNP cannot function
// without ES.
func isSEVSNPActive() bool {
	if goruntime.GOOS != "linux" {
		return false
	}

	// SEV-ES is a hard requirement for SNP. If ES is not active, SNP
	// cannot work regardless of what sev_snp reports.
	if !isSysfsParamActive("/sys/module/kvm_amd/parameters/sev_es") {
		return false
	}

	// Method 1: sysfs parameter (direct, preferred on kernels that support it)
	if isSysfsParamActive("/sys/module/kvm_amd/parameters/sev_snp") {
		return true
	}

	// Method 2: kernel boot params + SEV confirmation
	// On some kernels the sev_snp sysfs param doesn't exist but the boot
	// param was accepted. Verify SEV is also active as a sanity check.
	cmdline, err := os.ReadFile("/proc/cmdline")
	if err != nil {
		return false
	}
	if !containsBootParam(string(cmdline), "kvm_amd.sev_snp=1") {
		return false
	}
	return isSysfsParamActive("/sys/module/kvm_amd/parameters/sev")
}

// isSysfsParamActive reads a kvm_amd sysfs parameter and returns true if
// it is "Y" or "y" (the standard boolean format for kernel module params).
func isSysfsParamActive(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return len(data) > 0 && (data[0] == 'Y' || data[0] == 'y')
}

// containsBootParam checks if a kernel boot parameter exists in /proc/cmdline.
func containsBootParam(cmdline, param string) bool {
	for _, field := range strings.Fields(cmdline) {
		if field == param {
			return true
		}
	}
	return false
}

// DetectProviderTier classifies the node's provider tier from current system state.
func DetectProviderTier() types.ProviderTier {
	caps := DetectRuntime(RuntimeAuto)
	return caps.ProviderTier
}

// classifyTier returns the provider tier for a given runtime name and SEV-SNP state.
func classifyTier(runtimeName string, sevActive bool) types.ProviderTier {
	if goruntime.GOOS != "linux" {
		return types.ProviderTierDev
	}
	if isKataRuntime(runtimeName) && sevActive {
		return types.ProviderTierConfidential
	}
	return types.ProviderTierStandard
}

// isKataRuntime returns true if the runtime name is a Kata variant.
func isKataRuntime(name string) bool {
	switch name {
	case RuntimeKataV2, RuntimeKataQemuSNP:
		return true
	}
	return false
}
