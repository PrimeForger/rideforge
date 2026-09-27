# API Server Entrypoint (`cmd/api`)

## Purpose
Main application entrypoint that boots the RideForge HTTP server, WebSocket gateway, Kafka event consumers, background outbox and spatial reconciliation workers, and coordinates graceful process shutdown.

## Public Interface
Executable entrypoint (`main()`). Exposes the HTTP REST API, `/ws` WebSocket gateway, `/metrics` Prometheus endpoint, and health check routes configured by `internal/config`.

## Dependencies
Relies on `internal/bootstrap` for dependency injection, `internal/server` for HTTP/WS routing, `internal/infrastructure/observability` for telemetry, `internal/infrastructure/kafka` for consumers, and `internal/infrastructure/outbox` for event publishing.

## Invariants & Gotchas
Tracing and metrics must initialize before the container and server. Signal handling traps `SIGINT` and `SIGTERM` to trigger a 5-second graceful shutdown drain that flushes logs, closes WebSocket connections, and commits consumer offsets before exit.

## Human-Approval Required
No. Operational startup glue; does not own business or financial logic.

## Related ADRs
ADR-0002 (Architecture Style), ADR-0014 (API & Service Communication), ADR-0022 (Observability Strategy), ADR-0024 (Configuration & Environment Strategy).

## Known Issues / Tech Debt
Monolithic container startup; will decompose into separate domain microservice entrypoints (`cmd/dispatch`, `cmd/driver`, etc.) in future milestones.
