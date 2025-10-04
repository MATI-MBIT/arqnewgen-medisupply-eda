# Order Management Service - Kubernetes Configuration

This directory contains the Kubernetes configuration for the Order Management microservice, which follows hexagonal architecture and uses RabbitMQ for messaging.

## Configuration Overview

### Service Configuration
- **Port**: 8081 (HTTP API)
- **Health Check**: `/health` endpoint
- **Architecture**: Hexagonal (Ports & Adapters)
- **Messaging**: RabbitMQ

### RabbitMQ Integration

The service connects to RabbitMQ with the following configuration:

```yaml
# Connection Details
Host: rabbitmq.mediorder.svc.cluster.local
Port: 5672
Username: user
Password: Retrieved from Kubernetes secret 'rabbitmq'
VHost: /

# Exchange Configuration
Exchange Name: order-exchange
Exchange Type: direct
Durable: true

# Queue Configuration
Queue Name: order-queue
Routing Key: order.created
Durable: true
```

### Environment Variables

The service uses the following environment variables:

| Variable | Description | Source |
|----------|-------------|---------|
| `RABBITMQ_URL` | Complete RabbitMQ connection string | Constructed from secret |
| `RABBITMQ_EXCHANGE` | Exchange name for publishing events | ConfigMap |
| `RABBITMQ_QUEUE` | Queue name for consuming events | ConfigMap |
| `RABBITMQ_ROUTING_KEY` | Routing key for message routing | ConfigMap |
| `HTTP_PORT` | HTTP server port | ConfigMap |
| `RABBITMQ_PASSWORD` | RabbitMQ password | Kubernetes Secret |

### Resource Allocation

```yaml
Resources:
  Limits:
    CPU: 200m
    Memory: 256Mi
  Requests:
    CPU: 100m
    Memory: 128Mi
```

### Health Checks

- **Liveness Probe**: Checks if the service is running
  - Path: `/health`
  - Initial Delay: 60s
  - Period: 30s

- **Readiness Probe**: Checks if the service is ready to receive traffic
  - Path: `/health`
  - Initial Delay: 30s
  - Period: 15s

### Dependencies

The service depends on:
1. **RabbitMQ**: Message broker for event-driven communication
   - Service: `rabbitmq.mediorder.svc.cluster.local`
   - Secret: `rabbitmq` (for password)

### Istio Service Mesh

The service is configured for Istio service mesh:
- Sidecar injection enabled
- Labels for traffic management
- Ready for advanced routing and security policies

### API Endpoints

The service exposes the following REST API endpoints:

- `GET /health` - Health check
- `POST /api/v1/orders` - Create new order
- `GET /api/v1/orders` - Get all orders
- `GET /api/v1/orders/{id}` - Get order by ID
- `PUT /api/v1/orders/{id}/status` - Update order status

### Event Publishing

The service publishes the following events to RabbitMQ:

- `order.created` - When a new order is created
- `order.updated` - When an order is updated

Event format:
```json
{
  "event_type": "order.created",
  "order_id": "uuid",
  "order": { ... },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Deployment

To deploy the service:

```bash
# Using Helm
helm install order-management ./order-service-chart \
  -f order-service-values.yaml \
  -n medisupply

# Or using kubectl with generated manifests
kubectl apply -f generated-manifests/ -n medisupply
```

## Monitoring

The service is configured for monitoring with:
- Prometheus metrics (if enabled)
- Health check endpoints
- Istio observability features

## Security

Security considerations:
- RabbitMQ password stored in Kubernetes secrets
- Istio service mesh for secure communication
- Network policies (configurable)
- Pod security contexts

## Scaling

The service supports horizontal scaling:
- Default: 1 replica
- Autoscaling: Disabled by default
- Can be enabled with CPU-based scaling (80% threshold)

## Future Enhancements

The configuration includes placeholders for:
- Database persistence
- Advanced monitoring
- Backup and recovery
- Network policies
- Custom ingress rules