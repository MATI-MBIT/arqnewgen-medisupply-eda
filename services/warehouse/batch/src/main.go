package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/application"
	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/config"
	drivenadapters "github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/infrastructure/driven-adapters"
	drivingadapters "github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/infrastructure/driving-adapters"
)

func main() {
	log.Println("Starting warehouse batch application...")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration from environment variables
	cfg := config.LoadConfig()
	log.Printf("Configuration - Order Events Topic: %s, Group ID: %s, Broker: %s, HTTP Port: %s", 
		cfg.Kafka.OrderEventsTopic, cfg.Kafka.GroupID, cfg.Kafka.BrokerAddress, cfg.HTTP.Port)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize driven adapters (repositories)
	batchRepo := drivenadapters.NewBatchMemoryRepository()
	
	// Initialize application layer (business logic)
	batchService := application.NewBatchService(batchRepo)
	orderService := application.NewOrderService(batchService)

	// Initialize driving adapters
	// OrderEventConsumerAdapter for order events processing
	orderEventConsumerAdapter := drivingadapters.NewOrderEventConsumerAdapter(
		cfg.Kafka.BrokerAddress,
		cfg.Kafka.OrderEventsTopic,
		cfg.Kafka.GroupID,
		orderService,
	)
	
	// ApiServiceAdapter for synchronous HTTP requests
	apiServiceAdapter := drivingadapters.NewApiServiceAdapter(cfg.HTTP.Port, batchService)

	// Start the order event consumer adapter in a goroutine
	go orderEventConsumerAdapter.Start(ctx)

	// Start the HTTP API service adapter in a goroutine
	go apiServiceAdapter.Start(ctx)

	// Set up graceful shutdown
	setupGracefulShutdown(cancel)

	log.Println("Application shut down gracefully.")
}

// setupGracefulShutdown handles OS signals for graceful shutdown
func setupGracefulShutdown(cancel context.CancelFunc) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-sigchan
	log.Println("Shutdown signal received, cancelling context...")

	// Cancel the context to signal goroutines to stop
	cancel()

	// Give goroutines a moment to clean up
	time.Sleep(2 * time.Second)
}