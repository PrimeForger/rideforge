package events

import "github.com/google/uuid"

type Event interface {
	Name() string
}

type RideRequestedEvent struct {
	RideID uuid.UUID
}

func (e RideRequestedEvent) Name() string {
	return "ride.requested"
}

type RideAcceptedEvent struct {
	RideID   uuid.UUID
	DriverID uuid.UUID
}

func (e RideAcceptedEvent) Name() string {
	return "ride.accepted"
}
