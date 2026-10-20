# Domain Events (`internal/domain/events`)

## Purpose
Defines the schema and payload contracts for all domain events published across the system.

## Public Interface
`DriverStatusChangedEvent`, `DriverLocationUpdatedEvent`, `DriverDeviceHeartbeatEvent`, `MatchingAttemptStartedEvent`, `DriverOfferSentEvent`, event type constants.

## Dependencies
`github.com/google/uuid`, `time`.

## Invariants & Gotchas
Ride request and acceptance event payloads use the current snake_case JSON contract (`ride_id`, `pickup_latitude`, `pickup_longitude`, and `driver_id`). In-repository producers and consumers change together; no legacy decoding is provided.

## Human-Approval Required
No (unless changing existing field schemas).

## Related ADRs
ADR-0005 (Event-Driven Architecture), ADR-0006 (Kafka / Redpanda for Event Streaming).

## Known Issues / Tech Debt
JSON serialization is used; schema evolution tools (Protobuf/Avro) will be introduced for high-throughput topics.
