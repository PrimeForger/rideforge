# Telemetry & Metrics Infrastructure (`internal/infrastructure/observability`)

## Purpose
Initializes OpenTelemetry distributed tracing and registers Prometheus metrics for HTTP, dispatch, Kafka, Redis, and surge decisions.

## Public Interface
`InitTracing(...) func(ctx) error`, `Register()`, `SurgeMetrics`, and Prometheus metrics including `SurgeDecisionsTotal` and `SurgePolicyRefreshesTotal`.

## Dependencies
`github.com/prometheus/client_golang/prometheus`, `go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/sdk/trace`.

## Invariants & Gotchas
Prometheus label values must strictly maintain low cardinality; never use user IDs, ride IDs, H3 cells, locality IDs, or timestamps as metric labels. Surge decision and refresh metrics use fixed result labels.

## Human-Approval Required
No.

## Related ADRs
ADR-0022 (Observability Strategy), `docs/04-development/13.OBSERVABILITY_DEVELOPMENT.md`.

## Known Issues / Tech Debt
OTel collector endpoint is optional; traces drop silently if unconfigured.
