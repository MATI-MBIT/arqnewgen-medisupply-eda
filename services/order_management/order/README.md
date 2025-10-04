# Order Management Service

A microservice for managing orders built with hexagonal architecture (ports and adapters pattern) using Go and RabbitMQ.

## Architecture

This service follows the hexagonal architecture pattern with clear separation of concerns:

```
src/
├── main.go                           # Application entry point
├── config/                           # Configuration management
│   └── config.go
├── domain/                           # Business domain (core)
│   └── order.go                      # Order entities and interfaces
├── application/                      # Business logic (use cases)
│   └── order_service.go              # Order business operations
└── infrastructure/                   # External adapters
    ├── driving-adapters/             # Input adapters (primary)
    │   ├── api_service_adapter.go    # HTTP REST API
    │   └── order_consumer_adapter.go # RabbitMQ consumer
    └── driven-adapters/              # Output adapters (secondary)
        ├── memory_order_repository.go # In-memory data storage
        └── rabbitmq_publisher.go     # RabbitMQ event publisher
```

### Hexagonal Architecture Layers

1. **Domain Layer** (`domain/`): Contains business entities and interfaces (ports)
2. **Application Layer** (`application/`): Contains business logic and use cases
3. **Infrastructure Layer** (`infrastructure/`): Contains adapters that implement the ports
   - **Driving Adapters**: Handle incoming requests (HTTP API, message consumers)
   - **Driven Adapters**: Handle outgoing operations (database, message publishers)

## Features

- **RESTful API** for order management (create, read, update orders)
- **Event-driven architecture** with RabbitMQ integration
- **Hexagonal architecture** for clean separation of concerns
- **In-memory storage** for development and testing
- **Health check endpoint** for monitoring
- **Graceful shutdown** handling

## API Endpoints

### Health Check
- `GET /health` - Service health status

### Order Management
- `POST /api/v1/orders` - Create a new order
- `GET /api/v1/orders` - Get all orders
- `GET /api/v1/orders/{id}` - Get order by ID
- `PUT /api/v1/orders/{id}/status` - Update order status

### Example Requests

#### Create Order
```bash
curl -X POST http://localhost:8081/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer-123",
    "product_id": "product-456",
    "quantity": 2,
    "total_amount": 99.99
  }'
```

#### Get All Orders
```bash
curl http://localhost:8081/api/v1/orders
```

#### Update Order Status
```bash
curl -X PUT http://localhost:8081/api/v1/orders/{order-id}/status \
  -H "Content-Type: application/json" \
  -d '{"status": "shipped"}'
```

## Configuration

The service uses environment variables for configuration. Copy `.env.example` to `.env` and adjust as needed:

```bash
# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=order-exchange
RABBITMQ_QUEUE=order-queue
RABBITMQ_ROUTING_KEY=order.created

# HTTP Server Configuration
HTTP_PORT=8081
```

## Running the Service

### Prerequisites
- Go 1.24.1 or later
- RabbitMQ server running on localhost:5672

### Local Development
```bash
# Install dependencies
go mod tidy

# Run the service
go run src/main.go
```

### Using Docker
```bash
# Build the image
docker build -t order-management .

# Run with docker-compose
docker-compose up
```

## Event Publishing

The service publishes order events to RabbitMQ when:
- A new order is created (`order.created`)
- An order status is updated (`order.updated`)

Event format:
```json
{
  "event_type": "order.created",
  "order_id": "uuid",
  "order": {
    "id": "uuid",
    "customer_id": "customer-123",
    "product_id": "product-456",
    "quantity": 2,
    "status": "created",
    "total_amount": 99.99,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Development

### Adding New Features

1. **Domain Changes**: Add new entities or interfaces in `domain/`
2. **Business Logic**: Implement use cases in `application/`
3. **External Integration**: Create new adapters in `infrastructure/`

### Testing

The hexagonal architecture makes testing easier by allowing you to:
- Test business logic in isolation (application layer)
- Mock external dependencies using interfaces (ports)
- Test adapters independently

## Dependencies

- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/rabbitmq/amqp091-go` - RabbitMQ client
- `github.com/google/uuid` - UUID generation
- `github.com/joho/godotenv` - Environment variable loading