package ports

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/google/uuid"
)

type DriverRepository interface {
	GetAvailableDrivers(ctx context.Context) ([]*driver.Driver, error)
	GetByID(ctx context.Context, id uuid.UUID) (*driver.Driver, error)
	Save(ctx context.Context, d *driver.Driver) error
}
