package domain

import (
	"fmt"
	"time"
)

// BatchStatus represents the status of a batch
type BatchStatus string

const (
	BatchStatusPending    BatchStatus = "pending"
	BatchStatusProcessing BatchStatus = "processing"
	BatchStatusCompleted  BatchStatus = "completed"
	BatchStatusCancelled  BatchStatus = "cancelled"
	BatchStatusDamaged    BatchStatus = "damaged"
)

// BatchItem represents an item within a batch
type BatchItem struct {
	OrderID     string    `json:"order_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	Status      string    `json:"status"`
	AddedAt     time.Time `json:"added_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

// Batch represents a batch aggregate in the warehouse domain
type Batch struct {
	ID          string      `json:"id"`
	ProductID   string      `json:"product_id"`
	Status      BatchStatus `json:"status"`
	Items       []BatchItem `json:"items"`
	TotalItems  int         `json:"total_items"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	ProcessedAt *time.Time  `json:"processed_at,omitempty"`
}

// NewBatch creates a new batch with the given product ID
func NewBatch(id, productID string) *Batch {
	now := time.Now()
	return &Batch{
		ID:         id,
		ProductID:  productID,
		Status:     BatchStatusPending,
		Items:      make([]BatchItem, 0),
		TotalItems: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// AddItem adds an order item to the batch
func (b *Batch) AddItem(orderID, productID string, quantity int, status string) error {
	if b.ProductID != productID {
		return fmt.Errorf("product ID mismatch: batch is for %s, item is for %s", b.ProductID, productID)
	}

	if b.Status == BatchStatusCompleted || b.Status == BatchStatusCancelled {
		return fmt.Errorf("cannot add items to batch with status %s", b.Status)
	}

	// Check if order already exists in batch
	for i, item := range b.Items {
		if item.OrderID == orderID {
			// Update existing item
			b.Items[i].Quantity = quantity
			b.Items[i].Status = status
			b.Items[i].AddedAt = time.Now()
			b.UpdatedAt = time.Now()
			return nil
		}
	}

	// Add new item
	item := BatchItem{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Status:    status,
		AddedAt:   time.Now(),
	}

	b.Items = append(b.Items, item)
	b.TotalItems = len(b.Items)
	b.UpdatedAt = time.Now()

	return nil
}

// RemoveItem removes an order item from the batch
func (b *Batch) RemoveItem(orderID string) error {
	if b.Status == BatchStatusCompleted {
		return fmt.Errorf("cannot remove items from completed batch")
	}

	for i, item := range b.Items {
		if item.OrderID == orderID {
			// Remove item from slice
			b.Items = append(b.Items[:i], b.Items[i+1:]...)
			b.TotalItems = len(b.Items)
			b.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("order %s not found in batch", orderID)
}

// UpdateItemStatus updates the status of a specific item in the batch
func (b *Batch) UpdateItemStatus(orderID, status string) error {
	for i, item := range b.Items {
		if item.OrderID == orderID {
			b.Items[i].Status = status
			if status == "processed" || status == "shipped" || status == "delivered" {
				now := time.Now()
				b.Items[i].ProcessedAt = &now
			}
			b.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("order %s not found in batch", orderID)
}

// StartProcessing changes the batch status to processing
func (b *Batch) StartProcessing() error {
	if b.Status != BatchStatusPending {
		return fmt.Errorf("cannot start processing batch with status %s", b.Status)
	}

	b.Status = BatchStatusProcessing
	b.UpdatedAt = time.Now()
	return nil
}

// Complete marks the batch as completed
func (b *Batch) Complete() error {
	if b.Status != BatchStatusProcessing {
		return fmt.Errorf("cannot complete batch with status %s", b.Status)
	}

	b.Status = BatchStatusCompleted
	now := time.Now()
	b.ProcessedAt = &now
	b.UpdatedAt = now
	return nil
}

// Cancel marks the batch as cancelled
func (b *Batch) Cancel() error {
	if b.Status == BatchStatusCompleted {
		return fmt.Errorf("cannot cancel completed batch")
	}

	b.Status = BatchStatusCancelled
	b.UpdatedAt = time.Now()
	return nil
}

// MarkAsDamaged marks the batch as damaged
func (b *Batch) MarkAsDamaged() error {
	b.Status = BatchStatusDamaged
	b.UpdatedAt = time.Now()
	return nil
}

// GetItemByOrderID returns the batch item for a specific order ID
func (b *Batch) GetItemByOrderID(orderID string) (*BatchItem, error) {
	for _, item := range b.Items {
		if item.OrderID == orderID {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("order %s not found in batch", orderID)
}

// HasOrder checks if the batch contains a specific order
func (b *Batch) HasOrder(orderID string) bool {
	_, err := b.GetItemByOrderID(orderID)
	return err == nil
}

// IsEmpty returns true if the batch has no items
func (b *Batch) IsEmpty() bool {
	return len(b.Items) == 0
}

// GetTotalQuantity returns the total quantity of all items in the batch
func (b *Batch) GetTotalQuantity() int {
	total := 0
	for _, item := range b.Items {
		total += item.Quantity
	}
	return total
}