package ingress

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestAPIHandler(t *testing.T, resolver resolverFunc, method VerifyMethod, maxPerDep int) (*CustomDomainHandler, *DomainOwnershipStore) {
	t.Helper()
	store := NewDomainOwnershipStore(time.Hour)
	verifier := NewDomainVerifier(method, "moltbunker.dev", testSecret, resolver)
	return NewDomainVerifyHandler(verifier, store, testSecret, "moltbunker.dev", maxPerDep), store
}

func TestDomainAPI_Challenge_CNAME(t *testing.T) {
	h, _ := newTestAPIHandler(t, fakeResolver{}, MethodCNAME, 0)
	body := `{"host":"app.customer.com","deployment_id":"dep-123"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/ingress/custom-domain/challenge", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp challengeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Method != MethodCNAME || resp.CNAMETarget == "" {
		t.Fatalf("unexpected challenge response: %+v", resp)
	}
}

func TestDomainAPI_Verify_Success(t *testing.T) {
	host, dep := "app.customer.com", "dep-123"
	target := CNAMETarget(testSecret, "moltbunker.dev", host, dep)
	h, store := newTestAPIHandler(t, fakeResolver{cname: target}, MethodCNAME, 0)

	body := `{"host":"app.customer.com","deployment_id":"dep-123","owner_wallet":"0xabc"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/ingress/custom-domain/verify", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if _, ok := store.LookupByHost(host); !ok {
		t.Fatal("verified host not persisted to store")
	}
}

func TestDomainAPI_Verify_Failure(t *testing.T) {
	h, store := newTestAPIHandler(t, fakeResolver{cname: "wrong.moltbunker.dev"}, MethodCNAME, 0)
	body := `{"host":"app.customer.com","deployment_id":"dep-123"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/ingress/custom-domain/verify", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body=%s", rec.Code, rec.Body.String())
	}
	if store.Count() != 0 {
		t.Fatal("failed verification must not persist")
	}
}

func TestDomainAPI_Verify_PerDeploymentCap(t *testing.T) {
	// maxPerDep=1; first host succeeds, second host for same deployment is capped.
	h, store := newTestAPIHandler(t, fakeResolver{}, MethodCNAME, 1)
	// Pre-seed one verified host for dep-123.
	store.Store("a.customer.com", "dep-123", "", MethodCNAME)

	// Point the verifier's resolver at the correct target for b.customer.com so
	// the only reason to reject is the cap, not the DNS proof.
	target := CNAMETarget(testSecret, "moltbunker.dev", "b.customer.com", "dep-123")
	h.verifier = NewDomainVerifier(MethodCNAME, "moltbunker.dev", testSecret, fakeResolver{cname: target})

	body := `{"host":"b.customer.com","deployment_id":"dep-123"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/ingress/custom-domain/verify", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (cap); body=%s", rec.Code, rec.Body.String())
	}
}

func TestDomainAPI_MethodNotAllowed(t *testing.T) {
	h, _ := newTestAPIHandler(t, fakeResolver{}, MethodCNAME, 0)
	req := httptest.NewRequest(http.MethodGet, "/v1/ingress/custom-domain/challenge", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestDomainAPI_BadRequest(t *testing.T) {
	h, _ := newTestAPIHandler(t, fakeResolver{}, MethodCNAME, 0)
	req := httptest.NewRequest(http.MethodPost, "/v1/ingress/custom-domain/challenge", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

// Ensure the verifier honors context (smoke test the ctx plumbing compiles/runs).
func TestDomainVerifier_ContextWired(t *testing.T) {
	host, dep := "app.customer.com", "dep-123"
	target := CNAMETarget(testSecret, "moltbunker.dev", host, dep)
	v := NewDomainVerifier(MethodCNAME, "moltbunker.dev", testSecret, fakeResolver{cname: target})
	if err := v.Verify(context.Background(), host, dep); err != nil {
		t.Fatalf("verify: %v", err)
	}
}
