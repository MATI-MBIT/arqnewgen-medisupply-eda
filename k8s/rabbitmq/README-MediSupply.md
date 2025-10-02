# RabbitMQ Chart - MediSupply EDA

Sistema de colas de mensajes para la arquitectura Event-Driven de MediSupply. RabbitMQ actúa como destino de replicación desde Kafka para procesamiento asíncrono de pedidos y notificaciones.

## 🎯 Propósito en MediSupply EDA

RabbitMQ forma parte del flujo de replicación bidireccional en la arquitectura:

```
Kafka Principal ⟷ RabbitMQ (via kafka-rabbitmq-replicator)
     ↓                    ↓
Topics configurables → Queues específicas
```

### Funciones Principales

- **Procesamiento de pedidos**: Colas para órdenes de compra y inventario
- **Notificaciones**: Sistema de alertas y notificaciones push
- **Integración legacy**: Conectar sistemas existentes que usan AMQP
- **Dead letter queues**: Manejo de mensajes fallidos

## 🚀 Instalación MediSupply

### Instalación Estándar

```bash
# Desde el directorio k8s
helm install rabbitmq ./rabbitmq --namespace mediorder --create-namespace

# O usando el Makefile
make deploy  # Incluye RabbitMQ en el despliegue completo
```

### Verificación

```bash
# Verificar pods
kubectl get pods -l app.kubernetes.io/name=rabbitmq -n mediorder

# Obtener credenciales
kubectl get secret rabbitmq -n mediorder -o jsonpath="{.data.rabbitmq-password}" | base64 -d

# Acceder al Management UI
kubectl port-forward svc/rabbitmq 15672:15672 -n mediorder
```

## ⚙️ Configuración MediSupply

### Queues y Exchanges Utilizados

| Queue | Exchange | Routing Key | Propósito |
|-------|----------|-------------|-----------|
| `order-damage-queue` | `events` | `order.damage` | Procesamiento de daños en pedidos |
| `sensor-queue` | `events` | `sensor.data` | Datos de sensores IoT |
| `warehouse-events` | `warehouse` | `inventory.*` | Eventos de inventario |
| `notifications` | `notifications` | `alert.*` | Sistema de alertas |

### Configuración de Producción

```yaml
auth:
  username: medisupply
  password: "secure-password"
  erlangCookie: "medisupply-cookie"

clustering:
  enabled: true
  replicaCount: 3

persistence:
  enabled: true
  storageClass: "standard"
  size: 10Gi

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

metrics:
  enabled: true
  serviceMonitor:
    enabled: true
```

## 🌐 Acceso y Puertos

### Puertos Estándar

| Puerto | Protocolo | Descripción |
|--------|-----------|-------------|
| 5672 | AMQP | Conexiones AMQP estándar |
| 5671 | AMQPS | AMQP sobre SSL/TLS |
| 15672 | HTTP | Management UI y API |
| 15692 | HTTP | Métricas Prometheus |
| 25672 | Erlang | Comunicación entre nodos |

### Acceso al Management UI

```bash
# Port-forward para acceso local
kubectl port-forward svc/rabbitmq 15672:15672 -n mediorder

# Acceder en: http://localhost:15672
# Usuario: user (por defecto)
# Contraseña: obtener del secret
kubectl get secret rabbitmq -n mediorder -o jsonpath="{.data.rabbitmq-password}" | base64 -d
```

## 🔄 Integración con MediSupply

### Replicación desde Kafka

El `kafka-rabbitmq-replicator` maneja la sincronización:

```yaml
# Configuración del replicador
replication:
  kafkaToRabbitmq:
    enabled: true
    mappings:
      - kafkaTopic: "events-order-damage"
        rabbitmqQueue: "order-damage-queue"
        rabbitmqExchange: "events"
        rabbitmqRoutingKey: "order.damage"
```

Para más detalles sobre configuración, instalación y troubleshooting, consultar el README.md original del chart de Bitnami.