package domain

import (
	"time"
)

// Order represents a domain order entity
type Order struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	ProductID   string    `json:"product_id"`
	Quantity    int       `json:"quantity"`
	Status      string    `json:"status"`
	TotalAmount float64   `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// OrderEvent represents a domain event for orders
type OrderEvent struct {
	EventType string    `json:"event_type"`
	OrderID   string    `json:"order_id"`
	Order     Order     `json:"order"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderEventHandler defines the contract for handling order events
type OrderEventHandler interface {
	HandleOrderEvent(event OrderEvent) error
}

// OrderRepository defines the contract for order persistence
type OrderRepository interface {
	Save(order Order) error
	FindByID(id string) (*Order, error)
	FindAll() ([]Order, error)
	Update(order Order) error
	Delete(id string) error
}

// OrderEventPublisher defines the contract for publishing order events
type OrderEventPublisher interface {
	PublishOrderEvent(event OrderEvent) error
}