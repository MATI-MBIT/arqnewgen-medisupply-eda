# Warehouse Batch Helm Chart

Este chart despliega el servicio Warehouse Batch Event Processing en Kubernetes.

## Descripción

El Warehouse Batch es un servicio que consume eventos de Kafka de forma continua para procesamiento de operaciones de almacén. Implementa arquitectura hexagonal con adaptadores para Kafka y expone una API HTTP para monitoreo.

## Instalación

```bash
# Desde la carpeta k8s
helm upgrade --install warehouse-batch ./warehouse-batch \
  --namespace medisupply \
  --create-namespace
```

## Configuración

### Valores principales

| Parámetro | Descripción | Valor por defecto |
|-----------|-------------|-------------------|
| `replicaCount` | Número de réplicas | `1` |
| `image.repository` | Repositorio de la imagen | `warehouse-batch` |
| `image.tag` | Tag de la imagen | `latest` |
| `kafka.brokerAddress` | Dirección del broker Kafka | `kafka:9092` |
| `kafka.topic` | Topic de Kafka para consumir eventos | `warehouse-events` |
| `kafka.groupId` | ID del grupo de consumidores Kafka | `warehouse-batch-service` |
| `service.port` | Puerto HTTP del servicio | `8080` |

### Ejemplo de configuración personalizada

```yaml
# custom-values.yaml
replicaCount: 2

kafka:
  brokerAddress: "my-kafka.example.com:9092"
  topic: "warehouse-events-prod"
  groupId: "warehouse-batch-service-prod"

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi
```

```bash
helm upgrade --install warehouse-batch ./warehouse-batch \
  --namespace medisupply \
  --values custom-values.yaml
```

## Componentes

### Deployment
- Consumidor continuo de eventos Kafka
- Servicio HTTP para monitoreo y health checks
- Expone endpoints `/health` para verificación de estado
- Configurado con probes de liveness y readiness
- Implementa arquitectura hexagonal con adaptadores

### Service
- Expone el servicio HTTP internamente en el cluster
- Puerto configurable (por defecto 8080)

## Monitoreo

### Ver logs del servicio
```bash
kubectl logs -f deployment/warehouse-batch -n medisupply
```



### Ver estado
```bash
kubectl get pods -l app.kubernetes.io/name=warehouse-batch -n medisupply

```

### Ver API
```bash
kubectl port-forward svc/warehouse-batch 8080:8080 -n medisupply
curl http://localhost:8080/health
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

## Dependencias

- **Kafka**: Broker de mensajes para consumir eventos de warehouse
- **Namespace**: Con inyección de Istio habilitada (opcional)

## Arquitectura del Servicio

El servicio implementa arquitectura hexagonal:

- **Domain Layer**: Entidades de negocio y contratos
- **Application Layer**: Lógica de negocio para procesamiento de eventos
- **Infrastructure Layer**: 
  - **Driving Adapters**: EventConsumerAdapter (Kafka), ApiServiceAdapter (HTTP)
  - **Driven Adapters**: KafkaProducer (para demo)