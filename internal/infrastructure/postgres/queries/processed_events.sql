-- name: ProcessedEventIsProcessed :one
SELECT EXISTS (SELECT 1 FROM processed_events WHERE event_id = $1 AND consumer_name = $2);

-- name: ProcessedEventInsertIfNotExists :execrows
INSERT INTO processed_events (event_id, consumer_name)
	VALUES ($1, $2)
	ON CONFLICT DO NOTHING;
