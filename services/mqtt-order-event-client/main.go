package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	publisher "mqtt-order-event-client/publisher"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

// Event represents the structure of events received from mqtt-event-generator
type Event struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Data      EventData `json:"data"`
}

// EventData represents the data payload of an event
type EventData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Status      string  `json:"status"`
}

// EventStore manages received events in memory
type EventStore struct {
	mu      sync.RWMutex
	events  []Event
	maxSize int
}

// NewEventStore creates a new event store with a maximum size
func NewEventStore(maxSize int) *EventStore {
	return &EventStore{
		events:  make([]Event, 0),
		maxSize: maxSize,
	}
}

// AddEvent adds a new event to the store
func (es *EventStore) AddEvent(event Event) {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.events = append(es.events, event)

	// Keep only the last maxSize events
	if len(es.events) > es.maxSize {
		es.events = es.events[len(es.events)-es.maxSize:]
	}
}

// GetEvents returns all stored events
func (es *EventStore) GetEvents() []Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	// Return a copy to avoid race conditions
	events := make([]Event, len(es.events))
	copy(events, es.events)
	return events
}

// GetLatestEvent returns the most recent event
func (es *EventStore) GetLatestEvent() *Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	if len(es.events) == 0 {
		return nil
	}

	return &es.events[len(es.events)-1]
}

// GetEventCount returns the number of stored events
func (es *EventStore) GetEventCount() int {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return len(es.events)
}

var eventStore *EventStore
var orderPublisher *publisher.MqttPublisher

func main() {
	// Initialize event store with max 1000 events
	eventStore = NewEventStore(1000)

	// MQTT Configuration
	broker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	clientID := getEnv("MQTT_CLIENT_ID", "order-event-client")
	topic := getEnv("MQTT_TOPIC", "events/order-damage")
	username := getEnv("MQTT_USERNAME", "")
	password := getEnv("MQTT_PASSWORD", "")

	// HTTP Server Configuration
	httpPort := getEnv("HTTP_PORT", "8080")

	// Configure MQTT client options
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

	// Configure MQTT callbacks
	opts.SetDefaultPublishHandler(messageHandler)
	opts.SetOnConnectHandler(connectHandler)
	opts.SetConnectionLostHandler(connectLostHandler)

	// Create MQTT client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}

	log.Printf("Connected to MQTT broker: %s", broker)

	// Subscribe to the topic
	if token := client.Subscribe(topic, 1, nil); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to topic %s: %v", topic, token.Error())
	}

	log.Printf("Subscribed to topic: %s", topic)

	// Initialize MQTT Order Publisher from environment
	var err error
	orderPublisher, err = publisher.NewMqttPublisherFromEnv()
	if err != nil {
		log.Printf("Warning: could not initialize MQTT publisher: %v", err)
	} else {
		defer func() {
			if err := orderPublisher.Close(); err != nil {
				log.Printf("Error closing MQTT publisher: %v", err)
			}
		}()
	}

	// Start HTTP server in a goroutine
	go startHTTPServer(httpPort)

	// Handle system signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("MQTT Order Event Client started. HTTP server on port %s", httpPort)
	log.Println("Press Ctrl+C to exit...")

	// Wait for termination signal
	<-sigChan
	log.Println("Received termination signal, shutting down...")

	// Unsubscribe and disconnect
	if token := client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		log.Printf("Error unsubscribing: %v", token.Error())
	}

	client.Disconnect(250)
	log.Println("MQTT client disconnected")
}

// startHTTPServer starts the HTTP server with REST endpoints
func startHTTPServer(port string) {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "mqtt-order-event-client",
			"timestamp": time.Now(),
		})
	})

	// Get all events endpoint
	router.GET("/events", func(c *gin.Context) {
		events := eventStore.GetEvents()
		c.JSON(http.StatusOK, gin.H{
			"events": events,
			"count":  len(events),
		})
	})

	// Get latest event endpoint
	router.GET("/events/latest", func(c *gin.Context) {
		event := eventStore.GetLatestEvent()
		if event == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No events available",
			})
			return
		}
		c.JSON(http.StatusOK, event)
	})

	// Get events count endpoint
	router.GET("/events/count", func(c *gin.Context) {
		count := eventStore.GetEventCount()
		c.JSON(http.StatusOK, gin.H{
			"count": count,
		})
	})

	// Get events statistics endpoint
	router.GET("/events/stats", func(c *gin.Context) {
		events := eventStore.GetEvents()

		if len(events) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"count": 0,
				"stats": "No events available",
			})
			return
		}

		// Calculate basic statistics
		var totalTemp, totalHumidity float64
		var activeCount int

		for _, event := range events {
			totalTemp += event.Data.Temperature
			totalHumidity += event.Data.Humidity
			if event.Data.Status == "active" {
				activeCount++
			}
		}

		avgTemp := totalTemp / float64(len(events))
		avgHumidity := totalHumidity / float64(len(events))

		c.JSON(http.StatusOK, gin.H{
			"count":               len(events),
			"average_temperature": fmt.Sprintf("%.2f", avgTemp),
			"average_humidity":    fmt.Sprintf("%.2f", avgHumidity),
			"active_sensors":      activeCount,
			"latest_event":        events[len(events)-1],
		})
	})

	// Start server
	log.Printf("Starting HTTP server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// MQTT message handler - processes incoming events
var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message from topic %s", msg.Topic())

	var event Event
	if err := json.Unmarshal(msg.Payload(), &event); err != nil {
		log.Printf("Error unmarshaling event: %v", err)
		return
	}

	// Store the event
	eventStore.AddEvent(event)

	log.Printf("Event stored: ID=%s, Type=%s, Source=%s, Temp=%.2f°C, Humidity=%.2f%%",
		event.ID, event.Type, event.Source, event.Data.Temperature, event.Data.Humidity)
	log.Printf("Publishing order damage event for sensor/order id=%s", event.ID)
	// Publish order damage event to Kafka after logging and storing
	if orderPublisher != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := orderPublisher.PublishOrderDamageFromSensor(
			ctx,
			event.ID,
			"mqtt-order-event-client",
			event.Data.Temperature,
			event.Data.Humidity,
			event.Data.Status,
			msg.Topic(),
		); err != nil {
			log.Printf("Error publishing order damage event: %v", err)
		} else {
			log.Printf("Order damage event published for sensor/order id=%s", event.ID)
		}
	} else {
		log.Printf("Order publisher not initialized, skipping order damage event")
	}
}

// MQTT connection handler
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("MQTT client connected successfully")
}

// MQTT connection lost handler
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("MQTT connection lost: %v", err)
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
