package drivingadapters

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/oder/src/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

// OrderConsumerAdapter is responsible for consuming order events from RabbitMQ
// and translating them into domain events for the application layer
type OrderConsumerAdapter struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	queueName    string
	exchangeName string
	routingKey   string
	eventHandler domain.OrderEventHandler
}

// NewOrderConsumerAdapter creates a new OrderConsumerAdapter
func NewOrderConsumerAdapter(rabbitMQURL, exchangeName, queueName, routingKey string, eventHandler domain.OrderEventHandler) (*OrderConsumerAdapter, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declare the exchange
	err = channel.ExchangeDeclare(
		exchangeName, // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Declare the queue
	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Bind the queue to the exchange
	err = channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &OrderConsumerAdapter{
		conn:         conn,
		channel:      channel,
		queueName:    queueName,
		exchangeName: exchangeName,
		routingKey:   routingKey,
		eventHandler: eventHandler,
	}, nil
}

// Start begins consuming events from RabbitMQ
func (adapter *OrderConsumerAdapter) Start(ctx context.Context) {
	log.Println("Starting order consumer adapter...")

	// Start consuming messages
	msgs, err := adapter.channel.Consume(
		adapter.queueName, // queue
		"",                // consumer
		false,             // auto-ack is false, we will manually acknowledge
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		log.Printf("Failed to register a consumer: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Order consumer adapter stopping...")
			adapter.Close()
			return
		case delivery, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return
			}

			// Translate RabbitMQ message to domain event
			event, err := adapter.translateMessage(delivery.Body)
			if err != nil {
				log.Printf("Error translating message: %v", err)
				delivery.Nack(false, false) // Reject and don't requeue
				continue
			}

			// Handle the event through the application layer based on event type
			var handlingErr error
			switch e := event.(type) {
			case domain.OrderDamageEvent:
				handlingErr = adapter.eventHandler.HandleOrderDamageEvent(e)
			case domain.OrderEvent:
				handlingErr = adapter.eventHandler.HandleOrderEvent(e)
			default:
				log.Printf("Unknown event type received: %T", e)
				delivery.Nack(false, false) // Reject unknown event types
				continue
			}

			if handlingErr != nil {
				log.Printf("Error handling event: %v", handlingErr)
				delivery.Nack(false, true) // Reject and requeue for retry
			} else {
				delivery.Ack(false) // Acknowledge successful processing
			}
		}
	}
}

// translateMessage converts a RabbitMQ message to a domain order event
func (adapter *OrderConsumerAdapter) translateMessage(body []byte) (interface{}, error) {
	// First try to unmarshal as MQTT order event (for order damage events)
	var mqttEvent domain.MQTTOrderEvent
	if err := json.Unmarshal(body, &mqttEvent); err == nil {
		// Check if this is an order damage event
		if mqttEvent.MqttTopic == "events/order-damage" {
			return adapter.handleOrderDamageEvent(mqttEvent)
		}
	}

	// Try to unmarshal as regular order event
	var event domain.OrderEvent
	if err := json.Unmarshal(body, &event); err == nil {
		return event, nil
	}

	// If JSON unmarshaling fails, create a simple event from the message body
	event = domain.OrderEvent{
		EventType: "order.message",
		OrderID:   "unknown",
		Timestamp: time.Now(),
	}

	log.Printf("Received non-JSON message, created simple event: %s", string(body))
	return event, nil
}

// handleOrderDamageEvent processes order damage events from MQTT
func (adapter *OrderConsumerAdapter) handleOrderDamageEvent(mqttEvent domain.MQTTOrderEvent) (domain.OrderDamageEvent, error) {
	var damageEvent domain.OrderDamageEvent
	
	// Parse the nested JSON payload
	if err := json.Unmarshal([]byte(mqttEvent.Payload), &damageEvent); err != nil {
		return damageEvent, err
	}

	log.Printf("Received order damage event: OrderID=%s, Severity=%s, Description=%s", 
		damageEvent.OrderID, damageEvent.Severity, damageEvent.Description)
	
	return damageEvent, nil
}

// Close closes the RabbitMQ connection and channel
func (adapter *OrderConsumerAdapter) Close() error {
	if adapter.channel != nil {
		adapter.channel.Close()
	}
	if adapter.conn != nil {
		return adapter.conn.Close()
	}
	return nil
}