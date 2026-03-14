package matching

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
	"github.com/google/uuid"
)

type MatchingService struct {
	producer *kafka.Producer
}

func NewMatchingService(producer *kafka.Producer) *MatchingService {
	return &MatchingService{producer: producer}
}

func (m *MatchingService) HandleRideRequested(ctx context.Context, message []byte) error {
	var event events.Event
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	if event.Type != ride.EventRideRequested {
		return nil
	}

	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	var payload ride.RideRequested
	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	log.Println("Matching ride:", payload.RideID)

	// Simulate driver selection
	driverID := "driver-123"

	matchEvent := events.Event{
		ID:        uuid.NewString(),
		Type:      ride.EventDriverMatched,
		Aggregate: payload.RideID,
		Data: ride.DriverMatched{
			RideID:    payload.RideID,
			DriverID:  driverID,
			MatchedAt: time.Now(),
		},
		Occurred: time.Now(),
	}

	return m.producer.Publish(ctx, payload.RideID, matchEvent)
}
