package drivenadapters

import (
	"context"
	"encoding/json"
	"log"

	"github.com/MATI-MBIT/arqnewgen-medisupply-eda/simple-service/oder/src/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher handles message publication to RabbitMQ
type RabbitMQPublisher struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
	queueName    string
	routingKey   string
}

// NewRabbitMQPublisher creates a new RabbitMQPublisher
func NewRabbitMQPublisher(rabbitMQURL, exchangeName, queueName, routingKey string) (*RabbitMQPublisher, error) {
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

	// Declare the queue for order events
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

	log.Printf("RabbitMQ Publisher initialized - Exchange: %s, Queue: %s, RoutingKey: %s", 
		exchangeName, queueName, routingKey)

	return &RabbitMQPublisher{
		conn:         conn,
		channel:      channel,
		exchangeName: exchangeName,
		queueName:    queueName,
		routingKey:   routingKey,
	}, nil
}

// PublishOrderEvent publishes an order event to RabbitMQ
func (p *RabbitMQPublisher) PublishOrderEvent(event domain.OrderEvent) error {
	// Marshal the event to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publish the message
	err = p.channel.PublishWithContext(
		context.Background(),
		p.exchangeName, // exchange
		p.routingKey,   // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("Failed to publish order event: %v", err)
		return err
	}

	log.Printf("Order event published successfully: Type=%s, OrderID=%s", event.EventType, event.OrderID)
	return nil
}

// PublishMessage publishes a simple message to RabbitMQ (for demo purposes)
func (p *RabbitMQPublisher) PublishMessage(ctx context.Context, message string) error {
	err := p.channel.PublishWithContext(ctx,
		p.exchangeName, // exchange
		p.routingKey,   // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})

	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		return err
	}

	log.Printf("Message published successfully: %s", message)
	return nil
}

// Close closes the RabbitMQ connection and channel
func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}