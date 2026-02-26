package state

import "context"

// CurrentSchemaVersion is the latest schema version for the state database.
const CurrentSchemaVersion = 1

// Bucket names used by the bbolt implementation.
const (
	BucketMeta        = "meta"
	BucketDeployments = "deployments"
	BucketBans        = "bans"
	BucketPeers       = "peers"
	BucketCertPins    = "certpins"
	BucketAPIKeys     = "apikeys"
)

// Meta keys stored in the meta bucket.
const (
	MetaSchemaVersion = "schema_version"
	MetaCreatedAt     = "created_at"
	MetaMigratedFrom  = "migrated_from"
)

// StateStore is the persistence interface for all daemon state.
// Values are raw bytes — callers are responsible for marshaling their own types.
// This avoids import cycles between packages.
type StateStore interface {
	Close() error

	// Deployments
	PutDeployment(ctx context.Context, id string, data []byte) error
	GetDeployment(ctx context.Context, id string) ([]byte, error)
	DeleteDeployment(ctx context.Context, id string) error
	ListDeployments(ctx context.Context) (map[string][]byte, error)

	// Bans
	PutBan(ctx context.Context, peerID string, data []byte) error
	DeleteBan(ctx context.Context, peerID string) error
	ListBans(ctx context.Context) (map[string][]byte, error)

	// Peers (address book)
	PutPeer(ctx context.Context, peerID string, data []byte) error
	DeletePeer(ctx context.Context, peerID string) error
	ListPeers(ctx context.Context) (map[string][]byte, error)

	// Certificate Pins
	PutCertPin(ctx context.Context, nodeID string, hash []byte) error
	DeleteCertPin(ctx context.Context, nodeID string) error
	ListCertPins(ctx context.Context) (map[string][]byte, error)

	// API Keys
	PutAPIKey(ctx context.Context, id string, data []byte) error
	DeleteAPIKey(ctx context.Context, id string) error
	ListAPIKeys(ctx context.Context) (map[string][]byte, error)

	// Schema
	SchemaVersion(ctx context.Context) (int, error)
	SetSchemaVersion(ctx context.Context, version int) error

	// Metadata (timestamps, counters, etc.)
	PutMeta(ctx context.Context, key string, data []byte) error
	GetMeta(ctx context.Context, key string) ([]byte, error)
}
