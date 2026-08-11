# ADR-0005: Event-Driven Architecture

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Architecture  
> **Scope:** RideForge asynchronous communication and event-driven integration  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is a domain-oriented, modular, microservice-compatible ride-hailing platform.

The platform contains workflows that naturally cross domain boundaries:

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

Not every workflow should require direct synchronous communication between all participating components.

A ride-hailing platform also has workloads that may be:

```text
High Volume
Asynchronous
Latency Sensitive
Failure Prone
Independently Scalable
Operationally Independent
```

Examples include:

```text
Ride lifecycle events
Driver status changes
Matching events
Dispatch events
Notifications
Analytics
AI data pipelines
Operational integrations
```

A purely synchronous architecture could create excessive coupling and long request chains.

A purely asynchronous architecture would introduce unnecessary complexity for operations that genuinely require an immediate response.

RideForge therefore needs a balanced event-driven architecture.

---

# 2. Problem

Without a deliberate event-driven strategy, the system can develop:

```text
Long Synchronous Chains
Tight Service Coupling
Failure Propagation
Repeated Database Reads
Duplicate Business Logic
Difficult Scaling
```

At the same time, indiscriminate event usage can create:

```text
Event Spaghetti
Hidden Control Flow
Eventual Consistency Problems
Debugging Complexity
Ordering Problems
Duplicate Processing
Operational Overhead
```

The architecture must therefore define:

```text
When to use events
What events represent
Who owns events
How events are delivered
How failures are handled
How consumers remain reliable
```

---

# 3. Decision

RideForge will use an **event-driven architecture for appropriate asynchronous domain and integration workflows**, while retaining synchronous APIs for operations that require immediate responses.

The communication model is therefore:

```text
Synchronous APIs
+
Asynchronous Events
```

rather than:

```text
Synchronous Everything
```

or:

```text
Asynchronous Everything
```

---

# 4. Architectural Principle

The primary rule is:

> **Use synchronous communication when the caller needs an immediate result; use events when a meaningful fact can be processed asynchronously.**

---

# 5. Communication Model

The high-level communication model is:

```text
                    RideForge
                       │
          ┌────────────┴────────────┐
          │                         │
     Synchronous                Asynchronous
       APIs                       Events
          │                         │
          ▼                         ▼
 Immediate Response          Event Stream / Broker
                                    │
                       ┌────────────┼────────────┐
                       ▼            ▼            ▼
                    Consumer     Consumer     Consumer
```

---

# 6. Why Event-Driven Architecture

Event-driven integration provides RideForge with:

```text
Loose Coupling
Asynchronous Processing
Independent Consumers
Horizontal Scalability
Failure Isolation
Extensibility
Integration Flexibility
```

These benefits are particularly valuable for a ride-hailing platform with multiple real-time and background workloads.

---

# 7. Event-Driven Architecture Is Not Universal

RideForge will not convert every operation into an event.

For example:

```text
Create Ride
```

may require an immediate API response.

Whereas:

```text
Ride Created
```

may trigger multiple asynchronous consumers.

Therefore:

```text
Command / Request
→ Often Synchronous

Fact / Notification
→ Often Asynchronous
```

---

# 8. Synchronous Operations

Synchronous communication is appropriate when:

```text
Immediate Feedback Is Required
The Caller Needs the Result
The Operation Is Short-Lived
Strong Immediate Consistency Is Required
```

Examples include:

```text
Create Ride Request
Get Ride State
Retrieve Current User Information
Validate Immediate Request Data
```

The exact API contract is governed by API architecture and development documentation.

---

# 9. Asynchronous Operations

Events are appropriate when:

```text
Immediate Response Is Not Required
Multiple Consumers Need the Same Fact
Processing May Be Expensive
Processing Can Be Retried
Consumers Need Independent Scaling
```

Examples include:

```text
Notifications
Analytics
Operational Processing
AI Feature Collection
Some Dispatch Workflows
Some Integration Workflows
```

---

# 10. Domain Events

A domain event represents a meaningful business fact.

Examples:

```text
RideCreated
RideCancelled
RideCompleted
DriverStatusChanged
DriverAssigned
MatchCreated
```

A domain event should communicate:

> **Something meaningful happened.**

---

# 11. Integration Events

An integration event is intended to communicate information across a system boundary.

Conceptually:

```text
Domain Fact
     ↓
Integration Contract
     ↓
External / Other Service Consumer
```

The implementation may use the same event infrastructure, but the semantic purpose should remain clear.

---

# 12. Commands vs Events

RideForge should distinguish commands from events.

```text
Command
→ Request something to happen

Event
→ State something that happened
```

Example:

```text
AssignDriver
```

may represent a requested operation.

Whereas:

```text
DriverAssigned
```

represents a completed fact.

---

# 13. Event Ownership

The domain that owns the business fact should own the publication of that fact.

For example:

```text
Ride Domain
     ↓
RideCreated
```

The event should not be invented by an unrelated consumer merely because that consumer needs the information.

---

# 14. Event Meaning

Events should describe business or integration meaning rather than infrastructure details.

Prefer:

```text
RideCreated
```

over:

```text
RideTableUpdated
```

when the event is intended to represent the business fact.

---

# 15. Event Granularity

Events should contain enough information for their intended consumers without becoming unrestricted snapshots of the entire domain.

Avoid:

```text
Huge Event Payloads
Internal Implementation Details
Unnecessary Sensitive Data
```

---

# 16. Event Contracts

An event contract should define, as appropriate:

```text
Event Name
Event Version
Event ID
Timestamp
Producer
Aggregate / Entity Identifier
Payload
Correlation Information
```

The exact envelope structure is governed by the event implementation documentation.

---

# 17. Event Identity

Every published event should have a stable event identifier where the infrastructure requires one.

This supports:

```text
Deduplication
Tracing
Debugging
Auditability
Consumer Idempotency
```

---

# 18. Correlation

Events should support correlation across a business workflow.

Conceptually:

```text
Ride Request
      ↓
Correlation ID
      ↓
Ride Event
      ↓
Matching Event
      ↓
Dispatch Event
```

This makes distributed workflows easier to trace.

---

# 19. Causation

Where useful, events may also carry causation information.

Conceptually:

```text
Event B
```

may identify:

```text
Event A
```

as the event that caused the resulting action.

This can improve debugging of complex workflows.

---

# 20. Event Timestamp

Events should include an appropriate timestamp representing the event occurrence or publication semantics defined by the event contract.

Consumers should not assume that processing time is the same as event occurrence time.

---

# 21. Event Versioning

Events are contracts.

If the meaning or structure of an event changes incompatibly, the contract must be versioned or migrated deliberately.

Avoid silently changing an event in a way that breaks existing consumers.

---

# 22. Backward Compatibility

Consumers should be able to tolerate compatible producer evolution where practical.

For example:

```text
Adding Optional Information
```

may be less disruptive than:

```text
Removing Existing Fields
Changing Meaning
Changing Data Types
```

Compatibility requirements depend on the specific event contract.

---

# 23. Event Schema Evolution

When an event changes:

```text
Identify Consumers
      ↓
Assess Compatibility
      ↓
Choose Evolution Strategy
      ↓
Deploy Safely
      ↓
Monitor Consumers
```

---

# 24. Event Broker

RideForge uses a Kafka-compatible event-streaming architecture, with the current production direction using:

```text
Redpanda
```

The broker provides:

```text
Event Streaming
Persistence
Consumer Groups
Partitioning
Scalable Consumption
```

The architectural decision is about event-driven integration; the detailed infrastructure choice is recorded separately.

---

# 25. Broker Abstraction

Business logic should not depend directly on broker-specific implementation details.

Conceptually:

```text
Application / Domain
       ↓
Messaging Contract
       ↓
Redpanda / Kafka-Compatible Infrastructure
```

This keeps broker concerns within infrastructure boundaries.

---

# 26. Topics

Events should be organized into meaningful topics or equivalent streams.

A topic should represent a coherent event category rather than becoming a miscellaneous event bucket.

Examples may include:

```text
ride.events
match.events
```

Additional topics should be introduced based on actual domain and operational requirements.

---

# 27. Topic Ownership

Topic ownership should be explicit.

A producing domain should know which event contracts it owns.

Consumers should not assume ownership of a producer's topic merely because they consume from it.

---

# 28. Partitioning

Partitioning should be designed around the ordering and scaling requirements of the event stream.

Possible keys may include:

```text
Ride ID
Driver ID
Aggregate ID
Region
```

The correct key depends on the event's ordering requirements.

---

# 29. Ordering

RideForge should only rely on ordering guarantees that the messaging infrastructure and chosen partitioning strategy actually provide.

Do not assume:

```text
Global Ordering
```

unless it is explicitly implemented and justified.

---

# 30. Aggregate Ordering

When events for the same aggregate require ordering, they should use a stable partitioning strategy that preserves the required ordering within that aggregate.

For example:

```text
RideCreated
RideMatched
DriverAssigned
RideStarted
RideCompleted
```

may require ordered handling for a particular ride.

---

# 31. Consumer Groups

Independent processing workloads should use appropriate consumer groups.

Conceptually:

```text
Topic
 │
 ├── Matching Consumer Group
 ├── Notification Consumer Group
 ├── Analytics Consumer Group
 └── AI Consumer Group
```

Each consumer group can process the same event stream independently.

---

# 32. Consumer Independence

Consumers should remain independently deployable and scalable where appropriate.

One consumer becoming slow should not unnecessarily prevent unrelated consumers from processing the event stream.

---

# 33. Consumer Idempotency

Consumers must be designed to tolerate duplicate event delivery where the delivery model permits duplicates.

A consumer should be safe if the same event is processed more than once.

This is especially important for:

```text
Payments
Assignments
Notifications
State Updates
```

where duplicate side effects may be harmful.

---

# 34. Idempotency Strategy

Idempotency may be implemented through mechanisms such as:

```text
Event ID Tracking
Idempotency Keys
Unique Constraints
State Checks
Deduplication Records
```

The detailed platform-wide idempotency decision is documented separately.

---

# 35. At-Least-Once Processing

RideForge should design asynchronous consumers assuming that duplicate delivery or retry may occur.

The system should therefore favor:

```text
Reliable Delivery
+
Idempotent Consumers
```

over assuming exactly-once business effects everywhere.

---

# 36. Exactly-Once Consideration

Infrastructure-level exactly-once semantics should not be treated as a guarantee that an external business side effect occurs exactly once.

Business-level idempotency remains necessary.

---

# 37. Event Publication Reliability

A critical problem in event-driven systems is:

```text
Database Transaction Succeeds
```

while:

```text
Event Publication Fails
```

This can produce inconsistent system behaviour.

RideForge therefore uses the Outbox pattern for appropriate transactional event publication.

The detailed decision is documented separately in:

```text
ADR-0012 — Outbox Pattern
```

---

# 38. Transactional Event Flow

The preferred pattern is:

```text
Application Command
      ↓
Local Database Transaction
      │
      ├── Domain State Change
      │
      └── Outbox Record
              ↓
          Outbox Publisher
              ↓
            Broker
              ↓
          Consumers
```

This avoids relying on an unsafe dual write.

---

# 39. Consumer Failure

A consumer may fail because of:

```text
Application Error
Database Error
Dependency Failure
Temporary Network Failure
Invalid Input
Unexpected Event
```

The architecture must support retry and failure isolation.

---

# 40. Retry Strategy

Retry should be used for failures that are potentially recoverable.

Examples:

```text
Temporary Database Failure
Temporary Network Failure
Transient Provider Failure
```

Do not retry indefinitely when the event itself is permanently invalid.

---

# 41. Dead Letter Queue

Events that cannot be successfully processed after the defined retry policy may be routed to a:

```text
Dead Letter Queue
```

The detailed DLQ decision is documented separately.

---

# 42. DLQ Principle

A DLQ should provide:

```text
Failure Isolation
Investigation
Controlled Replay
Operational Visibility
```

A DLQ is not a place where failed events are permanently forgotten.

---

# 43. Poison Events

A poison event is an event that repeatedly fails processing because the problem is not transient.

Examples:

```text
Invalid Schema
Invalid Business Data
Unsupported Version
Programming Error
```

Such events should be isolated rather than endlessly retried.

---

# 44. Event Replay

Event replay may be supported where the event contract and consumer semantics permit it.

Replay must account for:

```text
Idempotency
Side Effects
Current State
Event Version
Ordering
```

Do not replay production events blindly.

---

# 45. Consumer Recovery

A recovering consumer should be able to resume processing without creating incorrect duplicate business effects.

This requires:

```text
Checkpointing
Idempotency
Durable State
Controlled Retries
```

according to the infrastructure implementation.

---

# 46. Event Retention

Event retention should be selected based on:

```text
Recovery Requirements
Replay Requirements
Storage Cost
Compliance
Operational Needs
```

Retention should not be unlimited by default.

---

# 47. Event Payload Design

Payloads should be:

```text
Minimal Enough
Self-Describing Enough
Stable Enough
Safe Enough
```

Avoid putting every available database field into an event.

---

# 48. Sensitive Data

Events should contain only data required by their consumers.

Avoid unnecessary publication of:

```text
Passwords
Secrets
Payment Credentials
Unnecessary Personal Data
Internal Security Information
```

---

# 49. Event Contracts and Privacy

Event-driven systems can distribute data widely.

Therefore event design must consider:

```text
Who Consumes This Event?
What Data Do They Need?
How Long Is It Retained?
Can It Be Replayed?
```

---

# 50. Event Contracts and Security

Consumers should not automatically receive privileged access merely because they can subscribe to a topic.

Messaging infrastructure should support appropriate access controls.

---

# 51. Event-Driven Ride Lifecycle

A conceptual ride workflow may look like:

```text
Client
  ↓
Create Ride
  ↓
Ride Domain
  ↓
RideCreated
  ↓
Matching / Dispatch
  ↓
Match / Assignment Events
  ↓
Ride Lifecycle
  ↓
Completion
```

Not every step must be asynchronous.

The exact workflow depends on business and latency requirements.

---

# 52. Event-Driven Matching

Matching may consume events such as:

```text
RideCreated
```

and use current operational state to identify candidates.

Matching should not assume that the event itself contains all current driver state.

---

# 53. Event-Driven Dispatch

Dispatch may consume:

```text
Ride
Match
Driver
Location
```

signals according to its design.

The dispatch strategy remains responsible for applying the appropriate operational rules.

---

# 54. Event-Driven Notifications

Notification processing is a strong asynchronous candidate.

Conceptually:

```text
Ride State Event
      ↓
Notification Consumer
      ↓
Channel Provider
```

Notification failure should generally not block the core ride lifecycle.

---

# 55. Event-Driven Analytics

Analytics consumers may independently process operational events.

This allows analytics workloads to scale without placing analytical queries directly on critical transactional workflows.

---

# 56. Event-Driven AI Data

AI and machine-learning pipelines may consume operational events to build:

```text
Features
Training Data
Feedback Signals
Prediction Inputs
```

The AI data pipeline should remain separate from critical transactional event handling where possible.

---

# 57. Event-Driven Observability

Operational events can support:

```text
Metrics
Tracing
Audit Information
Operational Dashboards
```

However, observability should not depend solely on business events.

Infrastructure-level telemetry remains necessary.

---

# 58. Event-Driven Auditability

Important domain events can provide useful historical context.

However, event streams should not automatically be treated as the complete audit log unless an explicit decision establishes that requirement.

---

# 59. Event-Driven Consistency

Consumers must explicitly understand whether their data is:

```text
Immediately Consistent
```

or:

```text
Eventually Consistent
```

Event-driven integration naturally introduces asynchronous processing boundaries.

---

# 60. User-Facing Consistency

User-facing operations should not expose confusing intermediate states without appropriate handling.

For example:

```text
Ride Created
```

may be returned immediately while:

```text
Matching
```

continues asynchronously.

The client-facing API should communicate the resulting state clearly.

---

# 61. Event Processing Time

Consumers should be designed for the possibility that:

```text
Event Creation Time
```

and:

```text
Event Processing Time
```

are different.

An event may arrive later due to:

```text
Backlog
Retry
Network Delay
Consumer Recovery
```

---

# 62. Out-of-Order Events

Consumers should not assume global event ordering.

Where ordering is required:

```text
Use Appropriate Partitioning
Validate State
Handle Unexpected Sequence
```

---

# 63. Stale Events

A consumer should protect itself against stale information when necessary.

For example:

```text
Older Driver State
```

should not overwrite:

```text
Newer Driver State
```

when the domain requires monotonic updates.

---

# 64. Event Time vs Processing Time

When domain semantics require it, consumers should distinguish:

```text
Occurred At
```

from:

```text
Processed At
```

This is particularly relevant to:

```text
Location
ETA
Demand
Supply
AI Features
Analytics
```

---

# 65. Backpressure

Consumers may process events slower than producers publish them.

The architecture must support controlled backpressure through:

```text
Consumer Scaling
Partitioning
Batching
Rate Limiting
Retry Control
```

as appropriate.

---

# 66. Consumer Lag

Consumer lag is an important operational signal.

Significant lag may indicate:

```text
Consumer Saturation
Broker Pressure
Dependency Failure
Insufficient Partitions
Processing Inefficiency
```

---

# 67. Event Processing Latency

For latency-sensitive consumers, monitor:

```text
Event Creation
→
Event Consumption
→
Processing Completion
```

to identify delays.

---

# 68. Event Failure Metrics

Important metrics may include:

```text
Events Published
Events Consumed
Consumer Errors
Retry Count
DLQ Count
Consumer Lag
Processing Latency
```

---

# 69. Event Correlation and Tracing

Distributed workflows should preserve correlation information so that an operational trace can follow:

```text
API Request
 ↓
Domain Operation
 ↓
Event
 ↓
Consumer
 ↓
External Dependency
```

where supported.

---

# 70. Event Testing

Event-driven code should be tested at multiple levels.

Recommended categories:

```text
Unit Tests
Contract Tests
Integration Tests
Failure Tests
Replay Tests
Idempotency Tests
Ordering Tests
```

---

# 71. Event Contract Testing

Consumers and producers should validate event contracts.

Tests should detect:

```text
Breaking Schema Changes
Missing Required Fields
Invalid Types
Unexpected Event Versions
```

---

# 72. Consumer Idempotency Testing

Tests should intentionally process the same event multiple times.

Expected result:

```text
One Correct Business Effect
```

rather than:

```text
Multiple Duplicate Effects
```

---

# 73. Failure Testing

Important failure scenarios include:

```text
Broker Unavailable
Consumer Database Unavailable
External Provider Unavailable
Malformed Event
Duplicate Event
Out-of-Order Event
Consumer Restart
Publisher Restart
```

---

# 74. Event Replay Testing

Where replay is supported, test:

```text
Historical Event
      ↓
Consumer
      ↓
Expected State
```

before using replay operationally.

---

# 75. Local Development

Local development should provide a practical event infrastructure.

Developers should be able to run relevant services with:

```text
Local Redpanda / Kafka-Compatible Broker
PostgreSQL
Redis
Application Services
```

as required by the workflow being developed.

---

# 76. Local Event Testing

A developer should be able to:

```text
Publish Test Event
Consume Event
Inspect Event
Trigger Retry
Inspect DLQ
```

without requiring production infrastructure.

---

# 77. Event Observability

Every important consumer should expose enough information to identify:

```text
Topic
Partition
Event ID
Consumer Group
Processing Result
Retry Count
Correlation ID
```

where operationally useful.

---

# 78. Event Security

Messaging access should follow least privilege.

A consumer should receive access only to the streams required for its responsibility.

---

# 79. Event Infrastructure Failure

If the broker becomes unavailable:

```text
Critical Domain Transaction
```

should not automatically be rolled back unless the business requirement explicitly demands synchronous event publication.

The Outbox pattern allows the local transaction to succeed while publication is retried later.

---

# 80. Event Broker Recovery

When the broker recovers:

```text
Pending Outbox Records
      ↓
Publisher
      ↓
Broker
      ↓
Consumers
```

should resume according to the retry and delivery strategy.

---

# 81. Event Publishing Failure

Publishing failure should be observable and recoverable.

The system should avoid silently dropping business-critical events.

---

# 82. Consumer Dependency Failure

If a consumer's downstream dependency is unavailable:

```text
Consume
  ↓
Process
  ↓
Dependency Failure
  ↓
Retry / Backoff
  ↓
DLQ if Necessary
```

The failure should not block unrelated consumers.

---

# 83. Event Ordering and Dispatch

Dispatch and matching workflows may require careful ordering.

For example:

```text
RideCreated
```

must not be interpreted as:

```text
RideCompleted
```

or:

```text
RideCancelled
```

if later state transitions have already occurred.

Consumers should validate state when ordering guarantees alone are insufficient.

---

# 84. Event-Driven Architecture and AI

AI consumers may process operational events for:

```text
Feature Generation
Demand Prediction
Supply Prediction
ETA Models
Ranking
Feedback
```

AI workloads should not be allowed to overload critical business event consumers.

Separate consumer groups and appropriate resource controls should be used where needed.

---

# 85. Event-Driven Architecture and Legal Constraints

Events must not bypass authoritative regional and legal validation.

For example:

```text
RideCreated
      ↓
Matching
      ↓
Regional Eligibility
      ↓
Dispatch
```

Legal and operational validation remains authoritative.

---

# 86. Event-Driven Architecture and Payments

Payment events should be treated as high-integrity domain/integration events.

Consumers must be idempotent because duplicate processing could create financial side effects.

---

# 87. Event-Driven Architecture and Notifications

Notification events should contain enough information to deliver the message but should avoid unnecessary sensitive information.

Notification consumers should support retries and controlled failure handling.

---

# 88. Event-Driven Architecture and Data Ownership

Events provide information.

They do not transfer ownership of the underlying domain state.

For example:

```text
DriverUpdated
```

does not make the consumer the owner of Driver state.

---

# 89. Event-Driven Architecture and Service Boundaries

Events should reinforce service boundaries rather than bypass them.

Preferred:

```text
Service A
    ↓
Event Contract
    ↓
Service B
```

Avoid:

```text
Service A
    ↓
Internal State of Service B
```

---

# 90. Event-Driven Architecture and DDD

DDD defines:

```text
Who Owns the Business Fact
```

Event-driven architecture defines:

```text
How That Fact Can Be Communicated
```

Therefore:

```text
DDD
 ↓
Domain Ownership
 ↓
Domain Event
 ↓
Integration Event
 ↓
Messaging Infrastructure
```

---

# 91. Event-Driven Architecture and Microservices

Microservices may communicate through:

```text
Synchronous APIs
```

and:

```text
Asynchronous Events
```

Event-driven architecture does not require every service interaction to be asynchronous.

---

# 92. Event-Driven Architecture and API Design

APIs and events are complementary.

A common workflow may be:

```text
API Request
      ↓
Local Transaction
      ↓
Event Publication
      ↓
Asynchronous Processing
```

---

# 93. Event-Driven Architecture and Database Transactions

A database transaction should normally protect the local domain state.

Event publication should be reliably connected to that transaction through the Outbox pattern where required.

---

# 94. Event-Driven Architecture and Idempotency

Because asynchronous processing may be retried, consumers should be idempotent.

This is a fundamental requirement, not an optional optimization.

---

# 95. Event-Driven Architecture and Dead Letter Queues

DLQs provide controlled handling of events that cannot be processed normally.

The architecture should provide:

```text
Retry
+
Backoff
+
DLQ
+
Investigation
+
Controlled Replay
```

where appropriate.

---

# 96. Event-Driven Architecture and Performance

Events can improve scalability by decoupling workloads, but they also introduce:

```text
Serialization
Broker Latency
Consumer Processing
Eventual Consistency
Operational Overhead
```

Performance decisions must consider the complete workflow.

---

# 97. Event-Driven Architecture and Cost

Messaging infrastructure introduces:

```text
Broker Cost
Storage Cost
Network Cost
Operational Cost
Engineering Cost
```

Event-driven architecture should therefore be applied where its benefits justify the cost.

---

# 98. Event-Driven Architecture and Simplicity

The simplest correct communication mechanism should be preferred.

If a local function call is sufficient:

```text
Do Not Introduce an Event
```

If an API is sufficient:

```text
Do Not Introduce Asynchronous Messaging Solely for Fashion
```

---

# 99. Decision Matrix

| Requirement | Preferred Communication |
|---|---|
| Immediate response required | Synchronous API |
| Immediate validation required | Synchronous API |
| Multiple independent consumers | Event |
| Background processing | Event |
| Analytics | Event |
| Notification | Event |
| AI feature pipeline | Event |
| Strong local transaction | Local transaction |
| Cross-domain asynchronous workflow | Event |
| Simple internal function call | Direct function call |
| External provider request requiring immediate result | Synchronous adapter |

This is a guideline rather than an absolute rule.

---

# 100. Event Adoption Checklist

Before introducing an event:

```text
[ ] Is there a meaningful business/integration fact?
[ ] Is asynchronous processing useful?
[ ] Is immediate response unnecessary?
[ ] Is the event owner clear?
[ ] Are consumers known?
[ ] Is the event contract defined?
[ ] Is idempotency addressed?
[ ] Is retry behaviour defined?
[ ] Is DLQ behaviour defined where necessary?
[ ] Is observability available?
[ ] Is sensitive data minimized?
```

---

# 101. New Consumer Checklist

When adding a consumer:

```text
[ ] Consumer responsibility is clear
[ ] Consumer group is defined
[ ] Event contract is validated
[ ] Idempotency is implemented
[ ] Retry policy is defined
[ ] Failure handling is defined
[ ] DLQ behaviour is defined
[ ] Metrics are available
[ ] Logs contain correlation context
[ ] Integration tests exist
```

---

# 102. New Event Checklist

When creating a new event:

```text
[ ] Event name is meaningful
[ ] Event owner is identified
[ ] Event semantics are documented
[ ] Payload is minimal
[ ] Sensitive data is reviewed
[ ] Event ID exists
[ ] Timestamp semantics are clear
[ ] Versioning strategy is defined
[ ] Consumers are identified
[ ] Ordering requirements are identified
[ ] Retry requirements are identified
[ ] Replay requirements are considered
```

---

# 103. Event Evolution Checklist

Before changing an event:

```text
[ ] Existing consumers identified
[ ] Compatibility reviewed
[ ] Versioning decision made
[ ] Migration strategy defined
[ ] Tests updated
[ ] Monitoring updated
[ ] Rollout plan defined
```

---

# 104. Consequences

## 104.1 Positive Consequences

The decision provides:

```text
Reduced Synchronous Coupling
Independent Consumer Scaling
Better Failure Isolation
Asynchronous Processing
Extensibility
Improved Integration
```

---

## 104.2 Negative Consequences

The architecture introduces:

```text
Eventual Consistency
Broker Dependency
Consumer Complexity
Retry Complexity
Ordering Challenges
Debugging Complexity
Operational Overhead
```

These are accepted trade-offs for workflows where asynchronous integration provides meaningful value.

---

# 105. Risks

## Risk 1 — Event Sprawl

Too many events can make the architecture difficult to understand.

### Mitigation

Create events only for meaningful business or integration facts.

---

## Risk 2 — Duplicate Processing

Consumers may process the same event more than once.

### Mitigation

Use idempotent consumer design.

---

## Risk 3 — Event Ordering Problems

Events may arrive in unexpected order.

### Mitigation

Use appropriate partitioning, state validation, and event sequencing rules.

---

## Risk 4 — Hidden Workflow Complexity

Asynchronous workflows can become difficult to trace.

### Mitigation

Use:

```text
Correlation IDs
Structured Logs
Metrics
Tracing
Clear Event Contracts
```

---

## Risk 5 — Broker Dependency

Critical workflows may become dependent on messaging infrastructure.

### Mitigation

Use local transactions and the Outbox pattern for appropriate critical events.

---

## Risk 6 — Sensitive Data Propagation

Events can distribute data to many consumers.

### Mitigation

Apply:

```text
Data Minimization
Access Control
Privacy Review
Retention Controls
```

---

# 106. Validation

The event-driven architecture should be validated through:

```text
Contract Tests
Integration Tests
Failure Tests
Retry Tests
Idempotency Tests
Ordering Tests
Replay Tests
Load Tests
Consumer Lag Monitoring
Production Metrics
```

---

# 107. Review Triggers

Revisit this ADR when:

```text
Messaging Platform Changes
Major Event Contract Changes
Event Volume Changes Significantly
New Critical Consumer Added
Major Reliability Failure Occurs
Eventual Consistency Becomes a Problem
A Major Workflow Changes from Sync to Async
```

---

# 108. Related Documentation

This ADR should be read with:

```text
03-components/
04-development/
05-ai/
```

Especially:

```text
Event and Messaging Development
Database Development
Redis Development
API Development
Testing
Integration Testing and Local Infrastructure
Observability Development
Performance and Optimization
AI Data Pipeline
AI Feedback and Learning Loop
AI Failure and Fallback Strategy
```

---

# 109. Related ADRs

This decision is directly related to:

```text
ADR-0002 — Architecture Style
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0012 — Outbox Pattern
ADR-0013 — Dead Letter Queue Strategy
ADR-0014 — API and Service Communication
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
```

---

# 110. Decision Summary

RideForge adopts event-driven architecture as a complementary communication model.

The platform will use:

```text
Synchronous APIs
+
Asynchronous Events
```

according to the needs of each workflow.

The event-driven system will emphasize:

```text
Explicit Event Ownership
+
Stable Contracts
+
Idempotent Consumers
+
Reliable Publication
+
Controlled Retries
+
Dead Letter Handling
+
Observability
+
Data Minimization
```

---

# 111. Final Principle

> **Events should represent meaningful facts and create useful decoupling; they should not be introduced merely to make the architecture look distributed.**

The intended architecture is:

```text
Domain Ownership
       ↓
Meaningful Event
       ↓
Reliable Publication
       ↓
Independent Consumer
       ↓
Idempotent Processing
       ↓
Observable Result
```

while preserving synchronous APIs wherever immediate interaction and consistency are required.

---

# 112. Status

```text
Decision: ACCEPTED

Event-Driven Architecture:
Synchronous APIs
+
Asynchronous Events
+
Reliable Event Publication
+
Idempotent Consumers
+
Controlled Failure Handling
```

This decision establishes the event-driven architectural foundation for the subsequent RideForge messaging, reliability, and integration ADRs.
