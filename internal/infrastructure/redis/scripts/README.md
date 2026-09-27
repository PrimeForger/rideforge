# Redis Lua Scripts (`internal/infrastructure/redis/scripts`)

## Purpose
Contains embedded atomic Lua scripts executed on Redis for concurrency-critical driver operations.

## Public Interface
`DriverLockScript`, `H3UpdateScript`, `OfferAckScript`, `DemandScripts`, `SequenceScript`, `DecodeScriptResult`.

## Dependencies
`github.com/redis/go-redis/v9`.

## Invariants & Gotchas
Lua scripts run atomically on the Redis single thread. They must remain O(1) and short to avoid stalling Redis command processing.

## Human-Approval Required
Yes. Controls atomic driver locks, cell membership index updates, and offer acknowledgments.

## Related ADRs
ADR-0009 (Redis for Real-Time State), ADR-0020 (Idempotency Strategy).

## Known Issues / Tech Debt
None.
