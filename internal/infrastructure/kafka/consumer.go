package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader      *kafka.Reader
	dlqProducer *Producer
}

func NewConsumer(brokers []string, topic, groupID string, dlq *Producer) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		dlqProducer: dlq,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, []byte) error) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		eventType := extractEventType(msg.Value)

		if err := handler(ctx, msg.Value); err != nil {
			observability.KafkaEventsProcessedTotal.WithLabelValues(eventType, "error").Inc()

			log.Println("handler failed, sending to DLQ:", err)

			dlqEvent := map[string]interface{}{
				"original_topic": msg.Topic,
				"partition":      msg.Partition,
				"offset":         msg.Offset,
				"key":            string(msg.Key),
				"value":          string(msg.Value),
				"error":          err.Error(),
				"failed_at":      time.Now(),
			}

			data, _ := json.Marshal(dlqEvent)

			if err := c.dlqProducer.PublishRaw(ctx, "dlq", data); err != nil {
				log.Println("failed to publish DLQ:", err)
			}

			// commit so it doesn't retry forever
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return err
			}

			continue
		}

		observability.KafkaEventsProcessedTotal.WithLabelValues(eventType, "success").Inc()

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func extractEventType(payload []byte) string {
	var e struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(payload, &e); err != nil {
		return "unknown"
	}

	if e.Type == "" {
		return "unknown"
	}

	return e.Type
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
