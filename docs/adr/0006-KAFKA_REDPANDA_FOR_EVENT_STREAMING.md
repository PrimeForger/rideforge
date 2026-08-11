# ADR-0006: Kafka / Redpanda for Event Streaming

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Infrastructure / Messaging  
> **Scope:** RideForge event-streaming infrastructure  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge uses an event-driven architecture for asynchronous communication between appropriate domain and platform components.

ADR-0005 established that RideForge will use:

```text
Synchronous APIs
+
Asynchronous Events
```

where each communication mechanism is selected according to the workflow requirements.

The platform requires an event-streaming infrastructure capable of supporting:

```text
Durable Event Streams
Consumer Groups
Partitioned Processing
Horizontal Consumer Scaling
Event Retention
Retry / Recovery Workflows
Independent Consumers
High Event Throughput
Operational Observability
```

RideForge is also being developed with cost and operational simplicity in mind.

The project has considered multiple messaging approaches, including:

```text
Redis Streams
Kafka
Redpanda
RabbitMQ
```

The selected event-streaming platform must therefore provide the required streaming characteristics without introducing unnecessary infrastructure cost or operational complexity.

---

# 2. Problem

RideForge needs a reliable event-streaming platform that can support workflows such as:

```text
Ride Events
Matching Events
Dispatch Events
Driver / Location Events
Notification Events
Analytics Events
AI Data Events
Operational Integration Events
```

The platform must also support:

```text
Consumer Groups
Partitioning
Durable Event Storage
Consumer Recovery
Replay Where Appropriate
Consumer Lag Monitoring
Independent Consumer Scaling
```

A lightweight queue alone may not provide the event-streaming capabilities required by the evolving architecture.

At the same time, adopting a large distributed infrastructure stack without a clear workload justification would increase cost and operational burden.

---

# 3. Decision

RideForge will use a:

> **Kafka-compatible event-streaming architecture, with Redpanda as the selected event-streaming platform.**

The architecture will preserve Kafka-compatible concepts and interfaces where practical.

Conceptually:

```text
RideForge Services
       ↓
Messaging Interfaces
       ↓
Kafka-Compatible Producer / Consumer
       ↓
Redpanda
       ↓
Topics / Partitions
       ↓
Consumer Groups
```

Redpanda is the current selected event-streaming infrastructure for RideForge.

---

# 4. Why Kafka-Compatible Streaming

The architecture requires capabilities associated with durable event streaming:

```text
Topics
Partitions
Consumer Groups
Durable Records
Offset Management
Retention
Replay
Horizontal Consumption
```

Kafka-compatible infrastructure provides these concepts and a mature ecosystem for event-driven systems.

---

# 5. Why Redpanda

RideForge selected Redpanda as the current event-streaming platform because it provides a Kafka-compatible event-streaming model while fitting the project's operational and cost objectives.

The decision prioritizes:

```text
Kafka Compatibility
Streaming Capabilities
Operational Simplicity
Resource Efficiency
Production Suitability
Future Ecosystem Compatibility
```

The selection is an architectural choice for the current platform and is not intended to claim that Redpanda is universally superior to Kafka for every workload.

---

# 6. Kafka Compatibility

RideForge will use Kafka-compatible concepts and protocols where practical.

This provides compatibility with:

```text
Kafka-Compatible Clients
Kafka-Compatible Tooling
Kafka-Compatible Event Concepts
```

The application should avoid unnecessary dependence on implementation-specific broker behaviour.

---

# 7. Application-Level Abstraction

RideForge application code should communicate through explicit messaging abstractions where appropriate.

Conceptually:

```text
Application
    ↓
Event / Messaging Contract
    ↓
Producer / Consumer Adapter
    ↓
Redpanda
```

The domain should not directly contain broker-specific code.

---

# 8. Producer Responsibility

A producer is responsible for publishing events according to an explicit event contract.

A producer should define:

```text
Event Type
Event Payload
Event Key
Event Metadata
Delivery Requirements
```

The producer should not assume that consumers will process the event immediately.

---

# 9. Consumer Responsibility

A consumer is responsible for:

```text
Receiving Events
Validating Events
Processing Events
Maintaining Required State
Handling Failures
Maintaining Idempotency
Reporting Operational State
```

Consumers should be independently scalable where appropriate.

---

# 10. Topic Strategy

RideForge will organize events into meaningful topics.

Examples include:

```text
ride.events
match.events
```

Additional topics should be introduced according to actual domain and operational requirements.

Avoid creating a single topic containing unrelated event categories without a clear reason.

---

# 11. Topic Ownership

Topic ownership should be explicit.

For example:

```text
Ride Domain
    ↓
ride.events
```

The owning domain is responsible for the semantic contract of the events it publishes.

Consumers do not become owners merely by subscribing.

---

# 12. Topic Naming

Topic names should be:

```text
Stable
Meaningful
Predictable
Domain-Oriented
```

The existing RideForge naming convention uses a pattern such as:

```text
<domain>.events
```

Examples:

```text
ride.events
match.events
```

Any deviation should have a clear reason.

---

# 13. Event Types Within Topics

A topic may contain multiple related event types where doing so provides a coherent event stream.

For example:

```text
ride.events

    RideCreated
    RideCancelled
    RideStarted
    RideCompleted
```

The event type must remain explicit in the event envelope.

---

# 14. Partitioning

Partitioning should support both:

```text
Throughput
```

and:

```text
Ordering Requirements
```

The partition key should be selected according to the aggregate or entity whose event ordering matters.

---

# 15. Partition Key

For event streams where ordering matters for a ride, a suitable key may be:

```text
Ride ID
```

For another workload, the correct key may instead be:

```text
Driver ID
Aggregate ID
Region
```

The key must be selected based on actual domain semantics.

---

# 16. Ordering

RideForge should only depend on ordering guarantees that are explicitly provided by the chosen partitioning strategy.

The architecture does not assume:

```text
Global Event Ordering
```

unless specifically implemented and justified.

---

# 17. Per-Aggregate Ordering

Where a domain requires ordered events for an aggregate:

```text
Same Aggregate
      ↓
Same Partition Key
      ↓
Ordered Processing
```

may be used.

For example:

```text
RideCreated
RideMatched
DriverAssigned
RideStarted
RideCompleted
```

may require ordered processing for the same ride.

---

# 18. Consumer Groups

Consumer groups provide independent processing of the same event stream.

Conceptually:

```text
ride.events
     │
     ├── Matching Consumer Group
     │
     ├── Notification Consumer Group
     │
     ├── Analytics Consumer Group
     │
     └── AI Consumer Group
```

Each group can maintain its own progress.

---

# 19. Consumer Group Independence

A slow consumer group should not unnecessarily prevent unrelated consumer groups from processing events.

This allows:

```text
Matching
Notification
Analytics
AI
```

to evolve independently.

---

# 20. Offset Management

Consumers must maintain reliable progress through the event stream.

Offset handling should support:

```text
Consumer Recovery
Restart
Retry
Controlled Replay
```

The application should not assume that an event is processed successfully merely because it was fetched.

---

# 21. Processing and Offset Commit

Where business processing must be completed before acknowledging progress, the consumer implementation should ensure that offset handling does not cause premature loss of an unprocessed event.

Conceptually:

```text
Fetch Event
    ↓
Process Event
    ↓
Successful Processing
    ↓
Commit Progress
```

The exact implementation depends on the messaging client and consumer architecture.

---

# 22. At-Least-Once Processing

RideForge consumers should be designed with duplicate processing in mind.

The architecture therefore favours:

```text
At-Least-Once Delivery
+
Idempotent Business Processing
```

rather than relying on infrastructure-level exactly-once semantics as the sole protection against duplicate business effects.

---

# 23. Idempotency

Consumers that create or modify business state must be idempotent where duplicate delivery can occur.

Possible mechanisms include:

```text
Event ID
Idempotency Key
Unique Constraint
Processed Event Record
State Validation
```

The platform-wide idempotency strategy is addressed by a separate ADR.

---

# 24. Event Persistence

Events should remain available according to the configured retention policy.

Retention should support the operational requirements for:

```text
Recovery
Consumer Restart
Debugging
Replay Where Supported
```

Retention should not be unlimited by default.

---

# 25. Event Retention

Retention should be determined by:

```text
Business Requirements
Recovery Requirements
Replay Requirements
Storage Cost
Privacy Requirements
Operational Requirements
```

Different topics may require different retention policies.

---

# 26. Replay

The event-streaming platform may support replaying historical events where appropriate.

Replay must be treated as an operational procedure rather than an ordinary consumer operation.

Before replaying events, evaluate:

```text
Idempotency
Side Effects
Ordering
Current State
Event Version
Consumer Behaviour
```

---

# 27. Replay Safety

Never assume that replaying an event is harmless.

A replay may trigger:

```text
Notifications
External API Calls
Payment Actions
State Changes
AI Processing
```

Therefore replay procedures must be controlled.

---

# 28. Dead Letter Handling

Failed events may eventually be moved to a dead-letter stream according to the retry policy.

The related architectural decision is:

```text
ADR-0013 — Dead Letter Queue Strategy
```

The event-streaming infrastructure must support the operational requirements for failure isolation.

---

# 29. Retry

Retry should be applied to recoverable failures.

Examples:

```text
Temporary Database Failure
Temporary Network Failure
Temporary External Provider Failure
```

Retry should not be used indefinitely for permanent errors.

---

# 30. Backoff

Repeated retries should use appropriate backoff to avoid:

```text
Consumer Saturation
Dependency Overload
Retry Storms
```

---

# 31. Poison Events

An event that repeatedly fails because of a permanent issue should be isolated.

Examples:

```text
Invalid Schema
Unsupported Event Version
Invalid Business Data
Programming Error
```

Such events should not remain in an infinite retry loop.

---

# 32. Event Envelope

RideForge events should use a consistent event envelope where appropriate.

A conceptual envelope may contain:

```text
Event ID
Event Type
Event Version
Occurred At
Producer
Aggregate ID
Correlation ID
Causation ID
Payload
```

The actual implementation remains governed by the event and messaging development documentation.

---

# 33. Event Payload

Payloads should contain the information required by the intended consumers.

Avoid:

```text
Entire Database Row
Internal Infrastructure Details
Unnecessary Personal Data
Secrets
```

---

# 34. Event Schema

Event schemas should be explicit enough to support:

```text
Validation
Compatibility
Consumer Development
Debugging
Versioning
```

---

# 35. Schema Evolution

Event schemas must evolve carefully.

Before changing an event:

```text
Identify Consumers
      ↓
Assess Compatibility
      ↓
Choose Migration Strategy
      ↓
Deploy
      ↓
Monitor
```

---

# 36. Breaking Event Changes

Breaking changes should not be introduced silently.

Depending on the situation, use:

```text
New Event Version
Migration Strategy
Compatibility Layer
Consumer Migration
```

The appropriate approach depends on the impact of the change.

---

# 37. Producer Failure

A producer must avoid silently losing business-critical events.

For transactional domain events, RideForge will use the Outbox pattern where appropriate.

Conceptually:

```text
Database Transaction
      │
      ├── Domain State
      │
      └── Outbox Record
                ↓
          Outbox Publisher
                ↓
             Redpanda
```

The detailed Outbox decision is recorded in ADR-0012.

---

# 38. Outbox and Redpanda

The event-streaming platform is not treated as part of the same database transaction.

The Outbox pattern bridges:

```text
Transactional Database
```

and:

```text
Event Broker
```

to provide reliable publication.

---

# 39. Broker Unavailability

If Redpanda becomes temporarily unavailable:

```text
Local Transaction
      ↓
Outbox Record
      ↓
Publisher Retry
      ↓
Redpanda Recovery
      ↓
Event Published
```

This prevents a temporary broker outage from automatically causing loss of the associated transactional event.

---

# 40. Broker Recovery

After broker recovery, pending events should be published according to the configured retry and publisher behaviour.

The system should expose enough observability to identify:

```text
Pending Outbox Records
Publication Failure
Publication Lag
Recovery Progress
```

---

# 41. Consumer Failure

If a consumer fails:

```text
Consumer Failure
      ↓
Retry / Recovery
      ↓
Continue Processing
```

The failure should not unnecessarily stop unrelated consumer groups.

---

# 42. Consumer Restart

A consumer restart should allow processing to resume from its persisted progress.

Consumers must be designed so that a restart does not cause incorrect duplicate business effects.

---

# 43. Broker Failure Isolation

A broker outage should be treated as an infrastructure failure with controlled degradation.

Critical transactional workflows should use the Outbox pattern where applicable.

Non-critical asynchronous features may be temporarily delayed.

---

# 44. Event Lag

Consumer lag should be monitored.

High lag may indicate:

```text
Insufficient Consumer Capacity
Slow Processing
Dependency Failure
Broker Pressure
Partition Imbalance
```

---

# 45. Scaling Consumers

Consumer groups should scale horizontally where the workload requires it.

Scaling is constrained by:

```text
Partition Count
Processing Characteristics
Ordering Requirements
Downstream Capacity
```

---

# 46. Partition Count

Partition count should be selected based on:

```text
Expected Throughput
Consumer Parallelism
Ordering Requirements
Future Growth
Operational Cost
```

Do not create excessive partitions without a workload justification.

---

# 47. Hot Partitions

Poor partition-key selection can concentrate traffic into a small number of partitions.

Potential hot keys should be identified through:

```text
Traffic Analysis
Partition Metrics
Load Testing
Production Monitoring
```

---

# 48. Consumer Backpressure

Consumers must handle cases where event production exceeds processing capacity.

Possible controls include:

```text
Consumer Scaling
Batch Processing
Backoff
Concurrency Limits
Rate Limiting
Downstream Protection
```

---

# 49. Event Throughput

The platform should be capacity-tested for expected RideForge workloads.

Important measurements include:

```text
Events / Second
Consumer Throughput
End-to-End Latency
Broker Latency
Consumer Lag
Processing Time
```

---

# 50. Latency

For latency-sensitive workflows, measure:

```text
Event Creation
      ↓
Broker Publication
      ↓
Consumer Fetch
      ↓
Business Processing
      ↓
Result
```

Optimizing only broker latency may not improve total workflow latency.

---

# 51. Event Ordering vs Throughput

Strong ordering requirements may reduce parallelism.

The architecture should balance:

```text
Ordering
+
Throughput
+
Scalability
```

rather than maximizing one without considering the others.

---

# 52. Event Size

Event payloads should remain reasonably small.

Large payloads increase:

```text
Network Usage
Storage
Serialization Cost
Consumer Processing
Broker Load
```

If a consumer requires large data, consider retrieving it through an appropriate read interface rather than embedding everything into the event.

---

# 53. Event Compression

Compression may be used when event volume and payload size justify it.

The decision should consider:

```text
CPU Cost
Network Savings
Storage Savings
Latency
```

Compression should not be introduced without measurable value.

---

# 54. Batch Processing

Consumers may batch processing when appropriate.

Batching can improve:

```text
Database Throughput
Network Efficiency
Consumer Efficiency
```

but may increase:

```text
Processing Latency
Failure Scope
```

---

# 55. Database Consumers

Consumers that write to PostgreSQL should use safe transaction boundaries.

A conceptual flow is:

```text
Event
 ↓
Validate
 ↓
Begin Local Transaction
 ↓
Apply State Change
 ↓
Commit
 ↓
Commit Consumer Progress
```

The exact sequence depends on the consumer implementation and failure model.

---

# 56. Redis Consumers

Consumers interacting with Redis should distinguish between:

```text
Ephemeral State
```

and:

```text
Durable Domain State
```

Redis should not become an accidental source of truth merely because a consumer can write to it.

---

# 57. AI Consumers

AI pipelines may consume events for:

```text
Feature Generation
Training Data
Feedback
Prediction
Demand / Supply Signals
```

AI consumers should use separate capacity and failure controls where required so that ML workloads do not overwhelm critical operational consumers.

---

# 58. Analytics Consumers

Analytics consumers should be isolated from critical operational workflows.

A slow analytics pipeline should not prevent:

```text
Matching
Dispatch
Ride Lifecycle
```

from operating.

---

# 59. Notification Consumers

Notification consumers are generally asynchronous.

They should support:

```text
Retry
Provider Failure Handling
Idempotency
Delivery Observability
```

---

# 60. Payment Consumers

Payment-related consumers require stronger protection against duplicate effects.

They should use:

```text
Idempotency
State Validation
Provider Correlation
Auditability
```

according to the payment domain design.

---

# 61. Event Security

Messaging access must follow least privilege.

Producers and consumers should receive only the access required for their responsibilities.

---

# 62. Topic Access

Access should be controlled by:

```text
Producer Permissions
Consumer Permissions
Topic Permissions
Environment Isolation
```

where supported by the deployment environment.

---

# 63. Environment Isolation

Development, testing, staging, and production event infrastructure should not unintentionally share operational event streams.

Use appropriate isolation through:

```text
Separate Clusters
Separate Namespaces
Separate Topics
Separate Credentials
```

according to the deployment architecture.

---

# 64. Secrets

Messaging credentials and certificates must not be stored in:

```text
Source Code
Git
Event Payloads
Logs
```

They should be provided through the project's configuration and secret-management mechanisms.

---

# 65. Observability

Redpanda and RideForge consumers should expose sufficient telemetry to understand:

```text
Broker Health
Topic Activity
Producer Errors
Consumer Errors
Consumer Lag
Processing Latency
Retry Count
DLQ Count
```

---

# 66. Logging

Structured logs should include relevant context such as:

```text
Event ID
Event Type
Topic
Partition
Offset
Consumer Group
Correlation ID
Processing Result
```

Sensitive payload data should not be logged indiscriminately.

---

# 67. Metrics

Important metrics include:

```text
Publish Rate
Consume Rate
Publish Errors
Consume Errors
Consumer Lag
Processing Latency
Retry Count
DLQ Count
Outbox Backlog
```

---

# 68. Tracing

Where distributed tracing is enabled, event processing should preserve correlation information across:

```text
Producer
Broker
Consumer
Downstream Dependency
```

where technically supported.

---

# 69. Health Checks

Applications using Redpanda should expose health information appropriate to their role.

Health checks should distinguish between:

```text
Application Healthy
Broker Dependency Healthy
Ready to Process Events
```

where necessary.

---

# 70. Local Development

Local development should provide a practical Kafka-compatible event environment.

The preferred local setup should allow developers to run:

```text
Redpanda
PostgreSQL
Redis
RideForge Services
```

as required by the workflow being developed.

---

# 71. Local Topic Creation

Local topic setup should be deterministic.

Development infrastructure should be able to create required topics consistently rather than requiring every developer to create them manually.

---

# 72. Local Event Inspection

Developers should be able to inspect:

```text
Topics
Events
Consumer Groups
Offsets
Lag
```

using appropriate local tooling.

---

# 73. Testing Environment

Integration tests involving event streaming should use a real Kafka-compatible broker where the test validates:

```text
Serialization
Partitioning
Consumer Behaviour
Offset Handling
Retry
Integration
```

Unit tests should not require a real broker when testing isolated business logic.

---

# 74. Contract Testing

Producer and consumer contracts should be validated independently where appropriate.

Contract testing should identify:

```text
Breaking Event Changes
Invalid Schemas
Missing Fields
Unexpected Versions
```

---

# 75. Failure Testing

The event infrastructure should be tested under scenarios such as:

```text
Broker Unavailable
Producer Restart
Consumer Restart
Duplicate Event
Malformed Event
Consumer Dependency Failure
Event Backlog
High Consumer Lag
```

---

# 76. Load Testing

Load testing should evaluate:

```text
Peak Event Rate
Consumer Scaling
Partition Distribution
End-to-End Latency
Broker Resource Usage
Consumer Lag
Recovery After Backlog
```

---

# 77. Disaster Recovery

Recovery planning should account for:

```text
Broker Failure
Data Loss Scenario
Consumer Recovery
Event Replay
Outbox Recovery
Configuration Recovery
```

The exact disaster-recovery architecture is governed by deployment and operations documentation.

---

# 78. Operational Runbooks

Production event-streaming operations should have runbooks for:

```text
Consumer Lag
Broker Unavailability
Failed Publication
DLQ Growth
Consumer Crash
Partition Problems
Event Replay
Outbox Backlog
```

---

# 79. Event Replay Procedure

A controlled replay should generally follow:

```text
Identify Event Range
      ↓
Identify Consumer
      ↓
Validate Idempotency
      ↓
Assess Side Effects
      ↓
Select Replay Environment / Scope
      ↓
Execute Controlled Replay
      ↓
Monitor
      ↓
Validate Result
```

---

# 80. Cost Management

Redpanda infrastructure costs should be monitored through:

```text
Compute
Memory
Storage
Network
Retention
Operational Overhead
```

Event retention and topic growth should be reviewed periodically.

---

# 81. Cost Optimization

Potential optimizations include:

```text
Appropriate Retention
Right-Sized Resources
Efficient Event Payloads
Controlled Topic Count
Consumer Scaling
Batch Processing
```

Optimization should not compromise required reliability or recovery capabilities.

---

# 82. Migration Considerations

The application should avoid unnecessary coupling to Redpanda-specific features where Kafka-compatible behaviour is sufficient.

This keeps future migration options open.

Potential future migration:

```text
Redpanda
   ↓
Another Kafka-Compatible Platform
```

should be possible without rewriting domain logic.

---

# 83. What This Decision Does Not Mean

This ADR does not mean:

```text
Every Operation Uses Events
```

It does not mean:

```text
Every Module Is a Microservice
```

It does not mean:

```text
Kafka Semantics Must Be Used Everywhere
```

It does not mean:

```text
Eventual Consistency Is Always Acceptable
```

It establishes the event-streaming infrastructure for workflows where event-driven architecture is appropriate.

---

# 84. Alternatives Considered

## 84.1 Redis Streams

Redis Streams were considered because RideForge already uses Redis for low-latency workloads and Redis Streams can provide:

```text
Streams
Consumer Groups
Persistence
Simple Operations
```

### Advantages

```text
Existing Redis Familiarity
Simple Operational Model
Potentially Lower Infrastructure Complexity
```

### Disadvantages

```text
Different Streaming Ecosystem
Less Alignment With Kafka-Oriented Tooling
Less Natural Fit for a Long-Term Event-Streaming Backbone
```

### Decision

Redis remains appropriate for:

```text
Caching
Ephemeral State
Real-Time Workloads
```

but Redpanda is selected as the primary event-streaming platform.

---

# 85. RabbitMQ

RabbitMQ was considered as a mature message-broker option.

### Advantages

```text
Strong Queueing Model
Routing Features
Mature Ecosystem
```

### Disadvantages

```text
Less Aligned With Long-Lived Event-Streaming Requirements
Different Operational Model
Less Natural Fit for Kafka-Compatible Event Streams
```

### Decision

RabbitMQ is not selected as the primary event-streaming platform.

This does not prevent future use for a specialized workload if a separate architectural decision justifies it.

---

# 86. Apache Kafka

Apache Kafka was considered as the reference Kafka-compatible event-streaming platform.

### Advantages

```text
Mature Ecosystem
Large Community
Broad Tooling
Strong Streaming Model
```

### Disadvantages

```text
Higher Operational Complexity in Some Deployment Models
Potentially Higher Infrastructure Overhead
```

### Decision

RideForge adopts the Kafka-compatible model while selecting Redpanda as the current event-streaming implementation.

---

# 87. Direct Database Polling

Database polling was considered as an alternative to event streaming.

### Advantages

```text
Simple Initial Implementation
No Broker Required
```

### Disadvantages

```text
Polling Load
Latency
Coupling
Poor Event Semantics
Scaling Challenges
```

### Decision

Database polling is not the primary integration mechanism for event-driven workflows.

The Outbox pattern and event broker provide a more explicit asynchronous architecture.

---

# 88. Synchronous Service Calls Only

A synchronous-only architecture was considered.

### Advantages

```text
Simple Request Flow
Immediate Responses
Easy Initial Reasoning
```

### Disadvantages

```text
Tight Coupling
Failure Propagation
Long Request Chains
Poor Background Processing Model
Reduced Consumer Independence
```

### Decision

RideForge retains synchronous APIs but uses asynchronous events where they provide clear architectural value.

---

# 89. Decision Drivers

The decision is primarily driven by:

```text
Event Durability
Consumer Groups
Partitioning
Scalability
Replay Capability
Operational Recovery
Kafka Compatibility
Cost Awareness
Future Evolution
Event-Driven Architecture
```

---

# 90. Consequences

## 90.1 Positive Consequences

The decision provides:

```text
Durable Event Streaming
Independent Consumers
Horizontal Consumer Scaling
Partition-Based Ordering
Event Retention
Replay Capability
Kafka-Compatible Ecosystem
Clear Event Infrastructure
```

---

## 90.2 Negative Consequences

The platform introduces:

```text
Broker Operations
Messaging Configuration
Consumer Complexity
Eventual Consistency
Partition Management
Offset Management
Retry Handling
DLQ Handling
Monitoring Requirements
```

These costs are accepted because event streaming is a core architectural capability for the evolving RideForge platform.

---

# 91. Risks

## Risk 1 — Overusing the Broker

### Mitigation

Use events only where asynchronous communication provides meaningful value.

---

## Risk 2 — Broker Operational Complexity

### Mitigation

Use standardized infrastructure, automation, observability, and runbooks.

---

## Risk 3 — Consumer Lag

### Mitigation

Monitor lag and provide appropriate scaling and backpressure controls.

---

## Risk 4 — Duplicate Processing

### Mitigation

Require idempotent consumer behaviour.

---

## Risk 5 — Event Contract Breakage

### Mitigation

Use explicit contracts, compatibility testing, and versioning.

---

## Risk 6 — Cost Growth

### Mitigation

Monitor:

```text
Retention
Storage
Traffic
Partitions
Consumer Capacity
```

and optimize based on actual workload.

---

## Risk 7 — Redpanda-Specific Coupling

### Mitigation

Keep application contracts and business logic independent from broker-specific implementation details.

---

# 92. Validation

The decision should be validated through:

```text
Integration Tests
Contract Tests
Load Tests
Failure Tests
Consumer Recovery Tests
Replay Tests
Idempotency Tests
Operational Metrics
Cost Monitoring
```

---

# 93. Review Triggers

Revisit this ADR when:

```text
Event Volume Changes Significantly
Redpanda No Longer Meets Requirements
Infrastructure Cost Becomes Material
A Major Messaging Workload Is Added
Kafka Compatibility Requirements Change
A New Messaging Platform Is Proposed
Operational Failures Reveal Architectural Problems
```

---

# 94. Related Documentation

This ADR should be read together with:

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
Integration Testing and Local Infrastructure
Logging and Debugging
Observability Development
Performance and Optimization
AI Machine Learning Data Pipeline
AI Feedback and Learning Loop
AI Failure and Fallback Strategy
```

---

# 95. Related ADRs

This decision is directly related to:

```text
ADR-0002 — Architecture Style
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
ADR-0005 — Event-Driven Architecture
ADR-0012 — Outbox Pattern
ADR-0013 — Dead Letter Queue Strategy
ADR-0014 — API and Service Communication
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
ADR-0028 — Cost Optimization Strategy
```

---

# 96. Decision Summary

RideForge will use:

```text
Kafka-Compatible Event Streaming
            +
Redpanda
```

as the primary asynchronous event-streaming infrastructure.

The architecture will use:

```text
Topics
+
Partitions
+
Consumer Groups
+
Durable Events
+
Offset Management
+
Retention
+
Controlled Replay
```

together with:

```text
Outbox
+
Idempotent Consumers
+
Retry
+
DLQ
+
Observability
```

to provide a reliable event-driven platform.

---

# 97. Final Principle

> **Redpanda is an infrastructure choice that supports RideForge's event-driven architecture; it must remain an implementation detail beneath stable application and domain contracts.**

The desired architecture is:

```text
Domain
   ↓
Event Contract
   ↓
Messaging Adapter
   ↓
Redpanda
   ↓
Consumer Groups
   ↓
Independent Processing
```

while maintaining:

```text
Reliable Publication
+
Idempotent Processing
+
Controlled Failure
+
Observable Operations
+
Cost Discipline
```

---

# 98. Status

```text
Decision: ACCEPTED

Event Streaming Platform:
Kafka-Compatible Architecture
+
Redpanda
```

This decision establishes Redpanda as the current RideForge event-streaming infrastructure and provides the foundation for subsequent messaging reliability, API communication, consistency, and operational ADRs.
