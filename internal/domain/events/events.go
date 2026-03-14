package events

import "github.com/google/uuid"

type Event interface {
	Name() string
}

type RideRequested struct {
	RideID uuid.UUID
}

func (e RideRequested) Name() string {
	return "ride.requested"
}

type RideAccepted struct {
	RideID   uuid.UUID
	DriverID uuid.UUID
}

func (e RideAccepted) Name() string {
	return "ride.accepted"
}
