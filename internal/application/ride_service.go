package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/domain/region"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type RideService struct {
	rideRepo ports.RideRepository
	// eventBus ports.EventBus
	txManager  *postgres.TxManager
	outboxRepo ports.OutboxRepository
}

func NewRideService(
	rideRepo ports.RideRepository,
	// eventBus ports.EventBus,
	txManager *postgres.TxManager,
	outboxRepo ports.OutboxRepository,
) *RideService {
	return &RideService{
		rideRepo: rideRepo,
		// eventBus: eventBus,
		txManager:  txManager,
		outboxRepo: outboxRepo,
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

	err := s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

		// 3. Save to repository
		if err := s.rideRepo.SaveTx(ctx, tx, r); err != nil {
			return err
		}

		// 4. Emit event
		domainEvent := events.RideRequestedEvent{RideID: r.ID}

		envelope := appevents.Envelope{
			ID:        uuid.NewString(),
			Type:      domainEvent.Name(),
			Aggregate: r.ID.String(),
			Data:      domainEvent,
			Occurred:  time.Now(),
		}

		payload, err := json.Marshal(envelope)
		if err != nil {
			return err
		}

		outboxEvent := outbox.NewEvent(r.ID, envelope.Type, payload)

		return s.outboxRepo.Insert(ctx, tx, outboxEvent)
	})

	if err != nil {
		return uuid.Nil, err
	}

	return r.ID, nil
}

func (s *RideService) AssignDriver(
	ctx context.Context,
	req AssignDriverRequest,
) error {

	var ErrOptimisticLockConflict = errors.New("optimistic lock conflict")
	const maxRetries = 3

	for i := 0; i < maxRetries; i++ {

		err := s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

			// Fetch inside transaction
			r, err := s.rideRepo.GetByIDTx(ctx, tx, req.RideID)
			if err != nil {
				return err
			}

			if err := r.AssignDriver(req.DriverID); err != nil {
				return err
			}

			if err := s.rideRepo.SaveTx(ctx, tx, r); err != nil {
				return err
			}

			domainEvent := events.RideAcceptedEvent{
				RideID:   r.ID,
				DriverID: req.DriverID,
			}

			envelope := appevents.Envelope{
				ID:        uuid.NewString(),
				Type:      domainEvent.Name(),
				Aggregate: r.ID.String(),
				Data:      domainEvent,
				Occurred:  time.Now(),
			}

			payload, err := json.Marshal(envelope)
			if err != nil {
				return err
			}

			outboxEvent := outbox.NewEvent(r.ID, envelope.Type, payload)

			return s.outboxRepo.Insert(ctx, tx, outboxEvent)
		})

		if err == nil {
			return nil
		}

		if errors.Is(err, ErrOptimisticLockConflict) {
			log.Println("retrying due to optimistic lock conflict")
			continue
		}

		return err
	}

	return errors.New("failed to assign driver after retries")
}

func (s *RideService) StartMatching(
	ctx context.Context,
	rideID uuid.UUID,
) error {

	return s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

		r, err := s.rideRepo.GetByIDTx(ctx, tx, rideID)
		if err != nil {
			return err
		}

		if err := r.StartMatching(); err != nil {
			return err
		}

		return s.rideRepo.SaveTx(ctx, tx, r)
	})
}
