package runtime

import (
	"runtime"
	"testing"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestDetectRuntime_ExplicitOverride(t *testing.T) {
	caps := DetectRuntime(RuntimeRuncV2)
	if caps.RuntimeName != RuntimeRuncV2 {
		t.Errorf("expected runtime %s, got %s", RuntimeRuncV2, caps.RuntimeName)
	}
	if caps.VMIsolation {
		t.Error("runc should not have VM isolation")
	}
}

func TestDetectRuntime_Auto(t *testing.T) {
	caps := DetectRuntime(RuntimeAuto)

	// On any platform, auto should return a valid runtime
	if caps.RuntimeName == "" {
		t.Error("auto detection returned empty runtime name")
	}

	// On macOS, should always be runc + dev tier
	if runtime.GOOS == "darwin" {
		if caps.RuntimeName != RuntimeRuncV2 {
			t.Errorf("macOS should use runc, got %s", caps.RuntimeName)
		}
		if caps.ProviderTier != types.ProviderTierDev {
			t.Errorf("macOS should be dev tier, got %s", caps.ProviderTier)
		}
		if caps.KataAvailable {
			t.Error("Kata should not be available on macOS")
		}
	}
}

func TestDetectRuntime_EmptyPreference(t *testing.T) {
	// Empty string should behave like "auto"
	caps := DetectRuntime("")
	if caps.RuntimeName == "" {
		t.Error("empty preference should auto-detect, got empty runtime")
	}
}

func TestIsKataRuntime(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{RuntimeKataV2, true},
		{RuntimeKataQemuSNP, true},
		{RuntimeRuncV2, false},
		{"io.containerd.runsc.v1", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := isKataRuntime(tt.name); got != tt.expected {
			t.Errorf("isKataRuntime(%q) = %v, want %v", tt.name, got, tt.expected)
		}
	}
}

func TestClassifyTier(t *testing.T) {
	if runtime.GOOS != "linux" {
		// On non-Linux, everything is dev tier
		tier := classifyTier(RuntimeKataV2, true)
		if tier != types.ProviderTierDev {
			t.Errorf("non-Linux should always be dev, got %s", tier)
		}
		return
	}

	tests := []struct {
		runtimeName string
		sevActive   bool
		expected    types.ProviderTier
	}{
		{RuntimeKataQemuSNP, true, types.ProviderTierConfidential},
		{RuntimeKataV2, true, types.ProviderTierConfidential},
		{RuntimeKataV2, false, types.ProviderTierStandard},
		{RuntimeRuncV2, false, types.ProviderTierStandard},
		{RuntimeRuncV2, true, types.ProviderTierStandard},
	}

	for _, tt := range tests {
		got := classifyTier(tt.runtimeName, tt.sevActive)
		if got != tt.expected {
			t.Errorf("classifyTier(%q, %v) = %s, want %s", tt.runtimeName, tt.sevActive, got, tt.expected)
		}
	}
}

func TestDetectProviderTier(t *testing.T) {
	tier := DetectProviderTier()
	// Just verify it returns a valid tier
	switch tier {
	case types.ProviderTierConfidential, types.ProviderTierStandard, types.ProviderTierDev:
		// OK
	default:
		t.Errorf("unexpected tier: %s", tier)
	}
}
