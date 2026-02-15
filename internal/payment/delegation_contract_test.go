package payment

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestDelegationContract_MockDelegate(t *testing.T) {
	dc := NewMockDelegationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	_, err := dc.Delegate(ctx, provider, parseWei("1000"))
	if err != nil {
		t.Fatalf("failed to delegate: %v", err)
	}

	total, err := dc.GetTotalDelegatedTo(ctx, provider)
	if err != nil {
		t.Fatalf("failed to get total delegated: %v", err)
	}
	if total.Cmp(parseWei("1000")) != 0 {
		t.Errorf("expected 1000 delegated, got %s", total.String())
	}
}

func TestDelegationContract_MockUndelegate(t *testing.T) {
	dc := NewMockDelegationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	_, _ = dc.Delegate(ctx, provider, parseWei("1000"))

	_, err := dc.RequestUndelegate(ctx, parseWei("500"))
	if err != nil {
		t.Fatalf("failed to request undelegate: %v", err)
	}

	// Completing should fail — unbonding period not over
	_, err = dc.CompleteUndelegate(ctx, big.NewInt(0))
	if err == nil {
		t.Fatal("expected error for incomplete unbonding period")
	}
}

func TestDelegationContract_MockGetDelegation(t *testing.T) {
	dc := NewMockDelegationContract()
	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")
	delegator := common.Address{} // zero address for mock (no baseClient)

	_, _ = dc.Delegate(ctx, provider, parseWei("500"))

	del, err := dc.GetDelegation(ctx, delegator)
	if err != nil {
		t.Fatalf("failed to get delegation: %v", err)
	}
	if del.Amount.Cmp(parseWei("500")) != 0 {
		t.Errorf("expected 500 delegated, got %s", del.Amount.String())
	}
	if del.Provider != provider {
		t.Errorf("expected provider %s, got %s", provider.Hex(), del.Provider.Hex())
	}
}

func TestDelegationContract_MockProviderConfig(t *testing.T) {
	dc := NewMockDelegationContract()
	ctx := context.Background()

	_, err := dc.SetDelegationConfig(ctx, 1500, 500)
	if err != nil {
		t.Fatalf("failed to set config: %v", err)
	}

	cfg, err := dc.GetProviderConfig(ctx, common.Address{})
	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}
	if cfg.RewardCutBps != 1500 {
		t.Errorf("expected rewardCutBps 1500, got %d", cfg.RewardCutBps)
	}
	if cfg.FeeShareBps != 500 {
		t.Errorf("expected feeShareBps 500, got %d", cfg.FeeShareBps)
	}
}

func TestDelegationContract_MockToggleDelegations(t *testing.T) {
	dc := NewMockDelegationContract()
	ctx := context.Background()

	_, err := dc.ToggleAcceptDelegations(ctx, false)
	if err != nil {
		t.Fatalf("failed to toggle: %v", err)
	}

	cfg, err := dc.GetProviderConfig(ctx, common.Address{})
	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}
	if cfg.AcceptDelegations {
		t.Error("expected AcceptDelegations=false")
	}
}
