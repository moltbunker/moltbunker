package payment

import (
	"context"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
)

const (
	eventBackfillBlocks = 100  // blocks to backfill on reconnect
	eventReconnectBase  = 2 * time.Second
	eventReconnectMax   = 60 * time.Second
	eventChannelBuffer  = 64
)

// EventWatcher manages WebSocket event subscriptions with automatic reconnection.
// It watches for Staked/Unstaked, the four escrow lifecycle events
// (ReservationCreated, PaymentReleased, Refunded, ReservationFinalized), and
// Slashed events, forwarding them to consumer channels for cache invalidation
// and state sync. All four escrow events share the single escrowEvents channel
// and are distinguished by EscrowEvent.Kind.
type EventWatcher struct {
	baseClient       *BaseClient
	stakingContract  *StakingContract
	escrowContract   *EscrowContract
	slashingContract *SlashingContract

	stakeEvents  chan *StakeEvent
	escrowEvents chan *EscrowEvent
	slashEvents  chan *SlashEvent

	lastBlock atomic.Uint64
	running   atomic.Bool
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewEventWatcher creates a new event watcher.
func NewEventWatcher(
	bc *BaseClient,
	sc *StakingContract,
	ec *EscrowContract,
	slc *SlashingContract,
) *EventWatcher {
	return &EventWatcher{
		baseClient:       bc,
		stakingContract:  sc,
		escrowContract:   ec,
		slashingContract: slc,
		stakeEvents:      make(chan *StakeEvent, eventChannelBuffer),
		escrowEvents:     make(chan *EscrowEvent, eventChannelBuffer),
		slashEvents:      make(chan *SlashEvent, eventChannelBuffer),
	}
}

// Start begins watching for on-chain events via WebSocket subscriptions.
// Each subscription runs in its own goroutine with automatic reconnection.
func (ew *EventWatcher) Start(ctx context.Context) error {
	if ew.running.Load() {
		return nil
	}

	if ew.baseClient == nil || !ew.baseClient.IsConnected() {
		logging.Info("event watcher: no chain connection, skipping")
		return nil
	}

	if !ew.baseClient.HasWSConfig() {
		logging.Info("event watcher: no WebSocket endpoint configured, event subscriptions disabled (using RPC polling)")
		return nil
	}

	// Snapshot current block for backfill baseline
	blockNum, err := ew.baseClient.GetBlockNumber(ctx)
	if err == nil {
		ew.lastBlock.Store(blockNum)
	}

	ctx, ew.cancel = context.WithCancel(ctx)
	ew.running.Store(true)

	// Launch one goroutine per event subscription. The four escrow lifecycle
	// events each get their own subscription (matching the stake/slash pattern)
	// but all feed the single escrowEvents channel.
	ew.wg.Add(6)
	util.SafeGoWithName("event-watcher-stake", func() {
		defer ew.wg.Done()
		ew.watchStakeEvents(ctx)
	})
	util.SafeGoWithName("event-watcher-escrow-created", func() {
		defer ew.wg.Done()
		ew.watchEscrowCreatedEvents(ctx)
	})
	util.SafeGoWithName("event-watcher-payment-released", func() {
		defer ew.wg.Done()
		ew.watchPaymentReleasedEvents(ctx)
	})
	util.SafeGoWithName("event-watcher-refunded", func() {
		defer ew.wg.Done()
		ew.watchRefundedEvents(ctx)
	})
	util.SafeGoWithName("event-watcher-finalized", func() {
		defer ew.wg.Done()
		ew.watchReservationFinalizedEvents(ctx)
	})
	util.SafeGoWithName("event-watcher-slash", func() {
		defer ew.wg.Done()
		ew.watchSlashEvents(ctx)
	})

	logging.Info("event watcher started", "block", ew.lastBlock.Load())
	return nil
}

// Stop stops the event watcher and waits for goroutines to exit.
// After Stop returns, all event channels are closed so consumers
// using range will unblock.
func (ew *EventWatcher) Stop() {
	if !ew.running.Load() {
		return
	}
	ew.cancel()
	ew.wg.Wait()
	ew.running.Store(false)

	// Close channels so consumers ranging over them will unblock.
	// Safe because all writer goroutines have exited (wg.Wait above).
	close(ew.stakeEvents)
	close(ew.escrowEvents)
	close(ew.slashEvents)

	logging.Info("event watcher stopped")
}

// StakeEvents returns the channel for receiving stake events.
func (ew *EventWatcher) StakeEvents() <-chan *StakeEvent {
	return ew.stakeEvents
}

// EscrowEvents returns the channel for receiving all escrow lifecycle events.
// Consumers switch on EscrowEvent.Kind to handle Created/PaymentReleased/
// Refunded/Finalized.
func (ew *EventWatcher) EscrowEvents() <-chan *EscrowEvent {
	return ew.escrowEvents
}

// SlashEvents returns the channel for receiving slash events.
func (ew *EventWatcher) SlashEvents() <-chan *SlashEvent {
	return ew.slashEvents
}

// watchStakeEvents subscribes to Staked/Unstaked events with reconnection.
func (ew *EventWatcher) watchStakeEvents(ctx context.Context) {
	topics := []common.Hash{}
	if ev, ok := ew.stakingContract.contractABI.Events["Staked"]; ok {
		topics = append(topics, ev.ID)
	}
	if ev, ok := ew.stakingContract.contractABI.Events["Unstaked"]; ok {
		topics = append(topics, ev.ID)
	}
	if len(topics) == 0 {
		logging.Warn("event watcher: no stake event topics found in ABI")
		return
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{ew.stakingContract.contractAddr},
		Topics:    [][]common.Hash{topics},
	}

	ew.subscribeWithReconnect(ctx, "stake", query, func(log ethtypes.Log) {
		event, err := ew.stakingContract.parseStakeEvent(log)
		if err != nil {
			return
		}
		select {
		case ew.stakeEvents <- event:
		default:
			logging.Warn("event watcher: stake event channel full, dropping")
		}
	})
}

// emitEscrowEvent forwards an escrow event to the shared channel, dropping it
// (with a warning) if the consumer is too slow and the buffer is full.
func (ew *EventWatcher) emitEscrowEvent(ev *EscrowEvent) {
	select {
	case ew.escrowEvents <- ev:
	default:
		logging.Warn("event watcher: escrow event channel full, dropping",
			"kind", ev.Kind.String())
	}
}

// watchSingleEvent resolves the topic ID for the named escrow event, builds a
// filter query for the escrow contract, and subscribes with reconnection. The
// handler is invoked for each matching log and is responsible for parsing and
// emitting an EscrowEvent of the given kind. Centralizes the repetitive
// topic-lookup + query-build + subscribe wiring shared by the four escrow
// watchers.
func (ew *EventWatcher) watchSingleEvent(
	ctx context.Context,
	name string,
	kind EscrowEventKind,
	handler func(ethtypes.Log) *EscrowEvent,
) {
	ev, ok := ew.escrowContract.contractABI.Events[name]
	if !ok {
		logging.Warn("event watcher: escrow event not found in ABI", "event", name)
		return
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{ew.escrowContract.contractAddr},
		Topics:    [][]common.Hash{{ev.ID}},
	}

	ew.subscribeWithReconnect(ctx, "escrow-"+kind.String(), query, func(log ethtypes.Log) {
		if out := handler(log); out != nil {
			ew.emitEscrowEvent(out)
		}
	})
}

// watchEscrowCreatedEvents subscribes to ReservationCreated events.
func (ew *EventWatcher) watchEscrowCreatedEvents(ctx context.Context) {
	ew.watchSingleEvent(ctx, "ReservationCreated", EscrowEventCreated, parseEscrowCreatedLog)
}

// watchPaymentReleasedEvents subscribes to PaymentReleased events.
func (ew *EventWatcher) watchPaymentReleasedEvents(ctx context.Context) {
	ew.watchSingleEvent(ctx, "PaymentReleased", EscrowEventPaymentReleased, parsePaymentReleasedLog)
}

// watchRefundedEvents subscribes to Refunded events.
func (ew *EventWatcher) watchRefundedEvents(ctx context.Context) {
	ew.watchSingleEvent(ctx, "Refunded", EscrowEventRefunded, parseRefundedLog)
}

// watchReservationFinalizedEvents subscribes to ReservationFinalized events.
func (ew *EventWatcher) watchReservationFinalizedEvents(ctx context.Context) {
	ew.watchSingleEvent(ctx, "ReservationFinalized", EscrowEventFinalized, parseReservationFinalizedLog)
}

// The four parse functions below own the topic-position → field mapping for the
// escrow lifecycle events. They are package-level (not closures) so the mapping
// can be locked by unit tests against synthetic logs. Each event's indexed
// topics occupy log.Topics[1..] (Topics[0] is the event signature); non-indexed
// fields are packed into 32-byte words in log.Data.

// parseEscrowCreatedLog maps a ReservationCreated log:
// ReservationCreated(uint256 indexed reservationId, address indexed requester,
// uint256 amount, uint256 duration). amount is the first data word.
func parseEscrowCreatedLog(log ethtypes.Log) *EscrowEvent {
	event := &EscrowEvent{Kind: EscrowEventCreated, Timestamp: time.Now()}
	if len(log.Topics) > 1 {
		event.ReservationID = new(big.Int).SetBytes(log.Topics[1].Bytes())
	}
	if len(log.Topics) > 2 {
		event.Requester = common.HexToAddress(log.Topics[2].Hex())
	}
	if len(log.Data) >= 32 {
		event.Amount = new(big.Int).SetBytes(log.Data[:32])
	}
	return event
}

// parsePaymentReleasedLog maps a PaymentReleased log:
// PaymentReleased(uint256 indexed reservationId, uint256 grossAmount, ...).
// grossAmount is the first non-indexed data word.
func parsePaymentReleasedLog(log ethtypes.Log) *EscrowEvent {
	event := &EscrowEvent{Kind: EscrowEventPaymentReleased, Timestamp: time.Now()}
	if len(log.Topics) > 1 {
		event.ReservationID = new(big.Int).SetBytes(log.Topics[1].Bytes())
	}
	if len(log.Data) >= 32 {
		event.Amount = new(big.Int).SetBytes(log.Data[:32])
	}
	return event
}

// parseRefundedLog maps a Refunded log:
// Refunded(uint256 indexed reservationId, address indexed requester,
// uint256 refundAmount). refundAmount is the first (only) non-indexed word.
func parseRefundedLog(log ethtypes.Log) *EscrowEvent {
	event := &EscrowEvent{Kind: EscrowEventRefunded, Timestamp: time.Now()}
	if len(log.Topics) > 1 {
		event.ReservationID = new(big.Int).SetBytes(log.Topics[1].Bytes())
	}
	if len(log.Topics) > 2 {
		event.Requester = common.HexToAddress(log.Topics[2].Hex())
	}
	if len(log.Data) >= 32 {
		event.Amount = new(big.Int).SetBytes(log.Data[:32])
	}
	return event
}

// parseReservationFinalizedLog maps a ReservationFinalized log:
// ReservationFinalized(uint256 indexed reservationId). No data words.
func parseReservationFinalizedLog(log ethtypes.Log) *EscrowEvent {
	event := &EscrowEvent{Kind: EscrowEventFinalized, Timestamp: time.Now()}
	if len(log.Topics) > 1 {
		event.ReservationID = new(big.Int).SetBytes(log.Topics[1].Bytes())
	}
	return event
}

// watchSlashEvents subscribes to Slashed events with reconnection.
func (ew *EventWatcher) watchSlashEvents(ctx context.Context) {
	ev, ok := ew.slashingContract.contractABI.Events["Slashed"]
	if !ok {
		logging.Warn("event watcher: Slashed event not found in ABI")
		return
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{ew.slashingContract.contractAddr},
		Topics:    [][]common.Hash{{ev.ID}},
	}

	ew.subscribeWithReconnect(ctx, "slash", query, func(log ethtypes.Log) {
		event := &SlashEvent{
			Timestamp: time.Now(),
		}
		if len(log.Topics) > 1 {
			event.Provider = common.HexToAddress(log.Topics[1].Hex())
		}
		if len(log.Data) >= 32 {
			event.Amount = new(big.Int).SetBytes(log.Data[:32])
		}
		select {
		case ew.slashEvents <- event:
		default:
			logging.Warn("event watcher: slash event channel full, dropping")
		}
	})
}

// subscribeWithReconnect manages a single event subscription with automatic
// reconnection and backfill on WS failure.
func (ew *EventWatcher) subscribeWithReconnect(
	ctx context.Context,
	name string,
	query ethereum.FilterQuery,
	handler func(ethtypes.Log),
) {
	delay := eventReconnectBase

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wsClient := ew.baseClient.WSClient()
		if wsClient == nil {
			// Try reconnecting WS
			if err := ew.baseClient.ReconnectWS(ctx); err != nil {
				logging.Warn("event watcher: WS reconnect failed",
					"subscription", name, "error", err)
				if !ew.sleepOrDone(ctx, delay) {
					return
				}
				delay = ew.nextDelay(delay)
				continue
			}
			wsClient = ew.baseClient.WSClient()
			if wsClient == nil {
				if !ew.sleepOrDone(ctx, delay) {
					return
				}
				delay = ew.nextDelay(delay)
				continue
			}
		}

		// Backfill events from last known block
		ew.backfillEvents(ctx, query, handler)

		// Subscribe
		logs := make(chan ethtypes.Log, 16)
		sub, err := wsClient.SubscribeFilterLogs(ctx, query, logs)
		if err != nil {
			logging.Warn("event watcher: subscribe failed",
				"subscription", name, "error", err)
			if !ew.sleepOrDone(ctx, delay) {
				return
			}
			delay = ew.nextDelay(delay)
			continue
		}

		// Reset delay on successful subscription
		delay = eventReconnectBase
		logging.Info("event watcher: subscribed", "subscription", name)

		// Process events until error or context cancellation
		done := ew.processEvents(ctx, name, sub, logs, handler)
		sub.Unsubscribe()
		if done {
			return
		}

		// Subscription dropped — reconnect WS and retry
		_ = ew.baseClient.ReconnectWS(ctx)
	}
}

// processEvents reads events from the subscription until an error or context done.
// Returns true if context was cancelled (should stop), false if subscription errored
// (should reconnect).
func (ew *EventWatcher) processEvents(
	ctx context.Context,
	name string,
	sub ethereum.Subscription,
	logs <-chan ethtypes.Log,
	handler func(ethtypes.Log),
) bool {
	for {
		select {
		case <-ctx.Done():
			return true
		case err := <-sub.Err():
			if err != nil {
				logging.Warn("event watcher: subscription error",
					"subscription", name, "error", err)
			}
			return false
		case log := <-logs:
			// Track latest block for backfill
			if log.BlockNumber > ew.lastBlock.Load() {
				ew.lastBlock.Store(log.BlockNumber)
			}
			handler(log)
		}
	}
}

// backfillEvents queries historical logs from lastBlock to catch events
// missed during a subscription gap.
func (ew *EventWatcher) backfillEvents(
	ctx context.Context,
	query ethereum.FilterQuery,
	handler func(ethtypes.Log),
) {
	last := ew.lastBlock.Load()
	if last == 0 {
		return
	}

	client := ew.baseClient.Client()
	if client == nil {
		return
	}

	// Backfill from last known block
	fromBlock := last
	if fromBlock > eventBackfillBlocks {
		fromBlock = last - eventBackfillBlocks
	}

	backfillQuery := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		Addresses: query.Addresses,
		Topics:    query.Topics,
	}

	logs, err := client.FilterLogs(ctx, backfillQuery)
	if err != nil {
		logging.Warn("event watcher: backfill failed", "error", err)
		return
	}

	for _, log := range logs {
		if log.BlockNumber > last {
			handler(log)
			if log.BlockNumber > ew.lastBlock.Load() {
				ew.lastBlock.Store(log.BlockNumber)
			}
		}
	}

	if len(logs) > 0 {
		logging.Info("event watcher: backfilled events",
			"count", len(logs), "from_block", fromBlock)
	}
}

// sleepOrDone sleeps for the given duration or returns false if context is cancelled.
func (ew *EventWatcher) sleepOrDone(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// nextDelay calculates the next exponential backoff delay.
func (ew *EventWatcher) nextDelay(current time.Duration) time.Duration {
	next := current * 2
	if next > eventReconnectMax {
		next = eventReconnectMax
	}
	return next
}
