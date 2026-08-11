# ADR-0019: Data Consistency and Transaction Boundaries

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Data Architecture / Transaction Management / Distributed Systems  
> **Scope:** PostgreSQL transactions, service boundaries, consistency guarantees, event publication, concurrency, and cross-service state changes  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is an event-driven, microservice-compatible ride-hailing platform.

The platform contains operations such as:

```text
Create Ride
Accept Ride
Assign Driver
Start Trip
Complete Trip
Cancel Ride
Update Driver Availability
Update Driver Location
Publish Ride Events
Publish Match Events
Update Operational State
```

These operations may involve:

```text
PostgreSQL
Redis
Kafka / Redpanda
External Providers
Multiple Application Services
```

A distributed system cannot safely assume that one transaction can atomically update every participating system.

RideForge therefore needs explicit rules defining:

```text
What must be atomic
What may be eventually consistent
Where transactions begin and end
How database state and events remain consistent
How concurrent requests are handled
How retries behave
How failures are recovered
```

---

# 2. Problem

Without explicit transaction boundaries, the system can produce states such as:

```text
Ride Created
but Event Missing

Driver Assigned
but Ride State Not Updated

Ride Cancelled
but Driver Still Reserved

Database Updated
but Kafka Event Lost

Event Published
but Database Transaction Rolled Back

Two Drivers Assigned to One Ride
```

These failures become especially dangerous in ride-hailing workflows because state transitions affect:

```text
Money
Driver Availability
Ride Lifecycle
Dispatch
Notifications
Analytics
AI Feedback
```

---

# 3. Decision

RideForge will use:

```text
Local ACID Transactions
+
Explicit Transaction Boundaries
+
Outbox Pattern
+
Idempotent Event Consumers
+
Optimistic / Pessimistic Concurrency Controls Where Required
+
Eventual Consistency Across Service Boundaries
```

The core principle is:

> **A transaction should protect the smallest meaningful consistency boundary, and distributed services should not attempt to create a global database transaction.**

---

# 4. Transaction Boundary Principle

A transaction should normally be owned by the application/service responsible for the state being changed.

Conceptually:

```text
Application Service
       ↓
Domain Operation
       ↓
Repository
       ↓
PostgreSQL Transaction
```

The transaction should contain the state changes that must succeed or fail together.

---

# 5. No Distributed Database Transactions

RideForge will not use distributed transactions such as:

```text
Two-Phase Commit
XA Transactions
Global Database Transaction
```

across independent services.

The architecture instead uses:

```text
Local Transaction
+
Outbox
+
Events
+
Idempotent Consumers
```

---

# 6. Why Local Transactions

Local PostgreSQL transactions provide:

```text
Atomicity
Consistency
Isolation
Durability
```

without introducing the operational complexity of distributed transactions.

---

# 7. Transaction Ownership

The service owning the aggregate/state should own its transaction.

Example:

```text
Ride Service
→ Ride Transaction

Driver Service
→ Driver Transaction

Payment Service
→ Payment Transaction
```

A service should not directly modify another service's database.

---

# 8. Database Ownership

Each service should logically own its persistent state.

Conceptually:

```text
Ride Service
    ↓
Ride Data

Driver Service
    ↓
Driver Data

Payment Service
    ↓
Payment Data
```

The exact physical database topology may evolve, but ownership boundaries must remain explicit.

---

# 9. Shared PostgreSQL

PostgreSQL may initially host multiple service schemas or modules for cost and operational simplicity.

This does not imply:

```text
Shared Tables Everywhere
```

Logical ownership must still be enforced.

---

# 10. Shared Database Does Not Mean Shared Transaction

Even if two modules use the same PostgreSQL cluster:

```text
Service A
Service B
```

should not automatically combine unrelated business operations into one transaction.

Transaction boundaries follow domain ownership, not merely physical database location.

---

# 11. Atomic Operation

An operation should be atomic when partial completion would create an invalid state.

Example:

```text
Ride Status
+
Assigned Driver
```

may need to change atomically when assignment is committed.

---

# 12. Example: Ride Creation

A ride creation transaction may contain:

```text
Create Ride Record
Create Required Ride State
Create Outbox Event
```

The desired result is:

```text
COMMIT
    ↓
Ride Exists
+
Outbox Event Exists
```

or:

```text
ROLLBACK
    ↓
Neither Exists
```

---

# 13. Ride Creation Flow

```text
BEGIN
   ↓
Validate Domain State
   ↓
Create Ride
   ↓
Create Outbox Event
   ↓
COMMIT
```

After commit:

```text
Outbox Publisher
      ↓
Kafka / Redpanda
```

---

# 14. Outbox Boundary

The outbox record must be created inside the same database transaction as the business state change.

Conceptually:

```text
BEGIN
 ├── Business State Change
 └── Outbox Insert
COMMIT
```

This is governed by:

```text
ADR-0012 — Outbox Pattern
```

---

# 15. Why Outbox Is Required

Without an outbox:

```text
BEGIN
 ↓
Update Database
 ↓
COMMIT
 ↓
Publish Event
 ↓
Kafka Failure
```

could produce:

```text
Database Updated
Event Missing
```

The outbox makes the state transition and event intent atomic.

---

# 16. Event Publication Is Asynchronous

The database transaction does not normally wait for Kafka/Redpanda publication.

Instead:

```text
PostgreSQL
   ↓
Outbox
   ↓
Publisher
   ↓
Kafka / Redpanda
```

This keeps the database transaction bounded.

---

# 17. Database and Kafka Are Not One Transaction

RideForge will not require:

```text
PostgreSQL Commit
+
Kafka Commit
```

to occur as one distributed transaction.

Instead:

```text
Database Transaction
      ↓
Durable Outbox
      ↓
Asynchronous Publication
```

---

# 18. Eventual Consistency

Cross-service state will normally be eventually consistent.

Example:

```text
Ride Created
      ↓
Ride Event
      ↓
Matching Service
      ↓
Candidate Evaluation
```

There may be a small delay between:

```text
Ride State Change
```

and:

```text
Consumer State Update
```

This is accepted by the architecture.

---

# 19. Strong Consistency vs Eventual Consistency

Use strong consistency for:

```text
Critical Aggregate State
Assignment Ownership
Financial State
State Transitions That Must Be Atomic
```

Use eventual consistency for:

```text
Analytics
Search Read Models
Notifications
AI Features
Operational Projections
Non-Critical Caches
```

---

# 20. Consistency Classification

Each piece of state should be classified as:

```text
Strongly Consistent
Eventually Consistent
Best Effort
Cached
Derived
```

This classification should be documented rather than assumed.

---

# 21. Ride Lifecycle State

Ride lifecycle state is business-critical.

Transitions such as:

```text
REQUESTED
→ SEARCHING
→ MATCHED
→ DRIVER_ARRIVING
→ DRIVER_ARRIVED
→ IN_PROGRESS
→ COMPLETED
```

must be protected against invalid concurrent transitions.

---

# 22. State Machine

Ride transitions should be represented as explicit domain rules.

For example:

```text
REQUESTED
   ↓
SEARCHING
   ↓
MATCHED
   ↓
DRIVER_ARRIVING
   ↓
DRIVER_ARRIVED
   ↓
IN_PROGRESS
   ↓
COMPLETED
```

Invalid transitions must be rejected.

---

# 23. Transaction + State Transition

A state transition should normally be committed within a transaction:

```text
BEGIN
 ↓
Verify Current State
 ↓
Validate Transition
 ↓
Update State
 ↓
Create Outbox Event
 ↓
COMMIT
```

---

# 24. Conditional Updates

Where appropriate, state changes should use conditional updates.

Conceptually:

```sql
UPDATE rides
SET status = 'MATCHED'
WHERE id = ?
  AND status = 'SEARCHING';
```

The affected-row count determines whether the transition succeeded.

---

# 25. Why Conditional Updates

This protects against:

```text
Concurrent Assignment
Duplicate Requests
Stale Workers
Retrying Consumers
Race Conditions
```

---

# 26. Assignment Atomicity

Driver assignment is a critical consistency boundary.

The system must prevent:

```text
One Ride
→ Two Drivers
```

and:

```text
One Driver
→ Two Conflicting Active Rides
```

where the business rules prohibit it.

---

# 27. Assignment Transaction

A simplified assignment transaction may be:

```text
BEGIN
 ↓
Lock / Validate Ride
 ↓
Lock / Validate Driver
 ↓
Validate Eligibility
 ↓
Update Ride Assignment
 ↓
Update Driver Assignment State
 ↓
Create Outbox Events
 ↓
COMMIT
```

The exact lock strategy depends on the implementation.

---

# 28. Pessimistic Locking

Pessimistic locking may be used where contention is high and duplicate assignment must be prevented.

Examples:

```text
SELECT ... FOR UPDATE
```

may protect:

```text
Ride Row
Driver Row
Reservation Row
```

when appropriate.

---

# 29. Do Not Lock Unnecessarily

Long-lived database locks are dangerous.

Avoid:

```text
BEGIN
 ↓
External API Call
 ↓
Wait
 ↓
AI Inference
 ↓
Commit
```

Instead:

```text
Validate / Prepare
 ↓
Short Transaction
 ↓
Commit
```

---

# 30. External Calls Inside Transactions

Avoid external calls inside database transactions.

Examples:

```text
HTTP API
Kafka
Redis Network Call
Map Provider
AI Model
Payment Gateway
```

unless there is a very specific and justified reason.

---

# 31. Why

External calls can cause:

```text
Long Transactions
Lock Contention
Connection Exhaustion
Timeout Cascades
Deadlocks
```

---

# 32. Transaction Duration

Transactions should be:

```text
Short
Deterministic
Bounded
```

The transaction should perform only the database work required for the atomic state change.

---

# 33. Connection Pool Interaction

A long transaction consumes a PostgreSQL connection for its entire lifetime.

This matters because RideForge uses:

```text
PostgreSQL
+
PgBouncer
```

as part of the database architecture.

Long transactions reduce effective connection-pool capacity.

---

# 34. Transaction Isolation

PostgreSQL's default isolation level:

```text
READ COMMITTED
```

should be the default unless a specific workflow requires stronger isolation.

Do not automatically use:

```text
SERIALIZABLE
```

for every transaction.

---

# 35. READ COMMITTED

READ COMMITTED is suitable for many normal operations because it provides a good balance between:

```text
Correctness
Concurrency
Performance
```

---

# 36. REPEATABLE READ

Use REPEATABLE READ only when the operation requires a stable snapshot across multiple reads.

---

# 37. SERIALIZABLE

Use SERIALIZABLE only where business correctness genuinely requires serializable behaviour and the retry implications are understood.

---

# 38. Serialization Failures

Serializable transactions may fail and require retry.

The application must treat serialization failures as:

```text
Retryable Transaction Error
```

when safe.

---

# 39. Deadlocks

Deadlocks can occur when concurrent transactions lock resources in different orders.

RideForge should define consistent lock ordering.

---

# 40. Lock Ordering

Where multiple resources must be locked:

```text
Ride
→ Driver
```

or another documented order should be used consistently.

Do not allow different code paths to use:

```text
Ride → Driver
```

in one place and:

```text
Driver → Ride
```

in another without justification.

---

# 41. Deadlock Retry

A deadlock may be retried when:

```text
Operation Is Safe
+
Transaction Is Idempotent
```

Retry count must be bounded.

---

# 42. Optimistic Concurrency

Optimistic concurrency may be preferred where contention is low.

A version field can be used:

```text
version = 10
```

Update:

```text
WHERE id = ?
AND version = 10
```

then:

```text
version = 11
```

---

# 43. Choosing Locking Strategy

Use:

```text
Pessimistic Locking
```

when:

```text
Contention Is High
Exclusive Ownership Is Critical
```

Use:

```text
Optimistic Concurrency
```

when:

```text
Conflicts Are Less Frequent
Retries Are Acceptable
```

---

# 44. Idempotency

Transaction correctness depends on idempotency.

RideForge must distinguish:

```text
First Successful Execution
```

from:

```text
Retry of the Same Request
```

This is governed by:

```text
ADR-0020 — Idempotency Strategy
```

---

# 45. Idempotent Ride Creation

A repeated ride creation request should not create duplicate rides when the same idempotency key represents the same logical request.

---

# 46. Idempotent State Transition

Repeated:

```text
Cancel Ride
```

should produce a deterministic result rather than creating multiple side effects.

---

# 47. Idempotent Event Consumers

Consumers may receive duplicate events.

Therefore:

```text
Event Processing
```

must be idempotent.

---

# 48. Exactly-Once Illusion

RideForge should not rely on global:

```text
Exactly Once
```

semantics across all distributed components.

Instead use:

```text
At-Least-Once Delivery
+
Idempotent Processing
```

where applicable.

---

# 49. Outbox Delivery

The outbox publisher may publish an event and crash before marking it published.

This can produce:

```text
Duplicate Event
```

Consumers must tolerate it.

---

# 50. Consumer Transaction

A consumer that updates its local database should generally use:

```text
BEGIN
 ↓
Check Idempotency
 ↓
Apply State Change
 ↓
Record Processed Event
 ↓
COMMIT
```

---

# 51. Processed Event Record

A consumer may store:

```text
consumer
event_id
processed_at
```

to prevent duplicate processing.

The exact implementation is defined by the messaging architecture.

---

# 52. Event Ordering

Events for the same aggregate may require ordering.

RideForge should define ordering keys such as:

```text
ride_id
driver_id
```

where event ordering is required.

---

# 53. Ordering Is Not Global

The system should not assume global ordering across:

```text
All Rides
All Drivers
All Events
```

Kafka/Redpanda ordering should be designed around the aggregate/key that requires it.

---

# 54. Stale Events

A consumer may receive:

```text
Event Version 4
```

after:

```text
Event Version 5
```

if the architecture permits such delivery across different streams or processing paths.

Consumers should use versioning or state validation where required.

---

# 55. Event Version

Important domain events should carry sufficient metadata to identify:

```text
event_id
aggregate_id
event_type
occurred_at
version
```

---

# 56. Database Version

Aggregate records may also contain:

```text
version
```

to protect against stale updates.

---

# 57. Transaction Boundary Around Events

The preferred pattern is:

```text
BEGIN
 ├── Update Aggregate
 ├── Increment Version
 └── Insert Outbox Event
COMMIT
```

---

# 58. Event Publication Failure

If Kafka/Redpanda is unavailable:

```text
Database Transaction
→ Still Commits
```

provided:

```text
Outbox Insert
→ Succeeds
```

The publisher retries later.

---

# 59. Outbox Failure

If the outbox insert fails:

```text
Business Transaction
→ Rollback
```

The system must not commit the business state without the required event intent.

---

# 60. Outbox Cleanup

Published outbox records may eventually be archived or deleted according to the retention policy.

Do not delete them before the publication guarantee has been satisfied.

---

# 61. Eventual Consistency Window

After a transaction commits:

```text
Database State = Current
```

while:

```text
Consumers = Potentially Slightly Behind
```

This window is expected.

---

# 62. Read-After-Write

When a user performs:

```text
Create Ride
```

and immediately reads the ride from the authoritative service:

```text
Strong Read-After-Write
```

should normally be available from the owning database.

---

# 63. Read Models

Derived read models may lag:

```text
Ride DB
 ↓
Event
 ↓
Projection
```

The product should tolerate this where eventual consistency is acceptable.

---

# 64. Avoid Read-Model Authority

A derived projection should not become the source of truth for the aggregate unless explicitly designed as such.

---

# 65. Redis Consistency

Redis is not the authoritative source for core transactional ride state unless explicitly defined by a domain decision.

Redis may hold:

```text
Cache
Location State
Ephemeral Availability
Rate Limits
Locks
```

according to the relevant architecture.

---

# 66. Redis Failure

A Redis failure should not corrupt PostgreSQL transactional state.

---

# 67. Cache-Aside

For ordinary cache data:

```text
Read
 ↓
Cache
 ↓
Miss
 ↓
Database
 ↓
Cache
```

The database remains authoritative.

---

# 68. Cache Invalidation

When persistent state changes:

```text
Database Commit
 ↓
Invalidate / Update Cache
```

Cache operations should not be required for the database transaction to succeed unless the cache itself is part of the authoritative state model.

---

# 69. Redis Inside Transaction

Do not assume PostgreSQL and Redis can commit atomically.

Avoid designs that require:

```text
PostgreSQL Commit
+
Redis Commit
```

as one transaction.

---

# 70. Distributed Locking

Redis locks may be used for selected coordination problems, but they must not replace database constraints for critical ownership guarantees.

For example:

```text
Unique Database Constraint
```

is preferred over relying solely on:

```text
Redis Lock
```

to prevent duplicate assignment.

---

# 71. Database Constraints

Critical invariants should be enforced at the database level where practical.

Examples:

```text
UNIQUE
NOT NULL
CHECK
FOREIGN KEY
EXCLUSION / Partial Unique Constraints
```

---

# 72. Application Validation vs Database Constraints

Use both:

```text
Application Validation
+
Database Constraint
```

for critical invariants.

Application validation provides:

```text
Better Errors
```

while database constraints provide:

```text
Final Integrity Protection
```

---

# 73. Unique Assignment Constraint

Where the data model allows it, enforce uniqueness for active assignment.

For example:

```text
One Active Ride Assignment
→ One Driver
```

The exact database constraint depends on the ride lifecycle schema.

---

# 74. Foreign Keys

Use foreign keys where the relationship is within the same database ownership boundary and referential integrity is required.

Avoid cross-service foreign keys that couple independent service databases.

---

# 75. Cross-Service References

Cross-service relationships should normally use identifiers:

```text
ride_id
driver_id
user_id
```

rather than cross-database foreign keys.

---

# 76. Referential Consistency Across Services

Cross-service references may temporarily point to data that a consumer has not received yet.

This is an accepted eventual consistency condition when the architecture requires it.

---

# 77. Delete Semantics

Avoid physical deletion of critical domain records when historical/audit requirements require preservation.

Prefer lifecycle states such as:

```text
ACTIVE
CANCELLED
COMPLETED
ARCHIVED
```

where appropriate.

---

# 78. Soft Delete

Soft deletion may be used when:

```text
History Matters
Audit Matters
References Must Remain
```

It should not be applied blindly to every table.

---

# 79. Transaction Scope and Repository Layer

Repositories should participate in a transaction supplied by the application/service layer.

Avoid repositories silently creating unrelated transactions inside a larger business operation.

---

# 80. Transaction Context

Conceptually:

```text
Application Service
      ↓
Transaction
      ↓
Repository A
Repository B
Repository C
      ↓
Commit
```

---

# 81. Repository Responsibility

Repositories should handle:

```text
Persistence
Queries
Updates
```

but should not decide:

```text
When a business transaction starts
```

unless explicitly designed as a unit-of-work abstraction.

---

# 82. Unit of Work

A unit-of-work abstraction may be used to coordinate repositories participating in the same transaction.

Conceptually:

```text
UnitOfWork
 ├── RideRepository
 ├── DriverRepository
 └── OutboxRepository
```

---

# 83. Transaction Helper

Infrastructure may provide a transaction helper that:

```text
Begins Transaction
Executes Callback
Commits on Success
Rolls Back on Error
```

The helper must correctly handle panic/error paths.

---

# 84. Transaction Error Handling

A failed transaction should return a meaningful error category.

Examples:

```text
Deadlock
SerializationFailure
UniqueViolation
ForeignKeyViolation
ConnectionFailure
Timeout
```

---

# 85. Unique Constraint Handling

A unique constraint violation may represent:

```text
Expected Concurrency Conflict
```

rather than a generic server error.

The application should translate it appropriately.

---

# 86. Transaction Timeout

Transactions should have bounded execution time.

A transaction that exceeds its expected duration should be observable and investigated.

---

# 87. Context Propagation

Go database operations should receive the request context.

Conceptually:

```text
HTTP Request
 ↓
Application Context
 ↓
DB Transaction
```

Cancellation should propagate where safe.

---

# 88. Context Cancellation

If the request is cancelled before commit:

```text
Rollback
```

should normally occur.

Do not leave an abandoned transaction open.

---

# 89. Panic Safety

Transaction wrappers must guarantee rollback when a panic occurs.

---

# 90. Commit Errors

Commit itself can fail.

The application must treat:

```text
Commit Error
```

as a transaction failure unless the database semantics establish otherwise.

---

# 91. Retry Boundaries

Retry the smallest safe unit.

Prefer:

```text
Retry Whole Transaction
```

for transaction serialization/deadlock failures.

Avoid:

```text
Retry Individual SQL Statement
```

when doing so could violate transaction semantics.

---

# 92. Retry Count

Retries must be:

```text
Bounded
```

Typical implementation values should be configured and tested rather than copied blindly.

---

# 93. Exponential Backoff

Retryable transaction failures should use:

```text
Backoff
+
Jitter
```

where appropriate.

---

# 94. External Retry vs Transaction Retry

These are different mechanisms.

```text
Database Transaction Retry
```

handles transactional conflicts.

```text
Provider Retry
```

handles external provider failures.

Do not combine them into an uncontrolled retry loop.

---

# 95. Nested Transactions

PostgreSQL does not provide independent nested transactions in the same sense as application-level transactions.

Use:

```text
Savepoints
```

only when a specific workflow requires partial rollback within a transaction.

---

# 96. Savepoints

Savepoints should not become a default substitute for clean transaction boundaries.

---

# 97. Long Workflows

A workflow such as:

```text
Ride Creation
→ Matching
→ Driver Acceptance
→ Pickup
→ Trip
→ Payment
```

must not be one database transaction.

Each state transition has its own local transaction.

---

# 98. Saga-Like Workflow

The ride lifecycle is therefore coordinated through:

```text
Local Transactions
+
Events
+
State Machines
+
Compensating Actions
```

rather than one global transaction.

---

# 99. Example: Driver Acceptance

```text
Driver Accept Request
        ↓
BEGIN
        ↓
Validate Ride State
        ↓
Validate Driver State
        ↓
Commit Acceptance
        ↓
Outbox Event
        ↓
COMMIT
```

Other services react asynchronously.

---

# 100. Example: Ride Cancellation

```text
BEGIN
 ↓
Validate Cancellation
 ↓
Update Ride State
 ↓
Create RideCancelled Event
 ↓
COMMIT
```

Matching/driver/notification consumers react afterward.

---

# 101. Compensation

When a downstream operation fails, do not attempt to roll back a previously committed distributed transaction.

Instead use a compensating domain action.

Example:

```text
Reservation Created
 ↓
Downstream Failure
 ↓
Release Reservation
```

---

# 102. Compensation Must Be Explicit

Compensating actions should be:

```text
Idempotent
Auditable
State-Aware
```

---

# 103. Compensation Is Not Rollback

Important distinction:

```text
Database Rollback
```

means:

```text
Transaction Never Committed
```

while:

```text
Compensation
```

means:

```text
A New State Change Corrects a Previously Committed State
```

---

# 104. Payment Example

A payment workflow should not hold a PostgreSQL transaction open while waiting for an external payment gateway.

Instead:

```text
Create Payment Intent
 ↓
Commit
 ↓
Call Gateway
 ↓
Receive Result
 ↓
Update Payment State
```

---

# 105. External Provider Example

For ETA:

```text
Load Required State
 ↓
Commit / End DB Transaction
 ↓
Call Route Provider
 ↓
Persist Result If Needed
```

Do not keep ride database locks while waiting for the provider.

---

# 106. AI Inference Example

Do not hold a database transaction while waiting for:

```text
AI Model
```

Instead:

```text
Read Candidate State
 ↓
End Transaction
 ↓
Run Inference
 ↓
Validate Current State
 ↓
Short Assignment Transaction
```

---

# 107. Stale Decision Protection

Because external computation may take time:

```text
AI Ranking
```

or:

```text
ETA Calculation
```

may become stale before assignment.

The final transaction should revalidate critical state.

---

# 108. Compare-and-Set

Use compare-and-set style operations where appropriate:

```text
Current State = EXPECTED
```

before applying the transition.

---

# 109. Reservation

If a dispatch workflow requires temporary reservation:

```text
Candidate Selected
 ↓
Reserve Driver
 ↓
Short TTL
 ↓
Accept / Release
```

The reservation state must have explicit ownership and expiry semantics.

---

# 110. Reservation Transaction

Creating or consuming a reservation should be atomic within the owning database.

---

# 111. Reservation Expiry

An expired reservation should not remain authoritative indefinitely.

The system may use:

```text
TTL
Scheduled Cleanup
Lazy Expiry
State Validation
```

depending on the implementation.

---

# 112. Transactional Messaging

The preferred pattern is:

```text
DB State
+
Outbox
```

not:

```text
DB State
+
Direct Kafka Publish
```

inside the same application transaction.

---

# 113. Outbox Publisher

The publisher should:

```text
Read Pending Outbox
 ↓
Publish Event
 ↓
Record Publication
```

with idempotent/retry-safe behaviour.

---

# 114. Publisher Failure

If publication fails:

```text
Keep Outbox Pending
```

and retry.

---

# 115. Poison Event

If an event repeatedly fails publication or processing:

```text
DLQ / Operational Handling
```

should be used according to:

```text
ADR-0013
```

---

# 116. Event Consumer Transaction

A consumer should generally make:

```text
State Change
+
Processed Event Marker
```

atomic.

---

# 117. Duplicate Event

If the same event is delivered twice:

```text
First Delivery
→ Apply

Second Delivery
→ Detect Duplicate
→ No Duplicate Side Effect
```

---

# 118. Event Processing Order

If a consumer requires strict ordering:

```text
Aggregate Key
+
Partitioning
+
Version Validation
```

should be used.

---

# 119. Eventual Consistency and User Experience

The UI must not assume that every projection updates instantly.

For example:

```text
Ride Created
```

may be immediately visible in the ride service while:

```text
Operations Dashboard
```

updates shortly afterward.

---

# 120. User-Facing State

The authoritative service should be used for critical user-facing state where immediate consistency matters.

---

# 121. Operational Dashboard

Operational dashboards may use eventually consistent projections where:

```text
Low Latency
```

is more valuable than:

```text
Perfect Immediate Consistency
```

provided the dashboard clearly reflects its freshness requirements.

---

# 122. Data Consistency Matrix

| Data | Consistency | Owner | Typical Mechanism |
|---|---|---|---|
| Ride lifecycle | Strong | Ride domain | PostgreSQL transaction |
| Assignment | Strong | Ride/dispatch domain | Transaction + constraints |
| Driver availability | Strong within owner | Driver domain | PostgreSQL transaction |
| Driver location | Real-time / specialized | Location subsystem | Redis / location store |
| Cache | Eventually consistent | Cache layer | Cache invalidation |
| Analytics | Eventual | Analytics | Events / pipelines |
| AI features | Eventual | AI platform | Event/data pipeline |
| Notifications | Eventual | Notification subsystem | Events |
| Search/read projections | Eventual | Projection | Events |
| Outbox | Strong with aggregate | Owning service | Same DB transaction |

---

# 123. Cross-Service Transaction Matrix

| Operation | Local Transaction | Cross-Service Event |
|---|---:|---:|
| Create Ride | Yes | Yes |
| Update Ride State | Yes | Yes |
| Driver Acceptance | Yes | Yes |
| Ride Completion | Yes | Yes |
| Notification | No global transaction | Yes |
| Analytics Update | No global transaction | Yes |
| AI Feature Update | No global transaction | Yes |
| External Payment | Local intent/state | Yes / callback |
| ETA Provider Call | No long transaction | Optional |
| Driver Location Broadcast | No global transaction | Optional |

---

# 124. Data Ownership Rule

Every mutable business entity must have a clearly defined owner.

Example:

```text
Ride
→ Ride Domain

Driver
→ Driver Domain

Payment
→ Payment Domain

Location
→ Location Subsystem
```

---

# 125. Single Writer Principle

Where practical, one domain owner should be responsible for authoritative writes to an aggregate.

This reduces:

```text
Concurrent Ownership
State Conflicts
Hidden Coupling
```

---

# 126. No Direct Cross-Service Database Writes

A service should not update another service's tables directly.

Instead:

```text
Command / API
```

or:

```text
Event
```

should cross the service boundary.

---

# 127. Commands vs Events

Use:

```text
Command
```

when asking another service to perform an action.

Use:

```text
Event
```

when notifying that something already happened.

---

# 128. Transaction Boundary and API Calls

An API call to another service should not automatically become part of the caller's database transaction.

Instead:

```text
Commit Local State
 ↓
Send Command / Event
 ↓
Remote Service Processes
```

with retry/idempotency controls.

---

# 129. Synchronous Service Call

A synchronous call may still be used when the caller needs an immediate response.

However:

```text
Caller DB Transaction
```

should generally not remain open while waiting for the remote service.

---

# 130. Distributed Consistency

When a synchronous dependency is required:

```text
Prepare
 ↓
Call Dependency
 ↓
Short Local Transaction
```

or:

```text
Local Transaction
 ↓
Commit
 ↓
Call Dependency
```

should be preferred over holding locks across the network.

---

# 131. Transactional Inbox

For critical consumers, an inbox/processed-event pattern may be used:

```text
Event
 ↓
Inbox
 ↓
Transaction
 ↓
State Change
```

This is useful where duplicate event delivery must be controlled.

---

# 132. Inbox + Outbox

A service may use:

```text
Inbox
+
Business State
+
Outbox
```

inside one local transaction.

This creates a reliable local processing boundary.

---

# 133. Example

```text
BEGIN
 ├── Record Incoming Event
 ├── Update Local State
 └── Create Outgoing Event
COMMIT
```

Then:

```text
Outgoing Event
→ Kafka / Redpanda
```

---

# 134. Exactly-Once Business Effect

The architecture aims for:

```text
Exactly-Once Business Effect
```

rather than claiming:

```text
Exactly-Once Network Delivery
```

This is achieved through:

```text
Idempotency
Constraints
Versioning
Transactional State Changes
```

---

# 135. Database Constraints as Final Guard

Application code may have race conditions.

Database constraints provide the final integrity boundary.

---

# 136. Example Race

Two workers:

```text
Worker A → Assign Driver X
Worker B → Assign Driver X
```

Both may see:

```text
Driver Available
```

before either writes.

The database must prevent an invalid final state.

---

# 137. Concurrency Strategy

For critical resources:

```text
Validate
+
Lock / Conditional Update
+
Constraint
```

should be used as appropriate.

---

# 138. Read-Modify-Write Risk

Avoid:

```text
SELECT Available
↓
Application Decision
↓
UPDATE
```

without concurrency protection.

Prefer:

```text
Conditional UPDATE
```

or:

```text
SELECT FOR UPDATE
```

where required.

---

# 139. Transaction Isolation Does Not Replace Domain Rules

A stronger isolation level cannot compensate for missing business constraints.

Both are required where appropriate.

---

# 140. Transaction Boundary Documentation

Every critical application service should document:

```text
Transaction Start
State Read
State Write
Outbox Write
Commit
External Calls
Retry Behaviour
```

---

# 141. Development Rule

Developers should not introduce a database transaction merely because:

```text
Multiple SQL Queries Exist
```

The transaction should exist because:

```text
Those Operations Must Commit or Roll Back Together
```

---

# 142. Transaction Scope Review

During code review, ask:

```text
What invariant does this transaction protect?
What state changes must be atomic?
Is the transaction too large?
Are there external calls inside it?
Can it deadlock?
Can it be retried?
Is the operation idempotent?
```

---

# 143. Performance

Transaction performance depends on:

```text
Duration
Lock Contention
Rows Touched
Indexes
Isolation Level
Connection Pool
Query Plan
```

---

# 144. Avoid Large Transactions

Do not process thousands of unrelated records inside a single transaction unless batch atomicity is explicitly required.

---

# 145. Batch Processing

Large background jobs should use bounded batches where possible:

```text
Batch 1
→ Commit

Batch 2
→ Commit

Batch 3
→ Commit
```

---

# 146. Transaction Metrics

Track:

```text
Transaction Count
Transaction Duration
Rollback Count
Deadlock Count
Serialization Failure Count
Lock Wait Time
```

---

# 147. Long Transaction Alert

A long-running transaction should be observable and alertable when it exceeds the expected operational threshold.

---

# 148. Database Lock Monitoring

Monitor:

```text
Blocked Queries
Lock Waits
Deadlocks
Long Transactions
```

to identify transaction design problems.

---

# 149. PgBouncer Considerations

When PgBouncer is used, transaction handling must match the configured pooling mode.

Application code should not assume that:

```text
Database Session
```

remains permanently tied to one physical PostgreSQL connection outside the transaction.

---

# 150. Prepared Statements

Database access patterns should be compatible with the chosen PgBouncer mode and driver configuration.

---

# 151. Connection Lifecycle

Application code should:

```text
Acquire
→ Use
→ Commit/Rollback
→ Release
```

connections promptly.

---

# 152. Transaction Leak

A transaction that is not committed or rolled back can hold:

```text
Connection
Locks
Snapshots
```

and eventually degrade the entire database.

---

# 153. Transaction Cleanup

All transaction paths must guarantee:

```text
Commit
OR
Rollback
```

---

# 154. Failure Scenarios

## Scenario 1 — Database Failure

```text
Transaction
→ DB Failure
→ Rollback / Connection Error
```

No partial committed state should be assumed.

---

## Scenario 2 — Kafka Failure

```text
DB Transaction
→ Commit
→ Outbox Pending
→ Kafka Retry
```

---

## Scenario 3 — Consumer Crash

```text
Event
→ Consumer Starts Transaction
→ Crash Before Commit
→ Event Redelivered
→ Idempotent Processing
```

---

## Scenario 4 — Duplicate Request

```text
Request
→ Transaction
→ Commit

Retry
→ Idempotency Check
→ No Duplicate Side Effect
```

---

## Scenario 5 — Concurrent Assignment

```text
Worker A ─┐
          ├→ Concurrency Control
Worker B ─┘
          ↓
One Valid Assignment
```

---

# 155. Failure and Degradation

Transaction failures should follow:

```text
ADR-0021 — Failure and Degradation Strategy
```

The transaction layer must not invent independent failure semantics that conflict with the platform-wide strategy.

---

# 156. Idempotency Relationship

Retry safety is governed by:

```text
ADR-0020 — Idempotency Strategy
```

Transactions alone do not guarantee safe retries.

---

# 157. Outbox Relationship

Database/event consistency is governed by:

```text
ADR-0012 — Outbox Pattern
```

---

# 158. Dead Letter Relationship

Persistent event-processing failures are governed by:

```text
ADR-0013 — Dead Letter Queue Strategy
```

---

# 159. Observability Relationship

Transaction and consistency metrics are governed by:

```text
ADR-0022 — Observability Strategy
```

---

# 160. Security Relationship

Transaction data and cross-service identifiers must follow:

```text
ADR-0023 — Security and Secret Management
```

---

# 161. Configuration Relationship

Transaction timeouts, retry limits, and related operational configuration follow:

```text
ADR-0024 — Configuration and Environment Strategy
```

---

# 162. Testing Relationship

Transaction, concurrency, and integration testing follow:

```text
ADR-0025 — Testing and Integration Strategy
```

---

# 163. Cost Relationship

Transaction and database efficiency should align with:

```text
ADR-0028 — Cost Optimization Strategy
```

---

# 164. Alternatives Considered

## 164.1 Global Distributed Transactions

### Advantages

```text
Strong Cross-Service Atomicity
```

### Disadvantages

```text
High Complexity
Poor Availability
Operational Overhead
Tight Coupling
Long Transaction Duration
```

### Decision

```text
Rejected.
```

---

# 165. Direct Database Updates Across Services

### Advantages

```text
Simple to Implement Initially
```

### Disadvantages

```text
Broken Ownership
Hidden Coupling
Hard-to-Control Transactions
Migration Risk
```

### Decision

```text
Rejected.
```

---

# 166. Direct Kafka Publish Inside DB Transaction

### Advantages

```text
Simple Mental Model
```

### Disadvantages

```text
DB/Kafka Atomicity Problem
Lost Events
Duplicate Events
Long Transactions
```

### Decision

```text
Rejected.
```

Use the outbox pattern.

---

# 167. Eventual Consistency Everywhere

### Advantages

```text
High Decoupling
High Availability
```

### Disadvantages

```text
Unsafe for Critical State
Duplicate Assignment Risk
Inconsistent Ownership
```

### Decision

```text
Rejected.
```

Use strong local consistency where required and eventual consistency across boundaries.

---

# 168. SERIALIZABLE Everywhere

### Advantages

```text
Strong Isolation
```

### Disadvantages

```text
Higher Contention
More Retries
Lower Throughput
Unnecessary Complexity
```

### Decision

```text
Rejected.
```

Use the weakest isolation level that safely satisfies the business invariant.

---

# 169. Redis as Primary Transactional Store

### Advantages

```text
Low Latency
High Throughput
```

### Disadvantages

```text
Poor Fit for Core Relational Transactions
Durability / Querying Trade-Offs
Complexity
```

### Decision

```text
Rejected for core transactional ride state.
```

---

# 170. Consequences

## 170.1 Positive Consequences

The decision provides:

```text
Clear Transaction Boundaries
Strong Local Consistency
Reliable Event Intent
Safe Distributed Evolution
Controlled Concurrency
Better Failure Recovery
Provider Independence
```

---

## 170.2 Negative Consequences

The architecture introduces:

```text
Eventual Consistency
Outbox Management
Idempotency Requirements
Concurrency Handling
More Complex Failure Scenarios
Additional Observability
```

These trade-offs are accepted.

---

# 171. Risks

## Risk 1 — Incorrect Transaction Boundary

### Mitigation

```text
Document Invariants
Review Transactions
Use Domain Ownership
```

---

## Risk 2 — Long Transactions

### Mitigation

```text
No External Calls
Short Queries
Lock Monitoring
Transaction Timeouts
```

---

## Risk 3 — Duplicate Events

### Mitigation

```text
Idempotent Consumers
Event IDs
Inbox / Processed Event Tracking
```

---

## Risk 4 — Duplicate Assignment

### Mitigation

```text
Conditional Updates
Locks
Database Constraints
Final Revalidation
```

---

## Risk 5 — Stale Read Model

### Mitigation

```text
Authoritative Reads
Event Versioning
Projection Monitoring
```

---

## Risk 6 — Retry Storm

### Mitigation

```text
Bounded Retries
Backoff
Jitter
Failure Classification
```

---

# 172. Validation

This ADR should be validated through:

```text
Transaction Unit Tests
Concurrency Tests
Deadlock Tests
Serialization Failure Tests
Outbox Tests
Duplicate Event Tests
Idempotency Tests
Assignment Race Tests
Failure Injection
Load Tests
Long Transaction Tests
```

---

# 173. Review Triggers

Revisit this ADR when:

```text
A New Service Boundary Is Added
Database Topology Changes
A New Messaging System Is Added
Transaction Contention Increases
A Distributed Transaction Is Proposed
Assignment Semantics Change
Payment Architecture Changes
Database Sharding Is Introduced
A New Consistency Requirement Appears
```

---

# 174. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
adr/
```

Especially:

```text
Database Development
Event and Messaging Development
Error Handling and Validation
Performance and Optimization
Integration Testing
Observability Development
```

---

# 175. Related ADRs

This decision is directly related to:

```text
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
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
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
ADR-0024 — Configuration and Environment Strategy
ADR-0025 — Testing and Integration Strategy
ADR-0028 — Cost Optimization Strategy
```

---

# 176. Decision Summary

RideForge adopts:

```text
Local ACID Transactions
        +
Explicit Domain Boundaries
        +
Database Constraints
        +
Outbox Pattern
        +
Idempotent Consumers
        +
Controlled Concurrency
        +
Eventual Consistency Across Services
```

The architecture does **not** attempt to make:

```text
PostgreSQL
+
Redis
+
Kafka
+
External APIs
+
AI Services
```

one global transaction.

Instead:

```text
                     Local Transaction
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
       Business State                 Outbox
             │                           │
             └──────────── COMMIT ──────┘
                                         │
                                         ▼
                                  Kafka / Redpanda
                                         │
                              ┌──────────┼──────────┐
                              ▼          ▼          ▼
                           Service A  Service B  Service C
                              │          │          │
                         Local Tx    Local Tx    Local Tx
```

---

# 177. Final Principle

> **RideForge will use strong local ACID transactions to protect critical domain invariants and use events, outbox delivery, idempotency, and explicit state machines to coordinate eventually consistent operations across service boundaries.**

The fundamental rule is:

```text
One Business Invariant
        ↓
One Owning Transaction Boundary
```

and:

```text
Cross-Service Coordination
        ↓
Events / Commands
        ↓
Local Transactions
```

not:

```text
Global Distributed Transaction
```

---

# 178. Status

```text
Decision: ACCEPTED

Core Transaction Model:
Local ACID

Cross-Service Consistency:
Eventual

Distributed Transactions:
Not Used

Outbox:
Required for Transactional Domain Events

Consumer Processing:
Idempotent

Critical Invariants:
Database + Application Protected

Assignment Consistency:
Strong

External Calls Inside Critical DB Transactions:
Avoided

Default Isolation:
READ COMMITTED

Stronger Isolation:
Only When Justified

Concurrency:
Explicitly Managed

Retry:
Bounded + Idempotent

Global Rollback:
Not Used

Compensation:
Used for Committed Distributed Workflows

Primary Goal:
Preserve Critical Business Invariants Without Introducing Distributed Transaction Complexity
```
