package domain

import (
	"testing"
	"time"
)

func TestNewBatchCreatedEvent(t *testing.T) {
	// Create a test batch
	batch := NewBatch("test-batch-1", "prod-123")
	
	// Create the event
	event := NewBatchCreatedEvent(batch)
	
	// Verify event properties
	if event.EventType != BatchEventCreated {
		t.Errorf("Expected event type %s, got %s", BatchEventCreated, event.EventType)
	}
	
	if event.BatchID != batch.ID {
		t.Errorf("Expected batch ID %s, got %s", batch.ID, event.BatchID)
	}
	
	if event.ProductID != batch.ProductID {
		t.Errorf("Expected product ID %s, got %s", batch.ProductID, event.ProductID)
	}
	
	if event.Batch == nil {
		t.Error("Expected batch to be included in event")
	}
	
	if event.OrderID != nil {
		t.Error("Expected order ID to be nil for batch created event")
	}
	
	if event.ItemDetails != nil {
		t.Error("Expected item details to be nil for batch created event")
	}
	
	// Verify timestamp is recent (within last minute)
	if time.Since(event.Timestamp) > time.Minute {
		t.Error("Event timestamp should be recent")
	}
}

func TestNewBatchItemAddedEvent(t *testing.T) {
	// Create a test batch and add an item
	batch := NewBatch("test-batch-1", "prod-123")
	orderID := "order-456"
	err := batch.AddItem(orderID, "prod-123", 5, "allocated")
	if err != nil {
		t.Fatalf("Failed to add item to batch: %v", err)
	}
	
	// Get the added item
	item, err := batch.GetItemByOrderID(orderID)
	if err != nil {
		t.Fatalf("Failed to get item from batch: %v", err)
	}
	
	// Create the event
	event := NewBatchItemAddedEvent(batch, orderID, item)
	
	// Verify event properties
	if event.EventType != BatchEventItemAdded {
		t.Errorf("Expected event type %s, got %s", BatchEventItemAdded, event.EventType)
	}
	
	if event.BatchID != batch.ID {
		t.Errorf("Expected batch ID %s, got %s", batch.ID, event.BatchID)
	}
	
	if event.OrderID == nil || *event.OrderID != orderID {
		t.Errorf("Expected order ID %s, got %v", orderID, event.OrderID)
	}
	
	if event.ItemDetails == nil {
		t.Error("Expected item details to be included in event")
	} else {
		if event.ItemDetails.OrderID != orderID {
			t.Errorf("Expected item order ID %s, got %s", orderID, event.ItemDetails.OrderID)
		}
		if event.ItemDetails.Quantity != 5 {
			t.Errorf("Expected item quantity 5, got %d", event.ItemDetails.Quantity)
		}
	}
}

func TestBatchEventTypes(t *testing.T) {
	expectedTypes := []BatchEventType{
		BatchEventCreated,
		BatchEventItemAdded,
		BatchEventItemRemoved,
		BatchEventItemUpdated,
		BatchEventProcessing,
		BatchEventCompleted,
		BatchEventCancelled,
		BatchEventDamaged,
	}
	
	expectedValues := []string{
		"batch.created",
		"batch.item_added",
		"batch.item_removed",
		"batch.item_updated",
		"batch.processing_started",
		"batch.completed",
		"batch.cancelled",
		"batch.marked_damaged",
	}
	
	for i, eventType := range expectedTypes {
		if string(eventType) != expectedValues[i] {
			t.Errorf("Expected event type %s, got %s", expectedValues[i], string(eventType))
		}
	}
}