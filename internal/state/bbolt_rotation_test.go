package state

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"

	bolt "go.etcd.io/bbolt"

	"github.com/moltbunker/moltbunker/internal/security"
)

func TestBboltStore_RotateKey(t *testing.T) {
	ctx := context.Background()
	oldKey := newKey(t)
	s, path := openEncrypted(t, oldKey)

	// Write 10 values across several buckets.
	want := map[string][]byte{}
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("dep-%d", i)
		v := []byte(fmt.Sprintf(`{"id":%q,"secret":"s-%d"}`, id, i))
		if err := s.PutDeployment(ctx, id, v); err != nil {
			t.Fatalf("PutDeployment: %v", err)
		}
		want[id] = v
	}
	if err := s.PutPeer(ctx, "peer-1", []byte("peer-data")); err != nil {
		t.Fatalf("PutPeer: %v", err)
	}

	// Rotate to a new key.
	newKeyBytes := newKey(t)
	m, err := s.RotateKey(ctx, newKeyBytes)
	if err != nil {
		t.Fatalf("RotateKey: %v", err)
	}
	if m.ValuesRotated < 11 {
		t.Errorf("ValuesRotated = %d, want >= 11", m.ValuesRotated)
	}

	// All values readable under the new key (same store, swapped key).
	for id, v := range want {
		got, err := s.GetDeployment(ctx, id)
		if err != nil {
			t.Fatalf("GetDeployment %s: %v", id, err)
		}
		if !bytes.Equal(got, v) {
			t.Errorf("%s mismatch after rotate: got %q want %q", id, got, v)
		}
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// A store opened with the OLD key must fail to decrypt the rotated values.
	// bbolt takes an exclusive file lock, so each reopen must be closed before
	// the next.
	old, err := NewBboltStore(path, oldKey)
	if err != nil {
		t.Fatalf("reopen with old key: %v", err)
	}
	if _, err := old.GetDeployment(ctx, "dep-0"); err == nil {
		_ = old.Close()
		t.Fatal("old key should NOT decrypt rotated values")
	}
	if err := old.Close(); err != nil {
		t.Fatalf("close old: %v", err)
	}

	// A store opened with the NEW key reads them fine.
	fresh, err := NewBboltStore(path, newKeyBytes)
	if err != nil {
		t.Fatalf("reopen with new key: %v", err)
	}
	defer fresh.Close()
	got, err := fresh.GetDeployment(ctx, "dep-3")
	if err != nil {
		t.Fatalf("fresh GetDeployment: %v", err)
	}
	if !bytes.Equal(got, want["dep-3"]) {
		t.Errorf("fresh read mismatch: got %q want %q", got, want["dep-3"])
	}
}

// TestBboltStore_RotateMagicVersion writes a value with the legacy MBENC1 magic
// directly, then verifies RotateKey migrates it to MBENC2.
func TestBboltStore_RotateMagicVersion(t *testing.T) {
	ctx := context.Background()
	key := newKey(t)
	s, path := openEncrypted(t, key)

	// Manually construct an MBENC1-tagged value and write it raw into bbolt.
	plain := []byte(`{"id":"legacy","v":1}`)
	ct, err := security.EncryptAES256GCM(key, plain)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	mbenc1 := append(append([]byte(nil), encMagic...), ct...)
	err = s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(BucketDeployments)).Put([]byte("legacy"), mbenc1)
	})
	if err != nil {
		t.Fatalf("raw put MBENC1: %v", err)
	}

	// Readable before rotation (decode accepts MBENC1).
	got, err := s.GetDeployment(ctx, "legacy")
	if err != nil {
		t.Fatalf("GetDeployment pre-rotate: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("pre-rotate mismatch: got %q want %q", got, plain)
	}

	// Rotate to the SAME key (idempotent re-tag).
	if _, err := s.RotateKey(ctx, key); err != nil {
		t.Fatalf("RotateKey: %v", err)
	}

	// The on-disk value must now carry MBENC2.
	var raw []byte
	err = s.db.View(func(tx *bolt.Tx) error {
		raw = append([]byte(nil), tx.Bucket([]byte(BucketDeployments)).Get([]byte("legacy"))...)
		return nil
	})
	if err != nil {
		t.Fatalf("View: %v", err)
	}
	if !bytes.HasPrefix(raw, encMagic2) {
		t.Fatalf("value not migrated to MBENC2: %x", raw[:8])
	}

	got, err = s.GetDeployment(ctx, "legacy")
	if err != nil {
		t.Fatalf("GetDeployment post-rotate: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("post-rotate mismatch: got %q want %q", got, plain)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	_ = path
}

// TestBboltStore_RotateConcurrent runs parallel puts alongside a RotateKey call
// and asserts no panic / no data loss / consistent reads afterward.
func TestBboltStore_RotateConcurrent(t *testing.T) {
	ctx := context.Background()
	oldKey := newKey(t)
	s, _ := openEncrypted(t, oldKey)
	defer s.Close()

	// Seed some values so rotation has work to do.
	for i := 0; i < 20; i++ {
		if err := s.PutDeployment(ctx, fmt.Sprintf("seed-%d", i), []byte(fmt.Sprintf("v%d", i))); err != nil {
			t.Fatalf("seed put: %v", err)
		}
	}

	newKeyBytes := newKey(t)

	var wg sync.WaitGroup
	errCh := make(chan error, 64)

	// 50 concurrent writers.
	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				id := fmt.Sprintf("c-%d-%d", g, i)
				if err := s.PutDeployment(ctx, id, []byte(id)); err != nil {
					errCh <- err
					return
				}
				if _, err := s.GetDeployment(ctx, id); err != nil {
					errCh <- err
					return
				}
			}
		}(g)
	}

	// Rotate mid-flight.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if _, err := s.RotateKey(ctx, newKeyBytes); err != nil {
			errCh <- err
		}
	}()

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent op failed: %v", err)
	}

	// After rotation, every concurrent value is readable and consistent.
	for g := 0; g < 50; g++ {
		for i := 0; i < 5; i++ {
			id := fmt.Sprintf("c-%d-%d", g, i)
			got, err := s.GetDeployment(ctx, id)
			if err != nil {
				t.Fatalf("post GetDeployment %s: %v", id, err)
			}
			if string(got) != id {
				t.Errorf("%s = %q, want %q", id, got, id)
			}
		}
	}
}
