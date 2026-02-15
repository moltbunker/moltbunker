package commands

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

// makeTerminalRaw puts stdin into raw mode and returns a restore function.
// The caller MUST defer the restore function to avoid leaving the terminal in a broken state.
func makeTerminalRaw() (restore func(), err error) {
	fd := int(syscall.Stdin)
	if !term.IsTerminal(fd) {
		return func() {}, fmt.Errorf("stdin is not a terminal")
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return func() {}, fmt.Errorf("failed to set raw mode: %w", err)
	}

	return func() {
		term.Restore(fd, oldState)
	}, nil
}

// terminalSize returns the current terminal dimensions (cols, rows).
// Falls back to 80x24 if detection fails.
func terminalSize() (cols, rows uint16) {
	fd := int(os.Stdout.Fd())
	w, h, err := term.GetSize(fd)
	if err != nil || w <= 0 || h <= 0 {
		return 80, 24
	}
	return uint16(w), uint16(h)
}
