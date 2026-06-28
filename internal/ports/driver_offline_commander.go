package ports

import (
	"context"

	"github.com/google/uuid"
)

type DriverOfflineCommander interface {
	GoOffline(ctx context.Context, driverID uuid.UUID, reason string) error
}
