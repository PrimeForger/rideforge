# RideForge AI Agent — Project Context Loading Instructions

> **Document Type:** AI Agent Context / Repository Documentation Loading Contract  
> **Purpose:** Give any AI coding, architecture, documentation, review, or planning agent a single file that explains how to load and maintain complete RideForge project context without requiring the user to manually list documentation files in every conversation.
> **Project:** RideForge
> **Primary Repository:** RideForge / `rideforge`
> **Scope:** Project documentation, architecture decisions, AI architecture, development documentation, and implementation context
> **Instruction Priority:** This document is the context-loading contract. Follow it before beginning substantial project work.

---

# 1. Purpose

You are an AI agent working on the **RideForge** project.

Before answering a substantial RideForge question or modifying the project, you must build context from the project's documentation instead of relying only on:

- the current conversation,
- general programming knowledge,
- assumptions about the architecture,
- memory of previous conversations,
- or generic ride-hailing architecture patterns.

The user should **not need to explicitly mention every documentation file**.

This document exists so the user can provide this single file to an AI agent and the agent can determine:

```text
What to read
In what order to read it
Which documents are authoritative
Which documents are relevant to a task
How to resolve conflicts
How to preserve existing architectural decisions
When additional documents must be loaded
```

---

# 2. Core Instruction

When this document is provided to you, treat the RideForge documentation tree as a **project knowledge base**.

Your first responsibility is:

```text
Load Project Context
        ↓
Understand Architecture
        ↓
Understand Existing Decisions
        ↓
Understand Relevant Domain
        ↓
Understand Current Implementation
        ↓
Only Then Plan / Modify / Review
```

Do **not** immediately propose a new architecture simply because a generic solution appears better.

First determine:

```text
What RideForge already decided
Why it was decided
What is already implemented
What is intentionally deferred
What constraints exist
```

---

# 3. Repository Documentation Root

The primary documentation root is expected to be the project's documentation directory.

If the repository contains a directory such as:

```text
docs/
```

treat it as the primary documentation root.

The expected high-level structure is:

```text
docs/
├── 01-...
├── 02-...
├── 03-...
├── 04-development/
├── 05-ai/
└── adr/
```

If the exact names of the first three directories differ, discover the actual directory names rather than assuming them.

The known completed documentation layers are:

```text
04-development/
05-ai/
adr/
```

---

# 4. Documentation Loading Rule

You are expected to recursively inspect Markdown documentation.

Primary file type:

```text
.md
```

Do not assume that only a few obvious files contain relevant context.

When the task is broad, architecture-related, or ambiguous, inspect the complete documentation structure first.

---

# 5. Required Context Loading Order

Load documentation in the following conceptual order:

```text
01 — Project / Product Foundation
        ↓
02 — Requirements / Domain / Business Context
        ↓
03 — Architecture / System Design
        ↓
04 — Development / Implementation
        ↓
05 — AI / ML
        ↓
ADR — Architecture Decision Records
        ↓
Source Code / Configuration / Infrastructure
```

The exact directory names may vary, but the **conceptual order must remain**.

---

# 6. Why This Order Matters

The documentation hierarchy establishes:

```text
Why the product exists
        ↓
What the system must do
        ↓
How the system is architected
        ↓
How it is implemented
        ↓
How AI fits into the system
        ↓
Why important architectural decisions were made
        ↓
What the actual code currently does
```

Do not reverse this process unless the task is specifically implementation-first.

---

# 7. Phase 1 — Project Foundation

First inspect the project/product foundation documentation.

Look for documents covering:

```text
Product Vision
Project Overview
Goals
Scope
Business Model
Users
Actors
Product Requirements
Functional Requirements
Non-Functional Requirements
Business Constraints
Regional Constraints
Legal Constraints
```

These documents establish the business context.

---

# 8. Phase 2 — Domain and Requirements

Next inspect documentation covering:

```text
Domain Model
Ride Lifecycle
Driver Lifecycle
Passenger Lifecycle
Dispatch
Booking
Payments
Regions
Legal Rules
Driver Operations
Customer Operations
Platform Operations
```

The domain documentation is critical because implementation decisions must preserve business semantics.

---

# 9. Phase 3 — Architecture

Next inspect the main architecture documentation.

Look for:

```text
System Architecture
Service Architecture
Domain Architecture
Data Architecture
Event Architecture
Communication Architecture
Infrastructure Architecture
Security Architecture
Deployment Architecture
Observability Architecture
```

This establishes the intended system design.

---

# 10. Phase 4 — Development Documentation

The completed development documentation is located under:

```text
04-development/
```

This directory describes implementation-oriented decisions and practices.

When working on implementation, testing, APIs, services, databases, infrastructure, or development workflow, this directory is mandatory context.

Do not treat the development documentation as optional when the task concerns code.

---

# 11. Phase 5 — AI Documentation

The completed AI documentation is located under:

```text
05-ai/
```

The planned AI documentation set is:

```text
05-ai/
│
├── 1.AI_STRATEGY_AND_VISION.md
├── 2.AI_ARCHITECTURE.md
├── 3.AI_COMPONENTS_AND_SERVICES.md
├── 4.AI_USE_CASES.md
├── 5.SMART_DISPATCH_AI.md
├── 6.ETA_AND_PREDICTION_SYSTEM.md
├── 7.DRIVER_DEMAND_AND_SUPPLY_PREDICTION.md
├── 8.AI_MATCHING_AND_RANKING.md
├── 9.FEATURE_ENGINEERING.md
├── 10.MACHINE_LEARNING_DATA_PIPELINE.md
├── 11.MODEL_TRAINING_AND_EVALUATION.md
├── 12.MODEL_SERVING_AND_INFERENCE.md
├── 13.MODEL_VERSIONING_AND_REGISTRY.md
├── 14.ONLINE_AND_OFFLINE_FEATURES.md
├── 15.AI_FEEDBACK_AND_LEARNING_LOOP.md
├── 16.AI_SAFETY_AND_GUARDRAILS.md
├── 17.AI_MONITORING_AND_MODEL_OBSERVABILITY.md
├── 18.AI_PERFORMANCE_AND_COST_OPTIMIZATION.md
├── 19.AI_EXPERIMENTATION_AND_A_B_TESTING.md
├── 20.AI_DATA_PRIVACY_AND_GOVERNANCE.md
├── 21.AI_FAILURE_AND_FALLBACK_STRATEGY.md
└── 22.AI_DEVELOPMENT_CHECKLIST.md
```

If these files exist, load them according to relevance.

For broad AI work, load the entire directory.

For a narrow AI task, load:

```text
1.AI_STRATEGY_AND_VISION.md
2.AI_ARCHITECTURE.md
```

first, then the specific relevant documents.

---

# 12. AI Documentation Dependency Rule

For AI-related tasks, do not read only the individual feature document.

For example, if asked to modify Smart Dispatch AI, first understand:

```text
AI Strategy
AI Architecture
AI Components
AI Use Cases
Smart Dispatch AI
AI Matching / Ranking
Feature Engineering
Model Serving
Failure / Fallback
Monitoring
Governance
```

This prevents locally correct but globally incompatible AI implementations.

---

# 13. Phase 6 — Architecture Decision Records

The ADR directory is:

```text
adr/
```

The ADR collection contains the project's architectural decisions.

Current ADR range:

```text
0001 → 0030
```

There are:

```text
30 ADR documents
```

---

# 14. ADR Loading Rule

For a broad architecture task:

```text
Read all ADRs.
```

For a narrow task:

```text
Read ADR-0001 first
Read ADR-0030 as the index
Then load all ADRs directly related to the task.
```

For high-risk architecture changes:

```text
Read the complete ADR collection.
```

---

# 15. ADR Directory

The current planned ADR set is:

```text
adr/
│
├── 0001-ADR_PROCESS_AND_GUIDELINES.md
├── 0002-ARCHITECTURE_STYLE.md
├── 0003-MICROSERVICE_BOUNDARIES.md
├── 0004-DOMAIN_DRIVEN_DESIGN.md
├── 0005-EVENT_DRIVEN_ARCHITECTURE.md
├── 0006-KAFKA_REDPANDA_FOR_EVENT_STREAMING.md
├── 0007-POSTGRESQL_AS_PRIMARY_DATABASE.md
├── 0008-POSTGIS_FOR_GEOSPATIAL_OPERATIONS.md
├── 0009-REDIS_FOR_REAL_TIME_STATE_AND_CACHING.md
├── 0010-DRIVER_LOCATION_STORAGE_STRATEGY.md
├── 0011-PGBOUNCER_FOR_DATABASE_CONNECTION_POOLING.md
├── 0012-OUTBOX_PATTERN.md
├── 0013-DEAD_LETTER_QUEUE_STRATEGY.md
├── 0014-API_AND_SERVICE_COMMUNICATION.md
├── 0015-SMART_DISPATCH_AND_STAND_DISPATCH.md
├── 0016-AI_ASSISTED_DISPATCH_STRATEGY.md
├── 0017-ETA_AND_ROUTE_PROVIDER_STRATEGY.md
├── 0018-REGIONAL_AND_LEGAL_RIDE_VALIDATION.md
├── 0019-DATA_CONSISTENCY_AND_TRANSACTION_BOUNDARIES.md
├── 0020-IDEMPOTENCY_STRATEGY.md
├── 0021-FAILURE_AND_DEGRADATION_STRATEGY.md
├── 0022-OBSERVABILITY_STRATEGY.md
├── 0023-SECURITY_AND_SECRET_MANAGEMENT.md
├── 0024-CONFIGURATION_AND_ENVIRONMENT_STRATEGY.md
├── 0025-TESTING_AND_INTEGRATION_STRATEGY.md
├── 0026-MODEL_AND_AI_GOVERNANCE.md
├── 0027-CLOUD_AND_DEPLOYMENT_STRATEGY.md
├── 0028-COST_OPTIMIZATION_STRATEGY.md
├── 0029-ARCHITECTURE_EVOLUTION_AND_MIGRATION.md
└── 0030-ADR_INDEX.md
```

If the actual repository contains additional ADRs beyond this list, treat those as part of the current architecture and update your context accordingly.

---

# 16. ADR Priority

When determining architectural intent, use this conceptual priority:

```text
Newer Accepted ADR
        >
Older Superseded ADR
```

However:

```text
Historical ADRs must not be ignored
```

because they explain architectural evolution.

If an ADR says it is:

```text
Superseded
```

follow the newer decision.

---

# 17. ADR-0030

Always use:

```text
adr/0030-ADR_INDEX.md
```

as the ADR navigation document.

It provides:

```text
ADR Catalog
Statuses
Categories
Relationships
Reading Order
Architecture Maps
Cross-References
ADR Lifecycle
Current Architecture Baseline
```

Do not assume the index contains every implementation detail.

Use it to identify which detailed ADRs must then be loaded.

---

# 18. ADR-0001

Always use:

```text
adr/0001-ADR_PROCESS_AND_GUIDELINES.md
```

to understand:

```text
How ADRs are created
How ADRs are changed
How ADRs are superseded
How architectural decisions are documented
```

---

# 19. ADR-0029

For architecture migrations, major refactoring, technology replacement, service extraction, database migration, infrastructure migration, or architecture evolution, load:

```text
adr/0029-ARCHITECTURE_EVOLUTION_AND_MIGRATION.md
```

This document governs how the current architecture should evolve safely.

---

# 20. Important Core ADRs

The following decisions are especially important to preserve:

```text
0002 — Architecture Style
0003 — Microservice Boundaries
0004 — Domain-Driven Design
0005 — Event-Driven Architecture
0006 — Kafka / Redpanda
0007 — PostgreSQL
0008 — PostGIS
0009 — Redis
0010 — Driver Location Storage
0012 — Outbox
0013 — DLQ
0014 — API / Service Communication
0015 — Stand + Smart Dispatch
0016 — AI-Assisted Dispatch
0017 — ETA / Route Providers
0018 — Regional / Legal Ride Validation
0019 — Data Consistency
0020 — Idempotency
0021 — Failure / Degradation
0022 — Observability
0023 — Security
0024 — Configuration
0025 — Testing
0026 — AI Governance
0027 — Cloud / Deployment
0028 — Cost Optimization
0029 — Architecture Evolution
```

---

# 21. Core Architecture Baseline

Unless newer documentation explicitly changes the decision, understand RideForge as:

```text
Microservice Architecture
+
Domain-Driven Design
+
Event-Driven Architecture
+
PostgreSQL
+
PostGIS
+
Redis
+
Kafka / Redpanda
+
Outbox Pattern
+
DLQ
+
Controlled Database Connection Pooling
+
API + Event Communication
+
Stand Dispatch
+
Smart Dispatch
+
AI-Assisted Dispatch
+
ETA / Route Provider Abstraction
+
Regional / Legal Ride Validation
+
Explicit Data Consistency
+
Idempotency
+
Controlled Failure Degradation
+
Observability
+
Security
+
Environment-Aware Configuration
+
Integration-Oriented Testing
+
Cloud Deployment
+
Cost-Aware Infrastructure
+
Incremental Architecture Evolution
```

Do not replace any of these with a generic alternative without first checking the relevant ADRs.

---

# 22. Current Implementation vs Intended Architecture

A critical rule:

```text
Documentation = Intended / Accepted Architecture
Code = Current Runtime Implementation
```

These can temporarily differ.

When they differ:

```text
Do not automatically assume the code is wrong.
Do not automatically assume the documentation is wrong.
```

Determine:

```text
Is the implementation incomplete?
Is the documentation outdated?
Is this an intentional migration?
Is there a newer ADR?
```

---

# 23. Conflict Resolution

If documents conflict, follow this process:

```text
1. Identify the conflicting statements.

2. Determine document dates / status.

3. Check for superseding ADRs.

4. Check ADR-0030.

5. Check ADR-0029 for migration context.

6. Check newer project documentation.

7. Check actual implementation.

8. Do not silently invent a resolution.

9. If ambiguity remains, state the conflict explicitly.
```

---

# 24. Never Silently Override Architecture

Do not propose:

```text
MongoDB
instead of PostgreSQL

RabbitMQ
instead of Kafka / Redpanda

Another cache
instead of Redis

Another geospatial database
instead of PostGIS

Pure AI dispatch
instead of hybrid dispatch

Microservice rewrite
instead of incremental evolution
```

simply because these are common alternatives.

First determine whether an existing ADR already governs the decision.

---

# 25. Task-Specific Context Loading

Use this matrix to determine which documentation must be loaded.

| Task | Minimum Context |
|---|---|
| General project work | Foundation + Architecture + ADR Index |
| Backend service | Architecture + Development + Relevant ADRs |
| Database | Architecture + Development + Data ADRs |
| PostgreSQL | ADR-0007 + ADR-0011 + ADR-0019 + ADR-0020 |
| PostGIS | ADR-0008 + ADR-0010 + relevant development docs |
| Redis | ADR-0009 + ADR-0010 + relevant development docs |
| Driver location | ADR-0009 + ADR-0010 + ADR-0022 + relevant AI/dispatch docs |
| Kafka / Redpanda | ADR-0005 + ADR-0006 + ADR-0012 + ADR-0013 + ADR-0021 |
| API | ADR-0014 + Development docs + relevant architecture |
| Dispatch | ADR-0015 + ADR-0016 + ADR-0017 + ADR-0018 + relevant AI docs |
| Smart Dispatch | ADR-0015 + ADR-0016 + ADR-0017 + ADR-0018 + full relevant AI docs |
| Stand Dispatch | ADR-0015 + ADR-0018 + relevant domain/development docs |
| ETA | ADR-0017 + `05-ai/6.ETA_AND_PREDICTION_SYSTEM.md` |
| AI | `05-ai/` + ADR-0016 + ADR-0026 |
| ML | `05-ai/` + ADR-0026 + relevant development docs |
| Security | ADR-0023 + ADR-0024 + deployment docs |
| Configuration | ADR-0024 + development + deployment docs |
| Testing | ADR-0025 + relevant architecture + development docs |
| Deployment | ADR-0027 + ADR-0028 + development/deployment docs |
| Cost | ADR-0028 + ADR-0027 + relevant infrastructure docs |
| Failure handling | ADR-0021 + ADR-0013 + ADR-0020 + ADR-0022 |
| Observability | ADR-0022 + ADR-0021 + development docs |
| Architecture change | Full ADR set + ADR-0029 |
| Migration | ADR-0029 + affected ADRs + affected development docs |
| Documentation change | Relevant documentation hierarchy + ADR-0001 + ADR-0030 |

---

# 26. Smart Dispatch Context Rule

Smart Dispatch is a cross-cutting capability.

When modifying it, do not load only:

```text
5.SMART_DISPATCH_AI.md
```

Also understand:

```text
ADR-0015
ADR-0016
ADR-0017
ADR-0018
ADR-0019
ADR-0020
ADR-0021
ADR-0022
ADR-0026
```

and the relevant `05-ai/` documents.

Smart Dispatch interacts with:

```text
Driver Location
Geospatial Search
Driver Eligibility
Regional Rules
ETA
Ranking
AI Models
Real-Time State
Event Processing
Failure Handling
Observability
```

---

# 27. Driver Location Context Rule

For driver location work, load at minimum:

```text
ADR-0008
ADR-0009
ADR-0010
ADR-0011
ADR-0019
ADR-0021
ADR-0022
```

Then inspect relevant implementation documentation.

Do not redesign location storage without understanding the existing strategy.

---

# 28. Event-Driven Context Rule

For event-driven work, load:

```text
ADR-0005
ADR-0006
ADR-0012
ADR-0013
ADR-0019
ADR-0020
ADR-0021
ADR-0022
```

Pay particular attention to:

```text
Ordering
Delivery
Idempotency
Retry
DLQ
Outbox
Consumer Behaviour
Observability
```

---

# 29. Database Context Rule

For database changes, inspect:

```text
Database Architecture
Development Documentation
ADR-0007
ADR-0008
ADR-0009
ADR-0010
ADR-0011
ADR-0019
ADR-0020
ADR-0021
ADR-0028
```

Do not optimize database infrastructure before understanding:

```text
Query Patterns
Transactions
Connection Pools
Indexes
Geospatial Queries
Caching
Eventual Consistency
```

---

# 30. AI Context Rule

For AI work, inspect:

```text
05-ai/
```

plus:

```text
ADR-0015
ADR-0016
ADR-0017
ADR-0021
ADR-0022
ADR-0026
ADR-0028
```

AI must remain compatible with:

```text
Dispatch
ETA
Failure / Fallback
Observability
Governance
Cost
```

---

# 31. Legal / Regional Context Rule

For anything involving:

```text
Ride Creation
Driver Matching
Dispatch
Cross-Region Rides
Kerala / Karnataka
Regional Operations
```

load:

```text
ADR-0018
```

and the relevant domain/business documentation.

Never remove or bypass legal validation merely to simplify an implementation.

---

# 32. Cost Context Rule

When introducing infrastructure or changing architecture, inspect:

```text
ADR-0027
ADR-0028
```

and evaluate:

```text
Compute
Database
Redis
Messaging
Network
Storage
Observability
AI
External Providers
```

Do not optimize solely for infrastructure price at the expense of required reliability.

---

# 33. Migration Context Rule

For any significant migration:

```text
Read ADR-0029 first.
```

Then read all affected ADRs.

Required concepts include:

```text
Backward Compatibility
Expand-and-Contract
Migration
Validation
Canary
Rollback
Forward Recovery
Decommissioning
Technical Debt
```

---

# 34. Code Context

Documentation is not enough for implementation work.

After understanding the documentation, inspect:

```text
Repository Structure
Relevant Services
Domain Models
Application Services
Repositories
Handlers
Consumers
Producers
Configuration
Tests
Infrastructure
Migrations
```

The code determines the actual current implementation state.

---

# 35. Do Not Read the Entire Codebase Blindly

For implementation tasks:

```text
Documentation First
        ↓
Identify Relevant Components
        ↓
Inspect Target Code
        ↓
Inspect Dependencies
        ↓
Inspect Tests
```

Avoid loading unrelated code unless necessary.

---

# 36. Test Context

Tests are architectural evidence.

Before modifying an existing behaviour, inspect relevant:

```text
Unit Tests
Integration Tests
Contract Tests
Repository Tests
Service Tests
Event Tests
End-to-End Tests
```

Tests may reveal constraints not obvious from documentation.

---

# 37. Configuration Context

Before changing configuration, inspect:

```text
Environment Files
Configuration Modules
Deployment Configuration
Secrets References
Documentation
ADR-0024
```

Never expose secrets in generated output.

---

# 38. Migration Context

Before database or infrastructure migration:

```text
Inspect Existing Migrations
Inspect Current Schema
Inspect Deployment Configuration
Inspect Rollback / Recovery Procedures
Inspect Tests
```

Do not assume the documented target state is already fully implemented.

---

# 39. Documentation Completeness Rule

If a requested task appears to contradict documentation:

```text
Do not immediately rewrite the documentation.
```

First determine whether:

```text
The implementation is incomplete
OR
The documentation is outdated
OR
The architecture is intentionally changing
```

If architecture is intentionally changing:

```text
Create / Update the appropriate ADR
```

according to ADR-0001 and ADR-0029.

---

# 40. No Hallucinated Project Context

Never claim:

```text
Implemented
Completed
Exists
Configured
Migrated
Tested
Production-Ready
```

unless the repository, files, tests, or user-provided evidence support the claim.

---

# 41. Distinguish Three Types of Knowledge

When answering, distinguish:

```text
DOCUMENTED FACT
```

from:

```text
CURRENT IMPLEMENTATION FACT
```

and:

```text
PROPOSED DESIGN
```

Example:

```text
Documented:
RideForge uses Kafka / Redpanda.

Implementation:
The current repository contains a specific producer / consumer implementation.

Proposal:
A new batching mechanism could be introduced.
```

Do not merge these categories.

---

# 42. External Knowledge

General engineering knowledge may be used when useful.

However:

```text
Project Documentation
```

takes precedence over generic architectural assumptions for RideForge-specific decisions.

If outside knowledge is used to challenge an existing decision, explicitly identify it as an alternative or recommendation.

---

# 43. Web Research

Web research is allowed when the task requires:

```text
Current Technology Information
Provider Documentation
Current Pricing
Current API Behaviour
Security Advisories
Current Cloud Capabilities
```

But external information must not silently replace RideForge's documented decisions.

State the distinction:

```text
Current RideForge Decision
vs
External Recommendation
```

---

# 44. Context Summary Before Major Work

Before a major implementation or architecture task, internally establish:

```text
Project Goal
Relevant Domain
Current Architecture
Relevant ADRs
Relevant AI Documents
Relevant Development Documents
Current Implementation
Constraints
Existing Tests
Potential Impact
```

Then proceed.

Do not require the user to repeat this information.

---

# 45. Minimal Context Rule

Do not always load every file.

Use:

```text
Broad Task
→ Broad Context

Narrow Task
→ Narrow Context

Architecture Change
→ Broad Context

Migration
→ Broad Context

Simple Local Fix
→ Relevant Documentation + Relevant Code
```

---

# 46. Full Context Trigger

Load the complete documentation tree when the user asks:

```text
Understand the project
Review architecture
Plan a major feature
Redesign architecture
Perform a migration
Audit the system
Evaluate scalability
Evaluate production readiness
Make major architectural changes
```

---

# 47. Narrow Context Trigger

Use focused loading when the task is clearly isolated.

Examples:

```text
Fix a repository query
Add a unit test
Change a DTO
Fix a typo
Update a local configuration
```

Still inspect the relevant architecture documentation before making changes that affect behaviour.

---

# 48. Context Refresh Rule

If the conversation becomes long or changes direction substantially:

```text
Refresh Relevant Documentation Context
```

Do not rely entirely on conversation memory.

---

# 49. After Implementation

After making code changes:

```text
1. Run relevant tests.

2. Compare implementation against the relevant ADR.

3. Compare implementation against development documentation.

4. Check whether documentation is now outdated.

5. Check whether a new ADR is required.

6. Report exactly what changed.
```

---

# 50. When to Create a New ADR

Consider a new ADR when the change affects:

```text
Architecture Style
Service Boundaries
Database Technology
Messaging Technology
Major Data Strategy
Dispatch Strategy
AI Architecture
Security Architecture
Deployment Architecture
Cloud Strategy
Cost Strategy
Migration Strategy
```

Do not create an ADR for ordinary implementation details unless they represent a significant architectural decision.

---

# 51. When Not to Create an ADR

Normally do not create an ADR for:

```text
Simple Bug Fix
Routine Refactor
Variable Naming
Minor Endpoint Addition
Small Test Addition
Formatting
Non-Architectural Documentation
```

unless the change introduces a significant architectural decision.

---

# 52. Architecture Decision Hierarchy

Use this hierarchy:

```text
Product / Business Constraints
        ↓
Domain Rules
        ↓
Accepted ADRs
        ↓
Architecture Documentation
        ↓
Development Documentation
        ↓
Implementation
        ↓
Tests
```

When a lower layer conflicts with a higher architectural constraint, investigate the conflict before changing behaviour.

---

# 53. Important RideForge Constraints

Always remain aware of these architectural constraints when relevant:

```text
PostgreSQL is the primary transactional database.

PostGIS is used for geospatial operations.

Redis is used for real-time state and caching.

Kafka / Redpanda is used for event streaming.

Outbox is used for reliable event publication.

DLQ is part of failure handling.

PgBouncer is used where database connection pooling requires it.

RideForge supports both Stand Dispatch and Smart Dispatch.

AI assists dispatch but does not replace hard eligibility and legal constraints.

ETA / routing is treated as a replaceable capability.

Regional and legal ride validation is a domain constraint.

Idempotency is required for retryable distributed workflows.

Failures should degrade safely.

Observability is a first-class concern.

Security and secret management are first-class concerns.

Configuration must be environment-aware.

Testing must protect distributed behaviour.

Cloud infrastructure should evolve incrementally.

Cost optimization must not compromise required reliability.

Architecture should evolve incrementally rather than through uncontrolled rewrites.
```

---

# 54. Documentation Change Rule

If you modify documentation:

```text
Preserve Existing Terminology
Preserve Existing Architecture
Preserve Cross-References
Preserve ADR Relationships
Do Not Invent Completed Work
```

If a document must change because the architecture changed:

```text
Update the relevant documentation
AND
Update/create the relevant ADR when required
```

---

# 55. Generated Documentation Rule

When generating a new documentation file for RideForge:

```text
Use Markdown (.md)
Follow the project's existing naming convention
Use professional production-grade documentation
Reference related documents
Reference relevant ADRs
Avoid duplicate decisions
Clearly distinguish current state from future plans
```

---

# 56. Documentation Status

The following documentation layers have been explicitly established as completed in the project context:

```text
04-development/
05-ai/
adr/
```

The ADR layer is:

```text
ADR-0001 → ADR-0030
```

The AI layer is:

```text
22 planned documents
```

The development layer is considered complete according to the project documentation plan.

---

# 57. Important Note About File Availability

If this context file is provided to an AI agent but the actual project documentation is not available in its environment:

```text
Do not pretend to have read the missing files.
```

Instead:

```text
1. Inspect the repository.
2. Locate the documentation directories.
3. Load the available files.
4. Report missing documentation only if it materially affects the task.
```

---

# 58. If Files Have Different Names

Do not fail simply because a filename differs.

Use semantic matching.

For example:

```text
05-ai/2.AI_ARCHITECTURE.md
```

may be represented as:

```text
05-ai/02-AI_ARCHITECTURE.md
```

or another equivalent filename.

The purpose is to locate the corresponding document, not blindly depend on exact numbering.

---

# 59. If Additional Documentation Exists

If the repository contains documentation not listed here:

```text
Inspect it.
```

Additional project documentation may be newer than this context file.

Determine:

```text
Purpose
Scope
Status
Relationship
```

before ignoring it.

---

# 60. If New ADRs Exist

If ADRs greater than `0030` exist:

```text
They are part of the current ADR collection.
```

Read:

```text
0030-ADR_INDEX.md
```

and then the newer ADRs.

Do not assume ADR-0030 remains the latest decision forever.

---

# 61. If ADRs Are Missing

If a referenced ADR is absent:

```text
Do not fabricate its contents.
```

Use the available documentation and state that the referenced decision could not be verified.

---

# 62. If Documentation Is Contradictory

Never silently choose the answer you prefer.

Report:

```text
Conflict
Possible Interpretation
Evidence
Recommended Resolution
```

For architecture conflicts, prefer creating or updating an ADR rather than hiding the inconsistency.

---

# 63. Context Loading Procedure

Use this procedure whenever this file is provided:

```text
STEP 1
Locate the repository.

        ↓

STEP 2
Locate docs/.

        ↓

STEP 3
Inspect the complete documentation tree.

        ↓

STEP 4
Read project/product foundation documents.

        ↓

STEP 5
Read requirements/domain documents.

        ↓

STEP 6
Read architecture documents.

        ↓

STEP 7
Read ADR-0001 and ADR-0030.

        ↓

STEP 8
Load all ADRs relevant to the task.

        ↓

STEP 9
Load 04-development documents relevant to implementation.

        ↓

STEP 10
Load 05-ai documents relevant to AI / dispatch / ETA / ML work.

        ↓

STEP 11
Inspect relevant source code.

        ↓

STEP 12
Inspect relevant tests and migrations.

        ↓

STEP 13
Establish current state vs intended state.

        ↓

STEP 14
Only then plan or implement.
```

---

# 64. Context Loading Output

You do not need to dump the contents of every document to the user.

Instead, internally establish context and, when useful, summarize:

```text
Documents Reviewed
Relevant ADRs
Relevant Architecture
Current Implementation
Important Constraints
Potential Conflicts
```

Keep the user's conversation focused on the actual task.

---

# 65. Context Ledger

For major tasks, maintain an internal conceptual ledger:

```text
PROJECT
├── Product
├── Domain
├── Architecture
├── Services
├── Data
├── Events
├── Dispatch
├── AI
├── Infrastructure
├── Security
├── Observability
├── Testing
└── Evolution
```

Map the task into this ledger before making architectural changes.

---

# 66. Context Relevance Levels

Classify documents as:

```text
P0 — Mandatory
P1 — Highly Relevant
P2 — Supporting
P3 — Not Required
```

For example:

```text
Smart Dispatch

P0:
ADR-0015
ADR-0016
05-ai/5.SMART_DISPATCH_AI.md

P1:
ADR-0017
ADR-0018
ADR-0019
ADR-0020
ADR-0021
ADR-0022
ADR-0026

P2:
ADR-0028
General development documentation

P3:
Unrelated documentation
```

---

# 67. Do Not Over-Context Simple Tasks

The purpose of this document is to eliminate repetitive manual context loading, not to force unnecessary reading for every tiny change.

Use judgment.

---

# 68. Do Not Under-Context Architecture Changes

Conversely, never make a major architecture change after reading only one feature document.

Architecture changes require:

```text
Broad Context
+
Affected ADRs
+
Current Implementation
```

---

# 69. Implementation Safety Rule

Before modifying production-critical components, understand:

```text
Dependencies
Failure Modes
Transactions
Events
Retries
Idempotency
Observability
Tests
Rollback
```

---

# 70. Production-Critical Areas

Treat these as high-impact:

```text
Ride Creation
Ride State
Driver State
Driver Location
Matching
Dispatch
ETA
Payments
Event Processing
Database Transactions
Regional / Legal Validation
Authentication / Authorization
```

For these areas, use broader context.

---

# 71. Dispatch Safety Rule

Never allow AI or optimization logic to bypass:

```text
Driver Eligibility
Regional Rules
Legal Restrictions
Ride State
Safety Constraints
```

Hard constraints must remain authoritative.

---

# 72. Event Safety Rule

Never change an event contract without checking:

```text
Producer
Consumers
Outbox
Retries
DLQ
Replay
Idempotency
Observability
```

---

# 73. Database Safety Rule

Never make a destructive schema or data change without checking:

```text
Current Schema
Migrations
Consumers
Transactions
Backups
Rollback / Recovery
```

---

# 74. AI Safety Rule

Never deploy an AI change without considering:

```text
Fallback
Model Version
Feature Compatibility
Monitoring
Latency
Cost
Safety
Business Constraints
```

---

# 75. Architecture Evolution Safety Rule

Never replace an architectural component merely because another technology is fashionable.

Evaluate:

```text
Problem
Current Limitations
Alternatives
Cost
Operational Complexity
Migration Risk
Rollback
Long-Term Value
```

---

# 76. Final Agent Instruction

When this file is provided, you should behave as though the user has said:

> **"Before doing anything substantial in RideForge, load and understand the project's documented context, follow the established architecture and ADRs, inspect the relevant implementation, and only then make recommendations or changes. I should not have to manually list every documentation file for you."**

That is the purpose of this document.

---

# 77. One-Line Operating Rule

```text
READ THE PROJECT CONTEXT → UNDERSTAND THE DECISIONS → CHECK THE IMPLEMENTATION → THEN ACT.
```

---

# 78. End of Context Contract

This document is intentionally designed to be passed to another AI agent as a **single context-loading instruction file**.

The agent should use it as the entry point into the RideForge documentation ecosystem.

