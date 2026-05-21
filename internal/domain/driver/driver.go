package driver

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Driver struct {
	ID               uuid.UUID
	Status           Status
	Rating           float64
	AcceptanceRate   float64
	CancellationRate float64
	TimeoutRate      float64
	CompletedRides   int
	LastAssignedAt   time.Time
	Lat              float64
	Lng              float64
}

func NewDriver() *Driver {
	return &Driver{
		ID:     uuid.New(),
		Status: StatusOffline,
	}
}

func (d *Driver) IsAvailable() bool {
	return d.Status == StatusOnline
}

func (d *Driver) GoOnline() {
	d.Status = StatusOnline
}

func (d *Driver) GoOffline() {
	d.Status = StatusOffline
}

func (d *Driver) AssignRide() error {
	if d.Status != StatusOnline {
		return errors.New("driver not available")
	}

	d.Status = StatusBusy
	return nil
}

func (d *Driver) CompleteRide() {
	d.Status = StatusOnline
}
