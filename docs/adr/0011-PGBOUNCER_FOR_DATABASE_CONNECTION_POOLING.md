# ADR-0011: PgBouncer for Database Connection Pooling

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Database Infrastructure / Performance  
> **Scope:** RideForge PostgreSQL connection management  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge uses PostgreSQL as its primary transactional database.

The platform is designed around:

```text
Go Services
PostgreSQL
Redis
Redpanda
Event-Driven Processing
Concurrent API Requests
Background Workers
```

Ride-hailing workloads can create significant concurrency.

For example:

```text
Ride Requests
Driver Updates
Matching
Dispatch
Payments
User Requests
Background Consumers
Administrative Operations
```

may execute concurrently.

Each application process may maintain PostgreSQL connections through its own connection pool.

As the number of application instances and concurrent workers increases, the total number of PostgreSQL connections can grow substantially.

PostgreSQL connections are a finite resource.

Excessive connection creation can result in:

```text
Connection Exhaustion
Increased Memory Usage
Connection Queueing
Higher Latency
Database Instability
Failed Requests
```

RideForge therefore needs a deliberate connection-management strategy.

---

# 2. Problem

RideForge must efficiently manage PostgreSQL connections across:

```text
Multiple Service Instances
Multiple Application Workers
Concurrent Requests
Background Consumers
Scaling Events
```

The solution must:

```text
Protect PostgreSQL
Control Connection Counts
Reduce Connection Establishment Overhead
Support Application Concurrency
Improve Resource Utilization
Preserve Transaction Correctness
```

The platform also needs to avoid treating connection pooling as a substitute for query optimization or database capacity planning.

---

# 3. Decision

RideForge will support:

> **PgBouncer as the database connection-pooling layer when PostgreSQL connection pressure or deployment topology justifies an external pool.**

The preferred architecture is:

```text
RideForge Services
       ↓
Application DB Pool
       ↓
PgBouncer
       ↓
PostgreSQL
```

PgBouncer is an infrastructure-level connection pool.

It does not replace:

```text
Application Pooling
Query Optimization
Transaction Management
PostgreSQL Capacity Planning
```

---

# 4. Primary Principle

> **PgBouncer is used to control and multiplex PostgreSQL connections; it is not a solution for inefficient queries, oversized transactions, or insufficient database capacity.**

---

# 5. Why Connection Pooling Is Required

Creating a new PostgreSQL connection for every request is inefficient.

Without pooling:

```text
Request
  ↓
Create Connection
  ↓
Authenticate / Establish Session
  ↓
Execute Query
  ↓
Close Connection
```

With application pooling:

```text
Application
    ↓
Connection Pool
    ↓
Reuse Connection
```

With PgBouncer:

```text
Application Pools
       ↓
   PgBouncer
       ↓
PostgreSQL Connections
```

This allows a larger number of application-side clients to share a controlled number of actual PostgreSQL connections.

---

# 6. Connection Exhaustion

PostgreSQL has a finite connection capacity.

A simplified example:

```text
10 Service Instances
×
20 Application Connections
=
200 Potential PostgreSQL Connections
```

The actual capacity also needs to account for:

```text
Administrative Connections
Background Workers
Monitoring
Migrations
Other Consumers
```

Therefore application pool sizes cannot be selected independently.

---

# 7. Connection Budget

RideForge should treat database connections as a shared resource.

A conceptual budget is:

```text
PostgreSQL Connection Capacity
        │
        ├── Application Traffic
        ├── Background Workers
        ├── Administrative Access
        ├── Monitoring
        └── Operational Headroom
```

The system should maintain sufficient headroom rather than operating continuously at the connection limit.

---

# 8. PgBouncer Role

PgBouncer sits between application processes and PostgreSQL:

```text
                ┌─────────────────┐
                │ RideForge       │
                │ Services        │
                └────────┬────────┘
                         │
                  Application Pool
                         │
                         ▼
                ┌─────────────────┐
                │    PgBouncer    │
                └────────┬────────┘
                         │
                  PostgreSQL Pool
                         │
                         ▼
                ┌─────────────────┐
                │   PostgreSQL    │
                └─────────────────┘
```

The purpose is to reduce the number of backend PostgreSQL connections required to serve application concurrency.

---

# 9. Pooling Layers

RideForge may have two pooling layers:

```text
Application Pool
       ↓
PgBouncer Pool
       ↓
PostgreSQL
```

These layers serve different purposes.

The application pool controls:

```text
How many requests can concurrently use DB connections
```

PgBouncer controls:

```text
How many PostgreSQL backend connections are maintained
```

Both must be configured together.

---

# 10. Pool Sizing Principle

Do not configure:

```text
Application Pool Size
```

and:

```text
PgBouncer Pool Size
```

independently.

They must be considered together with:

```text
Number of Instances
Request Concurrency
Query Duration
Worker Count
Database Capacity
```

---

# 11. Example Connection Budget

Consider:

```text
8 Application Instances
10 Connections Per Instance
```

This creates:

```text
80 Application-Side Connections
```

Without an external pool, PostgreSQL may need to support most of those connections.

With PgBouncer, the application layer may maintain many client connections while PgBouncer maintains a smaller controlled number of active PostgreSQL backend connections.

The actual numbers must be determined through load testing and production metrics.

---

# 12. Transaction Pooling

RideForge should prefer:

> **Transaction pooling**

when application compatibility allows it.

Conceptually:

```text
BEGIN
  ↓
PostgreSQL Connection Assigned
  ↓
Queries
  ↓
COMMIT
  ↓
Connection Returned to Pool
```

After the transaction completes, PgBouncer can reuse the backend PostgreSQL connection for another client.

This provides high connection multiplexing efficiency.

---

# 13. Session Pooling

Session pooling keeps a PostgreSQL backend connection associated with a client session.

Conceptually:

```text
Client
  ↓
PgBouncer
  ↓
Dedicated Backend Connection
  ↓
Until Session Ends
```

This provides stronger session compatibility but reduces multiplexing efficiency.

Session pooling should be used only when required by application behaviour.

---

# 14. Statement Pooling

Statement pooling provides even finer-grained multiplexing.

However, it is generally more restrictive and may not be suitable for RideForge's transactional application workloads.

Therefore it should not be the default mode.

---

# 15. Pooling Mode Decision

The preferred hierarchy is:

```text
Transaction Pooling
        ↓
Session Pooling If Required
        ↓
Statement Pooling Only For Specific Compatible Workloads
```

The actual production mode must be validated against:

```text
Driver Behaviour
Transactions
Prepared Statements
Session State
Extensions
Application Libraries
```

---

# 16. Go PostgreSQL Drivers

RideForge services are implemented primarily in Go.

The selected PostgreSQL driver and database library must be tested for compatibility with PgBouncer's selected pooling mode.

Particular attention should be given to:

```text
Prepared Statements
Transactions
Connection State
Session Variables
Temporary Objects
LISTEN / NOTIFY
Advisory Locks
```

---

# 17. Prepared Statements

Prepared statements can have session-level behaviour depending on the client and driver configuration.

With transaction pooling:

```text
Client
  ↓
Backend Connection A
  ↓
Transaction Ends
  ↓
Backend Connection A Reused
```

The next transaction may use a different backend connection.

Therefore prepared-statement behaviour must be explicitly tested.

---

# 18. Application Compatibility

Before enabling transaction pooling in production, verify:

```text
Go PostgreSQL Driver
ORM / Query Layer
Prepared Statement Behaviour
Transaction Behaviour
Session State Usage
Advisory Lock Usage
Temporary Tables
Connection-Level Settings
```

The implementation should not assume compatibility without testing.

---

# 19. Session State

Application code should minimize reliance on persistent PostgreSQL session state.

Avoid unnecessary use of:

```text
SET SESSION
Temporary Tables
Session-Specific Configuration
Persistent Prepared Statements
Session-Level Advisory State
```

when using transaction pooling.

---

# 20. Transaction Boundaries

PgBouncer does not define application transaction boundaries.

The application remains responsible for:

```text
BEGIN
Transaction Work
COMMIT / ROLLBACK
```

Transactions should remain:

```text
Short
Focused
Predictable
```

---

# 21. Long Transactions

Long-running transactions reduce the benefits of pooling.

They can also cause:

```text
Connection Occupancy
Lock Retention
Vacuum Delays
Resource Contention
```

Therefore transaction duration must be monitored.

---

# 22. External Calls Inside Transactions

Application code should not hold PostgreSQL transactions open while waiting for:

```text
External APIs
Routing Providers
Payment Providers
AI Services
Network Calls
User Input
```

Prefer:

```text
Database Transaction
      ↓
Commit
      ↓
External Operation
```

where business semantics permit.

---

# 23. Connection Lifetime

A PostgreSQL connection should be treated as a limited resource.

Application code must:

```text
Acquire
Use
Release
```

connections correctly.

Connection leaks can exhaust both:

```text
Application Pool
PgBouncer
PostgreSQL
```

---

# 24. Application Pool vs PgBouncer

The architecture should avoid excessive application-side pooling.

For example:

```text
Many Application Instances
×
Large Application Pools
```

can still overwhelm PgBouncer or create unnecessary client-side resource usage.

Pool sizes must reflect actual concurrency.

---

# 25. PgBouncer Pool Configuration

Important concepts include:

```text
max_client_conn
default_pool_size
min_pool_size
reserve_pool_size
reserve_pool_timeout
server_idle_timeout
server_lifetime
query_wait_timeout
```

The exact values are environment-specific.

They must be derived from:

```text
Load
Database Capacity
Service Count
Latency
Traffic Patterns
```

rather than copied blindly.

---

# 26. Client Connection Limits

PgBouncer should enforce a reasonable maximum number of client connections.

This protects PgBouncer itself from uncontrolled application growth.

Conceptually:

```text
Application Concurrency
        ↓
PgBouncer Client Limit
        ↓
Controlled Queue
        ↓
PostgreSQL Backend Pool
```

---

# 27. Backend Connection Limits

The number of PostgreSQL backend connections should be deliberately bounded.

A simplified capacity model is:

```text
PostgreSQL max_connections
        >
PgBouncer Backend Pool
        +
Operational Headroom
```

The exact capacity must account for PostgreSQL-specific reserved and administrative connections.

---

# 28. Queueing

If all backend connections are busy:

```text
Client
  ↓
PgBouncer
  ↓
Wait
  ↓
Backend Connection Available
```

This is preferable to allowing unlimited direct connections to PostgreSQL.

However, excessive queueing increases latency.

---

# 29. Query Wait Time

A request waiting too long for a database connection should fail predictably rather than remaining blocked indefinitely.

Use appropriate:

```text
Connection Timeout
Query Timeout
Application Deadline
```

policies.

---

# 30. Backpressure

PgBouncer provides an important backpressure boundary.

Instead of:

```text
Unlimited Application Requests
        ↓
Unlimited PostgreSQL Connections
```

the system should use:

```text
Application
      ↓
Bounded Pool
      ↓
PgBouncer
      ↓
Bounded PostgreSQL Connections
```

---

# 31. PgBouncer Does Not Increase Database Capacity

PgBouncer can improve:

```text
Connection Utilization
Connection Multiplexing
Connection Stability
```

but it does not automatically increase:

```text
CPU
Memory
Disk I/O
Query Processing Capacity
```

A slow query remains slow.

---

# 32. Query Optimization Remains Required

Before increasing pool sizes, investigate:

```text
Slow Queries
Missing Indexes
Large Result Sets
N+1 Queries
Lock Contention
Long Transactions
```

Increasing connections can make an overloaded database worse.

---

# 33. Connection Storm Protection

PgBouncer helps prevent connection storms caused by many application instances connecting directly to PostgreSQL.

However, application startup should still use controlled connection creation.

Avoid synchronized connection bursts where possible.

---

# 34. Service Scaling

When RideForge scales from:

```text
5 Instances
```

to:

```text
50 Instances
```

the database connection model must remain predictable.

PgBouncer provides a central connection-management boundary.

---

# 35. Autoscaling

Autoscaling can increase the number of application instances rapidly.

Without connection pooling:

```text
Instance Count ↑
      ↓
DB Connections ↑
      ↓
PostgreSQL Pressure ↑
```

With PgBouncer:

```text
Instance Count ↑
      ↓
Client Connections ↑
      ↓
Controlled Backend Pool
```

The backend pool still needs capacity planning.

---

# 36. Background Workers

Background workers can consume PostgreSQL connections just like HTTP services.

Examples include:

```text
Event Consumers
Outbox Publishers
Scheduled Jobs
Matching Workers
Payment Workers
Notification Workers
```

Their connection pools must be included in the global connection budget.

---

# 37. Redpanda Consumers

A Redpanda consumer that writes to PostgreSQL may generate bursty database traffic.

PgBouncer can help absorb connection demand, but the application should also control:

```text
Consumer Concurrency
Batch Size
Transaction Size
Database Write Rate
```

---

# 38. Outbox Publisher

The Outbox publisher may read PostgreSQL records frequently.

It should use:

```text
Short Transactions
Bounded Connections
Efficient Queries
Appropriate Indexes
```

PgBouncer should not be used to compensate for inefficient outbox polling.

---

# 39. Matching Workers

Matching workers can be highly concurrent.

The architecture should prevent:

```text
Worker Count ↑
      ↓
Unbounded DB Connections
```

Use bounded worker concurrency and appropriate pooling.

---

# 40. Database Connection Leaks

Connection leaks are especially dangerous when PgBouncer is present because the application may continue holding client-side connections even if backend connections are managed centrally.

Monitor:

```text
Application Pool Usage
PgBouncer Client Connections
PgBouncer Waiting Clients
PostgreSQL Backend Connections
```

---

# 41. Connection Metrics

Important metrics include:

```text
Application Pool Size
Application Pool In Use
PgBouncer Client Connections
PgBouncer Waiting Clients
PgBouncer Active Servers
PgBouncer Idle Servers
PostgreSQL Active Connections
PostgreSQL Idle Connections
Connection Wait Time
```

---

# 42. PgBouncer Observability

PgBouncer exposes administrative statistics that can be used to monitor:

```text
Pools
Clients
Servers
Waiting Clients
Traffic
Transaction Rate
Query Rate
Latency
```

Monitoring should be integrated into RideForge observability.

---

# 43. Queue Time vs Query Time

A critical diagnostic distinction is:

```text
Connection Wait Time
```

versus:

```text
Database Query Time
```

A request may be slow because:

```text
No Connection Available
```

or because:

```text
The Query Is Slow
```

These require different fixes.

---

# 44. Pool Saturation

Pool saturation should be monitored.

If:

```text
Waiting Clients ↑
```

while:

```text
Backend Connections = Capacity
```

the system may need:

```text
Query Optimization
Higher Appropriate Pool Capacity
Reduced Concurrency
Database Scaling
```

depending on the actual bottleneck.

---

# 45. PostgreSQL Saturation

If PostgreSQL is already CPU- or I/O-bound, increasing PgBouncer backend connections may worsen performance.

The decision must therefore consider:

```text
Database CPU
Database Memory
I/O
Locks
Query Latency
Connections
```

together.

---

# 46. PgBouncer Failure

PgBouncer is an infrastructure dependency.

If PgBouncer becomes unavailable:

```text
Application
    ↓
Database Connection Failure
```

unless a controlled fallback architecture exists.

The production deployment should therefore provide appropriate:

```text
Availability
Restart
Health Checks
Monitoring
```

---

# 47. PgBouncer High Availability

For production, PgBouncer availability should match the criticality of PostgreSQL access.

Possible approaches include:

```text
Multiple PgBouncer Instances
Load Balancer
Managed Database Proxy
Orchestrator-managed Instances
```

The exact deployment depends on the hosting environment.

---

# 48. PgBouncer Statelessness

PgBouncer should be treated primarily as connection-management infrastructure.

The application should not depend on durable state being stored inside PgBouncer.

This simplifies:

```text
Restart
Failover
Scaling
Deployment
```

---

# 49. Health Checks

PgBouncer deployment should expose health checks that verify:

```text
Process Health
Port Availability
PostgreSQL Connectivity
Pool Availability
```

where appropriate.

---

# 50. Deployment Ordering

A safe deployment sequence is generally:

```text
PostgreSQL
      ↓
PgBouncer
      ↓
Application Services
```

During shutdown:

```text
Application Services
      ↓
PgBouncer
      ↓
PostgreSQL
```

Graceful termination should allow in-flight database operations to complete or fail predictably.

---

# 51. Configuration

PgBouncer configuration should be environment-specific.

Do not hard-code production connection parameters into source code.

Configuration should be managed through:

```text
Environment Configuration
Secret Management
Deployment Configuration
```

---

# 52. Credentials

Database credentials used by PgBouncer must be protected.

Credentials must not be:

```text
Committed to Git
Logged
Embedded in Images
Hard-Coded in Source
```

---

# 53. Authentication

PgBouncer authentication must be configured consistently with the PostgreSQL authentication model.

The chosen mechanism should support:

```text
Security
Operational Simplicity
Credential Rotation
Service Authentication
```

---

# 54. TLS

Production database traffic should use appropriate transport security where required.

The architecture should consider encryption for:

```text
Application → PgBouncer
PgBouncer → PostgreSQL
```

depending on deployment topology and trust boundaries.

---

# 55. Network Isolation

PgBouncer and PostgreSQL should not be unnecessarily exposed publicly.

Preferred architecture:

```text
Application Network
        ↓
Private PgBouncer
        ↓
Private PostgreSQL
```

---

# 56. Migration Connections

Schema migrations may have different connection requirements from normal application traffic.

Migration jobs should be treated as a separate workload.

They may use:

```text
Direct PostgreSQL Connection
```

or:

```text
PgBouncer
```

depending on migration compatibility and deployment design.

---

# 57. Administrative Connections

Operational database access should not depend on the same pool capacity required by production application traffic.

Maintain appropriate administrative capacity.

---

# 58. Long-Running Queries

Long-running queries occupy backend connections.

They reduce the effectiveness of pooling.

Long-running queries should be:

```text
Identified
Measured
Optimized
Moved to Appropriate Workloads
```

where possible.

---

# 59. Transactions and Pool Reuse

Transaction pooling works best when transactions are:

```text
Short
Self-Contained
Stateless
```

This is aligned with RideForge's service-oriented architecture.

---

# 60. Advisory Locks

Advisory locks require special consideration with transaction pooling.

Application logic must understand whether a lock is:

```text
Transaction-Scoped
```

or:

```text
Session-Scoped
```

Session-scoped assumptions may be incompatible with transaction pooling.

Critical locking designs must be tested explicitly.

---

# 61. LISTEN / NOTIFY

Persistent PostgreSQL session features such as:

```text
LISTEN
NOTIFY
```

require careful consideration with transaction pooling.

RideForge's primary event architecture is Redpanda, so PostgreSQL LISTEN/NOTIFY should not become an implicit event-bus dependency.

---

# 62. Temporary Tables

Temporary tables are session-scoped PostgreSQL state.

Applications using transaction pooling should avoid relying on temporary tables across transactions unless compatibility is explicitly validated.

---

# 63. Session Variables

Session variables may persist differently depending on pooling mode.

Avoid relying on persistent session state unless the connection lifecycle guarantees it.

---

# 64. Prepared Statement Strategy

The Go data-access layer should define an explicit prepared-statement strategy compatible with the selected PgBouncer mode.

This should be validated with:

```text
Integration Tests
Load Tests
Production-Like Configuration
```

---

# 65. Application Timeout Strategy

Database operations should respect request deadlines.

Conceptually:

```text
HTTP / Event Deadline
        ↓
Database Context
        ↓
PgBouncer Wait
        ↓
PostgreSQL Query
```

When the request deadline expires, the database operation should be cancelled where supported.

---

# 66. Query Cancellation

Query cancellation helps prevent abandoned requests from continuing to consume PostgreSQL resources.

This is particularly important for:

```text
High-Concurrency APIs
Long Queries
Client Disconnects
Timeouts
```

---

# 67. Connection Reuse

PgBouncer should maximize safe reuse of backend PostgreSQL connections.

However, connections must not be reused across incompatible session states.

This is another reason to minimize session-specific application behaviour.

---

# 68. Pooling and Transactions

A transaction must remain logically complete before a backend connection is returned to the pool.

Applications must always:

```text
COMMIT
```

or:

```text
ROLLBACK
```

before ending the transaction.

Unfinished transactions can cause connection and state-management problems.

---

# 69. Error Handling

Applications should handle:

```text
Connection Refused
Pool Timeout
Query Timeout
Database Unavailable
PgBouncer Unavailable
Connection Reset
```

without blindly retrying indefinitely.

---

# 70. Retry Strategy

Database retries must be controlled.

Avoid:

```text
Immediate Infinite Retry
```

Prefer:

```text
Bounded Retry
Exponential Backoff
Jitter
Idempotency Where Required
```

according to the operation.

---

# 71. Retry and Transactions

A failed transaction may be retried only when the application can safely determine that retrying will not duplicate a business operation.

This is especially important for:

```text
Payments
Ride Creation
Driver Assignment
State Transitions
```

---

# 72. PgBouncer and Idempotency

PgBouncer does not provide business-level idempotency.

Application-level idempotency remains governed by the platform's idempotency strategy.

---

# 73. PgBouncer and Outbox

The Outbox pattern remains responsible for:

```text
Transactional State + Event Record
```

PgBouncer only manages the PostgreSQL connection layer.

The two concerns must not be conflated.

---

# 74. PgBouncer and PostgreSQL Transactions

PgBouncer does not change PostgreSQL transaction semantics.

The application must still define:

```text
Transaction Boundary
Isolation
Commit
Rollback
Consistency
```

correctly.

---

# 75. PgBouncer and Redis

PgBouncer has no role in Redis connection management.

The two systems have independent pools and capacity limits:

```text
PostgreSQL
→ PgBouncer

Redis
→ Redis Client Pool
```

---

# 76. PgBouncer and Redpanda

PgBouncer does not manage Redpanda connections.

Redpanda consumers and producers have their own:

```text
Connection
Concurrency
Batching
Backpressure
```

configuration.

---

# 77. PgBouncer and AI Services

AI services that query PostgreSQL must follow the same connection-management policy.

Large AI workloads should not bypass PgBouncer without a documented reason.

Offline model-training workloads may require a separate database-access strategy.

---

# 78. Analytics Workloads

Large analytical queries should not be routed through the same production pool without careful consideration.

Potential alternatives include:

```text
Read Replica
Analytics Database
Batch Pipeline
Dedicated Connection Pool
```

where required.

---

# 79. Read Replicas

If PostgreSQL read replicas are introduced, connection pooling must distinguish:

```text
Primary Connections
Read Replica Connections
```

according to consistency requirements.

This should be introduced through a future architecture decision if necessary.

---

# 80. Connection Routing

RideForge should not blindly route every database operation through the same pool when different consistency requirements exist.

Potential future pools include:

```text
Primary Pool
Read Pool
Analytics Pool
Migration / Admin Pool
```

The initial deployment should remain simple unless workload evidence requires separation.

---

# 81. Local Development

Local development may use:

```text
Application
   ↓
PgBouncer
   ↓
PostgreSQL
```

when testing production-like connection behaviour.

However, developers should also be able to run PostgreSQL directly when debugging database-specific issues.

---

# 82. Local Development Simplicity

PgBouncer should not become a mandatory local dependency unless the development workflow requires it.

A reasonable approach is:

```text
Basic Development
→ Application → PostgreSQL

Production-Like Testing
→ Application → PgBouncer → PostgreSQL
```

---

# 83. Integration Testing

Integration tests should test both:

```text
Direct PostgreSQL
```

and, where relevant:

```text
PostgreSQL Through PgBouncer
```

especially for:

```text
Transactions
Prepared Statements
Session State
Connection Reuse
Timeouts
```

---

# 84. Load Testing

Load tests should compare:

```text
Direct PostgreSQL
```

against:

```text
PgBouncer + PostgreSQL
```

where necessary.

Measure:

```text
Request Latency
Connection Wait Time
Database Connections
Throughput
CPU
Error Rate
```

---

# 85. Failure Testing

Test:

```text
PostgreSQL Restart
PgBouncer Restart
Application Restart
Connection Saturation
Pool Saturation
Database Failure
Network Interruption
```

The system should recover without uncontrolled connection storms.

---

# 86. Monitoring

Important PgBouncer metrics include:

```text
Client Connections
Waiting Clients
Active Server Connections
Idle Server Connections
Pool Saturation
Transaction Rate
Query Rate
Bytes Sent
Bytes Received
```

---

# 87. PostgreSQL Correlation

PgBouncer metrics should be correlated with PostgreSQL metrics:

```text
PgBouncer Waiting Clients
+
PostgreSQL CPU
+
Query Latency
+
PostgreSQL Connections
```

This helps determine the actual bottleneck.

---

# 88. Alerts

Potential alerts include:

```text
High Waiting Clients
High Connection Wait Time
Backend Pool Exhaustion
PostgreSQL Connection Saturation
PgBouncer Unavailable
Repeated Connection Errors
Unusual Connection Growth
```

---

# 89. Capacity Planning

Capacity planning should consider:

```text
Application Instances
Connections Per Instance
Worker Count
Request Rate
Query Duration
Peak Concurrency
Database CPU
Database Memory
```

Connection pool configuration should be reviewed whenever service topology changes materially.

---

# 90. Scaling Formula

A conceptual planning model is:

```text
Total Application Connections
=
Σ Connection Pool Size Per Instance
```

The maximum PostgreSQL backend pool should then be constrained by:

```text
PostgreSQL Capacity
-
Operational Headroom
```

This is a planning model, not a universal formula.

Actual limits must be validated through load testing.

---

# 91. Avoid Oversized Pools

A common mistake is:

```text
More Connections
=
More Performance
```

This is not generally true.

Too many active queries can increase:

```text
CPU Contention
Memory Pressure
Lock Contention
Context Switching
Disk I/O
Latency
```

---

# 92. Pool Sizing Strategy

The preferred approach is:

```text
Start Conservatively
      ↓
Load Test
      ↓
Measure
      ↓
Tune
      ↓
Monitor Production
      ↓
Adjust
```

---

# 93. Operational Runbook

If PgBouncer shows high waiting clients:

```text
1. Check PostgreSQL CPU
2. Check Query Latency
3. Check Active Backend Connections
4. Check Long Transactions
5. Check Lock Contention
6. Check Application Pool Usage
7. Check Request Rate
8. Determine Actual Bottleneck
```

Do not immediately increase pool size.

---

# 94. If PostgreSQL Is Healthy but PgBouncer Is Saturated

Investigate:

```text
PgBouncer Pool Configuration
Client Connection Count
Backend Pool Size
Connection Queueing
```

Then adjust capacity if justified.

---

# 95. If PostgreSQL Is Saturated

Do not blindly increase PgBouncer connections.

Instead investigate:

```text
Slow Queries
CPU
I/O
Locks
Indexes
Transactions
Read / Write Load
```

Potential solutions may include:

```text
Query Optimization
Caching
Read Replicas
Partitioning
Database Scaling
Workload Separation
```

---

# 96. If Connection Count Is High but Query Load Is Low

Investigate:

```text
Idle Connections
Application Pool Size
PgBouncer Pool Configuration
Connection Leaks
Long-Lived Sessions
```

The solution may be reducing pool sizes rather than increasing them.

---

# 97. Cost Considerations

PgBouncer itself is lightweight compared with database infrastructure.

Its primary value is:

```text
Better Connection Utilization
Reduced Connection Overhead
Controlled Database Capacity
```

It should be introduced because it solves a real connection-management problem, not simply because it is a common production component.

---

# 98. Alternatives Considered

## 98.1 Direct Application-to-PostgreSQL Connections

### Advantages

```text
Simple
Fewer Components
Easy Debugging
```

### Disadvantages

```text
Connection Multiplication Across Instances
Higher Connection Pressure
Less Centralized Pool Management
```

### Decision

Supported for small/local environments, but PgBouncer is preferred when production connection pressure justifies it.

---

# 99. Managed Database Proxy

A managed provider proxy may provide similar functionality.

### Advantages

```text
Managed Operations
High Availability
Provider Integration
```

### Disadvantages

```text
Provider Coupling
Potential Additional Cost
Different Feature Set
```

### Decision

A managed proxy remains a valid deployment option, but the architectural requirement is connection pooling rather than a specific vendor implementation.

---

# 100. Application Pooling Only

Application-level pooling is necessary but may become inefficient across many service instances.

For example:

```text
20 Instances
×
20 Connections
=
400 Potential Backend Connections
```

PgBouncer provides an additional multiplexing layer.

---

# 101. Pgpool-II

Pgpool-II is another PostgreSQL middleware option.

It provides broader database-proxy functionality.

RideForge does not require those additional capabilities merely for connection pooling.

Therefore PgBouncer is preferred for the focused pooling requirement.

---

# 102. What This Decision Does Not Mean

This ADR does not mean:

```text
Every Environment Must Always Run PgBouncer
```

It does not mean:

```text
PgBouncer Fixes Slow Queries
```

It does not mean:

```text
More Database Connections Improve Performance
```

It does not mean:

```text
Application Pooling Is No Longer Required
```

It does not mean:

```text
Transactions Can Be Held Open Indefinitely
```

---

# 103. Decision Matrix

| Requirement | Direct PostgreSQL | Application Pool | PgBouncer | Managed Proxy |
|---|---:|---:|---:|---:|
| Simple local development | **Primary** | Optional | Optional | No |
| Connection reuse | Limited | **Primary** | **Primary** | **Primary** |
| Multiplex many clients | No | Limited | **Primary** | **Primary** |
| Protect PostgreSQL connections | Limited | Some | **Strong** | **Strong** |
| Low operational complexity | **Strong** | **Strong** | Moderate | Strong |
| Vendor independence | **Strong** | **Strong** | **Strong** | Depends |
| Production scale | Conditional | **Required** | **Preferred when justified** | Alternative |
| Transaction pooling | No | No | **Yes** | Depends |

---

# 104. Consequences

## 104.1 Positive Consequences

The decision provides:

```text
Controlled PostgreSQL Connections
Connection Multiplexing
Better Scaling Across Service Instances
Reduced Connection Storm Risk
Centralized Pool Visibility
Improved Resource Utilization
```

---

## 104.2 Negative Consequences

The architecture introduces:

```text
Additional Infrastructure
Additional Configuration
Another Failure Point
Pooling Compatibility Constraints
Operational Monitoring Requirements
```

These trade-offs are accepted when connection pressure justifies the component.

---

# 105. Risks

## Risk 1 — Incorrect Pool Sizing

### Mitigation

Use:

```text
Load Testing
Connection Metrics
Database Metrics
Conservative Defaults
```

---

## Risk 2 — Pooling Mode Incompatibility

### Mitigation

Test:

```text
Prepared Statements
Transactions
Session State
Temporary Tables
Advisory Locks
```

before enabling transaction pooling.

---

## Risk 3 — PgBouncer Becomes a Bottleneck

### Mitigation

Monitor:

```text
Client Connections
Waiting Clients
Backend Connections
Latency
CPU
```

and scale PgBouncer where required.

---

## Risk 4 — PostgreSQL Is Already Saturated

### Mitigation

Do not increase backend connections blindly.

Investigate actual database bottlenecks first.

---

## Risk 5 — Connection Leaks

### Mitigation

Use:

```text
Context-Aware DB Access
Connection Pool Metrics
Integration Tests
Code Review
Timeouts
```

---

## Risk 6 — Long Transactions

### Mitigation

Monitor transaction duration and prohibit external network calls inside database transactions where possible.

---

# 106. Validation

The PgBouncer decision should be validated through:

```text
Integration Tests
Connection Pool Tests
Transaction Tests
Prepared Statement Tests
Load Tests
Failure Tests
Restart Tests
Connection Saturation Tests
Production Monitoring
```

---

# 107. Review Triggers

Revisit this ADR when:

```text
PostgreSQL Connection Pressure Changes
Service Instance Count Changes Significantly
PgBouncer Becomes a Performance Bottleneck
A Managed Database Proxy Is Introduced
Read Replicas Are Added
Database Topology Changes
Pooling Compatibility Requirements Change
```

---

# 108. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
```

Especially:

```text
Database Development
Configuration and Environment
Performance and Optimization
Error Handling and Validation
Logging and Debugging
Observability Development
Integration Testing and Local Infrastructure
Migrations and Schema Changes
```

---

# 109. Related ADRs

This decision is directly related to:

```text
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0007 — PostgreSQL as Primary Database
ADR-0008 — Redis for Real-Time State and Caching
ADR-0009 — PostGIS for Geospatial Operations
ADR-0010 — Driver Location Storage Strategy
ADR-0012 — Outbox Pattern
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
ADR-0024 — Configuration and Environment Strategy
ADR-0025 — Testing and Integration Strategy
ADR-0028 — Cost Optimization Strategy
```

---

# 110. Decision Summary

RideForge adopts:

```text
PgBouncer
```

as the preferred external PostgreSQL connection-pooling layer when production connection pressure, service scaling, or deployment topology justifies it.

The intended architecture is:

```text
RideForge Services
        ↓
Application Connection Pool
        ↓
PgBouncer
        ↓
PostgreSQL
```

with:

```text
Bounded Application Pools
+
Controlled PgBouncer Backend Pools
+
Short Transactions
+
Query Optimization
+
Connection Monitoring
```

The system should remain simple in environments where PgBouncer is unnecessary.

---

# 111. Final Principle

> **PgBouncer protects PostgreSQL from uncontrolled connection growth by multiplexing application clients onto a bounded backend connection pool; it complements application pooling and database optimization rather than replacing them.**

The intended production relationship is:

```text
                 RideForge Services
                        │
                Application Pools
                        │
                        ▼
                 ┌─────────────┐
                 │  PgBouncer  │
                 └──────┬──────┘
                        │
                Bounded Backend
                  Connections
                        │
                        ▼
                 ┌─────────────┐
                 │ PostgreSQL  │
                 └─────────────┘
```

This establishes a controlled PostgreSQL connection boundary while preserving application-level transaction management and PostgreSQL as the authoritative transactional database.

---

# 112. Status

```text
Decision: ACCEPTED

Connection Pooling:
Application Pool + PgBouncer Where Justified

Primary Database:
PostgreSQL

Preferred Pooling Mode:
Transaction Pooling Where Application Compatibility Allows
```

This decision establishes the RideForge PostgreSQL connection-pooling strategy and provides the foundation for controlled database concurrency as the number of services, workers, and application instances increases.
