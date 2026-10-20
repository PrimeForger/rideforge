# Locality Domain (`internal/domain/locality`)

## Purpose
Defines canonical platform locality identity, boundary metadata, and draft/active/inactive lifecycle state. PostGIS boundaries are authoritative; H3 mappings are derived infrastructure data.

## Public Interface
`Locality`, `Status`, `New`, and `CanResolve`.

## Dependencies
`strings` and `github.com/google/uuid`.

## Invariants & Gotchas
Locality identity is a stable UUID. Only active localities can resolve for surge use. The pure domain model has no database, PostGIS, H3, or authorization dependencies.

## Human-Approval Required
YES. Changing locality identity or lifecycle semantics affects geographic eligibility and pricing.

## Related ADRs
ADR-0007 (PostgreSQL as Primary Database), ADR-0008 (PostGIS for Geospatial Data), ADR-0019 (Data Consistency).

## Known Issues / Tech Debt
Admin authentication and authorization are not implemented. Boundary validation and H3 derivation belong to application/infrastructure adapters.
