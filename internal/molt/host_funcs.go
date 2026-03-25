package molt

import (
	"context"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"github.com/moltbunker/moltbunker/internal/logging"
)

const hostModuleName = "host"

// registerHostFunctions registers the "host" module with functions callable from WASM.
//
// Exported functions:
//
//	Logging:
//	  - host.log(level i32, ptr i32, len i32)                            — log message from WASM
//
//	Legacy stubs (v1 uses stdin/stdout):
//	  - host.get_request(ptr i32, len i32) -> i32                        — stub
//	  - host.set_response(ptr i32, len i32) -> i32                       — stub
//
//	Result handles (v2 — ptr/len data exchange):
//	  - host.result_size(handle i32) -> i32                              — byte length of handle
//	  - host.result_read(handle i32, dst_ptr i32, dst_len i32) -> i32   — copy data, free handle
//	  - host.error_message(handle i32, dst_ptr i32, dst_len i32) -> i32 — read error string
//
//	Utilities:
//	  - host.random_bytes(dst_ptr i32, dst_len i32) -> i32              — crypto/rand fill
//	  - host.time_now_ms() -> i64                                        — UTC millis
//
//	Services:
//	  - host.http_request(req_ptr i32, req_len i32) -> i32              — HTTP outbound
//	  - host.storage_put(req_ptr i32, req_len i32) -> i32               — store object
//	  - host.storage_get(req_ptr i32, req_len i32) -> i32               — retrieve object
//	  - host.storage_delete(req_ptr i32, req_len i32) -> i32            — delete object
//	  - host.storage_list(req_ptr i32, req_len i32) -> i32              — list objects
//	  - host.crawl_page(req_ptr i32, req_len i32) -> i32                — crawl single page
func registerHostFunctions(ctx context.Context, rt wazero.Runtime) error {
	_, err := rt.NewHostModuleBuilder(hostModuleName).
		// host.log — reads a string from WASM memory and logs it
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostLog), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, nil).
		WithParameterNames("level", "ptr", "len").
		Export("log").

		// host.get_request — stub, v1 uses stdin JSON instead
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostGetRequestStub), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("ptr", "len").
		Export("get_request").

		// host.set_response — stub, v1 uses stdout JSON instead
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostSetResponseStub), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("ptr", "len").
		Export("set_response").

		// --- Result handle functions ---

		// host.result_size — get byte length of a result handle
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostResultSize), []api.ValueType{api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("handle").
		Export("result_size").

		// host.result_read — copy result data into WASM memory, free handle
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostResultRead), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("handle", "dst_ptr", "dst_len").
		Export("result_read").

		// host.error_message — read error string for a negative handle
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostErrorMessage), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("handle", "dst_ptr", "dst_len").
		Export("error_message").

		// --- Utility functions ---

		// host.random_bytes — fill WASM memory with crypto/rand bytes
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostRandomBytes), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("dst_ptr", "dst_len").
		Export("random_bytes").

		// host.time_now_ms — current UTC time in milliseconds
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostTimeNowMs), nil, []api.ValueType{api.ValueTypeI64}).
		Export("time_now_ms").

		// --- Service functions ---

		// host.http_request — execute HTTP request via proxy
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostHTTPRequest), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("req_ptr", "req_len").
		Export("http_request").

		// host.storage_put — store object in bucket
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostStoragePut), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("req_ptr", "req_len").
		Export("storage_put").

		// host.storage_get — retrieve object from bucket
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostStorageGet), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("req_ptr", "req_len").
		Export("storage_get").

		// host.storage_delete — delete object from bucket
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostStorageDelete), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("req_ptr", "req_len").
		Export("storage_delete").

		// host.storage_list — list objects in bucket
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostStorageList), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("req_ptr", "req_len").
		Export("storage_list").

		// host.crawl_page — crawl a single web page
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(hostCrawlPage), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		WithParameterNames("req_ptr", "req_len").
		Export("crawl_page").
		Instantiate(ctx)

	return err
}

// hostLog reads a log message from WASM linear memory and routes it through slog.
// Parameters on stack: [level i32, ptr i32, len i32]
// Level mapping: 0=debug, 1=info, 2=warn, 3+=error
func hostLog(ctx context.Context, mod api.Module, stack []uint64) {
	level := api.DecodeI32(stack[0])
	ptr := api.DecodeU32(stack[1])
	length := api.DecodeU32(stack[2])

	mem := mod.Memory()
	if mem == nil {
		return
	}

	buf, ok := mem.Read(ptr, length)
	if !ok {
		logging.Warn("molt host.log: invalid memory read", "ptr", ptr, "len", length)
		return
	}

	msg := string(buf)
	name := mod.Name()

	switch {
	case level <= 0:
		logging.Debug(msg, "molt", name)
	case level == 1:
		logging.Info(msg, "molt", name)
	case level == 2:
		logging.Warn(msg, "molt", name)
	default:
		logging.Error(msg, "molt", name)
	}
}

// hostGetRequestStub is a no-op placeholder. V1 Molts read requests from stdin.
// Returns 0 (no bytes written).
func hostGetRequestStub(_ context.Context, _ api.Module, stack []uint64) {
	stack[0] = 0
}

// hostSetResponseStub is a no-op placeholder. V1 Molts write responses to stdout.
// Returns 0 (no bytes written).
func hostSetResponseStub(_ context.Context, _ api.Module, stack []uint64) {
	stack[0] = 0
}
