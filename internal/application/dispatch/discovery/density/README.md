# Spatial Density Classification (`internal/application/dispatch/discovery/density`)

## Purpose
Classifies spatial cells into density tiers (Sparse, Normal, Dense) based on active driver counts.

## Public Interface
`DensityClassifier` interface, `DriverDensityProvider`, `DensityClassification` enum (`Sparse`, `Normal`, `Dense`).

## Dependencies
`context`, `internal/config`.

## Invariants & Gotchas
In sparse areas, search policies expand rapidly with wider ring jumps; in dense areas, search remains tight to prevent excessive Redis pipeline traffic.

## Human-Approval Required
No.

## Related ADRs
ADR-0015 (Smart Dispatch), `docs/02-architecture/3.DISCOVERY_ENGINE.md`.

## Known Issues / Tech Debt
Density thresholds are globally configured; dynamic per-city thresholds will be introduced in Phase 2.
