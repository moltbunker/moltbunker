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

// BunkerEscrowReservation is an auto generated low-level Go binding around an user-defined struct.
type BunkerEscrowReservation struct {
	Requester      common.Address
	TotalAmount    *big.Int
	ReleasedAmount *big.Int
	Duration       *big.Int
	StartTime      *big.Int
	Status         uint8
	Providers      [3]common.Address
}

// BunkerEscrowMetaData contains all meta data concerning the BunkerEscrow contract.
var BunkerEscrowMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_treasury\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"BPS_DENOMINATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_PROTOCOL_FEE_BPS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"OPERATOR_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PROVIDERS_PER_RESERVATION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"calculateProtocolFee\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"fee\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"claim\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createReservation\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"duration\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"feeBurnBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"feeTreasuryBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"finalizeReservation\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getProviders\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address[3]\",\"internalType\":\"address[3]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getReservation\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structBunkerEscrow.Reservation\",\"components\":[{\"name\":\"requester\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"totalAmount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"releasedAmount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"duration\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"startTime\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumBunkerEscrow.Status\"},{\"name\":\"providers\",\"type\":\"address[3]\",\"internalType\":\"address[3]\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"increaseDeposit\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"lowBalanceThresholdBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nextReservationId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"paused\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"protocolFeeBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"refund\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"releasePayment\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"settledDuration\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"reservationFeeBps\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"selectProviders\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"providerAddrs\",\"type\":\"address[3]\",\"internalType\":\"address[3]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setFeeSplit\",\"inputs\":[{\"name\":\"burnBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"treasuryBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setLowBalanceThreshold\",\"inputs\":[{\"name\":\"newThresholdBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setProtocolFee\",\"inputs\":[{\"name\":\"newFeeBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setStakingContract\",\"inputs\":[{\"name\":\"_stakingContract\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTreasury\",\"inputs\":[{\"name\":\"newTreasury\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"settleDispute\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"requesterAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"providerAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"stakingContract\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"token\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalBurned\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalTreasuryFees\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"treasury\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"unpause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"DepositIncreased\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"requester\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"additionalAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newTotal\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"DisputeSettled\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"requesterAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"providerAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"FeeSplitUpdated\",\"inputs\":[{\"name\":\"burnBps\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"treasuryBps\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"LowBalance\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"remaining\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"threshold\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Paused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PaymentReleased\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"grossAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"netToProviders\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"protocolFee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"burnedAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"treasuryAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProtocolFeeUpdated\",\"inputs\":[{\"name\":\"oldFeeBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newFeeBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProvidersSelected\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"provider0\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"provider1\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"provider2\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Refunded\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"requester\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"refundAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReservationCreated\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"requester\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"duration\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReservationFinalized\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"StakingContractUpdated\",\"inputs\":[{\"name\":\"oldStaking\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newStaking\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TreasuryUpdated\",\"inputs\":[{\"name\":\"oldTreasury\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newTreasury\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Unpaused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"AmountOverflow\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"CannotDisableStakingVerification\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DuplicateProvider\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DurationOverflow\",\"inputs\":[{\"name\":\"duration\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"EnforcedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExpectedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidDisputeAmounts\",\"inputs\":[{\"name\":\"requesterAmt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"providerAmt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"available\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidFeeSplit\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidReservation\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidStatus\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"expected\",\"type\":\"uint8\",\"internalType\":\"enumBunkerEscrow.Status\"},{\"name\":\"actual\",\"type\":\"uint8\",\"internalType\":\"enumBunkerEscrow.Status\"}]},{\"type\":\"error\",\"name\":\"NotProvider\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"NotRequester\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"NotRequesterOrOperator\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"NothingToRelease\",\"inputs\":[{\"name\":\"reservationId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ProtocolFeeTooHigh\",\"inputs\":[{\"name\":\"requested\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maximum\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ProviderNotStaked\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"SettledDurationExceedsTotal\",\"inputs\":[{\"name\":\"settled\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"total\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ThresholdTooHigh\",\"inputs\":[{\"name\":\"requested\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"maximum\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]},{\"type\":\"error\",\"name\":\"ZeroAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroDuration\",\"inputs\":[]}]",
}

// BunkerEscrowABI is the input ABI used to generate the binding from.
// Deprecated: Use BunkerEscrowMetaData.ABI instead.
var BunkerEscrowABI = BunkerEscrowMetaData.ABI

// BunkerEscrow is an auto generated Go binding around an Ethereum contract.
type BunkerEscrow struct {
	BunkerEscrowCaller     // Read-only binding to the contract
	BunkerEscrowTransactor // Write-only binding to the contract
	BunkerEscrowFilterer   // Log filterer for contract events
}

// BunkerEscrowCaller is an auto generated read-only Go binding around an Ethereum contract.
type BunkerEscrowCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerEscrowTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BunkerEscrowTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerEscrowFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BunkerEscrowFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerEscrowSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BunkerEscrowSession struct {
	Contract     *BunkerEscrow     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BunkerEscrowCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BunkerEscrowCallerSession struct {
	Contract *BunkerEscrowCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// BunkerEscrowTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BunkerEscrowTransactorSession struct {
	Contract     *BunkerEscrowTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// BunkerEscrowRaw is an auto generated low-level Go binding around an Ethereum contract.
type BunkerEscrowRaw struct {
	Contract *BunkerEscrow // Generic contract binding to access the raw methods on
}

// BunkerEscrowCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BunkerEscrowCallerRaw struct {
	Contract *BunkerEscrowCaller // Generic read-only contract binding to access the raw methods on
}

// BunkerEscrowTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BunkerEscrowTransactorRaw struct {
	Contract *BunkerEscrowTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBunkerEscrow creates a new instance of BunkerEscrow, bound to a specific deployed contract.
func NewBunkerEscrow(address common.Address, backend bind.ContractBackend) (*BunkerEscrow, error) {
	contract, err := bindBunkerEscrow(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrow{BunkerEscrowCaller: BunkerEscrowCaller{contract: contract}, BunkerEscrowTransactor: BunkerEscrowTransactor{contract: contract}, BunkerEscrowFilterer: BunkerEscrowFilterer{contract: contract}}, nil
}

// NewBunkerEscrowCaller creates a new read-only instance of BunkerEscrow, bound to a specific deployed contract.
func NewBunkerEscrowCaller(address common.Address, caller bind.ContractCaller) (*BunkerEscrowCaller, error) {
	contract, err := bindBunkerEscrow(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowCaller{contract: contract}, nil
}

// NewBunkerEscrowTransactor creates a new write-only instance of BunkerEscrow, bound to a specific deployed contract.
func NewBunkerEscrowTransactor(address common.Address, transactor bind.ContractTransactor) (*BunkerEscrowTransactor, error) {
	contract, err := bindBunkerEscrow(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowTransactor{contract: contract}, nil
}

// NewBunkerEscrowFilterer creates a new log filterer instance of BunkerEscrow, bound to a specific deployed contract.
func NewBunkerEscrowFilterer(address common.Address, filterer bind.ContractFilterer) (*BunkerEscrowFilterer, error) {
	contract, err := bindBunkerEscrow(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowFilterer{contract: contract}, nil
}

// bindBunkerEscrow binds a generic wrapper to an already deployed contract.
func bindBunkerEscrow(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BunkerEscrowMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerEscrow *BunkerEscrowRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerEscrow.Contract.BunkerEscrowCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerEscrow *BunkerEscrowRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.BunkerEscrowTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerEscrow *BunkerEscrowRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.BunkerEscrowTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerEscrow *BunkerEscrowCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerEscrow.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerEscrow *BunkerEscrowTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerEscrow *BunkerEscrowTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.contract.Transact(opts, method, params...)
}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCaller) BPSDENOMINATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "BPS_DENOMINATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowSession) BPSDENOMINATOR() (*big.Int, error) {
	return _BunkerEscrow.Contract.BPSDENOMINATOR(&_BunkerEscrow.CallOpts)
}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCallerSession) BPSDENOMINATOR() (*big.Int, error) {
	return _BunkerEscrow.Contract.BPSDENOMINATOR(&_BunkerEscrow.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerEscrow *BunkerEscrowCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerEscrow *BunkerEscrowSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _BunkerEscrow.Contract.DEFAULTADMINROLE(&_BunkerEscrow.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerEscrow *BunkerEscrowCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _BunkerEscrow.Contract.DEFAULTADMINROLE(&_BunkerEscrow.CallOpts)
}

// MAXPROTOCOLFEEBPS is a free data retrieval call binding the contract method 0x6d947e4b.
//
// Solidity: function MAX_PROTOCOL_FEE_BPS() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCaller) MAXPROTOCOLFEEBPS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "MAX_PROTOCOL_FEE_BPS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXPROTOCOLFEEBPS is a free data retrieval call binding the contract method 0x6d947e4b.
//
// Solidity: function MAX_PROTOCOL_FEE_BPS() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowSession) MAXPROTOCOLFEEBPS() (*big.Int, error) {
	return _BunkerEscrow.Contract.MAXPROTOCOLFEEBPS(&_BunkerEscrow.CallOpts)
}

// MAXPROTOCOLFEEBPS is a free data retrieval call binding the contract method 0x6d947e4b.
//
// Solidity: function MAX_PROTOCOL_FEE_BPS() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCallerSession) MAXPROTOCOLFEEBPS() (*big.Int, error) {
	return _BunkerEscrow.Contract.MAXPROTOCOLFEEBPS(&_BunkerEscrow.CallOpts)
}

// OPERATORROLE is a free data retrieval call binding the contract method 0xf5b541a6.
//
// Solidity: function OPERATOR_ROLE() view returns(bytes32)
func (_BunkerEscrow *BunkerEscrowCaller) OPERATORROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "OPERATOR_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OPERATORROLE is a free data retrieval call binding the contract method 0xf5b541a6.
//
// Solidity: function OPERATOR_ROLE() view returns(bytes32)
func (_BunkerEscrow *BunkerEscrowSession) OPERATORROLE() ([32]byte, error) {
	return _BunkerEscrow.Contract.OPERATORROLE(&_BunkerEscrow.CallOpts)
}

// OPERATORROLE is a free data retrieval call binding the contract method 0xf5b541a6.
//
// Solidity: function OPERATOR_ROLE() view returns(bytes32)
func (_BunkerEscrow *BunkerEscrowCallerSession) OPERATORROLE() ([32]byte, error) {
	return _BunkerEscrow.Contract.OPERATORROLE(&_BunkerEscrow.CallOpts)
}

// PROVIDERSPERRESERVATION is a free data retrieval call binding the contract method 0x34f9ef48.
//
// Solidity: function PROVIDERS_PER_RESERVATION() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCaller) PROVIDERSPERRESERVATION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "PROVIDERS_PER_RESERVATION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PROVIDERSPERRESERVATION is a free data retrieval call binding the contract method 0x34f9ef48.
//
// Solidity: function PROVIDERS_PER_RESERVATION() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowSession) PROVIDERSPERRESERVATION() (*big.Int, error) {
	return _BunkerEscrow.Contract.PROVIDERSPERRESERVATION(&_BunkerEscrow.CallOpts)
}

// PROVIDERSPERRESERVATION is a free data retrieval call binding the contract method 0x34f9ef48.
//
// Solidity: function PROVIDERS_PER_RESERVATION() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCallerSession) PROVIDERSPERRESERVATION() (*big.Int, error) {
	return _BunkerEscrow.Contract.PROVIDERSPERRESERVATION(&_BunkerEscrow.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerEscrow *BunkerEscrowCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerEscrow *BunkerEscrowSession) VERSION() (string, error) {
	return _BunkerEscrow.Contract.VERSION(&_BunkerEscrow.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerEscrow *BunkerEscrowCallerSession) VERSION() (string, error) {
	return _BunkerEscrow.Contract.VERSION(&_BunkerEscrow.CallOpts)
}

// CalculateProtocolFee is a free data retrieval call binding the contract method 0x9c7c270c.
//
// Solidity: function calculateProtocolFee(uint256 amount) view returns(uint256 fee)
func (_BunkerEscrow *BunkerEscrowCaller) CalculateProtocolFee(opts *bind.CallOpts, amount *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "calculateProtocolFee", amount)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateProtocolFee is a free data retrieval call binding the contract method 0x9c7c270c.
//
// Solidity: function calculateProtocolFee(uint256 amount) view returns(uint256 fee)
func (_BunkerEscrow *BunkerEscrowSession) CalculateProtocolFee(amount *big.Int) (*big.Int, error) {
	return _BunkerEscrow.Contract.CalculateProtocolFee(&_BunkerEscrow.CallOpts, amount)
}

// CalculateProtocolFee is a free data retrieval call binding the contract method 0x9c7c270c.
//
// Solidity: function calculateProtocolFee(uint256 amount) view returns(uint256 fee)
func (_BunkerEscrow *BunkerEscrowCallerSession) CalculateProtocolFee(amount *big.Int) (*big.Int, error) {
	return _BunkerEscrow.Contract.CalculateProtocolFee(&_BunkerEscrow.CallOpts, amount)
}

// FeeBurnBps is a free data retrieval call binding the contract method 0x1d5514f7.
//
// Solidity: function feeBurnBps() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCaller) FeeBurnBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "feeBurnBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeeBurnBps is a free data retrieval call binding the contract method 0x1d5514f7.
//
// Solidity: function feeBurnBps() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowSession) FeeBurnBps() (*big.Int, error) {
	return _BunkerEscrow.Contract.FeeBurnBps(&_BunkerEscrow.CallOpts)
}

// FeeBurnBps is a free data retrieval call binding the contract method 0x1d5514f7.
//
// Solidity: function feeBurnBps() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCallerSession) FeeBurnBps() (*big.Int, error) {
	return _BunkerEscrow.Contract.FeeBurnBps(&_BunkerEscrow.CallOpts)
}

// FeeTreasuryBps is a free data retrieval call binding the contract method 0x6894fc82.
//
// Solidity: function feeTreasuryBps() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCaller) FeeTreasuryBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "feeTreasuryBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeeTreasuryBps is a free data retrieval call binding the contract method 0x6894fc82.
//
// Solidity: function feeTreasuryBps() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowSession) FeeTreasuryBps() (*big.Int, error) {
	return _BunkerEscrow.Contract.FeeTreasuryBps(&_BunkerEscrow.CallOpts)
}

// FeeTreasuryBps is a free data retrieval call binding the contract method 0x6894fc82.
//
// Solidity: function feeTreasuryBps() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCallerSession) FeeTreasuryBps() (*big.Int, error) {
	return _BunkerEscrow.Contract.FeeTreasuryBps(&_BunkerEscrow.CallOpts)
}

// GetProviders is a free data retrieval call binding the contract method 0xa9f11583.
//
// Solidity: function getProviders(uint256 reservationId) view returns(address[3])
func (_BunkerEscrow *BunkerEscrowCaller) GetProviders(opts *bind.CallOpts, reservationId *big.Int) ([3]common.Address, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "getProviders", reservationId)

	if err != nil {
		return *new([3]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([3]common.Address)).(*[3]common.Address)

	return out0, err

}

// GetProviders is a free data retrieval call binding the contract method 0xa9f11583.
//
// Solidity: function getProviders(uint256 reservationId) view returns(address[3])
func (_BunkerEscrow *BunkerEscrowSession) GetProviders(reservationId *big.Int) ([3]common.Address, error) {
	return _BunkerEscrow.Contract.GetProviders(&_BunkerEscrow.CallOpts, reservationId)
}

// GetProviders is a free data retrieval call binding the contract method 0xa9f11583.
//
// Solidity: function getProviders(uint256 reservationId) view returns(address[3])
func (_BunkerEscrow *BunkerEscrowCallerSession) GetProviders(reservationId *big.Int) ([3]common.Address, error) {
	return _BunkerEscrow.Contract.GetProviders(&_BunkerEscrow.CallOpts, reservationId)
}

// GetReservation is a free data retrieval call binding the contract method 0xd57763c3.
//
// Solidity: function getReservation(uint256 reservationId) view returns((address,uint128,uint128,uint48,uint48,uint8,address[3]))
func (_BunkerEscrow *BunkerEscrowCaller) GetReservation(opts *bind.CallOpts, reservationId *big.Int) (BunkerEscrowReservation, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "getReservation", reservationId)

	if err != nil {
		return *new(BunkerEscrowReservation), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerEscrowReservation)).(*BunkerEscrowReservation)

	return out0, err

}

// GetReservation is a free data retrieval call binding the contract method 0xd57763c3.
//
// Solidity: function getReservation(uint256 reservationId) view returns((address,uint128,uint128,uint48,uint48,uint8,address[3]))
func (_BunkerEscrow *BunkerEscrowSession) GetReservation(reservationId *big.Int) (BunkerEscrowReservation, error) {
	return _BunkerEscrow.Contract.GetReservation(&_BunkerEscrow.CallOpts, reservationId)
}

// GetReservation is a free data retrieval call binding the contract method 0xd57763c3.
//
// Solidity: function getReservation(uint256 reservationId) view returns((address,uint128,uint128,uint48,uint48,uint8,address[3]))
func (_BunkerEscrow *BunkerEscrowCallerSession) GetReservation(reservationId *big.Int) (BunkerEscrowReservation, error) {
	return _BunkerEscrow.Contract.GetReservation(&_BunkerEscrow.CallOpts, reservationId)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerEscrow *BunkerEscrowCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerEscrow *BunkerEscrowSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _BunkerEscrow.Contract.GetRoleAdmin(&_BunkerEscrow.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerEscrow *BunkerEscrowCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _BunkerEscrow.Contract.GetRoleAdmin(&_BunkerEscrow.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerEscrow *BunkerEscrowCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerEscrow *BunkerEscrowSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _BunkerEscrow.Contract.HasRole(&_BunkerEscrow.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerEscrow *BunkerEscrowCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _BunkerEscrow.Contract.HasRole(&_BunkerEscrow.CallOpts, role, account)
}

// LowBalanceThresholdBps is a free data retrieval call binding the contract method 0xd66b5ea6.
//
// Solidity: function lowBalanceThresholdBps() view returns(uint16)
func (_BunkerEscrow *BunkerEscrowCaller) LowBalanceThresholdBps(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "lowBalanceThresholdBps")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// LowBalanceThresholdBps is a free data retrieval call binding the contract method 0xd66b5ea6.
//
// Solidity: function lowBalanceThresholdBps() view returns(uint16)
func (_BunkerEscrow *BunkerEscrowSession) LowBalanceThresholdBps() (uint16, error) {
	return _BunkerEscrow.Contract.LowBalanceThresholdBps(&_BunkerEscrow.CallOpts)
}

// LowBalanceThresholdBps is a free data retrieval call binding the contract method 0xd66b5ea6.
//
// Solidity: function lowBalanceThresholdBps() view returns(uint16)
func (_BunkerEscrow *BunkerEscrowCallerSession) LowBalanceThresholdBps() (uint16, error) {
	return _BunkerEscrow.Contract.LowBalanceThresholdBps(&_BunkerEscrow.CallOpts)
}

// NextReservationId is a free data retrieval call binding the contract method 0xc70c26cd.
//
// Solidity: function nextReservationId() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCaller) NextReservationId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "nextReservationId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NextReservationId is a free data retrieval call binding the contract method 0xc70c26cd.
//
// Solidity: function nextReservationId() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowSession) NextReservationId() (*big.Int, error) {
	return _BunkerEscrow.Contract.NextReservationId(&_BunkerEscrow.CallOpts)
}

// NextReservationId is a free data retrieval call binding the contract method 0xc70c26cd.
//
// Solidity: function nextReservationId() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCallerSession) NextReservationId() (*big.Int, error) {
	return _BunkerEscrow.Contract.NextReservationId(&_BunkerEscrow.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerEscrow *BunkerEscrowCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerEscrow *BunkerEscrowSession) Owner() (common.Address, error) {
	return _BunkerEscrow.Contract.Owner(&_BunkerEscrow.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerEscrow *BunkerEscrowCallerSession) Owner() (common.Address, error) {
	return _BunkerEscrow.Contract.Owner(&_BunkerEscrow.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerEscrow *BunkerEscrowCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerEscrow *BunkerEscrowSession) Paused() (bool, error) {
	return _BunkerEscrow.Contract.Paused(&_BunkerEscrow.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerEscrow *BunkerEscrowCallerSession) Paused() (bool, error) {
	return _BunkerEscrow.Contract.Paused(&_BunkerEscrow.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerEscrow *BunkerEscrowCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerEscrow *BunkerEscrowSession) PendingOwner() (common.Address, error) {
	return _BunkerEscrow.Contract.PendingOwner(&_BunkerEscrow.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerEscrow *BunkerEscrowCallerSession) PendingOwner() (common.Address, error) {
	return _BunkerEscrow.Contract.PendingOwner(&_BunkerEscrow.CallOpts)
}

// ProtocolFeeBps is a free data retrieval call binding the contract method 0x35659fb8.
//
// Solidity: function protocolFeeBps() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCaller) ProtocolFeeBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "protocolFeeBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProtocolFeeBps is a free data retrieval call binding the contract method 0x35659fb8.
//
// Solidity: function protocolFeeBps() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowSession) ProtocolFeeBps() (*big.Int, error) {
	return _BunkerEscrow.Contract.ProtocolFeeBps(&_BunkerEscrow.CallOpts)
}

// ProtocolFeeBps is a free data retrieval call binding the contract method 0x35659fb8.
//
// Solidity: function protocolFeeBps() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCallerSession) ProtocolFeeBps() (*big.Int, error) {
	return _BunkerEscrow.Contract.ProtocolFeeBps(&_BunkerEscrow.CallOpts)
}

// ReservationFeeBps is a free data retrieval call binding the contract method 0x3d8b28ce.
//
// Solidity: function reservationFeeBps(uint256 ) view returns(uint16)
func (_BunkerEscrow *BunkerEscrowCaller) ReservationFeeBps(opts *bind.CallOpts, arg0 *big.Int) (uint16, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "reservationFeeBps", arg0)

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// ReservationFeeBps is a free data retrieval call binding the contract method 0x3d8b28ce.
//
// Solidity: function reservationFeeBps(uint256 ) view returns(uint16)
func (_BunkerEscrow *BunkerEscrowSession) ReservationFeeBps(arg0 *big.Int) (uint16, error) {
	return _BunkerEscrow.Contract.ReservationFeeBps(&_BunkerEscrow.CallOpts, arg0)
}

// ReservationFeeBps is a free data retrieval call binding the contract method 0x3d8b28ce.
//
// Solidity: function reservationFeeBps(uint256 ) view returns(uint16)
func (_BunkerEscrow *BunkerEscrowCallerSession) ReservationFeeBps(arg0 *big.Int) (uint16, error) {
	return _BunkerEscrow.Contract.ReservationFeeBps(&_BunkerEscrow.CallOpts, arg0)
}

// StakingContract is a free data retrieval call binding the contract method 0xee99205c.
//
// Solidity: function stakingContract() view returns(address)
func (_BunkerEscrow *BunkerEscrowCaller) StakingContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "stakingContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakingContract is a free data retrieval call binding the contract method 0xee99205c.
//
// Solidity: function stakingContract() view returns(address)
func (_BunkerEscrow *BunkerEscrowSession) StakingContract() (common.Address, error) {
	return _BunkerEscrow.Contract.StakingContract(&_BunkerEscrow.CallOpts)
}

// StakingContract is a free data retrieval call binding the contract method 0xee99205c.
//
// Solidity: function stakingContract() view returns(address)
func (_BunkerEscrow *BunkerEscrowCallerSession) StakingContract() (common.Address, error) {
	return _BunkerEscrow.Contract.StakingContract(&_BunkerEscrow.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerEscrow *BunkerEscrowCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerEscrow *BunkerEscrowSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _BunkerEscrow.Contract.SupportsInterface(&_BunkerEscrow.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerEscrow *BunkerEscrowCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _BunkerEscrow.Contract.SupportsInterface(&_BunkerEscrow.CallOpts, interfaceId)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_BunkerEscrow *BunkerEscrowCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_BunkerEscrow *BunkerEscrowSession) Token() (common.Address, error) {
	return _BunkerEscrow.Contract.Token(&_BunkerEscrow.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_BunkerEscrow *BunkerEscrowCallerSession) Token() (common.Address, error) {
	return _BunkerEscrow.Contract.Token(&_BunkerEscrow.CallOpts)
}

// TotalBurned is a free data retrieval call binding the contract method 0xd89135cd.
//
// Solidity: function totalBurned() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCaller) TotalBurned(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "totalBurned")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalBurned is a free data retrieval call binding the contract method 0xd89135cd.
//
// Solidity: function totalBurned() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowSession) TotalBurned() (*big.Int, error) {
	return _BunkerEscrow.Contract.TotalBurned(&_BunkerEscrow.CallOpts)
}

// TotalBurned is a free data retrieval call binding the contract method 0xd89135cd.
//
// Solidity: function totalBurned() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCallerSession) TotalBurned() (*big.Int, error) {
	return _BunkerEscrow.Contract.TotalBurned(&_BunkerEscrow.CallOpts)
}

// TotalTreasuryFees is a free data retrieval call binding the contract method 0xea280be3.
//
// Solidity: function totalTreasuryFees() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCaller) TotalTreasuryFees(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "totalTreasuryFees")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalTreasuryFees is a free data retrieval call binding the contract method 0xea280be3.
//
// Solidity: function totalTreasuryFees() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowSession) TotalTreasuryFees() (*big.Int, error) {
	return _BunkerEscrow.Contract.TotalTreasuryFees(&_BunkerEscrow.CallOpts)
}

// TotalTreasuryFees is a free data retrieval call binding the contract method 0xea280be3.
//
// Solidity: function totalTreasuryFees() view returns(uint256)
func (_BunkerEscrow *BunkerEscrowCallerSession) TotalTreasuryFees() (*big.Int, error) {
	return _BunkerEscrow.Contract.TotalTreasuryFees(&_BunkerEscrow.CallOpts)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_BunkerEscrow *BunkerEscrowCaller) Treasury(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerEscrow.contract.Call(opts, &out, "treasury")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_BunkerEscrow *BunkerEscrowSession) Treasury() (common.Address, error) {
	return _BunkerEscrow.Contract.Treasury(&_BunkerEscrow.CallOpts)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_BunkerEscrow *BunkerEscrowCallerSession) Treasury() (common.Address, error) {
	return _BunkerEscrow.Contract.Treasury(&_BunkerEscrow.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerEscrow *BunkerEscrowTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerEscrow *BunkerEscrowSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerEscrow.Contract.AcceptOwnership(&_BunkerEscrow.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerEscrow.Contract.AcceptOwnership(&_BunkerEscrow.TransactOpts)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 reservationId) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) Claim(opts *bind.TransactOpts, reservationId *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "claim", reservationId)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 reservationId) returns()
func (_BunkerEscrow *BunkerEscrowSession) Claim(reservationId *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.Claim(&_BunkerEscrow.TransactOpts, reservationId)
}

// Claim is a paid mutator transaction binding the contract method 0x379607f5.
//
// Solidity: function claim(uint256 reservationId) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) Claim(reservationId *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.Claim(&_BunkerEscrow.TransactOpts, reservationId)
}

// CreateReservation is a paid mutator transaction binding the contract method 0xbee9d69a.
//
// Solidity: function createReservation(uint256 amount, uint256 duration) returns(uint256 reservationId)
func (_BunkerEscrow *BunkerEscrowTransactor) CreateReservation(opts *bind.TransactOpts, amount *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "createReservation", amount, duration)
}

// CreateReservation is a paid mutator transaction binding the contract method 0xbee9d69a.
//
// Solidity: function createReservation(uint256 amount, uint256 duration) returns(uint256 reservationId)
func (_BunkerEscrow *BunkerEscrowSession) CreateReservation(amount *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.CreateReservation(&_BunkerEscrow.TransactOpts, amount, duration)
}

// CreateReservation is a paid mutator transaction binding the contract method 0xbee9d69a.
//
// Solidity: function createReservation(uint256 amount, uint256 duration) returns(uint256 reservationId)
func (_BunkerEscrow *BunkerEscrowTransactorSession) CreateReservation(amount *big.Int, duration *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.CreateReservation(&_BunkerEscrow.TransactOpts, amount, duration)
}

// FinalizeReservation is a paid mutator transaction binding the contract method 0xc0a1826c.
//
// Solidity: function finalizeReservation(uint256 reservationId) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) FinalizeReservation(opts *bind.TransactOpts, reservationId *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "finalizeReservation", reservationId)
}

// FinalizeReservation is a paid mutator transaction binding the contract method 0xc0a1826c.
//
// Solidity: function finalizeReservation(uint256 reservationId) returns()
func (_BunkerEscrow *BunkerEscrowSession) FinalizeReservation(reservationId *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.FinalizeReservation(&_BunkerEscrow.TransactOpts, reservationId)
}

// FinalizeReservation is a paid mutator transaction binding the contract method 0xc0a1826c.
//
// Solidity: function finalizeReservation(uint256 reservationId) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) FinalizeReservation(reservationId *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.FinalizeReservation(&_BunkerEscrow.TransactOpts, reservationId)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerEscrow *BunkerEscrowSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.GrantRole(&_BunkerEscrow.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.GrantRole(&_BunkerEscrow.TransactOpts, role, account)
}

// IncreaseDeposit is a paid mutator transaction binding the contract method 0xfd28e705.
//
// Solidity: function increaseDeposit(uint256 reservationId, uint256 amount) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) IncreaseDeposit(opts *bind.TransactOpts, reservationId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "increaseDeposit", reservationId, amount)
}

// IncreaseDeposit is a paid mutator transaction binding the contract method 0xfd28e705.
//
// Solidity: function increaseDeposit(uint256 reservationId, uint256 amount) returns()
func (_BunkerEscrow *BunkerEscrowSession) IncreaseDeposit(reservationId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.IncreaseDeposit(&_BunkerEscrow.TransactOpts, reservationId, amount)
}

// IncreaseDeposit is a paid mutator transaction binding the contract method 0xfd28e705.
//
// Solidity: function increaseDeposit(uint256 reservationId, uint256 amount) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) IncreaseDeposit(reservationId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.IncreaseDeposit(&_BunkerEscrow.TransactOpts, reservationId, amount)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerEscrow *BunkerEscrowTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerEscrow *BunkerEscrowSession) Pause() (*types.Transaction, error) {
	return _BunkerEscrow.Contract.Pause(&_BunkerEscrow.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) Pause() (*types.Transaction, error) {
	return _BunkerEscrow.Contract.Pause(&_BunkerEscrow.TransactOpts)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 reservationId) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) Refund(opts *bind.TransactOpts, reservationId *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "refund", reservationId)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 reservationId) returns()
func (_BunkerEscrow *BunkerEscrowSession) Refund(reservationId *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.Refund(&_BunkerEscrow.TransactOpts, reservationId)
}

// Refund is a paid mutator transaction binding the contract method 0x278ecde1.
//
// Solidity: function refund(uint256 reservationId) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) Refund(reservationId *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.Refund(&_BunkerEscrow.TransactOpts, reservationId)
}

// ReleasePayment is a paid mutator transaction binding the contract method 0x97e99a80.
//
// Solidity: function releasePayment(uint256 reservationId, uint256 settledDuration) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) ReleasePayment(opts *bind.TransactOpts, reservationId *big.Int, settledDuration *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "releasePayment", reservationId, settledDuration)
}

// ReleasePayment is a paid mutator transaction binding the contract method 0x97e99a80.
//
// Solidity: function releasePayment(uint256 reservationId, uint256 settledDuration) returns()
func (_BunkerEscrow *BunkerEscrowSession) ReleasePayment(reservationId *big.Int, settledDuration *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.ReleasePayment(&_BunkerEscrow.TransactOpts, reservationId, settledDuration)
}

// ReleasePayment is a paid mutator transaction binding the contract method 0x97e99a80.
//
// Solidity: function releasePayment(uint256 reservationId, uint256 settledDuration) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) ReleasePayment(reservationId *big.Int, settledDuration *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.ReleasePayment(&_BunkerEscrow.TransactOpts, reservationId, settledDuration)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerEscrow *BunkerEscrowTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerEscrow *BunkerEscrowSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerEscrow.Contract.RenounceOwnership(&_BunkerEscrow.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerEscrow.Contract.RenounceOwnership(&_BunkerEscrow.TransactOpts)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerEscrow *BunkerEscrowSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.RenounceRole(&_BunkerEscrow.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.RenounceRole(&_BunkerEscrow.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerEscrow *BunkerEscrowSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.RevokeRole(&_BunkerEscrow.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.RevokeRole(&_BunkerEscrow.TransactOpts, role, account)
}

// SelectProviders is a paid mutator transaction binding the contract method 0x202f5388.
//
// Solidity: function selectProviders(uint256 reservationId, address[3] providerAddrs) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) SelectProviders(opts *bind.TransactOpts, reservationId *big.Int, providerAddrs [3]common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "selectProviders", reservationId, providerAddrs)
}

// SelectProviders is a paid mutator transaction binding the contract method 0x202f5388.
//
// Solidity: function selectProviders(uint256 reservationId, address[3] providerAddrs) returns()
func (_BunkerEscrow *BunkerEscrowSession) SelectProviders(reservationId *big.Int, providerAddrs [3]common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SelectProviders(&_BunkerEscrow.TransactOpts, reservationId, providerAddrs)
}

// SelectProviders is a paid mutator transaction binding the contract method 0x202f5388.
//
// Solidity: function selectProviders(uint256 reservationId, address[3] providerAddrs) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) SelectProviders(reservationId *big.Int, providerAddrs [3]common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SelectProviders(&_BunkerEscrow.TransactOpts, reservationId, providerAddrs)
}

// SetFeeSplit is a paid mutator transaction binding the contract method 0xf8978401.
//
// Solidity: function setFeeSplit(uint16 burnBps, uint16 treasuryBps) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) SetFeeSplit(opts *bind.TransactOpts, burnBps uint16, treasuryBps uint16) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "setFeeSplit", burnBps, treasuryBps)
}

// SetFeeSplit is a paid mutator transaction binding the contract method 0xf8978401.
//
// Solidity: function setFeeSplit(uint16 burnBps, uint16 treasuryBps) returns()
func (_BunkerEscrow *BunkerEscrowSession) SetFeeSplit(burnBps uint16, treasuryBps uint16) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SetFeeSplit(&_BunkerEscrow.TransactOpts, burnBps, treasuryBps)
}

// SetFeeSplit is a paid mutator transaction binding the contract method 0xf8978401.
//
// Solidity: function setFeeSplit(uint16 burnBps, uint16 treasuryBps) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) SetFeeSplit(burnBps uint16, treasuryBps uint16) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SetFeeSplit(&_BunkerEscrow.TransactOpts, burnBps, treasuryBps)
}

// SetLowBalanceThreshold is a paid mutator transaction binding the contract method 0xb881f82e.
//
// Solidity: function setLowBalanceThreshold(uint16 newThresholdBps) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) SetLowBalanceThreshold(opts *bind.TransactOpts, newThresholdBps uint16) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "setLowBalanceThreshold", newThresholdBps)
}

// SetLowBalanceThreshold is a paid mutator transaction binding the contract method 0xb881f82e.
//
// Solidity: function setLowBalanceThreshold(uint16 newThresholdBps) returns()
func (_BunkerEscrow *BunkerEscrowSession) SetLowBalanceThreshold(newThresholdBps uint16) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SetLowBalanceThreshold(&_BunkerEscrow.TransactOpts, newThresholdBps)
}

// SetLowBalanceThreshold is a paid mutator transaction binding the contract method 0xb881f82e.
//
// Solidity: function setLowBalanceThreshold(uint16 newThresholdBps) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) SetLowBalanceThreshold(newThresholdBps uint16) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SetLowBalanceThreshold(&_BunkerEscrow.TransactOpts, newThresholdBps)
}

// SetProtocolFee is a paid mutator transaction binding the contract method 0x787dce3d.
//
// Solidity: function setProtocolFee(uint256 newFeeBps) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) SetProtocolFee(opts *bind.TransactOpts, newFeeBps *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "setProtocolFee", newFeeBps)
}

// SetProtocolFee is a paid mutator transaction binding the contract method 0x787dce3d.
//
// Solidity: function setProtocolFee(uint256 newFeeBps) returns()
func (_BunkerEscrow *BunkerEscrowSession) SetProtocolFee(newFeeBps *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SetProtocolFee(&_BunkerEscrow.TransactOpts, newFeeBps)
}

// SetProtocolFee is a paid mutator transaction binding the contract method 0x787dce3d.
//
// Solidity: function setProtocolFee(uint256 newFeeBps) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) SetProtocolFee(newFeeBps *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SetProtocolFee(&_BunkerEscrow.TransactOpts, newFeeBps)
}

// SetStakingContract is a paid mutator transaction binding the contract method 0x9dd373b9.
//
// Solidity: function setStakingContract(address _stakingContract) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) SetStakingContract(opts *bind.TransactOpts, _stakingContract common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "setStakingContract", _stakingContract)
}

// SetStakingContract is a paid mutator transaction binding the contract method 0x9dd373b9.
//
// Solidity: function setStakingContract(address _stakingContract) returns()
func (_BunkerEscrow *BunkerEscrowSession) SetStakingContract(_stakingContract common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SetStakingContract(&_BunkerEscrow.TransactOpts, _stakingContract)
}

// SetStakingContract is a paid mutator transaction binding the contract method 0x9dd373b9.
//
// Solidity: function setStakingContract(address _stakingContract) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) SetStakingContract(_stakingContract common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SetStakingContract(&_BunkerEscrow.TransactOpts, _stakingContract)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address newTreasury) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) SetTreasury(opts *bind.TransactOpts, newTreasury common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "setTreasury", newTreasury)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address newTreasury) returns()
func (_BunkerEscrow *BunkerEscrowSession) SetTreasury(newTreasury common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SetTreasury(&_BunkerEscrow.TransactOpts, newTreasury)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address newTreasury) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) SetTreasury(newTreasury common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SetTreasury(&_BunkerEscrow.TransactOpts, newTreasury)
}

// SettleDispute is a paid mutator transaction binding the contract method 0x8ba64218.
//
// Solidity: function settleDispute(uint256 reservationId, uint256 requesterAmount, uint256 providerAmount) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) SettleDispute(opts *bind.TransactOpts, reservationId *big.Int, requesterAmount *big.Int, providerAmount *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "settleDispute", reservationId, requesterAmount, providerAmount)
}

// SettleDispute is a paid mutator transaction binding the contract method 0x8ba64218.
//
// Solidity: function settleDispute(uint256 reservationId, uint256 requesterAmount, uint256 providerAmount) returns()
func (_BunkerEscrow *BunkerEscrowSession) SettleDispute(reservationId *big.Int, requesterAmount *big.Int, providerAmount *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SettleDispute(&_BunkerEscrow.TransactOpts, reservationId, requesterAmount, providerAmount)
}

// SettleDispute is a paid mutator transaction binding the contract method 0x8ba64218.
//
// Solidity: function settleDispute(uint256 reservationId, uint256 requesterAmount, uint256 providerAmount) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) SettleDispute(reservationId *big.Int, requesterAmount *big.Int, providerAmount *big.Int) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.SettleDispute(&_BunkerEscrow.TransactOpts, reservationId, requesterAmount, providerAmount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerEscrow *BunkerEscrowTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerEscrow *BunkerEscrowSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.TransferOwnership(&_BunkerEscrow.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerEscrow.Contract.TransferOwnership(&_BunkerEscrow.TransactOpts, newOwner)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerEscrow *BunkerEscrowTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerEscrow.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerEscrow *BunkerEscrowSession) Unpause() (*types.Transaction, error) {
	return _BunkerEscrow.Contract.Unpause(&_BunkerEscrow.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerEscrow *BunkerEscrowTransactorSession) Unpause() (*types.Transaction, error) {
	return _BunkerEscrow.Contract.Unpause(&_BunkerEscrow.TransactOpts)
}

// BunkerEscrowDepositIncreasedIterator is returned from FilterDepositIncreased and is used to iterate over the raw logs and unpacked data for DepositIncreased events raised by the BunkerEscrow contract.
type BunkerEscrowDepositIncreasedIterator struct {
	Event *BunkerEscrowDepositIncreased // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowDepositIncreasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowDepositIncreased)
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
		it.Event = new(BunkerEscrowDepositIncreased)
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
func (it *BunkerEscrowDepositIncreasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowDepositIncreasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowDepositIncreased represents a DepositIncreased event raised by the BunkerEscrow contract.
type BunkerEscrowDepositIncreased struct {
	ReservationId    *big.Int
	Requester        common.Address
	AdditionalAmount *big.Int
	NewTotal         *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterDepositIncreased is a free log retrieval operation binding the contract event 0x55d9d468c420c3892a429ebbcdaaf1ac722f8b49bba8b4f0b986cf6103f4fd29.
//
// Solidity: event DepositIncreased(uint256 indexed reservationId, address indexed requester, uint256 additionalAmount, uint256 newTotal)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterDepositIncreased(opts *bind.FilterOpts, reservationId []*big.Int, requester []common.Address) (*BunkerEscrowDepositIncreasedIterator, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "DepositIncreased", reservationIdRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowDepositIncreasedIterator{contract: _BunkerEscrow.contract, event: "DepositIncreased", logs: logs, sub: sub}, nil
}

// WatchDepositIncreased is a free log subscription operation binding the contract event 0x55d9d468c420c3892a429ebbcdaaf1ac722f8b49bba8b4f0b986cf6103f4fd29.
//
// Solidity: event DepositIncreased(uint256 indexed reservationId, address indexed requester, uint256 additionalAmount, uint256 newTotal)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchDepositIncreased(opts *bind.WatchOpts, sink chan<- *BunkerEscrowDepositIncreased, reservationId []*big.Int, requester []common.Address) (event.Subscription, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "DepositIncreased", reservationIdRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowDepositIncreased)
				if err := _BunkerEscrow.contract.UnpackLog(event, "DepositIncreased", log); err != nil {
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

// ParseDepositIncreased is a log parse operation binding the contract event 0x55d9d468c420c3892a429ebbcdaaf1ac722f8b49bba8b4f0b986cf6103f4fd29.
//
// Solidity: event DepositIncreased(uint256 indexed reservationId, address indexed requester, uint256 additionalAmount, uint256 newTotal)
func (_BunkerEscrow *BunkerEscrowFilterer) ParseDepositIncreased(log types.Log) (*BunkerEscrowDepositIncreased, error) {
	event := new(BunkerEscrowDepositIncreased)
	if err := _BunkerEscrow.contract.UnpackLog(event, "DepositIncreased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowDisputeSettledIterator is returned from FilterDisputeSettled and is used to iterate over the raw logs and unpacked data for DisputeSettled events raised by the BunkerEscrow contract.
type BunkerEscrowDisputeSettledIterator struct {
	Event *BunkerEscrowDisputeSettled // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowDisputeSettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowDisputeSettled)
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
		it.Event = new(BunkerEscrowDisputeSettled)
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
func (it *BunkerEscrowDisputeSettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowDisputeSettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowDisputeSettled represents a DisputeSettled event raised by the BunkerEscrow contract.
type BunkerEscrowDisputeSettled struct {
	ReservationId   *big.Int
	RequesterAmount *big.Int
	ProviderAmount  *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterDisputeSettled is a free log retrieval operation binding the contract event 0x3d8191766196c966cf505eceed889b6820ad155ebfa6d53b8b0760d144250a8c.
//
// Solidity: event DisputeSettled(uint256 indexed reservationId, uint256 requesterAmount, uint256 providerAmount)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterDisputeSettled(opts *bind.FilterOpts, reservationId []*big.Int) (*BunkerEscrowDisputeSettledIterator, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "DisputeSettled", reservationIdRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowDisputeSettledIterator{contract: _BunkerEscrow.contract, event: "DisputeSettled", logs: logs, sub: sub}, nil
}

// WatchDisputeSettled is a free log subscription operation binding the contract event 0x3d8191766196c966cf505eceed889b6820ad155ebfa6d53b8b0760d144250a8c.
//
// Solidity: event DisputeSettled(uint256 indexed reservationId, uint256 requesterAmount, uint256 providerAmount)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchDisputeSettled(opts *bind.WatchOpts, sink chan<- *BunkerEscrowDisputeSettled, reservationId []*big.Int) (event.Subscription, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "DisputeSettled", reservationIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowDisputeSettled)
				if err := _BunkerEscrow.contract.UnpackLog(event, "DisputeSettled", log); err != nil {
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

// ParseDisputeSettled is a log parse operation binding the contract event 0x3d8191766196c966cf505eceed889b6820ad155ebfa6d53b8b0760d144250a8c.
//
// Solidity: event DisputeSettled(uint256 indexed reservationId, uint256 requesterAmount, uint256 providerAmount)
func (_BunkerEscrow *BunkerEscrowFilterer) ParseDisputeSettled(log types.Log) (*BunkerEscrowDisputeSettled, error) {
	event := new(BunkerEscrowDisputeSettled)
	if err := _BunkerEscrow.contract.UnpackLog(event, "DisputeSettled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowFeeSplitUpdatedIterator is returned from FilterFeeSplitUpdated and is used to iterate over the raw logs and unpacked data for FeeSplitUpdated events raised by the BunkerEscrow contract.
type BunkerEscrowFeeSplitUpdatedIterator struct {
	Event *BunkerEscrowFeeSplitUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowFeeSplitUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowFeeSplitUpdated)
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
		it.Event = new(BunkerEscrowFeeSplitUpdated)
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
func (it *BunkerEscrowFeeSplitUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowFeeSplitUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowFeeSplitUpdated represents a FeeSplitUpdated event raised by the BunkerEscrow contract.
type BunkerEscrowFeeSplitUpdated struct {
	BurnBps     uint16
	TreasuryBps uint16
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterFeeSplitUpdated is a free log retrieval operation binding the contract event 0x63589f8fbdb3a42e8e990c853b4c0b704151472642497f8e5a435e7547d4b26b.
//
// Solidity: event FeeSplitUpdated(uint16 burnBps, uint16 treasuryBps)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterFeeSplitUpdated(opts *bind.FilterOpts) (*BunkerEscrowFeeSplitUpdatedIterator, error) {

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "FeeSplitUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowFeeSplitUpdatedIterator{contract: _BunkerEscrow.contract, event: "FeeSplitUpdated", logs: logs, sub: sub}, nil
}

// WatchFeeSplitUpdated is a free log subscription operation binding the contract event 0x63589f8fbdb3a42e8e990c853b4c0b704151472642497f8e5a435e7547d4b26b.
//
// Solidity: event FeeSplitUpdated(uint16 burnBps, uint16 treasuryBps)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchFeeSplitUpdated(opts *bind.WatchOpts, sink chan<- *BunkerEscrowFeeSplitUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "FeeSplitUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowFeeSplitUpdated)
				if err := _BunkerEscrow.contract.UnpackLog(event, "FeeSplitUpdated", log); err != nil {
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

// ParseFeeSplitUpdated is a log parse operation binding the contract event 0x63589f8fbdb3a42e8e990c853b4c0b704151472642497f8e5a435e7547d4b26b.
//
// Solidity: event FeeSplitUpdated(uint16 burnBps, uint16 treasuryBps)
func (_BunkerEscrow *BunkerEscrowFilterer) ParseFeeSplitUpdated(log types.Log) (*BunkerEscrowFeeSplitUpdated, error) {
	event := new(BunkerEscrowFeeSplitUpdated)
	if err := _BunkerEscrow.contract.UnpackLog(event, "FeeSplitUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowLowBalanceIterator is returned from FilterLowBalance and is used to iterate over the raw logs and unpacked data for LowBalance events raised by the BunkerEscrow contract.
type BunkerEscrowLowBalanceIterator struct {
	Event *BunkerEscrowLowBalance // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowLowBalanceIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowLowBalance)
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
		it.Event = new(BunkerEscrowLowBalance)
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
func (it *BunkerEscrowLowBalanceIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowLowBalanceIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowLowBalance represents a LowBalance event raised by the BunkerEscrow contract.
type BunkerEscrowLowBalance struct {
	ReservationId *big.Int
	Remaining     *big.Int
	Threshold     *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterLowBalance is a free log retrieval operation binding the contract event 0x08b79c53a755f9dd3a4d2c9461ab4c2eea8a0475de654ec1353add8a72f1b94f.
//
// Solidity: event LowBalance(uint256 indexed reservationId, uint256 remaining, uint256 threshold)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterLowBalance(opts *bind.FilterOpts, reservationId []*big.Int) (*BunkerEscrowLowBalanceIterator, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "LowBalance", reservationIdRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowLowBalanceIterator{contract: _BunkerEscrow.contract, event: "LowBalance", logs: logs, sub: sub}, nil
}

// WatchLowBalance is a free log subscription operation binding the contract event 0x08b79c53a755f9dd3a4d2c9461ab4c2eea8a0475de654ec1353add8a72f1b94f.
//
// Solidity: event LowBalance(uint256 indexed reservationId, uint256 remaining, uint256 threshold)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchLowBalance(opts *bind.WatchOpts, sink chan<- *BunkerEscrowLowBalance, reservationId []*big.Int) (event.Subscription, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "LowBalance", reservationIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowLowBalance)
				if err := _BunkerEscrow.contract.UnpackLog(event, "LowBalance", log); err != nil {
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

// ParseLowBalance is a log parse operation binding the contract event 0x08b79c53a755f9dd3a4d2c9461ab4c2eea8a0475de654ec1353add8a72f1b94f.
//
// Solidity: event LowBalance(uint256 indexed reservationId, uint256 remaining, uint256 threshold)
func (_BunkerEscrow *BunkerEscrowFilterer) ParseLowBalance(log types.Log) (*BunkerEscrowLowBalance, error) {
	event := new(BunkerEscrowLowBalance)
	if err := _BunkerEscrow.contract.UnpackLog(event, "LowBalance", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the BunkerEscrow contract.
type BunkerEscrowOwnershipTransferStartedIterator struct {
	Event *BunkerEscrowOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowOwnershipTransferStarted)
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
		it.Event = new(BunkerEscrowOwnershipTransferStarted)
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
func (it *BunkerEscrowOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the BunkerEscrow contract.
type BunkerEscrowOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerEscrowOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowOwnershipTransferStartedIterator{contract: _BunkerEscrow.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *BunkerEscrowOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowOwnershipTransferStarted)
				if err := _BunkerEscrow.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_BunkerEscrow *BunkerEscrowFilterer) ParseOwnershipTransferStarted(log types.Log) (*BunkerEscrowOwnershipTransferStarted, error) {
	event := new(BunkerEscrowOwnershipTransferStarted)
	if err := _BunkerEscrow.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BunkerEscrow contract.
type BunkerEscrowOwnershipTransferredIterator struct {
	Event *BunkerEscrowOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowOwnershipTransferred)
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
		it.Event = new(BunkerEscrowOwnershipTransferred)
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
func (it *BunkerEscrowOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowOwnershipTransferred represents a OwnershipTransferred event raised by the BunkerEscrow contract.
type BunkerEscrowOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerEscrowOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowOwnershipTransferredIterator{contract: _BunkerEscrow.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BunkerEscrowOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowOwnershipTransferred)
				if err := _BunkerEscrow.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BunkerEscrow *BunkerEscrowFilterer) ParseOwnershipTransferred(log types.Log) (*BunkerEscrowOwnershipTransferred, error) {
	event := new(BunkerEscrowOwnershipTransferred)
	if err := _BunkerEscrow.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the BunkerEscrow contract.
type BunkerEscrowPausedIterator struct {
	Event *BunkerEscrowPaused // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowPaused)
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
		it.Event = new(BunkerEscrowPaused)
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
func (it *BunkerEscrowPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowPaused represents a Paused event raised by the BunkerEscrow contract.
type BunkerEscrowPaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterPaused(opts *bind.FilterOpts) (*BunkerEscrowPausedIterator, error) {

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowPausedIterator{contract: _BunkerEscrow.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *BunkerEscrowPaused) (event.Subscription, error) {

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowPaused)
				if err := _BunkerEscrow.contract.UnpackLog(event, "Paused", log); err != nil {
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
func (_BunkerEscrow *BunkerEscrowFilterer) ParsePaused(log types.Log) (*BunkerEscrowPaused, error) {
	event := new(BunkerEscrowPaused)
	if err := _BunkerEscrow.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowPaymentReleasedIterator is returned from FilterPaymentReleased and is used to iterate over the raw logs and unpacked data for PaymentReleased events raised by the BunkerEscrow contract.
type BunkerEscrowPaymentReleasedIterator struct {
	Event *BunkerEscrowPaymentReleased // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowPaymentReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowPaymentReleased)
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
		it.Event = new(BunkerEscrowPaymentReleased)
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
func (it *BunkerEscrowPaymentReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowPaymentReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowPaymentReleased represents a PaymentReleased event raised by the BunkerEscrow contract.
type BunkerEscrowPaymentReleased struct {
	ReservationId  *big.Int
	GrossAmount    *big.Int
	NetToProviders *big.Int
	ProtocolFee    *big.Int
	BurnedAmount   *big.Int
	TreasuryAmount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterPaymentReleased is a free log retrieval operation binding the contract event 0x28e2a1a217824c8e7d741818f656c447b4f3cdbe3b07e34ac32d9dc13b36b212.
//
// Solidity: event PaymentReleased(uint256 indexed reservationId, uint256 grossAmount, uint256 netToProviders, uint256 protocolFee, uint256 burnedAmount, uint256 treasuryAmount)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterPaymentReleased(opts *bind.FilterOpts, reservationId []*big.Int) (*BunkerEscrowPaymentReleasedIterator, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "PaymentReleased", reservationIdRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowPaymentReleasedIterator{contract: _BunkerEscrow.contract, event: "PaymentReleased", logs: logs, sub: sub}, nil
}

// WatchPaymentReleased is a free log subscription operation binding the contract event 0x28e2a1a217824c8e7d741818f656c447b4f3cdbe3b07e34ac32d9dc13b36b212.
//
// Solidity: event PaymentReleased(uint256 indexed reservationId, uint256 grossAmount, uint256 netToProviders, uint256 protocolFee, uint256 burnedAmount, uint256 treasuryAmount)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchPaymentReleased(opts *bind.WatchOpts, sink chan<- *BunkerEscrowPaymentReleased, reservationId []*big.Int) (event.Subscription, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "PaymentReleased", reservationIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowPaymentReleased)
				if err := _BunkerEscrow.contract.UnpackLog(event, "PaymentReleased", log); err != nil {
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

// ParsePaymentReleased is a log parse operation binding the contract event 0x28e2a1a217824c8e7d741818f656c447b4f3cdbe3b07e34ac32d9dc13b36b212.
//
// Solidity: event PaymentReleased(uint256 indexed reservationId, uint256 grossAmount, uint256 netToProviders, uint256 protocolFee, uint256 burnedAmount, uint256 treasuryAmount)
func (_BunkerEscrow *BunkerEscrowFilterer) ParsePaymentReleased(log types.Log) (*BunkerEscrowPaymentReleased, error) {
	event := new(BunkerEscrowPaymentReleased)
	if err := _BunkerEscrow.contract.UnpackLog(event, "PaymentReleased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowProtocolFeeUpdatedIterator is returned from FilterProtocolFeeUpdated and is used to iterate over the raw logs and unpacked data for ProtocolFeeUpdated events raised by the BunkerEscrow contract.
type BunkerEscrowProtocolFeeUpdatedIterator struct {
	Event *BunkerEscrowProtocolFeeUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowProtocolFeeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowProtocolFeeUpdated)
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
		it.Event = new(BunkerEscrowProtocolFeeUpdated)
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
func (it *BunkerEscrowProtocolFeeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowProtocolFeeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowProtocolFeeUpdated represents a ProtocolFeeUpdated event raised by the BunkerEscrow contract.
type BunkerEscrowProtocolFeeUpdated struct {
	OldFeeBps *big.Int
	NewFeeBps *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterProtocolFeeUpdated is a free log retrieval operation binding the contract event 0xb404cac19fb1cbeff98d325795b08886e3cd8fe8cb1a2f193aac66f13fb239c3.
//
// Solidity: event ProtocolFeeUpdated(uint256 oldFeeBps, uint256 newFeeBps)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterProtocolFeeUpdated(opts *bind.FilterOpts) (*BunkerEscrowProtocolFeeUpdatedIterator, error) {

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "ProtocolFeeUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowProtocolFeeUpdatedIterator{contract: _BunkerEscrow.contract, event: "ProtocolFeeUpdated", logs: logs, sub: sub}, nil
}

// WatchProtocolFeeUpdated is a free log subscription operation binding the contract event 0xb404cac19fb1cbeff98d325795b08886e3cd8fe8cb1a2f193aac66f13fb239c3.
//
// Solidity: event ProtocolFeeUpdated(uint256 oldFeeBps, uint256 newFeeBps)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchProtocolFeeUpdated(opts *bind.WatchOpts, sink chan<- *BunkerEscrowProtocolFeeUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "ProtocolFeeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowProtocolFeeUpdated)
				if err := _BunkerEscrow.contract.UnpackLog(event, "ProtocolFeeUpdated", log); err != nil {
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

// ParseProtocolFeeUpdated is a log parse operation binding the contract event 0xb404cac19fb1cbeff98d325795b08886e3cd8fe8cb1a2f193aac66f13fb239c3.
//
// Solidity: event ProtocolFeeUpdated(uint256 oldFeeBps, uint256 newFeeBps)
func (_BunkerEscrow *BunkerEscrowFilterer) ParseProtocolFeeUpdated(log types.Log) (*BunkerEscrowProtocolFeeUpdated, error) {
	event := new(BunkerEscrowProtocolFeeUpdated)
	if err := _BunkerEscrow.contract.UnpackLog(event, "ProtocolFeeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowProvidersSelectedIterator is returned from FilterProvidersSelected and is used to iterate over the raw logs and unpacked data for ProvidersSelected events raised by the BunkerEscrow contract.
type BunkerEscrowProvidersSelectedIterator struct {
	Event *BunkerEscrowProvidersSelected // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowProvidersSelectedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowProvidersSelected)
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
		it.Event = new(BunkerEscrowProvidersSelected)
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
func (it *BunkerEscrowProvidersSelectedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowProvidersSelectedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowProvidersSelected represents a ProvidersSelected event raised by the BunkerEscrow contract.
type BunkerEscrowProvidersSelected struct {
	ReservationId *big.Int
	Provider0     common.Address
	Provider1     common.Address
	Provider2     common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterProvidersSelected is a free log retrieval operation binding the contract event 0x45e176f25eb06939d4744053c7279ff21fc46c1a62b147a766dd1b7e75123e24.
//
// Solidity: event ProvidersSelected(uint256 indexed reservationId, address provider0, address provider1, address provider2)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterProvidersSelected(opts *bind.FilterOpts, reservationId []*big.Int) (*BunkerEscrowProvidersSelectedIterator, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "ProvidersSelected", reservationIdRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowProvidersSelectedIterator{contract: _BunkerEscrow.contract, event: "ProvidersSelected", logs: logs, sub: sub}, nil
}

// WatchProvidersSelected is a free log subscription operation binding the contract event 0x45e176f25eb06939d4744053c7279ff21fc46c1a62b147a766dd1b7e75123e24.
//
// Solidity: event ProvidersSelected(uint256 indexed reservationId, address provider0, address provider1, address provider2)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchProvidersSelected(opts *bind.WatchOpts, sink chan<- *BunkerEscrowProvidersSelected, reservationId []*big.Int) (event.Subscription, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "ProvidersSelected", reservationIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowProvidersSelected)
				if err := _BunkerEscrow.contract.UnpackLog(event, "ProvidersSelected", log); err != nil {
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

// ParseProvidersSelected is a log parse operation binding the contract event 0x45e176f25eb06939d4744053c7279ff21fc46c1a62b147a766dd1b7e75123e24.
//
// Solidity: event ProvidersSelected(uint256 indexed reservationId, address provider0, address provider1, address provider2)
func (_BunkerEscrow *BunkerEscrowFilterer) ParseProvidersSelected(log types.Log) (*BunkerEscrowProvidersSelected, error) {
	event := new(BunkerEscrowProvidersSelected)
	if err := _BunkerEscrow.contract.UnpackLog(event, "ProvidersSelected", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowRefundedIterator is returned from FilterRefunded and is used to iterate over the raw logs and unpacked data for Refunded events raised by the BunkerEscrow contract.
type BunkerEscrowRefundedIterator struct {
	Event *BunkerEscrowRefunded // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowRefundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowRefunded)
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
		it.Event = new(BunkerEscrowRefunded)
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
func (it *BunkerEscrowRefundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowRefundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowRefunded represents a Refunded event raised by the BunkerEscrow contract.
type BunkerEscrowRefunded struct {
	ReservationId *big.Int
	Requester     common.Address
	RefundAmount  *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterRefunded is a free log retrieval operation binding the contract event 0x7ca5472b7ea78c2c0141c5a12ee6d170cf4ce8ed06be3d22c8252ddfc7a6a2c4.
//
// Solidity: event Refunded(uint256 indexed reservationId, address indexed requester, uint256 refundAmount)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterRefunded(opts *bind.FilterOpts, reservationId []*big.Int, requester []common.Address) (*BunkerEscrowRefundedIterator, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "Refunded", reservationIdRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowRefundedIterator{contract: _BunkerEscrow.contract, event: "Refunded", logs: logs, sub: sub}, nil
}

// WatchRefunded is a free log subscription operation binding the contract event 0x7ca5472b7ea78c2c0141c5a12ee6d170cf4ce8ed06be3d22c8252ddfc7a6a2c4.
//
// Solidity: event Refunded(uint256 indexed reservationId, address indexed requester, uint256 refundAmount)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchRefunded(opts *bind.WatchOpts, sink chan<- *BunkerEscrowRefunded, reservationId []*big.Int, requester []common.Address) (event.Subscription, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "Refunded", reservationIdRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowRefunded)
				if err := _BunkerEscrow.contract.UnpackLog(event, "Refunded", log); err != nil {
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

// ParseRefunded is a log parse operation binding the contract event 0x7ca5472b7ea78c2c0141c5a12ee6d170cf4ce8ed06be3d22c8252ddfc7a6a2c4.
//
// Solidity: event Refunded(uint256 indexed reservationId, address indexed requester, uint256 refundAmount)
func (_BunkerEscrow *BunkerEscrowFilterer) ParseRefunded(log types.Log) (*BunkerEscrowRefunded, error) {
	event := new(BunkerEscrowRefunded)
	if err := _BunkerEscrow.contract.UnpackLog(event, "Refunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowReservationCreatedIterator is returned from FilterReservationCreated and is used to iterate over the raw logs and unpacked data for ReservationCreated events raised by the BunkerEscrow contract.
type BunkerEscrowReservationCreatedIterator struct {
	Event *BunkerEscrowReservationCreated // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowReservationCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowReservationCreated)
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
		it.Event = new(BunkerEscrowReservationCreated)
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
func (it *BunkerEscrowReservationCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowReservationCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowReservationCreated represents a ReservationCreated event raised by the BunkerEscrow contract.
type BunkerEscrowReservationCreated struct {
	ReservationId *big.Int
	Requester     common.Address
	Amount        *big.Int
	Duration      *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterReservationCreated is a free log retrieval operation binding the contract event 0x33461db70ae41500095642332d9a39420948b9b034c6edc55b6f44c9819e68c6.
//
// Solidity: event ReservationCreated(uint256 indexed reservationId, address indexed requester, uint256 amount, uint256 duration)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterReservationCreated(opts *bind.FilterOpts, reservationId []*big.Int, requester []common.Address) (*BunkerEscrowReservationCreatedIterator, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "ReservationCreated", reservationIdRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowReservationCreatedIterator{contract: _BunkerEscrow.contract, event: "ReservationCreated", logs: logs, sub: sub}, nil
}

// WatchReservationCreated is a free log subscription operation binding the contract event 0x33461db70ae41500095642332d9a39420948b9b034c6edc55b6f44c9819e68c6.
//
// Solidity: event ReservationCreated(uint256 indexed reservationId, address indexed requester, uint256 amount, uint256 duration)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchReservationCreated(opts *bind.WatchOpts, sink chan<- *BunkerEscrowReservationCreated, reservationId []*big.Int, requester []common.Address) (event.Subscription, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}
	var requesterRule []interface{}
	for _, requesterItem := range requester {
		requesterRule = append(requesterRule, requesterItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "ReservationCreated", reservationIdRule, requesterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowReservationCreated)
				if err := _BunkerEscrow.contract.UnpackLog(event, "ReservationCreated", log); err != nil {
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

// ParseReservationCreated is a log parse operation binding the contract event 0x33461db70ae41500095642332d9a39420948b9b034c6edc55b6f44c9819e68c6.
//
// Solidity: event ReservationCreated(uint256 indexed reservationId, address indexed requester, uint256 amount, uint256 duration)
func (_BunkerEscrow *BunkerEscrowFilterer) ParseReservationCreated(log types.Log) (*BunkerEscrowReservationCreated, error) {
	event := new(BunkerEscrowReservationCreated)
	if err := _BunkerEscrow.contract.UnpackLog(event, "ReservationCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowReservationFinalizedIterator is returned from FilterReservationFinalized and is used to iterate over the raw logs and unpacked data for ReservationFinalized events raised by the BunkerEscrow contract.
type BunkerEscrowReservationFinalizedIterator struct {
	Event *BunkerEscrowReservationFinalized // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowReservationFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowReservationFinalized)
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
		it.Event = new(BunkerEscrowReservationFinalized)
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
func (it *BunkerEscrowReservationFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowReservationFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowReservationFinalized represents a ReservationFinalized event raised by the BunkerEscrow contract.
type BunkerEscrowReservationFinalized struct {
	ReservationId *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterReservationFinalized is a free log retrieval operation binding the contract event 0xe9a0dea3b9b7ff7aa31f919eb39def4bd959c40dfcef05fd5bc29f85bd90ff7b.
//
// Solidity: event ReservationFinalized(uint256 indexed reservationId)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterReservationFinalized(opts *bind.FilterOpts, reservationId []*big.Int) (*BunkerEscrowReservationFinalizedIterator, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "ReservationFinalized", reservationIdRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowReservationFinalizedIterator{contract: _BunkerEscrow.contract, event: "ReservationFinalized", logs: logs, sub: sub}, nil
}

// WatchReservationFinalized is a free log subscription operation binding the contract event 0xe9a0dea3b9b7ff7aa31f919eb39def4bd959c40dfcef05fd5bc29f85bd90ff7b.
//
// Solidity: event ReservationFinalized(uint256 indexed reservationId)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchReservationFinalized(opts *bind.WatchOpts, sink chan<- *BunkerEscrowReservationFinalized, reservationId []*big.Int) (event.Subscription, error) {

	var reservationIdRule []interface{}
	for _, reservationIdItem := range reservationId {
		reservationIdRule = append(reservationIdRule, reservationIdItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "ReservationFinalized", reservationIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowReservationFinalized)
				if err := _BunkerEscrow.contract.UnpackLog(event, "ReservationFinalized", log); err != nil {
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

// ParseReservationFinalized is a log parse operation binding the contract event 0xe9a0dea3b9b7ff7aa31f919eb39def4bd959c40dfcef05fd5bc29f85bd90ff7b.
//
// Solidity: event ReservationFinalized(uint256 indexed reservationId)
func (_BunkerEscrow *BunkerEscrowFilterer) ParseReservationFinalized(log types.Log) (*BunkerEscrowReservationFinalized, error) {
	event := new(BunkerEscrowReservationFinalized)
	if err := _BunkerEscrow.contract.UnpackLog(event, "ReservationFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the BunkerEscrow contract.
type BunkerEscrowRoleAdminChangedIterator struct {
	Event *BunkerEscrowRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowRoleAdminChanged)
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
		it.Event = new(BunkerEscrowRoleAdminChanged)
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
func (it *BunkerEscrowRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowRoleAdminChanged represents a RoleAdminChanged event raised by the BunkerEscrow contract.
type BunkerEscrowRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*BunkerEscrowRoleAdminChangedIterator, error) {

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

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowRoleAdminChangedIterator{contract: _BunkerEscrow.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *BunkerEscrowRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowRoleAdminChanged)
				if err := _BunkerEscrow.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_BunkerEscrow *BunkerEscrowFilterer) ParseRoleAdminChanged(log types.Log) (*BunkerEscrowRoleAdminChanged, error) {
	event := new(BunkerEscrowRoleAdminChanged)
	if err := _BunkerEscrow.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the BunkerEscrow contract.
type BunkerEscrowRoleGrantedIterator struct {
	Event *BunkerEscrowRoleGranted // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowRoleGranted)
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
		it.Event = new(BunkerEscrowRoleGranted)
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
func (it *BunkerEscrowRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowRoleGranted represents a RoleGranted event raised by the BunkerEscrow contract.
type BunkerEscrowRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BunkerEscrowRoleGrantedIterator, error) {

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

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowRoleGrantedIterator{contract: _BunkerEscrow.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *BunkerEscrowRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowRoleGranted)
				if err := _BunkerEscrow.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_BunkerEscrow *BunkerEscrowFilterer) ParseRoleGranted(log types.Log) (*BunkerEscrowRoleGranted, error) {
	event := new(BunkerEscrowRoleGranted)
	if err := _BunkerEscrow.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the BunkerEscrow contract.
type BunkerEscrowRoleRevokedIterator struct {
	Event *BunkerEscrowRoleRevoked // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowRoleRevoked)
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
		it.Event = new(BunkerEscrowRoleRevoked)
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
func (it *BunkerEscrowRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowRoleRevoked represents a RoleRevoked event raised by the BunkerEscrow contract.
type BunkerEscrowRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BunkerEscrowRoleRevokedIterator, error) {

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

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowRoleRevokedIterator{contract: _BunkerEscrow.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *BunkerEscrowRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowRoleRevoked)
				if err := _BunkerEscrow.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_BunkerEscrow *BunkerEscrowFilterer) ParseRoleRevoked(log types.Log) (*BunkerEscrowRoleRevoked, error) {
	event := new(BunkerEscrowRoleRevoked)
	if err := _BunkerEscrow.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowStakingContractUpdatedIterator is returned from FilterStakingContractUpdated and is used to iterate over the raw logs and unpacked data for StakingContractUpdated events raised by the BunkerEscrow contract.
type BunkerEscrowStakingContractUpdatedIterator struct {
	Event *BunkerEscrowStakingContractUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowStakingContractUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowStakingContractUpdated)
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
		it.Event = new(BunkerEscrowStakingContractUpdated)
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
func (it *BunkerEscrowStakingContractUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowStakingContractUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowStakingContractUpdated represents a StakingContractUpdated event raised by the BunkerEscrow contract.
type BunkerEscrowStakingContractUpdated struct {
	OldStaking common.Address
	NewStaking common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterStakingContractUpdated is a free log retrieval operation binding the contract event 0x7042586b23181180eb30b4798702d7a0233b7fc2551e89806770e8e5d9392e6a.
//
// Solidity: event StakingContractUpdated(address indexed oldStaking, address indexed newStaking)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterStakingContractUpdated(opts *bind.FilterOpts, oldStaking []common.Address, newStaking []common.Address) (*BunkerEscrowStakingContractUpdatedIterator, error) {

	var oldStakingRule []interface{}
	for _, oldStakingItem := range oldStaking {
		oldStakingRule = append(oldStakingRule, oldStakingItem)
	}
	var newStakingRule []interface{}
	for _, newStakingItem := range newStaking {
		newStakingRule = append(newStakingRule, newStakingItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "StakingContractUpdated", oldStakingRule, newStakingRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowStakingContractUpdatedIterator{contract: _BunkerEscrow.contract, event: "StakingContractUpdated", logs: logs, sub: sub}, nil
}

// WatchStakingContractUpdated is a free log subscription operation binding the contract event 0x7042586b23181180eb30b4798702d7a0233b7fc2551e89806770e8e5d9392e6a.
//
// Solidity: event StakingContractUpdated(address indexed oldStaking, address indexed newStaking)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchStakingContractUpdated(opts *bind.WatchOpts, sink chan<- *BunkerEscrowStakingContractUpdated, oldStaking []common.Address, newStaking []common.Address) (event.Subscription, error) {

	var oldStakingRule []interface{}
	for _, oldStakingItem := range oldStaking {
		oldStakingRule = append(oldStakingRule, oldStakingItem)
	}
	var newStakingRule []interface{}
	for _, newStakingItem := range newStaking {
		newStakingRule = append(newStakingRule, newStakingItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "StakingContractUpdated", oldStakingRule, newStakingRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowStakingContractUpdated)
				if err := _BunkerEscrow.contract.UnpackLog(event, "StakingContractUpdated", log); err != nil {
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
func (_BunkerEscrow *BunkerEscrowFilterer) ParseStakingContractUpdated(log types.Log) (*BunkerEscrowStakingContractUpdated, error) {
	event := new(BunkerEscrowStakingContractUpdated)
	if err := _BunkerEscrow.contract.UnpackLog(event, "StakingContractUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowTreasuryUpdatedIterator is returned from FilterTreasuryUpdated and is used to iterate over the raw logs and unpacked data for TreasuryUpdated events raised by the BunkerEscrow contract.
type BunkerEscrowTreasuryUpdatedIterator struct {
	Event *BunkerEscrowTreasuryUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowTreasuryUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowTreasuryUpdated)
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
		it.Event = new(BunkerEscrowTreasuryUpdated)
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
func (it *BunkerEscrowTreasuryUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowTreasuryUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowTreasuryUpdated represents a TreasuryUpdated event raised by the BunkerEscrow contract.
type BunkerEscrowTreasuryUpdated struct {
	OldTreasury common.Address
	NewTreasury common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterTreasuryUpdated is a free log retrieval operation binding the contract event 0x4ab5be82436d353e61ca18726e984e561f5c1cc7c6d38b29d2553c790434705a.
//
// Solidity: event TreasuryUpdated(address indexed oldTreasury, address indexed newTreasury)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterTreasuryUpdated(opts *bind.FilterOpts, oldTreasury []common.Address, newTreasury []common.Address) (*BunkerEscrowTreasuryUpdatedIterator, error) {

	var oldTreasuryRule []interface{}
	for _, oldTreasuryItem := range oldTreasury {
		oldTreasuryRule = append(oldTreasuryRule, oldTreasuryItem)
	}
	var newTreasuryRule []interface{}
	for _, newTreasuryItem := range newTreasury {
		newTreasuryRule = append(newTreasuryRule, newTreasuryItem)
	}

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "TreasuryUpdated", oldTreasuryRule, newTreasuryRule)
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowTreasuryUpdatedIterator{contract: _BunkerEscrow.contract, event: "TreasuryUpdated", logs: logs, sub: sub}, nil
}

// WatchTreasuryUpdated is a free log subscription operation binding the contract event 0x4ab5be82436d353e61ca18726e984e561f5c1cc7c6d38b29d2553c790434705a.
//
// Solidity: event TreasuryUpdated(address indexed oldTreasury, address indexed newTreasury)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchTreasuryUpdated(opts *bind.WatchOpts, sink chan<- *BunkerEscrowTreasuryUpdated, oldTreasury []common.Address, newTreasury []common.Address) (event.Subscription, error) {

	var oldTreasuryRule []interface{}
	for _, oldTreasuryItem := range oldTreasury {
		oldTreasuryRule = append(oldTreasuryRule, oldTreasuryItem)
	}
	var newTreasuryRule []interface{}
	for _, newTreasuryItem := range newTreasury {
		newTreasuryRule = append(newTreasuryRule, newTreasuryItem)
	}

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "TreasuryUpdated", oldTreasuryRule, newTreasuryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowTreasuryUpdated)
				if err := _BunkerEscrow.contract.UnpackLog(event, "TreasuryUpdated", log); err != nil {
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

// ParseTreasuryUpdated is a log parse operation binding the contract event 0x4ab5be82436d353e61ca18726e984e561f5c1cc7c6d38b29d2553c790434705a.
//
// Solidity: event TreasuryUpdated(address indexed oldTreasury, address indexed newTreasury)
func (_BunkerEscrow *BunkerEscrowFilterer) ParseTreasuryUpdated(log types.Log) (*BunkerEscrowTreasuryUpdated, error) {
	event := new(BunkerEscrowTreasuryUpdated)
	if err := _BunkerEscrow.contract.UnpackLog(event, "TreasuryUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerEscrowUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the BunkerEscrow contract.
type BunkerEscrowUnpausedIterator struct {
	Event *BunkerEscrowUnpaused // Event containing the contract specifics and raw log

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
func (it *BunkerEscrowUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerEscrowUnpaused)
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
		it.Event = new(BunkerEscrowUnpaused)
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
func (it *BunkerEscrowUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerEscrowUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerEscrowUnpaused represents a Unpaused event raised by the BunkerEscrow contract.
type BunkerEscrowUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_BunkerEscrow *BunkerEscrowFilterer) FilterUnpaused(opts *bind.FilterOpts) (*BunkerEscrowUnpausedIterator, error) {

	logs, sub, err := _BunkerEscrow.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &BunkerEscrowUnpausedIterator{contract: _BunkerEscrow.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_BunkerEscrow *BunkerEscrowFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *BunkerEscrowUnpaused) (event.Subscription, error) {

	logs, sub, err := _BunkerEscrow.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerEscrowUnpaused)
				if err := _BunkerEscrow.contract.UnpackLog(event, "Unpaused", log); err != nil {
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
func (_BunkerEscrow *BunkerEscrowFilterer) ParseUnpaused(log types.Log) (*BunkerEscrowUnpaused, error) {
	event := new(BunkerEscrowUnpaused)
	if err := _BunkerEscrow.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
