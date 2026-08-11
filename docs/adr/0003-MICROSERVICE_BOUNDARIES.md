# ADR-0003: Microservice Boundaries

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Architecture  
> **Scope:** RideForge service and domain boundaries  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is designed as a domain-oriented, modular, microservice-compatible platform.

The architecture established by ADR-0002 deliberately avoids both extremes:

```text
Monolith Everything
```

and:

```text
Microservice Everything
```

RideForge instead requires clear domain boundaries that can evolve into independently deployable services when operational or business requirements justify the separation.

A ride-hailing platform naturally contains multiple capabilities:

```text
Users
Drivers
Rides
Dispatch
Matching
Location
ETA
Payments
Notifications
AI / Optimization
```

These capabilities do not necessarily have identical:

```text
Scaling Requirements
Availability Requirements
Data Ownership
Latency Requirements
Deployment Frequency
Failure Characteristics
Security Requirements
Operational Responsibilities
```

Therefore, service boundaries must be based primarily on **business/domain ownership and operational characteristics**, rather than arbitrary technical decomposition.

---

# 2. Problem

RideForge needs a consistent method for deciding:

```text
What belongs inside one service?
What should remain a module?
What should become an independent service?
Who owns the data?
How should services communicate?
When should a service be extracted?
```

Without explicit boundary rules, the system can evolve into either:

```text
A Large Coupled Application
```

or:

```text
A Distributed Monolith
```

Both create long-term engineering problems.

---

# 3. Decision

RideForge will define service boundaries primarily around:

```text
Business Capability
+
Domain Ownership
+
Data Ownership
+
Independent Scaling
+
Failure Isolation
+
Deployment Independence
```

A component should become an independently deployable microservice only when the benefits of separation justify the additional distributed-system complexity.

---

# 4. Boundary Hierarchy

RideForge boundaries should be evaluated in the following order:

```text
Business Capability
        ↓
Domain Boundary
        ↓
Data Ownership
        ↓
Application Boundary
        ↓
Infrastructure Boundary
        ↓
Deployment Boundary
```

A deployment boundary should normally follow a meaningful domain boundary rather than create one artificially.

---

# 5. Primary Boundary Principle

> **A service owns a coherent business capability and the state required to operate that capability.**

A service should not exist merely because:

```text
The code is large
A package is large
A technology is different
A framework is different
A developer wants a separate repository
```

---

# 6. Candidate RideForge Domains

The initial architectural domains include:

```text
User
Driver
Ride
Dispatch
Matching
Location
ETA
Payment
Notification
AI / Optimization
```

These are conceptual domain boundaries.

They are not all required to become independent services immediately.

---

# 7. Ride Domain

The Ride domain owns the lifecycle of a ride.

Typical responsibilities include:

```text
Ride Creation
Ride State
Ride Lifecycle
Pickup / Destination Information
Ride Status
Ride Cancellation
Ride Completion
```

The exact implementation responsibilities are defined by the application and component documentation.

---

# 8. Ride Domain Ownership

The Ride domain should be authoritative for:

```text
Ride Identity
Ride Lifecycle State
Ride State Transitions
```

Other services may consume ride information but should not directly modify Ride-owned state without an explicit contract.

---

# 9. Driver Domain

The Driver domain represents driver-related business state.

Potential responsibilities include:

```text
Driver Identity
Driver Status
Driver Eligibility
Driver Operational State
Driver Profile Information
```

Real-time driver location may require a separate specialized location capability depending on scale and workload.

---

# 10. Driver Data Ownership

Driver-related data should have a clear owner.

Other domains should not directly write into driver-owned tables simply because they need driver information.

Preferred interaction:

```text
Driver Domain
     ↓
API / Event
     ↓
Consumer
```

---

# 11. Location Domain

Location is a specialized workload because ride-hailing requires:

```text
High Write Frequency
Low Latency
Freshness
Geospatial Queries
Short-Lived State
```

The Location capability may therefore have different scaling and infrastructure requirements from ordinary driver profile data.

---

# 12. Location Boundary

The architecture permits Location to evolve into an independently scalable service when:

```text
Location Traffic
```

creates materially different requirements from:

```text
Driver Business Data
```

The physical storage strategy is governed separately by the location-storage ADR.

---

# 13. Matching Domain

Matching determines suitable driver candidates for a ride.

Conceptually:

```text
Ride Request
      ↓
Eligible Driver Candidates
      ↓
Ranking / Selection
```

Matching should remain distinct from the broader ride lifecycle.

---

# 14. Matching Responsibilities

Matching may include:

```text
Candidate Retrieval
Eligibility Filtering
Candidate Ranking
Match Selection
```

Hard business constraints must remain authoritative.

AI ranking may optimize candidate ordering but must not bypass mandatory eligibility rules.

---

# 15. Dispatch Domain

Dispatch is responsible for coordinating how a ride request is offered or assigned according to the configured dispatch strategy.

RideForge supports:

```text
Stand Dispatch
+
Smart Dispatch
```

The dispatch boundary should therefore support multiple strategies without duplicating the entire Ride lifecycle.

---

# 16. Dispatch and Matching Relationship

Matching and Dispatch are related but conceptually different.

A simplified relationship is:

```text
Ride
 ↓
Dispatch Strategy
 ↓
Candidate Selection / Matching
 ↓
Assignment
```

The exact orchestration may evolve.

The important architectural principle is that:

```text
Candidate Selection
```

and:

```text
Ride Lifecycle
```

should not become one inseparable implementation.

---

# 17. Stand Dispatch

Stand Dispatch represents a deterministic operational strategy appropriate to locations or business environments where dispatch follows stand-based rules.

The architecture should allow:

```text
Stand Dispatch
```

to operate without requiring:

```text
AI
```

to be available.

---

# 18. Smart Dispatch

Smart Dispatch represents an optimization-oriented strategy that can use:

```text
Driver Availability
Location
ETA
Demand
Supply
Historical Signals
AI / Ranking
```

where appropriate.

Smart Dispatch remains constrained by:

```text
Domain Rules
Eligibility
Regional Rules
Legal Constraints
Operational Constraints
```

---

# 19. ETA Domain

ETA is a specialized capability because it may depend on:

```text
Location
Routing
Traffic
Historical Data
Prediction Models
External Providers
```

It can therefore have a distinct operational profile.

ETA should expose an explicit contract rather than leaking routing-provider implementation details into unrelated domains.

---

# 20. Payment Domain

Payment is a high-integrity domain.

Payment state should have a clearly defined owner.

Other domains should not directly manipulate payment records.

Typical interaction:

```text
Ride / Order Event
       ↓
Payment Service
       ↓
Payment Provider
       ↓
Payment Result
       ↓
Payment Event
```

The exact payment workflow is governed by payment-specific architecture documentation and future ADRs.

---

# 21. Notification Domain

Notifications are generally asynchronous and can have an independent operational profile.

Typical notification channels may include:

```text
Push
SMS
WhatsApp
Email
```

Notification processing should not unnecessarily block core ride operations.

---

# 22. AI / Optimization Domain

AI capabilities include:

```text
Smart Dispatch Optimization
ETA Prediction
Demand Prediction
Supply Prediction
Ranking
Other Future Optimization
```

AI should remain modular and replaceable.

AI services should not become the authoritative owner of core ride state.

---

# 23. AI Boundary

The preferred interaction is:

```text
Domain State
      ↓
AI Input
      ↓
Prediction / Recommendation
      ↓
Domain Validation
      ↓
Action
```

not:

```text
AI
 ↓
Directly Change Core Domain State
```

unless an explicit domain contract permits the operation.

---

# 24. Candidate Service Map

A possible future service decomposition is:

```text
                    RideForge Platform
                           │
       ┌───────────────────┼───────────────────┐
       │                   │                   │
     Ride              Driver              Payment
       │                   │                   │
       └──────────────┬────┴──────────────┐    │
                      │                   │    │
                  Dispatch           Location │
                      │                   │    │
                  Matching              ETA    │
                      │                   │    │
                      └─────────┬─────────┘    │
                                │              │
                           AI / Ranking        │
                                │              │
                                └──────┬───────┘
                                       │
                                  Events / APIs
```

This represents a conceptual boundary model, not a requirement that every box immediately become a separately deployed service.

---

# 25. Module Before Service

The preferred development progression is:

```text
Domain Module
      ↓
Stable Boundary
      ↓
Operational Evidence
      ↓
Service Candidate
      ↓
Independent Service
```

This prevents premature distribution.

---

# 26. Service Extraction Criteria

A module should be considered for extraction when several of these conditions are true:

```text
[ ] Clear Domain Ownership
[ ] Clear Data Ownership
[ ] Independent Scaling Need
[ ] Independent Deployment Need
[ ] Failure Isolation Benefit
[ ] Distinct Operational Profile
[ ] Stable API / Event Contract
[ ] Reasonable Migration Cost
```

The presence of only one weak criterion is normally insufficient.

---

# 27. Strong Extraction Signals

Strong reasons include:

### 27.1 Scaling Difference

Example:

```text
Location
```

requires significantly more throughput than:

```text
Driver Profile
```

---

### 27.2 Failure Isolation

A capability may need independent failure containment.

Example:

```text
AI Inference
```

should not bring down:

```text
Ride Creation
```

---

### 27.3 Deployment Independence

A capability may need to deploy independently because it changes at a different rate.

---

### 27.4 Security Boundary

A component may require stronger isolation because it handles sensitive data or privileged operations.

---

# 28. Weak Extraction Signals

Do not extract a service solely because:

```text
The package is large
The code has many files
A framework differs
The code feels separate
Microservices are popular
```

---

# 29. Data Ownership Rule

Each service must have clearly defined ownership of persistent business state.

A service should:

```text
Read Its Own State
Write Its Own State
Expose Required Data Through Contracts
```

rather than allowing unrestricted table access.

---

# 30. Shared Database Rule

During early development, multiple services or modules may physically use the same PostgreSQL deployment.

This is acceptable if:

```text
Logical Ownership
```

remains explicit.

However, direct cross-domain writes should be avoided.

---

# 31. Cross-Service Data Access

Preferred:

```text
Service A
   │
   ├── API Request
   │
   └── Event
          ↓
      Service B
```

Avoid:

```text
Service A
   ↓
Service B Database Tables
```

for normal business writes.

---

# 32. Read Models

Where a service needs information owned by another domain, suitable approaches include:

```text
API Query
Event-Driven Projection
Read Model
Cached Representation
```

The choice depends on:

```text
Freshness
Latency
Consistency
Complexity
```

---

# 33. Event-Based Synchronization

A consumer may maintain a local representation of another domain's state through events.

Example:

```text
Driver Updated
      ↓
Driver Event
      ↓
Matching Projection
```

This can reduce synchronous coupling.

---

# 34. Synchronous Dependency Rule

Synchronous calls are appropriate when immediate information is required.

However, avoid creating long chains:

```text
A → B → C → D → E
```

because a failure in one dependency can propagate across the chain.

---

# 35. Asynchronous Dependency Rule

Use asynchronous communication where:

```text
Immediate Response Is Not Required
```

and where eventual consistency is acceptable.

---

# 36. Event Ownership

The domain that owns a business fact should publish the corresponding event.

Example:

```text
Ride Domain
   ↓
Ride Created
```

rather than an unrelated service generating a duplicate representation of the same fact.

---

# 37. Event Contract

Events should have:

```text
Stable Meaning
Versioned Contract
Explicit Producer
Explicit Consumers
Idempotent Consumption Strategy
```

The detailed event strategy is documented separately.

---

# 38. API Contract

Services should expose explicit APIs when synchronous communication is required.

Contracts should define:

```text
Request
Response
Errors
Timeout Expectations
Authentication
Authorization
Compatibility
```

---

# 39. Service Independence

A service is not truly independent if it requires:

```text
Another Service's Database
Another Service's Internal Code
Another Service's Deployment
Another Service's Internal Events
```

to function for ordinary operations.

Dependencies should be explicit.

---

# 40. Deployment Independence

Independent deployment should be treated as a benefit of a stable service boundary, not as the starting assumption.

Before extraction, verify:

```text
Contract Stability
Dependency Isolation
Data Ownership
Operational Readiness
```

---

# 41. Failure Isolation

Services should have failure boundaries appropriate to their criticality.

For example:

```text
AI Failure
    ↓
Fallback
    ↓
Dispatch Continues
```

where supported by the business rules.

Similarly:

```text
Notification Failure
    ↓
Notification Retry
    ↓
Ride Lifecycle Continues
```

where notifications are not transaction-critical.

---

# 42. Criticality Classification

Not every service has the same criticality.

A conceptual classification is:

```text
Critical
Important
Supporting
Asynchronous / Non-Critical
```

Critical domains should have stronger:

```text
Availability
Failure Handling
Observability
Testing
Recovery
```

requirements.

---

# 43. Ride Lifecycle Criticality

The Ride lifecycle is a core platform capability.

Architectural changes affecting it require careful consideration of:

```text
Consistency
Availability
State Transitions
Failure Recovery
Idempotency
```

---

# 44. Dispatch Criticality

Dispatch is also operationally critical because it directly affects:

```text
Assignment
Driver Utilization
Passenger Experience
Operational Throughput
```

Therefore dispatch boundaries require strong failure and fallback design.

---

# 45. Matching Criticality

Matching must be able to degrade safely.

If advanced ranking becomes unavailable, the platform should be able to use an approved deterministic strategy where business rules permit it.

---

# 46. Location Criticality

Location data can be high-volume and latency-sensitive.

The architecture should therefore isolate location workload characteristics from slower transactional workflows where appropriate.

---

# 47. ETA Criticality

ETA improves user and dispatch decisions but should not unnecessarily become a hard dependency for every ride lifecycle operation.

A defined fallback must exist where ETA is required operationally.

---

# 48. AI Criticality

AI should generally be treated as:

```text
Optimization Capability
```

with explicit fallback.

AI availability should not automatically determine core platform availability.

---

# 49. Payment Criticality

Payment operations may require stronger consistency and external provider handling.

Payment should therefore maintain an explicit domain boundary.

---

# 50. Notification Criticality

Notifications are generally suitable for asynchronous processing.

A notification outage should not automatically stop the core ride lifecycle unless a specific business rule requires it.

---

# 51. Service-to-Service Security

As services become independently deployed, communication should be protected through appropriate:

```text
Authentication
Authorization
Network Controls
Secret Management
```

Service boundaries do not imply trust by default.

---

# 52. Service Configuration

Each service should have clearly defined configuration for:

```text
Database
Redis
Messaging
External Providers
Timeouts
Feature Flags
AI Models
```

Configuration should not be embedded in business logic.

---

# 53. Service Observability

Each independently deployed service should provide:

```text
Health
Readiness
Metrics
Structured Logs
Tracing Where Appropriate
Dependency Visibility
```

---

# 54. Service Testing

A service boundary should support:

```text
Unit Testing
Integration Testing
Contract Testing
Failure Testing
```

without requiring unrelated services to be fully operational for every unit-level test.

---

# 55. Local Development

Local development should allow developers to work on a service without requiring the entire distributed production environment.

Where necessary, developers should use:

```text
Local Dependencies
Containers
Test Doubles
Fixtures
Local Event Infrastructure
```

according to the development documentation.

---

# 56. Service Repository Strategy

Service extraction does not automatically require a separate repository.

The repository strategy should be chosen based on:

```text
Team Ownership
Deployment Pipeline
Release Independence
Code Sharing
Operational Requirements
```

A service may initially remain within a monorepo.

---

# 57. Monorepo Compatibility

The architecture supports a monorepo approach while maintaining strong service boundaries.

A monorepo does not mean:

```text
Shared Everything
```

and multiple repositories do not automatically mean:

```text
Independent Architecture
```

The actual boundaries are defined by:

```text
Contracts
Ownership
Dependencies
Data
Deployment
```

---

# 58. Shared Libraries

Shared libraries should be limited to genuinely cross-cutting capabilities.

Examples:

```text
Logging
Tracing
Configuration Utilities
Common Protocol Types
```

Avoid creating a large shared library containing domain logic for multiple services.

---

# 59. Shared Domain Logic

Do not share domain logic simply to avoid duplication if doing so creates strong coupling between service boundaries.

Prefer explicit domain ownership.

---

# 60. Versioning

Independent service contracts should be versioned when compatibility requires it.

Versioning applies to:

```text
API Contracts
Event Contracts
Shared Schemas
```

---

# 61. Backward Compatibility

When evolving a service:

```text
Existing Consumers
```

should continue working according to the supported compatibility policy.

Breaking changes should be deliberate.

---

# 62. Service Lifecycle

A service should have a lifecycle:

```text
Proposed
    ↓
Implemented
    ↓
Production
    ↓
Scaled / Evolved
    ↓
Deprecated
    ↓
Retired
```

Service creation should therefore include consideration of its eventual retirement.

---

# 63. Service Retirement

Before removing a service:

```text
[ ] Consumers identified
[ ] Data ownership migrated
[ ] Contracts deprecated
[ ] Traffic removed
[ ] Monitoring removed
[ ] Deployment removed
[ ] Documentation updated
```

---

# 64. Architecture Smell: Distributed Monolith

Indicators include:

```text
Every Request Requires Multiple Services
Shared Database Writes
Synchronous Chains
Coupled Deployments
Shared Internal Libraries
Frequent Cross-Service Changes
```

If these become common, the architecture should be reviewed.

---

# 65. Architecture Smell: Service Explosion

Indicators include:

```text
Very Small Services
Identical Scaling
Identical Deployment
High Network Traffic
Excessive Operational Overhead
No Clear Data Ownership
```

This indicates premature decomposition.

---

# 66. Architecture Smell: Hidden Coupling

Indicators include:

```text
Direct Table Access
Undocumented API Dependencies
Implicit Events
Shared Mutable State
Hidden Environment Dependencies
```

These should be treated as architectural defects.

---

# 67. Boundary Review Process

Before creating a new service, evaluate:

```text
1. What business capability does it own?
2. What state does it own?
3. Who consumes it?
4. Why does it need independent deployment?
5. Why does it need independent scaling?
6. What failures should it isolate?
7. What contracts does it expose?
8. What dependencies does it require?
9. What operational burden does it add?
10. Can the same benefit be achieved as a module?
```

---

# 68. Service Extraction Checklist

```text
[ ] Domain boundary is clear
[ ] Responsibility is cohesive
[ ] Data ownership is clear
[ ] API / event contract is clear
[ ] Dependencies are explicit
[ ] Shared database writes removed or controlled
[ ] Failure boundary defined
[ ] Fallback defined where required
[ ] Observability implemented
[ ] Tests implemented
[ ] Deployment defined
[ ] Configuration defined
[ ] Security defined
[ ] Rollback defined
```

---

# 69. Boundary Decision Matrix

| Criterion | Module | Service Candidate | Strong Service Candidate |
|---|---|---|---|
| Domain ownership | Moderate | Clear | Strong |
| Data ownership | Shared / emerging | Mostly clear | Independent |
| Scaling | Similar | Some difference | Strongly different |
| Deployment | Coupled | Useful to separate | Must separate |
| Failure isolation | Low | Useful | Critical |
| Contract stability | Emerging | Defined | Stable |
| Operational overhead | Low | Moderate | Justified |
| Extraction value | Low | Medium | High |

This matrix is a decision aid, not an automatic scoring system.

---

# 70. Consequences

## 70.1 Positive Consequences

The decision provides:

```text
Clear Domain Ownership
Controlled Service Growth
Independent Scaling When Needed
Failure Isolation
Explicit Contracts
Future Service Extraction
Reduced Direct Coupling
```

---

## 70.2 Negative Consequences

The architecture requires discipline around:

```text
Data Ownership
Contracts
Events
Service Dependencies
Observability
Testing
Deployment
```

As more services are extracted, operational complexity increases.

---

# 71. Risks

## Risk 1 — Incorrect Boundaries

A domain may be split at the wrong location.

### Mitigation

Use:

```text
Business Capability
Data Ownership
Change Frequency
Scaling
Failure Characteristics
```

to evaluate boundaries.

---

## Risk 2 — Distributed Monolith

### Mitigation

Avoid:

```text
Long Synchronous Chains
Shared Writes
Hidden Dependencies
```

---

## Risk 3 — Premature Extraction

### Mitigation

Start with a well-defined module and extract only when evidence supports the decision.

---

## Risk 4 — Excessive Shared Libraries

### Mitigation

Keep shared libraries focused on genuinely cross-cutting concerns.

---

## Risk 5 — Inconsistent Contracts

### Mitigation

Use explicit API and event contracts with compatibility rules.

---

# 72. Validation

Service boundaries should be validated through:

```text
Code Dependency Analysis
Integration Testing
Contract Testing
Load Testing
Failure Testing
Operational Monitoring
Production Experience
```

The boundary should be reconsidered if it consistently produces excessive coupling or operational overhead.

---

# 73. Review Triggers

Revisit this ADR when:

```text
A New Major Domain Is Introduced
A Service Is Extracted
A Service Is Merged
Database Ownership Changes
Traffic Characteristics Change
A Major Failure Reveals a Boundary Problem
A Team Ownership Model Changes
```

---

# 74. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
```

Relevant documentation includes:

```text
Components and Services
Code Structure and Conventions
Database Development
Redis Development
Event and Messaging Development
API Development
Testing
Observability
Performance
AI Architecture
AI Failure and Fallback
```

---

# 75. Related ADRs

This decision establishes the basis for:

```text
ADR-0002 — Architecture Style
ADR-0004 — Domain-Driven Design
ADR-0005 — Event-Driven Architecture
ADR-0012 — Outbox Pattern
ADR-0013 — Dead Letter Queue Strategy
ADR-0014 — API and Service Communication
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0029 — Architecture Evolution and Migration
```

---

# 76. Decision Summary

RideForge will define microservice boundaries around:

```text
Business Capability
+
Domain Ownership
+
Data Ownership
+
Independent Scaling
+
Failure Isolation
+
Deployment Independence
```

The platform will follow:

```text
Module First
      ↓
Stable Boundary
      ↓
Measure Requirements
      ↓
Extract When Justified
      ↓
Operate Independently
```

rather than:

```text
Create Service
      ↓
Find a Reason Later
```

---

# 77. Final Principle

> **A RideForge service boundary must represent a meaningful ownership boundary, not merely a code boundary.**

The target architecture is:

```text
Cohesive Domains
      +
Explicit Contracts
      +
Clear Data Ownership
      +
Controlled Distribution
      +
Independent Scaling Where Justified
```

The number of services is not the goal.

The goal is a platform that can evolve safely as RideForge's:

```text
Traffic
Business
Regions
Operational Requirements
AI Capabilities
Engineering Team
```

grow over time.

---

# 78. Status

```text
Decision: ACCEPTED

Microservice Boundary Strategy:
Domain-Oriented
+
Ownership-Driven
+
Evidence-Based
+
Incrementally Extracted
```

This decision is subordinate to higher-priority business, legal, and domain constraints and provides the boundary framework for subsequent RideForge architectural decisions.
