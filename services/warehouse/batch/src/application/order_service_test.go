package application

import (
	"encoding/json"
	"testing"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/batch/src/domain"
)

func TestOrderService_HandleOrderEvent(t *testing.T) {
	service := NewOrderService()

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