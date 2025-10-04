package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/oder/src/domain"
)

// Example demonstrating how the order damage event is processed
func main() {
	// Example MQTT message received (as shown in the user's request)
	mqttMessage := `{
		"mqtt_topic": "events/order-damage",
		"payload": "{\"eventId\":\"evt_1759555253\",\"type\":\"order.damage\",\"source\":\"mqtt-order-event-client\",\"occurredAt\":\"2025-10-04T05:20:53.182048461Z\",\"orderId\":\"evt_1759555253\",\"severity\":\"minor\",\"description\":\"Potential damage detected: temp=9.23C, humidity=58.00%\",\"details\":{\"temperature\":9.234309268737501,\"humidity\":58,\"status\":\"active\",\"mqttTopic\":\"events/sensor\"}}",
		"timestamp": 1759555253.2585864
	}`

	// Parse the MQTT message
	var mqttEvent domain.MQTTOrderEvent
	if err := json.Unmarshal([]byte(mqttMessage), &mqttEvent); err != nil {
		log.Fatalf("Failed to parse MQTT message: %v", err)
	}

	fmt.Printf("Received MQTT event on topic: %s\n", mqttEvent.MqttTopic)
	fmt.Printf("Timestamp: %f\n", mqttEvent.Timestamp)

	// Parse the nested payload
	var damageEvent domain.OrderDamageEvent
	if err := json.Unmarshal([]byte(mqttEvent.Payload), &damageEvent); err != nil {
		log.Fatalf("Failed to parse damage event payload: %v", err)
	}

	// Display the parsed damage event
	fmt.Printf("\n=== Order Damage Event Details ===\n")
	fmt.Printf("Event ID: %s\n", damageEvent.EventID)
	fmt.Printf("Type: %s\n", damageEvent.Type)
	fmt.Printf("Source: %s\n", damageEvent.Source)
	fmt.Printf("Occurred At: %s\n", damageEvent.OccurredAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Order ID: %s\n", damageEvent.OrderID)
	fmt.Printf("Severity: %s\n", damageEvent.Severity)
	fmt.Printf("Description: %s\n", damageEvent.Description)

	fmt.Printf("\n=== Sensor Details ===\n")
	fmt.Printf("Temperature: %.2f°C\n", damageEvent.Details.Temperature)
	fmt.Printf("Humidity: %d%%\n", damageEvent.Details.Humidity)
	fmt.Printf("Status: %s\n", damageEvent.Details.Status)
	fmt.Printf("MQTT Topic: %s\n", damageEvent.Details.MqttTopic)

	// Simulate processing the event
	fmt.Printf("\n=== Processing Event ===\n")
	switch damageEvent.Severity {
	case "minor":
		fmt.Printf("✓ Minor damage detected - Order %s flagged for monitoring\n", damageEvent.OrderID)
	case "major":
		fmt.Printf("⚠ Major damage detected - Order %s requires immediate attention\n", damageEvent.OrderID)
	case "critical":
		fmt.Printf("🚨 Critical damage detected - Order %s should be cancelled\n", damageEvent.OrderID)
	}

	fmt.Printf("\nEvent processing completed successfully!\n")
}