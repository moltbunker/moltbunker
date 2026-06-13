package deployment

import (
	"strings"
	"testing"
)

func TestLoadAddresses(t *testing.T) {
	m, err := LoadAddresses()
	if err != nil {
		t.Fatalf("LoadAddresses: %v", err)
	}
	if m == nil {
		t.Fatal("LoadAddresses returned nil manifest")
		return // unreachable after Fatal; satisfies staticcheck SA5011 nil-deref analysis
	}
	if _, ok := m.Chains["84532"]; !ok {
		t.Fatalf("expected Base Sepolia (84532) in manifest, got chains: %v", keys(m.Chains))
	}
	if _, ok := m.Chains["8453"]; !ok {
		t.Fatal("expected Base Mainnet (8453) in manifest")
	}
}

func TestAddressesForChain_Sepolia(t *testing.T) {
	cs, ok := AddressesForChain(84532)
	if !ok {
		t.Fatal("expected sepolia chain present")
	}
	const wantToken = "0x4cc3F5C0d2Ecb4118e214980906eFe5c880a6ceA"
	if cs.Token != wantToken {
		t.Fatalf("sepolia token = %q, want %q", cs.Token, wantToken)
	}
	if cs.Slashing == "" {
		t.Fatal("sepolia slashing address should be populated")
	}
}

func TestAddressesForChain_Unknown(t *testing.T) {
	if _, ok := AddressesForChain(99999); ok {
		t.Fatal("expected unknown chain 99999 to return false")
	}
}

func TestContractSetHasNoZeroOnTestnet(t *testing.T) {
	cs, ok := AddressesForChain(84532)
	if !ok {
		t.Fatal("expected sepolia present")
	}
	all := map[string]string{
		"token": cs.Token, "staking": cs.Staking, "escrow": cs.Escrow,
		"pricing": cs.Pricing, "timelock": cs.Timelock, "delegation": cs.Delegation,
		"reputation": cs.Reputation, "verification": cs.Verification,
		"registry": cs.Registry, "slashing": cs.Slashing,
	}
	for name, addr := range all {
		if isZero(addr) {
			t.Errorf("sepolia %s address is unexpectedly the zero address", name)
		}
		if !strings.HasPrefix(addr, "0x") {
			t.Errorf("sepolia %s address malformed: %q", name, addr)
		}
	}
}

// TestContractSetZeroOnMainnet documents the invariant that mainnet addresses
// remain the zero address until the real mainnet deploy lands in the manifest.
// When mainnet deploys, this test should be updated alongside addresses.json.
func TestContractSetZeroOnMainnet(t *testing.T) {
	cs, ok := AddressesForChain(8453)
	if !ok {
		t.Fatal("expected mainnet present")
	}
	if !isZero(cs.Token) {
		t.Fatalf("mainnet token is %q; if mainnet has deployed, update this test and the manifest invariant", cs.Token)
	}
}

func isZero(addr string) bool {
	h := strings.TrimPrefix(strings.TrimPrefix(addr, "0x"), "0X")
	if h == "" {
		return false
	}
	return strings.Trim(h, "0") == ""
}

func keys(m map[string]ChainAddresses) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
