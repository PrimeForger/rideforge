# ADR-0030: ADR Index

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Document Type:** Architecture Decision Record Index  
> **Scope:** Complete RideForge Architecture Decision Record catalog  
> **Owner:** RideForge Architecture / Engineering  
> **ADR Range:** `0001`–`0029`  
> **Purpose:** Central navigation, status tracking, dependency mapping, and lifecycle reference for all RideForge ADRs

---

# 1. Purpose

This document is the central index for the RideForge Architecture Decision Records.

The ADR collection records the major architectural decisions that shape:

```text
Architecture
Domain Design
Service Boundaries
Event Streaming
Data
Geospatial Operations
Real-Time State
Dispatch
AI
APIs
Reliability
Observability
Security
Configuration
Testing
Deployment
Cost
Architecture Evolution
```

This index exists to make those decisions:

```text
Discoverable
Traceable
Navigable
Reviewable
Maintainable
```

---

# 2. ADR Directory

The RideForge ADR directory is:

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

---

# 3. ADR Status Convention

RideForge uses the following ADR lifecycle states:

```text
Proposed
Accepted
Deprecated
Superseded
Rejected
```

---

# 4. Status Definitions

## 4.1 Proposed

The decision is under review and has not yet been formally accepted.

---

## 4.2 Accepted

The decision is currently part of the approved architecture.

---

## 4.3 Deprecated

The decision is no longer recommended for new implementation but may still exist in the platform.

---

## 4.4 Superseded

A newer ADR has replaced the decision.

The historical ADR remains preserved for architectural history.

---

## 4.5 Rejected

The decision was considered and explicitly rejected.

The record should remain available to explain why the alternative was not selected.

---

# 5. Current ADR Status

At the time of this index:

```text
ADR-0001 → Accepted
ADR-0002 → Accepted
ADR-0003 → Accepted
ADR-0004 → Accepted
ADR-0005 → Accepted
ADR-0006 → Accepted
ADR-0007 → Accepted
ADR-0008 → Accepted
ADR-0009 → Accepted
ADR-0010 → Accepted
ADR-0011 → Accepted
ADR-0012 → Accepted
ADR-0013 → Accepted
ADR-0014 → Accepted
ADR-0015 → Accepted
ADR-0016 → Accepted
ADR-0017 → Accepted
ADR-0018 → Accepted
ADR-0019 → Accepted
ADR-0020 → Accepted
ADR-0021 → Accepted
ADR-0022 → Accepted
ADR-0023 → Accepted
ADR-0024 → Accepted
ADR-0025 → Accepted
ADR-0026 → Accepted
ADR-0027 → Accepted
ADR-0028 → Accepted
ADR-0029 → Accepted
ADR-0030 → Accepted
```

---

# 6. Complete ADR Table

| ADR | Document | Decision Area | Status |
|---|---|---|---|
| 0001 | `ADR_PROCESS_AND_GUIDELINES` | ADR governance | Accepted |
| 0002 | `ARCHITECTURE_STYLE` | Overall architecture | Accepted |
| 0003 | `MICROSERVICE_BOUNDARIES` | Service boundaries | Accepted |
| 0004 | `DOMAIN_DRIVEN_DESIGN` | Domain architecture | Accepted |
| 0005 | `EVENT_DRIVEN_ARCHITECTURE` | Event architecture | Accepted |
| 0006 | `KAFKA_REDPANDA_FOR_EVENT_STREAMING` | Event streaming | Accepted |
| 0007 | `POSTGRESQL_AS_PRIMARY_DATABASE` | Primary database | Accepted |
| 0008 | `POSTGIS_FOR_GEOSPATIAL_OPERATIONS` | Geospatial database | Accepted |
| 0009 | `REDIS_FOR_REAL_TIME_STATE_AND_CACHING` | Cache / real-time state | Accepted |
| 0010 | `DRIVER_LOCATION_STORAGE_STRATEGY` | Driver location | Accepted |
| 0011 | `PGBOUNCER_FOR_DATABASE_CONNECTION_POOLING` | Database pooling | Accepted |
| 0012 | `OUTBOX_PATTERN` | Reliable event publication | Accepted |
| 0013 | `DEAD_LETTER_QUEUE_STRATEGY` | Failed event processing | Accepted |
| 0014 | `API_AND_SERVICE_COMMUNICATION` | Service communication | Accepted |
| 0015 | `SMART_DISPATCH_AND_STAND_DISPATCH` | Dispatch architecture | Accepted |
| 0016 | `AI_ASSISTED_DISPATCH_STRATEGY` | AI dispatch | Accepted |
| 0017 | `ETA_AND_ROUTE_PROVIDER_STRATEGY` | ETA / routing | Accepted |
| 0018 | `REGIONAL_AND_LEGAL_RIDE_VALIDATION` | Regional / legal rules | Accepted |
| 0019 | `DATA_CONSISTENCY_AND_TRANSACTION_BOUNDARIES` | Data consistency | Accepted |
| 0020 | `IDEMPOTENCY_STRATEGY` | Idempotency | Accepted |
| 0021 | `FAILURE_AND_DEGRADATION_STRATEGY` | Failure handling | Accepted |
| 0022 | `OBSERVABILITY_STRATEGY` | Observability | Accepted |
| 0023 | `SECURITY_AND_SECRET_MANAGEMENT` | Security | Accepted |
| 0024 | `CONFIGURATION_AND_ENVIRONMENT_STRATEGY` | Configuration | Accepted |
| 0025 | `TESTING_AND_INTEGRATION_STRATEGY` | Testing | Accepted |
| 0026 | `MODEL_AND_AI_GOVERNANCE` | AI governance | Accepted |
| 0027 | `CLOUD_AND_DEPLOYMENT_STRATEGY` | Cloud / deployment | Accepted |
| 0028 | `COST_OPTIMIZATION_STRATEGY` | Cost optimization | Accepted |
| 0029 | `ARCHITECTURE_EVOLUTION_AND_MIGRATION` | Architecture evolution | Accepted |
| 0030 | `ADR_INDEX` | ADR index / governance | Accepted |

---

# 7. Architecture Foundation ADRs

These ADRs establish the fundamental architecture.

```text
0001
ADR Process and Guidelines

0002
Architecture Style

0003
Microservice Boundaries

0004
Domain-Driven Design
```

Primary relationship:

```text
ADR-0001
    ↓
ADR-0002
    ↓
ADR-0003
    ↓
ADR-0004
```

---

# 8. Event Architecture ADRs

```text
0005 — Event-Driven Architecture
0006 — Kafka / Redpanda for Event Streaming
0012 — Outbox Pattern
0013 — Dead Letter Queue Strategy
```

Relationship:

```text
Event-Driven Architecture
        │
        ├── Kafka / Redpanda
        │
        ├── Outbox
        │
        └── DLQ
```

These decisions collectively define the event-driven execution model.

---

# 9. Data and Storage ADRs

```text
0007 — PostgreSQL as Primary Database
0008 — PostGIS for Geospatial Operations
0009 — Redis for Real-Time State and Caching
0010 — Driver Location Storage Strategy
0011 — PgBouncer for Database Connection Pooling
0019 — Data Consistency and Transaction Boundaries
```

Relationship:

```text
PostgreSQL
├── PostGIS
├── PgBouncer
└── Transaction Boundaries

Redis
└── Real-Time State / Cache

Driver Location
└── Location Storage Strategy
```

---

# 10. Communication ADRs

```text
0014 — API and Service Communication
0005 — Event-Driven Architecture
0006 — Kafka / Redpanda
0012 — Outbox
0013 — DLQ
```

Communication model:

```text
                Service
               /       \
              /         \
         HTTP/API       Events
                         │
                    Kafka/Redpanda
                         │
                    Consumers
```

---

# 11. Dispatch ADRs

```text
0015 — Smart Dispatch and Stand Dispatch
0016 — AI Assisted Dispatch Strategy
0017 — ETA and Route Provider Strategy
0018 — Regional and Legal Ride Validation
```

Conceptual relationship:

```text
Ride Request
     │
     ▼
Legal / Regional Validation
     │
     ▼
Dispatch Strategy
     │
     ├── Stand Dispatch
     │
     └── Smart Dispatch
             │
             ▼
        Candidate Ranking
             │
             ▼
          ETA / Route
```

---

# 12. Consistency and Reliability ADRs

```text
0019 — Data Consistency and Transaction Boundaries
0020 — Idempotency Strategy
0021 — Failure and Degradation Strategy
```

Relationship:

```text
Transactions
     +
Idempotency
     +
Failure Handling
     =
Reliable Distributed Workflow
```

---

# 13. Operational ADRs

```text
0022 — Observability Strategy
0023 — Security and Secret Management
0024 — Configuration and Environment Strategy
0025 — Testing and Integration Strategy
```

These define the operational foundation required to run the architecture safely.

---

# 14. AI ADRs

```text
0016 — AI Assisted Dispatch Strategy
0026 — Model and AI Governance
```

AI architecture must also remain aligned with the AI documentation under:

```text
05-ai/
```

The ADRs define architectural decisions.

The AI documentation defines the broader AI design, implementation, lifecycle, and operational strategy.

---

# 15. Deployment ADRs

```text
0027 — Cloud and Deployment Strategy
0028 — Cost Optimization Strategy
```

Relationship:

```text
Cloud / Deployment
        │
        └── Cost Optimization
```

---

# 16. Architecture Evolution ADR

```text
0029 — Architecture Evolution and Migration
```

This ADR governs how existing decisions can evolve.

It is therefore related to every other ADR.

---

# 17. ADR Governance

```text
0001 — ADR Process and Guidelines
0030 — ADR Index
```

These provide:

```text
ADR Creation
ADR Review
ADR Status
ADR Discovery
ADR Lifecycle
```

---

# 18. Core Architecture Dependency Map

The high-level architecture relationship is:

```text
                    ADR-0001
                 ADR Governance
                       │
                       ▼
                    ADR-0002
              Architecture Style
                       │
             ┌─────────┴─────────┐
             ▼                   ▼
          ADR-0003            ADR-0004
      Microservices             DDD
             │                   │
             └─────────┬─────────┘
                       ▼
                    ADR-0005
             Event-Driven Architecture
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
       ADR-0006     ADR-0012     ADR-0013
      Streaming      Outbox        DLQ
```

---

# 19. Data Architecture Dependency Map

```text
                  ADR-0007
               PostgreSQL
                    │
          ┌─────────┼─────────┐
          ▼         ▼         ▼
       ADR-0008  ADR-0011  ADR-0019
       PostGIS   PgBouncer  Consistency
          │
          ▼
       ADR-0010
    Driver Location

                  ADR-0009
              Redis / State
```

---

# 20. Dispatch Dependency Map

```text
                 Ride Request
                      │
                      ▼
                  ADR-0018
            Regional / Legal Rules
                      │
                      ▼
                  ADR-0015
          Dispatch Strategy Selection
                 /         \
                /           \
               ▼             ▼
       Stand Dispatch    Smart Dispatch
                              │
                              ▼
                          ADR-0016
                        AI Assistance
                              │
                              ▼
                          ADR-0017
                        ETA / Routing
```

---

# 21. Reliability Dependency Map

```text
             ADR-0019
          Data Consistency
                 │
                 ▼
             ADR-0020
           Idempotency
                 │
                 ▼
             ADR-0021
      Failure / Degradation
                 │
        ┌────────┴────────┐
        ▼                 ▼
     ADR-0013          ADR-0022
       DLQ           Observability
```

---

# 22. Operational Dependency Map

```text
               ADR-0027
          Cloud / Deployment
                  │
        ┌─────────┼─────────┐
        ▼         ▼         ▼
     ADR-0023  ADR-0024  ADR-0025
     Security  Config     Testing
        │         │         │
        └─────────┼─────────┘
                  ▼
               Production
                  │
                  ▼
               ADR-0022
             Observability
                  │
                  ▼
               ADR-0028
          Cost Optimization
```

---

# 23. AI Dependency Map

```text
                 ADR-0015
             Dispatch Strategy
                    │
                    ▼
                 ADR-0016
          AI-Assisted Dispatch
                    │
          ┌─────────┴─────────┐
          ▼                   ▼
       AI Models             ETA
          │                   │
          ▼                   ▼
      ADR-0026             ADR-0017
   AI Governance        ETA / Routing
          │
          ▼
      05-ai/
```

---

# 24. Evolution Dependency Map

```text
                ADR-0029
        Architecture Evolution
                 │
       ┌─────────┼─────────┐
       ▼         ▼         ▼
    Services   Data    Infrastructure
       │         │         │
       ▼         ▼         ▼
     ADR-0003  ADR-0007  ADR-0027
     ADR-0004  ADR-0008  ADR-0028
                ADR-0009
```

---

# 25. Complete Architectural Relationship

```text
                          ADR-0001
                       ADR Governance
                              │
                              ▼
                          ADR-0002
                    Architecture Style
                              │
                 ┌────────────┴────────────┐
                 ▼                         ▼
             ADR-0003                  ADR-0004
          Microservices                   DDD
                 │                         │
                 └────────────┬────────────┘
                              ▼
                          ADR-0005
                   Event-Driven Architecture
                              │
               ┌──────────────┼──────────────┐
               ▼              ▼              ▼
            ADR-0006       ADR-0012       ADR-0013
          Kafka/Redpanda    Outbox           DLQ
               │
               │
      ┌────────┴─────────────────────┐
      ▼                              ▼
  Data Layer                    Service Layer
      │                              │
      ▼                              ▼
  ADR-0007                       ADR-0014
 PostgreSQL                  API / Communication
      │                              │
 ┌────┼────┐                         │
 ▼    ▼    ▼                         ▼
0008 0011 0019                    ADR-0015
PostGIS PgBouncer Consistency     Dispatch
 │                │                │
 ▼                ▼           ┌────┴────┐
0010            0020          ▼         ▼
Location       Idempotency   Stand     Smart
                              │         │
                              │         ▼
                              │      ADR-0016
                              │       AI
                              │         │
                              └────┬────┘
                                   ▼
                                ADR-0017
                              ETA / Routing
                                   │
                                   ▼
                                ADR-0018
                           Regional / Legal
                                   │
                                   ▼
                                ADR-0021
                         Failure / Degradation
                                   │
              ┌────────────────────┼────────────────────┐
              ▼                    ▼                    ▼
           ADR-0022            ADR-0023             ADR-0024
        Observability           Security            Configuration
              │                    │                    │
              └────────────────────┼────────────────────┘
                                   ▼
                                ADR-0025
                                 Testing
                                   │
                                   ▼
                                ADR-0026
                             AI Governance
                                   │
                                   ▼
                                ADR-0027
                         Cloud / Deployment
                                   │
                                   ▼
                                ADR-0028
                           Cost Optimization
                                   │
                                   ▼
                                ADR-0029
                     Architecture Evolution
                                   │
                                   ▼
                                ADR-0030
                              ADR Index
```

---

# 26. ADR Categories

## 26.1 Governance

```text
0001
0030
```

---

## 26.2 Architecture

```text
0002
0003
0004
0029
```

---

## 26.3 Eventing

```text
0005
0006
0012
0013
```

---

## 26.4 Data

```text
0007
0008
0009
0010
0011
0019
```

---

## 26.5 Communication

```text
0014
```

---

## 26.6 Dispatch and Mobility

```text
0015
0016
0017
0018
```

---

## 26.7 Reliability

```text
0020
0021
```

---

## 26.8 Operations

```text
0022
0023
0024
0025
```

---

## 26.9 AI

```text
0026
```

with dispatch integration through:

```text
0016
```

---

## 26.10 Cloud and Infrastructure

```text
0027
0028
```

---

# 27. ADR Reading Order

For a new engineer or architect joining RideForge, the recommended reading order is:

```text
0001
ADR Process and Guidelines

0002
Architecture Style

0004
Domain-Driven Design

0003
Microservice Boundaries

0005
Event-Driven Architecture

0006
Kafka / Redpanda

0007
PostgreSQL

0008
PostGIS

0009
Redis

0010
Driver Location Storage

0011
PgBouncer

0012
Outbox

0013
DLQ

0014
API and Service Communication

0019
Data Consistency

0020
Idempotency

0021
Failure and Degradation

0015
Dispatch

0017
ETA and Route Providers

0018
Regional / Legal Validation

0016
AI-Assisted Dispatch

0026
AI Governance

0022
Observability

0023
Security

0024
Configuration

0025
Testing

0027
Cloud and Deployment

0028
Cost Optimization

0029
Architecture Evolution

0030
ADR Index
```

---

# 28. Recommended Reading by Concern

## If Working on Database

Read:

```text
0007
0008
0010
0011
0019
0020
0021
0028
```

---

## If Working on Dispatch

Read:

```text
0015
0016
0017
0018
0019
0020
0021
0022
0026
```

---

## If Working on Event Processing

Read:

```text
0005
0006
0012
0013
0019
0020
0021
0022
```

---

## If Working on AI

Read:

```text
0015
0016
0017
0026
```

and the complete:

```text
05-ai/
```

documentation set.

---

## If Working on Infrastructure

Read:

```text
0007
0009
0011
0022
0023
0024
0027
0028
0029
```

---

## If Working on Security

Read:

```text
0023
0024
0027
0029
```

---

## If Working on Testing

Read:

```text
0021
0022
0024
0025
0029
```

---

## If Working on Architecture Changes

Read:

```text
0001
0002
0003
0004
0027
0028
0029
0030
```

---

# 29. Cross-Reference Rules

ADR documents should reference related ADRs when a decision depends on another architectural decision.

Use the canonical format:

```text
ADR-XXXX — Title
```

Example:

```text
ADR-0007 — PostgreSQL as Primary Database
```

---

# 30. ADR Reference Integrity

When an ADR references another ADR:

```text
Referenced ADR must exist
ADR number must be correct
Title should match the indexed title
Relationship should remain meaningful
```

---

# 31. Adding a New ADR

New ADRs should:

```text
Use the next available number
Follow ADR-0001 guidelines
Use the standard filename convention
Be added to this index
Have an explicit status
Identify related ADRs
```

---

# 32. ADR Filename Convention

Use:

```text
NNNN-DESCRIPTIVE_TITLE.md
```

Examples:

```text
0031-SOME_ARCHITECTURAL_DECISION.md
0032-ANOTHER_ARCHITECTURAL_DECISION.md
```

---

# 33. ADR Numbering

ADR numbers are sequential.

Do not reuse an ADR number after retirement.

---

# 34. Historical ADR Preservation

Do not delete ADRs merely because they are no longer current.

Historical decisions provide:

```text
Context
Reasoning
Migration History
Architecture Evolution
```

---

# 35. Superseded ADRs

When an ADR is superseded:

```text
Old ADR
   │
   └── Superseded By → New ADR
```

The new ADR should reference the old decision.

---

# 36. ADR Lifecycle

The expected lifecycle is:

```text
Draft
  ↓
Review
  ↓
Accepted
  ↓
Implemented
  ↓
Reviewed
  ↓
Deprecated / Superseded
```

---

# 37. ADR Implementation Status

An accepted ADR does not automatically mean every implementation detail is complete.

The ADR records:

```text
Architectural Decision
```

while implementation tracking belongs in:

```text
Development Documentation
Project Tracking
Implementation Checklists
```

---

# 38. ADR and Implementation Documentation

The relationship is:

```text
ADR
↓
Why / Decision

Architecture Documentation
↓
What / Structure

Development Documentation
↓
How / Implementation

Runbooks
↓
How to Operate
```

---

# 39. ADR and AI Documentation

The relationship is:

```text
ADR
↓
Architectural Decision

05-ai/
↓
AI Architecture + Components + Lifecycle + Operations
```

---

# 40. ADR and Deployment Documentation

The relationship is:

```text
ADR-0027
↓
Deployment Decision

Deployment Documentation
↓
Concrete Infrastructure and Operational Procedures
```

---

# 41. ADR and Cost Documentation

The relationship is:

```text
ADR-0028
↓
Cost Decision

Cost / Operations Documentation
↓
Concrete Monitoring and Optimization Procedures
```

---

# 42. Architecture Evolution

ADR-0029 is the governing ADR for changing other architectural decisions.

Conceptually:

```text
Existing ADR
     │
     ▼
New Requirement
     │
     ▼
ADR-0029 Migration Principles
     │
     ▼
New / Updated ADR
     │
     ▼
Migration
```

---

# 43. Decision Ownership

Architectural decisions should have an accountable owner.

Possible ownership:

```text
Architecture
Platform Engineering
Backend Engineering
AI Engineering
Security
Operations
```

The exact organizational assignment may evolve.

---

# 44. ADR Review

ADR review should evaluate:

```text
Current Validity
Implementation Alignment
Operational Impact
Cost
Security
Scalability
New Requirements
```

---

# 45. ADR Review Triggers

Review an ADR when:

```text
Its Assumptions Change
Its Technology Changes
A Major Incident Challenges It
Cost Becomes Material
Scale Changes
Regulatory Requirements Change
A New Architecture Is Introduced
A Migration Is Planned
```

---

# 46. ADR Consistency

Related ADRs should not intentionally contradict one another.

If two decisions become incompatible:

```text
Create / Update ADR
```

rather than silently implementing a contradiction.

---

# 47. ADR Conflict Resolution

When architectural decisions conflict:

```text
Identify Conflict
      ↓
Identify Higher-Level Constraint
      ↓
Evaluate Current Requirements
      ↓
Create Updated Decision
      ↓
Supersede Old Decision if Necessary
```

---

# 48. Current Architectural Principles

The ADR set collectively establishes these high-level principles:

```text
1. Domain boundaries matter.

2. Services should have meaningful boundaries.

3. Events are used for asynchronous decoupling.

4. PostgreSQL is the primary transactional database.

5. PostGIS supports geospatial operations.

6. Redis supports real-time state and caching.

7. Kafka / Redpanda supports event streaming.

8. Outbox protects reliable event publication.

9. DLQ protects failed event processing.

10. APIs and events have explicit communication contracts.

11. Dispatch supports both Stand and Smart strategies.

12. AI assists dispatch without replacing hard business constraints.

13. ETA is treated as a replaceable capability.

14. Regional and legal ride rules remain domain constraints.

15. Data consistency is explicitly designed.

16. Idempotency is required for retryable distributed workflows.

17. Failures require controlled degradation.

18. Observability is a first-class production requirement.

19. Security and secrets require explicit controls.

20. Configuration is environment-aware.

21. Testing must cover distributed integration behaviour.

22. AI models require governance.

23. Cloud infrastructure should evolve incrementally.

24. Cost must be measured and controlled.

25. Architecture should evolve incrementally rather than through repeated rewrites.
```

---

# 49. Non-Goals of This Index

This document does not replace:

```text
Architecture Documentation
Development Documentation
AI Documentation
Deployment Runbooks
Operational Runbooks
API Documentation
Database Documentation
Testing Documentation
```

It is the navigation and decision registry for ADRs.

---

# 50. ADR Maintenance Checklist

When adding or modifying an ADR:

```text
□ ADR Number Verified
□ Filename Verified
□ Title Verified
□ Status Defined
□ Context Defined
□ Decision Defined
□ Alternatives Documented
□ Consequences Documented
□ Related ADRs Identified
□ Migration Considered
□ Rollback Considered
□ This Index Updated
```

---

# 51. Index Maintenance Checklist

Periodically verify:

```text
□ Every ADR File Exists
□ Every ADR Number Is Unique
□ Every ADR Has a Status
□ Every ADR Has a Title
□ References Are Valid
□ Superseded Relationships Are Correct
□ Categories Are Accurate
□ Architecture Relationships Are Current
□ No Historical ADR Was Accidentally Deleted
□ New ADRs Are Included
```

---

# 52. ADR File Naming Validation

The expected pattern is:

```text
^[0-9]{4}-[A-Z0-9_]+\.md$
```

Examples:

```text
0001-ADR_PROCESS_AND_GUIDELINES.md
0015-SMART_DISPATCH_AND_STAND_DISPATCH.md
0029-ARCHITECTURE_EVOLUTION_AND_MIGRATION.md
0030-ADR_INDEX.md
```

---

# 53. ADR Number Range

Current range:

```text
0001 → 0030
```

Total ADR documents:

```text
30
```

---

# 54. Current ADR Collection

```text
┌─────────────────────────────────────────────────────────────┐
│                    RIDEFORGE ADR SET                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  GOVERNANCE                                                 │
│  0001 ADR Process                                           │
│  0030 ADR Index                                             │
│                                                             │
│  ARCHITECTURE                                               │
│  0002 Architecture Style                                    │
│  0003 Microservice Boundaries                               │
│  0004 Domain-Driven Design                                  │
│  0029 Architecture Evolution                                │
│                                                             │
│  EVENTING                                                   │
│  0005 Event-Driven Architecture                             │
│  0006 Kafka / Redpanda                                      │
│  0012 Outbox                                                │
│  0013 DLQ                                                   │
│                                                             │
│  DATA                                                       │
│  0007 PostgreSQL                                            │
│  0008 PostGIS                                               │
│  0009 Redis                                                 │
│  0010 Driver Location                                       │
│  0011 PgBouncer                                              │
│  0019 Consistency                                           │
│                                                             │
│  COMMUNICATION                                              │
│  0014 API / Service Communication                           │
│                                                             │
│  DISPATCH                                                   │
│  0015 Dispatch                                              │
│  0016 AI Dispatch                                           │
│  0017 ETA / Routing                                         │
│  0018 Regional / Legal                                      │
│                                                             │
│  RELIABILITY                                                │
│  0020 Idempotency                                           │
│  0021 Failure / Degradation                                 │
│                                                             │
│  OPERATIONS                                                 │
│  0022 Observability                                         │
│  0023 Security                                              │
│  0024 Configuration                                         │
│  0025 Testing                                               │
│                                                             │
│  AI                                                         │
│  0026 AI Governance                                         │
│                                                             │
│  CLOUD / COST                                               │
│  0027 Cloud / Deployment                                    │
│  0028 Cost Optimization                                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

# 55. Architectural Decision Flow

The complete decision lifecycle is:

```text
Business Requirement
        │
        ▼
Architecture Problem
        │
        ▼
ADR
        │
        ▼
Implementation
        │
        ▼
Operations
        │
        ▼
Measurement
        │
        ▼
New Requirement
        │
        ▼
Architecture Review
        │
        ▼
New ADR / Superseding ADR
        │
        ▼
Migration
```

---

# 56. Current ADR Baseline

As of this index, the RideForge architecture baseline is:

```text
Architecture:
Microservice + Domain-Driven + Event-Driven

Primary Database:
PostgreSQL

Geospatial:
PostGIS

Real-Time State / Cache:
Redis

Event Streaming:
Kafka / Redpanda

Reliable Event Publication:
Outbox

Failed Event Handling:
DLQ

Database Pooling:
PgBouncer

Service Communication:
APIs + Events

Dispatch:
Stand + Smart

AI:
Assisted / Governed

ETA:
Provider / Model Abstraction

Regional Operations:
Explicit Legal Validation

Consistency:
Explicit Transaction Boundaries

Idempotency:
Required

Failure:
Controlled Degradation

Observability:
First-Class

Security:
First-Class

Configuration:
Environment-Aware

Testing:
Integration-Aware

Deployment:
Containerized + Automated + Incremental

Cost:
Measured + Workload-Driven

Evolution:
Incremental + ADR-Driven
```

---

# 57. Architecture Baseline Rule

This ADR index represents the current documented architecture baseline.

If implementation differs from an accepted ADR:

```text
Do Not Silently Ignore the Difference.
```

Instead:

```text
Identify Difference
      ↓
Determine Whether Implementation or ADR Is Incorrect
      ↓
Correct Implementation
OR
Create / Update ADR
```

---

# 58. ADR as Architectural Source of Truth

For architectural decisions:

```text
ADR
```

is the authoritative record of:

```text
Why a decision was made
What was selected
What alternatives were rejected
What consequences were accepted
```

Implementation code remains the source of truth for actual runtime behaviour.

---

# 59. Documentation Hierarchy

RideForge documentation should conceptually follow:

```text
                    ADRs
                     │
             Architectural Why
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
 Architecture Docs        AI Architecture
          │                     │
          ▼                     ▼
 Development Docs        AI Implementation
          │                     │
          └──────────┬──────────┘
                     ▼
                Runbooks
                     │
                     ▼
             Operational Reality
```

---

# 60. Final Principles

The following principles govern the ADR collection:

```text
1. ADRs record significant architectural decisions.

2. ADRs preserve architectural reasoning.

3. ADR numbers are sequential and never reused.

4. Historical ADRs remain preserved.

5. Current decisions must have explicit status.

6. Superseded decisions must identify their replacement.

7. Major architecture changes require ADR review.

8. Related ADRs should cross-reference one another.

9. This index must remain synchronized with the ADR directory.

10. The index must not replace the detailed ADR documents.

11. ADRs must not become implementation manuals.

12. Implementation details belong in appropriate technical documentation.

13. Architecture decisions should be based on explicit trade-offs.

14. Accepted decisions are subject to future review.

15. Architecture may evolve when requirements change.

16. New architecture must not be introduced merely for novelty.

17. Migration risk must be considered when changing decisions.

18. Cost must be considered in major architecture decisions.

19. Security must be considered in major architecture decisions.

20. Reliability must be considered in major architecture decisions.

21. Observability must be considered in major architecture decisions.

22. Legal and regional constraints must be considered where relevant.

23. AI decisions must remain governed by explicit AI architecture and governance.

24. Infrastructure decisions must remain consistent with deployment and cost ADRs.

25. The ADR collection should remain understandable to engineers joining the project.

26. The ADR collection should provide enough context to understand why the current architecture exists.

27. The architecture should evolve incrementally rather than through repeated uncontrolled rewrites.

28. This index is the central navigation point for the RideForge ADR collection.
```

---

# 61. Completion Status

```text
ADR-0001 → COMPLETE
ADR-0002 → COMPLETE
ADR-0003 → COMPLETE
ADR-0004 → COMPLETE
ADR-0005 → COMPLETE
ADR-0006 → COMPLETE
ADR-0007 → COMPLETE
ADR-0008 → COMPLETE
ADR-0009 → COMPLETE
ADR-0010 → COMPLETE
ADR-0011 → COMPLETE
ADR-0012 → COMPLETE
ADR-0013 → COMPLETE
ADR-0014 → COMPLETE
ADR-0015 → COMPLETE
ADR-0016 → COMPLETE
ADR-0017 → COMPLETE
ADR-0018 → COMPLETE
ADR-0019 → COMPLETE
ADR-0020 → COMPLETE
ADR-0021 → COMPLETE
ADR-0022 → COMPLETE
ADR-0023 → COMPLETE
ADR-0024 → COMPLETE
ADR-0025 → COMPLETE
ADR-0026 → COMPLETE
ADR-0027 → COMPLETE
ADR-0028 → COMPLETE
ADR-0029 → COMPLETE
ADR-0030 → COMPLETE
```

---

# 62. ADR Folder Completion

The planned ADR documentation set is now complete:

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

**ADR documentation set: COMPLETE.**

---

# 63. Next Documentation Layer

The ADR folder is now complete according to the planned architecture documentation sequence.

Future documentation work should not create duplicate ADRs for decisions already represented here.

New ADRs should only be introduced when a genuinely new architectural decision is required.

The next documentation phase should therefore be selected from the remaining project documentation structure rather than automatically extending the ADR folder.

