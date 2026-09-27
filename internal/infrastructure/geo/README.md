# H3 Geospatial Service Adapter (`internal/infrastructure/geo`)

## Purpose
Wraps the CGO-based `uber/h3-go/v4` spatial index, providing H3 cell calculations, grid disk generation, and spatial distances.

## Public Interface
`H3Service`, `NewH3Service(res int) *H3Service`, `CellForLocation(lat, lng) (string, error)`, `GridDisk(cell string, k int) ([]string, error)`, `DistanceKm(c1, c2 string) float64`.

## Dependencies
`github.com/uber/h3-go/v4`. Requires CGO.

## Invariants & Gotchas
Uses CGO bindings. Default resolution is Res 8 (~460m edge length), balanced for urban and semi-urban ride dispatch.

## Human-Approval Required
No.

## Related ADRs
ADR-0008 (PostGIS for Geospatial Data), ADR-0010 (Driver Location Storage Strategy).

## Known Issues / Tech Debt
Older Windows MinGW linkers emit warnings on `--high-entropy-va` with Go 1.25.
