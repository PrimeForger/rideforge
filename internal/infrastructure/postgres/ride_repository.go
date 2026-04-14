package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

func (r *RideRepository) SaveTx(ctx context.Context, tx *sql.Tx, rideEntity *ride.Ride) error {
	var ErrOptimisticLockConflict = errors.New("optimistic lock conflict")

	// CASE 1: New Ride
	if rideEntity.Version == 0 {

		query := `
		INSERT INTO rides (id, rider_id, driver_id, status, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err := tx.ExecContext(ctx, query,
			rideEntity.ID,
			rideEntity.RiderID,
			rideEntity.DriverID,
			string(rideEntity.Status),
			rideEntity.Version,
			rideEntity.CreatedAt,
			rideEntity.UpdatedAt,
		)

		return err
	}

	// CASE 2: Existing Ride -> Optimistic Lock Update
	query := `
	UPDATE rides
	SET driver_id = $1,
		status = $2,
		version = version + 1,
		updated_at = $3
	WHERE id = $4 AND version = $5
	`

	res, err := tx.ExecContext(ctx, query,
		rideEntity.DriverID,
		string(rideEntity.Status),
		time.Now(),
		rideEntity.ID,
		rideEntity.Version,
	)

	if err != nil {
		return nil
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil
	}

	if rows == 0 {
		return ErrOptimisticLockConflict
	}

	// Update in-memory version after success
	rideEntity.Version++

	return nil
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

func (r *RideRepository) GetByIDTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*ride.Ride, error) {

	query := `
	SELECT id, rider_id, driver_id, status, version, created_at, updated_at
	FROM rides WHERE id = $1
	`

	row := tx.QueryRowContext(ctx, query, id)

	var rideEntity ride.Ride
	var status string
	var driverID sql.NullString

	err := row.Scan(
		&rideEntity.ID,
		&rideEntity.RiderID,
		&driverID,
		&status,
		&rideEntity.Version,
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
