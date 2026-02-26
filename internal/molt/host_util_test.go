package molt

import (
	"testing"
	"time"
)

func TestHostResultSize_WithServices(t *testing.T) {
	rs := NewResultStore()
	data := []byte("test data 12345")
	handle := rs.Store(data)

	size, ok := rs.Size(handle)
	if !ok {
		t.Fatal("Size should return true for valid handle")
	}
	if size != 15 {
		t.Fatalf("Size = %d, want 15", size)
	}
}

func TestHostResultRead_ClampsToBuffer(t *testing.T) {
	rs := NewResultStore()
	data := []byte("hello world!") // 12 bytes
	handle := rs.Store(data)

	// Size reports full length
	size, ok := rs.Size(handle)
	if !ok || size != 12 {
		t.Fatalf("Size = %d, ok=%v, want 12", size, ok)
	}

	// Read returns full data
	got, ok := rs.Read(handle)
	if !ok {
		t.Fatal("Read returned false")
	}
	if string(got) != "hello world!" {
		t.Fatalf("Read = %q", string(got))
	}
}

func TestHostErrorMessage_NegativeHandle(t *testing.T) {
	rs := NewResultStore()
	errHandle := rs.StoreError("timeout exceeded")

	if errHandle >= 0 {
		t.Fatalf("StoreError should return negative handle, got %d", errHandle)
	}

	msg, ok := rs.ErrorMessage(errHandle)
	if !ok {
		t.Fatal("ErrorMessage returned false for valid error handle")
	}
	if msg != "timeout exceeded" {
		t.Fatalf("ErrorMessage = %q, want %q", msg, "timeout exceeded")
	}
}

func TestHostTimeNowMs_ReturnsRecentTime(t *testing.T) {
	before := time.Now().UnixMilli()
	// Simulate what hostTimeNowMs does
	now := time.Now().UnixMilli()
	after := time.Now().UnixMilli()

	if now < before || now > after {
		t.Fatalf("time_now_ms returned %d, not in range [%d, %d]", now, before, after)
	}
}

func TestResultStore_EmptyData(t *testing.T) {
	rs := NewResultStore()
	handle := rs.Store([]byte{})

	size, ok := rs.Size(handle)
	if !ok {
		t.Fatal("Size returned false for empty data handle")
	}
	if size != 0 {
		t.Fatalf("Size = %d, want 0", size)
	}

	got, ok := rs.Read(handle)
	if !ok {
		t.Fatal("Read returned false for empty data handle")
	}
	if len(got) != 0 {
		t.Fatalf("Read returned %d bytes, want 0", len(got))
	}
}

func TestResultStore_LargeData(t *testing.T) {
	rs := NewResultStore()
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}
	handle := rs.Store(data)

	size, ok := rs.Size(handle)
	if !ok {
		t.Fatal("Size returned false")
	}
	if size != len(data) {
		t.Fatalf("Size = %d, want %d", size, len(data))
	}

	got, ok := rs.Read(handle)
	if !ok {
		t.Fatal("Read returned false")
	}
	if len(got) != len(data) {
		t.Fatalf("Read returned %d bytes, want %d", len(got), len(data))
	}
	// Spot-check
	if got[0] != 0 || got[255] != 255 || got[256] != 0 {
		t.Fatal("data corruption detected")
	}
}
