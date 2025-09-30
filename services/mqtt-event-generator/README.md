# MQTT Event Generator

Generador de eventos MQTT en Go que publica eventos cada 30 segundos a un servidor EMQX.

## Características

- Genera eventos simulados de sensores cada 30 segundos
- Se conecta a EMQX via protocolo MQTT
- Servidor HTTP con endpoint de health check en `/health`
- Configurable mediante variables de entorno
- Manejo graceful de señales del sistema
- Logs detallados de conexión y publicación

## Configuración

El servicio se configura mediante las siguientes variables de entorno:

| Variable | Descripción | Valor por defecto |
|----------|-------------|-------------------|
| `MQTT_BROKER` | URL del broker MQTT | `tcp://localhost:1883` |
| `MQTT_CLIENT_ID` | ID del cliente MQTT | `event-generator` |
| `MQTT_TOPIC` | Topic donde publicar eventos | `events/sensor` |
| `EVENT_INTERVAL_SECONDS` | Frecuencia de eventos en segundos | `30` |
| `HTTP_PORT` | Puerto del servidor HTTP para health check | `8080` |
| `MQTT_USERNAME` | Usuario para autenticación | (vacío) |
| `MQTT_PASSWORD` | Contraseña para autenticación | (vacío) |

## Uso

### Desarrollo local

```bash
# Instalar dependencias
go mod tidy

# Ejecutar el generador
go run main.go

# O compilar y ejecutar
go build -o event-generator
./event-generator
```

### Con variables de entorno

```bash
export MQTT_BROKER="tcp://emqx:1883"
export MQTT_TOPIC="sensors/temperature"
export MQTT_USERNAME="admin"
export MQTT_PASSWORD="public"
export EVENT_INTERVAL_SECONDS="10"
go run main.go
```

## Estructura del evento

Los eventos generados tienen la siguiente estructura JSON:

```json
{
  "id": "evt_1703123456",
  "timestamp": "2023-12-21T10:30:45Z",
  "type": "sensor_reading",
  "source": "temperature_sensor_01",
  "data": {
    "temperature": 23.5,
    "humidity": 45.2,
    "status": "active"
  }
}
```

## Health Check

El servicio incluye un endpoint HTTP para health checks:

```bash
# Verificar estado del servicio
curl http://localhost:8080/health
```

Respuesta esperada:

```json
{
  "status": "healthy",
  "timestamp": "2023-12-21T10:30:45Z",
  "service": "mqtt-event-generator"
}
```

## Docker

Para ejecutar en contenedor, crear un Dockerfile:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o event-generator main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/event-generator .
CMD ["./event-generator"]
```