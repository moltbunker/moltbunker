package identity

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// geth's keystore.NewKeyStore() spawns fsnotify watchers that have
		// no Close() method, so they leak in tests. This is a known upstream issue.
		// IgnoreAnyFunction is needed because the goroutine is blocked in
		// syscall.syscall6 (kqueue), so readEvents isn't the top frame.
		goleak.IgnoreTopFunction("github.com/ethereum/go-ethereum/accounts/keystore.(*watcher).loop"),
		goleak.IgnoreAnyFunction("github.com/fsnotify/fsnotify.(*Watcher).readEvents"),
	)
}
