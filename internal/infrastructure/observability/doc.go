// Package observability provides the foundational metrics and tracing telemetry infrastructure for RideForge.
//
// Stage 10 / Phase B — Dispatch Latency Observability Model:
//
// 1. Candidate Lookup (dispatch.candidate_lookup / dispatch_candidate_lookup_duration_seconds):
//   - Start: Beginning of candidate search/discovery (MatchingEngine.HandleMatchingStarted -> CandidateSearcher.FindCandidates).
//   - End: Candidate search completion before candidate pipeline processing.
//   - Note: Represents outer candidate discovery latency.
//
// 2. Candidate Ranking (dispatch.ranking / dispatch_ranking_duration_seconds):
//   - Start: Beginning of candidate pipeline execution (MatchingEngine.HandleMatchingStarted -> CandidatePipeline.Execute).
//   - End: Candidate pipeline completion (filtering, scoring, ordering).
//
// 3. Driver Reservation (dispatch.reservation / dispatch_reservation_duration_seconds):
//   - Start: Driver locking attempt (MatchingEngine.offerCandidate -> DriverLocker.Reserve).
//   - End: Lock result returned (labels: result = success | skipped | error).
//
// 4. Driver Offer Delivery (dispatch.offer / dispatch_offer_duration_seconds):
//   - Start: Offer transmission attempt (DriverOfferGateway.SendOffer).
//   - End: WebSocket or Push delivery completed (labels: status = delivered_ws | ws_failed | delivered_push | push_failed).
//
// 5. Driver Acceptance Handling (dispatch.acceptance / dispatch_acceptance_duration_seconds):
//   - Start: Driver response processing (DriverResponseService.HandleDriverAccepted / HandleDriverRejected / HandleDriverTimeout).
//   - End: Response handling, outbox event generation, and state updates complete (labels: result = accepted | rejected | timeout | already_accepted | error).
//
// 6. End-to-End Matching Attempt (MatchingEngine.HandleMatchingStarted / matching_duration_seconds):
//   - Preserved from initial design to measure complete matching attempt duration.
//
// Stage 10 / Phase C — H3 & Candidate Retrieval Observability Model:
//
// 1. Low-Level H3 Cell Lookup Duration (h3_lookup_duration_seconds):
//   - Start: Beginning of Redis/H3 cell driver lookup (H3DriverIndex.GetDriversInCells).
//   - End: Completion of SMembers retrieval for queried H3 cells.
//   - Boundary Distinction: Measures lower-level H3/Redis retrieval specifically, unlike dispatch_candidate_lookup_duration_seconds which measures outer candidate discovery.
//
// 2. H3 Search Expansion & Volume:
//   - h3_search_rings_visited: Histogram tracking count of concentric H3 rings searched.
//   - h3_lookup_cell_count: Histogram tracking total number of H3 cells searched.
//   - h3_lookup_result_count: Histogram tracking count of raw driver IDs retrieved from H3 index.
//   - h3_cell_updates_total: Counter tracking driver H3 cell updates (labels: result = added | moved | unchanged | error).
//   - h3_driver_removals_total: Counter tracking driver H3 index removals (labels: result = success | not_found | error).
//
// 3. Candidate Pipeline Processing & Retrieval Outcome:
//   - candidate_pipeline_count: Histogram tracking candidate counts at pipeline stages (labels: stage = loaded | filtered | ranked).
//   - candidate_retrieval_outcome_total: Counter tracking candidate retrieval outcomes (labels: result = found | empty | error).
//
// Stage 10 / Phase D — Operational Spatial / Heatmap Observability Model:
//
// 1. Dispatch Density Classifications (dispatch_density_classifications_total):
//   - Counter tracking H3 ring driver density classifications evaluated during search expansion.
//   - Source of truth: DensityRule.Apply (labels: density = sparse | normal | dense).
//   - Semantics: Represents density evaluation events during dispatch search expansion, not a static count of dense cells.
//
// 2. Dispatch Outcomes & Failure Deduplication:
//   - Matching attempts & failure outcomes: Tracked by matching_attempts_total (labels: result = success | max_attempts_reached | no_drivers_reserved | candidate_search_error | candidate_pipeline_error).
//   - Offer timeouts: Tracked by driver_timeouts_total (labels: reason).
//   - Candidate retrieval outcomes: Tracked by candidate_retrieval_outcome_total (labels: result = found | empty | error).
//   - Deduplication Policy: Existing metrics are reused as single sources of truth without creating redundant failure counter wrappers.
//
// 3. Spatial Aggregation Architectural Distinction:
//   - Prometheus metrics store bounded operational telemetry only (density classifications, matching outcomes, indexed driver totals).
//   - H3 cell IDs, driver IDs, ride IDs, and coordinates are STRICTLY FORBIDDEN as Prometheus metric labels.
//   - Detailed per-cell heatmap visualization relies on spatial backend stores (e.g. Redis H3 sets `drivers:h3:<cell>`) and query endpoints.
//
// Telemetry Principles & Conventions:
//
// 1. Metric Naming:
//   - Preserves existing Prometheus metric naming conventions (e.g., matching_attempts_total, driver_offers_total, dispatch_*_duration_seconds, h3_*, dispatch_density_classifications_total).
//   - Uses Prometheus counter, histogram, and gauge primitives.
//
// 2. Span Naming:
//   - Follows hierarchical span naming conventions:
//     dispatch
//     ├── dispatch.candidate_lookup (attributes: search.backend, search.rings_visited, search.cells_visited, candidates.discovered, candidates.loaded, candidates.filtered, candidates.ranked, search.result)
//     ├── dispatch.ranking
//     ├── dispatch.reservation
//     ├── dispatch.offer
//     └── dispatch.acceptance
//
// 3. Cardinality Policy:
//   - High-cardinality values (ride_id, driver_id, user_id, request_id, cell ID, coordinates, raw error strings)
//     MUST NOT be used as metric labels. They belong exclusively in trace spans and log contexts.
//   - Metric labels MUST be strictly bounded low-cardinality dimensions (result, delivery_status, type, reason, status, operation, stage, density).
//
// 4. Testability:
//   - Supports registering metrics with custom Prometheus registries via RegisterWithRegistry for isolated testing.
//   - Supports graceful OpenTelemetry initialization and shutdown via InitTracing and NoopShutdown.
package observability
