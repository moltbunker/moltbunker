// Package golden hosts the E2E-01 golden-path acceptance test: a single
// in-process flow that exercises every major moltbunker product promise in
// sequence (wallet keygen -> on-chain escrow reserve -> image verify+scan gate
// -> container start -> subdomain resolve -> tunnel -> public HTTPS 200 -> stop
// -> escrow finalize).
//
// The substantive test files all carry the `//go:build e2e` constraint and run
// only under `go test -tags e2e ./tests/e2e/golden/...` (see the CI job
// `e2e-golden`). This always-compiled file exists so the package is non-empty
// in the default (untagged) build, keeping `go test ./tests/e2e/golden/...`
// green on every platform instead of erroring with "matched no packages".
//
// Mock-vs-real boundaries are documented per-leg in harness.go and annotated at
// runtime via t.Log("[MOCK: ...]") / t.Log("[REAL: ...]") in each phase.
package golden
