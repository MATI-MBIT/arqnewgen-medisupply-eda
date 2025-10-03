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
    -o warehouse-batch-service src/main.go

# Final image - use distroless for better security
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /app/warehouse-batch-service /warehouse-batch-service

# Use non-root user for security
USER nonroot:nonroot

# Set default environment variables
ENV KAFKA_TOPIC=warehouse-events
ENV KAFKA_BROKER_ADDRESS=kafka:9092
ENV HTTP_PORT=8080

# Expose HTTP port for API service adapter
EXPOSE 8080

# Default command
ENTRYPOINT ["/warehouse-batch-service"]