package builders

import "github.com/ashadashraf/ride-hail-app/internal/application/dispatch/discovery/refinement"

func BuildGeoRefiner(source refinement.GeoRefinementSource) refinement.Refiner {
	return refinement.NewRedisGeoRefiner(source)
}
