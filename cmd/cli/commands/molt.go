package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/moltbunker/moltbunker/internal/client"
	"github.com/spf13/cobra"
)

// NewMoltCmd creates the Molt (WASM serverless) command group.
func NewMoltCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "molt",
		Aliases: []string{"fn", "function"},
		Short:   "Manage Molt serverless functions",
		Long: `Deploy and manage Molt serverless WASM functions.

Molts are lightweight serverless functions powered by WebAssembly.
They start in milliseconds, use minimal memory, and scale automatically.

Examples:
  moltbunker molt deploy my_func.wasm                    # Deploy a WASM function
  moltbunker molt deploy my_func.wasm --timeout 5000     # 5s timeout
  moltbunker molt list                                   # List all Molts
  moltbunker molt invoke <id> --method POST --data '{}' # Invoke directly
  moltbunker molt stop <id>                              # Stop (keeps cache)
  moltbunker molt delete <id>                            # Remove entirely`,
	}

	cmd.AddCommand(
		newMoltDeployCmd(),
		newMoltListCmd(),
		newMoltGetCmd(),
		newMoltStopCmd(),
		newMoltDeleteCmd(),
		newMoltInvokeCmd(),
	)

	return cmd
}

var (
	moltMemoryLimitMB uint32
	moltTimeoutMs     int
	moltMaxInstances  int
	moltOwner         string
)

func newMoltDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy <wasm-file>",
		Short: "Deploy a WASM function as a Molt",
		Long: `Deploy a WebAssembly module as a serverless Molt function.

The WASM module must be WASI-compatible (compiled with wasm32-wasi target).
It reads a JSON request from stdin and writes a JSON response to stdout.

Examples:
  moltbunker molt deploy target/wasm32-wasi/release/handler.wasm
  moltbunker molt deploy echo.wasm --memory 128 --timeout 5000
  moltbunker molt deploy api.wasm --owner 0xAc1D...`,
		Args: cobra.ExactArgs(1),
		RunE: runMoltDeploy,
	}

	cmd.Flags().Uint32Var(&moltMemoryLimitMB, "memory", 0, "Max memory per instance in MB (default: 256)")
	cmd.Flags().IntVar(&moltTimeoutMs, "timeout", 0, "Max execution time in ms (default: 30000)")
	cmd.Flags().IntVar(&moltMaxInstances, "max-instances", 0, "Max concurrent instances (default: 100)")
	cmd.Flags().StringVar(&moltOwner, "owner", "", "Deployer wallet address")

	return cmd
}

func newMoltListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all deployed Molts",
		Args:    cobra.NoArgs,
		RunE:    runMoltList,
	}
}

func newMoltGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <deployment-id>",
		Short: "Get details for a Molt deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  runMoltGet,
	}
}

func newMoltStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <deployment-id>",
		Short: "Stop a Molt (keeps compiled cache for fast restart)",
		Args:  cobra.ExactArgs(1),
		RunE:  runMoltStop,
	}
}

func newMoltDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <deployment-id>",
		Aliases: []string{"rm"},
		Short:   "Delete a Molt deployment entirely",
		Args:    cobra.ExactArgs(1),
		RunE:    runMoltDelete,
	}
}

var (
	moltInvokeMethod string
	moltInvokePath   string
	moltInvokeData   string
)

func newMoltInvokeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoke <deployment-id>",
		Short: "Invoke a Molt directly (for testing)",
		Long: `Directly invoke a deployed Molt function for testing and debugging.

Examples:
  moltbunker molt invoke <id>                                    # GET /
  moltbunker molt invoke <id> --method POST --data '{"key":"v"}'
  moltbunker molt invoke <id> --path /api/hello`,
		Args: cobra.ExactArgs(1),
		RunE: runMoltInvoke,
	}

	cmd.Flags().StringVar(&moltInvokeMethod, "method", "GET", "HTTP method")
	cmd.Flags().StringVar(&moltInvokePath, "path", "/", "HTTP path")
	cmd.Flags().StringVar(&moltInvokeData, "data", "", "Request body (string)")

	return cmd
}

// ── Handlers ─────────────────────────────────────────────────────────────────

func runMoltDeploy(_ *cobra.Command, args []string) error {
	wasmPath := args[0]

	// Read WASM file
	// #nosec G304 -- wasmPath is the user-provided WASM module path (CLI arg); reading it is the command's purpose
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return fmt.Errorf("failed to read WASM file: %w", err)
	}

	if len(wasmBytes) < 8 {
		return fmt.Errorf("file too small to be a valid WASM module")
	}

	// Validate WASM magic number (\0asm)
	if wasmBytes[0] != 0x00 || wasmBytes[1] != 0x61 || wasmBytes[2] != 0x73 || wasmBytes[3] != 0x6d {
		return fmt.Errorf("file is not a valid WebAssembly module (bad magic number)")
	}

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	Info(fmt.Sprintf("Deploying Molt: %s (%d bytes)", wasmPath, len(wasmBytes)))

	req := &client.MoltDeployRequest{
		WasmBytes:     wasmBytes,
		MemoryLimitMB: moltMemoryLimitMB,
		TimeoutMs:     moltTimeoutMs,
		MaxInstances:  moltMaxInstances,
		Owner:         moltOwner,
	}

	var resp *client.MoltDeployResponse
	err = WithSpinner("Compiling and deploying", func() error {
		var e error
		resp, e = c.DaemonClient().MoltDeploy(req)
		return e
	})
	if err != nil {
		return fmt.Errorf("molt deploy failed: %w", err)
	}

	fields := [][2]string{
		{"Deployment", FormatNodeID(resp.DeploymentID)},
		{"Module CID", FormatNodeID(resp.ModuleCID)},
		{"Status", resp.Status},
	}

	fmt.Println(StatusBox("Molt Deployed", fields))
	fmt.Println(Hint(fmt.Sprintf("Invoke: moltbunker molt invoke %s", FormatNodeID(resp.DeploymentID))))

	return nil
}

func runMoltList(_ *cobra.Command, _ []string) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	molts, err := c.DaemonClient().MoltList()
	if err != nil {
		return fmt.Errorf("failed to list molts: %w", err)
	}

	if len(molts) == 0 {
		Info("No Molts deployed")
		fmt.Println(Hint("Deploy one: moltbunker molt deploy <file.wasm>"))
		return nil
	}

	headers := []string{"ID", "Status", "Memory", "Timeout", "Invocations", "Created"}
	rows := make([][]string, 0, len(molts))
	for _, m := range molts {
		invocations := "-"
		if m.Metrics != nil {
			invocations = fmt.Sprintf("%d", m.Metrics.TotalInvocations)
		}
		rows = append(rows, []string{
			FormatNodeID(m.ID),
			m.Status,
			fmt.Sprintf("%dMB", m.MemoryLimitMB),
			fmt.Sprintf("%dms", m.TimeoutMs),
			invocations,
			m.CreatedAt.Format("Jan 02 15:04"),
		})
	}

	fmt.Println(RenderTable(headers, rows))

	return nil
}

func runMoltGet(_ *cobra.Command, args []string) error {
	deploymentID := args[0]

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	info, err := c.DaemonClient().MoltGet(deploymentID)
	if err != nil {
		return fmt.Errorf("failed to get molt: %w", err)
	}

	fields := [][2]string{
		{"ID", info.ID},
		{"Module CID", FormatNodeID(info.ModuleCID)},
		{"Status", info.Status},
		{"Memory", fmt.Sprintf("%d MB", info.MemoryLimitMB)},
		{"Timeout", fmt.Sprintf("%d ms", info.TimeoutMs)},
		{"Created", info.CreatedAt.Format("2006-01-02 15:04:05")},
	}
	if info.Owner != "" {
		fields = append(fields, [2]string{"Owner", FormatNodeID(info.Owner)})
	}
	if info.Metrics != nil {
		fields = append(fields,
			[2]string{"Invocations", fmt.Sprintf("%d total, %d ok, %d err, %d timeout",
				info.Metrics.TotalInvocations,
				info.Metrics.SuccessInvocations,
				info.Metrics.ErrorInvocations,
				info.Metrics.TimeoutInvocations)},
			[2]string{"Avg Latency", info.Metrics.AvgLatency.String()},
		)
	}

	fmt.Println(StatusBox("Molt", fields))

	return nil
}

func runMoltStop(_ *cobra.Command, args []string) error {
	deploymentID := args[0]

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Stopping Molt", func() error {
		return c.DaemonClient().MoltStop(deploymentID)
	})
	if err != nil {
		return fmt.Errorf("stop failed: %w", err)
	}

	Success(fmt.Sprintf("Stopped Molt: %s", FormatNodeID(deploymentID)))
	fmt.Println(Hint("Compiled cache retained — redeploy is fast"))

	return nil
}

func runMoltDelete(_ *cobra.Command, args []string) error {
	deploymentID := args[0]

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	err = WithSpinner("Deleting Molt", func() error {
		return c.DaemonClient().MoltDelete(deploymentID)
	})
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	Success(fmt.Sprintf("Deleted Molt: %s", FormatNodeID(deploymentID)))

	return nil
}

func runMoltInvoke(_ *cobra.Command, args []string) error {
	deploymentID := args[0]

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if err := c.RequireDaemon(); err != nil {
		return err
	}

	req := &client.MoltInvokeRequest{
		DeploymentID: deploymentID,
		Method:       strings.ToUpper(moltInvokeMethod),
		Path:         moltInvokePath,
	}
	if moltInvokeData != "" {
		req.Body = []byte(moltInvokeData)
	}

	var resp *client.MoltInvokeResponse
	err = WithSpinner("Invoking Molt", func() error {
		var e error
		resp, e = c.DaemonClient().MoltInvoke(req)
		return e
	})
	if err != nil {
		return fmt.Errorf("invocation failed: %w", err)
	}

	fields := [][2]string{
		{"Status", fmt.Sprintf("%d", resp.StatusCode)},
		{"Duration", fmt.Sprintf("%dms", resp.DurationMs)},
	}
	if resp.Error != "" {
		fields = append(fields, [2]string{"Error", resp.Error})
	}
	for k, v := range resp.Headers {
		fields = append(fields, [2]string{fmt.Sprintf("Header[%s]", k), v})
	}

	fmt.Println(StatusBox("Molt Response", fields))

	if len(resp.Body) > 0 {
		fmt.Println(string(resp.Body))
	}

	return nil
}
