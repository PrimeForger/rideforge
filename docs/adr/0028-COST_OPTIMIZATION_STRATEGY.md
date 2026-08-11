# ADR-0028: Cost Optimization Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Cost / Infrastructure / Architecture / Operations  
> **Scope:** Cloud infrastructure, compute, databases, Redis, Kafka / Redpanda, networking, storage, observability, AI infrastructure, development environments, staging, production, scaling, capacity planning, and infrastructure evolution  
> **Owner:** RideForge Architecture / Platform Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is being designed as a production-grade ride-hailing platform with:

```text
Microservices
Event-Driven Architecture
PostgreSQL
PostGIS
Redis
Kafka / Redpanda
Real-Time Driver Location
Smart Dispatch
Stand Dispatch
AI / ML Components
External Providers
Containerized Services
Cloud Deployment
```

The architecture must eventually support:

```text
High Request Volume
Real-Time Location Updates
Dispatch
Event Processing
Geospatial Queries
AI Inference
Observability
High Availability
```

These capabilities can become expensive if infrastructure is scaled or selected without considering actual workload.

At the same time, excessive cost optimization can damage:

```text
Reliability
Performance
Developer Productivity
Operational Safety
Customer Experience
```

RideForge therefore requires a deliberate cost strategy that optimizes total system cost without compromising critical platform requirements.

---

# 2. Problem

Distributed systems can accumulate cost through:

```text
Over-Provisioned Compute
Oversized Databases
Excessive Redis Memory
Unnecessary Kafka / Redpanda Capacity
High Network Egress
Excessive Logging
Excessive Metrics Retention
Expensive AI Inference
Idle Development Infrastructure
Premature High-Availability Infrastructure
Premature Multi-Region Deployment
Uncontrolled Autoscaling
Unused Cloud Resources
```

The platform needs explicit principles for:

```text
What to optimize
When to optimize
What not to sacrifice
How to measure cost
How to scale infrastructure
How to control AI cost
How to evolve infrastructure economically
```

---

# 3. Decision

RideForge will adopt a:

```text
Measured
Workload-Driven
Risk-Aware
Incremental
Cost-Conscious
```

cost optimization strategy.

The primary rule is:

> **Optimize infrastructure cost based on measured workload and business value, while never sacrificing required reliability, security, data integrity, legal compliance, or critical ride-hailing functionality merely to reduce infrastructure cost.**

---

# 4. Cost Optimization Principles

RideForge follows:

```text
Measure Before Optimizing
Right-Size Before Over-Engineering
Scale With Demand
Avoid Idle Capacity
Prefer Managed Services When Economically Justified
Prefer Simpler Infrastructure at Lower Scale
Optimize High-Cost Paths First
Track Unit Economics
Protect Reliability
Automate Cost Controls
Review Cost Continuously
```

---

# 5. Cost Hierarchy

Infrastructure cost should be considered across:

```text
Compute
Database
Cache
Messaging
Storage
Networking
Observability
AI / ML
External Providers
CI/CD
Development
Staging
Production
```

---

# 6. Total Cost of Ownership

Cost decisions must consider more than the cloud invoice.

Total cost includes:

```text
Infrastructure Cost
Engineering Time
Operational Complexity
Maintenance
Monitoring
Incident Risk
Migration Cost
Vendor Lock-In
```

A cheaper infrastructure component is not automatically the better choice if it creates substantially higher operational burden.

---

# 7. Cost Per Business Unit

Where practical, RideForge should track cost relative to meaningful business units.

Examples:

```text
Cost per Ride
Cost per Completed Ride
Cost per Active Driver
Cost per 1,000 Ride Requests
Cost per 1 Million Location Updates
Cost per AI Prediction
```

The exact unit economics should evolve with the platform.

---

# 8. Cost Allocation

Costs should be attributable to major system areas where practical.

Conceptually:

```text
Ride Platform
├── API / Services
├── Database
├── Redis
├── Messaging
├── Location
├── Dispatch
├── AI
├── Observability
└── External Providers
```

---

# 9. Environment Cost Strategy

Different environments have different requirements.

```text
Development → Minimize Cost
Testing     → Short-Lived / Efficient
Staging     → Production-Relevant but Right-Sized
Production  → Reliability-First
```

---

# 10. Development Cost

Development environments should avoid continuously running production-sized infrastructure.

Prefer:

```text
Local Containers
Ephemeral Environments
Shared Non-Production Infrastructure
Scheduled Resources
```

where appropriate.

---

# 11. Local Development

Local development may use:

```text
Docker Compose
Local PostgreSQL
Local Redis
Local Kafka / Redpanda
```

to avoid unnecessary cloud spending.

---

# 12. Testing Cost

Testing infrastructure should be created only when required and destroyed when no longer needed where practical.

---

# 13. Ephemeral Test Environments

Ephemeral environments may be used for:

```text
Integration Tests
Pull Request Validation
Temporary QA
Migration Testing
```

when their setup and teardown cost is justified.

---

# 14. Staging Cost

Staging should be:

```text
Production-Like
Smaller
Right-Sized
```

rather than a permanent duplicate of production capacity.

---

# 15. Production Cost

Production must prioritize:

```text
Availability
Performance
Safety
Recovery
```

before aggressive cost reduction.

---

# 16. Compute Cost

Compute cost should be optimized through:

```text
Right-Sizing
Autoscaling
Workload Scheduling
Container Density
Idle Resource Removal
Efficient Runtime Configuration
```

---

# 17. Right-Sizing

Resources should be selected based on observed:

```text
CPU
Memory
Request Rate
Latency
Concurrency
Network
```

rather than arbitrary large instance sizes.

---

# 18. CPU Utilization

Persistently low CPU utilization may indicate over-provisioning.

However, CPU alone must not determine capacity.

---

# 19. Memory Utilization

Memory should be monitored carefully because memory pressure can cause:

```text
Out-of-Memory Failures
Garbage Collection Pressure
Performance Degradation
Restarts
```

---

# 20. Autoscaling

Autoscaling should be used where workloads vary sufficiently.

Autoscaling signals may include:

```text
CPU
Memory
Requests
Latency
Queue Lag
Concurrent Work
```

---

# 21. Autoscaling Limits

Every autoscaling configuration should have:

```text
Minimum Capacity
Maximum Capacity
Scaling Policy
Cooldown / Stabilization
```

where supported.

---

# 22. Autoscaling Must Not Overload Dependencies

Scaling application services aggressively can overload:

```text
PostgreSQL
Redis
Kafka / Redpanda
External Providers
```

Cost optimization must therefore consider the entire dependency chain.

---

# 23. Scale Based on Bottlenecks

The platform should optimize the actual bottleneck rather than blindly increasing compute.

Example:

```text
Application CPU Low
Database Saturated
```

Adding more application instances is unlikely to solve the underlying problem.

---

# 24. Container Efficiency

Containers should avoid unnecessary:

```text
Packages
Processes
Background Services
Memory Allocation
```

---

# 25. Database Cost

PostgreSQL is expected to be one of the most important infrastructure cost centers.

Optimization should focus on:

```text
Right-Sizing
Query Efficiency
Index Efficiency
Connection Management
Storage
Read / Write Load
Backup Retention
```

---

# 26. PostgreSQL Right-Sizing

Database capacity should be based on:

```text
CPU
Memory
IOPS
Storage
Connections
Query Latency
Transaction Rate
```

---

# 27. Query Optimization Before Scaling

Before increasing database capacity, investigate:

```text
Slow Queries
Missing Indexes
Inefficient Queries
Excessive Round Trips
Unnecessary Data Retrieval
```

where appropriate.

---

# 28. Index Cost

Indexes improve reads but consume:

```text
Storage
Write Performance
Maintenance Resources
```

Only useful indexes should be retained.

---

# 29. PostgreSQL Connection Cost

Excessive connections can increase:

```text
Memory Usage
Database Load
Connection Overhead
```

Use controlled pooling.

---

# 30. PgBouncer

PgBouncer should be used where database connection management requires it.

This follows:

```text
ADR-0011 — PgBouncer for Database Connection Pooling
```

---

# 31. Connection Pool Sizing

Pool sizes should be based on:

```text
Database Capacity
Service Count
Concurrency
Query Duration
Traffic
```

rather than setting unnecessarily large pools.

---

# 32. PostGIS Cost

Geospatial queries can become expensive at scale.

Optimization should include:

```text
Spatial Indexes
Query Bounding
Appropriate Radius
Selective Filtering
Efficient Geometry Operations
```

---

# 33. Geospatial Query Principle

Do not perform expensive geospatial calculations against unnecessarily large candidate sets.

Prefer:

```text
Filter
→ Reduce Candidate Set
→ Calculate Precise Distance
```

where appropriate.

---

# 34. Database Storage

Storage should be monitored for:

```text
Growth
Unused Data
Indexes
Logs
Historical Data
```

---

# 35. Data Retention

Data should not be retained indefinitely without a business or regulatory reason.

Retention policies may reduce:

```text
Storage
Backup Size
Query Cost
Recovery Time
```

---

# 36. Historical Data

Large historical datasets may eventually be moved to cheaper storage or analytical systems when operational PostgreSQL no longer needs to hold all of the data.

---

# 37. Archival Strategy

Archival should preserve required:

```text
Business Data
Auditability
Legal Requirements
Analytics
```

while reducing operational database load.

---

# 38. Redis Cost

Redis cost is primarily influenced by:

```text
Memory
Instance Size
Replication
Persistence
Connections
Traffic
```

---

# 39. Redis Memory Discipline

Do not store data in Redis indefinitely unless the workload requires it.

Use:

```text
TTL
Eviction
Explicit Deletion
Appropriate Key Design
```

where applicable.

---

# 40. Redis Data Classification

Classify Redis data as:

```text
Cache
Ephemeral State
Real-Time State
Critical State
```

before deciding its persistence and availability requirements.

---

# 41. Cache Cost

A cache should reduce expensive backend work.

If a cache does not provide meaningful value, its infrastructure cost should be reconsidered.

---

# 42. Cache Hit Rate

Monitor:

```text
Hit Rate
Miss Rate
Memory Usage
Evictions
```

to determine whether the cache is correctly sized.

---

# 43. Kafka / Redpanda Cost

Messaging cost is influenced by:

```text
Broker Capacity
Storage
Replication
Partitions
Retention
Network
Producer Volume
Consumer Volume
```

---

# 44. Topic Design

Avoid creating unnecessary topics and partitions.

Topic design should reflect actual event-domain requirements.

---

# 45. Partition Count

Do not over-partition topics without a workload requirement.

Excessive partitions increase operational and infrastructure overhead.

---

# 46. Message Retention

Retention should match:

```text
Replay Requirement
Recovery Requirement
Business Requirement
Storage Cost
```

---

# 47. Event Payload Size

Large event payloads increase:

```text
Storage
Network
Broker Load
Consumer Load
```

Events should contain the information required by consumers without unnecessarily embedding large payloads.

---

# 48. Duplicate Data

Avoid repeatedly publishing large data that consumers can obtain through appropriate references when doing so is safe and consistent with the event contract.

---

# 49. Consumer Efficiency

Consumers should avoid:

```text
Unnecessary Polling
Repeated Processing
Excessive Database Queries
```

---

# 50. Consumer Scaling

Scale consumers based on:

```text
Lag
Processing Time
Partition Count
Business Latency
```

rather than simply increasing consumer count.

---

# 51. Location Infrastructure Cost

Driver location updates can become a major workload.

Cost optimization should consider:

```text
Update Frequency
Payload Size
Storage Strategy
Geospatial Queries
Network Traffic
```

---

# 52. Location Update Frequency

Location update frequency should be appropriate for:

```text
Driver State
Ride State
Dispatch Requirements
ETA Requirements
```

rather than always using the highest possible frequency.

---

# 53. Adaptive Location Updates

Where appropriate, update frequency may vary according to state.

Conceptually:

```text
Offline
→ No Updates

Online / Idle
→ Lower Frequency

Available Near Demand
→ Higher Frequency

Active Ride
→ Higher Frequency
```

The exact policy belongs to the location subsystem.

---

# 54. Location Data Retention

High-frequency location data should not automatically remain in the hot operational database indefinitely.

---

# 55. Location Storage Strategy

The location-storage strategy follows:

```text
ADR-0010 — Driver Location Storage Strategy
```

---

# 56. Network Cost

Network cost can come from:

```text
Internet Egress
Cross-Region Traffic
Service-to-Service Traffic
Database Traffic
Object Storage Transfers
AI Provider Requests
```

---

# 57. Keep Chatty Services Close

Services with high communication volume should preferably operate within low-latency, cost-efficient network boundaries.

---

# 58. Cross-Region Traffic

Avoid unnecessary cross-region traffic.

Cross-region traffic can increase:

```text
Latency
Cost
Failure Complexity
```

---

# 59. Data Locality

Keep data and compute close when possible.

---

# 60. External Provider Traffic

Repeated calls to external providers can create both:

```text
Provider Cost
Network Cost
Latency
```

Caching and batching should be considered where appropriate.

---

# 61. API Call Optimization

Avoid unnecessary external API calls.

Prefer:

```text
Caching
Deduplication
Request Coalescing
Batching
Appropriate Refresh Intervals
```

where provider contracts allow it.

---

# 62. Routing Provider Cost

Routing and distance APIs can become significant costs.

Use them where they provide value and avoid repeated requests for unchanged data.

---

# 63. ETA Cost

ETA prediction should avoid expensive model inference when:

```text
A Cached Prediction Is Still Valid
A Deterministic Estimate Is Sufficient
The Route Has Not Materially Changed
```

where correctness permits.

---

# 64. AI Cost

AI is a major cost-control area.

AI cost may include:

```text
Training
Inference
GPU / CPU
Model Storage
Feature Infrastructure
External AI APIs
Observability
Data Processing
```

---

# 65. AI Model Selection

Use the simplest model that satisfies the business requirement.

Do not use a large model when a smaller model provides sufficient performance.

---

# 66. AI Inference Frequency

Avoid unnecessary real-time inference.

Use:

```text
Batch Prediction
Cached Prediction
Precomputation
```

when real-time inference is not required.

---

# 67. AI Model Caching

Where predictions remain valid for a defined period, caching may reduce inference cost.

---

# 68. AI Fallback

AI failure should use the approved deterministic fallback rather than repeatedly retrying expensive inference without bound.

---

# 69. AI Retry Limits

AI calls must have bounded:

```text
Timeout
Retry Count
Resource Usage
```

---

# 70. AI Provider Cost

For external AI providers, monitor:

```text
Requests
Tokens
Model
Latency
Cost
```

where applicable.

---

# 71. Model Routing

Where multiple models are available, workloads may use:

```text
Small Model
Medium Model
Large Model
```

according to task complexity.

---

# 72. Training Cost

Training should be scheduled and provisioned according to actual need.

Do not keep expensive training infrastructure running continuously when training is periodic.

---

# 73. GPU Cost

GPU resources should be used only when model requirements justify them.

---

# 74. CPU Models

If a model can meet required latency and accuracy using CPU infrastructure, CPU inference may be preferred for cost efficiency.

---

# 75. Model Quantization

Model optimization techniques such as:

```text
Quantization
Distillation
Pruning
```

may be evaluated when they provide meaningful cost or latency benefits without unacceptable quality loss.

---

# 76. Observability Cost

Observability can become a significant cost center.

Costs include:

```text
Logs
Metrics
Traces
Storage
Retention
External Observability Providers
```

---

# 77. Logging Discipline

Do not log:

```text
Large Payloads
Repeated High-Frequency Events
Sensitive Data
Unnecessary Debug Information
```

in production by default.

---

# 78. Log Levels

Production logging should normally use appropriate levels such as:

```text
INFO
WARN
ERROR
```

with debug-level logging enabled only when justified.

---

# 79. High-Frequency Location Logs

Avoid logging every driver location update at full payload detail.

This can create substantial:

```text
Storage
Ingestion
Network
Query
```

cost.

---

# 80. Metrics Cardinality

Avoid unnecessarily high-cardinality metrics labels.

Examples of risky dimensions:

```text
Ride ID
Driver ID
User ID
Request ID
```

used as permanent metric labels.

---

# 81. Trace Sampling

Distributed tracing should use appropriate sampling rather than recording every trace at maximum detail forever.

---

# 82. Retention Strategy

Retention should differ by data type:

```text
Critical Audit Data
→ Longer Retention

Operational Logs
→ Moderate Retention

Debug Data
→ Short Retention

High-Volume Telemetry
→ Controlled Retention
```

---

# 83. Storage Cost

Use appropriate storage tiers where available.

Examples:

```text
Hot Storage
Warm Storage
Cold / Archive Storage
```

---

# 84. Object Storage

Object storage should be preferred for large non-relational artifacts where appropriate.

Examples:

```text
Model Artifacts
Exports
Archived Data
Large Files
Backups
```

---

# 85. Database vs Object Storage

Do not store large binary artifacts in PostgreSQL unless the workload justifies it.

---

# 86. Backup Cost

Backup retention should balance:

```text
Recovery Requirements
Compliance
Storage Cost
```

---

# 87. Backup Redundancy

Additional backup copies should be justified by:

```text
Recovery Risk
Availability Requirements
Disaster Recovery
```

---

# 88. High Availability Cost

High availability introduces additional cost through:

```text
Replication
Standby Capacity
Multi-Zone Infrastructure
Network
Monitoring
```

---

# 89. HA Decision

High availability should be introduced according to business-criticality and SLO requirements.

---

# 90. Multi-Zone Cost

Multi-zone deployment should be used when the availability improvement justifies:

```text
Additional Compute
Network Cost
Operational Complexity
```

---

# 91. Multi-Region Cost

Multi-region architecture should not be adopted merely for theoretical scalability.

It should be justified by:

```text
Availability
Latency
Business Expansion
Regulatory Requirements
Disaster Recovery
```

---

# 92. Multi-Region Data Cost

Cross-region replication can create:

```text
Network Cost
Storage Cost
Operational Complexity
Consistency Complexity
```

---

# 93. Development Resource Scheduling

Development resources that do not need to run continuously should be stopped outside working periods where practical.

---

# 94. Staging Scheduling

Non-production resources may be scheduled to shut down during periods when they are not needed.

---

# 95. Ephemeral Infrastructure

Temporary infrastructure should have automatic cleanup where practical.

---

# 96. Resource Ownership

Every persistent cloud resource should have an owner or owning system.

---

# 97. Resource Tagging

Resources should be tagged or labeled with information such as:

```text
Environment
Service
Owner
Cost Center
Purpose
```

where the cloud platform supports it.

---

# 98. Unused Resource Detection

Periodically identify:

```text
Unused Volumes
Unused IPs
Idle Instances
Unused Databases
Unused Load Balancers
Old Snapshots
Unused Storage
```

and remove them after validation.

---

# 99. Orphaned Resources

Resources left behind by deployments or experiments should be cleaned up.

---

# 100. Temporary Resource Expiration

Temporary environments and resources should have an expiration mechanism where practical.

---

# 101. Reserved / Commitment Pricing

Long-term commitment pricing may be considered only after workload becomes sufficiently predictable.

---

# 102. Commitment Risk

Do not commit to long-term capacity when:

```text
Traffic Is Uncertain
Architecture Is Rapidly Changing
Migration Is Imminent
```

---

# 103. Spot / Preemptible Compute

Spot or preemptible compute may be used for workloads that tolerate interruption.

Suitable examples may include:

```text
Batch Jobs
Model Training
Non-Critical Processing
```

---

# 104. Critical Workloads

Avoid interruption-prone infrastructure for workloads that require continuous availability unless appropriate redundancy exists.

---

# 105. Queue-Based Workloads

Batch and asynchronous workloads should use queue-based execution where it improves:

```text
Resource Utilization
Scaling
Cost
```

---

# 106. Scheduled Workloads

Periodic jobs should run only when needed.

---

# 107. Serverless Components

Serverless infrastructure may be used when its cost and operational model are advantageous for:

```text
Infrequent Workloads
Event-Driven Jobs
Scheduled Jobs
Small APIs
```

---

# 108. Serverless Cost Evaluation

Serverless should not automatically be assumed cheaper.

Evaluate:

```text
Invocation Volume
Duration
Memory
Network
Cold Starts
Operational Complexity
```

---

# 109. Managed vs Self-Managed

The decision should compare:

```text
Infrastructure Price
+
Engineering Cost
+
Operational Risk
```

rather than infrastructure price alone.

---

# 110. Vendor Lock-In Cost

A service that is cheap today may become expensive to migrate later.

Architecture decisions should therefore consider:

```text
Migration Cost
Data Portability
API Coupling
Operational Knowledge
```

---

# 111. Cost and Reliability Trade-Off

Do not optimize cost by removing:

```text
Required Backups
Required Monitoring
Required Security
Required Recovery
Required Capacity
```

---

# 112. Cost and Performance Trade-Off

Do not accept significant customer-facing latency degradation merely to save marginal infrastructure cost.

---

# 113. Cost and Developer Productivity

Developer time is an economic resource.

Avoid infrastructure choices that create excessive operational burden solely for small infrastructure savings.

---

# 114. Cost Optimization Workflow

The preferred workflow is:

```text
Measure
  ↓
Identify Cost Driver
  ↓
Understand Workload
  ↓
Estimate Optimization
  ↓
Assess Risk
  ↓
Implement
  ↓
Measure Again
```

---

# 115. Cost Optimization Without Measurement

Do not optimize a component solely because it appears expensive without understanding:

```text
Traffic
Usage
Business Value
Dependency
```

---

# 116. Cost Dashboards

The platform should provide cost visibility at an appropriate level.

Possible categories:

```text
Compute
Database
Cache
Messaging
Network
Storage
Observability
AI
External Providers
```

---

# 117. Cost Alerts

Cost anomaly alerts should be introduced where cloud tooling supports them.

---

# 118. Cost Budgets

Budgets may be defined for:

```text
Development
Staging
Production
AI
```

according to organizational needs.

---

# 119. Cost Anomaly Detection

Unexpected cost increases should trigger investigation.

Possible causes:

```text
Traffic Spike
Autoscaling
Memory Leak
Log Explosion
Retry Storm
AI Usage Spike
Network Transfer
Resource Leak
```

---

# 120. Retry Storm Cost

Unbounded retries can increase:

```text
Compute
Network
Provider Charges
Queue Load
```

Retries must therefore be bounded.

---

# 121. Failure and Cost

Failure handling should consider cost.

For example:

```text
Provider Failure
→ Unbounded Retry
→ Increased Cost
→ More Failure
```

Use:

```text
Timeout
Backoff
Retry Limit
Circuit Breaker
Fallback
```

where appropriate.

---

# 122. Event Replay Cost

Large event replays can consume significant:

```text
Compute
Database
Messaging
```

capacity.

Replay procedures should be controlled.

---

# 123. DLQ Replay Cost

DLQ replay should be rate-limited where necessary to avoid destabilizing production infrastructure.

---

# 124. Data Processing Cost

Batch pipelines should avoid processing unchanged data unnecessarily.

Prefer incremental processing where appropriate.

---

# 125. Feature Pipeline Cost

AI feature generation should avoid recomputing expensive features unnecessarily.

---

# 126. Feature Reuse

Reusable features should be shared where the architecture supports it.

---

# 127. Model Serving Cost

Model serving capacity should reflect:

```text
Inference Volume
Latency Requirements
Model Size
Concurrency
```

---

# 128. AI Dispatch Cost

Smart dispatch inference should be invoked only after deterministic candidate filtering where practical.

This reduces:

```text
Inference Volume
Feature Computation
Latency
```

---

# 129. Candidate Reduction

The preferred flow is:

```text
Hard Eligibility
→ Geographic Filtering
→ Candidate Reduction
→ Feature Generation
→ AI Ranking
```

rather than running expensive AI inference over every possible driver.

---

# 130. ETA Cost Optimization

ETA should use the cheapest sufficiently accurate source according to context.

Potential hierarchy:

```text
Cached Valid ETA
      ↓
Fast Deterministic Estimate
      ↓
Routing Provider
      ↓
ML Prediction / Hybrid Model
```

The exact hierarchy belongs to the ETA architecture.

---

# 131. Provider Call Deduplication

Identical or near-identical provider requests should be deduplicated where safe.

---

# 132. Provider Rate Limits

External provider limits must be respected.

Cost optimization must never intentionally violate provider contracts.

---

# 133. Database Read Optimization

Where safe, use:

```text
Caching
Read Replicas
Materialized Data
Efficient Queries
```

before simply increasing the primary database size.

---

# 134. Read Replicas

Read replicas may reduce primary database load but introduce:

```text
Additional Cost
Replication Lag
Operational Complexity
```

They should be introduced only when justified.

---

# 135. Caching vs Database Scaling

Caching may be preferred when the workload contains highly repeated reads and acceptable staleness.

---

# 136. Cache Invalidation Cost

Caching should not be introduced when invalidation complexity exceeds the value it provides.

---

# 137. Service Count and Cost

Microservices increase:

```text
Compute Overhead
Networking
Observability
Deployment Complexity
```

Therefore, services should have meaningful boundaries.

---

# 138. Microservice Cost Principle

Do not split a service solely to follow a theoretical microservice count.

---

# 139. Service Scaling Independence

Service separation is economically valuable when workloads scale differently.

Example:

```text
Location Service → High Volume
Admin Service → Low Volume
```

Independent scaling can reduce overall cost.

---

# 140. Event-Driven Cost

Event-driven architecture can improve decoupling but introduces:

```text
Broker Cost
Storage
Consumers
Observability
```

Events should therefore have meaningful business or technical value.

---

# 141. Synchronous vs Asynchronous Cost

Use asynchronous processing when it improves:

```text
Scalability
Resilience
Throughput
Resource Utilization
```

but do not introduce asynchronous infrastructure for trivial workflows where it adds unnecessary cost.

---

# 142. Cost-Aware Architecture Reviews

Major architecture changes should include cost considerations.

---

# 143. Architecture Cost Questions

Before introducing infrastructure, ask:

```text
What problem does this solve?
What workload requires it?
What is the expected cost?
What is the operational cost?
Can an existing component solve it?
Can it be introduced later?
What is the migration cost?
```

---

# 144. Cost Review Triggers

Perform a cost review when:

```text
Traffic Doubles
Infrastructure Cost Spikes
New Region Is Added
New AI Model Is Added
New External Provider Is Added
Major Database Growth Occurs
Messaging Volume Increases
Observability Volume Increases
```

---

# 145. Capacity Planning

Capacity planning should estimate:

```text
Current Load
Expected Growth
Peak Load
Failure Capacity
Scaling Limits
```

---

# 146. Peak Demand

Ride-hailing systems can experience demand spikes.

Capacity planning must consider:

```text
Peak Hours
Events
Weather
Holidays
Regional Demand
```

where relevant.

---

# 147. Capacity Buffer

Production should maintain a reasonable capacity buffer.

Do not run infrastructure permanently at its absolute limit merely to minimize cost.

---

# 148. Capacity vs Cost

The objective is:

```text
Enough Capacity
+
Reasonable Headroom
```

rather than:

```text
Maximum Capacity
```

or:

```text
Minimum Possible Capacity
```

---

# 149. Load Testing for Cost

Load tests can identify:

```text
Cost per Request
Cost per Ride
Cost per Event
```

at different traffic levels.

---

# 150. Cost Regression

Infrastructure changes should be evaluated for unexpected cost regression where practical.

---

# 151. Cost Regression Examples

Examples:

```text
New Logging
→ 10x Log Volume

New Feature
→ 5x Database Queries

New Model
→ 20x Inference Cost

New Consumer
→ 2x Broker Storage
```

---

# 152. Cost Testing

Cost testing does not require perfect prediction.

The objective is to detect material cost changes before they become production problems.

---

# 153. Cost Ownership

Major cost centers should have owners.

Examples:

```text
Database → Platform
Messaging → Platform
AI → AI Engineering
Observability → Platform
External Routing → Mobility / Platform
```

---

# 154. Cost Accountability

Teams should understand how architectural decisions affect infrastructure cost.

---

# 155. Cost Documentation

Significant cost-related decisions should be documented.

---

# 156. Cost Optimization and Security

Do not reduce security controls solely to save cost without an explicit risk review.

---

# 157. Cost Optimization and Compliance

Compliance and legal requirements are not optional cost-reduction targets.

---

# 158. Cost Optimization and Data Retention

Retention reduction may be used when legally and operationally acceptable.

---

# 159. Cost Optimization and Backups

Backup retention may be optimized according to recovery requirements, but required recovery capability must remain intact.

---

# 160. Cost Optimization and Availability

Availability targets define the minimum acceptable infrastructure redundancy.

---

# 161. Cost Optimization and SLOs

Cost decisions should be evaluated against:

```text
Latency SLO
Availability SLO
Recovery SLO
Business SLO
```

---

# 162. Unit Economics

As RideForge grows, monitor:

```text
Infrastructure Cost / Ride
Infrastructure Cost / Completed Ride
Infrastructure Cost / Active Driver
AI Cost / Dispatch
Routing Cost / Ride
Observability Cost / Million Events
```

---

# 163. Cost Per Dispatch

Smart dispatch cost should be measurable.

---

# 164. Cost Per ETA

ETA provider and model cost should be measurable.

---

# 165. Cost Per Location Update

High-volume location processing should be measurable where practical.

---

# 166. Cost Per Event

Messaging infrastructure cost may be estimated using:

```text
Events
Bytes
Retention
Consumers
```

---

# 167. Cost Optimization Priorities

When optimization is necessary, prioritize:

```text
1. Large Waste
2. Idle Resources
3. Excessive Scaling
4. High-Volume Inefficiencies
5. Expensive External Calls
6. AI Inference
7. Observability Overhead
8. Storage Retention
```

The exact priority depends on measured cost.

---

# 168. Cost Optimization Anti-Patterns

Avoid:

```text
Optimizing Without Measurement
Removing Backups
Removing Monitoring
Disabling Security
Over-Compressing Critical Data
Using Unreliable Infrastructure
Premature Reserved Capacity
Premature Multi-Region
Overusing Kubernetes
Using Large AI Models Everywhere
Unlimited Logging
Unlimited Retries
Unlimited Event Retention
Oversized Connection Pools
```

---

# 169. Consequences

## 169.1 Positive Consequences

This strategy provides:

```text
Predictable Cost Growth
Better Resource Utilization
Reduced Waste
Controlled AI Spending
Improved Capacity Planning
Better Infrastructure Decisions
```

---

## 169.2 Negative Consequences

The strategy introduces:

```text
Cost Monitoring Work
Additional Dashboards
Capacity Analysis
Architecture Review Effort
Optimization Engineering
```

These costs are accepted.

---

# 170. Risks

## Risk 1 — Excessive Cost Cutting

### Mitigation

Protect:

```text
Reliability
Security
Recovery
Performance
```

as explicit constraints.

---

## Risk 2 — Under-Provisioning

### Mitigation

Use:

```text
Capacity Planning
Load Testing
Monitoring
Headroom
```

---

## Risk 3 — Cost Complexity

### Mitigation

Track major cost centers rather than attempting perfect accounting for every tiny component.

---

## Risk 4 — AI Cost Explosion

### Mitigation

Use:

```text
Model Selection
Caching
Batching
Candidate Reduction
Inference Limits
Cost Monitoring
```

---

## Risk 5 — Infrastructure Waste

### Mitigation

Use:

```text
Resource Ownership
Tagging
Idle Resource Detection
Automatic Cleanup
```

---

## Risk 6 — Cost Optimization Creates Technical Debt

### Mitigation

Document deliberate trade-offs and revisit them as scale changes.

---

# 171. Alternatives Considered

## 171.1 Optimize for Minimum Possible Cloud Cost

### Advantages

```text
Lower Infrastructure Invoice
```

### Disadvantages

```text
Reliability Risk
Performance Risk
Operational Risk
```

### Decision

```text
Rejected.
```

---

# 172. Optimize for Maximum Performance

### Advantages

```text
Large Capacity
Low Latency Headroom
```

### Disadvantages

```text
Excessive Cost
Unused Capacity
```

### Decision

```text
Rejected.
```

---

# 173. Fixed Infrastructure Forever

### Advantages

```text
Predictable
Simple
```

### Disadvantages

```text
Poor Scaling
Over-Provisioning
Under-Provisioning
```

### Decision

```text
Rejected.
```

---

# 174. Fully Serverless Cost Strategy

### Advantages

```text
Potentially Low Idle Cost
Automatic Scaling
```

### Disadvantages

```text
Variable Cost
Runtime Constraints
Potentially Higher Cost at Sustained Scale
```

### Decision

```text
Rejected as a universal strategy.
```

---

# 175. Validation

This ADR should be validated through:

```text
Cost Dashboards
Cloud Billing Analysis
Resource Utilization
Load Tests
Capacity Tests
AI Usage Monitoring
Database Monitoring
Messaging Monitoring
Network Analysis
Storage Analysis
Cost Anomaly Detection
Architecture Reviews
```

---

# 176. Review Triggers

Revisit this ADR when:

```text
Monthly Infrastructure Cost Changes Materially
Traffic Increases Significantly
New Cloud Provider Is Added
New Region Is Added
AI Usage Becomes Significant
Database Growth Accelerates
Messaging Volume Changes
Observability Cost Becomes Material
A Major Infrastructure Component Is Replaced
Business Unit Economics Change
```

---

# 177. Final Principles

The following principles are mandatory:

```text
1. Measure infrastructure cost before optimizing it.

2. Optimize based on workload rather than assumptions.

3. Infrastructure cost must be considered as part of total cost of ownership.

4. Engineering and operational complexity are part of cost.

5. Production reliability must not be sacrificed for marginal cost savings.

6. Security must not be removed merely to reduce infrastructure cost.

7. Required backups must remain enabled.

8. Required observability must remain enabled.

9. Required disaster recovery capability must remain available.

10. Development environments should be cost-efficient.

11. Local containers should be preferred where they reduce unnecessary cloud spending.

12. Testing infrastructure should be ephemeral or right-sized where practical.

13. Staging should be production-relevant but right-sized.

14. Production capacity must be based on real workload and business requirements.

15. Compute resources should be right-sized.

16. Autoscaling should be used where workload variability justifies it.

17. Autoscaling must consider downstream dependencies.

18. Database capacity must be optimized using query and workload analysis before blindly scaling infrastructure.

19. PostgreSQL indexes must provide measurable value.

20. Database connection pools must remain controlled.

21. PgBouncer should be used where connection management requires it.

22. PostGIS queries should reduce candidate sets before expensive calculations where practical.

23. Redis data must have appropriate lifecycle management.

24. Redis memory should not be consumed by indefinitely retained data without justification.

25. Kafka / Redpanda partitions should not be created unnecessarily.

26. Event retention must balance replay requirements and storage cost.

27. Event payloads should avoid unnecessary size.

28. High-frequency driver location traffic must be optimized.

29. Location update frequency should match operational requirements.

30. High-volume location data should not remain in hot storage indefinitely without justification.

31. Cross-region traffic should be minimized where practical.

32. External provider calls should be minimized through caching and deduplication where safe.

33. Routing provider usage must be monitored.

34. AI inference must be cost-aware.

35. The simplest model that meets requirements is preferred.

36. Expensive models should not be used for simple tasks without justification.

37. Real-time AI inference should be used only when real-time prediction is required.

38. AI predictions may be cached when validity permits.

39. AI retries must be bounded.

40. AI training infrastructure should not remain active when training is not running.

41. GPU resources should be used only when justified.

42. Model optimization may be used when it reduces cost without unacceptable quality loss.

43. Observability data must be retained according to business value and operational requirements.

44. High-frequency events should not generate unnecessary logs.

45. High-cardinality identifiers should not be used carelessly as metric labels.

46. Trace sampling should be used where full tracing is unnecessary.

47. Large artifacts should use appropriate object storage where practical.

48. Data retention must have a business, operational, or legal justification.

49. Unused resources should be identified and removed.

50. Temporary resources should have automatic cleanup where practical.

51. Persistent resources should have clear ownership.

52. Cloud resources should be tagged or labeled where supported.

53. Long-term capacity commitments should be used only when workload is sufficiently predictable.

54. Spot or preemptible compute may be used for interruption-tolerant workloads.

55. Critical workloads should not depend on interruption-prone capacity without sufficient redundancy.

56. Cost anomalies must be investigated.

57. Unbounded retries must be avoided.

58. Event replay must be controlled.

59. DLQ replay must be controlled.

60. Batch pipelines should avoid unnecessary recomputation.

61. Feature pipelines should avoid unnecessary expensive recomputation.

62. Microservice boundaries should provide meaningful operational or business value.

63. Microservices should not be created merely to increase service count.

64. Independent service scaling should be used where it provides real efficiency.

65. Infrastructure should be evaluated by total cost of ownership, not cloud invoice alone.

66. Cloud-provider lock-in should be considered in major cost decisions.

67. Read replicas should be introduced only when their cost and consistency trade-offs are justified.

68. Multi-zone redundancy should match availability requirements.

69. Multi-region infrastructure should not be introduced prematurely.

70. Cost decisions must consider latency and customer experience.

71. Cost decisions must consider driver and customer safety.

72. Cost decisions must consider legal and regulatory requirements.

73. Cost decisions must consider RTO and RPO.

74. Cost optimization should follow a measure → optimize → measure cycle.

75. Material cost changes should be reviewed.

76. Infrastructure owners should understand the cost impact of their systems.

77. Cost should be visible by major system area.

78. Unit economics should become increasingly important as RideForge scales.

79. Cost per ride should eventually become a meaningful operational metric.

80. Cost per AI prediction should be monitored when AI becomes material.

81. Cost per location update should be monitored when location processing becomes material.

82. Cost per event should be monitored when messaging becomes material.

83. Cost optimization must not create hidden reliability debt.

84. Cost optimization must not create hidden operational debt.

85. Cost optimization must remain aligned with the cloud and deployment strategy.

86. The platform should continuously evolve toward the lowest sustainable cost that satisfies its required reliability, performance, security, and business objectives.
```

---

# 178. Status

```text
Decision: ACCEPTED

Strategy:
Measured + Workload-Driven + Risk-Aware + Incremental

Primary Goal:
Optimize Total Cost of Ownership Without Sacrificing Required Reliability, Security, Performance, Recovery, Legal Compliance, or Core Ride-Hailing Functionality

Cost Areas:
Compute
Database
Redis
Kafka / Redpanda
Storage
Networking
Observability
AI / ML
External Providers
CI/CD
Development
Staging
Production

Compute:
Right-Sized + Autoscaled Where Appropriate

Database:
Query Optimization + Right-Sizing + Controlled Connections

PostgreSQL:
Primary Database

PostGIS:
Geospatial Optimization Required

PgBouncer:
Used Where Connection Management Requires It

Redis:
TTL + Memory Discipline + Right-Sizing

Kafka / Redpanda:
Controlled Partitions + Retention + Payload Size

Location:
Adaptive / Requirement-Based Update Frequency

Networking:
Minimize Unnecessary Cross-Region and External Traffic

AI:
Cost-Aware Model Selection + Controlled Inference

Observability:
Controlled Logging + Metrics + Tracing Retention

Storage:
Appropriate Storage Tier + Retention

Development:
Cost-Efficient / Local First Where Practical

Testing:
Ephemeral / Right-Sized Where Practical

Staging:
Production-Relevant but Right-Sized

Production:
Reliability-First + Right-Sized

Autoscaling:
Allowed With Downstream Capacity Controls

Multi-Region:
Requirement-Driven

High Availability:
Business-Requirement-Driven

Cost Visibility:
Required

Cost Anomaly Detection:
Recommended

Unit Economics:
Required as Scale Increases

Cloud Strategy:
ADR-0027

Security:
ADR-0023

Configuration:
ADR-0024

Testing:
ADR-0025

AI Governance:
ADR-0026

Failure Strategy:
ADR-0021

Observability:
ADR-0022

Next ADR:
0029-ARCHITECTURE_EVOLUTION_AND_MIGRATION.md
```

---

# 179. Decision Summary

RideForge adopts the following cost optimization loop:

```text
                    ACTUAL WORKLOAD
                           │
                           ▼
                    COST VISIBILITY
                           │
                           ▼
                  IDENTIFY COST DRIVER
                           │
                           ▼
                  UNDERSTAND BOTTLENECK
                           │
                           ▼
                 EVALUATE OPTIMIZATION
                           │
              ┌────────────┴────────────┐
              ▼                         ▼
        Cost Reduction             Risk Increase
              │                         │
              ▼                         ▼
        Implement Safely          Reject / Redesign
              │
              ▼
             TEST
              │
              ▼
          MEASURE AGAIN
              │
              ▼
        CONTINUOUS OPTIMIZATION
```

The architecture intentionally avoids two extremes:

```text
CHEAP AT ANY COST
```

and:

```text
SCALE AT ANY COST
```

Instead, RideForge aims for:

```text
RELIABLE
+
PERFORMANT
+
SECURE
+
COST-EFFICIENT
+
OPERATIONALLY SIMPLE
```

The central principle is:

> **Do not spend money to solve a problem that does not exist, and do not remove infrastructure that is required to operate RideForge safely and reliably.**

---

# 180. Status Metadata

| Field | Value |
|---|---|
| ADR | `0028` |
| Title | Cost Optimization Strategy |
| Status | Accepted |
| Category | Cost / Infrastructure / Architecture |
| Primary Strategy | Measured + Workload-Driven |
| Cost Model | Total Cost of Ownership |
| Unit Economics | Increasingly Required With Scale |
| Compute | Right-Sizing + Selective Autoscaling |
| Database | Right-Sizing + Query Optimization |
| PostgreSQL | Primary Database |
| PostGIS | Geospatial Optimization |
| PgBouncer | Connection Cost / Capacity Control |
| Redis | TTL + Memory Discipline |
| Kafka / Redpanda | Partition + Retention + Payload Control |
| Location | Adaptive Workload Optimization |
| Networking | Minimize Unnecessary Transfer |
| AI | Cost-Aware Model / Inference Strategy |
| Observability | Controlled Volume + Retention |
| Storage | Lifecycle + Appropriate Tier |
| Development | Cost-Efficient |
| Testing | Ephemeral / Right-Sized |
| Staging | Right-Sized |
| Production | Reliability-First |
| Multi-Region | Requirement-Driven |
| Cost Visibility | Required |
| Cost Anomaly Detection | Recommended |
| Cloud / Deployment | ADR-0027 |
| Security | ADR-0023 |
| Configuration | ADR-0024 |
| Testing | ADR-0025 |
| AI Governance | ADR-0026 |
| Failure Strategy | ADR-0021 |
| Observability | ADR-0022 |
| Next ADR | `0029-ARCHITECTURE_EVOLUTION_AND_MIGRATION.md` |
