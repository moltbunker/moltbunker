package p2p

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// Router starts a background cleanup goroutine via util.SafeGoWithName
		// that may still be shutting down when goleak checks after test completion.
		goleak.IgnoreAnyFunction("github.com/moltbunker/moltbunker/internal/util.SafeGoWithName.func1"),
		// HTTP transport goroutines from GeoLocator's http.Client connection pool.
		goleak.IgnoreTopFunction("net/http.(*persistConn).readLoop"),
		goleak.IgnoreTopFunction("net/http.(*persistConn).writeLoop"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
	)
}
