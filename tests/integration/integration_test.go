//go:build integration

// Package integration provides full-stack integration tests that wire together:
//   - Anvil (local Ethereum) with deployed smart contracts
//   - Real P2P nodes with mDNS discovery
//   - Real containerd via Colima (optional, macOS)
//
// Prerequisites:
//   - Foundry installed (~/.foundry/bin/anvil, forge, cast)
//   - Optionally: Colima running for real container tests
//
// Run with:
//
//	go test -tags integration -v -timeout 10m ./tests/integration/...
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ─── Package-level state shared across tests ────────────────────────────────

var (
	anvilPort    int
	anvilRPCURL  string
	anvilProcess *exec.Cmd

	// Contract addresses deployed by DeployLocal.s.sol
	tokenAddr   string
	stakingAddr string
	escrowAddr  string
	pricingAddr string

	// Anvil default accounts (deterministic)
	deployerPK  = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	deployerAddr = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	treasuryAddr = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
	operatorAddr = "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"
	operatorPK   = "0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"
	slasherAddr  = "0x90F79bf6EB2c4f870365E785982E1f101E93b906"
	provider1Addr = "0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65"
	provider1PK   = "0x47e179ec197488593b187f80a00eb0da91f1b9d0b13f8733639f19c30a34926a"
	provider2Addr = "0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc"
	provider2PK   = "0x8b3a350cf5c34c9194ca85829a2df0ec3153be0318b5e2d3348e872092edffba"
	provider3Addr = "0x976EA74026E726554dB657fA54763abd0C3a0aa9"
	provider3PK   = "0x92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e"
	requesterAddr = "0x14dC79964da2C08b23698B3D3cc7Ca32193d9955"
	requesterPK   = "0x4bbbf85ce3377467afe5d46f804f221813b2bb87f24d81f60f1fcdbf7cbf4356"

	// Tool paths
	foundryBin string

	// Feature flags
	hasColima bool
)

func TestMain(m *testing.M) {
	// 1. Find Foundry tools
	foundryBin = filepath.Join(os.Getenv("HOME"), ".foundry", "bin")
	if _, err := os.Stat(filepath.Join(foundryBin, "anvil")); os.IsNotExist(err) {
		fmt.Println("SKIP: Foundry not installed (~/.foundry/bin/anvil not found)")
		fmt.Println("Install with: curl -L https://foundry.paradigm.xyz | bash && foundryup")
		os.Exit(0)
	}

	// 2. Check Colima (optional)
	hasColima = checkColima()

	// 3. Start Anvil
	var err error
	anvilPort, err = findFreePort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: cannot find free port: %v\n", err)
		os.Exit(1)
	}
	anvilRPCURL = fmt.Sprintf("http://127.0.0.1:%d", anvilPort)

	if err := startAnvil(); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: cannot start Anvil: %v\n", err)
		os.Exit(1)
	}

	// 4. Deploy contracts
	if err := deployContracts(); err != nil {
		stopAnvil()
		fmt.Fprintf(os.Stderr, "FATAL: cannot deploy contracts: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n=== Integration Test Environment ===\n")
	fmt.Printf("Anvil RPC:  %s\n", anvilRPCURL)
	fmt.Printf("Token:      %s\n", tokenAddr)
	fmt.Printf("Staking:    %s\n", stakingAddr)
	fmt.Printf("Escrow:     %s\n", escrowAddr)
	fmt.Printf("Pricing:    %s\n", pricingAddr)
	fmt.Printf("Colima:     %v\n", hasColima)
	fmt.Printf("====================================\n\n")

	// 5. Run tests
	code := m.Run()

	// 6. Cleanup
	stopAnvil()
	os.Exit(code)
}

// ─── Anvil management ───────────────────────────────────────────────────────

func startAnvil() error {
	anvilPath := filepath.Join(foundryBin, "anvil")
	anvilProcess = exec.Command(anvilPath,
		"--port", strconv.Itoa(anvilPort),
		"--accounts", "10",
		"--balance", "10000",
		"--silent",
	)
	anvilProcess.Stdout = nil
	anvilProcess.Stderr = nil

	if err := anvilProcess.Start(); err != nil {
		return fmt.Errorf("failed to start anvil: %w", err)
	}

	// Wait for Anvil to be ready
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", anvilPort), 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("anvil did not start within 10s on port %d", anvilPort)
}

func stopAnvil() {
	if anvilProcess != nil && anvilProcess.Process != nil {
		anvilProcess.Process.Kill()
		anvilProcess.Wait()
	}
}

// ─── Contract deployment ────────────────────────────────────────────────────

func deployContracts() error {
	forgePath := filepath.Join(foundryBin, "forge")
	contractsDir := filepath.Join(projectRoot(), "contracts")

	cmd := exec.Command(forgePath, "script",
		"script/DeployLocal.s.sol",
		"--rpc-url", anvilRPCURL,
		"--broadcast",
	)
	cmd.Dir = contractsDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("forge script failed: %w\nOutput:\n%s", err, string(output))
	}

	// Parse addresses from broadcast JSON
	broadcastFile := filepath.Join(contractsDir, "broadcast", "DeployLocal.s.sol", "31337", "run-latest.json")
	return parseContractAddresses(broadcastFile)
}

func parseContractAddresses(broadcastFile string) error {
	data, err := os.ReadFile(broadcastFile)
	if err != nil {
		return fmt.Errorf("cannot read broadcast file: %w", err)
	}

	var broadcast struct {
		Transactions []struct {
			Hash            string `json:"hash"`
			TransactionType string `json:"transactionType"`
			ContractAddress string `json:"contractAddress"`
			ContractName    string `json:"contractName"`
		} `json:"transactions"`
	}

	if err := json.Unmarshal(data, &broadcast); err != nil {
		return fmt.Errorf("cannot parse broadcast JSON: %w", err)
	}

	for _, tx := range broadcast.Transactions {
		if tx.TransactionType != "CREATE" {
			continue
		}
		switch tx.ContractName {
		case "BunkerToken":
			tokenAddr = tx.ContractAddress
		case "BunkerStaking":
			stakingAddr = tx.ContractAddress
		case "BunkerEscrow":
			escrowAddr = tx.ContractAddress
		case "BunkerPricing":
			pricingAddr = tx.ContractAddress
		}
	}

	if tokenAddr == "" || stakingAddr == "" || escrowAddr == "" || pricingAddr == "" {
		return fmt.Errorf("missing contract addresses: token=%s staking=%s escrow=%s pricing=%s",
			tokenAddr, stakingAddr, escrowAddr, pricingAddr)
	}

	return nil
}

// ─── cast wrapper ───────────────────────────────────────────────────────────

// castCall executes a read-only contract call and returns the raw output.
func castCall(contract, sig string, args ...string) (string, error) {
	castPath := filepath.Join(foundryBin, "cast")
	cmdArgs := []string{"call", contract, sig}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--rpc-url", anvilRPCURL)

	cmd := exec.Command(castPath, cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cast call failed: %w\nOutput: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// castSend executes a state-changing transaction and waits for the receipt.
func castSend(privateKey, contract, sig string, args ...string) (string, error) {
	castPath := filepath.Join(foundryBin, "cast")
	cmdArgs := []string{"send", contract, sig}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, "--rpc-url", anvilRPCURL, "--private-key", privateKey)

	cmd := exec.Command(castPath, cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cast send failed: %w\nOutput: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// castDecode decodes a raw hex value to a specific type.
func castDecode(typ, value string) (string, error) {
	castPath := filepath.Join(foundryBin, "cast")
	cmd := exec.Command(castPath, "--to-base", value, "10")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cast decode failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ─── Utilities ──────────────────────────────────────────────────────────────

func projectRoot() string {
	// Walk up from the test file to find go.mod
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

func findFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func checkColima() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	sock := filepath.Join(home, ".colima", "default", "containerd.sock")
	conn, err := net.DialTimeout("unix", sock, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func ctxWithTimeout(t *testing.T, d time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), d)
	t.Cleanup(cancel)
	return ctx
}
