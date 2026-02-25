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
//   - host.log(level i32, ptr i32, len i32)          — fully implemented
//   - host.get_request(ptr i32, len i32) -> i32       — stub (v1 uses stdin)
//   - host.set_response(ptr i32, len i32) -> i32      — stub (v1 uses stdout)
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
