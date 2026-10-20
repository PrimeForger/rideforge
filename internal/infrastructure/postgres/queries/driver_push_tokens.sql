-- name: DriverPushTokenUpsertTx :execrows
INSERT INTO driver_push_tokens (
			driver_id, device_id, platform, token, enabled, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, TRUE, NOW(), NOW())
		ON CONFLICT (driver_id, device_id)
		DO UPDATE SET
			platform = EXCLUDED.platform,
			token = EXCLUDED.token,
			enabled = TRUE,
			updated_at = NOW();

-- name: DriverPushTokenGetActiveTokens :many
SELECT token
		FROM driver_push_tokens
		WHERE driver_id = $1
		AND enabled = TRUE;
