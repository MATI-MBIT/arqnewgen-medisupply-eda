# Build stage
FROM golang:1.21-alpine AS builder

# Instalar actualizaciones de seguridad
RUN apk update && apk upgrade && apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Compilar la aplicación con flags de seguridad
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o order-event-client main.go

# Imagen final - usar distroless para mayor seguridad
FROM gcr.io/distroless/static-debian12:nonroot

# Copiar certificados CA desde builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copiar el binario compilado
COPY --from=builder /app/order-event-client /order-event-client

# Usar usuario no-root por seguridad
USER nonroot:nonroot

# Exponer puerto (opcional, para métricas futuras)
EXPOSE 8010

# Set default environment variables
ENV MQTT_BROKER=tcp://localhost:1883
ENV MQTT_CLIENT_ID=order-event-client
ENV MQTT_TOPIC=events/sensor
ENV HTTP_PORT=8110

# Comando por defecto
ENTRYPOINT ["/order-event-client"]
