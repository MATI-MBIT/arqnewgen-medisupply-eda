package drivenadapters

import (
	"fmt"
	"sync"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
)

// BatchMemoryRepository implements BatchRepository using in-memory storage
type BatchMemoryRepository struct {
	batches map[string]*domain.Batch
	mutex   sync.RWMutex
}

// NewBatchMemoryRepository creates a new in-memory batch repository
func NewBatchMemoryRepository() *BatchMemoryRepository {
	return &BatchMemoryRepository{
		batches: make(map[string]*domain.Batch),
		mutex:   sync.RWMutex{},
	}
}

// Save stores or updates a batch
func (r *BatchMemoryRepository) Save(batch *domain.Batch) error {
	if batch == nil {
		return fmt.Errorf("batch cannot be nil")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Create a deep copy to avoid external modifications
	batchCopy := *batch
	itemsCopy := make([]domain.BatchItem, len(batch.Items))
	copy(itemsCopy, batch.Items)
	batchCopy.Items = itemsCopy

	r.batches[batch.ID] = &batchCopy
	return nil
}

// FindByID retrieves a batch by its ID
func (r *BatchMemoryRepository) FindByID(id string) (*domain.Batch, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	batch, exists := r.batches[id]
	if !exists {
		return nil, fmt.Errorf("batch with ID %s not found", id)
	}

	// Return a copy to avoid external modifications
	batchCopy := *batch
	itemsCopy := make([]domain.BatchItem, len(batch.Items))
	copy(itemsCopy, batch.Items)
	batchCopy.Items = itemsCopy

	return &batchCopy, nil
}

// FindByProductID retrieves all batches for a specific product
func (r *BatchMemoryRepository) FindByProductID(productID string) ([]*domain.Batch, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*domain.Batch
	for _, batch := range r.batches {
		if batch.ProductID == productID {
			// Create a copy
			batchCopy := *batch
			itemsCopy := make([]domain.BatchItem, len(batch.Items))
			copy(itemsCopy, batch.Items)
			batchCopy.Items = itemsCopy
			result = append(result, &batchCopy)
		}
	}

	return result, nil
}

// FindByStatus retrieves all batches with a specific status
func (r *BatchMemoryRepository) FindByStatus(status domain.BatchStatus) ([]*domain.Batch, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*domain.Batch
	for _, batch := range r.batches {
		if batch.Status == status {
			// Create a copy
			batchCopy := *batch
			itemsCopy := make([]domain.BatchItem, len(batch.Items))
			copy(itemsCopy, batch.Items)
			batchCopy.Items = itemsCopy
			result = append(result, &batchCopy)
		}
	}

	return result, nil
}

// FindByOrderID finds the batch containing a specific order
func (r *BatchMemoryRepository) FindByOrderID(orderID string) (*domain.Batch, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, batch := range r.batches {
		if batch.HasOrder(orderID) {
			// Create a copy
			batchCopy := *batch
			itemsCopy := make([]domain.BatchItem, len(batch.Items))
			copy(itemsCopy, batch.Items)
			batchCopy.Items = itemsCopy
			return &batchCopy, nil
		}
	}

	return nil, fmt.Errorf("no batch found containing order %s", orderID)
}

// FindPendingBatchForProduct finds a pending batch for a product (for adding new orders)
func (r *BatchMemoryRepository) FindPendingBatchForProduct(productID string) (*domain.Batch, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, batch := range r.batches {
		if batch.ProductID == productID && batch.Status == domain.BatchStatusPending {
			// Create a copy
			batchCopy := *batch
			itemsCopy := make([]domain.BatchItem, len(batch.Items))
			copy(itemsCopy, batch.Items)
			batchCopy.Items = itemsCopy
			return &batchCopy, nil
		}
	}

	return nil, fmt.Errorf("no pending batch found for product %s", productID)
}

// Delete removes a batch from the repository
func (r *BatchMemoryRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.batches[id]; !exists {
		return fmt.Errorf("batch with ID %s not found", id)
	}

	delete(r.batches, id)
	return nil
}

// GetAll retrieves all batches
func (r *BatchMemoryRepository) GetAll() ([]*domain.Batch, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*domain.Batch
	for _, batch := range r.batches {
		// Create a copy
		batchCopy := *batch
		itemsCopy := make([]domain.BatchItem, len(batch.Items))
		copy(itemsCopy, batch.Items)
		batchCopy.Items = itemsCopy
		result = append(result, &batchCopy)
	}

	return result, nil
}

// GetBatchCount returns the total number of batches (useful for testing)
func (r *BatchMemoryRepository) GetBatchCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.batches)
}