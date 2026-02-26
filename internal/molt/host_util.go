package molt

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/tetratelabs/wazero/api"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// hostResultSize returns the byte length of a result handle.
// Params: [handle i32] → [size i32]
// Returns 0 if the handle is invalid.
func hostResultSize(ctx context.Context, mod api.Module, stack []uint64) {
	handle := api.DecodeI32(stack[0])

	svc := servicesFromContext(ctx)
	if svc == nil {
		stack[0] = 0
		return
	}

	size, ok := svc.results.Size(handle)
	if !ok {
		stack[0] = 0
		return
	}
	stack[0] = api.EncodeI32(int32(size))
}

// hostResultRead copies result data into WASM memory and frees the handle.
// Params: [handle i32, dst_ptr i32, dst_len i32] → [bytes_written i32]
// Returns -1 if the handle is invalid. Frees the handle after reading.
func hostResultRead(ctx context.Context, mod api.Module, stack []uint64) {
	handle := api.DecodeI32(stack[0])
	dstPtr := api.DecodeU32(stack[1])
	dstLen := api.DecodeU32(stack[2])

	svc := servicesFromContext(ctx)
	if svc == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	data, ok := svc.results.Read(handle)
	if !ok {
		stack[0] = api.EncodeI32(-1)
		return
	}

	// Clamp to destination buffer size
	writeLen := uint32(len(data))
	if writeLen > dstLen {
		writeLen = dstLen
	}

	mem := mod.Memory()
	if mem == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	if writeLen > 0 {
		if !mem.Write(dstPtr, data[:writeLen]) {
			logging.Warn("host.result_read: memory write out of bounds", "ptr", dstPtr, "len", writeLen)
			stack[0] = api.EncodeI32(-1)
			return
		}
	}

	stack[0] = api.EncodeI32(int32(writeLen))
}

// hostErrorMessage copies an error string into WASM memory for a negative handle.
// Params: [handle i32, dst_ptr i32, dst_len i32] → [bytes_written i32]
// Returns -1 if the handle is invalid. The handle should be negative (from StoreError).
func hostErrorMessage(ctx context.Context, mod api.Module, stack []uint64) {
	handle := api.DecodeI32(stack[0])
	dstPtr := api.DecodeU32(stack[1])
	dstLen := api.DecodeU32(stack[2])

	svc := servicesFromContext(ctx)
	if svc == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	msg, ok := svc.results.ErrorMessage(handle)
	if !ok {
		stack[0] = api.EncodeI32(-1)
		return
	}

	msgBytes := []byte(msg)
	writeLen := uint32(len(msgBytes))
	if writeLen > dstLen {
		writeLen = dstLen
	}

	mem := mod.Memory()
	if mem == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	if writeLen > 0 {
		if !mem.Write(dstPtr, msgBytes[:writeLen]) {
			logging.Warn("host.error_message: memory write out of bounds", "ptr", dstPtr, "len", writeLen)
			stack[0] = api.EncodeI32(-1)
			return
		}
	}

	stack[0] = api.EncodeI32(int32(writeLen))
}

// hostRandomBytes fills WASM memory with cryptographically secure random bytes.
// Params: [dst_ptr i32, dst_len i32] → [bytes_written i32]
// Returns -1 on error.
func hostRandomBytes(_ context.Context, mod api.Module, stack []uint64) {
	dstPtr := api.DecodeU32(stack[0])
	dstLen := api.DecodeU32(stack[1])

	if dstLen == 0 {
		stack[0] = 0
		return
	}

	mem := mod.Memory()
	if mem == nil {
		stack[0] = api.EncodeI32(-1)
		return
	}

	buf := make([]byte, dstLen)
	if _, err := rand.Read(buf); err != nil {
		logging.Error("host.random_bytes: crypto/rand failed", "err", err)
		stack[0] = api.EncodeI32(-1)
		return
	}

	if !mem.Write(dstPtr, buf) {
		logging.Warn("host.random_bytes: memory write out of bounds", "ptr", dstPtr, "len", dstLen)
		stack[0] = api.EncodeI32(-1)
		return
	}

	stack[0] = api.EncodeI32(int32(dstLen))
}

// hostTimeNowMs returns the current UTC time in milliseconds since epoch.
// Params: [] → [ms i64]
func hostTimeNowMs(_ context.Context, _ api.Module, stack []uint64) {
	stack[0] = api.EncodeI64(time.Now().UnixMilli())
}
