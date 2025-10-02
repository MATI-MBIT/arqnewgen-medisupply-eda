# MQTT-Kafka Bridge Chart - MediSupply EDA

Chart de Helm para desplegar el puente entre MQTT y Kafka en la arquitectura Event-Driven de MediSupply. Este servicio conecta el ecosistema MQTT con Apache Kafka para persistencia y procesamiento distribuido.

## 🎯 Propósito en MediSupply EDA

El MQTT-Kafka Bridge completa el flujo de ingesta de eventos:

```
mqtt-event-generator → EMQX → mqtt-order-event-client → EMQX → mqtt-kafka-bridge → Kafka Principal
```

### Funciones Principales

- **Suscripción MQTT**: Consume eventos desde topics MQTT específicos
- **Publicación Kafka**: Envía eventos a topics de Kafka para persistencia
- **Transformación**: Convierte mensajes MQTT a formato Kafka
- **Routing**: Mapea topics MQTT a topics Kafka configurables

## 🚀 Instalación

### Instalación Estándar MediSupply

```bash
# Desde el directorio k8s
helm install mqtt-kafka-bridge ./mqtt-kafka-bridge \
  --namespace medisupply \
  --create-namespace

# O usando el Makefile
make deploy  # Incluye mqtt-kafka-bridge en el despliegue completo
```

### Verificación

```bash
# Verificar pods
kubectl get pods -l app.kubernetes.io/name=mqtt-kafka-bridge -n medisupply

# Ver logs en tiempo real
kubectl logs -f deployment/mqtt-kafka-bridge -n medisupply

# Verificar conectividad
kubectl exec -it deployment/mqtt-kafka-bridge -n medisupply -- nc -zv kafka 9092
kubectl exec -it deployment/mqtt-kafka-bridge -n medisupply -- nc -zv emqx 1883
```

## ⚙️ Configuración

### Valores Principales

| Parámetro | Descripción | Valor por defecto |
|-----------|-------------|-------------------|
| `replicaCount` | Número de réplicas | `1` |
| `image.repository` | Repositorio de la imagen | `mqtt-kafka-bridge` |
| `image.tag` | Tag de la imagen | `latest` |
| `mqtt.broker` | URL del broker MQTT | `tcp://emqx.medilogistic.svc.cluster.local:1883` |
| `mqtt.topics` | Topics MQTT a suscribir | `["orders/events"]` |
| `kafka.bootstrapServers` | Servidores Kafka | `kafka.medisupply.svc.cluster.local:9092` |
| `kafka.topics` | Topics Kafka destino | `["events-order"]` |

### Configuración de Mapeo

```yaml
# values.yaml personalizado
mqtt:
  broker: "tcp://emqx.medilogistic.svc.cluster.local:1883"
  username: "admin"
  password: "public"
  clientId: "mqtt-kafka-bridge"

kafka:
  bootstrapServers: "kafka.medisupply.svc.cluster.local:9092"
  
# Mapeo de topics
topicMapping:
  - mqttTopic: "orders/events"
    kafkaTopic: "events-order"
  - mqttTopic: "sensors/damage"
    kafkaTopic: "events-damage"
  - mqttTopic: "inventory/updates"
    kafkaTopic: "inventory-updates"

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

## 🔧 Configuración Avanzada

### Configuración de Kafka Producer

```yaml
kafka:
  producer:
    acks: "all"
    retries: 3
    batchSize: 16384
    lingerMs: 5
    bufferMemory: 33554432
    compressionType: "gzip"
```

### Configuración MQTT

```yaml
mqtt:
  qos: 1
  cleanSession: false
  keepAlive: 60
  connectionTimeout: 30
  reconnectDelay: 5
```

### Health Checks

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

## 📊 Monitoreo

### Métricas Expuestas

| Métrica | Descripción |
|---------|-------------|
| `mqtt_messages_received_total` | Total de mensajes MQTT recibidos |
| `kafka_messages_sent_total` | Total de mensajes enviados a Kafka |
| `bridge_errors_total` | Total de errores de procesamiento |
| `bridge_latency_seconds` | Latencia de procesamiento |

### Endpoints de Monitoreo

```bash
# Health check
curl http://localhost:8080/health

# Métricas Prometheus
curl http://localhost:8080/metrics

# Estado del bridge
curl http://localhost:8080/status
```

## 🔄 Integración con MediSupply

### Dependencias

- **EMQX**: Broker MQTT en namespace `medilogistic`
- **Kafka**: Cluster principal en namespace `medisupply`
- **mqtt-order-event-client**: Fuente de eventos procesados
- **Istio**: Service mesh para comunicación segura

### Flujo de Datos

1. **Suscribe** a topics MQTT desde `mqtt-order-event-client`
2. **Transforma** mensajes MQTT a formato Kafka
3. **Publica** eventos a topics Kafka correspondientes
4. **Mantiene** orden y garantías de entrega
5. **Reporta** métricas y estado de salud

### Topics Utilizados

#### MQTT (Source)
| Topic | Descripción | QoS |
|-------|-------------|-----|
| `orders/events` | Eventos de pedidos procesados | 1 |
| `sensors/damage` | Alertas de daños | 1 |
| `inventory/updates` | Actualizaciones de inventario | 1 |

#### Kafka (Destination)
| Topic | Descripción | Particiones |
|-------|-------------|-------------|
| `events-order` | Eventos de pedidos | 3 |
| `events-damage` | Eventos de daños | 1 |
| `inventory-updates` | Actualizaciones de inventario | 2 |

## 🚨 Troubleshooting

### Problemas Comunes

1. **No se conecta a MQTT**:

   ```bash
   # Verificar logs
   kubectl logs deployment/mqtt-kafka-bridge -n medisupply
   
   # Verificar conectividad
   kubectl exec -it deployment/mqtt-kafka-bridge -n medisupply -- nc -zv emqx.medilogistic.svc.cluster.local 1883
   ```

2. **No se conecta a Kafka**:

   ```bash
   # Verificar Kafka
   kubectl get pods -l app.kubernetes.io/name=kafka -n medisupply
   
   # Test de conectividad
   kubectl exec -it deployment/mqtt-kafka-bridge -n medisupply -- nc -zv kafka.medisupply.svc.cluster.local 9092
   ```

3. **Mensajes no se procesan**:

   ```bash
   # Verificar topics Kafka
   kubectl exec -it kafka-controller-0 -n medisupply -- kafka-topics.sh --bootstrap-server localhost:9092 --list
   
   # Verificar mensajes en Kafka
   kubectl exec -it kafka-controller-0 -n medisupply -- kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic events-order --from-beginning
   ```

### Comandos de Diagnóstico

```bash
# Estado del bridge
kubectl port-forward svc/mqtt-kafka-bridge 8080:8080 -n medisupply
curl http://localhost:8080/status

# Métricas detalladas
curl http://localhost:8080/metrics | grep bridge

# Logs con filtro
kubectl logs deployment/mqtt-kafka-bridge -n medisupply | grep -E "(ERROR|WARN|Connected|Disconnected)"
```

## 🔧 Desarrollo

### Testing Local

```bash
# Publicar mensaje de prueba a MQTT
mosquitto_pub -h localhost -p 1883 -t "orders/events" -m '{"test": "message"}'

# Verificar en Kafka
kubectl exec -it kafka-controller-0 -n medisupply -- kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic events-order --from-beginning
```

### Configuración de Desarrollo

```yaml
# dev-values.yaml
replicaCount: 1

mqtt:
  broker: "tcp://localhost:1883"
  
kafka:
  bootstrapServers: "localhost:9092"

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 50m
    memory: 128Mi

logging:
  level: DEBUG
```

## 📋 Configuración por Defecto

```yaml
replicaCount: 1

image:
  repository: mqtt-kafka-bridge
  tag: latest
  pullPolicy: IfNotPresent

mqtt:
  broker: "tcp://emqx.medilogistic.svc.cluster.local:1883"
  username: "admin"
  password: "public"
  clientId: "mqtt-kafka-bridge"
  qos: 1

kafka:
  bootstrapServers: "kafka.medisupply.svc.cluster.local:9092"
  
service:
  type: ClusterIP
  port: 8080

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi
```