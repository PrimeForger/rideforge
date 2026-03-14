package postgres

import (
	"context"
	"database/sql"

	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/google/uuid"
)

type RideRepository struct {
	db *sql.DB
}

func NewRideRepository(db *sql.DB) *RideRepository {
	return &RideRepository{db: db}
}

func (r *RideRepository) Save(ctx context.Context, rideEntity *ride.Ride) error {

	query := `
	INSERT INTO rides (id, rider_id, driver_id, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (id) DO UPDATE SET
		driver_id = EXCLUDED.driver_id,
		status = EXCLUDED.status,
		updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		rideEntity.ID,
		rideEntity.RiderID,
		rideEntity.DriverID,
		string(rideEntity.Status),
		rideEntity.CreatedAt,
		rideEntity.UpdatedAt,
	)

	return err
}

func (r *RideRepository) GetByID(ctx context.Context, id uuid.UUID) (*ride.Ride, error) {

	query := `
	SELECT id, rider_id, driver_id, status, created_at, updated_at
	FROM rides WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var rideEntity ride.Ride
	var status string
	var driverID sql.NullString

	err := row.Scan(
		&rideEntity.ID,
		&rideEntity.RiderID,
		&driverID,
		&status,
		&rideEntity.CreatedAt,
		&rideEntity.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if driverID.Valid {
		dID, _ := uuid.Parse(driverID.String)
		rideEntity.DriverID = dID
	}

	rideEntity.Status = ride.Status(status)

	return &rideEntity, nil
}
