# Dispatch Engine Factory Builders (`internal/bootstrap/builders`)

## Purpose
Provides factory builder functions that construct candidate searchers, candidate processing pipelines, multi-factor rankers, and routing providers for the matching engine.

## Public Interface
`BuildCandidatePipeline(...)`, `BuildCandidateSearcher(...)`, `BuildRanker(...)`, `BuildRefiners(...)`, `BuildRouting(...)`.

## Dependencies
Relies on `internal/application/dispatch/candidate/*`, `internal/application/dispatch/discovery/*`, `internal/application/dispatch/routing`, and `internal/config`.

## Invariants & Gotchas
Constructs pipelines with fixed stage ordering (Loader -> Validation -> Deduplication -> Availability -> Offered Filter -> Ranking -> Heap). Changing builder sequence directly alters candidate qualification order.

## Human-Approval Required
Yes. Modifying pipeline composition or ranker wiring changes driver selection and ranking fairness.

## Related ADRs
ADR-0015 (Smart Dispatch & Stand Dispatch), ADR-0016 (AI-Assisted Dispatch Strategy), `docs/02-architecture/4.CANDIDATE_PIPELINE.md`.

## Known Issues / Tech Debt
Pipeline construction is currently static; dynamic stage configuration based on ride category (e.g. ambulance vs premium) is planned.
