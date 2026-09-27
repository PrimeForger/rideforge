# Discovery Search Budget & State (`internal/application/dispatch/discovery/search`)

## Purpose
Tracks search execution state, budget limits, and candidate collection during spatial search runs.

## Public Interface
`SearchBudget`, `SearchBudgetFactory`, `SearchRequest`, `SearchResult`, `SearchState`.

## Dependencies
`internal/config`.

## Invariants & Gotchas
Budgets enforce hard caps on both the maximum number of H3 cells checked and total candidate IDs accumulated.

## Human-Approval Required
No.

## Related ADRs
`docs/02-architecture/3.DISCOVERY_ENGINE.md`.

## Known Issues / Tech Debt
None.
