package molt

import (
	"crypto/hmac"
	"crypto/rand"
	"testing"
)

// TestMemoryCanary_PlantAndVerify tests the basic canary lifecycle:
// create → plant in memory → verify all intact.
func TestMemoryCanary_PlantAndVerify(t *testing.T) {
	cs, err := NewMemoryCanarySet(8)
	if err != nil {
		t.Fatalf("NewMemoryCanarySet: %v", err)
	}
	if cs.Count() != 8 {
		t.Fatalf("Count = %d, want 8", cs.Count())
	}

	fakeMem := newFakeMemory(256 * 1024)
	if !cs.Plant(fakeMem) {
		t.Fatal("Plant failed")
	}

	intact, violated := cs.Verify(fakeMem)
	if violated != 0 {
		t.Fatalf("violated = %d, want 0", violated)
	}
	if intact != 8 {
		t.Fatalf("intact = %d, want 8", intact)
	}
}

// TestMemoryCanary_DefaultCount verifies default canary count.
func TestMemoryCanary_DefaultCount(t *testing.T) {
	cs, err := NewMemoryCanarySet(0) // should use default
	if err != nil {
		t.Fatalf("NewMemoryCanarySet: %v", err)
	}
	if cs.Count() != DefaultCanaryCount {
		t.Fatalf("Count = %d, want %d", cs.Count(), DefaultCanaryCount)
	}
}

// TestMemoryCanary_ChallengeResponse verifies the HMAC challenge/response protocol.
func TestMemoryCanary_ChallengeResponse(t *testing.T) {
	cs, err := NewMemoryCanarySet(4)
	if err != nil {
		t.Fatalf("NewMemoryCanarySet: %v", err)
	}

	// Use a fake memory to test challenge/response
	fakeMem := newFakeMemory(256 * 1024) // 256KB

	if !cs.Plant(fakeMem) {
		t.Fatal("Plant failed")
	}

	// Generate a nonce
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}

	// Get challenge response from "provider" (reads from memory)
	response, err := cs.ChallengeResponse(fakeMem, 0, nonce)
	if err != nil {
		t.Fatalf("ChallengeResponse: %v", err)
	}

	// Get expected response from "verifier" (uses stored values)
	expected, err := cs.ExpectedResponse(0, nonce)
	if err != nil {
		t.Fatalf("ExpectedResponse: %v", err)
	}

	// They should match (canary intact)
	if !hmac.Equal(response, expected) {
		t.Fatal("challenge response mismatch — canary may be corrupted")
	}
}

// TestMemoryCanary_TamperDetection verifies that modifying a canary
// is detected by Verify().
func TestMemoryCanary_TamperDetection(t *testing.T) {
	cs, err := NewMemoryCanarySet(4)
	if err != nil {
		t.Fatalf("NewMemoryCanarySet: %v", err)
	}

	fakeMem := newFakeMemory(256 * 1024)

	if !cs.Plant(fakeMem) {
		t.Fatal("Plant failed")
	}

	// Verify before tampering — all should be intact
	intact, violated := cs.Verify(fakeMem)
	if violated != 0 {
		t.Fatalf("before tampering: intact=%d, violated=%d", intact, violated)
	}
	if intact != 4 {
		t.Fatalf("before tampering: intact=%d, want 4", intact)
	}

	// Tamper with first canary
	offset := cs.canaries[0].Offset
	fakeMem.data[offset] ^= 0xFF // flip bits

	// Verify after tampering — first should be violated
	intact, violated = cs.Verify(fakeMem)
	if violated != 1 {
		t.Fatalf("after tampering: violated=%d, want 1", violated)
	}
	if intact != 3 {
		t.Fatalf("after tampering: intact=%d, want 3", intact)
	}
}

// TestMemoryCanary_ChallengeResponseTampered verifies that HMAC mismatch
// is detected when canary is modified in memory.
func TestMemoryCanary_ChallengeResponseTampered(t *testing.T) {
	cs, err := NewMemoryCanarySet(4)
	if err != nil {
		t.Fatalf("NewMemoryCanarySet: %v", err)
	}

	fakeMem := newFakeMemory(256 * 1024)
	if !cs.Plant(fakeMem) {
		t.Fatal("Plant failed")
	}

	nonce := make([]byte, 16)
	rand.Read(nonce) //nolint:errcheck

	// Tamper with canary 2
	offset := cs.canaries[2].Offset
	fakeMem.data[offset] ^= 0xFF

	// Provider response should differ from expected
	response, err := cs.ChallengeResponse(fakeMem, 2, nonce)
	if err != nil {
		t.Fatalf("ChallengeResponse: %v", err)
	}
	expected, err := cs.ExpectedResponse(2, nonce)
	if err != nil {
		t.Fatalf("ExpectedResponse: %v", err)
	}

	if hmac.Equal(response, expected) {
		t.Fatal("tampered canary should produce different HMAC")
	}
}

// TestMemoryCanary_SmallMemory verifies that Plant returns false
// when memory is too small for canaries.
func TestMemoryCanary_SmallMemory(t *testing.T) {
	cs, err := NewMemoryCanarySet(16)
	if err != nil {
		t.Fatalf("NewMemoryCanarySet: %v", err)
	}

	// 32KB is too small (need at least 64KB + guard + canary space)
	tinyMem := newFakeMemory(32 * 1024)
	if cs.Plant(tinyMem) {
		t.Fatal("Plant should fail on tiny memory")
	}
}

// TestMemoryCanary_OutOfRangeIndex verifies bounds checking on challenge.
func TestMemoryCanary_OutOfRangeIndex(t *testing.T) {
	cs, err := NewMemoryCanarySet(4)
	if err != nil {
		t.Fatalf("NewMemoryCanarySet: %v", err)
	}

	fakeMem := newFakeMemory(256 * 1024)
	cs.Plant(fakeMem)

	nonce := make([]byte, 16)
	if _, err := cs.ChallengeResponse(fakeMem, 99, nonce); err == nil {
		t.Fatal("expected error for out-of-range index")
	}
	if _, err := cs.ChallengeResponse(fakeMem, -1, nonce); err == nil {
		t.Fatal("expected error for negative index")
	}
}

// TestMemoryCanary_UniqueValues verifies all canary values are unique.
func TestMemoryCanary_UniqueValues(t *testing.T) {
	cs, err := NewMemoryCanarySet(16)
	if err != nil {
		t.Fatalf("NewMemoryCanarySet: %v", err)
	}

	seen := make(map[[CanarySize]byte]bool)
	for _, c := range cs.canaries {
		if seen[c.Value] {
			t.Fatal("duplicate canary value detected")
		}
		seen[c.Value] = true
	}
}

// TestMemoryCanary_NonOverlappingOffsets verifies canaries don't overlap.
func TestMemoryCanary_NonOverlappingOffsets(t *testing.T) {
	cs, err := NewMemoryCanarySet(16)
	if err != nil {
		t.Fatalf("NewMemoryCanarySet: %v", err)
	}

	fakeMem := newFakeMemory(512 * 1024)
	if !cs.Plant(fakeMem) {
		t.Fatal("Plant failed")
	}

	// Check no two canaries overlap
	for i := 0; i < len(cs.canaries); i++ {
		for j := i + 1; j < len(cs.canaries); j++ {
			a := cs.canaries[i]
			b := cs.canaries[j]
			// Check [a.Offset, a.Offset+CanarySize) doesn't overlap [b.Offset, b.Offset+CanarySize)
			if a.Offset < b.Offset+CanarySize && b.Offset < a.Offset+CanarySize {
				t.Fatalf("canary %d (offset %d) overlaps canary %d (offset %d)",
					i, a.Offset, j, b.Offset)
			}
		}
	}
}

// fakeMemory implements api.Memory for testing canary operations
// without a full WASM runtime.
type fakeMemory struct {
	data []byte
}

func newFakeMemory(size int) *fakeMemory {
	return &fakeMemory{data: make([]byte, size)}
}

// fakeMemory implements CanaryMemory for testing.
func (m *fakeMemory) Size() uint32 { return uint32(len(m.data)) }

func (m *fakeMemory) Read(offset, length uint32) ([]byte, bool) {
	end := uint64(offset) + uint64(length)
	if end > uint64(len(m.data)) {
		return nil, false
	}
	dst := make([]byte, length)
	copy(dst, m.data[offset:end])
	return dst, true
}

func (m *fakeMemory) Write(offset uint32, val []byte) bool {
	end := uint64(offset) + uint64(len(val))
	if end > uint64(len(m.data)) {
		return false
	}
	copy(m.data[offset:end], val)
	return true
}
