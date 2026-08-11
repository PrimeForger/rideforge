# ADR-0010: Driver Location Storage Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Data / Real-Time Infrastructure  
> **Scope:** RideForge driver location storage, retrieval, freshness, and lifecycle  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

Driver location is one of the most latency-sensitive data domains in RideForge.

A driver may continuously produce location updates containing information such as:

```text
Latitude
Longitude
Timestamp
Heading
Speed
Accuracy
Driver Availability Context
```

RideForge uses this information for:

```text
Driver Discovery
Dispatch
Matching
Smart Dispatch
Stand Dispatch
ETA
Driver Availability
Geospatial Filtering
Operational Monitoring
```

The location workload is fundamentally different from ordinary transactional business data.

A typical driver can generate frequent location updates while thousands of drivers may be active simultaneously.

Persisting every real-time location update directly into the primary PostgreSQL transactional workload can create unnecessary:

```text
Write Load
Connection Pressure
Index Maintenance
Storage Growth
Query Contention
Database Cost
Latency
```

At the same time, driver location must be available quickly for latency-sensitive dispatch and matching.

ADR-0007 establishes PostgreSQL as the primary transactional database.

ADR-0008 establishes Redis as the infrastructure for appropriate real-time, low-latency, cached, and ephemeral operational state.

ADR-0009 establishes PostGIS as the PostgreSQL-backed solution for durable relational geospatial data.

A specific decision is therefore required for the storage of **current driver location**.

---

# 2. Problem

RideForge needs a driver-location architecture that provides:

```text
Low-Latency Location Reads
High-Frequency Location Writes
Nearby Driver Discovery
Location Freshness
Stale Driver Detection
Horizontal Scalability
Controlled Memory Usage
Failure Recovery
Operational Simplicity
```

The architecture must also clearly distinguish:

```text
Current Driver Location
```

from:

```text
Durable Driver Business State
```

and:

```text
Historical Driver Location
```

---

# 3. Decision

RideForge will use:

> **Redis as the primary real-time storage and retrieval layer for current driver location.**

PostgreSQL/PostGIS remains responsible for appropriate durable geospatial data and historical data where explicitly required.

The default architecture is:

```text
Driver App
    ↓
Location Ingestion
    ↓
Redis
    ↓
Current Driver Location
    ↓
Dispatch / Matching / ETA
```

while:

```text
PostgreSQL + PostGIS
    ↓
Durable Geographic / Business Data
```

and:

```text
Historical Location Pipeline
    ↓
Durable / Analytical Storage
```

may be used for historical location data according to retention and analytical requirements.

---

# 4. Core Decision Principle

> **Current driver location is real-time operational state, not ordinary transactional business state.**

Therefore:

```text
Current Location
→ Redis

Durable Driver Profile
→ PostgreSQL

Durable Geographic Data
→ PostgreSQL + PostGIS

Historical Location
→ Dedicated historical pipeline/storage when required
```

---

# 5. Why Redis

Redis is selected for current driver location because the workload requires:

```text
Low Latency
High Write Throughput
High Read Throughput
Short-Lived State
Fast Geospatial Queries
Expiration
Operational Simplicity
```

Redis also integrates naturally with the already-established RideForge architecture.

---

# 6. Why Not PostgreSQL as the Primary Current-Location Store

PostgreSQL is capable of storing geographic points.

However, current driver location has a different access pattern:

```text
Frequent Updates
Frequent Reads
High Concurrency
Short Data Lifetime
```

Using PostgreSQL as the default destination for every GPS update can unnecessarily increase:

```text
Write Volume
Index Maintenance
Connection Usage
Storage Growth
Vacuum Work
Database Load
```

Therefore PostgreSQL should not be the default hot-path location store.

---

# 7. Why Not PostGIS as the Primary Real-Time Location Store

PostGIS is the correct solution for durable relational geospatial data.

It is not automatically the correct solution for high-frequency ephemeral driver location.

The distinction is:

```text
PostGIS
→ Durable Spatial State

Redis
→ Current Real-Time Spatial State
```

This avoids forcing a transactional database to serve as the high-frequency location cache.

---

# 8. Why Not Redpanda as the Current Location Store

Redpanda is the event-streaming platform.

It is appropriate for:

```text
Location Events
Location-Related Events
Event Processing
Data Pipelines
```

but it should not be treated as the primary queryable current-state store for dispatch.

The roles remain:

```text
Redpanda
→ Event Stream

Redis
→ Current Location State
```

---

# 9. Current Location Data Model

A current driver location record may conceptually contain:

```text
driver_id
latitude
longitude
timestamp
heading
speed
accuracy
```

Additional fields may be added only when they have a clear operational purpose.

Avoid storing unnecessary payloads in the hot location path.

---

# 10. Location Update Lifecycle

The intended lifecycle is:

```text
Driver App
    ↓
Location Update
    ↓
Validation
    ↓
Location Ingestion
    ↓
Redis
    ↓
Current Location
    ↓
Dispatch / Matching / ETA
```

If location history is required:

```text
Location Update
    ↓
Event / Historical Pipeline
    ↓
Historical Storage
```

The real-time path and historical path should remain logically separate.

---

# 11. Location Ingestion

The location ingestion layer should:

```text
Receive Location
Validate Input
Validate Driver Identity
Validate Coordinate Range
Check Timestamp
Update Real-Time State
Publish Historical/Event Data Where Required
```

The ingestion path should avoid unnecessary synchronous dependencies.

---

# 12. Coordinate Validation

Incoming coordinates must be validated.

At minimum:

```text
Latitude
Longitude
Timestamp
Driver Identity
```

should be checked.

Invalid coordinates must not overwrite valid current state.

---

# 13. Timestamp Validation

Location updates must carry a timestamp.

The system should detect:

```text
Future Timestamp
Very Old Timestamp
Out-of-Order Timestamp
Duplicate Timestamp
```

according to the defined ingestion policy.

---

# 14. Out-of-Order Updates

Mobile networks can deliver updates out of order.

Example:

```text
Update A → 10:00:05
Update B → 10:00:03
```

If B arrives after A, B should not automatically overwrite A.

The location update policy should prefer the newest valid location timestamp.

Conceptually:

```text
Incoming Timestamp
        ↓
Compare With Stored Timestamp
        ↓
Newer?
 ┌──────┴──────┐
Yes           No
 ↓             ↓
Update       Ignore
```

---

# 15. Duplicate Location Updates

Duplicate updates may occur because of:

```text
Network Retry
Client Retry
Transport Duplication
Mobile Connectivity
```

The ingestion layer should tolerate duplicate location updates without corrupting current state.

---

# 16. Location Freshness

Current location must always have an associated freshness concept.

A location should be classified conceptually as:

```text
Fresh
Stale
Expired
Unknown
```

The exact thresholds depend on RideForge operational requirements.

---

# 17. TTL

Current driver location should use an appropriate TTL.

Conceptually:

```text
Location Update
      ↓
Redis Key
      ↓
TTL
      ↓
Expiration
```

When the driver stops sending location updates, the location should eventually stop being considered current.

---

# 18. TTL Is a Safety Mechanism

TTL should prevent indefinitely stale locations from remaining in the real-time system.

However:

> **TTL alone is not a complete driver availability mechanism.**

Driver availability and driver location are related but distinct concepts.

---

# 19. Stale Location Handling

A stale location should not automatically qualify a driver for a latency-sensitive match.

Matching may require:

```text
Location Freshness
+
Driver Availability
+
Eligibility
+
Region
+
Dispatch Rules
```

---

# 20. Driver Availability vs Location

The architecture must distinguish:

```text
Driver Is Available
```

from:

```text
Driver Has a Recent Location
```

A driver may be:

```text
Available + Fresh Location
Available + Stale Location
Unavailable + Fresh Location
Unavailable + No Location
```

These states must not be conflated.

---

# 21. Redis Geospatial Index

Current driver locations may be represented using Redis geospatial capabilities.

Conceptually:

```text
driver locations
       ↓
Redis GEO structure
       ↓
Nearby Driver Query
```

The geospatial structure should contain only drivers whose current location is considered operationally valid.

---

# 22. Location Metadata

Geospatial indexing alone may not contain every field required for matching.

A complementary operational record may contain:

```text
Location
Timestamp
Availability
Vehicle Type
Operational State
```

However, durable business attributes should remain owned by their appropriate domain.

---

# 23. Nearby Driver Search

A typical lookup is:

```text
Pickup Location
      ↓
Redis Nearby Search
      ↓
Candidate Driver IDs
      ↓
Freshness Filter
      ↓
Availability Filter
      ↓
Eligibility Filter
      ↓
Ranking
```

Redis should primarily provide efficient candidate discovery.

---

# 24. Candidate Retrieval vs Candidate Eligibility

Location proximity does not establish eligibility.

The matching system must still evaluate:

```text
Driver Availability
Vehicle Compatibility
Region
Operating Rules
Ride Constraints
Driver Status
```

---

# 25. Candidate Retrieval vs Ranking

Redis may identify nearby candidates.

The ranking system may then evaluate:

```text
Distance
ETA
Driver State
Demand / Supply
Historical Features
AI Score
Dispatch Strategy
```

This separation prevents the location store from becoming the complete matching engine.

---

# 26. Location Update Frequency

The driver application should not send location updates more frequently than the platform can meaningfully use.

Update frequency should consider:

```text
Driver Movement
Ride State
Network Conditions
Battery
Server Capacity
Dispatch Requirements
```

The platform should support adaptive location frequency where appropriate.

---

# 27. Adaptive Location Updates

Possible behaviour:

```text
Driver Offline
→ No Location

Driver Available
→ Normal Location Frequency

Driver Near Pickup
→ Higher Frequency

Active Ride
→ Higher / Context-Specific Frequency

Idle for Long Period
→ Reduced Frequency
```

The exact values belong to implementation/configuration rather than this ADR.

---

# 28. Mobile Network Behaviour

Location ingestion must tolerate:

```text
Intermittent Connectivity
Delayed Packets
Duplicate Packets
Out-of-Order Packets
Temporary Disconnection
```

The system must not assume continuous network connectivity.

---

# 29. Driver Reconnection

When a driver reconnects:

```text
Driver Reconnects
      ↓
Fresh Location
      ↓
Redis State Updated
      ↓
Driver Becomes Location-Eligible
```

Old location should not remain indefinitely authoritative.

---

# 30. Driver Disconnect

A disconnected driver should eventually become:

```text
Location Stale / Expired
```

according to the freshness policy.

The system should not continue dispatching using an indefinitely old location.

---

# 31. Location Ingestion Idempotency

Location updates should tolerate retries.

If an update is received multiple times:

```text
Same Valid Update
       ↓
No Corruption
```

The ingestion mechanism may use:

```text
Timestamp
Sequence Number
Client Event ID
```

where necessary.

---

# 32. Client Sequence Numbers

A driver client may optionally provide a monotonically increasing location sequence number.

This can help identify:

```text
Out-of-Order Updates
Duplicate Updates
Missing Updates
```

when supported by the client protocol.

---

# 33. Location Accuracy

GPS accuracy should be considered when evaluating a location.

For example:

```text
High Accuracy
→ Suitable for Precise Use

Low Accuracy
→ May Require Filtering / Lower Confidence
```

The exact threshold is an operational decision.

---

# 34. Impossible Movement Detection

The system should be able to identify obviously invalid movement where practical.

Examples include:

```text
Impossible Speed
Large Location Jump
Invalid Coordinates
Timestamp Anomaly
```

Such updates should be rejected or flagged rather than blindly becoming current state.

---

# 35. Location Smoothing

Location smoothing should not be introduced blindly.

If smoothing is required, it should be handled in an appropriate location-processing layer and should preserve:

```text
Operational Accuracy
Latency
Traceability
```

---

# 36. Location Privacy

Driver location is sensitive operational information.

Access must be restricted to services that require it.

Avoid exposing precise driver locations to unauthorized consumers.

---

# 37. User-Facing Location Exposure

Customer-facing systems should not necessarily receive raw driver coordinates.

The platform may provide:

```text
Approximate Position
Map Position
ETA
Status
```

according to product and privacy requirements.

---

# 38. Location Security

Location APIs should require appropriate:

```text
Authentication
Authorization
Service Identity
Rate Limiting
Input Validation
```

---

# 39. Redis Key Strategy

A conceptual key structure may be:

```text
driver:location:<driver_id>
```

and a geospatial namespace may represent:

```text
drivers:geo
```

The exact key naming standard belongs to the Redis development conventions.

---

# 40. Key Ownership

The location subsystem owns its location keys.

Other services should not directly mutate location keys.

Prefer:

```text
Location Service / Ingestion
        ↓
Redis Location State
```

rather than:

```text
Multiple Services
        ↓
Direct Redis Mutation
```

---

# 41. Location Data Lifecycle

The lifecycle is:

```text
Created
   ↓
Updated
   ↓
Refreshed
   ↓
Stale
   ↓
Expired
   ↓
Removed
```

This lifecycle must be reflected in the Redis implementation.

---

# 42. Historical Location

Current location and historical location have different requirements.

Current location:

```text
Low Latency
Short Retention
High Update Rate
```

Historical location:

```text
Durability
Retention
Analytics
Auditing
Potentially Large Volume
```

They should therefore not automatically share the same storage path.

---

# 43. Historical Location Strategy

If historical location is required:

```text
Driver App
    ↓
Location Ingestion
    ├──────────────→ Redis
    │                 Current State
    │
    └──────────────→ Event / Pipeline
                      ↓
                Historical Storage
```

The historical storage technology may evolve independently.

---

# 44. Historical Storage

Historical location may eventually use:

```text
PostgreSQL Partitioned Tables
Object Storage
Analytics Database
Data Warehouse
Specialized Time-Series / Location Store
```

depending on:

```text
Volume
Retention
Query Requirements
Cost
```

This ADR does not force a single historical-storage technology.

---

# 45. Redpanda Integration

Location events may be published to Redpanda when downstream processing requires them.

Potential consumers include:

```text
Historical Pipeline
Analytics
AI Feature Pipeline
Monitoring
Fraud Detection
Operational Processing
```

Redpanda remains an event stream, not the current location query store.

---

# 46. Location Event Volume

Location updates can produce large event volumes.

The system should avoid publishing unnecessarily verbose events.

Location events should have:

```text
Defined Schema
Minimal Required Fields
Versioning
Retention Policy
```

---

# 47. Location Event vs Current State

These are distinct:

```text
Location Event
→ Something happened

Redis Location
→ Current known location
```

A location event stream may contain many historical updates while Redis contains only the current operational state.

---

# 48. Redis Failure

If Redis becomes unavailable:

```text
Current Location
→ Temporarily Unavailable
```

The platform must not silently use arbitrarily old location data.

Possible degradation includes:

```text
Reduce Matching Scope
Reject Stale Candidates
Temporarily Degrade Dispatch
Use Last Known Safe State Where Explicitly Allowed
```

---

# 49. Redis Recovery

After Redis recovers:

```text
Drivers
   ↓
Fresh Location Updates
   ↓
Redis Rebuilt
```

The system should prefer fresh incoming locations rather than treating pre-failure state as automatically current.

---

# 50. Redis Data Loss

Because current location is operational state:

```text
Redis Data Loss
```

must be recoverable through:

```text
Fresh Driver Updates
```

where possible.

This is one of the reasons current location should not be made dependent on durable PostgreSQL writes for every update.

---

# 51. PostgreSQL Fallback

PostgreSQL should not automatically be used as a fallback for every current-location lookup.

Using old PostgreSQL location data as though it were current can create unsafe dispatch decisions.

A fallback is valid only when:

```text
Freshness Is Known
Business Rules Allow It
```

---

# 52. Matching During Location Failure

If current location cannot be reliably obtained:

```text
Do Not Match Using Unknown Freshness
```

The matching system should prefer:

```text
No Candidate
```

over:

```text
Unsafe / Stale Candidate
```

when correctness requires it.

---

# 53. Dispatch Degradation

When location infrastructure is degraded, RideForge may fall back to:

```text
Stand Dispatch
Reduced Candidate Radius
Reduced Dispatch Capacity
Manual / Operational Handling
```

depending on the operating environment.

The exact degradation strategy is governed by the failure/degradation architecture.

---

# 54. Smart Dispatch During Location Failure

AI-assisted matching must not override location freshness safety rules.

AI may rank available candidates, but it must not transform stale or invalid location into valid real-time state.

---

# 55. Stand Dispatch During Location Failure

Stand Dispatch can provide a useful alternative where physical stand or queue state is available.

This supports the platform's hybrid dispatch strategy:

```text
Smart Dispatch
+
Stand Dispatch
```

without requiring invalid location data.

---

# 56. Location and ETA

ETA systems may consume:

```text
Current Driver Location
Pickup Location
Destination
Route Information
```

The current driver location must satisfy the ETA system's freshness requirements.

Stale location should reduce confidence or trigger fallback behaviour.

---

# 57. Location and Route Providers

Route providers may require current coordinates.

The platform should avoid repeatedly sending unchanged or stale coordinates to external providers.

Caching and update thresholds may reduce:

```text
API Cost
Network Traffic
Latency
```

where appropriate.

---

# 58. Location and AI Features

Location can contribute features such as:

```text
Distance to Pickup
Distance to Stand
Region
Movement Direction
Recent Movement
Estimated Arrival
```

Features should have timestamps and freshness semantics.

---

# 59. Location Feature Freshness

AI features derived from location should preserve the age of the underlying data where freshness matters.

Conceptually:

```text
Feature
+
Feature Timestamp
+
Source Timestamp
```

This prevents stale operational data from being treated as current.

---

# 60. Location and Demand Prediction

Demand prediction may use historical location information to derive:

```text
Demand Zones
Pickup Density
Spatial Demand Patterns
Time-Region Demand
```

Historical location should be processed through appropriate data pipelines rather than repeatedly querying the real-time Redis state.

---

# 61. Location and Supply Prediction

Supply prediction may use:

```text
Driver Distribution
Driver Availability
Regional Driver Density
Historical Driver Movement
```

Real-time and historical datasets should remain clearly separated.

---

# 62. Geospatial Indexing Strategy

Redis geospatial indexing should be used for current location discovery.

PostGIS should be used for durable spatial relationships.

The architecture is:

```text
Current Driver
     ↓
Redis GEO

Durable Region
     ↓
PostGIS
```

---

# 63. Geospatial Candidate Filtering

A nearby query should not automatically return every driver.

Apply appropriate filters for:

```text
Freshness
Availability
Vehicle Type
Region
Operational State
```

before final ranking.

---

# 64. Candidate Radius

The candidate radius may vary based on:

```text
Urban Density
Driver Supply
Demand
Ride Type
Traffic
Dispatch Strategy
```

The radius should be configurable rather than hard-coded into the architecture.

---

# 65. Progressive Radius Expansion

Where appropriate, matching may progressively expand the search radius:

```text
Radius 1
   ↓
No Suitable Candidate
   ↓
Radius 2
   ↓
No Suitable Candidate
   ↓
Radius 3
```

Each expansion should remain subject to:

```text
Freshness
Eligibility
Regional Rules
Dispatch Policy
```

---

# 66. Hot Regions

Some geographic areas may have much higher location traffic.

Monitor:

```text
Location Update Rate
Nearby Query Rate
Redis CPU
Redis Memory
Network
Hot Geospatial Areas
```

This may influence future sharding or partitioning strategies.

---

# 67. Scaling

The initial Redis location strategy should remain simple.

Scale only when measurements justify it.

Potential future strategies include:

```text
Redis Replication
Redis Cluster
Regional Redis
Location Sharding
Specialized Location Infrastructure
```

---

# 68. Regional Deployment

If RideForge expands geographically, location infrastructure may eventually require regional deployment.

A possible future model is:

```text
Region A
→ Redis Location

Region B
→ Redis Location

Region C
→ Redis Location
```

The exact architecture depends on:

```text
Latency
Data Locality
Availability
Operational Complexity
```

---

# 69. Cross-Region Location

Driver location should generally be processed close to the operational region where it is consumed when scale justifies regional deployment.

Cross-region synchronization should not be introduced unnecessarily.

---

# 70. Location Consistency

Current location does not require global strong consistency in the same way as financial or ride lifecycle state.

The primary requirement is:

```text
Fresh
Correctly Ordered
Operationally Available
```

within the appropriate regional context.

---

# 71. Location Ordering

The system should prefer the newest valid location.

Conceptually:

```text
Stored Timestamp = T1
Incoming Timestamp = T2

If T2 > T1
→ Update

If T2 <= T1
→ Ignore / Deduplicate
```

---

# 72. Location Atomicity

Location updates should be applied atomically enough to prevent partially updated state.

For example:

```text
Coordinates
+
Timestamp
+
Operational Metadata
```

should not result in a state where fields from unrelated updates are incorrectly combined.

---

# 73. Location Concurrency

Multiple location updates may arrive concurrently.

The storage mechanism should ensure that an older update cannot overwrite a newer valid update.

---

# 74. Location Security Boundary

Only authorized location-ingestion paths should be allowed to write current driver location.

Read access should also be restricted by service responsibility.

---

# 75. Location API Contract

The location ingestion API should define:

```text
Driver Identity
Latitude
Longitude
Timestamp
Optional Accuracy
Optional Heading
Optional Speed
```

The exact API contract belongs to API development documentation and service contracts.

---

# 76. Rate Limiting Location Updates

Location ingestion may require rate limiting or adaptive throttling to protect infrastructure.

Rate limiting must avoid causing legitimate active drivers to appear offline.

---

# 77. Backpressure

When location ingestion is under pressure, the platform should have controlled backpressure.

Possible mechanisms include:

```text
Sampling
Throttling
Batching
Queueing
Adaptive Frequency
```

The system should avoid uncontrolled growth of in-memory queues.

---

# 78. Location Batching

Batching may reduce network overhead when appropriate.

However, batching increases the time between individual location updates and therefore may reduce freshness.

Use batching only when:

```text
Latency Requirement
Network Cost
Update Frequency
```

allow it.

---

# 79. Location Compression

If location payload volume becomes material, payload compression or compact encoding may be considered.

Do not optimize prematurely.

---

# 80. Location Observability

The location system should expose operational metrics such as:

```text
Location Updates / Second
Location Update Latency
Redis Write Latency
Redis Read Latency
Fresh Drivers
Stale Drivers
Expired Drivers
Rejected Updates
Out-of-Order Updates
Duplicate Updates
```

---

# 81. Location Quality Metrics

Track:

```text
Invalid Coordinate Rate
GPS Accuracy Distribution
Timestamp Anomaly Rate
Impossible Movement Rate
Location Freshness
Update Success Rate
```

These metrics can reveal client and infrastructure problems.

---

# 82. Location Data Quality

The system should distinguish between:

```text
Storage Success
```

and:

```text
Location Quality
```

A successfully stored GPS coordinate may still be:

```text
Inaccurate
Stale
Impossible
```

---

# 83. Monitoring Freshness

A critical operational metric is:

```text
Current Location Age
```

For example:

```text
Now - Location Timestamp
```

This should be monitored across:

```text
All Drivers
Regions
Vehicle Types
Operational States
```

where useful.

---

# 84. Alerting

Alerts may be appropriate for:

```text
Large Increase in Stale Drivers
Redis Failure
Location Ingestion Failure
High Update Rejection Rate
High Redis Latency
Memory Pressure
Unexpected Location Drop
```

---

# 85. Testing

Location infrastructure should be tested at multiple levels.

### Unit Tests

```text
Coordinate Validation
Timestamp Ordering
Freshness Logic
Duplicate Handling
```

### Integration Tests

```text
Redis Writes
TTL
Geospatial Queries
Concurrency
Recovery
```

### Load Tests

```text
High Update Rate
High Read Rate
Large Driver Population
Hot Geographic Areas
```

---

# 86. Failure Testing

Test at minimum:

```text
Redis Unavailable
Redis Restart
Network Failure
Out-of-Order Updates
Duplicate Updates
Stale Driver
Expired Driver
Invalid Coordinate
High Update Volume
```

---

# 87. Recovery Testing

Verify that after Redis loss:

```text
Drivers Can Reconnect
Fresh Locations Reappear
Stale Drivers Are Not Treated as Fresh
Matching Recovers
Dispatch Recovers
```

---

# 88. Security Testing

Validate:

```text
Unauthorized Location Write
Unauthorized Location Read
Invalid Driver Identity
Rate-Limit Bypass
Malformed Coordinates
Malformed Payloads
```

---

# 89. Cost Considerations

Redis location storage is memory-sensitive.

Cost is influenced by:

```text
Driver Population
Active Driver Population
Location Update Rate
Key Count
Metadata Size
Geospatial Index Size
High Availability
Regional Deployment
```

The design should optimize the real-time state rather than retaining unnecessary historical data in Redis.

---

# 90. Memory Optimization

Current location records should remain compact.

Avoid storing:

```text
Large JSON Payloads
Historical Location Arrays
Unnecessary Metadata
Duplicate Business Data
```

inside every active-driver Redis record.

---

# 91. Data Retention

Redis should contain only the amount of location state required for current operations.

Historical location should leave Redis through:

```text
Expiration
Event Pipeline
Historical Storage
```

rather than accumulating indefinitely.

---

# 92. Privacy Retention

Location retention should follow the minimum necessary principle.

Different data types may have different retention:

```text
Current Location
→ Very Short

Operational Location History
→ Defined Business Retention

Analytics Data
→ Defined Analytics Retention
```

---

# 93. Alternative: PostgreSQL/PostGIS Only

### Advantages

```text
Single Durable Store
Strong Relational Integration
Simpler Data Model
```

### Disadvantages

```text
High Write Load
Connection Pressure
Storage Growth
Higher Hot-Path Database Dependency
```

### Decision

Rejected as the default current-location architecture.

PostGIS remains appropriate for durable spatial workloads.

---

# 94. Alternative: Redis + PostgreSQL Without Event Pipeline

### Advantages

```text
Simple
Low Latency
Few Components
```

### Disadvantages

```text
Poor Historical Data Pipeline
Limited Event Consumers
Difficult Analytics Integration
```

### Decision

Not sufficient for the long-term architecture where location events need to feed downstream systems.

---

# 95. Alternative: Redpanda as Current Location Store

### Advantages

```text
High Throughput
Durable Event Stream
Scalable Event Processing
```

### Disadvantages

```text
Poor Fit for Direct Current-State Queries
Additional State Reconstruction
Higher Query Complexity
```

### Decision

Rejected as the primary current-location query store.

Redpanda remains the event-streaming layer.

---

# 96. Alternative: Specialized Location Database

A specialized location database may eventually provide:

```text
Large-Scale Location Storage
High Throughput
Distributed Geospatial Queries
Regional Distribution
```

However, it introduces additional infrastructure.

### Decision

Deferred until workload evidence justifies the complexity.

---

# 97. Decision Matrix

| Requirement | Redis | PostgreSQL + PostGIS | Redpanda | Specialized Location Store |
|---|---:|---:|---:|---:|
| Current driver location | **Primary** | Secondary | No | Future option |
| High-frequency updates | **Primary** | Not preferred | Supporting | **Potential** |
| Nearby active drivers | **Primary** | Possible | No | **Potential** |
| Durable region data | No | **Primary** | No | No |
| Driver profile | No | **Primary** | No | No |
| Historical location | Temporary / No | Possible | **Pipeline source** | **Potential** |
| Event distribution | No | Outbox source | **Primary** | No |
| Low-latency current state | **Primary** | Secondary | No | **Potential** |
| TTL / freshness | **Primary** | Limited | Retention-based | Depends |
| Transactional business state | No | **Primary** | No | No |

---

# 98. Consequences

## 98.1 Positive Consequences

The decision provides:

```text
Low-Latency Driver Location
High-Frequency Update Support
Efficient Nearby Search
Reduced PostgreSQL Write Pressure
Natural TTL-Based Freshness
Clear Separation of Real-Time and Durable Data
```

---

## 98.2 Negative Consequences

The architecture introduces:

```text
Redis Dependency
Location Freshness Complexity
Event / Historical Pipeline Complexity
Additional Operational Monitoring
Potential Redis Memory Cost
Distributed-State Considerations
```

These trade-offs are accepted for the real-time location workload.

---

# 99. Risks

## Risk 1 — Redis Becomes a Hidden Source of Truth

### Mitigation

Document Redis location as current operational state and define recovery from fresh driver updates.

---

## Risk 2 — Stale Driver Location

### Mitigation

Use:

```text
Timestamp
TTL
Freshness Checks
Candidate Filtering
```

---

## Risk 3 — High Location Update Volume

### Mitigation

Use:

```text
Adaptive Update Frequency
Throttling
Batching Where Appropriate
Redis Scaling
Specialized Location Infrastructure When Justified
```

---

## Risk 4 — Redis Memory Exhaustion

### Mitigation

Keep current records compact and monitor:

```text
Memory
Key Count
Growth
Evictions
```

---

## Risk 5 — Out-of-Order Location Updates

### Mitigation

Compare timestamps or sequence numbers before accepting updates.

---

## Risk 6 — Location Quality Problems

### Mitigation

Validate:

```text
Coordinates
Timestamp
Accuracy
Movement
```

and monitor rejection/anomaly rates.

---

## Risk 7 — Historical Data Loss

### Mitigation

If history is required, publish location events to the appropriate durable pipeline rather than relying on Redis retention.

---

# 100. Validation

The decision should be validated through:

```text
Location Unit Tests
Redis Integration Tests
Geospatial Tests
Concurrency Tests
Load Tests
Failure Tests
Recovery Tests
Security Tests
Freshness Monitoring
Production Metrics
```

---

# 101. Review Triggers

Revisit this ADR when:

```text
Active Driver Population Increases Significantly
Location Update Rate Becomes Materially Larger
Redis Memory Cost Becomes Significant
Regional Deployment Is Required
Redis Cannot Meet Latency Requirements
Historical Location Volume Requires Dedicated Storage
A Specialized Location Database Is Proposed
Location Consistency Requirements Change
```

---

# 102. Related Documentation

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
API Development
Performance and Optimization
Observability Development
Integration Testing and Local Infrastructure
Smart Dispatch AI
ETA and Prediction System
AI Matching and Ranking
Machine Learning Data Pipeline
AI Failure and Fallback Strategy
```

---

# 103. Related ADRs

This decision is directly related to:

```text
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0007 — PostgreSQL as Primary Database
ADR-0008 — Redis for Real-Time State and Caching
ADR-0009 — PostGIS for Geospatial Operations
ADR-0011 — PgBouncer for Database Connection Pooling
ADR-0012 — Outbox Pattern
ADR-0013 — Dead Letter Queue Strategy
ADR-0015 — Smart Dispatch and Stand Dispatch
ADR-0016 — AI-Assisted Dispatch Strategy
ADR-0017 — ETA and Route Provider Strategy
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0021 — Failure and Degradation Strategy
ADR-0028 — Cost Optimization Strategy
```

---

# 104. Decision Summary

RideForge adopts:

```text
Redis
```

as the primary real-time storage layer for current driver location.

The complete location architecture is:

```text
Driver App
    │
    ▼
Location Ingestion
    │
    ├───────────────► Redis
    │                  │
    │                  ├── Current Location
    │                  ├── Freshness
    │                  └── Nearby Driver Search
    │
    └───────────────► Redpanda / Historical Pipeline
                           │
                           ▼
                    Historical Storage
```

While:

```text
PostgreSQL + PostGIS
        ↓
Durable Business / Geographic State
```

remains the authoritative storage for appropriate relational and spatial data.

---

# 105. Final Principle

> **Current driver location is a real-time operational state and should be optimized for freshness and low latency; durable business and geographic state remains in PostgreSQL/PostGIS, while historical location is handled through a dedicated durable pipeline when required.**

The intended separation is:

```text
                    DRIVER LOCATION
                          │
          ┌───────────────┼────────────────┐
          │               │                │
       Redis          Redpanda        PostgreSQL/PostGIS
          │               │                │
    Current State      Location        Durable Spatial
    + Nearby Search      Events            State
          │               │                │
          └───────────────┼────────────────┘
                          │
                     RideForge
                      Services
```

This provides a simple initial architecture that can scale without prematurely introducing a dedicated location database.

---

# 106. Status

```text
Decision: ACCEPTED

Current Driver Location:
Redis

Durable Geographic Data:
PostgreSQL + PostGIS

Location Events:
Redpanda

Historical Location:
Dedicated Durable Pipeline / Storage When Required
```

This decision establishes the RideForge driver-location storage strategy and provides the foundation for dispatch, matching, smart dispatch, ETA, location analytics, and future geographic scaling decisions.
