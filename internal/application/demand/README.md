# Demand Application Services

## Purpose

Coordinates ride demand projection, recovery, reconciliation, and demand event processing.

## Public Interface

`DemandService`, `DemandRecoveryService`, `DemandReconciliationWorker`, and `ProcessProjection`.

## Dependencies

`internal/domain`, `internal/application/events`, `internal/infrastructure/geo`, `internal/infrastructure/observability`, `internal/ports`, OpenTelemetry, Zap, and UUID.

## Invariants & Gotchas

Projection failures are returned without recording the event as processed. Recovery derives active demand from PostgreSQL and repairs Redis state.

## Human-Approval Required

YES — demand projection, recovery, and reconciliation changes require review.

## Related ADRs

ADR-0019, ADR-0020, ADR-0021, ADR-0022.

## Known Issues / Tech Debt

The reconciliation worker remains timer-based and may have a bounded startup recovery delay.
