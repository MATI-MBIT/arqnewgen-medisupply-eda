# MQTT Order Event Client

A Go microservice that acts as a client for the `mqtt-event-generator` service. It subscribes to MQTT events and exposes them through HTTP REST endpoints.

## Features

- **MQTT Subscription**: Subscribes to MQTT topics to receive events from mqtt-event-generator
- **Event Storage**: Stores received events in memory with configurable maximum size
- **REST API**: Provides HTTP endpoints to access stored events
- **Real-time Processing**: Processes events as they arrive
- **Statistics**: Calculates basic statistics from received events

## Configuration

The service can be configured using environment variables:

| Variable | Default Value | Description |
|----------|---------------|-------------|
| `MQTT_BROKER` | `tcp://localhost:1883` | MQTT broker URL |
| `MQTT_CLIENT_ID` | `order-event-client` | MQTT client identifier |
| `MQTT_TOPIC` | `events/sensor` | MQTT topic to subscribe to |
| `MQTT_USERNAME` | `` | MQTT username (optional) |
| `MQTT_PASSWORD` | `` | MQTT password (optional) |
| `HTTP_PORT` | `8080` | HTTP server port |

## API Endpoints

### Health Check
- **GET** `/health`
- Returns service health status

### Get All Events
- **GET** `/events`
- Returns all stored events with count

### Get Latest Event
- **GET** `/events/latest`
- Returns the most recent event

### Get Events Count
- **GET** `/events/count`
- Returns the number of stored events

### Get Events Statistics
- **GET** `/events/stats`
- Returns calculated statistics including:
  - Total event count
  - Average temperature
  - Average humidity
  - Number of active sensors
  - Latest event

## Running the Service

### Prerequisites
- Go 1.21 or higher
- Access to an MQTT broker
- Running `mqtt-event-generator` service (optional, for testing)

### Installation
```bash
# Navigate to the service directory
cd services/mqtt-order-event-client

# Download dependencies
go mod tidy

# Run the service
go run main.go
```

### Using Docker (Optional)
```bash
# Build Docker image
docker build -t mqtt-order-event-client .

# Run container
docker run -p 8080:8080 \
  -e MQTT_BROKER=tcp://your-mqtt-broker:1883 \
  -e MQTT_TOPIC=events/sensor \
  mqtt-order-event-client
```

## Example Usage

### Start the service
```bash
go run main.go
```

### Test the endpoints
```bash
# Check service health
curl http://localhost:8080/health

# Get all events
curl http://localhost:8080/events

# Get latest event
curl http://localhost:8080/events/latest

# Get event statistics
curl http://localhost:8080/events/stats
```

## Event Format

The service expects events in the following JSON format (same as mqtt-event-generator):

```json
{
  "id": "evt_1234567890",
  "timestamp": "2023-12-07T10:30:00Z",
  "type": "sensor_reading",
  "source": "temperature_sensor_01",
  "data": {
    "temperature": 23.5,
    "humidity": 45.2,
    "status": "active"
  }
}
```

## Architecture

```
mqtt-event-generator → MQTT Broker → mqtt-order-event-client → HTTP API
```

1. **mqtt-event-generator** publishes events to MQTT broker
2. **mqtt-order-event-client** subscribes to MQTT topics
3. Events are stored in memory and exposed via REST API
4. Clients can query events through HTTP endpoints

## Development

### Project Structure
```
mqtt-order-event-client/
├── main.go          # Main application code
├── go.mod           # Go module dependencies
└── README.md        # This documentation
```

### Adding New Features
- Modify the `Event` struct to support new event types
- Add new HTTP endpoints in the `startHTTPServer` function
- Extend the `EventStore` for additional storage capabilities

## Monitoring

The service provides basic logging for:
- MQTT connection status
- Event reception and processing
- HTTP server status
- Error conditions

Monitor the logs to ensure proper operation and troubleshoot issues.
