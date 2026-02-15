// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEscrow.sol";
import "../src/BunkerPricing.sol";

/// @title ForkTest
/// @notice Fork tests that deploy Moltbunker contracts onto a Base mainnet fork
///         to verify they interact correctly with real chain state.
/// @dev Requires BASE_RPC_URL env variable (e.g., https://mainnet.base.org).
///      Run with: forge test --match-contract ForkTest --fork-url $BASE_RPC_URL
///      or rely on vm.createSelectFork which reads from foundry.toml rpc_endpoints.
contract ForkTest is Test {
    BunkerToken token;
    BunkerStaking staking;
    BunkerEscrow escrow;
    BunkerPricing pricing;

    address owner;
    address treasury;
    address operator;
    address provider0;
    address provider1;
    address provider2;
    address requester;

    uint256 baseFork;

    function setUp() public {
        // Create a fork of Base mainnet. Uses BASE_RPC_URL env variable.
        // Falls back to public RPC if env is not set.
        string memory rpcUrl = vm.envOr("BASE_RPC_URL", string("https://mainnet.base.org"));
        baseFork = vm.createSelectFork(rpcUrl);

        // Create deterministic addresses
        owner = makeAddr("forkOwner");
        treasury = makeAddr("forkTreasury");
        operator = makeAddr("forkOperator");
        provider0 = makeAddr("forkProvider0");
        provider1 = makeAddr("forkProvider1");
        provider2 = makeAddr("forkProvider2");
        requester = makeAddr("forkRequester");

        // Fund accounts with ETH for gas
        vm.deal(owner, 10 ether);
        vm.deal(requester, 10 ether);
        vm.deal(provider0, 10 ether);
        vm.deal(provider1, 10 ether);
        vm.deal(provider2, 10 ether);

        // Deploy all contracts on the fork
        vm.startPrank(owner);
        token = new BunkerToken(owner);
        staking = new BunkerStaking(address(token), treasury, owner);
        escrow = new BunkerEscrow(address(token), treasury, owner);
        pricing = new BunkerPricing(owner);

        // Grant roles
        escrow.grantRole(escrow.OPERATOR_ROLE(), operator);
        staking.grantRole(staking.SLASHER_ROLE(), owner);
        staking.setSlashingEnabled(true);

        // Mint tokens to participants
        token.mint(requester, 1_000_000e18);
        token.mint(provider0, 200_000_000e18);
        token.mint(provider1, 200_000_000e18);
        token.mint(provider2, 200_000_000e18);
        vm.stopPrank();

        // Approvals
        vm.prank(requester);
        token.approve(address(escrow), type(uint256).max);

        vm.prank(provider0);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider1);
        token.approve(address(staking), type(uint256).max);
        vm.prank(provider2);
        token.approve(address(staking), type(uint256).max);
    }

    // ──────────────────────────────────────────────
    //  Fork State Tests
    // ──────────────────────────────────────────────

    /// @notice Verify we are on the Base mainnet fork (chainId = 8453).
    function test_fork_chainId() public view {
        assertEq(block.chainid, 8453, "Should be on Base mainnet");
    }

    /// @notice Verify the fork has a non-zero block number (real chain state).
    function test_fork_blockNumber() public view {
        assertGt(block.number, 0, "Block number should be > 0 on fork");
    }

    /// @notice Verify that ETH balances exist for funded addresses.
    function test_fork_ethBalances() public view {
        assertEq(owner.balance, 10 ether, "Owner should have 10 ETH");
        assertEq(requester.balance, 10 ether, "Requester should have 10 ETH");
    }

    /// @notice Verify that known Base contracts have code deployed.
    ///   WETH on Base is at 0x4200000000000000000000000000000000000006.
    function test_fork_baseContractHasCode() public view {
        address weth = 0x4200000000000000000000000000000000000006;
        uint256 codeSize;
        assembly {
            codeSize := extcodesize(weth)
        }
        assertGt(codeSize, 0, "WETH contract should have code on Base");
    }

    // ──────────────────────────────────────────────
    //  Contract Deployment on Fork
    // ──────────────────────────────────────────────

    /// @notice Verify all contracts deployed successfully on the fork.
    function test_fork_contractsDeployed() public view {
        assertTrue(address(token) != address(0), "Token should be deployed");
        assertTrue(address(staking) != address(0), "Staking should be deployed");
        assertTrue(address(escrow) != address(0), "Escrow should be deployed");
        assertTrue(address(pricing) != address(0), "Pricing should be deployed");

        // Verify contracts have code
        address tokenAddr = address(token);
        uint256 codeSize;
        assembly {
            codeSize := extcodesize(tokenAddr)
        }
        assertGt(codeSize, 0, "Token contract should have code");
    }

    /// @notice Verify contract metadata is correct on the fork.
    function test_fork_contractMetadata() public view {
        assertEq(token.name(), "Bunker Token");
        assertEq(token.symbol(), "BUNKER");
        assertEq(token.decimals(), 18);
        assertEq(token.SUPPLY_CAP(), 100_000_000_000e18);
        assertEq(keccak256(bytes(token.VERSION())), keccak256(bytes("1.0.0")));
        assertEq(keccak256(bytes(staking.VERSION())), keccak256(bytes("1.3.0")));
        assertEq(keccak256(bytes(escrow.VERSION())), keccak256(bytes("1.2.0")));
        assertEq(keccak256(bytes(pricing.VERSION())), keccak256(bytes("1.2.0")));
    }

    // ──────────────────────────────────────────────
    //  Full Lifecycle on Fork
    // ──────────────────────────────────────────────

    /// @notice End-to-end lifecycle on fork: stake -> price -> escrow -> pay.
    function test_fork_fullLifecycle() public {
        // 1. Providers stake
        vm.prank(provider0);
        staking.stake(10_000_000e18);
        vm.prank(provider1);
        staking.stake(10_000_000e18);
        vm.prank(provider2);
        staking.stake(10_000_000e18);

        assertTrue(staking.isActiveProvider(provider0));
        assertTrue(staking.isActiveProvider(provider1));
        assertTrue(staking.isActiveProvider(provider2));
        assertEq(staking.totalStaked(), 30_000_000e18);

        // 2. Calculate deployment cost
        BunkerPricing.ResourceRequest memory req = BunkerPricing.ResourceRequest({
            cpuCores: 2000,
            memoryGB: 4000,
            storageGB: 20,
            networkGB: 10,
            durationHours: 24,
            useGPUBasic: false,
            useGPUPremium: false,
            useTor: false,
            usePremiumSLA: false,
            useSpot: false
        });
        uint256 cost = pricing.calculateCost(req);
        assertGt(cost, 0, "Cost should be > 0");

        // 3. Create escrow reservation
        vm.prank(requester);
        uint256 reservationId = escrow.createReservation(cost, 24 hours);
        assertEq(reservationId, 1);
        assertEq(token.balanceOf(address(escrow)), cost);

        // 4. Select providers
        vm.prank(operator);
        escrow.selectProviders(reservationId, [provider0, provider1, provider2]);

        BunkerEscrow.Reservation memory res = escrow.getReservation(reservationId);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Active));

        // 5. Release 50% payment
        vm.prank(operator);
        escrow.releasePayment(reservationId, 12 hours);

        // 6. Finalize
        uint256 totalSupplyBefore = token.totalSupply();
        vm.prank(operator);
        escrow.finalizeReservation(reservationId);

        res = escrow.getReservation(reservationId);
        assertEq(uint8(res.status), uint8(BunkerEscrow.Status.Completed));

        // Escrow should be empty
        assertEq(token.balanceOf(address(escrow)), 0, "Escrow should be empty after finalization");

        // Total supply should have decreased (tokens burned from protocol fee)
        assertLt(token.totalSupply(), totalSupplyBefore, "Total supply should decrease due to burns");
    }

    /// @notice Staking and slashing on fork to verify token burn mechanics.
    function test_fork_stakingAndSlashing() public {
        // Provider stakes Gold tier
        vm.prank(provider0);
        staking.stake(100_000_000e18);
        assertEq(uint8(staking.getTier(provider0)), uint8(BunkerStaking.Tier.Gold));

        uint256 totalSupplyBefore = token.totalSupply();
        uint256 treasuryBefore = token.balanceOf(treasury);

        // Slash 10_000_000 tokens
        uint256 slashAmount = 10_000_000e18;
        vm.prank(owner);
        staking.slash(provider0, slashAmount);

        // 80% burned, 20% to treasury
        uint256 expectedBurn = (slashAmount * 8000) / 10000;
        uint256 expectedTreasury = slashAmount - expectedBurn;

        assertEq(token.totalSupply(), totalSupplyBefore - expectedBurn, "Burns should reduce supply");
        assertEq(
            token.balanceOf(treasury) - treasuryBefore,
            expectedTreasury,
            "Treasury should receive 20%"
        );

        // Provider should still be active (90M remaining, Silver tier)
        assertTrue(staking.isActiveProvider(provider0));
        assertEq(uint8(staking.getTier(provider0)), uint8(BunkerStaking.Tier.Silver));
    }
}
