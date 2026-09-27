# Discovery Candidate Selector (`internal/application/dispatch/discovery/selector`)

## Purpose
Selects and trims candidate sets discovered across spatial rings before handoff to candidate pipeline.

## Public Interface
`Selector` interface, `DefaultSelector`, `TopKSelector`.

## Dependencies
`internal/application/dispatch/discovery/contract`.

## Invariants & Gotchas
Maintains deterministic candidate ordering to ensure reproducible matching results.

## Human-Approval Required
No.

## Related ADRs
`docs/02-architecture/3.DISCOVERY_ENGINE.md`.

## Known Issues / Tech Debt
None.
