package tunnel

// Reverse tunnel control message types.
// Uses 0x10+ range to avoid collision with forward tunnel messages (0x01-0x05).
const (
	MsgTunnelRegister   byte = 0x10 // Provider → Ingress: register tunnel
	MsgTunnelRegistered byte = 0x11 // Ingress → Provider: registration confirmed
	MsgTunnelPing       byte = 0x12 // Ingress → Provider: liveness check
	MsgTunnelPong       byte = 0x13 // Provider → Ingress: alive
	MsgTunnelDeregister byte = 0x14 // Provider → Ingress: graceful disconnect
)

// TunnelRegisterRequest is sent by the provider to the ingress to register a reverse tunnel.
type TunnelRegisterRequest struct {
	NodeID        string       `json:"node_id"`
	Subdomain     string       `json:"subdomain,omitempty"`    // empty = auto-assign
	DeploymentID  string       `json:"deployment_id"`
	ContainerPort int          `json:"container_port"`
	Nonce         string       `json:"nonce"`                  // 32-byte hex
	Timestamp     int64        `json:"timestamp"`              // unix seconds
	TLSBinding    string       `json:"tls_binding"`            // hex(tls.ConnectionState().TLSUnique)
	WalletProof   *WalletProof `json:"wallet_proof,omitempty"` // for staked tiers
	ReconnToken   string       `json:"reconn_token,omitempty"` // reconnection
}

// TunnelRegisterResponse is sent by the ingress to confirm registration.
type TunnelRegisterResponse struct {
	Subdomain   string        `json:"subdomain"`
	ReconnToken string        `json:"reconn_token"`
	Limits      *TunnelLimits `json:"limits"`
	FullDomain  string        `json:"full_domain"` // e.g., "abc123.moltbunker.dev"
}

// TunnelLimits describes the rate limits for a reverse tunnel session.
type TunnelLimits struct {
	MaxRPS       int   `json:"max_rps"`
	MaxBandwidth int64 `json:"max_bandwidth_bps"`
	MaxStreams   int   `json:"max_streams"`
}

// WalletProof proves wallet ownership for staked tier access.
type WalletProof struct {
	Address   string `json:"address"`
	Signature string `json:"signature"` // EIP-191 sig over nodeID+nonce+timestamp
	Message   string `json:"message"`
}

// TunnelPing is sent by ingress to check liveness.
type TunnelPing struct {
	Challenge [16]byte `json:"challenge"` // random nonce
}

// TunnelPong is the provider's response to a ping.
type TunnelPong struct {
	Challenge [16]byte `json:"challenge"` // echo back
}

// Default limits by tier.
var (
	FreeTierLimits = &TunnelLimits{
		MaxRPS:       10,
		MaxBandwidth: 1 * 1024 * 1024, // 1 MB/s
		MaxStreams:   100,
	}
	StarterTierLimits = &TunnelLimits{
		MaxRPS:       50,
		MaxBandwidth: 5 * 1024 * 1024, // 5 MB/s
		MaxStreams:   500,
	}
	BronzeTierLimits = &TunnelLimits{
		MaxRPS:       200,
		MaxBandwidth: 10 * 1024 * 1024, // 10 MB/s
		MaxStreams:   1000,
	}
	SilverPlusTierLimits = &TunnelLimits{
		MaxRPS:       2000,
		MaxBandwidth: 50 * 1024 * 1024, // 50 MB/s
		MaxStreams:   10000,
	}
)

// MaxSubdomainsForTier returns the subdomain cap for a staking tier.
func MaxSubdomainsForTier(tier string) int {
	switch tier {
	case "starter":
		return 10
	case "bronze":
		return 25
	case "silver", "gold", "platinum":
		return 100
	default: // free
		return 3
	}
}

// LimitsForTier returns the tunnel limits for a staking tier.
func LimitsForTier(tier string) *TunnelLimits {
	switch tier {
	case "starter":
		return StarterTierLimits
	case "bronze":
		return BronzeTierLimits
	case "silver", "gold", "platinum":
		return SilverPlusTierLimits
	default:
		return FreeTierLimits
	}
}

// ReconnGraceForTier returns the reconnection grace period for a tier.
func ReconnGraceForTier(tier string) int {
	switch tier {
	case "starter":
		return 120 // 2 min
	case "bronze":
		return 300 // 5 min
	case "silver", "gold", "platinum":
		return 1800 // 30 min
	default:
		return 60 // 1 min
	}
}
