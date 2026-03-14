package rideapp

import (
	"context"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
	"github.com/google/uuid"
)

type RideService struct {
	producer *kafka.Producer
}

func NewRideService(producer *kafka.Producer) *RideService {
	return &RideService{producer: producer}
}

func (s *RideService) RequestRide(ctx context.Context, passenger, pickup, dropoff string) (string, error) {
	rideID := uuid.NewString()

	payload := ride.RideRequested{
		RideID:      rideID,
		Passenger:   passenger,
		Pickup:      pickup,
		Dropoff:     dropoff,
		RequestedAt: time.Now(),
	}

	event := events.Event{
		ID:        uuid.NewString(),
		Type:      ride.EventRideRequested,
		Aggregate: rideID,
		Data:      payload,
		Occurred:  time.Now(),
	}

	err := s.producer.Publish(ctx, rideID, event)
	if err != nil {
		return "", err
	}

	return rideID, nil
}
