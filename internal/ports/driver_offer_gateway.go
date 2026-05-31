package ports

import (
	"context"

	"github.com/google/uuid"
)

type DriverOfferGateway interface {
	SendOffer(ctx context.Context, rideID uuid.UUID, driverID uuid.UUID) error
}
