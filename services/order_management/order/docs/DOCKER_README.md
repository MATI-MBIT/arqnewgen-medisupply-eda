# Order Management Service - Docker Configuration

This document explains how to build and run the Order Management Service using Docker.

## Docker Image Overview

The Order Management Service uses a multi-stage Docker build for optimal security and size:

### Build Stages

1. **Builder Stage** (`golang:1.24-alpine`)
   - Downloads dependencies
   - Compiles the Go application with security flags
   - Creates a statically linked binary

2. **Runtime Stage** (`gcr.io/distroless/static-debian12:nonroot`)
   - Minimal distroless image for security
   - No shell, package manager, or unnecessary tools
   - Runs as non-root user
   - Only contains the compiled binary and CA certificates

### Image Specifications

```dockerfile
# Base Images
Builder: golang:1.24-alpine
Runtime: gcr.io/distroless/static-debian12:nonroot

# Security Features
- Non-root user execution
- Statically linked binary
- Minimal attack surface
- No shell access
- Security-hardened build flags

# Environment Variables
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=order-exchange
RABBITMQ_QUEUE=order-queue
RABBITMQ_ROUTING_KEY=order.created
HTTP_PORT=8080

# Exposed Ports
8080 (HTTP API)
```

## Building the Image

### Manual Build
```bash
# Build with default tag
docker build -t order_management/order:latest .

# Build with specific tag
docker build -t order_management/order:v1.0.0 .

# Build with build arguments (if needed)
docker build --build-arg GO_VERSION=1.24 -t order_management/order:latest .
```

### Using Build Script
```bash
# Make script executable (Linux/Mac)
chmod +x build-docker.sh
```

## Running the Container

### Standalone Container
```bash
# Basic run
docker run -p 8080:8080 \
  -e RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/" \
  order_management/order:latest

# With all environment variables
docker run -p 8080:8080 \
  -e RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/" \
  -e RABBITMQ_EXCHANGE="order-exchange" \
  -e RABBITMQ_QUEUE="order-queue" \
  -e RABBITMQ_ROUTING_KEY="order.created" \
  -e HTTP_PORT="8080" \
  order_management/order:latest

# With custom network
docker network create order-network
docker run --network order-network -p 8080:8080 \
  -e RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/" \
  order_management/order:latest
```

### Using Docker Compose

The service includes a complete docker-compose setup with RabbitMQ:

```bash
# Start all services
cd deployment/
docker-compose up -d

# View logs
docker-compose logs -f order-management

# Stop services
docker-compose down

# Rebuild and restart
docker-compose up --build -d
```

#### Docker Compose Services

1. **rabbitmq**: RabbitMQ message broker with management UI
2. **rabbitmq-init**: Initializes RabbitMQ topology (exchanges, queues)
3. **order-management**: The order management service

#### Service Dependencies

```
order-management → rabbitmq-init → rabbitmq
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `RABBITMQ_URL` | Complete RabbitMQ connection string | `amqp://guest:guest@localhost:5672/` | Yes |
| `RABBITMQ_EXCHANGE` | Exchange name for publishing events | `order-exchange` | Yes |
| `RABBITMQ_QUEUE` | Queue name for consuming events | `order-queue` | Yes |
| `RABBITMQ_ROUTING_KEY` | Routing key for message routing | `order.created` | Yes |
| `HTTP_PORT` | HTTP server port | `8080` | Yes |

## Health Checks

### HTTP Health Check
```bash
# Check service health
curl http://localhost:8080/health

# Expected response
{
  "status": "healthy",
  "service": "order-management",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Container Health Check
```bash
# Check if container is running
docker ps --filter "name=order-management"

# Check container logs
docker logs order-management

# Execute health check manually
curl -f http://localhost:8080/health || exit 1
```

## Networking

### Port Mapping
- **8080**: HTTP API server
- **5672**: RabbitMQ AMQP (when using docker-compose)
- **15672**: RabbitMQ Management UI (when using docker-compose)

### Service Discovery
When using docker-compose, services can communicate using service names:
- RabbitMQ: `rabbitmq:5672`
- Order Management: `order-management:8080`

## Volumes and Persistence

The order management service is stateless and doesn't require persistent volumes. However, RabbitMQ in the docker-compose setup can be configured with persistence:

```yaml
# Add to rabbitmq service in docker-compose.yml
volumes:
  - rabbitmq_data:/var/lib/rabbitmq

volumes:
  rabbitmq_data:
```

## Security Considerations

### Image Security
- Uses distroless base image (minimal attack surface)
- Runs as non-root user (`nonroot:nonroot`)
- No shell or package manager in runtime image
- Statically linked binary (no dynamic dependencies)

### Runtime Security
```bash
# Run with read-only filesystem
docker run --read-only -p 8080:8080 \
  -e RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/" \
  order_management/order:latest

# Run with limited capabilities
docker run --cap-drop=ALL -p 8080:8080 \
  -e RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/" \
  order_management/order:latest

# Run with memory limits
docker run -m 256m --cpus="0.5" -p 8080:8080 \
  -e RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/" \
  order_management/order:latest
```

## Troubleshooting

### Common Issues

1. **Container Won't Start**
   ```bash
   # Check logs
   docker logs order-management
   
   # Common causes:
   # - Invalid RABBITMQ_URL
   # - RabbitMQ not accessible
   # - Port already in use
   ```

2. **Cannot Connect to RabbitMQ**
   ```bash
   # Test RabbitMQ connectivity
   docker run --rm --network container:rabbitmq \
     rabbitmq:management-alpine \
     rabbitmqctl status
   
   # Check network connectivity
   docker run --rm --network container:order-management \
     alpine/curl curl -f http://rabbitmq:15672
   ```

3. **Health Check Fails**
   ```bash
   # Test health endpoint
   curl -v http://localhost:8080/health
   
   # Check if service is listening
   docker exec order-management netstat -tlnp
   ```

### Debug Mode

To run with debug logging:
```bash
docker run -p 8080:8080 \
  -e RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/" \
  -e LOG_LEVEL="DEBUG" \
  order_management/order:latest
```

## Production Deployment

### Registry Push
```bash
# Tag for registry
docker tag order_management/order:latest your-registry.com/order_management/order:v1.0.0

# Push to registry
docker push your-registry.com/order_management/order:v1.0.0
```

### Kubernetes Deployment
The Docker image is designed to work with the Kubernetes configuration in `k8s/config/services/order_management/order/`.

### Resource Recommendations
```yaml
# Kubernetes resource limits
resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

## Monitoring

### Metrics Collection
The service exposes metrics that can be collected by Prometheus:
- HTTP request metrics
- RabbitMQ connection status
- Application-specific metrics

### Log Aggregation
Logs are written to stdout/stderr and can be collected by:
- Docker logging drivers
- Kubernetes log aggregation
- Centralized logging systems (ELK, Fluentd, etc.)

## Development Workflow

```bash
# 1. Make code changes
# 2. Build new image
./build-docker.sh

# 3. Test locally
docker-compose up --build -d

# 4. Run tests
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/v1/orders -d '{"customer_id":"test",...}'

# 5. Clean up
docker-compose down
```