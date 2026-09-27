# Candidate Processing Pipeline (`internal/application/dispatch/candidate/pipeline`)

## Purpose
Orchestrates the multi-stage candidate processing pipeline (batch loading, filtering, scoring, and heap ordering).

## Public Interface
`Pipeline` interface (`Execute(ctx, request) (*Result, error)`), `DefaultPipeline`, `New(stages ...Stage) Pipeline`.

## Dependencies
`context`, `internal/application/dispatch/candidate`, `internal/application/dispatch/candidate/stage`.

## Invariants & Gotchas
Stages execute in strict sequence. If any stage filters out all candidates, the pipeline short-circuits and returns an empty result to avoid unnecessary downstream computations.

## Human-Approval Required
Yes. Controls the qualification pipeline that determines which drivers are eligible for ride offers.

## Related ADRs
ADR-0015 (Smart Dispatch), `docs/02-architecture/4.CANDIDATE_PIPELINE.md`.

## Known Issues / Tech Debt
Pipeline stages are statically configured; dynamic stage injection based on ride priority is planned.
