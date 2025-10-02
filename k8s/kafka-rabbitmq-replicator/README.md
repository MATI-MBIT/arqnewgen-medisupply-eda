# Kafka-RabbitMQ Replicator

Un replicador bidireccional de mensajes entre Apache Kafka y RabbitMQ, diseñado para ejecutarse en Kubernetes.

## Características

- **Replicación bidireccional**: Kafka → RabbitMQ y RabbitMQ → Kafka
- **Configuración flexible**: Mapeo personalizable de topics/queues
- **Deduplicación**: Evita procesamiento duplicado de mensajes
- **Manejo de errores robusto**: Reintentos automáticos y logging detallado
- **Graceful shutdown**: Manejo de señales SIGINT/SIGTERM
- **Health checks**: Probes de liveness y readiness
- **Métricas**: Logging de estadísticas y heartbeat

## Arquitectura

```
┌─────────────┐    K2R     ┌─────────────┐
│    Kafka    │ ────────► │  RabbitMQ   │
│   Topics    │           │   Queues    │
└─────────────┘           └─────────────┘
       ▲                         │
       │          R2K            │
       └─────────────────────────┘
```

## Configuración

### Kafka to RabbitMQ (K2R)

```yaml
replication:
  kafkaToRabbitmq:
    enabled: true
    mappings:
      - kafkaTopic: "events-order-damage"
        rabbitmqQueue: "order-damage-queue"
        rabbitmqExchange: "events"
        rabbitmqRoutingKey: "order.damage"
      - kafkaTopic: "events-sensor"
        rabbitmqQueue: "sensor-queue"
        rabbitmqExchange: "events"
        rabbitmqRoutingKey: "sensor.data"
```

### RabbitMQ to Kafka (R2K)

```yaml
replication:
  rabbitmqToKafka:
    enabled: true
    mappings:
      - rabbitmqQueue: "warehouse-events"
        kafkaTopic: "warehouse-events"
      - rabbitmqQueue: "notifications"
        kafkaTopic: "notifications"
```

## Instalación

### Prerequisitos

- Kubernetes cluster
- Helm 3.x
- Apache Kafka desplegado
- RabbitMQ desplegado

### Despliegue

1. **Configurar valores**:
   ```bash
   cp values.yaml my-values.yaml
   # Editar my-values.yaml con tu configuración
   ```

2. **Instalar el chart**:
   ```bash
   helm install kafka-rabbitmq-replicator . -f my-values.yaml
   ```

3. **Verificar despliegue**:
   ```bash
   kubectl get pods -l app.kubernetes.io/name=kafka-rabbitmq-replicator
   kubectl logs -l component=kafka-to-rabbitmq
   kubectl logs -l component=rabbitmq-to-kafka
   ```

## Configuración Detallada

### Conexiones

```yaml
# Kafka
kafka:
  bootstrapServers: "kafka:9092"

# RabbitMQ
rabbitmq:
  host: "rabbitmq"
  port: 5672
  username: "user"
  password: "password"
  vhost: "/"
```

### Consumer Groups

```yaml
consumerGroup:
  kafkaToRabbitmq: "kafka-rabbitmq-replicator-k2r"
  rabbitmqToKafka: "kafka-rabbitmq-replicator-r2k"
```

### Recursos

```yaml
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### Health Checks

```yaml
healthCheck:
  enabled: true
  initialDelaySeconds: 30
  periodSeconds: 10
```

## Mapeo de Mensajes

### Kafka → RabbitMQ

Los mensajes de Kafka se publican en RabbitMQ con:
- **Headers adicionales**: `kafka_topic`, `kafka_partition`, `kafka_offset`, `kafka_key`
- **Persistencia**: Mensajes marcados como persistentes
- **Exchange/Queue**: Según configuración del mapeo

### RabbitMQ → Kafka

Los mensajes de RabbitMQ se publican en Kafka con:
- **Headers adicionales**: `rabbitmq_queue`, `rabbitmq_exchange`, `rabbitmq_routing_key`
- **Key preservation**: Mantiene la key original si está disponible
- **Acknowledgment**: ACK solo después de confirmación de Kafka

## Monitoreo

### Logs

```bash
# Ver logs de K2R
kubectl logs -l component=kafka-to-rabbitmq -f

# Ver logs de R2K
kubectl logs -l component=rabbitmq-to-kafka -f
```

### Métricas en Logs

- **Heartbeat cada 60s**: Estadísticas de uptime, mensajes procesados, errores
- **Rate de procesamiento**: Mensajes por segundo
- **Deduplicación**: Tracking de mensajes ya procesados

## Troubleshooting

### Problemas Comunes

1. **Conexión a Kafka**:
   ```bash
   kubectl exec -it <pod> -- python -c "
   from kafka import KafkaConsumer
   consumer = KafkaConsumer(bootstrap_servers=['kafka:9092'])
   print('Kafka OK')
   "
   ```

2. **Conexión a RabbitMQ**:
   ```bash
   kubectl exec -it <pod> -- python -c "
   import pika
   connection = pika.BlockingConnection(pika.ConnectionParameters('rabbitmq'))
   print('RabbitMQ OK')
   "
   ```

3. **Verificar mappings**:
   ```bash
   kubectl get configmap kafka-rabbitmq-replicator-script -o yaml
   ```

### Logs de Debug

Para habilitar logging más detallado, modificar el script:
```python
logging.basicConfig(level=logging.DEBUG)
```

## Limitaciones

- **Orden de mensajes**: No garantizado entre particiones diferentes
- **Exactly-once**: Implementa at-least-once con deduplicación
- **Schemas**: No maneja schemas complejos automáticamente
- **Transacciones**: No soporta transacciones distribuidas

## Desarrollo

### Estructura del Proyecto

```
k8s/kafka-rabbitmq-replicator/
├── Chart.yaml
├── values.yaml
├── README.md
├── scripts/
│   └── replicator.py
└── templates/
    ├── _helpers.tpl
    ├── configmap.yaml
    └── deployment.yaml
```

### Testing Local

```bash
# Instalar dependencias
pip install kafka-python pika

# Ejecutar replicador
export KAFKA_BOOTSTRAP_SERVERS="localhost:9092"
export RABBITMQ_HOST="localhost"
export REPLICATION_MAPPINGS='[{"kafkaTopic":"test","rabbitmqQueue":"test"}]'
python scripts/replicator.py K2R
```

## Contribución

1. Fork el repositorio
2. Crear feature branch
3. Commit cambios
4. Push al branch
5. Crear Pull Request

## Licencia

MIT License