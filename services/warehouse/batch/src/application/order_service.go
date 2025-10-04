package application

import (
	"fmt"
	"log"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
)

// OrderService handles business logic for order events
type OrderService struct {
	batchService *BatchService
}

// NewOrderService creates a new OrderService
func NewOrderService(batchService *BatchService) *OrderService {
	return &OrderService{
		batchService: batchService,
	}
}

// HandleOrderEvent processes the received order event
func (s *OrderService) HandleOrderEvent(event domain.OrderEvent) error {
	log.Printf("Received order event: Type=%s, OrderID=%s, Status=%s", 
		event.EventType, event.OrderID, event.Order.Status)

	// Check if this event is relevant for warehouse processing
	if !event.IsWarehouseRelevant() {
		log.Printf("Event type %s is not relevant for warehouse processing, skipping", event.EventType)
		return nil
	}

	// Get the warehouse action for this event
	action := event.GetWarehouseAction()
	log.Printf("Processing warehouse action: %s for order %s", action, event.OrderID)

	// Process based on the warehouse action
	switch action {
	case "process_damage":
		return s.processDamage(event)
	case "allocate_inventory":
		return s.allocateInventory(event)
	case "release_inventory":
		return s.releaseInventory(event)
	case "update_inventory":
		return s.updateInventory(event)
	case "confirm_delivery":
		return s.confirmDelivery(event)
	case "process_return":
		return s.processReturn(event)
	case "confirm_allocation":
		return s.confirmAllocation(event)
	case "confirm_release":
		return s.confirmRelease(event)
	default:
		log.Printf("Unknown warehouse action: %s", action)
		return fmt.Errorf("unknown warehouse action: %s", action)
	}
}

// processDamage handles damage processing events
func (s *OrderService) processDamage(event domain.OrderEvent) error {
	log.Printf("Processing damage for order %s: Status=%s, Quantity=%d", 
		event.OrderID, event.Order.Status, event.Order.Quantity)
	
	// Business logic for damage processing
	switch event.Order.Status {
	case "damage_detected_minor":
		log.Printf("Minor damage detected for order %s - marking for inspection", event.OrderID)
		// Try to update order status in batch, if not found create new batch
		if err := s.batchService.UpdateOrderStatus(event.OrderID, "damage_minor"); err != nil {
			log.Printf("Order not found in existing batch, creating new batch for damage processing: %v", err)
			// Create new batch with the order for damage processing
			_, err := s.batchService.AddOrderToBatch(
				event.OrderID,
				event.Order.ProductID,
				event.Order.Quantity,
				"damage_minor",
			)
			if err != nil {
				log.Printf("Failed to create batch for damage processing: %v", err)
				return err
			}
			log.Printf("Created new batch for order %s with minor damage status", event.OrderID)
		}
	case "damage_detected_major":
		log.Printf("Major damage detected for order %s - marking as damaged", event.OrderID)
		// Try to update order status in batch, if not found create new batch
		if err := s.batchService.UpdateOrderStatus(event.OrderID, "damage_major"); err != nil {
			log.Printf("Order not found in existing batch, creating new batch for damage processing: %v", err)
			// Create new batch with the order for damage processing
			batch, err := s.batchService.AddOrderToBatch(
				event.OrderID,
				event.Order.ProductID,
				event.Order.Quantity,
				"damage_major",
			)
			if err != nil {
				log.Printf("Failed to create batch for damage processing: %v", err)
				return err
			}
			log.Printf("Created new batch %s for order %s with major damage status", batch.ID, event.OrderID)
			// Mark the entire batch as damaged since it's major damage
			if err := s.batchService.MarkBatchAsDamaged(batch.ID); err != nil {
				log.Printf("Failed to mark batch as damaged: %v", err)
			}
		} else {
			// Order was found and updated, now mark the batch as damaged
			batch, err := s.batchService.GetBatchByOrderID(event.OrderID)
			if err == nil {
				if err := s.batchService.MarkBatchAsDamaged(batch.ID); err != nil {
					log.Printf("Failed to mark batch as damaged: %v", err)
				}
			}
		}
	case "damage_processed":
		log.Printf("Damage processing completed for order %s", event.OrderID)
		// Try to update order status to processed, if not found create new batch
		if err := s.batchService.UpdateOrderStatus(event.OrderID, "damage_processed"); err != nil {
			log.Printf("Order not found in existing batch, creating new batch for damage processing: %v", err)
			// Create new batch with the order for damage processing completion
			_, err := s.batchService.AddOrderToBatch(
				event.OrderID,
				event.Order.ProductID,
				event.Order.Quantity,
				"damage_processed",
			)
			if err != nil {
				log.Printf("Failed to create batch for damage processing: %v", err)
				return err
			}
			log.Printf("Created new batch for order %s with damage processed status", event.OrderID)
		}
	default:
		log.Printf("Unknown damage status: %s for order %s", event.Order.Status, event.OrderID)
	}
	
	return nil
}

// allocateInventory handles inventory allocation for new orders
func (s *OrderService) allocateInventory(event domain.OrderEvent) error {
	log.Printf("Allocating inventory for order %s: ProductID=%s, Quantity=%d", 
		event.OrderID, event.Order.ProductID, event.Order.Quantity)
	
	// Add order to batch for processing
	batch, err := s.batchService.AddOrderToBatch(
		event.OrderID, 
		event.Order.ProductID, 
		event.Order.Quantity, 
		"allocated",
	)
	if err != nil {
		log.Printf("Failed to add order to batch: %v", err)
		return err
	}
	
	log.Printf("Order %s added to batch %s for inventory allocation", event.OrderID, batch.ID)
	return nil
}

// releaseInventory handles inventory release for cancelled orders
func (s *OrderService) releaseInventory(event domain.OrderEvent) error {
	log.Printf("Releasing inventory for cancelled order %s: ProductID=%s, Quantity=%d", 
		event.OrderID, event.Order.ProductID, event.Order.Quantity)
	
	// Remove order from batch since it's cancelled
	if err := s.batchService.RemoveOrderFromBatch(event.OrderID); err != nil {
		log.Printf("Failed to remove order from batch: %v", err)
		return err
	}
	
	log.Printf("Order %s removed from batch due to cancellation", event.OrderID)
	return nil
}

// updateInventory handles inventory updates for shipped orders
func (s *OrderService) updateInventory(event domain.OrderEvent) error {
	log.Printf("Updating inventory for shipped order %s: ProductID=%s, Quantity=%d", 
		event.OrderID, event.Order.ProductID, event.Order.Quantity)
	
	// Update order status to shipped in batch
	if err := s.batchService.UpdateOrderStatus(event.OrderID, "shipped"); err != nil {
		log.Printf("Failed to update order status in batch: %v", err)
		return err
	}
	
	log.Printf("Order %s status updated to shipped in batch", event.OrderID)
	return nil
}

// confirmDelivery handles delivery confirmation
func (s *OrderService) confirmDelivery(event domain.OrderEvent) error {
	log.Printf("Confirming delivery for order %s", event.OrderID)
	
	// Update order status to delivered in batch
	if err := s.batchService.UpdateOrderStatus(event.OrderID, "delivered"); err != nil {
		log.Printf("Failed to update order status in batch: %v", err)
		return err
	}
	
	log.Printf("Order %s status updated to delivered in batch", event.OrderID)
	return nil
}

// processReturn handles returned orders
func (s *OrderService) processReturn(event domain.OrderEvent) error {
	log.Printf("Processing return for order %s: ProductID=%s, Quantity=%d", 
		event.OrderID, event.Order.ProductID, event.Order.Quantity)
	
	// Update order status to returned in batch
	if err := s.batchService.UpdateOrderStatus(event.OrderID, "returned"); err != nil {
		log.Printf("Failed to update order status in batch: %v", err)
		return err
	}
	
	// Add returned item back to inventory by creating a new batch entry
	_, err := s.batchService.AddOrderToBatch(
		event.OrderID+"-return", 
		event.Order.ProductID, 
		event.Order.Quantity, 
		"returned",
	)
	if err != nil {
		log.Printf("Failed to add returned item to batch: %v", err)
		return err
	}
	
	log.Printf("Order %s processed as return and added back to inventory", event.OrderID)
	return nil
}

// confirmAllocation confirms inventory allocation
func (s *OrderService) confirmAllocation(event domain.OrderEvent) error {
	log.Printf("Confirming inventory allocation for order %s", event.OrderID)
	
	// Update order status to allocation confirmed in batch
	if err := s.batchService.UpdateOrderStatus(event.OrderID, "allocation_confirmed"); err != nil {
		log.Printf("Failed to update order status in batch: %v", err)
		return err
	}
	
	log.Printf("Order %s allocation confirmed in batch", event.OrderID)
	return nil
}

// confirmRelease confirms inventory release
func (s *OrderService) confirmRelease(event domain.OrderEvent) error {
	log.Printf("Confirming inventory release for order %s", event.OrderID)
	
	// Update order status to release confirmed in batch
	if err := s.batchService.UpdateOrderStatus(event.OrderID, "release_confirmed"); err != nil {
		log.Printf("Failed to update order status in batch: %v", err)
		return err
	}
	
	log.Printf("Order %s release confirmed in batch", event.OrderID)
	return nil
}