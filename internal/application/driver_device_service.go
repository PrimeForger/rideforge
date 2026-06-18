package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
	"github.com/google/uuid"
)

type DriverDeviceService struct {
	txManager     *postgres.TxManager
	pushTokenRepo ports.DriverPushTokenRepository
	outboxRepo    ports.OutboxRepository
}

func NewDriverDeviceService(
	txManager *postgres.TxManager,
	pushTokenRepo ports.DriverPushTokenRepository,
	outboxRepo ports.OutboxRepository,
) *DriverDeviceService {
	return &DriverDeviceService{
		txManager:     txManager,
		pushTokenRepo: pushTokenRepo,
		outboxRepo:    outboxRepo,
	}
}

func (s *DriverDeviceService) RegisterPushToken(
	ctx context.Context,
	driverID uuid.UUID,
	deviceID string,
	platform string,
	token string,
) error {
	deviceID = strings.TrimSpace(deviceID)
	platform = strings.ToLower(strings.TrimSpace(platform))
	token = strings.TrimSpace(token)

	if deviceID == "" {
		return errors.New("device_id is required")
	}

	if token == "" {
		return errors.New("token is required")
	}

	if platform != "android" && platform != "ios" {
		return errors.New("invalid platform")
	}

	return s.txManager.WithinTx(ctx, func(tx *sql.Tx) error {
		if err := s.pushTokenRepo.UpsertTx(
			ctx,
			tx,
			driverID,
			deviceID,
			platform,
			token,
		); err != nil {
			return err
		}

		event := events.DriverPushTokenUpdatedEvent{
			DriverID: driverID,
			DeviceID: deviceID,
			Platform: platform,
			Token:    token,
		}

		envelope := appevents.Envelope{
			ID:        uuid.NewString(),
			Type:      event.Name(),
			Aggregate: driverID.String(),
			Data:      event,
			Occurred:  time.Now(),
		}

		payload, err := json.Marshal(envelope)
		if err != nil {
			return err
		}

		return s.outboxRepo.Insert(ctx, tx,
			outbox.NewEvent(driverID, envelope.Type, payload),
		)
	})
}
