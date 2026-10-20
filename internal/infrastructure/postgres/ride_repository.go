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

	query := rideSave
	_, err := r.db.ExecContext(ctx, query,
		rideEntity.ID,
		rideEntity.RiderID,
		rideEntity.DriverID,
		rideEntity.PickupLatitude,
		rideEntity.PickupLongitude,
		string(rideEntity.Status),
		rideEntity.CreatedAt,
		rideEntity.UpdatedAt,
	)

	return err
}

var ErrOptimisticLockConflict = errors.New("optimistic lock conflict")

func (r *RideRepository) SaveTx(ctx context.Context, tx *sql.Tx, rideEntity *ride.Ride) error {

	// CASE 1: New Ride
	if rideEntity.Version == 0 {

		query := rideInsertTx
		_, err := tx.ExecContext(ctx, query,
			rideEntity.ID,
			rideEntity.RiderID,
			rideEntity.DriverID,
			rideEntity.PickupLatitude,
			rideEntity.PickupLongitude,
			string(rideEntity.Status),
			rideEntity.Version,
			rideEntity.CreatedAt,
			rideEntity.UpdatedAt,
		)

		return err
	}

	// CASE 2: Existing Ride -> Optimistic Lock Update
	query := rideUpdateVersionedTx

	res, err := tx.ExecContext(ctx, query,
		rideEntity.DriverID,
		rideEntity.PickupLatitude,
		rideEntity.PickupLongitude,
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

	query := rideGetByID

	row := r.db.QueryRowContext(ctx, query, id)

	var rideEntity ride.Ride
	var status string
	var driverID sql.NullString

	err := row.Scan(
		&rideEntity.ID,
		&rideEntity.RiderID,
		&driverID,
		&rideEntity.PickupLatitude,
		&rideEntity.PickupLongitude,
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

	query := rideGetByIDTx

	row := tx.QueryRowContext(ctx, query, id)

	var rideEntity ride.Ride
	var status string
	var driverID sql.NullString

	err := row.Scan(
		&rideEntity.ID,
		&rideEntity.RiderID,
		&driverID,
		&rideEntity.PickupLatitude,
		&rideEntity.PickupLongitude,
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

func (r *RideRepository) GetActiveRides(ctx context.Context) ([]*ride.Ride, error) {
	query := rideGetActiveRides

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []*ride.Ride

	for rows.Next() {
		var rideEntity ride.Ride
		var status string
		var driverID sql.NullString

		err := rows.Scan(
			&rideEntity.ID,
			&rideEntity.RiderID,
			&driverID,
			&rideEntity.PickupLatitude,
			&rideEntity.PickupLongitude,
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
		rides = append(rides, &rideEntity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rides, nil
}
