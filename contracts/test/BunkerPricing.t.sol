// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../src/BunkerPricing.sol";

contract BunkerPricingTest is Test {
    BunkerPricing public pricing;

    address public owner = makeAddr("owner");
    address public alice = makeAddr("alice");

    // Re-declare events for expectEmit matching.
    event PricesUpdated(
        uint256 cpuPerCoreHour,
        uint256 memoryPerGBHour,
        uint256 storagePerGBMonth,
        uint256 networkPerGB,
        uint256 gpuBasicPerHour,
        uint256 gpuPremiumPerHour
    );
    event MultipliersUpdated(
        uint256 redundancy,
        uint256 tor,
        uint256 premiumSLA,
        uint256 spot
    );
    event PriceOracleUpdated(address indexed oldOracle, address indexed newOracle);

    // Default price constants (BUNKER wei).
    uint256 constant CPU_PRICE       = 0.5e18;   // 500000000000000000
    uint256 constant MEMORY_PRICE    = 0.1e18;   // 100000000000000000
    uint256 constant STORAGE_PRICE   = 0.05e18;  // 50000000000000000
    uint256 constant NETWORK_PRICE   = 0.02e18;  // 20000000000000000
    uint256 constant GPU_BASIC       = 5e18;      // 5000000000000000000
    uint256 constant GPU_PREMIUM     = 15e18;     // 15000000000000000000

    // Default multiplier constants (basis points).
    uint256 constant REDUNDANCY_BPS  = 30000; // 3.0x
    uint256 constant TOR_BPS         = 12000; // 1.2x
    uint256 constant PREMIUM_SLA_BPS = 15000; // 1.5x
    uint256 constant SPOT_BPS        = 5000;  // 0.5x

    uint256 constant BPS = 10000;

    function setUp() public {
        pricing = new BunkerPricing(owner);
    }

    // -----------------------------------------------------------------------
    //  Helper: Build a ResourceRequest with sensible defaults (all zeros/false).
    // -----------------------------------------------------------------------

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

    // =======================================================================
    //  1. Deployment Tests
    // =======================================================================

    function test_Deployment_DefaultPrices() public view {
        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.cpuPerCoreHour,    CPU_PRICE,      "cpu default");
        assertEq(p.memoryPerGBHour,   MEMORY_PRICE,   "memory default");
        assertEq(p.storagePerGBMonth, STORAGE_PRICE,   "storage default");
        assertEq(p.networkPerGB,      NETWORK_PRICE,   "network default");
        assertEq(p.gpuBasicPerHour,   GPU_BASIC,       "gpu basic default");
        assertEq(p.gpuPremiumPerHour, GPU_PREMIUM,     "gpu premium default");
    }

    function test_Deployment_DefaultMultipliers() public view {
        BunkerPricing.Multipliers memory m = pricing.getMultipliers();
        assertEq(m.redundancy,  REDUNDANCY_BPS,  "redundancy default");
        assertEq(m.tor,         TOR_BPS,         "tor default");
        assertEq(m.premiumSLA,  PREMIUM_SLA_BPS, "premiumSLA default");
        assertEq(m.spot,        SPOT_BPS,        "spot default");
    }

    function test_Deployment_OwnerIsCorrect() public view {
        assertEq(pricing.owner(), owner);
    }

    function test_Deployment_VersionIs110() public view {
        assertEq(pricing.VERSION(), "1.2.0");
    }

    function test_Deployment_BpsDenominator() public view {
        assertEq(pricing.BPS_DENOMINATOR(), 10000);
    }

    function test_Deployment_PriceOracleIsZero() public view {
        assertEq(address(pricing.priceOracle()), address(0));
    }

    // =======================================================================
    //  2. Cost Calculation - Basic
    // =======================================================================

    /// @notice 1 CPU core for 1 hour, no extras.
    /// Math:
    ///   cpuCores = 1000 (milli), durationHours = 1
    ///   cpuCost  = (1000 * 0.5e18 * 1) / 1000 = 0.5e18
    ///   total before multipliers = 0.5e18
    ///   After redundancy (3x):   0.5e18 * 30000 / 10000 = 1.5e18
    function test_Cost_OneCpuOneHour() public view {
        BunkerPricing.ResourceRequest memory req = _req(1000, 0, 0, 0, 1);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 1.5e18);
    }

    /// @notice 2 CPU cores + 4 GB memory for 24 hours, no extras.
    /// Math:
    ///   cpuCost    = (2000 * 0.5e18 * 24) / 1000 = 24e18
    ///   memoryCost = (4000 * 0.1e18 * 24) / 1000 = 9.6e18
    ///   subtotal   = 33.6e18
    ///   After redundancy (3x): 33.6e18 * 3 = 100.8e18
    function test_Cost_TwoCpuFourGbMemory24Hours() public view {
        BunkerPricing.ResourceRequest memory req = _req(2000, 4000, 0, 0, 24);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 100.8e18);
    }

    /// @notice Storage cost: 10 GB for 730 hours (1 month).
    /// Math:
    ///   storageCost = (10 * 0.05e18 * 730) / 730 = 0.5e18
    ///   After redundancy (3x): 0.5e18 * 3 = 1.5e18
    function test_Cost_StorageOneMonth() public view {
        BunkerPricing.ResourceRequest memory req = _req(0, 0, 10, 0, 730);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 1.5e18);
    }

    /// @notice Storage cost for partial month (365 hours = 0.5 months).
    /// Math:
    ///   storageCost = (10 * 0.05e18 * 365) / 730 = 0.25e18
    ///   After redundancy (3x): 0.25e18 * 3 = 0.75e18
    function test_Cost_StorageHalfMonth() public view {
        BunkerPricing.ResourceRequest memory req = _req(0, 0, 10, 0, 365);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 0.75e18);
    }

    /// @notice Network cost: 100 GB egress. No duration factor.
    /// Math:
    ///   networkCost = 100 * 0.02e18 = 2e18
    ///   After redundancy (3x): 2e18 * 3 = 6e18
    function test_Cost_NetworkOnly() public view {
        BunkerPricing.ResourceRequest memory req = _req(0, 0, 0, 100, 0);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 6e18);
    }

    /// @notice Network cost does not depend on duration.
    /// Both requests should yield the same network cost component.
    function test_Cost_NetworkIndependentOfDuration() public view {
        BunkerPricing.ResourceRequest memory req1 = _req(0, 0, 0, 50, 1);
        BunkerPricing.ResourceRequest memory req2 = _req(0, 0, 0, 50, 1000);
        assertEq(pricing.calculateCost(req1), pricing.calculateCost(req2));
    }

    /// @notice GPU basic cost: 1 hour.
    /// Math:
    ///   gpuCost = 5e18 * 1 = 5e18
    ///   After redundancy (3x): 5e18 * 3 = 15e18
    function test_Cost_GpuBasicOneHour() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(0, 0, 0, 0, 1, true, false, false, false, false);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 15e18);
    }

    /// @notice GPU premium cost: 10 hours.
    /// Math:
    ///   gpuCost = 15e18 * 10 = 150e18
    ///   After redundancy (3x): 150e18 * 3 = 450e18
    function test_Cost_GpuPremiumTenHours() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(0, 0, 0, 0, 10, false, true, false, false, false);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 450e18);
    }

    /// @notice When both GPU flags are set, only basic GPU applies (else-if logic).
    /// Math:
    ///   gpuCost = 5e18 * 1 = 5e18 (basic, not premium)
    ///   After redundancy (3x): 15e18
    function test_Cost_GpuBothFlags_BasicTakesPrecedence() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(0, 0, 0, 0, 1, true, true, false, false, false);
        uint256 cost = pricing.calculateCost(req);
        // Should be basic GPU cost, not premium
        assertEq(cost, 15e18);
    }

    /// @notice Zero resources = zero cost.
    function test_Cost_ZeroResources() public view {
        BunkerPricing.ResourceRequest memory req = _req(0, 0, 0, 0, 0);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 0);
    }

    // =======================================================================
    //  3. Cost Calculation - Multipliers
    // =======================================================================

    /// @notice Tor multiplier: 1.2x on top of redundancy (3x).
    /// Math:
    ///   cpuCost = (1000 * 0.5e18 * 1) / 1000 = 0.5e18
    ///   After redundancy: 0.5e18 * 30000 / 10000 = 1.5e18
    ///   After tor:        1.5e18 * 12000 / 10000 = 1.8e18
    function test_Cost_WithTor() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(1000, 0, 0, 0, 1, false, false, true, false, false);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 1.8e18);
    }

    /// @notice PremiumSLA multiplier: 1.5x on top of redundancy (3x).
    /// Math:
    ///   cpuCost = 0.5e18 (same base)
    ///   After redundancy: 1.5e18
    ///   After premiumSLA: 1.5e18 * 15000 / 10000 = 2.25e18
    function test_Cost_WithPremiumSLA() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(1000, 0, 0, 0, 1, false, false, false, true, false);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 2.25e18);
    }

    /// @notice Spot multiplier: 0.5x on top of redundancy (3x).
    /// Math:
    ///   cpuCost = 0.5e18
    ///   After redundancy: 1.5e18
    ///   After spot:       1.5e18 * 5000 / 10000 = 0.75e18
    function test_Cost_WithSpot() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(1000, 0, 0, 0, 1, false, false, false, false, true);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 0.75e18);
    }

    /// @notice Tor + PremiumSLA combined.
    /// Math:
    ///   cpuCost = 0.5e18
    ///   After redundancy: 1.5e18
    ///   After tor:        1.5e18 * 12000 / 10000 = 1.8e18
    ///   After premiumSLA: 1.8e18 * 15000 / 10000 = 2.7e18
    function test_Cost_TorPlusPremiumSLA() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(1000, 0, 0, 0, 1, false, false, true, true, false);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 2.7e18);
    }

    /// @notice All optional multipliers combined: tor * premiumSLA * spot.
    /// Math:
    ///   cpuCost = 0.5e18
    ///   After redundancy:  1.5e18
    ///   After tor:         1.5e18   * 12000 / 10000 = 1.8e18
    ///   After premiumSLA:  1.8e18   * 15000 / 10000 = 2.7e18
    ///   After spot:        2.7e18   * 5000  / 10000 = 1.35e18
    function test_Cost_AllMultipliersCombined() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(1000, 0, 0, 0, 1, false, false, true, true, true);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 1.35e18);
    }

    // =======================================================================
    //  4. Cost Calculation - Worked Examples
    // =======================================================================

    /// @notice Example 1: Basic web server deployment.
    ///   2 CPU cores, 4 GB memory, 20 GB storage, 10 GB network, 720 hours (1 month).
    ///   No GPU, no optional flags.
    ///
    /// Step-by-step:
    ///   cpuCores = 2000 (milli), memoryGB = 4000 (milli)
    ///   cpuCost    = (2000 * 0.5e18  * 720) / 1000 = 720e18
    ///   memoryCost = (4000 * 0.1e18  * 720) / 1000 = 288e18
    ///   storageCost= (20   * 0.05e18 * 720) / 730  = 720e18 / 730
    ///              = 986301369863013698 (integer division)
    ///   networkCost= 10 * 0.02e18 = 0.2e18
    ///
    ///   subtotal = 720e18 + 288e18 + 986301369863013698 + 0.2e18
    ///            = 1009186301369863013698
    ///
    ///   After redundancy (3x): 1009186301369863013698 * 30000 / 10000
    ///            = 3027558904109589041094
    function test_Cost_WorkedExample_BasicWebServer() public view {
        BunkerPricing.ResourceRequest memory req = _req(2000, 4000, 20, 10, 720);
        uint256 cost = pricing.calculateCost(req);

        // Verify each component by computing in Solidity (using ceiling division
        // to match the H-12 fix in BunkerPricing).
        uint256 cpuCost     = _testCeilDiv(2000 * CPU_PRICE * 720, 1000);
        uint256 memoryCost  = _testCeilDiv(4000 * MEMORY_PRICE * 720, 1000);
        uint256 storageCost = _testCeilDiv(20 * STORAGE_PRICE * 720, 730);
        uint256 networkCost = 10 * NETWORK_PRICE;

        uint256 subtotal = cpuCost + memoryCost + storageCost + networkCost;
        uint256 expected = _testCeilDiv(subtotal * REDUNDANCY_BPS, BPS);

        assertEq(cost, expected, "web server total");
        assertEq(cpuCost,     720e18,                  "web server cpu");
        assertEq(memoryCost,  288e18,                  "web server memory");
        assertEq(networkCost, 0.2e18,                  "web server network");
    }

    /// @notice Example 2: GPU workload with premium SLA.
    ///   4 CPU cores, 16 GB memory, basic GPU, 24 hours, premiumSLA=true.
    ///   No storage, no network, no tor, no spot.
    ///
    /// Step-by-step:
    ///   cpuCores = 4000 (milli), memoryGB = 16000 (milli)
    ///   cpuCost    = (4000 * 0.5e18 * 24) / 1000 = 48e18
    ///   memoryCost = (16000 * 0.1e18 * 24) / 1000 = 38.4e18
    ///   gpuCost    = 5e18 * 24 = 120e18
    ///
    ///   subtotal   = 48e18 + 38.4e18 + 120e18 = 206.4e18
    ///   After redundancy (3x):   206.4e18 * 3 = 619.2e18
    ///   After premiumSLA (1.5x): 619.2e18 * 15000 / 10000 = 928.8e18
    function test_Cost_WorkedExample_GpuPremiumSLA() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(4000, 16000, 0, 0, 24, true, false, false, true, false);
        uint256 cost = pricing.calculateCost(req);

        uint256 cpuCost    = (4000 * CPU_PRICE * 24) / 1000;
        uint256 memoryCost = (16000 * MEMORY_PRICE * 24) / 1000;
        uint256 gpuCost    = GPU_BASIC * 24;

        uint256 subtotal = cpuCost + memoryCost + gpuCost;
        uint256 afterRedundancy = (subtotal * REDUNDANCY_BPS) / BPS;
        uint256 expected = (afterRedundancy * PREMIUM_SLA_BPS) / BPS;

        assertEq(cpuCost,    48e18,   "gpu example cpu");
        assertEq(memoryCost, 38.4e18, "gpu example memory");
        assertEq(gpuCost,    120e18,  "gpu example gpu");
        assertEq(cost,       expected, "gpu example total");
        assertEq(cost,       928.8e18, "gpu example exact");
    }

    /// @notice Example 3: Minimal spot instance with Tor.
    ///   0.5 CPU cores (500 milli), 0.5 GB memory (500 milli), 1 hour.
    ///   Tor + Spot.
    ///
    /// Step-by-step:
    ///   cpuCost    = (500 * 0.5e18 * 1) / 1000 = 0.25e18
    ///   memoryCost = (500 * 0.1e18 * 1) / 1000 = 0.05e18
    ///   subtotal   = 0.3e18
    ///   After redundancy (3x): 0.9e18
    ///   After tor (1.2x):      0.9e18 * 12000 / 10000 = 1.08e18
    ///   After spot (0.5x):     1.08e18 * 5000 / 10000 = 0.54e18
    function test_Cost_WorkedExample_SpotTor() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(500, 500, 0, 0, 1, false, false, true, false, true);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 0.54e18);
    }

    // =======================================================================
    //  5. Admin Price Updates
    // =======================================================================

    function test_SetPrices_Works() public {
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour:    1e18,
            memoryPerGBHour:   0.2e18,
            storagePerGBMonth: 0.1e18,
            networkPerGB:      0.05e18,
            gpuBasicPerHour:   10e18,
            gpuPremiumPerHour: 30e18
        });

        vm.prank(owner);
        pricing.setPrices(newPrices);

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.cpuPerCoreHour,    1e18);
        assertEq(p.memoryPerGBHour,   0.2e18);
        assertEq(p.storagePerGBMonth, 0.1e18);
        assertEq(p.networkPerGB,      0.05e18);
        assertEq(p.gpuBasicPerHour,   10e18);
        assertEq(p.gpuPremiumPerHour, 30e18);
    }

    function test_SetPrices_EmitsPricesUpdated() public {
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour:    1e18,
            memoryPerGBHour:   0.2e18,
            storagePerGBMonth: 0.1e18,
            networkPerGB:      0.05e18,
            gpuBasicPerHour:   10e18,
            gpuPremiumPerHour: 30e18
        });

        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit PricesUpdated(1e18, 0.2e18, 0.1e18, 0.05e18, 10e18, 30e18);
        pricing.setPrices(newPrices);
    }

    function test_SetPrices_NonOwnerReverts() public {
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour:    1e18,
            memoryPerGBHour:   0.2e18,
            storagePerGBMonth: 0.1e18,
            networkPerGB:      0.05e18,
            gpuBasicPerHour:   10e18,
            gpuPremiumPerHour: 30e18
        });

        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        pricing.setPrices(newPrices);
    }

    function test_SetPrices_ZeroCpuReverts() public {
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour:    0,       // zero CPU
            memoryPerGBHour:   0.2e18,
            storagePerGBMonth: 0.1e18,
            networkPerGB:      0.05e18,
            gpuBasicPerHour:   10e18,
            gpuPremiumPerHour: 30e18
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setPrices(newPrices);
    }

    function test_SetPrices_ZeroMemoryReverts() public {
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour:    1e18,
            memoryPerGBHour:   0,       // zero memory
            storagePerGBMonth: 0.1e18,
            networkPerGB:      0.05e18,
            gpuBasicPerHour:   10e18,
            gpuPremiumPerHour: 30e18
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setPrices(newPrices);
    }

    function test_SetPrices_ZeroStorageReverts() public {
        // All prices must be non-zero to prevent free resources.
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour:    1e18,
            memoryPerGBHour:   0.1e18,
            storagePerGBMonth: 0,
            networkPerGB:      0.05e18,
            gpuBasicPerHour:   5e18,
            gpuPremiumPerHour: 15e18
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setPrices(newPrices);
    }

    function test_SetPrices_ZeroGPUReverts() public {
        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour:    1e18,
            memoryPerGBHour:   0.1e18,
            storagePerGBMonth: 0.05e18,
            networkPerGB:      0.02e18,
            gpuBasicPerHour:   0,
            gpuPremiumPerHour: 15e18
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setPrices(newPrices);
    }

    // --- Individual price setters ---

    function test_SetCPUPrice_Works() public {
        vm.prank(owner);
        pricing.setCPUPrice(2e18);

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.cpuPerCoreHour, 2e18);
        // Other prices unchanged.
        assertEq(p.memoryPerGBHour, MEMORY_PRICE);
    }

    function test_SetCPUPrice_EmitsPricesUpdated() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit PricesUpdated(2e18, MEMORY_PRICE, STORAGE_PRICE, NETWORK_PRICE, GPU_BASIC, GPU_PREMIUM);
        pricing.setCPUPrice(2e18);
    }

    function test_SetCPUPrice_ZeroReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setCPUPrice(0);
    }

    function test_SetCPUPrice_NonOwnerReverts() public {
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        pricing.setCPUPrice(2e18);
    }

    function test_SetMemoryPrice_Works() public {
        vm.prank(owner);
        pricing.setMemoryPrice(0.5e18);

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.memoryPerGBHour, 0.5e18);
        assertEq(p.cpuPerCoreHour, CPU_PRICE); // unchanged
    }

    function test_SetMemoryPrice_ZeroReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setMemoryPrice(0);
    }

    function test_SetMemoryPrice_NonOwnerReverts() public {
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        pricing.setMemoryPrice(0.5e18);
    }

    function test_SetStoragePrice_Works() public {
        vm.prank(owner);
        pricing.setStoragePrice(0.2e18);

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.storagePerGBMonth, 0.2e18);
    }

    function test_SetStoragePrice_ZeroReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setStoragePrice(0);
    }

    function test_SetStoragePrice_NonOwnerReverts() public {
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        pricing.setStoragePrice(0.2e18);
    }

    function test_SetNetworkPrice_Works() public {
        vm.prank(owner);
        pricing.setNetworkPrice(0.1e18);

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.networkPerGB, 0.1e18);
    }

    function test_SetNetworkPrice_ZeroReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setNetworkPrice(0);
    }

    function test_SetNetworkPrice_NonOwnerReverts() public {
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        pricing.setNetworkPrice(0.1e18);
    }

    // --- GPU prices ---

    function test_SetGPUPrices_Works() public {
        vm.prank(owner);
        pricing.setGPUPrices(8e18, 25e18);

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.gpuBasicPerHour,   8e18);
        assertEq(p.gpuPremiumPerHour, 25e18);
    }

    function test_SetGPUPrices_EmitsPricesUpdated() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit PricesUpdated(CPU_PRICE, MEMORY_PRICE, STORAGE_PRICE, NETWORK_PRICE, 8e18, 25e18);
        pricing.setGPUPrices(8e18, 25e18);
    }

    function test_SetGPUPrices_ZeroBasicReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setGPUPrices(0, 25e18);
    }

    function test_SetGPUPrices_ZeroPremiumReverts() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroPrice.selector);
        pricing.setGPUPrices(8e18, 0);
    }

    function test_SetGPUPrices_NonOwnerReverts() public {
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        pricing.setGPUPrices(8e18, 25e18);
    }

    // --- Cost reflects updated prices ---

    function test_Cost_ReflectsUpdatedPrices() public {
        // Double CPU price.
        vm.prank(owner);
        pricing.setCPUPrice(1e18);

        // 1 CPU for 1 hour should now cost 2x the original.
        BunkerPricing.ResourceRequest memory req = _req(1000, 0, 0, 0, 1);
        uint256 cost = pricing.calculateCost(req);

        // cpuCost = (1000 * 1e18 * 1) / 1000 = 1e18
        // After redundancy: 1e18 * 3 = 3e18
        assertEq(cost, 3e18);
    }

    // =======================================================================
    //  6. Admin Multiplier Updates
    // =======================================================================

    function test_SetMultipliers_Works() public {
        BunkerPricing.Multipliers memory newMult = BunkerPricing.Multipliers({
            redundancy: 20000,  // 2x
            tor:        15000,  // 1.5x
            premiumSLA: 20000,  // 2x
            spot:       3000    // 0.3x
        });

        vm.prank(owner);
        pricing.setMultipliers(newMult);

        BunkerPricing.Multipliers memory m = pricing.getMultipliers();
        assertEq(m.redundancy,  20000);
        assertEq(m.tor,         15000);
        assertEq(m.premiumSLA,  20000);
        assertEq(m.spot,        3000);
    }

    function test_SetMultipliers_EmitsMultipliersUpdated() public {
        BunkerPricing.Multipliers memory newMult = BunkerPricing.Multipliers({
            redundancy: 20000,
            tor:        15000,
            premiumSLA: 20000,
            spot:       3000
        });

        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit MultipliersUpdated(20000, 15000, 20000, 3000);
        pricing.setMultipliers(newMult);
    }

    function test_SetMultipliers_ZeroRedundancyReverts() public {
        BunkerPricing.Multipliers memory newMult = BunkerPricing.Multipliers({
            redundancy: 0,
            tor:        12000,
            premiumSLA: 15000,
            spot:       5000
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroMultiplier.selector);
        pricing.setMultipliers(newMult);
    }

    function test_SetMultipliers_ZeroTorReverts() public {
        BunkerPricing.Multipliers memory newMult = BunkerPricing.Multipliers({
            redundancy: 30000,
            tor:        0,
            premiumSLA: 15000,
            spot:       5000
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroMultiplier.selector);
        pricing.setMultipliers(newMult);
    }

    function test_SetMultipliers_ZeroPremiumSLAReverts() public {
        BunkerPricing.Multipliers memory newMult = BunkerPricing.Multipliers({
            redundancy: 30000,
            tor:        12000,
            premiumSLA: 0,
            spot:       5000
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroMultiplier.selector);
        pricing.setMultipliers(newMult);
    }

    function test_SetMultipliers_ZeroSpotReverts() public {
        BunkerPricing.Multipliers memory newMult = BunkerPricing.Multipliers({
            redundancy: 30000,
            tor:        12000,
            premiumSLA: 15000,
            spot:       0
        });

        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ZeroMultiplier.selector);
        pricing.setMultipliers(newMult);
    }

    function test_SetMultipliers_NonOwnerReverts() public {
        BunkerPricing.Multipliers memory newMult = BunkerPricing.Multipliers({
            redundancy: 20000,
            tor:        15000,
            premiumSLA: 20000,
            spot:       3000
        });

        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        pricing.setMultipliers(newMult);
    }

    /// @notice Verify cost changes after multiplier update.
    function test_Cost_ReflectsUpdatedMultipliers() public {
        // Change redundancy from 3x to 2x.
        BunkerPricing.Multipliers memory newMult = BunkerPricing.Multipliers({
            redundancy: 20000,  // 2x
            tor:        12000,
            premiumSLA: 15000,
            spot:       5000
        });
        vm.prank(owner);
        pricing.setMultipliers(newMult);

        // 1 CPU for 1 hour: cpuCost = 0.5e18, after 2x redundancy = 1e18
        BunkerPricing.ResourceRequest memory req = _req(1000, 0, 0, 0, 1);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 1e18);
    }

    // =======================================================================
    //  7. Price Oracle
    // =======================================================================

    function test_SetPriceOracle_Works() public {
        address oracle = makeAddr("oracle");

        vm.prank(owner);
        pricing.setPriceOracle(oracle);

        assertEq(address(pricing.priceOracle()), oracle);
    }

    function test_SetPriceOracle_EmitsPriceOracleUpdated() public {
        address oracle = makeAddr("oracle");

        vm.prank(owner);
        vm.expectEmit(true, true, false, false);
        emit PriceOracleUpdated(address(0), oracle);
        pricing.setPriceOracle(oracle);
    }

    function test_SetPriceOracle_CanSetToZeroToDisable() public {
        address oracle = makeAddr("oracle");

        vm.startPrank(owner);
        pricing.setPriceOracle(oracle);
        assertEq(address(pricing.priceOracle()), oracle);

        pricing.setPriceOracle(address(0));
        assertEq(address(pricing.priceOracle()), address(0));
        vm.stopPrank();
    }

    function test_SetPriceOracle_EmitsCorrectOldAddress() public {
        address oracle1 = makeAddr("oracle1");
        address oracle2 = makeAddr("oracle2");

        vm.startPrank(owner);
        pricing.setPriceOracle(oracle1);

        vm.expectEmit(true, true, false, false);
        emit PriceOracleUpdated(oracle1, oracle2);
        pricing.setPriceOracle(oracle2);
        vm.stopPrank();
    }

    function test_SetPriceOracle_NonOwnerReverts() public {
        vm.prank(alice);
        vm.expectRevert(
            abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", alice)
        );
        pricing.setPriceOracle(makeAddr("oracle"));
    }

    // =======================================================================
    //  8. Edge Cases
    // =======================================================================

    /// @notice Very large resource values should not overflow.
    ///   Using type(uint128).max-scale values to stay within uint256 bounds.
    function test_Cost_LargeNumbersNoOverflow() public view {
        // Use moderate large values that won't overflow when multiplied together.
        // cpuCores=1e9 milli (1M cores), memoryGB=1e9 milli (1M GB), 8760 hours (1 year).
        BunkerPricing.ResourceRequest memory req = _req(1e9, 1e9, 1e9, 1e9, 8760);
        // Should not revert due to overflow.
        uint256 cost = pricing.calculateCost(req);
        assertTrue(cost > 0, "large cost should be nonzero");
    }

    /// @notice Zero duration means zero compute cost, but network cost remains.
    /// Math:
    ///   cpuCost    = (1000 * 0.5e18 * 0) / 1000 = 0
    ///   memoryCost = 0
    ///   storageCost = 0 (guarded by durationHours > 0 check)
    ///   networkCost = 50 * 0.02e18 = 1e18
    ///   subtotal = 1e18
    ///   After redundancy (3x): 3e18
    function test_Cost_ZeroDuration_NetworkStillCounts() public view {
        BunkerPricing.ResourceRequest memory req = _req(1000, 4000, 20, 50, 0);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 3e18);
    }

    /// @notice All-zero request yields zero cost.
    function test_Cost_AllZeroRequest() public view {
        BunkerPricing.ResourceRequest memory req = _req(0, 0, 0, 0, 0);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 0);
    }

    /// @notice All-zero request with all flags true still yields zero cost.
    ///   0 * any multiplier = 0.
    function test_Cost_AllZeroWithAllFlags() public view {
        BunkerPricing.ResourceRequest memory req = _reqFull(0, 0, 0, 0, 0, true, false, true, true, true);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 0);
    }

    /// @notice Fractional CPU (0.25 cores = 250 milli) for 1 hour.
    /// Math:
    ///   cpuCost = (250 * 0.5e18 * 1) / 1000 = 0.125e18
    ///   After redundancy (3x): 0.375e18
    function test_Cost_FractionalCpu() public view {
        BunkerPricing.ResourceRequest memory req = _req(250, 0, 0, 0, 1);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 0.375e18);
    }

    /// @notice Storage with zero GB but nonzero duration = zero storage cost.
    function test_Cost_ZeroStorageGB_NoStorageCost() public view {
        BunkerPricing.ResourceRequest memory req = _req(0, 0, 0, 0, 720);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 0);
    }

    /// @notice Single wei of CPU price still works correctly.
    function test_Cost_MinimalPrice() public {
        vm.prank(owner);
        pricing.setCPUPrice(1); // 1 wei per core-hour

        // 1000 milli-cores for 1 hour: (1000 * 1 * 1) / 1000 = 1 wei
        // After redundancy: 1 * 30000 / 10000 = 3 wei
        BunkerPricing.ResourceRequest memory req = _req(1000, 0, 0, 0, 1);
        uint256 cost = pricing.calculateCost(req);
        assertEq(cost, 3);
    }

    /// @notice Storage division truncation: small hours produce zero storage cost
    ///   due to integer division by 730.
    ///   1 GB storage for 1 hour: (1 * 0.05e18 * 1) / 730 = 68493150684931 (truncated)
    function test_Cost_StorageDivisionTruncation() public view {
        BunkerPricing.ResourceRequest memory req = _req(0, 0, 1, 0, 1);
        uint256 cost = pricing.calculateCost(req);

        uint256 storageCost = _testCeilDiv(1 * STORAGE_PRICE * 1, 730);
        uint256 expected = _testCeilDiv(storageCost * REDUNDANCY_BPS, BPS);

        assertEq(cost, expected);
        // Verify it is nonzero but small.
        assertTrue(cost > 0, "tiny storage cost should be nonzero");
        assertTrue(cost < 1e18, "tiny storage cost should be much less than 1 BUNKER");
    }

    // =======================================================================
    //  9. Fuzz Tests
    // =======================================================================

    /// @notice Fuzz cost calculation with varying resource inputs.
    ///   Verifies that calculateCost matches a manual Solidity calculation.
    function testFuzz_calculateCost(uint48 cpuCores, uint48 memoryGB, uint48 duration) public view {
        // Constrain inputs to avoid overflow while still being meaningful.
        // Using uint48 max (~281 trillion) scaled down to reasonable ranges.
        vm.assume(cpuCores <= 100_000);    // up to 100 cores in milli-units
        vm.assume(memoryGB <= 1_000_000);  // up to 1000 GB in milli-units
        vm.assume(duration <= 87_600);     // up to 10 years in hours

        BunkerPricing.ResourceRequest memory req = _req(
            uint256(cpuCores),
            uint256(memoryGB),
            0,   // no storage for simplicity
            0,   // no network
            uint256(duration)
        );

        uint256 cost = pricing.calculateCost(req);

        // Recompute expected cost manually
        uint256 cpuCost = (uint256(cpuCores) * CPU_PRICE * uint256(duration)) / 1000;
        uint256 memoryCost = (uint256(memoryGB) * MEMORY_PRICE * uint256(duration)) / 1000;
        uint256 subtotal = cpuCost + memoryCost;
        uint256 expected = (subtotal * REDUNDANCY_BPS) / BPS;

        assertEq(cost, expected, "fuzz cost mismatch");
    }

    /// @notice Fuzz price updates: set arbitrary prices and verify they persist.
    function testFuzz_setPrices(uint128 cpu, uint128 memory_, uint128 storage_, uint128 network) public {
        // All resource prices must be non-zero per contract validation
        vm.assume(cpu > 0);
        vm.assume(memory_ > 0);
        vm.assume(storage_ > 0);
        vm.assume(network > 0);

        BunkerPricing.ResourcePrices memory newPrices = BunkerPricing.ResourcePrices({
            cpuPerCoreHour:    uint256(cpu),
            memoryPerGBHour:   uint256(memory_),
            storagePerGBMonth: uint256(storage_),
            networkPerGB:      uint256(network),
            gpuBasicPerHour:   GPU_BASIC,
            gpuPremiumPerHour: GPU_PREMIUM
        });

        vm.prank(owner);
        pricing.setPrices(newPrices);

        BunkerPricing.ResourcePrices memory p = pricing.getPrices();
        assertEq(p.cpuPerCoreHour,    uint256(cpu),      "cpu price mismatch");
        assertEq(p.memoryPerGBHour,   uint256(memory_),  "memory price mismatch");
        assertEq(p.storagePerGBMonth, uint256(storage_),  "storage price mismatch");
        assertEq(p.networkPerGB,      uint256(network),   "network price mismatch");
    }

    /// @notice Fuzz multiplier stacking: apply combinations of multipliers and
    ///   verify the cost matches manual calculation.
    function testFuzz_multiplierStacking(uint16 redundancy, uint16 tor, uint16 sla) public {
        // All multipliers must be non-zero (contract rejects zero multipliers)
        vm.assume(redundancy > 0 && redundancy <= 50000);  // up to 5x
        vm.assume(tor > 0 && tor <= 50000);
        vm.assume(sla > 0 && sla <= 50000);

        BunkerPricing.Multipliers memory newMult = BunkerPricing.Multipliers({
            redundancy: uint256(redundancy),
            tor:        uint256(tor),
            premiumSLA: uint256(sla),
            spot:       SPOT_BPS  // keep spot constant
        });

        vm.prank(owner);
        pricing.setMultipliers(newMult);

        // Calculate cost for 1 CPU core for 1 hour with tor + premiumSLA
        BunkerPricing.ResourceRequest memory req = _reqFull(
            1000, 0, 0, 0, 1,     // 1 core, 1 hour
            false, false,          // no GPU
            true, true, false      // tor=true, premiumSLA=true, spot=false
        );

        uint256 cost = pricing.calculateCost(req);

        // Manual calculation
        // cpuCost = (1000 * CPU_PRICE * 1) / 1000 = CPU_PRICE = 0.5e18
        uint256 baseCost = CPU_PRICE;
        uint256 afterRedundancy = (baseCost * uint256(redundancy)) / BPS;
        uint256 afterTor = (afterRedundancy * uint256(tor)) / BPS;
        uint256 expected = (afterTor * uint256(sla)) / BPS;

        assertEq(cost, expected, "multiplier stacking mismatch");
    }

    /// @dev Ceiling division helper matching BunkerPricing._ceilDiv
    function _testCeilDiv(uint256 a, uint256 b) internal pure returns (uint256) {
        return a == 0 ? 0 : (a - 1) / b + 1;
    }

    // ──────────────────────────────────────────────
    //  setMaxMultiplierBps Tests
    // ──────────────────────────────────────────────

    function test_setMaxMultiplierBps_succeeds() public {
        vm.prank(owner);
        pricing.setMaxMultiplierBps(200000);
        assertEq(pricing.maxMultiplierBps(), 200000);
    }

    function test_setMaxMultiplierBps_revertsBelowMin() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.InvalidMultiplierCap.selector);
        pricing.setMaxMultiplierBps(9999);
    }

    function test_setMaxMultiplierBps_revertsAboveMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.InvalidMultiplierCap.selector);
        pricing.setMaxMultiplierBps(500001);
    }

    function test_setMaxMultiplierBps_revertsNonOwner() public {
        vm.prank(alice);
        vm.expectRevert();
        pricing.setMaxMultiplierBps(200000);
    }

    function test_setMaxMultiplierBps_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerPricing.MaxMultiplierUpdated(200000);
        pricing.setMaxMultiplierBps(200000);
    }

    // ──────────────────────────────────────────────
    //  setStaleThresholdBounds Tests
    // ──────────────────────────────────────────────

    function test_setStaleThresholdBounds_succeeds() public {
        vm.prank(owner);
        pricing.setStaleThresholdBounds(10 minutes, 12 hours);
        assertEq(pricing.minStaleThreshold(), 10 minutes);
        assertEq(pricing.maxStaleThreshold(), 12 hours);
    }

    function test_setStaleThresholdBounds_revertsZeroMin() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.InvalidThresholdBounds.selector);
        pricing.setStaleThresholdBounds(0, 12 hours);
    }

    function test_setStaleThresholdBounds_revertsMinGteMax() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.InvalidThresholdBounds.selector);
        pricing.setStaleThresholdBounds(12 hours, 12 hours);
    }

    function test_setStaleThresholdBounds_revertsMaxTooLong() public {
        vm.prank(owner);
        vm.expectRevert(BunkerPricing.ThresholdTooLong.selector);
        pricing.setStaleThresholdBounds(5 minutes, 8 days);
    }

    function test_setStaleThresholdBounds_revertsNonOwner() public {
        vm.prank(alice);
        vm.expectRevert();
        pricing.setStaleThresholdBounds(10 minutes, 12 hours);
    }

    function test_setStaleThresholdBounds_clampsCurrentThreshold() public {
        // stalePriceThreshold defaults to 1 hour
        vm.startPrank(owner);

        // Set bounds that make current threshold too high
        pricing.setStaleThresholdBounds(1 minutes, 30 minutes);
        assertEq(pricing.stalePriceThreshold(), 30 minutes);

        // Set bounds that make current threshold too low
        pricing.setStaleThresholdBounds(2 hours, 6 hours);
        assertEq(pricing.stalePriceThreshold(), 2 hours);

        vm.stopPrank();
    }

    function test_setStaleThresholdBounds_emitsEvent() public {
        vm.prank(owner);
        vm.expectEmit(false, false, false, true);
        emit BunkerPricing.StaleThresholdBoundsUpdated(10 minutes, 12 hours);
        pricing.setStaleThresholdBounds(10 minutes, 12 hours);
    }
}
