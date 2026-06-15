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

// BunkerRegistryMetaData contains all meta data concerning the BunkerRegistry contract.
var BunkerRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_treasury\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_registrationFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_stakingContract\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"BPS_DENOMINATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"BURN_BPS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_AVATAR_URL_LENGTH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_BULK_SIZE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_DESCRIPTION_LENGTH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_NAMES_PER_OWNER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MIN_REGISTRATION_FEE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PREMIUM_1_CHAR_MULTIPLIER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PREMIUM_2_CHAR_MULTIPLIER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PREMIUM_3_CHAR_MULTIPLIER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"PREMIUM_4_CHAR_MULTIPLIER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"batchReserveNames\",\"inputs\":[{\"name\":\"names\",\"type\":\"string[]\",\"internalType\":\"string[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"bulkRegister\",\"inputs\":[{\"name\":\"names\",\"type\":\"string[]\",\"internalType\":\"string[]\"},{\"name\":\"deploymentIDs\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"bulkRenew\",\"inputs\":[{\"name\":\"names\",\"type\":\"string[]\",\"internalType\":\"string[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"bunkerToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"calculatePrice\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"cancelReservation\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"changeFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"claimReservation\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentID\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"expirationPeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"gracePeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isAvailable\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isExpired\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isInGracePeriod\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"metadata\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"description\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"avatarURL\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nameCount\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nameOf\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ownedNameAt\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"paused\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"primaryName\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"reclaimSquatted\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"referralDiscountBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"referralRewardBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"register\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentID\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerWithReferral\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"deploymentID\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"referrer\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registrationFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"release\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renew\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"reservationPeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"reserve\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"reservedNames\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"resolve\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"deploymentID\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registeredAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"reverseResolve\",\"inputs\":[{\"name\":\"deploymentID\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setChangeFee\",\"inputs\":[{\"name\":\"fee\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setExpirationPeriod\",\"inputs\":[{\"name\":\"period\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setGracePeriod\",\"inputs\":[{\"name\":\"period\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMetadata\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"description\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"avatarURL\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setPrimaryName\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setReferralDiscountBps\",\"inputs\":[{\"name\":\"bps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setReferralRewardBps\",\"inputs\":[{\"name\":\"bps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setRegistrationFee\",\"inputs\":[{\"name\":\"newFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setReservationPeriod\",\"inputs\":[{\"name\":\"period\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setReservedName\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"reserved\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setShortNamesEnabled\",\"inputs\":[{\"name\":\"enabled\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSquattingGracePeriod\",\"inputs\":[{\"name\":\"period\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setStakingContract\",\"inputs\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTreasury\",\"inputs\":[{\"name\":\"newTreasury\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"shortNamesEnabled\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"squattingGracePeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"stakingContract\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIBunkerStakingTier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"subdomains\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"deploymentID\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"registeredAt\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"expiresAt\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"reservedUntil\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"referrer\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transfer\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"treasury\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"unpause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateDeployment\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"newDeploymentID\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"ChangeFeeUpdated\",\"inputs\":[{\"name\":\"oldFee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newFee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ExpirationPeriodUpdated\",\"inputs\":[{\"name\":\"oldPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"GracePeriodUpdated\",\"inputs\":[{\"name\":\"oldPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MetadataUpdated\",\"inputs\":[{\"name\":\"nameIndexed\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Paused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PrimaryNameSet\",\"inputs\":[{\"name\":\"deploymentID\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReferralDiscountUpdated\",\"inputs\":[{\"name\":\"oldBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReferralRewardUpdated\",\"inputs\":[{\"name\":\"oldBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RegistrationFeeUpdated\",\"inputs\":[{\"name\":\"oldFee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newFee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReservationCancelled\",\"inputs\":[{\"name\":\"nameIndexed\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReservationClaimed\",\"inputs\":[{\"name\":\"nameIndexed\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"deploymentID\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReservationPeriodUpdated\",\"inputs\":[{\"name\":\"oldPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ReservedNameUpdated\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"reserved\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ShortNamesEnabledUpdated\",\"inputs\":[{\"name\":\"enabled\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SquattedNameReclaimed\",\"inputs\":[{\"name\":\"nameIndexed\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"reclaimer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SquattingGracePeriodUpdated\",\"inputs\":[{\"name\":\"oldPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"StakingContractUpdated\",\"inputs\":[{\"name\":\"oldAddr\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"newAddr\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SubdomainRegistered\",\"inputs\":[{\"name\":\"nameIndexed\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"deploymentID\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"fee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SubdomainReleased\",\"inputs\":[{\"name\":\"nameIndexed\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SubdomainRenewed\",\"inputs\":[{\"name\":\"nameIndexed\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newExpiry\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"},{\"name\":\"fee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SubdomainReserved\",\"inputs\":[{\"name\":\"nameIndexed\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"reserver\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"reservedUntil\",\"type\":\"uint48\",\"indexed\":false,\"internalType\":\"uint48\"},{\"name\":\"fee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SubdomainTransferred\",\"inputs\":[{\"name\":\"nameIndexed\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"from\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SubdomainUpdated\",\"inputs\":[{\"name\":\"nameIndexed\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"name\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"},{\"name\":\"oldDeploymentID\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"newDeploymentID\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TreasuryUpdated\",\"inputs\":[{\"name\":\"oldTreasury\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"newTreasury\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Unpaused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"ArrayLengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ArrayTooLarge\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"max\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"CannotTransferToSelf\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"DeploymentNotOwned\",\"inputs\":[{\"name\":\"deploymentID\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"EnforcedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExpectedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"FeeBelowMinimum\",\"inputs\":[{\"name\":\"fee\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minimum\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"FeeTransferFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidDeploymentID\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidName\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"InvalidPeriod\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidReferrer\",\"inputs\":[{\"name\":\"referrer\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"MetadataAvatarURLTooLong\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"max\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"MetadataDescriptionTooLong\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"max\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"NameAlreadyRegistered\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"NameExpired\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"NameInGracePeriod\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"NameNotExpired\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"NameNotRegistered\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"NameNotSquatted\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"NameReserved\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"NotNameOwner\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"caller\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"NotReservationOwner\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"caller\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReservationExpired\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ShortNamesDisabled\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TooManyNames\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"max\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}]",
}

// BunkerRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use BunkerRegistryMetaData.ABI instead.
var BunkerRegistryABI = BunkerRegistryMetaData.ABI

// BunkerRegistry is an auto generated Go binding around an Ethereum contract.
type BunkerRegistry struct {
	BunkerRegistryCaller     // Read-only binding to the contract
	BunkerRegistryTransactor // Write-only binding to the contract
	BunkerRegistryFilterer   // Log filterer for contract events
}

// BunkerRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type BunkerRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BunkerRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BunkerRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BunkerRegistrySession struct {
	Contract     *BunkerRegistry   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BunkerRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BunkerRegistryCallerSession struct {
	Contract *BunkerRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// BunkerRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BunkerRegistryTransactorSession struct {
	Contract     *BunkerRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// BunkerRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type BunkerRegistryRaw struct {
	Contract *BunkerRegistry // Generic contract binding to access the raw methods on
}

// BunkerRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BunkerRegistryCallerRaw struct {
	Contract *BunkerRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// BunkerRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BunkerRegistryTransactorRaw struct {
	Contract *BunkerRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBunkerRegistry creates a new instance of BunkerRegistry, bound to a specific deployed contract.
func NewBunkerRegistry(address common.Address, backend bind.ContractBackend) (*BunkerRegistry, error) {
	contract, err := bindBunkerRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistry{BunkerRegistryCaller: BunkerRegistryCaller{contract: contract}, BunkerRegistryTransactor: BunkerRegistryTransactor{contract: contract}, BunkerRegistryFilterer: BunkerRegistryFilterer{contract: contract}}, nil
}

// NewBunkerRegistryCaller creates a new read-only instance of BunkerRegistry, bound to a specific deployed contract.
func NewBunkerRegistryCaller(address common.Address, caller bind.ContractCaller) (*BunkerRegistryCaller, error) {
	contract, err := bindBunkerRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryCaller{contract: contract}, nil
}

// NewBunkerRegistryTransactor creates a new write-only instance of BunkerRegistry, bound to a specific deployed contract.
func NewBunkerRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*BunkerRegistryTransactor, error) {
	contract, err := bindBunkerRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryTransactor{contract: contract}, nil
}

// NewBunkerRegistryFilterer creates a new log filterer instance of BunkerRegistry, bound to a specific deployed contract.
func NewBunkerRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*BunkerRegistryFilterer, error) {
	contract, err := bindBunkerRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryFilterer{contract: contract}, nil
}

// bindBunkerRegistry binds a generic wrapper to an already deployed contract.
func bindBunkerRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BunkerRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerRegistry *BunkerRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerRegistry.Contract.BunkerRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerRegistry *BunkerRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.BunkerRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerRegistry *BunkerRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.BunkerRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerRegistry *BunkerRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerRegistry *BunkerRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerRegistry *BunkerRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.contract.Transact(opts, method, params...)
}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) BPSDENOMINATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "BPS_DENOMINATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) BPSDENOMINATOR() (*big.Int, error) {
	return _BunkerRegistry.Contract.BPSDENOMINATOR(&_BunkerRegistry.CallOpts)
}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) BPSDENOMINATOR() (*big.Int, error) {
	return _BunkerRegistry.Contract.BPSDENOMINATOR(&_BunkerRegistry.CallOpts)
}

// BURNBPS is a free data retrieval call binding the contract method 0xa37a9fc0.
//
// Solidity: function BURN_BPS() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) BURNBPS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "BURN_BPS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BURNBPS is a free data retrieval call binding the contract method 0xa37a9fc0.
//
// Solidity: function BURN_BPS() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) BURNBPS() (*big.Int, error) {
	return _BunkerRegistry.Contract.BURNBPS(&_BunkerRegistry.CallOpts)
}

// BURNBPS is a free data retrieval call binding the contract method 0xa37a9fc0.
//
// Solidity: function BURN_BPS() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) BURNBPS() (*big.Int, error) {
	return _BunkerRegistry.Contract.BURNBPS(&_BunkerRegistry.CallOpts)
}

// MAXAVATARURLLENGTH is a free data retrieval call binding the contract method 0x8dbffb12.
//
// Solidity: function MAX_AVATAR_URL_LENGTH() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) MAXAVATARURLLENGTH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "MAX_AVATAR_URL_LENGTH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXAVATARURLLENGTH is a free data retrieval call binding the contract method 0x8dbffb12.
//
// Solidity: function MAX_AVATAR_URL_LENGTH() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) MAXAVATARURLLENGTH() (*big.Int, error) {
	return _BunkerRegistry.Contract.MAXAVATARURLLENGTH(&_BunkerRegistry.CallOpts)
}

// MAXAVATARURLLENGTH is a free data retrieval call binding the contract method 0x8dbffb12.
//
// Solidity: function MAX_AVATAR_URL_LENGTH() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) MAXAVATARURLLENGTH() (*big.Int, error) {
	return _BunkerRegistry.Contract.MAXAVATARURLLENGTH(&_BunkerRegistry.CallOpts)
}

// MAXBULKSIZE is a free data retrieval call binding the contract method 0xc3225549.
//
// Solidity: function MAX_BULK_SIZE() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) MAXBULKSIZE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "MAX_BULK_SIZE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXBULKSIZE is a free data retrieval call binding the contract method 0xc3225549.
//
// Solidity: function MAX_BULK_SIZE() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) MAXBULKSIZE() (*big.Int, error) {
	return _BunkerRegistry.Contract.MAXBULKSIZE(&_BunkerRegistry.CallOpts)
}

// MAXBULKSIZE is a free data retrieval call binding the contract method 0xc3225549.
//
// Solidity: function MAX_BULK_SIZE() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) MAXBULKSIZE() (*big.Int, error) {
	return _BunkerRegistry.Contract.MAXBULKSIZE(&_BunkerRegistry.CallOpts)
}

// MAXDESCRIPTIONLENGTH is a free data retrieval call binding the contract method 0x9201ea0a.
//
// Solidity: function MAX_DESCRIPTION_LENGTH() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) MAXDESCRIPTIONLENGTH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "MAX_DESCRIPTION_LENGTH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXDESCRIPTIONLENGTH is a free data retrieval call binding the contract method 0x9201ea0a.
//
// Solidity: function MAX_DESCRIPTION_LENGTH() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) MAXDESCRIPTIONLENGTH() (*big.Int, error) {
	return _BunkerRegistry.Contract.MAXDESCRIPTIONLENGTH(&_BunkerRegistry.CallOpts)
}

// MAXDESCRIPTIONLENGTH is a free data retrieval call binding the contract method 0x9201ea0a.
//
// Solidity: function MAX_DESCRIPTION_LENGTH() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) MAXDESCRIPTIONLENGTH() (*big.Int, error) {
	return _BunkerRegistry.Contract.MAXDESCRIPTIONLENGTH(&_BunkerRegistry.CallOpts)
}

// MAXNAMESPEROWNER is a free data retrieval call binding the contract method 0xe33ea45e.
//
// Solidity: function MAX_NAMES_PER_OWNER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) MAXNAMESPEROWNER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "MAX_NAMES_PER_OWNER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXNAMESPEROWNER is a free data retrieval call binding the contract method 0xe33ea45e.
//
// Solidity: function MAX_NAMES_PER_OWNER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) MAXNAMESPEROWNER() (*big.Int, error) {
	return _BunkerRegistry.Contract.MAXNAMESPEROWNER(&_BunkerRegistry.CallOpts)
}

// MAXNAMESPEROWNER is a free data retrieval call binding the contract method 0xe33ea45e.
//
// Solidity: function MAX_NAMES_PER_OWNER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) MAXNAMESPEROWNER() (*big.Int, error) {
	return _BunkerRegistry.Contract.MAXNAMESPEROWNER(&_BunkerRegistry.CallOpts)
}

// MINREGISTRATIONFEE is a free data retrieval call binding the contract method 0x418cc2ae.
//
// Solidity: function MIN_REGISTRATION_FEE() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) MINREGISTRATIONFEE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "MIN_REGISTRATION_FEE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINREGISTRATIONFEE is a free data retrieval call binding the contract method 0x418cc2ae.
//
// Solidity: function MIN_REGISTRATION_FEE() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) MINREGISTRATIONFEE() (*big.Int, error) {
	return _BunkerRegistry.Contract.MINREGISTRATIONFEE(&_BunkerRegistry.CallOpts)
}

// MINREGISTRATIONFEE is a free data retrieval call binding the contract method 0x418cc2ae.
//
// Solidity: function MIN_REGISTRATION_FEE() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) MINREGISTRATIONFEE() (*big.Int, error) {
	return _BunkerRegistry.Contract.MINREGISTRATIONFEE(&_BunkerRegistry.CallOpts)
}

// PREMIUM1CHARMULTIPLIER is a free data retrieval call binding the contract method 0x0b436a20.
//
// Solidity: function PREMIUM_1_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) PREMIUM1CHARMULTIPLIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "PREMIUM_1_CHAR_MULTIPLIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PREMIUM1CHARMULTIPLIER is a free data retrieval call binding the contract method 0x0b436a20.
//
// Solidity: function PREMIUM_1_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) PREMIUM1CHARMULTIPLIER() (*big.Int, error) {
	return _BunkerRegistry.Contract.PREMIUM1CHARMULTIPLIER(&_BunkerRegistry.CallOpts)
}

// PREMIUM1CHARMULTIPLIER is a free data retrieval call binding the contract method 0x0b436a20.
//
// Solidity: function PREMIUM_1_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) PREMIUM1CHARMULTIPLIER() (*big.Int, error) {
	return _BunkerRegistry.Contract.PREMIUM1CHARMULTIPLIER(&_BunkerRegistry.CallOpts)
}

// PREMIUM2CHARMULTIPLIER is a free data retrieval call binding the contract method 0xe420f838.
//
// Solidity: function PREMIUM_2_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) PREMIUM2CHARMULTIPLIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "PREMIUM_2_CHAR_MULTIPLIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PREMIUM2CHARMULTIPLIER is a free data retrieval call binding the contract method 0xe420f838.
//
// Solidity: function PREMIUM_2_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) PREMIUM2CHARMULTIPLIER() (*big.Int, error) {
	return _BunkerRegistry.Contract.PREMIUM2CHARMULTIPLIER(&_BunkerRegistry.CallOpts)
}

// PREMIUM2CHARMULTIPLIER is a free data retrieval call binding the contract method 0xe420f838.
//
// Solidity: function PREMIUM_2_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) PREMIUM2CHARMULTIPLIER() (*big.Int, error) {
	return _BunkerRegistry.Contract.PREMIUM2CHARMULTIPLIER(&_BunkerRegistry.CallOpts)
}

// PREMIUM3CHARMULTIPLIER is a free data retrieval call binding the contract method 0xb7f096a1.
//
// Solidity: function PREMIUM_3_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) PREMIUM3CHARMULTIPLIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "PREMIUM_3_CHAR_MULTIPLIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PREMIUM3CHARMULTIPLIER is a free data retrieval call binding the contract method 0xb7f096a1.
//
// Solidity: function PREMIUM_3_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) PREMIUM3CHARMULTIPLIER() (*big.Int, error) {
	return _BunkerRegistry.Contract.PREMIUM3CHARMULTIPLIER(&_BunkerRegistry.CallOpts)
}

// PREMIUM3CHARMULTIPLIER is a free data retrieval call binding the contract method 0xb7f096a1.
//
// Solidity: function PREMIUM_3_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) PREMIUM3CHARMULTIPLIER() (*big.Int, error) {
	return _BunkerRegistry.Contract.PREMIUM3CHARMULTIPLIER(&_BunkerRegistry.CallOpts)
}

// PREMIUM4CHARMULTIPLIER is a free data retrieval call binding the contract method 0x6e6902c8.
//
// Solidity: function PREMIUM_4_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) PREMIUM4CHARMULTIPLIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "PREMIUM_4_CHAR_MULTIPLIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PREMIUM4CHARMULTIPLIER is a free data retrieval call binding the contract method 0x6e6902c8.
//
// Solidity: function PREMIUM_4_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) PREMIUM4CHARMULTIPLIER() (*big.Int, error) {
	return _BunkerRegistry.Contract.PREMIUM4CHARMULTIPLIER(&_BunkerRegistry.CallOpts)
}

// PREMIUM4CHARMULTIPLIER is a free data retrieval call binding the contract method 0x6e6902c8.
//
// Solidity: function PREMIUM_4_CHAR_MULTIPLIER() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) PREMIUM4CHARMULTIPLIER() (*big.Int, error) {
	return _BunkerRegistry.Contract.PREMIUM4CHARMULTIPLIER(&_BunkerRegistry.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerRegistry *BunkerRegistryCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerRegistry *BunkerRegistrySession) VERSION() (string, error) {
	return _BunkerRegistry.Contract.VERSION(&_BunkerRegistry.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerRegistry *BunkerRegistryCallerSession) VERSION() (string, error) {
	return _BunkerRegistry.Contract.VERSION(&_BunkerRegistry.CallOpts)
}

// BunkerToken is a free data retrieval call binding the contract method 0x0b38d902.
//
// Solidity: function bunkerToken() view returns(address)
func (_BunkerRegistry *BunkerRegistryCaller) BunkerToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "bunkerToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BunkerToken is a free data retrieval call binding the contract method 0x0b38d902.
//
// Solidity: function bunkerToken() view returns(address)
func (_BunkerRegistry *BunkerRegistrySession) BunkerToken() (common.Address, error) {
	return _BunkerRegistry.Contract.BunkerToken(&_BunkerRegistry.CallOpts)
}

// BunkerToken is a free data retrieval call binding the contract method 0x0b38d902.
//
// Solidity: function bunkerToken() view returns(address)
func (_BunkerRegistry *BunkerRegistryCallerSession) BunkerToken() (common.Address, error) {
	return _BunkerRegistry.Contract.BunkerToken(&_BunkerRegistry.CallOpts)
}

// CalculatePrice is a free data retrieval call binding the contract method 0xf8a3fce1.
//
// Solidity: function calculatePrice(string name, address user) view returns(uint256 price)
func (_BunkerRegistry *BunkerRegistryCaller) CalculatePrice(opts *bind.CallOpts, name string, user common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "calculatePrice", name, user)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculatePrice is a free data retrieval call binding the contract method 0xf8a3fce1.
//
// Solidity: function calculatePrice(string name, address user) view returns(uint256 price)
func (_BunkerRegistry *BunkerRegistrySession) CalculatePrice(name string, user common.Address) (*big.Int, error) {
	return _BunkerRegistry.Contract.CalculatePrice(&_BunkerRegistry.CallOpts, name, user)
}

// CalculatePrice is a free data retrieval call binding the contract method 0xf8a3fce1.
//
// Solidity: function calculatePrice(string name, address user) view returns(uint256 price)
func (_BunkerRegistry *BunkerRegistryCallerSession) CalculatePrice(name string, user common.Address) (*big.Int, error) {
	return _BunkerRegistry.Contract.CalculatePrice(&_BunkerRegistry.CallOpts, name, user)
}

// ChangeFee is a free data retrieval call binding the contract method 0x0040ff6c.
//
// Solidity: function changeFee() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) ChangeFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "changeFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ChangeFee is a free data retrieval call binding the contract method 0x0040ff6c.
//
// Solidity: function changeFee() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) ChangeFee() (*big.Int, error) {
	return _BunkerRegistry.Contract.ChangeFee(&_BunkerRegistry.CallOpts)
}

// ChangeFee is a free data retrieval call binding the contract method 0x0040ff6c.
//
// Solidity: function changeFee() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) ChangeFee() (*big.Int, error) {
	return _BunkerRegistry.Contract.ChangeFee(&_BunkerRegistry.CallOpts)
}

// ExpirationPeriod is a free data retrieval call binding the contract method 0x8897cad3.
//
// Solidity: function expirationPeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) ExpirationPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "expirationPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ExpirationPeriod is a free data retrieval call binding the contract method 0x8897cad3.
//
// Solidity: function expirationPeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) ExpirationPeriod() (*big.Int, error) {
	return _BunkerRegistry.Contract.ExpirationPeriod(&_BunkerRegistry.CallOpts)
}

// ExpirationPeriod is a free data retrieval call binding the contract method 0x8897cad3.
//
// Solidity: function expirationPeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) ExpirationPeriod() (*big.Int, error) {
	return _BunkerRegistry.Contract.ExpirationPeriod(&_BunkerRegistry.CallOpts)
}

// GracePeriod is a free data retrieval call binding the contract method 0xa06db7dc.
//
// Solidity: function gracePeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) GracePeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "gracePeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GracePeriod is a free data retrieval call binding the contract method 0xa06db7dc.
//
// Solidity: function gracePeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) GracePeriod() (*big.Int, error) {
	return _BunkerRegistry.Contract.GracePeriod(&_BunkerRegistry.CallOpts)
}

// GracePeriod is a free data retrieval call binding the contract method 0xa06db7dc.
//
// Solidity: function gracePeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) GracePeriod() (*big.Int, error) {
	return _BunkerRegistry.Contract.GracePeriod(&_BunkerRegistry.CallOpts)
}

// IsAvailable is a free data retrieval call binding the contract method 0x965306aa.
//
// Solidity: function isAvailable(string name) view returns(bool)
func (_BunkerRegistry *BunkerRegistryCaller) IsAvailable(opts *bind.CallOpts, name string) (bool, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "isAvailable", name)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAvailable is a free data retrieval call binding the contract method 0x965306aa.
//
// Solidity: function isAvailable(string name) view returns(bool)
func (_BunkerRegistry *BunkerRegistrySession) IsAvailable(name string) (bool, error) {
	return _BunkerRegistry.Contract.IsAvailable(&_BunkerRegistry.CallOpts, name)
}

// IsAvailable is a free data retrieval call binding the contract method 0x965306aa.
//
// Solidity: function isAvailable(string name) view returns(bool)
func (_BunkerRegistry *BunkerRegistryCallerSession) IsAvailable(name string) (bool, error) {
	return _BunkerRegistry.Contract.IsAvailable(&_BunkerRegistry.CallOpts, name)
}

// IsExpired is a free data retrieval call binding the contract method 0xc64fafbc.
//
// Solidity: function isExpired(string name) view returns(bool)
func (_BunkerRegistry *BunkerRegistryCaller) IsExpired(opts *bind.CallOpts, name string) (bool, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "isExpired", name)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsExpired is a free data retrieval call binding the contract method 0xc64fafbc.
//
// Solidity: function isExpired(string name) view returns(bool)
func (_BunkerRegistry *BunkerRegistrySession) IsExpired(name string) (bool, error) {
	return _BunkerRegistry.Contract.IsExpired(&_BunkerRegistry.CallOpts, name)
}

// IsExpired is a free data retrieval call binding the contract method 0xc64fafbc.
//
// Solidity: function isExpired(string name) view returns(bool)
func (_BunkerRegistry *BunkerRegistryCallerSession) IsExpired(name string) (bool, error) {
	return _BunkerRegistry.Contract.IsExpired(&_BunkerRegistry.CallOpts, name)
}

// IsInGracePeriod is a free data retrieval call binding the contract method 0x1499651f.
//
// Solidity: function isInGracePeriod(string name) view returns(bool)
func (_BunkerRegistry *BunkerRegistryCaller) IsInGracePeriod(opts *bind.CallOpts, name string) (bool, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "isInGracePeriod", name)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsInGracePeriod is a free data retrieval call binding the contract method 0x1499651f.
//
// Solidity: function isInGracePeriod(string name) view returns(bool)
func (_BunkerRegistry *BunkerRegistrySession) IsInGracePeriod(name string) (bool, error) {
	return _BunkerRegistry.Contract.IsInGracePeriod(&_BunkerRegistry.CallOpts, name)
}

// IsInGracePeriod is a free data retrieval call binding the contract method 0x1499651f.
//
// Solidity: function isInGracePeriod(string name) view returns(bool)
func (_BunkerRegistry *BunkerRegistryCallerSession) IsInGracePeriod(name string) (bool, error) {
	return _BunkerRegistry.Contract.IsInGracePeriod(&_BunkerRegistry.CallOpts, name)
}

// Metadata is a free data retrieval call binding the contract method 0x7122ba06.
//
// Solidity: function metadata(bytes32 ) view returns(string description, string avatarURL)
func (_BunkerRegistry *BunkerRegistryCaller) Metadata(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Description string
	AvatarURL   string
}, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "metadata", arg0)

	outstruct := new(struct {
		Description string
		AvatarURL   string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Description = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.AvatarURL = *abi.ConvertType(out[1], new(string)).(*string)

	return *outstruct, err

}

// Metadata is a free data retrieval call binding the contract method 0x7122ba06.
//
// Solidity: function metadata(bytes32 ) view returns(string description, string avatarURL)
func (_BunkerRegistry *BunkerRegistrySession) Metadata(arg0 [32]byte) (struct {
	Description string
	AvatarURL   string
}, error) {
	return _BunkerRegistry.Contract.Metadata(&_BunkerRegistry.CallOpts, arg0)
}

// Metadata is a free data retrieval call binding the contract method 0x7122ba06.
//
// Solidity: function metadata(bytes32 ) view returns(string description, string avatarURL)
func (_BunkerRegistry *BunkerRegistryCallerSession) Metadata(arg0 [32]byte) (struct {
	Description string
	AvatarURL   string
}, error) {
	return _BunkerRegistry.Contract.Metadata(&_BunkerRegistry.CallOpts, arg0)
}

// NameCount is a free data retrieval call binding the contract method 0x79bba73b.
//
// Solidity: function nameCount(address owner) view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) NameCount(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "nameCount", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NameCount is a free data retrieval call binding the contract method 0x79bba73b.
//
// Solidity: function nameCount(address owner) view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) NameCount(owner common.Address) (*big.Int, error) {
	return _BunkerRegistry.Contract.NameCount(&_BunkerRegistry.CallOpts, owner)
}

// NameCount is a free data retrieval call binding the contract method 0x79bba73b.
//
// Solidity: function nameCount(address owner) view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) NameCount(owner common.Address) (*big.Int, error) {
	return _BunkerRegistry.Contract.NameCount(&_BunkerRegistry.CallOpts, owner)
}

// NameOf is a free data retrieval call binding the contract method 0xe532a034.
//
// Solidity: function nameOf(bytes32 ) view returns(string)
func (_BunkerRegistry *BunkerRegistryCaller) NameOf(opts *bind.CallOpts, arg0 [32]byte) (string, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "nameOf", arg0)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// NameOf is a free data retrieval call binding the contract method 0xe532a034.
//
// Solidity: function nameOf(bytes32 ) view returns(string)
func (_BunkerRegistry *BunkerRegistrySession) NameOf(arg0 [32]byte) (string, error) {
	return _BunkerRegistry.Contract.NameOf(&_BunkerRegistry.CallOpts, arg0)
}

// NameOf is a free data retrieval call binding the contract method 0xe532a034.
//
// Solidity: function nameOf(bytes32 ) view returns(string)
func (_BunkerRegistry *BunkerRegistryCallerSession) NameOf(arg0 [32]byte) (string, error) {
	return _BunkerRegistry.Contract.NameOf(&_BunkerRegistry.CallOpts, arg0)
}

// OwnedNameAt is a free data retrieval call binding the contract method 0x7059c7b8.
//
// Solidity: function ownedNameAt(address owner, uint256 index) view returns(bytes32)
func (_BunkerRegistry *BunkerRegistryCaller) OwnedNameAt(opts *bind.CallOpts, owner common.Address, index *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "ownedNameAt", owner, index)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// OwnedNameAt is a free data retrieval call binding the contract method 0x7059c7b8.
//
// Solidity: function ownedNameAt(address owner, uint256 index) view returns(bytes32)
func (_BunkerRegistry *BunkerRegistrySession) OwnedNameAt(owner common.Address, index *big.Int) ([32]byte, error) {
	return _BunkerRegistry.Contract.OwnedNameAt(&_BunkerRegistry.CallOpts, owner, index)
}

// OwnedNameAt is a free data retrieval call binding the contract method 0x7059c7b8.
//
// Solidity: function ownedNameAt(address owner, uint256 index) view returns(bytes32)
func (_BunkerRegistry *BunkerRegistryCallerSession) OwnedNameAt(owner common.Address, index *big.Int) ([32]byte, error) {
	return _BunkerRegistry.Contract.OwnedNameAt(&_BunkerRegistry.CallOpts, owner, index)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerRegistry *BunkerRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerRegistry *BunkerRegistrySession) Owner() (common.Address, error) {
	return _BunkerRegistry.Contract.Owner(&_BunkerRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerRegistry *BunkerRegistryCallerSession) Owner() (common.Address, error) {
	return _BunkerRegistry.Contract.Owner(&_BunkerRegistry.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerRegistry *BunkerRegistryCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerRegistry *BunkerRegistrySession) Paused() (bool, error) {
	return _BunkerRegistry.Contract.Paused(&_BunkerRegistry.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerRegistry *BunkerRegistryCallerSession) Paused() (bool, error) {
	return _BunkerRegistry.Contract.Paused(&_BunkerRegistry.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerRegistry *BunkerRegistryCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerRegistry *BunkerRegistrySession) PendingOwner() (common.Address, error) {
	return _BunkerRegistry.Contract.PendingOwner(&_BunkerRegistry.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerRegistry *BunkerRegistryCallerSession) PendingOwner() (common.Address, error) {
	return _BunkerRegistry.Contract.PendingOwner(&_BunkerRegistry.CallOpts)
}

// PrimaryName is a free data retrieval call binding the contract method 0x8f87f7a8.
//
// Solidity: function primaryName(bytes32 ) view returns(bytes32)
func (_BunkerRegistry *BunkerRegistryCaller) PrimaryName(opts *bind.CallOpts, arg0 [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "primaryName", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PrimaryName is a free data retrieval call binding the contract method 0x8f87f7a8.
//
// Solidity: function primaryName(bytes32 ) view returns(bytes32)
func (_BunkerRegistry *BunkerRegistrySession) PrimaryName(arg0 [32]byte) ([32]byte, error) {
	return _BunkerRegistry.Contract.PrimaryName(&_BunkerRegistry.CallOpts, arg0)
}

// PrimaryName is a free data retrieval call binding the contract method 0x8f87f7a8.
//
// Solidity: function primaryName(bytes32 ) view returns(bytes32)
func (_BunkerRegistry *BunkerRegistryCallerSession) PrimaryName(arg0 [32]byte) ([32]byte, error) {
	return _BunkerRegistry.Contract.PrimaryName(&_BunkerRegistry.CallOpts, arg0)
}

// ReferralDiscountBps is a free data retrieval call binding the contract method 0x30ab6943.
//
// Solidity: function referralDiscountBps() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) ReferralDiscountBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "referralDiscountBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ReferralDiscountBps is a free data retrieval call binding the contract method 0x30ab6943.
//
// Solidity: function referralDiscountBps() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) ReferralDiscountBps() (*big.Int, error) {
	return _BunkerRegistry.Contract.ReferralDiscountBps(&_BunkerRegistry.CallOpts)
}

// ReferralDiscountBps is a free data retrieval call binding the contract method 0x30ab6943.
//
// Solidity: function referralDiscountBps() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) ReferralDiscountBps() (*big.Int, error) {
	return _BunkerRegistry.Contract.ReferralDiscountBps(&_BunkerRegistry.CallOpts)
}

// ReferralRewardBps is a free data retrieval call binding the contract method 0x75f4c059.
//
// Solidity: function referralRewardBps() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) ReferralRewardBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "referralRewardBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ReferralRewardBps is a free data retrieval call binding the contract method 0x75f4c059.
//
// Solidity: function referralRewardBps() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) ReferralRewardBps() (*big.Int, error) {
	return _BunkerRegistry.Contract.ReferralRewardBps(&_BunkerRegistry.CallOpts)
}

// ReferralRewardBps is a free data retrieval call binding the contract method 0x75f4c059.
//
// Solidity: function referralRewardBps() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) ReferralRewardBps() (*big.Int, error) {
	return _BunkerRegistry.Contract.ReferralRewardBps(&_BunkerRegistry.CallOpts)
}

// RegistrationFee is a free data retrieval call binding the contract method 0x14c44e09.
//
// Solidity: function registrationFee() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) RegistrationFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "registrationFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RegistrationFee is a free data retrieval call binding the contract method 0x14c44e09.
//
// Solidity: function registrationFee() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) RegistrationFee() (*big.Int, error) {
	return _BunkerRegistry.Contract.RegistrationFee(&_BunkerRegistry.CallOpts)
}

// RegistrationFee is a free data retrieval call binding the contract method 0x14c44e09.
//
// Solidity: function registrationFee() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) RegistrationFee() (*big.Int, error) {
	return _BunkerRegistry.Contract.RegistrationFee(&_BunkerRegistry.CallOpts)
}

// ReservationPeriod is a free data retrieval call binding the contract method 0x89af5e60.
//
// Solidity: function reservationPeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) ReservationPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "reservationPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ReservationPeriod is a free data retrieval call binding the contract method 0x89af5e60.
//
// Solidity: function reservationPeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) ReservationPeriod() (*big.Int, error) {
	return _BunkerRegistry.Contract.ReservationPeriod(&_BunkerRegistry.CallOpts)
}

// ReservationPeriod is a free data retrieval call binding the contract method 0x89af5e60.
//
// Solidity: function reservationPeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) ReservationPeriod() (*big.Int, error) {
	return _BunkerRegistry.Contract.ReservationPeriod(&_BunkerRegistry.CallOpts)
}

// ReservedNames is a free data retrieval call binding the contract method 0x389c0748.
//
// Solidity: function reservedNames(bytes32 ) view returns(bool)
func (_BunkerRegistry *BunkerRegistryCaller) ReservedNames(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "reservedNames", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ReservedNames is a free data retrieval call binding the contract method 0x389c0748.
//
// Solidity: function reservedNames(bytes32 ) view returns(bool)
func (_BunkerRegistry *BunkerRegistrySession) ReservedNames(arg0 [32]byte) (bool, error) {
	return _BunkerRegistry.Contract.ReservedNames(&_BunkerRegistry.CallOpts, arg0)
}

// ReservedNames is a free data retrieval call binding the contract method 0x389c0748.
//
// Solidity: function reservedNames(bytes32 ) view returns(bool)
func (_BunkerRegistry *BunkerRegistryCallerSession) ReservedNames(arg0 [32]byte) (bool, error) {
	return _BunkerRegistry.Contract.ReservedNames(&_BunkerRegistry.CallOpts, arg0)
}

// Resolve is a free data retrieval call binding the contract method 0x461a4478.
//
// Solidity: function resolve(string name) view returns(address owner, bytes32 deploymentID, uint256 registeredAt)
func (_BunkerRegistry *BunkerRegistryCaller) Resolve(opts *bind.CallOpts, name string) (struct {
	Owner        common.Address
	DeploymentID [32]byte
	RegisteredAt *big.Int
}, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "resolve", name)

	outstruct := new(struct {
		Owner        common.Address
		DeploymentID [32]byte
		RegisteredAt *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Owner = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.DeploymentID = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)
	outstruct.RegisteredAt = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Resolve is a free data retrieval call binding the contract method 0x461a4478.
//
// Solidity: function resolve(string name) view returns(address owner, bytes32 deploymentID, uint256 registeredAt)
func (_BunkerRegistry *BunkerRegistrySession) Resolve(name string) (struct {
	Owner        common.Address
	DeploymentID [32]byte
	RegisteredAt *big.Int
}, error) {
	return _BunkerRegistry.Contract.Resolve(&_BunkerRegistry.CallOpts, name)
}

// Resolve is a free data retrieval call binding the contract method 0x461a4478.
//
// Solidity: function resolve(string name) view returns(address owner, bytes32 deploymentID, uint256 registeredAt)
func (_BunkerRegistry *BunkerRegistryCallerSession) Resolve(name string) (struct {
	Owner        common.Address
	DeploymentID [32]byte
	RegisteredAt *big.Int
}, error) {
	return _BunkerRegistry.Contract.Resolve(&_BunkerRegistry.CallOpts, name)
}

// ReverseResolve is a free data retrieval call binding the contract method 0x6315f9d0.
//
// Solidity: function reverseResolve(bytes32 deploymentID) view returns(string name)
func (_BunkerRegistry *BunkerRegistryCaller) ReverseResolve(opts *bind.CallOpts, deploymentID [32]byte) (string, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "reverseResolve", deploymentID)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// ReverseResolve is a free data retrieval call binding the contract method 0x6315f9d0.
//
// Solidity: function reverseResolve(bytes32 deploymentID) view returns(string name)
func (_BunkerRegistry *BunkerRegistrySession) ReverseResolve(deploymentID [32]byte) (string, error) {
	return _BunkerRegistry.Contract.ReverseResolve(&_BunkerRegistry.CallOpts, deploymentID)
}

// ReverseResolve is a free data retrieval call binding the contract method 0x6315f9d0.
//
// Solidity: function reverseResolve(bytes32 deploymentID) view returns(string name)
func (_BunkerRegistry *BunkerRegistryCallerSession) ReverseResolve(deploymentID [32]byte) (string, error) {
	return _BunkerRegistry.Contract.ReverseResolve(&_BunkerRegistry.CallOpts, deploymentID)
}

// ShortNamesEnabled is a free data retrieval call binding the contract method 0xbae7e108.
//
// Solidity: function shortNamesEnabled() view returns(bool)
func (_BunkerRegistry *BunkerRegistryCaller) ShortNamesEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "shortNamesEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ShortNamesEnabled is a free data retrieval call binding the contract method 0xbae7e108.
//
// Solidity: function shortNamesEnabled() view returns(bool)
func (_BunkerRegistry *BunkerRegistrySession) ShortNamesEnabled() (bool, error) {
	return _BunkerRegistry.Contract.ShortNamesEnabled(&_BunkerRegistry.CallOpts)
}

// ShortNamesEnabled is a free data retrieval call binding the contract method 0xbae7e108.
//
// Solidity: function shortNamesEnabled() view returns(bool)
func (_BunkerRegistry *BunkerRegistryCallerSession) ShortNamesEnabled() (bool, error) {
	return _BunkerRegistry.Contract.ShortNamesEnabled(&_BunkerRegistry.CallOpts)
}

// SquattingGracePeriod is a free data retrieval call binding the contract method 0xa5b67776.
//
// Solidity: function squattingGracePeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCaller) SquattingGracePeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "squattingGracePeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SquattingGracePeriod is a free data retrieval call binding the contract method 0xa5b67776.
//
// Solidity: function squattingGracePeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistrySession) SquattingGracePeriod() (*big.Int, error) {
	return _BunkerRegistry.Contract.SquattingGracePeriod(&_BunkerRegistry.CallOpts)
}

// SquattingGracePeriod is a free data retrieval call binding the contract method 0xa5b67776.
//
// Solidity: function squattingGracePeriod() view returns(uint256)
func (_BunkerRegistry *BunkerRegistryCallerSession) SquattingGracePeriod() (*big.Int, error) {
	return _BunkerRegistry.Contract.SquattingGracePeriod(&_BunkerRegistry.CallOpts)
}

// StakingContract is a free data retrieval call binding the contract method 0xee99205c.
//
// Solidity: function stakingContract() view returns(address)
func (_BunkerRegistry *BunkerRegistryCaller) StakingContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "stakingContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakingContract is a free data retrieval call binding the contract method 0xee99205c.
//
// Solidity: function stakingContract() view returns(address)
func (_BunkerRegistry *BunkerRegistrySession) StakingContract() (common.Address, error) {
	return _BunkerRegistry.Contract.StakingContract(&_BunkerRegistry.CallOpts)
}

// StakingContract is a free data retrieval call binding the contract method 0xee99205c.
//
// Solidity: function stakingContract() view returns(address)
func (_BunkerRegistry *BunkerRegistryCallerSession) StakingContract() (common.Address, error) {
	return _BunkerRegistry.Contract.StakingContract(&_BunkerRegistry.CallOpts)
}

// Subdomains is a free data retrieval call binding the contract method 0x9d79d081.
//
// Solidity: function subdomains(bytes32 ) view returns(address owner, bytes32 deploymentID, uint48 registeredAt, uint48 expiresAt, uint48 reservedUntil, address referrer)
func (_BunkerRegistry *BunkerRegistryCaller) Subdomains(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Owner         common.Address
	DeploymentID  [32]byte
	RegisteredAt  *big.Int
	ExpiresAt     *big.Int
	ReservedUntil *big.Int
	Referrer      common.Address
}, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "subdomains", arg0)

	outstruct := new(struct {
		Owner         common.Address
		DeploymentID  [32]byte
		RegisteredAt  *big.Int
		ExpiresAt     *big.Int
		ReservedUntil *big.Int
		Referrer      common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Owner = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.DeploymentID = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)
	outstruct.RegisteredAt = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.ExpiresAt = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.ReservedUntil = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.Referrer = *abi.ConvertType(out[5], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// Subdomains is a free data retrieval call binding the contract method 0x9d79d081.
//
// Solidity: function subdomains(bytes32 ) view returns(address owner, bytes32 deploymentID, uint48 registeredAt, uint48 expiresAt, uint48 reservedUntil, address referrer)
func (_BunkerRegistry *BunkerRegistrySession) Subdomains(arg0 [32]byte) (struct {
	Owner         common.Address
	DeploymentID  [32]byte
	RegisteredAt  *big.Int
	ExpiresAt     *big.Int
	ReservedUntil *big.Int
	Referrer      common.Address
}, error) {
	return _BunkerRegistry.Contract.Subdomains(&_BunkerRegistry.CallOpts, arg0)
}

// Subdomains is a free data retrieval call binding the contract method 0x9d79d081.
//
// Solidity: function subdomains(bytes32 ) view returns(address owner, bytes32 deploymentID, uint48 registeredAt, uint48 expiresAt, uint48 reservedUntil, address referrer)
func (_BunkerRegistry *BunkerRegistryCallerSession) Subdomains(arg0 [32]byte) (struct {
	Owner         common.Address
	DeploymentID  [32]byte
	RegisteredAt  *big.Int
	ExpiresAt     *big.Int
	ReservedUntil *big.Int
	Referrer      common.Address
}, error) {
	return _BunkerRegistry.Contract.Subdomains(&_BunkerRegistry.CallOpts, arg0)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_BunkerRegistry *BunkerRegistryCaller) Treasury(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerRegistry.contract.Call(opts, &out, "treasury")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_BunkerRegistry *BunkerRegistrySession) Treasury() (common.Address, error) {
	return _BunkerRegistry.Contract.Treasury(&_BunkerRegistry.CallOpts)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_BunkerRegistry *BunkerRegistryCallerSession) Treasury() (common.Address, error) {
	return _BunkerRegistry.Contract.Treasury(&_BunkerRegistry.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerRegistry *BunkerRegistryTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerRegistry *BunkerRegistrySession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerRegistry.Contract.AcceptOwnership(&_BunkerRegistry.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerRegistry.Contract.AcceptOwnership(&_BunkerRegistry.TransactOpts)
}

// BatchReserveNames is a paid mutator transaction binding the contract method 0x559e6d80.
//
// Solidity: function batchReserveNames(string[] names) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) BatchReserveNames(opts *bind.TransactOpts, names []string) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "batchReserveNames", names)
}

// BatchReserveNames is a paid mutator transaction binding the contract method 0x559e6d80.
//
// Solidity: function batchReserveNames(string[] names) returns()
func (_BunkerRegistry *BunkerRegistrySession) BatchReserveNames(names []string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.BatchReserveNames(&_BunkerRegistry.TransactOpts, names)
}

// BatchReserveNames is a paid mutator transaction binding the contract method 0x559e6d80.
//
// Solidity: function batchReserveNames(string[] names) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) BatchReserveNames(names []string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.BatchReserveNames(&_BunkerRegistry.TransactOpts, names)
}

// BulkRegister is a paid mutator transaction binding the contract method 0xbba509b8.
//
// Solidity: function bulkRegister(string[] names, bytes32[] deploymentIDs) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) BulkRegister(opts *bind.TransactOpts, names []string, deploymentIDs [][32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "bulkRegister", names, deploymentIDs)
}

// BulkRegister is a paid mutator transaction binding the contract method 0xbba509b8.
//
// Solidity: function bulkRegister(string[] names, bytes32[] deploymentIDs) returns()
func (_BunkerRegistry *BunkerRegistrySession) BulkRegister(names []string, deploymentIDs [][32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.BulkRegister(&_BunkerRegistry.TransactOpts, names, deploymentIDs)
}

// BulkRegister is a paid mutator transaction binding the contract method 0xbba509b8.
//
// Solidity: function bulkRegister(string[] names, bytes32[] deploymentIDs) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) BulkRegister(names []string, deploymentIDs [][32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.BulkRegister(&_BunkerRegistry.TransactOpts, names, deploymentIDs)
}

// BulkRenew is a paid mutator transaction binding the contract method 0x96202308.
//
// Solidity: function bulkRenew(string[] names) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) BulkRenew(opts *bind.TransactOpts, names []string) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "bulkRenew", names)
}

// BulkRenew is a paid mutator transaction binding the contract method 0x96202308.
//
// Solidity: function bulkRenew(string[] names) returns()
func (_BunkerRegistry *BunkerRegistrySession) BulkRenew(names []string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.BulkRenew(&_BunkerRegistry.TransactOpts, names)
}

// BulkRenew is a paid mutator transaction binding the contract method 0x96202308.
//
// Solidity: function bulkRenew(string[] names) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) BulkRenew(names []string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.BulkRenew(&_BunkerRegistry.TransactOpts, names)
}

// CancelReservation is a paid mutator transaction binding the contract method 0x99e4b8e1.
//
// Solidity: function cancelReservation(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) CancelReservation(opts *bind.TransactOpts, name string) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "cancelReservation", name)
}

// CancelReservation is a paid mutator transaction binding the contract method 0x99e4b8e1.
//
// Solidity: function cancelReservation(string name) returns()
func (_BunkerRegistry *BunkerRegistrySession) CancelReservation(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.CancelReservation(&_BunkerRegistry.TransactOpts, name)
}

// CancelReservation is a paid mutator transaction binding the contract method 0x99e4b8e1.
//
// Solidity: function cancelReservation(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) CancelReservation(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.CancelReservation(&_BunkerRegistry.TransactOpts, name)
}

// ClaimReservation is a paid mutator transaction binding the contract method 0x594ae315.
//
// Solidity: function claimReservation(string name, bytes32 deploymentID) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) ClaimReservation(opts *bind.TransactOpts, name string, deploymentID [32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "claimReservation", name, deploymentID)
}

// ClaimReservation is a paid mutator transaction binding the contract method 0x594ae315.
//
// Solidity: function claimReservation(string name, bytes32 deploymentID) returns()
func (_BunkerRegistry *BunkerRegistrySession) ClaimReservation(name string, deploymentID [32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.ClaimReservation(&_BunkerRegistry.TransactOpts, name, deploymentID)
}

// ClaimReservation is a paid mutator transaction binding the contract method 0x594ae315.
//
// Solidity: function claimReservation(string name, bytes32 deploymentID) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) ClaimReservation(name string, deploymentID [32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.ClaimReservation(&_BunkerRegistry.TransactOpts, name, deploymentID)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerRegistry *BunkerRegistryTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerRegistry *BunkerRegistrySession) Pause() (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Pause(&_BunkerRegistry.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) Pause() (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Pause(&_BunkerRegistry.TransactOpts)
}

// ReclaimSquatted is a paid mutator transaction binding the contract method 0x3197770c.
//
// Solidity: function reclaimSquatted(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) ReclaimSquatted(opts *bind.TransactOpts, name string) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "reclaimSquatted", name)
}

// ReclaimSquatted is a paid mutator transaction binding the contract method 0x3197770c.
//
// Solidity: function reclaimSquatted(string name) returns()
func (_BunkerRegistry *BunkerRegistrySession) ReclaimSquatted(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.ReclaimSquatted(&_BunkerRegistry.TransactOpts, name)
}

// ReclaimSquatted is a paid mutator transaction binding the contract method 0x3197770c.
//
// Solidity: function reclaimSquatted(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) ReclaimSquatted(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.ReclaimSquatted(&_BunkerRegistry.TransactOpts, name)
}

// Register is a paid mutator transaction binding the contract method 0x656afdee.
//
// Solidity: function register(string name, bytes32 deploymentID) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) Register(opts *bind.TransactOpts, name string, deploymentID [32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "register", name, deploymentID)
}

// Register is a paid mutator transaction binding the contract method 0x656afdee.
//
// Solidity: function register(string name, bytes32 deploymentID) returns()
func (_BunkerRegistry *BunkerRegistrySession) Register(name string, deploymentID [32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Register(&_BunkerRegistry.TransactOpts, name, deploymentID)
}

// Register is a paid mutator transaction binding the contract method 0x656afdee.
//
// Solidity: function register(string name, bytes32 deploymentID) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) Register(name string, deploymentID [32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Register(&_BunkerRegistry.TransactOpts, name, deploymentID)
}

// RegisterWithReferral is a paid mutator transaction binding the contract method 0xe7707dba.
//
// Solidity: function registerWithReferral(string name, bytes32 deploymentID, address referrer) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) RegisterWithReferral(opts *bind.TransactOpts, name string, deploymentID [32]byte, referrer common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "registerWithReferral", name, deploymentID, referrer)
}

// RegisterWithReferral is a paid mutator transaction binding the contract method 0xe7707dba.
//
// Solidity: function registerWithReferral(string name, bytes32 deploymentID, address referrer) returns()
func (_BunkerRegistry *BunkerRegistrySession) RegisterWithReferral(name string, deploymentID [32]byte, referrer common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.RegisterWithReferral(&_BunkerRegistry.TransactOpts, name, deploymentID, referrer)
}

// RegisterWithReferral is a paid mutator transaction binding the contract method 0xe7707dba.
//
// Solidity: function registerWithReferral(string name, bytes32 deploymentID, address referrer) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) RegisterWithReferral(name string, deploymentID [32]byte, referrer common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.RegisterWithReferral(&_BunkerRegistry.TransactOpts, name, deploymentID, referrer)
}

// Release is a paid mutator transaction binding the contract method 0xf34e3723.
//
// Solidity: function release(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) Release(opts *bind.TransactOpts, name string) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "release", name)
}

// Release is a paid mutator transaction binding the contract method 0xf34e3723.
//
// Solidity: function release(string name) returns()
func (_BunkerRegistry *BunkerRegistrySession) Release(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Release(&_BunkerRegistry.TransactOpts, name)
}

// Release is a paid mutator transaction binding the contract method 0xf34e3723.
//
// Solidity: function release(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) Release(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Release(&_BunkerRegistry.TransactOpts, name)
}

// Renew is a paid mutator transaction binding the contract method 0xa4a9a612.
//
// Solidity: function renew(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) Renew(opts *bind.TransactOpts, name string) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "renew", name)
}

// Renew is a paid mutator transaction binding the contract method 0xa4a9a612.
//
// Solidity: function renew(string name) returns()
func (_BunkerRegistry *BunkerRegistrySession) Renew(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Renew(&_BunkerRegistry.TransactOpts, name)
}

// Renew is a paid mutator transaction binding the contract method 0xa4a9a612.
//
// Solidity: function renew(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) Renew(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Renew(&_BunkerRegistry.TransactOpts, name)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerRegistry *BunkerRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerRegistry *BunkerRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerRegistry.Contract.RenounceOwnership(&_BunkerRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerRegistry.Contract.RenounceOwnership(&_BunkerRegistry.TransactOpts)
}

// Reserve is a paid mutator transaction binding the contract method 0xae999ece.
//
// Solidity: function reserve(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) Reserve(opts *bind.TransactOpts, name string) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "reserve", name)
}

// Reserve is a paid mutator transaction binding the contract method 0xae999ece.
//
// Solidity: function reserve(string name) returns()
func (_BunkerRegistry *BunkerRegistrySession) Reserve(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Reserve(&_BunkerRegistry.TransactOpts, name)
}

// Reserve is a paid mutator transaction binding the contract method 0xae999ece.
//
// Solidity: function reserve(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) Reserve(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Reserve(&_BunkerRegistry.TransactOpts, name)
}

// SetChangeFee is a paid mutator transaction binding the contract method 0xc108adab.
//
// Solidity: function setChangeFee(uint256 fee) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetChangeFee(opts *bind.TransactOpts, fee *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setChangeFee", fee)
}

// SetChangeFee is a paid mutator transaction binding the contract method 0xc108adab.
//
// Solidity: function setChangeFee(uint256 fee) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetChangeFee(fee *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetChangeFee(&_BunkerRegistry.TransactOpts, fee)
}

// SetChangeFee is a paid mutator transaction binding the contract method 0xc108adab.
//
// Solidity: function setChangeFee(uint256 fee) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetChangeFee(fee *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetChangeFee(&_BunkerRegistry.TransactOpts, fee)
}

// SetExpirationPeriod is a paid mutator transaction binding the contract method 0xf6eea098.
//
// Solidity: function setExpirationPeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetExpirationPeriod(opts *bind.TransactOpts, period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setExpirationPeriod", period)
}

// SetExpirationPeriod is a paid mutator transaction binding the contract method 0xf6eea098.
//
// Solidity: function setExpirationPeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetExpirationPeriod(period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetExpirationPeriod(&_BunkerRegistry.TransactOpts, period)
}

// SetExpirationPeriod is a paid mutator transaction binding the contract method 0xf6eea098.
//
// Solidity: function setExpirationPeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetExpirationPeriod(period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetExpirationPeriod(&_BunkerRegistry.TransactOpts, period)
}

// SetGracePeriod is a paid mutator transaction binding the contract method 0xf2f65960.
//
// Solidity: function setGracePeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetGracePeriod(opts *bind.TransactOpts, period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setGracePeriod", period)
}

// SetGracePeriod is a paid mutator transaction binding the contract method 0xf2f65960.
//
// Solidity: function setGracePeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetGracePeriod(period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetGracePeriod(&_BunkerRegistry.TransactOpts, period)
}

// SetGracePeriod is a paid mutator transaction binding the contract method 0xf2f65960.
//
// Solidity: function setGracePeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetGracePeriod(period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetGracePeriod(&_BunkerRegistry.TransactOpts, period)
}

// SetMetadata is a paid mutator transaction binding the contract method 0x0890d80c.
//
// Solidity: function setMetadata(string name, string description, string avatarURL) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetMetadata(opts *bind.TransactOpts, name string, description string, avatarURL string) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setMetadata", name, description, avatarURL)
}

// SetMetadata is a paid mutator transaction binding the contract method 0x0890d80c.
//
// Solidity: function setMetadata(string name, string description, string avatarURL) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetMetadata(name string, description string, avatarURL string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetMetadata(&_BunkerRegistry.TransactOpts, name, description, avatarURL)
}

// SetMetadata is a paid mutator transaction binding the contract method 0x0890d80c.
//
// Solidity: function setMetadata(string name, string description, string avatarURL) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetMetadata(name string, description string, avatarURL string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetMetadata(&_BunkerRegistry.TransactOpts, name, description, avatarURL)
}

// SetPrimaryName is a paid mutator transaction binding the contract method 0xab66abd5.
//
// Solidity: function setPrimaryName(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetPrimaryName(opts *bind.TransactOpts, name string) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setPrimaryName", name)
}

// SetPrimaryName is a paid mutator transaction binding the contract method 0xab66abd5.
//
// Solidity: function setPrimaryName(string name) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetPrimaryName(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetPrimaryName(&_BunkerRegistry.TransactOpts, name)
}

// SetPrimaryName is a paid mutator transaction binding the contract method 0xab66abd5.
//
// Solidity: function setPrimaryName(string name) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetPrimaryName(name string) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetPrimaryName(&_BunkerRegistry.TransactOpts, name)
}

// SetReferralDiscountBps is a paid mutator transaction binding the contract method 0x98da62d5.
//
// Solidity: function setReferralDiscountBps(uint256 bps) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetReferralDiscountBps(opts *bind.TransactOpts, bps *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setReferralDiscountBps", bps)
}

// SetReferralDiscountBps is a paid mutator transaction binding the contract method 0x98da62d5.
//
// Solidity: function setReferralDiscountBps(uint256 bps) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetReferralDiscountBps(bps *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetReferralDiscountBps(&_BunkerRegistry.TransactOpts, bps)
}

// SetReferralDiscountBps is a paid mutator transaction binding the contract method 0x98da62d5.
//
// Solidity: function setReferralDiscountBps(uint256 bps) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetReferralDiscountBps(bps *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetReferralDiscountBps(&_BunkerRegistry.TransactOpts, bps)
}

// SetReferralRewardBps is a paid mutator transaction binding the contract method 0xb09e603f.
//
// Solidity: function setReferralRewardBps(uint256 bps) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetReferralRewardBps(opts *bind.TransactOpts, bps *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setReferralRewardBps", bps)
}

// SetReferralRewardBps is a paid mutator transaction binding the contract method 0xb09e603f.
//
// Solidity: function setReferralRewardBps(uint256 bps) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetReferralRewardBps(bps *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetReferralRewardBps(&_BunkerRegistry.TransactOpts, bps)
}

// SetReferralRewardBps is a paid mutator transaction binding the contract method 0xb09e603f.
//
// Solidity: function setReferralRewardBps(uint256 bps) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetReferralRewardBps(bps *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetReferralRewardBps(&_BunkerRegistry.TransactOpts, bps)
}

// SetRegistrationFee is a paid mutator transaction binding the contract method 0xc320c727.
//
// Solidity: function setRegistrationFee(uint256 newFee) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetRegistrationFee(opts *bind.TransactOpts, newFee *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setRegistrationFee", newFee)
}

// SetRegistrationFee is a paid mutator transaction binding the contract method 0xc320c727.
//
// Solidity: function setRegistrationFee(uint256 newFee) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetRegistrationFee(newFee *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetRegistrationFee(&_BunkerRegistry.TransactOpts, newFee)
}

// SetRegistrationFee is a paid mutator transaction binding the contract method 0xc320c727.
//
// Solidity: function setRegistrationFee(uint256 newFee) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetRegistrationFee(newFee *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetRegistrationFee(&_BunkerRegistry.TransactOpts, newFee)
}

// SetReservationPeriod is a paid mutator transaction binding the contract method 0xceb20a25.
//
// Solidity: function setReservationPeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetReservationPeriod(opts *bind.TransactOpts, period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setReservationPeriod", period)
}

// SetReservationPeriod is a paid mutator transaction binding the contract method 0xceb20a25.
//
// Solidity: function setReservationPeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetReservationPeriod(period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetReservationPeriod(&_BunkerRegistry.TransactOpts, period)
}

// SetReservationPeriod is a paid mutator transaction binding the contract method 0xceb20a25.
//
// Solidity: function setReservationPeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetReservationPeriod(period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetReservationPeriod(&_BunkerRegistry.TransactOpts, period)
}

// SetReservedName is a paid mutator transaction binding the contract method 0x9248cd43.
//
// Solidity: function setReservedName(string name, bool reserved) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetReservedName(opts *bind.TransactOpts, name string, reserved bool) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setReservedName", name, reserved)
}

// SetReservedName is a paid mutator transaction binding the contract method 0x9248cd43.
//
// Solidity: function setReservedName(string name, bool reserved) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetReservedName(name string, reserved bool) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetReservedName(&_BunkerRegistry.TransactOpts, name, reserved)
}

// SetReservedName is a paid mutator transaction binding the contract method 0x9248cd43.
//
// Solidity: function setReservedName(string name, bool reserved) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetReservedName(name string, reserved bool) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetReservedName(&_BunkerRegistry.TransactOpts, name, reserved)
}

// SetShortNamesEnabled is a paid mutator transaction binding the contract method 0x1482887f.
//
// Solidity: function setShortNamesEnabled(bool enabled) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetShortNamesEnabled(opts *bind.TransactOpts, enabled bool) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setShortNamesEnabled", enabled)
}

// SetShortNamesEnabled is a paid mutator transaction binding the contract method 0x1482887f.
//
// Solidity: function setShortNamesEnabled(bool enabled) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetShortNamesEnabled(enabled bool) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetShortNamesEnabled(&_BunkerRegistry.TransactOpts, enabled)
}

// SetShortNamesEnabled is a paid mutator transaction binding the contract method 0x1482887f.
//
// Solidity: function setShortNamesEnabled(bool enabled) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetShortNamesEnabled(enabled bool) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetShortNamesEnabled(&_BunkerRegistry.TransactOpts, enabled)
}

// SetSquattingGracePeriod is a paid mutator transaction binding the contract method 0x113f1bc4.
//
// Solidity: function setSquattingGracePeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetSquattingGracePeriod(opts *bind.TransactOpts, period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setSquattingGracePeriod", period)
}

// SetSquattingGracePeriod is a paid mutator transaction binding the contract method 0x113f1bc4.
//
// Solidity: function setSquattingGracePeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetSquattingGracePeriod(period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetSquattingGracePeriod(&_BunkerRegistry.TransactOpts, period)
}

// SetSquattingGracePeriod is a paid mutator transaction binding the contract method 0x113f1bc4.
//
// Solidity: function setSquattingGracePeriod(uint256 period) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetSquattingGracePeriod(period *big.Int) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetSquattingGracePeriod(&_BunkerRegistry.TransactOpts, period)
}

// SetStakingContract is a paid mutator transaction binding the contract method 0x9dd373b9.
//
// Solidity: function setStakingContract(address addr) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetStakingContract(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setStakingContract", addr)
}

// SetStakingContract is a paid mutator transaction binding the contract method 0x9dd373b9.
//
// Solidity: function setStakingContract(address addr) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetStakingContract(addr common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetStakingContract(&_BunkerRegistry.TransactOpts, addr)
}

// SetStakingContract is a paid mutator transaction binding the contract method 0x9dd373b9.
//
// Solidity: function setStakingContract(address addr) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetStakingContract(addr common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetStakingContract(&_BunkerRegistry.TransactOpts, addr)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address newTreasury) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) SetTreasury(opts *bind.TransactOpts, newTreasury common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "setTreasury", newTreasury)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address newTreasury) returns()
func (_BunkerRegistry *BunkerRegistrySession) SetTreasury(newTreasury common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetTreasury(&_BunkerRegistry.TransactOpts, newTreasury)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address newTreasury) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) SetTreasury(newTreasury common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.SetTreasury(&_BunkerRegistry.TransactOpts, newTreasury)
}

// Transfer is a paid mutator transaction binding the contract method 0xfbf58b3e.
//
// Solidity: function transfer(string name, address newOwner) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) Transfer(opts *bind.TransactOpts, name string, newOwner common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "transfer", name, newOwner)
}

// Transfer is a paid mutator transaction binding the contract method 0xfbf58b3e.
//
// Solidity: function transfer(string name, address newOwner) returns()
func (_BunkerRegistry *BunkerRegistrySession) Transfer(name string, newOwner common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Transfer(&_BunkerRegistry.TransactOpts, name, newOwner)
}

// Transfer is a paid mutator transaction binding the contract method 0xfbf58b3e.
//
// Solidity: function transfer(string name, address newOwner) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) Transfer(name string, newOwner common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Transfer(&_BunkerRegistry.TransactOpts, name, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerRegistry *BunkerRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.TransferOwnership(&_BunkerRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.TransferOwnership(&_BunkerRegistry.TransactOpts, newOwner)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerRegistry *BunkerRegistryTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerRegistry *BunkerRegistrySession) Unpause() (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Unpause(&_BunkerRegistry.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) Unpause() (*types.Transaction, error) {
	return _BunkerRegistry.Contract.Unpause(&_BunkerRegistry.TransactOpts)
}

// UpdateDeployment is a paid mutator transaction binding the contract method 0xd572671f.
//
// Solidity: function updateDeployment(string name, bytes32 newDeploymentID) returns()
func (_BunkerRegistry *BunkerRegistryTransactor) UpdateDeployment(opts *bind.TransactOpts, name string, newDeploymentID [32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.contract.Transact(opts, "updateDeployment", name, newDeploymentID)
}

// UpdateDeployment is a paid mutator transaction binding the contract method 0xd572671f.
//
// Solidity: function updateDeployment(string name, bytes32 newDeploymentID) returns()
func (_BunkerRegistry *BunkerRegistrySession) UpdateDeployment(name string, newDeploymentID [32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.UpdateDeployment(&_BunkerRegistry.TransactOpts, name, newDeploymentID)
}

// UpdateDeployment is a paid mutator transaction binding the contract method 0xd572671f.
//
// Solidity: function updateDeployment(string name, bytes32 newDeploymentID) returns()
func (_BunkerRegistry *BunkerRegistryTransactorSession) UpdateDeployment(name string, newDeploymentID [32]byte) (*types.Transaction, error) {
	return _BunkerRegistry.Contract.UpdateDeployment(&_BunkerRegistry.TransactOpts, name, newDeploymentID)
}

// BunkerRegistryChangeFeeUpdatedIterator is returned from FilterChangeFeeUpdated and is used to iterate over the raw logs and unpacked data for ChangeFeeUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryChangeFeeUpdatedIterator struct {
	Event *BunkerRegistryChangeFeeUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryChangeFeeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryChangeFeeUpdated)
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
		it.Event = new(BunkerRegistryChangeFeeUpdated)
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
func (it *BunkerRegistryChangeFeeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryChangeFeeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryChangeFeeUpdated represents a ChangeFeeUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryChangeFeeUpdated struct {
	OldFee *big.Int
	NewFee *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterChangeFeeUpdated is a free log retrieval operation binding the contract event 0x55116a86ab3cee7c9c9b730bcda216ebe2ca2660136b7b36ca7138603a2bd698.
//
// Solidity: event ChangeFeeUpdated(uint256 oldFee, uint256 newFee)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterChangeFeeUpdated(opts *bind.FilterOpts) (*BunkerRegistryChangeFeeUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "ChangeFeeUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryChangeFeeUpdatedIterator{contract: _BunkerRegistry.contract, event: "ChangeFeeUpdated", logs: logs, sub: sub}, nil
}

// WatchChangeFeeUpdated is a free log subscription operation binding the contract event 0x55116a86ab3cee7c9c9b730bcda216ebe2ca2660136b7b36ca7138603a2bd698.
//
// Solidity: event ChangeFeeUpdated(uint256 oldFee, uint256 newFee)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchChangeFeeUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryChangeFeeUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "ChangeFeeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryChangeFeeUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "ChangeFeeUpdated", log); err != nil {
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

// ParseChangeFeeUpdated is a log parse operation binding the contract event 0x55116a86ab3cee7c9c9b730bcda216ebe2ca2660136b7b36ca7138603a2bd698.
//
// Solidity: event ChangeFeeUpdated(uint256 oldFee, uint256 newFee)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseChangeFeeUpdated(log types.Log) (*BunkerRegistryChangeFeeUpdated, error) {
	event := new(BunkerRegistryChangeFeeUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "ChangeFeeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryExpirationPeriodUpdatedIterator is returned from FilterExpirationPeriodUpdated and is used to iterate over the raw logs and unpacked data for ExpirationPeriodUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryExpirationPeriodUpdatedIterator struct {
	Event *BunkerRegistryExpirationPeriodUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryExpirationPeriodUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryExpirationPeriodUpdated)
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
		it.Event = new(BunkerRegistryExpirationPeriodUpdated)
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
func (it *BunkerRegistryExpirationPeriodUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryExpirationPeriodUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryExpirationPeriodUpdated represents a ExpirationPeriodUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryExpirationPeriodUpdated struct {
	OldPeriod *big.Int
	NewPeriod *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterExpirationPeriodUpdated is a free log retrieval operation binding the contract event 0x4a8ef160cb5b411fb018f9d68556ec33f8e7901166d29bc96fc5ca7d5eac0980.
//
// Solidity: event ExpirationPeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterExpirationPeriodUpdated(opts *bind.FilterOpts) (*BunkerRegistryExpirationPeriodUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "ExpirationPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryExpirationPeriodUpdatedIterator{contract: _BunkerRegistry.contract, event: "ExpirationPeriodUpdated", logs: logs, sub: sub}, nil
}

// WatchExpirationPeriodUpdated is a free log subscription operation binding the contract event 0x4a8ef160cb5b411fb018f9d68556ec33f8e7901166d29bc96fc5ca7d5eac0980.
//
// Solidity: event ExpirationPeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchExpirationPeriodUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryExpirationPeriodUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "ExpirationPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryExpirationPeriodUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "ExpirationPeriodUpdated", log); err != nil {
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

// ParseExpirationPeriodUpdated is a log parse operation binding the contract event 0x4a8ef160cb5b411fb018f9d68556ec33f8e7901166d29bc96fc5ca7d5eac0980.
//
// Solidity: event ExpirationPeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseExpirationPeriodUpdated(log types.Log) (*BunkerRegistryExpirationPeriodUpdated, error) {
	event := new(BunkerRegistryExpirationPeriodUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "ExpirationPeriodUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryGracePeriodUpdatedIterator is returned from FilterGracePeriodUpdated and is used to iterate over the raw logs and unpacked data for GracePeriodUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryGracePeriodUpdatedIterator struct {
	Event *BunkerRegistryGracePeriodUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryGracePeriodUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryGracePeriodUpdated)
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
		it.Event = new(BunkerRegistryGracePeriodUpdated)
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
func (it *BunkerRegistryGracePeriodUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryGracePeriodUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryGracePeriodUpdated represents a GracePeriodUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryGracePeriodUpdated struct {
	OldPeriod *big.Int
	NewPeriod *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterGracePeriodUpdated is a free log retrieval operation binding the contract event 0x55c7a79c45e9a972909cd640f9336a14a84adbaf756211f16267001854110191.
//
// Solidity: event GracePeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterGracePeriodUpdated(opts *bind.FilterOpts) (*BunkerRegistryGracePeriodUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "GracePeriodUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryGracePeriodUpdatedIterator{contract: _BunkerRegistry.contract, event: "GracePeriodUpdated", logs: logs, sub: sub}, nil
}

// WatchGracePeriodUpdated is a free log subscription operation binding the contract event 0x55c7a79c45e9a972909cd640f9336a14a84adbaf756211f16267001854110191.
//
// Solidity: event GracePeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchGracePeriodUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryGracePeriodUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "GracePeriodUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryGracePeriodUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "GracePeriodUpdated", log); err != nil {
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

// ParseGracePeriodUpdated is a log parse operation binding the contract event 0x55c7a79c45e9a972909cd640f9336a14a84adbaf756211f16267001854110191.
//
// Solidity: event GracePeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseGracePeriodUpdated(log types.Log) (*BunkerRegistryGracePeriodUpdated, error) {
	event := new(BunkerRegistryGracePeriodUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "GracePeriodUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryMetadataUpdatedIterator is returned from FilterMetadataUpdated and is used to iterate over the raw logs and unpacked data for MetadataUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryMetadataUpdatedIterator struct {
	Event *BunkerRegistryMetadataUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryMetadataUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryMetadataUpdated)
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
		it.Event = new(BunkerRegistryMetadataUpdated)
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
func (it *BunkerRegistryMetadataUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryMetadataUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryMetadataUpdated represents a MetadataUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryMetadataUpdated struct {
	NameIndexed common.Hash
	Name        string
	Owner       common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterMetadataUpdated is a free log retrieval operation binding the contract event 0x216811aefb54a27978bda071ef6a3fac1ab61bbf27fe10ade3a8fe021f445dd7.
//
// Solidity: event MetadataUpdated(string indexed nameIndexed, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterMetadataUpdated(opts *bind.FilterOpts, nameIndexed []string, owner []common.Address) (*BunkerRegistryMetadataUpdatedIterator, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "MetadataUpdated", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryMetadataUpdatedIterator{contract: _BunkerRegistry.contract, event: "MetadataUpdated", logs: logs, sub: sub}, nil
}

// WatchMetadataUpdated is a free log subscription operation binding the contract event 0x216811aefb54a27978bda071ef6a3fac1ab61bbf27fe10ade3a8fe021f445dd7.
//
// Solidity: event MetadataUpdated(string indexed nameIndexed, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchMetadataUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryMetadataUpdated, nameIndexed []string, owner []common.Address) (event.Subscription, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "MetadataUpdated", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryMetadataUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "MetadataUpdated", log); err != nil {
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

// ParseMetadataUpdated is a log parse operation binding the contract event 0x216811aefb54a27978bda071ef6a3fac1ab61bbf27fe10ade3a8fe021f445dd7.
//
// Solidity: event MetadataUpdated(string indexed nameIndexed, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseMetadataUpdated(log types.Log) (*BunkerRegistryMetadataUpdated, error) {
	event := new(BunkerRegistryMetadataUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "MetadataUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the BunkerRegistry contract.
type BunkerRegistryOwnershipTransferStartedIterator struct {
	Event *BunkerRegistryOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryOwnershipTransferStarted)
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
		it.Event = new(BunkerRegistryOwnershipTransferStarted)
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
func (it *BunkerRegistryOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the BunkerRegistry contract.
type BunkerRegistryOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerRegistryOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryOwnershipTransferStartedIterator{contract: _BunkerRegistry.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *BunkerRegistryOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryOwnershipTransferStarted)
				if err := _BunkerRegistry.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_BunkerRegistry *BunkerRegistryFilterer) ParseOwnershipTransferStarted(log types.Log) (*BunkerRegistryOwnershipTransferStarted, error) {
	event := new(BunkerRegistryOwnershipTransferStarted)
	if err := _BunkerRegistry.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BunkerRegistry contract.
type BunkerRegistryOwnershipTransferredIterator struct {
	Event *BunkerRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryOwnershipTransferred)
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
		it.Event = new(BunkerRegistryOwnershipTransferred)
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
func (it *BunkerRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the BunkerRegistry contract.
type BunkerRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryOwnershipTransferredIterator{contract: _BunkerRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BunkerRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryOwnershipTransferred)
				if err := _BunkerRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BunkerRegistry *BunkerRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*BunkerRegistryOwnershipTransferred, error) {
	event := new(BunkerRegistryOwnershipTransferred)
	if err := _BunkerRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the BunkerRegistry contract.
type BunkerRegistryPausedIterator struct {
	Event *BunkerRegistryPaused // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryPaused)
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
		it.Event = new(BunkerRegistryPaused)
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
func (it *BunkerRegistryPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryPaused represents a Paused event raised by the BunkerRegistry contract.
type BunkerRegistryPaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterPaused(opts *bind.FilterOpts) (*BunkerRegistryPausedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryPausedIterator{contract: _BunkerRegistry.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *BunkerRegistryPaused) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryPaused)
				if err := _BunkerRegistry.contract.UnpackLog(event, "Paused", log); err != nil {
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
func (_BunkerRegistry *BunkerRegistryFilterer) ParsePaused(log types.Log) (*BunkerRegistryPaused, error) {
	event := new(BunkerRegistryPaused)
	if err := _BunkerRegistry.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryPrimaryNameSetIterator is returned from FilterPrimaryNameSet and is used to iterate over the raw logs and unpacked data for PrimaryNameSet events raised by the BunkerRegistry contract.
type BunkerRegistryPrimaryNameSetIterator struct {
	Event *BunkerRegistryPrimaryNameSet // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryPrimaryNameSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryPrimaryNameSet)
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
		it.Event = new(BunkerRegistryPrimaryNameSet)
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
func (it *BunkerRegistryPrimaryNameSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryPrimaryNameSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryPrimaryNameSet represents a PrimaryNameSet event raised by the BunkerRegistry contract.
type BunkerRegistryPrimaryNameSet struct {
	DeploymentID [32]byte
	Name         string
	Owner        common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterPrimaryNameSet is a free log retrieval operation binding the contract event 0x7a8ce68d980cbfc1310c4a07d785ecdb27cb0fbcbd8eaf27d76f7dd7d1b0d7f9.
//
// Solidity: event PrimaryNameSet(bytes32 indexed deploymentID, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterPrimaryNameSet(opts *bind.FilterOpts, deploymentID [][32]byte, owner []common.Address) (*BunkerRegistryPrimaryNameSetIterator, error) {

	var deploymentIDRule []interface{}
	for _, deploymentIDItem := range deploymentID {
		deploymentIDRule = append(deploymentIDRule, deploymentIDItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "PrimaryNameSet", deploymentIDRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryPrimaryNameSetIterator{contract: _BunkerRegistry.contract, event: "PrimaryNameSet", logs: logs, sub: sub}, nil
}

// WatchPrimaryNameSet is a free log subscription operation binding the contract event 0x7a8ce68d980cbfc1310c4a07d785ecdb27cb0fbcbd8eaf27d76f7dd7d1b0d7f9.
//
// Solidity: event PrimaryNameSet(bytes32 indexed deploymentID, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchPrimaryNameSet(opts *bind.WatchOpts, sink chan<- *BunkerRegistryPrimaryNameSet, deploymentID [][32]byte, owner []common.Address) (event.Subscription, error) {

	var deploymentIDRule []interface{}
	for _, deploymentIDItem := range deploymentID {
		deploymentIDRule = append(deploymentIDRule, deploymentIDItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "PrimaryNameSet", deploymentIDRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryPrimaryNameSet)
				if err := _BunkerRegistry.contract.UnpackLog(event, "PrimaryNameSet", log); err != nil {
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

// ParsePrimaryNameSet is a log parse operation binding the contract event 0x7a8ce68d980cbfc1310c4a07d785ecdb27cb0fbcbd8eaf27d76f7dd7d1b0d7f9.
//
// Solidity: event PrimaryNameSet(bytes32 indexed deploymentID, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) ParsePrimaryNameSet(log types.Log) (*BunkerRegistryPrimaryNameSet, error) {
	event := new(BunkerRegistryPrimaryNameSet)
	if err := _BunkerRegistry.contract.UnpackLog(event, "PrimaryNameSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryReferralDiscountUpdatedIterator is returned from FilterReferralDiscountUpdated and is used to iterate over the raw logs and unpacked data for ReferralDiscountUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryReferralDiscountUpdatedIterator struct {
	Event *BunkerRegistryReferralDiscountUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryReferralDiscountUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryReferralDiscountUpdated)
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
		it.Event = new(BunkerRegistryReferralDiscountUpdated)
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
func (it *BunkerRegistryReferralDiscountUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryReferralDiscountUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryReferralDiscountUpdated represents a ReferralDiscountUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryReferralDiscountUpdated struct {
	OldBps *big.Int
	NewBps *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterReferralDiscountUpdated is a free log retrieval operation binding the contract event 0x062059aa293a11d5474590143d5f595371302263fea1cd1bb4627d7c9d242fd0.
//
// Solidity: event ReferralDiscountUpdated(uint256 oldBps, uint256 newBps)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterReferralDiscountUpdated(opts *bind.FilterOpts) (*BunkerRegistryReferralDiscountUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "ReferralDiscountUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryReferralDiscountUpdatedIterator{contract: _BunkerRegistry.contract, event: "ReferralDiscountUpdated", logs: logs, sub: sub}, nil
}

// WatchReferralDiscountUpdated is a free log subscription operation binding the contract event 0x062059aa293a11d5474590143d5f595371302263fea1cd1bb4627d7c9d242fd0.
//
// Solidity: event ReferralDiscountUpdated(uint256 oldBps, uint256 newBps)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchReferralDiscountUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryReferralDiscountUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "ReferralDiscountUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryReferralDiscountUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "ReferralDiscountUpdated", log); err != nil {
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

// ParseReferralDiscountUpdated is a log parse operation binding the contract event 0x062059aa293a11d5474590143d5f595371302263fea1cd1bb4627d7c9d242fd0.
//
// Solidity: event ReferralDiscountUpdated(uint256 oldBps, uint256 newBps)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseReferralDiscountUpdated(log types.Log) (*BunkerRegistryReferralDiscountUpdated, error) {
	event := new(BunkerRegistryReferralDiscountUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "ReferralDiscountUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryReferralRewardUpdatedIterator is returned from FilterReferralRewardUpdated and is used to iterate over the raw logs and unpacked data for ReferralRewardUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryReferralRewardUpdatedIterator struct {
	Event *BunkerRegistryReferralRewardUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryReferralRewardUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryReferralRewardUpdated)
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
		it.Event = new(BunkerRegistryReferralRewardUpdated)
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
func (it *BunkerRegistryReferralRewardUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryReferralRewardUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryReferralRewardUpdated represents a ReferralRewardUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryReferralRewardUpdated struct {
	OldBps *big.Int
	NewBps *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterReferralRewardUpdated is a free log retrieval operation binding the contract event 0x0ba2611e03c210377c7111f9c3f59febd751b32e081daa8d153f9e5aefc4d44a.
//
// Solidity: event ReferralRewardUpdated(uint256 oldBps, uint256 newBps)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterReferralRewardUpdated(opts *bind.FilterOpts) (*BunkerRegistryReferralRewardUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "ReferralRewardUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryReferralRewardUpdatedIterator{contract: _BunkerRegistry.contract, event: "ReferralRewardUpdated", logs: logs, sub: sub}, nil
}

// WatchReferralRewardUpdated is a free log subscription operation binding the contract event 0x0ba2611e03c210377c7111f9c3f59febd751b32e081daa8d153f9e5aefc4d44a.
//
// Solidity: event ReferralRewardUpdated(uint256 oldBps, uint256 newBps)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchReferralRewardUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryReferralRewardUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "ReferralRewardUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryReferralRewardUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "ReferralRewardUpdated", log); err != nil {
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

// ParseReferralRewardUpdated is a log parse operation binding the contract event 0x0ba2611e03c210377c7111f9c3f59febd751b32e081daa8d153f9e5aefc4d44a.
//
// Solidity: event ReferralRewardUpdated(uint256 oldBps, uint256 newBps)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseReferralRewardUpdated(log types.Log) (*BunkerRegistryReferralRewardUpdated, error) {
	event := new(BunkerRegistryReferralRewardUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "ReferralRewardUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryRegistrationFeeUpdatedIterator is returned from FilterRegistrationFeeUpdated and is used to iterate over the raw logs and unpacked data for RegistrationFeeUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryRegistrationFeeUpdatedIterator struct {
	Event *BunkerRegistryRegistrationFeeUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryRegistrationFeeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryRegistrationFeeUpdated)
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
		it.Event = new(BunkerRegistryRegistrationFeeUpdated)
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
func (it *BunkerRegistryRegistrationFeeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryRegistrationFeeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryRegistrationFeeUpdated represents a RegistrationFeeUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryRegistrationFeeUpdated struct {
	OldFee *big.Int
	NewFee *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRegistrationFeeUpdated is a free log retrieval operation binding the contract event 0x50b218c5a101ad05d53ab0a964d01da639ee79525ae4b7802ed714249740a8d5.
//
// Solidity: event RegistrationFeeUpdated(uint256 oldFee, uint256 newFee)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterRegistrationFeeUpdated(opts *bind.FilterOpts) (*BunkerRegistryRegistrationFeeUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "RegistrationFeeUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryRegistrationFeeUpdatedIterator{contract: _BunkerRegistry.contract, event: "RegistrationFeeUpdated", logs: logs, sub: sub}, nil
}

// WatchRegistrationFeeUpdated is a free log subscription operation binding the contract event 0x50b218c5a101ad05d53ab0a964d01da639ee79525ae4b7802ed714249740a8d5.
//
// Solidity: event RegistrationFeeUpdated(uint256 oldFee, uint256 newFee)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchRegistrationFeeUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryRegistrationFeeUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "RegistrationFeeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryRegistrationFeeUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "RegistrationFeeUpdated", log); err != nil {
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

// ParseRegistrationFeeUpdated is a log parse operation binding the contract event 0x50b218c5a101ad05d53ab0a964d01da639ee79525ae4b7802ed714249740a8d5.
//
// Solidity: event RegistrationFeeUpdated(uint256 oldFee, uint256 newFee)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseRegistrationFeeUpdated(log types.Log) (*BunkerRegistryRegistrationFeeUpdated, error) {
	event := new(BunkerRegistryRegistrationFeeUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "RegistrationFeeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryReservationCancelledIterator is returned from FilterReservationCancelled and is used to iterate over the raw logs and unpacked data for ReservationCancelled events raised by the BunkerRegistry contract.
type BunkerRegistryReservationCancelledIterator struct {
	Event *BunkerRegistryReservationCancelled // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryReservationCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryReservationCancelled)
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
		it.Event = new(BunkerRegistryReservationCancelled)
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
func (it *BunkerRegistryReservationCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryReservationCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryReservationCancelled represents a ReservationCancelled event raised by the BunkerRegistry contract.
type BunkerRegistryReservationCancelled struct {
	NameIndexed common.Hash
	Name        string
	Owner       common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterReservationCancelled is a free log retrieval operation binding the contract event 0xd7ecf76eb8f3c47f70a74aeac050fdf2d18b948a777c8aa2835bc54ddeda231e.
//
// Solidity: event ReservationCancelled(string indexed nameIndexed, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterReservationCancelled(opts *bind.FilterOpts, nameIndexed []string, owner []common.Address) (*BunkerRegistryReservationCancelledIterator, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "ReservationCancelled", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryReservationCancelledIterator{contract: _BunkerRegistry.contract, event: "ReservationCancelled", logs: logs, sub: sub}, nil
}

// WatchReservationCancelled is a free log subscription operation binding the contract event 0xd7ecf76eb8f3c47f70a74aeac050fdf2d18b948a777c8aa2835bc54ddeda231e.
//
// Solidity: event ReservationCancelled(string indexed nameIndexed, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchReservationCancelled(opts *bind.WatchOpts, sink chan<- *BunkerRegistryReservationCancelled, nameIndexed []string, owner []common.Address) (event.Subscription, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "ReservationCancelled", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryReservationCancelled)
				if err := _BunkerRegistry.contract.UnpackLog(event, "ReservationCancelled", log); err != nil {
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

// ParseReservationCancelled is a log parse operation binding the contract event 0xd7ecf76eb8f3c47f70a74aeac050fdf2d18b948a777c8aa2835bc54ddeda231e.
//
// Solidity: event ReservationCancelled(string indexed nameIndexed, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseReservationCancelled(log types.Log) (*BunkerRegistryReservationCancelled, error) {
	event := new(BunkerRegistryReservationCancelled)
	if err := _BunkerRegistry.contract.UnpackLog(event, "ReservationCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryReservationClaimedIterator is returned from FilterReservationClaimed and is used to iterate over the raw logs and unpacked data for ReservationClaimed events raised by the BunkerRegistry contract.
type BunkerRegistryReservationClaimedIterator struct {
	Event *BunkerRegistryReservationClaimed // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryReservationClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryReservationClaimed)
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
		it.Event = new(BunkerRegistryReservationClaimed)
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
func (it *BunkerRegistryReservationClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryReservationClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryReservationClaimed represents a ReservationClaimed event raised by the BunkerRegistry contract.
type BunkerRegistryReservationClaimed struct {
	NameIndexed  common.Hash
	Name         string
	Owner        common.Address
	DeploymentID [32]byte
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterReservationClaimed is a free log retrieval operation binding the contract event 0xaf9cf492d92f51be75e0a41717d2eca55872e9c9640d946ea9276c417bafd0d3.
//
// Solidity: event ReservationClaimed(string indexed nameIndexed, string name, address indexed owner, bytes32 deploymentID)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterReservationClaimed(opts *bind.FilterOpts, nameIndexed []string, owner []common.Address) (*BunkerRegistryReservationClaimedIterator, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "ReservationClaimed", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryReservationClaimedIterator{contract: _BunkerRegistry.contract, event: "ReservationClaimed", logs: logs, sub: sub}, nil
}

// WatchReservationClaimed is a free log subscription operation binding the contract event 0xaf9cf492d92f51be75e0a41717d2eca55872e9c9640d946ea9276c417bafd0d3.
//
// Solidity: event ReservationClaimed(string indexed nameIndexed, string name, address indexed owner, bytes32 deploymentID)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchReservationClaimed(opts *bind.WatchOpts, sink chan<- *BunkerRegistryReservationClaimed, nameIndexed []string, owner []common.Address) (event.Subscription, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "ReservationClaimed", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryReservationClaimed)
				if err := _BunkerRegistry.contract.UnpackLog(event, "ReservationClaimed", log); err != nil {
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

// ParseReservationClaimed is a log parse operation binding the contract event 0xaf9cf492d92f51be75e0a41717d2eca55872e9c9640d946ea9276c417bafd0d3.
//
// Solidity: event ReservationClaimed(string indexed nameIndexed, string name, address indexed owner, bytes32 deploymentID)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseReservationClaimed(log types.Log) (*BunkerRegistryReservationClaimed, error) {
	event := new(BunkerRegistryReservationClaimed)
	if err := _BunkerRegistry.contract.UnpackLog(event, "ReservationClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryReservationPeriodUpdatedIterator is returned from FilterReservationPeriodUpdated and is used to iterate over the raw logs and unpacked data for ReservationPeriodUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryReservationPeriodUpdatedIterator struct {
	Event *BunkerRegistryReservationPeriodUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryReservationPeriodUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryReservationPeriodUpdated)
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
		it.Event = new(BunkerRegistryReservationPeriodUpdated)
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
func (it *BunkerRegistryReservationPeriodUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryReservationPeriodUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryReservationPeriodUpdated represents a ReservationPeriodUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryReservationPeriodUpdated struct {
	OldPeriod *big.Int
	NewPeriod *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterReservationPeriodUpdated is a free log retrieval operation binding the contract event 0x1f5290971d7202a921176342fab127484a469b06550ba25e19bc5baa00b67332.
//
// Solidity: event ReservationPeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterReservationPeriodUpdated(opts *bind.FilterOpts) (*BunkerRegistryReservationPeriodUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "ReservationPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryReservationPeriodUpdatedIterator{contract: _BunkerRegistry.contract, event: "ReservationPeriodUpdated", logs: logs, sub: sub}, nil
}

// WatchReservationPeriodUpdated is a free log subscription operation binding the contract event 0x1f5290971d7202a921176342fab127484a469b06550ba25e19bc5baa00b67332.
//
// Solidity: event ReservationPeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchReservationPeriodUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryReservationPeriodUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "ReservationPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryReservationPeriodUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "ReservationPeriodUpdated", log); err != nil {
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

// ParseReservationPeriodUpdated is a log parse operation binding the contract event 0x1f5290971d7202a921176342fab127484a469b06550ba25e19bc5baa00b67332.
//
// Solidity: event ReservationPeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseReservationPeriodUpdated(log types.Log) (*BunkerRegistryReservationPeriodUpdated, error) {
	event := new(BunkerRegistryReservationPeriodUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "ReservationPeriodUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryReservedNameUpdatedIterator is returned from FilterReservedNameUpdated and is used to iterate over the raw logs and unpacked data for ReservedNameUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryReservedNameUpdatedIterator struct {
	Event *BunkerRegistryReservedNameUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryReservedNameUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryReservedNameUpdated)
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
		it.Event = new(BunkerRegistryReservedNameUpdated)
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
func (it *BunkerRegistryReservedNameUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryReservedNameUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryReservedNameUpdated represents a ReservedNameUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryReservedNameUpdated struct {
	Name     string
	Reserved bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterReservedNameUpdated is a free log retrieval operation binding the contract event 0x007b3f58008b6ebd21be6f46bb687e22a47a1e623603588c7889c4f23681d55e.
//
// Solidity: event ReservedNameUpdated(string name, bool reserved)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterReservedNameUpdated(opts *bind.FilterOpts) (*BunkerRegistryReservedNameUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "ReservedNameUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryReservedNameUpdatedIterator{contract: _BunkerRegistry.contract, event: "ReservedNameUpdated", logs: logs, sub: sub}, nil
}

// WatchReservedNameUpdated is a free log subscription operation binding the contract event 0x007b3f58008b6ebd21be6f46bb687e22a47a1e623603588c7889c4f23681d55e.
//
// Solidity: event ReservedNameUpdated(string name, bool reserved)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchReservedNameUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryReservedNameUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "ReservedNameUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryReservedNameUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "ReservedNameUpdated", log); err != nil {
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

// ParseReservedNameUpdated is a log parse operation binding the contract event 0x007b3f58008b6ebd21be6f46bb687e22a47a1e623603588c7889c4f23681d55e.
//
// Solidity: event ReservedNameUpdated(string name, bool reserved)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseReservedNameUpdated(log types.Log) (*BunkerRegistryReservedNameUpdated, error) {
	event := new(BunkerRegistryReservedNameUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "ReservedNameUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryShortNamesEnabledUpdatedIterator is returned from FilterShortNamesEnabledUpdated and is used to iterate over the raw logs and unpacked data for ShortNamesEnabledUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryShortNamesEnabledUpdatedIterator struct {
	Event *BunkerRegistryShortNamesEnabledUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryShortNamesEnabledUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryShortNamesEnabledUpdated)
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
		it.Event = new(BunkerRegistryShortNamesEnabledUpdated)
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
func (it *BunkerRegistryShortNamesEnabledUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryShortNamesEnabledUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryShortNamesEnabledUpdated represents a ShortNamesEnabledUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryShortNamesEnabledUpdated struct {
	Enabled bool
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterShortNamesEnabledUpdated is a free log retrieval operation binding the contract event 0xb5bd277f5de8e3ab27b8ec677e165e9b14868fb26bc569da39e21b3ed5c69ccd.
//
// Solidity: event ShortNamesEnabledUpdated(bool enabled)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterShortNamesEnabledUpdated(opts *bind.FilterOpts) (*BunkerRegistryShortNamesEnabledUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "ShortNamesEnabledUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryShortNamesEnabledUpdatedIterator{contract: _BunkerRegistry.contract, event: "ShortNamesEnabledUpdated", logs: logs, sub: sub}, nil
}

// WatchShortNamesEnabledUpdated is a free log subscription operation binding the contract event 0xb5bd277f5de8e3ab27b8ec677e165e9b14868fb26bc569da39e21b3ed5c69ccd.
//
// Solidity: event ShortNamesEnabledUpdated(bool enabled)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchShortNamesEnabledUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryShortNamesEnabledUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "ShortNamesEnabledUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryShortNamesEnabledUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "ShortNamesEnabledUpdated", log); err != nil {
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

// ParseShortNamesEnabledUpdated is a log parse operation binding the contract event 0xb5bd277f5de8e3ab27b8ec677e165e9b14868fb26bc569da39e21b3ed5c69ccd.
//
// Solidity: event ShortNamesEnabledUpdated(bool enabled)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseShortNamesEnabledUpdated(log types.Log) (*BunkerRegistryShortNamesEnabledUpdated, error) {
	event := new(BunkerRegistryShortNamesEnabledUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "ShortNamesEnabledUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistrySquattedNameReclaimedIterator is returned from FilterSquattedNameReclaimed and is used to iterate over the raw logs and unpacked data for SquattedNameReclaimed events raised by the BunkerRegistry contract.
type BunkerRegistrySquattedNameReclaimedIterator struct {
	Event *BunkerRegistrySquattedNameReclaimed // Event containing the contract specifics and raw log

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
func (it *BunkerRegistrySquattedNameReclaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistrySquattedNameReclaimed)
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
		it.Event = new(BunkerRegistrySquattedNameReclaimed)
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
func (it *BunkerRegistrySquattedNameReclaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistrySquattedNameReclaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistrySquattedNameReclaimed represents a SquattedNameReclaimed event raised by the BunkerRegistry contract.
type BunkerRegistrySquattedNameReclaimed struct {
	NameIndexed common.Hash
	Name        string
	Reclaimer   common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSquattedNameReclaimed is a free log retrieval operation binding the contract event 0x5e9814c56725f590f0878c0872db7875aed61bb268867d274e4c59dae11e3923.
//
// Solidity: event SquattedNameReclaimed(string indexed nameIndexed, string name, address indexed reclaimer)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterSquattedNameReclaimed(opts *bind.FilterOpts, nameIndexed []string, reclaimer []common.Address) (*BunkerRegistrySquattedNameReclaimedIterator, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var reclaimerRule []interface{}
	for _, reclaimerItem := range reclaimer {
		reclaimerRule = append(reclaimerRule, reclaimerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "SquattedNameReclaimed", nameIndexedRule, reclaimerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistrySquattedNameReclaimedIterator{contract: _BunkerRegistry.contract, event: "SquattedNameReclaimed", logs: logs, sub: sub}, nil
}

// WatchSquattedNameReclaimed is a free log subscription operation binding the contract event 0x5e9814c56725f590f0878c0872db7875aed61bb268867d274e4c59dae11e3923.
//
// Solidity: event SquattedNameReclaimed(string indexed nameIndexed, string name, address indexed reclaimer)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchSquattedNameReclaimed(opts *bind.WatchOpts, sink chan<- *BunkerRegistrySquattedNameReclaimed, nameIndexed []string, reclaimer []common.Address) (event.Subscription, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var reclaimerRule []interface{}
	for _, reclaimerItem := range reclaimer {
		reclaimerRule = append(reclaimerRule, reclaimerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "SquattedNameReclaimed", nameIndexedRule, reclaimerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistrySquattedNameReclaimed)
				if err := _BunkerRegistry.contract.UnpackLog(event, "SquattedNameReclaimed", log); err != nil {
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

// ParseSquattedNameReclaimed is a log parse operation binding the contract event 0x5e9814c56725f590f0878c0872db7875aed61bb268867d274e4c59dae11e3923.
//
// Solidity: event SquattedNameReclaimed(string indexed nameIndexed, string name, address indexed reclaimer)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseSquattedNameReclaimed(log types.Log) (*BunkerRegistrySquattedNameReclaimed, error) {
	event := new(BunkerRegistrySquattedNameReclaimed)
	if err := _BunkerRegistry.contract.UnpackLog(event, "SquattedNameReclaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistrySquattingGracePeriodUpdatedIterator is returned from FilterSquattingGracePeriodUpdated and is used to iterate over the raw logs and unpacked data for SquattingGracePeriodUpdated events raised by the BunkerRegistry contract.
type BunkerRegistrySquattingGracePeriodUpdatedIterator struct {
	Event *BunkerRegistrySquattingGracePeriodUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistrySquattingGracePeriodUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistrySquattingGracePeriodUpdated)
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
		it.Event = new(BunkerRegistrySquattingGracePeriodUpdated)
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
func (it *BunkerRegistrySquattingGracePeriodUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistrySquattingGracePeriodUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistrySquattingGracePeriodUpdated represents a SquattingGracePeriodUpdated event raised by the BunkerRegistry contract.
type BunkerRegistrySquattingGracePeriodUpdated struct {
	OldPeriod *big.Int
	NewPeriod *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSquattingGracePeriodUpdated is a free log retrieval operation binding the contract event 0x3e452c0fd86ddf6f5ade3ddcc4e4ff19a6daf5e5e0798a19ad36b85299c7c0db.
//
// Solidity: event SquattingGracePeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterSquattingGracePeriodUpdated(opts *bind.FilterOpts) (*BunkerRegistrySquattingGracePeriodUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "SquattingGracePeriodUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistrySquattingGracePeriodUpdatedIterator{contract: _BunkerRegistry.contract, event: "SquattingGracePeriodUpdated", logs: logs, sub: sub}, nil
}

// WatchSquattingGracePeriodUpdated is a free log subscription operation binding the contract event 0x3e452c0fd86ddf6f5ade3ddcc4e4ff19a6daf5e5e0798a19ad36b85299c7c0db.
//
// Solidity: event SquattingGracePeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchSquattingGracePeriodUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistrySquattingGracePeriodUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "SquattingGracePeriodUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistrySquattingGracePeriodUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "SquattingGracePeriodUpdated", log); err != nil {
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

// ParseSquattingGracePeriodUpdated is a log parse operation binding the contract event 0x3e452c0fd86ddf6f5ade3ddcc4e4ff19a6daf5e5e0798a19ad36b85299c7c0db.
//
// Solidity: event SquattingGracePeriodUpdated(uint256 oldPeriod, uint256 newPeriod)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseSquattingGracePeriodUpdated(log types.Log) (*BunkerRegistrySquattingGracePeriodUpdated, error) {
	event := new(BunkerRegistrySquattingGracePeriodUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "SquattingGracePeriodUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryStakingContractUpdatedIterator is returned from FilterStakingContractUpdated and is used to iterate over the raw logs and unpacked data for StakingContractUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryStakingContractUpdatedIterator struct {
	Event *BunkerRegistryStakingContractUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryStakingContractUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryStakingContractUpdated)
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
		it.Event = new(BunkerRegistryStakingContractUpdated)
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
func (it *BunkerRegistryStakingContractUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryStakingContractUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryStakingContractUpdated represents a StakingContractUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryStakingContractUpdated struct {
	OldAddr common.Address
	NewAddr common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterStakingContractUpdated is a free log retrieval operation binding the contract event 0x7042586b23181180eb30b4798702d7a0233b7fc2551e89806770e8e5d9392e6a.
//
// Solidity: event StakingContractUpdated(address oldAddr, address newAddr)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterStakingContractUpdated(opts *bind.FilterOpts) (*BunkerRegistryStakingContractUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "StakingContractUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryStakingContractUpdatedIterator{contract: _BunkerRegistry.contract, event: "StakingContractUpdated", logs: logs, sub: sub}, nil
}

// WatchStakingContractUpdated is a free log subscription operation binding the contract event 0x7042586b23181180eb30b4798702d7a0233b7fc2551e89806770e8e5d9392e6a.
//
// Solidity: event StakingContractUpdated(address oldAddr, address newAddr)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchStakingContractUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryStakingContractUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "StakingContractUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryStakingContractUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "StakingContractUpdated", log); err != nil {
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
// Solidity: event StakingContractUpdated(address oldAddr, address newAddr)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseStakingContractUpdated(log types.Log) (*BunkerRegistryStakingContractUpdated, error) {
	event := new(BunkerRegistryStakingContractUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "StakingContractUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistrySubdomainRegisteredIterator is returned from FilterSubdomainRegistered and is used to iterate over the raw logs and unpacked data for SubdomainRegistered events raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainRegisteredIterator struct {
	Event *BunkerRegistrySubdomainRegistered // Event containing the contract specifics and raw log

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
func (it *BunkerRegistrySubdomainRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistrySubdomainRegistered)
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
		it.Event = new(BunkerRegistrySubdomainRegistered)
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
func (it *BunkerRegistrySubdomainRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistrySubdomainRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistrySubdomainRegistered represents a SubdomainRegistered event raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainRegistered struct {
	NameIndexed  common.Hash
	Name         string
	Owner        common.Address
	DeploymentID [32]byte
	Fee          *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterSubdomainRegistered is a free log retrieval operation binding the contract event 0x90cdbde7f62657b2152b1e93f0ac32008757530f4d60f8292ba2e440d0015459.
//
// Solidity: event SubdomainRegistered(string indexed nameIndexed, string name, address indexed owner, bytes32 deploymentID, uint256 fee)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterSubdomainRegistered(opts *bind.FilterOpts, nameIndexed []string, owner []common.Address) (*BunkerRegistrySubdomainRegisteredIterator, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "SubdomainRegistered", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistrySubdomainRegisteredIterator{contract: _BunkerRegistry.contract, event: "SubdomainRegistered", logs: logs, sub: sub}, nil
}

// WatchSubdomainRegistered is a free log subscription operation binding the contract event 0x90cdbde7f62657b2152b1e93f0ac32008757530f4d60f8292ba2e440d0015459.
//
// Solidity: event SubdomainRegistered(string indexed nameIndexed, string name, address indexed owner, bytes32 deploymentID, uint256 fee)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchSubdomainRegistered(opts *bind.WatchOpts, sink chan<- *BunkerRegistrySubdomainRegistered, nameIndexed []string, owner []common.Address) (event.Subscription, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "SubdomainRegistered", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistrySubdomainRegistered)
				if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainRegistered", log); err != nil {
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

// ParseSubdomainRegistered is a log parse operation binding the contract event 0x90cdbde7f62657b2152b1e93f0ac32008757530f4d60f8292ba2e440d0015459.
//
// Solidity: event SubdomainRegistered(string indexed nameIndexed, string name, address indexed owner, bytes32 deploymentID, uint256 fee)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseSubdomainRegistered(log types.Log) (*BunkerRegistrySubdomainRegistered, error) {
	event := new(BunkerRegistrySubdomainRegistered)
	if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistrySubdomainReleasedIterator is returned from FilterSubdomainReleased and is used to iterate over the raw logs and unpacked data for SubdomainReleased events raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainReleasedIterator struct {
	Event *BunkerRegistrySubdomainReleased // Event containing the contract specifics and raw log

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
func (it *BunkerRegistrySubdomainReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistrySubdomainReleased)
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
		it.Event = new(BunkerRegistrySubdomainReleased)
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
func (it *BunkerRegistrySubdomainReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistrySubdomainReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistrySubdomainReleased represents a SubdomainReleased event raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainReleased struct {
	NameIndexed common.Hash
	Name        string
	Owner       common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSubdomainReleased is a free log retrieval operation binding the contract event 0x614e1837690e3a5f44d0508c4bfd8b1ed7c5c0712b3f73bb80bd6182ed3e0711.
//
// Solidity: event SubdomainReleased(string indexed nameIndexed, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterSubdomainReleased(opts *bind.FilterOpts, nameIndexed []string, owner []common.Address) (*BunkerRegistrySubdomainReleasedIterator, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "SubdomainReleased", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistrySubdomainReleasedIterator{contract: _BunkerRegistry.contract, event: "SubdomainReleased", logs: logs, sub: sub}, nil
}

// WatchSubdomainReleased is a free log subscription operation binding the contract event 0x614e1837690e3a5f44d0508c4bfd8b1ed7c5c0712b3f73bb80bd6182ed3e0711.
//
// Solidity: event SubdomainReleased(string indexed nameIndexed, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchSubdomainReleased(opts *bind.WatchOpts, sink chan<- *BunkerRegistrySubdomainReleased, nameIndexed []string, owner []common.Address) (event.Subscription, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "SubdomainReleased", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistrySubdomainReleased)
				if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainReleased", log); err != nil {
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

// ParseSubdomainReleased is a log parse operation binding the contract event 0x614e1837690e3a5f44d0508c4bfd8b1ed7c5c0712b3f73bb80bd6182ed3e0711.
//
// Solidity: event SubdomainReleased(string indexed nameIndexed, string name, address indexed owner)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseSubdomainReleased(log types.Log) (*BunkerRegistrySubdomainReleased, error) {
	event := new(BunkerRegistrySubdomainReleased)
	if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainReleased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistrySubdomainRenewedIterator is returned from FilterSubdomainRenewed and is used to iterate over the raw logs and unpacked data for SubdomainRenewed events raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainRenewedIterator struct {
	Event *BunkerRegistrySubdomainRenewed // Event containing the contract specifics and raw log

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
func (it *BunkerRegistrySubdomainRenewedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistrySubdomainRenewed)
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
		it.Event = new(BunkerRegistrySubdomainRenewed)
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
func (it *BunkerRegistrySubdomainRenewedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistrySubdomainRenewedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistrySubdomainRenewed represents a SubdomainRenewed event raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainRenewed struct {
	NameIndexed common.Hash
	Name        string
	Owner       common.Address
	NewExpiry   *big.Int
	Fee         *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSubdomainRenewed is a free log retrieval operation binding the contract event 0x2e98e19c2a8c2ce70e2441a77a2770df5db240889d406093fd58187d6e1e23df.
//
// Solidity: event SubdomainRenewed(string indexed nameIndexed, string name, address indexed owner, uint48 newExpiry, uint256 fee)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterSubdomainRenewed(opts *bind.FilterOpts, nameIndexed []string, owner []common.Address) (*BunkerRegistrySubdomainRenewedIterator, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "SubdomainRenewed", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistrySubdomainRenewedIterator{contract: _BunkerRegistry.contract, event: "SubdomainRenewed", logs: logs, sub: sub}, nil
}

// WatchSubdomainRenewed is a free log subscription operation binding the contract event 0x2e98e19c2a8c2ce70e2441a77a2770df5db240889d406093fd58187d6e1e23df.
//
// Solidity: event SubdomainRenewed(string indexed nameIndexed, string name, address indexed owner, uint48 newExpiry, uint256 fee)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchSubdomainRenewed(opts *bind.WatchOpts, sink chan<- *BunkerRegistrySubdomainRenewed, nameIndexed []string, owner []common.Address) (event.Subscription, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "SubdomainRenewed", nameIndexedRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistrySubdomainRenewed)
				if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainRenewed", log); err != nil {
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

// ParseSubdomainRenewed is a log parse operation binding the contract event 0x2e98e19c2a8c2ce70e2441a77a2770df5db240889d406093fd58187d6e1e23df.
//
// Solidity: event SubdomainRenewed(string indexed nameIndexed, string name, address indexed owner, uint48 newExpiry, uint256 fee)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseSubdomainRenewed(log types.Log) (*BunkerRegistrySubdomainRenewed, error) {
	event := new(BunkerRegistrySubdomainRenewed)
	if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainRenewed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistrySubdomainReservedIterator is returned from FilterSubdomainReserved and is used to iterate over the raw logs and unpacked data for SubdomainReserved events raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainReservedIterator struct {
	Event *BunkerRegistrySubdomainReserved // Event containing the contract specifics and raw log

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
func (it *BunkerRegistrySubdomainReservedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistrySubdomainReserved)
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
		it.Event = new(BunkerRegistrySubdomainReserved)
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
func (it *BunkerRegistrySubdomainReservedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistrySubdomainReservedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistrySubdomainReserved represents a SubdomainReserved event raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainReserved struct {
	NameIndexed   common.Hash
	Name          string
	Reserver      common.Address
	ReservedUntil *big.Int
	Fee           *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSubdomainReserved is a free log retrieval operation binding the contract event 0xeff1b6ffacd6e50cc00d5a112fdfccf13d7871b360f51c34018056f13aec8406.
//
// Solidity: event SubdomainReserved(string indexed nameIndexed, string name, address indexed reserver, uint48 reservedUntil, uint256 fee)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterSubdomainReserved(opts *bind.FilterOpts, nameIndexed []string, reserver []common.Address) (*BunkerRegistrySubdomainReservedIterator, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var reserverRule []interface{}
	for _, reserverItem := range reserver {
		reserverRule = append(reserverRule, reserverItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "SubdomainReserved", nameIndexedRule, reserverRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistrySubdomainReservedIterator{contract: _BunkerRegistry.contract, event: "SubdomainReserved", logs: logs, sub: sub}, nil
}

// WatchSubdomainReserved is a free log subscription operation binding the contract event 0xeff1b6ffacd6e50cc00d5a112fdfccf13d7871b360f51c34018056f13aec8406.
//
// Solidity: event SubdomainReserved(string indexed nameIndexed, string name, address indexed reserver, uint48 reservedUntil, uint256 fee)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchSubdomainReserved(opts *bind.WatchOpts, sink chan<- *BunkerRegistrySubdomainReserved, nameIndexed []string, reserver []common.Address) (event.Subscription, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var reserverRule []interface{}
	for _, reserverItem := range reserver {
		reserverRule = append(reserverRule, reserverItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "SubdomainReserved", nameIndexedRule, reserverRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistrySubdomainReserved)
				if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainReserved", log); err != nil {
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

// ParseSubdomainReserved is a log parse operation binding the contract event 0xeff1b6ffacd6e50cc00d5a112fdfccf13d7871b360f51c34018056f13aec8406.
//
// Solidity: event SubdomainReserved(string indexed nameIndexed, string name, address indexed reserver, uint48 reservedUntil, uint256 fee)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseSubdomainReserved(log types.Log) (*BunkerRegistrySubdomainReserved, error) {
	event := new(BunkerRegistrySubdomainReserved)
	if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainReserved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistrySubdomainTransferredIterator is returned from FilterSubdomainTransferred and is used to iterate over the raw logs and unpacked data for SubdomainTransferred events raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainTransferredIterator struct {
	Event *BunkerRegistrySubdomainTransferred // Event containing the contract specifics and raw log

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
func (it *BunkerRegistrySubdomainTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistrySubdomainTransferred)
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
		it.Event = new(BunkerRegistrySubdomainTransferred)
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
func (it *BunkerRegistrySubdomainTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistrySubdomainTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistrySubdomainTransferred represents a SubdomainTransferred event raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainTransferred struct {
	NameIndexed common.Hash
	Name        string
	From        common.Address
	To          common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSubdomainTransferred is a free log retrieval operation binding the contract event 0xff6d4da279e86476cdb01668762ba2693d53c1d0169b50ca4c6bb52c1c72f7a1.
//
// Solidity: event SubdomainTransferred(string indexed nameIndexed, string name, address indexed from, address indexed to)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterSubdomainTransferred(opts *bind.FilterOpts, nameIndexed []string, from []common.Address, to []common.Address) (*BunkerRegistrySubdomainTransferredIterator, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "SubdomainTransferred", nameIndexedRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistrySubdomainTransferredIterator{contract: _BunkerRegistry.contract, event: "SubdomainTransferred", logs: logs, sub: sub}, nil
}

// WatchSubdomainTransferred is a free log subscription operation binding the contract event 0xff6d4da279e86476cdb01668762ba2693d53c1d0169b50ca4c6bb52c1c72f7a1.
//
// Solidity: event SubdomainTransferred(string indexed nameIndexed, string name, address indexed from, address indexed to)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchSubdomainTransferred(opts *bind.WatchOpts, sink chan<- *BunkerRegistrySubdomainTransferred, nameIndexed []string, from []common.Address, to []common.Address) (event.Subscription, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "SubdomainTransferred", nameIndexedRule, fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistrySubdomainTransferred)
				if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainTransferred", log); err != nil {
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

// ParseSubdomainTransferred is a log parse operation binding the contract event 0xff6d4da279e86476cdb01668762ba2693d53c1d0169b50ca4c6bb52c1c72f7a1.
//
// Solidity: event SubdomainTransferred(string indexed nameIndexed, string name, address indexed from, address indexed to)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseSubdomainTransferred(log types.Log) (*BunkerRegistrySubdomainTransferred, error) {
	event := new(BunkerRegistrySubdomainTransferred)
	if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistrySubdomainUpdatedIterator is returned from FilterSubdomainUpdated and is used to iterate over the raw logs and unpacked data for SubdomainUpdated events raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainUpdatedIterator struct {
	Event *BunkerRegistrySubdomainUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistrySubdomainUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistrySubdomainUpdated)
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
		it.Event = new(BunkerRegistrySubdomainUpdated)
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
func (it *BunkerRegistrySubdomainUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistrySubdomainUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistrySubdomainUpdated represents a SubdomainUpdated event raised by the BunkerRegistry contract.
type BunkerRegistrySubdomainUpdated struct {
	NameIndexed     common.Hash
	Name            string
	OldDeploymentID [32]byte
	NewDeploymentID [32]byte
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterSubdomainUpdated is a free log retrieval operation binding the contract event 0xed824acc29caa90335a2426c549a404f7cf5863b73983c53b8b03493870d1fd8.
//
// Solidity: event SubdomainUpdated(string indexed nameIndexed, string name, bytes32 oldDeploymentID, bytes32 newDeploymentID)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterSubdomainUpdated(opts *bind.FilterOpts, nameIndexed []string) (*BunkerRegistrySubdomainUpdatedIterator, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "SubdomainUpdated", nameIndexedRule)
	if err != nil {
		return nil, err
	}
	return &BunkerRegistrySubdomainUpdatedIterator{contract: _BunkerRegistry.contract, event: "SubdomainUpdated", logs: logs, sub: sub}, nil
}

// WatchSubdomainUpdated is a free log subscription operation binding the contract event 0xed824acc29caa90335a2426c549a404f7cf5863b73983c53b8b03493870d1fd8.
//
// Solidity: event SubdomainUpdated(string indexed nameIndexed, string name, bytes32 oldDeploymentID, bytes32 newDeploymentID)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchSubdomainUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistrySubdomainUpdated, nameIndexed []string) (event.Subscription, error) {

	var nameIndexedRule []interface{}
	for _, nameIndexedItem := range nameIndexed {
		nameIndexedRule = append(nameIndexedRule, nameIndexedItem)
	}

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "SubdomainUpdated", nameIndexedRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistrySubdomainUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainUpdated", log); err != nil {
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

// ParseSubdomainUpdated is a log parse operation binding the contract event 0xed824acc29caa90335a2426c549a404f7cf5863b73983c53b8b03493870d1fd8.
//
// Solidity: event SubdomainUpdated(string indexed nameIndexed, string name, bytes32 oldDeploymentID, bytes32 newDeploymentID)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseSubdomainUpdated(log types.Log) (*BunkerRegistrySubdomainUpdated, error) {
	event := new(BunkerRegistrySubdomainUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "SubdomainUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryTreasuryUpdatedIterator is returned from FilterTreasuryUpdated and is used to iterate over the raw logs and unpacked data for TreasuryUpdated events raised by the BunkerRegistry contract.
type BunkerRegistryTreasuryUpdatedIterator struct {
	Event *BunkerRegistryTreasuryUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryTreasuryUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryTreasuryUpdated)
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
		it.Event = new(BunkerRegistryTreasuryUpdated)
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
func (it *BunkerRegistryTreasuryUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryTreasuryUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryTreasuryUpdated represents a TreasuryUpdated event raised by the BunkerRegistry contract.
type BunkerRegistryTreasuryUpdated struct {
	OldTreasury common.Address
	NewTreasury common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterTreasuryUpdated is a free log retrieval operation binding the contract event 0x4ab5be82436d353e61ca18726e984e561f5c1cc7c6d38b29d2553c790434705a.
//
// Solidity: event TreasuryUpdated(address oldTreasury, address newTreasury)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterTreasuryUpdated(opts *bind.FilterOpts) (*BunkerRegistryTreasuryUpdatedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "TreasuryUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryTreasuryUpdatedIterator{contract: _BunkerRegistry.contract, event: "TreasuryUpdated", logs: logs, sub: sub}, nil
}

// WatchTreasuryUpdated is a free log subscription operation binding the contract event 0x4ab5be82436d353e61ca18726e984e561f5c1cc7c6d38b29d2553c790434705a.
//
// Solidity: event TreasuryUpdated(address oldTreasury, address newTreasury)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchTreasuryUpdated(opts *bind.WatchOpts, sink chan<- *BunkerRegistryTreasuryUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "TreasuryUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryTreasuryUpdated)
				if err := _BunkerRegistry.contract.UnpackLog(event, "TreasuryUpdated", log); err != nil {
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
// Solidity: event TreasuryUpdated(address oldTreasury, address newTreasury)
func (_BunkerRegistry *BunkerRegistryFilterer) ParseTreasuryUpdated(log types.Log) (*BunkerRegistryTreasuryUpdated, error) {
	event := new(BunkerRegistryTreasuryUpdated)
	if err := _BunkerRegistry.contract.UnpackLog(event, "TreasuryUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerRegistryUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the BunkerRegistry contract.
type BunkerRegistryUnpausedIterator struct {
	Event *BunkerRegistryUnpaused // Event containing the contract specifics and raw log

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
func (it *BunkerRegistryUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerRegistryUnpaused)
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
		it.Event = new(BunkerRegistryUnpaused)
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
func (it *BunkerRegistryUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerRegistryUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerRegistryUnpaused represents a Unpaused event raised by the BunkerRegistry contract.
type BunkerRegistryUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_BunkerRegistry *BunkerRegistryFilterer) FilterUnpaused(opts *bind.FilterOpts) (*BunkerRegistryUnpausedIterator, error) {

	logs, sub, err := _BunkerRegistry.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &BunkerRegistryUnpausedIterator{contract: _BunkerRegistry.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_BunkerRegistry *BunkerRegistryFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *BunkerRegistryUnpaused) (event.Subscription, error) {

	logs, sub, err := _BunkerRegistry.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerRegistryUnpaused)
				if err := _BunkerRegistry.contract.UnpackLog(event, "Unpaused", log); err != nil {
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
func (_BunkerRegistry *BunkerRegistryFilterer) ParseUnpaused(log types.Log) (*BunkerRegistryUnpaused, error) {
	event := new(BunkerRegistryUnpaused)
	if err := _BunkerRegistry.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
