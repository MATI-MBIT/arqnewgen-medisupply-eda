# Usar imagen base más reciente y segura
FROM golang:1.22-alpine3.19 AS builder

# Instalar actualizaciones de seguridad
RUN apk update && apk upgrade && apk add --no-cache git ca-certificates

WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download && go mod verify

# Copiar código fuente
COPY . .

# Compilar la aplicación con flags de seguridad
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o event-generator main.go

# Imagen final - usar distroless para mayor seguridad
FROM gcr.io/distroless/static-debian12:nonroot

# Copiar certificados CA desde builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copiar el binario compilado
COPY --from=builder /app/event-generator /event-generator

# Usar usuario no-root por seguridad
USER nonroot:nonroot

# Exponer puerto (opcional, para métricas futuras)
EXPOSE 8080

# Comando por defecto
ENTRYPOINT ["/event-generator"]