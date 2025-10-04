# Kafka Replicator Chart

Este chart de Helm despliega un replicador de tópicos de Kafka basado en Python y kafka-python, que permite replicar tópicos entre dos clusters de Kafka de forma bidireccional con soporte completo para consumer groups.

## Características

- **Replicación bidireccional**: Permite replicar tópicos de Kafka a Kafka-warehouse y viceversa
- **Consumer Groups**: Implementación completa de consumer groups para escalabilidad y tolerancia a fallos
- **Escalabilidad horizontal**: Soporte para múltiples réplicas que se distribuyen automáticamente las particiones
- **Configuración flexible**: Define fácilmente qué tópicos replicar en cada dirección
- **Basado en Python**: Utiliza kafka-python para una replicación simple y confiable
- **Contenedores separados**: Un contenedor por dirección de replicación para mejor aislamiento
- **Monitoreo**: Incluye logging detallado para monitorear el estado de la replicación
- **Gestión de offsets**: Soporte para auto-commit y commit manual de offsets

## Configuración

### Clusters de Kafka

```yaml
sourceKafka:
  bootstrapServers: "kafka:9092"
  
targetKafka:
  bootstrapServers: "kafka-warehouse:9092"
```

### Consumer Groups

```yaml
consumerGroup:
  sourceToTarget: "kafka-replicator-s2t-v4"
  targetToSource: "kafka-replicator-t2s-v4"
```

### Escalabilidad

```yaml
replicaCount: 3  # Múltiples instancias para alta disponibilidad
```

### Replicación de Tópicos

#### De Kafka a Kafka-warehouse
```yaml
replication:
  sourceToTarget:
    enabled: true
    topics:
      - sourceTopicName: "damage"
        targetTopicName: "damage"
      - sourceTopicName: "events-sensor"
        targetTopicName: "events-sensor"
```

#### De Kafka-warehouse a Kafka
```yaml
replication:
  targetToSource:
    enabled: true
    topics:
      - sourceTopicName: "warehouse-events"
        targetTopicName: "warehouse-events"
```

## Consumer Groups

Este chart implementa consumer groups de Kafka para proporcionar:

### Escalabilidad Horizontal
- **Múltiples réplicas**: Puedes ejecutar múltiples instancias del replicador
- **Distribución automática**: Las particiones se distribuyen automáticamente entre las instancias
- **Balanceado de carga**: Cada instancia procesa un subconjunto de particiones

### Tolerancia a Fallos
- **Rebalanceo automático**: Si una instancia falla, sus particiones se reasignan automáticamente
- **Gestión de offsets**: Los offsets se almacenan en Kafka para recuperación automática
- **Detección de fallos**: Heartbeats y timeouts configurables

### Configuración de Consumer Groups

El chart está optimizado para baja latencia y sin lag:
- **Auto-commit**: Habilitado con intervalo de 1 segundo
- **Auto offset reset**: `earliest` (procesa todos los mensajes disponibles)
- **Polling**: Timeout corto (1s) para mejor responsividad
- **Producer**: Flush inmediato después de cada mensaje
- **Batching**: Optimizado para latencia (`linger_ms=0`, `batch_size=1024`)

### Configuración de Latencia

```yaml
consumer:
  autoOffsetReset: "earliest"  # earliest = todos los mensajes, latest = solo nuevos
```

## Instalación

```bash
helm install kafka-replicator ./k8s/kafka-replicator
```

### Instalación con múltiples réplicas

```bash
helm install kafka-replicator ./k8s/kafka-replicator --set replicaCount=3
```

## Verificación

Para verificar que los replicadores están funcionando:

```bash
# Ver el estado de todos los pods
kubectl get pods -l app=kafka-replicator

# Ver logs de una instancia específica
kubectl logs kafka-replicator-<hash> -c source-to-target-replicator

# Ver logs de todas las instancias
kubectl logs -l app=kafka-replicator -c source-to-target-replicator

# Ver logs en tiempo real
kubectl logs -f -l app=kafka-replicator --all-containers=true

# Verificar consumer groups en Kafka
kubectl exec -it kafka-0 -- kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group kafka-replicator-s2t-v4
```

### Monitoreo de Consumer Groups

```bash
# Ver estado del consumer group
kubectl exec -it kafka-0 -- kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group kafka-replicator-s2t-v4

# Ver lag por partición
kubectl exec -it kafka-0 -- kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group kafka-replicator-s2t-v4 \
  --verbose

# Listar todos los consumer groups
kubectl exec -it kafka-0 -- kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --list
```

## Personalización

Puedes personalizar la configuración creando tu propio archivo `values.yaml`:

```yaml
# custom-values.yaml
replication:
  sourceToTarget:
    enabled: true
    topics:
      - sourceTopicName: "mi-topico-origen"
        targetTopicName: "mi-topico-destino"

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
```

Luego instalar con:
```bash
helm install kafka-replicator ./k8s/kafka-replicator -f custom-values.yaml
```
## T
roubleshooting Consumer Groups

### Problema: Instancias no procesan mensajes

```bash
# Verificar asignación de particiones
kubectl logs -l app=kafka-replicator -c source-to-target-replicator | grep "assignment"

# Verificar que hay particiones disponibles
kubectl exec -it kafka-0 -- kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --describe --topic events-order-damage
```

### Problema: Rebalanceo constante

```bash
# Verificar logs de rebalanceo
kubectl logs -l app=kafka-replicator -c source-to-target-replicator | grep -i rebalance

# Ajustar timeouts si es necesario
helm upgrade kafka-replicator ./k8s/kafka-replicator \
  --set consumerGroup.sourceToTarget.sessionTimeoutMs=45000 \
  --set consumerGroup.sourceToTarget.maxPollIntervalMs=600000
```

### Problema: Lag alto en consumer group

```bash
# Verificar lag
kubectl exec -it kafka-0 -- kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group kafka-replicator-s2t-v4

# Escalar horizontalmente
helm upgrade kafka-replicator ./k8s/kafka-replicator --set replicaCount=5
```

### Problema: Mensajes duplicados

```bash
# Verificar configuración de auto-commit
kubectl logs -l app=kafka-replicator -c source-to-target-replicator | grep "Auto commit"
```

### Problema: Lag de 1 mensaje (mensaje atrasado)

Este problema se ha solucionado con las siguientes optimizaciones:

```bash
# Verificar configuración de offset reset
kubectl logs -l app=kafka-replicator -c source-to-target-replicator | grep "AUTO_OFFSET_RESET"

# Si aún hay lag, verificar que esté usando 'earliest'
helm upgrade kafka-replicator ./k8s/kafka-replicator \
  --set consumer.autoOffsetReset=earliest
```

**Optimizaciones aplicadas:**
- `auto_offset_reset=earliest`: Procesa todos los mensajes disponibles
- `auto_commit_interval_ms=1000`: Commit más frecuente
- `linger_ms=0`: Envío inmediato sin esperar
- `flush()` después de cada mensaje: Garantiza entrega inmediata
- Polling timeout corto: Mejor responsividad

## Mejores Prácticas

### Configuración de Producción

```yaml
# values-production.yaml
replicaCount: 3

consumerGroup:
  sourceToTarget:
    groupId: "kafka-replicator-s2t-prod"
    sessionTimeoutMs: 45000
    heartbeatIntervalMs: 3000
    maxPollRecords: 100
    maxPollIntervalMs: 600000
    autoOffsetReset: "earliest"
    enableAutoCommit: false  # Control manual para mayor confiabilidad
    autoCommitIntervalMs: 5000

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi
```

### Monitoreo Continuo

1. **Configurar alertas** para lag alto en consumer groups
2. **Monitorear logs** para errores de rebalanceo
3. **Verificar métricas** de throughput y latencia
4. **Revisar recursos** de CPU y memoria regularmente

### Escalado

- **Horizontal**: Aumentar `replicaCount` para más throughput
- **Vertical**: Aumentar recursos por pod para mensajes grandes
- **Particiones**: Asegurar suficientes particiones para paralelismo óptimo