package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

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

		if err := handler(ctx, msg.Value); err != nil {

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

			_ = c.dlqProducer.Publish(ctx, "dlq", data)

			// commit so it doesn't retry forever
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return err
			}

			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
