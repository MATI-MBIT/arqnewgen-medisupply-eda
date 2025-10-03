package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Global constants for RabbitMQ configuration
const (
	amqpURL      = "amqp://guest:guest@localhost:5672/"
	exchangeName = "my-direct-exchange"
	queueName    = "my-queue"
	routingKey   = "my-key"
)

// failOnError is a helper function to exit if an error occurs
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// publish sends messages to the RabbitMQ exchange
func publish(ctx context.Context, conn *amqp.Connection) {
	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare the exchange
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// Loop to send a message every second
	for {
		select {
		case <-ctx.Done():
			log.Println("Publisher shutting down...")
			return
		case <-time.After(1 * time.Second):
			body := fmt.Sprintf("Hello RabbitMQ at %s", time.Now().Format(time.RFC3339))
			err = ch.PublishWithContext(ctx,
				exchangeName, // exchange
				routingKey,   // routing key
				false,        // mandatory
				false,        // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				})
			if err != nil {
				log.Println("Failed to publish a message:", err)
			} else {
				log.Println("Message sent successfully")
			}
		}
	}
}

// consume reads messages from the RabbitMQ queue
func consume(ctx context.Context, conn *amqp.Connection) {
	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare the queue we will be consuming from.
	// Declaring a queue is idempotent; it will only be created if it doesn't exist.
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Bind the queue to the exchange with the routing key
	err = ch.QueueBind(
		q.Name,       // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	// Start consuming messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack is false, we will manually acknowledge
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	log.Println("Waiting for messages. To exit press CTRL+C")
	// Loop to process messages
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer shutting down...")
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			log.Printf("Received a message: %s", d.Body)
			// Acknowledge the message so RabbitMQ knows it can be deleted
			d.Ack(false)
		}
	}
}

func main() {
	// Connect to RabbitMQ server
	conn, err := amqp.Dial(amqpURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Run publisher and consumer in separate goroutines
	go publish(ctx, conn)
	go consume(ctx, conn)

	// Set up graceful shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan

	log.Println("Shutdown signal received, cancelling context...")
	cancel()

	time.Sleep(1 * time.Second)
	log.Println("Application shut down gracefully.")
}