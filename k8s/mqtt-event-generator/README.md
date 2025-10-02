# MQTT Event Generator Chart - MediSupply EDA

Chart de Helm para desplegar el generador de eventos IoT simulados en la arquitectura Event-Driven de MediSupply. Este servicio simula sensores de temperatura, humedad y estado que publican eventos a EMQX.

## 🎯 Propósito en MediSupply EDA

El MQTT Event Generator inicia el flujo de eventos en la arquitectura:

```
mqtt-event-generator → EMQX (events/sensor) → mqtt-order-event-client → EMQX (orders/events) → mqtt-kafka-bridge → Kafka
```

### Funciones Principales

- **Simulación de sensores IoT**: Genera eventos de temperatura, humedad y estado
- **Publicación MQTT**: Envía eventos al broker EMQX cada 30 segundos
- **Health monitoring**: Endpoint HTTP para verificación de estado
- **Configuración flexible**: Intervalos y topics configurables

## 🚀 Instalación

### Instalación Estándar MediSupply

```bash
# Desde el directorio k8s
helm install mqtt-event-generator ./mqtt-event-generator \
  --namespace medilogistic \
  --create-namespace

# O usando el Makefile
make deploy  # Incluye mqtt-event-generator en el despliegue completo
```

### Verificación

```bash
# Verificar pods
kubectl get pods -l app.kubernetes.io/name=mqtt-event-generator -n medilogistic

# Ver logs en tiempo real
kubectl logs -f deployment/mqtt-event-generator -n medilogistic

# Verificar health check
kubectl port-forward deployment/mqtt-event-generator 8080:8080 -n medilogistic
curl http://localhost:8080/health
```

## Configuración

### Valores principales

| Parámetro | Descripción | Valor por defecto |
|-----------|-------------|-------------------|
| `replicaCount` | Número de réplicas | `1` |
| `image.repository` | Repositorio de la imagen | `mqtt-event-generator` |
| `image.tag` | Tag de la imagen | `latest` |
| `mqtt.broker` | URL del broker MQTT | `tcp://emqx:1883` |
| `mqtt.clientId` | ID del cliente MQTT | `event-generator-k8s` |
| `mqtt.topic` | Topic donde publicar | `events/sensor` |
| `mqtt.username` | Usuario MQTT | `admin` |
| `mqtt.password` | Contraseña MQTT | `public` |
| `eventGenerator.intervalSeconds` | Frecuencia de eventos (segundos) | `30` |

### Ejemplo de configuración personalizada

```yaml
# custom-values.yaml
replicaCount: 2

mqtt:
  broker: "tcp://my-mqtt-broker:1883"
  topic: "sensors/temperature"
  username: "myuser"
  password: "mypassword"

eventGenerator:
  intervalSeconds: 10

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

```bash
helm upgrade --install mqtt-event-generator ./mqtt-event-generator \
  --namespace medisupply \
  --values custom-values.yaml
```

## Monitoreo

### Ver logs
```bash
kubectl logs -f deployment/mqtt-event-generator -n medisupply
```

### Ver estado
```bash
kubectl get pods -l app.kubernetes.io/name=mqtt-event-generator -n medisupply
```

### Ver eventos generados
Conectarse al broker MQTT y suscribirse al topic configurado para ver los eventos.

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

## Dependencias

- **EMQX**: Debe estar desplegado en el mismo namespace (`medisupply`)
- **Namespace**: Con inyección de Istio habilitada (opcional)

### Verificar conectividad con EMQX

```bash
# Verificar que EMQX esté ejecutándose
kubectl get pods -l app.kubernetes.io/name=emqx -n medisupply

# Verificar el servicio EMQX
kubectl get svc emqx -n medisupply

# Verificar health check del event generator
curl http://localhost:8080/health  # Después de port-forward

# Port-forward para testing
kubectl port-forward deployment/mqtt-event-generator 8080:8080 -n medisupply
```

## 📊 Eventos Generados

### Estructura del Evento

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

### Topics MQTT

| Topic | Descripción | Frecuencia |
|-------|-------------|------------|
| `events/sensor` | Eventos de sensores IoT | Cada 30s (configurable) |

## 🔧 Integración con MediSupply

### Dependencias

- **EMQX**: Debe estar desplegado en namespace `medilogistic`
- **mqtt-order-event-client**: Consumidor de los eventos generados
- **Istio**: Inyección de sidecar habilitada (opcional)

### Flujo de Datos

1. **mqtt-event-generator** publica eventos a `events/sensor`
2. **mqtt-order-event-client** consume eventos de `events/sensor`
3. **mqtt-order-event-client** procesa y publica a `orders/events`
4. **mqtt-kafka-bridge** consume de `orders/events` y envía a Kafka

### Configuración de EMQX

El chart asume que EMQX está configurado con:

- **Usuario**: `admin`
- **Contraseña**: `public`
- **Puerto MQTT**: `1883`
- **Servicio**: `emqx.medilogistic.svc.cluster.local`

## 🚨 Troubleshooting

### Problemas Comunes

1. **No se conecta a EMQX**:

   ```bash
   # Verificar logs
   kubectl logs deployment/mqtt-event-generator -n medilogistic
   
   # Verificar configuración
   kubectl get configmap mqtt-event-generator -n medilogistic -o yaml
   ```

2. **Eventos no se publican**:

   ```bash
   # Verificar topic en EMQX dashboard
   kubectl port-forward svc/emqx 18083:18083 -n medilogistic
   # Ir a http://localhost:18083 → WebSocket → Subscribe to events/sensor
   ```

3. **Pod no inicia**:

   ```bash
   kubectl describe pod -l app.kubernetes.io/name=mqtt-event-generator -n medilogistic
   ```

### Verificar Conectividad

```bash
# Verificar que EMQX esté ejecutándose
kubectl get pods -l app.kubernetes.io/name=emqx -n medilogistic

# Verificar el servicio EMQX
kubectl get svc emqx -n medilogistic

# Verificar conectividad desde el pod
kubectl exec -it deployment/mqtt-event-generator -n medilogistic -- nc -zv emqx 1883
```