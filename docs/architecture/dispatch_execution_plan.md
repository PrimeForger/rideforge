# Dispatch Engine Execution Plan

This document defines the exact implementation order for the dispatch
engine.

The order is intentional.

Each milestone finishes in a deployable state.

Nothing should be partially implemented.

---

# Milestone 1

Production Grade H3 Index

Status

✅ Complete

Completed

- Atomic Lua updates
- Atomic removal
- Empty cell cleanup
- TTL refresh
- Pipeline lookup
- Metrics
- Tracing

Exit Criteria

✓ H3 index survives concurrent updates

✓ No stale memberships

✓ Lookup is pipeline based

✓ Metrics available

✓ Tracing available

---

# Milestone 2

Adaptive Candidate Search

Status

🟡 In Progress

Goal

Search only as much as necessary.

Tasks

□ Progressive ring expansion

□ Density-aware expansion

□ Search budget

Deliverables

SearchStrategy

ProgressiveRingSearcher

SearchBudget

SearchResult

Exit Criteria

✓ No fixed GridDisk search

✓ Search stops early

✓ Sparse and dense cities behave differently

✓ Candidate target respected

---

# Milestone 3

Candidate Retrieval Pipeline

Status

⬜ Pending

Goal

Decouple searching from processing.

Tasks

□ Candidate Provider

□ Candidate Iterator

□ Candidate Pipeline

□ Candidate Filters

□ Candidate Builder

Exit Criteria

✓ MatchingEngine receives candidates only

✓ DriverCache hidden

✓ Redis hidden

✓ Streaming supported

---

# Milestone 4

Production Ranking

Status

⬜ Pending

Goal

Ranking becomes a replaceable module.

Tasks

□ ETA feature

□ Distance feature

□ Idle time feature

□ Acceptance feature

□ Quality feature

□ Score normalization

Exit Criteria

✓ Ranking independent

✓ Feature weights configurable

✓ New features require no engine changes

---

# Milestone 5

Dispatch Optimizer

Status

⬜ Pending

Goal

Dispatch policy becomes adaptive.

Tasks

□ Adaptive batch size

□ Dynamic timeout

□ Retry optimizer

Exit Criteria

✓ Offer size adapts

✓ Timeout adapts

✓ Retry adapts

---

# Milestone 6

Geo Engine

Status

⬜ Pending

Goal

Separate spatial indexing from routing.

Tasks

□ H3 search

□ GEO refinement

□ Haversine fallback

□ ETA provider interface

Exit Criteria

✓ Ranking never performs geo lookup

✓ Search never computes ETA

---

# Milestone 7

Hot Region Scaling

Status

⬜ Pending

Goal

Support demand-aware dispatch.

Tasks

□ Cell occupancy cache

□ Demand cache

□ Supply cache

□ Hot cell detector

□ Surge calculator

Exit Criteria

✓ Cell state available

✓ Surge computed

---

# Milestone 8

Distributed Dispatch

Status

⬜ Pending

Goal

Scale dispatch across nodes.

Tasks

□ Regional partitioning

□ Cell ownership

□ Local dispatch

□ Ownership transfer

Exit Criteria

✓ Multiple dispatch nodes

✓ No duplicate dispatch

---

# Milestone 9

Reliability

Status

⬜ Pending

Goal

Automatic recovery.

Tasks

□ Redis rebuild

□ Background reconciliation

□ Stale cleanup

□ Consistency verification

Exit Criteria

✓ Self-healing index

✓ Self-healing cache

---

# Milestone 10

Observability

Status

⬜ Pending

Goal

Everything measurable.

Tasks

□ Dispatch latency

□ Ring metrics

□ Candidate metrics

□ Heatmaps

□ Reservation metrics

Exit Criteria

✓ Every dispatch phase measurable

---

# Milestone 11

Machine Learning Ready

Status

⬜ Pending

Goal

Replace heuristics without architecture changes.

Tasks

□ Acceptance prediction

□ ETA prediction

□ Policy learning

Exit Criteria

✓ Rule engine replaceable

✓ ML pluggable

---

# Refactoring Rules

Every refactoring must satisfy all of the following.

□ Public APIs remain stable whenever possible.

□ MatchingEngine becomes smaller after refactoring.

□ Infrastructure never owns business decisions.

□ Every new component has one responsibility.

□ Every dependency points downward.

□ No duplicate logic.

□ Every feature has metrics.

□ Every feature has tracing.

□ Every feature has integration tests.

---

# Definition of Done

A milestone is complete only if all of the following are true.

□ Code implemented.

□ Unit tests pass.

□ Integration tests pass.

□ Metrics exported.

□ Traces visible.

□ Structured logs added.

□ Documentation updated.

□ Existing functionality preserved.

□ Benchmarks executed (if performance-critical).

□ No TODOs remain.

---

# Final Target Architecture

                        Ride Request
                              │
                              ▼
                    Dispatch Orchestrator
                              │
               ┌──────────────┴──────────────┐
               ▼                             ▼
       Search Strategy                Retry Strategy
               │
               ▼
        Candidate Search
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
      Driver Response Engine
               │
               ▼
      Assignment / Retry
