# Candidate Pipeline Stages (`internal/application/dispatch/candidate/stage`)

## Purpose
Implements modular stages for candidate processing: batch driver loading, validation, deduplication, availability filtering, ranking, and heap building.

## Public Interface
`Stage` interface, `DriverLoader`, `ValidationFilter`, `DuplicateFilter`, `AvailabilityFilter`, `AlreadyOfferedFilter`, `RankingStage`, `HeapBuilder`.

## Dependencies
`internal/application/dispatch/candidate`, `internal/application/dispatch/candidate/ranking`, `internal/application/dispatch/candidate/heap`.

## Invariants & Gotchas
Duplicate filter must execute before driver batch loading to eliminate redundant database or Redis driver queries.

## Human-Approval Required
Yes. Filters exclude drivers from matching.

## Related ADRs
`docs/02-architecture/4.CANDIDATE_PIPELINE.md`.

## Known Issues / Tech Debt
None.
