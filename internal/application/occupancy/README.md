# Occupancy Application Services

## Purpose

Calculates cell occupancy from demand and available driver supply and exposes the occupancy provider used by surge and dispatch.

## Public Interface

`OccupancyService`.

## Dependencies

`internal/domain/occupancy`, `internal/infrastructure/geo`, `internal/infrastructure/observability`, `internal/ports`, OpenTelemetry, and Zap.

## Invariants & Gotchas

Occupancy remains derived real-time state. Invalid or unavailable inputs are surfaced to callers so higher-level policies can apply their documented fallback.

## Human-Approval Required

YES — occupancy classification and thresholds require review.

## Related ADRs

ADR-0009, ADR-0015, ADR-0019, ADR-0022.

## Known Issues / Tech Debt

Occupancy metrics remain dependent on the active-cell refresh path.
