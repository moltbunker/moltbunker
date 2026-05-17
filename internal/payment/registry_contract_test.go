package payment

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestMockRegistry_RegisterAndResolve(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("test-deployment-001"))

	_, err := rc.Register(ctx, "myapp", depID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	reg, err := rc.Resolve(ctx, "myapp")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if reg.Name != "myapp" {
		t.Errorf("expected name 'myapp', got %q", reg.Name)
	}
	if reg.DeploymentID != depID {
		t.Errorf("deployment ID mismatch")
	}
}

func TestMockRegistry_RegisterAndRelease(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("dep1"))

	_, err := rc.Register(ctx, "releaseme", depID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	avail, err := rc.IsAvailable(ctx, "releaseme")
	if err != nil {
		t.Fatalf("IsAvailable failed: %v", err)
	}
	if avail {
		t.Error("expected name to NOT be available after registration")
	}

	_, err = rc.Release(ctx, "releaseme")
	if err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	avail, err = rc.IsAvailable(ctx, "releaseme")
	if err != nil {
		t.Fatalf("IsAvailable failed: %v", err)
	}
	if !avail {
		t.Error("expected name to be available after release")
	}
}

func TestMockRegistry_ReserveAndClaim(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	_, err := rc.Reserve(ctx, "reserved-name")
	if err != nil {
		t.Fatalf("Reserve failed: %v", err)
	}

	// Name should not be available after reservation
	avail, err := rc.IsAvailable(ctx, "reserved-name")
	if err != nil {
		t.Fatalf("IsAvailable failed: %v", err)
	}
	if avail {
		t.Error("expected name to NOT be available after reservation")
	}

	// Claim with deployment ID
	var depID [32]byte
	copy(depID[:], []byte("my-deployment"))

	_, err = rc.ClaimReservation(ctx, "reserved-name", depID)
	if err != nil {
		t.Fatalf("ClaimReservation failed: %v", err)
	}

	// Resolve should now return the deployment
	reg, err := rc.Resolve(ctx, "reserved-name")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if reg.DeploymentID != depID {
		t.Error("deployment ID mismatch after claim")
	}
}

func TestMockRegistry_CancelReservation(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	_, err := rc.Reserve(ctx, "cancel-me")
	if err != nil {
		t.Fatalf("Reserve failed: %v", err)
	}

	_, err = rc.CancelReservation(ctx, "cancel-me")
	if err != nil {
		t.Fatalf("CancelReservation failed: %v", err)
	}

	avail, err := rc.IsAvailable(ctx, "cancel-me")
	if err != nil {
		t.Fatalf("IsAvailable failed: %v", err)
	}
	if !avail {
		t.Error("expected name to be available after cancellation")
	}
}

func TestMockRegistry_Transfer(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("dep-transfer"))

	_, err := rc.Register(ctx, "xferme", depID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	newOwner := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	_, err = rc.Transfer(ctx, "xferme", newOwner)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	reg, err := rc.Resolve(ctx, "xferme")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if reg.Owner != newOwner {
		t.Errorf("expected owner %s, got %s", newOwner.Hex(), reg.Owner.Hex())
	}
}

func TestMockRegistry_UpdateDeployment(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID1 [32]byte
	copy(depID1[:], []byte("dep-v1"))
	_, err := rc.Register(ctx, "updatable", depID1)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	var depID2 [32]byte
	copy(depID2[:], []byte("dep-v2"))
	_, err = rc.UpdateDeployment(ctx, "updatable", depID2)
	if err != nil {
		t.Fatalf("UpdateDeployment failed: %v", err)
	}

	reg, err := rc.Resolve(ctx, "updatable")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if reg.DeploymentID != depID2 {
		t.Error("deployment ID not updated")
	}
}

func TestMockRegistry_SetMetadataAndGet(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("dep-meta"))
	_, err := rc.Register(ctx, "metaapp", depID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err = rc.SetMetadata(ctx, "metaapp", "My cool app", "https://example.com/avatar.png")
	if err != nil {
		t.Fatalf("SetMetadata failed: %v", err)
	}

	meta, err := rc.GetMetadata(ctx, "metaapp")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}
	if meta.Description != "My cool app" {
		t.Errorf("expected description 'My cool app', got %q", meta.Description)
	}
	if meta.AvatarURL != "https://example.com/avatar.png" {
		t.Errorf("expected avatar URL, got %q", meta.AvatarURL)
	}
}

func TestMockRegistry_SetPrimaryAndReverse(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("dep-primary"))
	_, err := rc.Register(ctx, "primary-app", depID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err = rc.SetPrimaryName(ctx, "primary-app")
	if err != nil {
		t.Fatalf("SetPrimaryName failed: %v", err)
	}

	name, err := rc.ReverseResolve(ctx, depID)
	if err != nil {
		t.Fatalf("ReverseResolve failed: %v", err)
	}
	if name != "primary-app" {
		t.Errorf("expected 'primary-app', got %q", name)
	}
}

func TestMockRegistry_ReverseResolve_NoPrimary(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("dep-no-primary"))
	_, err := rc.Register(ctx, "no-primary", depID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Without SetPrimaryName, ReverseResolve should return empty
	name, err := rc.ReverseResolve(ctx, depID)
	if err != nil {
		t.Fatalf("ReverseResolve failed: %v", err)
	}
	if name != "" {
		t.Errorf("expected empty name, got %q", name)
	}
}

func TestMockRegistry_Renew(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("dep-renew"))
	_, err := rc.Register(ctx, "renewable", depID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Get initial expiry
	rc.mockMu.RLock()
	initialExpiry := rc.mockNames["renewable"].ExpiresAt
	rc.mockMu.RUnlock()

	_, err = rc.Renew(ctx, "renewable")
	if err != nil {
		t.Fatalf("Renew failed: %v", err)
	}

	rc.mockMu.RLock()
	newExpiry := rc.mockNames["renewable"].ExpiresAt
	rc.mockMu.RUnlock()

	// Expiry should be extended by ~365 days
	diff := newExpiry.Sub(initialExpiry)
	if diff < 364*24*3600e9 || diff > 366*24*3600e9 { // Allow 1 day tolerance
		t.Errorf("expected ~365 day extension, got %v", diff)
	}
}

func TestMockRegistry_ReclaimSquatted(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("squatted-dep"))
	_, err := rc.Register(ctx, "squatted", depID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err = rc.ReclaimSquatted(ctx, "squatted")
	if err != nil {
		t.Fatalf("ReclaimSquatted failed: %v", err)
	}

	avail, err := rc.IsAvailable(ctx, "squatted")
	if err != nil {
		t.Fatalf("IsAvailable failed: %v", err)
	}
	if !avail {
		t.Error("expected name to be available after reclaim")
	}
}

func TestMockRegistry_ListOwnedNames(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	owner := mockDefaultOwner // mock mode uses deterministic default owner

	var depID1, depID2 [32]byte
	copy(depID1[:], []byte("dep-list-1"))
	copy(depID2[:], []byte("dep-list-2"))

	if _, err := rc.Register(ctx, "list-one", depID1); err != nil {
		t.Fatalf("Register list-one: %v", err)
	}
	if _, err := rc.Register(ctx, "list-two", depID2); err != nil {
		t.Fatalf("Register list-two: %v", err)
	}

	names, err := rc.ListOwnedNames(ctx, owner)
	if err != nil {
		t.Fatalf("ListOwnedNames failed: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestMockRegistry_DuplicateRegister(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("dep-dup"))

	_, err := rc.Register(ctx, "taken", depID)
	if err != nil {
		t.Fatalf("First register failed: %v", err)
	}

	_, err = rc.Register(ctx, "taken", depID)
	if err == nil {
		t.Error("expected error on duplicate registration")
	}
}

func TestMockRegistry_ResolveNonExistent(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	_, err := rc.Resolve(ctx, "does-not-exist")
	if err == nil {
		t.Error("expected error resolving non-existent name")
	}
}

func TestMockRegistry_IsExpired(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("dep-expiry"))

	_, err := rc.Register(ctx, "checkexpiry", depID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Freshly registered — should not be expired
	expired, err := rc.IsExpired(ctx, "checkexpiry")
	if err != nil {
		t.Fatalf("IsExpired failed: %v", err)
	}
	if expired {
		t.Error("expected freshly registered name to not be expired")
	}

	// Non-existent name — should not be expired
	expired, err = rc.IsExpired(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("IsExpired failed: %v", err)
	}
	if expired {
		t.Error("expected non-existent name to not be expired")
	}
}

func TestMockRegistry_GetMetadata_NonExistent(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	meta, err := rc.GetMetadata(ctx, "no-such-name")
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}
	if meta.Description != "" || meta.AvatarURL != "" {
		t.Error("expected empty metadata for non-existent name")
	}
}

func TestMockRegistry_RegisterWithReferral(t *testing.T) {
	rc := NewMockRegistryContract()
	ctx := context.Background()

	var depID [32]byte
	copy(depID[:], []byte("dep-referral"))
	referrer := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

	_, err := rc.RegisterWithReferral(ctx, "referral-app", depID, referrer)
	if err != nil {
		t.Fatalf("RegisterWithReferral failed: %v", err)
	}

	reg, err := rc.Resolve(ctx, "referral-app")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if reg.Name != "referral-app" {
		t.Errorf("expected name 'referral-app', got %q", reg.Name)
	}
}
