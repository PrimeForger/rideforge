# Surge Application Services

## Purpose

Resolves locality-specific surge decisions and manages bounded, per-instance policy snapshots.

## Public Interface

`SurgeService`, `SurgePolicyCache`, and `SurgePolicyRefreshWorker`.

## Dependencies

`internal/domain/occupancy`, `internal/domain/surgepolicy`, `internal/ports`, Zap, and UUID.

## Invariants & Gotchas

PostgreSQL is authoritative for policies. Successful refreshes atomically replace complete snapshots; failed refreshes retain the last known-good snapshot until maximum staleness, after which decisions fall back to 1.00x.

## Human-Approval Required

YES — surge policy, thresholds, and multiplier behavior require review.

## Related ADRs

ADR-0019, ADR-0021, ADR-0022, ADR-0024.

## Known Issues / Tech Debt

The fare or quote boundary that consumes surge decisions is intentionally deferred until a pricing operation exists.
