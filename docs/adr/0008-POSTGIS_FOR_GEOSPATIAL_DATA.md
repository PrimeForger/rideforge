# ADR-0008: PostGIS for Geospatial Data

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Data / Geospatial Infrastructure  
> **Scope:** RideForge geospatial data and PostgreSQL-backed spatial workloads  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is a location-driven platform.

Core workflows depend on geographic information such as:

```text
Driver Location
Pickup Location
Drop Location
Service Regions
Operational Boundaries
Geographic Eligibility
Distance
Nearby Entity Queries
```

RideForge already uses:

```text
PostgreSQL
Redis
Redpanda
```

ADR-0007 established PostgreSQL as the primary transactional database.

ADR-0008 established Redis as the infrastructure for appropriate real-time and low-latency operational state.

RideForge therefore needs a clear decision for geospatial data that belongs with durable relational data.

---

# 2. Problem

Geospatial workloads require capabilities beyond ordinary relational comparisons.

Examples include:

```text
Point-in-Region Queries
Distance Calculations
Spatial Filtering
Bounding-Box Queries
Region Membership
Geographic Relationships
```

The platform must determine:

```text
Which geospatial data belongs in PostgreSQL
Which data belongs in Redis
When specialized geospatial infrastructure is justified
```

The solution must also preserve the distinction between:

```text
Durable Geographic Data
```

and:

```text
High-Frequency Real-Time Location State
```

---

# 3. Decision

RideForge will use:

> **PostGIS as the primary PostgreSQL geospatial extension for durable relational geospatial data.**

PostGIS extends the existing PostgreSQL foundation rather than introducing a separate geospatial database by default.

The intended model is:

```text
PostgreSQL + PostGIS
        ↓
Durable Relational + Geospatial Data
```

while:

```text
Redis
        ↓
Real-Time / High-Frequency Location State
```

and:

```text
Redpanda
        ↓
Location / Domain Events Where Required
```

---

# 4. Primary Principle

> **Use PostGIS when geospatial information is durable, relational, queryable, and belongs to PostgreSQL-owned domain state.**

Use Redis or specialized infrastructure when the workload is dominated by:

```text
High-Frequency Updates
Very Low-Latency Current State
Large-Scale Real-Time Location Processing
```

---

# 5. Why PostGIS

PostGIS allows RideForge to keep relational and spatial data within the same PostgreSQL platform.

This provides a unified environment for:

```text
Transactions
Relational Constraints
Indexes
SQL
Geospatial Queries
Geographic Relationships
```

It avoids introducing a separate geospatial database before the workload requires one.

---

# 6. Geospatial Data Categories

RideForge should distinguish between several categories.

## 6.1 Durable Geographic Data

Examples:

```text
Service Regions
Region Boundaries
Pickup / Drop Coordinates
Operational Areas
Geographic Restrictions
```

PostGIS is the preferred PostgreSQL-backed solution.

## 6.2 Real-Time Location State

Examples:

```text
Current Driver Location
Recent Driver Position
Live Availability Location
```

Redis or specialized location infrastructure may be more appropriate.

## 6.3 Historical Location Data

Examples:

```text
Driver Location History
Ride Route History
Historical Spatial Events
```

Storage should be selected according to:

```text
Volume
Retention
Analytics Requirements
Query Requirements
Cost
```

---

# 7. PostgreSQL + PostGIS Architecture

The conceptual architecture is:

```text
                    Geospatial Workloads
                           │
          ┌────────────────┴────────────────┐
          │                                 │
   Durable Spatial Data             Real-Time Location
          │                                 │
          ▼                                 ▼
 PostgreSQL + PostGIS                     Redis
          │                                 │
          └────────────────┬────────────────┘
                           │
                    RideForge Services
```

A specialized geospatial store may be introduced later if actual scale justifies it.

---

# 8. Spatial Data Ownership

Geospatial data should follow domain ownership.

For example:

```text
Region Domain
    ↓
Region Geometry
    ↓
PostgreSQL + PostGIS
```

A service should not directly modify another domain's spatial data merely because the database is physically shared.

---

# 9. Geometry vs Geography

RideForge should choose the appropriate PostGIS spatial type according to the intended semantics.

Conceptually:

```text
Geometry
→ Coordinate-space operations

Geography
→ Earth-based geographic calculations
```

The choice should be made based on:

```text
Coordinate System
Distance Requirements
Query Type
Performance
```

rather than using one type indiscriminately.

---

# 10. Coordinate Reference Systems

Geospatial data must have an explicitly understood coordinate reference system.

For common GPS coordinates:

```text
Latitude
Longitude
```

the application should consistently use the agreed geographic coordinate system.

Mixed coordinate systems must not be silently combined.

---

# 11. Coordinate Validation

Incoming coordinates should be validated before persistence.

At minimum:

```text
Latitude Range
Longitude Range
Expected Coordinate System
```

should be validated.

Invalid coordinates must not silently enter durable spatial data.

---

# 12. Spatial Precision

RideForge should avoid storing unnecessary precision when it provides no business value.

Precision should be selected according to:

```text
Location Accuracy
Use Case
Storage Cost
Query Requirements
Privacy Requirements
```

---

# 13. Spatial Indexing

PostGIS spatial indexes should be used for spatial queries that require them.

Indexing decisions should be based on:

```text
Query Patterns
Dataset Size
Spatial Selectivity
Write Frequency
```

Do not add spatial indexes without an actual query requirement.

---

# 14. Spatial Index Trade-Off

Spatial indexes improve relevant query performance but introduce:

```text
Storage Cost
Write Cost
Maintenance Cost
Index Build Time
```

Therefore index design must be workload-driven.

---

# 15. Region Boundaries

Service and operational regions are strong candidates for PostGIS.

Examples include:

```text
Service Areas
City Boundaries
Operational Zones
Restricted Areas
Pickup Zones
```

A region can be represented as a spatial geometry and queried using spatial relationships.

---

# 16. Point-in-Polygon Queries

A common RideForge requirement is determining whether a location belongs to a region.

Conceptually:

```text
Driver / Pickup Point
        ↓
Point-in-Polygon Query
        ↓
Region
```

This may support:

```text
Service Eligibility
Regional Configuration
Operational Routing
```

---

# 17. Distance Queries

PostGIS may be used for durable spatial queries involving distance.

Examples include:

```text
Distance Between Locations
Nearby Stored Locations
Spatial Filtering
Regional Distance Analysis
```

For real-time high-frequency candidate discovery, Redis may still be preferred where it provides the required latency.

---

# 18. Nearest-Entity Queries

PostGIS may support nearest-entity queries over durable spatial datasets.

The query should use appropriate spatial indexing and bounded result sets.

Do not perform unrestricted distance calculations over large tables when an indexed spatial query can reduce the candidate set.

---

# 19. Spatial Query Design

Spatial queries should generally follow:

```text
Indexed Spatial Filter
        ↓
Reduced Candidate Set
        ↓
Precise Spatial Calculation
        ↓
Business Filtering
```

This avoids unnecessarily expensive calculations across entire datasets.

---

# 20. PostGIS and Driver Matching

PostGIS can support durable geographic filtering.

For example:

```text
Ride Pickup Point
      ↓
PostGIS Spatial Query
      ↓
Geographically Relevant Drivers / Regions
```

However, high-frequency driver location should generally remain in the real-time location infrastructure.

Matching should combine:

```text
Real-Time Location
+
Durable Eligibility
+
Business Rules
+
Dispatch Strategy
```

---

# 21. PostGIS and Dispatch

Dispatch may use PostGIS for durable geographic information such as:

```text
Service Regions
Operational Zones
Geographic Constraints
Region Membership
```

Real-time dispatch state may remain in Redis.

---

# 22. PostGIS and Stand Dispatch

Stand Dispatch may use PostGIS for:

```text
Stand Locations
Service Areas
Region Boundaries
Geographic Eligibility
```

while Redis may maintain:

```text
Current Driver Position
Driver Availability
Temporary Stand State
```

where appropriate.

---

# 23. PostGIS and Smart Dispatch

Smart Dispatch may use PostGIS-derived information as part of feature generation or eligibility filtering.

Examples include:

```text
Region
Distance
Geographic Zone
Historical Spatial Context
```

Real-time features may come from Redis or other low-latency systems.

---

# 24. PostGIS and ETA

ETA systems may use PostGIS for:

```text
Historical Spatial Analysis
Region Mapping
Geographic Features
```

but should not depend on expensive transactional geospatial queries for every latency-sensitive ETA prediction.

---

# 25. PostGIS and AI

PostGIS data may contribute to AI/ML datasets such as:

```text
Pickup Region
Drop Region
Distance
Geographic Features
Historical Spatial Context
```

Spatial features should be transformed into model-appropriate representations rather than blindly exposing raw geometry to every model.

---

# 26. PostGIS and Regional / Legal Rules

Geographic eligibility may be represented using PostGIS boundaries.

However:

> **Spatial containment is not itself the complete legal decision.**

A workflow may require:

```text
Spatial Region
+
Vehicle / Driver Eligibility
+
Operating Rules
+
Regulatory Constraints
+
Business Policy
```

PostGIS provides geographic facts; domain logic determines the business decision.

---

# 27. Geographic Boundaries

Boundaries should be versioned or otherwise managed carefully when historical correctness matters.

A region can change over time.

Therefore workflows involving historical analysis may need:

```text
Effective From
Effective Until
Version
```

or an equivalent strategy.

---

# 28. Region Versioning

If a region boundary changes:

```text
Old Boundary
      ↓
Historical Interpretation

New Boundary
      ↓
Current Operations
```

should remain distinguishable where required.

---

# 29. Historical Geographic Data

Historical spatial data may require different retention policies from current operational data.

Potential categories include:

```text
Current Region
Historical Region
Historical Ride Point
Historical Route
Historical Location Sample
```

Retention should follow actual business and analytical requirements.

---

# 30. Location History

High-frequency location history should not automatically be stored indefinitely in PostgreSQL.

At scale, consider:

```text
Aggregation
Partitioning
Archival
Event-Based Storage
Analytics Storage
```

according to workload requirements.

---

# 31. PostGIS vs Redis

The intended distinction is:

```text
PostGIS
→ Durable Relational Geospatial Data

Redis
→ Current / Low-Latency Geospatial State
```

For example:

```text
Service Region Polygon
→ PostGIS

Current Driver Position
→ Redis
```

---

# 32. PostGIS vs Redpanda

The intended distinction is:

```text
PostGIS
→ Current / Durable Spatial State

Redpanda
→ Spatial Events
```

For example:

```text
Region Updated
→ Event

Region Geometry
→ PostgreSQL + PostGIS
```

---

# 33. PostGIS vs Specialized Geospatial Databases

A specialized geospatial or distributed location database may eventually be justified if:

```text
Scale
Throughput
Latency
Global Distribution
Query Requirements
```

exceed what the PostgreSQL + Redis architecture can efficiently support.

Such a transition requires a separate ADR.

---

# 34. Database Transactions

Spatial updates that belong to a transactional domain should participate in the same PostgreSQL transaction as the related relational state where appropriate.

Conceptually:

```text
BEGIN
  │
  ├── Update Domain State
  │
  └── Update Spatial State
  │
COMMIT
```

This is an important advantage of keeping durable spatial data within PostgreSQL.

---

# 35. Spatial Constraints

Where appropriate, database constraints and application validation should protect:

```text
Valid Coordinates
Required Geometry
Expected Geometry Type
Required Relationships
```

Domain rules remain the responsibility of application/domain logic.

---

# 36. Geometry Validity

Polygon and other complex geometries should be validated before they become authoritative region data.

Invalid geometry can cause:

```text
Query Errors
Incorrect Spatial Relationships
Unexpected Performance
```

---

# 37. Geometry Repair

If geometry repair is required, it must be deliberate.

Do not silently transform invalid business boundaries without understanding the resulting shape and operational meaning.

---

# 38. Spatial Data Ingestion

External geographic datasets should follow:

```text
Ingest
  ↓
Validate
  ↓
Normalize Coordinate System
  ↓
Validate Geometry
  ↓
Review
  ↓
Persist
```

---

# 39. Spatial Data Quality

Important quality checks include:

```text
Coordinate Validity
Geometry Validity
Coordinate Reference System
Unexpected Empty Geometry
Unexpected Large Geometry
Boundary Consistency
Duplicate Geometry
```

---

# 40. Spatial Data Migration

Changes to spatial schemas or data must be managed through version-controlled database migrations.

Large spatial migrations should consider:

```text
Table Size
Index Build Time
Lock Duration
Deployment Strategy
Rollback / Recovery
```

---

# 41. Spatial Index Migration

Creating or rebuilding large spatial indexes may be operationally expensive.

Production index changes should consider:

```text
Locking
Build Time
CPU
Disk
Concurrent Traffic
```

---

# 42. Spatial Query Performance

Performance analysis should use PostgreSQL query-planning tools where appropriate.

Important measurements include:

```text
Execution Time
Rows Examined
Index Usage
Candidate Set Size
Spatial Calculation Cost
```

---

# 43. Bounding-Box Filtering

For appropriate workloads, an initial bounding-box or indexed spatial filter can reduce the candidate set before more expensive exact spatial calculations.

Conceptually:

```text
Large Dataset
      ↓
Indexed Spatial Filter
      ↓
Small Candidate Set
      ↓
Exact Spatial Predicate
```

---

# 44. Avoid Full Spatial Scans

Unbounded spatial scans should be avoided for latency-sensitive workloads.

If a query repeatedly scans large spatial datasets:

```text
Review Query
      ↓
Review Index
      ↓
Review Data Model
      ↓
Review Access Pattern
```

before simply scaling hardware.

---

# 45. Pagination

Large geospatial result sets should be bounded.

Avoid returning thousands of spatial candidates when the business workflow needs only a small number of candidates.

---

# 46. Query Timeouts

Spatial queries should have appropriate timeouts.

An expensive spatial query must not indefinitely consume:

```text
Database Connection
CPU
Application Worker
```

---

# 47. Connection Management

PostGIS uses the same PostgreSQL connection infrastructure.

Therefore the same principles apply:

```text
Bounded Connection Pools
Connection Timeouts
Query Timeouts
Proper Connection Release
```

---

# 48. Caching Spatial Results

Spatial query results may be cached in Redis when:

```text
The Result Is Safe to Cache
Freshness Is Defined
Invalidation Is Understood
```

Examples may include:

```text
Region Lookup
Static Geographic Configuration
Frequently Requested Spatial Metadata
```

---

# 49. Spatial Cache Invalidation

When durable spatial data changes:

```text
PostGIS Update
      ↓
Cache Invalidation / Refresh
      ↓
Redis
```

must occur according to the selected cache strategy.

---

# 50. Geospatial Eventing

Changes to important geographic data may produce events.

For example:

```text
RegionCreated
RegionUpdated
RegionDeactivated
```

The event should represent the business fact, while PostGIS remains the durable spatial store.

---

# 51. Outbox Integration

If a spatial update must reliably publish an event:

```text
PostgreSQL Transaction
      │
      ├── Spatial Change
      │
      ├── Domain State Change
      │
      └── Outbox Record
                ↓
             Redpanda
```

The Outbox pattern remains the preferred reliability mechanism.

---

# 52. Spatial Data Security

Geospatial data may contain sensitive operational information.

Access should follow:

```text
Least Privilege
Domain Ownership
Need to Know
```

---

# 53. Location Privacy

Location data may reveal:

```text
Driver Movement
User Pickup / Drop Locations
Operational Patterns
```

Retention and access should therefore be controlled.

---

# 54. Location Logging

Applications should not log complete location streams by default.

When location is logged for debugging:

```text
Minimize Precision
Minimize Retention
Restrict Access
```

where appropriate.

---

# 55. Geospatial Data Retention

Retention should be defined by data type.

For example:

```text
Current Region Boundary
→ Long-Lived

Current Driver Location
→ Short-Lived

Historical Location
→ Explicit Retention Policy
```

---

# 56. Spatial Data Backup

Durable PostGIS data must be included in PostgreSQL backup and recovery procedures.

Backup validation must include restoration of spatial data.

---

# 57. Disaster Recovery

Recovery procedures should verify:

```text
Spatial Tables
Spatial Indexes
Coordinate Metadata
Region Boundaries
Related Relational State
```

---

# 58. Local Development

Local RideForge development should support PostGIS where geospatial features are being developed.

The local environment may include:

```text
PostgreSQL + PostGIS
Redis
Redpanda
RideForge Services
```

as required by the workflow.

---

# 59. Local Spatial Data

Local spatial datasets should be:

```text
Small
Deterministic
Safe to Reset
Representative Enough for Tests
```

Production geographic datasets should not be copied into local development unnecessarily.

---

# 60. Geospatial Testing

Tests should cover:

```text
Coordinate Validation
Geometry Validity
Point-in-Polygon
Distance Queries
Nearby Queries
Region Boundaries
Spatial Index Behaviour
Migration Behaviour
```

---

# 61. Spatial Integration Tests

Real PostgreSQL/PostGIS infrastructure should be used when testing actual:

```text
Spatial Operators
Spatial Indexes
Coordinate Systems
Geometry Behaviour
```

Mocks cannot validate real PostGIS query semantics.

---

# 62. Boundary Tests

Region tests should include:

```text
Point Clearly Inside
Point Clearly Outside
Point Near Boundary
Point Exactly on Boundary
Invalid Coordinate
```

Boundary semantics must be explicit.

---

# 63. Distance Tests

Distance calculations should be tested with known geographic examples and expected tolerances.

Tests should distinguish:

```text
Geographic Distance
Coordinate-Space Distance
```

according to the selected spatial type and query semantics.

---

# 64. Spatial Failure Tests

Important failure scenarios include:

```text
Invalid Geometry
Invalid Coordinate
Missing Spatial Data
PostGIS Unavailable
Slow Spatial Query
Corrupted / Invalid Imported Data
```

---

# 65. Load Testing

Geospatial workloads should be load-tested where they are latency-sensitive.

Measure:

```text
Queries / Second
Execution Latency
Candidate Set Size
CPU
Memory
Index Efficiency
Database Connections
```

---

# 66. Observability

Important geospatial database telemetry includes:

```text
Spatial Query Latency
Slow Spatial Queries
Index Usage
Database Connections
CPU
Memory
Storage
Table Growth
Index Growth
```

Operational services should additionally monitor:

```text
Invalid Coordinates
Stale Location State
Region Lookup Failures
```

where relevant.

---

# 67. Spatial Query Metrics

For critical queries, monitor:

```text
Query Count
Latency
Rows Examined
Rows Returned
Timeouts
Errors
```

This helps distinguish:

```text
Database Problem
```

from:

```text
Application / Query Design Problem
```

---

# 68. Cost Management

PostGIS is preferred partly because it extends the existing PostgreSQL platform.

This avoids introducing a separate geospatial database before there is a measurable need.

However, large spatial datasets still have:

```text
Storage Cost
Index Cost
CPU Cost
Backup Cost
Query Cost
```

These must be monitored.

---

# 69. Performance Optimization Principles

Use this optimization order:

```text
Correct Spatial Model
      ↓
Correct Coordinate System
      ↓
Correct Spatial Predicate
      ↓
Correct Spatial Index
      ↓
Candidate Reduction
      ↓
Query Optimization
      ↓
Caching
      ↓
Scaling
```

Do not immediately introduce a new database because one spatial query is slow.

---

# 70. Real-Time Location Boundary

PostGIS should not automatically become the destination for every driver GPS update.

At high update frequencies, prefer:

```text
Redis
```

or another specialized location system where justified.

PostGIS remains appropriate for durable spatial state and selected analytical workloads.

---

# 71. Location Sampling

If historical location must be retained, the system should consider whether every GPS sample is actually required.

Possible strategies include:

```text
Sampling
Aggregation
Route Simplification
Time-Based Retention
Event-Based Retention
```

The correct strategy depends on business and analytical requirements.

---

# 72. Geographic Feature Engineering

AI pipelines may derive features such as:

```text
Distance to Pickup
Distance to Region Boundary
Pickup Region
Drop Region
Historical Demand Zone
```

Feature generation should avoid repeatedly performing expensive spatial queries on the transactional database during model inference.

---

# 73. Offline Spatial Processing

Historical spatial features should preferably be generated through:

```text
Batch Processing
Feature Pipelines
Precomputed Tables
Read Models
```

when online query cost becomes significant.

---

# 74. Online Spatial Processing

Online spatial queries should be:

```text
Bounded
Indexed
Latency-Aware
Observable
```

They should not perform large historical scans during critical request paths.

---

# 75. Service Region Lookup

A typical service-region workflow may be:

```text
Pickup Coordinate
      ↓
PostGIS Region Query
      ↓
Region ID
      ↓
Business Rules
      ↓
Ride Eligibility
```

The region result may be cached when appropriate.

---

# 76. Cross-Region Decisions

Geospatial containment can identify:

```text
Origin Region
Destination Region
```

but the final ride decision may require:

```text
Origin Region
+
Destination Region
+
Vehicle Rules
+
Operating Rules
+
Legal Rules
```

Geospatial data is an input to the decision, not the entire decision.

---

# 77. Geographic Data Ownership During Service Extraction

If a geographic domain is later extracted into a separate service:

```text
Current PostgreSQL Ownership
      ↓
Logical Data Isolation
      ↓
Service Contract
      ↓
Independent Storage Where Justified
```

PostGIS does not require a separate service or database from the beginning.

---

# 78. Alternatives Considered

## 78.1 PostgreSQL Without PostGIS

### Advantages

```text
Simpler Initial Database
Fewer Extensions
```

### Disadvantages

```text
Weak Native Spatial Capabilities
Manual Spatial Logic
Poorer Spatial Query Model
```

### Decision

PostGIS is preferred for durable spatial workloads.

---

## 78.2 Redis Only for Geospatial Data

### Advantages

```text
Very Low Latency
Real-Time Access
Efficient Nearby Queries
```

### Disadvantages

```text
Not Ideal for Durable Relational Spatial State
Limited Transactional Domain Integration
Different Persistence Model
```

### Decision

Redis remains appropriate for real-time location, while PostGIS owns durable relational geospatial data.

---

## 78.3 Dedicated Geospatial Database

A dedicated geospatial database was considered.

### Advantages

```text
Specialized Spatial Capabilities
Potential Large-Scale Optimization
```

### Disadvantages

```text
Additional Infrastructure
Additional Operational Complexity
Additional Cost
Data Synchronization
```

### Decision

Not required as the default architecture. Introduce only when workload evidence justifies it.

---

## 78.4 Application-Level Geospatial Calculations

Performing all geographic calculations inside application code was considered.

### Advantages

```text
Simple Database Schema
No Spatial Extension
```

### Disadvantages

```text
Inefficient Large Dataset Processing
Poor Queryability
Increased Application Complexity
Difficult Indexing
```

### Decision

PostGIS is preferred for durable database-side spatial operations.

---

# 79. Decision Drivers

The decision is primarily driven by:

```text
Existing PostgreSQL Architecture
Durable Spatial Data
Relational + Spatial Queries
Operational Simplicity
Cost Efficiency
Geospatial Query Capability
Transaction Integration
Future Scalability
```

---

# 80. What This Decision Does Not Mean

This ADR does not mean:

```text
All Driver Locations Must Be Stored in PostGIS
```

It does not mean:

```text
PostGIS Replaces Redis
```

It does not mean:

```text
PostGIS Replaces a Future Specialized Location System
```

It does not mean:

```text
Every Geospatial Query Must Run on PostgreSQL
```

It establishes PostGIS as the default PostgreSQL-backed solution for durable relational geospatial data.

---

# 81. Decision Matrix

| Workload                        |     PostgreSQL + PostGIS |                 Redis |            Redpanda | Specialized Store |
| ------------------------------- | -----------------------: | --------------------: | ------------------: | ----------------: |
| Service region polygons         |              **Primary** |                 Cache |        Event output |      Not required |
| Region membership               |              **Primary** |        Possible cache |                  No |      Not required |
| Durable pickup/drop coordinates |              **Primary** |        Possible cache | Events where useful |      Not required |
| Current driver location         |                 Possible | **Primary candidate** | Events where useful |     Future option |
| Nearby active drivers           |                 Possible | **Primary candidate** |                  No |     Future option |
| Historical spatial analysis     |    **Primary candidate** |                    No |   Supporting events |     Future option |
| Geographic configuration        |              **Primary** |                 Cache |              Events |                No |
| Durable spatial business state  |              **Primary** |                    No |                  No |                No |
| High-frequency global location  | Not preferred by default |             Candidate |          Supporting |     Future option |

---

# 82. Consequences

## 82.1 Positive Consequences

The decision provides:

```text
PostgreSQL + Spatial Capability
Durable Spatial State
Transactional Integration
Spatial Indexing
Relational + Geographic Queries
Reduced Infrastructure Count
Clear Separation from Real-Time State
```

---

## 82.2 Negative Consequences

The architecture introduces:

```text
PostGIS Extension Management
Spatial Schema Complexity
Spatial Index Maintenance
Coordinate-System Complexity
Spatial Query Optimization
Potentially Large Spatial Storage
```

These costs are accepted because geographic data is fundamental to RideForge.

---

# 83. Risks

## Risk 1 — PostGIS Used for High-Frequency Location Writes

### Mitigation

Keep high-frequency current location in Redis or a specialized location system when PostgreSQL write pressure becomes excessive.

---

## Risk 2 — Incorrect Coordinate Systems

### Mitigation

Standardize coordinate systems and validate all spatial inputs.

---

## Risk 3 — Invalid Geometries

### Mitigation

Validate geometry before it becomes authoritative.

---

## Risk 4 — Slow Spatial Queries

### Mitigation

Use:

```text
Spatial Indexes
Candidate Reduction
Bounded Queries
EXPLAIN / EXPLAIN ANALYZE
Caching
```

where appropriate.

---

## Risk 5 — Geographic Data Becomes Too Large

### Mitigation

Use:

```text
Partitioning
Aggregation
Archival
Specialized Storage
```

when evidence justifies it.

---

## Risk 6 — Spatial Data Becomes Inconsistent Across Systems

### Mitigation

Define:

```text
Authoritative Source
Synchronization Strategy
Event Contracts
Cache Invalidation
```

for every duplicated spatial dataset.

---

# 84. Validation

The PostGIS decision should be validated through:

```text
Spatial Unit Tests
Integration Tests
Boundary Tests
Query Plan Analysis
Load Tests
Migration Tests
Failure Tests
Backup / Restore Tests
Production Monitoring
```

---

# 85. Review Triggers

Revisit this ADR when:

```text
Spatial Query Volume Increases Significantly
PostGIS Becomes a Major Performance Bottleneck
Real-Time Location Volume Increases Significantly
A Dedicated Geospatial Store Is Proposed
Global Geographic Scale Is Introduced
PostGIS Storage Cost Becomes Material
Spatial Workloads Require Capabilities Beyond PostgreSQL
```

---

# 86. Related Documentation

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
Smart Dispatch AI
ETA and Prediction System
Driver Demand and Supply Prediction
AI Matching and Ranking
Feature Engineering
```

---

# 87. Related ADRs

This decision is directly related to:

```text
ADR-0002 — Architecture Style
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0007 — PostgreSQL as Primary Database
ADR-0008 — Redis for Real-Time State
ADR-0012 — Outbox Pattern
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0021 — Failure and Degradation Strategy
ADR-0028 — Cost Optimization Strategy
```

---

# 88. Decision Summary

RideForge adopts:

```text
PostgreSQL + PostGIS
```

as the primary solution for durable relational geospatial data.

The intended architecture is:

```text
PostGIS
→ Durable Geographic Data

Redis
→ Real-Time Location / Low-Latency State

Redpanda
→ Geographic / Domain Events

Specialized Geospatial Store
→ Introduced Only When Workload Evidence Justifies It
```

This allows RideForge to use PostgreSQL's transactional and relational capabilities while supporting geographic operations without prematurely adding another database platform.

---

# 89. Final Principle

> **PostGIS is the default home for durable relational geospatial data; Redis handles real-time location workloads where low latency and high update frequency matter, and specialized geospatial infrastructure is introduced only when actual scale requires it.**

The intended relationship is:

```text
                    RideForge Geospatial Architecture
                               │
                ┌──────────────┼──────────────┐
                │              │              │
           PostGIS           Redis         Redpanda
                │              │              │
        Durable Spatial    Real-Time      Spatial /
             Data          Location       Domain Events
                │              │              │
                └──────────────┼──────────────┘
                               │
                          RideForge
                           Services
```

This establishes a clear boundary between:

```text
Durable Spatial State
```

and:

```text
Real-Time Spatial State
```

without prematurely introducing a dedicated geospatial database.

---

# 90. Status

```text
Decision: ACCEPTED

Primary Geospatial Store:
PostgreSQL + PostGIS

Primary Role:
Durable Relational Geospatial Data

Real-Time Location:
Redis / Specialized Infrastructure Where Justified
```

This decision establishes PostGIS as the RideForge geospatial foundation and provides the basis for future decisions involving location architecture, geospatial scaling, and specialized spatial infrastructure.
