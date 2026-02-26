package state

import "context"

// CurrentSchemaVersion is the latest schema version for the state database.
const CurrentSchemaVersion = 2

// Bucket names used by the bbolt implementation.
const (
	BucketMeta        = "meta"
	BucketDeployments = "deployments"
	BucketBans        = "bans"
	BucketPeers       = "peers"
	BucketCertPins    = "certpins"
	BucketAPIKeys     = "apikeys"

	// P0 service buckets (added in schema v2)
	BucketStorageBuckets = "storage_buckets"
	BucketStorageObjects = "storage_objects"
	BucketProxySessions  = "proxy_sessions"
	BucketCrawlJobs      = "crawl_jobs"
	BucketAgentState     = "agent_state"
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

	// Storage Buckets (P0 object storage)
	PutStorageBucket(ctx context.Context, name string, data []byte) error
	GetStorageBucket(ctx context.Context, name string) ([]byte, error)
	DeleteStorageBucket(ctx context.Context, name string) error
	ListStorageBuckets(ctx context.Context) (map[string][]byte, error)

	// Storage Objects (P0 object storage)
	PutStorageObject(ctx context.Context, key string, data []byte) error
	GetStorageObject(ctx context.Context, key string) ([]byte, error)
	DeleteStorageObject(ctx context.Context, key string) error
	ListStorageObjects(ctx context.Context) (map[string][]byte, error)

	// Proxy Sessions (P0 decentralized proxy)
	PutProxySession(ctx context.Context, id string, data []byte) error
	GetProxySession(ctx context.Context, id string) ([]byte, error)
	DeleteProxySession(ctx context.Context, id string) error
	ListProxySessions(ctx context.Context) (map[string][]byte, error)

	// Crawl Jobs (P0 web crawling)
	PutCrawlJob(ctx context.Context, id string, data []byte) error
	GetCrawlJob(ctx context.Context, id string) ([]byte, error)
	DeleteCrawlJob(ctx context.Context, id string) error
	ListCrawlJobs(ctx context.Context) (map[string][]byte, error)

	// Agent State (P0 AI agent runtime)
	PutAgentState(ctx context.Context, id string, data []byte) error
	GetAgentState(ctx context.Context, id string) ([]byte, error)
	DeleteAgentState(ctx context.Context, id string) error
	ListAgentState(ctx context.Context) (map[string][]byte, error)

	// Schema
	SchemaVersion(ctx context.Context) (int, error)
	SetSchemaVersion(ctx context.Context, version int) error

	// Metadata (timestamps, counters, etc.)
	PutMeta(ctx context.Context, key string, data []byte) error
	GetMeta(ctx context.Context, key string) ([]byte, error)
}
