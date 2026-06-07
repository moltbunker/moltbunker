package state

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/moltbunker/moltbunker/internal/security"
)

// BboltStore implements StateStore using bbolt (embedded B+ tree KV database).
// Each category of state (deployments, bans, peers, etc.) is stored in its own bucket.
// All operations are ACID — partial writes on crash are impossible.
//
// R8 — encryption at rest: when encKey is non-nil, every value is stored as
// encMagic || AES-256-GCM(value) and transparently decrypted on read. A nil
// encKey disables encryption (plaintext), preserving the legacy behavior.
//
// Threat model: see internal/state/statekey.go. The on-disk key mitigates
// stolen-disk / leaked-backup / casual filesystem access, not a live host-root
// attacker who can also read state.key.
type BboltStore struct {
	db     *bolt.DB
	encKey []byte // nil => encryption disabled (plaintext)
}

// encMagic is a fixed, non-JSON byte prefix marking an encrypted value on disk.
// Plaintext state values are JSON or short ASCII (schema version, timestamps),
// none of which begin with these bytes, so the prefix unambiguously
// distinguishes encrypted blobs from legacy plaintext during lazy migration.
var encMagic = []byte{0x4D, 0x42, 0x45, 0x4E, 0x43, 0x31, 0x00} // "MBENC1\x00"

// allBuckets is the list of buckets created on database open.
var allBuckets = []string{
	BucketMeta,
	BucketDeployments,
	BucketBans,
	BucketPeers,
	BucketCertPins,
	BucketAPIKeys,
	// P0 service buckets (schema v2)
	BucketStorageBuckets,
	BucketStorageObjects,
	BucketProxySessions,
	BucketCrawlJobs,
	BucketAgentState,
}

// NewBboltStore opens or creates a bbolt database at the given path.
// Parent directories are created if they don't exist. If encKey is non-nil it
// must be a 32-byte AES-256 key; values are then encrypted at rest. Pass nil to
// store values as plaintext (legacy behavior).
func NewBboltStore(path string, encKey []byte) (*BboltStore, error) {
	if encKey != nil && len(encKey) != 32 {
		return nil, fmt.Errorf("state encryption key must be 32 bytes, got %d", len(encKey))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create state db directory: %w", err)
	}

	db, err := bolt.Open(path, 0600, &bolt.Options{
		Timeout:      1 * time.Second,
		FreelistType: bolt.FreelistMapType,
	})
	if err != nil {
		return nil, fmt.Errorf("open state db: %w", err)
	}

	// Create all buckets in a single transaction
	err = db.Update(func(tx *bolt.Tx) error {
		for _, name := range allBuckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize buckets: %w", err)
	}

	return &BboltStore{db: db, encKey: encKey}, nil
}

func (s *BboltStore) Close() error {
	return s.db.Close()
}

// --- encryption helpers ---

// encode returns the bytes to store on disk for a value. With encryption enabled
// it produces encMagic || AES-256-GCM(data); otherwise data is returned as-is.
func (s *BboltStore) encode(data []byte) ([]byte, error) {
	if s.encKey == nil {
		return data, nil
	}
	ct, err := security.EncryptAES256GCM(s.encKey, data)
	if err != nil {
		return nil, fmt.Errorf("encrypt state value: %w", err)
	}
	out := make([]byte, 0, len(encMagic)+len(ct))
	out = append(out, encMagic...)
	out = append(out, ct...)
	return out, nil
}

// decode reverses encode for a value read from disk. A value carrying encMagic
// is decrypted (and a decrypt failure is a hard error — ciphertext is never
// returned). A value without the magic prefix is returned verbatim: this is the
// back-compat / lazy-migration path (legacy plaintext, or values written while
// encryption was disabled). With encryption disabled, values are returned as-is.
func (s *BboltStore) decode(stored []byte) ([]byte, error) {
	if stored == nil {
		return nil, nil
	}
	if s.encKey == nil {
		return stored, nil
	}
	if !bytes.HasPrefix(stored, encMagic) {
		return stored, nil
	}
	pt, err := security.DecryptAES256GCM(s.encKey, stored[len(encMagic):])
	if err != nil {
		return nil, fmt.Errorf("decrypt state value: %w", err)
	}
	return pt, nil
}

// --- internal helpers ---

func (s *BboltStore) put(bucket, key string, data []byte) error {
	stored, err := s.encode(data)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.Put([]byte(key), stored)
	})
}

func (s *BboltStore) get(bucket, key string) ([]byte, error) {
	var stored []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		v := b.Get([]byte(key))
		if v != nil {
			stored = make([]byte, len(v))
			copy(stored, v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.decode(stored)
}

func (s *BboltStore) del(bucket, key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.Delete([]byte(key))
	})
}

func (s *BboltStore) list(bucket string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		return b.ForEach(func(k, v []byte) error {
			cp := make([]byte, len(v))
			copy(cp, v)
			plain, derr := s.decode(cp)
			if derr != nil {
				return fmt.Errorf("decode value for key %q: %w", string(k), derr)
			}
			result[string(k)] = plain
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, err
}

// --- Deployments ---

func (s *BboltStore) PutDeployment(_ context.Context, id string, data []byte) error {
	return s.put(BucketDeployments, id, data)
}

func (s *BboltStore) GetDeployment(_ context.Context, id string) ([]byte, error) {
	return s.get(BucketDeployments, id)
}

func (s *BboltStore) DeleteDeployment(_ context.Context, id string) error {
	return s.del(BucketDeployments, id)
}

func (s *BboltStore) ListDeployments(_ context.Context) (map[string][]byte, error) {
	return s.list(BucketDeployments)
}

// --- Bans ---

func (s *BboltStore) PutBan(_ context.Context, peerID string, data []byte) error {
	return s.put(BucketBans, peerID, data)
}

func (s *BboltStore) DeleteBan(_ context.Context, peerID string) error {
	return s.del(BucketBans, peerID)
}

func (s *BboltStore) ListBans(_ context.Context) (map[string][]byte, error) {
	return s.list(BucketBans)
}

// --- Peers ---

func (s *BboltStore) PutPeer(_ context.Context, peerID string, data []byte) error {
	return s.put(BucketPeers, peerID, data)
}

func (s *BboltStore) DeletePeer(_ context.Context, peerID string) error {
	return s.del(BucketPeers, peerID)
}

func (s *BboltStore) ListPeers(_ context.Context) (map[string][]byte, error) {
	return s.list(BucketPeers)
}

// --- Certificate Pins ---

func (s *BboltStore) PutCertPin(_ context.Context, nodeID string, hash []byte) error {
	return s.put(BucketCertPins, nodeID, hash)
}

func (s *BboltStore) DeleteCertPin(_ context.Context, nodeID string) error {
	return s.del(BucketCertPins, nodeID)
}

func (s *BboltStore) ListCertPins(_ context.Context) (map[string][]byte, error) {
	return s.list(BucketCertPins)
}

// --- API Keys ---

func (s *BboltStore) PutAPIKey(_ context.Context, id string, data []byte) error {
	return s.put(BucketAPIKeys, id, data)
}

func (s *BboltStore) DeleteAPIKey(_ context.Context, id string) error {
	return s.del(BucketAPIKeys, id)
}

func (s *BboltStore) ListAPIKeys(_ context.Context) (map[string][]byte, error) {
	return s.list(BucketAPIKeys)
}

// --- Storage Buckets ---

func (s *BboltStore) PutStorageBucket(_ context.Context, name string, data []byte) error {
	return s.put(BucketStorageBuckets, name, data)
}

func (s *BboltStore) GetStorageBucket(_ context.Context, name string) ([]byte, error) {
	return s.get(BucketStorageBuckets, name)
}

func (s *BboltStore) DeleteStorageBucket(_ context.Context, name string) error {
	return s.del(BucketStorageBuckets, name)
}

func (s *BboltStore) ListStorageBuckets(_ context.Context) (map[string][]byte, error) {
	return s.list(BucketStorageBuckets)
}

// --- Storage Objects ---

func (s *BboltStore) PutStorageObject(_ context.Context, key string, data []byte) error {
	return s.put(BucketStorageObjects, key, data)
}

func (s *BboltStore) GetStorageObject(_ context.Context, key string) ([]byte, error) {
	return s.get(BucketStorageObjects, key)
}

func (s *BboltStore) DeleteStorageObject(_ context.Context, key string) error {
	return s.del(BucketStorageObjects, key)
}

func (s *BboltStore) ListStorageObjects(_ context.Context) (map[string][]byte, error) {
	return s.list(BucketStorageObjects)
}

// --- Proxy Sessions ---

func (s *BboltStore) PutProxySession(_ context.Context, id string, data []byte) error {
	return s.put(BucketProxySessions, id, data)
}

func (s *BboltStore) GetProxySession(_ context.Context, id string) ([]byte, error) {
	return s.get(BucketProxySessions, id)
}

func (s *BboltStore) DeleteProxySession(_ context.Context, id string) error {
	return s.del(BucketProxySessions, id)
}

func (s *BboltStore) ListProxySessions(_ context.Context) (map[string][]byte, error) {
	return s.list(BucketProxySessions)
}

// --- Crawl Jobs ---

func (s *BboltStore) PutCrawlJob(_ context.Context, id string, data []byte) error {
	return s.put(BucketCrawlJobs, id, data)
}

func (s *BboltStore) GetCrawlJob(_ context.Context, id string) ([]byte, error) {
	return s.get(BucketCrawlJobs, id)
}

func (s *BboltStore) DeleteCrawlJob(_ context.Context, id string) error {
	return s.del(BucketCrawlJobs, id)
}

func (s *BboltStore) ListCrawlJobs(_ context.Context) (map[string][]byte, error) {
	return s.list(BucketCrawlJobs)
}

// --- Agent State ---

func (s *BboltStore) PutAgentState(_ context.Context, id string, data []byte) error {
	return s.put(BucketAgentState, id, data)
}

func (s *BboltStore) GetAgentState(_ context.Context, id string) ([]byte, error) {
	return s.get(BucketAgentState, id)
}

func (s *BboltStore) DeleteAgentState(_ context.Context, id string) error {
	return s.del(BucketAgentState, id)
}

func (s *BboltStore) ListAgentState(_ context.Context) (map[string][]byte, error) {
	return s.list(BucketAgentState)
}

// --- Schema ---

func (s *BboltStore) SchemaVersion(_ context.Context) (int, error) {
	data, err := s.get(BucketMeta, MetaSchemaVersion)
	if err != nil {
		return 0, err
	}
	if data == nil {
		return 0, nil
	}
	var v int
	_, err = fmt.Sscanf(string(data), "%d", &v)
	return v, err
}

func (s *BboltStore) SetSchemaVersion(_ context.Context, version int) error {
	return s.put(BucketMeta, MetaSchemaVersion, []byte(fmt.Sprintf("%d", version)))
}

// --- Metadata ---

func (s *BboltStore) PutMeta(_ context.Context, key string, data []byte) error {
	return s.put(BucketMeta, key, data)
}

func (s *BboltStore) GetMeta(_ context.Context, key string) ([]byte, error) {
	return s.get(BucketMeta, key)
}
