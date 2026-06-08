package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/containers/ocicrypt"
	"github.com/containers/ocicrypt/config"
	"github.com/containers/ocicrypt/keywrap"

	"github.com/moltbunker/moltbunker/internal/security"
)

// R5 — Image content encryption at rest (custom ocicrypt keywrapper).
//
// ocicrypt encrypts each image layer with a fresh symmetric key, then asks a
// registered keywrapper to "wrap" (seal) that key to the recipients who are
// allowed to decrypt the layer. The standard wrappers (jwe/pgp/pkcs7/pkcs11)
// assume X.509 / PGP / NIST-curve recipient keys. Moltbunker already issues
// every provider a stable X25519 keypair (used for exec-key delivery — see
// internal/security/ecies.go and internal/daemon/provider_key.go), so this file
// registers a moltbunker-native keywrapper that seals each layer key to those
// X25519 keys via the existing SealToX25519 / OpenFromX25519 ECIES primitives.
//
// The result is a standard OCI encrypted image (standard per-layer AES, standard
// encrypted media types, imgcrypt-native decrypt) whose layer keys are wrapped
// with moltbunker's own per-recipient key delivery — no external keyprovider
// binary or gRPC service, no extra NIST keypair on every node.
//
// Design boundaries (mirror image_verify.go's R3 notes):
//   - The keywrapper is pure crypto: it seals/opens an opaque optsData blob (the
//     per-layer symmetric key bundle ocicrypt hands it). It knows nothing about
//     where recipient keys come from — the daemon zone sources them from the
//     deploy request / gossiped Deployment and passes them in via the ocicrypt
//     EncryptConfig/DecryptConfig Parameters maps.
//   - Recipients are caller-supplied X25519 public keys. The keywrapper itself
//     is N-recipient capable, but R5 v1 passes a SELF-RECIPIENT set (this node's
//     own key only) — each provider encrypts the images it pulls to itself; see
//     the recipient-model note in image_encrypt.go and security_policy.go.
//   - Decryption requires THIS node's stable X25519 private key.

const (
	// X25519Scheme is the ocicrypt keywrapper scheme name registered for the
	// moltbunker X25519 wrapper.
	X25519Scheme = "moltbunker.x25519"

	// x25519AnnotationID is the OCI image annotation key under which wrapped
	// layer keys are stored. It follows the ocicrypt
	// "org.opencontainers.image.enc.keys.<scheme>" convention.
	x25519AnnotationID = "org.opencontainers.image.enc.keys.moltbunker.x25519"

	// X25519RecipientsParam is the EncryptConfig.Parameters key under which the
	// caller passes the recipient X25519 public keys (each a 32-byte slice).
	X25519RecipientsParam = "moltbunker.x25519-recipients"

	// X25519PrivKeyParam is the DecryptConfig.Parameters key under which the
	// caller passes this node's stable X25519 private key (a 32-byte slice).
	X25519PrivKeyParam = "moltbunker.x25519-privkey"
)

// wrappedLayerKey is the JSON payload stored in the per-layer annotation: one
// ECIES envelope per recipient. On decrypt, a node opens whichever envelope was
// sealed to its own X25519 key.
type wrappedLayerKey struct {
	Recipients []*security.X25519Envelope `json:"recipients"`
}

// x25519KeyWrapper implements ocicrypt's keywrap.KeyWrapper using moltbunker's
// X25519 ECIES envelopes.
type x25519KeyWrapper struct{}

// GetAnnotationID implements keywrap.KeyWrapper.
func (kw *x25519KeyWrapper) GetAnnotationID() string { return x25519AnnotationID }

// WrapKeys seals optsData (the per-layer symmetric key bundle) to every
// recipient X25519 public key found in ec.Parameters[X25519RecipientsParam].
//
// Returning (nil, nil) when no recipients are configured is intentional and
// matches the built-in jwe wrapper: it lets the scheme stay dormant (no
// annotation emitted) when moltbunker X25519 recipients aren't in use, so this
// wrapper never interferes with images encrypted by other schemes.
func (kw *x25519KeyWrapper) WrapKeys(ec *config.EncryptConfig, optsData []byte) ([]byte, error) {
	recipients := ec.Parameters[X25519RecipientsParam]
	if len(recipients) == 0 {
		return nil, nil
	}

	wrapped := wrappedLayerKey{Recipients: make([]*security.X25519Envelope, 0, len(recipients))}
	for _, pub := range recipients {
		env, err := security.SealToX25519(pub, optsData)
		if err != nil {
			return nil, fmt.Errorf("x25519 keywrap: seal layer key to recipient: %w", err)
		}
		wrapped.Recipients = append(wrapped.Recipients, env)
	}

	out, err := json.Marshal(wrapped)
	if err != nil {
		return nil, fmt.Errorf("x25519 keywrap: marshal wrapped keys: %w", err)
	}
	return out, nil
}

// UnwrapKey opens the first recipient envelope that this node's X25519 private
// key (dc.Parameters[X25519PrivKeyParam]) can decrypt, returning the original
// optsData. It is fail-closed: any failure returns an error and never returns
// partial/ciphertext bytes.
func (kw *x25519KeyWrapper) UnwrapKey(dc *config.DecryptConfig, annotation []byte) ([]byte, error) {
	privKeys := dc.Parameters[X25519PrivKeyParam]
	if len(privKeys) == 0 {
		return nil, errors.New("x25519 keywrap: no X25519 private key available for decryption")
	}

	var wrapped wrappedLayerKey
	if err := json.Unmarshal(annotation, &wrapped); err != nil {
		return nil, fmt.Errorf("x25519 keywrap: unmarshal wrapped keys: %w", err)
	}
	if len(wrapped.Recipients) == 0 {
		return nil, errors.New("x25519 keywrap: annotation carries no recipient envelopes")
	}

	for _, priv := range privKeys {
		for _, env := range wrapped.Recipients {
			if env == nil {
				continue
			}
			if plain, err := security.OpenFromX25519(priv, env); err == nil {
				return plain, nil
			}
		}
	}
	return nil, errors.New("x25519 keywrap: no private key matches any recipient envelope")
}

// NoPossibleKeys implements keywrap.KeyWrapper: decryption is impossible if no
// X25519 private key is configured.
func (kw *x25519KeyWrapper) NoPossibleKeys(dcparameters map[string][][]byte) bool {
	return len(kw.GetPrivateKeys(dcparameters)) == 0
}

// GetPrivateKeys implements keywrap.KeyWrapper.
func (kw *x25519KeyWrapper) GetPrivateKeys(dcparameters map[string][][]byte) [][]byte {
	return dcparameters[X25519PrivKeyParam]
}

// GetKeyIdsFromPacket implements keywrap.KeyWrapper. The X25519 scheme has no
// notion of key IDs.
func (kw *x25519KeyWrapper) GetKeyIdsFromPacket(_ string) ([]uint64, error) { return nil, nil }

// GetRecipients implements keywrap.KeyWrapper. Recipient X25519 keys are not
// recoverable from the wrapped packet, so a stable sentinel is returned.
func (kw *x25519KeyWrapper) GetRecipients(_ string) ([]string, error) {
	return []string{"[moltbunker.x25519]"}, nil
}

// Compile-time assertion that the wrapper satisfies the ocicrypt interface.
var _ keywrap.KeyWrapper = (*x25519KeyWrapper)(nil)

var registerX25519Once sync.Once

// RegisterX25519KeyWrapper registers the moltbunker X25519 keywrapper with
// ocicrypt's global keywrapper registry. It is idempotent and safe to call from
// every ImageCrypter constructor; the underlying registry mutation runs exactly
// once. Registration wires both the encrypt path (scheme -> wrapper) and the
// decrypt path (annotation -> scheme), so encrypted images produced by this
// node are decryptable by any node that has also registered the wrapper.
func RegisterX25519KeyWrapper() {
	registerX25519Once.Do(func() {
		ocicrypt.RegisterKeyWrapper(X25519Scheme, &x25519KeyWrapper{})
	})
}
