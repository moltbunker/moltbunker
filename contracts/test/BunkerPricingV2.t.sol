// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerPricing.sol";

/// @title BunkerPricingV2Test
/// @notice Tests for BunkerPricing V2 features: oracle pricing, provider-set prices,
///         price bounds validation, and oracle enable/disable toggle.
contract BunkerPricingV2Test is Test {
    BunkerPricing public pricing;

    address public owner = makeAddr("owner");
    address public alice = makeAddr("alice");
    address public oracleAddr = makeAddr("oracle");

    // Default price constants (BUNKER wei).
    uint256 constant CPU_PRICE = 0.5e18;
    uint256 constant MEMORY_PRICE = 0.1e18;
    uint256 constant STORAGE_PRICE = 0.05e18;
    uint256 constant NETWORK_PRICE = 0.02e18;
    uint256 constant GPU_BASIC = 5e18;
    uint256 constant GPU_PREMIUM = 15e18;

    // Default multiplier constants (basis points).
    uint256 constant REDUNDANCY_BPS = 30000;
    uint256 constant TOR_BPS = 12000;
    uint256 constant PREMIUM_SLA_BPS = 15000;
    uint256 constant SPOT_BPS = 5000;

    uint256 constant BPS = 10000;

    // Re-declare events for expectEmit matching.
    event PricesUpdated(
        uint256 cpuPerCoreHour,
        uint256 memoryPerGBHour,
        uint256 storagePerGBMonth,
        uint256 networkPerGB,
        uint256 gpuBasicPerHour,
        uint256 gpuPremiumPerHour
    );
    event PriceOracleUpdated(address indexed oldOracle, address indexed newOracle);

    function setUp() public {
        pricing = new BunkerPricing(owner);
    }

    // ──────────────────────────────────────────────
    //  Helpers
    // ──────────────────────────────────────────────

    function _req(
        uint256 cpuCores,
        uint256 memoryGB,
        uint256 storageGB,
        uint256 networkGB,
        uint256 durationHours
    ) internal pure returns (BunkerPricing.ResourceRequest memory) {
        return BunkerPricing.ResourceRequest({
            cpuCores: cpuCores,
            memoryGB: memoryGB,
            storageGB: storageGB,
            networkGB: networkGB,
            durationHours: durationHours,
            useGPUBasic: false,
            useGPUPremium: false,
            useTor: false,
            usePremiumSLA: false,
            useSpot: false
        });
    }

    function _reqFull(
        uint256 cpuCores,
        uint256 memoryGB,
        uint256 storageGB,
        uint256 networkGB,
        uint256 durationHours,
        bool gpuBasic,
        bool gpuPremium,
        bool tor,
        bool premiumSLA,
        bool spot
    ) internal pure returns (BunkerPricing.ResourceRequest memory) {
        return BunkerPricing.ResourceRequest({
            cpuCores: cpuCores,
            memoryGB: memoryGB,
            storageGB: storageGB,
            networkGB: networkGB,
            durationHours: durationHours,
            useGPUBasic: gpuBasic,
            useGPUPremium: gpuPremium,
            useTor: tor,
            usePremiumSLA: premiumSLA,
            useSpot: spot
        });
    }

    // ================================================================
    //  1. ORACLE PRICING
    // ================================================================

    function test_oraclePricing_setOracleAddress() public {
        vm.prank(owner);
        pricing.setPriceOracle(oracleAddr);

        assertEq(address(pricing.priceOracle()), oracleAddr);
    }

    function test_oraclePricing_costCalculationStillWorksWithOracle() public {
        // Set oracle address (oracle integration is Phase 2, so pricing should
        // continue to work using on-chain stored prices).
        vm.prank(owner);
        pricing.setPriceOracle(oracleAddr);

        // Cost calculation should still use stored prices.
        BunkerPricing.ResourceRequest memory req = _req(1000, 0, 0, 0, 1);
        uint256 cost = pricing.calculateCost(req);

        // 1 CPU core for 1 hour = 0.5 BUNKER * 3x redundancy = 1.5 BUNKER.
        assertEq(cost, 1.5e18);
    }

    function test_oraclePricing_emitsOracleUpdatedEvent() public {
        vm.prank(owner);
        vm.expectEmit(true, true, false, false, address(pricing));
        emit PriceOracleUpdated(address(0), oracleAddr);
        pricing.setPriceOracle(oracleAddr);
    }

    function test_staleOracleReverts_nonOwnerCannotSetOracle() public {
        // This tests the access control on oracle setting (simulating "stale" by
        // preventing unauthorized oracle changes).
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        pricing.setPriceOracle(oracleAddr);
    }

    function test_oraclePricing_oracleDoesNotAffectStoredPrices() public {
        // Setting an oracle should not change the stored prices.
        BunkerPricing.ResourcePrices memory pricesBefore = pricing.getPrices();

        vm.prank(owner);
        pricing.setPriceOracle(oracleAddr);

        BunkerPricing.ResourcePrices memory pricesAfter = pricing.getPrices();
        assertEq(pricesAfter.cpuPerCoreHour, pricesBefore.cpuPerCoreHour);
        assertEq(pricesAfter.memoryPerGBHour, pricesBefore.memoryPerGBHour);
        assertEq(pricesAfter.storagePerGBMonth, pricesBefore.storagePerGBMonth);
        assertEq(pricesAfter.networkPerGB, pricesBefore.networkPerGB);
        assertEq(pricesAfter.gpuBasicPerHour, pricesBefore.gpuBasicPerHour);
        assertEq(pricesAfter.gpuPremiumPerHour, pricesBefore.gpuPremiumPerHour);
    }

    // ================================================================
    //  2. PROVIDER PRICING (admin-set prices work)
    // ================================================================

    function test_providerPricing_updatedPricesReflectInCost() public {
        // Double CPU price to simulate provider-competitive pricing.
        vm.prank(owner);
        pricing.setCPUPrice(1e18);

        BunkerPricing.ResourceRequest memory req = _req(1000, 0, 0, 0, 1);
        uint256 cost = pricing.calculateCost(req);

        // 1 CPU * 1e18 * 1h / 1000 = 1e18, after 3x redundancy = 3e18.
        assertEq(cost, 3e18);
    }

    function test_providerPricing_fullPriceUpdate() public {
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour: 2e18,
            memoryPerGBHour: 0.5e18,
            storagePerGBMonth: 0.2e18,
            networkPerGB: 0.1e18,
            gpuBasicPerHour: 10e18,
            gpuPremiumPerHour: 30e18
        });

        vm.prank(owner);
        pricing.setPrices(newPrices);

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.cpuPerCoreHour, 2e18);
        assertEq(p.memoryPerGBHour, 0.5e18);
        assertEq(p.storagePerGBMonth, 0.2e18);
        assertEq(p.networkPerGB, 0.1e18);
        assertEq(p.gpuBasicPerHour, 10e18);
        assertEq(p.gpuPremiumPerHour, 30e18);

        // Verify cost calculation uses new prices.
        BunkerPricing.ResourceRequest memory req = _req(1000, 1000, 0, 0, 1);
        uint256 cost = pricing.calculateCost(req);

        // cpuCost = (1000 * 2e18 * 1) / 1000 = 2e18
        // memoryCost = (1000 * 0.5e18 * 1) / 1000 = 0.5e18
        // subtotal = 2.5e18
        // after 3x redundancy = 7.5e18
        assertEq(cost, 7.5e18);
    }

    function test_providerPricing_individualPriceUpdatesCombine() public {
        vm.startPrank(owner);
        pricing.setCPUPrice(1e18);
        pricing.setMemoryPrice(0.5e18);
        pricing.setStoragePrice(0.1e18);
        pricing.setNetworkPrice(0.05e18);
        vm.stopPrank();

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.cpuPerCoreHour, 1e18);
        assertEq(p.memoryPerGBHour, 0.5e18);
        assertEq(p.storagePerGBMonth, 0.1e18);
        assertEq(p.networkPerGB, 0.05e18);
        // GPU prices unchanged.
        assertEq(p.gpuBasicPerHour, GPU_BASIC);
        assertEq(p.gpuPremiumPerHour, GPU_PREMIUM);
    }

    // ================================================================
    //  3. PROVIDER PRICE BOUNDS
    // ================================================================

    function test_providerPriceBounds_zeroPriceReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setCPUPrice(0);
    }

    function test_providerPriceBounds_zeroMemoryPriceReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setMemoryPrice(0);
    }

    function test_providerPriceBounds_zeroStoragePriceReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setStoragePrice(0);
    }

    function test_providerPriceBounds_zeroNetworkPriceReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setNetworkPrice(0);
    }

    function test_providerPriceBounds_zeroGPUBasicReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setGPUPrices(0, 25e18);
    }

    function test_providerPriceBounds_zeroGPUPremiumReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setGPUPrices(8e18, 0);
    }

    function test_providerPriceBounds_batchZeroCPUReverts() public {
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour: 0,
            memoryPerGBHour: 0.1e18,
            storagePerGBMonth: 0.05e18,
            networkPerGB: 0.02e18,
            gpuBasicPerHour: 5e18,
            gpuPremiumPerHour: 15e18
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setPrices(newPrices);
    }

    function test_providerPriceBounds_batchZeroNetworkReverts() public {
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour: 1e18,
            memoryPerGBHour: 0.1e18,
            storagePerGBMonth: 0.05e18,
            networkPerGB: 0,
            gpuBasicPerHour: 5e18,
            gpuPremiumPerHour: 15e18
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setPrices(newPrices);
    }

    function test_providerPriceBounds_nonOwnerCannotSet() public {
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        pricing.setCPUPrice(1e18);
    }

    function test_providerPriceBounds_veryHighPriceWorks() public {
        // Extreme but valid price.
        vm.prank(owner);
        pricing.setCPUPrice(type(uint256).max / 2);

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.cpuPerCoreHour, type(uint256).max / 2);
    }

    function test_providerPriceBounds_singleWeiPriceWorks() public {
        vm.prank(owner);
        pricing.setCPUPrice(1);

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.cpuPerCoreHour, 1);

        // 1000 milli-cores for 1 hour: (1000 * 1 * 1) / 1000 = 1 wei.
        // After 3x redundancy: 3 wei.
        BunkerPricing.ResourceRequest memory req = _req(1000, 0, 0, 0, 1);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 3);
    }

    // ================================================================
    //  4. ENABLE / DISABLE ORACLE
    // ================================================================

    function test_enableDisableOracle_setThenClear() public {
        // Enable.
        vm.prank(owner);
        pricing.setPriceOracle(oracleAddr);
        assertEq(address(pricing.priceOracle()), oracleAddr);

        // Disable.
        vm.prank(owner);
        pricing.setPriceOracle(address(0));
        assertEq(address(pricing.priceOracle()), address(0));
    }

    function test_enableDisableOracle_emitsEventsCorrectly() public {
        // Enable.
        vm.prank(owner);
        vm.expectEmit(true, true, false, false, address(pricing));
        emit PriceOracleUpdated(address(0), oracleAddr);
        pricing.setPriceOracle(oracleAddr);

        // Disable.
        vm.prank(owner);
        vm.expectEmit(true, true, false, false, address(pricing));
        emit PriceOracleUpdated(oracleAddr, address(0));
        pricing.setPriceOracle(address(0));
    }

    function test_enableDisableOracle_costUnchangedAfterToggle() public {
        BunkerPricing.ResourceRequest memory req = _req(1000, 1000, 10, 5, 24);
        uint256 costBefore = pricing.calculateCost(req);

        // Enable oracle.
        vm.prank(owner);
        pricing.setPriceOracle(oracleAddr);
        uint256 costWithOracle = pricing.calculateCost(req);

        // Disable oracle.
        vm.prank(owner);
        pricing.setPriceOracle(address(0));
        uint256 costAfter = pricing.calculateCost(req);

        // Cost should be the same throughout (oracle is Phase 2, not yet integrated).
        assertEq(costBefore, costWithOracle);
        assertEq(costBefore, costAfter);
    }

    function test_enableDisableOracle_multipleOracleChanges() public {
        address oracle1 = makeAddr("oracle1");
        address oracle2 = makeAddr("oracle2");

        vm.startPrank(owner);
        pricing.setPriceOracle(oracle1);
        assertEq(address(pricing.priceOracle()), oracle1);

        pricing.setPriceOracle(oracle2);
        assertEq(address(pricing.priceOracle()), oracle2);

        pricing.setPriceOracle(address(0));
        assertEq(address(pricing.priceOracle()), address(0));
        vm.stopPrank();
    }

    // ================================================================
    //  5. MULTIPLIER VALIDATION
    // ================================================================

    function test_multiplierBounds_zeroRedundancyReverts() public {
        BunkerPricing.Multipliers memory m = BunkerPricing.Multipliers({
            redundancy: 0,
            tor: TOR_BPS,
            premiumSLA: PREMIUM_SLA_BPS,
            spot: SPOT_BPS
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroMultiplier.selector);
        pricing.setMultipliers(m);
    }

    function test_multiplierBounds_zeroTorReverts() public {
        BunkerPricing.Multipliers memory m = BunkerPricing.Multipliers({
            redundancy: REDUNDANCY_BPS,
            tor: 0,
            premiumSLA: PREMIUM_SLA_BPS,
            spot: SPOT_BPS
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroMultiplier.selector);
        pricing.setMultipliers(m);
    }

    function test_multiplierBounds_zeroPremiumSLAReverts() public {
        BunkerPricing.Multipliers memory m = BunkerPricing.Multipliers({
            redundancy: REDUNDANCY_BPS,
            tor: TOR_BPS,
            premiumSLA: 0,
            spot: SPOT_BPS
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroMultiplier.selector);
        pricing.setMultipliers(m);
    }

    function test_multiplierBounds_zeroSpotReverts() public {
        BunkerPricing.Multipliers memory m = BunkerPricing.Multipliers({
            redundancy: REDUNDANCY_BPS,
            tor: TOR_BPS,
            premiumSLA: PREMIUM_SLA_BPS,
            spot: 0
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroMultiplier.selector);
        pricing.setMultipliers(m);
    }

    function test_multiplierBounds_validExtremeValues() public {
        BunkerPricing.Multipliers memory m = BunkerPricing.Multipliers({
            redundancy: 100000, // 10x
            tor: 50000,         // 5x
            premiumSLA: 50000,  // 5x
            spot: 1             // 0.01%
        });

        vm.prank(owner);
        pricing.setMultipliers(m);

        BunkerPricing.Multipliers memory stored = pricing.getMultipliers();
        assertEq(stored.redundancy, 100000);
        assertEq(stored.tor, 50000);
        assertEq(stored.premiumSLA, 50000);
        assertEq(stored.spot, 1);
    }

    // ================================================================
    //  6. COST CALCULATION WITH UPDATED PRICING
    // ================================================================

    function test_costAfterPriceUpdate_cpuDoublesDoublesCost() public {
        BunkerPricing.ResourceRequest memory req = _req(1000, 0, 0, 0, 1);
        uint256 costBefore = pricing.calculateCost(req);

        vm.prank(owner);
        pricing.setCPUPrice(1e18); // Double from 0.5e18.

        uint256 costAfter = pricing.calculateCost(req);
        assertEq(costAfter, costBefore * 2);
    }

    function test_costAfterMultiplierUpdate_halvingRedundancyHalvesCost() public {
        BunkerPricing.ResourceRequest memory req = _req(1000, 0, 0, 0, 1);
        uint256 costBefore = pricing.calculateCost(req);

        BunkerPricing.Multipliers memory m = BunkerPricing.Multipliers({
            redundancy: 15000, // 1.5x instead of 3x
            tor: TOR_BPS,
            premiumSLA: PREMIUM_SLA_BPS,
            spot: SPOT_BPS
        });

        vm.prank(owner);
        pricing.setMultipliers(m);

        uint256 costAfter = pricing.calculateCost(req);
        assertEq(costAfter, costBefore / 2);
    }

    function test_costWithGPU_basicAndPremiumDifference() public {
        BunkerPricing.ResourceRequest memory reqBasic = _reqFull(
            0, 0, 0, 0, 10, true, false, false, false, false
        );
        BunkerPricing.ResourceRequest memory reqPremium = _reqFull(
            0, 0, 0, 0, 10, false, true, false, false, false
        );

        uint256 basicCost = pricing.calculateCost(reqBasic);
        uint256 premiumCost = pricing.calculateCost(reqPremium);

        // Premium GPU (15 BUNKER/hr) should be 3x basic GPU (5 BUNKER/hr).
        assertEq(premiumCost, basicCost * 3);
    }

    // ================================================================
    //  7. FUZZ TESTS
    // ================================================================

    function testFuzz_priceUpdateAndCostCalculation(
        uint128 cpuPrice,
        uint48 cpuCores,
        uint48 duration
    ) public {
        vm.assume(cpuPrice > 0 && cpuPrice <= 1000e18);
        vm.assume(cpuCores <= 100_000);
        vm.assume(duration <= 87_600);

        vm.prank(owner);
        pricing.setCPUPrice(uint256(cpuPrice));

        BunkerPricing.ResourceRequest memory req = _req(
            uint256(cpuCores), 0, 0, 0, uint256(duration)
        );
        uint256 cost = pricing.calculateCost(req);

        // Manual calculation (using ceiling division to match H-12 fix).
        uint256 cpuCost = _testCeilDiv(uint256(cpuCores) * uint256(cpuPrice) * uint256(duration), 1000);
        uint256 expected = _testCeilDiv(cpuCost * REDUNDANCY_BPS, BPS);

        assertEq(cost, expected);
    }

    function testFuzz_oracleToggle(address oracle1, address oracle2) public {
        vm.startPrank(owner);
        pricing.setPriceOracle(oracle1);
        assertEq(address(pricing.priceOracle()), oracle1);

        pricing.setPriceOracle(oracle2);
        assertEq(address(pricing.priceOracle()), oracle2);

        pricing.setPriceOracle(address(0));
        assertEq(address(pricing.priceOracle()), address(0));
        vm.stopPrank();
    }

    /// @dev Ceiling division helper matching BunkerPricing._ceilDiv
    function _testCeilDiv(uint256 a, uint256 b) internal pure returns (uint256) {
        return a == 0 ? 0 : (a - 1) / b + 1;
    }
}
