;; Minimal WASM module that exports _start and does nothing.
;; Used for compile tests and basic invocation tests (empty stdout).
(module
  (func $start)
  (export "_start" (func $start))
)
