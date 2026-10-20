Stage 7 — Phase A

Cell Demand Tracking Foundation — Implementation Plan (Revised)

Project: RideForge (Go Ride-Hailing Backend)
Document Type: Architectural & Implementation Specification
Stage: Stage 7 — Demand & Surge / Hot Region Scaling
Phase: Phase A — Cell Demand Tracking Foundation
Status: REVISED SPECIFICATION (Ready for Approval)
Date: August 23, 2026

## Current Phase A Implementation Contract (2026-10-03)

The corrections implemented in code supersede older examples below where they conflict: ride creation requires pickup coordinates; `(0,0)` is valid; ride events use snake_case JSON; demand projection uses the separate `demand-projection` processed-event identity and records it only after Redis succeeds; and recovery reconciles the union of PostgreSQL-expected cells and the Redis cell registry using per-cell Lua scripts. Convergence after a concurrent snapshot race is eventual on a later reconciliation run.

## Phase A Implementation Verification (2026-10-03)

Implemented and verified: focused application, event, PostgreSQL repository, and live Redis integration tests; `go test -ldflags=-linkmode=internal ./...`; `go vet ./...`; and `go test -cover -ldflags=-linkmode=internal ./...` pass. The Redis integration tests were also run uncached against a local Redis instance. `go test -race ./...` was attempted but cannot link with the host Windows MinGW/sanitizer toolchain (missing Windows runtime symbols). The versioned pickup-coordinate SQL migration and repository binding/scanning were reviewed/tested, but not applied to PostgreSQL because no local PostgreSQL client/server is available.

Phase A is awaiting user verification. Do not start Phase B or Phase C until the user verifies this phase.

1. Objective

The objective of Stage 7 — Phase A is to establish a high-performance, fault-tolerant, and idempotent real-time demand tracking foundation at the Uber H3 spatial cell level (Resolution 8).

Phase A will answer the fundamental question:

"For any given H3 cell, how many active ride requests currently exist?"

Phase A focuses exclusively on the demand tracking foundation. It will not implement hot-cell classification, imbalance ratios, or surge pricing multipliers (which are reserved for Phase B and Phase C).

2. Current-State Dependencies & Audit Corrections

Following code-level verification, Phase A incorporates the following architectural corrections:

Geospatial H3 Resolution:
Reuses geo.H3Service for lat/lng to H3 cell resolution (CellForLocation at Resolution 8, ~0.73 km² per cell).

Driver Supply Infrastructure (Stage 6 / Stage 9):
Reuses H3DriverIndex and Redis SETs drivers:h3:<cell_id>. No new driver spatial index will be built.

Decoupled Recovery (Stage 9 vs Stage 7):
Stage 9 H3RecoveryService will NOT be modified to handle ride demand. Driver recovery (PostgreSQL Drivers $\rightarrow$ H3/Geo) and Ride Demand recovery (PostgreSQL Rides $\rightarrow$ H3 Demand) remain strictly separated into dedicated services (H3RecoveryService vs DemandRecoveryService).

Verified Domain Events:
Code audit confirms matching.failed and ride.cancelled do NOT currently exist in internal/domain/events/. Phase A will NOT invent fake event types. Demand increments on ride.requested and decrements on driver.accepted. Terminal transitions for cancelled/exhausted rides are handled via service methods and convergent reconciliation.

Non-Destructive Convergent Recovery:
Phase A eliminates ClearAllDemand() (which creates race conditions with live traffic). Instead, it implements convergent diff reconciliation (ZADD/ZREM diffs).

3. Active Demand Definition & Verified Lifecycle

Analysis of RideForge Ride Lifecycle

In RideForge, a ride aggregate transitions through the following statuses defined in internal/domain/ride/status.go:

               ┌─────────────────────────────────────────────────┐
               │                                                 │
               ▼                                                 │

[ Created ] ──────> REQUESTED ──> MATCHING ──> ACCEPTED ──> COMPLETED
(API) │ │ │
└──────────────┴─────────────┴──────> CANCELLED

REQUESTED: Ride aggregate created via RideService.CreateRide, stored in PostgreSQL, and ride.requested event emitted to Outbox.

MATCHING: RideService.StartMatchingTx transitions ride to matching mode; matching engine searches candidates and sends driver offers (driver.offered).

ACCEPTED: Driver accepts offer via DriverResponseService.HandleDriverAccepted; ride is assigned to driver (r.AssignDriver(driverID)).

CANCELLED: Ride is cancelled by passenger/system or exhausts matching attempts.

Canonical Definition of Active Demand

In RideForge, Active Demand is defined as:

The set of all ride requests currently seeking a driver in a given H3 cell (statuses REQUESTED and MATCHING).

Rationale:

ride.requested / matching.started = Demand Enters Active State: The passenger is waiting for a driver match in that cell.

driver.offered = Demand REMAINS Active: Offering a ride to a candidate driver does not satisfy demand. The driver may reject or time out. Demand remains active until accepted.

driver.accepted = Demand Leaves Active State (Satisfied): The driver has committed to the ride. Demand is fulfilled.

ride.cancelled / Cancel Service Call = Demand Leaves Active State (Terminated): The ride was cancelled or terminated. Demand is removed.

4. Ride Lifecycle Analysis

The table below traces how each lifecycle event impacts Active Demand:

Lifecycle Event

Event / Source

Ride Status

Active Demand Impact

Reason

Ride Created

ride.requested (Event)

REQUESTED

INCREMENT (+1)

New passenger request entering the pickup H3 cell.

Matching Started

matching.started (Event)

MATCHING

NO-OP (Keep)

Transition from REQUESTED to MATCHING; demand is already active.

Matching Retry

matching.retry (Event)

MATCHING

NO-OP (Keep)

Subsequent candidate search attempt; demand remains active.

Driver Offered

driver.offered (Event)

MATCHING

NO-OP (Keep)

Offer extended to driver, but not yet accepted.

Offer Acked

driver.offer.acked (Event)

MATCHING

NO-OP (Keep)

Driver device acknowledged offer delivery; offer still pending.

Driver Rejected

driver.rejected (Event)

MATCHING

NO-OP (Keep)

Driver declined offer; matching engine will retry next driver.

Driver Timeout

driver.timeout (Event)

MATCHING

NO-OP (Keep)

Offer expired without response; matching engine will retry next driver.

Driver Accepted

driver.accepted (Event)

ACCEPTED

DECREMENT (-1)

Driver accepted ride; demand satisfied.

Ride Cancelled

Service Call / Event

CANCELLED

DECREMENT (-1)

Passenger or system cancelled request; demand terminated.

Ride Completed

ride.completed (Event)

COMPLETED

NO-OP

Ride completed (demand was already decremented at acceptance).

5. Event Mapping & Codebase Verification

+-----------------------+---------------------+---------------------------------------------------------------+
| Codebase Event | Demand Action | Verification Status |
+-----------------------+---------------------+---------------------------------------------------------------+
| ride.requested | AddActiveDemand | CONFIRMED (Emitted in ride_service.go:L65) |
| driver.accepted | RemoveActiveDemand | CONFIRMED (Emitted in ride_service.go:L118) |
| ride.cancelled | RemoveActiveDemand | PROPOSED (To be added when CancelRide service method is wired)|
| matching.failed | N/A (Do not invent) | CONFIRMED MISSING (Handled via DemandReconciliationService) |
+-----------------------+---------------------+---------------------------------------------------------------+

6.  State Model & Source of Truth Architecture

             AUTHORITATIVE DATA (PostgreSQL)

    ┌──────────────────────────────────────────────┐
    │ rides (id, status, pickup_latitude, pickup_longitude) │
    └──────────────────────┬───────────────────────┘
    │ Asynchronous Domain Events (Outbox)
    ▼
    DERIVED REAL-TIME STATE (Redis)
    ┌──────────────────────────────────────────────┐
    │ demand:ride:<ride_id> ──> "cell_id" │ (String, configured TTL)
    │ demand:h3:<cell_id>:rides ──> ZSET │ (Member=ride_id, Score=timestamp)
    └──────────────────────┬───────────────────────┘
    │ Read / Prune Operations
    ▼

1.  ZREMRANGEBYSCORE demand:h3:<cell_id>:rides -inf (now - configured TTL) [Projection Freshness]
1.  ZCARD demand:h3:<cell_id>:rides ───────────────────────────> Real-Time Demand

Authoritative Source of Truth:
PostgreSQL rides table (status IN ('REQUESTED', 'MATCHING')).

Derived Real-Time State:
Redis ZSETs updated asynchronously by the demand projection after matching/event handling. Projection idempotency is recorded separately and only after a Redis operation succeeds; score pruning and PostgreSQL-backed reconciliation repair derived state.

State Lifespan:
Transient (lives while ride is active in matching). Derived state can be fully rebuilt or reconciled from PostgreSQL at any time.

7. Redis Storage Options Comparison

Four Redis state storage options were evaluated for cell demand tracking:

Criteria

Option A: Simple Integer (INCR/DECR)

Option B: Cell Ride Set (SADD/SREM)

Option C: Cell ZSET (ZADD/ZREM)

Option D (REVISED & SELECTED): Hybrid Pointer + Cell ZSET with Timestamp Scores

Redis Keys

demand:h3:<cell_id>

demand:h3:<cell_id>:rides

demand:h3:<cell_id>:zset

demand:ride:<ride_id> + demand:h3:<cell_id>:rides

Data Type

String (Integer)

Set of ride_id UUIDs

ZSet (ride_id, score=ts)

String (cell_id) + ZSET (member=ride_id, score=ts)

TTL Leak Risk

HIGH (Leaks over time)

HIGH (Set members never expire)

MEDIUM (Requires score pruning)

ZERO (ZSET score pruning + demand:ride:<id> configured TTL)

Idempotency

POOR (INCR twice = +2)

EXCELLENT (SADD twice = 1 item)

EXCELLENT (ZADD twice = 1 item)

EXCELLENT (ZADD / SET are natively idempotent)

Missing Coord Support

POOR (Needs coords on DECR)

POOR (Needs cell_id on SREM)

POOR (Needs cell_id on ZREM)

EXCELLENT (demand:ride:<id> provides cell on DECR)

Drift Safety

POOR (Can go negative)

GOOD (SCard $\ge 0$)

GOOD (ZCard $\ge 0$)

EXCELLENT (ZCard $\ge 0$, no negative counts)

Read Complexity

$O(1)$ (GET)

$O(1)$ (SCard)

$O(1)$ (ZCard)

$O(1)$ (ZREMRANGEBYSCORE + ZCard)

Write Complexity

$O(1)$

$O(1)$

$O(\log N)$

$O(\log N)$ (Lua script)

8. Selected Storage Model

Option D (Revised) — Hybrid Ride-Cell Pointer + Cell Ride ZSET with Timestamp Scores

Description:

Option D utilizes two complementary Redis structures:

Ride-to-Cell Pointer:
demand:ride:<ride_id> -> Value: <cell_id> (String with configured TTL; 30-minute default).

Cell Active Rides ZSET:
demand:h3:<cell_id>:rides -> ZSET where member = ride_id and score = unix_timestamp_seconds.

Solving the Cell Set Leak Issue:

In plain Sets (SADD), expiring demand:ride:<ride_id> leaves the member orphaned in demand:h3:<cell_id>:rides.
By using ZSET scores for the last projection refresh, entries that have not been refreshed within the configured demand TTL can be pruned natively in Redis. Reconciliation refreshes scores for rides still active in PostgreSQL.

Read Operation:

To get active demand count for cell <cell_id>:

ZREMRANGEBYSCORE demand:h3:<cell_id>:rides -inf (now - configured TTL)

ZCARD demand:h3:<cell_id>:rides $\rightarrow$ Active Demand Count.

Write Operations (Lua Script Encapsulation):

AddActiveDemand (on ride.requested):

SET demand:ride:<ride_id> <cell_id> EX <configured TTL>

ZADD demand:h3:<cell_id>:rides <now_timestamp> <ride_id>

EXPIRE demand:h3:<cell_id>:rides <configured TTL>

RemoveActiveDemand (on driver.accepted / ride.cancelled):

GET demand:ride:<ride_id> $\rightarrow$ get <cell_id> (fallback to event cell_id if present).

ZREM demand:h3:<cell_id>:rides <ride_id>

DEL demand:ride:<ride_id>

9. Idempotency vs Derived-State Correctness Strategy

A critical architectural distinction is enforced between Event Idempotency and Derived-State Correctness:

Event Idempotency (PostgreSQL processed_events):
Prevents duplicate execution of side-effects when Kafka redelivers the same event message.

Derived-State Correctness (PostgreSQL Authoritative Reconciler):
processed_events does NOT guarantee Redis derived-state correctness (e.g. if Redis fails after idempotency check, or event updates fail silently).
Therefore, Redis demand state is strictly a derived projection. It can be completely rebuilt or reconciled from PostgreSQL rides (status IN ('REQUESTED', 'MATCHING')) at any time, independently of event processing history.

10. Concurrency Strategy

No Distributed Locking Required:
Demand updates do not acquire driver reservation locks (RedisDriverLocker).

Atomic Lua Scripts:
Add/Remove operations on Redis use atomic Lua scripts to execute demand:ride:<id> and demand:h3:<cell>:rides updates in a single execution step.

Safe Concurrent Event Processing:
Multiple dispatch nodes can concurrently process ride.requested and driver.accepted without race conditions or mutex locking.

11. Failure & Convergent Recovery Strategy

Decoupled Demand Recovery Architecture

Rather than polluting Stage 9's H3RecoveryService, Stage 7 introduces a dedicated DemandRecoveryService:

Stage 9 (Driver Spatial Index Recovery)
┌─────────────────────────────────────────────────────────────┐
│ H3RecoveryService ──> PostgreSQL Drivers ──> Driver H3/Geo │
└─────────────────────────────────────────────────────────────┘

Stage 7 (Ride Demand Index Recovery)
┌─────────────────────────────────────────────────────────────┐
│ DemandRecoveryService ──> PostgreSQL Rides ──> Demand ZSETs │
└─────────────────────────────────────────────────────────────┘

Non-Destructive Convergent Diff Reconciliation

Phase A rejects ClearAllDemand() / DEL ALL because clearing Redis live state creates severe race conditions with incoming concurrent live traffic.

Instead, DemandRecoveryService executes Convergent Diff Reconciliation:

Fetch Authoritative Active Rides:
Query PostgreSQL rides table for active rides:

SELECT id, pickup_latitude, pickup_longitude, created_at
FROM rides
WHERE status IN ('REQUESTED', 'MATCHING');

Group Active Rides by Cell:
Resolve pickup_lat, pickup_lng $\rightarrow$ cell_id. Build expectedRides[cell_id] = map[ride_id]timestamp.

Execute Convergent Diff per Cell:
For each cell:

Read known Redis cells from `demand:h3:cells`; reconcile the union of Redis-known cells and PostgreSQL-expected cells.

For each cell, run one atomic Lua reconciliation. Remove stale ZSET members and delete a ride pointer only when it still points to the cell being reconciled.

Add Missing Rides (ZADD): Any active ride_id in PostgreSQL missing from Redis is added and its ride pointer refreshed to the authoritative cell.

Prune scores that have not been refreshed within the configured demand TTL. Reconciliation refreshes scores for active PostgreSQL rides.

Concurrent writes can temporarily race with the recovery snapshot. Subsequent reconciliation runs restore the projection from current PostgreSQL state; no global Redis clear is used.

12. H3 Cell Resolution

Resolution Source:
Pickup coordinates (lat, lng) from CreateRideRequest or RideRequestedEvent.

Resolution Mechanism:
Call geo.H3Service.CellForLocation(lat, lng) at Resolution 8.

Validation:
Coordinates are required at ride creation and validated for latitude/longitude bounds; `(0,0)` is valid. Missing coordinates are represented explicitly in the request contract, not inferred from a coordinate value.

13. Architecture & Port Design

RideForge's Clean Architecture layer separation will be preserved:

internal/
├── domain/
│ └── occupancy/
│ └── demand.go <-- Domain entity & active demand value objects
├── ports/
│ └── demand_store.go <-- Storage interface definition
├── application/
│ ├── demand_service.go <-- Application service handling events
│ ├── demand_recovery_service.go <-- Dedicated Stage 7 recovery service
│ └── demand_reconciliation_worker.go<-- Periodic background reconciliation worker
└── infrastructure/
└── redis/
├── demand_store.go <-- Redis implementation (Lua + ZSETs)
└── scripts/
└── demand_scripts.go <-- Atomic Lua scripts

Port Definition (internal/ports/demand_store.go)

package ports

import (
"context"
"time"
"github.com/google/uuid"
)

type CellDemandStore interface {
AddActiveDemand(ctx context.Context, rideID uuid.UUID, cellID string) error
RemoveActiveDemand(ctx context.Context, rideID uuid.UUID, cellID string) error
GetCellDemand(ctx context.Context, cellID string) (int, error)
GetMultiCellDemand(ctx context.Context, cellIDs []string) (map[string]int, error)
ListDemandCells(ctx context.Context) ([]string, error)
ReconcileCellDemand(ctx context.Context, cellID string, activeRideIDs map[uuid.UUID]time.Time) error
}

14. EventRouter Integration

Demand projection runs after the matching transaction or side-effect handler succeeds, under the separate `demand-projection` processed-event identity. `ride.requested` carries `ride_id`, `pickup_latitude`, and `pickup_longitude`; `ride.accepted` carries `ride_id` and `driver_id`. Projection errors are returned and the demand marker is written only after projection succeeds.

This keeps demand updates decoupled from synchronous ride creation and matching transactions while making projection errors visible to the consumer.

15. Stage 9 & Stage 7 Reconciliation Separation

Stage 9 H3ReconciliationWorker: Runs H3RecoveryService (rebuilding driver spatial index). Unchanged.

Stage 7 DemandReconciliationWorker: Runs DemandRecoveryService (rebuilding ride demand ZSETs). New worker running on configurable interval (default 300s).

16. Observability Integration

Per Stage 10 guidelines, NO high-cardinality Prometheus labels (such as H3 Cell ID, Ride ID, or coordinates) will be created.

New metrics added to internal/infrastructure/observability/metrics.go:

var (
CellDemandEventsTotal = prometheus.NewCounterVec(
prometheus.CounterOpts{
Name: "cell_demand_events_total",
Help: "Total cell demand events processed",
},
[]string{"event_type", "result"},
)

    CellActiveDemandTotal = prometheus.NewGauge(
    	prometheus.GaugeOpts{
    		Name: "cell_active_demand_total",
    		Help: "Current total active ride requests across all cells",
    	},
    )

    DemandReconciliationDurationSeconds = prometheus.NewHistogram(
    	prometheus.HistogramOpts{
    		Name:    "demand_reconciliation_duration_seconds",
    		Help:    "Demand reconciliation execution duration in seconds",
    		Buckets: prometheus.DefBuckets,
    	},
    )

)

17. Files Expected to Change / Be Created

CONFIRMED: No code will be written until this revised plan receives explicit approval.

New Files to Create:

internal/domain/occupancy/demand.go — Domain demand definitions.

internal/ports/demand_store.go — Port interface for demand storage.

internal/infrastructure/redis/demand_store.go — Redis ZSET implementation.

internal/infrastructure/redis/scripts/demand_scripts.go — Atomic Lua scripts for ZSET updates and score pruning.

internal/application/demand_service.go — Application service handling demand lifecycle.

internal/application/demand_recovery_service.go — Dedicated Stage 7 demand recovery service.

internal/application/demand_reconciliation_worker.go — Background reconciliation worker.

internal/application/demand_service_test.go — Unit tests.

internal/infrastructure/redis/demand_store_test.go — Redis ZSET integration tests.

Existing Files to Modify:

internal/ports/ride_repository.go — Add GetActiveRides(ctx context.Context) ([]\*ride.Ride, error).

internal/infrastructure/postgres/ride_repository.go — Implement GetActiveRides.

internal/application/event_router.go — Wire DemandService into ride.requested and driver.accepted handlers.

internal/infrastructure/observability/metrics.go — Register Phase A metrics.

internal/bootstrap/container.go — Wire DemandStore, DemandService, DemandRecoveryService, and DemandReconciliationWorker.

18. Tests Plan

The future Phase A implementation will be verified with 10 dedicated test suites:

TestDemandStore_AddActiveDemand: Verify ZADD and SET atomic key creation.

TestDemandStore_RemoveActiveDemand: Verify ZREM and DEL key removal.

TestDemandStore_ZSetAgePruning: Verify ZREMRANGEBYSCORE prunes entries not refreshed within the configured TTL while active ride reconciliation refreshes active members.

TestDemandStore_Idempotency: Execute AddActiveDemand 10x for same ride; verify ZCard == 1.

TestDemandService_HandleRideRequested: Verify coordinate-to-cell mapping and demand addition.

TestDemandService_HandleDriverAccepted: Verify demand removal when ride is matched.

TestEventRouter_DemandIntegration: End-to-end event router dispatch to demand service.

TestDemandRecovery_ConvergentReconciliation: Simulate stale rides in Redis vs PostgreSQL, run DemandRecoveryService, verify Redis ZSETs converge cleanly without wiping live concurrent demand.

TestDemandStore_Concurrency: 50 concurrent goroutines adding/removing demand in same cell without race conditions (go test -race).

TestObservability_DemandMetrics: Verify Prometheus metrics increment correctly on demand events.

19. Step-by-Step Implementation Sequence

Step 1: Define Ports & Domain Models
Create ports/demand_store.go and domain/occupancy/demand.go.
↓
Step 2: Implement Redis ZSET Storage & Lua Scripts
Create infrastructure/redis/scripts/demand_scripts.go and infrastructure/redis/demand_store.go.
↓
Step 3: Implement Application Demand Service
Create application/demand_service.go.
↓
Step 4: Update PostgreSQL Ride Repository
Add GetActiveRides to ports/ride_repository.go and postgres/ride_repository.go.
↓
Step 5: Implement Dedicated Demand Recovery Service & Worker
Create application/demand_recovery_service.go and application/demand_reconciliation_worker.go.
↓
Step 6: Wire EventRouter Side-Effect Handlers
Update application/event_router.go to invoke DemandService on ride events.
↓
Step 7: Add Prometheus Metrics
Register new metrics in infrastructure/observability/metrics.go.
↓
Step 8: Wire Bootstrap Container
Update bootstrap/container.go to instantiate and inject demand dependencies.
↓
Step 9: Run Automated Verification Suite
Execute unit tests, integration tests, race detector, and full repository tests.

20. Non-Goals (What Phase A Must NOT Implement)

NO Hot-Cell Detection: Do not write logic comparing Demand > Supply or tagging cells as HOT (reserved for Phase B).

NO Surge Computation: Do not implement pricing multipliers or fare adjustments (reserved for Phase C).

NO Changes to Candidate Search / DensityRule: Do not alter DensityRule or candidate ranking.

NO Invented Event Types: Do not invent matching.failed or fake event types not present in codebase.

NO Multi-Node Clustering: Do not introduce Stage 8 multi-node gossip or Redis Cluster sharding.

NO Machine Learning Prediction: Do not introduce Stage 11 ML prediction models.

21. Risks & Mitigations

Identified Risk

Impact

Mitigation Strategy

Leaked Demand Keys (Missing cancellation event)

Overcounts demand in cell

Option D uses ZSET projection-refresh scores + configured-TTL pruning + DemandRecoveryService reconciliation.

Race Conditions During Recovery

Live demand wiped out

Eliminated ClearAllDemand(); replaced with non-destructive convergent diff reconciliation (ZADD/ZREM).

Coupling Driver & Demand Recovery

Fragile Stage 9 service

Separated into DemandRecoveryService; Stage 9 H3RecoveryService remains untouched.

Event Replay Duplicate Counting

Artificial demand spikes

Enforced by PostgreSQL processed_events table + Redis ZADD Set idempotency.

22. Verification Plan

Upon approval and implementation of Phase A, the following automated commands must pass cleanly:

# 1. Code formatting check

gofmt -s -w .

# 2. Go static analysis

go vet ./...

# 3. Focused Phase A unit & integration tests with race detection

go test -v -race ./internal/application/... ./internal/infrastructure/redis/...

# 4. Full workspace test suite

go test -v ./...

23. Final Approval Checklist

Corrected Redis TTL leak assumption by using ZSET timestamp scores (ZREMRANGEBYSCORE).

Explicitly separated Event Idempotency (processed_events) from Derived-State Correctness (PostgreSQL active rides reconciler).

Verified codebase event definitions (ride.requested, driver.accepted) and avoided un-emitted event types.

Created dedicated DemandRecoveryService instead of polluting Stage 9 H3RecoveryService.

Replaced destructive ClearAllDemand() with non-destructive convergent diff reconciliation.

Retained Active Demand definition (REQUESTED + MATCHING).

Retained H3 Resolution 8, Clean Architecture layers, and Stage 10 Observability bounds.

End of Revised Implementation Plan.
