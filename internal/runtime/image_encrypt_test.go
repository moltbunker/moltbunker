package runtime

import (
	"bytes"
	"io"
	"testing"

	"github.com/containers/ocicrypt"
	"github.com/containers/ocicrypt/config"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// TestEncryptLayer_EndToEndViaRegisteredWrapper proves that ocicrypt's real
// layer encryption path invokes the registered moltbunker X25519 keywrapper:
// it encrypts a layer to three recipients, then decrypts it with each
// recipient's private key, recovering the exact plaintext. This exercises the
// full EncryptLayer -> finalizer (WrapKeys) -> DecryptLayer (UnwrapKey) cycle
// without needing a containerd content store.
func TestEncryptLayer_EndToEndViaRegisteredWrapper(t *testing.T) {
	RegisterX25519KeyWrapper()

	pub1, priv1 := genX25519(t)
	pub2, priv2 := genX25519(t)
	pub3, priv3 := genX25519(t)

	plaintext := bytes.Repeat([]byte("moltbunker-proprietary-image-layer\n"), 64)
	plainDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageLayer,
		Digest:    digest.FromBytes(plaintext),
		Size:      int64(len(plaintext)),
	}

	ec := &config.EncryptConfig{
		Parameters: map[string][][]byte{
			X25519RecipientsParam: {pub1, pub2, pub3},
		},
	}

	encReader, finalizer, err := ocicrypt.EncryptLayer(ec, bytes.NewReader(plaintext), plainDesc)
	if err != nil {
		t.Fatalf("EncryptLayer: %v", err)
	}
	encBytes, err := io.ReadAll(encReader)
	if err != nil {
		t.Fatalf("read encrypted layer: %v", err)
	}
	if bytes.Equal(encBytes, plaintext) {
		t.Fatal("encrypted layer bytes equal plaintext — not encrypted")
	}

	annotations, err := finalizer()
	if err != nil {
		t.Fatalf("finalizer: %v", err)
	}
	// The moltbunker X25519 annotation must be present — proof ocicrypt called
	// our WrapKeys.
	if _, ok := annotations[x25519AnnotationID]; !ok {
		t.Fatalf("encrypted layer missing %q annotation; got keys %v", x25519AnnotationID, keysOf(annotations))
	}

	encDesc := plainDesc
	encDesc.Digest = digest.FromBytes(encBytes)
	encDesc.Size = int64(len(encBytes))
	encDesc.Annotations = annotations

	for i, priv := range [][]byte{priv1, priv2, priv3} {
		dc := &config.DecryptConfig{
			Parameters: map[string][][]byte{X25519PrivKeyParam: {priv}},
		}
		decReader, _, err := ocicrypt.DecryptLayer(dc, bytes.NewReader(encBytes), encDesc, false)
		if err != nil {
			t.Fatalf("DecryptLayer replica %d: %v", i, err)
		}
		decBytes, err := io.ReadAll(decReader)
		if err != nil {
			t.Fatalf("read decrypted layer replica %d: %v", i, err)
		}
		if !bytes.Equal(decBytes, plaintext) {
			t.Fatalf("replica %d: decrypted layer != original plaintext", i)
		}
	}

	// A non-recipient must not be able to decrypt.
	_, outsiderPriv := genX25519(t)
	dc := &config.DecryptConfig{
		Parameters: map[string][][]byte{X25519PrivKeyParam: {outsiderPriv}},
	}
	if _, _, err := ocicrypt.DecryptLayer(dc, bytes.NewReader(encBytes), encDesc, false); err == nil {
		t.Fatal("non-recipient was able to decrypt the layer")
	}
}

func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
