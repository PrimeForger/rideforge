# ADR-0024: Configuration and Environment Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Configuration / Platform Architecture / Operations  
> **Scope:** Application configuration, environment separation, runtime configuration, secrets integration, feature flags, service configuration, configuration validation, deployment configuration, and configuration lifecycle  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is a distributed ride-hailing platform composed of multiple services and infrastructure components.

The platform contains configuration for:

```text
Application Services
Databases
Redis
Kafka / Redpanda
External Providers
Routing Providers
Payment Providers
Notification Providers
AI Services
Observability
Security
Deployment
Feature Flags
Regional Operations
Dispatch Modes
```

Configuration exists at multiple levels:

```text
Application
Service
Environment
Infrastructure
Deployment
Runtime
Feature Flag
Secret
```

Without a consistent configuration strategy, systems tend to develop:

```text
Hard-Coded Values
Environment Drift
Secret Leakage
Inconsistent Defaults
Unclear Ownership
Unsafe Production Changes
Difficult Local Development
Deployment Errors
```

RideForge therefore requires a predictable configuration model that separates:

```text
Code
Configuration
Secrets
Environment
Runtime State
```

---

# 2. Problem

The platform must support multiple environments while maintaining the same application architecture.

At minimum:

```text
Development
Testing
Staging
Production
```

Different environments require different:

```text
Database Endpoints
Redis Endpoints
Kafka Endpoints
Provider Endpoints
Credentials
Logging Levels
Feature Flags
Resource Limits
Operational Policies
```

However, environment-specific configuration must not cause application behaviour to become fundamentally different without an explicit architectural reason.

---

# 3. Decision

RideForge will use an:

```text
Environment-Aware
Configuration-Driven
Secret-Separated
Validated
Versioned
Runtime-Injectable
```

configuration strategy.

The core principle is:

> **Application code defines behaviour; configuration defines environment-specific values and controlled operational policies; secrets are managed separately from ordinary configuration.**

---

# 4. Configuration Categories

Configuration will be divided conceptually into:

```text
1. Static Application Configuration
2. Environment Configuration
3. Infrastructure Configuration
4. Runtime Configuration
5. Feature Flags
6. Secrets
7. Derived Configuration
```

---

# 5. Static Application Configuration

Static configuration contains values that rarely change and are part of application behaviour.

Examples:

```text
Default Timeout
Retry Limits
Internal Constants
Supported Event Names
Validation Limits
```

These may be represented in source code when they are genuinely code-level constants.

---

# 6. Environment Configuration

Environment configuration identifies where the application is running.

Examples:

```text
ENVIRONMENT=development
ENVIRONMENT=staging
ENVIRONMENT=production
```

---

# 7. Infrastructure Configuration

Infrastructure configuration includes:

```text
Database Host
Redis Host
Kafka Brokers
Object Storage Endpoint
Telemetry Endpoint
```

These values should not be hard-coded into business logic.

---

# 8. Runtime Configuration

Runtime configuration controls operational behaviour.

Examples:

```text
Request Timeout
Database Pool Size
Kafka Batch Size
Cache TTL
Provider Timeout
Log Level
```

---

# 9. Feature Flags

Feature flags control explicitly approved runtime behaviour.

Examples:

```text
SMART_DISPATCH_ENABLED
AI_RANKING_ENABLED
NEW_ETA_ENGINE_ENABLED
```

Feature flags must not become a replacement for proper architecture or configuration management.

---

# 10. Secrets

Secrets include:

```text
Passwords
API Keys
Tokens
Private Keys
Signing Keys
Encryption Keys
Provider Credentials
```

Secrets follow:

```text
ADR-0023 — Security and Secret Management
```

---

# 11. Derived Configuration

Some values may be derived from other configuration.

Example:

```text
DATABASE_URL
```

may be constructed from:

```text
DATABASE_HOST
DATABASE_PORT
DATABASE_NAME
DATABASE_USER
DATABASE_PASSWORD
```

However, the system should avoid unnecessarily duplicating configuration sources.

---

# 12. Configuration Precedence

When multiple configuration sources exist, precedence must be explicit.

A conceptual order is:

```text
Application Defaults
        ↓
Environment Configuration
        ↓
Deployment Configuration
        ↓
Runtime Overrides
```

Secrets should be injected through the approved secret-management mechanism.

---

# 13. Configuration vs Secrets

Configuration and secrets must remain conceptually separate.

```text
Configuration
→ How the system behaves

Secret
→ How the system proves access
```

Example:

```text
REDIS_TIMEOUT=500ms
```

is configuration.

```text
REDIS_PASSWORD=...
```

is a secret.

---

# 14. Source Code

Source code should not contain environment-specific infrastructure endpoints.

Avoid:

```text
if production {
    connectToProductionDatabase()
}
```

Prefer configuration-driven dependency setup.

---

# 15. Environment Variables

Environment variables are an accepted runtime configuration mechanism.

Examples:

```text
APP_ENV
HTTP_PORT
DATABASE_HOST
DATABASE_PORT
REDIS_HOST
KAFKA_BROKERS
LOG_LEVEL
```

---

# 16. Environment Variables and Secrets

Environment variables may receive secrets at runtime, but secret lifecycle management should remain outside source code.

Conceptually:

```text
Secret Manager
      ↓
Runtime Injection
      ↓
Environment / Process
      ↓
Application
```

---

# 17. Local `.env` Files

Local development may use:

```text
.env
.env.local
.env.development
```

where appropriate.

Files containing secrets must not be committed.

---

# 18. Example Local Configuration

```text
APP_ENV=development
HTTP_PORT=8080

DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=rideforge

REDIS_HOST=localhost
REDIS_PORT=6379

KAFKA_BROKERS=localhost:9092

LOG_LEVEL=debug
```

This is an example structure, not a mandatory exact environment-variable list.

---

# 19. Production Configuration

Production configuration should be supplied through controlled deployment mechanisms.

Do not require engineers to manually edit production configuration files on running servers.

---

# 20. Environment Isolation

Each environment must have separate configuration.

At minimum:

```text
Development
Testing
Staging
Production
```

---

# 21. Credential Isolation

Credentials must never be shared unnecessarily between environments.

Example:

```text
Development DB Credential
≠
Staging DB Credential
≠
Production DB Credential
```

---

# 22. Resource Isolation

Where practical, environments should use separate:

```text
Databases
Redis Instances
Kafka / Redpanda Topics or Clusters
Storage
Credentials
Telemetry Destinations
```

---

# 23. Development Environment

Development should optimize for:

```text
Fast Feedback
Easy Debugging
Local Reproducibility
Low Cost
```

---

# 24. Testing Environment

Testing should optimize for:

```text
Determinism
Isolation
Repeatability
Automation
```

---

# 25. Staging Environment

Staging should approximate production configuration sufficiently to validate:

```text
Deployment
Integration
Performance
Security
Operational Behaviour
```

---

# 26. Production Environment

Production configuration must prioritize:

```text
Security
Reliability
Observability
Performance
Controlled Changes
```

---

# 27. Configuration Schema

Each service should define a clear configuration schema.

Conceptually:

```text
Configuration
├── Application
├── Server
├── Database
├── Redis
├── Messaging
├── External Providers
├── Observability
├── Security
└── Feature Flags
```

---

# 28. Configuration Validation

Services must validate required configuration during startup.

Examples:

```text
Required Database Host Missing
Invalid Port
Invalid Duration
Invalid URL
Missing Provider Credential
Unsupported Environment
```

---

# 29. Fail Fast

If a required configuration value is missing or invalid, the service should normally fail startup rather than run in an unsafe partial state.

---

# 30. Configuration Error

Configuration errors should clearly identify:

```text
Configuration Key
Expected Type / Constraint
Actual Problem
```

but must not expose secret values.

---

# 31. Secret Validation

A missing secret may be reported as:

```text
PAYMENT_PROVIDER_API_KEY is not configured
```

but not:

```text
PAYMENT_PROVIDER_API_KEY=actual-secret-value
```

---

# 32. Type Validation

Configuration should be parsed into appropriate types.

Examples:

```text
Port → Integer
Timeout → Duration
Enabled → Boolean
URL → URL
Pool Size → Integer
```

---

# 33. Duration Configuration

Duration values should use an unambiguous representation.

Example:

```text
DATABASE_TIMEOUT=2s
```

rather than:

```text
DATABASE_TIMEOUT=2
```

when the unit would otherwise be unclear.

---

# 34. Boolean Configuration

Boolean values should use a consistent representation.

Prefer:

```text
true
false
```

rather than accepting many ambiguous forms unless the configuration parser explicitly supports them.

---

# 35. URL Validation

External service URLs should be validated during startup.

Examples:

```text
DATABASE_URL
REDIS_URL
PROVIDER_URL
TELEMETRY_ENDPOINT
```

---

# 36. Configuration Defaults

Defaults may be provided for safe non-secret settings.

Examples:

```text
LOG_LEVEL=info
HTTP_PORT=8080
```

Defaults must not silently create unsafe production behaviour.

---

# 37. Dangerous Defaults

Avoid defaults for:

```text
Production Passwords
Signing Keys
Payment Credentials
Encryption Keys
Privileged Tokens
```

These must be explicitly supplied.

---

# 38. Safe Defaults

Defaults should generally favour:

```text
Security
Bounded Resources
Reasonable Timeouts
Observability
```

---

# 39. Configuration Naming

Configuration names should be:

```text
Explicit
Consistent
Predictable
Environment-Friendly
```

Example:

```text
DATABASE_MAX_CONNECTIONS
```

rather than:

```text
DBMAX
```

---

# 40. Naming Convention

Environment variables should use:

```text
UPPER_SNAKE_CASE
```

where environment variables are used.

---

# 41. Service Prefixes

For large systems, service-specific prefixes may be useful.

Example:

```text
RIDE_SERVICE_HTTP_PORT
```

However, avoid excessive naming complexity.

---

# 42. Configuration Ownership

Every important configuration setting should have an owner.

Examples:

```text
Database Pool Size
→ Platform / Service Owner

Dispatch Timeout
→ Dispatch Domain Owner

AI Model Version
→ AI Owner

Payment Provider
→ Payment Owner
```

---

# 43. Configuration Documentation

Each service should document:

```text
Configuration Key
Type
Required / Optional
Default
Description
Security Classification
Example
```

---

# 44. Configuration Reference

A service should maintain a machine-readable or documented configuration reference where practical.

Example:

```text
| Key | Type | Required | Default | Description |
|---|---|---:|---|---|
| HTTP_PORT | int | No | 8080 | HTTP listener port |
| LOG_LEVEL | string | No | info | Logging level |
| DATABASE_HOST | string | Yes | - | PostgreSQL host |
```

---

# 45. Configuration Versioning

Configuration changes that materially affect architecture or behaviour should be tracked.

---

# 46. Git Tracking

Non-sensitive configuration templates may be committed.

Example:

```text
.env.example
config.example.yaml
```

These must contain placeholders rather than real secrets.

---

# 47. Example Configuration Template

```text
DATABASE_HOST=
DATABASE_PORT=5432
DATABASE_NAME=
DATABASE_USER=
DATABASE_PASSWORD=

REDIS_HOST=
REDIS_PORT=6379

KAFKA_BROKERS=

PAYMENT_PROVIDER_API_KEY=
```

The actual secret values must not be committed.

---

# 48. Configuration Drift

Environment drift occurs when:

```text
Development
≠
Staging
≠
Production
```

in unexpected ways.

---

# 49. Drift Prevention

Use:

```text
Shared Configuration Schema
Environment Templates
Infrastructure as Code
Deployment Automation
Configuration Validation
```

where practical.

---

# 50. Configuration as Code

Infrastructure and deployment configuration should be managed as code where appropriate.

Examples:

```text
Terraform
Helm
Kubernetes Manifests
Docker Compose
CI/CD Configuration
```

The exact technology is deployment-specific.

---

# 51. Infrastructure Configuration

Infrastructure configuration should not be hidden in manual server changes.

Manual changes create:

```text
Drift
Unclear Ownership
Poor Reproducibility
Difficult Recovery
```

---

# 52. Configuration Changes

Configuration changes should follow an appropriate review process.

High-risk configuration changes should receive stronger review.

Examples:

```text
Database Endpoint
Payment Provider
Authentication Policy
Security Rules
Dispatch Policy
```

---

# 53. Runtime Configuration Changes

Runtime configuration may be changed without redeployment only when the architecture explicitly supports safe dynamic configuration.

---

# 54. Static vs Dynamic Configuration

Static configuration:

```text
Loaded at Startup
```

Dynamic configuration:

```text
Can Change During Runtime
```

These must not be confused.

---

# 55. Prefer Static Configuration Initially

For most services, configuration should be loaded at startup.

Dynamic configuration should be introduced only where operational value justifies its complexity.

---

# 56. Dynamic Configuration

Potential candidates include:

```text
Feature Flags
Operational Thresholds
Provider Routing
Experiment Allocation
```

---

# 57. Dynamic Configuration Safety

Dynamic changes should have:

```text
Validation
Authorization
Auditability
Rollback
```

where applicable.

---

# 58. Configuration Reload

If runtime reload is supported, services must define:

```text
What Can Reload
What Requires Restart
How Reload Is Triggered
What Happens If Reload Fails
```

---

# 59. Immutable Configuration

Critical infrastructure configuration should generally be changed through deployment rather than arbitrary runtime mutation.

---

# 60. Feature Flags

Feature flags are intended for:

```text
Controlled Rollout
Experimentation
Temporary Operational Control
Emergency Disablement
```

---

# 61. Feature Flag Lifecycle

Every temporary feature flag should have:

```text
Owner
Purpose
Creation Date
Expected Removal Date
Default State
```

---

# 62. Feature Flag Debt

Old feature flags must be removed.

A permanent collection of flags creates:

```text
Complexity
Unclear Behaviour
Testing Explosion
```

---

# 63. AI Feature Flags

AI capabilities may use flags such as:

```text
SMART_DISPATCH_ENABLED
AI_RANKING_ENABLED
AI_ETA_ENABLED
```

but hard business constraints must remain enforced regardless of AI flag state.

---

# 64. Dispatch Configuration

Dispatch configuration may include:

```text
Dispatch Mode
Candidate Limits
Search Radius
Timeout
Fallback Policy
```

---

# 65. Hybrid Dispatch

RideForge supports both:

```text
Smart Dispatch
Stand Dispatch
```

and configuration must allow the appropriate strategy to be selected by operating context.

---

# 66. Dispatch Mode Selection

Dispatch mode may be determined by:

```text
Region
Operating Zone
Business Configuration
Feature Flag
Operational Policy
```

It must not be selected through arbitrary hard-coded conditionals spread across services.

---

# 67. Legal Configuration

Legal and regional rules may require configuration, but legal validation must remain explicit domain logic.

Configuration must not become a mechanism for accidentally disabling mandatory legal constraints.

---

# 68. Regional Configuration

Regional configuration may include:

```text
Region Identifier
Operating Status
Dispatch Mode
Provider Selection
Currency
Timezone
```

---

# 69. Region Configuration

Regional configuration should be versioned and auditable where it affects ride eligibility or operational behaviour.

---

# 70. Timezone Configuration

Business-date calculations should use explicit regional timezones.

Do not assume:

```text
UTC
```

is always the business timezone.

---

# 71. Clock Configuration

Services should rely on a consistent time source and avoid hard-coded local times.

---

# 72. Provider Configuration

External provider configuration should include:

```text
Provider Name
Endpoint
Timeout
Retry Policy
Circuit Policy
Credential Reference
```

where applicable.

---

# 73. Provider Switching

Provider selection should be configurable where the architecture supports multiple providers.

Example:

```text
ROUTE_PROVIDER=provider_a
```

However, provider selection should not bypass business validation.

---

# 74. Provider Fallback

Provider fallback configuration should follow:

```text
ADR-0021 — Failure and Degradation Strategy
```

---

# 75. Database Configuration

Database configuration may include:

```text
Host
Port
Database
User
Password
Pool Size
Timeout
SSL Mode
```

Secrets remain governed by:

```text
ADR-0023
```

---

# 76. Connection Pool Configuration

Pool sizing must consider:

```text
Service Instance Count
PgBouncer
Database Capacity
Expected Concurrency
```

Avoid independently increasing every service's pool size without considering aggregate database connections.

---

# 77. Redis Configuration

Redis configuration may include:

```text
Endpoint
Port
TLS
Connection Pool
Timeout
Database Index
```

---

# 78. Kafka / Redpanda Configuration

Messaging configuration may include:

```text
Brokers
Client ID
Security Mode
Topic Names
Consumer Group
Producer Settings
Consumer Settings
```

---

# 79. Topic Names

Topic names should be configuration-driven rather than scattered as literals throughout business code.

---

# 80. Consumer Group

Consumer group identity is operationally significant and should be explicitly configured.

---

# 81. Event Configuration

Configuration should not change event schema semantics silently.

Schema evolution remains governed by event architecture and messaging standards.

---

# 82. Observability Configuration

Observability configuration may include:

```text
Log Level
Metrics Endpoint
Trace Endpoint
Sampling Rate
Telemetry Exporter
```

Sensitive telemetry credentials must use secret management.

---

# 83. Logging Configuration

Log level should normally be configurable by environment.

Typical defaults:

```text
Development → debug
Staging → info
Production → info
```

Exact levels should depend on operational needs.

---

# 84. Production Debug Logging

Production DEBUG logging should not be enabled casually.

It can create:

```text
Cost
Noise
Sensitive Data Risk
Performance Impact
```

---

# 85. Configuration and Performance

Configuration should support operational tuning for:

```text
Timeouts
Connection Pools
Batch Sizes
Worker Counts
Concurrency
Cache TTLs
```

Changes must still be validated through performance testing.

---

# 86. Configuration and Security

Configuration must not allow operators to accidentally disable critical security controls.

For example:

```text
AUTH_REQUIRED=false
```

should not be an unrestricted production configuration switch.

---

# 87. Security-Sensitive Flags

Security-sensitive configuration should require stronger controls and review.

Examples:

```text
Authentication Policy
Authorization Policy
TLS Verification
Encryption
Public Exposure
```

---

# 88. TLS Verification

Production systems should not disable TLS certificate verification as a general workaround.

---

# 89. Dangerous Configuration

Configurations that can materially weaken security should:

```text
Have Safe Defaults
Require Explicit Override
Be Audited
```

where appropriate.

---

# 90. Configuration Validation by Environment

Some values may be invalid in production even if they are useful locally.

Example:

```text
ALLOW_INSECURE_LOCAL_AUTH=true
```

may be acceptable only in local development.

---

# 91. Environment-Specific Validation

Configuration validators should be aware of environment-specific security requirements.

---

# 92. Production Guardrails

Production startup should reject unsafe configuration combinations.

Examples:

```text
Production + Debug Authentication
Production + Missing TLS
Production + Placeholder Credentials
Production + Invalid Provider
```

---

# 93. Configuration Dependency

Some configuration values depend on others.

Example:

```text
TLS_ENABLED=true
```

may require:

```text
TLS_CERT
TLS_KEY
```

The configuration validator should detect missing dependencies.

---

# 94. Configuration Consistency

Configuration combinations should be validated.

Example:

```text
POOL_MAX < POOL_MIN
```

must fail validation.

---

# 95. Numeric Bounds

Configuration values should have sensible bounds.

Examples:

```text
Timeout > 0
Pool Size > 0
Retry Count >= 0
Batch Size > 0
```

---

# 96. Timeout Strategy

Timeouts should exist for external dependencies.

Examples:

```text
Database
Redis
Kafka
Routing Provider
Payment Provider
Notification Provider
AI Provider
```

---

# 97. Retry Configuration

Retry policies should be explicit.

Configuration may include:

```text
Max Attempts
Backoff
Maximum Delay
```

Retries must follow the relevant failure strategy.

---

# 98. Retry Safety

Configuration must not enable retries for non-idempotent operations without appropriate safeguards.

---

# 99. Circuit Breaker Configuration

If circuit breakers are used, configuration may include:

```text
Failure Threshold
Open Duration
Half-Open Limits
```

---

# 100. Configuration and Idempotency

Critical operations must follow:

```text
ADR-0020 — Idempotency Strategy
```

Configuration must not accidentally remove idempotency protection.

---

# 101. Configuration and Degradation

Fallback and degradation configuration follows:

```text
ADR-0021 — Failure and Degradation Strategy
```

---

# 102. Configuration and Observability

Configuration changes should be observable.

Important configuration changes should expose:

```text
Who
What
When
Environment
Version
```

without exposing secret values.

---

# 103. Configuration Version

Where practical, deployments should expose a configuration version or revision identifier.

This helps correlate:

```text
Incident
→ Configuration Change
```

---

# 104. Configuration Audit

High-impact configuration changes should be auditable.

Examples:

```text
Dispatch Mode
Payment Provider
Authentication Policy
Regional Operating Status
Provider Routing
```

---

# 105. Configuration Rollback

Configuration changes should have a rollback strategy.

Prefer:

```text
Previous Known-Good Configuration
```

over manual reconstruction.

---

# 106. Configuration Deployment

Configuration changes should use the same controlled deployment process as application changes where appropriate.

---

# 107. Atomic Configuration Changes

Related configuration changes should be applied consistently.

Avoid partial updates where one setting changes while its required dependency does not.

---

# 108. Configuration Validation Before Deployment

CI/CD should validate configuration templates before deployment.

Checks may include:

```text
Syntax
Schema
Required Keys
Allowed Values
Type
Environment Rules
Security Constraints
```

---

# 109. Secret Validation in CI

CI should verify that required secret references exist conceptually without printing or exposing secret values.

---

# 110. Configuration Templates

Maintain templates such as:

```text
.env.example
config.example.yaml
```

to make required configuration discoverable.

---

# 111. Configuration Documentation

Documentation should remain synchronized with actual configuration.

When a configuration key is:

```text
Added
Removed
Renamed
Deprecated
```

documentation should be updated.

---

# 112. Configuration Deprecation

Deprecated configuration should have:

```text
Deprecation Notice
Migration Path
Removal Target
Owner
```

---

# 113. Unknown Configuration

Services should decide explicitly whether unknown configuration keys are:

```text
Rejected
Ignored with Warning
```

For critical production services, rejecting unexpected configuration can help detect typos.

---

# 114. Configuration Typos

Example:

```text
DATABASE_HOTS
```

should not silently become:

```text
DATABASE_HOST
```

A configuration parser should detect invalid keys where practical.

---

# 115. Runtime Configuration Object

Application code should preferably consume a validated configuration object rather than repeatedly reading raw environment variables throughout the codebase.

Conceptually:

```text
Environment
    ↓
Parser
    ↓
Validator
    ↓
Configuration Object
    ↓
Application
```

---

# 116. Avoid Direct Environment Access Everywhere

Avoid:

```text
os.Getenv(...)
```

or equivalent environment access scattered across business logic.

Prefer centralized configuration loading.

---

# 117. Dependency Injection

Services should receive configuration-dependent dependencies through constructors or explicit dependency wiring.

---

# 118. Configuration and Domain Logic

Domain logic should not depend directly on environment variables.

Bad conceptual design:

```text
Domain Rule
→ Read ENV
```

Better:

```text
Configuration
→ Application Layer
→ Domain Policy
```

---

# 119. Configuration and Infrastructure

Infrastructure-specific configuration should remain outside domain entities.

---

# 120. Configuration and Tests

Tests should be able to construct configuration explicitly.

This improves:

```text
Determinism
Isolation
Readability
```

---

# 121. Test Configuration

Integration tests should use dedicated test infrastructure configuration.

Examples:

```text
Test PostgreSQL
Test Redis
Test Kafka / Redpanda
Mock Providers
```

---

# 122. Test Secrets

Use:

```text
Dedicated Test Credentials
```

and never real production secrets.

---

# 123. Configuration in Unit Tests

Unit tests should avoid depending on the developer's machine environment.

---

# 124. Configuration in Integration Tests

Integration tests may use environment variables or test configuration files to connect to controlled infrastructure.

---

# 125. Local Development

Local development should be reproducible through:

```text
Docker Compose
Environment Templates
Seed Configuration
Local Documentation
```

where appropriate.

---

# 126. Local Configuration Safety

Local defaults should not accidentally point to production.

---

# 127. Production Endpoint Protection

A developer environment should not silently connect to production because of a missing local variable.

---

# 128. Explicit Environment Selection

The environment should be explicit.

Example:

```text
APP_ENV=development
```

---

# 129. Environment Mismatch Protection

Services should validate that:

```text
APP_ENV
```

matches deployment context where infrastructure supports such validation.

---

# 130. Configuration and Containers

Container images should remain environment-neutral where practical.

Do not bake production credentials into images.

---

# 131. Container Runtime Configuration

Runtime environment should supply:

```text
Endpoints
Credentials
Operational Settings
Feature Flags
```

---

# 132. Configuration and Kubernetes

Where Kubernetes is used:

```text
ConfigMap
→ Non-Secret Configuration

Secret / External Secret Mechanism
→ Sensitive Configuration
```

subject to the chosen secret-management architecture.

---

# 133. Configuration and Cloud

Cloud-specific configuration should be managed through:

```text
Infrastructure as Code
Environment Configuration
Secret Management
Deployment Automation
```

---

# 134. Configuration and CI/CD

CI/CD should inject deployment-specific configuration without exposing secrets in build logs.

---

# 135. Configuration and Build Artifacts

Build artifacts should not contain environment-specific secrets.

---

# 136. Build-Time vs Runtime Configuration

Prefer runtime configuration for environment-specific values.

This allows the same artifact to be promoted across environments where practical.

---

# 137. Immutable Artifact Principle

Conceptually:

```text
Build Once
    ↓
Test
    ↓
Promote
    ↓
Configure at Runtime
```

rather than:

```text
Build Development
Build Staging
Build Production
```

with different embedded secrets.

---

# 138. Deployment Promotion

A tested artifact should be promotable without rebuilding solely to change environment configuration.

---

# 139. Configuration Drift Detection

Where practical, automation should identify unexpected differences between intended and actual configuration.

---

# 140. Configuration Inventory

Maintain awareness of:

```text
Configuration Keys
Owners
Environment Scope
Sensitivity
Default
Source
```

---

# 141. Configuration Source

Each key should have a known source such as:

```text
Code
Environment
Config File
Secret Manager
Feature Flag System
Infrastructure
```

---

# 142. Configuration Sensitivity

Each configuration key should be classified:

```text
Public
Internal
Sensitive
Secret
```

---

# 143. Example Inventory

```text
| Key | Source | Sensitivity | Environment | Owner |
|---|---|---|---|---|
| HTTP_PORT | Environment | Internal | All | Platform |
| LOG_LEVEL | Environment | Internal | All | Platform |
| DATABASE_HOST | Environment | Internal | All | Platform |
| DATABASE_PASSWORD | Secret Manager | Secret | All | Platform |
| DISPATCH_MODE | Config | Sensitive | Regional | Dispatch |
```

---

# 144. Configuration Review

Review configuration when:

```text
Service Added
Provider Added
Security Boundary Changed
Database Changed
Messaging Changed
AI Model Changed
Region Added
Deployment Architecture Changed
```

---

# 145. Regional Configuration Governance

Because RideForge may operate across different regions, regional configuration should be controlled carefully.

Regional settings must not accidentally override:

```text
Global Security Controls
Legal Constraints
Data Protection Controls
```

---

# 146. Regional Dispatch Configuration

Regional dispatch mode may select:

```text
SMART
STAND
```

or an approved fallback strategy.

The configuration must not bypass eligibility rules.

---

# 147. Regional Provider Configuration

Different regions may use different:

```text
Routing Provider
Payment Provider
Notification Provider
```

where legally and operationally required.

---

# 148. Regional Data Configuration

Data residency requirements may influence configuration such as:

```text
Database Region
Storage Region
Telemetry Destination
```

These decisions must remain aligned with security and privacy requirements.

---

# 149. AI Configuration

AI-related configuration may include:

```text
Model Name
Model Version
Inference Endpoint
Timeout
Fallback
Feature Flag
```

Sensitive provider credentials remain secrets.

---

# 150. AI Configuration Governance

AI configuration changes should follow:

```text
ADR-0026 — Model and AI Governance
```

---

# 151. Model Rollback

Model configuration should support rollback to a previously approved model version.

---

# 152. Configuration and Experimentation

Experiment configuration may include:

```text
Experiment ID
Variant
Traffic Allocation
Eligibility
```

Experimentation follows the AI experimentation and product experimentation policies.

---

# 153. Configuration Safety

No configuration should silently disable:

```text
Authentication
Authorization
Legal Validation
Data Integrity
Idempotency
Critical Observability
```

---

# 154. Safe Configuration Hierarchy

Conceptually:

```text
Security Constraints
        ↓
Legal Constraints
        ↓
Data Integrity Constraints
        ↓
Operational Configuration
        ↓
Optimization Configuration
```

Lower-level configuration must not override higher-level constraints.

---

# 155. Configuration and Business Rules

Business rules that require strong correctness should remain explicit domain logic.

Configuration may supply:

```text
Policy Values
Thresholds
Regional Parameters
```

but should not turn critical domain invariants into arbitrary strings.

---

# 156. Configuration and Schema

Configuration changes should not silently change persistent data schemas.

Database schema changes follow:

```text
ADR-0019
```

and migration documentation.

---

# 157. Configuration and Event Contracts

Changing configuration must not silently break event contracts.

Event schema changes require the appropriate messaging and migration process.

---

# 158. Configuration and API Contracts

API configuration must not alter public API semantics unexpectedly.

---

# 159. Configuration Change Classification

Configuration changes can be classified:

```text
LOW
MEDIUM
HIGH
CRITICAL
```

---

# 160. Low-Risk Configuration

Examples:

```text
Log Level
Non-Critical Cache TTL
Development Port
```

---

# 161. Medium-Risk Configuration

Examples:

```text
Worker Count
Timeouts
Retry Limits
```

---

# 162. High-Risk Configuration

Examples:

```text
Provider Selection
Database Endpoint
Dispatch Strategy
Regional Configuration
```

---

# 163. Critical Configuration

Examples:

```text
Authentication Policy
Authorization Policy
Encryption
Security Boundaries
Production Access
```

---

# 164. Change Approval

Higher-risk configuration changes should require stronger approval.

---

# 165. Configuration Rollback Plan

Every high-risk configuration change should identify:

```text
Previous Value
Rollback Procedure
Owner
Validation Signal
```

---

# 166. Configuration Change Observability

After a high-risk change, monitor:

```text
Errors
Latency
Business Success
Fallback
Dependency Health
Security Signals
```

---

# 167. Configuration and Deployment Strategy

Deployment architecture is governed by:

```text
ADR-0027 — Cloud and Deployment Strategy
```

This ADR defines how configuration participates in deployment.

---

# 168. Configuration and Cost

Configuration may influence:

```text
Resource Counts
Database Pooling
Telemetry Volume
AI Usage
Provider Usage
```

Cost-sensitive changes should follow:

```text
ADR-0028 — Cost Optimization Strategy
```

---

# 169. Configuration and Failure Strategy

Configuration must provide safe operational controls for:

```text
Timeout
Retry
Fallback
Provider Selection
Feature Disablement
```

without allowing unsafe bypasses.

---

# 170. Configuration and Degradation

When a dependency is degraded, configuration may control approved fallback behaviour.

Example:

```text
Primary ETA Provider
        ↓
Timeout
        ↓
Fallback Provider
```

---

# 171. Configuration Recovery

If configuration becomes invalid, recovery should use:

```text
Last Known Good Configuration
```

where possible.

---

# 172. Configuration Backup

Important non-secret configuration should be version-controlled or otherwise recoverable.

Secrets should remain in dedicated secret-management systems.

---

# 173. Configuration Disaster Recovery

Disaster recovery should include the ability to reconstruct:

```text
Application Configuration
Infrastructure Configuration
Secret References
Feature Flag State
Regional Configuration
```

without exposing secret values.

---

# 174. Configuration Testing Matrix

Test configuration across:

```text
Development
Testing
Staging
Production
```

and important feature combinations.

---

# 175. Invalid Configuration Tests

Test:

```text
Missing Required Key
Invalid Type
Invalid Value
Invalid Dependency
Unsafe Production Setting
Unknown Key
```

---

# 176. Configuration Contract Tests

Services consuming shared configuration should validate that the expected configuration contract remains compatible.

---

# 177. Configuration Security Tests

Verify:

```text
Secrets Not Logged
Secrets Not Committed
Production Credentials Not Used in Tests
Unsafe Flags Rejected
Unauthorized Configuration Changes Blocked
```

---

# 178. Configuration Performance Tests

Validate configuration changes that affect:

```text
Concurrency
Database Connections
Workers
Timeouts
Batching
Caching
```

under realistic load.

---

# 179. Configuration Documentation Lifecycle

Configuration documentation should be updated as part of the same change that modifies configuration.

---

# 180. Configuration Change Checklist

Before changing a production configuration value:

```text
□ Identify Owner
□ Identify Risk
□ Validate Value
□ Check Dependencies
□ Review Security Impact
□ Review Business Impact
□ Prepare Rollback
□ Apply Change
□ Verify Telemetry
□ Verify Business Health
□ Record Change
```

---

# 181. Configuration Review Triggers

Revisit this ADR when:

```text
A New Environment Is Added
A New Secret Manager Is Added
Dynamic Configuration Is Introduced
Feature Flag Infrastructure Is Added
A New Cloud Provider Is Added
Multi-Region Deployment Is Added
Configuration Drift Becomes Significant
A Configuration Incident Occurs
A Major Provider Is Replaced
Security Requirements Change
```

---

# 182. Consequences

## 182.1 Positive Consequences

This strategy provides:

```text
Predictable Environments
Reduced Configuration Drift
Safer Deployments
Better Local Development
Clear Secret Separation
Improved Rollback
Better Operational Control
Improved Auditability
```

---

## 182.2 Negative Consequences

The architecture introduces:

```text
Configuration Management Complexity
Schema Validation Work
Environment Management
Feature Flag Lifecycle
Deployment Automation Requirements
Configuration Documentation
Operational Governance
```

These costs are accepted.

---

# 183. Risks

## Risk 1 — Configuration Drift

### Mitigation

```text
Configuration as Code
Validation
Automation
Environment Templates
```

---

## Risk 2 — Secret Leakage

### Mitigation

```text
ADR-0023
Secret Manager
Secret Scanning
Log Redaction
```

---

## Risk 3 — Unsafe Production Configuration

### Mitigation

```text
Environment-Aware Validation
Safe Defaults
Change Review
Production Guardrails
```

---

## Risk 4 — Feature Flag Accumulation

### Mitigation

```text
Owner
Expiration / Removal Target
Regular Cleanup
```

---

## Risk 5 — Configuration Becomes Business Logic

### Mitigation

Keep:

```text
Critical Invariants
Legal Rules
Security Rules
Data Integrity Rules
```

in explicit application/domain logic.

---

## Risk 6 — Dynamic Configuration Creates Hidden State

### Mitigation

Use dynamic configuration only where necessary and make changes:

```text
Auditable
Versioned
Observable
Reversible
```

---

# 184. Alternatives Considered

## 184.1 Hard-Code Environment Values

### Advantages

```text
Simple
```

### Disadvantages

```text
Unsafe
Inflexible
Difficult Deployment
Configuration Drift
```

### Decision

```text
Rejected.
```

---

# 185. Environment-Specific Builds

### Advantages

```text
Simple Runtime
```

### Disadvantages

```text
Artifact Duplication
Secret Leakage Risk
Promotion Complexity
```

### Decision

```text
Rejected as the default strategy.
```

---

# 186. One Global Configuration File

### Advantages

```text
Centralized
```

### Disadvantages

```text
Poor Environment Isolation
Secret Risk
Large Blast Radius
```

### Decision

```text
Rejected.
```

---

# 187. Fully Dynamic Configuration

### Advantages

```text
Runtime Changes
No Restart
```

### Disadvantages

```text
Hidden Runtime State
Complexity
Harder Reproduction
Operational Risk
```

### Decision

```text
Rejected as the default strategy.
```

Dynamic configuration is allowed selectively.

---

# 188. No Configuration Validation

### Advantages

```text
Less Code
```

### Disadvantages

```text
Late Failures
Unsafe Defaults
Harder Debugging
```

### Decision

```text
Rejected.
```

---

# 189. Validation

This ADR should be validated through:

```text
Configuration Unit Tests
Startup Validation Tests
Environment Tests
CI Validation
Secret Scanning
Deployment Tests
Configuration Drift Checks
Rollback Tests
Feature Flag Tests
Security Tests
Integration Tests
```

---

# 190. Final Principles

The following principles are mandatory:

```text
1. Application code defines behaviour; configuration defines environment-specific values and controlled operational policies.

2. Secrets are separate from ordinary configuration.

3. Production secrets follow ADR-0023.

4. Environment-specific infrastructure endpoints must not be hard-coded into business logic.

5. Configuration must be validated at startup.

6. Invalid required configuration should normally fail fast.

7. Configuration errors must never expose secret values.

8. Environment variables are an accepted runtime configuration mechanism.

9. Local .env files must not be committed when they contain secrets.

10. Development, testing, staging, and production must have isolated configuration.

11. Production credentials must not be reused across lower environments.

12. Configuration should be centrally parsed and validated.

13. Business logic should not read raw environment variables directly.

14. Services should consume validated configuration objects.

15. Configuration keys should have clear names and ownership.

16. Configuration templates should be maintained for discoverability.

17. Non-sensitive configuration templates may be version-controlled.

18. Secrets must never be stored in source-controlled configuration.

19. Production configuration must have security guardrails.

20. Unsafe security configuration must not be enabled accidentally.

21. Critical legal, security, and data-integrity constraints must not be bypassable through ordinary configuration.

22. Feature flags must have owners and lifecycle management.

23. Temporary feature flags should be removed after their purpose is complete.

24. Dynamic configuration should be introduced only where operational value justifies its complexity.

25. High-risk configuration changes require stronger review.

26. Important configuration changes must be observable and auditable where appropriate.

27. Configuration changes require rollback strategies.

28. Environment drift should be detected and controlled.

29. Infrastructure configuration should be managed as code where practical.

30. Production artifacts should remain environment-neutral where practical.

31. Runtime configuration should be preferred over embedding environment-specific secrets into artifacts.

32. Database pool configuration must consider aggregate infrastructure capacity.

33. Timeout and retry configuration must be explicit.

34. Retry configuration must not make non-idempotent operations unsafe.

35. Provider configuration must support approved fallback strategies.

36. Regional configuration must not override mandatory legal or security constraints.

37. Dispatch configuration must support Smart Dispatch and Stand Dispatch without scattering mode logic across services.

38. AI configuration must remain compatible with AI governance requirements.

39. Configuration changes affecting performance must be validated under realistic load.

40. Configuration changes affecting security must receive appropriate security review.

41. Configuration must be recoverable after infrastructure failure.

42. The configuration strategy must evolve with the architecture.
```

---

# 191. Status

```text
Decision: ACCEPTED

Configuration Model:
Environment-Aware + Validated + Runtime-Injectable

Configuration Categories:
Application
Environment
Infrastructure
Runtime
Feature Flags
Secrets
Derived Configuration

Environment Separation:
Development
Testing
Staging
Production

Secrets:
Separate from Ordinary Configuration

Secret Management:
ADR-0023

Configuration Validation:
Required

Startup Validation:
Required for Critical Configuration

Configuration Naming:
Consistent UPPER_SNAKE_CASE for Environment Variables

Local Configuration:
Supported Through Local Environment Files / Equivalent Mechanisms

Production Configuration:
Controlled Deployment Mechanism

Infrastructure Configuration:
Configuration as Code Where Practical

Dynamic Configuration:
Selective Only

Feature Flags:
Controlled + Owned + Lifecycle Managed

Configuration Drift:
Actively Controlled

Production Guardrails:
Required

Configuration Rollback:
Required for High-Risk Changes

Observability:
Configuration Changes Must Be Observable Where Appropriate

Security:
Configuration Must Not Disable Mandatory Security Controls

Legal:
Configuration Must Not Bypass Mandatory Legal Validation

Regional:
Supported

Dispatch:
Smart Dispatch + Stand Dispatch Supported

AI:
Governed by AI Architecture and Model Governance

Primary Goal:
Provide Predictable, Secure, Validated, Reproducible, and Environment-Aware Configuration Across the RideForge Platform
```

---

# 192. Decision Summary

RideForge adopts:

```text
                     CONFIGURATION
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
          Application   Environment   Runtime
              │            │            │
              └────────────┼────────────┘
                           ▼
                     VALIDATION
                           │
                 ┌─────────┴─────────┐
                 ▼                   ▼
          Normal Config           Secrets
                 │                   │
                 │             Secret Manager
                 │                   │
                 └─────────┬─────────┘
                           ▼
                      APPLICATION
                           │
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
     Database            Redis          Kafka / Redpanda
        │                  │                  │
        └──────────────────┼──────────────────┘
                           ▼
                  External Providers
```

The system will keep:

```text
Code
Configuration
Secrets
Runtime State
```

as distinct concerns while providing controlled integration between them.

---

# 193. Status Metadata

| Field | Value |
|---|---|
| ADR | `0024` |
| Title | Configuration and Environment Strategy |
| Status | Accepted |
| Category | Configuration / Platform |
| Primary Concern | Environment and Runtime Configuration |
| Secret Strategy | Dedicated Secret Management |
| Validation | Required |
| Environment Isolation | Required |
| Dynamic Configuration | Selective |
| Feature Flags | Controlled |
| Configuration Drift | Controlled |
| Rollback | Required for High-Risk Changes |
| Related Security ADR | `0023` |
| Related Observability ADR | `0022` |
| Next ADR | `0025-TESTING_AND_INTEGRATION_STRATEGY.md` |

---

# 24. Dispatch Strategy Configuration Clarification

The dispatch configuration model defined by this ADR is hierarchical and must be treated separately from infrastructure/environment configuration.

## 24.1 Two Primary Dispatch Strategies

RideForge has two primary dispatch strategies:

```text
Smart Stand Dispatch
Smart Dispatch
```

AI-assisted dispatch is not a third primary dispatch strategy. It is an optimization capability that may operate within either strategy.

```text
Configuration Resolution
        ↓
Effective Dispatch Strategy
        ↓
Strategy Execution
        ↓
AI-Assisted Optimization where enabled
```

---

## 24.2 Hierarchical Dispatch Strategy Configuration

Dispatch strategy may be configured at different levels of the business/location hierarchy.

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

It is not necessary to configure every level.

The system resolves the effective strategy by starting at the most specific applicable configuration and moving upward until an explicit dispatch strategy is found.

```text
Most Specific Applicable Level
        ↓
Explicit Dispatch Strategy?
   ├── YES → Use it
   └── NO
        ↓
Parent Level
        ↓
Continue Upward
        ↓
System Default if Nothing Is Configured
```

The canonical rule is:

> **Specific configuration overrides inherited configuration.**

Example:

```text
District      → Smart Dispatch
Town A        → Smart Stand Dispatch
Town B        → No explicit configuration
```

Result:

```text
Town A ride → Smart Stand Dispatch
Town B ride → Smart Dispatch
```

The hierarchy must not be hard-coded to only:

```text
State → District → Town → Stand
```

The implementation must support configured intermediate levels and a parent relationship that can evolve with the domain model.

---

## 24.3 Configuration Resolution and Strategy Execution Are Separate

The configuration layer determines **which strategy applies**.

The dispatch engine determines **how that strategy executes**.

Therefore:

```text
Configuration Hierarchy
        ↓
Effective Dispatch Strategy
        ↓
Dispatch Strategy Execution
```

Configuration code must not contain the complete Smart Dispatch or Smart Stand Dispatch algorithm.

Likewise, the dispatch engine must not silently invent configuration inheritance rules.

---

## 24.4 Smart Stand Dispatch Configuration

A configuration resolving to Smart Stand Dispatch means:

> Use stand-preferred dispatch behavior when the rider is within the applicable stand radius, while retaining broader candidate fallback according to the dispatch rules.

It does not mean:

```text
Only drivers currently at the stand may receive the ride.
```

When the rider is inside a configured stand radius:

```text
Preferred Stand
    ↓
Eligible Stand Drivers
    ↓
Stand Queue / Ordering
    ↓
Suitable driver available?
   ├── YES → preferred dispatch
   └── NO  → broader eligible candidate discovery
```

Broader candidates may include:

```text
Drivers outside the preferred stand
Drivers at nearby stands
Drivers from nearby locations
```

If the rider is outside all configured stand radii, Smart Stand Dispatch must not create a stand-only candidate pool.

---

## 24.5 Smart Dispatch Configuration

A configuration resolving to Smart Dispatch means:

```text
Search eligible nearby drivers
        ↓
Evaluate dispatch factors
        ↓
Select best eligible candidate
```

Stand membership is not a dispatch preference under Smart Dispatch.

Eligible drivers may be:

```text
At an auto stand
Outside an auto stand
Associated with another nearby location
```

subject to the applicable hard constraints.

---

## 24.6 Cross-Location Dispatch

The location/configuration hierarchy does not create an automatic hard boundary for candidate discovery.

For example:

```text
Location A → Smart Dispatch
Location B → Smart Stand Dispatch
```

A ride originating in Location A may consider eligible drivers from Location B if local supply is insufficient and geographic expansion permits it.

Location B's dispatch strategy provides context for how its local drivers are prioritized; it does not automatically make those drivers ineligible for the ride.

The candidate context should preserve:

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

## 24.7 Configuration Does Not Override Hard Constraints

The resolved dispatch strategy does not override:

```text
Legal / Regional Restrictions
Driver Eligibility
Driver Availability
Vehicle / Service Compatibility
Safety Constraints
Ride Constraints
Other Hard Business Rules
```

For example:

```text
Smart Stand Dispatch
    ↓
Preferred stand driver
    ↓
Legal validation
    ↓
Rejected
```

The system may continue to broader eligible candidates if the dispatch rules permit it.

Configuration selects the strategy; it does not authorize otherwise invalid candidates.

---

## 24.8 AI-Assisted Dispatch Configuration

AI assistance should be configured independently from the primary dispatch strategy.

Conceptually:

```text
strategy = SMART_STAND
ai_assisted = true
```

or:

```text
strategy = SMART
ai_assisted = true
```

AI assistance must not redefine the configured primary strategy.

If AI is unavailable:

```text
Smart Stand Dispatch + AI
        ↓
AI unavailable
        ↓
Deterministic Smart Stand Dispatch
```

or:

```text
Smart Dispatch + AI
        ↓
AI unavailable
        ↓
Deterministic Smart Dispatch
```

AI failure must not silently change the effective dispatch strategy.

---

## 24.9 Configuration Change Semantics

Changing a dispatch strategy at a configuration level is a business-behavior change and must be treated as an auditable configuration change.

The effective strategy for a ride should be resolved deterministically from the configuration applicable to that ride.

Runtime processing should preserve sufficient context to identify:

```text
Effective Dispatch Strategy
Configuration Level
Configuration Source
Resolution Path where required
```

This supports debugging and operational auditing.

---

## 24.10 Configuration Precedence Guardrails

Implementations must not:

```text
Let a parent configuration override an explicit child configuration.

Require every hierarchy level to contain a configuration.

Hard-code the hierarchy to a fixed set of geographic levels.

Use environment variables to represent business dispatch strategy inheritance.

Resolve dispatch strategy independently in multiple services with different precedence rules.

Treat Smart Stand Dispatch as a stand-only strategy.

Treat location configuration as a hard candidate-discovery boundary.

Allow AI configuration to replace the primary dispatch strategy.

Silently switch Smart Stand Dispatch to Smart Dispatch during fallback.

Use configuration inheritance to bypass legal, safety, or eligibility constraints.
```

There must be one authoritative strategy-resolution rule shared by all consumers of dispatch configuration.

