package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	appevents "github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/bootstrap"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/observability"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func main() {

	container, err := bootstrap.NewContainer()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = container.Logger.Sync()
	}()

	var shutdownTracing func(context.Context) error

	if container.Config.Observability.TracingEnabled {
		shutdown, err := observability.InitTracing(
			context.Background(),
			container.Config.Observability.ServiceName,
			container.Config.Observability.Environment,
			container.Config.Observability.OTLPEndpoint,
		)
		if err != nil {
			container.Logger.Fatal("failed to initialize tracing", zap.Error(err))
		}

		shutdownTracing = shutdown
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = shutdownTracing(ctx)
		}()
	}

	observability.Register()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := server.NewServer(
		container.RideService,
		container.DriverService,
		container.DriverResponseCommandService,
		container.DriverDeviceService,
		container.RealtimeHub,
		container.GeoService,
		container.DriverCache,
		&container.Config.Realtime,
	)

	srv.RegisterRoutes(mux)

	handler := otelhttp.NewHandler(mux, "http.server")

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	consumer := kafka.NewConsumer(
		[]string{"localhost:9092"},
		"ride.events",
		"matching-group",
		container.DLQProducer,
	)

	outboxWorker := outbox.NewWorker(container.OutboxRepo, container.TxManager, container.RideProducer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()

		container.Logger.Info("http server starting",
			zap.String("addr", ":8080"),
		)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			container.Logger.Fatal("http server failed",
				zap.Error(err),
			)
		}
	}()

	go func() {
		defer wg.Done()

		tracer := otel.Tracer("kafka.consumer")

		err := consumer.Consume(ctx, func(ctx context.Context, msg []byte) error {
			var envelope appevents.Envelope
			if err := json.Unmarshal(msg, &envelope); err != nil {
				return err
			}

			ctx, span := tracer.Start(ctx, "event.process")
			defer span.End()

			span.SetAttributes(
				attribute.String("event.type", envelope.Type),
				attribute.String("event.id", envelope.ID),
				attribute.String("event.aggregate", envelope.Aggregate),
			)

			if err := container.EventRouter.Handle(ctx, envelope); err != nil {
				span.RecordError(err)
				return err
			}

			return nil
		})

		if err != nil && ctx.Err() == nil {
			container.Logger.Error("consumer error",
				zap.Error(err),
			)
		}
	}()

	go func() {
		defer wg.Done()
		container.HeartbeatRecoveryWorker.Start(ctx)
	}()

	go func() {
		defer wg.Done()
		outboxWorker.Start(ctx)
	}()

	go func() {
		defer wg.Done()
		container.TimeoutScheduler.Start(ctx)
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	<-quit

	container.Logger.Info("shutdown signal received")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		container.Logger.Error("http server shutdown failed",
			zap.Error(err),
		)
	}

	if err := consumer.Close(); err != nil {
		container.Logger.Error("consumer close failed",
			zap.Error(err),
		)
	}

	wg.Wait()

	if err := container.RideProducer.Close(); err != nil {
		container.Logger.Error("ride producer close failed",
			zap.Error(err),
		)
	}

	if err := container.MatchProducer.Close(); err != nil {
		container.Logger.Error("match producer close failed",
			zap.Error(err),
		)
	}

	if err := container.DLQProducer.Close(); err != nil {
		container.Logger.Error("dlq producer close failed",
			zap.Error(err),
		)
	}
}
