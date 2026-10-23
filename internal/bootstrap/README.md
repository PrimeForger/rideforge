# Dependency Injection & Composition Root (`internal/bootstrap`)

## Purpose

Serves as the application's single composition root, instantiating database connections, Redis clients, Kafka producers, repositories, application services, and background workers.

## Public Interface

`Container` struct (including locality and surge policy repositories/services, policy snapshot, refresh worker, and surge calculator), `NewContainer() (*Container, error)`.

## Dependencies

Imports all `internal/ports`, `internal/application`, `internal/infrastructure/*`, and `internal/config` packages.

## Invariants & Gotchas

This is the ONLY package in the repository permitted to wire concrete infrastructure adapters to application services and ports. PostgreSQL connects through the required `DATABASE_URL` runtime setting; an unset value fails startup without logging credentials. The startup ping uses the configured database connection timeout. The per-process surge policy snapshot is refreshed on a configured interval and expires at its configured maximum staleness; PostgreSQL remains authoritative.

## Human-Approval Required

Yes. Changing container bindings alters dependency graphs, lifecycle management, and transactional boundary wiring.

## Related ADRs

ADR-0002 (Architecture Style), ADR-0004 (Domain-Driven Design), ADR-0019 (Data Consistency and Transaction Boundaries).

## Known Issues / Tech Debt

Single unified container manages all services; will be refactored into smaller modular sub-containers as independent services emerge.
