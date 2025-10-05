package application

import (
	"fmt"
	"log"
	"time"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
)

// BatchService handles business logic for batch operations
type BatchService struct {
	batchRepo      domain.BatchRepository
	eventPublisher domain.BatchEventPublisher
}

// NewBatchService creates a new BatchService
func NewBatchService(batchRepo domain.BatchRepository, eventPublisher domain.BatchEventPublisher) *BatchService {
	return &BatchService{
		batchRepo:      batchRepo,
		eventPublisher: eventPublisher,
	}
}

// AddOrderToBatch adds an order to an appropriate batch
func (s *BatchService) AddOrderToBatch(orderID, productID string, quantity int, status string) (*domain.Batch, error) {
	log.Printf("Adding order %s to batch for product %s (quantity: %d, status: %s)", 
		orderID, productID, quantity, status)

	// Try to find an existing pending batch for this product
	batch, err := s.batchRepo.FindPendingBatchForProduct(productID)
	isNewBatch := false
	if err != nil {
		// No pending batch found, create a new one
		batchID := s.generateBatchID(productID)
		batch = domain.NewBatch(batchID, productID)
		isNewBatch = true
		log.Printf("Created new batch %s for product %s", batchID, productID)
	} else {
		log.Printf("Found existing pending batch %s for product %s", batch.ID, productID)
	}

	// Add the order to the batch
	if err := batch.AddItem(orderID, productID, quantity, status); err != nil {
		return nil, fmt.Errorf("failed to add order to batch: %w", err)
	}

	// Save the batch
	if err := s.batchRepo.Save(batch); err != nil {
		return nil, fmt.Errorf("failed to save batch: %w", err)
	}

	// Publish events
	if isNewBatch {
		// Publish batch created event
		if err := s.eventPublisher.PublishBatchEvent(domain.NewBatchCreatedEvent(batch)); err != nil {
			log.Printf("Failed to publish batch created event: %v", err)
		}
	}

	// Get the added item for the event
	item, err := batch.GetItemByOrderID(orderID)
	if err != nil {
		log.Printf("Failed to get item for event publishing: %v", err)
	} else {
		// Publish item added event
		if err := s.eventPublisher.PublishBatchEvent(domain.NewBatchItemAddedEvent(batch, orderID, item)); err != nil {
			log.Printf("Failed to publish batch item added event: %v", err)
		}
	}

	log.Printf("Successfully added order %s to batch %s", orderID, batch.ID)
	return batch, nil
}

// RemoveOrderFromBatch removes an order from its batch
func (s *BatchService) RemoveOrderFromBatch(orderID string) error {
	log.Printf("Removing order %s from batch", orderID)

	// Find the batch containing this order
	batch, err := s.batchRepo.FindByOrderID(orderID)
	if err != nil {
		return fmt.Errorf("failed to find batch for order %s: %w", orderID, err)
	}

	// Remove the order from the batch
	if err := batch.RemoveItem(orderID); err != nil {
		return fmt.Errorf("failed to remove order from batch: %w", err)
	}

	// Publish item removed event
	if err := s.eventPublisher.PublishBatchEvent(domain.NewBatchItemRemovedEvent(batch, orderID)); err != nil {
		log.Printf("Failed to publish batch item removed event: %v", err)
	}

	// If batch is empty, delete it; otherwise save the updated batch
	if batch.IsEmpty() {
		log.Printf("Batch %s is now empty, deleting it", batch.ID)
		if err := s.batchRepo.Delete(batch.ID); err != nil {
			return fmt.Errorf("failed to delete empty batch: %w", err)
		}
	} else {
		if err := s.batchRepo.Save(batch); err != nil {
			return fmt.Errorf("failed to save updated batch: %w", err)
		}
	}

	log.Printf("Successfully removed order %s from batch %s", orderID, batch.ID)
	return nil
}

// UpdateOrderStatus updates the status of an order within its batch
func (s *BatchService) UpdateOrderStatus(orderID, status string) error {
	log.Printf("Updating order %s status to %s", orderID, status)

	// Find the batch containing this order
	batch, err := s.batchRepo.FindByOrderID(orderID)
	if err != nil {
		return fmt.Errorf("failed to find batch for order %s: %w", orderID, err)
	}

	// Update the order status
	if err := batch.UpdateItemStatus(orderID, status); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Save the updated batch
	if err := s.batchRepo.Save(batch); err != nil {
		return fmt.Errorf("failed to save updated batch: %w", err)
	}

	// Get the updated item for the event
	item, err := batch.GetItemByOrderID(orderID)
	if err != nil {
		log.Printf("Failed to get updated item for event publishing: %v", err)
	} else {
		// Publish item updated event
		if err := s.eventPublisher.PublishBatchEvent(domain.NewBatchItemUpdatedEvent(batch, orderID, item)); err != nil {
			log.Printf("Failed to publish batch item updated event: %v", err)
		}
	}

	log.Printf("Successfully updated order %s status to %s in batch %s", orderID, status, batch.ID)
	return nil
}

// ProcessBatch starts processing a batch
func (s *BatchService) ProcessBatch(batchID string) error {
	log.Printf("Starting to process batch %s", batchID)

	batch, err := s.batchRepo.FindByID(batchID)
	if err != nil {
		return fmt.Errorf("failed to find batch %s: %w", batchID, err)
	}

	if err := batch.StartProcessing(); err != nil {
		return fmt.Errorf("failed to start processing batch: %w", err)
	}

	if err := s.batchRepo.Save(batch); err != nil {
		return fmt.Errorf("failed to save batch: %w", err)
	}

	// Publish processing started event
	if err := s.eventPublisher.PublishBatchEvent(domain.NewBatchProcessingStartedEvent(batch)); err != nil {
		log.Printf("Failed to publish batch processing started event: %v", err)
	}

	log.Printf("Successfully started processing batch %s", batchID)
	return nil
}

// CompleteBatch marks a batch as completed
func (s *BatchService) CompleteBatch(batchID string) error {
	log.Printf("Completing batch %s", batchID)

	batch, err := s.batchRepo.FindByID(batchID)
	if err != nil {
		return fmt.Errorf("failed to find batch %s: %w", batchID, err)
	}

	if err := batch.Complete(); err != nil {
		return fmt.Errorf("failed to complete batch: %w", err)
	}

	if err := s.batchRepo.Save(batch); err != nil {
		return fmt.Errorf("failed to save batch: %w", err)
	}

	// Publish batch completed event
	if err := s.eventPublisher.PublishBatchEvent(domain.NewBatchCompletedEvent(batch)); err != nil {
		log.Printf("Failed to publish batch completed event: %v", err)
	}

	log.Printf("Successfully completed batch %s", batchID)
	return nil
}

// CancelBatch cancels a batch
func (s *BatchService) CancelBatch(batchID string) error {
	log.Printf("Cancelling batch %s", batchID)

	batch, err := s.batchRepo.FindByID(batchID)
	if err != nil {
		return fmt.Errorf("failed to find batch %s: %w", batchID, err)
	}

	if err := batch.Cancel(); err != nil {
		return fmt.Errorf("failed to cancel batch: %w", err)
	}

	if err := s.batchRepo.Save(batch); err != nil {
		return fmt.Errorf("failed to save batch: %w", err)
	}

	// Publish batch cancelled event
	if err := s.eventPublisher.PublishBatchEvent(domain.NewBatchCancelledEvent(batch)); err != nil {
		log.Printf("Failed to publish batch cancelled event: %v", err)
	}

	log.Printf("Successfully cancelled batch %s", batchID)
	return nil
}

// MarkBatchAsDamaged marks a batch as damaged
func (s *BatchService) MarkBatchAsDamaged(batchID string) error {
	log.Printf("Marking batch %s as damaged", batchID)

	batch, err := s.batchRepo.FindByID(batchID)
	if err != nil {
		return fmt.Errorf("failed to find batch %s: %w", batchID, err)
	}

	if err := batch.MarkAsDamaged(); err != nil {
		return fmt.Errorf("failed to mark batch as damaged: %w", err)
	}

	if err := s.batchRepo.Save(batch); err != nil {
		return fmt.Errorf("failed to save batch: %w", err)
	}

	// Publish batch damaged event
	if err := s.eventPublisher.PublishBatchEvent(domain.NewBatchDamagedEvent(batch)); err != nil {
		log.Printf("Failed to publish batch damaged event: %v", err)
	}

	log.Printf("Successfully marked batch %s as damaged", batchID)
	return nil
}

// GetBatchByOrderID retrieves the batch containing a specific order
func (s *BatchService) GetBatchByOrderID(orderID string) (*domain.Batch, error) {
	return s.batchRepo.FindByOrderID(orderID)
}

// GetBatchesByProductID retrieves all batches for a specific product
func (s *BatchService) GetBatchesByProductID(productID string) ([]*domain.Batch, error) {
	return s.batchRepo.FindByProductID(productID)
}

// GetBatchesByStatus retrieves all batches with a specific status
func (s *BatchService) GetBatchesByStatus(status domain.BatchStatus) ([]*domain.Batch, error) {
	return s.batchRepo.FindByStatus(status)
}

// GetAllBatches retrieves all batches
func (s *BatchService) GetAllBatches() ([]*domain.Batch, error) {
	return s.batchRepo.GetAll()
}

// generateBatchID generates a unique batch ID
func (s *BatchService) generateBatchID(productID string) string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("BATCH-%s-%s", productID, timestamp)
}