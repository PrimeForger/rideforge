# Driver Application Services

## Purpose

Coordinates driver lifecycle, device, response, metrics, and driver-facing ride workflows.

## Public Interface

`DriverService`, `DriverResponseService`, `DriverResponseCommandService`, `DriverDeviceService`, and `DriverMetricsService`.

## Dependencies

`internal/application/events`, `internal/application/location`, `internal/domain`, `internal/infrastructure`, `internal/ports`, Zap, and UUID.

## Invariants & Gotchas

State changes use repository transactions and outbox events. Driver location orchestration is provided by the location application package.

## Human-Approval Required

YES — driver state, response, and offer processing changes require review.

## Related ADRs

ADR-0012, ADR-0015, ADR-0019, ADR-0020, ADR-0021.

## Known Issues / Tech Debt

Driver and dispatch orchestration remain in the same deployable application by design.
