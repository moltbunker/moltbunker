package payment

import (
	"context"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func sampleEdgeData() *EdgeProviderData {
	var nodeID, region [32]byte
	copy(nodeID[:], []byte("edge-node-1"))
	copy(region[:], []byte("us-east"))
	return &EdgeProviderData{
		NodeID:        nodeID,
		Region:        region,
		EndpointURL:   "https://edge1.moltbunker.dev",
		TLSPubkeyHash: []byte{0x00, 0x11, 0x22, 0x33},
		RegisteredAt:  time.Unix(1_700_000_000, 0),
		Active:        true,
		Frozen:        false,
	}
}

func TestMockEdgeRegistry_IsActive(t *testing.T) {
	ctx := context.Background()
	m := NewMockEdgeRegistryReader()
	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")

	// Unknown address returns false, no error.
	active, err := m.IsActiveEdgeProvider(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error for unknown address: %v", err)
	}
	if active {
		t.Fatalf("expected unknown address to be inactive")
	}

	// After registration, returns true.
	m.RegisterMock(addr, sampleEdgeData())
	active, err = m.IsActiveEdgeProvider(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error after register: %v", err)
	}
	if !active {
		t.Fatalf("expected registered active provider to be active")
	}

	// A frozen provider is reported inactive.
	frozen := sampleEdgeData()
	frozen.Frozen = true
	m.RegisterMock(addr, frozen)
	active, err = m.IsActiveEdgeProvider(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error for frozen provider: %v", err)
	}
	if active {
		t.Fatalf("expected frozen provider to be inactive")
	}

	// An explicitly inactive provider is reported inactive.
	inactive := sampleEdgeData()
	inactive.Active = false
	m.RegisterMock(addr, inactive)
	active, _ = m.IsActiveEdgeProvider(ctx, addr)
	if active {
		t.Fatalf("expected inactive provider to be inactive")
	}

	// After unregister, returns false.
	m.RegisterMock(addr, sampleEdgeData())
	m.UnregisterMock(addr)
	active, err = m.IsActiveEdgeProvider(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error after unregister: %v", err)
	}
	if active {
		t.Fatalf("expected unregistered address to be inactive")
	}
}

func TestMockEdgeRegistry_GetInfo(t *testing.T) {
	ctx := context.Background()
	m := NewMockEdgeRegistryReader()
	addr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	want := sampleEdgeData()
	m.RegisterMock(addr, want)

	got, err := m.GetEdgeProviderInfo(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil edge provider data")
	}
	if got.NodeID != want.NodeID {
		t.Errorf("NodeID mismatch: got %x want %x", got.NodeID, want.NodeID)
	}
	if got.Region != want.Region {
		t.Errorf("Region mismatch: got %x want %x", got.Region, want.Region)
	}
	if got.EndpointURL != want.EndpointURL {
		t.Errorf("EndpointURL mismatch: got %q want %q", got.EndpointURL, want.EndpointURL)
	}
	if string(got.TLSPubkeyHash) != string(want.TLSPubkeyHash) {
		t.Errorf("TLSPubkeyHash mismatch: got %x want %x", got.TLSPubkeyHash, want.TLSPubkeyHash)
	}
	if !got.RegisteredAt.Equal(want.RegisteredAt) {
		t.Errorf("RegisteredAt mismatch: got %v want %v", got.RegisteredAt, want.RegisteredAt)
	}

	// RegisterMock must store a copy: mutating the caller's struct afterward
	// must not change the stored record.
	want.EndpointURL = "https://mutated.example"
	got2, _ := m.GetEdgeProviderInfo(ctx, addr)
	if got2.EndpointURL == "https://mutated.example" {
		t.Errorf("stored record was mutated through the caller's pointer")
	}
}

func TestMockEdgeRegistry_ZeroAddress(t *testing.T) {
	ctx := context.Background()
	m := NewMockEdgeRegistryReader()

	got, err := m.GetEdgeProviderInfo(ctx, common.Address{})
	if err != nil {
		t.Fatalf("expected nil error for zero address, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil data for zero address, got %+v", got)
	}
}

func TestMockEdgeRegistry_Concurrency(t *testing.T) {
	ctx := context.Background()
	m := NewMockEdgeRegistryReader()

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			addr := common.BigToAddress(big.NewInt(int64(n) + 1))
			data := sampleEdgeData()
			for j := 0; j < 100; j++ {
				m.RegisterMock(addr, data)
				_, _ = m.IsActiveEdgeProvider(ctx, addr)
				_, _ = m.GetEdgeProviderInfo(ctx, addr)
				if j%2 == 0 {
					m.UnregisterMock(addr)
				}
			}
		}(i)
	}

	wg.Wait()
}

// abiEdgeProviderInfo mirrors the on-chain getEdgeProviderInfo tuple components
// in declaration order. It is only used by the ABI encoder in the test below to
// produce wire bytes; the decode under test happens through edgeProviderInfoTuple.
type abiEdgeProviderInfo struct {
	NodeId        [32]byte
	Region        [32]byte
	RegisteredAt  *big.Int
	Active        bool
	Frozen        bool
	EndpointURL   string
	TlsPubkeyHash []byte
}

// TestEdgeProviderInfo_ABIDecode exercises the on-chain decode path of
// GetEdgeProviderInfo without a live backend: it ABI-encodes a sample
// EdgeProviderInfo tuple with EdgeRegistryABI's getEdgeProviderInfo output
// arguments, then runs the exact production decode — abi.Unpack followed by the
// decodeEdgeProviderInfoResult type-assert + translation that
// GetEdgeProviderInfo performs on result[0]. This confirms the struct decode is
// correct, in particular the uint48 registeredAt (decoded as *big.Int) and the
// two bool fields, which previously had zero coverage. (It also pins the
// regression where the original results-slice form panicked on this tuple.)
func TestEdgeProviderInfo_ABIDecode(t *testing.T) {
	parsedABI, err := abi.JSON(strings.NewReader(EdgeRegistryABI))
	if err != nil {
		t.Fatalf("failed to parse edge registry ABI: %v", err)
	}
	method, ok := parsedABI.Methods["getEdgeProviderInfo"]
	if !ok {
		t.Fatalf("getEdgeProviderInfo not present in ABI")
	}

	var nodeID, region [32]byte
	copy(nodeID[:], []byte("edge-node-abc"))
	copy(region[:], []byte("eu-west"))

	const registeredAtUnix int64 = 1_700_000_123
	sample := abiEdgeProviderInfo{
		NodeId:        nodeID,
		Region:        region,
		RegisteredAt:  big.NewInt(registeredAtUnix),
		Active:        true,
		Frozen:        true,
		EndpointURL:   "https://edge-abc.moltbunker.dev",
		TlsPubkeyHash: []byte{0xde, 0xad, 0xbe, 0xef, 0x00, 0x99},
	}

	// Encode exactly as the contract would return it (the tuple is the single
	// output argument of getEdgeProviderInfo).
	encoded, err := method.Outputs.Pack(sample)
	if err != nil {
		t.Fatalf("failed to ABI-encode sample tuple: %v", err)
	}

	// Decode through the exact production path bind.BoundContract.Call drives:
	// UnpackIntoInterface into &edgeInfoResult{} (its Info field is the tuple),
	// then the tuple->EdgeProviderData translation.
	out := edgeInfoResult{}
	if err := parsedABI.UnpackIntoInterface(&out, "getEdgeProviderInfo", encoded); err != nil {
		t.Fatalf("UnpackIntoInterface failed: %v", err)
	}
	data := out.Info.toEdgeProviderData()

	// Production translation into the daemon-side struct.
	if data.NodeID != nodeID {
		t.Errorf("data.NodeID mismatch: got %x want %x", data.NodeID, nodeID)
	}
	if data.Region != region {
		t.Errorf("data.Region mismatch: got %x want %x", data.Region, region)
	}
	if !data.RegisteredAt.Equal(time.Unix(registeredAtUnix, 0)) {
		t.Errorf("data.RegisteredAt mismatch: got %v want %v", data.RegisteredAt, time.Unix(registeredAtUnix, 0))
	}
	if !data.Active {
		t.Errorf("data.Active mismatch: got false want true")
	}
	if !data.Frozen {
		t.Errorf("data.Frozen mismatch: got false want true")
	}
	if data.EndpointURL != sample.EndpointURL {
		t.Errorf("data.EndpointURL mismatch: got %q want %q", data.EndpointURL, sample.EndpointURL)
	}
	if string(data.TLSPubkeyHash) != string(sample.TlsPubkeyHash) {
		t.Errorf("data.TLSPubkeyHash mismatch: got %x want %x", data.TLSPubkeyHash, sample.TlsPubkeyHash)
	}
}

// TestEdgeProviderInfo_ABIDecode_ZeroRegisteredAt confirms a zero uint48
// registeredAt maps to the Go zero time (not a 1970 epoch timestamp).
func TestEdgeProviderInfo_ABIDecode_ZeroRegisteredAt(t *testing.T) {
	parsedABI, err := abi.JSON(strings.NewReader(EdgeRegistryABI))
	if err != nil {
		t.Fatalf("failed to parse edge registry ABI: %v", err)
	}
	method := parsedABI.Methods["getEdgeProviderInfo"]

	sample := abiEdgeProviderInfo{
		RegisteredAt:  big.NewInt(0),
		TlsPubkeyHash: []byte{},
	}
	encoded, err := method.Outputs.Pack(sample)
	if err != nil {
		t.Fatalf("failed to ABI-encode sample tuple: %v", err)
	}

	out := edgeInfoResult{}
	if err := parsedABI.UnpackIntoInterface(&out, "getEdgeProviderInfo", encoded); err != nil {
		t.Fatalf("UnpackIntoInterface failed: %v", err)
	}
	data := out.Info.toEdgeProviderData()
	if !data.RegisteredAt.IsZero() {
		t.Errorf("expected zero RegisteredAt for unset uint48, got %v", data.RegisteredAt)
	}
	if data.Active || data.Frozen {
		t.Errorf("expected false bools for empty tuple, got active=%v frozen=%v", data.Active, data.Frozen)
	}
}

func TestMockEdgeRegistry_NilDataIgnored(t *testing.T) {
	ctx := context.Background()
	m := NewMockEdgeRegistryReader()
	addr := common.HexToAddress("0x3333333333333333333333333333333333333333")

	// Registering nil data is a no-op (no panic, address stays unknown).
	m.RegisterMock(addr, nil)
	active, err := m.IsActiveEdgeProvider(ctx, addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if active {
		t.Fatalf("expected address to stay inactive after nil register")
	}
}
