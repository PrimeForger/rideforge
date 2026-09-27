# Application Orchestration Services (`internal/application`)

## Purpose
Coordinates domain use cases, manages transactional workflows, orchestrates dispatch, and delegates persistence and messaging to ports.

## Public Interface
`RideService`, `DriverService`, `DriverLocationService`, `DriverResponseService`, `DriverResponseCommandService`, `DriverDeviceService`, `DriverMetricsService`, `DemandService`, `DemandRecoveryService`, `OccupancyService`, `GeoService`, `IdempotencyService`, `EventRouter`, DTOs.

## Dependencies
`internal/domain/*`, `internal/ports`, `go.uber.org/zap`, `go.opentelemetry.io/otel`.

## Invariants & Gotchas
Services execute use cases but never write raw SQL queries or Redis commands directly. Transactions must be managed through `ports.TxManager` to maintain consistency.

## Human-Approval Required
Yes. Coordinates core transactional operations, driver locking, offer processing, and demand reconciliation.

## Related ADRs
ADR-0002 (Architecture Style), ADR-0015 (Smart Dispatch), ADR-0019 (Data Consistency), ADR-0020 (Idempotency).

## Known Issues / Tech Debt
`EventRouter` currently acts as a central hub; will split into dedicated consumer handlers per domain topic.
