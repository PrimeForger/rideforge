package redis

import (
	"context"
	"encoding/json"

	"github.com/ashadashraf/ride-hail-app/internal/domain/events"
	"github.com/redis/go-redis/v9"
)

type EventBus struct {
	client *redis.Client
	stream string
}

func NewEventBus(addr string, stream string) *EventBus {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &EventBus{
		client: rdb,
		stream: stream,
	}
}

func (e *EventBus) Publish(ctx context.Context, event events.Event) error {

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return e.client.XAdd(ctx, &redis.XAddArgs{
		Stream: e.stream,
		Values: map[string]interface{}{
			"type":    event.Name(),
			"payload": payload,
		},
	}).Err()
}
