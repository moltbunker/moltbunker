package payment

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/moltbunker/moltbunker/internal/logging"
)

const (
	attestationInterval     = 24 * time.Hour
	attestationMaxMissed    = 3 // strikes before suspension
	attestationSuspendDays  = 7
)

// VerificationContract provides interface to the BunkerVerification smart contract.
type VerificationContract struct {
	baseClient   *BaseClient
	contract     *bind.BoundContract
	contractABI  abi.ABI
	contractAddr common.Address
	mockMode     bool

	// Mock state
	mockAttestations map[common.Address]*AttestationData
	mockMu           sync.RWMutex
}

// NewVerificationContract creates a new verification contract client.
func NewVerificationContract(baseClient *BaseClient, contractAddr common.Address) (*VerificationContract, error) {
	vc := &VerificationContract{
		baseClient:       baseClient,
		contractAddr:     contractAddr,
		mockAttestations: make(map[common.Address]*AttestationData),
	}

	// Require a connected base client; use NewMockVerificationContract() for testing
	if baseClient == nil {
		return nil, fmt.Errorf("base client is required (use NewMockVerificationContract for testing)")
	}
	if !baseClient.IsConnected() {
		return nil, fmt.Errorf("base client not connected to RPC")
	}

	parsedABI, err := abi.JSON(strings.NewReader(VerificationContractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse verification ABI: %w", err)
	}
	vc.contractABI = parsedABI

	client := baseClient.Client()
	vc.contract = bind.NewBoundContract(contractAddr, parsedABI, client, client, client)

	return vc, nil
}

// NewMockVerificationContract creates a mock verification contract for testing.
func NewMockVerificationContract() *VerificationContract {
	return &VerificationContract{
		mockMode:         true,
		mockAttestations: make(map[common.Address]*AttestationData),
	}
}

// IsMockMode returns whether running in mock mode.
func (vc *VerificationContract) IsMockMode() bool {
	return vc.mockMode
}

// SubmitAttestation submits a hardware/software attestation hash.
func (vc *VerificationContract) SubmitAttestation(ctx context.Context, hash [32]byte) (*types.Transaction, error) {
	if vc.mockMode {
		return vc.mockSubmitAttestation(ctx, hash)
	}

	auth, err := vc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := vc.contract.Transact(auth, "submitAttestation", hash)
	if err != nil {
		return nil, fmt.Errorf("failed to submit attestation: %w", err)
	}
	return tx, nil
}

func (vc *VerificationContract) mockSubmitAttestation(_ context.Context, hash [32]byte) (*types.Transaction, error) {
	vc.mockMu.Lock()
	defer vc.mockMu.Unlock()

	var provider common.Address
	if vc.baseClient != nil {
		provider = vc.baseClient.Address()
	}

	att, exists := vc.mockAttestations[provider]
	if !exists {
		att = &AttestationData{}
		vc.mockAttestations[provider] = att
	}

	att.LastHash = hash
	att.LastTime = time.Now()
	att.MissedCount = 0
	att.Suspended = false
	att.SuspendedUntil = time.Time{}

	logging.Info("submitted attestation",
		"provider", provider.Hex(),
		"hash", fmt.Sprintf("%x", hash[:8]))
	return nil, nil
}

// CheckMissedAttestations checks and records missed attestations for a provider.
func (vc *VerificationContract) CheckMissedAttestations(ctx context.Context, provider common.Address) (*types.Transaction, error) {
	if vc.mockMode {
		return vc.mockCheckMissed(ctx, provider)
	}

	auth, err := vc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := vc.contract.Transact(auth, "checkMissedAttestations", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to check missed attestations: %w", err)
	}
	return tx, nil
}

func (vc *VerificationContract) mockCheckMissed(_ context.Context, provider common.Address) (*types.Transaction, error) {
	vc.mockMu.Lock()
	defer vc.mockMu.Unlock()

	att, exists := vc.mockAttestations[provider]
	if !exists {
		return nil, nil // Not registered
	}

	if time.Since(att.LastTime) > attestationInterval {
		att.MissedCount++
		if att.MissedCount >= attestationMaxMissed {
			att.Suspended = true
			att.SuspendedUntil = time.Now().Add(time.Duration(attestationSuspendDays) * 24 * time.Hour)
			logging.Info("provider suspended for missed attestations",
				"provider", provider.Hex(),
				"missed", att.MissedCount)
		}
	}

	return nil, nil
}

// ChallengeAttestation challenges a provider's attestation with proof of fraud.
func (vc *VerificationContract) ChallengeAttestation(ctx context.Context, provider common.Address, proof []byte) (*types.Transaction, error) {
	if vc.mockMode {
		vc.mockMu.Lock()
		defer vc.mockMu.Unlock()

		att, exists := vc.mockAttestations[provider]
		if !exists {
			return nil, fmt.Errorf("no attestation found for provider")
		}
		att.Suspended = true
		att.SuspendedUntil = time.Now().Add(time.Duration(attestationSuspendDays) * 24 * time.Hour)
		logging.Info("attestation challenged", "provider", provider.Hex())
		return nil, nil
	}

	auth, err := vc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}

	tx, err := vc.contract.Transact(auth, "challengeAttestation", provider, proof)
	if err != nil {
		return nil, fmt.Errorf("failed to challenge attestation: %w", err)
	}
	return tx, nil
}

// GetAttestation returns the attestation data for a provider.
func (vc *VerificationContract) GetAttestation(ctx context.Context, provider common.Address) (*AttestationData, error) {
	if vc.mockMode {
		vc.mockMu.RLock()
		defer vc.mockMu.RUnlock()
		att, exists := vc.mockAttestations[provider]
		if !exists {
			return &AttestationData{}, nil
		}
		return &AttestationData{
			LastHash:       att.LastHash,
			LastTime:       att.LastTime,
			MissedCount:    att.MissedCount,
			Suspended:      att.Suspended,
			SuspendedUntil: att.SuspendedUntil,
		}, nil
	}

	var result []interface{}
	err := vc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "getAttestation", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get attestation: %w", err)
	}

	data := &AttestationData{}
	if len(result) > 0 {
		if res, ok := result[0].(struct {
			LastHash       [32]byte
			LastTime       *big.Int
			MissedCount    uint16
			Suspended      bool
			SuspendedUntil *big.Int
		}); ok {
			data.LastHash = res.LastHash
			if res.LastTime != nil {
				data.LastTime = time.Unix(res.LastTime.Int64(), 0)
			}
			data.MissedCount = res.MissedCount
			data.Suspended = res.Suspended
			if res.SuspendedUntil != nil {
				data.SuspendedUntil = time.Unix(res.SuspendedUntil.Int64(), 0)
			}
		}
	}
	return data, nil
}

// IsAttestationCurrent checks if a provider's attestation is up to date.
func (vc *VerificationContract) IsAttestationCurrent(ctx context.Context, provider common.Address) (bool, error) {
	if vc.mockMode {
		vc.mockMu.RLock()
		defer vc.mockMu.RUnlock()
		att, exists := vc.mockAttestations[provider]
		if !exists {
			return false, nil
		}
		return time.Since(att.LastTime) <= attestationInterval && !att.Suspended, nil
	}

	var result []interface{}
	err := vc.contract.Call(&bind.CallOpts{Context: ctx}, &result, "isAttestationCurrent", provider)
	if err != nil {
		return false, fmt.Errorf("failed to check attestation: %w", err)
	}
	if len(result) > 0 {
		if current, ok := result[0].(bool); ok {
			return current, nil
		}
	}
	return false, nil
}

// ─── Admin Setters ───────────────────────────────────────────────────────────

// SetReinstatementCooldown adjusts the reinstatement cooldown after suspension.
func (vc *VerificationContract) SetReinstatementCooldown(ctx context.Context, newCooldown *big.Int) (*types.Transaction, error) {
	if vc.mockMode {
		logging.Info("mock: setReinstatementCooldown cooldown=%s", newCooldown)
		return nil, nil
	}
	auth, err := vc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	tx, err := vc.contract.Transact(auth, "setReinstatementCooldown", newCooldown)
	if err != nil {
		return nil, fmt.Errorf("failed to set reinstatement cooldown: %w", err)
	}
	return tx, nil
}
