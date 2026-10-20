# Occupancy & Demand Domain (`internal/domain/occupancy`)

## Purpose
Models ride demand aggregates and driver occupancy state across spatial H3 cells.

## Public Interface
`CellDemand` (`CellID`, `ActiveCount`, `UpdatedAt`), `RideDemandPointer` (`RideID`, `CellID`, `CreatedAt`), `CellOccupancy` (`CellID`, `DriverCount`, `DemandCount`).

## Dependencies
`github.com/google/uuid`, `time`. Pure domain.

## Invariants & Gotchas
Phase B's configurable minimum demand for HOT classification defaults to 2; Phase C's independently configured surge-policy minimum demand defaults to 3 and can keep surge at 1.00x even when a cell is HOT.
Ride demand pointers tie active rides to specific cells with strict TTLs to ensure that phantom demand counts expire automatically if ride cancellation events are missed.

## Human-Approval Required
No.

## Related ADRs
ADR-0009 (Redis for Real-Time State), ADR-0015 (Smart Dispatch & Stand Dispatch).

## Known Issues / Tech Debt
Dynamic passenger movement before driver acceptance is not currently reflected in cell demand.
