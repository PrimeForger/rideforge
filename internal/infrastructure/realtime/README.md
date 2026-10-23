# Real-Time Gateway & WebSocket Hub (`internal/infrastructure/realtime`)

## Purpose

Manages active WebSocket connections with driver mobile apps, heartbeat monitoring, and offer dispatch notifications.

## Public Interface

`Hub`, `Client`, `DriverOfferGateway`, `PushNotifier`, `HeartbeatRecoveryWorker`.

## Dependencies

`github.com/gorilla/websocket`, `internal/ports`, `go.uber.org/zap`.

## Invariants & Gotchas

Client writes are synchronized through a channel pump to prevent concurrent WebSocket write panics. Stale connections are disconnected if heartbeats lapse.

## Human-Approval Required

Yes.

## Related ADRs

ADR-0014 (API and Service Communication), ADR-0021 (Failure and Degradation Strategy).

## Known Issues / Tech Debt

WebSocket hub is single-instance; Redis Pub/Sub adapter is needed for multi-node deployments.
