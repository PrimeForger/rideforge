# RideForge Documentation Audit Report (Phase 1)

> **Document Type:** Read-Only Documentation Audit Report  
> **Status:** Pending User Review & Approval (Phase 1 Deliverable)  
> **Repository:** `rideforge`  
> **Scope:** All 120 files across `docs/` (`00-project`, `01-business`, `02-architecture`, `03-components`, `04-development`, `05-ai`, `adr`, `architecture`, `diagrams`, and root files).  
> **Audit Constraint:** Read-only analysis. No documentation files have been modified, moved, or deleted.

---

## 1. Executive Summary

A comprehensive, read-only audit of the entire `docs/` directory (120 files, ~3.8 MB of text) was conducted to evaluate purpose, overlap, consistency, currency, and context weight.

### Key Audit Findings:
1. **Parallel Architecture Directories (`docs/02-architecture/` vs `docs/architecture/`)**:
   - `docs/architecture/` contains 4 files (`DISCOVERY_ENGINE_ARCHITECTURE.md`, `dispatch_engine_implementation_roadmap.md`, `dispatch_execution_plan.md`, `matching_engine_current_plan.md`) that overlap heavily with `docs/02-architecture/` and `docs/00-project/4.ROADMAP.md`.
2. **Duplicate Architecture Contracts within `02-architecture/`**:
   - `docs/02-architecture/4.CANDIDATE_PIPELINE.md` and `docs/02-architecture/9.CANDIDATE_PIPELINE_ARCHITECTURE.md` both document the Candidate Pipeline. The latter explicitly notes in its header that both describe the same pipeline at differing granularity and must be kept in sync.
3. **Dual Layering between Architecture and Component Specifications**:
   - Every major system (`DISPATCH_ENGINE`, `DISCOVERY_ENGINE`, `CANDIDATE_PIPELINE`, `RANKING_ENGINE`, `GEO_ENGINE`) is documented twice: once in `02-architecture/` and once in `03-components/`. While the component layer is intended to be code/interface-oriented, large conceptual sections are duplicated almost verbatim.
4. **Development Testing & Observability Fragmentation**:
   - Testing is split across three files in `04-development/` (`4.TESTING_STRATEGY.md`, `5.TESTING_GUIDELINES.md`, and `19.INTEGRATION_TESTING_AND_LOCAL_INFRASTRUCTURE.md`).
   - Logging and Observability are split between `12.LOGGING_AND_DEBUGGING.md` and `13.OBSERVABILITY_DEVELOPMENT.md`.
5. **ADR Filename Reference Mismatches**:
   - In `RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md`, ADR-0008 is referenced as `0008-POSTGIS_FOR_GEOSPATIAL_OPERATIONS.md` (actual file: `0008-POSTGIS_FOR_GEOSPATIAL_DATA.md`), and ADR-0009 is referenced as `0009-REDIS_FOR_REAL_TIME_STATE_AND_CACHING.md` (actual file: `0009-REDIS_FOR_REAL_TIME_STATE.md`).
6. **AI/ML Platform Documentation vs Current Codebase Reality**:
   - `docs/05-ai/` contains 22 detailed files (~750 KB) describing a full-scale machine learning platform (model registry, feature store, online/offline pipelines, model serving). The current Go codebase contains no ML services or Python model runners; it is currently focused on core heuristic dispatch, spatial indexing, and transactional state.

---

## 2. File-by-File Inventory & Audit Classification

| # | File Path | Stated Purpose | Status | Audit Notes |
|---|---|---|---|---|
| 1 | `docs/initial_conversation_format.md` | User-facing prompt template to paste into an AI chat directing it to load context instructions. | **Consolidate** | Useful as a template, but redundant when `.agents/rules/` automatically loads context. Propose moving to `docs/00-project/` or `.agents/`. |
| 2 | `docs/RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md` | Master context-loading contract and reading order for AI agents. | **Keep** | Essential core document. Needs minor updates to fix 2 ADR filenames and mention `docs/architecture/` consolidation. |
| 3 | `docs/00-project/1.PROJECT_VISION.md` | Problem statement, target market (Tier 2/3/rural India), value proposition. | **Keep** | Foundational and concise. |
| 4 | `docs/00-project/2.SYSTEM_GOALS.md` | Quantitative latency SLAs (p99 dispatch <500ms), availability, scalability goals. | **Keep** | Clear engineering benchmarks. |
| 5 | `docs/00-project/3.ARCHITECTURE_PRINCIPLES.md` | High-level engineering principles (Single Responsibility, Ports/Adapters, Event-driven). | **Keep** | High-level reference; consistent with ADR-0002. |
| 6 | `docs/00-project/4.ROADMAP.md` | Platform engineering phases (Core dispatch -> Zones/Stands -> AI/ETA -> Scaling). | **Consolidate** | Overlaps with `docs/architecture/dispatch_engine_implementation_roadmap.md` and `dispatch_execution_plan.md`. |
| 7 | `docs/00-project/5.TERMINOLOGY.md` | Comprehensive glossary of domain terms (Stand Dispatch, Smart Dispatch, Driver Lock, etc.). | **Keep** | Authoritative terminology dictionary. |
| 8 | `docs/01-business/1.DISPATCH_BUSINESS_RULES.md` | Primary business rules for matching, stand priority, geographic expansion. | **Keep** | Critical business logic. Heavily referenced by ADRs. |
| 9 | `docs/01-business/2.DRIVER_LIFECYCLE.md` | Driver state machine (Offline -> Available -> Reserved -> Offered -> OnRide). | **Keep** | Clean and authoritative state machine. |
| 10 | `docs/01-business/3.RIDE_LIFECYCLE.md` | Ride state machine (Requested -> Matching -> Offered -> Accepted -> Completed). | **Keep** | Clean and authoritative state machine. |
| 11 | `docs/01-business/4.MATCHING_RULES.md` | Matching criteria, candidate ranking heuristics, offer sequencing. | **Merge-Candidate** | ~60% overlap with `1.DISPATCH_BUSINESS_RULES.md`. Both describe candidate filtering and offer creation. |
| 12 | `docs/01-business/5.RETRY_POLICY.md` | Retry intervals, exponential backoff, search expansion progression. | **Consolidate** | Overlaps with `1.DISPATCH_BUSINESS_RULES.md` Section 9 and `adr/0021`. Could be a dedicated sub-section of business rules. |
| 13 | `docs/02-architecture/1.SYSTEM_ARCHITECTURE.md` | Overall service boundaries, synchronous vs asynchronous communication. | **Keep** | Core architectural blueprint. |
| 14 | `docs/02-architecture/2.DISPATCH_ENGINE.md` | Dispatch orchestration architecture, transactional boundaries, attempt policies. | **Keep** | Authoritative dispatch engine design. |
| 15 | `docs/02-architecture/3.DISCOVERY_ENGINE.md` | H3 search, spatial density classification, ring expansion architecture. | **Keep** | Authoritative discovery design. |
| 16 | `docs/02-architecture/4.CANDIDATE_PIPELINE.md` | Candidate pipeline stages (enrich, filter, rank, heap). | **Merge-Candidate** | Duplicates `9.CANDIDATE_PIPELINE_ARCHITECTURE.md` in the same directory. |
| 17 | `docs/02-architecture/5.RANKING_ENGINE.md` | Driver ranking architecture, multi-factor scoring (ETA, stand, quality). | **Keep** | Authoritative ranking design. |
| 18 | `docs/02-architecture/6.GEO_ENGINE.md` | H3 resolutions (res 7, 8, 9), distance calculation, spatial indexing. | **Keep** | Authoritative geo design. |
| 19 | `docs/02-architecture/7.EVENT_DRIVEN_ARCHITECTURE.md` | Event streaming, Kafka topic design, outbox pattern. | **Keep** | Authoritative event streaming design. |
| 20 | `docs/02-architecture/8.RELIABILITY_ARCHITECTURE.md` | Circuit breakers, graceful degradation, fallback mechanisms. | **Keep** | Authoritative reliability design. |
| 21 | `docs/02-architecture/9.CANDIDATE_PIPELINE_ARCHITECTURE.md` | Detailed candidate pipeline architecture, batching, stage interfaces. | **Merge-Candidate** | See row 16. Should be unified with `4.CANDIDATE_PIPELINE.md` into a single canonical doc. |
| 22 | `docs/03-components/1.DISPATCH_ENGINE.md` | Component-level specification for Dispatch Engine Go packages. | **Keep** | Useful Go package-level specification. |
| 23 | `docs/03-components/2.DISCOVERY_COMPONENTS.md` | Component-level specification for Discovery Go packages. | **Keep** | Detailed component breakdown. |
| 24 | `docs/03-components/3.CANDIDATE_PIPELINE_COMPONENTS.md` | Component-level specification for Candidate Pipeline Go packages. | **Keep** | Detailed component breakdown. |
| 25 | `docs/03-components/4.RANKING_COMPONENTS.md` | Component-level specification for Ranking Go packages. | **Keep** | Detailed component breakdown. |
| 26 | `docs/03-components/5.GEO_COMPONENTS.md` | Component-level specification for Geospatial Go packages. | **Keep** | Detailed component breakdown. |
| 27 | `docs/03-components/6.REALTIME_COMPONENTS.md` | Component-level specification for WebSocket/real-time state. | **Keep** | Detailed component breakdown. |
| 28 | `docs/03-components/7.RESERVATION_COMPONENTS.md` | Component-level specification for atomic driver reservation. | **Keep** | Detailed component breakdown. |
| 29 | `docs/03-components/8.OFFER_AND_DRIVER_RESPONSE_COMPONENTS.md` | Component-level specification for driver offer delivery and response. | **Keep** | Detailed component breakdown. |
| 30 | `docs/03-components/9.EVENT_AND_MESSAGING_COMPONENTS.md` | Component-level specification for Kafka producers/consumers. | **Keep** | Detailed component breakdown. |
| 31 | `docs/03-components/10.PERSISTENCE_COMPONENTS.md` | Component-level specification for PostgreSQL repositories. | **Keep** | Detailed component breakdown. |
| 32 | `docs/03-components/11.OBSERVABILITY_COMPONENTS.md` | Component-level specification for Prometheus metrics, OTel tracing. | **Keep** | Detailed component breakdown. |
| 33 | `docs/03-components/12.CONFIGURATION_AND_COMPOSITION_COMPONENTS.md` | Component-level specification for configuration loading and container DI. | **Keep** | Detailed component breakdown. |
| 34 | `docs/03-components/13.INFRASTRUCTURE_COMPONENTS.md` | Component-level specification for Docker, networking, external services. | **Keep** | Detailed component breakdown. |
| 35 | `docs/03-components/14.APPLICATION_SERVICES.md` | Component-level specification for Application orchestration layer. | **Keep** | Detailed component breakdown. |
| 36 | `docs/04-development/1.DEVELOPMENT_GUIDELINES.md` | General engineering philosophy, standards, and definition of done. | **Keep** | Foundational development guide. |
| 37 | `docs/04-development/2.CODE_STRUCTURE_AND_CONVENTIONS.md` | Project file structure, directory conventions, naming rules. | **Keep** | Practical codebase reference. |
| 38 | `docs/04-development/3.GO_DEVELOPMENT_STANDARDS.md` | Idiomatic Go standards, error wrapping, interface sizing, concurrency. | **Keep** | Core language guidelines. |
| 39 | `docs/04-development/4.TESTING_STRATEGY.md` | Testing pyramid, unit vs integration vs concurrency testing strategy. | **Merge-Candidate** | Heavily overlaps with `5.TESTING_GUIDELINES.md`. |
| 40 | `docs/04-development/5.TESTING_GUIDELINES.md` | Practical test implementation guidelines (test names, table tests, mocks). | **Merge-Candidate** | Should be consolidated with `4.TESTING_STRATEGY.md` into a single complete testing guide. |
| 41 | `docs/04-development/6.DATABASE_DEVELOPMENT.md` | PostgreSQL query rules, connection management, indexing patterns. | **Consolidate** | Overlaps with `16.MIGRATIONS_AND_SCHEMA_CHANGES.md`. |
| 42 | `docs/04-development/7.REDIS_DEVELOPMENT.md` | Redis keyspace design, Lua scripts, pipeline usage, TTLs. | **Keep** | Practical Redis development standards. |
| 43 | `docs/04-development/8.EVENT_AND_MESSAGING_DEVELOPMENT.md` | Kafka producer/consumer implementation rules, partition keys. | **Keep** | Practical Kafka development standards. |
| 44 | `docs/04-development/9.API_DEVELOPMENT.md` | HTTP/REST endpoints, error response schemas, input validation. | **Keep** | Practical API development standards. |
| 45 | `docs/04-development/10.CONFIGURATION_AND_ENVIRONMENT.md` | Environment variables, `.env` file management, config structs. | **Keep** | Practical configuration standards. |
| 46 | `docs/04-development/11.ERROR_HANDLING_AND_VALIDATION.md` | Error categorization, custom error types, input validation. | **Keep** | Critical Go error handling guidelines. |
| 47 | `docs/04-development/12.LOGGING_AND_DEBUGGING.md` | Structured logging (Zap), log levels, context propagation. | **Merge-Candidate** | Overlaps with `13.OBSERVABILITY_DEVELOPMENT.md`. |
| 48 | `docs/04-development/13.OBSERVABILITY_DEVELOPMENT.md` | Metrics instrumentation, OTel traces, health checks. | **Merge-Candidate** | Should merge with `12.LOGGING_AND_DEBUGGING.md` into a unified observability guide. |
| 49 | `docs/04-development/14.PERFORMANCE_AND_OPTIMIZATION.md` | Memory allocation reduction, pprof profiling, bench testing. | **Keep** | Essential for latency-critical dispatch path. |
| 50 | `docs/04-development/15.LOCAL_DEVELOPMENT_SETUP.md` | Local setup instructions (Docker Compose, prerequisites, Go tools). | **Keep** | Developer onboarding guide. |
| 51 | `docs/04-development/16.MIGRATIONS_AND_SCHEMA_CHANGES.md` | Database migration practices, zero-downtime expand/contract rules. | **Consolidate** | Can remain standalone or merge into `6.DATABASE_DEVELOPMENT.md`. |
| 52 | `docs/04-development/17.GIT_AND_BRANCHING_WORKFLOW.md` | Branching conventions, commit message standards, PR guidelines. | **Keep** | Git workflow standards. |
| 53 | `docs/04-development/18.CODE_REVIEW_GUIDELINES.md` | Code review checklist, reviewer expectations, PR hygiene. | **Keep** | Clear review criteria. |
| 54 | `docs/04-development/19.INTEGRATION_TESTING_AND_LOCAL_INFRASTRUCTURE.md` | Local integration testing with Docker/Testcontainers. | **Consolidate** | Closely related to `4.TESTING_STRATEGY.md` and `15.LOCAL_DEVELOPMENT_SETUP.md`. |
| 55 | `docs/04-development/20.DEVELOPMENT_CHECKLIST.md` | Pre-commit and pre-PR engineering checklist. | **Keep** | High-utility summary checklist. |
| 56–77 | `docs/05-ai/1.AI_STRATEGY_AND_VISION.md` through `22.AI_DEVELOPMENT_CHECKLIST.md` (22 files) | Comprehensive AI/ML platform design (feature store, training, serving, A/B testing, fallback). | **Keep (Deferred)** | These 22 files are complete and coherent as a future design suite. However, they describe features not yet in the codebase. Propose keeping them intact in `05-ai/` with an explicit "Future Roadmap / Planned Capabilities" banner so agents know not to look for runtime code yet. |
| 78–107 | `docs/adr/0001-ADR_PROCESS_AND_GUIDELINES.md` through `0030-ADR_INDEX.md` (30 files) | Architectural Decision Records (immutable historical decisions). | **Keep (Immutable)** | Ground Rule 4 applies: Keep ADRs as an immutable record. Reorganize index links or fix filenames if approved, but do not rewrite past decisions. |
| 108 | `docs/architecture/DISCOVERY_ENGINE_ARCHITECTURE.md` | Deep dive into Discovery Engine architecture and H3 design. | **Misplaced / Merge-Candidate** | Sits in `docs/architecture/` rather than `docs/02-architecture/`. Almost identical in scope to `docs/02-architecture/3.DISCOVERY_ENGINE.md`. |
| 109 | `docs/architecture/dispatch_engine_implementation_roadmap.md` | Detailed milestone roadmap for building the dispatch engine. | **Misplaced / Consolidate** | Should be unified with `docs/00-project/4.ROADMAP.md` or `docs/architecture/dispatch_execution_plan.md`. |
| 110 | `docs/architecture/dispatch_execution_plan.md` | Task-level checklist for Milestones 1, 2, 3 (H3 index, adaptive search, candidate pipeline). | **Misplaced / Consolidate** | Represents live implementation progress. Ideal candidate to form the foundation of `docs/STATUS.md` in Phase 3! |
| 111 | `docs/architecture/matching_engine_current_plan.md` | ASCII diagrams and notes on `MatchingEngine -> CandidateSearcher -> H3Strategy`. | **Orphaned / Merge-Candidate** | Quick scratchpad/working note. Should be merged into the relevant architecture doc or archived. |
| 112–121 | `docs/diagrams/01-SYSTEM_CONTEXT_AND_HIGH_LEVEL_ARCHITECTURE.md` through `10-DIAGRAM_INDEX.md` (10 files) | Text/Mermaid/ASCII architecture diagrams. | **Keep** | Well structured, cleanly indexed, provides fast visual comprehension without excessive token overhead. |

---

## 3. Numbered List of Open Questions

The following open questions require your input and decision before any modifications are made in Phase 2.

### Question 1: Resolving the Dual Architecture Directories (`docs/02-architecture/` vs `docs/architecture/`)
* **Files Involved**:
  - `docs/architecture/DISCOVERY_ENGINE_ARCHITECTURE.md`
  - `docs/architecture/dispatch_engine_implementation_roadmap.md`
  - `docs/architecture/dispatch_execution_plan.md`
  - `docs/architecture/matching_engine_current_plan.md`
  - vs `docs/02-architecture/3.DISCOVERY_ENGINE.md` and `docs/00-project/4.ROADMAP.md`
* **Ambiguity / Conflict**:
  RideForge has two separate architecture directories: `docs/02-architecture/` (which follows the numbered `00`, `01`, `02` convention) and `docs/architecture/` (an unnumbered folder containing 4 working drafts/milestones). `DISCOVERY_ENGINE_ARCHITECTURE.md` covers the exact same subsystem as `02-architecture/3.DISCOVERY_ENGINE.md`.
* **Why It Matters**:
  AI agents and human developers have to guess which document is the active source of truth. Having parallel directories causes confusion and wastes context window tokens.
* **Proposed Options**:
  - **Option 1A (Recommended)**: Consolidate all architectural content into `docs/02-architecture/`. Merge unique points from `DISCOVERY_ENGINE_ARCHITECTURE.md` into `02-architecture/3.DISCOVERY_ENGINE.md`. Move the execution milestones (`dispatch_execution_plan.md`) into the new living `docs/STATUS.md` (Phase 3). Remove the redundant `docs/architecture/` folder.
  - **Option 1B**: Keep `docs/architecture/` strictly for "in-flight execution plans / roadmaps" and rename it to `docs/plans/`, keeping `docs/02-architecture/` for completed, static specifications.
  - **Option 1C**: Leave both directories as-is and cross-link them.

---

### Question 2: Consolidating Candidate Pipeline Architecture within `02-architecture/`
* **Files Involved**:
  - `docs/02-architecture/4.CANDIDATE_PIPELINE.md`
  - `docs/02-architecture/9.CANDIDATE_PIPELINE_ARCHITECTURE.md`
* **Ambiguity / Conflict**:
  Both files reside in the same folder and document the same Candidate Pipeline. File `9` even starts with a warning: *"Important: This document is the detailed candidate-pipeline architecture. 02-architecture/4.CANDIDATE_PIPELINE.md describes the higher-level pipeline contract; both documents must remain behaviorally consistent."* Maintaining two large files on the exact same pipeline creates maintenance drift.
* **Why It Matters**:
  If a developer or agent updates one pipeline stage (e.g. fairness filter or heap size) in one file, the other becomes outdated or contradictory.
* **Proposed Options**:
  - **Option 2A (Recommended)**: Merge `9.CANDIDATE_PIPELINE_ARCHITECTURE.md` into `4.CANDIDATE_PIPELINE.md` (combining the high-level contract and the detailed stage-by-stage architecture into a single authoritative document), and remove file `9`.
  - **Option 2B**: Keep both files, but make `4.CANDIDATE_PIPELINE.md` a concise high-level overview that directly links to `9` for all implementation details without repeating the stage specifications.
  - **Option 2C**: Keep both files completely independent as they are today.

---

### Question 3: Matching Rules vs Dispatch Business Rules
* **Files Involved**:
  - `docs/01-business/1.DISPATCH_BUSINESS_RULES.md`
  - `docs/01-business/4.MATCHING_RULES.md`
* **Ambiguity / Conflict**:
  `1.DISPATCH_BUSINESS_RULES.md` defines the entire dispatch sequence (Discovery -> Enrichment -> Filtering -> Ranking -> Offers -> Reservation). `4.MATCHING_RULES.md` defines the matching process using the exact same sequence. Both discuss candidate eligibility, stand preferences, offer timeouts, and retries.
* **Why It Matters**:
  It is unclear where a new business rule regarding driver selection should be added: in `DISPATCH_BUSINESS_RULES` or `MATCHING_RULES`?
* **Proposed Options**:
  - **Option 3A (Recommended)**: Clarify the boundary: `1.DISPATCH_BUSINESS_RULES.md` remains the overarching business policy (stand priority, regional regulations, rider/driver operational policy), while `4.MATCHING_RULES.md` focuses strictly on the matching algorithm's tactical criteria (filtering predicates, scoring weights, tie-breaking). Cross-reference cleanly to eliminate copy-pasted sections.
  - **Option 3B**: Merge `4.MATCHING_RULES.md` directly into `1.DISPATCH_BUSINESS_RULES.md` as Section 6 ("Matching Rules"), producing a single unified business rules document.
  - **Option 3C**: Keep both separate as-is.

---

### Question 4: Testing Documentation Fragmentation in `04-development/`
* **Files Involved**:
  - `docs/04-development/4.TESTING_STRATEGY.md`
  - `docs/04-development/5.TESTING_GUIDELINES.md`
  - `docs/04-development/19.INTEGRATION_TESTING_AND_LOCAL_INFRASTRUCTURE.md`
* **Ambiguity / Conflict**:
  Testing is split across three separate documents: Strategy (`4`), Tactical/Code guidelines (`5`), and Integration/Infrastructure (`19`). A developer writing a test has to consult all three documents.
* **Why It Matters**:
  Splitting testing guidance across three files fragments standards (e.g. mocking rules vs race testing vs testcontainer setup).
* **Proposed Options**:
  - **Option 4A (Recommended)**: Consolidate `4.TESTING_STRATEGY.md` and `5.TESTING_GUIDELINES.md` into a comprehensive `docs/04-development/4.TESTING_GUIDELINES.md` (containing both the high-level testing pyramid strategy and practical Go test examples). Keep `19.INTEGRATION_TESTING...` as a dedicated guide for running local Docker/Kafka/PostgreSQL test environments.
  - **Option 4B**: Merge all three into a single, complete `docs/04-development/4.TESTING.md`.
  - **Option 4C**: Leave all three as separate files.

---

### Question 5: Logging vs Observability in `04-development/`
* **Files Involved**:
  - `docs/04-development/12.LOGGING_AND_DEBUGGING.md`
  - `docs/04-development/13.OBSERVABILITY_DEVELOPMENT.md`
* **Ambiguity / Conflict**:
  `12.LOGGING_AND_DEBUGGING.md` covers Zap structured logging and context propagation. `13.OBSERVABILITY_DEVELOPMENT.md` covers metrics, tracing, and also logs and context propagation.
* **Why It Matters**:
  Tracing and logging both rely on `trace_id` in log fields. Having two separate guidelines risks divergence in logging conventions.
* **Proposed Options**:
  - **Option 5A (Recommended)**: Merge `12.LOGGING_AND_DEBUGGING.md` into `13.OBSERVABILITY_DEVELOPMENT.md` as "Section 2: Structured Logging Standards", providing a unified observability guide (Logs, Metrics, Tracing, Debugging).
  - **Option 5B**: Keep `12` strictly for local debugging routines (pprof, delve, local logs) and `13` strictly for production telemetry (Prometheus metrics, OTel spans).
  - **Option 5C**: Leave both as separate files.

---

### Question 6: ADR Filename Discrepancies in Context Loading Instructions
* **Files Involved**:
  - `docs/RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md` (lines 400–401)
  - `docs/adr/0008-POSTGIS_FOR_GEOSPATIAL_DATA.md`
  - `docs/adr/0009-REDIS_FOR_REAL_TIME_STATE.md`
* **Ambiguity / Conflict**:
  The context-loading instructions reference:
  - `0008-POSTGIS_FOR_GEOSPATIAL_OPERATIONS.md` (actual file: `0008-POSTGIS_FOR_GEOSPATIAL_DATA.md`)
  - `0009-REDIS_FOR_REAL_TIME_STATE_AND_CACHING.md` (actual file: `0009-REDIS_FOR_REAL_TIME_STATE.md`)
* **Why It Matters**:
  Any automated script or strict agent looking for the exact filename listed in the instructions will fail to locate the file.
* **Proposed Options**:
  - **Option 6A (Recommended)**: Update the text in `RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md` to match the actual existing ADR filenames (`0008-POSTGIS_FOR_GEOSPATIAL_DATA.md` and `0009-REDIS_FOR_REAL_TIME_STATE.md`).
  - **Option 6B**: Rename the actual ADR files in `docs/adr/` to match the longer names in the instructions. (Note: Ground Rule 4 advises against modifying ADRs unless necessary).

---

### Question 7: Handling `docs/05-ai/` (22 Files Describing Future ML Architecture)
* **Files Involved**:
  - `docs/05-ai/1.AI_STRATEGY_AND_VISION.md` through `22.AI_DEVELOPMENT_CHECKLIST.md` (22 files)
* **Ambiguity / Conflict**:
  These 22 files represent a thorough, fully fleshed-out vision of an ML platform (feature store, model serving, drift monitoring, etc.). However, the current Go codebase implements heuristic matching, not Python/ONNX ML models.
* **Why It Matters**:
  AI coding agents reading these files might attempt to import nonexistent Python ML libraries or search for feature stores that do not yet exist in the Go backend.
* **Proposed Options**:
  - **Option 7A (Recommended)**: Keep all 22 files intact, but add a standard status header to `05-ai/` documents: `> **Status:** Planned / Future Architecture (Phase 3/4 Roadmap) — Not Yet Implemented in Runtime Codebase`. This preserves 100% of the research/design while preventing AI agents from hallucinating active ML services.
  - **Option 7B**: Consolidate the 22 files into 4–5 higher-level summary documents (Strategy, Architecture, Pipeline, Governance) to significantly reduce repo size.
  - **Option 7C**: Leave as-is with no status headers.

---

## 4. Proposed Consolidation Plan (For Review & Approval)

If you approve the recommended options above, the Phase 2 execution plan will be:

### A. Pre-requisite Checkpoint
* Execute `git commit -m "checkpoint: pre-audit docs snapshot"` to create a complete safety recovery point before touching any file.

### B. Structural Relocations & Cleanups
1. **Eliminate redundant `docs/architecture/` folder**:
   - Merge valuable details from `docs/architecture/DISCOVERY_ENGINE_ARCHITECTURE.md` into `docs/02-architecture/3.DISCOVERY_ENGINE.md`.
   - Incorporate `docs/architecture/dispatch_execution_plan.md` into the new living `docs/STATUS.md` (Phase 3).
   - Merge `docs/architecture/matching_engine_current_plan.md` ASCII diagrams into `docs/02-architecture/2.DISPATCH_ENGINE.md`.
   - Safely remove `docs/architecture/`.
2. **Consolidate Candidate Pipeline in `docs/02-architecture/`**:
   - Merge `9.CANDIDATE_PIPELINE_ARCHITECTURE.md` into `4.CANDIDATE_PIPELINE.md`.
   - Renumber subsequent files in `02-architecture/` cleanly.
3. **Consolidate Development Testing & Observability**:
   - Merge `04-development/4.TESTING_STRATEGY.md` and `04-development/5.TESTING_GUIDELINES.md` into a single, cohesive `4.TESTING_GUIDELINES.md`.
   - Merge `04-development/12.LOGGING_AND_DEBUGGING.md` into `04-development/13.OBSERVABILITY_DEVELOPMENT.md`.
4. **Fix Context Loading References**:
   - Fix ADR-0008 and ADR-0009 filename references in `docs/RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md`.
5. **Mark AI Architecture State**:
   - Add status callouts to `docs/05-ai/` clarifying that the ML platform is planned architecture.

---

## 5. Next Steps

Per the prompt instructions:
1. **No edits have been made to existing documentation.**
2. Please review the **7 Open Questions** above and provide your decisions/preferences.
3. Once approved, we will execute the Git checkpoint and begin Phase 2.
