package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application"
	"github.com/ashadashraf/ride-hail-app/internal/domain/occupancy"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
)

type mockOccupancyStore struct {
	supplies map[string]int
	demands  map[string]int
}

func newMockOccupancyStore() *mockOccupancyStore {
	return &mockOccupancyStore{
		supplies: make(map[string]int),
		demands:  make(map[string]int),
	}
}

func (m *mockOccupancyStore) GetCellSupply(ctx context.Context, cellID string) (int, error) {
	return m.supplies[cellID], nil
}

func (m *mockOccupancyStore) GetMultiCellSupply(ctx context.Context, cellIDs []string) (map[string]int, error) {
	res := make(map[string]int)
	for _, c := range cellIDs {
		res[c] = m.supplies[c]
	}
	return res, nil
}

func (m *mockOccupancyStore) GetCellOccupancy(ctx context.Context, cellID string) (occupancy.CellOccupancy, error) {
	supply := m.supplies[cellID]
	demand := m.demands[cellID]
	return occupancy.NewCellOccupancy(cellID, supply, demand, 1.2, 1.5, 2, time.Now()), nil
}

func (m *mockOccupancyStore) GetMultiCellOccupancy(ctx context.Context, cellIDs []string) (map[string]occupancy.CellOccupancy, error) {
	res := make(map[string]occupancy.CellOccupancy)
	for _, c := range cellIDs {
		supply := m.supplies[c]
		demand := m.demands[c]
		res[c] = occupancy.NewCellOccupancy(c, supply, demand, 1.2, 1.5, 2, time.Now())
	}
	return res, nil
}

func TestOccupancyService_GetCellOccupancy(t *testing.T) {
	store := newMockOccupancyStore()
	store.supplies["cell-1"] = 10
	store.demands["cell-1"] = 20

	h3 := geo.NewH3Service(8, 1)
	svc := application.NewOccupancyService(store, h3, nil)

	ctx := context.Background()

	// Valid cell
	occ, err := svc.GetCellOccupancy(ctx, "cell-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if occ.SupplyCount != 10 {
		t.Errorf("expected supply 10, got %d", occ.SupplyCount)
	}
	if occ.DemandCount != 20 {
		t.Errorf("expected demand 20, got %d", occ.DemandCount)
	}
	if occ.ImbalanceRatio != 2.0 {
		t.Errorf("expected imbalance ratio 2.0, got %f", occ.ImbalanceRatio)
	}
	if occ.Status != occupancy.StatusHot {
		t.Errorf("expected status HOT, got %s", occ.Status)
	}

	// Empty cellID
	_, err = svc.GetCellOccupancy(ctx, "")
	if err == nil {
		t.Error("expected error for empty cellID, got nil")
	}
}

func TestOccupancyService_GetMultiCellOccupancy(t *testing.T) {
	store := newMockOccupancyStore()
	store.supplies["cell-1"] = 10
	store.demands["cell-1"] = 10
	store.supplies["cell-2"] = 10
	store.demands["cell-2"] = 13
	store.supplies["cell-3"] = 5
	store.demands["cell-3"] = 15

	h3 := geo.NewH3Service(8, 1)
	svc := application.NewOccupancyService(store, h3, nil)

	ctx := context.Background()
	res, err := svc.GetMultiCellOccupancy(ctx, []string{"cell-1", "cell-2", "cell-3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res) != 3 {
		t.Fatalf("expected 3 results, got %d", len(res))
	}

	if res["cell-1"].Status != occupancy.StatusNormal {
		t.Errorf("expected cell-1 NORMAL, got %s", res["cell-1"].Status)
	}
	if res["cell-2"].Status != occupancy.StatusWarm {
		t.Errorf("expected cell-2 WARM, got %s", res["cell-2"].Status)
	}
	if res["cell-3"].Status != occupancy.StatusHot {
		t.Errorf("expected cell-3 HOT, got %s", res["cell-3"].Status)
	}
}

func TestOccupancyService_IsCellHot(t *testing.T) {
	store := newMockOccupancyStore()
	store.supplies["hot-cell"] = 2
	store.demands["hot-cell"] = 10
	store.supplies["normal-cell"] = 10
	store.demands["normal-cell"] = 5

	h3 := geo.NewH3Service(8, 1)
	svc := application.NewOccupancyService(store, h3, nil)

	ctx := context.Background()

	isHot, err := svc.IsCellHot(ctx, "hot-cell")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isHot {
		t.Error("expected hot-cell to be hot, got false")
	}

	isNormalHot, err := svc.IsCellHot(ctx, "normal-cell")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isNormalHot {
		t.Error("expected normal-cell not to be hot, got true")
	}
}

func TestOccupancyService_DetectHotCells(t *testing.T) {
	store := newMockOccupancyStore()
	store.supplies["cell-1"] = 10
	store.demands["cell-1"] = 5 // Normal
	store.supplies["cell-2"] = 10
	store.demands["cell-2"] = 13 // Warm
	store.supplies["cell-3"] = 5
	store.demands["cell-3"] = 15 // Hot
	store.supplies["cell-4"] = 0
	store.demands["cell-4"] = 3 // Hot

	h3 := geo.NewH3Service(8, 1)
	svc := application.NewOccupancyService(store, h3, nil)

	ctx := context.Background()
	hotCells, err := svc.DetectHotCells(ctx, []string{"cell-1", "cell-2", "cell-3", "cell-4"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hotCells) != 2 {
		t.Fatalf("expected 2 hot cells, got %d", len(hotCells))
	}

	cellIDs := make(map[string]bool)
	for _, c := range hotCells {
		cellIDs[c.CellID] = true
	}

	if !cellIDs["cell-3"] || !cellIDs["cell-4"] {
		t.Errorf("expected cell-3 and cell-4 in hot cells, got %v", cellIDs)
	}
}

func TestOccupancyService_GetRingOccupancy(t *testing.T) {
	store := newMockOccupancyStore()
	h3 := geo.NewH3Service(8, 1)
	svc := application.NewOccupancyService(store, h3, nil)

	ctx := context.Background()

	// Kochi central cell
	centerCell, _ := h3.CellForLocation(9.9312, 76.2673)

	res, err := svc.GetRingOccupancy(ctx, centerCell, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ring 1 (hollow ring) around an H3 cell contains 6 cells
	if len(res) != 6 {
		t.Fatalf("expected 6 cells in ring 1, got %d", len(res))
	}

	// Empty center cell returns error
	_, err = svc.GetRingOccupancy(ctx, "", 1)
	if err == nil {
		t.Error("expected error for empty centerCell, got nil")
	}

	// Negative ring returns error
	_, err = svc.GetRingOccupancy(ctx, centerCell, -1)
	if err == nil {
		t.Error("expected error for negative ring, got nil")
	}
}

func TestOccupancyService_EmptyCellList(t *testing.T) {
	store := newMockOccupancyStore()
	h3 := geo.NewH3Service(8, 1)
	svc := application.NewOccupancyService(store, h3, nil)

	ctx := context.Background()
	res, err := svc.GetMultiCellOccupancy(ctx, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("expected empty map, got %d items", len(res))
	}
}

func TestOccupancyService_Concurrency(t *testing.T) {
	store := newMockOccupancyStore()
	for i := 0; i < 10; i++ {
		cell := "cell-" + string(rune('a'+i))
		store.supplies[cell] = i * 2
		store.demands[cell] = i * 3
	}

	h3 := geo.NewH3Service(8, 1)
	svc := application.NewOccupancyService(store, h3, nil)

	ctx := context.Background()
	concurrency := 20
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			cell := "cell-" + string(rune('a'+(idx%10)))
			_, err := svc.GetCellOccupancy(ctx, cell)
			if err != nil {
				t.Errorf("concurrent GetCellOccupancy failed: %v", err)
			}
			_, _ = svc.IsCellHot(ctx, cell)
			done <- true
		}(i)
	}

	for i := 0; i < concurrency; i++ {
		<-done
	}
}
