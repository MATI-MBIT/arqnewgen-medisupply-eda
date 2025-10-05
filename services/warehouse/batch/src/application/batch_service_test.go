package application

import (
	"testing"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
	drivenadapters "github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/infrastructure/driven-adapters"
)

func TestBatchService_AddOrderToBatch(t *testing.T) {
	// Setup
	repo := drivenadapters.NewBatchMemoryRepository()
	mockPublisher := domain.NewMockBatchEventPublisher()
	service := NewBatchService(repo, mockPublisher)

	// Test data
	orderID := "order-123"
	productID := "product-456"
	quantity := 10
	status := "allocated"

	// Execute
	batch, err := service.AddOrderToBatch(orderID, productID, quantity, status)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if batch == nil {
		t.Fatal("Expected batch to be created, got nil")
	}

	if batch.ProductID != productID {
		t.Errorf("Expected product ID %s, got %s", productID, batch.ProductID)
	}

	if batch.Status != domain.BatchStatusPending {
		t.Errorf("Expected batch status %s, got %s", domain.BatchStatusPending, batch.Status)
	}

	if len(batch.Items) != 1 {
		t.Errorf("Expected 1 item in batch, got %d", len(batch.Items))
	}

	item := batch.Items[0]
	if item.OrderID != orderID {
		t.Errorf("Expected order ID %s, got %s", orderID, item.OrderID)
	}

	if item.Quantity != quantity {
		t.Errorf("Expected quantity %d, got %d", quantity, item.Quantity)
	}

	if item.Status != status {
		t.Errorf("Expected status %s, got %s", status, item.Status)
	}
}

func TestBatchService_AddMultipleOrdersToSameBatch(t *testing.T) {
	// Setup
	repo := drivenadapters.NewBatchMemoryRepository()
	mockPublisher := domain.NewMockBatchEventPublisher()
	service := NewBatchService(repo, mockPublisher)

	productID := "product-456"

	// Add first order
	batch1, err := service.AddOrderToBatch("order-1", productID, 5, "allocated")
	if err != nil {
		t.Fatalf("Failed to add first order: %v", err)
	}

	// Add second order (should go to same batch)
	batch2, err := service.AddOrderToBatch("order-2", productID, 3, "allocated")
	if err != nil {
		t.Fatalf("Failed to add second order: %v", err)
	}

	// Both should be the same batch
	if batch1.ID != batch2.ID {
		t.Errorf("Expected orders to be in same batch, got different batch IDs: %s vs %s", batch1.ID, batch2.ID)
	}

	if len(batch2.Items) != 2 {
		t.Errorf("Expected 2 items in batch, got %d", len(batch2.Items))
	}

	if batch2.TotalItems != 2 {
		t.Errorf("Expected total items to be 2, got %d", batch2.TotalItems)
	}
}

func TestBatchService_RemoveOrderFromBatch(t *testing.T) {
	// Setup
	repo := drivenadapters.NewBatchMemoryRepository()
	mockPublisher := domain.NewMockBatchEventPublisher()
	service := NewBatchService(repo, mockPublisher)

	// Add order to batch
	orderID := "order-123"
	productID := "product-456"
	batch, err := service.AddOrderToBatch(orderID, productID, 10, "allocated")
	if err != nil {
		t.Fatalf("Failed to add order: %v", err)
	}

	// Remove order from batch
	err = service.RemoveOrderFromBatch(orderID)
	if err != nil {
		t.Fatalf("Failed to remove order: %v", err)
	}

	// Verify batch is deleted (since it's empty)
	_, err = repo.FindByID(batch.ID)
	if err == nil {
		t.Error("Expected batch to be deleted after removing last order")
	}
}

func TestBatchService_UpdateOrderStatus(t *testing.T) {
	// Setup
	repo := drivenadapters.NewBatchMemoryRepository()
	mockPublisher := domain.NewMockBatchEventPublisher()
	service := NewBatchService(repo, mockPublisher)

	// Add order to batch
	orderID := "order-123"
	productID := "product-456"
	_, err := service.AddOrderToBatch(orderID, productID, 10, "allocated")
	if err != nil {
		t.Fatalf("Failed to add order: %v", err)
	}

	// Update order status
	newStatus := "shipped"
	err = service.UpdateOrderStatus(orderID, newStatus)
	if err != nil {
		t.Fatalf("Failed to update order status: %v", err)
	}

	// Verify status was updated
	batch, err := service.GetBatchByOrderID(orderID)
	if err != nil {
		t.Fatalf("Failed to get batch: %v", err)
	}

	item, err := batch.GetItemByOrderID(orderID)
	if err != nil {
		t.Fatalf("Failed to get item: %v", err)
	}

	if item.Status != newStatus {
		t.Errorf("Expected status %s, got %s", newStatus, item.Status)
	}
}

func TestBatchService_ProcessBatch(t *testing.T) {
	// Setup
	repo := drivenadapters.NewBatchMemoryRepository()
	mockPublisher := domain.NewMockBatchEventPublisher()
	service := NewBatchService(repo, mockPublisher)

	// Add order to batch
	batch, err := service.AddOrderToBatch("order-123", "product-456", 10, "allocated")
	if err != nil {
		t.Fatalf("Failed to add order: %v", err)
	}

	// Process batch
	err = service.ProcessBatch(batch.ID)
	if err != nil {
		t.Fatalf("Failed to process batch: %v", err)
	}

	// Verify batch status
	updatedBatch, err := service.GetBatchByOrderID("order-123")
	if err != nil {
		t.Fatalf("Failed to get batch: %v", err)
	}

	if updatedBatch.Status != domain.BatchStatusProcessing {
		t.Errorf("Expected batch status %s, got %s", domain.BatchStatusProcessing, updatedBatch.Status)
	}
}

func TestBatchService_CompleteBatch(t *testing.T) {
	// Setup
	repo := drivenadapters.NewBatchMemoryRepository()
	mockPublisher := domain.NewMockBatchEventPublisher()
	service := NewBatchService(repo, mockPublisher)

	// Add order and start processing
	batch, err := service.AddOrderToBatch("order-123", "product-456", 10, "allocated")
	if err != nil {
		t.Fatalf("Failed to add order: %v", err)
	}

	err = service.ProcessBatch(batch.ID)
	if err != nil {
		t.Fatalf("Failed to process batch: %v", err)
	}

	// Complete batch
	err = service.CompleteBatch(batch.ID)
	if err != nil {
		t.Fatalf("Failed to complete batch: %v", err)
	}

	// Verify batch status
	updatedBatch, err := service.GetBatchByOrderID("order-123")
	if err != nil {
		t.Fatalf("Failed to get batch: %v", err)
	}

	if updatedBatch.Status != domain.BatchStatusCompleted {
		t.Errorf("Expected batch status %s, got %s", domain.BatchStatusCompleted, updatedBatch.Status)
	}

	if updatedBatch.ProcessedAt == nil {
		t.Error("Expected ProcessedAt to be set")
	}
}