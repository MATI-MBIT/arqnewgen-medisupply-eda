# MediSupply EDA - Servicios

Microservicios desarrollados en Go para la arquitectura Event-Driven de MediSupply. Estos servicios manejan la generación, procesamiento y distribución de eventos IoT y de pedidos.

## 🏗️ Arquitectura de Servicios

```
mqtt-event-generator → EMQX → mqtt-order-event-client → EMQX → mqtt-kafka-bridge → Kafka
```

## 📦 Servicios Disponibles

### 1. mqtt-event-generator
**Generador de eventos IoT simulados**

- **Propósito**: Simula sensores IoT generando eventos cada 30 segundos
- **Tecnología**: Go + MQTT
- **Output**: Eventos JSON a topic MQTT `events/sensor`
- **Health Check**: HTTP endpoint en `/health`

### 2. mqtt-order-event-client
**Cliente procesador de eventos de pedidos**

- **Propósito**: Suscribe a eventos MQTT y los expone via REST API
- **Tecnología**: Go + MQTT + HTTP REST
- **Input**: Eventos desde topic MQTT `events/sensor`
- **Output**: API REST con estadísticas y eventos almacenados

## 🚀 Desarrollo Local

### Prerrequisitos

- Go 1.21+
- Docker
- EMQX broker (local o remoto)

### Configuración

Cada servicio incluye un archivo `.env.example` con las variables necesarias:

```bash
# Copiar configuración de ejemplo
cp mqtt-event-generator/.env.example mqtt-event-generator/.env
cp mqtt-order-event-client/.env.example mqtt-order-event-client/.env

# Editar configuraciones según tu entorno
```

### Ejecución Local

```bash
# Ejecutar generador de eventos
cd mqtt-event-generator
go mod tidy
go run main.go

# En otra terminal, ejecutar cliente
cd mqtt-order-event-client
go mod tidy
go run main.go
```

## 🐳 Docker

### Construcción de Imágenes

```bash
# Construir todas las imágenes
make build-all

# Construir imagen específica
make build SERVICE=mqtt-event-generator

# Ver imágenes construidas
make list-images
```

### Para Desarrollo con Kubernetes Local

```bash
# Construir y cargar en Kind/Minikube
make build-load-all

# O servicio específico
make build-load SERVICE=mqtt-event-generator
```

## ⚙️ Configuración

### Variables de Entorno Comunes

| Variable | Descripción | Valor por defecto |
|----------|-------------|-------------------|
| `MQTT_BROKER` | URL del broker MQTT | `tcp://localhost:1883` |
| `MQTT_USERNAME` | Usuario MQTT | (vacío) |
| `MQTT_PASSWORD` | Contraseña MQTT | (vacío) |
| `HTTP_PORT` | Puerto HTTP para APIs | `8080` |

### Configuración Específica por Servicio

#### mqtt-event-generator
```bash
MQTT_CLIENT_ID=event-generator
MQTT_TOPIC=events/sensor
EVENT_INTERVAL_SECONDS=30
```

#### mqtt-order-event-client
```bash
MQTT_CLIENT_ID=order-event-client
MQTT_TOPIC=events/sensor
```

## 🔧 Comandos Make Disponibles

```bash
make help                    # Mostrar ayuda completa
make build SERVICE=<name>    # Construir imagen específica
make build-all              # Construir todas las imágenes
make push SERVICE=<name>     # Publicar imagen al registry
make push-all               # Publicar todas las imágenes
make build-push-all         # Construir y publicar todo
make load-to-k8s SERVICE=<name> # Cargar imagen a K8s local
make load-all-to-k8s        # Cargar todas a K8s local
make build-load SERVICE=<name>  # Construir y cargar específica
make build-load-all         # Construir y cargar todas
make clean-images           # Limpiar imágenes locales
```

## 📊 APIs y Endpoints

### mqtt-event-generator

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/health` | GET | Health check del servicio |

**Respuesta Health Check:**
```json
{
  "status": "healthy",
  "timestamp": "2023-12-21T10:30:45Z",
  "service": "mqtt-event-generator"
}
```

### mqtt-order-event-client

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/health` | GET | Health check del servicio |
| `/events` | GET | Obtener todos los eventos almacenados |
| `/events/latest` | GET | Obtener el evento más reciente |
| `/events/count` | GET | Obtener contador de eventos |
| `/events/stats` | GET | Obtener estadísticas calculadas |

**Ejemplo de evento:**
```json
{
  "id": "evt_1703123456",
  "timestamp": "2023-12-21T10:30:45Z",
  "type": "sensor_reading",
  "source": "temperature_sensor_01",
  "data": {
    "temperature": 23.5,
    "humidity": 45.2,
    "status": "active"
  }
}
```

## 🧪 Testing

### Testing Local con EMQX

```bash
# Ejecutar EMQX en Docker
docker run -d --name emqx -p 1883:1883 -p 8083:8083 -p 8084:8084 -p 8883:8883 -p 18083:18083 emqx/emqx:latest

# Configurar servicios para usar EMQX local
export MQTT_BROKER="tcp://localhost:1883"
export MQTT_USERNAME="admin"
export MQTT_PASSWORD="public"
```

### Verificar Flujo de Eventos

```bash
# 1. Iniciar generador
cd mqtt-event-generator && go run main.go &

# 2. Iniciar cliente
cd mqtt-order-event-client && go run main.go &

# 3. Verificar eventos
curl http://localhost:8080/events/latest
curl http://localhost:8080/events/stats
```

## 🔍 Monitoreo y Logs

### Logs de Desarrollo

Los servicios incluyen logging detallado:

```bash
# Ver logs del generador
go run main.go 2>&1 | grep "Event published"

# Ver logs del cliente
go run main.go 2>&1 | grep "Event received"
```

### Métricas

- **Eventos generados por segundo**
- **Eventos procesados por segundo**
- **Estadísticas de temperatura/humedad**
- **Estado de conexiones MQTT**

## 🚨 Troubleshooting

### Problemas Comunes

1. **Error de conexión MQTT**:
   ```bash
   # Verificar que EMQX esté ejecutándose
   docker ps | grep emqx
   
   # Verificar conectividad
   telnet localhost 1883
   ```

2. **Dependencias Go**:
   ```bash
   go mod tidy
   go mod download
   ```

3. **Puerto ocupado**:
   ```bash
   # Cambiar puerto HTTP
   export HTTP_PORT=8081
   ```

### Debugging

```bash
# Habilitar logs detallados
export LOG_LEVEL=debug

# Verificar variables de entorno
env | grep MQTT
```

## 🔄 Integración con Kubernetes

Los servicios se despliegan automáticamente en Kubernetes usando los charts en `k8s/`:

```bash
# Desde el directorio k8s
make deploy

# Verificar despliegue
kubectl get pods -l app=mqtt-event-generator
kubectl get pods -l app=mqtt-order-event-client
```

## 🤝 Desarrollo

### Estructura de Código

```
services/
├── mqtt-event-generator/
│   ├── main.go              # Aplicación principal
│   ├── Dockerfile           # Imagen Docker
│   ├── go.mod              # Dependencias Go
│   ├── .env.example        # Configuración ejemplo
│   └── README.md           # Documentación específica
├── mqtt-order-event-client/
│   ├── main.go              # Aplicación principal
│   ├── publisher/           # Módulos adicionales
│   ├── Dockerfile           # Imagen Docker
│   ├── go.mod              # Dependencias Go
│   ├── .env.example        # Configuración ejemplo
│   └── README.md           # Documentación específica
├── Makefile                # Comandos de build/deploy
└── README.md               # Esta documentación
```

### Agregar Nuevo Servicio

1. Crear directorio del servicio
2. Implementar aplicación Go
3. Crear Dockerfile
4. Agregar al Makefile en variable `SERVICES`
5. Crear chart de Helm en `k8s/`
6. Documentar en README específico

## 📄 Licencia

MIT License - Ver archivo LICENSE en el directorio raíz.