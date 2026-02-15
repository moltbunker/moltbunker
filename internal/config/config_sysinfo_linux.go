package config

import (
	"os"
	"runtime"
	"strings"
	"syscall"
)

// detectMemoryGB returns total system memory in GB on Linux
func detectMemoryGB() int {
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err == nil && info.Totalram > 0 {
		gb := int(info.Totalram * uint64(info.Unit) / (1024 * 1024 * 1024))
		if gb > 0 {
			return gb
		}
	}
	return 8
}

// detectStorageGB returns total disk space in GB on Linux
func detectStorageGB() int {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err == nil {
		gb := int(stat.Blocks * uint64(stat.Bsize) / (1024 * 1024 * 1024))
		if gb > 0 {
			return gb
		}
	}
	return 100
}

// detectHardwareProfile returns a HardwareProfile with auto-detected values on Linux
func detectHardwareProfile() HardwareProfile {
	hw := HardwareProfile{
		CPUCores:      runtime.NumCPU(),
		CPUThreads:    runtime.NumCPU(),
		CPUSockets:    1,
		MemoryGB:      detectMemoryGB(),
		StorageGB:     detectStorageGB(),
		MemoryType:    "unknown",
		StorageType:   "unknown",
		BandwidthMbps: 100,
		SEVSNPLevel:   "none",
		TPMVersion:    "none",
		OS:            "linux",
	}
	if hw.CPUCores < 1 {
		hw.CPUCores = 1
	}

	// CPU architecture
	hw.CPUArch = runtime.GOARCH
	if hw.CPUArch == "amd64" {
		hw.CPUArch = "x86_64"
	}

	// CPU model from /proc/cpuinfo
	hw.CPUModel, hw.CPUCores, hw.CPUSockets = parseProcCPUInfo()
	if hw.CPUCores < 1 {
		hw.CPUCores = runtime.NumCPU()
	}

	// Storage type detection via sysfs
	hw.StorageType = detectStorageType()

	// SEV-SNP detection
	hw.SEVSNPSupported, hw.SEVSNPLevel = detectSEVSNP()

	// TPM detection
	hw.TPMVersion = detectTPM()

	// OS version from /etc/os-release
	hw.OSVersion = detectOSVersion()

	// Kernel version via uname
	hw.Kernel = detectKernel()

	// Hostname
	if host, err := os.Hostname(); err == nil {
		hw.Hostname = host
	}

	return hw
}

// parseProcCPUInfo extracts CPU model, physical core count, and socket count
func parseProcCPUInfo() (model string, cores int, sockets int) {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "", 0, 1
	}

	physicalIDs := make(map[string]bool)
	coreIDs := make(map[string]bool) // "physID:coreID" for unique physical cores

	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "model name":
			if model == "" {
				model = val
			}
		case "physical id":
			physicalIDs[val] = true
		case "core id":
			// Build composite key with last physical ID
			var lastPhysID string
			for pid := range physicalIDs {
				lastPhysID = pid
			}
			coreIDs[lastPhysID+":"+val] = true
		}
	}

	sockets = len(physicalIDs)
	if sockets < 1 {
		sockets = 1
	}
	cores = len(coreIDs)
	return model, cores, sockets
}

// detectStorageType checks sysfs for NVMe/SSD/HDD
func detectStorageType() string {
	// Check for NVMe devices first
	if entries, err := os.ReadDir("/sys/class/nvme"); err == nil && len(entries) > 0 {
		return "NVMe"
	}

	// Check rotational flag on block devices
	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return "unknown"
	}

	for _, entry := range entries {
		name := entry.Name()
		// Skip loop, ram, dm devices
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") || strings.HasPrefix(name, "dm-") {
			continue
		}
		data, err := os.ReadFile("/sys/block/" + name + "/queue/rotational")
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(data)) == "0" {
			return "SSD"
		}
		return "HDD"
	}
	return "unknown"
}

// detectSEVSNP checks for AMD SEV-SNP support
func detectSEVSNP() (supported bool, level string) {
	// Check SEV-SNP parameter
	if data, err := os.ReadFile("/sys/module/kvm_amd/parameters/sev_snp"); err == nil {
		val := strings.TrimSpace(string(data))
		if val == "Y" || val == "1" {
			return true, "snp"
		}
	}

	// Check basic SEV
	if data, err := os.ReadFile("/sys/module/kvm_amd/parameters/sev"); err == nil {
		val := strings.TrimSpace(string(data))
		if val == "Y" || val == "1" {
			return true, "sev"
		}
	}

	// Check /dev/sev device
	if _, err := os.Stat("/dev/sev"); err == nil {
		return true, "sev"
	}

	return false, "none"
}

// detectTPM checks for TPM hardware
func detectTPM() string {
	data, err := os.ReadFile("/sys/class/tpm/tpm0/tpm_version_major")
	if err != nil {
		return "none"
	}
	ver := strings.TrimSpace(string(data))
	if ver == "2" {
		return "2.0"
	}
	if ver == "1" {
		return "1.2"
	}
	return ver
}

// detectOSVersion reads /etc/os-release for PRETTY_NAME
func detectOSVersion() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Linux"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			val := strings.TrimPrefix(line, "PRETTY_NAME=")
			val = strings.Trim(val, "\"")
			return val
		}
	}
	return "Linux"
}

// detectKernel returns the kernel version via uname syscall
func detectKernel() string {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return ""
	}
	// Convert [65]int8 to string
	var buf []byte
	for _, b := range uname.Release {
		if b == 0 {
			break
		}
		buf = append(buf, byte(b))
	}
	return string(buf)
}
