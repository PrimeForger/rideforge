package pipeline

import (
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	candidateheap "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/heap"
)

type HeapCandidateSource struct {
	heap *candidateheap.Heap
}

func NewHeapCandidateSource(
	heap *candidateheap.Heap,
) *HeapCandidateSource {

	return &HeapCandidateSource{
		heap: heap,
	}
}

func (s *HeapCandidateSource) Next() (
	*candidate.Candidate,
	bool,
) {
	return s.heap.PopCandidate()
}

var _ CandidateSource = (*HeapCandidateSource)(nil)
