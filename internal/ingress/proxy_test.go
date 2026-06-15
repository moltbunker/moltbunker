package ingress

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/moltbunker/moltbunker/internal/tunnel"
)

// newBlockTestProxy builds a Proxy with an empty resolver (no registered
// services) and a wired blocklist. The forward-path takedown check in
// handleRequest fires before any tunnel dial, so a nil tunnel client is fine
// for asserting the 403 short-circuit.
func newBlockTestProxy(t *testing.T, bl tunnel.BlocklistChecker, cd *DomainOwnershipStore) *Proxy {
	t.Helper()
	resolver := NewResolver(nil, nil)
	p := NewProxy(resolver, nil, "moltbunker.dev")
	p.SetBlocklist(bl)
	if cd != nil {
		p.SetCustomDomains(cd)
	}
	return p
}

// TestProxy_Blocklist_SubdomainForwardPath asserts that a blocked subdomain is
// rejected with 403 at the ingress proxy itself — i.e. on the primary forward
// path, before resolver.Resolve / OpenTunnel are ever consulted. EDGE-02.
func TestProxy_Blocklist_SubdomainForwardPath(t *testing.T) {
	bl := tunnel.NewBlocklist()
	bl.Block("a1b2c3d4", "abuse")
	p := newBlockTestProxy(t, bl, nil)

	req := httptest.NewRequest(http.MethodGet, "http://a1b2c3d4.moltbunker.dev/", nil)
	rec := httptest.NewRecorder()
	p.handleRequest(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("blocked subdomain: status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

// TestProxy_Blocklist_CustomHost asserts that blocking the customer's OWN
// hostname (the natural takedown target) severs the deployment on the forward
// path, even though routing internally keys on the deployment ID. EDGE-02.
func TestProxy_Blocklist_CustomHost(t *testing.T) {
	cd := NewDomainOwnershipStore(time.Hour)
	cd.Store("app.customer.com", "dep-abc12345", "", MethodCNAME)

	bl := tunnel.NewBlocklist()
	bl.Block("app.customer.com", "DMCA") // block by custom host, not deployment ID
	p := newBlockTestProxy(t, bl, cd)

	req := httptest.NewRequest(http.MethodGet, "http://app.customer.com/", nil)
	req.Host = "app.customer.com"
	rec := httptest.NewRecorder()
	p.handleRequest(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("blocked custom host: status = %d, want 403; body=%s", rec.Code, rec.Body.String())
	}
}

// TestProxy_Blocklist_NotBlockedPassesThrough confirms a non-blocked request is
// NOT short-circuited by the takedown check (it proceeds to routing and fails
// with the normal not-found, proving the 403 is specific to blocked hosts).
func TestProxy_Blocklist_NotBlockedPassesThrough(t *testing.T) {
	bl := tunnel.NewBlocklist()
	bl.Block("someoneelse", "abuse")
	p := newBlockTestProxy(t, bl, nil)

	req := httptest.NewRequest(http.MethodGet, "http://a1b2c3d4.moltbunker.dev/", nil)
	rec := httptest.NewRecorder()
	p.handleRequest(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Fatalf("non-blocked subdomain must not be 403; got %d", rec.Code)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("non-blocked, unresolvable subdomain: status = %d, want 404", rec.Code)
	}
}

// TestProxy_NilBlocklistIsNoop confirms the seam is optional: with no blocklist
// wired, requests are never rejected by the takedown check.
func TestProxy_NilBlocklistIsNoop(t *testing.T) {
	p := newBlockTestProxy(t, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "http://a1b2c3d4.moltbunker.dev/", nil)
	rec := httptest.NewRecorder()
	p.handleRequest(rec, req)
	if rec.Code == http.StatusForbidden {
		t.Fatalf("nil blocklist must not 403; got %d", rec.Code)
	}
}

// TestStripInternalHeaders confirms every client-supplied X-Moltbunker-* header
// is removed at ingress so it can neither be forged nor leaked to a backend.
func TestStripInternalHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://a1b2c3d4.moltbunker.dev/", nil)
	req.Header.Set("X-Moltbunker-CustomDomain", "true")
	req.Header.Set("X-Moltbunker-Provider", "evil")
	req.Header.Set("x-moltbunker-deployment", "spoofed") // lowercase variant
	req.Header.Set("X-Real-Header", "keep")

	stripInternalHeaders(req)

	for k := range req.Header {
		if strings.HasPrefix(http.CanonicalHeaderKey(k), "X-Moltbunker-") {
			t.Fatalf("internal header %q was not stripped", k)
		}
	}
	if req.Header.Get("X-Real-Header") != "keep" {
		t.Fatal("non-internal header was incorrectly stripped")
	}
}

// fakeReverseOpener returns one end of a pipe; the test writes a canned HTTP
// response on the other end so the forward/reverse proxy code can read it.
type fakeReverseOpener struct {
	t        *testing.T
	gotReqCh chan *http.Request
}

func (f fakeReverseOpener) OpenStream(string) (net.Conn, error) {
	client, server := net.Pipe()
	go func() {
		// Read the proxied request off the stream so we can assert on it.
		br := bufio.NewReader(server)
		req, err := http.ReadRequest(br)
		if err != nil {
			_ = server.Close()
			return
		}
		f.gotReqCh <- req
		// Write a minimal HTTP/1.1 response back.
		_, _ = server.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nhi"))
		_ = server.Close()
	}()
	return client, nil
}

// TestProxy_ForgedCustomDomainHeader_NotReflected proves that a client sending
// X-Moltbunker-CustomDomain: true on a normal *.moltbunker.dev request (a) does
// NOT get the "verified custom domain" marker stamped on the response, and (b)
// does not leak the forged control header to the backend. EDGE-02.
func TestProxy_ForgedCustomDomainHeader_NotReflected(t *testing.T) {
	gotReqCh := make(chan *http.Request, 1)
	p := newBlockTestProxy(t, nil, nil)
	p.SetReverseStreamOpener(fakeReverseOpener{t: t, gotReqCh: gotReqCh})

	req := httptest.NewRequest(http.MethodGet, "http://a1b2c3d4.moltbunker.dev/", nil)
	req.Header.Set("X-Moltbunker-CustomDomain", "true") // forged by client
	rec := httptest.NewRecorder()
	p.handleRequest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if v := rec.Header().Get("X-Moltbunker-CustomDomain"); v != "" {
		t.Fatalf("forged custom-domain marker reflected to response: %q", v)
	}

	select {
	case backendReq := <-gotReqCh:
		if v := backendReq.Header.Get("X-Moltbunker-CustomDomain"); v != "" {
			t.Fatalf("forged control header leaked to backend: %q", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("backend did not receive proxied request")
	}
}

// TestProxy_VerifiedCustomDomain_MarkerSet proves the marker IS set on the
// response for a genuine verified BYO host (trusted resolveCustomDomain branch),
// confirming the signal is driven by trusted state, not a request header.
func TestProxy_VerifiedCustomDomain_MarkerSet(t *testing.T) {
	cd := NewDomainOwnershipStore(time.Hour)
	cd.Store("app.customer.com", "dep-abc12345", "", MethodCNAME)

	gotReqCh := make(chan *http.Request, 1)
	p := newBlockTestProxy(t, nil, cd)
	p.SetReverseStreamOpener(fakeReverseOpener{t: t, gotReqCh: gotReqCh})

	req := httptest.NewRequest(http.MethodGet, "http://app.customer.com/", nil)
	req.Host = "app.customer.com"
	rec := httptest.NewRecorder()
	p.handleRequest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if v := rec.Header().Get("X-Moltbunker-CustomDomain"); v != "true" {
		t.Fatalf("verified custom-domain marker not set on response: %q", v)
	}
	<-gotReqCh // drain
}
