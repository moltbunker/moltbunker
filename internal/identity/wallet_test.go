package identity

import (
	"crypto/ecdsa"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

func TestLoadWalletManager_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	wm, err := LoadWalletManager(dir)
	if err != nil {
		t.Fatalf("LoadWalletManager on empty dir: %v", err)
	}
	if wm != nil {
		t.Fatal("expected nil WalletManager for empty keystore dir")
	}
}

func TestLoadWalletManager_NonExistentDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nonexistent", "keystore")

	wm, err := LoadWalletManager(dir)
	if err != nil {
		t.Fatalf("LoadWalletManager on non-existent dir: %v", err)
	}
	if wm != nil {
		t.Fatal("expected nil WalletManager for non-existent dir")
	}

	// Directory should have been created
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected directory to be created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected a directory")
	}
}

func TestCreateWalletManager(t *testing.T) {
	dir := t.TempDir()
	password := "test-password-123"

	wm, err := CreateWalletManager(dir, password)
	if err != nil {
		t.Fatalf("CreateWalletManager: %v", err)
	}
	if wm == nil {
		t.Fatal("expected non-nil WalletManager")
	}
	if !wm.IsLoaded() {
		t.Error("expected IsLoaded() to be true")
	}
	if wm.Address().Hex() == "0x0000000000000000000000000000000000000000" {
		t.Error("expected non-zero address")
	}
	if wm.KeystoreDir() != dir {
		t.Errorf("expected KeystoreDir=%s, got %s", dir, wm.KeystoreDir())
	}
}

func TestCreateWalletManager_AlreadyExists(t *testing.T) {
	dir := t.TempDir()

	// Create first wallet
	_, err := CreateWalletManager(dir, "password1234")
	if err != nil {
		t.Fatalf("first CreateWalletManager: %v", err)
	}

	// Second create should fail
	_, err = CreateWalletManager(dir, "password5678")
	if err == nil {
		t.Fatal("expected error when wallet already exists")
	}
}

func TestCreateWalletManager_ThenLoad(t *testing.T) {
	dir := t.TempDir()
	password := "test-password-123"

	// Create
	created, err := CreateWalletManager(dir, password)
	if err != nil {
		t.Fatalf("CreateWalletManager: %v", err)
	}

	// Load
	loaded, err := LoadWalletManager(dir)
	if err != nil {
		t.Fatalf("LoadWalletManager: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil WalletManager after load")
	}
	if !loaded.IsLoaded() {
		t.Error("expected IsLoaded() to be true")
	}

	// Addresses should match
	if created.Address() != loaded.Address() {
		t.Errorf("address mismatch: created=%s loaded=%s", created.Address().Hex(), loaded.Address().Hex())
	}
}

func TestCreateWalletManager_UnlockWithPassword(t *testing.T) {
	dir := t.TempDir()
	password := "test-password-123"

	wm, err := CreateWalletManager(dir, password)
	if err != nil {
		t.Fatalf("CreateWalletManager: %v", err)
	}

	// Unlock with correct password
	privKey, err := wm.PrivateKey(password)
	if err != nil {
		t.Fatalf("PrivateKey with correct password: %v", err)
	}
	if privKey == nil {
		t.Fatal("expected non-nil private key")
	}

	// Verify the key matches the address
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	if addr != wm.Address() {
		t.Errorf("derived address %s doesn't match wallet address %s", addr.Hex(), wm.Address().Hex())
	}
}

func TestCreateWalletManager_WrongPassword(t *testing.T) {
	dir := t.TempDir()

	wm, err := CreateWalletManager(dir, "correct-password")
	if err != nil {
		t.Fatalf("CreateWalletManager: %v", err)
	}

	// Try wrong password
	_, err = wm.PrivateKey("wrong-password")
	if err == nil {
		t.Fatal("expected error with wrong password")
	}
}

func TestImportWalletManager(t *testing.T) {
	dir := t.TempDir()
	password := "import-password-123"

	// Generate a test private key
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	expectedAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	privKeyHex := crypto.FromECDSA(privKey)

	// Import
	wm, err := ImportWalletManager(dir, toHex(privKeyHex), password)
	if err != nil {
		t.Fatalf("ImportWalletManager: %v", err)
	}
	if wm == nil {
		t.Fatal("expected non-nil WalletManager")
	}
	if !wm.IsLoaded() {
		t.Error("expected IsLoaded() to be true")
	}
	if wm.Address() != expectedAddr {
		t.Errorf("address mismatch: expected=%s got=%s", expectedAddr.Hex(), wm.Address().Hex())
	}
}

func TestImportWalletManager_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	password := "roundtrip-password"

	// Generate a key
	origKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	origKeyHex := toHex(crypto.FromECDSA(origKey))

	// Import
	wm, err := ImportWalletManager(dir, origKeyHex, password)
	if err != nil {
		t.Fatalf("ImportWalletManager: %v", err)
	}

	// Export and compare
	exported, err := wm.ExportKey(password)
	if err != nil {
		t.Fatalf("ExportKey: %v", err)
	}

	if !origKey.Equal(exported) {
		t.Error("exported key doesn't match original")
	}
}

func TestImportWalletManager_AlreadyExists(t *testing.T) {
	dir := t.TempDir()

	// Create a wallet first
	_, err := CreateWalletManager(dir, "password1234")
	if err != nil {
		t.Fatalf("CreateWalletManager: %v", err)
	}

	// Import should fail
	privKey, _ := crypto.GenerateKey()
	_, err = ImportWalletManager(dir, toHex(crypto.FromECDSA(privKey)), "password5678")
	if err == nil {
		t.Fatal("expected error when wallet already exists")
	}
}

func TestImportWalletManager_InvalidHex(t *testing.T) {
	dir := t.TempDir()

	_, err := ImportWalletManager(dir, "not-valid-hex", "password1234")
	if err == nil {
		t.Fatal("expected error with invalid hex key")
	}
}

func TestIsLoaded_NilReceiver(t *testing.T) {
	var wm *WalletManager
	if wm.IsLoaded() {
		t.Error("expected IsLoaded() to be false for nil WalletManager")
	}
}

func TestIsLoaded_UnloadedWallet(t *testing.T) {
	// A WalletManager with loaded=false
	wm := &WalletManager{}
	if wm.IsLoaded() {
		t.Error("expected IsLoaded() to be false for unloaded wallet")
	}
}

func TestCreateWalletManager_KeystoreDirPermissions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "secure-keystore")

	_, err := CreateWalletManager(dir, "password1234")
	if err != nil {
		t.Fatalf("CreateWalletManager: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat keystore dir: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("expected keystore dir permissions 0700, got %04o", perm)
	}
}

func TestLoadWalletManager_ThenUnlockCaches(t *testing.T) {
	dir := t.TempDir()
	password := "cache-test-password"

	// Create wallet
	_, err := CreateWalletManager(dir, password)
	if err != nil {
		t.Fatalf("CreateWalletManager: %v", err)
	}

	// Load wallet fresh
	wm, err := LoadWalletManager(dir)
	if err != nil {
		t.Fatalf("LoadWalletManager: %v", err)
	}

	// First unlock with correct password
	key1, err := wm.PrivateKey(password)
	if err != nil {
		t.Fatalf("first PrivateKey call: %v", err)
	}

	// Second call should return cached key (even with empty password since it's cached)
	key2, err := wm.PrivateKey("")
	if err != nil {
		t.Fatalf("second PrivateKey call (cached): %v", err)
	}

	if !key1.Equal(key2) {
		t.Error("cached key doesn't match first key")
	}
}

// toHex converts bytes to a hex string without 0x prefix.
func toHex(b []byte) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, len(b)*2)
	for i, v := range b {
		result[i*2] = hexChars[v>>4]
		result[i*2+1] = hexChars[v&0x0f]
	}
	return string(result)
}

// ecdsaKeysEqual compares two ECDSA private keys.
func ecdsaKeysEqual(a, b *ecdsa.PrivateKey) bool {
	return a.D.Cmp(b.D) == 0 &&
		a.PublicKey.X.Cmp(b.PublicKey.X) == 0 &&
		a.PublicKey.Y.Cmp(b.PublicKey.Y) == 0
}
