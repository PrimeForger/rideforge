# `docs/architecture/DISCOVERY_ENGINE_ARCHITECTURE.md`

# Part 1 — Overview, Philosophy & High-Level Architecture

---

# Discovery Engine Architecture

**Project:** RideForge Ride-Hailing Platform

**Module:** Dispatch Discovery Engine

**Status:** Production Architecture

**Document Version:** 1.0

---

# Table of Contents

1. Introduction
2. Why This Module Exists
3. High-Level Responsibilities
4. Design Philosophy
5. Core Design Principles
6. High-Level Architecture
7. Complete Discovery Flow
8. Package Overview
9. Dependency Rules
10. Architectural Decisions
11. Why We Didn't Choose Simpler Designs

---

# 1. Introduction

The Discovery Engine is responsible for finding the most suitable drivers for a ride request.

It is **not** responsible for:

- dispatching offers
- driver locking
- ETA calculation
- ranking
- reservation
- surge pricing
- ride assignment

Its responsibility begins with:

```
Ride Request
```

and ends with

```
Candidate Driver IDs
```

Everything after candidate discovery belongs to other modules.

This strict separation allows each module to evolve independently while maintaining a clean architecture.

---

# 2. Why This Module Exists

A naive ride-hailing system typically performs the following steps:

```
Ride Request

↓

Search nearby drivers

↓

Return all drivers

↓

Filter

↓

Dispatch
```

While this approach is simple, it suffers from several problems:

- unnecessary Redis traffic
- unnecessary H3 cell lookups
- large memory allocations
- poor extensibility
- tightly coupled business logic
- difficult testing
- poor scalability

As the system grows, new requirements emerge:

- urban vs rural behavior
- retry-aware search
- airport-specific search
- premium ride search
- marketplace-aware search
- ML-driven search profiles

If these behaviors are implemented directly inside the matching engine, the code quickly becomes difficult to maintain.

The Discovery Engine exists to isolate search behavior into a dedicated subsystem with well-defined responsibilities and extension points.

---

# 3. High-Level Responsibilities

The Discovery Engine is responsible for:

- locating candidate drivers
- expanding search progressively
- controlling search budgets
- selecting search policies
- adapting search behavior based on context
- exposing candidate driver IDs to downstream modules

It deliberately avoids business concerns that belong elsewhere.

For example:

**Not Responsible**

- driver ranking
- ETA estimation
- driver reservation
- dispatch offers
- acceptance handling
- ride assignment

These are handled by dedicated modules.

---

# 4. Design Philosophy

The architecture follows one fundamental principle:

> **Configuration decides behavior. Execution follows configuration.**

Instead of hardcoding dispatch logic:

```
if dense {
    ...
}

if retry {
    ...
}
```

the system constructs a search policy and executes it.

The execution engine should never know _why_ a particular strategy was selected.

It only executes the selected policy.

---

# 5. Core Design Principles

## 5.1 Single Responsibility Principle

Every component owns exactly one responsibility.

Examples:

| Component       | Responsibility          |
| --------------- | ----------------------- |
| Selector        | Select search profile   |
| Pipeline        | Execute rules           |
| Rule            | Modify builder          |
| Builder         | Construct SearchProfile |
| ExpansionPolicy | Expansion behavior      |
| RingExpander    | Execute ring search     |
| H3Strategy      | H3-based search         |
| GeoStrategy     | Redis GEO search        |

No component performs multiple unrelated tasks.

---

## 5.2 Open for Extension

The system is designed so that future behavior can be added without modifying existing code.

Future additions may include:

```
AirportRule

NightModeRule

MarketplaceRule

PremiumRideRule

SurgeRule

MLRule
```

Each becomes a new rule inside the pipeline.

Existing components remain unchanged.

---

## 5.3 Dependency Inversion

High-level modules never depend on infrastructure implementations.

Example:

```
MatchingEngine

↓

CandidateSearcher

↓

H3Strategy
```

The matching engine has no knowledge of Redis, H3, or search implementation details.

---

## 5.4 Composition over Conditionals

Instead of:

```
if urban {
    ...
}

if airport {
    ...
}

if retry {
    ...
}
```

the system composes behavior through:

```
Pipeline

↓

Rules

↓

SearchProfile
```

Behavior is assembled rather than hardcoded.

---

## 5.5 Immutable Search Profiles

A SearchProfile is treated as immutable once constructed.

This provides:

- predictable behavior
- thread safety
- easier testing
- easier debugging

All mutation occurs during profile construction.

After construction, the profile is read-only.

---

# 6. High-Level Architecture

```
                        Ride Request
                              │
                              ▼
                     Matching Engine
                              │
                              ▼
                   CandidateSearcher
                              │
            ┌─────────────────┴─────────────────┐
            │                                   │
            ▼                                   ▼
      H3 Strategy                        GEO Strategy
            │
            ▼
      Ring Expander
            │
            ▼
 Adaptive Ring Expansion Policy
            │
            ▼
     Search Profile Selector
            │
            ▼
      Profile Pipeline
            │
            ▼
        Search Rules
            │
            ▼
     Search Profile Builder
            │
            ▼
       Search Profile
            │
            ▼
      Expansion Policy
            │
            ▼
      Progressive Ring Search
            │
            ▼
      Candidate Driver IDs
```

Notice that every layer performs exactly one task.

---

# 7. Complete Discovery Flow

A complete search request flows through the system as follows:

```
Ride Request

↓

Matching Engine

↓

Candidate Searcher

↓

H3 Strategy

↓

Center H3 Cell

↓

Ring Expander

↓

Adaptive Policy

↓

Selector

↓

Pipeline

↓

Density Rule

↓

Builder

↓

Search Profile

↓

Expansion Policy

↓

Ring Decision

↓

Lookup Drivers

↓

Budget Check

↓

Next Ring

↓

Enough Candidates?

↓

Yes

↓

Return Driver IDs
```

Every step has a single responsibility.

---

# 8. Package Overview

```
dispatch/
└── discovery/

    contract/
        Shared contracts and enums.

    density/
        Driver density abstraction and classification.

    expansion/
        Progressive search execution.

    lookup/
        Driver lookup abstraction.

    pipeline/
        Executes search rules.

    profile/
        Immutable search policies.

    rule/
        Individual policy rules.

    search/
        Search request state.

    selector/
        Search profile construction.

    strategy/
        H3 / GEO search strategies.
```

Every package represents one concept.

Packages do not overlap responsibilities.

---

# 9. Dependency Rules

Dependencies always move downward.

```
MatchingEngine

↓

Strategy

↓

RingExpander

↓

Policy

↓

Selector

↓

Pipeline

↓

Rules

↓

Builder

↓

SearchProfile
```

The reverse is never allowed.

For example:

**Allowed**

```
Rule

↓

Builder
```

**Forbidden**

```
Builder

↓

Rule
```

This keeps the architecture acyclic.

---

# 10. Architectural Decisions

Several important architectural decisions were intentionally made during development.

## Decision 1

Search behavior is separated from matching.

Reason:

Candidate discovery and dispatch evolve independently.

---

## Decision 2

Policies are immutable.

Reason:

Immutable objects reduce bugs and simplify concurrent execution.

---

## Decision 3

Behavior is composed.

Reason:

Composition scales better than nested conditionals.

---

## Decision 4

Execution and configuration are separated.

Reason:

Configuration should decide behavior.

Execution should only execute.

---

## Decision 5

Rules modify builders.

Reason:

Rules should not mutate immutable objects directly.

Builders exist specifically for controlled construction.

---

# 11. Why We Didn't Choose Simpler Designs

Several simpler architectures were considered.

## Option A

```
MatchingEngine

↓

Redis
```

Rejected because all search logic becomes tightly coupled.

---

## Option B

```
MatchingEngine

↓

H3 Service

↓

Redis
```

Rejected because business rules accumulate inside H3 search.

---

## Option C

```
Selector

↓

switch density

↓

switch retry

↓

switch airport
```

Rejected because selectors become God Objects over time.

---

## Option D

```
ExpansionProfile

↓

Expose MaxRing

↓

Policy reads configuration
```

Rejected because execution becomes coupled to configuration.

Instead:

```
ExpansionPolicy

↓

ConfigureBudget()

↓

NextRing()
```

The policy exposes behavior rather than data.

---

# Summary

The Discovery Engine is designed as a layered, extensible subsystem that isolates candidate discovery from the rest of the dispatch workflow. Every component has a single responsibility, dependencies flow in one direction, and behavior is composed through immutable profiles and policy execution rather than hardcoded conditionals.

This foundation allows future capabilities—such as retry-aware search, airport-specific behavior, surge-aware expansion, regional policies, and machine learning—to be added as independent extensions without requiring structural changes to the core discovery engine.

---

## Part 2 — Discovery Pipeline, Search Profiles & Progressive Ring Expansion

- Table of Contents
  Discovery Pipeline
  Search Strategies
  Progressive Ring Expansion
  Search State
  Search Budget
  Ring Expansion Policy
  Search Profile System
  Rule Pipeline
  Expansion Policy
  Sequence Diagram
  Design Rationale

1. Discovery Pipeline

The Discovery Engine is composed of multiple small layers rather than one large search function.

Ride Request
│
▼
Matching Engine
│
▼
Candidate Searcher
│
▼
Search Strategy
│
▼
Ring Expander
│
▼
Expansion Policy
│
▼
Selector
│
▼
Pipeline
│
▼
Rules
│
▼
Search Profile
│
▼
Driver Lookup
│
▼
Candidate Drivers

Every layer performs a single responsibility.

No layer performs business logic that belongs elsewhere.

2. Search Strategies

The Discovery Engine supports multiple search strategies.

Current implementations:

CandidateSearcher

        │

        ├──────────────┐
        ▼              ▼

H3Strategy GeoStrategy

The Matching Engine never knows which implementation is being used.

Instead it depends on

CandidateSearcher

This allows future strategies to be introduced without modifying Matching Engine.

Possible future strategies include:

RegionStrategy

CachedStrategy

HybridStrategy

MarketplaceStrategy

MultiNodeStrategy

Each strategy remains isolated.

H3 Strategy

The H3 strategy performs candidate discovery using hierarchical hexagonal indexing.

Responsibilities:

Determine center H3 cell
Start progressive expansion
Return candidate driver IDs

The H3 strategy does not decide:

search policy
density
expansion profile
search limits

Those responsibilities belong elsewhere.

GEO Strategy

The GEO strategy performs candidate discovery using Redis GEO.

Responsibilities:

Radius search
Candidate lookup

It intentionally mirrors the H3 strategy interface.

This allows both implementations to be interchangeable.

3. Progressive Ring Expansion

Traditional implementations search

GridDisk(maxRing)

which retrieves every cell immediately.

Example:

Ring 0

Ring 1

Ring 2

Ring 3

This performs unnecessary Redis lookups when nearby drivers already satisfy the request.

The Discovery Engine instead performs progressive expansion.

Ring 0

↓

Enough candidates?

↓

YES

Stop

↓

NO

↓

Ring 1

↓

Enough?

↓

YES

Stop

↓

NO

↓

Ring 2

Search expands only when necessary.

This dramatically reduces:

Redis commands
H3 computations
memory usage
latency 4. Search State

Every search request owns an independent SearchState.

SearchState represents the current execution state of the search.

Typical contents include:

Center Cell

Current Ring

Visited Rings

Visited Cells

Driver IDs

Seen Drivers

Search Budget

Last Density

Termination Reason

SearchState is mutable.

It exists only during one search request.

It is never shared.

Why SearchState Exists

Without SearchState every layer would need dozens of parameters.

Example:

CenterCell

CurrentRing

VisitedCells

VisitedRings

CandidateLimit

RemainingBudget

Density

...

Instead

SearchState

acts as the execution context for the search.

5. Search Budget

Searching should not continue indefinitely.

The Discovery Engine therefore owns a dedicated SearchBudget.

Current budgets include:

Candidate Budget

Ring Budget

Cell Budget

Example:

Need

40 drivers

Current result

25 drivers

Remaining candidate budget

15

The next lookup requests only the remaining budget.

Instead of

Redis

↓

50 drivers

↓

Discard 35

the system requests

Redis

↓

15 drivers

This minimizes unnecessary work.

Ring Budget

Ring Budget prevents excessive expansion.

Example

Maximum Ring

4

After Ring 4

Search Terminates

No additional H3 lookups occur.

Cell Budget

Searching many H3 cells can become expensive.

The Cell Budget limits the total number of visited cells.

This protects the system against excessive searches in sparse regions.

6. Ring Expansion Policy

The Ring Expander executes searches.

The Ring Expansion Policy decides what happens next.

Responsibilities of the policy:

terminate search
choose next ring
configure budget
evaluate search profile

The policy never performs Redis lookups.

The expander performs execution.

This separation is intentional.

Policy

↓

Decision

↓

Expander

↓

Execution 7. Search Profile System

The Search Profile represents the policy selected for one search request.

It is immutable.

Example

Urban Profile

Airport Profile

Rural Profile

Night Profile

Each profile can configure different search behavior.

Current profile contents

Expansion Policy

Future additions

Retry Policy

Ranking Policy

Offer Policy

Routing Policy

SearchProfile is intentionally designed for future growth.

Why Immutable Profiles

Mutable policy objects easily become difficult to reason about.

Instead

Rules

↓

Builder

↓

Immutable Profile

After construction the profile never changes.

8. Rule Pipeline

Profiles are constructed through a pipeline.

Selector

↓

Pipeline

↓

Rules

↓

Builder

↓

SearchProfile

Each rule contributes one concern.

Current rule:

DensityRule

Future rules:

RetryRule

MarketplaceRule

AirportRule

HotRegionRule

MLRule

Rules never know about one another.

They communicate only through the Builder.

Why a Pipeline?

Without a pipeline

Selector

↓

if density

↓

if airport

↓

if retry

↓

if marketplace

The selector eventually becomes hundreds of lines long.

With the pipeline

Selector

↓

Pipeline

↓

Rule

↓

Rule

↓

Rule

Each rule remains independently testable.

9. Expansion Policy

The Expansion Policy owns expansion behavior.

Responsibilities:

configure search budgets
calculate next ring

Consumers never inspect internal configuration.

Instead

ExpansionPolicy.ConfigureBudget(...)

and

ExpansionPolicy.NextRing(...)

are used.

This hides implementation details.

Why Behavior Instead of Configuration

Old design

Policy

↓

ExpansionProfile

↓

MaxRing

↓

RingStep

↓

MaxCells

Consumers directly inspected configuration.

Current design

Policy

↓

ExpansionPolicy

↓

ConfigureBudget()

↓

NextRing()

Consumers execute behavior instead of reading configuration.

This reduces coupling.

10. Complete Sequence Diagram
    Ride Request

↓

Matching Engine

↓

CandidateSearcher

↓

H3 Strategy

↓

Center Cell

↓

Ring Expander

↓

Adaptive Policy

↓

Selector

↓

Pipeline

↓

Density Rule

↓

Density Provider

↓

Density Classifier

↓

Expansion Profile Provider

↓

Builder

↓

Search Profile

↓

Expansion Policy

↓

Configure Budget

↓

Next Ring

↓

Driver Lookup

↓

Candidate IDs

↓

Enough Candidates?

↓

YES

↓

Return

Every step owns exactly one responsibility.

11. Design Rationale

The Discovery Engine intentionally separates:

Configuration

Expansion Profile

from

Behavior

Expansion Policy

and separates

Decision Making

Selector

Pipeline

Rules

from

Execution

Ring Expander

This architecture follows a consistent pattern throughout the module:

Configuration

↓

Builder

↓

Immutable Policy

↓

Execution

As new requirements emerge—such as retry-aware searches, airport-specific behavior, surge-aware expansion, regional dispatch, or ML-driven policies—the system can be extended by introducing new rules or policies without modifying the existing execution flow.

## Part 3 — H3 Discovery, Redis Integration, Dependency Injection & Bootstrap

- Table of Contents
  Introduction
  H3-Based Candidate Discovery
  Redis Responsibilities
  Candidate Lookup
  Dependency Injection
  Bootstrap Composition
  Runtime Object Graph
  Discovery Module Boundaries
  Error Handling Strategy
  Performance Considerations
  Testing Strategy
  Future Scalability

1. Introduction

The previous sections explained how the Discovery Engine makes decisions.

This section explains how those decisions are executed.

Specifically:

H3 search
Redis lookup
dependency injection
runtime composition
module boundaries
testing strategy

This document intentionally separates decision-making from execution.

2. H3-Based Candidate Discovery

The Discovery Engine currently uses H3 as its primary spatial index.

High-level flow:

Ride Request
│
▼
Pickup Latitude / Longitude
│
▼
Center H3 Cell
│
▼
Progressive Ring Expansion
│
▼
Redis Cell Lookup
│
▼
Candidate Driver IDs

Notice that H3 itself is not responsible for candidate selection.

Its responsibility ends after identifying the H3 cells that should be searched.

Why H3?

The Discovery Engine uses H3 because:

deterministic spatial partitioning
efficient neighborhood traversal
hierarchical indexing
predictable expansion
excellent scalability

H3 dramatically reduces the search space before Redis is queried.

3. Redis Responsibilities

Redis is treated as an execution engine, not a decision engine.

Redis never decides:

which ring to search
whether expansion should continue
candidate limits
density policies
search budgets

Redis only executes lookup requests.

Responsibilities include:

H3 cell membership
driver ID retrieval
driver density counts
driver location cache
Example

Instead of:

Redis

↓

Find nearest drivers

the Discovery Engine performs:

Policy

↓

Ring Decision

↓

Ring Expander

↓

Redis Lookup

Redis executes.

Policies decide.

4. Candidate Lookup

The lookup layer isolates Redis from the search engine.

Ring Expander
│
▼
CellDriverLookup
│
▼
Redis

This abstraction provides several benefits.

Isolation

The Ring Expander never knows Redis commands.

It only knows:

GetDriversInCells(...)
Testability

Unit tests can inject:

FakeLookup

instead of Redis.

This allows deterministic tests without infrastructure.

Future Compatibility

Future implementations could include:

Redis Lookup

In-Memory Lookup

Regional Lookup

Distributed Lookup

without changing the Ring Expander.

5. Dependency Injection

The Discovery Engine uses constructor injection exclusively.

Objects never instantiate their own dependencies.

Example:

MatchingEngine

↓

CandidateSearcher

↓

H3Strategy

↓

RingExpander

↓

AdaptivePolicy

↓

Selector

↓

Pipeline

↓

Rules

Dependencies always flow downward.

Why Constructor Injection?

Benefits include:

explicit dependencies
easier testing
no hidden state
immutable object graphs
simpler reviews

No service performs:

redis.New(...)

internally.

All dependencies are injected from the composition root.

6. Bootstrap Composition

The composition root is responsible for assembling the Discovery Engine.

Example flow:

Redis Client
│
▼
Density Provider
│
▼
Density Classifier
│
▼
Expansion Profile Provider
│
▼
Density Rule
│
▼
Profile Pipeline
│
▼
Selector
│
▼
Adaptive Ring Expansion Policy
│
▼
Ring Expander
│
▼
H3 Strategy
│
▼
CandidateSearcher
│
▼
MatchingEngine

Each dependency is created exactly once.

Why Bootstrap Owns Composition

Business objects should never decide:

which implementation to use
how dependencies are wired
how policies are assembled

Those decisions belong exclusively to the composition root.

7. Runtime Object Graph

The runtime object graph currently resembles:

MatchingEngine
│
▼
CandidateSearcher
│
▼
H3Strategy
│
▼
RingExpander
│
▼
AdaptiveRingExpansionPolicy
│
▼
Selector
│
▼
Pipeline
│
▼
DensityRule
│
▼
DensityProvider
│
▼
Redis

Notice that every object has one parent.

No circular dependencies exist.

8. Discovery Module Boundaries

The Discovery Engine intentionally exposes a very small API.

External modules should only depend on:

CandidateSearcher

Everything below it is internal implementation.

Modules outside Discovery should never know:

H3
SearchProfile
ExpansionPolicy
DensityRule
Pipeline
Selector

Those are private implementation details.

Stable Boundary

The Discovery Engine presents one stable contract:

FindCandidates(...)

Everything else may evolve without affecting consumers.

9. Error Handling Strategy

Errors propagate upward without being swallowed.

Example:

Redis Error

↓

Lookup

↓

Ring Expander

↓

H3 Strategy

↓

CandidateSearcher

↓

MatchingEngine

No intermediate layer silently ignores infrastructure failures.

Why?

Silent failures create:

incomplete candidate lists
inconsistent dispatch behavior
difficult debugging

The Discovery Engine favors explicit failures over hidden degradation.

10. Performance Considerations

The architecture was designed with performance in mind.

Progressive Expansion

Search expands only when necessary.

Instead of:

GridDisk(5)

the engine performs:

Ring 0

↓

Ring 1

↓

Ring 2

This significantly reduces Redis traffic.

Search Budget

Only the remaining candidate budget is requested.

Example:

Need

50

Already found

38

Next lookup requests only:

12

instead of another full search.

Immutable Policies

Search profiles are immutable.

This removes:

synchronization concerns
accidental mutations
hidden side effects
Small Components

Every object owns one responsibility.

Small components:

allocate less memory
are easier to optimize
are easier to benchmark
are easier to replace 11. Testing Strategy

The Discovery Engine is designed for layered testing.

Unit Tests

Each component can be tested independently.

Examples:

DensityRule

AdaptiveRingExpansionPolicy

RingExpander

Selector

Builder

No Redis required.

Integration Tests

Integration tests validate:

Redis Lookup

H3 Search

Progressive Expansion

Candidate Retrieval

These tests verify infrastructure behavior.

End-to-End Tests

Complete flow:

Ride Request

↓

Candidate Search

↓

Dispatch

These tests validate the complete dispatch pipeline.

12. Future Scalability

The architecture intentionally prepares for future features.

Possible future additions include:

AirportRule

RetryRule

SurgeRule

MarketplaceRule

VIPRule

MLRule

RegionalRule

Each becomes another pipeline stage.

No existing execution logic changes.

Distributed Discovery

The current architecture can evolve toward:

MatchingEngine
│
▼
Regional Candidate Searcher
│
├──────────┐
▼ ▼

Region A Region B

without changing the Discovery Engine contract.

Multi-Resolution H3

The architecture also supports future multi-resolution H3 indexing.

For example:

Resolution 9

↓

Not enough drivers

↓

Resolution 8

↓

Resolution 7

This enhancement can be introduced within the H3 strategy layer without affecting:

Ring Expander
Search Profiles
Rules
Selector
Pipeline

The separation between policy and execution makes this evolution straightforward.

Summary

The Discovery Engine is built around a layered execution model where decision-making and execution are deliberately separated.

H3 identifies spatial search regions.
Redis executes data retrieval.
Ring Expander performs the search.
Adaptive Policy decides how the search progresses.
Selector, Pipeline, and Rules construct immutable search policies.
Dependency Injection assembles the object graph at the composition root.

This structure produces a highly testable, extensible, and maintainable subsystem that is prepared for future enhancements such as adaptive retry policies, regional dispatch, hot-region optimization, and multi-resolution H3 search without requiring significant architectural changes.

## Part 4 — Extension Guide, Development Standards, Anti-Patterns & Future Evolution

- Table of Contents
  Introduction
  Extension Philosophy
  Adding a New Search Rule
  Adding a New Search Strategy
  Adding a New Search Profile
  Adding a New Expansion Policy
  Dependency Rules
  Development Standards
  Anti-Patterns
  Code Review Checklist
  Performance Guidelines
  Future Evolution
  Final Architecture Summary

1. Introduction

The Discovery Engine was intentionally designed to evolve.

Most ride-hailing systems continue to grow for years:

new dispatch algorithms
better routing
surge optimization
ML recommendations
regional dispatch
airport behavior

The architecture should allow those additions without rewriting the existing system.

This document explains how future development should happen.

2. Extension Philosophy

Every new feature should answer one question first:

Which responsibility is changing?

Never ask:

"Where can I put this code?"

Instead ask:

What responsibility does this feature belong to?

Examples:

Feature Belongs To
Search radius policy Expansion Policy
Driver density Density Rule
Airport behavior New Rule
Retry behavior New Rule
Candidate retrieval Lookup
Redis optimization Lookup
Multi-resolution H3 H3 Strategy
Ranking Ranking Engine
Driver reservation Dispatch Module

Never add unrelated responsibilities to an existing component.

3. Adding a New Search Rule

Search rules are the preferred extension mechanism.

Current pipeline:

Pipeline

↓

DensityRule

Future:

Pipeline

↓

DensityRule

↓

RetryRule

↓

AirportRule

↓

MarketplaceRule

↓

HotRegionRule

↓

MLRule

Every rule should own one concern.

Rule Checklist

A rule should:

inspect SearchState
inspect external providers if required
modify Builder
return an error if it cannot complete

A rule should not:

execute Redis lookups unrelated to its responsibility
execute H3 searches
dispatch rides
rank drivers
mutate SearchProfile directly

Rules only modify the Builder.

Example
RetryRule

↓

Read retry count

↓

Configure retry profile

↓

Done

No additional responsibilities.

4. Adding a New Search Strategy

The strategy layer changes how candidates are discovered.

Current:

CandidateSearcher

↓

H3Strategy

↓

GeoStrategy

Future:

CandidateSearcher

↓

H3Strategy

↓

GeoStrategy

↓

HybridStrategy

↓

RegionalStrategy

↓

CachedStrategy

A strategy decides:

where candidates come from

It does not decide:

search policy
retry
expansion
budgets

Those belong elsewhere.

5. Adding a New Search Profile

SearchProfile represents immutable policy.

Future example:

SearchProfile

├── ExpansionPolicy

├── RetryPolicy

├── RankingPolicy

├── OfferPolicy

└── RoutingPolicy

Each policy should remain independent.

Avoid creating:

MegaSearchProfile

where one object owns unrelated behavior.

Builder Guidelines

Only the Builder constructs SearchProfile.

Never construct profiles manually.

Incorrect:

SearchProfile{
...
}

Correct:

Builder

↓

Build() 6. Adding a New Expansion Policy

Expansion policies encapsulate expansion behavior.

Future examples:

DefaultExpansionPolicy

UrbanExpansionPolicy

AirportExpansionPolicy

SurgeExpansionPolicy

MLExpansionPolicy

Every expansion policy must implement the same contract.

The consumer must never care which implementation it receives.

Responsibilities

ExpansionPolicy owns:

configuring budgets
selecting next ring
expansion behavior

It should never:

query Redis
classify density
retrieve drivers

Those responsibilities belong elsewhere.

7. Dependency Rules

Dependencies always move downward.

MatchingEngine

↓

Searcher

↓

Strategy

↓

Expander

↓

Policy

↓

Selector

↓

Pipeline

↓

Rule

↓

Builder

↓

SearchProfile

No layer should import upward.

Allowed
Rule

↓

Builder
Forbidden
Builder

↓

Rule
Allowed
Strategy

↓

Expander
Forbidden
Expander

↓

Strategy

This keeps the dependency graph acyclic.

8. Development Standards

Every new component should satisfy these principles.

Single Responsibility

One class.

One reason to change.

Constructor Injection

Never instantiate dependencies internally.

Incorrect:

func NewRule() {

    redis.New(...)

}

Correct:

func NewRule(
provider DriverDensityProvider,
)
Interfaces

Depend on behavior.

Never on implementation.

Correct:

DriverDensityProvider

Incorrect:

RedisDensityProvider
Immutability

SearchProfile should remain immutable.

Construction happens through Builder.

Execution never mutates profiles.

Explicit Composition

Bootstrap owns object construction.

Business objects never decide:

implementation
dependency graph
configuration 9. Anti-Patterns

The following patterns are intentionally forbidden.

God Objects

Incorrect:

Selector

↓

Everything

Selector should orchestrate.

Nothing more.

Nested Conditionals

Avoid:

if retry {

    if airport {

        if dense {

Instead:

Pipeline

↓

Rules
Infrastructure Leakage

Business objects should never know Redis commands.

Incorrect:

SMEMBERS

inside policy.

Correct:

CellDriverLookup
Mutable Global State

Never share SearchState.

Each search request owns its own state.

Circular Dependencies

Never introduce:

Rule

↓

Profile

↓

Rule

Keep dependencies one-directional.

10. Code Review Checklist

Every Discovery Engine pull request should answer:

Responsibilities
Does this component own one responsibility?
Dependencies
Does dependency direction remain downward?
Testability
Can this component be unit tested?
Extensibility
Could another implementation be introduced later?
Coupling
Does this introduce unnecessary knowledge?
Performance
Does this reduce unnecessary Redis work?
Immutability
Is SearchProfile still immutable?
Bootstrap
Is composition performed only in bootstrap? 11. Performance Guidelines

Discovery performance should prioritize:

Minimize Redis Round Trips

Prefer:

Pipeline

↓

One network trip

over repeated lookups.

Progressive Expansion

Never retrieve more cells than necessary.

Budget Driven Search

Never request more candidates than required.

Stateless Rules

Rules should not maintain mutable state.

This improves concurrency.

Small Objects

Prefer many focused objects over large multipurpose classes.

12. Future Evolution

The Discovery Engine was designed to support future stages without architectural changes.

Stage 3

Candidate Retrieval Pipeline

Lookup

↓

Availability

↓

Reservation

↓

Ranking
Stage 4

Ranking Engine

SearchProfile can later include:

RankingPolicy

No Discovery refactor required.

Stage 5

Dispatch Optimizer

Retry-aware profiles become another rule.

Stage 6

Geo Engine

Multi-resolution H3

Routing providers

Distance refinement

No Discovery refactor required.

Stage 7

Hot Region Scaling

Add:

HotRegionRule

Pipeline remains unchanged.

Stage 8

Regional Dispatch

Introduce:

RegionalStrategy

No MatchingEngine changes.

Stage 9

Reliability

Background reconciliation

Index rebuilding

Search architecture unchanged.

Stage 10

Observability

Each layer becomes independently instrumented.

Example:

Strategy

↓

Policy

↓

Lookup

↓

Pipeline

Each exposes metrics and traces.

Stage 11

Machine Learning

Pipeline becomes:

DensityRule

↓

RetryRule

↓

MLRule

No architectural change required.

Only another rule.

13. Final Architecture Summary

The Discovery Engine follows a layered architecture where each component owns exactly one responsibility.

Ride Request
│
▼
Matching Engine
│
▼
CandidateSearcher
│
▼
Search Strategy
│
▼
Ring Expander
│
▼
Ring Expansion Policy
│
▼
Selector
│
▼
Pipeline
│
▼
Rules
│
▼
Builder
│
▼
Search Profile
│
▼
Expansion Policy
│
▼
Driver Lookup
│
▼
Candidate Driver IDs

The architecture is intentionally designed around composition, immutability, dependency inversion, and single responsibility. New capabilities should be added by introducing new rules, policies, or strategies rather than modifying existing execution paths. This minimizes regression risk, keeps components independently testable, and allows the Discovery Engine to evolve into advanced dispatch features—such as regional search, surge-aware behavior, airport optimization, and ML-assisted decision making—without requiring significant structural refactoring.
