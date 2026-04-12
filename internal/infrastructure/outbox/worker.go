package outbox

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/postgres"
	"github.com/ashadashraf/ride-hail-app/internal/ports"
)

type Worker struct {
	repo      ports.OutboxRepository
	txManager *postgres.TxManager
	producer  *kafka.Producer
}

func NewWorker(repo ports.OutboxRepository, txManager *postgres.TxManager, producer *kafka.Producer) *Worker {
	return &Worker{
		repo:      repo,
		txManager: txManager,
		producer:  producer,
	}
}

func (w *Worker) Start(ctx context.Context) {

	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-ticker.C:
			w.process(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) process(ctx context.Context) {

	err := w.txManager.WithinTx(ctx, func(tx *sql.Tx) error {

		events, err := w.repo.GetUnpublishedTx(ctx, tx)
		if err != nil {
			log.Println("outbox fetch error:", err)
			return err
		}

		for _, e := range events {

			if err := w.producer.PublishRaw(ctx, e.EventType, e.Payload); err != nil {
				log.Println("publish failed:", err)
				return err
			}

			if err := w.repo.MarkPublishedTx(ctx, tx, e.ID); err != nil {
				log.Println("mark published failed:", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Println("outbox worker error:", err)
	}
}
