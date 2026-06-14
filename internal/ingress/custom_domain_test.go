package ingress

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeResolver is an injectable resolverFunc for DNS proof tests (no real DNS).
type fakeResolver struct {
	cname    string
	cnameErr error
	txt      []string
	txtErr   error
}

func (f fakeResolver) LookupCNAME(_ context.Context, _ string) (string, error) {
	return f.cname, f.cnameErr
}
func (f fakeResolver) LookupTXT(_ context.Context, _ string) ([]string, error) {
	return f.txt, f.txtErr
}

var testSecret = []byte("edge-02-test-hmac-secret-not-a-real-credential")

func TestGenerateVerificationToken_Deterministic(t *testing.T) {
	a := GenerateVerificationToken(testSecret, "app.customer.com", "dep-123")
	b := GenerateVerificationToken(testSecret, "app.customer.com", "dep-123")
	if a != b {
		t.Fatalf("token not deterministic: %q != %q", a, b)
	}
	if a == "" {
		t.Fatal("token must not be empty")
	}
}

func TestGenerateVerificationToken_DifferentInputsDiffer(t *testing.T) {
	base := GenerateVerificationToken(testSecret, "app.customer.com", "dep-123")
	if GenerateVerificationToken(testSecret, "evil.customer.com", "dep-123") == base {
		t.Fatal("different host produced same token")
	}
	if GenerateVerificationToken(testSecret, "app.customer.com", "dep-999") == base {
		t.Fatal("different deployment produced same token")
	}
	if GenerateVerificationToken([]byte("other-secret"), "app.customer.com", "dep-123") == base {
		t.Fatal("different secret produced same token")
	}
}

func TestCNAMEVerifier_Accept(t *testing.T) {
	host := "app.customer.com"
	dep := "dep-123"
	target := CNAMETarget(testSecret, "moltbunker.dev", host, dep) + "." // trailing dot like real CNAME
	v := NewDomainVerifier(MethodCNAME, "moltbunker.dev", testSecret, fakeResolver{cname: target})
	if err := v.Verify(context.Background(), host, dep); err != nil {
		t.Fatalf("expected accept, got %v", err)
	}
	if v.Method() != MethodCNAME {
		t.Fatalf("Method = %q", v.Method())
	}
}

func TestCNAMEVerifier_WrongTarget(t *testing.T) {
	v := NewDomainVerifier(MethodCNAME, "moltbunker.dev", testSecret, fakeResolver{cname: "wrong.moltbunker.dev."})
	if err := v.Verify(context.Background(), "app.customer.com", "dep-123"); err == nil {
		t.Fatal("expected error for wrong CNAME target")
	}
}

func TestCNAMEVerifier_LookupError(t *testing.T) {
	v := NewDomainVerifier(MethodCNAME, "moltbunker.dev", testSecret, fakeResolver{cnameErr: errors.New("nxdomain")})
	if err := v.Verify(context.Background(), "app.customer.com", "dep-123"); err == nil {
		t.Fatal("expected lookup error to propagate")
	}
}

func TestTXTVerifier_Accept(t *testing.T) {
	host := "app.customer.com"
	dep := "dep-123"
	_, value := TXTRecord(testSecret, host, dep)
	v := NewDomainVerifier(MethodTXT, "moltbunker.dev", testSecret, fakeResolver{txt: []string{"unrelated", value}})
	if err := v.Verify(context.Background(), host, dep); err != nil {
		t.Fatalf("expected accept, got %v", err)
	}
	if v.Method() != MethodTXT {
		t.Fatalf("Method = %q", v.Method())
	}
}

func TestTXTVerifier_MissingRecord(t *testing.T) {
	v := NewDomainVerifier(MethodTXT, "moltbunker.dev", testSecret, fakeResolver{txt: []string{"some-other-record"}})
	if err := v.Verify(context.Background(), "app.customer.com", "dep-123"); err == nil {
		t.Fatal("expected error when TXT record missing")
	}
}

func TestDomainOwnershipStore_StoreAndLookup(t *testing.T) {
	s := NewDomainOwnershipStore(time.Hour)
	s.Store("App.Customer.com", "dep-123", "0xwallet", MethodCNAME)
	rec, ok := s.LookupByHost("app.customer.com") // case-insensitive
	if !ok {
		t.Fatal("expected lookup hit")
	}
	if rec.DeploymentID != "dep-123" {
		t.Fatalf("DeploymentID = %q", rec.DeploymentID)
	}
	// Lookup with port should normalize.
	if _, ok := s.LookupByHost("app.customer.com:443"); !ok {
		t.Fatal("expected lookup hit with port")
	}
}

func TestDomainOwnershipStore_Expiry(t *testing.T) {
	s := NewDomainOwnershipStore(time.Hour)
	// Manually insert an already-expired record.
	s.mu.Lock()
	s.domains["expired.customer.com"] = VerifiedDomain{
		Host:         "expired.customer.com",
		DeploymentID: "dep-x",
		ExpiresAt:    time.Now().Add(-time.Minute),
	}
	s.mu.Unlock()
	if _, ok := s.LookupByHost("expired.customer.com"); ok {
		t.Fatal("expired record must not be returned")
	}
}

func TestDomainOwnershipStore_Remove(t *testing.T) {
	s := NewDomainOwnershipStore(time.Hour)
	s.Store("app.customer.com", "dep-1", "", MethodCNAME)
	s.Remove("app.customer.com")
	if _, ok := s.LookupByHost("app.customer.com"); ok {
		t.Fatal("record should be gone after Remove")
	}
}

func TestDomainOwnershipStore_CountForDeployment(t *testing.T) {
	s := NewDomainOwnershipStore(time.Hour)
	s.Store("a.customer.com", "dep-1", "", MethodCNAME)
	s.Store("b.customer.com", "dep-1", "", MethodCNAME)
	s.Store("c.customer.com", "dep-2", "", MethodCNAME)
	if got := s.CountForDeployment("dep-1"); got != 2 {
		t.Fatalf("CountForDeployment(dep-1) = %d, want 2", got)
	}
}

func TestStoreIfUnderCap_EnforcesCap(t *testing.T) {
	s := NewDomainOwnershipStore(time.Hour)
	if _, ok := s.StoreIfUnderCap("a.customer.com", "dep-1", "", MethodCNAME, 1); !ok {
		t.Fatal("first host under cap should store")
	}
	if _, ok := s.StoreIfUnderCap("b.customer.com", "dep-1", "", MethodCNAME, 1); ok {
		t.Fatal("second host for same deployment must be capped")
	}
	// A different deployment is unaffected by dep-1's cap.
	if _, ok := s.StoreIfUnderCap("c.customer.com", "dep-2", "", MethodCNAME, 1); !ok {
		t.Fatal("different deployment must not be capped by dep-1")
	}
}

func TestStoreIfUnderCap_RefreshNotCounted(t *testing.T) {
	s := NewDomainOwnershipStore(time.Hour)
	if _, ok := s.StoreIfUnderCap("a.customer.com", "dep-1", "", MethodCNAME, 1); !ok {
		t.Fatal("first store should succeed")
	}
	// Re-verifying the SAME host for the SAME deployment is an in-place refresh
	// and must not be blocked by the cap.
	if _, ok := s.StoreIfUnderCap("a.customer.com", "dep-1", "0xnew", MethodCNAME, 1); !ok {
		t.Fatal("re-verify of existing host must not be capped")
	}
	if s.CountForDeployment("dep-1") != 1 {
		t.Fatalf("refresh must not increase count, got %d", s.CountForDeployment("dep-1"))
	}
}

func TestStoreIfUnderCap_Unlimited(t *testing.T) {
	s := NewDomainOwnershipStore(time.Hour)
	for _, h := range []string{"a.c.com", "b.c.com", "c.c.com"} {
		if _, ok := s.StoreIfUnderCap(h, "dep-1", "", MethodCNAME, 0); !ok {
			t.Fatalf("maxPerDep=0 (unlimited) should always store, failed on %s", h)
		}
	}
	if s.CountForDeployment("dep-1") != 3 {
		t.Fatalf("unlimited cap: count = %d, want 3", s.CountForDeployment("dep-1"))
	}
}

// TestStoreIfUnderCap_AtomicNoTOCTOU hammers the same deployment from many
// goroutines with a cap of N and asserts the cap is never exceeded — proving
// the check+insert is atomic (a separate Count+Store would let multiple
// concurrent callers slip past).
func TestStoreIfUnderCap_AtomicNoTOCTOU(t *testing.T) {
	const cap = 5
	const goroutines = 64
	s := NewDomainOwnershipStore(time.Hour)

	var wg sync.WaitGroup
	var mu sync.Mutex
	stored := 0
	start := make(chan struct{})
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		host := "host" + string(rune('a'+i%26)) + string(rune('0'+i/26)) + ".customer.com"
		go func(h string) {
			defer wg.Done()
			<-start
			if _, ok := s.StoreIfUnderCap(h, "dep-race", "", MethodCNAME, cap); ok {
				mu.Lock()
				stored++
				mu.Unlock()
			}
		}(host)
	}
	close(start)
	wg.Wait()

	if stored > cap {
		t.Fatalf("cap exceeded under concurrency: stored %d > cap %d", stored, cap)
	}
	if got := s.CountForDeployment("dep-race"); got > cap {
		t.Fatalf("store holds %d > cap %d", got, cap)
	}
}
