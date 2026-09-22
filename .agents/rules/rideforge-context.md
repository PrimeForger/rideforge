---
paths:
  - "**/*"
---
# RideForge Architecture & Context Loading Invariant

> **Critical Guideline:** You are working on the **RideForge** ride-hailing backend project.

Before planning, refactoring, or modifying components in this repository:

1. **Mandatory Status & Documentation Inspection**:
   - First, consult [`docs/STATUS.md`](../../docs/STATUS.md) to orient on the current phase, completed items, active blockers, and next 1–3 tasks.
   - You MUST consult and follow [`docs/RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md`](../../docs/RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md).
   - Check existing Architectural Decision Records (ADRs) under [`docs/adr/`](../../docs/adr/) before proposing any structural or architectural changes.
   - Respect the documented standards in [`docs/04-development/`](../../docs/04-development/) (specifically Go standards, testing guidelines, and error handling).

2. **Clean / Hexagonal Architecture Boundaries**:
   - `internal/domain`: Pure business logic, value objects, entities. Strictly no I/O, no database drivers, no HTTP or framework dependencies.
   - `internal/application`: Use cases and orchestration. Interacts with ports only.
   - `internal/ports`: Inbound and outbound contracts (interfaces).
   - `internal/infrastructure`: Concrete implementations of ports (PostgreSQL, Redis, Kafka, OTel).
   - `internal/bootstrap`: Composition root where concrete adapters are wired together.

3. **High-Performance & Low-Latency Constraints**:
   - RideForge handles real-time driver/rider location tracking and dispatching.
   - Avoid unbounded memory allocations in hot paths.
   - Always run table-driven tests with `-race` detection enabled.
