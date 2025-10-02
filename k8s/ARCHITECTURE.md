# Arquitectura MediSupply EDA - Diseño con Helm

## 🎯 Filosofía de Diseño

Esta implementación sigue las mejores prácticas de Helm, reutilizando charts existentes con diferentes configuraciones en lugar de duplicar código.

## 📦 Reutilización de Charts

### Kafka (2 instancias)
- **kafka**: Cluster principal (puerto 9092)
  - Configuración: `config/kafka-values.yaml`
  - Propósito: Recepción de eventos desde MQTT Bridge
  
- **kafka-pedidos**: Cluster secundario (puerto 9093)
  - Configuración: `config/kafka-pedidos-values.yaml`
  - Propósito: Destino de MirrorMaker2 para procesamiento de pedidos

### Kafka UI (2 instancias)
- **kafka-ui**: Monitoreo del cluster principal
  - Configuración: valores por defecto
  - URL: http://localhost:9090
  
- **kafka-ui-pedidos**: Monitoreo del cluster secundario
  - Configuración: `config/kafka-ui-pedidos-values.yaml`
  - URL: http://localhost:9091

### Strimzi Operator (1 instancia)
- **strimzi-operator**: Gestiona ambos clusters y recursos
  - Configuración: `config/strimzi-values.yaml`
  - Gestiona: Kafka Connect, MirrorMaker2, Topics, Conectores

## 🔄 Flujos de Datos

### Flujo 1: Eventos → RabbitMQ
```
mqtt-event-generator → EMQX → mqtt-order-event-client → EMQX → mqtt-kafka-bridge → Kafka (9092) → Kafka Connect → RabbitMQ
```

### Flujo 2: Eventos → Kafka Pedidos
```
mqtt-event-generator → EMQX → mqtt-order-event-client → EMQX → mqtt-kafka-bridge → Kafka (9092) → MirrorMaker2 → Kafka Pedidos (9093)
```

## 🏗️ Ventajas del Diseño

### ✅ Reutilización de Código
- Un solo chart de Kafka para múltiples instancias
- Un solo chart de Kafka UI para múltiples clusters
- Configuración específica via values files

### ✅ Mantenibilidad
- Actualizaciones centralizadas en charts base
- Configuraciones separadas y versionables
- Fácil escalabilidad horizontal

### ✅ Consistencia
- Misma configuración base para todos los clusters
- Patrones de naming consistentes
- Gestión unificada de recursos

### ✅ Flexibilidad
- Fácil adición de nuevos clusters
- Configuraciones independientes por entorno
- Despliegue selectivo de componentes

## 📁 Estructura de Configuración

```
config/
├── kafka-values.yaml           # Cluster principal
├── kafka-pedidos-values.yaml   # Cluster secundario
├── kafka-ui-pedidos-values.yaml # UI para cluster secundario
├── strimzi-values.yaml         # Operador Strimzi
├── kind-config.yaml            # Configuración Kind
└── minikube-config.yaml        # Configuración Minikube
```

## 🚀 Comandos de Despliegue

### Despliegue Completo
```bash
make init deploy
```

### Despliegue por Componentes
```bash
# Solo clusters Kafka
helm upgrade --install kafka ./kafka --values ./config/kafka-values.yaml -n medisupply
helm upgrade --install kafka-pedidos ./kafka --values ./config/kafka-pedidos-values.yaml -n medisupply

# Solo UIs
helm upgrade --install kafka-ui ./kafka-ui -n medisupply
helm upgrade --install kafka-ui-pedidos ./kafka-ui --values ./config/kafka-ui-pedidos-values.yaml -n medisupply

# Solo Strimzi
helm upgrade --install strimzi-operator ./strimzi-kafka-operator --values ./config/strimzi-values.yaml -n medisupply
```

## 🔧 Configuraciones Específicas

### Kafka Pedidos
- **Puerto**: 9093 (evita conflictos)
- **Recursos**: Optimizados para cluster secundario
- **Persistencia**: Deshabilitada (desarrollo)
- **Replicación**: Factor 1

### Kafka UI Pedidos
- **Bootstrap**: kafka-pedidos:9093
- **Nombre**: kafka-pedidos
- **Puerto**: 8080 (interno)

### Strimzi Operator
- **Namespace**: medisupply únicamente
- **Recursos**: Optimizados para desarrollo
- **Features**: Configuración mínima

## 🔍 Monitoreo y Observabilidad

### Dashboards Disponibles
| Servicio | URL | Cluster |
|----------|-----|---------|
| Kafka UI Principal | http://localhost:9090 | kafka:9092 |
| Kafka UI Pedidos | http://localhost:9091 | kafka-pedidos:9093 |
| EMQX Dashboard | http://localhost:18083 | MQTT Broker |
| RabbitMQ Management | http://localhost:15672 | Message Queue |
| Kiali | http://localhost:20001 | Service Mesh |

### Comandos de Estado
```bash
make status                    # Estado general
kubectl get kafka -n medisupply    # Clusters Kafka
kubectl get kafkaconnect -n medisupply # Conectores
kubectl get kafkamirrormaker2 -n medisupply # MM2
```

## 🛠️ Troubleshooting

### Verificar Conectividad entre Clusters
```bash
# Desde Kafka principal
kubectl exec -n medisupply kafka-controller-0 -- kafka-topics.sh --bootstrap-server localhost:9092 --list

# Desde Kafka pedidos
kubectl exec -n medisupply kafka-pedidos-controller-0 -- kafka-topics.sh --bootstrap-server localhost:9093 --list
```

### Verificar MirrorMaker2
```bash
# Estado de MM2
kubectl describe kafkamirrormaker2 kafka-mm2 -n medisupply

# Topics replicados
kubectl exec -n medisupply kafka-pedidos-controller-0 -- kafka-topics.sh --bootstrap-server localhost:9093 --list | grep source
```

Esta arquitectura proporciona una base sólida, mantenible y escalable para el sistema MediSupply EDA.