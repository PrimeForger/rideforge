package stage

import (
	"context"

	"github.com/google/uuid"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
)

type BatchDataSource interface {
	LoadDrivers(
		ctx context.Context,
		driverIDs []uuid.UUID,
	) ([]*driver.Driver, error)
}
