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
)

// EdgeRegistryABI is the minimal read-only ABI for the BunkerEdgeRegistry
// contract. It contains only the two view functions EDGE-02 needs
// (isActiveEdgeProvider, getEdgeProviderInfo). Keeping the surface tiny avoids
// coupling the daemon to the full contract ABI.
// #nosec G101 -- not a credential: this is a Solidity contract ABI (JSON interface definition)
const EdgeRegistryABI = `[
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "isActiveEdgeProvider",
		"outputs": [{"name": "active", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getEdgeProviderInfo",
		"outputs": [
			{
				"name": "info",
				"type": "tuple",
				"components": [
					{"name": "nodeId", "type": "bytes32"},
					{"name": "region", "type": "bytes32"},
					{"name": "registeredAt", "type": "uint48"},
					{"name": "active", "type": "bool"},
					{"name": "frozen", "type": "bool"},
					{"name": "endpointURL", "type": "string"},
					{"name": "tlsPubkeyHash", "type": "bytes"}
				]
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`

// EdgeProviderData is the daemon-side representation of an edge provider's
// on-chain registration metadata. It mirrors BunkerEdgeRegistry.EdgeProviderInfo
// but uses Go-native types.
type EdgeProviderData struct {
	NodeID        [32]byte
	Region        [32]byte
	EndpointURL   string
	TLSPubkeyHash []byte
	RegisteredAt  time.Time
	Active        bool
	Frozen        bool
}

// EdgeRegistryReader is the read-only seam EDGE-02 consumes to gate reverse
// tunnels on edge-provider registration. Both the production contract reader
// and the in-memory mock implement it, so EDGE-02 can be developed and tested
// on darwin without a deployed contract.
type EdgeRegistryReader interface {
	// IsActiveEdgeProvider reports whether the address is a registered,
	// active, and unfrozen edge provider.
	IsActiveEdgeProvider(ctx context.Context, addr common.Address) (bool, error)

	// GetEdgeProviderInfo returns the edge provider's metadata. For an
	// unregistered address it returns (nil, nil) rather than an error.
	GetEdgeProviderInfo(ctx context.Context, addr common.Address) (*EdgeProviderData, error)
}

// edgeProviderInfoTuple matches the on-chain getEdgeProviderInfo tuple return
// for ABI decoding. The field order and Go types must mirror the tuple
// components positionally — go-ethereum copies the decoded tuple into this
// struct field-by-field (see edgeInfoResult), so no `abi` struct tags are
// required; uint48 registeredAt decodes to *big.Int.
type edgeProviderInfoTuple struct {
	NodeId        [32]byte
	Region        [32]byte
	RegisteredAt  *big.Int
	Active        bool
	Frozen        bool
	EndpointURL   string
	TlsPubkeyHash []byte
}

// EdgeRegistryContract is the production EdgeRegistryReader backed by an on-chain
// BunkerEdgeRegistry deployment. The contract address is supplied at runtime
// (from config), never hard-coded.
type EdgeRegistryContract struct {
	baseClient   *BaseClient
	contract     *bind.BoundContract
	contractABI  abi.ABI
	contractAddr common.Address
}

// NewEdgeRegistryContract creates a production edge-registry reader. The address
// is sourced from config at runtime (address-as-parameter, never a source
// constant). Returns an error if baseClient is nil or not connected.
func NewEdgeRegistryContract(baseClient *BaseClient, contractAddr common.Address) (*EdgeRegistryContract, error) {
	if baseClient == nil {
		return nil, fmt.Errorf("base client is required (use NewMockEdgeRegistryReader for testing)")
	}
	if !baseClient.IsConnected() {
		return nil, fmt.Errorf("base client not connected to RPC")
	}

	parsedABI, err := abi.JSON(strings.NewReader(EdgeRegistryABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse edge registry ABI: %w", err)
	}

	client := baseClient.Client()
	return &EdgeRegistryContract{
		baseClient:   baseClient,
		contract:     bind.NewBoundContract(contractAddr, parsedABI, client, client, client),
		contractABI:  parsedABI,
		contractAddr: contractAddr,
	}, nil
}

// IsActiveEdgeProvider implements EdgeRegistryReader against the on-chain contract.
func (ec *EdgeRegistryContract) IsActiveEdgeProvider(ctx context.Context, addr common.Address) (bool, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	var result []interface{}
	if err := ec.contract.Call(callOpts, &result, "isActiveEdgeProvider", addr); err != nil {
		return false, fmt.Errorf("isActiveEdgeProvider failed: %w", err)
	}
	if len(result) < 1 {
		return false, fmt.Errorf("unexpected result length")
	}
	active, ok := result[0].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected result type for isActiveEdgeProvider")
	}
	return active, nil
}

// edgeInfoResult wraps the single tuple output of getEdgeProviderInfo. Decoding
// the tuple into a struct requires a wrapper whose first field is the tuple
// struct: bind.BoundContract.Call routes a single-element results slice through
// abi.UnpackIntoInterface, whose copyAtomic step sets dst.Field(0) from the
// decoded tuple. Pointing Field(0) at edgeProviderInfoTuple makes go-ethereum
// copy the tuple field-by-field (positionally, no struct tags required).
// Decoding directly into edgeProviderInfoTuple instead would mis-target only
// its first field and panic.
type edgeInfoResult struct {
	Info edgeProviderInfoTuple
}

// GetEdgeProviderInfo implements EdgeRegistryReader against the on-chain contract.
func (ec *EdgeRegistryContract) GetEdgeProviderInfo(ctx context.Context, addr common.Address) (*EdgeProviderData, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	out := edgeInfoResult{}
	results := []interface{}{&out}
	if err := ec.contract.Call(callOpts, &results, "getEdgeProviderInfo", addr); err != nil {
		return nil, fmt.Errorf("getEdgeProviderInfo failed: %w", err)
	}
	return out.Info.toEdgeProviderData(), nil
}

// toEdgeProviderData converts the ABI-decoded tuple into the daemon-side
// representation. Kept separate from the RPC call so the struct-tag decode +
// type translation can be unit-tested without a live backend.
func (t edgeProviderInfoTuple) toEdgeProviderData() *EdgeProviderData {
	var registeredAt time.Time
	if t.RegisteredAt != nil && t.RegisteredAt.Sign() > 0 {
		registeredAt = time.Unix(t.RegisteredAt.Int64(), 0)
	}

	return &EdgeProviderData{
		NodeID:        t.NodeId,
		Region:        t.Region,
		EndpointURL:   t.EndpointURL,
		TLSPubkeyHash: t.TlsPubkeyHash,
		RegisteredAt:  registeredAt,
		Active:        t.Active,
		Frozen:        t.Frozen,
	}
}

// MockEdgeRegistryReader is an in-memory EdgeRegistryReader for tests and darwin
// builds. It is safe for concurrent use.
type MockEdgeRegistryReader struct {
	mu        sync.RWMutex
	providers map[common.Address]*EdgeProviderData
}

// NewMockEdgeRegistryReader creates an empty in-memory edge-registry reader.
func NewMockEdgeRegistryReader() *MockEdgeRegistryReader {
	return &MockEdgeRegistryReader{
		providers: make(map[common.Address]*EdgeProviderData),
	}
}

// RegisterMock adds or replaces an edge provider in the mock store. A nil data
// pointer is ignored.
func (m *MockEdgeRegistryReader) RegisterMock(addr common.Address, data *EdgeProviderData) {
	if data == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	// Store a copy so callers cannot mutate the stored record after the fact.
	cp := *data
	m.providers[addr] = &cp
}

// UnregisterMock removes an edge provider from the mock store.
func (m *MockEdgeRegistryReader) UnregisterMock(addr common.Address) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.providers, addr)
}

// IsActiveEdgeProvider implements EdgeRegistryReader against the in-memory store.
func (m *MockEdgeRegistryReader) IsActiveEdgeProvider(_ context.Context, addr common.Address) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.providers[addr]
	if !ok {
		return false, nil
	}
	return data.Active && !data.Frozen, nil
}

// GetEdgeProviderInfo implements EdgeRegistryReader against the in-memory store.
// An unregistered address returns (nil, nil), not an error.
func (m *MockEdgeRegistryReader) GetEdgeProviderInfo(_ context.Context, addr common.Address) (*EdgeProviderData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.providers[addr]
	if !ok {
		return nil, nil
	}
	cp := *data
	return &cp, nil
}

// Compile-time assertions that both implementations satisfy the interface.
var (
	_ EdgeRegistryReader = (*EdgeRegistryContract)(nil)
	_ EdgeRegistryReader = (*MockEdgeRegistryReader)(nil)
)
