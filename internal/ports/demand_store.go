package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CellDemandStore interface {
	AddActiveDemand(ctx context.Context, rideID uuid.UUID, cellID string, createdAt time.Time) error
	RemoveActiveDemand(ctx context.Context, rideID uuid.UUID, cellID string) error
	GetCellDemand(ctx context.Context, cellID string) (int, error)
	GetMultiCellDemand(ctx context.Context, cellIDs []string) (map[string]int, error)
	ReconcileCellDemand(ctx context.Context, cellID string, activeRideIDs map[uuid.UUID]time.Time) error
}
