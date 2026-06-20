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

// BunkerStakingProviderInfo is an auto generated low-level Go binding around an user-defined struct.
type BunkerStakingProviderInfo struct {
	StakedAmount   *big.Int
	TotalUnbonding *big.Int
	Beneficiary    common.Address
	RegisteredAt   *big.Int
	Active         bool
	NodeId         [32]byte
	Region         [32]byte
	Capabilities   uint64
	Frozen         bool
}

// BunkerStakingSlashProposal is an auto generated low-level Go binding around an user-defined struct.
type BunkerStakingSlashProposal struct {
	Provider             common.Address
	Amount               *big.Int
	Reason               string
	ProposedAt           *big.Int
	Executed             bool
	Appealed             bool
	Resolved             bool
	SlashReason          uint8
	AppealWindowSnapshot *big.Int
	SlashBurnBpsSnapshot uint16
}

// BunkerStakingUnstakeRequest is an auto generated low-level Go binding around an user-defined struct.
type BunkerStakingUnstakeRequest struct {
	Amount     *big.Int
	UnlockTime *big.Int
	Completed  bool
}

// BunkerStakingMetaData contains all meta data concerning the BunkerStaking contract.
var BunkerStakingMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_treasury\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"BENEFICIARY_TIMELOCK\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"BPS_DENOMINATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_EMISSION_MULTIPLIER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SLASHER_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"acceptOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"appealSlash\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"appealWindow\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"claimRewards\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"claimVestedRewards\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"completeUnstake\",\"inputs\":[{\"name\":\"requestIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"earned\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"emissionMultiplier\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"executeBeneficiaryChange\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"executeSlash\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"freezeProvider\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getClaimableVested\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"claimable\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getProviderInfo\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"info\",\"type\":\"tuple\",\"internalType\":\"structBunkerStaking.ProviderInfo\",\"components\":[{\"name\":\"stakedAmount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"totalUnbonding\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"beneficiary\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"registeredAt\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"active\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"nodeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"region\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"capabilities\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"frozen\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSlashProposal\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"proposal\",\"type\":\"tuple\",\"internalType\":\"structBunkerStaking.SlashProposal\",\"components\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"reason\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"proposedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"executed\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"appealed\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"resolved\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"slashReason\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.SlashReason\"},{\"name\":\"appealWindowSnapshot\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"slashBurnBpsSnapshot\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getSlashableBalance\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"total\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTier\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"tier\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.Tier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTierForAmount\",\"inputs\":[{\"name\":\"stakeAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"tier\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.Tier\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUnstakeQueueLength\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUnstakeRequest\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"request\",\"type\":\"tuple\",\"internalType\":\"structBunkerStaking.UnstakeRequest\",\"components\":[{\"name\":\"amount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"unlockTime\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"completed\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVestedRewardCount\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"immediateReleaseBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initiateBeneficiaryChange\",\"inputs\":[{\"name\":\"newBeneficiary\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isActiveProvider\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"active\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"lastTimeRewardApplicable\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"lastUpdateTime\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxEmissionRate\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"maxTierMultiplierBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nodeIdToProvider\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"notifyRewardAmount\",\"inputs\":[{\"name\":\"reward\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"paused\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingBeneficiaries\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"newBeneficiary\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"effectiveTime\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"pendingOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"periodFinish\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proposeSlash\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"reason\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"proposeSlashByReason\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"reason\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.SlashReason\"}],\"outputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"providers\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"stakedAmount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"totalUnbonding\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"beneficiary\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"registeredAt\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"active\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"nodeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"region\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"capabilities\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"frozen\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"reportComputeHours\",\"inputs\":[{\"name\":\"hours_\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"requestUnstake\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"resolveAppeal\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"uphold\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"rewardPerToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"rewardPerTokenStored\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"rewardRate\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"rewards\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"rewardsDuration\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"rewardsToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setAppealWindow\",\"inputs\":[{\"name\":\"newWindow\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setEmissionMultiplier\",\"inputs\":[{\"name\":\"multiplierBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxEmissionRate\",\"inputs\":[{\"name\":\"maxRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMaxTierMultiplierBps\",\"inputs\":[{\"name\":\"newMax\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setRewardsDuration\",\"inputs\":[{\"name\":\"_rewardsDuration\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSlashFeeSplit\",\"inputs\":[{\"name\":\"burnBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"treasuryBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSlashPercentage\",\"inputs\":[{\"name\":\"reason\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.SlashReason\"},{\"name\":\"bps\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSlashingEnabled\",\"inputs\":[{\"name\":\"enabled\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTierMinStake\",\"inputs\":[{\"name\":\"tier\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.Tier\"},{\"name\":\"minStake\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTierRewardMultiplier\",\"inputs\":[{\"name\":\"tier\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.Tier\"},{\"name\":\"multiplierBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTreasury\",\"inputs\":[{\"name\":\"newTreasury\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setUnbondingPeriod\",\"inputs\":[{\"name\":\"newPeriod\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setVestingParams\",\"inputs\":[{\"name\":\"_vestingPeriod\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_immediateReleaseBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"slash\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"slashBurnBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"slashImmediate\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"slashPercentageBps\",\"inputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.SlashReason\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"slashProposalCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"slashProposals\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"reason\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"proposedAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"executed\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"appealed\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"resolved\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"slashReason\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.SlashReason\"},{\"name\":\"appealWindowSnapshot\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"slashBurnBpsSnapshot\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"slashTreasuryBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"slashingEnabled\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"stake\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"stakeWithIdentity\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"region\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"capabilities\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"tierConfigs\",\"inputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.Tier\"}],\"outputs\":[{\"name\":\"minStake\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxConcurrentJobs\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"rewardMultiplierBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"},{\"name\":\"priorityQueue\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"governance\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"token\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalComputeHoursReported\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalStaked\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalUnbonding\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"treasury\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"unbondingPeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"unfreezeProvider\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unpause\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unstakeQueues\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"amount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"unlockTime\",\"type\":\"uint48\",\"internalType\":\"uint48\"},{\"name\":\"completed\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"updateIdentity\",\"inputs\":[{\"name\":\"nodeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"region\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"capabilities\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"userRewardPerTokenPaid\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"vestedRewards\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"totalAmount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"releasedAmount\",\"type\":\"uint128\",\"internalType\":\"uint128\"},{\"name\":\"vestingStart\",\"type\":\"uint48\",\"internalType\":\"uint48\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"vestingPeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"AppealResolved\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"upheld\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AppealWindowUpdated\",\"inputs\":[{\"name\":\"newWindow\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"BeneficiaryChangeInitiated\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newBeneficiary\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"effectiveTime\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"BeneficiaryChanged\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oldBeneficiary\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newBeneficiary\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ComputeHoursReported\",\"inputs\":[{\"name\":\"hours_\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"totalComputeHours\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EmissionMultiplierUpdated\",\"inputs\":[{\"name\":\"multiplierBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MaxEmissionRateUpdated\",\"inputs\":[{\"name\":\"maxRate\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MaxTierMultiplierUpdated\",\"inputs\":[{\"name\":\"newMax\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferStarted\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Paused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderDeregistered\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderFrozen\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"by\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderIdentityUpdated\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"nodeId\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"region\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"capabilities\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderRegistered\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"stakeAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"tier\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumBunkerStaking.Tier\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProviderUnfrozen\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RewardClaimed\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RewardEpochStarted\",\"inputs\":[{\"name\":\"reward\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"rewardRate\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"periodFinish\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RewardVested\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"totalAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"immediateAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"vestedAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RewardsDurationUpdated\",\"inputs\":[{\"name\":\"newDuration\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SlashAppealed\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SlashFeeSplitUpdated\",\"inputs\":[{\"name\":\"burnBps\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"},{\"name\":\"treasuryBps\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SlashPercentageUpdated\",\"inputs\":[{\"name\":\"reason\",\"type\":\"uint8\",\"indexed\":true,\"internalType\":\"enumBunkerStaking.SlashReason\"},{\"name\":\"bps\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SlashProposalExecuted\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SlashProposed\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"reason\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SlashProposedByReason\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"reason\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumBunkerStaking.SlashReason\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Slashed\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"totalSlashed\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"burnedAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"treasuryAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SlashingEnabledUpdated\",\"inputs\":[{\"name\":\"enabled\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Staked\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"totalStake\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"tier\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumBunkerStaking.Tier\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TierConfigUpdated\",\"inputs\":[{\"name\":\"tier\",\"type\":\"uint8\",\"indexed\":true,\"internalType\":\"enumBunkerStaking.Tier\"},{\"name\":\"minStake\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TierRewardMultiplierUpdated\",\"inputs\":[{\"name\":\"tier\",\"type\":\"uint8\",\"indexed\":true,\"internalType\":\"enumBunkerStaking.Tier\"},{\"name\":\"multiplierBps\",\"type\":\"uint16\",\"indexed\":false,\"internalType\":\"uint16\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TreasuryUpdated\",\"inputs\":[{\"name\":\"oldTreasury\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newTreasury\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UnbondingPeriodUpdated\",\"inputs\":[{\"name\":\"newPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Unpaused\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UnstakeCompleted\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"requestIndex\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UnstakeRequested\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"unlockTime\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"requestIndex\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"VestedRewardsClaimed\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"VestedRewardsForfeited\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"VestingParamsUpdated\",\"inputs\":[{\"name\":\"vestingPeriod\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"immediateReleaseBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"AlreadyRegistered\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"AppealWindowElapsed\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"AppealWindowNotElapsed\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"windowEnd\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"BelowMinimumStake\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"minimum\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"BeneficiaryTimelockNotElapsed\",\"inputs\":[{\"name\":\"effectiveTime\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"currentTime\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"CannotConfigureNoneTier\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EnforcedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ExpectedPause\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientSlashableBalance\",\"inputs\":[{\"name\":\"requested\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"available\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InsufficientStake\",\"inputs\":[{\"name\":\"requested\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"available\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidAppealWindow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidEmissionMultiplier\",\"inputs\":[{\"name\":\"multiplier\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidFeeSplit\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidImmediateReleaseBps\",\"inputs\":[{\"name\":\"bps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidMultiplierCap\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidProposalId\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidSlashPercentage\",\"inputs\":[{\"name\":\"bps\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]},{\"type\":\"error\",\"name\":\"InvalidSlashReason\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidUnbondingPeriod\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidUnstakeIndex\",\"inputs\":[{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"queueLength\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidVestingPeriod\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MultiplierTooHigh\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MultiplierTooLow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NoPendingBeneficiaryChange\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"NoRewardsToClaimError\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NoVestedRewardsClaimable\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NodeIdAlreadyClaimed\",\"inputs\":[{\"name\":\"nodeId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"NotProposalProvider\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"caller\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ProposalAlreadyAppealed\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ProposalAlreadyExecuted\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ProposalAlreadyResolved\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ProposalAppealed\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ProposalNotAppealed\",\"inputs\":[{\"name\":\"proposalId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ProviderIsFrozen\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ProviderNotActive\",\"inputs\":[{\"name\":\"provider\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"RewardDurationNotFinished\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"RewardRateTooHigh\",\"inputs\":[{\"name\":\"rewardRate\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"balance\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"RewardsDurationZero\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"SlashingNotEnabled\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TierOrderViolation\",\"inputs\":[{\"name\":\"tier\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.Tier\"},{\"name\":\"minStake\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"adjacentTier\",\"type\":\"uint8\",\"internalType\":\"enumBunkerStaking.Tier\"},{\"name\":\"adjacentMinStake\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"TokenBurnFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TooManyUnstakeRequests\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UnbondingNotReady\",\"inputs\":[{\"name\":\"unlockTime\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"currentTime\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"UnstakeAlreadyCompleted\",\"inputs\":[{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"UseUpdateIdentity\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroAmount\",\"inputs\":[]}]",
}

// BunkerStakingABI is the input ABI used to generate the binding from.
// Deprecated: Use BunkerStakingMetaData.ABI instead.
var BunkerStakingABI = BunkerStakingMetaData.ABI

// BunkerStaking is an auto generated Go binding around an Ethereum contract.
type BunkerStaking struct {
	BunkerStakingCaller     // Read-only binding to the contract
	BunkerStakingTransactor // Write-only binding to the contract
	BunkerStakingFilterer   // Log filterer for contract events
}

// BunkerStakingCaller is an auto generated read-only Go binding around an Ethereum contract.
type BunkerStakingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerStakingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BunkerStakingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerStakingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BunkerStakingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BunkerStakingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BunkerStakingSession struct {
	Contract     *BunkerStaking    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BunkerStakingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BunkerStakingCallerSession struct {
	Contract *BunkerStakingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// BunkerStakingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BunkerStakingTransactorSession struct {
	Contract     *BunkerStakingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// BunkerStakingRaw is an auto generated low-level Go binding around an Ethereum contract.
type BunkerStakingRaw struct {
	Contract *BunkerStaking // Generic contract binding to access the raw methods on
}

// BunkerStakingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BunkerStakingCallerRaw struct {
	Contract *BunkerStakingCaller // Generic read-only contract binding to access the raw methods on
}

// BunkerStakingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BunkerStakingTransactorRaw struct {
	Contract *BunkerStakingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBunkerStaking creates a new instance of BunkerStaking, bound to a specific deployed contract.
func NewBunkerStaking(address common.Address, backend bind.ContractBackend) (*BunkerStaking, error) {
	contract, err := bindBunkerStaking(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BunkerStaking{BunkerStakingCaller: BunkerStakingCaller{contract: contract}, BunkerStakingTransactor: BunkerStakingTransactor{contract: contract}, BunkerStakingFilterer: BunkerStakingFilterer{contract: contract}}, nil
}

// NewBunkerStakingCaller creates a new read-only instance of BunkerStaking, bound to a specific deployed contract.
func NewBunkerStakingCaller(address common.Address, caller bind.ContractCaller) (*BunkerStakingCaller, error) {
	contract, err := bindBunkerStaking(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingCaller{contract: contract}, nil
}

// NewBunkerStakingTransactor creates a new write-only instance of BunkerStaking, bound to a specific deployed contract.
func NewBunkerStakingTransactor(address common.Address, transactor bind.ContractTransactor) (*BunkerStakingTransactor, error) {
	contract, err := bindBunkerStaking(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingTransactor{contract: contract}, nil
}

// NewBunkerStakingFilterer creates a new log filterer instance of BunkerStaking, bound to a specific deployed contract.
func NewBunkerStakingFilterer(address common.Address, filterer bind.ContractFilterer) (*BunkerStakingFilterer, error) {
	contract, err := bindBunkerStaking(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingFilterer{contract: contract}, nil
}

// bindBunkerStaking binds a generic wrapper to an already deployed contract.
func bindBunkerStaking(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BunkerStakingMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerStaking *BunkerStakingRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerStaking.Contract.BunkerStakingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerStaking *BunkerStakingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerStaking.Contract.BunkerStakingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerStaking *BunkerStakingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerStaking.Contract.BunkerStakingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BunkerStaking *BunkerStakingCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BunkerStaking.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BunkerStaking *BunkerStakingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerStaking.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BunkerStaking *BunkerStakingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BunkerStaking.Contract.contract.Transact(opts, method, params...)
}

// BENEFICIARYTIMELOCK is a free data retrieval call binding the contract method 0x982bd124.
//
// Solidity: function BENEFICIARY_TIMELOCK() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) BENEFICIARYTIMELOCK(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "BENEFICIARY_TIMELOCK")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BENEFICIARYTIMELOCK is a free data retrieval call binding the contract method 0x982bd124.
//
// Solidity: function BENEFICIARY_TIMELOCK() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) BENEFICIARYTIMELOCK() (*big.Int, error) {
	return _BunkerStaking.Contract.BENEFICIARYTIMELOCK(&_BunkerStaking.CallOpts)
}

// BENEFICIARYTIMELOCK is a free data retrieval call binding the contract method 0x982bd124.
//
// Solidity: function BENEFICIARY_TIMELOCK() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) BENEFICIARYTIMELOCK() (*big.Int, error) {
	return _BunkerStaking.Contract.BENEFICIARYTIMELOCK(&_BunkerStaking.CallOpts)
}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) BPSDENOMINATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "BPS_DENOMINATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) BPSDENOMINATOR() (*big.Int, error) {
	return _BunkerStaking.Contract.BPSDENOMINATOR(&_BunkerStaking.CallOpts)
}

// BPSDENOMINATOR is a free data retrieval call binding the contract method 0xe1a45218.
//
// Solidity: function BPS_DENOMINATOR() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) BPSDENOMINATOR() (*big.Int, error) {
	return _BunkerStaking.Contract.BPSDENOMINATOR(&_BunkerStaking.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerStaking *BunkerStakingCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerStaking *BunkerStakingSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _BunkerStaking.Contract.DEFAULTADMINROLE(&_BunkerStaking.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_BunkerStaking *BunkerStakingCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _BunkerStaking.Contract.DEFAULTADMINROLE(&_BunkerStaking.CallOpts)
}

// MAXEMISSIONMULTIPLIER is a free data retrieval call binding the contract method 0xa452ab9f.
//
// Solidity: function MAX_EMISSION_MULTIPLIER() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) MAXEMISSIONMULTIPLIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "MAX_EMISSION_MULTIPLIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXEMISSIONMULTIPLIER is a free data retrieval call binding the contract method 0xa452ab9f.
//
// Solidity: function MAX_EMISSION_MULTIPLIER() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) MAXEMISSIONMULTIPLIER() (*big.Int, error) {
	return _BunkerStaking.Contract.MAXEMISSIONMULTIPLIER(&_BunkerStaking.CallOpts)
}

// MAXEMISSIONMULTIPLIER is a free data retrieval call binding the contract method 0xa452ab9f.
//
// Solidity: function MAX_EMISSION_MULTIPLIER() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) MAXEMISSIONMULTIPLIER() (*big.Int, error) {
	return _BunkerStaking.Contract.MAXEMISSIONMULTIPLIER(&_BunkerStaking.CallOpts)
}

// SLASHERROLE is a free data retrieval call binding the contract method 0x5095af64.
//
// Solidity: function SLASHER_ROLE() view returns(bytes32)
func (_BunkerStaking *BunkerStakingCaller) SLASHERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "SLASHER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SLASHERROLE is a free data retrieval call binding the contract method 0x5095af64.
//
// Solidity: function SLASHER_ROLE() view returns(bytes32)
func (_BunkerStaking *BunkerStakingSession) SLASHERROLE() ([32]byte, error) {
	return _BunkerStaking.Contract.SLASHERROLE(&_BunkerStaking.CallOpts)
}

// SLASHERROLE is a free data retrieval call binding the contract method 0x5095af64.
//
// Solidity: function SLASHER_ROLE() view returns(bytes32)
func (_BunkerStaking *BunkerStakingCallerSession) SLASHERROLE() ([32]byte, error) {
	return _BunkerStaking.Contract.SLASHERROLE(&_BunkerStaking.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerStaking *BunkerStakingCaller) VERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerStaking *BunkerStakingSession) VERSION() (string, error) {
	return _BunkerStaking.Contract.VERSION(&_BunkerStaking.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(string)
func (_BunkerStaking *BunkerStakingCallerSession) VERSION() (string, error) {
	return _BunkerStaking.Contract.VERSION(&_BunkerStaking.CallOpts)
}

// AppealWindow is a free data retrieval call binding the contract method 0x3fd2938a.
//
// Solidity: function appealWindow() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) AppealWindow(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "appealWindow")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AppealWindow is a free data retrieval call binding the contract method 0x3fd2938a.
//
// Solidity: function appealWindow() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) AppealWindow() (*big.Int, error) {
	return _BunkerStaking.Contract.AppealWindow(&_BunkerStaking.CallOpts)
}

// AppealWindow is a free data retrieval call binding the contract method 0x3fd2938a.
//
// Solidity: function appealWindow() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) AppealWindow() (*big.Int, error) {
	return _BunkerStaking.Contract.AppealWindow(&_BunkerStaking.CallOpts)
}

// Earned is a free data retrieval call binding the contract method 0x008cc262.
//
// Solidity: function earned(address provider) view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) Earned(opts *bind.CallOpts, provider common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "earned", provider)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Earned is a free data retrieval call binding the contract method 0x008cc262.
//
// Solidity: function earned(address provider) view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) Earned(provider common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.Earned(&_BunkerStaking.CallOpts, provider)
}

// Earned is a free data retrieval call binding the contract method 0x008cc262.
//
// Solidity: function earned(address provider) view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) Earned(provider common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.Earned(&_BunkerStaking.CallOpts, provider)
}

// EmissionMultiplier is a free data retrieval call binding the contract method 0x37de2e19.
//
// Solidity: function emissionMultiplier() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) EmissionMultiplier(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "emissionMultiplier")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EmissionMultiplier is a free data retrieval call binding the contract method 0x37de2e19.
//
// Solidity: function emissionMultiplier() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) EmissionMultiplier() (*big.Int, error) {
	return _BunkerStaking.Contract.EmissionMultiplier(&_BunkerStaking.CallOpts)
}

// EmissionMultiplier is a free data retrieval call binding the contract method 0x37de2e19.
//
// Solidity: function emissionMultiplier() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) EmissionMultiplier() (*big.Int, error) {
	return _BunkerStaking.Contract.EmissionMultiplier(&_BunkerStaking.CallOpts)
}

// GetClaimableVested is a free data retrieval call binding the contract method 0x73522681.
//
// Solidity: function getClaimableVested(address provider) view returns(uint256 claimable)
func (_BunkerStaking *BunkerStakingCaller) GetClaimableVested(opts *bind.CallOpts, provider common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "getClaimableVested", provider)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetClaimableVested is a free data retrieval call binding the contract method 0x73522681.
//
// Solidity: function getClaimableVested(address provider) view returns(uint256 claimable)
func (_BunkerStaking *BunkerStakingSession) GetClaimableVested(provider common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.GetClaimableVested(&_BunkerStaking.CallOpts, provider)
}

// GetClaimableVested is a free data retrieval call binding the contract method 0x73522681.
//
// Solidity: function getClaimableVested(address provider) view returns(uint256 claimable)
func (_BunkerStaking *BunkerStakingCallerSession) GetClaimableVested(provider common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.GetClaimableVested(&_BunkerStaking.CallOpts, provider)
}

// GetProviderInfo is a free data retrieval call binding the contract method 0x7583902f.
//
// Solidity: function getProviderInfo(address provider) view returns((uint128,uint128,address,uint48,bool,bytes32,bytes32,uint64,bool) info)
func (_BunkerStaking *BunkerStakingCaller) GetProviderInfo(opts *bind.CallOpts, provider common.Address) (BunkerStakingProviderInfo, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "getProviderInfo", provider)

	if err != nil {
		return *new(BunkerStakingProviderInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerStakingProviderInfo)).(*BunkerStakingProviderInfo)

	return out0, err

}

// GetProviderInfo is a free data retrieval call binding the contract method 0x7583902f.
//
// Solidity: function getProviderInfo(address provider) view returns((uint128,uint128,address,uint48,bool,bytes32,bytes32,uint64,bool) info)
func (_BunkerStaking *BunkerStakingSession) GetProviderInfo(provider common.Address) (BunkerStakingProviderInfo, error) {
	return _BunkerStaking.Contract.GetProviderInfo(&_BunkerStaking.CallOpts, provider)
}

// GetProviderInfo is a free data retrieval call binding the contract method 0x7583902f.
//
// Solidity: function getProviderInfo(address provider) view returns((uint128,uint128,address,uint48,bool,bytes32,bytes32,uint64,bool) info)
func (_BunkerStaking *BunkerStakingCallerSession) GetProviderInfo(provider common.Address) (BunkerStakingProviderInfo, error) {
	return _BunkerStaking.Contract.GetProviderInfo(&_BunkerStaking.CallOpts, provider)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerStaking *BunkerStakingCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerStaking *BunkerStakingSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _BunkerStaking.Contract.GetRoleAdmin(&_BunkerStaking.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_BunkerStaking *BunkerStakingCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _BunkerStaking.Contract.GetRoleAdmin(&_BunkerStaking.CallOpts, role)
}

// GetSlashProposal is a free data retrieval call binding the contract method 0x97b5103c.
//
// Solidity: function getSlashProposal(uint256 proposalId) view returns((address,uint256,string,uint256,bool,bool,bool,uint8,uint256,uint16) proposal)
func (_BunkerStaking *BunkerStakingCaller) GetSlashProposal(opts *bind.CallOpts, proposalId *big.Int) (BunkerStakingSlashProposal, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "getSlashProposal", proposalId)

	if err != nil {
		return *new(BunkerStakingSlashProposal), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerStakingSlashProposal)).(*BunkerStakingSlashProposal)

	return out0, err

}

// GetSlashProposal is a free data retrieval call binding the contract method 0x97b5103c.
//
// Solidity: function getSlashProposal(uint256 proposalId) view returns((address,uint256,string,uint256,bool,bool,bool,uint8,uint256,uint16) proposal)
func (_BunkerStaking *BunkerStakingSession) GetSlashProposal(proposalId *big.Int) (BunkerStakingSlashProposal, error) {
	return _BunkerStaking.Contract.GetSlashProposal(&_BunkerStaking.CallOpts, proposalId)
}

// GetSlashProposal is a free data retrieval call binding the contract method 0x97b5103c.
//
// Solidity: function getSlashProposal(uint256 proposalId) view returns((address,uint256,string,uint256,bool,bool,bool,uint8,uint256,uint16) proposal)
func (_BunkerStaking *BunkerStakingCallerSession) GetSlashProposal(proposalId *big.Int) (BunkerStakingSlashProposal, error) {
	return _BunkerStaking.Contract.GetSlashProposal(&_BunkerStaking.CallOpts, proposalId)
}

// GetSlashableBalance is a free data retrieval call binding the contract method 0x61860e4c.
//
// Solidity: function getSlashableBalance(address provider) view returns(uint256 total)
func (_BunkerStaking *BunkerStakingCaller) GetSlashableBalance(opts *bind.CallOpts, provider common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "getSlashableBalance", provider)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSlashableBalance is a free data retrieval call binding the contract method 0x61860e4c.
//
// Solidity: function getSlashableBalance(address provider) view returns(uint256 total)
func (_BunkerStaking *BunkerStakingSession) GetSlashableBalance(provider common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.GetSlashableBalance(&_BunkerStaking.CallOpts, provider)
}

// GetSlashableBalance is a free data retrieval call binding the contract method 0x61860e4c.
//
// Solidity: function getSlashableBalance(address provider) view returns(uint256 total)
func (_BunkerStaking *BunkerStakingCallerSession) GetSlashableBalance(provider common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.GetSlashableBalance(&_BunkerStaking.CallOpts, provider)
}

// GetTier is a free data retrieval call binding the contract method 0xb45aae52.
//
// Solidity: function getTier(address provider) view returns(uint8 tier)
func (_BunkerStaking *BunkerStakingCaller) GetTier(opts *bind.CallOpts, provider common.Address) (uint8, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "getTier", provider)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetTier is a free data retrieval call binding the contract method 0xb45aae52.
//
// Solidity: function getTier(address provider) view returns(uint8 tier)
func (_BunkerStaking *BunkerStakingSession) GetTier(provider common.Address) (uint8, error) {
	return _BunkerStaking.Contract.GetTier(&_BunkerStaking.CallOpts, provider)
}

// GetTier is a free data retrieval call binding the contract method 0xb45aae52.
//
// Solidity: function getTier(address provider) view returns(uint8 tier)
func (_BunkerStaking *BunkerStakingCallerSession) GetTier(provider common.Address) (uint8, error) {
	return _BunkerStaking.Contract.GetTier(&_BunkerStaking.CallOpts, provider)
}

// GetTierForAmount is a free data retrieval call binding the contract method 0x1237739b.
//
// Solidity: function getTierForAmount(uint256 stakeAmount) view returns(uint8 tier)
func (_BunkerStaking *BunkerStakingCaller) GetTierForAmount(opts *bind.CallOpts, stakeAmount *big.Int) (uint8, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "getTierForAmount", stakeAmount)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetTierForAmount is a free data retrieval call binding the contract method 0x1237739b.
//
// Solidity: function getTierForAmount(uint256 stakeAmount) view returns(uint8 tier)
func (_BunkerStaking *BunkerStakingSession) GetTierForAmount(stakeAmount *big.Int) (uint8, error) {
	return _BunkerStaking.Contract.GetTierForAmount(&_BunkerStaking.CallOpts, stakeAmount)
}

// GetTierForAmount is a free data retrieval call binding the contract method 0x1237739b.
//
// Solidity: function getTierForAmount(uint256 stakeAmount) view returns(uint8 tier)
func (_BunkerStaking *BunkerStakingCallerSession) GetTierForAmount(stakeAmount *big.Int) (uint8, error) {
	return _BunkerStaking.Contract.GetTierForAmount(&_BunkerStaking.CallOpts, stakeAmount)
}

// GetUnstakeQueueLength is a free data retrieval call binding the contract method 0x2eb3328a.
//
// Solidity: function getUnstakeQueueLength(address provider) view returns(uint256 count)
func (_BunkerStaking *BunkerStakingCaller) GetUnstakeQueueLength(opts *bind.CallOpts, provider common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "getUnstakeQueueLength", provider)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUnstakeQueueLength is a free data retrieval call binding the contract method 0x2eb3328a.
//
// Solidity: function getUnstakeQueueLength(address provider) view returns(uint256 count)
func (_BunkerStaking *BunkerStakingSession) GetUnstakeQueueLength(provider common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.GetUnstakeQueueLength(&_BunkerStaking.CallOpts, provider)
}

// GetUnstakeQueueLength is a free data retrieval call binding the contract method 0x2eb3328a.
//
// Solidity: function getUnstakeQueueLength(address provider) view returns(uint256 count)
func (_BunkerStaking *BunkerStakingCallerSession) GetUnstakeQueueLength(provider common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.GetUnstakeQueueLength(&_BunkerStaking.CallOpts, provider)
}

// GetUnstakeRequest is a free data retrieval call binding the contract method 0x4ae6c4ae.
//
// Solidity: function getUnstakeRequest(address provider, uint256 index) view returns((uint128,uint48,bool) request)
func (_BunkerStaking *BunkerStakingCaller) GetUnstakeRequest(opts *bind.CallOpts, provider common.Address, index *big.Int) (BunkerStakingUnstakeRequest, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "getUnstakeRequest", provider, index)

	if err != nil {
		return *new(BunkerStakingUnstakeRequest), err
	}

	out0 := *abi.ConvertType(out[0], new(BunkerStakingUnstakeRequest)).(*BunkerStakingUnstakeRequest)

	return out0, err

}

// GetUnstakeRequest is a free data retrieval call binding the contract method 0x4ae6c4ae.
//
// Solidity: function getUnstakeRequest(address provider, uint256 index) view returns((uint128,uint48,bool) request)
func (_BunkerStaking *BunkerStakingSession) GetUnstakeRequest(provider common.Address, index *big.Int) (BunkerStakingUnstakeRequest, error) {
	return _BunkerStaking.Contract.GetUnstakeRequest(&_BunkerStaking.CallOpts, provider, index)
}

// GetUnstakeRequest is a free data retrieval call binding the contract method 0x4ae6c4ae.
//
// Solidity: function getUnstakeRequest(address provider, uint256 index) view returns((uint128,uint48,bool) request)
func (_BunkerStaking *BunkerStakingCallerSession) GetUnstakeRequest(provider common.Address, index *big.Int) (BunkerStakingUnstakeRequest, error) {
	return _BunkerStaking.Contract.GetUnstakeRequest(&_BunkerStaking.CallOpts, provider, index)
}

// GetVestedRewardCount is a free data retrieval call binding the contract method 0xc4a24845.
//
// Solidity: function getVestedRewardCount(address provider) view returns(uint256 count)
func (_BunkerStaking *BunkerStakingCaller) GetVestedRewardCount(opts *bind.CallOpts, provider common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "getVestedRewardCount", provider)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetVestedRewardCount is a free data retrieval call binding the contract method 0xc4a24845.
//
// Solidity: function getVestedRewardCount(address provider) view returns(uint256 count)
func (_BunkerStaking *BunkerStakingSession) GetVestedRewardCount(provider common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.GetVestedRewardCount(&_BunkerStaking.CallOpts, provider)
}

// GetVestedRewardCount is a free data retrieval call binding the contract method 0xc4a24845.
//
// Solidity: function getVestedRewardCount(address provider) view returns(uint256 count)
func (_BunkerStaking *BunkerStakingCallerSession) GetVestedRewardCount(provider common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.GetVestedRewardCount(&_BunkerStaking.CallOpts, provider)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerStaking *BunkerStakingCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerStaking *BunkerStakingSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _BunkerStaking.Contract.HasRole(&_BunkerStaking.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_BunkerStaking *BunkerStakingCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _BunkerStaking.Contract.HasRole(&_BunkerStaking.CallOpts, role, account)
}

// ImmediateReleaseBps is a free data retrieval call binding the contract method 0x193f8d93.
//
// Solidity: function immediateReleaseBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) ImmediateReleaseBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "immediateReleaseBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ImmediateReleaseBps is a free data retrieval call binding the contract method 0x193f8d93.
//
// Solidity: function immediateReleaseBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) ImmediateReleaseBps() (*big.Int, error) {
	return _BunkerStaking.Contract.ImmediateReleaseBps(&_BunkerStaking.CallOpts)
}

// ImmediateReleaseBps is a free data retrieval call binding the contract method 0x193f8d93.
//
// Solidity: function immediateReleaseBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) ImmediateReleaseBps() (*big.Int, error) {
	return _BunkerStaking.Contract.ImmediateReleaseBps(&_BunkerStaking.CallOpts)
}

// IsActiveProvider is a free data retrieval call binding the contract method 0x9b5e1a3b.
//
// Solidity: function isActiveProvider(address provider) view returns(bool active)
func (_BunkerStaking *BunkerStakingCaller) IsActiveProvider(opts *bind.CallOpts, provider common.Address) (bool, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "isActiveProvider", provider)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsActiveProvider is a free data retrieval call binding the contract method 0x9b5e1a3b.
//
// Solidity: function isActiveProvider(address provider) view returns(bool active)
func (_BunkerStaking *BunkerStakingSession) IsActiveProvider(provider common.Address) (bool, error) {
	return _BunkerStaking.Contract.IsActiveProvider(&_BunkerStaking.CallOpts, provider)
}

// IsActiveProvider is a free data retrieval call binding the contract method 0x9b5e1a3b.
//
// Solidity: function isActiveProvider(address provider) view returns(bool active)
func (_BunkerStaking *BunkerStakingCallerSession) IsActiveProvider(provider common.Address) (bool, error) {
	return _BunkerStaking.Contract.IsActiveProvider(&_BunkerStaking.CallOpts, provider)
}

// LastTimeRewardApplicable is a free data retrieval call binding the contract method 0x80faa57d.
//
// Solidity: function lastTimeRewardApplicable() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) LastTimeRewardApplicable(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "lastTimeRewardApplicable")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastTimeRewardApplicable is a free data retrieval call binding the contract method 0x80faa57d.
//
// Solidity: function lastTimeRewardApplicable() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) LastTimeRewardApplicable() (*big.Int, error) {
	return _BunkerStaking.Contract.LastTimeRewardApplicable(&_BunkerStaking.CallOpts)
}

// LastTimeRewardApplicable is a free data retrieval call binding the contract method 0x80faa57d.
//
// Solidity: function lastTimeRewardApplicable() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) LastTimeRewardApplicable() (*big.Int, error) {
	return _BunkerStaking.Contract.LastTimeRewardApplicable(&_BunkerStaking.CallOpts)
}

// LastUpdateTime is a free data retrieval call binding the contract method 0xc8f33c91.
//
// Solidity: function lastUpdateTime() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) LastUpdateTime(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "lastUpdateTime")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastUpdateTime is a free data retrieval call binding the contract method 0xc8f33c91.
//
// Solidity: function lastUpdateTime() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) LastUpdateTime() (*big.Int, error) {
	return _BunkerStaking.Contract.LastUpdateTime(&_BunkerStaking.CallOpts)
}

// LastUpdateTime is a free data retrieval call binding the contract method 0xc8f33c91.
//
// Solidity: function lastUpdateTime() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) LastUpdateTime() (*big.Int, error) {
	return _BunkerStaking.Contract.LastUpdateTime(&_BunkerStaking.CallOpts)
}

// MaxEmissionRate is a free data retrieval call binding the contract method 0x142fc582.
//
// Solidity: function maxEmissionRate() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) MaxEmissionRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "maxEmissionRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxEmissionRate is a free data retrieval call binding the contract method 0x142fc582.
//
// Solidity: function maxEmissionRate() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) MaxEmissionRate() (*big.Int, error) {
	return _BunkerStaking.Contract.MaxEmissionRate(&_BunkerStaking.CallOpts)
}

// MaxEmissionRate is a free data retrieval call binding the contract method 0x142fc582.
//
// Solidity: function maxEmissionRate() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) MaxEmissionRate() (*big.Int, error) {
	return _BunkerStaking.Contract.MaxEmissionRate(&_BunkerStaking.CallOpts)
}

// MaxTierMultiplierBps is a free data retrieval call binding the contract method 0xedb38703.
//
// Solidity: function maxTierMultiplierBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) MaxTierMultiplierBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "maxTierMultiplierBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxTierMultiplierBps is a free data retrieval call binding the contract method 0xedb38703.
//
// Solidity: function maxTierMultiplierBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) MaxTierMultiplierBps() (*big.Int, error) {
	return _BunkerStaking.Contract.MaxTierMultiplierBps(&_BunkerStaking.CallOpts)
}

// MaxTierMultiplierBps is a free data retrieval call binding the contract method 0xedb38703.
//
// Solidity: function maxTierMultiplierBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) MaxTierMultiplierBps() (*big.Int, error) {
	return _BunkerStaking.Contract.MaxTierMultiplierBps(&_BunkerStaking.CallOpts)
}

// NodeIdToProvider is a free data retrieval call binding the contract method 0x04e05e8c.
//
// Solidity: function nodeIdToProvider(bytes32 ) view returns(address)
func (_BunkerStaking *BunkerStakingCaller) NodeIdToProvider(opts *bind.CallOpts, arg0 [32]byte) (common.Address, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "nodeIdToProvider", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NodeIdToProvider is a free data retrieval call binding the contract method 0x04e05e8c.
//
// Solidity: function nodeIdToProvider(bytes32 ) view returns(address)
func (_BunkerStaking *BunkerStakingSession) NodeIdToProvider(arg0 [32]byte) (common.Address, error) {
	return _BunkerStaking.Contract.NodeIdToProvider(&_BunkerStaking.CallOpts, arg0)
}

// NodeIdToProvider is a free data retrieval call binding the contract method 0x04e05e8c.
//
// Solidity: function nodeIdToProvider(bytes32 ) view returns(address)
func (_BunkerStaking *BunkerStakingCallerSession) NodeIdToProvider(arg0 [32]byte) (common.Address, error) {
	return _BunkerStaking.Contract.NodeIdToProvider(&_BunkerStaking.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerStaking *BunkerStakingCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerStaking *BunkerStakingSession) Owner() (common.Address, error) {
	return _BunkerStaking.Contract.Owner(&_BunkerStaking.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BunkerStaking *BunkerStakingCallerSession) Owner() (common.Address, error) {
	return _BunkerStaking.Contract.Owner(&_BunkerStaking.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerStaking *BunkerStakingCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerStaking *BunkerStakingSession) Paused() (bool, error) {
	return _BunkerStaking.Contract.Paused(&_BunkerStaking.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_BunkerStaking *BunkerStakingCallerSession) Paused() (bool, error) {
	return _BunkerStaking.Contract.Paused(&_BunkerStaking.CallOpts)
}

// PendingBeneficiaries is a free data retrieval call binding the contract method 0x73605e7a.
//
// Solidity: function pendingBeneficiaries(address ) view returns(address newBeneficiary, uint48 effectiveTime)
func (_BunkerStaking *BunkerStakingCaller) PendingBeneficiaries(opts *bind.CallOpts, arg0 common.Address) (struct {
	NewBeneficiary common.Address
	EffectiveTime  *big.Int
}, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "pendingBeneficiaries", arg0)

	outstruct := new(struct {
		NewBeneficiary common.Address
		EffectiveTime  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.NewBeneficiary = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.EffectiveTime = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// PendingBeneficiaries is a free data retrieval call binding the contract method 0x73605e7a.
//
// Solidity: function pendingBeneficiaries(address ) view returns(address newBeneficiary, uint48 effectiveTime)
func (_BunkerStaking *BunkerStakingSession) PendingBeneficiaries(arg0 common.Address) (struct {
	NewBeneficiary common.Address
	EffectiveTime  *big.Int
}, error) {
	return _BunkerStaking.Contract.PendingBeneficiaries(&_BunkerStaking.CallOpts, arg0)
}

// PendingBeneficiaries is a free data retrieval call binding the contract method 0x73605e7a.
//
// Solidity: function pendingBeneficiaries(address ) view returns(address newBeneficiary, uint48 effectiveTime)
func (_BunkerStaking *BunkerStakingCallerSession) PendingBeneficiaries(arg0 common.Address) (struct {
	NewBeneficiary common.Address
	EffectiveTime  *big.Int
}, error) {
	return _BunkerStaking.Contract.PendingBeneficiaries(&_BunkerStaking.CallOpts, arg0)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerStaking *BunkerStakingCaller) PendingOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "pendingOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerStaking *BunkerStakingSession) PendingOwner() (common.Address, error) {
	return _BunkerStaking.Contract.PendingOwner(&_BunkerStaking.CallOpts)
}

// PendingOwner is a free data retrieval call binding the contract method 0xe30c3978.
//
// Solidity: function pendingOwner() view returns(address)
func (_BunkerStaking *BunkerStakingCallerSession) PendingOwner() (common.Address, error) {
	return _BunkerStaking.Contract.PendingOwner(&_BunkerStaking.CallOpts)
}

// PeriodFinish is a free data retrieval call binding the contract method 0xebe2b12b.
//
// Solidity: function periodFinish() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) PeriodFinish(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "periodFinish")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PeriodFinish is a free data retrieval call binding the contract method 0xebe2b12b.
//
// Solidity: function periodFinish() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) PeriodFinish() (*big.Int, error) {
	return _BunkerStaking.Contract.PeriodFinish(&_BunkerStaking.CallOpts)
}

// PeriodFinish is a free data retrieval call binding the contract method 0xebe2b12b.
//
// Solidity: function periodFinish() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) PeriodFinish() (*big.Int, error) {
	return _BunkerStaking.Contract.PeriodFinish(&_BunkerStaking.CallOpts)
}

// Providers is a free data retrieval call binding the contract method 0x0787bc27.
//
// Solidity: function providers(address ) view returns(uint128 stakedAmount, uint128 totalUnbonding, address beneficiary, uint48 registeredAt, bool active, bytes32 nodeId, bytes32 region, uint64 capabilities, bool frozen)
func (_BunkerStaking *BunkerStakingCaller) Providers(opts *bind.CallOpts, arg0 common.Address) (struct {
	StakedAmount   *big.Int
	TotalUnbonding *big.Int
	Beneficiary    common.Address
	RegisteredAt   *big.Int
	Active         bool
	NodeId         [32]byte
	Region         [32]byte
	Capabilities   uint64
	Frozen         bool
}, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "providers", arg0)

	outstruct := new(struct {
		StakedAmount   *big.Int
		TotalUnbonding *big.Int
		Beneficiary    common.Address
		RegisteredAt   *big.Int
		Active         bool
		NodeId         [32]byte
		Region         [32]byte
		Capabilities   uint64
		Frozen         bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.StakedAmount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TotalUnbonding = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Beneficiary = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.RegisteredAt = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.Active = *abi.ConvertType(out[4], new(bool)).(*bool)
	outstruct.NodeId = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Region = *abi.ConvertType(out[6], new([32]byte)).(*[32]byte)
	outstruct.Capabilities = *abi.ConvertType(out[7], new(uint64)).(*uint64)
	outstruct.Frozen = *abi.ConvertType(out[8], new(bool)).(*bool)

	return *outstruct, err

}

// Providers is a free data retrieval call binding the contract method 0x0787bc27.
//
// Solidity: function providers(address ) view returns(uint128 stakedAmount, uint128 totalUnbonding, address beneficiary, uint48 registeredAt, bool active, bytes32 nodeId, bytes32 region, uint64 capabilities, bool frozen)
func (_BunkerStaking *BunkerStakingSession) Providers(arg0 common.Address) (struct {
	StakedAmount   *big.Int
	TotalUnbonding *big.Int
	Beneficiary    common.Address
	RegisteredAt   *big.Int
	Active         bool
	NodeId         [32]byte
	Region         [32]byte
	Capabilities   uint64
	Frozen         bool
}, error) {
	return _BunkerStaking.Contract.Providers(&_BunkerStaking.CallOpts, arg0)
}

// Providers is a free data retrieval call binding the contract method 0x0787bc27.
//
// Solidity: function providers(address ) view returns(uint128 stakedAmount, uint128 totalUnbonding, address beneficiary, uint48 registeredAt, bool active, bytes32 nodeId, bytes32 region, uint64 capabilities, bool frozen)
func (_BunkerStaking *BunkerStakingCallerSession) Providers(arg0 common.Address) (struct {
	StakedAmount   *big.Int
	TotalUnbonding *big.Int
	Beneficiary    common.Address
	RegisteredAt   *big.Int
	Active         bool
	NodeId         [32]byte
	Region         [32]byte
	Capabilities   uint64
	Frozen         bool
}, error) {
	return _BunkerStaking.Contract.Providers(&_BunkerStaking.CallOpts, arg0)
}

// RewardPerToken is a free data retrieval call binding the contract method 0xcd3daf9d.
//
// Solidity: function rewardPerToken() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) RewardPerToken(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "rewardPerToken")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RewardPerToken is a free data retrieval call binding the contract method 0xcd3daf9d.
//
// Solidity: function rewardPerToken() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) RewardPerToken() (*big.Int, error) {
	return _BunkerStaking.Contract.RewardPerToken(&_BunkerStaking.CallOpts)
}

// RewardPerToken is a free data retrieval call binding the contract method 0xcd3daf9d.
//
// Solidity: function rewardPerToken() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) RewardPerToken() (*big.Int, error) {
	return _BunkerStaking.Contract.RewardPerToken(&_BunkerStaking.CallOpts)
}

// RewardPerTokenStored is a free data retrieval call binding the contract method 0xdf136d65.
//
// Solidity: function rewardPerTokenStored() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) RewardPerTokenStored(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "rewardPerTokenStored")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RewardPerTokenStored is a free data retrieval call binding the contract method 0xdf136d65.
//
// Solidity: function rewardPerTokenStored() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) RewardPerTokenStored() (*big.Int, error) {
	return _BunkerStaking.Contract.RewardPerTokenStored(&_BunkerStaking.CallOpts)
}

// RewardPerTokenStored is a free data retrieval call binding the contract method 0xdf136d65.
//
// Solidity: function rewardPerTokenStored() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) RewardPerTokenStored() (*big.Int, error) {
	return _BunkerStaking.Contract.RewardPerTokenStored(&_BunkerStaking.CallOpts)
}

// RewardRate is a free data retrieval call binding the contract method 0x7b0a47ee.
//
// Solidity: function rewardRate() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) RewardRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "rewardRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RewardRate is a free data retrieval call binding the contract method 0x7b0a47ee.
//
// Solidity: function rewardRate() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) RewardRate() (*big.Int, error) {
	return _BunkerStaking.Contract.RewardRate(&_BunkerStaking.CallOpts)
}

// RewardRate is a free data retrieval call binding the contract method 0x7b0a47ee.
//
// Solidity: function rewardRate() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) RewardRate() (*big.Int, error) {
	return _BunkerStaking.Contract.RewardRate(&_BunkerStaking.CallOpts)
}

// Rewards is a free data retrieval call binding the contract method 0x0700037d.
//
// Solidity: function rewards(address ) view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) Rewards(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "rewards", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Rewards is a free data retrieval call binding the contract method 0x0700037d.
//
// Solidity: function rewards(address ) view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) Rewards(arg0 common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.Rewards(&_BunkerStaking.CallOpts, arg0)
}

// Rewards is a free data retrieval call binding the contract method 0x0700037d.
//
// Solidity: function rewards(address ) view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) Rewards(arg0 common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.Rewards(&_BunkerStaking.CallOpts, arg0)
}

// RewardsDuration is a free data retrieval call binding the contract method 0x386a9525.
//
// Solidity: function rewardsDuration() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) RewardsDuration(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "rewardsDuration")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RewardsDuration is a free data retrieval call binding the contract method 0x386a9525.
//
// Solidity: function rewardsDuration() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) RewardsDuration() (*big.Int, error) {
	return _BunkerStaking.Contract.RewardsDuration(&_BunkerStaking.CallOpts)
}

// RewardsDuration is a free data retrieval call binding the contract method 0x386a9525.
//
// Solidity: function rewardsDuration() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) RewardsDuration() (*big.Int, error) {
	return _BunkerStaking.Contract.RewardsDuration(&_BunkerStaking.CallOpts)
}

// RewardsToken is a free data retrieval call binding the contract method 0xd1af0c7d.
//
// Solidity: function rewardsToken() view returns(address)
func (_BunkerStaking *BunkerStakingCaller) RewardsToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "rewardsToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RewardsToken is a free data retrieval call binding the contract method 0xd1af0c7d.
//
// Solidity: function rewardsToken() view returns(address)
func (_BunkerStaking *BunkerStakingSession) RewardsToken() (common.Address, error) {
	return _BunkerStaking.Contract.RewardsToken(&_BunkerStaking.CallOpts)
}

// RewardsToken is a free data retrieval call binding the contract method 0xd1af0c7d.
//
// Solidity: function rewardsToken() view returns(address)
func (_BunkerStaking *BunkerStakingCallerSession) RewardsToken() (common.Address, error) {
	return _BunkerStaking.Contract.RewardsToken(&_BunkerStaking.CallOpts)
}

// SlashBurnBps is a free data retrieval call binding the contract method 0x4a928939.
//
// Solidity: function slashBurnBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) SlashBurnBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "slashBurnBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlashBurnBps is a free data retrieval call binding the contract method 0x4a928939.
//
// Solidity: function slashBurnBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) SlashBurnBps() (*big.Int, error) {
	return _BunkerStaking.Contract.SlashBurnBps(&_BunkerStaking.CallOpts)
}

// SlashBurnBps is a free data retrieval call binding the contract method 0x4a928939.
//
// Solidity: function slashBurnBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) SlashBurnBps() (*big.Int, error) {
	return _BunkerStaking.Contract.SlashBurnBps(&_BunkerStaking.CallOpts)
}

// SlashPercentageBps is a free data retrieval call binding the contract method 0xab53fb0b.
//
// Solidity: function slashPercentageBps(uint8 ) view returns(uint16)
func (_BunkerStaking *BunkerStakingCaller) SlashPercentageBps(opts *bind.CallOpts, arg0 uint8) (uint16, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "slashPercentageBps", arg0)

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// SlashPercentageBps is a free data retrieval call binding the contract method 0xab53fb0b.
//
// Solidity: function slashPercentageBps(uint8 ) view returns(uint16)
func (_BunkerStaking *BunkerStakingSession) SlashPercentageBps(arg0 uint8) (uint16, error) {
	return _BunkerStaking.Contract.SlashPercentageBps(&_BunkerStaking.CallOpts, arg0)
}

// SlashPercentageBps is a free data retrieval call binding the contract method 0xab53fb0b.
//
// Solidity: function slashPercentageBps(uint8 ) view returns(uint16)
func (_BunkerStaking *BunkerStakingCallerSession) SlashPercentageBps(arg0 uint8) (uint16, error) {
	return _BunkerStaking.Contract.SlashPercentageBps(&_BunkerStaking.CallOpts, arg0)
}

// SlashProposalCount is a free data retrieval call binding the contract method 0x95bcbe6f.
//
// Solidity: function slashProposalCount() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) SlashProposalCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "slashProposalCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlashProposalCount is a free data retrieval call binding the contract method 0x95bcbe6f.
//
// Solidity: function slashProposalCount() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) SlashProposalCount() (*big.Int, error) {
	return _BunkerStaking.Contract.SlashProposalCount(&_BunkerStaking.CallOpts)
}

// SlashProposalCount is a free data retrieval call binding the contract method 0x95bcbe6f.
//
// Solidity: function slashProposalCount() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) SlashProposalCount() (*big.Int, error) {
	return _BunkerStaking.Contract.SlashProposalCount(&_BunkerStaking.CallOpts)
}

// SlashProposals is a free data retrieval call binding the contract method 0xe7666dba.
//
// Solidity: function slashProposals(uint256 ) view returns(address provider, uint256 amount, string reason, uint256 proposedAt, bool executed, bool appealed, bool resolved, uint8 slashReason, uint256 appealWindowSnapshot, uint16 slashBurnBpsSnapshot)
func (_BunkerStaking *BunkerStakingCaller) SlashProposals(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Provider             common.Address
	Amount               *big.Int
	Reason               string
	ProposedAt           *big.Int
	Executed             bool
	Appealed             bool
	Resolved             bool
	SlashReason          uint8
	AppealWindowSnapshot *big.Int
	SlashBurnBpsSnapshot uint16
}, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "slashProposals", arg0)

	outstruct := new(struct {
		Provider             common.Address
		Amount               *big.Int
		Reason               string
		ProposedAt           *big.Int
		Executed             bool
		Appealed             bool
		Resolved             bool
		SlashReason          uint8
		AppealWindowSnapshot *big.Int
		SlashBurnBpsSnapshot uint16
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Provider = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Amount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Reason = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ProposedAt = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.Executed = *abi.ConvertType(out[4], new(bool)).(*bool)
	outstruct.Appealed = *abi.ConvertType(out[5], new(bool)).(*bool)
	outstruct.Resolved = *abi.ConvertType(out[6], new(bool)).(*bool)
	outstruct.SlashReason = *abi.ConvertType(out[7], new(uint8)).(*uint8)
	outstruct.AppealWindowSnapshot = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)
	outstruct.SlashBurnBpsSnapshot = *abi.ConvertType(out[9], new(uint16)).(*uint16)

	return *outstruct, err

}

// SlashProposals is a free data retrieval call binding the contract method 0xe7666dba.
//
// Solidity: function slashProposals(uint256 ) view returns(address provider, uint256 amount, string reason, uint256 proposedAt, bool executed, bool appealed, bool resolved, uint8 slashReason, uint256 appealWindowSnapshot, uint16 slashBurnBpsSnapshot)
func (_BunkerStaking *BunkerStakingSession) SlashProposals(arg0 *big.Int) (struct {
	Provider             common.Address
	Amount               *big.Int
	Reason               string
	ProposedAt           *big.Int
	Executed             bool
	Appealed             bool
	Resolved             bool
	SlashReason          uint8
	AppealWindowSnapshot *big.Int
	SlashBurnBpsSnapshot uint16
}, error) {
	return _BunkerStaking.Contract.SlashProposals(&_BunkerStaking.CallOpts, arg0)
}

// SlashProposals is a free data retrieval call binding the contract method 0xe7666dba.
//
// Solidity: function slashProposals(uint256 ) view returns(address provider, uint256 amount, string reason, uint256 proposedAt, bool executed, bool appealed, bool resolved, uint8 slashReason, uint256 appealWindowSnapshot, uint16 slashBurnBpsSnapshot)
func (_BunkerStaking *BunkerStakingCallerSession) SlashProposals(arg0 *big.Int) (struct {
	Provider             common.Address
	Amount               *big.Int
	Reason               string
	ProposedAt           *big.Int
	Executed             bool
	Appealed             bool
	Resolved             bool
	SlashReason          uint8
	AppealWindowSnapshot *big.Int
	SlashBurnBpsSnapshot uint16
}, error) {
	return _BunkerStaking.Contract.SlashProposals(&_BunkerStaking.CallOpts, arg0)
}

// SlashTreasuryBps is a free data retrieval call binding the contract method 0x246c7e09.
//
// Solidity: function slashTreasuryBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) SlashTreasuryBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "slashTreasuryBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlashTreasuryBps is a free data retrieval call binding the contract method 0x246c7e09.
//
// Solidity: function slashTreasuryBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) SlashTreasuryBps() (*big.Int, error) {
	return _BunkerStaking.Contract.SlashTreasuryBps(&_BunkerStaking.CallOpts)
}

// SlashTreasuryBps is a free data retrieval call binding the contract method 0x246c7e09.
//
// Solidity: function slashTreasuryBps() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) SlashTreasuryBps() (*big.Int, error) {
	return _BunkerStaking.Contract.SlashTreasuryBps(&_BunkerStaking.CallOpts)
}

// SlashingEnabled is a free data retrieval call binding the contract method 0x321ba6fd.
//
// Solidity: function slashingEnabled() view returns(bool)
func (_BunkerStaking *BunkerStakingCaller) SlashingEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "slashingEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SlashingEnabled is a free data retrieval call binding the contract method 0x321ba6fd.
//
// Solidity: function slashingEnabled() view returns(bool)
func (_BunkerStaking *BunkerStakingSession) SlashingEnabled() (bool, error) {
	return _BunkerStaking.Contract.SlashingEnabled(&_BunkerStaking.CallOpts)
}

// SlashingEnabled is a free data retrieval call binding the contract method 0x321ba6fd.
//
// Solidity: function slashingEnabled() view returns(bool)
func (_BunkerStaking *BunkerStakingCallerSession) SlashingEnabled() (bool, error) {
	return _BunkerStaking.Contract.SlashingEnabled(&_BunkerStaking.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerStaking *BunkerStakingCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerStaking *BunkerStakingSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _BunkerStaking.Contract.SupportsInterface(&_BunkerStaking.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_BunkerStaking *BunkerStakingCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _BunkerStaking.Contract.SupportsInterface(&_BunkerStaking.CallOpts, interfaceId)
}

// TierConfigs is a free data retrieval call binding the contract method 0x9da9db2a.
//
// Solidity: function tierConfigs(uint8 ) view returns(uint256 minStake, uint16 maxConcurrentJobs, uint16 rewardMultiplierBps, bool priorityQueue, bool governance)
func (_BunkerStaking *BunkerStakingCaller) TierConfigs(opts *bind.CallOpts, arg0 uint8) (struct {
	MinStake            *big.Int
	MaxConcurrentJobs   uint16
	RewardMultiplierBps uint16
	PriorityQueue       bool
	Governance          bool
}, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "tierConfigs", arg0)

	outstruct := new(struct {
		MinStake            *big.Int
		MaxConcurrentJobs   uint16
		RewardMultiplierBps uint16
		PriorityQueue       bool
		Governance          bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.MinStake = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.MaxConcurrentJobs = *abi.ConvertType(out[1], new(uint16)).(*uint16)
	outstruct.RewardMultiplierBps = *abi.ConvertType(out[2], new(uint16)).(*uint16)
	outstruct.PriorityQueue = *abi.ConvertType(out[3], new(bool)).(*bool)
	outstruct.Governance = *abi.ConvertType(out[4], new(bool)).(*bool)

	return *outstruct, err

}

// TierConfigs is a free data retrieval call binding the contract method 0x9da9db2a.
//
// Solidity: function tierConfigs(uint8 ) view returns(uint256 minStake, uint16 maxConcurrentJobs, uint16 rewardMultiplierBps, bool priorityQueue, bool governance)
func (_BunkerStaking *BunkerStakingSession) TierConfigs(arg0 uint8) (struct {
	MinStake            *big.Int
	MaxConcurrentJobs   uint16
	RewardMultiplierBps uint16
	PriorityQueue       bool
	Governance          bool
}, error) {
	return _BunkerStaking.Contract.TierConfigs(&_BunkerStaking.CallOpts, arg0)
}

// TierConfigs is a free data retrieval call binding the contract method 0x9da9db2a.
//
// Solidity: function tierConfigs(uint8 ) view returns(uint256 minStake, uint16 maxConcurrentJobs, uint16 rewardMultiplierBps, bool priorityQueue, bool governance)
func (_BunkerStaking *BunkerStakingCallerSession) TierConfigs(arg0 uint8) (struct {
	MinStake            *big.Int
	MaxConcurrentJobs   uint16
	RewardMultiplierBps uint16
	PriorityQueue       bool
	Governance          bool
}, error) {
	return _BunkerStaking.Contract.TierConfigs(&_BunkerStaking.CallOpts, arg0)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_BunkerStaking *BunkerStakingCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_BunkerStaking *BunkerStakingSession) Token() (common.Address, error) {
	return _BunkerStaking.Contract.Token(&_BunkerStaking.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_BunkerStaking *BunkerStakingCallerSession) Token() (common.Address, error) {
	return _BunkerStaking.Contract.Token(&_BunkerStaking.CallOpts)
}

// TotalComputeHoursReported is a free data retrieval call binding the contract method 0xd424b7c3.
//
// Solidity: function totalComputeHoursReported() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) TotalComputeHoursReported(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "totalComputeHoursReported")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalComputeHoursReported is a free data retrieval call binding the contract method 0xd424b7c3.
//
// Solidity: function totalComputeHoursReported() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) TotalComputeHoursReported() (*big.Int, error) {
	return _BunkerStaking.Contract.TotalComputeHoursReported(&_BunkerStaking.CallOpts)
}

// TotalComputeHoursReported is a free data retrieval call binding the contract method 0xd424b7c3.
//
// Solidity: function totalComputeHoursReported() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) TotalComputeHoursReported() (*big.Int, error) {
	return _BunkerStaking.Contract.TotalComputeHoursReported(&_BunkerStaking.CallOpts)
}

// TotalStaked is a free data retrieval call binding the contract method 0x817b1cd2.
//
// Solidity: function totalStaked() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) TotalStaked(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "totalStaked")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalStaked is a free data retrieval call binding the contract method 0x817b1cd2.
//
// Solidity: function totalStaked() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) TotalStaked() (*big.Int, error) {
	return _BunkerStaking.Contract.TotalStaked(&_BunkerStaking.CallOpts)
}

// TotalStaked is a free data retrieval call binding the contract method 0x817b1cd2.
//
// Solidity: function totalStaked() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) TotalStaked() (*big.Int, error) {
	return _BunkerStaking.Contract.TotalStaked(&_BunkerStaking.CallOpts)
}

// TotalUnbonding is a free data retrieval call binding the contract method 0x350fd0be.
//
// Solidity: function totalUnbonding() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) TotalUnbonding(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "totalUnbonding")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalUnbonding is a free data retrieval call binding the contract method 0x350fd0be.
//
// Solidity: function totalUnbonding() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) TotalUnbonding() (*big.Int, error) {
	return _BunkerStaking.Contract.TotalUnbonding(&_BunkerStaking.CallOpts)
}

// TotalUnbonding is a free data retrieval call binding the contract method 0x350fd0be.
//
// Solidity: function totalUnbonding() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) TotalUnbonding() (*big.Int, error) {
	return _BunkerStaking.Contract.TotalUnbonding(&_BunkerStaking.CallOpts)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_BunkerStaking *BunkerStakingCaller) Treasury(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "treasury")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_BunkerStaking *BunkerStakingSession) Treasury() (common.Address, error) {
	return _BunkerStaking.Contract.Treasury(&_BunkerStaking.CallOpts)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_BunkerStaking *BunkerStakingCallerSession) Treasury() (common.Address, error) {
	return _BunkerStaking.Contract.Treasury(&_BunkerStaking.CallOpts)
}

// UnbondingPeriod is a free data retrieval call binding the contract method 0x6cf6d675.
//
// Solidity: function unbondingPeriod() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) UnbondingPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "unbondingPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UnbondingPeriod is a free data retrieval call binding the contract method 0x6cf6d675.
//
// Solidity: function unbondingPeriod() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) UnbondingPeriod() (*big.Int, error) {
	return _BunkerStaking.Contract.UnbondingPeriod(&_BunkerStaking.CallOpts)
}

// UnbondingPeriod is a free data retrieval call binding the contract method 0x6cf6d675.
//
// Solidity: function unbondingPeriod() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) UnbondingPeriod() (*big.Int, error) {
	return _BunkerStaking.Contract.UnbondingPeriod(&_BunkerStaking.CallOpts)
}

// UnstakeQueues is a free data retrieval call binding the contract method 0xf0023200.
//
// Solidity: function unstakeQueues(address , uint256 ) view returns(uint128 amount, uint48 unlockTime, bool completed)
func (_BunkerStaking *BunkerStakingCaller) UnstakeQueues(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	Amount     *big.Int
	UnlockTime *big.Int
	Completed  bool
}, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "unstakeQueues", arg0, arg1)

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

// UnstakeQueues is a free data retrieval call binding the contract method 0xf0023200.
//
// Solidity: function unstakeQueues(address , uint256 ) view returns(uint128 amount, uint48 unlockTime, bool completed)
func (_BunkerStaking *BunkerStakingSession) UnstakeQueues(arg0 common.Address, arg1 *big.Int) (struct {
	Amount     *big.Int
	UnlockTime *big.Int
	Completed  bool
}, error) {
	return _BunkerStaking.Contract.UnstakeQueues(&_BunkerStaking.CallOpts, arg0, arg1)
}

// UnstakeQueues is a free data retrieval call binding the contract method 0xf0023200.
//
// Solidity: function unstakeQueues(address , uint256 ) view returns(uint128 amount, uint48 unlockTime, bool completed)
func (_BunkerStaking *BunkerStakingCallerSession) UnstakeQueues(arg0 common.Address, arg1 *big.Int) (struct {
	Amount     *big.Int
	UnlockTime *big.Int
	Completed  bool
}, error) {
	return _BunkerStaking.Contract.UnstakeQueues(&_BunkerStaking.CallOpts, arg0, arg1)
}

// UserRewardPerTokenPaid is a free data retrieval call binding the contract method 0x8b876347.
//
// Solidity: function userRewardPerTokenPaid(address ) view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) UserRewardPerTokenPaid(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "userRewardPerTokenPaid", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UserRewardPerTokenPaid is a free data retrieval call binding the contract method 0x8b876347.
//
// Solidity: function userRewardPerTokenPaid(address ) view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) UserRewardPerTokenPaid(arg0 common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.UserRewardPerTokenPaid(&_BunkerStaking.CallOpts, arg0)
}

// UserRewardPerTokenPaid is a free data retrieval call binding the contract method 0x8b876347.
//
// Solidity: function userRewardPerTokenPaid(address ) view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) UserRewardPerTokenPaid(arg0 common.Address) (*big.Int, error) {
	return _BunkerStaking.Contract.UserRewardPerTokenPaid(&_BunkerStaking.CallOpts, arg0)
}

// VestedRewards is a free data retrieval call binding the contract method 0xfcab880f.
//
// Solidity: function vestedRewards(address , uint256 ) view returns(uint128 totalAmount, uint128 releasedAmount, uint48 vestingStart)
func (_BunkerStaking *BunkerStakingCaller) VestedRewards(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	TotalAmount    *big.Int
	ReleasedAmount *big.Int
	VestingStart   *big.Int
}, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "vestedRewards", arg0, arg1)

	outstruct := new(struct {
		TotalAmount    *big.Int
		ReleasedAmount *big.Int
		VestingStart   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.TotalAmount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.ReleasedAmount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.VestingStart = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// VestedRewards is a free data retrieval call binding the contract method 0xfcab880f.
//
// Solidity: function vestedRewards(address , uint256 ) view returns(uint128 totalAmount, uint128 releasedAmount, uint48 vestingStart)
func (_BunkerStaking *BunkerStakingSession) VestedRewards(arg0 common.Address, arg1 *big.Int) (struct {
	TotalAmount    *big.Int
	ReleasedAmount *big.Int
	VestingStart   *big.Int
}, error) {
	return _BunkerStaking.Contract.VestedRewards(&_BunkerStaking.CallOpts, arg0, arg1)
}

// VestedRewards is a free data retrieval call binding the contract method 0xfcab880f.
//
// Solidity: function vestedRewards(address , uint256 ) view returns(uint128 totalAmount, uint128 releasedAmount, uint48 vestingStart)
func (_BunkerStaking *BunkerStakingCallerSession) VestedRewards(arg0 common.Address, arg1 *big.Int) (struct {
	TotalAmount    *big.Int
	ReleasedAmount *big.Int
	VestingStart   *big.Int
}, error) {
	return _BunkerStaking.Contract.VestedRewards(&_BunkerStaking.CallOpts, arg0, arg1)
}

// VestingPeriod is a free data retrieval call binding the contract method 0x7313ee5a.
//
// Solidity: function vestingPeriod() view returns(uint256)
func (_BunkerStaking *BunkerStakingCaller) VestingPeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BunkerStaking.contract.Call(opts, &out, "vestingPeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VestingPeriod is a free data retrieval call binding the contract method 0x7313ee5a.
//
// Solidity: function vestingPeriod() view returns(uint256)
func (_BunkerStaking *BunkerStakingSession) VestingPeriod() (*big.Int, error) {
	return _BunkerStaking.Contract.VestingPeriod(&_BunkerStaking.CallOpts)
}

// VestingPeriod is a free data retrieval call binding the contract method 0x7313ee5a.
//
// Solidity: function vestingPeriod() view returns(uint256)
func (_BunkerStaking *BunkerStakingCallerSession) VestingPeriod() (*big.Int, error) {
	return _BunkerStaking.Contract.VestingPeriod(&_BunkerStaking.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerStaking *BunkerStakingTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerStaking *BunkerStakingSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerStaking.Contract.AcceptOwnership(&_BunkerStaking.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_BunkerStaking *BunkerStakingTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _BunkerStaking.Contract.AcceptOwnership(&_BunkerStaking.TransactOpts)
}

// AppealSlash is a paid mutator transaction binding the contract method 0x949a218d.
//
// Solidity: function appealSlash(uint256 proposalId) returns()
func (_BunkerStaking *BunkerStakingTransactor) AppealSlash(opts *bind.TransactOpts, proposalId *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "appealSlash", proposalId)
}

// AppealSlash is a paid mutator transaction binding the contract method 0x949a218d.
//
// Solidity: function appealSlash(uint256 proposalId) returns()
func (_BunkerStaking *BunkerStakingSession) AppealSlash(proposalId *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.AppealSlash(&_BunkerStaking.TransactOpts, proposalId)
}

// AppealSlash is a paid mutator transaction binding the contract method 0x949a218d.
//
// Solidity: function appealSlash(uint256 proposalId) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) AppealSlash(proposalId *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.AppealSlash(&_BunkerStaking.TransactOpts, proposalId)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0x372500ab.
//
// Solidity: function claimRewards() returns()
func (_BunkerStaking *BunkerStakingTransactor) ClaimRewards(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "claimRewards")
}

// ClaimRewards is a paid mutator transaction binding the contract method 0x372500ab.
//
// Solidity: function claimRewards() returns()
func (_BunkerStaking *BunkerStakingSession) ClaimRewards() (*types.Transaction, error) {
	return _BunkerStaking.Contract.ClaimRewards(&_BunkerStaking.TransactOpts)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0x372500ab.
//
// Solidity: function claimRewards() returns()
func (_BunkerStaking *BunkerStakingTransactorSession) ClaimRewards() (*types.Transaction, error) {
	return _BunkerStaking.Contract.ClaimRewards(&_BunkerStaking.TransactOpts)
}

// ClaimVestedRewards is a paid mutator transaction binding the contract method 0xadb50861.
//
// Solidity: function claimVestedRewards() returns()
func (_BunkerStaking *BunkerStakingTransactor) ClaimVestedRewards(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "claimVestedRewards")
}

// ClaimVestedRewards is a paid mutator transaction binding the contract method 0xadb50861.
//
// Solidity: function claimVestedRewards() returns()
func (_BunkerStaking *BunkerStakingSession) ClaimVestedRewards() (*types.Transaction, error) {
	return _BunkerStaking.Contract.ClaimVestedRewards(&_BunkerStaking.TransactOpts)
}

// ClaimVestedRewards is a paid mutator transaction binding the contract method 0xadb50861.
//
// Solidity: function claimVestedRewards() returns()
func (_BunkerStaking *BunkerStakingTransactorSession) ClaimVestedRewards() (*types.Transaction, error) {
	return _BunkerStaking.Contract.ClaimVestedRewards(&_BunkerStaking.TransactOpts)
}

// CompleteUnstake is a paid mutator transaction binding the contract method 0x552b54ec.
//
// Solidity: function completeUnstake(uint256 requestIndex) returns()
func (_BunkerStaking *BunkerStakingTransactor) CompleteUnstake(opts *bind.TransactOpts, requestIndex *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "completeUnstake", requestIndex)
}

// CompleteUnstake is a paid mutator transaction binding the contract method 0x552b54ec.
//
// Solidity: function completeUnstake(uint256 requestIndex) returns()
func (_BunkerStaking *BunkerStakingSession) CompleteUnstake(requestIndex *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.CompleteUnstake(&_BunkerStaking.TransactOpts, requestIndex)
}

// CompleteUnstake is a paid mutator transaction binding the contract method 0x552b54ec.
//
// Solidity: function completeUnstake(uint256 requestIndex) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) CompleteUnstake(requestIndex *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.CompleteUnstake(&_BunkerStaking.TransactOpts, requestIndex)
}

// ExecuteBeneficiaryChange is a paid mutator transaction binding the contract method 0x78cc30ee.
//
// Solidity: function executeBeneficiaryChange() returns()
func (_BunkerStaking *BunkerStakingTransactor) ExecuteBeneficiaryChange(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "executeBeneficiaryChange")
}

// ExecuteBeneficiaryChange is a paid mutator transaction binding the contract method 0x78cc30ee.
//
// Solidity: function executeBeneficiaryChange() returns()
func (_BunkerStaking *BunkerStakingSession) ExecuteBeneficiaryChange() (*types.Transaction, error) {
	return _BunkerStaking.Contract.ExecuteBeneficiaryChange(&_BunkerStaking.TransactOpts)
}

// ExecuteBeneficiaryChange is a paid mutator transaction binding the contract method 0x78cc30ee.
//
// Solidity: function executeBeneficiaryChange() returns()
func (_BunkerStaking *BunkerStakingTransactorSession) ExecuteBeneficiaryChange() (*types.Transaction, error) {
	return _BunkerStaking.Contract.ExecuteBeneficiaryChange(&_BunkerStaking.TransactOpts)
}

// ExecuteSlash is a paid mutator transaction binding the contract method 0x69ae9af8.
//
// Solidity: function executeSlash(uint256 proposalId) returns()
func (_BunkerStaking *BunkerStakingTransactor) ExecuteSlash(opts *bind.TransactOpts, proposalId *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "executeSlash", proposalId)
}

// ExecuteSlash is a paid mutator transaction binding the contract method 0x69ae9af8.
//
// Solidity: function executeSlash(uint256 proposalId) returns()
func (_BunkerStaking *BunkerStakingSession) ExecuteSlash(proposalId *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.ExecuteSlash(&_BunkerStaking.TransactOpts, proposalId)
}

// ExecuteSlash is a paid mutator transaction binding the contract method 0x69ae9af8.
//
// Solidity: function executeSlash(uint256 proposalId) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) ExecuteSlash(proposalId *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.ExecuteSlash(&_BunkerStaking.TransactOpts, proposalId)
}

// FreezeProvider is a paid mutator transaction binding the contract method 0x9980a9e8.
//
// Solidity: function freezeProvider(address provider) returns()
func (_BunkerStaking *BunkerStakingTransactor) FreezeProvider(opts *bind.TransactOpts, provider common.Address) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "freezeProvider", provider)
}

// FreezeProvider is a paid mutator transaction binding the contract method 0x9980a9e8.
//
// Solidity: function freezeProvider(address provider) returns()
func (_BunkerStaking *BunkerStakingSession) FreezeProvider(provider common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.FreezeProvider(&_BunkerStaking.TransactOpts, provider)
}

// FreezeProvider is a paid mutator transaction binding the contract method 0x9980a9e8.
//
// Solidity: function freezeProvider(address provider) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) FreezeProvider(provider common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.FreezeProvider(&_BunkerStaking.TransactOpts, provider)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerStaking *BunkerStakingTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerStaking *BunkerStakingSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.GrantRole(&_BunkerStaking.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.GrantRole(&_BunkerStaking.TransactOpts, role, account)
}

// InitiateBeneficiaryChange is a paid mutator transaction binding the contract method 0x32a10dfd.
//
// Solidity: function initiateBeneficiaryChange(address newBeneficiary) returns()
func (_BunkerStaking *BunkerStakingTransactor) InitiateBeneficiaryChange(opts *bind.TransactOpts, newBeneficiary common.Address) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "initiateBeneficiaryChange", newBeneficiary)
}

// InitiateBeneficiaryChange is a paid mutator transaction binding the contract method 0x32a10dfd.
//
// Solidity: function initiateBeneficiaryChange(address newBeneficiary) returns()
func (_BunkerStaking *BunkerStakingSession) InitiateBeneficiaryChange(newBeneficiary common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.InitiateBeneficiaryChange(&_BunkerStaking.TransactOpts, newBeneficiary)
}

// InitiateBeneficiaryChange is a paid mutator transaction binding the contract method 0x32a10dfd.
//
// Solidity: function initiateBeneficiaryChange(address newBeneficiary) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) InitiateBeneficiaryChange(newBeneficiary common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.InitiateBeneficiaryChange(&_BunkerStaking.TransactOpts, newBeneficiary)
}

// NotifyRewardAmount is a paid mutator transaction binding the contract method 0x3c6b16ab.
//
// Solidity: function notifyRewardAmount(uint256 reward) returns()
func (_BunkerStaking *BunkerStakingTransactor) NotifyRewardAmount(opts *bind.TransactOpts, reward *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "notifyRewardAmount", reward)
}

// NotifyRewardAmount is a paid mutator transaction binding the contract method 0x3c6b16ab.
//
// Solidity: function notifyRewardAmount(uint256 reward) returns()
func (_BunkerStaking *BunkerStakingSession) NotifyRewardAmount(reward *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.NotifyRewardAmount(&_BunkerStaking.TransactOpts, reward)
}

// NotifyRewardAmount is a paid mutator transaction binding the contract method 0x3c6b16ab.
//
// Solidity: function notifyRewardAmount(uint256 reward) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) NotifyRewardAmount(reward *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.NotifyRewardAmount(&_BunkerStaking.TransactOpts, reward)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerStaking *BunkerStakingTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerStaking *BunkerStakingSession) Pause() (*types.Transaction, error) {
	return _BunkerStaking.Contract.Pause(&_BunkerStaking.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_BunkerStaking *BunkerStakingTransactorSession) Pause() (*types.Transaction, error) {
	return _BunkerStaking.Contract.Pause(&_BunkerStaking.TransactOpts)
}

// ProposeSlash is a paid mutator transaction binding the contract method 0xa7c23fec.
//
// Solidity: function proposeSlash(address provider, uint256 amount, string reason) returns(uint256 proposalId)
func (_BunkerStaking *BunkerStakingTransactor) ProposeSlash(opts *bind.TransactOpts, provider common.Address, amount *big.Int, reason string) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "proposeSlash", provider, amount, reason)
}

// ProposeSlash is a paid mutator transaction binding the contract method 0xa7c23fec.
//
// Solidity: function proposeSlash(address provider, uint256 amount, string reason) returns(uint256 proposalId)
func (_BunkerStaking *BunkerStakingSession) ProposeSlash(provider common.Address, amount *big.Int, reason string) (*types.Transaction, error) {
	return _BunkerStaking.Contract.ProposeSlash(&_BunkerStaking.TransactOpts, provider, amount, reason)
}

// ProposeSlash is a paid mutator transaction binding the contract method 0xa7c23fec.
//
// Solidity: function proposeSlash(address provider, uint256 amount, string reason) returns(uint256 proposalId)
func (_BunkerStaking *BunkerStakingTransactorSession) ProposeSlash(provider common.Address, amount *big.Int, reason string) (*types.Transaction, error) {
	return _BunkerStaking.Contract.ProposeSlash(&_BunkerStaking.TransactOpts, provider, amount, reason)
}

// ProposeSlashByReason is a paid mutator transaction binding the contract method 0x340178f2.
//
// Solidity: function proposeSlashByReason(address provider, uint8 reason) returns(uint256 proposalId)
func (_BunkerStaking *BunkerStakingTransactor) ProposeSlashByReason(opts *bind.TransactOpts, provider common.Address, reason uint8) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "proposeSlashByReason", provider, reason)
}

// ProposeSlashByReason is a paid mutator transaction binding the contract method 0x340178f2.
//
// Solidity: function proposeSlashByReason(address provider, uint8 reason) returns(uint256 proposalId)
func (_BunkerStaking *BunkerStakingSession) ProposeSlashByReason(provider common.Address, reason uint8) (*types.Transaction, error) {
	return _BunkerStaking.Contract.ProposeSlashByReason(&_BunkerStaking.TransactOpts, provider, reason)
}

// ProposeSlashByReason is a paid mutator transaction binding the contract method 0x340178f2.
//
// Solidity: function proposeSlashByReason(address provider, uint8 reason) returns(uint256 proposalId)
func (_BunkerStaking *BunkerStakingTransactorSession) ProposeSlashByReason(provider common.Address, reason uint8) (*types.Transaction, error) {
	return _BunkerStaking.Contract.ProposeSlashByReason(&_BunkerStaking.TransactOpts, provider, reason)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerStaking *BunkerStakingTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerStaking *BunkerStakingSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerStaking.Contract.RenounceOwnership(&_BunkerStaking.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BunkerStaking *BunkerStakingTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BunkerStaking.Contract.RenounceOwnership(&_BunkerStaking.TransactOpts)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerStaking *BunkerStakingTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerStaking *BunkerStakingSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.RenounceRole(&_BunkerStaking.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.RenounceRole(&_BunkerStaking.TransactOpts, role, callerConfirmation)
}

// ReportComputeHours is a paid mutator transaction binding the contract method 0xd85bc13f.
//
// Solidity: function reportComputeHours(uint256 hours_) returns()
func (_BunkerStaking *BunkerStakingTransactor) ReportComputeHours(opts *bind.TransactOpts, hours_ *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "reportComputeHours", hours_)
}

// ReportComputeHours is a paid mutator transaction binding the contract method 0xd85bc13f.
//
// Solidity: function reportComputeHours(uint256 hours_) returns()
func (_BunkerStaking *BunkerStakingSession) ReportComputeHours(hours_ *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.ReportComputeHours(&_BunkerStaking.TransactOpts, hours_)
}

// ReportComputeHours is a paid mutator transaction binding the contract method 0xd85bc13f.
//
// Solidity: function reportComputeHours(uint256 hours_) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) ReportComputeHours(hours_ *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.ReportComputeHours(&_BunkerStaking.TransactOpts, hours_)
}

// RequestUnstake is a paid mutator transaction binding the contract method 0x23095721.
//
// Solidity: function requestUnstake(uint256 amount) returns()
func (_BunkerStaking *BunkerStakingTransactor) RequestUnstake(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "requestUnstake", amount)
}

// RequestUnstake is a paid mutator transaction binding the contract method 0x23095721.
//
// Solidity: function requestUnstake(uint256 amount) returns()
func (_BunkerStaking *BunkerStakingSession) RequestUnstake(amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.RequestUnstake(&_BunkerStaking.TransactOpts, amount)
}

// RequestUnstake is a paid mutator transaction binding the contract method 0x23095721.
//
// Solidity: function requestUnstake(uint256 amount) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) RequestUnstake(amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.RequestUnstake(&_BunkerStaking.TransactOpts, amount)
}

// ResolveAppeal is a paid mutator transaction binding the contract method 0x1c74cf1d.
//
// Solidity: function resolveAppeal(uint256 proposalId, bool uphold) returns()
func (_BunkerStaking *BunkerStakingTransactor) ResolveAppeal(opts *bind.TransactOpts, proposalId *big.Int, uphold bool) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "resolveAppeal", proposalId, uphold)
}

// ResolveAppeal is a paid mutator transaction binding the contract method 0x1c74cf1d.
//
// Solidity: function resolveAppeal(uint256 proposalId, bool uphold) returns()
func (_BunkerStaking *BunkerStakingSession) ResolveAppeal(proposalId *big.Int, uphold bool) (*types.Transaction, error) {
	return _BunkerStaking.Contract.ResolveAppeal(&_BunkerStaking.TransactOpts, proposalId, uphold)
}

// ResolveAppeal is a paid mutator transaction binding the contract method 0x1c74cf1d.
//
// Solidity: function resolveAppeal(uint256 proposalId, bool uphold) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) ResolveAppeal(proposalId *big.Int, uphold bool) (*types.Transaction, error) {
	return _BunkerStaking.Contract.ResolveAppeal(&_BunkerStaking.TransactOpts, proposalId, uphold)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerStaking *BunkerStakingTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerStaking *BunkerStakingSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.RevokeRole(&_BunkerStaking.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.RevokeRole(&_BunkerStaking.TransactOpts, role, account)
}

// SetAppealWindow is a paid mutator transaction binding the contract method 0x031fc067.
//
// Solidity: function setAppealWindow(uint256 newWindow) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetAppealWindow(opts *bind.TransactOpts, newWindow *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setAppealWindow", newWindow)
}

// SetAppealWindow is a paid mutator transaction binding the contract method 0x031fc067.
//
// Solidity: function setAppealWindow(uint256 newWindow) returns()
func (_BunkerStaking *BunkerStakingSession) SetAppealWindow(newWindow *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetAppealWindow(&_BunkerStaking.TransactOpts, newWindow)
}

// SetAppealWindow is a paid mutator transaction binding the contract method 0x031fc067.
//
// Solidity: function setAppealWindow(uint256 newWindow) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetAppealWindow(newWindow *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetAppealWindow(&_BunkerStaking.TransactOpts, newWindow)
}

// SetEmissionMultiplier is a paid mutator transaction binding the contract method 0x4817cfaa.
//
// Solidity: function setEmissionMultiplier(uint256 multiplierBps) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetEmissionMultiplier(opts *bind.TransactOpts, multiplierBps *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setEmissionMultiplier", multiplierBps)
}

// SetEmissionMultiplier is a paid mutator transaction binding the contract method 0x4817cfaa.
//
// Solidity: function setEmissionMultiplier(uint256 multiplierBps) returns()
func (_BunkerStaking *BunkerStakingSession) SetEmissionMultiplier(multiplierBps *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetEmissionMultiplier(&_BunkerStaking.TransactOpts, multiplierBps)
}

// SetEmissionMultiplier is a paid mutator transaction binding the contract method 0x4817cfaa.
//
// Solidity: function setEmissionMultiplier(uint256 multiplierBps) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetEmissionMultiplier(multiplierBps *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetEmissionMultiplier(&_BunkerStaking.TransactOpts, multiplierBps)
}

// SetMaxEmissionRate is a paid mutator transaction binding the contract method 0x8c3aeb08.
//
// Solidity: function setMaxEmissionRate(uint256 maxRate) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetMaxEmissionRate(opts *bind.TransactOpts, maxRate *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setMaxEmissionRate", maxRate)
}

// SetMaxEmissionRate is a paid mutator transaction binding the contract method 0x8c3aeb08.
//
// Solidity: function setMaxEmissionRate(uint256 maxRate) returns()
func (_BunkerStaking *BunkerStakingSession) SetMaxEmissionRate(maxRate *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetMaxEmissionRate(&_BunkerStaking.TransactOpts, maxRate)
}

// SetMaxEmissionRate is a paid mutator transaction binding the contract method 0x8c3aeb08.
//
// Solidity: function setMaxEmissionRate(uint256 maxRate) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetMaxEmissionRate(maxRate *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetMaxEmissionRate(&_BunkerStaking.TransactOpts, maxRate)
}

// SetMaxTierMultiplierBps is a paid mutator transaction binding the contract method 0x0b69caff.
//
// Solidity: function setMaxTierMultiplierBps(uint16 newMax) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetMaxTierMultiplierBps(opts *bind.TransactOpts, newMax uint16) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setMaxTierMultiplierBps", newMax)
}

// SetMaxTierMultiplierBps is a paid mutator transaction binding the contract method 0x0b69caff.
//
// Solidity: function setMaxTierMultiplierBps(uint16 newMax) returns()
func (_BunkerStaking *BunkerStakingSession) SetMaxTierMultiplierBps(newMax uint16) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetMaxTierMultiplierBps(&_BunkerStaking.TransactOpts, newMax)
}

// SetMaxTierMultiplierBps is a paid mutator transaction binding the contract method 0x0b69caff.
//
// Solidity: function setMaxTierMultiplierBps(uint16 newMax) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetMaxTierMultiplierBps(newMax uint16) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetMaxTierMultiplierBps(&_BunkerStaking.TransactOpts, newMax)
}

// SetRewardsDuration is a paid mutator transaction binding the contract method 0xcc1a378f.
//
// Solidity: function setRewardsDuration(uint256 _rewardsDuration) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetRewardsDuration(opts *bind.TransactOpts, _rewardsDuration *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setRewardsDuration", _rewardsDuration)
}

// SetRewardsDuration is a paid mutator transaction binding the contract method 0xcc1a378f.
//
// Solidity: function setRewardsDuration(uint256 _rewardsDuration) returns()
func (_BunkerStaking *BunkerStakingSession) SetRewardsDuration(_rewardsDuration *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetRewardsDuration(&_BunkerStaking.TransactOpts, _rewardsDuration)
}

// SetRewardsDuration is a paid mutator transaction binding the contract method 0xcc1a378f.
//
// Solidity: function setRewardsDuration(uint256 _rewardsDuration) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetRewardsDuration(_rewardsDuration *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetRewardsDuration(&_BunkerStaking.TransactOpts, _rewardsDuration)
}

// SetSlashFeeSplit is a paid mutator transaction binding the contract method 0xe6a4347d.
//
// Solidity: function setSlashFeeSplit(uint16 burnBps, uint16 treasuryBps) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetSlashFeeSplit(opts *bind.TransactOpts, burnBps uint16, treasuryBps uint16) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setSlashFeeSplit", burnBps, treasuryBps)
}

// SetSlashFeeSplit is a paid mutator transaction binding the contract method 0xe6a4347d.
//
// Solidity: function setSlashFeeSplit(uint16 burnBps, uint16 treasuryBps) returns()
func (_BunkerStaking *BunkerStakingSession) SetSlashFeeSplit(burnBps uint16, treasuryBps uint16) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetSlashFeeSplit(&_BunkerStaking.TransactOpts, burnBps, treasuryBps)
}

// SetSlashFeeSplit is a paid mutator transaction binding the contract method 0xe6a4347d.
//
// Solidity: function setSlashFeeSplit(uint16 burnBps, uint16 treasuryBps) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetSlashFeeSplit(burnBps uint16, treasuryBps uint16) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetSlashFeeSplit(&_BunkerStaking.TransactOpts, burnBps, treasuryBps)
}

// SetSlashPercentage is a paid mutator transaction binding the contract method 0xc87a1ec8.
//
// Solidity: function setSlashPercentage(uint8 reason, uint16 bps) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetSlashPercentage(opts *bind.TransactOpts, reason uint8, bps uint16) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setSlashPercentage", reason, bps)
}

// SetSlashPercentage is a paid mutator transaction binding the contract method 0xc87a1ec8.
//
// Solidity: function setSlashPercentage(uint8 reason, uint16 bps) returns()
func (_BunkerStaking *BunkerStakingSession) SetSlashPercentage(reason uint8, bps uint16) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetSlashPercentage(&_BunkerStaking.TransactOpts, reason, bps)
}

// SetSlashPercentage is a paid mutator transaction binding the contract method 0xc87a1ec8.
//
// Solidity: function setSlashPercentage(uint8 reason, uint16 bps) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetSlashPercentage(reason uint8, bps uint16) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetSlashPercentage(&_BunkerStaking.TransactOpts, reason, bps)
}

// SetSlashingEnabled is a paid mutator transaction binding the contract method 0x3d357473.
//
// Solidity: function setSlashingEnabled(bool enabled) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetSlashingEnabled(opts *bind.TransactOpts, enabled bool) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setSlashingEnabled", enabled)
}

// SetSlashingEnabled is a paid mutator transaction binding the contract method 0x3d357473.
//
// Solidity: function setSlashingEnabled(bool enabled) returns()
func (_BunkerStaking *BunkerStakingSession) SetSlashingEnabled(enabled bool) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetSlashingEnabled(&_BunkerStaking.TransactOpts, enabled)
}

// SetSlashingEnabled is a paid mutator transaction binding the contract method 0x3d357473.
//
// Solidity: function setSlashingEnabled(bool enabled) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetSlashingEnabled(enabled bool) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetSlashingEnabled(&_BunkerStaking.TransactOpts, enabled)
}

// SetTierMinStake is a paid mutator transaction binding the contract method 0xe5ae2240.
//
// Solidity: function setTierMinStake(uint8 tier, uint256 minStake) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetTierMinStake(opts *bind.TransactOpts, tier uint8, minStake *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setTierMinStake", tier, minStake)
}

// SetTierMinStake is a paid mutator transaction binding the contract method 0xe5ae2240.
//
// Solidity: function setTierMinStake(uint8 tier, uint256 minStake) returns()
func (_BunkerStaking *BunkerStakingSession) SetTierMinStake(tier uint8, minStake *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetTierMinStake(&_BunkerStaking.TransactOpts, tier, minStake)
}

// SetTierMinStake is a paid mutator transaction binding the contract method 0xe5ae2240.
//
// Solidity: function setTierMinStake(uint8 tier, uint256 minStake) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetTierMinStake(tier uint8, minStake *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetTierMinStake(&_BunkerStaking.TransactOpts, tier, minStake)
}

// SetTierRewardMultiplier is a paid mutator transaction binding the contract method 0x647b65ed.
//
// Solidity: function setTierRewardMultiplier(uint8 tier, uint16 multiplierBps) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetTierRewardMultiplier(opts *bind.TransactOpts, tier uint8, multiplierBps uint16) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setTierRewardMultiplier", tier, multiplierBps)
}

// SetTierRewardMultiplier is a paid mutator transaction binding the contract method 0x647b65ed.
//
// Solidity: function setTierRewardMultiplier(uint8 tier, uint16 multiplierBps) returns()
func (_BunkerStaking *BunkerStakingSession) SetTierRewardMultiplier(tier uint8, multiplierBps uint16) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetTierRewardMultiplier(&_BunkerStaking.TransactOpts, tier, multiplierBps)
}

// SetTierRewardMultiplier is a paid mutator transaction binding the contract method 0x647b65ed.
//
// Solidity: function setTierRewardMultiplier(uint8 tier, uint16 multiplierBps) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetTierRewardMultiplier(tier uint8, multiplierBps uint16) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetTierRewardMultiplier(&_BunkerStaking.TransactOpts, tier, multiplierBps)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address newTreasury) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetTreasury(opts *bind.TransactOpts, newTreasury common.Address) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setTreasury", newTreasury)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address newTreasury) returns()
func (_BunkerStaking *BunkerStakingSession) SetTreasury(newTreasury common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetTreasury(&_BunkerStaking.TransactOpts, newTreasury)
}

// SetTreasury is a paid mutator transaction binding the contract method 0xf0f44260.
//
// Solidity: function setTreasury(address newTreasury) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetTreasury(newTreasury common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetTreasury(&_BunkerStaking.TransactOpts, newTreasury)
}

// SetUnbondingPeriod is a paid mutator transaction binding the contract method 0x114eaf55.
//
// Solidity: function setUnbondingPeriod(uint256 newPeriod) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetUnbondingPeriod(opts *bind.TransactOpts, newPeriod *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setUnbondingPeriod", newPeriod)
}

// SetUnbondingPeriod is a paid mutator transaction binding the contract method 0x114eaf55.
//
// Solidity: function setUnbondingPeriod(uint256 newPeriod) returns()
func (_BunkerStaking *BunkerStakingSession) SetUnbondingPeriod(newPeriod *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetUnbondingPeriod(&_BunkerStaking.TransactOpts, newPeriod)
}

// SetUnbondingPeriod is a paid mutator transaction binding the contract method 0x114eaf55.
//
// Solidity: function setUnbondingPeriod(uint256 newPeriod) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetUnbondingPeriod(newPeriod *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetUnbondingPeriod(&_BunkerStaking.TransactOpts, newPeriod)
}

// SetVestingParams is a paid mutator transaction binding the contract method 0xeb658878.
//
// Solidity: function setVestingParams(uint256 _vestingPeriod, uint256 _immediateReleaseBps) returns()
func (_BunkerStaking *BunkerStakingTransactor) SetVestingParams(opts *bind.TransactOpts, _vestingPeriod *big.Int, _immediateReleaseBps *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "setVestingParams", _vestingPeriod, _immediateReleaseBps)
}

// SetVestingParams is a paid mutator transaction binding the contract method 0xeb658878.
//
// Solidity: function setVestingParams(uint256 _vestingPeriod, uint256 _immediateReleaseBps) returns()
func (_BunkerStaking *BunkerStakingSession) SetVestingParams(_vestingPeriod *big.Int, _immediateReleaseBps *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetVestingParams(&_BunkerStaking.TransactOpts, _vestingPeriod, _immediateReleaseBps)
}

// SetVestingParams is a paid mutator transaction binding the contract method 0xeb658878.
//
// Solidity: function setVestingParams(uint256 _vestingPeriod, uint256 _immediateReleaseBps) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SetVestingParams(_vestingPeriod *big.Int, _immediateReleaseBps *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SetVestingParams(&_BunkerStaking.TransactOpts, _vestingPeriod, _immediateReleaseBps)
}

// Slash is a paid mutator transaction binding the contract method 0x02fb4d85.
//
// Solidity: function slash(address provider, uint256 amount) returns()
func (_BunkerStaking *BunkerStakingTransactor) Slash(opts *bind.TransactOpts, provider common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "slash", provider, amount)
}

// Slash is a paid mutator transaction binding the contract method 0x02fb4d85.
//
// Solidity: function slash(address provider, uint256 amount) returns()
func (_BunkerStaking *BunkerStakingSession) Slash(provider common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.Slash(&_BunkerStaking.TransactOpts, provider, amount)
}

// Slash is a paid mutator transaction binding the contract method 0x02fb4d85.
//
// Solidity: function slash(address provider, uint256 amount) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) Slash(provider common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.Slash(&_BunkerStaking.TransactOpts, provider, amount)
}

// SlashImmediate is a paid mutator transaction binding the contract method 0x817b467a.
//
// Solidity: function slashImmediate(address provider, uint256 amount) returns()
func (_BunkerStaking *BunkerStakingTransactor) SlashImmediate(opts *bind.TransactOpts, provider common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "slashImmediate", provider, amount)
}

// SlashImmediate is a paid mutator transaction binding the contract method 0x817b467a.
//
// Solidity: function slashImmediate(address provider, uint256 amount) returns()
func (_BunkerStaking *BunkerStakingSession) SlashImmediate(provider common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SlashImmediate(&_BunkerStaking.TransactOpts, provider, amount)
}

// SlashImmediate is a paid mutator transaction binding the contract method 0x817b467a.
//
// Solidity: function slashImmediate(address provider, uint256 amount) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) SlashImmediate(provider common.Address, amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.SlashImmediate(&_BunkerStaking.TransactOpts, provider, amount)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 amount) returns()
func (_BunkerStaking *BunkerStakingTransactor) Stake(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "stake", amount)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 amount) returns()
func (_BunkerStaking *BunkerStakingSession) Stake(amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.Stake(&_BunkerStaking.TransactOpts, amount)
}

// Stake is a paid mutator transaction binding the contract method 0xa694fc3a.
//
// Solidity: function stake(uint256 amount) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) Stake(amount *big.Int) (*types.Transaction, error) {
	return _BunkerStaking.Contract.Stake(&_BunkerStaking.TransactOpts, amount)
}

// StakeWithIdentity is a paid mutator transaction binding the contract method 0xaf428c3b.
//
// Solidity: function stakeWithIdentity(uint256 amount, bytes32 nodeId, bytes32 region, uint64 capabilities) returns()
func (_BunkerStaking *BunkerStakingTransactor) StakeWithIdentity(opts *bind.TransactOpts, amount *big.Int, nodeId [32]byte, region [32]byte, capabilities uint64) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "stakeWithIdentity", amount, nodeId, region, capabilities)
}

// StakeWithIdentity is a paid mutator transaction binding the contract method 0xaf428c3b.
//
// Solidity: function stakeWithIdentity(uint256 amount, bytes32 nodeId, bytes32 region, uint64 capabilities) returns()
func (_BunkerStaking *BunkerStakingSession) StakeWithIdentity(amount *big.Int, nodeId [32]byte, region [32]byte, capabilities uint64) (*types.Transaction, error) {
	return _BunkerStaking.Contract.StakeWithIdentity(&_BunkerStaking.TransactOpts, amount, nodeId, region, capabilities)
}

// StakeWithIdentity is a paid mutator transaction binding the contract method 0xaf428c3b.
//
// Solidity: function stakeWithIdentity(uint256 amount, bytes32 nodeId, bytes32 region, uint64 capabilities) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) StakeWithIdentity(amount *big.Int, nodeId [32]byte, region [32]byte, capabilities uint64) (*types.Transaction, error) {
	return _BunkerStaking.Contract.StakeWithIdentity(&_BunkerStaking.TransactOpts, amount, nodeId, region, capabilities)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerStaking *BunkerStakingTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerStaking *BunkerStakingSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.TransferOwnership(&_BunkerStaking.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.TransferOwnership(&_BunkerStaking.TransactOpts, newOwner)
}

// UnfreezeProvider is a paid mutator transaction binding the contract method 0x349ec18a.
//
// Solidity: function unfreezeProvider(address provider) returns()
func (_BunkerStaking *BunkerStakingTransactor) UnfreezeProvider(opts *bind.TransactOpts, provider common.Address) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "unfreezeProvider", provider)
}

// UnfreezeProvider is a paid mutator transaction binding the contract method 0x349ec18a.
//
// Solidity: function unfreezeProvider(address provider) returns()
func (_BunkerStaking *BunkerStakingSession) UnfreezeProvider(provider common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.UnfreezeProvider(&_BunkerStaking.TransactOpts, provider)
}

// UnfreezeProvider is a paid mutator transaction binding the contract method 0x349ec18a.
//
// Solidity: function unfreezeProvider(address provider) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) UnfreezeProvider(provider common.Address) (*types.Transaction, error) {
	return _BunkerStaking.Contract.UnfreezeProvider(&_BunkerStaking.TransactOpts, provider)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerStaking *BunkerStakingTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerStaking *BunkerStakingSession) Unpause() (*types.Transaction, error) {
	return _BunkerStaking.Contract.Unpause(&_BunkerStaking.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_BunkerStaking *BunkerStakingTransactorSession) Unpause() (*types.Transaction, error) {
	return _BunkerStaking.Contract.Unpause(&_BunkerStaking.TransactOpts)
}

// UpdateIdentity is a paid mutator transaction binding the contract method 0x1607da4c.
//
// Solidity: function updateIdentity(bytes32 nodeId, bytes32 region, uint64 capabilities) returns()
func (_BunkerStaking *BunkerStakingTransactor) UpdateIdentity(opts *bind.TransactOpts, nodeId [32]byte, region [32]byte, capabilities uint64) (*types.Transaction, error) {
	return _BunkerStaking.contract.Transact(opts, "updateIdentity", nodeId, region, capabilities)
}

// UpdateIdentity is a paid mutator transaction binding the contract method 0x1607da4c.
//
// Solidity: function updateIdentity(bytes32 nodeId, bytes32 region, uint64 capabilities) returns()
func (_BunkerStaking *BunkerStakingSession) UpdateIdentity(nodeId [32]byte, region [32]byte, capabilities uint64) (*types.Transaction, error) {
	return _BunkerStaking.Contract.UpdateIdentity(&_BunkerStaking.TransactOpts, nodeId, region, capabilities)
}

// UpdateIdentity is a paid mutator transaction binding the contract method 0x1607da4c.
//
// Solidity: function updateIdentity(bytes32 nodeId, bytes32 region, uint64 capabilities) returns()
func (_BunkerStaking *BunkerStakingTransactorSession) UpdateIdentity(nodeId [32]byte, region [32]byte, capabilities uint64) (*types.Transaction, error) {
	return _BunkerStaking.Contract.UpdateIdentity(&_BunkerStaking.TransactOpts, nodeId, region, capabilities)
}

// BunkerStakingAppealResolvedIterator is returned from FilterAppealResolved and is used to iterate over the raw logs and unpacked data for AppealResolved events raised by the BunkerStaking contract.
type BunkerStakingAppealResolvedIterator struct {
	Event *BunkerStakingAppealResolved // Event containing the contract specifics and raw log

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
func (it *BunkerStakingAppealResolvedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingAppealResolved)
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
		it.Event = new(BunkerStakingAppealResolved)
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
func (it *BunkerStakingAppealResolvedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingAppealResolvedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingAppealResolved represents a AppealResolved event raised by the BunkerStaking contract.
type BunkerStakingAppealResolved struct {
	ProposalId *big.Int
	Upheld     bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterAppealResolved is a free log retrieval operation binding the contract event 0x89cd618842088e9ada48e481bba3a9320eba61a9329ef23ede5abd02837084a5.
//
// Solidity: event AppealResolved(uint256 indexed proposalId, bool upheld)
func (_BunkerStaking *BunkerStakingFilterer) FilterAppealResolved(opts *bind.FilterOpts, proposalId []*big.Int) (*BunkerStakingAppealResolvedIterator, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "AppealResolved", proposalIdRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingAppealResolvedIterator{contract: _BunkerStaking.contract, event: "AppealResolved", logs: logs, sub: sub}, nil
}

// WatchAppealResolved is a free log subscription operation binding the contract event 0x89cd618842088e9ada48e481bba3a9320eba61a9329ef23ede5abd02837084a5.
//
// Solidity: event AppealResolved(uint256 indexed proposalId, bool upheld)
func (_BunkerStaking *BunkerStakingFilterer) WatchAppealResolved(opts *bind.WatchOpts, sink chan<- *BunkerStakingAppealResolved, proposalId []*big.Int) (event.Subscription, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "AppealResolved", proposalIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingAppealResolved)
				if err := _BunkerStaking.contract.UnpackLog(event, "AppealResolved", log); err != nil {
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

// ParseAppealResolved is a log parse operation binding the contract event 0x89cd618842088e9ada48e481bba3a9320eba61a9329ef23ede5abd02837084a5.
//
// Solidity: event AppealResolved(uint256 indexed proposalId, bool upheld)
func (_BunkerStaking *BunkerStakingFilterer) ParseAppealResolved(log types.Log) (*BunkerStakingAppealResolved, error) {
	event := new(BunkerStakingAppealResolved)
	if err := _BunkerStaking.contract.UnpackLog(event, "AppealResolved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingAppealWindowUpdatedIterator is returned from FilterAppealWindowUpdated and is used to iterate over the raw logs and unpacked data for AppealWindowUpdated events raised by the BunkerStaking contract.
type BunkerStakingAppealWindowUpdatedIterator struct {
	Event *BunkerStakingAppealWindowUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingAppealWindowUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingAppealWindowUpdated)
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
		it.Event = new(BunkerStakingAppealWindowUpdated)
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
func (it *BunkerStakingAppealWindowUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingAppealWindowUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingAppealWindowUpdated represents a AppealWindowUpdated event raised by the BunkerStaking contract.
type BunkerStakingAppealWindowUpdated struct {
	NewWindow *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAppealWindowUpdated is a free log retrieval operation binding the contract event 0x54413d229c766aa747f8b521ac83d355bbe788dfe3c7c6ec5595911a9fcd6a83.
//
// Solidity: event AppealWindowUpdated(uint256 newWindow)
func (_BunkerStaking *BunkerStakingFilterer) FilterAppealWindowUpdated(opts *bind.FilterOpts) (*BunkerStakingAppealWindowUpdatedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "AppealWindowUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingAppealWindowUpdatedIterator{contract: _BunkerStaking.contract, event: "AppealWindowUpdated", logs: logs, sub: sub}, nil
}

// WatchAppealWindowUpdated is a free log subscription operation binding the contract event 0x54413d229c766aa747f8b521ac83d355bbe788dfe3c7c6ec5595911a9fcd6a83.
//
// Solidity: event AppealWindowUpdated(uint256 newWindow)
func (_BunkerStaking *BunkerStakingFilterer) WatchAppealWindowUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingAppealWindowUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "AppealWindowUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingAppealWindowUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "AppealWindowUpdated", log); err != nil {
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

// ParseAppealWindowUpdated is a log parse operation binding the contract event 0x54413d229c766aa747f8b521ac83d355bbe788dfe3c7c6ec5595911a9fcd6a83.
//
// Solidity: event AppealWindowUpdated(uint256 newWindow)
func (_BunkerStaking *BunkerStakingFilterer) ParseAppealWindowUpdated(log types.Log) (*BunkerStakingAppealWindowUpdated, error) {
	event := new(BunkerStakingAppealWindowUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "AppealWindowUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingBeneficiaryChangeInitiatedIterator is returned from FilterBeneficiaryChangeInitiated and is used to iterate over the raw logs and unpacked data for BeneficiaryChangeInitiated events raised by the BunkerStaking contract.
type BunkerStakingBeneficiaryChangeInitiatedIterator struct {
	Event *BunkerStakingBeneficiaryChangeInitiated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingBeneficiaryChangeInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingBeneficiaryChangeInitiated)
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
		it.Event = new(BunkerStakingBeneficiaryChangeInitiated)
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
func (it *BunkerStakingBeneficiaryChangeInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingBeneficiaryChangeInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingBeneficiaryChangeInitiated represents a BeneficiaryChangeInitiated event raised by the BunkerStaking contract.
type BunkerStakingBeneficiaryChangeInitiated struct {
	Provider       common.Address
	NewBeneficiary common.Address
	EffectiveTime  *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterBeneficiaryChangeInitiated is a free log retrieval operation binding the contract event 0x98da5d426da0c3b191c5600bb598cf868f68f6592e8aae0ffcc8722e050f9208.
//
// Solidity: event BeneficiaryChangeInitiated(address indexed provider, address indexed newBeneficiary, uint256 effectiveTime)
func (_BunkerStaking *BunkerStakingFilterer) FilterBeneficiaryChangeInitiated(opts *bind.FilterOpts, provider []common.Address, newBeneficiary []common.Address) (*BunkerStakingBeneficiaryChangeInitiatedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var newBeneficiaryRule []interface{}
	for _, newBeneficiaryItem := range newBeneficiary {
		newBeneficiaryRule = append(newBeneficiaryRule, newBeneficiaryItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "BeneficiaryChangeInitiated", providerRule, newBeneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingBeneficiaryChangeInitiatedIterator{contract: _BunkerStaking.contract, event: "BeneficiaryChangeInitiated", logs: logs, sub: sub}, nil
}

// WatchBeneficiaryChangeInitiated is a free log subscription operation binding the contract event 0x98da5d426da0c3b191c5600bb598cf868f68f6592e8aae0ffcc8722e050f9208.
//
// Solidity: event BeneficiaryChangeInitiated(address indexed provider, address indexed newBeneficiary, uint256 effectiveTime)
func (_BunkerStaking *BunkerStakingFilterer) WatchBeneficiaryChangeInitiated(opts *bind.WatchOpts, sink chan<- *BunkerStakingBeneficiaryChangeInitiated, provider []common.Address, newBeneficiary []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var newBeneficiaryRule []interface{}
	for _, newBeneficiaryItem := range newBeneficiary {
		newBeneficiaryRule = append(newBeneficiaryRule, newBeneficiaryItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "BeneficiaryChangeInitiated", providerRule, newBeneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingBeneficiaryChangeInitiated)
				if err := _BunkerStaking.contract.UnpackLog(event, "BeneficiaryChangeInitiated", log); err != nil {
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

// ParseBeneficiaryChangeInitiated is a log parse operation binding the contract event 0x98da5d426da0c3b191c5600bb598cf868f68f6592e8aae0ffcc8722e050f9208.
//
// Solidity: event BeneficiaryChangeInitiated(address indexed provider, address indexed newBeneficiary, uint256 effectiveTime)
func (_BunkerStaking *BunkerStakingFilterer) ParseBeneficiaryChangeInitiated(log types.Log) (*BunkerStakingBeneficiaryChangeInitiated, error) {
	event := new(BunkerStakingBeneficiaryChangeInitiated)
	if err := _BunkerStaking.contract.UnpackLog(event, "BeneficiaryChangeInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingBeneficiaryChangedIterator is returned from FilterBeneficiaryChanged and is used to iterate over the raw logs and unpacked data for BeneficiaryChanged events raised by the BunkerStaking contract.
type BunkerStakingBeneficiaryChangedIterator struct {
	Event *BunkerStakingBeneficiaryChanged // Event containing the contract specifics and raw log

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
func (it *BunkerStakingBeneficiaryChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingBeneficiaryChanged)
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
		it.Event = new(BunkerStakingBeneficiaryChanged)
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
func (it *BunkerStakingBeneficiaryChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingBeneficiaryChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingBeneficiaryChanged represents a BeneficiaryChanged event raised by the BunkerStaking contract.
type BunkerStakingBeneficiaryChanged struct {
	Provider       common.Address
	OldBeneficiary common.Address
	NewBeneficiary common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterBeneficiaryChanged is a free log retrieval operation binding the contract event 0xdc2fee73e6c685172c975dd7a10bdc4da20294ae742590263dfcd6a59681dcc5.
//
// Solidity: event BeneficiaryChanged(address indexed provider, address indexed oldBeneficiary, address indexed newBeneficiary)
func (_BunkerStaking *BunkerStakingFilterer) FilterBeneficiaryChanged(opts *bind.FilterOpts, provider []common.Address, oldBeneficiary []common.Address, newBeneficiary []common.Address) (*BunkerStakingBeneficiaryChangedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var oldBeneficiaryRule []interface{}
	for _, oldBeneficiaryItem := range oldBeneficiary {
		oldBeneficiaryRule = append(oldBeneficiaryRule, oldBeneficiaryItem)
	}
	var newBeneficiaryRule []interface{}
	for _, newBeneficiaryItem := range newBeneficiary {
		newBeneficiaryRule = append(newBeneficiaryRule, newBeneficiaryItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "BeneficiaryChanged", providerRule, oldBeneficiaryRule, newBeneficiaryRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingBeneficiaryChangedIterator{contract: _BunkerStaking.contract, event: "BeneficiaryChanged", logs: logs, sub: sub}, nil
}

// WatchBeneficiaryChanged is a free log subscription operation binding the contract event 0xdc2fee73e6c685172c975dd7a10bdc4da20294ae742590263dfcd6a59681dcc5.
//
// Solidity: event BeneficiaryChanged(address indexed provider, address indexed oldBeneficiary, address indexed newBeneficiary)
func (_BunkerStaking *BunkerStakingFilterer) WatchBeneficiaryChanged(opts *bind.WatchOpts, sink chan<- *BunkerStakingBeneficiaryChanged, provider []common.Address, oldBeneficiary []common.Address, newBeneficiary []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var oldBeneficiaryRule []interface{}
	for _, oldBeneficiaryItem := range oldBeneficiary {
		oldBeneficiaryRule = append(oldBeneficiaryRule, oldBeneficiaryItem)
	}
	var newBeneficiaryRule []interface{}
	for _, newBeneficiaryItem := range newBeneficiary {
		newBeneficiaryRule = append(newBeneficiaryRule, newBeneficiaryItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "BeneficiaryChanged", providerRule, oldBeneficiaryRule, newBeneficiaryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingBeneficiaryChanged)
				if err := _BunkerStaking.contract.UnpackLog(event, "BeneficiaryChanged", log); err != nil {
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

// ParseBeneficiaryChanged is a log parse operation binding the contract event 0xdc2fee73e6c685172c975dd7a10bdc4da20294ae742590263dfcd6a59681dcc5.
//
// Solidity: event BeneficiaryChanged(address indexed provider, address indexed oldBeneficiary, address indexed newBeneficiary)
func (_BunkerStaking *BunkerStakingFilterer) ParseBeneficiaryChanged(log types.Log) (*BunkerStakingBeneficiaryChanged, error) {
	event := new(BunkerStakingBeneficiaryChanged)
	if err := _BunkerStaking.contract.UnpackLog(event, "BeneficiaryChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingComputeHoursReportedIterator is returned from FilterComputeHoursReported and is used to iterate over the raw logs and unpacked data for ComputeHoursReported events raised by the BunkerStaking contract.
type BunkerStakingComputeHoursReportedIterator struct {
	Event *BunkerStakingComputeHoursReported // Event containing the contract specifics and raw log

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
func (it *BunkerStakingComputeHoursReportedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingComputeHoursReported)
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
		it.Event = new(BunkerStakingComputeHoursReported)
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
func (it *BunkerStakingComputeHoursReportedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingComputeHoursReportedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingComputeHoursReported represents a ComputeHoursReported event raised by the BunkerStaking contract.
type BunkerStakingComputeHoursReported struct {
	Hours             *big.Int
	TotalComputeHours *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterComputeHoursReported is a free log retrieval operation binding the contract event 0xcd4e5f23e306594173b24ca7f1a71b37ffb973b416d19566af45e2139dca1b30.
//
// Solidity: event ComputeHoursReported(uint256 hours_, uint256 totalComputeHours)
func (_BunkerStaking *BunkerStakingFilterer) FilterComputeHoursReported(opts *bind.FilterOpts) (*BunkerStakingComputeHoursReportedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "ComputeHoursReported")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingComputeHoursReportedIterator{contract: _BunkerStaking.contract, event: "ComputeHoursReported", logs: logs, sub: sub}, nil
}

// WatchComputeHoursReported is a free log subscription operation binding the contract event 0xcd4e5f23e306594173b24ca7f1a71b37ffb973b416d19566af45e2139dca1b30.
//
// Solidity: event ComputeHoursReported(uint256 hours_, uint256 totalComputeHours)
func (_BunkerStaking *BunkerStakingFilterer) WatchComputeHoursReported(opts *bind.WatchOpts, sink chan<- *BunkerStakingComputeHoursReported) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "ComputeHoursReported")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingComputeHoursReported)
				if err := _BunkerStaking.contract.UnpackLog(event, "ComputeHoursReported", log); err != nil {
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

// ParseComputeHoursReported is a log parse operation binding the contract event 0xcd4e5f23e306594173b24ca7f1a71b37ffb973b416d19566af45e2139dca1b30.
//
// Solidity: event ComputeHoursReported(uint256 hours_, uint256 totalComputeHours)
func (_BunkerStaking *BunkerStakingFilterer) ParseComputeHoursReported(log types.Log) (*BunkerStakingComputeHoursReported, error) {
	event := new(BunkerStakingComputeHoursReported)
	if err := _BunkerStaking.contract.UnpackLog(event, "ComputeHoursReported", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingEmissionMultiplierUpdatedIterator is returned from FilterEmissionMultiplierUpdated and is used to iterate over the raw logs and unpacked data for EmissionMultiplierUpdated events raised by the BunkerStaking contract.
type BunkerStakingEmissionMultiplierUpdatedIterator struct {
	Event *BunkerStakingEmissionMultiplierUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingEmissionMultiplierUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingEmissionMultiplierUpdated)
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
		it.Event = new(BunkerStakingEmissionMultiplierUpdated)
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
func (it *BunkerStakingEmissionMultiplierUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingEmissionMultiplierUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingEmissionMultiplierUpdated represents a EmissionMultiplierUpdated event raised by the BunkerStaking contract.
type BunkerStakingEmissionMultiplierUpdated struct {
	MultiplierBps *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterEmissionMultiplierUpdated is a free log retrieval operation binding the contract event 0x3665e914cbbfb56f8df84838e27e44aab666f6d78acbaf60f7c541a398fd6f54.
//
// Solidity: event EmissionMultiplierUpdated(uint256 multiplierBps)
func (_BunkerStaking *BunkerStakingFilterer) FilterEmissionMultiplierUpdated(opts *bind.FilterOpts) (*BunkerStakingEmissionMultiplierUpdatedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "EmissionMultiplierUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingEmissionMultiplierUpdatedIterator{contract: _BunkerStaking.contract, event: "EmissionMultiplierUpdated", logs: logs, sub: sub}, nil
}

// WatchEmissionMultiplierUpdated is a free log subscription operation binding the contract event 0x3665e914cbbfb56f8df84838e27e44aab666f6d78acbaf60f7c541a398fd6f54.
//
// Solidity: event EmissionMultiplierUpdated(uint256 multiplierBps)
func (_BunkerStaking *BunkerStakingFilterer) WatchEmissionMultiplierUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingEmissionMultiplierUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "EmissionMultiplierUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingEmissionMultiplierUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "EmissionMultiplierUpdated", log); err != nil {
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

// ParseEmissionMultiplierUpdated is a log parse operation binding the contract event 0x3665e914cbbfb56f8df84838e27e44aab666f6d78acbaf60f7c541a398fd6f54.
//
// Solidity: event EmissionMultiplierUpdated(uint256 multiplierBps)
func (_BunkerStaking *BunkerStakingFilterer) ParseEmissionMultiplierUpdated(log types.Log) (*BunkerStakingEmissionMultiplierUpdated, error) {
	event := new(BunkerStakingEmissionMultiplierUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "EmissionMultiplierUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingMaxEmissionRateUpdatedIterator is returned from FilterMaxEmissionRateUpdated and is used to iterate over the raw logs and unpacked data for MaxEmissionRateUpdated events raised by the BunkerStaking contract.
type BunkerStakingMaxEmissionRateUpdatedIterator struct {
	Event *BunkerStakingMaxEmissionRateUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingMaxEmissionRateUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingMaxEmissionRateUpdated)
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
		it.Event = new(BunkerStakingMaxEmissionRateUpdated)
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
func (it *BunkerStakingMaxEmissionRateUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingMaxEmissionRateUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingMaxEmissionRateUpdated represents a MaxEmissionRateUpdated event raised by the BunkerStaking contract.
type BunkerStakingMaxEmissionRateUpdated struct {
	MaxRate *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMaxEmissionRateUpdated is a free log retrieval operation binding the contract event 0xc44e0289ac09e3f9a9c9e667e4c25e2aea77e0fc8a89677a961c423a35082e81.
//
// Solidity: event MaxEmissionRateUpdated(uint256 maxRate)
func (_BunkerStaking *BunkerStakingFilterer) FilterMaxEmissionRateUpdated(opts *bind.FilterOpts) (*BunkerStakingMaxEmissionRateUpdatedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "MaxEmissionRateUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingMaxEmissionRateUpdatedIterator{contract: _BunkerStaking.contract, event: "MaxEmissionRateUpdated", logs: logs, sub: sub}, nil
}

// WatchMaxEmissionRateUpdated is a free log subscription operation binding the contract event 0xc44e0289ac09e3f9a9c9e667e4c25e2aea77e0fc8a89677a961c423a35082e81.
//
// Solidity: event MaxEmissionRateUpdated(uint256 maxRate)
func (_BunkerStaking *BunkerStakingFilterer) WatchMaxEmissionRateUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingMaxEmissionRateUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "MaxEmissionRateUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingMaxEmissionRateUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "MaxEmissionRateUpdated", log); err != nil {
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

// ParseMaxEmissionRateUpdated is a log parse operation binding the contract event 0xc44e0289ac09e3f9a9c9e667e4c25e2aea77e0fc8a89677a961c423a35082e81.
//
// Solidity: event MaxEmissionRateUpdated(uint256 maxRate)
func (_BunkerStaking *BunkerStakingFilterer) ParseMaxEmissionRateUpdated(log types.Log) (*BunkerStakingMaxEmissionRateUpdated, error) {
	event := new(BunkerStakingMaxEmissionRateUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "MaxEmissionRateUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingMaxTierMultiplierUpdatedIterator is returned from FilterMaxTierMultiplierUpdated and is used to iterate over the raw logs and unpacked data for MaxTierMultiplierUpdated events raised by the BunkerStaking contract.
type BunkerStakingMaxTierMultiplierUpdatedIterator struct {
	Event *BunkerStakingMaxTierMultiplierUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingMaxTierMultiplierUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingMaxTierMultiplierUpdated)
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
		it.Event = new(BunkerStakingMaxTierMultiplierUpdated)
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
func (it *BunkerStakingMaxTierMultiplierUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingMaxTierMultiplierUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingMaxTierMultiplierUpdated represents a MaxTierMultiplierUpdated event raised by the BunkerStaking contract.
type BunkerStakingMaxTierMultiplierUpdated struct {
	NewMax uint16
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMaxTierMultiplierUpdated is a free log retrieval operation binding the contract event 0x51842c09c69e9b362aa5f1861b50689dd0af1ae0111cea7050599ccb2105e31b.
//
// Solidity: event MaxTierMultiplierUpdated(uint16 newMax)
func (_BunkerStaking *BunkerStakingFilterer) FilterMaxTierMultiplierUpdated(opts *bind.FilterOpts) (*BunkerStakingMaxTierMultiplierUpdatedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "MaxTierMultiplierUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingMaxTierMultiplierUpdatedIterator{contract: _BunkerStaking.contract, event: "MaxTierMultiplierUpdated", logs: logs, sub: sub}, nil
}

// WatchMaxTierMultiplierUpdated is a free log subscription operation binding the contract event 0x51842c09c69e9b362aa5f1861b50689dd0af1ae0111cea7050599ccb2105e31b.
//
// Solidity: event MaxTierMultiplierUpdated(uint16 newMax)
func (_BunkerStaking *BunkerStakingFilterer) WatchMaxTierMultiplierUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingMaxTierMultiplierUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "MaxTierMultiplierUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingMaxTierMultiplierUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "MaxTierMultiplierUpdated", log); err != nil {
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

// ParseMaxTierMultiplierUpdated is a log parse operation binding the contract event 0x51842c09c69e9b362aa5f1861b50689dd0af1ae0111cea7050599ccb2105e31b.
//
// Solidity: event MaxTierMultiplierUpdated(uint16 newMax)
func (_BunkerStaking *BunkerStakingFilterer) ParseMaxTierMultiplierUpdated(log types.Log) (*BunkerStakingMaxTierMultiplierUpdated, error) {
	event := new(BunkerStakingMaxTierMultiplierUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "MaxTierMultiplierUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingOwnershipTransferStartedIterator is returned from FilterOwnershipTransferStarted and is used to iterate over the raw logs and unpacked data for OwnershipTransferStarted events raised by the BunkerStaking contract.
type BunkerStakingOwnershipTransferStartedIterator struct {
	Event *BunkerStakingOwnershipTransferStarted // Event containing the contract specifics and raw log

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
func (it *BunkerStakingOwnershipTransferStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingOwnershipTransferStarted)
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
		it.Event = new(BunkerStakingOwnershipTransferStarted)
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
func (it *BunkerStakingOwnershipTransferStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingOwnershipTransferStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingOwnershipTransferStarted represents a OwnershipTransferStarted event raised by the BunkerStaking contract.
type BunkerStakingOwnershipTransferStarted struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferStarted is a free log retrieval operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerStaking *BunkerStakingFilterer) FilterOwnershipTransferStarted(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerStakingOwnershipTransferStartedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingOwnershipTransferStartedIterator{contract: _BunkerStaking.contract, event: "OwnershipTransferStarted", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferStarted is a free log subscription operation binding the contract event 0x38d16b8cac22d99fc7c124b9cd0de2d3fa1faef420bfe791d8c362d765e22700.
//
// Solidity: event OwnershipTransferStarted(address indexed previousOwner, address indexed newOwner)
func (_BunkerStaking *BunkerStakingFilterer) WatchOwnershipTransferStarted(opts *bind.WatchOpts, sink chan<- *BunkerStakingOwnershipTransferStarted, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "OwnershipTransferStarted", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingOwnershipTransferStarted)
				if err := _BunkerStaking.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
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
func (_BunkerStaking *BunkerStakingFilterer) ParseOwnershipTransferStarted(log types.Log) (*BunkerStakingOwnershipTransferStarted, error) {
	event := new(BunkerStakingOwnershipTransferStarted)
	if err := _BunkerStaking.contract.UnpackLog(event, "OwnershipTransferStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BunkerStaking contract.
type BunkerStakingOwnershipTransferredIterator struct {
	Event *BunkerStakingOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BunkerStakingOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingOwnershipTransferred)
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
		it.Event = new(BunkerStakingOwnershipTransferred)
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
func (it *BunkerStakingOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingOwnershipTransferred represents a OwnershipTransferred event raised by the BunkerStaking contract.
type BunkerStakingOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerStaking *BunkerStakingFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BunkerStakingOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingOwnershipTransferredIterator{contract: _BunkerStaking.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BunkerStaking *BunkerStakingFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BunkerStakingOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingOwnershipTransferred)
				if err := _BunkerStaking.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BunkerStaking *BunkerStakingFilterer) ParseOwnershipTransferred(log types.Log) (*BunkerStakingOwnershipTransferred, error) {
	event := new(BunkerStakingOwnershipTransferred)
	if err := _BunkerStaking.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the BunkerStaking contract.
type BunkerStakingPausedIterator struct {
	Event *BunkerStakingPaused // Event containing the contract specifics and raw log

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
func (it *BunkerStakingPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingPaused)
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
		it.Event = new(BunkerStakingPaused)
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
func (it *BunkerStakingPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingPaused represents a Paused event raised by the BunkerStaking contract.
type BunkerStakingPaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_BunkerStaking *BunkerStakingFilterer) FilterPaused(opts *bind.FilterOpts) (*BunkerStakingPausedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingPausedIterator{contract: _BunkerStaking.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0x62e78cea01bee320cd4e420270b5ea74000d11b0c9f74754ebdbfc544b05a258.
//
// Solidity: event Paused(address account)
func (_BunkerStaking *BunkerStakingFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *BunkerStakingPaused) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "Paused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingPaused)
				if err := _BunkerStaking.contract.UnpackLog(event, "Paused", log); err != nil {
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
func (_BunkerStaking *BunkerStakingFilterer) ParsePaused(log types.Log) (*BunkerStakingPaused, error) {
	event := new(BunkerStakingPaused)
	if err := _BunkerStaking.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingProviderDeregisteredIterator is returned from FilterProviderDeregistered and is used to iterate over the raw logs and unpacked data for ProviderDeregistered events raised by the BunkerStaking contract.
type BunkerStakingProviderDeregisteredIterator struct {
	Event *BunkerStakingProviderDeregistered // Event containing the contract specifics and raw log

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
func (it *BunkerStakingProviderDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingProviderDeregistered)
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
		it.Event = new(BunkerStakingProviderDeregistered)
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
func (it *BunkerStakingProviderDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingProviderDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingProviderDeregistered represents a ProviderDeregistered event raised by the BunkerStaking contract.
type BunkerStakingProviderDeregistered struct {
	Provider common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderDeregistered is a free log retrieval operation binding the contract event 0xf04091b4a187e321a42001e46961e45b6a75b203fc6fb766b7e05505f6080abb.
//
// Solidity: event ProviderDeregistered(address indexed provider)
func (_BunkerStaking *BunkerStakingFilterer) FilterProviderDeregistered(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingProviderDeregisteredIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "ProviderDeregistered", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingProviderDeregisteredIterator{contract: _BunkerStaking.contract, event: "ProviderDeregistered", logs: logs, sub: sub}, nil
}

// WatchProviderDeregistered is a free log subscription operation binding the contract event 0xf04091b4a187e321a42001e46961e45b6a75b203fc6fb766b7e05505f6080abb.
//
// Solidity: event ProviderDeregistered(address indexed provider)
func (_BunkerStaking *BunkerStakingFilterer) WatchProviderDeregistered(opts *bind.WatchOpts, sink chan<- *BunkerStakingProviderDeregistered, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "ProviderDeregistered", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingProviderDeregistered)
				if err := _BunkerStaking.contract.UnpackLog(event, "ProviderDeregistered", log); err != nil {
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

// ParseProviderDeregistered is a log parse operation binding the contract event 0xf04091b4a187e321a42001e46961e45b6a75b203fc6fb766b7e05505f6080abb.
//
// Solidity: event ProviderDeregistered(address indexed provider)
func (_BunkerStaking *BunkerStakingFilterer) ParseProviderDeregistered(log types.Log) (*BunkerStakingProviderDeregistered, error) {
	event := new(BunkerStakingProviderDeregistered)
	if err := _BunkerStaking.contract.UnpackLog(event, "ProviderDeregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingProviderFrozenIterator is returned from FilterProviderFrozen and is used to iterate over the raw logs and unpacked data for ProviderFrozen events raised by the BunkerStaking contract.
type BunkerStakingProviderFrozenIterator struct {
	Event *BunkerStakingProviderFrozen // Event containing the contract specifics and raw log

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
func (it *BunkerStakingProviderFrozenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingProviderFrozen)
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
		it.Event = new(BunkerStakingProviderFrozen)
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
func (it *BunkerStakingProviderFrozenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingProviderFrozenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingProviderFrozen represents a ProviderFrozen event raised by the BunkerStaking contract.
type BunkerStakingProviderFrozen struct {
	Provider common.Address
	By       common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderFrozen is a free log retrieval operation binding the contract event 0x65eea5ca6f5e9a2ded541a1264e7cc2d2c3ebd7e8de01cbf48d345ea50920f75.
//
// Solidity: event ProviderFrozen(address indexed provider, address indexed by)
func (_BunkerStaking *BunkerStakingFilterer) FilterProviderFrozen(opts *bind.FilterOpts, provider []common.Address, by []common.Address) (*BunkerStakingProviderFrozenIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var byRule []interface{}
	for _, byItem := range by {
		byRule = append(byRule, byItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "ProviderFrozen", providerRule, byRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingProviderFrozenIterator{contract: _BunkerStaking.contract, event: "ProviderFrozen", logs: logs, sub: sub}, nil
}

// WatchProviderFrozen is a free log subscription operation binding the contract event 0x65eea5ca6f5e9a2ded541a1264e7cc2d2c3ebd7e8de01cbf48d345ea50920f75.
//
// Solidity: event ProviderFrozen(address indexed provider, address indexed by)
func (_BunkerStaking *BunkerStakingFilterer) WatchProviderFrozen(opts *bind.WatchOpts, sink chan<- *BunkerStakingProviderFrozen, provider []common.Address, by []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var byRule []interface{}
	for _, byItem := range by {
		byRule = append(byRule, byItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "ProviderFrozen", providerRule, byRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingProviderFrozen)
				if err := _BunkerStaking.contract.UnpackLog(event, "ProviderFrozen", log); err != nil {
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

// ParseProviderFrozen is a log parse operation binding the contract event 0x65eea5ca6f5e9a2ded541a1264e7cc2d2c3ebd7e8de01cbf48d345ea50920f75.
//
// Solidity: event ProviderFrozen(address indexed provider, address indexed by)
func (_BunkerStaking *BunkerStakingFilterer) ParseProviderFrozen(log types.Log) (*BunkerStakingProviderFrozen, error) {
	event := new(BunkerStakingProviderFrozen)
	if err := _BunkerStaking.contract.UnpackLog(event, "ProviderFrozen", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingProviderIdentityUpdatedIterator is returned from FilterProviderIdentityUpdated and is used to iterate over the raw logs and unpacked data for ProviderIdentityUpdated events raised by the BunkerStaking contract.
type BunkerStakingProviderIdentityUpdatedIterator struct {
	Event *BunkerStakingProviderIdentityUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingProviderIdentityUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingProviderIdentityUpdated)
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
		it.Event = new(BunkerStakingProviderIdentityUpdated)
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
func (it *BunkerStakingProviderIdentityUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingProviderIdentityUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingProviderIdentityUpdated represents a ProviderIdentityUpdated event raised by the BunkerStaking contract.
type BunkerStakingProviderIdentityUpdated struct {
	Provider     common.Address
	NodeId       [32]byte
	Region       [32]byte
	Capabilities uint64
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterProviderIdentityUpdated is a free log retrieval operation binding the contract event 0x1cee51a2b872997d671c1f8951c74077c6e67098f7a60021794d5ac33f597c0f.
//
// Solidity: event ProviderIdentityUpdated(address indexed provider, bytes32 nodeId, bytes32 region, uint64 capabilities)
func (_BunkerStaking *BunkerStakingFilterer) FilterProviderIdentityUpdated(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingProviderIdentityUpdatedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "ProviderIdentityUpdated", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingProviderIdentityUpdatedIterator{contract: _BunkerStaking.contract, event: "ProviderIdentityUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderIdentityUpdated is a free log subscription operation binding the contract event 0x1cee51a2b872997d671c1f8951c74077c6e67098f7a60021794d5ac33f597c0f.
//
// Solidity: event ProviderIdentityUpdated(address indexed provider, bytes32 nodeId, bytes32 region, uint64 capabilities)
func (_BunkerStaking *BunkerStakingFilterer) WatchProviderIdentityUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingProviderIdentityUpdated, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "ProviderIdentityUpdated", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingProviderIdentityUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "ProviderIdentityUpdated", log); err != nil {
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

// ParseProviderIdentityUpdated is a log parse operation binding the contract event 0x1cee51a2b872997d671c1f8951c74077c6e67098f7a60021794d5ac33f597c0f.
//
// Solidity: event ProviderIdentityUpdated(address indexed provider, bytes32 nodeId, bytes32 region, uint64 capabilities)
func (_BunkerStaking *BunkerStakingFilterer) ParseProviderIdentityUpdated(log types.Log) (*BunkerStakingProviderIdentityUpdated, error) {
	event := new(BunkerStakingProviderIdentityUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "ProviderIdentityUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingProviderRegisteredIterator is returned from FilterProviderRegistered and is used to iterate over the raw logs and unpacked data for ProviderRegistered events raised by the BunkerStaking contract.
type BunkerStakingProviderRegisteredIterator struct {
	Event *BunkerStakingProviderRegistered // Event containing the contract specifics and raw log

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
func (it *BunkerStakingProviderRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingProviderRegistered)
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
		it.Event = new(BunkerStakingProviderRegistered)
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
func (it *BunkerStakingProviderRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingProviderRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingProviderRegistered represents a ProviderRegistered event raised by the BunkerStaking contract.
type BunkerStakingProviderRegistered struct {
	Provider    common.Address
	StakeAmount *big.Int
	Tier        uint8
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterProviderRegistered is a free log retrieval operation binding the contract event 0x57d314643bba38495581cebe1f7627d299bd42671765bfd77105c4f5093a530a.
//
// Solidity: event ProviderRegistered(address indexed provider, uint256 stakeAmount, uint8 tier)
func (_BunkerStaking *BunkerStakingFilterer) FilterProviderRegistered(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingProviderRegisteredIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "ProviderRegistered", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingProviderRegisteredIterator{contract: _BunkerStaking.contract, event: "ProviderRegistered", logs: logs, sub: sub}, nil
}

// WatchProviderRegistered is a free log subscription operation binding the contract event 0x57d314643bba38495581cebe1f7627d299bd42671765bfd77105c4f5093a530a.
//
// Solidity: event ProviderRegistered(address indexed provider, uint256 stakeAmount, uint8 tier)
func (_BunkerStaking *BunkerStakingFilterer) WatchProviderRegistered(opts *bind.WatchOpts, sink chan<- *BunkerStakingProviderRegistered, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "ProviderRegistered", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingProviderRegistered)
				if err := _BunkerStaking.contract.UnpackLog(event, "ProviderRegistered", log); err != nil {
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

// ParseProviderRegistered is a log parse operation binding the contract event 0x57d314643bba38495581cebe1f7627d299bd42671765bfd77105c4f5093a530a.
//
// Solidity: event ProviderRegistered(address indexed provider, uint256 stakeAmount, uint8 tier)
func (_BunkerStaking *BunkerStakingFilterer) ParseProviderRegistered(log types.Log) (*BunkerStakingProviderRegistered, error) {
	event := new(BunkerStakingProviderRegistered)
	if err := _BunkerStaking.contract.UnpackLog(event, "ProviderRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingProviderUnfrozenIterator is returned from FilterProviderUnfrozen and is used to iterate over the raw logs and unpacked data for ProviderUnfrozen events raised by the BunkerStaking contract.
type BunkerStakingProviderUnfrozenIterator struct {
	Event *BunkerStakingProviderUnfrozen // Event containing the contract specifics and raw log

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
func (it *BunkerStakingProviderUnfrozenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingProviderUnfrozen)
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
		it.Event = new(BunkerStakingProviderUnfrozen)
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
func (it *BunkerStakingProviderUnfrozenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingProviderUnfrozenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingProviderUnfrozen represents a ProviderUnfrozen event raised by the BunkerStaking contract.
type BunkerStakingProviderUnfrozen struct {
	Provider common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterProviderUnfrozen is a free log retrieval operation binding the contract event 0x33fa283d77758f91ac312d2ad1600b5797226bb3f77ebe1b2aaad28a4713b4a1.
//
// Solidity: event ProviderUnfrozen(address indexed provider)
func (_BunkerStaking *BunkerStakingFilterer) FilterProviderUnfrozen(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingProviderUnfrozenIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "ProviderUnfrozen", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingProviderUnfrozenIterator{contract: _BunkerStaking.contract, event: "ProviderUnfrozen", logs: logs, sub: sub}, nil
}

// WatchProviderUnfrozen is a free log subscription operation binding the contract event 0x33fa283d77758f91ac312d2ad1600b5797226bb3f77ebe1b2aaad28a4713b4a1.
//
// Solidity: event ProviderUnfrozen(address indexed provider)
func (_BunkerStaking *BunkerStakingFilterer) WatchProviderUnfrozen(opts *bind.WatchOpts, sink chan<- *BunkerStakingProviderUnfrozen, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "ProviderUnfrozen", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingProviderUnfrozen)
				if err := _BunkerStaking.contract.UnpackLog(event, "ProviderUnfrozen", log); err != nil {
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

// ParseProviderUnfrozen is a log parse operation binding the contract event 0x33fa283d77758f91ac312d2ad1600b5797226bb3f77ebe1b2aaad28a4713b4a1.
//
// Solidity: event ProviderUnfrozen(address indexed provider)
func (_BunkerStaking *BunkerStakingFilterer) ParseProviderUnfrozen(log types.Log) (*BunkerStakingProviderUnfrozen, error) {
	event := new(BunkerStakingProviderUnfrozen)
	if err := _BunkerStaking.contract.UnpackLog(event, "ProviderUnfrozen", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingRewardClaimedIterator is returned from FilterRewardClaimed and is used to iterate over the raw logs and unpacked data for RewardClaimed events raised by the BunkerStaking contract.
type BunkerStakingRewardClaimedIterator struct {
	Event *BunkerStakingRewardClaimed // Event containing the contract specifics and raw log

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
func (it *BunkerStakingRewardClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingRewardClaimed)
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
		it.Event = new(BunkerStakingRewardClaimed)
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
func (it *BunkerStakingRewardClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingRewardClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingRewardClaimed represents a RewardClaimed event raised by the BunkerStaking contract.
type BunkerStakingRewardClaimed struct {
	Provider common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRewardClaimed is a free log retrieval operation binding the contract event 0x106f923f993c2149d49b4255ff723acafa1f2d94393f561d3eda32ae348f7241.
//
// Solidity: event RewardClaimed(address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) FilterRewardClaimed(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingRewardClaimedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "RewardClaimed", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingRewardClaimedIterator{contract: _BunkerStaking.contract, event: "RewardClaimed", logs: logs, sub: sub}, nil
}

// WatchRewardClaimed is a free log subscription operation binding the contract event 0x106f923f993c2149d49b4255ff723acafa1f2d94393f561d3eda32ae348f7241.
//
// Solidity: event RewardClaimed(address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) WatchRewardClaimed(opts *bind.WatchOpts, sink chan<- *BunkerStakingRewardClaimed, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "RewardClaimed", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingRewardClaimed)
				if err := _BunkerStaking.contract.UnpackLog(event, "RewardClaimed", log); err != nil {
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

// ParseRewardClaimed is a log parse operation binding the contract event 0x106f923f993c2149d49b4255ff723acafa1f2d94393f561d3eda32ae348f7241.
//
// Solidity: event RewardClaimed(address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) ParseRewardClaimed(log types.Log) (*BunkerStakingRewardClaimed, error) {
	event := new(BunkerStakingRewardClaimed)
	if err := _BunkerStaking.contract.UnpackLog(event, "RewardClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingRewardEpochStartedIterator is returned from FilterRewardEpochStarted and is used to iterate over the raw logs and unpacked data for RewardEpochStarted events raised by the BunkerStaking contract.
type BunkerStakingRewardEpochStartedIterator struct {
	Event *BunkerStakingRewardEpochStarted // Event containing the contract specifics and raw log

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
func (it *BunkerStakingRewardEpochStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingRewardEpochStarted)
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
		it.Event = new(BunkerStakingRewardEpochStarted)
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
func (it *BunkerStakingRewardEpochStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingRewardEpochStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingRewardEpochStarted represents a RewardEpochStarted event raised by the BunkerStaking contract.
type BunkerStakingRewardEpochStarted struct {
	Reward       *big.Int
	RewardRate   *big.Int
	PeriodFinish *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterRewardEpochStarted is a free log retrieval operation binding the contract event 0x347a584423bcd4ce3b9eef1f89febdd551066b57473a8c30c7cb7b5384defd10.
//
// Solidity: event RewardEpochStarted(uint256 reward, uint256 rewardRate, uint256 periodFinish)
func (_BunkerStaking *BunkerStakingFilterer) FilterRewardEpochStarted(opts *bind.FilterOpts) (*BunkerStakingRewardEpochStartedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "RewardEpochStarted")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingRewardEpochStartedIterator{contract: _BunkerStaking.contract, event: "RewardEpochStarted", logs: logs, sub: sub}, nil
}

// WatchRewardEpochStarted is a free log subscription operation binding the contract event 0x347a584423bcd4ce3b9eef1f89febdd551066b57473a8c30c7cb7b5384defd10.
//
// Solidity: event RewardEpochStarted(uint256 reward, uint256 rewardRate, uint256 periodFinish)
func (_BunkerStaking *BunkerStakingFilterer) WatchRewardEpochStarted(opts *bind.WatchOpts, sink chan<- *BunkerStakingRewardEpochStarted) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "RewardEpochStarted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingRewardEpochStarted)
				if err := _BunkerStaking.contract.UnpackLog(event, "RewardEpochStarted", log); err != nil {
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

// ParseRewardEpochStarted is a log parse operation binding the contract event 0x347a584423bcd4ce3b9eef1f89febdd551066b57473a8c30c7cb7b5384defd10.
//
// Solidity: event RewardEpochStarted(uint256 reward, uint256 rewardRate, uint256 periodFinish)
func (_BunkerStaking *BunkerStakingFilterer) ParseRewardEpochStarted(log types.Log) (*BunkerStakingRewardEpochStarted, error) {
	event := new(BunkerStakingRewardEpochStarted)
	if err := _BunkerStaking.contract.UnpackLog(event, "RewardEpochStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingRewardVestedIterator is returned from FilterRewardVested and is used to iterate over the raw logs and unpacked data for RewardVested events raised by the BunkerStaking contract.
type BunkerStakingRewardVestedIterator struct {
	Event *BunkerStakingRewardVested // Event containing the contract specifics and raw log

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
func (it *BunkerStakingRewardVestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingRewardVested)
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
		it.Event = new(BunkerStakingRewardVested)
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
func (it *BunkerStakingRewardVestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingRewardVestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingRewardVested represents a RewardVested event raised by the BunkerStaking contract.
type BunkerStakingRewardVested struct {
	Provider        common.Address
	TotalAmount     *big.Int
	ImmediateAmount *big.Int
	VestedAmount    *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterRewardVested is a free log retrieval operation binding the contract event 0x475055cbae9aa34fd8d18d08d5012fc98dcbbcdfc5540d79b84e2a0caa7ab42d.
//
// Solidity: event RewardVested(address indexed provider, uint256 totalAmount, uint256 immediateAmount, uint256 vestedAmount)
func (_BunkerStaking *BunkerStakingFilterer) FilterRewardVested(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingRewardVestedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "RewardVested", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingRewardVestedIterator{contract: _BunkerStaking.contract, event: "RewardVested", logs: logs, sub: sub}, nil
}

// WatchRewardVested is a free log subscription operation binding the contract event 0x475055cbae9aa34fd8d18d08d5012fc98dcbbcdfc5540d79b84e2a0caa7ab42d.
//
// Solidity: event RewardVested(address indexed provider, uint256 totalAmount, uint256 immediateAmount, uint256 vestedAmount)
func (_BunkerStaking *BunkerStakingFilterer) WatchRewardVested(opts *bind.WatchOpts, sink chan<- *BunkerStakingRewardVested, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "RewardVested", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingRewardVested)
				if err := _BunkerStaking.contract.UnpackLog(event, "RewardVested", log); err != nil {
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

// ParseRewardVested is a log parse operation binding the contract event 0x475055cbae9aa34fd8d18d08d5012fc98dcbbcdfc5540d79b84e2a0caa7ab42d.
//
// Solidity: event RewardVested(address indexed provider, uint256 totalAmount, uint256 immediateAmount, uint256 vestedAmount)
func (_BunkerStaking *BunkerStakingFilterer) ParseRewardVested(log types.Log) (*BunkerStakingRewardVested, error) {
	event := new(BunkerStakingRewardVested)
	if err := _BunkerStaking.contract.UnpackLog(event, "RewardVested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingRewardsDurationUpdatedIterator is returned from FilterRewardsDurationUpdated and is used to iterate over the raw logs and unpacked data for RewardsDurationUpdated events raised by the BunkerStaking contract.
type BunkerStakingRewardsDurationUpdatedIterator struct {
	Event *BunkerStakingRewardsDurationUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingRewardsDurationUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingRewardsDurationUpdated)
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
		it.Event = new(BunkerStakingRewardsDurationUpdated)
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
func (it *BunkerStakingRewardsDurationUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingRewardsDurationUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingRewardsDurationUpdated represents a RewardsDurationUpdated event raised by the BunkerStaking contract.
type BunkerStakingRewardsDurationUpdated struct {
	NewDuration *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterRewardsDurationUpdated is a free log retrieval operation binding the contract event 0xfb46ca5a5e06d4540d6387b930a7c978bce0db5f449ec6b3f5d07c6e1d44f2d3.
//
// Solidity: event RewardsDurationUpdated(uint256 newDuration)
func (_BunkerStaking *BunkerStakingFilterer) FilterRewardsDurationUpdated(opts *bind.FilterOpts) (*BunkerStakingRewardsDurationUpdatedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "RewardsDurationUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingRewardsDurationUpdatedIterator{contract: _BunkerStaking.contract, event: "RewardsDurationUpdated", logs: logs, sub: sub}, nil
}

// WatchRewardsDurationUpdated is a free log subscription operation binding the contract event 0xfb46ca5a5e06d4540d6387b930a7c978bce0db5f449ec6b3f5d07c6e1d44f2d3.
//
// Solidity: event RewardsDurationUpdated(uint256 newDuration)
func (_BunkerStaking *BunkerStakingFilterer) WatchRewardsDurationUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingRewardsDurationUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "RewardsDurationUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingRewardsDurationUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "RewardsDurationUpdated", log); err != nil {
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

// ParseRewardsDurationUpdated is a log parse operation binding the contract event 0xfb46ca5a5e06d4540d6387b930a7c978bce0db5f449ec6b3f5d07c6e1d44f2d3.
//
// Solidity: event RewardsDurationUpdated(uint256 newDuration)
func (_BunkerStaking *BunkerStakingFilterer) ParseRewardsDurationUpdated(log types.Log) (*BunkerStakingRewardsDurationUpdated, error) {
	event := new(BunkerStakingRewardsDurationUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "RewardsDurationUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the BunkerStaking contract.
type BunkerStakingRoleAdminChangedIterator struct {
	Event *BunkerStakingRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *BunkerStakingRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingRoleAdminChanged)
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
		it.Event = new(BunkerStakingRoleAdminChanged)
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
func (it *BunkerStakingRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingRoleAdminChanged represents a RoleAdminChanged event raised by the BunkerStaking contract.
type BunkerStakingRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BunkerStaking *BunkerStakingFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*BunkerStakingRoleAdminChangedIterator, error) {

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

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingRoleAdminChangedIterator{contract: _BunkerStaking.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_BunkerStaking *BunkerStakingFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *BunkerStakingRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

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

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingRoleAdminChanged)
				if err := _BunkerStaking.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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
func (_BunkerStaking *BunkerStakingFilterer) ParseRoleAdminChanged(log types.Log) (*BunkerStakingRoleAdminChanged, error) {
	event := new(BunkerStakingRoleAdminChanged)
	if err := _BunkerStaking.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the BunkerStaking contract.
type BunkerStakingRoleGrantedIterator struct {
	Event *BunkerStakingRoleGranted // Event containing the contract specifics and raw log

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
func (it *BunkerStakingRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingRoleGranted)
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
		it.Event = new(BunkerStakingRoleGranted)
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
func (it *BunkerStakingRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingRoleGranted represents a RoleGranted event raised by the BunkerStaking contract.
type BunkerStakingRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerStaking *BunkerStakingFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BunkerStakingRoleGrantedIterator, error) {

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

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingRoleGrantedIterator{contract: _BunkerStaking.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerStaking *BunkerStakingFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *BunkerStakingRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingRoleGranted)
				if err := _BunkerStaking.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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
func (_BunkerStaking *BunkerStakingFilterer) ParseRoleGranted(log types.Log) (*BunkerStakingRoleGranted, error) {
	event := new(BunkerStakingRoleGranted)
	if err := _BunkerStaking.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the BunkerStaking contract.
type BunkerStakingRoleRevokedIterator struct {
	Event *BunkerStakingRoleRevoked // Event containing the contract specifics and raw log

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
func (it *BunkerStakingRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingRoleRevoked)
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
		it.Event = new(BunkerStakingRoleRevoked)
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
func (it *BunkerStakingRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingRoleRevoked represents a RoleRevoked event raised by the BunkerStaking contract.
type BunkerStakingRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerStaking *BunkerStakingFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*BunkerStakingRoleRevokedIterator, error) {

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

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingRoleRevokedIterator{contract: _BunkerStaking.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_BunkerStaking *BunkerStakingFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *BunkerStakingRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingRoleRevoked)
				if err := _BunkerStaking.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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
func (_BunkerStaking *BunkerStakingFilterer) ParseRoleRevoked(log types.Log) (*BunkerStakingRoleRevoked, error) {
	event := new(BunkerStakingRoleRevoked)
	if err := _BunkerStaking.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingSlashAppealedIterator is returned from FilterSlashAppealed and is used to iterate over the raw logs and unpacked data for SlashAppealed events raised by the BunkerStaking contract.
type BunkerStakingSlashAppealedIterator struct {
	Event *BunkerStakingSlashAppealed // Event containing the contract specifics and raw log

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
func (it *BunkerStakingSlashAppealedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingSlashAppealed)
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
		it.Event = new(BunkerStakingSlashAppealed)
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
func (it *BunkerStakingSlashAppealedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingSlashAppealedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingSlashAppealed represents a SlashAppealed event raised by the BunkerStaking contract.
type BunkerStakingSlashAppealed struct {
	ProposalId *big.Int
	Provider   common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSlashAppealed is a free log retrieval operation binding the contract event 0xbd0352ee5139a1b683bb023cc369fd06666afe06da9100507190d7dea2d85e6f.
//
// Solidity: event SlashAppealed(uint256 indexed proposalId, address indexed provider)
func (_BunkerStaking *BunkerStakingFilterer) FilterSlashAppealed(opts *bind.FilterOpts, proposalId []*big.Int, provider []common.Address) (*BunkerStakingSlashAppealedIterator, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "SlashAppealed", proposalIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingSlashAppealedIterator{contract: _BunkerStaking.contract, event: "SlashAppealed", logs: logs, sub: sub}, nil
}

// WatchSlashAppealed is a free log subscription operation binding the contract event 0xbd0352ee5139a1b683bb023cc369fd06666afe06da9100507190d7dea2d85e6f.
//
// Solidity: event SlashAppealed(uint256 indexed proposalId, address indexed provider)
func (_BunkerStaking *BunkerStakingFilterer) WatchSlashAppealed(opts *bind.WatchOpts, sink chan<- *BunkerStakingSlashAppealed, proposalId []*big.Int, provider []common.Address) (event.Subscription, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "SlashAppealed", proposalIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingSlashAppealed)
				if err := _BunkerStaking.contract.UnpackLog(event, "SlashAppealed", log); err != nil {
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

// ParseSlashAppealed is a log parse operation binding the contract event 0xbd0352ee5139a1b683bb023cc369fd06666afe06da9100507190d7dea2d85e6f.
//
// Solidity: event SlashAppealed(uint256 indexed proposalId, address indexed provider)
func (_BunkerStaking *BunkerStakingFilterer) ParseSlashAppealed(log types.Log) (*BunkerStakingSlashAppealed, error) {
	event := new(BunkerStakingSlashAppealed)
	if err := _BunkerStaking.contract.UnpackLog(event, "SlashAppealed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingSlashFeeSplitUpdatedIterator is returned from FilterSlashFeeSplitUpdated and is used to iterate over the raw logs and unpacked data for SlashFeeSplitUpdated events raised by the BunkerStaking contract.
type BunkerStakingSlashFeeSplitUpdatedIterator struct {
	Event *BunkerStakingSlashFeeSplitUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingSlashFeeSplitUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingSlashFeeSplitUpdated)
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
		it.Event = new(BunkerStakingSlashFeeSplitUpdated)
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
func (it *BunkerStakingSlashFeeSplitUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingSlashFeeSplitUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingSlashFeeSplitUpdated represents a SlashFeeSplitUpdated event raised by the BunkerStaking contract.
type BunkerStakingSlashFeeSplitUpdated struct {
	BurnBps     uint16
	TreasuryBps uint16
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSlashFeeSplitUpdated is a free log retrieval operation binding the contract event 0x15b73021de90b6b84357315508cf8ed811c03180d19a0b761081c55467dab9ea.
//
// Solidity: event SlashFeeSplitUpdated(uint16 burnBps, uint16 treasuryBps)
func (_BunkerStaking *BunkerStakingFilterer) FilterSlashFeeSplitUpdated(opts *bind.FilterOpts) (*BunkerStakingSlashFeeSplitUpdatedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "SlashFeeSplitUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingSlashFeeSplitUpdatedIterator{contract: _BunkerStaking.contract, event: "SlashFeeSplitUpdated", logs: logs, sub: sub}, nil
}

// WatchSlashFeeSplitUpdated is a free log subscription operation binding the contract event 0x15b73021de90b6b84357315508cf8ed811c03180d19a0b761081c55467dab9ea.
//
// Solidity: event SlashFeeSplitUpdated(uint16 burnBps, uint16 treasuryBps)
func (_BunkerStaking *BunkerStakingFilterer) WatchSlashFeeSplitUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingSlashFeeSplitUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "SlashFeeSplitUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingSlashFeeSplitUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "SlashFeeSplitUpdated", log); err != nil {
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

// ParseSlashFeeSplitUpdated is a log parse operation binding the contract event 0x15b73021de90b6b84357315508cf8ed811c03180d19a0b761081c55467dab9ea.
//
// Solidity: event SlashFeeSplitUpdated(uint16 burnBps, uint16 treasuryBps)
func (_BunkerStaking *BunkerStakingFilterer) ParseSlashFeeSplitUpdated(log types.Log) (*BunkerStakingSlashFeeSplitUpdated, error) {
	event := new(BunkerStakingSlashFeeSplitUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "SlashFeeSplitUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingSlashPercentageUpdatedIterator is returned from FilterSlashPercentageUpdated and is used to iterate over the raw logs and unpacked data for SlashPercentageUpdated events raised by the BunkerStaking contract.
type BunkerStakingSlashPercentageUpdatedIterator struct {
	Event *BunkerStakingSlashPercentageUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingSlashPercentageUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingSlashPercentageUpdated)
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
		it.Event = new(BunkerStakingSlashPercentageUpdated)
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
func (it *BunkerStakingSlashPercentageUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingSlashPercentageUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingSlashPercentageUpdated represents a SlashPercentageUpdated event raised by the BunkerStaking contract.
type BunkerStakingSlashPercentageUpdated struct {
	Reason uint8
	Bps    uint16
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSlashPercentageUpdated is a free log retrieval operation binding the contract event 0xd9f07f2676b0b065e4321fe7215b8a66130f4a15cd8a0b9583429ad3c41cc63c.
//
// Solidity: event SlashPercentageUpdated(uint8 indexed reason, uint16 bps)
func (_BunkerStaking *BunkerStakingFilterer) FilterSlashPercentageUpdated(opts *bind.FilterOpts, reason []uint8) (*BunkerStakingSlashPercentageUpdatedIterator, error) {

	var reasonRule []interface{}
	for _, reasonItem := range reason {
		reasonRule = append(reasonRule, reasonItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "SlashPercentageUpdated", reasonRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingSlashPercentageUpdatedIterator{contract: _BunkerStaking.contract, event: "SlashPercentageUpdated", logs: logs, sub: sub}, nil
}

// WatchSlashPercentageUpdated is a free log subscription operation binding the contract event 0xd9f07f2676b0b065e4321fe7215b8a66130f4a15cd8a0b9583429ad3c41cc63c.
//
// Solidity: event SlashPercentageUpdated(uint8 indexed reason, uint16 bps)
func (_BunkerStaking *BunkerStakingFilterer) WatchSlashPercentageUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingSlashPercentageUpdated, reason []uint8) (event.Subscription, error) {

	var reasonRule []interface{}
	for _, reasonItem := range reason {
		reasonRule = append(reasonRule, reasonItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "SlashPercentageUpdated", reasonRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingSlashPercentageUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "SlashPercentageUpdated", log); err != nil {
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

// ParseSlashPercentageUpdated is a log parse operation binding the contract event 0xd9f07f2676b0b065e4321fe7215b8a66130f4a15cd8a0b9583429ad3c41cc63c.
//
// Solidity: event SlashPercentageUpdated(uint8 indexed reason, uint16 bps)
func (_BunkerStaking *BunkerStakingFilterer) ParseSlashPercentageUpdated(log types.Log) (*BunkerStakingSlashPercentageUpdated, error) {
	event := new(BunkerStakingSlashPercentageUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "SlashPercentageUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingSlashProposalExecutedIterator is returned from FilterSlashProposalExecuted and is used to iterate over the raw logs and unpacked data for SlashProposalExecuted events raised by the BunkerStaking contract.
type BunkerStakingSlashProposalExecutedIterator struct {
	Event *BunkerStakingSlashProposalExecuted // Event containing the contract specifics and raw log

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
func (it *BunkerStakingSlashProposalExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingSlashProposalExecuted)
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
		it.Event = new(BunkerStakingSlashProposalExecuted)
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
func (it *BunkerStakingSlashProposalExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingSlashProposalExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingSlashProposalExecuted represents a SlashProposalExecuted event raised by the BunkerStaking contract.
type BunkerStakingSlashProposalExecuted struct {
	ProposalId *big.Int
	Provider   common.Address
	Amount     *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSlashProposalExecuted is a free log retrieval operation binding the contract event 0x918b19ebaa57b40a3d5b3863d0d0d8697941812ac1e24e08f1226af6c441b0c4.
//
// Solidity: event SlashProposalExecuted(uint256 indexed proposalId, address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) FilterSlashProposalExecuted(opts *bind.FilterOpts, proposalId []*big.Int, provider []common.Address) (*BunkerStakingSlashProposalExecutedIterator, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "SlashProposalExecuted", proposalIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingSlashProposalExecutedIterator{contract: _BunkerStaking.contract, event: "SlashProposalExecuted", logs: logs, sub: sub}, nil
}

// WatchSlashProposalExecuted is a free log subscription operation binding the contract event 0x918b19ebaa57b40a3d5b3863d0d0d8697941812ac1e24e08f1226af6c441b0c4.
//
// Solidity: event SlashProposalExecuted(uint256 indexed proposalId, address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) WatchSlashProposalExecuted(opts *bind.WatchOpts, sink chan<- *BunkerStakingSlashProposalExecuted, proposalId []*big.Int, provider []common.Address) (event.Subscription, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "SlashProposalExecuted", proposalIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingSlashProposalExecuted)
				if err := _BunkerStaking.contract.UnpackLog(event, "SlashProposalExecuted", log); err != nil {
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

// ParseSlashProposalExecuted is a log parse operation binding the contract event 0x918b19ebaa57b40a3d5b3863d0d0d8697941812ac1e24e08f1226af6c441b0c4.
//
// Solidity: event SlashProposalExecuted(uint256 indexed proposalId, address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) ParseSlashProposalExecuted(log types.Log) (*BunkerStakingSlashProposalExecuted, error) {
	event := new(BunkerStakingSlashProposalExecuted)
	if err := _BunkerStaking.contract.UnpackLog(event, "SlashProposalExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingSlashProposedIterator is returned from FilterSlashProposed and is used to iterate over the raw logs and unpacked data for SlashProposed events raised by the BunkerStaking contract.
type BunkerStakingSlashProposedIterator struct {
	Event *BunkerStakingSlashProposed // Event containing the contract specifics and raw log

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
func (it *BunkerStakingSlashProposedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingSlashProposed)
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
		it.Event = new(BunkerStakingSlashProposed)
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
func (it *BunkerStakingSlashProposedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingSlashProposedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingSlashProposed represents a SlashProposed event raised by the BunkerStaking contract.
type BunkerStakingSlashProposed struct {
	ProposalId *big.Int
	Provider   common.Address
	Amount     *big.Int
	Reason     string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSlashProposed is a free log retrieval operation binding the contract event 0xf9a4da41e5e2e24a83fa7b480cb6c8c275cf2bac4a62d7a085a42eb5f18fdf91.
//
// Solidity: event SlashProposed(uint256 indexed proposalId, address indexed provider, uint256 amount, string reason)
func (_BunkerStaking *BunkerStakingFilterer) FilterSlashProposed(opts *bind.FilterOpts, proposalId []*big.Int, provider []common.Address) (*BunkerStakingSlashProposedIterator, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "SlashProposed", proposalIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingSlashProposedIterator{contract: _BunkerStaking.contract, event: "SlashProposed", logs: logs, sub: sub}, nil
}

// WatchSlashProposed is a free log subscription operation binding the contract event 0xf9a4da41e5e2e24a83fa7b480cb6c8c275cf2bac4a62d7a085a42eb5f18fdf91.
//
// Solidity: event SlashProposed(uint256 indexed proposalId, address indexed provider, uint256 amount, string reason)
func (_BunkerStaking *BunkerStakingFilterer) WatchSlashProposed(opts *bind.WatchOpts, sink chan<- *BunkerStakingSlashProposed, proposalId []*big.Int, provider []common.Address) (event.Subscription, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "SlashProposed", proposalIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingSlashProposed)
				if err := _BunkerStaking.contract.UnpackLog(event, "SlashProposed", log); err != nil {
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

// ParseSlashProposed is a log parse operation binding the contract event 0xf9a4da41e5e2e24a83fa7b480cb6c8c275cf2bac4a62d7a085a42eb5f18fdf91.
//
// Solidity: event SlashProposed(uint256 indexed proposalId, address indexed provider, uint256 amount, string reason)
func (_BunkerStaking *BunkerStakingFilterer) ParseSlashProposed(log types.Log) (*BunkerStakingSlashProposed, error) {
	event := new(BunkerStakingSlashProposed)
	if err := _BunkerStaking.contract.UnpackLog(event, "SlashProposed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingSlashProposedByReasonIterator is returned from FilterSlashProposedByReason and is used to iterate over the raw logs and unpacked data for SlashProposedByReason events raised by the BunkerStaking contract.
type BunkerStakingSlashProposedByReasonIterator struct {
	Event *BunkerStakingSlashProposedByReason // Event containing the contract specifics and raw log

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
func (it *BunkerStakingSlashProposedByReasonIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingSlashProposedByReason)
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
		it.Event = new(BunkerStakingSlashProposedByReason)
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
func (it *BunkerStakingSlashProposedByReasonIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingSlashProposedByReasonIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingSlashProposedByReason represents a SlashProposedByReason event raised by the BunkerStaking contract.
type BunkerStakingSlashProposedByReason struct {
	ProposalId *big.Int
	Provider   common.Address
	Amount     *big.Int
	Reason     uint8
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSlashProposedByReason is a free log retrieval operation binding the contract event 0xadfef744d6013ba990154f6243ea13ab87ec3400c8d78d9186ca1b19c376fd3f.
//
// Solidity: event SlashProposedByReason(uint256 indexed proposalId, address indexed provider, uint256 amount, uint8 reason)
func (_BunkerStaking *BunkerStakingFilterer) FilterSlashProposedByReason(opts *bind.FilterOpts, proposalId []*big.Int, provider []common.Address) (*BunkerStakingSlashProposedByReasonIterator, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "SlashProposedByReason", proposalIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingSlashProposedByReasonIterator{contract: _BunkerStaking.contract, event: "SlashProposedByReason", logs: logs, sub: sub}, nil
}

// WatchSlashProposedByReason is a free log subscription operation binding the contract event 0xadfef744d6013ba990154f6243ea13ab87ec3400c8d78d9186ca1b19c376fd3f.
//
// Solidity: event SlashProposedByReason(uint256 indexed proposalId, address indexed provider, uint256 amount, uint8 reason)
func (_BunkerStaking *BunkerStakingFilterer) WatchSlashProposedByReason(opts *bind.WatchOpts, sink chan<- *BunkerStakingSlashProposedByReason, proposalId []*big.Int, provider []common.Address) (event.Subscription, error) {

	var proposalIdRule []interface{}
	for _, proposalIdItem := range proposalId {
		proposalIdRule = append(proposalIdRule, proposalIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "SlashProposedByReason", proposalIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingSlashProposedByReason)
				if err := _BunkerStaking.contract.UnpackLog(event, "SlashProposedByReason", log); err != nil {
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

// ParseSlashProposedByReason is a log parse operation binding the contract event 0xadfef744d6013ba990154f6243ea13ab87ec3400c8d78d9186ca1b19c376fd3f.
//
// Solidity: event SlashProposedByReason(uint256 indexed proposalId, address indexed provider, uint256 amount, uint8 reason)
func (_BunkerStaking *BunkerStakingFilterer) ParseSlashProposedByReason(log types.Log) (*BunkerStakingSlashProposedByReason, error) {
	event := new(BunkerStakingSlashProposedByReason)
	if err := _BunkerStaking.contract.UnpackLog(event, "SlashProposedByReason", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingSlashedIterator is returned from FilterSlashed and is used to iterate over the raw logs and unpacked data for Slashed events raised by the BunkerStaking contract.
type BunkerStakingSlashedIterator struct {
	Event *BunkerStakingSlashed // Event containing the contract specifics and raw log

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
func (it *BunkerStakingSlashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingSlashed)
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
		it.Event = new(BunkerStakingSlashed)
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
func (it *BunkerStakingSlashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingSlashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingSlashed represents a Slashed event raised by the BunkerStaking contract.
type BunkerStakingSlashed struct {
	Provider       common.Address
	TotalSlashed   *big.Int
	BurnedAmount   *big.Int
	TreasuryAmount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSlashed is a free log retrieval operation binding the contract event 0x23ee33e2cc85d581547d857dc227450a3e2ef8666fa2faa5b13f0a0893e4d4ad.
//
// Solidity: event Slashed(address indexed provider, uint256 totalSlashed, uint256 burnedAmount, uint256 treasuryAmount)
func (_BunkerStaking *BunkerStakingFilterer) FilterSlashed(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingSlashedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "Slashed", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingSlashedIterator{contract: _BunkerStaking.contract, event: "Slashed", logs: logs, sub: sub}, nil
}

// WatchSlashed is a free log subscription operation binding the contract event 0x23ee33e2cc85d581547d857dc227450a3e2ef8666fa2faa5b13f0a0893e4d4ad.
//
// Solidity: event Slashed(address indexed provider, uint256 totalSlashed, uint256 burnedAmount, uint256 treasuryAmount)
func (_BunkerStaking *BunkerStakingFilterer) WatchSlashed(opts *bind.WatchOpts, sink chan<- *BunkerStakingSlashed, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "Slashed", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingSlashed)
				if err := _BunkerStaking.contract.UnpackLog(event, "Slashed", log); err != nil {
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

// ParseSlashed is a log parse operation binding the contract event 0x23ee33e2cc85d581547d857dc227450a3e2ef8666fa2faa5b13f0a0893e4d4ad.
//
// Solidity: event Slashed(address indexed provider, uint256 totalSlashed, uint256 burnedAmount, uint256 treasuryAmount)
func (_BunkerStaking *BunkerStakingFilterer) ParseSlashed(log types.Log) (*BunkerStakingSlashed, error) {
	event := new(BunkerStakingSlashed)
	if err := _BunkerStaking.contract.UnpackLog(event, "Slashed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingSlashingEnabledUpdatedIterator is returned from FilterSlashingEnabledUpdated and is used to iterate over the raw logs and unpacked data for SlashingEnabledUpdated events raised by the BunkerStaking contract.
type BunkerStakingSlashingEnabledUpdatedIterator struct {
	Event *BunkerStakingSlashingEnabledUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingSlashingEnabledUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingSlashingEnabledUpdated)
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
		it.Event = new(BunkerStakingSlashingEnabledUpdated)
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
func (it *BunkerStakingSlashingEnabledUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingSlashingEnabledUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingSlashingEnabledUpdated represents a SlashingEnabledUpdated event raised by the BunkerStaking contract.
type BunkerStakingSlashingEnabledUpdated struct {
	Enabled bool
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterSlashingEnabledUpdated is a free log retrieval operation binding the contract event 0xc3ef19d6884f0eda58a206cc3949f40500fd0e290ba63d4448dd37c491b606f8.
//
// Solidity: event SlashingEnabledUpdated(bool enabled)
func (_BunkerStaking *BunkerStakingFilterer) FilterSlashingEnabledUpdated(opts *bind.FilterOpts) (*BunkerStakingSlashingEnabledUpdatedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "SlashingEnabledUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingSlashingEnabledUpdatedIterator{contract: _BunkerStaking.contract, event: "SlashingEnabledUpdated", logs: logs, sub: sub}, nil
}

// WatchSlashingEnabledUpdated is a free log subscription operation binding the contract event 0xc3ef19d6884f0eda58a206cc3949f40500fd0e290ba63d4448dd37c491b606f8.
//
// Solidity: event SlashingEnabledUpdated(bool enabled)
func (_BunkerStaking *BunkerStakingFilterer) WatchSlashingEnabledUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingSlashingEnabledUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "SlashingEnabledUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingSlashingEnabledUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "SlashingEnabledUpdated", log); err != nil {
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

// ParseSlashingEnabledUpdated is a log parse operation binding the contract event 0xc3ef19d6884f0eda58a206cc3949f40500fd0e290ba63d4448dd37c491b606f8.
//
// Solidity: event SlashingEnabledUpdated(bool enabled)
func (_BunkerStaking *BunkerStakingFilterer) ParseSlashingEnabledUpdated(log types.Log) (*BunkerStakingSlashingEnabledUpdated, error) {
	event := new(BunkerStakingSlashingEnabledUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "SlashingEnabledUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingStakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the BunkerStaking contract.
type BunkerStakingStakedIterator struct {
	Event *BunkerStakingStaked // Event containing the contract specifics and raw log

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
func (it *BunkerStakingStakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingStaked)
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
		it.Event = new(BunkerStakingStaked)
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
func (it *BunkerStakingStakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingStakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingStaked represents a Staked event raised by the BunkerStaking contract.
type BunkerStakingStaked struct {
	Provider   common.Address
	Amount     *big.Int
	TotalStake *big.Int
	Tier       uint8
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free log retrieval operation binding the contract event 0x11046f741396910185536955c79ec22a76d04229d188bc13962f8331b81008a9.
//
// Solidity: event Staked(address indexed provider, uint256 amount, uint256 totalStake, uint8 tier)
func (_BunkerStaking *BunkerStakingFilterer) FilterStaked(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingStakedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "Staked", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingStakedIterator{contract: _BunkerStaking.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free log subscription operation binding the contract event 0x11046f741396910185536955c79ec22a76d04229d188bc13962f8331b81008a9.
//
// Solidity: event Staked(address indexed provider, uint256 amount, uint256 totalStake, uint8 tier)
func (_BunkerStaking *BunkerStakingFilterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *BunkerStakingStaked, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "Staked", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingStaked)
				if err := _BunkerStaking.contract.UnpackLog(event, "Staked", log); err != nil {
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

// ParseStaked is a log parse operation binding the contract event 0x11046f741396910185536955c79ec22a76d04229d188bc13962f8331b81008a9.
//
// Solidity: event Staked(address indexed provider, uint256 amount, uint256 totalStake, uint8 tier)
func (_BunkerStaking *BunkerStakingFilterer) ParseStaked(log types.Log) (*BunkerStakingStaked, error) {
	event := new(BunkerStakingStaked)
	if err := _BunkerStaking.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingTierConfigUpdatedIterator is returned from FilterTierConfigUpdated and is used to iterate over the raw logs and unpacked data for TierConfigUpdated events raised by the BunkerStaking contract.
type BunkerStakingTierConfigUpdatedIterator struct {
	Event *BunkerStakingTierConfigUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingTierConfigUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingTierConfigUpdated)
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
		it.Event = new(BunkerStakingTierConfigUpdated)
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
func (it *BunkerStakingTierConfigUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingTierConfigUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingTierConfigUpdated represents a TierConfigUpdated event raised by the BunkerStaking contract.
type BunkerStakingTierConfigUpdated struct {
	Tier     uint8
	MinStake *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterTierConfigUpdated is a free log retrieval operation binding the contract event 0x058e74ecb6b959ca816a971cf6a85ecc6cb00effd98a67f02478c7fe3d40ec67.
//
// Solidity: event TierConfigUpdated(uint8 indexed tier, uint256 minStake)
func (_BunkerStaking *BunkerStakingFilterer) FilterTierConfigUpdated(opts *bind.FilterOpts, tier []uint8) (*BunkerStakingTierConfigUpdatedIterator, error) {

	var tierRule []interface{}
	for _, tierItem := range tier {
		tierRule = append(tierRule, tierItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "TierConfigUpdated", tierRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingTierConfigUpdatedIterator{contract: _BunkerStaking.contract, event: "TierConfigUpdated", logs: logs, sub: sub}, nil
}

// WatchTierConfigUpdated is a free log subscription operation binding the contract event 0x058e74ecb6b959ca816a971cf6a85ecc6cb00effd98a67f02478c7fe3d40ec67.
//
// Solidity: event TierConfigUpdated(uint8 indexed tier, uint256 minStake)
func (_BunkerStaking *BunkerStakingFilterer) WatchTierConfigUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingTierConfigUpdated, tier []uint8) (event.Subscription, error) {

	var tierRule []interface{}
	for _, tierItem := range tier {
		tierRule = append(tierRule, tierItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "TierConfigUpdated", tierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingTierConfigUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "TierConfigUpdated", log); err != nil {
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

// ParseTierConfigUpdated is a log parse operation binding the contract event 0x058e74ecb6b959ca816a971cf6a85ecc6cb00effd98a67f02478c7fe3d40ec67.
//
// Solidity: event TierConfigUpdated(uint8 indexed tier, uint256 minStake)
func (_BunkerStaking *BunkerStakingFilterer) ParseTierConfigUpdated(log types.Log) (*BunkerStakingTierConfigUpdated, error) {
	event := new(BunkerStakingTierConfigUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "TierConfigUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingTierRewardMultiplierUpdatedIterator is returned from FilterTierRewardMultiplierUpdated and is used to iterate over the raw logs and unpacked data for TierRewardMultiplierUpdated events raised by the BunkerStaking contract.
type BunkerStakingTierRewardMultiplierUpdatedIterator struct {
	Event *BunkerStakingTierRewardMultiplierUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingTierRewardMultiplierUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingTierRewardMultiplierUpdated)
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
		it.Event = new(BunkerStakingTierRewardMultiplierUpdated)
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
func (it *BunkerStakingTierRewardMultiplierUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingTierRewardMultiplierUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingTierRewardMultiplierUpdated represents a TierRewardMultiplierUpdated event raised by the BunkerStaking contract.
type BunkerStakingTierRewardMultiplierUpdated struct {
	Tier          uint8
	MultiplierBps uint16
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterTierRewardMultiplierUpdated is a free log retrieval operation binding the contract event 0x3ddbe19f2e67fb27e670a8e6970e56ed4fc0969de981bcfc91ef749b8a5143f8.
//
// Solidity: event TierRewardMultiplierUpdated(uint8 indexed tier, uint16 multiplierBps)
func (_BunkerStaking *BunkerStakingFilterer) FilterTierRewardMultiplierUpdated(opts *bind.FilterOpts, tier []uint8) (*BunkerStakingTierRewardMultiplierUpdatedIterator, error) {

	var tierRule []interface{}
	for _, tierItem := range tier {
		tierRule = append(tierRule, tierItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "TierRewardMultiplierUpdated", tierRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingTierRewardMultiplierUpdatedIterator{contract: _BunkerStaking.contract, event: "TierRewardMultiplierUpdated", logs: logs, sub: sub}, nil
}

// WatchTierRewardMultiplierUpdated is a free log subscription operation binding the contract event 0x3ddbe19f2e67fb27e670a8e6970e56ed4fc0969de981bcfc91ef749b8a5143f8.
//
// Solidity: event TierRewardMultiplierUpdated(uint8 indexed tier, uint16 multiplierBps)
func (_BunkerStaking *BunkerStakingFilterer) WatchTierRewardMultiplierUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingTierRewardMultiplierUpdated, tier []uint8) (event.Subscription, error) {

	var tierRule []interface{}
	for _, tierItem := range tier {
		tierRule = append(tierRule, tierItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "TierRewardMultiplierUpdated", tierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingTierRewardMultiplierUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "TierRewardMultiplierUpdated", log); err != nil {
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

// ParseTierRewardMultiplierUpdated is a log parse operation binding the contract event 0x3ddbe19f2e67fb27e670a8e6970e56ed4fc0969de981bcfc91ef749b8a5143f8.
//
// Solidity: event TierRewardMultiplierUpdated(uint8 indexed tier, uint16 multiplierBps)
func (_BunkerStaking *BunkerStakingFilterer) ParseTierRewardMultiplierUpdated(log types.Log) (*BunkerStakingTierRewardMultiplierUpdated, error) {
	event := new(BunkerStakingTierRewardMultiplierUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "TierRewardMultiplierUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingTreasuryUpdatedIterator is returned from FilterTreasuryUpdated and is used to iterate over the raw logs and unpacked data for TreasuryUpdated events raised by the BunkerStaking contract.
type BunkerStakingTreasuryUpdatedIterator struct {
	Event *BunkerStakingTreasuryUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingTreasuryUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingTreasuryUpdated)
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
		it.Event = new(BunkerStakingTreasuryUpdated)
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
func (it *BunkerStakingTreasuryUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingTreasuryUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingTreasuryUpdated represents a TreasuryUpdated event raised by the BunkerStaking contract.
type BunkerStakingTreasuryUpdated struct {
	OldTreasury common.Address
	NewTreasury common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterTreasuryUpdated is a free log retrieval operation binding the contract event 0x4ab5be82436d353e61ca18726e984e561f5c1cc7c6d38b29d2553c790434705a.
//
// Solidity: event TreasuryUpdated(address indexed oldTreasury, address indexed newTreasury)
func (_BunkerStaking *BunkerStakingFilterer) FilterTreasuryUpdated(opts *bind.FilterOpts, oldTreasury []common.Address, newTreasury []common.Address) (*BunkerStakingTreasuryUpdatedIterator, error) {

	var oldTreasuryRule []interface{}
	for _, oldTreasuryItem := range oldTreasury {
		oldTreasuryRule = append(oldTreasuryRule, oldTreasuryItem)
	}
	var newTreasuryRule []interface{}
	for _, newTreasuryItem := range newTreasury {
		newTreasuryRule = append(newTreasuryRule, newTreasuryItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "TreasuryUpdated", oldTreasuryRule, newTreasuryRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingTreasuryUpdatedIterator{contract: _BunkerStaking.contract, event: "TreasuryUpdated", logs: logs, sub: sub}, nil
}

// WatchTreasuryUpdated is a free log subscription operation binding the contract event 0x4ab5be82436d353e61ca18726e984e561f5c1cc7c6d38b29d2553c790434705a.
//
// Solidity: event TreasuryUpdated(address indexed oldTreasury, address indexed newTreasury)
func (_BunkerStaking *BunkerStakingFilterer) WatchTreasuryUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingTreasuryUpdated, oldTreasury []common.Address, newTreasury []common.Address) (event.Subscription, error) {

	var oldTreasuryRule []interface{}
	for _, oldTreasuryItem := range oldTreasury {
		oldTreasuryRule = append(oldTreasuryRule, oldTreasuryItem)
	}
	var newTreasuryRule []interface{}
	for _, newTreasuryItem := range newTreasury {
		newTreasuryRule = append(newTreasuryRule, newTreasuryItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "TreasuryUpdated", oldTreasuryRule, newTreasuryRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingTreasuryUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "TreasuryUpdated", log); err != nil {
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
func (_BunkerStaking *BunkerStakingFilterer) ParseTreasuryUpdated(log types.Log) (*BunkerStakingTreasuryUpdated, error) {
	event := new(BunkerStakingTreasuryUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "TreasuryUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingUnbondingPeriodUpdatedIterator is returned from FilterUnbondingPeriodUpdated and is used to iterate over the raw logs and unpacked data for UnbondingPeriodUpdated events raised by the BunkerStaking contract.
type BunkerStakingUnbondingPeriodUpdatedIterator struct {
	Event *BunkerStakingUnbondingPeriodUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingUnbondingPeriodUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingUnbondingPeriodUpdated)
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
		it.Event = new(BunkerStakingUnbondingPeriodUpdated)
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
func (it *BunkerStakingUnbondingPeriodUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingUnbondingPeriodUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingUnbondingPeriodUpdated represents a UnbondingPeriodUpdated event raised by the BunkerStaking contract.
type BunkerStakingUnbondingPeriodUpdated struct {
	NewPeriod *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterUnbondingPeriodUpdated is a free log retrieval operation binding the contract event 0x38a1644f59872db6fd17fdced395495fcaa3ca7d2825a0704db5a3acbd1dd064.
//
// Solidity: event UnbondingPeriodUpdated(uint256 newPeriod)
func (_BunkerStaking *BunkerStakingFilterer) FilterUnbondingPeriodUpdated(opts *bind.FilterOpts) (*BunkerStakingUnbondingPeriodUpdatedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "UnbondingPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingUnbondingPeriodUpdatedIterator{contract: _BunkerStaking.contract, event: "UnbondingPeriodUpdated", logs: logs, sub: sub}, nil
}

// WatchUnbondingPeriodUpdated is a free log subscription operation binding the contract event 0x38a1644f59872db6fd17fdced395495fcaa3ca7d2825a0704db5a3acbd1dd064.
//
// Solidity: event UnbondingPeriodUpdated(uint256 newPeriod)
func (_BunkerStaking *BunkerStakingFilterer) WatchUnbondingPeriodUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingUnbondingPeriodUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "UnbondingPeriodUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingUnbondingPeriodUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "UnbondingPeriodUpdated", log); err != nil {
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
func (_BunkerStaking *BunkerStakingFilterer) ParseUnbondingPeriodUpdated(log types.Log) (*BunkerStakingUnbondingPeriodUpdated, error) {
	event := new(BunkerStakingUnbondingPeriodUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "UnbondingPeriodUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the BunkerStaking contract.
type BunkerStakingUnpausedIterator struct {
	Event *BunkerStakingUnpaused // Event containing the contract specifics and raw log

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
func (it *BunkerStakingUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingUnpaused)
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
		it.Event = new(BunkerStakingUnpaused)
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
func (it *BunkerStakingUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingUnpaused represents a Unpaused event raised by the BunkerStaking contract.
type BunkerStakingUnpaused struct {
	Account common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_BunkerStaking *BunkerStakingFilterer) FilterUnpaused(opts *bind.FilterOpts) (*BunkerStakingUnpausedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingUnpausedIterator{contract: _BunkerStaking.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x5db9ee0a495bf2e6ff9c91a7834c1ba4fdd244a5e8aa4e537bd38aeae4b073aa.
//
// Solidity: event Unpaused(address account)
func (_BunkerStaking *BunkerStakingFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *BunkerStakingUnpaused) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "Unpaused")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingUnpaused)
				if err := _BunkerStaking.contract.UnpackLog(event, "Unpaused", log); err != nil {
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
func (_BunkerStaking *BunkerStakingFilterer) ParseUnpaused(log types.Log) (*BunkerStakingUnpaused, error) {
	event := new(BunkerStakingUnpaused)
	if err := _BunkerStaking.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingUnstakeCompletedIterator is returned from FilterUnstakeCompleted and is used to iterate over the raw logs and unpacked data for UnstakeCompleted events raised by the BunkerStaking contract.
type BunkerStakingUnstakeCompletedIterator struct {
	Event *BunkerStakingUnstakeCompleted // Event containing the contract specifics and raw log

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
func (it *BunkerStakingUnstakeCompletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingUnstakeCompleted)
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
		it.Event = new(BunkerStakingUnstakeCompleted)
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
func (it *BunkerStakingUnstakeCompletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingUnstakeCompletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingUnstakeCompleted represents a UnstakeCompleted event raised by the BunkerStaking contract.
type BunkerStakingUnstakeCompleted struct {
	Provider     common.Address
	Amount       *big.Int
	RequestIndex *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterUnstakeCompleted is a free log retrieval operation binding the contract event 0xe8cf66c4a1bfe34e4e2e2af0a72f79d3482978050d726564ff9d2e4220835d63.
//
// Solidity: event UnstakeCompleted(address indexed provider, uint256 amount, uint256 requestIndex)
func (_BunkerStaking *BunkerStakingFilterer) FilterUnstakeCompleted(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingUnstakeCompletedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "UnstakeCompleted", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingUnstakeCompletedIterator{contract: _BunkerStaking.contract, event: "UnstakeCompleted", logs: logs, sub: sub}, nil
}

// WatchUnstakeCompleted is a free log subscription operation binding the contract event 0xe8cf66c4a1bfe34e4e2e2af0a72f79d3482978050d726564ff9d2e4220835d63.
//
// Solidity: event UnstakeCompleted(address indexed provider, uint256 amount, uint256 requestIndex)
func (_BunkerStaking *BunkerStakingFilterer) WatchUnstakeCompleted(opts *bind.WatchOpts, sink chan<- *BunkerStakingUnstakeCompleted, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "UnstakeCompleted", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingUnstakeCompleted)
				if err := _BunkerStaking.contract.UnpackLog(event, "UnstakeCompleted", log); err != nil {
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

// ParseUnstakeCompleted is a log parse operation binding the contract event 0xe8cf66c4a1bfe34e4e2e2af0a72f79d3482978050d726564ff9d2e4220835d63.
//
// Solidity: event UnstakeCompleted(address indexed provider, uint256 amount, uint256 requestIndex)
func (_BunkerStaking *BunkerStakingFilterer) ParseUnstakeCompleted(log types.Log) (*BunkerStakingUnstakeCompleted, error) {
	event := new(BunkerStakingUnstakeCompleted)
	if err := _BunkerStaking.contract.UnpackLog(event, "UnstakeCompleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingUnstakeRequestedIterator is returned from FilterUnstakeRequested and is used to iterate over the raw logs and unpacked data for UnstakeRequested events raised by the BunkerStaking contract.
type BunkerStakingUnstakeRequestedIterator struct {
	Event *BunkerStakingUnstakeRequested // Event containing the contract specifics and raw log

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
func (it *BunkerStakingUnstakeRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingUnstakeRequested)
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
		it.Event = new(BunkerStakingUnstakeRequested)
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
func (it *BunkerStakingUnstakeRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingUnstakeRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingUnstakeRequested represents a UnstakeRequested event raised by the BunkerStaking contract.
type BunkerStakingUnstakeRequested struct {
	Provider     common.Address
	Amount       *big.Int
	UnlockTime   *big.Int
	RequestIndex *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterUnstakeRequested is a free log retrieval operation binding the contract event 0x6930caaa0f0843978bf16992d58b9fd54913ce2e45b8acdd34f5b44f95419db2.
//
// Solidity: event UnstakeRequested(address indexed provider, uint256 amount, uint256 unlockTime, uint256 requestIndex)
func (_BunkerStaking *BunkerStakingFilterer) FilterUnstakeRequested(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingUnstakeRequestedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "UnstakeRequested", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingUnstakeRequestedIterator{contract: _BunkerStaking.contract, event: "UnstakeRequested", logs: logs, sub: sub}, nil
}

// WatchUnstakeRequested is a free log subscription operation binding the contract event 0x6930caaa0f0843978bf16992d58b9fd54913ce2e45b8acdd34f5b44f95419db2.
//
// Solidity: event UnstakeRequested(address indexed provider, uint256 amount, uint256 unlockTime, uint256 requestIndex)
func (_BunkerStaking *BunkerStakingFilterer) WatchUnstakeRequested(opts *bind.WatchOpts, sink chan<- *BunkerStakingUnstakeRequested, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "UnstakeRequested", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingUnstakeRequested)
				if err := _BunkerStaking.contract.UnpackLog(event, "UnstakeRequested", log); err != nil {
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

// ParseUnstakeRequested is a log parse operation binding the contract event 0x6930caaa0f0843978bf16992d58b9fd54913ce2e45b8acdd34f5b44f95419db2.
//
// Solidity: event UnstakeRequested(address indexed provider, uint256 amount, uint256 unlockTime, uint256 requestIndex)
func (_BunkerStaking *BunkerStakingFilterer) ParseUnstakeRequested(log types.Log) (*BunkerStakingUnstakeRequested, error) {
	event := new(BunkerStakingUnstakeRequested)
	if err := _BunkerStaking.contract.UnpackLog(event, "UnstakeRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingVestedRewardsClaimedIterator is returned from FilterVestedRewardsClaimed and is used to iterate over the raw logs and unpacked data for VestedRewardsClaimed events raised by the BunkerStaking contract.
type BunkerStakingVestedRewardsClaimedIterator struct {
	Event *BunkerStakingVestedRewardsClaimed // Event containing the contract specifics and raw log

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
func (it *BunkerStakingVestedRewardsClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingVestedRewardsClaimed)
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
		it.Event = new(BunkerStakingVestedRewardsClaimed)
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
func (it *BunkerStakingVestedRewardsClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingVestedRewardsClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingVestedRewardsClaimed represents a VestedRewardsClaimed event raised by the BunkerStaking contract.
type BunkerStakingVestedRewardsClaimed struct {
	Provider common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterVestedRewardsClaimed is a free log retrieval operation binding the contract event 0xf295d7b1dc83525860d9a2626877f603a897f06bb6533c5a1ddffa679868dc98.
//
// Solidity: event VestedRewardsClaimed(address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) FilterVestedRewardsClaimed(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingVestedRewardsClaimedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "VestedRewardsClaimed", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingVestedRewardsClaimedIterator{contract: _BunkerStaking.contract, event: "VestedRewardsClaimed", logs: logs, sub: sub}, nil
}

// WatchVestedRewardsClaimed is a free log subscription operation binding the contract event 0xf295d7b1dc83525860d9a2626877f603a897f06bb6533c5a1ddffa679868dc98.
//
// Solidity: event VestedRewardsClaimed(address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) WatchVestedRewardsClaimed(opts *bind.WatchOpts, sink chan<- *BunkerStakingVestedRewardsClaimed, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "VestedRewardsClaimed", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingVestedRewardsClaimed)
				if err := _BunkerStaking.contract.UnpackLog(event, "VestedRewardsClaimed", log); err != nil {
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

// ParseVestedRewardsClaimed is a log parse operation binding the contract event 0xf295d7b1dc83525860d9a2626877f603a897f06bb6533c5a1ddffa679868dc98.
//
// Solidity: event VestedRewardsClaimed(address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) ParseVestedRewardsClaimed(log types.Log) (*BunkerStakingVestedRewardsClaimed, error) {
	event := new(BunkerStakingVestedRewardsClaimed)
	if err := _BunkerStaking.contract.UnpackLog(event, "VestedRewardsClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingVestedRewardsForfeitedIterator is returned from FilterVestedRewardsForfeited and is used to iterate over the raw logs and unpacked data for VestedRewardsForfeited events raised by the BunkerStaking contract.
type BunkerStakingVestedRewardsForfeitedIterator struct {
	Event *BunkerStakingVestedRewardsForfeited // Event containing the contract specifics and raw log

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
func (it *BunkerStakingVestedRewardsForfeitedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingVestedRewardsForfeited)
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
		it.Event = new(BunkerStakingVestedRewardsForfeited)
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
func (it *BunkerStakingVestedRewardsForfeitedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingVestedRewardsForfeitedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingVestedRewardsForfeited represents a VestedRewardsForfeited event raised by the BunkerStaking contract.
type BunkerStakingVestedRewardsForfeited struct {
	Provider common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterVestedRewardsForfeited is a free log retrieval operation binding the contract event 0xec349f79d94fe65197ec13d98c43134830b550972fdd1990702f29d62c17aa7c.
//
// Solidity: event VestedRewardsForfeited(address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) FilterVestedRewardsForfeited(opts *bind.FilterOpts, provider []common.Address) (*BunkerStakingVestedRewardsForfeitedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "VestedRewardsForfeited", providerRule)
	if err != nil {
		return nil, err
	}
	return &BunkerStakingVestedRewardsForfeitedIterator{contract: _BunkerStaking.contract, event: "VestedRewardsForfeited", logs: logs, sub: sub}, nil
}

// WatchVestedRewardsForfeited is a free log subscription operation binding the contract event 0xec349f79d94fe65197ec13d98c43134830b550972fdd1990702f29d62c17aa7c.
//
// Solidity: event VestedRewardsForfeited(address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) WatchVestedRewardsForfeited(opts *bind.WatchOpts, sink chan<- *BunkerStakingVestedRewardsForfeited, provider []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "VestedRewardsForfeited", providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingVestedRewardsForfeited)
				if err := _BunkerStaking.contract.UnpackLog(event, "VestedRewardsForfeited", log); err != nil {
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

// ParseVestedRewardsForfeited is a log parse operation binding the contract event 0xec349f79d94fe65197ec13d98c43134830b550972fdd1990702f29d62c17aa7c.
//
// Solidity: event VestedRewardsForfeited(address indexed provider, uint256 amount)
func (_BunkerStaking *BunkerStakingFilterer) ParseVestedRewardsForfeited(log types.Log) (*BunkerStakingVestedRewardsForfeited, error) {
	event := new(BunkerStakingVestedRewardsForfeited)
	if err := _BunkerStaking.contract.UnpackLog(event, "VestedRewardsForfeited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BunkerStakingVestingParamsUpdatedIterator is returned from FilterVestingParamsUpdated and is used to iterate over the raw logs and unpacked data for VestingParamsUpdated events raised by the BunkerStaking contract.
type BunkerStakingVestingParamsUpdatedIterator struct {
	Event *BunkerStakingVestingParamsUpdated // Event containing the contract specifics and raw log

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
func (it *BunkerStakingVestingParamsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BunkerStakingVestingParamsUpdated)
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
		it.Event = new(BunkerStakingVestingParamsUpdated)
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
func (it *BunkerStakingVestingParamsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BunkerStakingVestingParamsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BunkerStakingVestingParamsUpdated represents a VestingParamsUpdated event raised by the BunkerStaking contract.
type BunkerStakingVestingParamsUpdated struct {
	VestingPeriod       *big.Int
	ImmediateReleaseBps *big.Int
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterVestingParamsUpdated is a free log retrieval operation binding the contract event 0xf9933d06956f6b5949e6bca44fbed2a305cbd0eeb438eb592eff60a65d156bf5.
//
// Solidity: event VestingParamsUpdated(uint256 vestingPeriod, uint256 immediateReleaseBps)
func (_BunkerStaking *BunkerStakingFilterer) FilterVestingParamsUpdated(opts *bind.FilterOpts) (*BunkerStakingVestingParamsUpdatedIterator, error) {

	logs, sub, err := _BunkerStaking.contract.FilterLogs(opts, "VestingParamsUpdated")
	if err != nil {
		return nil, err
	}
	return &BunkerStakingVestingParamsUpdatedIterator{contract: _BunkerStaking.contract, event: "VestingParamsUpdated", logs: logs, sub: sub}, nil
}

// WatchVestingParamsUpdated is a free log subscription operation binding the contract event 0xf9933d06956f6b5949e6bca44fbed2a305cbd0eeb438eb592eff60a65d156bf5.
//
// Solidity: event VestingParamsUpdated(uint256 vestingPeriod, uint256 immediateReleaseBps)
func (_BunkerStaking *BunkerStakingFilterer) WatchVestingParamsUpdated(opts *bind.WatchOpts, sink chan<- *BunkerStakingVestingParamsUpdated) (event.Subscription, error) {

	logs, sub, err := _BunkerStaking.contract.WatchLogs(opts, "VestingParamsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BunkerStakingVestingParamsUpdated)
				if err := _BunkerStaking.contract.UnpackLog(event, "VestingParamsUpdated", log); err != nil {
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

// ParseVestingParamsUpdated is a log parse operation binding the contract event 0xf9933d06956f6b5949e6bca44fbed2a305cbd0eeb438eb592eff60a65d156bf5.
//
// Solidity: event VestingParamsUpdated(uint256 vestingPeriod, uint256 immediateReleaseBps)
func (_BunkerStaking *BunkerStakingFilterer) ParseVestingParamsUpdated(log types.Log) (*BunkerStakingVestingParamsUpdated, error) {
	event := new(BunkerStakingVestingParamsUpdated)
	if err := _BunkerStaking.contract.UnpackLog(event, "VestingParamsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
