# Redis Lua Scripts (`internal/infrastructure/redis/scripts`)

## Purpose
Contains embedded atomic Lua scripts executed on Redis for concurrency-critical driver and demand-projection operations.

## Public Interface
`DriverLockScript`, `H3UpdateScript`, `OfferAckScript`, `AddActiveDemandScript`, `RemoveActiveDemandScript`, `ReconcileCellDemandScript`, `GetAndPruneCellDemandScript`, `SequenceScript`, `DecodeScriptResult`.

## Dependencies
`github.com/redis/go-redis/v9`.

## Invariants & Gotchas
Lua scripts run atomically on the Redis single thread. Demand reconciliation is bounded to one H3 cell per invocation and corrects membership and pointers atomically.

## Human-Approval Required
Yes. Controls atomic driver locks, cell membership index updates, offer acknowledgments, and ride-demand projection consistency.

## Related ADRs
ADR-0009 (Redis for Real-Time State), ADR-0020 (Idempotency Strategy).

## Known Issues / Tech Debt
None.
