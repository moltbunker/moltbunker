package commands

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/spf13/cobra"
)

// Requester command flags
var (
	requesterOutputDir  string
	requesterFollow     bool
	requesterDecrypt    bool
	requesterKeyFile    string
)

// NewRequesterCmd creates the requester command with subcommands
func NewRequesterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "requester",
		Short: "Requester operations (jobs, outputs, keys)",
		Long: `Commands for requesters who submit jobs to the Moltbunker network.

Requesters submit containerized workloads to the network and pay BUNKER tokens.
All job outputs are encrypted and only accessible to the requester.

Examples:
  moltbunker requester keys init    # Initialize encryption keys
  moltbunker requester jobs         # List submitted jobs
  moltbunker requester outputs      # Retrieve encrypted outputs`,
	}

	cmd.AddCommand(newRequesterKeysCmd())
	cmd.AddCommand(newRequesterJobsCmd())
	cmd.AddCommand(newRequesterOutputsCmd())
	cmd.AddCommand(newRequesterBalanceCmd())

	return cmd
}

// newRequesterKeysCmd creates the keys subcommand
func newRequesterKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage requester encryption keys",
		Long: `Manage encryption keys for secure job output retrieval.

All job outputs are encrypted with your public key and can only be
decrypted with your private key. Keep your private key safe!

Subcommands:
  keys init      - Generate a new key pair
  keys show      - Display your public key
  keys export    - Export public key to file
  keys delete    - Delete your keys (WARNING: irreversible)`,
	}

	cmd.AddCommand(newRequesterKeysInitCmd())
	cmd.AddCommand(newRequesterKeysShowCmd())
	cmd.AddCommand(newRequesterKeysExportCmd())
	cmd.AddCommand(newRequesterKeysDeleteCmd())

	return cmd
}

func newRequesterKeysInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate new encryption keys",
		Long: `Generate a new X25519 key pair for encrypting job outputs.

You will be prompted to create a passphrase to protect your private key.
This passphrase is required whenever you decrypt job outputs.

WARNING: If you lose your passphrase, you cannot recover your keys!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyStorePath := getKeyStorePath()
			rkm := security.NewRequesterKeyManager(keyStorePath)

			if rkm.KeysExist() {
				return fmt.Errorf("keys already exist at %s. Use 'keys delete' first to replace them", keyStorePath)
			}

			// Prompt for passphrase
			fmt.Println("Generate Encryption Keys")
			fmt.Println("========================")
			fmt.Println()
			fmt.Print("Enter passphrase to protect your private key: ")
			passphrase, err := readPassword()
			if err != nil {
				return fmt.Errorf("failed to read passphrase: %w", err)
			}

			fmt.Print("Confirm passphrase: ")
			confirm, err := readPassword()
			if err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}

			if passphrase != confirm {
				return fmt.Errorf("passphrases do not match")
			}

			if len(passphrase) < 8 {
				return fmt.Errorf("passphrase must be at least 8 characters")
			}

			fmt.Println()
			fmt.Println("Generating keys...")

			if err := rkm.GenerateNewKeys(); err != nil {
				return fmt.Errorf("failed to generate keys: %w", err)
			}

			if err := rkm.SaveKeys(passphrase); err != nil {
				return fmt.Errorf("failed to save keys: %w", err)
			}

			pubKeyHex, _ := rkm.PublicKeyHex()

			fmt.Println()
			fmt.Println("Keys Generated Successfully")
			fmt.Println("===========================")
			fmt.Printf("Key File:   %s\n", keyStorePath)
			fmt.Printf("Public Key: %s\n", pubKeyHex)
			fmt.Println()
			fmt.Println("Your public key will be sent to providers when you submit jobs.")
			fmt.Println("Keep your passphrase safe - it's required to decrypt job outputs!")

			return nil
		},
	}
}

func newRequesterKeysShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display your public key",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyStorePath := getKeyStorePath()
			rkm := security.NewRequesterKeyManager(keyStorePath)

			if !rkm.KeysExist() {
				return fmt.Errorf("no keys found. Run 'moltbunker requester keys init' first")
			}

			fmt.Print("Enter passphrase: ")
			passphrase, err := readPassword()
			if err != nil {
				return err
			}
			fmt.Println()

			if err := rkm.LoadKeys(passphrase); err != nil {
				return fmt.Errorf("failed to load keys: %w", err)
			}

			pubKeyHex, _ := rkm.PublicKeyHex()

			fmt.Println("Public Key")
			fmt.Println("==========")
			fmt.Println(pubKeyHex)
			fmt.Println()
			fmt.Println("Share this key with the network when submitting jobs.")

			return nil
		},
	}
}

func newRequesterKeysExportCmd() *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export public key to file",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyStorePath := getKeyStorePath()
			rkm := security.NewRequesterKeyManager(keyStorePath)

			if !rkm.KeysExist() {
				return fmt.Errorf("no keys found")
			}

			fmt.Print("Enter passphrase: ")
			passphrase, err := readPassword()
			if err != nil {
				return err
			}
			fmt.Println()

			if err := rkm.LoadKeys(passphrase); err != nil {
				return fmt.Errorf("failed to load keys: %w", err)
			}

			pubKey, _ := rkm.GetPublicKey()

			if outputFile == "" {
				outputFile = "moltbunker_pubkey.bin"
			}

			if err := os.WriteFile(outputFile, pubKey, 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			fmt.Printf("Public key exported to: %s\n", outputFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: moltbunker_pubkey.bin)")

	return cmd
}

func newRequesterKeysDeleteCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete your encryption keys (WARNING: irreversible)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				fmt.Println("WARNING: This will permanently delete your encryption keys!")
				fmt.Println("You will NOT be able to decrypt any job outputs encrypted with these keys.")
				fmt.Println()
				fmt.Println("To proceed, run: moltbunker requester keys delete --confirm")
				return nil
			}

			keyStorePath := getKeyStorePath()
			rkm := security.NewRequesterKeyManager(keyStorePath)

			if !rkm.KeysExist() {
				return fmt.Errorf("no keys found")
			}

			if err := rkm.DeleteKeys(); err != nil {
				return fmt.Errorf("failed to delete keys: %w", err)
			}

			fmt.Println("Keys deleted successfully.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm key deletion")

	return cmd
}

// newRequesterJobsCmd creates the jobs subcommand
func newRequesterJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "List and manage submitted jobs",
		Long: `List jobs you have submitted to the Moltbunker network.

Shows:
- Active jobs currently running
- Completed jobs with status
- Job costs and duration

Examples:
  moltbunker requester jobs
  moltbunker requester jobs --status running`,
		RunE: runRequesterJobs,
	}

	return cmd
}

func runRequesterJobs(cmd *cobra.Command, args []string) error {
	daemonClient := client.NewDaemonClient(SocketPath)
	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running")
	}
	defer daemonClient.Close()

	resp, err := daemonClient.RequesterJobs()
	if err != nil {
		return fmt.Errorf("failed to get jobs: %w", err)
	}

	if len(resp.Jobs) == 0 {
		fmt.Println("No jobs found.")
		fmt.Println()
		fmt.Println("Submit a job with: moltbunker deploy <image>")
		return nil
	}

	fmt.Println("Your Jobs")
	fmt.Println("=========")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tImage\tStatus\tDuration\tCost\tReplicas")
	fmt.Fprintln(w, "--\t-----\t------\t--------\t----\t--------")
	for _, j := range resp.Jobs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s BUNKER\t%d/3\n",
			j.ID[:12], j.Image, j.Status, j.Duration, j.Cost, j.ReplicaCount)
	}
	w.Flush()

	return nil
}

// newRequesterOutputsCmd creates the outputs subcommand
func newRequesterOutputsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outputs [job-id]",
		Short: "Retrieve and decrypt job outputs",
		Long: `Retrieve encrypted outputs from your jobs and decrypt them.

Outputs are encrypted with your public key and can only be decrypted
with your private key. You will be prompted for your passphrase.

Examples:
  moltbunker requester outputs abc123        # Get outputs for job
  moltbunker requester outputs abc123 --save # Save to files`,
		Args: cobra.ExactArgs(1),
		RunE: runRequesterOutputs,
	}

	cmd.Flags().StringVarP(&requesterOutputDir, "output", "o", "", "Directory to save outputs")
	cmd.Flags().BoolVar(&requesterDecrypt, "decrypt", true, "Decrypt outputs (default: true)")
	cmd.Flags().StringVar(&requesterKeyFile, "key-file", "", "Path to key file (default: ~/.moltbunker/requester_keys.json)")

	return cmd
}

func runRequesterOutputs(cmd *cobra.Command, args []string) error {
	jobID := args[0]

	daemonClient := client.NewDaemonClient(SocketPath)
	if err := daemonClient.Connect(); err != nil {
		return fmt.Errorf("daemon not running")
	}
	defer daemonClient.Close()

	// Get encrypted outputs from daemon
	resp, err := daemonClient.RequesterOutputs(jobID)
	if err != nil {
		return fmt.Errorf("failed to get outputs: %w", err)
	}

	if len(resp.Outputs) == 0 {
		fmt.Println("No outputs available for this job.")
		return nil
	}

	fmt.Printf("Job Outputs for %s\n", jobID)
	fmt.Println("===================")
	fmt.Println()

	if !requesterDecrypt {
		// Just show encrypted output info
		for _, out := range resp.Outputs {
			fmt.Printf("Type: %s, Size: %d bytes, Timestamp: %s\n",
				out.Type, len(out.EncryptedData), out.Timestamp)
		}
		fmt.Println()
		fmt.Println("Outputs are encrypted. Use --decrypt to view contents.")
		return nil
	}

	// Load keys for decryption
	keyStorePath := requesterKeyFile
	if keyStorePath == "" {
		keyStorePath = getKeyStorePath()
	}

	rkm := security.NewRequesterKeyManager(keyStorePath)
	if !rkm.KeysExist() {
		return fmt.Errorf("no encryption keys found. Run 'moltbunker requester keys init' first")
	}

	fmt.Print("Enter passphrase: ")
	passphrase, err := readPassword()
	if err != nil {
		return err
	}
	fmt.Println()

	if err := rkm.LoadKeys(passphrase); err != nil {
		return fmt.Errorf("failed to load keys: %w", err)
	}

	decryptor, err := rkm.CreateDecryptor()
	if err != nil {
		return fmt.Errorf("failed to create decryptor: %w", err)
	}

	// Decrypt and display/save outputs
	for i, out := range resp.Outputs {
		// Parse metadata
		metadata := &security.EncryptionMetadata{
			DeploymentID:   out.DeploymentID,
			ProviderPubKey: out.ProviderPubKey,
			EncryptedDEK:   out.EncryptedDEK,
			DEKNonce:       out.DEKNonce,
		}

		decrypted, err := decryptor.DecryptOutput(metadata, out.EncryptedData)
		if err != nil {
			fmt.Printf("Failed to decrypt output %d: %v\n", i, err)
			continue
		}

		if requesterOutputDir != "" {
			// Save to file
			filename := fmt.Sprintf("%s_%s_%d.txt", jobID[:12], out.Type, i)
			outputPath := filepath.Join(requesterOutputDir, filename)
			if err := os.WriteFile(outputPath, decrypted, 0600); err != nil {
				fmt.Printf("Failed to save output: %v\n", err)
				continue
			}
			fmt.Printf("Saved %s to %s\n", out.Type, outputPath)
		} else {
			// Display
			fmt.Printf("\n--- %s (from %s) ---\n", out.Type, out.Timestamp)
			fmt.Println(string(decrypted))
		}
	}

	return nil
}

// newRequesterBalanceCmd creates the balance subcommand
func newRequesterBalanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "Show BUNKER token balance",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemonClient := client.NewDaemonClient(SocketPath)
			if err := daemonClient.Connect(); err != nil {
				return fmt.Errorf("daemon not running")
			}
			defer daemonClient.Close()

			resp, err := daemonClient.RequesterBalance()
			if err != nil {
				return fmt.Errorf("failed to get balance: %w", err)
			}

			fmt.Println("BUNKER Balance")
			fmt.Println("==============")
			fmt.Printf("Available:   %s BUNKER\n", resp.Available)
			fmt.Printf("Reserved:    %s BUNKER (for active jobs)\n", resp.Reserved)
			fmt.Printf("Total Spent: %s BUNKER\n", resp.TotalSpent)

			return nil
		},
	}
}

// Helper functions

func getKeyStorePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/moltbunker_keys.json"
	}
	return filepath.Join(homeDir, ".moltbunker", "requester_keys.json")
}

func readPassword() (string, error) {
	// Simple implementation - in production use golang.org/x/term
	var password string
	_, err := fmt.Scanln(&password)
	return password, err
}

// Additional deploy enhancements for requester flow

// RequesterDeployFlags contains deployment flags for requesters
var (
	requesterPubKeyHex string
)

// EnhanceDeployCmd adds requester-specific flags to the deploy command
func EnhanceDeployCmd(cmd *cobra.Command) {
	cmd.Flags().StringVar(&requesterPubKeyHex, "pub-key", "", "Requester public key for output encryption (hex)")
}

// GetRequesterPubKey returns the requester's public key for deployment
func GetRequesterPubKey() ([]byte, error) {
	if requesterPubKeyHex != "" {
		return hex.DecodeString(requesterPubKeyHex)
	}

	// Try to load from key file
	keyStorePath := getKeyStorePath()
	rkm := security.NewRequesterKeyManager(keyStorePath)

	if !rkm.KeysExist() {
		return nil, fmt.Errorf("no encryption keys found. Run 'moltbunker requester keys init' or provide --pub-key")
	}

	fmt.Print("Enter passphrase to unlock encryption keys: ")
	passphrase, err := readPassword()
	if err != nil {
		return nil, err
	}
	fmt.Println()

	if err := rkm.LoadKeys(passphrase); err != nil {
		return nil, fmt.Errorf("failed to load keys: %w", err)
	}

	return rkm.GetPublicKey()
}
