# H3 Geospatial Service Adapter (`internal/infrastructure/geo`)

## Purpose
Wraps the CGO-based `uber/h3-go/v4` spatial index, providing H3 cell calculations, grid disk generation, and spatial distances.

## Public Interface
`H3Service`, `NewH3Service(resolution, searchRing int) *H3Service`, `Resolution()`, `MaxSearchRing()`, `CellForLocation`, `CellsForGeoJSON`, `CellsInRing`, `DiskCells`, `CellFromString`.

## Dependencies
`github.com/uber/h3-go/v4`. Requires CGO.

## Invariants & Gotchas
Uses CGO bindings. Polygon coverage uses the service's configured operational resolution and returns sorted, deduplicated cells. The caller remains responsible for storing derived locality coverage atomically with the authoritative boundary.

## Human-Approval Required
No.

## Related ADRs
ADR-0008 (PostGIS for Geospatial Data), ADR-0010 (Driver Location Storage Strategy).

## Known Issues / Tech Debt
Older Windows MinGW linkers emit warnings on `--high-entropy-va` with Go 1.25.
