# ADR-0013: Dead Letter Queue Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Event Reliability / Failure Handling  
> **Scope:** RideForge event-processing failures that cannot be successfully processed through normal retry mechanisms  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge uses an event-driven architecture based on:

```text
Go Services
PostgreSQL
Redis
Redpanda
Transactional Outbox
Asynchronous Consumers
```

The architecture depends on asynchronous event processing for workflows such as:

```text
Ride Events
Driver Events
Matching Events
Dispatch Events
Payment Events
Notification Events
Operational Events
AI / Data Pipeline Events
```

Normal event processing is expected to tolerate transient failures.

However, some events may continue to fail even after retries.

Examples include:

```text
Invalid Event Payload
Unsupported Schema
Malformed Data
Permanent Business Validation Failure
Unexpected Consumer Bug
Incompatible Event Version
Poison Message
```

If such an event is retried indefinitely, it can cause:

```text
Consumer Blocking
Repeated Processing
Resource Exhaustion
Increasing Lag
Operational Instability
```

RideForge therefore requires a controlled mechanism for isolating events that cannot be successfully processed through normal retry handling.

---

# 2. Problem

The event-processing architecture must distinguish between:

```text
Temporary Failure
```

and:

```text
Permanent / Unrecoverable Failure
```

The system must provide a controlled path for failed events without silently losing them.

The solution must support:

```text
Failure Isolation
Bounded Retry
Operational Visibility
Manual Investigation
Controlled Replay
Consumer Recovery
Poison Message Handling
Auditability
```

---

# 3. Decision

RideForge will use:

> **A Dead Letter Queue (DLQ) strategy for events that cannot be successfully processed after the configured retry policy or are determined to be non-retryable.**

The conceptual flow is:

```text
Redpanda Topic
      ↓
Consumer
      ↓
Processing
   ┌──┴───────────────┐
   │                  │
Success            Failure
   │                  │
   ▼                  ▼
Commit            Retry
                      │
                 ┌────┴────┐
                 │         │
              Success   Exhausted /
                         Permanent
                            │
                            ▼
                           DLQ
                            │
                            ▼
                     Investigation
                            │
                  ┌─────────┴─────────┐
                  │                   │
                Replay              Discard
              After Fix          Only If Explicitly
                                   Approved
```

---

# 4. Core Principle

> **A DLQ is an exception-handling mechanism, not a normal event-processing path.**

The normal lifecycle should remain:

```text
Event
  ↓
Consumer
  ↓
Process
  ↓
Success
```

The DLQ should be used only when normal processing cannot safely continue.

---

# 5. Failure Classification

RideForge will classify event-processing failures into at least:

```text
Transient
Permanent
Unknown
```

### Transient

A failure that may succeed later.

Examples:

```text
Temporary Broker Failure
Temporary Database Failure
Network Timeout
Temporary Dependency Failure
Resource Exhaustion
```

Expected action:

```text
Retry
```

### Permanent

A failure unlikely to succeed without changing the event or system.

Examples:

```text
Invalid Payload
Unsupported Schema
Malformed Data
Known Business Rejection
```

Expected action:

```text
DLQ / Quarantine
```

### Unknown

A failure whose recoverability cannot immediately be determined.

Expected action:

```text
Bounded Retry
+
Observation
```

If retries remain unsuccessful:

```text
DLQ
```

---

# 6. Why a DLQ Is Required

Without a DLQ:

```text
Poison Event
    ↓
Retry
    ↓
Retry
    ↓
Retry
    ↓
Retry Forever
```

This can block healthy event processing.

With a DLQ:

```text
Poison Event
    ↓
Retry
    ↓
Failure Threshold
    ↓
DLQ
    ↓
Healthy Events Continue
```

---

# 7. DLQ and Outbox Relationship

ADR-0012 establishes the Transactional Outbox for reliable publication.

The responsibilities are different:

```text
Outbox
→ Reliable Publication From PostgreSQL

DLQ
→ Isolation of Events That Cannot Be Successfully Processed
```

A typical producer-side flow is:

```text
PostgreSQL
   ↓
Outbox
   ↓
Redpanda
   ↓
Consumer
   ↓
Retry
   ↓
DLQ
```

---

# 8. Producer-Side vs Consumer-Side Failures

The DLQ strategy primarily addresses:

```text
Consumer Processing Failures
```

Producer publication failures are primarily handled through:

```text
Outbox Retry
```

For example:

```text
PostgreSQL
   ↓
Outbox
   ↓
Redpanda Unavailable
   ↓
Outbox Retry
```

rather than immediately placing the event into a consumer DLQ.

---

# 9. Consumer DLQ

A consumer DLQ isolates an event that the consumer could not safely process.

Example:

```text
ride.events
     ↓
Matching Consumer
     ↓
Processing Failure
     ↓
Retry
     ↓
DLQ
```

---

# 10. DLQ Topic

Each DLQ should have a predictable naming convention.

Conceptually:

```text
<source-topic>.dlq
```

Examples:

```text
ride.events.dlq
match.events.dlq
driver.events.dlq
```

The exact naming convention should follow RideForge event and messaging standards.

---

# 11. DLQ Message Envelope

A DLQ message should preserve the original event and add failure metadata.

Conceptually:

```text
original_event
failure_metadata
```

Failure metadata may include:

```text
event_id
event_type
source_topic
source_partition
source_offset
consumer
consumer_group
failure_reason
failure_class
attempt_count
failed_at
first_failed_at
last_failed_at
correlation_id
```

---

# 12. Original Event Preservation

The original event payload should be preserved where operationally safe and appropriate.

This allows:

```text
Investigation
Replay
Root-Cause Analysis
```

without reconstructing the original event manually.

---

# 13. Failure Metadata

Failure metadata should make the event independently diagnosable.

At minimum, operators should be able to determine:

```text
What failed?
Where did it fail?
Which consumer failed?
How many attempts occurred?
When did it fail?
Why did it fail?
```

---

# 14. Stable Event Identity

The original:

```text
event_id
```

must remain unchanged when the event is moved to the DLQ.

This preserves identity across:

```text
Producer
Redpanda
Consumer
Retries
DLQ
Replay
```

---

# 15. Retry Count

The DLQ record should retain the number of processing attempts.

Conceptually:

```text
attempt_count = N
```

This helps distinguish:

```text
Immediate Permanent Failure
```

from:

```text
Repeated Transient Failure
```

---

# 16. Retry Strategy

Normal consumer processing should use bounded retries.

Conceptually:

```text
Event
 ↓
Attempt 1
 ↓
Failure
 ↓
Backoff
 ↓
Attempt 2
 ↓
Failure
 ↓
Backoff
 ↓
Attempt N
 ↓
DLQ
```

The exact retry count and delays are configuration decisions.

---

# 17. Exponential Backoff

Retries should generally use increasing delays.

Conceptually:

```text
Attempt 1 → Short Delay
Attempt 2 → Longer Delay
Attempt 3 → Longer Delay
Attempt N → DLQ
```

Jitter should be considered to avoid synchronized retry storms.

---

# 18. Retryable Failures

Examples include:

```text
Temporary Database Unavailability
Temporary Network Failure
Transient Dependency Failure
Temporary Resource Exhaustion
Temporary Service Unavailability
```

These should normally remain in the retry path.

---

# 19. Non-Retryable Failures

Examples include:

```text
Malformed Payload
Invalid Required Field
Unsupported Event Version
Invalid Serialization
Known Permanent Business Rejection
```

These may bypass repeated retries and move directly to DLQ.

---

# 20. Unknown Failures

Unexpected errors should initially be treated conservatively.

A reasonable strategy is:

```text
Unknown Error
    ↓
Bounded Retry
    ↓
Still Failing
    ↓
DLQ
```

This prevents an unexpected temporary problem from becoming a permanent message loss.

---

# 21. Poison Message

A poison message is an event that repeatedly causes consumer failure.

Example:

```text
Event E123
   ↓
Attempt 1 → Failure
Attempt 2 → Failure
Attempt 3 → Failure
Attempt 4 → Failure
```

If the failure is caused by the event itself:

```text
E123 → DLQ
```

---

# 22. Poison Message Isolation

A poison message must not indefinitely prevent healthy events from being processed.

Therefore:

```text
Poison Event
→ Isolate
```

while:

```text
Healthy Events
→ Continue
```

according to the required ordering guarantees.

---

# 23. Ordering Considerations

DLQ handling must respect event ordering requirements.

For streams where ordering is important:

```text
Event A
Event B
Event C
```

and B fails, blindly processing C may violate business semantics.

The consumer must determine whether:

```text
C Can Continue
```

or:

```text
C Must Wait
```

based on the aggregate/event contract.

---

# 24. Aggregate Ordering

For aggregate-keyed events such as ride lifecycle events:

```text
RideCreated
    ↓
DriverAssigned
    ↓
RideStarted
    ↓
RideCompleted
```

a failure in an earlier event may require special handling.

The DLQ strategy therefore does not automatically imply that later events can always proceed.

---

# 25. DLQ Does Not Solve Ordering

A DLQ isolates failures.

It does not automatically solve:

```text
Ordering
Causality
Business Dependencies
State Synchronization
```

Those concerns remain part of consumer design.

---

# 26. Consumer Acknowledgement

An event should be acknowledged only after successful processing.

Conceptually:

```text
Receive Event
     ↓
Process
     ↓
Success
     ↓
Commit / Acknowledge
```

If processing fails:

```text
Do Not Acknowledge
```

until the retry/DLQ mechanism has safely handled the failure.

---

# 27. Redpanda Consumer Model

RideForge uses Redpanda as the event-streaming platform.

The consumer flow should conceptually be:

```text
Redpanda
   ↓
Consumer Group
   ↓
Receive Event
   ↓
Process
   ↓
Success → Commit Offset
   ↓
Failure → Retry / DLQ
```

---

# 28. Offset Commit Principle

The consumer should not advance its source offset in a way that causes an unprocessed event to be silently lost.

The ordering of:

```text
Business Commit
+
Offset Commit
```

must be designed carefully.

---

# 29. Database Transaction + Event Processing

A database-backed consumer may use:

```text
BEGIN

Apply Business Change
Record Processed Event

COMMIT
```

followed by the appropriate offset acknowledgement.

This can provide durable idempotency.

---

# 30. Duplicate Processing

DLQ handling does not eliminate duplicates.

A consumer may process an event successfully but fail before its source offset is committed.

The event may then be delivered again.

Therefore:

```text
Consumer Idempotency
```

remains mandatory for business operations where duplicates are unsafe.

---

# 31. DLQ and Idempotency

The relationship is:

```text
Outbox
→ Reliable Publication

Redpanda
→ Event Delivery

Consumer
→ Business Processing

Idempotency
→ Duplicate Safety

DLQ
→ Failed Event Isolation
```

Each mechanism solves a different failure mode.

---

# 32. DLQ Replay

A DLQ must support controlled replay when the underlying problem has been fixed.

Conceptually:

```text
DLQ
 ↓
Investigation
 ↓
Root Cause Fixed
 ↓
Replay
 ↓
Original Topic
 ↓
Consumer
```

---

# 33. Replay Must Be Explicit

DLQ events must not automatically replay indefinitely.

Replay should be:

```text
Controlled
Auditable
Rate-Limited
Observable
```

---

# 34. Replay Identity

A replayed event should preserve its original:

```text
event_id
```

when the event represents the same original business fact.

A separate replay identifier may be added for operational tracing.

---

# 35. Replay Attempts

Replay metadata may track:

```text
replay_count
last_replayed_at
replayed_by
replay_reason
```

This helps prevent repeated uncontrolled replay cycles.

---

# 36. Replay After Code Fix

A common workflow is:

```text
Buggy Consumer
      ↓
Events Fail
      ↓
DLQ
      ↓
Bug Fixed
      ↓
Deploy
      ↓
Replay DLQ
```

This is one of the primary operational benefits of the DLQ.

---

# 37. Replay After Schema Fix

If events failed because of a schema incompatibility:

```text
Schema Updated
      ↓
Consumer Updated
      ↓
Validate
      ↓
Replay
```

The replay must remain compatible with the original event contract.

---

# 38. Replay Rate Limiting

Do not replay an entire large DLQ at maximum speed without checking downstream capacity.

Prefer:

```text
Small Batch
   ↓
Observe
   ↓
Increase Gradually
```

This avoids turning recovery into another outage.

---

# 39. Replay Validation

Before replay:

```text
Confirm Root Cause Fixed
Confirm Consumer Version
Confirm Event Schema
Confirm Dependencies
Confirm Database Capacity
Confirm Broker Capacity
```

---

# 40. Replay Safety

Replay should require confidence that:

```text
Duplicate Processing
```

will not produce incorrect business effects.

Use:

```text
event_id
Idempotency
Business Constraints
```

where required.

---

# 41. Manual DLQ Operations

Operational tooling should eventually support:

```text
List
Inspect
Search
Replay
Quarantine
Archive
Delete
```

with appropriate authorization.

---

# 42. DLQ Access Control

DLQ operations can affect production business state.

Therefore:

```text
View
Replay
Delete
```

should have appropriate permissions.

Replay and deletion should require stronger privileges than read access.

---

# 43. Auditability

Operational actions should be auditable.

Record:

```text
Who
What
When
Why
Which Event
Which DLQ
Which Replay
```

---

# 44. DLQ Deletion

Deleting a DLQ event should be an explicit operational action.

Do not automatically delete failed events immediately after reaching the DLQ.

Retention should provide time for:

```text
Investigation
Recovery
Replay
Audit
```

---

# 45. DLQ Retention

DLQ retention should be longer than normal transient retry state where business requirements justify it.

The retention period should be defined by:

```text
Operational Need
Data Sensitivity
Storage Cost
Compliance
Recovery Requirements
```

---

# 46. Sensitive Data

DLQ messages may contain business or personal information.

Therefore:

```text
Access
Encryption
Logging
Retention
Deletion
```

must follow RideForge security and data-governance requirements.

---

# 47. Avoid Sensitive Logging

The system should not log full DLQ payloads by default.

Prefer logging:

```text
event_id
event_type
aggregate_id
failure_reason
consumer
correlation_id
```

and only the minimum required payload metadata.

---

# 48. DLQ Topic Security

DLQ topics should have appropriate:

```text
Authentication
Authorization
Network Controls
Retention
Encryption
```

according to the event platform's security configuration.

---

# 49. DLQ Naming

A consistent naming strategy should be used.

Recommended conceptual form:

```text
<source-topic>.dlq
```

Examples:

```text
ride.events.dlq
match.events.dlq
ride.events.v1.dlq
```

Avoid inconsistent names across services.

---

# 50. Consumer Group Isolation

DLQ consumers should not accidentally use the same consumer group as the original processing consumer.

Replay tooling should use a dedicated operational process or consumer group.

---

# 51. DLQ Partitioning

DLQ partitioning should generally preserve enough information to support:

```text
Investigation
Ordering
Replay
Scalability
```

When replaying to the original topic, the original partition key should be preserved where appropriate.

---

# 52. Original Topic Metadata

A DLQ record should preserve metadata such as:

```text
source_topic
source_partition
source_offset
```

where available.

This helps operators locate the original event and understand its source position.

---

# 53. Original Headers

Original event headers should be preserved where practical.

They may contain:

```text
Correlation ID
Causation ID
Event Version
Trace Context
Producer Metadata
```

---

# 54. Failure Stack Information

The DLQ metadata may contain an error summary.

For example:

```text
failure_class
failure_code
failure_message
```

Detailed stack traces should be handled carefully because they may contain sensitive information or excessive data.

---

# 55. Failure Code

Where possible, consumers should classify errors using stable failure codes.

Examples:

```text
INVALID_EVENT
UNSUPPORTED_VERSION
DEPENDENCY_TIMEOUT
DATABASE_UNAVAILABLE
BUSINESS_VALIDATION_FAILED
SERIALIZATION_ERROR
```

This makes operations easier to automate.

---

# 56. Error Message Stability

Human-readable error messages may change.

Operational automation should prefer:

```text
failure_code
```

over parsing free-form error text.

---

# 57. DLQ Monitoring

Important metrics include:

```text
DLQ Message Count
DLQ Messages / Second
DLQ Growth Rate
Oldest DLQ Message Age
Retry Count
Replay Count
Failure Categories
```

---

# 58. DLQ Alerting

Alerts should be configured for:

```text
Unexpected DLQ Growth
Critical Event DLQ Entry
High Failure Rate
Large DLQ Backlog
Old DLQ Messages
Repeated Replay Failure
```

---

# 59. Critical vs Non-Critical Events

Not every event has the same business impact.

Events should be classified where useful:

```text
Critical
Important
Operational
Analytical
```

Critical events should receive stronger:

```text
Alerting
Retention
Recovery
Ownership
```

---

# 60. Critical Event Failure

For a critical event:

```text
Event
 ↓
Failure
 ↓
Retry
 ↓
DLQ
```

should trigger operational visibility.

The DLQ should not become a silent storage location.

---

# 61. DLQ Ownership

Every DLQ should have an operational owner.

The owner should be responsible for:

```text
Monitoring
Incident Response
Replay
Retention
Schema Compatibility
Failure Analysis
```

---

# 62. DLQ Runbook

Each important DLQ should have a runbook covering:

```text
What Produces This DLQ?
What Does the Event Mean?
What Failures Are Expected?
How Do We Inspect It?
How Do We Replay?
When Should We Not Replay?
Who Owns It?
```

---

# 63. Root-Cause Analysis

When a DLQ event is investigated, determine:

```text
Event Origin
Consumer Version
Failure Type
Failure Count
Dependency State
Schema Version
Data State
```

Do not replay blindly.

---

# 64. Replay Failure

If replay fails again:

```text
Replay
 ↓
Failure
 ↓
Retry
 ↓
DLQ Again
```

This must be observable.

Avoid automatic infinite replay cycles.

---

# 65. Replay Loop Protection

Track replay metadata.

For example:

```text
replay_count
```

If the same event repeatedly enters the DLQ:

```text
Investigate
```

rather than:

```text
Replay Forever
```

---

# 66. Quarantine

Some events may require quarantine rather than immediate replay.

Quarantine means:

```text
Keep Event
Prevent Normal Processing
Investigate
Decide Explicitly
```

A DLQ can serve as a quarantine mechanism.

---

# 67. DLQ vs Retry Topic

Retry mechanisms and DLQs have different roles.

```text
Retry Topic / Retry State
→ Temporary Failure

DLQ
→ Exhausted / Permanent Failure
```

The architecture may use delayed retry topics or application-level retry depending on the event-processing implementation.

---

# 68. Retry Topic Consideration

If retry delays become complex, dedicated retry topics may be introduced.

Example:

```text
ride.events
     ↓
ride.events.retry.1
     ↓
ride.events.retry.2
     ↓
ride.events.dlq
```

This is optional and should not be introduced unless required.

---

# 69. Initial Strategy

RideForge should prefer the simplest reliable retry implementation initially.

Use:

```text
Consumer Retry
+
Backoff
+
DLQ
```

before introducing multiple retry topics.

---

# 70. Avoid Over-Engineering

The DLQ architecture should remain:

```text
Simple
Predictable
Observable
Recoverable
```

Additional retry infrastructure should be introduced only when actual event volume or latency requirements justify it.

---

# 71. Consumer Isolation

A failing consumer should not unnecessarily affect unrelated consumer groups.

Redpanda consumer groups provide logical isolation.

A DLQ strategy should preserve that separation.

---

# 72. Multiple Consumers

The same event may be consumed by:

```text
Matching
Analytics
Notifications
AI Pipeline
```

Each consumer may have its own failure and DLQ path.

For example:

```text
ride.events
   ├── matching-group
   │      └── matching DLQ
   │
   ├── notification-group
   │      └── notification DLQ
   │
   └── analytics-group
          └── analytics DLQ
```

A failure in one consumer should not automatically move the event out of the source topic for every consumer.

---

# 73. Consumer-Specific DLQ

The DLQ should identify the consumer context.

This is important because:

```text
Event Is Valid Globally
```

but:

```text
Consumer A Cannot Process It
```

The event may still be valid for:

```text
Consumer B
Consumer C
```

---

# 74. DLQ Naming by Consumer

Where necessary, use:

```text
<source-topic>.<consumer>.dlq
```

Example:

```text
ride.events.matching.dlq
```

The exact naming convention should be standardized before implementation.

---

# 75. Shared DLQ

A shared DLQ may be simpler for some deployments.

However, it can make:

```text
Ownership
Filtering
Replay
Monitoring
```

more difficult.

Prefer consumer-specific DLQs when operational isolation is important.

---

# 76. Decision on DLQ Isolation

The initial RideForge strategy is:

> **Prefer consumer-specific DLQs when different consumers have independent failure and replay requirements.**

A shared DLQ may be used for tightly coupled processing paths where it provides a clear operational advantage.

---

# 77. Event Replay Destination

Replay should normally return the event to:

```text
Original Source Topic
```

rather than publishing it permanently into a new business event stream.

This preserves the original event contract.

---

# 78. Replay to New Topic

Replay to a new topic may be appropriate for:

```text
Testing
Migration
Schema Transformation
Special Recovery
```

but should be explicitly controlled.

---

# 79. Replay Transformations

Transforming a DLQ event during replay can change its semantics.

If transformation is required:

```text
Original Event
→ Explicit Migration / Transformation
→ New Versioned Event
```

Do not silently modify production events.

---

# 80. Event Schema Evolution and DLQ

An event may enter the DLQ because:

```text
Consumer Version < Event Version
```

The solution may be:

```text
Upgrade Consumer
```

rather than changing the event.

---

# 81. Backward Compatibility

Consumers should support the event versions they are expected to process.

Schema compatibility reduces unnecessary DLQ traffic.

---

# 82. Deployment Strategy

Consumer deployments should avoid causing unnecessary message failures.

Use:

```text
Backward-Compatible Changes
Rolling Deployment
Schema Compatibility
```

where practical.

---

# 83. Consumer Bug

If a consumer has a code bug:

```text
Events
 ↓
Failures
 ↓
DLQ
```

After the bug is fixed:

```text
Deploy Fix
 ↓
Validate
 ↓
Replay
```

This is a primary DLQ recovery workflow.

---

# 84. Dependency Failure

If a downstream dependency is unavailable:

```text
Consumer
 ↓
Dependency Failure
 ↓
Retry
```

The event should not immediately be placed in the DLQ unless the failure is known to be permanent.

---

# 85. Database Failure

If the consumer's PostgreSQL database is temporarily unavailable:

```text
Retry
```

rather than:

```text
Immediate DLQ
```

assuming the event itself is valid.

---

# 86. Invalid Business State

If the event is valid but the business state does not permit processing:

```text
Business Validation Failure
```

the event may require:

```text
DLQ
```

or a domain-specific compensating workflow.

The DLQ must not become a substitute for proper domain workflow design.

---

# 87. Compensation

Some failures require a compensating action rather than replay.

For example:

```text
External Side Effect
      ↓
Consumer Failure
```

Blind replay may duplicate the side effect.

In such cases:

```text
Investigate
+
Compensate
+
Then Decide Whether Replay Is Safe
```

---

# 88. External Side Effects

Consumers that perform:

```text
Payments
Notifications
Third-Party APIs
Driver Communications
```

must use idempotency controls where replay can repeat external effects.

---

# 89. DLQ and Payments

Payment-related events require special care.

Before replaying:

```text
Payment Event
```

verify whether the original side effect already occurred.

The DLQ must never be treated as proof that an external payment did not happen.

---

# 90. DLQ and Notifications

For notifications:

```text
Notification Sent
```

may have occurred before the consumer failed.

Replay can cause duplicate messages.

Use:

```text
Notification Idempotency
```

where business requirements justify it.

---

# 91. DLQ and Matching

Matching events may affect:

```text
Driver Assignment
Ride State
Dispatch
```

Before replaying a matching event, verify the current ride and driver state.

The event's historical validity does not guarantee that replaying it against current state is safe.

---

# 92. DLQ and AI

AI-related events may enter a DLQ due to:

```text
Model Failure
Feature Missing
Inference Timeout
Schema Mismatch
Dependency Failure
```

AI consumers should still follow the same:

```text
Retry
DLQ
Replay
Idempotency
```

principles.

---

# 93. DLQ and Analytics

Analytical consumers may have different recovery requirements.

A failed analytics event may be replayed after:

```text
Pipeline Fix
Schema Fix
Data Dependency Recovery
```

without necessarily affecting transactional business state.

---

# 94. Data Privacy

DLQ retention must respect data privacy requirements.

If an event contains sensitive information:

```text
Retention
Encryption
Access
Deletion
Replay
```

must follow applicable governance policies.

---

# 95. Data Minimization

The DLQ should preserve enough information for recovery without unnecessarily duplicating sensitive data.

Where possible:

```text
Reference
+
Minimal Required Payload
```

may be preferable to storing redundant sensitive data.

---

# 96. DLQ Storage Cost

DLQ growth can become expensive.

Monitor:

```text
Message Count
Message Size
Retention
Replay Rate
Storage Cost
```

---

# 97. DLQ Cleanup

Cleanup should be policy-driven.

Never delete DLQ messages simply because they are old without considering:

```text
Retention Policy
Incident State
Recovery Need
Compliance
```

---

# 98. Disaster Recovery

DLQ data may be required after a major incident.

The deployment should determine whether DLQ topics are:

```text
Replicated
Backed Up
Retained
Reconstructible
```

according to business criticality.

---

# 99. Cross-Region DLQ

If RideForge later becomes multi-region, define:

```text
Regional DLQ
Global DLQ
Replay Region
Ownership
Failover
```

before implementing cross-region event recovery.

---

# 100. Observability Correlation

DLQ events should remain traceable using:

```text
event_id
correlation_id
causation_id
aggregate_id
```

This allows operators to follow:

```text
Request
→ Event
→ Consumer
→ Failure
→ DLQ
→ Replay
```

---

# 101. Operational Dashboard

A future operations dashboard should expose:

```text
DLQ Count
DLQ Rate
Oldest Event
Top Failure Codes
Top Consumers
Replay Activity
Repeated Failures
Critical Event Failures
```

---

# 102. Alert Severity

Alert severity should depend on:

```text
Event Criticality
Failure Rate
Backlog
Age
Business Impact
```

One failed low-priority analytics event should not necessarily page an operator.

---

# 103. Incident Workflow

A standard DLQ incident workflow is:

```text
Detect
  ↓
Classify
  ↓
Investigate
  ↓
Fix Root Cause
  ↓
Validate Fix
  ↓
Replay
  ↓
Monitor
  ↓
Close Incident
```

---

# 104. Replay Approval

Production replay of critical events should require appropriate operational approval.

At minimum:

```text
Identify Event
Confirm Cause
Confirm Fix
Confirm Replay Safety
Execute Replay
Monitor Result
```

---

# 105. Automated Replay

Automatic replay may be used only for well-understood, transient failure classes.

For example:

```text
Known Temporary Dependency Failure
```

may support controlled automatic recovery.

Unknown or business-sensitive failures should normally remain in the DLQ for investigation.

---

# 106. Replay Limits

Automatic replay should have:

```text
Maximum Attempts
Maximum Rate
Maximum Duration
```

to prevent replay storms.

---

# 107. DLQ Replay Metrics

Track:

```text
Replay Count
Replay Success Rate
Replay Failure Rate
Replay Duration
Events Re-DLQed
```

A high re-DLQ rate indicates the underlying problem remains unresolved.

---

# 108. Re-DLQ Handling

If a replayed event fails again:

```text
DLQ
 ↓
Replay
 ↓
Failure
 ↓
DLQ
```

the system should increment replay/failure metadata.

Repeated failures should trigger investigation rather than unlimited replay.

---

# 109. Manual Event Modification

Operators should generally not edit production event payloads manually.

If modification is necessary:

```text
Create Explicit Corrected Event
```

rather than silently changing the original historical event.

---

# 110. Event Integrity

The DLQ should preserve the original event as faithfully as practical.

The original event should be treated as immutable historical input.

Operational metadata can be added separately.

---

# 111. Audit Trail

Every replay or destructive operation should create an audit record containing:

```text
event_id
operator
timestamp
action
reason
result
```

---

# 112. Testing Strategy

DLQ behaviour must be tested at multiple levels.

### Unit Tests

```text
Failure Classification
Retry Policy
DLQ Metadata
Replay Validation
```

### Integration Tests

```text
Redpanda Consumer
Retry
DLQ Publication
Replay
Offset Handling
```

### Failure Tests

```text
Database Failure
Dependency Failure
Invalid Payload
Consumer Crash
Broker Failure
```

---

# 113. Poison Event Test

Create an event that deterministically fails.

Verify:

```text
Retry Count Increases
Event Eventually Enters DLQ
Healthy Events Continue
DLQ Metadata Is Correct
```

---

# 114. Retryable Failure Test

Simulate:

```text
Temporary Dependency Failure
```

Then recover the dependency.

Verify:

```text
Event Retries
Event Eventually Succeeds
Event Does Not Enter DLQ
```

---

# 115. Replay Test

Create a DLQ event.

Then:

```text
Fix Consumer
 ↓
Replay Event
 ↓
Process Successfully
```

Verify:

```text
Original event_id preserved
Business Effect correct
Replay metadata recorded
```

---

# 116. Duplicate Replay Test

Replay an already successfully processed event.

Verify:

```text
No Unsafe Duplicate Business Effect
```

through consumer idempotency.

---

# 117. Ordering Test

Create ordered events:

```text
E1
E2
E3
```

Make E2 fail.

Verify the consumer behaves according to the defined aggregate-ordering policy.

---

# 118. DLQ Recovery Test

Simulate:

```text
Consumer Bug
 ↓
DLQ
 ↓
Bug Fixed
 ↓
Replay
```

Verify the full recovery process.

---

# 119. Load Test

Test:

```text
High Event Rate
High Failure Rate
Large DLQ
Replay Burst
```

Measure:

```text
Consumer Throughput
Redpanda Lag
DLQ Growth
Replay Throughput
Database Load
```

---

# 120. Security Test

Verify:

```text
Unauthorized DLQ Read
Unauthorized Replay
Unauthorized Delete
Unauthorized Topic Access
```

are denied.

---

# 121. Operational Runbook

When DLQ traffic increases:

```text
1. Identify Affected Consumer
2. Identify Failure Code
3. Check Recent Deployments
4. Check Dependencies
5. Check Event Schema
6. Check Source Topic
7. Check Failure Rate
8. Determine Transient vs Permanent
9. Fix Root Cause
10. Replay Carefully
```

---

# 122. If DLQ Growth Is Sudden

Check:

```text
Recent Deployment
Schema Change
Database Availability
External Dependency
Credential Changes
Redpanda Health
Consumer Configuration
```

---

# 123. If Every Event Enters DLQ

Likely causes include:

```text
Consumer Configuration Error
Schema Incompatibility
Database Connection Failure
Authorization Failure
Broken Deployment
Malformed Event Contract
```

Treat this as a high-severity incident.

---

# 124. If Only One Event Type Fails

Investigate:

```text
Event Schema
Producer Change
Consumer Handler
Domain State
Event Version
```

rather than assuming the entire messaging system is unhealthy.

---

# 125. If DLQ Replay Fails

Do not repeatedly replay.

Instead:

```text
Stop Replay
Inspect Failure
Fix Root Cause
Validate
Replay Again
```

---

# 126. Cost Considerations

DLQ infrastructure introduces:

```text
Storage
Retention
Monitoring
Operational Tooling
Replay Infrastructure
```

However, the cost is justified by the ability to preserve and recover failed events rather than silently losing them or blocking healthy processing.

---

# 127. Alternatives Considered

## 127.1 Infinite Retry

### Advantages

```text
Simple Concept
No Separate DLQ
```

### Disadvantages

```text
Poison Message Blocking
Unbounded Resource Usage
No Operational Isolation
Difficult Recovery
```

### Decision

Rejected.

---

# 128. Retry and Drop

### Advantages

```text
Simple
Low Storage
```

### Disadvantages

```text
Event Loss
No Recovery
Poor Auditability
```

### Decision

Rejected for business-critical events.

---

# 129. Manual Database Recovery

### Advantages

```text
No Messaging Infrastructure
```

### Disadvantages

```text
Unsafe
Difficult to Audit
Manual Data Manipulation
High Operational Risk
```

### Decision

Rejected.

---

# 130. Global Shared DLQ

### Advantages

```text
Fewer Topics
Simple Initial Infrastructure
```

### Disadvantages

```text
Mixed Ownership
Harder Filtering
Harder Replay
Consumer Coupling
```

### Decision

Not the default. Prefer consumer-specific DLQs when independent recovery is required.

---

# 131. Decision Matrix

| Requirement | Infinite Retry | Drop Failed Event | Shared DLQ | Consumer-Specific DLQ |
|---|---:|---:|---:|---:|
| Failure isolation | No | No | **Yes** | **Yes** |
| Event recovery | Limited | No | **Yes** | **Yes** |
| Consumer isolation | No | No | Limited | **Strong** |
| Operational clarity | Poor | Poor | Moderate | **Strong** |
| Poison message handling | Poor | Yes, by loss | **Yes** | **Yes** |
| Replay | Difficult | No | **Yes** | **Yes** |
| RideForge fit | No | No | Conditional | **Primary** |

---

# 132. Consequences

## 132.1 Positive Consequences

The decision provides:

```text
Controlled Failure Isolation
Poison Message Protection
Event Recovery
Operational Visibility
Controlled Replay
Consumer Independence
Better Incident Handling
```

---

## 132.2 Negative Consequences

The architecture introduces:

```text
Additional Topics
Retry Logic
DLQ Monitoring
Replay Tooling
Retention Management
Operational Complexity
Consumer Idempotency Requirements
```

These trade-offs are accepted.

---

# 133. Risks

## Risk 1 — DLQ Becomes a Permanent Event Graveyard

### Mitigation

Use:

```text
Ownership
Alerting
Runbooks
Retention
Replay Procedures
```

---

## Risk 2 — Infinite Replay Loop

### Mitigation

Use:

```text
Replay Count
Bounded Replay
Manual Approval
Root-Cause Validation
```

---

## Risk 3 — Sensitive Data Retention

### Mitigation

Apply:

```text
Encryption
Access Control
Data Minimization
Retention Policies
```

---

## Risk 4 — Duplicate Business Effects During Replay

### Mitigation

Use:

```text
event_id
Consumer Idempotency
External Idempotency
```

where required.

---

## Risk 5 — DLQ Storage Growth

### Mitigation

Monitor:

```text
Size
Age
Growth Rate
Retention
```

and investigate abnormal growth.

---

## Risk 6 — Ordering Violations

### Mitigation

Define ordering requirements per event stream and aggregate.

---

# 134. Validation

The DLQ strategy should be validated through:

```text
Retry Tests
Poison Message Tests
DLQ Integration Tests
Replay Tests
Duplicate Tests
Ordering Tests
Failure Tests
Load Tests
Security Tests
Operational Recovery Tests
```

---

# 135. Review Triggers

Revisit this ADR when:

```text
Event Volume Changes Significantly
Retry Architecture Changes
Redpanda Topology Changes
Consumer Architecture Changes
Multi-Region Processing Is Introduced
Replay Requirements Become More Complex
Event Retention Requirements Change
A Dedicated Workflow Engine Is Introduced
```

---

# 136. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
```

Especially:

```text
Event and Messaging Development
Error Handling and Validation
Logging and Debugging
Observability Development
Performance and Optimization
Configuration and Environment
Integration Testing and Local Infrastructure
AI Failure and Fallback Strategy
```

---

# 137. Related ADRs

This decision is directly related to:

```text
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0012 — Outbox Pattern
ADR-0014 — API and Service Communication
ADR-0015 — Smart Dispatch and Stand Dispatch
ADR-0016 — AI-Assisted Dispatch Strategy
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
ADR-0023 — Security and Secret Management
ADR-0025 — Testing and Integration Strategy
ADR-0028 — Cost Optimization Strategy
```

---

# 138. Decision Summary

RideForge adopts a:

```text
Bounded Retry
+
Dead Letter Queue
+
Controlled Replay
+
Consumer Idempotency
```

strategy for event-processing failures.

The intended lifecycle is:

```text
                    Event
                      │
                      ▼
                  Consumer
                      │
                 ┌────┴────┐
                 │         │
              Success    Failure
                 │         │
                 ▼         ▼
               Commit    Classify
                           │
                    ┌──────┴──────┐
                    │             │
                 Retryable     Permanent /
                 Failure       Exhausted
                    │             │
                    ▼             ▼
                  Retry          DLQ
                    │             │
              ┌─────┴─────┐       │
              │           │       ▼
           Success     Failure  Investigate
              │           │       │
              ▼           ▼       ▼
            Commit       DLQ    Replay
```

The DLQ is therefore a controlled exception path rather than a normal processing mechanism.

---

# 139. Final Principle

> **Retry transient failures, isolate permanent or exhausted failures in a DLQ, preserve the original event identity and failure context, and replay only after the underlying problem has been understood and corrected.**

The reliability model is:

```text
                  Producer
                     │
                  Outbox
                     │
                     ▼
                 Redpanda
                     │
                     ▼
                  Consumer
                     │
            ┌────────┴────────┐
            │                 │
         Success            Failure
            │                 │
            ▼                 ▼
         Commit            Retry
                              │
                         ┌────┴────┐
                         │         │
                      Success   Exhausted /
                                Permanent
                                   │
                                   ▼
                                  DLQ
                                   │
                              Investigation
                                   │
                              Controlled
                                Replay
```

This establishes the standard RideForge failure-isolation mechanism for asynchronous event processing while preserving recoverability, operational visibility, and consumer independence.

---

# 140. Status

```text
Decision: ACCEPTED

Failure Handling:
Bounded Retry + Dead Letter Queue

Primary Event Platform:
Redpanda

DLQ Purpose:
Failed Event Isolation and Recovery

Retry Semantics:
Bounded

Replay:
Controlled and Auditable

Consumer Requirement:
Idempotent Processing Where Business Effects Require It

Primary Goal:
Prevent Event Loss and Poison-Message Blocking
```

This decision establishes the RideForge Dead Letter Queue strategy and provides the operational foundation for reliable failure handling across the event-driven architecture.
