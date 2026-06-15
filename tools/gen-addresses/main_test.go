package main

import (
	"os"
	"strings"
	"testing"
)

// minimalManifest is a small two-chain manifest used by the generator tests.
func minimalManifest() *AddressManifest {
	full := func() map[string]string {
		mm := map[string]string{}
		for _, n := range contractNames {
			mm[n] = "0x1111111111111111111111111111111111111111"
		}
		return mm
	}
	return &AddressManifest{
		Chains: map[string]ChainEntry{
			"84532": {
				ChainName:  "Base Sepolia",
				RPCURL:     "https://sepolia.base.org",
				Contracts:  full(),
				DeployedAt: "2026-02-26T00:00:00Z",
				Note:       "testnet",
			},
			"8453": {
				ChainName: "Base Mainnet",
				RPCURL:    "https://mainnet.base.org",
				Contracts: func() map[string]string {
					mm := map[string]string{}
					for _, n := range contractNames {
						mm[n] = "0x0000000000000000000000000000000000000000"
					}
					return mm
				}(),
				Note: "pending mainnet deploy",
			},
		},
	}
}

func TestParseManifestValid(t *testing.T) {
	m, err := parseManifest("../../deployments/addresses.json")
	if err != nil {
		t.Fatalf("parseManifest: %v", err)
	}
	if _, ok := m.Chains["84532"]; !ok {
		t.Fatalf("expected chain 84532 in manifest")
	}
	if got := m.Chains["84532"].Contracts["token"]; !strings.HasPrefix(got, "0x") {
		t.Fatalf("token address malformed: %q", got)
	}
}

func TestParseManifestMissingFile(t *testing.T) {
	if _, err := parseManifest("../../deployments/does-not-exist.json"); err == nil {
		t.Fatal("expected error for missing manifest file")
	}
}

func TestValidateManifestMissingContract(t *testing.T) {
	m := minimalManifest()
	// Remove a required contract from one chain.
	c := m.Chains["84532"].Contracts
	delete(c, "slashing")
	m.Chains["84532"] = ChainEntry{
		ChainName: "Base Sepolia",
		RPCURL:    "x",
		Contracts: c,
	}
	if err := validateManifest(m); err == nil {
		t.Fatal("expected error when a required contract is missing")
	}
}

func TestValidateAddress(t *testing.T) {
	good := "0x4cc3F5C0d2Ecb4118e214980906eFe5c880a6ceA"
	zero := "0x0000000000000000000000000000000000000000"
	cases := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{"valid", good, false},
		{"zero-allowed", zero, false},
		{"empty", "", true},
		{"no-0x-prefix", "4cc3F5C0d2Ecb4118e214980906eFe5c880a6ceA", true},
		{"too-short", "0x1234", true},
		{"too-long", good + "ab", true},
		{"non-hex", "0xZZc3F5C0d2Ecb4118e214980906eFe5c880a6ceA", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAddress(tc.addr)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %q", tc.addr)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.addr, err)
			}
		})
	}
}

func TestGenerateTS(t *testing.T) {
	out, err := generateTS(minimalManifest())
	if err != nil {
		t.Fatalf("generateTS: %v", err)
	}
	mustContain := []string{
		"AUTO-GENERATED",
		"export interface ContractAddresses",
		"export const CHAIN_CONFIGS: Record<number, ChainConfig>",
		"export function getContracts",
		"84532:",
		"8453:",
		"chainName: \"Base Sepolia\"",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("generated TS missing %q\n---\n%s", s, out)
		}
	}
	for _, n := range contractNames {
		if !strings.Contains(out, n+":") {
			t.Errorf("generated TS missing contract field %q", n)
		}
	}
	// No trailing comma directly before a closing brace of the top-level map
	// would be a syntax error in some strict parsers; our template uses
	// trailing commas which are valid TS, so just sanity-check it parses-shaped.
	if strings.Contains(out, ",,") {
		t.Errorf("generated TS has a double comma")
	}
}

func TestGenerateYAML(t *testing.T) {
	out, err := generateYAML(minimalManifest())
	if err != nil {
		t.Fatalf("generateYAML: %v", err)
	}
	for _, s := range []string{"AUTO-GENERATED", "token_address", "governance_address", "subdomain_registry_address", "chain_id: 84532", "chain_id: 8453"} {
		if !strings.Contains(out, s) {
			t.Errorf("generated YAML missing %q\n---\n%s", s, out)
		}
	}
}

func TestGenerateEnvExample(t *testing.T) {
	out, err := generateEnvExample(minimalManifest())
	if err != nil {
		t.Fatalf("generateEnvExample: %v", err)
	}
	for _, s := range []string{"AUTO-GENERATED", "VITE_CHAIN_ID=84532", "VITE_TOKEN_ADDRESS", "OPTIONAL", "chain_id: 8453"} {
		if !strings.Contains(out, s) {
			t.Errorf("generated .env.example missing %q\n---\n%s", s, out)
		}
	}
}

func TestRunWritesYAML(t *testing.T) {
	dir := t.TempDir()
	out := dir + "/addresses-fragment.yaml"
	if err := run("../../deployments/addresses.json", out, "", "", ""); err != nil {
		t.Fatalf("run: %v", err)
	}
	data, err := os.ReadFile(out) // #nosec G304 -- test-controlled temp path
	if err != nil {
		t.Fatalf("read generated yaml: %v", err)
	}
	if !strings.Contains(string(data), "token_address") {
		t.Fatalf("generated yaml missing token_address")
	}
}
