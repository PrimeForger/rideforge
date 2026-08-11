# 03 — Ride and Driver Lifecycle

> **Format:** Markdown + Mermaid  
> **Scope:** Ride lifecycle, driver lifecycle, assignment relationship, recovery paths  
> **Purpose:** Compact lifecycle context for AI agents and engineers.

---

## 1. Purpose

This document visualizes the two core operational lifecycles in RideForge:

```text
Ride Lifecycle
Driver Lifecycle
```

It also shows how they connect through:

```text
Validation
Dispatch
Assignment
Acceptance
Pickup
Trip
Completion
Cancellation
Re-dispatch
```

Detailed business rules remain in the domain, development, AI, and ADR documentation.

---

## 2. High-Level Relationship

```text
Passenger
    ↓
Ride Request
    ↓
Validation
    ↓
Dispatch
    ↓
Driver Assignment
    ↓
Driver Acceptance
    ↓
Pickup
    ↓
Trip
    ↓
Ride Completion
    ↓
Driver Available Again
```

---

## 3. Ride Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Requested

    Requested --> Validating
    Validating --> Dispatching
    Validating --> Cancelled

    Dispatching --> DriverAssigned
    Dispatching --> Cancelled
    Dispatching --> Expired

    DriverAssigned --> DriverAccepted
    DriverAssigned --> Dispatching
    DriverAssigned --> Cancelled
    DriverAssigned --> Expired

    DriverAccepted --> DriverArriving
    DriverAccepted --> Cancelled

    DriverArriving --> DriverArrived
    DriverArriving --> Cancelled

    DriverArrived --> InProgress
    DriverArrived --> Cancelled

    InProgress --> Completed
    InProgress --> Cancelled

    Completed --> [*]
    Cancelled --> [*]
    Expired --> [*]
```

---

## 4. Ride State Meaning

| State | Meaning |
|---|---|
| `Requested` | Passenger requested a ride |
| `Validating` | Required ride and operating constraints are checked |
| `Dispatching` | System is finding an eligible driver |
| `DriverAssigned` | Driver has been selected |
| `DriverAccepted` | Selected driver accepted |
| `DriverArriving` | Driver is travelling to pickup |
| `DriverArrived` | Driver reached pickup |
| `InProgress` | Trip is active |
| `Completed` | Trip completed successfully |
| `Cancelled` | Ride was cancelled |
| `Expired` | Ride could not proceed within its applicable lifecycle window |

The exact production state model may contain additional implementation states.

---

## 5. Driver Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Offline

    Offline --> Available
    Available --> Assigned

    Assigned --> Accepted
    Assigned --> Available
    Assigned --> Offline

    Accepted --> Arriving
    Accepted --> Offline

    Arriving --> AtPickup
    Arriving --> Offline

    AtPickup --> OnTrip
    AtPickup --> Offline

    OnTrip --> Available
    OnTrip --> Offline

    Available --> Offline
```

---

## 6. Driver State Meaning

| State | Meaning |
|---|---|
| `Offline` | Driver is not participating in active dispatch |
| `Available` | Driver can potentially receive a ride |
| `Assigned` | Ride has been assigned |
| `Accepted` | Driver accepted the ride |
| `Arriving` | Driver is travelling toward pickup |
| `AtPickup` | Driver reached pickup |
| `OnTrip` | Driver is serving the active trip |

Actual implementation may contain additional operational states.

---

## 7. Combined Lifecycle

```mermaid
flowchart TB
    subgraph Ride["Ride"]
        R1["Requested"]
        R2["Validating"]
        R3["Dispatching"]
        R4["Driver Assigned"]
        R5["Driver Accepted"]
        R6["Driver Arriving"]
        R7["Driver Arrived"]
        R8["In Progress"]
        R9["Completed / Cancelled"]

        R1 --> R2 --> R3 --> R4 --> R5 --> R6 --> R7 --> R8 --> R9
    end

    subgraph Driver["Driver"]
        D1["Available"]
        D2["Assigned"]
        D3["Accepted"]
        D4["Arriving"]
        D5["At Pickup"]
        D6["On Trip"]
        D7["Available Again"]

        D1 --> D2 --> D3 --> D4 --> D5 --> D6 --> D7
    end

    R4 -. assignment .-> D2
    R5 -. acceptance .-> D3
    R6 -. movement .-> D4
    R7 -. pickup .-> D5
    R8 -. trip .-> D6
    R9 -. completion .-> D7
```

The two domains remain separate. The assignment relationship connects them.

---

## 8. Validation Before Dispatch

```mermaid
flowchart LR
    Request["Ride Request"]
    Validation["Ride + Regional / Legal Validation"]
    Dispatch["Dispatch"]
    Rejected["Rejected"]

    Request --> Validation
    Validation -->|Allowed| Dispatch
    Validation -->|Not Allowed| Rejected
```

The distinction is:

```text
Regional / Legal Validation
    = Can this ride proceed?

Dispatch
    = Which eligible driver should receive it?
```

---

## 9. Driver Candidate Flow

```text
All Drivers
    ↓
Location Candidates
    ↓
Eligibility Filtering
    ↓
Dispatch Candidates
    ↓
Ranking / Selection
    ↓
Driver Assignment
```

A nearby driver is not automatically an eligible driver.

Eligibility may depend on:

```text
Availability
Location Freshness
Vehicle / Service Requirements
Regional Rules
Operational Rules
Existing Assignment
Other Hard Constraints
```

---

## 10. Assignment Flow

```mermaid
sequenceDiagram
    participant D as Dispatch
    participant R as Ride
    participant Dr as Driver
    participant E as Event Platform

    D->>R: Select Driver
    R->>Dr: Assignment
    R->>E: Assignment Event

    alt Driver accepts
        Dr-->>R: Accept
        R->>E: Acceptance Event
    else Rejects / times out
        Dr-->>R: Reject / Timeout
        R->>D: Re-dispatch
    end
```

---

## 11. Re-Dispatch

A ride may return to dispatch when:

```text
Driver Rejects
Driver Times Out
Driver Becomes Unavailable
Assignment Becomes Invalid
Operational Failure Occurs
```

```mermaid
flowchart LR
    Assigned["Driver Assigned"]
    Failure["Reject / Timeout / Invalid"]
    Redispatch["Re-dispatch"]

    Assigned --> Failure --> Redispatch --> Assigned
```

The retry/reassignment rules are governed by the dispatch and failure strategies.

---

## 12. Stand and Smart Dispatch

Both strategies converge on the same ride/driver lifecycle:

```mermaid
flowchart TB
    ValidRide["Valid Ride"]
    Strategy["Dispatch Strategy"]

    Stand["Stand Dispatch"]
    Smart["Smart Dispatch"]

    Assignment["Driver Assignment"]
    Accepted["Driver Accepted"]
    Trip["Trip"]

    ValidRide --> Strategy
    Strategy --> Stand
    Strategy --> Smart

    Stand --> Assignment
    Smart --> Assignment

    Assignment --> Accepted --> Trip
```

Therefore:

```text
Dispatch Strategy
```

can evolve independently from:

```text
Ride Lifecycle
Driver Lifecycle
```

---

## 13. AI Relationship

AI may assist:

```text
Candidate Ranking
Driver Selection
ETA Prediction
Demand / Supply Prediction
```

but does not own:

```text
Ride State
Driver State
Legal Eligibility
Hard Safety Constraints
Payment State
```

```mermaid
flowchart LR
    Candidates["Eligible Candidates"]
    AI["AI / Ranking"]
    Decision["Dispatch Decision"]
    Ride["Ride State"]
    Driver["Driver State"]

    Candidates --> AI --> Decision
    Decision --> Ride
    Decision --> Driver
```

---

## 14. Location Relationship

Driver location supports several lifecycle stages:

```text
Before Assignment
    ↓
Candidate Discovery

After Assignment
    ↓
Driver Arrival

During Trip
    ↓
Trip Tracking
```

Conceptually:

```mermaid
flowchart TB
    Driver["Driver"]
    State["Driver State"]
    Location["Current Location"]
    Freshness["Location Freshness"]
    Dispatch["Dispatch"]

    Driver --> State
    Driver --> Location
    Location --> Freshness

    State --> Dispatch
    Location --> Dispatch
    Freshness --> Dispatch
```

Location is a real-time capability rather than ordinary driver profile state.

---

## 15. Cancellation

Cancellation may occur at different stages:

```mermaid
flowchart TB
    Requested["Requested"]
    Dispatching["Dispatching"]
    Assigned["Assigned"]
    Accepted["Accepted"]
    Arriving["Arriving"]
    Arrived["Arrived"]
    Trip["In Progress"]

    Cancelled["Cancelled"]

    Requested --> Cancelled
    Dispatching --> Cancelled
    Assigned --> Cancelled
    Accepted --> Cancelled
    Arriving --> Cancelled
    Arrived --> Cancelled
    Trip --> Cancelled
```

Whether cancellation is permitted at a particular stage and its consequences are business rules.

---

## 16. Normal Completion

```text
Driver Arrived
      ↓
Trip Started
      ↓
Trip In Progress
      ↓
Destination Reached
      ↓
Ride Completed
      ↓
Driver Released
      ↓
Driver Available Again
```

```mermaid
sequenceDiagram
    participant D as Driver
    participant R as Ride
    participant E as Event Platform

    D->>R: Start Trip
    R->>E: Trip Started

    D->>R: Trip Progress
    R->>E: Trip Progress

    D->>R: Complete Trip
    R->>E: Ride Completed
```

---

## 17. Lifecycle Events

Important transitions may produce domain events:

```text
Ride Created
Ride Dispatching
Driver Assigned
Driver Accepted
Trip Started
Ride Completed
Ride Cancelled

Driver Available
Driver Assigned
Driver Accepted
Driver Arriving
Driver At Pickup
Driver On Trip
Driver Available Again
```

Exact event names and schemas belong to the event architecture documentation and implementation.

---

## 18. Outbox Relationship

Lifecycle state changes and event publication follow the established event architecture:

```mermaid
flowchart LR
    State["Ride / Driver State Change"]
    Tx["Database Transaction"]
    Outbox["Outbox"]
    Stream["Kafka / Redpanda"]

    State --> Tx
    Tx --> Outbox
    Outbox --> Stream
```

This preserves the relationship between transactional state and event publication.

---

## 19. Idempotency

Lifecycle operations may be retried or delivered more than once.

Important operations include:

```text
Assignment
Acceptance
Trip Start
Trip Completion
Cancellation
```

Conceptually:

```text
Event / Command
      ↓
Idempotency
      ↓
Valid State Transition
```

Consumers should not create duplicate business effects when processing the same operation more than once.

---

## 20. Invalid State Transitions

Lifecycle state must not move arbitrarily.

For example:

```text
Completed → Driver Assigned
Cancelled → In Progress
```

are not normal transitions.

The owning domain must reject invalid transitions unless an explicit recovery/business operation exists.

---

## 21. Concurrent Lifecycle Operations

Lifecycle transitions may race.

Example:

```text
Driver Accepts
```

while:

```text
Assignment Timeout
```

is being processed.

The implementation must protect the authoritative state from ending in an invalid condition.

The exact concurrency mechanism belongs to the data consistency and development documentation.

---

## 22. Authoritative State

Events are facts about state changes.

They are not themselves the authoritative state.

```text
Domain State
    ↓
State Transition
    ↓
Domain Event
    ↓
Consumers / Derived State
```

The owning domain remains authoritative.

---

## 23. Failure and Recovery

A lifecycle failure should produce an explicit outcome:

```text
Retry
Reconciliation
Re-dispatch
Fallback
Cancellation
Expiration
Recovery
```

The system should not silently leave a ride permanently stuck between lifecycle states.

---

## 24. Lifecycle Observability

Important transitions should be observable through:

```text
Logs
Metrics
Traces
Events
Alerts
```

Useful lifecycle measurements include:

```text
Time to Assignment
Assignment Acceptance Rate
Assignment Timeout Rate
Re-Dispatch Rate
Time to Pickup
Cancellation Rate
Completion Rate
Ride Expiration Rate
Trip Duration
```

These are representative metrics, not a complete mandatory metric list.

---

## 25. Core Lifecycle Rules

```text
1. Ride and Driver are separate lifecycle domains.

2. Dispatch connects the two lifecycles.

3. Validation occurs before dispatch.

4. Regional and legal rules remain authoritative.

5. Driver eligibility is checked before assignment.

6. Driver location supports candidate discovery and trip operations.

7. Stand Dispatch and Smart Dispatch converge on the same lifecycle.

8. AI may assist decisions but does not own lifecycle state.

9. ETA supports lifecycle decisions but does not own ride state.

10. Assignment is a critical state transition.

11. Assignment must safely handle retries and concurrency.

12. Retryable lifecycle operations require appropriate idempotency.

13. Events represent state changes; they do not replace authoritative state.

14. Derived consumers may temporarily lag authoritative state.

15. Invalid state transitions must be rejected.

16. Terminal ride states should not normally return to active states.

17. Failures must have explicit recovery or terminal outcomes.

18. Important lifecycle transitions must remain observable.

19. The lifecycle should remain stable when dispatch implementations evolve.
```

---

## 26. AI Agent Context

For lifecycle-related work, load this document with:

```text
02-SERVICE_AND_DOMAIN_ARCHITECTURE.md
04-DISPATCH_ARCHITECTURE.md
06-EVENT_DRIVEN_AND_DATA_FLOW.md
08-SECURITY_FAILURE_AND_OBSERVABILITY.md
```

Relevant ADRs include:

```text
ADR-0015 — Smart Dispatch and Stand Dispatch
ADR-0018 — Regional and Legal Ride Validation
ADR-0019 — Data Consistency and Transaction Boundaries
ADR-0020 — Idempotency Strategy
ADR-0021 — Failure and Degradation Strategy
ADR-0022 — Observability Strategy
```

For AI-assisted lifecycle decisions also load:

```text
ADR-0016 — AI-Assisted Dispatch Strategy
ADR-0026 — Model and AI Governance
07-AI_AND_ML_ARCHITECTURE.md
```

---

## 27. Related Diagrams

```text
01-SYSTEM_CONTEXT_AND_HIGH_LEVEL_ARCHITECTURE.md
02-SERVICE_AND_DOMAIN_ARCHITECTURE.md
04-DISPATCH_ARCHITECTURE.md
05-DRIVER_LOCATION_AND_GEOSPATIAL_FLOW.md
06-EVENT_DRIVEN_AND_DATA_FLOW.md
07-AI_AND_ML_ARCHITECTURE.md
08-SECURITY_FAILURE_AND_OBSERVABILITY.md
09-CLOUD_DEPLOYMENT_AND_INFRASTRUCTURE.md
10-DIAGRAM_INDEX.md
```

---

## 28. Maintenance Rules

Update this document when:

```text
Ride lifecycle states materially change
Driver lifecycle states materially change
Assignment semantics change
Cancellation semantics change
Dispatch-to-lifecycle integration changes
Lifecycle event architecture changes
```

Do not update it for ordinary implementation details.

---

## 29. Completion Criteria

```text
□ Ride Lifecycle Represented
□ Driver Lifecycle Represented
□ Ride / Driver Relationship Represented
□ Validation Before Dispatch Represented
□ Assignment Represented
□ Re-Dispatch Represented
□ Cancellation Represented
□ Completion Represented
□ Location Relationship Represented
□ Event Relationship Represented
□ Outbox Relationship Represented
□ Idempotency Considered
□ Failure / Recovery Considered
□ Stand / Smart Dispatch Convergence Represented
□ AI Role Represented
□ Observability Represented
□ Related ADRs Referenced
□ Related Diagrams Referenced
```

---

## 30. Status

```text
Status: Complete

Document:
03-RIDE_AND_DRIVER_LIFECYCLE.md

Diagram Type:
Ride Lifecycle + Driver Lifecycle

Primary Audience:
AI Agents
Architects
Backend Engineers
Dispatch Engineers

Previous Diagram:
02-SERVICE_AND_DOMAIN_ARCHITECTURE.md

Next Diagram:
04-DISPATCH_ARCHITECTURE.md
```
