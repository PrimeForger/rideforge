# ADR-0007: PostgreSQL as Primary Database

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Data / Infrastructure  
> **Scope:** RideForge primary transactional data storage  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge requires a durable primary database for transactional and relational business data.

The platform contains domain areas such as:

```text
Users
Drivers
Rides
Dispatch
Matching
Payments
Regions
Operational Configuration
```

These domains require:

```text
Transactional Integrity
Relational Modeling
Strong Consistency Where Required
Constraints
Indexes
Reliable Queries
Migrations
Backup and Recovery
Operational Maturity
```

RideForge also has specialized workloads that may require different infrastructure:

```text
Redis
→ Low-latency / ephemeral workloads

Redpanda
→ Event streaming

Geospatial Infrastructure
→ Location-heavy workloads

AI / ML Infrastructure
→ Prediction and model workloads
```

The primary transactional database therefore needs to remain focused on durable business state rather than becoming the storage solution for every workload.

---

# 2. Problem

RideForge needs a primary database that can support:

```text
Transactional Ride State
Driver State
User Data
Payment State
Business Rules
Relational Queries
Data Integrity
Concurrent Operations
```

The database must also work effectively with the project's existing technology direction:

```text
Go
PostgreSQL Drivers
PostGIS Where Required
Redis
Redpanda
Docker
Cloud Infrastructure
```

The platform is also being designed with cost efficiency in mind.

---

# 3. Decision

RideForge will use:

> **PostgreSQL as the primary transactional database.**

PostgreSQL will be the authoritative durable store for core relational business state unless a separate architectural decision explicitly assigns a workload to another storage system.

---

# 4. Primary Database Principle

The primary database should be:

```text
PostgreSQL
```

for durable transactional domain state.

Other infrastructure should not become an accidental replacement for PostgreSQL.

Conceptually:

```text
Core Business State
        ↓
    PostgreSQL
```

while:

```text
Cache / Ephemeral State
        ↓
      Redis

Event Stream
        ↓
    Redpanda
```

---

# 5. Why PostgreSQL

PostgreSQL provides capabilities that align well with RideForge's transactional requirements:

```text
ACID Transactions
Relational Modeling
Foreign Keys
Constraints
Indexes
Advanced SQL
JSON Support
Geospatial Extensions
Mature Tooling
Reliable Backup / Recovery Ecosystem
```

It also provides a strong foundation for evolving RideForge without prematurely introducing multiple specialized databases.

---

# 6. Transactional Authority

PostgreSQL will be authoritative for durable transactional state owned by the application domains.

Examples include:

```text
Ride Lifecycle State
Driver Business State
User Business State
Payment State
Configuration State
```

The exact ownership of each table must remain aligned with the domain model.

---

# 7. Database Ownership

Database ownership should follow domain ownership.

Conceptually:

```text
Domain
   ↓
Application Service
   ↓
Repository / Data Access
   ↓
PostgreSQL
```

A service or module should not modify another domain's data merely because the tables are physically accessible.

---

# 8. Shared PostgreSQL Deployment

RideForge may use one PostgreSQL deployment for multiple domains during early and intermediate stages.

This does not mean that all domains share unrestricted ownership.

The architecture distinguishes between:

```text
Physical Database
```

and:

```text
Logical Data Ownership
```

---

# 9. Logical Ownership

Even when domains share the same PostgreSQL instance:

```text
Ride
Driver
Payment
Dispatch
```

should maintain clear ownership boundaries.

The preferred rule is:

> **Physical database sharing is acceptable; uncontrolled logical ownership is not.**

---

# 10. Direct Cross-Domain Writes

Direct writes into another domain's tables should be avoided.

Prefer:

```text
Domain A
   ↓
API / Application Contract / Event
   ↓
Domain B
   ↓
Domain B Repository
   ↓
PostgreSQL
```

rather than:

```text
Domain A
   ↓
Domain B Tables
```

---

# 11. ACID Transactions

PostgreSQL transactions should be used where multiple changes must remain consistent within the local transaction boundary.

Examples may include:

```text
Ride State Change
Assignment State Change
Payment State Change
Outbox Record Creation
```

The exact transaction boundary must follow domain consistency requirements.

---

# 12. Transaction Boundary Principle

A transaction should protect a clearly defined consistency boundary.

Prefer:

```text
Small, Meaningful Transaction
```

over:

```text
Large Transaction Across Unrelated Domains
```

---

# 13. Cross-Service Transactions

RideForge will not rely on distributed database transactions as the default mechanism for coordinating independent services.

For cross-boundary workflows, use appropriate mechanisms such as:

```text
Events
Outbox
Idempotency
Compensation
Retry
```

according to the workflow.

---

# 14. PostgreSQL and Outbox

When a business transaction must reliably produce an event, PostgreSQL may store the corresponding Outbox record in the same transaction.

Conceptually:

```text
BEGIN
  │
  ├── Update Domain State
  │
  └── Insert Outbox Record
  │
COMMIT
```

Then:

```text
Outbox Publisher
       ↓
Redpanda
```

This connects PostgreSQL transactional state with reliable event publication.

---

# 15. Schema Design

PostgreSQL schemas should represent durable business state clearly.

Schema design should prioritize:

```text
Correctness
Integrity
Queryability
Maintainability
Operational Performance
```

Avoid designing tables solely around API response shapes.

---

# 16. Relational Modeling

Use relational modeling when relationships are meaningful.

Examples include:

```text
Ride → Driver
Ride → User
Ride → Region
Payment → Ride
```

Foreign keys and constraints should be used where they protect important integrity requirements.

---

# 17. Database Constraints

Important invariants should be protected at the database level where appropriate.

Examples include:

```text
NOT NULL
UNIQUE
FOREIGN KEY
CHECK
PRIMARY KEY
```

Database constraints complement domain validation.

They do not replace domain rules.

---

# 18. Domain Validation vs Database Constraints

Use:

```text
Domain Validation
```

for business rules.

Use:

```text
Database Constraints
```

for data integrity.

For important invariants, both may be appropriate.

---

# 19. Primary Keys

RideForge should use stable identifiers for durable entities.

Identifiers should support:

```text
Uniqueness
Distributed Processing
Event Correlation
Database Relationships
```

The exact identifier format should remain consistent with the application's established conventions.

---

# 20. Indexing

Indexes should be created based on actual access patterns.

Important workloads may include:

```text
Ride Lookup
Driver Lookup
Assignment Lookup
Status Filtering
Time-Based Queries
Foreign-Key Queries
Geospatial Queries
```

Avoid adding indexes without a query or constraint justification.

---

# 21. Index Trade-Off

Indexes improve reads but add cost to:

```text
Writes
Storage
Maintenance
Vacuum / Cleanup
```

Therefore every important index should have a clear purpose.

---

# 22. Query Design

Queries should be designed around known application access patterns.

Avoid:

```text
SELECT *
```

when only a subset of fields is required.

Prefer:

```text
Explicit Columns
Bounded Results
Appropriate Indexes
```

where practical.

---

# 23. Pagination

Large result sets should use controlled pagination.

For high-volume datasets, prefer pagination strategies appropriate to the access pattern rather than repeatedly scanning large offsets.

---

# 24. N+1 Query Prevention

Application code should avoid repeatedly querying PostgreSQL for related records when a suitable query or batching strategy can retrieve the required data efficiently.

Monitor for:

```text
N+1 Queries
Repeated Lookups
Unbounded Queries
```

---

# 25. Connection Management

PostgreSQL connections are a limited resource.

Applications must use:

```text
Connection Pooling
Bounded Pool Sizes
Timeouts
Proper Connection Release
```

rather than creating unrestricted connections.

---

# 26. Connection Pooling

RideForge should use an appropriate PostgreSQL connection pool.

Pool sizing should account for:

```text
Application Concurrency
Database Capacity
Number of Application Instances
PgBouncer Where Used
Query Duration
```

---

# 27. PgBouncer

PgBouncer may be used where connection management requires an additional pooling layer.

Conceptually:

```text
Application
     ↓
PgBouncer
     ↓
PostgreSQL
```

PgBouncer should be introduced based on measured connection pressure rather than automatically added to every environment.

---

# 28. Connection Limits

PostgreSQL connection limits must be treated as a capacity constraint.

Increasing application concurrency without considering database connections can result in:

```text
Connection Exhaustion
Queueing
Timeouts
Latency Spikes
Database Instability
```

---

# 29. Query Timeouts

Long-running queries should be controlled through appropriate:

```text
Application Timeouts
Database Timeouts
Cancellation
```

where appropriate.

A slow query should not indefinitely consume a connection.

---

# 30. Transaction Timeouts

Transactions should remain as short as practical.

Avoid holding transactions open while waiting for:

```text
External APIs
Network Calls
User Input
Long Computation
Unbounded Messaging Operations
```

---

# 31. Locking

PostgreSQL locking should be used deliberately.

Application workflows should avoid unnecessary lock contention.

For concurrent operations, evaluate:

```text
Row Locks
Optimistic Concurrency
Atomic Updates
Unique Constraints
Transaction Isolation
```

---

# 32. Concurrency

RideForge has concurrency-sensitive workflows such as:

```text
Driver Assignment
Ride State Changes
Payment State Changes
```

Database operations must protect against race conditions.

---

# 33. Atomic State Changes

Where possible, use database operations that make state transitions atomic.

Conceptually:

```text
UPDATE ...
WHERE current_state = expected_state
```

can prevent a stale caller from applying an invalid concurrent transition.

---

# 34. Optimistic Concurrency

Optimistic concurrency may be used when multiple actors can update the same state.

Potential mechanisms include:

```text
Version Column
Updated Timestamp
Expected State
Conditional Update
```

The correct mechanism depends on the domain.

---

# 35. Pessimistic Locking

Pessimistic locking may be appropriate when a short critical section must prevent competing operations from acting simultaneously.

Use it carefully because excessive locking can reduce throughput.

---

# 36. Isolation Levels

PostgreSQL transaction isolation should be selected according to business requirements.

Do not automatically use the strongest isolation level everywhere.

Consider:

```text
Correctness
Contention
Throughput
Latency
```

---

# 37. Ride Assignment

Ride assignment is a concurrency-sensitive workflow.

The implementation must prevent multiple competing workers from incorrectly assigning the same driver or ride.

Database techniques may be combined with:

```text
Atomic State Changes
Locks
Constraints
Idempotency
```

as appropriate.

---

# 38. Driver State

Driver state may change frequently.

PostgreSQL should remain the source of truth for durable driver business state, while high-frequency operational state may use specialized infrastructure when appropriate.

---

# 39. Location Data

High-frequency driver location updates should not automatically be written to PostgreSQL at an unrestricted rate.

Location workloads may require:

```text
Redis
Geospatial Store
Specialized Location Infrastructure
```

depending on scale.

PostgreSQL remains the durable relational database, not necessarily the optimal destination for every real-time location update.

---

# 40. PostGIS

PostGIS may be used where PostgreSQL-backed geospatial capabilities provide clear value.

Potential uses include:

```text
Geospatial Queries
Region Boundaries
Spatial Filtering
Distance-Based Queries
Geographic Data
```

PostGIS should be adopted based on workload requirements.

---

# 41. Redis vs PostgreSQL

RideForge should distinguish:

```text
Durable Business State
```

from:

```text
Fast / Ephemeral Operational State
```

A conceptual model is:

```text
PostgreSQL
→ Durable Source of Truth

Redis
→ Cache / Ephemeral / Low-Latency State

Redpanda
→ Event Stream
```

---

# 42. Redpanda vs PostgreSQL

Redpanda stores event streams.

PostgreSQL stores transactional domain state.

They serve different purposes.

```text
PostgreSQL
→ What the system currently knows as durable state

Redpanda
→ What happened / what should be processed asynchronously
```

Neither should silently become a replacement for the other.

---

# 43. Cache Strategy

Redis may cache PostgreSQL-derived information.

The cache should not automatically become the authoritative source.

Conceptually:

```text
PostgreSQL
     ↓
Redis Cache
```

rather than:

```text
Redis
     ↓
PostgreSQL
```

for ordinary durable state.

---

# 44. Cache Invalidation

Cache invalidation should be explicitly designed.

Possible strategies include:

```text
TTL
Event-Based Invalidation
Write-Through
Cache-Aside
```

The selected strategy should depend on the workload.

---

# 45. JSONB

PostgreSQL JSONB may be used when semi-structured data is genuinely useful.

It should not be used as a substitute for relational modeling of stable business concepts.

Prefer relational columns for:

```text
Frequently Queried Fields
Strongly Constrained Fields
Important Relationships
```

---

# 46. Database Normalization

Normalize data where it improves:

```text
Integrity
Consistency
Maintainability
```

Denormalization may be introduced when measurable read performance or specialized access patterns justify it.

---

# 47. Denormalization

Denormalization should be:

```text
Intentional
Documented
Measured
Tested
```

Do not duplicate data merely to avoid writing a join without evidence that it is necessary.

---

# 48. Read Models

Dedicated read models may be used when:

```text
Read Workload Differs Significantly
Complex Queries Are Repeated
A Projection Improves Performance
```

The read model should have clear ownership and refresh semantics.

---

# 49. Materialized Views

Materialized views may be considered for suitable analytical or read-heavy workloads.

They should not be used for rapidly changing transactional state unless refresh requirements are well understood.

---

# 50. Analytics

Heavy analytical workloads should not unnecessarily compete with critical transactional workloads.

Depending on scale, analytics may use:

```text
Read Replicas
Dedicated Analytics Storage
Event-Based Pipelines
```

rather than executing large analytical queries against the primary transactional database.

---

# 51. Read Replicas

Read replicas may be introduced when read scaling requirements justify them.

Before introducing replicas, evaluate:

```text
Read Load
Replication Lag
Consistency Requirements
Operational Complexity
```

---

# 52. Replica Consistency

A read replica may lag behind the primary.

Do not use a replica for a read that requires immediate read-after-write consistency unless the architecture explicitly handles that requirement.

---

# 53. Database Availability

Production PostgreSQL deployment should provide availability appropriate to the platform's operational requirements.

The exact infrastructure strategy may include:

```text
Managed PostgreSQL
High Availability
Replication
Automated Recovery
```

as deployment decisions evolve.

---

# 54. Backups

PostgreSQL data must have a defined backup strategy.

Backups should support:

```text
Accidental Data Recovery
Operational Recovery
Disaster Recovery
```

---

# 55. Backup Validation

A backup strategy is incomplete without recovery testing.

Periodically verify:

```text
Backup Exists
Backup Is Usable
Restore Works
Recovery Time Is Acceptable
Recovery Point Is Acceptable
```

---

# 56. Point-in-Time Recovery

Point-in-time recovery should be considered for production environments where recovery from accidental or corrupted changes is important.

The exact implementation depends on the chosen PostgreSQL hosting model.

---

# 57. Database Security

Production PostgreSQL access should follow least privilege.

Applications should receive only the permissions required for their responsibilities.

Avoid using unrestricted administrative credentials from application processes.

---

# 58. Credentials

Database credentials must not be committed to source control.

Use appropriate:

```text
Environment Configuration
Secret Management
Credential Rotation
```

---

# 59. Encryption

Database traffic should use appropriate transport security in production environments.

Storage encryption should be enabled according to the selected infrastructure provider and security requirements.

---

# 60. Sensitive Data

Sensitive data stored in PostgreSQL should be:

```text
Minimized
Protected
Access Controlled
Audited Where Required
```

Do not store unnecessary sensitive information.

---

# 61. Migration Strategy

PostgreSQL schema changes must be managed through version-controlled migrations.

Migrations should be:

```text
Deterministic
Reviewable
Repeatable
Tested
Environment-Aware
```

---

# 62. Migration Ownership

Schema changes should be reviewed according to the domain that owns the affected data.

A migration should not silently modify unrelated domain structures.

---

# 63. Backward-Compatible Migrations

Production schema changes should prefer an approach that supports safe rolling deployment where required.

A common strategy is:

```text
Expand
   ↓
Deploy Compatible Code
   ↓
Migrate Data
   ↓
Switch Usage
   ↓
Contract
```

---

# 64. Destructive Migrations

Destructive changes such as:

```text
DROP COLUMN
DROP TABLE
Irreversible Data Transformation
```

should be handled carefully.

They should not be included casually in the same deployment as application code that still depends on the old structure.

---

# 65. Migration Testing

Every significant migration should be tested against:

```text
Fresh Database
Existing Database
Representative Production-Like Data
Rollback / Recovery Procedure Where Applicable
```

---

# 66. Schema Drift

Production schema must remain aligned with version-controlled migration history.

Manual production changes should be avoided except for controlled emergency operations.

Emergency changes must be reconciled into the migration history afterward.

---

# 67. Database Development Workflow

A typical workflow is:

```text
Define Domain Change
      ↓
Design Schema Change
      ↓
Create Migration
      ↓
Update Repository / Application
      ↓
Run Tests
      ↓
Run Migration Locally
      ↓
Integration Test
      ↓
Review
      ↓
Deploy
```

---

# 68. Database Testing

Database-related tests should cover:

```text
Constraints
Queries
Transactions
Migrations
Indexes
Concurrency
Repository Behaviour
```

---

# 69. Integration Testing

Repository integration tests should use a real PostgreSQL-compatible environment when testing actual database behaviour.

Mocks should not be relied upon to validate:

```text
SQL Semantics
Constraints
Transactions
Indexes
Locking
```

---

# 70. Test Database Isolation

Tests should avoid corrupting shared development data.

Use appropriate:

```text
Test Database
Transaction Rollback
Fixtures
Containers
Database Reset
```

according to the test type.

---

# 71. Query Performance Testing

Important queries should be evaluated using PostgreSQL query-planning tools when performance matters.

Relevant analysis may include:

```text
EXPLAIN
EXPLAIN ANALYZE
Index Usage
Rows Examined
Execution Time
```

---

# 72. Slow Query Detection

Production systems should provide a mechanism for identifying slow database queries.

Slow queries should be investigated before simply increasing database resources.

---

# 73. Database Observability

Important database telemetry includes:

```text
Connections
Connection Saturation
Query Latency
Slow Queries
Locks
Deadlocks
CPU
Memory
Storage
Replication Lag
Transaction Rate
```

---

# 74. Connection Saturation

Connection saturation is especially important for RideForge because high application concurrency can overwhelm PostgreSQL even when the database has available CPU.

Monitor:

```text
Active Connections
Idle Connections
Waiting Connections
Pool Utilization
Database Connection Limits
```

---

# 75. Deadlocks

Deadlocks should be monitored and investigated.

Common causes include:

```text
Inconsistent Lock Ordering
Long Transactions
Unnecessary Locks
Concurrent Updates
```

---

# 76. Vacuum and Maintenance

PostgreSQL maintenance operations such as:

```text
VACUUM
ANALYZE
Autovacuum
```

are important for long-running production systems.

The deployment environment should provide appropriate configuration and monitoring.

---

# 77. Table Growth

Large tables should be monitored for:

```text
Row Growth
Index Growth
Query Performance
Storage Growth
Vacuum Behaviour
```

High-growth tables may eventually require partitioning or archival strategies.

---

# 78. Partitioning

PostgreSQL table partitioning may be introduced when a table reaches a scale or access pattern where partitioning provides measurable benefits.

Potential candidates could include high-volume historical data such as:

```text
Ride History
Operational Events
Time-Series-Like Data
```

Partitioning should not be introduced prematurely.

---

# 79. Archival

Historical data should have a defined lifecycle when it becomes too large or no longer belongs in the primary operational workload.

Possible approaches include:

```text
Archival Tables
Cold Storage
Analytics Storage
Retention Policies
```

---

# 80. Database Performance Principles

The preferred optimization order is:

```text
Correct Query
      ↓
Correct Index
      ↓
Correct Data Access Pattern
      ↓
Connection Management
      ↓
Caching
      ↓
Read Scaling
      ↓
Infrastructure Scaling
```

Do not immediately scale hardware before understanding the workload.

---

# 81. PostgreSQL as Source of Truth

For durable transactional domains:

```text
PostgreSQL
```

is the source of truth unless a separate ADR explicitly assigns ownership elsewhere.

---

# 82. Redis as Derived State

When Redis represents PostgreSQL-owned data, treat it as:

```text
Derived / Cached / Operational State
```

unless a separate architecture decision explicitly defines Redis as authoritative for a specialized domain.

---

# 83. Redpanda as Event History

Redpanda event streams should not automatically be treated as the authoritative current-state database.

The event stream communicates durable events according to its retention policy.

PostgreSQL remains the primary transactional store.

---

# 84. Data Consistency

RideForge should explicitly identify whether each workflow requires:

```text
Strong Consistency
Eventual Consistency
Read-After-Write Consistency
```

The database choice alone does not determine the consistency model of the entire distributed workflow.

---

# 85. Database and Event Consistency

The combination:

```text
PostgreSQL
+
Outbox
+
Redpanda
```

provides a foundation for reliable transactional event publication without requiring distributed transactions across the database and broker.

---

# 86. Database and Idempotency

Database constraints can support idempotency.

Examples include:

```text
UNIQUE(event_id)
UNIQUE(idempotency_key)
Conditional State Updates
```

The database can therefore be an important part of business-level duplicate protection.

---

# 87. Database and Legal Rules

Legal and regional rules remain domain logic.

PostgreSQL provides persistence and integrity; it does not determine whether a ride is legally permitted.

---

# 88. Database and AI

AI systems may read historical or operational data from PostgreSQL through controlled interfaces.

AI models should not directly modify core transactional state without going through appropriate domain/application boundaries.

---

# 89. Database and Dispatch

Dispatch may require current operational information from:

```text
PostgreSQL
Redis
Location Infrastructure
ETA Systems
```

depending on the workload.

PostgreSQL should remain authoritative for durable business state.

---

# 90. Database and Matching

Matching may query PostgreSQL for durable eligibility and business state while using specialized stores for high-frequency location information.

The matching system should avoid unnecessarily expensive database scans in latency-critical paths.

---

# 91. Database and ETA

ETA prediction may use PostgreSQL historical data for:

```text
Training
Analysis
Historical Features
```

but latency-sensitive ETA requests should not depend on large analytical queries against the primary database.

---

# 92. Database and Notifications

Notification systems may consume events and retrieve required durable information through appropriate APIs or read models.

They should not directly access unrelated domain tables merely because PostgreSQL is physically shared.

---

# 93. Database and Microservices

As RideForge evolves toward independently deployable services, PostgreSQL ownership should evolve accordingly.

A possible future model is:

```text
Service
   ↓
Owned Schema / Tables
   ↓
PostgreSQL Cluster
```

or, when justified:

```text
Service
   ↓
Independent PostgreSQL Database
```

The architecture does not require a separate database per service from day one.

---

# 94. Database Extraction Strategy

If a domain is eventually extracted into an independent service:

```text
Shared PostgreSQL
      ↓
Logical Ownership
      ↓
Data Isolation
      ↓
Migration
      ↓
Independent Database Where Justified
```

The migration should be deliberate and reversible where practical.

---

# 95. Avoid Database-Per-Service Dogma

RideForge does not adopt:

```text
One Service = One Database
```

as an unconditional rule.

Database separation should be driven by:

```text
Ownership
Scaling
Failure Isolation
Security
Operational Independence
```

---

# 96. Alternative Databases Considered

RideForge has considered several database technologies for different workloads.

Potential alternatives include:

```text
MySQL
MongoDB
ScyllaDB
Redis
```

They are not selected as the primary transactional database because PostgreSQL provides the required relational and transactional capabilities while allowing specialized infrastructure where needed.

---

# 97. MongoDB

MongoDB may be useful for certain document-oriented workloads.

However, RideForge's core transactional domain requires strong relational modelling and consistency.

Therefore PostgreSQL remains the primary database.

---

# 98. MySQL

MySQL is a mature relational database and could support many RideForge workloads.

PostgreSQL is selected because of its:

```text
Advanced SQL Capabilities
Extensibility
PostGIS Support
Strong Ecosystem
```

and alignment with the project's established architecture.

---

# 99. ScyllaDB

ScyllaDB may be useful for extremely high-scale distributed workloads such as specialized location or time-series-like data.

It is not required as the primary transactional database at the current architectural stage.

A specialized storage decision may be made later if workload evidence justifies it.

---

# 100. Redis

Redis is not selected as the primary transactional database.

It remains appropriate for:

```text
Caching
Real-Time State
Ephemeral Data
Low-Latency Access
```

according to the Redis architecture.

---

# 101. Cost Considerations

PostgreSQL provides a strong cost-to-capability balance for the current RideForge architecture.

The platform should avoid adding specialized databases until there is measurable justification.

---

# 102. Operational Simplicity

A primary PostgreSQL deployment reduces the number of database technologies that engineering teams must operate during early platform growth.

This supports:

```text
Simpler Development
Simpler Testing
Simpler Backups
Simpler Monitoring
Lower Operational Overhead
```

---

# 103. Future Scaling

PostgreSQL can scale through multiple approaches before requiring a fundamentally different primary data architecture:

```text
Query Optimization
Indexing
Connection Pooling
PgBouncer
Vertical Scaling
Read Replicas
Partitioning
Archival
Caching
Specialized Stores
```

The architecture should scale based on evidence.

---

# 104. What This Decision Does Not Mean

This ADR does not mean:

```text
Every Data Point Must Be Stored in PostgreSQL
```

It does not mean:

```text
Redis Cannot Own Specialized Operational State
```

It does not mean:

```text
Redpanda Is a Database Replacement
```

It does not mean:

```text
Every Service Must Have Its Own PostgreSQL Instance
```

It establishes PostgreSQL as the primary transactional database.

---

# 105. Decision Matrix

| Requirement | PostgreSQL | Redis | Redpanda |
|---|---:|---:|---:|
| Durable relational state | **Primary** | No | No |
| ACID transactions | **Primary** | Limited / workload-specific | No |
| Business relationships | **Primary** | No | No |
| High-frequency ephemeral state | Secondary | **Primary** | No |
| Event streaming | No | Secondary / specialized | **Primary** |
| Durable event retention | No | Possible / workload-specific | **Primary** |
| Geospatial relational queries | **Primary with PostGIS** | Specialized | No |
| Cache | No | **Primary** | No |
| Transactional outbox | **Primary** | No | Destination |
| Consumer groups | No | Possible | **Primary** |

---

# 106. Database Development Checklist

Before introducing or modifying PostgreSQL data:

```text
[ ] Domain ownership is clear
[ ] Schema design is reviewed
[ ] Constraints are defined
[ ] Indexes have a purpose
[ ] Migration is created
[ ] Migration is tested
[ ] Repository changes are tested
[ ] Transaction boundary is clear
[ ] Concurrency risks are considered
[ ] Performance impact is considered
[ ] Backup / recovery impact is considered
[ ] Security / privacy impact is reviewed
```

---

# 107. New Table Checklist

```text
[ ] Business purpose is documented
[ ] Owning domain is identified
[ ] Primary key is defined
[ ] Foreign keys are defined where appropriate
[ ] NOT NULL requirements are defined
[ ] UNIQUE constraints are considered
[ ] CHECK constraints are considered
[ ] Indexes are justified
[ ] Retention requirements are known
[ ] Migration is created
[ ] Integration tests exist
```

---

# 108. New Index Checklist

```text
[ ] Query requiring the index is identified
[ ] Existing indexes reviewed
[ ] Index selectivity considered
[ ] Write overhead considered
[ ] Storage overhead considered
[ ] Query plan validated
[ ] Production impact considered
```

---

# 109. Database Performance Checklist

```text
[ ] Query plan inspected
[ ] Appropriate index exists
[ ] Result set is bounded
[ ] N+1 behaviour checked
[ ] Connection usage checked
[ ] Transaction duration checked
[ ] Lock contention checked
[ ] Cache strategy considered
[ ] Load tested where necessary
```

---

# 110. Migration Checklist

```text
[ ] Migration is version controlled
[ ] Existing data considered
[ ] Production rollout considered
[ ] Backward compatibility considered
[ ] Lock duration considered
[ ] Large-table impact considered
[ ] Rollback / recovery strategy considered
[ ] Application compatibility verified
```

---

# 111. Consequences

## 111.1 Positive Consequences

The decision provides:

```text
Strong Transactional Foundation
Relational Data Integrity
Mature Ecosystem
PostGIS Support
Flexible Querying
Cost Efficiency
Simplified Initial Operations
Compatibility With Go
Strong Migration Tooling
```

---

## 111.2 Negative Consequences

The architecture requires careful management of:

```text
Connections
Indexes
Query Performance
Vacuum / Maintenance
Storage Growth
Scaling
High-Frequency Workloads
```

PostgreSQL should not be forced to handle workloads for which specialized infrastructure is more appropriate.

---

# 112. Risks

## Risk 1 — PostgreSQL Becomes a Bottleneck

### Mitigation

Use:

```text
Query Optimization
Indexing
Connection Pooling
PgBouncer
Caching
Read Replicas
Partitioning
```

according to measured requirements.

---

## Risk 2 — Database Connection Exhaustion

### Mitigation

Control:

```text
Pool Size
Application Concurrency
PgBouncer
Database Limits
```

and monitor connection saturation.

---

## Risk 3 — High-Frequency Location Writes

### Mitigation

Use specialized real-time infrastructure rather than blindly writing every location update to PostgreSQL.

---

## Risk 4 — Cross-Domain Database Coupling

### Mitigation

Maintain logical data ownership and avoid direct cross-domain writes.

---

## Risk 5 — Large Analytical Queries

### Mitigation

Use:

```text
Read Replicas
Event Pipelines
Analytics Storage
Read Models
```

when required.

---

## Risk 6 — Premature Database Scaling

### Mitigation

Measure the actual bottleneck before introducing additional infrastructure.

---

# 113. Validation

The PostgreSQL decision should be validated through:

```text
Integration Tests
Load Tests
Query Plan Analysis
Concurrency Tests
Migration Tests
Failure Tests
Backup / Restore Tests
Production Monitoring
```

---

# 114. Review Triggers

Revisit this ADR when:

```text
PostgreSQL No Longer Meets Workload Requirements
Database Cost Becomes Material
Major Data Domain Is Added
Specialized Storage Becomes Necessary
Scaling Characteristics Change
High-Frequency Workloads Increase Significantly
A Service Requires Independent Database Ownership
```

---

# 115. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
```

Especially:

```text
Database Development
Redis Development
Event and Messaging Development
Migrations and Schema Changes
Performance and Optimization
Integration Testing and Local Infrastructure
Observability Development
AI Machine Learning Data Pipeline
```

---

# 116. Related ADRs

This decision is directly related to:

```text
ADR-0002 — Architecture Style
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0012 — Outbox Pattern
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0028 — Cost Optimization Strategy
```

---

# 117. Decision Summary

RideForge adopts:

```text
PostgreSQL
```

as the primary transactional database.

The database will be the authoritative durable store for core relational business state while specialized infrastructure is used where appropriate:

```text
PostgreSQL
→ Transactional / Durable State

Redis
→ Cache / Real-Time / Ephemeral State

Redpanda
→ Event Streaming

PostGIS
→ PostgreSQL Geospatial Capability

Specialized Storage
→ Introduced Only When Workload Evidence Justifies It
```

---

# 118. Final Principle

> **PostgreSQL is the default source of truth for RideForge's durable transactional domain state; specialized infrastructure should be introduced only when a workload has requirements that PostgreSQL should not be forced to satisfy.**

The intended data architecture is:

```text
                    RideForge Data Layer
                           │
          ┌────────────────┼────────────────┐
          │                │                │
     PostgreSQL          Redis           Redpanda
          │                │                │
   Durable State      Fast / Ephemeral    Events
          │                │                │
          └────────────────┼────────────────┘
                           │
                  Specialized Stores
                  Where Justified
```

This provides a strong transactional foundation while preserving the ability to scale specialized workloads independently.

---

# 119. Status

```text
Decision: ACCEPTED

Primary Database:
PostgreSQL

Primary Role:
Durable Transactional Domain State
```

This decision establishes PostgreSQL as the primary database foundation for RideForge and provides the basis for subsequent decisions concerning data consistency, transaction boundaries, schema ownership, migrations, and specialized storage.
