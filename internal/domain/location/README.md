# Location Domain (`internal/domain/location`)

## Purpose
Encapsulates geographic coordinates, GPS accuracy metrics, bearing, and basic coordinate validation.

## Public Interface
`Location` struct (`Latitude`, `Longitude`, `Bearing`, `Accuracy`, `Timestamp`), `IsValid() bool`.

## Dependencies
`time`. Pure domain.

## Invariants & Gotchas
Latitude must strictly lie in [-90.0, 90.0] and Longitude in [-180.0, 180.0]. Point (0.0, 0.0) is treated as an invalid GPS null-island glitch and rejected.

## Human-Approval Required
No.

## Related ADRs
ADR-0008 (PostGIS for Geospatial Data), ADR-0010 (Driver Location Storage Strategy).

## Known Issues / Tech Debt
Altitude and speed fields are omitted; will be added when highway telematics are supported.
