package events

import "github.com/google/uuid"

type Event interface {
	Name() string
}

type RideRequestedEvent struct {
	RideID          uuid.UUID `json:"ride_id"`
	PickupLatitude  float64   `json:"pickup_latitude"`
	PickupLongitude float64   `json:"pickup_longitude"`
}

func (e RideRequestedEvent) Name() string {
	return "ride.requested"
}

type RideAcceptedEvent struct {
	RideID   uuid.UUID `json:"ride_id"`
	DriverID uuid.UUID `json:"driver_id"`
}

func (e RideAcceptedEvent) Name() string {
	return "ride.accepted"
}
