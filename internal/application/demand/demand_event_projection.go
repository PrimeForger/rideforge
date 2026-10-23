package demand

import (
	"context"
	"fmt"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/google/uuid"
)

const demandProjectionConsumer = "demand-projection"

type demandProjectionService interface {
	HandleRideRequested(context.Context, uuid.UUID, float64, float64) error
	HandleDriverAccepted(context.Context, uuid.UUID) error
}

type demandProjectionProcessedEvents interface {
	IsProcessed(context.Context, uuid.UUID, string) (bool, error)
	InsertIfNotExists(context.Context, uuid.UUID, string) (bool, error)
}

func ProcessProjection(
	ctx context.Context,
	envelope appevents.Envelope,
	demandService demandProjectionService,
	processedEvents demandProjectionProcessedEvents,
) error {
	if demandService == nil {
		return fmt.Errorf("demand projection service is not configured")
	}
	if processedEvents == nil {
		return fmt.Errorf("demand projection processed-event repository is not configured")
	}
	eventID, err := uuid.Parse(envelope.ID)
	if err != nil {
		return fmt.Errorf("parse demand projection event ID: %w", err)
	}
	processed, err := processedEvents.IsProcessed(ctx, eventID, demandProjectionConsumer)
	if err != nil {
		return fmt.Errorf("check demand projection idempotency: %w", err)
	}
	if processed {
		return nil
	}

	switch envelope.Type {
	case "ride.requested":
		var data struct {
			RideID          string  `json:"ride_id"`
			PickupLatitude  float64 `json:"pickup_latitude"`
			PickupLongitude float64 `json:"pickup_longitude"`
		}
		if err := decodeEventData(envelope, &data); err != nil {
			return fmt.Errorf("decode ride.requested demand payload: %w", err)
		}
		rideID, err := parseUUID(data.RideID, "ride_id")
		if err != nil {
			return err
		}
		if err := demandService.HandleRideRequested(ctx, rideID, data.PickupLatitude, data.PickupLongitude); err != nil {
			return fmt.Errorf("project ride.requested demand: %w", err)
		}
	case "ride.accepted":
		var data struct {
			RideID string `json:"ride_id"`
		}
		if err := decodeEventData(envelope, &data); err != nil {
			return fmt.Errorf("decode ride.accepted demand payload: %w", err)
		}
		rideID, err := parseUUID(data.RideID, "ride_id")
		if err != nil {
			return err
		}
		if err := demandService.HandleDriverAccepted(ctx, rideID); err != nil {
			return fmt.Errorf("project ride.accepted demand: %w", err)
		}
	default:
		return fmt.Errorf("unsupported demand projection event %q", envelope.Type)
	}

	if _, err := processedEvents.InsertIfNotExists(ctx, eventID, demandProjectionConsumer); err != nil {
		return fmt.Errorf("record demand projection event: %w", err)
	}
	return nil
}

func processDemandProjection(
	ctx context.Context,
	envelope appevents.Envelope,
	demandService demandProjectionService,
	processedEvents demandProjectionProcessedEvents,
) error {
	return ProcessProjection(ctx, envelope, demandService, processedEvents)
}
