package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application/events"
	"github.com/ashadashraf/ride-hail-app/internal/bootstrap"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/outbox"
	"github.com/ashadashraf/ride-hail-app/internal/server"
	"github.com/google/uuid"
)

func main() {

	container, err := bootstrap.NewContainer()
	if err != nil {
		log.Fatal(err)
	}

	srv := server.NewServer(
		container.RideService,
		container.DriverService,
		container.DriverResponseCommandService,
		container.DriverDeviceService,
		container.RealtimeHub,
		container.GeoService,
		container.DriverCache,
	)

	srv.RegisterRoutes()

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	consumer := kafka.NewConsumer(
		[]string{"localhost:9092"},
		"ride.events",
		"matching-group",
		container.DLQProducer,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		log.Println("HTTP server running on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()

		err := consumer.Consume(ctx, func(ctx context.Context, msg []byte) error {

			var envelope events.Envelope
			if err := json.Unmarshal(msg, &envelope); err != nil {
				return err
			}

			return container.TxManager.WithinTx(ctx, func(tx *sql.Tx) error {

				eventID, err := uuid.Parse(envelope.ID)
				if err != nil {
					return err
				}

				// IDEMPOTENCY CHECK
				inserted, err := container.ProcessedEventRepo.InsertIfNotExists(
					ctx,
					tx,
					eventID,
					"matching-service",
				)
				if err != nil {
					return err
				}

				if !inserted {
					log.Println("duplicate event skipped:", envelope.ID)
					return nil
				}

				// ROUTE BASED ON EVENT TYPE
				switch envelope.Type {

				case "ride.requested":
					var data struct {
						RideID string `json:"ride_id"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					rideID, _ := uuid.Parse(data.RideID)

					return container.RideService.StartMatchingTx(ctx, tx, rideID)

				case "ride.accepted":

					var data struct {
						RideID string `json:"ride_id"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					rideID, _ := uuid.Parse(data.RideID)

					return container.RideEventHandler.HandleRideAccepted(ctx, rideID)

				case "matching.started":

					var data struct {
						RideID string `json:"ride_id"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					rideID, _ := uuid.Parse(data.RideID)

					return container.MatchingEngine.HandleMatchingStarted(ctx, tx, rideID)

				case "matching.retry":

					var data struct {
						RideID string `json:"ride_id"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					rideID, _ := uuid.Parse(data.RideID)

					return container.MatchingEngine.HandleMatchingStarted(ctx, tx, rideID)

				case "driver.online":

					var data struct {
						DriverID string  `json:"driver_id"`
						Lat      float64 `json:"lat"`
						Lng      float64 `json:"lng"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					driverID, _ := uuid.Parse(data.DriverID)

					if err := container.GeoService.UpdateDriverLocation(ctx, driverID, data.Lat, data.Lng); err != nil {
						return err
					}

					return container.DriverCache.MarkOnline(ctx, driverID, data.Lat, data.Lng)

				case "driver.offline":

					var data struct {
						DriverID string `json:"driver_id"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					driverID, _ := uuid.Parse(data.DriverID)

					// Remove from GEO
					if err := container.GeoService.RemoveDriver(ctx, driverID); err != nil {
						return err
					}

					if err := container.DriverCache.MarkOffline(ctx, driverID); err != nil {
						return err
					}

					// Force release lock
					if err := container.DriverLocker.ForceRelease(ctx, driverID); err != nil {
						return err
					}

					return nil

				case "driver.offered":

					var data struct {
						RideID   string `json:"ride_id"`
						DriverID string `json:"driver_id"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					rideID, _ := uuid.Parse(data.RideID)
					driverID, _ := uuid.Parse(data.DriverID)
					return container.DriverOfferHandler.HandleDriverOffered(ctx, rideID, driverID)

				case "driver.accepted":

					var data struct {
						RideID   string `json:"ride_id"`
						DriverID string `json:"driver_id"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					rideID, _ := uuid.Parse(data.RideID)
					driverID, _ := uuid.Parse(data.DriverID)

					return container.DriverResponseService.HandleDriverAccepted(ctx, tx, rideID, driverID)

				case "driver.rejected":

					var data struct {
						RideID   string `json:"ride_id"`
						DriverID string `json:"driver_id"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					rideID, _ := uuid.Parse(data.RideID)
					driverID, _ := uuid.Parse(data.DriverID)

					return container.DriverResponseService.HandleDriverRejected(ctx, tx, rideID, driverID)

				case "driver.rejected.processed":

					var data struct {
						RideID   string `json:"ride_id"`
						DriverID string `json:"driver_id"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					rideID, _ := uuid.Parse(data.RideID)
					driverID, _ := uuid.Parse(data.DriverID)

					return container.RideEventHandler.HandleDriverRejected(ctx, rideID, driverID)

				case "driver.timeout":

					var data struct {
						RideID   string `json:"ride_id"`
						DriverID string `json:"driver_id"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					rideID, err := uuid.Parse(data.RideID)
					if err != nil {
						return err
					}

					driverID, err := uuid.Parse(data.DriverID)
					if err != nil {
						return err
					}

					acked, err := container.DriverCache.IsOfferAcked(ctx, rideID, driverID)
					if err != nil {
						return err
					}

					deliveryStatus, err := container.DriverCache.GetOfferDeliveryStatus(ctx, rideID, driverID)
					if err != nil {
						return err
					}

					return container.DriverResponseService.HandleDriverTimeout(ctx, tx, rideID, driverID, acked, string(deliveryStatus))

				case "driver.push_token.updated":

					var data struct {
						DriverID string `json:"driver_id"`
						DeviceID string `json:"device_id"`
						Platform string `json:"platform"`
						Token    string `json:"token"`
					}

					raw, _ := json.Marshal(envelope.Data)
					if err := json.Unmarshal(raw, &data); err != nil {
						return err
					}

					driverID, err := uuid.Parse(data.DriverID)
					if err != nil {
						return err
					}

					return container.DriverCache.AddPushToken(ctx, driverID, data.Token)

				default:
					return nil
				}
			})
		})

		if err != nil && ctx.Err() == nil {
			log.Println("consumer error:", err)
		}
	}()

	outboxWorker := outbox.NewWorker(container.OutboxRepo, container.TxManager, container.RideProducer)

	wg.Add(2)
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
	<-quit

	log.Println("Shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	httpServer.Shutdown(shutdownCtx)

	wg.Wait()

	consumer.Close()
	container.RideProducer.Close()
	container.MatchProducer.Close()
	container.DLQProducer.Close()
}
