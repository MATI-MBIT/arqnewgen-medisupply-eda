# MQTT Order Event Client Chart - MediSupply EDA

Chart de Helm para desplegar el cliente procesador de eventos de pedidos en la arquitectura Event-Driven de MediSupply. Este servicio actúa como intermediario entre el generador de eventos y el puente hacia Kafka.

## 🎯 Propósito en MediSupply EDA

El MQTT Order Event Client procesa eventos en el flujo de la arquitectura:

```
mqtt-event-generator → EMQX (events/sensor) → mqtt-order-event-client → EMQX (orders/events) → mqtt-kafka-bridge → Kafka
```

### Funciones Principales

- **Suscripción MQTT**: Recibe eventos de sensores desde `events/sensor`
- **Procesamiento**: Almacena eventos en memoria y calcula estadísticas
- **API REST**: Expone endpoints HTTP para consultar eventos y métricas
- **Republishing**: Procesa y reenvía eventos a `orders/events` (configurable)

## 🚀 Instalación

### Instalación Estándar MediSupply

```bash
# Desde el directorio k8s
helm install mqtt-order-event-client ./mqtt-order-event-client \
  --namespace medilogistic \
  --create-namespace

# O usando el Makefile
make deploy  # Incluye mqtt-order-event-client en el despliegue completo
```

### Verificación

```bash
# Verificar pods
kubectl get pods -l app.kubernetes.io/name=mqtt-order-event-client -n medilogistic

# Ver logs en tiempo real
kubectl logs -f deployment/mqtt-order-event-client -n medilogistic

# Verificar API REST
kubectl port-forward svc/mqtt-order-event-client 8080:8080 -n medilogistic
curl http://localhost:8080/health
```

## Configuración

### Valores principales

| Parámetro | Descripción | Valor por defecto |
|-----------|-------------|-------------------|
| `replicaCount` | Número de réplicas | `1` |
| `image.repository` | Repositorio de la imagen | `mqtt-order-event-client` |
| `image.tag` | Tag de la imagen | `latest` |
| `mqtt.broker` | URL del broker MQTT | `tcp://emqx.medisupply.svc.cluster.local:1883` |
| `mqtt.clientId` | ID del cliente MQTT | `order-event-client` |
| `mqtt.topic` | Topic a suscribir | `events/sensor` |
| `mqtt.username` | Usuario MQTT | `admin` |
| `mqtt.password` | Contraseña MQTT | `public` |
| `service.port` | Puerto HTTP expuesto por el servicio | `8080` |

### Ejemplo de configuración personalizada

```yaml
# custom-values.yaml
replicaCount: 2

mqtt:
  broker: "tcp://my-mqtt-broker:1883"
  topic: "sensors/temperature"
  username: "myuser"
  password: "mypassword"

service:
  port: 9090

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

```bash
helm upgrade --install mqtt-order-event-client ./mqtt-order-event-client \
  --namespace medisupply \
  --values custom-values.yaml
```

## Monitoreo

### Ver logs
```bash
kubectl logs -f deployment/mqtt-order-event-client -n medisupply
```

### Ver estado
```bash
kubectl get pods -l app.kubernetes.io/name=mqtt-order-event-client -n medisupply
```

### Ver API
```bash
kubectl port-forward svc/mqtt-order-event-client 8080:8080 -n medisupply
curl http://localhost:8080/health
curl http://localhost:8080/events
```

## Integración con Istio

El chart incluye automáticamente las anotaciones necesarias para la inyección del sidecar de Istio:

```yaml
podAnnotations:
  sidecar.istio.io/inject: "true"
```

## Autoscaling

Para habilitar el autoscaling horizontal:

```yaml
autoscaling:
  enabled: true
  minReplicas: 1
  maxReplicas: 5
  targetCPUUtilizationPercentage: 80
```

## 📊 API REST Endpoints

### Endpoints Disponibles

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/health` | GET | Health check del servicio |
| `/events` | GET | Obtener todos los eventos almacenados |
| `/events/latest` | GET | Obtener el evento más reciente |
| `/events/count` | GET | Obtener contador de eventos |
| `/events/stats` | GET | Obtener estadísticas calculadas |

### Ejemplos de Uso

```bash
# Health check
curl http://localhost:8080/health

# Obtener todos los eventos
curl http://localhost:8080/events

# Obtener estadísticas
curl http://localhost:8080/events/stats
```

### Respuesta de Estadísticas

```json
{
  "total_events": 150,
  "average_temperature": 24.3,
  "average_humidity": 58.7,
  "active_sensors": 3,
  "latest_event": {
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
}
```

## 🔧 Integración con MediSupply

### Dependencias

- **EMQX**: Debe estar desplegado en namespace `medilogistic`
- **mqtt-event-generator**: Generador de eventos fuente
- **mqtt-kafka-bridge**: Consumidor de eventos procesados
- **Istio**: Inyección de sidecar habilitada (opcional)

### Flujo de Datos

1. **Suscribe** a eventos desde `mqtt-event-generator` en topic `events/sensor`
2. **Almacena** eventos en memoria con límite configurable
3. **Calcula** estadísticas en tiempo real (temperatura, humedad promedio)
4. **Expone** datos via API REST para monitoreo
5. **Republica** eventos procesados a `orders/events` (opcional)

### Topics MQTT

| Topic | Dirección | Propósito |
|-------|-----------|-----------|
| `events/sensor` | Subscribe | Recibir eventos de sensores |
| `orders/events` | Publish | Enviar eventos procesados |

## 🚨 Troubleshooting

### Problemas Comunes

1. **No recibe eventos**:

   ```bash
   # Verificar logs
   kubectl logs deployment/mqtt-order-event-client -n medilogistic
   
   # Verificar que mqtt-event-generator esté publicando
   kubectl logs deployment/mqtt-event-generator -n medilogistic
   ```

2. **API no responde**:

   ```bash
   # Verificar estado del pod
   kubectl get pods -l app.kubernetes.io/name=mqtt-order-event-client -n medilogistic
   
   # Verificar servicio
   kubectl get svc mqtt-order-event-client -n medilogistic
   ```

3. **Conexión MQTT falla**:

   ```bash
   # Verificar conectividad a EMQX
   kubectl exec -it deployment/mqtt-order-event-client -n medilogistic -- nc -zv emqx 1883
   ```

### Verificar Funcionamiento

```bash
# Ver eventos recibidos en tiempo real
kubectl port-forward svc/mqtt-order-event-client 8080:8080 -n medilogistic
watch -n 5 'curl -s http://localhost:8080/events/count'

# Verificar estadísticas
curl -s http://localhost:8080/events/stats | jq '.'
```
