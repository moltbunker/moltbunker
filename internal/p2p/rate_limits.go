package p2p

import (
	"time"

	"github.com/moltbunker/moltbunker/pkg/types"
	"golang.org/x/time/rate"
)

// TierRateLimit defines the message rate limit and burst for a staking tier.
type TierRateLimit struct {
	Rate  rate.Limit
	Burst int
}

// TierBanDuration defines the auto-ban duration for rate limit violations per tier.
// Higher-staked peers get shorter bans (they have more to lose economically).
type TierBanDuration struct {
	Duration time.Duration
}

// tierRateLimits maps staking tiers to their rate limits.
var tierRateLimits = map[string]TierRateLimit{
	"unstaked": {Rate: 10, Burst: 20},
	"unknown":  {Rate: 20, Burst: 40}, // announced but stake not yet verified
	string(types.StakingTierStarter):  {Rate: 50, Burst: 100},
	string(types.StakingTierBronze):   {Rate: 100, Burst: 200},
	string(types.StakingTierSilver):   {Rate: 200, Burst: 400},
	string(types.StakingTierGold):     {Rate: 500, Burst: 1000},
	string(types.StakingTierPlatinum): {Rate: 1000, Burst: 2000},
}

// tierBanDurations maps staking tiers to their auto-ban durations.
var tierBanDurations = map[string]time.Duration{
	"unstaked":                         1 * time.Hour,
	"unknown":                          1 * time.Hour,
	string(types.StakingTierStarter):   30 * time.Minute,
	string(types.StakingTierBronze):    30 * time.Minute,
	string(types.StakingTierSilver):    15 * time.Minute,
	string(types.StakingTierGold):      5 * time.Minute,
	string(types.StakingTierPlatinum):  5 * time.Minute,
}

// GetTierRateLimit returns the rate limit for a given tier string.
// Returns the "unstaked" limit if the tier is not recognized.
func GetTierRateLimit(tier string) TierRateLimit {
	if rl, ok := tierRateLimits[tier]; ok {
		return rl
	}
	return tierRateLimits["unstaked"]
}

// GetTierBanDuration returns the auto-ban duration for a given tier string.
func GetTierBanDuration(tier string) time.Duration {
	if d, ok := tierBanDurations[tier]; ok {
		return d
	}
	return tierBanDurations["unstaked"]
}

// NewTieredLimiter creates a rate.Limiter sized for the given tier.
func NewTieredLimiter(tier string) *rate.Limiter {
	rl := GetTierRateLimit(tier)
	return rate.NewLimiter(rl.Rate, rl.Burst)
}
