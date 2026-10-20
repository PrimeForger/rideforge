-- name: OutboxInsert :execrows
INSERT INTO outbox_events (id, aggregate_id, event_type, payload, created_at, published, processed_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: OutboxGetUnpublishedTx :many
SELECT id, aggregate_id, event_type, payload, created_at, published, processed_at
	FROM outbox_events
	WHERE published = false
	ORDER BY created_at ASC
	FOR UPDATE SKIP LOCKED
	LIMIT 50;

-- name: OutboxMarkPublishedTx :execrows
UPDATE outbox_events 
		SET published = true, processed_at = NOW()
		WHERE id = $1;
