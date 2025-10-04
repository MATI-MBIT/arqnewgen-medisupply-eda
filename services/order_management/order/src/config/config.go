package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the application
type Config struct {
	RabbitMQ RabbitMQConfig
	HTTP     HTTPConfig
}

// RabbitMQConfig holds RabbitMQ-specific configuration
type RabbitMQConfig struct {
	URL          string
	ExchangeName string
	// Consumer configuration (for receiving damage events)
	ConsumerQueueName    string
	ConsumerRoutingKey   string
	// Publisher configuration (for publishing order events)
	PublisherQueueName   string
	PublisherRoutingKey  string
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Port string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Construct RabbitMQ URL from components if individual parts are provided
	rabbitmqURL := getEnv("RABBITMQ_URL", "")
	if rabbitmqURL == "" {
		// Build URL from components
		host := getEnv("RABBITMQ_HOST", "localhost")
		port := getEnv("RABBITMQ_PORT", "5672")
		user := getEnv("RABBITMQ_USER", "guest")
		password := getEnv("RABBITMQ_PASSWORD", "guest")
		rabbitmqURL = fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
		fmt.Printf("DEBUG: Built RabbitMQ URL from components - Host: %s, Port: %s, User: %s\n", host, port, user)
	} else {
		fmt.Printf("DEBUG: Using provided RABBITMQ_URL\n")
	}

	return &Config{
		RabbitMQ: RabbitMQConfig{
			URL:          rabbitmqURL,
			ExchangeName: getEnv("RABBITMQ_EXCHANGE", "events"),
			// Consumer configuration (for receiving damage events)
			ConsumerQueueName:    getEnv("RABBITMQ_CONSUMER_QUEUE", "order-damage-queue"),
			ConsumerRoutingKey:   getEnv("RABBITMQ_CONSUMER_ROUTING_KEY", "order.damage"),
			// Publisher configuration (for publishing order events)
			PublisherQueueName:   getEnv("RABBITMQ_PUBLISHER_QUEUE", "order-events-queue"),
			PublisherRoutingKey:  getEnv("RABBITMQ_PUBLISHER_ROUTING_KEY", "order.events"),
		},
		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8081"),
		},
	}
}

// getEnv returns environment variable value or default if not set
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}