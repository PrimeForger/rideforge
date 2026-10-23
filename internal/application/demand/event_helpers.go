package demand

import (
	"encoding/json"
	"errors"

	"github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/google/uuid"
)

func decodeEventData(envelope events.Envelope, dst any) error {
	raw, err := json.Marshal(envelope.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}

func parseUUID(raw string, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.New("invalid " + field)
	}
	return id, nil
}
