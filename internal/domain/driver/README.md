# Driver Domain (`internal/domain/driver`)

## Purpose
Encapsulates the driver entity, lifecycle state machine, status transitions, offline reasons, and driver quality metrics.

## Public Interface
`Driver` entity, `Status` enum (`DriverOffline`, `DriverAvailable`, `DriverReserved`, `DriverOffered`, `DriverDispatched`, `DriverOnRide`, `DriverPaused`), `OfflineReason` enum (`Manual`, `HeartbeatTimeout`, `BatteryLow`, `LocationDisabled`), `DriverMetrics`.

## Dependencies
`time`, `github.com/google/uuid`. Pure domain logic.

## Invariants & Gotchas
Driver state transitions must be strictly enforced. A driver can only be reserved if currently in `DriverAvailable` state. A paused or offline driver cannot receive offers.

## Human-Approval Required
Yes. Driver lifecycle states directly affect dispatch eligibility, lock safety, and driver operations.

## Related ADRs
ADR-0004 (Domain-Driven Design), ADR-0010 (Driver Location Storage Strategy), ADR-0015 (Smart Dispatch & Stand Dispatch).

## Known Issues / Tech Debt
Driver ratings and acceptance rate calculations are static structs; streaming metric computation will be added in Phase 2.
