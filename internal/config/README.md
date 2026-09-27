# Configuration Management (`internal/config`)

## Purpose
Parses environment variables into typed configuration structs with strict default values, providing immutable runtime settings across the application.

## Public Interface
`Config` struct, `Load() *Config`, sub-configs: `ServerConfig`, `DatabaseConfig`, `RedisConfig`, `KafkaConfig`, `ObservabilityConfig`, `H3Config`, `RankingConfig`, `RealtimeConfig`, `MatchingConfig`, `MatchingRetryConfig`, `LockingConfig`, `DemandConfig`, `OccupancyConfig`, `RoutingConfig`, `ExpansionProfileConfig`.

## Dependencies
`os`, `strconv`, `time`. Standard library only; zero internal or external framework dependencies.

## Invariants & Gotchas
Values are loaded once at startup and treated as immutable thereafter. Unset environment variables fall back to production-safe defaults specified in `getEnvOrDefault` helpers.

## Human-Approval Required
No (unless changing critical matching/locking default timeouts, which require alignment with dispatch rules).

## Related ADRs
ADR-0024 (Configuration and Environment Strategy).

## Known Issues / Tech Debt
Environment variables are read directly from `os.Getenv`; external secret manager integration (HashiCorp Vault / AWS Secrets Manager) is planned for production deployment.
