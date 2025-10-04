package domain

import "time"

// Order represents an order in the system
type Order struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	ProductID    string    `json:"product_id"`
	Quantity     int       `json:"quantity"`
	Status       string    `json:"status"`
	TotalAmount  float64   `json:"total_amount"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OrderEvent represents an order event from the order-events topic
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

// IsWarehouseRelevant checks if the order event is relevant for warehouse processing
func (oe *OrderEvent) IsWarehouseRelevant() bool {
	warehouseRelevantEvents := map[string]bool{
		"order.damage_processed":     true,
		"order.created":             true,
		"order.cancelled":           true,
		"order.shipped":             true,
		"order.delivered":           true,
		"order.returned":            true,
		"order.inventory_allocated": true,
		"order.inventory_released":  true,
	}
	
	return warehouseRelevantEvents[oe.EventType]
}

// GetWarehouseAction returns the warehouse action based on the event type
func (oe *OrderEvent) GetWarehouseAction() string {
	switch oe.EventType {
	case "order.damage_processed":
		return "process_damage"
	case "order.created":
		return "allocate_inventory"
	case "order.cancelled":
		return "release_inventory"
	case "order.shipped":
		return "update_inventory"
	case "order.delivered":
		return "confirm_delivery"
	case "order.returned":
		return "process_return"
	case "order.inventory_allocated":
		return "confirm_allocation"
	case "order.inventory_released":
		return "confirm_release"
	default:
		return "unknown"
	}
}