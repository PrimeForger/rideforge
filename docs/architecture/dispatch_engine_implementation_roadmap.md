# Dispatch Engine Implementation Roadmap

> Status: In Progress
>
> Goal:
> Build a production-grade dispatch engine capable of scaling from a single city to a multi-region ride-hailing platform while minimizing future architectural refactoring.

---

# Architecture Principles

These principles must never be violated.

## 1. Single Responsibility

Every package, service and function should have exactly one responsibility.

Examples

- Candidate search only searches candidates.
- Ranking only ranks.
- Reservation only reserves.
- Matching engine only orchestrates.

---

## 2. Orchestration vs Business Logic

Application services orchestrate.

Infrastructure performs infrastructure work.

Domain performs business rules.

Example

MatchingEngine

✓ Candidate Search

✓ Ranking

✓ Reservation

✓ Offer Creation

✗ Redis implementation

✗ H3 implementation

✗ Kafka implementation

---

## 3. Dependency Direction

Application

↓

Domain

↓

Ports

↓

Infrastructure

Infrastructure never owns business decisions.

---

## 4. Performance First

The dispatch path executes for every ride request.

Every additional Redis call,
allocation,
database query,
or network round-trip must be justified.

---

## 5. Scale Horizontally

Every component must be replaceable by a distributed implementation without changing business logic.

Examples

Current

Redis

↓

Future

Redis Cluster

Current

Single Dispatcher

↓

Future

Regional Dispatch Nodes

---

## 6. Observable By Default

Every critical path must expose

- metrics
- tracing
- structured logs

No hidden operations.

---

## 7. Failure Safety

Infrastructure failures must never leave inconsistent dispatch state.

Examples

- reservation cleanup
- atomic Redis updates
- outbox consistency
- retry safety

---

# Overall Progress

| Stage | Name                             | Status         |
| ----- | -------------------------------- | -------------- |
| 1     | Production Grade H3 Driver Index | 🟡 In Progress |
| 2     | Intelligent Candidate Search     | ⬜ Not Started |
| 3     | Candidate Retrieval Pipeline     | ⬜ Not Started |
| 4     | Ranking Engine                   | ⬜ Not Started |
| 5     | Dispatch Optimizer               | ⬜ Not Started |
| 6     | Geo Engine                       | ⬜ Not Started |
| 7     | Hot Region Scaling               | ⬜ Not Started |
| 8     | Multi-node Scaling               | ⬜ Not Started |
| 9     | Reliability                      | ⬜ Not Started |
| 10    | Observability                    | ⬜ Not Started |
| 11    | ML Readiness                     | ⬜ Not Started |

---

# Stage 1 — Production Grade H3 Driver Index

## Goal

Build an H3 indexing layer that remains correct under concurrency,
minimizes Redis writes,
and is ready for Redis Cluster deployment.

---

## Completed

### Atomic Driver Cell Movement

Status

✅ Completed

Features

- Lua based atomic movement
- Single Redis round trip
- No stale memberships
- Driver mapping updated atomically

---

### Skip Same Cell Updates

Status

✅ Completed

Features

- No unnecessary SREM
- No unnecessary SADD
- TTL refresh only

---

### Empty Cell Cleanup

Status

✅ Completed

Features

- Remove empty H3 sets
- Prevent orphan cell keys

---

### Pipeline Cell Lookup

Status

✅ Completed

Features

- Pipeline SMEMBERS
- Single network round trip

---

### Metrics

Status

✅ Completed

Metrics

- h3_cell_changes_total
- h3_same_cell_updates_total
- h3_driver_additions_total
- h3_driver_removals_total
- h3_lookup_duration_seconds
- h3_drivers_per_lookup

---

### Tracing

Status

✅ Completed

Tracing

- backend
- lookup latency
- ring count
- pipeline size
- returned drivers

---

## Remaining

None

---

## Completion Criteria

- [x] Atomic updates
- [x] No race conditions
- [x] Pipeline lookup
- [x] Metrics
- [x] Tracing
- [x] Cluster compatible

---

# Stage 2 — Intelligent Candidate Search

## Goal

Replace today's fixed-radius / fixed-ring candidate lookup with an adaptive,
incremental search engine that minimizes Redis work while maintaining the
highest probability of quickly finding the best drivers.

This stage changes **how candidates are discovered**, not how they are ranked.

After this stage the Matching Engine will no longer ask:

"Give me every nearby driver."

Instead it will ask:

"Keep expanding until enough good candidates are found."

This dramatically reduces:

- Redis work
- network traffic
- heap allocations
- unnecessary ranking
- candidate filtering cost

while improving dispatch latency.

---

# Current Flow

Current implementation

Ride

↓

RetryPolicy

↓

Radius

↓

NeighborCells()

↓

GetDriversInCells()

↓

Return all drivers

↓

Ranking

Problems

• Always searches the entire configured ring.

• Dense city and sparse village behave exactly the same.

• Wastes Redis calls.

• Wastes CPU ranking unnecessary drivers.

• Not adaptive.

---

# Target Flow

Ride

↓

Search Strategy

↓

Ring 0

↓

Enough candidates?

↓

YES

↓

Stop

↓

NO

↓

Ring 1

↓

Enough?

↓

YES

↓

Stop

↓

NO

↓

Ring 2

↓

...

↓

Ranking

Search stops immediately once enough candidates exist.

---

# Stage Overview

Stage 2 consists of three independent capabilities.

1.

Progressive Ring Expansion

2.

Density-aware Expansion

3.

Search Budget

Each capability builds on the previous one.

Do not skip any.

---

# Step 7 — Progressive Ring Expansion

Status

⬜ Pending

Priority

★★★★★

---

## Why

Today

GridDisk(4)

returns every cell inside ring 4.

That means

Ring0

-

Ring1

-

Ring2

-

Ring3

-

Ring4

are searched every time.

Even if Ring0 already has enough drivers.

That is unnecessary work.

---

## New Behaviour

Search

Ring0

↓

Enough?

↓

Yes

↓

Stop

↓

No

↓

Ring1

↓

Enough?

↓

Stop

↓

Ring2

↓

...

Each ring is searched independently.

---

## Required Refactoring

Remove the concept of

NeighborCells()

from the dispatch flow.

Instead expose two responsibilities.

CellForLocation()

returns only the center cell.

GridRing()

returns one exact ring.

NOT GridDisk.

Matching should orchestrate ring expansion.

The H3 service should never decide search strategy.

---

## New Responsibilities

H3Service

Responsible for

- coordinate validation
- center cell calculation
- exact ring calculation

NOT

- stopping criteria
- candidate limits
- dispatch logic

H3DriverIndex

Responsible for

- lookup for supplied cells

NOT

- ring expansion

Matching

Responsible for

- deciding whether another ring is required

This keeps responsibilities clean.

---

## Acceptance Criteria

✓ Ring0 searched first

✓ Stop immediately when enough candidates found

✓ Ring1 only if required

✓ Ring2 only if required

✓ No duplicated drivers

✓ No duplicated Redis lookups

---

## Testing

Dense city

Expected

Ring0 only

Suburban

Expected

Ring0

↓

Ring1

Rural

Expected

Ring0

↓

Ring1

↓

Ring2

↓

Ring3

Maximum ring

Expected

Stop after configured maximum.

---

# Step 8 — Density-aware Expansion

Status

⬜ Pending

Priority

★★★★★

---

## Why

Searching one ring at a time is still not optimal.

Different areas behave differently.

City

Thousands of drivers.

Village

Five drivers.

Airport

Hundreds.

Highway

Almost none.

Dispatch should adapt automatically.

---

## Behaviour

Sparse region

Expand aggressively.

Dense region

Stay local.

---

## Inputs

Current candidate count

Current ring

Configured candidate target

Maximum ring

These are enough for version one.

No machine learning required.

---

## Dispatch Decision

Pseudo flow

Ring0

↓

Found 3

↓

Need 40

↓

Expand immediately

versus

Ring0

↓

Found 45

↓

Stop

---

## Responsibilities

Search Strategy

Owns expansion policy.

Matching Engine

Consumes the strategy.

Retry Policy

Must not know H3.

This is important.

Retry Policy manages retries.

Search Strategy manages geography.

These are different responsibilities.

---

## Acceptance Criteria

Dense city

Search stops immediately.

Sparse village

Search expands automatically.

No hardcoded behaviour.

---

# Step 9 — Search Budget

Status

⬜ Pending

Priority

★★★★★

---

## Why

Current lookup usually asks Redis for

CandidateLimit

drivers.

But dispatch doesn't actually need every driver.

It only needs enough to satisfy

OfferBatch

-

Filtering Loss

-

Ranking Margin

Everything else wastes work.

---

## Example

Need

20

drivers

Ring0

↓

12

Need

8

more

Ring1

↓

15

Enough

Stop

Never visit Ring2.

---

## Budget Ownership

Search Strategy

Maintains

RemainingBudget

Each lookup consumes part of it.

When budget reaches zero

Search stops.

---

## Benefits

Lower Redis traffic

Lower allocations

Lower heap size

Lower ranking cost

Lower latency

---

## Acceptance Criteria

Search stops exactly when budget satisfied.

No unnecessary Redis requests.

No unnecessary rings searched.

## Budget never becomes negative.

# Stage 3 — Candidate Retrieval Pipeline

## Goal

Transform candidate retrieval from a monolithic function into a
streaming, composable pipeline.

Current implementation tightly couples

- H3 lookup
- cache loading
- filtering
- ranking
- reservation

inside MatchingEngine.

Production systems separate these concerns.

---

# Current Flow

Ride

↓

findCandidateDriverIDs()

↓

GetDrivers()

↓

Filter

↓

Heap

↓

Reserve

↓

Offer

Problems

• MatchingEngine owns too much logic

• Difficult to extend

• Difficult to test

• Difficult to profile

• Every new filter modifies MatchingEngine

• Ranking depends on retrieval implementation

---

# Target Flow

Ride

↓

Candidate Search

↓

Driver Provider

↓

Candidate Pipeline

↓

Ranking

↓

Reservation

↓

Offer

Each stage has exactly one responsibility.

---

# Architecture

                 Ride Request
                       │
                       ▼
             Candidate Search Strategy
                       │
                       ▼
             Candidate Provider
                       │
                       ▼
             Candidate Pipeline
             ├──────────────┐
             ▼              ▼
         Filter         Enrichment
             │              │
             └──────┬───────┘
                    ▼
              Candidate Iterator
                    ▼
                 Ranking
                    ▼
               Reservation
                    ▼
                 Offer Batch

MatchingEngine orchestrates.

Nothing else.

---

# Candidate Object

Current

MatchingEngine constantly converts

Driver

↓

Candidate

↓

Heap Candidate

↓

Offer Candidate

Future

One immutable candidate object flows through
the entire dispatch pipeline.

Example

Candidate

- DriverID
- Lat
- Lng
- Distance
- ETA
- Score
- Availability
- Metadata

Later

Acceptance Probability

Idle Time

Driver Rating

Current Ride Count

can be added without changing
MatchingEngine.

---

# Pipeline Philosophy

Each stage receives

Candidate

↓

returns

Candidate

or

Reject

No stage should know about
other stages.

---

# Step 10 — Streaming Candidate Iterator

Status

⬜ Pending

Priority

★★★★★

---

## Why

Current implementation

collect every driver

↓

filter every driver

↓

rank every driver

↓

offer few drivers

Large cities

1000+

drivers

This wastes

memory

CPU

cache lookups

Instead

yield candidates gradually.

---

## New Flow

Candidate Search

↓

Yield Candidate

↓

Filter

↓

Rank

↓

Reserve

↓

Offer

↓

Next Candidate

Only process
what is actually needed.

---

## Responsibilities

Candidate Provider

Owns

iteration.

Pipeline

Consumes.

Matching Engine

Stops iteration
once enough drivers are reserved.

---

## Benefits

Constant memory usage.

Early termination.

Lower GC pressure.

Faster dispatch.

---

## Acceptance Criteria

No "collect all first."

Pipeline stops immediately
after enough successful reservations.

---

# Step 11 — Candidate Filtering Pipeline

Status

⬜ Pending

Priority

★★★★★

---

## Why

Current code

if unavailable

continue

if offered

continue

if reserved

continue

...

This logic grows forever.

Future filters

driver suspension

battery

network quality

vehicle type

ETA

fraud

will explode MatchingEngine.

---

## Target

Candidate

↓

Availability Filter

↓

Already Offered Filter

↓

Reservation Filter

↓

Policy Filter

↓

ETA Filter

↓

Ranking

Each filter

independent

testable

replaceable.

---

## Standard Interface

Each filter receives

Candidate

↓

Decision

Accept

Reject

Reject Reason

No filter modifies
other filters.

---

## Benefits

Adding a new filter

=

new file.

No MatchingEngine modification.

---

## Acceptance Criteria

Each filter isolated.

MatchingEngine unaware
of filtering logic.

---

# Step 12 — Batch Cache Fetch

Status

⬜ Pending

Priority

★★★★★

---

## Why

Candidate search may return

200

driver ids.

Fetching individually

Redis

↓

Redis

↓

Redis

↓

Redis

...

kills latency.

---

## Target

Batch

MGET

↓

Decode

↓

Candidate objects

One Redis trip.

---

## Current State

DriverCache already supports
batch retrieval.

Good.

However

MatchingEngine currently owns

conversion

distance calculation

candidate creation

Move this responsibility
into Candidate Provider.

---

## Candidate Provider Responsibilities

Input

Driver IDs

Output

Candidate Iterator

Internally

Batch cache fetch

↓

Distance enrichment

↓

Candidate construction

MatchingEngine never sees raw Driver objects.

---

## Acceptance Criteria

MatchingEngine no longer depends on DriverCache.

MatchingEngine receives Candidates only.

---

# Refactoring Targets

MatchingEngine

Responsibilities after Stage 3

✓ Invoke Search Strategy

✓ Iterate Candidates

✓ Reserve Drivers

✓ Publish Offers

Nothing else.

---

Candidate Search Strategy

Responsibilities

✓ Progressive ring expansion

✓ Search budget

✓ Density adaptation

---

Candidate Provider

Responsibilities

✓ Redis batch fetch

✓ Candidate creation

✓ Distance calculation

✓ Streaming iterator

---

Candidate Pipeline

Responsibilities

✓ Candidate filters

✓ Candidate enrichment

✓ Reject reasons

---

DriverCache

Responsibilities

✓ Batch retrieval

Only cache.

Never business logic.

---

# Success Criteria

By the end of Stage 3

MatchingEngine should become approximately
40–60% smaller.

Adding a new dispatch filter
should require

one new file

and

zero MatchingEngine changes.

Candidate retrieval should support
constant-memory streaming
instead of collect-all processing.

<!-- implement Stage 3 immediately after Step 7 (Progressive Ring Expansion), rather than waiting until all of Stage 2 is complete.

The reason is structural: Progressive Ring Expansion (Step 7) changes how candidates are discovered, while Stage 3 changes how those candidates flow through the system. Once the streaming candidate pipeline exists, the remaining Stage 2 features—density-aware expansion (Step 8) and search budget (Step 9)—become much easier to implement because they can simply decide when to stop producing candidates instead of working with large pre-collected slices. This sequencing also avoids reworking the matching flow twice -->

---

# Dispatch Engine Final Architecture

This section defines the long-term architecture that all future
implementations must follow.

It acts as the contract for the dispatch system.

Future features must extend these components instead of introducing
new orchestration inside MatchingEngine.

---

# High-Level Dispatch Flow

                        Ride Request
                              │
                              ▼
                  Dispatch Orchestrator
                              │
               ┌──────────────┴──────────────┐
               ▼                             ▼
      Candidate Search Strategy      Retry Strategy
               │
               ▼
       Candidate Search Engine
               │
               ▼
        Candidate Provider
               │
               ▼
        Candidate Pipeline
               │
               ▼
          Ranking Engine
               │
               ▼
       Reservation Manager
               │
               ▼
          Offer Dispatcher
               │
               ▼
      Driver Response Handler
               │
               ▼
        Retry / Assignment

MatchingEngine should eventually become
DispatchOrchestrator.

It should contain almost no business logic.

Its responsibility is orchestration only.

---

# Component Responsibilities

## Dispatch Orchestrator

Owns

✓ dispatch workflow

✓ calling components

✓ stopping conditions

✓ transaction boundaries

Never owns

× H3

× Redis

× Ranking

× Candidate filtering

× ETA

---

## Candidate Search Strategy

Responsible for

• progressive ring expansion

• density adaptation

• search budget

• stopping policy

Input

Ride

RetryDecision

Output

Candidate Search Plan

---

## Candidate Search Engine

Responsible for

• H3 traversal

• Redis lookup

• progressive expansion

• multi-resolution search (future)

Never ranks drivers.

Never filters drivers.

---

## Candidate Provider

Responsible for

• cache batch loading

• candidate construction

• location enrichment

• metadata enrichment

Input

Driver IDs

Output

Candidate Iterator

---

## Candidate Pipeline

Responsible for

independent filters

Availability

↓

Already Offered

↓

Reservation

↓

Vehicle Capability

↓

Policy

↓

ETA

↓

Custom Rules

Each filter

one file

one responsibility

---

## Ranking Engine

Responsible for

calculating

Dispatch Score

using

ETA

Quality

Idle Time

Demand

Acceptance Probability

etc.

Ranking should know nothing about Redis.

---

## Reservation Manager

Responsible for

driver locking

reservation release

reservation timeout

future distributed reservation

MatchingEngine never manipulates Redis directly.

---

## Offer Dispatcher

Responsible for

offer creation

offer persistence

outbox

websocket

push notification

future SMS fallback

---

## Driver Response Handler

Responsible for

accept

reject

timeout

disconnect

late response

Everything after an offer has been sent.

---

# Data Objects

Dispatch should operate on immutable objects.

RideRequest

↓

SearchPlan

↓

Candidate

↓

RankedCandidate

↓

ReservedCandidate

↓

DriverOffer

Every stage transforms one object into another.

Avoid modifying shared mutable objects.

---

# Dependency Rules

Allowed

Dispatch Orchestrator

↓

Search Strategy

↓

Search Engine

↓

Provider

↓

Pipeline

↓

Ranking

↓

Reservation

↓

Offer

Not Allowed

Search Strategy

↓

Redis

Pipeline

↓

H3

Ranking

↓

Redis

Reservation

↓

Ranking

Offer

↓

Matching

Every dependency points downward.

---

# Extension Points

The architecture should allow adding new behavior without modifying
existing orchestrators.

Examples

Future

Surge-aware filter

↓

new CandidateFilter

No MatchingEngine changes.

Future

ML Ranking

↓

new Ranker

No MatchingEngine changes.

Future

New Search Algorithm

↓

new SearchStrategy

No MatchingEngine changes.

Future

Google ETA

↓

new ETAProvider

No MatchingEngine changes.

---

# Testing Strategy

Each component should be testable independently.

Candidate Search

↓

unit tests

Candidate Pipeline

↓

unit tests

Ranking

↓

unit tests

Reservation

↓

integration tests

Offer Dispatcher

↓

integration tests

Dispatch Orchestrator

↓

end-to-end tests

---

# Performance Goals

Ride Dispatch

Target

<150 ms

Candidate Lookup

<20 ms

Ranking

<10 ms

Reservation

<20 ms

Offer Creation

<10 ms

Redis Calls

Minimum possible.

No unnecessary allocations.

No duplicated candidate processing.
