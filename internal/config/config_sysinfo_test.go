package config

import (
	"runtime"
	"testing"
)

func TestDetectMemoryGB(t *testing.T) {
	mem := detectMemoryGB()
	if mem < 1 {
		t.Errorf("detectMemoryGB() returned %d, expected at least 1 GB", mem)
	}
	// Sanity upper bound: no consumer/server machine has 100TB
	if mem > 100*1024 {
		t.Errorf("detectMemoryGB() returned %d GB, seems unreasonably high", mem)
	}
}

func TestDetectStorageGB(t *testing.T) {
	storage := detectStorageGB()
	if storage < 1 {
		t.Errorf("detectStorageGB() returned %d, expected at least 1 GB", storage)
	}
	// Sanity upper bound: 1PB
	if storage > 1024*1024 {
		t.Errorf("detectStorageGB() returned %d GB, seems unreasonably high", storage)
	}
}

func TestDetectHardwareProfile(t *testing.T) {
	hw := detectHardwareProfile()

	// CPU
	if hw.CPUCores < 1 {
		t.Errorf("CPUCores = %d, expected at least 1", hw.CPUCores)
	}
	if hw.CPUThreads < 1 {
		t.Errorf("CPUThreads = %d, expected at least 1", hw.CPUThreads)
	}
	if hw.CPUSockets < 1 {
		t.Errorf("CPUSockets = %d, expected at least 1", hw.CPUSockets)
	}
	if hw.CPUThreads < hw.CPUCores {
		t.Errorf("CPUThreads (%d) < CPUCores (%d), threads should be >= cores", hw.CPUThreads, hw.CPUCores)
	}

	// CPU architecture must match runtime
	expectedArch := runtime.GOARCH
	if expectedArch == "amd64" {
		expectedArch = "x86_64"
	}
	if hw.CPUArch != expectedArch {
		t.Errorf("CPUArch = %q, expected %q", hw.CPUArch, expectedArch)
	}

	// Memory
	if hw.MemoryGB < 1 {
		t.Errorf("MemoryGB = %d, expected at least 1", hw.MemoryGB)
	}

	// Storage
	if hw.StorageGB < 1 {
		t.Errorf("StorageGB = %d, expected at least 1", hw.StorageGB)
	}

	// OS must match runtime
	if hw.OS != runtime.GOOS {
		t.Errorf("OS = %q, expected %q", hw.OS, runtime.GOOS)
	}

	// Hostname should be non-empty
	if hw.Hostname == "" {
		t.Error("Hostname should not be empty")
	}

	// BandwidthMbps should have a default
	if hw.BandwidthMbps < 1 {
		t.Errorf("BandwidthMbps = %d, expected at least 1", hw.BandwidthMbps)
	}

	// SEV-SNP level should be one of known values
	validLevels := map[string]bool{"none": true, "sev": true, "snp": true}
	if !validLevels[hw.SEVSNPLevel] {
		t.Errorf("SEVSNPLevel = %q, expected one of none/sev/snp", hw.SEVSNPLevel)
	}

	// TPM version should be one of known values
	validTPM := map[string]bool{"none": true, "1.2": true, "2.0": true}
	if !validTPM[hw.TPMVersion] {
		t.Errorf("TPMVersion = %q, expected one of none/1.2/2.0", hw.TPMVersion)
	}
}

func TestDetectHardwareProfile_Consistency(t *testing.T) {
	// Call twice and verify results are consistent
	hw1 := detectHardwareProfile()
	hw2 := detectHardwareProfile()

	if hw1.CPUCores != hw2.CPUCores {
		t.Errorf("CPUCores inconsistent: %d vs %d", hw1.CPUCores, hw2.CPUCores)
	}
	if hw1.MemoryGB != hw2.MemoryGB {
		t.Errorf("MemoryGB inconsistent: %d vs %d", hw1.MemoryGB, hw2.MemoryGB)
	}
	if hw1.OS != hw2.OS {
		t.Errorf("OS inconsistent: %q vs %q", hw1.OS, hw2.OS)
	}
	if hw1.CPUArch != hw2.CPUArch {
		t.Errorf("CPUArch inconsistent: %q vs %q", hw1.CPUArch, hw2.CPUArch)
	}
}
