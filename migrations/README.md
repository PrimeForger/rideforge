# PostgreSQL schema migrations

The numbered `*.up.sql` files are the authoritative schema history. Run them in order with `golang-migrate` v4.20.1. The matching `*.down.sql` files support verification on disposable databases; production schema corrections should use a new forward migration. Once a migration has been deployed, leave it unchanged.

`000001` creates the original core tables. `000002` adds required ride pickup coordinates, `000003` adds locality and H3 mapping tables, `000004` adds surge policy tables, and `000005` rejects overlapping active policy periods. The `000002` migration assumes an empty initial database, as agreed for the current development setup.

## Apply

Install the pinned [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate), including its PostgreSQL driver. Supply a direct PostgreSQL connection URL through `RIDEFORGE_MIGRATION_DATABASE_URL`. The application reads `DATABASE_URL`; both URLs must target the same logical database and may use different roles. The database role needs schema and extension creation privileges, and the PostgreSQL server must have PostGIS and `btree_gist` available. Use a dedicated deployment job for migrations; application startup does not run them.

From the repository root in PowerShell:

```powershell
migrate -path migrations -database $env:RIDEFORGE_MIGRATION_DATABASE_URL version
migrate -path migrations -database $env:RIDEFORGE_MIGRATION_DATABASE_URL up
```

Use a direct database connection for migrations when the application uses PgBouncer transaction pooling. Migration failure leaves the database version dirty; inspect the failed change and actual schema before any manual version repair. Never automate `force` against an unknown schema.

For a disposable database only, verify reversibility with `migrate ... down 5` followed by `migrate ... up`. Existing manually provisioned development databases should be recreated empty before adopting this migration history.

The complete schema snapshot at `internal/infrastructure/postgres/migrations.sql` is a review reference. Do not apply it as an installation script.

For local Docker Compose, set `POSTGRES_PASSWORD` in your environment before starting the PostGIS service, then set the application `DATABASE_URL` and migration URL for that database. Keep credentials out of committed files.
