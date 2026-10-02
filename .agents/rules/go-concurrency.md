---
paths:
  - "**/*.go"
---
# RideForge Go Concurrency Rules

- Every goroutine must have a clear owner, exit condition, and cancellation path.
- Prefer `context.Context`, channel closure, and `sync.WaitGroup` over sleeps or
  polling for lifecycle coordination.
- Bound worker counts, channel capacity, retries, fan-out, and background work.
- Never copy a mutex or use a map concurrently without synchronization.
- Document ownership when a goroutine reads or mutates shared state.
- Use atomic operations only for simple independent values; use a mutex or an
  actor/owner goroutine for compound invariants.
- Make shutdown idempotent and test cancellation, queue saturation, duplicate
  delivery, timeout, and partial failure paths.
- Run `go test -race` for changed packages and the full repository before
  declaring concurrency-sensitive work complete.
- Avoid allocations and blocking network calls in dispatch/location hot paths;
  measure before optimizing and preserve correctness under load.

