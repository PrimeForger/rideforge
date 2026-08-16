package refinement

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/candidate"
	"github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/search"
	"github.com/google/uuid"
)

type RedisGeoRefiner struct {
	source GeoRefinementSource
}

func NewRedisGeoRefiner(
	source GeoRefinementSource,
) *RedisGeoRefiner {

	return &RedisGeoRefiner{
		source: source,
	}
}

func (r *RedisGeoRefiner) Refine(
	ctx context.Context,
	result search.Result,
	req search.Request,
) (search.Result, error) {

	driverIDs := candidateIDs(result.Candidates)

	refined, err := r.source.NearestDrivers(
		ctx,
		req.PickupLat,
		req.PickupLng,
		req.RadiusKm,
		driverIDs,
	)
	if err != nil {
		return result, err
	}

	result.Candidates = candidate.NewCollectionFromIDs(refined)

	return result, nil
}

func candidateIDs(
	collection *candidate.Collection,
) []uuid.UUID {

	ids := make([]uuid.UUID, 0, collection.Len())

	it := collection.Iterator()

	for {
		c, ok := it.Next()
		if !ok {
			break
		}

		ids = append(ids, c.ID)
	}

	return ids
}
