# ADR-0009: Redis for Real-Time State

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Data / Infrastructure  
> **Scope:** RideForge real-time, low-latency, cached, and ephemeral operational state  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge requires storage for workloads that are fundamentally different from durable transactional business data.

Examples include:

```text
Driver Availability
Driver Location
Short-Lived Dispatch State
Matching State
Temporary Locks
Caching
Rate Limiting
Temporary Coordination
```

These workloads can have:

```text
High Read Frequency
High Write Frequency
Low Latency Requirements
Short Data Lifetimes
Frequent Updates
High Concurrency
```

ADR-0007 established PostgreSQL as the primary transactional database.

RideForge therefore needs a complementary low-latency state system without turning that system into an accidental replacement for PostgreSQL.

---

# 2. Problem

RideForge needs an infrastructure component capable of efficiently supporting:

```text
Low-Latency Reads
High-Frequency Writes
Ephemeral State
Caching
Distributed Coordination
Real-Time Operational Workloads
```

The solution must work alongside:

```text
PostgreSQL
Redpanda
Go Services
Dispatch
Matching
ETA
AI
```

without replacing PostgreSQL as the primary source of durable transactional state or Redpanda as the primary event-streaming platform.

---

# 3. Decision

RideForge will use:

> **Redis for real-time, low-latency, cached, and appropriate ephemeral operational state.**

Redis complements PostgreSQL rather than replacing it.

The intended storage model is:

```text
PostgreSQL
→ Durable Transactional State

Redis
→ Real-Time / Derived / Ephemeral State

Redpanda
→ Event Streaming
```

---

# 4. Primary Principle

> **PostgreSQL owns durable transactional state; Redis serves low-latency and ephemeral operational workloads where its characteristics provide a clear advantage.**

Every Redis dataset must have an explicit:

```text
Owner
Purpose
Lifecycle
Freshness Requirement
Failure Behaviour
```

---

# 5. Appropriate Redis Responsibilities

Redis may be used for:

```text
Caching
Driver Availability
Real-Time Driver Location
Short-Lived Matching State
Dispatch Coordination
Distributed Locks
Rate Limiting
Temporary Coordination
Expiring Data
```

The exact use must be justified by workload characteristics.

---

# 6. Redis Is Not the Default Source of Truth

Redis should not become authoritative merely because it is faster.

For durable business state:

```text
PostgreSQL
```

remains authoritative unless a separate ADR explicitly establishes another source of truth.

---

# 7. Durable State vs Real-Time State

RideForge must distinguish:

```text
Durable Business State
```

from:

```text
Real-Time Operational State
```

For example:

```text
Driver Profile
→ PostgreSQL

Driver Current Location
→ Redis / Specialized Location Infrastructure

Ride Lifecycle
→ PostgreSQL

Temporary Dispatch Candidate State
→ Redis where appropriate
```

---

# 8. Driver Availability

Driver availability is a strong Redis candidate when it requires:

```text
Frequent Updates
Fast Reads
Short-Lived State
High Concurrency
```

The durable driver business record remains in PostgreSQL.

Redis availability state must have explicit freshness semantics.

---

# 9. Driver Location

Real-time driver location can generate extremely high write volume.

Redis may be used for:

```text
Current Location
Recent Location
Geospatial Lookup
Nearby Driver Discovery
```

when the workload and scale justify it.

Historical location data must be handled separately according to retention and analytics requirements.

---

# 10. Location Freshness

Real-time location is time-sensitive.

Location entries should include appropriate freshness information and/or expiration.

Conceptually:

```text
Location Update
      ↓
Redis
      ↓
TTL / Timestamp
      ↓
Fresh / Stale / Expired
```

A stale key must not automatically be treated as a current location.

---

# 11. Redis Geospatial Capabilities

Redis geospatial functionality may be used for:

```text
Nearby Driver Search
Radius Queries
Location-Based Candidate Discovery
```

Candidate proximity alone does not establish eligibility.

The complete decision remains:

```text
Location
+
Availability
+
Eligibility
+
Regional Rules
+
Dispatch Strategy
+
Ranking
```

---

# 12. Redis and Matching

A conceptual matching flow may be:

```text
Ride Request
      ↓
Redis Operational / Geospatial State
      ↓
Candidate Drivers
      ↓
Eligibility Validation
      ↓
Ranking
      ↓
Match
```

Redis candidate retrieval must not bypass authoritative business and legal rules.

---

# 13. Redis and Dispatch

Dispatch may use Redis for:

```text
Current Driver Availability
Candidate State
Temporary Assignment Coordination
Short-Lived Dispatch State
```

Redis should not become the authoritative owner of durable ride lifecycle state.

---

# 14. Redis and Smart Dispatch

Smart Dispatch may consume low-latency operational signals such as:

```text
Driver Location
Driver Availability
Recent Operational State
Short-Lived Features
```

AI or ranking systems must not assume that Redis contains complete historical data.

---

# 15. Cache-Aside Pattern

For cacheable data, RideForge may use:

```text
Application
    ↓
Redis
    │
    ├── Hit → Return
    │
    └── Miss
          ↓
      PostgreSQL
          ↓
        Redis
          ↓
       Return
```

This is appropriate when cache staleness is acceptable within the defined policy.

---

# 16. Cache Invalidation

Cache invalidation must be explicit.

Possible strategies include:

```text
TTL
Event-Based Invalidation
Explicit Delete
Write-Through
Cache-Aside
```

The choice depends on:

```text
Freshness
Read Frequency
Write Frequency
Failure Behaviour
```

---

# 17. TTL

TTL should be used for data with a defined lifetime.

Examples include:

```text
Temporary Locks
Rate-Limit Counters
Short-Lived Sessions
Current Location
Temporary Dispatch State
Cache Entries
```

TTL must not be used as a substitute for durable business lifecycle management.

For example, a ride must not become completed merely because a Redis key expires.

---

# 18. Redis Key Design

Redis keys should be:

```text
Predictable
Namespaced
Consistent
Collision-Resistant
Easy to Inspect
```

A conceptual convention is:

```text
<domain>:<resource>:<identifier>
```

Examples:

```text
driver:availability:<driver_id>
driver:location:<driver_id>
ride:dispatch:<ride_id>
```

The exact implementation convention is governed by development standards.

---

# 19. Key Ownership

Each Redis key namespace must have a clear owner.

Conceptually:

```text
driver:*  → Driver / Operational Ownership
ride:*    → Ride / Dispatch Ownership
match:*   → Matching Ownership
```

A service must not freely modify another domain's Redis state.

---

# 20. Value Design

Redis values should be:

```text
Small
Purpose-Specific
Efficient
Versionable Where Necessary
Expirable Where Appropriate
```

Avoid storing unrestricted application objects or unnecessarily large payloads.

---

# 21. Redis Data Structures

Use Redis data structures according to the access pattern.

### Strings

Suitable for:

```text
Simple Values
Counters
Serialized Small Objects
```

### Hashes

Suitable for:

```text
Related Operational Fields
Driver State
Temporary Session State
```

### Sets

Suitable for:

```text
Membership
Unique Candidate IDs
Temporary Groups
```

### Sorted Sets

Suitable for:

```text
Ranking
Priority
Time Ordering
Dispatch Ordering
```

### Geospatial Data

Suitable for:

```text
Nearby Driver Discovery
Radius Searches
```

---

# 22. Redis Streams

Redis Streams may be used for specialized workloads where appropriate.

However:

> **Redpanda remains the primary event-streaming platform for RideForge.**

Redis Streams must not become a second general-purpose event backbone without a separate architectural decision.

---

# 23. Redis Pub/Sub

Redis Pub/Sub may be used for transient notification patterns where message durability is not required.

Business-critical durable events should use:

```text
Redpanda
```

rather than relying on Pub/Sub delivery.

---

# 24. Distributed Locks

Redis may be used for short-lived distributed coordination.

A lock should have:

```text
Owner Identity
Expiration
Controlled Acquisition
Controlled Release
Failure Handling
```

Locks should remain short-lived.

---

# 25. Lock Safety

A Redis lock must not be treated as the sole protection for critical business invariants.

Where correctness matters, also use appropriate:

```text
PostgreSQL Transactions
Atomic State Changes
Constraints
Idempotency
```

---

# 26. Rate Limiting

Redis is suitable for high-throughput rate limiting.

Potential uses include:

```text
API Rate Limits
OTP Request Limits
Authentication Attempts
Driver Action Limits
External Provider Limits
```

Rate-limit keys should have controlled TTLs.

---

# 27. Redis Failure Modes

Redis failure must have an explicit degradation strategy.

For cache workloads:

```text
Redis Failure
      ↓
Cache Bypass
      ↓
Authoritative Store
```

may be appropriate.

For real-time state:

```text
Redis Failure
      ↓
Reject Stale State / Reduce Capability / Use Valid Fallback
```

depending on the domain.

Missing state must never silently be interpreted as fresh state.

---

# 28. Redis Recovery

After recovery:

```text
Authoritative Source / Current Operational Signals
                 ↓
          Rebuild / Refresh
                 ↓
               Redis
```

Derived Redis state should be recoverable where practical.

---

# 29. Redis as Derived State

Where Redis represents PostgreSQL-owned data:

```text
PostgreSQL
     ↓
Redis
```

Redis is derived state.

If Redis is lost, the durable business record should remain available in PostgreSQL.

---

# 30. Redis as Ephemeral State

Some state may intentionally exist only in Redis:

```text
Temporary Lock
Rate Limit
Short-Lived Candidate State
Current Location
```

Each such workload must explicitly define what happens when the data disappears.

---

# 31. Persistence

Redis persistence must be selected according to workload criticality.

For cache-only workloads:

```text
Data Loss May Be Acceptable
```

For operational state:

```text
Recovery Requirements Must Be Evaluated
```

A single persistence policy must not automatically be applied to every Redis workload.

---

# 32. Memory Management

Redis is memory-oriented infrastructure.

Monitor:

```text
Used Memory
Peak Memory
Memory Growth
Evictions
Fragmentation
Large Keys
```

Memory capacity must include operational headroom.

---

# 33. Eviction Policy

Eviction policy must match the workload.

Cache data may tolerate eviction.

Critical operational state may not.

Never assume that a cache-friendly eviction policy is safe for business-critical operational state.

---

# 34. Large Keys

Avoid large values and uncontrolled collections.

Large keys can cause:

```text
Memory Pressure
Latency Spikes
Slow Operations
Replication Pressure
Operational Difficulty
```

---

# 35. Cache Stampede

Large simultaneous cache misses can overload PostgreSQL.

Possible mitigations include:

```text
Request Coalescing
Short Locks
TTL Jitter
Prewarming
Stale-While-Revalidate
```

Use only where workload evidence justifies them.

---

# 36. Cache Penetration

Repeated requests for nonexistent data can bypass the cache.

Possible mitigations include:

```text
Negative Caching
Input Validation
Rate Limiting
```

where appropriate.

---

# 37. Cache Avalanche

Mass expiration can create sudden database load.

Where necessary, use:

```text
TTL Jitter
Staggered Expiration
Controlled Refresh
```

---

# 38. Concurrency

Redis is designed for highly concurrent operations, but application concurrency must still be controlled.

Monitor:

```text
Connections
Commands
Pipelines
Large Operations
Hot Keys
```

---

# 39. Connection Management

Redis clients should use appropriate pooling or multiplexing.

Applications must not create unrestricted Redis connections.

Redis calls should also have appropriate timeouts.

---

# 40. Atomic Operations

Prefer native atomic Redis commands when they provide the required semantics.

Examples include:

```text
SET NX
INCR
SADD
ZADD
```

Multi-step operations should not assume atomicity unless Redis semantics actually provide it.

---

# 41. Pipelining

Pipelining may be used when multiple independent commands can be combined to reduce network round trips.

It should be introduced where measurements demonstrate value.

---

# 42. Server-Side Scripts

Lua or equivalent server-side scripting may be used for small atomic operations when necessary.

Scripts should remain:

```text
Small
Deterministic
Tested
Operationally Safe
```

Do not turn Redis scripts into a second domain-logic layer.

---

# 43. Redis and PostgreSQL Consistency

Redis and PostgreSQL may temporarily disagree when Redis contains cached or derived state.

This is acceptable only when the workflow explicitly allows it.

For example:

```text
Cache
→ Eventual Consistency

Ride Lifecycle
→ Stronger Transactional Consistency
```

---

# 44. Redis and Redpanda

Redis and Redpanda serve different roles:

```text
Redis
→ Fast State / Cache / Coordination

Redpanda
→ Durable Event Stream
```

Avoid duplicating responsibilities across them without a clear architectural reason.

---

# 45. Redis and AI

AI systems may consume Redis state for low-latency prediction inputs.

However:

```text
Current Redis State
```

must not automatically be treated as:

```text
Historical Training Data
```

Training pipelines require their own data-quality and historical-data strategy.

---

# 46. Redis and ETA

Redis may cache:

```text
Recent ETA
Routing Results
Short-Lived Prediction Results
```

where caching improves latency or reduces external provider cost.

Cached ETA values must have explicit freshness requirements.

---

# 47. Redis and External Providers

External provider responses may be cached when:

```text
The Response Is Safe to Cache
Freshness Is Defined
The Cache Key Is Correct
```

Sensitive or user-specific data requires additional privacy review.

---

# 48. Redis and Notifications

Redis may support:

```text
Rate Limiting
Deduplication
Temporary Delivery State
Short-Lived Coordination
```

Durable notification events should continue to use the event-streaming architecture.

---

# 49. Redis and Microservices

Multiple services may use Redis, but logical ownership remains mandatory.

A shared physical Redis deployment does not imply unrestricted shared state.

Prefer:

```text
Service A
   ↓
Contract / Event
   ↓
Service B
   ↓
Service B Redis State
```

rather than direct modification of Service B's keys.

---

# 50. Redis Schema Evolution

Redis values that survive deployments require compatibility consideration.

Possible strategies include:

```text
Versioned Values
Dual Read
Dual Write
Migration
TTL-Based Natural Expiration
```

The strategy should match the lifetime and criticality of the data.

---

# 51. Testing

Redis-backed functionality should use a real Redis-compatible environment when testing:

```text
TTL
Atomicity
Data Structures
Expiration
Concurrency
Geospatial Queries
Serialization
```

Mocks may be used for isolated unit tests but must not replace integration tests for actual Redis behaviour.

---

# 52. Failure Testing

Important scenarios include:

```text
Redis Unavailable
Redis Restart
Key Expiration
Memory Pressure
Connection Failure
Stale State
Cache Miss
```

The expected degradation behaviour must be tested.

---

# 53. Load Testing

High-frequency workloads should be tested for:

```text
Commands / Second
Latency
Connection Usage
Memory Growth
CPU
Network
Hot Keys
```

---

# 54. Observability

Important Redis telemetry includes:

```text
Memory Usage
Command Rate
Command Latency
Connections
Evictions
Expired Keys
Cache Hit Rate
Cache Miss Rate
Replication Health
CPU
Network
```

Operational state should additionally expose relevant:

```text
Fresh State Count
Expired State Count
Update Rate
Stale State Rate
```

---

# 55. Security

Redis access must follow:

```text
Least Privilege
Network Isolation
Authentication
Secure Transport Where Required
```

Redis should not be publicly exposed without an explicit security architecture.

Credentials must never be committed to source control.

---

# 56. Privacy

Redis may contain sensitive operational information such as:

```text
Location
Driver State
User-Related Temporary State
```

Data should be minimized and retained only as long as necessary.

Complete location streams should not be logged unnecessarily.

---

# 57. Availability

Redis deployment must match workload criticality.

Possible production approaches include:

```text
Replication
Sentinel
Managed Redis
Redis Cluster
```

The chosen deployment should be based on:

```text
Availability
Memory
Throughput
Failover
Recovery
```

requirements.

---

# 58. Redis Cluster

Redis Cluster should be introduced only when:

```text
Memory Capacity
Throughput
Availability
Horizontal Scaling
```

requirements justify its complexity.

Do not introduce clustering merely because Redis is present.

---

# 59. Local Development

Local development should provide a simple disposable Redis environment.

The local stack may include:

```text
PostgreSQL
Redis
Redpanda
RideForge Services
```

as required by the workflow.

Local Redis data should normally be resettable.

---

# 60. Development Workflow

A typical Redis development workflow is:

```text
Define Operational State
      ↓
Determine Durability Requirement
      ↓
Define Key / Data Structure
      ↓
Define TTL
      ↓
Define Failure Behaviour
      ↓
Implement
      ↓
Integration Test
      ↓
Load Test Where Required
      ↓
Observe
```

---

# 61. Redis Adoption Checklist

```text
[ ] Latency requirement identified
[ ] Read/write frequency identified
[ ] Durability requirement identified
[ ] Source of truth identified
[ ] TTL requirement identified
[ ] Failure behaviour defined
[ ] Key ownership defined
[ ] Memory impact considered
[ ] Eviction impact considered
[ ] Security/privacy impact considered
[ ] Observability defined
[ ] Integration tests added
```

---

# 62. New Redis Key Checklist

```text
[ ] Namespace defined
[ ] Owning domain defined
[ ] Value structure defined
[ ] TTL defined
[ ] Maximum size considered
[ ] Serialization defined
[ ] Invalidation strategy defined
[ ] Failure behaviour defined
[ ] Security/privacy reviewed
```

---

# 63. Cache Checklist

```text
[ ] Authoritative source identified
[ ] Cache key defined
[ ] TTL defined
[ ] Invalidation strategy defined
[ ] Cache miss behaviour defined
[ ] Stampede risk considered
[ ] Stale-data tolerance defined
[ ] Metrics added
```

---

# 64. Real-Time State Checklist

```text
[ ] State ownership defined
[ ] Freshness requirement defined
[ ] TTL defined
[ ] Stale-state behaviour defined
[ ] Recovery strategy defined
[ ] Failure mode defined
[ ] Concurrency model defined
[ ] Load-tested where necessary
```

---

# 65. Lock Checklist

```text
[ ] Lock purpose defined
[ ] Lock owner identified
[ ] TTL defined
[ ] Release behaviour defined
[ ] Failure behaviour defined
[ ] Critical invariant protected elsewhere if required
[ ] Contention considered
[ ] Recovery tested
```

---

# 66. Decision Matrix

| Workload                      | PostgreSQL    | Redis                                     | Redpanda            |
| ----------------------------- | ------------- | ----------------------------------------- | ------------------- |
| Durable ride state            | **Primary**   | No                                        | Event output        |
| Driver profile                | **Primary**   | Cache                                     | Event output        |
| Current driver location       | Possible      | **Primary candidate**                     | Events where useful |
| Driver availability           | Durable state | **Primary candidate for real-time state** | Events              |
| Cache                         | No            | **Primary**                               | No                  |
| Rate limiting                 | No            | **Primary**                               | No                  |
| Temporary lock                | No            | **Primary candidate**                     | No                  |
| Durable event stream          | No            | No                                        | **Primary**         |
| Notification events           | No            | Supporting                                | **Primary**         |
| Analytics events              | No            | No                                        | **Primary**         |
| Historical transactional data | **Primary**   | No                                        | No                  |

---

# 67. Alternatives Considered

## 67.1 PostgreSQL Only

### Advantages

```text
Single Data Technology
Simple Source of Truth
Strong Transactions
```

### Disadvantages

```text
High-Frequency Writes
Connection Pressure
Latency
Unnecessary Database Load
Large Operational Tables
```

### Decision

PostgreSQL remains primary, while Redis handles workloads that benefit materially from low-latency operational storage.

---

## 67.2 Redis as Primary Database

### Advantages

```text
Very Low Latency
High Throughput
Simple Key-Value Access
```

### Disadvantages

```text
Poor Fit for Core Relational Transactions
Limited Relational Integrity
Different Durability Model
Complex Business Relationships
```

### Decision

Redis is not the primary transactional database.

---

## 67.3 Redis Streams as Primary Event Backbone

### Advantages

```text
Existing Redis Infrastructure
Consumer Groups
Stream Semantics
```

### Disadvantages

```text
Duplicates Event Infrastructure Responsibilities
Less Alignment With Selected Kafka-Compatible Architecture
```

### Decision

Redpanda remains the primary event-streaming platform.

---

## 67.4 Specialized Location Database

A specialized geospatial or distributed database may eventually be considered for very large location workloads.

At the current architectural stage, Redis is an appropriate real-time state option where workload evidence supports it.

A future specialized location store requires a separate ADR.

---

# 68. Cost Considerations

Redis introduces:

```text
Memory Cost
High-Availability Cost
Operational Cost
Monitoring Cost
```

Therefore Redis should not be used merely because it is fast.

Every Redis workload should have a measurable or architectural justification.

---

# 69. Performance Principles

Redis optimization should consider:

```text
Network Latency
Serialization
Connection Usage
Command Complexity
Memory Pressure
Hot Keys
Payload Size
```

A faster data store does not eliminate inefficient application access patterns.

---

# 70. Data Lifecycle

Every Redis dataset should have an explicit lifecycle:

```text
Created
   ↓
Updated
   ↓
Refreshed
   ↓
Expired / Invalidated
   ↓
Removed
```

Data without a defined lifecycle should not be placed in Redis casually.

---

# 71. Consequences

## 71.1 Positive Consequences

The decision provides:

```text
Low-Latency Operational State
Reduced PostgreSQL Pressure
Efficient Caching
Fast Candidate Retrieval
Real-Time Coordination
High-Concurrency Support
```

## 71.2 Negative Consequences

The architecture introduces:

```text
Additional Infrastructure
Cache Invalidation Complexity
Stale Data Risk
Memory Management
Redis Availability Requirements
Additional Monitoring
```

These trade-offs are accepted for workloads that benefit from Redis.

---

# 72. Risks

## Risk 1 — Redis Becomes a Hidden Source of Truth

**Mitigation:** Explicitly document ownership and durability for every Redis dataset.

## Risk 2 — Stale Operational State

**Mitigation:** Use TTL, timestamps, freshness validation, and state expiration where required.

## Risk 3 — Memory Exhaustion

**Mitigation:** Monitor memory, growth, evictions, large keys, and capacity headroom.

## Risk 4 — Cache Invalidation Bugs

**Mitigation:** Define invalidation strategy and test stale-data scenarios.

## Risk 5 — Redis Becomes a Second Event Broker

**Mitigation:** Keep Redpanda as the primary event-streaming platform.

## Risk 6 — Distributed Lock Misuse

**Mitigation:** Keep locks short-lived and protect critical business invariants through authoritative transactional mechanisms.

---

# 73. Validation

The Redis decision should be validated through:

```text
Integration Tests
Concurrency Tests
Failure Tests
Load Tests
Memory Tests
TTL Tests
Cache Behaviour Tests
Recovery Tests
Production Observability
```

---

# 74. Review Triggers

Revisit this ADR when:

```text
Redis Memory Becomes a Material Cost
Location Volume Increases Significantly
Redis No Longer Meets Latency Requirements
A Specialized Location Store Is Proposed
A New Durable Redis Workload Is Proposed
Redis Becomes a Critical Dependency
A Major Consistency Problem Appears
```

---

# 75. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
```

Especially:

```text
Redis Development
Database Development
Event and Messaging Development
Performance and Optimization
Error Handling and Validation
Integration Testing and Local Infrastructure
Observability Development
Smart Dispatch AI
AI Matching and Ranking
AI Failure and Fallback Strategy
```

---

# 76. Related ADRs

This decision is directly related to:

```text
ADR-0002 — Architecture Style
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0007 — PostgreSQL as Primary Database
ADR-0012 — Outbox Pattern
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0028 — Cost Optimization Strategy
```

---

# 77. Decision Summary

RideForge adopts Redis as the infrastructure for appropriate:

```text
Real-Time State
+
Low-Latency State
+
Caching
+
Ephemeral Operational Data
+
Coordination
```

while maintaining:

```text
PostgreSQL
→ Durable Transactional Source of Truth

Redis
→ Real-Time / Derived / Ephemeral State

Redpanda
→ Durable Event Streaming
```

Every Redis workload must have explicit:

```text
Ownership
+
Lifecycle
+
Freshness
+
Failure Strategy
```

---

# 78. Final Principle

> **Redis exists to make the right workloads fast and operationally efficient; it must not silently become the source of truth for durable business state.**

The intended relationship is:

```text
                RideForge Data Architecture
                         │
          ┌──────────────┼──────────────┐
          │              │              │
     PostgreSQL        Redis         Redpanda
          │              │              │
     Durable State   Real-Time      Event Stream
                      State
          │              │              │
          └──────┬───────┴──────────────┘
                 │
             Services
```

Redis should therefore be introduced deliberately with explicit:

```text
TTL
Ownership
Freshness
Failure Behaviour
Memory Limits
Observability
```

---

# 79. Status

```text
Decision: ACCEPTED

Redis Role:
Real-Time
+
Low-Latency
+
Ephemeral
+
Cache
+
Operational State
```

This decision establishes Redis as RideForge's real-time state and caching infrastructure while preserving PostgreSQL as the primary transactional database and Redpanda as the primary event-streaming platform.
