# Adaptive Ring Expansion (`internal/application/dispatch/discovery/expansion`)

## Purpose
Calculates progressive H3 ring expansion steps based on density classification and search budgets.

## Public Interface
`RingExpander` interface, `AdaptiveExpansionPolicy`, `ExpansionStep` (`Ring`, `RadiusKm`, `CellCount`).

## Dependencies
`internal/infrastructure/geo`.

## Invariants & Gotchas
Ring expansion terminates immediately once candidate target counts or maximum budget limits are reached.

## Human-Approval Required
Yes. Governs spatial search radius and candidate discovery breadth.

## Related ADRs
`docs/02-architecture/3.DISCOVERY_ENGINE.md`.

## Known Issues / Tech Debt
None.
