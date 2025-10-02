# MediSupply EDA - Kubernetes Infrastructure

Infraestructura Kubernetes completa para el sistema MediSupply con arquitectura Event-Driven usando Istio Service Mesh, Apache Kafka, Strimzi y KEDA.

## 🏗️ Arquitectura Event-Driven

### Flujo Principal: MQTT → Kafka
```
mqtt-event-generator → EMQX → mqtt-order-event-client → EMQX → mqtt-kafka-bridge → Kafka (Principal)
```

### Replicación Bidireccional: MirrorMaker 2
```
Kafka Principal ⟷ Kafka Warehouse
├── damage, events-sensor → (Principal → Warehouse)
└── new, inventory-updates ← (Principal ← Warehouse)
```

## 🔧 Componentes Principales

- **Istio Service Mesh**: Comunicación segura entre servicios
- **Apache Kafka**: Sistema de mensajería central para eventos (2 clusters: Principal y Warehouse)
- **MirrorMaker 2**: Replicación bidireccional parametrizable por topic entre clusters

- **EMQX**: Broker MQTT para IoT y eventos en tiempo real
- **RabbitMQ**: Sistema de colas para procesamiento asíncrono
- **KEDA**: Autoescalado basado en eventos

## 🚀 Inicio Rápido

### Prerrequisitos

- Docker
- kubectl
- helm
- kind o minikube

### Despliegue Completo

```bash
# Crear cluster e instalar toda la infraestructura
make init

# Desplegar arquitectura EDA completa
make deploy

# Verificar estado de componentes
make status
```

### Comandos Disponibles

```bash
make help                    # Mostrar ayuda
make init                    # Crear cluster con Kind (default)
make init PROVIDER=minikube  # Crear cluster con Minikube
make deploy                  # Desplegar arquitectura EDA completa
make status                  # Mostrar estado de componentes
make kafka-ui                # Abrir Kafka UI (http://localhost:9090)
make kafka-pedidos-ui        # Abrir Kafka Pedidos UI (http://localhost:9091)
make mqtt                    # Abrir EMQX dashboard (http://localhost:18083)
make rabbitmq                # Abrir RabbitMQ UI (http://localhost:15672)
make kiali                   # Abrir Kiali dashboard (http://localhost:20001)
make clean                   # Eliminar charts del cluster
make destroy                 # Eliminar cluster completamente
```

## 🔧 Componentes

### Istio Service Mesh
- **Base**: Componentes fundamentales
- **Istiod**: Plano de control
- **Gateway**: Punto de entrada (NodePort)
- **Addons**: Prometheus, Jaeger, Grafana, Kiali

### Apache Kafka
- **Versión**: 4.0.0
- **Configuración**: Desarrollo sin persistencia
- **Protocolo**: PLAINTEXT (sin SASL)
- **Replicación**: Factor 1
- **Namespace**: medisupply

### KEDA
- **Versión**: 2.17.2
- **Autoescalado**: Basado en métricas de Kafka
- **Namespace**: keda-system

### Kafka UI
- **Puerto**: 9090 (local)
- **Funcionalidades**:
  - Gestión de topics
  - Producir/consumir mensajes
  - Monitoreo del cluster
  - Gestión de consumer groups

## 🌐 Acceso a Dashboards

| Servicio | URL | Descripción |
|----------|-----|-------------|
| Kafka UI | http://localhost:9090 | Interfaz de gestión de ambos clusters Kafka |
| Kafka Pedidos UI | http://localhost:9091 | Interfaz de gestión de Kafka Pedidos |
| EMQX Dashboard | http://localhost:18083 | Gestión del broker MQTT |
| RabbitMQ Management | http://localhost:15672 | Gestión de colas RabbitMQ |
| Kiali | http://localhost:20001 | Observabilidad de Istio |

## 📁 Estructura del Proyecto

```
k8s/
├── istio/                    # Charts de Istio
│   ├── base/                 # Componentes base
│   ├── istiod/               # Plano de control
│   └── gateway/              # Gateway de entrada
├── kafka/                    # Chart de Apache Kafka Central
├── kafka-ui/                 # Chart de Kafka UI

├── mqtt/                     # Charts MQTT
│   └── emqx/                 # EMQX broker
├── rabbitmq/                 # Chart de RabbitMQ
├── mqtt-event-generator/     # Chart del generador de eventos
├── mqtt-order-event-client/  # Chart del cliente de eventos
├── mqtt-kafka-bridge/        # Chart del puente MQTT-Kafka
├── strimzi-kafka-operator/   # Chart de Strimzi Operator
├── strimzi-resources/        # Recursos CRDs de Strimzi
│   ├── kafka-topics.yaml
│   ├── kafka-connect.yaml
│   ├── kafka-mirrormaker2.yaml
│   ├── kafka-connect-rabbitmq-connector.yaml
│   ├── rabbitmq-credentials-secret.yaml
│   └── README.md
├── keda/                     # Chart de KEDA
├── kafka-mirror-maker2/      # Chart de MirrorMaker 2
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
├── config/                   # Configuraciones
│   ├── kind-config.yaml
│   ├── minikube-config.yaml
│   ├── kafka-values.yaml
│   ├── kafka-warehouse-values.yaml
│   └── mirror-maker2-values.yaml
├── Makefile                  # Comandos de gestión
└── README.md                 # Este archivo
```

## ⚙️ Configuración

### Cluster Local
- **Kind**: Puertos 80/443 expuestos
- **Minikube**: 2 CPUs, 6GB RAM, 20GB disco

### Kafka
- **Bootstrap Servers**: kafka:9092
- **Auto Create Topics**: Habilitado
- **Persistencia**: Deshabilitada (desarrollo)

### Observabilidad
- **Métricas**: Prometheus
- **Trazas**: Jaeger
- **Dashboards**: Grafana
- **Service Mesh**: Kiali

## 🔍 Troubleshooting

### Kafka UI muestra "Cluster Offline"
```bash
# Verificar pods de Kafka
kubectl get pods -n medisupply -l app.kubernetes.io/name=kafka

# Ver logs de Kafka
kubectl logs -n medisupply kafka-controller-0 -c kafka
```

### Port-forward falla
```bash
# Verificar que el pod esté corriendo
kubectl get pods -n medisupply -l app.kubernetes.io/name=kafka-ui

# Reiniciar el pod
kubectl delete pod -n medisupply -l app.kubernetes.io/name=kafka-ui
```

### Recrear desde cero
```bash
make destroy  # Eliminar cluster
make init     # Crear nuevo cluster
make deploy   # Desplegar servicios
```

## 📝 Notas de Desarrollo

- **Persistencia**: Deshabilitada para desarrollo rápido
- **Seguridad**: PLAINTEXT para simplicidad
- **Recursos**: Configuración mínima para desarrollo local
- **Istio**: Inyección automática de sidecars habilitada

## 🤝 Contribución

1. Modificar configuraciones en `/config`
2. Actualizar charts en sus respectivos directorios
3. Probar con `make clean && make deploy`
4. Documentar cambios en este README
