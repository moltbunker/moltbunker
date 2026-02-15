package payment

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestVerificationContract_MockSubmitAttestation(t *testing.T) {
	vc := NewMockVerificationContract()
	ctx := context.Background()

	hash := [32]byte{1, 2, 3, 4}
	_, err := vc.SubmitAttestation(ctx, hash)
	if err != nil {
		t.Fatalf("failed to submit attestation: %v", err)
	}

	att, err := vc.GetAttestation(ctx, common.Address{})
	if err != nil {
		t.Fatalf("failed to get attestation: %v", err)
	}
	if att.LastHash != hash {
		t.Error("attestation hash mismatch")
	}
	if att.Suspended {
		t.Error("expected not suspended after fresh attestation")
	}
}

func TestVerificationContract_MockIsAttestationCurrent(t *testing.T) {
	vc := NewMockVerificationContract()
	ctx := context.Background()
	provider := common.Address{}

	// Not registered → not current
	current, _ := vc.IsAttestationCurrent(ctx, provider)
	if current {
		t.Error("expected not current before any attestation")
	}

	// Submit → current
	_, _ = vc.SubmitAttestation(ctx, [32]byte{1})
	current, _ = vc.IsAttestationCurrent(ctx, provider)
	if !current {
		t.Error("expected current after fresh attestation")
	}
}

func TestVerificationContract_MockChallenge(t *testing.T) {
	vc := NewMockVerificationContract()
	ctx := context.Background()
	provider := common.Address{}

	_, _ = vc.SubmitAttestation(ctx, [32]byte{1})

	_, err := vc.ChallengeAttestation(ctx, provider, []byte("fraud proof"))
	if err != nil {
		t.Fatalf("failed to challenge: %v", err)
	}

	att, _ := vc.GetAttestation(ctx, provider)
	if !att.Suspended {
		t.Error("expected suspended after challenge")
	}

	current, _ := vc.IsAttestationCurrent(ctx, provider)
	if current {
		t.Error("expected not current when suspended")
	}
}

func TestVerificationContract_MockChallengeNoAttestation(t *testing.T) {
	vc := NewMockVerificationContract()
	ctx := context.Background()

	_, err := vc.ChallengeAttestation(ctx, common.HexToAddress("0xdead"), []byte("proof"))
	if err == nil {
		t.Error("expected error when challenging non-existent attestation")
	}
}
