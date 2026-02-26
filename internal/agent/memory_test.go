package agent

import (
	"testing"
)

func TestMemory_PutAndGet(t *testing.T) {
	m := NewMemoryStore()

	if err := m.Put("agent-1", "context", "important data"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	entry, err := m.Get("agent-1", "context")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if entry.Value != "important data" {
		t.Errorf("value = %q, want 'important data'", entry.Value)
	}
	if entry.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestMemory_GetNotFound(t *testing.T) {
	m := NewMemoryStore()

	_, err := m.Get("agent-1", "missing")
	if err == nil {
		t.Error("expected error for missing agent")
	}

	m.Put("agent-1", "key", "val")
	_, err = m.Get("agent-1", "missing-key")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestMemory_Overwrite(t *testing.T) {
	m := NewMemoryStore()

	m.Put("a1", "key", "v1")
	m.Put("a1", "key", "v2")

	entry, _ := m.Get("a1", "key")
	if entry.Value != "v2" {
		t.Errorf("value = %q, want v2", entry.Value)
	}
}

func TestMemory_Delete(t *testing.T) {
	m := NewMemoryStore()

	m.Put("a1", "key", "val")
	m.Delete("a1", "key")

	_, err := m.Get("a1", "key")
	if err == nil {
		t.Error("key should be deleted")
	}
}

func TestMemory_List(t *testing.T) {
	m := NewMemoryStore()

	m.Put("a1", "k1", "v1")
	m.Put("a1", "k2", "v2")
	m.Put("a2", "k1", "v3")

	entries := m.List("a1")
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}

	entries = m.List("a2")
	if len(entries) != 1 {
		t.Fatalf("a2 entries = %d, want 1", len(entries))
	}

	entries = m.List("nonexistent")
	if entries != nil {
		t.Errorf("nonexistent should return nil, got %d entries", len(entries))
	}
}

func TestMemory_Clear(t *testing.T) {
	m := NewMemoryStore()

	m.Put("a1", "k1", "v1")
	m.Put("a1", "k2", "v2")
	m.Clear("a1")

	if m.Size("a1") != 0 {
		t.Error("size should be 0 after clear")
	}
}

func TestMemory_Size(t *testing.T) {
	m := NewMemoryStore()

	if m.Size("a1") != 0 {
		t.Error("empty agent should have size 0")
	}

	m.Put("a1", "k1", "v1")
	m.Put("a1", "k2", "v2")

	if m.Size("a1") != 2 {
		t.Errorf("size = %d, want 2", m.Size("a1"))
	}
}

func TestMemory_EmptyKeyOrAgent(t *testing.T) {
	m := NewMemoryStore()

	if err := m.Put("", "key", "val"); err == nil {
		t.Error("expected error for empty agent ID")
	}
	if err := m.Put("a1", "", "val"); err == nil {
		t.Error("expected error for empty key")
	}
}

func TestMemory_GetReturnsCopy(t *testing.T) {
	m := NewMemoryStore()
	m.Put("a1", "key", "original")

	entry, _ := m.Get("a1", "key")
	entry.Value = "mutated"

	original, _ := m.Get("a1", "key")
	if original.Value != "original" {
		t.Error("mutation of copy should not affect original")
	}
}
