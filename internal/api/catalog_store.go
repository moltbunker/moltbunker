package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// Catalog validation limits
const (
	maxCatalogPresets    = 200
	maxCatalogCategories = 50
	maxCatalogTiers      = 20
	maxPresetNameLen     = 128
	maxPresetDescLen     = 512
	maxPresetImageLen    = 256
	maxPresetTags        = 20
	maxTagLen            = 64
	maxCategoryLabelLen  = 128
	maxTierNameLen       = 128
	maxTierDescLen       = 256
	maxTierSpecLen       = 64
)

// CatalogPreset represents a deployable image preset
type CatalogPreset struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Image       string   `json:"image"`
	Description string   `json:"description"`
	CategoryID  string   `json:"category_id"`
	DefaultTier string   `json:"default_tier"`
	Tags        []string `json:"tags"`
	Enabled     bool     `json:"enabled"`
	SortOrder   int      `json:"sort_order"`
}

// CatalogCategory represents a grouping of presets
type CatalogCategory struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sort_order"`
}

// CatalogTier represents a resource tier
type CatalogTier struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CPU         string `json:"cpu"`
	Memory      string `json:"memory"`
	Storage     string `json:"storage"`
	Monthly     int64  `json:"monthly"`
	Enabled     bool   `json:"enabled"`
	Popular     bool   `json:"popular"`
	SortOrder   int    `json:"sort_order"`
}

// CatalogConfig holds the complete deploy catalog
type CatalogConfig struct {
	Presets    []CatalogPreset   `json:"presets"`
	Categories []CatalogCategory `json:"categories"`
	Tiers      []CatalogTier     `json:"tiers"`
	UpdatedAt  time.Time         `json:"updated_at"`
	UpdatedBy  string            `json:"updated_by"`
	Version    int               `json:"version"`
}

// CatalogStore persists the deploy catalog to a JSON file
type CatalogStore struct {
	mu       sync.RWMutex
	catalog  *CatalogConfig
	filePath string
}

// NewCatalogStore creates a store, loading from disk or using defaults
func NewCatalogStore(filePath string) *CatalogStore {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		logging.Warn("failed to create catalog directory",
			"dir", dir,
			"error", err.Error(),
			logging.Component("api"))
	}

	s := &CatalogStore{
		catalog:  defaultCatalog(),
		filePath: filePath,
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		logging.Warn("failed to load catalog",
			"error", err.Error(),
			logging.Component("api"))
	}
	return s
}

// load reads catalog from disk
func (s *CatalogStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var catalog CatalogConfig
	if err := json.Unmarshal(data, &catalog); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.catalog = &catalog
	return nil
}

// saveLocked writes catalog to disk.
// MUST be called while s.mu is held.
func (s *CatalogStore) saveLocked() error {
	data, err := json.MarshalIndent(s.catalog, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0600)
}

// Get returns a deep copy of the full catalog (including disabled items)
func (s *CatalogStore) Get() *CatalogConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.Marshal(s.catalog)
	if err != nil {
		cp := *s.catalog
		return &cp
	}
	var cp CatalogConfig
	if err := json.Unmarshal(data, &cp); err != nil {
		shallow := *s.catalog
		return &shallow
	}
	return &cp
}

// GetPublic returns a filtered catalog containing only enabled items
func (s *CatalogStore) GetPublic() *CatalogConfig {
	full := s.Get()

	// Build set of enabled categories and tiers
	enabledCats := make(map[string]bool, len(full.Categories))
	categories := make([]CatalogCategory, 0, len(full.Categories))
	for _, c := range full.Categories {
		if c.Enabled {
			categories = append(categories, c)
			enabledCats[c.ID] = true
		}
	}

	// Only include presets whose category is also enabled
	presets := make([]CatalogPreset, 0, len(full.Presets))
	for _, p := range full.Presets {
		if p.Enabled && enabledCats[p.CategoryID] {
			presets = append(presets, p)
		}
	}

	tiers := make([]CatalogTier, 0, len(full.Tiers))
	for _, t := range full.Tiers {
		if t.Enabled {
			tiers = append(tiers, t)
		}
	}

	return &CatalogConfig{
		Presets:    presets,
		Categories: categories,
		Tiers:      tiers,
		UpdatedAt:  full.UpdatedAt,
		Version:    full.Version,
	}
}

// Replace replaces the entire catalog and persists to disk
func (s *CatalogStore) Replace(catalog *CatalogConfig) error {
	if err := validateCatalog(catalog); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	catalog.UpdatedAt = time.Now()
	catalog.Version = s.catalog.Version + 1
	s.catalog = catalog
	return s.saveLocked()
}

// AddPreset adds a preset and persists to disk
func (s *CatalogStore) AddPreset(preset CatalogPreset) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validatePreset(&preset); err != nil {
		return err
	}

	// Check for duplicate ID
	for _, p := range s.catalog.Presets {
		if p.ID == preset.ID {
			return fmt.Errorf("duplicate preset ID: %q", preset.ID)
		}
	}

	if len(s.catalog.Presets) >= maxCatalogPresets {
		return fmt.Errorf("too many presets (max %d)", maxCatalogPresets)
	}

	// Validate category reference
	if !s.hasCategoryLocked(preset.CategoryID) {
		return fmt.Errorf("unknown category_id: %q", preset.CategoryID)
	}

	// Validate tier reference
	if preset.DefaultTier != "" && !s.hasTierLocked(preset.DefaultTier) {
		return fmt.Errorf("unknown default_tier: %q", preset.DefaultTier)
	}

	s.catalog.Presets = append(s.catalog.Presets, preset)
	s.catalog.UpdatedAt = time.Now()
	s.catalog.Version++
	return s.saveLocked()
}

// UpdatePreset updates a preset by ID and persists to disk
func (s *CatalogStore) UpdatePreset(id string, preset CatalogPreset) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validatePreset(&preset); err != nil {
		return err
	}

	// Validate category reference
	if !s.hasCategoryLocked(preset.CategoryID) {
		return fmt.Errorf("unknown category_id: %q", preset.CategoryID)
	}

	// Validate tier reference
	if preset.DefaultTier != "" && !s.hasTierLocked(preset.DefaultTier) {
		return fmt.Errorf("unknown default_tier: %q", preset.DefaultTier)
	}

	for i, p := range s.catalog.Presets {
		if p.ID == id {
			preset.ID = id // ID cannot be changed
			s.catalog.Presets[i] = preset
			s.catalog.UpdatedAt = time.Now()
			s.catalog.Version++
			return s.saveLocked()
		}
	}
	return fmt.Errorf("preset not found: %q", id)
}

// DeletePreset removes a preset by ID and persists to disk
func (s *CatalogStore) DeletePreset(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.catalog.Presets {
		if p.ID == id {
			s.catalog.Presets = append(s.catalog.Presets[:i], s.catalog.Presets[i+1:]...)
			s.catalog.UpdatedAt = time.Now()
			s.catalog.Version++
			return s.saveLocked()
		}
	}
	return fmt.Errorf("preset not found: %q", id)
}

// AddCategory adds a category and persists to disk
func (s *CatalogStore) AddCategory(cat CatalogCategory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateCategory(&cat); err != nil {
		return err
	}

	for _, c := range s.catalog.Categories {
		if c.ID == cat.ID {
			return fmt.Errorf("duplicate category ID: %q", cat.ID)
		}
	}

	if len(s.catalog.Categories) >= maxCatalogCategories {
		return fmt.Errorf("too many categories (max %d)", maxCatalogCategories)
	}

	s.catalog.Categories = append(s.catalog.Categories, cat)
	s.catalog.UpdatedAt = time.Now()
	s.catalog.Version++
	return s.saveLocked()
}

// UpdateCategory updates a category by ID and persists to disk
func (s *CatalogStore) UpdateCategory(id string, cat CatalogCategory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateCategory(&cat); err != nil {
		return err
	}

	for i, c := range s.catalog.Categories {
		if c.ID == id {
			cat.ID = id
			s.catalog.Categories[i] = cat
			s.catalog.UpdatedAt = time.Now()
			s.catalog.Version++
			return s.saveLocked()
		}
	}
	return fmt.Errorf("category not found: %q", id)
}

// DeleteCategory removes a category by ID and persists to disk.
// Fails if any preset references this category.
func (s *CatalogStore) DeleteCategory(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for referencing presets
	for _, p := range s.catalog.Presets {
		if p.CategoryID == id {
			return fmt.Errorf("category %q is referenced by preset %q", id, p.ID)
		}
	}

	for i, c := range s.catalog.Categories {
		if c.ID == id {
			s.catalog.Categories = append(s.catalog.Categories[:i], s.catalog.Categories[i+1:]...)
			s.catalog.UpdatedAt = time.Now()
			s.catalog.Version++
			return s.saveLocked()
		}
	}
	return fmt.Errorf("category not found: %q", id)
}

// UpdateTier updates a tier by ID and persists to disk
func (s *CatalogStore) UpdateTier(id string, tier CatalogTier) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateTier(&tier); err != nil {
		return err
	}

	for i, t := range s.catalog.Tiers {
		if t.ID == id {
			tier.ID = id
			s.catalog.Tiers[i] = tier
			s.catalog.UpdatedAt = time.Now()
			s.catalog.Version++
			return s.saveLocked()
		}
	}
	return fmt.Errorf("tier not found: %q", id)
}

// hasCategoryLocked checks if a category ID exists. Must hold s.mu.
func (s *CatalogStore) hasCategoryLocked(id string) bool {
	for _, c := range s.catalog.Categories {
		if c.ID == id {
			return true
		}
	}
	return false
}

// hasTierLocked checks if a tier ID exists. Must hold s.mu.
func (s *CatalogStore) hasTierLocked(id string) bool {
	for _, t := range s.catalog.Tiers {
		if t.ID == id {
			return true
		}
	}
	return false
}

// ── Validation ───────────────────────────────────────────────────────────────

var catalogIDPattern = adminNodeIDPattern // reuse: 8-64 hex chars is too strict; override below

func init() {
	// Allow alphanumeric + hyphens for catalog IDs (e.g., "code-server", "sd-webui")
	// Minimum 1 char, maximum 64 chars
	catalogIDPattern = nil // unused, we validate inline
}

func validateID(id string) error {
	if len(id) == 0 || len(id) > 64 {
		return fmt.Errorf("id must be 1-64 characters")
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("id must be lowercase alphanumeric, hyphens, or underscores")
		}
	}
	return nil
}

func validatePreset(p *CatalogPreset) error {
	if err := validateID(p.ID); err != nil {
		return fmt.Errorf("preset %s", err)
	}
	if len(p.Name) == 0 || len(p.Name) > maxPresetNameLen {
		return fmt.Errorf("preset name must be 1-%d characters", maxPresetNameLen)
	}
	if len(p.Image) == 0 || len(p.Image) > maxPresetImageLen {
		return fmt.Errorf("preset image must be 1-%d characters", maxPresetImageLen)
	}
	if len(p.Description) > maxPresetDescLen {
		return fmt.Errorf("preset description too long (max %d)", maxPresetDescLen)
	}
	if len(p.Tags) > maxPresetTags {
		return fmt.Errorf("too many tags (max %d)", maxPresetTags)
	}
	for _, tag := range p.Tags {
		if len(tag) > maxTagLen {
			return fmt.Errorf("tag too long (max %d)", maxTagLen)
		}
	}
	return nil
}

func validateCategory(c *CatalogCategory) error {
	if err := validateID(c.ID); err != nil {
		return fmt.Errorf("category %s", err)
	}
	if len(c.Label) == 0 || len(c.Label) > maxCategoryLabelLen {
		return fmt.Errorf("category label must be 1-%d characters", maxCategoryLabelLen)
	}
	return nil
}

func validateTier(t *CatalogTier) error {
	if err := validateID(t.ID); err != nil {
		return fmt.Errorf("tier %s", err)
	}
	if len(t.Name) == 0 || len(t.Name) > maxTierNameLen {
		return fmt.Errorf("tier name must be 1-%d characters", maxTierNameLen)
	}
	if len(t.Description) > maxTierDescLen {
		return fmt.Errorf("tier description too long (max %d)", maxTierDescLen)
	}
	if len(t.CPU) > maxTierSpecLen || len(t.Memory) > maxTierSpecLen || len(t.Storage) > maxTierSpecLen {
		return fmt.Errorf("tier spec values too long (max %d)", maxTierSpecLen)
	}
	if t.Monthly < 0 {
		return fmt.Errorf("tier monthly cost cannot be negative")
	}
	return nil
}

func validateCatalog(c *CatalogConfig) error {
	if len(c.Presets) > maxCatalogPresets {
		return fmt.Errorf("too many presets (max %d)", maxCatalogPresets)
	}
	if len(c.Categories) > maxCatalogCategories {
		return fmt.Errorf("too many categories (max %d)", maxCatalogCategories)
	}
	if len(c.Tiers) > maxCatalogTiers {
		return fmt.Errorf("too many tiers (max %d)", maxCatalogTiers)
	}

	// Validate each item
	catIDs := make(map[string]bool, len(c.Categories))
	for _, cat := range c.Categories {
		if err := validateCategory(&cat); err != nil {
			return err
		}
		if catIDs[cat.ID] {
			return fmt.Errorf("duplicate category ID: %q", cat.ID)
		}
		catIDs[cat.ID] = true
	}

	tierIDs := make(map[string]bool, len(c.Tiers))
	for _, tier := range c.Tiers {
		if err := validateTier(&tier); err != nil {
			return err
		}
		if tierIDs[tier.ID] {
			return fmt.Errorf("duplicate tier ID: %q", tier.ID)
		}
		tierIDs[tier.ID] = true
	}

	presetIDs := make(map[string]bool, len(c.Presets))
	for _, preset := range c.Presets {
		if err := validatePreset(&preset); err != nil {
			return err
		}
		if presetIDs[preset.ID] {
			return fmt.Errorf("duplicate preset ID: %q", preset.ID)
		}
		presetIDs[preset.ID] = true
		if !catIDs[preset.CategoryID] {
			return fmt.Errorf("preset %q references unknown category %q", preset.ID, preset.CategoryID)
		}
		if preset.DefaultTier != "" && !tierIDs[preset.DefaultTier] {
			return fmt.Errorf("preset %q references unknown tier %q", preset.ID, preset.DefaultTier)
		}
	}

	return nil
}

// ── Default catalog ──────────────────────────────────────────────────────────

func defaultCatalog() *CatalogConfig {
	return &CatalogConfig{
		Categories: []CatalogCategory{
			{ID: "ai", Label: "AI / ML", Enabled: true, SortOrder: 0},
			{ID: "infrastructure", Label: "Infrastructure", Enabled: true, SortOrder: 1},
			{ID: "database", Label: "Database", Enabled: true, SortOrder: 2},
			{ID: "dev", Label: "Developer Tools", Enabled: true, SortOrder: 3},
		},
		Tiers: []CatalogTier{
			{ID: "minimal", Name: "Minimal", Description: "Lightweight tasks & bots", CPU: "1 vCPU", Memory: "1 GB", Storage: "10 GB", Monthly: 100_000, Enabled: true, SortOrder: 0},
			{ID: "standard", Name: "Standard", Description: "Web apps & APIs", CPU: "2 vCPU", Memory: "4 GB", Storage: "50 GB", Monthly: 400_000, Enabled: true, SortOrder: 1},
			{ID: "performance", Name: "Performance", Description: "ML inference & databases", CPU: "4 vCPU", Memory: "8 GB", Storage: "200 GB", Monthly: 1_500_000, Enabled: true, Popular: true, SortOrder: 2},
			{ID: "enterprise", Name: "Enterprise", Description: "Heavy compute & training", CPU: "8 vCPU", Memory: "16 GB", Storage: "500 GB", Monthly: 5_000_000, Enabled: true, SortOrder: 3},
		},
		Presets: []CatalogPreset{
			// AI/ML
			{ID: "ollama", Name: "Ollama", Image: "ollama/ollama:latest", Description: "Run LLMs locally — Llama 3, Mistral, Gemma, Phi", CategoryID: "ai", DefaultTier: "performance", Tags: []string{"llm", "inference", "gpu"}, Enabled: true, SortOrder: 0},
			{ID: "vllm", Name: "vLLM", Image: "vllm/vllm-openai:latest", Description: "High-throughput LLM serving with PagedAttention", CategoryID: "ai", DefaultTier: "enterprise", Tags: []string{"llm", "serving", "gpu"}, Enabled: true, SortOrder: 1},
			{ID: "tgi", Name: "Text Generation Inference", Image: "ghcr.io/huggingface/text-generation-inference:latest", Description: "HuggingFace optimized inference for text generation", CategoryID: "ai", DefaultTier: "performance", Tags: []string{"llm", "inference", "huggingface"}, Enabled: true, SortOrder: 2},
			{ID: "sd-webui", Name: "Stable Diffusion WebUI", Image: "stabilityai/stable-diffusion:latest", Description: "Image generation with AUTOMATIC1111 WebUI", CategoryID: "ai", DefaultTier: "performance", Tags: []string{"image", "gpu", "diffusion"}, Enabled: true, SortOrder: 3},
			{ID: "comfyui", Name: "ComfyUI", Image: "comfyanonymous/comfyui:latest", Description: "Node-based Stable Diffusion workflow editor", CategoryID: "ai", DefaultTier: "performance", Tags: []string{"image", "gpu", "workflow"}, Enabled: true, SortOrder: 4},
			{ID: "jupyter", Name: "Jupyter Lab", Image: "jupyter/scipy-notebook:latest", Description: "Interactive notebooks for data science and ML", CategoryID: "ai", DefaultTier: "standard", Tags: []string{"notebook", "python", "data"}, Enabled: true, SortOrder: 5},
			// Infrastructure
			{ID: "nginx", Name: "Nginx", Image: "nginx:alpine", Description: "High-performance web server and reverse proxy", CategoryID: "infrastructure", DefaultTier: "minimal", Tags: []string{"web", "proxy", "http"}, Enabled: true, SortOrder: 0},
			{ID: "redis", Name: "Redis", Image: "redis:7-alpine", Description: "In-memory data store, cache, and message broker", CategoryID: "infrastructure", DefaultTier: "standard", Tags: []string{"cache", "kv", "pubsub"}, Enabled: true, SortOrder: 1},
			{ID: "minio", Name: "MinIO", Image: "minio/minio:latest", Description: "S3-compatible object storage", CategoryID: "infrastructure", DefaultTier: "standard", Tags: []string{"storage", "s3", "object"}, Enabled: true, SortOrder: 2},
			{ID: "traefik", Name: "Traefik", Image: "traefik:v3.0", Description: "Cloud-native reverse proxy and load balancer", CategoryID: "infrastructure", DefaultTier: "minimal", Tags: []string{"proxy", "lb", "routing"}, Enabled: true, SortOrder: 3},
			// Database
			{ID: "postgres", Name: "PostgreSQL", Image: "postgres:16-alpine", Description: "Advanced open-source relational database", CategoryID: "database", DefaultTier: "standard", Tags: []string{"sql", "relational", "acid"}, Enabled: true, SortOrder: 0},
			{ID: "mongodb", Name: "MongoDB", Image: "mongo:7", Description: "Document-oriented NoSQL database", CategoryID: "database", DefaultTier: "standard", Tags: []string{"nosql", "document", "json"}, Enabled: true, SortOrder: 1},
			{ID: "chromadb", Name: "ChromaDB", Image: "chromadb/chroma:latest", Description: "Open-source vector database for AI embeddings", CategoryID: "database", DefaultTier: "standard", Tags: []string{"vector", "embeddings", "ai"}, Enabled: true, SortOrder: 2},
			// Dev
			{ID: "code-server", Name: "VS Code Server", Image: "codercom/code-server:latest", Description: "Run VS Code in the browser", CategoryID: "dev", DefaultTier: "standard", Tags: []string{"ide", "editor", "browser"}, Enabled: true, SortOrder: 0},
			{ID: "n8n", Name: "n8n", Image: "n8nio/n8n:latest", Description: "Workflow automation platform", CategoryID: "dev", DefaultTier: "standard", Tags: []string{"automation", "workflow", "integration"}, Enabled: true, SortOrder: 1},
		},
		Version: 1,
	}
}
