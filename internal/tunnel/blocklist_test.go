package tunnel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBlocklist_BlockUnblock(t *testing.T) {
	bl := NewBlocklist()
	if blocked, _ := bl.IsBlocked("evil"); blocked {
		t.Fatal("nothing should be blocked initially")
	}
	bl.Block("Evil", "abuse") // case-insensitive
	blocked, reason := bl.IsBlocked("evil")
	if !blocked || reason != "abuse" {
		t.Fatalf("blocked=%v reason=%q", blocked, reason)
	}
	if !bl.Unblock("evil") {
		t.Fatal("Unblock should report a removed entry")
	}
	if blocked, _ := bl.IsBlocked("evil"); blocked {
		t.Fatal("should be unblocked")
	}
	if bl.Unblock("evil") {
		t.Fatal("Unblock of absent entry should report false")
	}
}

func TestBlocklist_List(t *testing.T) {
	bl := NewBlocklist()
	bl.Block("b", "")
	bl.Block("a", "reason-a")
	list := bl.List()
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	if list[0].Subdomain != "a" || list[1].Subdomain != "b" {
		t.Fatalf("expected sorted list, got %+v", list)
	}
}

func TestBlocklist_EmptySubdomainIgnored(t *testing.T) {
	bl := NewBlocklist()
	bl.Block("   ", "x")
	if bl.Len() != 0 {
		t.Fatal("empty subdomain must not be blocked")
	}
}

func TestBlocklistAdmin_PostGetDelete(t *testing.T) {
	bl := NewBlocklist()
	h := NewBlocklistAdminHandler(bl)

	// POST a block.
	post := httptest.NewRequest(http.MethodPost, "/v1/ingress/blocklist",
		strings.NewReader(`{"subdomain":"evil","reason":"takedown"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, post)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if blocked, _ := bl.IsBlocked("evil"); !blocked {
		t.Fatal("subdomain not blocked after POST")
	}

	// GET the list.
	get := httptest.NewRequest(http.MethodGet, "/v1/ingress/blocklist", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, get)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", rec.Code)
	}
	var list []BlockEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 || list[0].Subdomain != "evil" {
		t.Fatalf("unexpected list: %+v", list)
	}

	// DELETE (unblock).
	del := httptest.NewRequest(http.MethodDelete, "/v1/ingress/blocklist?subdomain=evil", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, del)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want 204", rec.Code)
	}
	if blocked, _ := bl.IsBlocked("evil"); blocked {
		t.Fatal("subdomain still blocked after DELETE")
	}
}

func TestBlocklistAdmin_Errors(t *testing.T) {
	h := NewBlocklistAdminHandler(NewBlocklist())

	// POST without subdomain.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST empty status = %d, want 400", rec.Code)
	}

	// DELETE without subdomain query.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/x", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("DELETE no-param status = %d, want 400", rec.Code)
	}

	// DELETE absent subdomain.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/x?subdomain=missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("DELETE missing status = %d, want 404", rec.Code)
	}

	// Unsupported method.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/x", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("PUT status = %d, want 405", rec.Code)
	}
}
