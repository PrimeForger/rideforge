# Ride Application Use Cases (`internal/application/ride`)

## Purpose
Implements specific ride command handlers, including ride request submission and legal pre-validation.

## Public Interface
`RequestRideUseCase`, `RequestRideCommand`, `RequestRideResult`.

## Dependencies
`internal/domain/ride`, `internal/domain/region`, `internal/ports`.

## Invariants & Gotchas
Checks coordinate boundaries and enforces interstate border rules via `region.IsRideAllowed` before committing a ride request.

## Human-Approval Required
Yes. Entrypoint for passenger ride bookings and legal border enforcement.

## Related ADRs
ADR-0014 (API & Service Communication), ADR-0018 (Regional & Legal Ride Validation), ADR-0019 (Data Consistency).

## Known Issues / Tech Debt
Fare quote estimation is mocked; integration with full pricing service is pending.
