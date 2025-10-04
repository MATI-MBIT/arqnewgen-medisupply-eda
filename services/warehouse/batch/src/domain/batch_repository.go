package domain

// BatchRepository defines the contract for batch persistence
type BatchRepository interface {
	// Save stores or updates a batch
	Save(batch *Batch) error
	
	// FindByID retrieves a batch by its ID
	FindByID(id string) (*Batch, error)
	
	// FindByProductID retrieves all batches for a specific product
	FindByProductID(productID string) ([]*Batch, error)
	
	// FindByStatus retrieves all batches with a specific status
	FindByStatus(status BatchStatus) ([]*Batch, error)
	
	// FindByOrderID finds the batch containing a specific order
	FindByOrderID(orderID string) (*Batch, error)
	
	// FindPendingBatchForProduct finds a pending batch for a product (for adding new orders)
	FindPendingBatchForProduct(productID string) (*Batch, error)
	
	// Delete removes a batch from the repository
	Delete(id string) error
	
	// GetAll retrieves all batches
	GetAll() ([]*Batch, error)
}