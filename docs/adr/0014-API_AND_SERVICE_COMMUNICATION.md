# ADR-0014: API and Service Communication

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Service Communication / Integration Architecture  
> **Scope:** Communication between RideForge services and external-facing clients  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is designed as a modular, service-oriented and event-driven ride-hailing platform.

The architecture contains multiple logical services and infrastructure components, including:

```text
API / Edge Layer
User Services
Ride Services
Driver Services
Matching Services
Dispatch Services
Payment Services
Notification Services
AI Services
ETA Services
PostgreSQL
Redis
Redpanda
External Route Providers
External Payment Providers
```

These components do not all require the same communication mechanism.

Some operations require:

```text
Immediate Response
Strong Request/Response Semantics
Synchronous Validation
```

while others are better represented as:

```text
Asynchronous Events
Eventual Consistency
Background Processing
Decoupled Consumers
```

RideForge therefore requires a clear communication strategy that prevents:

```text
Unnecessary Service Coupling
Long Synchronous Chains
Distributed Failure Propagation
Unclear Ownership
Inconsistent API Contracts
```

---

# 2. Problem

The platform needs to determine:

```text
When to use synchronous API communication
When to use asynchronous event communication
How services discover each other
How APIs are secured
How requests are traced
How timeouts and retries work
How failures propagate
How service boundaries remain independent
```

The architecture must remain practical for the initial deployment while supporting future service scaling.

---

# 3. Decision

RideForge will use a **hybrid service communication model**:

```text
Synchronous HTTP APIs
+
Asynchronous Redpanda Events
```

The primary rule is:

> **Use synchronous communication when the caller requires an immediate result; use asynchronous events when the operation can be processed independently after the initiating transaction.**

The conceptual architecture is:

```text
                    Clients
                       │
                       ▼
                API / Edge Layer
                       │
              Synchronous HTTP
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
     Ride API      User API       Driver API
        │              │              │
        └──────────────┼──────────────┘
                       │
                 Domain Services
                       │
                  PostgreSQL
                       │
                    Redis

Domain Transactions
        │
        ▼
Transactional Outbox
        │
        ▼
     Redpanda
        │
        ├──────────────┬──────────────┐
        ▼              ▼              ▼
   Matching       Notification       AI
   Consumer         Consumer       Consumer
```

---

# 4. Communication Principles

RideForge communication must follow these principles:

```text
1. Prefer the simplest mechanism that satisfies the requirement.
2. Do not use asynchronous messaging when an immediate response is required.
3. Do not create synchronous chains for work that can be asynchronous.
4. Services communicate through explicit contracts.
5. Service ownership must remain clear.
6. Timeouts are mandatory for synchronous remote calls.
7. Retries must be bounded.
8. Events must be idempotently consumable.
9. Internal APIs must not expose database implementation details.
10. Observability must cross service boundaries.
```

---

# 5. Communication Categories

RideForge communication is divided into:

```text
Client → Platform
Service → Service
Service → Event Bus
Service → External Provider
Service → Infrastructure
```

Each category has different requirements.

---

# 6. Client-to-Service Communication

External clients such as:

```text
Rider Mobile App
Driver Mobile App
Admin Dashboard
Partner Applications
```

should communicate through the platform's API boundary.

The preferred model is:

```text
Client
  ↓
API / Edge
  ↓
Application Service
```

Clients should not communicate directly with internal databases or internal infrastructure.

---

# 7. Service-to-Service Communication

Service-to-service communication may use:

```text
Synchronous HTTP
```

or:

```text
Asynchronous Redpanda Events
```

depending on the business requirement.

---

# 8. Synchronous Communication

Synchronous communication is appropriate when the caller needs a result before continuing.

Examples:

```text
Validate User
Fetch Ride Details
Check Driver Eligibility
Calculate Immediate Quote
Request Current ETA
Retrieve Configuration
```

Conceptually:

```text
Service A
   │
   │ HTTP Request
   ▼
Service B
   │
   │ Response
   ▼
Service A
```

---

# 9. Asynchronous Communication

Asynchronous events are appropriate when the caller does not need an immediate response.

Examples:

```text
RideCreated
DriverAssigned
RideStarted
RideCompleted
PaymentCompleted
NotificationRequested
AnalyticsEvent
AI Feature Update
```

Conceptually:

```text
Service A
   │
   ▼
Outbox
   │
   ▼
Redpanda
   │
   ├── Service B
   ├── Service C
   └── Service D
```

---

# 10. Decision Rule

Use this question:

> **Does the caller need the result to complete the current operation?**

If:

```text
Yes
```

prefer:

```text
Synchronous API
```

If:

```text
No
```

consider:

```text
Asynchronous Event
```

---

# 11. Synchronous Communication Should Be Narrow

A synchronous call should return the information necessary for the current operation.

Avoid turning one request into:

```text
A
 ↓
B
 ↓
C
 ↓
D
 ↓
E
```

Long chains increase:

```text
Latency
Failure Probability
Debugging Complexity
Dependency Coupling
```

---

# 12. Synchronous Call Chain

Prefer:

```text
Client
  ↓
Ride Service
```

over:

```text
Client
  ↓
Ride Service
  ↓
Driver Service
  ↓
Pricing Service
  ↓
Location Service
  ↓
Notification Service
```

when the downstream operations can be performed asynchronously.

---

# 13. Communication Ownership

The service owning a business capability owns the corresponding API contract.

For example:

```text
Ride Service
→ Ride API

Driver Service
→ Driver API

Payment Service
→ Payment API
```

Other services should consume these contracts rather than directly accessing another service's database.

---

# 14. No Cross-Service Database Access

RideForge services must not directly access another service's private database tables.

Avoid:

```text
Service A
    ↓
Service B Database
```

Prefer:

```text
Service A
    ↓
Service B API
```

or:

```text
Service A
    ↓
Event
    ↓
Service B
```

---

# 15. Database Ownership

Each service owns its persistence model.

The database boundary should reinforce the service boundary.

Conceptually:

```text
Ride Service
    ↓
Ride Data

Driver Service
    ↓
Driver Data
```

Shared database infrastructure does not imply shared ownership of tables.

---

# 16. HTTP as Initial Synchronous Protocol

RideForge will use HTTP-based APIs for synchronous service communication.

This provides:

```text
Simple Tooling
Broad Language Support
Easy Debugging
Well-Defined Semantics
Good Operational Support
```

---

# 17. REST-Oriented APIs

Internal HTTP APIs should generally follow resource-oriented design.

Examples:

```text
GET /rides/{ride_id}
POST /rides
POST /rides/{ride_id}/cancel
GET /drivers/{driver_id}
```

The exact endpoint contract is governed by API development standards.

---

# 18. API Contract

Every internal API should define:

```text
Endpoint
Method
Request Schema
Response Schema
Status Codes
Authentication
Authorization
Timeout Expectations
Error Schema
Version
```

---

# 19. Internal API Contracts

Internal APIs should be treated as real contracts.

They must not rely on:

```text
Undocumented Fields
Database Columns
Implicit Behaviour
Service Implementation Details
```

---

# 20. API Versioning

Breaking API changes require explicit versioning.

A common approach is:

```text
/api/v1/...
```

Future breaking changes may use:

```text
/api/v2/...
```

Versioning should be introduced deliberately rather than for every minor change.

---

# 21. Backward Compatibility

Prefer backward-compatible changes such as:

```text
Adding Optional Fields
Adding New Endpoints
Adding New Event Fields
```

Avoid breaking existing consumers without a migration plan.

---

# 22. Request Timeout

Every synchronous service call must have a timeout.

Conceptually:

```text
Service A
    │
    ├── Request
    │
    └── Deadline
           ↓
        Service B
```

A remote call must not wait indefinitely.

---

# 23. Timeout Propagation

The caller's remaining deadline should be propagated where practical.

For example:

```text
Incoming Request Deadline
        ↓
Service A
        ↓
Service B
        ↓
Database
```

Each downstream operation should respect the remaining budget.

---

# 24. Latency Budget

Synchronous workflows should have explicit latency budgets.

Conceptually:

```text
Total Request Budget
        │
        ├── Service A
        ├── Service B
        └── Database
```

A downstream service must not consume the entire caller's latency budget unnecessarily.

---

# 25. Retry Policy

Not every failed HTTP request should be retried.

Retry only when:

```text
Failure Is Likely Transient
+
Operation Is Safe to Retry
```

Examples of potentially retryable failures:

```text
Timeout
Temporary Network Failure
503 Service Unavailable
```

---

# 26. Non-Retryable Requests

Avoid automatically retrying operations such as:

```text
Create Payment
Create Ride
Assign Driver
Cancel Ride
```

unless the operation has explicit idempotency protection.

---

# 27. Idempotency

Retries must be combined with idempotency where duplicate execution could produce an incorrect business result.

This is especially important for:

```text
Ride Creation
Payment
Driver Assignment
State Transitions
Notifications
```

---

# 28. Exponential Backoff

Retryable synchronous calls should use:

```text
Bounded Retry
Exponential Backoff
Jitter
```

Avoid immediate retry loops.

---

# 29. Retry Budget

Each request should have a limited retry budget.

Conceptually:

```text
Request Deadline
      ↓
Retry 1
      ↓
Retry 2
      ↓
Stop
```

The retry strategy must respect the overall request deadline.

---

# 30. Circuit Breaking

For repeatedly failing dependencies, a circuit-breaker mechanism may be used.

Conceptually:

```text
Healthy
   ↓
Failures Increase
   ↓
Open Circuit
   ↓
Reject Fast
   ↓
Recovery Probe
   ↓
Healthy
```

Circuit breaking should be introduced where failure propagation is a real operational concern.

---

# 31. Bulkheads

High-risk dependencies should not be allowed to consume all service resources.

Use bounded:

```text
Concurrency
Connection Pools
Worker Pools
Queues
```

where required.

---

# 32. Service Availability

A synchronous dependency should not automatically make the entire platform unavailable.

Where possible:

```text
Dependency Failure
      ↓
Fallback / Degraded Behaviour
```

should be considered.

---

# 33. Asynchronous Preference for Non-Critical Work

The following often belongs in asynchronous processing:

```text
Notifications
Analytics
Search Index Updates
AI Feature Processing
Operational Metrics
Historical Data Processing
```

This prevents non-critical work from extending request latency.

---

# 34. Ride Creation Example

A simplified ride creation workflow may be:

```text
Client
  ↓
Ride API
  ↓
Validate Request
  ↓
Create Ride + Outbox
  ↓
Commit
  ↓
Return Ride Created
```

Then:

```text
Outbox
  ↓
Redpanda
  ↓
Matching / Dispatch
```

This prevents ride creation from depending synchronously on the complete matching workflow.

---

# 35. Matching Example

After:

```text
RideCreated
```

the matching service can consume the event:

```text
Redpanda
   ↓
Matching Service
   ↓
Candidate Discovery
   ↓
Ranking
   ↓
Assignment
```

The result can then produce:

```text
DriverAssigned
```

---

# 36. Dispatch Example

Dispatch-related state changes should use appropriate domain ownership.

For example:

```text
Matching Service
   ↓
Driver Assignment
   ↓
Ride Service / Authoritative State
   ↓
DriverAssigned Event
```

The exact transaction boundary must remain explicit.

---

# 37. AI Communication

AI services should not unnecessarily sit in critical synchronous request paths.

For example:

```text
Ride Request
   ↓
Core Ride Creation
   ↓
Event
   ↓
AI / Smart Dispatch
```

is often preferable to:

```text
Ride Request
   ↓
AI Service
   ↓
Ride Service
```

when the business flow allows asynchronous AI processing.

---

# 38. AI in Critical Path

If AI must provide an immediate decision, it may be called synchronously.

However, the system must define:

```text
Timeout
Fallback
Failure Policy
Cost Limit
```

before placing AI in the critical path.

---

# 39. ETA Communication

ETA may require synchronous communication when the user requests a current ETA.

Conceptually:

```text
Client
  ↓
ETA API
  ↓
Current Location
  ↓
Route Provider / ETA Engine
  ↓
Response
```

Caching and asynchronous updates may be used where appropriate.

---

# 40. External Provider Communication

External providers include:

```text
Maps / Routing
Payments
SMS
WhatsApp
Push Notifications
Identity Services
```

These dependencies should have explicit:

```text
Timeout
Retry
Rate Limit
Circuit Breaker
Fallback
```

policies.

---

# 41. External API Isolation

Do not expose external provider-specific contracts directly to RideForge domain consumers.

Prefer:

```text
RideForge Interface
      ↓
Provider Adapter
      ↓
External Provider
```

This allows providers to be replaced or changed without spreading provider-specific code across services.

---

# 42. Route Provider Adapter

Conceptually:

```text
ETA Service
    ↓
Route Provider Interface
    ↓
Provider Adapter
    ├── Provider A
    ├── Provider B
    └── Future Provider
```

---

# 43. Payment Provider Adapter

Similarly:

```text
Payment Service
    ↓
Payment Interface
    ↓
Provider Adapter
    ├── Razorpay
    ├── Stripe
    └── Future Provider
```

The business domain should not depend directly on provider SDK semantics.

---

# 44. Notification Providers

Notification delivery may be asynchronous:

```text
Domain Event
   ↓
Notification Consumer
   ↓
Notification Provider
```

This prevents third-party notification latency from blocking the originating transaction.

---

# 45. Event-Driven Communication

Events should represent facts that have occurred.

Examples:

```text
RideCreated
DriverAssigned
RideStarted
RideCompleted
PaymentCompleted
```

Events should not be confused with commands.

---

# 46. Commands vs Events

A command:

```text
Do Something
```

An event:

```text
Something Happened
```

Examples:

```text
AssignDriver
```

is a command.

```text
DriverAssigned
```

is an event.

---

# 47. Event Ownership

The producing service owns the event contract.

Consumers should not mutate the original event.

Events should be treated as immutable messages.

---

# 48. Event Versioning

Event contracts should support controlled evolution.

A conceptual envelope includes:

```text
event_type
event_version
event_id
aggregate_id
occurred_at
payload
```

---

# 49. Event Delivery Semantics

RideForge uses:

```text
At-Least-Once Event Processing
```

where required.

Consumers must therefore support:

```text
Duplicate Delivery
Retry
Replay
Out-of-Order Conditions Where Applicable
```

---

# 50. Event Partitioning

When ordering is required for an aggregate:

```text
aggregate_id
```

should generally be used as the partition key.

For example:

```text
ride_id
```

for ride lifecycle events.

---

# 51. Event Consumer Groups

Different capabilities should use independent consumer groups.

For example:

```text
ride.events
   ├── matching-group
   ├── notification-group
   ├── analytics-group
   └── ai-group
```

A failure in one consumer should not prevent independent consumers from receiving the event.

---

# 52. Synchronous vs Asynchronous Decision Matrix

| Requirement | Synchronous HTTP | Asynchronous Event |
|---|---:|---:|
| Immediate response required | **Yes** | No |
| User-facing request | **Usually** | Sometimes |
| Durable business event | No | **Yes** |
| Notifications | Usually No | **Yes** |
| Analytics | No | **Yes** |
| AI background processing | No | **Yes** |
| Current ETA | **Yes** | Sometimes |
| Long-running operation | No | **Yes** |
| Cross-service state propagation | Conditional | **Often** |
| Critical validation | **Yes** | No |
| Eventual consistency acceptable | Sometimes | **Yes** |

---

# 53. API Gateway / Edge Layer

RideForge should provide a controlled API boundary for external clients.

The edge layer may handle:

```text
Authentication
Authorization
Rate Limiting
Routing
Request Validation
Tracing
API Versioning
```

Internal services should not all need to implement external-facing concerns independently.

---

# 54. Internal Service APIs

Internal APIs should remain accessible only through appropriate private network boundaries.

Avoid exposing internal service endpoints directly to the public internet.

---

# 55. Service Discovery

Service-to-service communication requires a way to resolve service endpoints.

The deployment environment may provide:

```text
DNS
Container Service Discovery
Kubernetes Service
Service Registry
Cloud Service Discovery
```

The architecture should depend on logical service names rather than hard-coded IP addresses.

---

# 56. Service Address Configuration

Avoid:

```text
http://10.0.0.25:8080
```

hard-coded in application code.

Prefer configurable logical addresses such as:

```text
RIDE_SERVICE_URL
DRIVER_SERVICE_URL
ETA_SERVICE_URL
```

or platform-native service discovery.

---

# 57. Service Identity

Service-to-service requests should carry an identifiable service identity.

Conceptually:

```text
Caller Service
      ↓
Authentication
      ↓
Target Service
```

The exact authentication mechanism depends on deployment security architecture.

---

# 58. Authentication

Internal APIs must not assume that being inside a private network is sufficient authorization.

Use appropriate:

```text
Service Credentials
Tokens
mTLS
Signed Requests
Platform Identity
```

depending on the deployment environment.

---

# 59. Authorization

Authentication answers:

```text
Who Are You?
```

Authorization answers:

```text
What Are You Allowed To Do?
```

Both must be considered for sensitive internal APIs.

---

# 60. Least Privilege

A service should receive only the permissions required for its responsibilities.

For example:

```text
Matching Service
→ Matching-related operations

Payment Service
→ Payment-related operations
```

Avoid broad cross-service permissions.

---

# 61. Request Context

Synchronous requests should propagate useful context such as:

```text
Correlation ID
Trace ID
Request ID
Deadline
Authentication Context
```

where appropriate.

---

# 62. Correlation ID

A correlation ID allows a complete request workflow to be traced:

```text
Client
  ↓
API
  ↓
Ride Service
  ↓
Event
  ↓
Matching
  ↓
Driver Assignment
```

---

# 63. Trace Context

Distributed tracing should propagate trace context across:

```text
HTTP
Redpanda
Background Workers
External Calls
```

where supported.

---

# 64. Request ID

A request ID identifies an individual request.

It may differ from:

```text
Correlation ID
```

which can represent a broader business workflow.

---

# 65. Error Contract

All synchronous APIs should use a consistent error structure.

Conceptually:

```text
code
message
details
request_id
```

The exact contract should be defined centrally.

---

# 66. Error Categories

Common categories include:

```text
Validation Error
Authentication Error
Authorization Error
Not Found
Conflict
Rate Limited
Dependency Failure
Internal Error
```

---

# 67. Do Not Leak Internal Errors

Internal service errors should not expose:

```text
Database Details
Stack Traces
Secrets
Infrastructure Addresses
Internal Credentials
```

to external clients.

---

# 68. HTTP Status Codes

Use HTTP status codes consistently.

Examples:

```text
200 OK
201 Created
204 No Content
400 Bad Request
401 Unauthorized
403 Forbidden
404 Not Found
409 Conflict
422 Unprocessable Entity
429 Too Many Requests
500 Internal Server Error
502 Bad Gateway
503 Service Unavailable
504 Gateway Timeout
```

Only use codes that accurately represent the failure.

---

# 69. Synchronous Error Propagation

A service should translate downstream failures into an appropriate response.

Avoid blindly returning raw downstream errors.

Conceptually:

```text
Service B
   ↓
503
   ↓
Service A
   ↓
Controlled Application Error
   ↓
Client
```

---

# 70. Partial Failure

A service may successfully complete its local operation while a downstream service fails.

The architecture should distinguish:

```text
Local Transaction Success
```

from:

```text
Asynchronous Workflow Completion
```

This is another reason to use events for independent downstream work.

---

# 71. Long-Running Operations

For operations that take significant time:

```text
POST /operation
```

may return:

```text
Accepted
+
Operation ID
```

followed by asynchronous processing.

Conceptually:

```text
Request
 ↓
202 Accepted
 ↓
Background Processing
 ↓
Event / Status
```

---

# 72. Synchronous Service Composition

Synchronous composition should be limited to operations where immediate consistency or a response is necessary.

Avoid synchronous composition for:

```text
Notifications
Analytics
Historical Storage
AI Training
Non-Critical Enrichment
```

---

# 73. Communication During Degradation

When a dependency fails, the service should follow its defined degradation policy.

Possible outcomes include:

```text
Fail Fast
Fallback
Return Cached Result
Return Partial Result
Queue Work
```

The choice depends on business criticality.

---

# 74. Fallbacks

Fallbacks must not silently return misleading data.

For example:

```text
Stale Driver Location
```

must not be presented as:

```text
Current Driver Location
```

---

# 75. Communication and Data Consistency

Synchronous APIs can provide immediate responses, but they do not automatically provide distributed transactionality.

For cross-service state changes:

```text
Local Transaction
+
Event
```

is generally preferred over distributed transactions.

---

# 76. Distributed Transactions

RideForge will avoid distributed transactions across service boundaries by default.

Prefer:

```text
Local Transaction
      ↓
Outbox
      ↓
Event
      ↓
Consumer
```

rather than:

```text
Service A Transaction
+
Service B Transaction
+
Service C Transaction
```

---

# 77. Saga / Workflow Considerations

Complex multi-service workflows may eventually require:

```text
Saga
Process Manager
Workflow Engine
```

but these should be introduced only when the workflow complexity justifies them.

The initial architecture should prefer local transactions plus events.

---

# 78. Communication and Transactions

A service should not keep a local database transaction open while making synchronous calls to another service.

Avoid:

```text
BEGIN
  ↓
Call Service B
  ↓
Wait
  ↓
Call Service C
  ↓
COMMIT
```

This creates long transactions and distributed coupling.

Prefer:

```text
Local Transaction
  ↓
COMMIT
  ↓
Asynchronous Event
```

where business semantics permit.

---

# 79. Communication and Locks

Avoid holding database locks while waiting for remote services.

Remote latency is unpredictable.

---

# 80. API Rate Limiting

Internal APIs may require rate limits to protect critical services from excessive callers.

Rate limits should be based on:

```text
Service
Endpoint
Operation
Business Criticality
```

---

# 81. Backpressure

Synchronous APIs should enforce bounded concurrency.

Asynchronous consumers should enforce:

```text
Consumer Concurrency
Batch Size
Queue Capacity
```

This prevents overload from propagating through the system.

---

# 82. Connection Pools

Each service must use bounded connection pools for:

```text
PostgreSQL
Redis
HTTP Clients
```

where applicable.

The connection strategy must remain consistent with:

```text
ADR-0011
```

for PostgreSQL and PgBouncer.

---

# 83. HTTP Client Reuse

Services should reuse HTTP clients and transport connections rather than creating a new HTTP client for every request.

This reduces:

```text
Connection Overhead
Latency
Resource Consumption
```

---

# 84. HTTP Keep-Alive

Persistent HTTP connections should be used where supported and appropriate.

Connection pool settings should be tuned according to actual traffic.

---

# 85. HTTP Client Timeouts

At minimum, consider:

```text
Connection Timeout
Request Timeout
Response Header Timeout
Idle Connection Timeout
```

according to the chosen Go HTTP client configuration.

---

# 86. Service Communication Metrics

Track:

```text
Request Count
Success Rate
Error Rate
Latency
Timeouts
Retries
Circuit Breaker State
Connection Pool Usage
```

---

# 87. Event Communication Metrics

Track:

```text
Published Events
Consumed Events
Consumer Lag
Processing Latency
Retries
DLQ Count
Replay Count
```

---

# 88. Dependency Metrics

For each important downstream service:

```text
Request Rate
P50 Latency
P95 Latency
P99 Latency
Error Rate
Timeout Rate
```

should be observable.

---

# 89. Logging

Service communication logs should include:

```text
Request ID
Correlation ID
Trace ID
Caller
Target
Operation
Latency
Status
Error Code
```

Avoid logging sensitive payloads unnecessarily.

---

# 90. Distributed Tracing

Distributed tracing should show:

```text
Incoming Request
   ↓
Service A
   ↓
Service B
   ↓
Database
```

and asynchronous flows:

```text
Service A
   ↓
Outbox
   ↓
Redpanda
   ↓
Service B
```

---

# 91. Eventual Consistency Visibility

Operational dashboards should distinguish:

```text
Transaction Completed
```

from:

```text
Downstream Event Processed
```

This helps operators understand asynchronous propagation delays.

---

# 92. Health Checks

Services should expose appropriate health endpoints.

Typical concepts:

```text
Liveness
Readiness
Dependency Health
```

Health checks must not perform expensive business operations.

---

# 93. Readiness

A service should be considered ready only when it can perform the operations required by its role.

For example:

```text
Database Connectivity
Required Configuration
Required Dependencies
```

should be validated according to the service's startup requirements.

---

# 94. Liveness

Liveness should primarily determine whether the process is functioning.

Do not make liveness depend on every external dependency unless there is a strong operational reason.

Otherwise a dependency failure can cause cascading restarts.

---

# 95. Graceful Shutdown

Services must stop accepting new work before terminating.

A conceptual sequence is:

```text
Stop New Requests
      ↓
Stop New Event Consumption
      ↓
Finish In-Flight Work
      ↓
Close Connections
      ↓
Exit
```

---

# 96. Event Consumer Shutdown

Consumers should stop taking new messages before shutdown while allowing in-flight processing to complete where possible.

Offsets must not be committed for work that did not complete successfully.

---

# 97. API Deployment Compatibility

During rolling deployments:

```text
Old Service
+
New Service
```

may run simultaneously.

Therefore API contracts should remain backward-compatible during the deployment window.

---

# 98. Event Deployment Compatibility

Similarly:

```text
Old Consumer
+
New Producer
```

may coexist.

Event schema evolution must support this transition.

---

# 99. Service Contract Testing

Contract tests should verify:

```text
Request Schema
Response Schema
Error Schema
Event Schema
Compatibility
```

between producers and consumers.

---

# 100. Integration Testing

Integration tests should verify:

```text
Service A
   ↓
Service B
```

and:

```text
Service A
   ↓
Redpanda
   ↓
Service B
```

using production-like infrastructure where practical.

---

# 101. Failure Testing

Test:

```text
Service Unavailable
Network Timeout
Database Failure
Redpanda Failure
Dependency Timeout
Duplicate Event
Out-of-Order Event
```

and verify defined degradation behaviour.

---

# 102. Load Testing

Measure service communication under:

```text
Normal Load
Peak Load
Burst Load
Dependency Slowdown
Dependency Failure
```

---

# 103. Communication Cost

Synchronous service calls have a direct latency and infrastructure cost.

Asynchronous events have:

```text
Broker Cost
Storage
Consumer Complexity
Eventual Consistency
```

The simplest architecture satisfying the business requirement should be selected.

---

# 104. Avoid Chatty Services

A service should not make many small calls when a well-designed operation can provide the required information efficiently.

Avoid:

```text
Service A
 → B
 → B
 → B
 → B
```

Prefer:

```text
One Purposeful Request
```

or:

```text
Local Read Model
```

when justified.

---

# 105. Read Models

For high-volume read requirements, asynchronous projections may be preferable to repeated synchronous calls.

Conceptually:

```text
Domain Event
   ↓
Projection Consumer
   ↓
Read Model
   ↓
Fast Query
```

This may reduce cross-service request chains.

---

# 106. Caching

Caching may reduce synchronous dependency pressure.

Use:

```text
Redis
```

where appropriate.

Cached data must have:

```text
Freshness
Invalidation
Fallback
```

semantics.

---

# 107. Cache Is Not Source of Truth

A cached service response must not become an accidental authoritative data store.

The owning service remains authoritative for its business state.

---

# 108. Service Communication Security

Communication channels should provide:

```text
Authentication
Authorization
Encryption
Input Validation
Rate Limiting
Auditability
```

according to the trust boundary.

---

# 109. Secret Management

Service credentials must come from the approved configuration and secret-management mechanism.

Never embed:

```text
API Keys
Database Passwords
Tokens
Private Keys
```

in source code.

---

# 110. External API Credentials

External provider credentials should be scoped to the specific service that needs them.

Do not share one broad credential across unrelated services.

---

# 111. Network Boundaries

A preferred production model is:

```text
Public Internet
      ↓
API / Edge
      ↓
Private Services
      ↓
Private Infrastructure
```

Only required endpoints should be public.

---

# 112. Communication Failure Taxonomy

Failures should be categorized as:

```text
Timeout
Connection Failure
Authentication Failure
Authorization Failure
Validation Failure
Rate Limit
Service Unavailable
Dependency Failure
Protocol Error
Business Error
```

This supports consistent retry and fallback decisions.

---

# 113. Retry Classification Matrix

| Failure | Default Action |
|---|---|
| Timeout | Retry only if safe |
| Connection Failure | Bounded Retry |
| 429 | Respect Retry-After / Backoff |
| 500 | Retry only if operation is safe |
| 502 | Bounded Retry |
| 503 | Bounded Retry |
| 504 | Retry only if safe |
| 400 | Do Not Retry |
| 401 | Do Not Blindly Retry |
| 403 | Do Not Retry |
| 404 | Usually Do Not Retry |
| Business Conflict | Resolve / Return Conflict |
| Invalid Payload | Do Not Retry |

The exact policy depends on operation semantics.

---

# 114. Synchronous Communication Checklist

```text
[ ] Timeout configured
[ ] Retry policy defined
[ ] Idempotency considered
[ ] Error mapping defined
[ ] Authentication configured
[ ] Authorization configured
[ ] Connection pooling configured
[ ] Metrics configured
[ ] Tracing configured
[ ] Failure fallback defined
```

---

# 115. Asynchronous Communication Checklist

```text
[ ] Event contract defined
[ ] Event version defined
[ ] Event ID defined
[ ] Partition key defined
[ ] Producer ownership defined
[ ] Consumer group defined
[ ] Idempotency strategy defined
[ ] Retry policy defined
[ ] DLQ strategy defined
[ ] Observability configured
```

---

# 116. Decision Matrix

| Communication Option | Primary Use | RideForge Decision |
|---|---|---|
| HTTP API | Immediate request/response | **Primary synchronous mechanism** |
| Redpanda Event | Asynchronous state propagation | **Primary asynchronous mechanism** |
| Direct DB Access | Cross-service data sharing | **Rejected** |
| Shared Redis State | Cross-service ownership | **Restricted** |
| Distributed Transaction | Cross-service atomicity | **Avoid by default** |
| Workflow Engine | Complex long-running workflows | **Future / conditional** |
| gRPC | High-performance internal RPC | **Future / conditional** |

---

# 117. gRPC Consideration

gRPC may eventually be useful for:

```text
High-Frequency Internal RPC
Strict Typed Contracts
Low-Latency Service Communication
Streaming
```

However, introducing a second synchronous protocol adds:

```text
Operational Complexity
Tooling
Deployment Configuration
Contract Management
```

Therefore HTTP remains the initial standard unless measured requirements justify gRPC.

---

# 118. Service Mesh Consideration

A service mesh may eventually provide:

```text
mTLS
Traffic Management
Retries
Observability
Service Discovery
```

but it should not be introduced prematurely.

RideForge should first establish these requirements at the application and infrastructure levels.

---

# 119. API Gateway vs Service Mesh

These solve different problems.

```text
API Gateway
→ External / Edge Traffic

Service Mesh
→ Service-to-Service Traffic
```

Neither is a substitute for clear service contracts.

---

# 120. Initial Architecture

The initial communication architecture is intentionally:

```text
External
   ↓
API / Edge
   ↓
HTTP
   ↓
Services
   │
   ├── PostgreSQL
   ├── Redis
   └── Outbox
          ↓
       Redpanda
          ↓
       Consumers
```

This keeps the architecture understandable while supporting future evolution.

---

# 121. Future Evolution

The communication architecture may evolve toward:

```text
API Gateway
+
Service Discovery
+
HTTP / gRPC
+
Redpanda
+
Service Mesh
```

only when scale or operational requirements justify additional infrastructure.

---

# 122. Consequences

## 122.1 Positive Consequences

The decision provides:

```text
Clear Communication Boundaries
Simple Initial Architecture
Immediate Request/Response Support
Asynchronous Event Processing
Reduced Distributed Coupling
Explicit Service Ownership
Failure Isolation
Scalable Event Consumption
```

---

## 122.2 Negative Consequences

The architecture introduces:

```text
Two Communication Models
Eventual Consistency
Retry Complexity
Distributed Debugging
API Contract Management
Event Contract Management
Additional Observability Requirements
```

These trade-offs are accepted.

---

# 123. Risks

## Risk 1 — Too Many Synchronous Dependencies

### Mitigation

Use:

```text
Events
Read Models
Caching
Bounded Dependency Chains
```

where appropriate.

---

## Risk 2 — Eventual Consistency Misunderstanding

### Mitigation

Document:

```text
Authoritative State
Event Propagation
Expected Delay
```

for each asynchronous workflow.

---

## Risk 3 — Retry Storm

### Mitigation

Use:

```text
Bounded Retries
Exponential Backoff
Jitter
Circuit Breaking
```

where required.

---

## Risk 4 — Duplicate Event Processing

### Mitigation

Use:

```text
Stable event_id
Consumer Idempotency
Transactional Consumer State
```

where required.

---

## Risk 5 — API Contract Drift

### Mitigation

Use:

```text
Contract Testing
Versioning
Backward Compatibility
Code Review
```

---

## Risk 6 — Service Coupling Through APIs

### Mitigation

Keep APIs:

```text
Capability-Oriented
Minimal
Explicit
Owned by One Service
```

---

# 124. Validation

The API and service communication strategy should be validated through:

```text
API Contract Tests
Integration Tests
Event Contract Tests
Failure Tests
Timeout Tests
Retry Tests
Load Tests
Security Tests
Distributed Tracing
Production Metrics
```

---

# 125. Review Triggers

Revisit this ADR when:

```text
Service Count Changes Significantly
HTTP Becomes a Measured Bottleneck
gRPC Is Proposed
Service Mesh Is Proposed
Multi-Region Deployment Is Introduced
Service Discovery Requirements Change
Synchronous Chains Become Excessive
Event Volume Changes Significantly
A Workflow Engine Is Introduced
```

---

# 126. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
```

Especially:

```text
API Development
Event and Messaging Development
Error Handling and Validation
Configuration and Environment
Logging and Debugging
Observability Development
Performance and Optimization
Integration Testing and Local Infrastructure
Security and Secret Management
```

---

# 127. Related ADRs

This decision is directly related to:

```text
ADR-0002 — Architecture Style
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0011 — PgBouncer for Database Connection Pooling
ADR-0012 — Outbox Pattern
ADR-0013 — Dead Letter Queue Strategy
ADR-0015 — Smart Dispatch and Stand Dispatch
ADR-0016 — AI-Assisted Dispatch Strategy
ADR-0017 — ETA and Route Provider Strategy
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
ADR-0023 — Security and Secret Management
ADR-0024 — Configuration and Environment Strategy
ADR-0025 — Testing and Integration Strategy
```

---

# 128. Decision Summary

RideForge adopts a hybrid communication architecture:

```text
Synchronous HTTP
+
Asynchronous Redpanda Events
```

The decision can be summarized as:

```text
Immediate Result Required
        ↓
   HTTP API

Independent / Background Work
        ↓
   Redpanda Event

Cross-Service Transaction
        ↓
Local Transaction
+
Outbox
+
Event

Cross-Service Database Access
        ↓
        Avoid
```

The preferred service communication model is:

```text
                    External Clients
                           │
                           ▼
                     API / Edge
                           │
                     HTTP / HTTPS
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
   Ride Service       Driver Service     User Service
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                     Local Databases
                           │
                     Transactional
                         Outbox
                           │
                           ▼
                       Redpanda
                           │
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
      Matching         Notification       AI / Data
       Consumer          Consumer          Consumer
```

---

# 129. Final Principle

> **Use synchronous HTTP for operations that require an immediate response and asynchronous Redpanda events for independent work and state propagation; keep service ownership explicit, avoid cross-service database access, bound remote calls with timeouts and retries, and use local transactions plus the outbox pattern instead of distributed transactions.**

The architecture intentionally keeps the initial communication model simple:

```text
HTTP
+
Redpanda
+
Explicit Contracts
+
Bounded Timeouts
+
Idempotent Processing
+
Strong Observability
```

This provides a practical foundation for RideForge while leaving room for future adoption of:

```text
gRPC
Service Mesh
Advanced Service Discovery
Workflow Engines
Regional Service Communication
```

when actual system requirements justify them.

---

# 130. Status

```text
Decision: ACCEPTED

Synchronous Protocol:
HTTP / HTTPS

Asynchronous Protocol:
Redpanda Events

Cross-Service Database Access:
Not Allowed

Distributed Transactions:
Avoided by Default

Reliable Event Publication:
Transactional Outbox

Failure Handling:
Timeout + Bounded Retry + Fallback / DLQ

Service Discovery:
Platform / Environment-Based

Communication Security:
Authentication + Authorization + Encryption

Primary Goal:
Clear, Reliable, Observable Service Communication
```

This decision establishes the standard RideForge communication model between external clients, internal services, event infrastructure, and external providers while preserving the platform's microservice boundaries and event-driven architecture.
