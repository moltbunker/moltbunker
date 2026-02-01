package commands

import (
	"testing"
)

func TestNewInstallCmd(t *testing.T) {
	cmd := NewInstallCmd()

	if cmd == nil {
		t.Fatal("NewInstallCmd returned nil")
	}

	if cmd.Use != "install" {
		t.Errorf("Use mismatch: got %s, want install", cmd.Use)
	}
}

func TestNewStartCmd(t *testing.T) {
	cmd := NewStartCmd()

	if cmd == nil {
		t.Fatal("NewStartCmd returned nil")
	}

	if cmd.Use != "start" {
		t.Errorf("Use mismatch: got %s, want start", cmd.Use)
	}
}

func TestNewStopCmd(t *testing.T) {
	cmd := NewStopCmd()

	if cmd == nil {
		t.Fatal("NewStopCmd returned nil")
	}

	if cmd.Use != "stop" {
		t.Errorf("Use mismatch: got %s, want stop", cmd.Use)
	}
}

func TestNewStatusCmd(t *testing.T) {
	cmd := NewStatusCmd()

	if cmd == nil {
		t.Fatal("NewStatusCmd returned nil")
	}

	if cmd.Use != "status" {
		t.Errorf("Use mismatch: got %s, want status", cmd.Use)
	}
}

func TestNewDeployCmd(t *testing.T) {
	cmd := NewDeployCmd()

	if cmd == nil {
		t.Fatal("NewDeployCmd returned nil")
	}

	if cmd.Use != "deploy [image]" {
		t.Errorf("Use mismatch: got %s, want deploy [image]", cmd.Use)
	}

	// Check flags exist
	torOnlyFlag := cmd.Flags().Lookup("tor-only")
	if torOnlyFlag == nil {
		t.Error("--tor-only flag should exist")
	}

	onionServiceFlag := cmd.Flags().Lookup("onion-service")
	if onionServiceFlag == nil {
		t.Error("--onion-service flag should exist")
	}
}

func TestNewLogsCmd(t *testing.T) {
	cmd := NewLogsCmd()

	if cmd == nil {
		t.Fatal("NewLogsCmd returned nil")
	}

	if cmd.Use != "logs [container-id]" {
		t.Errorf("Use mismatch: got %s, want logs [container-id]", cmd.Use)
	}

	// Check flags exist
	followFlag := cmd.Flags().Lookup("follow")
	if followFlag == nil {
		t.Error("--follow flag should exist")
	}

	tailFlag := cmd.Flags().Lookup("tail")
	if tailFlag == nil {
		t.Error("--tail flag should exist")
	}
}

func TestNewMonitorCmd(t *testing.T) {
	cmd := NewMonitorCmd()

	if cmd == nil {
		t.Fatal("NewMonitorCmd returned nil")
	}

	if cmd.Use != "monitor" {
		t.Errorf("Use mismatch: got %s, want monitor", cmd.Use)
	}
}

func TestNewConfigCmd(t *testing.T) {
	cmd := NewConfigCmd()

	if cmd == nil {
		t.Fatal("NewConfigCmd returned nil")
	}

	if cmd.Use != "config" {
		t.Errorf("Use mismatch: got %s, want config", cmd.Use)
	}

	// Check subcommands
	if !cmd.HasSubCommands() {
		t.Error("config should have subcommands")
	}
}

func TestNewInteractiveCmd(t *testing.T) {
	cmd := NewInteractiveCmd()

	if cmd == nil {
		t.Fatal("NewInteractiveCmd returned nil")
	}

	if cmd.Use != "interactive" {
		t.Errorf("Use mismatch: got %s, want interactive", cmd.Use)
	}
}

func TestNewTorCmd(t *testing.T) {
	cmd := NewTorCmd()

	if cmd == nil {
		t.Fatal("NewTorCmd returned nil")
	}

	if cmd.Use != "tor" {
		t.Errorf("Use mismatch: got %s, want tor", cmd.Use)
	}

	// Check subcommands
	if !cmd.HasSubCommands() {
		t.Error("tor should have subcommands")
	}

	// Verify subcommand names
	subCommands := cmd.Commands()
	expectedSubCmds := map[string]bool{
		"start":  false,
		"status": false,
		"onion":  false,
		"rotate": false,
	}

	for _, subCmd := range subCommands {
		if _, exists := expectedSubCmds[subCmd.Name()]; exists {
			expectedSubCmds[subCmd.Name()] = true
		}
	}

	for name, found := range expectedSubCmds {
		if !found {
			t.Errorf("Missing tor subcommand: %s", name)
		}
	}
}

func TestTruncateID(t *testing.T) {
	tests := []struct {
		id     string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactlength", 11, "exactlength"},
		{"thisisaverylongid", 10, "thisisa..."}, // 7 chars + "..." = 10
		{"", 10, ""},
	}

	for _, tt := range tests {
		got := truncateID(tt.id, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateID(%s, %d) = %s, want %s", tt.id, tt.maxLen, got, tt.want)
		}
	}
}
