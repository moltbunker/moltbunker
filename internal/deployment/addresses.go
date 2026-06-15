// Package deployment loads the canonical contract-address manifest
// (deployments/addresses.json) and exposes typed, chain-keyed lookups for the
// daemon.
//
// The manifest is embedded at build time via the root "deployments" package,
// so address resolution requires no filesystem access at startup. Contract
// addresses are public on-chain facts committed to source; this package never
// touches or assumes any private key material.
package deployment

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/moltbunker/moltbunker/deployments"
)

// ContractSet holds the deployed address of every moltbunker contract on a
// single chain. The JSON keys use the canonical short names shared with the
// web dapp's ContractName union and the Python SDK.
type ContractSet struct {
	Token        string `json:"token"`
	Staking      string `json:"staking"`
	Escrow       string `json:"escrow"`
	Pricing      string `json:"pricing"`
	Timelock     string `json:"timelock"`
	Delegation   string `json:"delegation"`
	Reputation   string `json:"reputation"`
	Verification string `json:"verification"`
	Registry     string `json:"registry"`
	Slashing     string `json:"slashing"`
}

// ChainAddresses is a single chain entry in the manifest.
type ChainAddresses struct {
	ChainName  string      `json:"chainName"`
	RPCURL     string      `json:"rpcUrl"`
	Contracts  ContractSet `json:"contracts"`
	DeployedAt string      `json:"deployedAt"`
	Note       string      `json:"note"`
}

// AddressManifest is the top-level manifest schema. Chains are keyed by
// stringified EVM chainId (JSON object keys must be strings).
type AddressManifest struct {
	Chains map[string]ChainAddresses `json:"chains"`
}

var (
	loadOnce       sync.Once
	loadedManifest *AddressManifest
	loadErr        error
)

// LoadAddresses parses the embedded manifest and returns it. The result is
// cached after the first successful parse; the embedded bytes never change at
// runtime, so repeated calls are cheap.
func LoadAddresses() (*AddressManifest, error) {
	loadOnce.Do(func() {
		var m AddressManifest
		if err := json.Unmarshal(deployments.AddressManifestJSON, &m); err != nil {
			loadErr = fmt.Errorf("parse embedded addresses.json: %w", err)
			return
		}
		if len(m.Chains) == 0 {
			loadErr = fmt.Errorf("embedded addresses.json contains no chains")
			return
		}
		loadedManifest = &m
	})
	return loadedManifest, loadErr
}

// AddressesForChain returns the ContractSet for the given EVM chainId. The
// second return value is false if the chain is not present in the manifest or
// the manifest failed to load.
func AddressesForChain(chainID int64) (*ContractSet, bool) {
	m, err := LoadAddresses()
	if err != nil || m == nil {
		return nil, false
	}
	entry, ok := m.Chains[fmt.Sprintf("%d", chainID)]
	if !ok {
		return nil, false
	}
	cs := entry.Contracts
	return &cs, true
}
