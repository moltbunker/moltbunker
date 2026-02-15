package p2p

import (
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestGetTierRateLimit_AllTiers(t *testing.T) {
	tests := []struct {
		tier  string
		rate  float64
		burst int
	}{
		{"unstaked", 10, 20},
		{"unknown", 20, 40},
		{string(types.StakingTierStarter), 50, 100},
		{string(types.StakingTierBronze), 100, 200},
		{string(types.StakingTierSilver), 200, 400},
		{string(types.StakingTierGold), 500, 1000},
		{string(types.StakingTierPlatinum), 1000, 2000},
	}

	for _, tt := range tests {
		rl := GetTierRateLimit(tt.tier)
		if float64(rl.Rate) != tt.rate {
			t.Errorf("tier %s: expected rate %f, got %f", tt.tier, tt.rate, float64(rl.Rate))
		}
		if rl.Burst != tt.burst {
			t.Errorf("tier %s: expected burst %d, got %d", tt.tier, tt.burst, rl.Burst)
		}
	}
}

func TestGetTierRateLimit_UnknownFallsToUnstaked(t *testing.T) {
	rl := GetTierRateLimit("nonexistent")
	expected := GetTierRateLimit("unstaked")
	if rl.Rate != expected.Rate || rl.Burst != expected.Burst {
		t.Errorf("unknown tier should fall back to unstaked limits")
	}
}

func TestGetTierBanDuration(t *testing.T) {
	tests := []struct {
		tier     string
		expected time.Duration
	}{
		{"unstaked", 1 * time.Hour},
		{string(types.StakingTierStarter), 30 * time.Minute},
		{string(types.StakingTierGold), 5 * time.Minute},
		{string(types.StakingTierPlatinum), 5 * time.Minute},
	}

	for _, tt := range tests {
		d := GetTierBanDuration(tt.tier)
		if d != tt.expected {
			t.Errorf("tier %s: expected ban %v, got %v", tt.tier, tt.expected, d)
		}
	}
}

func TestNewTieredLimiter(t *testing.T) {
	limiter := NewTieredLimiter(string(types.StakingTierSilver))
	if limiter == nil {
		t.Fatal("expected non-nil limiter")
	}
	if limiter.Burst() != 400 {
		t.Errorf("expected burst 400, got %d", limiter.Burst())
	}
}

func TestTierHierarchy(t *testing.T) {
	// Verify that higher tiers get higher limits
	tiers := []string{
		"unstaked",
		"unknown",
		string(types.StakingTierStarter),
		string(types.StakingTierBronze),
		string(types.StakingTierSilver),
		string(types.StakingTierGold),
		string(types.StakingTierPlatinum),
	}

	for i := 1; i < len(tiers); i++ {
		prev := GetTierRateLimit(tiers[i-1])
		curr := GetTierRateLimit(tiers[i])
		if curr.Rate < prev.Rate {
			t.Errorf("tier %s rate (%f) should be >= tier %s rate (%f)",
				tiers[i], float64(curr.Rate), tiers[i-1], float64(prev.Rate))
		}
	}
}
