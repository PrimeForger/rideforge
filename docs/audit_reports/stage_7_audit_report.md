Stage 7 — Demand & Surge / Hot Region Scaling

Current-State Audit Report

Project: RideForge (Go Ride-Hailing Backend)
Audit Scope: Stage 7 — Demand & Surge / Hot Region Scaling Readiness & Implementation Boundary
Auditor: Senior Go Backend Architect
Date: August 23, 2026

1. Executive Summary

This codebase audit evaluates the readiness of the RideForge repository for Stage 7 — Demand & Surge / Hot Region Scaling. RideForge has successfully completed Stage 5 (Dispatch Intelligence), Stage 6 (Geo Engine), Stage 9 (Reliability & H3 Recovery), and Stage 10 (Observability).

Key Findings:

Supply Tracking is Already Implemented (Stage 6 / Stage 9):
Driver supply per H3 cell is actively maintained in Redis using atomic Lua scripts (UpdateDriverCellScript in internal/infrastructure/redis/h3_driver_index.go) and Redis SETs (drivers:h3:<cell_id>). Real-time driver supply per cell is readable via H3DensityProvider.DriverCountInRing (SCard).

Existing "Density System" is a Search-Expansion Dispatch Policy, NOT Demand/Surge:
The current DensityClassifier and DensityRule evaluate only supply count during candidate discovery ring expansion. It is not demand-aware, does not compute imbalance ratios, does not classify hot cells, and does not compute surge pricing.

Demand Tracking is Completely Missing:
While ride requests flow through RideService and emit ride.requested domain events via PostgreSQL Outbox, there are no Redis counters or aggregation mechanisms tracking active demand per H3 cell.

Hot Cell Detection and Surge Computation are Completely Missing:
A search across all .go files confirms zero matches for surge or hot-cell imbalance classification.

Stage 9 Reliability & Stage 10 Observability are Ready for Reuse:
The existing H3RecoveryService, HeartbeatRecoveryWorker, and Prometheus metrics suite in internal/infrastructure/observability/metrics.go provide a robust foundation for cell state recovery and metric collection.

Verdict:
Stage 7 is NOT implemented. However, because Stage 6 and Stage 9 built full H3 cell driver supply indexing, Phase 1 of Stage 7 (Cell Supply) is ~70% complete. Stage 7 implementation can proceed immediately by building on top of the existing H3 index and EventRouter without duplicating spatial or reliability components.

2. Stage 7 Original Requirements Mapping

The table below maps the original Stage 7 requirements against the current codebase implementation status:

Requirement

Status

Existing Component

Evidence

Gap

23. Cell occupancy cache (Supply)

COMPLETE

H3DriverIndex, H3DensityProvider

h3_driver_index.go:L57-L60, h3_density_provider.go:L26-L68

None. Tracks active drivers in Redis SET drivers:h3:<cell_id>.

23. Cell occupancy cache (Demand)

NOT IMPLEMENTED

None

No Redis key or repository exists for cell demand counts.

Missing demand:h3:<cell_id> tracking active unmatched requests per cell.

23. Cell occupancy cache (Requests/Rides)

PARTIAL

PostgreSQL rides table

ride_service.go:L42-L90

Rides are stored transactionally in PostgreSQL, but not aggregated per H3 cell in real-time.

24. Hot cell detection (Drivers < Requests)

NOT IMPLEMENTED

None

Zero occurrences of hot cell logic or imbalance thresholds in codebase.

Missing HotCellDetector and cell status classification (NORMAL, WARM, HOT).

25. Surge computation (Surge per H3 cell)

NOT IMPLEMENTED

None

Zero occurrences of surge logic in .go code files.

Missing surge calculation engine, pricing multipliers, and surge configuration.

3. Existing Architecture Relevant to Stage 7

RideForge's current architecture provides strong abstractions that Stage 7 must integrate into:

[ Rider Request ] ──> RideService (PostgreSQL + Outbox)
│
▼
Domain Event: "ride.requested"
│
▼
Kafka Event Bus ──> EventRouter (Side-Effect Handlers)
│
├──> GeoService (Redis GEO)
├──> H3DriverIndex (Redis SET drivers:h3:<cell_id>)
└──> [ Stage 7 Target: Demand & Surge Service ]

Driver Lifecycle & Location:
Real-time driver location updates enter via DriverLocationService.ProcessRealtimeLocation. It validates coordinates, checks sequence numbers (AcceptLocationSeq), updates Redis GEO, updates the H3 cell index (H3DriverIndex.UpdateDriverCell), updates DriverCache details, and refreshes the driver's liveness heartbeat.

Ride Lifecycle & Event Pipeline:
Rides are created in RideService.CreateRide inside a PostgreSQL transaction (TxManager). An outbox record is created atomically. Kafka consumer receives ride.requested and routes it through EventRouter.Handle.

Geospatial & H3 Engine:
geo.H3Service handles Uber H3 resolution (default Resolution 8, ~0.73 km² cell area).

Reliability Infrastructure (Stage 9):
H3RecoveryService reconciles derived Redis indexes against PostgreSQL source of truth (GetAvailableDrivers). HeartbeatRecoveryWorker purges offline/stale drivers when heartbeats expire (default 30s TTL).

4. Existing Density System Audit

A critical question for Stage 7 is whether the existing density system satisfies Stage 7 or should be reused.

Detailed Code Audit of Density Components

DensityClass: Enum (DensitySparse, DensityNormal, DensityDense).

DensityClassifier: Classifies a single integer (driverCount) against sparseUpperBound and denseUpperBound.

DriverDensityProvider: Interface requiring DriverCountInRing(ctx, centerCell, ring).

H3DensityProvider: Implementation that resolves H3 ring cells and pipelines SCard calls across Redis drivers:h3:<cell_id> keys.

DensityRule: Discovery policy executed during candidate search. It sets SearchState.LastDensity and picks an expansion profile (DefaultExpansionProfileProvider).

Density System Audit Findings

Inputs: driverCount (count of online drivers in surrounding H3 rings).

Outputs: Candidate search expansion profile (e.g., how aggressively to expand k-ring search radii).

Persistence: Transient. Evaluated in-memory per candidate search request. Not stored in Redis or database.

Demand/Surge Awareness: None. Does not count ride requests, active rides, or demand/supply ratios.

Can the existing density system be reused for Stage 7?

No, but it can consume Stage 7.
The existing DensityRule is merely a candidate search expansion policy inside the dispatch discovery engine. It must remain intact to prevent breaking candidate search. However, Stage 7's cell occupancy tracking will expose a richer CellOccupancyProvider that DensityRule can query in the future if search expansion needs to become demand-aware.

5. Demand State Audit

Current State of Demand

Origin: Rides are created via HTTP POST /rides handled by RideHandler.CreateRide. RideService.CreateRide validates regional legal permissions, saves the Ride record in PostgreSQL, and inserts a ride.requested outbox event.

Available Events:

ride.requested (Demand created)

matching.started (Matching initiated)

driver.accepted / ride.driver_matched (Demand satisfied / resolved)

ride.completed / ride.cancelled / driver.timeout (Demand closed / retried)

Counters / Aggregates: None. No Redis key or memory structure tracks active demand per cell.

Missing Pieces:

H3 cell resolution for ride pickup coordinates (lat, lng -> H3 cell ID).

Redis atomic counter (INCRBY / DECRBY on demand:h3:<cell_id> or sliding-window ZSET).

Demand event handlers wired into EventRouter.

6. Supply State Audit

Current State of Supply

Authoritative Source of Truth: PostgreSQL drivers table (status = 'ONLINE').

Real-Time Index: Redis SET drivers:h3:<cell_id> managed by H3DriverIndex.

Liveness & Heartbeat: DriverCache maintains driver:heartbeat:<driver_id> (30s TTL). HeartbeatRecoveryWorker automatically marks drivers offline when heartbeats expire, purging them from drivers:h3:<cell_id>.

Conclusion for Stage 7 Supply:
Stage 7 does NOT need to create a new supply counter or supply cache. Real-time supply per H3 cell is already available via SCard drivers:h3:<cell_id>.

7. Cell State Audit

Existing Redis keys and data structures in the codebase:

Purpose

Redis Key Pattern

Redis Data Type

TTL

Defined In

Driver H3 Cell Pointer

driver:h3_cell:<driver_id>

String

90s (configurable)

keys.go:L4

Cell Drivers Index

drivers:h3:<cell_id>

Set

90s (configurable)

keys.go:L5

Driver Details

driver:<driver_id>

Hash

None

driver_cache.go:L90

Online Drivers Set

drivers:online

Set

None

driver_cache.go:L307

Driver Heartbeat

driver:heartbeat:<driver_id>

String

30s

driver_cache.go:L336

Driver Lock

driver:lock:<driver_id>

String

20s

driver_locker.go

Gap for Stage 7:
Missing Redis keys:

demand:h3:<cell_id> (String/Hash tracking active unmatched request count per cell).

cell:state:<cell_id> (Hash tracking computed imbalance ratio, status HOT/NORMAL, surge multiplier, updated timestamp).

8. Hot Cell Detection Audit

Current Implementation: None.

Codebase Search: No matches for hot cell thresholding, Demand > Supply ratio evaluation, or cell status tagging.

Gap: Need a HotCellDetector domain service that evaluates:
$$\text{Imbalance Ratio} = \frac{\text{Active Demand}}{\max(\text{Active Supply}, 1)}$$
If ratio $> \text{HotThreshold}$ (e.g. $1.5$) and Demand $\ge \text{MinDemandThreshold}$, mark cell HOT.

9. Surge Audit

Current Implementation: None.

Codebase Search: Zero occurrences of surge computation or pricing multiplier logic in any .go file.

Explicit Statement: Surge computation is not currently implemented.

Required Boundary: Stage 7 must introduce a decoupled SurgeCalculator that computes multiplier $M \in [1.0, \text{MaxSurge}]$ based on cell imbalance ratio and supply elasticity without embedding pricing logic directly inside the candidate search or driver locker.

10. Event Flow Analysis

The diagram below shows how existing domain events flow through RideForge and where Stage 7 state hooks should naturally attach:

Ride Lifecycle Events Stage 7 State Impact

---

[ Ride Created ] ──> ride.requested ─────────> INCR demand:h3:<pickup_cell>
│
[ Matching Fails ] ──> matching.retry ─────────> (Maintain / Refresh TTL)
│
[ Driver Accepts ] ──> driver.accepted ────────> DECR demand:h3:<pickup_cell>
│
[ Ride Cancelled ] ──> ride.cancelled ─────────> DECR demand:h3:<pickup_cell>

Driver Lifecycle Events Stage 7 State Impact

---

[ Driver Location ] ──> driver.location ───────> Atomic update drivers:h3:<cell> (Stage 6)
│
[ Driver Online ] ──> driver.online ───────────> SADD drivers:h3:<cell> (Stage 6)
│
[ Driver Offline ] ──> driver.offline ──────────> SREM drivers:h3:<cell> (Stage 6)
│
▼
[ Recalculate Hot Cell & Surge ]

By subscribing to ride.requested, driver.accepted, and ride.cancelled inside EventRouter, Stage 7 can maintain real-time demand state without changing any core ride transaction logic.

11. Reliability Analysis (Stage 9 Integration)

Impact of Failure Scenarios on Stage 7 State

Redis Restart / Crash:

Supply: Fully recovered by H3RecoveryService.RebuildIndex which queries PostgreSQL available drivers and rebuilds drivers:h3:<cell_id>.

Demand: Demands in Redis will be wiped out. Stage 7 can implement a DemandRecoveryService that queries active rides in state REQUESTED/MATCHING from PostgreSQL and repopulates demand:h3:<cell_id>.

Stale Driver / Heartbeat Expiry:

Handled seamlessly by HeartbeatRecoveryWorker. When a heartbeat expires, the worker triggers GoOffline, emitting driver.offline which removes the driver from drivers:h3:<cell_id>.

Duplicate Events / Replays:

Handled seamlessly by ProcessedEventRepository.InsertIfNotExistsTx. Duplicate event IDs are skipped before side-effects execute.

Out-of-Order Events:

Location sequences are enforced by AcceptLocationSeq. For demand, setting TTLs (e.g. 5 minutes) on demand keys prevents leaked counters if a cancellation event is dropped.

12. Observability Analysis (Stage 10 Integration)

Stage 10 already includes comprehensive Prometheus metrics in metrics.go:

Existing Stage 10 Metrics Reusable by Stage 7

h3_indexed_drivers (Gauge): Live count of supply drivers indexed in H3.

h3_cell_updates_total (CounterVec: added, moved, unchanged, error): Track driver movement across cells.

h3_lookup_result_count (Histogram): Track candidate driver retrieval density.

dispatch_density_classifications_total (CounterVec: sparse, normal, dense): Track search expansion classifications.

Necessary New Stage 7 Metrics

To operate and debug Stage 7, only 4 new metrics are required:

cell_demand_active (GaugeVec: cell_id): Current active ride requests per H3 cell.

cell_imbalance_ratio (HistogramVec: cell_id): Demand/Supply ratio distribution.

hot_cells_active_total (GaugeVec: status = hot, warm, normal): Total active hot cells.

surge_multiplier_applied (Histogram): Applied surge multipliers.

13. Data Ownership / Source of Truth Matrix

Data Element

Source of Truth

Derived Real-Time State

Storage Location

Update Mechanism

Driver Availability

PostgreSQL drivers.status

Redis driver:<id> Hash

PostgreSQL / Redis

Transactional API / Outbox

Driver Location

Redis / WebSocket feed

Redis GEO & drivers:h3:<cell>

Redis

DriverLocationService

H3 Cell Supply

Derived from Driver Location

Redis SET drivers:h3:<cell_id>

Redis

H3DriverIndex.UpdateDriverCell (Lua)

Active Ride Requests

PostgreSQL rides.status

Redis String demand:h3:<cell_id>

PostgreSQL / Redis

EventRouter on ride.requested / driver.accepted

Hot-Cell Status

Computed on-the-fly

Redis Hash cell:state:<cell_id>

Redis

HotCellService evaluation worker / trigger

Surge Multiplier

Computed on-the-fly

Redis Hash cell:state:<cell_id>

Redis

SurgeCalculator

Historical Occupancy/Surge

PostgreSQL (Not yet created)

N/A

PostgreSQL cell_occupancy_history

Background periodic flush worker

14. Architectural Risks

Risk of Duplicating Density Systems:
Creating a separate DensityService for Stage 7 would duplicate DensityClassifier.
Mitigation: Keep DensityRule strictly for candidate search expansion, and create CellOccupancyService strictly for demand/supply imbalance and surge pricing.

Risk of Leaked Demand Counters:
If a ride is abandoned without emitting a completion/cancellation event, Redis demand:h3:<cell_id> could stay artificially high.
Mitigation: Use sliding-window ZSETs or short TTLs (e.g. 3-5 minutes) with periodic reconciliation against PostgreSQL rides WHERE status IN ('REQUESTED', 'MATCHING').

Risk of High-Cardinality Prometheus Metrics:
Labeling Prometheus metrics with every raw H3 Cell ID (millions globally) will crash Prometheus.
Mitigation: Aggregate Prometheus metrics by Region / District, while keeping granular H3 cell counts in Redis.

Risk of Coupling Stage 7 to Dispatch Transaction Path:
Calculating surge pricing synchronously inside the matching loop will increase dispatch latency.
Mitigation: Surge calculation must be asynchronous or cached in Redis cell:state:<cell_id> so reads are $O(1)$.

15. What MUST NOT Be Built

Stage 7 implementation MUST REUSE the following existing codebase components and MUST NOT DUPLICATE them:

Do NOT build a new H3 cell calculator:
Reuse geo.H3Service for location-to-cell mapping (CellForLocation).

Do NOT build a new driver cell index or Redis key schema for drivers:
Reuse H3DriverIndex and Redis SET drivers:h3:<cell_id>.

Do NOT build a new driver liveness or heartbeat mechanism:
Reuse Stage 9 HeartbeatRecoveryWorker.

Do NOT build a new event bus or outbox consumer:
Reuse Kafka consumer infrastructure and EventRouter.

Do NOT build a new driver recovery mechanism:
Reuse Stage 9 H3RecoveryService.

16. Stage 7 Gap Matrix

Capability

Existing

Partial

Missing

Recommended Action

Cell Driver Supply Occupancy

YES

—

—

Reuse H3DriverIndex and h3CellDriversKey(cell).

Cell Request Demand Occupancy

—

—

YES

Create demand:h3:<cell_id> Redis counter and event handlers.

Demand Tracking Engine

—

—

YES

Build DemandTracker listening to ride.requested, driver.accepted, ride.cancelled.

Supply Tracking Engine

YES

—

—

Reuse H3DensityProvider & H3DriverIndex.

Hot Cell Detection

—

—

YES

Build HotCellDetector service calculating $\frac{\text{Demand}}{\text{Supply}}$ imbalance.

Surge Computation

—

—

YES

Build SurgeCalculator with configurable multiplier curves.

Cell State Recovery

—

YES

—

Extend H3RecoveryService with active ride demand reconciliation.

Surge Configuration

—

—

YES

Add SurgeConfig to internal/config/config.go.

Occupancy & Surge APIs

—

—

YES

Add HTTP/admin query endpoints /occupancy/cell/:id and /surge/cell/:id.

Stage 7 Observability

—

YES

—

Add 4 target metrics to observability/metrics.go.

17. Recommended Stage 7 Boundaries

Based strictly on this audit, Stage 7 boundaries are defined as follows:

What Stage 7 Should Own:

Real-time cell demand aggregation (demand:h3:<cell_id>).

Demand/Supply imbalance ratio calculation per H3 cell.

Hot cell state classification (NORMAL, WARM, HOT).

Cell surge pricing multiplier computation ($1.0 \times - 3.0 \times$).

Storage of cell state snapshots in Redis cell:state:<cell_id>.

Exposing CellOccupancyService and SurgeService query ports for ride pricing and dispatch.

What Stage 7 Should Reuse:

H3DriverIndex and drivers:h3:<cell_id> for supply counts.

EventRouter and Outbox pattern for consuming lifecycle events.

H3RecoveryService for reliability.

metrics.go for Prometheus integration.

What Stage 7 Should NOT Touch:

Core driver candidate search pipeline filters (e.g. rating, experience).

Driver lock reservation (RedisDriverLocker).

Stage 8 multi-node clustering or distributed partition consensus.

Stage 11 machine learning model serving / prediction pipelines.

18. Recommended Implementation Phases

To deliver Stage 7 cleanly without architectural risk, the following 4-phase sequence is recommended:

Phase A: Cell Demand Tracking Foundation
│
▼
Phase B: Hot Cell Detection & Imbalance Engine
│
▼
Phase C: Surge Pricing Calculation Engine
│
▼
Phase D: Cell Recovery, APIs & Observability

Phase A — Cell Demand Tracking Foundation

Purpose: Track real-time active ride requests per H3 cell in Redis.

Dependencies: geo.H3Service, EventRouter, Outbox event pipeline.

Affected Components: internal/infrastructure/redis/demand_store.go, internal/application/demand_service.go, EventRouter side-effect handlers.

Expected Outcome: demand:h3:<cell_id> counters increment on ride.requested and decrement on driver.accepted/ride.cancelled.

Phase B — Hot Cell Detection & Imbalance Engine

Purpose: Compute real-time demand-to-supply ratios per H3 cell and classify hot cells.

Dependencies: Phase A demand counters + existing Stage 6 H3DensityProvider supply counts.

Affected Components: internal/domain/occupancy/hot_cell.go, internal/application/hot_cell_service.go.

Expected Outcome: Atomic calculation of imbalance ratio and tagging cells as HOT, WARM, or NORMAL.

Phase C — Surge Pricing Calculation Engine

Purpose: Compute surge multipliers based on hot cell status and imbalance ratios.

Dependencies: Phase B hot cell classifications + SurgeConfig configuration.

Affected Components: internal/domain/surge/calculator.go, internal/config/config.go, internal/application/surge_service.go.

Expected Outcome: Computed surge multipliers cached in Redis cell:state:<cell_id> available for fare quotes.

Phase D — Cell Recovery, APIs & Observability

Purpose: Ensure cell state recovers cleanly after Redis restart, expose inspection APIs, and add Prometheus metrics.

Dependencies: Phases A–C.

Affected Components: internal/application/h3_recovery_service.go, internal/server/, internal/infrastructure/observability/metrics.go.

Expected Outcome: 100% production-ready Stage 7 capability with metrics, recovery, and HTTP monitoring endpoints.

19. Verification Strategy

Future Stage 7 implementation must be verified using the following automated test suite:

Unit Tests:

hot_cell_test.go: Test imbalance ratio calculation edge cases (0 supply, 0 demand, negative values, high demand).

surge_calculator_test.go: Test multiplier curves, minimum/maximum clamps, and configuration thresholds.

Integration Tests (Redis):

demand_store_test.go: Verify atomic INCR/DECR and TTL expiration on Redis keys using real/miniredis.

cell_occupancy_test.go: Verify simultaneous reading of supply from drivers:h3:<cell> and demand from demand:h3:<cell>.

Event-Flow Tests:

demand_event_handler_test.go: Simulate ride.requested -> driver.accepted event flow through EventRouter and verify demand counter accuracy.

Recovery Tests:

demand_recovery_test.go: Flush Redis, simulate active rides in PostgreSQL rides table, trigger recovery worker, and verify demand state repopulation.

Concurrency & Idempotency Tests:

Run 100 concurrent ride.requested events with duplicate event IDs to verify ProcessedEventRepository prevents double counting.

Full Repository Verification:

Execute go test -v ./... across the entire workspace to ensure zero regression in dispatch, geo, matching, or reliability.

20. Final Recommendation

Decision

Stage 7 partially exists (Supply side) and should be completed.

Concise Answers to Architectural Questions:

What should we implement first?
Implement Phase A (Cell Demand Tracking Foundation) by adding demand:h3:<cell_id> Redis storage and wiring event handlers into EventRouter.

What existing system should be reused?
Reuse existing H3Service for cell resolution, H3DriverIndex (drivers:h3:<cell_id>) for supply counts, EventRouter for event consumption, and H3RecoveryService for reliability.

What should we avoid building?
Avoid building a duplicate driver location index, a duplicate density classifier, or a separate event consumer group.

What is the minimum production-grade Stage 7 scope?
Real-time cell demand counters in Redis, hot-cell imbalance classification, surge multiplier calculation, demand recovery integration, and 4 Prometheus metrics.

What should be deferred to Stage 8?
Multi-node Redis cluster partitioning, cross-node gossip protocols, and distributed lock sharding.

What should be deferred to Stage 11?
Machine-learning-based predictive demand forecasting, spatial neural networks, and historical feature store offline ingestion pipelines.

End of Audit Report.
