package molt

import (
	"context"
	"sync"
	"testing"
)

func TestResultStore_StoreAndRead(t *testing.T) {
	rs := NewResultStore()

	data := []byte("hello world")
	handle := rs.Store(data)
	if handle <= 0 {
		t.Fatalf("expected positive handle, got %d", handle)
	}

	// Size should report correct length
	size, ok := rs.Size(handle)
	if !ok {
		t.Fatal("Size returned false for valid handle")
	}
	if size != len(data) {
		t.Fatalf("Size = %d, want %d", size, len(data))
	}

	// Read should return data and free the handle (one-shot)
	got, ok := rs.Read(handle)
	if !ok {
		t.Fatal("Read returned false for valid handle")
	}
	if string(got) != "hello world" {
		t.Fatalf("Read = %q, want %q", string(got), "hello world")
	}

	// Second read should fail (one-shot)
	_, ok = rs.Read(handle)
	if ok {
		t.Fatal("expected second Read to return false (one-shot)")
	}
}

func TestResultStore_StoreError(t *testing.T) {
	rs := NewResultStore()

	handle := rs.StoreError("something went wrong")
	if handle >= 0 {
		t.Fatalf("expected negative handle for error, got %d", handle)
	}

	// ErrorMessage should return the error string
	msg, ok := rs.ErrorMessage(handle)
	if !ok {
		t.Fatal("ErrorMessage returned false for valid error handle")
	}
	if msg != "something went wrong" {
		t.Fatalf("ErrorMessage = %q, want %q", msg, "something went wrong")
	}

	// Second read should fail (one-shot)
	_, ok = rs.ErrorMessage(handle)
	if ok {
		t.Fatal("expected second ErrorMessage to return false")
	}
}

func TestResultStore_InvalidHandle(t *testing.T) {
	rs := NewResultStore()

	_, ok := rs.Size(999)
	if ok {
		t.Fatal("Size should return false for invalid handle")
	}

	_, ok = rs.Read(999)
	if ok {
		t.Fatal("Read should return false for invalid handle")
	}

	_, ok = rs.ErrorMessage(-999)
	if ok {
		t.Fatal("ErrorMessage should return false for invalid handle")
	}
}

func TestResultStore_Close(t *testing.T) {
	rs := NewResultStore()

	rs.Store([]byte("a"))
	rs.Store([]byte("b"))
	rs.StoreError("c")

	if rs.Len() != 3 {
		t.Fatalf("Len = %d, want 3", rs.Len())
	}

	rs.Close()
	if rs.Len() != 0 {
		t.Fatalf("Len = %d after Close, want 0", rs.Len())
	}
}

func TestResultStore_Concurrent(t *testing.T) {
	rs := NewResultStore()
	const n = 100

	var wg sync.WaitGroup
	handles := make([]int32, n)

	// Concurrent writes
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			handles[idx] = rs.Store([]byte("data"))
		}(i)
	}
	wg.Wait()

	if rs.Len() != n {
		t.Fatalf("Len = %d, want %d", rs.Len(), n)
	}

	// Concurrent reads
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, ok := rs.Read(handles[idx])
			if !ok {
				t.Errorf("Read(%d) returned false", handles[idx])
			}
		}(i)
	}
	wg.Wait()

	if rs.Len() != 0 {
		t.Fatalf("Len = %d after reading all, want 0", rs.Len())
	}
}

func TestResultStore_MonotonicHandles(t *testing.T) {
	rs := NewResultStore()

	h1 := rs.Store([]byte("a"))
	h2 := rs.Store([]byte("b"))
	h3 := rs.StoreError("c")

	if h1 >= h2 {
		t.Fatalf("handles should be monotonically increasing: h1=%d, h2=%d", h1, h2)
	}
	// Error handle counter still increments (just negative)
	if -h3 <= h2 {
		t.Fatalf("error handle counter should keep increasing: h2=%d, h3=%d", h2, h3)
	}
}

func TestHostServices_ContextRoundtrip(t *testing.T) {
	svc := NewHostServices(HostCapabilities{HTTPEnabled: true})
	svc.Owner = "0xTestWallet"

	ctx := withHostServices(context.Background(), svc)

	got := servicesFromContext(ctx)
	if got == nil {
		t.Fatal("servicesFromContext returned nil")
	}
	if got.Owner != "0xTestWallet" {
		t.Fatalf("Owner = %q, want %q", got.Owner, "0xTestWallet")
	}
	if !got.Config.HTTPEnabled {
		t.Fatal("HTTPEnabled should be true")
	}
}

func TestHostServices_ContextNil(t *testing.T) {
	got := servicesFromContext(context.Background())
	if got != nil {
		t.Fatal("expected nil from context without services")
	}
}
