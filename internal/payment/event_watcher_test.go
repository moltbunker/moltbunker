package payment

import (
	"context"
	"testing"
	"time"
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
