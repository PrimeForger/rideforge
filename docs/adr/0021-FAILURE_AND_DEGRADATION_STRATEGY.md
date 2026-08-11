# ADR-0021: Failure and Degradation Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Reliability / Resilience / Distributed Systems  
> **Scope:** Service failures, dependency failures, timeouts, retries, degradation, fallback, recovery, and operational resilience  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is a distributed ride-hailing platform composed of multiple services and infrastructure components.

The platform depends on:

```text
API Services
PostgreSQL
PgBouncer
Redis
Kafka / Redpanda
Location Infrastructure
Routing / ETA Providers
Dispatch Services
AI Services
Payment Providers
Notification Providers
Cloud Infrastructure
```

A production ride-hailing platform cannot assume that every dependency will always be:

```text
Available
Fast
Correct
Reachable
Healthy
```

Failures can occur at any layer.

Examples:

```text
Database unavailable
Redis unavailable
Kafka unavailable
Consumer crashed
Route provider timeout
AI inference timeout
Payment provider unavailable
Driver location stream delayed
Network partition
Connection pool exhaustion
Service overload
Deployment failure
```

RideForge therefore needs an explicit strategy for:

```text
Detecting failures
Containing failures
Retrying safely
Degrading functionality
Using fallbacks
Protecting critical state
Recovering automatically
Recovering manually when required
```

---

# 2. Problem

Without a unified failure strategy, individual services may implement inconsistent behaviour such as:

```text
Infinite Retries
Long Timeouts
Retry Storms
Cascading Failures
Unsafe Fallbacks
Silent Data Loss
Incorrect Ride State
Duplicate Business Effects
```

A failure in one dependency should not automatically cause the entire platform to fail.

For example:

```text
AI Service Down
```

must not necessarily mean:

```text
Ride Booking Down
```

Similarly:

```text
ETA Provider Down
```

must not necessarily mean:

```text
Ride Creation Down
```

The platform therefore requires explicit failure boundaries and degradation modes.

---

# 3. Decision

RideForge will use a layered failure and degradation strategy based on:

```text
Failure Classification
+
Timeouts
+
Bounded Retries
+
Exponential Backoff
+
Jitter
+
Circuit Breaking Where Appropriate
+
Bulkheads
+
Rate Limiting
+
Fallbacks
+
Graceful Degradation
+
Idempotency
+
Eventual Recovery
+
Observability
```

The core principle is:

> **A failure should be contained within the smallest possible boundary, while critical ride and financial state remains protected.**

---

# 4. Failure Classification

Failures should be classified before deciding how to respond.

Primary categories:

```text
Transient Failure
Permanent Failure
Dependency Failure
Application Failure
Infrastructure Failure
Data Failure
Capacity Failure
Unknown Outcome
```

---

# 5. Transient Failure

A transient failure may succeed if retried later.

Examples:

```text
Temporary Network Failure
Database Connection Reset
Temporary Kafka Failure
Temporary Provider 5xx
Temporary Resource Contention
```

These may be retryable.

---

# 6. Permanent Failure

A permanent failure is unlikely to succeed without changing the request or state.

Examples:

```text
Invalid Input
Unauthorized Request
Forbidden Operation
Invalid Ride State
Unsupported Region
Invalid Driver
```

These should normally not be retried.

---

# 7. Dependency Failure

A dependency failure occurs when another system required by a service is unavailable or unhealthy.

Examples:

```text
PostgreSQL
Redis
Kafka / Redpanda
Map Provider
Payment Provider
AI Service
Notification Provider
```

The service must determine whether the dependency is:

```text
Critical
Important
Optional
```

for the specific operation.

---

# 8. Application Failure

Application failures include:

```text
Unhandled Exception
Panic
Invalid State
Programming Error
Serialization Error
Unexpected Domain Error
```

Application failures should be observable and should not be hidden behind generic retries.

---

# 9. Infrastructure Failure

Examples:

```text
Node Failure
Container Failure
Network Failure
Load Balancer Failure
Disk Failure
Availability Zone Failure
Cloud Service Failure
```

The architecture should limit the blast radius of such failures.

---

# 10. Capacity Failure

Capacity failures occur when the system is overloaded.

Examples:

```text
CPU Saturation
Memory Pressure
Database Connection Exhaustion
Kafka Consumer Lag
Redis Saturation
Request Queue Growth
```

Capacity failures require load shedding or scaling rather than blind retries.

---

# 11. Unknown Outcome

An unknown outcome occurs when the client does not know whether the operation succeeded.

Example:

```text
Request Sent
      ↓
Server Processes
      ↓
Network Connection Lost
      ↓
Client Receives No Response
```

The operation may have already committed.

Therefore:

```text
No Response
≠
No Operation
```

This is governed by:

```text
ADR-0020 — Idempotency Strategy
```

---

# 12. Critical vs Non-Critical Dependencies

Each dependency must be classified according to the operation.

Example:

| Dependency | Ride Creation | Dispatch | Analytics |
|---|---:|---:|---:|
| PostgreSQL | Critical | Critical | Important |
| Redis | Important | Critical/Important | Optional |
| Kafka | Important | Important | Important |
| AI Service | Optional | Optional/Important | Optional |
| Route Provider | Important | Important | Optional |
| Notification Provider | Optional | Optional | Optional |
| Payment Provider | Not always required | Not required | Not required |

The exact classification depends on the workflow.

---

# 13. Critical Dependency

A critical dependency means the operation cannot safely proceed without it.

Example:

```text
PostgreSQL
```

for a transaction that must durably change authoritative ride state.

The system should fail safely rather than inventing state.

---

# 14. Optional Dependency

An optional dependency can fail while the core operation continues.

Examples:

```text
AI Ranking
Analytics
Non-Critical Notification
Recommendation
```

---

# 15. Graceful Degradation

When an optional dependency fails:

```text
Primary Function
→ Continue
```

while:

```text
Optional Feature
→ Disabled / Simplified
```

Example:

```text
AI Smart Dispatch Unavailable
        ↓
Fallback to Deterministic Dispatch
```

---

# 16. Degradation Must Be Explicit

Every important dependency should define:

```text
Healthy Mode
Degraded Mode
Unavailable Mode
Recovery Mode
```

Do not rely on accidental behaviour.

---

# 17. Failure Boundary

A dependency failure should be contained as close to the dependency boundary as possible.

Conceptually:

```text
Service
  ↓
Dependency Adapter
  ↓
Failure Handling
  ↓
Fallback / Error
```

Business logic should not be filled with provider-specific failure handling.

---

# 18. Dependency Adapters

External providers should be accessed through explicit interfaces/adapters.

Examples:

```text
RouteProvider
PaymentProvider
NotificationProvider
AIProvider
LocationProvider
```

This allows:

```text
Provider Failure
→ Adapter Handles
→ Application Receives Stable Result
```

---

# 19. Timeout Strategy

Every network dependency must have a bounded timeout.

Avoid:

```text
Infinite Wait
```

---

# 20. Timeout Layers

Timeouts may exist at:

```text
HTTP Client
Database Query
Redis Operation
Kafka Operation
External Provider
Application Request
```

Timeouts must be coordinated rather than arbitrarily stacked.

---

# 21. Timeout Budget

A request should have a total latency budget.

Conceptually:

```text
Total Request Budget
        ↓
Dependency A
+
Dependency B
+
Database
+
Application Processing
```

The sum should remain within the API's expected response time.

---

# 22. Timeout Propagation

Request context deadlines should propagate to downstream operations where appropriate.

Conceptually:

```text
Client
 ↓
API
 ↓
Application
 ↓
Dependency
```

---

# 23. Avoid Timeout Multiplication

If:

```text
API Timeout = 5s
```

do not configure:

```text
Provider Timeout = 30s
```

inside the same synchronous request.

The provider should have a timeout compatible with the remaining request budget.

---

# 24. Database Timeouts

Database operations should have bounded execution time.

Long-running queries can cause:

```text
Connection Pool Exhaustion
Lock Contention
Request Queue Growth
```

---

# 25. Redis Timeouts

Redis calls must also have bounded timeouts.

A Redis outage should not cause application goroutines to wait indefinitely.

---

# 26. Kafka Timeouts

Kafka/Redpanda operations should use bounded network and processing timeouts appropriate to the producer/consumer role.

---

# 27. External Provider Timeouts

Route, payment, notification, and AI provider calls must use bounded timeouts.

---

# 28. Retry Strategy

Retries should only be used when:

```text
Failure Is Potentially Transient
+
Operation Is Safe to Retry
```

---

# 29. Retry Is Not a Default

Do not automatically retry:

```text
Every Error
```

or:

```text
Every HTTP 5xx
```

without considering operation semantics.

---

# 30. Retryable Errors

Potential retry candidates include:

```text
Network Timeout
Connection Reset
Temporary 5xx
Database Deadlock
Serialization Failure
Temporary Kafka Failure
Temporary Provider Unavailability
```

---

# 31. Non-Retryable Errors

Examples:

```text
Validation Error
Authentication Failure
Authorization Failure
Invalid State
Unsupported Region
Invalid Request
Business Rule Violation
Idempotency Conflict
```

---

# 32. Exponential Backoff

Retries should normally use:

```text
Exponential Backoff
```

Example:

```text
Attempt 1 → Short Delay
Attempt 2 → Longer Delay
Attempt 3 → Longer Delay
```

---

# 33. Jitter

Add jitter to prevent many workers from retrying simultaneously.

Conceptually:

```text
Backoff
+
Randomized Jitter
```

---

# 34. Retry Limit

Retries must be bounded.

Never implement:

```text
while failure:
    retry forever
```

---

# 35. Retry Budget

Services should have an operational retry budget.

Excessive retries can consume more resources than the original failure.

---

# 36. Retry Storm

A retry storm occurs when:

```text
Dependency Fails
      ↓
Requests Retry
      ↓
Load Increases
      ↓
Dependency Becomes More Unhealthy
      ↓
More Retries
```

This creates a feedback loop.

---

# 37. Retry Storm Prevention

Use:

```text
Bounded Retries
Backoff
Jitter
Circuit Breakers
Rate Limits
Load Shedding
```

where appropriate.

---

# 38. Circuit Breaker

A circuit breaker may be used for unstable remote dependencies.

Conceptually:

```text
CLOSED
  ↓
Failures Increase
  ↓
OPEN
  ↓
Wait
  ↓
HALF-OPEN
  ↓
Test Request
  ↓
CLOSED / OPEN
```

---

# 39. Circuit Breaker Purpose

The circuit breaker prevents repeatedly calling a dependency that is already failing.

---

# 40. Circuit Breaker Is Not a Fix

A circuit breaker does not solve the underlying failure.

It provides:

```text
Failure Containment
```

and:

```text
Resource Protection
```

---

# 41. Circuit Breaker Scope

Circuit breakers should normally be scoped to:

```text
Dependency
+
Operation
```

rather than globally disabling unrelated functionality.

---

# 42. Circuit Breaker Example

If:

```text
Route Provider
```

is failing:

```text
Route Provider Circuit
→ OPEN
```

but:

```text
Ride Creation
→ May Continue
```

if an acceptable fallback exists.

---

# 43. Bulkheads

Bulkheads isolate resources between workloads.

Examples:

```text
Separate Worker Pools
Separate Connection Pools
Separate Concurrency Limits
Separate Queues
```

---

# 44. Why Bulkheads

Without isolation:

```text
AI Requests
```

could consume all available resources and prevent:

```text
Ride Requests
```

from being processed.

---

# 45. Database Bulkhead

Where necessary, separate workloads logically so that:

```text
Analytics Queries
```

cannot exhaust resources needed for:

```text
Ride Transactions
```

---

# 46. Worker Bulkhead

Use separate worker/concurrency controls for:

```text
Critical Ride Events
AI Jobs
Analytics
Notifications
```

where operationally justified.

---

# 47. Load Shedding

When the system is overloaded, reject or defer work that is less important.

Priority should favour:

```text
Active Ride Operations
Critical State Transitions
Safety-Critical Operations
Payment State
```

over:

```text
Analytics
Non-Critical Recommendations
Optional AI Processing
```

---

# 48. Backpressure

Consumers and workers should apply backpressure rather than processing unlimited work.

---

# 49. Queue Growth

Monitor:

```text
Queue Depth
Consumer Lag
Processing Latency
Retry Queue Size
DLQ Size
```

---

# 50. Kafka / Redpanda Failure

If Kafka/Redpanda becomes unavailable:

```text
Database Transaction
+
Outbox
```

should continue to provide durable event intent where possible.

---

# 51. Kafka Recovery

After Kafka/Redpanda recovery:

```text
Outbox Publisher
→ Resume
→ Publish Pending Events
```

---

# 52. Consumer Failure

If a consumer crashes:

```text
Message Remains Available / Is Redelivered
```

according to the messaging semantics.

The consumer must be idempotent.

---

# 53. Consumer Lag

Temporary consumer lag should not automatically be treated as data loss.

The system should distinguish:

```text
Delayed
```

from:

```text
Lost
```

---

# 54. Dead Letter Queue

Messages that cannot be processed after appropriate retries may be moved to:

```text
DLQ
```

according to:

```text
ADR-0013 — Dead Letter Queue Strategy
```

---

# 55. PostgreSQL Failure

PostgreSQL is authoritative for core transactional state.

If PostgreSQL is unavailable:

```text
Do Not Invent Transactional State
```

Critical write operations should fail safely.

---

# 56. PostgreSQL Connection Exhaustion

Connection exhaustion may result from:

```text
Traffic Spike
Long Transactions
Leaked Connections
Slow Queries
Incorrect Pool Configuration
```

The system should protect the database through:

```text
Connection Limits
Timeouts
PgBouncer
Load Shedding
Query Optimization
```

---

# 57. PgBouncer

PgBouncer is part of the connection-management strategy.

The application should not attempt to solve database overload only by increasing application connection counts.

---

# 58. Database Read Degradation

Where read replicas or projections exist, non-critical reads may use them.

Critical authoritative reads should use the source of truth when required.

---

# 59. Redis Failure

Redis may support:

```text
Caching
Real-Time State
Location
Rate Limiting
Ephemeral Coordination
```

depending on the subsystem.

Redis failure must be classified by use case.

---

# 60. Cache Failure

For cache-only data:

```text
Redis Down
→ Read From Database
```

where the database fallback is operationally safe.

---

# 61. Cache Fallback

Do not use a database fallback blindly if it can cause:

```text
Database Overload
```

Use:

```text
Rate Limits
Local Cache
Request Coalescing
Load Shedding
```

where appropriate.

---

# 62. Critical Real-Time State

If Redis or another real-time state system is authoritative for a specific subsystem, its failure strategy must be explicitly defined for that subsystem.

Do not assume cache fallback is sufficient.

---

# 63. Driver Location Failure

Location data may become:

```text
Delayed
Stale
Unavailable
```

The dispatch system must not treat stale driver locations as current without validation.

---

# 64. Stale Location Handling

A driver may be marked:

```text
LOCATION_STALE
```

after a defined freshness threshold.

Such a driver may be excluded or down-ranked from dispatch.

---

# 65. Location Failure and Dispatch

Dispatch should prefer:

```text
Known Fresh Location
```

over:

```text
Unknown / Stale Location
```

---

# 66. Route Provider Failure

If the primary route provider fails:

```text
Fallback Provider
```

may be used where configured.

---

# 67. Route Provider Fallback

Fallback selection should consider:

```text
Availability
Latency
Coverage
Accuracy
Cost
Regional Support
```

---

# 68. ETA Degradation

If route-based ETA is unavailable, the system may use a simpler fallback:

```text
Historical ETA
Distance-Based Estimate
Cached Route
Regional Default
```

The fallback must be clearly marked internally as lower confidence.

---

# 69. ETA Confidence

An ETA result should carry sufficient metadata to distinguish:

```text
Provider ETA
Fallback ETA
Cached ETA
Approximate ETA
```

where the product needs this distinction.

---

# 70. AI Service Failure

AI must never become an unavoidable dependency for core ride safety and lifecycle operations unless explicitly approved.

---

# 71. AI Dispatch Failure

AI-assisted dispatch is an optimization capability within the resolved primary dispatch strategy.

If AI ranking, prediction, or inference fails or times out:

```text
Resolved Dispatch Strategy
        ↓
AI Assistance Unavailable
        ↓
Deterministic Execution of Same Strategy
```

AI failure must not automatically change the effective dispatch strategy.

For example:

```text
Smart Stand Dispatch
        ↓
AI Failure
        ↓
Deterministic Smart Stand Dispatch
```

and:

```text
Smart Dispatch
        ↓
AI Failure
        ↓
Deterministic Smart Dispatch
```

---

# 72. Smart Stand Dispatch vs Smart Dispatch

RideForge has two primary dispatch strategies:

```text
Smart Stand Dispatch
Smart Dispatch
```

AI-assisted dispatch is not a third primary strategy.

### Smart Stand Dispatch

Smart Stand Dispatch is:

```text
Stand-preferred
```

not:

```text
Stand-exclusive
```

When the rider is within a configured stand radius:

```text
Preferred Stand
      ↓
Eligible Stand Drivers
      ↓
Stand Queue / Ordering
```

If suitable stand supply is unavailable, the candidate search may expand to:

```text
Non-Stand Drivers
Nearby Stand Drivers
Drivers from Nearby Locations
```

If the rider is outside all configured stand radii, Smart Stand Dispatch must not restrict the candidate pool to stand drivers.

### Smart Dispatch

Smart Dispatch is stand-agnostic.

It may consider any eligible nearby driver regardless of stand membership.

For both strategies, failure handling must preserve:

```text
Regional Rules
Legal Validation
Driver Eligibility
Safety Constraints
Ride Constraints
```

---

# 73. AI Failure Principle

AI may optimize a dispatch decision.

AI must not remove mandatory business constraints or change the resolved primary strategy.

Conceptually:

```text
Resolved Dispatch Strategy
        ↓
Hard Constraints
        ↓
Candidate Set
        ↓
AI Ranking
        ↓
Final Strategy-Compatible Selection
```

If AI is unavailable:

```text
Resolved Dispatch Strategy
        ↓
Hard Constraints
        ↓
Deterministic Strategy-Compatible Selection
```

AI output must never authorize an otherwise invalid candidate.

---

# 74. AI Fallback

The fallback path must preserve the effective dispatch strategy.

For Smart Stand Dispatch:

```text
Smart Stand Dispatch
        ↓
AI unavailable
        ↓
Deterministic Smart Stand Dispatch
        ↓
Preferred Stand / Queue Rules
        ↓
Broader Candidate Expansion if configured
```

For Smart Dispatch:

```text
Smart Dispatch
        ↓
AI unavailable
        ↓
Deterministic Smart Dispatch
```

Broader candidate discovery is not equivalent to switching strategies.

For example:

```text
Smart Stand Dispatch
        ↓
Preferred Stand has no suitable driver
        ↓
Expand to broader eligible candidates
        ↓
Still Smart Stand Dispatch
```

A transition to another primary strategy is permitted only when an explicit business/configuration rule defines that transition.

---

# 75. AI Timeout

AI inference must have a bounded timeout.

If the timeout expires:

```text
Do Not Hold Critical Transaction
```

and:

```text
Continue with Safe Strategy-Compatible Fallback
```

where available.

AI timeout must not silently change:

```text
Smart Stand Dispatch
```

into:

```text
Smart Dispatch
```

unless explicitly configured.

---

# 76. AI Model Failure

A model failure includes:

```text
Model Unavailable
Model Timeout
Invalid Output
Unexpected Score
Feature Missing
Model Version Error
```

These should trigger controlled degradation.

---

# 77. Invalid AI Output

AI output must be validated before it affects a business decision.

Invalid output should be treated as:

```text
AI Failure
```

not trusted blindly.

---

# 78. Payment Provider Failure

Payment systems require special handling because an external request can have an unknown outcome.

Use:

```text
Idempotency
+
Provider Status Query
+
Webhook
+
Reconciliation
```

where supported.

---

# 79. Payment Must Not Be Blindly Retried

A timeout does not mean:

```text
Payment Failed
```

The system should determine the provider-side state before issuing another payment operation.

---

# 80. Notification Failure

Notification failure should generally not roll back a committed ride state transition.

Example:

```text
Ride Accepted
+
Notification Failed
```

The ride should remain:

```text
ACCEPTED
```

while notification delivery is retried asynchronously.

---

# 81. Notification Retry

Use:

```text
Outbox / Queue
+
Retry
+
Provider Idempotency
```

where applicable.

---

# 82. Critical State vs Side Effect

A useful rule is:

```text
Core State
→ Must Be Correct

Side Effect
→ May Be Delayed
```

Examples of side effects:

```text
Push Notification
Analytics
AI Feedback
Email
SMS
```

---

# 83. Failure Containment Matrix

| Component | Failure Impact | Primary Strategy |
|---|---|---|
| PostgreSQL | Critical | Fail Safe |
| PgBouncer | Critical | Pool Recovery / Failover |
| Kafka / Redpanda | Event Delay | Outbox |
| Redis Cache | Read Degradation | Cache Fallback |
| Location Store | Dispatch Degradation | Freshness Rules |
| Route Provider | ETA Degradation | Provider Fallback |
| AI Service | Dispatch Degradation | Deterministic Fallback |
| Payment Provider | Payment Uncertainty | Idempotency + Reconciliation |
| Notification Provider | Side-Effect Delay | Async Retry |
| Analytics | Delayed Analytics | Eventual Recovery |

---

# 84. Failure Severity

Failures should be classified by impact:

```text
SEV-1
SEV-2
SEV-3
SEV-4
```

The exact operational definitions should be maintained by the observability/operations documentation.

---

# 85. User-Facing Errors

Errors should be:

```text
Stable
Actionable
Non-Leaky
Consistent
```

Do not expose:

```text
Database Stack Trace
Kafka Internal Error
Provider Credentials
Internal Hostnames
```

---

# 86. Error Translation

Internal failure:

```text
DATABASE_CONNECTION_ERROR
```

may become a public response such as:

```text
SERVICE_UNAVAILABLE
```

depending on API semantics.

---

# 87. Fail Fast vs Fail Safe

Use:

```text
Fail Fast
```

when continuing would create an invalid or unsafe state.

Use:

```text
Fail Safe / Degrade
```

when the operation can safely continue with reduced functionality.

---

# 88. Example

If:

```text
Legal Ride Validation
```

fails:

```text
Fail
```

Do not allow the ride merely because the validation service is unavailable.

---

# 89. Example

If:

```text
AI Ranking
```

fails:

```text
Fallback
```

may be appropriate because deterministic dispatch can still enforce the required constraints.

---

# 90. Legal Validation

Legal/regional validation is a hard constraint.

A failure in legal validation must not be treated as:

```text
Optional Dependency
```

for a workflow that requires the validation.

This is governed by:

```text
ADR-0018 — Regional and Legal Ride Validation
```

---

# 91. Safety Constraints

Safety-related constraints must fail closed where required.

Do not convert:

```text
Unknown Safety State
```

into:

```text
Assume Safe
```

---

# 92. Dispatch Safety

Dispatch fallback must preserve:

```text
Driver Eligibility
Vehicle Eligibility
Regional Restrictions
Ride Constraints
Safety Rules
```

---

# 93. Stand Dispatch Fallback

Smart Stand Dispatch must not be treated as a generic fallback that is automatically substituted for Smart Dispatch.

The effective dispatch strategy is resolved from the applicable hierarchical configuration:

```text
Most Specific Applicable Configuration
        ↓
Explicit Strategy?
   ├── YES → Use It
   └── NO
        ↓
Parent Configuration
        ↓
Continue Upward
        ↓
System Default
```

If the resolved strategy is Smart Stand Dispatch, broader candidate discovery may occur when the preferred stand cannot provide a suitable driver.

That expansion may include:

```text
Non-Stand Drivers
Nearby Stand Drivers
Drivers from Nearby Locations
```

provided they satisfy the applicable hard constraints.

This is:

```text
Candidate Expansion
```

not:

```text
Strategy Switching
```

A strategy switch is permitted only when an explicit business/configuration rule defines it.

---

# 94. Recovery

Recovery means returning from:

```text
Degraded
```

to:

```text
Healthy
```

without creating inconsistent state.

---

# 95. Automatic Recovery

Where possible, recovery should be automatic.

Examples:

```text
Retry Outbox
Reconnect Kafka
Restart Consumer
Reopen Circuit
Refresh Provider Connection
```

---

# 96. Manual Recovery

Manual intervention may be required for:

```text
Data Corruption
Poison Messages
Payment Reconciliation
Persistent Provider Inconsistency
Operational Configuration Errors
```

---

# 97. Recovery Must Be Idempotent

Recovery actions must follow:

```text
ADR-0020 — Idempotency Strategy
```

Do not assume recovery runs only once.

---

# 98. Reconciliation

Reconciliation is required when external state can diverge from internal state.

Examples:

```text
Payment
Provider Booking
Notification Delivery
Location State
```

---

# 99. Reconciliation Principle

Use:

```text
Authoritative State
+
External State
+
Operation Identity
```

to determine the correct action.

---

# 100. Do Not Repair Blindly

A recovery worker should never blindly execute:

```text
Retry Everything
```

It should classify the state first.

---

# 101. Disaster Recovery

Disaster recovery must address:

```text
Database Failure
Message Infrastructure Failure
Region Failure
Service Failure
Configuration Loss
Data Loss
```

---

# 102. Recovery Objectives

The platform should define:

```text
RTO — Recovery Time Objective
RPO — Recovery Point Objective
```

for critical systems.

The exact targets belong in deployment/operations documentation and should be based on business requirements.

---

# 103. PostgreSQL Recovery

PostgreSQL recovery should provide:

```text
Backups
Point-in-Time Recovery
Replication / Failover
Migration Safety
```

as appropriate to deployment maturity.

---

# 104. Kafka / Redpanda Recovery

Messaging recovery should consider:

```text
Retention
Replication
Consumer Offsets
Outbox Backlog
Consumer Replay
```

---

# 105. Redis Recovery

Redis recovery depends on whether its data is:

```text
Cache
Ephemeral State
Real-Time State
Coordination State
```

Each category requires different recovery expectations.

---

# 106. Cache Recovery

Cache data may often be rebuilt from the source of truth.

---

# 107. Ephemeral State Recovery

Ephemeral state may be reconstructed or allowed to expire.

---

# 108. Authoritative Real-Time State

If a subsystem makes Redis authoritative for a particular state, recovery must explicitly define:

```text
Persistence
Rebuild
Replay
Failover
Consistency Verification
```

---

# 109. Dependency Health

Health checks should distinguish:

```text
Process Alive
```

from:

```text
Dependency Ready
```

---

# 110. Liveness

Liveness answers:

```text
Is the process alive?
```

It should not fail merely because a dependency is temporarily unavailable if restarting the process would not help.

---

# 111. Readiness

Readiness answers:

```text
Can this instance safely receive this workload?
```

Readiness may reflect critical dependency availability.

---

# 112. Dependency-Specific Readiness

A service may remain alive while becoming unready for certain operations.

Avoid unnecessarily restarting healthy processes because of temporary dependency failures.

---

# 113. Health Check Failure Loops

Poor health checks can cause:

```text
Dependency Failure
→ Service Marked Unhealthy
→ Restart
→ Dependency Still Down
→ Restart
→ More Load
```

Health checks must be designed carefully.

---

# 114. Deployment Failure

Deployments should support:

```text
Rollback
Health Verification
Gradual Rollout
```

where operationally appropriate.

---

# 115. Graceful Shutdown

Services should stop accepting new work and allow in-flight work to finish safely within a bounded shutdown period.

---

# 116. Consumer Shutdown

Consumers should:

```text
Stop New Consumption
Finish / Safely Abort Current Work
Commit Appropriate Offset
Release Resources
```

according to the consumer design.

---

# 117. HTTP Shutdown

HTTP servers should:

```text
Stop Accepting New Connections
Drain In-Flight Requests
Close Dependencies
Exit
```

within a bounded period.

---

# 118. Database Shutdown

Database connections should be released cleanly.

Open transactions must not be abandoned.

---

# 119. Failure Injection

The system should periodically test:

```text
Database Unavailable
Redis Unavailable
Kafka Unavailable
Provider Timeout
AI Timeout
Consumer Crash
Network Failure
```

---

# 120. Chaos Testing

Chaos testing should focus on:

```text
Failure Containment
Recovery
Data Integrity
User Impact
```

rather than introducing failures without measurable objectives.

---

# 121. Observability

Failure handling depends on observability.

Track:

```text
Error Rate
Timeout Rate
Retry Rate
Circuit State
Queue Depth
Consumer Lag
Dependency Latency
Fallback Rate
Degraded Mode Rate
Recovery Time
```

---

# 122. Fallback Metrics

Every meaningful fallback should be observable.

Examples:

```text
AI_FALLBACK_COUNT
ETA_FALLBACK_COUNT
ROUTE_PROVIDER_FAILOVER_COUNT
NOTIFICATION_RETRY_COUNT
```

---

# 123. Silent Degradation Is Forbidden

The system should not silently switch to a degraded mode without generating operational telemetry.

---

# 124. User Experience During Degradation

The product should distinguish:

```text
Normal
Delayed
Approximate
Unavailable
```

where this distinction materially affects user expectations.

---

# 125. ETA Degraded Experience

If ETA confidence decreases:

```text
Internal Confidence
```

should reflect it.

The product may communicate an approximate ETA rather than falsely presenting high precision.

---

# 126. Dispatch Degraded Experience

If AI is unavailable:

```text
Dispatch Continues
```

through the configured fallback.

The system should record:

```text
Fallback Reason
```

for later analysis.

---

# 127. Regional Degradation

Different operating regions may use different:

```text
Dispatch Mode
Route Provider
Fallback
Legal Validation
```

Failure handling must respect regional configuration.

---

# 128. No Global Assumption

A provider failure in:

```text
Region A
```

should not automatically disable functionality in:

```text
Region B
```

unless the dependency is globally shared and actually affected.

---

# 129. Regional Isolation

Where practical, failures should be isolated by:

```text
Region
Service
Provider
Queue
Worker Pool
```

---

# 130. Provider Failover

Provider failover should be configurable.

Do not hard-code a provider-specific fallback into core domain logic.

---

# 131. Provider Priority

A provider strategy may define:

```text
Primary
Secondary
Emergency Fallback
```

with explicit selection rules.

---

# 132. Cost vs Resilience

Fallbacks have cost implications.

For example:

```text
Primary Provider
→ Low Cost

Secondary Provider
→ Higher Cost
```

The platform should track fallback cost where material.

---

# 133. Avoid Over-Engineering Fallbacks

Not every dependency requires multiple providers.

Fallback infrastructure should be justified by:

```text
Business Criticality
Availability Requirements
Failure Frequency
Cost
Operational Complexity
```

---

# 134. Single Provider Strategy

A single provider may be acceptable when:

```text
Business Risk Is Low
Fallback Is Not Cost-Effective
Core Operation Has Another Safe Degradation
```

---

# 135. Failure Policy by Dependency

Each dependency should document:

```text
Timeout
Retry
Circuit Breaker
Fallback
Failure Mode
Recovery
Observability
```

---

# 136. Dependency Failure Template

```text
Dependency:
Criticality:
Timeout:
Retryable Errors:
Retry Limit:
Backoff:
Circuit Breaker:
Fallback:
Failure Mode:
Recovery:
Metrics:
Alerts:
```

---

# 137. Core Principle for Critical State

When authoritative state cannot be safely persisted:

```text
Reject / Defer Operation
```

rather than:

```text
Invent State
```

---

# 138. Core Principle for Optional Features

When optional functionality fails:

```text
Continue Core Operation
+
Record Degradation
```

---

# 139. Failure Decision Tree

```text
Failure
  │
  ├── Is it transient?
  │       │
  │       ├── Yes → Is retry safe?
  │       │             │
  │       │             ├── Yes → Bounded Retry
  │       │             └── No  → Fail / Reconcile
  │       │
  │       └── No → Is fallback safe?
  │                     │
  │                     ├── Yes → Degrade
  │                     └── No  → Fail Safely
  │
  └── Record Telemetry
```

---

# 140. Failure Decision Matrix

| Failure | Retry | Fallback | Fail Closed | Recovery |
|---|---:|---:|---:|---:|
| DB transient error | Yes | Usually no | Yes for critical writes | Automatic |
| Redis cache failure | Limited | Often yes | No | Automatic |
| Kafka unavailable | Publisher retry | Outbox | No immediate data loss | Automatic |
| AI unavailable | Limited | Yes | No if deterministic fallback exists | Automatic |
| Route provider down | Limited | Yes | No if ETA fallback exists | Automatic |
| Legal validation unavailable | Limited | No unsafe fallback | Yes | Automatic/manual |
| Payment timeout | Carefully | Reconcile | Do not duplicate | Reconciliation |
| Notification failure | Yes | Queue | No | Automatic |
| Consumer crash | Redelivery | N/A | No | Automatic |

---

# 141. Consequences

## 141.1 Positive Consequences

This strategy provides:

```text
Failure Containment
Predictable Degradation
Safer Retries
Better Availability
Reduced Cascading Failure
Improved Recovery
Clear Operational Behaviour
```

---

## 141.2 Negative Consequences

The architecture introduces:

```text
More Configuration
More Failure States
Fallback Complexity
Additional Monitoring
Recovery Workflows
Operational Testing
Provider Management
```

These trade-offs are accepted.

---

# 142. Risks

## Risk 1 — Retry Storm

### Mitigation

```text
Bounded Retry
Backoff
Jitter
Circuit Breakers
```

---

## Risk 2 — Unsafe Fallback

### Mitigation

```text
Hard Constraints First
Fallback Validation
Explicit Degradation Design
```

---

## Risk 3 — Silent Data Loss

### Mitigation

```text
Outbox
Durable State
Reconciliation
Monitoring
```

---

## Risk 4 — Cascading Failure

### Mitigation

```text
Bulkheads
Timeouts
Circuit Breakers
Load Shedding
```

---

## Risk 5 — Recovery Causes Duplicate Effects

### Mitigation

```text
Idempotency
State Validation
Reconciliation
```

---

## Risk 6 — Fallback Becomes Permanent

### Mitigation

Track:

```text
Fallback Duration
Fallback Rate
Provider Recovery
```

and alert on prolonged degradation.

---

# 143. Alternatives Considered

## 143.1 Retry Everything

### Advantages

```text
Simple
```

### Disadvantages

```text
Retry Storms
Higher Load
Duplicate Effects
Longer Latency
```

### Decision

```text
Rejected.
```

---

# 144. Fail Everything on Any Dependency Failure

### Advantages

```text
Simple
Predictable
```

### Disadvantages

```text
Poor Availability
Unnecessary User Impact
Optional Features Become Critical
```

### Decision

```text
Rejected.
```

---

# 145. Ignore Dependency Failures

### Advantages

```text
High Apparent Availability
```

### Disadvantages

```text
Incorrect State
Silent Failures
Data Loss
Safety Risk
```

### Decision

```text
Rejected.
```

---

# 146. Unlimited Multi-Provider Fallbacks

### Advantages

```text
High Theoretical Availability
```

### Disadvantages

```text
Cost
Operational Complexity
Provider Drift
Testing Complexity
```

### Decision

```text
Rejected.
```

Fallbacks must be justified per dependency.

---

# 147. AI as Mandatory Dependency

### Advantages

```text
Potentially Better Optimization
```

### Disadvantages

```text
AI Failure Becomes Ride Failure
Higher Latency
Higher Operational Risk
```

### Decision

```text
Rejected.
```

AI must remain degradable for core ride operations.

---

# 148. Validation

This ADR should be validated through:

```text
Unit Tests
Integration Tests
Failure Injection
Chaos Tests
Load Tests
Timeout Tests
Retry Tests
Circuit Breaker Tests
Provider Failover Tests
Database Failure Tests
Kafka Failure Tests
Redis Failure Tests
AI Failure Tests
Recovery Tests
```

---

# 149. Review Triggers

Revisit this ADR when:

```text
A New Critical Dependency Is Added
A New Provider Is Added
A New Fallback Is Introduced
Multi-Region Operation Is Introduced
A Major Outage Occurs
Retry Behaviour Changes
Database Architecture Changes
Messaging Architecture Changes
AI Becomes More Central to Dispatch
RTO / RPO Requirements Change
```

---

# 150. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
adr/
```

Especially:

```text
Error Handling and Validation
Logging and Debugging
Observability Development
Performance and Optimization
Integration Testing
Configuration and Environment
Event and Messaging Development
```

---

# 151. Related ADRs

This decision is directly related to:

```text
ADR-0003 — Microservice Boundaries
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0007 — PostgreSQL as Primary Database
ADR-0009 — Redis for Real-Time State and Caching
ADR-0010 — Driver Location Storage Strategy
ADR-0012 — Outbox Pattern
ADR-0013 — Dead Letter Queue Strategy
ADR-0014 — API and Service Communication
ADR-0015 — Smart Dispatch and Stand Dispatch
ADR-0016 — AI-Assisted Dispatch Strategy
ADR-0017 — ETA and Route Provider Strategy
ADR-0018 — Regional and Legal Ride Validation
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0022 — Observability Strategy
ADR-0023 — Security and Secret Management
ADR-0024 — Configuration and Environment Strategy
ADR-0025 — Testing and Integration Strategy
ADR-0026 — Model and AI Governance
ADR-0027 — Cloud and Deployment Strategy
ADR-0028 — Cost Optimization Strategy
```

---

# 152. Decision Summary

RideForge adopts:

```text
                     FAILURE
                        │
                        ▼
                CLASSIFY FAILURE
                        │
          ┌─────────────┼─────────────┐
          ▼             ▼             ▼
       Transient     Permanent     Dependency
          │             │             │
          ▼             ▼             ▼
     Retry Safely    Fail Safely   Criticality
          │                           │
          ▼                    ┌──────┴──────┐
   Bounded Retry               ▼             ▼
   + Backoff              Critical        Optional
   + Jitter                   │             │
                              ▼             ▼
                         Fail Safely     Degrade
                                            │
                                            ▼
                                        Fallback
```

The platform will use:

```text
Timeouts
Bounded Retries
Exponential Backoff
Jitter
Circuit Breakers
Bulkheads
Load Shedding
Backpressure
Fallbacks
Graceful Degradation
Outbox
Idempotency
Reconciliation
Observability
```

---

# 153. Final Principles

The following principles are mandatory:

```text
1. Every external dependency must have a bounded timeout.

2. Retries must be explicitly justified.

3. Retries must be bounded.

4. Retryable operations must be idempotent or otherwise safe.

5. Exponential backoff and jitter should be used for repeated transient failures.

6. Critical dependencies must fail safely.

7. Optional dependencies should degrade rather than unnecessarily break core functionality.

8. AI must not become a mandatory dependency for core ride lifecycle correctness unless explicitly approved.

9. AI failures should fall back to deterministic mechanisms where a safe fallback exists.

10. Legal and safety constraints must fail closed when required.

11. Payment failures must account for unknown external outcomes.

12. Notification failures must not normally roll back committed business state.

13. Kafka / Redpanda failures must be buffered through the outbox where applicable.

14. Redis failures must be classified according to whether Redis is acting as cache, ephemeral state, or authoritative subsystem state.

15. Database failures must never result in invented transactional state.

16. Fallbacks must preserve hard business constraints.

17. Degraded mode must be observable.

18. Recovery operations must be idempotent.

19. Reconciliation must be used when internal and external state can diverge.

20. Failure containment is preferred over global failure.

21. A dependency failure in one region should not unnecessarily disable unrelated regions.

22. No failure mechanism should rely on infinite retries.

23. Health checks must distinguish liveness from readiness.

24. Graceful shutdown must protect in-flight work.

25. Failure strategies must be tested through integration, failure-injection, and recovery testing.
```

---

# 154. Status

```text
Decision: ACCEPTED

Failure Model:
Explicit Classification

Retries:
Bounded

Backoff:
Exponential + Jitter

Timeouts:
Required

Circuit Breakers:
Used Where Appropriate

Bulkheads:
Used Where Appropriate

Load Shedding:
Used During Capacity Pressure

Critical State:
Fail Safe

Optional Features:
Graceful Degradation

AI:
Degradable

Dispatch:
Smart + Deterministic / Stand Fallback According to Configuration

Legal Validation:
Hard Constraint

Payment:
Idempotency + Reconciliation

Kafka / Redpanda:
Outbox-Based Recovery

Redis:
Failure Behaviour Depends on State Criticality

Database:
Authoritative for Core Transactional State

Recovery:
Automatic Where Possible

Reconciliation:
Required for External State Divergence

Observability:
Required

Primary Goal:
Contain Failures, Protect Critical Business State, and Maintain the Highest Safe Level of Ride-Hailing Availability
```

---

# 148. Dispatch Strategy Preservation During Failure and Degradation

Failure and degradation handling must preserve the dispatch strategy resolved for the ride unless an explicit business/configuration rule defines a strategy transition.

## 148.1 Hierarchical Strategy Resolution

The effective dispatch strategy is resolved by starting at the most specific applicable configuration level and moving upward until an explicit strategy is found.

Possible levels include:

```text
State
District
City / Town
Rural Area
Auto Stand
Specific Ride Level
Other Configured Intermediate Levels
```

Not every level requires configuration.

The canonical rule is:

> **Specific configuration overrides inherited configuration.**

If no explicit strategy exists anywhere in the applicable hierarchy, the system default is used.

Failure handling must not bypass this resolution process.

---

## 148.2 Two Primary Strategies

The primary dispatch strategies are:

```text
Smart Stand Dispatch
Smart Dispatch
```

AI-assisted dispatch is an optimization capability, not a third primary strategy.

---

## 148.3 Smart Stand Dispatch Degradation

Smart Stand Dispatch is stand-preferred, not stand-exclusive.

When the rider is within a stand radius:

```text
Preferred Stand
    ↓
Eligible Stand Drivers
    ↓
Queue / Ordering Rules
```

If suitable stand supply is unavailable, the system may expand the candidate search to:

```text
Drivers outside the stand
Drivers at nearby stands
Drivers from nearby locations
```

If the rider is outside all stand radii, the system may consider all eligible nearby drivers without stand restriction.

During degradation:

```text
Smart Stand Dispatch
        ↓
Broader Candidate Discovery
```

must not silently become:

```text
Smart Dispatch
```

---

## 148.4 Smart Dispatch Degradation

Smart Dispatch remains stand-agnostic during degradation.

If AI ranking fails:

```text
Smart Dispatch
        ↓
Deterministic Smart Dispatch
```

Stand membership must not become an implicit preference merely because AI is unavailable.

---

## 148.5 Cross-Location Degradation

If local supply is insufficient, candidate discovery may expand to nearby locations.

A nearby location may use a different dispatch strategy.

For example:

```text
Location A → Smart Dispatch
Location B → Smart Stand Dispatch
```

A candidate from Location B may still be considered for a ride from Location A when geographic, operational, legal, and other hard constraints permit.

The candidate context should preserve:

```text
Candidate Location
Candidate Location Strategy
Stand Membership
Relevant Stand
Queue Position
Discovery Source
Expansion Level
```

A different source-location strategy is not itself a reason to reject a candidate.

---

## 148.6 Hard Constraints Remain Authoritative

All failure, retry, fallback, and degradation paths must continue enforcing:

```text
Driver Eligibility
Driver Availability
Vehicle / Service Compatibility
Regional / Legal Constraints
Safety Constraints
Ride Constraints
Location Freshness
Other Hard Business Rules
```

AI or degraded operation cannot bypass these constraints.

---

## 148.7 Failure Categories Must Remain Separate

The implementation must distinguish:

```text
AI Failure
Candidate Expansion
Retry
Degradation
Strategy Switching
```

These are not interchangeable.

For example:

```text
Preferred Stand Unavailable
        ↓
Candidate Expansion
        ↓
Still Smart Stand Dispatch
```

while:

```text
AI Unavailable
        ↓
Deterministic Ranking
        ↓
Same Primary Strategy
```

A strategy switch requires an explicit rule.

---

## 148.8 Stand Queue Semantics During Failure

If Smart Stand Dispatch uses first-eligible-in-queue behavior, failure handling must preserve that business rule.

AI failure must not turn:

```text
First Eligible Driver in Queue
```

into:

```text
Highest AI Score
```

unless the business configuration explicitly defines such a ranking model.

---

## 148.9 AI-Agent Implementation Guardrails

Implementations must not:

```text
Treat Smart Stand Dispatch as stand-only.

Treat Smart Stand Dispatch as a generic fallback for Smart Dispatch.

Switch Smart Stand Dispatch to Smart Dispatch merely because the preferred stand has no suitable driver.

Switch strategies merely because AI is unavailable.

Reject non-stand candidates solely because the rider is inside a stand radius.

Reject nearby-location candidates solely because their source location uses another strategy.

Bypass hierarchical configuration during fallback.

Treat geographic expansion as strategy switching.

Allow AI to override legal, safety, eligibility, or operational constraints.

Replace configured stand queue semantics with an arbitrary AI score.

Hard-code the configuration hierarchy into failure handling.
```

The failure/degradation layer must preserve the effective dispatch strategy and only broaden, retry, or degrade execution according to explicitly defined business rules.

