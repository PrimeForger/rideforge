package application

import (
	"context"
	"errors"

	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type MatchingService struct {
	driverRepo ports.DriverRepository
	rideSvc    *RideService
}

func NewMatchingService(
	driverRepo ports.DriverRepository,
	rideSvc *RideService,
) *MatchingService {
	return &MatchingService{
		driverRepo: driverRepo,
		rideSvc:    rideSvc,
	}
}

func (m *MatchingService) MatchRide(
	ctx context.Context,
	rideID string,
) error {

	drivers, err := m.driverRepo.GetAvailableDrivers(ctx)

	if err != nil {
		return err
	}

	if len(drivers) == 0 {
		return errors.New("no drivers available")
	}

	driver := drivers[0]

	rideUUID, _ := uuid.Parse(rideID)

	req := AssignDriverRequest{
		RideID:   rideUUID,
		DriverID: driver.ID,
	}

	return m.rideSvc.AssignDriver(ctx, req)
}
