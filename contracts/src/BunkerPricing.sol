// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable2Step.sol";

/// @dev Chainlink price feed interface for BUNKER/USD conversion.
interface AggregatorV3Interface {
    function latestRoundData() external view returns (
        uint80 roundId,
        int256 answer,
        uint256 startedAt,
        uint256 updatedAt,
        uint80 answeredInRound
    );
    function decimals() external view returns (uint8);
}

/// @title BunkerPricing
/// @author Moltbunker
/// @notice Stores and manages configurable resource pricing parameters for
///         the Moltbunker decentralized compute network. Supports oracle-based
///         USD pricing and provider-set custom rates.
/// @dev No token transfers. Read-heavy contract. Admin updates prices.
///      All prices in BUNKER wei (18 decimals). Multipliers in basis points
///      (10000 = 1.00x). Oracle prices use 8 decimals (Chainlink standard).
contract BunkerPricing is Ownable2Step {

    // ──────────────────────────────────────────────
    //  Constants
    // ──────────────────────────────────────────────

    /// @notice Contract version.
    string public constant VERSION = "1.2.0";

    /// @notice Basis points denominator for multipliers.
    uint256 public constant BPS_DENOMINATOR = 10000;

    /// @notice Minimum allowed stale price threshold (default 5 minutes).
    uint256 public minStaleThreshold = 5 minutes;

    /// @notice Maximum allowed stale price threshold (default 24 hours).
    uint256 public maxStaleThreshold = 24 hours;

    /// @notice Maximum multiplier in basis points (default 10x = 100000).
    uint256 public maxMultiplierBps = 100000;

    // ──────────────────────────────────────────────
    //  Types
    // ──────────────────────────────────────────────

    /// @notice Resource pricing configuration.
    struct ResourcePrices {
        uint256 cpuPerCoreHour;      // BUNKER wei per CPU core-hour
        uint256 memoryPerGBHour;     // BUNKER wei per GB-hour
        uint256 storagePerGBMonth;   // BUNKER wei per GB-month
        uint256 networkPerGB;        // BUNKER wei per GB egress
        uint256 gpuBasicPerHour;     // BUNKER wei per basic GPU-hour
        uint256 gpuPremiumPerHour;   // BUNKER wei per premium GPU-hour
    }

    /// @notice Multiplier configuration (basis points, 10000 = 1.00x).
    struct Multipliers {
        uint256 redundancy;   // 3x = 30000
        uint256 tor;          // 1.2x = 12000
        uint256 premiumSLA;   // 1.5x = 15000
        uint256 spot;         // 0.5x = 5000
    }

    /// @notice Resource requirements for cost calculation.
    struct ResourceRequest {
        uint256 cpuCores;     // Number of CPU cores (whole or fractional * 1000)
        uint256 memoryGB;     // GB of memory (scaled by 1000 for milli-precision)
        uint256 storageGB;    // GB of storage
        uint256 networkGB;    // GB of expected egress
        uint256 durationHours; // Duration in hours
        bool useGPUBasic;     // Request basic GPU
        bool useGPUPremium;   // Request premium GPU
        bool useTor;          // Tor-only deployment
        bool usePremiumSLA;   // 99.99% SLA
        bool useSpot;         // Preemptible instance
    }

    /// @notice Provider-specific pricing configuration.
    struct ProviderPricing {
        uint256 cpuPerCoreHour;
        uint256 memoryPerGBHour;
        uint256 storagePerGBMonth;
        uint256 networkPerGB;
        bool isSet; // Whether provider has custom pricing
    }

    // ──────────────────────────────────────────────
    //  State
    // ──────────────────────────────────────────────

    /// @notice Current resource prices.
    ResourcePrices public prices;

    /// @notice Current multipliers.
    Multipliers public multipliers;

    /// @notice Chainlink price feed for BUNKER/USD conversion.
    AggregatorV3Interface public priceOracle;

    /// @notice Maximum age of oracle price data before it is considered stale.
    uint256 public stalePriceThreshold = 1 hours;

    /// @notice Toggle for oracle-based USD pricing mode.
    bool public useOracleForPricing;

    /// @notice Provider-specific pricing overrides.
    mapping(address => ProviderPricing) public providerPrices;

    /// @notice Minimum price multiplier in basis points (floor = 0.5x of default).
    uint256 public minPriceMultiplierBps = 5000;

    /// @notice Maximum price multiplier in basis points (ceiling = 3.0x of default).
    uint256 public maxPriceMultiplierBps = 30000;

    // ──────────────────────────────────────────────
    //  Errors
    // ──────────────────────────────────────────────

    error ZeroPrice();
    error ZeroMultiplier();
    error InvalidMultiplier(uint256 multiplier);
    error StalePriceData(uint256 updatedAt, uint256 threshold);
    error NegativePrice();
    error OracleNotSet();
    error PriceBelowMinimum(uint256 price, uint256 minimum);
    error PriceAboveMaximum(uint256 price, uint256 maximum);
    error InvalidPriceBounds(uint256 minBps, uint256 maxBps);
    error InvalidStalePriceThreshold(uint256 threshold);
    error MultiplierTooHigh();
    error InvalidMultiplierCap();
    error InvalidThresholdBounds();
    error ThresholdTooLong();

    // ──────────────────────────────────────────────
    //  Events
    // ──────────────────────────────────────────────

    /// @notice Emitted when any resource price is updated.
    event PricesUpdated(
        uint256 cpuPerCoreHour,
        uint256 memoryPerGBHour,
        uint256 storagePerGBMonth,
        uint256 networkPerGB,
        uint256 gpuBasicPerHour,
        uint256 gpuPremiumPerHour
    );

    /// @notice Emitted when multipliers are updated.
    event MultipliersUpdated(
        uint256 redundancy,
        uint256 tor,
        uint256 premiumSLA,
        uint256 spot
    );

    /// @notice Emitted when the price oracle address is set.
    event PriceOracleUpdated(address indexed oldOracle, address indexed newOracle);

    /// @notice Emitted when oracle pricing mode is toggled.
    event OraclePricingEnabled(bool enabled);

    /// @notice Emitted when the stale price threshold is updated.
    event StalePriceThresholdUpdated(uint256 oldThreshold, uint256 newThreshold);

    /// @notice Emitted when a provider sets custom pricing.
    event ProviderPricesSet(
        address indexed provider,
        uint256 cpu,
        uint256 memory_,
        uint256 storage_,
        uint256 network
    );

    /// @notice Emitted when provider price bounds are updated.
    event ProviderPriceBoundsUpdated(uint256 minBps, uint256 maxBps);

    /// @notice Emitted when a provider clears their custom pricing.
    event ProviderPricesCleared(address indexed provider);

    /// @notice Emitted when the max multiplier cap is updated.
    event MaxMultiplierUpdated(uint256 newMax);

    /// @notice Emitted when the stale threshold bounds are updated.
    event StaleThresholdBoundsUpdated(uint256 newMin, uint256 newMax);

    // ──────────────────────────────────────────────
    //  Constructor
    // ──────────────────────────────────────────────

    /// @param _initialOwner Admin wallet address.
    constructor(address _initialOwner) Ownable(_initialOwner) {
        // Set default prices (matching Go DefaultPricingConfig).
        prices = ResourcePrices({
            cpuPerCoreHour:    500000000000000000,    // 0.50 BUNKER
            memoryPerGBHour:   100000000000000000,    // 0.10 BUNKER
            storagePerGBMonth: 50000000000000000,     // 0.05 BUNKER
            networkPerGB:      20000000000000000,     // 0.02 BUNKER
            gpuBasicPerHour:   5000000000000000000,   // 5.00 BUNKER
            gpuPremiumPerHour: 15000000000000000000   // 15.00 BUNKER
        });

        // Set default multipliers (in basis points).
        multipliers = Multipliers({
            redundancy: 30000,  // 3.00x
            tor:        12000,  // 1.20x
            premiumSLA: 15000,  // 1.50x
            spot:        5000   // 0.50x
        });
    }

    // ──────────────────────────────────────────────
    //  External: Cost Calculation (View)
    // ──────────────────────────────────────────────

    /// @notice Calculate the total cost for a resource request.
    /// @dev All inputs use whole units except cpuCores and memoryGB which use
    ///      milli-precision (multiply actual value by 1000). durationHours is in whole hours.
    ///      Redundancy multiplier is always applied. Additional multipliers are
    ///      applied based on the request flags. If oracle pricing is enabled, the
    ///      base cost (calculated in BUNKER wei) is treated as USD-denominated and
    ///      converted to BUNKER tokens using the oracle price.
    /// @param req The resource request specification.
    /// @return totalCost Total cost in BUNKER wei.
    function calculateCost(ResourceRequest calldata req)
        external
        view
        returns (uint256 totalCost)
    {
        totalCost = _calculateBaseCost(req, prices);

        // If oracle pricing is enabled, the base prices represent USD values
        // (8 decimals like Chainlink). Convert the USD cost to BUNKER tokens.
        if (useOracleForPricing) {
            totalCost = _convertUsdToBunker(totalCost);
        }
    }

    /// @notice Calculate cost in USD (8 decimals matching Chainlink).
    /// @dev Uses the same base prices and multipliers but returns the result
    ///      as a USD amount without oracle conversion.
    /// @param req The resource request specification.
    /// @return usdCost Cost in USD with 8 decimals.
    function calculateCostUSD(ResourceRequest calldata req)
        external
        view
        returns (uint256 usdCost)
    {
        usdCost = _calculateBaseCost(req, prices);
    }

    /// @notice Calculate cost using a specific provider's custom pricing.
    /// @dev Falls back to default prices if the provider has not set custom rates.
    ///      GPU prices always use the default rates (providers cannot override GPU).
    ///      Provider prices are clamped to current bounds at read time to handle
    ///      cases where admin changes defaults after provider set their prices.
    /// @param provider The provider address.
    /// @param req The resource request specification.
    /// @return totalCost Total cost in BUNKER wei.
    function calculateProviderCost(address provider, ResourceRequest calldata req)
        external
        view
        returns (uint256 totalCost)
    {
        ProviderPricing storage pp = providerPrices[provider];
        if (!pp.isSet) {
            // No custom pricing, use default.
            totalCost = _calculateBaseCost(req, prices);
        } else {
            // Build a ResourcePrices struct from provider overrides + default GPU prices.
            // Clamp provider prices to current bounds in case defaults changed.
            ResourcePrices memory providerResourcePrices = ResourcePrices({
                cpuPerCoreHour: _clampToProviderBounds(pp.cpuPerCoreHour, prices.cpuPerCoreHour),
                memoryPerGBHour: _clampToProviderBounds(pp.memoryPerGBHour, prices.memoryPerGBHour),
                storagePerGBMonth: _clampToProviderBounds(pp.storagePerGBMonth, prices.storagePerGBMonth),
                networkPerGB: _clampToProviderBounds(pp.networkPerGB, prices.networkPerGB),
                gpuBasicPerHour: prices.gpuBasicPerHour,
                gpuPremiumPerHour: prices.gpuPremiumPerHour
            });
            totalCost = _calculateBaseCost(req, providerResourcePrices);
        }

        if (useOracleForPricing) {
            totalCost = _convertUsdToBunker(totalCost);
        }
    }

    /// @notice Get the current BUNKER/USD price from the Chainlink oracle.
    /// @dev Validates roundId, answeredInRound, answer sign, and staleness.
    /// @return price The token price (scaled by oracle decimals).
    /// @return decimals_ The number of decimals in the price.
    function getTokenPrice() public view returns (uint256 price, uint8 decimals_) {
        if (address(priceOracle) == address(0)) revert OracleNotSet();
        (uint80 roundId, int256 answer,, uint256 updatedAt, uint80 answeredInRound) =
            priceOracle.latestRoundData();
        if (answer <= 0) revert NegativePrice();
        if (answeredInRound < roundId) revert StalePriceData(updatedAt, stalePriceThreshold);
        if (block.timestamp - updatedAt > stalePriceThreshold)
            revert StalePriceData(updatedAt, stalePriceThreshold);
        return (uint256(answer), priceOracle.decimals());
    }

    /// @notice Get all current prices.
    /// @return p The current resource prices.
    function getPrices() external view returns (ResourcePrices memory p) {
        return prices;
    }

    /// @notice Get all current multipliers.
    /// @return m The current multipliers.
    function getMultipliers() external view returns (Multipliers memory m) {
        return multipliers;
    }

    /// @notice Get a provider's custom pricing configuration.
    /// @param provider The provider address.
    /// @return pp The provider's pricing (isSet=false if using defaults).
    function getProviderPrices(address provider)
        external
        view
        returns (ProviderPricing memory pp)
    {
        return providerPrices[provider];
    }

    // ──────────────────────────────────────────────
    //  External: Provider Pricing
    // ──────────────────────────────────────────────

    /// @notice Set custom pricing for the calling provider.
    /// @dev Each price must be within [minPriceMultiplierBps, maxPriceMultiplierBps]
    ///      of the corresponding default price (enforced per resource type).
    /// @param pp The provider pricing configuration (isSet field is ignored).
    function setProviderPrices(ProviderPricing calldata pp) external {
        // Validate each price against bounds relative to default prices.
        _validateProviderPrice(pp.cpuPerCoreHour, prices.cpuPerCoreHour, "cpu");
        _validateProviderPrice(pp.memoryPerGBHour, prices.memoryPerGBHour, "memory");
        _validateProviderPrice(pp.storagePerGBMonth, prices.storagePerGBMonth, "storage");
        _validateProviderPrice(pp.networkPerGB, prices.networkPerGB, "network");

        providerPrices[msg.sender] = ProviderPricing({
            cpuPerCoreHour: pp.cpuPerCoreHour,
            memoryPerGBHour: pp.memoryPerGBHour,
            storagePerGBMonth: pp.storagePerGBMonth,
            networkPerGB: pp.networkPerGB,
            isSet: true
        });

        emit ProviderPricesSet(
            msg.sender,
            pp.cpuPerCoreHour,
            pp.memoryPerGBHour,
            pp.storagePerGBMonth,
            pp.networkPerGB
        );
    }

    /// @notice Clear custom pricing for the calling provider, reverting to defaults.
    function clearProviderPrices() external {
        delete providerPrices[msg.sender];
        emit ProviderPricesCleared(msg.sender);
    }

    // ──────────────────────────────────────────────
    //  External: Admin Price Updates
    // ──────────────────────────────────────────────

    /// @notice Update all resource prices at once.
    /// @param newPrices New resource pricing configuration.
    function setPrices(ResourcePrices calldata newPrices) external onlyOwner {
        if (newPrices.cpuPerCoreHour == 0 || newPrices.memoryPerGBHour == 0 ||
            newPrices.storagePerGBMonth == 0 || newPrices.networkPerGB == 0 ||
            newPrices.gpuBasicPerHour == 0 || newPrices.gpuPremiumPerHour == 0)
            revert ZeroPrice();

        prices = newPrices;

        emit PricesUpdated(
            newPrices.cpuPerCoreHour,
            newPrices.memoryPerGBHour,
            newPrices.storagePerGBMonth,
            newPrices.networkPerGB,
            newPrices.gpuBasicPerHour,
            newPrices.gpuPremiumPerHour
        );
    }

    /// @notice Update individual CPU price.
    /// @param newPrice New CPU price per core-hour in BUNKER wei.
    function setCPUPrice(uint256 newPrice) external onlyOwner {
        if (newPrice == 0) revert ZeroPrice();
        prices.cpuPerCoreHour = newPrice;
        _emitPricesUpdated();
    }

    /// @notice Update individual memory price.
    /// @param newPrice New memory price per GB-hour in BUNKER wei.
    function setMemoryPrice(uint256 newPrice) external onlyOwner {
        if (newPrice == 0) revert ZeroPrice();
        prices.memoryPerGBHour = newPrice;
        _emitPricesUpdated();
    }

    /// @notice Update individual storage price.
    /// @param newPrice New storage price per GB-month in BUNKER wei.
    function setStoragePrice(uint256 newPrice) external onlyOwner {
        if (newPrice == 0) revert ZeroPrice();
        prices.storagePerGBMonth = newPrice;
        _emitPricesUpdated();
    }

    /// @notice Update individual network price.
    /// @param newPrice New network price per GB in BUNKER wei.
    function setNetworkPrice(uint256 newPrice) external onlyOwner {
        if (newPrice == 0) revert ZeroPrice();
        prices.networkPerGB = newPrice;
        _emitPricesUpdated();
    }

    /// @notice Update GPU prices.
    /// @param basicPerHour New basic GPU price per hour in BUNKER wei.
    /// @param premiumPerHour New premium GPU price per hour in BUNKER wei.
    function setGPUPrices(uint256 basicPerHour, uint256 premiumPerHour) external onlyOwner {
        if (basicPerHour == 0 || premiumPerHour == 0) revert ZeroPrice();
        prices.gpuBasicPerHour = basicPerHour;
        prices.gpuPremiumPerHour = premiumPerHour;
        _emitPricesUpdated();
    }

    // ──────────────────────────────────────────────
    //  External: Admin Multiplier Updates
    // ──────────────────────────────────────────────

    /// @notice Update all multipliers at once.
    /// @param newMultipliers New multiplier configuration (basis points).
    ///        Each multiplier must be > 0 and <= maxMultiplierBps (10x).
    function setMultipliers(Multipliers calldata newMultipliers) external onlyOwner {
        if (newMultipliers.redundancy == 0 || newMultipliers.tor == 0 ||
            newMultipliers.premiumSLA == 0 || newMultipliers.spot == 0)
            revert ZeroMultiplier();
        if (newMultipliers.redundancy > maxMultiplierBps ||
            newMultipliers.tor > maxMultiplierBps ||
            newMultipliers.premiumSLA > maxMultiplierBps ||
            newMultipliers.spot > maxMultiplierBps)
            revert MultiplierTooHigh();

        multipliers = newMultipliers;

        emit MultipliersUpdated(
            newMultipliers.redundancy,
            newMultipliers.tor,
            newMultipliers.premiumSLA,
            newMultipliers.spot
        );
    }

    // ──────────────────────────────────────────────
    //  External: Admin Oracle
    // ──────────────────────────────────────────────

    /// @notice Set the price oracle address for USD-denominated pricing.
    /// @dev If setting to address(0) while oracle pricing is enabled, oracle pricing
    ///      is automatically disabled to prevent reverts in cost calculations.
    /// @param newOracle The oracle contract address (or address(0) to disable).
    function setPriceOracle(address newOracle) external onlyOwner {
        address old = address(priceOracle);
        priceOracle = AggregatorV3Interface(newOracle);
        if (newOracle == address(0) && useOracleForPricing) {
            useOracleForPricing = false;
            emit OraclePricingEnabled(false);
        }
        emit PriceOracleUpdated(old, newOracle);
    }

    /// @notice Enable or disable oracle-based pricing.
    /// @dev Reverts if enabling when no oracle is set.
    /// @param enabled True to use oracle pricing, false for direct BUNKER pricing.
    function enableOraclePricing(bool enabled) external onlyOwner {
        if (enabled && address(priceOracle) == address(0)) revert OracleNotSet();
        useOracleForPricing = enabled;
        emit OraclePricingEnabled(enabled);
    }

    /// @notice Update the stale price threshold for oracle data.
    /// @param newThreshold New maximum age in seconds for oracle price data.
    ///        Must be between MIN_STALE_THRESHOLD (5 minutes) and MAX_STALE_THRESHOLD (24 hours).
    function setStalePriceThreshold(uint256 newThreshold) external onlyOwner {
        if (newThreshold < minStaleThreshold || newThreshold > maxStaleThreshold)
            revert InvalidStalePriceThreshold(newThreshold);
        uint256 old = stalePriceThreshold;
        stalePriceThreshold = newThreshold;
        emit StalePriceThresholdUpdated(old, newThreshold);
    }

    // ──────────────────────────────────────────────
    //  External: Admin Provider Price Bounds
    // ──────────────────────────────────────────────

    /// @notice Set the minimum and maximum bounds for provider-set pricing.
    /// @param minBps Minimum multiplier in basis points (e.g., 5000 = 0.5x).
    /// @param maxBps Maximum multiplier in basis points (e.g., 30000 = 3.0x).
    function setProviderPriceBounds(uint256 minBps, uint256 maxBps) external onlyOwner {
        if (minBps == 0 || maxBps == 0 || minBps >= maxBps)
            revert InvalidPriceBounds(minBps, maxBps);
        minPriceMultiplierBps = minBps;
        maxPriceMultiplierBps = maxBps;
        emit ProviderPriceBoundsUpdated(minBps, maxBps);
    }

    /// @notice Adjust the maximum multiplier cap for pricing multipliers.
    /// @param newMax New cap in basis points (BPS_DENOMINATOR = 1x min, 500000 = 50x max).
    function setMaxMultiplierBps(uint256 newMax) external onlyOwner {
        if (newMax < BPS_DENOMINATOR || newMax > 500000) revert InvalidMultiplierCap();
        maxMultiplierBps = newMax;
        emit MaxMultiplierUpdated(newMax);
    }

    /// @notice Adjust the stale price threshold bounds for oracle data.
    /// @param newMin New minimum stale threshold in seconds.
    /// @param newMax New maximum stale threshold in seconds.
    function setStaleThresholdBounds(uint256 newMin, uint256 newMax) external onlyOwner {
        if (newMin == 0 || newMin >= newMax) revert InvalidThresholdBounds();
        if (newMax > 7 days) revert ThresholdTooLong();
        minStaleThreshold = newMin;
        maxStaleThreshold = newMax;
        // Clamp current threshold if out of new bounds.
        if (stalePriceThreshold < newMin) stalePriceThreshold = newMin;
        if (stalePriceThreshold > newMax) stalePriceThreshold = newMax;
        emit StaleThresholdBoundsUpdated(newMin, newMax);
    }

    // ──────────────────────────────────────────────
    //  Internal
    // ──────────────────────────────────────────────

    /// @dev Calculate the base cost using the given price set and apply multipliers.
    ///      When oracle pricing is enabled, these values represent USD-denominated costs
    ///      (same scale as BUNKER wei but representing USD). Admin must set appropriate
    ///      prices before enabling oracle mode via `enableOraclePricing`.
    ///      Uses ceiling division to prevent truncation from undercharging.
    function _calculateBaseCost(ResourceRequest calldata req, ResourcePrices memory p)
        internal
        view
        returns (uint256 totalCost)
    {
        // Base compute cost (cpuCores and memoryGB are in milli-units).
        uint256 cpuCost = _ceilDiv(req.cpuCores * p.cpuPerCoreHour * req.durationHours, 1000);
        uint256 memoryCost = _ceilDiv(req.memoryGB * p.memoryPerGBHour * req.durationHours, 1000);

        // Storage cost (convert hours to months: hours / 730).
        uint256 storageCost = 0;
        if (req.storageGB > 0 && req.durationHours > 0) {
            storageCost = _ceilDiv(req.storageGB * p.storagePerGBMonth * req.durationHours, 730);
        }

        // Network cost.
        uint256 networkCost = req.networkGB * p.networkPerGB;

        // GPU cost.
        uint256 gpuCost = 0;
        if (req.useGPUBasic) {
            gpuCost = p.gpuBasicPerHour * req.durationHours;
        } else if (req.useGPUPremium) {
            gpuCost = p.gpuPremiumPerHour * req.durationHours;
        }

        totalCost = cpuCost + memoryCost + storageCost + networkCost + gpuCost;

        // Apply redundancy multiplier (mandatory).
        totalCost = _ceilDiv(totalCost * multipliers.redundancy, BPS_DENOMINATOR);

        // Apply optional multipliers.
        if (req.useTor) {
            totalCost = _ceilDiv(totalCost * multipliers.tor, BPS_DENOMINATOR);
        }
        if (req.usePremiumSLA) {
            totalCost = _ceilDiv(totalCost * multipliers.premiumSLA, BPS_DENOMINATOR);
        }
        if (req.useSpot) {
            totalCost = _ceilDiv(totalCost * multipliers.spot, BPS_DENOMINATOR);
        }
    }

    /// @dev Convert a USD-denominated cost to BUNKER tokens using the oracle.
    ///      Uses ceiling division to prevent truncation from undercharging.
    /// @param usdCost The cost in USD (same decimals as the base prices).
    /// @return bunkerCost The cost in BUNKER wei (18 decimals).
    function _convertUsdToBunker(uint256 usdCost) internal view returns (uint256 bunkerCost) {
        (uint256 tokenPrice, uint8 oracleDecimals) = getTokenPrice();
        // tokenPrice = BUNKER price in USD (e.g., 8 decimals).
        // bunkerCost = usdCost * 10^oracleDecimals / tokenPrice
        // This converts USD amount into BUNKER amount.
        bunkerCost = _ceilDiv(usdCost * (10 ** oracleDecimals), tokenPrice);
    }

    /// @dev Validate that a provider's price is within allowed bounds of the default.
    function _validateProviderPrice(
        uint256 providerPrice,
        uint256 defaultPrice,
        string memory /* resourceName */
    ) internal view {
        if (providerPrice == 0) revert ZeroPrice();
        uint256 minAllowed = (defaultPrice * minPriceMultiplierBps) / BPS_DENOMINATOR;
        uint256 maxAllowed = (defaultPrice * maxPriceMultiplierBps) / BPS_DENOMINATOR;
        if (providerPrice < minAllowed) revert PriceBelowMinimum(providerPrice, minAllowed);
        if (providerPrice > maxAllowed) revert PriceAboveMaximum(providerPrice, maxAllowed);
    }

    /// @dev Ceiling division: returns ceil(a / b). Returns 0 if a is 0.
    function _ceilDiv(uint256 a, uint256 b) internal pure returns (uint256) {
        return a == 0 ? 0 : (a - 1) / b + 1;
    }

    /// @dev Clamp a provider price to the current allowed bounds relative to the default.
    ///      If the provider price is below the minimum, returns the minimum.
    ///      If the provider price is above the maximum, returns the maximum.
    function _clampToProviderBounds(uint256 providerPrice, uint256 defaultPrice)
        internal
        view
        returns (uint256)
    {
        uint256 minPrice = (defaultPrice * minPriceMultiplierBps) / BPS_DENOMINATOR;
        uint256 maxPrice = (defaultPrice * maxPriceMultiplierBps) / BPS_DENOMINATOR;
        if (providerPrice < minPrice) return minPrice;
        if (providerPrice > maxPrice) return maxPrice;
        return providerPrice;
    }

    /// @dev Emit the full prices updated event.
    function _emitPricesUpdated() internal {
        emit PricesUpdated(
            prices.cpuPerCoreHour,
            prices.memoryPerGBHour,
            prices.storagePerGBMonth,
            prices.networkPerGB,
            prices.gpuBasicPerHour,
            prices.gpuPremiumPerHour
        );
    }
}
