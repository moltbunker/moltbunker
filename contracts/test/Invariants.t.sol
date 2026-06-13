// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "forge-std/StdInvariant.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEscrow.sol";

/// @title Invariants
/// @notice Stateful, handler-based Foundry invariant suites for the three
///         accounting-critical Moltbunker contracts. These complement the
///         single-call fuzz tests by driving arbitrary multi-call sequences and
///         asserting global accounting invariants after every step.
///
/// @dev Pure in-process Foundry — no external mocks. Each suite uses the
///      StdInvariant `targetContract` pattern so the fuzzer only calls the
///      bounded handler functions (never the raw contract), keeping every
///      sequence reachable from a realistic external actor.

// ──────────────────────────────────────────────────────────────────────────
//  (1) BunkerToken: totalSupply <= SUPPLY_CAP at all times
// ──────────────────────────────────────────────────────────────────────────

/// @dev Drives mint/burn against BunkerToken from the owner. Mint is bounded to
///      the remaining mintable supply so legitimate calls never revert; the
///      invariant proves no reachable sequence can exceed the cap.
contract TokenHandler is Test {
    BunkerToken public token;
    address public owner;

    constructor(BunkerToken _token, address _owner) {
        token = _token;
        owner = _owner;
    }

    /// @dev Mint a bounded amount to this handler.
    function handler_mint(uint256 amount) external {
        uint256 mintable = token.mintableSupply();
        if (mintable == 0) return; // cap reached — nothing to do
        amount = bound(amount, 1, mintable);
        vm.prank(owner);
        token.mint(address(this), amount);
    }

    /// @dev Burn a bounded amount the handler currently holds.
    function handler_burn(uint256 amount) external {
        uint256 bal = token.balanceOf(address(this));
        if (bal == 0) return;
        amount = bound(amount, 0, bal);
        if (amount == 0) return;
        token.burn(amount);
    }
}

contract BunkerTokenInvariantTest is StdInvariant, Test {
    BunkerToken public token;
    TokenHandler public handler;

    address public owner = makeAddr("tokenOwner");

    function setUp() public {
        vm.prank(owner);
        token = new BunkerToken(owner);

        handler = new TokenHandler(token, owner);

        // Only fuzz the handler; never the token directly.
        targetContract(address(handler));
        excludeContract(address(token));
    }

    /// @notice totalSupply must never exceed the hard supply cap.
    function invariant_totalSupplyNeverExceedsCap() public view {
        assertLe(token.totalSupply(), token.SUPPLY_CAP());
    }
}

// ──────────────────────────────────────────────────────────────────────────
//  (2) BunkerStaking: stake accounting conservation + no counter underflow
// ──────────────────────────────────────────────────────────────────────────

/// @dev Drives stake / requestUnstake / completeUnstake / slashImmediate against
///      a single staker (this handler). All amounts are bounded to currently
///      legal ranges so calls do not revert spuriously; the invariants then prove
///      the global counters stay consistent with the contract token balance.
contract StakingHandler is Test {
    BunkerToken public token;
    BunkerStaking public staking;
    address public slasher;

    uint256 public constant STARTER_MIN = 1_000_000e18;

    constructor(BunkerToken _token, BunkerStaking _staking, address _slasher) {
        token = _token;
        staking = _staking;
        slasher = _slasher;
    }

    /// @dev Stake a bounded amount. First stake must clear the Starter minimum.
    function handler_stake(uint256 amount) external {
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(address(this));
        if (info.frozen) return;
        uint256 bal = token.balanceOf(address(this));
        if (bal < STARTER_MIN) return;

        uint256 lower = info.active ? 1 : STARTER_MIN;
        if (bal < lower) return;
        amount = bound(amount, lower, bal);

        token.approve(address(staking), amount);
        staking.stake(amount);
    }

    /// @dev Request unstake of a bounded portion of the active stake.
    function handler_requestUnstake(uint256 amount) external {
        BunkerStaking.ProviderInfo memory info = staking.getProviderInfo(address(this));
        if (info.frozen || !info.active) return;
        if (info.stakedAmount == 0) return;
        // Stay under the 50-entry queue bound.
        if (staking.getUnstakeQueueLength(address(this)) >= 50) return;
        amount = bound(amount, 1, info.stakedAmount);
        staking.requestUnstake(amount);
    }

    /// @dev Complete a matured unstake request, warping past the unbonding period.
    function handler_completeUnstake(uint256 idx) external {
        if (staking.getProviderInfo(address(this)).frozen) return;
        uint256 len = staking.getUnstakeQueueLength(address(this));
        if (len == 0) return;
        idx = bound(idx, 0, len - 1);
        BunkerStaking.UnstakeRequest memory req = staking.getUnstakeRequest(address(this), idx);
        if (req.completed) return;
        if (block.timestamp < req.unlockTime) {
            vm.warp(uint256(req.unlockTime) + 1);
        }
        staking.completeUnstake(idx);
    }

    /// @dev Slash a bounded portion of the handler's slashable balance.
    function handler_slashImmediate(uint256 amount) external {
        uint256 slashable = staking.getSlashableBalance(address(this));
        if (slashable == 0) return;
        amount = bound(amount, 1, slashable);
        vm.prank(slasher);
        staking.slashImmediate(address(this), amount);
    }
}

contract BunkerStakingInvariantTest is StdInvariant, Test {
    BunkerToken public token;
    BunkerStaking public staking;
    StakingHandler public handler;

    address public owner = makeAddr("stakingOwner");
    address public treasury = makeAddr("stakingTreasury");
    address public slasher = makeAddr("stakingSlasher");

    uint256 public constant SEED = 50_000_000_000e18; // 50B to the handler

    function setUp() public {
        vm.startPrank(owner);
        token = new BunkerToken(owner);
        staking = new BunkerStaking(address(token), treasury, owner);
        staking.grantRole(staking.SLASHER_ROLE(), slasher);
        staking.setSlashingEnabled(true);
        vm.stopPrank();

        handler = new StakingHandler(token, staking, slasher);

        // Seed the handler with tokens to stake.
        vm.prank(owner);
        token.mint(address(handler), SEED);

        targetContract(address(handler));
        excludeContract(address(staking));
        excludeContract(address(token));
    }

    /// @notice The contract's token balance always backs the tracked stake + unbonding.
    /// @dev Slashing burns/transfers tokens out and decrements the counters in lockstep,
    ///      so the balance is always >= the sum (extra balance can only appear via
    ///      reward injection, which this suite does not perform).
    function invariant_stakingTokenConservation() public view {
        uint256 tracked = staking.totalStaked() + staking.totalUnbonding();
        assertLe(tracked, token.balanceOf(address(staking)));
    }

    /// @notice The global totalStaked counter must never wrap past uint128 range
    ///         (the per-provider fields are uint128; the global is uint256 but must
    ///         stay within the same bound given all stake originates from uint128 fields).
    function invariant_totalStakedNoUnderflow() public view {
        assertLe(staking.totalStaked(), type(uint128).max);
    }

    /// @notice The global totalUnbonding counter must never wrap past uint128 range.
    function invariant_totalUnbondingNoUnderflow() public view {
        assertLe(staking.totalUnbonding(), type(uint128).max);
    }
}

// ──────────────────────────────────────────────────────────────────────────
//  (3) BunkerEscrow: sum of unreleased reservation balances <= contract balance
// ──────────────────────────────────────────────────────────────────────────

/// @dev Drives the escrow reservation lifecycle. The requester (this handler)
///      creates reservations; the operator (this handler too, via OPERATOR_ROLE)
///      selects providers, releases progressive payments, finalizes, and refunds.
///      A ghost set of reservation IDs lets the invariant sum unreleased balances.
contract EscrowHandler is Test {
    BunkerToken public token;
    BunkerEscrow public escrow;

    address public p0;
    address public p1;
    address public p2;

    uint256[] public reservationIds;

    uint256 public constant MAX_AMOUNT = 1_000_000e18;
    uint256 public constant MAX_DURATION = 30 days;

    constructor(
        BunkerToken _token,
        BunkerEscrow _escrow,
        address _p0,
        address _p1,
        address _p2
    ) {
        token = _token;
        escrow = _escrow;
        p0 = _p0;
        p1 = _p1;
        p2 = _p2;
    }

    function reservationCount() external view returns (uint256) {
        return reservationIds.length;
    }

    /// @dev Create a reservation funded from the handler's balance.
    function handler_createReservation(uint256 amount, uint256 duration) external {
        uint256 bal = token.balanceOf(address(this));
        if (bal == 0) return;
        amount = bound(amount, 1, bal > MAX_AMOUNT ? MAX_AMOUNT : bal);
        duration = bound(duration, 1, MAX_DURATION);

        token.approve(address(escrow), amount);
        uint256 id = escrow.createReservation(amount, duration);
        reservationIds.push(id);
    }

    /// @dev Move a Created reservation to Active with the three fixed providers.
    function handler_selectProviders(uint256 reservId) external {
        if (reservationIds.length == 0) return;
        reservId = reservationIds[bound(reservId, 0, reservationIds.length - 1)];
        BunkerEscrow.Reservation memory r = escrow.getReservation(reservId);
        if (r.status != BunkerEscrow.Status.Created) return;
        escrow.selectProviders(reservId, [p0, p1, p2]);
    }

    /// @dev Release payment for a settled portion of an Active reservation.
    function handler_releasePayment(uint256 reservId, uint256 settledDuration) external {
        if (reservationIds.length == 0) return;
        reservId = reservationIds[bound(reservId, 0, reservationIds.length - 1)];
        BunkerEscrow.Reservation memory r = escrow.getReservation(reservId);
        if (r.status != BunkerEscrow.Status.Active) return;
        settledDuration = bound(settledDuration, 0, r.duration);
        // releasePayment reverts when there is nothing new to release; guard it.
        uint256 proportional = (uint256(r.totalAmount) * settledDuration) / uint256(r.duration);
        if (proportional <= r.releasedAmount) return;
        escrow.releasePayment(reservId, settledDuration);
    }

    /// @dev Refund an Active or Created reservation back to the requester.
    function handler_refund(uint256 reservId) external {
        if (reservationIds.length == 0) return;
        reservId = reservationIds[bound(reservId, 0, reservationIds.length - 1)];
        BunkerEscrow.Reservation memory r = escrow.getReservation(reservId);
        if (r.status != BunkerEscrow.Status.Active && r.status != BunkerEscrow.Status.Created) {
            return;
        }
        escrow.refund(reservId);
    }

    /// @dev Finalize an Active reservation (releases the remainder).
    function handler_finalizeReservation(uint256 reservId) external {
        if (reservationIds.length == 0) return;
        reservId = reservationIds[bound(reservId, 0, reservationIds.length - 1)];
        BunkerEscrow.Reservation memory r = escrow.getReservation(reservId);
        if (r.status != BunkerEscrow.Status.Active) return;
        escrow.finalizeReservation(reservId);
    }

    /// @dev Advance time so progressive settlement / refund proportions vary.
    function handler_warp(uint256 secs) external {
        secs = bound(secs, 1, 2 days);
        vm.warp(block.timestamp + secs);
    }
}

contract BunkerEscrowInvariantTest is StdInvariant, Test {
    BunkerToken public token;
    BunkerStaking public staking;
    BunkerEscrow public escrow;
    EscrowHandler public handler;

    address public owner = makeAddr("escrowOwner");
    address public treasury = makeAddr("escrowTreasury");

    // Provider addresses distinct from the handler to avoid re-entrancy/ghost mismatch.
    address public p0 = makeAddr("escrowProvider0");
    address public p1 = makeAddr("escrowProvider1");
    address public p2 = makeAddr("escrowProvider2");

    uint256 public constant SEED = 10_000_000_000e18; // 10B to the handler

    function setUp() public {
        vm.startPrank(owner);
        token = new BunkerToken(owner);
        staking = new BunkerStaking(address(token), treasury, owner);
        escrow = new BunkerEscrow(address(token), treasury, owner);
        escrow.setStakingContract(address(staking));
        vm.stopPrank();

        // Stake the three providers so selectProviders' active-provider check passes.
        _stakeProviders();

        handler = new EscrowHandler(token, escrow, p0, p1, p2);

        vm.startPrank(owner);
        escrow.grantRole(escrow.OPERATOR_ROLE(), address(handler));
        token.mint(address(handler), SEED);
        vm.stopPrank();

        targetContract(address(handler));
        excludeContract(address(staking));
        excludeContract(address(escrow));
        excludeContract(address(token));
    }

    /// @dev Stake the three providers so selectProviders' isActiveProvider check passes.
    function _stakeProviders() internal {
        uint256 min = 1_000_000e18; // Starter minimum
        address[3] memory ps = [p0, p1, p2];
        for (uint256 i = 0; i < 3; i++) {
            vm.prank(owner);
            token.mint(ps[i], min);
            vm.startPrank(ps[i]);
            token.approve(address(staking), min);
            staking.stake(min);
            vm.stopPrank();
        }
    }

    /// @notice The escrow contract balance always covers every unreleased reservation
    ///         balance still owed (Created or Active). Released funds have already left
    ///         the contract (to providers/treasury/burn), so they are excluded.
    function invariant_escrowConservation() public view {
        uint256 owed;
        uint256 count = handler.reservationCount();
        for (uint256 i = 0; i < count; i++) {
            uint256 id = handler.reservationIds(i);
            BunkerEscrow.Reservation memory r = escrow.getReservation(id);
            if (r.status == BunkerEscrow.Status.Created || r.status == BunkerEscrow.Status.Active) {
                owed += uint256(r.totalAmount) - uint256(r.releasedAmount);
            }
        }
        assertLe(owed, token.balanceOf(address(escrow)));
    }
}
