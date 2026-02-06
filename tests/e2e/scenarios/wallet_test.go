//go:build e2e

package scenarios

import (
	"path/filepath"
	"testing"

	"github.com/moltbunker/moltbunker/internal/identity"
	"github.com/moltbunker/moltbunker/tests/e2e/testutil"
)

// TestE2E_WalletCreation tests wallet creation and key management
func TestE2E_WalletCreation(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	// Create wallet manager in temp directory
	keystoreDir := h.TempSubDir("keystore")
	wm, err := identity.NewWalletManager(keystoreDir)
	assert.NoError(err, "WalletManager creation should succeed")
	assert.NotNil(wm)

	// Should have an address
	addr := wm.AddressString()
	assert.NotEmpty(addr, "Wallet address should not be empty")
	assert.Contains(addr, "0x", "Address should start with 0x")

	t.Logf("Wallet address: %s", addr)
}

// TestE2E_WalletSigning tests message signing with wallet
func TestE2E_WalletSigning(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	keystoreDir := h.TempSubDir("keystore")
	wm, err := identity.NewWalletManager(keystoreDir)
	assert.NoError(err)

	// Sign a hash (empty password for test accounts)
	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = byte(i)
	}

	signature, err := wm.SignHash(hash, "")
	assert.NoError(err, "Signing should succeed")
	assert.NotEmpty(signature, "Signature should not be empty")
	assert.Equal(65, len(signature), "Signature should be 65 bytes (r + s + v)")
}

// TestE2E_WalletImportExport tests key import/export
func TestE2E_WalletImportExport(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	keystoreDir := h.TempSubDir("keystore")
	wm, err := identity.NewWalletManager(keystoreDir)
	assert.NoError(err)

	// Export private key
	key, err := wm.ExportKey("")
	assert.NoError(err, "Export should succeed")
	assert.NotNil(key, "Key should not be nil")

	// Import into a new wallet
	keystoreDir2 := h.TempSubDir("keystore2")
	wm2, err := identity.NewWalletManager(keystoreDir2)
	assert.NoError(err)

	addr, err := wm2.ImportKey(key, "testpassword")
	assert.NoError(err, "Import should succeed")
	assert.Equal(wm.Address(), addr, "Imported address should match original")
}

// TestE2E_NodeIdentity tests node identity key management
func TestE2E_NodeIdentity(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	// Create key manager (use non-existent path so it generates new keys)
	keyPath := filepath.Join(h.TempSubDir("keys"), "node.key")
	km, err := identity.NewKeyManager(keyPath)
	assert.NoError(err, "KeyManager creation should succeed")
	assert.NotNil(km)

	// Should have a node ID
	nodeID := km.NodeID()
	assert.NotEqual([32]byte{}, nodeID, "Node ID should not be zero")

	// Should have a public key
	pubKey := km.PublicKey()
	assert.NotNil(pubKey, "Public key should not be nil")
	assert.Equal(32, len(pubKey), "Ed25519 public key should be 32 bytes")

	// Creating another key manager with same path should load same key
	km2, err := identity.NewKeyManager(keyPath)
	assert.NoError(err)
	assert.Equal(km.NodeID(), km2.NodeID(), "Same key path should produce same node ID")

	t.Logf("Node ID: %s", nodeID.String()[:16])
}

// TestE2E_CertificateGeneration tests TLS certificate generation from node identity
func TestE2E_CertificateGeneration(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	keyPath := filepath.Join(h.TempSubDir("cert-keys"), "cert-node.key")
	km, err := identity.NewKeyManager(keyPath)
	assert.NoError(err)

	// Create certificate manager
	certMgr, err := identity.NewCertificateManager(km)
	assert.NoError(err, "CertificateManager creation should succeed")
	assert.NotNil(certMgr)

	// Should have a valid TLS config
	tlsConfig := certMgr.TLSConfig()
	assert.NotNil(tlsConfig, "TLS config should not be nil")
	assert.NotEmpty(tlsConfig.Certificates, "Should have at least one certificate")

	// Should have a certificate
	cert := certMgr.Certificate()
	assert.NotNil(cert, "Certificate should not be nil")
}

// TestE2E_MultipleWallets tests managing multiple wallets
func TestE2E_MultipleWallets(t *testing.T) {
	h := testutil.NewTestHarness(t)
	assert := testutil.NewAssertions(t)

	// Create first wallet (auto-generated)
	ks1Dir := h.TempSubDir("wallet1")
	wm1, err := identity.NewWalletManager(ks1Dir)
	assert.NoError(err)

	// Save original address before creating another account
	originalAddr := wm1.Address()

	// Create second wallet with explicit password
	addr2, err := wm1.CreateAccount("password123")
	assert.NoError(err)
	assert.NotEqual(originalAddr, addr2, "New account should have different address")

	// Create wallet in different directory
	ks2Dir := h.TempSubDir("wallet2")
	wm2, err := identity.NewWalletManager(ks2Dir)
	assert.NoError(err)
	assert.NotEqual(wm1.AddressString(), wm2.AddressString(),
		"Different keystores should have different addresses")

	t.Logf("Wallet 1: %s", wm1.AddressString())
	t.Logf("Wallet 2: %s", wm2.AddressString())
}
