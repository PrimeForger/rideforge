# Discovery Pipeline (`internal/application/dispatch/discovery/pipeline`)

## Purpose
Coordinates candidate discovery execution, managing lookup, expansion, and refinement workflows.

## Public Interface
`DiscoveryPipeline`, `Execute(ctx, request) (*DiscoveryResult, error)`.

## Dependencies
`context`, `internal/application/dispatch/discovery/*`.

## Invariants & Gotchas
Enforces a strict execution deadline (p99 < 50ms) to preserve overall dispatch latency SLAs.

## Human-Approval Required
Yes. Controls candidate discovery execution budget.

## Related ADRs
`docs/02-architecture/3.DISCOVERY_ENGINE.md`.

## Known Issues / Tech Debt
None.
