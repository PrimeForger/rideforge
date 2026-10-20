package application

import (
	"errors"
	"math"

	"github.com/google/uuid"
)

type CreateRideRequest struct {
	RiderID         uuid.UUID
	PickupLatitude  *float64
	PickupLongitude *float64
	From            string
	To              string
}

func (r CreateRideRequest) ValidatePickup() error {
	if r.PickupLatitude == nil || r.PickupLongitude == nil {
		return errors.New("pickup latitude and longitude are required")
	}
	lat, lng := *r.PickupLatitude, *r.PickupLongitude
	if math.IsNaN(lat) || math.IsInf(lat, 0) || lat < -90 || lat > 90 {
		return errors.New("pickup latitude must be between -90 and 90")
	}
	if math.IsNaN(lng) || math.IsInf(lng, 0) || lng < -180 || lng > 180 {
		return errors.New("pickup longitude must be between -180 and 180")
	}
	return nil
}

type AssignDriverRequest struct {
	RideID   uuid.UUID
	DriverID uuid.UUID
}
