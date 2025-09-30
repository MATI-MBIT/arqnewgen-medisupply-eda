package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MqttPublisher publishes OrderDamageEvent messages to an MQTT topic.
type MqttPublisher struct {
	client mqtt.Client
	Topic  string
}

// NewMqttPublisherFromEnv creates and connects an MQTT publisher using env vars.
// Env vars:
// - MQTT_BROKER (default: tcp://localhost:1883)
// - MQTT_PUB_CLIENT_ID (default: order-event-client-pub)
// - MQTT_PUB_TOPIC (default: events/order-damage)
// - MQTT_USERNAME (optional)
// - MQTT_PASSWORD (optional)
func NewMqttPublisherFromEnv() (*MqttPublisher, error) {
	broker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	clientID := getEnv("MQTT_PUB_CLIENT_ID", "order-event-client")
	topic := getEnv("MQTT_PUB_TOPIC", "events/order-damage")
	username := getEnv("MQTT_USERNAME", "")
	password := getEnv("MQTT_PASSWORD", "")

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)

	if username != "" {
		opts.SetUsername(username)
	}
	if password != "" {
		opts.SetPassword(password)
	}

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		// no-op; could log if needed
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		// no-op; could log if needed
		_ = err
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect error: %w", token.Error())
	}

	return &MqttPublisher{client: client, Topic: topic}, nil
}

// PublishOrderDamageFromSensor builds an OrderDamageEvent and publishes it as JSON to MQTT.
func (p *MqttPublisher) PublishOrderDamageFromSensor(ctx context.Context, sensorID, source string, temperature, humidity float64, status, mqttTopic string) error {
	if p == nil || p.client == nil {
		return nil
	}

	severity := deriveSeverity(temperature, humidity)
	evt := OrderDamageEvent{
		EventID:     sensorID,
		Type:        "order.damage",
		Source:      source,
		OccurredAt:  time.Now().UTC(),
		OrderID:     sensorID,
		Severity:    severity,
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

	// Respect context by polling until done or publish finished
	done := make(chan error, 1)
	go func() {
		t := p.client.Publish(p.Topic, 0, false, payload)
		t.Wait()
		done <- t.Error()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// Close disconnects the MQTT client.
func (p *MqttPublisher) Close() error {
	if p == nil || p.client == nil {
		return nil
	}
	p.client.Disconnect(250)
	return nil
}
