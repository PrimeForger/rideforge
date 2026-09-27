# Candidate Ranking & Scoring (`internal/application/dispatch/candidate/ranking`)

## Purpose
Computes multi-factor scores for candidate drivers combining travel time (ETA), driver rating, acceptance history, and fairness.

## Public Interface
`Ranker` interface (`Score(ctx, ride, candidate) (float64, error)`), `WeightedRanker`, `ScoreWeights`, `FairnessFeatures`.

## Dependencies
`internal/application/dispatch/candidate`, `internal/config`.

## Invariants & Gotchas
Scores are strictly normalized between 0.0 and 1.0. Drivers exceeding maximum ETA receive heavy penalty scores. Weights must sum to 1.0.

## Human-Approval Required
Yes. Ranking algorithms directly influence driver income distribution, rider wait times, and system fairness.

## Related ADRs
ADR-0015 (Smart Dispatch), ADR-0016 (AI-Assisted Dispatch Strategy).

## Known Issues / Tech Debt
Static linear scoring weights will be augmented with machine-learning ranking models in Phase 3.
