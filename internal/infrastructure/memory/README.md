# In-Memory Mock Repositories (`internal/infrastructure/memory`)

## Purpose
Provides in-memory mock repository implementations for unit testing and local development without databases.

## Public Interface
`InMemoryRideRepository`, `NewInMemoryRideRepository()`.

## Dependencies
`sync.RWMutex`, `internal/domain/ride`, `internal/ports`.

## Invariants & Gotchas
Thread-safe via RWMutex. State is strictly ephemeral and lost on process restart. Never use in production.

## Human-Approval Required
No.

## Related ADRs
`docs/04-development/4.TESTING_GUIDELINES.md`.

## Known Issues / Tech Debt
None.
