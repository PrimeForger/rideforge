package outbox

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID
	AggregateID uuid.UUID
	EventType   string
	Payload     []byte
	CreatedAt   time.Time
	Published   bool
	ProcessedAt *time.Time
}

func NewEvent(aggregateID uuid.UUID, eventType string, payload []byte) *Event {
	return &Event{
		ID:          uuid.New(),
		AggregateID: aggregateID,
		EventType:   eventType,
		Payload:     payload,
		CreatedAt:   time.Now(),
		Published:   false,
		ProcessedAt: nil,
	}
}
