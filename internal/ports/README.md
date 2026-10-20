# Domain & Application Ports (`internal/ports`)

## Purpose
Defines the abstract interface contracts (inbound and outbound ports) that decouple business logic from databases, message queues, caches, and external networks.

## Public Interface
Interfaces include `RideRepository`, `DriverRepository`, `DriverLocker`, `H3DriverIndex`, `RealtimeDriverCache`, `CellDemandStore`, `CellOccupancyStore`, `LocalityRepository`, `LocalityResolver`, `LocalityBoundaryIndexer`, `H3CellResolver`, `SurgePolicyRepository`, `SurgePolicyProvider`, `SurgeObserver`, and event ports. Demand-store cell listing supports recovery of Redis-only stale state.

## Dependencies
`context`, standard domain packages (including locality and surge policy), and `github.com/google/uuid`. No infrastructure imports.

## Invariants & Gotchas
All I/O operations accept `context.Context` first. Interfaces model domain capabilities, not storage technologies. Locality resolution reports unresolved and ambiguous cells explicitly so callers never select an arbitrary match.

## Human-Approval Required
Yes. Modifying port contracts impacts all infrastructure adapters and use-case orchestrators across the platform.

## Related ADRs
ADR-0002 (Architecture Style), ADR-0004 (Domain-Driven Design), ADR-0019 (Data Consistency and Transaction Boundaries).

## Known Issues / Tech Debt
Bulk batch operations on repositories currently lack streaming iterator contracts.
