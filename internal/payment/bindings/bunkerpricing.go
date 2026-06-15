// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// BunkerPricingMultipliers is an auto generated low-level Go binding around an user-defined struct.
type BunkerPricingMultipliers struct {
	Redundancy *big.Int
	Tor        *big.Int
	PremiumSLA *big.Int
	Spot       *big.Int
}

// BunkerPricingProviderPricing is an auto generated low-level Go binding around an user-defined struct.
type BunkerPricingProviderPricing struct {
	CpuPerCoreHour    *big.Int
	MemoryPerGBHour   *big.Int
	StoragePerGBMonth *big.Int
	NetworkPerGB      *big.Int
	IsSet             bool
}

// BunkerPricingResourcePrices is an auto generated low-level Go binding around an user-defined struct.
type BunkerPricingResourcePrices struct {
	CpuPerCoreHour    *big.Int
	MemoryPerGBHour   *big.Int
	StoragePerGBMonth *big.Int
	NetworkPerGB      *big.Int
	GpuBasicPerHour   *big.Int
	GpuPremiumPerHour *big.Int
}

// BunkerPricingResourceRequest is an auto generated low-level Go binding around an user-defined struct.
type BunkerPricingResourceRequest struct {
	CpuCores      *big.Int
	MemoryGB      *big.Int
	StorageGB     *big.Int
	NetworkGB     *big.Int
	DurationHours *big.Int
	UseGPUBasic   bool
	UseGPUPremium bool
	UseTor        bool
	UsePremiumSLA bool
	UseSpot       bool
}

// BunkerPricingMetaData contains all meta data concerning the BunkerPricing contract.
var BunkerPricingMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"BPS_DENOMINATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"calculateCost\",\"inputs\":[{\"name\":\"req\",\"type\":\"tuple\",\"internalType\":\"structBunkerPricing.ResourceRequest\",\"components\":[{\"name\":\"cpuCores\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"memoryGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"storageGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"networkGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"durationHours\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"useGPUBasic\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"useGPUPremium\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"useTor\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"usePremiumSLA\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"useSpot\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"totalCost\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"calculateCostUSD\",\"inputs\":[{\"name\":\"req\",\"type\":\"tuple\",\"internalType\":\"structBunkerPricing.ResourceRequest\",\"components\":[{\"name\":\"cpuCores\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"memoryGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"storageGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"networkGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"durationHours\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"useGPUBasic\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"useGPUPremium\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"useTor\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"usePremiumSLA\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"useSpot\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"usdCost\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"calculateProviderCost\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"req\",\"type\":\"tuple\",\"internalType\":\"structBunkerPricing.ResourceRequest\",\"components\":[{\"name\":\"cpuCores\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"memoryGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"storageGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"networkGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"durationHours\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"useGPUBasic\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"useGPUPremium\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"useTor\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"usePremiumSLA\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"useSpot\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"totalCost\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"clearProviderPrices\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"enableOraclePricing\",\"inputs\":[{\"name\":\"enabled\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getMultipliers\",\"inputs\":[],\"outputs\":[{\"name\":\"m\",\"type\":\"tuple\",\"internalType\":\"structBunkerPricing.Multipliers\",\"components\":[{\"name\":\"redundancy\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"tor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"premiumSLA\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"spot\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPrices\",\"inputs\":[],\"outputs\":[{\"name\":\"p\",\"type\":\"tuple\",\"internalType\":\"structBunkerPricing.ResourcePrices\",\"components\":[{\"name\":\"cpuPerCoreHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"memoryPerGBHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"storagePerGBMonth\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"networkPerGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gpuBasicPerHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gpuPremiumPerHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getProviderPrices\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"pp\",\"type\":\"tuple\",\"internalType\":\"structBunkerPricing.ProviderPricing\",\"components\":[{\"name\":\"cpuPerCoreHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"memoryPerGBHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"storagePerGBMonth\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"networkPerGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"isSet\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTokenPrice\",\"inputs\":[],\"outputs\":[{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"decimals_\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxMultiplierBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxPriceMultiplierBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxStaleThreshold\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"minPriceMultiplierBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"minStaleThreshold\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"multipliers\",\"inputs\":[],\"outputs\":[{\"name\":\"redundancy\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"tor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"premiumSLA\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"spot\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"priceOracle\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractAggregatorV3Interface\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"prices\",\"inputs\":[],\"outputs\":[{\"name\":\"cpuPerCoreHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"memoryPerGBHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"storagePerGBMonth\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"networkPerGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gpuBasicPerHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gpuPremiumPerHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"providerPrices\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"cpuPerCoreHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"memoryPerGBHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"storagePerGBMonth\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"networkPerGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"isSet\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setCPUPrice\",\"inputs\":[{\"name\":\"newPrice\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setGPUPrices\",\"inputs\":[{\"name\":\"basicPerHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"premiumPerHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxMultiplierBps\",\"inputs\":[{\"name\":\"newMax\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMemoryPrice\",\"inputs\":[{\"name\":\"newPrice\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMultipliers\",\"inputs\":[{\"name\":\"newMultipliers\",\"type\":\"tuple\",\"internalType\":\"structBunkerPricing.Multipliers\",\"components\":[{\"name\":\"redundancy\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"tor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"premiumSLA\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"spot\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setNetworkPrice\",\"inputs\":[{\"name\":\"newPrice\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setPriceOracle\",\"inputs\":[{\"name\":\"newOracle\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setPrices\",\"inputs\":[{\"name\":\"newPrices\",\"type\":\"tuple\",\"internalType\":\"structBunkerPricing.ResourcePrices\",\"components\":[{\"name\":\"cpuPerCoreHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"memoryPerGBHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"storagePerGBMonth\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"networkPerGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gpuBasicPerHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gpuPremiumPerHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setProviderPriceBounds\",\"inputs\":[{\"name\":\"minBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setProviderPrices\",\"inputs\":[{\"name\":\"pp\",\"type\":\"tuple\",\"internalType\":\"structBunkerPricing.ProviderPricing\",\"components\":[{\"name\":\"cpuPerCoreHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"memoryPerGBHour\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"storagePerGBMonth\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"networkPerGB\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"isSet\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setStalePriceThreshold\",\"inputs\":[{\"name\":\"newThreshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setStaleThresholdBounds\",\"inputs\":[{\"name\":\"newMin\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"newMax\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setStoragePrice\",\"inputs\":[{\"name\":\"newPrice\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"stalePriceThreshold\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"useOracleForPricing\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"MaxMultiplierUpdated\",\"inputs\":[{\"name\":\"newMax\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MultipliersUpdated\",\"inputs\":[{\"name\":\"redundancy\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"tor\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"premiumSLA\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"spot\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OraclePricingEnabled\",\"inputs\":[{\"name\":\"enabled\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PriceOracleUpdated\",\"inputs\":[{\"name\":\"oldOracle\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOracle\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PricesUpdated\",\"inputs\":[{\"name\":\"cpuPerCoreHour\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"memoryPerGBHour\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"storagePerGBMonth\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"networkPerGB\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"gpuBasicPerHour\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"gpuPremiumPerHour\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderPriceBoundsUpdated\",\"inputs\":[{\"name\":\"minBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"maxBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderPricesCleared\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderPricesSet\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"cpu\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"memory_\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"storage_\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"network\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"StalePriceThresholdUpdated\",\"inputs\":[{\"name\":\"oldThreshold\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newThreshold\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"StaleThresholdBoundsUpdated\",\"inputs\":[{\"name\":\"newMin\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newMax\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InvalidMultiplier\",\"inputs\":[{\"name\":\"multiplier\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidMultiplierCap\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPriceBounds\",\"inputs\":[{\"name\":\"minBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidStalePriceThreshold\",\"inputs\":[{\"name\":\"threshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidThresholdBounds\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MultiplierTooHigh\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NegativePrice\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OracleNotSet\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"PriceAboveMaximum\",\"inputs\":[{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maximum\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"PriceBelowMinimum\",\"inputs\":[{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minimum\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"StalePriceData\",\"inputs\":[{\"name\":\"updatedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"threshold\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ThresholdTooLong\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroMultiplier\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroPrice\",\"inputs\":[]}]",
}

// BunkerPricingABI is the input ABI used to generate the binding from.
// Deprecated: Use BunkerPricingMetaData.ABI instead.
var BunkerPricingABI = BunkerPricingMetaData.ABI

// BunkerPricing is an auto generated Go binding around an Ethereum contract.
type BunkerPricing struct {
	BunkerPricingCaller     // Read-only binding to the contract
	BunkerPricingTransactor // Write-only binding to the contract
	BunkerPricingFilterer   // Log filterer for contract events
}

// BunkerPricingCaller is an auto generated read-only Go binding around an Ethereum contract.
type BunkerPricingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerPricingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BunkerPricingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerPricingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BunkerPricingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerPricingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BunkerPricingSession struct {
	Contract     *BunkerPricing    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BunkerPricingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BunkerPricingCallerSession struct {
	Contract *BunkerPricingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// BunkerPricingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BunkerPricingTransactorSession struct {
	Contract     *BunkerPricingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// BunkerPricingRaw is an auto generated low-level Go binding around an Ethereum contract.
type BunkerPricingRaw struct {
	Contract *BunkerPricing // Generic contract binding to access the raw methods on
}

// BunkerPricingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BunkerPricingCallerRaw struct {
	Contract *BunkerPricingCaller // Generic read-only contract binding to access the raw methods on
}

// BunkerPricingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BunkerPricingTransactorRaw struct {
	Contract *BunkerPricingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBunkerPricing creates a new instance of BunkerPricing, bound to a specific deployed contract.
func NewBunkerPricing(address common.Address, backend bind.ContractBackend) (*BunkerPricing, error) {
	contract, err := bindBunkerPricing(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BunkerPricing{BunkerPricingCaller: BunkerPricingCaller{contract: contract}, BunkerPricingTransactor: BunkerPricingTransactor{contract: contract}, BunkerPricingFilterer: BunkerPricingFilterer{contract: contract}}, nil
}

// NewBunkerPricingCaller creates a new read-only instance of BunkerPricing, bound to a specific deployed contract.
func NewBunkerPricingCaller(address common.Address, caller bind.ContractCaller) (*BunkerPricingCaller, error) {
	contract, err := bindBunkerPricing(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerPricingCaller{contract: contract}, nil
}

// NewBunkerPricingTransactor creates a new write-only instance of BunkerPricing, bound to a specific deployed contract.
func NewBunkerPricingTransactor(address common.Address, transactor bind.ContractTransactor) (*BunkerPricingTransactor, error) {
	contract, err := bindBunkerPricing(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerPricingTransactor{contract: contract}, nil
}

// NewBunkerPricingFilterer creates a new log filterer instance of BunkerPricing, bound to a specific deployed contract.
func NewBunkerPricingFilterer(address common.Address, filterer bind.ContractFilterer) (*BunkerPricingFilterer, error) {
	contract, err := bindBunkerPricing(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BunkerPricingFilterer{contract: contract}, nil
}

// bindBunkerPricing binds a generic wrapper to an already deployed contract.
func bindBunkerPricing(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BunkerPricingMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerPricing *BunkerPricingRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerPricing.Contract.BunkerPricingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerPricing *BunkerPricingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerPricing.Contract.BunkerPricingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerPricing *BunkerPricingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerPricing.Contract.BunkerPricingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerPricing *BunkerPricingCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerPricing.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerPricing *BunkerPricingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerPricing.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerPricing *BunkerPricingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerPricing.Contract.contract.Transact(opts, method, params...)
}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerPricing *BunkerPricingCaller) BPSDENOMINATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "BPS_DENOMINATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerPricing *BunkerPricingSession) BPSDENOMINATOR() (*big.Int, error) {
	return _BunkerPricing.Contract.BPSDENOMINATOR(&_BunkerPricing.CallOpts)
}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerPricing *BunkerPricingCallerSession) BPSDENOMINATOR() (*big.Int, error) {
	return _BunkerPricing.Contract.BPSDENOMINATOR(&_BunkerPricing.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerPricing *BunkerPricingCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerPricing *BunkerPricingSession) VERSION() (string, error) {
	return _BunkerPricing.Contract.VERSION(&_BunkerPricing.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerPricing *BunkerPricingCallerSession) VERSION() (string, error) {
	return _BunkerPricing.Contract.VERSION(&_BunkerPricing.CallOpts)
}

// CalculateCost is a free data retrieval call binding the contract method 0x0a5b1570.
//
// Solidity: function calculateCost((uint256,uint256,uint256,uint256,uint256,bool,bool,bool,bool,bool) req) view returns(uint256 totalCost)
func (_BunkerPricing *BunkerPricingCaller) CalculateCost(opts *bind.CallOpts, req BunkerPricingResourceRequest) (*big.Int, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "calculateCost", req)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateCost is a free data retrieval call binding the contract method 0x0a5b1570.
//
// Solidity: function calculateCost((uint256,uint256,uint256,uint256,uint256,bool,bool,bool,bool,bool) req) view returns(uint256 totalCost)
func (_BunkerPricing *BunkerPricingSession) CalculateCost(req BunkerPricingResourceRequest) (*big.Int, error) {
	return _BunkerPricing.Contract.CalculateCost(&_BunkerPricing.CallOpts, req)
}

// CalculateCost is a free data retrieval call binding the contract method 0x0a5b1570.
//
// Solidity: function calculateCost((uint256,uint256,uint256,uint256,uint256,bool,bool,bool,bool,bool) req) view returns(uint256 totalCost)
func (_BunkerPricing *BunkerPricingCallerSession) CalculateCost(req BunkerPricingResourceRequest) (*big.Int, error) {
	return _BunkerPricing.Contract.CalculateCost(&_BunkerPricing.CallOpts, req)
}

// CalculateCostUSD is a free data retrieval call binding the contract method 0x9201d7cd.
//
// Solidity: function calculateCostUSD((uint256,uint256,uint256,uint256,uint256,bool,bool,bool,bool,bool) req) view returns(uint256 usdCost)
func (_BunkerPricing *BunkerPricingCaller) CalculateCostUSD(opts *bind.CallOpts, req BunkerPricingResourceRequest) (*big.Int, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "calculateCostUSD", req)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateCostUSD is a free data retrieval call binding the contract method 0x9201d7cd.
//
// Solidity: function calculateCostUSD((uint256,uint256,uint256,uint256,uint256,bool,bool,bool,bool,bool) req) view returns(uint256 usdCost)
func (_BunkerPricing *BunkerPricingSession) CalculateCostUSD(req BunkerPricingResourceRequest) (*big.Int, error) {
	return _BunkerPricing.Contract.CalculateCostUSD(&_BunkerPricing.CallOpts, req)
}

// CalculateCostUSD is a free data retrieval call binding the contract method 0x9201d7cd.
//
// Solidity: function calculateCostUSD((uint256,uint256,uint256,uint256,uint256,bool,bool,bool,bool,bool) req) view returns(uint256 usdCost)
func (_BunkerPricing *BunkerPricingCallerSession) CalculateCostUSD(req BunkerPricingResourceRequest) (*big.Int, error) {
	return _BunkerPricing.Contract.CalculateCostUSD(&_BunkerPricing.CallOpts, req)
}

// CalculateProviderCost is a free data retrieval call binding the contract method 0x0f9198d8.
//
// Solidity: function calculateProviderCost(address provider, (uint256,uint256,uint256,uint256,uint256,bool,bool,bool,bool,bool) req) view returns(uint256 totalCost)
func (_BunkerPricing *BunkerPricingCaller) CalculateProviderCost(opts *bind.CallOpts, provider common.Address, req BunkerPricingResourceRequest) (*big.Int, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "calculateProviderCost", provider, req)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateProviderCost is a free data retrieval call binding the contract method 0x0f9198d8.
//
// Solidity: function calculateProviderCost(address provider, (uint256,uint256,uint256,uint256,uint256,bool,bool,bool,bool,bool) req) view returns(uint256 totalCost)
func (_BunkerPricing *BunkerPricingSession) CalculateProviderCost(provider common.Address, req BunkerPricingResourceRequest) (*big.Int, error) {
	return _BunkerPricing.Contract.CalculateProviderCost(&_BunkerPricing.CallOpts, provider, req)
}

// CalculateProviderCost is a free data retrieval call binding the contract method 0x0f9198d8.
//
// Solidity: function calculateProviderCost(address provider, (uint256,uint256,uint256,uint256,uint256,bool,bool,bool,bool,bool) req) view returns(uint256 totalCost)
func (_BunkerPricing *BunkerPricingCallerSession) CalculateProviderCost(provider common.Address, req BunkerPricingResourceRequest) (*big.Int, error) {
	return _BunkerPricing.Contract.CalculateProviderCost(&_BunkerPricing.CallOpts, provider, req)
}

// GetMultipliers is a free data retrieval call binding the contract method 0x79873f8a.
//
// Solidity: function getMultipliers() view returns((uint256,uint256,uint256,uint256) m)
func (_BunkerPricing *BunkerPricingCaller) GetMultipliers(opts *bind.CallOpts) (BunkerPricingMultipliers, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "getMultipliers")

	if err != nil {
		return *new(BunkerPricingMultipliers), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerPricingMultipliers)).(*BunkerPricingMultipliers)

	return out0, err

}

// GetMultipliers is a free data retrieval call binding the contract method 0x79873f8a.
//
// Solidity: function getMultipliers() view returns((uint256,uint256,uint256,uint256) m)
func (_BunkerPricing *BunkerPricingSession) GetMultipliers() (BunkerPricingMultipliers, error) {
	return _BunkerPricing.Contract.GetMultipliers(&_BunkerPricing.CallOpts)
}

// GetMultipliers is a free data retrieval call binding the contract method 0x79873f8a.
//
// Solidity: function getMultipliers() view returns((uint256,uint256,uint256,uint256) m)
func (_BunkerPricing *BunkerPricingCallerSession) GetMultipliers() (BunkerPricingMultipliers, error) {
	return _BunkerPricing.Contract.GetMultipliers(&_BunkerPricing.CallOpts)
}

// GetPrices is a free data retrieval call binding the contract method 0xbd9a548b.
//
// Solidity: function getPrices() view returns((uint256,uint256,uint256,uint256,uint256,uint256) p)
func (_BunkerPricing *BunkerPricingCaller) GetPrices(opts *bind.CallOpts) (BunkerPricingResourcePrices, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "getPrices")

	if err != nil {
		return *new(BunkerPricingResourcePrices), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerPricingResourcePrices)).(*BunkerPricingResourcePrices)

	return out0, err

}

// GetPrices is a free data retrieval call binding the contract method 0xbd9a548b.
//
// Solidity: function getPrices() view returns((uint256,uint256,uint256,uint256,uint256,uint256) p)
func (_BunkerPricing *BunkerPricingSession) GetPrices() (BunkerPricingResourcePrices, error) {
	return _BunkerPricing.Contract.GetPrices(&_BunkerPricing.CallOpts)
}

// GetPrices is a free data retrieval call binding the contract method 0xbd9a548b.
//
// Solidity: function getPrices() view returns((uint256,uint256,uint256,uint256,uint256,uint256) p)
func (_BunkerPricing *BunkerPricingCallerSession) GetPrices() (BunkerPricingResourcePrices, error) {
	return _BunkerPricing.Contract.GetPrices(&_BunkerPricing.CallOpts)
}

// GetProviderPrices is a free data retrieval call binding the contract method 0x106859b6.
//
// Solidity: function getProviderPrices(address provider) view returns((uint256,uint256,uint256,uint256,bool) pp)
func (_BunkerPricing *BunkerPricingCaller) GetProviderPrices(opts *bind.CallOpts, provider common.Address) (BunkerPricingProviderPricing, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "getProviderPrices", provider)

	if err != nil {
		return *new(BunkerPricingProviderPricing), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerPricingProviderPricing)).(*BunkerPricingProviderPricing)

	return out0, err

}

// GetProviderPrices is a free data retrieval call binding the contract method 0x106859b6.
//
// Solidity: function getProviderPrices(address provider) view returns((uint256,uint256,uint256,uint256,bool) pp)
func (_BunkerPricing *BunkerPricingSession) GetProviderPrices(provider common.Address) (BunkerPricingProviderPricing, error) {
	return _BunkerPricing.Contract.GetProviderPrices(&_BunkerPricing.CallOpts, provider)
}

// GetProviderPrices is a free data retrieval call binding the contract method 0x106859b6.
//
// Solidity: function getProviderPrices(address provider) view returns((uint256,uint256,uint256,uint256,bool) pp)
func (_BunkerPricing *BunkerPricingCallerSession) GetProviderPrices(provider common.Address) (BunkerPricingProviderPricing, error) {
	return _BunkerPricing.Contract.GetProviderPrices(&_BunkerPricing.CallOpts, provider)
}

// GetTokenPrice is a free data retrieval call binding the contract method 0x4b94f50e.
//
// Solidity: function getTokenPrice() view returns(uint256 price, uint8 decimals_)
func (_BunkerPricing *BunkerPricingCaller) GetTokenPrice(opts *bind.CallOpts) (struct {
	Price    *big.Int
	Decimals uint8
}, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "getTokenPrice")

	outstruct := new(struct {
		Price    *big.Int
		Decimals uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Price = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Decimals = *abi.ConvertType(out[1], new(uint8)).(*uint8)

	return *outstruct, err

}

// GetTokenPrice is a free data retrieval call binding the contract method 0x4b94f50e.
//
// Solidity: function getTokenPrice() view returns(uint256 price, uint8 decimals_)
func (_BunkerPricing *BunkerPricingSession) GetTokenPrice() (struct {
	Price    *big.Int
	Decimals uint8
}, error) {
	return _BunkerPricing.Contract.GetTokenPrice(&_BunkerPricing.CallOpts)
}

// GetTokenPrice is a free data retrieval call binding the contract method 0x4b94f50e.
//
// Solidity: function getTokenPrice() view returns(uint256 price, uint8 decimals_)
func (_BunkerPricing *BunkerPricingCallerSession) GetTokenPrice() (struct {
	Price    *big.Int
	Decimals uint8
}, error) {
	return _BunkerPricing.Contract.GetTokenPrice(&_BunkerPricing.CallOpts)
}

// MaxMultiplierBps is a free data retrieval call binding the contract method 0x97313841.
//
// Solidity: function maxMultiplierBps() view returns(uint256)
func (_BunkerPricing *BunkerPricingCaller) MaxMultiplierBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "maxMultiplierBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxMultiplierBps is a free data retrieval call binding the contract method 0x97313841.
//
// Solidity: function maxMultiplierBps() view returns(uint256)
func (_BunkerPricing *BunkerPricingSession) MaxMultiplierBps() (*big.Int, error) {
	return _BunkerPricing.Contract.MaxMultiplierBps(&_BunkerPricing.CallOpts)
}

// MaxMultiplierBps is a free data retrieval call binding the contract method 0x97313841.
//
// Solidity: function maxMultiplierBps() view returns(uint256)
func (_BunkerPricing *BunkerPricingCallerSession) MaxMultiplierBps() (*big.Int, error) {
	return _BunkerPricing.Contract.MaxMultiplierBps(&_BunkerPricing.CallOpts)
}

// MaxPriceMultiplierBps is a free data retrieval call binding the contract method 0xad2e64d0.
//
// Solidity: function maxPriceMultiplierBps() view returns(uint256)
func (_BunkerPricing *BunkerPricingCaller) MaxPriceMultiplierBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "maxPriceMultiplierBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxPriceMultiplierBps is a free data retrieval call binding the contract method 0xad2e64d0.
//
// Solidity: function maxPriceMultiplierBps() view returns(uint256)
func (_BunkerPricing *BunkerPricingSession) MaxPriceMultiplierBps() (*big.Int, error) {
	return _BunkerPricing.Contract.MaxPriceMultiplierBps(&_BunkerPricing.CallOpts)
}

// MaxPriceMultiplierBps is a free data retrieval call binding the contract method 0xad2e64d0.
//
// Solidity: function maxPriceMultiplierBps() view returns(uint256)
func (_BunkerPricing *BunkerPricingCallerSession) MaxPriceMultiplierBps() (*big.Int, error) {
	return _BunkerPricing.Contract.MaxPriceMultiplierBps(&_BunkerPricing.CallOpts)
}

// MaxStaleThreshold is a free data retrieval call binding the contract method 0x13594a90.
//
// Solidity: function maxStaleThreshold() view returns(uint256)
func (_BunkerPricing *BunkerPricingCaller) MaxStaleThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "maxStaleThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxStaleThreshold is a free data retrieval call binding the contract method 0x13594a90.
//
// Solidity: function maxStaleThreshold() view returns(uint256)
func (_BunkerPricing *BunkerPricingSession) MaxStaleThreshold() (*big.Int, error) {
	return _BunkerPricing.Contract.MaxStaleThreshold(&_BunkerPricing.CallOpts)
}

// MaxStaleThreshold is a free data retrieval call binding the contract method 0x13594a90.
//
// Solidity: function maxStaleThreshold() view returns(uint256)
func (_BunkerPricing *BunkerPricingCallerSession) MaxStaleThreshold() (*big.Int, error) {
	return _BunkerPricing.Contract.MaxStaleThreshold(&_BunkerPricing.CallOpts)
}

// MinPriceMultiplierBps is a free data retrieval call binding the contract method 0xa8b80047.
//
// Solidity: function minPriceMultiplierBps() view returns(uint256)
func (_BunkerPricing *BunkerPricingCaller) MinPriceMultiplierBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "minPriceMultiplierBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinPriceMultiplierBps is a free data retrieval call binding the contract method 0xa8b80047.
//
// Solidity: function minPriceMultiplierBps() view returns(uint256)
func (_BunkerPricing *BunkerPricingSession) MinPriceMultiplierBps() (*big.Int, error) {
	return _BunkerPricing.Contract.MinPriceMultiplierBps(&_BunkerPricing.CallOpts)
}

// MinPriceMultiplierBps is a free data retrieval call binding the contract method 0xa8b80047.
//
// Solidity: function minPriceMultiplierBps() view returns(uint256)
func (_BunkerPricing *BunkerPricingCallerSession) MinPriceMultiplierBps() (*big.Int, error) {
	return _BunkerPricing.Contract.MinPriceMultiplierBps(&_BunkerPricing.CallOpts)
}

// MinStaleThreshold is a free data retrieval call binding the contract method 0x3165e9d3.
//
// Solidity: function minStaleThreshold() view returns(uint256)
func (_BunkerPricing *BunkerPricingCaller) MinStaleThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "minStaleThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinStaleThreshold is a free data retrieval call binding the contract method 0x3165e9d3.
//
// Solidity: function minStaleThreshold() view returns(uint256)
func (_BunkerPricing *BunkerPricingSession) MinStaleThreshold() (*big.Int, error) {
	return _BunkerPricing.Contract.MinStaleThreshold(&_BunkerPricing.CallOpts)
}

// MinStaleThreshold is a free data retrieval call binding the contract method 0x3165e9d3.
//
// Solidity: function minStaleThreshold() view returns(uint256)
func (_BunkerPricing *BunkerPricingCallerSession) MinStaleThreshold() (*big.Int, error) {
	return _BunkerPricing.Contract.MinStaleThreshold(&_BunkerPricing.CallOpts)
}

// Multipliers is a free data retrieval call binding the contract method 0x9d7d6667.
//
// Solidity: function multipliers() view returns(uint256 redundancy, uint256 tor, uint256 premiumSLA, uint256 spot)
func (_BunkerPricing *BunkerPricingCaller) Multipliers(opts *bind.CallOpts) (struct {
	Redundancy *big.Int
	Tor        *big.Int
	PremiumSLA *big.Int
	Spot       *big.Int
}, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "multipliers")

	outstruct := new(struct {
		Redundancy *big.Int
		Tor        *big.Int
		PremiumSLA *big.Int
		Spot       *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Redundancy = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Tor = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.PremiumSLA = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Spot = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Multipliers is a free data retrieval call binding the contract method 0x9d7d6667.
//
// Solidity: function multipliers() view returns(uint256 redundancy, uint256 tor, uint256 premiumSLA, uint256 spot)
func (_BunkerPricing *BunkerPricingSession) Multipliers() (struct {
	Redundancy *big.Int
	Tor        *big.Int
	PremiumSLA *big.Int
	Spot       *big.Int
}, error) {
	return _BunkerPricing.Contract.Multipliers(&_BunkerPricing.CallOpts)
}

// Multipliers is a free data retrieval call binding the contract method 0x9d7d6667.
//
// Solidity: function multipliers() view returns(uint256 redundancy, uint256 tor, uint256 premiumSLA, uint256 spot)
func (_BunkerPricing *BunkerPricingCallerSession) Multipliers() (struct {
	Redundancy *big.Int
	Tor        *big.Int
	PremiumSLA *big.Int
	Spot       *big.Int
}, error) {
	return _BunkerPricing.Contract.Multipliers(&_BunkerPricing.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerPricing *BunkerPricingCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerPricing *BunkerPricingSession) Owner() (common.Address, error) {
	return _BunkerPricing.Contract.Owner(&_BunkerPricing.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerPricing *BunkerPricingCallerSession) Owner() (common.Address, error) {
	return _BunkerPricing.Contract.Owner(&_BunkerPricing.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerPricing *BunkerPricingCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerPricing *BunkerPricingSession) PendingOwner() (common.Address, error) {
	return _BunkerPricing.Contract.PendingOwner(&_BunkerPricing.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerPricing *BunkerPricingCallerSession) PendingOwner() (common.Address, error) {
	return _BunkerPricing.Contract.PendingOwner(&_BunkerPricing.CallOpts)
}

// PriceOracle is a free data retrieval call binding the contract method 0x2630c12f.
//
// Solidity: function priceOracle() view returns(address)
func (_BunkerPricing *BunkerPricingCaller) PriceOracle(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "priceOracle")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PriceOracle is a free data retrieval call binding the contract method 0x2630c12f.
//
// Solidity: function priceOracle() view returns(address)
func (_BunkerPricing *BunkerPricingSession) PriceOracle() (common.Address, error) {
	return _BunkerPricing.Contract.PriceOracle(&_BunkerPricing.CallOpts)
}

// PriceOracle is a free data retrieval call binding the contract method 0x2630c12f.
//
// Solidity: function priceOracle() view returns(address)
func (_BunkerPricing *BunkerPricingCallerSession) PriceOracle() (common.Address, error) {
	return _BunkerPricing.Contract.PriceOracle(&_BunkerPricing.CallOpts)
}

// Prices is a free data retrieval call binding the contract method 0xd3419bf3.
//
// Solidity: function prices() view returns(uint256 cpuPerCoreHour, uint256 memoryPerGBHour, uint256 storagePerGBMonth, uint256 networkPerGB, uint256 gpuBasicPerHour, uint256 gpuPremiumPerHour)
func (_BunkerPricing *BunkerPricingCaller) Prices(opts *bind.CallOpts) (struct {
	CpuPerCoreHour    *big.Int
	MemoryPerGBHour   *big.Int
	StoragePerGBMonth *big.Int
	NetworkPerGB      *big.Int
	GpuBasicPerHour   *big.Int
	GpuPremiumPerHour *big.Int
}, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "prices")

	outstruct := new(struct {
		CpuPerCoreHour    *big.Int
		MemoryPerGBHour   *big.Int
		StoragePerGBMonth *big.Int
		NetworkPerGB      *big.Int
		GpuBasicPerHour   *big.Int
		GpuPremiumPerHour *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.CpuPerCoreHour = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.MemoryPerGBHour = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.StoragePerGBMonth = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.NetworkPerGB = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.GpuBasicPerHour = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.GpuPremiumPerHour = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Prices is a free data retrieval call binding the contract method 0xd3419bf3.
//
// Solidity: function prices() view returns(uint256 cpuPerCoreHour, uint256 memoryPerGBHour, uint256 storagePerGBMonth, uint256 networkPerGB, uint256 gpuBasicPerHour, uint256 gpuPremiumPerHour)
func (_BunkerPricing *BunkerPricingSession) Prices() (struct {
	CpuPerCoreHour    *big.Int
	MemoryPerGBHour   *big.Int
	StoragePerGBMonth *big.Int
	NetworkPerGB      *big.Int
	GpuBasicPerHour   *big.Int
	GpuPremiumPerHour *big.Int
}, error) {
	return _BunkerPricing.Contract.Prices(&_BunkerPricing.CallOpts)
}

// Prices is a free data retrieval call binding the contract method 0xd3419bf3.
//
// Solidity: function prices() view returns(uint256 cpuPerCoreHour, uint256 memoryPerGBHour, uint256 storagePerGBMonth, uint256 networkPerGB, uint256 gpuBasicPerHour, uint256 gpuPremiumPerHour)
func (_BunkerPricing *BunkerPricingCallerSession) Prices() (struct {
	CpuPerCoreHour    *big.Int
	MemoryPerGBHour   *big.Int
	StoragePerGBMonth *big.Int
	NetworkPerGB      *big.Int
	GpuBasicPerHour   *big.Int
	GpuPremiumPerHour *big.Int
}, error) {
	return _BunkerPricing.Contract.Prices(&_BunkerPricing.CallOpts)
}

// ProviderPrices is a free data retrieval call binding the contract method 0x48595aad.
//
// Solidity: function providerPrices(address ) view returns(uint256 cpuPerCoreHour, uint256 memoryPerGBHour, uint256 storagePerGBMonth, uint256 networkPerGB, bool isSet)
func (_BunkerPricing *BunkerPricingCaller) ProviderPrices(opts *bind.CallOpts, arg0 common.Address) (struct {
	CpuPerCoreHour    *big.Int
	MemoryPerGBHour   *big.Int
	StoragePerGBMonth *big.Int
	NetworkPerGB      *big.Int
	IsSet             bool
}, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "providerPrices", arg0)

	outstruct := new(struct {
		CpuPerCoreHour    *big.Int
		MemoryPerGBHour   *big.Int
		StoragePerGBMonth *big.Int
		NetworkPerGB      *big.Int
		IsSet             bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.CpuPerCoreHour = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.MemoryPerGBHour = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.StoragePerGBMonth = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.NetworkPerGB = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.IsSet = *abi.ConvertType(out[4], new(bool)).(*bool)

	return *outstruct, err

}

// ProviderPrices is a free data retrieval call binding the contract method 0x48595aad.
//
// Solidity: function providerPrices(address ) view returns(uint256 cpuPerCoreHour, uint256 memoryPerGBHour, uint256 storagePerGBMonth, uint256 networkPerGB, bool isSet)
func (_BunkerPricing *BunkerPricingSession) ProviderPrices(arg0 common.Address) (struct {
	CpuPerCoreHour    *big.Int
	MemoryPerGBHour   *big.Int
	StoragePerGBMonth *big.Int
	NetworkPerGB      *big.Int
	IsSet             bool
}, error) {
	return _BunkerPricing.Contract.ProviderPrices(&_BunkerPricing.CallOpts, arg0)
}

// ProviderPrices is a free data retrieval call binding the contract method 0x48595aad.
//
// Solidity: function providerPrices(address ) view returns(uint256 cpuPerCoreHour, uint256 memoryPerGBHour, uint256 storagePerGBMonth, uint256 networkPerGB, bool isSet)
func (_BunkerPricing *BunkerPricingCallerSession) ProviderPrices(arg0 common.Address) (struct {
	CpuPerCoreHour    *big.Int
	MemoryPerGBHour   *big.Int
	StoragePerGBMonth *big.Int
	NetworkPerGB      *big.Int
	IsSet             bool
}, error) {
	return _BunkerPricing.Contract.ProviderPrices(&_BunkerPricing.CallOpts, arg0)
}

// StalePriceThreshold is a free data retrieval call binding the contract method 0x08cb2210.
//
// Solidity: function stalePriceThreshold() view returns(uint256)
func (_BunkerPricing *BunkerPricingCaller) StalePriceThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "stalePriceThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StalePriceThreshold is a free data retrieval call binding the contract method 0x08cb2210.
//
// Solidity: function stalePriceThreshold() view returns(uint256)
func (_BunkerPricing *BunkerPricingSession) StalePriceThreshold() (*big.Int, error) {
	return _BunkerPricing.Contract.StalePriceThreshold(&_BunkerPricing.CallOpts)
}

// StalePriceThreshold is a free data retrieval call binding the contract method 0x08cb2210.
//
// Solidity: function stalePriceThreshold() view returns(uint256)
func (_BunkerPricing *BunkerPricingCallerSession) StalePriceThreshold() (*big.Int, error) {
	return _BunkerPricing.Contract.StalePriceThreshold(&_BunkerPricing.CallOpts)
}

// UseOracleForPricing is a free data retrieval call binding the contract method 0xcadf8ac0.
//
// Solidity: function useOracleForPricing() view returns(bool)
func (_BunkerPricing *BunkerPricingCaller) UseOracleForPricing(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BunkerPricing.contract.Call(opts, &out, "useOracleForPricing")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// UseOracleForPricing is a free data retrieval call binding the contract method 0xcadf8ac0.
//
// Solidity: function useOracleForPricing() view returns(bool)
func (_BunkerPricing *BunkerPricingSession) UseOracleForPricing() (bool, error) {
	return _BunkerPricing.Contract.UseOracleForPricing(&_BunkerPricing.CallOpts)
}

// UseOracleForPricing is a free data retrieval call binding the contract method 0xcadf8ac0.
//
// Solidity: function useOracleForPricing() view returns(bool)
func (_BunkerPricing *BunkerPricingCallerSession) UseOracleForPricing() (bool, error) {
	return _BunkerPricing.Contract.UseOracleForPricing(&_BunkerPricing.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerPricing *BunkerPricingTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerPricing *BunkerPricingSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerPricing.Contract.AcceptOwnership(&_BunkerPricing.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerPricing *BunkerPricingTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerPricing.Contract.AcceptOwnership(&_BunkerPricing.TransactOpts)
}

// ClearProviderPrices is a paid mutator transaction binding the contract method 0x6152ee17.
//
// Solidity: function clearProviderPrices() returns()
func (_BunkerPricing *BunkerPricingTransactor) ClearProviderPrices(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "clearProviderPrices")
}

// ClearProviderPrices is a paid mutator transaction binding the contract method 0x6152ee17.
//
// Solidity: function clearProviderPrices() returns()
func (_BunkerPricing *BunkerPricingSession) ClearProviderPrices() (*types.Transaction, error) {
	return _BunkerPricing.Contract.ClearProviderPrices(&_BunkerPricing.TransactOpts)
}

// ClearProviderPrices is a paid mutator transaction binding the contract method 0x6152ee17.
//
// Solidity: function clearProviderPrices() returns()
func (_BunkerPricing *BunkerPricingTransactorSession) ClearProviderPrices() (*types.Transaction, error) {
	return _BunkerPricing.Contract.ClearProviderPrices(&_BunkerPricing.TransactOpts)
}

// EnableOraclePricing is a paid mutator transaction binding the contract method 0x827055f5.
//
// Solidity: function enableOraclePricing(bool enabled) returns()
func (_BunkerPricing *BunkerPricingTransactor) EnableOraclePricing(opts *bind.TransactOpts, enabled bool) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "enableOraclePricing", enabled)
}

// EnableOraclePricing is a paid mutator transaction binding the contract method 0x827055f5.
//
// Solidity: function enableOraclePricing(bool enabled) returns()
func (_BunkerPricing *BunkerPricingSession) EnableOraclePricing(enabled bool) (*types.Transaction, error) {
	return _BunkerPricing.Contract.EnableOraclePricing(&_BunkerPricing.TransactOpts, enabled)
}

// EnableOraclePricing is a paid mutator transaction binding the contract method 0x827055f5.
//
// Solidity: function enableOraclePricing(bool enabled) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) EnableOraclePricing(enabled bool) (*types.Transaction, error) {
	return _BunkerPricing.Contract.EnableOraclePricing(&_BunkerPricing.TransactOpts, enabled)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerPricing *BunkerPricingTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerPricing *BunkerPricingSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerPricing.Contract.RenounceOwnership(&_BunkerPricing.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerPricing *BunkerPricingTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerPricing.Contract.RenounceOwnership(&_BunkerPricing.TransactOpts)
}

// SetCPUPrice is a paid mutator transaction binding the contract method 0xd20bb1d3.
//
// Solidity: function setCPUPrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetCPUPrice(opts *bind.TransactOpts, newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setCPUPrice", newPrice)
}

// SetCPUPrice is a paid mutator transaction binding the contract method 0xd20bb1d3.
//
// Solidity: function setCPUPrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingSession) SetCPUPrice(newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetCPUPrice(&_BunkerPricing.TransactOpts, newPrice)
}

// SetCPUPrice is a paid mutator transaction binding the contract method 0xd20bb1d3.
//
// Solidity: function setCPUPrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetCPUPrice(newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetCPUPrice(&_BunkerPricing.TransactOpts, newPrice)
}

// SetGPUPrices is a paid mutator transaction binding the contract method 0xac2cc61e.
//
// Solidity: function setGPUPrices(uint256 basicPerHour, uint256 premiumPerHour) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetGPUPrices(opts *bind.TransactOpts, basicPerHour *big.Int, premiumPerHour *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setGPUPrices", basicPerHour, premiumPerHour)
}

// SetGPUPrices is a paid mutator transaction binding the contract method 0xac2cc61e.
//
// Solidity: function setGPUPrices(uint256 basicPerHour, uint256 premiumPerHour) returns()
func (_BunkerPricing *BunkerPricingSession) SetGPUPrices(basicPerHour *big.Int, premiumPerHour *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetGPUPrices(&_BunkerPricing.TransactOpts, basicPerHour, premiumPerHour)
}

// SetGPUPrices is a paid mutator transaction binding the contract method 0xac2cc61e.
//
// Solidity: function setGPUPrices(uint256 basicPerHour, uint256 premiumPerHour) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetGPUPrices(basicPerHour *big.Int, premiumPerHour *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetGPUPrices(&_BunkerPricing.TransactOpts, basicPerHour, premiumPerHour)
}

// SetMaxMultiplierBps is a paid mutator transaction binding the contract method 0x1f403947.
//
// Solidity: function setMaxMultiplierBps(uint256 newMax) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetMaxMultiplierBps(opts *bind.TransactOpts, newMax *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setMaxMultiplierBps", newMax)
}

// SetMaxMultiplierBps is a paid mutator transaction binding the contract method 0x1f403947.
//
// Solidity: function setMaxMultiplierBps(uint256 newMax) returns()
func (_BunkerPricing *BunkerPricingSession) SetMaxMultiplierBps(newMax *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetMaxMultiplierBps(&_BunkerPricing.TransactOpts, newMax)
}

// SetMaxMultiplierBps is a paid mutator transaction binding the contract method 0x1f403947.
//
// Solidity: function setMaxMultiplierBps(uint256 newMax) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetMaxMultiplierBps(newMax *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetMaxMultiplierBps(&_BunkerPricing.TransactOpts, newMax)
}

// SetMemoryPrice is a paid mutator transaction binding the contract method 0x65b1b3ef.
//
// Solidity: function setMemoryPrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetMemoryPrice(opts *bind.TransactOpts, newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setMemoryPrice", newPrice)
}

// SetMemoryPrice is a paid mutator transaction binding the contract method 0x65b1b3ef.
//
// Solidity: function setMemoryPrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingSession) SetMemoryPrice(newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetMemoryPrice(&_BunkerPricing.TransactOpts, newPrice)
}

// SetMemoryPrice is a paid mutator transaction binding the contract method 0x65b1b3ef.
//
// Solidity: function setMemoryPrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetMemoryPrice(newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetMemoryPrice(&_BunkerPricing.TransactOpts, newPrice)
}

// SetMultipliers is a paid mutator transaction binding the contract method 0xee02c360.
//
// Solidity: function setMultipliers((uint256,uint256,uint256,uint256) newMultipliers) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetMultipliers(opts *bind.TransactOpts, newMultipliers BunkerPricingMultipliers) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setMultipliers", newMultipliers)
}

// SetMultipliers is a paid mutator transaction binding the contract method 0xee02c360.
//
// Solidity: function setMultipliers((uint256,uint256,uint256,uint256) newMultipliers) returns()
func (_BunkerPricing *BunkerPricingSession) SetMultipliers(newMultipliers BunkerPricingMultipliers) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetMultipliers(&_BunkerPricing.TransactOpts, newMultipliers)
}

// SetMultipliers is a paid mutator transaction binding the contract method 0xee02c360.
//
// Solidity: function setMultipliers((uint256,uint256,uint256,uint256) newMultipliers) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetMultipliers(newMultipliers BunkerPricingMultipliers) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetMultipliers(&_BunkerPricing.TransactOpts, newMultipliers)
}

// SetNetworkPrice is a paid mutator transaction binding the contract method 0xd2c5ed86.
//
// Solidity: function setNetworkPrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetNetworkPrice(opts *bind.TransactOpts, newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setNetworkPrice", newPrice)
}

// SetNetworkPrice is a paid mutator transaction binding the contract method 0xd2c5ed86.
//
// Solidity: function setNetworkPrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingSession) SetNetworkPrice(newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetNetworkPrice(&_BunkerPricing.TransactOpts, newPrice)
}

// SetNetworkPrice is a paid mutator transaction binding the contract method 0xd2c5ed86.
//
// Solidity: function setNetworkPrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetNetworkPrice(newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetNetworkPrice(&_BunkerPricing.TransactOpts, newPrice)
}

// SetPriceOracle is a paid mutator transaction binding the contract method 0x530e784f.
//
// Solidity: function setPriceOracle(address newOracle) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetPriceOracle(opts *bind.TransactOpts, newOracle common.Address) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setPriceOracle", newOracle)
}

// SetPriceOracle is a paid mutator transaction binding the contract method 0x530e784f.
//
// Solidity: function setPriceOracle(address newOracle) returns()
func (_BunkerPricing *BunkerPricingSession) SetPriceOracle(newOracle common.Address) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetPriceOracle(&_BunkerPricing.TransactOpts, newOracle)
}

// SetPriceOracle is a paid mutator transaction binding the contract method 0x530e784f.
//
// Solidity: function setPriceOracle(address newOracle) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetPriceOracle(newOracle common.Address) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetPriceOracle(&_BunkerPricing.TransactOpts, newOracle)
}

// SetPrices is a paid mutator transaction binding the contract method 0x03d4b7c1.
//
// Solidity: function setPrices((uint256,uint256,uint256,uint256,uint256,uint256) newPrices) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetPrices(opts *bind.TransactOpts, newPrices BunkerPricingResourcePrices) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setPrices", newPrices)
}

// SetPrices is a paid mutator transaction binding the contract method 0x03d4b7c1.
//
// Solidity: function setPrices((uint256,uint256,uint256,uint256,uint256,uint256) newPrices) returns()
func (_BunkerPricing *BunkerPricingSession) SetPrices(newPrices BunkerPricingResourcePrices) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetPrices(&_BunkerPricing.TransactOpts, newPrices)
}

// SetPrices is a paid mutator transaction binding the contract method 0x03d4b7c1.
//
// Solidity: function setPrices((uint256,uint256,uint256,uint256,uint256,uint256) newPrices) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetPrices(newPrices BunkerPricingResourcePrices) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetPrices(&_BunkerPricing.TransactOpts, newPrices)
}

// SetProviderPriceBounds is a paid mutator transaction binding the contract method 0xaa1328f1.
//
// Solidity: function setProviderPriceBounds(uint256 minBps, uint256 maxBps) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetProviderPriceBounds(opts *bind.TransactOpts, minBps *big.Int, maxBps *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setProviderPriceBounds", minBps, maxBps)
}

// SetProviderPriceBounds is a paid mutator transaction binding the contract method 0xaa1328f1.
//
// Solidity: function setProviderPriceBounds(uint256 minBps, uint256 maxBps) returns()
func (_BunkerPricing *BunkerPricingSession) SetProviderPriceBounds(minBps *big.Int, maxBps *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetProviderPriceBounds(&_BunkerPricing.TransactOpts, minBps, maxBps)
}

// SetProviderPriceBounds is a paid mutator transaction binding the contract method 0xaa1328f1.
//
// Solidity: function setProviderPriceBounds(uint256 minBps, uint256 maxBps) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetProviderPriceBounds(minBps *big.Int, maxBps *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetProviderPriceBounds(&_BunkerPricing.TransactOpts, minBps, maxBps)
}

// SetProviderPrices is a paid mutator transaction binding the contract method 0x5efb84df.
//
// Solidity: function setProviderPrices((uint256,uint256,uint256,uint256,bool) pp) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetProviderPrices(opts *bind.TransactOpts, pp BunkerPricingProviderPricing) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setProviderPrices", pp)
}

// SetProviderPrices is a paid mutator transaction binding the contract method 0x5efb84df.
//
// Solidity: function setProviderPrices((uint256,uint256,uint256,uint256,bool) pp) returns()
func (_BunkerPricing *BunkerPricingSession) SetProviderPrices(pp BunkerPricingProviderPricing) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetProviderPrices(&_BunkerPricing.TransactOpts, pp)
}

// SetProviderPrices is a paid mutator transaction binding the contract method 0x5efb84df.
//
// Solidity: function setProviderPrices((uint256,uint256,uint256,uint256,bool) pp) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetProviderPrices(pp BunkerPricingProviderPricing) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetProviderPrices(&_BunkerPricing.TransactOpts, pp)
}

// SetStalePriceThreshold is a paid mutator transaction binding the contract method 0x2599af34.
//
// Solidity: function setStalePriceThreshold(uint256 newThreshold) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetStalePriceThreshold(opts *bind.TransactOpts, newThreshold *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setStalePriceThreshold", newThreshold)
}

// SetStalePriceThreshold is a paid mutator transaction binding the contract method 0x2599af34.
//
// Solidity: function setStalePriceThreshold(uint256 newThreshold) returns()
func (_BunkerPricing *BunkerPricingSession) SetStalePriceThreshold(newThreshold *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetStalePriceThreshold(&_BunkerPricing.TransactOpts, newThreshold)
}

// SetStalePriceThreshold is a paid mutator transaction binding the contract method 0x2599af34.
//
// Solidity: function setStalePriceThreshold(uint256 newThreshold) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetStalePriceThreshold(newThreshold *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetStalePriceThreshold(&_BunkerPricing.TransactOpts, newThreshold)
}

// SetStaleThresholdBounds is a paid mutator transaction binding the contract method 0x2ac310a3.
//
// Solidity: function setStaleThresholdBounds(uint256 newMin, uint256 newMax) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetStaleThresholdBounds(opts *bind.TransactOpts, newMin *big.Int, newMax *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setStaleThresholdBounds", newMin, newMax)
}

// SetStaleThresholdBounds is a paid mutator transaction binding the contract method 0x2ac310a3.
//
// Solidity: function setStaleThresholdBounds(uint256 newMin, uint256 newMax) returns()
func (_BunkerPricing *BunkerPricingSession) SetStaleThresholdBounds(newMin *big.Int, newMax *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetStaleThresholdBounds(&_BunkerPricing.TransactOpts, newMin, newMax)
}

// SetStaleThresholdBounds is a paid mutator transaction binding the contract method 0x2ac310a3.
//
// Solidity: function setStaleThresholdBounds(uint256 newMin, uint256 newMax) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetStaleThresholdBounds(newMin *big.Int, newMax *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetStaleThresholdBounds(&_BunkerPricing.TransactOpts, newMin, newMax)
}

// SetStoragePrice is a paid mutator transaction binding the contract method 0x6e838911.
//
// Solidity: function setStoragePrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingTransactor) SetStoragePrice(opts *bind.TransactOpts, newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "setStoragePrice", newPrice)
}

// SetStoragePrice is a paid mutator transaction binding the contract method 0x6e838911.
//
// Solidity: function setStoragePrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingSession) SetStoragePrice(newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetStoragePrice(&_BunkerPricing.TransactOpts, newPrice)
}

// SetStoragePrice is a paid mutator transaction binding the contract method 0x6e838911.
//
// Solidity: function setStoragePrice(uint256 newPrice) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) SetStoragePrice(newPrice *big.Int) (*types.Transaction, error) {
	return _BunkerPricing.Contract.SetStoragePrice(&_BunkerPricing.TransactOpts, newPrice)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerPricing *BunkerPricingTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BunkerPricing.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerPricing *BunkerPricingSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerPricing.Contract.TransferOwnership(&_BunkerPricing.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerPricing *BunkerPricingTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerPricing.Contract.TransferOwnership(&_BunkerPricing.TransactOpts, newOwner)
}

// BunkerPricingMaxMultiplierUpdatedIterator is returned from FilterMaxMultiplierUpdated and is used to iterate over the raw logs and unpacked data for MaxMultiplierUpdated events raised by the BunkerPricing contract.
type BunkerPricingMaxMultiplierUpdatedIterator struct {
	Event *BunkerPricingMaxMultiplierUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingMaxMultiplierUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingMaxMultiplierUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingMaxMultiplierUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingMaxMultiplierUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingMaxMultiplierUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingMaxMultiplierUpdated represents a MaxMultiplierUpdated event raised by the BunkerPricing contract.
type BunkerPricingMaxMultiplierUpdated struct {
	NewMax *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMaxMultiplierUpdated is a free log retrieval operation binding the contract event 0xc1ba18bc88cda71f03136908dce4e1a2319aeed38972c3c1a6a470f2d87a7432.
//
// Solidity: event MaxMultiplierUpdated(uint256 newMax)
func (_BunkerPricing *BunkerPricingFilterer) FilterMaxMultiplierUpdated(opts *bind.FilterOpts) (*BunkerPricingMaxMultiplierUpdatedIterator, error) {

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "MaxMultiplierUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerPricingMaxMultiplierUpdatedIterator{contract: _BunkerPricing.contract, event: "MaxMultiplierUpdated", logs: logs, sub: sub}, nil
}

// WatchMaxMultiplierUpdated is a free log subscription operation binding the contract event 0xc1ba18bc88cda71f03136908dce4e1a2319aeed38972c3c1a6a470f2d87a7432.
//
// Solidity: event MaxMultiplierUpdated(uint256 newMax)
func (_BunkerPricing *BunkerPricingFilterer) WatchMaxMultiplierUpdated(opts *bind.WatchOpts, sink chan<- *BunkerPricingMaxMultiplierUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "MaxMultiplierUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingMaxMultiplierUpdated)
				if err := _BunkerPricing.contract.UnpackLog(event, "MaxMultiplierUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMaxMultiplierUpdated is a log parse operation binding the contract event 0xc1ba18bc88cda71f03136908dce4e1a2319aeed38972c3c1a6a470f2d87a7432.
//
// Solidity: event MaxMultiplierUpdated(uint256 newMax)
func (_BunkerPricing *BunkerPricingFilterer) ParseMaxMultiplierUpdated(log types.Log) (*BunkerPricingMaxMultiplierUpdated, error) {
	event := new(BunkerPricingMaxMultiplierUpdated)
	if err := _BunkerPricing.contract.UnpackLog(event, "MaxMultiplierUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingMultipliersUpdatedIterator is returned from FilterMultipliersUpdated and is used to iterate over the raw logs and unpacked data for MultipliersUpdated events raised by the BunkerPricing contract.
type BunkerPricingMultipliersUpdatedIterator struct {
	Event *BunkerPricingMultipliersUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingMultipliersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingMultipliersUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingMultipliersUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingMultipliersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingMultipliersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingMultipliersUpdated represents a MultipliersUpdated event raised by the BunkerPricing contract.
type BunkerPricingMultipliersUpdated struct {
	Redundancy *big.Int
	Tor        *big.Int
	PremiumSLA *big.Int
	Spot       *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterMultipliersUpdated is a free log retrieval operation binding the contract event 0x0d49a1e5ea8841770dfddfd4823e83d41b1f8f573baf50150c7455986f5033c1.
//
// Solidity: event MultipliersUpdated(uint256 redundancy, uint256 tor, uint256 premiumSLA, uint256 spot)
func (_BunkerPricing *BunkerPricingFilterer) FilterMultipliersUpdated(opts *bind.FilterOpts) (*BunkerPricingMultipliersUpdatedIterator, error) {

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "MultipliersUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerPricingMultipliersUpdatedIterator{contract: _BunkerPricing.contract, event: "MultipliersUpdated", logs: logs, sub: sub}, nil
}

// WatchMultipliersUpdated is a free log subscription operation binding the contract event 0x0d49a1e5ea8841770dfddfd4823e83d41b1f8f573baf50150c7455986f5033c1.
//
// Solidity: event MultipliersUpdated(uint256 redundancy, uint256 tor, uint256 premiumSLA, uint256 spot)
func (_BunkerPricing *BunkerPricingFilterer) WatchMultipliersUpdated(opts *bind.WatchOpts, sink chan<- *BunkerPricingMultipliersUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "MultipliersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingMultipliersUpdated)
				if err := _BunkerPricing.contract.UnpackLog(event, "MultipliersUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMultipliersUpdated is a log parse operation binding the contract event 0x0d49a1e5ea8841770dfddfd4823e83d41b1f8f573baf50150c7455986f5033c1.
//
// Solidity: event MultipliersUpdated(uint256 redundancy, uint256 tor, uint256 premiumSLA, uint256 spot)
func (_BunkerPricing *BunkerPricingFilterer) ParseMultipliersUpdated(log types.Log) (*BunkerPricingMultipliersUpdated, error) {
	event := new(BunkerPricingMultipliersUpdated)
	if err := _BunkerPricing.contract.UnpackLog(event, "MultipliersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingOraclePricingEnabledIterator is returned from FilterOraclePricingEnabled and is used to iterate over the raw logs and unpacked data for OraclePricingEnabled events raised by the BunkerPricing contract.
type BunkerPricingOraclePricingEnabledIterator struct {
	Event *BunkerPricingOraclePricingEnabled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingOraclePricingEnabledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingOraclePricingEnabled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingOraclePricingEnabled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingOraclePricingEnabledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingOraclePricingEnabledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingOraclePricingEnabled represents a OraclePricingEnabled event raised by the BunkerPricing contract.
type BunkerPricingOraclePricingEnabled struct {
	Enabled bool
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterOraclePricingEnabled is a free log retrieval operation binding the contract event 0xfceaa4c758e5decf6cc9a516d36edad4529d05d7f343a5e589b1ed36ee49991f.
//
// Solidity: event OraclePricingEnabled(bool enabled)
func (_BunkerPricing *BunkerPricingFilterer) FilterOraclePricingEnabled(opts *bind.FilterOpts) (*BunkerPricingOraclePricingEnabledIterator, error) {

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "OraclePricingEnabled")
	if err != nil {
		return nil, err
	}
	return &BunkerPricingOraclePricingEnabledIterator{contract: _BunkerPricing.contract, event: "OraclePricingEnabled", logs: logs, sub: sub}, nil
}

// WatchOraclePricingEnabled is a free log subscription operation binding the contract event 0xfceaa4c758e5decf6cc9a516d36edad4529d05d7f343a5e589b1ed36ee49991f.
//
// Solidity: event OraclePricingEnabled(bool enabled)
func (_BunkerPricing *BunkerPricingFilterer) WatchOraclePricingEnabled(opts *bind.WatchOpts, sink chan<- *BunkerPricingOraclePricingEnabled) (event.Subscription, error) {

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "OraclePricingEnabled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingOraclePricingEnabled)
				if err := _BunkerPricing.contract.UnpackLog(event, "OraclePricingEnabled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOraclePricingEnabled is a log parse operation binding the contract event 0xfceaa4c758e5decf6cc9a516d36edad4529d05d7f343a5e589b1ed36ee49991f.
//
// Solidity: event OraclePricingEnabled(bool enabled)
func (_BunkerPricing *BunkerPricingFilterer) ParseOraclePricingEnabled(log types.Log) (*BunkerPricingOraclePricingEnabled, error) {
	event := new(BunkerPricingOraclePricingEnabled)
	if err := _BunkerPricing.contract.UnpackLog(event, "OraclePricingEnabled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the BunkerPricing contract.
type BunkerPricingOwnershipTransferStartedIterator struct {
	Event *BunkerPricingOwnershipTransferStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingOwnershipTransferStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingOwnershipTransferStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the BunkerPricing contract.
type BunkerPricingOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerPricing *BunkerPricingFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerPricingOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerPricingOwnershipTransferStartedIterator{contract: _BunkerPricing.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerPricing *BunkerPricingFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *BunkerPricingOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingOwnershipTransferStarted)
				if err := _BunkerPricing.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferStarted is a log parse operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerPricing *BunkerPricingFilterer) ParseOwnershipTransferStarted(log types.Log) (*BunkerPricingOwnershipTransferStarted, error) {
	event := new(BunkerPricingOwnershipTransferStarted)
	if err := _BunkerPricing.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BunkerPricing contract.
type BunkerPricingOwnershipTransferredIterator struct {
	Event *BunkerPricingOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingOwnershipTransferred represents a OwnershipTransferred event raised by the BunkerPricing contract.
type BunkerPricingOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerPricing *BunkerPricingFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerPricingOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerPricingOwnershipTransferredIterator{contract: _BunkerPricing.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerPricing *BunkerPricingFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BunkerPricingOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingOwnershipTransferred)
				if err := _BunkerPricing.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerPricing *BunkerPricingFilterer) ParseOwnershipTransferred(log types.Log) (*BunkerPricingOwnershipTransferred, error) {
	event := new(BunkerPricingOwnershipTransferred)
	if err := _BunkerPricing.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingPriceOracleUpdatedIterator is returned from FilterPriceOracleUpdated and is used to iterate over the raw logs and unpacked data for PriceOracleUpdated events raised by the BunkerPricing contract.
type BunkerPricingPriceOracleUpdatedIterator struct {
	Event *BunkerPricingPriceOracleUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingPriceOracleUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingPriceOracleUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingPriceOracleUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingPriceOracleUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingPriceOracleUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingPriceOracleUpdated represents a PriceOracleUpdated event raised by the BunkerPricing contract.
type BunkerPricingPriceOracleUpdated struct {
	OldOracle common.Address
	NewOracle common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPriceOracleUpdated is a free log retrieval operation binding the contract event 0x56b5f80d8cac1479698aa7d01605fd6111e90b15fc4d2b377417f46034876cbd.
//
// Solidity: event PriceOracleUpdated(address indexed oldOracle, address indexed newOracle)
func (_BunkerPricing *BunkerPricingFilterer) FilterPriceOracleUpdated(opts *bind.FilterOpts, oldOracle []common.Address, newOracle []common.Address) (*BunkerPricingPriceOracleUpdatedIterator, error) {

	var oldOracleRule []interface{}
	for _, oldOracleItem := range oldOracle {
		oldOracleRule = append(oldOracleRule, oldOracleItem)
	}
	var newOracleRule []interface{}
	for _, newOracleItem := range newOracle {
		newOracleRule = append(newOracleRule, newOracleItem)
	}

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "PriceOracleUpdated", oldOracleRule, newOracleRule)
	if err != nil {
		return nil, err
	}
	return &BunkerPricingPriceOracleUpdatedIterator{contract: _BunkerPricing.contract, event: "PriceOracleUpdated", logs: logs, sub: sub}, nil
}

// WatchPriceOracleUpdated is a free log subscription operation binding the contract event 0x56b5f80d8cac1479698aa7d01605fd6111e90b15fc4d2b377417f46034876cbd.
//
// Solidity: event PriceOracleUpdated(address indexed oldOracle, address indexed newOracle)
func (_BunkerPricing *BunkerPricingFilterer) WatchPriceOracleUpdated(opts *bind.WatchOpts, sink chan<- *BunkerPricingPriceOracleUpdated, oldOracle []common.Address, newOracle []common.Address) (event.Subscription, error) {

	var oldOracleRule []interface{}
	for _, oldOracleItem := range oldOracle {
		oldOracleRule = append(oldOracleRule, oldOracleItem)
	}
	var newOracleRule []interface{}
	for _, newOracleItem := range newOracle {
		newOracleRule = append(newOracleRule, newOracleItem)
	}

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "PriceOracleUpdated", oldOracleRule, newOracleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingPriceOracleUpdated)
				if err := _BunkerPricing.contract.UnpackLog(event, "PriceOracleUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePriceOracleUpdated is a log parse operation binding the contract event 0x56b5f80d8cac1479698aa7d01605fd6111e90b15fc4d2b377417f46034876cbd.
//
// Solidity: event PriceOracleUpdated(address indexed oldOracle, address indexed newOracle)
func (_BunkerPricing *BunkerPricingFilterer) ParsePriceOracleUpdated(log types.Log) (*BunkerPricingPriceOracleUpdated, error) {
	event := new(BunkerPricingPriceOracleUpdated)
	if err := _BunkerPricing.contract.UnpackLog(event, "PriceOracleUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingPricesUpdatedIterator is returned from FilterPricesUpdated and is used to iterate over the raw logs and unpacked data for PricesUpdated events raised by the BunkerPricing contract.
type BunkerPricingPricesUpdatedIterator struct {
	Event *BunkerPricingPricesUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingPricesUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingPricesUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingPricesUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingPricesUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingPricesUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingPricesUpdated represents a PricesUpdated event raised by the BunkerPricing contract.
type BunkerPricingPricesUpdated struct {
	CpuPerCoreHour    *big.Int
	MemoryPerGBHour   *big.Int
	StoragePerGBMonth *big.Int
	NetworkPerGB      *big.Int
	GpuBasicPerHour   *big.Int
	GpuPremiumPerHour *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterPricesUpdated is a free log retrieval operation binding the contract event 0xc73ec39986e464014d46dae8bdefb70d710adbcc1c342b49333817b40f4ab525.
//
// Solidity: event PricesUpdated(uint256 cpuPerCoreHour, uint256 memoryPerGBHour, uint256 storagePerGBMonth, uint256 networkPerGB, uint256 gpuBasicPerHour, uint256 gpuPremiumPerHour)
func (_BunkerPricing *BunkerPricingFilterer) FilterPricesUpdated(opts *bind.FilterOpts) (*BunkerPricingPricesUpdatedIterator, error) {

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "PricesUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerPricingPricesUpdatedIterator{contract: _BunkerPricing.contract, event: "PricesUpdated", logs: logs, sub: sub}, nil
}

// WatchPricesUpdated is a free log subscription operation binding the contract event 0xc73ec39986e464014d46dae8bdefb70d710adbcc1c342b49333817b40f4ab525.
//
// Solidity: event PricesUpdated(uint256 cpuPerCoreHour, uint256 memoryPerGBHour, uint256 storagePerGBMonth, uint256 networkPerGB, uint256 gpuBasicPerHour, uint256 gpuPremiumPerHour)
func (_BunkerPricing *BunkerPricingFilterer) WatchPricesUpdated(opts *bind.WatchOpts, sink chan<- *BunkerPricingPricesUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "PricesUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingPricesUpdated)
				if err := _BunkerPricing.contract.UnpackLog(event, "PricesUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePricesUpdated is a log parse operation binding the contract event 0xc73ec39986e464014d46dae8bdefb70d710adbcc1c342b49333817b40f4ab525.
//
// Solidity: event PricesUpdated(uint256 cpuPerCoreHour, uint256 memoryPerGBHour, uint256 storagePerGBMonth, uint256 networkPerGB, uint256 gpuBasicPerHour, uint256 gpuPremiumPerHour)
func (_BunkerPricing *BunkerPricingFilterer) ParsePricesUpdated(log types.Log) (*BunkerPricingPricesUpdated, error) {
	event := new(BunkerPricingPricesUpdated)
	if err := _BunkerPricing.contract.UnpackLog(event, "PricesUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingProviderPriceBoundsUpdatedIterator is returned from FilterProviderPriceBoundsUpdated and is used to iterate over the raw logs and unpacked data for ProviderPriceBoundsUpdated events raised by the BunkerPricing contract.
type BunkerPricingProviderPriceBoundsUpdatedIterator struct {
	Event *BunkerPricingProviderPriceBoundsUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingProviderPriceBoundsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingProviderPriceBoundsUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingProviderPriceBoundsUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingProviderPriceBoundsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingProviderPriceBoundsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingProviderPriceBoundsUpdated represents a ProviderPriceBoundsUpdated event raised by the BunkerPricing contract.
type BunkerPricingProviderPriceBoundsUpdated struct {
	MinBps *big.Int
	MaxBps *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterProviderPriceBoundsUpdated is a free log retrieval operation binding the contract event 0x6d6efa34d855535fa7f3ce1ee66a59c789265ead9418832223c17979816badb8.
//
// Solidity: event ProviderPriceBoundsUpdated(uint256 minBps, uint256 maxBps)
func (_BunkerPricing *BunkerPricingFilterer) FilterProviderPriceBoundsUpdated(opts *bind.FilterOpts) (*BunkerPricingProviderPriceBoundsUpdatedIterator, error) {

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "ProviderPriceBoundsUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerPricingProviderPriceBoundsUpdatedIterator{contract: _BunkerPricing.contract, event: "ProviderPriceBoundsUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderPriceBoundsUpdated is a free log subscription operation binding the contract event 0x6d6efa34d855535fa7f3ce1ee66a59c789265ead9418832223c17979816badb8.
//
// Solidity: event ProviderPriceBoundsUpdated(uint256 minBps, uint256 maxBps)
func (_BunkerPricing *BunkerPricingFilterer) WatchProviderPriceBoundsUpdated(opts *bind.WatchOpts, sink chan<- *BunkerPricingProviderPriceBoundsUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "ProviderPriceBoundsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingProviderPriceBoundsUpdated)
				if err := _BunkerPricing.contract.UnpackLog(event, "ProviderPriceBoundsUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderPriceBoundsUpdated is a log parse operation binding the contract event 0x6d6efa34d855535fa7f3ce1ee66a59c789265ead9418832223c17979816badb8.
//
// Solidity: event ProviderPriceBoundsUpdated(uint256 minBps, uint256 maxBps)
func (_BunkerPricing *BunkerPricingFilterer) ParseProviderPriceBoundsUpdated(log types.Log) (*BunkerPricingProviderPriceBoundsUpdated, error) {
	event := new(BunkerPricingProviderPriceBoundsUpdated)
	if err := _BunkerPricing.contract.UnpackLog(event, "ProviderPriceBoundsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingProviderPricesClearedIterator is returned from FilterProviderPricesCleared and is used to iterate over the raw logs and unpacked data for ProviderPricesCleared events raised by the BunkerPricing contract.
type BunkerPricingProviderPricesClearedIterator struct {
	Event *BunkerPricingProviderPricesCleared // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingProviderPricesClearedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingProviderPricesCleared)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingProviderPricesCleared)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingProviderPricesClearedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingProviderPricesClearedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingProviderPricesCleared represents a ProviderPricesCleared event raised by the BunkerPricing contract.
type BunkerPricingProviderPricesCleared struct {
	Provider common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderPricesCleared is a free log retrieval operation binding the contract event 0xe71ed8b6608cadd98920dbe666b4ed5a79bbe13107ca2685bda11eabc73aef06.
//
// Solidity: event ProviderPricesCleared(address indexed provider)
func (_BunkerPricing *BunkerPricingFilterer) FilterProviderPricesCleared(opts *bind.FilterOpts, provider []common.Address) (*BunkerPricingProviderPricesClearedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "ProviderPricesCleared", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerPricingProviderPricesClearedIterator{contract: _BunkerPricing.contract, event: "ProviderPricesCleared", logs: logs, sub: sub}, nil
}

// WatchProviderPricesCleared is a free log subscription operation binding the contract event 0xe71ed8b6608cadd98920dbe666b4ed5a79bbe13107ca2685bda11eabc73aef06.
//
// Solidity: event ProviderPricesCleared(address indexed provider)
func (_BunkerPricing *BunkerPricingFilterer) WatchProviderPricesCleared(opts *bind.WatchOpts, sink chan<- *BunkerPricingProviderPricesCleared, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "ProviderPricesCleared", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingProviderPricesCleared)
				if err := _BunkerPricing.contract.UnpackLog(event, "ProviderPricesCleared", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderPricesCleared is a log parse operation binding the contract event 0xe71ed8b6608cadd98920dbe666b4ed5a79bbe13107ca2685bda11eabc73aef06.
//
// Solidity: event ProviderPricesCleared(address indexed provider)
func (_BunkerPricing *BunkerPricingFilterer) ParseProviderPricesCleared(log types.Log) (*BunkerPricingProviderPricesCleared, error) {
	event := new(BunkerPricingProviderPricesCleared)
	if err := _BunkerPricing.contract.UnpackLog(event, "ProviderPricesCleared", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingProviderPricesSetIterator is returned from FilterProviderPricesSet and is used to iterate over the raw logs and unpacked data for ProviderPricesSet events raised by the BunkerPricing contract.
type BunkerPricingProviderPricesSetIterator struct {
	Event *BunkerPricingProviderPricesSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingProviderPricesSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingProviderPricesSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingProviderPricesSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingProviderPricesSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingProviderPricesSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingProviderPricesSet represents a ProviderPricesSet event raised by the BunkerPricing contract.
type BunkerPricingProviderPricesSet struct {
	Provider common.Address
	Cpu      *big.Int
	Memory   *big.Int
	Storage  *big.Int
	Network  *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderPricesSet is a free log retrieval operation binding the contract event 0xa1dab754312a0c4e3f91ed4dc76f7d8467b9e0405c1531fc59049da78aabef80.
//
// Solidity: event ProviderPricesSet(address indexed provider, uint256 cpu, uint256 memory_, uint256 storage_, uint256 network)
func (_BunkerPricing *BunkerPricingFilterer) FilterProviderPricesSet(opts *bind.FilterOpts, provider []common.Address) (*BunkerPricingProviderPricesSetIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "ProviderPricesSet", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerPricingProviderPricesSetIterator{contract: _BunkerPricing.contract, event: "ProviderPricesSet", logs: logs, sub: sub}, nil
}

// WatchProviderPricesSet is a free log subscription operation binding the contract event 0xa1dab754312a0c4e3f91ed4dc76f7d8467b9e0405c1531fc59049da78aabef80.
//
// Solidity: event ProviderPricesSet(address indexed provider, uint256 cpu, uint256 memory_, uint256 storage_, uint256 network)
func (_BunkerPricing *BunkerPricingFilterer) WatchProviderPricesSet(opts *bind.WatchOpts, sink chan<- *BunkerPricingProviderPricesSet, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "ProviderPricesSet", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingProviderPricesSet)
				if err := _BunkerPricing.contract.UnpackLog(event, "ProviderPricesSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderPricesSet is a log parse operation binding the contract event 0xa1dab754312a0c4e3f91ed4dc76f7d8467b9e0405c1531fc59049da78aabef80.
//
// Solidity: event ProviderPricesSet(address indexed provider, uint256 cpu, uint256 memory_, uint256 storage_, uint256 network)
func (_BunkerPricing *BunkerPricingFilterer) ParseProviderPricesSet(log types.Log) (*BunkerPricingProviderPricesSet, error) {
	event := new(BunkerPricingProviderPricesSet)
	if err := _BunkerPricing.contract.UnpackLog(event, "ProviderPricesSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingStalePriceThresholdUpdatedIterator is returned from FilterStalePriceThresholdUpdated and is used to iterate over the raw logs and unpacked data for StalePriceThresholdUpdated events raised by the BunkerPricing contract.
type BunkerPricingStalePriceThresholdUpdatedIterator struct {
	Event *BunkerPricingStalePriceThresholdUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingStalePriceThresholdUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingStalePriceThresholdUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingStalePriceThresholdUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingStalePriceThresholdUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingStalePriceThresholdUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingStalePriceThresholdUpdated represents a StalePriceThresholdUpdated event raised by the BunkerPricing contract.
type BunkerPricingStalePriceThresholdUpdated struct {
	OldThreshold *big.Int
	NewThreshold *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterStalePriceThresholdUpdated is a free log retrieval operation binding the contract event 0x8f83b069a278bb20c05e2fd368a9a7a7bca85964a493d16acea1bf1864df3920.
//
// Solidity: event StalePriceThresholdUpdated(uint256 oldThreshold, uint256 newThreshold)
func (_BunkerPricing *BunkerPricingFilterer) FilterStalePriceThresholdUpdated(opts *bind.FilterOpts) (*BunkerPricingStalePriceThresholdUpdatedIterator, error) {

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "StalePriceThresholdUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerPricingStalePriceThresholdUpdatedIterator{contract: _BunkerPricing.contract, event: "StalePriceThresholdUpdated", logs: logs, sub: sub}, nil
}

// WatchStalePriceThresholdUpdated is a free log subscription operation binding the contract event 0x8f83b069a278bb20c05e2fd368a9a7a7bca85964a493d16acea1bf1864df3920.
//
// Solidity: event StalePriceThresholdUpdated(uint256 oldThreshold, uint256 newThreshold)
func (_BunkerPricing *BunkerPricingFilterer) WatchStalePriceThresholdUpdated(opts *bind.WatchOpts, sink chan<- *BunkerPricingStalePriceThresholdUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "StalePriceThresholdUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingStalePriceThresholdUpdated)
				if err := _BunkerPricing.contract.UnpackLog(event, "StalePriceThresholdUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStalePriceThresholdUpdated is a log parse operation binding the contract event 0x8f83b069a278bb20c05e2fd368a9a7a7bca85964a493d16acea1bf1864df3920.
//
// Solidity: event StalePriceThresholdUpdated(uint256 oldThreshold, uint256 newThreshold)
func (_BunkerPricing *BunkerPricingFilterer) ParseStalePriceThresholdUpdated(log types.Log) (*BunkerPricingStalePriceThresholdUpdated, error) {
	event := new(BunkerPricingStalePriceThresholdUpdated)
	if err := _BunkerPricing.contract.UnpackLog(event, "StalePriceThresholdUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerPricingStaleThresholdBoundsUpdatedIterator is returned from FilterStaleThresholdBoundsUpdated and is used to iterate over the raw logs and unpacked data for StaleThresholdBoundsUpdated events raised by the BunkerPricing contract.
type BunkerPricingStaleThresholdBoundsUpdatedIterator struct {
	Event *BunkerPricingStaleThresholdBoundsUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *BunkerPricingStaleThresholdBoundsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerPricingStaleThresholdBoundsUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(BunkerPricingStaleThresholdBoundsUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *BunkerPricingStaleThresholdBoundsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerPricingStaleThresholdBoundsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerPricingStaleThresholdBoundsUpdated represents a StaleThresholdBoundsUpdated event raised by the BunkerPricing contract.
type BunkerPricingStaleThresholdBoundsUpdated struct {
	NewMin *big.Int
	NewMax *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterStaleThresholdBoundsUpdated is a free log retrieval operation binding the contract event 0xd38397e3dccaeef2b961fb2d310645505e2c8df40fbbbcea679d2689b79698e1.
//
// Solidity: event StaleThresholdBoundsUpdated(uint256 newMin, uint256 newMax)
func (_BunkerPricing *BunkerPricingFilterer) FilterStaleThresholdBoundsUpdated(opts *bind.FilterOpts) (*BunkerPricingStaleThresholdBoundsUpdatedIterator, error) {

	logs, sub, err := _BunkerPricing.contract.FilterLogs(opts, "StaleThresholdBoundsUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerPricingStaleThresholdBoundsUpdatedIterator{contract: _BunkerPricing.contract, event: "StaleThresholdBoundsUpdated", logs: logs, sub: sub}, nil
}

// WatchStaleThresholdBoundsUpdated is a free log subscription operation binding the contract event 0xd38397e3dccaeef2b961fb2d310645505e2c8df40fbbbcea679d2689b79698e1.
//
// Solidity: event StaleThresholdBoundsUpdated(uint256 newMin, uint256 newMax)
func (_BunkerPricing *BunkerPricingFilterer) WatchStaleThresholdBoundsUpdated(opts *bind.WatchOpts, sink chan<- *BunkerPricingStaleThresholdBoundsUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerPricing.contract.WatchLogs(opts, "StaleThresholdBoundsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerPricingStaleThresholdBoundsUpdated)
				if err := _BunkerPricing.contract.UnpackLog(event, "StaleThresholdBoundsUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStaleThresholdBoundsUpdated is a log parse operation binding the contract event 0xd38397e3dccaeef2b961fb2d310645505e2c8df40fbbbcea679d2689b79698e1.
//
// Solidity: event StaleThresholdBoundsUpdated(uint256 newMin, uint256 newMax)
func (_BunkerPricing *BunkerPricingFilterer) ParseStaleThresholdBoundsUpdated(log types.Log) (*BunkerPricingStaleThresholdBoundsUpdated, error) {
	event := new(BunkerPricingStaleThresholdBoundsUpdated)
	if err := _BunkerPricing.contract.UnpackLog(event, "StaleThresholdBoundsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
