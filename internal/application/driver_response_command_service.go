package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type DriverResponseCommandService struct {
	txManager  *postgres.TxManager
	outboxRepo ports.OutboxRepository
}

func NewDriverResponseCommandService(
	txManager *postgres.TxManager,
	outboxRepo ports.OutboxRepository,
) *DriverResponseCommandService {
	return &DriverResponseCommandService{
		txManager:  txManager,
		outboxRepo: outboxRepo,
	}
}

func (s *DriverResponseCommandService) AcceptRide(
	ctx context.Context,
	rideID,
	driverID uuid.UUID,
) error {
	return s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {
		event := events.DriverAcceptedEvent{
			RideID:   rideID,
			DriverID: driverID,
		}

		return s.insertEvent(ctx, tx, rideID, event)
	})
}

func (s *DriverResponseCommandService) RejectRide(
	ctx context.Context,
	rideID,
	driverID uuid.UUID,
) error {
	return s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {
		event := events.DriverRejectedEvent{
			RideID:   rideID,
			DriverID: driverID,
		}

		return s.insertEvent(ctx, tx, rideID, event)
	})
}

func (s *DriverResponseCommandService) insertEvent(
	ctx context.Context,
	tx *sql.Tx,
	aggregateID uuid.UUID,
	event events.Event,
) error {
	envelope := appevents.Envelope{
		ID:        uuid.NewString(),
		Type:      event.Name(),
		Aggregate: aggregateID.String(),
		Data:      event,
		Occurred:  time.Now(),
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	return s.outboxRepo.Insert(ctx, tx,
		outbox.NewEvent(aggregateID, envelope.Type, payload),
	)
}
