package api

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// WalletAuthManager starts a background cleanup goroutine (cleanupLoop)
		// that runs a ticker forever and has no shutdown mechanism.
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreTopFunction("runtime.gopark"),
		goleak.IgnoreAnyFunction("github.com/moltbunker/moltbunker/internal/api.(*WalletAuthManager).cleanupLoop"),
		goleak.IgnoreAnyFunction("github.com/moltbunker/moltbunker/internal/api.(*ExecSessionManager).reapLoop"),
	)
}
