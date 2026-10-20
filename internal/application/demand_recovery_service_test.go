package application

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/google/uuid"
)

type demandRecoveryRideRepositoryFake struct {
	rides []*ride.Ride
}

func (f *demandRecoveryRideRepositoryFake) Save(context.Context, *ride.Ride) error { return nil }
func (f *demandRecoveryRideRepositoryFake) SaveTx(context.Context, *sql.Tx, *ride.Ride) error {
	return nil
}
func (f *demandRecoveryRideRepositoryFake) GetByID(context.Context, uuid.UUID) (*ride.Ride, error) {
	return nil, nil
}
func (f *demandRecoveryRideRepositoryFake) GetByIDTx(context.Context, *sql.Tx, uuid.UUID) (*ride.Ride, error) {
	return nil, nil
}
func (f *demandRecoveryRideRepositoryFake) GetActiveRides(context.Context) ([]*ride.Ride, error) {
	return f.rides, nil
}

func TestDemandRecoveryReconcilesExpectedAndRedisOnlyCells(t *testing.T) {
	h3 := geo.NewH3Service(8, 1)
	lat, lng := 9.9312, 76.2673
	activeCell, err := h3.CellForLocation(lat, lng)
	if err != nil {
		t.Fatal(err)
	}
	wrongCell, err := h3.CellForLocation(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	staleCell, err := h3.CellForLocation(40.7128, -74.0060)
	if err != nil {
		t.Fatal(err)
	}
	activeID := uuid.New()
	terminalID := uuid.New()
	activeRide := &ride.Ride{
		ID: activeID, Status: ride.StatusMatching,
		PickupLatitude: lat, PickupLongitude: lng, CreatedAt: time.Now(),
	}
	terminalRide := &ride.Ride{
		ID: terminalID, Status: ride.StatusCancelled,
		PickupLatitude: lat, PickupLongitude: lng, CreatedAt: time.Now(),
	}
	store := newMockDemandStore()
	store.added[wrongCell] = map[uuid.UUID]time.Time{activeID: time.Now()}
	store.added[staleCell] = map[uuid.UUID]time.Time{uuid.New(): time.Now()}
	service := NewDemandRecoveryService(
		&demandRecoveryRideRepositoryFake{rides: []*ride.Ride{activeRide, terminalRide}},
		h3,
		store,
		nil,
	)

	result, err := service.RebuildDemand(context.Background())
	if err != nil {
		t.Fatalf("RebuildDemand() error = %v", err)
	}
	if result.TotalActiveRides != 1 {
		t.Fatalf("TotalActiveRides = %d, want 1", result.TotalActiveRides)
	}
	if _, ok := store.reconcile[activeCell][activeID]; !ok {
		t.Fatalf("active ride was not reconciled into derived cell %s", activeCell)
	}
	if len(store.reconcile[wrongCell]) != 0 {
		t.Fatalf("wrong cell %s should be reconciled empty, got %d rides", wrongCell, len(store.reconcile[wrongCell]))
	}
	if len(store.reconcile[staleCell]) != 0 {
		t.Fatalf("stale-only cell %s should be reconciled empty, got %d rides", staleCell, len(store.reconcile[staleCell]))
	}
	if _, ok := store.reconcile[activeCell][terminalID]; ok {
		t.Fatal("terminal ride was included in demand reconciliation")
	}
}
