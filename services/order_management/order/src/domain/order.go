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

// OrderDamageEvent represents an order damage event from MQTT
type OrderDamageEvent struct {
	EventID     string                 `json:"eventId"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	OccurredAt  time.Time              `json:"occurredAt"`
	OrderID     string                 `json:"orderId"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Details     OrderDamageDetails     `json:"details"`
}

// OrderDamageDetails contains the sensor data that triggered the damage event
type OrderDamageDetails struct {
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Status      string  `json:"status"`
	MqttTopic   string  `json:"mqttTopic"`
}

// MQTTOrderEvent represents the wrapper structure for MQTT events
type MQTTOrderEvent struct {
	MqttTopic string  `json:"mqtt_topic"`
	Payload   string  `json:"payload"`
	Timestamp float64 `json:"timestamp"`
}

// OrderEventHandler defines the contract for handling order events
type OrderEventHandler interface {
	HandleOrderEvent(event OrderEvent) error
	HandleOrderDamageEvent(event OrderDamageEvent) error
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