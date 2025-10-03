package drivingadapters

import (
	"context"
	"log"
	"time"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
	"github.com/segmentio/kafka-go"
)

// EventConsumerAdapter is responsible for consuming events from Kafka
// and translating them into domain events for the application layer
type EventConsumerAdapter struct {
	reader       *kafka.Reader
	eventHandler domain.EventHandler
}

// NewEventConsumerAdapter creates a new EventConsumerAdapter
func NewEventConsumerAdapter(brokerAddress, topic, groupID string, eventHandler domain.EventHandler) *EventConsumerAdapter {
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

	return &EventConsumerAdapter{
		reader:       reader,
		eventHandler: eventHandler,
	}
}

// Start begins consuming events from the message broker
func (adapter *EventConsumerAdapter) Start(ctx context.Context) {
	log.Println("Starting event consumer adapter...")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Event consumer adapter stopping...")
			adapter.Close()
			return
		default:
			// Create a context with timeout for reading messages
			readCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			
			// Fetch the next message from Kafka
			msg, err := adapter.reader.ReadMessage(readCtx)
			cancel()
			
			if err != nil {
				log.Printf("Error reading message: %v", err)
				// Add backoff for coordinator errors
				if isCoordinatorError(err) {
					log.Println("Group coordinator not available, retrying in 5 seconds...")
					time.Sleep(5 * time.Second)
				}
				continue
			}

			// Translate Kafka message to domain event
			event := adapter.translateMessage(msg)
			
			// Handle the event through the application layer
			if err := adapter.eventHandler.HandleEvent(event); err != nil {
				log.Printf("Error handling event: %v", err)
			}
		}
	}
}

// translateMessage converts a Kafka message to a domain event
func (adapter *EventConsumerAdapter) translateMessage(msg kafka.Message) domain.Event {
	return domain.Event{
		Key:       string(msg.Key),
		Value:     string(msg.Value),
		Timestamp: time.Now(),
	}
}

// isCoordinatorError checks if the error is related to group coordinator
func isCoordinatorError(err error) bool {
	return err != nil && (
		err.Error() == "Group Coordinator Not Available" ||
		err.Error() == "[15] Group Coordinator Not Available")
}

// Close closes the Kafka reader
func (adapter *EventConsumerAdapter) Close() error {
	if adapter.reader != nil {
		return adapter.reader.Close()
	}
	return nil
}