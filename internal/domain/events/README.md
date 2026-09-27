# Domain Events (`internal/domain/events`)

## Purpose
Defines the schema and payload contracts for all domain events published across the system.

## Public Interface
`DriverStatusChangedEvent`, `DriverLocationUpdatedEvent`, `DriverDeviceHeartbeatEvent`, `MatchingAttemptStartedEvent`, `DriverOfferSentEvent`, event type constants.

## Dependencies
`github.com/google/uuid`, `time`.

## Invariants & Gotchas
Event payloads must remain backward-compatible to prevent consumer deserialization failures during zero-downtime rolling upgrades.

## Human-Approval Required
No (unless changing existing field schemas in a breaking manner).

## Related ADRs
ADR-0005 (Event-Driven Architecture), ADR-0006 (Kafka / Redpanda for Event Streaming).

## Known Issues / Tech Debt
JSON serialization is used; schema evolution tools (Protobuf/Avro) will be introduced for high-throughput topics.
