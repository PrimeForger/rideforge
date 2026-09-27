# Candidate Representation (`internal/application/dispatch/candidate`)

## Purpose
Defines candidate driver data models, candidate collections, and candidate sources used throughout the dispatch pipeline.

## Public Interface
`Candidate` struct (`DriverID`, `Location`, `ETA`, `Score`, `Metadata`), `CandidateSource` interface, `CandidateCollection`.

## Dependencies
`internal/domain/driver`, `internal/domain/location`.

## Invariants & Gotchas
Candidate objects are transient pipeline representations; possessing a candidate object does NOT mean a driver lock has been obtained.

## Human-Approval Required
No.

## Related ADRs
ADR-0015 (Smart Dispatch), `docs/02-architecture/4.CANDIDATE_PIPELINE.md`.

## Known Issues / Tech Debt
Candidate slices allocate memory in hot paths; object pooling will be evaluated under high load.
