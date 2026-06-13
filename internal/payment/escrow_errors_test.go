package payment

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// newRealEscrowForTest builds a non-mock EscrowContract with initialized caches
// but no base client. This exercises the REAL getReservationID gate (the code
// path that the orphan reconciler depends on) without needing a live chain.
func newRealEscrowForTest() *EscrowContract {
	return &EscrowContract{
		mockMode: false,
		// A zero-value base client (no private key) lets calls that pass the
		// reservation gate fail cleanly with "no private key configured" instead
		// of nil-dereferencing.
		baseClient:       &BaseClient{},
		reservationIDs:   make(map[[32]byte]*big.Int),
		reservationTimes: make(map[[32]byte]time.Time),
	}
}

// TestFinalizeEscrow_NoReservationIDAfterRestart documents the HIGH-1 bug: with
// an empty in-memory cache (the state after a daemon restart), the REAL escrow
// contract's FinalizeEscrow cannot resolve a reservation ID and fails before
// ever touching the chain. This is the failure the persisted Deployment
// .ReservationID + reconciler rehydration fixes.
func TestFinalizeEscrow_NoReservationIDAfterRestart(t *testing.T) {
	ec := newRealEscrowForTest()
	jobID := JobIDFromString("dep-orphan")

	_, err := ec.FinalizeEscrow(context.Background(), jobID)
	if err == nil {
		t.Fatal("expected finalize to fail with empty reservation cache")
	}
	if !strings.Contains(err.Error(), "no reservation ID found") {
		t.Fatalf("expected 'no reservation ID found', got %v", err)
	}
}

// TestReservationIDForJob_RoundTrip verifies the getter used by the deploy path
// to persist the reservation ID, and that rehydration via
// StoreExternalReservationID restores it after a simulated restart.
func TestReservationIDForJob_RoundTrip(t *testing.T) {
	ec := newRealEscrowForTest()
	jobID := JobIDFromString("dep-1")

	if _, ok := ec.ReservationIDForJob(jobID); ok {
		t.Fatal("expected no reservation before store")
	}

	want := big.NewInt(4242)
	ec.StoreExternalReservationID(jobID, want)

	got, ok := ec.ReservationIDForJob(jobID)
	if !ok {
		t.Fatal("expected reservation after store")
	}
	if got.Cmp(want) != 0 {
		t.Fatalf("reservation = %s, want %s", got, want)
	}

	// Simulate restart: fresh contract, empty cache.
	restarted := newRealEscrowForTest()
	if _, ok := restarted.ReservationIDForJob(jobID); ok {
		t.Fatal("fresh contract must start with empty cache")
	}
	// Rehydrate from the persisted decimal string.
	rehydrated := new(big.Int)
	rehydrated.SetString(got.String(), 10)
	restarted.StoreExternalReservationID(jobID, rehydrated)

	if back, ok := restarted.ReservationIDForJob(jobID); !ok || back.Cmp(want) != 0 {
		t.Fatalf("rehydrated reservation = %v ok=%v, want %s", back, ok, want)
	}
	// And FinalizeEscrow no longer trips the "no reservation ID" guard (it now
	// proceeds past getReservationID; the nil baseClient causes a later, distinct
	// failure — which proves the reservation gate is satisfied).
	if _, err := restarted.FinalizeEscrow(context.Background(), jobID); err != nil {
		if strings.Contains(err.Error(), "no reservation ID found") {
			t.Fatalf("reservation gate should be satisfied after rehydration, got %v", err)
		}
	}
}

// --- HIGH-2: IsAlreadyTerminalEscrowErr precision ---

func TestIsAlreadyTerminalEscrowErr_NilAndTransient(t *testing.T) {
	if IsAlreadyTerminalEscrowErr(nil) {
		t.Error("nil error must not be terminal")
	}

	// The headline regression: a message containing the substring "finalized"
	// but meaning the OPPOSITE must be treated as TRANSIENT.
	transient := []string{
		"execution reverted: reservation not yet finalizable",
		"dial tcp: connection refused",
		"context deadline exceeded",
		"reservation not active yet",    // collides with old "not active" catch-all
		"invalid state transition wait", // collides with old "invalid state" catch-all
		"escrow not yet finalized, retry later",
	}
	for _, m := range transient {
		if IsAlreadyTerminalEscrowErr(errors.New(m)) {
			t.Errorf("error %q must be treated as TRANSIENT (not terminal)", m)
		}
	}
}

func TestIsAlreadyTerminalEscrowErr_PreciseStrings(t *testing.T) {
	terminal := []string{
		"execution reverted: reservation already finalized",
		"reservation already completed",
		"reservation already refunded",
		"execution reverted: InvalidStatus(5, 2, 3)",
	}
	for _, m := range terminal {
		if !IsAlreadyTerminalEscrowErr(errors.New(m)) {
			t.Errorf("error %q must be treated as terminal", m)
		}
	}
}

// dataErr is a test rpc.DataError carrying ABI-encoded custom-error bytes.
type dataErr struct {
	msg  string
	data string
}

func (d *dataErr) Error() string          { return d.msg }
func (d *dataErr) ErrorCode() int         { return 3 } // execution reverted
func (d *dataErr) ErrorData() interface{} { return d.data }

// packInvalidStatus ABI-encodes InvalidStatus(reservationId, expected, actual).
func packInvalidStatus(t *testing.T, reservationID int64, expected, actual uint8) string {
	t.Helper()
	parsed, err := abi.JSON(strings.NewReader(EscrowContractABI))
	if err != nil {
		t.Fatalf("parse ABI: %v", err)
	}
	abiErr, ok := parsed.Errors["InvalidStatus"]
	if !ok {
		t.Fatal("InvalidStatus not in ABI")
	}
	args, err := abiErr.Inputs.Pack(big.NewInt(reservationID), expected, actual)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	out := append(abiErr.ID.Bytes()[:4], args...)
	return "0x" + hexEncode(out)
}

func hexEncode(b []byte) string {
	const hexchars = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, c := range b {
		out[i*2] = hexchars[c>>4]
		out[i*2+1] = hexchars[c&0x0f]
	}
	return string(out)
}

func TestIsAlreadyTerminalEscrowErr_TypedCustomError(t *testing.T) {
	// expected=Active(2). actual values per Status enum:
	// None=0, Created=1, Active=2, Completed=3, Refunded=4, Disputed=5.
	cases := []struct {
		name   string
		actual uint8
		want   bool
	}{
		{"completed-is-terminal", uint8(EscrowStateCompleted), true},
		{"refunded-is-terminal", uint8(EscrowStateRefunded), true},
		{"disputed-is-terminal", uint8(EscrowStateDisputed), true},
		{"created-is-transient", uint8(EscrowStateCreated), false},
		{"none-is-transient", uint8(EscrowStateNone), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			de := &dataErr{
				msg:  "execution reverted",
				data: packInvalidStatus(t, 7, uint8(EscrowStateActive), tc.actual),
			}
			if got := IsAlreadyTerminalEscrowErr(de); got != tc.want {
				t.Errorf("actual=%d: got terminal=%v, want %v", tc.actual, got, tc.want)
			}
		})
	}
}
