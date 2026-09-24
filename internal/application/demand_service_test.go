package application

import (
	"context"
	"testing"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/google/uuid"
)

type mockDemandStore struct {
	added     map[string]map[uuid.UUID]time.Time
	removed   map[uuid.UUID]string
	reconcile map[string]map[uuid.UUID]time.Time
}

func newMockDemandStore() *mockDemandStore {
	return &mockDemandStore{
		added:     make(map[string]map[uuid.UUID]time.Time),
		removed:   make(map[uuid.UUID]string),
		reconcile: make(map[string]map[uuid.UUID]time.Time),
	}
}

func (m *mockDemandStore) AddActiveDemand(ctx context.Context, rideID uuid.UUID, cellID string, createdAt time.Time) error {
	if m.added[cellID] == nil {
		m.added[cellID] = make(map[uuid.UUID]time.Time)
	}
	m.added[cellID][rideID] = createdAt
	return nil
}

func (m *mockDemandStore) RemoveActiveDemand(ctx context.Context, rideID uuid.UUID, cellID string) error {
	m.removed[rideID] = cellID
	for c, set := range m.added {
		delete(set, rideID)
		if len(set) == 0 {
			delete(m.added, c)
		}
	}
	return nil
}

func (m *mockDemandStore) GetCellDemand(ctx context.Context, cellID string) (int, error) {
	return len(m.added[cellID]), nil
}

func (m *mockDemandStore) GetMultiCellDemand(ctx context.Context, cellIDs []string) (map[string]int, error) {
	res := make(map[string]int)
	for _, c := range cellIDs {
		res[c] = len(m.added[c])
	}
	return res, nil
}

func (m *mockDemandStore) ReconcileCellDemand(ctx context.Context, cellID string, activeRideIDs map[uuid.UUID]time.Time) error {
	m.reconcile[cellID] = activeRideIDs
	return nil
}

func TestDemandService_HandleRideRequested(t *testing.T) {
	store := newMockDemandStore()
	h3 := geo.NewH3Service(8, 1)
	service := NewDemandService(store, h3, nil)

	ctx := context.Background()
	rideID := uuid.New()

	// Kochi coordinates
	lat, lng := 9.9312, 76.2673
	err := service.HandleRideRequested(ctx, rideID, lat, lng, time.Now())
	if err != nil {
		t.Fatalf("unexpected error adding ride requested demand: %v", err)
	}

	cellID, err := h3.CellForLocation(lat, lng)
	if err != nil {
		t.Fatalf("failed to calculate H3 cell: %v", err)
	}

	count, err := service.GetCellDemand(ctx, cellID)
	if err != nil {
		t.Fatalf("failed to get cell demand: %v", err)
	}

	if count != 1 {
		t.Errorf("expected demand count 1, got %d", count)
	}
}

func TestDemandService_HandleDriverAccepted(t *testing.T) {
	store := newMockDemandStore()
	h3 := geo.NewH3Service(8, 1)
	service := NewDemandService(store, h3, nil)

	ctx := context.Background()
	rideID := uuid.New()

	lat, lng := 9.9312, 76.2673
	_ = service.HandleRideRequested(ctx, rideID, lat, lng, time.Now())

	cellID, _ := h3.CellForLocation(lat, lng)

	err := service.HandleDriverAccepted(ctx, rideID)
	if err != nil {
		t.Fatalf("unexpected error removing active demand: %v", err)
	}

	count, _ := service.GetCellDemand(ctx, cellID)
	if count != 0 {
		t.Errorf("expected demand count 0 after acceptance, got %d", count)
	}
}

func TestDemandService_InvalidCoordsSkipped(t *testing.T) {
	store := newMockDemandStore()
	h3 := geo.NewH3Service(8, 1)
	service := NewDemandService(store, h3, nil)

	ctx := context.Background()
	rideID := uuid.New()

	// 0,0 invalid coords
	err := service.HandleRideRequested(ctx, rideID, 0, 0, time.Now())
	if err != nil {
		t.Fatalf("unexpected error for invalid coords: %v", err)
	}

	if len(store.added) != 0 {
		t.Errorf("expected no demand added for invalid coords, got %d cells", len(store.added))
	}
}
