package drivenadapters

import (
	"errors"
	"testing"
	"time"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
)

func TestBatchEventPublisherAdapter_ErrorRecovery(t *testing.T) {
	// This test demonstrates the error recovery logic
	// Note: This is a unit test for the error detection logic only
	// Integration tests with actual Kafka would require a test environment
	
	t.Run("Error Detection", func(t *testing.T) {
		// Test various error formats that should trigger recovery
		errorCases := []struct {
			name  string
			err   error
			shouldRecover bool
		}{
			{
				name:  "Standard Kafka error code 3",
				err:   errors.New("[3] Unknown Topic Or Partition: the request is for a topic or partition that does not exist on this broker"),
				shouldRecover: true,
			},
			{
				name:  "Lowercase version",
				err:   errors.New("[3] unknown topic or partition: topic does not exist"),
				shouldRecover: true,
			},
			{
				name:  "UnknownTopicOrPartition format",
				err:   errors.New("kafka: UnknownTopicOrPartition"),
				shouldRecover: true,
			},
			{
				name:  "Generic topic not found",
				err:   errors.New("topic or partition that does not exist"),
				shouldRecover: true,
			},
			{
				name:  "Different error should not trigger recovery",
				err:   errors.New("connection refused"),
				shouldRecover: false,
			},
			{
				name:  "Different Kafka error should not trigger recovery",
				err:   errors.New("[1] OffsetOutOfRange"),
				shouldRecover: false,
			},
		}

		for _, tc := range errorCases {
			t.Run(tc.name, func(t *testing.T) {
				result := isUnknownTopicOrPartitionError(tc.err)
				if result != tc.shouldRecover {
					t.Errorf("Expected recovery decision %v for error '%v', got %v", 
						tc.shouldRecover, tc.err, result)
				}
			})
		}
	})

	t.Run("Writer Recreation", func(t *testing.T) {
		// Test that writer recreation works
		adapter := NewBatchEventPublisherAdapter("localhost:9092", "test-topic")
		
		// Store original writer reference
		originalWriter := adapter.writer
		
		// Recreate writer
		adapter.recreateWriter()
		
		// Verify new writer was created
		if adapter.writer == originalWriter {
			t.Error("Expected new writer instance after recreation")
		}
		
		if adapter.writer == nil {
			t.Error("Expected writer to be recreated, got nil")
		}
		
		// Verify configuration is preserved
		if adapter.topic != "test-topic" {
			t.Errorf("Expected topic to be preserved as 'test-topic', got '%s'", adapter.topic)
		}
		
		if adapter.brokerAddress != "localhost:9092" {
			t.Errorf("Expected broker address to be preserved as 'localhost:9092', got '%s'", adapter.brokerAddress)
		}
		
		// Clean up
		adapter.Close()
	})
}

func TestBatchEventPublisherAdapter_EventCreation(t *testing.T) {
	// Test that we can create events that would be published
	batch := domain.NewBatch("test-batch-1", "prod-123")
	
	// Test batch created event
	event := domain.NewBatchCreatedEvent(batch)
	if event.EventType != domain.BatchEventCreated {
		t.Errorf("Expected event type %s, got %s", domain.BatchEventCreated, event.EventType)
	}
	
	// Add an item to test item events
	err := batch.AddItem("order-456", "prod-123", 5, "allocated")
	if err != nil {
		t.Fatalf("Failed to add item to batch: %v", err)
	}
	
	item, err := batch.GetItemByOrderID("order-456")
	if err != nil {
		t.Fatalf("Failed to get item: %v", err)
	}
	
	// Test item added event
	itemEvent := domain.NewBatchItemAddedEvent(batch, "order-456", item)
	if itemEvent.EventType != domain.BatchEventItemAdded {
		t.Errorf("Expected event type %s, got %s", domain.BatchEventItemAdded, itemEvent.EventType)
	}
	
	if itemEvent.OrderID == nil || *itemEvent.OrderID != "order-456" {
		t.Errorf("Expected order ID 'order-456', got %v", itemEvent.OrderID)
	}
	
	if itemEvent.ItemDetails == nil {
		t.Error("Expected item details to be included")
	}
	
	// Verify timestamp is recent
	if time.Since(itemEvent.Timestamp) > time.Minute {
		t.Error("Event timestamp should be recent")
	}
}