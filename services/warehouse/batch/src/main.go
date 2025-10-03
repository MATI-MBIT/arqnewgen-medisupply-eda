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
	drivingadapters "github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/infrastructure/driving-adapters"
	drivenadapters "github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/infrastructure/driven-adapters"
)

func main() {
	log.Println("Starting warehouse batch application...")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration from environment variables
	cfg := config.LoadConfig()
	log.Printf("Configuration - Topic: %s, Broker: %s, HTTP Port: %s", 
		cfg.Kafka.Topic, cfg.Kafka.BrokerAddress, cfg.HTTP.Port)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize application layer (business logic)
	eventService := application.NewEventService()

	// Initialize driving adapters
	// EventConsumerAdapter for async event processing
	eventConsumerAdapter := drivingadapters.NewEventConsumerAdapter(
		cfg.Kafka.BrokerAddress,
		cfg.Kafka.Topic,
		eventService,
	)
	
	// ApiServiceAdapter for synchronous HTTP requests
	apiServiceAdapter := drivingadapters.NewApiServiceAdapter(cfg.HTTP.Port)

	// Initialize driven adapter (KafkaProducer) - optional for demo
	kafkaProducer := drivenadapters.NewKafkaProducer(cfg.Kafka.BrokerAddress, cfg.Kafka.Topic)

	// Start the event consumer adapter in a goroutine
	go eventConsumerAdapter.Start(ctx)

	// Start the HTTP API service adapter in a goroutine
	go apiServiceAdapter.Start(ctx)

	// Start the producer in a goroutine (for demo purposes)
	go kafkaProducer.StartProducing(ctx)

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