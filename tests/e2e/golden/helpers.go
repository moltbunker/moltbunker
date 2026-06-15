// This file is intentionally NOT behind the e2e build tag. The pure helpers
// below are shared by both the e2e-tagged golden suite and the untagged sanity
// test (golden_sanity_test.go), so there is a single implementation and no risk
// of the two copies silently drifting. Keep this file dependency-free (stdlib
// math/big only) so it compiles in the default, no-tag build on every platform.

package golden

import "math/big"

// BunkerToWei converts a whole-number BUNKER amount to wei (18 decimals). Same
// pattern as escrow_test.go's bunkerToWei helper.
func BunkerToWei(n int64) *big.Int {
	decimals := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	return new(big.Int).Mul(big.NewInt(n), decimals)
}

// bareID strips the conventional "dep-" prefix so the 8-char hex prefix used for
// auto-assigned subdomains can be taken from the start of the hex id.
func bareID(deploymentID string) string {
	const p = "dep-"
	if len(deploymentID) > len(p) && deploymentID[:len(p)] == p {
		return deploymentID[len(p):]
	}
	return deploymentID
}

// contains is strings.Contains without importing strings (kept tiny so the
// no-tag build pulls in nothing beyond math/big).
func contains(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
