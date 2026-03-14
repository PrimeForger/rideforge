package application

import (
	"context"
	"errors"

	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/region"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type RideService struct {
	rideRepo ports.RideRepository
	eventBus ports.EventBus
}

func NewRideService(
	rideRepo ports.RideRepository,
	eventBus ports.EventBus,
) *RideService {
	return &RideService{
		rideRepo: rideRepo,
		eventBus: eventBus,
	}
}

func (s *RideService) CreateRide(
	ctx context.Context,
	req CreateRideRequest,
	fromRegion region.Region,
	toRegion region.Region,
) (uuid.UUID, error) {

	// 1. Legal region validation
	if !region.IsRideAllowed(fromRegion, toRegion) {
		return uuid.Nil, errors.New("cross-region rides not allowed")
	}

	// 2. Create domain aggregate
	r := ride.NewRide(req.RiderID)

	// 3. Save to repository
	if err := s.rideRepo.Save(ctx, r); err != nil {
		return uuid.Nil, err
	}

	// 4. Emit event
	event := events.RideRequested{RideID: r.ID}
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return uuid.Nil, err
	}

	return r.ID, nil
}

func (s *RideService) AssignDriver(
	ctx context.Context,
	req AssignDriverRequest,
) error {

	r, err := s.rideRepo.GetByID(ctx, req.RideID)
	if err != nil {
		return err
	}

	if err := r.AssignDriver(req.DriverID); err != nil {
		return err
	}

	if err := s.rideRepo.Save(ctx, r); err != nil {
		return err
	}

	event := events.RideAccepted{
		RideID:   r.ID,
		DriverID: req.DriverID,
	}

	return s.eventBus.Publish(ctx, event)
}
