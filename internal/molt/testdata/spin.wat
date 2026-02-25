;; Infinite loop module for timeout testing.
;; _start enters an infinite loop that never returns.
(module
  (func $start
    (loop $inf
      (br $inf)
    )
  )
  (export "_start" (func $start))
)
