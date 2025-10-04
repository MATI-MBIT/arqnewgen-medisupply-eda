package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/oder/src/application"
	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/oder/src/config"
	drivingadapters "github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/oder/src/infrastructure/driving-adapters"
	drivenadapters "github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/oder/src/infrastructure/driven-adapters"
)

func main() {
	log.Println("Starting order management application...")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration from environment variables
	cfg := config.LoadConfig()
	log.Printf("Configuration - Exchange: %s, Consumer Queue: %s, Publisher Queue: %s, HTTP Port: %s", 
		cfg.RabbitMQ.ExchangeName, cfg.RabbitMQ.ConsumerQueueName, cfg.RabbitMQ.PublisherQueueName, cfg.HTTP.Port)
	log.Printf("RabbitMQ URL: %s", cfg.RabbitMQ.URL)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize driven adapters (infrastructure)
	// Order repository for data persistence
	orderRepo := drivenadapters.NewMemoryOrderRepository()
	
	// RabbitMQ publisher for event publishing
	eventPublisher, err := drivenadapters.NewRabbitMQPublisher(
		cfg.RabbitMQ.URL,
		cfg.RabbitMQ.ExchangeName,
		cfg.RabbitMQ.PublisherQueueName,
		cfg.RabbitMQ.PublisherRoutingKey,
	)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ publisher: %v", err)
	}
	defer eventPublisher.Close()

	// Initialize application layer (business logic)
	orderService := application.NewOrderService(orderRepo, eventPublisher)

	// Initialize driving adapters
	// Order consumer adapter for async event processing
	orderConsumerAdapter, err := drivingadapters.NewOrderConsumerAdapter(
		cfg.RabbitMQ.URL,
		cfg.RabbitMQ.ExchangeName,
		cfg.RabbitMQ.ConsumerQueueName,
		cfg.RabbitMQ.ConsumerRoutingKey,
		orderService,
	)
	if err != nil {
		log.Fatalf("Failed to create order consumer adapter: %v", err)
	}
	defer orderConsumerAdapter.Close()
	
	// API service adapter for synchronous HTTP requests
	apiServiceAdapter := drivingadapters.NewApiServiceAdapter(cfg.HTTP.Port, orderService)

	// Start the order consumer adapter in a goroutine
	go orderConsumerAdapter.Start(ctx)

	// Start the HTTP API service adapter in a goroutine
	go apiServiceAdapter.Start(ctx)

	// Set up graceful shutdown
	setupGracefulShutdown(cancel)

	log.Println("Order management application shut down gracefully.")
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