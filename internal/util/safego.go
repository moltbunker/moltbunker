package util

import (
	"runtime/debug"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// SafeGo wraps a goroutine function with panic recovery and logging.
// Use this in place of bare `go` statements to ensure panics are
// caught, logged with stack traces, and don't crash the entire process.
//
// Example:
//
//	util.SafeGo(func() {
//	    // goroutine code here
//	})
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				logging.Error("goroutine panic recovered",
					"panic", r,
					"stack", string(stack),
				)
			}
		}()
		fn()
	}()
}

// SafeGoWithName wraps a goroutine function with panic recovery and logging,
// including a descriptive name for the goroutine for better debugging.
//
// Example:
//
//	util.SafeGoWithName("connection-handler", func() {
//	    // goroutine code here
//	})
func SafeGoWithName(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				logging.Error("goroutine panic recovered",
					"goroutine", name,
					"panic", r,
					"stack", string(stack),
				)
			}
		}()
		fn()
	}()
}
