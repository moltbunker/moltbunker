package payment

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/payment/bindings"
)

// TestDelegationABIRoundTrip pins the getDelegation output tuple shape in the
// hand-typed DelegationContractABI (still loaded for the Transact methods until
// ABIGEN-01 Phase 2) so a future edit that reintroduces the phantom rewardDebt
// field or the wrong integer width is caught immediately.
func TestDelegationABIRoundTrip(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(DelegationContractABI))
	if err != nil {
		t.Fatalf("parse delegation ABI: %v", err)
	}

	method, ok := parsed.Methods["getDelegation"]
	if !ok {
		t.Fatal("getDelegation method missing from ABI")
	}
	if len(method.Outputs) != 1 {
		t.Fatalf("expected 1 output (the tuple), got %d", len(method.Outputs))
	}

	tuple := method.Outputs[0].Type
	if tuple.T != abi.TupleTy {
		t.Fatalf("expected tuple output, got %v", tuple.T)
	}

	wantNames := []string{"provider", "amount", "delegatedAt", "active"}
	if len(tuple.TupleElems) != len(wantNames) {
		t.Fatalf("expected %d tuple components, got %d (%v)",
			len(wantNames), len(tuple.TupleElems), tuple.TupleRawNames)
	}
	for i, want := range wantNames {
		if tuple.TupleRawNames[i] != want {
			t.Errorf("component %d: want name %q, got %q", i, want, tuple.TupleRawNames[i])
		}
	}

	for _, n := range tuple.TupleRawNames {
		if n == "rewardDebt" {
			t.Fatal("phantom rewardDebt component must not exist in getDelegation tuple")
		}
	}

	if got := tuple.TupleElems[0].String(); got != "address" {
		t.Errorf("provider type: want address, got %s", got)
	}
	if got := tuple.TupleElems[1].String(); got != "uint128" {
		t.Errorf("amount type: want uint128, got %s", got)
	}
	if got := tuple.TupleElems[2].String(); got != "uint48" {
		t.Errorf("delegatedAt type: want uint48, got %s", got)
	}
	if got := tuple.TupleElems[3].String(); got != "bool" {
		t.Errorf("active type: want bool, got %s", got)
	}
}

// TestDelegationGeneratedBindingDecode packs a DelegationInfo tuple using the
// generated binding's ABI and decodes it into the generated
// BunkerDelegationDelegationInfo struct — proving the production read path
// (GetDelegation -> caller.GetDelegation) round-trips and active=true survives
// (the old mismatched ABI silently zeroed it).
func TestDelegationGeneratedBindingDecode(t *testing.T) {
	parsed, err := bindings.BunkerDelegationMetaData.GetAbi()
	if err != nil {
		t.Fatalf("load generated binding ABI: %v", err)
	}
	args := parsed.Methods["getDelegation"].Outputs

	provider := common.HexToAddress("0x00000000000000000000000000000000000000aB")
	amount := big.NewInt(1000)
	delegatedAt := big.NewInt(1718000000)

	packed, err := args.Pack(bindings.BunkerDelegationDelegationInfo{
		Provider:    provider,
		Amount:      amount,
		DelegatedAt: delegatedAt,
		Active:      true,
	})
	if err != nil {
		t.Fatalf("pack tuple: %v", err)
	}

	var out struct {
		Info bindings.BunkerDelegationDelegationInfo
	}
	if err := parsed.UnpackIntoInterface(&out, "getDelegation", packed); err != nil {
		t.Fatalf("UnpackIntoInterface: %v", err)
	}
	res := out.Info
	if res.Provider != provider {
		t.Errorf("provider: want %s, got %s", provider.Hex(), res.Provider.Hex())
	}
	if res.Amount == nil || res.Amount.Cmp(amount) != 0 {
		t.Errorf("amount: want %s, got %v", amount, res.Amount)
	}
	if res.DelegatedAt == nil || res.DelegatedAt.Cmp(delegatedAt) != 0 {
		t.Errorf("delegatedAt: want %s, got %v", delegatedAt, res.DelegatedAt)
	}
	if !res.Active {
		t.Error("active: want true (the old bug always returned false)")
	}
}

// TestProviderConfigGeneratedBindingDecode packs a ProviderDelegationConfig tuple
// using the generated binding's ABI (six fields, on-chain order) and decodes it
// into BunkerDelegationProviderDelegationConfig. This is the field that the
// hand-typed ABI got wrong (five fields, wrong order, acceptDelegations instead
// of acceptingDelegations, no totalDelegated) — the generated binding decodes
// acceptingDelegations and totalDelegated correctly.
func TestProviderConfigGeneratedBindingDecode(t *testing.T) {
	parsed, err := bindings.BunkerDelegationMetaData.GetAbi()
	if err != nil {
		t.Fatalf("load generated binding ABI: %v", err)
	}
	args := parsed.Methods["getProviderConfig"].Outputs
	tuple := args[0].Type
	if tuple.T != abi.TupleTy {
		t.Fatalf("expected tuple output, got %v", tuple.T)
	}
	if len(tuple.TupleElems) != 6 {
		t.Fatalf("expected 6 tuple components, got %d", len(tuple.TupleElems))
	}

	packed, err := args.Pack(bindings.BunkerDelegationProviderDelegationConfig{
		RewardCutBps:         1000,
		PendingRewardCutBps:  1500,
		RewardCutEffectiveAt: big.NewInt(1718000000),
		FeeShareBps:          250,
		TotalDelegated:       big.NewInt(42),
		AcceptingDelegations: true,
	})
	if err != nil {
		t.Fatalf("pack config tuple: %v", err)
	}

	var out struct {
		Config bindings.BunkerDelegationProviderDelegationConfig
	}
	if err := parsed.UnpackIntoInterface(&out, "getProviderConfig", packed); err != nil {
		t.Fatalf("UnpackIntoInterface: %v", err)
	}
	res := out.Config
	if res.RewardCutBps != 1000 {
		t.Errorf("rewardCutBps: want 1000, got %d", res.RewardCutBps)
	}
	if res.PendingRewardCutBps != 1500 {
		t.Errorf("pendingRewardCutBps: want 1500, got %d", res.PendingRewardCutBps)
	}
	if res.FeeShareBps != 250 {
		t.Errorf("feeShareBps: want 250, got %d", res.FeeShareBps)
	}
	if res.TotalDelegated == nil || res.TotalDelegated.Cmp(big.NewInt(42)) != 0 {
		t.Errorf("totalDelegated: want 42, got %v", res.TotalDelegated)
	}
	if !res.AcceptingDelegations {
		t.Error("acceptingDelegations: want true (hand-typed ABI could not decode this field)")
	}
	if res.RewardCutEffectiveAt == nil || res.RewardCutEffectiveAt.Cmp(big.NewInt(1718000000)) != 0 {
		t.Errorf("rewardCutEffectiveAt: want 1718000000, got %v", res.RewardCutEffectiveAt)
	}
}

// forgeArtifact is the minimal shape we read out of a forge build artifact.
type forgeArtifact struct {
	ABI json.RawMessage `json:"abi"`
}

// TestDelegationABIMatchesSolidity is the schema-drift regression guard: it loads
// the canonical forge artifact for BunkerDelegation and asserts that the
// getDelegation and getProviderConfig output tuples match exactly what the
// generated Go binding expects. If forge has not produced the artifact (e.g. CI
// without the contracts toolchain), the test is skipped so `go test ./...` stays
// green. When the artifact is present, this prevents silent re-introduction of
// the decode mismatch that PAY-01/ABIGEN-01 addressed.
func TestDelegationABIMatchesSolidity(t *testing.T) {
	// internal/payment -> repo root is two levels up.
	artifactPath := filepath.Join("..", "..", "contracts", "out", "BunkerDelegation.sol", "BunkerDelegation.json")
	raw, err := os.ReadFile(artifactPath) // #nosec G304 -- fixed, repo-relative path to a build artifact in tests
	if err != nil {
		t.Skipf("forge artifact not present (%s); run `make gen-bindings` or `forge build`: %v", artifactPath, err)
	}

	var art forgeArtifact
	if err := json.Unmarshal(raw, &art); err != nil {
		t.Fatalf("unmarshal artifact: %v", err)
	}
	solABI, err := abi.JSON(strings.NewReader(string(art.ABI)))
	if err != nil {
		t.Fatalf("parse solidity ABI from artifact: %v", err)
	}

	// getDelegation: [provider address, amount uint128, delegatedAt uint48, active bool]
	delTuple := solABI.Methods["getDelegation"].Outputs[0].Type
	wantDel := []struct{ name, typ string }{
		{"provider", "address"},
		{"amount", "uint128"},
		{"delegatedAt", "uint48"},
		{"active", "bool"},
	}
	if len(delTuple.TupleElems) != len(wantDel) {
		t.Fatalf("getDelegation: want %d components, got %d", len(wantDel), len(delTuple.TupleElems))
	}
	for i, w := range wantDel {
		if delTuple.TupleRawNames[i] != w.name {
			t.Errorf("getDelegation component %d: want name %q, got %q", i, w.name, delTuple.TupleRawNames[i])
		}
		if got := delTuple.TupleElems[i].String(); got != w.typ {
			t.Errorf("getDelegation component %d (%s): want type %q, got %q", i, w.name, w.typ, got)
		}
	}

	// getProviderConfig: the on-chain six-field tuple in declaration order.
	cfgTuple := solABI.Methods["getProviderConfig"].Outputs[0].Type
	wantCfg := []struct{ name, typ string }{
		{"rewardCutBps", "uint16"},
		{"pendingRewardCutBps", "uint16"},
		{"rewardCutEffectiveAt", "uint48"},
		{"feeShareBps", "uint16"},
		{"totalDelegated", "uint128"},
		{"acceptingDelegations", "bool"},
	}
	if len(cfgTuple.TupleElems) != len(wantCfg) {
		t.Fatalf("getProviderConfig: want %d components, got %d (%v)",
			len(wantCfg), len(cfgTuple.TupleElems), cfgTuple.TupleRawNames)
	}
	for i, w := range wantCfg {
		if cfgTuple.TupleRawNames[i] != w.name {
			t.Errorf("getProviderConfig component %d: want name %q, got %q", i, w.name, cfgTuple.TupleRawNames[i])
		}
		if got := cfgTuple.TupleElems[i].String(); got != w.typ {
			t.Errorf("getProviderConfig component %d (%s): want type %q, got %q", i, w.name, w.typ, got)
		}
	}
}
