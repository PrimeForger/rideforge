# Spatial Candidate Lookup (`internal/application/dispatch/discovery/lookup`)

## Purpose
Executes batch candidate ID lookups across sets of H3 spatial cells in Redis.

## Public Interface
`CandidateLookup` interface, `RedisH3Lookup`.

## Dependencies
`context`, `internal/ports`.

## Invariants & Gotchas
All cell lookups must use pipelined multi-key commands to ensure O(1) network round-trip overhead regardless of ring size.

## Human-Approval Required
No.

## Related ADRs
ADR-0009 (Redis for Real-Time State), ADR-0010 (Driver Location Storage).

## Known Issues / Tech Debt
None.
