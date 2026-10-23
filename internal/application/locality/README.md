# Locality Application Services

## Purpose

Coordinates locality creation and replacement of derived H3 coverage mappings.

## Public Interface

`LocalityService`.

## Dependencies

`internal/domain/locality`, `internal/ports`, and UUID.

## Invariants & Gotchas

Boundary validation and H3 coverage generation must succeed before locality persistence is completed. Locality repositories remain authoritative for persisted locality state.

## Human-Approval Required

YES — locality and geographic boundary changes require review.

## Related ADRs

ADR-0008, ADR-0018, ADR-0019.

## Known Issues / Tech Debt

Administrative authentication and authorization are outside this application service boundary.
