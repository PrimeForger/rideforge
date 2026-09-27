# Logging Infrastructure (`internal/infrastructure/logger`)

## Purpose
Configures and initializes Uber Zap structured loggers.

## Public Interface
`NewLogger(env string) *zap.Logger`.

## Dependencies
`go.uber.org/zap`, `go.uber.org/zap/zapcore`.

## Invariants & Gotchas
Production mode outputs structured JSON logs; development mode outputs formatted text. Logger must be flushed with `Sync()` at shutdown.

## Human-Approval Required
No.

## Related ADRs
ADR-0022 (Observability Strategy), `docs/04-development/12.LOGGING_AND_DEBUGGING.md`.

## Known Issues / Tech Debt
None.
