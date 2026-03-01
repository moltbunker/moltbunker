package molt

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/security"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// TestMoltLifecycle_DeployInvokeStop tests the full Molt lifecycle:
// compile → invoke → metrics → close.
func TestMoltLifecycle_DeployInvokeStop(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "echo.wasm")

	// Phase 1: Compile
	compiled, err := rt.Compile(context.Background(), wasm, "lifecycle-cid")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if compiled.CID != "lifecycle-cid" {
		t.Fatalf("CID = %q, want lifecycle-cid", compiled.CID)
	}

	// Phase 2: Invoke multiple times
	bodies := []string{
		`{"action":"create"}`,
		`{"action":"read"}`,
		`{"action":"update"}`,
	}

	for i, body := range bodies {
		result, err := rt.Invoke(context.Background(), compiled, MoltInvocation{
			DeploymentID: "lifecycle-deploy",
			Method:       "POST",
			Path:         "/api/v1",
			Body:         []byte(body),
		})
		if err != nil {
			t.Fatalf("Invoke %d: %v", i, err)
		}
		if result.StatusCode != 200 {
			t.Fatalf("Invoke %d: status = %d, want 200", i, result.StatusCode)
		}
		if string(result.Body) != body {
			t.Fatalf("Invoke %d: body = %q, want %q", i, string(result.Body), body)
		}
		if result.Duration <= 0 {
			t.Fatalf("Invoke %d: duration = %v, want > 0", i, result.Duration)
		}
	}

	// Phase 3: Verify metrics
	stats := rt.Metrics().GetStats("lifecycle-deploy")
	if stats.TotalInvocations != 3 {
		t.Fatalf("TotalInvocations = %d, want 3", stats.TotalInvocations)
	}
	if stats.SuccessInvocations != 3 {
		t.Fatalf("SuccessInvocations = %d, want 3", stats.SuccessInvocations)
	}

	// Phase 4: Close
	if err := rt.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Phase 5: Invoke after close should fail
	_, err = rt.Invoke(context.Background(), compiled, MoltInvocation{
		DeploymentID: "lifecycle-deploy",
		Method:       "GET",
		Path:         "/",
	})
	if err == nil {
		t.Fatal("expected error invoking after close")
	}
}

// TestMoltLifecycle_BillingIntegration tests the Molt billing flow:
// deposit credits → invoke → calculate price → deduct → verify balance.
func TestMoltLifecycle_BillingIntegration(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "echo.wasm")

	compiled, err := rt.Compile(context.Background(), wasm, "billing-cid")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	// Set up credit manager
	creditMgr := payment.NewMoltCreditManager()
	requester := "0xdeadbeef1234567890abcdef1234567890abcdef"

	// Deposit 10 BUNKER (10^19 wei)
	deposit := new(big.Int).Mul(big.NewInt(10), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	creditMgr.Deposit(requester, deposit)

	initialBalance := creditMgr.GetBalance(requester)
	if initialBalance.Cmp(deposit) != 0 {
		t.Fatalf("initial balance = %s, want %s", initialBalance.String(), deposit.String())
	}

	// Set up pricing calculator
	pricingCalc := payment.NewPricingCalculator(big.NewInt(1000))
	pricingCfg := types.DefaultPricingConfig()

	// Invoke 5 times and bill for each
	totalCost := new(big.Int)
	for i := 0; i < 5; i++ {
		body := strings.Repeat("x", 1024)
		result, err := rt.Invoke(context.Background(), compiled, MoltInvocation{
			DeploymentID: "billing-deploy",
			Method:       "POST",
			Path:         "/process",
			Body:         []byte(body),
		})
		if err != nil {
			t.Fatalf("Invoke %d: %v", i, err)
		}
		if result.StatusCode != 200 {
			t.Fatalf("Invoke %d: status = %d", i, result.StatusCode)
		}

		// Calculate cost
		cost := pricingCalc.CalculateMoltInvocationPrice(
			result.Duration,
			int64(result.MemoryUsedBytes),
			pricingCfg,
		)
		if cost.Sign() <= 0 {
			t.Fatalf("Invoke %d: cost = %s, want > 0", i, cost.String())
		}

		// Deduct from credits
		if err := creditMgr.Deduct(requester, cost); err != nil {
			t.Fatalf("Invoke %d: Deduct: %v", i, err)
		}
		totalCost.Add(totalCost, cost)
	}

	// Verify final balance = initial - total cost
	finalBalance := creditMgr.GetBalance(requester)
	expected := new(big.Int).Sub(deposit, totalCost)
	if finalBalance.Cmp(expected) != 0 {
		t.Fatalf("final balance = %s, want %s (total cost: %s)",
			finalBalance.String(), expected.String(), totalCost.String())
	}

	// Verify credit details
	credit := creditMgr.GetCredit(requester)
	if credit == nil {
		t.Fatal("credit record not found")
	}
	if credit.TotalSpent.Cmp(totalCost) != 0 {
		t.Fatalf("TotalSpent = %s, want %s", credit.TotalSpent.String(), totalCost.String())
	}
	if credit.TotalDeposited.Cmp(deposit) != 0 {
		t.Fatalf("TotalDeposited = %s, want %s", credit.TotalDeposited.String(), deposit.String())
	}

	t.Logf("Billing: 5 invocations, total cost: %s BUNKER wei, remaining: %s",
		totalCost.String(), finalBalance.String())
}

// TestMoltLifecycle_InsufficientCredits verifies that invocations fail
// gracefully when credits are depleted.
func TestMoltLifecycle_InsufficientCredits(t *testing.T) {
	creditMgr := payment.NewMoltCreditManager()
	requester := "0xdeadbeef"

	// Deposit tiny amount (1 wei)
	creditMgr.Deposit(requester, big.NewInt(1))

	// Try to deduct more than available
	cost := new(big.Int).Mul(big.NewInt(1000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	err := creditMgr.Deduct(requester, cost)
	if err == nil {
		t.Fatal("expected insufficient credits error")
	}
	if !strings.Contains(err.Error(), "insufficient Molt credits") {
		t.Fatalf("error = %q, want 'insufficient Molt credits'", err.Error())
	}

	// Balance should be unchanged
	bal := creditMgr.GetBalance(requester)
	if bal.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("balance = %s, want 1", bal.String())
	}
}

// TestMoltLifecycle_CreditRefund tests the full deposit → use → refund cycle.
func TestMoltLifecycle_CreditRefund(t *testing.T) {
	creditMgr := payment.NewMoltCreditManager()
	requester := "0xcafe0001"

	// Deposit 100 BUNKER
	deposit := new(big.Int).Mul(big.NewInt(100), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	creditMgr.Deposit(requester, deposit)

	// Spend 30 BUNKER
	spend := new(big.Int).Mul(big.NewInt(30), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	if err := creditMgr.Deduct(requester, spend); err != nil {
		t.Fatalf("Deduct: %v", err)
	}

	// Refund remaining
	refund := creditMgr.RefundAll(requester)
	expected := new(big.Int).Sub(deposit, spend)
	if refund.Cmp(expected) != 0 {
		t.Fatalf("refund = %s, want %s", refund.String(), expected.String())
	}

	// After refund, credit should be gone
	if creditMgr.GetCredit(requester) != nil {
		t.Fatal("credit record should be removed after RefundAll")
	}

	// Balance should be 0
	bal := creditMgr.GetBalance(requester)
	if bal.Sign() != 0 {
		t.Fatalf("balance after refund = %s, want 0", bal.String())
	}
}

// TestMoltLifecycle_PricingMinimumFloor verifies the 100ms minimum billing floor.
func TestMoltLifecycle_PricingMinimumFloor(t *testing.T) {
	pricingCalc := payment.NewPricingCalculator(big.NewInt(1000))
	cfg := types.DefaultPricingConfig()

	// Very short invocation (1ms) should be billed at minimum (100ms)
	price1ms := pricingCalc.CalculateMoltInvocationPrice(1*time.Millisecond, 1024*1024, cfg)
	price100ms := pricingCalc.CalculateMoltInvocationPrice(100*time.Millisecond, 1024*1024, cfg)

	// Both should be the same (floor at 100ms)
	if price1ms.Cmp(price100ms) != 0 {
		t.Fatalf("1ms price = %s, 100ms price = %s — should match due to 100ms floor",
			price1ms.String(), price100ms.String())
	}

	// 200ms should cost more
	price200ms := pricingCalc.CalculateMoltInvocationPrice(200*time.Millisecond, 1024*1024, cfg)
	if price200ms.Cmp(price100ms) <= 0 {
		t.Fatalf("200ms price = %s should be > 100ms price = %s",
			price200ms.String(), price100ms.String())
	}

	// More memory should cost more
	price10MB := pricingCalc.CalculateMoltInvocationPrice(100*time.Millisecond, 10*1024*1024, cfg)
	if price10MB.Cmp(price100ms) <= 0 {
		t.Fatalf("10MB price = %s should be > 1MB price = %s",
			price10MB.String(), price100ms.String())
	}

	t.Logf("Pricing: 1ms=%s, 100ms=%s, 200ms=%s, 10MB=%s BUNKER wei",
		price1ms.String(), price100ms.String(), price200ms.String(), price10MB.String())
}

// TestMoltLifecycle_EncryptedInvocationWithBilling tests the combined
// E2E encryption + billing flow.
func TestMoltLifecycle_EncryptedInvocationWithBilling(t *testing.T) {
	rt := newTestRuntime(t)
	wasm := loadTestWASM(t, "echo.wasm")

	compiled, err := rt.Compile(context.Background(), wasm, "enc-bill-cid")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	// Set up encryption
	requesterPub, requesterPriv, err := security.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair: %v", err)
	}
	em := security.NewDeploymentEncryptionManager(t.TempDir())
	if _, err := em.SetupDeploymentEncryption("enc-bill-deploy", requesterPub); err != nil {
		t.Fatalf("SetupDeploymentEncryption: %v", err)
	}
	decryptor, err := security.NewRequesterDecryptor(requesterPriv, requesterPub)
	if err != nil {
		t.Fatalf("NewRequesterDecryptor: %v", err)
	}

	// Set up billing
	creditMgr := payment.NewMoltCreditManager()
	requester := "0xaabb1122"
	deposit := new(big.Int).Mul(big.NewInt(10), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	creditMgr.Deposit(requester, deposit)

	pricingCalc := payment.NewPricingCalculator(big.NewInt(1000))
	pricingCfg := types.DefaultPricingConfig()

	// Invoke with encrypted data
	plaintext := `{"secret":"classified-data"}`
	encrypted, err := em.EncryptData("enc-bill-deploy", []byte(plaintext))
	if err != nil {
		t.Fatalf("EncryptData: %v", err)
	}

	// Decrypt on "provider side" and invoke
	decrypted, err := em.DecryptData("enc-bill-deploy", encrypted)
	if err != nil {
		t.Fatalf("DecryptData: %v", err)
	}

	result, err := rt.Invoke(context.Background(), compiled, MoltInvocation{
		DeploymentID: "enc-bill-deploy",
		Method:       "POST",
		Path:         "/secret",
		Body:         decrypted,
	})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if result.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", result.StatusCode)
	}

	// Encrypt the response
	encryptedResponse, err := em.EncryptData("enc-bill-deploy", result.Body)
	if err != nil {
		t.Fatalf("EncryptData response: %v", err)
	}

	// Requester decrypts
	metadata, err := em.GetEncryptionMetadata("enc-bill-deploy")
	if err != nil {
		t.Fatalf("GetEncryptionMetadata: %v", err)
	}
	decryptedResponse, err := decryptor.DecryptOutput(metadata, encryptedResponse)
	if err != nil {
		t.Fatalf("DecryptOutput: %v", err)
	}
	if string(decryptedResponse) != plaintext {
		t.Fatalf("decrypted = %q, want %q", string(decryptedResponse), plaintext)
	}

	// Bill for invocation
	cost := pricingCalc.CalculateMoltInvocationPrice(
		result.Duration, int64(result.MemoryUsedBytes), pricingCfg,
	)
	if err := creditMgr.Deduct(requester, cost); err != nil {
		t.Fatalf("Deduct: %v", err)
	}

	finalBalance := creditMgr.GetBalance(requester)
	if finalBalance.Cmp(deposit) >= 0 {
		t.Fatal("balance should have decreased after billing")
	}

	t.Logf("Encrypted + billed invocation: cost=%s wei, remaining=%s wei",
		cost.String(), finalBalance.String())
}

// TestMoltLifecycle_TimeoutInvocation verifies that timeout produces
// a 504 result and is counted in error metrics.
func TestMoltLifecycle_TimeoutInvocation(t *testing.T) {
	ctx := context.Background()
	cfg := MoltConfig{
		MemoryLimitMB:   64,
		TimeoutMs:       100, // 100ms timeout — very short
		MaxInstances:    10,
		CacheDir:        t.TempDir(),
		MaxCacheEntries: 16,
	}
	rt, err := NewMoltRuntime(ctx, cfg)
	if err != nil {
		t.Fatalf("NewMoltRuntime: %v", err)
	}
	t.Cleanup(func() { rt.Close(ctx) })

	wasm := loadTestWASM(t, "spin.wasm")
	compiled, err := rt.Compile(ctx, wasm, "spin-cid")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}

	result, err := rt.Invoke(ctx, compiled, MoltInvocation{
		DeploymentID: "timeout-deploy",
		Method:       "GET",
		Path:         "/infinite",
	})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}

	if result.StatusCode != 504 {
		t.Fatalf("status = %d, want 504 (timeout)", result.StatusCode)
	}
	if result.Error == "" {
		t.Fatal("expected error message for timeout")
	}

	// Metrics should show 1 timeout
	stats := rt.Metrics().GetStats("timeout-deploy")
	if stats.TimeoutInvocations != 1 {
		t.Fatalf("TimeoutInvocations = %d, want 1", stats.TimeoutInvocations)
	}
}
