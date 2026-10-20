# PostgreSQL Repositories & Transaction Manager (`internal/infrastructure/postgres`)

## Purpose
Provides PostgreSQL database access, connection management, repository implementations, generated SQL queries, and transaction orchestration.

## Public Interface
`NewDB`, `TxManager`, `RideRepository`, `LocalityRepository`, `SurgePolicyRepository`, `DriverRepository`, `OutboxRepository`, `ProcessedEventRepository`, `DriverMetricsRepository`, `DriverPushTokenRepository`, `DBDriverLocker`. The generated `Queries`, `DBTX`, and `New` are local SQL adapter helpers.

## Dependencies
`database/sql`, `github.com/lib/pq`, `github.com/google/uuid`, `internal/ports`, `internal/domain/*`, `internal/infrastructure/geo`. `sqlc` v1.30.0 generates Go query code from `queries/*.sql` and root `migrations/` using `sqlc.yaml`; it is a development tool, not an application runtime dependency.

Install the pinned [sqlc v1.30.0 release](https://github.com/sqlc-dev/sqlc/releases/tag/v1.30.0) for your platform and run `sqlc generate` from the repository root after editing SQL or migrations. The generated Go files are committed with the source SQL.

## Invariants & Gotchas
PostgreSQL is authoritative for surge policies and tiers. A partial GiST exclusion constraint with half-open effective ranges rejects overlapping ACTIVE policy schedules for one locality; the `btree_gist` extension supports locality UUID equality in that constraint.
Repository SQL lives in `queries/*.sql`; generated `*.sql.go`, `db.go`, and `models.go` are checked in and must be regenerated with `sqlc generate` after query or schema changes. Repository adapters retain domain mapping and transaction ownership. `NewDB` uses a caller-bounded context for its startup ping and closes the handle on failure. Queries must pass context. Connection pool parameters must remain aligned with PgBouncer transaction-pooling limits. Locality boundary and derived H3 mapping replacements run in one PostgreSQL transaction; PostGIS boundaries remain authoritative. Active locality resolution requires the operational H3 resolution and rejects ambiguity. Surge policies and ratio tiers load from PostgreSQL.

## Human-Approval Required
Yes. Governs relational persistence, schema integrity, and database transaction boundaries.

## Related ADRs
ADR-0007 (PostgreSQL as Primary Database), ADR-0011 (PgBouncer), ADR-0019 (Data Consistency).

## Known Issues / Tech Debt
Root `migrations/` is the authoritative versioned schema history. The external `golang-migrate` CLI applies pending versions as a deployment step; the application does not migrate on startup. `migrations.sql` is a reference snapshot of version 000005. The current baseline migrations assume an empty database. The live repository smoke test requires `RIDEFORGE_TEST_DATABASE_URL` pointing to a disposable migrated database; CI supplies it. See `migrations/README.md` for application, rollback, and recovery procedures.
