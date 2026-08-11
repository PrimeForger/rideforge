# ADR-0020: Idempotency Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Reliability / Distributed Systems / API and Event Processing  
> **Scope:** API requests, commands, ride lifecycle operations, event consumers, retries, duplicate delivery, and distributed workflows  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is an event-driven ride-hailing platform operating across:

```text
Mobile / Web Clients
        ↓
API Services
        ↓
Application Services
        ↓
PostgreSQL
        ↓
Outbox
        ↓
Kafka / Redpanda
        ↓
Consumers
        ↓
Other Services
```

Distributed systems routinely experience:

```text
Network Retries
Client Retries
HTTP Timeouts
Message Redelivery
Consumer Restarts
Worker Restarts
Database Failures
Kafka / Redpanda Failures
Load Balancer Retries
Mobile Connectivity Problems
```

As a result, the same logical operation may be submitted or delivered more than once.

Examples:

```text
Create Ride
Accept Ride
Cancel Ride
Assign Driver
Start Trip
Complete Trip
Payment Initiation
Notification Request
Event Consumption
```

RideForge must therefore ensure that retries and duplicate delivery do not create duplicate business effects.

---

# 2. Problem

Without an explicit idempotency strategy, the same logical operation can produce:

```text
Two Rides
Two Driver Assignments
Two Payments
Two Cancellations
Two Notifications
Two Trip Completions
Duplicate Database Records
Duplicate Domain Events
```

A simple retry such as:

```text
Request
   ↓
Server Processes
   ↓
Network Timeout
   ↓
Client Retries
```

does not necessarily mean the first request failed.

The server may have already committed the operation.

Therefore:

```text
Timeout
≠
Operation Failed
```

This is the central reliability problem addressed by this ADR.

---

# 3. Decision

RideForge will adopt **idempotency as a first-class reliability mechanism** across APIs, commands, event consumers, and critical asynchronous workflows.

The platform will use:

```text
Idempotency Keys
+
Database Uniqueness
+
Processed Event Tracking
+
Conditional State Transitions
+
Version Checks
+
Deterministic Retry Behaviour
+
Bounded Retries
```

The fundamental principle is:

> **The same logical operation must produce the same business effect regardless of how many times that operation is retried or delivered.**

---

# 4. Idempotency Definition

An operation is idempotent when repeating the same logical operation does not create additional unintended business effects.

Conceptually:

```text
Operation(X)
+
Operation(X)
+
Operation(X)
```

must result in the same intended business state as:

```text
Operation(X)
```

---

# 5. Idempotency Is Not the Same as Duplicate HTTP Requests

Two requests may look identical but represent different operations.

For example:

```text
POST /rides
```

with the same pickup and destination could represent:

```text
Request A
```

and:

```text
Request B
```

from the same user at different times.

Therefore idempotency requires a stable identifier for the logical operation.

---

# 6. Idempotency Key

For operations requiring request-level idempotency, the client should provide an idempotency key.

Conceptually:

```text
Idempotency-Key: <unique-operation-key>
```

The key represents:

```text
One Logical Operation
```

not:

```text
One User
```

and not:

```text
One Endpoint
```

---

# 7. Key Scope

An idempotency key must be scoped appropriately.

A useful logical scope is:

```text
Tenant / Account
+
Operation
+
Idempotency Key
```

The exact scope depends on the API.

---

# 8. Key Uniqueness

The system must prevent two concurrent requests with the same logical key from creating separate business effects.

The database should provide the final uniqueness guarantee where appropriate.

---

# 9. Idempotency Record

For APIs that require durable request idempotency, RideForge may persist:

```text
idempotency_key
operation
scope
request_hash
status
response_reference
created_at
expires_at
```

The exact schema is implementation-specific.

---

# 10. Request Hash

The original request payload may be associated with the idempotency key.

This allows the system to detect:

```text
Same Key
+
Different Payload
```

which should normally be rejected.

---

# 11. Same Key, Different Payload

Example:

```text
Request 1
Key: ABC
Pickup: Location A
Destination: Location B
```

followed by:

```text
Request 2
Key: ABC
Pickup: Location C
Destination: Location D
```

must not silently reuse the first operation.

The system should return an idempotency conflict.

---

# 12. Idempotency Conflict

Use a dedicated internal error category such as:

```text
IDEMPOTENCY_KEY_REUSED
```

The public API should return a stable, documented error representation.

---

# 13. Key Lifecycle

An idempotency key should have a defined lifecycle:

```text
NEW
 ↓
PROCESSING
 ↓
COMPLETED
```

or:

```text
NEW
 ↓
PROCESSING
 ↓
FAILED
```

depending on implementation.

---

# 14. Processing State

A request may be marked:

```text
PROCESSING
```

while another worker is executing it.

A concurrent request using the same key must not execute the same business operation independently.

---

# 15. Completed State

After successful completion:

```text
COMPLETED
```

the original result may be replayed for a repeated request.

---

# 16. Response Replay

For APIs where appropriate, a duplicate request should receive the same logical result as the original request.

Conceptually:

```text
First Request
→ Process
→ Store Result

Retry
→ Find Existing Result
→ Return Existing Result
```

---

# 17. Response Storage

The platform does not need to store every complete response body indefinitely.

Depending on the endpoint, it may store:

```text
Response Reference
Resource ID
Status
Result Metadata
```

and reconstruct the current response.

---

# 18. Idempotency TTL

Idempotency records should have an explicit retention period.

The TTL should reflect:

```text
Client Retry Window
Network Behaviour
Business Risk
Operational Requirements
Storage Cost
```

---

# 19. Do Not Use Unlimited Idempotency Storage

Indefinite retention of every idempotency key creates unnecessary storage growth.

Use:

```text
Bounded Retention
+
Cleanup
```

where business semantics allow it.

---

# 20. Critical Operations

The following operations should receive strong idempotency protection where applicable:

```text
Ride Creation
Driver Acceptance
Ride Cancellation
Trip Start
Trip Completion
Payment Initiation
Refund Initiation
Driver Assignment
Important Administrative Commands
```

The exact list depends on the final service/API design.

---

# 21. Ride Creation

Ride creation is a primary idempotency use case.

Without idempotency:

```text
User Taps Book
       ↓
Request Times Out
       ↓
User Taps Again
       ↓
Two Rides
```

With idempotency:

```text
User Taps Book
       ↓
Operation Key
       ↓
Ride Created
       ↓
Retry
       ↓
Existing Ride Returned
```

---

# 22. Ride Creation Identity

The ride itself receives a separate:

```text
ride_id
```

The idempotency key is the identifier of the creation operation.

Do not confuse:

```text
idempotency_key
```

with:

```text
ride_id
```

---

# 23. Ride ID vs Idempotency Key

```text
Idempotency Key
→ Identifies a Request / Operation

Ride ID
→ Identifies the Domain Entity
```

One successful operation may produce:

```text
Idempotency Key
→ Ride ID
```

---

# 24. Driver Acceptance

A driver acceptance operation must be idempotent.

Repeated acceptance should not:

```text
Create Multiple Acceptance Records
```

or:

```text
Emit Multiple Business Effects
```

unless duplicate events are intentionally supported.

---

# 25. Conditional Acceptance

Driver acceptance should verify the expected ride state.

Conceptually:

```text
Accept only if:
ride.status = SEARCHING
```

or the applicable state.

This provides:

```text
Idempotency
+
Concurrency Protection
```

---

# 26. Ride Cancellation

Repeated cancellation should be deterministic.

Example:

```text
Cancel
→ CANCELLED

Cancel Again
→ Already Cancelled
```

The second request should not trigger a second cancellation side effect.

---

# 27. Cancellation Side Effects

Side effects such as:

```text
Driver Release
Notification
Refund Request
Analytics Event
```

must also be protected from duplication.

---

# 28. Trip Start

Trip start should be idempotent.

Repeated start requests should not create:

```text
Multiple Start Timestamps
Multiple Trip Sessions
Multiple Start Events
```

---

# 29. Trip Completion

Trip completion is especially sensitive because it may trigger:

```text
Fare Calculation
Payment
Invoice
Driver Earnings
Analytics
```

The completion operation must therefore have strong idempotency and state-transition protection.

---

# 30. Payment Initiation

Payment initiation should use an idempotency key or provider-supported equivalent.

Never assume:

```text
Payment Request Timeout
→ Payment Did Not Happen
```

---

# 31. Payment Provider Idempotency

If a payment provider supports idempotency keys, RideForge should propagate a stable operation identifier where appropriate.

The provider key and RideForge key should have a documented relationship.

---

# 32. Refunds

Refund operations must be idempotent.

A retry must not create:

```text
Two Refunds
```

for the same logical refund operation.

---

# 33. Driver Assignment

Assignment must be protected against duplicate commands.

Example:

```text
Assign Ride R
→ Driver D
```

Repeated assignment command:

```text
Assign Ride R
→ Driver D
```

must not create a second conflicting assignment.

---

# 34. Assignment Idempotency

Assignment correctness should combine:

```text
Idempotency
+
State Validation
+
Concurrency Control
+
Database Constraints
```

---

# 35. Idempotency Does Not Replace Concurrency Control

Idempotency handles:

```text
Same Logical Operation
```

Concurrency control handles:

```text
Competing Operations
```

These are different problems.

---

# 36. Example

Two workers may receive:

```text
Worker A → Assign Driver X
Worker B → Assign Driver Y
```

These may have different idempotency keys.

Idempotency alone cannot decide which assignment wins.

The domain state machine and concurrency controls must decide.

---

# 37. Database Constraints

Database constraints provide a final integrity boundary.

Examples:

```text
UNIQUE
PARTIAL UNIQUE
CHECK
FOREIGN KEY
```

where appropriate.

---

# 38. Event Idempotency

Kafka/Redpanda consumers must assume that an event may be delivered more than once.

Therefore:

```text
Event Processing
```

must be idempotent.

---

# 39. Event ID

Every important domain event should have a unique:

```text
event_id
```

The event ID is used to detect duplicate delivery.

---

# 40. Event ID vs Aggregate ID

Do not confuse:

```text
event_id
```

with:

```text
aggregate_id
```

Example:

```text
event_id
→ uniquely identifies this event

ride_id
→ identifies the ride aggregate
```

A ride can generate many events:

```text
RideCreated
RideMatched
RideStarted
RideCompleted
```

---

# 41. Processed Event Store

A consumer may maintain a processed-event record:

```text
consumer_name
event_id
processed_at
```

with a uniqueness constraint.

---

# 42. Consumer Transaction

A safe consumer pattern is:

```text
BEGIN
 ↓
Check event_id
 ↓
If already processed → stop
 ↓
Apply state change
 ↓
Record event_id
 ↓
COMMIT
```

---

# 43. Duplicate Event

Example:

```text
RideCompleted
event_id = E123
```

First delivery:

```text
Process
→ Update State
→ Record E123
→ Commit
```

Second delivery:

```text
Find E123
→ Already Processed
→ No Duplicate Effect
```

---

# 44. Event Processing Race

Two workers may process the same event simultaneously.

The database uniqueness constraint should protect the processed-event record.

Conceptually:

```text
UNIQUE(consumer, event_id)
```

---

# 45. Insert-First Pattern

A consumer may use an atomic insert to claim the event.

Conceptually:

```text
INSERT event_id
ON CONFLICT DO NOTHING
```

If insertion succeeds:

```text
Process
```

If insertion conflicts:

```text
Duplicate
```

The exact transaction design must ensure the event is not marked processed before its business effect is safely committed.

---

# 46. Inbox Pattern

For more complex consumers, RideForge may use an inbox:

```text
Incoming Event
      ↓
Inbox
      ↓
Local Transaction
      ↓
Business State
```

This is compatible with the data consistency strategy defined in:

```text
ADR-0019
```

---

# 47. Event + Outbox

A service that consumes one event and produces another may use:

```text
Inbox
+
Business State
+
Outbox
```

in one local transaction.

Conceptually:

```text
BEGIN
 ├── Record / Claim Event
 ├── Update Local State
 └── Create Outbox Event
COMMIT
```

---

# 48. Exactly-Once Delivery

RideForge will not depend on universal:

```text
Exactly-Once Delivery
```

across the entire distributed platform.

Instead:

```text
At-Least-Once Delivery
+
Idempotent Processing
```

will provide the required business semantics.

---

# 49. Exactly-Once Business Effect

The target is:

```text
Exactly-Once Business Effect
```

rather than:

```text
Exactly-Once Network Delivery
```

---

# 50. Retry Classification

Every retryable operation should be classified as:

```text
Safe to Retry
Unsafe to Retry
Retry Only With Idempotency
Retry Only After State Check
```

---

# 51. Retryable Errors

Typical retryable conditions include:

```text
Transient Network Failure
Connection Reset
Database Serialization Failure
Database Deadlock
Temporary Kafka Failure
Temporary Provider Failure
```

provided the operation is safe to retry.

---

# 52. Non-Retryable Errors

Examples:

```text
Validation Failure
Unauthorized
Forbidden
Invalid State Transition
Idempotency Conflict
Permanent Business Rule Failure
```

These should normally not be blindly retried.

---

# 53. Retry Boundary

Retry the smallest safe unit.

For a database transaction:

```text
Retry Whole Transaction
```

when required.

For an API request:

```text
Retry Entire Idempotent Operation
```

rather than replaying partial internal steps.

---

# 54. Retry Count

Retries must be:

```text
Bounded
```

Never use:

```text
Infinite Retry
```

---

# 55. Backoff

Use:

```text
Exponential Backoff
+
Jitter
```

for appropriate transient failures.

---

# 56. Retry Storm Prevention

Idempotency does not prevent retry storms by itself.

The platform must also use:

```text
Backoff
Jitter
Circuit Breaking Where Appropriate
Rate Limits
Bounded Retries
```

---

# 57. Client-Side Retries

Mobile and web clients may retry requests because of:

```text
Timeout
Connection Loss
Backgrounding
Network Switching
```

Critical APIs must therefore be designed to tolerate safe retries.

---

# 58. API Gateway Retries

If infrastructure performs automatic retries, the operation must remain safe.

Automatic retry behaviour should not be enabled indiscriminately for non-idempotent endpoints.

---

# 59. HTTP Methods

HTTP semantics provide useful defaults:

```text
GET
→ Naturally Idempotent

PUT
→ Intended to Be Idempotent

DELETE
→ Should Be Idempotent at the Resource Semantics Level

POST
→ Not Naturally Idempotent
```

Therefore POST operations that create important business effects should use explicit idempotency where appropriate.

---

# 60. GET

GET requests should not create business side effects.

---

# 61. PUT

PUT should represent replacement/update semantics that can safely be repeated.

---

# 62. DELETE

Repeated deletion should normally produce the same final state.

However, associated side effects must also be protected.

---

# 63. POST

POST operations that create:

```text
Ride
Payment
Refund
Booking
Assignment
```

should use explicit idempotency when retry safety is required.

---

# 64. API Contract

The API documentation should clearly specify:

```text
Whether Idempotency Is Required
Whether Idempotency Is Optional
Key Format
Key Scope
Key Lifetime
Conflict Behaviour
Response Replay Behaviour
```

---

# 65. Missing Idempotency Key

For an endpoint requiring idempotency:

```text
Missing Key
→ Validation Error
```

Do not silently generate an unpredictable key on behalf of the client if doing so could make retries unsafe.

---

# 66. Optional Idempotency

Some endpoints may allow optional idempotency.

The endpoint documentation must define:

```text
What Happens With Key
What Happens Without Key
```

---

# 67. Key Format

Keys should be:

```text
Unique
Opaque
Non-Sensitive
Stable
```

Do not encode sensitive business information in the key.

---

# 68. Key Entropy

Keys must have sufficient uniqueness to avoid accidental collisions.

UUID-based identifiers are a suitable default where appropriate.

---

# 69. Do Not Use Predictable Sequential Keys

Avoid keys such as:

```text
1
2
3
```

when they can collide across clients or expose operational information.

---

# 70. Key Ownership

The client generating the key owns its logical operation identity.

The server owns:

```text
Validation
Storage
Execution
Replay
Expiry
```

---

# 71. Idempotency Scope

An idempotency key should not accidentally collide across:

```text
Different Users
Different Tenants
Different Operations
Different Endpoints
```

unless intentionally scoped that way.

---

# 72. Multi-Tenant Systems

Where RideForge services support tenant boundaries, idempotency records must respect tenant isolation.

Conceptually:

```text
tenant_id
+
operation
+
idempotency_key
```

---

# 73. Security

Idempotency keys must not be treated as authorization credentials.

Possession of a key must not grant access to the underlying resource.

Authorization must still be checked.

---

# 74. Response Replay Authorization

When replaying a stored idempotent result, the current caller must still satisfy authorization requirements.

Do not allow a leaked idempotency key to bypass access control.

---

# 75. Sensitive Responses

Avoid storing sensitive response bodies unnecessarily.

Prefer:

```text
Resource ID
+
Status
+
Safe Result Metadata
```

when the response can be reconstructed securely.

---

# 76. Idempotency Record Cleanup

Expired idempotency records may be cleaned up through:

```text
Scheduled Job
Partitioning
TTL-Based Storage
Background Cleanup
```

depending on the database architecture.

---

# 77. Cleanup Safety

Cleanup must not occur while the key may still reasonably be retried.

---

# 78. Idempotency and Database Transactions

The idempotency record should be coordinated with the business operation.

Preferred conceptual flow:

```text
BEGIN
 ↓
Create / Claim Idempotency Record
 ↓
Perform Business Change
 ↓
Create Outbox Event
 ↓
COMMIT
```

---

# 79. First Request Failure

If the transaction rolls back:

```text
Business Change
→ Rolled Back
```

the idempotency state must not incorrectly indicate:

```text
SUCCESS
```

---

# 80. Processing Crash

If the process crashes:

```text
After Business Commit
Before Response
```

a retry must discover the committed result through durable state.

---

# 81. Processing State Recovery

A stuck:

```text
PROCESSING
```

record must have a recovery mechanism.

Possible mechanisms:

```text
Lease
Timeout
Heartbeat
Recovery Worker
State Reconciliation
```

---

# 82. Do Not Leave Permanent PROCESSING

A request should not remain indefinitely in:

```text
PROCESSING
```

because a worker crashed.

---

# 83. Lease-Based Processing

For long operations, a processing record may contain:

```text
lease_until
worker_id
updated_at
```

so another worker can safely recover after the lease expires.

---

# 84. Avoid Duplicate Execution During Lease

A new worker should only take over when the previous lease is safely expired.

---

# 85. Idempotency and State Machines

State machines are a major part of idempotency.

Example:

```text
SEARCHING
→ MATCHED
```

Repeated:

```text
SEARCHING → MATCHED
```

should not produce another assignment if the state is already:

```text
MATCHED
```

---

# 86. Terminal States

Terminal states should be explicitly handled.

Examples:

```text
COMPLETED
CANCELLED
EXPIRED
```

A repeated operation against a terminal state should produce deterministic behaviour.

---

# 87. No-Op Idempotency

Some repeated operations may safely become:

```text
NO-OP
```

Example:

```text
Cancel already-cancelled ride
→ No additional state change
```

---

# 88. Idempotent Event Handlers

Event handlers should be designed around:

```text
Current State
+
Event Identity
+
Event Version
```

rather than assuming every event arrives exactly once.

---

# 89. Stale Events

An event may be duplicated or stale.

Consumers should use:

```text
Version Checks
State Validation
Event Ordering
```

where necessary.

---

# 90. Event Versioning

Important events should include an aggregate version when ordering/state evolution matters.

Example:

```text
ride_version = 12
```

---

# 91. Version Guard

A consumer may reject or ignore an event if:

```text
event.version <= current.version
```

where this rule matches the aggregate's semantics.

---

# 92. Idempotency and Outbox

The outbox guarantees:

```text
Business State
+
Event Intent
```

are committed together.

Idempotency guarantees:

```text
Repeated Operation
```

does not create repeated business effects.

These solve different problems.

---

# 93. Idempotency and DLQ

A duplicate event should normally not enter the DLQ merely because it was already processed.

Instead:

```text
Duplicate
→ Ignore / Acknowledge
```

according to consumer semantics.

---

# 94. DLQ Use

The DLQ is for:

```text
Poison Messages
Permanent Processing Failures
Malformed Events
Unrecoverable Business Processing Errors
```

not ordinary duplicate delivery.

---

# 95. Idempotency and AI

AI workflows can also be retried.

Examples:

```text
ETA Prediction
Dispatch Ranking
Demand Prediction
Feature Generation
```

If the operation produces persistent side effects, it should have a deterministic operation identity.

---

# 96. AI Inference

Pure inference without side effects is naturally easier to retry:

```text
Same Input
+
Same Model Version
+
Same Feature Snapshot
```

can produce the same logical prediction subject to model determinism.

---

# 97. AI Side Effects

If AI inference triggers:

```text
Assignment
Driver Notification
Stored Decision
```

the resulting business operation must use the normal idempotency and concurrency controls.

AI itself must not bypass them.

---

# 98. Notification Idempotency

Notifications may be duplicated if an event is redelivered.

Important notifications should have a stable logical notification identity.

Example:

```text
ride_id
+
notification_type
+
recipient
```

or an explicit event/notification ID.

---

# 99. Email / SMS / Push

External notification providers may have their own deduplication mechanisms.

RideForge should use provider-supported idempotency or application-level deduplication where available.

---

# 100. WhatsApp Notifications

WhatsApp message workflows should similarly distinguish:

```text
Message Request
```

from:

```text
Message Delivery
```

and protect against duplicate sends where business requirements demand it.

---

# 101. Webhooks

Inbound webhooks from external providers must be treated as potentially duplicated.

Every webhook handler should identify the provider's event/reference ID where available.

---

# 102. Webhook Processing

Preferred pattern:

```text
Webhook
 ↓
Validate Signature
 ↓
Identify Event
 ↓
Check Processed State
 ↓
Apply State Change
 ↓
Record Event
 ↓
COMMIT
```

---

# 103. Webhook Duplicate

A duplicate webhook should normally return a successful acknowledgement after confirming that the event has already been processed.

---

# 104. Provider Timeout

If an external provider request times out:

```text
Unknown Outcome
```

must be assumed when the provider may have processed the request.

Do not blindly retry without idempotency.

---

# 105. Unknown Outcome

This is a critical distributed-systems state:

```text
Request Sent
+
Response Not Received
```

The platform cannot assume:

```text
Not Processed
```

---

# 106. Recovery From Unknown Outcome

Use:

```text
Idempotency Key
Provider Query
Status Reconciliation
Webhook
```

where available.

---

# 107. Reconciliation

For important external operations, periodic reconciliation may verify:

```text
Ride State
Payment State
Provider State
```

and repair discrepancies through controlled domain actions.

---

# 108. Reconciliation Is Not Blind Replay

A reconciliation worker should inspect:

```text
Current State
Provider State
Operation Identity
```

before applying changes.

---

# 109. Idempotency in Background Jobs

Background jobs must also be idempotent.

Examples:

```text
Expire Ride
Release Driver
Generate Invoice
Calculate Earnings
Send Notification
Update Projection
```

---

# 110. Job Identity

A job should have a stable logical identity when duplicate execution is possible.

---

# 111. Scheduled Jobs

A scheduled job may run more than once because of:

```text
Worker Restart
Leader Change
Scheduler Retry
Deployment
Clock / Scheduling Issues
```

Therefore critical scheduled work should be idempotent.

---

# 112. Distributed Cron

Do not assume:

```text
Only One Scheduler
```

unless the infrastructure guarantees it.

Use:

```text
Lease
Database Lock
Unique Job Key
```

where required.

---

# 113. Cleanup Jobs

Cleanup operations should be safe to execute repeatedly.

Example:

```text
DELETE expired records
```

or:

```text
Mark expired
```

should produce the same final state when repeated.

---

# 114. Idempotency and Caching

Cache writes may be repeated safely in most cases, but cache updates must not become the authority for critical state.

---

# 115. Idempotency and Redis Locks

A Redis lock is not a substitute for idempotency.

Use:

```text
Lock
+
Idempotency
+
Database Constraint
```

where the workflow requires all three.

---

# 116. API Idempotency Table

A conceptual table may be:

| Field | Purpose |
|---|---|
| `scope` | Tenant/account/operation scope |
| `operation` | Logical operation |
| `idempotency_key` | Client operation identity |
| `request_hash` | Detect payload mismatch |
| `status` | Processing state |
| `resource_id` | Created/affected resource |
| `created_at` | Creation time |
| `expires_at` | Retention boundary |

---

# 117. Unique Constraint

A suitable uniqueness rule may be:

```text
UNIQUE(scope, operation, idempotency_key)
```

The exact database design depends on service ownership and API semantics.

---

# 118. Event Processing Table

A conceptual consumer table:

| Field | Purpose |
|---|---|
| `consumer_name` | Consumer identity |
| `event_id` | Unique event identity |
| `processed_at` | Processing timestamp |
| `aggregate_id` | Related entity |
| `version` | Optional event version |

---

# 119. Event Unique Constraint

A suitable constraint may be:

```text
UNIQUE(consumer_name, event_id)
```

---

# 120. Idempotency Error Taxonomy

Recommended internal categories:

```text
IDEMPOTENCY_KEY_MISSING
IDEMPOTENCY_KEY_REUSED
IDEMPOTENCY_REQUEST_MISMATCH
IDEMPOTENCY_IN_PROGRESS
IDEMPOTENCY_EXPIRED
DUPLICATE_EVENT
STALE_EVENT
```

Not every error must be exposed directly to clients.

---

# 121. API Status Behaviour

The exact HTTP status codes should be defined by the API standard, but the semantics should distinguish:

```text
New Operation
Existing Successful Operation
Operation In Progress
Key Conflict
Invalid Request
```

---

# 122. In-Progress Request

If a second request arrives while the first operation is:

```text
PROCESSING
```

the system may:

```text
Wait
Return In-Progress
Return Retryable Response
```

depending on endpoint semantics.

It must not execute the business operation independently.

---

# 123. Concurrency on Same Key

Two identical requests may arrive simultaneously:

```text
Request A ─┐
           ├→ Same Idempotency Key
Request B ─┘
```

Only one should become the authoritative executor.

---

# 124. Atomic Claim

The first worker must atomically claim the key.

Possible mechanisms:

```text
Unique Constraint
INSERT
ON CONFLICT
Row Lock
```

The database should provide the final guarantee.

---

# 125. Request Hash Canonicalization

If request hashes are used, canonicalization must be deterministic.

Equivalent JSON payloads should not accidentally produce different hashes merely because:

```text
Field Order
Whitespace
Serialization Format
```

differs.

---

# 126. Do Not Hash Sensitive Data Blindly

If request payloads contain sensitive information, hashing/storage must follow the security and privacy policies.

---

# 127. Idempotency Key Logging

Logs should record a safe identifier/reference.

Avoid logging sensitive payloads alongside the key.

---

# 128. Trace Correlation

An idempotency key may be associated with:

```text
request_id
trace_id
ride_id
event_id
```

for observability.

These identifiers have different meanings and should not be conflated.

---

# 129. Identifier Model

```text
request_id
→ One incoming request attempt

idempotency_key
→ One logical client operation

event_id
→ One domain event

aggregate_id
→ One domain entity

trace_id
→ One distributed execution trace
```

---

# 130. Idempotency Metrics

Track:

```text
Idempotent Request Count
Duplicate Request Count
Idempotency Conflict Count
In-Progress Collision Count
Duplicate Event Count
Duplicate Webhook Count
Retry Count
```

---

# 131. Operational Signals

Unexpected spikes may indicate:

```text
Network Problems
Client Bugs
Load Balancer Retry Configuration
Consumer Instability
Provider Timeouts
```

---

# 132. Alerting

Potential alerts:

```text
High Duplicate Request Rate
High Idempotency Conflict Rate
High Stuck Processing Rate
High Duplicate Event Rate
High Provider Unknown-Outcome Rate
```

---

# 133. Testing

Idempotency must be tested as a first-class behaviour.

Required test categories:

```text
Unit Tests
Integration Tests
Concurrency Tests
Retry Tests
Failure Injection
Event Redelivery Tests
Webhook Duplicate Tests
```

---

# 134. API Idempotency Test

Test:

```text
Request 1
→ Success

Request 2 with Same Key
→ Same Logical Result
```

---

# 135. Payload Conflict Test

Test:

```text
Request 1
Key = A
Payload = X

Request 2
Key = A
Payload = Y

Expected:
Conflict
```

---

# 136. Concurrent Request Test

Run:

```text
100 Requests
Same Idempotency Key
```

Expected:

```text
One Business Operation
```

with deterministic responses for the rest.

---

# 137. Consumer Duplicate Test

Deliver:

```text
Event E1
Event E1
Event E1
```

Expected:

```text
One Business Effect
```

---

# 138. Crash Test

Simulate:

```text
Commit
+
Process Crash
+
Retry
```

Expected:

```text
No Duplicate Business Effect
```

---

# 139. Outbox Duplicate Test

Simulate:

```text
Publish Event
+
Publisher Crash Before Marking Published
+
Republish
```

Expected:

```text
Consumer Handles Duplicate Safely
```

---

# 140. Webhook Duplicate Test

Deliver the same provider webhook multiple times.

Expected:

```text
One Business Effect
```

---

# 141. Unknown Outcome Test

Simulate:

```text
Provider Receives Request
Provider Response Lost
Client Retries
```

Expected:

```text
No Duplicate External Operation
```

where provider idempotency is supported.

---

# 142. State Transition Test

Repeat:

```text
Cancel
Cancel
Cancel
```

Expected:

```text
One Valid Terminal State
```

and no repeated irreversible side effects.

---

# 143. Database Constraint Test

Verify that concurrent requests cannot violate:

```text
Unique Business Invariants
```

even if application-level coordination fails.

---

# 144. Load Testing

Test idempotency under:

```text
High Request Volume
High Retry Rate
Consumer Restart
Database Contention
```

---

# 145. Chaos Testing

Inject:

```text
Network Failure
Kafka Failure
Database Connection Failure
Worker Crash
Provider Timeout
```

and verify retry safety.

---

# 146. Performance

Idempotency must not become a major request bottleneck.

Use:

```text
Indexed Keys
Bounded Records
Efficient Lookups
Appropriate TTL
```

---

# 147. Database Indexing

Idempotency lookup columns must be indexed according to the uniqueness and query pattern.

---

# 148. Storage Growth

Monitor:

```text
Idempotency Record Count
Processed Event Record Count
Cleanup Lag
Storage Size
```

---

# 149. Cleanup Monitoring

A cleanup failure should not silently cause unbounded growth.

---

# 150. Hot Keys

A single idempotency key may become highly contended if a client retries aggressively.

The system should avoid expensive repeated work for such keys.

---

# 151. Rate Limiting

Rate limits may protect against abusive or malfunctioning clients generating excessive retries.

---

# 152. Idempotency and Rate Limits

Rate limiting and idempotency solve different problems:

```text
Rate Limit
→ Controls Request Volume

Idempotency
→ Controls Duplicate Business Effects
```

Both may be required.

---

# 153. Idempotency and Authentication

Idempotency does not replace authentication or authorization.

Every request must still be authorized.

---

# 154. Multi-Region Considerations

If RideForge eventually operates across multiple regions, idempotency records must have a clearly defined ownership and consistency strategy.

Do not assume a local idempotency store provides global uniqueness.

---

# 155. Multi-Region Strategy

Before active-active multi-region operation, define:

```text
Idempotency Key Ownership
Global Key Routing
Replication
Conflict Resolution
Regional Failover
```

---

# 156. Current Architecture

For the current architecture, prefer:

```text
Single Authoritative Operation Owner
```

rather than prematurely implementing globally distributed idempotency.

---

# 157. Migration Considerations

When extracting a service:

```text
Existing Idempotency Semantics
```

must remain compatible.

Do not accidentally reset operation identity during migration.

---

# 158. Backward Compatibility

API changes must preserve:

```text
Existing Idempotency-Key Semantics
```

or introduce an explicit versioned change.

---

# 159. API Versioning

If idempotency behaviour changes materially:

```text
API Version
```

or another explicit compatibility mechanism should be used.

---

# 160. Documentation Requirements

Each critical API must document:

```text
Idempotency Requirement
Key Format
Key Scope
TTL
Conflict Semantics
Retry Semantics
Response Replay
```

---

# 161. Developer Rule

Developers must ask:

```text
Can This Operation Be Retried?
Can It Be Delivered Twice?
Can It Time Out After Commit?
Can Two Workers Execute It?
```

If the answer to any is yes, idempotency must be considered explicitly.

---

# 162. Code Review Checklist

Reviewers should verify:

```text
[ ] Operation identity is explicit
[ ] Duplicate execution is safe
[ ] Database constraints exist where needed
[ ] State transitions are conditional
[ ] Events have unique IDs
[ ] Consumers are idempotent
[ ] Retries are bounded
[ ] Unknown outcomes are handled
[ ] External calls use provider idempotency where available
[ ] Sensitive data is not stored unnecessarily
```

---

# 163. Consequences

## 163.1 Positive Consequences

The decision provides:

```text
Safe Client Retries
Safe Consumer Redelivery
Duplicate Protection
Better Failure Recovery
Reliable Event Processing
Safer External Provider Integration
Stronger Ride Lifecycle Integrity
```

---

## 163.2 Negative Consequences

The architecture introduces:

```text
Additional Database State
Idempotency Record Cleanup
More Complex API Contracts
Request Hashing
Consumer Deduplication
Additional Testing
Operational Monitoring
```

These trade-offs are accepted.

---

# 164. Risks

## Risk 1 — Incorrect Key Scope

### Mitigation

```text
Explicit Scope Definition
API Documentation
Integration Tests
```

---

## Risk 2 — Key Reuse With Different Payload

### Mitigation

```text
Request Hash
Conflict Detection
```

---

## Risk 3 — Stuck Processing State

### Mitigation

```text
Lease
Timeout
Recovery Worker
```

---

## Risk 4 — Duplicate Event Effects

### Mitigation

```text
Event ID
Processed Event Store
Database Constraint
```

---

## Risk 5 — Retry Storm

### Mitigation

```text
Backoff
Jitter
Rate Limits
Bounded Retries
```

---

## Risk 6 — False Idempotency

An operation may appear idempotent but still trigger duplicate side effects.

### Mitigation

Test the complete workflow:

```text
State Change
+
Event
+
Notification
+
Payment
+
Analytics
```

rather than only the primary database record.

---

# 165. Alternatives Considered

## 165.1 No Explicit Idempotency

### Advantages

```text
Simple
Less Storage
```

### Disadvantages

```text
Duplicate Rides
Duplicate Payments
Duplicate Events
Unsafe Retries
```

### Decision

```text
Rejected.
```

---

# 166. Client-Only Idempotency

### Advantages

```text
Less Server Complexity
```

### Disadvantages

```text
Server Cannot Guarantee Safety
Consumer Retries Still Exist
External Webhooks Still Duplicate
```

### Decision

```text
Rejected.
```

---

# 167. Redis-Only Idempotency

### Advantages

```text
Fast
Simple TTL
```

### Disadvantages

```text
Not Suitable as Sole Durable Guarantee for Critical Business Operations
Failure / Eviction Risks
Cross-System Consistency Problems
```

### Decision

```text
Rejected as the sole mechanism for critical operations.
```

Redis may support low-risk coordination where appropriate.

---

# 168. Kafka Exactly-Once as the Complete Solution

### Advantages

```text
Strong Messaging Semantics
```

### Disadvantages

```text
Does Not Solve API Retries
Does Not Solve External Providers
Does Not Replace Database Constraints
Does Not Make All Business Effects Globally Exactly Once
```

### Decision

```text
Rejected as the sole idempotency strategy.
```

---

# 169. Global Distributed Lock

### Advantages

```text
Can Prevent Some Concurrent Executions
```

### Disadvantages

```text
Complexity
Failure Modes
Latency
Lock Expiry Problems
Does Not Replace Durable Business Identity
```

### Decision

```text
Rejected as the primary idempotency mechanism.
```

---

# 170. Validation

This ADR should be validated through:

```text
API Retry Tests
Concurrent Request Tests
Duplicate Event Tests
Webhook Replay Tests
Provider Timeout Tests
Database Constraint Tests
Outbox Tests
Consumer Restart Tests
Chaos Tests
Load Tests
```

---

# 171. Review Triggers

Revisit this ADR when:

```text
A New Critical API Is Added
A New Payment Provider Is Added
Multi-Region Active/Active Is Introduced
A New Messaging System Is Added
A New External Webhook Integration Is Added
Idempotency Storage Becomes a Bottleneck
A Duplicate-Effect Incident Occurs
API Retry Semantics Change
```

---

# 172. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
adr/
```

Especially:

```text
API Development
Event and Messaging Development
Database Development
Error Handling and Validation
Testing Strategy
Integration Testing
Performance and Optimization
Observability Development
```

---

# 173. Related ADRs

This decision is directly related to:

```text
ADR-0003 — Microservice Boundaries
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0007 — PostgreSQL as Primary Database
ADR-0009 — Redis for Real-Time State and Caching
ADR-0012 — Outbox Pattern
ADR-0013 — Dead Letter Queue Strategy
ADR-0014 — API and Service Communication
ADR-0015 — Smart Dispatch and Stand Dispatch
ADR-0016 — AI-Assisted Dispatch Strategy
ADR-0017 — ETA and Route Provider Strategy
ADR-0018 — Regional and Legal Ride Validation
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
ADR-0023 — Security and Secret Management
ADR-0024 — Configuration and Environment Strategy
ADR-0025 — Testing and Integration Strategy
ADR-0026 — Model and AI Governance
ADR-0028 — Cost Optimization Strategy
```

---

# 174. Decision Summary

RideForge adopts:

```text
                         Operation
                             │
                             ▼
                    Idempotency Identity
                             │
                    ┌────────┴────────┐
                    ▼                 ▼
              New Operation      Existing Key
                    │                 │
                    ▼                 ▼
              Claim Atomically    Validate Request
                    │                 │
                    ▼                 ▼
             Business Transaction   Replay / Status
                    │
          ┌─────────┴─────────┐
          ▼                   ▼
    Business State          Outbox
          │                   │
          └──────── COMMIT ───┘
                              │
                              ▼
                       Event Delivery
                              │
                              ▼
                       Consumer Inbox /
                     Processed Event State
                              │
                              ▼
                     Idempotent Effect
```

The platform will rely on:

```text
Idempotency Keys
Database Constraints
Conditional State Transitions
Event IDs
Processed Event Tracking
Outbox Pattern
Bounded Retries
Backoff + Jitter
Provider Idempotency
Reconciliation
```

---

# 175. Final Principles

The following principles are mandatory:

```text
1. A timeout does not prove that an operation failed.

2. Critical retryable APIs must have explicit operation identity.

3. Idempotency keys identify logical operations, not resources.

4. The same idempotency key must not silently represent different payloads.

5. Database constraints provide the final integrity boundary.

6. Idempotency does not replace concurrency control.

7. Event consumers must tolerate duplicate delivery.

8. Every important event must have a unique event identity.

9. Outbox and idempotency solve different reliability problems.

10. External provider calls must account for unknown outcomes.

11. Payment and refund operations require strong duplicate protection.

12. Background jobs must be safe against duplicate execution.

13. Retries must be bounded and use appropriate backoff.

14. PROCESSING states must have recovery semantics.

15. Idempotency keys must not be treated as authorization credentials.

16. Sensitive response data should not be retained unnecessarily.

17. Multi-region global idempotency must be explicitly designed before active-active operation.

18. The target is exactly-once business effect, not universal exactly-once delivery.
```

---

# 176. Status

```text
Decision: ACCEPTED

API Idempotency:
Required for Critical Retryable Operations

Idempotency Keys:
Supported

Request Hash:
Used Where Payload Conflict Detection Is Required

Database Constraints:
Required for Critical Invariants

Event Idempotency:
Required

Event IDs:
Required

Processed Event Tracking:
Supported

Outbox:
Required for Transactional Domain Events

External Provider Idempotency:
Used Where Supported

Retries:
Bounded

Backoff:
Required for Appropriate Retry Classes

Unknown Provider Outcomes:
Explicitly Handled

Global Distributed Transactions:
Not Used

Redis:
Not the Sole Durable Idempotency Guarantee

Exactly-Once Delivery:
Not Assumed

Exactly-Once Business Effect:
Target

Primary Goal:
Ensure Retries, Duplicate Requests, and Duplicate Event Delivery Do Not Produce Duplicate Business Effects
```
