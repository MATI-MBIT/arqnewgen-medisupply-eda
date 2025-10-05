# Warehouse Batch Service - Hexagonal Architecture

This service has been refactored to follow hexagonal architecture principles, providing clear separation of concerns and improved testability.

## Architecture Overview

```
services/warehouse/batch/
├── src/
│   ├── domain/                     # Core business entities and interfaces
│   │   └── order.go               # Order domain models and interfaces
│   ├── application/               # Business logic and use cases
│   │   └── order_service.go       # Order event processing business logic
│   ├── config/                    # Configuration management
│   │   └── config.go              # Environment variable configuration
│   ├── infrastructure/
│   │   └── driving-adapters/      # External interfaces that drive the application
│   │       ├── order_event_consumer_adapter.go  # Order event consumer adapter
│   │       └── api_service_adapter.go            # HTTP REST API adapter
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
- **Order**: Core domain entity representing an order
- **OrderEvent**: Domain entity representing order events
- **OrderEventHandler**: Interface defining the contract for order event handling

### Application Layer
- **OrderService**: Contains the business logic for processing order events

### Configuration Layer
- **Config**: Manages application configuration from environment variables with sensible defaults

### Infrastructure Layer

#### Driving Adapters
- **OrderEventConsumerAdapter**: 
  - **Architectural Role**: Adapter that subscribes to order events from Kafka
  - **Responsibility**: Listens for order events, parses JSON messages, and translates them into domain order events for processing
- **ApiServiceAdapter**: 
  - **Architectural Role**: HTTP REST API adapter that exposes application capabilities
  - **Responsibility**: Provides synchronous HTTP endpoints for health checks and batch management operations

#### Driven Adapters
- **BatchMemoryRepository**: 
  - **Architectural Role**: In-memory implementation of the batch repository
  - **Responsibility**: Provides data persistence for batch entities using in-memory storage
- **BatchEventPublisherAdapter**: 
  - **Architectural Role**: Kafka event publisher adapter for batch events
  - **Responsibility**: Publishes batch domain events to the warehouse-batch-events Kafka topic

## Key Benefits

1. **Separation of Concerns**: Each layer has a clear responsibility
2. **Testability**: Business logic is isolated and can be easily unit tested
3. **Flexibility**: Easy to swap out infrastructure components without affecting business logic
4. **Maintainability**: Clear boundaries make the code easier to understand and modify

## Configuration

The application can be configured using environment variables:

| Environment Variable | Default Value | Description |
|---------------------|---------------|-------------|
| `KAFKA_ORDER_EVENTS_TOPIC` | `order-events` | Kafka topic for consuming order events |
| `KAFKA_BATCH_EVENTS_TOPIC` | `warehouse-batch-events` | Kafka topic for publishing batch events |
| `KAFKA_BROKER_ADDRESS` | `localhost:9092` | Kafka broker address |
| `KAFKA_GROUP_ID` | `warehouse-batch-service` | Kafka consumer group ID |
| `HTTP_PORT` | `8080` | HTTP port for the API service adapter |

### Example Configuration

Copy the example environment file:
```bash
cp .env.example .env
```

Edit `.env` with your configuration:
```bash
KAFKA_ORDER_EVENTS_TOPIC=order-events
KAFKA_BATCH_EVENTS_TOPIC=warehouse-batch-events
KAFKA_BROKER_ADDRESS=kafka:9092
KAFKA_GROUP_ID=warehouse-batch-service
HTTP_PORT=8080
```

## Development

### Running Tests

```bash
# Run all tests
cd services/warehouse/batch
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./src/domain/
go test ./src/application/
```

### Building the Application

```bash
# Build the application
cd services/warehouse/batch
go build -o bin/warehouse-batch-service src/main.go

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/warehouse-batch-service-linux src/main.go
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

### Batch Management API (v1)

#### Get All Batches
- **Endpoint**: `GET /api/v1/batches`
- **Description**: Retrieves all batches in the system
- **Response**: 
  ```json
  {
    "batches": [
      {
        "id": "batch_123",
        "product_id": "prod_456",
        "status": "pending",
        "items": [
          {
            "order_id": "order_789",
            "product_id": "prod_456",
            "quantity": 5,
            "status": "allocated",
            "added_at": "2024-01-01T12:00:00Z",
            "processed_at": null
          }
        ],
        "total_items": 1,
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-01T12:00:00Z",
        "processed_at": null
      }
    ],
    "count": 1
  }
  ```

#### Get Batches by Product ID
- **Endpoint**: `GET /api/v1/batches/product/{productId}`
- **Description**: Retrieves all batches for a specific product
- **Parameters**: 
  - `productId` (path): The product identifier
- **Response**: 
  ```json
  {
    "product_id": "prod_456",
    "batches": [...],
    "count": 2
  }
  ```

#### Get Batches by Status
- **Endpoint**: `GET /api/v1/batches/status/{status}`
- **Description**: Retrieves all batches with a specific status
- **Parameters**: 
  - `status` (path): Batch status (pending, processing, completed, cancelled, damaged)
- **Response**: 
  ```json
  {
    "status": "pending",
    "batches": [...],
    "count": 3
  }
  ```

#### Get Batch by Order ID
- **Endpoint**: `GET /api/v1/batches/order/{orderId}`
- **Description**: Retrieves the batch containing a specific order
- **Parameters**: 
  - `orderId` (path): The order identifier
- **Response**: 
  ```json
  {
    "order_id": "order_789",
    "batch": {
      "id": "batch_123",
      "product_id": "prod_456",
      "status": "pending",
      "items": [...],
      "total_items": 1,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z",
      "processed_at": null
    }
  }
  ```

### Batch Status Values

The following status values are supported:
- `pending` - Batch is created and waiting for processing
- `processing` - Batch is currently being processed
- `completed` - Batch processing has been completed
- `cancelled` - Batch has been cancelled
- `damaged` - Batch contains damaged items

### Error Responses

All endpoints may return error responses in the following format:
```json
{
  "error": "Error description",
  "details": "Detailed error message"
}
```

Common HTTP status codes:
- `200 OK` - Successful request
- `400 Bad Request` - Invalid parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

### Testing the API

```bash
# Health check
curl http://localhost:8080/health

# Get all batches
curl http://localhost:8080/api/v1/batches

# Get batches for a specific product
curl http://localhost:8080/api/v1/batches/product/prod_456

# Get batches by status
curl http://localhost:8080/api/v1/batches/status/pending

# Get batch containing a specific order
curl http://localhost:8080/api/v1/batches/order/order_789

# With Docker
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/batches
```

## Event Processing

### Order Events Processing

The warehouse batch service consumes order events from the `order-events` topic. It handles the following event types:

- `order.damage_processed` - Processes damage reports and updates inventory status
- `order.created` - Allocates inventory for new orders
- `order.cancelled` - Releases allocated inventory
- `order.shipped` - Updates inventory after shipping
- `order.delivered` - Confirms delivery and closes warehouse operations
- `order.returned` - Processes returned items and updates inventory
- `order.inventory_allocated` - Confirms inventory allocation
- `order.inventory_released` - Confirms inventory release

### Order Event Format

The service expects order events in the following JSON format:

```json
{
  "event_type": "order.damage_processed",
  "order_id": "evt_1759598824",
  "order": {
    "id": "evt_1759598824",
    "customer_id": "unknown",
    "product_id": "unknown",
    "quantity": 1,
    "status": "damage_detected_minor",
    "total_amount": 0,
    "created_at": "2025-10-04T17:27:04.082881166Z",
    "updated_at": "2025-10-04T17:36:13.584671556Z"
  },
  "timestamp": "2025-10-04T17:36:13.58470126Z"
}
```

### Batch Events Publishing

The warehouse batch service publishes batch events to the `warehouse-batch-events` topic whenever significant batch operations occur. This enables other services to react to batch changes in real-time.

#### Published Event Types

- `batch.created` - Published when a new batch is created
- `batch.item_added` - Published when an order item is added to a batch
- `batch.item_removed` - Published when an order item is removed from a batch
- `batch.item_updated` - Published when an order item status is updated within a batch
- `batch.processing_started` - Published when batch processing begins
- `batch.completed` - Published when a batch is completed
- `batch.cancelled` - Published when a batch is cancelled
- `batch.marked_damaged` - Published when a batch is marked as damaged

#### Batch Event Format

All batch events follow this JSON structure:

```json
{
  "event_type": "batch.created",
  "batch_id": "BATCH-prod_456-20241201120000",
  "product_id": "prod_456",
  "batch": {
    "id": "BATCH-prod_456-20241201120000",
    "product_id": "prod_456",
    "status": "pending",
    "items": [
      {
        "order_id": "order_789",
        "product_id": "prod_456",
        "quantity": 5,
        "status": "allocated",
        "added_at": "2024-12-01T12:00:00Z",
        "processed_at": null
      }
    ],
    "total_items": 1,
    "created_at": "2024-12-01T12:00:00Z",
    "updated_at": "2024-12-01T12:00:00Z",
    "processed_at": null
  },
  "order_id": "order_789",
  "item_details": {
    "order_id": "order_789",
    "product_id": "prod_456",
    "quantity": 5,
    "status": "allocated",
    "added_at": "2024-12-01T12:00:00Z",
    "processed_at": null
  },
  "timestamp": "2024-12-01T12:00:00Z"
}
```

#### Event Headers

Each published event includes Kafka headers for efficient filtering and routing:

- `event_type` - The type of batch event
- `batch_id` - The batch identifier
- `product_id` - The product identifier
- `order_id` - The order identifier (for item-specific events)
- `timestamp` - Event timestamp in RFC3339 format

#### Event Partitioning

Events are partitioned by `batch_id` to ensure all events for a specific batch are processed in order by downstream consumers.

## Application Behavior

The application will:
1. Load configuration from environment variables (with fallback to defaults)
2. Initialize the batch event publisher for publishing to the warehouse-batch-events topic
3. Start the OrderEventConsumerAdapter to listen for order events
4. Start the ApiServiceAdapter to serve HTTP requests on the configured port
5. Process incoming order events through the OrderService
6. Publish batch events whenever batch operations occur (creation, updates, status changes)
7. Handle graceful shutdown on SIGINT/SIGTERM signals, including proper cleanup of Kafka connections

### Event Publishing Behavior

- **Synchronous Publishing**: Batch events are published synchronously to ensure data consistency
- **Error Handling**: Event publishing failures are logged but don't prevent the main operation from completing
- **Auto-Recovery**: Automatically recovers from "Unknown Topic Or Partition" errors by recreating the Kafka writer
- **Retry Logic**: Failed publishes due to topic/partition issues are automatically retried after writer recreation
- **Partitioning**: Events are partitioned by batch ID to maintain ordering for each batch
- **Headers**: Rich metadata is included in Kafka headers for efficient filtering and routing
- **Graceful Shutdown**: The event publisher is properly closed during application shutdown

#### Error Recovery

The event publisher includes automatic recovery for common Kafka topic issues:

1. **Topic Not Found**: When a topic doesn't exist or isn't available, the publisher detects the error
2. **Writer Recreation**: The Kafka writer is automatically recreated with fresh connections
3. **Retry Attempt**: The failed message is retried once with the new writer
4. **Logging**: All recovery attempts are logged for monitoring and debugging

This ensures reliable event publishing even when topics are created after the service starts or during temporary Kafka issues.

## Docker Image Features

- **Multi-stage build**: Optimized for size and security
- **Distroless base image**: Minimal attack surface with no shell or package manager
- **Non-root user**: Runs as `nonroot:nonroot` for enhanced security
- **Static binary**: No external dependencies required at runtime
- **CA certificates**: Included for secure external connections