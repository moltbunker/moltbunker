package types

import (
	"fmt"
	"math/big"
	"time"
)

// NodeRole defines the operational mode of a node
type NodeRole string

const (
	// NodeRoleProvider - Node provides compute resources and earns BUNKER
	NodeRoleProvider NodeRole = "provider"
	// NodeRoleRequester - Node submits jobs and pays BUNKER
	NodeRoleRequester NodeRole = "requester"
	// NodeRoleHybrid - Node can both provide and request (for testing/development)
	NodeRoleHybrid NodeRole = "hybrid"
)

// IsValid checks if the node role is valid
func (r NodeRole) IsValid() bool {
	switch r {
	case NodeRoleProvider, NodeRoleRequester, NodeRoleHybrid:
		return true
	default:
		return false
	}
}

// StakingTier represents provider staking levels
type StakingTier string

const (
	StakingTierStarter  StakingTier = "starter"
	StakingTierBronze   StakingTier = "bronze"
	StakingTierSilver   StakingTier = "silver"
	StakingTierGold     StakingTier = "gold"
	StakingTierPlatinum StakingTier = "platinum"
)

// StakingTierConfig defines requirements and limits for a staking tier
type StakingTierConfig struct {
	Name              StakingTier `yaml:"name" json:"name"`
	MinStake          *big.Int    `yaml:"-" json:"-"`                                     // Minimum stake in wei
	MinStakeString    string      `yaml:"min_stake" json:"min_stake"`                     // String representation for config
	MaxConcurrentJobs int         `yaml:"max_concurrent_jobs" json:"max_concurrent_jobs"` // Max jobs at once
	RewardMultiplier  float64     `yaml:"reward_multiplier" json:"reward_multiplier"`     // Earnings multiplier
	PriorityQueue     bool        `yaml:"priority_queue" json:"priority_queue"`           // Access to priority jobs
	Governance        bool        `yaml:"governance" json:"governance"`                   // Can vote on proposals
	RevenueShare      bool        `yaml:"revenue_share" json:"revenue_share"`             // Share of protocol revenue
}

// ParseMinStake parses the min stake string to big.Int
func (s *StakingTierConfig) ParseMinStake() error {
	if s.MinStakeString == "" {
		s.MinStake = big.NewInt(0)
		return nil
	}
	s.MinStake = new(big.Int)
	_, ok := s.MinStake.SetString(s.MinStakeString, 10)
	if !ok {
		return fmt.Errorf("invalid min_stake value: %s", s.MinStakeString)
	}
	return nil
}

// DefaultStakingTiers returns the default staking tier configuration
func DefaultStakingTiers() map[StakingTier]*StakingTierConfig {
	return map[StakingTier]*StakingTierConfig{
		StakingTierStarter: {
			Name:              StakingTierStarter,
			MinStakeString:    "1000000000000000000000000",      // 1,000,000 BUNKER (18 decimals)
			MaxConcurrentJobs: 3,
			RewardMultiplier:  1.0,
			PriorityQueue:     false,
			Governance:        false,
			RevenueShare:      false,
		},
		StakingTierBronze: {
			Name:              StakingTierBronze,
			MinStakeString:    "5000000000000000000000000",      // 5,000,000 BUNKER
			MaxConcurrentJobs: 10,
			RewardMultiplier:  1.0,
			PriorityQueue:     false,
			Governance:        false,
			RevenueShare:      false,
		},
		StakingTierSilver: {
			Name:              StakingTierSilver,
			MinStakeString:    "10000000000000000000000000",     // 10,000,000 BUNKER
			MaxConcurrentJobs: 50,
			RewardMultiplier:  1.05,
			PriorityQueue:     true,
			Governance:        false,
			RevenueShare:      false,
		},
		StakingTierGold: {
			Name:              StakingTierGold,
			MinStakeString:    "100000000000000000000000000",    // 100,000,000 BUNKER
			MaxConcurrentJobs: 200,
			RewardMultiplier:  1.1,
			PriorityQueue:     true,
			Governance:        true,
			RevenueShare:      false,
		},
		StakingTierPlatinum: {
			Name:              StakingTierPlatinum,
			MinStakeString:    "1000000000000000000000000000",   // 1,000,000,000 BUNKER
			MaxConcurrentJobs: -1,                               // Unlimited
			RewardMultiplier:  1.2,
			PriorityQueue:     true,
			Governance:        true,
			RevenueShare:      true,
		},
	}
}

// PricingConfig defines resource pricing (in BUNKER wei per unit)
type PricingConfig struct {
	CPUPerCoreHour     string `yaml:"cpu_per_core_hour" json:"cpu_per_core_hour"`         // BUNKER per CPU core-hour
	MemoryPerGBHour    string `yaml:"memory_per_gb_hour" json:"memory_per_gb_hour"`       // BUNKER per GB-hour
	StoragePerGBMonth  string `yaml:"storage_per_gb_month" json:"storage_per_gb_month"`   // BUNKER per GB-month
	NetworkPerGB       string `yaml:"network_per_gb" json:"network_per_gb"`               // BUNKER per GB egress
	GPUBasicPerHour    string `yaml:"gpu_basic_per_hour" json:"gpu_basic_per_hour"`       // BUNKER per basic GPU-hour
	GPUPremiumPerHour  string `yaml:"gpu_premium_per_hour" json:"gpu_premium_per_hour"`   // BUNKER per premium GPU-hour

	// Multipliers
	RedundancyMultiplier float64 `yaml:"redundancy_multiplier" json:"redundancy_multiplier"` // For 3x redundancy (default: 3.0)
	TorMultiplier        float64 `yaml:"tor_multiplier" json:"tor_multiplier"`               // For Tor-only mode
	PremiumSLAMultiplier float64 `yaml:"premium_sla_multiplier" json:"premium_sla_multiplier"` // For 99.99% SLA
	SpotMultiplier       float64 `yaml:"spot_multiplier" json:"spot_multiplier"`             // For preemptible instances

	// Minimums
	MinimumServiceHours   int `yaml:"minimum_service_hours" json:"minimum_service_hours"`     // Min billing for services
	MinimumJobMinutes     int `yaml:"minimum_job_minutes" json:"minimum_job_minutes"`         // Min billing for jobs
	MinimumFunctionMillis int `yaml:"minimum_function_millis" json:"minimum_function_millis"` // Min billing for functions
}

// DefaultPricingConfig returns the default pricing configuration
func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		// Base prices in BUNKER wei (18 decimals)
		// 0.50 BUNKER per CPU core-hour
		CPUPerCoreHour:    "500000000000000000",
		// 0.10 BUNKER per GB-hour
		MemoryPerGBHour:   "100000000000000000",
		// 0.05 BUNKER per GB-month
		StoragePerGBMonth: "50000000000000000",
		// 0.02 BUNKER per GB
		NetworkPerGB:      "20000000000000000",
		// 5.00 BUNKER per basic GPU-hour
		GPUBasicPerHour:   "5000000000000000000",
		// 15.00 BUNKER per premium GPU-hour
		GPUPremiumPerHour: "15000000000000000000",

		// Multipliers
		RedundancyMultiplier: 3.0,  // 3x for mandatory redundancy
		TorMultiplier:        1.2,  // 20% premium for Tor
		PremiumSLAMultiplier: 1.5,  // 50% premium for 99.99% SLA
		SpotMultiplier:       0.5,  // 50% discount for preemptible

		// Minimums
		MinimumServiceHours:   1,
		MinimumJobMinutes:     1,
		MinimumFunctionMillis: 100,
	}
}

// SlashingViolation represents types of slashable offenses
type SlashingViolation string

const (
	SlashingViolationDowntime        SlashingViolation = "extended_downtime"
	SlashingViolationJobAbandonment  SlashingViolation = "job_abandonment"
	SlashingViolationHealthFailure   SlashingViolation = "health_check_failure"
	SlashingViolationResourceFraud   SlashingViolation = "resource_fraud"
	SlashingViolationReplicaMismatch SlashingViolation = "replica_mismatch"
	SlashingViolationSecurityBreach  SlashingViolation = "security_violation"
	SlashingViolationMalicious       SlashingViolation = "malicious_activity"
)

// SlashingConfig defines slashing parameters for violations
type SlashingConfig struct {
	// Violation percentages (0-100)
	DowntimePerHourPercent      float64 `yaml:"downtime_per_hour_percent" json:"downtime_per_hour_percent"`           // Per hour of downtime
	DowntimeMaxPercent          float64 `yaml:"downtime_max_percent" json:"downtime_max_percent"`                     // Maximum for downtime
	JobAbandonmentPercent       float64 `yaml:"job_abandonment_percent" json:"job_abandonment_percent"`
	HealthFailurePercent        float64 `yaml:"health_failure_percent" json:"health_failure_percent"`
	ResourceFraudPercent        float64 `yaml:"resource_fraud_percent" json:"resource_fraud_percent"`
	ReplicaMismatchPercent      float64 `yaml:"replica_mismatch_percent" json:"replica_mismatch_percent"`
	SecurityViolationPercent    float64 `yaml:"security_violation_percent" json:"security_violation_percent"`
	MaliciousActivityPercent    float64 `yaml:"malicious_activity_percent" json:"malicious_activity_percent"`

	// Appeal windows (in hours)
	DowntimeAppealHours         int `yaml:"downtime_appeal_hours" json:"downtime_appeal_hours"`
	JobAbandonmentAppealHours   int `yaml:"job_abandonment_appeal_hours" json:"job_abandonment_appeal_hours"`
	HealthFailureAppealHours    int `yaml:"health_failure_appeal_hours" json:"health_failure_appeal_hours"`
	ResourceFraudAppealHours    int `yaml:"resource_fraud_appeal_hours" json:"resource_fraud_appeal_hours"`
	ReplicaMismatchAppealHours  int `yaml:"replica_mismatch_appeal_hours" json:"replica_mismatch_appeal_hours"`
	SecurityViolationAppealHours int `yaml:"security_violation_appeal_hours" json:"security_violation_appeal_hours"`
	MaliciousActivityAppealHours int `yaml:"malicious_activity_appeal_hours" json:"malicious_activity_appeal_hours"`

	// Distribution percentages (must sum to 100)
	BurnPercent      float64 `yaml:"burn_percent" json:"burn_percent"`           // Burned (deflationary)
	RequesterPercent float64 `yaml:"requester_percent" json:"requester_percent"` // To affected requester
	ReporterPercent  float64 `yaml:"reporter_percent" json:"reporter_percent"`   // To reporter (if community report)
}

// DefaultSlashingConfig returns the default slashing configuration
func DefaultSlashingConfig() *SlashingConfig {
	return &SlashingConfig{
		// Violation percentages
		DowntimePerHourPercent:   2.0,
		DowntimeMaxPercent:       20.0,
		JobAbandonmentPercent:    15.0,
		HealthFailurePercent:     10.0,
		ResourceFraudPercent:     30.0,
		ReplicaMismatchPercent:   40.0,
		SecurityViolationPercent: 75.0,
		MaliciousActivityPercent: 100.0,

		// Appeal windows
		DowntimeAppealHours:        0,  // No appeal for downtime
		JobAbandonmentAppealHours:  12,
		HealthFailureAppealHours:   24,
		ResourceFraudAppealHours:   24,
		ReplicaMismatchAppealHours: 24,
		SecurityViolationAppealHours: 48,
		MaliciousActivityAppealHours: 72,

		// Distribution (50% burn, 30% requester, 20% reporter)
		BurnPercent:      50.0,
		RequesterPercent: 30.0,
		ReporterPercent:  20.0,
	}
}

// ReputationConfig defines reputation scoring parameters
type ReputationConfig struct {
	// Score range
	MinScore     int `yaml:"min_score" json:"min_score"`
	MaxScore     int `yaml:"max_score" json:"max_score"`
	InitialScore int `yaml:"initial_score" json:"initial_score"`

	// Positive actions
	JobCompletedPoints        int `yaml:"job_completed_points" json:"job_completed_points"`
	JobEarlyCompletionPoints  int `yaml:"job_early_completion_points" json:"job_early_completion_points"`
	PerfectMonthPoints        int `yaml:"perfect_month_points" json:"perfect_month_points"`
	CommunityVerificationPoints int `yaml:"community_verification_points" json:"community_verification_points"`

	// Negative actions
	JobTimeoutPenalty        int `yaml:"job_timeout_penalty" json:"job_timeout_penalty"`
	HealthFailurePenalty     int `yaml:"health_failure_penalty" json:"health_failure_penalty"`
	ReplicaMismatchPenalty   int `yaml:"replica_mismatch_penalty" json:"replica_mismatch_penalty"`
	SlashingEventPenalty     int `yaml:"slashing_event_penalty" json:"slashing_event_penalty"`
	SecurityViolationPenalty int `yaml:"security_violation_penalty" json:"security_violation_penalty"`

	// Decay
	InactivityDecayPerWeek int `yaml:"inactivity_decay_per_week" json:"inactivity_decay_per_week"`
	DecayFloor             int `yaml:"decay_floor" json:"decay_floor"` // Minimum score from decay

	// Tier thresholds
	EliteThreshold      int     `yaml:"elite_threshold" json:"elite_threshold"`
	TrustedThreshold    int     `yaml:"trusted_threshold" json:"trusted_threshold"`
	StandardThreshold   int     `yaml:"standard_threshold" json:"standard_threshold"`
	ProbationThreshold  int     `yaml:"probation_threshold" json:"probation_threshold"`

	// Tier multipliers
	EliteMultiplier     float64 `yaml:"elite_multiplier" json:"elite_multiplier"`
	TrustedMultiplier   float64 `yaml:"trusted_multiplier" json:"trusted_multiplier"`
	StandardMultiplier  float64 `yaml:"standard_multiplier" json:"standard_multiplier"`
	ProbationMultiplier float64 `yaml:"probation_multiplier" json:"probation_multiplier"`
	RestrictedMultiplier float64 `yaml:"restricted_multiplier" json:"restricted_multiplier"`
}

// DefaultReputationConfig returns the default reputation configuration
func DefaultReputationConfig() *ReputationConfig {
	return &ReputationConfig{
		// Score range
		MinScore:     0,
		MaxScore:     1000,
		InitialScore: 500,

		// Positive actions
		JobCompletedPoints:          5,
		JobEarlyCompletionPoints:    10,
		PerfectMonthPoints:          20,
		CommunityVerificationPoints: 50,

		// Negative actions
		JobTimeoutPenalty:        -10,
		HealthFailurePenalty:     -25,
		ReplicaMismatchPenalty:   -50,
		SlashingEventPenalty:     -100,
		SecurityViolationPenalty: -200,

		// Decay
		InactivityDecayPerWeek: 1,
		DecayFloor:             100,

		// Tier thresholds
		EliteThreshold:     900,
		TrustedThreshold:   750,
		StandardThreshold:  500,
		ProbationThreshold: 250,

		// Tier multipliers
		EliteMultiplier:      1.2,
		TrustedMultiplier:    1.1,
		StandardMultiplier:   1.0,
		ProbationMultiplier:  0.9,
		RestrictedMultiplier: 0.8,
	}
}

// ProviderState represents the current state of a provider
type ProviderState struct {
	NodeID            NodeID      `json:"node_id"`
	WalletAddress     string      `json:"wallet_address"`
	StakedAmount      *big.Int    `json:"staked_amount"`
	Tier              StakingTier `json:"tier"`
	ReputationScore   int         `json:"reputation_score"`
	Uptime30Days      float64     `json:"uptime_30_days"`       // Percentage
	JobsCompleted     int         `json:"jobs_completed"`
	JobsCompleted7Days int        `json:"jobs_completed_7_days"`
	ActiveJobs        int         `json:"active_jobs"`
	TotalEarnings     *big.Int    `json:"total_earnings"`
	RegisteredAt      time.Time   `json:"registered_at"`
	LastSeenAt        time.Time   `json:"last_seen_at"`
	UnstakeInitiated  *time.Time  `json:"unstake_initiated,omitempty"`
	Region            string      `json:"region"`
	Country           string      `json:"country"`
}

// FairnessConfig defines fairness algorithm parameters
type FairnessConfig struct {
	// Weight distribution (should sum to 1.0)
	ReputationWeight float64 `yaml:"reputation_weight" json:"reputation_weight"`
	UptimeWeight     float64 `yaml:"uptime_weight" json:"uptime_weight"`
	CapacityWeight   float64 `yaml:"capacity_weight" json:"capacity_weight"`
	LatencyWeight    float64 `yaml:"latency_weight" json:"latency_weight"`
	LongevityWeight  float64 `yaml:"longevity_weight" json:"longevity_weight"`

	// Fairness boost
	MaxFairnessBoost     float64 `yaml:"max_fairness_boost" json:"max_fairness_boost"`           // Maximum boost (e.g., 0.5 = 50%)
	FairnessWindowDays   int     `yaml:"fairness_window_days" json:"fairness_window_days"`       // Look back period
	MinReputationToAccept int    `yaml:"min_reputation_to_accept" json:"min_reputation_to_accept"` // Minimum rep to accept jobs
}

// DefaultFairnessConfig returns the default fairness configuration
func DefaultFairnessConfig() *FairnessConfig {
	return &FairnessConfig{
		// Weights (sum to 1.0)
		ReputationWeight: 0.35,
		UptimeWeight:     0.25,
		CapacityWeight:   0.20,
		LatencyWeight:    0.10,
		LongevityWeight:  0.10,

		// Fairness boost
		MaxFairnessBoost:      0.5,  // 50% max boost
		FairnessWindowDays:    7,    // 7-day window
		MinReputationToAccept: 250,  // Minimum 250 rep to accept jobs
	}
}

// ProtocolFees defines protocol fee distribution
type ProtocolFees struct {
	TotalFeePercent   float64 `yaml:"total_fee_percent" json:"total_fee_percent"`     // Total protocol fee (e.g., 5%)
	BurnPercent       float64 `yaml:"burn_percent" json:"burn_percent"`               // Percentage of fee burned
	TreasuryPercent   float64 `yaml:"treasury_percent" json:"treasury_percent"`       // Percentage to treasury
}

// DefaultProtocolFees returns the default protocol fee configuration
func DefaultProtocolFees() *ProtocolFees {
	return &ProtocolFees{
		TotalFeePercent: 5.0,  // 5% total fee
		BurnPercent:     80.0, // 80% of fee burned (4% of total)
		TreasuryPercent: 20.0, // 20% of fee to treasury (1% of total)
	}
}

// UnstakingConfig defines unstaking parameters
type UnstakingConfig struct {
	CooldownDays       int `yaml:"cooldown_days" json:"cooldown_days"`               // Days to wait before unstaking
	MaxActiveJobsHours int `yaml:"max_active_jobs_hours" json:"max_active_jobs_hours"` // Max time to finish active jobs
}

// DefaultUnstakingConfig returns the default unstaking configuration
func DefaultUnstakingConfig() *UnstakingConfig {
	return &UnstakingConfig{
		CooldownDays:       7,  // 7-day cooldown
		MaxActiveJobsHours: 24, // 24 hours to finish jobs
	}
}
