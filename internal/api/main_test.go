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
		// 99designs/keyring's KWallet backend opens a DBus connection at
		// package init (kwallet.go); the resulting inWorker goroutine runs
		// for the lifetime of the process and has no exposed shutdown.
		goleak.IgnoreAnyFunction("github.com/godbus/dbus.(*Conn).inWorker"),
	)
}
