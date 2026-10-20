package redis

import (
	"context"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

func newDemandStoreForTest(t *testing.T) (*RedisDemandStore, context.Context) {
	t.Helper()
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("REDIS_TEST_ADDR is not set")
	}
	client, err := NewClient(addr)
	if err != nil {
		t.Fatalf("connect to test Redis: %v", err)
	}
	t.Cleanup(func() { _ = client.GetRaw().Close() })
	return NewRedisDemandStore(client, DemandStoreOptions{DemandTTL: time.Hour}), context.Background()
}

func cleanupDemandTestState(t *testing.T, store *RedisDemandStore, ctx context.Context, rideIDs []uuid.UUID, cells ...string) {
	t.Helper()
	for _, rideID := range rideIDs {
		if err := store.RemoveActiveDemand(ctx, rideID, ""); err != nil {
			t.Errorf("cleanup ride %s: %v", rideID, err)
		}
	}
	for _, cellID := range cells {
		if err := store.client.GetRaw().Del(ctx, h3CellDemandKey(cellID)).Err(); err != nil {
			t.Errorf("cleanup cell %s: %v", cellID, err)
		}
		if err := store.client.GetRaw().SRem(ctx, demandCellRegistryKey, cellID).Err(); err != nil {
			t.Errorf("cleanup registry cell %s: %v", cellID, err)
		}
	}
}

func TestRedisDemandStoreAddRemoveAndReconcile(t *testing.T) {
	store, ctx := newDemandStoreForTest(t)
	targetCell := "test-target-" + uuid.NewString()
	wrongCell := "test-wrong-" + uuid.NewString()
	staleCell := "test-stale-" + uuid.NewString()
	activeID, repairID, staleID := uuid.New(), uuid.New(), uuid.New()
	cleanupDemandTestState(t, store, ctx, []uuid.UUID{activeID, repairID, staleID}, targetCell, wrongCell, staleCell)
	t.Cleanup(func() {
		cleanupDemandTestState(t, store, ctx, []uuid.UUID{activeID, repairID, staleID}, targetCell, wrongCell, staleCell)
	})

	createdAt := time.Now()
	if err := store.AddActiveDemand(ctx, activeID, wrongCell); err != nil {
		t.Fatal(err)
	}
	if err := store.AddActiveDemand(ctx, activeID, targetCell); err != nil {
		t.Fatal(err)
	}
	if err := store.AddActiveDemand(ctx, activeID, targetCell); err != nil {
		t.Fatal(err)
	}
	if got, err := store.GetCellDemand(ctx, targetCell); err != nil || got != 1 {
		t.Fatalf("target demand = %d, error = %v; want 1, nil", got, err)
	}
	if got, err := store.GetCellDemand(ctx, wrongCell); err != nil || got != 0 {
		t.Fatalf("wrong-cell demand = %d, error = %v; want 0, nil", got, err)
	}

	if err := store.AddActiveDemand(ctx, staleID, staleCell); err != nil {
		t.Fatal(err)
	}
	if err := store.AddActiveDemand(ctx, repairID, wrongCell); err != nil {
		t.Fatal(err)
	}
	// Simulate a wrong-cell membership while the ride pointer still names wrongCell.
	if err := store.client.GetRaw().ZAdd(ctx, h3CellDemandKey(targetCell), goredis.Z{Score: float64(createdAt.Unix()), Member: repairID.String()}).Err(); err != nil {
		t.Fatal(err)
	}
	if err := store.ReconcileCellDemand(ctx, targetCell, map[uuid.UUID]time.Time{activeID: createdAt, repairID: createdAt}); err != nil {
		t.Fatalf("reconcile target cell: %v", err)
	}
	if got, err := store.GetCellDemand(ctx, targetCell); err != nil || got != 2 {
		t.Fatalf("target demand after reconciliation = %d, error = %v; want 2, nil", got, err)
	}
	if err := store.ReconcileCellDemand(ctx, staleCell, nil); err != nil {
		t.Fatalf("reconcile stale-only cell: %v", err)
	}
	if cells, err := store.ListDemandCells(ctx); err != nil {
		t.Fatal(err)
	} else {
		for _, cell := range cells {
			if cell == staleCell || cell == wrongCell {
				t.Fatalf("empty cell %s remains in demand registry", cell)
			}
		}
	}
	if err := store.RemoveActiveDemand(ctx, activeID, ""); err != nil {
		t.Fatal(err)
	}
	if err := store.RemoveActiveDemand(ctx, activeID, ""); err != nil {
		t.Fatalf("duplicate remove: %v", err)
	}
	if err := store.RemoveActiveDemand(ctx, repairID, ""); err != nil {
		t.Fatal(err)
	}
	if got, err := store.GetCellDemand(ctx, targetCell); err != nil || got != 0 {
		t.Fatalf("demand after removal = %d, error = %v; want 0, nil", got, err)
	}
}

func TestRedisDemandStoreConcurrentUpdates(t *testing.T) {
	store, ctx := newDemandStoreForTest(t)
	cellID := "test-concurrent-" + uuid.NewString()
	rideIDs := make([]uuid.UUID, 64)
	for i := range rideIDs {
		rideIDs[i] = uuid.New()
	}
	t.Cleanup(func() { cleanupDemandTestState(t, store, ctx, rideIDs, cellID) })

	var wg sync.WaitGroup
	for _, rideID := range rideIDs {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			if err := store.AddActiveDemand(ctx, id, cellID); err != nil {
				t.Errorf("concurrent add: %v", err)
			}
		}(rideID)
	}
	wg.Wait()
	if got, err := store.GetCellDemand(ctx, cellID); err != nil || got != len(rideIDs) {
		t.Fatalf("concurrent demand = %d, error = %v; want %d, nil", got, err, len(rideIDs))
	}

	for _, rideID := range rideIDs {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			if err := store.RemoveActiveDemand(ctx, id, ""); err != nil {
				t.Errorf("concurrent remove: %v", err)
			}
		}(rideID)
	}
	wg.Wait()
	if got, err := store.GetCellDemand(ctx, cellID); err != nil || got != 0 {
		t.Fatalf("demand after concurrent removal = %d, error = %v; want 0, nil", got, err)
	}
}

func TestRedisOccupancyReadsLoadScriptsOnColdCache(t *testing.T) {
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("REDIS_TEST_ADDR is not set")
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil || !(strings.EqualFold(host, "localhost") || (net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback())) {
		t.Fatalf("cold-cache integration test requires a loopback Redis test server, got %q", addr)
	}

	store, ctx := newDemandStoreForTest(t)
	cellID := "test-cold-cache-" + uuid.NewString()
	badCellID := "test-cold-cache-wrong-type-" + uuid.NewString()
	rideID := uuid.New()
	driverID := uuid.New()
	supplyKey := h3CellDriversKey(cellID)
	badSupplyKey := h3CellDriversKey(badCellID)
	badDemandKey := h3CellDemandKey(badCellID)
	cleanupDemandTestState(t, store, ctx, []uuid.UUID{rideID}, cellID, badCellID)
	t.Cleanup(func() {
		cleanupDemandTestState(t, store, ctx, []uuid.UUID{rideID}, cellID, badCellID)
		_ = store.client.GetRaw().Del(ctx, supplyKey, badSupplyKey, badDemandKey).Err()
	})
	if err := store.client.GetRaw().SAdd(ctx, supplyKey, driverID.String()).Err(); err != nil {
		t.Fatal(err)
	}
	if err := store.AddActiveDemand(ctx, rideID, cellID); err != nil {
		t.Fatal(err)
	}
	if err := store.client.GetRaw().ScriptFlush(ctx).Err(); err != nil {
		t.Fatalf("flush Redis script cache: %v", err)
	}

	occupancyStore := NewRedisOccupancyStore(store.client, OccupancyStoreOptions{DemandTTL: time.Hour})
	got, err := occupancyStore.GetCellOccupancy(ctx, cellID)
	if err != nil {
		t.Fatalf("GetCellOccupancy with cold script cache: %v", err)
	}
	if got.SupplyCount != 1 || got.DemandCount != 1 {
		t.Fatalf("occupancy = supply %d demand %d, want 1 and 1", got.SupplyCount, got.DemandCount)
	}

	if err := store.client.GetRaw().ScriptFlush(ctx).Err(); err != nil {
		t.Fatalf("flush Redis script cache before multi-cell read: %v", err)
	}
	occupancies, err := occupancyStore.GetMultiCellOccupancy(ctx, []string{cellID})
	if err != nil {
		t.Fatalf("GetMultiCellOccupancy with cold script cache: %v", err)
	}
	if got := occupancies[cellID]; got.SupplyCount != 1 || got.DemandCount != 1 {
		t.Fatalf("multi-cell occupancy = supply %d demand %d, want 1 and 1", got.SupplyCount, got.DemandCount)
	}

	if err := store.client.GetRaw().ScriptFlush(ctx).Err(); err != nil {
		t.Fatalf("flush Redis script cache before multi-demand read: %v", err)
	}
	demands, err := store.GetMultiCellDemand(ctx, []string{cellID})
	if err != nil {
		t.Fatalf("GetMultiCellDemand with cold script cache: %v", err)
	}
	if demands[cellID] != 1 {
		t.Fatalf("multi-cell demand = %d, want 1", demands[cellID])
	}

	if err := store.client.GetRaw().Set(ctx, badSupplyKey, "wrong-type", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := occupancyStore.GetMultiCellSupply(ctx, []string{cellID, badCellID}); err == nil {
		t.Fatal("GetMultiCellSupply succeeded after a cell supply key had the wrong Redis type")
	}
	if _, err := occupancyStore.GetMultiCellOccupancy(ctx, []string{cellID, badCellID}); err == nil {
		t.Fatal("GetMultiCellOccupancy succeeded after a cell supply key had the wrong Redis type")
	}
	if err := store.client.GetRaw().Set(ctx, badDemandKey, "wrong-type", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetMultiCellDemand(ctx, []string{cellID, badCellID}); err == nil {
		t.Fatal("GetMultiCellDemand succeeded after a cell demand key had the wrong Redis type")
	}
}
