# Domain & Application Ports (`internal/ports`)

## Purpose
Defines the abstract interface contracts (inbound and outbound ports) that decouple business logic from databases, message queues, caches, and external networks.

## Public Interface
Interfaces: `RideRepository`, `DriverRepository`, `DriverLocker`, `H3DriverIndex`, `CellDemandStore`, `CellOccupancyStore`, `EventBus`, `OutboxRepository`, `DriverOfferGateway`, `PushNotifier`, `DriverMetricsRepository`, `DriverOfflineCommander`, `DriverPushTokenRepository`.

## Dependencies
`context`, standard domain packages (`domain/driver`, `domain/ride`, `domain/occupancy`, `domain/outbox`). Strictly no infrastructure or third-party library imports.

## Invariants & Gotchas
All I/O operations must accept `context.Context` as their first parameter and return `error` as their last. Interfaces must model domain capabilities, not storage technologies.

## Human-Approval Required
Yes. Modifying port contracts impacts all infrastructure adapters and use-case orchestrators across the platform.

## Related ADRs
ADR-0002 (Architecture Style), ADR-0004 (Domain-Driven Design), ADR-0019 (Data Consistency and Transaction Boundaries).

## Known Issues / Tech Debt
Bulk batch operations on repositories currently lack streaming iterator contracts.
