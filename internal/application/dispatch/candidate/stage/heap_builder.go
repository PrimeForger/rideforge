package stage

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	candidateheap "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/heap"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
)

// Convert the ranked collection into an ordered candidate source.
// The MatchingEngine consumes candidates without knowing the ordering strategy.

type HeapBuilder struct{}

func NewHeapBuilder() *HeapBuilder {

	return &HeapBuilder{}
}

func (b *HeapBuilder) Execute(
	ctx context.Context,
	pipelineCtx *pipeline.Context,
	candidates *candidate.Collection,
) error {

	heap := candidateheap.NewHeap()

	it := candidates.Iterator()

	for {

		candidate, ok := it.Next()
		if !ok {
			break
		}

		if candidate.Metadata.Score == nil {
			continue
		}

		heap.PushCandidate(candidate)
	}

	pipelineCtx.Result.Candidates =
		pipeline.NewHeapCandidateSource(heap)

	pipelineCtx.Result.RankedCandidates =
		heap.Len()

	return nil
}

var _ pipeline.Stage = (*HeapBuilder)(nil)
