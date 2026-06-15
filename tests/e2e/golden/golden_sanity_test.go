package golden

import (
	"math/big"
	"testing"
)

// This sanity test is intentionally NOT gated behind the e2e build tag, so the
// default `go test ./tests/e2e/golden/...` (no tags) actually executes a test
// on every platform. It covers the pure helpers used by the e2e-tagged golden
// path so a regression in them is caught even without the e2e tag. The helpers
// themselves live in the untagged helpers.go and are the SAME implementations
// the e2e suite uses — there is no separate sanity copy to drift. The full
// acceptance flow lives in the `//go:build e2e` files and runs in the CI
// `e2e-golden` job.

func TestSanity_BunkerToWei(t *testing.T) {
	got := BunkerToWei(100)
	want := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	if got.Cmp(want) != 0 {
		t.Fatalf("BunkerToWei(100) = %s, want %s", got, want)
	}
}

func TestSanity_BareIDStripsPrefix(t *testing.T) {
	if got := bareID("dep-deadbeef"); got != "deadbeef" {
		t.Fatalf("bareID(dep-deadbeef) = %q, want deadbeef", got)
	}
	if got := bareID("noprefix"); got != "noprefix" {
		t.Fatalf("bareID(noprefix) = %q, want noprefix", got)
	}
}

func TestSanity_Contains(t *testing.T) {
	if !contains("169.254.169.254/32 drop", "169.254.169.254/32") {
		t.Fatal("contains should find the cidr substring")
	}
	if contains("accept", "drop") {
		t.Fatal("contains should not match a missing substring")
	}
}
