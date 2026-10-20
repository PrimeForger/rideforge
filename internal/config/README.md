# Configuration Management (`internal/config`)

## Purpose
Parses environment variables into typed configuration structs with strict default values, providing immutable runtime settings across the application.

## Public Interface
`Config` struct, `Load() *Config`, and sub-configs: `H3Config`, `RankingConfig`, `RealtimeConfig`, `MatchingConfig`, `MatchingRetryConfig`, `LockingConfig`, `DemandConfig`, `OccupancyConfig`, `SurgePolicyConfig`, `RoutingConfig`, `ExpansionProfileConfig`, and `ObservabilityConfig`. `Config.DatabaseURL` reads the required runtime `DATABASE_URL`.

## Dependencies
`os`, `strconv`, `time`. Standard library only; zero internal or external framework dependencies.

## Invariants & Gotchas
`OCCUPANCY_MIN_DEMAND_FOR_HOT` defaults to 2 and controls Phase B cell classification. The separate PostgreSQL surge policy `minimum_demand` defaults to 3 and controls whether Phase C applies a multiplier, so a HOT cell can still receive the base multiplier.
Values are loaded once at startup and treated as immutable thereafter. Unset or invalid environment variables fall back to production-safe defaults specified in the environment parsing helpers. Surge policy snapshots refresh using `SURGE_POLICY_REFRESH_INTERVAL_SECONDS`; consumers stop using a snapshot older than `SURGE_POLICY_MAX_STALENESS_SECONDS` and safely use the base multiplier. Both settings are normalized so the staleness limit exceeds the refresh interval.
`DATABASE_URL` has no built-in credential default; bootstrap returns an error when it is unset. `DATABASE_CONNECT_TIMEOUT` controls the bounded startup ping and defaults to 5 seconds when unset or invalid. The migration job uses a separately supplied connection URL for the same logical database.

## Human-Approval Required
No (unless changing critical matching/locking default timeouts, which require alignment with dispatch rules).

## Related ADRs
ADR-0024 (Configuration and Environment Strategy).

## Known Issues / Tech Debt
Environment variables are read directly from `os.Getenv`; external secret manager integration (HashiCorp Vault / AWS Secrets Manager) is planned for production deployment.
