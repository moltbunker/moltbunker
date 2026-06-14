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

// BunkerReputationReputationData is an auto generated low-level Go binding around an user-defined struct.
type BunkerReputationReputationData struct {
	Score         uint32
	JobsCompleted uint32
	JobsFailed    uint32
	SlashCount    uint32
	LastUpdated   *big.Int
	RegisteredAt  *big.Int
}

// BunkerReputationMetaData contains all meta data concerning the BunkerReputation contract.
var BunkerReputationMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"INITIAL_SCORE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_SCORE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"REPORTER_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"WEEK\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"applyDecay\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"decayFloor\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"decayRate\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getReputation\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"data\",\"type\":\"tuple\",\"internalType\":\"structBunkerReputation.ReputationData\",\"components\":[{\"name\":\"score\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"jobsCompleted\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"jobsFailed\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"slashCount\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"lastUpdated\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"registeredAt\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getScore\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTier\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumBunkerReputation.ReputationTier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"healthFailDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isEligibleForJobs\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"eligible\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"jobCompletedDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"jobEarlyDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"jobTimeoutDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxCustomDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"minCustomDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"minScoreForJobs\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"perfectUptimeDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"recordEvent\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"delta\",\"type\":\"int16\",\"internalType\":\"int16\"},{\"name\":\"reason\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"recordJobCompleted\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"recordJobFailed\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"recordSlashEvent\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerProvider\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"replicaMismatchDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"reputations\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"score\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"jobsCompleted\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"jobsFailed\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"slashCount\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"lastUpdated\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"registeredAt\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"securityViolationDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setDecayParams\",\"inputs\":[{\"name\":\"rate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"floor\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setDeltaParameters\",\"inputs\":[{\"name\":\"_jobCompletedDelta\",\"type\":\"int16\",\"internalType\":\"int16\"},{\"name\":\"_jobEarlyDelta\",\"type\":\"int16\",\"internalType\":\"int16\"},{\"name\":\"_perfectUptimeDelta\",\"type\":\"int16\",\"internalType\":\"int16\"},{\"name\":\"_jobTimeoutDelta\",\"type\":\"int16\",\"internalType\":\"int16\"},{\"name\":\"_healthFailDelta\",\"type\":\"int16\",\"internalType\":\"int16\"},{\"name\":\"_replicaMismatchDelta\",\"type\":\"int16\",\"internalType\":\"int16\"},{\"name\":\"_slashEventDelta\",\"type\":\"int16\",\"internalType\":\"int16\"},{\"name\":\"_securityViolationDelta\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxCustomDelta\",\"inputs\":[{\"name\":\"maxDelta\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMinScoreForJobs\",\"inputs\":[{\"name\":\"_minScore\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTierThresholds\",\"inputs\":[{\"name\":\"probation\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"standard\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"trusted\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"elite\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"slashEventDelta\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"int16\",\"internalType\":\"int16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"tierElite\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"tierProbation\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"tierStandard\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"tierTrusted\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"DecayParamsUpdated\",\"inputs\":[{\"name\":\"rate\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"floor\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"DeltaParametersUpdated\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MaxCustomDeltaUpdated\",\"inputs\":[{\"name\":\"maxDelta\",\"type\":\"int16\",\"indexed\":false,\"internalType\":\"int16\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MinScoreForJobsUpdated\",\"inputs\":[{\"name\":\"minScore\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderRegistered\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ScoreUpdated\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oldScore\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"newScore\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"},{\"name\":\"reason\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TierThresholdsUpdated\",\"inputs\":[{\"name\":\"probation\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"standard\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"trusted\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"elite\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"DeltaOutOfBounds\",\"inputs\":[{\"name\":\"delta\",\"type\":\"int16\",\"internalType\":\"int16\"}]},{\"type\":\"error\",\"name\":\"FloorTooHigh\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidDelta\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NegativeDeltaRequired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"PositiveDeltaRequired\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ProviderAlreadyRegistered\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ProviderNotRegistered\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ScoreExceedsMax\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ScoreOverflow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ThresholdExceedsMax\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ThresholdsNotAscending\",\"inputs\":[]}]",
}

// BunkerReputationABI is the input ABI used to generate the binding from.
// Deprecated: Use BunkerReputationMetaData.ABI instead.
var BunkerReputationABI = BunkerReputationMetaData.ABI

// BunkerReputation is an auto generated Go binding around an Ethereum contract.
type BunkerReputation struct {
	BunkerReputationCaller     // Read-only binding to the contract
	BunkerReputationTransactor // Write-only binding to the contract
	BunkerReputationFilterer   // Log filterer for contract events
}

// BunkerReputationCaller is an auto generated read-only Go binding around an Ethereum contract.
type BunkerReputationCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerReputationTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BunkerReputationTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerReputationFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BunkerReputationFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerReputationSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BunkerReputationSession struct {
	Contract     *BunkerReputation // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BunkerReputationCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BunkerReputationCallerSession struct {
	Contract *BunkerReputationCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// BunkerReputationTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BunkerReputationTransactorSession struct {
	Contract     *BunkerReputationTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// BunkerReputationRaw is an auto generated low-level Go binding around an Ethereum contract.
type BunkerReputationRaw struct {
	Contract *BunkerReputation // Generic contract binding to access the raw methods on
}

// BunkerReputationCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BunkerReputationCallerRaw struct {
	Contract *BunkerReputationCaller // Generic read-only contract binding to access the raw methods on
}

// BunkerReputationTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BunkerReputationTransactorRaw struct {
	Contract *BunkerReputationTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBunkerReputation creates a new instance of BunkerReputation, bound to a specific deployed contract.
func NewBunkerReputation(address common.Address, backend bind.ContractBackend) (*BunkerReputation, error) {
	contract, err := bindBunkerReputation(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BunkerReputation{BunkerReputationCaller: BunkerReputationCaller{contract: contract}, BunkerReputationTransactor: BunkerReputationTransactor{contract: contract}, BunkerReputationFilterer: BunkerReputationFilterer{contract: contract}}, nil
}

// NewBunkerReputationCaller creates a new read-only instance of BunkerReputation, bound to a specific deployed contract.
func NewBunkerReputationCaller(address common.Address, caller bind.ContractCaller) (*BunkerReputationCaller, error) {
	contract, err := bindBunkerReputation(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerReputationCaller{contract: contract}, nil
}

// NewBunkerReputationTransactor creates a new write-only instance of BunkerReputation, bound to a specific deployed contract.
func NewBunkerReputationTransactor(address common.Address, transactor bind.ContractTransactor) (*BunkerReputationTransactor, error) {
	contract, err := bindBunkerReputation(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerReputationTransactor{contract: contract}, nil
}

// NewBunkerReputationFilterer creates a new log filterer instance of BunkerReputation, bound to a specific deployed contract.
func NewBunkerReputationFilterer(address common.Address, filterer bind.ContractFilterer) (*BunkerReputationFilterer, error) {
	contract, err := bindBunkerReputation(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BunkerReputationFilterer{contract: contract}, nil
}

// bindBunkerReputation binds a generic wrapper to an already deployed contract.
func bindBunkerReputation(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BunkerReputationMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerReputation *BunkerReputationRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerReputation.Contract.BunkerReputationCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerReputation *BunkerReputationRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerReputation.Contract.BunkerReputationTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerReputation *BunkerReputationRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerReputation.Contract.BunkerReputationTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerReputation *BunkerReputationCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerReputation.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerReputation *BunkerReputationTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerReputation.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerReputation *BunkerReputationTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerReputation.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerReputation *BunkerReputationCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerReputation *BunkerReputationSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _BunkerReputation.Contract.DEFAULTADMINROLE(&_BunkerReputation.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerReputation *BunkerReputationCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _BunkerReputation.Contract.DEFAULTADMINROLE(&_BunkerReputation.CallOpts)
}

// INITIALSCORE is a free data retrieval call binding the contract method 0xde330373.
//
// Solidity: function INITIAL_SCORE() view returns(uint256)
func (_BunkerReputation *BunkerReputationCaller) INITIALSCORE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "INITIAL_SCORE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// INITIALSCORE is a free data retrieval call binding the contract method 0xde330373.
//
// Solidity: function INITIAL_SCORE() view returns(uint256)
func (_BunkerReputation *BunkerReputationSession) INITIALSCORE() (*big.Int, error) {
	return _BunkerReputation.Contract.INITIALSCORE(&_BunkerReputation.CallOpts)
}

// INITIALSCORE is a free data retrieval call binding the contract method 0xde330373.
//
// Solidity: function INITIAL_SCORE() view returns(uint256)
func (_BunkerReputation *BunkerReputationCallerSession) INITIALSCORE() (*big.Int, error) {
	return _BunkerReputation.Contract.INITIALSCORE(&_BunkerReputation.CallOpts)
}

// MAXSCORE is a free data retrieval call binding the contract method 0x27ff6223.
//
// Solidity: function MAX_SCORE() view returns(uint256)
func (_BunkerReputation *BunkerReputationCaller) MAXSCORE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "MAX_SCORE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXSCORE is a free data retrieval call binding the contract method 0x27ff6223.
//
// Solidity: function MAX_SCORE() view returns(uint256)
func (_BunkerReputation *BunkerReputationSession) MAXSCORE() (*big.Int, error) {
	return _BunkerReputation.Contract.MAXSCORE(&_BunkerReputation.CallOpts)
}

// MAXSCORE is a free data retrieval call binding the contract method 0x27ff6223.
//
// Solidity: function MAX_SCORE() view returns(uint256)
func (_BunkerReputation *BunkerReputationCallerSession) MAXSCORE() (*big.Int, error) {
	return _BunkerReputation.Contract.MAXSCORE(&_BunkerReputation.CallOpts)
}

// REPORTERROLE is a free data retrieval call binding the contract method 0x3f60d799.
//
// Solidity: function REPORTER_ROLE() view returns(bytes32)
func (_BunkerReputation *BunkerReputationCaller) REPORTERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "REPORTER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// REPORTERROLE is a free data retrieval call binding the contract method 0x3f60d799.
//
// Solidity: function REPORTER_ROLE() view returns(bytes32)
func (_BunkerReputation *BunkerReputationSession) REPORTERROLE() ([32]byte, error) {
	return _BunkerReputation.Contract.REPORTERROLE(&_BunkerReputation.CallOpts)
}

// REPORTERROLE is a free data retrieval call binding the contract method 0x3f60d799.
//
// Solidity: function REPORTER_ROLE() view returns(bytes32)
func (_BunkerReputation *BunkerReputationCallerSession) REPORTERROLE() ([32]byte, error) {
	return _BunkerReputation.Contract.REPORTERROLE(&_BunkerReputation.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerReputation *BunkerReputationCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerReputation *BunkerReputationSession) VERSION() (string, error) {
	return _BunkerReputation.Contract.VERSION(&_BunkerReputation.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerReputation *BunkerReputationCallerSession) VERSION() (string, error) {
	return _BunkerReputation.Contract.VERSION(&_BunkerReputation.CallOpts)
}

// WEEK is a free data retrieval call binding the contract method 0xf4359ce5.
//
// Solidity: function WEEK() view returns(uint256)
func (_BunkerReputation *BunkerReputationCaller) WEEK(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "WEEK")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WEEK is a free data retrieval call binding the contract method 0xf4359ce5.
//
// Solidity: function WEEK() view returns(uint256)
func (_BunkerReputation *BunkerReputationSession) WEEK() (*big.Int, error) {
	return _BunkerReputation.Contract.WEEK(&_BunkerReputation.CallOpts)
}

// WEEK is a free data retrieval call binding the contract method 0xf4359ce5.
//
// Solidity: function WEEK() view returns(uint256)
func (_BunkerReputation *BunkerReputationCallerSession) WEEK() (*big.Int, error) {
	return _BunkerReputation.Contract.WEEK(&_BunkerReputation.CallOpts)
}

// DecayFloor is a free data retrieval call binding the contract method 0x9861f58c.
//
// Solidity: function decayFloor() view returns(uint256)
func (_BunkerReputation *BunkerReputationCaller) DecayFloor(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "decayFloor")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DecayFloor is a free data retrieval call binding the contract method 0x9861f58c.
//
// Solidity: function decayFloor() view returns(uint256)
func (_BunkerReputation *BunkerReputationSession) DecayFloor() (*big.Int, error) {
	return _BunkerReputation.Contract.DecayFloor(&_BunkerReputation.CallOpts)
}

// DecayFloor is a free data retrieval call binding the contract method 0x9861f58c.
//
// Solidity: function decayFloor() view returns(uint256)
func (_BunkerReputation *BunkerReputationCallerSession) DecayFloor() (*big.Int, error) {
	return _BunkerReputation.Contract.DecayFloor(&_BunkerReputation.CallOpts)
}

// DecayRate is a free data retrieval call binding the contract method 0xa9c1f2f1.
//
// Solidity: function decayRate() view returns(uint256)
func (_BunkerReputation *BunkerReputationCaller) DecayRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "decayRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DecayRate is a free data retrieval call binding the contract method 0xa9c1f2f1.
//
// Solidity: function decayRate() view returns(uint256)
func (_BunkerReputation *BunkerReputationSession) DecayRate() (*big.Int, error) {
	return _BunkerReputation.Contract.DecayRate(&_BunkerReputation.CallOpts)
}

// DecayRate is a free data retrieval call binding the contract method 0xa9c1f2f1.
//
// Solidity: function decayRate() view returns(uint256)
func (_BunkerReputation *BunkerReputationCallerSession) DecayRate() (*big.Int, error) {
	return _BunkerReputation.Contract.DecayRate(&_BunkerReputation.CallOpts)
}

// GetReputation is a free data retrieval call binding the contract method 0x9c89a0e2.
//
// Solidity: function getReputation(address provider) view returns((uint32,uint32,uint32,uint32,uint48,uint48) data)
func (_BunkerReputation *BunkerReputationCaller) GetReputation(opts *bind.CallOpts, provider common.Address) (BunkerReputationReputationData, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "getReputation", provider)

	if err != nil {
		return *new(BunkerReputationReputationData), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerReputationReputationData)).(*BunkerReputationReputationData)

	return out0, err

}

// GetReputation is a free data retrieval call binding the contract method 0x9c89a0e2.
//
// Solidity: function getReputation(address provider) view returns((uint32,uint32,uint32,uint32,uint48,uint48) data)
func (_BunkerReputation *BunkerReputationSession) GetReputation(provider common.Address) (BunkerReputationReputationData, error) {
	return _BunkerReputation.Contract.GetReputation(&_BunkerReputation.CallOpts, provider)
}

// GetReputation is a free data retrieval call binding the contract method 0x9c89a0e2.
//
// Solidity: function getReputation(address provider) view returns((uint32,uint32,uint32,uint32,uint48,uint48) data)
func (_BunkerReputation *BunkerReputationCallerSession) GetReputation(provider common.Address) (BunkerReputationReputationData, error) {
	return _BunkerReputation.Contract.GetReputation(&_BunkerReputation.CallOpts, provider)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerReputation *BunkerReputationCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerReputation *BunkerReputationSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _BunkerReputation.Contract.GetRoleAdmin(&_BunkerReputation.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerReputation *BunkerReputationCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _BunkerReputation.Contract.GetRoleAdmin(&_BunkerReputation.CallOpts, role)
}

// GetScore is a free data retrieval call binding the contract method 0xd47875d0.
//
// Solidity: function getScore(address provider) view returns(uint32)
func (_BunkerReputation *BunkerReputationCaller) GetScore(opts *bind.CallOpts, provider common.Address) (uint32, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "getScore", provider)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetScore is a free data retrieval call binding the contract method 0xd47875d0.
//
// Solidity: function getScore(address provider) view returns(uint32)
func (_BunkerReputation *BunkerReputationSession) GetScore(provider common.Address) (uint32, error) {
	return _BunkerReputation.Contract.GetScore(&_BunkerReputation.CallOpts, provider)
}

// GetScore is a free data retrieval call binding the contract method 0xd47875d0.
//
// Solidity: function getScore(address provider) view returns(uint32)
func (_BunkerReputation *BunkerReputationCallerSession) GetScore(provider common.Address) (uint32, error) {
	return _BunkerReputation.Contract.GetScore(&_BunkerReputation.CallOpts, provider)
}

// GetTier is a free data retrieval call binding the contract method 0xb45aae52.
//
// Solidity: function getTier(address provider) view returns(uint8)
func (_BunkerReputation *BunkerReputationCaller) GetTier(opts *bind.CallOpts, provider common.Address) (uint8, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "getTier", provider)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetTier is a free data retrieval call binding the contract method 0xb45aae52.
//
// Solidity: function getTier(address provider) view returns(uint8)
func (_BunkerReputation *BunkerReputationSession) GetTier(provider common.Address) (uint8, error) {
	return _BunkerReputation.Contract.GetTier(&_BunkerReputation.CallOpts, provider)
}

// GetTier is a free data retrieval call binding the contract method 0xb45aae52.
//
// Solidity: function getTier(address provider) view returns(uint8)
func (_BunkerReputation *BunkerReputationCallerSession) GetTier(provider common.Address) (uint8, error) {
	return _BunkerReputation.Contract.GetTier(&_BunkerReputation.CallOpts, provider)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerReputation *BunkerReputationCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerReputation *BunkerReputationSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _BunkerReputation.Contract.HasRole(&_BunkerReputation.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerReputation *BunkerReputationCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _BunkerReputation.Contract.HasRole(&_BunkerReputation.CallOpts, role, account)
}

// HealthFailDelta is a free data retrieval call binding the contract method 0x364c3b63.
//
// Solidity: function healthFailDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCaller) HealthFailDelta(opts *bind.CallOpts) (int16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "healthFailDelta")

	if err != nil {
		return *new(int16), err
	}

	out0 := *abi.ConvertType(out[0], new(int16)).(*int16)

	return out0, err

}

// HealthFailDelta is a free data retrieval call binding the contract method 0x364c3b63.
//
// Solidity: function healthFailDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationSession) HealthFailDelta() (int16, error) {
	return _BunkerReputation.Contract.HealthFailDelta(&_BunkerReputation.CallOpts)
}

// HealthFailDelta is a free data retrieval call binding the contract method 0x364c3b63.
//
// Solidity: function healthFailDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCallerSession) HealthFailDelta() (int16, error) {
	return _BunkerReputation.Contract.HealthFailDelta(&_BunkerReputation.CallOpts)
}

// IsEligibleForJobs is a free data retrieval call binding the contract method 0x9d9718fe.
//
// Solidity: function isEligibleForJobs(address provider) view returns(bool eligible)
func (_BunkerReputation *BunkerReputationCaller) IsEligibleForJobs(opts *bind.CallOpts, provider common.Address) (bool, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "isEligibleForJobs", provider)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsEligibleForJobs is a free data retrieval call binding the contract method 0x9d9718fe.
//
// Solidity: function isEligibleForJobs(address provider) view returns(bool eligible)
func (_BunkerReputation *BunkerReputationSession) IsEligibleForJobs(provider common.Address) (bool, error) {
	return _BunkerReputation.Contract.IsEligibleForJobs(&_BunkerReputation.CallOpts, provider)
}

// IsEligibleForJobs is a free data retrieval call binding the contract method 0x9d9718fe.
//
// Solidity: function isEligibleForJobs(address provider) view returns(bool eligible)
func (_BunkerReputation *BunkerReputationCallerSession) IsEligibleForJobs(provider common.Address) (bool, error) {
	return _BunkerReputation.Contract.IsEligibleForJobs(&_BunkerReputation.CallOpts, provider)
}

// JobCompletedDelta is a free data retrieval call binding the contract method 0x0b67ca75.
//
// Solidity: function jobCompletedDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCaller) JobCompletedDelta(opts *bind.CallOpts) (int16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "jobCompletedDelta")

	if err != nil {
		return *new(int16), err
	}

	out0 := *abi.ConvertType(out[0], new(int16)).(*int16)

	return out0, err

}

// JobCompletedDelta is a free data retrieval call binding the contract method 0x0b67ca75.
//
// Solidity: function jobCompletedDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationSession) JobCompletedDelta() (int16, error) {
	return _BunkerReputation.Contract.JobCompletedDelta(&_BunkerReputation.CallOpts)
}

// JobCompletedDelta is a free data retrieval call binding the contract method 0x0b67ca75.
//
// Solidity: function jobCompletedDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCallerSession) JobCompletedDelta() (int16, error) {
	return _BunkerReputation.Contract.JobCompletedDelta(&_BunkerReputation.CallOpts)
}

// JobEarlyDelta is a free data retrieval call binding the contract method 0xc3ae1c30.
//
// Solidity: function jobEarlyDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCaller) JobEarlyDelta(opts *bind.CallOpts) (int16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "jobEarlyDelta")

	if err != nil {
		return *new(int16), err
	}

	out0 := *abi.ConvertType(out[0], new(int16)).(*int16)

	return out0, err

}

// JobEarlyDelta is a free data retrieval call binding the contract method 0xc3ae1c30.
//
// Solidity: function jobEarlyDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationSession) JobEarlyDelta() (int16, error) {
	return _BunkerReputation.Contract.JobEarlyDelta(&_BunkerReputation.CallOpts)
}

// JobEarlyDelta is a free data retrieval call binding the contract method 0xc3ae1c30.
//
// Solidity: function jobEarlyDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCallerSession) JobEarlyDelta() (int16, error) {
	return _BunkerReputation.Contract.JobEarlyDelta(&_BunkerReputation.CallOpts)
}

// JobTimeoutDelta is a free data retrieval call binding the contract method 0xacb5fe0f.
//
// Solidity: function jobTimeoutDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCaller) JobTimeoutDelta(opts *bind.CallOpts) (int16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "jobTimeoutDelta")

	if err != nil {
		return *new(int16), err
	}

	out0 := *abi.ConvertType(out[0], new(int16)).(*int16)

	return out0, err

}

// JobTimeoutDelta is a free data retrieval call binding the contract method 0xacb5fe0f.
//
// Solidity: function jobTimeoutDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationSession) JobTimeoutDelta() (int16, error) {
	return _BunkerReputation.Contract.JobTimeoutDelta(&_BunkerReputation.CallOpts)
}

// JobTimeoutDelta is a free data retrieval call binding the contract method 0xacb5fe0f.
//
// Solidity: function jobTimeoutDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCallerSession) JobTimeoutDelta() (int16, error) {
	return _BunkerReputation.Contract.JobTimeoutDelta(&_BunkerReputation.CallOpts)
}

// MaxCustomDelta is a free data retrieval call binding the contract method 0xaaee8cfd.
//
// Solidity: function maxCustomDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCaller) MaxCustomDelta(opts *bind.CallOpts) (int16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "maxCustomDelta")

	if err != nil {
		return *new(int16), err
	}

	out0 := *abi.ConvertType(out[0], new(int16)).(*int16)

	return out0, err

}

// MaxCustomDelta is a free data retrieval call binding the contract method 0xaaee8cfd.
//
// Solidity: function maxCustomDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationSession) MaxCustomDelta() (int16, error) {
	return _BunkerReputation.Contract.MaxCustomDelta(&_BunkerReputation.CallOpts)
}

// MaxCustomDelta is a free data retrieval call binding the contract method 0xaaee8cfd.
//
// Solidity: function maxCustomDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCallerSession) MaxCustomDelta() (int16, error) {
	return _BunkerReputation.Contract.MaxCustomDelta(&_BunkerReputation.CallOpts)
}

// MinCustomDelta is a free data retrieval call binding the contract method 0xc7b71111.
//
// Solidity: function minCustomDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCaller) MinCustomDelta(opts *bind.CallOpts) (int16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "minCustomDelta")

	if err != nil {
		return *new(int16), err
	}

	out0 := *abi.ConvertType(out[0], new(int16)).(*int16)

	return out0, err

}

// MinCustomDelta is a free data retrieval call binding the contract method 0xc7b71111.
//
// Solidity: function minCustomDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationSession) MinCustomDelta() (int16, error) {
	return _BunkerReputation.Contract.MinCustomDelta(&_BunkerReputation.CallOpts)
}

// MinCustomDelta is a free data retrieval call binding the contract method 0xc7b71111.
//
// Solidity: function minCustomDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCallerSession) MinCustomDelta() (int16, error) {
	return _BunkerReputation.Contract.MinCustomDelta(&_BunkerReputation.CallOpts)
}

// MinScoreForJobs is a free data retrieval call binding the contract method 0x95519b96.
//
// Solidity: function minScoreForJobs() view returns(uint256)
func (_BunkerReputation *BunkerReputationCaller) MinScoreForJobs(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "minScoreForJobs")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinScoreForJobs is a free data retrieval call binding the contract method 0x95519b96.
//
// Solidity: function minScoreForJobs() view returns(uint256)
func (_BunkerReputation *BunkerReputationSession) MinScoreForJobs() (*big.Int, error) {
	return _BunkerReputation.Contract.MinScoreForJobs(&_BunkerReputation.CallOpts)
}

// MinScoreForJobs is a free data retrieval call binding the contract method 0x95519b96.
//
// Solidity: function minScoreForJobs() view returns(uint256)
func (_BunkerReputation *BunkerReputationCallerSession) MinScoreForJobs() (*big.Int, error) {
	return _BunkerReputation.Contract.MinScoreForJobs(&_BunkerReputation.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerReputation *BunkerReputationCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerReputation *BunkerReputationSession) Owner() (common.Address, error) {
	return _BunkerReputation.Contract.Owner(&_BunkerReputation.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerReputation *BunkerReputationCallerSession) Owner() (common.Address, error) {
	return _BunkerReputation.Contract.Owner(&_BunkerReputation.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerReputation *BunkerReputationCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerReputation *BunkerReputationSession) PendingOwner() (common.Address, error) {
	return _BunkerReputation.Contract.PendingOwner(&_BunkerReputation.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerReputation *BunkerReputationCallerSession) PendingOwner() (common.Address, error) {
	return _BunkerReputation.Contract.PendingOwner(&_BunkerReputation.CallOpts)
}

// PerfectUptimeDelta is a free data retrieval call binding the contract method 0x684e30fd.
//
// Solidity: function perfectUptimeDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCaller) PerfectUptimeDelta(opts *bind.CallOpts) (int16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "perfectUptimeDelta")

	if err != nil {
		return *new(int16), err
	}

	out0 := *abi.ConvertType(out[0], new(int16)).(*int16)

	return out0, err

}

// PerfectUptimeDelta is a free data retrieval call binding the contract method 0x684e30fd.
//
// Solidity: function perfectUptimeDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationSession) PerfectUptimeDelta() (int16, error) {
	return _BunkerReputation.Contract.PerfectUptimeDelta(&_BunkerReputation.CallOpts)
}

// PerfectUptimeDelta is a free data retrieval call binding the contract method 0x684e30fd.
//
// Solidity: function perfectUptimeDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCallerSession) PerfectUptimeDelta() (int16, error) {
	return _BunkerReputation.Contract.PerfectUptimeDelta(&_BunkerReputation.CallOpts)
}

// ReplicaMismatchDelta is a free data retrieval call binding the contract method 0xfade295d.
//
// Solidity: function replicaMismatchDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCaller) ReplicaMismatchDelta(opts *bind.CallOpts) (int16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "replicaMismatchDelta")

	if err != nil {
		return *new(int16), err
	}

	out0 := *abi.ConvertType(out[0], new(int16)).(*int16)

	return out0, err

}

// ReplicaMismatchDelta is a free data retrieval call binding the contract method 0xfade295d.
//
// Solidity: function replicaMismatchDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationSession) ReplicaMismatchDelta() (int16, error) {
	return _BunkerReputation.Contract.ReplicaMismatchDelta(&_BunkerReputation.CallOpts)
}

// ReplicaMismatchDelta is a free data retrieval call binding the contract method 0xfade295d.
//
// Solidity: function replicaMismatchDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCallerSession) ReplicaMismatchDelta() (int16, error) {
	return _BunkerReputation.Contract.ReplicaMismatchDelta(&_BunkerReputation.CallOpts)
}

// Reputations is a free data retrieval call binding the contract method 0x06fa15a7.
//
// Solidity: function reputations(address ) view returns(uint32 score, uint32 jobsCompleted, uint32 jobsFailed, uint32 slashCount, uint48 lastUpdated, uint48 registeredAt)
func (_BunkerReputation *BunkerReputationCaller) Reputations(opts *bind.CallOpts, arg0 common.Address) (struct {
	Score         uint32
	JobsCompleted uint32
	JobsFailed    uint32
	SlashCount    uint32
	LastUpdated   *big.Int
	RegisteredAt  *big.Int
}, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "reputations", arg0)

	outstruct := new(struct {
		Score         uint32
		JobsCompleted uint32
		JobsFailed    uint32
		SlashCount    uint32
		LastUpdated   *big.Int
		RegisteredAt  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Score = *abi.ConvertType(out[0], new(uint32)).(*uint32)
	outstruct.JobsCompleted = *abi.ConvertType(out[1], new(uint32)).(*uint32)
	outstruct.JobsFailed = *abi.ConvertType(out[2], new(uint32)).(*uint32)
	outstruct.SlashCount = *abi.ConvertType(out[3], new(uint32)).(*uint32)
	outstruct.LastUpdated = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.RegisteredAt = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Reputations is a free data retrieval call binding the contract method 0x06fa15a7.
//
// Solidity: function reputations(address ) view returns(uint32 score, uint32 jobsCompleted, uint32 jobsFailed, uint32 slashCount, uint48 lastUpdated, uint48 registeredAt)
func (_BunkerReputation *BunkerReputationSession) Reputations(arg0 common.Address) (struct {
	Score         uint32
	JobsCompleted uint32
	JobsFailed    uint32
	SlashCount    uint32
	LastUpdated   *big.Int
	RegisteredAt  *big.Int
}, error) {
	return _BunkerReputation.Contract.Reputations(&_BunkerReputation.CallOpts, arg0)
}

// Reputations is a free data retrieval call binding the contract method 0x06fa15a7.
//
// Solidity: function reputations(address ) view returns(uint32 score, uint32 jobsCompleted, uint32 jobsFailed, uint32 slashCount, uint48 lastUpdated, uint48 registeredAt)
func (_BunkerReputation *BunkerReputationCallerSession) Reputations(arg0 common.Address) (struct {
	Score         uint32
	JobsCompleted uint32
	JobsFailed    uint32
	SlashCount    uint32
	LastUpdated   *big.Int
	RegisteredAt  *big.Int
}, error) {
	return _BunkerReputation.Contract.Reputations(&_BunkerReputation.CallOpts, arg0)
}

// SecurityViolationDelta is a free data retrieval call binding the contract method 0xd4b378bf.
//
// Solidity: function securityViolationDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCaller) SecurityViolationDelta(opts *bind.CallOpts) (int16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "securityViolationDelta")

	if err != nil {
		return *new(int16), err
	}

	out0 := *abi.ConvertType(out[0], new(int16)).(*int16)

	return out0, err

}

// SecurityViolationDelta is a free data retrieval call binding the contract method 0xd4b378bf.
//
// Solidity: function securityViolationDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationSession) SecurityViolationDelta() (int16, error) {
	return _BunkerReputation.Contract.SecurityViolationDelta(&_BunkerReputation.CallOpts)
}

// SecurityViolationDelta is a free data retrieval call binding the contract method 0xd4b378bf.
//
// Solidity: function securityViolationDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCallerSession) SecurityViolationDelta() (int16, error) {
	return _BunkerReputation.Contract.SecurityViolationDelta(&_BunkerReputation.CallOpts)
}

// SlashEventDelta is a free data retrieval call binding the contract method 0x8bf51c2c.
//
// Solidity: function slashEventDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCaller) SlashEventDelta(opts *bind.CallOpts) (int16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "slashEventDelta")

	if err != nil {
		return *new(int16), err
	}

	out0 := *abi.ConvertType(out[0], new(int16)).(*int16)

	return out0, err

}

// SlashEventDelta is a free data retrieval call binding the contract method 0x8bf51c2c.
//
// Solidity: function slashEventDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationSession) SlashEventDelta() (int16, error) {
	return _BunkerReputation.Contract.SlashEventDelta(&_BunkerReputation.CallOpts)
}

// SlashEventDelta is a free data retrieval call binding the contract method 0x8bf51c2c.
//
// Solidity: function slashEventDelta() view returns(int16)
func (_BunkerReputation *BunkerReputationCallerSession) SlashEventDelta() (int16, error) {
	return _BunkerReputation.Contract.SlashEventDelta(&_BunkerReputation.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerReputation *BunkerReputationCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerReputation *BunkerReputationSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _BunkerReputation.Contract.SupportsInterface(&_BunkerReputation.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerReputation *BunkerReputationCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _BunkerReputation.Contract.SupportsInterface(&_BunkerReputation.CallOpts, interfaceId)
}

// TierElite is a free data retrieval call binding the contract method 0xa240b04f.
//
// Solidity: function tierElite() view returns(uint16)
func (_BunkerReputation *BunkerReputationCaller) TierElite(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "tierElite")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// TierElite is a free data retrieval call binding the contract method 0xa240b04f.
//
// Solidity: function tierElite() view returns(uint16)
func (_BunkerReputation *BunkerReputationSession) TierElite() (uint16, error) {
	return _BunkerReputation.Contract.TierElite(&_BunkerReputation.CallOpts)
}

// TierElite is a free data retrieval call binding the contract method 0xa240b04f.
//
// Solidity: function tierElite() view returns(uint16)
func (_BunkerReputation *BunkerReputationCallerSession) TierElite() (uint16, error) {
	return _BunkerReputation.Contract.TierElite(&_BunkerReputation.CallOpts)
}

// TierProbation is a free data retrieval call binding the contract method 0xad74b060.
//
// Solidity: function tierProbation() view returns(uint16)
func (_BunkerReputation *BunkerReputationCaller) TierProbation(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "tierProbation")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// TierProbation is a free data retrieval call binding the contract method 0xad74b060.
//
// Solidity: function tierProbation() view returns(uint16)
func (_BunkerReputation *BunkerReputationSession) TierProbation() (uint16, error) {
	return _BunkerReputation.Contract.TierProbation(&_BunkerReputation.CallOpts)
}

// TierProbation is a free data retrieval call binding the contract method 0xad74b060.
//
// Solidity: function tierProbation() view returns(uint16)
func (_BunkerReputation *BunkerReputationCallerSession) TierProbation() (uint16, error) {
	return _BunkerReputation.Contract.TierProbation(&_BunkerReputation.CallOpts)
}

// TierStandard is a free data retrieval call binding the contract method 0xc233b616.
//
// Solidity: function tierStandard() view returns(uint16)
func (_BunkerReputation *BunkerReputationCaller) TierStandard(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "tierStandard")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// TierStandard is a free data retrieval call binding the contract method 0xc233b616.
//
// Solidity: function tierStandard() view returns(uint16)
func (_BunkerReputation *BunkerReputationSession) TierStandard() (uint16, error) {
	return _BunkerReputation.Contract.TierStandard(&_BunkerReputation.CallOpts)
}

// TierStandard is a free data retrieval call binding the contract method 0xc233b616.
//
// Solidity: function tierStandard() view returns(uint16)
func (_BunkerReputation *BunkerReputationCallerSession) TierStandard() (uint16, error) {
	return _BunkerReputation.Contract.TierStandard(&_BunkerReputation.CallOpts)
}

// TierTrusted is a free data retrieval call binding the contract method 0x125f377a.
//
// Solidity: function tierTrusted() view returns(uint16)
func (_BunkerReputation *BunkerReputationCaller) TierTrusted(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _BunkerReputation.contract.Call(opts, &out, "tierTrusted")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// TierTrusted is a free data retrieval call binding the contract method 0x125f377a.
//
// Solidity: function tierTrusted() view returns(uint16)
func (_BunkerReputation *BunkerReputationSession) TierTrusted() (uint16, error) {
	return _BunkerReputation.Contract.TierTrusted(&_BunkerReputation.CallOpts)
}

// TierTrusted is a free data retrieval call binding the contract method 0x125f377a.
//
// Solidity: function tierTrusted() view returns(uint16)
func (_BunkerReputation *BunkerReputationCallerSession) TierTrusted() (uint16, error) {
	return _BunkerReputation.Contract.TierTrusted(&_BunkerReputation.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerReputation *BunkerReputationTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerReputation *BunkerReputationSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerReputation.Contract.AcceptOwnership(&_BunkerReputation.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerReputation *BunkerReputationTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerReputation.Contract.AcceptOwnership(&_BunkerReputation.TransactOpts)
}

// ApplyDecay is a paid mutator transaction binding the contract method 0xedfcc437.
//
// Solidity: function applyDecay(address provider) returns()
func (_BunkerReputation *BunkerReputationTransactor) ApplyDecay(opts *bind.TransactOpts, provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "applyDecay", provider)
}

// ApplyDecay is a paid mutator transaction binding the contract method 0xedfcc437.
//
// Solidity: function applyDecay(address provider) returns()
func (_BunkerReputation *BunkerReputationSession) ApplyDecay(provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.ApplyDecay(&_BunkerReputation.TransactOpts, provider)
}

// ApplyDecay is a paid mutator transaction binding the contract method 0xedfcc437.
//
// Solidity: function applyDecay(address provider) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) ApplyDecay(provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.ApplyDecay(&_BunkerReputation.TransactOpts, provider)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerReputation *BunkerReputationTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerReputation *BunkerReputationSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.GrantRole(&_BunkerReputation.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.GrantRole(&_BunkerReputation.TransactOpts, role, account)
}

// RecordEvent is a paid mutator transaction binding the contract method 0x68cad238.
//
// Solidity: function recordEvent(address provider, int16 delta, string reason) returns()
func (_BunkerReputation *BunkerReputationTransactor) RecordEvent(opts *bind.TransactOpts, provider common.Address, delta int16, reason string) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "recordEvent", provider, delta, reason)
}

// RecordEvent is a paid mutator transaction binding the contract method 0x68cad238.
//
// Solidity: function recordEvent(address provider, int16 delta, string reason) returns()
func (_BunkerReputation *BunkerReputationSession) RecordEvent(provider common.Address, delta int16, reason string) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RecordEvent(&_BunkerReputation.TransactOpts, provider, delta, reason)
}

// RecordEvent is a paid mutator transaction binding the contract method 0x68cad238.
//
// Solidity: function recordEvent(address provider, int16 delta, string reason) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) RecordEvent(provider common.Address, delta int16, reason string) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RecordEvent(&_BunkerReputation.TransactOpts, provider, delta, reason)
}

// RecordJobCompleted is a paid mutator transaction binding the contract method 0xc53fba6d.
//
// Solidity: function recordJobCompleted(address provider) returns()
func (_BunkerReputation *BunkerReputationTransactor) RecordJobCompleted(opts *bind.TransactOpts, provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "recordJobCompleted", provider)
}

// RecordJobCompleted is a paid mutator transaction binding the contract method 0xc53fba6d.
//
// Solidity: function recordJobCompleted(address provider) returns()
func (_BunkerReputation *BunkerReputationSession) RecordJobCompleted(provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RecordJobCompleted(&_BunkerReputation.TransactOpts, provider)
}

// RecordJobCompleted is a paid mutator transaction binding the contract method 0xc53fba6d.
//
// Solidity: function recordJobCompleted(address provider) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) RecordJobCompleted(provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RecordJobCompleted(&_BunkerReputation.TransactOpts, provider)
}

// RecordJobFailed is a paid mutator transaction binding the contract method 0x1dd98b7e.
//
// Solidity: function recordJobFailed(address provider) returns()
func (_BunkerReputation *BunkerReputationTransactor) RecordJobFailed(opts *bind.TransactOpts, provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "recordJobFailed", provider)
}

// RecordJobFailed is a paid mutator transaction binding the contract method 0x1dd98b7e.
//
// Solidity: function recordJobFailed(address provider) returns()
func (_BunkerReputation *BunkerReputationSession) RecordJobFailed(provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RecordJobFailed(&_BunkerReputation.TransactOpts, provider)
}

// RecordJobFailed is a paid mutator transaction binding the contract method 0x1dd98b7e.
//
// Solidity: function recordJobFailed(address provider) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) RecordJobFailed(provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RecordJobFailed(&_BunkerReputation.TransactOpts, provider)
}

// RecordSlashEvent is a paid mutator transaction binding the contract method 0x93198fd0.
//
// Solidity: function recordSlashEvent(address provider) returns()
func (_BunkerReputation *BunkerReputationTransactor) RecordSlashEvent(opts *bind.TransactOpts, provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "recordSlashEvent", provider)
}

// RecordSlashEvent is a paid mutator transaction binding the contract method 0x93198fd0.
//
// Solidity: function recordSlashEvent(address provider) returns()
func (_BunkerReputation *BunkerReputationSession) RecordSlashEvent(provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RecordSlashEvent(&_BunkerReputation.TransactOpts, provider)
}

// RecordSlashEvent is a paid mutator transaction binding the contract method 0x93198fd0.
//
// Solidity: function recordSlashEvent(address provider) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) RecordSlashEvent(provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RecordSlashEvent(&_BunkerReputation.TransactOpts, provider)
}

// RegisterProvider is a paid mutator transaction binding the contract method 0x0e260016.
//
// Solidity: function registerProvider(address provider) returns()
func (_BunkerReputation *BunkerReputationTransactor) RegisterProvider(opts *bind.TransactOpts, provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "registerProvider", provider)
}

// RegisterProvider is a paid mutator transaction binding the contract method 0x0e260016.
//
// Solidity: function registerProvider(address provider) returns()
func (_BunkerReputation *BunkerReputationSession) RegisterProvider(provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RegisterProvider(&_BunkerReputation.TransactOpts, provider)
}

// RegisterProvider is a paid mutator transaction binding the contract method 0x0e260016.
//
// Solidity: function registerProvider(address provider) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) RegisterProvider(provider common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RegisterProvider(&_BunkerReputation.TransactOpts, provider)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerReputation *BunkerReputationTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerReputation *BunkerReputationSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerReputation.Contract.RenounceOwnership(&_BunkerReputation.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerReputation *BunkerReputationTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerReputation.Contract.RenounceOwnership(&_BunkerReputation.TransactOpts)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerReputation *BunkerReputationTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerReputation *BunkerReputationSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RenounceRole(&_BunkerReputation.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RenounceRole(&_BunkerReputation.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerReputation *BunkerReputationTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerReputation *BunkerReputationSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RevokeRole(&_BunkerReputation.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.RevokeRole(&_BunkerReputation.TransactOpts, role, account)
}

// SetDecayParams is a paid mutator transaction binding the contract method 0xb0b88d74.
//
// Solidity: function setDecayParams(uint256 rate, uint256 floor) returns()
func (_BunkerReputation *BunkerReputationTransactor) SetDecayParams(opts *bind.TransactOpts, rate *big.Int, floor *big.Int) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "setDecayParams", rate, floor)
}

// SetDecayParams is a paid mutator transaction binding the contract method 0xb0b88d74.
//
// Solidity: function setDecayParams(uint256 rate, uint256 floor) returns()
func (_BunkerReputation *BunkerReputationSession) SetDecayParams(rate *big.Int, floor *big.Int) (*types.Transaction, error) {
	return _BunkerReputation.Contract.SetDecayParams(&_BunkerReputation.TransactOpts, rate, floor)
}

// SetDecayParams is a paid mutator transaction binding the contract method 0xb0b88d74.
//
// Solidity: function setDecayParams(uint256 rate, uint256 floor) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) SetDecayParams(rate *big.Int, floor *big.Int) (*types.Transaction, error) {
	return _BunkerReputation.Contract.SetDecayParams(&_BunkerReputation.TransactOpts, rate, floor)
}

// SetDeltaParameters is a paid mutator transaction binding the contract method 0x579c28e6.
//
// Solidity: function setDeltaParameters(int16 _jobCompletedDelta, int16 _jobEarlyDelta, int16 _perfectUptimeDelta, int16 _jobTimeoutDelta, int16 _healthFailDelta, int16 _replicaMismatchDelta, int16 _slashEventDelta, int16 _securityViolationDelta) returns()
func (_BunkerReputation *BunkerReputationTransactor) SetDeltaParameters(opts *bind.TransactOpts, _jobCompletedDelta int16, _jobEarlyDelta int16, _perfectUptimeDelta int16, _jobTimeoutDelta int16, _healthFailDelta int16, _replicaMismatchDelta int16, _slashEventDelta int16, _securityViolationDelta int16) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "setDeltaParameters", _jobCompletedDelta, _jobEarlyDelta, _perfectUptimeDelta, _jobTimeoutDelta, _healthFailDelta, _replicaMismatchDelta, _slashEventDelta, _securityViolationDelta)
}

// SetDeltaParameters is a paid mutator transaction binding the contract method 0x579c28e6.
//
// Solidity: function setDeltaParameters(int16 _jobCompletedDelta, int16 _jobEarlyDelta, int16 _perfectUptimeDelta, int16 _jobTimeoutDelta, int16 _healthFailDelta, int16 _replicaMismatchDelta, int16 _slashEventDelta, int16 _securityViolationDelta) returns()
func (_BunkerReputation *BunkerReputationSession) SetDeltaParameters(_jobCompletedDelta int16, _jobEarlyDelta int16, _perfectUptimeDelta int16, _jobTimeoutDelta int16, _healthFailDelta int16, _replicaMismatchDelta int16, _slashEventDelta int16, _securityViolationDelta int16) (*types.Transaction, error) {
	return _BunkerReputation.Contract.SetDeltaParameters(&_BunkerReputation.TransactOpts, _jobCompletedDelta, _jobEarlyDelta, _perfectUptimeDelta, _jobTimeoutDelta, _healthFailDelta, _replicaMismatchDelta, _slashEventDelta, _securityViolationDelta)
}

// SetDeltaParameters is a paid mutator transaction binding the contract method 0x579c28e6.
//
// Solidity: function setDeltaParameters(int16 _jobCompletedDelta, int16 _jobEarlyDelta, int16 _perfectUptimeDelta, int16 _jobTimeoutDelta, int16 _healthFailDelta, int16 _replicaMismatchDelta, int16 _slashEventDelta, int16 _securityViolationDelta) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) SetDeltaParameters(_jobCompletedDelta int16, _jobEarlyDelta int16, _perfectUptimeDelta int16, _jobTimeoutDelta int16, _healthFailDelta int16, _replicaMismatchDelta int16, _slashEventDelta int16, _securityViolationDelta int16) (*types.Transaction, error) {
	return _BunkerReputation.Contract.SetDeltaParameters(&_BunkerReputation.TransactOpts, _jobCompletedDelta, _jobEarlyDelta, _perfectUptimeDelta, _jobTimeoutDelta, _healthFailDelta, _replicaMismatchDelta, _slashEventDelta, _securityViolationDelta)
}

// SetMaxCustomDelta is a paid mutator transaction binding the contract method 0x4dd3c431.
//
// Solidity: function setMaxCustomDelta(int16 maxDelta) returns()
func (_BunkerReputation *BunkerReputationTransactor) SetMaxCustomDelta(opts *bind.TransactOpts, maxDelta int16) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "setMaxCustomDelta", maxDelta)
}

// SetMaxCustomDelta is a paid mutator transaction binding the contract method 0x4dd3c431.
//
// Solidity: function setMaxCustomDelta(int16 maxDelta) returns()
func (_BunkerReputation *BunkerReputationSession) SetMaxCustomDelta(maxDelta int16) (*types.Transaction, error) {
	return _BunkerReputation.Contract.SetMaxCustomDelta(&_BunkerReputation.TransactOpts, maxDelta)
}

// SetMaxCustomDelta is a paid mutator transaction binding the contract method 0x4dd3c431.
//
// Solidity: function setMaxCustomDelta(int16 maxDelta) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) SetMaxCustomDelta(maxDelta int16) (*types.Transaction, error) {
	return _BunkerReputation.Contract.SetMaxCustomDelta(&_BunkerReputation.TransactOpts, maxDelta)
}

// SetMinScoreForJobs is a paid mutator transaction binding the contract method 0x7aa4d004.
//
// Solidity: function setMinScoreForJobs(uint256 _minScore) returns()
func (_BunkerReputation *BunkerReputationTransactor) SetMinScoreForJobs(opts *bind.TransactOpts, _minScore *big.Int) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "setMinScoreForJobs", _minScore)
}

// SetMinScoreForJobs is a paid mutator transaction binding the contract method 0x7aa4d004.
//
// Solidity: function setMinScoreForJobs(uint256 _minScore) returns()
func (_BunkerReputation *BunkerReputationSession) SetMinScoreForJobs(_minScore *big.Int) (*types.Transaction, error) {
	return _BunkerReputation.Contract.SetMinScoreForJobs(&_BunkerReputation.TransactOpts, _minScore)
}

// SetMinScoreForJobs is a paid mutator transaction binding the contract method 0x7aa4d004.
//
// Solidity: function setMinScoreForJobs(uint256 _minScore) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) SetMinScoreForJobs(_minScore *big.Int) (*types.Transaction, error) {
	return _BunkerReputation.Contract.SetMinScoreForJobs(&_BunkerReputation.TransactOpts, _minScore)
}

// SetTierThresholds is a paid mutator transaction binding the contract method 0xdcd28478.
//
// Solidity: function setTierThresholds(uint16 probation, uint16 standard, uint16 trusted, uint16 elite) returns()
func (_BunkerReputation *BunkerReputationTransactor) SetTierThresholds(opts *bind.TransactOpts, probation uint16, standard uint16, trusted uint16, elite uint16) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "setTierThresholds", probation, standard, trusted, elite)
}

// SetTierThresholds is a paid mutator transaction binding the contract method 0xdcd28478.
//
// Solidity: function setTierThresholds(uint16 probation, uint16 standard, uint16 trusted, uint16 elite) returns()
func (_BunkerReputation *BunkerReputationSession) SetTierThresholds(probation uint16, standard uint16, trusted uint16, elite uint16) (*types.Transaction, error) {
	return _BunkerReputation.Contract.SetTierThresholds(&_BunkerReputation.TransactOpts, probation, standard, trusted, elite)
}

// SetTierThresholds is a paid mutator transaction binding the contract method 0xdcd28478.
//
// Solidity: function setTierThresholds(uint16 probation, uint16 standard, uint16 trusted, uint16 elite) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) SetTierThresholds(probation uint16, standard uint16, trusted uint16, elite uint16) (*types.Transaction, error) {
	return _BunkerReputation.Contract.SetTierThresholds(&_BunkerReputation.TransactOpts, probation, standard, trusted, elite)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerReputation *BunkerReputationTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BunkerReputation.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerReputation *BunkerReputationSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.TransferOwnership(&_BunkerReputation.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerReputation *BunkerReputationTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerReputation.Contract.TransferOwnership(&_BunkerReputation.TransactOpts, newOwner)
}

// BunkerReputationDecayParamsUpdatedIterator is returned from FilterDecayParamsUpdated and is used to iterate over the raw logs and unpacked data for DecayParamsUpdated events raised by the BunkerReputation contract.
type BunkerReputationDecayParamsUpdatedIterator struct {
	Event *BunkerReputationDecayParamsUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerReputationDecayParamsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationDecayParamsUpdated)
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
		it.Event = new(BunkerReputationDecayParamsUpdated)
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
func (it *BunkerReputationDecayParamsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationDecayParamsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationDecayParamsUpdated represents a DecayParamsUpdated event raised by the BunkerReputation contract.
type BunkerReputationDecayParamsUpdated struct {
	Rate  *big.Int
	Floor *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDecayParamsUpdated is a free log retrieval operation binding the contract event 0x8fc7937c0b577a9b967c8027dda6d2507bc421730a34aa1f323d475c8674ee49.
//
// Solidity: event DecayParamsUpdated(uint256 rate, uint256 floor)
func (_BunkerReputation *BunkerReputationFilterer) FilterDecayParamsUpdated(opts *bind.FilterOpts) (*BunkerReputationDecayParamsUpdatedIterator, error) {

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "DecayParamsUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerReputationDecayParamsUpdatedIterator{contract: _BunkerReputation.contract, event: "DecayParamsUpdated", logs: logs, sub: sub}, nil
}

// WatchDecayParamsUpdated is a free log subscription operation binding the contract event 0x8fc7937c0b577a9b967c8027dda6d2507bc421730a34aa1f323d475c8674ee49.
//
// Solidity: event DecayParamsUpdated(uint256 rate, uint256 floor)
func (_BunkerReputation *BunkerReputationFilterer) WatchDecayParamsUpdated(opts *bind.WatchOpts, sink chan<- *BunkerReputationDecayParamsUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "DecayParamsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationDecayParamsUpdated)
				if err := _BunkerReputation.contract.UnpackLog(event, "DecayParamsUpdated", log); err != nil {
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

// ParseDecayParamsUpdated is a log parse operation binding the contract event 0x8fc7937c0b577a9b967c8027dda6d2507bc421730a34aa1f323d475c8674ee49.
//
// Solidity: event DecayParamsUpdated(uint256 rate, uint256 floor)
func (_BunkerReputation *BunkerReputationFilterer) ParseDecayParamsUpdated(log types.Log) (*BunkerReputationDecayParamsUpdated, error) {
	event := new(BunkerReputationDecayParamsUpdated)
	if err := _BunkerReputation.contract.UnpackLog(event, "DecayParamsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationDeltaParametersUpdatedIterator is returned from FilterDeltaParametersUpdated and is used to iterate over the raw logs and unpacked data for DeltaParametersUpdated events raised by the BunkerReputation contract.
type BunkerReputationDeltaParametersUpdatedIterator struct {
	Event *BunkerReputationDeltaParametersUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerReputationDeltaParametersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationDeltaParametersUpdated)
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
		it.Event = new(BunkerReputationDeltaParametersUpdated)
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
func (it *BunkerReputationDeltaParametersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationDeltaParametersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationDeltaParametersUpdated represents a DeltaParametersUpdated event raised by the BunkerReputation contract.
type BunkerReputationDeltaParametersUpdated struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterDeltaParametersUpdated is a free log retrieval operation binding the contract event 0xbbe3d3400bbdd1c673beee5e0de381892cf0afcab96f92e98a7ab105f9bb6deb.
//
// Solidity: event DeltaParametersUpdated()
func (_BunkerReputation *BunkerReputationFilterer) FilterDeltaParametersUpdated(opts *bind.FilterOpts) (*BunkerReputationDeltaParametersUpdatedIterator, error) {

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "DeltaParametersUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerReputationDeltaParametersUpdatedIterator{contract: _BunkerReputation.contract, event: "DeltaParametersUpdated", logs: logs, sub: sub}, nil
}

// WatchDeltaParametersUpdated is a free log subscription operation binding the contract event 0xbbe3d3400bbdd1c673beee5e0de381892cf0afcab96f92e98a7ab105f9bb6deb.
//
// Solidity: event DeltaParametersUpdated()
func (_BunkerReputation *BunkerReputationFilterer) WatchDeltaParametersUpdated(opts *bind.WatchOpts, sink chan<- *BunkerReputationDeltaParametersUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "DeltaParametersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationDeltaParametersUpdated)
				if err := _BunkerReputation.contract.UnpackLog(event, "DeltaParametersUpdated", log); err != nil {
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

// ParseDeltaParametersUpdated is a log parse operation binding the contract event 0xbbe3d3400bbdd1c673beee5e0de381892cf0afcab96f92e98a7ab105f9bb6deb.
//
// Solidity: event DeltaParametersUpdated()
func (_BunkerReputation *BunkerReputationFilterer) ParseDeltaParametersUpdated(log types.Log) (*BunkerReputationDeltaParametersUpdated, error) {
	event := new(BunkerReputationDeltaParametersUpdated)
	if err := _BunkerReputation.contract.UnpackLog(event, "DeltaParametersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationMaxCustomDeltaUpdatedIterator is returned from FilterMaxCustomDeltaUpdated and is used to iterate over the raw logs and unpacked data for MaxCustomDeltaUpdated events raised by the BunkerReputation contract.
type BunkerReputationMaxCustomDeltaUpdatedIterator struct {
	Event *BunkerReputationMaxCustomDeltaUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerReputationMaxCustomDeltaUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationMaxCustomDeltaUpdated)
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
		it.Event = new(BunkerReputationMaxCustomDeltaUpdated)
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
func (it *BunkerReputationMaxCustomDeltaUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationMaxCustomDeltaUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationMaxCustomDeltaUpdated represents a MaxCustomDeltaUpdated event raised by the BunkerReputation contract.
type BunkerReputationMaxCustomDeltaUpdated struct {
	MaxDelta int16
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMaxCustomDeltaUpdated is a free log retrieval operation binding the contract event 0x098d1be9c3c19e7a0e573dcca5929a3e9a44f7bdde2d01345335e60d5d5c564e.
//
// Solidity: event MaxCustomDeltaUpdated(int16 maxDelta)
func (_BunkerReputation *BunkerReputationFilterer) FilterMaxCustomDeltaUpdated(opts *bind.FilterOpts) (*BunkerReputationMaxCustomDeltaUpdatedIterator, error) {

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "MaxCustomDeltaUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerReputationMaxCustomDeltaUpdatedIterator{contract: _BunkerReputation.contract, event: "MaxCustomDeltaUpdated", logs: logs, sub: sub}, nil
}

// WatchMaxCustomDeltaUpdated is a free log subscription operation binding the contract event 0x098d1be9c3c19e7a0e573dcca5929a3e9a44f7bdde2d01345335e60d5d5c564e.
//
// Solidity: event MaxCustomDeltaUpdated(int16 maxDelta)
func (_BunkerReputation *BunkerReputationFilterer) WatchMaxCustomDeltaUpdated(opts *bind.WatchOpts, sink chan<- *BunkerReputationMaxCustomDeltaUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "MaxCustomDeltaUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationMaxCustomDeltaUpdated)
				if err := _BunkerReputation.contract.UnpackLog(event, "MaxCustomDeltaUpdated", log); err != nil {
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

// ParseMaxCustomDeltaUpdated is a log parse operation binding the contract event 0x098d1be9c3c19e7a0e573dcca5929a3e9a44f7bdde2d01345335e60d5d5c564e.
//
// Solidity: event MaxCustomDeltaUpdated(int16 maxDelta)
func (_BunkerReputation *BunkerReputationFilterer) ParseMaxCustomDeltaUpdated(log types.Log) (*BunkerReputationMaxCustomDeltaUpdated, error) {
	event := new(BunkerReputationMaxCustomDeltaUpdated)
	if err := _BunkerReputation.contract.UnpackLog(event, "MaxCustomDeltaUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationMinScoreForJobsUpdatedIterator is returned from FilterMinScoreForJobsUpdated and is used to iterate over the raw logs and unpacked data for MinScoreForJobsUpdated events raised by the BunkerReputation contract.
type BunkerReputationMinScoreForJobsUpdatedIterator struct {
	Event *BunkerReputationMinScoreForJobsUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerReputationMinScoreForJobsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationMinScoreForJobsUpdated)
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
		it.Event = new(BunkerReputationMinScoreForJobsUpdated)
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
func (it *BunkerReputationMinScoreForJobsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationMinScoreForJobsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationMinScoreForJobsUpdated represents a MinScoreForJobsUpdated event raised by the BunkerReputation contract.
type BunkerReputationMinScoreForJobsUpdated struct {
	MinScore *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMinScoreForJobsUpdated is a free log retrieval operation binding the contract event 0xaa35cd770c922777c44e00b6d7435031c4be4751bf9b17dd944e5d7fb116a1e6.
//
// Solidity: event MinScoreForJobsUpdated(uint256 minScore)
func (_BunkerReputation *BunkerReputationFilterer) FilterMinScoreForJobsUpdated(opts *bind.FilterOpts) (*BunkerReputationMinScoreForJobsUpdatedIterator, error) {

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "MinScoreForJobsUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerReputationMinScoreForJobsUpdatedIterator{contract: _BunkerReputation.contract, event: "MinScoreForJobsUpdated", logs: logs, sub: sub}, nil
}

// WatchMinScoreForJobsUpdated is a free log subscription operation binding the contract event 0xaa35cd770c922777c44e00b6d7435031c4be4751bf9b17dd944e5d7fb116a1e6.
//
// Solidity: event MinScoreForJobsUpdated(uint256 minScore)
func (_BunkerReputation *BunkerReputationFilterer) WatchMinScoreForJobsUpdated(opts *bind.WatchOpts, sink chan<- *BunkerReputationMinScoreForJobsUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "MinScoreForJobsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationMinScoreForJobsUpdated)
				if err := _BunkerReputation.contract.UnpackLog(event, "MinScoreForJobsUpdated", log); err != nil {
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

// ParseMinScoreForJobsUpdated is a log parse operation binding the contract event 0xaa35cd770c922777c44e00b6d7435031c4be4751bf9b17dd944e5d7fb116a1e6.
//
// Solidity: event MinScoreForJobsUpdated(uint256 minScore)
func (_BunkerReputation *BunkerReputationFilterer) ParseMinScoreForJobsUpdated(log types.Log) (*BunkerReputationMinScoreForJobsUpdated, error) {
	event := new(BunkerReputationMinScoreForJobsUpdated)
	if err := _BunkerReputation.contract.UnpackLog(event, "MinScoreForJobsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the BunkerReputation contract.
type BunkerReputationOwnershipTransferStartedIterator struct {
	Event *BunkerReputationOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *BunkerReputationOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationOwnershipTransferStarted)
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
		it.Event = new(BunkerReputationOwnershipTransferStarted)
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
func (it *BunkerReputationOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the BunkerReputation contract.
type BunkerReputationOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerReputation *BunkerReputationFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerReputationOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerReputationOwnershipTransferStartedIterator{contract: _BunkerReputation.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerReputation *BunkerReputationFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *BunkerReputationOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationOwnershipTransferStarted)
				if err := _BunkerReputation.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_BunkerReputation *BunkerReputationFilterer) ParseOwnershipTransferStarted(log types.Log) (*BunkerReputationOwnershipTransferStarted, error) {
	event := new(BunkerReputationOwnershipTransferStarted)
	if err := _BunkerReputation.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BunkerReputation contract.
type BunkerReputationOwnershipTransferredIterator struct {
	Event *BunkerReputationOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BunkerReputationOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationOwnershipTransferred)
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
		it.Event = new(BunkerReputationOwnershipTransferred)
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
func (it *BunkerReputationOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationOwnershipTransferred represents a OwnershipTransferred event raised by the BunkerReputation contract.
type BunkerReputationOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerReputation *BunkerReputationFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerReputationOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerReputationOwnershipTransferredIterator{contract: _BunkerReputation.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerReputation *BunkerReputationFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BunkerReputationOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationOwnershipTransferred)
				if err := _BunkerReputation.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BunkerReputation *BunkerReputationFilterer) ParseOwnershipTransferred(log types.Log) (*BunkerReputationOwnershipTransferred, error) {
	event := new(BunkerReputationOwnershipTransferred)
	if err := _BunkerReputation.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationProviderRegisteredIterator is returned from FilterProviderRegistered and is used to iterate over the raw logs and unpacked data for ProviderRegistered events raised by the BunkerReputation contract.
type BunkerReputationProviderRegisteredIterator struct {
	Event *BunkerReputationProviderRegistered // Event containing the contract specifics and raw log

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
func (it *BunkerReputationProviderRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationProviderRegistered)
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
		it.Event = new(BunkerReputationProviderRegistered)
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
func (it *BunkerReputationProviderRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationProviderRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationProviderRegistered represents a ProviderRegistered event raised by the BunkerReputation contract.
type BunkerReputationProviderRegistered struct {
	Provider common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderRegistered is a free log retrieval operation binding the contract event 0x70abce74777b3838ae60a33a6b9a87d9d25532668fe4fea548554c55868579c0.
//
// Solidity: event ProviderRegistered(address indexed provider)
func (_BunkerReputation *BunkerReputationFilterer) FilterProviderRegistered(opts *bind.FilterOpts, provider []common.Address) (*BunkerReputationProviderRegisteredIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "ProviderRegistered", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerReputationProviderRegisteredIterator{contract: _BunkerReputation.contract, event: "ProviderRegistered", logs: logs, sub: sub}, nil
}

// WatchProviderRegistered is a free log subscription operation binding the contract event 0x70abce74777b3838ae60a33a6b9a87d9d25532668fe4fea548554c55868579c0.
//
// Solidity: event ProviderRegistered(address indexed provider)
func (_BunkerReputation *BunkerReputationFilterer) WatchProviderRegistered(opts *bind.WatchOpts, sink chan<- *BunkerReputationProviderRegistered, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "ProviderRegistered", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationProviderRegistered)
				if err := _BunkerReputation.contract.UnpackLog(event, "ProviderRegistered", log); err != nil {
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

// ParseProviderRegistered is a log parse operation binding the contract event 0x70abce74777b3838ae60a33a6b9a87d9d25532668fe4fea548554c55868579c0.
//
// Solidity: event ProviderRegistered(address indexed provider)
func (_BunkerReputation *BunkerReputationFilterer) ParseProviderRegistered(log types.Log) (*BunkerReputationProviderRegistered, error) {
	event := new(BunkerReputationProviderRegistered)
	if err := _BunkerReputation.contract.UnpackLog(event, "ProviderRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the BunkerReputation contract.
type BunkerReputationRoleAdminChangedIterator struct {
	Event *BunkerReputationRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *BunkerReputationRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationRoleAdminChanged)
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
		it.Event = new(BunkerReputationRoleAdminChanged)
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
func (it *BunkerReputationRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationRoleAdminChanged represents a RoleAdminChanged event raised by the BunkerReputation contract.
type BunkerReputationRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BunkerReputation *BunkerReputationFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*BunkerReputationRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &BunkerReputationRoleAdminChangedIterator{contract: _BunkerReputation.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BunkerReputation *BunkerReputationFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *BunkerReputationRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationRoleAdminChanged)
				if err := _BunkerReputation.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BunkerReputation *BunkerReputationFilterer) ParseRoleAdminChanged(log types.Log) (*BunkerReputationRoleAdminChanged, error) {
	event := new(BunkerReputationRoleAdminChanged)
	if err := _BunkerReputation.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the BunkerReputation contract.
type BunkerReputationRoleGrantedIterator struct {
	Event *BunkerReputationRoleGranted // Event containing the contract specifics and raw log

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
func (it *BunkerReputationRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationRoleGranted)
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
		it.Event = new(BunkerReputationRoleGranted)
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
func (it *BunkerReputationRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationRoleGranted represents a RoleGranted event raised by the BunkerReputation contract.
type BunkerReputationRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerReputation *BunkerReputationFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BunkerReputationRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BunkerReputationRoleGrantedIterator{contract: _BunkerReputation.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerReputation *BunkerReputationFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *BunkerReputationRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationRoleGranted)
				if err := _BunkerReputation.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerReputation *BunkerReputationFilterer) ParseRoleGranted(log types.Log) (*BunkerReputationRoleGranted, error) {
	event := new(BunkerReputationRoleGranted)
	if err := _BunkerReputation.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the BunkerReputation contract.
type BunkerReputationRoleRevokedIterator struct {
	Event *BunkerReputationRoleRevoked // Event containing the contract specifics and raw log

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
func (it *BunkerReputationRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationRoleRevoked)
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
		it.Event = new(BunkerReputationRoleRevoked)
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
func (it *BunkerReputationRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationRoleRevoked represents a RoleRevoked event raised by the BunkerReputation contract.
type BunkerReputationRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerReputation *BunkerReputationFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BunkerReputationRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BunkerReputationRoleRevokedIterator{contract: _BunkerReputation.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerReputation *BunkerReputationFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *BunkerReputationRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationRoleRevoked)
				if err := _BunkerReputation.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerReputation *BunkerReputationFilterer) ParseRoleRevoked(log types.Log) (*BunkerReputationRoleRevoked, error) {
	event := new(BunkerReputationRoleRevoked)
	if err := _BunkerReputation.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationScoreUpdatedIterator is returned from FilterScoreUpdated and is used to iterate over the raw logs and unpacked data for ScoreUpdated events raised by the BunkerReputation contract.
type BunkerReputationScoreUpdatedIterator struct {
	Event *BunkerReputationScoreUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerReputationScoreUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationScoreUpdated)
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
		it.Event = new(BunkerReputationScoreUpdated)
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
func (it *BunkerReputationScoreUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationScoreUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationScoreUpdated represents a ScoreUpdated event raised by the BunkerReputation contract.
type BunkerReputationScoreUpdated struct {
	Provider common.Address
	OldScore uint32
	NewScore uint32
	Reason   string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterScoreUpdated is a free log retrieval operation binding the contract event 0xbbca41c649d7fd3213364f8d335fde7b4993be09ac1a455a7d6644cc8b3fe4bb.
//
// Solidity: event ScoreUpdated(address indexed provider, uint32 oldScore, uint32 newScore, string reason)
func (_BunkerReputation *BunkerReputationFilterer) FilterScoreUpdated(opts *bind.FilterOpts, provider []common.Address) (*BunkerReputationScoreUpdatedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "ScoreUpdated", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerReputationScoreUpdatedIterator{contract: _BunkerReputation.contract, event: "ScoreUpdated", logs: logs, sub: sub}, nil
}

// WatchScoreUpdated is a free log subscription operation binding the contract event 0xbbca41c649d7fd3213364f8d335fde7b4993be09ac1a455a7d6644cc8b3fe4bb.
//
// Solidity: event ScoreUpdated(address indexed provider, uint32 oldScore, uint32 newScore, string reason)
func (_BunkerReputation *BunkerReputationFilterer) WatchScoreUpdated(opts *bind.WatchOpts, sink chan<- *BunkerReputationScoreUpdated, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "ScoreUpdated", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationScoreUpdated)
				if err := _BunkerReputation.contract.UnpackLog(event, "ScoreUpdated", log); err != nil {
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

// ParseScoreUpdated is a log parse operation binding the contract event 0xbbca41c649d7fd3213364f8d335fde7b4993be09ac1a455a7d6644cc8b3fe4bb.
//
// Solidity: event ScoreUpdated(address indexed provider, uint32 oldScore, uint32 newScore, string reason)
func (_BunkerReputation *BunkerReputationFilterer) ParseScoreUpdated(log types.Log) (*BunkerReputationScoreUpdated, error) {
	event := new(BunkerReputationScoreUpdated)
	if err := _BunkerReputation.contract.UnpackLog(event, "ScoreUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerReputationTierThresholdsUpdatedIterator is returned from FilterTierThresholdsUpdated and is used to iterate over the raw logs and unpacked data for TierThresholdsUpdated events raised by the BunkerReputation contract.
type BunkerReputationTierThresholdsUpdatedIterator struct {
	Event *BunkerReputationTierThresholdsUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerReputationTierThresholdsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerReputationTierThresholdsUpdated)
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
		it.Event = new(BunkerReputationTierThresholdsUpdated)
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
func (it *BunkerReputationTierThresholdsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerReputationTierThresholdsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerReputationTierThresholdsUpdated represents a TierThresholdsUpdated event raised by the BunkerReputation contract.
type BunkerReputationTierThresholdsUpdated struct {
	Probation uint16
	Standard  uint16
	Trusted   uint16
	Elite     uint16
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTierThresholdsUpdated is a free log retrieval operation binding the contract event 0x10dda4fe76711959fec19b1e9e1ae3674fed9d64c297bea0b351c15f0fc867eb.
//
// Solidity: event TierThresholdsUpdated(uint16 probation, uint16 standard, uint16 trusted, uint16 elite)
func (_BunkerReputation *BunkerReputationFilterer) FilterTierThresholdsUpdated(opts *bind.FilterOpts) (*BunkerReputationTierThresholdsUpdatedIterator, error) {

	logs, sub, err := _BunkerReputation.contract.FilterLogs(opts, "TierThresholdsUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerReputationTierThresholdsUpdatedIterator{contract: _BunkerReputation.contract, event: "TierThresholdsUpdated", logs: logs, sub: sub}, nil
}

// WatchTierThresholdsUpdated is a free log subscription operation binding the contract event 0x10dda4fe76711959fec19b1e9e1ae3674fed9d64c297bea0b351c15f0fc867eb.
//
// Solidity: event TierThresholdsUpdated(uint16 probation, uint16 standard, uint16 trusted, uint16 elite)
func (_BunkerReputation *BunkerReputationFilterer) WatchTierThresholdsUpdated(opts *bind.WatchOpts, sink chan<- *BunkerReputationTierThresholdsUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerReputation.contract.WatchLogs(opts, "TierThresholdsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerReputationTierThresholdsUpdated)
				if err := _BunkerReputation.contract.UnpackLog(event, "TierThresholdsUpdated", log); err != nil {
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

// ParseTierThresholdsUpdated is a log parse operation binding the contract event 0x10dda4fe76711959fec19b1e9e1ae3674fed9d64c297bea0b351c15f0fc867eb.
//
// Solidity: event TierThresholdsUpdated(uint16 probation, uint16 standard, uint16 trusted, uint16 elite)
func (_BunkerReputation *BunkerReputationFilterer) ParseTierThresholdsUpdated(log types.Log) (*BunkerReputationTierThresholdsUpdated, error) {
	event := new(BunkerReputationTierThresholdsUpdated)
	if err := _BunkerReputation.contract.UnpackLog(event, "TierThresholdsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
