package application

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/config"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/geo"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/redis"
	"github.com/google/uuid"
)

func TestDriverLocationService_H3IndexTracksAvailability(t *testing.T) {
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		t.Skip("REDIS_TEST_ADDR is not set")
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil || !(strings.EqualFold(host, "localhost") || (net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback())) {
		t.Fatalf("availability integration test requires a loopback Redis test server, got %q", addr)
	}

	client, err := redis.NewClient(addr)
	if err != nil {
		t.Fatalf("connect to test Redis: %v", err)
	}
	defer client.GetRaw().Close()

	ctx := context.Background()
	driverID := uuid.New()
	lat, lng := 9.9312, 76.2673
	h3 := geo.NewH3Service(8, 1)
	cellID, err := h3.CellForLocation(lat, lng)
	if err != nil {
		t.Fatal(err)
	}
	cache := redis.NewDriverCache(client, redis.DriverCacheOptions{})
	index := redis.NewH3DriverIndex(client, redis.H3DriverIndexOptions{DriverCellTTL: time.Minute})
	geoIndex := redis.NewGeoService(client)
	t.Cleanup(func() {
		_, _ = index.RemoveDriver(ctx, driverID)
		_ = geoIndex.RemoveDriver(ctx, driverID)
		_ = cache.MarkOffline(ctx, driverID)
	})

	service := NewDriverLocationService(
		geoIndex,
		cache,
		h3,
		index,
		&config.Config{H3: config.H3Config{Enabled: true}},
	)

	if err := cache.MarkOnline(ctx, driverID, lat, lng); err != nil {
		t.Fatal(err)
	}
	if err := service.updateSpatialIndexes(ctx, driverID, lat, lng); err != nil {
		t.Fatalf("index available driver: %v", err)
	}
	assertDriverIndexed(t, ctx, index, driverID, cellID, true)

	newLat, newLng := 32.7767, -96.7970
	newCellID, err := h3.CellForLocation(newLat, newLng)
	if err != nil {
		t.Fatal(err)
	}
	if newCellID == cellID {
		t.Fatal("test locations unexpectedly resolved to the same H3 cell")
	}
	if err := service.updateSpatialIndexes(ctx, driverID, newLat, newLng); err != nil {
		t.Fatalf("move available driver: %v", err)
	}
	assertDriverIndexed(t, ctx, index, driverID, cellID, false)
	assertDriverIndexed(t, ctx, index, driverID, newCellID, true)

	if err := cache.MarkBusy(ctx, driverID); err != nil {
		t.Fatal(err)
	}
	if err := service.updateSpatialIndexes(ctx, driverID, newLat, newLng); err != nil {
		t.Fatalf("update busy driver location: %v", err)
	}
	assertDriverIndexed(t, ctx, index, driverID, newCellID, false)

	if err := cache.MarkOnline(ctx, driverID, newLat, newLng); err != nil {
		t.Fatal(err)
	}
	if err := service.updateSpatialIndexes(ctx, driverID, newLat, newLng); err != nil {
		t.Fatalf("reindex available driver: %v", err)
	}
	assertDriverIndexed(t, ctx, index, driverID, newCellID, true)
}

func assertDriverIndexed(
	t *testing.T,
	ctx context.Context,
	index *redis.H3DriverIndex,
	driverID uuid.UUID,
	cellID string,
	want bool,
) {
	t.Helper()
	drivers, err := index.GetDriversInCells(ctx, []string{cellID}, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, id := range drivers {
		if id == driverID {
			found = true
			break
		}
	}
	if found != want {
		t.Fatalf("driver H3 membership = %v, want %v", found, want)
	}
}
