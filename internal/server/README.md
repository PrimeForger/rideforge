# HTTP & WebSocket Presentation Layer (`internal/server`)

## Purpose
Handles incoming HTTP requests and WebSocket connections, validates parameters, and maps transport requests to application service calls and JSON responses.

## Public Interface
`Server` struct, `NewServer(...) *Server`, `Start() error`, `Shutdown(ctx) error`, HTTP endpoints (`/rides`, `/drivers`, `/ws`).

## Dependencies
Relies on `net/http`, `github.com/gorilla/websocket`, `internal/application/*`, `internal/infrastructure/realtime`, and `internal/config`.

## Invariants & Gotchas
Presentation layer only. Contains zero business logic, zero SQL queries, and zero direct Redis calls; purely decodes JSON/query params, invokes application services, and writes HTTP status codes.

## Human-Approval Required
No. Transport glue layer.

## Related ADRs
ADR-0014 (API and Service Communication).

## Known Issues / Tech Debt
Rate-limiting middleware and JWT authentication validation middleware are currently pending integration.
