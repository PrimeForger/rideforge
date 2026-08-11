# ADR-0017: ETA and Route Provider Strategy

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** ETA / Routing / External Provider Architecture  
> **Scope:** Route calculation, travel-time estimation, ETA generation, and provider integration for RideForge  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge requires accurate travel-time and route information for several core capabilities:

```text
Ride Pricing
Driver Dispatch
Smart Matching
ETA Display
Pickup Prediction
Trip Progress
Route Guidance
Demand / Supply Analysis
AI Features
```

A ride-hailing platform cannot reliably use straight-line geographic distance as the primary measure of travel time.

Actual travel time depends on:

```text
Road Network
Traffic
Road Restrictions
Turn Restrictions
Road Quality
Bridges
One-Way Roads
Vehicle Restrictions
Border / Regional Rules
Time of Day
Road Closures
```

RideForge therefore needs a routing and ETA architecture that can support:

```text
Real-Time ETA
Route Calculation
Provider Failover
Provider Replacement
Cost Control
Latency Control
Regional Differences
Future AI-Based ETA
```

The platform should not make the core domain directly dependent on one external map provider.

---

# 2. Problem

RideForge must determine:

```text
How routes are calculated
How ETA is calculated
Which provider is used
How multiple providers are supported
How provider failures are handled
How provider costs are controlled
How provider-specific APIs are isolated
How ETA is consumed by dispatch
How AI ETA can be introduced later
How route data is cached
How stale ETA is handled
```

The architecture must balance:

```text
Accuracy
Latency
Availability
Cost
Coverage
Operational Simplicity
Provider Independence
```

---

# 3. Decision

RideForge will use an **abstraction-based route and ETA provider architecture**.

The core domain will depend on internal interfaces rather than directly on a specific external provider.

Conceptually:

```text
RideForge Domain
      │
      ▼
ETA / Routing Interface
      │
      ▼
Provider Selection Layer
      │
      ├── Route Provider A
      ├── Route Provider B
      ├── Route Provider C
      └── Future AI ETA
```

The primary rule is:

> **External route providers are infrastructure dependencies and must remain behind a RideForge-owned routing / ETA abstraction.**

---

# 4. Core Principle

The application should ask:

```text
"Give me the ETA / route I need."
```

rather than:

```text
"Call Provider X's API."
```

This keeps provider-specific concerns outside the domain.

---

# 5. Routing and ETA Are Related but Different

The architecture distinguishes:

```text
Routing
```

from:

```text
ETA Prediction
```

Routing answers:

```text
Which route should be taken?
```

ETA answers:

```text
How long will the journey take?
```

A route provider may supply both, but RideForge should not assume that they are the same capability.

---

# 6. Routing

Routing may provide:

```text
Route Geometry
Distance
Duration
Instructions
Waypoints
Road Information
```

The exact provider response should be translated into a RideForge-owned representation.

---

# 7. ETA

ETA may be calculated from:

```text
Route Duration
Traffic
Current Driver Location
Pickup Location
Historical Travel Time
Real-Time Signals
```

The ETA subsystem should be able to evolve independently from the route provider.

---

# 8. Provider Abstraction

The architecture should expose internal capabilities such as:

```text
RouteProvider
ETAProvider
```

The exact Go interfaces belong to implementation documentation.

The architectural boundary is:

```text
Application
    ↓
RideForge Interface
    ↓
Provider Adapter
    ↓
External Provider
```

---

# 9. Provider Adapter

Each external provider should have an adapter.

Conceptually:

```text
ETA / Routing Interface
        │
        ├── Provider A Adapter
        ├── Provider B Adapter
        └── Provider C Adapter
```

The adapter is responsible for translating:

```text
RideForge Request
      ↓
Provider Request

Provider Response
      ↓
RideForge Response
```

---

# 10. No Provider-Specific Domain Logic

The domain must not contain logic such as:

```text
If Google:
    use field X

If Mapbox:
    use field Y
```

Provider-specific behaviour belongs inside adapters.

---

# 11. Provider Independence

The architecture should allow the platform to change providers without changing:

```text
Ride Domain
Dispatch Domain
Matching Domain
Pricing Domain
AI Domain
```

Only the provider integration layer should normally change.

---

# 12. Primary Provider

RideForge should configure one approved primary provider for each supported routing capability.

Conceptually:

```text
Primary Route Provider
Primary ETA Provider
```

The exact provider may vary by:

```text
Market
Region
Cost
Coverage
Traffic Availability
Contract
```

---

# 13. Secondary Provider

Where justified, a secondary provider may be configured for:

```text
Fallback
Regional Coverage
Cost Optimization
Validation
Specific Routing Requirements
```

Multiple providers should not be introduced without a measurable operational reason.

---

# 14. Provider Selection

Provider selection may consider:

```text
Region
Country
Service Capability
Availability
Latency
Cost
Coverage
Configured Policy
```

Conceptually:

```text
ETA Request
    ↓
Provider Policy
    ↓
Selected Provider
```

---

# 15. Provider Selection Is Policy

Provider selection should be configuration/policy driven rather than hard-coded throughout application services.

Example:

```text
Region A
→ Provider A

Region B
→ Provider B
```

---

# 16. Provider Capability Matrix

Providers may differ in:

```text
Traffic Support
Route Quality
Geographic Coverage
Vehicle Profiles
Rate Limits
Cost
Response Latency
```

The provider abstraction must not assume that every provider supports every feature.

---

# 17. Capability Detection

A provider configuration should identify supported capabilities.

Conceptually:

```text
Provider A
├── Routing: Yes
├── Traffic ETA: Yes
└── Truck Profile: No

Provider B
├── Routing: Yes
├── Traffic ETA: Yes
└── Truck Profile: Yes
```

---

# 18. Unsupported Capability

If a provider cannot satisfy a required capability:

```text
Do Not Silently Approximate
```

unless the fallback behaviour is explicitly defined.

Possible actions:

```text
Use Secondary Provider
Use Cached Result
Use Deterministic Estimate
Return Controlled Error
```

---

# 19. Route Request

A conceptual route request may contain:

```text
origin
destination
waypoints
vehicle_type
route_preferences
region
requested_at
```

The exact API contract is implementation-specific.

---

# 20. ETA Request

A conceptual ETA request may contain:

```text
origin
destination
current_time
vehicle_type
traffic_required
region
```

Additional fields may be included as the ETA system evolves.

---

# 21. Route Response

A RideForge-owned route representation may include:

```text
distance
duration
geometry
provider
provider_version
calculated_at
```

Provider-specific fields should remain inside the adapter unless they are explicitly part of the platform contract.

---

# 22. ETA Response

A RideForge-owned ETA representation may include:

```text
eta_seconds
distance_meters
calculated_at
expires_at
source
confidence
```

Not every provider needs to supply every field.

---

# 23. ETA Freshness

ETA is time-sensitive.

A response should therefore have a freshness policy.

Conceptually:

```text
Fresh
Stale
Expired
Unavailable
```

The exact thresholds depend on the use case.

---

# 24. ETA Must Not Be Treated as Permanent Data

An ETA calculated at:

```text
10:00
```

should not automatically be considered valid at:

```text
10:30
```

unless the use case explicitly allows stale data.

---

# 25. ETA Consumers

ETA may be consumed by:

```text
Rider App
Driver App
Dispatch
Matching
Pricing
Operations
Analytics
AI Systems
```

Each consumer may require different freshness.

---

# 26. ETA for Rider Display

Rider-facing ETA should prioritize:

```text
Freshness
Predictability
Low Latency
```

A slightly older cached ETA may be preferable to a long provider timeout, depending on the UX policy.

---

# 27. ETA for Dispatch

Dispatch ETA is more sensitive to:

```text
Current Driver Location
Pickup Route
Traffic
```

Therefore dispatch should use an appropriately fresh ETA.

---

# 28. ETA for Historical Analytics

Historical analytics may use:

```text
Stored ETA
Actual Travel Time
Provider Data
```

with much less stringent real-time requirements.

---

# 29. ETA for AI

AI systems may use:

```text
Provider ETA
Historical Travel Time
Current Traffic
Predicted ETA
```

but the feature source and timestamp must remain identifiable.

---

# 30. Provider Latency

Provider calls must have strict timeouts.

Avoid:

```text
Dispatch
   ↓
Wait indefinitely for Route Provider
```

---

# 31. Provider Timeout

If the primary provider exceeds the configured timeout:

```text
Primary Provider
      ↓
Timeout
      ↓
Fallback Policy
```

may invoke a secondary provider or deterministic fallback.

---

# 32. Provider Failure

Provider failures include:

```text
Timeout
Network Failure
HTTP 5xx
Rate Limit
Invalid Response
Authentication Failure
Coverage Failure
Service Outage
```

Each should be classified appropriately.

---

# 33. Provider Retry

Provider requests may be retried only when:

```text
The failure is transient
+
The operation is safe to retry
+
The retry fits within the latency budget
```

---

# 34. Retry Policy

Provider retries should use:

```text
Bounded Attempts
Exponential Backoff
Jitter
Deadline Awareness
```

Avoid retry storms.

---

# 35. Provider Rate Limits

External providers may enforce rate limits.

RideForge must respect:

```text
Requests per Second
Requests per Minute
Daily Quotas
Concurrent Request Limits
```

according to provider contracts.

---

# 36. Rate Limit Handling

When a provider returns a rate-limit response:

```text
Do Not Immediately Retry Repeatedly
```

Use:

```text
Retry-After
Backoff
Secondary Provider
Cache
Fallback
```

where supported.

---

# 37. Circuit Breaker

A provider that repeatedly fails may be isolated using a circuit breaker.

Conceptually:

```text
Healthy
   ↓
Repeated Failures
   ↓
Open
   ↓
Fail Fast / Fallback
   ↓
Recovery Probe
   ↓
Healthy
```

---

# 38. Provider Health

Track provider health using:

```text
Latency
Error Rate
Timeout Rate
Rate-Limit Rate
Availability
```

Provider selection can use this information where policy allows.

---

# 39. Provider Failover

A representative flow is:

```text
ETA Request
    ↓
Primary Provider
    │
    ├── Success → Return Result
    │
    └── Failure
          ↓
     Secondary Provider
          │
          ├── Success → Return Result
          │
          └── Failure
                ↓
          Safe Fallback
```

---

# 40. Failover Must Be Bounded

Do not chain many providers:

```text
A → B → C → D → E
```

because total latency can become unacceptable.

A preferred initial model is:

```text
Primary
+
One Secondary
+
Safe Fallback
```

unless operational requirements justify more.

---

# 41. Provider Failover and Cost

Fallback should consider cost.

A cheap primary provider may be used normally while a more expensive provider is reserved for:

```text
Outage
High-Value Requests
Critical Operations
Coverage Gaps
```

only when policy permits.

---

# 42. Provider Failover and Accuracy

The secondary provider may produce a different route or ETA.

The response should retain its source:

```text
source = primary
```

or:

```text
source = fallback
```

so downstream systems understand the result.

---

# 43. Fallback ETA

When no provider is available, RideForge may use a deterministic fallback only when the business use case permits it.

For example:

```text
Straight-Line Distance
+
Historical Speed Estimate
```

may produce a coarse estimate.

It must not be presented as equivalent to a live traffic ETA.

---

# 44. Fallback Labeling

Fallback results should be distinguishable internally.

Conceptually:

```text
source = DETERMINISTIC_FALLBACK
```

This is important for:

```text
Observability
Analytics
AI Training
Quality Evaluation
```

---

# 45. Dispatch Fallback

If ETA provider failure occurs during Smart Dispatch:

```text
AI / ETA unavailable
       ↓
Deterministic dispatch ranking
```

may be used where possible.

The dispatch system must never assign an otherwise ineligible driver merely because ETA is unavailable.

---

# 46. Cached ETA

Recent ETA results may be cached when the use case allows.

Potential cache key:

```text
origin_bucket
destination_bucket
vehicle_type
route_profile
```

The actual key design must account for location precision and freshness.

---

# 47. Cache Freshness

Cached ETA must have a short, explicit TTL appropriate to the use case.

The TTL should not be globally assumed to be correct for every consumer.

---

# 48. Cache Is Not Source of Truth

The cache is an optimization layer.

The provider or ETA engine remains the source of the calculation.

---

# 49. Cache Failure

If Redis is unavailable:

```text
Do Not Fail Entire ETA System Automatically
```

The system should call the provider directly when safe and practical.

---

# 50. Negative Caching

Negative provider responses such as:

```text
No Route
```

may be cached only when the semantics are stable enough to justify it.

Avoid caching transient provider failures as permanent route failures.

---

# 51. Route Geometry Caching

Route geometry may be cached for repeated routes when:

```text
Route Stability
Storage Cost
Privacy
Freshness
```

requirements permit.

Traffic-sensitive duration should not automatically inherit the same TTL as static route geometry.

---

# 52. Static Route vs Traffic ETA

The architecture should distinguish:

```text
Static Route Geometry
```

from:

```text
Traffic-Aware Duration
```

A route geometry may remain valid while travel time changes.

---

# 53. Current Driver ETA

For a driver approaching a pickup:

```text
Driver Location
+
Pickup Location
+
Current Time
```

should be used where live ETA is required.

---

# 54. Location Freshness

ETA quality depends directly on driver location freshness.

A stale driver location can produce a stale ETA even when the route provider itself is healthy.

---

# 55. Driver Location Integration

The ETA system should consume the approved driver-location architecture defined by:

```text
ADR-0010 — Driver Location Storage Strategy
```

It should not create an independent driver-location source.

---

# 56. ETA Update Frequency

ETA refresh frequency should depend on:

```text
Distance
Trip Phase
Traffic Volatility
Driver Movement
Consumer
```

There is no requirement for every consumer to request a new route calculation at every second.

---

# 57. Avoid Excessive Provider Calls

Calling a route provider continuously can create:

```text
High Cost
Rate Limiting
Latency
Unnecessary Load
```

Use appropriate:

```text
Refresh Intervals
Caching
Location Thresholds
Change Detection
```

---

# 58. Location Change Threshold

A new ETA calculation may be triggered when the driver's location changes sufficiently.

Conceptually:

```text
Small Location Change
→ Reuse Current ETA

Significant Location Change
→ Recalculate
```

The exact threshold depends on operational testing.

---

# 59. ETA Change Threshold

Similarly, a meaningful ETA change may be required before publishing a new value to clients.

This reduces unnecessary updates.

---

# 60. Route Recalculation

A route may need recalculation when:

```text
Driver Deviates
Road Closure Occurs
Destination Changes
Traffic Changes Significantly
Waypoint Changes
```

---

# 61. Route Recalculation Policy

Recalculation should be bounded.

Do not continuously recompute a route merely because a location update arrived.

---

# 62. Route Provider Vehicle Profiles

If supported, the platform should use the appropriate vehicle profile.

Examples:

```text
Car
Motorcycle
Taxi
Commercial Vehicle
```

The route provider adapter should translate RideForge vehicle categories into provider-specific profiles.

---

# 63. Vehicle Profile Mismatch

If a provider cannot represent a required vehicle profile:

```text
Use Approved Approximation
```

only if policy allows it.

Otherwise:

```text
Use Another Provider
```

---

# 64. Route Preferences

RideForge may eventually support preferences such as:

```text
Fastest Route
Shortest Route
Avoid Toll
Avoid Highway
Vehicle Restrictions
```

These should be represented as platform-level preferences and translated by the provider adapter.

---

# 65. Provider-Specific Route Options

Provider-specific options should not leak into the core domain unless they become part of a genuine RideForge capability.

---

# 66. Regional Route Behaviour

Different regions may require different routing behaviour because of:

```text
Road Network
Traffic
Border Rules
Vehicle Restrictions
Provider Coverage
```

Provider policy may therefore be region-specific.

---

# 67. Border and Legal Considerations

Routing must not be treated as authorization to perform a ride.

A route provider may return a route across a regional boundary even when RideForge business policy prohibits that ride.

Therefore:

```text
Route Provider
≠
Legal Authorization
```

---

# 68. Legal Validation Boundary

Ride legality remains governed by:

```text
ADR-0018 — Regional and Legal Ride Validation
```

The routing system provides route information; it does not decide whether the ride is legally permitted.

---

# 69. Route Provider and Cross-Region Routes

When a route crosses regions:

```text
Route Calculation
```

may still succeed.

RideForge must separately evaluate:

```text
Regional Policy
Legal Constraints
Driver Eligibility
```

before dispatch.

---

# 70. ETA Provider and Dispatch

Smart Dispatch may use ETA as a ranking signal:

```text
Candidate Drivers
      ↓
ETA
      ↓
Ranking
```

but ETA must remain one signal among the defined dispatch features.

---

# 71. ETA Provider and AI

AI-assisted dispatch may consume provider ETA as a feature.

Conceptually:

```text
Provider ETA
    ↓
Feature Engineering
    ↓
AI Model
    ↓
Candidate Score
```

Provider ETA should remain identifiable as an input source.

---

# 72. AI-Predicted ETA

Future models may predict ETA directly:

```text
Route + Traffic + Historical Data
        ↓
AI ETA Model
        ↓
Predicted ETA
```

This should be introduced through a separate model lifecycle and validated against provider and actual travel outcomes.

---

# 73. Provider ETA vs AI ETA

The system may eventually support:

```text
Provider ETA
AI ETA
Hybrid ETA
```

The selection should be policy-driven.

---

# 74. Hybrid ETA

A future hybrid model may combine:

```text
Provider Route
+
Provider Travel Time
+
Historical Data
+
AI Correction
```

This can improve local prediction accuracy if validated.

---

# 75. AI Must Not Silently Replace Provider ETA

Any AI ETA replacement should be:

```text
Explicit
Versioned
Monitored
Experimented
Reversible
```

---

# 76. Actual Travel Time

RideForge should record actual travel outcomes where permitted:

```text
Pickup Arrival Time
Trip Start
Trip End
Actual Route
Actual Duration
```

These can be used to evaluate ETA quality.

---

# 77. ETA Accuracy Metrics

Measure:

```text
MAE
RMSE
MAPE
Median Absolute Error
P50 Error
P95 Error
```

The final metric set should reflect the actual product requirement.

---

# 78. ETA Bias

Monitor whether ETA systematically:

```text
Overestimates
```

or:

```text
Underestimates
```

actual travel time.

---

# 79. ETA Calibration

For predicted arrival probabilities or ranges, evaluate calibration where applicable.

---

# 80. Provider Comparison

When multiple providers are available, compare:

```text
Accuracy
Latency
Coverage
Cost
Availability
```

using controlled measurements.

---

# 81. Provider Evaluation

A new provider should not be selected solely because:

```text
Price Is Lower
```

or:

```text
Demo ETA Looks Better
```

Evaluate production-relevant workloads.

---

# 82. Shadow Provider

A candidate provider may be evaluated in shadow mode:

```text
Primary Provider
→ Real Result

Candidate Provider
→ Shadow Result
```

The candidate does not control the production route.

---

# 83. Provider A/B Testing

Provider experiments should consider:

```text
Region
Traffic
Ride Type
Vehicle Type
Time
```

to avoid comparing providers under unrelated traffic conditions.

---

# 84. Provider Cost Monitoring

Track:

```text
Requests
Cost per Request
Cost per Ride
Cost per Region
Cache Hit Rate
Fallback Rate
```

---

# 85. Provider Quotas

Provider quotas should be monitored before approaching limits.

Alerts should be configured for:

```text
Quota Utilization
Rate Limit Events
Unexpected Traffic Increase
```

---

# 86. Provider Contract Changes

External provider contracts may change.

Provider adapters should isolate:

```text
API Version
Request Format
Response Format
Authentication
Error Mapping
```

from the rest of the platform.

---

# 87. Provider API Versioning

Provider API versions should be explicitly configured and tested.

Do not depend on an unpinned provider API version when the provider supports version selection.

---

# 88. Provider Authentication

Provider credentials must be managed through the approved secret-management mechanism.

Never embed provider credentials in:

```text
Source Code
Git Repository
Model Artifacts
Logs
```

---

# 89. Provider Security

External provider calls should use:

```text
TLS
Credential Rotation
Least Privilege
Request Validation
Response Validation
```

as appropriate.

---

# 90. Response Validation

Never assume an external provider response is valid.

Validate:

```text
HTTP Status
Schema
Required Fields
Distance Range
Duration Range
Geometry
Provider Metadata
```

---

# 91. Invalid Provider Response

If a provider returns malformed or impossible data:

```text
Reject Response
Record Provider Error
Fallback
```

Do not propagate invalid route data into dispatch.

---

# 92. Impossible ETA

Examples:

```text
Negative Duration
Zero Duration for Long Route
Extreme Distance
Invalid Coordinates
```

should trigger validation failure.

---

# 93. Coordinate Validation

Before provider calls, validate:

```text
Latitude Range
Longitude Range
Coordinate Presence
Coordinate Precision
```

---

# 94. Provider Input Normalization

RideForge should normalize:

```text
Coordinates
Vehicle Type
Units
Route Preferences
Region
```

before sending provider requests.

---

# 95. Unit Normalization

Internally standardize units such as:

```text
Distance → meters
Duration → seconds
Coordinates → decimal degrees
```

Provider-specific units should be converted inside the adapter.

---

# 96. Time Handling

ETA calculations should use:

```text
UTC / Instant-Based Timestamps
```

internally where appropriate.

Local timezone handling should remain at presentation or policy boundaries.

---

# 97. Request Idempotency

Route/ETA queries are generally read operations, but request retries should still be bounded.

The system must avoid accidental duplicate billable requests caused by uncontrolled retry loops.

---

# 98. Provider Request Deduplication

Where appropriate, concurrent identical requests may be coalesced.

Conceptually:

```text
Request A ─┐
Request B ─┼→ Shared Provider Request
Request C ─┘
```

This can reduce provider cost and load.

This optimization should only be introduced when measured demand justifies it.

---

# 99. ETA Request Coalescing

High-frequency ETA requests for the same driver/ride may be combined when:

```text
Same Origin
Same Destination
Freshness Window
Same Vehicle Profile
```

are satisfied.

---

# 100. Rate Limiting Inside RideForge

RideForge may apply internal rate limits before calling external providers.

This protects provider quotas and system resources.

---

# 101. Backpressure

ETA workloads should use bounded:

```text
Concurrency
Connection Pools
Queues
Worker Pools
```

where required.

---

# 102. Provider Connection Reuse

HTTP clients should reuse connections.

Avoid creating a new HTTP client for every route request.

---

# 103. Provider Timeout Hierarchy

The timeout hierarchy should remain bounded:

```text
Overall Request Deadline
      ↓
ETA / Route Operation Deadline
      ↓
Provider Request Timeout
```

A provider should not consume the entire application request budget.

---

# 104. Observability

Every provider call should expose appropriate metrics:

```text
Request Count
Success Count
Error Count
Timeout Count
Latency
Provider
Region
Operation
Fallback Count
```

---

# 105. Provider Health Dashboard

Operations should be able to see:

```text
Provider Availability
P50 Latency
P95 Latency
P99 Latency
Error Rate
Rate Limit Rate
Fallback Rate
Cost
```

---

# 106. ETA Quality Dashboard

Monitor:

```text
Predicted ETA
Actual Arrival
Absolute Error
Bias
Provider
Region
Ride Type
```

---

# 107. Distributed Tracing

Route/ETA calls should be traceable:

```text
Ride Request
   ↓
Dispatch
   ↓
ETA Service
   ↓
Provider
```

This makes latency investigation easier.

---

# 108. Logging

Provider integration logs should include:

```text
Request ID
Trace ID
Provider
Operation
Latency
Status
Error Category
Fallback
```

Do not log unnecessary:

```text
Precise Personal Locations
Credentials
Sensitive Payloads
```

---

# 109. Privacy

Location data is sensitive operational data.

Access to route and ETA data should follow:

```text
Least Privilege
Purpose Limitation
Retention Rules
Access Control
Auditability
```

---

# 110. Data Retention

Provider request/response data should not be retained indefinitely unless there is a documented reason.

Store only what is needed for:

```text
Operations
Analytics
Billing / Provider Reconciliation
Model Evaluation
Dispute Handling
```

---

# 111. Route Geometry Privacy

Detailed route geometry can expose sensitive travel information.

Retention and access should therefore be controlled.

---

# 112. Analytics Separation

Analytical workloads should not directly overload the real-time ETA provider integration.

Use:

```text
Events
Data Pipelines
Read Models
```

where appropriate.

---

# 113. Event-Based ETA Outcomes

Actual ETA outcomes may be published through the event system for:

```text
Analytics
AI Training
Monitoring
```

The event architecture remains governed by:

```text
ADR-0005
ADR-0006
ADR-0012
ADR-0013
```

---

# 114. ETA Event Examples

Potential events include:

```text
ETAUpdated
RouteCalculated
PickupArrived
TripStarted
TripCompleted
```

The exact event catalog is maintained separately.

---

# 115. Provider Fallback and Events

Fallback usage may be included in internal telemetry.

For example:

```text
eta_source = PRIMARY
```

or:

```text
eta_source = SECONDARY
```

or:

```text
eta_source = FALLBACK
```

---

# 116. AI Training Data

ETA provider results should not automatically become ground truth.

For model training, distinguish:

```text
Predicted ETA
```

from:

```text
Actual Travel Time
```

The actual observed outcome is the relevant target for ETA prediction.

---

# 117. Provider ETA as Feature

Provider ETA can be used as a feature, but training pipelines should retain:

```text
Provider
Timestamp
Prediction
Actual Outcome
```

where needed.

---

# 118. ETA Feedback Loop

The preferred feedback loop is:

```text
ETA Prediction
     ↓
Ride Execution
     ↓
Actual Outcome
     ↓
Error Calculation
     ↓
Monitoring / Training
```

---

# 119. AI ETA Future Architecture

A future architecture may be:

```text
Route Provider
      ↓
Route
      │
      ├── Traffic Data
      ├── Historical Travel Time
      └── Real-Time Features
                 ↓
             AI ETA Model
                 ↓
             ETA Engine
                 ↓
             Consumers
```

This should be introduced only after sufficient data and evaluation.

---

# 120. Provider vs AI Decision

A future policy may define:

```text
Provider ETA
```

for:

```text
Low Data Regions
Cold Start
Model Failure
Fallback
```

and:

```text
AI ETA
```

for:

```text
Well-Supported Regions
High Data Availability
Validated Models
```

---

# 121. Cold Start

New regions may not have enough historical data for reliable AI ETA.

The system should therefore support:

```text
Provider-Based ETA
```

during cold start.

---

# 122. Model Failure

If an AI ETA model fails:

```text
AI ETA
   ↓
Provider ETA
```

should remain available where configured.

---

# 123. Provider Failure

If a provider fails:

```text
Provider ETA
   ↓
Secondary Provider
   ↓
AI ETA / Deterministic Fallback
```

may be used according to policy.

---

# 124. Fallback Hierarchy

A representative hierarchy is:

```text
Primary Provider
       ↓
Secondary Provider
       ↓
Approved AI / Historical Estimate
       ↓
Deterministic Fallback
```

The exact hierarchy is configuration-dependent.

---

# 125. Fallback Quality

Fallback quality should be measured separately.

Do not mix:

```text
Primary ETA Accuracy
```

with:

```text
Fallback ETA Accuracy
```

in a way that hides provider problems.

---

# 126. Route Provider and Pricing

Pricing may use distance and duration.

Pricing must define whether it uses:

```text
Estimated Route
Actual Route
Estimated Duration
Actual Duration
```

and this must be consistent with the pricing architecture.

---

# 127. Route Provider and Dispatch

Dispatch should prefer:

```text
Pickup ETA
```

over raw route distance when reliable ETA is available.

---

# 128. Route Provider and Rider App

Rider-facing ETA should be generated by a centralized ETA capability rather than independently calculated by every client.

This prevents inconsistent ETA logic across:

```text
Android
iOS
Web
Driver App
```

---

# 129. Client Responsibilities

Clients should display server-provided ETA values.

They should not independently choose external providers for core RideForge ETA.

---

# 130. ETA Consistency

A rider and driver should not receive incompatible ETA calculations merely because their clients use different map providers.

RideForge should define the authoritative ETA representation for each product surface.

---

# 131. Operational Overrides

Operations may need to temporarily override provider selection.

Such overrides must be:

```text
Authorized
Audited
Time-Bounded
Reversible
Observable
```

---

# 132. Provider Maintenance Mode

A provider can be temporarily disabled:

```text
Provider A
→ Disabled

Provider B
→ Active
```

without changing domain code.

---

# 133. Provider Configuration

Configuration should include:

```text
Provider Enabled
API Endpoint
Credential Reference
Timeout
Retry Policy
Rate Limit
Region
Capabilities
Priority
```

Secrets must remain outside normal configuration values where the deployment platform supports secret references.

---

# 134. Provider Priority

A provider priority may be represented as:

```text
Primary
Secondary
Disabled
```

Avoid complex dynamic provider scoring until the platform actually needs it.

---

# 135. Provider Selection Simplicity

The initial provider selection should remain deterministic:

```text
Region
+
Capability
+
Configured Priority
```

Dynamic selection based on real-time provider health can be introduced later if justified.

---

# 136. Health-Based Provider Selection

Future provider selection may consider:

```text
Current Error Rate
Current Latency
Quota
Cost
```

but should remain bounded and observable.

---

# 137. Avoid Provider Flapping

If provider selection changes too frequently:

```text
Provider A
→ B
→ A
→ B
```

metrics and behaviour become difficult to reason about.

Use stable selection with controlled failover and recovery.

---

# 138. Provider Recovery

When a failed provider recovers:

```text
Open Circuit
   ↓
Probe
   ↓
Healthy
   ↓
Controlled Return to Primary
```

Do not immediately send full traffic to a previously failing provider.

---

# 139. Provider Canary

A recovered or newly deployed provider integration may receive limited traffic before becoming primary.

---

# 140. Testing

The ETA and routing system requires:

```text
Unit Tests
Integration Tests
Provider Contract Tests
Failure Tests
Load Tests
Latency Tests
Accuracy Tests
```

---

# 141. Provider Adapter Tests

Each adapter should test:

```text
Request Mapping
Response Mapping
Error Mapping
Authentication
Timeout
Rate Limit
Malformed Response
```

---

# 142. Contract Tests

Provider integrations should use provider-compatible contract fixtures where possible.

---

# 143. Mocking

Unit tests should not call live external providers.

Use:

```text
Mocks
Fakes
Recorded Fixtures
```

for unit and most integration tests.

---

# 144. Integration Tests

Real provider calls may be used in controlled integration environments where:

```text
Credentials
Cost
Rate Limits
Provider Terms
```

permit them.

---

# 145. Failure Tests

Test:

```text
Timeout
5xx
429
Malformed Response
Network Failure
Invalid Route
No Route
```

and verify fallback behaviour.

---

# 146. Accuracy Tests

Compare provider and ETA outputs against known routes and observed travel outcomes where available.

---

# 147. Load Tests

Measure:

```text
Requests per Second
Latency
Provider Quota Consumption
Cache Hit Rate
Fallback Rate
```

under expected traffic.

---

# 148. Cost Tests

Measure provider cost under:

```text
Normal Traffic
Peak Traffic
Cache Enabled
Cache Disabled
Failover
```

---

# 149. Security Tests

Verify:

```text
Credential Handling
TLS
Authorization
Sensitive Logging
Provider Access
```

---

# 150. Deployment

Provider adapters should be deployed as part of the appropriate service boundary.

Changing a provider should not require changing every consumer.

---

# 151. Configuration Deployment

Provider changes should preferably be configurable without rebuilding business logic.

However, configuration changes must still pass validation and deployment controls.

---

# 152. Environment Separation

Provider configuration should differ across:

```text
Development
Staging
Production
```

Never use production provider credentials in local development.

---

# 153. Local Development

Local development should support a deterministic fake provider.

For example:

```text
Fake Route Provider
Fake ETA Provider
```

This allows developers to work without external API costs.

---

# 154. Fake Provider Behaviour

The fake provider should support scenarios such as:

```text
Success
Timeout
Error
No Route
Slow Response
Malformed Response
```

for development and testing.

---

# 155. Production Provider Isolation

Production provider credentials should be accessible only to the services that require them.

---

# 156. Observability Checklist

```text
[ ] Provider request count
[ ] Provider latency
[ ] Provider error rate
[ ] Provider timeout rate
[ ] Provider rate-limit rate
[ ] Provider fallback rate
[ ] ETA accuracy
[ ] ETA bias
[ ] Cache hit rate
[ ] Cost per request
[ ] Cost per ride
[ ] Trace propagation
[ ] Provider health
```

---

# 157. Operational Checklist

```text
[ ] Primary provider configured
[ ] Secondary provider configured where justified
[ ] Provider capabilities documented
[ ] Timeout configured
[ ] Retry policy configured
[ ] Circuit breaker configured where required
[ ] Rate limits monitored
[ ] Credentials secured
[ ] Fallback defined
[ ] Kill switch available
[ ] Provider recovery tested
```

---

# 158. Decision Matrix

| Capability | Primary Provider | Secondary Provider | AI / Prediction | Deterministic Fallback |
|---|---:|---:|---:|---:|
| Route geometry | **Yes** | Optional | No | Limited |
| Traffic ETA | **Yes** | Optional | Optional | Limited |
| Dispatch ETA | **Yes** | Optional | Optional | Yes |
| Rider ETA | **Yes** | Optional | Optional | Limited |
| Historical ETA | No | No | **Yes** | Yes |
| Provider outage | No | **Yes** | Optional | **Yes** |
| New-region cold start | **Yes** | Optional | Usually No | Yes |
| AI model failure | Provider | Optional | No | **Yes** |

---

# 159. Alternatives Considered

## 159.1 Single Provider Everywhere

### Advantages

```text
Simple
Low Operational Complexity
Easy Integration
```

### Disadvantages

```text
Single Vendor Dependency
Outage Risk
Regional Coverage Risk
Pricing Risk
```

### Decision

```text
Rejected as the long-term architectural assumption.
```

A single primary provider may still be used initially.

---

# 160. Direct Provider Calls From Every Service

Example:

```text
Dispatch → Provider
Pricing → Provider
Rider API → Provider
Driver API → Provider
```

### Disadvantages

```text
Duplicated Logic
Credential Distribution
Inconsistent ETA
Provider Lock-In
Difficult Provider Replacement
```

### Decision

```text
Rejected.
```

---

# 161. Centralized Routing / ETA Capability

Example:

```text
All Consumers
      ↓
ETA / Routing Service
      ↓
Provider Layer
```

### Advantages

```text
Centralized Policy
Consistent ETA
Centralized Caching
Provider Isolation
Observability
Cost Control
```

### Decision

```text
Accepted as the preferred architecture.
```

---

# 162. AI-Only ETA

### Disadvantages

```text
Cold Start
Training Dependency
Model Failure
Data Requirements
Operational Complexity
```

### Decision

```text
Rejected as the sole ETA source.
```

AI may complement provider ETA.

---

# 163. Provider-Only ETA Forever

### Disadvantages

```text
Limited Local Optimization
Vendor Dependency
Potentially Higher Cost
Limited Learning From RideForge-Specific Data
```

### Decision

```text
Rejected as the permanent architectural assumption.
```

AI-based ETA may be introduced later.

---

# 164. Consequences

## 164.1 Positive Consequences

The decision provides:

```text
Provider Independence
Centralized ETA Logic
Controlled Failover
Cost Management
Regional Flexibility
AI Compatibility
Consistent Client Behaviour
Better Observability
```

---

## 164.2 Negative Consequences

The architecture introduces:

```text
Provider Abstraction Layer
Provider Configuration
Fallback Complexity
Caching Complexity
ETA Quality Monitoring
Multiple Integration Tests
```

These trade-offs are accepted.

---

# 165. Risks

## Risk 1 — Provider Lock-In

### Mitigation

```text
Internal Provider Interface
Adapter Pattern
Provider-Agnostic Domain Models
```

---

## Risk 2 — Provider Outage

### Mitigation

```text
Secondary Provider
Circuit Breaker
Fallback
Cached Results
```

---

## Risk 3 — Provider Cost Explosion

### Mitigation

```text
Caching
Request Coalescing
Rate Limits
Cost Monitoring
Provider Selection
```

---

## Risk 4 — Stale ETA

### Mitigation

```text
Freshness Metadata
TTL
Location Refresh
Consumer-Specific Freshness Rules
```

---

## Risk 5 — Invalid Provider Data

### Mitigation

```text
Response Validation
Range Checks
Fallback
Provider Monitoring
```

---

## Risk 6 — AI ETA Regression

### Mitigation

```text
Provider ETA Fallback
Shadow Mode
Canary
Offline Evaluation
Rollback
```

---

## Risk 7 — Regional Coverage Problems

### Mitigation

```text
Provider Capability Matrix
Regional Provider Policy
Secondary Provider
```

---

# 166. Validation

This ADR should be validated through:

```text
Provider Adapter Tests
Integration Tests
Provider Failure Tests
Latency Tests
Load Tests
Cost Tests
ETA Accuracy Evaluation
Shadow Provider Testing
AI ETA Shadow Testing
Failover Tests
Rollback Tests
```

---

# 167. Review Triggers

Revisit this ADR when:

```text
A New Route Provider Is Proposed
Provider Costs Change Significantly
Provider Coverage Changes
A Provider Has Repeated Outages
AI ETA Is Ready for Production
Multi-Region Routing Becomes Significant
Route Requirements Become More Complex
Real-Time Traffic Requirements Change
A Dedicated Routing Infrastructure Is Introduced
```

---

# 168. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
```

Especially:

```text
ETA and Prediction System
Smart Dispatch AI
Driver Demand and Supply Prediction
AI Matching and Ranking
Feature Engineering
Model Training and Evaluation
Model Serving and Inference
Online and Offline Features
AI Failure and Fallback Strategy
Performance and Optimization
Redis Development
Event and Messaging Development
API Development
Observability Development
Data Privacy and Governance
```

---

# 169. Related ADRs

This decision is directly related to:

```text
ADR-0003 — Microservice Boundaries
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0008 — PostGIS for Geospatial Operations
ADR-0009 — Redis for Real-Time State and Caching
ADR-0010 — Driver Location Storage Strategy
ADR-0014 — API and Service Communication
ADR-0015 — Smart Dispatch and Stand Dispatch
ADR-0016 — AI-Assisted Dispatch Strategy
ADR-0018 — Regional and Legal Ride Validation
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
ADR-0023 — Security and Secret Management
ADR-0024 — Configuration and Environment Strategy
ADR-0025 — Testing and Integration Strategy
ADR-0026 — Model and AI Governance
ADR-0028 — Cost Optimization Strategy
```

---

# 170. Decision Summary

RideForge adopts a provider-agnostic routing and ETA architecture:

```text
                     RideForge
                         │
                         ▼
                ETA / Routing API
                         │
                         ▼
                 Provider Policy
                         │
            ┌────────────┼────────────┐
            ▼            ▼            ▼
        Primary      Secondary      AI ETA
        Provider      Provider       Model
            │            │            │
            └────────────┼────────────┘
                         ▼
                  ETA / Route Result
                         │
              ┌──────────┼──────────┐
              ▼          ▼          ▼
           Dispatch     Rider      Driver
```

The preferred failure hierarchy is:

```text
Primary Provider
       ↓
Secondary Provider
       ↓
Approved AI / Historical Estimate
       ↓
Deterministic Fallback
```

where the configured use case permits each level.

The architecture separates:

```text
Route Calculation
ETA Prediction
Provider Integration
Provider Selection
Caching
Dispatch Consumption
AI Prediction
```

so each concern can evolve independently.

---

# 171. Final Principle

> **RideForge will treat routing and ETA as platform capabilities behind a provider-independent abstraction, use a configured primary provider with bounded failover, keep provider-specific logic inside adapters, and allow future AI-based ETA improvements without making AI or any single external provider a mandatory source of truth.**

The operational hierarchy is:

```text
RideForge ETA / Routing Capability
            ↓
     Provider Selection
            ↓
   Primary / Secondary
            ↓
     Validated Result
            ↓
     Consumer-Specific
      Freshness Policy
            ↓
       Dispatch / UX
```

The system must remain capable of:

```text
Changing Providers
Adding Providers
Disabling Providers
Failing Over
Caching Safely
Measuring Accuracy
Controlling Cost
Introducing AI ETA
Rolling Back AI ETA
Operating During Provider Outages
```

without changing the core ride domain.

---

# 172. Status

```text
Decision: ACCEPTED

Routing Architecture:
Provider-Agnostic

ETA Architecture:
Provider-Agnostic

Primary Provider:
Configured

Secondary Provider:
Optional / Recommended Where Operationally Justified

Provider Integration:
Adapter-Based

Provider-Specific Logic:
Infrastructure Layer Only

ETA Freshness:
Explicit

Caching:
Allowed With Explicit TTL / Freshness Policy

AI ETA:
Future / Optional

AI Failure:
Provider ETA / Deterministic Fallback

Provider Failure:
Secondary Provider / Approved Fallback

Legal Authorization:
Outside Routing System

Primary Goal:
Accurate, Available, Cost-Controlled and Replaceable ETA / Routing
```
