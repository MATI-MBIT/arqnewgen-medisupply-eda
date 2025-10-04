# Local Testing Guide for Order Management Service

This guide shows you how to test the order damage event handling locally using Docker Compose.

## Prerequisites

- Docker and Docker Compose installed
- Python 3 with `pika` library (for Python test script)
- `jq` and `curl` (for bash test script)

## Setup

### 1. Start the Services

```bash
# Navigate to the deployment directory
cd services/order_management/order/deployment

# Start RabbitMQ and the order service
docker-compose up -d

# Check if services are running
docker-compose ps
```

### 2. Verify Setup

```bash
# Check RabbitMQ is healthy
docker logs rabbitmq

# Check order service is running
docker logs order-management

# Access RabbitMQ Management UI
# Open http://localhost:15672 (guest/guest)
```

## Testing Methods

### Method 1: Python Test Script (Recommended)

```bash
# Install required Python package
pip install pika

# Send a minor damage event
python test/send_damage_event.py --severity minor

# Send a major damage event with specific order ID
python test/send_damage_event.py --severity major --order-id ORDER123

# Send multiple critical damage events
python test/send_damage_event.py --severity critical --count 3

# Available options:
# --severity: minor, major, critical (default: minor)
# --order-id: Custom order ID (auto-generated if not provided)
# --count: Number of events to send (default: 1)
```

### Method 2: Bash Test Script

```bash
# Send a minor damage event
./test/send_damage_event.sh minor

# Send a major damage event with specific order ID
./test/send_damage_event.sh major ORDER456

# Send a critical damage event
./test/send_damage_event.sh critical ORDER789
```

### Method 3: Manual RabbitMQ Management UI

1. Open http://localhost:15672 (guest/guest)
2. Go to "Queues" tab
3. Click on "order-damage-queue"
4. Scroll down to "Publish a message"
5. Set routing key to: `order.damage`
6. Use this payload:

```json
{
  "mqtt_topic": "events/order-damage",
  "payload": "{\"eventId\":\"evt_manual_test\",\"type\":\"order.damage\",\"source\":\"manual-test\",\"occurredAt\":\"2025-10-04T10:30:00.000Z\",\"orderId\":\"MANUAL_ORDER_123\",\"severity\":\"minor\",\"description\":\"Manual test damage event\",\"details\":{\"temperature\":15.5,\"humidity\":45,\"status\":\"active\",\"mqttTopic\":\"events/sensor\"}}",
  "timestamp": 1728036600.123
}
```

### Method 4: Direct Go Testing

```bash
# Test the parsing logic without RabbitMQ
go run examples/order_damage_example.go
```

## Monitoring and Verification

### Check Order Service Logs

```bash
# Follow logs in real-time
docker logs order-management -f

# Check recent logs
docker logs order-management --tail 50
```

### Expected Log Output

When a damage event is processed, you should see:

```
Processing order damage event: EventID=evt_xxx, OrderID=xxx, Severity=minor, OccurredAt=2025-10-04 10:30:00
Damage details: Temperature=15.50°C, Humidity=45%, Status=active
Damage description: Manual test damage event
Order ORDER_123 not found, creating new order from damage event
Created new order from damage event: ID=ORDER_123
Order ORDER_123 status updated to: damage_detected_minor
```

### Check RabbitMQ Queues

1. Open http://localhost:15672
2. Go to "Queues" tab
3. Check message counts in:
   - `order-damage-queue` (should decrease as messages are consumed)
   - Check "Message rates" for activity

### Health Check

```bash
# Check if the service is healthy
curl http://localhost:8080/health

# Expected response: HTTP 200 OK
```

## Test Scenarios

### Scenario 1: New Order with Minor Damage
```bash
python test/send_damage_event.py --severity minor --order-id NEW_ORDER_001
```
**Expected**: New order created with status `damage_detected_minor`

### Scenario 2: Existing Order with Major Damage
```bash
# First create an order, then send major damage
python test/send_damage_event.py --severity minor --order-id EXISTING_ORDER_002
python test/send_damage_event.py --severity major --order-id EXISTING_ORDER_002
```
**Expected**: Order status updated to `damage_detected_major`

### Scenario 3: Critical Damage (Auto-cancel)
```bash
python test/send_damage_event.py --severity critical --order-id CRITICAL_ORDER_003
```
**Expected**: Order status set to `cancelled_damage`

### Scenario 4: Bulk Testing
```bash
# Send multiple events with different severities
python test/send_damage_event.py --severity minor --count 5
python test/send_damage_event.py --severity major --count 3
python test/send_damage_event.py --severity critical --count 2
```

## Troubleshooting

### Service Won't Start
```bash
# Check Docker logs
docker-compose logs

# Rebuild the image
docker-compose build --no-cache order-management
docker-compose up -d
```

### RabbitMQ Connection Issues
```bash
# Check RabbitMQ is ready
docker exec rabbitmq rabbitmq-diagnostics ping

# Check network connectivity
docker exec order-management nslookup rabbitmq
```

### No Messages Being Processed
```bash
# Check queue bindings
curl -u guest:guest http://localhost:15672/api/bindings

# Check exchange exists
curl -u guest:guest http://localhost:15672/api/exchanges/%2F/events
```

## Cleanup

```bash
# Stop services
docker-compose down

# Remove volumes (if needed)
docker-compose down -v

# Remove images (if needed)
docker-compose down --rmi all
```

## Integration with Full System

To test with the complete MQTT → Kafka → RabbitMQ flow:

```bash
# Use the full Kubernetes deployment
cd ../../../../k8s
make init
make deploy

# Then use the MQTT Event Generator or send MQTT messages directly
```

This local setup is perfect for development and testing the order damage event handling logic before deploying to the full Kubernetes environment.