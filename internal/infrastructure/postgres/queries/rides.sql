-- name: RideSave :execrows
INSERT INTO rides (id, rider_id, driver_id, pickup_latitude, pickup_longitude, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (id) DO UPDATE SET
		driver_id = EXCLUDED.driver_id,
		pickup_latitude = EXCLUDED.pickup_latitude,
		pickup_longitude = EXCLUDED.pickup_longitude,
		status = EXCLUDED.status,
		updated_at = EXCLUDED.updated_at;

-- name: RideInsertTx :execrows
INSERT INTO rides (id, rider_id, driver_id, pickup_latitude, pickup_longitude, status, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: RideUpdateVersionedTx :execrows
UPDATE rides
	SET driver_id = $1,
		pickup_latitude = $2,
		pickup_longitude = $3,
		status = $4,
		version = version + 1,
		updated_at = $5
	WHERE id = $6 AND version = $7;

-- name: RideGetByID :one
SELECT id, rider_id, driver_id, pickup_latitude, pickup_longitude, status, created_at, updated_at
	FROM rides WHERE id = $1;

-- name: RideGetByIDTx :one
SELECT id, rider_id, driver_id, pickup_latitude, pickup_longitude, status, version, created_at, updated_at
	FROM rides WHERE id = $1;

-- name: RideGetActiveRides :many
SELECT id, rider_id, driver_id, pickup_latitude, pickup_longitude, status, version, created_at, updated_at
	FROM rides
	WHERE status IN ('REQUESTED', 'MATCHING');
