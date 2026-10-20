package application

import (
	"context"
	"errors"
	"testing"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/google/uuid"
)

type projectionDemandServiceFake struct {
	requested int
	accepted  int
	err       error
	lat       float64
	lng       float64
}

func (f *projectionDemandServiceFake) HandleRideRequested(_ context.Context, _ uuid.UUID, lat, lng float64) error {
	f.requested++
	f.lat, f.lng = lat, lng
	return f.err
}

func (f *projectionDemandServiceFake) HandleDriverAccepted(_ context.Context, _ uuid.UUID) error {
	f.accepted++
	return f.err
}

type projectionProcessedEventsFake struct {
	processed map[string]bool
	err       error
}

func (f *projectionProcessedEventsFake) IsProcessed(_ context.Context, eventID uuid.UUID, consumer string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.processed[eventID.String()+":"+consumer], nil
}

func (f *projectionProcessedEventsFake) InsertIfNotExists(_ context.Context, eventID uuid.UUID, consumer string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	key := eventID.String() + ":" + consumer
	if f.processed[key] {
		return false, nil
	}
	f.processed[key] = true
	return true, nil
}

func TestProcessDemandProjectionRecordsOnlySuccessfulProjection(t *testing.T) {
	eventID := uuid.New()
	rideID := uuid.New()
	marker := eventID.String() + ":" + demandProjectionConsumer
	failure := errors.New("redis unavailable")

	t.Run("failed projection stays unprocessed", func(t *testing.T) {
		demand := &projectionDemandServiceFake{err: failure}
		processed := &projectionProcessedEventsFake{processed: make(map[string]bool)}
		envelope := appevents.Envelope{
			ID:       eventID.String(),
			Type:     "ride.requested",
			Data:     events.RideRequestedEvent{RideID: rideID, PickupLatitude: 0, PickupLongitude: 0},
			Occurred: time.Now(),
		}
		if err := processDemandProjection(context.Background(), envelope, demand, processed); !errors.Is(err, failure) {
			t.Fatalf("processDemandProjection() error = %v, want %v", err, failure)
		}
		if processed.processed[marker] {
			t.Fatal("failed demand projection was marked processed")
		}
	})

	t.Run("successful duplicate is idempotently skipped", func(t *testing.T) {
		demand := &projectionDemandServiceFake{}
		processed := &projectionProcessedEventsFake{processed: make(map[string]bool)}
		envelope := appevents.Envelope{
			ID:       eventID.String(),
			Type:     "ride.requested",
			Data:     events.RideRequestedEvent{RideID: rideID, PickupLatitude: 0, PickupLongitude: 0},
			Occurred: time.Now(),
		}
		for range 2 {
			if err := processDemandProjection(context.Background(), envelope, demand, processed); err != nil {
				t.Fatalf("processDemandProjection() error = %v", err)
			}
		}
		if demand.requested != 1 {
			t.Fatalf("requested calls = %d, want 1", demand.requested)
		}
		if demand.lat != 0 || demand.lng != 0 {
			t.Fatalf("coordinates = (%v, %v), want (0, 0)", demand.lat, demand.lng)
		}
	})

	t.Run("accepted event removes demand", func(t *testing.T) {
		demand := &projectionDemandServiceFake{}
		processed := &projectionProcessedEventsFake{processed: make(map[string]bool)}
		envelope := appevents.Envelope{
			ID:   uuid.NewString(),
			Type: "ride.accepted",
			Data: events.RideAcceptedEvent{RideID: rideID, DriverID: uuid.New()},
		}
		if err := processDemandProjection(context.Background(), envelope, demand, processed); err != nil {
			t.Fatalf("processDemandProjection() error = %v", err)
		}
		if demand.accepted != 1 {
			t.Fatalf("accepted calls = %d, want 1", demand.accepted)
		}
	})

	t.Run("requested event derives the demand cell", func(t *testing.T) {
		store := newMockDemandStore()
		h3 := geo.NewH3Service(8, 1)
		demand := NewDemandService(store, h3, nil)
		processed := &projectionProcessedEventsFake{processed: make(map[string]bool)}
		lat, lng := 9.9312, 76.2673
		envelope := appevents.Envelope{
			ID:       uuid.NewString(),
			Type:     "ride.requested",
			Data:     events.RideRequestedEvent{RideID: rideID, PickupLatitude: lat, PickupLongitude: lng},
			Occurred: time.Now(),
		}
		if err := processDemandProjection(context.Background(), envelope, demand, processed); err != nil {
			t.Fatalf("processDemandProjection() error = %v", err)
		}
		cellID, err := h3.CellForLocation(lat, lng)
		if err != nil {
			t.Fatal(err)
		}
		if got := len(store.added[cellID]); got != 1 {
			t.Fatalf("demand in derived cell %s = %d, want 1", cellID, got)
		}
	})
}
