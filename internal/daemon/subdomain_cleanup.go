package daemon

import (
	"context"
	"strings"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/p2p"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/internal/util"
)

// StartSubdomainCleanup starts a background goroutine that periodically
// removes expired subdomain mappings from gossip state. This prevents stale
// "subdomain:<name>" entries from persisting indefinitely after the on-chain
// registration has expired.
//
// The cleanup runs every hour and checks each gossip "subdomain:*" entry
// against the on-chain BunkerRegistry to see if the name has expired.
func StartSubdomainCleanup(ctx context.Context, gossip *p2p.GossipProtocol, ps *payment.PaymentService) {
	if gossip == nil || ps == nil {
		return
	}

	util.SafeGoWithName("subdomain-expiry-cleanup", func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cleanExpiredSubdomains(ctx, gossip, ps)
			}
		}
	})
}

// cleanExpiredSubdomains iterates gossip "subdomain:*" entries and removes
// any that are expired on-chain.
func cleanExpiredSubdomains(ctx context.Context, gossip *p2p.GossipProtocol, ps *payment.PaymentService) {
	entries := gossip.GetStateByPrefix("subdomain:")
	if len(entries) == 0 {
		return
	}

	removed := 0
	for key := range entries {
		// Extract name from "subdomain:<name>"
		name := strings.TrimPrefix(key, "subdomain:")
		if name == "" {
			continue
		}

		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		expired, err := ps.IsSubdomainExpired(checkCtx, name)
		cancel()

		if err != nil {
			// Skip entries we can't check (RPC error, etc.)
			continue
		}

		if expired {
			gossip.UpdateState(key, nil)
			removed++
		}
	}

	if removed > 0 {
		logging.Info("cleaned expired subdomain gossip entries",
			"removed", removed,
			"checked", len(entries),
			logging.Component("subdomain-cleanup"))
	}
}
