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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/moltbunker/moltbunker/internal/logging"
)

// RegistryContract provides an interface to the BunkerRegistry smart contract.
type RegistryContract struct {
	baseClient    *BaseClient
	tokenContract *TokenContract
	contract      *bind.BoundContract
	contractABI   abi.ABI
	contractAddr  common.Address
	mockMode      bool

	// Mock state
	mockNames map[string]*mockSubdomain // name → record
	mockMu    sync.RWMutex
}

type mockSubdomain struct {
	Owner         common.Address
	DeploymentID  [32]byte
	RegisteredAt  time.Time
	ExpiresAt     time.Time
	ReservedUntil time.Time
	Referrer      common.Address
	Description   string
	AvatarURL     string
	PrimaryName   bool // whether this is the primary name for its deployment
}

// NewRegistryContract creates a new registry contract client.
func NewRegistryContract(baseClient *BaseClient, tokenContract *TokenContract, contractAddr common.Address) (*RegistryContract, error) {
	rc := &RegistryContract{
		baseClient:    baseClient,
		tokenContract: tokenContract,
		contractAddr:  contractAddr,
		mockNames:     make(map[string]*mockSubdomain),
	}

	if baseClient == nil {
		return nil, fmt.Errorf("base client is required (use NewMockRegistryContract for testing)")
	}
	if !baseClient.IsConnected() {
		return nil, fmt.Errorf("base client not connected to RPC")
	}

	parsedABI, err := abi.JSON(strings.NewReader(BunkerRegistryABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry ABI: %w", err)
	}
	rc.contractABI = parsedABI

	client := baseClient.Client()
	rc.contract = bind.NewBoundContract(contractAddr, parsedABI, client, client, client)

	return rc, nil
}

// mockDefaultOwner is a deterministic non-zero address used as the default
// owner in mock mode when no baseClient is available. This avoids the issue
// where Resolve rejects zero-address owners as "not registered".
var mockDefaultOwner = common.HexToAddress("0x0000000000000000000000000000000000000001")

// NewMockRegistryContract creates a mock registry contract for testing.
func NewMockRegistryContract() *RegistryContract {
	return &RegistryContract{
		mockMode:  true,
		mockNames: make(map[string]*mockSubdomain),
	}
}

// Register registers a vanity subdomain name.
func (rc *RegistryContract) Register(ctx context.Context, name string, deploymentID [32]byte) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		if _, exists := rc.mockNames[name]; exists {
			return nil, fmt.Errorf("name already registered: %s", name)
		}
		owner := mockDefaultOwner
		if rc.baseClient != nil {
			owner = rc.baseClient.Address()
		}
		rc.mockNames[name] = &mockSubdomain{
			Owner:        owner,
			DeploymentID: deploymentID,
			RegisteredAt: time.Now(),
			ExpiresAt:    time.Now().Add(365 * 24 * time.Hour),
		}
		return nil, nil
	}

	if err := rc.approveNameFee(ctx, name); err != nil {
		return nil, err
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}

	tx, err := rc.contract.Transact(opts, "register", name, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}

	logging.Info("subdomain registration TX sent",
		"name", name,
		"tx", tx.Hash().Hex()[:16])
	return tx, nil
}

// RegisterWithReferral registers a subdomain with a referral discount.
func (rc *RegistryContract) RegisterWithReferral(ctx context.Context, name string, deploymentID [32]byte, referrer common.Address) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		if _, exists := rc.mockNames[name]; exists {
			return nil, fmt.Errorf("name already registered: %s", name)
		}
		owner := mockDefaultOwner
		if rc.baseClient != nil {
			owner = rc.baseClient.Address()
		}
		rc.mockNames[name] = &mockSubdomain{
			Owner:        owner,
			DeploymentID: deploymentID,
			RegisteredAt: time.Now(),
			ExpiresAt:    time.Now().Add(365 * 24 * time.Hour),
			Referrer:     referrer,
		}
		return nil, nil
	}

	if err := rc.approveNameFee(ctx, name); err != nil {
		return nil, err
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	tx, err := rc.contract.Transact(opts, "registerWithReferral", name, deploymentID, referrer)
	if err != nil {
		return nil, fmt.Errorf("registerWithReferral failed: %w", err)
	}
	logging.Info("subdomain registration with referral TX sent",
		"name", name,
		"referrer", referrer.Hex()[:10],
		"tx", tx.Hash().Hex()[:16])
	return tx, nil
}

// Release releases a subdomain name.
func (rc *RegistryContract) Release(ctx context.Context, name string) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		delete(rc.mockNames, name)
		return nil, nil
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "release", name)
}

// Transfer transfers a subdomain to a new owner.
func (rc *RegistryContract) Transfer(ctx context.Context, name string, newOwner common.Address) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		if rec, exists := rc.mockNames[name]; exists {
			rec.Owner = newOwner
		}
		return nil, nil
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "transfer", name, newOwner)
}

// UpdateDeployment updates the deployment ID a name points to.
func (rc *RegistryContract) UpdateDeployment(ctx context.Context, name string, newDeploymentID [32]byte) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		if rec, exists := rc.mockNames[name]; exists {
			rec.DeploymentID = newDeploymentID
		}
		return nil, nil
	}

	if err := rc.approveChangeFee(ctx); err != nil {
		return nil, err
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "updateDeployment", name, newDeploymentID)
}

// Renew extends a name's expiration.
func (rc *RegistryContract) Renew(ctx context.Context, name string) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		if rec, exists := rc.mockNames[name]; exists {
			rec.ExpiresAt = rec.ExpiresAt.Add(365 * 24 * time.Hour)
		}
		return nil, nil
	}

	if err := rc.approveNameFee(ctx, name); err != nil {
		return nil, err
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "renew", name)
}

// Reserve reserves a name for a limited time.
func (rc *RegistryContract) Reserve(ctx context.Context, name string) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		if _, exists := rc.mockNames[name]; exists {
			return nil, fmt.Errorf("name already registered: %s", name)
		}
		owner := mockDefaultOwner
		if rc.baseClient != nil {
			owner = rc.baseClient.Address()
		}
		rc.mockNames[name] = &mockSubdomain{
			Owner:         owner,
			RegisteredAt:  time.Now(),
			ExpiresAt:     time.Now().Add(365 * 24 * time.Hour),
			ReservedUntil: time.Now().Add(48 * time.Hour),
		}
		return nil, nil
	}

	if err := rc.approveNameFee(ctx, name); err != nil {
		return nil, err
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "reserve", name)
}

// ClaimReservation sets the deployment ID for a reserved name.
func (rc *RegistryContract) ClaimReservation(ctx context.Context, name string, deploymentID [32]byte) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		if rec, exists := rc.mockNames[name]; exists {
			rec.DeploymentID = deploymentID
			rec.ReservedUntil = time.Time{}
		}
		return nil, nil
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "claimReservation", name, deploymentID)
}

// CancelReservation cancels a name reservation.
func (rc *RegistryContract) CancelReservation(ctx context.Context, name string) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		delete(rc.mockNames, name)
		return nil, nil
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "cancelReservation", name)
}

// SetMetadata sets description and avatar URL for a name.
func (rc *RegistryContract) SetMetadata(ctx context.Context, name string, description string, avatarURL string) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		if rec, exists := rc.mockNames[name]; exists {
			rec.Description = description
			rec.AvatarURL = avatarURL
		}
		return nil, nil
	}

	if err := rc.approveChangeFee(ctx); err != nil {
		return nil, err
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "setMetadata", name, description, avatarURL)
}

// SetPrimaryName sets the primary name for reverse resolution.
func (rc *RegistryContract) SetPrimaryName(ctx context.Context, name string) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		if rec, exists := rc.mockNames[name]; exists {
			rec.PrimaryName = true
		}
		return nil, nil
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "setPrimaryName", name)
}

// ReclaimSquatted reclaims a squatted name.
func (rc *RegistryContract) ReclaimSquatted(ctx context.Context, name string) (*types.Transaction, error) {
	if rc.mockMode {
		rc.mockMu.Lock()
		defer rc.mockMu.Unlock()
		delete(rc.mockNames, name)
		return nil, nil
	}

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "reclaimSquatted", name)
}

// Resolve returns the registration data for a subdomain name.
func (rc *RegistryContract) Resolve(ctx context.Context, name string) (*SubdomainRegistration, error) {
	if rc.mockMode {
		rc.mockMu.RLock()
		defer rc.mockMu.RUnlock()
		rec, exists := rc.mockNames[name]
		if !exists || rec.Owner == (common.Address{}) {
			return nil, fmt.Errorf("name not registered: %s", name)
		}
		return &SubdomainRegistration{
			Name:         name,
			Owner:        rec.Owner,
			DeploymentID: rec.DeploymentID,
			RegisteredAt: rec.RegisteredAt,
		}, nil
	}

	callOpts := &bind.CallOpts{Context: ctx}
	var result []interface{}
	err := rc.contract.Call(callOpts, &result, "resolve", name)
	if err != nil {
		return nil, fmt.Errorf("resolve failed: %w", err)
	}
	if len(result) < 3 {
		return nil, fmt.Errorf("unexpected result length: %d", len(result))
	}

	owner := result[0].(common.Address)
	if owner == (common.Address{}) {
		return nil, fmt.Errorf("name not registered: %s", name)
	}

	depID := result[1].([32]byte)
	regAt := result[2].(*big.Int)

	return &SubdomainRegistration{
		Name:         name,
		Owner:        owner,
		DeploymentID: depID,
		RegisteredAt: time.Unix(regAt.Int64(), 0),
	}, nil
}

// IsAvailable checks if a subdomain name is available.
func (rc *RegistryContract) IsAvailable(ctx context.Context, name string) (bool, error) {
	if rc.mockMode {
		rc.mockMu.RLock()
		defer rc.mockMu.RUnlock()
		_, exists := rc.mockNames[name]
		return !exists, nil
	}

	callOpts := &bind.CallOpts{Context: ctx}
	var result []interface{}
	err := rc.contract.Call(callOpts, &result, "isAvailable", name)
	if err != nil {
		return false, fmt.Errorf("isAvailable failed: %w", err)
	}
	if len(result) < 1 {
		return false, fmt.Errorf("unexpected result length")
	}
	return result[0].(bool), nil
}

// IsExpired checks if a subdomain name is expired.
func (rc *RegistryContract) IsExpired(ctx context.Context, name string) (bool, error) {
	if rc.mockMode {
		rc.mockMu.RLock()
		defer rc.mockMu.RUnlock()
		rec, exists := rc.mockNames[name]
		if !exists {
			return false, nil
		}
		if rec.ExpiresAt.IsZero() {
			return false, nil
		}
		return time.Now().After(rec.ExpiresAt), nil
	}

	callOpts := &bind.CallOpts{Context: ctx}
	var result []interface{}
	err := rc.contract.Call(callOpts, &result, "isExpired", name)
	if err != nil {
		return false, fmt.Errorf("isExpired failed: %w", err)
	}
	if len(result) < 1 {
		return false, fmt.Errorf("unexpected result length")
	}
	return result[0].(bool), nil
}

// IsInGracePeriod checks if a name is in its post-expiry grace period.
func (rc *RegistryContract) IsInGracePeriod(ctx context.Context, name string) (bool, error) {
	if rc.mockMode {
		return false, nil
	}

	callOpts := &bind.CallOpts{Context: ctx}
	var result []interface{}
	err := rc.contract.Call(callOpts, &result, "isInGracePeriod", name)
	if err != nil {
		return false, fmt.Errorf("isInGracePeriod failed: %w", err)
	}
	if len(result) < 1 {
		return false, fmt.Errorf("unexpected result length")
	}
	return result[0].(bool), nil
}

// ReverseResolve looks up the primary name for a deployment ID.
func (rc *RegistryContract) ReverseResolve(ctx context.Context, deploymentID [32]byte) (string, error) {
	if rc.mockMode {
		rc.mockMu.RLock()
		defer rc.mockMu.RUnlock()
		for name, rec := range rc.mockNames {
			if rec.DeploymentID == deploymentID && rec.PrimaryName {
				return name, nil
			}
		}
		return "", nil
	}

	callOpts := &bind.CallOpts{Context: ctx}
	var result []interface{}
	err := rc.contract.Call(callOpts, &result, "reverseResolve", deploymentID)
	if err != nil {
		return "", fmt.Errorf("reverseResolve failed: %w", err)
	}
	if len(result) < 1 {
		return "", fmt.Errorf("unexpected result length")
	}
	return result[0].(string), nil
}

// CalculatePrice returns the registration price for a name including premium and staking discounts.
func (rc *RegistryContract) CalculatePrice(ctx context.Context, name string, user common.Address) (*big.Int, error) {
	if rc.mockMode {
		return big.NewInt(1000000), nil
	}

	callOpts := &bind.CallOpts{Context: ctx}
	var result []interface{}
	err := rc.contract.Call(callOpts, &result, "calculatePrice", name, user)
	if err != nil {
		return nil, fmt.Errorf("calculatePrice failed: %w", err)
	}
	if len(result) < 1 {
		return nil, fmt.Errorf("unexpected result length")
	}
	return result[0].(*big.Int), nil
}

// GetRegistrationFee returns the current base registration fee in BUNKER tokens.
func (rc *RegistryContract) GetRegistrationFee(ctx context.Context) (*big.Int, error) {
	if rc.mockMode {
		return big.NewInt(1000000), nil // 1M BUNKER (without decimals for mock)
	}

	callOpts := &bind.CallOpts{Context: ctx}
	var result []interface{}
	err := rc.contract.Call(callOpts, &result, "registrationFee")
	if err != nil {
		return nil, fmt.Errorf("registrationFee failed: %w", err)
	}
	if len(result) < 1 {
		return nil, fmt.Errorf("unexpected result length")
	}
	return result[0].(*big.Int), nil
}

// GetChangeFee returns the current change fee for updateDeployment and setMetadata.
func (rc *RegistryContract) GetChangeFee(ctx context.Context) (*big.Int, error) {
	if rc.mockMode {
		return big.NewInt(10000), nil // 10K BUNKER
	}

	callOpts := &bind.CallOpts{Context: ctx}
	var result []interface{}
	err := rc.contract.Call(callOpts, &result, "changeFee")
	if err != nil {
		return nil, fmt.Errorf("changeFee failed: %w", err)
	}
	if len(result) < 1 {
		return nil, fmt.Errorf("unexpected result length")
	}
	return result[0].(*big.Int), nil
}

// GetMetadata returns the metadata for a subdomain.
func (rc *RegistryContract) GetMetadata(ctx context.Context, name string) (*SubdomainMetadata, error) {
	if rc.mockMode {
		rc.mockMu.RLock()
		defer rc.mockMu.RUnlock()
		rec, exists := rc.mockNames[name]
		if !exists {
			return &SubdomainMetadata{}, nil
		}
		return &SubdomainMetadata{
			Description: rec.Description,
			AvatarURL:   rec.AvatarURL,
		}, nil
	}

	callOpts := &bind.CallOpts{Context: ctx}
	nameHash := crypto.Keccak256Hash([]byte(name))
	var result []interface{}
	err := rc.contract.Call(callOpts, &result, "metadata", nameHash)
	if err != nil {
		return nil, fmt.Errorf("metadata failed: %w", err)
	}
	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected result length: %d", len(result))
	}
	return &SubdomainMetadata{
		Description: result[0].(string),
		AvatarURL:   result[1].(string),
	}, nil
}

// approveNameFee calculates the actual price (with premiums + staking discounts) and approves the registry contract.
func (rc *RegistryContract) approveNameFee(ctx context.Context, name string) error {
	if rc.tokenContract == nil {
		return nil
	}
	fee, err := rc.CalculatePrice(ctx, name, rc.baseClient.Address())
	if err != nil {
		return fmt.Errorf("failed to calculate price for %q: %w", name, err)
	}
	if fee.Sign() == 0 {
		return nil
	}
	if _, err := rc.tokenContract.Approve(ctx, rc.contractAddr, fee); err != nil {
		return fmt.Errorf("failed to approve token spend: %w", err)
	}
	return nil
}

// approveChangeFee reads the on-chain changeFee and approves the registry contract.
func (rc *RegistryContract) approveChangeFee(ctx context.Context) error {
	if rc.tokenContract == nil {
		return nil
	}
	fee, err := rc.GetChangeFee(ctx)
	if err != nil {
		return fmt.Errorf("failed to get change fee: %w", err)
	}
	if fee.Sign() == 0 {
		return nil
	}
	if _, err := rc.tokenContract.Approve(ctx, rc.contractAddr, fee); err != nil {
		return fmt.Errorf("failed to approve token spend: %w", err)
	}
	return nil
}

// ListOwnedNames returns all subdomain registrations owned by an address.
func (rc *RegistryContract) ListOwnedNames(ctx context.Context, owner common.Address) ([]SubdomainRegistration, error) {
	if rc.mockMode {
		rc.mockMu.RLock()
		defer rc.mockMu.RUnlock()
		var results []SubdomainRegistration
		for name, rec := range rc.mockNames {
			if rec.Owner == owner {
				results = append(results, SubdomainRegistration{
					Name:         name,
					Owner:        rec.Owner,
					DeploymentID: rec.DeploymentID,
					RegisteredAt: rec.RegisteredAt,
				})
			}
		}
		return results, nil
	}

	// Get count first
	callOpts := &bind.CallOpts{Context: ctx}
	var countResult []interface{}
	if err := rc.contract.Call(callOpts, &countResult, "nameCount", owner); err != nil {
		return nil, fmt.Errorf("nameCount failed: %w", err)
	}
	count := countResult[0].(*big.Int).Int64()

	var results []SubdomainRegistration
	for i := int64(0); i < count; i++ {
		// Get name hash at index
		var hashResult []interface{}
		if err := rc.contract.Call(callOpts, &hashResult, "ownedNameAt", owner, big.NewInt(i)); err != nil {
			continue
		}
		nameHash := hashResult[0].([32]byte)

		// Get name string from hash
		var nameResult []interface{}
		if err := rc.contract.Call(callOpts, &nameResult, "nameOf", nameHash); err != nil {
			continue
		}
		name := nameResult[0].(string)
		if name == "" {
			continue
		}

		// Resolve full record
		reg, err := rc.Resolve(ctx, name)
		if err != nil {
			continue
		}
		results = append(results, *reg)
	}

	return results, nil
}
