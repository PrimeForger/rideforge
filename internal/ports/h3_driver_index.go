package ports

import (
	"context"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/google/uuid"
)

type DriverCellUpdateStatus int

const (
	DriverCellUnchanged DriverCellUpdateStatus = iota
	DriverCellMoved
	DriverCellAdded
)

type CellUpdateResult struct {
	Status  DriverCellUpdateStatus
	OldCell string
	NewCell string
}

type DriverRemoveResult struct {
	Removed bool
	OldCell string
}

type H3CellCalculator interface {
	CellForLocation(lat, lng float64) (string, error)
}

type H3DriverIndexer interface {
	UpdateDriverCell(ctx context.Context, driverID uuid.UUID, newCell string) (CellUpdateResult, error)
	RemoveDriver(ctx context.Context, driverID uuid.UUID) (DriverRemoveResult, error)
}

type GeoIndexer interface {
	UpdateDriverLocation(ctx context.Context, driverID uuid.UUID, lat, lng float64) error
	RemoveDriver(ctx context.Context, driverID uuid.UUID) error
}

type RealtimeDriverCache interface {
	GetOnlineDriverIDs(ctx context.Context) ([]uuid.UUID, error)
	LoadDrivers(ctx context.Context, driverIDs []uuid.UUID) ([]*driver.Driver, error)
}
