# Kafka Replicator Chart

Este chart de Helm despliega un replicador de tópicos de Kafka basado en Python y kafka-python, que permite replicar tópicos entre dos clusters de Kafka de forma bidireccional.

## Características

- **Replicación bidireccional**: Permite replicar tópicos de Kafka a Kafka-warehouse y viceversa
- **Configuración flexible**: Define fácilmente qué tópicos replicar en cada dirección
- **Basado en Python**: Utiliza kafka-python para una replicación simple y confiable
- **Contenedores separados**: Un contenedor por dirección de replicación para mejor aislamiento
- **Monitoreo**: Incluye logging detallado para monitorear el estado de la replicación

## Configuración

### Clusters de Kafka

```yaml
sourceKafka:
  name: "kafka"
  bootstrapServers: "kafka:9092"
  
targetKafka:
  name: "kafka-warehouse"
  bootstrapServers: "kafka-warehouse:9092"
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

## Instalación

```bash
helm install kafka-replicator ./k8s/kafka-replicator
```

## Verificación

Para verificar que los replicadores están funcionando:

```bash
# Ver el estado del pod
kubectl get pods -l app=kafka-replicator

# Ver logs del replicador source-to-target
kubectl logs -l app=kafka-replicator -c source-to-target-replicator

# Ver logs del replicador target-to-source
kubectl logs -l app=kafka-replicator -c target-to-source-replicator

# Ver todos los logs
kubectl logs -l app=kafka-replicator --all-containers=true
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