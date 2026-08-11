# ADR-0022: Observability Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Reliability / Operations / Platform Architecture  
> **Scope:** Logs, metrics, traces, health signals, alerts, dashboards, dependency monitoring, business observability, AI observability, and incident diagnosis  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is a distributed, event-driven ride-hailing platform.

The platform contains multiple components and execution paths:

```text
Clients
   ↓
API Gateway / Services
   ↓
Application Services
   ↓
PostgreSQL
Redis
Kafka / Redpanda
   ↓
Consumers
   ↓
Dispatch
ETA
Payments
Notifications
AI
External Providers
```

A distributed system cannot be operated reliably through application logs alone.

A single ride request may pass through:

```text
API
 ↓
Ride Service
 ↓
PostgreSQL
 ↓
Outbox
 ↓
Kafka / Redpanda
 ↓
Matching Service
 ↓
ETA
 ↓
AI Ranking
 ↓
Driver Assignment
 ↓
Notification
```

When something goes wrong, engineers must be able to determine:

```text
What happened?
Where did it happen?
When did it happen?
Which request was affected?
Which ride was affected?
Which dependency failed?
Was the system degraded?
Did a fallback execute?
Was data committed?
Was an event published?
Was the event consumed?
Was the final business effect correct?
```

Therefore RideForge requires a unified observability strategy covering:

```text
Logs
Metrics
Distributed Traces
Health
Events
Business Signals
Infrastructure
Dependencies
AI / ML
```

---

# 2. Problem

Without standardized observability, production incidents can become:

```text
Hard to Detect
Hard to Diagnose
Hard to Correlate
Hard to Reproduce
Hard to Measure
Hard to Recover
```

Examples:

```text
Ride Creation Latency Increased
```

could be caused by:

```text
API Saturation
Database Slow Query
PgBouncer Exhaustion
Redis Delay
Kafka Delay
External Provider Latency
AI Inference
```

Without correlated telemetry, the root cause becomes guesswork.

---

# 3. Decision

RideForge will use a unified observability model based on:

```text
Structured Logs
+
Metrics
+
Distributed Tracing
+
Health / Readiness Signals
+
Business Metrics
+
Dependency Metrics
+
Event / Messaging Metrics
+
AI / Model Metrics
+
Dashboards
+
Alerting
+
Correlation Identifiers
```

The fundamental principle is:

> **Every important production operation must be observable from request initiation through final business effect.**

---

# 4. Three Pillars

The core technical observability model is:

```text
Logs
Metrics
Traces
```

These should work together rather than being treated as independent systems.

---

# 5. Logs

Logs answer:

```text
What happened?
```

Examples:

```text
Ride Created
Assignment Attempted
Provider Timeout
Consumer Retry
Fallback Activated
Database Error
```

---

# 6. Metrics

Metrics answer:

```text
How often?
How much?
How fast?
How many?
```

Examples:

```text
Requests Per Second
Error Rate
P95 Latency
Consumer Lag
Database Connections
Fallback Rate
```

---

# 7. Traces

Traces answer:

```text
Where did the time go?
Which services participated?
Which dependency caused the delay?
```

---

# 8. Correlation

Logs, metrics, and traces must be correlated where practical.

Important identifiers include:

```text
request_id
trace_id
span_id
ride_id
driver_id
event_id
idempotency_key
```

These identifiers have different meanings and must not be conflated.

---

# 9. Identifier Model

```text
request_id
→ Identifies one request attempt

trace_id
→ Identifies one distributed execution trace

span_id
→ Identifies one trace operation

ride_id
→ Identifies one ride

driver_id
→ Identifies one driver

event_id
→ Identifies one event

idempotency_key
→ Identifies one logical retryable operation
```

---

# 10. Trace Context Propagation

Distributed trace context should propagate across service boundaries.

Conceptually:

```text
Client
 ↓
API
 ↓
Ride Service
 ↓
Kafka / Redpanda
 ↓
Matching Service
 ↓
Driver Service
```

where the messaging and tracing infrastructure supports appropriate propagation.

---

# 11. Request Correlation

Every incoming request should receive or propagate a request correlation identifier.

If the client supplies an acceptable request ID, the service may propagate it according to the API contract.

Otherwise the service should generate one.

---

# 12. Trace ID

A trace ID should be generated or propagated for distributed request tracing.

It should not be treated as:

```text
Authentication Credential
```

or:

```text
Business Identifier
```

---

# 13. Business Correlation

Important business identifiers should be attached to telemetry where safe.

Examples:

```text
ride_id
driver_id
region
dispatch_mode
```

Sensitive personal data should not be logged merely for convenience.

---

# 14. Structured Logging

RideForge will use structured logs rather than relying on free-form text as the primary production format.

Conceptually:

```json
{
  "level": "info",
  "service": "ride-service",
  "event": "ride.created",
  "ride_id": "ride-123",
  "trace_id": "trace-456",
  "request_id": "req-789"
}
```

The exact logging library and serialization format may evolve.

---

# 15. Log Levels

Standard levels should include:

```text
DEBUG
INFO
WARN
ERROR
```

Additional levels should only be introduced when operationally justified.

---

# 16. DEBUG

Use DEBUG for:

```text
Development Diagnostics
Detailed Execution Information
Non-Critical Internal State
```

DEBUG should generally be disabled or heavily sampled in high-volume production paths.

---

# 17. INFO

Use INFO for meaningful lifecycle events:

```text
Service Started
Ride Created
Ride Matched
Consumer Started
Configuration Loaded
Fallback Activated
```

Do not log every internal function call at INFO.

---

# 18. WARN

Use WARN when:

```text
Something Unexpected Happened
But Processing Can Continue
```

Examples:

```text
Fallback Used
Retry Attempt
Stale Location
Provider Degraded
Consumer Lag Increasing
```

---

# 19. ERROR

Use ERROR for failures requiring attention or failed operations that materially affect functionality.

Examples:

```text
Database Operation Failed
Provider Failure
Unrecoverable Consumer Error
Invalid External Response
```

Do not use ERROR for expected validation failures.

---

# 20. Avoid Log Noise

Do not log:

```text
Every Loop
Every SQL Statement
Every Cache Hit
Every Function Entry
```

unless temporarily enabled for diagnosis.

Excessive logs increase:

```text
Cost
Storage
Noise
Query Difficulty
```

---

# 21. Sensitive Data

Never log:

```text
Passwords
Access Tokens
Secret Keys
Payment Credentials
Private Keys
Sensitive Authentication Data
```

---

# 22. Personal Data

Personally identifiable information should be minimized.

Avoid logging unnecessary:

```text
Phone Numbers
Email Addresses
Full Names
Precise Personal Addresses
Government IDs
```

when an internal identifier is sufficient.

---

# 23. Location Privacy

Ride-hailing location data can be sensitive.

Do not emit highly precise location information into general-purpose logs unless there is a justified operational requirement.

Prefer:

```text
Geohash / Region
Approximate Coordinates
Internal Location ID
```

where appropriate.

---

# 24. Log Redaction

If sensitive information can enter a request or provider response, logging layers should support redaction.

---

# 25. Structured Event Names

Important logs should use stable event names.

Examples:

```text
ride.created
ride.cancelled
ride.matched
ride.completed
driver.assigned
driver.released
dispatch.fallback
provider.timeout
consumer.retry
consumer.dlq
```

Stable names make operational querying easier.

---

# 26. Error Classification

Errors should include structured categories where possible:

```text
error_code
error_type
dependency
retryable
```

---

# 27. Example

```text
error_code:
DATABASE_TIMEOUT

dependency:
postgres

retryable:
true
```

This is more useful than:

```text
something went wrong
```

---

# 28. Metrics

Metrics should represent:

```text
Availability
Latency
Traffic
Errors
Saturation
Business Health
Dependency Health
```

---

# 29. RED Metrics

For request-driven services, track:

```text
Rate
Errors
Duration
```

---

# 30. USE Metrics

For infrastructure resources, track:

```text
Utilization
Saturation
Errors
```

where applicable.

---

# 31. API Metrics

Important API metrics include:

```text
Request Count
Request Rate
Success Rate
Error Rate
Latency
Timeout Rate
Status Code Distribution
```

---

# 32. Latency Percentiles

Track:

```text
P50
P90
P95
P99
```

where useful.

Average latency alone is insufficient for production diagnosis.

---

# 33. Why P95 / P99

Averages can hide tail latency.

For example:

```text
Average = 120ms
P99 = 4.2s
```

may indicate a serious user-facing problem.

---

# 34. Endpoint Metrics

Track important endpoints individually.

Examples:

```text
POST /rides
POST /rides/{id}/cancel
POST /rides/{id}/accept
GET /rides/{id}
```

Avoid creating excessive high-cardinality endpoint labels.

---

# 35. High Cardinality

Avoid metric labels such as:

```text
ride_id
request_id
user_id
event_id
```

because they can create enormous metric cardinality.

Use them in:

```text
Logs
Traces
```

instead.

---

# 36. Recommended Metric Labels

Use bounded dimensions such as:

```text
service
endpoint
method
status_class
region
operation
provider
dispatch_mode
```

where appropriate.

---

# 37. Database Metrics

Track:

```text
Connection Pool Usage
Active Connections
Idle Connections
Query Latency
Transaction Duration
Rollback Count
Deadlock Count
Lock Wait Time
Error Rate
```

---

# 38. PgBouncer Metrics

Track where available:

```text
Client Connections
Server Connections
Pool Utilization
Waiting Clients
Transaction Pool Saturation
```

---

# 39. Long Transactions

Monitor:

```text
Transaction Duration
```

and alert when transactions exceed expected operational thresholds.

---

# 40. Redis Metrics

Track:

```text
Command Latency
Memory Usage
Connection Count
Evictions
Errors
Hit Rate
Miss Rate
CPU
Replication Health
```

according to deployment architecture.

---

# 41. Kafka / Redpanda Metrics

Track:

```text
Producer Error Rate
Producer Latency
Consumer Lag
Consumer Throughput
Consumer Error Rate
Partition Health
Retry Count
DLQ Count
```

---

# 42. Consumer Lag

Consumer lag is one of the most important messaging health signals.

High lag can indicate:

```text
Consumer Failure
Insufficient Capacity
Slow Processing
Downstream Dependency Failure
```

---

# 43. Outbox Metrics

Track:

```text
Pending Outbox Records
Publication Rate
Publication Failures
Publication Latency
Oldest Pending Record Age
```

---

# 44. Outbox Backlog

A growing outbox backlog indicates:

```text
Kafka Unavailability
Publisher Failure
Publisher Capacity Problem
Database / Query Problem
```

---

# 45. DLQ Metrics

Track:

```text
DLQ Message Count
DLQ Rate
Oldest DLQ Message Age
Consumer Error Categories
```

A DLQ should never become invisible storage for permanent failures.

---

# 46. Dispatch Metrics

Important dispatch metrics include:

```text
Matching Attempts
Match Success Rate
Time to Match
Candidate Count
Assignment Success Rate
Assignment Failure Rate
Fallback Rate
```

---

# 47. Smart Dispatch Metrics

Track:

```text
AI Ranking Invocation Count
AI Ranking Latency
AI Ranking Failure Rate
AI Fallback Rate
AI Decision Acceptance Rate
```

---

# 48. Stand Dispatch Metrics

Track:

```text
Stand Assignment Count
Stand Match Latency
Stand Queue Depth
Stand Dispatch Success Rate
```

---

# 49. Dispatch Mode Metric

Track:

```text
dispatch_mode
```

with bounded values such as:

```text
SMART
STAND
FALLBACK
```

according to the configured domain model.

---

# 50. ETA Metrics

Track:

```text
ETA Request Count
ETA Latency
Provider Failure Rate
Fallback Rate
Prediction Error
Provider Distribution
```

---

# 51. ETA Accuracy

Where actual arrival data is available, track:

```text
Predicted ETA
Actual Arrival
Absolute Error
Percentage Error
```

over appropriate reporting windows.

---

# 52. Driver Location Metrics

Track:

```text
Location Update Rate
Location Freshness
Stale Driver Count
Location Processing Latency
Location Ingestion Errors
```

---

# 53. Location Freshness

A key metric is:

```text
Current Time
-
Last Driver Location Timestamp
```

This determines whether dispatch is working with current information.

---

# 54. Payment Metrics

Track:

```text
Payment Success Rate
Payment Failure Rate
Provider Timeout Rate
Unknown Outcome Count
Refund Success Rate
Reconciliation Count
```

---

# 55. Notification Metrics

Track:

```text
Notification Requested
Notification Sent
Notification Failed
Notification Retried
Provider Latency
Delivery Status
```

---

# 56. AI / ML Metrics

AI observability must cover more than infrastructure.

Track:

```text
Inference Latency
Inference Error Rate
Fallback Rate
Model Version
Feature Availability
Prediction Distribution
Prediction Quality
Drift Indicators
```

---

# 57. Model Version

Every important model-generated decision should be traceable to:

```text
model_name
model_version
```

where applicable.

---

# 58. Feature Version

Where feature pipelines are versioned, important AI decisions should also be traceable to:

```text
feature_version
```

---

# 59. AI Decision Traceability

For important decisions:

```text
ride_id
+
dispatch_mode
+
model_version
+
feature_version
+
decision_timestamp
```

should be available through safe telemetry or decision records.

---

# 60. AI Failure

Track:

```text
Model Timeout
Model Error
Invalid Output
Missing Feature
Fallback
```

---

# 61. AI Fallback Rate

A high AI fallback rate may indicate:

```text
Model Failure
Infrastructure Failure
Feature Pipeline Failure
Latency Budget Problem
```

---

# 62. Model Quality

Business metrics should be used to evaluate AI outcomes.

Examples:

```text
Match Time
Driver Acceptance
Cancellation
ETA Accuracy
Ride Completion
```

---

# 63. AI Observability Relationship

Detailed AI monitoring is governed by:

```text
05-ai/
```

and:

```text
ADR-0026 — Model and AI Governance
```

This ADR defines the platform-wide observability principles.

---

# 64. Business Metrics

Technical health alone is insufficient.

RideForge must monitor business health.

Examples:

```text
Ride Requests
Successful Matches
Driver Acceptance
Ride Completion
Cancellation
Average Match Time
Average ETA
Payment Success
```

---

# 65. Business SLIs

Potential service-level indicators include:

```text
Ride Creation Success Rate
Ride Matching Success Rate
Time to Match
Trip Completion Rate
Payment Success Rate
Notification Success Rate
```

---

# 66. Availability

Availability should be measured for meaningful user operations rather than only:

```text
HTTP 200 Rate
```

---

# 67. Example

A service returning:

```text
HTTP 200
```

while returning invalid ride state is not necessarily healthy.

Business correctness matters.

---

# 68. Service-Level Objectives

SLOs should be defined for critical operations.

Examples:

```text
Ride Creation Availability
Ride Matching Availability
API Latency
Event Processing Delay
Payment Processing Success
```

Exact numerical targets should be determined from business requirements and operational maturity rather than invented arbitrarily.

---

# 69. Error Budgets

Once SLOs exist, error budgets may be used to balance:

```text
Reliability
+
Feature Delivery
```

---

# 70. Alerting

Alerts should identify conditions requiring action.

Avoid alerting on every error.

---

# 71. Alert Quality

Every alert should answer:

```text
What is wrong?
Why does it matter?
Who should act?
What is the likely next step?
```

---

# 72. Alert Categories

Use categories such as:

```text
Availability
Latency
Errors
Capacity
Data Integrity
Messaging
Dependency
Security
Business
AI
```

---

# 73. Critical Alerts

Examples:

```text
Ride Creation Unavailable
Database Unavailable
Critical Assignment Failure
Payment Integrity Issue
Major Regional Outage
```

---

# 74. Warning Alerts

Examples:

```text
Increasing Consumer Lag
High Fallback Rate
High Database Pool Usage
Increasing Provider Latency
```

---

# 75. Alert Fatigue

Too many alerts reduce operational effectiveness.

Every alert should have:

```text
Action
Owner
Severity
Threshold
Recovery Condition
```

---

# 76. Alert Recovery

Alerts should automatically resolve when the monitored condition returns to normal where the monitoring system supports it.

---

# 77. Dashboards

Dashboards should be organized by operational purpose.

Recommended categories:

```text
Platform Overview
API Health
Ride Operations
Dispatch
Database
Redis
Kafka / Redpanda
External Providers
AI
Business
```

---

# 78. Platform Overview Dashboard

Should answer:

```text
Is the platform healthy?
Are users successfully booking rides?
Are matches succeeding?
Are major dependencies healthy?
```

---

# 79. Ride Operations Dashboard

Include:

```text
Ride Requests
Ride Creation Success
Active Rides
Matching Success
Cancellation
Completion
```

---

# 80. Dispatch Dashboard

Include:

```text
Match Rate
Time to Match
Candidate Count
Smart Dispatch
Stand Dispatch
Fallback Rate
Driver Availability
```

---

# 81. Database Dashboard

Include:

```text
Connections
Pool Usage
Query Latency
Transaction Duration
Locks
Deadlocks
Errors
```

---

# 82. Messaging Dashboard

Include:

```text
Producer Errors
Consumer Lag
Throughput
Retry Rate
DLQ
Outbox Backlog
```

---

# 83. Dependency Dashboard

Include:

```text
Provider Latency
Provider Errors
Timeouts
Circuit State
Fallback Rate
```

---

# 84. AI Dashboard

Include:

```text
Inference Requests
Latency
Errors
Fallback
Model Version
Feature Availability
Prediction Quality
```

---

# 85. Business Dashboard

Include:

```text
Ride Requests
Matches
Completion
Cancellation
Revenue / Payment Success
Driver Activity
```

---

# 86. Tracing

Distributed tracing should cover critical request paths.

Example:

```text
POST /rides
   │
   ├── Ride Service
   │      └── PostgreSQL
   │
   └── Outbox
          └── Kafka
                 └── Matching Service
                        ├── Location
                        ├── ETA
                        └── AI
```

---

# 87. Trace Boundaries

Create spans around meaningful operations.

Examples:

```text
ride.create
db.transaction
db.query
outbox.publish
dispatch.match
eta.calculate
ai.rank
driver.assign
notification.send
```

---

# 88. Do Not Trace Every Function

Tracing every internal function can produce:

```text
High Overhead
High Storage
High Noise
```

Trace meaningful boundaries.

---

# 89. Database Tracing

Database operations should be traceable enough to identify:

```text
Slow Query
Transaction Delay
Connection Wait
```

without exposing sensitive query parameters.

---

# 90. External Provider Tracing

Provider calls should include spans showing:

```text
Provider
Operation
Latency
Result Category
```

Do not store secrets or sensitive payloads.

---

# 91. Messaging Tracing

Where supported, trace:

```text
Producer
→ Topic
→ Consumer
```

with propagated trace context.

---

# 92. Asynchronous Trace Boundaries

An asynchronous event does not always represent a direct continuation of the original HTTP request.

Telemetry should preserve both:

```text
Causal Relationship
```

and:

```text
New Processing Execution
```

where appropriate.

---

# 93. Trace Sampling

Tracing may use sampling to control cost.

Critical or anomalous requests may receive higher sampling priority.

---

# 94. Error Sampling

Errors should generally be retained at a higher sampling rate than ordinary successful traffic.

---

# 95. High-Value Trace Retention

Longer retention may be justified for:

```text
Failed Rides
Payment Failures
Dispatch Failures
Major Incidents
```

subject to privacy and cost requirements.

---

# 96. Log Retention

Retention should be based on:

```text
Operational Need
Compliance
Privacy
Cost
Incident Investigation
```

---

# 97. Metric Retention

Metrics generally require longer aggregation windows than raw logs.

Use:

```text
High Resolution Short Term
Aggregated Long Term
```

where appropriate.

---

# 98. Trace Retention

Trace retention should balance:

```text
Debugging Value
Storage Cost
Privacy
```

---

# 99. Data Privacy

Observability data can itself become sensitive data.

Therefore:

```text
Logs
Metrics
Traces
```

must follow the same security and privacy principles as application data.

---

# 100. Access Control

Observability systems should have role-based access.

Not every engineer needs unrestricted access to:

```text
Production Logs
Payment Data
Location Data
Personal Information
```

---

# 101. Auditability

Access to highly sensitive observability data should be auditable where required.

---

# 102. Production vs Development

Production telemetry should be:

```text
Structured
Sanitized
Controlled
Queryable
```

Development telemetry may be more verbose.

---

# 103. Local Development

Local development should provide enough observability to reproduce production behaviour without requiring the production observability stack.

---

# 104. Local Correlation

Local logs should still expose:

```text
request_id
trace_id
service
operation
```

where practical.

---

# 105. Error Logs

An error should contain enough context to diagnose it.

Recommended fields:

```text
timestamp
level
service
environment
operation
error_code
error_type
request_id
trace_id
resource_id
dependency
retryable
```

---

# 106. Do Not Log Stack Traces Everywhere

Stack traces should be emitted when they add diagnostic value.

Avoid duplicating the same stack trace across multiple layers.

---

# 107. Error Ownership

The layer that handles an error should generally log it at the appropriate level.

Avoid:

```text
Repository ERROR
Service ERROR
HTTP ERROR
```

all logging the exact same failure as independent errors.

This creates noise.

---

# 108. Error Propagation

Errors should carry structured context through application layers.

The final log should contain enough information without requiring every layer to emit duplicate logs.

---

# 109. Metrics vs Logs

Do not use logs for high-volume counting when a metric is more appropriate.

For example:

```text
Every Successful Ride
```

may generate an event/log, while:

```text
Ride Success Rate
```

should be a metric.

---

# 110. Metrics vs Traces

Use metrics for:

```text
Aggregated Health
```

Use traces for:

```text
Individual Execution Paths
```

---

# 111. Logs vs Traces

Use logs for:

```text
Detailed Event Context
```

Use traces for:

```text
Distributed Timing and Causality
```

---

# 112. Observability Correlation Example

For a failed ride:

```text
ride_id = R123
```

the engineer should be able to locate:

```text
API Trace
Ride Service Logs
Database Span
Outbox Event
Kafka Event
Matching Trace
ETA Trace
AI Decision
Assignment Result
```

where those components participated.

---

# 113. Incident Investigation

A standard investigation path should be:

```text
Business Symptom
 ↓
Metric
 ↓
Affected Service
 ↓
Trace
 ↓
Logs
 ↓
Dependency
 ↓
Root Cause
```

---

# 114. Incident Example

Symptom:

```text
Time to Match Increased
```

Investigate:

```text
Dispatch Metric
 ↓
Matching Latency
 ↓
Trace
 ↓
ETA Span
 ↓
Provider Latency
 ↓
Provider Failure
```

---

# 115. Operational Runbooks

Important alerts should link to runbooks.

Runbooks should define:

```text
Symptoms
Checks
Likely Causes
Immediate Actions
Recovery
Escalation
```

---

# 116. Dependency Runbook

For each critical dependency, maintain:

```text
Health Check
Failure Symptoms
Fallback
Recovery
Escalation
```

---

# 117. Database Runbook

Should cover:

```text
Connection Exhaustion
Slow Queries
Lock Contention
Deadlocks
Replication Problems
Backup / Recovery
```

---

# 118. Kafka / Redpanda Runbook

Should cover:

```text
Producer Failure
Consumer Lag
Partition Problems
Outbox Backlog
DLQ Growth
Consumer Restart
```

---

# 119. Redis Runbook

Should cover:

```text
Memory Pressure
High Latency
Connection Problems
Evictions
Failover
Data Rebuild
```

---

# 120. AI Runbook

Should cover:

```text
Model Failure
High Latency
Fallback Spike
Feature Failure
Model Version Problem
Prediction Drift
```

---

# 121. Regional Observability

Metrics should support regional analysis.

Useful bounded dimensions:

```text
region
country
operating_zone
dispatch_mode
```

Avoid high-cardinality raw geographic identifiers in metrics.

---

# 122. Legal / Regional Failures

Track failures related to:

```text
Regional Validation
Ride Eligibility
Cross-Region Restrictions
```

This is especially important because a legal validation failure is not equivalent to an infrastructure failure.

---

# 123. Dispatch Observability by Mode

The system should make it possible to compare:

```text
Smart Dispatch
Stand Dispatch
Fallback Dispatch
```

without mixing them into one opaque metric.

---

# 124. Cost Observability

Observability itself has a cost.

Track:

```text
Log Volume
Metric Cardinality
Trace Volume
Storage
Ingestion Cost
Retention Cost
```

---

# 125. Cost Control

Use:

```text
Sampling
Aggregation
Retention Policies
Log Levels
Cardinality Control
```

rather than disabling useful observability entirely.

---

# 126. High-Cardinality Protection

Never add:

```text
ride_id
user_id
driver_id
request_id
event_id
```

as unbounded metric labels.

Use them in logs and traces.

---

# 127. Metric Naming

Metrics should follow a consistent naming convention.

Examples:

```text
http_requests_total
http_request_duration_seconds
ride_matches_total
ride_match_duration_seconds
kafka_consumer_lag
outbox_pending_total
```

The exact naming convention should be standardized in development documentation.

---

# 128. Units

Metric names should make units clear where the telemetry system does not enforce them.

Examples:

```text
_seconds
_bytes
_total
_ratio
```

---

# 129. Counter Metrics

Use counters for monotonically increasing events:

```text
requests_total
errors_total
rides_created_total
```

---

# 130. Gauge Metrics

Use gauges for current state:

```text
active_rides
active_drivers
queue_depth
connections
```

---

# 131. Histogram Metrics

Use histograms for distributions:

```text
request_duration
db_query_duration
eta_error
ai_inference_duration
```

---

# 132. SLO Metrics

SLO-related metrics should be derived from stable, well-defined indicators.

---

# 133. Health Signals

Every service should expose appropriate:

```text
Liveness
Readiness
Dependency Health
```

signals.

---

# 134. Liveness Principle

Liveness should answer:

```text
Can this process continue running?
```

It should not be used to force restarts for every dependency outage.

---

# 135. Readiness Principle

Readiness should answer:

```text
Can this instance safely accept the relevant workload?
```

---

# 136. Dependency Health

Dependency health should be visible without necessarily making the entire service unhealthy.

---

# 137. Startup Observability

Service startup should emit:

```text
Service Name
Version
Environment
Configuration Version
Startup Result
Dependency Initialization Result
```

without logging secrets.

---

# 138. Version Observability

Every service should expose its deployed version/build identifier.

This enables:

```text
Incident
→ Version
→ Deployment
```

correlation.

---

# 139. Deployment Correlation

Observability should allow engineers to determine:

```text
Did this problem start after a deployment?
```

---

# 140. Feature Flags

If feature flags exist, important traces/logs should expose the relevant bounded feature state when it materially affects behaviour.

---

# 141. Dispatch Configuration

For dispatch decisions, telemetry should identify:

```text
Region
Dispatch Mode
Algorithm Version
Fallback
```

where appropriate.

---

# 142. AI Decision Auditability

For important AI-assisted decisions, the system should be able to answer:

```text
Which model?
Which version?
Which features?
Which decision?
When?
Why did fallback occur?
```

without exposing sensitive model internals unnecessarily.

---

# 143. Model Observability

Detailed model monitoring remains governed by:

```text
05-ai/17.AI_MONITORING_AND_MODEL_OBSERVABILITY.md
```

and:

```text
ADR-0026 — Model and AI Governance
```

---

# 144. Security Monitoring

Observability should also detect:

```text
Authentication Failures
Authorization Failures
Suspicious Request Volume
Secret Access Problems
Unusual Provider Behaviour
```

Security-specific telemetry must not expose secrets.

---

# 145. Security Events

Security events should be distinguishable from normal application events.

---

# 146. Audit Logs

Business actions requiring auditability should use durable audit records where required.

Do not assume ordinary application logs are a substitute for an authoritative audit trail.

---

# 147. Audit vs Application Logs

```text
Application Log
→ Operational Diagnosis

Audit Record
→ Durable Record of Significant Action
```

They serve different purposes.

---

# 148. Data Integrity Monitoring

Monitor signals such as:

```text
Duplicate Assignment
Invalid State Transition
Orphaned Record
Unexpected Status
Reconciliation Failure
```

---

# 149. Business Invariant Alerts

A business invariant violation may be more serious than a technical error.

Example:

```text
One Driver
→ Two Active Rides
```

should trigger immediate investigation.

---

# 150. Data Freshness

Monitor freshness for:

```text
Driver Location
Read Models
AI Features
ETA Data
Event Processing
```

---

# 151. Freshness Metric

A useful general measure is:

```text
Current Time
-
Last Successful Update
```

---

# 152. Event Freshness

For event-driven systems, monitor:

```text
Event Age
Consumer Processing Delay
Oldest Unprocessed Event
```

---

# 153. Outbox Freshness

Monitor:

```text
Age of Oldest Pending Outbox Record
```

This can detect event publication degradation earlier than simple queue counts.

---

# 154. Dependency Latency

Track dependency latency separately from overall request latency.

This allows:

```text
API Slow
→ Which Dependency?
```

to be answered quickly.

---

# 155. Provider Error Categories

External provider errors should be classified into bounded categories:

```text
TIMEOUT
UNAVAILABLE
RATE_LIMITED
INVALID_RESPONSE
AUTHENTICATION_ERROR
BUSINESS_ERROR
UNKNOWN
```

---

# 156. Rate Limits

Provider rate-limit events should be observable separately from general provider failures.

---

# 157. Circuit State Metrics

Where circuit breakers exist, track:

```text
OPEN
HALF_OPEN
CLOSED
```

and transition counts.

---

# 158. Fallback Metrics

Fallback events should identify:

```text
fallback_type
reason
region
service
```

using bounded labels where metrics are involved.

---

# 159. Recovery Metrics

Track:

```text
Time to Detect
Time to Recover
Failure Duration
Fallback Duration
```

---

# 160. MTTR

Mean Time to Recovery can be used as an operational indicator.

It should not be the only reliability metric.

---

# 161. MTTD

Mean Time to Detect measures how quickly incidents become visible.

Observability should reduce:

```text
MTTD
```

---

# 162. Incident Correlation

Major incidents should preserve:

```text
Incident ID
Trace IDs
Service Versions
Deployment IDs
Relevant Logs
Relevant Metrics
```

where practical.

---

# 163. Retrospective Data

After an incident, observability should allow engineers to determine:

```text
What happened?
Why wasn't it detected earlier?
Which signal should have alerted us?
What telemetry was missing?
```

---

# 164. Observability Gap

If an incident cannot be diagnosed because telemetry is missing, record an:

```text
Observability Gap
```

and address it as engineering work.

---

# 165. Testing Observability

Observability itself must be tested.

Verify:

```text
Logs Are Emitted
Metrics Update
Trace Context Propagates
Alerts Trigger
Fallbacks Are Visible
Health Checks Behave Correctly
```

---

# 166. Failure Injection Observability

When injecting failures, verify that telemetry identifies:

```text
Failure
Impact
Fallback
Recovery
```

---

# 167. Logging Tests

Test structured log fields for important events.

---

# 168. Metric Tests

Verify critical metrics are emitted and updated correctly.

Avoid coupling application unit tests excessively to implementation-specific metric names unless those names are part of the operational contract.

---

# 169. Trace Tests

Integration tests should verify trace propagation across critical service boundaries where practical.

---

# 170. Alert Tests

Critical alerts should be tested periodically to prevent stale monitoring rules.

---

# 171. Dashboard Validation

Dashboards should be reviewed after major architecture changes.

A dashboard that references removed services or metrics creates false confidence.

---

# 172. Observability Ownership

Every important metric/log/alert should have an owner.

Ownership may be:

```text
Service Team
Platform Team
Operations
Security
Data / AI Team
```

---

# 173. Observability Standards

Development standards should define:

```text
Log Format
Metric Naming
Trace Naming
Error Codes
Correlation IDs
Sensitive Data Rules
```

---

# 174. Configuration

Observability configuration should follow:

```text
ADR-0024 — Configuration and Environment Strategy
```

Avoid hard-coding:

```text
Telemetry Endpoints
Credentials
Sampling Policies
Retention Settings
```

---

# 175. Secret Management

Telemetry credentials and provider credentials must follow:

```text
ADR-0023 — Security and Secret Management
```

---

# 176. Performance

Telemetry must not materially degrade critical application performance.

Use:

```text
Asynchronous Logging
Sampling
Batch Export
Efficient Serialization
Bounded Buffers
```

where appropriate.

---

# 177. Telemetry Backpressure

Telemetry systems can also fail.

If telemetry exporters become unavailable:

```text
Application Core Functionality
```

must not normally fail solely because telemetry cannot be delivered.

---

# 178. Telemetry Failure

Telemetry failure should be:

```text
Contained
Observable Locally
Recoverable
```

without creating a cascading failure.

---

# 179. Local Buffering

Where appropriate, telemetry exporters may buffer data temporarily.

Buffers must be bounded.

---

# 180. Never Block Critical Business Operations Indefinitely

A failed telemetry backend must not hold:

```text
Ride Creation
Driver Assignment
Payment State
```

indefinitely.

---

# 181. Cost Optimization

Observability must align with:

```text
ADR-0028 — Cost Optimization Strategy
```

Use cost controls without sacrificing critical signals.

---

# 182. Observability Priority

If telemetry volume must be reduced, preserve in roughly this order:

```text
Critical Errors
Security Events
Business Integrity Signals
Critical Traces
Operational Metrics
Important Logs
Debug Logs
```

The exact priority may vary by incident.

---

# 183. Observability Architecture

Conceptually:

```text
                 ┌─────────────────┐
                 │   Applications  │
                 └────────┬────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
       Logs            Metrics          Traces
          │               │               │
          └───────────────┼───────────────┘
                          ▼
                 Telemetry Pipeline
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
       Storage         Monitoring       Alerting
          │               │               │
          └───────────────┼───────────────┘
                          ▼
                    Dashboards
                          │
                          ▼
                    Engineering /
                     Operations
```

---

# 184. End-to-End Ride Observability

A successful ride should be traceable approximately as:

```text
Ride Request
     │
     ▼
API Request
     │
     ▼
Ride Creation
     │
     ├── PostgreSQL
     │
     └── Outbox
            │
            ▼
        Kafka / Redpanda
            │
            ▼
      Matching Service
            │
      ┌─────┼─────┐
      ▼     ▼     ▼
   Location ETA    AI
      │     │      │
      └─────┼──────┘
            ▼
      Driver Assignment
            │
            ▼
       Notification
```

Telemetry should make the critical path visible.

---

# 185. End-to-End Failure Example

Suppose:

```text
ETA Provider
```

becomes slow.

Observability should show:

```text
Provider Latency ↑
       ↓
ETA Latency ↑
       ↓
Dispatch Latency ↑
       ↓
Time to Match ↑
```

and ideally:

```text
Fallback Rate ↑
```

if the fallback activates.

---

# 186. Alert Example

A useful alert is:

```text
Time to Match SLO Breached
```

rather than:

```text
Some Service Error Increased
```

because the former directly represents business impact.

---

# 187. Dependency-to-Business Correlation

The observability platform should allow operators to correlate:

```text
Technical Failure
```

with:

```text
Business Impact
```

Example:

```text
Route Provider Failure
→ ETA Fallback
→ Match Latency
→ Ride Completion Impact
```

---

# 188. Regional Analysis

For regional operations, dashboards should support:

```text
Region
Dispatch Mode
Provider
```

comparison.

---

# 189. Avoid Unbounded Geographic Labels

Do not use exact:

```text
Latitude
Longitude
Address
```

as metric labels.

---

# 190. Privacy-Aware Observability

Observability design must minimize exposure of:

```text
User Identity
Driver Identity
Location
Payment Data
Personal Communication
```

while retaining enough information for operations.

---

# 191. Observability and Compliance

If regulatory or legal requirements require auditability, the observability architecture must preserve appropriate records without treating ordinary logs as a compliance substitute.

---

# 192. Incident Severity

Telemetry should support rapid identification of:

```text
Platform-Wide Incident
Regional Incident
Service Incident
Dependency Incident
Data Integrity Incident
```

---

# 193. Regional Incident

A regional outage should be identifiable without being hidden inside global averages.

---

# 194. Service Incident

Service-specific metrics should allow isolation of:

```text
Ride Service
Matching Service
Driver Service
Payment Service
AI Service
```

---

# 195. Dependency Incident

Provider-specific telemetry should allow isolation of:

```text
Map Provider
Payment Provider
Notification Provider
AI Provider
```

---

# 196. Data Integrity Incident

Monitor for:

```text
Invalid State
Duplicate Assignment
Missing Event
Stuck Outbox
Stuck Processing
Reconciliation Failure
```

---

# 197. Observability and Architecture Evolution

When a service is split or merged, telemetry ownership and dashboards must evolve with it.

---

# 198. Migration Requirement

Architecture migrations should preserve enough telemetry to compare:

```text
Before
vs
After
```

---

# 199. Deployment Comparison

Important metrics should be compared around deployments:

```text
Latency
Errors
Throughput
Fallback
Business Success
```

---

# 200. Canary Observability

If canary deployment is used, telemetry should distinguish:

```text
Canary
vs
Stable
```

using bounded dimensions.

---

# 201. Rollback Signal

A deployment should be eligible for rollback when it causes meaningful degradation in:

```text
Error Rate
Latency
Business Success
Data Integrity
```

according to deployment policy.

---

# 202. Operational Principle

Observability is not merely:

```text
Logging
```

It is the ability to understand:

```text
System State
Business State
Failure State
Dependency State
Recovery State
```

---

# 203. Consequences

## 203.1 Positive Consequences

The strategy provides:

```text
Faster Incident Detection
Faster Diagnosis
Better Root-Cause Analysis
Better Business Visibility
Safer AI Operations
Improved Reliability
Better Capacity Planning
Better Architecture Decisions
```

---

## 203.2 Negative Consequences

The architecture introduces:

```text
Telemetry Infrastructure
Storage Cost
Operational Complexity
Instrumentation Work
Privacy Considerations
Cardinality Management
Dashboard Maintenance
Alert Maintenance
```

These trade-offs are accepted.

---

# 204. Risks

## Risk 1 — Excessive Telemetry Cost

### Mitigation

```text
Sampling
Aggregation
Retention Policies
Cardinality Control
```

---

## Risk 2 — High Cardinality

### Mitigation

Keep high-cardinality identifiers in:

```text
Logs
Traces
```

rather than metrics.

---

## Risk 3 — Sensitive Data Leakage

### Mitigation

```text
Redaction
Access Control
Data Minimization
Review
```

---

## Risk 4 — Alert Fatigue

### Mitigation

```text
Actionable Alerts
Severity
Ownership
SLO-Based Alerting
```

---

## Risk 5 — Telemetry Failure Affects Application

### Mitigation

```text
Bounded Buffers
Asynchronous Export
Fail-Open Telemetry
```

where appropriate.

---

## Risk 6 — False Confidence

A green technical dashboard does not guarantee business correctness.

### Mitigation

Track:

```text
Business Metrics
Data Integrity Signals
```

alongside infrastructure metrics.

---

# 205. Alternatives Considered

## 205.1 Logs Only

### Advantages

```text
Simple
```

### Disadvantages

```text
Poor Aggregation
Poor Distributed Timing
Difficult Capacity Analysis
```

### Decision

```text
Rejected.
```

---

# 206. Metrics Only

### Advantages

```text
Low Storage
Good Aggregation
```

### Disadvantages

```text
Insufficient Diagnostic Detail
Poor Individual Request Investigation
```

### Decision

```text
Rejected.
```

---

# 207. Traces Only

### Advantages

```text
Excellent Request Visibility
```

### Disadvantages

```text
Expensive
Not Ideal for Long-Term Aggregated Business Metrics
```

### Decision

```text
Rejected.
```

---

# 208. No Distributed Tracing

### Advantages

```text
Lower Complexity
```

### Disadvantages

```text
Harder Microservice Diagnosis
Harder Dependency Analysis
```

### Decision

```text
Rejected.
```

---

# 209. Log Everything

### Advantages

```text
Maximum Raw Detail
```

### Disadvantages

```text
Cost
Noise
Privacy Risk
Performance Impact
```

### Decision

```text
Rejected.
```

---

# 210. Validation

This ADR should be validated through:

```text
Integration Tests
Failure Injection
Trace Propagation Tests
Logging Tests
Metric Tests
Alert Tests
Dashboard Reviews
Load Tests
Chaos Tests
Incident Drills
```

---

# 211. Review Triggers

Revisit this ADR when:

```text
A New Service Is Added
A New Critical Dependency Is Added
Observability Cost Becomes Excessive
A Major Incident Reveals Missing Telemetry
Metric Cardinality Becomes a Problem
Tracing Architecture Changes
AI Architecture Changes
Multi-Region Deployment Is Introduced
Compliance Requirements Change
```

---

# 212. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
adr/
```

Especially:

```text
Logging and Debugging
Observability Development
Performance and Optimization
Error Handling and Validation
Event and Messaging Development
Testing and Integration Testing
Configuration and Environment
```

---

# 213. Related ADRs

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
ADR-0021 — Failure and Degradation Strategy
ADR-0023 — Security and Secret Management
ADR-0024 — Configuration and Environment Strategy
ADR-0025 — Testing and Integration Strategy
ADR-0026 — Model and AI Governance
ADR-0027 — Cloud and Deployment Strategy
ADR-0028 — Cost Optimization Strategy
```

---

# 214. Decision Summary

RideForge adopts:

```text
                     PRODUCTION SYSTEM
                            │
          ┌─────────────────┼─────────────────┐
          ▼                 ▼                 ▼
        LOGS             METRICS            TRACES
          │                 │                 │
          └─────────────────┼─────────────────┘
                            ▼
                    CORRELATED TELEMETRY
                            │
          ┌─────────────────┼─────────────────┐
          ▼                 ▼                 ▼
       TECHNICAL         BUSINESS          DEPENDENCY
       HEALTH            HEALTH             HEALTH
          │                 │                 │
          └─────────────────┼─────────────────┘
                            ▼
                     ALERTS / DASHBOARDS
                            │
                            ▼
                     INCIDENT RESPONSE
                            │
                            ▼
                     ROOT-CAUSE ANALYSIS
                            │
                            ▼
                       IMPROVEMENT
```

The platform will make important operations observable across:

```text
API
Database
Redis
Kafka / Redpanda
Outbox
Consumers
Dispatch
Location
ETA
AI
Payments
Notifications
External Providers
Business Workflows
```

---

# 215. Final Principles

The following principles are mandatory:

```text
1. Every critical production operation must be observable.

2. Logs, metrics, and traces must complement one another.

3. Structured logs are the production logging standard.

4. High-cardinality identifiers belong primarily in logs and traces, not metrics.

5. request_id, trace_id, event_id, ride_id, driver_id, and idempotency_key have distinct meanings.

6. Distributed trace context should propagate across service boundaries.

7. Business metrics must exist alongside technical metrics.

8. Technical health does not guarantee business correctness.

9. Critical dependencies must have measurable health signals.

10. Outbox backlog and consumer lag must be observable.

11. Dispatch latency and match success must be observable.

12. Smart Dispatch, Stand Dispatch, and fallback behaviour must be distinguishable.

13. ETA provider failures and fallback usage must be observable.

14. AI model version and fallback behaviour must be traceable where applicable.

15. Payment failures and reconciliation states must be observable.

16. Location freshness must be observable.

17. Important data-integrity violations must generate operational signals.

18. Alerts must be actionable and owned.

19. Alert fatigue must be actively controlled.

20. Observability data must follow privacy and security requirements.

21. Secrets must never be emitted into telemetry.

22. Telemetry failure must not normally take down critical business operations.

23. Observability systems must have bounded resource usage.

24. Sampling and retention must control cost without removing critical signals.

25. Every major incident should identify observability gaps.

26. Observability must evolve with architecture changes.

27. Deployment changes must be correlatable with system health.

28. Regional failures must be distinguishable from global failures.

29. Degraded modes must be visible.

30. Recovery must be observable.

31. Critical business workflows should be traceable end-to-end.

32. The goal of observability is not maximum telemetry; it is maximum useful understanding of system and business state.
```

---

# 216. Status

```text
Decision: ACCEPTED

Observability Model:
Logs + Metrics + Traces

Logging:
Structured

Metrics:
Technical + Business

Tracing:
Distributed

Correlation:
request_id + trace_id + business identifiers

API Observability:
Required

Database Observability:
Required

Redis Observability:
Required

Kafka / Redpanda Observability:
Required

Outbox Observability:
Required

DLQ Observability:
Required

Dispatch Observability:
Required

ETA Observability:
Required

Location Freshness:
Required

AI Observability:
Required

Payment Observability:
Required

Notification Observability:
Required

Health:
Liveness + Readiness + Dependency Signals

Alerting:
Actionable + Owned

Dashboards:
Operational + Business + Dependency

Privacy:
Required

Security:
Required

High Cardinality:
Controlled

Telemetry Cost:
Controlled Through Sampling / Retention / Aggregation

Incident Investigation:
Traceable End-to-End

Primary Goal:
Provide Reliable, Correlated, Privacy-Aware Visibility Into Technical Health, Business Health, Failures, and Recovery
```
