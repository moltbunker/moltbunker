;; Echo module: reads stdin into memory, writes it to stdout.
;; Implements the Molt stdin/stdout JSON protocol.
(module
  ;; WASI imports
  (import "wasi_snapshot_preview1" "fd_read"
    (func $fd_read (param i32 i32 i32 i32) (result i32)))
  (import "wasi_snapshot_preview1" "fd_write"
    (func $fd_write (param i32 i32 i32 i32) (result i32)))
  (import "wasi_snapshot_preview1" "proc_exit"
    (func $proc_exit (param i32)))

  (memory 1)
  (export "memory" (memory 0))

  ;; Memory layout:
  ;; 0-7:    iovec for fd_read  (buf_ptr=1024, buf_len=8192)
  ;; 8-11:   nread result
  ;; 12-19:  iovec for fd_write (buf_ptr=1024, buf_len=nread)
  ;; 20-23:  nwritten result
  ;; 1024+:  data buffer

  (func $start
    ;; Set up iovec for reading: buf_ptr=1024, buf_len=8192
    (i32.store (i32.const 0) (i32.const 1024))   ;; iov_base = 1024
    (i32.store (i32.const 4) (i32.const 8192))    ;; iov_len = 8192

    ;; fd_read(stdin=0, iovs=0, iovs_len=1, nread=8)
    (call $fd_read (i32.const 0) (i32.const 0) (i32.const 1) (i32.const 8))
    drop

    ;; Load nread
    ;; Set up iovec for writing: buf_ptr=1024, buf_len=nread
    (i32.store (i32.const 12) (i32.const 1024))           ;; iov_base = 1024
    (i32.store (i32.const 16) (i32.load (i32.const 8)))    ;; iov_len = nread

    ;; fd_write(stdout=1, iovs=12, iovs_len=1, nwritten=20)
    (call $fd_write (i32.const 1) (i32.const 12) (i32.const 1) (i32.const 20))
    drop

    ;; exit(0)
    (call $proc_exit (i32.const 0))
  )

  (export "_start" (func $start))
)
