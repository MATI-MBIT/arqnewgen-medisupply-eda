package application

import (
	"encoding/json"
	"testing"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
	drivenadapters "github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/infrastructure/driven-adapters"
)

func TestOrderService_HandleOrderEvent(t *testing.T) {
	// Setup dependencies
	repo := drivenadapters.NewBatchMemoryRepository()
	batchService := NewBatchService(repo)
	service := NewOrderService(batchService)

	// Test event JSON from the user's example
	eventJSON := `{
		"event_type": "order.damage_processed",
		"order_id": "evt_1759598824",
		"order": {
			"id": "evt_1759598824",
			"customer_id": "unknown",
			"product_id": "unknown",
			"quantity": 1,
			"status": "damage_detected_minor",
			"total_amount": 0,
			"created_at": "2025-10-04T17:27:04.082881166Z",
			"updated_at": "2025-10-04T17:36:13.584671556Z"
		},
		"timestamp": "2025-10-04T17:36:13.58470126Z"
	}`

	var orderEvent domain.OrderEvent
	err := json.Unmarshal([]byte(eventJSON), &orderEvent)
	if err != nil {
		t.Fatalf("Failed to unmarshal test event: %v", err)
	}

	// Test that the event is parsed correctly
	if orderEvent.EventType != "order.damage_processed" {
		t.Errorf("Expected event type 'order.damage_processed', got '%s'", orderEvent.EventType)
	}

	if orderEvent.OrderID != "evt_1759598824" {
		t.Errorf("Expected order ID 'evt_1759598824', got '%s'", orderEvent.OrderID)
	}

	if orderEvent.Order.Status != "damage_detected_minor" {
		t.Errorf("Expected order status 'damage_detected_minor', got '%s'", orderEvent.Order.Status)
	}

	// Test that the event is warehouse relevant
	if !orderEvent.IsWarehouseRelevant() {
		t.Error("Expected damage_processed event to be warehouse relevant")
	}

	// Test warehouse action
	action := orderEvent.GetWarehouseAction()
	if action != "process_damage" {
		t.Errorf("Expected warehouse action 'process_damage', got '%s'", action)
	}

	// First add the order to a batch (simulate it was created earlier)
	_, err = batchService.AddOrderToBatch(orderEvent.OrderID, "product-123", 1, "allocated")
	if err != nil {
		t.Fatalf("Failed to add order to batch: %v", err)
	}

	// Test handling the event
	err = service.HandleOrderEvent(orderEvent)
	if err != nil {
		t.Errorf("Failed to handle order event: %v", err)
	}
}

func TestOrderEvent_IsWarehouseRelevant(t *testing.T) {
	tests := []struct {
		eventType string
		expected  bool
	}{
		{"order.damage_processed", true},
		{"order.created", true},
		{"order.cancelled", true},
		{"order.shipped", true},
		{"order.delivered", true},
		{"order.returned", true},
		{"order.inventory_allocated", true},
		{"order.inventory_released", true},
		{"order.payment_processed", false},
		{"order.notification_sent", false},
	}

	for _, test := range tests {
		event := domain.OrderEvent{EventType: test.eventType}
		result := event.IsWarehouseRelevant()
		if result != test.expected {
			t.Errorf("For event type '%s', expected %v, got %v", test.eventType, test.expected, result)
		}
	}
}

func TestOrderEvent_GetWarehouseAction(t *testing.T) {
	tests := []struct {
		eventType string
		expected  string
	}{
		{"order.damage_processed", "process_damage"},
		{"order.created", "allocate_inventory"},
		{"order.cancelled", "release_inventory"},
		{"order.shipped", "update_inventory"},
		{"order.delivered", "confirm_delivery"},
		{"order.returned", "process_return"},
		{"order.inventory_allocated", "confirm_allocation"},
		{"order.inventory_released", "confirm_release"},
		{"order.unknown_event", "unknown"},
	}

	for _, test := range tests {
		event := domain.OrderEvent{EventType: test.eventType}
		result := event.GetWarehouseAction()
		if result != test.expected {
			t.Errorf("For event type '%s', expected action '%s', got '%s'", test.eventType, test.expected, result)
		}
	}
}

func TestOrderService_ProcessDamage_CreatesBatchWhenNotExists(t *testing.T) {
	// Setup dependencies
	repo := drivenadapters.NewBatchMemoryRepository()
	batchService := NewBatchService(repo)
	service := NewOrderService(batchService)

	tests := []struct {
		name           string
		damageStatus   string
		expectedStatus string
		shouldMarkDamaged bool
	}{
		{
			name:           "Minor damage creates batch",
			damageStatus:   "damage_detected_minor",
			expectedStatus: "damage_minor",
			shouldMarkDamaged: false,
		},
		{
			name:           "Major damage creates batch and marks as damaged",
			damageStatus:   "damage_detected_major",
			expectedStatus: "damage_major",
			shouldMarkDamaged: true,
		},
		{
			name:           "Damage processed creates batch",
			damageStatus:   "damage_processed",
			expectedStatus: "damage_processed",
			shouldMarkDamaged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create damage event for non-existing order
			orderEvent := domain.OrderEvent{
				EventType: "order.damage_processed",
				OrderID:   "damage-order-" + tt.name,
				Order: domain.Order{
					ID:        "damage-order-" + tt.name,
					ProductID: "product-damage-123",
					Quantity:  5,
					Status:    tt.damageStatus,
				},
			}

			// Verify order doesn't exist in any batch initially
			_, err := batchService.GetBatchByOrderID(orderEvent.OrderID)
			if err == nil {
				t.Fatal("Expected order to not exist in any batch initially")
			}

			// Process the damage event
			err = service.HandleOrderEvent(orderEvent)
			if err != nil {
				t.Fatalf("Failed to handle damage event: %v", err)
			}

			// Verify batch was created with the order
			batch, err := batchService.GetBatchByOrderID(orderEvent.OrderID)
			if err != nil {
				t.Fatalf("Expected batch to be created for order: %v", err)
			}

			// Verify batch contains the order with correct status
			item, err := batch.GetItemByOrderID(orderEvent.OrderID)
			if err != nil {
				t.Fatalf("Expected order to be in batch: %v", err)
			}

			if item.Status != tt.expectedStatus {
				t.Errorf("Expected order status '%s', got '%s'", tt.expectedStatus, item.Status)
			}

			if item.ProductID != orderEvent.Order.ProductID {
				t.Errorf("Expected product ID '%s', got '%s'", orderEvent.Order.ProductID, item.ProductID)
			}

			if item.Quantity != orderEvent.Order.Quantity {
				t.Errorf("Expected quantity %d, got %d", orderEvent.Order.Quantity, item.Quantity)
			}

			// Verify batch damage status for major damage
			if tt.shouldMarkDamaged {
				if batch.Status != domain.BatchStatusDamaged {
					t.Errorf("Expected batch to be marked as damaged, got status '%s'", batch.Status)
				}
			}
		})
	}
}

func TestOrderService_ProcessDamage_UpdatesExistingBatch(t *testing.T) {
	// Setup dependencies
	repo := drivenadapters.NewBatchMemoryRepository()
	batchService := NewBatchService(repo)
	service := NewOrderService(batchService)

	// Create an order in a batch first
	orderID := "existing-order-123"
	productID := "product-456"
	
	_, err := batchService.AddOrderToBatch(orderID, productID, 3, "allocated")
	if err != nil {
		t.Fatalf("Failed to create initial batch: %v", err)
	}

	// Create damage event for existing order
	orderEvent := domain.OrderEvent{
		EventType: "order.damage_processed",
		OrderID:   orderID,
		Order: domain.Order{
			ID:        orderID,
			ProductID: productID,
			Quantity:  3,
			Status:    "damage_detected_major",
		},
	}

	// Process the damage event
	err = service.HandleOrderEvent(orderEvent)
	if err != nil {
		t.Fatalf("Failed to handle damage event: %v", err)
	}

	// Verify order status was updated
	batch, err := batchService.GetBatchByOrderID(orderID)
	if err != nil {
		t.Fatalf("Failed to get batch: %v", err)
	}

	item, err := batch.GetItemByOrderID(orderID)
	if err != nil {
		t.Fatalf("Failed to get order item: %v", err)
	}

	if item.Status != "damage_major" {
		t.Errorf("Expected order status 'damage_major', got '%s'", item.Status)
	}

	// Verify batch was marked as damaged for major damage
	if batch.Status != domain.BatchStatusDamaged {
		t.Errorf("Expected batch to be marked as damaged, got status '%s'", batch.Status)
	}
}