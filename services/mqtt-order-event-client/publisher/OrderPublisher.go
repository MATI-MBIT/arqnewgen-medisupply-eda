package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	kplain "github.com/segmentio/kafka-go/sasl/plain"
)

// OrderDamageEvent is the entity published to Kafka for order damage notifications
// Kept here to encapsulate the public schema used on the Kafka topic.
type OrderDamageEvent struct {
	EventID     string        `json:"eventId"`
	Type        string        `json:"type"`   // e.g. "order.damage"
	Source      string        `json:"source"` // service/source name
	OccurredAt  time.Time     `json:"occurredAt"`
	OrderID     string        `json:"orderId"`
	Severity    string        `json:"severity"` // minor|major|critical
	Description string        `json:"description"`
	Details     DamageDetails `json:"details"`
}

type DamageDetails struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Status      string  `json:"status"`
	MqttTopic   string  `json:"mqttTopic"`
}

// Publisher wraps a Kafka writer.
type Publisher struct {
	writer *kafka.Writer
	Topic  string
}

// NewPublisherFromEnv creates a Kafka publisher using environment variables.
// Env vars:
// - KAFKA_BROKERS (comma-separated, default: kafka:9092)
// - KAFKA_TOPIC (default: order-status-events)
// - KAFKA_SASL_ENABLE (true/false, default: false)
// - KAFKA_USERNAME, KAFKA_PASSWORD (when SASL enabled)
func NewPublisherFromEnv() (*Publisher, error) {
	brokers := getEnv("KAFKA_BROKERS", "kafka:9092")
	topic := getEnv("KAFKA_TOPIC", "order-status-events")
	saslEnable := strings.ToLower(getEnv("KAFKA_SASL_ENABLE", "false")) == "true"
	username := getEnv("KAFKA_USERNAME", "")
	password := getEnv("KAFKA_PASSWORD", "")

	var transport kafka.RoundTripper
	if saslEnable && username != "" {
		mech := kplain.Mechanism{Username: username, Password: password}
		transport = &kafka.Transport{SASL: mech}
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(brokers, ",")...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		BatchTimeout: 200 * time.Millisecond,
		Transport:    transport,
	}

	return &Publisher{writer: w, Topic: topic}, nil
}

// Close closes the underlying Kafka writer.
func (p *Publisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

// PublishOrderDamageFromSensor builds and publishes an OrderDamageEvent using sensor data.
func (p *Publisher) PublishOrderDamageFromSensor(ctx context.Context, sensorID, source string, temperature, humidity float64, status, mqttTopic string) error {
	if p == nil || p.writer == nil {
		return nil
    }

    severity := deriveSeverity(temperature, humidity)
    evt := OrderDamageEvent{
        EventID:    sensorID,
        Type:       "order.damage",
        Source:     source,
        OccurredAt: time.Now().UTC(),
        OrderID:    sensorID,
        Severity:   severity,
        Description: fmt.Sprintf("Potential damage detected: temp=%.2fC, humidity=%.2f%%", temperature, humidity),
        Details: DamageDetails{
            Temperature: temperature,
            Humidity:    humidity,
            Status:      status,
            MqttTopic:   mqttTopic,
        },
    }

    payload, err := json.Marshal(evt)
    if err != nil {
        return err
    }

    return p.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte(evt.OrderID),
        Value: payload,
    })
}

func deriveSeverity(temperature, humidity float64) string {
	severity := "minor"
	if temperature >= 30 || humidity >= 80 {
		severity = "major"
	}
	if temperature >= 40 || humidity >= 90 {
		severity = "critical"
	}
	return severity
}

// Deprecated: kept for backward-compatibility if needed.
// Prefer using fmt.Sprintf directly where needed.
func formatFloat(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", v), "0"), ".")
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
