package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"golang.org/x/crypto/bcrypt"
)

// APIKey represents an API key with metadata
type APIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	KeyHash     string    `json:"key_hash"`     // bcrypt hash of the key
	KeyPrefix   string    `json:"key_prefix"`   // First 8 chars for identification
	Permissions []string  `json:"permissions"`  // e.g., ["read", "write", "admin"]
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
	Enabled     bool      `json:"enabled"`
	RateLimit   int       `json:"rate_limit,omitempty"` // Custom rate limit for this key
}

// APIKeyManager manages API keys
type APIKeyManager struct {
	keys     map[string]*APIKey // keyed by ID
	prefixes map[string]string  // prefix -> ID for quick lookup
	mu       sync.RWMutex
	filePath string
}

// NewAPIKeyManager creates a new API key manager with file storage
func NewAPIKeyManager(filePath string) (*APIKeyManager, error) {
	m := &APIKeyManager{
		keys:     make(map[string]*APIKey),
		prefixes: make(map[string]string),
		filePath: filePath,
	}

	// Load existing keys
	if err := m.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load API keys: %w", err)
	}

	return m, nil
}

// NewAPIKeyManagerInMemory creates an in-memory API key manager
func NewAPIKeyManagerInMemory() *APIKeyManager {
	m := &APIKeyManager{
		keys:     make(map[string]*APIKey),
		prefixes: make(map[string]string),
	}

	// Create a default development key
	key, plainKey, _ := m.CreateKey("default", []string{"read", "write"}, 0)
	logging.Info("created default API key (development mode)",
		"key_prefix", key.KeyPrefix,
		"plain_key", plainKey,
		logging.Component("api"))

	return m
}

// load loads keys from file
func (m *APIKeyManager) load() error {
	if m.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	var keys []*APIKey
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		m.keys[key.ID] = key
		m.prefixes[key.KeyPrefix] = key.ID
	}

	return nil
}

// save saves keys to file
func (m *APIKeyManager) save() error {
	if m.filePath == "" {
		return nil
	}

	m.mu.RLock()
	keys := make([]*APIKey, 0, len(m.keys))
	for _, key := range m.keys {
		keys = append(keys, key)
	}
	m.mu.RUnlock()

	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0600)
}

// CreateKey creates a new API key and returns it (the plain key is only returned once)
func (m *APIKeyManager) CreateKey(name string, permissions []string, expiresInDays int) (*APIKey, string, error) {
	// Generate random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate key: %w", err)
	}

	// Create the key string: mb_<random>
	plainKey := "mb_" + hex.EncodeToString(keyBytes)

	// Hash the key for storage
	hash, err := bcrypt.GenerateFromPassword([]byte(plainKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash key: %w", err)
	}

	// Generate ID
	idBytes := make([]byte, 8)
	rand.Read(idBytes)
	id := hex.EncodeToString(idBytes)

	// Create key object
	key := &APIKey{
		ID:          id,
		Name:        name,
		KeyHash:     string(hash),
		KeyPrefix:   plainKey[:11], // "mb_" + 8 chars
		Permissions: permissions,
		CreatedAt:   time.Now(),
		Enabled:     true,
	}

	if expiresInDays > 0 {
		key.ExpiresAt = time.Now().AddDate(0, 0, expiresInDays)
	}

	// Store
	m.mu.Lock()
	m.keys[id] = key
	m.prefixes[key.KeyPrefix] = id
	m.mu.Unlock()

	// Save to file
	if err := m.save(); err != nil {
		logging.Warn("failed to save API keys",
			"error", err.Error(),
			logging.Component("api"))
	}

	logging.Info("API key created",
		"key_id", id,
		"name", name,
		"prefix", key.KeyPrefix,
		logging.Component("api"))

	return key, plainKey, nil
}

// ValidateKey validates an API key and returns true if valid
func (m *APIKeyManager) ValidateKey(key string) bool {
	if len(key) < 11 {
		return false
	}

	prefix := key[:11]

	m.mu.RLock()
	id, exists := m.prefixes[prefix]
	if !exists {
		m.mu.RUnlock()
		return false
	}

	apiKey := m.keys[id]
	m.mu.RUnlock()

	if apiKey == nil || !apiKey.Enabled {
		return false
	}

	// Check expiration
	if !apiKey.ExpiresAt.IsZero() && time.Now().After(apiKey.ExpiresAt) {
		return false
	}

	// Verify hash
	if err := bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(key)); err != nil {
		return false
	}

	// Update last used time (async to not block validation)
	go func() {
		m.mu.Lock()
		if k, ok := m.keys[id]; ok {
			k.LastUsedAt = time.Now()
		}
		m.mu.Unlock()
		m.save()
	}()

	return true
}

// ValidateKeyWithPermission validates a key and checks if it has the required permission
func (m *APIKeyManager) ValidateKeyWithPermission(key, permission string) bool {
	if len(key) < 11 {
		return false
	}

	prefix := key[:11]

	m.mu.RLock()
	id, exists := m.prefixes[prefix]
	if !exists {
		m.mu.RUnlock()
		return false
	}

	apiKey := m.keys[id]
	m.mu.RUnlock()

	if apiKey == nil || !apiKey.Enabled {
		return false
	}

	// Check expiration
	if !apiKey.ExpiresAt.IsZero() && time.Now().After(apiKey.ExpiresAt) {
		return false
	}

	// Verify hash
	if err := bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(key)); err != nil {
		return false
	}

	// Check permission
	hasPermission := false
	for _, p := range apiKey.Permissions {
		if p == permission || p == "admin" {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return false
	}

	// Update last used
	go func() {
		m.mu.Lock()
		if k, ok := m.keys[id]; ok {
			k.LastUsedAt = time.Now()
		}
		m.mu.Unlock()
		m.save()
	}()

	return true
}

// GetKeyByID retrieves a key by ID (without the actual key value)
func (m *APIKeyManager) GetKeyByID(id string) (*APIKey, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key, exists := m.keys[id]
	return key, exists
}

// ListKeys returns all keys (without actual key values)
func (m *APIKeyManager) ListKeys() []*APIKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]*APIKey, 0, len(m.keys))
	for _, key := range m.keys {
		keys = append(keys, key)
	}
	return keys
}

// RevokeKey revokes (disables) a key
func (m *APIKeyManager) RevokeKey(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key, exists := m.keys[id]
	if !exists {
		return fmt.Errorf("key not found: %s", id)
	}

	key.Enabled = false

	if err := m.save(); err != nil {
		return err
	}

	logging.Info("API key revoked",
		"key_id", id,
		"name", key.Name,
		logging.Component("api"))

	return nil
}

// DeleteKey permanently deletes a key
func (m *APIKeyManager) DeleteKey(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key, exists := m.keys[id]
	if !exists {
		return fmt.Errorf("key not found: %s", id)
	}

	delete(m.prefixes, key.KeyPrefix)
	delete(m.keys, id)

	if err := m.save(); err != nil {
		return err
	}

	logging.Info("API key deleted",
		"key_id", id,
		"name", key.Name,
		logging.Component("api"))

	return nil
}

// UpdateKeyPermissions updates permissions for a key
func (m *APIKeyManager) UpdateKeyPermissions(id string, permissions []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key, exists := m.keys[id]
	if !exists {
		return fmt.Errorf("key not found: %s", id)
	}

	key.Permissions = permissions

	return m.save()
}

// UpdateKey updates mutable fields of an API key (name, rate limit, enabled).
// Nil pointer fields are left unchanged.
func (m *APIKeyManager) UpdateKey(id string, name *string, rateLimit *int, enabled *bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key, exists := m.keys[id]
	if !exists {
		return fmt.Errorf("key not found: %s", id)
	}

	if name != nil {
		key.Name = *name
	}
	if rateLimit != nil {
		key.RateLimit = *rateLimit
	}
	if enabled != nil {
		key.Enabled = *enabled
	}

	if err := m.save(); err != nil {
		return err
	}

	logging.Info("API key updated",
		"key_id", id,
		"name", key.Name,
		logging.Component("api"))

	return nil
}

// CountByStatus returns (total, active, revoked) key counts
func (m *APIKeyManager) CountByStatus() (total, active, revoked int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total = len(m.keys)
	for _, key := range m.keys {
		if key.Enabled {
			active++
		} else {
			revoked++
		}
	}
	return
}

// GetKeyFromRequest extracts and validates API key from request header
func (m *APIKeyManager) GetKeyFromRequest(apiKey string) (*APIKey, bool) {
	if len(apiKey) < 11 {
		return nil, false
	}

	prefix := apiKey[:11]

	m.mu.RLock()
	id, exists := m.prefixes[prefix]
	if !exists {
		m.mu.RUnlock()
		return nil, false
	}

	key := m.keys[id]
	m.mu.RUnlock()

	if key == nil || !key.Enabled {
		return nil, false
	}

	// Verify hash
	if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(apiKey)); err != nil {
		return nil, false
	}

	return key, true
}

// HashAPIKey creates a hash of an API key for comparison
// This is useful for quick comparisons without bcrypt
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// SecureCompare performs constant-time comparison
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
