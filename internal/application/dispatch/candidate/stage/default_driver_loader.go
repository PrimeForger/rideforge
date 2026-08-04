package stage

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate/pipeline"
	"github.com/google/uuid"
)

// Load driver information for every discovered candidate in a single batch.
// Subsequent stages operate only on enriched candidates.

type DefaultDriverLoader struct {
	source BatchDataSource
}

func NewDefaultDriverLoader(
	source BatchDataSource,
) *DefaultDriverLoader {

	return &DefaultDriverLoader{
		source: source,
	}
}

func (l *DefaultDriverLoader) Execute(
	ctx context.Context,
	pipelineCtx *pipeline.Context,
	candidates *candidate.Collection,
) error {

	ids := make([]uuid.UUID, 0, candidates.Len())

	it := candidates.Iterator()

	for {
		candidate, ok := it.Next()
		if !ok {
			break
		}

		ids = append(ids, candidate.ID)
	}

	drivers, err := l.source.LoadDrivers(
		ctx,
		ids,
	)
	if err != nil {
		return err
	}

	candidateIndex := make(map[uuid.UUID]*candidate.Candidate, candidates.Len())

	candidates.ForEach(func(c *candidate.Candidate) {
		candidateIndex[c.ID] = c
	})

	for _, d := range drivers {

		candidate := candidateIndex[d.ID]

		if candidate == nil {
			continue
		}

		candidate.Driver = d
	}

	pipelineCtx.Result.LoadedCandidates += len(drivers)

	return nil
}

var _ pipeline.Stage = (*DefaultDriverLoader)(nil)
