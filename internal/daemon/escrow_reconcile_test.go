package daemon

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// mockEscrowFinalizer records FinalizeJob + RegisterExternalReservation calls
// and can be configured to return a canned error. It satisfies the
// escrowFinalizer interface. Used for the status/replica/transient unit tests
// that do not need the real reservation-cache behavior.
type mockEscrowFinalizer struct {
	mu         sync.Mutex
	calls      int
	jobIDs     [][32]byte
	registered map[[32]byte]*big.Int
	retErr     error
	done       chan struct{} // closed-once signal after the expected number of calls
	expected   int
}

func newMockFinalizer(expected int) *mockEscrowFinalizer {
	return &mockEscrowFinalizer{
		done:       make(chan struct{}),
		expected:   expected,
		registered: make(map[[32]byte]*big.Int),
	}
}

func (m *mockEscrowFinalizer) RegisterExternalReservation(jobID [32]byte, reservationID *big.Int) {
	m.mu.Lock()
	m.registered[jobID] = new(big.Int).Set(reservationID)
	m.mu.Unlock()
}

func (m *mockEscrowFinalizer) FinalizeJob(_ context.Context, jobID [32]byte) error {
	m.mu.Lock()
	m.calls++
	m.jobIDs = append(m.jobIDs, jobID)
	reached := m.calls >= m.expected
	err := m.retErr
	m.mu.Unlock()
	if reached {
		// Non-blocking close-once.
		select {
		case <-m.done:
		default:
			close(m.done)
		}
	}
	return err
}

func (m *mockEscrowFinalizer) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// newReconcileCM builds a minimal ContainerManager wired with a node ID and a
// temp data dir (so saveStateAsync writes to JSON without error).
func newReconcileCM(t *testing.T, nodeID types.NodeID) *ContainerManager {
	t.Helper()
	return &ContainerManager{
		deployments: make(map[string]*Deployment),
		dataDir:     t.TempDir(),
		node:        &Node{nodeInfo: &types.Node{ID: nodeID}},
	}
}

func nodeIDFrom(b byte) types.NodeID {
	var id types.NodeID
	id[0] = b
	return id
}

// waitFor polls fn until it returns true or the deadline elapses.
func waitFor(t *testing.T, fn func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

func TestReconcileEscrowsOnStartup_FinalizesStopped(t *testing.T) {
	nodeID := nodeIDFrom(0x01)
	cm := newReconcileCM(t, nodeID)
	cm.deployments["dep-stop"] = &Deployment{
		ID:              "dep-stop",
		Status:          types.ContainerStatusStopped,
		OriginatorID:    nodeID,
		EscrowFinalized: false,
	}

	fin := newMockFinalizer(1)
	cm.reconcileEscrowsOnStartup(context.Background(), fin)

	if !waitFor(t, func() bool { return fin.callCount() == 1 }) {
		t.Fatalf("expected FinalizeJob called once, got %d", fin.callCount())
	}
	if !waitFor(t, func() bool {
		cm.mu.RLock()
		defer cm.mu.RUnlock()
		return cm.deployments["dep-stop"].EscrowFinalized
	}) {
		t.Fatal("expected EscrowFinalized to become true")
	}
}

func TestReconcileEscrowsOnStartup_SkipsAlreadyFinalized(t *testing.T) {
	nodeID := nodeIDFrom(0x02)
	cm := newReconcileCM(t, nodeID)
	cm.deployments["dep-done"] = &Deployment{
		ID:              "dep-done",
		Status:          types.ContainerStatusStopped,
		OriginatorID:    nodeID,
		EscrowFinalized: true, // already finalized
	}

	fin := newMockFinalizer(1)
	cm.reconcileEscrowsOnStartup(context.Background(), fin)

	// Give any (erroneously) spawned goroutine a chance to run.
	time.Sleep(100 * time.Millisecond)
	if got := fin.callCount(); got != 0 {
		t.Fatalf("expected FinalizeJob NOT called, got %d", got)
	}
}

func TestReconcileEscrowsOnStartup_SkipsReplicas(t *testing.T) {
	nodeID := nodeIDFrom(0x03)
	otherNode := nodeIDFrom(0x99)
	cm := newReconcileCM(t, nodeID)
	cm.deployments["dep-replica"] = &Deployment{
		ID:              "dep-replica",
		Status:          types.ContainerStatusStopped,
		OriginatorID:    otherNode, // not this node => replica, must be skipped
		EscrowFinalized: false,
	}

	fin := newMockFinalizer(1)
	cm.reconcileEscrowsOnStartup(context.Background(), fin)

	time.Sleep(100 * time.Millisecond)
	if got := fin.callCount(); got != 0 {
		t.Fatalf("expected FinalizeJob NOT called for replica, got %d", got)
	}
	cm.mu.RLock()
	finalized := cm.deployments["dep-replica"].EscrowFinalized
	cm.mu.RUnlock()
	if finalized {
		t.Error("replica deployment must not be marked finalized")
	}
}

func TestReconcileEscrowsOnStartup_AlreadyFinalizedOnChain(t *testing.T) {
	nodeID := nodeIDFrom(0x04)
	cm := newReconcileCM(t, nodeID)
	cm.deployments["dep-onchain"] = &Deployment{
		ID:              "dep-onchain",
		Status:          types.ContainerStatusFailed,
		OriginatorID:    nodeID,
		EscrowFinalized: false,
	}

	fin := newMockFinalizer(1)
	// Precise terminal custom-error phrasing (InvalidStatus actual=Completed).
	fin.retErr = fmt.Errorf("execution reverted: reservation already finalized")
	cm.reconcileEscrowsOnStartup(context.Background(), fin)

	if !waitFor(t, func() bool { return fin.callCount() == 1 }) {
		t.Fatalf("expected FinalizeJob called once, got %d", fin.callCount())
	}
	// Despite the error, the "already finalized" message must mark it done.
	if !waitFor(t, func() bool {
		cm.mu.RLock()
		defer cm.mu.RUnlock()
		return cm.deployments["dep-onchain"].EscrowFinalized
	}) {
		t.Fatal("expected EscrowFinalized=true after already-finalized error")
	}
}

func TestReconcileEscrowsOnStartup_TransientErrorLeavesUnfinalized(t *testing.T) {
	nodeID := nodeIDFrom(0x05)
	cm := newReconcileCM(t, nodeID)
	cm.deployments["dep-transient"] = &Deployment{
		ID:              "dep-transient",
		Status:          types.ContainerStatusStopped,
		OriginatorID:    nodeID,
		EscrowFinalized: false,
	}

	fin := newMockFinalizer(1)
	fin.retErr = fmt.Errorf("dial tcp: connection refused")
	cm.reconcileEscrowsOnStartup(context.Background(), fin)

	if !waitFor(t, func() bool { return fin.callCount() == 1 }) {
		t.Fatalf("expected FinalizeJob called once, got %d", fin.callCount())
	}
	// A transient error must leave the flag false so a later restart retries.
	time.Sleep(100 * time.Millisecond)
	cm.mu.RLock()
	finalized := cm.deployments["dep-transient"].EscrowFinalized
	cm.mu.RUnlock()
	if finalized {
		t.Error("transient error must NOT mark EscrowFinalized true")
	}
}

// TestReconcileEscrowsOnStartup_NotYetFinalizableIsTransient asserts the HIGH-2
// regression: an ambiguous "not yet finalizable" revert (which contains the
// substring "finalized") must be treated as TRANSIENT — the live escrow must
// NOT be marked finalized.
func TestReconcileEscrowsOnStartup_NotYetFinalizableIsTransient(t *testing.T) {
	nodeID := nodeIDFrom(0x07)
	cm := newReconcileCM(t, nodeID)
	cm.deployments["dep-notyet"] = &Deployment{
		ID:              "dep-notyet",
		Status:          types.ContainerStatusStopped,
		OriginatorID:    nodeID,
		EscrowFinalized: false,
	}

	fin := newMockFinalizer(1)
	fin.retErr = fmt.Errorf("execution reverted: reservation not yet finalizable")
	cm.reconcileEscrowsOnStartup(context.Background(), fin)

	if !waitFor(t, func() bool { return fin.callCount() == 1 }) {
		t.Fatalf("expected FinalizeJob called once, got %d", fin.callCount())
	}
	time.Sleep(100 * time.Millisecond)
	cm.mu.RLock()
	finalized := cm.deployments["dep-notyet"].EscrowFinalized
	cm.mu.RUnlock()
	if finalized {
		t.Error("'not yet finalizable' must be TRANSIENT — escrow must NOT be abandoned")
	}
}

func TestReconcileEscrowsOnStartup_NilFinalizer(t *testing.T) {
	cm := newReconcileCM(t, nodeIDFrom(0x06))
	cm.deployments["d"] = &Deployment{
		ID: "d", Status: types.ContainerStatusStopped, OriginatorID: nodeIDFrom(0x06),
	}
	// Must not panic with a nil finalizer.
	cm.reconcileEscrowsOnStartup(context.Background(), nil)
}

// TestReconcileEscrowsOnStartup_RehydratesPersistedReservation asserts the
// reconciler rehydrates the persisted Deployment.ReservationID into the cache
// (via RegisterExternalReservation) BEFORE finalizing, and skips orphans with
// no persisted reservation ID for rehydration.
func TestReconcileEscrowsOnStartup_RehydratesPersistedReservation(t *testing.T) {
	nodeID := nodeIDFrom(0x08)
	cm := newReconcileCM(t, nodeID)
	cm.deployments["dep-with-res"] = &Deployment{
		ID:            "dep-with-res",
		Status:        types.ContainerStatusStopped,
		OriginatorID:  nodeID,
		ReservationID: "12345",
	}

	fin := newMockFinalizer(1)
	cm.reconcileEscrowsOnStartup(context.Background(), fin)

	if !waitFor(t, func() bool { return fin.callCount() == 1 }) {
		t.Fatalf("expected FinalizeJob called once, got %d", fin.callCount())
	}
	// The persisted reservation ID must have been rehydrated for the right job.
	wantJob := payment.JobIDFromString("dep-with-res")
	fin.mu.Lock()
	got, ok := fin.registered[wantJob]
	fin.mu.Unlock()
	if !ok {
		t.Fatal("expected RegisterExternalReservation to be called for the orphan")
	}
	if got.Cmp(big.NewInt(12345)) != 0 {
		t.Fatalf("rehydrated reservation = %s, want 12345", got)
	}
}

// TestReconcileEscrowsOnStartup_RealEscrowEndToEnd drives the reconciler through
// a REAL *payment.PaymentService (mock mode) — not a finalizer mock that
// bypasses the reservation cache. It reproduces the crash-recovery sequence:
// create escrow, persist the on-chain reservation ID, clear the in-memory cache
// (simulating a restart), then reconcile. The reconciler must rehydrate the
// cache from the persisted ID and finalize successfully.
func TestReconcileEscrowsOnStartup_RealEscrowEndToEnd(t *testing.T) {
	svc, err := payment.NewPaymentService(&payment.PaymentServiceConfig{MockMode: true})
	if err != nil {
		t.Fatalf("new mock payment service: %v", err)
	}

	ctx := context.Background()
	const depID = "dep-e2e"
	jobID := payment.JobIDFromString(depID)

	// Create the escrow as the deploy path would, then read back + persist the
	// assigned reservation ID.
	if err := svc.CreateJobEscrow(ctx, jobID, common.Address{}, big.NewInt(1_000_000), time.Hour); err != nil {
		t.Fatalf("create escrow: %v", err)
	}
	resID, ok := svc.ReservationIDForJob(jobID)
	if !ok {
		t.Fatal("expected a reservation ID after CreateJobEscrow")
	}

	nodeID := nodeIDFrom(0x09)
	cm := newReconcileCM(t, nodeID)
	cm.deployments[depID] = &Deployment{
		ID:            depID,
		Status:        types.ContainerStatusStopped,
		OriginatorID:  nodeID,
		ReservationID: resID.String(), // persisted at create time, survives restart
	}

	// Simulate a daemon restart: the in-memory jobID->reservationID cache is gone.
	svc.InvalidateEscrowReservation(jobID)
	if _, stillCached := svc.ReservationIDForJob(jobID); stillCached {
		t.Fatal("precondition: cache should be empty after simulated restart")
	}

	// Wrap the real service so we can confirm rehydration ran via the real path.
	spy := &reservationSpy{PaymentService: svc}
	cm.reconcileEscrowsOnStartup(ctx, spy)

	if !waitFor(t, func() bool {
		cm.mu.RLock()
		defer cm.mu.RUnlock()
		return cm.deployments[depID].EscrowFinalized
	}) {
		t.Fatal("expected EscrowFinalized=true after real end-to-end reconcile")
	}
	// The cache was rehydrated through the real EscrowContract.
	if spy.registerCount() != 1 {
		t.Fatalf("expected exactly one RegisterExternalReservation, got %d", spy.registerCount())
	}
}

// reservationSpy embeds a real *payment.PaymentService and counts
// RegisterExternalReservation calls while delegating to the real path.
type reservationSpy struct {
	*payment.PaymentService
	mu        sync.Mutex
	registers int
}

func (s *reservationSpy) RegisterExternalReservation(jobID [32]byte, reservationID *big.Int) {
	s.mu.Lock()
	s.registers++
	s.mu.Unlock()
	s.PaymentService.RegisterExternalReservation(jobID, reservationID)
}

func (s *reservationSpy) registerCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.registers
}
