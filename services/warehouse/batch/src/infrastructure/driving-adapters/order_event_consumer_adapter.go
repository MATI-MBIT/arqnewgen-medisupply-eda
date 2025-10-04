package drivingadapters

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
	"github.com/segmentio/kafka-go"
)

// OrderEventConsumerAdapter is responsible for consuming order events from Kafka
// and translating them into domain order events for the application layer
type OrderEventConsumerAdapter struct {
	reader            *kafka.Reader
	orderEventHandler domain.OrderEventHandler
}

// NewOrderEventConsumerAdapter creates a new OrderEventConsumerAdapter
func NewOrderEventConsumerAdapter(brokerAddress, topic, groupID string, orderEventHandler domain.OrderEventHandler) *OrderEventConsumerAdapter {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{brokerAddress},
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafka.LastOffset,
		// Add retry configurations for Kubernetes
		MaxAttempts: 3,
		Dialer: &kafka.Dialer{
			Timeout: 10 * time.Second,
		},
	})

	return &OrderEventConsumerAdapter{
		reader:            reader,
		orderEventHandler: orderEventHandler,
	}
}

// Start begins consuming order events from the message broker
func (adapter *OrderEventConsumerAdapter) Start(ctx context.Context) {
	config := adapter.reader.Config()
	log.Printf("Starting order event consumer adapter with group ID: %s", config.GroupID)
	log.Printf("Consuming from topic: %s, brokers: %v", config.Topic, config.Brokers)
	log.Printf("Waiting for order events... (timeout errors are normal when no messages are available)")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Order event consumer adapter stopping...")
			adapter.Close()
			return
		default:
			// Create a context with timeout for reading messages
			readCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			
			// Fetch the next message from Kafka
			msg, err := adapter.reader.ReadMessage(readCtx)
			cancel()
			
			if err != nil {
				// Only log non-timeout errors to reduce noise
				if !strings.Contains(err.Error(), "context deadline exceeded") {
					log.Printf("Error reading order event message: %v", err)
				}
				// Add backoff for connection errors
				time.Sleep(2 * time.Second)
				continue
			}

			// Translate Kafka message to domain order event
			orderEvent, err := adapter.translateMessage(msg)
			if err != nil {
				log.Printf("Error translating order event message: %v", err)
				continue
			}
			
			// Handle the order event through the application layer
			if err := adapter.orderEventHandler.HandleOrderEvent(orderEvent); err != nil {
				log.Printf("Error handling order event: %v", err)
			}
		}
	}
}

// translateMessage converts a Kafka message to a domain order event
func (adapter *OrderEventConsumerAdapter) translateMessage(msg kafka.Message) (domain.OrderEvent, error) {
	var orderEvent domain.OrderEvent
	
	// Parse the JSON message value
	if err := json.Unmarshal(msg.Value, &orderEvent); err != nil {
		log.Printf("Failed to unmarshal order event JSON: %v", err)
		log.Printf("Message value: %s", string(msg.Value))
		return orderEvent, err
	}
	
	log.Printf("Successfully parsed order event: Type=%s, OrderID=%s", 
		orderEvent.EventType, orderEvent.OrderID)
	
	return orderEvent, nil
}

// Close closes the Kafka reader
func (adapter *OrderEventConsumerAdapter) Close() error {
	if adapter.reader != nil {
		return adapter.reader.Close()
	}
	return nil
}