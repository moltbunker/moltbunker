//go:build e2e

package golden

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// anvilRPCEnv is the env var that opts a developer / future CI job into the
// real on-chain leg. When unset, TestGoldenPath_EscrowAnvilMode skips cleanly so
// the default darwin/linux suite never needs a chain.
const anvilRPCEnv = "MOLTBUNKER_ANVIL_RPC"

// TestGoldenPath_EscrowMockMode exercises the full escrow settlement cycle
// (CreateEscrow -> SelectProviders -> ReleasePayment -> FinalizeEscrow) against
// the in-memory MockEscrowContract. This runs in every CI environment without a
// chain. [MOCK: in-memory MockEscrowContract]
func TestGoldenPath_EscrowMockMode(t *testing.T) {
	a := testutil.NewAssertions(t)
	t.Log("[MOCK: escrow contract] full create -> select -> release -> finalize cycle")

	ec := payment.NewMockEscrowContract()
	a.True(ec.IsMockMode(), "should be in mock mode")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	jobID := payment.JobIDFromString("golden-escrow-mock-001")
	provider := common.HexToAddress("0x1111111111111111111111111111111111111111")
	amount := BunkerToWei(100)
	durationSecs := big.NewInt(3600)

	_, err := ec.CreateEscrow(ctx, jobID, provider, amount, durationSecs)
	a.NoError(err, "CreateEscrow should succeed")

	esc, err := ec.GetEscrow(ctx, jobID)
	a.NoError(err)
	a.Equal(payment.EscrowStateCreated.String(), esc.State.String(),
		"escrow should be Created before selection")

	providers := [3]common.Address{provider, {}, {}}
	_, err = ec.SelectProviders(ctx, jobID, providers)
	a.NoError(err, "SelectProviders should succeed")

	esc, err = ec.GetEscrow(ctx, jobID)
	a.NoError(err)
	a.Equal(payment.EscrowStateActive.String(), esc.State.String(),
		"escrow should be Active after selection")

	_, err = ec.ReleasePayment(ctx, jobID, durationSecs) // full duration
	a.NoError(err, "ReleasePayment should succeed")

	esc, err = ec.GetEscrow(ctx, jobID)
	a.NoError(err)
	a.Equal(amount.String(), esc.Released.String(),
		"full amount should be released after full duration")
	a.Equal(big.NewInt(0).String(), ec.CalculateRemainingEscrow(esc).String(),
		"no remaining escrow after full release")

	_, err = ec.FinalizeEscrow(ctx, jobID)
	a.NoError(err, "FinalizeEscrow should succeed")

	esc, err = ec.GetEscrow(ctx, jobID)
	a.NoError(err)
	a.Equal(payment.EscrowStateCompleted.String(), esc.State.String(),
		"escrow should be Completed after finalization")
}

// TestGoldenPath_EscrowAnvilMode is the opt-in real-chain counterpart. It SKIPS
// unless MOLTBUNKER_ANVIL_RPC points at a running anvil node. When set, it mints
// a fresh in-memory account (never written to disk), dials the RPC, and verifies
// the daemon can construct + connect a real payment.BaseClient against the local
// chain — proving the on-chain leg is reachable end-to-end.
//
// [REAL: payment.BaseClient against anvil]
//
// NOTE on scope: a full real CreateEscrow->Finalize against anvil requires the
// BunkerEscrow + token contracts to be DEPLOYED first (forge script), which the
// existing tests/integration job already does with the foundry-toolchain. That
// deployment is intentionally out of scope for this golden test; wiring it
// (foundry-toolchain@v1 + forge deploy + contract addresses) is the follow-on
// tracked in daemon-todo.md. This test proves the chain dial + client
// construction leg, and the mock-mode test above proves the settlement state
// machine.
func TestGoldenPath_EscrowAnvilMode(t *testing.T) {
	rpc := os.Getenv(anvilRPCEnv)
	if rpc == "" {
		t.Skipf("%s not set; skipping real on-chain escrow leg (mock-mode test covers settlement)", anvilRPCEnv)
	}
	a := testutil.NewAssertions(t)
	t.Logf("[REAL: anvil] dialing %s and constructing a connected BaseClient", rpc)

	// Fresh ECDSA account — in memory only, never persisted.
	key, err := ethcrypto.GenerateKey()
	a.NoError(err, "ephemeral account generation should succeed")
	addr := ethcrypto.PubkeyToAddress(key.PublicKey)
	a.NotEqual(common.Address{}, addr, "ephemeral account address should be non-zero")

	cfg := &payment.BaseClientConfig{
		RPCURL:  rpc,
		ChainID: 31337, // anvil default
	}
	bc, err := payment.NewBaseClient(cfg, key)
	a.NoError(err, "BaseClient construction should succeed")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	a.NoError(bc.Connect(ctx), "BaseClient should connect to anvil")
	defer bc.Close()
	a.True(bc.IsConnected(), "BaseClient should report connected after Connect")
}
