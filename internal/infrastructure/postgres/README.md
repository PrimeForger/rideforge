# PostgreSQL Repositories & Transaction Manager (`internal/infrastructure/postgres`)

## Purpose
Provides PostgreSQL database access, connection management, repository implementations, and transaction orchestration.

## Public Interface
`DB`, `TxManager`, `RideRepository`, `DriverRepository`, `OutboxRepository`, `ProcessedEventRepository`, `DriverMetricsRepository`, `DriverPushTokenRepository`, `DriverLocker`.

## Dependencies
`database/sql`, `github.com/lib/pq`, `internal/ports`, `internal/domain/*`.

## Invariants & Gotchas
Queries must pass context. Connection pool parameters must remain aligned with PgBouncer transaction-pooling limits.

## Human-Approval Required
Yes. Governs relational persistence, schema integrity, and database transaction boundaries.

## Related ADRs
ADR-0007 (PostgreSQL as Primary Database), ADR-0011 (PgBouncer), ADR-0019 (Data Consistency).

## Known Issues / Tech Debt
Schema migrations are run via external SQL scripts rather than an automated migration tool.
