# Candidate Search Strategy (`internal/application/dispatch/discovery/strategy`)

## Purpose
Implements candidate search strategies, primarily the H3 hexagonal spatial search strategy.

## Public Interface
`SearchStrategy` interface, `H3Strategy`, `NewH3Strategy(...)`.

## Dependencies
`internal/infrastructure/geo`, `internal/ports`, `internal/application/dispatch/discovery/*`.

## Invariants & Gotchas
Checks density classification and evaluates search budget before expanding each concentric H3 ring.

## Human-Approval Required
Yes. Core candidate discovery algorithm.

## Related ADRs
ADR-0008 (PostGIS), ADR-0010 (Driver Location Storage), `docs/02-architecture/3.DISCOVERY_ENGINE.md`.

## Known Issues / Tech Debt
Early-exit wiring is currently being completed in Milestone 2.
