package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/google/uuid"
)

func TestRepositoriesAgainstMigratedPostgres(t *testing.T) {
	url := os.Getenv("RIDEFORGE_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("RIDEFORGE_TEST_DATABASE_URL is unset")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := NewDB(ctx, url)
	if err != nil {
		t.Fatalf("connect to migrated test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	createdAt := time.Now().UTC()
	entity := &ride.Ride{
		ID: uuid.New(), RiderID: uuid.New(), Status: ride.StatusRequested,
		PickupLatitude: 12.5, PickupLongitude: 75.1,
		CreatedAt: createdAt, UpdatedAt: createdAt,
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin ride test transaction: %v", err)
	}
	defer tx.Rollback()
	rides := NewRideRepository(db)
	if err := rides.SaveTx(ctx, tx, entity); err != nil {
		t.Fatalf("save ride in migrated schema: %v", err)
	}
	loaded, err := rides.GetByIDTx(ctx, tx, entity.ID)
	if err != nil {
		t.Fatalf("load ride from migrated schema: %v", err)
	}
	if loaded.PickupLatitude != entity.PickupLatitude || loaded.PickupLongitude != entity.PickupLongitude {
		t.Fatalf("pickup round trip = (%v, %v), want (%v, %v)", loaded.PickupLatitude, loaded.PickupLongitude, entity.PickupLatitude, entity.PickupLongitude)
	}

	driverID := uuid.New()
	if _, err := db.ExecContext(ctx, `INSERT INTO drivers (id, status, lat, lng) VALUES ($1, 'ONLINE', 0, 0)`, driverID); err != nil {
		t.Fatalf("insert driver in migrated schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := db.ExecContext(cleanupCtx, `DELETE FROM drivers WHERE id = $1`, driverID); err != nil {
			t.Errorf("clean up test driver: %v", err)
		}
	})
	drivers, err := NewDriverRepository(db).GetByIDs(ctx, []uuid.UUID{driverID})
	if err != nil {
		t.Fatalf("lookup UUID array in migrated schema: %v", err)
	}
	if len(drivers) != 1 || drivers[0].ID != driverID {
		t.Fatalf("driver lookup returned %v, want driver %s", drivers, driverID)
	}
}
