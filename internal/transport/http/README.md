# HTTP Transport (`internal/transport/http`)

## Purpose
Composes the HTTP and WebSocket handlers and registers their routes on the caller's `http.ServeMux`. The API entrypoint owns the listener, telemetry wrapper, and shutdown lifecycle.

## Public Interface
`Server`, `NewServer(...) *Server`, and `(*Server).RegisterRoutes(*http.ServeMux)`.

## Dependencies
Relies on `net/http`, `internal/application`, `internal/application/driver`, `internal/application/location`, `internal/config`, `internal/infrastructure/realtime`, `internal/infrastructure/redis`, and `internal/transport/http/handlers`.

## Invariants & Gotchas
Route registration preserves the existing paths. This package composes handlers and does not own business rules or the HTTP listener.

## Human-Approval Required
NO. Transport wiring only.

## Related ADRs
ADR-0014 (API and Service Communication).

## Known Issues / Tech Debt
Authentication and rate limiting are not yet registered. No separate middleware package is needed until middleware is implemented.
