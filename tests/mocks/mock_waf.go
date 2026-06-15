package mocks

import (
	"net/http"
	"sync"

	"github.com/moltbunker/moltbunker/internal/ingress"
)

// MockWAFEngine implements ingress.WAFEngine for tests so callers never need
// the real Coraza engine or the OWASP CRS files. Configure AlwaysBlock to force
// a blocked verdict and inspect RecordedCalls to assert the engine was invoked.
type MockWAFEngine struct {
	// AlwaysBlock forces every Inspect to return Blocked=true.
	AlwaysBlock bool
	// BlockMode is the mode reported by Mode(); defaults to detection.
	BlockMode ingress.WAFMode
	// MatchRuleIDs are returned as MatchedRules on every inspection.
	MatchRuleIDs []string

	mu            sync.Mutex
	RecordedCalls []string
}

// Inspect records the call and returns a result driven by AlwaysBlock.
func (m *MockWAFEngine) Inspect(r *http.Request, tenantID string) (*ingress.WAFResult, error) {
	m.mu.Lock()
	m.RecordedCalls = append(m.RecordedCalls, tenantID)
	m.mu.Unlock()

	res := &ingress.WAFResult{
		MatchedPhases: make(map[string]int),
		MatchedRules:  append([]string(nil), m.MatchRuleIDs...),
	}
	for _, id := range m.MatchRuleIDs {
		res.MatchedPhases[id] = 2
	}
	if m.AlwaysBlock {
		res.Blocked = true
		res.Phase = 2
	}
	return res, nil
}

// Mode reports the configured block mode (detection by default).
func (m *MockWAFEngine) Mode(string) ingress.WAFMode {
	if m.BlockMode != "" {
		return m.BlockMode
	}
	return ingress.ModeDetection
}

// Calls returns a copy of the recorded tenant IDs.
func (m *MockWAFEngine) Calls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.RecordedCalls...)
}

// Ensure MockWAFEngine satisfies the interface at compile time.
var _ ingress.WAFEngine = (*MockWAFEngine)(nil)
