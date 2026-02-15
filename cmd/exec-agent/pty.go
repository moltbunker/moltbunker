//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

// PTY wraps a pseudo-terminal master/slave pair with a spawned shell process.
type PTY struct {
	master *os.File
	slave  *os.File
	cmd    *exec.Cmd
}

// openPTY allocates a new PTY pair using /dev/ptmx (Linux).
// This avoids external dependencies and works with CGO_ENABLED=0.
func openPTY() (master, slave *os.File, err error) {
	master, err = os.OpenFile("/dev/ptmx", os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("open /dev/ptmx: %w", err)
	}

	// Unlock the slave side
	var unlock int
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&unlock))); errno != 0 {
		master.Close()
		return nil, nil, fmt.Errorf("TIOCSPTLCK: %w", errno)
	}

	// Get the slave device number
	var ptyno uint32
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&ptyno))); errno != 0 {
		master.Close()
		return nil, nil, fmt.Errorf("TIOCGPTN: %w", errno)
	}

	slavePath := fmt.Sprintf("/dev/pts/%d", ptyno)
	slave, err = os.OpenFile(slavePath, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("open slave %s: %w", slavePath, err)
	}

	return master, slave, nil
}

// SpawnShell allocates a PTY, starts a shell process, and returns the PTY handle.
// The shell runs as a login shell (/bin/sh) with a new session.
func SpawnShell(cols, rows uint16) (*PTY, error) {
	master, slave, err := openPTY()
	if err != nil {
		return nil, fmt.Errorf("allocate PTY: %w", err)
	}

	// Set initial terminal size
	if err := setWinsize(master, cols, rows); err != nil {
		master.Close()
		slave.Close()
		return nil, fmt.Errorf("set initial winsize: %w", err)
	}

	// Determine shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	cmd.Stdin = slave
	cmd.Stdout = slave
	cmd.Stderr = slave
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
		Ctty:    0, // slave fd will be stdin (fd 0) after exec
	}

	if err := cmd.Start(); err != nil {
		master.Close()
		slave.Close()
		return nil, fmt.Errorf("start shell: %w", err)
	}

	// Close slave on our side — the child process owns it now
	slave.Close()

	return &PTY{
		master: master,
		slave:  nil, // closed
		cmd:    cmd,
	}, nil
}

// Read reads from the PTY master (shell output).
func (p *PTY) Read(buf []byte) (int, error) {
	return p.master.Read(buf)
}

// Write writes to the PTY master (shell input).
func (p *PTY) Write(data []byte) (int, error) {
	return p.master.Write(data)
}

// Resize changes the PTY window size (TIOCSWINSZ).
func (p *PTY) Resize(cols, rows uint16) error {
	return setWinsize(p.master, cols, rows)
}

// Close terminates the shell and closes the PTY.
func (p *PTY) Close() error {
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Signal(syscall.SIGHUP)
		_ = p.cmd.Wait()
	}
	return p.master.Close()
}

// Wait waits for the shell process to exit and returns the error.
func (p *PTY) Wait() error {
	if p.cmd == nil {
		return nil
	}
	return p.cmd.Wait()
}

// winsize matches the kernel struct winsize for TIOCSWINSZ/TIOCGWINSZ.
type winsize struct {
	Rows uint16
	Cols uint16
	X    uint16 // unused
	Y    uint16 // unused
}

func setWinsize(f *os.File, cols, rows uint16) error {
	ws := winsize{Rows: rows, Cols: cols}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&ws)),
	)
	if errno != 0 {
		return fmt.Errorf("TIOCSWINSZ: %w", errno)
	}
	return nil
}
