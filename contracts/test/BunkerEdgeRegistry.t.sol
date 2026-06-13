// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerToken.sol";
import "../src/BunkerStaking.sol";
import "../src/BunkerEdgeRegistry.sol";

contract BunkerEdgeRegistryTest is Test {
    BunkerToken public token;
    BunkerStaking public staking;
    BunkerEdgeRegistry public registry;

    address public owner = makeAddr("owner");
    address public treasury = makeAddr("treasury");
    address public slasher = makeAddr("slasher");
    address public edge1 = makeAddr("edge1");
    address public edge2 = makeAddr("edge2");
    address public lowStaker = makeAddr("lowStaker");
    address public stranger = makeAddr("stranger");

    // BunkerStaking tier minimums (mirrored from BunkerStaking constructor).
    uint256 constant STARTER_MIN = 1_000_000e18;
    uint256 constant SILVER_MIN = 10_000_000e18;
    uint256 constant GOLD_MIN = 100_000_000e18;
    uint256 constant PLATINUM_MIN = 1_000_000_000e18;

    // BunkerEdgeRegistry defaults.
    uint256 constant DEFAULT_MIN_EDGE_STAKE = 50_000_000e18;
    uint256 constant DEFAULT_APPEAL_WINDOW = 48 hours;

    bytes32 constant NODE_ID_1 = keccak256("edge1-node");
    bytes32 constant NODE_ID_2 = keccak256("edge2-node");
    bytes32 constant REGION_US = bytes32("us-east");
    bytes32 constant REGION_EU = bytes32("eu-west");

    string constant ENDPOINT_1 = "https://edge1.moltbunker.dev";
    bytes constant TLS_HASH_1 = hex"0011223344556677";

    function setUp() public {
        vm.startPrank(owner);
        token = new BunkerToken(owner);
        staking = new BunkerStaking(address(token), treasury, owner);

        // Grant slasher role on staking and enable slashing so routed slashes work.
        staking.grantRole(staking.SLASHER_ROLE(), slasher);
        staking.setSlashingEnabled(true);

        // Mint enough for the edge providers to clear the 50 M threshold.
        token.mint(edge1, 200_000_000e18);
        token.mint(edge2, 200_000_000e18);
        token.mint(lowStaker, 200_000_000e18);
        vm.stopPrank();

        // edge1 stakes 50 M (exactly at the edge threshold).
        vm.startPrank(edge1);
        token.approve(address(staking), type(uint256).max);
        staking.stake(DEFAULT_MIN_EDGE_STAKE);
        vm.stopPrank();

        // edge2 stakes 100 M (comfortably above the threshold).
        vm.startPrank(edge2);
        token.approve(address(staking), type(uint256).max);
        staking.stake(100_000_000e18);
        vm.stopPrank();

        // lowStaker stakes only 10 M (below the 50 M edge threshold).
        vm.startPrank(lowStaker);
        token.approve(address(staking), type(uint256).max);
        staking.stake(SILVER_MIN);
        vm.stopPrank();

        // Deploy the edge registry and wire up the slasher role.
        vm.startPrank(owner);
        registry = new BunkerEdgeRegistry(address(staking), owner);
        registry.grantRole(registry.SLASHER_ROLE(), slasher);

        // Grant the registry SLASHER_ROLE on BunkerStaking so routed slashes work.
        staking.grantRole(staking.SLASHER_ROLE(), address(registry));
        vm.stopPrank();
    }

    // ================================================================
    //  1. CONSTRUCTOR
    // ================================================================

    function test_constructor_setsStakingContract() public view {
        assertEq(address(registry.stakingContract()), address(staking));
    }

    function test_constructor_setsOwner() public view {
        assertEq(registry.owner(), owner);
    }

    function test_constructor_grantsDefaultAdminRole() public view {
        assertTrue(registry.hasRole(registry.DEFAULT_ADMIN_ROLE(), owner));
    }

    function test_constructor_defaultMinStake() public view {
        assertEq(registry.minStakeForEdge(), DEFAULT_MIN_EDGE_STAKE);
    }

    function test_constructor_defaultAppealWindow() public view {
        assertEq(registry.appealWindow(), DEFAULT_APPEAL_WINDOW);
    }

    function test_constructor_slashingDisabledByDefault() public view {
        assertFalse(registry.slashingEnabled());
    }

    function test_constructor_version() public view {
        assertEq(registry.VERSION(), "1.0.0");
    }

    function test_constructor_revertsZeroStaking() public {
        vm.prank(owner);
        vm.expectRevert(BunkerEdgeRegistry.ZeroAddress.selector);
        new BunkerEdgeRegistry(address(0), owner);
    }

    function test_constructor_revertsZeroOwner() public {
        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSelector(bytes4(keccak256("OwnableInvalidOwner(address)")), address(0))
        );
        new BunkerEdgeRegistry(address(staking), address(0));
    }

    function test_constants() public view {
        assertEq(registry.SLASHER_ROLE(), keccak256("SLASHER_ROLE"));
        assertEq(registry.MIN_EDGE_STAKE_FLOOR(), STARTER_MIN);
        assertEq(registry.MIN_EDGE_STAKE_CEILING(), PLATINUM_MIN);
        assertEq(registry.APPEAL_WINDOW_FLOOR(), 12 hours);
        assertEq(registry.APPEAL_WINDOW_CEILING(), 14 days);
    }

    // ================================================================
    //  2. registerEdgeProvider
    // ================================================================

    function test_register_happyPath() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);

        BunkerEdgeRegistry.EdgeProviderInfo memory info = registry.getEdgeProviderInfo(edge1);
        assertEq(info.nodeId, NODE_ID_1);
        assertEq(info.region, REGION_US);
        assertEq(info.endpointURL, ENDPOINT_1);
        assertEq(info.tlsPubkeyHash, TLS_HASH_1);
        assertTrue(info.active);
        assertFalse(info.frozen);
        assertGt(info.registeredAt, 0);
    }

    function test_register_emitsEvent() public {
        vm.prank(edge1);
        vm.expectEmit(true, false, false, true, address(registry));
        emit BunkerEdgeRegistry.EdgeProviderRegistered(edge1, NODE_ID_1, REGION_US);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
    }

    function test_register_setsNodeIdMapping() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        assertEq(registry.nodeIdToEdgeProvider(NODE_ID_1), edge1);
    }

    function test_register_secondTimeReverts() public {
        vm.startPrank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEdgeRegistry.AlreadyRegistered.selector, edge1)
        );
        registry.registerEdgeProvider(NODE_ID_2, REGION_EU, ENDPOINT_1, TLS_HASH_1);
        vm.stopPrank();
    }

    function test_register_duplicateNodeIdReverts() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);

        vm.prank(edge2);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEdgeRegistry.NodeIdAlreadyClaimed.selector, NODE_ID_1)
        );
        registry.registerEdgeProvider(NODE_ID_1, REGION_EU, ENDPOINT_1, TLS_HASH_1);
    }

    function test_register_insufficientStakeReverts() public {
        vm.prank(lowStaker);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEdgeRegistry.InsufficientStake.selector, SILVER_MIN, DEFAULT_MIN_EDGE_STAKE
            )
        );
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
    }

    function test_register_noStakeReverts() public {
        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEdgeRegistry.InsufficientStake.selector, 0, DEFAULT_MIN_EDGE_STAKE
            )
        );
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
    }

    function test_register_exactThresholdSucceeds() public {
        // edge1 stakes exactly 50 M — must be allowed.
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        assertTrue(registry.isActiveEdgeProvider(edge1));
    }

    function test_register_worksWhileStakingPaused() public {
        // getProviderInfo reads storage and is not pause-guarded, so the stake
        // check must still pass while BunkerStaking is paused (risk #1).
        vm.prank(owner);
        staking.pause();

        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        assertTrue(registry.isActiveEdgeProvider(edge1));
    }

    function test_register_zeroNodeIdSkipsMapping() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(bytes32(0), REGION_US, ENDPOINT_1, TLS_HASH_1);
        // A zero node ID does not claim the mapping (so it stays re-usable).
        assertEq(registry.nodeIdToEdgeProvider(bytes32(0)), address(0));
        assertTrue(registry.isActiveEdgeProvider(edge1));
    }

    function test_register_revertsWhenPaused() public {
        vm.prank(owner);
        registry.pause();

        vm.prank(edge1);
        vm.expectRevert(abi.encodeWithSelector(bytes4(keccak256("EnforcedPause()"))));
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
    }

    // ================================================================
    //  3. deregisterEdgeProvider
    // ================================================================

    function test_deregister_happyPath() public {
        vm.startPrank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        registry.deregisterEdgeProvider();
        vm.stopPrank();

        BunkerEdgeRegistry.EdgeProviderInfo memory info = registry.getEdgeProviderInfo(edge1);
        assertFalse(info.active);
        assertEq(info.nodeId, bytes32(0));
        assertEq(registry.nodeIdToEdgeProvider(NODE_ID_1), address(0));
    }

    function test_deregister_clearsMetadata() public {
        vm.startPrank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        registry.deregisterEdgeProvider();
        vm.stopPrank();

        // No stale routing metadata must remain after deregistration.
        BunkerEdgeRegistry.EdgeProviderInfo memory info = registry.getEdgeProviderInfo(edge1);
        assertFalse(info.active);
        assertEq(info.nodeId, bytes32(0));
        assertEq(info.region, bytes32(0));
        assertEq(info.endpointURL, "");
        assertEq(info.tlsPubkeyHash, hex"");
    }

    function test_deregister_emitsEvent() public {
        vm.startPrank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        vm.expectEmit(true, false, false, false, address(registry));
        emit BunkerEdgeRegistry.EdgeProviderDeregistered(edge1);
        registry.deregisterEdgeProvider();
        vm.stopPrank();
    }

    function test_deregister_notActiveReverts() public {
        vm.prank(edge1);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.NotActive.selector, edge1));
        registry.deregisterEdgeProvider();
    }

    function test_deregister_twiceReverts() public {
        vm.startPrank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        registry.deregisterEdgeProvider();
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.NotActive.selector, edge1));
        registry.deregisterEdgeProvider();
        vm.stopPrank();
    }

    function test_deregister_allowsNodeIdReclaim() public {
        vm.startPrank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        registry.deregisterEdgeProvider();
        vm.stopPrank();

        // edge2 can now claim the released node ID.
        vm.prank(edge2);
        registry.registerEdgeProvider(NODE_ID_1, REGION_EU, ENDPOINT_1, TLS_HASH_1);
        assertEq(registry.nodeIdToEdgeProvider(NODE_ID_1), edge2);
    }

    // ================================================================
    //  4. updateEndpoint
    // ================================================================

    function test_updateEndpoint_happyPath() public {
        vm.startPrank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        registry.updateEndpoint("https://new.edge.dev", hex"aabbcc");
        vm.stopPrank();

        BunkerEdgeRegistry.EdgeProviderInfo memory info = registry.getEdgeProviderInfo(edge1);
        assertEq(info.endpointURL, "https://new.edge.dev");
        assertEq(info.tlsPubkeyHash, hex"aabbcc");
    }

    function test_updateEndpoint_emitsEvent() public {
        vm.startPrank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        vm.expectEmit(true, false, false, true, address(registry));
        emit BunkerEdgeRegistry.EdgeEndpointUpdated(edge1, "https://new.edge.dev");
        registry.updateEndpoint("https://new.edge.dev", hex"aabbcc");
        vm.stopPrank();
    }

    function test_updateEndpoint_notActiveReverts() public {
        vm.prank(edge1);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.NotActive.selector, edge1));
        registry.updateEndpoint("https://x.dev", hex"00");
    }

    function test_updateEndpoint_frozenReverts() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);

        vm.prank(slasher);
        registry.freezeEdgeProvider(edge1);

        vm.prank(edge1);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.ProviderIsFrozen.selector, edge1));
        registry.updateEndpoint("https://x.dev", hex"00");
    }

    // ================================================================
    //  5. freeze / unfreeze
    // ================================================================

    function test_freeze_bySlasher() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);

        vm.prank(slasher);
        registry.freezeEdgeProvider(edge1);

        assertTrue(registry.getEdgeProviderInfo(edge1).frozen);
        assertFalse(registry.isActiveEdgeProvider(edge1)); // frozen => not active to consumers
    }

    function test_freeze_emitsEvent() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);

        vm.prank(slasher);
        vm.expectEmit(true, true, false, false, address(registry));
        emit BunkerEdgeRegistry.EdgeProviderFrozen(edge1, slasher);
        registry.freezeEdgeProvider(edge1);
    }

    function test_freeze_nonSlasherReverts() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);

        bytes32 slasherRole = registry.SLASHER_ROLE();
        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(
                bytes4(keccak256("AccessControlUnauthorizedAccount(address,bytes32)")),
                stranger,
                slasherRole
            )
        );
        registry.freezeEdgeProvider(edge1);
    }

    function test_freeze_notActiveReverts() public {
        vm.prank(slasher);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.NotActive.selector, edge1));
        registry.freezeEdgeProvider(edge1);
    }

    function test_unfreeze_byOwner() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);

        vm.prank(slasher);
        registry.freezeEdgeProvider(edge1);

        vm.prank(owner);
        registry.unfreezeEdgeProvider(edge1);

        assertFalse(registry.getEdgeProviderInfo(edge1).frozen);
        assertTrue(registry.isActiveEdgeProvider(edge1));
    }

    function test_unfreeze_emitsEvent() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        vm.prank(slasher);
        registry.freezeEdgeProvider(edge1);

        vm.prank(owner);
        vm.expectEmit(true, false, false, false, address(registry));
        emit BunkerEdgeRegistry.EdgeProviderUnfrozen(edge1);
        registry.unfreezeEdgeProvider(edge1);
    }

    function test_unfreeze_nonOwnerReverts() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        vm.prank(slasher);
        registry.freezeEdgeProvider(edge1);

        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(
                bytes4(keccak256("OwnableUnauthorizedAccount(address)")), stranger
            )
        );
        registry.unfreezeEdgeProvider(edge1);
    }

    function test_frozenProviderCannotReregister() public {
        vm.prank(edge1);
        registry.registerEdgeProvider(NODE_ID_1, REGION_US, ENDPOINT_1, TLS_HASH_1);
        vm.prank(slasher);
        registry.freezeEdgeProvider(edge1);

        // deregister first so 'active' is false, then re-register is blocked by frozen.
        vm.startPrank(edge1);
        registry.deregisterEdgeProvider();
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.ProviderIsFrozen.selector, edge1));
        registry.registerEdgeProvider(NODE_ID_2, REGION_EU, ENDPOINT_1, TLS_HASH_1);
        vm.stopPrank();
    }

    // ================================================================
    //  6. proposeEdgeSlash
    // ================================================================

    function _register(address who, bytes32 nodeId, bytes32 region) internal {
        vm.prank(who);
        registry.registerEdgeProvider(nodeId, region, ENDPOINT_1, TLS_HASH_1);
    }

    function test_propose_happyPath() public {
        _register(edge1, NODE_ID_1, REGION_US);

        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");

        assertEq(id, 0);
        assertEq(registry.slashProposalCount(), 1);

        BunkerEdgeRegistry.EdgeSlashProposal memory p = registry.getEdgeSlashProposal(id);
        assertEq(p.provider, edge1);
        assertEq(p.amount, 1_000_000e18);
        assertEq(p.reason, "downtime");
        assertEq(p.appealWindowSnapshot, DEFAULT_APPEAL_WINDOW);
        assertFalse(p.executed);
        assertFalse(p.appealed);
    }

    function test_propose_emitsEvent() public {
        _register(edge1, NODE_ID_1, REGION_US);

        vm.prank(slasher);
        vm.expectEmit(true, true, false, true, address(registry));
        emit BunkerEdgeRegistry.EdgeSlashProposed(0, edge1, 1_000_000e18, "downtime");
        registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
    }

    function test_propose_snapshotsAppealWindow() public {
        _register(edge1, NODE_ID_1, REGION_US);

        // Change the appeal window AFTER the proposal — snapshot must hold the old value.
        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");

        vm.prank(owner);
        registry.setAppealWindow(14 days);

        assertEq(registry.getEdgeSlashProposal(id).appealWindowSnapshot, DEFAULT_APPEAL_WINDOW);
    }

    function test_propose_nonSlasherReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);

        bytes32 slasherRole = registry.SLASHER_ROLE();
        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(
                bytes4(keccak256("AccessControlUnauthorizedAccount(address,bytes32)")),
                stranger,
                slasherRole
            )
        );
        registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
    }

    function test_propose_zeroAmountReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);

        vm.prank(slasher);
        vm.expectRevert(BunkerEdgeRegistry.ZeroAmount.selector);
        registry.proposeEdgeSlash(edge1, 0, "downtime");
    }

    function test_propose_zeroProviderReverts() public {
        vm.prank(slasher);
        vm.expectRevert(BunkerEdgeRegistry.ZeroAddress.selector);
        registry.proposeEdgeSlash(address(0), 1_000_000e18, "downtime");
    }

    function test_propose_nonRegisteredProviderReverts() public {
        // edge2 has plenty of BunkerStaking stake but never registered as an edge
        // provider. The registry holds SLASHER_ROLE on BunkerStaking, so it must
        // not be usable to propose slashes against stakers it does not govern.
        vm.prank(slasher);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.NotActive.selector, edge2));
        registry.proposeEdgeSlash(edge2, 1_000_000e18, "downtime");
    }

    function test_propose_deregisteredProviderReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(edge1);
        registry.deregisterEdgeProvider();

        // Once deregistered the provider is no longer active, so proposing a slash
        // against it must revert even though its BunkerStaking stake persists.
        vm.prank(slasher);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.NotActive.selector, edge1));
        registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
    }

    function test_propose_exceedingSlashableBalanceReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);

        // edge1 has exactly DEFAULT_MIN_EDGE_STAKE (50 M) staked and nothing
        // unbonding, so the slashable balance is 50 M. A proposal for more must
        // revert at propose time (defensive pre-check mirroring BunkerStaking H-05).
        uint256 slashable = staking.getSlashableBalance(edge1);
        assertEq(slashable, DEFAULT_MIN_EDGE_STAKE);

        vm.prank(slasher);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEdgeRegistry.InsufficientSlashableBalance.selector,
                slashable + 1,
                slashable
            )
        );
        registry.proposeEdgeSlash(edge1, uint128(slashable + 1), "downtime");
    }

    function test_propose_atSlashableBalanceSucceeds() public {
        _register(edge1, NODE_ID_1, REGION_US);

        uint256 slashable = staking.getSlashableBalance(edge1);
        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, uint128(slashable), "full slash");
        assertEq(registry.getEdgeSlashProposal(id).amount, slashable);
    }

    // ================================================================
    //  7. executeEdgeSlash
    // ================================================================

    function test_execute_slashingDisabledReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);

        vm.startPrank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
        vm.warp(block.timestamp + DEFAULT_APPEAL_WINDOW + 1);

        // registry.slashingEnabled defaults to false.
        vm.expectRevert(BunkerEdgeRegistry.SlashingNotEnabled.selector);
        registry.executeEdgeSlash(id);
        vm.stopPrank();
    }

    function test_execute_windowNotElapsedReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);

        vm.prank(owner);
        registry.setSlashingEnabled(true);

        vm.startPrank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEdgeRegistry.AppealWindowNotElapsed.selector, id)
        );
        registry.executeEdgeSlash(id);
        vm.stopPrank();
    }

    function test_execute_happyPath_callsStakingSlash() public {
        _register(edge1, NODE_ID_1, REGION_US);

        vm.prank(owner);
        registry.setSlashingEnabled(true);

        uint256 slashAmount = 1_000_000e18;
        uint256 stakedBefore = staking.getProviderInfo(edge1).stakedAmount;

        vm.startPrank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, uint128(slashAmount), "downtime");
        vm.warp(block.timestamp + DEFAULT_APPEAL_WINDOW + 1);

        vm.expectEmit(true, true, false, true, address(registry));
        emit BunkerEdgeRegistry.EdgeSlashExecuted(id, edge1, slashAmount);
        registry.executeEdgeSlash(id);
        vm.stopPrank();

        // The slash actually hit BunkerStaking stake.
        uint256 stakedAfter = staking.getProviderInfo(edge1).stakedAmount;
        assertEq(stakedBefore - stakedAfter, slashAmount);
        assertTrue(registry.getEdgeSlashProposal(id).executed);
    }

    function test_execute_twiceReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(owner);
        registry.setSlashingEnabled(true);

        vm.startPrank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
        vm.warp(block.timestamp + DEFAULT_APPEAL_WINDOW + 1);
        registry.executeEdgeSlash(id);

        vm.expectRevert(
            abi.encodeWithSelector(BunkerEdgeRegistry.ProposalAlreadyExecuted.selector, id)
        );
        registry.executeEdgeSlash(id);
        vm.stopPrank();
    }

    function test_execute_appealedReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(owner);
        registry.setSlashingEnabled(true);

        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");

        vm.prank(edge1);
        registry.appealEdgeSlash(id);

        vm.warp(block.timestamp + DEFAULT_APPEAL_WINDOW + 1);
        vm.prank(slasher);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.ProposalAppealed.selector, id));
        registry.executeEdgeSlash(id);
    }

    function test_execute_invalidProposalIdReverts() public {
        vm.prank(owner);
        registry.setSlashingEnabled(true);
        vm.prank(slasher);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.InvalidProposalId.selector, 99));
        registry.executeEdgeSlash(99);
    }

    // ================================================================
    //  8. appealEdgeSlash
    // ================================================================

    function test_appeal_byTargetedProvider() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");

        vm.prank(edge1);
        vm.expectEmit(true, true, false, false, address(registry));
        emit BunkerEdgeRegistry.EdgeSlashAppealed(id, edge1);
        registry.appealEdgeSlash(id);

        assertTrue(registry.getEdgeSlashProposal(id).appealed);
    }

    function test_appeal_nonProviderReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");

        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(
                BunkerEdgeRegistry.NotProposalProvider.selector, id, stranger
            )
        );
        registry.appealEdgeSlash(id);
    }

    function test_appeal_afterWindowReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");

        vm.warp(block.timestamp + DEFAULT_APPEAL_WINDOW + 1);
        vm.prank(edge1);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.AppealWindowElapsed.selector, id));
        registry.appealEdgeSlash(id);
    }

    function test_appeal_twiceReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");

        vm.startPrank(edge1);
        registry.appealEdgeSlash(id);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.ProposalAppealed.selector, id));
        registry.appealEdgeSlash(id);
        vm.stopPrank();
    }

    function test_appeal_executedReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(owner);
        registry.setSlashingEnabled(true);

        vm.startPrank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
        vm.warp(block.timestamp + DEFAULT_APPEAL_WINDOW + 1);
        registry.executeEdgeSlash(id);
        vm.stopPrank();

        vm.prank(edge1);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEdgeRegistry.ProposalAlreadyExecuted.selector, id)
        );
        registry.appealEdgeSlash(id);
    }

    // ================================================================
    //  9. resolveEdgeSlashAppeal
    // ================================================================

    function test_resolve_upheldExecutesSlash() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(owner);
        registry.setSlashingEnabled(true);

        uint256 slashAmount = 1_000_000e18;
        uint256 stakedBefore = staking.getProviderInfo(edge1).stakedAmount;

        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, uint128(slashAmount), "downtime");
        vm.prank(edge1);
        registry.appealEdgeSlash(id);

        vm.prank(owner);
        vm.expectEmit(true, false, false, true, address(registry));
        emit BunkerEdgeRegistry.EdgeAppealResolved(id, true);
        registry.resolveEdgeSlashAppeal(id, true);

        uint256 stakedAfter = staking.getProviderInfo(edge1).stakedAmount;
        assertEq(stakedBefore - stakedAfter, slashAmount);
        assertTrue(registry.getEdgeSlashProposal(id).executed);
        assertTrue(registry.getEdgeSlashProposal(id).resolved);
    }

    function test_resolve_dismissedDoesNotSlash() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(owner);
        registry.setSlashingEnabled(true);

        uint256 stakedBefore = staking.getProviderInfo(edge1).stakedAmount;

        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
        vm.prank(edge1);
        registry.appealEdgeSlash(id);

        vm.prank(owner);
        registry.resolveEdgeSlashAppeal(id, false);

        uint256 stakedAfter = staking.getProviderInfo(edge1).stakedAmount;
        assertEq(stakedBefore, stakedAfter); // no slash applied
        assertFalse(registry.getEdgeSlashProposal(id).executed);
        assertTrue(registry.getEdgeSlashProposal(id).resolved);
    }

    function test_resolve_nonOwnerReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
        vm.prank(edge1);
        registry.appealEdgeSlash(id);

        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(
                bytes4(keccak256("OwnableUnauthorizedAccount(address)")), stranger
            )
        );
        registry.resolveEdgeSlashAppeal(id, true);
    }

    function test_resolve_notAppealedReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");

        vm.prank(owner);
        vm.expectRevert(abi.encodeWithSelector(BunkerEdgeRegistry.ProposalNotAppealed.selector, id));
        registry.resolveEdgeSlashAppeal(id, true);
    }

    function test_resolve_twiceReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(owner);
        registry.setSlashingEnabled(true);

        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
        vm.prank(edge1);
        registry.appealEdgeSlash(id);

        vm.startPrank(owner);
        registry.resolveEdgeSlashAppeal(id, false);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEdgeRegistry.ProposalAlreadyResolved.selector, id)
        );
        registry.resolveEdgeSlashAppeal(id, false);
        vm.stopPrank();
    }

    function test_resolve_upheldWhileSlashingDisabledReverts() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(owner);
        registry.setSlashingEnabled(true);

        vm.prank(slasher);
        uint256 id = registry.proposeEdgeSlash(edge1, 1_000_000e18, "downtime");
        vm.prank(edge1);
        registry.appealEdgeSlash(id);

        // Disable slashing before resolving — upheld must revert.
        vm.startPrank(owner);
        registry.setSlashingEnabled(false);
        vm.expectRevert(BunkerEdgeRegistry.SlashingNotEnabled.selector);
        registry.resolveEdgeSlashAppeal(id, true);
        vm.stopPrank();
    }

    // ================================================================
    //  10. setMinStakeForEdge / setAppealWindow / setSlashingEnabled
    // ================================================================

    function test_setMinStake_byOwner() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true, address(registry));
        emit BunkerEdgeRegistry.MinStakeUpdated(DEFAULT_MIN_EDGE_STAKE, GOLD_MIN);
        registry.setMinStakeForEdge(GOLD_MIN);
        assertEq(registry.minStakeForEdge(), GOLD_MIN);
    }

    function test_setMinStake_nonOwnerReverts() public {
        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(
                bytes4(keccak256("OwnableUnauthorizedAccount(address)")), stranger
            )
        );
        registry.setMinStakeForEdge(GOLD_MIN);
    }

    function test_setMinStake_belowFloorReverts() public {
        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEdgeRegistry.InvalidMinStake.selector, STARTER_MIN - 1)
        );
        registry.setMinStakeForEdge(STARTER_MIN - 1);
    }

    function test_setMinStake_aboveCeilingReverts() public {
        vm.prank(owner);
        vm.expectRevert(
            abi.encodeWithSelector(BunkerEdgeRegistry.InvalidMinStake.selector, PLATINUM_MIN + 1)
        );
        registry.setMinStakeForEdge(PLATINUM_MIN + 1);
    }

    function test_setMinStake_atFloorSucceeds() public {
        vm.prank(owner);
        registry.setMinStakeForEdge(STARTER_MIN);
        assertEq(registry.minStakeForEdge(), STARTER_MIN);
    }

    function test_setMinStake_atCeilingSucceeds() public {
        vm.prank(owner);
        registry.setMinStakeForEdge(PLATINUM_MIN);
        assertEq(registry.minStakeForEdge(), PLATINUM_MIN);
    }

    function test_setAppealWindow_byOwner() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true, address(registry));
        emit BunkerEdgeRegistry.AppealWindowUpdated(7 days);
        registry.setAppealWindow(7 days);
        assertEq(registry.appealWindow(), 7 days);
    }

    function test_setAppealWindow_belowFloorReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerEdgeRegistry.InvalidAppealWindow.selector);
        registry.setAppealWindow(12 hours - 1);
    }

    function test_setAppealWindow_aboveCeilingReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerEdgeRegistry.InvalidAppealWindow.selector);
        registry.setAppealWindow(14 days + 1);
    }

    function test_setAppealWindow_nonOwnerReverts() public {
        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(
                bytes4(keccak256("OwnableUnauthorizedAccount(address)")), stranger
            )
        );
        registry.setAppealWindow(7 days);
    }

    function test_setSlashingEnabled_byOwner() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true, address(registry));
        emit BunkerEdgeRegistry.SlashingEnabledUpdated(true);
        registry.setSlashingEnabled(true);
        assertTrue(registry.slashingEnabled());
    }

    function test_setSlashingEnabled_nonOwnerReverts() public {
        vm.prank(stranger);
        vm.expectRevert(
            abi.encodeWithSelector(
                bytes4(keccak256("OwnableUnauthorizedAccount(address)")), stranger
            )
        );
        registry.setSlashingEnabled(true);
    }

    // ================================================================
    //  11. isActiveEdgeProvider view
    // ================================================================

    function test_isActive_falseWhenUnregistered() public view {
        assertFalse(registry.isActiveEdgeProvider(edge1));
    }

    function test_isActive_trueAfterRegister() public {
        _register(edge1, NODE_ID_1, REGION_US);
        assertTrue(registry.isActiveEdgeProvider(edge1));
    }

    function test_isActive_falseAfterDeregister() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(edge1);
        registry.deregisterEdgeProvider();
        assertFalse(registry.isActiveEdgeProvider(edge1));
    }

    function test_isActive_falseWhenFrozen() public {
        _register(edge1, NODE_ID_1, REGION_US);
        vm.prank(slasher);
        registry.freezeEdgeProvider(edge1);
        assertFalse(registry.isActiveEdgeProvider(edge1));
    }

    // ================================================================
    //  12. Ownership transfer syncs admin role
    // ================================================================

    function test_transferOwnership_syncsAdminRole() public {
        address newOwner = makeAddr("newOwner");

        vm.prank(owner);
        registry.transferOwnership(newOwner);
        vm.prank(newOwner);
        registry.acceptOwnership();

        assertEq(registry.owner(), newOwner);
        assertTrue(registry.hasRole(registry.DEFAULT_ADMIN_ROLE(), newOwner));
        assertFalse(registry.hasRole(registry.DEFAULT_ADMIN_ROLE(), owner));
    }

    // ================================================================
    //  13. FUZZ
    // ================================================================

    function testFuzz_registerAndQuery(bytes32 nodeId, bytes32 region, uint128 stakeAmount)
        public
    {
        // Bound the stake to a sane mintable range, then skip below-threshold runs.
        stakeAmount = uint128(bound(uint256(stakeAmount), 0, 1_500_000_000e18));
        vm.assume(nodeId != bytes32(0));
        vm.assume(registry.nodeIdToEdgeProvider(nodeId) == address(0));

        address fuzzEdge = makeAddr("fuzzEdge");
        vm.prank(owner);
        token.mint(fuzzEdge, uint256(stakeAmount) + 1);

        if (stakeAmount > 0) {
            vm.startPrank(fuzzEdge);
            token.approve(address(staking), type(uint256).max);
            // staking requires at least the Starter minimum to register.
            if (stakeAmount >= STARTER_MIN) {
                staking.stake(stakeAmount);
            }
            vm.stopPrank();
        }

        if (stakeAmount < DEFAULT_MIN_EDGE_STAKE) {
            vm.prank(fuzzEdge);
            vm.expectRevert(
                abi.encodeWithSelector(
                    BunkerEdgeRegistry.InsufficientStake.selector,
                    stakeAmount >= STARTER_MIN ? uint256(stakeAmount) : 0,
                    DEFAULT_MIN_EDGE_STAKE
                )
            );
            registry.registerEdgeProvider(nodeId, region, ENDPOINT_1, TLS_HASH_1);
            return;
        }

        vm.prank(fuzzEdge);
        registry.registerEdgeProvider(nodeId, region, ENDPOINT_1, TLS_HASH_1);

        assertTrue(registry.isActiveEdgeProvider(fuzzEdge));
        BunkerEdgeRegistry.EdgeProviderInfo memory info = registry.getEdgeProviderInfo(fuzzEdge);
        assertEq(info.nodeId, nodeId);
        assertEq(info.region, region);
        assertEq(registry.nodeIdToEdgeProvider(nodeId), fuzzEdge);
    }
}
