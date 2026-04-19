package postgres

import (
	"context"
	"database/sql"

	"github.com/ashadashraf/ride-hail-app/internal/domain/driver"
	"github.com/google/uuid"
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

func (r *DriverRepository) InsertRideOfferTx(
	ctx context.Context,
	tx *sql.Tx,
	rideID uuid.UUID,
	driverID uuid.UUID,
	attempt int,
) error {

	_, err := tx.ExecContext(ctx, `
		INSERT INTO ride_driver_offers (ride_id, driver_id, status, attempt)
		VALUES ($1, $2, 'OFFERED', $3)
	`, rideID, driverID, attempt)

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
		SET status = 'REJECTED'
		WHERE ride_id = $1 AND driver_id = $2
		AND status = 'OFFERED'
	`, rideID, driverID)

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
