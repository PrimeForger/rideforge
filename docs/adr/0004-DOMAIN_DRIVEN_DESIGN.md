# ADR-0004: Domain-Driven Design

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Architecture  
> **Scope:** RideForge domain and application architecture  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is a ride-hailing platform whose core behaviour is driven by business concepts rather than by infrastructure alone.

Important concepts include:

```text
Ride
Driver
Location
Matching
Dispatch
ETA
Payment
Notification
AI / Optimization
Region
```

These concepts have different:

```text
Responsibilities
Rules
State
Lifecycles
Consistency Requirements
Failure Characteristics
```

The architecture therefore needs to keep business rules understandable and protected from unnecessary infrastructure coupling.

ADR-0002 established a domain-oriented, modular, microservice-compatible architecture.

ADR-0003 established that service boundaries should follow meaningful business capability, domain ownership, data ownership, scaling requirements, and failure boundaries.

Domain-Driven Design provides the architectural discipline for defining and protecting those domain boundaries.

---

# 2. Problem

Without explicit domain boundaries, business logic can gradually become distributed across:

```text
HTTP Handlers
Application Services
Repositories
Database Code
Redis Code
Kafka Consumers
External Providers
AI Services
```

This creates several risks:

```text
Business Logic Duplication
Hidden Coupling
Unclear Ownership
Infrastructure-Driven Design
Difficult Testing
Unsafe Service Extraction
```

RideForge therefore needs a consistent domain model and dependency structure.

---

# 3. Decision

RideForge will use **Domain-Driven Design (DDD) principles** as the primary approach for organizing business logic.

The architecture will distinguish between:

```text
Domain
Application
Infrastructure
Interfaces
```

with the domain remaining the primary owner of business rules.

DDD will be applied pragmatically.

RideForge will use DDD where it provides meaningful value and will avoid introducing domain patterns merely for ceremony.

---

# 4. Core DDD Principles

RideForge will prioritize:

```text
Explicit Domain Concepts
Clear Business Rules
Bounded Responsibilities
Domain Ownership
High Cohesion
Low Coupling
Explicit State Transitions
Domain Events Where Appropriate
```

---

# 5. Domain Model

The domain model should represent concepts that matter to the ride-hailing business.

Examples include:

```text
Ride
Driver
Vehicle
Location
Dispatch Strategy
Match
Assignment
ETA
Payment
Region
```

The exact model evolves as the product evolves.

---

# 6. Ubiquitous Language

RideForge should use consistent terminology across:

```text
Code
Documentation
APIs
Events
Database Concepts
Architecture Discussions
```

The same business concept should not receive unrelated names across different parts of the platform unless there is a deliberate domain distinction.

---

# 7. Why Ubiquitous Language Matters

Consistent language reduces ambiguity.

For example, the system should clearly distinguish concepts such as:

```text
Candidate
Match
Assignment
Dispatch
Ride
Driver Availability
Driver Location
```

These terms may be related but should not automatically be treated as interchangeable.

---

# 8. Bounded Contexts

RideForge will use bounded contexts to separate areas of the domain that have:

```text
Different Models
Different Responsibilities
Different Rules
Different Ownership
```

A bounded context may eventually correspond to:

```text
A Module
```

or:

```text
An Independent Service
```

but the two concepts are not automatically identical.

---

# 9. Initial Domain Areas

The initial conceptual domain areas include:

```text
Ride
Driver
Location
Matching
Dispatch
ETA
Payment
Notification
AI / Optimization
```

These boundaries may evolve as the platform grows.

---

# 10. Ride Context

The Ride context represents the lifecycle of a ride.

Its model should capture concepts such as:

```text
Ride Identity
Pickup
Destination
Ride Status
Ride Lifecycle
Cancellation
Completion
```

The Ride context is authoritative for ride lifecycle state.

---

# 11. Driver Context

The Driver context represents driver-related business state.

Potential concepts include:

```text
Driver Identity
Driver Eligibility
Driver Operational Status
Driver Availability
Driver Profile
```

The exact ownership of real-time location is separated from ordinary driver profile data when required by workload characteristics.

---

# 12. Location Context

Location is a specialized domain capability because it involves:

```text
High-Frequency Updates
Freshness
Geospatial Queries
Low-Latency Access
Ephemeral State
```

The Location context may therefore use infrastructure different from ordinary transactional driver data.

---

# 13. Matching Context

The Matching context is responsible for determining suitable driver candidates for a ride.

Conceptually:

```text
Ride Request
      ↓
Candidate Retrieval
      ↓
Eligibility
      ↓
Ranking
      ↓
Match
```

The exact implementation may evolve.

---

# 14. Dispatch Context

The Dispatch context coordinates the strategy used to move a ride toward assignment.

RideForge supports:

```text
Stand Dispatch
+
Smart Dispatch
```

The domain model should therefore represent dispatch as a business capability with interchangeable strategies where appropriate.

---

# 15. ETA Context

ETA is a specialized capability responsible for estimating travel or arrival time.

It may use:

```text
Routing
Traffic
Location
Historical Data
Machine Learning
External Providers
```

ETA output should remain a prediction or estimate rather than silently becoming authoritative domain state.

---

# 16. Payment Context

Payment is a separate domain because it has:

```text
Financial State
Provider Integrations
Consistency Requirements
Failure Handling
Security Requirements
```

Payment ownership should remain explicit.

---

# 17. Notification Context

Notification is responsible for communicating events or state changes to users and other operational actors.

It should generally remain asynchronous where immediate completion is not required.

---

# 18. AI / Optimization Context

AI capabilities support:

```text
Prediction
Ranking
Optimization
Forecasting
Recommendation
```

AI should not automatically become the owner of the business state it helps optimize.

---

# 19. AI and Domain Authority

The preferred model is:

```text
Domain State
      ↓
AI Input
      ↓
AI Prediction / Recommendation
      ↓
Domain Validation
      ↓
Business Action
```

This preserves deterministic business authority.

---

# 20. Domain Entities

Entities represent domain concepts whose identity persists across state changes.

Examples may include:

```text
Ride
Driver
Payment
```

An entity is identified by its domain identity rather than solely by its current attribute values.

---

# 21. Entity Identity

Entity identity should be:

```text
Stable
Explicit
Domain Meaningful
```

Changing an entity's attributes should not automatically create a new entity.

---

# 22. Value Objects

Value objects represent concepts defined by their values rather than independent identity.

Examples may include:

```text
Coordinates
Money
Address
Region
Distance
Duration
```

Value objects should be used where they make domain rules clearer and safer.

---

# 23. Value Object Characteristics

A value object should generally be:

```text
Immutable
Validated
Meaningful
Self-Consistent
```

Where practical.

---

# 24. Aggregates

Aggregates define consistency boundaries within the domain.

An aggregate should protect business invariants that must remain consistent together.

The aggregate should have a clear:

```text
Aggregate Root
```

through which state changes are controlled.

---

# 25. Aggregate Design Principle

Do not create large aggregates merely because multiple entities are related.

Prefer:

```text
Small Consistency Boundaries
```

when business rules permit.

---

# 26. Aggregate Example

Conceptually:

```text
Ride Aggregate
      │
      ├── Ride Identity
      ├── Ride State
      └── Ride Lifecycle Rules
```

The exact internal aggregate model should follow the actual domain requirements.

---

# 27. Aggregate Boundaries and Transactions

Transactions should generally remain within an aggregate or an appropriate domain transaction boundary.

Avoid requiring a single transaction to span unrelated services.

---

# 28. Domain Services

A domain service may be used when business logic:

```text
Is Domain Logic
```

but:

```text
Does Not Naturally Belong to One Entity or Value Object
```

Examples may include complex domain operations involving multiple domain concepts.

---

# 29. Domain Service Restrictions

Domain services should not become a generic dumping ground for business logic.

Avoid creating:

```text
God Domain Service
```

containing unrelated operations.

---

# 30. Application Services

Application services orchestrate use cases.

They may:

```text
Receive a Request
Load Domain State
Invoke Domain Behaviour
Persist Changes
Publish Events
Return a Result
```

They should not become the primary location for domain rules that belong in the domain model.

---

# 31. Domain vs Application Responsibility

A useful distinction is:

```text
Domain
→ What the business allows

Application
→ How a use case is orchestrated
```

---

# 32. Example

A domain rule might be:

```text
A ride cannot transition from completed back to active.
```

The application service may orchestrate:

```text
Load Ride
    ↓
Request Transition
    ↓
Persist Ride
    ↓
Publish Event
```

The validity of the transition belongs to the domain.

---

# 33. Repository Abstractions

Repositories represent domain-facing persistence requirements.

The domain/application layer should depend on an appropriate abstraction rather than directly depending on:

```text
SQL Driver
Redis Client
ORM Implementation
```

where practical.

---

# 34. Repository Responsibility

Repositories should primarily handle:

```text
Persistence
Retrieval
Querying Required State
```

They should not become a replacement for domain services.

---

# 35. Repository Boundaries

Repository interfaces should be designed around domain needs rather than exposing the entire database.

Avoid generic abstractions such as:

```text
SaveEverything()
GetAnything()
```

when they hide domain ownership.

---

# 36. Domain Events

Domain events represent meaningful facts that occurred in the domain.

Examples may include:

```text
RideCreated
RideCancelled
RideCompleted
DriverStatusChanged
MatchCreated
AssignmentCompleted
```

Only meaningful events should be introduced.

---

# 37. Domain Event Principle

A domain event should communicate:

> **Something meaningful happened.**

It should not merely expose:

```text
A Database Row Changed
```

unless that change itself is a meaningful domain fact.

---

# 38. Event Ownership

The domain that owns the fact should be responsible for producing the corresponding domain event.

Example:

```text
Ride Context
    ↓
Ride Created
```

---

# 39. Domain Events vs Integration Events

The concepts should remain distinguishable.

```text
Domain Event
→ Internal domain fact

Integration Event
→ Contract intended for communication across boundaries
```

They may share implementation mechanisms, but their semantic purpose is different.

---

# 40. Domain Invariants

Critical business rules should be represented explicitly.

Examples may include:

```text
Invalid Ride State Transition
Invalid Driver Assignment
Illegal Region Transition
Invalid Payment State
Invalid Dispatch Action
```

The exact invariant set is domain-specific.

---

# 41. Invariant Ownership

An invariant should be enforced by the boundary that owns the state.

For example:

```text
Ride Lifecycle Rule
→ Ride Domain
```

rather than allowing every API handler to implement a different version of the rule.

---

# 42. Domain Validation

Validation should be separated into:

```text
Input Validation
Domain Validation
Infrastructure Validation
```

These have different responsibilities.

---

# 43. Input Validation

Input validation checks whether external input is structurally acceptable.

Examples:

```text
Required Fields
Format
Type
Basic Range
```

---

# 44. Domain Validation

Domain validation determines whether the operation is valid according to business rules.

Examples:

```text
Can this ride transition to this state?
Can this driver be assigned?
Is this operation allowed in this region?
```

---

# 45. Infrastructure Validation

Infrastructure validation handles requirements such as:

```text
Database Availability
External Provider Response
Message Broker Availability
```

These are not domain rules.

---

# 46. Regional Rules

Regional and legal restrictions are domain-level constraints.

They must remain authoritative over optimization mechanisms.

Conceptually:

```text
Region Rule
      ↓
Eligibility
      ↓
Matching
      ↓
Dispatch
```

not:

```text
AI Ranking
      ↓
Override Region Rule
```

---

# 47. Dispatch Strategy as Domain Policy

Dispatch strategy can be represented as a domain policy or strategy where appropriate.

For example:

```text
Stand Dispatch Strategy
Smart Dispatch Strategy
```

The selection of a strategy should respect region and operational configuration.

---

# 48. Strategy Pattern

Where multiple business strategies exist, the implementation may use a strategy abstraction.

Conceptually:

```text
Dispatch Strategy
       │
       ├── Stand Dispatch
       │
       └── Smart Dispatch
```

This is an implementation technique, not a requirement to over-abstract the domain.

---

# 49. Domain Policy

A policy represents a business rule or decision criterion that may apply across domain objects.

Examples may include:

```text
Driver Eligibility Policy
Dispatch Eligibility Policy
Regional Ride Policy
```

Policies should remain explicit and testable.

---

# 50. Domain Fact vs Optimization

RideForge must distinguish between:

```text
Fact
```

and:

```text
Prediction
```

For example:

```text
Driver is currently available
```

is operational state.

Whereas:

```text
Driver is likely to accept this ride
```

is a prediction.

The latter must not silently replace the former.

---

# 51. Domain Fact vs Recommendation

Similarly:

```text
Driver X is eligible
```

is a domain determination.

Whereas:

```text
Driver X is recommended
```

may be an optimization output.

The architecture should preserve that distinction.

---

# 52. External Services in the Domain

External providers should not leak their models throughout the domain.

For example:

```text
Google Maps
Mapbox
Payment Provider
AI Provider
```

should be represented through appropriate integration boundaries.

---

# 53. Anti-Corruption Layer

Where an external system has a significantly different model, an anti-corruption layer may be used to prevent the external model from contaminating the RideForge domain model.

Conceptually:

```text
RideForge Domain
      ↓
Adapter / ACL
      ↓
External Provider
```

Use this where the complexity is justified.

---

# 54. Domain Model and PostgreSQL

PostgreSQL is infrastructure.

The domain should not depend on PostgreSQL-specific concepts unless those concepts are genuinely part of the required domain behaviour.

Database implementation remains an infrastructure concern.

---

# 55. Domain Model and Redis

Redis is infrastructure for workloads such as:

```text
Caching
Real-Time State
Ephemeral Data
```

The domain should express required behaviour rather than directly depending on Redis operations.

---

# 56. Domain Model and Redpanda

The domain may produce meaningful events.

The infrastructure layer is responsible for delivering those events through the selected messaging platform.

The domain should not contain broker-specific implementation details.

---

# 57. Domain Model and HTTP

HTTP is an interface mechanism.

Domain logic should not depend directly on:

```text
HTTP Request
HTTP Response
Fiber Context
```

unless an explicit boundary requires it.

---

# 58. Dependency Direction

The preferred dependency direction is:

```text
Interfaces
     ↓
Application
     ↓
Domain
```

while infrastructure implements required interfaces:

```text
Infrastructure
     ↓
Application / Domain Contracts
```

A more complete conceptual model is:

```text
              Interfaces
                  │
                  ▼
             Application
                  │
                  ▼
                Domain
                  ▲
                  │
            Infrastructure
```

---

# 59. Dependency Inversion

High-level business logic should not depend directly on low-level infrastructure implementation details.

Instead:

```text
Business Requirement
       ↓
Interface / Contract
       ↓
Infrastructure Implementation
```

---

# 60. Domain Purity

Domain purity should be applied pragmatically.

The goal is not to create a completely technology-free codebase at any cost.

The goal is to prevent infrastructure concerns from controlling business rules.

---

# 61. Avoid Anemic Domain Models

Where meaningful domain behaviour exists, avoid reducing entities to simple data containers while placing all business rules into unrelated services.

Prefer:

```text
Data
+
Relevant Behaviour
+
Invariants
```

within the appropriate domain boundary.

---

# 62. Avoid God Objects

No single domain entity or service should become responsible for unrelated parts of the platform.

Examples of undesirable ownership:

```text
Ride
→ Payment
→ Notifications
→ AI
→ Driver Location
→ Analytics
```

Such behaviour should be separated by domain responsibility.

---

# 63. Avoid God Services

Similarly, avoid:

```text
RideManager
PlatformService
BusinessService
CommonService
```

that contains unrelated business logic.

---

# 64. Domain Coupling

Domains should communicate through:

```text
Explicit Contracts
Domain Events
Integration Events
APIs
```

rather than direct manipulation of internal objects owned by another domain.

---

# 65. Cross-Domain Transactions

Cross-domain operations should not automatically require a single database transaction.

Prefer:

```text
Local Transaction
      ↓
Event
      ↓
Next Domain
```

where business requirements permit eventual consistency.

---

# 66. Consistency Boundaries

The domain model should explicitly identify which state must be consistent immediately.

Examples may include:

```text
Ride State Transition
Assignment State
Payment State
```

Other information may tolerate eventual consistency.

---

# 67. Domain State Transitions

State machines should be explicit for lifecycle-heavy concepts.

For example:

```text
Requested
   ↓
Searching
   ↓
Matched
   ↓
Driver Assigned
   ↓
Driver Arriving
   ↓
Trip Started
   ↓
Completed
```

The exact RideForge state model is defined by the actual ride lifecycle implementation.

---

# 68. Invalid Transitions

Invalid transitions should be rejected by the domain rather than relying only on API-level checks.

For example:

```text
Completed
   ↓
Requested
```

should not be possible if the domain does not permit that transition.

---

# 69. Domain Commands

Where useful, application operations can be represented conceptually as commands:

```text
CreateRide
CancelRide
AssignDriver
StartRide
CompleteRide
```

Commands represent requested actions.

---

# 70. Domain Events

Events represent completed facts:

```text
RideCreated
RideCancelled
DriverAssigned
RideStarted
RideCompleted
```

The distinction is:

```text
Command
→ Request

Event
→ Fact
```

---

# 71. Domain Queries

Queries retrieve information without changing domain state.

Examples:

```text
GetRide
GetDriver
GetCurrentAssignment
GetETA
```

Queries should not introduce hidden state mutations.

---

# 72. CQRS

RideForge does not require full CQRS everywhere.

CQRS concepts may be used selectively when:

```text
Read and Write Workloads Differ Significantly
```

or:

```text
A Dedicated Read Model Provides Clear Value
```

Avoid adopting full CQRS merely for architectural fashion.

---

# 73. Event Sourcing

RideForge does not adopt Event Sourcing as a default domain persistence strategy.

Event-driven architecture does not imply Event Sourcing.

If Event Sourcing becomes necessary for a specific domain, it requires a separate architectural decision.

---

# 74. Domain Model Versioning

When a domain model changes:

```text
Existing State
Existing Events
Existing APIs
```

must be considered.

Breaking domain changes should be deliberate and migration-aware.

---

# 75. Domain Ownership and Database Migration

Schema migrations should respect domain ownership.

A migration should not silently change another domain's data assumptions without coordinated review.

---

# 76. Domain Ownership and Events

If an event changes meaning, the event contract should be reviewed independently of the internal domain model.

Internal refactoring should not accidentally become an external contract change.

---

# 77. Testing Domain Logic

Domain rules should be highly testable without requiring:

```text
PostgreSQL
Redis
Redpanda
HTTP
External Providers
```

where those dependencies are not required for the rule itself.

---

# 78. Domain Testing

Important tests include:

```text
Entity Behaviour
Value Object Validation
State Transitions
Invariants
Domain Policies
Domain Services
```

---

# 79. Application Testing

Application services should be tested for:

```text
Use Case Orchestration
Repository Interaction
Event Publication
Error Handling
Transaction Behaviour
```

---

# 80. Infrastructure Testing

Infrastructure should be tested separately for:

```text
Database Behaviour
Redis Behaviour
Messaging Behaviour
External Provider Behaviour
```

---

# 81. Domain and Observability

Domain events and state transitions should provide enough information for operational observability without exposing unnecessary sensitive data.

---

# 82. Domain and Logging

Logs should describe useful domain context such as:

```text
Ride ID
Driver ID
Assignment ID
Event Type
State Transition
Correlation ID
```

subject to privacy and security requirements.

---

# 83. Domain and Security

Authorization should be enforced at appropriate application/domain boundaries.

A domain entity should not assume that every caller is trusted merely because the entity exists internally.

---

# 84. Domain and Privacy

Domain models should avoid carrying unnecessary personal or sensitive information.

Use the minimum data required for the domain responsibility.

---

# 85. Domain and Performance

DDD should not justify unnecessary object graphs, database calls, or abstractions.

Performance-sensitive workloads such as:

```text
Location
Matching
Dispatch
ETA
```

may require specialized implementation techniques while preserving domain boundaries.

---

# 86. Domain and Scalability

The domain model should make high-scale workloads identifiable.

For example:

```text
Location Updates
```

may require different scaling than:

```text
Ride Lifecycle
```

The domain boundary helps make this distinction explicit.

---

# 87. Domain and Service Extraction

A well-defined bounded context should make service extraction easier.

Conceptually:

```text
Bounded Context
      ↓
Stable Contract
      ↓
Independent Deployment
```

This connects DDD directly with the microservice boundary decision.

---

# 88. Domain Evolution

The domain model should evolve with the business.

Changes should be reviewed when:

```text
Business Rules Change
New Region Added
Dispatch Strategy Changes
New Payment Model Added
New Ride Type Added
AI Capability Changes
```

---

# 89. Domain Model Review

Review domain boundaries when:

```text
A Domain Has Too Many Responsibilities
Two Domains Are Tightly Coupled
Business Rules Are Duplicated
Data Ownership Is Unclear
State Transitions Are Scattered
```

These are signals that the model may require refinement.

---

# 90. Consequences

## 90.1 Positive Consequences

The decision provides:

```text
Clear Business Ownership
Better Business Rule Isolation
Improved Testability
Reduced Infrastructure Coupling
Better Service Boundaries
Safer Evolution
Clearer Terminology
```

---

## 90.2 Negative Consequences

DDD introduces additional discipline:

```text
More Explicit Concepts
More Interfaces
More Deliberate Boundaries
More Domain Modeling
```

It may require more upfront thought than a purely CRUD-oriented implementation.

---

# 91. Risks

## Risk 1 — Over-Engineering

DDD can become ceremony if every concept is modeled excessively.

### Mitigation

Use DDD pragmatically and only introduce patterns that provide real value.

---

## Risk 2 — Anemic Domain

Business rules may still leak into application services.

### Mitigation

Review ownership of invariants and domain behaviour.

---

## Risk 3 — Incorrect Bounded Contexts

A context may be split incorrectly.

### Mitigation

Review boundaries using business capability, ownership, consistency, and change characteristics.

---

## Risk 4 — Shared Models

Shared domain models can create strong coupling between contexts.

### Mitigation

Prefer explicit contracts and context-specific models where necessary.

---

## Risk 5 — Infrastructure Leakage

Database or framework concerns may enter domain logic.

### Mitigation

Maintain dependency direction and infrastructure adapters.

---

# 92. Validation

DDD adoption should be evaluated through:

```text
Code Structure
Dependency Analysis
Domain Tests
Integration Tests
Service Boundaries
Data Ownership
Production Evolution
```

A successful DDD implementation should make business rules easier to locate and reason about.

---

# 93. Review Triggers

Revisit this ADR when:

```text
Major Domain Added
Bounded Context Changed
Major Service Extraction
Major Business Rule Change
Domain Model Becomes Difficult to Maintain
Cross-Domain Coupling Increases
```

---

# 94. Related Documentation

This ADR should be read with:

```text
03-components/
04-development/
05-ai/
```

Relevant areas include:

```text
Code Structure and Conventions
Go Development Standards
Database Development
Event and Messaging Development
API Development
Error Handling and Validation
Testing
Observability
AI Architecture
AI Matching and Ranking
AI Failure and Fallback
```

---

# 95. Related ADRs

This decision is directly related to:

```text
ADR-0002 — Architecture Style
ADR-0003 — Microservice Boundaries
ADR-0005 — Event-Driven Architecture
ADR-0012 — Outbox Pattern
ADR-0014 — API and Service Communication
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0029 — Architecture Evolution and Migration
```

---

# 96. Decision Summary

RideForge adopts Domain-Driven Design as the architectural approach for organizing business logic.

The implementation will emphasize:

```text
Ubiquitous Language
+
Bounded Contexts
+
Entities
+
Value Objects
+
Aggregates
+
Domain Services
+
Domain Events
+
Explicit Invariants
+
Clear Ownership
```

while avoiding unnecessary ceremony.

---

# 97. Final Principle

> **Business rules belong to the domain that owns them. Infrastructure should support the domain, not define it.**

The desired dependency model is:

```text
Business Concepts
       ↓
Domain Rules
       ↓
Application Use Cases
       ↓
Infrastructure / Interfaces
```

and not:

```text
Database
       ↓
Application
       ↓
Business Rules
```

RideForge will therefore use DDD as a practical tool for maintaining:

```text
Business Clarity
+
Domain Integrity
+
Service Boundaries
+
Testability
+
Long-Term Evolution
```

---

# 98. Status

```text
Decision: ACCEPTED

Domain Architecture:
Domain-Driven Design
+
Bounded Contexts
+
Explicit Domain Ownership
+
Pragmatic DDD
```

This decision establishes the domain modeling foundation for subsequent RideForge architectural decisions.
