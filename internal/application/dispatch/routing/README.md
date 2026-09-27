# Dispatch Routing & ETA Provider (`internal/application/dispatch/routing`)

## Purpose
Computes route distance and estimated travel times between candidate drivers and ride pickup locations.

## Public Interface
`RouteCalculator`, `RouteResult` (`DistanceMeters`, `DurationSeconds`), `OSRMProvider`, `HaversineFallbackProvider`.

## Dependencies
`context`, `internal/domain/location`, `internal/shared/geo`.

## Invariants & Gotchas
If the primary routing service fails or times out, the calculator automatically falls back to Haversine distance with speed heuristics to prevent dispatch halts.

## Human-Approval Required
No.

## Related ADRs
ADR-0017 (ETA & Route Provider Strategy).

## Known Issues / Tech Debt
Real-time traffic congestion data is currently estimated via static time-of-day multipliers.
