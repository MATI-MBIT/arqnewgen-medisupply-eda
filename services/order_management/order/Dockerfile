# Build stage
FROM golang:1.24-alpine AS builder

# Install security updates
RUN apk update && apk upgrade && apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY src/ ./src/

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o order-management-service src/main.go

# Final image - use distroless for better security
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /app/order-management-service /order-management-service

# Use non-root user for security
USER nonroot:nonroot

# Set default environment variables for RabbitMQ and HTTP
ENV RABBITMQ_URL=amqp://guest:guest@localhost:5672/
ENV RABBITMQ_EXCHANGE=order-exchange
ENV RABBITMQ_QUEUE=order-queue
ENV RABBITMQ_ROUTING_KEY=order.created
ENV HTTP_PORT=8081

# Expose HTTP port for API service adapter
EXPOSE 8080

# Default command
ENTRYPOINT ["/order-management-service"]