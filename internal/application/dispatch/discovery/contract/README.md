# Discovery Contracts & Request Models (`internal/application/dispatch/discovery/contract`)

## Purpose
Defines contracts and data structures for spatial candidate discovery, isolating discovery requests from downstream pipeline stages.

## Public Interface
`CandidateSearcher` interface, `DiscoveryRequest`, `DiscoveryResult`, `SearchContext`.

## Dependencies
`context`, `internal/domain/location`.

## Invariants & Gotchas
Discovery contracts only return candidate IDs and spatial metadata; they must NOT attempt to load full driver models or perform driver status checks.

## Human-Approval Required
No.

## Related ADRs
ADR-0015 (Smart Dispatch), `docs/02-architecture/3.DISCOVERY_ENGINE.md`.

## Known Issues / Tech Debt
None.
