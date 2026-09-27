# Outbox Relay Worker (`internal/infrastructure/outbox`)

## Purpose
Background worker polling pending outbox events from PostgreSQL and publishing them reliably to Kafka.

## Public Interface
`OutboxWorker`, `NewOutboxWorker(...)`, `Start(ctx)`, `Stop()`.

## Dependencies
`context`, `internal/ports`, `go.uber.org/zap`.

## Invariants & Gotchas
Uses `FOR UPDATE SKIP LOCKED` during polling to allow multiple concurrent outbox workers without duplicate processing.

## Human-Approval Required
Yes. Relays transactional events to external message streams.

## Related ADRs
ADR-0012 (Outbox Pattern), ADR-0019 (Data Consistency & Transaction Boundaries).

## Known Issues / Tech Debt
Database polling will be upgraded to change-data-capture (Debezium) at high scale.
