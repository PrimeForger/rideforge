package application

import "github.com/google/uuid"

type CreateRideRequest struct {
	RiderID uuid.UUID
	From    string
	To      string
}

type AssignDriverRequest struct {
	RideID   uuid.UUID
	DriverID uuid.UUID
}
