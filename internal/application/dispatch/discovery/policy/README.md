# Discovery Policies & Budget Rules (`internal/application/dispatch/discovery/policy`)

## Purpose
Defines expansion policies and search budget constraints governing spatial search depth.

## Public Interface
`ExpansionPolicy` interface, `SearchBudgetPolicy`, `PolicyInput`, `BudgetRule`.

## Dependencies
`context`, `internal/config`.

## Invariants & Gotchas
`PolicyInput` is decoupled from runtime search state so that retry-aware or regional budget rules can be added without altering search algorithms.

## Human-Approval Required
Yes. Search budget parameters control system fan-out and resource consumption.

## Related ADRs
`docs/02-architecture/3.DISCOVERY_ENGINE.md`.

## Known Issues / Tech Debt
None.
