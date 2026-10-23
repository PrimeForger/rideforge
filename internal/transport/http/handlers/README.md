# HTTP Handlers (`internal/transport/http/handlers`)

## Purpose

Decodes ride and driver HTTP requests, writes HTTP responses, and upgrades driver WebSocket connections. Ride creation requires pickup latitude and longitude.

## Public Interface

`RideHandler`, `NewRideHandler`, `DriverHandler`, `NewDriverHandler`, `WebSocketHandler`, and `NewWebSocketHandler`. Their exported methods handle the routes registered by the parent transport package.

## Dependencies

Uses `net/http`, `internal/application`, `internal/application/driver`, `internal/application/location`, `internal/config`, `internal/domain/driver`, `internal/domain/region`, `internal/infrastructure/realtime`, `internal/infrastructure/redis`, `github.com/google/uuid`, and `github.com/gorilla/websocket`.

## Invariants & Gotchas

Handlers validate and map transport input before invoking application services. The WebSocket handler upgrades the HTTP request and delegates client lifecycle to the realtime adapter. Authentication and origin restrictions are not yet enforced by this package.

## Human-Approval Required

NO. Transport request and response mapping only.

## Related ADRs

ADR-0014 (API and Service Communication), ADR-0023 (Security and Secret Management).

## Known Issues / Tech Debt

Driver WebSocket identity currently comes from a query parameter, and the upgrader accepts all origins. Authentication and origin validation remain pending.
