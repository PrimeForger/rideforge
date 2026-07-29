## AI Development Guide — Discovery Engine

- Project: RideForge Ride-Hailing Platform

Module: Dispatch Discovery Engine

Audience: AI Coding Agents (ChatGPT, Claude, Cursor, Copilot, Windsurf, Gemini, etc.)

Version: 1.0

Purpose

This document defines the architectural standards that every AI agent must follow when modifying or extending the Discovery Engine.

This is not an architecture explanation.

It is an engineering contract.

If an AI agent generates code that violates these rules, the implementation should be considered incorrect even if it compiles.

Table of Contents
Module Overview
Core Philosophy
Architecture Rules
Dependency Rules
Package Responsibilities
Extension Rules
Development Standards
Forbidden Patterns
Dependency Injection Rules
Testing Rules
Performance Rules
Naming Standards
Pull Request Checklist
Stable APIs
Future Extension Guidelines

1. Module Overview

The Discovery Engine is responsible only for candidate discovery.

It begins with:

Ride Request

It ends with:

Candidate Driver IDs

Everything else belongs to other modules.

The Discovery Engine must never perform:

Driver ranking
ETA calculation
Offer dispatch
Ride assignment
Driver reservation
Surge pricing
Pricing
Billing

Those belong to separate modules.

2. Core Philosophy

Always remember:

Decision and execution are separate responsibilities.

The engine follows this principle:

Configuration

↓

Policy

↓

Execution

Never mix these responsibilities.

Golden Rule

If adding a feature requires modifying five existing classes,

you are probably adding it incorrectly.

Prefer:

New Rule

or

New Policy

or

New Strategy

instead of modifying existing components.

3. Architecture Rules

The Discovery Engine follows strict layering.

MatchingEngine

↓

CandidateSearcher

↓

Search Strategy

↓

Ring Expander

↓

Ring Expansion Policy

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

↓

ExpansionPolicy

AI agents must preserve this order.

Never reverse dependencies.

4. Dependency Rules

Dependencies always point downward.

Allowed:

Strategy

↓

RingExpander

Forbidden:

RingExpander

↓

Strategy

Allowed:

Rule

↓

Builder

Forbidden:

Builder

↓

Rule

Allowed:

Selector

↓

Pipeline

Forbidden:

Pipeline

↓

Selector

Never introduce circular dependencies.

5. Package Responsibilities
   strategy/

Responsible for:

candidate search algorithm

Not responsible for:

density
retry
ranking
policies
expansion/

Responsible for:

progressive ring execution

Not responsible for:

Redis
SearchProfile construction
selector/

Responsible for:

orchestrating SearchProfile construction

Should remain extremely small.

pipeline/

Responsible for:

executing SearchProfileRules

Nothing else.

rule/

Responsible for:

modifying Builder

Never modify SearchProfile directly.

profile/

Responsible for:

immutable policies

Never execute Redis here.

density/

Responsible for:

density classification

Not responsible for:

ring expansion
lookup/

Responsible for:

retrieving drivers

Never contain policy logic.

search/

Responsible for:

runtime execution state

Should never contain business rules.

6. Extension Rules
   Adding New Search Behavior

Always prefer:

New Rule

Examples:

RetryRule

AirportRule

MarketplaceRule

VIPRule

MLRule

Do NOT add:

if retry {

}

if airport {

}

inside Selector.

Adding New Expansion Logic

Create:

New ExpansionPolicy

Never modify RingExpander.

Adding New Candidate Search

Create:

New Strategy

Never modify MatchingEngine.

Adding New Infrastructure

Create:

New Lookup

or

New Provider

Never leak infrastructure into business logic.

7. Development Standards

Every component should satisfy:

✅ One responsibility

✅ Constructor injection

✅ Interface driven

✅ Independently testable

✅ Immutable where possible

Never introduce:

static globals
mutable shared state
singleton business objects 8. Forbidden Patterns
Forbidden

God Objects

Example:

MatchingEngine

↓

Everything

Forbidden

Nested business conditionals

Example:

if airport {

    if retry {

        if dense {

        }

    }

}

Instead:

Pipeline

↓

Rules

Forbidden

Business logic inside Redis adapters

Forbidden

Business logic inside H3Service

Forbidden

Redis commands inside policies

Forbidden

Reading ExpansionProfile internals directly.

Correct:

policy.ConfigureBudget(...)

Incorrect:

profile.MaxRing 9. Dependency Injection Rules

All dependencies must be injected from bootstrap.

Never do:

redis.New(...)

inside services.

Never instantiate providers inside policies.

Composition belongs only in:

bootstrap/container.go 10. Testing Rules

Every new object must be independently testable.

Preferred test hierarchy:

Unit Test

↓

Integration Test

↓

End-to-End Test

Rules should be testable without Redis.

Policies should be testable without H3.

Builders should require no infrastructure.

11. Performance Rules

Always optimize for:

fewer Redis calls
fewer H3 lookups
fewer allocations
fewer duplicate driver lookups

Never request:

100 Drivers

if only

12

remain in the candidate budget.

Prefer progressive search.

Never expand unnecessarily.

12. Naming Standards

Use nouns for models.

Examples:

SearchProfile

SearchBudget

SearchState

Use verbs for behaviors.

Examples:

Select()

Apply()

Execute()

ConfigureBudget()

NextRing()

Interfaces should describe behavior.

Good:

CandidateSearcher

DriverDensityProvider

CellDriverLookup

Bad:

RedisLookup

RedisProvider 13. Pull Request Checklist

Every PR should answer:

Responsibilities

Does every component still own one responsibility?

Dependencies

Did dependency direction remain downward?

Extensibility

Could another implementation be added without modifying existing code?

Testability

Can every new component be unit tested?

Performance

Does this increase Redis traffic?

Does this increase allocations?

Simplicity

Could this be implemented as a Rule instead?

Stability

Does this preserve public contracts?

14. Stable APIs

The following interfaces should be treated as stable contracts.

CandidateSearcher
RingExpansionPolicy
ExpansionPolicy
DriverDensityProvider
CellDriverLookup
Selector
Pipeline
SearchProfileRule

Changes to these interfaces require architectural review.

15. Future Extension Guidelines

The Discovery Engine is expected to evolve.

Preferred extension points:

New Search Behavior

→ New Rule

New Candidate Retrieval

→ New Strategy

New Infrastructure

→ New Provider

New Expansion Logic

→ New ExpansionPolicy

New Search Policy

→ Extend SearchProfile

New Runtime Context

→ SearchState

New Budget Type

→ SearchBudget

Future Features

The following features should integrate without modifying existing architecture:

Airport Dispatch

Retry Profiles

Premium Dispatch

VIP Drivers

Scheduled Rides

Regional Dispatch

Marketplace Balancing

Demand Prediction

Machine Learning Ranking

Dynamic Ring Expansion

Weather-Aware Search

Traffic-Aware Search

Driver Preference Matching

Each should be implemented through existing extension points rather than by introducing conditional logic into core execution paths.

AI Agent Final Instructions

Before generating code, ask these questions:

Which layer owns this responsibility?
Can this be implemented as a new Rule, Strategy, Policy, or Provider instead of modifying existing code?
Does this preserve the downward dependency flow?
Will this component remain independently testable?
Am I exposing behavior instead of implementation details?
Am I introducing coupling that future features will have to undo?

If the answer to any of these questions is no, stop and redesign the implementation before generating code.

Architecture Contract

Every AI-generated contribution to the Discovery Engine must preserve these principles:

Single Responsibility: Every component owns one clear responsibility.
Composition over Conditionals: New behavior is composed through Rules, Policies, and Strategies—not nested if statements.
Dependency Inversion: High-level modules depend on abstractions, never concrete infrastructure.
Immutable Policies: SearchProfile and policy objects are immutable after construction.
Behavior over Configuration: Consumers invoke behavior through interfaces instead of inspecting configuration data.
Constructor Injection: All object composition happens in bootstrap/container.go.
Open for Extension: Future features should be added through extension points, not by modifying stable execution paths.

These rules are the long-term contract for the Discovery Engine. Any future implementation—whether written by a human or an AI agent—should maintain them to preserve the architecture's consistency, scalability, and maintainability.
