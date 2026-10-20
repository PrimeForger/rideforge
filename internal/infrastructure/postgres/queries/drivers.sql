-- name: DriverGetAvailableDrivers :many
SELECT id, status, lat, lng FROM drivers WHERE status = 'ONLINE';

-- name: DriverGetAvailableDriversExcludingTx :many
SELECT id, status FROM drivers WHERE status = 'ONLINE'
	AND id NOT IN (
		SELECT driver_id FROM ride_driver_offers WHERE ride_id = $1
	)
	LIMIT 10;

-- name: DriverGetEligibleDriversTx :many
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
	);

-- name: DriverGetByID :one
SELECT id, status FROM drivers WHERE id=$1;

-- name: DriverGetByIDs :many
SELECT id, status FROM drivers WHERE id = ANY($1);

-- name: DriverSave :execrows
INSERT INTO drivers (id, status, lat, lng)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE SET
	status = EXCLUDED.status,
	lat = EXCLUDED.lat,
	lng = EXCLUDED.lng;

-- name: DriverInsertRideOfferTx :execrows
INSERT INTO ride_driver_offers (ride_id, driver_id, status, attempt)
		VALUES ($1, $2, $3, $4);

-- name: DriverUpdatePendingOfferStatusTx :execrows
UPDATE ride_driver_offers
		SET status = $3
		WHERE ride_id = $1 AND driver_id = $2
		AND status = $4;

-- name: DriverCountRideAttemptsTx :one
SELECT COUNT(*) FROM ride_driver_offers WHERE ride_id = $1;

-- name: DriverUpdateOfferStatusTx :execrows
UPDATE ride_driver_offers
		SET status = $3
		WHERE ride_id = $1 AND driver_id = $2;

-- name: DriverGetActiveOfferDriversTx :many
SELECT driver_id
		FROM ride_driver_offers
		WHERE ride_id = $1
		AND driver_id <> $2
		AND status = 'OFFERED';

-- name: DriverExpireOtherOffersTx :execrows
UPDATE ride_driver_offers
		SET status = 'EXPIRED'
		WHERE ride_id = $1
		AND driver_id <> $2
		AND status = 'OFFERED';

-- name: DriverMarkDriverBusyTx :execrows
UPDATE drivers
		SET status = 'BUSY',
			reserved_for_ride = NULL,
			reserved_at = NULL
		WHERE id = $1;

-- name: DriverLockerReserve :execrows
UPDATE drivers
		SET status = 'RESERVED',
			reserved_for_ride = $2,
			reserved_at = NOW()
		WHERE id = $1 AND status = 'ONLINE';

-- name: DriverLockerRelease :execrows
UPDATE drivers 
		SET status = 'ONLINE',
			reserved_for_ride = NULL,
			reserved_at = NULL
		WHERE id = $1
		AND reserved_for_ride = $2
		AND status = 'RESERVED';

-- name: DriverLockerForceRelease :execrows
UPDATE drivers
		SET status = 'ONLINE',
			reserved_for_ride = NULL,
			reserved_at = NULL
		WHERE id = $1
		AND status = 'RESERVED';
