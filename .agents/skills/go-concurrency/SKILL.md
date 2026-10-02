---
name: go-concurrency
description: Design and review bounded, cancellation-safe Go concurrency for RideForge dispatch, location, event, and reconciliation workloads.
metadata:
  origin: ECC-adapted
---

# Go Concurrency for RideForge

Use for workers, queues, background reconciliation, event consumers, location
updates, dispatch fan-out, timers, locks, and shutdown behavior.

## Review sequence

1. Identify the owner of every mutable value.
2. Draw the goroutine and channel lifecycle, including cancellation and shutdown.
3. Identify bounds for goroutines, queues, retries, timers, and retained data.
4. Check ordering, duplicate delivery, timeout, and partial failure semantics.
5. Add deterministic tests for cancellation, saturation, races, and shutdown.
6. Run targeted tests with `-race`, then the full race suite.

## Required properties

- No goroutine leak when the parent context is canceled.
- No unbounded channel, retry, fan-out, or cache growth.
- No unsynchronized map access or copied synchronization primitive.
- Shutdown is idempotent and does not strand work silently.
- Metrics/logging expose queue depth, retries, lag, and dropped/degraded work
  without exposing sensitive payloads.

