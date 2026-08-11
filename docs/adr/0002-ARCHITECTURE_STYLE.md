# ADR-0002: Architecture Style

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Architecture  
> **Scope:** RideForge platform architecture  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is a production-oriented ride-hailing platform that must support:

- Ride lifecycle management
- Driver availability and location
- Matching and dispatch
- Stand-based dispatch
- Smart dispatch
- ETA and routing
- Payments and other supporting capabilities
- Event-driven workflows
- Real-time operational state
- AI-assisted optimization
- Regional and legal operating constraints

The platform is expected to evolve over time.

The architecture therefore needs to balance:

```text
Domain Isolation
+
Independent Evolution
+
Operational Reliability
+
Development Speed
+
Infrastructure Cost
+
Future Scalability
```

A ride-hailing system can become overly coupled if every capability is implemented inside one large application.

At the same time, prematurely splitting every capability into an independent deployable service creates unnecessary:

```text
Network Calls
Operational Complexity
Deployment Complexity
Distributed Transactions
Observability Requirements
Infrastructure Cost
```

The architecture must therefore provide clear boundaries without forcing premature distribution.

---

# 2. Problem

RideForge needs an architectural style that can support:

```text
Current Development
        ↓
Production MVP
        ↓
Growing Traffic
        ↓
Independent Scaling
        ↓
Service Extraction
```

without requiring a complete rewrite between these stages.

The key architectural question is:

> **Should RideForge be implemented as a monolith, a fully distributed microservice system from the beginning, or a modular architecture that preserves future microservice boundaries?**

---

# 3. Decision

RideForge will use a:

> **Domain-oriented, modular, microservice-compatible architecture with event-driven integration where appropriate.**

The architecture will prioritize:

```text
Strong Domain Boundaries
+
Clear Application Boundaries
+
Explicit Infrastructure Boundaries
+
Independent Components
+
Event-Driven Integration Where Valuable
+
Incremental Service Extraction
```

RideForge will **not** require every domain component to become an independently deployable microservice immediately.

---

# 4. Architectural Direction

The intended evolution is:

```text
Domain-Oriented Components
          ↓
Modular Application Boundaries
          ↓
Production Components
          ↓
Independently Scalable Services
```

The architecture should make service extraction possible without making service extraction mandatory.

---

# 5. Architecture Style

The selected style combines several complementary principles:

```text
Domain-Driven Design
+
Modular Architecture
+
Microservice-Compatible Boundaries
+
Event-Driven Architecture
+
API-Based Integration
```

These are not competing architectural styles in RideForge.

They operate at different levels.

---

# 6. Domain-Driven Boundaries

RideForge will organize business capabilities around domains rather than technical layers alone.

Examples include:

```text
Ride
Driver
Matching
Dispatch
Location
ETA
Payment
Notification
AI / Optimization
```

The exact service boundaries may evolve as the system grows.

---

# 7. Domain Ownership

Each domain should have clear ownership over its business rules.

For example:

```text
Ride Domain
    ↓
Ride Lifecycle

Matching / Dispatch
    ↓
Candidate Selection / Dispatch Strategy

Location
    ↓
Driver Location State

AI
    ↓
Prediction / Optimization
```

AI does not become the owner of domain rules merely because it participates in a decision.

---

# 8. Modular Architecture

Within a deployable application, modules should remain logically separated.

A module should expose:

```text
Public Contract
```

rather than allowing unrelated modules to depend directly on internal implementation details.

---

# 9. Module Boundaries

A module should generally contain:

```text
Domain Logic
Application Logic
Persistence Interfaces
Integration Interfaces
```

according to the responsibilities of that domain.

The goal is:

```text
High Cohesion
+
Low Coupling
```

---

# 10. Dependency Direction

Dependencies should generally flow toward stable domain contracts.

A simplified direction is:

```text
Interfaces
    ↓
Application
    ↓
Domain

Infrastructure
    ↓
Implements Required Interfaces
```

Infrastructure should not silently become the owner of business decisions.

---

# 11. Domain Independence

Domain logic should avoid unnecessary dependencies on:

```text
HTTP
Kafka
Redis
PostgreSQL
External APIs
Cloud Providers
```

where those dependencies are not intrinsic to the business rule.

---

# 12. Infrastructure Independence

Infrastructure implementations should be replaceable where practical.

For example:

```text
Repository Interface
        ↓
PostgreSQL Implementation
```

rather than embedding PostgreSQL-specific behaviour throughout domain logic.

---

# 13. API Boundaries

Synchronous communication should use explicit APIs or service contracts where appropriate.

An API boundary should define:

```text
Request
Response
Errors
Authentication
Timeout
Versioning
```

---

# 14. Event Boundaries

Asynchronous communication should use explicit domain or integration events where appropriate.

Events should represent meaningful facts or integration messages rather than arbitrary internal implementation details.

---

# 15. Synchronous vs Asynchronous Communication

RideForge will use:

```text
Synchronous Communication
```

when the caller requires an immediate response.

Examples:

```text
Create Ride
Validate Request
Retrieve Current State
```

Use:

```text
Asynchronous Events
```

when immediate coupling is unnecessary.

Examples may include:

```text
Ride Created
Ride State Changed
Match Completed
Notification Trigger
Analytics Event
```

The exact event set is defined by the relevant domain and event documentation.

---

# 16. Avoid Distributed Everything

RideForge will not introduce a separate network service merely because a component can technically be separated.

A service boundary should have a meaningful reason such as:

```text
Independent Scaling
Independent Deployment
Strong Domain Ownership
Failure Isolation
Security Boundary
Operational Isolation
Team Ownership
```

---

# 17. Service Extraction Principle

A module becomes a strong candidate for independent service extraction when one or more of the following become significant:

```text
Different Scaling Profile
Different Availability Requirement
Independent Deployment Requirement
Independent Data Ownership
Independent Failure Boundary
Independent Operational Lifecycle
```

---

# 18. Service Extraction Must Preserve Domain Boundaries

When extracting a service:

```text
Module
  ↓
Service Boundary
  ↓
API / Event Contract
  ↓
Independent Deployment
```

The extraction should not require unrelated domains to become tightly coupled.

---

# 19. Database Ownership

As services become independently deployable, database ownership should follow domain ownership.

A service should not depend on arbitrary writes into another service's private tables.

Preferred interaction:

```text
Service A
   ↓
API / Event
   ↓
Service B
```

rather than:

```text
Service A
   ↓
Directly Modify Service B Tables
```

---

# 20. Shared Database Consideration

During early development, a PostgreSQL instance may host multiple domain schemas or tables.

This does not automatically mean the architecture is tightly coupled.

The important distinction is:

```text
Physical Database Sharing
```

versus:

```text
Uncontrolled Data Ownership
```

Logical ownership should remain explicit.

---

# 21. Event-Driven Integration

RideForge will use event-driven integration where asynchronous communication provides clear value.

The event-driven architecture supports:

```text
Loose Coupling
Asynchronous Processing
Integration
Scalability
Failure Isolation
```

It should not be used for every operation.

---

# 22. Eventual Consistency

Some cross-domain workflows may use eventual consistency.

The architecture should explicitly identify where:

```text
Immediate Consistency
```

is required and where:

```text
Eventual Consistency
```

is acceptable.

---

# 23. Transaction Boundaries

A transaction should generally remain inside the domain/service boundary that owns the state.

Avoid distributed transactions unless there is a strong and documented reason.

Preferred pattern:

```text
Local Transaction
      ↓
Outbox
      ↓
Event
      ↓
Other Domain
```

---

# 24. Outbox Compatibility

RideForge's architecture is compatible with the Outbox pattern for reliable publication of events associated with local database transactions.

The detailed Outbox decision is documented separately.

---

# 25. Messaging Platform

RideForge uses an event-streaming platform based on:

```text
Kafka-Compatible Messaging
```

with the current production direction using:

```text
Redpanda
```

where appropriate.

The architectural style does not make every interaction asynchronous.

---

# 26. Real-Time State

Ride-hailing requires low-latency operational state.

Examples:

```text
Driver Availability
Driver Location
Active Dispatch State
Short-Lived Operational Data
```

These workloads may use Redis or other purpose-built mechanisms where appropriate.

The real-time state strategy remains separate from the core transactional source of truth.

---

# 27. PostgreSQL as Transactional Core

PostgreSQL remains the primary transactional database for core relational business state.

The architecture therefore separates:

```text
Transactional State
```

from:

```text
Real-Time / Ephemeral State
```

where the workload requires it.

---

# 28. Geospatial Architecture

RideForge may use PostgreSQL/PostGIS and other specialized location mechanisms according to workload requirements.

The architecture should not force all geospatial operations into a single storage technology.

The correct storage mechanism depends on:

```text
Query Pattern
Freshness
Scale
Latency
Durability
Cost
```

---

# 29. AI as an Optimization Layer

AI is treated as a specialized decision-support and optimization capability.

Conceptually:

```text
Authoritative Domain State
          ↓
AI / Optimization
          ↓
Prediction / Recommendation
          ↓
Domain Validation
          ↓
Action
```

AI does not automatically own:

```text
Legal Rules
Hard Constraints
Domain State
Transaction Integrity
```

---

# 30. Smart Dispatch and Stand Dispatch

RideForge supports both:

```text
Smart Dispatch
+
Stand Dispatch
```

as configurable operational strategies.

This allows the platform to select the appropriate dispatch approach according to:

```text
Operating Region
Business Model
Operational Requirements
Available Data
AI Availability
```

The dispatch strategy must remain compatible with the authoritative ride and driver domain rules.

---

# 31. AI Failure Must Not Become Platform Failure

AI components must have defined fallback behaviour.

Conceptually:

```text
AI Available
    ↓
Use AI

AI Unavailable
    ↓
Deterministic / Validated Fallback
```

The detailed failure strategy is documented separately.

---

# 32. Regional and Legal Constraints

Regional and legal rules remain authoritative.

The architecture must prevent:

```text
AI
Dispatch
Matching
Routing
```

from bypassing:

```text
Legal Validation
Regional Constraints
Operational Restrictions
```

This is a domain-level architectural constraint.

---

# 33. External Providers

External services should be accessed through explicit integration boundaries.

Examples:

```text
Routing Provider
Payment Provider
Messaging Provider
Maps / Geospatial Provider
AI Provider
```

Business logic should avoid being deeply coupled to a single provider where practical.

---

# 34. Provider Abstraction

A provider abstraction is justified when:

```text
Provider Replacement Is Plausible
Provider Failure Requires Fallback
Provider Differences Need Isolation
```

Do not create abstractions solely for theoretical future providers.

---

# 35. Failure Isolation

The architecture should isolate failures where practical.

For example:

```text
ETA Provider Failure
```

should not automatically stop:

```text
Ride Creation
```

Similarly:

```text
Demand Forecast Failure
```

should not automatically stop:

```text
Core Ride Lifecycle
```

---

# 36. Resilience Boundaries

Each critical component should have an identified:

```text
Failure Boundary
Fallback
Recovery Strategy
```

This is especially important for:

```text
Dispatch
Matching
ETA
Payments
Messaging
AI
```

---

# 37. Observability as a Cross-Cutting Capability

Every significant service or module should expose enough information to understand:

```text
Health
Latency
Errors
State Transitions
Dependencies
```

Observability is part of the architecture, not an afterthought.

---

# 38. Security Boundaries

Service boundaries should support:

```text
Authentication
Authorization
Secret Isolation
Network Controls
Data Access Controls
```

Security requirements should influence boundaries where necessary.

---

# 39. Configuration

Configuration should be separated from business logic.

The architecture should support:

```text
Environment Configuration
Feature Configuration
Model Configuration
Provider Configuration
Operational Configuration
```

without embedding environment-specific values in source code.

---

# 40. Deployment Architecture

The selected architecture must support independent deployment where justified.

A component should be deployable independently when:

```text
Its Boundary Is Stable
+
Its Dependencies Are Explicit
+
Its Operational Lifecycle Is Independent
```

---

# 41. Containerization

RideForge services should be container-compatible.

This supports:

```text
Local Development
CI/CD
Deployment
Scaling
Isolation
```

The architecture should remain usable without requiring a large orchestration platform for every local development task.

---

# 42. Local Development Principle

Local development should remain practical.

Developers should not need to reproduce the entire production infrastructure just to implement a small domain change.

Use lightweight local equivalents or subsets where appropriate.

---

# 43. Scalability Principle

RideForge should scale the component that requires scaling rather than scaling the entire platform unnecessarily.

For example:

```text
High Location Traffic
```

should not automatically require scaling:

```text
Payment Processing
```

---

# 44. Independent Scaling

Components with materially different workloads should be candidates for independent scaling.

Examples:

```text
Location
Matching
Dispatch
AI Inference
Event Consumers
```

---

# 45. Cost Principle

Architecture decisions should consider:

```text
Infrastructure Cost
Operational Cost
Engineering Cost
Scaling Cost
```

The most distributed architecture is not automatically the best architecture.

---

# 46. Avoid Premature Microservices

RideForge will not create services merely to increase the service count.

The desired architecture is:

```text
Small Enough to Understand
+
Separated Enough to Evolve
```

rather than:

```text
Maximum Number of Services
```

---

# 47. Coupling Principle

Minimize:

```text
Shared Internal State
Shared Database Writes
Synchronous Service Chains
Hidden Dependencies
```

Prefer:

```text
Explicit Contracts
Events
APIs
Domain Ownership
```

---

# 48. Cohesion Principle

A module or service should contain behaviour that changes together.

If two components:

```text
Always Change Together
Always Deploy Together
Always Scale Together
```

there may be little value in separating them prematurely.

---

# 49. Distributed Transaction Principle

Avoid distributed transactions where possible.

Prefer:

```text
Local Transaction
+
Reliable Event Publication
+
Idempotent Consumer
+
Eventual Consistency
```

where the business process allows it.

---

# 50. Synchronous Chain Principle

Avoid long synchronous chains such as:

```text
API
 ↓
Ride
 ↓
Matching
 ↓
Dispatch
 ↓
ETA
 ↓
Location
 ↓
External Provider
```

where every dependency must respond before the request can complete.

Long synchronous chains increase:

```text
Latency
Failure Propagation
Operational Complexity
```

---

# 51. Asynchronous Workflow Principle

Use asynchronous workflows when:

```text
Immediate Response Is Not Required
```

Examples:

```text
Notifications
Analytics
Non-Critical Processing
Learning Pipelines
Some Integration Workflows
```

---

# 52. Core Transaction Principle

Core transactional operations should remain explicit and deterministic.

For example:

```text
Ride Creation
Ride State Transition
Payment State
Driver Assignment State
```

must not depend on an opaque AI side effect.

---

# 53. Domain State Authority

For every important state, identify one authoritative owner.

Example:

```text
Ride State
→ Ride Domain

Driver Availability
→ Driver / Location Domain as defined by the architecture

Payment State
→ Payment Domain

Prediction
→ AI Capability
```

---

# 54. Event Ownership

Events should be published by the domain that owns the underlying fact.

For example:

```text
Ride Created
```

should originate from the owner of the ride lifecycle rather than an unrelated consumer.

---

# 55. Event Consumers

Consumers should treat events as integration contracts.

Consumers must not assume they control the producer's internal state.

---

# 56. Contract Stability

API and event contracts should be stable enough to allow independent evolution.

Breaking changes should be deliberate and versioned where required.

---

# 57. Backward Compatibility

When services evolve:

```text
Old Consumers
```

should not unexpectedly fail because of an internal implementation change.

Compatibility strategy should be chosen based on the contract and migration requirements.

---

# 58. Architecture Evolution

The architecture is intentionally evolutionary.

A typical progression may be:

```text
Well-Modularized Application
        ↓
Extract High-Value Boundary
        ↓
Independent Service
        ↓
Independent Scaling
        ↓
Further Domain Evolution
```

Not every component must follow this path.

---

# 59. Extraction Triggers

Consider extracting a component when:

```text
Scaling Pressure
+
Deployment Independence
+
Failure Isolation
+
Domain Ownership
```

justify the additional operational cost.

---

# 60. Do Not Extract for Vanity

The number of services is not an architecture quality metric.

The goal is:

```text
Business Capability Isolation
```

not:

```text
Service Count
```

---

# 61. Architecture Decision Hierarchy

RideForge architectural decisions should follow this hierarchy:

```text
Business / Legal Constraints
          ↓
Domain Rules
          ↓
Architecture
          ↓
Application Services
          ↓
Infrastructure
          ↓
Optimization
```

AI optimization operates within the constraints established above it.

---

# 62. Architecture Consistency

New components should be evaluated against:

```text
Existing Domain Boundaries
Existing Contracts
Existing Data Ownership
Existing Failure Model
Existing Observability
Existing Security Model
```

A technically attractive component should not be introduced if it creates unnecessary architectural inconsistency.

---

# 63. Technology Choice Principle

Technology should follow workload requirements.

The decision process should consider:

```text
Requirement
    ↓
Workload
    ↓
Constraints
    ↓
Options
    ↓
Trade-offs
    ↓
Technology
```

not:

```text
Technology
    ↓
Find a Problem for It
```

---

# 64. Architecture and Go

RideForge backend services are designed around Go where Go is the selected backend implementation technology.

The architecture should still prioritize:

```text
Domain Boundaries
Contracts
Testability
Operational Simplicity
```

rather than language-specific abstractions.

---

# 65. Architecture and Fiber

Where Fiber is used for HTTP services, HTTP framework concerns should remain at the interface layer.

Domain logic should not become dependent on framework-specific request handling.

---

# 66. Architecture and PostgreSQL

PostgreSQL is the primary transactional relational data store.

The architecture should use it for:

```text
Durable Business State
Transactional Operations
Relational Queries
```

while specialized workloads may use other infrastructure where justified.

---

# 67. Architecture and Redis

Redis is appropriate for:

```text
Low-Latency State
Caching
Ephemeral Data
Real-Time Workloads
```

but should not silently become the authoritative store for durable business state unless an explicit decision establishes that responsibility.

---

# 68. Architecture and Redpanda

Redpanda provides the event-streaming backbone for appropriate asynchronous workflows.

It should be used where:

```text
Event Streaming
Decoupling
Asynchronous Processing
```

provide meaningful value.

It should not replace ordinary synchronous APIs simply because an event platform exists.

---

# 69. Architecture and AI

AI capabilities remain modular and replaceable where practical.

The architecture should support:

```text
Model Replacement
Model Versioning
AI Disablement
Fallback
Controlled Rollout
```

without requiring a redesign of core ride lifecycle logic.

---

# 70. Architecture and Testing

Architectural boundaries should be testable.

At minimum:

```text
Domain Tests
Integration Tests
Contract Tests
Failure Tests
```

should be possible without requiring every test to deploy the entire platform.

---

# 71. Architecture and Observability

A service boundary is incomplete if its operational behaviour cannot be understood.

Important components should expose:

```text
Health
Metrics
Logs
Tracing
Dependency Information
```

where appropriate.

---

# 72. Architecture and Security

The architecture should follow:

```text
Least Privilege
Explicit Access
Minimal Exposure
Secure Defaults
```

Service boundaries should not be used as a substitute for proper authorization.

---

# 73. Architecture and Privacy

Data ownership should be explicit.

A service should access only the data it needs for its responsibility.

AI and analytics workloads should not receive unrestricted access to all operational data.

---

# 74. Architecture and Regional Operations

RideForge may operate in different regions with different operational and legal constraints.

Therefore architecture should support:

```text
Regional Configuration
Regional Dispatch Strategy
Regional Validation
Regional Provider Selection
```

where required.

---

# 75. Regional Strategy Selection

A region may use:

```text
Smart Dispatch
```

or:

```text
Stand Dispatch
```

according to configured operational requirements.

The architecture should allow this without duplicating the entire ride lifecycle system.

---

# 76. Architecture and Cross-Region Operations

Cross-region ride behaviour must pass through explicit validation.

Architecture should not assume that:

```text
Origin Region
+
Destination Region
```

automatically implies that the ride is operationally or legally allowed.

---

# 77. Architecture and Failure

Every critical dependency should have:

```text
Timeout
Failure Detection
Fallback / Degradation
Recovery
```

where appropriate.

---

# 78. Architecture and Performance

Performance optimization should happen at the appropriate layer.

Avoid prematurely introducing:

```text
Additional Services
Additional Caches
Additional Databases
Additional Queues
```

unless the measured workload justifies them.

---

# 79. Architecture and Cost

RideForge should prefer the simplest architecture that satisfies:

```text
Correctness
Reliability
Performance
Scalability
Security
Operational Requirements
```

while remaining economically viable.

---

# 80. Consequences

## 80.1 Positive Consequences

This decision provides:

```text
Clear Domain Boundaries
Future Service Extraction
Reduced Coupling
Incremental Scaling
Better Failure Isolation
Independent Evolution
Clear Ownership
```

It also allows the system to remain practical during early development.

---

## 80.2 Negative Consequences

The architecture requires discipline around:

```text
Module Boundaries
Contracts
Data Ownership
Event Design
Dependency Direction
```

Poor discipline can cause a modular architecture to degrade into a tightly coupled system.

---

## 80.3 Operational Consequences

As services are extracted, operational complexity increases:

```text
More Deployments
More Monitoring
More Network Calls
More Failure Modes
More Configuration
```

Therefore extraction must be justified by real operational needs.

---

# 81. Risks

## Risk 1 — Modular Monolith Becomes a Distributed Monolith

If services communicate through excessive synchronous calls and share data improperly, the system may become:

```text
Distributed Monolith
```

### Mitigation

Maintain:

```text
Clear Domain Ownership
Explicit Contracts
Limited Synchronous Chains
Event-Driven Integration
```

where appropriate.

---

## Risk 2 — Premature Service Extraction

Too many services too early can create unnecessary operational complexity.

### Mitigation

Use the extraction criteria defined in this ADR.

---

## Risk 3 — Shared Database Coupling

Multiple components may become dependent on each other's tables.

### Mitigation

Maintain:

```text
Logical Data Ownership
Repository Boundaries
Explicit Integration Contracts
```

and progressively enforce stronger ownership as services are extracted.

---

## Risk 4 — Event Overuse

Using events for every operation can make workflows difficult to understand.

### Mitigation

Use events when asynchronous decoupling provides meaningful value.

---

## Risk 5 — Excessive Synchronous Communication

Long synchronous chains can amplify failures.

### Mitigation

Use:

```text
Timeouts
Fallbacks
Asynchronous Events
```

where appropriate.

---

## Risk 6 — AI Coupling

AI may gradually become required for core domain operations.

### Mitigation

Maintain:

```text
Deterministic Domain Rules
AI as Optimization
Explicit Fallbacks
```

---

# 82. Implementation Guidelines

The architecture should be implemented with the following structure:

```text
Domain
├── Entities
├── Value Objects
├── Domain Services
├── Domain Events
└── Repository Interfaces

Application
├── Use Cases
├── Application Services
├── DTOs
└── Orchestration

Infrastructure
├── PostgreSQL
├── Redis
├── Redpanda
├── External Providers
└── Other Adapters

Interfaces
├── HTTP
├── Events
└── External Entry Points
```

The exact package structure remains governed by the relevant development documentation.

---

# 83. Architecture Evolution Strategy

RideForge should evolve through measured extraction:

```text
Stage 1
Well-Modularized Application
        ↓
Stage 2
Stable Domain Boundaries
        ↓
Stage 3
Identify High-Pressure Boundary
        ↓
Stage 4
Extract Service
        ↓
Stage 5
Independent Scaling
        ↓
Stage 6
Independent Evolution
```

---

# 84. When Not to Extract a Service

Do not extract a service when:

```text
Traffic Is Low
Scaling Profile Is Identical
Deployment Is Always Coupled
Data Is Tightly Coupled
Operational Benefit Is Minimal
```

unless another strong architectural reason exists.

---

# 85. When to Extract a Service

Extraction becomes more attractive when:

```text
Traffic Is High
Scaling Is Independent
Failure Isolation Matters
Deployment Independence Matters
Domain Ownership Is Strong
Data Ownership Is Clear
Operational Boundary Is Valuable
```

---

# 86. Validation

The architecture decision should be validated continuously through:

```text
Code Structure
Dependency Analysis
Integration Tests
Performance Testing
Operational Metrics
Failure Testing
Production Experience
```

The architecture is successful if the system can evolve without excessive coupling or unnecessary operational complexity.

---

# 87. Architecture Review Triggers

Revisit this ADR when:

```text
Major Traffic Growth
Major Service Extraction
Major Database Change
Messaging Platform Change
New Deployment Model
Major AI Integration
Major Regional Expansion
Significant Operational Failure
```

occurs.

---

# 88. Related Documentation

The following documentation defines the implementation details that follow this architectural decision:

```text
03-components/
04-development/
05-ai/
```

Relevant areas include:

```text
Components and Services
Development Guidelines
Code Structure
Database Development
Redis Development
Event and Messaging Development
API Development
Observability
Performance
AI Architecture
AI Failure and Fallback
```

---

# 89. Related ADRs

This ADR establishes the architectural foundation for subsequent decisions.

Planned related ADRs include:

```text
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
ADR-0005 — Event-Driven Architecture
ADR-0012 — Outbox Pattern
ADR-0014 — API and Service Communication
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0021 — Failure and Degradation Strategy
ADR-0027 — Cloud and Deployment Strategy
ADR-0029 — Architecture Evolution and Migration
```

These ADRs should refine or specialize this decision rather than contradict it without explicitly superseding it.

---

# 90. Decision Summary

RideForge adopts:

```text
Domain-Oriented
+
Modular
+
Microservice-Compatible
+
Event-Driven Where Appropriate
```

architecture.

The platform will not pursue:

```text
Monolith by Default
```

or:

```text
Microservices Everywhere
```

Instead:

```text
Build Clear Boundaries First
        ↓
Measure Real Requirements
        ↓
Extract Where Justified
        ↓
Scale Independently Where Valuable
```

---

# 91. Final Principle

> **RideForge should be designed as a system that can become more distributed without requiring it to be distributed prematurely.**

The architectural objective is:

```text
Strong Boundaries
+
Low Coupling
+
Clear Ownership
+
Independent Evolution
+
Controlled Complexity
```

The architecture should allow RideForge to grow from:

```text
A practical production system
```

into:

```text
A scalable distributed platform
```

without sacrificing:

```text
Domain Integrity
Operational Reliability
Developer Productivity
Cost Discipline
```

---

# 92. Status

```text
Decision: ACCEPTED

Architecture Style:
Domain-Oriented Modular Architecture
+
Microservice-Compatible Boundaries
+
Event-Driven Integration Where Appropriate
```

This decision is the architectural baseline for the subsequent RideForge ADRs.
