package daemon

import (
	"os/exec"
	"path/filepath"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/networking"
	"github.com/moltbunker/moltbunker/internal/runtime"
)

// attachProfileStore wires the R20 per-tenant security-profile store onto the
// concrete containerd client rooted at <dataDir>/profiles. On error it logs and
// leaves persistence disabled (the documented safe-but-downgraded default)
// rather than failing daemon startup. Must be called BEFORE
// LoadExistingContainers so reattached containers recover their stored profile.
func attachProfileStore(cc *runtime.ContainerdClient, dataDir string) {
	ps, err := runtime.NewProfileStore(filepath.Join(dataDir, "profiles"))
	if err != nil {
		logging.Warn("failed to init profile store; restart may downgrade tenant profiles",
			logging.Err(err))
		return
	}
	cc.SetProfileStore(ps)
}

// buildImageScanner returns the R4 image scanner. It returns a real Trivy
// scanner (wrapped in an in-memory cache) ONLY when scanning is enabled and the
// trivy binary is on PATH; otherwise it returns a NoopScanner. This guarantees
// a deploy on a host without trivy never fails the scan gate. The result is
// never nil.
func buildImageScanner(enabled bool) runtime.ImageScanner {
	if !enabled {
		return runtime.NewNoopScanner()
	}
	if _, err := exec.LookPath("trivy"); err != nil {
		logging.Warn("image scanning enabled but trivy binary not found on PATH; using no-op scanner",
			logging.Err(err))
		return runtime.NewNoopScanner()
	}
	logging.Info("image vulnerability scanning enabled (trivy)")
	return runtime.NewCachedScanner(runtime.NewTrivyCLIScanner())
}

// buildImageCrypter returns the R5 image-at-rest crypter. It returns a real
// ocicrypt/X25519-backed crypter ONLY when image encryption is explicitly
// enabled; otherwise a NoopImageCrypter (pass-through, Enabled()=false) so the
// decrypt/encrypt hooks in CreateSecureContainer never touch unencrypted public
// images. The result is never nil. Opt-in and R11-gated: see the
// ImageEncryptionEnabled config doc.
func buildImageCrypter(enabled bool) runtime.ImageCrypter {
	if !enabled {
		return runtime.NewNoopImageCrypter()
	}
	logging.Info("image content encryption at rest enabled (ocicrypt/X25519)")
	return runtime.NewOcicryptImageCrypter()
}

// imageDecryptKey returns this node's stable X25519 private key for decrypting
// image layers at rest, or nil when image encryption is disabled or the
// provider key is unavailable. nil makes the runtime decrypt hook a no-op.
func (cm *ContainerManager) imageDecryptKey() []byte {
	if cm.imageCrypter == nil || !cm.imageCrypter.Enabled() || cm.providerKey == nil {
		return nil
	}
	return cm.providerKey.privateKey()
}

// imageEncryptRecipients returns the X25519 public keys an image should be
// encrypted to at rest. v1 uses a SELF-RECIPIENT model: each provider encrypts
// the images it pulls to its OWN key. This needs no cross-node key delivery
// because, in the current data plane, every provider (originator + replicas)
// pulls the same image from the registry independently — so each can protect its
// own at-rest copy with zero coordination. Sealing to all replicas' keys (so one
// provider can encrypt-and-distribute) is a follow-up gated on per-node X25519
// advertisement (the same capability E2E exec replica-seeding needs). Returns
// nil when encryption is disabled or the provider key is unavailable.
func (cm *ContainerManager) imageEncryptRecipients() [][]byte {
	if cm.imageCrypter == nil || !cm.imageCrypter.Enabled() || cm.providerKey == nil {
		return nil
	}
	return [][]byte{cm.providerKey.PublicKey()}
}

// security_policy.go — daemon-zone translation of per-deployment security policy
// (R3 image signature / trust, R4 scan policy, R13/R14 network/egress policy)
// from the wire types (DeployRequest / Deployment) into the runtime and
// networking primitives that the gates actually consume.
//
// GUIDING INVARIANT: every translator here is opt-out by default. A request or
// deployment that carries none of the new fields produces a zero-valued
// TrustPolicy (RequireSignature=false), a nil ImageSignature, the daemon's
// default ScanPolicy, and an allow-all NetworkPolicy — i.e. behavior identical
// to before this wiring existed.

// toImageSignature converts the wire signature spec to the runtime type.
// Returns nil when no signature was supplied (the R3 gate stays dormant).
func toImageSignature(spec *ImageSignatureSpec) *runtime.ImageSignature {
	if spec == nil {
		return nil
	}
	return &runtime.ImageSignature{
		Digest:      runtime.ImageDigest(spec.Digest),
		PublisherID: spec.PublisherID,
		Signature:   spec.Signature,
	}
}

// toTrustPolicy builds the runtime trust policy from request fields.
//
// SAFETY: RequireSignature is only honored when at least one trusted publisher
// is supplied. A bare RequireSignature=true with an empty trust list would be a
// deny-all sentinel (every image rejected), which is almost never what a caller
// intends and would break the deploy. We therefore down-grade it to "off" and
// rely on a caller that genuinely wants enforcement to also pass publishers.
// When ImageSignature is supplied without RequireSignature, verification still
// runs (the runtime gate fires on ImageSignature != nil) but unsigned images
// are not rejected.
func toTrustPolicy(requireSignature bool, trustedPublishers []string) runtime.TrustPolicy {
	require := requireSignature && len(trustedPublishers) > 0
	return runtime.TrustPolicy{
		RequireSignature:  require,
		TrustedPublishers: trustedPublishers,
	}
}

// resolveScanPolicy returns the scan policy for a deployment. Today the only
// per-deployment knob exposed on the wire is the CVE allowlist; everything else
// inherits runtime.DefaultScanPolicy() (block HIGH/CRITICAL, never RequireScan).
// DefaultScanPolicy never hard-fails a clean image and — critically — only runs
// at all when a non-nil Scanner is wired in (see ContainerManager.imageScanner,
// which falls back to a NoopScanner when trivy is absent).
func resolveScanPolicy(ignoreCVEs []string) runtime.ScanPolicy {
	p := runtime.DefaultScanPolicy()
	p.IgnoreCVEs = ignoreCVEs
	return p
}

// applyNetworkPolicy installs the per-deployment R13/R14 network/egress policy
// for a container whose IP has just been allocated. A nil spec means the caller
// requested no policy: we skip enforcement entirely so behavior is identical to
// before (allow-all). When a spec is present we hand it to the enforcer (a
// no-op recorder off Linux; a stubbed nft applier on Linux). Failures are
// logged, not fatal — a deploy must not break because policy plumbing failed.
func (cm *ContainerManager) applyNetworkPolicy(deploymentID, containerIP string, spec *NetworkPolicySpec) {
	if cm.policyEnforcer == nil {
		return
	}
	policy, present := toNetworkPolicy(spec)
	if !present {
		// No policy requested — leave the container in its default allow-all
		// state, exactly as before this wiring existed.
		return
	}
	if containerIP == "" {
		logging.Warn("skipping network policy: container IP unavailable",
			logging.ContainerID(deploymentID))
		return
	}
	if err := cm.policyEnforcer.Apply(deploymentID, containerIP, policy); err != nil {
		logging.Warn("failed to apply network policy",
			logging.ContainerID(deploymentID),
			logging.Err(err))
	}
}

// removeNetworkPolicy tears down any policy rules previously applied for a
// deployment. Safe to call even when none were installed.
func (cm *ContainerManager) removeNetworkPolicy(deploymentID string) {
	if cm.policyEnforcer == nil {
		return
	}
	if err := cm.policyEnforcer.Remove(deploymentID); err != nil {
		logging.Warn("failed to remove network policy",
			logging.ContainerID(deploymentID),
			logging.Err(err))
	}
}

// toNetworkPolicy converts the wire network-policy spec into the networking
// type. A nil spec yields DefaultNetworkPolicy() (allow-all egress, full
// lateral isolation) — the current behavior. The second return reports whether
// a caller-supplied policy was present at all, so callers can skip enforcement
// entirely when nothing was requested.
func toNetworkPolicy(spec *NetworkPolicySpec) (networking.NetworkPolicy, bool) {
	if spec == nil {
		return networking.DefaultNetworkPolicy(), false
	}
	mode := networking.EgressDefaultAllow
	if spec.EgressDeny {
		mode = networking.EgressDefaultDeny
	}
	return networking.NetworkPolicy{
		AllowedPeers: spec.AllowedPeers,
		EgressMode:   mode,
		EgressAllow:  spec.EgressAllow,
		EgressDeny:   spec.EgressBlock,
	}, true
}
