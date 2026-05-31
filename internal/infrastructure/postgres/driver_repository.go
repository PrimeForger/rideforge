package postgres

import (
	"context"
	"database/sql"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/ashadashraf/ride-hail-app/internal/domain/ride"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type DriverRepository struct {
	db *sql.DB
}

func NewDriverRepository(db *sql.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

func (d *DriverRepository) GetAvailableDrivers(ctx context.Context) ([]*driver.Driver, error) {

	query := `SELECT id, status FROM drivers WHERE status = 'ONLINE'`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []*driver.Driver

	for rows.Next() {
		var dr driver.Driver
		var status string

		if err := rows.Scan(&dr.ID, &status); err != nil {
			return nil, err
		}

		dr.Status = driver.Status(status)
		drivers = append(drivers, &dr)
	}

	return drivers, nil
}

func (r *DriverRepository) GetAvailableDriversExcludingTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
) ([]*driver.Driver, error) {

	query := `
	SELECT id, status FROM drivers WHERE status = 'ONLINE'
	AND id NOT IN (
		SELECT driver_id FROM ride_driver_offers WHERE ride_id = $1
	)
	LIMIT 10
	`

	rows, err := tx.QueryContext(ctx, query, rideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []*driver.Driver

	for rows.Next() {
		var d driver.Driver
		if err := rows.Scan(&d.ID, &d.Status); err != nil {
			return nil, err
		}
		drivers = append(drivers, &d)
	}

	return drivers, nil
}

func (r *DriverRepository) GetEligibleDriversTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverIDs []uuid.UUID,
) ([]*driver.Driver, error) {

	if len(driverIDs) == 0 {
		return nil, nil
	}

	query := `
	SELECT d.id, d.status,
	       d.acceptance_rate,
	       d.cancellation_rate,
		   d.timeout_rate,
	       d.rating,
	       d.completed_rides,
	       d.last_assigned_at,
	       d.lat,
	       d.lng
	FROM drivers d
	WHERE d.id = ANY($1)
	AND d.status = 'ONLINE'
	AND NOT EXISTS (
		SELECT 1 FROM ride_driver_offers rdo
		WHERE rdo.ride_id = $2
		AND rdo.driver_id = d.id
	)
	`

	rows, err := tx.QueryContext(ctx, query, pq.Array(driverIDs), rideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []*driver.Driver

	for rows.Next() {
		var d driver.Driver

		err := rows.Scan(
			&d.ID,
			&d.Status,
			&d.AcceptanceRate,
			&d.CancellationRate,
			&d.TimeoutRate,
			&d.Rating,
			&d.CompletedRides,
			&d.LastAssignedAt,
			&d.Lat,
			&d.Lng,
		)
		if err != nil {
			return nil, err
		}

		drivers = append(drivers, &d)
	}

	return drivers, nil
}

func (d *DriverRepository) GetByID(ctx context.Context, id uuid.UUID) (*driver.Driver, error) {

	query := `SELECT id, status FROM drivers WHERE id=$1`

	row := d.db.QueryRowContext(ctx, query, id)

	var dr driver.Driver
	var status string

	if err := row.Scan(&dr.ID, &status); err != nil {
		return nil, err
	}

	dr.Status = driver.Status(status)

	return &dr, nil
}

func (d *DriverRepository) GetByIDTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*driver.Driver, error) {

	query := `SELECT id, status FROM drivers WHERE id=$1`

	row := tx.QueryRowContext(ctx, query, id)

	var dr driver.Driver
	var status string

	if err := row.Scan(&dr.ID, &status); err != nil {
		return nil, err
	}

	dr.Status = driver.Status(status)

	return &dr, nil
}

func (d *DriverRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*driver.Driver, error) {

	query := `SELECT id, status FROM drivers WHERE id = ANY($1)`

	rows, err := d.db.QueryContext(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drivers []*driver.Driver

	for rows.Next() {
		var dr driver.Driver
		var status string

		if err := rows.Scan(&dr.ID, &status); err != nil {
			return nil, err
		}

		dr.Status = driver.Status(status)
		drivers = append(drivers, &dr)
	}

	return drivers, nil
}

func (d *DriverRepository) Save(ctx context.Context, dr *driver.Driver) error {

	query := `
	INSERT INTO drivers (id, status)
	VALUES ($1, $2)
	ON CONFLICT (id) DO UPDATE SET
	status = EXCLUDED.status
	`

	_, err := d.db.ExecContext(ctx, query, dr.ID, string(dr.Status))
	return err
}

func (d *DriverRepository) SaveTx(ctx context.Context, tx *sql.Tx, dr *driver.Driver) error {

	query := `
	INSERT INTO drivers (id, status)
	VALUES ($1, $2)
	ON CONFLICT (id) DO UPDATE SET
	status = EXCLUDED.status
	`

	_, err := tx.ExecContext(ctx, query, dr.ID, string(dr.Status))
	return err
}

func (r *DriverRepository) InsertRideOfferTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
	attempt int,
) error {

	_, err := tx.ExecContext(ctx, `
		INSERT INTO ride_driver_offers (ride_id, driver_id, status, attempt)
		VALUES ($1, $2, $3, $4)
	`, rideID, driverID, ride.OfferStatusOffered, attempt)

	return err
}

func (r *DriverRepository) MarkDriverAcceptedTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

	_, err := tx.ExecContext(ctx, `
		UPDATE ride_driver_offers
		SET status = $3
		WHERE ride_id = $1 AND driver_id = $2
		AND status = $4
	`, rideID, driverID, ride.OfferStatusAccepted, ride.OfferStatusOffered)

	return err
}

func (r *DriverRepository) MarkDriverRejectedTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

	_, err := tx.ExecContext(ctx, `
		UPDATE ride_driver_offers
		SET status = $3
		WHERE ride_id = $1 AND driver_id = $2
		AND status = $4
	`, rideID, driverID, ride.OfferStatusRejected, ride.OfferStatusOffered)

	return err
}

func (r *DriverRepository) MarkDriverTimeoutTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
) error {

	_, err := tx.ExecContext(ctx, `
		UPDATE ride_driver_offers
		SET status = $3
		WHERE ride_id = $1 AND driver_id = $2
		AND status = $4
	`, rideID, driverID, ride.OfferStatusTimeout, ride.OfferStatusOffered)

	return err
}

func (r *DriverRepository) CountRideAttemptsTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
) (int, error) {

	var count int

	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM ride_driver_offers WHERE ride_id = $1`,
		rideID,
	).Scan(&count)

	return count, err
}

func (r *DriverRepository) UpdateOfferStatusTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
	status ride.OfferStatus,
) error {

	_, err := tx.ExecContext(ctx, `
		UPDATE ride_driver_offers
		SET status = $3
		WHERE ride_id = $1 AND driver_id = $2
	`, rideID, driverID, string(status))

	return err
}

func (r *DriverRepository) GetActiveOfferDriversTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	excludeDriverID uuid.UUID,
) ([]uuid.UUID, error) {

	rows, err := tx.QueryContext(ctx, `
		SELECT driver_id
		FROM ride_driver_offers
		WHERE ride_id = $1
		AND driver_id <> $2
		AND status = 'OFFERED'
	`, rideID, excludeDriverID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (r *DriverRepository) ExpireOtherOffersTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	acceptedDriverID uuid.UUID,
) error {

	_, err := tx.ExecContext(ctx, `
		UPDATE ride_driver_offers
		SET status = 'EXPIRED'
		WHERE ride_id = $1
		AND driver_id <> $2
		AND status = 'OFFERED'
	`, rideID, acceptedDriverID)

	return err
}

func (r *DriverRepository) MarkDriverBusyTx(
	ctx context.Context,
	tx *sql.Tx,
	driverID uuid.UUID,
) error {

	_, err := tx.ExecContext(ctx, `
		UPDATE drivers
		SET status = 'BUSY',
			reserved_for_ride = NULL,
			reserved_at = NULL
		WHERE id = $1
	`, driverID)

	return err
}
