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

// BunkerDelegationDelegationInfo is an auto generated low-level Go binding around an user-defined struct.
type BunkerDelegationDelegationInfo struct {
	Provider    common.Address
	Amount      *big.Int
	DelegatedAt *big.Int
	Active      bool
}

// BunkerDelegationProviderDelegationConfig is an auto generated low-level Go binding around an user-defined struct.
type BunkerDelegationProviderDelegationConfig struct {
	RewardCutBps         uint16
	PendingRewardCutBps  uint16
	RewardCutEffectiveAt *big.Int
	FeeShareBps          uint16
	TotalDelegated       *big.Int
	AcceptingDelegations bool
}

// BunkerDelegationUnbondingRequest is an auto generated low-level Go binding around an user-defined struct.
type BunkerDelegationUnbondingRequest struct {
	Amount     *big.Int
	UnlockTime *big.Int
	Completed  bool
}

// BunkerDelegationMetaData contains all meta data concerning the BunkerDelegation contract.
var BunkerDelegationMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_stakingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"BPS_DENOMINATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_UNBONDING_QUEUE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"completeUndelegate\",\"inputs\":[{\"name\":\"requestIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"delegate\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"delegations\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"delegatedAt\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"active\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"finalizeRewardCut\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getDelegation\",\"inputs\":[{\"name\":\"delegator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"info\",\"type\":\"tuple\",\"internalType\":\"structBunkerDelegation.DelegationInfo\",\"components\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"delegatedAt\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"active\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getProviderConfig\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"config\",\"type\":\"tuple\",\"internalType\":\"structBunkerDelegation.ProviderDelegationConfig\",\"components\":[{\"name\":\"rewardCutBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"pendingRewardCutBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"rewardCutEffectiveAt\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"feeShareBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"totalDelegated\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"acceptingDelegations\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTotalDelegatedTo\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"total\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUnbondingQueueLength\",\"inputs\":[{\"name\":\"delegator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUnbondingRequest\",\"inputs\":[{\"name\":\"delegator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"request\",\"type\":\"tuple\",\"internalType\":\"structBunkerDelegation.UnbondingRequest\",\"components\":[{\"name\":\"amount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"unlockTime\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"completed\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxRewardCutBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"paused\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"providerConfigs\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"rewardCutBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"pendingRewardCutBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"rewardCutEffectiveAt\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"feeShareBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"totalDelegated\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"acceptingDelegations\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"requestUndelegate\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setDelegationConfig\",\"inputs\":[{\"name\":\"rewardCutBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"feeShareBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxRewardCutBps\",\"inputs\":[{\"name\":\"newMax\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setStakingContract\",\"inputs\":[{\"name\":\"_stakingContract\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setUnbondingPeriod\",\"inputs\":[{\"name\":\"newPeriod\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"stakingContract\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"toggleAcceptDelegations\",\"inputs\":[{\"name\":\"accepting\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"token\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalDelegated\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unbondingPeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"unbondingQueues\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"amount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"unlockTime\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"completed\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"unpause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"Delegated\",\"inputs\":[{\"name\":\"delegator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"DelegationAcceptanceToggled\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"accepting\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"DelegationWithdrawn\",\"inputs\":[{\"name\":\"delegator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MaxRewardCutUpdated\",\"inputs\":[{\"name\":\"newMax\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Paused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderConfigUpdated\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"rewardCutBps\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"feeShareBps\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RewardCutFinalized\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"rewardCutBps\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RewardCutIncreaseScheduled\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newRewardCutBps\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"effectiveAt\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"StakingContractUpdated\",\"inputs\":[{\"name\":\"oldStaking\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newStaking\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UnbondingPeriodUpdated\",\"inputs\":[{\"name\":\"newPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UndelegateCompleted\",\"inputs\":[{\"name\":\"delegator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UndelegateRequested\",\"inputs\":[{\"name\":\"delegator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"unlockTime\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Unpaused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AlreadyDelegated\",\"inputs\":[{\"name\":\"delegator\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"EnforcedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExpectedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientDelegation\",\"inputs\":[{\"name\":\"requested\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"available\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidRewardCutCap\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidUnbondingIndex\",\"inputs\":[{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"queueLength\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidUnbondingPeriod\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NoPendingRewardCut\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotDelegated\",\"inputs\":[{\"name\":\"delegator\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ProviderNotAccepting\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ProviderNotActive\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"RewardCutTimelockActive\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"RewardCutTooHigh\",\"inputs\":[{\"name\":\"requested\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"maximum\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"TooManyUnbondingRequests\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UnbondingAlreadyCompleted\",\"inputs\":[{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"UnbondingNotReady\",\"inputs\":[{\"name\":\"unlockTime\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"currentTime\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ZeroAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroAmount\",\"inputs\":[]}]",
}

// BunkerDelegationABI is the input ABI used to generate the binding from.
// Deprecated: Use BunkerDelegationMetaData.ABI instead.
var BunkerDelegationABI = BunkerDelegationMetaData.ABI

// BunkerDelegation is an auto generated Go binding around an Ethereum contract.
type BunkerDelegation struct {
	BunkerDelegationCaller     // Read-only binding to the contract
	BunkerDelegationTransactor // Write-only binding to the contract
	BunkerDelegationFilterer   // Log filterer for contract events
}

// BunkerDelegationCaller is an auto generated read-only Go binding around an Ethereum contract.
type BunkerDelegationCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerDelegationTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BunkerDelegationTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerDelegationFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BunkerDelegationFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerDelegationSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BunkerDelegationSession struct {
	Contract     *BunkerDelegation // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BunkerDelegationCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BunkerDelegationCallerSession struct {
	Contract *BunkerDelegationCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// BunkerDelegationTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BunkerDelegationTransactorSession struct {
	Contract     *BunkerDelegationTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// BunkerDelegationRaw is an auto generated low-level Go binding around an Ethereum contract.
type BunkerDelegationRaw struct {
	Contract *BunkerDelegation // Generic contract binding to access the raw methods on
}

// BunkerDelegationCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BunkerDelegationCallerRaw struct {
	Contract *BunkerDelegationCaller // Generic read-only contract binding to access the raw methods on
}

// BunkerDelegationTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BunkerDelegationTransactorRaw struct {
	Contract *BunkerDelegationTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBunkerDelegation creates a new instance of BunkerDelegation, bound to a specific deployed contract.
func NewBunkerDelegation(address common.Address, backend bind.ContractBackend) (*BunkerDelegation, error) {
	contract, err := bindBunkerDelegation(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegation{BunkerDelegationCaller: BunkerDelegationCaller{contract: contract}, BunkerDelegationTransactor: BunkerDelegationTransactor{contract: contract}, BunkerDelegationFilterer: BunkerDelegationFilterer{contract: contract}}, nil
}

// NewBunkerDelegationCaller creates a new read-only instance of BunkerDelegation, bound to a specific deployed contract.
func NewBunkerDelegationCaller(address common.Address, caller bind.ContractCaller) (*BunkerDelegationCaller, error) {
	contract, err := bindBunkerDelegation(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationCaller{contract: contract}, nil
}

// NewBunkerDelegationTransactor creates a new write-only instance of BunkerDelegation, bound to a specific deployed contract.
func NewBunkerDelegationTransactor(address common.Address, transactor bind.ContractTransactor) (*BunkerDelegationTransactor, error) {
	contract, err := bindBunkerDelegation(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationTransactor{contract: contract}, nil
}

// NewBunkerDelegationFilterer creates a new log filterer instance of BunkerDelegation, bound to a specific deployed contract.
func NewBunkerDelegationFilterer(address common.Address, filterer bind.ContractFilterer) (*BunkerDelegationFilterer, error) {
	contract, err := bindBunkerDelegation(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationFilterer{contract: contract}, nil
}

// bindBunkerDelegation binds a generic wrapper to an already deployed contract.
func bindBunkerDelegation(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BunkerDelegationMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerDelegation *BunkerDelegationRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerDelegation.Contract.BunkerDelegationCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerDelegation *BunkerDelegationRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.BunkerDelegationTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerDelegation *BunkerDelegationRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.BunkerDelegationTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerDelegation *BunkerDelegationCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerDelegation.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerDelegation *BunkerDelegationTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerDelegation *BunkerDelegationTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.contract.Transact(opts, method, params...)
}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationCaller) BPSDENOMINATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "BPS_DENOMINATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationSession) BPSDENOMINATOR() (*big.Int, error) {
	return _BunkerDelegation.Contract.BPSDENOMINATOR(&_BunkerDelegation.CallOpts)
}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationCallerSession) BPSDENOMINATOR() (*big.Int, error) {
	return _BunkerDelegation.Contract.BPSDENOMINATOR(&_BunkerDelegation.CallOpts)
}

// MAXUNBONDINGQUEUE is a free data retrieval call binding the contract method 0x5a14a6ef.
//
// Solidity: function MAX_UNBONDING_QUEUE() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationCaller) MAXUNBONDINGQUEUE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "MAX_UNBONDING_QUEUE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXUNBONDINGQUEUE is a free data retrieval call binding the contract method 0x5a14a6ef.
//
// Solidity: function MAX_UNBONDING_QUEUE() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationSession) MAXUNBONDINGQUEUE() (*big.Int, error) {
	return _BunkerDelegation.Contract.MAXUNBONDINGQUEUE(&_BunkerDelegation.CallOpts)
}

// MAXUNBONDINGQUEUE is a free data retrieval call binding the contract method 0x5a14a6ef.
//
// Solidity: function MAX_UNBONDING_QUEUE() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationCallerSession) MAXUNBONDINGQUEUE() (*big.Int, error) {
	return _BunkerDelegation.Contract.MAXUNBONDINGQUEUE(&_BunkerDelegation.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerDelegation *BunkerDelegationCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerDelegation *BunkerDelegationSession) VERSION() (string, error) {
	return _BunkerDelegation.Contract.VERSION(&_BunkerDelegation.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerDelegation *BunkerDelegationCallerSession) VERSION() (string, error) {
	return _BunkerDelegation.Contract.VERSION(&_BunkerDelegation.CallOpts)
}

// Delegations is a free data retrieval call binding the contract method 0xbffe3486.
//
// Solidity: function delegations(address ) view returns(address provider, uint128 amount, uint48 delegatedAt, bool active)
func (_BunkerDelegation *BunkerDelegationCaller) Delegations(opts *bind.CallOpts, arg0 common.Address) (struct {
	Provider    common.Address
	Amount      *big.Int
	DelegatedAt *big.Int
	Active      bool
}, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "delegations", arg0)

	outstruct := new(struct {
		Provider    common.Address
		Amount      *big.Int
		DelegatedAt *big.Int
		Active      bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Provider = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Amount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.DelegatedAt = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Active = *abi.ConvertType(out[3], new(bool)).(*bool)

	return *outstruct, err

}

// Delegations is a free data retrieval call binding the contract method 0xbffe3486.
//
// Solidity: function delegations(address ) view returns(address provider, uint128 amount, uint48 delegatedAt, bool active)
func (_BunkerDelegation *BunkerDelegationSession) Delegations(arg0 common.Address) (struct {
	Provider    common.Address
	Amount      *big.Int
	DelegatedAt *big.Int
	Active      bool
}, error) {
	return _BunkerDelegation.Contract.Delegations(&_BunkerDelegation.CallOpts, arg0)
}

// Delegations is a free data retrieval call binding the contract method 0xbffe3486.
//
// Solidity: function delegations(address ) view returns(address provider, uint128 amount, uint48 delegatedAt, bool active)
func (_BunkerDelegation *BunkerDelegationCallerSession) Delegations(arg0 common.Address) (struct {
	Provider    common.Address
	Amount      *big.Int
	DelegatedAt *big.Int
	Active      bool
}, error) {
	return _BunkerDelegation.Contract.Delegations(&_BunkerDelegation.CallOpts, arg0)
}

// GetDelegation is a free data retrieval call binding the contract method 0x2b293768.
//
// Solidity: function getDelegation(address delegator) view returns((address,uint128,uint48,bool) info)
func (_BunkerDelegation *BunkerDelegationCaller) GetDelegation(opts *bind.CallOpts, delegator common.Address) (BunkerDelegationDelegationInfo, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "getDelegation", delegator)

	if err != nil {
		return *new(BunkerDelegationDelegationInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerDelegationDelegationInfo)).(*BunkerDelegationDelegationInfo)

	return out0, err

}

// GetDelegation is a free data retrieval call binding the contract method 0x2b293768.
//
// Solidity: function getDelegation(address delegator) view returns((address,uint128,uint48,bool) info)
func (_BunkerDelegation *BunkerDelegationSession) GetDelegation(delegator common.Address) (BunkerDelegationDelegationInfo, error) {
	return _BunkerDelegation.Contract.GetDelegation(&_BunkerDelegation.CallOpts, delegator)
}

// GetDelegation is a free data retrieval call binding the contract method 0x2b293768.
//
// Solidity: function getDelegation(address delegator) view returns((address,uint128,uint48,bool) info)
func (_BunkerDelegation *BunkerDelegationCallerSession) GetDelegation(delegator common.Address) (BunkerDelegationDelegationInfo, error) {
	return _BunkerDelegation.Contract.GetDelegation(&_BunkerDelegation.CallOpts, delegator)
}

// GetProviderConfig is a free data retrieval call binding the contract method 0x0edcd564.
//
// Solidity: function getProviderConfig(address provider) view returns((uint16,uint16,uint48,uint16,uint128,bool) config)
func (_BunkerDelegation *BunkerDelegationCaller) GetProviderConfig(opts *bind.CallOpts, provider common.Address) (BunkerDelegationProviderDelegationConfig, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "getProviderConfig", provider)

	if err != nil {
		return *new(BunkerDelegationProviderDelegationConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerDelegationProviderDelegationConfig)).(*BunkerDelegationProviderDelegationConfig)

	return out0, err

}

// GetProviderConfig is a free data retrieval call binding the contract method 0x0edcd564.
//
// Solidity: function getProviderConfig(address provider) view returns((uint16,uint16,uint48,uint16,uint128,bool) config)
func (_BunkerDelegation *BunkerDelegationSession) GetProviderConfig(provider common.Address) (BunkerDelegationProviderDelegationConfig, error) {
	return _BunkerDelegation.Contract.GetProviderConfig(&_BunkerDelegation.CallOpts, provider)
}

// GetProviderConfig is a free data retrieval call binding the contract method 0x0edcd564.
//
// Solidity: function getProviderConfig(address provider) view returns((uint16,uint16,uint48,uint16,uint128,bool) config)
func (_BunkerDelegation *BunkerDelegationCallerSession) GetProviderConfig(provider common.Address) (BunkerDelegationProviderDelegationConfig, error) {
	return _BunkerDelegation.Contract.GetProviderConfig(&_BunkerDelegation.CallOpts, provider)
}

// GetTotalDelegatedTo is a free data retrieval call binding the contract method 0xd2032d71.
//
// Solidity: function getTotalDelegatedTo(address provider) view returns(uint256 total)
func (_BunkerDelegation *BunkerDelegationCaller) GetTotalDelegatedTo(opts *bind.CallOpts, provider common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "getTotalDelegatedTo", provider)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalDelegatedTo is a free data retrieval call binding the contract method 0xd2032d71.
//
// Solidity: function getTotalDelegatedTo(address provider) view returns(uint256 total)
func (_BunkerDelegation *BunkerDelegationSession) GetTotalDelegatedTo(provider common.Address) (*big.Int, error) {
	return _BunkerDelegation.Contract.GetTotalDelegatedTo(&_BunkerDelegation.CallOpts, provider)
}

// GetTotalDelegatedTo is a free data retrieval call binding the contract method 0xd2032d71.
//
// Solidity: function getTotalDelegatedTo(address provider) view returns(uint256 total)
func (_BunkerDelegation *BunkerDelegationCallerSession) GetTotalDelegatedTo(provider common.Address) (*big.Int, error) {
	return _BunkerDelegation.Contract.GetTotalDelegatedTo(&_BunkerDelegation.CallOpts, provider)
}

// GetUnbondingQueueLength is a free data retrieval call binding the contract method 0xdcb67d53.
//
// Solidity: function getUnbondingQueueLength(address delegator) view returns(uint256 count)
func (_BunkerDelegation *BunkerDelegationCaller) GetUnbondingQueueLength(opts *bind.CallOpts, delegator common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "getUnbondingQueueLength", delegator)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUnbondingQueueLength is a free data retrieval call binding the contract method 0xdcb67d53.
//
// Solidity: function getUnbondingQueueLength(address delegator) view returns(uint256 count)
func (_BunkerDelegation *BunkerDelegationSession) GetUnbondingQueueLength(delegator common.Address) (*big.Int, error) {
	return _BunkerDelegation.Contract.GetUnbondingQueueLength(&_BunkerDelegation.CallOpts, delegator)
}

// GetUnbondingQueueLength is a free data retrieval call binding the contract method 0xdcb67d53.
//
// Solidity: function getUnbondingQueueLength(address delegator) view returns(uint256 count)
func (_BunkerDelegation *BunkerDelegationCallerSession) GetUnbondingQueueLength(delegator common.Address) (*big.Int, error) {
	return _BunkerDelegation.Contract.GetUnbondingQueueLength(&_BunkerDelegation.CallOpts, delegator)
}

// GetUnbondingRequest is a free data retrieval call binding the contract method 0x9e79b122.
//
// Solidity: function getUnbondingRequest(address delegator, uint256 index) view returns((uint128,uint48,bool) request)
func (_BunkerDelegation *BunkerDelegationCaller) GetUnbondingRequest(opts *bind.CallOpts, delegator common.Address, index *big.Int) (BunkerDelegationUnbondingRequest, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "getUnbondingRequest", delegator, index)

	if err != nil {
		return *new(BunkerDelegationUnbondingRequest), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerDelegationUnbondingRequest)).(*BunkerDelegationUnbondingRequest)

	return out0, err

}

// GetUnbondingRequest is a free data retrieval call binding the contract method 0x9e79b122.
//
// Solidity: function getUnbondingRequest(address delegator, uint256 index) view returns((uint128,uint48,bool) request)
func (_BunkerDelegation *BunkerDelegationSession) GetUnbondingRequest(delegator common.Address, index *big.Int) (BunkerDelegationUnbondingRequest, error) {
	return _BunkerDelegation.Contract.GetUnbondingRequest(&_BunkerDelegation.CallOpts, delegator, index)
}

// GetUnbondingRequest is a free data retrieval call binding the contract method 0x9e79b122.
//
// Solidity: function getUnbondingRequest(address delegator, uint256 index) view returns((uint128,uint48,bool) request)
func (_BunkerDelegation *BunkerDelegationCallerSession) GetUnbondingRequest(delegator common.Address, index *big.Int) (BunkerDelegationUnbondingRequest, error) {
	return _BunkerDelegation.Contract.GetUnbondingRequest(&_BunkerDelegation.CallOpts, delegator, index)
}

// MaxRewardCutBps is a free data retrieval call binding the contract method 0xbfb38478.
//
// Solidity: function maxRewardCutBps() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationCaller) MaxRewardCutBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "maxRewardCutBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxRewardCutBps is a free data retrieval call binding the contract method 0xbfb38478.
//
// Solidity: function maxRewardCutBps() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationSession) MaxRewardCutBps() (*big.Int, error) {
	return _BunkerDelegation.Contract.MaxRewardCutBps(&_BunkerDelegation.CallOpts)
}

// MaxRewardCutBps is a free data retrieval call binding the contract method 0xbfb38478.
//
// Solidity: function maxRewardCutBps() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationCallerSession) MaxRewardCutBps() (*big.Int, error) {
	return _BunkerDelegation.Contract.MaxRewardCutBps(&_BunkerDelegation.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerDelegation *BunkerDelegationCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerDelegation *BunkerDelegationSession) Owner() (common.Address, error) {
	return _BunkerDelegation.Contract.Owner(&_BunkerDelegation.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerDelegation *BunkerDelegationCallerSession) Owner() (common.Address, error) {
	return _BunkerDelegation.Contract.Owner(&_BunkerDelegation.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerDelegation *BunkerDelegationCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerDelegation *BunkerDelegationSession) Paused() (bool, error) {
	return _BunkerDelegation.Contract.Paused(&_BunkerDelegation.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerDelegation *BunkerDelegationCallerSession) Paused() (bool, error) {
	return _BunkerDelegation.Contract.Paused(&_BunkerDelegation.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerDelegation *BunkerDelegationCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerDelegation *BunkerDelegationSession) PendingOwner() (common.Address, error) {
	return _BunkerDelegation.Contract.PendingOwner(&_BunkerDelegation.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerDelegation *BunkerDelegationCallerSession) PendingOwner() (common.Address, error) {
	return _BunkerDelegation.Contract.PendingOwner(&_BunkerDelegation.CallOpts)
}

// ProviderConfigs is a free data retrieval call binding the contract method 0x96d509f2.
//
// Solidity: function providerConfigs(address ) view returns(uint16 rewardCutBps, uint16 pendingRewardCutBps, uint48 rewardCutEffectiveAt, uint16 feeShareBps, uint128 totalDelegated, bool acceptingDelegations)
func (_BunkerDelegation *BunkerDelegationCaller) ProviderConfigs(opts *bind.CallOpts, arg0 common.Address) (struct {
	RewardCutBps         uint16
	PendingRewardCutBps  uint16
	RewardCutEffectiveAt *big.Int
	FeeShareBps          uint16
	TotalDelegated       *big.Int
	AcceptingDelegations bool
}, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "providerConfigs", arg0)

	outstruct := new(struct {
		RewardCutBps         uint16
		PendingRewardCutBps  uint16
		RewardCutEffectiveAt *big.Int
		FeeShareBps          uint16
		TotalDelegated       *big.Int
		AcceptingDelegations bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.RewardCutBps = *abi.ConvertType(out[0], new(uint16)).(*uint16)
	outstruct.PendingRewardCutBps = *abi.ConvertType(out[1], new(uint16)).(*uint16)
	outstruct.RewardCutEffectiveAt = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.FeeShareBps = *abi.ConvertType(out[3], new(uint16)).(*uint16)
	outstruct.TotalDelegated = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.AcceptingDelegations = *abi.ConvertType(out[5], new(bool)).(*bool)

	return *outstruct, err

}

// ProviderConfigs is a free data retrieval call binding the contract method 0x96d509f2.
//
// Solidity: function providerConfigs(address ) view returns(uint16 rewardCutBps, uint16 pendingRewardCutBps, uint48 rewardCutEffectiveAt, uint16 feeShareBps, uint128 totalDelegated, bool acceptingDelegations)
func (_BunkerDelegation *BunkerDelegationSession) ProviderConfigs(arg0 common.Address) (struct {
	RewardCutBps         uint16
	PendingRewardCutBps  uint16
	RewardCutEffectiveAt *big.Int
	FeeShareBps          uint16
	TotalDelegated       *big.Int
	AcceptingDelegations bool
}, error) {
	return _BunkerDelegation.Contract.ProviderConfigs(&_BunkerDelegation.CallOpts, arg0)
}

// ProviderConfigs is a free data retrieval call binding the contract method 0x96d509f2.
//
// Solidity: function providerConfigs(address ) view returns(uint16 rewardCutBps, uint16 pendingRewardCutBps, uint48 rewardCutEffectiveAt, uint16 feeShareBps, uint128 totalDelegated, bool acceptingDelegations)
func (_BunkerDelegation *BunkerDelegationCallerSession) ProviderConfigs(arg0 common.Address) (struct {
	RewardCutBps         uint16
	PendingRewardCutBps  uint16
	RewardCutEffectiveAt *big.Int
	FeeShareBps          uint16
	TotalDelegated       *big.Int
	AcceptingDelegations bool
}, error) {
	return _BunkerDelegation.Contract.ProviderConfigs(&_BunkerDelegation.CallOpts, arg0)
}

// StakingContract is a free data retrieval call binding the contract method 0xee99205c.
//
// Solidity: function stakingContract() view returns(address)
func (_BunkerDelegation *BunkerDelegationCaller) StakingContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "stakingContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakingContract is a free data retrieval call binding the contract method 0xee99205c.
//
// Solidity: function stakingContract() view returns(address)
func (_BunkerDelegation *BunkerDelegationSession) StakingContract() (common.Address, error) {
	return _BunkerDelegation.Contract.StakingContract(&_BunkerDelegation.CallOpts)
}

// StakingContract is a free data retrieval call binding the contract method 0xee99205c.
//
// Solidity: function stakingContract() view returns(address)
func (_BunkerDelegation *BunkerDelegationCallerSession) StakingContract() (common.Address, error) {
	return _BunkerDelegation.Contract.StakingContract(&_BunkerDelegation.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_BunkerDelegation *BunkerDelegationCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_BunkerDelegation *BunkerDelegationSession) Token() (common.Address, error) {
	return _BunkerDelegation.Contract.Token(&_BunkerDelegation.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_BunkerDelegation *BunkerDelegationCallerSession) Token() (common.Address, error) {
	return _BunkerDelegation.Contract.Token(&_BunkerDelegation.CallOpts)
}

// TotalDelegated is a free data retrieval call binding the contract method 0x80d04de8.
//
// Solidity: function totalDelegated() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationCaller) TotalDelegated(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "totalDelegated")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalDelegated is a free data retrieval call binding the contract method 0x80d04de8.
//
// Solidity: function totalDelegated() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationSession) TotalDelegated() (*big.Int, error) {
	return _BunkerDelegation.Contract.TotalDelegated(&_BunkerDelegation.CallOpts)
}

// TotalDelegated is a free data retrieval call binding the contract method 0x80d04de8.
//
// Solidity: function totalDelegated() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationCallerSession) TotalDelegated() (*big.Int, error) {
	return _BunkerDelegation.Contract.TotalDelegated(&_BunkerDelegation.CallOpts)
}

// UnbondingPeriod is a free data retrieval call binding the contract method 0x6cf6d675.
//
// Solidity: function unbondingPeriod() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationCaller) UnbondingPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "unbondingPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UnbondingPeriod is a free data retrieval call binding the contract method 0x6cf6d675.
//
// Solidity: function unbondingPeriod() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationSession) UnbondingPeriod() (*big.Int, error) {
	return _BunkerDelegation.Contract.UnbondingPeriod(&_BunkerDelegation.CallOpts)
}

// UnbondingPeriod is a free data retrieval call binding the contract method 0x6cf6d675.
//
// Solidity: function unbondingPeriod() view returns(uint256)
func (_BunkerDelegation *BunkerDelegationCallerSession) UnbondingPeriod() (*big.Int, error) {
	return _BunkerDelegation.Contract.UnbondingPeriod(&_BunkerDelegation.CallOpts)
}

// UnbondingQueues is a free data retrieval call binding the contract method 0x28a78f43.
//
// Solidity: function unbondingQueues(address , uint256 ) view returns(uint128 amount, uint48 unlockTime, bool completed)
func (_BunkerDelegation *BunkerDelegationCaller) UnbondingQueues(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	Amount     *big.Int
	UnlockTime *big.Int
	Completed  bool
}, error) {
	var out []interface{}
	err := _BunkerDelegation.contract.Call(opts, &out, "unbondingQueues", arg0, arg1)

	outstruct := new(struct {
		Amount     *big.Int
		UnlockTime *big.Int
		Completed  bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Amount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.UnlockTime = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Completed = *abi.ConvertType(out[2], new(bool)).(*bool)

	return *outstruct, err

}

// UnbondingQueues is a free data retrieval call binding the contract method 0x28a78f43.
//
// Solidity: function unbondingQueues(address , uint256 ) view returns(uint128 amount, uint48 unlockTime, bool completed)
func (_BunkerDelegation *BunkerDelegationSession) UnbondingQueues(arg0 common.Address, arg1 *big.Int) (struct {
	Amount     *big.Int
	UnlockTime *big.Int
	Completed  bool
}, error) {
	return _BunkerDelegation.Contract.UnbondingQueues(&_BunkerDelegation.CallOpts, arg0, arg1)
}

// UnbondingQueues is a free data retrieval call binding the contract method 0x28a78f43.
//
// Solidity: function unbondingQueues(address , uint256 ) view returns(uint128 amount, uint48 unlockTime, bool completed)
func (_BunkerDelegation *BunkerDelegationCallerSession) UnbondingQueues(arg0 common.Address, arg1 *big.Int) (struct {
	Amount     *big.Int
	UnlockTime *big.Int
	Completed  bool
}, error) {
	return _BunkerDelegation.Contract.UnbondingQueues(&_BunkerDelegation.CallOpts, arg0, arg1)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerDelegation *BunkerDelegationTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerDelegation *BunkerDelegationSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerDelegation.Contract.AcceptOwnership(&_BunkerDelegation.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerDelegation.Contract.AcceptOwnership(&_BunkerDelegation.TransactOpts)
}

// CompleteUndelegate is a paid mutator transaction binding the contract method 0xdfb22ad1.
//
// Solidity: function completeUndelegate(uint256 requestIndex) returns()
func (_BunkerDelegation *BunkerDelegationTransactor) CompleteUndelegate(opts *bind.TransactOpts, requestIndex *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "completeUndelegate", requestIndex)
}

// CompleteUndelegate is a paid mutator transaction binding the contract method 0xdfb22ad1.
//
// Solidity: function completeUndelegate(uint256 requestIndex) returns()
func (_BunkerDelegation *BunkerDelegationSession) CompleteUndelegate(requestIndex *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.CompleteUndelegate(&_BunkerDelegation.TransactOpts, requestIndex)
}

// CompleteUndelegate is a paid mutator transaction binding the contract method 0xdfb22ad1.
//
// Solidity: function completeUndelegate(uint256 requestIndex) returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) CompleteUndelegate(requestIndex *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.CompleteUndelegate(&_BunkerDelegation.TransactOpts, requestIndex)
}

// Delegate is a paid mutator transaction binding the contract method 0x026e402b.
//
// Solidity: function delegate(address provider, uint256 amount) returns()
func (_BunkerDelegation *BunkerDelegationTransactor) Delegate(opts *bind.TransactOpts, provider common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "delegate", provider, amount)
}

// Delegate is a paid mutator transaction binding the contract method 0x026e402b.
//
// Solidity: function delegate(address provider, uint256 amount) returns()
func (_BunkerDelegation *BunkerDelegationSession) Delegate(provider common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.Delegate(&_BunkerDelegation.TransactOpts, provider, amount)
}

// Delegate is a paid mutator transaction binding the contract method 0x026e402b.
//
// Solidity: function delegate(address provider, uint256 amount) returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) Delegate(provider common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.Delegate(&_BunkerDelegation.TransactOpts, provider, amount)
}

// FinalizeRewardCut is a paid mutator transaction binding the contract method 0x6063f05c.
//
// Solidity: function finalizeRewardCut() returns()
func (_BunkerDelegation *BunkerDelegationTransactor) FinalizeRewardCut(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "finalizeRewardCut")
}

// FinalizeRewardCut is a paid mutator transaction binding the contract method 0x6063f05c.
//
// Solidity: function finalizeRewardCut() returns()
func (_BunkerDelegation *BunkerDelegationSession) FinalizeRewardCut() (*types.Transaction, error) {
	return _BunkerDelegation.Contract.FinalizeRewardCut(&_BunkerDelegation.TransactOpts)
}

// FinalizeRewardCut is a paid mutator transaction binding the contract method 0x6063f05c.
//
// Solidity: function finalizeRewardCut() returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) FinalizeRewardCut() (*types.Transaction, error) {
	return _BunkerDelegation.Contract.FinalizeRewardCut(&_BunkerDelegation.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerDelegation *BunkerDelegationTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerDelegation *BunkerDelegationSession) Pause() (*types.Transaction, error) {
	return _BunkerDelegation.Contract.Pause(&_BunkerDelegation.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) Pause() (*types.Transaction, error) {
	return _BunkerDelegation.Contract.Pause(&_BunkerDelegation.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerDelegation *BunkerDelegationTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerDelegation *BunkerDelegationSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerDelegation.Contract.RenounceOwnership(&_BunkerDelegation.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerDelegation.Contract.RenounceOwnership(&_BunkerDelegation.TransactOpts)
}

// RequestUndelegate is a paid mutator transaction binding the contract method 0xf86bc80c.
//
// Solidity: function requestUndelegate(uint256 amount) returns()
func (_BunkerDelegation *BunkerDelegationTransactor) RequestUndelegate(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "requestUndelegate", amount)
}

// RequestUndelegate is a paid mutator transaction binding the contract method 0xf86bc80c.
//
// Solidity: function requestUndelegate(uint256 amount) returns()
func (_BunkerDelegation *BunkerDelegationSession) RequestUndelegate(amount *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.RequestUndelegate(&_BunkerDelegation.TransactOpts, amount)
}

// RequestUndelegate is a paid mutator transaction binding the contract method 0xf86bc80c.
//
// Solidity: function requestUndelegate(uint256 amount) returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) RequestUndelegate(amount *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.RequestUndelegate(&_BunkerDelegation.TransactOpts, amount)
}

// SetDelegationConfig is a paid mutator transaction binding the contract method 0x9047b71f.
//
// Solidity: function setDelegationConfig(uint16 rewardCutBps, uint16 feeShareBps) returns()
func (_BunkerDelegation *BunkerDelegationTransactor) SetDelegationConfig(opts *bind.TransactOpts, rewardCutBps uint16, feeShareBps uint16) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "setDelegationConfig", rewardCutBps, feeShareBps)
}

// SetDelegationConfig is a paid mutator transaction binding the contract method 0x9047b71f.
//
// Solidity: function setDelegationConfig(uint16 rewardCutBps, uint16 feeShareBps) returns()
func (_BunkerDelegation *BunkerDelegationSession) SetDelegationConfig(rewardCutBps uint16, feeShareBps uint16) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.SetDelegationConfig(&_BunkerDelegation.TransactOpts, rewardCutBps, feeShareBps)
}

// SetDelegationConfig is a paid mutator transaction binding the contract method 0x9047b71f.
//
// Solidity: function setDelegationConfig(uint16 rewardCutBps, uint16 feeShareBps) returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) SetDelegationConfig(rewardCutBps uint16, feeShareBps uint16) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.SetDelegationConfig(&_BunkerDelegation.TransactOpts, rewardCutBps, feeShareBps)
}

// SetMaxRewardCutBps is a paid mutator transaction binding the contract method 0x6f91a0dd.
//
// Solidity: function setMaxRewardCutBps(uint16 newMax) returns()
func (_BunkerDelegation *BunkerDelegationTransactor) SetMaxRewardCutBps(opts *bind.TransactOpts, newMax uint16) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "setMaxRewardCutBps", newMax)
}

// SetMaxRewardCutBps is a paid mutator transaction binding the contract method 0x6f91a0dd.
//
// Solidity: function setMaxRewardCutBps(uint16 newMax) returns()
func (_BunkerDelegation *BunkerDelegationSession) SetMaxRewardCutBps(newMax uint16) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.SetMaxRewardCutBps(&_BunkerDelegation.TransactOpts, newMax)
}

// SetMaxRewardCutBps is a paid mutator transaction binding the contract method 0x6f91a0dd.
//
// Solidity: function setMaxRewardCutBps(uint16 newMax) returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) SetMaxRewardCutBps(newMax uint16) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.SetMaxRewardCutBps(&_BunkerDelegation.TransactOpts, newMax)
}

// SetStakingContract is a paid mutator transaction binding the contract method 0x9dd373b9.
//
// Solidity: function setStakingContract(address _stakingContract) returns()
func (_BunkerDelegation *BunkerDelegationTransactor) SetStakingContract(opts *bind.TransactOpts, _stakingContract common.Address) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "setStakingContract", _stakingContract)
}

// SetStakingContract is a paid mutator transaction binding the contract method 0x9dd373b9.
//
// Solidity: function setStakingContract(address _stakingContract) returns()
func (_BunkerDelegation *BunkerDelegationSession) SetStakingContract(_stakingContract common.Address) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.SetStakingContract(&_BunkerDelegation.TransactOpts, _stakingContract)
}

// SetStakingContract is a paid mutator transaction binding the contract method 0x9dd373b9.
//
// Solidity: function setStakingContract(address _stakingContract) returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) SetStakingContract(_stakingContract common.Address) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.SetStakingContract(&_BunkerDelegation.TransactOpts, _stakingContract)
}

// SetUnbondingPeriod is a paid mutator transaction binding the contract method 0x114eaf55.
//
// Solidity: function setUnbondingPeriod(uint256 newPeriod) returns()
func (_BunkerDelegation *BunkerDelegationTransactor) SetUnbondingPeriod(opts *bind.TransactOpts, newPeriod *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "setUnbondingPeriod", newPeriod)
}

// SetUnbondingPeriod is a paid mutator transaction binding the contract method 0x114eaf55.
//
// Solidity: function setUnbondingPeriod(uint256 newPeriod) returns()
func (_BunkerDelegation *BunkerDelegationSession) SetUnbondingPeriod(newPeriod *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.SetUnbondingPeriod(&_BunkerDelegation.TransactOpts, newPeriod)
}

// SetUnbondingPeriod is a paid mutator transaction binding the contract method 0x114eaf55.
//
// Solidity: function setUnbondingPeriod(uint256 newPeriod) returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) SetUnbondingPeriod(newPeriod *big.Int) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.SetUnbondingPeriod(&_BunkerDelegation.TransactOpts, newPeriod)
}

// ToggleAcceptDelegations is a paid mutator transaction binding the contract method 0xf86d6829.
//
// Solidity: function toggleAcceptDelegations(bool accepting) returns()
func (_BunkerDelegation *BunkerDelegationTransactor) ToggleAcceptDelegations(opts *bind.TransactOpts, accepting bool) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "toggleAcceptDelegations", accepting)
}

// ToggleAcceptDelegations is a paid mutator transaction binding the contract method 0xf86d6829.
//
// Solidity: function toggleAcceptDelegations(bool accepting) returns()
func (_BunkerDelegation *BunkerDelegationSession) ToggleAcceptDelegations(accepting bool) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.ToggleAcceptDelegations(&_BunkerDelegation.TransactOpts, accepting)
}

// ToggleAcceptDelegations is a paid mutator transaction binding the contract method 0xf86d6829.
//
// Solidity: function toggleAcceptDelegations(bool accepting) returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) ToggleAcceptDelegations(accepting bool) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.ToggleAcceptDelegations(&_BunkerDelegation.TransactOpts, accepting)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerDelegation *BunkerDelegationTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerDelegation *BunkerDelegationSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.TransferOwnership(&_BunkerDelegation.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerDelegation.Contract.TransferOwnership(&_BunkerDelegation.TransactOpts, newOwner)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerDelegation *BunkerDelegationTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerDelegation.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerDelegation *BunkerDelegationSession) Unpause() (*types.Transaction, error) {
	return _BunkerDelegation.Contract.Unpause(&_BunkerDelegation.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerDelegation *BunkerDelegationTransactorSession) Unpause() (*types.Transaction, error) {
	return _BunkerDelegation.Contract.Unpause(&_BunkerDelegation.TransactOpts)
}

// BunkerDelegationDelegatedIterator is returned from FilterDelegated and is used to iterate over the raw logs and unpacked data for Delegated events raised by the BunkerDelegation contract.
type BunkerDelegationDelegatedIterator struct {
	Event *BunkerDelegationDelegated // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationDelegatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationDelegated)
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
		it.Event = new(BunkerDelegationDelegated)
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
func (it *BunkerDelegationDelegatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationDelegatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationDelegated represents a Delegated event raised by the BunkerDelegation contract.
type BunkerDelegationDelegated struct {
	Delegator common.Address
	Provider  common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegated is a free log retrieval operation binding the contract event 0xe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b.
//
// Solidity: event Delegated(address indexed delegator, address indexed provider, uint256 amount)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterDelegated(opts *bind.FilterOpts, delegator []common.Address, provider []common.Address) (*BunkerDelegationDelegatedIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "Delegated", delegatorRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationDelegatedIterator{contract: _BunkerDelegation.contract, event: "Delegated", logs: logs, sub: sub}, nil
}

// WatchDelegated is a free log subscription operation binding the contract event 0xe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b.
//
// Solidity: event Delegated(address indexed delegator, address indexed provider, uint256 amount)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchDelegated(opts *bind.WatchOpts, sink chan<- *BunkerDelegationDelegated, delegator []common.Address, provider []common.Address) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "Delegated", delegatorRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationDelegated)
				if err := _BunkerDelegation.contract.UnpackLog(event, "Delegated", log); err != nil {
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

// ParseDelegated is a log parse operation binding the contract event 0xe5541a6b6103d4fa7e021ed54fad39c66f27a76bd13d374cf6240ae6bd0bb72b.
//
// Solidity: event Delegated(address indexed delegator, address indexed provider, uint256 amount)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseDelegated(log types.Log) (*BunkerDelegationDelegated, error) {
	event := new(BunkerDelegationDelegated)
	if err := _BunkerDelegation.contract.UnpackLog(event, "Delegated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationDelegationAcceptanceToggledIterator is returned from FilterDelegationAcceptanceToggled and is used to iterate over the raw logs and unpacked data for DelegationAcceptanceToggled events raised by the BunkerDelegation contract.
type BunkerDelegationDelegationAcceptanceToggledIterator struct {
	Event *BunkerDelegationDelegationAcceptanceToggled // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationDelegationAcceptanceToggledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationDelegationAcceptanceToggled)
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
		it.Event = new(BunkerDelegationDelegationAcceptanceToggled)
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
func (it *BunkerDelegationDelegationAcceptanceToggledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationDelegationAcceptanceToggledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationDelegationAcceptanceToggled represents a DelegationAcceptanceToggled event raised by the BunkerDelegation contract.
type BunkerDelegationDelegationAcceptanceToggled struct {
	Provider  common.Address
	Accepting bool
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegationAcceptanceToggled is a free log retrieval operation binding the contract event 0x6b98bc340f3988300f115b17fe5da378d561c6317fcc3fbfda4033e77f2699ef.
//
// Solidity: event DelegationAcceptanceToggled(address indexed provider, bool accepting)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterDelegationAcceptanceToggled(opts *bind.FilterOpts, provider []common.Address) (*BunkerDelegationDelegationAcceptanceToggledIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "DelegationAcceptanceToggled", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationDelegationAcceptanceToggledIterator{contract: _BunkerDelegation.contract, event: "DelegationAcceptanceToggled", logs: logs, sub: sub}, nil
}

// WatchDelegationAcceptanceToggled is a free log subscription operation binding the contract event 0x6b98bc340f3988300f115b17fe5da378d561c6317fcc3fbfda4033e77f2699ef.
//
// Solidity: event DelegationAcceptanceToggled(address indexed provider, bool accepting)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchDelegationAcceptanceToggled(opts *bind.WatchOpts, sink chan<- *BunkerDelegationDelegationAcceptanceToggled, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "DelegationAcceptanceToggled", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationDelegationAcceptanceToggled)
				if err := _BunkerDelegation.contract.UnpackLog(event, "DelegationAcceptanceToggled", log); err != nil {
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

// ParseDelegationAcceptanceToggled is a log parse operation binding the contract event 0x6b98bc340f3988300f115b17fe5da378d561c6317fcc3fbfda4033e77f2699ef.
//
// Solidity: event DelegationAcceptanceToggled(address indexed provider, bool accepting)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseDelegationAcceptanceToggled(log types.Log) (*BunkerDelegationDelegationAcceptanceToggled, error) {
	event := new(BunkerDelegationDelegationAcceptanceToggled)
	if err := _BunkerDelegation.contract.UnpackLog(event, "DelegationAcceptanceToggled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationDelegationWithdrawnIterator is returned from FilterDelegationWithdrawn and is used to iterate over the raw logs and unpacked data for DelegationWithdrawn events raised by the BunkerDelegation contract.
type BunkerDelegationDelegationWithdrawnIterator struct {
	Event *BunkerDelegationDelegationWithdrawn // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationDelegationWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationDelegationWithdrawn)
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
		it.Event = new(BunkerDelegationDelegationWithdrawn)
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
func (it *BunkerDelegationDelegationWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationDelegationWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationDelegationWithdrawn represents a DelegationWithdrawn event raised by the BunkerDelegation contract.
type BunkerDelegationDelegationWithdrawn struct {
	Delegator common.Address
	Provider  common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegationWithdrawn is a free log retrieval operation binding the contract event 0xaa0001b0e2e76b9a1b925257fefa58b756aebc52d2d7c8a85ea5beacc77a2100.
//
// Solidity: event DelegationWithdrawn(address indexed delegator, address indexed provider, uint256 amount)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterDelegationWithdrawn(opts *bind.FilterOpts, delegator []common.Address, provider []common.Address) (*BunkerDelegationDelegationWithdrawnIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "DelegationWithdrawn", delegatorRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationDelegationWithdrawnIterator{contract: _BunkerDelegation.contract, event: "DelegationWithdrawn", logs: logs, sub: sub}, nil
}

// WatchDelegationWithdrawn is a free log subscription operation binding the contract event 0xaa0001b0e2e76b9a1b925257fefa58b756aebc52d2d7c8a85ea5beacc77a2100.
//
// Solidity: event DelegationWithdrawn(address indexed delegator, address indexed provider, uint256 amount)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchDelegationWithdrawn(opts *bind.WatchOpts, sink chan<- *BunkerDelegationDelegationWithdrawn, delegator []common.Address, provider []common.Address) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "DelegationWithdrawn", delegatorRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationDelegationWithdrawn)
				if err := _BunkerDelegation.contract.UnpackLog(event, "DelegationWithdrawn", log); err != nil {
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

// ParseDelegationWithdrawn is a log parse operation binding the contract event 0xaa0001b0e2e76b9a1b925257fefa58b756aebc52d2d7c8a85ea5beacc77a2100.
//
// Solidity: event DelegationWithdrawn(address indexed delegator, address indexed provider, uint256 amount)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseDelegationWithdrawn(log types.Log) (*BunkerDelegationDelegationWithdrawn, error) {
	event := new(BunkerDelegationDelegationWithdrawn)
	if err := _BunkerDelegation.contract.UnpackLog(event, "DelegationWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationMaxRewardCutUpdatedIterator is returned from FilterMaxRewardCutUpdated and is used to iterate over the raw logs and unpacked data for MaxRewardCutUpdated events raised by the BunkerDelegation contract.
type BunkerDelegationMaxRewardCutUpdatedIterator struct {
	Event *BunkerDelegationMaxRewardCutUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationMaxRewardCutUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationMaxRewardCutUpdated)
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
		it.Event = new(BunkerDelegationMaxRewardCutUpdated)
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
func (it *BunkerDelegationMaxRewardCutUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationMaxRewardCutUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationMaxRewardCutUpdated represents a MaxRewardCutUpdated event raised by the BunkerDelegation contract.
type BunkerDelegationMaxRewardCutUpdated struct {
	NewMax uint16
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMaxRewardCutUpdated is a free log retrieval operation binding the contract event 0xfe2dd0e7c42a6b3aaeb9c8487d1fa6c72bebea32edce7b0d34d2b86b4b90844a.
//
// Solidity: event MaxRewardCutUpdated(uint16 newMax)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterMaxRewardCutUpdated(opts *bind.FilterOpts) (*BunkerDelegationMaxRewardCutUpdatedIterator, error) {

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "MaxRewardCutUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationMaxRewardCutUpdatedIterator{contract: _BunkerDelegation.contract, event: "MaxRewardCutUpdated", logs: logs, sub: sub}, nil
}

// WatchMaxRewardCutUpdated is a free log subscription operation binding the contract event 0xfe2dd0e7c42a6b3aaeb9c8487d1fa6c72bebea32edce7b0d34d2b86b4b90844a.
//
// Solidity: event MaxRewardCutUpdated(uint16 newMax)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchMaxRewardCutUpdated(opts *bind.WatchOpts, sink chan<- *BunkerDelegationMaxRewardCutUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "MaxRewardCutUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationMaxRewardCutUpdated)
				if err := _BunkerDelegation.contract.UnpackLog(event, "MaxRewardCutUpdated", log); err != nil {
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

// ParseMaxRewardCutUpdated is a log parse operation binding the contract event 0xfe2dd0e7c42a6b3aaeb9c8487d1fa6c72bebea32edce7b0d34d2b86b4b90844a.
//
// Solidity: event MaxRewardCutUpdated(uint16 newMax)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseMaxRewardCutUpdated(log types.Log) (*BunkerDelegationMaxRewardCutUpdated, error) {
	event := new(BunkerDelegationMaxRewardCutUpdated)
	if err := _BunkerDelegation.contract.UnpackLog(event, "MaxRewardCutUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the BunkerDelegation contract.
type BunkerDelegationOwnershipTransferStartedIterator struct {
	Event *BunkerDelegationOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationOwnershipTransferStarted)
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
		it.Event = new(BunkerDelegationOwnershipTransferStarted)
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
func (it *BunkerDelegationOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the BunkerDelegation contract.
type BunkerDelegationOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerDelegationOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationOwnershipTransferStartedIterator{contract: _BunkerDelegation.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *BunkerDelegationOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationOwnershipTransferStarted)
				if err := _BunkerDelegation.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_BunkerDelegation *BunkerDelegationFilterer) ParseOwnershipTransferStarted(log types.Log) (*BunkerDelegationOwnershipTransferStarted, error) {
	event := new(BunkerDelegationOwnershipTransferStarted)
	if err := _BunkerDelegation.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BunkerDelegation contract.
type BunkerDelegationOwnershipTransferredIterator struct {
	Event *BunkerDelegationOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationOwnershipTransferred)
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
		it.Event = new(BunkerDelegationOwnershipTransferred)
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
func (it *BunkerDelegationOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationOwnershipTransferred represents a OwnershipTransferred event raised by the BunkerDelegation contract.
type BunkerDelegationOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerDelegationOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationOwnershipTransferredIterator{contract: _BunkerDelegation.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BunkerDelegationOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationOwnershipTransferred)
				if err := _BunkerDelegation.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BunkerDelegation *BunkerDelegationFilterer) ParseOwnershipTransferred(log types.Log) (*BunkerDelegationOwnershipTransferred, error) {
	event := new(BunkerDelegationOwnershipTransferred)
	if err := _BunkerDelegation.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the BunkerDelegation contract.
type BunkerDelegationPausedIterator struct {
	Event *BunkerDelegationPaused // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationPaused)
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
		it.Event = new(BunkerDelegationPaused)
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
func (it *BunkerDelegationPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationPaused represents a Paused event raised by the BunkerDelegation contract.
type BunkerDelegationPaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterPaused(opts *bind.FilterOpts) (*BunkerDelegationPausedIterator, error) {

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationPausedIterator{contract: _BunkerDelegation.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *BunkerDelegationPaused) (event.Subscription, error) {

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationPaused)
				if err := _BunkerDelegation.contract.UnpackLog(event, "Paused", log); err != nil {
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

// ParsePaused is a log parse operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_BunkerDelegation *BunkerDelegationFilterer) ParsePaused(log types.Log) (*BunkerDelegationPaused, error) {
	event := new(BunkerDelegationPaused)
	if err := _BunkerDelegation.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationProviderConfigUpdatedIterator is returned from FilterProviderConfigUpdated and is used to iterate over the raw logs and unpacked data for ProviderConfigUpdated events raised by the BunkerDelegation contract.
type BunkerDelegationProviderConfigUpdatedIterator struct {
	Event *BunkerDelegationProviderConfigUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationProviderConfigUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationProviderConfigUpdated)
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
		it.Event = new(BunkerDelegationProviderConfigUpdated)
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
func (it *BunkerDelegationProviderConfigUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationProviderConfigUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationProviderConfigUpdated represents a ProviderConfigUpdated event raised by the BunkerDelegation contract.
type BunkerDelegationProviderConfigUpdated struct {
	Provider     common.Address
	RewardCutBps uint16
	FeeShareBps  uint16
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterProviderConfigUpdated is a free log retrieval operation binding the contract event 0x06f77c1359530c8fc190310202965155808170209f916f4b0a5efac34342e1d7.
//
// Solidity: event ProviderConfigUpdated(address indexed provider, uint16 rewardCutBps, uint16 feeShareBps)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterProviderConfigUpdated(opts *bind.FilterOpts, provider []common.Address) (*BunkerDelegationProviderConfigUpdatedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "ProviderConfigUpdated", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationProviderConfigUpdatedIterator{contract: _BunkerDelegation.contract, event: "ProviderConfigUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderConfigUpdated is a free log subscription operation binding the contract event 0x06f77c1359530c8fc190310202965155808170209f916f4b0a5efac34342e1d7.
//
// Solidity: event ProviderConfigUpdated(address indexed provider, uint16 rewardCutBps, uint16 feeShareBps)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchProviderConfigUpdated(opts *bind.WatchOpts, sink chan<- *BunkerDelegationProviderConfigUpdated, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "ProviderConfigUpdated", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationProviderConfigUpdated)
				if err := _BunkerDelegation.contract.UnpackLog(event, "ProviderConfigUpdated", log); err != nil {
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

// ParseProviderConfigUpdated is a log parse operation binding the contract event 0x06f77c1359530c8fc190310202965155808170209f916f4b0a5efac34342e1d7.
//
// Solidity: event ProviderConfigUpdated(address indexed provider, uint16 rewardCutBps, uint16 feeShareBps)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseProviderConfigUpdated(log types.Log) (*BunkerDelegationProviderConfigUpdated, error) {
	event := new(BunkerDelegationProviderConfigUpdated)
	if err := _BunkerDelegation.contract.UnpackLog(event, "ProviderConfigUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationRewardCutFinalizedIterator is returned from FilterRewardCutFinalized and is used to iterate over the raw logs and unpacked data for RewardCutFinalized events raised by the BunkerDelegation contract.
type BunkerDelegationRewardCutFinalizedIterator struct {
	Event *BunkerDelegationRewardCutFinalized // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationRewardCutFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationRewardCutFinalized)
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
		it.Event = new(BunkerDelegationRewardCutFinalized)
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
func (it *BunkerDelegationRewardCutFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationRewardCutFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationRewardCutFinalized represents a RewardCutFinalized event raised by the BunkerDelegation contract.
type BunkerDelegationRewardCutFinalized struct {
	Provider     common.Address
	RewardCutBps uint16
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterRewardCutFinalized is a free log retrieval operation binding the contract event 0xb850bcde10c1aa68133a57670c0c9f790f4a8e2454836924c876c79a01c259f2.
//
// Solidity: event RewardCutFinalized(address indexed provider, uint16 rewardCutBps)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterRewardCutFinalized(opts *bind.FilterOpts, provider []common.Address) (*BunkerDelegationRewardCutFinalizedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "RewardCutFinalized", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationRewardCutFinalizedIterator{contract: _BunkerDelegation.contract, event: "RewardCutFinalized", logs: logs, sub: sub}, nil
}

// WatchRewardCutFinalized is a free log subscription operation binding the contract event 0xb850bcde10c1aa68133a57670c0c9f790f4a8e2454836924c876c79a01c259f2.
//
// Solidity: event RewardCutFinalized(address indexed provider, uint16 rewardCutBps)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchRewardCutFinalized(opts *bind.WatchOpts, sink chan<- *BunkerDelegationRewardCutFinalized, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "RewardCutFinalized", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationRewardCutFinalized)
				if err := _BunkerDelegation.contract.UnpackLog(event, "RewardCutFinalized", log); err != nil {
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

// ParseRewardCutFinalized is a log parse operation binding the contract event 0xb850bcde10c1aa68133a57670c0c9f790f4a8e2454836924c876c79a01c259f2.
//
// Solidity: event RewardCutFinalized(address indexed provider, uint16 rewardCutBps)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseRewardCutFinalized(log types.Log) (*BunkerDelegationRewardCutFinalized, error) {
	event := new(BunkerDelegationRewardCutFinalized)
	if err := _BunkerDelegation.contract.UnpackLog(event, "RewardCutFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationRewardCutIncreaseScheduledIterator is returned from FilterRewardCutIncreaseScheduled and is used to iterate over the raw logs and unpacked data for RewardCutIncreaseScheduled events raised by the BunkerDelegation contract.
type BunkerDelegationRewardCutIncreaseScheduledIterator struct {
	Event *BunkerDelegationRewardCutIncreaseScheduled // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationRewardCutIncreaseScheduledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationRewardCutIncreaseScheduled)
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
		it.Event = new(BunkerDelegationRewardCutIncreaseScheduled)
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
func (it *BunkerDelegationRewardCutIncreaseScheduledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationRewardCutIncreaseScheduledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationRewardCutIncreaseScheduled represents a RewardCutIncreaseScheduled event raised by the BunkerDelegation contract.
type BunkerDelegationRewardCutIncreaseScheduled struct {
	Provider        common.Address
	NewRewardCutBps uint16
	EffectiveAt     *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterRewardCutIncreaseScheduled is a free log retrieval operation binding the contract event 0x13c6fcb618d3034412ea56025117271a589dcb26f773ea346a1d057d2bdf4fde.
//
// Solidity: event RewardCutIncreaseScheduled(address indexed provider, uint16 newRewardCutBps, uint48 effectiveAt)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterRewardCutIncreaseScheduled(opts *bind.FilterOpts, provider []common.Address) (*BunkerDelegationRewardCutIncreaseScheduledIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "RewardCutIncreaseScheduled", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationRewardCutIncreaseScheduledIterator{contract: _BunkerDelegation.contract, event: "RewardCutIncreaseScheduled", logs: logs, sub: sub}, nil
}

// WatchRewardCutIncreaseScheduled is a free log subscription operation binding the contract event 0x13c6fcb618d3034412ea56025117271a589dcb26f773ea346a1d057d2bdf4fde.
//
// Solidity: event RewardCutIncreaseScheduled(address indexed provider, uint16 newRewardCutBps, uint48 effectiveAt)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchRewardCutIncreaseScheduled(opts *bind.WatchOpts, sink chan<- *BunkerDelegationRewardCutIncreaseScheduled, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "RewardCutIncreaseScheduled", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationRewardCutIncreaseScheduled)
				if err := _BunkerDelegation.contract.UnpackLog(event, "RewardCutIncreaseScheduled", log); err != nil {
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

// ParseRewardCutIncreaseScheduled is a log parse operation binding the contract event 0x13c6fcb618d3034412ea56025117271a589dcb26f773ea346a1d057d2bdf4fde.
//
// Solidity: event RewardCutIncreaseScheduled(address indexed provider, uint16 newRewardCutBps, uint48 effectiveAt)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseRewardCutIncreaseScheduled(log types.Log) (*BunkerDelegationRewardCutIncreaseScheduled, error) {
	event := new(BunkerDelegationRewardCutIncreaseScheduled)
	if err := _BunkerDelegation.contract.UnpackLog(event, "RewardCutIncreaseScheduled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationStakingContractUpdatedIterator is returned from FilterStakingContractUpdated and is used to iterate over the raw logs and unpacked data for StakingContractUpdated events raised by the BunkerDelegation contract.
type BunkerDelegationStakingContractUpdatedIterator struct {
	Event *BunkerDelegationStakingContractUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationStakingContractUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationStakingContractUpdated)
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
		it.Event = new(BunkerDelegationStakingContractUpdated)
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
func (it *BunkerDelegationStakingContractUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationStakingContractUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationStakingContractUpdated represents a StakingContractUpdated event raised by the BunkerDelegation contract.
type BunkerDelegationStakingContractUpdated struct {
	OldStaking common.Address
	NewStaking common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterStakingContractUpdated is a free log retrieval operation binding the contract event 0x7042586b23181180eb30b4798702d7a0233b7fc2551e89806770e8e5d9392e6a.
//
// Solidity: event StakingContractUpdated(address indexed oldStaking, address indexed newStaking)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterStakingContractUpdated(opts *bind.FilterOpts, oldStaking []common.Address, newStaking []common.Address) (*BunkerDelegationStakingContractUpdatedIterator, error) {

	var oldStakingRule []interface{}
	for _, oldStakingItem := range oldStaking {
		oldStakingRule = append(oldStakingRule, oldStakingItem)
	}
	var newStakingRule []interface{}
	for _, newStakingItem := range newStaking {
		newStakingRule = append(newStakingRule, newStakingItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "StakingContractUpdated", oldStakingRule, newStakingRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationStakingContractUpdatedIterator{contract: _BunkerDelegation.contract, event: "StakingContractUpdated", logs: logs, sub: sub}, nil
}

// WatchStakingContractUpdated is a free log subscription operation binding the contract event 0x7042586b23181180eb30b4798702d7a0233b7fc2551e89806770e8e5d9392e6a.
//
// Solidity: event StakingContractUpdated(address indexed oldStaking, address indexed newStaking)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchStakingContractUpdated(opts *bind.WatchOpts, sink chan<- *BunkerDelegationStakingContractUpdated, oldStaking []common.Address, newStaking []common.Address) (event.Subscription, error) {

	var oldStakingRule []interface{}
	for _, oldStakingItem := range oldStaking {
		oldStakingRule = append(oldStakingRule, oldStakingItem)
	}
	var newStakingRule []interface{}
	for _, newStakingItem := range newStaking {
		newStakingRule = append(newStakingRule, newStakingItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "StakingContractUpdated", oldStakingRule, newStakingRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationStakingContractUpdated)
				if err := _BunkerDelegation.contract.UnpackLog(event, "StakingContractUpdated", log); err != nil {
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

// ParseStakingContractUpdated is a log parse operation binding the contract event 0x7042586b23181180eb30b4798702d7a0233b7fc2551e89806770e8e5d9392e6a.
//
// Solidity: event StakingContractUpdated(address indexed oldStaking, address indexed newStaking)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseStakingContractUpdated(log types.Log) (*BunkerDelegationStakingContractUpdated, error) {
	event := new(BunkerDelegationStakingContractUpdated)
	if err := _BunkerDelegation.contract.UnpackLog(event, "StakingContractUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationUnbondingPeriodUpdatedIterator is returned from FilterUnbondingPeriodUpdated and is used to iterate over the raw logs and unpacked data for UnbondingPeriodUpdated events raised by the BunkerDelegation contract.
type BunkerDelegationUnbondingPeriodUpdatedIterator struct {
	Event *BunkerDelegationUnbondingPeriodUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationUnbondingPeriodUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationUnbondingPeriodUpdated)
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
		it.Event = new(BunkerDelegationUnbondingPeriodUpdated)
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
func (it *BunkerDelegationUnbondingPeriodUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationUnbondingPeriodUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationUnbondingPeriodUpdated represents a UnbondingPeriodUpdated event raised by the BunkerDelegation contract.
type BunkerDelegationUnbondingPeriodUpdated struct {
	NewPeriod *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterUnbondingPeriodUpdated is a free log retrieval operation binding the contract event 0x38a1644f59872db6fd17fdced395495fcaa3ca7d2825a0704db5a3acbd1dd064.
//
// Solidity: event UnbondingPeriodUpdated(uint256 newPeriod)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterUnbondingPeriodUpdated(opts *bind.FilterOpts) (*BunkerDelegationUnbondingPeriodUpdatedIterator, error) {

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "UnbondingPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationUnbondingPeriodUpdatedIterator{contract: _BunkerDelegation.contract, event: "UnbondingPeriodUpdated", logs: logs, sub: sub}, nil
}

// WatchUnbondingPeriodUpdated is a free log subscription operation binding the contract event 0x38a1644f59872db6fd17fdced395495fcaa3ca7d2825a0704db5a3acbd1dd064.
//
// Solidity: event UnbondingPeriodUpdated(uint256 newPeriod)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchUnbondingPeriodUpdated(opts *bind.WatchOpts, sink chan<- *BunkerDelegationUnbondingPeriodUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "UnbondingPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationUnbondingPeriodUpdated)
				if err := _BunkerDelegation.contract.UnpackLog(event, "UnbondingPeriodUpdated", log); err != nil {
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

// ParseUnbondingPeriodUpdated is a log parse operation binding the contract event 0x38a1644f59872db6fd17fdced395495fcaa3ca7d2825a0704db5a3acbd1dd064.
//
// Solidity: event UnbondingPeriodUpdated(uint256 newPeriod)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseUnbondingPeriodUpdated(log types.Log) (*BunkerDelegationUnbondingPeriodUpdated, error) {
	event := new(BunkerDelegationUnbondingPeriodUpdated)
	if err := _BunkerDelegation.contract.UnpackLog(event, "UnbondingPeriodUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationUndelegateCompletedIterator is returned from FilterUndelegateCompleted and is used to iterate over the raw logs and unpacked data for UndelegateCompleted events raised by the BunkerDelegation contract.
type BunkerDelegationUndelegateCompletedIterator struct {
	Event *BunkerDelegationUndelegateCompleted // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationUndelegateCompletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationUndelegateCompleted)
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
		it.Event = new(BunkerDelegationUndelegateCompleted)
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
func (it *BunkerDelegationUndelegateCompletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationUndelegateCompletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationUndelegateCompleted represents a UndelegateCompleted event raised by the BunkerDelegation contract.
type BunkerDelegationUndelegateCompleted struct {
	Delegator common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterUndelegateCompleted is a free log retrieval operation binding the contract event 0xf3b3534c5f214996f6f38d120ceaa4508da213ebe303320f0ab3f4cf08e13397.
//
// Solidity: event UndelegateCompleted(address indexed delegator, uint256 amount)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterUndelegateCompleted(opts *bind.FilterOpts, delegator []common.Address) (*BunkerDelegationUndelegateCompletedIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "UndelegateCompleted", delegatorRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationUndelegateCompletedIterator{contract: _BunkerDelegation.contract, event: "UndelegateCompleted", logs: logs, sub: sub}, nil
}

// WatchUndelegateCompleted is a free log subscription operation binding the contract event 0xf3b3534c5f214996f6f38d120ceaa4508da213ebe303320f0ab3f4cf08e13397.
//
// Solidity: event UndelegateCompleted(address indexed delegator, uint256 amount)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchUndelegateCompleted(opts *bind.WatchOpts, sink chan<- *BunkerDelegationUndelegateCompleted, delegator []common.Address) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "UndelegateCompleted", delegatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationUndelegateCompleted)
				if err := _BunkerDelegation.contract.UnpackLog(event, "UndelegateCompleted", log); err != nil {
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

// ParseUndelegateCompleted is a log parse operation binding the contract event 0xf3b3534c5f214996f6f38d120ceaa4508da213ebe303320f0ab3f4cf08e13397.
//
// Solidity: event UndelegateCompleted(address indexed delegator, uint256 amount)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseUndelegateCompleted(log types.Log) (*BunkerDelegationUndelegateCompleted, error) {
	event := new(BunkerDelegationUndelegateCompleted)
	if err := _BunkerDelegation.contract.UnpackLog(event, "UndelegateCompleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationUndelegateRequestedIterator is returned from FilterUndelegateRequested and is used to iterate over the raw logs and unpacked data for UndelegateRequested events raised by the BunkerDelegation contract.
type BunkerDelegationUndelegateRequestedIterator struct {
	Event *BunkerDelegationUndelegateRequested // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationUndelegateRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationUndelegateRequested)
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
		it.Event = new(BunkerDelegationUndelegateRequested)
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
func (it *BunkerDelegationUndelegateRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationUndelegateRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationUndelegateRequested represents a UndelegateRequested event raised by the BunkerDelegation contract.
type BunkerDelegationUndelegateRequested struct {
	Delegator  common.Address
	Amount     *big.Int
	UnlockTime *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterUndelegateRequested is a free log retrieval operation binding the contract event 0xb8f4399de78dca166c5bc08ed55854b9f2c74a607cecd2fd94e9db166d51c425.
//
// Solidity: event UndelegateRequested(address indexed delegator, uint256 amount, uint256 unlockTime)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterUndelegateRequested(opts *bind.FilterOpts, delegator []common.Address) (*BunkerDelegationUndelegateRequestedIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "UndelegateRequested", delegatorRule)
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationUndelegateRequestedIterator{contract: _BunkerDelegation.contract, event: "UndelegateRequested", logs: logs, sub: sub}, nil
}

// WatchUndelegateRequested is a free log subscription operation binding the contract event 0xb8f4399de78dca166c5bc08ed55854b9f2c74a607cecd2fd94e9db166d51c425.
//
// Solidity: event UndelegateRequested(address indexed delegator, uint256 amount, uint256 unlockTime)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchUndelegateRequested(opts *bind.WatchOpts, sink chan<- *BunkerDelegationUndelegateRequested, delegator []common.Address) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "UndelegateRequested", delegatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationUndelegateRequested)
				if err := _BunkerDelegation.contract.UnpackLog(event, "UndelegateRequested", log); err != nil {
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

// ParseUndelegateRequested is a log parse operation binding the contract event 0xb8f4399de78dca166c5bc08ed55854b9f2c74a607cecd2fd94e9db166d51c425.
//
// Solidity: event UndelegateRequested(address indexed delegator, uint256 amount, uint256 unlockTime)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseUndelegateRequested(log types.Log) (*BunkerDelegationUndelegateRequested, error) {
	event := new(BunkerDelegationUndelegateRequested)
	if err := _BunkerDelegation.contract.UnpackLog(event, "UndelegateRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerDelegationUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the BunkerDelegation contract.
type BunkerDelegationUnpausedIterator struct {
	Event *BunkerDelegationUnpaused // Event containing the contract specifics and raw log

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
func (it *BunkerDelegationUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerDelegationUnpaused)
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
		it.Event = new(BunkerDelegationUnpaused)
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
func (it *BunkerDelegationUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerDelegationUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerDelegationUnpaused represents a Unpaused event raised by the BunkerDelegation contract.
type BunkerDelegationUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_BunkerDelegation *BunkerDelegationFilterer) FilterUnpaused(opts *bind.FilterOpts) (*BunkerDelegationUnpausedIterator, error) {

	logs, sub, err := _BunkerDelegation.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &BunkerDelegationUnpausedIterator{contract: _BunkerDelegation.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_BunkerDelegation *BunkerDelegationFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *BunkerDelegationUnpaused) (event.Subscription, error) {

	logs, sub, err := _BunkerDelegation.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerDelegationUnpaused)
				if err := _BunkerDelegation.contract.UnpackLog(event, "Unpaused", log); err != nil {
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

// ParseUnpaused is a log parse operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_BunkerDelegation *BunkerDelegationFilterer) ParseUnpaused(log types.Log) (*BunkerDelegationUnpaused, error) {
	event := new(BunkerDelegationUnpaused)
	if err := _BunkerDelegation.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
