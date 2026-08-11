# ADR-0018: Regional and Legal Ride Validation

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Regulatory / Domain Validation / Dispatch Safety  
> **Scope:** Regional eligibility, permit-aware ride validation, cross-region operations, and legal policy enforcement  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is intended to operate in real geographic markets where transportation rules can vary by:

```text
State
Region
Permit Area
Vehicle Category
Driver Category
Ride Type
Operating Authority
Contract / Permit Conditions
```

This is especially important for RideForge's intended operating environment around the Kerala–Karnataka border, where a ride may involve:

```text
Pickup Region
Destination Region
Driver Registration State
Driver Permit
Vehicle Type
Permit Area
Inter-State Authorization
```

A route provider may be able to calculate a route across a state boundary, but that does **not** establish that the ride is legally permitted.

The Motor Vehicles Act, 1988 contains permit-related provisions including the necessity for permits, contract-carriage permits, permit conditions, and validation of permits for use outside the region in which they were granted. citeturn0search0turn0search3turn0search6

Kerala's Motor Vehicles Department materials also distinguish vehicles operating interstate under the relevant permit framework, demonstrating that interstate operation must be treated as a regulated capability rather than as an automatic consequence of geographic routing. citeturn0search21turn0search22

Therefore, RideForge must make legal/regulatory eligibility an explicit part of the ride lifecycle.

---

# 2. Problem

RideForge must determine:

```text
Whether a ride is permitted under the platform's configured rules
Whether a driver is eligible for the ride
Whether the vehicle is eligible
Whether the requested origin/destination crosses a regulated boundary
Whether the driver's permit covers the intended operation
Whether an inter-state operation is authorized
Whether dispatch must reject or restrict a ride
How regulatory rules are represented
How rule changes are deployed
How validation decisions are audited
```

The system must avoid:

```text
Assuming Same-State Operation Is Always Allowed
Assuming Cross-State Operation Is Always Allowed
Treating Route Availability as Legal Authorization
Hard-Coding One Region's Rules
Allowing AI to Override Legal Rules
Allowing Dispatch to Bypass Validation
```

---

# 3. Decision

RideForge will implement **regional and legal ride validation as a deterministic policy boundary before dispatch and assignment**.

The core decision flow is:

```text
Ride Request
     │
     ▼
Geographic Resolution
     │
     ▼
Regional Policy Resolution
     │
     ▼
Ride Legal Eligibility
     │
     ▼
Driver / Vehicle Eligibility
     │
     ▼
Dispatch Strategy
     │
     ▼
Assignment
```

The primary rule is:

> **A route being technically possible does not make a ride legally eligible. Legal and regional validation must be satisfied before dispatch can assign a driver.**

---

# 4. Legal Validation Is a Hard Constraint

Legal eligibility is not a ranking signal.

It must be evaluated before:

```text
Smart Ranking
AI Ranking
Stand Assignment
Driver Reservation
Final Assignment
```

The architecture is:

```text
Legal Validation
      ↓
Eligible Ride
      ↓
Eligible Drivers
      ↓
Dispatch
```

not:

```text
Dispatch
   ↓
Check Legal Rules Later
```

---

# 5. Legal Rules Are Not Hard-Coded in Dispatch Algorithms

Dispatch should not contain scattered rules such as:

```text
if Kerala -> ...
if Karnataka -> ...
if border -> ...
```

Instead:

```text
Ride
 ↓
Regional / Legal Policy Engine
 ↓
Validation Result
 ↓
Dispatch
```

This keeps regulatory policy separate from dispatch algorithms.

---

# 6. Regulatory Policy vs Application Logic

RideForge must distinguish between:

```text
Regulatory Policy
```

and:

```text
Application Implementation
```

Regulatory policy determines:

```text
What Is Allowed
```

Application logic determines:

```text
How That Rule Is Enforced
```

---

# 7. Legal Authority

RideForge must treat official laws, regulations, permit conditions, government notifications, transport-authority requirements, and applicable legal advice as the authoritative sources for regulatory policy.

The application itself is **not** the legal authority.

The platform must not assume that a rule remains valid indefinitely.

---

# 8. Current Legal Reference

The Motor Vehicles Act, 1988 includes:

```text
Section 66 — Necessity for permits
Section 73 — Application for contract carriage permit
Section 74 — Grant of contract carriage permit
Section 84 — General conditions attaching to permits
Section 86 — Cancellation and suspension of permits
Section 88 — Validation of permits for use outside the region
Section 88A — Inter-State transport schemes
Section 93 — Agent / aggregator licensing
```

These provisions demonstrate that permit and operating-area requirements are material inputs to the platform's regulatory model. citeturn0search0turn0search3turn0search6

This ADR does **not** attempt to interpret every provision or provide legal advice.

---

# 9. No Universal Cross-Region Assumption

RideForge must not implement:

```text
Origin Region ≠ Destination Region
→ Always Allowed
```

or:

```text
Origin Region ≠ Destination Region
→ Always Rejected
```

Instead:

```text
Cross-Region
     ↓
Applicable Policy
     ↓
Permit / Authorization
     ↓
Eligibility Decision
```

---

# 10. Regional Classification

RideForge should maintain a canonical geographic classification.

Conceptually:

```text
Country
   ↓
State
   ↓
Transport Region
   ↓
City / Municipality
   ↓
Zone
   ↓
Pickup / Destination Area
```

The exact hierarchy depends on the operating market.

---

# 11. Region Identity

Each relevant operating region should have a stable internal identifier.

Example:

```text
region_id = KL
region_id = KA
```

The actual identifiers should be configuration-owned and should not be derived ad hoc throughout the application.

---

# 12. Geographic Resolution

A ride request should resolve:

```text
pickup_region
destination_region
pickup_zone
destination_zone
```

before legal validation.

The geographic resolver may use:

```text
PostGIS
Administrative Boundary Data
Geocoding
Configured Zones
```

according to the platform architecture.

---

# 13. Coordinate Is Not Legal Status

A GPS coordinate alone does not establish complete legal eligibility.

For example:

```text
Coordinate
→ Region
```

is only one input.

Other inputs may include:

```text
Driver Permit
Vehicle Permit
Ride Type
Operating Authority
Applicable Policy
```

---

# 14. Region Resolution Confidence

If geographic resolution is ambiguous:

```text
Do Not Guess
```

The system should use:

```text
Boundary Resolution
Configured Administrative Data
Manual Review
Safe Rejection
```

according to operational policy.

---

# 15. Boundary Handling

Border areas require special treatment.

A coordinate near a boundary may be subject to:

```text
Geocoding Error
Boundary Data Error
GPS Accuracy
Administrative Boundary Complexity
```

Therefore RideForge should use authoritative boundary datasets and explicit tolerances.

---

# 16. Border Zone

The platform may define a controlled:

```text
Border Zone
```

for locations where regional classification requires additional verification.

This does not itself mean:

```text
Allowed
```

or:

```text
Rejected
```

It means:

```text
Additional Validation Required
```

---

# 17. Origin Region

The origin region represents where the ride begins.

It is a core input to:

```text
Ride Validation
Permit Validation
Dispatch Policy
```

---

# 18. Destination Region

The destination region represents where the ride is intended to end.

It may affect:

```text
Permit Compatibility
Inter-State Operation
Regional Policy
Driver Eligibility
```

---

# 19. Route Region

A ride may pass through additional regions without ending there.

Therefore the system may eventually need:

```text
Origin Region
Destination Region
Route Regions
```

However, the initial implementation should not over-engineer route-region legal inference.

Only introduce route-region validation where the applicable regulatory policy requires it.

---

# 20. Driver Home Region

A driver's registration or operating region may differ from:

```text
Current Location
```

and:

```text
Ride Origin
```

Therefore:

```text
Current GPS Region
```

must not be treated as:

```text
Driver Legal Operating Region
```

---

# 21. Driver Registration

The platform should retain relevant vehicle/driver registration information required for regulatory validation.

Examples may include:

```text
Registration State
Vehicle Registration
Permit Information
Permit Validity
Vehicle Category
```

The exact data model belongs to the driver/vehicle domain.

---

# 22. Permit Information

Where applicable, RideForge should represent permit-related data such as:

```text
permit_id
permit_type
issuing_authority
valid_from
valid_until
authorized_region
authorized_route
vehicle_id
status
```

Only fields required by the actual regulatory model should be implemented.

---

# 23. Permit Is an External Legal Artifact

The platform's stored permit data is a representation of an external legal/administrative authorization.

It must not be treated as authoritative merely because it exists in the RideForge database.

---

# 24. Permit Verification

Where practical, permit information should be:

```text
Verified
Timestamped
Audited
Renewed
Revoked When Necessary
```

The verification mechanism depends on available authority integrations and operational processes.

---

# 25. Permit Expiry

An expired permit must not be treated as valid.

Conceptually:

```text
Permit Valid
   ↓
Expiry
   ↓
Permit Invalid
```

The platform should support proactive expiry detection.

---

# 26. Permit Suspension

A suspended permit must be treated as unavailable for the affected operation.

---

# 27. Permit Revocation

A revoked permit must immediately prevent applicable operations once the platform receives reliable information about the revocation.

---

# 28. Missing Permit Data

If a ride requires permit validation and the required permit information is unavailable:

```text
Do Not Assume Validity
```

The policy may choose:

```text
Reject
Manual Review
Restrict Operation
```

depending on the business and legal requirement.

---

# 29. Permit Scope

Permit validation must consider scope.

A permit may contain conditions relating to:

```text
Area
Route
Vehicle
Passenger Capacity
Type of Operation
Other Conditions
```

The Motor Vehicles Act specifically permits conditions to be attached to contract carriage permits, including specified areas or routes and conditions on contracts entered outside a specified area. citeturn0search6turn0search19

---

# 30. Permit Validation Is Contextual

The question is not merely:

```text
Does Driver Have Permit?
```

It is:

```text
Does Driver Have the Required Authorization
for This Specific Operation?
```

---

# 31. Ride Type

Legal validation may depend on ride type.

Examples:

```text
Point-to-Point Ride
Scheduled Ride
Airport Ride
Inter-State Ride
Special Contract Ride
Corporate Ride
```

The actual supported ride types should be defined by business and regulatory requirements.

---

# 32. Vehicle Type

Vehicle category may affect eligibility.

Examples:

```text
Motor Cab
Taxi
Auto Rickshaw
Tourist Vehicle
Other Passenger Vehicle
```

The platform should use the regulatory classification applicable to the operating market.

---

# 33. Driver Eligibility

A driver may be:

```text
Online
Available
Geographically Nearby
```

and still be:

```text
Legally Ineligible
```

Therefore legal validation must precede assignment.

---

# 34. Vehicle Eligibility

Likewise:

```text
Driver Eligible
```

does not automatically mean:

```text
Vehicle Eligible
```

Both must be validated.

---

# 35. Ride Eligibility

A ride itself must be validated independently.

Conceptually:

```text
Ride
 ↓
Origin
 ↓
Destination
 ↓
Ride Type
 ↓
Regional Policy
 ↓
Ride Eligible?
```

---

# 36. Driver + Vehicle + Ride Matrix

The final assignment eligibility can be represented as:

```text
Ride
+
Driver
+
Vehicle
+
Region
+
Permit
+
Policy
```

All required conditions must pass.

---

# 37. Legal Validation Result

The validation service should produce a structured result.

Conceptually:

```text
status
reason_code
policy_version
evaluated_at
region_context
```

Possible statuses:

```text
ALLOWED
DENIED
REVIEW_REQUIRED
UNKNOWN
```

---

# 38. ALLOWED

Means:

```text
Known Applicable Rules
+
Required Data
+
Required Authorizations
```

indicate that the operation is permitted under the platform's configured policy.

It does not constitute a legal opinion.

---

# 39. DENIED

Means:

```text
A Known Rule or Missing Required Authorization
```

prevents the operation according to configured policy.

---

# 40. REVIEW_REQUIRED

Used when:

```text
The case is ambiguous
```

and automated decision-making should not guess.

---

# 41. UNKNOWN

Represents insufficient information to make an automated decision.

The default operational policy should generally avoid treating UNKNOWN as ALLOWED.

---

# 42. Fail-Safe Principle

For high-risk regulatory decisions:

```text
Unknown
→ Do Not Automatically Assign
```

The exact user-facing behaviour should be defined by product policy.

---

# 43. Reason Codes

Use stable reason codes rather than only free-form messages.

Examples:

```text
REGION_NOT_ALLOWED
PERMIT_REQUIRED
PERMIT_EXPIRED
PERMIT_OUTSIDE_SCOPE
VEHICLE_NOT_ELIGIBLE
DRIVER_NOT_ELIGIBLE
INTERSTATE_AUTHORIZATION_MISSING
POLICY_RESTRICTED
REGION_UNRESOLVED
VALIDATION_DATA_MISSING
```

---

# 44. Reason Codes Are Internal Contracts

Reason codes should be:

```text
Stable
Documented
Versioned
Auditable
```

Clients may receive a safe public representation rather than internal regulatory details.

---

# 45. Validation Policy Version

Every legal validation decision should identify the policy version.

Example:

```text
policy_version = regional-policy-v7
```

This is important because rules can change over time.

---

# 46. Effective Dates

Regulatory policies should support:

```text
effective_from
effective_until
```

where required.

This prevents historical decisions from being evaluated against today's rules.

---

# 47. Historical Reproducibility

For an old ride, the platform should be able to identify:

```text
Which policy was active
```

when the decision was made.

---

# 48. Policy Change

When a regulatory rule changes:

```text
New Policy Version
       ↓
Validation Engine
       ↓
New Decisions
```

Existing completed rides should not be retroactively changed unless a legal or operational requirement explicitly requires it.

---

# 49. Policy Deployment

Regulatory policy changes should follow controlled deployment:

```text
Draft
 ↓
Review
 ↓
Validation
 ↓
Effective Date
 ↓
Production
```

---

# 50. Legal Review

Changes that encode regulatory interpretation should be reviewed by the appropriate legal/compliance authority before production activation.

Engineering should not independently invent legal interpretations.

---

# 51. Legal Rules vs Configuration

Not every rule should become a simple boolean configuration.

For example:

```text
cross_region_allowed = true
```

may be insufficient if actual eligibility depends on:

```text
Permit Type
Vehicle Type
Region
Ride Type
Time
Authority
```

The configuration model must reflect the actual policy complexity.

---

# 52. Policy Engine

RideForge should use a dedicated regional/legal validation boundary.

Conceptually:

```text
LegalValidationService
        │
        ├── Region Rules
        ├── Permit Rules
        ├── Vehicle Rules
        ├── Ride-Type Rules
        └── Operational Restrictions
```

The exact service boundary may evolve with platform scale.

---

# 53. Initial Implementation

The initial implementation should remain simple.

Prefer:

```text
Explicit Policy
+
Deterministic Validation
+
Versioned Configuration
```

over introducing a generalized rules engine before it is necessary.

---

# 54. Avoid Over-Engineering

Do not initially introduce:

```text
Complex DSL
Generic Policy Programming Language
External Rules Engine
Automated Legal Reasoning AI
```

unless actual regulatory complexity requires it.

---

# 55. AI Is Not the Legal Authority

AI must never be responsible for final legal eligibility.

AI may assist with:

```text
Policy Documentation
Data Classification
Monitoring
Anomaly Detection
```

but final automated validation must remain deterministic and governed.

---

# 56. AI Dispatch Interaction

The correct ordering is:

```text
Legal Validation
      ↓
Eligible Candidate Set
      ↓
AI Ranking
```

not:

```text
AI Ranking
      ↓
Legal Validation
```

---

# 57. Stand Dispatch Interaction

Stand dispatch must also respect legal validation.

A driver at the front of a stand queue cannot be assigned if:

```text
Driver / Vehicle / Ride
```

is legally incompatible.

---

# 58. Smart Dispatch Interaction

Smart Dispatch must exclude legally ineligible drivers before ranking.

---

# 59. Candidate Filtering

The preferred candidate pipeline is:

```text
All Nearby Drivers
       ↓
Availability
       ↓
Vehicle Compatibility
       ↓
Regional / Legal Eligibility
       ↓
Candidate Set
       ↓
Smart / AI Ranking
```

---

# 60. Legal Validation and Driver Reservation

Legal validation should occur before reservation where practical.

If policy changes or information becomes stale, final assignment should still revalidate critical constraints.

---

# 61. Final Assignment Revalidation

Before assignment:

```text
Ride State
Driver State
Vehicle State
Regional Policy
Permit Status
```

should be revalidated where required.

---

# 62. Race Conditions

A driver can become legally or operationally ineligible between:

```text
Candidate Discovery
```

and:

```text
Assignment
```

The final assignment boundary must protect against this.

---

# 63. Permit Expiry During Dispatch

Example:

```text
10:00 — Driver Eligible
10:01 — Permit Expires
10:02 — Assignment Attempt
```

The final assignment should not use stale eligibility information.

---

# 64. Policy Change During Dispatch

If a legal policy changes while a ride is in-flight, the system must have an explicit policy-transition rule.

Do not silently assume that the old decision remains valid forever.

---

# 65. Safe Transition

A policy transition may require:

```text
Finish Existing Assignment
Stop New Assignment
Revalidate Pending Rides
```

depending on legal requirements.

---

# 66. Cross-Region Ride

A cross-region ride should follow:

```text
Pickup Region
+
Destination Region
+
Driver / Vehicle Authorization
+
Ride Type
+
Applicable Policy
```

before dispatch.

---

# 67. Same-Region Ride

Same-region rides should still pass the standard validation pipeline.

Do not assume:

```text
Same Region
→ Automatically Legal
```

because other requirements may still apply.

---

# 68. Inter-State Operation

The platform must explicitly model inter-state operation where relevant.

The Motor Vehicles Act contains provisions governing use of vehicles outside the region in which permits were granted, including Section 88 and related permit mechanisms. citeturn0search0turn0search11

The platform must therefore distinguish:

```text
Same-Region
```

from:

```text
Inter-State / Cross-Region
```

operations.

---

# 69. No Blanket Inter-State Rule

RideForge must not encode:

```text
Inter-State = Allowed
```

or:

```text
Inter-State = Denied
```

as a universal rule.

The applicable permit and regulatory context must determine the decision.

---

# 70. Kerala / Karnataka Border

The initial regional architecture should explicitly support border scenarios such as:

```text
Kerala → Karnataka
Karnataka → Kerala
Kerala → Kerala
Karnataka → Karnataka
```

without assuming that all four cases have identical regulatory treatment.

---

# 71. Border Ride Example

A ride:

```text
Pickup: Kerala
Destination: Karnataka
```

must trigger:

```text
Cross-Region Policy
```

before dispatch.

The route provider alone must not authorize the ride.

---

# 72. Reverse Border Ride

Similarly:

```text
Pickup: Karnataka
Destination: Kerala
```

must be evaluated independently.

Rules may be directional or asymmetric depending on the applicable regulatory framework.

---

# 73. Driver Origin vs Ride Origin

A driver currently located in Karnataka may still be associated with:

```text
Vehicle Registration
Permit
Operating Authorization
```

from another jurisdiction.

Therefore location alone cannot determine legal eligibility.

---

# 74. Border Pickup

A pickup near a state boundary should use authoritative region resolution.

Do not infer legal status from:

```text
Nearest City Name
Postal Code Alone
GPS Without Boundary Validation
```

---

# 75. Border Destination

The destination should similarly be resolved against authoritative administrative data.

---

# 76. Route Crossing Without Destination Crossing

A route may pass through another administrative area without the destination being there.

The system should not automatically apply a cross-region restriction unless the applicable legal policy requires route-level validation.

---

# 77. Route Provider Independence

A route provider may return:

```text
Kerala → Karnataka Route
```

even if RideForge policy denies the ride.

Therefore:

```text
Routing Success
≠
Legal Validation Success
```

---

# 78. Legal Validation Before Pricing

Where legal eligibility affects whether a ride can be offered at all:

```text
Legal Validation
```

should occur before the platform presents a confirmed bookable ride.

Pricing should not create an impression that an illegal operation is available.

---

# 79. Legal Validation Before Dispatch

Mandatory:

```text
Validation
 ↓
Dispatch
```

---

# 80. Legal Validation Before Driver Offer

Mandatory where driver eligibility depends on the ride's regional/legal context.

---

# 81. Legal Validation Before Assignment

Mandatory.

---

# 82. User-Facing Error

When a ride is rejected, the public message should be:

```text
Clear
Non-accusatory
Non-technical
```

Do not expose sensitive internal regulatory data unnecessarily.

---

# 83. Internal Reason

Internally retain:

```text
reason_code
policy_version
region_context
validation_timestamp
```

for operational investigation.

---

# 84. Regulatory Audit Trail

Important validation decisions should be auditable.

Record:

```text
ride_id
driver_id
vehicle_id
origin_region
destination_region
policy_version
decision
reason_code
evaluated_at
```

Only retain fields required by governance policy.

---

# 85. Audit Immutability

Critical regulatory decisions should not be silently overwritten.

Use an append-oriented audit model where practical.

---

# 86. Audit Retention

Retention should be defined according to:

```text
Legal Requirements
Business Requirements
Privacy Requirements
Operational Investigation
```

Do not retain sensitive data indefinitely by default.

---

# 87. Regulatory Evidence

Where a decision depends on an external artifact such as:

```text
Permit
Verification
Government Record
Compliance Review
```

the platform should retain an appropriate reference to that evidence where legally and operationally justified.

---

# 88. Document Storage

Do not store sensitive legal documents inside arbitrary application tables without a defined document-storage and access-control strategy.

---

# 89. Access Control

Only authorized roles should access:

```text
Permit Data
Compliance Records
Legal Validation Overrides
Regulatory Evidence
```

---

# 90. Manual Review

Some cases may require manual review.

Examples:

```text
Unknown Permit Status
Boundary Ambiguity
New Regulatory Rule
Incomplete Vehicle Data
Special Contract
```

---

# 91. Manual Review Workflow

Conceptually:

```text
Ride
 ↓
Validation
 ↓
REVIEW_REQUIRED
 ↓
Compliance / Operations
 ↓
Approve / Reject
```

---

# 92. Manual Approval

Manual approval should be:

```text
Authorized
Audited
Time-Bounded
Ride-Specific
Policy-Aware
```

---

# 93. Manual Override

A manual override must not bypass hard legal restrictions unless the override itself represents a formally authorized regulatory/business action.

The system should never provide an unrestricted:

```text
force_allow = true
```

mechanism.

---

# 94. Override Audit

Record:

```text
Operator
Reason
Time
Original Decision
New Decision
Policy Version
Reference
```

---

# 95. Operational Emergency Mode

An emergency operational mode must not silently disable legal validation.

Safety and legal constraints remain active unless a properly governed emergency policy explicitly changes the applicable rule.

---

# 96. Data Freshness

Regulatory data must have freshness metadata.

Examples:

```text
permit_verified_at
policy_effective_from
policy_updated_at
```

---

# 97. Stale Regulatory Data

If required regulatory data is too stale:

```text
Do Not Treat It as Current Without Policy Approval
```

Use:

```text
Reverification
Review
Safe Denial
```

as applicable.

---

# 98. Regulatory Data Synchronization

Where external sources are available, updates may be synchronized through:

```text
Scheduled Import
Manual Verification
Official API
Operational Workflow
```

The exact mechanism depends on authority availability.

---

# 99. External Authority Integration

An external government system may eventually provide authoritative permit status.

The architecture should support an adapter:

```text
RideForge Permit Service
       ↓
Authority Integration
```

without coupling the domain directly to the external API.

---

# 100. No Assumed Government API

Do not assume that a government authority provides:

```text
Real-Time API
```

unless verified.

The system must support manual/operational verification where automated integration is unavailable.

---

# 101. Regulatory Policy Repository

Regulatory rules should have a controlled source.

Conceptually:

```text
Regulatory Policy
    ↓
Versioned Policy Repository
    ↓
Validation Service
```

---

# 102. Policy Ownership

Each policy should have:

```text
Owner
Effective Date
Review Date
Source
Status
Version
```

---

# 103. Policy Source

The source should identify the basis of the rule, such as:

```text
Law
Government Notification
Permit Condition
Authority Guidance
Approved Legal Interpretation
```

---

# 104. Policy Review

Regulatory policies should be periodically reviewed.

Especially when:

```text
Law Changes
Permit Rules Change
Aggregator Rules Change
State Rules Change
Operating Region Changes
```

---

# 105. No Automatic Legal Inference

RideForge should not infer legal rules from:

```text
Historical Ride Success
Competitor Behaviour
Map Provider Routes
Driver Behaviour
User Reports
```

These may be signals for review but are not legal authority.

---

# 106. Regulatory Change Detection

Future tooling may detect changes in official sources.

However:

```text
Detected Change
≠
Automatically Activated Rule
```

Human/legal validation remains required for consequential policy changes.

---

# 107. Policy Testing

Every policy change should be tested against representative scenarios.

Examples:

```text
Allowed Same-Region Ride
Denied Cross-Region Ride
Permit Expired
Permit Valid
Unknown Permit
Border Pickup
Border Destination
Vehicle Mismatch
```

---

# 108. Regression Suite

Maintain a legal/regional regression suite.

This protects against accidental changes in validation behaviour.

---

# 109. Policy Simulation

Before activating a new policy, evaluate it against:

```text
Historical Rides
Synthetic Border Cases
Driver Permit Scenarios
Vehicle Categories
```

where data is available and privacy requirements permit.

---

# 110. Shadow Policy

A new regulatory policy implementation may be run in shadow mode:

```text
Current Policy
→ Real Decision

New Policy
→ Shadow Decision
```

Differences should be reviewed before activation.

---

# 111. Policy Rollout

A policy rollout may use:

```text
Draft
 ↓
Test
 ↓
Shadow
 ↓
Review
 ↓
Effective
```

---

# 112. Policy Rollback

Where legally appropriate, a faulty implementation should be reversible.

Historical decisions must retain the policy version used at the time.

---

# 113. Regulatory Incidents

An incident may occur if:

```text
Illegal Ride Offered
Illegal Driver Assigned
Expired Permit Accepted
Cross-Region Restriction Bypassed
Incorrect Region Resolved
```

Such incidents require:

```text
Immediate Containment
Audit
Root Cause Analysis
Policy Review
```

---

# 114. Incident Kill Switch

Operations should be able to disable a problematic ride type or regional flow quickly.

Examples:

```text
Disable Cross-Region Bookings
Disable Region Pair
Disable Vehicle Category
```

This must be audited and controlled.

---

# 115. Region Pair Policy

The system may represent regional policy using an explicit pair:

```text
origin_region
destination_region
```

Example:

```text
KL → KA
KA → KL
KL → KL
KA → KA
```

This allows directional policy.

---

# 116. Directional Rules

Do not assume:

```text
KL → KA
```

has identical policy to:

```text
KA → KL
```

The configuration model should support directional evaluation.

---

# 117. Region Pair Matrix

A conceptual matrix:

| Origin | Destination | Policy |
|---|---|---|
| Region A | Region A | Configured |
| Region A | Region B | Configured |
| Region B | Region A | Configured |
| Region B | Region B | Configured |

The values must come from approved policy, not engineering assumptions.

---

# 118. Vehicle-Specific Region Policy

The same region pair may have different rules for:

```text
Vehicle Category A
Vehicle Category B
```

where applicable.

---

# 119. Ride-Type-Specific Policy

Likewise:

```text
Point-to-Point
Scheduled
Special Contract
```

may have different requirements.

---

# 120. Driver-Specific Eligibility

Driver-specific attributes may affect eligibility:

```text
Permit
Vehicle
Registration
Authorization
Status
```

These must be evaluated through authoritative driver data.

---

# 121. Policy Evaluation Order

A practical validation order is:

```text
1. Resolve Origin
2. Resolve Destination
3. Resolve Applicable Policy
4. Validate Ride Type
5. Validate Vehicle Requirements
6. Validate Driver Authorization
7. Validate Permit Scope
8. Validate Cross-Region Rules
9. Produce Decision
```

The exact order may vary when a rule requires a different dependency.

---

# 122. Dispatch Policy Interaction

After legal validation:

```text
Allowed
 ↓
Dispatch Policy
 ↓
Stand / Smart / AI
```

The legal policy does not choose the dispatch algorithm unless explicitly configured to do so.

---

# 123. Legal Validation Does Not Optimize

The legal validator answers:

```text
Allowed?
```

It should not answer:

```text
Which Driver Is Best?
```

That remains dispatch responsibility.

---

# 124. Legal Validation Does Not Route

The legal validator may consume geographic information, but it should not become the route engine.

Routing remains governed by:

```text
ADR-0017
```

---

# 125. Legal Validation Does Not Own Driver Location

Driver location remains governed by:

```text
ADR-0010
```

---

# 126. Legal Validation Does Not Own Assignment

Assignment remains governed by the dispatch and ride domain.

---

# 127. Event-Driven Integration

Regional/legal policy changes may be propagated through events where appropriate.

Potential event:

```text
RegionalPolicyUpdated
```

Consumers may include:

```text
Dispatch
Driver Eligibility
Operations
Analytics
```

---

# 128. Policy Event Reliability

Important policy changes should use reliable event publication where asynchronous propagation is required.

The event architecture follows:

```text
ADR-0005
ADR-0006
ADR-0012
ADR-0013
```

---

# 129. Policy Cache

Legal policies may be cached for low-latency validation.

The cache must include:

```text
Policy Version
Effective Time
Expiry / Refresh
```

---

# 130. Policy Cache Failure

If the policy cache is unavailable:

```text
Use Authoritative Policy Store
```

where feasible.

Do not silently fall back to an outdated policy indefinitely.

---

# 131. Regulatory Validation Latency

Legal validation should be fast enough to participate in:

```text
Ride Creation
Dispatch
Assignment
```

without becoming a major latency bottleneck.

---

# 132. Precomputed Policy

Where policies are stable, precomputed region matrices may be used.

Example:

```text
KL → KA
→ RESTRICTED

KA → KL
→ RESTRICTED
```

The matrix must still be versioned and traceable to its source policy.

---

# 133. Driver Eligibility Cache

Driver permit eligibility may be cached, but the cache must have explicit freshness.

Critical final assignment checks should revalidate stale or high-risk data.

---

# 134. Regulatory Data and Redis

Redis may be used for:

```text
Cached Policy
Cached Eligibility
Fast Region Lookups
```

but Redis is not the legal source of truth.

---

# 135. Regulatory Data and PostgreSQL

PostgreSQL may hold:

```text
Policy Metadata
Permit Records
Audit References
Region Configuration
```

according to domain ownership.

---

# 136. Geospatial Data

PostGIS may support:

```text
Point-in-Polygon
Region Resolution
Zone Lookup
Boundary Detection
```

as defined by:

```text
ADR-0008
```

---

# 137. Route Provider Separation

Route providers may determine:

```text
Route
Distance
Travel Time
```

but not:

```text
Legal Authorization
```

---

# 138. Regulatory Validation and Pricing

If a ride is legally unavailable:

```text
Do Not Present a Normal Confirmable Fare
```

unless product policy explicitly supports a pre-validation state.

---

# 139. Regulatory Validation and Search

Search results for drivers should exclude drivers who are legally ineligible for the ride.

---

# 140. Regulatory Validation and Stand Queue

Stand queues may contain drivers who are generally available but not eligible for a specific ride.

The queue should not be mutated merely because a particular ride cannot use the driver unless the stand policy explicitly requires it.

---

# 141. Regulatory Validation and Driver Offers

A driver should not receive an offer for a ride they are not legally eligible to perform.

---

# 142. Regulatory Validation and Notifications

If a ride is rejected for regional/legal reasons, notifications should use a safe product-level message.

Internal regulatory reason codes should remain protected.

---

# 143. Security

Regulatory data is sensitive operational data.

Protect it with:

```text
Authentication
Authorization
Encryption
Audit Logging
Least Privilege
```

---

# 144. Sensitive Permit Data

Do not expose:

```text
Full Permit Documents
Sensitive Identifiers
Internal Compliance Notes
```

to ordinary users unless required.

---

# 145. Compliance Access

Compliance/admin users should have explicit access roles.

---

# 146. Logging Restrictions

Do not log full:

```text
Permit Documents
Sensitive Driver Documents
Personal Information
```

in application logs.

---

# 147. Error Handling

Regulatory validation failures should produce typed internal errors.

Examples:

```text
RegionNotAllowed
PermitInvalid
PermitExpired
PermitScopeMismatch
RegionResolutionFailed
ValidationDataUnavailable
```

---

# 148. Unknown State

When the validator cannot establish eligibility:

```text
UNKNOWN
```

should not silently become:

```text
ALLOWED
```

---

# 149. Dependency Failure

If a required regulatory dependency is unavailable:

```text
Fail Safe
```

according to risk classification.

For high-risk assignment decisions:

```text
Do Not Assign Without Required Validation
```

---

# 150. Observability

Track:

```text
Validation Requests
Allowed Count
Denied Count
Review Count
Unknown Count
Validation Latency
Reason Codes
Region Pair
Policy Version
```

---

# 151. Regulatory Metrics

Useful metrics include:

```text
Cross-Region Attempt Rate
Cross-Region Denial Rate
Permit Failure Rate
Unknown Validation Rate
Manual Review Rate
Policy Evaluation Latency
```

---

# 152. Anomaly Detection

Future monitoring may identify:

```text
Unexpected Increase in Cross-Region Attempts
Unexpected Denial Spike
Unexpected Allowed Spike
Region Resolution Errors
Permit Verification Failures
```

AI may assist with anomaly detection but must not silently change legal policy.

---

# 153. Alerting

Alerts may be configured for:

```text
Large Policy Behaviour Change
High UNKNOWN Rate
High Validation Failure Rate
Permit Verification Outage
Region Resolution Outage
Unexpected Cross-Region Activity
```

---

# 154. Testing Strategy

Regional/legal validation requires:

```text
Unit Tests
Policy Tests
Integration Tests
Geospatial Tests
Boundary Tests
Concurrency Tests
Regression Tests
```

---

# 155. Region Tests

Test:

```text
Inside Region
Outside Region
Boundary
Unknown Region
Invalid Coordinate
```

---

# 156. Cross-Region Tests

Test:

```text
Same Region
Region A → Region B
Region B → Region A
```

---

# 157. Permit Tests

Test:

```text
Valid Permit
Expired Permit
Suspended Permit
Revoked Permit
Wrong Vehicle
Wrong Region
Missing Permit
```

---

# 158. Policy Version Tests

Test:

```text
Policy v1
Policy v2
Effective Date
Expired Policy
Future Policy
```

---

# 159. Dispatch Integration Tests

Verify:

```text
Legal Allowed
→ Driver Candidate

Legal Denied
→ No Candidate

Legal Unknown
→ Safe Handling
```

---

# 160. AI Integration Tests

Verify:

```text
Legally Eligible Candidate
→ AI Ranking

Legally Ineligible Candidate
→ Never Sent to AI Ranking
```

---

# 161. Stand Integration Tests

Verify:

```text
Queue-Head Driver Ineligible
→ Correct Queue Behaviour
```

without violating queue policy.

---

# 162. Race Tests

Simulate:

```text
Permit Expires During Dispatch
Policy Changes During Dispatch
Driver Becomes Ineligible During Assignment
```

and verify safe final validation.

---

# 163. Historical Regression

Maintain representative historical scenarios for:

```text
Border Rides
Regional Restrictions
Permit Conditions
Policy Changes
```

where data use is legally permitted.

---

# 164. Policy Simulation

Before a policy becomes active:

```text
Run Historical / Synthetic Scenarios
Compare Old vs New
Review Differences
Approve
Activate
```

---

# 165. Performance

Legal validation should not require expensive route calculations for every ride unless the applicable rule actually depends on route-level information.

Prefer:

```text
Cheap Region Resolution
+
Policy Lookup
+
Eligibility Validation
```

before expensive routing.

---

# 166. Avoid Route-Based Legal Overengineering

Do not calculate a complete route merely to determine:

```text
State A → State B
```

if origin and destination boundaries already provide the necessary policy inputs.

Route-level validation should be introduced only when legally required.

---

# 167. Cost Optimization

The validation architecture should minimize:

```text
External API Calls
Database Queries
Geospatial Computation
Manual Review
```

while maintaining correctness.

---

# 168. Precomputed Region Mapping

Stable administrative boundaries can be preprocessed for efficient lookup.

---

# 169. Permit Eligibility Cache

Permit eligibility may be cached with:

```text
Short / Appropriate TTL
Version
Verification Timestamp
```

---

# 170. Fail-Safe vs Fail-Open

For critical legal constraints:

```text
Fail-Closed
```

is the default architectural principle.

That means:

```text
Required Validation Unavailable
        ↓
Do Not Automatically Permit Assignment
```

---

# 171. Exceptions

A fail-open behaviour may be used only where:

```text
Risk Is Low
Policy Explicitly Allows It
Legal Review Supports It
```

The exception must be documented.

---

# 172. Regulatory Data Ownership

The appropriate domain should own:

```text
Driver Permit Data
Vehicle Permit Data
Regional Policy
```

The validation layer consumes these sources rather than duplicating ownership.

---

# 173. Service Boundary

The validation capability may initially exist within the Ride/Dispatch domain if scale is small.

It may later become:

```text
Regional Policy Service
Compliance Service
Eligibility Service
```

if justified.

---

# 174. Do Not Prematurely Create a Compliance Microservice

The architecture should not create a standalone service solely because the concept is named "legal validation."

Start with a clear module/application boundary.

Extract a service when:

```text
Independent Scaling
Independent Ownership
Operational Complexity
Multiple Consumers
```

justify it.

---

# 175. Configuration Example

A conceptual policy record may contain:

```text
policy_id
version
origin_region
destination_region
ride_type
vehicle_type
required_authorization
status
effective_from
effective_until
source_reference
```

This is illustrative, not a mandatory schema.

---

# 176. Policy Source Reference

Every significant policy should reference its source.

Example:

```text
source_reference:
  authority: "Approved Regulatory Source"
  reference: "Document / Notification / Rule Identifier"
```

The exact source must be maintained by the compliance process.

---

# 177. Source Verification

Engineering should not activate a policy merely because a document exists.

The applicable source must be reviewed and approved through the project's legal/compliance workflow.

---

# 178. Legal Documentation Boundary

This ADR defines the **software architecture for enforcing legal/regional policy**.

It is not:

```text
Legal Advice
Permit Guidance
Legal Opinion
Regulatory Certification
```

---

# 179. Consequences

## 179.1 Positive Consequences

The decision provides:

```text
Explicit Legal Safety Boundary
Regional Flexibility
Permit-Aware Dispatch
Directional Region Policies
Auditable Decisions
Versioned Regulatory Policies
Safe AI Integration
Controlled Border Operations
```

---

## 179.2 Negative Consequences

The architecture introduces:

```text
Regulatory Data Management
Policy Versioning
Manual Review
Permit Verification
Additional Validation Latency
Compliance Operations
Additional Testing
```

These trade-offs are accepted.

---

# 180. Risks

## Risk 1 — Regulatory Policy Becomes Outdated

### Mitigation

```text
Effective Dates
Review Dates
Policy Ownership
Official Source References
Periodic Review
```

---

## Risk 2 — Engineering Misinterprets a Law

### Mitigation

```text
Legal / Compliance Review
Explicit Source References
No Unapproved Legal Inference
```

---

## Risk 3 — Stale Permit Data

### Mitigation

```text
Verification Timestamp
Expiry Monitoring
Reverification
Fail-Safe Assignment
```

---

## Risk 4 — Boundary Resolution Error

### Mitigation

```text
Authoritative Boundary Data
Border Handling
Geospatial Tests
Manual Review
```

---

## Risk 5 — AI Bypasses Legal Constraints

### Mitigation

```text
Legal Validation Before AI
Hard Constraints
Final Revalidation
```

---

## Risk 6 — Manual Override Abuse

### Mitigation

```text
Role-Based Access
Audit Trail
Reason Required
Time-Bounded Overrides
```

---

## Risk 7 — Excessive Complexity

### Mitigation

Start with:

```text
Deterministic Policy
Versioned Configuration
Explicit Validation
```

and introduce more advanced rule infrastructure only when required.

---

# 181. Validation

This ADR should be validated through:

```text
Policy Unit Tests
Regional Boundary Tests
Permit Tests
Cross-Region Tests
Dispatch Integration Tests
AI Integration Tests
Concurrency Tests
Policy Regression Tests
Policy Simulation
Manual Review Tests
Fail-Safe Tests
Audit Tests
```

---

# 182. Review Triggers

Revisit this ADR when:

```text
A New State / Market Is Added
A New Vehicle Category Is Added
Permit Rules Change
Aggregator Regulations Change
A New Inter-State Operation Is Introduced
Government Data Integration Becomes Available
A Dedicated Compliance Service Is Required
A Major Regulatory Incident Occurs
AI Is Proposed for Legal Validation
```

---

# 183. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
```

Especially:

```text
Regional / Legal Ride Validation
Smart Dispatch AI
ETA and Prediction System
AI Matching and Ranking
AI Safety and Guardrails
AI Failure and Fallback Strategy
AI Data Privacy and Governance
Database Development
PostGIS / Geospatial Development
Redis Development
API Development
Error Handling and Validation
Observability Development
Configuration and Environment
Testing and Integration
```

---

# 184. Related ADRs

This decision is directly related to:

```text
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
ADR-0008 — PostGIS for Geospatial Operations
ADR-0009 — Redis for Real-Time State and Caching
ADR-0010 — Driver Location Storage Strategy
ADR-0014 — API and Service Communication
ADR-0015 — Smart Dispatch and Stand Dispatch
ADR-0016 — AI-Assisted Dispatch Strategy
ADR-0017 — ETA and Route Provider Strategy
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
ADR-0023 — Security and Secret Management
ADR-0024 — Configuration and Environment Strategy
ADR-0025 — Testing and Integration Strategy
ADR-0026 — Model and AI Governance
ADR-0028 — Cost Optimization Strategy
```

---

# 185. Decision Summary

RideForge adopts a deterministic regional and legal validation boundary:

```text
                         Ride Request
                              │
                              ▼
                     Geographic Resolution
                              │
                              ▼
                   Regional Policy Resolution
                              │
                              ▼
                     Ride Legal Validation
                              │
                    ┌─────────┼─────────┐
                    ▼         ▼         ▼
                 ALLOWED    DENIED    REVIEW
                    │
                    ▼
             Driver / Vehicle
                Eligibility
                    │
                    ▼
             Dispatch Strategy
              /      |       \
           Stand    Smart      AI
                    │
                    ▼
             Final Revalidation
                    │
                    ▼
                 Assignment
```

The core rules are:

```text
Route Availability
        ≠
Legal Authorization

Driver Availability
        ≠
Legal Eligibility

AI Ranking
        ≠
Legal Decision

Permit Exists
        ≠
Permit Automatically Covers Every Ride
```

---

# 186. Final Principle

> **RideForge must treat regional and legal eligibility as a deterministic, versioned, auditable hard constraint that is evaluated before dispatch and assignment, while keeping regulatory interpretation outside the AI and routing systems.**

The validation hierarchy is:

```text
1. Resolve Geographic Context
2. Resolve Applicable Policy
3. Validate Ride
4. Validate Driver
5. Validate Vehicle
6. Validate Permit / Authorization
7. Produce Deterministic Decision
8. Dispatch Only If Allowed
9. Revalidate Before Assignment
10. Record the Decision and Policy Version
```

For uncertain cases:

```text
UNKNOWN
     ↓
Do Not Automatically Assign
```

For prohibited cases:

```text
DENIED
     ↓
Do Not Dispatch
```

For permitted cases:

```text
ALLOWED
     ↓
Dispatch
```

This architecture allows RideForge to support different operating regions and regulatory environments without embedding fragile, region-specific legal assumptions throughout the ride, matching, dispatch, routing, or AI systems.

---

# 187. Status

```text
Decision: ACCEPTED

Regional Validation:
Required

Legal Validation:
Required

Cross-Region Validation:
Explicit

Permit Validation:
Required Where Applicable

AI Legal Authority:
Not Allowed

Routing Legal Authority:
Not Allowed

Dispatch Before Legal Validation:
Not Allowed

Final Assignment Revalidation:
Required

Policy Versioning:
Required

Auditability:
Required

Unknown State:
Fail-Safe

Manual Review:
Supported

Manual Override:
Restricted + Audited

Primary Goal:
Prevent Illegal or Unauthorized Ride Assignment While Supporting Regional Operational Flexibility
```

---

# 24. Clarification: Regional/Legal Validation and the Dispatch Strategy Model

Regional and legal validation is a **hard eligibility constraint** within the dispatch flow. It does not determine which dispatch strategy is configured for a ride.

The responsibilities must remain separate:

```text
Hierarchical Configuration
        ↓
Resolve Effective Dispatch Strategy
        ↓
Candidate Discovery
        ↓
Regional / Legal Validation
        ↓
Strategy-Specific Prioritization / Ranking
        ↓
Assignment
```

A geographically nearby driver is not automatically legally eligible, and a legally eligible driver is not automatically the best dispatch candidate.

---

## 24.1 Two Primary Dispatch Strategies

RideForge has two primary dispatch strategies:

```text
Smart Stand Dispatch
Smart Dispatch
```

AI-assisted dispatch is an optimization capability and is not a third primary dispatch strategy.

Regional and legal validation applies regardless of which primary strategy is resolved.

---

## 24.2 Smart Stand Dispatch and Legal Validation

Smart Stand Dispatch is **stand-preferred, not stand-exclusive**.

When the rider is inside a configured stand radius:

```text
Preferred Stand
    ↓
Eligible Stand Drivers
    ↓
Stand Queue / Ordering
```

If suitable stand supply is unavailable, broader candidates may be considered:

```text
Non-Stand Drivers
Nearby Stand Drivers
Drivers from Nearby Locations
```

However, every expanded candidate must still pass the applicable regional and legal validation.

Therefore:

```text
Stand Preference
    ≠
Legal Eligibility
```

and:

```text
Inside Stand Radius
    ≠
Automatically Legally Eligible
```

---

## 24.3 Smart Dispatch and Legal Validation

Smart Dispatch is stand-agnostic.

It may consider any geographically suitable driver who satisfies the applicable eligibility and legal constraints.

Stand membership must not override legal restrictions.

```text
Smart Dispatch
    ↓
Nearby Eligible Drivers
    ↓
Regional / Legal Validation
    ↓
Valid Candidates
```

---

## 24.4 Cross-Location Dispatch and Legal Boundaries

Cross-location candidate discovery does not bypass regional or legal restrictions.

Example:

```text
Location A
    ↓
No suitable local driver
    ↓
Expand to Location B
    ↓
Discover nearby candidates
    ↓
Regional / Legal Validation
    ↓
Only legally eligible candidates continue
```

Location B may use a different dispatch strategy from Location A.

For example:

```text
Location A → Smart Dispatch
Location B → Smart Stand Dispatch
```

A driver from Location B may still be considered for a ride originating in Location A if:

1. The candidate is geographically suitable.
2. The candidate is operationally eligible.
3. The applicable regional/legal rules permit the ride.
4. The candidate satisfies the remaining dispatch constraints.

A different source-location strategy does not itself make a candidate illegal or eligible.

---

## 24.5 Dispatch Strategy Is Not a Legal Boundary

The configured dispatch strategy must not be interpreted as a legal or geographic boundary.

These concepts remain separate:

```text
Dispatch Strategy
Configuration Hierarchy
Geographic Candidate Discovery
Regional / Legal Eligibility
Strategy-Specific Prioritization
```

Therefore:

```text
Smart Stand Dispatch
    ≠
Stand-only legal boundary
```

and:

```text
Location Strategy
    ≠
Legal authorization to cross regions
```

---

## 24.6 Hierarchical Dispatch Configuration

The effective dispatch strategy is resolved through hierarchical configuration:

```text
Most Specific Applicable Configuration
        ↓
Explicit Strategy?
   ├── YES → Effective Strategy
   └── NO
        ↓
Parent Configuration
        ↓
Continue Upward
        ↓
System Default
```

Possible configuration levels include:

```text
State
District
City / Town
Rural Area
Auto Stand
Specific Ride Level
Other Intermediate Levels
```

The nearest explicit configuration takes precedence.

Regional/legal validation must not alter this inheritance rule.

Instead, the resolved strategy is executed while legal validation remains a hard constraint.

---

## 24.7 Legal Validation During Candidate Expansion

When candidate discovery expands, each newly discovered candidate must independently pass the applicable legal checks.

Example:

```text
Preferred Stand Candidate
        ↓
Legal Validation
        ↓
Rejected
        ↓
Expand Candidate Search
        ↓
Nearby Candidate
        ↓
Legal Validation
        ↓
Candidate Accepted / Rejected
```

The system must not assume that because one candidate from a location is legally eligible, every other candidate from that location is automatically eligible.

Legal eligibility is evaluated according to the applicable ride, driver, vehicle, service, and regional constraints.

---

## 24.8 Regional Rules Must Be Applied Before Final Assignment

Regional/legal validation should be treated as a hard constraint before final assignment.

Conceptually:

```text
Candidate Discovery
        ↓
Hard Eligibility
        ↓
Regional / Legal Validation
        ↓
Strategy-Specific Prioritization
        ↓
Ranking
        ↓
Assignment
```

AI ranking must operate only on candidates that have passed the applicable hard constraints.

AI must never be used to decide that a legally invalid candidate is acceptable because the candidate has a better ETA, distance, or predicted score.

---

## 24.9 Cross-Region Does Not Mean Cross-Legal-Boundary

A nearby location can participate in candidate discovery without implying that the ride is legally permitted across the relevant regional boundary.

Therefore:

```text
Geographic Proximity
    ≠
Legal Permission
```

and:

```text
Candidate Discovery
    ≠
Legal Authorization
```

The discovery layer may identify a nearby candidate; the legal validation layer determines whether that candidate can actually participate in the ride.

---

## 24.10 AI-Assisted Dispatch and Legal Constraints

AI-assisted ranking may use regional/geographic signals, but it must operate below legal validation.

```text
Legal Validation
        ↓
Allowed Candidate Set
        ↓
AI-Assisted Ranking
```

AI cannot:

```text
Override regional restrictions
Authorize an otherwise prohibited ride
Bypass legal validation
Treat geographic proximity as legal permission
```

If AI is unavailable, deterministic dispatch continues using the same legally valid candidate set.

---

## 24.11 Failure and Fallback

Regional/legal validation remains authoritative during all fallback paths.

For Smart Stand Dispatch:

```text
Smart Stand Dispatch
        ↓
Preferred Stand
        ↓
No suitable legally eligible candidate
        ↓
Broader Candidate Discovery
        ↓
Regional / Legal Validation
        ↓
Continue with valid candidates
```

For Smart Dispatch:

```text
Smart Dispatch
        ↓
Nearby Candidate Discovery
        ↓
Regional / Legal Validation
        ↓
Deterministic / AI-assisted Ranking
```

Failure or expansion must not bypass legal validation.

A fallback strategy may only change the dispatch strategy if an explicit business/configuration rule permits such a transition.

---

## 24.12 Implementation Guardrails

Implementations must not:

```text
Treat dispatch strategy as a legal boundary.

Assume every driver in a nearby location is legally eligible.

Allow cross-location expansion to bypass regional validation.

Treat stand membership as proof of legal eligibility.

Allow AI ranking to override legal restrictions.

Treat geographic proximity as legal authorization.

Skip legal validation for fallback candidates.

Change the resolved dispatch strategy merely because a legally eligible driver was unavailable.

Hard-code legal behavior into the Smart Stand Dispatch implementation when the applicable regional rules belong to the legal validation domain.
```

Regional and legal validation must remain a hard constraint shared by both primary dispatch strategies and all candidate-expansion/fallback paths.

