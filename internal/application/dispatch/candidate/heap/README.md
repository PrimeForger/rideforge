# Bounded Candidate Heap (`internal/application/dispatch/candidate/heap`)

## Purpose
Provides a memory-bounded min/max heap for ordering candidates by score and extracting top-K candidates efficiently.

## Public Interface
`CandidateHeap`, `NewCandidateHeap(maxSize int)`, `Push(c *Candidate)`, `Pop() *Candidate`, `Len() int`, `ToSlice() []*Candidate`.

## Dependencies
`container/heap`, `internal/application/dispatch/candidate`.

## Invariants & Gotchas
Heap size is strictly capped at `maxSize`. When full, incoming candidates with lower scores than the root are dropped immediately with O(log K) complexity.

## Human-Approval Required
No.

## Related ADRs
`docs/02-architecture/4.CANDIDATE_PIPELINE.md`.

## Known Issues / Tech Debt
None.
