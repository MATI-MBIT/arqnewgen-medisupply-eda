package application

import (
	"fmt"
	"log"
	"time"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/oder/src/domain"
	"github.com/google/uuid"
)

// OrderService handles business logic for orders
type OrderService struct {
	orderRepo      domain.OrderRepository
	eventPublisher domain.OrderEventPublisher
}

// NewOrderService creates a new OrderService
func NewOrderService(orderRepo domain.OrderRepository, eventPublisher domain.OrderEventPublisher) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		eventPublisher: eventPublisher,
	}
}

// CreateOrder creates a new order and publishes an event
func (s *OrderService) CreateOrder(customerID, productID string, quantity int, totalAmount float64) (*domain.Order, error) {
	// Create new order
	order := domain.Order{
		ID:          uuid.New().String(),
		CustomerID:  customerID,
		ProductID:   productID,
		Quantity:    quantity,
		Status:      "created",
		TotalAmount: totalAmount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save order
	if err := s.orderRepo.Save(order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// Publish order created event
	event := domain.OrderEvent{
		EventType: "order.created",
		OrderID:   order.ID,
		Order:     order,
		Timestamp: time.Now(),
	}

	if err := s.eventPublisher.PublishOrderEvent(event); err != nil {
		log.Printf("Failed to publish order created event: %v", err)
		// Note: In a real system, you might want to implement compensation logic
	}

	log.Printf("Order created successfully: ID=%s, CustomerID=%s, ProductID=%s", 
		order.ID, order.CustomerID, order.ProductID)

	return &order, nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(id string) (*domain.Order, error) {
	return s.orderRepo.FindByID(id)
}

// GetAllOrders retrieves all orders
func (s *OrderService) GetAllOrders() ([]domain.Order, error) {
	return s.orderRepo.FindAll()
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(id, status string) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	order.Status = status
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(*order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Publish order updated event
	event := domain.OrderEvent{
		EventType: "order.updated",
		OrderID:   order.ID,
		Order:     *order,
		Timestamp: time.Now(),
	}

	if err := s.eventPublisher.PublishOrderEvent(event); err != nil {
		log.Printf("Failed to publish order updated event: %v", err)
	}

	log.Printf("Order status updated: ID=%s, Status=%s", order.ID, order.Status)

	return order, nil
}

// HandleOrderEvent processes incoming order events
func (s *OrderService) HandleOrderEvent(event domain.OrderEvent) error {
	log.Printf("Processing order event: Type=%s, OrderID=%s, Timestamp=%s", 
		event.EventType, event.OrderID, event.Timestamp.Format("2006-01-02 15:04:05"))
	
	// Business logic for processing different event types
	switch event.EventType {
	case "order.created":
		log.Printf("Order created event processed: %s", event.OrderID)
	case "order.updated":
		log.Printf("Order updated event processed: %s", event.OrderID)
	case "order.cancelled":
		log.Printf("Order cancelled event processed: %s", event.OrderID)
	default:
		log.Printf("Unknown event type: %s", event.EventType)
	}
	
	return nil
}

// HandleOrderDamageEvent processes incoming order damage events from MQTT
func (s *OrderService) HandleOrderDamageEvent(event domain.OrderDamageEvent) error {
	log.Printf("Processing order damage event: EventID=%s, OrderID=%s, Severity=%s, OccurredAt=%s", 
		event.EventID, event.OrderID, event.Severity, event.OccurredAt.Format("2006-01-02 15:04:05"))
	
	log.Printf("Damage details: Temperature=%.2f°C, Humidity=%d%%, Status=%s", 
		event.Details.Temperature, event.Details.Humidity, event.Details.Status)
	
	log.Printf("Damage description: %s", event.Description)
	
	// Check if order exists, if not create a new one
	order, err := s.orderRepo.FindByID(event.OrderID)
	if err != nil {
		log.Printf("Order %s not found, creating new order from damage event", event.OrderID)
		
		// Create new order with the received order ID
		newOrder := domain.Order{
			ID:          event.OrderID,
			CustomerID:  "unknown", // Default value since not provided in damage event
			ProductID:   "unknown", // Default value since not provided in damage event
			Quantity:    1,         // Default value
			Status:      "created_from_damage_event",
			TotalAmount: 0.0,       // Default value
			CreatedAt:   event.OccurredAt,
			UpdatedAt:   time.Now(),
		}
		
		// Save the new order
		if err := s.orderRepo.Save(newOrder); err != nil {
			return fmt.Errorf("failed to create order from damage event: %w", err)
		}
		
		log.Printf("Created new order from damage event: ID=%s", newOrder.ID)
		order = &newOrder
	} else {
		log.Printf("Found existing order %s", event.OrderID)
	}
	
	// Determine the new status based on damage severity
	var newStatus string
	switch event.Severity {
	case "minor":
		log.Printf("Minor damage detected for order %s - monitoring required", event.OrderID)
		newStatus = "damage_detected_minor"
		
	case "major":
		log.Printf("Major damage detected for order %s - immediate action required", event.OrderID)
		newStatus = "damage_detected_major"
		
	case "critical":
		log.Printf("Critical damage detected for order %s - order should be cancelled", event.OrderID)
		newStatus = "cancelled_damage"
		
	default:
		log.Printf("Unknown damage severity: %s for order %s", event.Severity, event.OrderID)
		newStatus = "damage_detected_unknown"
	}
	
	// Update order status
	order.Status = newStatus
	order.UpdatedAt = time.Now()
	
	if err := s.orderRepo.Update(*order); err != nil {
		return fmt.Errorf("failed to update order status after damage event: %w", err)
	}
	
	// Publish order updated event
	orderEvent := domain.OrderEvent{
		EventType: "order.damage_processed",
		OrderID:   order.ID,
		Order:     *order,
		Timestamp: time.Now(),
	}
	
	if err := s.eventPublisher.PublishOrderEvent(orderEvent); err != nil {
		log.Printf("Failed to publish order damage processed event: %v", err)
	}
	
	log.Printf("Order %s status updated to: %s", order.ID, order.Status)
	
	// Additional business logic could include:
	// - Sending notifications to warehouse staff
	// - Creating damage reports
	// - Triggering insurance claims
	// - Updating inventory status
	
	return nil
}