package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestManifestPopulatesEmptyAddresses verifies that when mock_payments is false
// and the operator supplies only chain_id (no contract addresses), Load() fills
// them in from the embedded canonical manifest.
func TestManifestPopulatesEmptyAddresses(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	yaml := `
node:
  role: provider
economics:
  mock_payments: false
  chain_id: 84532
`
	if err := os.WriteFile(configPath, []byte(yaml), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	const wantToken = "0x4cc3F5C0d2Ecb4118e214980906eFe5c880a6ceA"
	if cfg.Economics.TokenAddress != wantToken {
		t.Errorf("token_address = %q, want %q (from manifest)", cfg.Economics.TokenAddress, wantToken)
	}
	if cfg.Economics.StakingAddress == "" {
		t.Error("staking_address should be populated from manifest")
	}
	if cfg.Economics.VerificationAddress == "" {
		t.Error("verification_address should be populated from manifest")
	}
	// timelock -> governance_address mapping
	const wantGov = "0xcD8af28808749CD4B55a970f14DA250C8EAEd3C9"
	if cfg.Economics.GovernanceAddress != wantGov {
		t.Errorf("governance_address = %q, want %q (manifest timelock)", cfg.Economics.GovernanceAddress, wantGov)
	}
	// registry -> subdomain_registry_address mapping
	const wantReg = "0x3559A7D2E6F09eA74a295e654e0D6C22F921D4b5"
	if cfg.Economics.SubdomainRegistryAddress != wantReg {
		t.Errorf("subdomain_registry_address = %q, want %q", cfg.Economics.SubdomainRegistryAddress, wantReg)
	}
}

// TestManifestRespectsOperatorOverride verifies an operator-supplied address in
// YAML wins over the manifest value (manifest is fallback only).
func TestManifestRespectsOperatorOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	const override = "0xAAAAaAAaAAaAAaAAAAAaaaAaAaaaAAAAaaAaaAA1"
	yaml := `
node:
  role: provider
economics:
  mock_payments: false
  chain_id: 84532
  token_address: "` + override + `"
`
	if err := os.WriteFile(configPath, []byte(yaml), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Economics.TokenAddress != override {
		t.Errorf("operator override lost: token_address = %q, want %q", cfg.Economics.TokenAddress, override)
	}
	// Other fields still come from the manifest.
	if cfg.Economics.StakingAddress == "" {
		t.Error("staking_address should still be populated from manifest")
	}
}

// TestManifestNoopInMockMode verifies mock mode leaves addresses untouched.
func TestManifestNoopInMockMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Economics.MockPayments = true
	cfg.Economics.ChainID = 84532
	cfg.Economics.TokenAddress = ""
	populateAddressesFromManifest(cfg)
	if cfg.Economics.TokenAddress != "" {
		t.Errorf("mock mode should not populate addresses, got %q", cfg.Economics.TokenAddress)
	}
}

// TestManifestUnknownChainIsNoop verifies an unknown chain leaves fields empty
// (Validate() then surfaces the missing-address errors as before).
func TestManifestUnknownChainIsNoop(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Economics.MockPayments = false
	cfg.Economics.ChainID = 424242
	cfg.Economics.TokenAddress = ""
	populateAddressesFromManifest(cfg)
	if cfg.Economics.TokenAddress != "" {
		t.Errorf("unknown chain should not populate addresses, got %q", cfg.Economics.TokenAddress)
	}
}

func TestIsZeroAddress(t *testing.T) {
	cases := map[string]bool{
		"0x0000000000000000000000000000000000000000": true,
		"0000000000000000000000000000000000000000":   true,
		"0x4cc3F5C0d2Ecb4118e214980906eFe5c880a6ceA": false,
		"":   false,
		"0x": false,
	}
	for addr, want := range cases {
		if got := isZeroAddress(addr); got != want {
			t.Errorf("isZeroAddress(%q) = %v, want %v", addr, got, want)
		}
	}
}
