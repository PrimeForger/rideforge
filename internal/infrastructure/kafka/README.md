# Kafka Event Streaming Adapter (`internal/infrastructure/kafka`)

## Purpose
Manages Kafka/Redpanda topic producers and consumer groups using `segmentio/kafka-go`.

## Public Interface
`Producer`, `NewProducer(...)`, `Publish(ctx, topic, key, payload)`, `Consumer`, `NewConsumer(...)`, `Start(handler)`.

## Dependencies
`github.com/segmentio/kafka-go`, `go.uber.org/zap`, `go.opentelemetry.io/otel`.

## Invariants & Gotchas
Messages must use entity ID (e.g. `ride_id` or `driver_id`) as the partition key to guarantee in-order message processing per entity.

## Human-Approval Required
No.

## Related ADRs
ADR-0005 (Event-Driven Architecture), ADR-0006 (Kafka / Redpanda), ADR-0013 (Dead Letter Queue Strategy).

## Known Issues / Tech Debt
Consumer group offset commits use default auto-commit; manual synchronous commit is planned for critical topics.
