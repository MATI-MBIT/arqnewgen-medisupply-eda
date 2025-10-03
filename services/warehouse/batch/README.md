# Warehouse Batch Service - Hexagonal Architecture

This service has been refactored to follow hexagonal architecture principles, providing clear separation of concerns and improved testability.

## Architecture Overview

```
services/warehouse/batch/
├── src/
│   ├── domain/                     # Core business entities and interfaces
│   │   └── event.go               # Event domain model and EventHandler interface
│   ├── application/               # Business logic and use cases
│   │   └── event_service.go       # Event processing business logic
│   ├── config/                    # Configuration management
│   │   └── config.go              # Environment variable configuration
│   ├── infrastructure/
│   │   ├── driving-adapters/      # External interfaces that drive the application
│   │   │   ├── event_consumer_adapter.go  # Kafka event consumer adapter
│   │   │   └── api_service_adapter.go     # HTTP REST API adapter
│   │   └── driven-adapters/       # External dependencies driven by the application
│   │       └── kafka_producer.go  # Kafka producer adapter
│   └── main.go                    # Application entry point and dependency injection
├── deployment/
│   └── Dockerfile                 # Multi-stage Docker build configuration
├── .dockerignore                  # Docker build context exclusions
├── .env.example                   # Environment variables template
├── go.mod                         # Go module definition
├── go.sum                         # Go module checksums
└── README.md                      # This documentation
```

## Components

### Domain Layer
- **Event**: Core domain entity representing an event
- **EventHandler**: Interface defining the contract for event handling

### Application Layer
- **EventService**: Contains the business logic for processing events

### Configuration Layer
- **Config**: Manages application configuration from environment variables with sensible defaults

### Infrastructure Layer

#### Driving Adapters
- **EventConsumerAdapter**: 
  - **Architectural Role**: Adapter that subscribes to message-oriented middleware
  - **Responsibility**: Listens for asynchronous messages from Kafka, translates external events into domain events that the application layer can understand and execute
- **ApiServiceAdapter**: 
  - **Architectural Role**: HTTP REST API adapter that exposes application capabilities
  - **Responsibility**: Provides synchronous HTTP endpoints for health checks and potential future API operations

#### Driven Adapters
- **KafkaProducer**: Handles message production to Kafka (for demo purposes)

## Key Benefits

1. **Separation of Concerns**: Each layer has a clear responsibility
2. **Testability**: Business logic is isolated and can be easily unit tested
3. **Flexibility**: Easy to swap out infrastructure components without affecting business logic
4. **Maintainability**: Clear boundaries make the code easier to understand and modify

## Configuration

The application can be configured using environment variables:

| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `KAFKA_TOPIC` | `my-topic` | Kafka topic to consume from and produce to |
| `KAFKA_BROKER_ADDRESS` | `localhost:9092` | Kafka broker address |
| `HTTP_PORT` | `8080` | HTTP port for the API service adapter |

### Example Configuration

Copy the example environment file:
```bash
cp .env.example .env
```

Edit `.env` with your configuration:
```bash
KAFKA_TOPIC=warehouse-events
KAFKA_BROKER_ADDRESS=kafka:9092
HTTP_PORT=8080
```

## Running the Application

### Local Development

#### With default configuration:
```bash
cd services/warehouse/batch
go run src/main.go
```

#### With custom environment variables:
```bash
cd services/warehouse/batch
export KAFKA_TOPIC=warehouse-events
export KAFKA_BROKER_ADDRESS=kafka:9092
export HTTP_PORT=8080
go run src/main.go
```

#### Using .env file (with a tool like `direnv` or manually):
```bash
cd services/warehouse/batch
source .env
go run src/main.go
```

### Docker Deployment

#### Build the Docker image:
```bash
cd services/warehouse/batch
docker build -f deployment/Dockerfile -t warehouse-batch-service:latest .
```

#### Run with Docker:
```bash
# With default configuration
docker run --rm warehouse-batch-service:latest

# With custom environment variables
docker run --rm \
  -p 8080:8080 \
  -e KAFKA_TOPIC=warehouse-events \
  -e KAFKA_BROKER_ADDRESS=kafka:9092 \
  -e HTTP_PORT=8080 \
  warehouse-batch-service:latest

# With Docker Compose (if you have a docker-compose.yml)
docker-compose up -d
```

#### Multi-platform build:
```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  -f deployment/Dockerfile \
  -t warehouse-batch-service:latest .
```

### Kubernetes Deployment

The Docker image is designed to work seamlessly in Kubernetes environments:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: warehouse-batch-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: warehouse-batch-service
  template:
    metadata:
      labels:
        app: warehouse-batch-service
    spec:
      containers:
      - name: warehouse-batch-service
        image: warehouse-batch-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: KAFKA_TOPIC
          value: "warehouse-events"
        - name: KAFKA_BROKER_ADDRESS
          value: "kafka:9092"
        - name: HTTP_PORT
          value: "8080"
```

## HTTP API Endpoints

The ApiServiceAdapter exposes the following HTTP endpoints:

### Health Check
- **Endpoint**: `GET /health`
- **Description**: Returns the health status of the service
- **Response**: 
  ```json
  {
    "status": "healthy",
    "service": "warehouse-batch",
    "timestamp": "2024-01-01T12:00:00Z"
  }
  ```

### Testing the API

```bash
# Health check
curl http://localhost:8080/health

# With Docker
curl http://localhost:8080/health
```

## Application Behavior

The application will:
1. Load configuration from environment variables (with fallback to defaults)
2. Start the EventConsumerAdapter to listen for Kafka messages
3. Start the ApiServiceAdapter to serve HTTP requests on the configured port
4. Start a KafkaProducer (for demo purposes) to generate test messages
5. Process incoming events through the EventService
6. Handle graceful shutdown on SIGINT/SIGTERM signals

## Docker Image Features

- **Multi-stage build**: Optimized for size and security
- **Distroless base image**: Minimal attack surface with no shell or package manager
- **Non-root user**: Runs as `nonroot:nonroot` for enhanced security
- **Static binary**: No external dependencies required at runtime
- **CA certificates**: Included for secure external connections