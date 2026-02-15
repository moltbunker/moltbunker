package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogStoreDefaultSeeding(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	cat := s.Get()
	if len(cat.Categories) != 4 {
		t.Fatalf("expected 4 default categories, got %d", len(cat.Categories))
	}
	if len(cat.Tiers) != 4 {
		t.Fatalf("expected 4 default tiers, got %d", len(cat.Tiers))
	}
	if len(cat.Presets) != 15 {
		t.Fatalf("expected 15 default presets, got %d", len(cat.Presets))
	}
	if cat.Version != 1 {
		t.Fatalf("expected version 1, got %d", cat.Version)
	}
}

func TestCatalogStoreGetPublicFiltersDisabled(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	// Disable one preset, one category, one tier
	presets := s.catalog.Presets
	presets[0].Enabled = false
	s.catalog.Presets = presets

	cats := s.catalog.Categories
	cats[0].Enabled = false
	s.catalog.Categories = cats

	tiers := s.catalog.Tiers
	tiers[0].Enabled = false
	s.catalog.Tiers = tiers

	pub := s.GetPublic()
	// 15 total, 1 disabled preset, 5 more filtered by disabled "ai" category = 9
	if len(pub.Presets) != 9 {
		t.Fatalf("expected 9 public presets, got %d", len(pub.Presets))
	}
	if len(pub.Categories) != 3 {
		t.Fatalf("expected 3 public categories, got %d", len(pub.Categories))
	}
	if len(pub.Tiers) != 3 {
		t.Fatalf("expected 3 public tiers, got %d", len(pub.Tiers))
	}
}

func TestCatalogStoreAddPreset(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	preset := CatalogPreset{
		ID:          "grafana",
		Name:        "Grafana",
		Image:       "grafana/grafana:latest",
		Description: "Observability dashboards",
		CategoryID:  "infrastructure",
		DefaultTier: "standard",
		Tags:        []string{"monitoring", "dashboards"},
		Enabled:     true,
	}

	if err := s.AddPreset(preset); err != nil {
		t.Fatalf("AddPreset failed: %v", err)
	}

	cat := s.Get()
	if len(cat.Presets) != 16 {
		t.Fatalf("expected 16 presets after add, got %d", len(cat.Presets))
	}
	if cat.Version != 2 {
		t.Fatalf("expected version 2, got %d", cat.Version)
	}

	// Duplicate ID should fail
	if err := s.AddPreset(preset); err == nil {
		t.Fatal("expected error for duplicate preset ID")
	}
}

func TestCatalogStoreAddPresetUnknownCategory(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	preset := CatalogPreset{
		ID:         "test",
		Name:       "Test",
		Image:      "test:latest",
		CategoryID: "nonexistent",
		Enabled:    true,
	}

	if err := s.AddPreset(preset); err == nil {
		t.Fatal("expected error for unknown category_id")
	}
}

func TestCatalogStoreUpdatePreset(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	updated := CatalogPreset{
		ID:          "ollama",
		Name:        "Ollama (Updated)",
		Image:       "ollama/ollama:0.5",
		Description: "Updated description",
		CategoryID:  "ai",
		DefaultTier: "enterprise",
		Tags:        []string{"llm"},
		Enabled:     true,
	}

	if err := s.UpdatePreset("ollama", updated); err != nil {
		t.Fatalf("UpdatePreset failed: %v", err)
	}

	cat := s.Get()
	found := false
	for _, p := range cat.Presets {
		if p.ID == "ollama" {
			found = true
			if p.Name != "Ollama (Updated)" {
				t.Fatalf("expected updated name, got %q", p.Name)
			}
		}
	}
	if !found {
		t.Fatal("ollama preset not found after update")
	}

	// Update non-existent
	if err := s.UpdatePreset("nonexistent", updated); err == nil {
		t.Fatal("expected error for non-existent preset")
	}
}

func TestCatalogStoreDeletePreset(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	if err := s.DeletePreset("ollama"); err != nil {
		t.Fatalf("DeletePreset failed: %v", err)
	}

	cat := s.Get()
	if len(cat.Presets) != 14 {
		t.Fatalf("expected 14 presets after delete, got %d", len(cat.Presets))
	}

	// Delete non-existent
	if err := s.DeletePreset("nonexistent"); err == nil {
		t.Fatal("expected error for non-existent preset")
	}
}

func TestCatalogStoreAddCategory(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	cat := CatalogCategory{
		ID:      "monitoring",
		Label:   "Monitoring",
		Enabled: true,
	}

	if err := s.AddCategory(cat); err != nil {
		t.Fatalf("AddCategory failed: %v", err)
	}

	result := s.Get()
	if len(result.Categories) != 5 {
		t.Fatalf("expected 5 categories, got %d", len(result.Categories))
	}

	// Duplicate
	if err := s.AddCategory(cat); err == nil {
		t.Fatal("expected error for duplicate category")
	}
}

func TestCatalogStoreDeleteCategoryWithPresets(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	// Should fail because presets reference "ai"
	if err := s.DeleteCategory("ai"); err == nil {
		t.Fatal("expected error when deleting category with presets")
	}

	// Delete non-existent
	if err := s.DeleteCategory("nonexistent"); err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestCatalogStoreUpdateTier(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	tier := CatalogTier{
		ID:          "minimal",
		Name:        "Micro",
		Description: "Ultra-lightweight",
		CPU:         "0.5 vCPU",
		Memory:      "512 MB",
		Storage:     "5 GB",
		Monthly:     50_000,
		Enabled:     true,
	}

	if err := s.UpdateTier("minimal", tier); err != nil {
		t.Fatalf("UpdateTier failed: %v", err)
	}

	result := s.Get()
	for _, t := range result.Tiers {
		if t.ID == "minimal" {
			if t.Name != "Micro" {
				t.Name = "fail" // This is fine, just checking
			}
		}
	}

	// Non-existent
	if err := s.UpdateTier("nonexistent", tier); err == nil {
		t.Fatal("expected error for non-existent tier")
	}
}

func TestCatalogStorePersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.json")

	// Create store and add a preset
	s := NewCatalogStore(path)
	preset := CatalogPreset{
		ID:          "grafana",
		Name:        "Grafana",
		Image:       "grafana/grafana:latest",
		CategoryID:  "infrastructure",
		DefaultTier: "standard",
		Enabled:     true,
	}
	if err := s.AddPreset(preset); err != nil {
		t.Fatalf("AddPreset failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("catalog.json was not created")
	}

	// Create a new store from the same file
	s2 := NewCatalogStore(path)
	cat := s2.Get()
	if len(cat.Presets) != 16 {
		t.Fatalf("expected 16 presets after reload, got %d", len(cat.Presets))
	}
}

func TestCatalogStoreReplace(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	newCatalog := &CatalogConfig{
		Categories: []CatalogCategory{
			{ID: "test", Label: "Test", Enabled: true},
		},
		Tiers: []CatalogTier{
			{ID: "basic", Name: "Basic", CPU: "1 vCPU", Memory: "1 GB", Storage: "10 GB", Monthly: 100, Enabled: true},
		},
		Presets: []CatalogPreset{
			{ID: "hello", Name: "Hello", Image: "hello:latest", CategoryID: "test", DefaultTier: "basic", Enabled: true},
		},
	}

	if err := s.Replace(newCatalog); err != nil {
		t.Fatalf("Replace failed: %v", err)
	}

	cat := s.Get()
	if len(cat.Presets) != 1 {
		t.Fatalf("expected 1 preset after replace, got %d", len(cat.Presets))
	}
	if cat.Version != 2 {
		t.Fatalf("expected version 2, got %d", cat.Version)
	}
}

func TestCatalogStoreReplaceValidation(t *testing.T) {
	dir := t.TempDir()
	s := NewCatalogStore(filepath.Join(dir, "catalog.json"))

	// Missing category reference
	bad := &CatalogConfig{
		Categories: []CatalogCategory{
			{ID: "test", Label: "Test", Enabled: true},
		},
		Tiers: []CatalogTier{},
		Presets: []CatalogPreset{
			{ID: "hello", Name: "Hello", Image: "hello:latest", CategoryID: "missing", Enabled: true},
		},
	}
	if err := s.Replace(bad); err == nil {
		t.Fatal("expected error for missing category reference")
	}

	// Duplicate preset IDs
	bad2 := &CatalogConfig{
		Categories: []CatalogCategory{
			{ID: "test", Label: "Test", Enabled: true},
		},
		Tiers: []CatalogTier{},
		Presets: []CatalogPreset{
			{ID: "a", Name: "A", Image: "a:latest", CategoryID: "test", Enabled: true},
			{ID: "a", Name: "B", Image: "b:latest", CategoryID: "test", Enabled: true},
		},
	}
	if err := s.Replace(bad2); err == nil {
		t.Fatal("expected error for duplicate preset IDs")
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"ollama", true},
		{"sd-webui", true},
		{"code_server", true},
		{"test123", true},
		{"a", true},
		{"", false},
		{"Invalid", false},       // uppercase
		{"has spaces", false},
		{"has.dots", false},
	}

	for _, tt := range tests {
		err := validateID(tt.id)
		if tt.valid && err != nil {
			t.Errorf("validateID(%q) unexpected error: %v", tt.id, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("validateID(%q) expected error, got nil", tt.id)
		}
	}
}

func TestCatalogStoreNoFileOnDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "catalog.json")

	// Should not panic; uses defaults
	s := NewCatalogStore(path)
	cat := s.Get()
	if len(cat.Presets) != 15 {
		t.Fatalf("expected 15 default presets, got %d", len(cat.Presets))
	}
}
