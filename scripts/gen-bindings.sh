#!/usr/bin/env bash
#
# gen-bindings.sh — regenerate the Go contract bindings in
# internal/payment/bindings/ from the canonical forge artifacts in
# contracts/out/.
#
# Pipeline:
#   1. forge build           -> refresh contracts/out/<Name>.sol/<Name>.json
#   2. jq -r '.abi' …        -> extract the ABI array from each artifact
#   3. abigen --abi … --out  -> emit a typed Go binding into bindings/
#
# The generated files are committed to the repo, so a normal `go build` needs
# only the Go toolchain. This script is a developer tool, invoked intentionally
# (via `make gen-bindings` / `go generate ./internal/payment/...`) when the
# contracts change. CI uses `make bindings-check` to detect drift.
#
# Requirements on PATH (only when running this script, not for normal builds):
#   - forge   (foundry)   : compile Solidity
#   - jq                  : extract the .abi field from forge JSON
#   - abigen  (go-ethereum): generate Go bindings
#
# NOTE on abigen version: the blueprint specified pinning abigen to the go.mod
# go-ethereum version via `go run github.com/ethereum/go-ethereum/cmd/abigen`.
# That path does not link on modern Go toolchains (the v1.13.10 abigen depends
# on github.com/fjl/memsize, which references the now-removed runtime.stopTheWorld
# symbol). We therefore use the `abigen` binary on PATH. Its output is the
# stable abigen format and compiles cleanly against the pinned go-ethereum in
# go.mod (verified). The bindings-check target guards against any drift.

set -euo pipefail

# Resolve repo root from this script's location (scripts/ lives at the root).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

CONTRACTS_DIR="${ROOT_DIR}/contracts"
BINDINGS_DIR="${ROOT_DIR}/internal/payment/bindings"
PKG="bindings"

# Foundry may not be on PATH but installed in the standard location; mirror the
# Makefile's FOUNDRY_BIN fallback.
if ! command -v forge >/dev/null 2>&1 && [ -x "${HOME}/.foundry/bin/forge" ]; then
	export PATH="${HOME}/.foundry/bin:${PATH}"
fi

die() {
	echo "error: $*" >&2
	exit 1
}

# ─── Tool checks ──────────────────────────────────────────────────────────────
command -v forge >/dev/null 2>&1 || die "'forge' not found on PATH. Install foundry: https://getfoundry.sh (curl -L https://foundry.paradigm.xyz | bash && foundryup)"
command -v jq >/dev/null 2>&1 || die "'jq' not found on PATH. Install it (macOS: brew install jq, Debian/Ubuntu: apt-get install jq)."
command -v abigen >/dev/null 2>&1 || die "'abigen' not found on PATH. Install it: go install github.com/ethereum/go-ethereum/cmd/abigen@latest"

# ─── 1. Compile Solidity, refresh out/ ──────────────────────────────────────────
echo ">> forge build (refreshing contracts/out/)"
( cd "${CONTRACTS_DIR}" && forge build --quiet )

# ─── 2. + 3. Extract ABI + generate binding for each contract ───────────────────
mkdir -p "${BINDINGS_DIR}"

# The eight first-party contracts whose bindings we generate. Phase 1 (ABIGEN-01)
# only ADOPTS bunkerdelegation in code; the rest are generated so Phase 2 can
# adopt them file-by-file without re-running the toolchain setup.
CONTRACTS=(
	BunkerToken
	BunkerStaking
	BunkerEscrow
	BunkerPricing
	BunkerDelegation
	BunkerReputation
	BunkerVerification
	BunkerRegistry
)

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

for name in "${CONTRACTS[@]}"; do
	artifact="${CONTRACTS_DIR}/out/${name}.sol/${name}.json"
	lower="$(echo "${name}" | tr '[:upper:]' '[:lower:]')"
	abi_file="${TMP_DIR}/${lower}.abi"
	out_file="${BINDINGS_DIR}/${lower}.go"

	[ -f "${artifact}" ] || die "artifact not found: ${artifact} (did 'forge build' succeed?)"

	# Extract the .abi array. Forge writes it as a JSON array under the top-level
	# "abi" key; emit it compactly so abigen reads a clean ABI document.
	jq -c '.abi' "${artifact}" > "${abi_file}"
	[ -s "${abi_file}" ] || die "empty ABI extracted for ${name}"

	echo ">> abigen ${name} -> internal/payment/bindings/${lower}.go"
	abigen --abi "${abi_file}" --pkg "${PKG}" --type "${name}" --out "${out_file}"
	# abigen writes 0600; normalize to 0644 so the committed source mode is stable
	# and `make bindings-check` does not flag a mode-only diff.
	chmod 0644 "${out_file}"
done

echo ">> done. Generated ${#CONTRACTS[@]} bindings in internal/payment/bindings/"
