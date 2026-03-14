package driver

import (
	"errors"

	"github.com/google/uuid"
)

type Driver struct {
	ID     uuid.UUID
	Status Status
}

func NewDriver() *Driver {
	return &Driver{
		ID:     uuid.New(),
		Status: StatusOffline,
	}
}
func (d *Driver) GoOnline() {
	d.Status = StatusOnline
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
