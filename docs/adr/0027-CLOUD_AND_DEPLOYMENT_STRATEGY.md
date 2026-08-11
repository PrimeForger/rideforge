# ADR-0027: Cloud and Deployment Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Cloud / Infrastructure / Deployment / Operations  
> **Scope:** Cloud infrastructure, environments, containerization, orchestration, service deployment, networking, databases, messaging, caching, observability, CI/CD, scaling, availability, disaster recovery, regional deployment, cost control, and infrastructure evolution  
> **Owner:** RideForge Architecture / Platform Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is a distributed ride-hailing platform designed around:

```text
Microservices
Domain-Driven Design
Event-Driven Architecture
PostgreSQL
PostGIS
Redis
Kafka / Redpanda
HTTP APIs
Real-Time Location
Smart Dispatch
Stand Dispatch
AI / ML Services
External Providers
```

The platform must run reliably across multiple environments and eventually support production workloads with:

```text
Horizontal Scaling
Service Isolation
Rolling Deployments
Failure Recovery
Observability
Security
Regional Operation
Cost Control
```

RideForge also has an explicit requirement to avoid unnecessary infrastructure complexity during the early stages of the platform.

The deployment architecture therefore needs to balance:

```text
Reliability
Scalability
Operational Simplicity
Cost
Developer Productivity
Security
Future Migration
```

---

# 2. Problem

A ride-hailing system cannot rely on a single manually configured server once it reaches meaningful production scale.

At the same time, adopting a highly complex cloud-native platform too early creates unnecessary:

```text
Infrastructure Cost
Operational Burden
Deployment Complexity
Debugging Complexity
Engineering Overhead
```

RideForge therefore requires a deployment strategy that:

```text
Starts Simple
Scales Incrementally
Preserves Architectural Boundaries
Supports Automation
Allows Infrastructure Evolution
```

---

# 3. Decision

RideForge will use a:

```text
Containerized
Automated
Environment-Separated
Cloud-Deployable
Incrementally Scalable
Infrastructure-as-Code-Friendly
```

deployment strategy.

The platform will initially optimize for:

```text
Production Reliability
Low Operational Complexity
Controlled Cost
Fast Deployment
Clear Service Boundaries
```

and will adopt more advanced orchestration only when operational requirements justify it.

---

# 4. Core Deployment Principles

RideForge follows:

```text
Build Once, Deploy Many
Immutable Artifacts
Environment Separation
Automated Deployment
Infrastructure as Code
Least Privilege
Horizontal Scaling Where Appropriate
Stateless Services Where Practical
Managed Infrastructure Where Cost-Effective
Observable Deployments
Rollback Capability
```

---

# 5. Deployment Architecture

The conceptual deployment architecture is:

```text
                         Internet
                            │
                            ▼
                     Load Balancer
                            │
                    ┌───────┴───────┐
                    ▼               ▼
               API Gateway      Web / Client
                    │
                    ▼
             Application Services
                    │
        ┌───────────┼────────────┐
        ▼           ▼            ▼
   PostgreSQL     Redis     Kafka / Redpanda
        │           │            │
        └───────────┼────────────┘
                    │
                    ▼
             External Providers
```

AI services may operate as separate deployable components:

```text
Application Service
       │
       ▼
AI / Inference Service
       │
       ▼
Model Runtime
```

---

# 6. Environment Strategy

The platform will maintain distinct environments:

```text
Development
Testing
Staging
Production
```

These environments follow:

```text
ADR-0024 — Configuration and Environment Strategy
```

---

# 7. Development Environment

Development prioritizes:

```text
Fast Feedback
Low Cost
Easy Debugging
Reproducibility
```

Local infrastructure may use containers.

---

# 8. Local Development

Docker Compose or equivalent tooling may provide:

```text
PostgreSQL
PostGIS
Redis
Kafka / Redpanda
Application Services
Supporting Dependencies
```

where appropriate.

---

# 9. Development Environment Isolation

Local development must not accidentally connect to production resources.

---

# 10. Testing Environment

Testing environments exist to validate:

```text
Application Behaviour
Integration
Database
Messaging
Configuration
Deployment
```

---

# 11. Staging Environment

Staging should approximate production sufficiently to validate:

```text
Deployment
Networking
Configuration
Scaling
Observability
Database
Messaging
External Provider Integration
```

---

# 12. Production Environment

Production prioritizes:

```text
Availability
Security
Performance
Observability
Recovery
Controlled Changes
```

---

# 13. Cloud Strategy

RideForge will use cloud infrastructure where it provides meaningful operational value.

The architecture should remain portable enough to avoid unnecessary coupling to a single cloud provider.

---

# 14. Cloud Provider Independence

Application services should avoid embedding cloud-provider-specific assumptions in domain logic.

Cloud-specific integrations belong in:

```text
Infrastructure Layer
Platform Layer
Provider Adapters
Deployment Configuration
```

---

# 15. Managed Services

Managed cloud services should be preferred where they materially reduce operational burden without creating unacceptable cost or architectural lock-in.

Examples:

```text
Managed PostgreSQL
Managed Redis
Managed Object Storage
Managed Load Balancer
Managed Container Runtime
```

---

# 16. Self-Managed Infrastructure

Self-managed infrastructure may be used when:

```text
Cost Benefit Is Significant
Operational Requirements Are Known
Team Can Operate It Reliably
Managed Alternative Is Impractical
```

---

# 17. Database Deployment

PostgreSQL remains the primary database.

This follows:

```text
ADR-0007 — PostgreSQL as Primary Database
```

---

# 18. PostgreSQL Production Strategy

Production PostgreSQL should prioritize:

```text
Durability
Backups
Monitoring
Connection Management
Recovery
Capacity Planning
```

---

# 19. PostGIS

PostGIS remains part of the PostgreSQL deployment where geospatial operations require it.

This follows:

```text
ADR-0008 — PostGIS for Geospatial Operations
```

---

# 20. Database Connection Pooling

Database connections should be controlled through:

```text
Application Pooling
PgBouncer
Database Capacity
```

where appropriate.

This follows:

```text
ADR-0011 — PgBouncer for Database Connection Pooling
```

---

# 21. Database Deployment Principle

Application services should not each independently consume unlimited database connections.

Aggregate connection capacity must be considered.

---

# 22. Database Backups

Production PostgreSQL must have automated backups appropriate to the required recovery objectives.

---

# 23. Backup Validation

Backups are not considered reliable until restoration has been tested.

---

# 24. Database Recovery

Recovery procedures should define:

```text
Restore Source
Restore Procedure
Expected Recovery Time
Data Loss Expectation
Validation
```

---

# 25. Database High Availability

High availability should be introduced when:

```text
Traffic
Business Criticality
Recovery Requirements
```

justify the additional cost and complexity.

---

# 26. Redis Deployment

Redis is used for:

```text
Caching
Real-Time State
Location-Related Workloads
Operational Coordination
```

according to:

```text
ADR-0009 — Redis for Real-Time State and Caching
```

---

# 27. Redis Durability

Redis data should be classified according to whether it is:

```text
Recoverable Cache
Critical Runtime State
Ephemeral State
```

Deployment durability should match that classification.

---

# 28. Redis Failure

Services must not assume Redis is permanently available.

Critical workflows should follow the approved degradation strategy.

---

# 29. Messaging Deployment

Kafka / Redpanda is the event-streaming infrastructure.

This follows:

```text
ADR-0006 — Kafka / Redpanda for Event Streaming
```

---

# 30. Messaging Production Strategy

Production messaging infrastructure should provide:

```text
Durability
Monitoring
Access Control
Capacity Planning
Consumer Lag Monitoring
Recovery
```

---

# 31. Topic Management

Topics should be managed through controlled infrastructure or provisioning automation.

---

# 32. Consumer Groups

Consumer group configuration must be deterministic and deployment-safe.

---

# 33. Messaging Scaling

Consumers should scale independently where workload patterns require it.

---

# 34. Containerization

RideForge services should be deployable as containers.

Benefits include:

```text
Reproducibility
Isolation
Portability
Consistent Runtime
Simplified Deployment
```

---

# 35. Container Images

Images should be:

```text
Minimal
Versioned
Reproducible
Secure
Scannable
```

---

# 36. Immutable Images

Production images should not be modified after build.

---

# 37. Image Tagging

Avoid relying exclusively on:

```text
latest
```

for production deployments.

Prefer immutable identifiers such as:

```text
Version
Git Commit SHA
Build ID
```

---

# 38. Image Registry

Production images should be stored in a controlled container registry.

---

# 39. Image Security

Images should be scanned for:

```text
Known Vulnerabilities
Outdated Dependencies
Unnecessary Packages
```

where tooling supports this.

---

# 40. Container Runtime

The initial production container runtime should be selected based on:

```text
Traffic
Operational Capability
Cost
Scaling Requirements
```

rather than adopting orchestration complexity by default.

---

# 41. Orchestration Strategy

RideForge supports eventual orchestration through platforms such as:

```text
Kubernetes
Managed Container Platforms
Equivalent Orchestration Systems
```

but orchestration complexity should be introduced only when justified.

---

# 42. Kubernetes Strategy

Kubernetes may be used when the platform requires:

```text
Many Services
Complex Scheduling
Independent Scaling
Advanced Networking
Multi-Region Workloads
High Deployment Frequency
```

---

# 43. Kubernetes Is Not a Requirement From Day One

The architecture should not require Kubernetes merely because RideForge uses microservices.

---

# 44. Initial Deployment Simplicity

Early production deployment may use:

```text
Managed Container Runtime
Virtual Machines
Container Services
```

provided that:

```text
Deployment
Monitoring
Scaling
Security
Recovery
```

are sufficiently supported.

---

# 45. Service Statelessness

Application services should remain stateless where practical.

State should reside in appropriate infrastructure:

```text
PostgreSQL
Redis
Kafka / Redpanda
Object Storage
```

---

# 46. Stateful Services

Stateful infrastructure should be treated separately from stateless application services.

---

# 47. Horizontal Scaling

Stateless services should support horizontal scaling.

Conceptually:

```text
                 Load Balancer
                      │
          ┌───────────┼───────────┐
          ▼           ▼           ▼
      Instance A  Instance B  Instance C
```

---

# 48. Vertical Scaling

Vertical scaling may be used when:

```text
Horizontal Scaling Is Not Efficient
Stateful Infrastructure Requires It
Traffic Is Small
Cost Is Lower
```

---

# 49. Scaling Strategy

Scaling should be driven by:

```text
CPU
Memory
Request Rate
Latency
Queue Lag
Database Capacity
Business Traffic
```

rather than CPU alone.

---

# 50. Autoscaling

Autoscaling should be introduced where workload variability justifies it.

---

# 51. Autoscaling Safety

Autoscaling must consider downstream capacity.

Scaling application instances without scaling database capacity can cause:

```text
Connection Exhaustion
Database Overload
Latency
Failure
```

---

# 52. Database-Aware Scaling

Service scaling must account for:

```text
PgBouncer
PostgreSQL Connections
Query Load
Transaction Rate
```

---

# 53. Kafka-Aware Scaling

Consumer scaling must consider:

```text
Partition Count
Consumer Group
Processing Rate
Lag
```

---

# 54. Redis-Aware Scaling

Redis workloads must consider:

```text
Memory
Connection Count
Command Rate
Latency
Key Distribution
```

---

# 55. Load Balancing

Production HTTP services should be deployed behind an appropriate load-balancing layer.

---

# 56. Health Checks

Services should expose health signals appropriate for deployment systems.

Conceptually:

```text
Liveness
Readiness
Startup
```

where the runtime supports these concepts.

---

# 57. Liveness

Liveness should indicate whether the process is functioning sufficiently to remain running.

It should not depend on every external dependency being healthy.

---

# 58. Readiness

Readiness should indicate whether the service can safely receive traffic.

---

# 59. Startup

Startup checks may be used when initialization requires time.

---

# 60. Dependency Health

A service may be alive while a dependency is unavailable.

Health checks should distinguish:

```text
Process Health
Dependency Health
Traffic Readiness
```

---

# 61. Deployment Strategy

RideForge will prefer controlled deployments such as:

```text
Rolling Deployment
Canary Deployment
Blue-Green Deployment
```

depending on service risk and infrastructure capabilities.

---

# 62. Rolling Deployment

Rolling deployment should be the default for services where compatible versions can coexist.

---

# 63. Canary Deployment

Canary deployment should be used for higher-risk changes where traffic can be gradually shifted.

---

# 64. Blue-Green Deployment

Blue-green deployment may be used when:

```text
Fast Rollback
Environment Duplication
```

justify its cost.

---

# 65. Database Compatibility During Deployment

Application versions must be compatible with database schema transitions during rolling deployments.

Use:

```text
Expand
→ Migrate
→ Deploy
→ Contract
```

where necessary.

---

# 66. Event Compatibility During Deployment

Event producers and consumers may temporarily run different versions.

Event schema evolution must preserve compatibility according to the event strategy.

---

# 67. Deployment Ordering

For changes with dependency relationships, deployment order must be explicit.

Example:

```text
Infrastructure
→ Database Expansion
→ Compatible Application
→ Consumer
→ Producer
→ Cleanup
```

---

# 68. Build Once, Deploy Many

The same immutable application artifact should be promotable between environments where practical.

Conceptually:

```text
Build
 ↓
Test
 ↓
Staging
 ↓
Production
```

without rebuilding solely to embed environment-specific values.

---

# 69. Runtime Configuration

Environment-specific values should be injected at runtime.

This follows:

```text
ADR-0024 — Configuration and Environment Strategy
```

---

# 70. Secrets

Secrets must be provided through approved secret-management mechanisms.

This follows:

```text
ADR-0023 — Security and Secret Management
```

---

# 71. CI/CD

Deployment should be automated through CI/CD where practical.

Pipeline stages may include:

```text
Source
 ↓
Build
 ↓
Test
 ↓
Security Scan
 ↓
Image Build
 ↓
Image Scan
 ↓
Registry
 ↓
Staging
 ↓
Validation
 ↓
Production
```

---

# 72. Deployment Authorization

Production deployment must require appropriate authorization.

---

# 73. Deployment Auditability

Deployment systems should record:

```text
Artifact
Commit
Environment
Timestamp
Actor / Automation
Result
```

---

# 74. Manual Deployment

Manual production deployment should not be the normal operating model.

Emergency manual procedures may exist with appropriate controls.

---

# 75. Rollback

Every production deployment must have a rollback strategy.

---

# 76. Application Rollback

Rollback should normally restore a previously known-good artifact.

---

# 77. Database Rollback

Database rollback must not rely on blindly reversing migrations.

Prefer backward-compatible migration strategies where possible.

---

# 78. Event Rollback

Event schema changes should be designed so that deployment rollback does not create incompatible event streams.

---

# 79. Configuration Rollback

Configuration rollback follows:

```text
ADR-0024
```

---

# 80. AI Rollback

Model rollback follows:

```text
ADR-0026
```

---

# 81. Zero-Downtime Goal

Production services should aim for zero or minimal downtime during normal deployments.

---

# 82. Graceful Shutdown

Services must support graceful shutdown where possible.

Graceful shutdown should allow:

```text
Stop New Work
Finish Safe Work
Commit State
Flush Events
Close Connections
Exit
```

---

# 83. Event Consumer Shutdown

Consumers should stop accepting new work before terminating and should avoid losing acknowledged processing state.

---

# 84. HTTP Shutdown

HTTP services should stop accepting new requests according to the runtime's graceful shutdown mechanism and finish active requests within a bounded period.

---

# 85. Deployment and WebSockets

Real-time services using:

```text
WebSockets
WebRTC
Long-Lived Connections
```

require deployment strategies that account for connection draining.

---

# 86. Real-Time Location Services

Driver location services may require special handling during deployment because they can receive high-frequency updates.

---

# 87. Location Service Scaling

Location ingestion should scale independently where traffic requires it.

---

# 88. Dispatch Service Scaling

Dispatch services should be horizontally scalable where state ownership permits it.

---

# 89. Distributed Coordination

If multiple service instances participate in the same workflow, coordination should use approved infrastructure rather than process-local memory.

---

# 90. Process-Local State

Avoid storing critical shared state only in:

```text
In-Memory Maps
Local Files
Process Memory
```

---

# 91. Temporary Local State

Process-local caching may be used for non-critical data when invalidation and consistency requirements permit it.

---

# 92. Network Architecture

Production networking should separate:

```text
Public Entry Points
Private Services
Data Infrastructure
Management Interfaces
```

where supported by the deployment environment.

---

# 93. Public Exposure

Only services that need public exposure should be publicly reachable.

---

# 94. Internal Services

Internal microservices should communicate through controlled private networking where practical.

---

# 95. Database Exposure

Production databases should not be publicly accessible unless there is an explicitly justified requirement.

---

# 96. Redis Exposure

Production Redis should remain on private networking where practical.

---

# 97. Kafka Exposure

Kafka / Redpanda brokers should use controlled network access and authentication.

---

# 98. Network Security

Use appropriate:

```text
Firewall Rules
Security Groups
Network Policies
Private Networking
TLS
```

depending on the platform.

---

# 99. TLS

Production external traffic should use TLS.

Internal traffic should use appropriate encryption based on threat model and infrastructure capabilities.

---

# 100. Service Authentication

Internal services should authenticate where required by the security architecture.

---

# 101. Identity and Access

Deployment identities should use least privilege.

---

# 102. Deployment Credentials

CI/CD should not use broad permanent credentials when short-lived or scoped credentials are available.

---

# 103. Infrastructure as Code

Production infrastructure should be represented as code where practical.

Potential tooling includes:

```text
Terraform
Pulumi
Cloud Provider IaC
Kubernetes Manifests
Helm
```

The exact tooling is not fixed by this ADR.

---

# 104. Infrastructure Review

Infrastructure changes should go through code review.

---

# 105. Infrastructure State

Infrastructure state must be stored safely and protected against accidental corruption.

---

# 106. Infrastructure Drift

Actual infrastructure should be compared with declared infrastructure where tooling supports it.

---

# 107. Environment Reproducibility

A new environment should be reconstructable from:

```text
Infrastructure Code
Configuration
Secret References
Application Artifacts
Database Migrations
```

subject to external provider dependencies.

---

# 108. Deployment Dependency Management

Deployment should make dependencies explicit.

Example:

```text
Network
 ↓
Database / Redis / Messaging
 ↓
Core Services
 ↓
Dependent Services
 ↓
API / Gateway
```

---

# 109. Database Migration Deployment

Database migrations should run through controlled automation.

---

# 110. Migration Safety

Destructive migrations should not be deployed in the same step as application code that still requires the removed structure.

---

# 111. Backup Before High-Risk Migration

High-risk database changes should have validated recovery mechanisms before execution.

---

# 112. Messaging Infrastructure Deployment

Messaging infrastructure changes should consider:

```text
Topics
Partitions
Retention
Consumers
Producers
Compatibility
```

---

# 113. Redis Infrastructure Deployment

Redis changes should consider:

```text
Memory
Persistence
Eviction
Connection Limits
Availability
```

---

# 114. Observability Deployment

Every production service should integrate with the platform observability stack.

This follows:

```text
ADR-0022 — Observability Strategy
```

---

# 115. Deployment Metrics

Monitor:

```text
Deployment Success
Deployment Duration
Rollback Rate
Startup Time
Health Check Failures
```

---

# 116. Runtime Metrics

Monitor:

```text
CPU
Memory
Network
Requests
Latency
Errors
Connections
Queue Lag
```

---

# 117. Business Metrics

Infrastructure monitoring must also consider:

```text
Ride Creation Success
Dispatch Success
Driver Assignment
Customer Wait
Cancellation
Payment Success
```

---

# 118. Alerting

Alerts should focus on actionable conditions.

Avoid alerting on every minor infrastructure fluctuation.

---

# 119. Incident Integration

Deployment and infrastructure alerts should integrate with the incident-management process.

---

# 120. Logging

Production logs should be:

```text
Structured
Centralized
Searchable
Correlated
Redacted
```

where the platform supports these capabilities.

---

# 121. Distributed Tracing

Distributed tracing should be used for important cross-service workflows where practical.

Example:

```text
Ride Request
→ Ride Service
→ Dispatch
→ Matching
→ Driver
```

---

# 122. Correlation

Requests and events should have correlation identifiers where appropriate.

---

# 123. Disaster Recovery

Production infrastructure must have a documented disaster-recovery strategy.

---

# 124. Recovery Objectives

The platform should define:

```text
RTO
RPO
```

according to business requirements.

---

# 125. Recovery Testing

Disaster recovery procedures should be tested periodically.

---

# 126. Regional Recovery

If the platform eventually operates across multiple regions, recovery strategy should account for:

```text
Regional Failure
Network Partition
Database Failure
Provider Failure
```

---

# 127. Multi-Region Strategy

Multi-region deployment is not required from the first production release.

It should be introduced when justified by:

```text
Business Scale
Availability Requirements
Latency
Regulatory Requirements
Disaster Recovery
```

---

# 128. Regional Deployment

When multiple regions are introduced, the platform must clearly define:

```text
Data Ownership
Traffic Routing
Service Placement
Database Strategy
Messaging Strategy
Failover
```

---

# 129. Data Locality

Regional deployment must consider applicable data residency and privacy requirements.

---

# 130. Cross-Region Ride Operations

RideForge has regional and legal operating constraints.

Deployment architecture must not assume that geographic proximity automatically permits cross-region ride operations.

---

# 131. Regional Legal Validation

Legal ride eligibility remains an application/domain concern.

Infrastructure routing must not bypass it.

---

# 132. Region-Aware Services

Services may require awareness of:

```text
Region
Timezone
Operating Zone
Provider
Dispatch Strategy
```

through approved configuration.

---

# 133. Regional Dispatch

Regional operations may select:

```text
Smart Dispatch
Stand Dispatch
```

according to approved operational policy.

---

# 134. Regional Failure

If one regional dependency fails, the system should use only approved failover behaviour.

Do not automatically redirect rides across legal or operational boundaries.

---

# 135. Cloud Cost Strategy

Infrastructure should be sized according to actual demand.

Avoid:

```text
Over-Provisioning
Unused Managed Services
Premature Multi-Region
Unnecessary Kubernetes Complexity
```

---

# 136. Cost Visibility

Track major cost sources:

```text
Compute
Database
Redis
Kafka / Redpanda
Network
Storage
AI Inference
Observability
```

---

# 137. Cost-Aware Scaling

Autoscaling policies should consider both:

```text
Performance
Cost
```

---

# 138. Development Cost

Local environments should use cost-efficient infrastructure.

---

# 139. Staging Cost

Staging should approximate production where necessary but may use smaller capacity.

---

# 140. Production Cost

Production capacity should be based on:

```text
Traffic
SLO
Capacity
Growth
Recovery Requirements
```

---

# 141. Cost Optimization ADR

Detailed cost optimization follows:

```text
ADR-0028 — Cost Optimization Strategy
```

---

# 142. Security

Cloud and deployment security must follow:

```text
ADR-0023 — Security and Secret Management
```

---

# 143. Configuration

Deployment configuration follows:

```text
ADR-0024 — Configuration and Environment Strategy
```

---

# 144. Testing

Deployment validation follows:

```text
ADR-0025 — Testing and Integration Strategy
```

---

# 145. AI Deployment

AI model deployment follows:

```text
ADR-0026 — Model and AI Governance
```

---

# 146. Failure Strategy

Deployment and infrastructure failure handling follows:

```text
ADR-0021 — Failure and Degradation Strategy
```

---

# 147. Architecture Evolution

Infrastructure migration and long-term evolution follow:

```text
ADR-0029 — Architecture Evolution and Migration
```

---

# 148. Deployment Maturity Model

RideForge may evolve through:

```text
Stage 1
Simple Container Deployment

        ↓

Stage 2
Automated CI/CD + Managed Infrastructure

        ↓

Stage 3
Independent Service Scaling

        ↓

Stage 4
Advanced Orchestration

        ↓

Stage 5
Multi-Region / Advanced Resilience
```

---

# 149. Stage 1 — Early Production

Characteristics:

```text
Containerized Services
Managed / Simple Infrastructure
Automated Build
Basic Deployment Automation
Centralized Observability
Backups
```

---

# 150. Stage 2 — Growing Production

Introduce:

```text
Independent Scaling
Autoscaling
Stronger CI/CD
Advanced Monitoring
Canary Deployments
Capacity Automation
```

---

# 151. Stage 3 — High Scale

Potentially introduce:

```text
Kubernetes
Advanced Scheduling
Multiple Service Pools
Advanced Event Scaling
Regional Infrastructure
```

---

# 152. Stage 4 — Multi-Region

Potentially introduce:

```text
Regional Service Deployment
Regional Data Strategy
Traffic Steering
Cross-Region Recovery
Regional Observability
```

---

# 153. Stage 5 — Mature Platform

Potentially introduce:

```text
Multi-Region Active / Active
Advanced Service Mesh
Advanced Policy Automation
Automated Disaster Recovery
Advanced Capacity Management
```

These capabilities are not mandatory until justified.

---

# 154. Avoid Premature Complexity

Do not introduce infrastructure merely because it is common in large companies.

Every infrastructure component should have a reason.

---

# 155. Infrastructure Selection Criteria

Evaluate infrastructure based on:

```text
Business Need
Reliability
Cost
Operational Burden
Scalability
Security
Team Capability
Migration Cost
```

---

# 156. Service Deployment Checklist

Before production deployment:

```text
□ Container Image Built
□ Image Scanned
□ Configuration Defined
□ Secrets Configured
□ Health Checks Defined
□ Resource Limits Defined
□ Networking Defined
□ Database Dependencies Verified
□ Messaging Dependencies Verified
□ Observability Enabled
□ Rollback Tested
□ Migration Plan Verified
□ Security Review Completed
```

---

# 157. Production Readiness Checklist

```text
□ Monitoring
□ Logging
□ Tracing
□ Alerts
□ Backups
□ Restore Validation
□ Capacity Plan
□ Scaling Plan
□ Failure Plan
□ Rollback
□ Security
□ Cost Visibility
□ Incident Procedure
```

---

# 158. Deployment Change Checklist

```text
□ Change Identified
□ Risk Classified
□ Dependencies Identified
□ Compatibility Checked
□ Tests Passed
□ Deployment Plan Ready
□ Rollback Plan Ready
□ Monitoring Ready
□ Deployment Executed
□ Health Verified
□ Business Metrics Verified
```

---

# 159. Deployment Failure Checklist

If deployment fails:

```text
□ Stop Further Rollout
□ Assess Health
□ Check Logs
□ Check Metrics
□ Check Dependencies
□ Roll Back if Required
□ Verify Recovery
□ Preserve Evidence
□ Record Incident
□ Identify Root Cause
```

---

# 160. Infrastructure Failure Checklist

```text
□ Identify Failed Component
□ Determine Blast Radius
□ Protect Customer / Driver Workflows
□ Activate Approved Fallback
□ Recover Infrastructure
□ Validate Data
□ Validate Events
□ Validate Service Health
□ Resume Normal Operations
```

---

# 161. Deployment Anti-Patterns

Avoid:

```text
Manual Server Configuration
Mutable Production Containers
latest-only Image Tags
Production Secrets in Images
Public Databases
Unbounded Autoscaling
Unreviewed Infrastructure Changes
No Rollback Plan
No Backups
Untested Backups
Premature Multi-Region
Premature Kubernetes
Infrastructure Without Ownership
```

---

# 162. Consequences

## 162.1 Positive Consequences

This strategy provides:

```text
Predictable Deployments
Controlled Infrastructure Growth
Lower Initial Complexity
Better Production Reliability
Improved Security
Clear Environment Separation
Automated Delivery
Future Scaling Path
Cloud Portability
```

---

## 162.2 Negative Consequences

The strategy introduces:

```text
Container Management
CI/CD Infrastructure
Infrastructure-as-Code Maintenance
Cloud Operations
Monitoring Costs
Deployment Complexity
Migration Planning
```

These costs are accepted.

---

# 163. Risks

## Risk 1 — Premature Kubernetes Adoption

### Mitigation

Adopt orchestration only when scale or operational requirements justify it.

---

## Risk 2 — Cloud Cost Growth

### Mitigation

Track infrastructure cost and follow:

```text
ADR-0028
```

---

## Risk 3 — Provider Lock-In

### Mitigation

Keep cloud-specific code in infrastructure/provider boundaries.

---

## Risk 4 — Database Overload During Scaling

### Mitigation

Coordinate:

```text
Service Scaling
PgBouncer
PostgreSQL Capacity
```

---

## Risk 5 — Deployment Causes Data Incompatibility

### Mitigation

Use backward-compatible migration strategies.

---

## Risk 6 — Infrastructure Drift

### Mitigation

Use:

```text
Infrastructure as Code
Automated Validation
Drift Detection
```

---

## Risk 7 — Insufficient Disaster Recovery

### Mitigation

Backups, restore testing, recovery procedures, and explicit RTO/RPO.

---

# 164. Alternatives Considered

## 164.1 Deploy Everything on One Server

### Advantages

```text
Very Simple
Low Initial Cost
Easy Local Administration
```

### Disadvantages

```text
Single Point of Failure
Poor Scaling
Service Coupling
Difficult Recovery
```

### Decision

```text
Rejected for production architecture beyond the earliest controlled stage.
```

---

# 165. Kubernetes From Day One

### Advantages

```text
Advanced Scheduling
Scaling
Service Deployment
Strong Ecosystem
```

### Disadvantages

```text
High Operational Complexity
Higher Cost
More Infrastructure Work
```

### Decision

```text
Rejected as a mandatory initial deployment platform.
```

Kubernetes remains an approved future evolution path.

---

# 166. Fully Serverless Architecture

### Advantages

```text
Low Server Management
Elastic Scaling
```

### Disadvantages

```text
Potential Cost at Sustained Load
Runtime Constraints
Complex Real-Time Workloads
Messaging / State Complexity
```

### Decision

```text
Rejected as the universal deployment model.
```

Individual serverless components may still be used when justified.

---

# 167. Fully Self-Managed Infrastructure

### Advantages

```text
Maximum Control
Potential Cost Savings at Scale
```

### Disadvantages

```text
Operational Burden
Patch Management
Backup Responsibility
High Availability Complexity
```

### Decision

```text
Rejected as the default strategy.
```

---

# 168. Fully Managed Everything

### Advantages

```text
Lower Operational Burden
Fast Setup
```

### Disadvantages

```text
Potentially High Cost
Provider Lock-In
Limited Control
```

### Decision

```text
Rejected as a blanket rule.
```

Use managed infrastructure selectively.

---

# 169. Validation

This ADR should be validated through:

```text
Infrastructure Provisioning
Container Builds
Deployment Pipelines
Environment Provisioning
Health Checks
Scaling Tests
Rollback Tests
Migration Tests
Backup / Restore Tests
Failure Tests
Security Tests
Cost Reviews
Disaster Recovery Tests
Staging Deployment
Production Smoke Tests
```

---

# 170. Review Triggers

Revisit this ADR when:

```text
Production Scale Increases Significantly
A New Cloud Provider Is Considered
Kubernetes Becomes Operationally Necessary
Multi-Region Deployment Is Required
Database HA Requirements Change
Messaging Scale Changes
Major Cost Increase Occurs
Disaster Recovery Requirements Change
A Major Infrastructure Incident Occurs
Deployment Frequency Changes
Service Count Increases Significantly
```

---

# 171. Final Principles

The following principles are mandatory:

```text
1. RideForge uses containerized, automated, environment-separated deployment.

2. Development, testing, staging, and production are separate environments.

3. Production deployment must be automated where practical.

4. Production artifacts should be immutable.

5. Application images must not contain production secrets.

6. Environment-specific configuration is injected at runtime.

7. Infrastructure should be managed as code where practical.

8. Cloud-specific dependencies must remain outside core domain logic.

9. Managed services should be preferred when they reduce operational burden at acceptable cost.

10. Self-managed infrastructure may be used when justified by cost or operational requirements.

11. PostgreSQL remains the primary production database.

12. PostGIS remains part of the PostgreSQL strategy where required.

13. Redis remains the real-time state and caching infrastructure.

14. Kafka / Redpanda remains the event-streaming infrastructure.

15. Database connection capacity must be considered across all service instances.

16. PgBouncer should be used where required by database connection scaling.

17. Stateless services should support horizontal scaling where practical.

18. Stateful infrastructure must be treated separately from stateless application services.

19. Autoscaling must consider downstream capacity.

20. Production services must have appropriate health checks.

21. Liveness and readiness must not be treated as identical concepts.

22. Production services should support graceful shutdown.

23. Deployments should minimize customer-visible downtime.

24. Rolling deployment is the default where compatibility permits.

25. Canary or blue-green deployment may be used for higher-risk changes.

26. Database migrations must support rolling deployment compatibility.

27. Event schema changes must support deployment compatibility.

28. Every production deployment must have a rollback strategy.

29. Database rollback must not depend on unsafe destructive reversal.

30. Production databases must have automated backups.

31. Backup restoration must be tested.

32. Disaster recovery must define RTO and RPO.

33. Multi-region deployment is not required until justified.

34. Regional deployment must respect legal and data-residency requirements.

35. Infrastructure routing must never bypass regional ride eligibility rules.

36. Regional failover must not automatically create legally invalid ride operations.

37. Public network exposure must be minimized.

38. Production databases, Redis, and messaging infrastructure should remain private where practical.

39. Production traffic must use TLS.

40. Deployment identities must follow least privilege.

41. CI/CD credentials must be protected.

42. Deployment changes must be auditable.

43. Infrastructure changes must be reviewed.

44. Infrastructure drift should be controlled.

45. Container images should be versioned with immutable identifiers.

46. Production deployments should not rely exclusively on the latest image tag.

47. Container images should be scanned for vulnerabilities.

48. Observability must be enabled before production deployment.

49. Deployment monitoring must include infrastructure and business metrics.

50. Deployment failures must stop or roll back safely.

51. Infrastructure scaling must consider database, Redis, and messaging capacity.

52. Cloud cost must be monitored.

53. Premature infrastructure complexity should be avoided.

54. Kubernetes is an evolution option, not a mandatory initial requirement.

55. Multi-region active/active is an evolution option, not an initial requirement.

56. Infrastructure technology should be selected according to business and operational need.

57. The simplest architecture that satisfies current reliability requirements is preferred.

58. Infrastructure should evolve incrementally with platform scale.

59. Production infrastructure must remain observable, recoverable, and reproducible.

60. The deployment strategy must preserve the architectural boundaries defined by the RideForge ADR set.
```

---

# 172. Status

```text
Decision: ACCEPTED

Deployment Model:
Containerized + Automated + Environment-Separated

Primary Environments:
Development
Testing
Staging
Production

Application Deployment:
Immutable Artifacts

Containerization:
Required

Infrastructure as Code:
Preferred

CI/CD:
Required for Normal Production Delivery

Initial Orchestration:
Simple / Managed Container Runtime Where Appropriate

Kubernetes:
Future Option, Not Mandatory Initially

Database:
PostgreSQL + PostGIS

Database Pooling:
PgBouncer Where Required

Cache / Runtime State:
Redis

Event Streaming:
Kafka / Redpanda

Scaling:
Horizontal Where Practical

Autoscaling:
Selective

Health Checks:
Required

Graceful Shutdown:
Required

Deployment:
Rolling by Default Where Compatible

Canary:
For Higher-Risk Changes Where Appropriate

Rollback:
Required

Backups:
Required

Restore Testing:
Required

Disaster Recovery:
Required

Multi-Region:
Future / Requirement-Driven

Cloud Strategy:
Managed Where Valuable, Portable Where Practical

Cost:
Continuously Controlled

Security:
ADR-0023

Configuration:
ADR-0024

Testing:
ADR-0025

AI:
ADR-0026

Failure Strategy:
ADR-0021

Observability:
ADR-0022

Cost Optimization:
ADR-0028

Architecture Evolution:
ADR-0029

Primary Goal:
Deploy RideForge Reliably and Securely While Keeping Infrastructure Simple Enough to Operate, Cost-Efficient Enough to Grow, and Flexible Enough to Scale
```

---

# 173. Decision Summary

RideForge adopts an incremental deployment model:

```text
                     SOURCE CODE
                          │
                          ▼
                    CI / BUILD
                          │
                          ▼
                  TEST + SECURITY
                          │
                          ▼
                  IMMUTABLE IMAGE
                          │
                          ▼
                     REGISTRY
                          │
             ┌────────────┴────────────┐
             ▼                         ▼
         STAGING                   VALIDATION
             │                         │
             └────────────┬────────────┘
                          ▼
                    PRODUCTION
                          │
             ┌────────────┼────────────┐
             ▼            ▼            ▼
          Service      Database     Messaging
          Runtime     PostgreSQL    Kafka/Redpanda
             │            │            │
             └────────────┼────────────┘
                          ▼
                    OBSERVABILITY
                          │
                          ▼
                 HEALTH + BUSINESS
                    VERIFICATION
```

The infrastructure evolution path is intentionally incremental:

```text
Simple Container Deployment
          ↓
Automated Production Deployment
          ↓
Independent Service Scaling
          ↓
Advanced Orchestration
          ↓
Multi-Region Infrastructure
```

The platform does not adopt infrastructure complexity merely because the architecture is microservice-based.

The governing principle is:

> **Use the simplest deployment architecture that reliably satisfies the current operational requirements, while preserving a clear path to scale.**

---

# 174. Status Metadata

| Field | Value |
|---|---|
| ADR | `0027` |
| Title | Cloud and Deployment Strategy |
| Status | Accepted |
| Category | Cloud / Infrastructure / Deployment |
| Deployment Model | Containerized + Automated |
| Environment Separation | Required |
| Immutable Artifacts | Required |
| CI/CD | Required |
| Infrastructure as Code | Preferred |
| Database | PostgreSQL + PostGIS |
| Connection Pooling | PgBouncer Where Required |
| Runtime State / Cache | Redis |
| Event Streaming | Kafka / Redpanda |
| Horizontal Scaling | Preferred for Stateless Services |
| Autoscaling | Selective |
| Health Checks | Required |
| Graceful Shutdown | Required |
| Rolling Deployment | Default Where Compatible |
| Canary | Risk-Based |
| Rollback | Required |
| Backups | Required |
| Restore Testing | Required |
| Disaster Recovery | Required |
| Multi-Region | Requirement-Driven |
| Kubernetes | Future Evolution Option |
| Cloud Strategy | Managed Where Valuable |
| Cost Strategy | ADR-0028 |
| Security | ADR-0023 |
| Configuration | ADR-0024 |
| Testing | ADR-0025 |
| AI Governance | ADR-0026 |
| Failure Strategy | ADR-0021 |
| Observability | ADR-0022 |
| Architecture Evolution | ADR-0029 |
| Next ADR | `0028-COST_OPTIMIZATION_STRATEGY.md` |
