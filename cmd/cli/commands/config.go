package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long: `Manage moltbunker configuration.

Subcommands:
  get [key]       - Get a configuration value
  set [key] [val] - Set a configuration value
  show            - Show all configuration
  edit            - Open config file in editor`,
	}

	cmd.AddCommand(NewConfigGetCmd())
	cmd.AddCommand(NewConfigSetCmd())
	cmd.AddCommand(NewConfigShowCmd())
	cmd.AddCommand(NewConfigEditCmd())

	return cmd
}

func NewConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			// First try to get from running daemon
			daemonClient := client.NewDaemonClient("")
			if err := daemonClient.Connect(); err == nil {
				defer daemonClient.Close()
				value, err := daemonClient.ConfigGet(key)
				if err == nil {
					fmt.Printf("%s = %v\n", key, value)
					return nil
				}
			}

			// Fall back to config file
			config, err := loadConfigFile()
			if err != nil {
				return err
			}

			value, found := getNestedValue(config, key)
			if !found {
				return fmt.Errorf("key not found: %s", key)
			}

			fmt.Printf("%s = %v\n", key, value)
			return nil
		},
	}
}

func NewConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			config, err := loadConfigFile()
			if err != nil {
				return err
			}

			setNestedValue(config, key, value)

			if err := saveConfigFile(config); err != nil {
				return err
			}

			fmt.Printf("Set %s = %s\n", key, value)
			fmt.Println("Note: Restart daemon for changes to take effect")
			return nil
		},
	}
}

func NewConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show all configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := getConfigPath()

			data, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read config: %w", err)
			}

			fmt.Printf("Configuration file: %s\n", configPath)
			fmt.Println("---")
			fmt.Println(string(data))

			return nil
		},
	}
}

func NewConfigEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open config file in editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := getConfigPath()

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}

			fmt.Printf("Opening %s with %s...\n", configPath, editor)

			// Use exec to replace current process
			editorCmd := fmt.Sprintf("%s %s", editor, configPath)
			return runCommand(editorCmd)
		},
	}
}

func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".moltbunker", "config.yaml")
}

func loadConfigFile() (map[string]interface{}, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

func saveConfigFile(config map[string]interface{}) error {
	configPath := getConfigPath()

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func getNestedValue(config map[string]interface{}, key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
	current := config

	for i, part := range parts {
		val, exists := current[part]
		if !exists {
			return nil, false
		}

		if i == len(parts)-1 {
			return val, true
		}

		next, ok := val.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current = next
	}

	return nil, false
}

func setNestedValue(config map[string]interface{}, key string, value interface{}) {
	parts := strings.Split(key, ".")
	current := config

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}

		val, exists := current[part]
		if !exists {
			current[part] = make(map[string]interface{})
			val = current[part]
		}

		next, ok := val.(map[string]interface{})
		if !ok {
			current[part] = make(map[string]interface{})
			next = current[part].(map[string]interface{})
		}
		current = next
	}
}

func runCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Use os/exec to run the command
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
