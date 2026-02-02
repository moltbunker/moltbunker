package payment

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

// ProviderRegistryABI is the ABI for the provider registry contract
const ProviderRegistryABI = `[
	{
		"constant": false,
		"inputs": [
			{"name": "resources", "type": "tuple", "components": [
				{"name": "cpuCores", "type": "uint32"},
				{"name": "memoryGB", "type": "uint32"},
				{"name": "storageGB", "type": "uint32"},
				{"name": "bandwidthMbps", "type": "uint32"},
				{"name": "gpuCount", "type": "uint8"},
				{"name": "gpuModel", "type": "string"}
			]},
			{"name": "region", "type": "string"},
			{"name": "endpoint", "type": "string"}
		],
		"name": "registerProvider",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [],
		"name": "unregisterProvider",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "resources", "type": "tuple", "components": [
				{"name": "cpuCores", "type": "uint32"},
				{"name": "memoryGB", "type": "uint32"},
				{"name": "storageGB", "type": "uint32"},
				{"name": "bandwidthMbps", "type": "uint32"},
				{"name": "gpuCount", "type": "uint8"},
				{"name": "gpuModel", "type": "string"}
			]}
		],
		"name": "updateResources",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getProvider",
		"outputs": [
			{"name": "registered", "type": "bool"},
			{"name": "staked", "type": "uint256"},
			{"name": "reputation", "type": "uint256"},
			{"name": "region", "type": "string"},
			{"name": "endpoint", "type": "string"}
		],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "isRegistered",
		"outputs": [{"name": "", "type": "bool"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "getActiveProviders",
		"outputs": [{"name": "", "type": "address[]"}],
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "region", "type": "string"}
		],
		"name": "ProviderRegistered",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"}
		],
		"name": "ProviderUnregistered",
		"type": "event"
	}
]`

// StakingContractABI is the ABI for the staking contract
const StakingContractABI = `[
	{
		"constant": false,
		"inputs": [{"name": "amount", "type": "uint256"}],
		"name": "stake",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [{"name": "amount", "type": "uint256"}],
		"name": "requestUnstake",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [],
		"name": "withdraw",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getStake",
		"outputs": [{"name": "", "type": "uint256"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getPendingUnstake",
		"outputs": [
			{"name": "amount", "type": "uint256"},
			{"name": "unlockTime", "type": "uint256"}
		],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getTier",
		"outputs": [{"name": "", "type": "uint8"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "tier", "type": "uint8"}],
		"name": "getTierMinStake",
		"outputs": [{"name": "", "type": "uint256"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "totalStaked",
		"outputs": [{"name": "", "type": "uint256"}],
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "totalStake", "type": "uint256"}
		],
		"name": "Staked",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "unlockTime", "type": "uint256"}
		],
		"name": "UnstakeRequested",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "Withdrawn",
		"type": "event"
	}
]`

// EscrowContractABI is the ABI for the payment escrow contract
const EscrowContractABI = `[
	{
		"constant": false,
		"inputs": [
			{"name": "jobId", "type": "bytes32"},
			{"name": "provider", "type": "address"},
			{"name": "amount", "type": "uint256"},
			{"name": "duration", "type": "uint256"}
		],
		"name": "createEscrow",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "jobId", "type": "bytes32"},
			{"name": "uptime", "type": "uint256"}
		],
		"name": "releasePayment",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [{"name": "jobId", "type": "bytes32"}],
		"name": "refund",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [{"name": "jobId", "type": "bytes32"}],
		"name": "finalizeEscrow",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "jobId", "type": "bytes32"}],
		"name": "getEscrow",
		"outputs": [
			{"name": "requester", "type": "address"},
			{"name": "provider", "type": "address"},
			{"name": "amount", "type": "uint256"},
			{"name": "released", "type": "uint256"},
			{"name": "duration", "type": "uint256"},
			{"name": "startTime", "type": "uint256"},
			{"name": "state", "type": "uint8"}
		],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "requester", "type": "address"}],
		"name": "getRequesterEscrows",
		"outputs": [{"name": "", "type": "bytes32[]"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getProviderEscrows",
		"outputs": [{"name": "", "type": "bytes32[]"}],
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "jobId", "type": "bytes32"},
			{"indexed": true, "name": "requester", "type": "address"},
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "EscrowCreated",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "jobId", "type": "bytes32"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "PaymentReleased",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "jobId", "type": "bytes32"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "Refunded",
		"type": "event"
	}
]`

// SlashingContractABI is the ABI for the slashing contract
const SlashingContractABI = `[
	{
		"constant": false,
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "jobId", "type": "bytes32"},
			{"name": "reason", "type": "uint8"},
			{"name": "evidence", "type": "bytes"}
		],
		"name": "reportViolation",
		"outputs": [{"name": "disputeId", "type": "bytes32"}],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "disputeId", "type": "bytes32"},
			{"name": "evidence", "type": "bytes"}
		],
		"name": "submitDefense",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "disputeId", "type": "bytes32"},
			{"name": "slashAmount", "type": "uint256"}
		],
		"name": "resolveDispute",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "provider", "type": "address"},
			{"name": "amount", "type": "uint256"},
			{"name": "reason", "type": "uint8"}
		],
		"name": "slash",
		"outputs": [],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "disputeId", "type": "bytes32"}],
		"name": "getDispute",
		"outputs": [
			{"name": "reporter", "type": "address"},
			{"name": "provider", "type": "address"},
			{"name": "jobId", "type": "bytes32"},
			{"name": "reason", "type": "uint8"},
			{"name": "state", "type": "uint8"},
			{"name": "timestamp", "type": "uint256"}
		],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "provider", "type": "address"}],
		"name": "getSlashingHistory",
		"outputs": [
			{"name": "totalSlashed", "type": "uint256"},
			{"name": "violations", "type": "uint256"}
		],
		"type": "function"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "disputeId", "type": "bytes32"},
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "reason", "type": "uint8"}
		],
		"name": "ViolationReported",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "provider", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"},
			{"indexed": false, "name": "reason", "type": "uint8"}
		],
		"name": "Slashed",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "disputeId", "type": "bytes32"},
			{"indexed": false, "name": "slashAmount", "type": "uint256"}
		],
		"name": "DisputeResolved",
		"type": "event"
	}
]`

// ViolationReason represents reasons for slashing
type ViolationReason uint8

const (
	ViolationNone ViolationReason = iota
	ViolationDowntime
	ViolationDataLoss
	ViolationSecurityBreach
	ViolationSLAViolation
	ViolationMaliciousBehavior
)

func (v ViolationReason) String() string {
	switch v {
	case ViolationNone:
		return "none"
	case ViolationDowntime:
		return "downtime"
	case ViolationDataLoss:
		return "data_loss"
	case ViolationSecurityBreach:
		return "security_breach"
	case ViolationSLAViolation:
		return "sla_violation"
	case ViolationMaliciousBehavior:
		return "malicious_behavior"
	default:
		return "unknown"
	}
}

// EscrowState represents the state of an escrow
type EscrowState uint8

const (
	EscrowStateActive EscrowState = iota
	EscrowStateCompleted
	EscrowStateRefunded
	EscrowStateDisputed
)

func (s EscrowState) String() string {
	switch s {
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
