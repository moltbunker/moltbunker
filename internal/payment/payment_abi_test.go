package payment

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// TestDelegationABIRoundTrip pins the getDelegation output tuple shape at ABI
// parse time so a future edit that reintroduces the phantom rewardDebt field or
// the wrong integer width is caught immediately (regression guard for the
// silent-zero decode bug).
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

	// Exactly four components, in declaration order, with the correct types.
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

	// Assert no rewardDebt component lingers anywhere in the tuple.
	for _, n := range tuple.TupleRawNames {
		if n == "rewardDebt" {
			t.Fatal("phantom rewardDebt component must not exist in getDelegation tuple")
		}
	}

	// Type assertions: provider=address, amount=uint128, delegatedAt=uint48, active=bool.
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

// TestDelegationDecodeSilentZero_Regression encodes a DelegationInfo tuple with
// the corrected ABI and decodes it the same way production does — via
// UnpackIntoInterface into the canonical delegationInfoTuple — proving the
// decode succeeds and active=true is preserved (the old mismatched ABI silently
// returned zero values, so active was always false).
func TestDelegationDecodeSilentZero_Regression(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(DelegationContractABI))
	if err != nil {
		t.Fatalf("parse delegation ABI: %v", err)
	}
	args := parsed.Methods["getDelegation"].Outputs

	provider := common.HexToAddress("0x00000000000000000000000000000000000000aB")
	amount := big.NewInt(1000)
	delegatedAt := big.NewInt(1718000000)

	packed, err := args.Pack(delegationInfoTuple{
		Provider:    provider,
		Amount:      amount,
		DelegatedAt: delegatedAt,
		Active:      true,
	})
	if err != nil {
		t.Fatalf("pack tuple: %v", err)
	}

	// Decode exactly as GetDelegation does: tuple is the single field of an
	// outer wrapper, mapped positionally by UnpackIntoInterface.
	var out struct {
		Info delegationInfoTuple
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
		t.Error("active: want true (the bug always returned false)")
	}
}

// TestGetProviderConfigDecode encodes a ProviderDelegationConfig tuple in the
// Go-side ABI's declared shape and verifies all five components decode via the
// same UnpackIntoInterface path the contract client uses.
func TestGetProviderConfigDecode(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(DelegationContractABI))
	if err != nil {
		t.Fatalf("parse delegation ABI: %v", err)
	}
	method, ok := parsed.Methods["getProviderConfig"]
	if !ok {
		t.Fatal("getProviderConfig method missing from ABI")
	}
	args := method.Outputs
	tuple := args[0].Type
	if tuple.T != abi.TupleTy {
		t.Fatalf("expected tuple output, got %v", tuple.T)
	}
	if len(tuple.TupleElems) != 5 {
		t.Fatalf("expected 5 tuple components, got %d", len(tuple.TupleElems))
	}

	packed, err := args.Pack(providerConfigTuple{
		RewardCutBps:         1000,
		FeeShareBps:          250,
		AcceptDelegations:    true,
		PendingRewardCutBps:  1500,
		RewardCutEffectiveAt: big.NewInt(1718000000),
	})
	if err != nil {
		t.Fatalf("pack config tuple: %v", err)
	}

	var out struct {
		Config providerConfigTuple
	}
	if err := parsed.UnpackIntoInterface(&out, "getProviderConfig", packed); err != nil {
		t.Fatalf("UnpackIntoInterface: %v", err)
	}
	res := out.Config
	if res.RewardCutBps != 1000 {
		t.Errorf("rewardCutBps: want 1000, got %d", res.RewardCutBps)
	}
	if res.FeeShareBps != 250 {
		t.Errorf("feeShareBps: want 250, got %d", res.FeeShareBps)
	}
	if !res.AcceptDelegations {
		t.Error("acceptDelegations: want true")
	}
	if res.PendingRewardCutBps != 1500 {
		t.Errorf("pendingRewardCutBps: want 1500, got %d", res.PendingRewardCutBps)
	}
	if res.RewardCutEffectiveAt == nil || res.RewardCutEffectiveAt.Cmp(big.NewInt(1718000000)) != 0 {
		t.Errorf("rewardCutEffectiveAt: want 1718000000, got %v", res.RewardCutEffectiveAt)
	}
}
