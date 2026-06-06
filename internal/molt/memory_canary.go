package molt

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"sync"
)

// CanaryMemory is the subset of wazero api.Memory used for canary operations.
// Satisfied by api.Memory, testable with fakes.
type CanaryMemory interface {
	Size() uint32
	Read(offset, byteCount uint32) ([]byte, bool)
	Write(offset uint32, val []byte) bool
}

const (
	// DefaultCanaryCount is the number of canaries per invocation.
	DefaultCanaryCount = 16
	// CanarySize is the byte length of each canary value.
	CanarySize = 32
	// canaryGuardZone is the minimum gap from memory end to avoid overlap.
	canaryGuardZone = 256
)

// canaryEntry tracks a single canary placed in WASM memory.
type canaryEntry struct {
	Offset uint32
	Value  [CanarySize]byte
}

// MemoryCanarySet manages integrity canaries placed in WASM linear memory.
// Canaries are random 32-byte values written at pseudo-random offsets within
// the WASM heap. After invocation, Verify() checks that all values are intact.
//
// Threat model: Tier 2 providers (no SEV-SNP) could read or tamper with
// WASM memory via the host process. Canary verification detects modification.
type MemoryCanarySet struct {
	canaries []canaryEntry
	secret   [32]byte // HMAC key for challenge/response
	mu       sync.Mutex
}

// NewMemoryCanarySet creates a new canary set with the given count.
func NewMemoryCanarySet(count int) (*MemoryCanarySet, error) {
	if count <= 0 {
		count = DefaultCanaryCount
	}

	cs := &MemoryCanarySet{
		canaries: make([]canaryEntry, count),
	}

	// Generate HMAC secret for challenge/response
	if _, err := rand.Read(cs.secret[:]); err != nil {
		return nil, fmt.Errorf("generating canary secret: %w", err)
	}

	// Pre-generate canary values
	for i := range cs.canaries {
		if _, err := rand.Read(cs.canaries[i].Value[:]); err != nil {
			return nil, fmt.Errorf("generating canary %d: %w", i, err)
		}
	}

	return cs, nil
}

// Plant writes all canaries into the given WASM memory at pseudo-random offsets.
// Offsets are chosen to avoid the first 64KB (stack/globals) and the last 256
// bytes (guard zone). Returns false if memory is too small for placement.
func (cs *MemoryCanarySet) Plant(mem CanaryMemory) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	memSize := mem.Size()
	if memSize == 0 {
		return false
	}

	// Usable range: after first page (64KB) to before guard zone
	const minOffset uint32 = 65536 // 1 WASM page
	maxOffset := memSize - canaryGuardZone - CanarySize
	if maxOffset <= minOffset {
		return false // memory too small
	}

	usableRange := maxOffset - minOffset
	// #nosec G115 -- canary count is a small fixed-size slice length, fits in uint32
	if uint32(len(cs.canaries))*CanarySize > usableRange {
		return false // not enough room for all canaries
	}

	// Distribute canaries evenly across usable range with random jitter
	// #nosec G115 -- canary count is a small fixed-size slice length, fits in uint32
	stride := usableRange / uint32(len(cs.canaries))
	for i := range cs.canaries {
		// Deterministic base + random jitter within stride
		var jitter [4]byte
		rand.Read(jitter[:]) //nolint:errcheck
		j := (uint32(jitter[0])<<8 | uint32(jitter[1])) % (stride - CanarySize)
		offset := minOffset + uint32(i)*stride + j

		cs.canaries[i].Offset = offset
		if !mem.Write(offset, cs.canaries[i].Value[:]) {
			return false
		}
	}

	return true
}

// Verify reads back all canary values from WASM memory and checks integrity.
// Returns the number of intact canaries and the number of violations.
func (cs *MemoryCanarySet) Verify(mem CanaryMemory) (intact, violated int) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for _, c := range cs.canaries {
		if c.Offset == 0 {
			continue // not planted
		}
		data, ok := mem.Read(c.Offset, CanarySize)
		if !ok {
			violated++
			continue
		}
		if hmac.Equal(data, c.Value[:]) {
			intact++
		} else {
			violated++
		}
	}
	return intact, violated
}

// ChallengeResponse computes an HMAC-SHA256 response for a specific canary.
// Used by the challenge/response protocol: verifier sends (index, nonce),
// provider reads canary from memory and returns HMAC(canary_value, nonce).
func (cs *MemoryCanarySet) ChallengeResponse(mem CanaryMemory, index int, nonce []byte) ([]byte, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if index < 0 || index >= len(cs.canaries) {
		return nil, fmt.Errorf("canary index %d out of range [0, %d)", index, len(cs.canaries))
	}

	c := cs.canaries[index]
	if c.Offset == 0 {
		return nil, fmt.Errorf("canary %d not planted", index)
	}

	// Read current canary value from memory
	data, ok := mem.Read(c.Offset, CanarySize)
	if !ok {
		return nil, fmt.Errorf("failed to read canary %d at offset %d", index, c.Offset)
	}

	// HMAC(secret || canary_value, nonce)
	mac := hmac.New(sha256.New, cs.secret[:])
	mac.Write(data)
	mac.Write(nonce)
	return mac.Sum(nil), nil
}

// ExpectedResponse computes the expected HMAC for a canary challenge
// using the originally planted value (not read from memory).
// Used by the verifier to check the provider's response.
func (cs *MemoryCanarySet) ExpectedResponse(index int, nonce []byte) ([]byte, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if index < 0 || index >= len(cs.canaries) {
		return nil, fmt.Errorf("canary index %d out of range", index)
	}

	mac := hmac.New(sha256.New, cs.secret[:])
	mac.Write(cs.canaries[index].Value[:])
	mac.Write(nonce)
	return mac.Sum(nil), nil
}

// Count returns the number of canaries in this set.
func (cs *MemoryCanarySet) Count() int {
	return len(cs.canaries)
}
