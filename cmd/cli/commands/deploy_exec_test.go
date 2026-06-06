package commands

import (
	"errors"
	"testing"
)

// stubFetcher implements execPubKeyFetcher for tests.
type stubFetcher struct {
	pub []byte
	err error
}

func (s stubFetcher) GetExecPubKey() ([]byte, error) { return s.pub, s.err }

// When the provider pubkey cannot be fetched (older daemon / key load failure),
// generateSealedExecKey must fail BEFORE touching the wallet so the deploy
// silently falls back to no-E2E rather than sending a cleartext key.
func TestGenerateSealedExecKey_NoProviderPubKey(t *testing.T) {
	_, _, err := generateSealedExecKey(stubFetcher{err: errors.New("unavailable")})
	if err == nil {
		t.Fatal("expected error when provider pubkey is unavailable")
	}
}
