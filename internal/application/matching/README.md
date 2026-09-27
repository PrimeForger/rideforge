# Matching Engine (`internal/application/matching`)

## Purpose
Orchestrates driver offer matching, priority queues, offer timeouts, and sequential retry dispatch loops.

## Public Interface
`MatchingEngine`, `RideEventHandler`, `DriverOfferHandler`, `MatchingConsumer`, `PriorityQueue`, `Ranker`, `RankingEngine`, `RetryPolicy`.

## Dependencies
`internal/domain/*`, `internal/ports`, `internal/application/dispatch/*`.

## Invariants & Gotchas
Only one driver is locked and offered a ride at a time per matching attempt. Atomic locks must be verified before offer delivery. Expired offers trigger retry with expanded radius.

## Human-Approval Required
Yes. Core matching algorithm, offer dispatch rules, and driver assignment logic.

## Related ADRs
ADR-0015 (Smart Dispatch & Stand Dispatch), ADR-0016 (AI-Assisted Dispatch Strategy), ADR-0020 (Idempotency).

## Known Issues / Tech Debt
Sequential serial offers are used; parallel offer broadcast is planned for high-density surge tiers.
