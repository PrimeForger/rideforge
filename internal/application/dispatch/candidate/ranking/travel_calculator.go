package ranking

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
)

type TravelCalculator interface {
	Calculate(
		ctx context.Context,
		pickupLat float64,
		pickupLng float64,
		driver *driver.Driver,
	) (TravelFeatures, error)
}
