package stage

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"

	"github.com/google/uuid"
)

type DriverCache interface {
	GetDrivers(
		ctx context.Context,
		driverIDs []uuid.UUID,
	) ([]*driver.Driver, error)
}
