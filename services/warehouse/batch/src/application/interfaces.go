package application

import (
	"time"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
)

// BatchServiceInterface defines the contract for batch operations
type BatchServiceInterface interface {
	AddOrderToBatch(orderID, productID string, quantity int, status string) (*domain.Batch, error)
	RemoveOrderFromBatch(orderID string) error
	UpdateOrderStatus(orderID, status string) error
	ProcessBatch(batchID string) error
	CompleteBatch(batchID string) error
	CancelBatch(batchID string) error
	MarkBatchAsDamaged(batchID string) error
	GetBatchByOrderID(orderID string) (*domain.Batch, error)
	GetBatchesByProductID(productID string) ([]*domain.Batch, error)
	GetBatchesByStatus(status domain.BatchStatus) ([]*domain.Batch, error)
	GetAllBatches() ([]*domain.Batch, error)
}

// BatchDTO represents a batch for API responses
type BatchDTO struct {
	ID          string        `json:"id"`
	ProductID   string        `json:"product_id"`
	Status      string        `json:"status"`
	Items       []BatchItemDTO `json:"items"`
	TotalItems  int           `json:"total_items"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	ProcessedAt *time.Time    `json:"processed_at,omitempty"`
}

// BatchItemDTO represents an item within a batch for API responses
type BatchItemDTO struct {
	OrderID     string     `json:"order_id"`
	ProductID   string     `json:"product_id"`
	Quantity    int        `json:"quantity"`
	Status      string     `json:"status"`
	AddedAt     time.Time  `json:"added_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// ToBatchDTO converts a domain batch to a DTO
func ToBatchDTO(batch *domain.Batch) *BatchDTO {
	items := make([]BatchItemDTO, len(batch.Items))
	for i, item := range batch.Items {
		items[i] = BatchItemDTO{
			OrderID:     item.OrderID,
			ProductID:   item.ProductID,
			Quantity:    item.Quantity,
			Status:      item.Status,
			AddedAt:     item.AddedAt,
			ProcessedAt: item.ProcessedAt,
		}
	}

	return &BatchDTO{
		ID:          batch.ID,
		ProductID:   batch.ProductID,
		Status:      string(batch.Status),
		Items:       items,
		TotalItems:  batch.TotalItems,
		CreatedAt:   batch.CreatedAt,
		UpdatedAt:   batch.UpdatedAt,
		ProcessedAt: batch.ProcessedAt,
	}
}

// ToBatchDTOs converts a slice of domain batches to DTOs
func ToBatchDTOs(batches []*domain.Batch) []*BatchDTO {
	dtos := make([]*BatchDTO, len(batches))
	for i, batch := range batches {
		dtos[i] = ToBatchDTO(batch)
	}
	return dtos
}