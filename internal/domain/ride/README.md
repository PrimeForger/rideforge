# Ride Domain (`internal/domain/ride`)

## Purpose
Encapsulates the ride lifecycle state machine, offer tracking, pricing models, and ride entity invariants. Pickup coordinates are part of persisted ride state so demand recovery can derive the correct cell.

## Public Interface
`Ride` entity, `Status` enum (`RideRequested`, `RideMatching`, `RideOffered`, `RideAccepted`, `RideDriverArrived`, `RideInProgress`, `RideCompleted`, `RideCancelled`, `RideExpired`), `OfferStatus`, `OfferDeliveryStatus`.

## Dependencies
`github.com/google/uuid`, `time`. Pure domain.

## Invariants & Gotchas
Terminal states (`RideCompleted`, `RideCancelled`, `RideExpired`) are final and cannot transition back to active states. A ride can only be accepted if currently in `RideOffered` state.

## Human-Approval Required
Yes. Governs the core business entity, payment triggers, and lifecycle state transitions.

## Related ADRs
ADR-0004 (Domain-Driven Design), ADR-0015 (Smart Dispatch & Stand Dispatch), ADR-0019 (Data Consistency).

## Known Issues / Tech Debt
Multi-stop trips and route waypoint updates during transit are not yet implemented.
