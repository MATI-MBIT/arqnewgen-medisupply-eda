package drivenadapters

import (
	"fmt"
	"sync"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/oder/src/domain"
)

// MemoryOrderRepository is an in-memory implementation of the OrderRepository
// This is suitable for development and testing purposes
type MemoryOrderRepository struct {
	orders map[string]domain.Order
	mutex  sync.RWMutex
}

// NewMemoryOrderRepository creates a new MemoryOrderRepository
func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		orders: make(map[string]domain.Order),
	}
}

// Save stores an order in memory
func (r *MemoryOrderRepository) Save(order domain.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.orders[order.ID] = order
	return nil
}

// FindByID retrieves an order by its ID
func (r *MemoryOrderRepository) FindByID(id string) (*domain.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	order, exists := r.orders[id]
	if !exists {
		return nil, fmt.Errorf("order with ID %s not found", id)
	}
	
	return &order, nil
}

// FindAll retrieves all orders
func (r *MemoryOrderRepository) FindAll() ([]domain.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	orders := make([]domain.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}
	
	return orders, nil
}

// Update updates an existing order
func (r *MemoryOrderRepository) Update(order domain.Order) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.orders[order.ID]; !exists {
		return fmt.Errorf("order with ID %s not found", order.ID)
	}
	
	r.orders[order.ID] = order
	return nil
}

// Delete removes an order by its ID
func (r *MemoryOrderRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.orders[id]; !exists {
		return fmt.Errorf("order with ID %s not found", id)
	}
	
	delete(r.orders, id)
	return nil
}