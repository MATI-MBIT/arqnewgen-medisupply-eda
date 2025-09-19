package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Event struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Data      EventData `json:"data"`
}

type EventData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Status      string  `json:"status"`
}

func main() {
	// Configuración MQTT
	broker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	clientID := getEnv("MQTT_CLIENT_ID", "event-generator")
	topic := getEnv("MQTT_TOPIC", "events/sensor")
	username := getEnv("MQTT_USERNAME", "")
	password := getEnv("MQTT_PASSWORD", "")
	
	// Configuración de frecuencia de eventos
	eventInterval := getEventInterval("EVENT_INTERVAL_SECONDS", 30)

	// Configurar opciones del cliente MQTT
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

	// Configurar callbacks
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.SetOnConnectHandler(connectHandler)
	opts.SetConnectionLostHandler(connectLostHandler)

	// Crear cliente MQTT
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error conectando a MQTT broker: %v", token.Error())
	}

	log.Printf("Conectado al broker MQTT: %s", broker)
	log.Printf("Publicando eventos en el topic: %s", topic)
	log.Printf("Frecuencia de eventos: cada %d segundos", eventInterval)

	// Canal para manejar señales del sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ticker para generar eventos según la configuración
	ticker := time.NewTicker(time.Duration(eventInterval) * time.Second)
	defer ticker.Stop()

	// Generar evento inicial
	publishEvent(client, topic)

	// Loop principal
	for {
		select {
		case <-ticker.C:
			publishEvent(client, topic)
		case <-sigChan:
			log.Println("Recibida señal de terminación, cerrando...")
			client.Disconnect(250)
			return
		}
	}
}

func publishEvent(client mqtt.Client, topic string) {
	event := generateEvent()
	
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error serializando evento: %v", err)
		return
	}

	token := client.Publish(topic, 0, false, payload)
	token.Wait()
	
	if token.Error() != nil {
		log.Printf("Error publicando evento: %v", token.Error())
	} else {
		log.Printf("Evento publicado: %s", event.ID)
	}
}

func generateEvent() Event {
	return Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Type:      "sensor_reading",
		Source:    "temperature_sensor_01",
		Data: EventData{
			Temperature: 20.0 + (float64(time.Now().Unix()%20) - 10), // Simula temperatura entre 10-30°C
			Humidity:    50.0 + (float64(time.Now().Unix()%30) - 15), // Simula humedad entre 35-65%
			Status:      "active",
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEventInterval(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if interval, err := strconv.Atoi(value); err == nil && interval > 0 {
			return interval
		}
		log.Printf("⚠️  Valor inválido para %s: %s, usando valor por defecto: %d segundos", key, value, defaultValue)
	}
	return defaultValue
}

// Callbacks MQTT
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Mensaje recibido: %s desde topic: %s", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Cliente MQTT conectado")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Conexión MQTT perdida: %v", err)
}