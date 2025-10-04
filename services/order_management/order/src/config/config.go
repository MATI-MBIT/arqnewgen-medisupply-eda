package config

import "os"

// Config holds all configuration for the application
type Config struct {
	RabbitMQ RabbitMQConfig
	HTTP     HTTPConfig
}

// RabbitMQConfig holds RabbitMQ-specific configuration
type RabbitMQConfig struct {
	URL          string
	ExchangeName string
	QueueName    string
	RoutingKey   string
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Port string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		RabbitMQ: RabbitMQConfig{
			URL:          getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			ExchangeName: getEnv("RABBITMQ_EXCHANGE", "order-exchange"),
			QueueName:    getEnv("RABBITMQ_QUEUE", "order-queue"),
			RoutingKey:   getEnv("RABBITMQ_ROUTING_KEY", "order.created"),
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