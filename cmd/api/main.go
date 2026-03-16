package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ashadashraf/ride-hail-app/internal/application/matching"
	rideapp "github.com/ashadashraf/ride-hail-app/internal/application/ride"
	"github.com/ashadashraf/ride-hail-app/internal/bootstrap"
	"github.com/ashadashraf/ride-hail-app/internal/infrastructure/kafka"
	"github.com/ashadashraf/ride-hail-app/internal/server"
)

func main() {

	container, err := bootstrap.NewContainer()
	if err != nil {
		log.Fatal(err)
	}

	srv := server.NewServer(container.RideService)
	srv.RegisterRoutes()

	server := &http.Server{
		Addr:         ":8080",
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	rideProducer := kafka.NewProducer([]string{"localhost:9092"}, "ride.events")
	matchProducer := kafka.NewProducer([]string{"localhost:9092"}, "match.events")
	dlqProducer := kafka.NewProducer([]string{"localhost:9092"}, "ride.events.dlq")

	rideService := rideapp.NewRideService(rideProducer)
	matchingService := matching.NewMatchingService(matchProducer)

	rideID, err := rideService.RequestRide(
		context.Background(),
		"john",
		"airport",
		"hotel",
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Ride requested", rideID)

	consumer := kafka.NewConsumer(
		[]string{"localhost:9092"},
		"ride.events",
		"matching-group",
		dlqProducer,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		log.Println("HTTP server running on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()

		err := consumer.Consume(ctx, matchingService.HandleRideRequested)
		if err != nil && ctx.Err() == nil {
			log.Println("consumer error:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	server.Shutdown(shutdownCtx)

	wg.Wait()

	consumer.Close()
	rideProducer.Close()
	matchProducer.Close()
	dlqProducer.Close()
}
