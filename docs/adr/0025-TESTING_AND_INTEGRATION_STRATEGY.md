# ADR-0025: Testing and Integration Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Testing / Quality / Architecture / Integration  
> **Scope:** Unit testing, integration testing, contract testing, end-to-end testing, infrastructure-dependent testing, event-driven systems, database testing, API testing, AI testing, failure testing, CI/CD validation, and test environments  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is a distributed ride-hailing platform built around:

```text
Microservices
Domain-Driven Design
Event-Driven Architecture
PostgreSQL
PostGIS
Redis
Kafka / Redpanda
HTTP APIs
External Providers
AI Components
Real-Time Dispatch
```

The platform contains business-critical workflows such as:

```text
Ride Creation
Driver Availability
Driver Matching
Dispatch
ETA Calculation
Ride State Transitions
Payment
Notifications
Driver Location
Regional / Legal Validation
Event Publishing
Event Consumption
Fallback Processing
```

These workflows cannot be reliably validated through unit tests alone.

A production-grade testing strategy must validate multiple levels:

```text
Domain Logic
Application Logic
Infrastructure Integration
Service Boundaries
Event Contracts
Database Behaviour
Real-Time State
External Dependencies
Failure Handling
End-to-End Business Flows
```

The architecture therefore requires a layered testing strategy that balances:

```text
Fast Feedback
Confidence
Determinism
Production Similarity
Cost
Maintenance
```

---

# 2. Problem

Without a defined testing architecture, distributed systems commonly develop:

```text
Too Many Mock-Based Tests
Weak Integration Coverage
Brittle End-to-End Tests
Unverified Event Contracts
Database-Specific Bugs
Environment-Specific Failures
Uncovered Failure Paths
Slow CI Pipelines
False Confidence
```

RideForge needs explicit rules defining:

```text
What belongs in unit tests
What belongs in integration tests
When real infrastructure is required
How services are tested together
How events are tested
How external providers are tested
How failures are tested
What must run in CI
```

---

# 3. Decision

RideForge will use a layered testing strategy:

```text
                 END-TO-END
                     │
             CONTRACT / SYSTEM
                     │
              INTEGRATION
                     │
                   UNIT
                     │
                  DOMAIN
```

The primary strategy is:

```text
Fast Unit Tests
+
Realistic Integration Tests
+
Contract Tests
+
Selective End-to-End Tests
+
Failure / Resilience Tests
```

The platform will prefer **real infrastructure for infrastructure behaviour** rather than replacing every dependency with mocks.

---

# 4. Core Testing Principles

RideForge follows:

```text
Test Behaviour, Not Implementation
Fast Feedback First
Real Infrastructure Where Behaviour Matters
Deterministic Tests
Isolated Tests
Production-Relevant Integration
Explicit Failure Testing
Contract Verification
Repeatability
```

---

# 5. Testing Pyramid

The platform will follow a practical testing pyramid.

```text
                 ┌───────────────┐
                 │   E2E Tests   │
                 │   Few / High  │
                 └───────┬───────┘
                         │
                ┌────────▼────────┐
                │ Contract/System  │
                │     Tests        │
                └────────┬─────────┘
                         │
              ┌──────────▼──────────┐
              │ Integration Tests   │
              │ More / Real Infra   │
              └──────────┬──────────┘
                         │
             ┌───────────▼───────────┐
             │     Unit Tests        │
             │      Many / Fast      │
             └───────────┬───────────┘
                         │
                ┌────────▼────────┐
                │ Domain Invariants│
                └─────────────────┘
```

---

# 6. Test Categories

RideForge testing is divided into:

```text
1. Domain Tests
2. Unit Tests
3. Integration Tests
4. Contract Tests
5. API Tests
6. Event Tests
7. Database Tests
8. Infrastructure Tests
9. End-to-End Tests
10. Failure / Resilience Tests
11. Performance Tests
12. Security Tests
13. AI / Model Tests
```

---

# 7. Domain Tests

Domain tests validate business rules without infrastructure dependencies.

Examples:

```text
Ride State Transitions
Ride Eligibility
Driver Eligibility
Regional Rules
Matching Constraints
Fare Rules
Assignment Rules
```

These should be fast and deterministic.

---

# 8. Domain Invariants

Critical business invariants must have direct tests.

Examples:

```text
Completed Ride Cannot Return to Requested
Cancelled Ride Cannot Be Assigned
Unauthorized Driver Cannot Accept Ride
Illegal Regional Ride Cannot Be Created
Invalid State Transition Is Rejected
```

---

# 9. Domain Test Characteristics

Domain tests should generally:

```text
Avoid Database
Avoid Redis
Avoid Kafka
Avoid HTTP
Avoid External Providers
Avoid Real Network
```

They should test domain behaviour directly.

---

# 10. Unit Tests

Unit tests validate isolated application components.

Examples:

```text
Service Methods
Validators
Mappers
Policies
Rankers
Retry Calculators
Configuration Validators
Event Serializers
```

---

# 11. Unit Test Dependencies

Dependencies may be replaced with controlled test doubles when testing isolated logic.

Examples:

```text
Repository Interface
Provider Interface
Clock
ID Generator
Event Publisher
```

---

# 12. Mock Usage

Mocks are allowed but should not become the primary strategy for testing infrastructure behaviour.

Use mocks when the purpose is:

```text
Isolate Logic
Control Failure
Verify Interaction
Test Rare Conditions
```

---

# 13. Mock Abuse

Avoid tests that only prove:

```text
Method A
called Mock B
with Parameter C
```

while never verifying meaningful business behaviour.

---

# 14. Integration Tests

Integration tests verify real interaction between application code and infrastructure.

Examples:

```text
Service + PostgreSQL
Service + Redis
Service + Kafka / Redpanda
Service + PostGIS
Service + HTTP Provider Stub
```

---

# 15. Real Infrastructure Principle

When behaviour depends on infrastructure semantics, use the real infrastructure.

Examples:

```text
PostgreSQL SQL Behaviour
PostGIS Queries
Redis TTL
Kafka Delivery
Transaction Semantics
Database Constraints
```

These should not be validated solely through mocks.

---

# 16. PostgreSQL Integration Tests

PostgreSQL integration tests should validate:

```text
Queries
Transactions
Constraints
Indexes
Unique Rules
Foreign Keys
Pagination
Locking Where Applicable
```

---

# 17. PostGIS Tests

Geospatial behaviour should be tested against actual PostGIS.

Examples:

```text
Nearby Driver Query
Radius Search
Distance Calculation
Geospatial Filtering
Spatial Index Behaviour
```

---

# 18. Database Transaction Tests

Transactions should be tested using real PostgreSQL where transaction semantics matter.

Examples:

```text
Commit
Rollback
Concurrent Update
Constraint Failure
Partial Failure
```

---

# 19. Redis Integration Tests

Redis integration tests should validate actual Redis semantics.

Examples:

```text
TTL
Expiration
Atomic Operations
Key Structure
Caching
Real-Time State
Distributed Locking Where Used
```

---

# 20. Redis Geospatial Tests

If Redis geospatial functionality is used, test:

```text
Location Write
Location Update
Radius Search
Expiration
Removal
```

against real Redis.

---

# 21. Kafka / Redpanda Integration Tests

Messaging integration tests should validate:

```text
Producer
Consumer
Topic
Consumer Group
Serialization
Deserialization
Acknowledgement
Retry
DLQ
Ordering Where Required
```

---

# 22. Event Delivery Tests

Test important event flows such as:

```text
Ride Created
→ Event Published
→ Consumer Receives
→ State Updated
```

---

# 23. Outbox Tests

The Outbox Pattern must be tested as a transactional workflow.

Validate:

```text
Business Transaction
+
Outbox Insert
```

commit together.

Also validate:

```text
Business Failure
→ No Invalid Outbox Event
```

and:

```text
Publish Failure
→ Outbox Event Remains Available
```

---

# 24. Dead Letter Queue Tests

DLQ behaviour must be tested.

Examples:

```text
Invalid Event
Repeated Processing Failure
Retry Exhaustion
DLQ Publication
DLQ Metadata
```

---

# 25. Event Serialization Tests

Event payloads must be tested for:

```text
Required Fields
Types
Schema
Version
Compatibility
```

---

# 26. Event Contract Tests

Event producers and consumers must agree on event contracts.

Contract tests should validate:

```text
Event Name
Envelope
Schema
Required Fields
Version
Semantic Meaning
```

---

# 27. API Tests

API tests should validate:

```text
Status Codes
Request Validation
Response Schema
Authentication
Authorization
Error Responses
Pagination
Idempotency
```

---

# 28. API Contract Tests

Public and internal APIs should have contract validation where multiple services depend on them.

---

# 29. API Error Tests

Every important API should test:

```text
400
401
403
404
409
422
429
500
503
```

where applicable.

The exact response codes depend on the API contract.

---

# 30. Authentication Tests

Test:

```text
Valid Credentials
Invalid Credentials
Expired Token
Revoked Token
Missing Token
Malformed Token
```

---

# 31. Authorization Tests

Test:

```text
Allowed Access
Denied Access
Cross-User Access
Cross-Driver Access
Administrative Access
Resource Ownership
```

---

# 32. Idempotency Tests

Critical operations must test repeated requests.

Example:

```text
Create Ride
Request 1 → Creates Ride

Same Idempotency Key
Request 2 → Does Not Create Duplicate Ride
```

---

# 33. Concurrent Request Tests

Important state-changing operations should be tested under concurrency.

Examples:

```text
Two Drivers Accept Same Ride
Two Assignment Attempts
Duplicate Payment Callback
Concurrent Ride Cancellation
```

---

# 34. Race Condition Testing

Where state is updated concurrently, tests should attempt to expose:

```text
Lost Updates
Duplicate Assignments
Invalid State Transitions
Double Processing
```

---

# 35. Driver Availability Tests

Driver state must be tested across:

```text
Online
Offline
Available
Busy
Reserved
Suspended
```

according to the implemented domain model.

---

# 36. Driver Location Tests

Location systems should test:

```text
Location Update
Stale Location
Invalid Coordinates
Out-of-Bounds Coordinates
Location Expiration
Nearby Search
```

---

# 37. Matching Tests

Matching logic must be tested independently from infrastructure and with integration tests against real candidate data.

---

# 38. Smart Dispatch Tests

Smart dispatch tests should validate:

```text
Candidate Generation
Feature Availability
Ranking
Hard Constraints
Assignment
Fallback
```

---

# 39. Stand Dispatch Tests

Stand dispatch tests should validate:

```text
Stand Selection
Queue Order
Driver Availability
Assignment
Fallback
```

---

# 40. Hybrid Dispatch Tests

Because RideForge supports both dispatch strategies, test:

```text
Stand Mode
Smart Mode
Regional Selection
Mode Switching
Fallback
```

---

# 41. Dispatch Safety Tests

AI or ranking must never bypass hard constraints.

Test scenarios such as:

```text
AI Ranks Ineligible Driver First
→ Driver Must Still Be Rejected
```

---

# 42. ETA Tests

ETA systems should test:

```text
Input Validation
Route Provider Response
Prediction
Fallback
Timeout
Invalid Provider Response
```

---

# 43. Routing Provider Tests

External routing providers should normally be replaced by controlled provider stubs in integration tests.

Test:

```text
Success
Timeout
5xx
Invalid Response
Slow Response
Rate Limit
Provider Unavailable
```

---

# 44. Payment Integration Tests

Payment workflows should test:

```text
Payment Initiation
Success
Failure
Timeout
Duplicate Callback
Webhook Verification
Refund
```

where applicable.

---

# 45. External Provider Strategy

External providers should not be called against production from automated tests.

Use:

```text
Provider Sandbox
Controlled Stub
Mock Server
Recorded Fixture
```

according to provider capabilities.

---

# 46. Test Doubles

Approved test doubles include:

```text
Stub
Fake
Mock
Spy
Simulator
Provider Sandbox
```

Each should be used for an explicit purpose.

---

# 47. Test Double Preference

Prefer:

```text
Real Infrastructure
```

for infrastructure semantics, and:

```text
Stub / Fake / Mock
```

for external systems that are expensive, unavailable, or nondeterministic.

---

# 48. Test Fixtures

Fixtures should represent meaningful domain states.

Examples:

```text
Available Driver
Assigned Driver
Requested Ride
Active Ride
Completed Ride
Cancelled Ride
```

---

# 49. Test Data

Test data should be:

```text
Deterministic
Minimal
Representative
Non-Sensitive
Reproducible
```

---

# 50. Production Data

Production data should not be copied directly into automated test environments.

If production-like data is required, it should be:

```text
Synthetic
Anonymized
Sanitized
```

as appropriate.

---

# 51. Database Isolation

Integration tests should isolate their data.

Approaches may include:

```text
Transactional Rollback
Dedicated Test Database
Per-Test Schema
Containerized Database
Truncate / Reset
```

The approach should be selected based on test behaviour and performance.

---

# 52. Parallel Tests

Tests may run in parallel when isolation is guaranteed.

Tests sharing mutable global state should not be parallelized without safeguards.

---

# 53. Test Determinism

Tests must avoid uncontrolled dependencies on:

```text
Current Time
Random Values
External Network
Machine State
Execution Order
```

unless those behaviours are explicitly under test.

---

# 54. Time Injection

Time-dependent code should support a controlled clock.

Examples:

```text
Token Expiration
Ride Timeout
Driver Staleness
ETA
Retry Backoff
```

---

# 55. Randomness

Random values should be controllable in tests.

---

# 56. IDs

Where practical, tests should use deterministic IDs or injectable ID generators.

---

# 57. Integration Environment

Local integration testing should use infrastructure matching production technology where practical.

For example:

```text
PostgreSQL
Redis
Kafka / Redpanda
```

should be tested using the same major technologies used in production.

---

# 58. Docker-Based Integration Infrastructure

Docker Compose or equivalent local infrastructure may be used to run:

```text
PostgreSQL
PostGIS
Redis
Kafka / Redpanda
```

for integration tests.

---

# 59. Test Infrastructure Lifecycle

Test infrastructure should be:

```text
Provisioned
Ready
Tested
Cleaned Up
```

automatically where practical.

---

# 60. Health Checks

Integration test infrastructure should provide readiness checks before tests begin.

---

# 61. Database Migration Testing

Migrations must be executed against real PostgreSQL during integration validation.

Test:

```text
Fresh Database
Existing Database
Upgrade Path
Constraints
Indexes
```

where applicable.

---

# 62. Migration Rollback

If rollback migrations exist, they should be tested.

---

# 63. Schema Compatibility

Database changes should be validated for compatibility with application versions that may temporarily coexist during deployment.

---

# 64. Event Schema Compatibility

Event changes should be tested against existing consumers where backward compatibility is required.

---

# 65. Consumer Compatibility

When an event producer changes, test:

```text
New Producer
+
Existing Consumer
```

and where necessary:

```text
Existing Producer
+
New Consumer
```

---

# 66. API Compatibility

API changes should be evaluated for:

```text
Backward Compatibility
Client Impact
Consumer Impact
```

---

# 67. Contract Testing Strategy

Contract tests should focus on boundaries where independent components communicate.

Examples:

```text
Service → Service
Producer → Consumer
API → Client
Service → External Provider
```

---

# 68. Contract Ownership

The team responsible for changing a contract must identify affected consumers.

---

# 69. Breaking Changes

Breaking contract changes require:

```text
Migration Plan
Versioning Where Appropriate
Consumer Updates
Validation
```

---

# 70. End-to-End Testing

End-to-end tests validate complete business workflows.

They should be limited to high-value flows.

---

# 71. Core E2E Flows

At minimum, important ride lifecycle scenarios should be represented.

Example:

```text
Customer Requests Ride
        ↓
Ride Created
        ↓
Candidate Drivers Found
        ↓
Driver Selected
        ↓
Driver Accepts
        ↓
Ride Starts
        ↓
Ride Completes
```

---

# 72. E2E Cancellation Flow

Test:

```text
Ride Requested
→ Driver Assigned
→ Ride Cancelled
→ State Updated
→ Required Events Published
```

---

# 73. E2E Failure Flow

Test important degraded workflows.

Example:

```text
Primary Dispatch Fails
→ Fallback
→ Driver Still Assigned
```

where such fallback is part of the approved system design.

---

# 74. E2E Regional Validation

Test:

```text
Allowed Ride
→ Accepted

Disallowed Ride
→ Rejected
```

according to configured and legally approved regional rules.

---

# 75. E2E Hybrid Dispatch

Test both:

```text
Stand Dispatch
Smart Dispatch
```

in representative operating contexts.

---

# 76. E2E External Provider Flow

Where applicable:

```text
Application
→ Provider
→ Callback
→ State Update
```

should be tested through controlled environments.

---

# 77. E2E Test Quantity

Avoid creating hundreds of E2E tests for every small variation.

Use lower-level tests for detailed edge cases.

---

# 78. Failure Testing

Failure testing is mandatory for critical dependencies.

Test:

```text
Database Failure
Redis Failure
Kafka Failure
Provider Failure
Timeout
Network Failure
Invalid Event
Duplicate Event
```

---

# 79. Database Failure

Test application behaviour when:

```text
Connection Fails
Query Times Out
Transaction Fails
Database Becomes Unavailable
```

---

# 80. Redis Failure

Test:

```text
Redis Unavailable
Redis Timeout
Cache Miss
State Read Failure
```

and verify the approved fallback.

---

# 81. Kafka / Redpanda Failure

Test:

```text
Broker Unavailable
Publish Failure
Consumer Failure
Repeated Processing Failure
```

---

# 82. Event Duplicate Testing

Consumers should be tested with duplicate events.

Expected behaviour should follow:

```text
Idempotent Processing
```

where required.

---

# 83. Event Reordering

Where event ordering is not guaranteed globally, test out-of-order delivery where the domain requires protection.

---

# 84. Event Delay

Test delayed event processing.

The system should distinguish:

```text
Delayed
```

from:

```text
Lost
```

where the architecture supports recovery.

---

# 85. DLQ Recovery Testing

Test:

```text
Failure
→ Retry
→ DLQ
→ Corrective Action
→ Replay
```

where replay is supported.

---

# 86. Outbox Recovery Testing

Test:

```text
Business Transaction Succeeds
+
Publisher Fails
```

and verify that the outbox record remains available for later processing.

---

# 87. Connection Pool Testing

Test database behaviour under realistic concurrency.

Particularly validate:

```text
Pool Limits
PgBouncer
Database Connection Limits
Timeouts
```

---

# 88. Concurrency Load Tests

Critical operations should be tested under concurrent requests.

Examples:

```text
Ride Creation
Driver Acceptance
Matching
Location Updates
Payment Callback
```

---

# 89. Performance Testing

Performance tests should validate:

```text
Latency
Throughput
Resource Usage
Connection Usage
Event Processing Rate
```

---

# 90. Performance Test Isolation

Performance testing should be separated from ordinary functional CI tests.

---

# 91. Load Testing

Load tests should represent realistic traffic patterns.

Examples:

```text
Ride Requests
Location Updates
Driver Availability
Dispatch
```

---

# 92. Spike Testing

Spike tests may be used to validate sudden demand increases.

Example:

```text
Normal Traffic
→ Sudden Demand Surge
```

---

# 93. Soak Testing

Long-running tests may validate:

```text
Memory Leaks
Connection Leaks
Consumer Lag
Resource Exhaustion
```

---

# 94. Security Testing

Security tests should include:

```text
Authentication
Authorization
Input Validation
Rate Limiting
Secret Exposure
Dependency Vulnerabilities
```

---

# 95. API Security Tests

Test:

```text
Unauthorized Request
Privilege Escalation
Resource Enumeration
Malformed Input
Rate Limit
```

---

# 96. Event Security Tests

Validate:

```text
Unauthorized Producer
Unauthorized Consumer
Malformed Event
Unexpected Event Source
```

where infrastructure supports access controls.

---

# 97. AI Testing

AI-enabled components require additional testing.

Examples:

```text
Feature Validation
Model Output Validation
Ranking Quality
Fallback
Latency
Drift
Safety Constraints
```

---

# 98. AI Hard Constraints

AI tests must verify that model output cannot bypass deterministic business constraints.

---

# 99. AI Ranking Tests

Test:

```text
Valid Candidate
Invalid Candidate
Missing Feature
Unexpected Model Output
Model Timeout
Model Unavailable
```

---

# 100. AI Fallback Tests

Test:

```text
AI Available
AI Timeout
AI Error
AI Unavailable
```

and verify fallback behaviour.

---

# 101. ETA Model Tests

ETA systems should be tested for:

```text
Prediction Availability
Prediction Bounds
Fallback
Provider Failure
Unexpected Input
```

---

# 102. Model Regression Tests

Approved model versions should have regression datasets where appropriate.

---

# 103. AI Data Pipeline Tests

Data pipelines should test:

```text
Schema
Missing Values
Invalid Values
Feature Generation
Data Freshness
```

---

# 104. Feature Tests

Important model features should have validation for:

```text
Type
Range
Freshness
Availability
Meaning
```

---

# 105. Model Serving Tests

Model-serving endpoints should be tested for:

```text
Request Schema
Response Schema
Latency
Timeout
Invalid Input
Unavailable Model
```

---

# 106. AI Governance

AI testing follows:

```text
ADR-0026 — Model and AI Governance
```

---

# 107. Test Naming

Test names should clearly describe behaviour.

Prefer:

```text
RejectsRideWhenCrossRegionOperationIsNotAllowed
```

over:

```text
TestRide1
```

---

# 108. Test Structure

Tests should generally follow:

```text
Arrange
Act
Assert
```

or an equivalent readable structure.

---

# 109. Assertion Quality

Assertions should verify meaningful outcomes.

Avoid tests that pass without validating the important state transition.

---

# 110. Test Isolation

Each test should be independently understandable and should not depend on another test's execution.

---

# 111. Test Cleanup

Tests must clean up:

```text
Database Data
Redis Keys
Topics / Consumer Groups Where Necessary
Temporary Files
Resources
```

---

# 112. Test Data Cleanup

Prefer isolated test environments over fragile cleanup assumptions.

---

# 113. Test Parallelism

Parallel execution should be enabled where:

```text
Isolation
Determinism
Resource Capacity
```

permit it.

---

# 114. Flaky Tests

Flaky tests are treated as defects.

Do not permanently hide flakiness with:

```text
Retries
Skipped Tests
Disabled Tests
```

without tracking the underlying problem.

---

# 115. Test Retry

CI retries may be used temporarily for infrastructure instability, but repeated test retries must not conceal application defects.

---

# 116. Test Timeouts

Every integration and E2E test suite should have reasonable timeouts.

A test that hangs indefinitely is itself a failure.

---

# 117. Test Observability

Test infrastructure should expose enough information to diagnose failures.

Examples:

```text
Application Logs
Database Logs
Broker Logs
Container Logs
Test Output
```

---

# 118. Test Failure Artifacts

CI should retain useful failure artifacts where appropriate.

Examples:

```text
Logs
Screenshots
Trace
Request / Response Metadata
Container Logs
```

Sensitive data must be redacted.

---

# 119. CI Testing Layers

CI should execute tests in stages.

Conceptually:

```text
Commit
  ↓
Format / Static Checks
  ↓
Unit Tests
  ↓
Integration Tests
  ↓
Contract Tests
  ↓
Build
  ↓
Selected E2E
```

---

# 120. Fast Feedback

Fast tests should run first.

Expensive suites should run only after earlier validation passes unless there is a specific reason otherwise.

---

# 121. Pull Request Testing

Pull requests should validate at least:

```text
Formatting
Linting
Unit Tests
Relevant Integration Tests
Build
```

---

# 122. Main Branch Testing

Main branch validation should include stronger checks such as:

```text
Integration
Contract
Selected E2E
Security
```

as appropriate.

---

# 123. Release Testing

Release validation should include:

```text
Full Integration
Contract
Critical E2E
Security
Migration
Deployment Validation
```

---

# 124. Nightly Testing

Long-running suites may run on a scheduled basis.

Examples:

```text
Soak Tests
Large Integration Matrix
Extended E2E
Dependency Scanning
```

---

# 125. Test Matrix

The test matrix should cover important combinations such as:

```text
Dispatch Mode
Region
Provider
Environment
Feature Flag
```

without creating unnecessary combinatorial explosion.

---

# 126. Risk-Based Testing

Not every combination requires the same depth of testing.

Higher-risk areas receive stronger coverage.

Examples:

```text
Payments
Authentication
Dispatch
Ride State
Legal Validation
Data Consistency
```

---

# 127. Coverage

Code coverage is a signal, not the objective.

A high percentage does not guarantee correct behaviour.

---

# 128. Coverage Priorities

Prioritize coverage for:

```text
Domain Rules
State Machines
Transactions
Authorization
Payments
Dispatch
Event Processing
Failure Handling
```

---

# 129. Coverage Thresholds

Coverage thresholds may be introduced for important packages, but thresholds should not encourage meaningless tests.

---

# 130. Mutation Testing

Mutation testing may be used for critical domain logic to determine whether tests detect meaningful behavioural changes.

---

# 131. Test Review

Tests are production code and should receive code review.

Review for:

```text
Correctness
Isolation
Readability
Coverage
Failure Cases
Maintenance Cost
```

---

# 132. Testing and Architecture

Architecture changes should include appropriate tests.

Examples:

```text
New Service
→ Service Tests

New Event
→ Event Contract Tests

New Database Repository
→ Integration Tests

New Provider
→ Provider Integration Tests

New Dispatch Strategy
→ Domain + Integration + E2E Tests
```

---

# 133. New Microservice Test Requirements

A new service should normally include:

```text
Unit Tests
Integration Tests
API Tests if applicable
Event Tests if applicable
Configuration Tests
Health Tests
Failure Tests
```

---

# 134. New Database Repository

A repository implementation should include integration tests against real PostgreSQL when SQL behaviour is material.

---

# 135. New Redis Repository

A Redis repository should include integration tests against real Redis.

---

# 136. New Kafka Consumer

A consumer should include tests for:

```text
Valid Event
Invalid Event
Duplicate Event
Retry
DLQ
```

where applicable.

---

# 137. New External Provider

A provider integration should include:

```text
Success
Failure
Timeout
Invalid Response
Authentication Failure
Rate Limit
```

as applicable.

---

# 138. New API Endpoint

An endpoint should include:

```text
Happy Path
Validation
Authentication
Authorization
Not Found
Conflict
Failure
```

as applicable.

---

# 139. New Domain Rule

A new domain rule should have direct domain tests.

---

# 140. New Dispatch Algorithm

A new dispatch algorithm should be tested against:

```text
Candidate Eligibility
Ranking
Assignment
Concurrency
Fallback
```

---

# 141. Migration Testing

Every migration should be validated against:

```text
Fresh Schema
Representative Existing Schema
Application Compatibility
```

where applicable.

---

# 142. Backward Compatibility

During rolling deployments, old and new application versions may temporarily coexist.

Tests should verify compatibility for changes that cross deployment boundaries.

---

# 143. Expand-and-Contract Strategy

For incompatible schema changes, prefer:

```text
Expand
→ Migrate
→ Deploy Compatibility
→ Switch
→ Contract
```

rather than immediate destructive changes.

---

# 144. Event Migration

Event contract migrations should similarly allow compatible transition periods where necessary.

---

# 145. Test Environment Security

Test infrastructure must not expose:

```text
Production Secrets
Production Credentials
Sensitive Production Data
```

---

# 146. Test Environment Cost

Integration infrastructure should be designed to minimize cost while preserving meaningful behaviour.

---

# 147. Local vs CI Infrastructure

Local and CI integration environments should use compatible versions of:

```text
PostgreSQL
PostGIS
Redis
Kafka / Redpanda
```

where practical.

---

# 148. Version Pinning

Infrastructure versions used for integration tests should be controlled.

Avoid silently pulling arbitrary latest versions.

---

# 149. Infrastructure Upgrade Testing

Upgrades to:

```text
PostgreSQL
Redis
Kafka / Redpanda
Go
Libraries
```

should include integration validation.

---

# 150. Compatibility Testing

Major dependency upgrades should validate:

```text
Application
Database
Messaging
Serialization
Performance
```

as appropriate.

---

# 151. Test Reporting

CI should report:

```text
Passed
Failed
Skipped
Duration
Coverage
Relevant Artifacts
```

---

# 152. Test Failure Classification

Failures should be classified where practical:

```text
Application Defect
Test Defect
Infrastructure Failure
Dependency Failure
Flaky Test
Environment Failure
```

---

# 153. Test Ownership

Every important test suite should have an owner.

Examples:

```text
Domain Tests → Domain Team
Database Tests → Service / Platform Team
Messaging Tests → Event / Platform Team
AI Tests → AI Team
E2E Tests → Product / QA / Engineering
```

---

# 154. Test Maintenance

Tests should be updated when:

```text
Business Rules Change
API Changes
Event Schema Changes
Database Schema Changes
Provider Changes
Architecture Changes
```

---

# 155. Obsolete Tests

Tests that validate removed behaviour should be deleted rather than left as misleading historical artifacts.

---

# 156. Test Documentation

Complex integration suites should document:

```text
Prerequisites
Infrastructure
Execution
Cleanup
Troubleshooting
```

---

# 157. Local Test Commands

Each service should expose predictable commands for:

```text
Unit Tests
Integration Tests
All Tests
Coverage
```

The exact command depends on the service language and tooling.

---

# 158. Go Testing

Go services should use the standard Go testing ecosystem and project-approved tooling.

Typical layers:

```text
go test
Package Tests
Integration Tests
Race Detection
Benchmarks
```

where appropriate.

---

# 159. Race Detector

Concurrency-sensitive Go services should use race detection during appropriate CI or scheduled validation.

---

# 160. Benchmarks

Benchmarks may be added for:

```text
Matching
Serialization
Geospatial Processing
Hot Paths
Ranking
```

when performance matters.

---

# 161. API Integration Environment

API integration tests should run against the service with realistic middleware:

```text
Router
Authentication
Validation
Database
Relevant Dependencies
```

rather than bypassing the entire application stack.

---

# 162. Event Integration Environment

Event integration tests should run through the actual producer/consumer mechanisms where feasible.

---

# 163. Database Seed Strategy

Seed data should be:

```text
Minimal
Deterministic
Versioned
Purposeful
```

---

# 164. Test Scenario Catalog

Critical workflows should have named scenarios.

Examples:

```text
Successful Ride Assignment
No Driver Available
Driver Accept Race
Regional Ride Rejection
Dispatch Timeout
Provider Failure
Duplicate Event
Outbox Recovery
```

---

# 165. Regression Suite

Every production defect that can reasonably be reproduced should result in a regression test.

---

# 166. Production Incident → Test

The workflow should be:

```text
Incident
  ↓
Root Cause
  ↓
Reproduction
  ↓
Regression Test
  ↓
Fix
  ↓
Verification
```

---

# 167. Failure Injection

Controlled failure injection may be used to validate resilience.

Examples:

```text
Kill Consumer
Block Database
Delay Provider
Drop Broker Connection
Expire Redis
```

---

# 168. Chaos Testing

Chaos testing may be introduced after core resilience tests are mature.

It should not replace deterministic failure tests.

---

# 169. Test Safety

Failure injection must be restricted to:

```text
Local
Test
Controlled Staging
```

unless an explicitly approved production exercise exists.

---

# 170. Staging Validation

Staging should validate:

```text
Deployment
Configuration
Networking
Service Discovery
Database
Messaging
Observability
```

before production where practical.

---

# 171. Production Smoke Tests

After deployment, a limited smoke suite may verify:

```text
Service Health
Authentication
Critical API
Database Connectivity
Messaging
Core Ride Flow
```

without creating unsafe real-world side effects.

---

# 172. Test Data Cleanup in Staging

Staging test data should be identifiable and removable.

---

# 173. Test Isolation from Real Users

Staging and test environments must not accidentally send:

```text
Real Notifications
Real Payment Charges
Real Customer Messages
```

unless explicitly intended and controlled.

---

# 174. Provider Sandbox

External providers should use sandbox/test credentials whenever available.

---

# 175. Notification Testing

Notification systems should support test destinations.

Examples:

```text
Test Phone
Test Email
Sandbox Account
```

---

# 176. Payment Testing

Payment testing should use provider-approved test instruments and environments.

---

# 177. Location Testing

Location tests should use synthetic coordinates representing:

```text
Urban
Suburban
Rural
Boundary
High-Density
Low-Density
```

scenarios where relevant.

---

# 178. Regional Testing

Regional test data should include:

```text
Allowed
Restricted
Boundary
```

cases.

---

# 179. Legal Rule Testing

Legal/operational constraints should have explicit deterministic tests.

These tests should not depend solely on AI or external services.

---

# 180. Test Strategy for Hybrid Dispatch

The minimum test layers are:

```text
Stand Dispatch
→ Domain
→ Unit
→ Integration
→ E2E

Smart Dispatch
→ Domain
→ Unit
→ AI / Ranking
→ Integration
→ E2E

Hybrid Selection
→ Configuration
→ Regional Policy
→ Integration
→ E2E
```

---

# 181. Test Strategy for ETA

```text
ETA Domain Logic
→ Unit

Route Provider Integration
→ Integration

Prediction Model
→ Model Tests

Fallback
→ Failure Tests

Complete Ride ETA
→ E2E
```

---

# 182. Test Strategy for Driver Location

```text
Location Validation
→ Unit

Redis / Location Store
→ Integration

Geospatial Search
→ Integration

Stale Location
→ Failure / Integration

Dispatch Consumption
→ E2E
```

---

# 183. Test Strategy for Event-Driven Workflows

```text
Domain Event
→ Unit

Serialization
→ Unit / Contract

Outbox
→ Integration

Broker
→ Integration

Consumer
→ Integration

Retry / DLQ
→ Failure Tests

Complete Workflow
→ E2E
```

---

# 184. Test Strategy for PostgreSQL + PgBouncer

```text
Repository
→ Unit / Integration

SQL
→ Integration

Transactions
→ Integration

Pool
→ Integration

PgBouncer
→ Infrastructure Integration

Concurrency
→ Load / Integration
```

---

# 185. Test Strategy for External Providers

```text
Provider Adapter
→ Unit

Provider Protocol
→ Integration / Sandbox

Timeout
→ Failure

Invalid Response
→ Failure

Fallback
→ Integration / E2E
```

---

# 186. Testing Anti-Patterns

Avoid:

```text
Testing Only Happy Paths
Mocking Everything
Ignoring Integration Tests
Ignoring Failure Tests
Massive E2E Suites
Flaky Tests
Hard-Coded Test Order
Production Data in Tests
Real Payment Charges
Real Customer Notifications
Ignoring Event Contracts
Ignoring Database Semantics
```

---

# 187. Consequences

## 187.1 Positive Consequences

This strategy provides:

```text
Higher Confidence
Fast Developer Feedback
Realistic Infrastructure Validation
Safer Distributed-System Changes
Better Failure Coverage
Improved Regression Protection
Better Deployment Confidence
```

---

## 187.2 Negative Consequences

The strategy introduces:

```text
More Test Infrastructure
Longer CI Pipelines
Infrastructure Maintenance
More Test Data Management
Integration Test Complexity
Higher CI Resource Usage
```

These costs are accepted because RideForge is a distributed, real-time system where unit tests alone are insufficient.

---

# 188. Risks

## Risk 1 — Integration Tests Become Slow

### Mitigation

```text
Parallel Execution
Test Isolation
Focused Integration Suites
Reusable Infrastructure
Fast Health Checks
```

---

## Risk 2 — Tests Become Flaky

### Mitigation

```text
Deterministic Data
Controlled Time
Realistic Readiness Checks
Proper Cleanup
Stable Infrastructure Versions
```

---

## Risk 3 — Too Many Mocks

### Mitigation

Use real infrastructure for infrastructure semantics.

---

## Risk 4 — Too Many E2E Tests

### Mitigation

Keep E2E coverage focused on high-value business workflows.

---

## Risk 5 — Test Environment Differs From Production

### Mitigation

Use production-relevant technologies and controlled infrastructure versions.

---

## Risk 6 — Test Coverage Becomes a Vanity Metric

### Mitigation

Prioritize:

```text
Behaviour
Business Risk
Failure Modes
Critical Invariants
```

over raw percentage.

---

# 189. Alternatives Considered

## 189.1 Unit Tests Only

### Advantages

```text
Fast
Simple
Cheap
```

### Disadvantages

```text
Misses Database Bugs
Misses Broker Bugs
Misses Integration Bugs
Misses Configuration Problems
```

### Decision

```text
Rejected.
```

---

# 190. Mock Everything

### Advantages

```text
Fast
Isolated
```

### Disadvantages

```text
Mocks Do Not Reproduce Real Infrastructure Semantics
False Confidence
```

### Decision

```text
Rejected as the primary strategy.
```

---

# 191. End-to-End Tests Only

### Advantages

```text
High-Level Realism
```

### Disadvantages

```text
Slow
Brittle
Difficult Debugging
Expensive
```

### Decision

```text
Rejected.
```

---

# 192. Manual Testing Only

### Advantages

```text
Low Initial Automation Cost
```

### Disadvantages

```text
Not Reproducible
Slow
Poor Regression Protection
Human Error
```

### Decision

```text
Rejected.
```

---

# 193. Production Testing as Primary Validation

### Advantages

```text
Real Environment
```

### Disadvantages

```text
Safety Risk
Data Risk
Customer Impact
```

### Decision

```text
Rejected.
```

---

# 194. Validation

This ADR should be validated through:

```text
Unit Test Execution
Integration Test Execution
Contract Test Execution
E2E Test Execution
Failure Injection
Migration Testing
Security Testing
Performance Testing
CI Validation
Staging Validation
Production Smoke Testing
```

---

# 195. Review Triggers

Revisit this ADR when:

```text
A New Service Is Added
A New Infrastructure Dependency Is Added
A New Event Contract Is Added
A New Database Is Added
A New Dispatch Strategy Is Added
A New AI Model Is Added
A New External Provider Is Added
CI Architecture Changes
Testing Infrastructure Changes
A Major Production Incident Reveals a Testing Gap
```

---

# 196. Final Principles

The following principles are mandatory:

```text
1. RideForge uses layered testing rather than relying on one test type.

2. Domain invariants must have direct deterministic tests.

3. Unit tests should be fast and focused on isolated behaviour.

4. Mocks must not replace real infrastructure testing where infrastructure semantics matter.

5. PostgreSQL behaviour must be tested against real PostgreSQL.

6. PostGIS behaviour must be tested against real PostGIS.

7. Redis behaviour must be tested against real Redis.

8. Kafka / Redpanda behaviour must be tested against the real messaging infrastructure where practical.

9. Event contracts must be explicitly tested.

10. Outbox behaviour must be integration tested.

11. DLQ and retry behaviour must be tested.

12. APIs must test validation, authentication, authorization, and important error paths.

13. Critical operations must test idempotency.

14. Concurrent state changes must be tested where race conditions are possible.

15. External providers must not be called against production by automated tests.

16. External provider failures must be explicitly tested.

17. Production data must not be copied directly into test environments.

18. Test data must be deterministic and non-sensitive.

19. Tests should be isolated and independently reproducible.

20. Time-dependent logic should support controlled time in tests.

21. Flaky tests are defects and must not be permanently hidden.

22. Integration infrastructure must be version-controlled and reproducible where practical.

23. Database migrations must be tested against real PostgreSQL.

24. Event schema changes must be compatibility-tested.

25. API contract changes must be compatibility-tested where required.

26. High-value ride lifecycle flows must have end-to-end coverage.

27. Stand Dispatch and Smart Dispatch must both be tested.

28. Hybrid dispatch selection must be tested.

29. AI output must be tested against hard business constraints.

30. AI failure and fallback paths must be tested.

31. ETA prediction and provider fallback must be tested.

32. Driver location and geospatial behaviour must be tested with realistic scenarios.

33. Security-sensitive behaviour must have dedicated tests.

34. Failure testing is required for critical infrastructure dependencies.

35. Production defects should result in regression tests where reasonably reproducible.

36. Performance testing should be separated from ordinary functional testing.

37. CI should run fast tests before expensive tests.

38. Pull requests must execute appropriate unit, integration, and build validation.

39. Release validation must include critical integration and end-to-end workflows.

40. Long-running or expensive suites may run on scheduled pipelines.

41. Test artifacts must not expose secrets or sensitive data.

42. Test infrastructure must not accidentally interact with production systems.

43. Production smoke tests must be controlled and safe.

44. Coverage percentage is a signal, not the objective.

45. Tests are production code and must be reviewed.

46. Test ownership must be explicit.

47. Obsolete tests must be removed.

48. Architecture changes must include corresponding test coverage.

49. Testing strategy must evolve with the system architecture.

50. The goal of testing is confidence in system behaviour, not merely test count or coverage percentage.
```

---

# 197. Status

```text
Decision: ACCEPTED

Testing Model:
Layered Testing Strategy

Primary Layers:
Domain
Unit
Integration
Contract
End-to-End
Failure / Resilience

Infrastructure Testing:
Real Infrastructure Where Semantics Matter

Database:
Real PostgreSQL / PostGIS

Cache / Real-Time State:
Real Redis

Messaging:
Real Kafka / Redpanda Where Practical

External Providers:
Sandbox / Controlled Test Doubles

API:
Contract + Integration + Security

Events:
Schema + Contract + Integration

Outbox:
Integration Tested

DLQ:
Failure Tested

Idempotency:
Concurrency + Repeat Request Tested

Dispatch:
Stand + Smart + Hybrid Tested

ETA:
Prediction + Provider + Fallback Tested

AI:
Model + Safety + Fallback Tested

Migrations:
Real Database Validation

Performance:
Dedicated Test Layer

Security:
Dedicated Test Layer

CI:
Layered and Risk-Based

Production:
Controlled Smoke Validation

Primary Goal:
Provide Reliable, Repeatable, Production-Relevant Confidence in RideForge Behaviour Across Services, Infrastructure, Events, AI, and Critical Business Workflows
```

---

# 198. Decision Summary

RideForge adopts the following testing flow:

```text
                         CODE CHANGE
                              │
                              ▼
                    ┌──────────────────┐
                    │ Static Validation│
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Domain / Unit    │
                    │      Tests       │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Integration      │
                    │      Tests       │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Contract / Event │
                    │      Tests       │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Critical E2E     │
                    │      Tests       │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Release / Staging│
                    │    Validation    │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │ Production Smoke │
                    └──────────────────┘
```

The objective is not to maximize the number of tests.

The objective is to ensure that every important RideForge boundary has an appropriate level of confidence:

```text
Domain
  ↓
Service
  ↓
Database
  ↓
Cache
  ↓
Messaging
  ↓
External Providers
  ↓
AI
  ↓
Dispatch
  ↓
Complete Ride Lifecycle
```

---

# 199. Status Metadata

| Field | Value |
|---|---|
| ADR | `0025` |
| Title | Testing and Integration Strategy |
| Status | Accepted |
| Category | Testing / Quality / Architecture |
| Primary Strategy | Layered Testing |
| Unit Testing | Required |
| Integration Testing | Required |
| Contract Testing | Required for Important Boundaries |
| E2E Testing | Required for Critical Workflows |
| Failure Testing | Required for Critical Dependencies |
| Database Testing | Real PostgreSQL / PostGIS |
| Redis Testing | Real Redis |
| Messaging Testing | Real Kafka / Redpanda Where Practical |
| External Providers | Sandbox / Controlled Doubles |
| AI Testing | Required |
| Security Testing | Required |
| Performance Testing | Dedicated |
| Migration Testing | Required |
| CI Validation | Layered |
| Production Smoke Tests | Controlled |
| Related ADR | `0024-CONFIGURATION_AND_ENVIRONMENT_STRATEGY.md` |
| Next ADR | `0026-MODEL_AND_AI_GOVERNANCE.md` |
