package runtime

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/containers/ocicrypt"
	"github.com/containers/ocicrypt/config"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/moltbunker/moltbunker/internal/security"
)

// genX25519 returns a fresh X25519 keypair for tests.
func genX25519(t *testing.T) (pub, priv []byte) {
	t.Helper()
	pub, priv, err := security.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateX25519KeyPair: %v", err)
	}
	return pub, priv
}

func TestX25519KeyWrapper_Registration(t *testing.T) {
	RegisterX25519KeyWrapper()
	if kw := ocicrypt.GetKeyWrapper(X25519Scheme); kw == nil {
		t.Fatalf("keywrapper not registered under scheme %q", X25519Scheme)
	}
	// The annotation->scheme reverse mapping must also be wired so the decrypt
	// path can find the wrapper by annotation.
	kw := &x25519KeyWrapper{}
	if got := kw.GetAnnotationID(); got != x25519AnnotationID {
		t.Fatalf("GetAnnotationID = %q, want %q", got, x25519AnnotationID)
	}
}

func TestX25519KeyWrapper_RoundTripMultiRecipient(t *testing.T) {
	kw := &x25519KeyWrapper{}

	// Three replicas, mirroring a real 3-region deployment.
	pub1, priv1 := genX25519(t)
	pub2, priv2 := genX25519(t)
	pub3, priv3 := genX25519(t)

	optsData := []byte("this-is-the-per-layer-symmetric-key-bundle")

	ec := &config.EncryptConfig{
		Parameters: map[string][][]byte{
			X25519RecipientsParam: {pub1, pub2, pub3},
		},
	}
	wrapped, err := kw.WrapKeys(ec, optsData)
	if err != nil {
		t.Fatalf("WrapKeys: %v", err)
	}
	if len(wrapped) == 0 {
		t.Fatal("WrapKeys returned empty annotation")
	}

	// Every replica must be able to recover the original optsData with its own
	// private key.
	for i, priv := range [][]byte{priv1, priv2, priv3} {
		dc := &config.DecryptConfig{
			Parameters: map[string][][]byte{X25519PrivKeyParam: {priv}},
		}
		got, err := kw.UnwrapKey(dc, wrapped)
		if err != nil {
			t.Fatalf("UnwrapKey replica %d: %v", i, err)
		}
		if !bytes.Equal(got, optsData) {
			t.Fatalf("UnwrapKey replica %d = %q, want %q", i, got, optsData)
		}
	}
}

func TestX25519KeyWrapper_NonRecipientCannotUnwrap(t *testing.T) {
	kw := &x25519KeyWrapper{}
	pub1, _ := genX25519(t)
	_, outsiderPriv := genX25519(t) // not a recipient

	ec := &config.EncryptConfig{
		Parameters: map[string][][]byte{X25519RecipientsParam: {pub1}},
	}
	wrapped, err := kw.WrapKeys(ec, []byte("secret-layer-key"))
	if err != nil {
		t.Fatalf("WrapKeys: %v", err)
	}

	dc := &config.DecryptConfig{
		Parameters: map[string][][]byte{X25519PrivKeyParam: {outsiderPriv}},
	}
	if _, err := kw.UnwrapKey(dc, wrapped); err == nil {
		t.Fatal("UnwrapKey with a non-recipient key must fail, got nil error")
	}
}

func TestX25519KeyWrapper_NoRecipientsIsDormant(t *testing.T) {
	kw := &x25519KeyWrapper{}
	// No recipients configured: the wrapper must stay dormant (nil, nil) so it
	// never emits an annotation for images it isn't protecting.
	ec := &config.EncryptConfig{Parameters: map[string][][]byte{}}
	out, err := kw.WrapKeys(ec, []byte("ignored"))
	if err != nil {
		t.Fatalf("WrapKeys with no recipients: unexpected error %v", err)
	}
	if out != nil {
		t.Fatalf("WrapKeys with no recipients = %v, want nil", out)
	}
}

func TestX25519KeyWrapper_NoPrivateKeyFailsClosed(t *testing.T) {
	kw := &x25519KeyWrapper{}
	pub1, _ := genX25519(t)
	ec := &config.EncryptConfig{
		Parameters: map[string][][]byte{X25519RecipientsParam: {pub1}},
	}
	wrapped, err := kw.WrapKeys(ec, []byte("secret"))
	if err != nil {
		t.Fatalf("WrapKeys: %v", err)
	}

	dc := &config.DecryptConfig{Parameters: map[string][][]byte{}}
	if _, err := kw.UnwrapKey(dc, wrapped); err == nil {
		t.Fatal("UnwrapKey with no private key must fail closed, got nil error")
	}
	if !kw.NoPossibleKeys(dc.Parameters) {
		t.Fatal("NoPossibleKeys must be true when no X25519 private key is present")
	}
}

func TestX25519KeyWrapper_TamperedAnnotationFails(t *testing.T) {
	kw := &x25519KeyWrapper{}
	pub1, priv1 := genX25519(t)
	ec := &config.EncryptConfig{
		Parameters: map[string][][]byte{X25519RecipientsParam: {pub1}},
	}
	wrapped, err := kw.WrapKeys(ec, []byte("secret-layer-key"))
	if err != nil {
		t.Fatalf("WrapKeys: %v", err)
	}

	// Flip a byte inside the ciphertext of the (only) recipient envelope.
	var wk wrappedLayerKey
	if err := json.Unmarshal(wrapped, &wk); err != nil {
		t.Fatalf("unmarshal wrapped: %v", err)
	}
	if len(wk.Recipients) == 0 || len(wk.Recipients[0].Ciphertext) == 0 {
		t.Fatal("wrapped payload missing recipient ciphertext")
	}
	wk.Recipients[0].Ciphertext[len(wk.Recipients[0].Ciphertext)-1] ^= 0xFF
	tampered, err := json.Marshal(wk)
	if err != nil {
		t.Fatalf("marshal tampered: %v", err)
	}

	dc := &config.DecryptConfig{
		Parameters: map[string][][]byte{X25519PrivKeyParam: {priv1}},
	}
	if _, err := kw.UnwrapKey(dc, tampered); err == nil {
		t.Fatal("UnwrapKey of a tampered annotation must fail (AEAD), got nil error")
	}
}

func TestNoopImageCrypter_PassThrough(t *testing.T) {
	c := NewNoopImageCrypter()
	if c.Enabled() {
		t.Fatal("NoopImageCrypter.Enabled() must be false")
	}
	// Encrypt is a programming error on the noop crypter.
	if _, err := c.Encrypt(nil, nil, ocispec.Descriptor{}, nil); err == nil {
		t.Fatal("NoopImageCrypter.Encrypt must return an error")
	}
	// Decrypt passes the descriptor through unchanged.
	in := ocispec.Descriptor{MediaType: "application/vnd.oci.image.manifest.v1+json"}
	out, err := c.Decrypt(nil, nil, in, nil)
	if err != nil {
		t.Fatalf("NoopImageCrypter.Decrypt: %v", err)
	}
	if out.MediaType != in.MediaType {
		t.Fatalf("NoopImageCrypter.Decrypt changed descriptor: got %q", out.MediaType)
	}
}
