# Queue Separation Update

This document describes the changes made to separate input and output queues for the order management service.

## Overview

The order management service now uses two separate RabbitMQ queues:
- **Input Queue**: `order-damage-queue` - receives damage events
- **Output Queue**: `order-events-queue` - publishes order lifecycle events

## Changes Made

### 1. Configuration Updates (`src/config/config.go`)

**Before:**
```go
type RabbitMQConfig struct {
    URL          string
    ExchangeName string
    QueueName    string
    RoutingKey   string
}
```

**After:**
```go
type RabbitMQConfig struct {
    URL          string
    ExchangeName string
    // Consumer configuration (for receiving damage events)
    ConsumerQueueName    string
    ConsumerRoutingKey   string
    // Publisher configuration (for publishing order events)
    PublisherQueueName   string
    PublisherRoutingKey  string
}
```

### 2. Environment Variables

**New Environment Variables:**
- `RABBITMQ_CONSUMER_QUEUE` - Queue for receiving damage events (default: "order-damage-queue")
- `RABBITMQ_CONSUMER_ROUTING_KEY` - Routing key for damage events (default: "order.damage")
- `RABBITMQ_PUBLISHER_QUEUE` - Queue for publishing order events (default: "order-events-queue")
- `RABBITMQ_PUBLISHER_ROUTING_KEY` - Routing key for order events (default: "order.events")

**Deprecated Environment Variables:**
- `RABBITMQ_QUEUE` - Replaced by separate consumer/publisher queues
- `RABBITMQ_ROUTING_KEY` - Replaced by separate consumer/publisher routing keys

### 3. RabbitMQ Publisher Updates (`src/infrastructure/driven-adapters/rabbitmq_publisher.go`)

**Changes:**
- Added queue name parameter to constructor
- Publisher now creates and binds its own queue
- Enhanced logging for better debugging

**New Constructor:**
```go
func NewRabbitMQPublisher(rabbitMQURL, exchangeName, queueName, routingKey string) (*RabbitMQPublisher, error)
```

### 4. Docker Compose Updates (`deployment/docker-compose.yml`)

**Queue Creation:**
```bash
# Create exchange and queues
rabbitmqadmin declare exchange name=events type=direct durable=true
rabbitmqadmin declare queue name=order-damage-queue durable=true
rabbitmqadmin declare queue name=order-events-queue durable=true

# Create bindings
rabbitmqadmin declare binding source=events destination=order-damage-queue routing_key=order.damage
rabbitmqadmin declare binding source=events destination=order-events-queue routing_key=order.events
```

**Environment Variables:**
```yaml
environment:
  - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
  - RABBITMQ_EXCHANGE=events
  # Consumer configuration
  - RABBITMQ_CONSUMER_QUEUE=order-damage-queue
  - RABBITMQ_CONSUMER_ROUTING_KEY=order.damage
  # Publisher configuration
  - RABBITMQ_PUBLISHER_QUEUE=order-events-queue
  - RABBITMQ_PUBLISHER_ROUTING_KEY=order.events
  - HTTP_PORT=8080
```

### 5. Kubernetes Values Updates (`k8s/config/services/order_management/order/order-service-values.yaml`)

**Updated Configuration:**
```yaml
app:
  config:
    RABBITMQ_URL: "amqp://user:password@rabbitmq.mediorder.svc.cluster.local:5672/"
    RABBITMQ_EXCHANGE: "events"
    # Consumer configuration
    RABBITMQ_CONSUMER_QUEUE: "order-damage-queue"
    RABBITMQ_CONSUMER_ROUTING_KEY: "order.damage"
    # Publisher configuration
    RABBITMQ_PUBLISHER_QUEUE: "order-events-queue"
    RABBITMQ_PUBLISHER_ROUTING_KEY: "order.events"
    HTTP_PORT: "8080"
```

### 6. New Testing Scripts

**Complete Flow Test (`test/test_complete_flow.py`):**
- Sends damage events to input queue
- Monitors order events from output queue
- Verifies end-to-end functionality
- Tests multiple scenarios

**Order Events Monitor (`test/monitor_order_events.py`):**
- Real-time monitoring of order events
- Queue status checking
- Event analysis and display

## Message Flow

```
Damage Event → order-damage-queue → Order Service → order-events-queue → Order Event
```

### Input Flow (Damage Events)
1. MQTT sensors detect damage
2. Damage events sent to `order-damage-queue` with routing key `order.damage`
3. Order service consumes and processes damage events

### Output Flow (Order Events)
1. Order service processes damage events
2. Creates/updates orders in database
3. Publishes order lifecycle events to `order-events-queue` with routing key `order.events`
4. Other services can consume order events for further processing

## Event Types Published

The order service now publishes these event types to `order-events-queue`:

| Event Type | Description | Trigger |
|------------|-------------|---------|
| `order.created_from_damage` | New order created from damage event | Order doesn't exist when damage detected |
| `order.damage_processed` | Order updated after damage processing | Order status updated due to damage |
| `order.created` | Regular order creation | Manual order creation via API |
| `order.updated` | Order status updated | Manual status updates |

## Testing

### Quick Test
```bash
# Start services
cd services/order_management/order/deployment
docker-compose up -d

# Run complete flow test
python ../test/test_complete_flow.py
```

### Monitor Events
```bash
# In one terminal - monitor order events
python test/monitor_order_events.py

# In another terminal - send damage events
python test/send_damage_event.py --severity minor --order-id TEST123
```

## Benefits

✅ **Separation of Concerns**: Input and output are clearly separated  
✅ **Scalability**: Different queues can be scaled independently  
✅ **Monitoring**: Easier to monitor input vs output message rates  
✅ **Integration**: Other services can easily consume order events  
✅ **Debugging**: Clear visibility into message flow  
✅ **Configuration**: Flexible queue and routing key configuration  

## Migration Notes

**For Existing Deployments:**
1. Update environment variables to use new consumer/publisher configuration
2. Ensure both queues are created and properly bound
3. Update any monitoring or alerting to check both queues
4. Test the complete flow before deploying to production

**Backward Compatibility:**
- Old environment variables still work with default values
- Existing functionality remains unchanged
- Only new queue separation functionality is added