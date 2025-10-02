package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	// Límites de temperatura para la generación aleatoria
	MinTemperature = 10.0
	MaxTemperature = 30.0
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

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

type TemperatureLimitsRequest struct {
	MinTemperature float64 `json:"min_temperature"`
	MaxTemperature float64 `json:"max_temperature"`
}

type TemperatureLimitsResponse struct {
	Message        string    `json:"message"`
	MinTemperature float64   `json:"min_temperature"`
	MaxTemperature float64   `json:"max_temperature"`
	Timestamp      time.Time `json:"timestamp"`
}

func main() {
	// Configuración MQTT
	broker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	clientID := getEnv("MQTT_CLIENT_ID", "event-generator")
	topic := getEnv("MQTT_TOPIC", "events/sensor")
	username := getEnv("MQTT_USERNAME", "")
	password := getEnv("MQTT_PASSWORD", "")

	// Configuración de frecuencia de eventos
	eventInterval := getEventInterval("EVENT_INTERVAL_SECONDS", 10)

	// Configuración del servidor HTTP
	httpPort := getEnv("HTTP_PORT", "8080")

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

	// Iniciar servidor HTTP para health check
	server := &http.Server{
		Addr:    ":" + httpPort,
		Handler: setupRoutes(),
	}

	go func() {
		log.Printf("Servidor HTTP iniciado en puerto %s", httpPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error iniciando servidor HTTP: %v", err)
		}
	}()

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

			// Cerrar servidor HTTP
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				log.Printf("Error cerrando servidor HTTP: %v", err)
			}

			// Desconectar cliente MQTT
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
		log.Printf("Evento publicado en topic %s: %s", topic, event.ID)
	}
}

func generateEvent() Event {
	return Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Type:      "sensor_reading",
		Source:    "temperature_sensor_02",
		Data: EventData{
			Temperature: GetTemperatureRandom(MinTemperature, MaxTemperature),
			Humidity:    50.0 + (float64(time.Now().Unix()%30) - 15), // Simula humedad entre 35-65%
			Status:      "active",
		},
	}
}

// GetTemperatureRandom genera una temperatura aleatoria entre los límites especificados
func GetTemperatureRandom(lowerLimit, upperLimit float64) float64 {
	// Generar temperatura aleatoria entre lowerLimit y upperLimit
	return lowerLimit + rand.Float64()*(upperLimit-lowerLimit)
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

// Configuración de rutas HTTP
func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/temperature-limits", temperatureLimitsHandler)
	return mux
}

// Handler para el endpoint de health check
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "mqtt-event-generator",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Handler para el endpoint de configuración de límites de temperatura
func temperatureLimitsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TemperatureLimitsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validar que los valores sean válidos
	if req.MinTemperature >= req.MaxTemperature {
		http.Error(w, "MinTemperature must be less than MaxTemperature", http.StatusBadRequest)
		return
	}

	// Actualizar las variables globales
	MinTemperature = req.MinTemperature
	MaxTemperature = req.MaxTemperature

	log.Printf("Límites de temperatura actualizados: Min=%.2f, Max=%.2f", MinTemperature, MaxTemperature)

	// Responder con los nuevos valores
	response := TemperatureLimitsResponse{
		Message:        "Temperature limits updated successfully",
		MinTemperature: MinTemperature,
		MaxTemperature: MaxTemperature,
		Timestamp:      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding temperature limits response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
