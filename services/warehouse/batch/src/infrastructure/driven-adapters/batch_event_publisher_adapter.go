package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
)

// BatchEventPublisherAdapter implements the BatchEventPublisher interface using Kafka
type BatchEventPublisherAdapter struct {
	writer        *kafka.Writer
	topic         string
	brokerAddress string
}

// NewBatchEventPublisherAdapter creates a new BatchEventPublisherAdapter
func NewBatchEventPublisherAdapter(brokerAddress, topic string) *BatchEventPublisherAdapter {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokerAddress),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false, // Synchronous writes for reliability
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	return &BatchEventPublisherAdapter{
		writer:        writer,
		topic:         topic,
		brokerAddress: brokerAddress,
	}
}

// PublishBatchEvent publishes a batch event to Kafka
func (p *BatchEventPublisherAdapter) PublishBatchEvent(event *domain.BatchEvent) error {
	// Serialize the event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal batch event: %w", err)
	}

	// Create Kafka message
	message := kafka.Message{
		Key:   []byte(event.BatchID), // Use batch ID as partition key
		Value: eventData,
		Headers: []kafka.Header{
			{
				Key:   "event_type",
				Value: []byte(event.EventType),
			},
			{
				Key:   "batch_id",
				Value: []byte(event.BatchID),
			},
			{
				Key:   "product_id",
				Value: []byte(event.ProductID),
			},
			{
				Key:   "timestamp",
				Value: []byte(event.Timestamp.Format(time.RFC3339)),
			},
		},
	}

	// Add order_id header if present
	if event.OrderID != nil {
		message.Headers = append(message.Headers, kafka.Header{
			Key:   "order_id",
			Value: []byte(*event.OrderID),
		})
	}

	// Write message to Kafka with retry logic for topic/partition errors
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		// Check if it's an "Unknown Topic Or Partition" error (Kafka error code 3)
		if isUnknownTopicOrPartitionError(err) {
			log.Printf("Unknown topic or partition error detected for topic '%s', attempting to recreate writer: %v", p.topic, err)
			
			// Close the current writer
			if closeErr := p.writer.Close(); closeErr != nil {
				log.Printf("Warning: failed to close old writer: %v", closeErr)
			}
			
			// Recreate the writer
			p.recreateWriter()
			
			// Wait a moment for topic to be available
			log.Printf("Waiting 2 seconds for topic '%s' to become available...", p.topic)
			time.Sleep(2 * time.Second)
			
			// Retry the write operation with new writer
			log.Printf("Retrying batch event publish for batch %s with recreated writer", event.BatchID)
			retryCtx, retryCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer retryCancel()
			
			if retryErr := p.writer.WriteMessages(retryCtx, message); retryErr != nil {
				log.Printf("Retry failed for batch event %s (batch %s): %v", event.EventType, event.BatchID, retryErr)
				return fmt.Errorf("failed to write batch event to Kafka after retry: %w", retryErr)
			}
			
			log.Printf("Successfully published batch event after writer recreation: %s for batch %s", event.EventType, event.BatchID)
			return nil
		}
		
		return fmt.Errorf("failed to write batch event to Kafka: %w", err)
	}

	log.Printf("Successfully published batch event: %s for batch %s", event.EventType, event.BatchID)
	return nil
}

// recreateWriter creates a new Kafka writer instance
func (p *BatchEventPublisherAdapter) recreateWriter() {
	log.Printf("Recreating Kafka writer for topic %s", p.topic)
	
	p.writer = &kafka.Writer{
		Addr:         kafka.TCP(p.brokerAddress),
		Topic:        p.topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false, // Synchronous writes for reliability
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	
	log.Printf("Kafka writer recreated successfully for topic %s", p.topic)
}

// isUnknownTopicOrPartitionError checks if the error is related to unknown topic or partition
func isUnknownTopicOrPartitionError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for Kafka error code 3 (UnknownTopicOrPartition)
	// This can appear in different error message formats
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "[3] unknown topic or partition") ||
		   strings.Contains(errStr, "unknowntopicorpartition") ||
		   strings.Contains(errStr, "unknown topic or partition") ||
		   strings.Contains(errStr, "topic or partition that does not exist")
}

// Close closes the Kafka writer
func (p *BatchEventPublisherAdapter) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}