# ADR-0015: Smart Dispatch and Stand Dispatch

> **Status:** Accepted  
> **Date:** 2026-08-09  
> **Decision Type:** Dispatch Architecture / Ride Assignment  
> **Scope:** Driver assignment strategies used by the RideForge dispatch platform  
> **Owner:** RideForge Architecture / Engineering  
> **Supersedes:** None  
> **Superseded By:** None

---

# 1. Context

RideForge is being designed as a ride-hailing platform that must support different operational environments.

A single driver-dispatch strategy is not appropriate for every operating area.

Some locations may require:

```text
Queue-Based / Stand-Based Dispatch
```

while other locations may benefit from:

```text
Dynamic / Smart Dispatch
```

The platform therefore needs to support both models without creating two unrelated dispatch systems.

The selected architecture must allow dispatch behaviour to vary by:

```text
Operating Region
Market
Location
Stand
Vehicle Category
Ride Type
Business Policy
Operational Mode
```

The system must also support future AI-assisted dispatch without making AI a mandatory dependency for every ride.

---

# 2. Problem

RideForge must answer:

```text
How should a driver be selected for a ride?

Should every ride use the same algorithm?

How should stand-based operations work?

How should dynamic proximity-based dispatch work?

How should AI-assisted ranking be introduced?

How should the platform switch between strategies?

What happens when the preferred dispatch mechanism is unavailable?
```

The architecture must preserve:

```text
Fairness
Operational Control
Predictability
Driver Experience
Rider Experience
Legal / Regional Constraints
Low Latency
Failure Recovery
Future AI Compatibility
```

---

# 3. Decision


RideForge will support **two primary dispatch strategies** within one common dispatch platform:

```text
Smart Stand Dispatch
Smart Dispatch
```

AI-assisted dispatch is an **optimization capability applied to a resolved strategy**, not a third primary dispatch strategy.

Conceptually:

```text
Ride Request
     │
     ▼
Dispatch Configuration Resolution
     │
     ▼
Effective Dispatch Strategy
     │
     ├───────────────┐
     ▼               ▼
Smart Stand      Smart Dispatch
Dispatch         │
     │           │
     └──────┬────┘
            ▼
Candidate Discovery / Processing
            │
            ▼
Strategy-Specific Prioritization
            │
            ▼
AI-Assisted Optimization (where applicable)
            │
            ▼
Assignment
```

The important architectural decision is:

> **Smart Stand Dispatch and Smart Dispatch are first-class strategies within one dispatch platform. Smart Stand Dispatch is stand-preferred, not stand-exclusive. AI is an optional optimization layer rather than a separate mandatory dispatch strategy.**

# 4. Core Principle

RideForge will not hard-code one global dispatch algorithm.

Instead:

```text
Dispatch Policy
      ↓
Select Strategy
      ↓
Execute Strategy
      ↓
Produce Candidate / Assignment Decision
```

This allows the platform to operate differently in different locations while preserving one common dispatch domain.

---

# 5. Dispatch Strategy Abstraction


The dispatch domain should expose a strategy-oriented abstraction for the two primary strategies:

```text
DispatchStrategy

    ├── SmartStandDispatchStrategy
    └── SmartDispatchStrategy
```

AI assistance should be modeled as a ranking/optimization capability that can be used by either strategy where applicable, rather than as a third fundamental strategy.

The exact Go interface and package structure are implementation details, but the architectural boundary should remain explicit.

# 6. Strategy Selection


A dispatch configuration hierarchy determines the effective strategy for a ride.

The system starts at the **most specific applicable configuration level** and moves upward until it finds an explicitly configured dispatch strategy.

Conceptually:

```text
Most Specific Applicable Level
        ↓
Explicit Strategy?
   ├── YES → Effective Strategy
   └── NO
        ↓
Parent Level
        ↓
Continue Upward
        ↓
System Default
```

Supported levels may include:

```text
State
District
City / Town
Rural Area
Auto Stand
Specific Ride Level
Other configured intermediate levels
```

The hierarchy must not be hard-coded to a fixed set of levels.

The canonical rule is:

> **Specific configuration overrides inherited configuration.**

Example:

```text
District → Smart Dispatch
Town A   → Smart Stand Dispatch
Town B   → No explicit configuration
```

Therefore:

```text
Town A ride → Smart Stand Dispatch
Town B ride → Smart Dispatch
```

Configuration resolution determines **which strategy applies**. It does not determine the complete driver candidate boundary.

# 7. Region-Level Configuration

A region may define its default dispatch strategy.

Example:

```text
Region A
→ STAND

Region B
→ SMART
```

This allows different operating markets to use different operational models.

---

# 8. Location-Level Configuration


Dispatch configuration may be defined at any supported level in the configuration hierarchy.

For example:

```text
State
   ↓
District
   ↓
City / Town
   ↓
Rural Area
   ↓
Auto Stand
   ↓
Ride-level configuration
```

Not every level requires explicit configuration.

A lower-level ride inherits the nearest explicitly configured strategy when no explicit strategy exists at that level.

This means:

```text
Specific configuration
        >
Inherited configuration
        >
System default
```

The configuration hierarchy determines the **effective dispatch strategy**. It must not be interpreted as a hard boundary that prevents eligible drivers from nearby locations from being considered.

# 9. Configuration Precedence


The canonical precedence model is:

```text
Most Specific Applicable Configuration
        ↓
Parent Configuration
        ↓
...
        ↓
System Default
```

The first explicit dispatch strategy encountered wins.

For example:

```text
District      → SMART
Town A        → STAND
Town B        → [unset]
```

results in:

```text
Town A → Smart Stand Dispatch
Town B → Smart Dispatch
```

Configuration inheritance is evaluated at dispatch-strategy resolution time. The dispatch engine must not require every hierarchy level to have an explicit configuration.

# 10. Stand Dispatch


**Smart Stand Dispatch** is a stand-preferred dispatch strategy.

When a rider is within the configured radius of an auto stand:

```text
Ride
 ↓
Applicable Stand
 ↓
Prefer eligible drivers at that stand
 ↓
Apply stand queue / ordering rules
```

The stand is a **preferred dispatch source, not an exclusive candidate pool**.

If the preferred stand cannot provide a suitable eligible driver, the dispatch process may consider broader eligible candidates, including:

```text
Drivers outside the preferred stand
Drivers at nearby stands
Drivers from nearby locations
```

If the rider is outside the radius of all configured stands, Smart Stand Dispatch must not restrict the candidate search to stand drivers. It may consider all eligible nearby drivers.

The strategy therefore means:

```text
Stand Preference
+
Broader Eligible Fallback
```

not:

```text
Stand Only
```

# 11. Why Stand Dispatch Exists

Stand dispatch can be appropriate when:

```text
Physical Driver Queues Exist
Local Transport Rules Require Ordered Allocation
Drivers Operate From Designated Stands
Fair Queue Rotation Is Important
Operational Staff Need Predictable Allocation
```

It provides a highly understandable operational model.

---

# 12. Stand Queue

A stand may maintain a logical queue:

```text
Position 1 → Driver A
Position 2 → Driver B
Position 3 → Driver C
Position 4 → Driver D
```

When a compatible ride arrives:

```text
Driver A
→ Candidate

If eligible:
→ Assign

Queue advances
```

---

# 13. Stand Eligibility

Being at the front of a queue does not automatically guarantee assignment.

The driver must satisfy:

```text
Available
Eligible
Correct Vehicle Category
Operationally Valid
Not Already Assigned
Within Required Stand Rules
```

---

# 14. Stand Queue Fairness

The stand strategy should preserve predictable fairness.

The queue should not be bypassed arbitrarily by:

```text
Distance
AI Score
Driver Preference
Random Selection
```

unless explicit stand policy permits it.

---

# 15. Queue Advancement

After a driver receives a ride, the queue position should advance according to stand policy.

For example:

```text
Before:
A → B → C → D

A Assigned

After:
B → C → D
```

The exact re-entry rule depends on operational policy.

---

# 16. Driver Re-Entry

A driver may re-enter the stand queue after completing a ride or returning to the stand.

The platform should define:

```text
Re-entry Conditions
Re-entry Position
Cooldown
Eligibility
```

to prevent unfair queue manipulation.

---

# 17. Stand No-Show / Decline

If the driver at the front is not eligible or cannot accept the ride, the system should follow a deterministic policy.

Possible actions:

```text
Skip
Temporary Hold
Move Driver
Mark Unavailable
Requeue
```

The behaviour must be explicit rather than dependent on ad-hoc application logic.

---

# 18. Stand Dispatch Advantages

Stand dispatch provides:

```text
Predictability
Transparency
Operational Simplicity
Queue Fairness
Low Computational Cost
Easy Human Understanding
```

---

# 19. Stand Dispatch Limitations

Stand dispatch may perform poorly when:

```text
Drivers Are Highly Distributed
Pickup Demand Is Spatially Dynamic
Travel Time Differences Are Large
Driver Supply Is Sparse
Demand Changes Rapidly
```

This is where Smart Dispatch can provide better allocation.

---

# 20. Smart Dispatch


**Smart Dispatch** is the stand-agnostic dispatch strategy.

When Smart Dispatch is the effective strategy, candidate discovery and ranking do not depend on whether a driver is currently at an auto stand.

Eligible candidates may include:

```text
Drivers at auto stands
Drivers outside auto stands
Drivers from nearby locations
```

subject to hard eligibility, legal, service, availability, and geographic constraints.

The core flow remains:

```text
Ride Request
      ↓
Candidate Discovery
      ↓
Eligibility Filtering
      ↓
Candidate Ranking
      ↓
Best Eligible Candidate
      ↓
Assignment
```

Smart Dispatch should optimize the overall dispatch decision rather than simply selecting the first or nearest driver.

# 21. Smart Dispatch Objective

Smart Dispatch attempts to optimize the overall dispatch decision rather than simply selecting:

```text
First Driver
```

or:

```text
Nearest Driver
```

Potential factors include:

```text
ETA
Distance
Travel Time
Driver Availability
Vehicle Type
Ride Type
Driver State
Demand
Supply
Operational Rules
Fairness
Historical Performance
```

---

# 22. Candidate Discovery

Smart Dispatch should first identify a bounded candidate set.

Conceptually:

```text
All Drivers
     ↓
Spatial Filtering
     ↓
Availability Filtering
     ↓
Vehicle Filtering
     ↓
Operational Filtering
     ↓
Candidate Set
```

The ranking system should not evaluate every driver in the entire platform.

---

# 23. Candidate Eligibility

Eligibility must be applied before ranking.

Examples:

```text
Driver Online
Driver Available
Vehicle Compatible
Region Allowed
Driver Not Already Assigned
Required Documents Valid
Operational Restrictions Satisfied
```

An ineligible driver must not win because of a high score.

---

# 24. Ranking

After eligibility filtering:

```text
Eligible Candidates
       ↓
Ranking
       ↓
Best Candidate
```

The ranking mechanism may initially be deterministic and later become AI-assisted.

---

# 25. Initial Smart Ranking

The initial implementation should prefer a simple deterministic ranking model.

Potential ordering:

```text
Estimated Pickup ETA
+
Distance
+
Eligibility
+
Operational Priority
```

The exact ranking formula belongs to the dispatch implementation and should remain configurable.

---

# 26. AI-Assisted Ranking

AI may later improve candidate ranking using learned features.

Conceptually:

```text
Candidate Features
       ↓
AI / ML Model
       ↓
Candidate Score
       ↓
Dispatch Policy
       ↓
Assignment
```

AI should remain an optional strategy or scoring component rather than an unavoidable dependency.

---

# 27. AI Must Not Replace Core Safety Rules

AI ranking must never override hard constraints such as:

```text
Driver Availability
Vehicle Eligibility
Regional Restrictions
Legal Constraints
Ride Type Compatibility
Driver Assignment State
```

The architecture is:

```text
Hard Constraints
      ↓
AI / Ranking
      ↓
Decision
```

not:

```text
AI
 ↓
Everything
```

---

# 28. Hybrid Dispatch Model

The platform therefore supports:

```text
Stand Dispatch
```

for environments where queue-based allocation is required, and:

```text
Smart Dispatch
```

for environments where dynamic optimization is preferred.

AI can enhance Smart Dispatch where appropriate.

---

# 29. Example Operating Modes

A market may be configured as:

```text
Market A
→ Stand Dispatch

Market B
→ Smart Dispatch

Market C
→ Smart Dispatch + AI Ranking
```

This allows operational flexibility without maintaining separate dispatch platforms.

---

# 30. Dispatch Policy

A dispatch policy should determine:

```text
Dispatch Mode
Candidate Rules
Search Radius
Vehicle Rules
Retry Rules
Assignment Timeout
Fallback Strategy
```

The policy may vary by region and operating model.

---

# 31. Policy Is Not Algorithm

The system should distinguish:

```text
Policy
```

from:

```text
Algorithm
```

Policy determines:

```text
What Is Allowed
```

Algorithm determines:

```text
How Candidates Are Selected
```

---

# 32. Example

Policy:

```text
Region X
→ Smart Dispatch
→ Search up to 5 km
→ Sedan required
```

Algorithm:

```text
Rank eligible drivers by estimated pickup ETA
```

Keeping these concerns separate makes the system easier to configure and evolve.

---

# 33. Dispatch Request

A conceptual dispatch request may contain:

```text
ride_id
pickup_location
destination_location
vehicle_type
ride_type
region
zone
dispatch_policy
requested_at
```

The exact domain model is implementation-specific.

---

# 34. Dispatch Result

A dispatch decision may contain:

```text
ride_id
driver_id
strategy
score / ranking metadata
decision_time
policy_version
```

Sensitive internal scoring details should not necessarily be exposed to external clients.

---

# 35. Strategy Metadata


The system should record the effective dispatch strategy that produced an assignment.

Example:

```text
strategy = SMART_STAND
```

or:

```text
strategy = SMART
```

AI involvement should be recorded separately rather than represented as a third primary strategy.

For example:

```text
strategy = SMART
ai_assisted = true
model_version = dispatch-ranker-v4
```

or:

```text
strategy = SMART_STAND
ai_assisted = false
```

This distinction is important for analytics, experimentation, debugging, fairness analysis, and performance comparison.

# 36. Strategy Version


Where dispatch strategies evolve, record the effective strategy and relevant policy/configuration version.

Example:

```text
strategy = SMART
strategy_version = 2
policy_version = dispatch-policy-v5
```

For AI-assisted execution:

```text
strategy = SMART
ai_assisted = true
model_version = dispatch-ranker-v4
```

For Smart Stand Dispatch:

```text
strategy = SMART_STAND
strategy_version = 1
```

This supports reproducibility without treating AI assistance as a separate primary dispatch strategy.

# 37. Assignment Atomicity

Selecting a driver and marking the driver assigned must be protected against races.

Avoid:

```text
Check Driver Available
      ↓
Wait
      ↓
Assign
```

without transactional or concurrency protection.

---

# 38. Race Condition Example

Two rides may attempt:

```text
Ride A → Driver X
Ride B → Driver X
```

at the same time.

Only one assignment should succeed.

---

# 39. Assignment Protection

The assignment boundary should use appropriate concurrency control such as:

```text
Database Transaction
Optimistic Concurrency
Row Lock
Atomic State Transition
```

according to the implementation.

---

# 40. Driver State

Driver availability should have explicit states.

Conceptually:

```text
OFFLINE
AVAILABLE
RESERVED
ASSIGNED
ON_TRIP
UNAVAILABLE
```

The exact state machine is defined by the driver/dispatch domain.

---

# 41. Reservation Before Assignment

Smart Dispatch may temporarily reserve a candidate while the assignment is being confirmed.

Conceptually:

```text
AVAILABLE
    ↓
RESERVED
    ↓
ASSIGNED
```

If assignment fails:

```text
RESERVED
    ↓
AVAILABLE
```

---

# 42. Reservation Timeout

A reservation must not remain indefinitely.

Use:

```text
Reservation Expiry
```

to recover from:

```text
Process Crash
Network Failure
Assignment Timeout
```

---

# 43. Stand Dispatch Concurrency

Stand queue operations also require concurrency protection.

Two ride requests must not both assign the same front-of-queue driver.

The queue operation should be atomic.

---

# 44. Queue Reservation

A stand dispatcher may use:

```text
Queue Head
   ↓
Reserve Driver
   ↓
Confirm Assignment
```

to prevent concurrent assignment.

---

# 45. Assignment Confirmation

Where driver confirmation is required, the assignment workflow should define:

```text
Offer Sent
Offer Timeout
Accepted
Rejected
Expired
Cancelled
```

The exact lifecycle is domain-specific.

---

# 46. Driver Decline

A driver decline should not corrupt the queue or candidate pool.

The system should apply the configured policy.

For Smart Dispatch:

```text
Candidate Rejected
   ↓
Next Candidate
```

may be used.

For Stand Dispatch:

```text
Queue Policy
```

determines the next action.

---

# 47. Dispatch Retry

A failed assignment may trigger another dispatch attempt.

However, retries must respect:

```text
Ride State
Assignment Attempts
Driver Eligibility
Dispatch Timeout
Business Policy
```

---

# 48. Dispatch Timeout

Every dispatch attempt should have a bounded time window.

Conceptually:

```text
Dispatch Start
      ↓
Candidate Selection
      ↓
Assignment Offer
      ↓
Timeout
      ↓
Next Action
```

---

# 49. No Candidate


If no suitable candidate exists in the current search scope, the system follows the configured discovery/retry policy.

Possible actions include:

```text
Expand Search
Wait
Retry
Continue With Broader Eligible Candidates
Notify User
Escalate
```

For Smart Stand Dispatch, failure of the preferred stand does **not** automatically mean that the strategy has changed to Smart Dispatch. The system may broaden candidate discovery while preserving the effective Smart Stand Dispatch strategy.

A strategy change may occur only when an explicit business/configuration rule defines that transition.

# 50. Search Expansion


Search expansion is a **candidate discovery mechanism**, not a dispatch-strategy switch.

Smart Dispatch may progressively expand its geographic search according to policy.

Smart Stand Dispatch may expand beyond a preferred stand when the stand cannot provide a suitable candidate.

Conceptually:

```text
Preferred / Local Search
        ↓
Insufficient Suitable Supply
        ↓
Broader Eligible Search
        ↓
Nearby Stands / Non-Stand Drivers / Nearby Locations
```

Expansion must be bounded and policy-driven.

The effective dispatch strategy remains unchanged unless an explicit configuration defines a strategy transition.

# 51. Strategy Fallback


Fallback must distinguish between **candidate expansion**, **AI degradation**, and **strategy switching**.

For Smart Stand Dispatch:

```text
Preferred Stand
    ↓
No Suitable Candidate
    ↓
Broader Eligible Candidate Discovery
    ↓
Continue Under Smart Stand Dispatch
```

For AI-assisted ranking:

```text
AI Ranking
    ↓
Unavailable / Timeout
    ↓
Deterministic Strategy-Compatible Ranking
```

A fallback must not silently convert Smart Stand Dispatch into Smart Dispatch.

A strategy switch is allowed only when explicitly defined by business/configuration policy.

# 52. AI Failure Fallback


AI must never become a single point of failure for dispatch.

Preferred flow:

```text
AI Ranking / Prediction
        ↓
Failure / Timeout
        ↓
Deterministic Strategy-Compatible Processing
        ↓
Assignment
```

For Smart Stand Dispatch, deterministic fallback must preserve stand preference and configured queue semantics.

For Smart Dispatch, deterministic fallback must preserve stand-agnostic ranking.

AI failure must not bypass:

```text
Hard Eligibility
Legal / Regional Constraints
Safety Constraints
Vehicle / Service Compatibility
Driver Availability
```

# 53. Smart Dispatch Failure


If Smart Dispatch cannot operate because of a dependency or internal failure, the system follows the configured degradation policy.

A fallback to another dispatch strategy is permitted **only when explicitly configured and legally/operationally valid**.

The absence of a Smart Dispatch dependency does not itself authorize a strategy change.

# 54. Stand Dispatch Failure


Smart Stand Dispatch has its own recovery path.

If the preferred stand cannot provide a suitable candidate, the system may expand to broader eligible candidates:

```text
Preferred Stand
    ↓
No Suitable Candidate
    ↓
Non-Stand / Nearby Stand / Nearby-Location Candidates
```

This is still Smart Stand Dispatch unless an explicit configuration defines another strategy.

Operational or technical stand failure may additionally follow:

```text
Retry Queue Operation
Rebuild Queue State
Use Configured Alternate Dispatch
Escalate Operationally
```

A silent strategy change is prohibited.

# 55. Regional and Legal Constraints

Dispatch strategy selection must respect regional and legal policy.

The dispatch engine must not assign a driver merely because the driver ranks highly.

It must first verify:

```text
Region Eligibility
Operational Permission
Ride Legality
Vehicle Requirements
```

---

# 56. Legal Validation Boundary

Legal/region validation should remain a hard eligibility constraint.

Conceptually:

```text
Ride
 ↓
Region Validation
 ↓
Eligible Drivers
 ↓
Dispatch Strategy
```

not:

```text
Dispatch
 ↓
Maybe Legal
```

---

# 57. Border / Cross-Region Operations

RideForge may operate in areas where regional boundaries affect ride legality.

Therefore:

```text
Region Policy
+
Ride Validation
+
Driver Eligibility
```

must be evaluated before assignment.

---

# 58. Stand Dispatch and Regional Rules

A stand may have its own:

```text
Region
Zone
Vehicle Rules
Queue Rules
Operating Hours
```

Dispatch must respect these rules.

---

# 59. Smart Dispatch and Regional Rules

Smart Dispatch must not use geographic proximity alone.

Candidate ranking must occur only after:

```text
Regional Eligibility
```

has been established.

---

# 60. Driver Location

Smart Dispatch depends on current driver location.

The location architecture is governed by:

```text
ADR-0010 — Driver Location Storage Strategy
```

The dispatch system should consume the appropriate location representation rather than inventing a second location system.

---

# 61. Location Freshness

A driver's location must have a freshness policy.

For example:

```text
Fresh
Stale
Unavailable
```

A stale location should not be treated as equally reliable as a current location.

---

# 62. Stale Driver Location

If location is stale:

```text
Candidate Ranking
```

may:

```text
Penalize Candidate
Exclude Candidate
Request Location Refresh
```

depending on policy.

---

# 63. Smart Dispatch Geospatial Search

Candidate discovery may use the platform's geospatial capabilities.

Potential components include:

```text
PostGIS
Redis
Driver Location Store
```

The exact selection follows the location storage ADR and current implementation.

---

# 64. ETA as Dispatch Signal

Smart Dispatch may use:

```text
Estimated Pickup ETA
```

rather than raw geographic distance.

This is important because:

```text
5 km
```

does not necessarily mean:

```text
5 minutes
```

due to:

```text
Road Network
Traffic
Turn Restrictions
Road Quality
Bridges
Border Constraints
```

---

# 65. ETA Provider Strategy

ETA and routing decisions are governed by:

```text
ADR-0017 — ETA and Route Provider Strategy
```

Smart Dispatch should consume the ETA abstraction rather than hard-code a specific map provider.

---

# 66. Fairness

Dispatch should consider fairness where the operating model requires it.

Fairness can mean:

```text
Queue Fairness
Driver Opportunity
Assignment Distribution
Earnings Distribution
Stand Rotation
```

The exact fairness objective depends on the dispatch mode.

---

# 67. Stand Fairness

Stand Dispatch naturally provides a strong fairness mechanism through queue ordering.

The system should avoid unnecessary ranking that undermines this property.

---

# 68. Smart Dispatch Fairness

Smart Dispatch may require explicit fairness mechanisms.

Examples:

```text
Assignment Frequency
Recent Assignment Penalty
Idle-Time Weight
Queue Priority
Fairness Constraints
```

These should be introduced deliberately rather than through arbitrary scoring.

---

# 69. AI Fairness

AI ranking must be monitored for unintended driver-level bias.

Metrics may include:

```text
Assignment Distribution
Idle Time
Acceptance Opportunities
Trip Distance
Earnings Distribution
```

AI should not optimize one metric while producing unacceptable operational outcomes.

---

# 70. Rider Experience

Dispatch quality should ultimately improve:

```text
Pickup ETA
Assignment Speed
Reliability
Cancellation Rate
```

---

# 71. Driver Experience

Dispatch should consider:

```text
Fair Opportunity
Reduced Empty Travel
Predictable Queue Behaviour
Assignment Transparency Where Appropriate
```

---

# 72. Dispatch Objective

RideForge should not define dispatch as:

```text
Always choose nearest driver
```

The objective is broader:

```text
Find an eligible driver
+
Minimize pickup friction
+
Respect operational policy
+
Respect fairness
+
Maintain system efficiency
```

AI may optimize these objectives later.

---

# 73. Smart Dispatch Scoring

A conceptual score may combine:

```text
Pickup ETA
Distance
Idle Time
Driver State
Demand/Supply
Fairness
Vehicle Compatibility
Operational Priority
```

The exact mathematical formulation should be documented separately in the dispatch implementation and AI documentation.

---

# 74. Hard Constraints vs Soft Signals

This distinction is mandatory.

### Hard Constraints

```text
Legal Eligibility
Driver Availability
Vehicle Compatibility
Region Rules
Ride Type
Driver State
```

### Soft Signals

```text
Distance
ETA
Idle Time
Fairness Score
Historical Performance
AI Score
```

Hard constraints must be applied first.

---

# 75. Dispatch Pipeline


The canonical dispatch pipeline is:

```text
1. Receive Ride
2. Validate Ride State
3. Resolve Hierarchical Dispatch Configuration
4. Resolve Effective Dispatch Strategy
5. Validate Regional / Legal Constraints
6. Discover Candidates
7. Apply Hard Constraints
8. Apply Strategy-Specific Candidate Processing
9. Compute Features / ETA Where Required
10. Rank or Apply Queue Ordering
11. Apply Fairness / Policy Rules
12. Reserve Candidate
13. Confirm Assignment
14. Emit Assignment Event
```

The key separation is:

```text
Configuration Resolution
        ↓
Strategy Resolution
        ↓
Candidate Discovery
        ↓
Strategy-Specific Processing
        ↓
Ranking / Queue Selection
        ↓
Assignment
```

# 76. Stand Dispatch Pipeline


The canonical Smart Stand Dispatch pipeline is:

```text
1. Receive Ride
2. Resolve Dispatch Configuration
3. Determine Effective Smart Stand Dispatch Strategy
4. Determine Whether Pickup Is Within a Configured Stand Radius
5. If Applicable, Identify Preferred Stand
6. Inspect Eligible Stand Drivers / Queue
7. Select According to Stand Queue / Ordering Rules
8. If No Suitable Stand Candidate Exists, Expand to Broader Eligible Candidates
9. Consider Nearby Stands / Non-Stand / Nearby-Location Candidates as Applicable
10. Apply Strategy-Specific Candidate Evaluation
11. Reserve Candidate
12. Confirm Assignment
13. Advance / Update Queue According to Policy
14. Emit Assignment Event
```

If the pickup is outside all stand radii, the pipeline does not create a stand-only candidate pool:

```text
Outside All Stand Radii
        ↓
Eligible Nearby Drivers
```

The effective strategy remains Smart Stand Dispatch throughout this process unless an explicit strategy transition is configured.

# 77. Common Assignment Layer

Both strategies should converge on a common assignment boundary.

Conceptually:

```text
Stand Strategy ───┐
                  ├──> Assignment Engine
Smart Strategy ───┤
                  │
AI Strategy ──────┘
```

This prevents each strategy from implementing its own incompatible assignment mechanism.

---

# 78. Strategy Output

Strategies should primarily answer:

```text
Which driver should be considered next?
```

The assignment layer should handle:

```text
Reservation
Concurrency
State Transition
Persistence
Event Publication
```

where appropriate.

---

# 79. Separation of Concerns

Avoid:

```text
Smart Dispatch
    ↓
Direct Database State Mutation
    ↓
Direct Event Publication
```

when the domain architecture provides dedicated application/assignment boundaries.

Prefer:

```text
Strategy
 ↓
Decision
 ↓
Application Service
 ↓
Transaction
 ↓
State + Outbox
```

---

# 80. Dispatch Events

Important dispatch events may include:

```text
DriverAssignmentRequested
DriverAssigned
DriverAssignmentRejected
DriverAssignmentExpired
DispatchFailed
```

The exact event catalog should be maintained by the event architecture.

---

# 81. Assignment Event

After successful assignment, the system should emit a durable event.

Conceptually:

```text
Assignment Transaction
       │
       ├── Ride / Driver State
       │
       └── Outbox: DriverAssigned
               │
             COMMIT
               │
               ▼
            Redpanda
```

---

# 82. Dispatch State

Dispatch attempts should be observable.

Possible states include:

```text
PENDING
SEARCHING
OFFERED
ASSIGNED
REJECTED
EXPIRED
FAILED
CANCELLED
```

The exact state machine should remain consistent across the dispatch implementation.

---

# 83. Dispatch Attempt Identity

Each dispatch attempt should have an identifiable record or event context.

Useful identifiers include:

```text
ride_id
dispatch_attempt_id
driver_id
event_id
correlation_id
```

This makes repeated attempts diagnosable.

---

# 84. Dispatch Attempt Limits

The system should prevent unbounded assignment attempts.

Example:

```text
Attempt 1
Attempt 2
Attempt 3
Stop / Escalate
```

The limit should be policy-driven.

---

# 85. Driver Offer Timeout

If the system uses driver offers, the offer must expire.

Conceptually:

```text
Offer Sent
    ↓
Acceptance Window
    ↓
Accepted → Assignment
    ↓
Rejected / Timeout → Next Candidate
```

---

# 86. Simultaneous Offers

Sending offers to multiple drivers may improve assignment latency but can create:

```text
Race Conditions
Multiple Acceptances
Driver Confusion
```

Therefore simultaneous offers require explicit reservation semantics.

---

# 87. Sequential Offers

Sequential offers are simpler:

```text
Driver A
 ↓
Reject
 ↓
Driver B
 ↓
Reject
 ↓
Driver C
```

but may increase assignment latency.

The strategy should be configurable based on operating requirements.

---

# 88. Batch Offers

Batch offers may be introduced when required.

If used, the system must define:

```text
Maximum Candidates
Reservation Rules
Winner Selection
Cancellation of Losing Offers
```

---

# 89. Stand vs Smart Offer Behaviour

The offer model may differ:

```text
Stand
→ Usually Queue-Head Assignment

Smart
→ Ranked Candidate Selection
```

Both must eventually use the same assignment state machine.

---

# 90. Dispatch Cancellation

If the rider cancels while dispatch is active:

```text
Dispatch
→ Stop Further Assignment
```

Any reserved driver should be released safely.

---

# 91. Driver Goes Offline

If a candidate becomes unavailable during dispatch:

```text
Candidate
→ Invalid
```

and the strategy should continue according to retry/fallback policy.

---

# 92. Location Update During Dispatch

A driver's location may change while ranking is occurring.

The system should accept that dispatch decisions are made using a snapshot of available information.

Critical assignment state must still be validated before final assignment.

---

# 93. Stale Candidate Protection

Before final assignment:

```text
Candidate Selected
      ↓
Revalidate Eligibility
      ↓
Reserve / Assign
```

This protects against:

```text
Driver Offline
Driver Assigned Elsewhere
Vehicle Changed
Region Changed
```

---

# 94. Dispatch Decision Revalidation

The system should not blindly trust a candidate list generated earlier.

The final assignment must verify:

```text
Driver State
Ride State
Policy
Eligibility
Reservation
```

---

# 95. Operational Mode Switching

A region may switch from:

```text
STAND
```

to:

```text
SMART
```

or vice versa.

However, mode switching must account for:

```text
Active Rides
Existing Queue State
Pending Offers
In-Flight Assignments
Operational Staff
```

---

# 96. Safe Mode Transition

A safe transition may be:

```text
Current Mode
     ↓
Stop New Dispatch Decisions
     ↓
Finish / Cancel In-Flight Operations
     ↓
Apply New Policy
     ↓
Resume Dispatch
```

The exact procedure depends on operational requirements.

---

# 97. Configuration Changes

Dispatch strategy configuration should be:

```text
Versioned
Auditable
Validated
Observable
```

Avoid uncontrolled runtime configuration changes.

---

# 98. Configuration Rollback

A bad dispatch configuration should be reversible.

Maintain:

```text
Previous Valid Configuration
```

where operationally important.

---

# 99. Experimentation

Smart Dispatch and AI ranking may be evaluated using controlled experimentation.

Examples:

```text
Strategy A
vs
Strategy B
```

or:

```text
Deterministic Ranking
vs
AI Ranking
```

Experiments must preserve hard constraints.

---

# 100. A/B Testing

A/B tests should define:

```text
Population
Control
Treatment
Metrics
Duration
Safety Limits
Rollback Conditions
```

The experimentation architecture is governed separately by the AI and experimentation documentation.

---

# 101. Metrics

Dispatch should measure:

```text
Time to Assignment
Pickup ETA
Assignment Success Rate
Driver Acceptance Rate
Driver Decline Rate
Ride Cancellation Rate
Dispatch Attempts
No-Candidate Rate
```

---

# 102. Strategy Comparison Metrics

Compare:

```text
Stand Dispatch
Smart Dispatch
AI-Assisted Dispatch
```

using consistent metrics.

Examples:

```text
Median Pickup ETA
P95 Pickup ETA
Assignment Latency
Driver Idle Time
Driver Utilization
Cancellation Rate
```

---

# 103. Fairness Metrics

Monitor:

```text
Assignments per Driver
Idle Time
Trip Opportunity
Earnings Distribution
Stand Queue Position
```

where relevant.

---

# 104. AI Metrics

For AI-assisted dispatch, monitor:

```text
Model Latency
Prediction Quality
Ranking Quality
Fallback Rate
Model Errors
Feature Availability
Model Version
```

---

# 105. Fallback Metrics

Track:

```text
AI → Smart Fallback
Smart → Stand Fallback
No-Candidate Fallback
Retry Fallback
```

A high fallback rate may indicate a system problem.

---

# 106. Dispatch Observability

Every dispatch decision should be traceable to:

```text
Ride
Strategy
Policy Version
Candidate Set
Decision
Assignment
Fallback
```

Sensitive model information should not be exposed unnecessarily.

---

# 107. Decision Logging

Record enough information to reconstruct why a driver was selected.

For example:

```text
strategy
policy_version
candidate_count
selected_driver
decision_timestamp
```

Detailed candidate scores may be stored according to privacy and storage requirements.

---

# 108. Debugging

A dispatch investigation should answer:

```text
Which strategy ran?
Which policy was active?
Which drivers were eligible?
Why was the selected driver chosen?
Was a fallback used?
Was assignment successful?
```

---

# 109. Failure Handling

Dispatch failures should be classified as:

```text
No Candidate
Candidate Became Ineligible
Assignment Conflict
Dependency Failure
Timeout
Policy Failure
AI Failure
Infrastructure Failure
```

Each class should have a defined recovery path.

---

# 110. No Candidate

If no driver qualifies:

```text
Retry
Expand Search
Wait
Fallback
Notify
```

according to policy.

---

# 111. Assignment Conflict

If a candidate is already assigned:

```text
Discard Candidate
```

and continue with the next candidate or retry according to policy.

---

# 112. Dependency Failure

If a required dependency fails:

```text
Retry if transient
Fallback if supported
Fail safely if necessary
```

Never assign an ineligible driver merely to avoid a dispatch failure.

---

# 113. AI Failure

If the AI model times out:

```text
AI Ranking
   ↓
Timeout
   ↓
Deterministic Ranking
```

The ride should not become unserviceable solely because AI is unavailable when a deterministic dispatch path exists.

---

# 114. Stand Data Failure

If stand queue state is unavailable:

```text
Do Not Invent Queue Position
```

Use the configured recovery strategy.

---

# 115. Legal Failure

If regional/legal validation fails:

```text
Do Not Assign
```

This is a hard constraint.

---

# 116. Security

Dispatch APIs and internal operations must enforce:

```text
Authentication
Authorization
Input Validation
Auditability
```

Only authorized components should be able to:

```text
Change Dispatch Policy
Force Assignment
Override Queue
Switch Strategy
Replay Assignment Events
```

---

# 117. Manual Override

Operational tools may eventually allow manual dispatch overrides.

Manual overrides must:

```text
Be Authorized
Be Audited
Respect Hard Legal Constraints
Record Operator Identity
Record Reason
```

---

# 118. Manual Override and AI

A manual override should be represented explicitly rather than pretending the AI or Smart Dispatch made the decision.

For example:

```text
strategy = MANUAL_OVERRIDE
```

with appropriate audit metadata.

---

# 119. Auditability

Dispatch changes should be traceable:

```text
Policy Change
Strategy Change
Manual Override
Assignment
Reassignment
Driver Rejection
Replay
```

---

# 120. Testing Strategy

Dispatch requires:

```text
Unit Tests
Integration Tests
Concurrency Tests
Failure Tests
Load Tests
Simulation Tests
```

---

# 121. Stand Dispatch Tests

Test:

```text
Queue Ordering
Queue Advancement
Driver Eligibility
Driver Re-entry
Decline Handling
Concurrent Ride Requests
Queue Recovery
```

---

# 122. Smart Dispatch Tests

Test:

```text
Candidate Discovery
Eligibility Filtering
Ranking
Search Expansion
Candidate Reservation
Concurrent Assignments
No-Candidate Handling
```

---

# 123. AI Dispatch Tests

Test:

```text
Feature Availability
Model Timeout
Model Failure
Model Version Selection
Fallback Ranking
Invalid Model Output
```

---

# 124. Concurrency Tests

Simulate:

```text
Two Rides
One Driver
```

and verify:

```text
Only One Assignment
```

Also test:

```text
Many Rides
Many Drivers
```

under concurrent dispatch.

---

# 125. Regional Tests

Test:

```text
Allowed Ride
Disallowed Ride
Boundary Location
Region-Specific Strategy
Stand-Specific Strategy
Policy Override
```

---

# 126. Failure Tests

Test:

```text
Redis Failure
PostgreSQL Failure
Redpanda Failure
Location Store Failure
ETA Provider Failure
AI Service Failure
Driver State Race
```

and verify the appropriate fallback.

---

# 127. Load Testing

Measure:

```text
Dispatch Decisions / Second
Candidate Search Latency
Assignment Latency
Database Load
Redis Load
Redpanda Load
AI Inference Latency
```

---

# 128. Simulation

A dispatch simulator should eventually allow replaying historical or synthetic scenarios:

```text
Drivers
Locations
Demand
Ride Requests
Traffic
Stand Queues
```

This can help compare strategies without affecting production.

---

# 129. Offline Evaluation

Smart and AI dispatch strategies should be evaluated against historical or simulated data before production rollout where feasible.

Potential metrics:

```text
Assignment Rate
Pickup ETA
Idle Time
Driver Utilization
Cancellation Rate
```

---

# 130. Production Rollout

New dispatch strategies should be introduced gradually.

Possible rollout:

```text
Development
   ↓
Staging
   ↓
Internal / Test Market
   ↓
Small Production Region
   ↓
Controlled Expansion
```

---

# 131. AI Rollout

AI-assisted dispatch should use a controlled rollout:

```text
Shadow Mode
   ↓
Evaluation
   ↓
Small Traffic
   ↓
A/B Test
   ↓
Expanded Traffic
```

The deterministic strategy should remain available as fallback.

---

# 132. Shadow Mode

In shadow mode:

```text
Production Decision
       │
       ├── Existing Strategy → Real Assignment
       │
       └── New Strategy → Prediction Only
```

The new strategy does not control actual assignments.

This allows comparison before activation.

---

# 133. Rollback

Every dispatch strategy change must have a rollback path.

Example:

```text
AI_ASSISTED
    ↓
Problem
    ↓
SMART
```

or:

```text
SMART
    ↓
Problem
    ↓
STAND
```

only where policy permits.

---

# 134. Operational Guardrails

Dispatch must have guardrails for:

```text
Maximum Search Radius
Maximum Assignment Attempts
Maximum AI Latency
Maximum Offer Duration
Maximum Retry Count
Fallback Strategy
```

---

# 135. Cost Optimization

Smart Dispatch may require:

```text
Geospatial Queries
ETA Calculations
Feature Computation
AI Inference
```

These should be bounded.

Do not call expensive external or AI services for every candidate when a cheaper filter can eliminate candidates first.

---

# 136. Candidate Filtering Before Expensive Scoring

Preferred pipeline:

```text
Cheap Hard Filters
      ↓
Small Candidate Set
      ↓
Moderate-Cost Features
      ↓
Expensive ETA / AI Scoring
```

This reduces dispatch cost and latency.

---

# 137. AI Cost Control

AI inference should be used only when it can provide measurable value.

Possible controls:

```text
Model Tier
Candidate Limit
Inference Timeout
Caching
Fallback
Traffic Percentage
```

---

# 138. Latency Budget

Dispatch should define a target budget for:

```text
Policy Resolution
Candidate Discovery
Feature Computation
Ranking
Reservation
Assignment
```

The AI model must operate within the allowed dispatch budget when enabled.

---

# 139. Dispatch Performance Principle

> **A slightly less sophisticated dispatch decision delivered quickly is often better than a theoretically optimal decision delivered too late.**

The architecture should optimize:

```text
Decision Quality
+
Decision Latency
```

rather than quality alone.

---

# 140. Data Dependencies

Smart Dispatch may consume:

```text
Driver Location
Driver State
Ride State
ETA
Demand
Supply
Historical Features
AI Features
```

Each dependency should have:

```text
Freshness
Availability
Fallback
```

semantics.

---

# 141. Stale Features

AI and Smart Dispatch should not silently treat stale information as current.

Feature metadata may include:

```text
timestamp
source
freshness
```

where required.

---

# 142. Missing Features

If a feature is unavailable:

```text
Fallback Feature
Default Value
Candidate Exclusion
Deterministic Strategy
```

may be used according to policy.

AI should not fail unpredictably because one optional feature is missing.

---

# 143. Strategy Determinism

The deterministic Smart and Stand strategies should be reproducible for debugging.

Given the same:

```text
Input
Policy
Candidate State
Strategy Version
```

the decision should be explainable.

---

# 144. AI Explainability

AI ranking does not require exposing proprietary model internals to users.

However, the system should retain operational metadata sufficient to answer:

```text
Which model was used?
Which policy was used?
Was fallback used?
```

---

# 145. Dispatch Data Privacy

Dispatch systems may process:

```text
Driver Location
Rider Location
Travel History
Assignment History
```

These data must follow RideForge privacy and governance requirements.

---

# 146. Location Privacy

Driver location should be used only for legitimate dispatch purposes and stored according to the location-storage and privacy policies.

---

# 147. Data Retention

Dispatch decision history should have a defined retention strategy.

Retain enough information for:

```text
Operations
Analytics
Dispute Investigation
Model Evaluation
```

without retaining unnecessary sensitive information indefinitely.

---

# 148. Dispatch Architecture Boundary

The dispatch subsystem owns:

```text
Candidate Selection
Strategy Execution
Assignment Decision
Dispatch Attempts
```

It does not own:

```text
User Identity
Payment
External Notifications
Map Provider Internals
```

unless explicitly defined by service boundaries.

---

# 149. Integration With Ride Service

The Ride Service remains responsible for authoritative ride lifecycle state.

Dispatch should not independently create conflicting ride state.

Conceptually:

```text
Ride Service
   ↓
RideCreated
   ↓
Dispatch
   ↓
Assignment Decision
   ↓
Authoritative Assignment Transaction
```

The exact ownership boundary should follow the service architecture.

---

# 150. Integration With Driver Service

Driver-related availability and eligibility should come from the authoritative driver domain.

Dispatch should consume:

```text
Driver State
Location
Eligibility
```

through the defined service/data interfaces.

---

# 151. Integration With ETA Service

Smart Dispatch should use the ETA abstraction rather than embedding route-provider logic.

```text
Dispatch
   ↓
ETA Interface
   ↓
ETA Provider / Engine
```

---

# 152. Integration With AI Services

AI-assisted dispatch should use the AI inference boundary.

```text
Dispatch
   ↓
AI Ranking Interface
   ↓
Model Serving
```

The dispatch engine should remain functional when the AI dependency is unavailable.

---

# 153. Event Flow

A representative event-driven flow is:

```text
Ride Created
     │
     ▼
Redpanda
     │
     ▼
Dispatch Consumer
     │
     ▼
Policy Resolution
     │
     ├── Stand
     │
     ├── Smart
     │
     └── AI-Assisted
            │
            ▼
       Candidate Selection
            │
            ▼
         Assignment
            │
            ▼
       DriverAssigned
            │
            ▼
         Redpanda
```

---

# 154. Decision Matrix

| Capability | Stand Dispatch | Smart Dispatch | AI-Assisted Dispatch |
|---|---:|---:|---:|
| Queue fairness | **Strong** | Conditional | Conditional |
| Dynamic proximity | Limited | **Strong** | **Strong** |
| Operational predictability | **Strong** | Strong | Moderate |
| AI dependency | None | None | **Yes, optional** |
| Computational cost | **Low** | Moderate | Higher |
| Geographic optimization | Limited | **Strong** | **Strong** |
| Easy manual explanation | **Strong** | **Strong** | Moderate |
| Regional configurability | **Yes** | **Yes** | **Yes** |
| Failure fallback | **Yes** | **Yes** | **Required** |
| Initial implementation fit | **Yes** | **Yes** | Conditional |

---

# 155. Why Not Use Only Smart Dispatch?

A single Smart Dispatch algorithm would reduce architectural duplication but could be unsuitable for environments where:

```text
Stand Queue Fairness
Operational Rules
Local Transport Practices
```

are more important than dynamic optimization.

Therefore:

```text
Smart Dispatch Only
```

is rejected.

---

# 156. Why Not Use Only Stand Dispatch?

Stand-only dispatch is simple but cannot efficiently handle environments where:

```text
Drivers Are Distributed
Demand Is Dynamic
Pickup Time Matters
No Physical Queue Exists
```

Therefore:

```text
Stand Dispatch Only
```

is rejected.

---

# 157. Why Not Make AI Mandatory?

Mandatory AI would introduce:

```text
Inference Dependency
Latency
Cost
Operational Complexity
Model Failure Risk
```

into the critical dispatch path.

Therefore:

```text
AI as Mandatory Dispatch Engine
```

is rejected.

---

# 158. Decision Matrix — Alternatives

| Approach | Flexibility | Complexity | Failure Resilience | RideForge Fit |
|---|---:|---:|---:|---:|
| Stand Only | Low | Low | High | No |
| Smart Only | Moderate | Moderate | High | No |
| AI Only | High | High | Low without fallback | No |
| Stand + Smart | **High** | **Moderate** | **High** | **Yes** |
| Stand + Smart + Optional AI | **Very High** | Moderate/High | **High** | **Primary** |

---

# 159. Consequences

## 159.1 Positive Consequences

The decision provides:

```text
Operational Flexibility
Regional Strategy Selection
Stand-Based Fairness
Dynamic Dispatch Optimization
Future AI Integration
Controlled AI Adoption
Common Assignment Boundary
Deterministic Fallbacks
```

---

## 159.2 Negative Consequences

The architecture introduces:

```text
Multiple Dispatch Strategies
Strategy Configuration
More Testing
Strategy Comparison
Operational Complexity
Additional Metrics
```

These trade-offs are accepted.

---

# 160. Risks

## Risk 1 — Strategy Configuration Becomes Too Complex

### Mitigation

Use:

```text
Central Policy Model
Explicit Precedence
Versioned Configuration
```

---

## Risk 2 — AI Becomes a Hidden Hard Dependency

### Mitigation

Require:

```text
Deterministic Smart Fallback
AI Timeout
AI Failure Handling
```

---

## Risk 3 — Dispatch Strategies Diverge

### Mitigation

Keep a common:

```text
Assignment Boundary
Dispatch State Model
Event Contract
Observability Model
```

---

## Risk 4 — Inconsistent Fairness

### Mitigation

Define fairness objectives explicitly for each dispatch mode.

---

## Risk 5 — Race Conditions

### Mitigation

Use:

```text
Reservation
Atomic State Transitions
Transactions
Concurrency Testing
```

---

## Risk 6 — Legal Constraints Bypassed

### Mitigation

Apply regional and legal validation as a hard eligibility constraint before ranking.

---

# 161. Validation

This ADR should be validated through:

```text
Stand Dispatch Tests
Smart Dispatch Tests
Concurrency Tests
Regional Policy Tests
Fallback Tests
AI Failure Tests
Load Tests
Simulation
A/B Testing
Production Metrics
```

---

# 162. Review Triggers

Revisit this ADR when:

```text
A New Dispatch Strategy Is Introduced
Stand Operations Change
Regional Rules Change
AI Becomes a Mandatory Dependency
Dispatch Scale Changes Significantly
A Centralized Fleet Optimization System Is Introduced
Multi-Region Dispatch Is Introduced
A New Assignment Model Is Proposed
```

---

# 163. Related Documentation

This ADR should be read together with:

```text
03-components/
04-development/
05-ai/
```

Especially:

```text
Smart Dispatch AI
ETA and Prediction System
Driver Demand and Supply Prediction
AI Matching and Ranking
Feature Engineering
AI Failure and Fallback Strategy
AI Monitoring and Model Observability
AI Experimentation and A/B Testing
Event and Messaging Development
Database Development
Redis Development
Performance and Optimization
```

---

# 164. Related ADRs

This decision is directly related to:

```text
ADR-0002 — Architecture Style
ADR-0003 — Microservice Boundaries
ADR-0004 — Domain-Driven Design
ADR-0005 — Event-Driven Architecture
ADR-0006 — Kafka / Redpanda for Event Streaming
ADR-0008 — PostGIS for Geospatial Operations
ADR-0009 — Redis for Real-Time State and Caching
ADR-0010 — Driver Location Storage Strategy
ADR-0012 — Outbox Pattern
ADR-0013 — Dead Letter Queue Strategy
ADR-0014 — API and Service Communication
ADR-0016 — AI-Assisted Dispatch Strategy
ADR-0017 — ETA and Route Provider Strategy
ADR-0018 — Regional and Legal Ride Validation
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
ADR-0025 — Testing and Integration Strategy
ADR-0026 — Model and AI Governance
ADR-0028 — Cost Optimization Strategy
```

---

# 165. Decision Summary

RideForge adopts a hybrid dispatch architecture:

```text
                    Dispatch Platform
                           │
                    Dispatch Policy
                           │
             ┌─────────────┼─────────────┐
             ▼             ▼             ▼
          STAND          SMART       AI-ASSISTED
             │             │             │
             │             │       AI Ranking
             │             │             │
             └─────────────┼─────────────┘
                           ▼
                   Common Assignment
                       Boundary
                           │
                           ▼
                    Driver Assignment
                           │
                           ▼
                     Ride State
                           │
                           ▼
                       Redpanda
```

The operating model is:

```text
Stand Dispatch
→ Queue-Based, Predictable, Fair

Smart Dispatch
→ Dynamic, Location-Aware, Deterministic

AI-Assisted Dispatch
→ Optional Intelligence Layer With Safe Fallback
```

---

# 166. Final Principle

> **RideForge will support both Stand Dispatch and Smart Dispatch as first-class strategies, selected by explicit operational policy, while AI-assisted ranking remains an optional enhancement with deterministic fallback.**

The dispatch decision hierarchy is:

```text
1. Validate Ride
2. Validate Region / Legal Rules
3. Resolve Dispatch Policy
4. Select Dispatch Strategy
5. Apply Hard Eligibility Constraints
6. Rank or Select Candidate
7. Revalidate Candidate
8. Reserve Driver
9. Assign Driver Atomically
10. Publish Assignment Event
```

The architecture deliberately separates:

```text
Policy
+
Strategy
+
Ranking
+
Assignment
+
Event Publication
```

so that RideForge can operate different dispatch models in different markets without creating separate dispatch platforms.

---

# 167. Status

```text
Decision: ACCEPTED

Dispatch Architecture:
Hybrid

Supported Strategies:
Stand Dispatch
Smart Dispatch
AI-Assisted Dispatch

Strategy Selection:
Policy-Based

Hard Constraints:
Always Applied Before Ranking

AI:
Optional

AI Failure:
Deterministic Fallback Required

Assignment:
Common Assignment Boundary

Regional / Legal Validation:
Mandatory

Primary Goal:
Operational Flexibility + Efficient Driver Assignment
```

This decision establishes the RideForge dispatch strategy architecture and preserves the ability to operate both stand-based and smart dynamic dispatch according to the needs of each operating environment.


---

# 104. Canonical Dispatch Model Clarification

This section is authoritative for the relationship between dispatch strategy, configuration, candidate discovery, and AI assistance.

## 104.1 Two Primary Strategies

```text
Smart Stand Dispatch
Smart Dispatch
```

AI-assisted dispatch is an optimization capability, not a third primary strategy.

## 104.2 Hierarchical Strategy Configuration

The effective strategy is resolved from the most specific applicable configuration upward:

```text
Most Specific Explicit Configuration
        ↓
Parent Configuration
        ↓
...
        ↓
System Default
```

Specific configuration overrides inherited configuration.

## 104.3 Smart Stand Dispatch

```text
Inside Stand Radius
        ↓
Prefer Applicable Stand
        ↓
Apply Stand Queue / Ordering
        ↓
No Suitable Stand Candidate?
        ↓
Broaden Eligible Candidate Discovery
```

Smart Stand Dispatch is therefore:

> **Stand-preferred, not stand-exclusive.**

Outside all stand radii, the candidate search is not restricted to stand drivers.

## 104.4 Smart Dispatch

Smart Dispatch is stand-agnostic:

```text
Eligible Nearby Drivers
        ↓
Ranking
        ↓
Best Candidate
```

Stand membership does not create an implicit preference.

## 104.5 Cross-Location Dispatch

A location's strategy is not a hard boundary preventing its drivers from serving nearby locations.

Example:

```text
Location A → Smart Dispatch
Location B → Smart Stand Dispatch

Ride in A
   ↓
No Suitable Local Driver
   ↓
Expand to B
   ↓
Evaluate B's Eligible Drivers
```

When evaluating B candidates, preserve B's strategy/stand/queue context for strategy-specific prioritization.

## 104.6 Separation of Concerns

```text
Configuration Hierarchy
        ↓
Effective Dispatch Strategy
        ↓
Candidate Discovery
        ↓
Eligibility
        ↓
Strategy-Specific Prioritization
        ↓
Ranking / Queue Selection
        ↓
Assignment
```

The following must not be conflated:

```text
Dispatch Strategy
≠ Candidate Pool Boundary

Candidate Expansion
≠ Strategy Switching

AI Failure
≠ Strategy Switching

Not Preferred
≠ Ineligible
```

## 104.7 Hard Constraints

All strategies remain subject to:

```text
Legal / Regional Constraints
Driver Availability
Vehicle / Service Compatibility
Safety Constraints
Ride Constraints
Other Hard Eligibility Rules
```

AI and soft ranking signals cannot override these constraints.

## 104.8 Implementation Guardrails

Implementations must not:

```text
Treat Smart Stand Dispatch as stand-only.

Reject non-stand drivers solely because the rider is inside a stand radius.

Reject nearby-location drivers solely because their source location uses another dispatch strategy.

Hard-code a fixed configuration hierarchy.

Resolve configuration inheritance inside ranking or geo components.

Switch Smart Stand Dispatch to Smart Dispatch merely because stand supply is unavailable.

Treat AI failure as permission to change dispatch strategy.

Use AI score to bypass a configured stand queue rule.

Treat candidate preference as candidate eligibility.
```
