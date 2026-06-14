package payment

// Contract bindings in internal/payment/bindings/ are generated from the
// canonical forge artifacts in contracts/out/ by scripts/gen-bindings.sh.
//
// Run `go generate ./internal/payment/...` (or `make gen-bindings`) after the
// Solidity contracts change. The generated files are committed, so a normal
// `go build` needs only the Go toolchain — the generation toolchain (forge, jq,
// abigen) is required only when regenerating. `make bindings-check` detects
// drift between the committed bindings and a fresh generation.

//go:generate ../../scripts/gen-bindings.sh
