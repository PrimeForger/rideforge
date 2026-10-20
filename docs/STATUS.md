# RideForge — Living System Status

> **Document Type:** Living Operational Status & Next Tasks  
> **Last Updated:** 2026-10-20
> **Rule for AI Agents:** Read this file first when starting a session to orient on current state. Only mark tasks "Done" after automated tests/builds pass. Never append past session history here (use `docs/sessions/YYYY-MM-DD.md` for session archives).

---

## 1. Current Phase & Milestone

* **Active Phase:** Phase 1 — Core Dispatch Engine & Spatial State
* **Active Milestone:** Milestone 2 — Adaptive Candidate Search & Density-Aware Expansion
* **Recent Subsystem Focus:** Driver Occupancy & Ride Demand Tracking (`internal/domain/occupancy/`, `internal/application/demand_service.go`)

---

## 2. Completed Capabilities (Verified)

* [x] **H3 Spatial Indexing in Redis:** Atomic Lua script updates, empty cell cleanup, pipeline lookups (`internal/infrastructure/redis/h3_driver_index.go`).
* [x] **Outbox Pattern:** Transactional event publishing for driver state and ride status (`internal/domain/outbox/`, `internal/infrastructure/outbox/`).
* [x] **Hexagonal Architecture Foundation:** Clear separation across `domain`, `application`, `ports`, and `infrastructure`.
* [x] **Documentation System Consolidation:** Complete docs audit, unified testing guidelines (`4.TESTING_GUIDELINES.md`), consolidated candidate pipeline architecture (`4.CANDIDATE_PIPELINE.md`), and sharpened business/matching rule boundaries.
* [x] **Agent Harness Tooling:** Selective integration of ECC Go patterns, testing workflows, and Antigravity customization rules in `.agents/`.
* [x] **Module README & Enforcement Architecture:** Generated code-grounded README.md files across all 46 Go packages with 7-section templates and human-approval guards; mechanically enforced via `.agents/rules/module-readme-sync.md`, `/doc-sync-check` workflow, `.git/hooks/pre-commit`, and GitHub Actions CI.
* [x] **Stage 7 Phase A Demand Projection Recovery:** Pickup coordinates are required, persisted, and carried in the current ride event contract; Redis demand updates are idempotent and atomic per affected cell, with PostgreSQL-backed reconciliation. Focused and project Go checks passed except the Windows race-toolchain limitation recorded below.
* [x] **Stage 7 Phase B Occupancy Corrections:** Redis occupancy reads recover after script-cache loss and propagate command failures; available-driver supply excludes busy drivers and H3 membership cleanup works on cell moves/removal; reconciliation refreshes active hot/warm cell gauges. Full tests, vet, and coverage pass. Phase C implementation is authorized and underway.
* [x] **Stage 7 Phase C Locality Surge (Complete within approved scope):** PostGIS boundaries and operational-resolution H3 mappings update atomically; locality resolution is deterministic; PostgreSQL policy schedules reject overlapping active periods; invalid refreshes retain the last known-good snapshot until maximum staleness; active policies require valid tiers; surge uses Phase B occupancy and falls back to 1.00x for invalid or unavailable inputs. Quote/fare integration is deferred because no quote operation exists and pricing API work is outside this phase. Full tests, vet, and coverage pass. Windows race linking is unavailable in this environment.
* [x] **PostgreSQL Migration and Query Layout:** Root `migrations/` versions 000001–000005 are the schema authority; `sqlc` v1.30.0 validates grouped SQL files and generates repository query code. Fresh PostGIS `up`, full `down`, and `up` passed; the final application tables, constraints, and indexes match the schema snapshot. Live ride persistence and driver UUID-array lookup passed against the migrated schema. The PostgreSQL package, full Go suite, vet, and coverage checks passed with the documented Windows linker setting. A CI workflow was added for migration order, rollback, query generation, and live adapter smoke tests.

---

## 3. In-Progress Capabilities

* [ ] **Adaptive Ring Expansion:** Dynamic ring expansion based on spatial driver density classification (`internal/application/dispatch/discovery/`).
* [ ] **Search Budget Policy:** Enforcing candidate search budgets via `PolicyInput` and `SearchBudgetFactory` without coupling to runtime state.

---

## 4. Active Blockers & Known Issues

* [x] **Windows CGO / MinGW Linker Resolution:** Older MinGW GCC (`ld.exe`) on the host throws `unrecognized option '--high-entropy-va'` when invoking external linking. Resolved for binaries by building with `-ldflags="-linkmode=internal"`, allowing `cmd/api` to compile and link cleanly. Pure-Go packages test with `CGO_ENABLED=0`.
* [ ] **Windows Race Detector Toolchain:** `go test -race ./...` cannot link with the installed Windows Go/MinGW toolchain: the race runtime requires external linking, while the configured internal link mode rejects `runtime/race`. Non-race tests, vet, and coverage checks pass.
* [x] **PostgreSQL Migration Execution:** Migrations 000001–000005, including surge policy overlap enforcement, were applied, rolled back, and reapplied on a disposable PostGIS 15 database. This verifies local schema behavior; no deployed database was changed.

---

## 5. Next 1–3 Concrete Tasks

1. **Quote/fare integration (deferred):** The repository has no fare/quote operation and pricing API work is outside the approved Phase C scope; separate scope approval is needed before implementation.
2. **Complete Progressive Ring Searcher Implementation:** Wire `H3Strategy` with `SearchBudget` to exit early once candidate target counts are satisfied in dense cells.
3. **Connect Demand Metrics to Prometheus:** Expose active cell demand counters and reconciliation lag metrics in `internal/infrastructure/observability/metrics.go`.
