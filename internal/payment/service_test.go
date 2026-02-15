package payment

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/pkg/types"
)

func TestPaymentService_MockMode(t *testing.T) {
	config := &PaymentServiceConfig{
		MockMode: true,
	}

	ps, err := NewPaymentService(config)
	if err != nil {
		t.Fatalf("failed to create payment service: %v", err)
	}

	if !ps.IsMockMode() {
		t.Error("expected mock mode to be true")
	}

	ctx := context.Background()
	if err := ps.Start(ctx); err != nil {
		t.Fatalf("failed to start payment service: %v", err)
	}
	defer ps.Stop()

	if !ps.IsConnected() {
		t.Error("expected to be connected in mock mode")
	}
}

func TestTokenContract_MockOperations(t *testing.T) {
	tc := NewMockTokenContract()

	ctx := context.Background()
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	// Set mock balance
	balance := parseWei("1000")
	tc.SetMockBalance(addr1, balance)

	// Check balance
	got, err := tc.BalanceOf(ctx, addr1)
	if err != nil {
		t.Fatalf("failed to get balance: %v", err)
	}
	if got.Cmp(balance) != 0 {
		t.Errorf("expected balance %s, got %s", balance.String(), got.String())
	}

	// Check zero balance
	got, err = tc.BalanceOf(ctx, addr2)
	if err != nil {
		t.Fatalf("failed to get balance: %v", err)
	}
	if got.Sign() != 0 {
		t.Errorf("expected zero balance, got %s", got.String())
	}
}

func TestStakingContract_MockOperations(t *testing.T) {
	sc := NewMockStakingContract()

	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")

	// Initial stake should be zero
	stake, err := sc.GetStake(ctx, provider)
	if err != nil {
		t.Fatalf("failed to get stake: %v", err)
	}
	if stake.Sign() != 0 {
		t.Errorf("expected zero stake, got %s", stake.String())
	}

	// Check tier for zero stake
	tier, err := sc.GetTier(ctx, provider)
	if err != nil {
		t.Fatalf("failed to get tier: %v", err)
	}
	if tier != "" {
		t.Errorf("expected empty tier, got %v", tier)
	}

	// Check minimum stake
	hasMin, err := sc.HasMinimumStake(ctx, provider)
	if err != nil {
		t.Fatalf("failed to check minimum stake: %v", err)
	}
	if hasMin {
		t.Error("expected to not have minimum stake")
	}
}

func TestEscrowContract_MockOperations(t *testing.T) {
	ec := NewMockEscrowContract()

	ctx := context.Background()
	jobID := [32]byte{1, 2, 3, 4}
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")
	amount := parseWei("100")
	duration := big.NewInt(3600) // 1 hour

	// Create escrow (puts it in Created state)
	_, err := ec.CreateEscrow(ctx, jobID, provider, amount, duration)
	if err != nil {
		t.Fatalf("failed to create escrow: %v", err)
	}

	// Get escrow — should be in Created state before provider selection
	escrow, err := ec.GetEscrow(ctx, jobID)
	if err != nil {
		t.Fatalf("failed to get escrow: %v", err)
	}

	if escrow.Amount.Cmp(amount) != 0 {
		t.Errorf("expected amount %s, got %s", amount.String(), escrow.Amount.String())
	}
	if escrow.State != EscrowStateCreated {
		t.Errorf("expected state Created, got %v", escrow.State)
	}

	// Select providers to transition escrow to Active
	providers := [3]common.Address{provider, {}, {}}
	if _, err := ec.SelectProviders(ctx, jobID, providers); err != nil {
		t.Fatalf("failed to select providers: %v", err)
	}

	escrow, err = ec.GetEscrow(ctx, jobID)
	if err != nil {
		t.Fatalf("failed to get escrow after provider selection: %v", err)
	}
	if escrow.State != EscrowStateActive {
		t.Errorf("expected state Active after provider selection, got %v", escrow.State)
	}

	// Release partial payment
	uptime := big.NewInt(1800) // 30 minutes
	_, err = ec.ReleasePayment(ctx, jobID, uptime)
	if err != nil {
		t.Fatalf("failed to release payment: %v", err)
	}

	// Check released amount
	escrow, _ = ec.GetEscrow(ctx, jobID)
	expectedReleased := new(big.Int).Div(amount, big.NewInt(2)) // 50%
	if escrow.Released.Cmp(expectedReleased) != 0 {
		t.Errorf("expected released %s, got %s", expectedReleased.String(), escrow.Released.String())
	}

	// Finalize escrow
	_, err = ec.FinalizeEscrow(ctx, jobID)
	if err != nil {
		t.Fatalf("failed to finalize escrow: %v", err)
	}

	escrow, _ = ec.GetEscrow(ctx, jobID)
	if escrow.State != EscrowStateCompleted {
		t.Errorf("expected state Completed, got %v", escrow.State)
	}
}

func TestSlashingContract_MockOperations(t *testing.T) {
	sc := NewMockSlashingContract()

	ctx := context.Background()
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")
	jobID := [32]byte{1, 2, 3, 4}
	evidence := []byte("proof of violation")

	// Report violation
	disputeID, _, err := sc.ReportViolation(ctx, provider, jobID, ViolationDowntime, evidence)
	if err != nil {
		t.Fatalf("failed to report violation: %v", err)
	}

	// Get dispute
	dispute, err := sc.GetDispute(ctx, disputeID)
	if err != nil {
		t.Fatalf("failed to get dispute: %v", err)
	}

	if dispute.Provider != provider {
		t.Errorf("expected provider %s, got %s", provider.Hex(), dispute.Provider.Hex())
	}
	if dispute.Reason != ViolationDowntime {
		t.Errorf("expected reason Downtime, got %v", dispute.Reason)
	}
	if dispute.State != DisputeStatePending {
		t.Errorf("expected state Pending, got %v", dispute.State)
	}

	// Submit defense
	defense := []byte("defense evidence")
	_, err = sc.SubmitDefense(ctx, disputeID, defense)
	if err != nil {
		t.Fatalf("failed to submit defense: %v", err)
	}

	dispute, _ = sc.GetDispute(ctx, disputeID)
	if dispute.State != DisputeStateDefenseSubmitted {
		t.Errorf("expected state DefenseSubmitted, got %v", dispute.State)
	}

	// Resolve dispute
	slashAmount := parseWei("10")
	_, err = sc.ResolveDispute(ctx, disputeID, slashAmount)
	if err != nil {
		t.Fatalf("failed to resolve dispute: %v", err)
	}

	dispute, _ = sc.GetDispute(ctx, disputeID)
	if dispute.State != DisputeStateResolved {
		t.Errorf("expected state Resolved, got %v", dispute.State)
	}

	// Check slashing history
	history, err := sc.GetSlashingHistory(ctx, provider)
	if err != nil {
		t.Fatalf("failed to get slashing history: %v", err)
	}
	if history.TotalSlashed.Cmp(slashAmount) != 0 {
		t.Errorf("expected total slashed %s, got %s", slashAmount.String(), history.TotalSlashed.String())
	}
	if history.Violations != 1 {
		t.Errorf("expected 1 violation, got %d", history.Violations)
	}
}

func TestSlashingContract_CalculateSlashAmount(t *testing.T) {
	sc := NewMockSlashingContract()
	stake := parseWei("10000") // 10000 BUNKER

	tests := []struct {
		reason         ViolationReason
		expectedPercent int64
	}{
		{ViolationDowntime, 5},
		{ViolationSLAViolation, 10},
		{ViolationJobAbandonment, 15},
		{ViolationSecurityViolation, 50},
		{ViolationFraud, 100},
		{ViolationDataLoss, 25},
		{ViolationNone, 0},
	}

	for _, tt := range tests {
		t.Run(tt.reason.String(), func(t *testing.T) {
			amount := sc.CalculateSlashAmount(stake, tt.reason)
			expected := new(big.Int).Mul(stake, big.NewInt(tt.expectedPercent))
			expected.Div(expected, big.NewInt(100))

			if amount.Cmp(expected) != 0 {
				t.Errorf("expected slash %s, got %s", expected.String(), amount.String())
			}
		})
	}
}

func TestTokenAmount_FormatAndParse(t *testing.T) {
	tests := []struct {
		amount   string
		expected string
	}{
		{"1000000000000000000", "1"},      // 1 token
		{"1500000000000000000", "1.5"},    // 1.5 tokens
		{"123456789000000000000", "123.4567"}, // Truncated to 4 decimals
		{"0", "0"},
	}

	for _, tt := range tests {
		amount, _ := new(big.Int).SetString(tt.amount, 10)
		formatted := FormatTokenAmount(amount)
		if formatted != tt.expected {
			t.Errorf("FormatTokenAmount(%s) = %s, expected %s", tt.amount, formatted, tt.expected)
		}
	}
}

func TestTokenAmount_Parse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1", "1000000000000000000"},
		{"1.5", "1500000000000000000"},
		{"100.25", "100250000000000000000"},
	}

	for _, tt := range tests {
		amount, err := ParseTokenAmount(tt.input)
		if err != nil {
			t.Errorf("ParseTokenAmount(%s) error: %v", tt.input, err)
			continue
		}
		expected, _ := new(big.Int).SetString(tt.expected, 10)
		if amount.Cmp(expected) != 0 {
			t.Errorf("ParseTokenAmount(%s) = %s, expected %s", tt.input, amount.String(), expected.String())
		}
	}
}

func TestPricingCalculator_CalculatePriceWithResources(t *testing.T) {
	basePricePerHour := parseWei("1") // 1 BUNKER per hour
	pc := NewPricingCalculator(basePricePerHour)

	resources := types.ResourceLimits{
		CPUQuota:    100000,           // 1 CPU
		MemoryLimit: 2 * 1024 * 1024 * 1024, // 2GB
		DiskLimit:   10 * 1024 * 1024 * 1024, // 10GB
		NetworkBW:   100 * 1024 * 1024, // 100 MB/s
	}

	duration := 1 * time.Hour
	price := pc.CalculatePrice(resources, duration)

	if price.Sign() <= 0 {
		t.Error("expected positive price")
	}
}

func TestEscrowState_String(t *testing.T) {
	tests := []struct {
		state    EscrowState
		expected string
	}{
		{EscrowStateActive, "active"},
		{EscrowStateCompleted, "completed"},
		{EscrowStateRefunded, "refunded"},
		{EscrowStateDisputed, "disputed"},
		{EscrowState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("EscrowState(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}

func TestDisputeState_String(t *testing.T) {
	tests := []struct {
		state    DisputeState
		expected string
	}{
		{DisputeStatePending, "pending"},
		{DisputeStateDefenseSubmitted, "defense_submitted"},
		{DisputeStateResolved, "resolved"},
		{DisputeStateExpired, "expired"},
		{DisputeState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("DisputeState(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}

func TestViolationReason_String(t *testing.T) {
	tests := []struct {
		reason   ViolationReason
		expected string
	}{
		{ViolationNone, "none"},
		{ViolationDowntime, "downtime"},
		{ViolationJobAbandonment, "job_abandonment"},
		{ViolationSecurityViolation, "security_violation"},
		{ViolationFraud, "fraud"},
		{ViolationSLAViolation, "sla_violation"},
		{ViolationDataLoss, "data_loss"},
		{ViolationReason(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.reason.String(); got != tt.expected {
			t.Errorf("ViolationReason(%d).String() = %q, want %q", tt.reason, got, tt.expected)
		}
	}
}

func TestJobID_Conversion(t *testing.T) {
	str := "test-job-12345"
	id := JobIDFromString(str)
	hex := JobIDToHex(id)

	if len(hex) != 64 {
		t.Errorf("expected hex length 64, got %d", len(hex))
	}
}
