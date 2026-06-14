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

// BunkerVerificationAttestationRecord is an auto generated low-level Go binding around an user-defined struct.
type BunkerVerificationAttestationRecord struct {
	LastAttestationHash [32]byte
	LastAttestationTime *big.Int
	TotalAttestations   uint32
	MissedAttestations  uint32
	ConsecutiveMissed   uint32
	Suspended           bool
}

// BunkerVerificationMetaData contains all meta data concerning the BunkerVerification contract.
var BunkerVerificationMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERIFIER_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"attestationInterval\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"attestations\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"lastAttestationHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"lastAttestationTime\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"totalAttestations\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"missedAttestations\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"consecutiveMissed\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"suspended\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"challengeAttestation\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"fraudProof\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"checkMissedAttestations\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getAttestation\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"record\",\"type\":\"tuple\",\"internalType\":\"structBunkerVerification.AttestationRecord\",\"components\":[{\"name\":\"lastAttestationHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"lastAttestationTime\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"totalAttestations\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"missedAttestations\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"consecutiveMissed\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"suspended\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isAttestationCurrent\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"current\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxMissedAttestations\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"reinstateProvider\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"reinstatementCooldown\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAttestationInterval\",\"inputs\":[{\"name\":\"interval\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxMissedAttestations\",\"inputs\":[{\"name\":\"maxMissed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setReinstatementCooldown\",\"inputs\":[{\"name\":\"newCooldown\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitAttestation\",\"inputs\":[{\"name\":\"attestationHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"suspendedAt\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"AttestationChallenged\",\"inputs\":[{\"name\":\"challenger\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"attestationHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AttestationIntervalUpdated\",\"inputs\":[{\"name\":\"oldInterval\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newInterval\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AttestationMissed\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"consecutiveMissed\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AttestationSubmitted\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"attestationHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"totalAttestations\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MaxMissedAttestationsUpdated\",\"inputs\":[{\"name\":\"oldMax\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newMax\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderReinstated\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderSuspended\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"missedCount\",\"type\":\"uint32\",\"indexed\":false,\"internalType\":\"uint32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReinstatementCooldownUpdated\",\"inputs\":[{\"name\":\"newCooldown\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"AttestationTooEarly\",\"inputs\":[{\"name\":\"nextAllowed\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"current\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidAttestation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidCooldown\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInterval\",\"inputs\":[{\"name\":\"interval\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidMaxMissed\",\"inputs\":[{\"name\":\"maxMissed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ProviderNotAttesting\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ProviderSuspendedError\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ReinstatementCooldownActive\",\"inputs\":[]}]",
}

// BunkerVerificationABI is the input ABI used to generate the binding from.
// Deprecated: Use BunkerVerificationMetaData.ABI instead.
var BunkerVerificationABI = BunkerVerificationMetaData.ABI

// BunkerVerification is an auto generated Go binding around an Ethereum contract.
type BunkerVerification struct {
	BunkerVerificationCaller     // Read-only binding to the contract
	BunkerVerificationTransactor // Write-only binding to the contract
	BunkerVerificationFilterer   // Log filterer for contract events
}

// BunkerVerificationCaller is an auto generated read-only Go binding around an Ethereum contract.
type BunkerVerificationCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerVerificationTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BunkerVerificationTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerVerificationFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BunkerVerificationFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerVerificationSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BunkerVerificationSession struct {
	Contract     *BunkerVerification // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// BunkerVerificationCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BunkerVerificationCallerSession struct {
	Contract *BunkerVerificationCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// BunkerVerificationTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BunkerVerificationTransactorSession struct {
	Contract     *BunkerVerificationTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// BunkerVerificationRaw is an auto generated low-level Go binding around an Ethereum contract.
type BunkerVerificationRaw struct {
	Contract *BunkerVerification // Generic contract binding to access the raw methods on
}

// BunkerVerificationCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BunkerVerificationCallerRaw struct {
	Contract *BunkerVerificationCaller // Generic read-only contract binding to access the raw methods on
}

// BunkerVerificationTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BunkerVerificationTransactorRaw struct {
	Contract *BunkerVerificationTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBunkerVerification creates a new instance of BunkerVerification, bound to a specific deployed contract.
func NewBunkerVerification(address common.Address, backend bind.ContractBackend) (*BunkerVerification, error) {
	contract, err := bindBunkerVerification(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BunkerVerification{BunkerVerificationCaller: BunkerVerificationCaller{contract: contract}, BunkerVerificationTransactor: BunkerVerificationTransactor{contract: contract}, BunkerVerificationFilterer: BunkerVerificationFilterer{contract: contract}}, nil
}

// NewBunkerVerificationCaller creates a new read-only instance of BunkerVerification, bound to a specific deployed contract.
func NewBunkerVerificationCaller(address common.Address, caller bind.ContractCaller) (*BunkerVerificationCaller, error) {
	contract, err := bindBunkerVerification(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationCaller{contract: contract}, nil
}

// NewBunkerVerificationTransactor creates a new write-only instance of BunkerVerification, bound to a specific deployed contract.
func NewBunkerVerificationTransactor(address common.Address, transactor bind.ContractTransactor) (*BunkerVerificationTransactor, error) {
	contract, err := bindBunkerVerification(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationTransactor{contract: contract}, nil
}

// NewBunkerVerificationFilterer creates a new log filterer instance of BunkerVerification, bound to a specific deployed contract.
func NewBunkerVerificationFilterer(address common.Address, filterer bind.ContractFilterer) (*BunkerVerificationFilterer, error) {
	contract, err := bindBunkerVerification(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationFilterer{contract: contract}, nil
}

// bindBunkerVerification binds a generic wrapper to an already deployed contract.
func bindBunkerVerification(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BunkerVerificationMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerVerification *BunkerVerificationRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerVerification.Contract.BunkerVerificationCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerVerification *BunkerVerificationRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerVerification.Contract.BunkerVerificationTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerVerification *BunkerVerificationRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerVerification.Contract.BunkerVerificationTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerVerification *BunkerVerificationCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerVerification.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerVerification *BunkerVerificationTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerVerification.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerVerification *BunkerVerificationTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerVerification.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerVerification *BunkerVerificationCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerVerification *BunkerVerificationSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _BunkerVerification.Contract.DEFAULTADMINROLE(&_BunkerVerification.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerVerification *BunkerVerificationCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _BunkerVerification.Contract.DEFAULTADMINROLE(&_BunkerVerification.CallOpts)
}

// VERIFIERROLE is a free data retrieval call binding the contract method 0xe7705db6.
//
// Solidity: function VERIFIER_ROLE() view returns(bytes32)
func (_BunkerVerification *BunkerVerificationCaller) VERIFIERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "VERIFIER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VERIFIERROLE is a free data retrieval call binding the contract method 0xe7705db6.
//
// Solidity: function VERIFIER_ROLE() view returns(bytes32)
func (_BunkerVerification *BunkerVerificationSession) VERIFIERROLE() ([32]byte, error) {
	return _BunkerVerification.Contract.VERIFIERROLE(&_BunkerVerification.CallOpts)
}

// VERIFIERROLE is a free data retrieval call binding the contract method 0xe7705db6.
//
// Solidity: function VERIFIER_ROLE() view returns(bytes32)
func (_BunkerVerification *BunkerVerificationCallerSession) VERIFIERROLE() ([32]byte, error) {
	return _BunkerVerification.Contract.VERIFIERROLE(&_BunkerVerification.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerVerification *BunkerVerificationCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerVerification *BunkerVerificationSession) VERSION() (string, error) {
	return _BunkerVerification.Contract.VERSION(&_BunkerVerification.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerVerification *BunkerVerificationCallerSession) VERSION() (string, error) {
	return _BunkerVerification.Contract.VERSION(&_BunkerVerification.CallOpts)
}

// AttestationInterval is a free data retrieval call binding the contract method 0x5ee7bd80.
//
// Solidity: function attestationInterval() view returns(uint256)
func (_BunkerVerification *BunkerVerificationCaller) AttestationInterval(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "attestationInterval")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AttestationInterval is a free data retrieval call binding the contract method 0x5ee7bd80.
//
// Solidity: function attestationInterval() view returns(uint256)
func (_BunkerVerification *BunkerVerificationSession) AttestationInterval() (*big.Int, error) {
	return _BunkerVerification.Contract.AttestationInterval(&_BunkerVerification.CallOpts)
}

// AttestationInterval is a free data retrieval call binding the contract method 0x5ee7bd80.
//
// Solidity: function attestationInterval() view returns(uint256)
func (_BunkerVerification *BunkerVerificationCallerSession) AttestationInterval() (*big.Int, error) {
	return _BunkerVerification.Contract.AttestationInterval(&_BunkerVerification.CallOpts)
}

// Attestations is a free data retrieval call binding the contract method 0x20357048.
//
// Solidity: function attestations(address ) view returns(bytes32 lastAttestationHash, uint48 lastAttestationTime, uint32 totalAttestations, uint32 missedAttestations, uint32 consecutiveMissed, bool suspended)
func (_BunkerVerification *BunkerVerificationCaller) Attestations(opts *bind.CallOpts, arg0 common.Address) (struct {
	LastAttestationHash [32]byte
	LastAttestationTime *big.Int
	TotalAttestations   uint32
	MissedAttestations  uint32
	ConsecutiveMissed   uint32
	Suspended           bool
}, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "attestations", arg0)

	outstruct := new(struct {
		LastAttestationHash [32]byte
		LastAttestationTime *big.Int
		TotalAttestations   uint32
		MissedAttestations  uint32
		ConsecutiveMissed   uint32
		Suspended           bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.LastAttestationHash = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.LastAttestationTime = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.TotalAttestations = *abi.ConvertType(out[2], new(uint32)).(*uint32)
	outstruct.MissedAttestations = *abi.ConvertType(out[3], new(uint32)).(*uint32)
	outstruct.ConsecutiveMissed = *abi.ConvertType(out[4], new(uint32)).(*uint32)
	outstruct.Suspended = *abi.ConvertType(out[5], new(bool)).(*bool)

	return *outstruct, err

}

// Attestations is a free data retrieval call binding the contract method 0x20357048.
//
// Solidity: function attestations(address ) view returns(bytes32 lastAttestationHash, uint48 lastAttestationTime, uint32 totalAttestations, uint32 missedAttestations, uint32 consecutiveMissed, bool suspended)
func (_BunkerVerification *BunkerVerificationSession) Attestations(arg0 common.Address) (struct {
	LastAttestationHash [32]byte
	LastAttestationTime *big.Int
	TotalAttestations   uint32
	MissedAttestations  uint32
	ConsecutiveMissed   uint32
	Suspended           bool
}, error) {
	return _BunkerVerification.Contract.Attestations(&_BunkerVerification.CallOpts, arg0)
}

// Attestations is a free data retrieval call binding the contract method 0x20357048.
//
// Solidity: function attestations(address ) view returns(bytes32 lastAttestationHash, uint48 lastAttestationTime, uint32 totalAttestations, uint32 missedAttestations, uint32 consecutiveMissed, bool suspended)
func (_BunkerVerification *BunkerVerificationCallerSession) Attestations(arg0 common.Address) (struct {
	LastAttestationHash [32]byte
	LastAttestationTime *big.Int
	TotalAttestations   uint32
	MissedAttestations  uint32
	ConsecutiveMissed   uint32
	Suspended           bool
}, error) {
	return _BunkerVerification.Contract.Attestations(&_BunkerVerification.CallOpts, arg0)
}

// GetAttestation is a free data retrieval call binding the contract method 0xf9b71797.
//
// Solidity: function getAttestation(address provider) view returns((bytes32,uint48,uint32,uint32,uint32,bool) record)
func (_BunkerVerification *BunkerVerificationCaller) GetAttestation(opts *bind.CallOpts, provider common.Address) (BunkerVerificationAttestationRecord, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "getAttestation", provider)

	if err != nil {
		return *new(BunkerVerificationAttestationRecord), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerVerificationAttestationRecord)).(*BunkerVerificationAttestationRecord)

	return out0, err

}

// GetAttestation is a free data retrieval call binding the contract method 0xf9b71797.
//
// Solidity: function getAttestation(address provider) view returns((bytes32,uint48,uint32,uint32,uint32,bool) record)
func (_BunkerVerification *BunkerVerificationSession) GetAttestation(provider common.Address) (BunkerVerificationAttestationRecord, error) {
	return _BunkerVerification.Contract.GetAttestation(&_BunkerVerification.CallOpts, provider)
}

// GetAttestation is a free data retrieval call binding the contract method 0xf9b71797.
//
// Solidity: function getAttestation(address provider) view returns((bytes32,uint48,uint32,uint32,uint32,bool) record)
func (_BunkerVerification *BunkerVerificationCallerSession) GetAttestation(provider common.Address) (BunkerVerificationAttestationRecord, error) {
	return _BunkerVerification.Contract.GetAttestation(&_BunkerVerification.CallOpts, provider)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerVerification *BunkerVerificationCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerVerification *BunkerVerificationSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _BunkerVerification.Contract.GetRoleAdmin(&_BunkerVerification.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerVerification *BunkerVerificationCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _BunkerVerification.Contract.GetRoleAdmin(&_BunkerVerification.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerVerification *BunkerVerificationCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerVerification *BunkerVerificationSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _BunkerVerification.Contract.HasRole(&_BunkerVerification.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerVerification *BunkerVerificationCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _BunkerVerification.Contract.HasRole(&_BunkerVerification.CallOpts, role, account)
}

// IsAttestationCurrent is a free data retrieval call binding the contract method 0x6159f222.
//
// Solidity: function isAttestationCurrent(address provider) view returns(bool current)
func (_BunkerVerification *BunkerVerificationCaller) IsAttestationCurrent(opts *bind.CallOpts, provider common.Address) (bool, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "isAttestationCurrent", provider)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAttestationCurrent is a free data retrieval call binding the contract method 0x6159f222.
//
// Solidity: function isAttestationCurrent(address provider) view returns(bool current)
func (_BunkerVerification *BunkerVerificationSession) IsAttestationCurrent(provider common.Address) (bool, error) {
	return _BunkerVerification.Contract.IsAttestationCurrent(&_BunkerVerification.CallOpts, provider)
}

// IsAttestationCurrent is a free data retrieval call binding the contract method 0x6159f222.
//
// Solidity: function isAttestationCurrent(address provider) view returns(bool current)
func (_BunkerVerification *BunkerVerificationCallerSession) IsAttestationCurrent(provider common.Address) (bool, error) {
	return _BunkerVerification.Contract.IsAttestationCurrent(&_BunkerVerification.CallOpts, provider)
}

// MaxMissedAttestations is a free data retrieval call binding the contract method 0x88af3fe0.
//
// Solidity: function maxMissedAttestations() view returns(uint256)
func (_BunkerVerification *BunkerVerificationCaller) MaxMissedAttestations(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "maxMissedAttestations")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxMissedAttestations is a free data retrieval call binding the contract method 0x88af3fe0.
//
// Solidity: function maxMissedAttestations() view returns(uint256)
func (_BunkerVerification *BunkerVerificationSession) MaxMissedAttestations() (*big.Int, error) {
	return _BunkerVerification.Contract.MaxMissedAttestations(&_BunkerVerification.CallOpts)
}

// MaxMissedAttestations is a free data retrieval call binding the contract method 0x88af3fe0.
//
// Solidity: function maxMissedAttestations() view returns(uint256)
func (_BunkerVerification *BunkerVerificationCallerSession) MaxMissedAttestations() (*big.Int, error) {
	return _BunkerVerification.Contract.MaxMissedAttestations(&_BunkerVerification.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerVerification *BunkerVerificationCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerVerification *BunkerVerificationSession) Owner() (common.Address, error) {
	return _BunkerVerification.Contract.Owner(&_BunkerVerification.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerVerification *BunkerVerificationCallerSession) Owner() (common.Address, error) {
	return _BunkerVerification.Contract.Owner(&_BunkerVerification.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerVerification *BunkerVerificationCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerVerification *BunkerVerificationSession) PendingOwner() (common.Address, error) {
	return _BunkerVerification.Contract.PendingOwner(&_BunkerVerification.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerVerification *BunkerVerificationCallerSession) PendingOwner() (common.Address, error) {
	return _BunkerVerification.Contract.PendingOwner(&_BunkerVerification.CallOpts)
}

// ReinstatementCooldown is a free data retrieval call binding the contract method 0x047bccf5.
//
// Solidity: function reinstatementCooldown() view returns(uint256)
func (_BunkerVerification *BunkerVerificationCaller) ReinstatementCooldown(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "reinstatementCooldown")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ReinstatementCooldown is a free data retrieval call binding the contract method 0x047bccf5.
//
// Solidity: function reinstatementCooldown() view returns(uint256)
func (_BunkerVerification *BunkerVerificationSession) ReinstatementCooldown() (*big.Int, error) {
	return _BunkerVerification.Contract.ReinstatementCooldown(&_BunkerVerification.CallOpts)
}

// ReinstatementCooldown is a free data retrieval call binding the contract method 0x047bccf5.
//
// Solidity: function reinstatementCooldown() view returns(uint256)
func (_BunkerVerification *BunkerVerificationCallerSession) ReinstatementCooldown() (*big.Int, error) {
	return _BunkerVerification.Contract.ReinstatementCooldown(&_BunkerVerification.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerVerification *BunkerVerificationCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerVerification *BunkerVerificationSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _BunkerVerification.Contract.SupportsInterface(&_BunkerVerification.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerVerification *BunkerVerificationCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _BunkerVerification.Contract.SupportsInterface(&_BunkerVerification.CallOpts, interfaceId)
}

// SuspendedAt is a free data retrieval call binding the contract method 0x989b1d60.
//
// Solidity: function suspendedAt(address ) view returns(uint256)
func (_BunkerVerification *BunkerVerificationCaller) SuspendedAt(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerVerification.contract.Call(opts, &out, "suspendedAt", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SuspendedAt is a free data retrieval call binding the contract method 0x989b1d60.
//
// Solidity: function suspendedAt(address ) view returns(uint256)
func (_BunkerVerification *BunkerVerificationSession) SuspendedAt(arg0 common.Address) (*big.Int, error) {
	return _BunkerVerification.Contract.SuspendedAt(&_BunkerVerification.CallOpts, arg0)
}

// SuspendedAt is a free data retrieval call binding the contract method 0x989b1d60.
//
// Solidity: function suspendedAt(address ) view returns(uint256)
func (_BunkerVerification *BunkerVerificationCallerSession) SuspendedAt(arg0 common.Address) (*big.Int, error) {
	return _BunkerVerification.Contract.SuspendedAt(&_BunkerVerification.CallOpts, arg0)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerVerification *BunkerVerificationTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerVerification *BunkerVerificationSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerVerification.Contract.AcceptOwnership(&_BunkerVerification.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerVerification.Contract.AcceptOwnership(&_BunkerVerification.TransactOpts)
}

// ChallengeAttestation is a paid mutator transaction binding the contract method 0x904aa811.
//
// Solidity: function challengeAttestation(address provider, bytes fraudProof) returns()
func (_BunkerVerification *BunkerVerificationTransactor) ChallengeAttestation(opts *bind.TransactOpts, provider common.Address, fraudProof []byte) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "challengeAttestation", provider, fraudProof)
}

// ChallengeAttestation is a paid mutator transaction binding the contract method 0x904aa811.
//
// Solidity: function challengeAttestation(address provider, bytes fraudProof) returns()
func (_BunkerVerification *BunkerVerificationSession) ChallengeAttestation(provider common.Address, fraudProof []byte) (*types.Transaction, error) {
	return _BunkerVerification.Contract.ChallengeAttestation(&_BunkerVerification.TransactOpts, provider, fraudProof)
}

// ChallengeAttestation is a paid mutator transaction binding the contract method 0x904aa811.
//
// Solidity: function challengeAttestation(address provider, bytes fraudProof) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) ChallengeAttestation(provider common.Address, fraudProof []byte) (*types.Transaction, error) {
	return _BunkerVerification.Contract.ChallengeAttestation(&_BunkerVerification.TransactOpts, provider, fraudProof)
}

// CheckMissedAttestations is a paid mutator transaction binding the contract method 0x9b22ab77.
//
// Solidity: function checkMissedAttestations(address provider) returns()
func (_BunkerVerification *BunkerVerificationTransactor) CheckMissedAttestations(opts *bind.TransactOpts, provider common.Address) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "checkMissedAttestations", provider)
}

// CheckMissedAttestations is a paid mutator transaction binding the contract method 0x9b22ab77.
//
// Solidity: function checkMissedAttestations(address provider) returns()
func (_BunkerVerification *BunkerVerificationSession) CheckMissedAttestations(provider common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.CheckMissedAttestations(&_BunkerVerification.TransactOpts, provider)
}

// CheckMissedAttestations is a paid mutator transaction binding the contract method 0x9b22ab77.
//
// Solidity: function checkMissedAttestations(address provider) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) CheckMissedAttestations(provider common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.CheckMissedAttestations(&_BunkerVerification.TransactOpts, provider)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerVerification *BunkerVerificationTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerVerification *BunkerVerificationSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.GrantRole(&_BunkerVerification.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.GrantRole(&_BunkerVerification.TransactOpts, role, account)
}

// ReinstateProvider is a paid mutator transaction binding the contract method 0x99e82915.
//
// Solidity: function reinstateProvider(address provider) returns()
func (_BunkerVerification *BunkerVerificationTransactor) ReinstateProvider(opts *bind.TransactOpts, provider common.Address) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "reinstateProvider", provider)
}

// ReinstateProvider is a paid mutator transaction binding the contract method 0x99e82915.
//
// Solidity: function reinstateProvider(address provider) returns()
func (_BunkerVerification *BunkerVerificationSession) ReinstateProvider(provider common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.ReinstateProvider(&_BunkerVerification.TransactOpts, provider)
}

// ReinstateProvider is a paid mutator transaction binding the contract method 0x99e82915.
//
// Solidity: function reinstateProvider(address provider) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) ReinstateProvider(provider common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.ReinstateProvider(&_BunkerVerification.TransactOpts, provider)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerVerification *BunkerVerificationTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerVerification *BunkerVerificationSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerVerification.Contract.RenounceOwnership(&_BunkerVerification.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerVerification.Contract.RenounceOwnership(&_BunkerVerification.TransactOpts)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerVerification *BunkerVerificationTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerVerification *BunkerVerificationSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.RenounceRole(&_BunkerVerification.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.RenounceRole(&_BunkerVerification.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerVerification *BunkerVerificationTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerVerification *BunkerVerificationSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.RevokeRole(&_BunkerVerification.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.RevokeRole(&_BunkerVerification.TransactOpts, role, account)
}

// SetAttestationInterval is a paid mutator transaction binding the contract method 0xf5264a2a.
//
// Solidity: function setAttestationInterval(uint256 interval) returns()
func (_BunkerVerification *BunkerVerificationTransactor) SetAttestationInterval(opts *bind.TransactOpts, interval *big.Int) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "setAttestationInterval", interval)
}

// SetAttestationInterval is a paid mutator transaction binding the contract method 0xf5264a2a.
//
// Solidity: function setAttestationInterval(uint256 interval) returns()
func (_BunkerVerification *BunkerVerificationSession) SetAttestationInterval(interval *big.Int) (*types.Transaction, error) {
	return _BunkerVerification.Contract.SetAttestationInterval(&_BunkerVerification.TransactOpts, interval)
}

// SetAttestationInterval is a paid mutator transaction binding the contract method 0xf5264a2a.
//
// Solidity: function setAttestationInterval(uint256 interval) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) SetAttestationInterval(interval *big.Int) (*types.Transaction, error) {
	return _BunkerVerification.Contract.SetAttestationInterval(&_BunkerVerification.TransactOpts, interval)
}

// SetMaxMissedAttestations is a paid mutator transaction binding the contract method 0x7b8093cd.
//
// Solidity: function setMaxMissedAttestations(uint256 maxMissed) returns()
func (_BunkerVerification *BunkerVerificationTransactor) SetMaxMissedAttestations(opts *bind.TransactOpts, maxMissed *big.Int) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "setMaxMissedAttestations", maxMissed)
}

// SetMaxMissedAttestations is a paid mutator transaction binding the contract method 0x7b8093cd.
//
// Solidity: function setMaxMissedAttestations(uint256 maxMissed) returns()
func (_BunkerVerification *BunkerVerificationSession) SetMaxMissedAttestations(maxMissed *big.Int) (*types.Transaction, error) {
	return _BunkerVerification.Contract.SetMaxMissedAttestations(&_BunkerVerification.TransactOpts, maxMissed)
}

// SetMaxMissedAttestations is a paid mutator transaction binding the contract method 0x7b8093cd.
//
// Solidity: function setMaxMissedAttestations(uint256 maxMissed) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) SetMaxMissedAttestations(maxMissed *big.Int) (*types.Transaction, error) {
	return _BunkerVerification.Contract.SetMaxMissedAttestations(&_BunkerVerification.TransactOpts, maxMissed)
}

// SetReinstatementCooldown is a paid mutator transaction binding the contract method 0xe37020c6.
//
// Solidity: function setReinstatementCooldown(uint256 newCooldown) returns()
func (_BunkerVerification *BunkerVerificationTransactor) SetReinstatementCooldown(opts *bind.TransactOpts, newCooldown *big.Int) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "setReinstatementCooldown", newCooldown)
}

// SetReinstatementCooldown is a paid mutator transaction binding the contract method 0xe37020c6.
//
// Solidity: function setReinstatementCooldown(uint256 newCooldown) returns()
func (_BunkerVerification *BunkerVerificationSession) SetReinstatementCooldown(newCooldown *big.Int) (*types.Transaction, error) {
	return _BunkerVerification.Contract.SetReinstatementCooldown(&_BunkerVerification.TransactOpts, newCooldown)
}

// SetReinstatementCooldown is a paid mutator transaction binding the contract method 0xe37020c6.
//
// Solidity: function setReinstatementCooldown(uint256 newCooldown) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) SetReinstatementCooldown(newCooldown *big.Int) (*types.Transaction, error) {
	return _BunkerVerification.Contract.SetReinstatementCooldown(&_BunkerVerification.TransactOpts, newCooldown)
}

// SubmitAttestation is a paid mutator transaction binding the contract method 0x4506b5c5.
//
// Solidity: function submitAttestation(bytes32 attestationHash) returns()
func (_BunkerVerification *BunkerVerificationTransactor) SubmitAttestation(opts *bind.TransactOpts, attestationHash [32]byte) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "submitAttestation", attestationHash)
}

// SubmitAttestation is a paid mutator transaction binding the contract method 0x4506b5c5.
//
// Solidity: function submitAttestation(bytes32 attestationHash) returns()
func (_BunkerVerification *BunkerVerificationSession) SubmitAttestation(attestationHash [32]byte) (*types.Transaction, error) {
	return _BunkerVerification.Contract.SubmitAttestation(&_BunkerVerification.TransactOpts, attestationHash)
}

// SubmitAttestation is a paid mutator transaction binding the contract method 0x4506b5c5.
//
// Solidity: function submitAttestation(bytes32 attestationHash) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) SubmitAttestation(attestationHash [32]byte) (*types.Transaction, error) {
	return _BunkerVerification.Contract.SubmitAttestation(&_BunkerVerification.TransactOpts, attestationHash)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerVerification *BunkerVerificationTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BunkerVerification.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerVerification *BunkerVerificationSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.TransferOwnership(&_BunkerVerification.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerVerification *BunkerVerificationTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerVerification.Contract.TransferOwnership(&_BunkerVerification.TransactOpts, newOwner)
}

// BunkerVerificationAttestationChallengedIterator is returned from FilterAttestationChallenged and is used to iterate over the raw logs and unpacked data for AttestationChallenged events raised by the BunkerVerification contract.
type BunkerVerificationAttestationChallengedIterator struct {
	Event *BunkerVerificationAttestationChallenged // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationAttestationChallengedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationAttestationChallenged)
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
		it.Event = new(BunkerVerificationAttestationChallenged)
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
func (it *BunkerVerificationAttestationChallengedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationAttestationChallengedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationAttestationChallenged represents a AttestationChallenged event raised by the BunkerVerification contract.
type BunkerVerificationAttestationChallenged struct {
	Challenger      common.Address
	Provider        common.Address
	AttestationHash [32]byte
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterAttestationChallenged is a free log retrieval operation binding the contract event 0x48d84a903b64623652d0d22252a4a66df37f59e49c2e25804d03eaa1e2a3c425.
//
// Solidity: event AttestationChallenged(address indexed challenger, address indexed provider, bytes32 attestationHash)
func (_BunkerVerification *BunkerVerificationFilterer) FilterAttestationChallenged(opts *bind.FilterOpts, challenger []common.Address, provider []common.Address) (*BunkerVerificationAttestationChallengedIterator, error) {

	var challengerRule []interface{}
	for _, challengerItem := range challenger {
		challengerRule = append(challengerRule, challengerItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "AttestationChallenged", challengerRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationAttestationChallengedIterator{contract: _BunkerVerification.contract, event: "AttestationChallenged", logs: logs, sub: sub}, nil
}

// WatchAttestationChallenged is a free log subscription operation binding the contract event 0x48d84a903b64623652d0d22252a4a66df37f59e49c2e25804d03eaa1e2a3c425.
//
// Solidity: event AttestationChallenged(address indexed challenger, address indexed provider, bytes32 attestationHash)
func (_BunkerVerification *BunkerVerificationFilterer) WatchAttestationChallenged(opts *bind.WatchOpts, sink chan<- *BunkerVerificationAttestationChallenged, challenger []common.Address, provider []common.Address) (event.Subscription, error) {

	var challengerRule []interface{}
	for _, challengerItem := range challenger {
		challengerRule = append(challengerRule, challengerItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "AttestationChallenged", challengerRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationAttestationChallenged)
				if err := _BunkerVerification.contract.UnpackLog(event, "AttestationChallenged", log); err != nil {
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

// ParseAttestationChallenged is a log parse operation binding the contract event 0x48d84a903b64623652d0d22252a4a66df37f59e49c2e25804d03eaa1e2a3c425.
//
// Solidity: event AttestationChallenged(address indexed challenger, address indexed provider, bytes32 attestationHash)
func (_BunkerVerification *BunkerVerificationFilterer) ParseAttestationChallenged(log types.Log) (*BunkerVerificationAttestationChallenged, error) {
	event := new(BunkerVerificationAttestationChallenged)
	if err := _BunkerVerification.contract.UnpackLog(event, "AttestationChallenged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationAttestationIntervalUpdatedIterator is returned from FilterAttestationIntervalUpdated and is used to iterate over the raw logs and unpacked data for AttestationIntervalUpdated events raised by the BunkerVerification contract.
type BunkerVerificationAttestationIntervalUpdatedIterator struct {
	Event *BunkerVerificationAttestationIntervalUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationAttestationIntervalUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationAttestationIntervalUpdated)
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
		it.Event = new(BunkerVerificationAttestationIntervalUpdated)
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
func (it *BunkerVerificationAttestationIntervalUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationAttestationIntervalUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationAttestationIntervalUpdated represents a AttestationIntervalUpdated event raised by the BunkerVerification contract.
type BunkerVerificationAttestationIntervalUpdated struct {
	OldInterval *big.Int
	NewInterval *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterAttestationIntervalUpdated is a free log retrieval operation binding the contract event 0x269143a4310afed4c9c59d993804ae3a0c5a60064c6b5aa267857857f98f4bd1.
//
// Solidity: event AttestationIntervalUpdated(uint256 oldInterval, uint256 newInterval)
func (_BunkerVerification *BunkerVerificationFilterer) FilterAttestationIntervalUpdated(opts *bind.FilterOpts) (*BunkerVerificationAttestationIntervalUpdatedIterator, error) {

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "AttestationIntervalUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationAttestationIntervalUpdatedIterator{contract: _BunkerVerification.contract, event: "AttestationIntervalUpdated", logs: logs, sub: sub}, nil
}

// WatchAttestationIntervalUpdated is a free log subscription operation binding the contract event 0x269143a4310afed4c9c59d993804ae3a0c5a60064c6b5aa267857857f98f4bd1.
//
// Solidity: event AttestationIntervalUpdated(uint256 oldInterval, uint256 newInterval)
func (_BunkerVerification *BunkerVerificationFilterer) WatchAttestationIntervalUpdated(opts *bind.WatchOpts, sink chan<- *BunkerVerificationAttestationIntervalUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "AttestationIntervalUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationAttestationIntervalUpdated)
				if err := _BunkerVerification.contract.UnpackLog(event, "AttestationIntervalUpdated", log); err != nil {
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

// ParseAttestationIntervalUpdated is a log parse operation binding the contract event 0x269143a4310afed4c9c59d993804ae3a0c5a60064c6b5aa267857857f98f4bd1.
//
// Solidity: event AttestationIntervalUpdated(uint256 oldInterval, uint256 newInterval)
func (_BunkerVerification *BunkerVerificationFilterer) ParseAttestationIntervalUpdated(log types.Log) (*BunkerVerificationAttestationIntervalUpdated, error) {
	event := new(BunkerVerificationAttestationIntervalUpdated)
	if err := _BunkerVerification.contract.UnpackLog(event, "AttestationIntervalUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationAttestationMissedIterator is returned from FilterAttestationMissed and is used to iterate over the raw logs and unpacked data for AttestationMissed events raised by the BunkerVerification contract.
type BunkerVerificationAttestationMissedIterator struct {
	Event *BunkerVerificationAttestationMissed // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationAttestationMissedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationAttestationMissed)
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
		it.Event = new(BunkerVerificationAttestationMissed)
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
func (it *BunkerVerificationAttestationMissedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationAttestationMissedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationAttestationMissed represents a AttestationMissed event raised by the BunkerVerification contract.
type BunkerVerificationAttestationMissed struct {
	Provider          common.Address
	ConsecutiveMissed uint32
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterAttestationMissed is a free log retrieval operation binding the contract event 0x280ba3e3e6e91883f617dd57a3f1a133eadd1364f3f394d6551bb09d1ba77581.
//
// Solidity: event AttestationMissed(address indexed provider, uint32 consecutiveMissed)
func (_BunkerVerification *BunkerVerificationFilterer) FilterAttestationMissed(opts *bind.FilterOpts, provider []common.Address) (*BunkerVerificationAttestationMissedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "AttestationMissed", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationAttestationMissedIterator{contract: _BunkerVerification.contract, event: "AttestationMissed", logs: logs, sub: sub}, nil
}

// WatchAttestationMissed is a free log subscription operation binding the contract event 0x280ba3e3e6e91883f617dd57a3f1a133eadd1364f3f394d6551bb09d1ba77581.
//
// Solidity: event AttestationMissed(address indexed provider, uint32 consecutiveMissed)
func (_BunkerVerification *BunkerVerificationFilterer) WatchAttestationMissed(opts *bind.WatchOpts, sink chan<- *BunkerVerificationAttestationMissed, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "AttestationMissed", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationAttestationMissed)
				if err := _BunkerVerification.contract.UnpackLog(event, "AttestationMissed", log); err != nil {
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

// ParseAttestationMissed is a log parse operation binding the contract event 0x280ba3e3e6e91883f617dd57a3f1a133eadd1364f3f394d6551bb09d1ba77581.
//
// Solidity: event AttestationMissed(address indexed provider, uint32 consecutiveMissed)
func (_BunkerVerification *BunkerVerificationFilterer) ParseAttestationMissed(log types.Log) (*BunkerVerificationAttestationMissed, error) {
	event := new(BunkerVerificationAttestationMissed)
	if err := _BunkerVerification.contract.UnpackLog(event, "AttestationMissed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationAttestationSubmittedIterator is returned from FilterAttestationSubmitted and is used to iterate over the raw logs and unpacked data for AttestationSubmitted events raised by the BunkerVerification contract.
type BunkerVerificationAttestationSubmittedIterator struct {
	Event *BunkerVerificationAttestationSubmitted // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationAttestationSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationAttestationSubmitted)
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
		it.Event = new(BunkerVerificationAttestationSubmitted)
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
func (it *BunkerVerificationAttestationSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationAttestationSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationAttestationSubmitted represents a AttestationSubmitted event raised by the BunkerVerification contract.
type BunkerVerificationAttestationSubmitted struct {
	Provider          common.Address
	AttestationHash   [32]byte
	TotalAttestations uint32
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterAttestationSubmitted is a free log retrieval operation binding the contract event 0x424013c4cbe6b87d14eb842300bd7feb22f8a369fe66c29ed215cbe89b966e51.
//
// Solidity: event AttestationSubmitted(address indexed provider, bytes32 attestationHash, uint32 totalAttestations)
func (_BunkerVerification *BunkerVerificationFilterer) FilterAttestationSubmitted(opts *bind.FilterOpts, provider []common.Address) (*BunkerVerificationAttestationSubmittedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "AttestationSubmitted", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationAttestationSubmittedIterator{contract: _BunkerVerification.contract, event: "AttestationSubmitted", logs: logs, sub: sub}, nil
}

// WatchAttestationSubmitted is a free log subscription operation binding the contract event 0x424013c4cbe6b87d14eb842300bd7feb22f8a369fe66c29ed215cbe89b966e51.
//
// Solidity: event AttestationSubmitted(address indexed provider, bytes32 attestationHash, uint32 totalAttestations)
func (_BunkerVerification *BunkerVerificationFilterer) WatchAttestationSubmitted(opts *bind.WatchOpts, sink chan<- *BunkerVerificationAttestationSubmitted, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "AttestationSubmitted", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationAttestationSubmitted)
				if err := _BunkerVerification.contract.UnpackLog(event, "AttestationSubmitted", log); err != nil {
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

// ParseAttestationSubmitted is a log parse operation binding the contract event 0x424013c4cbe6b87d14eb842300bd7feb22f8a369fe66c29ed215cbe89b966e51.
//
// Solidity: event AttestationSubmitted(address indexed provider, bytes32 attestationHash, uint32 totalAttestations)
func (_BunkerVerification *BunkerVerificationFilterer) ParseAttestationSubmitted(log types.Log) (*BunkerVerificationAttestationSubmitted, error) {
	event := new(BunkerVerificationAttestationSubmitted)
	if err := _BunkerVerification.contract.UnpackLog(event, "AttestationSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationMaxMissedAttestationsUpdatedIterator is returned from FilterMaxMissedAttestationsUpdated and is used to iterate over the raw logs and unpacked data for MaxMissedAttestationsUpdated events raised by the BunkerVerification contract.
type BunkerVerificationMaxMissedAttestationsUpdatedIterator struct {
	Event *BunkerVerificationMaxMissedAttestationsUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationMaxMissedAttestationsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationMaxMissedAttestationsUpdated)
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
		it.Event = new(BunkerVerificationMaxMissedAttestationsUpdated)
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
func (it *BunkerVerificationMaxMissedAttestationsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationMaxMissedAttestationsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationMaxMissedAttestationsUpdated represents a MaxMissedAttestationsUpdated event raised by the BunkerVerification contract.
type BunkerVerificationMaxMissedAttestationsUpdated struct {
	OldMax *big.Int
	NewMax *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMaxMissedAttestationsUpdated is a free log retrieval operation binding the contract event 0x5a1cf7b040e3f28a0035107d9e150b6593f46205cb2c131835c5e26141708a64.
//
// Solidity: event MaxMissedAttestationsUpdated(uint256 oldMax, uint256 newMax)
func (_BunkerVerification *BunkerVerificationFilterer) FilterMaxMissedAttestationsUpdated(opts *bind.FilterOpts) (*BunkerVerificationMaxMissedAttestationsUpdatedIterator, error) {

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "MaxMissedAttestationsUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationMaxMissedAttestationsUpdatedIterator{contract: _BunkerVerification.contract, event: "MaxMissedAttestationsUpdated", logs: logs, sub: sub}, nil
}

// WatchMaxMissedAttestationsUpdated is a free log subscription operation binding the contract event 0x5a1cf7b040e3f28a0035107d9e150b6593f46205cb2c131835c5e26141708a64.
//
// Solidity: event MaxMissedAttestationsUpdated(uint256 oldMax, uint256 newMax)
func (_BunkerVerification *BunkerVerificationFilterer) WatchMaxMissedAttestationsUpdated(opts *bind.WatchOpts, sink chan<- *BunkerVerificationMaxMissedAttestationsUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "MaxMissedAttestationsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationMaxMissedAttestationsUpdated)
				if err := _BunkerVerification.contract.UnpackLog(event, "MaxMissedAttestationsUpdated", log); err != nil {
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

// ParseMaxMissedAttestationsUpdated is a log parse operation binding the contract event 0x5a1cf7b040e3f28a0035107d9e150b6593f46205cb2c131835c5e26141708a64.
//
// Solidity: event MaxMissedAttestationsUpdated(uint256 oldMax, uint256 newMax)
func (_BunkerVerification *BunkerVerificationFilterer) ParseMaxMissedAttestationsUpdated(log types.Log) (*BunkerVerificationMaxMissedAttestationsUpdated, error) {
	event := new(BunkerVerificationMaxMissedAttestationsUpdated)
	if err := _BunkerVerification.contract.UnpackLog(event, "MaxMissedAttestationsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the BunkerVerification contract.
type BunkerVerificationOwnershipTransferStartedIterator struct {
	Event *BunkerVerificationOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationOwnershipTransferStarted)
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
		it.Event = new(BunkerVerificationOwnershipTransferStarted)
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
func (it *BunkerVerificationOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the BunkerVerification contract.
type BunkerVerificationOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerVerification *BunkerVerificationFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerVerificationOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationOwnershipTransferStartedIterator{contract: _BunkerVerification.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerVerification *BunkerVerificationFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *BunkerVerificationOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationOwnershipTransferStarted)
				if err := _BunkerVerification.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_BunkerVerification *BunkerVerificationFilterer) ParseOwnershipTransferStarted(log types.Log) (*BunkerVerificationOwnershipTransferStarted, error) {
	event := new(BunkerVerificationOwnershipTransferStarted)
	if err := _BunkerVerification.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BunkerVerification contract.
type BunkerVerificationOwnershipTransferredIterator struct {
	Event *BunkerVerificationOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationOwnershipTransferred)
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
		it.Event = new(BunkerVerificationOwnershipTransferred)
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
func (it *BunkerVerificationOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationOwnershipTransferred represents a OwnershipTransferred event raised by the BunkerVerification contract.
type BunkerVerificationOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerVerification *BunkerVerificationFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerVerificationOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationOwnershipTransferredIterator{contract: _BunkerVerification.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerVerification *BunkerVerificationFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BunkerVerificationOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationOwnershipTransferred)
				if err := _BunkerVerification.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BunkerVerification *BunkerVerificationFilterer) ParseOwnershipTransferred(log types.Log) (*BunkerVerificationOwnershipTransferred, error) {
	event := new(BunkerVerificationOwnershipTransferred)
	if err := _BunkerVerification.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationProviderReinstatedIterator is returned from FilterProviderReinstated and is used to iterate over the raw logs and unpacked data for ProviderReinstated events raised by the BunkerVerification contract.
type BunkerVerificationProviderReinstatedIterator struct {
	Event *BunkerVerificationProviderReinstated // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationProviderReinstatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationProviderReinstated)
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
		it.Event = new(BunkerVerificationProviderReinstated)
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
func (it *BunkerVerificationProviderReinstatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationProviderReinstatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationProviderReinstated represents a ProviderReinstated event raised by the BunkerVerification contract.
type BunkerVerificationProviderReinstated struct {
	Provider common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderReinstated is a free log retrieval operation binding the contract event 0x0b8ceb75fcf9401085a9e463cacfa4e32c097b1c72ed746b9a140e7ddb628e6e.
//
// Solidity: event ProviderReinstated(address indexed provider)
func (_BunkerVerification *BunkerVerificationFilterer) FilterProviderReinstated(opts *bind.FilterOpts, provider []common.Address) (*BunkerVerificationProviderReinstatedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "ProviderReinstated", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationProviderReinstatedIterator{contract: _BunkerVerification.contract, event: "ProviderReinstated", logs: logs, sub: sub}, nil
}

// WatchProviderReinstated is a free log subscription operation binding the contract event 0x0b8ceb75fcf9401085a9e463cacfa4e32c097b1c72ed746b9a140e7ddb628e6e.
//
// Solidity: event ProviderReinstated(address indexed provider)
func (_BunkerVerification *BunkerVerificationFilterer) WatchProviderReinstated(opts *bind.WatchOpts, sink chan<- *BunkerVerificationProviderReinstated, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "ProviderReinstated", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationProviderReinstated)
				if err := _BunkerVerification.contract.UnpackLog(event, "ProviderReinstated", log); err != nil {
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

// ParseProviderReinstated is a log parse operation binding the contract event 0x0b8ceb75fcf9401085a9e463cacfa4e32c097b1c72ed746b9a140e7ddb628e6e.
//
// Solidity: event ProviderReinstated(address indexed provider)
func (_BunkerVerification *BunkerVerificationFilterer) ParseProviderReinstated(log types.Log) (*BunkerVerificationProviderReinstated, error) {
	event := new(BunkerVerificationProviderReinstated)
	if err := _BunkerVerification.contract.UnpackLog(event, "ProviderReinstated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationProviderSuspendedIterator is returned from FilterProviderSuspended and is used to iterate over the raw logs and unpacked data for ProviderSuspended events raised by the BunkerVerification contract.
type BunkerVerificationProviderSuspendedIterator struct {
	Event *BunkerVerificationProviderSuspended // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationProviderSuspendedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationProviderSuspended)
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
		it.Event = new(BunkerVerificationProviderSuspended)
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
func (it *BunkerVerificationProviderSuspendedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationProviderSuspendedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationProviderSuspended represents a ProviderSuspended event raised by the BunkerVerification contract.
type BunkerVerificationProviderSuspended struct {
	Provider    common.Address
	MissedCount uint32
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterProviderSuspended is a free log retrieval operation binding the contract event 0xca4b96ff49bd532b6699fa016a2eb5604552e789c2d54dc64bfc4a3e9beb2072.
//
// Solidity: event ProviderSuspended(address indexed provider, uint32 missedCount)
func (_BunkerVerification *BunkerVerificationFilterer) FilterProviderSuspended(opts *bind.FilterOpts, provider []common.Address) (*BunkerVerificationProviderSuspendedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "ProviderSuspended", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationProviderSuspendedIterator{contract: _BunkerVerification.contract, event: "ProviderSuspended", logs: logs, sub: sub}, nil
}

// WatchProviderSuspended is a free log subscription operation binding the contract event 0xca4b96ff49bd532b6699fa016a2eb5604552e789c2d54dc64bfc4a3e9beb2072.
//
// Solidity: event ProviderSuspended(address indexed provider, uint32 missedCount)
func (_BunkerVerification *BunkerVerificationFilterer) WatchProviderSuspended(opts *bind.WatchOpts, sink chan<- *BunkerVerificationProviderSuspended, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "ProviderSuspended", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationProviderSuspended)
				if err := _BunkerVerification.contract.UnpackLog(event, "ProviderSuspended", log); err != nil {
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

// ParseProviderSuspended is a log parse operation binding the contract event 0xca4b96ff49bd532b6699fa016a2eb5604552e789c2d54dc64bfc4a3e9beb2072.
//
// Solidity: event ProviderSuspended(address indexed provider, uint32 missedCount)
func (_BunkerVerification *BunkerVerificationFilterer) ParseProviderSuspended(log types.Log) (*BunkerVerificationProviderSuspended, error) {
	event := new(BunkerVerificationProviderSuspended)
	if err := _BunkerVerification.contract.UnpackLog(event, "ProviderSuspended", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationReinstatementCooldownUpdatedIterator is returned from FilterReinstatementCooldownUpdated and is used to iterate over the raw logs and unpacked data for ReinstatementCooldownUpdated events raised by the BunkerVerification contract.
type BunkerVerificationReinstatementCooldownUpdatedIterator struct {
	Event *BunkerVerificationReinstatementCooldownUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationReinstatementCooldownUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationReinstatementCooldownUpdated)
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
		it.Event = new(BunkerVerificationReinstatementCooldownUpdated)
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
func (it *BunkerVerificationReinstatementCooldownUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationReinstatementCooldownUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationReinstatementCooldownUpdated represents a ReinstatementCooldownUpdated event raised by the BunkerVerification contract.
type BunkerVerificationReinstatementCooldownUpdated struct {
	NewCooldown *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterReinstatementCooldownUpdated is a free log retrieval operation binding the contract event 0x7ed9b7a5ed787cf9733dfeaa734c68bb16202158613f1c65e62bb45f60425539.
//
// Solidity: event ReinstatementCooldownUpdated(uint256 newCooldown)
func (_BunkerVerification *BunkerVerificationFilterer) FilterReinstatementCooldownUpdated(opts *bind.FilterOpts) (*BunkerVerificationReinstatementCooldownUpdatedIterator, error) {

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "ReinstatementCooldownUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationReinstatementCooldownUpdatedIterator{contract: _BunkerVerification.contract, event: "ReinstatementCooldownUpdated", logs: logs, sub: sub}, nil
}

// WatchReinstatementCooldownUpdated is a free log subscription operation binding the contract event 0x7ed9b7a5ed787cf9733dfeaa734c68bb16202158613f1c65e62bb45f60425539.
//
// Solidity: event ReinstatementCooldownUpdated(uint256 newCooldown)
func (_BunkerVerification *BunkerVerificationFilterer) WatchReinstatementCooldownUpdated(opts *bind.WatchOpts, sink chan<- *BunkerVerificationReinstatementCooldownUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "ReinstatementCooldownUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationReinstatementCooldownUpdated)
				if err := _BunkerVerification.contract.UnpackLog(event, "ReinstatementCooldownUpdated", log); err != nil {
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

// ParseReinstatementCooldownUpdated is a log parse operation binding the contract event 0x7ed9b7a5ed787cf9733dfeaa734c68bb16202158613f1c65e62bb45f60425539.
//
// Solidity: event ReinstatementCooldownUpdated(uint256 newCooldown)
func (_BunkerVerification *BunkerVerificationFilterer) ParseReinstatementCooldownUpdated(log types.Log) (*BunkerVerificationReinstatementCooldownUpdated, error) {
	event := new(BunkerVerificationReinstatementCooldownUpdated)
	if err := _BunkerVerification.contract.UnpackLog(event, "ReinstatementCooldownUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the BunkerVerification contract.
type BunkerVerificationRoleAdminChangedIterator struct {
	Event *BunkerVerificationRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationRoleAdminChanged)
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
		it.Event = new(BunkerVerificationRoleAdminChanged)
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
func (it *BunkerVerificationRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationRoleAdminChanged represents a RoleAdminChanged event raised by the BunkerVerification contract.
type BunkerVerificationRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BunkerVerification *BunkerVerificationFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*BunkerVerificationRoleAdminChangedIterator, error) {

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

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationRoleAdminChangedIterator{contract: _BunkerVerification.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BunkerVerification *BunkerVerificationFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *BunkerVerificationRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationRoleAdminChanged)
				if err := _BunkerVerification.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_BunkerVerification *BunkerVerificationFilterer) ParseRoleAdminChanged(log types.Log) (*BunkerVerificationRoleAdminChanged, error) {
	event := new(BunkerVerificationRoleAdminChanged)
	if err := _BunkerVerification.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the BunkerVerification contract.
type BunkerVerificationRoleGrantedIterator struct {
	Event *BunkerVerificationRoleGranted // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationRoleGranted)
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
		it.Event = new(BunkerVerificationRoleGranted)
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
func (it *BunkerVerificationRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationRoleGranted represents a RoleGranted event raised by the BunkerVerification contract.
type BunkerVerificationRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerVerification *BunkerVerificationFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BunkerVerificationRoleGrantedIterator, error) {

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

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationRoleGrantedIterator{contract: _BunkerVerification.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerVerification *BunkerVerificationFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *BunkerVerificationRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationRoleGranted)
				if err := _BunkerVerification.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_BunkerVerification *BunkerVerificationFilterer) ParseRoleGranted(log types.Log) (*BunkerVerificationRoleGranted, error) {
	event := new(BunkerVerificationRoleGranted)
	if err := _BunkerVerification.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerVerificationRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the BunkerVerification contract.
type BunkerVerificationRoleRevokedIterator struct {
	Event *BunkerVerificationRoleRevoked // Event containing the contract specifics and raw log

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
func (it *BunkerVerificationRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerVerificationRoleRevoked)
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
		it.Event = new(BunkerVerificationRoleRevoked)
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
func (it *BunkerVerificationRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerVerificationRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerVerificationRoleRevoked represents a RoleRevoked event raised by the BunkerVerification contract.
type BunkerVerificationRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerVerification *BunkerVerificationFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BunkerVerificationRoleRevokedIterator, error) {

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

	logs, sub, err := _BunkerVerification.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BunkerVerificationRoleRevokedIterator{contract: _BunkerVerification.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerVerification *BunkerVerificationFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *BunkerVerificationRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _BunkerVerification.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerVerificationRoleRevoked)
				if err := _BunkerVerification.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_BunkerVerification *BunkerVerificationFilterer) ParseRoleRevoked(log types.Log) (*BunkerVerificationRoleRevoked, error) {
	event := new(BunkerVerificationRoleRevoked)
	if err := _BunkerVerification.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
