package realtime

import (
	"context"
	"log"

	"github.com/google/uuid"
)

type DriverOfferGateway struct{}

func NewDriverOfferGateway() *DriverOfferGateway {
	return &DriverOfferGateway{}
}

func (g *DriverOfferGateway) SendOffer(
	ctx context.Context,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {
	log.Printf("sending ride offer to driver: ride_id=%s driver_id=%s", rideID, driverID)
	return nil
}
