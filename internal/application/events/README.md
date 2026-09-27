# Application Event Routing (`internal/application/events`)

## Purpose
Defines event envelopes, header propagation, and event routing metadata for application messaging.

## Public Interface
`AppEvent` struct, event naming constants, traceparent extraction helpers.

## Dependencies
`time`, `context`.

## Invariants & Gotchas
Distributed trace context (`traceparent`, `tracestate`) must be preserved in headers across event boundaries.

## Human-Approval Required
No.

## Related ADRs
ADR-0005 (Event-Driven Architecture), ADR-0022 (Observability Strategy).

## Known Issues / Tech Debt
None.
