# Transactional Outbox Domain (`internal/domain/outbox`)

## Purpose
Defines the outbox event entity ensuring at-least-once asynchronous event publishing across database transaction boundaries.

## Public Interface
`OutboxEvent` struct (`ID`, `AggregateType`, `AggregateID`, `EventType`, `Payload`, `Headers`, `Status`, `RetryCount`, `CreatedAt`), `OutboxStatus` enum (`Pending`, `Published`, `Failed`).

## Dependencies
`github.com/google/uuid`, `time`. Pure domain.

## Invariants & Gotchas
Outbox records must ALWAYS be inserted inside the same database transaction as the state change they announce (e.g. creating a ride or accepting an offer).

## Human-Approval Required
Yes. Modifying outbox behavior directly impacts event publishing consistency and data integrity.

## Related ADRs
ADR-0012 (Outbox Pattern), ADR-0019 (Data Consistency & Transaction Boundaries).

## Known Issues / Tech Debt
Payloads are raw JSON bytes; schema versioning headers need standardization.
