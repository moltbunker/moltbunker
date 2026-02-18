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

// RegistryContract provides an interface to the BunkerRegistry smart contract.
type RegistryContract struct {
	baseClient   *BaseClient
	contract     *bind.BoundContract
	contractABI  abi.ABI
	contractAddr common.Address
	mockMode     bool

	// Mock state
	mockNames map[string]*mockSubdomain // name → record
	mockMu    sync.RWMutex
}

type mockSubdomain struct {
	Owner        common.Address
	DeploymentID [32]byte
	RegisteredAt time.Time
}

// NewRegistryContract creates a new registry contract client.
func NewRegistryContract(baseClient *BaseClient, contractAddr common.Address) (*RegistryContract, error) {
	rc := &RegistryContract{
		baseClient:   baseClient,
		contractAddr: contractAddr,
		mockNames:    make(map[string]*mockSubdomain),
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
		owner := common.Address{}
		if rc.baseClient != nil {
			owner = rc.baseClient.Address()
		}
		rc.mockNames[name] = &mockSubdomain{
			Owner:        owner,
			DeploymentID: deploymentID,
			RegisteredAt: time.Now(),
		}
		return nil, nil
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

	opts, err := rc.baseClient.GetTransactOpts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transact opts: %w", err)
	}
	return rc.contract.Transact(opts, "updateDeployment", name, newDeploymentID)
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

// GetRegistrationFee returns the current registration fee in BUNKER tokens.
func (rc *RegistryContract) GetRegistrationFee(ctx context.Context) (*big.Int, error) {
	if rc.mockMode {
		return big.NewInt(10000), nil // 10K BUNKER (without decimals for mock)
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
