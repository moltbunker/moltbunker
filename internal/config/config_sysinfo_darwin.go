package config

import (
	"os"
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"
)

// detectMemoryGB returns total system memory in GB on macOS
func detectMemoryGB() int {
	mem, err := unix.SysctlUint64("hw.memsize")
	if err == nil && mem > 0 {
		gb := int(mem / (1024 * 1024 * 1024))
		if gb > 0 {
			return gb
		}
	}
	return 8
}

// detectStorageGB returns total disk space in GB on macOS
func detectStorageGB() int {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err == nil {
		gb := int(uint64(stat.Blocks) * uint64(stat.Bsize) / (1024 * 1024 * 1024))
		if gb > 0 {
			return gb
		}
	}
	return 100
}

// detectHardwareProfile returns a HardwareProfile with auto-detected values on macOS
func detectHardwareProfile() HardwareProfile {
	hw := HardwareProfile{
		CPUCores:   runtime.NumCPU(),
		CPUThreads: runtime.NumCPU(),
		CPUSockets: 1,
		MemoryGB:   detectMemoryGB(),
		StorageGB:  detectStorageGB(),
		// macOS defaults
		MemoryType:      "unknown",
		StorageType:     "SSD",
		BandwidthMbps:   100,
		SEVSNPSupported: false,
		SEVSNPLevel:     "none",
		TPMVersion:      "none",
		OS:              "darwin",
	}
	if hw.CPUCores < 1 {
		hw.CPUCores = 1
	}

	// CPU model via sysctl
	if model, err := unix.Sysctl("machdep.cpu.brand_string"); err == nil && model != "" {
		hw.CPUModel = model
	}

	// CPU architecture
	hw.CPUArch = runtime.GOARCH
	if hw.CPUArch == "amd64" {
		hw.CPUArch = "x86_64"
	}

	// Physical core count
	if cores, err := unix.SysctlUint32("hw.physicalcpu"); err == nil && cores > 0 {
		hw.CPUCores = int(cores)
	}

	// OS version
	if ver, err := unix.Sysctl("kern.osproductversion"); err == nil && ver != "" {
		hw.OSVersion = "macOS " + ver
	}

	// Kernel version
	if kern, err := unix.Sysctl("kern.osrelease"); err == nil && kern != "" {
		hw.Kernel = kern
	}

	// Hostname
	if host, err := os.Hostname(); err == nil {
		hw.Hostname = host
	}

	return hw
}
