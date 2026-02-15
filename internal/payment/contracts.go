package payment

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	pkgtypes "github.com/moltbunker/moltbunker/pkg/types"
)

// Contract ABIs for Moltbunker smart contracts on Base network
// These ABIs define the interface for interacting with deployed contracts

// BunkerTokenABI is the ABI for the BUNKER ERC20 token contract
const BunkerTokenABI = `[
	{
		"constant": true,
		"inputs": [{"name": "account", "type": "address"}],
		"name": "balanceOf",
		"outputs": [{"name": "", "type": "uint256"}],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "spender", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "approve",
		"outputs": [{"name": "", "type": "bool"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{"name": "owner", "type": "address"},
			{"name": "spender", "type": "address"}
		],
		"name": "allowance",
		"outputs": [{"name": "", "type": "uint256"}],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "to", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "transfer",
		"outputs": [{"name": "", "type": "bool"}],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "from", "type": "address"},
			{"name": "to", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "transferFrom",
		"outputs": [{"name": "", "type": "bool"}],
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "from", "type": "address"},
			{"indexed": true, "name": "to", "type": "address"},
			{"indexed": false, "name": "value", "type": "uint256"}
		],
		"name": "Transfer",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "owner", "type": "address"},
			{"indexed": true, "name": "spender", "type": "address"},
			{"indexed": false, "name": "value", "type": "uint256"}
		],
		"name": "Approval",
		"type": "event"
	}
]`

// StakingContractABI is the ABI for the BunkerStaking contract
const StakingContractABI = `[
	{
		"inputs": [{"name": "amount", "type": "uint256"}],
		"name": "stake",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "amount", "type": "uint256"}],
		"name": "requestUnstake",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "requestIndex", "type": "uint256"}],
		"name": "completeUnstake",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getProviderInfo",
		"outputs": [
			{"name": "info", "type": "tuple", "components": [
				{"name": "stakedAmount", "type": "uint128"},
				{"name": "totalUnbonding", "type": "uint128"},
				{"name": "beneficiary", "type": "address"},
				{"name": "registeredAt", "type": "uint48"},
				{"name": "active", "type": "bool"},
				{"name": "nodeId", "type": "bytes32"},
				{"name": "region", "type": "bytes32"},
				{"name": "capabilities", "type": "uint64"},
				{"name": "frozen", "type": "bool"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "isActiveProvider",
		"outputs": [{"name": "active", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getStake",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getTier",
		"outputs": [{"name": "tier", "type": "uint8"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "stakeAmount", "type": "uint256"}],
		"name": "getTierForAmount",
		"outputs": [{"name": "tier", "type": "uint8"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "totalStaked",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getUnstakeQueueLength",
		"outputs": [{"name": "count", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "index", "type": "uint256"}
		],
		"name": "getUnstakeRequest",
		"outputs": [
			{"name": "request", "type": "tuple", "components": [
				{"name": "amount", "type": "uint128"},
				{"name": "unlockTime", "type": "uint48"},
				{"name": "completed", "type": "bool"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getSlashableBalance",
		"outputs": [{"name": "total", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "slashImmediate",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "slash",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "totalStake", "type": "uint256"},
			{"indexed": false, "name": "tier", "type": "uint8"}
		],
		"name": "Staked",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "unlockTime", "type": "uint256"},
			{"indexed": false, "name": "requestIndex", "type": "uint256"}
		],
		"name": "UnstakeRequested",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "requestIndex", "type": "uint256"}
		],
		"name": "UnstakeCompleted",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "totalSlashed", "type": "uint256"},
			{"indexed": false, "name": "burnedAmount", "type": "uint256"},
			{"indexed": false, "name": "treasuryAmount", "type": "uint256"}
		],
		"name": "Slashed",
		"type": "event"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "earned",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "claimRewards",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "reward", "type": "uint256"}],
		"name": "notifyRewardAmount",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "_rewardsDuration", "type": "uint256"}],
		"name": "setRewardsDuration",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "lastTimeRewardApplicable",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "rewardPerToken",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "rewardRate",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "periodFinish",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "rewardsDuration",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "amount", "type": "uint256"},
			{"name": "reason", "type": "string"}
		],
		"name": "proposeSlash",
		"outputs": [{"name": "proposalId", "type": "uint256"}],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "proposalId", "type": "uint256"}],
		"name": "executeSlash",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "proposalId", "type": "uint256"}],
		"name": "appealSlash",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "proposalId", "type": "uint256"},
			{"name": "uphold", "type": "bool"}
		],
		"name": "resolveAppeal",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "proposalId", "type": "uint256"}],
		"name": "getSlashProposal",
		"outputs": [
			{"name": "proposal", "type": "tuple", "components": [
				{"name": "provider", "type": "address"},
				{"name": "amount", "type": "uint256"},
				{"name": "reason", "type": "string"},
				{"name": "proposedAt", "type": "uint256"},
				{"name": "executed", "type": "bool"},
				{"name": "appealed", "type": "bool"},
				{"name": "resolved", "type": "bool"},
				{"name": "slashReason", "type": "uint8"},
				{"name": "appealWindowSnapshot", "type": "uint256"},
				{"name": "slashBurnBpsSnapshot", "type": "uint16"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "newBeneficiary", "type": "address"}],
		"name": "initiateBeneficiaryChange",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "executeBeneficiaryChange",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "RewardClaimed",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "reward", "type": "uint256"},
			{"indexed": false, "name": "rewardRate", "type": "uint256"},
			{"indexed": false, "name": "periodFinish", "type": "uint256"}
		],
		"name": "RewardEpochStarted",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "proposalId", "type": "uint256"},
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "reason", "type": "string"}
		],
		"name": "SlashProposed",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "proposalId", "type": "uint256"},
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "SlashProposalExecuted",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "proposalId", "type": "uint256"},
			{"indexed": true, "name": "provider", "type": "address"}
		],
		"name": "SlashAppealed",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "proposalId", "type": "uint256"},
			{"indexed": false, "name": "upheld", "type": "bool"}
		],
		"name": "AppealResolved",
		"type": "event"
	},
	{
		"inputs": [
			{"name": "amount", "type": "uint256"},
			{"name": "nodeId", "type": "bytes32"},
			{"name": "region", "type": "bytes32"},
			{"name": "capabilities", "type": "uint64"}
		],
		"name": "stakeWithIdentity",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "nodeId", "type": "bytes32"},
			{"name": "region", "type": "bytes32"},
			{"name": "capabilities", "type": "uint64"}
		],
		"name": "updateIdentity",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "freezeProvider",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "unfreezeProvider",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "reason", "type": "uint8"}
		],
		"name": "proposeSlashByReason",
		"outputs": [{"name": "proposalId", "type": "uint256"}],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "reason", "type": "uint8"},
			{"name": "bps", "type": "uint16"}
		],
		"name": "setSlashPercentage",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "claimVestedRewards",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getClaimableVested",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "_vestingPeriod", "type": "uint256"},
			{"name": "_immediateReleaseBps", "type": "uint256"}
		],
		"name": "setVestingParams",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "hours_", "type": "uint256"}],
		"name": "reportComputeHours",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "_multiplier", "type": "uint256"}],
		"name": "setEmissionMultiplier",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "_maxRate", "type": "uint256"}],
		"name": "setMaxEmissionRate",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "nodeId", "type": "bytes32"}],
		"name": "nodeIdToProvider",
		"outputs": [{"name": "", "type": "address"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "nodeId", "type": "bytes32"},
			{"indexed": false, "name": "region", "type": "bytes32"},
			{"indexed": false, "name": "capabilities", "type": "uint64"}
		],
		"name": "ProviderIdentityUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": true, "name": "slasher", "type": "address"}
		],
		"name": "ProviderFrozen",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"}
		],
		"name": "ProviderUnfrozen",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "proposalId", "type": "uint256"},
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "reason", "type": "uint8"}
		],
		"name": "SlashProposedByReason",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "vestingEnd", "type": "uint256"},
			{"indexed": false, "name": "vestingIndex", "type": "uint256"}
		],
		"name": "RewardVested",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "VestedRewardsClaimed",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "VestedRewardsForfeited",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "hours_", "type": "uint256"},
			{"indexed": false, "name": "total", "type": "uint256"}
		],
		"name": "ComputeHoursReported",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "multiplier", "type": "uint256"}
		],
		"name": "EmissionMultiplierUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "maxRate", "type": "uint256"}
		],
		"name": "MaxEmissionRateUpdated",
		"type": "event"
	},
	{
		"inputs": [
			{"name": "tier", "type": "uint8"},
			{"name": "multiplierBps", "type": "uint16"}
		],
		"name": "setTierRewardMultiplier",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "newMax", "type": "uint16"}],
		"name": "setMaxTierMultiplierBps",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "newPeriod", "type": "uint256"}],
		"name": "setUnbondingPeriod",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "burnBps", "type": "uint16"},
			{"name": "treasuryBps", "type": "uint16"}
		],
		"name": "setSlashFeeSplit",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "newWindow", "type": "uint256"}],
		"name": "setAppealWindow",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "enabled", "type": "bool"}],
		"name": "setSlashingEnabled",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "slashingEnabled",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "tier", "type": "uint8"},
			{"indexed": false, "name": "multiplierBps", "type": "uint16"}
		],
		"name": "TierRewardMultiplierUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "newMax", "type": "uint16"}
		],
		"name": "MaxTierMultiplierUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "newPeriod", "type": "uint256"}
		],
		"name": "UnbondingPeriodUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "burnBps", "type": "uint16"},
			{"indexed": false, "name": "treasuryBps", "type": "uint16"}
		],
		"name": "SlashFeeSplitUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "newWindow", "type": "uint256"}
		],
		"name": "AppealWindowUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "enabled", "type": "bool"}
		],
		"name": "SlashingEnabledUpdated",
		"type": "event"
	}
]`

// EscrowContractABI is the ABI for the BunkerEscrow contract
const EscrowContractABI = `[
	{
		"inputs": [
			{"name": "amount", "type": "uint256"},
			{"name": "duration", "type": "uint256"}
		],
		"name": "createReservation",
		"outputs": [{"name": "reservationId", "type": "uint256"}],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "reservationId", "type": "uint256"},
			{"name": "providerAddrs", "type": "address[3]"}
		],
		"name": "selectProviders",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "reservationId", "type": "uint256"},
			{"name": "settledDuration", "type": "uint256"}
		],
		"name": "releasePayment",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "reservationId", "type": "uint256"}],
		"name": "finalizeReservation",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "reservationId", "type": "uint256"}],
		"name": "refund",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "reservationId", "type": "uint256"}],
		"name": "getReservation",
		"outputs": [
			{"name": "", "type": "tuple", "components": [
				{"name": "requester", "type": "address"},
				{"name": "totalAmount", "type": "uint128"},
				{"name": "releasedAmount", "type": "uint128"},
				{"name": "duration", "type": "uint48"},
				{"name": "startTime", "type": "uint48"},
				{"name": "status", "type": "uint8"},
				{"name": "providers", "type": "address[3]"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "reservationId", "type": "uint256"}],
		"name": "getProviders",
		"outputs": [{"name": "", "type": "address[3]"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "amount", "type": "uint256"}],
		"name": "calculateProtocolFee",
		"outputs": [{"name": "fee", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "reservationId", "type": "uint256"},
			{"indexed": true, "name": "requester", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "duration", "type": "uint256"}
		],
		"name": "ReservationCreated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "reservationId", "type": "uint256"},
			{"indexed": false, "name": "provider0", "type": "address"},
			{"indexed": false, "name": "provider1", "type": "address"},
			{"indexed": false, "name": "provider2", "type": "address"}
		],
		"name": "ProvidersSelected",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "reservationId", "type": "uint256"},
			{"indexed": false, "name": "grossAmount", "type": "uint256"},
			{"indexed": false, "name": "netToProviders", "type": "uint256"},
			{"indexed": false, "name": "protocolFee", "type": "uint256"},
			{"indexed": false, "name": "burnedAmount", "type": "uint256"},
			{"indexed": false, "name": "treasuryAmount", "type": "uint256"}
		],
		"name": "PaymentReleased",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "reservationId", "type": "uint256"}
		],
		"name": "ReservationFinalized",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "reservationId", "type": "uint256"},
			{"indexed": true, "name": "requester", "type": "address"},
			{"indexed": false, "name": "refundAmount", "type": "uint256"}
		],
		"name": "Refunded",
		"type": "event"
	},
	{
		"inputs": [
			{"name": "reservationId", "type": "uint256"},
			{"name": "additionalAmount", "type": "uint256"}
		],
		"name": "increaseDeposit",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "reservationId", "type": "uint256"}],
		"name": "claim",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "thresholdBps", "type": "uint16"}],
		"name": "setLowBalanceThreshold",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "_stakingContract", "type": "address"}],
		"name": "setStakingContract",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "reservationId", "type": "uint256"},
			{"indexed": true, "name": "requester", "type": "address"},
			{"indexed": false, "name": "additionalAmount", "type": "uint256"},
			{"indexed": false, "name": "newTotal", "type": "uint256"}
		],
		"name": "DepositIncreased",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "reservationId", "type": "uint256"},
			{"indexed": false, "name": "remaining", "type": "uint256"},
			{"indexed": false, "name": "threshold", "type": "uint256"}
		],
		"name": "LowBalance",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "oldContract", "type": "address"},
			{"indexed": true, "name": "newContract", "type": "address"}
		],
		"name": "StakingContractUpdated",
		"type": "event"
	},
	{
		"inputs": [
			{"name": "burnBps", "type": "uint16"},
			{"name": "treasuryBps", "type": "uint16"}
		],
		"name": "setFeeSplit",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "burnBps", "type": "uint16"},
			{"indexed": false, "name": "treasuryBps", "type": "uint16"}
		],
		"name": "FeeSplitUpdated",
		"type": "event"
	}
]`

// SlashingContractABI is the ABI for the slashing/dispute functions on BunkerStaking.
// The slashing system is built into BunkerStaking.sol; this ABI subset covers the
// proposeSlash/appealSlash/resolveAppeal workflow and immediate slashing functions.
const SlashingContractABI = `[
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "amount", "type": "uint256"},
			{"name": "reason", "type": "string"}
		],
		"name": "proposeSlash",
		"outputs": [{"name": "proposalId", "type": "uint256"}],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "proposalId", "type": "uint256"}],
		"name": "appealSlash",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "proposalId", "type": "uint256"},
			{"name": "uphold", "type": "bool"}
		],
		"name": "resolveAppeal",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "proposalId", "type": "uint256"}],
		"name": "executeSlash",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "slashImmediate",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "slash",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "proposalId", "type": "uint256"}],
		"name": "getSlashProposal",
		"outputs": [
			{"name": "proposal", "type": "tuple", "components": [
				{"name": "provider", "type": "address"},
				{"name": "amount", "type": "uint256"},
				{"name": "reason", "type": "string"},
				{"name": "proposedAt", "type": "uint256"},
				{"name": "executed", "type": "bool"},
				{"name": "appealed", "type": "bool"},
				{"name": "resolved", "type": "bool"},
				{"name": "slashReason", "type": "uint8"},
				{"name": "appealWindowSnapshot", "type": "uint256"},
				{"name": "slashBurnBpsSnapshot", "type": "uint16"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getSlashableBalance",
		"outputs": [{"name": "total", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "proposalId", "type": "uint256"},
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "reason", "type": "string"}
		],
		"name": "SlashProposed",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "totalSlashed", "type": "uint256"},
			{"indexed": false, "name": "burnedAmount", "type": "uint256"},
			{"indexed": false, "name": "treasuryAmount", "type": "uint256"}
		],
		"name": "Slashed",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "proposalId", "type": "uint256"},
			{"indexed": false, "name": "upheld", "type": "bool"}
		],
		"name": "AppealResolved",
		"type": "event"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "reason", "type": "uint8"}
		],
		"name": "proposeSlashByReason",
		"outputs": [{"name": "proposalId", "type": "uint256"}],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "reason", "type": "uint8"},
			{"name": "bps", "type": "uint16"}
		],
		"name": "setSlashPercentage",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "proposalId", "type": "uint256"},
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "reason", "type": "uint8"}
		],
		"name": "SlashProposedByReason",
		"type": "event"
	}
]`

// ViolationReason represents reasons for slashing.
// Must match Solidity: enum SlashReason { None, Downtime, JobAbandonment, SecurityViolation, Fraud, SLAViolation, DataLoss }
type ViolationReason uint8

const (
	ViolationNone              ViolationReason = iota // 0
	ViolationDowntime                                 // 1
	ViolationJobAbandonment                           // 2
	ViolationSecurityViolation                        // 3
	ViolationFraud                                    // 4
	ViolationSLAViolation                             // 5
	ViolationDataLoss                                 // 6
)

func (v ViolationReason) String() string {
	switch v {
	case ViolationNone:
		return "none"
	case ViolationDowntime:
		return "downtime"
	case ViolationJobAbandonment:
		return "job_abandonment"
	case ViolationSecurityViolation:
		return "security_violation"
	case ViolationFraud:
		return "fraud"
	case ViolationSLAViolation:
		return "sla_violation"
	case ViolationDataLoss:
		return "data_loss"
	default:
		return "unknown"
	}
}

// ViolationReasonFromSlashingViolation converts the pkg/types SlashingViolation
// string type to the ABI-compatible ViolationReason uint8.
// This bridges the config-level string constants with on-chain enum values.
func ViolationReasonFromSlashingViolation(sv pkgtypes.SlashingViolation) ViolationReason {
	switch sv {
	case pkgtypes.SlashingViolationDowntime:
		return ViolationDowntime
	case pkgtypes.SlashingViolationJobAbandonment:
		return ViolationJobAbandonment
	case pkgtypes.SlashingViolationSecurityBreach:
		return ViolationSecurityViolation
	case pkgtypes.SlashingViolationResourceFraud:
		return ViolationFraud
	case pkgtypes.SlashingViolationHealthFailure:
		return ViolationSLAViolation // health failures map to SLA violations
	case pkgtypes.SlashingViolationReplicaMismatch:
		return ViolationSLAViolation // replica issues map to SLA violations
	case pkgtypes.SlashingViolationMalicious:
		return ViolationFraud // malicious activity maps to fraud
	default:
		return ViolationNone
	}
}

// EscrowState represents the state of an escrow (matches Solidity Status enum)
type EscrowState uint8

const (
	EscrowStateNone      EscrowState = iota // 0 - default
	EscrowStateCreated                      // 1 - reservation created, providers not yet selected
	EscrowStateActive                       // 2 - providers selected, job running
	EscrowStateCompleted                    // 3 - finalized
	EscrowStateRefunded                     // 4 - refunded to requester
	EscrowStateDisputed                     // 5 - under dispute
)

func (s EscrowState) String() string {
	switch s {
	case EscrowStateNone:
		return "none"
	case EscrowStateCreated:
		return "created"
	case EscrowStateActive:
		return "active"
	case EscrowStateCompleted:
		return "completed"
	case EscrowStateRefunded:
		return "refunded"
	case EscrowStateDisputed:
		return "disputed"
	default:
		return "unknown"
	}
}

// DisputeState represents the state of a dispute
type DisputeState uint8

const (
	DisputeStatePending DisputeState = iota
	DisputeStateDefenseSubmitted
	DisputeStateResolved
	DisputeStateExpired
)

func (s DisputeState) String() string {
	switch s {
	case DisputeStatePending:
		return "pending"
	case DisputeStateDefenseSubmitted:
		return "defense_submitted"
	case DisputeStateResolved:
		return "resolved"
	case DisputeStateExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// ─── Governance Contract ABIs ────────────────────────────────────────────────

// DelegationContractABI is the ABI for the BunkerDelegation contract
const DelegationContractABI = `[
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "delegate",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "amount", "type": "uint256"}],
		"name": "requestUndelegate",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "index", "type": "uint256"}],
		"name": "completeUndelegate",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "delegator", "type": "address"}],
		"name": "getDelegation",
		"outputs": [
			{"name": "", "type": "tuple", "components": [
				{"name": "provider", "type": "address"},
				{"name": "amount", "type": "uint256"},
				{"name": "rewardDebt", "type": "uint256"},
				{"name": "delegatedAt", "type": "uint48"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getProviderConfig",
		"outputs": [
			{"name": "", "type": "tuple", "components": [
				{"name": "rewardCutBps", "type": "uint16"},
				{"name": "feeShareBps", "type": "uint16"},
				{"name": "acceptDelegations", "type": "bool"},
				{"name": "pendingRewardCutBps", "type": "uint16"},
				{"name": "rewardCutEffectiveAt", "type": "uint48"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getTotalDelegatedTo",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "rewardCutBps", "type": "uint16"},
			{"name": "feeShareBps", "type": "uint16"}
		],
		"name": "setDelegationConfig",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "accept", "type": "bool"}],
		"name": "toggleAcceptDelegations",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "totalDelegated",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "delegator", "type": "address"},
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "Delegated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "delegator", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "unlockTime", "type": "uint256"}
		],
		"name": "UndelegateRequested",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "delegator", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "UndelegateCompleted",
		"type": "event"
	},
	{
		"inputs": [{"name": "newPeriod", "type": "uint256"}],
		"name": "setUnbondingPeriod",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "newMax", "type": "uint16"}],
		"name": "setMaxRewardCutBps",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "newPeriod", "type": "uint256"}
		],
		"name": "UnbondingPeriodUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "newMax", "type": "uint16"}
		],
		"name": "MaxRewardCutUpdated",
		"type": "event"
	}
]`

// ReputationContractABI is the ABI for the BunkerReputation contract
const ReputationContractABI = `[
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getScore",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getTier",
		"outputs": [{"name": "", "type": "uint8"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "isEligibleForJobs",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getReputation",
		"outputs": [
			{"name": "", "type": "tuple", "components": [
				{"name": "score", "type": "uint256"},
				{"name": "jobsCompleted", "type": "uint256"},
				{"name": "jobsFailed", "type": "uint256"},
				{"name": "slashEvents", "type": "uint256"},
				{"name": "lastDecay", "type": "uint48"},
				{"name": "registered", "type": "bool"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "registerProvider",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "recordJobCompleted",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "recordJobFailed",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "recordSlashEvent",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "delta", "type": "int256"},
			{"name": "reason", "type": "string"}
		],
		"name": "recordEvent",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "applyDecay",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "newScore", "type": "uint256"},
			{"indexed": false, "name": "delta", "type": "int256"},
			{"indexed": false, "name": "reason", "type": "string"}
		],
		"name": "ScoreUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "oldTier", "type": "uint8"},
			{"indexed": false, "name": "newTier", "type": "uint8"}
		],
		"name": "TierChanged",
		"type": "event"
	},
	{
		"inputs": [
			{"name": "probation", "type": "uint16"},
			{"name": "standard", "type": "uint16"},
			{"name": "trusted", "type": "uint16"},
			{"name": "elite", "type": "uint16"}
		],
		"name": "setTierThresholds",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "rate", "type": "uint16"},
			{"name": "floor", "type": "uint16"}
		],
		"name": "setDecayParams",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "minScore", "type": "uint16"}],
		"name": "setMinScoreForJobs",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "maxDelta", "type": "int16"}],
		"name": "setMaxCustomDelta",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "probation", "type": "uint16"},
			{"indexed": false, "name": "standard", "type": "uint16"},
			{"indexed": false, "name": "trusted", "type": "uint16"},
			{"indexed": false, "name": "elite", "type": "uint16"}
		],
		"name": "TierThresholdsUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "rate", "type": "uint16"},
			{"indexed": false, "name": "floor", "type": "uint16"}
		],
		"name": "DecayParamsUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "minScore", "type": "uint16"}
		],
		"name": "MinScoreForJobsUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "maxDelta", "type": "int16"}
		],
		"name": "MaxCustomDeltaUpdated",
		"type": "event"
	}
]`

// VerificationContractABI is the ABI for the BunkerVerification contract
const VerificationContractABI = `[
	{
		"inputs": [{"name": "hash", "type": "bytes32"}],
		"name": "submitAttestation",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "checkMissedAttestations",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "proof", "type": "bytes"}
		],
		"name": "challengeAttestation",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getAttestation",
		"outputs": [
			{"name": "", "type": "tuple", "components": [
				{"name": "lastHash", "type": "bytes32"},
				{"name": "lastTime", "type": "uint48"},
				{"name": "missedCount", "type": "uint16"},
				{"name": "suspended", "type": "bool"},
				{"name": "suspendedUntil", "type": "uint48"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "isAttestationCurrent",
		"outputs": [{"name": "", "type": "bool"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "hash", "type": "bytes32"},
			{"indexed": false, "name": "timestamp", "type": "uint256"}
		],
		"name": "AttestationSubmitted",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "missedCount", "type": "uint16"}
		],
		"name": "AttestationMissed",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "suspendedUntil", "type": "uint256"}
		],
		"name": "ProviderSuspended",
		"type": "event"
	},
	{
		"inputs": [{"name": "newCooldown", "type": "uint256"}],
		"name": "setReinstatementCooldown",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "newCooldown", "type": "uint256"}
		],
		"name": "ReinstatementCooldownUpdated",
		"type": "event"
	}
]`

// PricingContractABI is the ABI for the BunkerPricing contract
const PricingContractABI = `[
	{
		"inputs": [
			{"name": "req", "type": "tuple", "components": [
				{"name": "cpuCores", "type": "uint16"},
				{"name": "memoryGB", "type": "uint16"},
				{"name": "storageGB", "type": "uint16"},
				{"name": "bandwidthMbps", "type": "uint16"},
				{"name": "durationHours", "type": "uint32"},
				{"name": "gpuCount", "type": "uint8"}
			]}
		],
		"name": "calculateCost",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "req", "type": "tuple", "components": [
				{"name": "cpuCores", "type": "uint16"},
				{"name": "memoryGB", "type": "uint16"},
				{"name": "storageGB", "type": "uint16"},
				{"name": "bandwidthMbps", "type": "uint16"},
				{"name": "durationHours", "type": "uint32"},
				{"name": "gpuCount", "type": "uint8"}
			]}
		],
		"name": "calculateProviderCost",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "getTokenPrice",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "getPrices",
		"outputs": [
			{"name": "", "type": "tuple", "components": [
				{"name": "cpuPerCoreHour", "type": "uint256"},
				{"name": "memoryPerGBHour", "type": "uint256"},
				{"name": "storagePerGBHour", "type": "uint256"},
				{"name": "bandwidthPerGBHour", "type": "uint256"},
				{"name": "gpuPerHour", "type": "uint256"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "getMultipliers",
		"outputs": [
			{"name": "", "type": "tuple", "components": [
				{"name": "demandMultiplier", "type": "uint256"},
				{"name": "regionMultiplier", "type": "uint256"},
				{"name": "tierDiscount", "type": "uint256"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getProviderPrices",
		"outputs": [
			{"name": "", "type": "tuple", "components": [
				{"name": "cpuMultiplier", "type": "uint256"},
				{"name": "memoryMultiplier", "type": "uint256"},
				{"name": "storageMultiplier", "type": "uint256"},
				{"name": "minPrice", "type": "uint256"},
				{"name": "maxPrice", "type": "uint256"}
			]}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "tokenPrice", "type": "uint256"},
			{"indexed": false, "name": "timestamp", "type": "uint256"}
		],
		"name": "PriceUpdated",
		"type": "event"
	},
	{
		"inputs": [{"name": "newMax", "type": "uint256"}],
		"name": "setMaxMultiplierBps",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "newMin", "type": "uint256"},
			{"name": "newMax", "type": "uint256"}
		],
		"name": "setStaleThresholdBounds",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "newMax", "type": "uint256"}
		],
		"name": "MaxMultiplierUpdated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": false, "name": "newMin", "type": "uint256"},
			{"indexed": false, "name": "newMax", "type": "uint256"}
		],
		"name": "StaleThresholdBoundsUpdated",
		"type": "event"
	}
]`

// ─── Governance Go Types ─────────────────────────────────────────────────────

// ReputationTier represents a provider's reputation tier.
// Matches Solidity: enum Tier { Untrusted, Newcomer, Reliable, Established, Elite }
type ReputationTier uint8

const (
	ReputationTierUntrusted   ReputationTier = iota // 0
	ReputationTierNewcomer                          // 1
	ReputationTierReliable                          // 2
	ReputationTierEstablished                       // 3
	ReputationTierElite                             // 4
)

func (t ReputationTier) String() string {
	switch t {
	case ReputationTierUntrusted:
		return "untrusted"
	case ReputationTierNewcomer:
		return "newcomer"
	case ReputationTierReliable:
		return "reliable"
	case ReputationTierEstablished:
		return "established"
	case ReputationTierElite:
		return "elite"
	default:
		return "unknown"
	}
}

// DelegationData represents a delegation record from BunkerDelegation.
type DelegationData struct {
	Provider    common.Address
	Amount      *big.Int
	RewardDebt  *big.Int
	DelegatedAt time.Time
}

// ProviderDelegationConfigData represents a provider's delegation configuration.
type ProviderDelegationConfigData struct {
	RewardCutBps           uint16
	FeeShareBps            uint16
	AcceptDelegations      bool
	PendingRewardCutBps    uint16
	RewardCutEffectiveAt   time.Time
}

// UnbondingRequestData represents a pending undelegation request.
type UnbondingRequestData struct {
	Amount     *big.Int
	UnlockTime time.Time
	Completed  bool
}

// ReputationDataOnChain represents a provider's on-chain reputation data.
type ReputationDataOnChain struct {
	Score         *big.Int
	JobsCompleted *big.Int
	JobsFailed    *big.Int
	SlashEvents   *big.Int
	LastDecay     time.Time
	Registered    bool
}

// AttestationData represents a provider's attestation state.
type AttestationData struct {
	LastHash       [32]byte
	LastTime       time.Time
	MissedCount    uint16
	Suspended      bool
	SuspendedUntil time.Time
}

// ResourcePricesData represents on-chain resource pricing.
type ResourcePricesData struct {
	CPUPerCoreHour       *big.Int
	MemoryPerGBHour      *big.Int
	StoragePerGBHour     *big.Int
	BandwidthPerGBHour   *big.Int
	GPUPerHour           *big.Int
}

// MultipliersData represents on-chain pricing multipliers.
type MultipliersData struct {
	DemandMultiplier *big.Int
	RegionMultiplier *big.Int
	TierDiscount     *big.Int
}

// ProviderPricingData represents a provider's custom pricing.
type ProviderPricingData struct {
	CPUMultiplier     *big.Int
	MemoryMultiplier  *big.Int
	StorageMultiplier *big.Int
	MinPrice          *big.Int
	MaxPrice          *big.Int
}

// ResourceRequestData represents a resource request for pricing calculation.
type ResourceRequestData struct {
	CPUCores      uint16
	MemoryGB      uint16
	StorageGB     uint16
	BandwidthMbps uint16
	DurationHours uint32
	GPUCount      uint8
}
