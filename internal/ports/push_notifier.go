package ports

import (
	"context"

	"github.com/google/uuid"
)

type PushNotifier interface {
	SendRideOffer(ctx context.Context, rideID uuid.UUID, driverID uuid.UUID) error
}
