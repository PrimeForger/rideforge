# RideForge — Living System Status

> **Document Type:** Living Operational Status & Next Tasks  
> **Last Updated:** 2026-09-23  
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

---

## 3. In-Progress Capabilities

* [ ] **Adaptive Ring Expansion:** Dynamic ring expansion based on spatial driver density classification (`internal/application/dispatch/discovery/`).
* [ ] **Search Budget Policy:** Enforcing candidate search budgets via `PolicyInput` and `SearchBudgetFactory` without coupling to runtime state.
* [ ] **Driver Demand & Occupancy Reconciliation:** Background reconciliation workers and TTL recovery for cell demand pointers (`internal/application/demand_recovery_service.go`).

---

## 4. Active Blockers & Known Issues

* [ ] **Windows CGO / MinGW Linker Warning:** Older MinGW GCC (`ld.exe`) on the host throws `unrecognized option '--high-entropy-va'` when compiling tests with CGO enabled. Native pure-Go packages test clean; CGO-dependent packages (`uber/h3-go`) require an updated GCC toolchain or `CGO_ENABLED=0` for pure Go packages.

---

## 5. Next 1–3 Concrete Tasks

1. **Verify Demand & Occupancy Unit Tests:** Ensure test suites for `internal/application/demand_service_test.go` and `internal/application/occupancy_service_test.go` pass and adhere to `4.TESTING_GUIDELINES.md`.
2. **Complete Progressive Ring Searcher Implementation:** Wire `H3Strategy` with `SearchBudget` to exit early once candidate target counts are satisfied in dense cells.
3. **Connect Demand Metrics to Prometheus:** Expose active cell demand counters and reconciliation lag metrics in `internal/infrastructure/observability/metrics.go`.
