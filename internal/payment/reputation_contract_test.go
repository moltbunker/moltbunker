package payment

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestReputationContract_MockRegisterAndScore(t *testing.T) {
	rc := NewMockReputationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	_, err := rc.RegisterProvider(ctx, provider)
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	score, err := rc.GetScore(ctx, provider)
	if err != nil {
		t.Fatalf("failed to get score: %v", err)
	}
	if score.Cmp(big.NewInt(500)) != 0 {
		t.Errorf("expected initial score 500, got %s", score.String())
	}
}

func TestReputationContract_MockJobCompleted(t *testing.T) {
	rc := NewMockReputationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	_, _ = rc.RegisterProvider(ctx, provider)
	_, _ = rc.RecordJobCompleted(ctx, provider)

	score, _ := rc.GetScore(ctx, provider)
	if score.Cmp(big.NewInt(510)) != 0 {
		t.Errorf("expected score 510 after job completed, got %s", score.String())
	}

	rep, _ := rc.GetReputation(ctx, provider)
	if rep.JobsCompleted.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("expected 1 job completed, got %s", rep.JobsCompleted.String())
	}
}

func TestReputationContract_MockJobFailed(t *testing.T) {
	rc := NewMockReputationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	_, _ = rc.RegisterProvider(ctx, provider)
	_, _ = rc.RecordJobFailed(ctx, provider)

	score, _ := rc.GetScore(ctx, provider)
	if score.Cmp(big.NewInt(450)) != 0 {
		t.Errorf("expected score 450 after job failed, got %s", score.String())
	}
}

func TestReputationContract_MockSlashEvent(t *testing.T) {
	rc := NewMockReputationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	_, _ = rc.RegisterProvider(ctx, provider)
	_, _ = rc.RecordSlashEvent(ctx, provider)

	score, _ := rc.GetScore(ctx, provider)
	if score.Cmp(big.NewInt(400)) != 0 {
		t.Errorf("expected score 400 after slash, got %s", score.String())
	}
}

func TestReputationContract_MockScoreFloors(t *testing.T) {
	rc := NewMockReputationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	_, _ = rc.RegisterProvider(ctx, provider)
	// 6 slashes × -100 = -600, from 500 should floor at 0
	for i := 0; i < 6; i++ {
		_, _ = rc.RecordSlashEvent(ctx, provider)
	}

	score, _ := rc.GetScore(ctx, provider)
	if score.Sign() != 0 {
		t.Errorf("expected score 0 (floored), got %s", score.String())
	}
}

func TestReputationContract_MockScoreCaps(t *testing.T) {
	rc := NewMockReputationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	_, _ = rc.RegisterProvider(ctx, provider)
	// 60 completions × +10 = +600, from 500 = 1100, should cap at 1000
	for i := 0; i < 60; i++ {
		_, _ = rc.RecordJobCompleted(ctx, provider)
	}

	score, _ := rc.GetScore(ctx, provider)
	if score.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("expected score capped at 1000, got %s", score.String())
	}
}

func TestReputationContract_MockTier(t *testing.T) {
	rc := NewMockReputationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	_, _ = rc.RegisterProvider(ctx, provider)

	tier, err := rc.GetTier(ctx, provider)
	if err != nil {
		t.Fatalf("failed to get tier: %v", err)
	}
	if tier != ReputationTierReliable {
		t.Errorf("expected Reliable tier for score 500, got %s", tier)
	}
}

func TestReputationContract_MockEligibility(t *testing.T) {
	rc := NewMockReputationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	// Not registered → not eligible
	eligible, _ := rc.IsEligibleForJobs(ctx, provider)
	if eligible {
		t.Error("expected not eligible before registration")
	}

	_, _ = rc.RegisterProvider(ctx, provider)

	// Score 500 ≥ 250 → eligible
	eligible, _ = rc.IsEligibleForJobs(ctx, provider)
	if !eligible {
		t.Error("expected eligible with score 500")
	}
}

func TestReputationContract_MockRecordEvent(t *testing.T) {
	rc := NewMockReputationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	_, _ = rc.RegisterProvider(ctx, provider)

	// Delta clamped to ±200
	_, err := rc.RecordEvent(ctx, provider, big.NewInt(500), "bonus")
	if err != nil {
		t.Fatalf("failed to record event: %v", err)
	}

	score, _ := rc.GetScore(ctx, provider)
	// 500 + 200 (clamped) = 700
	if score.Cmp(big.NewInt(700)) != 0 {
		t.Errorf("expected score 700 (clamped delta), got %s", score.String())
	}
}
