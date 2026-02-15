package daemon

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// p2p.Router starts a background cleanup goroutine via util.SafeGoWithName
		// that may still be shutting down when goleak checks after test completion.
		goleak.IgnoreAnyFunction("github.com/moltbunker/moltbunker/internal/util.SafeGoWithName.func1"),
		// Test API server spawns goroutines for connection handling
		// that may not fully drain before test cleanup.
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
	)
}
