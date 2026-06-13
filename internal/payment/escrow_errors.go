package payment

import (
	"errors"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// escrowABIOnce lazily parses the escrow ABI once for custom-error decoding.
var (
	escrowABIOnce   sync.Once
	escrowABIParsed abi.ABI
	escrowABIErr    error
)

func escrowABI() (abi.ABI, error) {
	escrowABIOnce.Do(func() {
		escrowABIParsed, escrowABIErr = abi.JSON(strings.NewReader(EscrowContractABI))
	})
	return escrowABIParsed, escrowABIErr
}

// IsAlreadyTerminalEscrowErr reports whether a finalize/refund error indicates
// the on-chain reservation has ALREADY reached a terminal state (Completed,
// Refunded, or Disputed) — i.e. re-finalizing is a confirmed no-op and the
// orphan reconciler may mark the deployment done.
//
// It is intentionally precise. The only terminal signal the BunkerEscrow
// contract emits for a finalize/refund against a terminal reservation is the
// custom error InvalidStatus(reservationId, expected, actual) where `actual`
// is one of the terminal Status values. This function prefers decoding that
// typed custom-error data (via rpc.DataError) over any substring matching.
//
// Crucially it does NOT match ambiguous transient strings. In particular an
// error like "reservation not yet finalizable" — which contains the substring
// "finalized" — must be treated as TRANSIENT (return false) so a live escrow is
// never abandoned by the reconciler.
func IsAlreadyTerminalEscrowErr(err error) bool {
	if err == nil {
		return false
	}

	// 1) Preferred path: typed custom-error data. go-ethereum surfaces a revert
	//    during gas estimation / eth_call as an rpc.DataError carrying the ABI-
	//    encoded custom-error bytes. Decode and inspect the actual status.
	if data, ok := extractRevertData(err); ok {
		if terminal, decoded := isTerminalStatusRevert(data); decoded {
			return terminal
		}
	}

	// 2) Narrow string fallback for nodes/providers that only relay the decoded
	//    revert reason as a message. Match ONLY the precise, unambiguous terminal
	//    phrasings — never the bare "finalized"/"not active"/"invalid state"
	//    catch-alls, which collide with transient messages like
	//    "reservation not yet finalizable".
	msg := strings.ToLower(err.Error())
	for _, precise := range []string{
		"reservation already finalized",
		"reservation already completed",
		"reservation already refunded",
		"already finalized",
		"invalidstatus", // raw custom-error name some nodes echo verbatim
	} {
		if strings.Contains(msg, precise) {
			return true
		}
	}
	return false
}

// extractRevertData pulls the raw ABI-encoded revert/custom-error bytes from an
// error if it implements rpc.DataError. The ErrorData() value is typically a
// 0x-prefixed hex string.
func extractRevertData(err error) ([]byte, bool) {
	var de rpc.DataError
	if !errors.As(err, &de) {
		return nil, false
	}
	raw := de.ErrorData()
	switch v := raw.(type) {
	case string:
		b, decErr := hexutil.Decode(v)
		if decErr != nil || len(b) < 4 {
			return nil, false
		}
		return b, true
	case []byte:
		if len(v) < 4 {
			return nil, false
		}
		return v, true
	default:
		return nil, false
	}
}

// isTerminalStatusRevert decodes ABI-encoded custom-error data against the
// escrow ABI. It returns (terminal, true) when the data decodes to an
// InvalidStatus error whose actual status is terminal (Completed/Refunded/
// Disputed). The second return is false when the data is not a recognizable
// escrow custom error (caller should fall through to string matching).
func isTerminalStatusRevert(data []byte) (terminal bool, decoded bool) {
	parsed, err := escrowABI()
	if err != nil || len(data) < 4 {
		return false, false
	}
	var selector [4]byte
	copy(selector[:], data[:4])

	abiErr, err := parsed.ErrorByID(selector)
	if err != nil {
		return false, false
	}

	switch abiErr.Name {
	case "InvalidStatus":
		vals, uErr := abiErr.Unpack(data)
		if uErr != nil {
			return false, false
		}
		args, ok := vals.([]interface{})
		// InvalidStatus(reservationId uint256, expected uint8, actual uint8)
		if !ok || len(args) != 3 {
			// Decoded as InvalidStatus but shape unexpected: it IS a status
			// revert, but we cannot read `actual`. Be conservative — treat as
			// transient (do not abandon the escrow).
			return false, true
		}
		actual, ok := args[2].(uint8)
		if !ok {
			return false, true
		}
		switch EscrowState(actual) {
		case EscrowStateCompleted, EscrowStateRefunded, EscrowStateDisputed:
			return true, true
		default:
			// e.g. actual == Created: the reservation exists but is NOT yet
			// finalizable. Transient from the reconciler's perspective.
			return false, true
		}
	case "InvalidReservation":
		// Reservation does not exist on-chain (never created or wrong ID).
		// Not "already finalized"; leave it for retry/operator inspection.
		return false, true
	default:
		return false, false
	}
}
