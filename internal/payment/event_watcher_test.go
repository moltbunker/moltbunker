package payment

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func TestEventWatcher_NewAndChannels(t *testing.T) {
	ew := NewEventWatcher(nil, nil, nil, nil)
	if ew == nil {
		t.Fatal("expected non-nil event watcher")
	}
	if cap(ew.stakeEvents) != eventChannelBuffer {
		t.Errorf("expected stake channel buffer %d, got %d", eventChannelBuffer, cap(ew.stakeEvents))
	}
	if cap(ew.escrowEvents) != eventChannelBuffer {
		t.Errorf("expected escrow channel buffer %d, got %d", eventChannelBuffer, cap(ew.escrowEvents))
	}
	if cap(ew.slashEvents) != eventChannelBuffer {
		t.Errorf("expected slash channel buffer %d, got %d", eventChannelBuffer, cap(ew.slashEvents))
	}

	// Channels should be readable
	_ = ew.StakeEvents()
	_ = ew.EscrowEvents()
	_ = ew.SlashEvents()
}

func TestEventWatcher_StartNoConnection(t *testing.T) {
	ew := NewEventWatcher(nil, nil, nil, nil)

	ctx := context.Background()
	err := ew.Start(ctx)
	if err != nil {
		t.Fatalf("expected nil error for nil base client, got %v", err)
	}

	// Should not be running since there's no connection
	if ew.running.Load() {
		t.Error("expected not running with nil base client")
	}
}

func TestEventWatcher_StopIdempotent(t *testing.T) {
	ew := NewEventWatcher(nil, nil, nil, nil)

	// Stop when not started should be a no-op
	ew.Stop()
	ew.Stop()
}

func TestEventWatcher_NextDelay(t *testing.T) {
	ew := NewEventWatcher(nil, nil, nil, nil)

	d := eventReconnectBase
	d = ew.nextDelay(d)
	if d != eventReconnectBase*2 {
		t.Errorf("expected %v, got %v", eventReconnectBase*2, d)
	}

	// Should cap at max
	d = eventReconnectMax
	d = ew.nextDelay(d)
	if d != eventReconnectMax {
		t.Errorf("expected %v (capped), got %v", eventReconnectMax, d)
	}
}

func TestEventWatcher_SleepOrDone(t *testing.T) {
	ew := NewEventWatcher(nil, nil, nil, nil)

	// Cancelled context should return false immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := ew.sleepOrDone(ctx, 1*time.Hour)
	if result {
		t.Error("expected false for cancelled context")
	}

	// Short sleep should return true
	ctx2 := context.Background()
	result = ew.sleepOrDone(ctx2, 1*time.Millisecond)
	if !result {
		t.Error("expected true for short sleep")
	}
}

func TestEscrowEventKinds(t *testing.T) {
	ew := NewEventWatcher(nil, nil, nil, nil)

	// EscrowEvents must be a receive channel of the discriminated union type.
	var ch <-chan *EscrowEvent = ew.EscrowEvents()
	if ch == nil {
		t.Fatal("expected non-nil escrow event channel")
	}

	// All four kinds must stringify distinctly (guards the const block).
	kinds := map[EscrowEventKind]string{
		EscrowEventCreated:         "created",
		EscrowEventPaymentReleased: "payment_released",
		EscrowEventRefunded:        "refunded",
		EscrowEventFinalized:       "finalized",
	}
	seen := make(map[string]bool)
	for k, want := range kinds {
		if got := k.String(); got != want {
			t.Errorf("kind %d: want %q, got %q", k, want, got)
		}
		if seen[want] {
			t.Errorf("duplicate kind string %q", want)
		}
		seen[want] = true
	}
}

func TestWatchEscrowAllTopics(t *testing.T) {
	// A mock escrow contract parses the real ABI, so all four escrow event
	// topic IDs must be resolvable. Confirm each is present in the ABI and that
	// the four watcher entry points run without panicking against a cancelled
	// context (they short-circuit because there is no WS client).
	ec := NewMockEscrowContract()
	for _, name := range []string{
		"ReservationCreated",
		"PaymentReleased",
		"Refunded",
		"ReservationFinalized",
	} {
		ev, ok := ec.contractABI.Events[name]
		if !ok {
			t.Fatalf("escrow ABI missing event %q", name)
		}
		if ev.ID == [32]byte{} {
			t.Errorf("event %q has zero topic ID", name)
		}
	}

	ew := NewEventWatcher(nil, nil, ec, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // make subscribeWithReconnect return immediately

	// None of these should panic; with a nil base client + cancelled context
	// they return promptly.
	ew.watchEscrowCreatedEvents(ctx)
	ew.watchPaymentReleasedEvents(ctx)
	ew.watchRefundedEvents(ctx)
	ew.watchReservationFinalizedEvents(ctx)
}

// uint256Topic encodes a uint256 as a 32-byte indexed topic.
func uint256Topic(v int64) common.Hash {
	return common.BigToHash(big.NewInt(v))
}

// addressTopic encodes an address as a 32-byte indexed topic (left-padded).
func addressTopic(a common.Address) common.Hash {
	return common.BytesToHash(a.Bytes())
}

// nonIndexedData ABI-packs the non-indexed inputs of the named escrow event.
func nonIndexedData(t *testing.T, escrowABI abi.ABI, eventName string, vals ...interface{}) []byte {
	t.Helper()
	ev, ok := escrowABI.Events[eventName]
	if !ok {
		t.Fatalf("event %q not in ABI", eventName)
	}
	var nonIndexed abi.Arguments
	for _, in := range ev.Inputs {
		if !in.Indexed {
			nonIndexed = append(nonIndexed, in)
		}
	}
	data, err := nonIndexed.Pack(vals...)
	if err != nil {
		t.Fatalf("pack %s data: %v", eventName, err)
	}
	return data
}

// TestParseEscrowEventLogs (LOW-1) packs a representative log for each of the
// four escrow events — with realistic indexed topics and ABI-encoded data
// words — and asserts the parsed EscrowEvent fields. This locks the
// topic-position → field mapping so a future ABI/handler edit that shifts a
// topic index is caught.
func TestParseEscrowEventLogs(t *testing.T) {
	escrowABI, err := abi.JSON(strings.NewReader(EscrowContractABI))
	if err != nil {
		t.Fatalf("parse escrow ABI: %v", err)
	}

	requester := common.HexToAddress("0x00000000000000000000000000000000000000aB")

	t.Run("ReservationCreated", func(t *testing.T) {
		// indexed: reservationId(1), requester(2); data: amount, duration.
		log := ethtypes.Log{
			Topics: []common.Hash{
				escrowABI.Events["ReservationCreated"].ID,
				uint256Topic(101),
				addressTopic(requester),
			},
			Data: nonIndexedData(t, escrowABI, "ReservationCreated", big.NewInt(5000), big.NewInt(3600)),
		}
		ev := parseEscrowCreatedLog(log)
		if ev.Kind != EscrowEventCreated {
			t.Errorf("kind = %v, want Created", ev.Kind)
		}
		if ev.ReservationID == nil || ev.ReservationID.Int64() != 101 {
			t.Errorf("reservationID = %v, want 101", ev.ReservationID)
		}
		if ev.Requester != requester {
			t.Errorf("requester = %s, want %s", ev.Requester, requester)
		}
		if ev.Amount == nil || ev.Amount.Int64() != 5000 {
			t.Errorf("amount = %v, want 5000", ev.Amount)
		}
	})

	t.Run("PaymentReleased", func(t *testing.T) {
		// indexed: reservationId(1); data: grossAmount(first word), ...
		log := ethtypes.Log{
			Topics: []common.Hash{
				escrowABI.Events["PaymentReleased"].ID,
				uint256Topic(202),
			},
			Data: nonIndexedData(t, escrowABI, "PaymentReleased",
				big.NewInt(9000), // grossAmount
				big.NewInt(8000), // netToProviders
				big.NewInt(1000), // protocolFee
				big.NewInt(400),  // burnedAmount
				big.NewInt(600),  // treasuryAmount
			),
		}
		ev := parsePaymentReleasedLog(log)
		if ev.Kind != EscrowEventPaymentReleased {
			t.Errorf("kind = %v, want PaymentReleased", ev.Kind)
		}
		if ev.ReservationID == nil || ev.ReservationID.Int64() != 202 {
			t.Errorf("reservationID = %v, want 202", ev.ReservationID)
		}
		// Amount must be the FIRST data word (grossAmount), not a later one.
		if ev.Amount == nil || ev.Amount.Int64() != 9000 {
			t.Errorf("amount = %v, want 9000 (grossAmount)", ev.Amount)
		}
		// PaymentReleased has no indexed requester.
		if ev.Requester != (common.Address{}) {
			t.Errorf("requester should be zero for PaymentReleased, got %s", ev.Requester)
		}
	})

	t.Run("Refunded", func(t *testing.T) {
		// indexed: reservationId(1), requester(2); data: refundAmount.
		log := ethtypes.Log{
			Topics: []common.Hash{
				escrowABI.Events["Refunded"].ID,
				uint256Topic(303),
				addressTopic(requester),
			},
			Data: nonIndexedData(t, escrowABI, "Refunded", big.NewInt(7777)),
		}
		ev := parseRefundedLog(log)
		if ev.Kind != EscrowEventRefunded {
			t.Errorf("kind = %v, want Refunded", ev.Kind)
		}
		if ev.ReservationID == nil || ev.ReservationID.Int64() != 303 {
			t.Errorf("reservationID = %v, want 303", ev.ReservationID)
		}
		if ev.Requester != requester {
			t.Errorf("requester = %s, want %s", ev.Requester, requester)
		}
		if ev.Amount == nil || ev.Amount.Int64() != 7777 {
			t.Errorf("amount = %v, want 7777 (refundAmount)", ev.Amount)
		}
	})

	t.Run("ReservationFinalized", func(t *testing.T) {
		// indexed: reservationId(1); no data words.
		log := ethtypes.Log{
			Topics: []common.Hash{
				escrowABI.Events["ReservationFinalized"].ID,
				uint256Topic(404),
			},
		}
		ev := parseReservationFinalizedLog(log)
		if ev.Kind != EscrowEventFinalized {
			t.Errorf("kind = %v, want Finalized", ev.Kind)
		}
		if ev.ReservationID == nil || ev.ReservationID.Int64() != 404 {
			t.Errorf("reservationID = %v, want 404", ev.ReservationID)
		}
		// No data => Amount stays nil; no requester topic => zero.
		if ev.Amount != nil {
			t.Errorf("amount should be nil for Finalized, got %v", ev.Amount)
		}
		if ev.Requester != (common.Address{}) {
			t.Errorf("requester should be zero for Finalized, got %s", ev.Requester)
		}
	})
}

func TestEventWatcher_StartStopWithMockContracts(t *testing.T) {
	// Verify that Start+Stop on mock contracts doesn't panic
	sc := NewMockStakingContract()
	ec := NewMockEscrowContract()
	slc := NewMockSlashingContract()

	ew := NewEventWatcher(nil, sc, ec, slc)

	ctx := context.Background()
	err := ew.Start(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No base client means it should skip
	if ew.running.Load() {
		t.Error("expected not running with nil base client")
	}

	ew.Stop()
}
