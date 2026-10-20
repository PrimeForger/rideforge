# Location Domain (`internal/domain/location`)

## Purpose
Encapsulates geographic coordinates and coordinate validation.

## Public Interface
`Coordinate` struct (`Lat`, `Lng`), `Coordinate.IsValid() bool`, and `Coordinate.DistanceTo(other Coordinate) float64`.

## Dependencies
Standard library `math`. Pure domain.

## Invariants & Gotchas
Latitude and longitude must be finite and within inclusive geographic bounds ([-90, 90] and [-180, 180]). Point (0, 0) is treated as an invalid GPS null-island sentinel and rejected by `Coordinate.IsValid`.

## Human-Approval Required
No.

## Related ADRs
ADR-0008 (PostGIS for Geospatial Data), ADR-0010 (Driver Location Storage Strategy).

## Known Issues / Tech Debt
Altitude and speed fields are omitted; will be added when highway telematics are supported.
