# Shared Geospatial Math (`internal/shared/geo`)

## Purpose
Provides pure mathematical distance and bearing calculations without external libraries or CGO dependencies.

## Public Interface
`HaversineDistance(lat1, lon1, lat2, lon2 float64) float64`, `Bearing(...) float64`.

## Dependencies
`math`. Pure Go standard library.

## Invariants & Gotchas
Assumes spherical Earth with radius 6371.0 km. Safe to import from any domain or application package without dependency inversion issues.

## Human-Approval Required
No.

## Related ADRs
ADR-0008 (PostGIS for Geospatial Data).

## Known Issues / Tech Debt
None.
