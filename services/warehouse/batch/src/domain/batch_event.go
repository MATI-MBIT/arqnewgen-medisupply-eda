package domain

import (
	"time"
)

// BatchEventType represents the type of batch event
type BatchEventType string

const (
	BatchEventCreated       BatchEventType = "batch.created"
	BatchEventItemAdded     BatchEventType = "batch.item_added"
	BatchEventItemRemoved   BatchEventType = "batch.item_removed"
	BatchEventItemUpdated   BatchEventType = "batch.item_updated"
	BatchEventProcessing    BatchEventType = "batch.processing_started"
	BatchEventCompleted     BatchEventType = "batch.completed"
	BatchEventCancelled     BatchEventType = "batch.cancelled"
	BatchEventDamaged       BatchEventType = "batch.marked_damaged"
)

// BatchEvent represents a domain event for batch operations
type BatchEvent struct {
	EventType   BatchEventType `json:"event_type"`
	BatchID     string         `json:"batch_id"`
	ProductID   string         `json:"product_id"`
	Batch       *Batch         `json:"batch"`
	OrderID     *string        `json:"order_id,omitempty"`     // For item-specific events
	ItemDetails *BatchItem     `json:"item_details,omitempty"` // For item-specific events
	Timestamp   time.Time      `json:"timestamp"`
}

// NewBatchCreatedEvent creates a new batch created event
func NewBatchCreatedEvent(batch *Batch) *BatchEvent {
	return &BatchEvent{
		EventType: BatchEventCreated,
		BatchID:   batch.ID,
		ProductID: batch.ProductID,
		Batch:     batch,
		Timestamp: time.Now().UTC(),
	}
}

// NewBatchItemAddedEvent creates a new batch item added event
func NewBatchItemAddedEvent(batch *Batch, orderID string, item *BatchItem) *BatchEvent {
	return &BatchEvent{
		EventType:   BatchEventItemAdded,
		BatchID:     batch.ID,
		ProductID:   batch.ProductID,
		Batch:       batch,
		OrderID:     &orderID,
		ItemDetails: item,
		Timestamp:   time.Now().UTC(),
	}
}

// NewBatchItemRemovedEvent creates a new batch item removed event
func NewBatchItemRemovedEvent(batch *Batch, orderID string) *BatchEvent {
	return &BatchEvent{
		EventType: BatchEventItemRemoved,
		BatchID:   batch.ID,
		ProductID: batch.ProductID,
		Batch:     batch,
		OrderID:   &orderID,
		Timestamp: time.Now().UTC(),
	}
}

// NewBatchItemUpdatedEvent creates a new batch item updated event
func NewBatchItemUpdatedEvent(batch *Batch, orderID string, item *BatchItem) *BatchEvent {
	return &BatchEvent{
		EventType:   BatchEventItemUpdated,
		BatchID:     batch.ID,
		ProductID:   batch.ProductID,
		Batch:       batch,
		OrderID:     &orderID,
		ItemDetails: item,
		Timestamp:   time.Now().UTC(),
	}
}

// NewBatchProcessingStartedEvent creates a new batch processing started event
func NewBatchProcessingStartedEvent(batch *Batch) *BatchEvent {
	return &BatchEvent{
		EventType: BatchEventProcessing,
		BatchID:   batch.ID,
		ProductID: batch.ProductID,
		Batch:     batch,
		Timestamp: time.Now().UTC(),
	}
}

// NewBatchCompletedEvent creates a new batch completed event
func NewBatchCompletedEvent(batch *Batch) *BatchEvent {
	return &BatchEvent{
		EventType: BatchEventCompleted,
		BatchID:   batch.ID,
		ProductID: batch.ProductID,
		Batch:     batch,
		Timestamp: time.Now().UTC(),
	}
}

// NewBatchCancelledEvent creates a new batch cancelled event
func NewBatchCancelledEvent(batch *Batch) *BatchEvent {
	return &BatchEvent{
		EventType: BatchEventCancelled,
		BatchID:   batch.ID,
		ProductID: batch.ProductID,
		Batch:     batch,
		Timestamp: time.Now().UTC(),
	}
}

// NewBatchDamagedEvent creates a new batch damaged event
func NewBatchDamagedEvent(batch *Batch) *BatchEvent {
	return &BatchEvent{
		EventType: BatchEventDamaged,
		BatchID:   batch.ID,
		ProductID: batch.ProductID,
		Batch:     batch,
		Timestamp: time.Now().UTC(),
	}
}

// BatchEventPublisher defines the interface for publishing batch events
type BatchEventPublisher interface {
	PublishBatchEvent(event *BatchEvent) error
}