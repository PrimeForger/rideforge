package ride

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Ride struct {
	ID        uuid.UUID
	DriverID  uuid.UUID
	RiderID   uuid.UUID
	Status    Status
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewRide(riderID uuid.UUID) *Ride {
	now := time.Now()

	return &Ride{
		ID:        uuid.New(),
		RiderID:   riderID,
		Status:    StatusRequested,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (r *Ride) Transition(next Status) error {
	if !r.Status.CanTransitionTo(next) {
		return errors.New("invalid ride state transition")
	}

	r.Status = next
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Ride) AssignDriver(driverID uuid.UUID) error {
	if r.Status != StatusMatching {
		return errors.New("driver can only be assigned during matching")
	}

	r.DriverID = driverID
	return r.Transition(StatusAccepted)
}
