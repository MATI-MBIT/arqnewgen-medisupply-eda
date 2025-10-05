package config

import "os"

// Config holds all configuration for the application
type Config struct {
	Kafka KafkaConfig
	HTTP  HTTPConfig
}

// KafkaConfig holds Kafka-specific configuration
type KafkaConfig struct {
	OrderEventsTopic      string
	BatchEventsTopic      string
	BrokerAddress         string
	GroupID               string
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Port string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Kafka: KafkaConfig{
			OrderEventsTopic: getEnv("KAFKA_ORDER_EVENTS_TOPIC", "order-events"),
			BatchEventsTopic: getEnv("KAFKA_BATCH_EVENTS_TOPIC", "warehouse-batch-events"),
			BrokerAddress:    getEnv("KAFKA_BROKER_ADDRESS", "localhost:9092"),
			GroupID:          getEnv("KAFKA_GROUP_ID", "warehouse-batch-service"),
		},
		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8080"),
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