package matching

import "github.com/ashadashraf/ride-hail-app/internal/domain/driver"

type Ranker interface {
	Score(d *driver.Driver, distanceKm float64) float64
}
