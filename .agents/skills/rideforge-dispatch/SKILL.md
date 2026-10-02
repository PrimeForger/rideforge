---
name: rideforge-dispatch
description: Work safely on RideForge dispatch, candidate discovery, ranking, occupancy, demand, H3, and driver-location behavior using the documented architecture and ADRs.
metadata:
  origin: RideForge/ECC-adapted
---

# RideForge Dispatch Engineering

Use for smart dispatch, stand dispatch, candidate search, H3/geo behavior,
driver eligibility, demand/occupancy, ranking, ETA, and real-time state.

## Required context

Read `docs/STATUS.md`, the context-loading contract, the relevant business and
component documents, and the dispatch-related ADRs before changing behavior.
At minimum, inspect the relevant rules for matching, dispatch, location,
failure/degradation, idempotency, observability, and configuration.

## Design constraints

- Keep domain policies pure and put Redis/PostgreSQL/Kafka/provider behavior in
  infrastructure adapters behind ports.
- Preserve deterministic candidate eligibility and ranking inputs before adding
  adaptive or AI-assisted behavior.
- Make timeouts, search budgets, freshness windows, retries, and fallback modes
  explicit and observable.
- Preserve idempotency and event ordering for ride/driver state transitions.
- Treat location and demand data as sensitive; minimize retention and logging.
- Benchmark or measure hot paths before claiming a latency improvement.

## Verification checklist

- Table-driven unit tests cover normal, empty, stale, invalid, timeout, and
  degraded-provider cases.
- Race tests cover concurrent updates and reconciliation.
- Property/fuzz tests are considered for geo boundaries and input normalization.
- Metrics cover latency, freshness, candidate counts, queue/lag, retries, and
  fallback activation.
- Relevant module README and ADR documentation are updated when contracts or
  invariants change.

