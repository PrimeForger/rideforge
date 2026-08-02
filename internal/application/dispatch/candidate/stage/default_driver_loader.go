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
	driverCache DriverCache
}

func NewDefaultDriverLoader(
	driverCache DriverCache,
) *DefaultDriverLoader {

	return &DefaultDriverLoader{
		driverCache: driverCache,
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

	drivers, err := l.driverCache.GetDrivers(
		ctx,
		ids,
	)
	if err != nil {
		return err
	}

	for _, d := range drivers {

		candidate := candidates.FindByID(
			d.ID,
		)

		if candidate == nil {
			continue
		}

		candidate.Driver = d
	}

	pipelineCtx.Result.LoadedCandidates = len(drivers)

	return nil
}

var _ pipeline.Stage = (*DefaultDriverLoader)(nil)
