# Application Orchestration Services (`internal/application`)

## Purpose
Coordinates domain use cases, manages transactional workflows, orchestrates dispatch, and delegates persistence and messaging to ports. Stage 7 Phase C adds locality coverage orchestration, per-locality surge calculation, and bounded policy snapshot refresh with base-fare fallback.

## Public Interface
`RideService`, `MatchingService`, `MatchingEngine`, `EventRouter`, `IdempotencyService`, DTOs, and shared application errors. Feature-specific services are owned by `demand`, `driver`, `location`, `locality`, `occupancy`, and `surge` subpackages.

## Dependencies
`internal/domain/*`, `internal/ports`, `internal/application/events`, `internal/application/matching`, `internal/infrastructure`, `go.uber.org/zap`, `go.opentelemetry.io/otel`.

## Invariants & Gotchas
Services execute use cases but never write raw SQL queries or Redis commands directly. Transactions must be managed through repository/transaction ports to maintain consistency. Demand projection failures are returned and remain unprocessed so redelivery or reconciliation can repair the projection. Surge resolution requires one active locality and a fresh valid policy; all unavailable or invalid inputs produce a 1.00x multiplier and a bounded-cardinality metric. Policy refresh atomically publishes a complete valid snapshot; any failed refresh retains the last known-good snapshot without advancing its success time, and the snapshot is unavailable after maximum staleness.

## Human-Approval Required
Yes. Coordinates core transactional operations, driver locking, offer processing, and demand reconciliation.

## Related ADRs
ADR-0002 (Architecture Style), ADR-0015 (Smart Dispatch), ADR-0019 (Data Consistency), ADR-0020 (Idempotency).

## Known Issues / Tech Debt
`EventRouter` currently acts as a central hub; will split into dedicated consumer handlers per domain topic. No fare quote boundary currently consumes `SurgeService`; the multiplier remains available for the existing pricing integration point once that boundary exists.
