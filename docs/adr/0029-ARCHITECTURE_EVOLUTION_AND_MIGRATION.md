# ADR-0029: Architecture Evolution and Migration

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Architecture / Evolution / Migration / Platform Strategy  
> **Scope:** Architecture evolution, service extraction, infrastructure migration, database evolution, event migration, API evolution, dispatch evolution, AI evolution, cloud migration, regional expansion, deprecation, backward compatibility, rollout, rollback, and long-term platform modernization  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is being developed as a production-grade ride-hailing platform with:

```text
Microservices
Domain-Driven Design
Event-Driven Architecture
PostgreSQL
PostGIS
Redis
Kafka / Redpanda
Real-Time Driver Location
Stand Dispatch
Smart Dispatch
ETA Prediction
AI / ML Components
External Providers
Cloud Infrastructure
```

The platform is expected to evolve significantly as:

```text
Traffic Increases
Regions Expand
Services Mature
Business Rules Change
Infrastructure Requirements Grow
AI Capabilities Improve
Operational Experience Increases
```

An architecture that is appropriate for an early-stage platform may not remain appropriate at higher scale.

The architecture therefore needs an explicit evolution strategy.

---

# 2. Problem

Without a deliberate migration strategy, architecture changes can create:

```text
Extended Downtime
Data Loss
Event Incompatibility
Service Coupling
Migration Lock-In
Operational Risk
Rollback Difficulty
Dual-System Complexity
Technical Debt
```

RideForge needs a consistent approach for changing:

```text
Services
Databases
Messaging
Caching
APIs
Dispatch Algorithms
AI Models
Cloud Infrastructure
Regional Architecture
```

without destabilizing the production platform.

---

# 3. Decision

RideForge will evolve incrementally through:

```text
Backward-Compatible Changes
Controlled Migration
Incremental Rollout
Observable Validation
Explicit Rollback
Controlled Deprecation
```

The platform will prefer:

> **Evolution over replacement, incremental migration over big-bang migration, and reversible changes over irreversible changes.**

---

# 4. Core Evolution Principles

RideForge follows:

```text
Evolve Incrementally
Preserve Compatibility
Migrate Before Removing
Measure Before Replacing
Keep Rollback Available
Avoid Big-Bang Migration
Protect Data
Protect Events
Protect Customers
Protect Drivers
```

---

# 5. Architecture Is Not Static

The architecture is expected to change.

A valid architecture decision today may become inappropriate later because:

```text
Scale Changes
Cost Changes
Technology Changes
Business Requirements Change
Regulation Changes
Operational Experience Changes
```

Architecture evolution is therefore a normal engineering activity.

---

# 6. ADR-Driven Evolution

Major architecture changes must be documented through ADRs.

The ADR should explain:

```text
Current State
Problem
Decision
Alternatives
Migration Strategy
Risks
Rollback
Consequences
```

---

# 7. Existing ADRs Are Historical Decisions

An ADR records the decision that was accepted at a particular point in the architecture lifecycle.

A later ADR may replace it.

Do not silently rewrite historical architectural decisions to make them appear as though they were never made.

---

# 8. Superseding an ADR

When an architectural decision changes:

```text
New ADR
```

should explicitly identify:

```text
Superseded ADR
```

and explain why the architecture changed.

---

# 9. Architecture Evolution Lifecycle

The preferred evolution flow is:

```text
Identify Problem
      ↓
Measure Current State
      ↓
Define Target State
      ↓
Evaluate Alternatives
      ↓
Create ADR
      ↓
Design Migration
      ↓
Implement Compatibility
      ↓
Migrate Incrementally
      ↓
Validate
      ↓
Complete Cutover
      ↓
Remove Legacy Path
      ↓
Update Documentation
```

---

# 10. Current State Before Target State

Migration planning must explicitly document:

```text
Current Architecture
```

before defining:

```text
Target Architecture
```

---

# 11. Target Architecture

The target architecture should define:

```text
Services
Data Ownership
Communication
Infrastructure
Operational Model
Scaling
Security
Observability
```

---

# 12. Migration Boundary

Every migration should have a clear boundary.

Examples:

```text
Service
Database Table
API
Event
Infrastructure Component
AI Model
Region
```

---

# 13. Migration Ownership

Every major migration must have:

```text
Owner
Scope
Start Condition
Completion Condition
Rollback Owner
```

---

# 14. Migration Plan

A migration plan should identify:

```text
Step
Dependency
Risk
Validation
Rollback
Completion Criteria
```

---

# 15. Expand-and-Contract Strategy

For changes requiring compatibility across versions, prefer:

```text
Expand
    ↓
Migrate
    ↓
Switch
    ↓
Contract
```

---

# 16. Expand Phase

During expansion:

```text
New Structure
```

is introduced without immediately removing the old structure.

Examples:

```text
New Column
New Event Field
New API Endpoint
New Service
New Topic
```

---

# 17. Migrate Phase

Data or traffic is moved gradually.

Examples:

```text
Backfill Data
Dual Write
Dual Read
Shadow Traffic
Traffic Mirroring
```

where appropriate.

---

# 18. Switch Phase

The system begins using the new implementation as the primary path.

---

# 19. Contract Phase

After successful migration:

```text
Legacy Structure
```

may be removed.

---

# 20. Big-Bang Migration

Big-bang migration should be avoided for critical production systems unless:

```text
Migration Is Simple
Downtime Is Acceptable
Rollback Is Straightforward
Risk Is Low
```

---

# 21. Backward Compatibility

Compatibility should be preserved whenever multiple versions may coexist.

This applies to:

```text
APIs
Events
Database Schemas
Configuration
Service Interfaces
```

---

# 22. API Evolution

API changes should prefer:

```text
Additive Change
```

over destructive change.

---

# 23. Additive API Change

Prefer:

```text
Add Optional Field
```

over:

```text
Rename Existing Required Field Immediately
```

when compatibility matters.

---

# 24. API Versioning

API versioning should be introduced when a change cannot remain backward compatible.

---

# 25. API Deprecation

Deprecated APIs should have:

```text
Deprecation Date
Replacement
Migration Guidance
Removal Target
Usage Monitoring
```

---

# 26. API Removal

Do not remove an API solely because a replacement exists.

Verify:

```text
Consumer Migration
Usage
Operational Safety
```

first.

---

# 27. Event Evolution

Event schemas must support compatible evolution.

---

# 28. Event Additive Changes

Adding optional event fields is generally preferred over removing or changing the meaning of existing fields.

---

# 29. Event Semantic Changes

Changing the meaning of an existing field should be treated as a potentially breaking change.

Create:

```text
New Field
```

or:

```text
New Event Version
```

where appropriate.

---

# 30. Event Consumer Migration

Before removing an event field or changing an event contract:

```text
Identify Consumers
Update Consumers
Deploy Consumers
Verify Compatibility
Remove Legacy Contract
```

---

# 31. Event Replay Consideration

Event schema changes must consider historical events that may be replayed.

---

# 32. Database Evolution

Database changes must support application deployment compatibility.

---

# 33. Database Additive Change

Prefer:

```text
Add Column
Add Table
Add Index
```

before removing existing structures.

---

# 34. Database Backfill

Large data migrations should be performed in controlled batches where necessary.

---

# 35. Backfill Safety

Backfills should have:

```text
Rate Limit
Progress Tracking
Retry Strategy
Validation
Rollback / Recovery Plan
```

where appropriate.

---

# 36. Dual Write

Dual writes may be used temporarily when migrating data between representations.

They must have:

```text
Consistency Strategy
Failure Handling
Reconciliation
Removal Plan
```

---

# 37. Dual Read

Dual reads may be used during validation of a new data source.

The system should define:

```text
Primary Source
Comparison Source
Mismatch Handling
Cutover Criteria
```

---

# 38. Data Reconciliation

When migrating data, reconciliation should verify:

```text
Count
Identity
Important Fields
Relationships
Business Invariants
```

---

# 39. Data Migration Validation

Migration completion requires explicit validation.

---

# 40. Database Ownership Evolution

As services evolve, database ownership may change.

The target state must clearly identify:

```text
Owner
Read Access
Write Access
Migration Ownership
```

---

# 41. Shared Database Evolution

If multiple services initially share PostgreSQL, future service extraction should gradually establish stronger ownership boundaries.

---

# 42. Database-per-Service Evolution

Database-per-service may be introduced when:

```text
Service Scale
Ownership
Failure Isolation
Independent Deployment
Data Boundaries
```

justify the additional complexity.

---

# 43. Do Not Split Databases Prematurely

Database separation introduces:

```text
Distributed Transactions
Data Duplication
Eventual Consistency
Operational Complexity
```

Therefore it should be driven by actual requirements.

---

# 44. Service Extraction

A service may be extracted when there is a meaningful reason such as:

```text
Independent Scaling
Independent Deployment
Clear Domain Boundary
Security Isolation
Operational Ownership
```

---

# 45. Service Extraction Strategy

Prefer:

```text
Identify Domain
      ↓
Define Contract
      ↓
Introduce New Service
      ↓
Route New Traffic
      ↓
Migrate Existing Traffic
      ↓
Remove Old Path
```

---

# 46. Strangler Pattern

The Strangler pattern may be used for gradual service replacement.

Conceptually:

```text
Existing System
      │
      ├── Legacy Path
      │
      └── New Path

New Path Gradually Expands
        ↓
Legacy Path Shrinks
        ↓
Legacy Removed
```

---

# 47. Service Boundary Validation

Before extracting a service, verify:

```text
Domain Ownership
Data Ownership
API
Events
Failure Behaviour
Observability
Deployment
```

---

# 48. Monolith-to-Microservice Evolution

If RideForge begins with a modular monolith or simplified deployment, services may be extracted later.

Microservice extraction should be driven by:

```text
Operational Need
Scaling Need
Team Ownership
Domain Boundary
```

rather than ideology.

---

# 49. Microservice Consolidation

The reverse is also allowed.

If two services create unnecessary complexity, they may be consolidated.

---

# 50. Service Consolidation Criteria

Consider consolidation when:

```text
Low Independent Scaling
Strong Coupling
High Communication Overhead
Shared Deployment Lifecycle
Minimal Domain Separation
```

---

# 51. Infrastructure Evolution

Infrastructure may evolve from:

```text
Simple Deployment
```

to:

```text
Managed Containers
```

then:

```text
Advanced Orchestration
```

and eventually:

```text
Multi-Region Infrastructure
```

when justified.

---

# 52. Kubernetes Migration

If Kubernetes becomes necessary, migration should be incremental.

Example:

```text
Existing Runtime
      ↓
Containerize Services
      ↓
Deploy Selected Service
      ↓
Validate
      ↓
Expand Migration
```

---

# 53. Cloud Provider Migration

Cloud-provider migration should avoid rewriting domain logic.

---

# 54. Provider Adapter Strategy

Cloud-specific capabilities should remain behind:

```text
Infrastructure Interfaces
Provider Adapters
Deployment Configuration
```

where practical.

---

# 55. Cloud Migration Steps

Conceptually:

```text
Inventory Dependencies
      ↓
Classify Provider Coupling
      ↓
Create Abstraction Where Valuable
      ↓
Deploy Target Infrastructure
      ↓
Replicate / Migrate Data
      ↓
Shift Traffic
      ↓
Validate
      ↓
Decommission Source
```

---

# 56. Cloud Migration Warning

Do not create abstractions merely for theoretical portability.

Abstraction is justified when:

```text
Migration Probability
Operational Value
Cost
Risk
```

justify it.

---

# 57. Database Technology Migration

Changing databases is a high-risk migration.

It requires:

```text
Schema Mapping
Data Migration
Query Compatibility
Performance Testing
Consistency Validation
Rollback
```

---

# 58. PostgreSQL Migration Principle

PostgreSQL should remain the default unless a new database provides a material advantage.

---

# 59. Database Replacement Criteria

A database replacement requires demonstrated benefit in areas such as:

```text
Scale
Latency
Availability
Cost
Workload Fit
```

---

# 60. Database Migration Strategy

Prefer:

```text
Dual Write / Replication
+
Validation
+
Gradual Read Cutover
+
Legacy Decommission
```

where appropriate.

---

# 61. Redis Evolution

Redis may evolve through:

```text
Single Instance
→ Managed Redis
→ Replication / HA
→ Clustered Architecture
```

as workload and availability requirements increase.

---

# 62. Redis Migration

Redis migration must consider:

```text
TTL
Key Format
Data Volatility
Cache Rebuild
Real-Time State
```

---

# 63. Kafka / Redpanda Evolution

Messaging infrastructure may evolve in:

```text
Broker Capacity
Partitions
Replication
Retention
Deployment Model
```

without changing domain event semantics unnecessarily.

---

# 64. Kafka / Redpanda Provider Migration

If migrating between Kafka-compatible platforms:

```text
Validate Protocol Compatibility
Validate Semantics
Validate Retention
Validate Consumer Behaviour
Validate Operations
```

before cutover.

---

# 65. Event Infrastructure Migration

Messaging migration should use:

```text
Replication
Dual Publishing
Controlled Consumer Migration
Traffic Cutover
```

where appropriate.

---

# 66. Outbox Preservation

Infrastructure migrations must preserve the Outbox guarantee.

This follows:

```text
ADR-0012 — Outbox Pattern
```

---

# 67. DLQ Preservation

Migration must preserve:

```text
Retry
DLQ
Replay
Failure Metadata
```

behaviour.

---

# 68. Dispatch Evolution

RideForge supports:

```text
Stand Dispatch
Smart Dispatch
Hybrid Dispatch
```

---

# 69. Dispatch Algorithm Evolution

A new dispatch algorithm should first operate in:

```text
Offline Evaluation
```

then:

```text
Shadow
```

then:

```text
Controlled Rollout
```

before full production use.

---

# 70. Stand Dispatch Preservation

Stand dispatch must remain independently operational when required.

---

# 71. Smart Dispatch Evolution

Smart dispatch models may evolve without changing the core dispatch domain contract.

---

# 72. AI Model Migration

AI model migration follows:

```text
Train
→ Evaluate
→ Register
→ Shadow
→ Canary
→ Production
→ Monitor
→ Retire Old Model
```

---

# 73. AI Model Rollback

A previous approved model must remain available until the new model is considered stable.

---

# 74. AI Model Replacement

Model replacement should preserve:

```text
Inference Contract
Feature Semantics
Fallback
Monitoring
```

where possible.

---

# 75. Feature Evolution

Feature changes should be versioned when semantics materially change.

---

# 76. Training / Serving Compatibility

During model migration:

```text
Training Features
```

and:

```text
Serving Features
```

must remain compatible.

---

# 77. ETA Evolution

ETA providers and models may evolve independently.

The application should depend on an ETA capability rather than a specific provider implementation.

---

# 78. ETA Provider Migration

A provider migration may use:

```text
Shadow Comparison
```

between:

```text
Current Provider
New Provider
```

before switching production traffic.

---

# 79. External Provider Replacement

Provider adapters should isolate external API changes.

---

# 80. Provider Migration

Preferred flow:

```text
Adapter
→ New Provider Adapter
→ Shadow
→ Compare
→ Canary
→ Cutover
→ Remove Legacy Adapter
```

---

# 81. Regional Expansion

Adding a new operating region should be treated as an architecture change when it affects:

```text
Legal Rules
Dispatch
Data
Providers
Infrastructure
Operations
```

---

# 82. Region Onboarding

A new region should follow:

```text
Legal Validation
      ↓
Configuration
      ↓
Provider Validation
      ↓
Operational Testing
      ↓
Staging
      ↓
Controlled Launch
      ↓
Monitoring
```

---

# 83. Regional Configuration

Region-specific behaviour should be configuration-driven where appropriate.

Avoid duplicating entire services for small regional differences.

---

# 84. Regional Exceptions

Regional differences should be explicit.

Do not hide region-specific rules inside generic algorithms.

---

# 85. Legal Architecture Evolution

If legal requirements change:

```text
Domain Rules
Configuration
Tests
Dispatch
Routing
```

may require coordinated migration.

---

# 86. Data Residency Evolution

If data residency requirements change, migration must consider:

```text
Data Location
Backups
Replication
Logs
Analytics
AI Training
```

---

# 87. Observability Evolution

Observability systems may evolve as traffic grows.

Migration should preserve:

```text
Logs
Metrics
Traces
Correlation
Alerts
```

where operationally required.

---

# 88. Observability Cost Migration

Higher-scale observability may require:

```text
Sampling
Retention Changes
Storage Tiering
Aggregation
```

according to:

```text
ADR-0028
```

---

# 89. Security Evolution

Security requirements may increase over time.

Architecture migration must preserve:

```text
Authentication
Authorization
Secret Management
Encryption
Auditability
```

---

# 90. Configuration Evolution

Configuration changes should remain compatible across rolling deployments.

---

# 91. Feature Flags

Feature flags may be used to control migration traffic.

Examples:

```text
ENABLE_NEW_DISPATCH
USE_NEW_ETA_PROVIDER
USE_NEW_RANKING_MODEL
```

---

# 92. Feature Flag Governance

Feature flags must have:

```text
Owner
Purpose
Default
Created Date
Removal Plan
```

---

# 93. Temporary Flags

Migration flags should be temporary.

Long-lived flags create configuration complexity.

---

# 94. Traffic Migration

Traffic may be shifted using:

```text
Percentage
Region
User Segment
Driver Segment
Operational Zone
```

where appropriate.

---

# 95. Traffic Migration Safety

Traffic shifts should be:

```text
Gradual
Observable
Reversible
```

---

# 96. Canary Migration

Canary migration may begin with:

```text
1%
5%
10%
25%
50%
100%
```

or another risk-appropriate progression.

---

# 97. Canary Evaluation

Before increasing traffic, evaluate:

```text
Error Rate
Latency
Business Metrics
Resource Usage
Failure Rate
```

---

# 98. Automatic Traffic Rollback

Automatic rollback may be used when thresholds are clear and safe.

---

# 99. Migration Observability

Every migration should have dedicated observability for:

```text
Old Path
New Path
Traffic Distribution
Errors
Latency
Business Outcomes
```

---

# 100. Migration Success Criteria

A migration should have explicit completion criteria.

Examples:

```text
100% Traffic Migrated
Zero Critical Errors
Data Reconciled
Performance Within Target
Legacy Usage Near Zero
Rollback Window Complete
```

---

# 101. Migration Failure Criteria

Define conditions that trigger:

```text
Pause
Rollback
Investigation
```

before migration begins.

---

# 102. Rollback

Rollback must be planned before migration begins.

---

# 103. Rollback Simplicity

Prefer migrations where rollback is:

```text
Fast
Safe
Tested
```

---

# 104. Irreversible Changes

If a change is difficult to reverse:

```text
Increase Validation
Increase Rollout Control
Increase Backup / Recovery Preparation
```

before executing it.

---

# 105. Rollback Limitations

Not every migration can support perfect rollback.

For example:

```text
Destructive Data Transformation
External Side Effects
Irreversible Provider Actions
```

may require recovery rather than reversal.

---

# 106. Forward Recovery

When rollback is impossible, define:

```text
Forward Fix
Data Repair
Compensation
Recovery Procedure
```

---

# 107. Migration Checkpoints

Long migrations should have checkpoints.

Example:

```text
Phase 1 Complete
→ Validate

Phase 2 Complete
→ Validate

Phase 3 Complete
→ Validate
```

---

# 108. Migration Pause

The migration process must allow controlled pause where possible.

---

# 109. Migration Abort

Abort criteria must be defined before high-risk migrations.

---

# 110. Migration Windows

High-risk migrations may be scheduled during appropriate operational windows.

---

# 111. Migration and Peak Traffic

Avoid high-risk infrastructure migrations during predictable peak demand unless required.

---

# 112. Migration Communication

Important production migrations should be communicated to relevant operators and owners.

---

# 113. Migration Runbook

Major migrations should have a runbook containing:

```text
Prerequisites
Commands / Actions
Validation
Rollback
Escalation
Completion
```

---

# 114. Migration Dry Run

Where practical, perform a dry run in:

```text
Development
Staging
```

before production.

---

# 115. Production Migration

Production migration should follow the tested sequence as closely as possible.

---

# 116. Migration Audit

Record:

```text
Who
What
When
Why
Result
```

for important migrations.

---

# 117. Migration Testing

Migration tests should include:

```text
Fresh Environment
Existing Environment
Representative Data
Failure Conditions
Rollback / Recovery
```

where applicable.

---

# 118. Migration Performance

Large migrations must consider:

```text
CPU
Memory
IO
Database Locks
Network
Replication Lag
```

---

# 119. Online Migration

Prefer online or low-impact migration techniques for critical production databases where practical.

---

# 120. Lock Management

Database migrations must avoid unnecessary long-running locks.

---

# 121. Batch Migration

Large data transformations should use controlled batches when practical.

---

# 122. Migration Rate Limiting

Migration throughput may need to be limited to protect production workloads.

---

# 123. Migration Backpressure

If migration impacts production latency or capacity, slow or pause the migration.

---

# 124. Data Integrity

Migration must preserve:

```text
Identity
Relationships
Constraints
Business Invariants
```

---

# 125. Data Validation

Validate both:

```text
Source
Target
```

before decommissioning the source.

---

# 126. Reconciliation

For important migrations, reconciliation should identify mismatches.

---

# 127. Migration Idempotency

Migration scripts should be safely repeatable where practical.

---

# 128. Migration Checkpointing

Large migrations should track progress so they can resume safely.

---

# 129. Migration Tooling

Use appropriate tooling for:

```text
Database Migration
Data Transfer
Infrastructure Provisioning
Traffic Routing
```

The specific tool is not mandated by this ADR.

---

# 130. Legacy Systems

Legacy systems should be classified as:

```text
Active
Migration
Deprecated
Retired
```

---

# 131. Legacy Ownership

Every legacy system should have an owner until retirement.

---

# 132. Legacy Monitoring

Legacy components should remain observable while they serve production traffic.

---

# 133. Legacy Decommissioning

Before removal:

```text
Verify No Traffic
Verify No Consumers
Verify No Dependencies
Verify Data Retention
Verify Recovery
```

---

# 134. Decommissioning

Decommission in stages:

```text
Stop New Writes
→ Drain Traffic
→ Verify
→ Disable
→ Observe
→ Remove Infrastructure
→ Archive Required Artifacts
```

---

# 135. Decommissioning Data

Do not delete historical data without confirming:

```text
Retention Requirements
Legal Requirements
Operational Requirements
Recovery Requirements
```

---

# 136. Decommissioning Events

Before removing a topic or event:

```text
Verify Producers
Verify Consumers
Verify Replay Requirements
```

---

# 137. Decommissioning APIs

Before removing an API:

```text
Usage = Zero
Consumers Migrated
Deprecation Completed
```

---

# 138. Technical Debt

Architecture evolution should explicitly track technical debt created by transitional states.

---

# 139. Transitional Architecture

Temporary architecture may include:

```text
Dual Writes
Dual Reads
Compatibility Adapters
Feature Flags
Shadow Systems
```

---

# 140. Transitional Debt Deadline

Every temporary migration mechanism should have a planned removal condition.

---

# 141. Migration Debt Register

Major migrations may maintain:

```text
Migration Item
Owner
Status
Target Removal
Risk
```

---

# 142. Avoid Permanent Transitional Systems

A migration mechanism that remains indefinitely becomes part of the architecture and increases complexity.

---

# 143. Architecture Fitness

After migration, evaluate whether the new architecture actually improved:

```text
Reliability
Performance
Cost
Scalability
Maintainability
Developer Experience
```

---

# 144. Post-Migration Review

Major migrations should receive a post-migration review.

---

# 145. Post-Migration Questions

Review:

```text
Did We Achieve the Goal?
What Broke?
What Cost More?
What Became Simpler?
What Became More Complex?
What Should Be Removed?
```

---

# 146. Migration Lessons

Important migration lessons should update:

```text
Documentation
Runbooks
Testing
ADRs
Architecture Standards
```

---

# 147. Architecture Fitness Review

Architecture should periodically be reviewed against actual workload.

---

# 148. Architecture Evolution Triggers

Review architecture when:

```text
Traffic Increases Significantly
Latency Targets Change
Service Count Increases
Database Becomes Bottleneck
Messaging Becomes Bottleneck
Cloud Cost Becomes Material
New Region Is Added
AI Becomes Operationally Critical
Legal Requirements Change
Major Incident Occurs
```

---

# 149. Do Not Optimize Hypothetical Scale

Architecture should evolve based on:

```text
Observed Need
Measured Bottleneck
Expected Near-Term Requirement
```

rather than theoretical extreme scale.

---

# 150. Avoid Premature Distributed Complexity

Do not introduce:

```text
Additional Databases
Service Mesh
Multi-Region Active/Active
Complex Event Topologies
Advanced Kubernetes
```

without a demonstrated requirement.

---

# 151. Preserve Future Options

Avoid early decisions that unnecessarily prevent future evolution.

---

# 152. Stable Domain Boundaries

Even when infrastructure changes, domain boundaries should remain stable where they continue to represent valid business concepts.

---

# 153. Domain vs Infrastructure Evolution

Infrastructure may change without changing the domain.

Example:

```text
Redis
→ Different Redis Deployment
```

should not require rewriting ride-domain rules.

---

# 154. Service Contract Stability

Stable contracts make infrastructure migration easier.

---

# 155. Adapter Strategy

External dependencies should be isolated behind adapters where practical.

---

# 156. Architecture Compatibility

When changing an implementation, preserve:

```text
Contract
Semantics
Failure Behaviour
Observability
Security
```

where possible.

---

# 157. Migration Risk Classification

Migrations may be classified:

```text
Low
Medium
High
Critical
```

based on:

```text
Data Risk
Customer Impact
Operational Impact
Rollback Difficulty
```

---

# 158. Low-Risk Migration

Examples:

```text
Non-Critical Configuration
Internal Refactoring
Small Additive Schema
```

---

# 159. Medium-Risk Migration

Examples:

```text
Service Extraction
Provider Change
Moderate Data Migration
```

---

# 160. High-Risk Migration

Examples:

```text
Primary Database Migration
Core Dispatch Change
Messaging Infrastructure Migration
Regional Architecture Change
```

---

# 161. Critical Migration

Examples:

```text
Production Data Destructive Migration
Major Payment Architecture Change
Core Ride State Migration
```

Critical migrations require stronger approval and validation.

---

# 162. Migration Approval

The required approval level should match migration risk.

---

# 163. Migration Freeze

During high-risk migration, unrelated architecture changes may be temporarily restricted.

---

# 164. Migration Dependency Control

Avoid executing multiple major migrations simultaneously unless dependencies require it and risk is explicitly managed.

---

# 165. Migration Sequencing

Prefer:

```text
One Major Boundary at a Time
```

unless there is a strong reason to combine changes.

---

# 166. Architecture Evolution and Cost

Every migration should consider:

```text
Migration Cost
Operating Cost
Savings
Engineering Cost
```

---

# 167. Architecture Evolution and Reliability

Every migration must preserve or intentionally improve:

```text
Availability
Durability
Recovery
```

---

# 168. Architecture Evolution and Security

Every migration must preserve or improve:

```text
Authentication
Authorization
Secret Management
Encryption
Auditability
```

---

# 169. Architecture Evolution and Observability

Every migration must preserve sufficient:

```text
Logs
Metrics
Traces
Alerts
```

---

# 170. Architecture Evolution and Testing

Every migration must have appropriate:

```text
Unit
Integration
Contract
E2E
Failure
Performance
```

testing based on risk.

---

# 171. Architecture Evolution and AI

AI migrations must follow:

```text
ADR-0026 — Model and AI Governance
```

---

# 172. Architecture Evolution and Deployment

Deployment evolution follows:

```text
ADR-0027 — Cloud and Deployment Strategy
```

---

# 173. Architecture Evolution and Cost

Migration cost evaluation follows:

```text
ADR-0028 — Cost Optimization Strategy
```

---

# 174. Architecture Evolution and Failure

Migration failure handling follows:

```text
ADR-0021 — Failure and Degradation Strategy
```

---

# 175. Architecture Evolution and Observability

Migration observability follows:

```text
ADR-0022 — Observability Strategy
```

---

# 176. Architecture Evolution and Testing

Migration testing follows:

```text
ADR-0025 — Testing and Integration Strategy
```

---

# 177. Architecture Evolution and Configuration

Migration configuration follows:

```text
ADR-0024 — Configuration and Environment Strategy
```

---

# 178. Architecture Evolution and Security

Migration security follows:

```text
ADR-0023 — Security and Secret Management
```

---

# 179. Architecture Evolution Checklist

Before starting a major migration:

```text
□ Current State Documented
□ Target State Documented
□ Problem Clearly Defined
□ Alternatives Evaluated
□ ADR Approved
□ Owner Assigned
□ Dependencies Identified
□ Risk Classified
□ Migration Plan Written
□ Rollback Plan Written
□ Validation Criteria Defined
□ Monitoring Ready
□ Test Plan Ready
□ Data Backup Verified
□ Communication Plan Ready
```

---

# 180. Migration Execution Checklist

```text
□ Preconditions Verified
□ Baseline Metrics Captured
□ Compatibility Layer Deployed
□ Migration Started
□ Progress Monitored
□ Data Validated
□ Error Rate Monitored
□ Latency Monitored
□ Business Metrics Monitored
□ Rollout Increased Carefully
□ Cutover Completed
□ Legacy Traffic Verified
□ Migration Completion Confirmed
```

---

# 181. Migration Completion Checklist

```text
□ 100% Required Traffic Migrated
□ Data Reconciled
□ Error Rates Normal
□ Latency Normal
□ Business Metrics Acceptable
□ Monitoring Stable
□ Rollback Window Evaluated
□ Legacy Path Disabled
□ Temporary Flags Removed
□ Temporary Infrastructure Removed
□ Documentation Updated
□ ADR Updated if Necessary
□ Post-Migration Review Completed
```

---

# 182. Decommission Checklist

```text
□ No Active Traffic
□ No Active Consumers
□ No Active Producers
□ No Required Dependencies
□ Data Retention Confirmed
□ Backup Requirements Confirmed
□ Replacement Verified
□ Monitoring Removed Carefully
□ Infrastructure Removed
□ Secrets Removed
□ Documentation Updated
```

---

# 183. Migration Anti-Patterns

Avoid:

```text
Big-Bang Migration Without Need
No Rollback Plan
No Data Validation
No Monitoring
No Compatibility Layer
Unbounded Dual Writes
Permanent Feature Flags
Permanent Dual Reads
Untracked Technical Debt
Migrating During Peak Traffic Without Reason
Changing Multiple Critical Boundaries Simultaneously
Deleting Legacy Systems Immediately
```

---

# 184. Consequences

## 184.1 Positive Consequences

This strategy provides:

```text
Safer Architecture Evolution
Lower Migration Risk
Better Rollback
Improved Compatibility
Controlled Technical Debt
Clear Ownership
Better Long-Term Maintainability
```

---

## 184.2 Negative Consequences

The strategy introduces:

```text
Temporary Dual Systems
Migration Complexity
Additional Monitoring
Compatibility Code
Migration Documentation
Temporary Infrastructure Cost
```

These costs are accepted because safe evolution is preferable to risky big-bang replacement.

---

# 185. Risks

## Risk 1 — Transitional Systems Become Permanent

### Mitigation

Every temporary migration mechanism requires:

```text
Owner
Removal Condition
Target Date / Milestone
```

where practical.

---

## Risk 2 — Migration Becomes Too Complex

### Mitigation

Break migrations into smaller independently validated phases.

---

## Risk 3 — Rollback Becomes Impossible

### Mitigation

Prefer backward-compatible and reversible migration steps.

---

## Risk 4 — Data Inconsistency

### Mitigation

Use:

```text
Reconciliation
Checksums / Counts Where Appropriate
Business Invariants
Controlled Writes
```

---

## Risk 5 — Cost Explosion

### Mitigation

Monitor temporary infrastructure and remove it after migration.

---

## Risk 6 — Architecture Drift

### Mitigation

Update ADRs and architecture documentation after major migrations.

---

## Risk 7 — Excessive Abstraction

### Mitigation

Introduce abstraction only when migration probability and value justify it.

---

# 186. Alternatives Considered

## 186.1 Big-Bang Rewrite

### Advantages

```text
Clean Target Architecture
No Transitional Code
```

### Disadvantages

```text
High Risk
Long Development Time
Large Rollback Surface
Potential Data Loss
```

### Decision

```text
Rejected as the default migration strategy.
```

---

# 187. Replace Everything at Once

### Advantages

```text
Fast Conceptual Transition
```

### Disadvantages

```text
High Operational Risk
Hard Debugging
Difficult Rollback
```

### Decision

```text
Rejected.
```

---

# 188. Never Change Architecture

### Advantages

```text
No Migration Cost
```

### Disadvantages

```text
Technical Debt
Scaling Limits
Increasing Cost
Poor Maintainability
```

### Decision

```text
Rejected.
```

---

# 189. Rewrite Whenever Technology Changes

### Advantages

```text
Modern Stack
```

### Disadvantages

```text
Constant Migration
Business Risk
Engineering Waste
```

### Decision

```text
Rejected.
```

---

# 190. Build for Maximum Future Scale From Day One

### Advantages

```text
Potentially Fewer Large Migrations
```

### Disadvantages

```text
High Initial Complexity
High Cost
Unvalidated Assumptions
```

### Decision

```text
Rejected.
```

---

# 191. Validation

This ADR should be validated through:

```text
Migration Dry Runs
Staging Migrations
Data Reconciliation
Contract Tests
Integration Tests
Failure Tests
Rollback Tests
Load Tests
Performance Tests
Security Validation
Observability Validation
Cost Validation
Production Controlled Rollout
Post-Migration Review
```

---

# 192. Review Triggers

Revisit this ADR when:

```text
A Major Architecture Boundary Changes
A Database Is Replaced
A Messaging Platform Is Replaced
A Cloud Provider Is Changed
A Service Is Extracted
Services Are Consolidated
Dispatch Architecture Changes
AI Architecture Changes
A New Region Is Added
A Major Migration Incident Occurs
Technical Debt Becomes Material
```

---

# 193. Final Principles

The following principles are mandatory:

```text
1. RideForge architecture is expected to evolve.

2. Architecture evolution is a normal engineering activity.

3. Major architectural changes must be documented.

4. Historical ADRs must not be silently rewritten.

5. A new ADR should supersede an old ADR when a decision materially changes.

6. Current state must be understood before defining target state.

7. Major migrations require an explicit migration plan.

8. Migration ownership must be explicit.

9. Migration risk must be classified.

10. Backward compatibility should be preserved whenever practical.

11. Additive changes are preferred over destructive changes.

12. Expand-and-contract is the default strategy for compatibility-sensitive changes.

13. Big-bang migrations should be avoided.

14. Database migrations should preserve application compatibility.

15. Event migrations should preserve producer-consumer compatibility.

16. API migrations should preserve client compatibility where possible.

17. Data migrations must protect data integrity.

18. Large migrations should be performed in controlled batches where appropriate.

19. Data reconciliation is required for important data migrations.

20. Dual writes require an explicit consistency and removal strategy.

21. Dual reads require an explicit comparison and cutover strategy.

22. Temporary migration mechanisms must have removal conditions.

23. Transitional architecture must not silently become permanent.

24. Feature flags used for migration must have owners and removal plans.

25. Traffic migration should be gradual where risk justifies it.

26. Canary rollout should be used for high-risk changes where practical.

27. Migration success criteria must be defined before execution.

28. Migration failure criteria must be defined before execution.

29. Rollback must be considered before migration begins.

30. Irreversible migrations require stronger validation.

31. When rollback is impossible, forward recovery must be defined.

32. Major migrations should have runbooks.

33. Major migrations should be dry-run in lower environments where practical.

34. Production migrations should follow validated procedures.

35. Migration progress must be observable.

36. Migration metrics must include technical and business outcomes.

37. Database migrations must consider locks and production workload.

38. Large data migrations should support checkpointing where practical.

39. Migration scripts should be repeatable where practical.

40. Service extraction must be driven by meaningful domain or operational value.

41. Microservices may be consolidated when separation no longer provides sufficient value.

42. Database separation should not be introduced prematurely.

43. Cloud-specific implementations should remain outside core domain logic.

44. Provider adapters should be used where they provide meaningful migration value.

45. Infrastructure can evolve independently from domain logic where boundaries remain stable.

46. Dispatch algorithms must evolve through controlled validation.

47. Stand Dispatch must remain available where operationally required.

48. Smart Dispatch changes must preserve deterministic hard constraints.

49. AI model migrations must follow model governance.

50. ETA provider migrations should support controlled comparison where practical.

51. Regional expansion must consider legal, operational, data, and infrastructure requirements.

52. Regional failover must not bypass legal ride restrictions.

53. Architecture should not be optimized for hypothetical extreme scale without evidence.

54. New infrastructure should be introduced when measurable requirements justify it.

55. Migration cost must be considered alongside operating cost.

56. Migration complexity is part of architecture cost.

57. Security must not be weakened during migration.

58. Observability must not be removed during migration.

59. Required backups must be verified before high-risk data migrations.

60. Migration failures must be recoverable.

61. Legacy systems must have clear ownership until retirement.

62. Legacy traffic must be verified before decommissioning.

63. Legacy data must not be deleted without retention validation.

64. Legacy events must not be removed without producer and consumer validation.

65. Legacy APIs must not be removed without consumer validation.

66. Architecture changes should be validated after migration.

67. Major migrations should receive a post-migration review.

68. Migration lessons should update architecture documentation and standards.

69. The platform should evolve incrementally rather than through repeated rewrites.

70. The preferred architecture is the simplest architecture that satisfies current requirements while preserving reasonable future options.

71. Architecture evolution must protect customer experience.

72. Architecture evolution must protect driver experience.

73. Architecture evolution must protect data integrity.

74. Architecture evolution must protect operational reliability.

75. Architecture evolution must remain observable.

76. Architecture evolution must remain controlled.

77. Architecture evolution must remain reversible where technically possible.

78. The goal of migration is not technological novelty but measurable improvement.

79. Existing architecture should be replaced only when the new architecture provides sufficient value to justify migration risk.

80. RideForge should continuously evolve without continuously rewriting the platform.
```

---

# 194. Status

```text
Decision: ACCEPTED

Evolution Model:
Incremental + Compatible + Observable + Reversible Where Possible

Primary Strategy:
Expand → Migrate → Switch → Contract

Migration:
Risk-Based

Architecture Changes:
ADR-Driven

API Evolution:
Backward-Compatible Where Practical

Event Evolution:
Compatibility-First

Database Evolution:
Expand-and-Contract

Data Migration:
Controlled + Validated + Reconciled

Service Extraction:
Requirement-Driven

Service Consolidation:
Allowed When Boundaries No Longer Provide Value

Infrastructure Evolution:
Incremental

Cloud Migration:
Provider-Isolated Where Valuable

Dispatch Evolution:
Controlled Validation

AI Evolution:
ADR-0026 Governed

Regional Evolution:
Legal + Operational + Infrastructure Aware

Traffic Migration:
Gradual Where Appropriate

Canary:
Recommended for Higher-Risk Changes

Rollback:
Required Where Technically Possible

Forward Recovery:
Required for Irreversible Changes

Legacy Systems:
Owned Until Retirement

Temporary Migration Systems:
Must Have Removal Conditions

Feature Flags:
Temporary + Owned

Observability:
Required During Migration

Testing:
Required According to Risk

Security:
Required Throughout Migration

Cost:
ADR-0028

Deployment:
ADR-0027

Testing:
ADR-0025

AI Governance:
ADR-0026

Failure Strategy:
ADR-0021

Observability:
ADR-0022

Security:
ADR-0023

Configuration:
ADR-0024

Primary Goal:
Allow RideForge to Evolve Safely From an Early Production Platform Into a Larger, More Capable, More Reliable System Without Requiring Repeated Big-Bang Rewrites
```

---

# 195. Decision Summary

RideForge adopts the following architecture evolution model:

```text
                    CURRENT SYSTEM
                          │
                          ▼
                   IDENTIFY NEED
                          │
                          ▼
                   MEASURE PROBLEM
                          │
                          ▼
                 DEFINE TARGET STATE
                          │
                          ▼
                  CREATE / UPDATE ADR
                          │
                          ▼
                 DESIGN MIGRATION
                          │
                          ▼
               ┌─────────────────────┐
               │ COMPATIBILITY LAYER │
               └──────────┬──────────┘
                          │
                          ▼
                    MIGRATE DATA
                          │
                          ▼
                   MIGRATE TRAFFIC
                          │
                          ▼
                  VALIDATE NEW PATH
                          │
                ┌─────────┴─────────┐
                │                   │
             Healthy            Unhealthy
                │                   │
                ▼                   ▼
          Increase Traffic       Rollback /
                │                Recovery
                ▼
             CUTOVER
                │
                ▼
         REMOVE LEGACY PATH
                │
                ▼
        POST-MIGRATION REVIEW
                │
                ▼
           UPDATED STATE
```

The long-term architecture should therefore evolve through controlled transitions rather than repeated rewrites:

```text
Stable Domain Boundaries
        +
Compatible Contracts
        +
Incremental Migration
        +
Strong Observability
        +
Explicit Rollback
        +
Controlled Decommissioning
```

The central principle is:

> **RideForge should evolve continuously, but each evolution must be deliberate, measurable, observable, and safe enough to operate in production.**

---

# 196. Status Metadata

| Field | Value |
|---|---|
| ADR | `0029` |
| Title | Architecture Evolution and Migration |
| Status | Accepted |
| Category | Architecture / Evolution / Migration |
| Evolution Model | Incremental |
| Migration Model | Expand → Migrate → Switch → Contract |
| ADR Governance | Required |
| Backward Compatibility | Preferred |
| API Evolution | Additive Where Practical |
| Event Evolution | Compatibility-First |
| Database Evolution | Expand-and-Contract |
| Data Migration | Controlled + Reconciled |
| Service Extraction | Requirement-Driven |
| Service Consolidation | Allowed |
| Infrastructure Evolution | Incremental |
| Cloud Migration | Provider-Isolated Where Valuable |
| Dispatch Evolution | Controlled |
| AI Evolution | ADR-0026 |
| Regional Expansion | Legal + Operational + Infrastructure Aware |
| Traffic Migration | Gradual Where Appropriate |
| Canary | Risk-Based |
| Rollback | Required Where Possible |
| Forward Recovery | Required for Irreversible Changes |
| Legacy Ownership | Required Until Retirement |
| Temporary Migration Systems | Removal Condition Required |
| Feature Flags | Temporary + Owned |
| Observability | Required |
| Testing | Risk-Based and Required |
| Security | Required |
| Deployment | ADR-0027 |
| Cost | ADR-0028 |
| Failure Strategy | ADR-0021 |
| Observability | ADR-0022 |
| Security | ADR-0023 |
| Configuration | ADR-0024 |
| Testing | ADR-0025 |
| AI Governance | ADR-0026 |
| Next ADR | `0030-ADR_INDEX.md` |

---

# 24. Dispatch Strategy Evolution and Migration Clarifications

Future architecture evolution must preserve the canonical RideForge dispatch strategy model unless an explicit ADR changes it.

## 24.1 Canonical Dispatch Model

RideForge has two primary dispatch strategies:

```text
Smart Stand Dispatch
Smart Dispatch
```

AI-assisted dispatch is an optimization capability and is **not a third primary dispatch strategy**.

The architecture should therefore evolve around:

```text
Configuration Resolution
        ↓
Effective Dispatch Strategy
        ↓
Candidate Discovery
        ↓
Candidate Pipeline
        ↓
Strategy-Specific Prioritization / Ranking
        ↓
Assignment
```

AI may participate in optimization/ranking without replacing the resolved primary strategy.

---

## 24.2 Hierarchical Configuration Must Remain Extensible

Dispatch strategy configuration is hierarchical.

Possible levels include:

```text
State
District
City / Town
Rural Area
Auto Stand
Specific Ride Level
Other Intermediate Levels
```

Not every level requires configuration.

The effective strategy is resolved from the most specific applicable level upward:

```text
Most Specific Applicable Configuration
        ↓
Explicit Strategy?
   ├── YES → Use it
   └── NO
        ↓
Parent Configuration
        ↓
Continue Upward
        ↓
System Default
```

The canonical precedence rule is:

> **Specific configuration overrides inherited configuration.**

Migration work must preserve this behavior.

The hierarchy must not be migrated into a rigid, permanently hard-coded list of levels.

---

## 24.3 Smart Stand Dispatch Must Remain Stand-Preferred

Architecture migrations must not regress Smart Stand Dispatch into stand-only dispatch.

The canonical behavior is:

```text
Rider inside stand radius
        ↓
Prefer eligible drivers from the relevant stand
        ↓
Apply stand queue / ordering rules
        ↓
Suitable driver available?
   ├── YES → continue toward assignment
   └── NO  → expand candidate discovery
```

Expanded candidates may include:

```text
Drivers outside the stand
Drivers at nearby stands
Drivers from nearby locations
```

If the rider is outside the radius of all configured stands, Smart Stand Dispatch does not impose a stand-only candidate pool.

Any migration that introduces stand-only filtering is a behavioral regression.

---

## 24.4 Smart Dispatch Must Remain Stand-Agnostic

Smart Dispatch must continue to consider eligible nearby drivers without using auto-stand membership as an inherent dispatch preference.

Migration work must preserve the ability to consider:

```text
Drivers at stands
Drivers outside stands
Drivers from nearby locations
```

subject to hard eligibility, legal, service, and geographic constraints.

---

## 24.5 Cross-Location Dispatch Must Survive Architecture Changes

Location boundaries must not accidentally become hard dispatch boundaries during migration.

For example:

```text
Location A → Smart Dispatch
Location B → Smart Stand Dispatch
```

If candidate discovery expands from A to B, eligible candidates from B may be considered.

The source location's dispatch strategy provides strategy context; it does not automatically make the candidate ineligible.

Migration must preserve candidate context such as:

```text
Candidate Location
Candidate Location Strategy
Stand Membership
Relevant Stand
Queue Position
Discovery Source
Expansion Level
```

---

## 24.6 Strategy Switching Must Be Explicit

Architecture evolution must distinguish:

```text
Candidate Expansion
AI Degradation
Retry
Fallback
Strategy Switching
```

These are not interchangeable.

For example:

```text
Smart Stand Dispatch
        ↓
Preferred stand unavailable
        ↓
Broader candidate discovery
        ↓
Still Smart Stand Dispatch
```

Likewise:

```text
Smart Stand Dispatch + AI
        ↓
AI unavailable
        ↓
Deterministic Smart Stand Dispatch
```

A migration must not introduce implicit:

```text
Smart Stand Dispatch
        ↓
Smart Dispatch
```

behavior.

A strategy transition is valid only when explicitly defined by the business/configuration model.

---

## 24.7 AI Must Remain an Optimization Capability

Architecture migrations involving AI must preserve the distinction between:

```text
Primary Dispatch Strategy
```

and:

```text
AI Assistance
```

For example:

```text
strategy = SMART_STAND
ai_assisted = true
```

or:

```text
strategy = SMART
ai_assisted = true
```

AI may assist with:

```text
Ranking
ETA Prediction
Acceptance Prediction
Demand / Supply Prediction
Other Approved Optimization Signals
```

AI must not silently:

```text
Change the primary strategy
Override hard eligibility
Override legal restrictions
Override safety constraints
Replace configured stand queue semantics
```

---

## 24.8 Legal and Regional Constraints Must Survive Migration

Dispatch strategy and geographic proximity must not be treated as legal authorization.

Migration must preserve:

```text
Geographic Discovery
        ↓
Regional / Legal Validation
        ↓
Strategy-Specific Prioritization
```

The following distinctions remain authoritative:

```text
Geographic Proximity ≠ Legal Permission
Candidate Discovery ≠ Legal Authorization
Dispatch Strategy ≠ Legal Boundary
```

Cross-location candidate expansion must continue to enforce applicable regional and legal rules.

---

## 24.9 Migration Validation Requirements

Any architecture migration affecting dispatch must validate at minimum:

### Configuration

- Most-specific explicit configuration wins.
- Parent configuration is inherited when no child configuration exists.
- System default applies when nothing is configured.
- Intermediate configuration levels remain supported.

### Smart Stand Dispatch

- Stand preference applies when the rider is inside the stand radius.
- Stand queue/order semantics are preserved.
- Non-stand candidates are not automatically rejected.
- Broader candidate discovery is possible when preferred stand supply is insufficient.
- Stand radius is not treated as a universal candidate boundary.

### Smart Dispatch

- Stand membership does not create an implicit preference.
- Eligible nearby drivers remain discoverable regardless of stand membership.

### Cross-Location Dispatch

- Nearby-location candidates remain discoverable when expansion is allowed.
- Different source-location strategies do not automatically reject candidates.
- Candidate source context is preserved.

### AI

- AI assistance remains optional/independent of the primary strategy.
- AI failure preserves the primary strategy.
- Deterministic fallback remains available.
- AI cannot override hard constraints.

### Legal / Regional

- Every expanded candidate is independently validated.
- Cross-location discovery does not bypass legal rules.

---

## 24.10 Migration Guardrails

Architecture migrations must not:

```text
Convert Smart Stand Dispatch into stand-only dispatch.

Convert Smart Dispatch into stand-aware dispatch without an explicit business decision.

Treat dispatch strategy as a hard geographic candidate boundary.

Hard-code a fixed State → District → Town → Stand hierarchy.

Let parent configuration override explicit child configuration.

Treat candidate expansion as strategy switching.

Treat AI-assisted dispatch as a third primary strategy.

Allow AI failure to silently switch strategies.

Allow cross-location discovery to bypass regional/legal validation.

Discard candidate source context required for strategy-specific prioritization.

Replace configured stand queue semantics with an arbitrary AI score.

Duplicate dispatch strategy resolution across services with inconsistent precedence rules.
```

Any change that intentionally alters one of these behaviors must be documented through a new or superseding ADR.

---

## 24.11 Migration Principle

The governing migration principle is:

> **Architecture evolution may change implementation mechanisms, but it must not accidentally change the established dispatch business semantics.**

When a future architectural change intentionally changes dispatch semantics, the change must be explicitly proposed, documented, reviewed, and versioned rather than being introduced as an incidental migration side effect.

