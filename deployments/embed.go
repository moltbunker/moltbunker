// Package deployments embeds the canonical contract-address manifest
// (addresses.json) so that the daemon can resolve contract addresses at
// runtime without any filesystem access.
//
// The manifest is the single source of truth for contract addresses across
// the daemon, the web dapp, the web-admin panel, and the Python SDK. It is
// regenerated for the non-Go consumers by tools/gen-addresses. Contract
// addresses are public on-chain facts and are committed to source; private
// keys and mnemonics MUST NEVER appear in this file.
package deployments

import _ "embed"

// AddressManifestJSON is the raw bytes of the embedded addresses.json manifest.
// Consumers (e.g. internal/deployment) unmarshal it into typed structs. The
// embed lives in this root package because //go:embed paths cannot traverse
// parent directories ("..") and addresses.json sits at the repository root.
//
//go:embed addresses.json
var AddressManifestJSON []byte
