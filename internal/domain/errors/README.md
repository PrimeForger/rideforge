# Domain Errors (`internal/domain/errors`)

## Purpose
Defines standard sentinel errors representing business rule and domain invariant violations.

## Public Interface
`ErrRideNotFound`, `ErrRideAlreadyAccepted`, `ErrRideCancelled`, `ErrDriverNotFound`, `ErrDriverUnavailable`, `ErrDriverAlreadyReserved`, `ErrDriverOffline`, `ErrInvalidCoordinates`, etc.

## Dependencies
`errors`. Standard library only.

## Invariants & Gotchas
All sentinel errors must be wrapped with `%w` when propagated through application and infrastructure layers to enable `errors.Is` comparison.

## Human-Approval Required
No.

## Related ADRs
ADR-0004 (Domain-Driven Design).

## Known Issues / Tech Debt
Needs centralized mapping helpers to map domain errors to HTTP/gRPC status codes.
