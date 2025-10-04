# RabbitMQ Configuration for Order Management Service

## Overview

This document outlines the RabbitMQ configuration for the Order Management service, adapted from the Kafka-based warehouse batch service configuration.

## Key Differences: Kafka vs RabbitMQ

| Aspect | Warehouse (Kafka) | Order Management (RabbitMQ) |
|--------|-------------------|----------------------------|
| **Message Broker** | Kafka | RabbitMQ |
| **Connection String** | `KAFKA_BROKER_ADDRESS: "kafka-warehouse:9092"` | `RABBITMQ_URL: "amqp://user:password@rabbitmq.mediorder.svc.cluster.local:5672/"` |
| **Topic/Exchange** | `KAFKA_TOPIC: "warehouse-order-damage"` | `RABBITMQ_EXCHANGE: "order-exchange"` |
| **Consumer Group** | `KAFKA_GROUP_ID: "warehouse-batch-service"` | `RABBITMQ_QUEUE: "order-queue"` |
| **Routing** | Kafka partitions | RabbitMQ routing keys |
| **Port** | 9092 | 5672 |

## RabbitMQ Service Discovery

The service connects to RabbitMQ using Kubernetes service discovery:

```yaml
# Service FQDN
rabbitmq.mediorder.svc.cluster.local:5672

# This resolves to:
# - Namespace: mediorder
# - Service: rabbitmq  
# - Port: 5672 (AMQP)
```

## Environment Variables Configuration

### Primary Configuration
```yaml
app:
  config:
    RABBITMQ_URL: "amqp://user:password@rabbitmq.mediorder.svc.cluster.local:5672/"
    RABBITMQ_EXCHANGE: "order-exchange"
    RABBITMQ_QUEUE: "order-queue"
    RABBITMQ_ROUTING_KEY: "order.created"
    HTTP_PORT: "8081"
```

### Secret-based Password Injection
```yaml
env:
  custom:
    - name: RABBITMQ_PASSWORD
      valueFrom:
        secretKeyRef:
          name: rabbitmq
          key: rabbitmq-password
    - name: RABBITMQ_URL
      value: "amqp://user:$(RABBITMQ_PASSWORD)@rabbitmq.mediorder.svc.cluster.local:5672/"
```

## RabbitMQ Topology

### Exchange Configuration
```yaml
rabbitmq:
  exchange:
    name: "order-exchange"
    type: "direct"
    durable: true
```

### Queue Configuration
```yaml
rabbitmq:
  queue:
    name: "order-queue"
    durable: true
    routingKey: "order.created"
```

### Message Flow
```
Order Service → order-exchange → order-queue → Order Service Consumer
                    ↓
              (routing key: order.created)
```

## Connection Details

### Authentication
- **Username**: `user` (default RabbitMQ user)
- **Password**: Retrieved from Kubernetes secret `rabbitmq`
- **VHost**: `/` (default virtual host)

### Connection String Format
```
amqp://[username]:[password]@[host]:[port]/[vhost]
```

### Example
```
amqp://user:secretpassword@rabbitmq.mediorder.svc.cluster.local:5672/
```

## Security Considerations

### Secret Management
```yaml
# RabbitMQ password is stored in Kubernetes secret
secretName: "rabbitmq"
secretKey: "rabbitmq-password"

# Referenced in deployment as environment variable
- name: RABBITMQ_PASSWORD
  valueFrom:
    secretKeyRef:
      name: rabbitmq
      key: rabbitmq-password
```

### Network Security
- Service-to-service communication within Kubernetes cluster
- Istio service mesh for additional security layers
- Network policies (configurable)

## Integration with Kafka-RabbitMQ Replicator

The service integrates with the existing Kafka-RabbitMQ replicator:

```yaml
# From kafka-rabbitmq-replicator values.yaml
replication:
  kafkaToRabbitmq:
    enabled: true
    mappings:
      - kafkaTopic: "events-order-damage"
        rabbitmqQueue: "order-damage-queue"
        rabbitmqExchange: "events"
        rabbitmqRoutingKey: "order.damage"
```

This allows the order service to:
1. Receive events from Kafka topics via RabbitMQ
2. Publish events to RabbitMQ that can be replicated to Kafka

## Event Patterns

### Publishing Events
```json
{
  "event_type": "order.created",
  "order_id": "uuid",
  "order": {
    "id": "uuid",
    "customer_id": "customer-123",
    "status": "created",
    ...
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Routing Keys
- `order.created` - New order events
- `order.updated` - Order status changes
- `order.cancelled` - Order cancellations

## Monitoring and Observability

### Health Checks
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8081

readinessProbe:
  httpGet:
    path: /health
    port: 8081
```

### Metrics (Future)
- RabbitMQ connection status
- Message publish/consume rates
- Queue depth monitoring
- Error rates

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check RabbitMQ service is running in `mediorder` namespace
   - Verify service name resolution
   - Check network policies

2. **Authentication Failed**
   - Verify secret `rabbitmq` exists
   - Check username/password combination
   - Ensure secret is in correct namespace

3. **Queue/Exchange Not Found**
   - Verify RabbitMQ topology is created
   - Check exchange and queue names
   - Ensure proper permissions

### Debug Commands
```bash
# Check RabbitMQ service
kubectl get svc rabbitmq -n mediorder

# Check RabbitMQ secret
kubectl get secret rabbitmq -n mediorder -o yaml

# Check order service logs
kubectl logs -f deployment/order-management -n medisupply

# Test RabbitMQ connectivity
kubectl run rabbitmq-test --rm -it --image=rabbitmq:management \
  -- rabbitmqctl -n rabbit@rabbitmq.mediorder.svc.cluster.local list_queues
```

## Migration Notes

When migrating from Kafka to RabbitMQ:

1. **Message Format**: Maintain JSON structure for compatibility
2. **Error Handling**: Implement proper acknowledgment patterns
3. **Ordering**: Consider message ordering requirements
4. **Durability**: Ensure queues and exchanges are durable
5. **Monitoring**: Update monitoring dashboards for RabbitMQ metrics