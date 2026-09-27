# Geographic Candidate Refinement (`internal/application/dispatch/discovery/refinement`)

## Purpose
Refines H3 hexagonal candidate sets by calculating precise Haversine distance from pickup coordinates.

## Public Interface
`CandidateRefiner` interface, `HaversineRefiner`, `NoopRefiner`.

## Dependencies
`internal/domain/location`, `internal/shared/geo`.

## Invariants & Gotchas
H3 cells contain area outside circular radius bounds; refinement removes candidates situated outside strict kilometer cutoffs.

## Human-Approval Required
No.

## Related ADRs
ADR-0008 (PostGIS for Geospatial Data).

## Known Issues / Tech Debt
None.
