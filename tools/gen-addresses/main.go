// Command gen-addresses is the single-source-of-truth codegen for moltbunker
// contract addresses. It reads the canonical manifest (deployments/addresses.json)
// and emits derived, committed artifacts for every consumer:
//
//   - configs/addresses-fragment.yaml : human-readable reference / paste-in
//     template for daemon.yaml (the daemon itself resolves addresses via the
//     Go embed in internal/deployment, not this file).
//   - a TypeScript module (web + web-admin) exporting a typed CONTRACTS map.
//   - a web-admin .env.example with VITE_* override reference lines.
//
// The TS / env emitters are scaffolds: by default this tool only writes the
// in-repo YAML fragment. Pointing --out-web / --out-admin / --out-admin-env at
// the sibling web/ and web-admin/ repos is opt-in and lives in those repos'
// own sync workflows. This tool never reads or writes private key material;
// contract addresses are public on-chain facts.
//
// Mainnet cutover: edit deployments/addresses.json (the "8453" block), run
// `make gen-addresses`, and commit the diff.
//
// Uses only the Go standard library.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// contractNames is the canonical ordered list of contract short names. It must
// stay in sync with internal/deployment.ContractSet, the web ContractName
// union, and the Python SDK CONTRACT_NAMES list.
var contractNames = []string{
	"token", "staking", "escrow", "pricing", "timelock",
	"delegation", "reputation", "verification", "registry", "slashing",
}

// ChainEntry is a single chain block in the manifest.
type ChainEntry struct {
	ChainName  string            `json:"chainName"`
	RPCURL     string            `json:"rpcUrl"`
	Contracts  map[string]string `json:"contracts"`
	DeployedAt string            `json:"deployedAt"`
	Note       string            `json:"note"`
}

// AddressManifest is the top-level manifest schema.
type AddressManifest struct {
	Chains map[string]ChainEntry `json:"chains"`
}

func main() {
	manifestPath := flag.String("manifest", "deployments/addresses.json",
		"path to the canonical addresses.json manifest")
	outYAML := flag.String("out-yaml", "configs/addresses-fragment.yaml",
		"output path for the daemon YAML reference fragment")
	outWeb := flag.String("out-web", "",
		"output path for the web TypeScript module (empty = skip)")
	outAdmin := flag.String("out-admin", "",
		"output path for the web-admin TypeScript module (empty = skip)")
	outAdminEnv := flag.String("out-admin-env", "",
		"output path for the web-admin .env.example (empty = skip)")
	flag.Parse()

	if err := run(*manifestPath, *outYAML, *outWeb, *outAdmin, *outAdminEnv); err != nil {
		fmt.Fprintf(os.Stderr, "gen-addresses: %v\n", err)
		os.Exit(1)
	}
}

func run(manifestPath, outYAML, outWeb, outAdmin, outAdminEnv string) error {
	m, err := parseManifest(manifestPath)
	if err != nil {
		return err
	}
	if err := validateManifest(m); err != nil {
		return err
	}

	if outYAML != "" {
		yamlOut, err := generateYAML(m)
		if err != nil {
			return err
		}
		if err := writeFile(outYAML, yamlOut); err != nil {
			return err
		}
	}
	if outWeb != "" {
		tsOut, err := generateTS(m)
		if err != nil {
			return err
		}
		if err := writeFile(outWeb, tsOut); err != nil {
			return err
		}
	}
	if outAdmin != "" {
		tsOut, err := generateTS(m)
		if err != nil {
			return err
		}
		if err := writeFile(outAdmin, tsOut); err != nil {
			return err
		}
	}
	if outAdminEnv != "" {
		envOut, err := generateEnvExample(m)
		if err != nil {
			return err
		}
		if err := writeFile(outAdminEnv, envOut); err != nil {
			return err
		}
	}
	return nil
}

func parseManifest(path string) (*AddressManifest, error) {
	// #nosec G304 -- path is an operator/CI-supplied codegen input path, not request input
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest %q: %w", path, err)
	}
	var m AddressManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest %q: %w", path, err)
	}
	if len(m.Chains) == 0 {
		return nil, fmt.Errorf("manifest %q contains no chains", path)
	}
	return &m, nil
}

// validateManifest verifies every chain has every contract name and that each
// address is well-formed (0x + 40 hex). The zero address is permitted (a
// not-yet-deployed contract), but malformed addresses are a hard error.
func validateManifest(m *AddressManifest) error {
	for id, entry := range m.Chains {
		if entry.ChainName == "" {
			return fmt.Errorf("chain %s: missing chainName", id)
		}
		for _, name := range contractNames {
			addr, ok := entry.Contracts[name]
			if !ok {
				return fmt.Errorf("chain %s: missing required contract %q", id, name)
			}
			if err := validateAddress(addr); err != nil {
				return fmt.Errorf("chain %s contract %q: %w", id, name, err)
			}
		}
	}
	return nil
}

// validateAddress checks an address is 0x-prefixed and exactly 40 hex chars.
// The zero address is allowed; emptiness, bad prefix, wrong length, and
// non-hex characters are rejected.
func validateAddress(addr string) error {
	if addr == "" {
		return fmt.Errorf("empty address")
	}
	if !strings.HasPrefix(addr, "0x") && !strings.HasPrefix(addr, "0X") {
		return fmt.Errorf("address %q must start with 0x", addr)
	}
	hexPart := addr[2:]
	if len(hexPart) != 40 {
		return fmt.Errorf("address %q must be 42 chars (0x + 40 hex), got %d", addr, len(addr))
	}
	for _, c := range hexPart {
		isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !isHex {
			return fmt.Errorf("address %q contains non-hex character %q", addr, string(c))
		}
	}
	return nil
}

// sortedChainIDs returns chain ID keys in stable ascending order so generated
// output is deterministic (important for the stale-file git diff check).
func sortedChainIDs(m *AddressManifest) []string {
	ids := make([]string, 0, len(m.Chains))
	for id := range m.Chains {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

const codegenHeader = "AUTO-GENERATED by tools/gen-addresses — do not edit by hand. Source: deployments/addresses.json"

// goFieldFor maps a manifest contract short name to its daemon.yaml key.
// timelock -> governance_address and registry -> subdomain_registry_address
// reflect the existing EconomicsConfig field naming.
var yamlKeyFor = map[string]string{
	"token":        "token_address",
	"staking":      "staking_address",
	"escrow":       "escrow_address",
	"pricing":      "pricing_address",
	"timelock":     "governance_address",
	"delegation":   "delegation_address",
	"reputation":   "reputation_address",
	"verification": "verification_address",
	"registry":     "subdomain_registry_address",
	"slashing":     "slashing_address",
}

func generateYAML(m *AddressManifest) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", codegenHeader)
	b.WriteString("# Human-readable reference of resolved contract addresses per chain.\n")
	b.WriteString("# The daemon resolves these automatically from the embedded manifest\n")
	b.WriteString("# (internal/deployment) when economics.mock_payments is false; this file\n")
	b.WriteString("# is a paste source for operators who want to pin overrides in daemon.yaml.\n")
	for _, id := range sortedChainIDs(m) {
		entry := m.Chains[id]
		fmt.Fprintf(&b, "\n# %s (chain_id: %s)", entry.ChainName, id)
		if entry.Note != "" {
			fmt.Fprintf(&b, " — %s", entry.Note)
		}
		b.WriteString("\n")
		fmt.Fprintf(&b, "%s:\n", id)
		fmt.Fprintf(&b, "  chain_id: %s\n", id)
		fmt.Fprintf(&b, "  rpc_url: %q\n", entry.RPCURL)
		for _, name := range contractNames {
			fmt.Fprintf(&b, "  %s: %q\n", yamlKeyFor[name], entry.Contracts[name])
		}
	}
	return b.String(), nil
}

const tsTemplate = `// {{ .Header }}
// Single source of truth: deployments/addresses.json. Regenerate via ` + "`make gen-addresses`" + `.

export interface ContractAddresses {
{{- range .ContractNames }}
  {{ . }}: ` + "`0x${string}`" + `;
{{- end }}
}

export interface ChainConfig {
  chainName: string;
  rpcUrl: string;
  contracts: ContractAddresses;
}

export const CHAIN_CONFIGS: Record<number, ChainConfig> = {
{{- range .Chains }}
  {{ .ID }}: {
    chainName: {{ .ChainNameQ }},
    rpcUrl: {{ .RPCURLQ }},
    contracts: {
{{- range .Contracts }}
      {{ .Name }}: '{{ .Addr }}',
{{- end }}
    },
  },
{{- end }}
};

export function getContracts(chainId: number): ContractAddresses | undefined {
  return CHAIN_CONFIGS[chainId]?.contracts;
}
`

type tsContract struct {
	Name string
	Addr string
}

type tsChain struct {
	ID         string
	ChainNameQ string
	RPCURLQ    string
	Contracts  []tsContract
}

type tsData struct {
	Header        string
	ContractNames []string
	Chains        []tsChain
}

func generateTS(m *AddressManifest) (string, error) {
	data := tsData{Header: codegenHeader, ContractNames: contractNames}
	for _, id := range sortedChainIDs(m) {
		entry := m.Chains[id]
		ch := tsChain{
			ID:         id,
			ChainNameQ: jsonQuote(entry.ChainName),
			RPCURLQ:    jsonQuote(entry.RPCURL),
		}
		for _, name := range contractNames {
			ch.Contracts = append(ch.Contracts, tsContract{Name: name, Addr: entry.Contracts[name]})
		}
		data.Chains = append(data.Chains, ch)
	}
	tmpl, err := template.New("ts").Parse(tsTemplate)
	if err != nil {
		return "", fmt.Errorf("parse TS template: %w", err)
	}
	var b strings.Builder
	if err := tmpl.Execute(&b, data); err != nil {
		return "", fmt.Errorf("render TS: %w", err)
	}
	return b.String(), nil
}

// jsonQuote returns a double-quoted, JSON-escaped string literal usable in TS.
func jsonQuote(s string) string {
	out, _ := json.Marshal(s)
	return string(out)
}

// envVarFor maps a contract short name to its VITE_ env var name.
func envVarFor(name string) string {
	return "VITE_" + strings.ToUpper(name) + "_ADDRESS"
}

func generateEnvExample(m *AddressManifest) (string, error) {
	const defaultChain = "84532"
	entry, ok := m.Chains[defaultChain]
	if !ok {
		return "", fmt.Errorf("env example: default chain %s missing from manifest", defaultChain)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", codegenHeader)
	b.WriteString("#\n")
	b.WriteString("# These VITE_* lines are OPTIONAL per-deploy OVERRIDES. The manifest values\n")
	b.WriteString("# (deployments/addresses.json) are baked into the build via generated-addresses.ts,\n")
	b.WriteString("# so the base case is zero-config. Uncomment a line only to override one address.\n")
	b.WriteString("# NEVER put private keys, mnemonics, or API secrets in a committed file —\n")
	b.WriteString("# real secrets belong in the gitignored .env.local.\n")
	b.WriteString("\n")
	fmt.Fprintf(&b, "VITE_CHAIN_ID=%s\n", defaultChain)
	fmt.Fprintf(&b, "\n# %s (chain_id: %s) contract overrides:\n", entry.ChainName, defaultChain)
	for _, name := range contractNames {
		fmt.Fprintf(&b, "# %s=%s\n", envVarFor(name), entry.Contracts[name])
	}
	if mainnet, ok := m.Chains["8453"]; ok {
		fmt.Fprintf(&b, "\n# --- %s (chain_id: 8453) — uncomment after mainnet deploy ---\n", mainnet.ChainName)
		b.WriteString("# VITE_CHAIN_ID=8453\n")
		for _, name := range contractNames {
			fmt.Fprintf(&b, "# %s=%s\n", envVarFor(name), mainnet.Contracts[name])
		}
	}
	return b.String(), nil
}

func writeFile(path, content string) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create dir for %q: %w", path, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}
	return nil
}
