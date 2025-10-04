# Order Management Service - Hexagonal Architecture Refactoring

## Overview

The order management microservice has been successfully refactored from a simple monolithic structure to follow the hexagonal architecture pattern (ports and adapters), similar to the warehouse batch project but adapted for RabbitMQ instead of Kafka.

## Architecture Changes

### Before (Monolithic)
```
src/
└── main.go  # All logic in one file
```

### After (Hexagonal Architecture)
```
src/
├── main.go                           # Application entry point & dependency injection
├── config/                           # Configuration management
│   └── config.go
├── domain/                           # Business domain (core)
│   └── order.go                      # Order entities and interfaces (ports)
├── application/                      # Business logic (use cases)
│   └── order_service.go              # Order business operations
└── infrastructure/                   # External adapters
    ├── driving-adapters/             # Input adapters (primary ports)
    │   ├── api_service_adapter.go    # HTTP REST API
    │   └── order_consumer_adapter.go # RabbitMQ consumer
    └── driven-adapters/              # Output adapters (secondary ports)
        ├── memory_order_repository.go # In-memory data storage
        └── rabbitmq_publisher.go     # RabbitMQ event publisher
```

## Key Improvements

### 1. Separation of Concerns
- **Domain Layer**: Pure business entities and interfaces
- **Application Layer**: Business logic without infrastructure dependencies
- **Infrastructure Layer**: External system integrations

### 2. Dependency Inversion
- Business logic depends on abstractions (interfaces), not concrete implementations
- Infrastructure adapters implement the domain interfaces
- Easy to swap implementations (e.g., memory storage → database)

### 3. Testability
- Business logic can be tested in isolation
- Mock implementations can be easily created for interfaces
- Each adapter can be tested independently

### 4. Maintainability
- Clear boundaries between layers
- Changes in external systems only affect their respective adapters
- Business rules are centralized in the application layer

## RabbitMQ vs Kafka Adaptations

| Aspect | Warehouse (Kafka) | Order Management (RabbitMQ) |
|--------|-------------------|----------------------------|
| **Message Broker** | Kafka with segmentio/kafka-go | RabbitMQ with amqp091-go |
| **Connection Model** | Reader/Writer | Connection/Channel |
| **Message Routing** | Topics and partitions | Exchanges, queues, routing keys |
| **Consumer Groups** | Kafka consumer groups | RabbitMQ queues with multiple consumers |
| **Message Format** | Simple key-value | JSON events with structured data |

## New Features Added

### 1. RESTful API
- Complete CRUD operations for orders
- Health check endpoint
- Proper HTTP status codes and error handling

### 2. Event-Driven Architecture
- Order events published on create/update
- Structured event format with metadata
- Event consumption for processing external events

### 3. Configuration Management
- Environment-based configuration
- Separate config layer with defaults
- Support for .env files

### 4. Graceful Shutdown
- Context-based cancellation
- Proper resource cleanup
- Signal handling for containers

## Domain Model

### Order Entity
```go
type Order struct {
    ID          string    `json:"id"`
    CustomerID  string    `json:"customer_id"`
    ProductID   string    `json:"product_id"`
    Quantity    int       `json:"quantity"`
    Status      string    `json:"status"`
    TotalAmount float64   `json:"total_amount"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Event Model
```go
type OrderEvent struct {
    EventType string    `json:"event_type"`
    OrderID   string    `json:"order_id"`
    Order     Order     `json:"order"`
    Timestamp time.Time `json:"timestamp"`
}
```

## Ports (Interfaces)

### Primary Ports (Driving)
- `OrderEventHandler` - For processing incoming events

### Secondary Ports (Driven)
- `OrderRepository` - For data persistence
- `OrderEventPublisher` - For event publishing

## Adapters Implementation

### Driving Adapters (Input)
1. **ApiServiceAdapter** - HTTP REST API using Gin framework
2. **OrderConsumerAdapter** - RabbitMQ message consumer

### Driven Adapters (Output)
1. **MemoryOrderRepository** - In-memory data storage
2. **RabbitMQPublisher** - Event publishing to RabbitMQ

## Configuration

### Environment Variables
```bash
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=order-exchange
RABBITMQ_QUEUE=order-queue
RABBITMQ_ROUTING_KEY=order.created
HTTP_PORT=8081
```

## Dependencies Added
- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/google/uuid` - UUID generation
- `github.com/joho/godotenv` - Environment variable loading

## Usage Examples

### API Calls
```bash
# Create order
curl -X POST http://localhost:8081/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "customer-123", "product_id": "product-456", "quantity": 2, "total_amount": 99.99}'

# Get all orders
curl http://localhost:8081/api/v1/orders

# Update order status
curl -X PUT http://localhost:8081/api/v1/orders/{id}/status \
  -H "Content-Type: application/json" \
  -d '{"status": "shipped"}'
```

## Benefits Achieved

1. **Clean Architecture**: Clear separation between business logic and infrastructure
2. **Flexibility**: Easy to change external dependencies without affecting business logic
3. **Testability**: Each layer can be tested independently
4. **Maintainability**: Changes are localized to specific adapters
5. **Scalability**: Architecture supports adding new adapters and use cases
6. **Event-Driven**: Supports asynchronous processing and system integration

## Next Steps

1. **Database Integration**: Replace MemoryOrderRepository with a database adapter
2. **Authentication**: Add authentication middleware to the API adapter
3. **Monitoring**: Add metrics and logging adapters
4. **Testing**: Implement comprehensive unit and integration tests
5. **Validation**: Add input validation in the API adapter
6. **Error Handling**: Implement proper error handling and recovery mechanisms