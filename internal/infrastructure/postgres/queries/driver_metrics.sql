-- name: DriverMetricsInsertMetricEventTx :execrows
INSERT INTO driver_metric_events (event_id, driver_id, event_type)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING;

-- name: DriverMetricsIncrementOfferedTx :execrows
INSERT INTO driver_metrics (driver_id, offered_count, last_offered_at, updated_at)
		VALUES ($1, 1, NOW(), NOW())
		ON CONFLICT (driver_id)
		DO UPDATE SET
			offered_count = driver_metrics.offered_count + 1,
			last_offered_at = NOW(),
			updated_at = NOW();

-- name: DriverMetricsIncrementAckedTx :execrows
INSERT INTO driver_metrics (driver_id, acked_count, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (driver_id)
		DO UPDATE SET
			acked_count = driver_metrics.acked_count + 1,
			updated_at = NOW();

-- name: DriverMetricsIncrementAcceptedTx :execrows
INSERT INTO driver_metrics (driver_id, accepted_count, last_accepted_at, updated_at)
		VALUES ($1, 1, NOW(), NOW())
		ON CONFLICT (driver_id)
		DO UPDATE SET
			accepted_count = driver_metrics.accepted_count + 1,
			last_accepted_at = NOW(),
			updated_at = NOW();

-- name: DriverMetricsIncrementRejectedTx :execrows
INSERT INTO driver_metrics (driver_id, rejected_count, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (driver_id)
		DO UPDATE SET
			rejected_count = driver_metrics.rejected_count + 1,
			updated_at = NOW();

-- name: DriverMetricsIncrementTimeoutTx :execrows
INSERT INTO driver_metrics (driver_id, timeout_count, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (driver_id)
		DO UPDATE SET
			timeout_count = driver_metrics.timeout_count + 1,
			updated_at = NOW();

-- name: DriverMetricsIncrementCompletedTx :execrows
INSERT INTO driver_metrics (driver_id, completed_rides, last_completed_at, updated_at)
		VALUES ($1, 1, NOW(), NOW())
		ON CONFLICT (driver_id)
		DO UPDATE SET
			completed_rides = driver_metrics.completed_rides + 1,
			last_completed_at = NOW(),
			updated_at = NOW();

-- name: DriverMetricsIncrementCancelledTx :execrows
INSERT INTO driver_metrics (driver_id, cancelled_count, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (driver_id)
		DO UPDATE SET
			cancelled_count = driver_metrics.cancelled_count + 1,
			updated_at = NOW();

-- name: DriverMetricsRecalculateRatesTx :execrows
UPDATE driver_metrics
		SET
			acceptance_rate = CASE
				WHEN offered_count = 0 THEN 1.0
				ELSE accepted_count::DOUBLE PRECISION / offered_count
			END,
			rejection_rate = CASE
				WHEN offered_count = 0 THEN 0.0
				ELSE rejected_count::DOUBLE PRECISION / offered_count
			END,
			timeout_rate = CASE
				WHEN offered_count = 0 THEN 0.0
				ELSE timeout_count::DOUBLE PRECISION / offered_count
			END,
			cancellation_rate = CASE
				WHEN accepted_count = 0 THEN 0.0
				ELSE cancelled_count::DOUBLE PRECISION / accepted_count
			END,
			updated_at = NOW()
		WHERE driver_id = $1;

-- name: DriverMetricsGetMetricsTx :one
SELECT driver_id, acceptance_rate, cancellation_rate, timeout_rate, completed_rides
		FROM driver_metrics
		WHERE driver_id = $1;
