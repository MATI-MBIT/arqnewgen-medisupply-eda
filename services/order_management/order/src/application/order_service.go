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