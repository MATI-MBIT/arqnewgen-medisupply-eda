# MediSupply EDA - Arquitectura Event-Driven

Sistema completo de arquitectura Event-Driven para MediSupply, implementando un flujo de eventos desde sensores IoT hasta sistemas de gestión de inventario y pedidos, utilizando tecnologías como MQTT, Apache Kafka, RabbitMQ y Kubernetes.

## 🏗️ Arquitectura General

### Flujo Principal de Eventos
```
Sensores IoT → MQTT → Kafka → RabbitMQ/Kafka Warehouse
     ↓
mqtt-event-generator → EMQX → mqtt-order-event-client → mqtt-kafka-bridge → Kafka Principal
                                                                                    ↓
                                                                          Replicación Bidireccional
                                                                                    ↓
                                                                    Kafka Warehouse ⟷ RabbitMQ
```

### Componentes Principales

- **🔧 Infraestructura K8s** (`k8s/`): Charts de Helm para despliegue completo
- **📡 Servicios** (`services/`): Microservicios en Go para generación y procesamiento de eventos
- **🔄 Replicación**: Sistemas bidireccionales entre Kafka clusters y RabbitMQ

## 🚀 Inicio Rápido

### Prerrequisitos

- Docker
- kubectl
- Helm 3.x
- Kind o Minikube

### Despliegue Completo

```bash
# Clonar el repositorio
git clone <repository-url>
cd arqnewgen-medisupply-eda

# Crear cluster e instalar infraestructura
cd k8s
make init

# Desplegar toda la arquitectura EDA
make deploy

# Verificar estado
make status
```

### Acceso a Dashboards

| Servicio | URL | Descripción |
|----------|-----|-------------|
| Kafka UI | http://localhost:9090 | Gestión de clusters Kafka |
| EMQX Dashboard | http://localhost:18083 | Broker MQTT |
| RabbitMQ Management | http://localhost:15672 | Sistema de colas |
| Kiali | http://localhost:20001 | Service Mesh (Istio) |

```bash
# Abrir dashboards
make kafka-ui    # Kafka UI
make mqtt        # EMQX
make rabbitmq    # RabbitMQ
make kiali       # Kiali
```

## 📁 Estructura del Proyecto

```
arqnewgen-medisupply-eda/
├── k8s/                           # Infraestructura Kubernetes
│   ├── istio/                     # Service Mesh
│   ├── kafka/                     # Apache Kafka (Principal)
│   ├── kafka-ui/                  # Interfaz web para Kafka
│   ├── mqtt/                      # EMQX Broker
│   ├── rabbitmq/                  # RabbitMQ
│   ├── mqtt-event-generator/      # Chart del generador
│   ├── mqtt-order-event-client/   # Chart del cliente
│   ├── mqtt-kafka-bridge/         # Puente MQTT-Kafka
│   ├── kafka-replicator/          # Replicador Kafka-Kafka
│   ├── kafka-rabbitmq-replicator/ # Replicador Kafka-RabbitMQ
│   ├── keda/                      # Autoescalado basado en eventos
│   ├── config/                    # Configuraciones
│   ├── Makefile                   # Comandos de gestión
│   ├── README.md                  # Documentación K8s
│   └── ARCHITECTURE.md            # Arquitectura detallada
├── services/                      # Microservicios
│   ├── mqtt-event-generator/      # Generador de eventos IoT
│   ├── mqtt-order-event-client/   # Cliente de eventos de pedidos
│   ├── Makefile                   # Build y despliegue
│   └── README.md                  # Documentación servicios
├── README.md                      # Este archivo
└── .gitignore
```

## 🔄 Flujos de Datos

### 1. Generación de Eventos IoT
```
mqtt-event-generator → EMQX (events/sensor) → mqtt-order-event-client
```

### 2. Procesamiento de Pedidos
```
mqtt-order-event-client → EMQX (orders/events) → mqtt-kafka-bridge → Kafka
```

### 3. Replicación Bidireccional
```
Kafka Principal ⟷ Kafka Warehouse (damage, events-sensor → / ← warehouse-events)
Kafka Principal ⟷ RabbitMQ (configurable por topic/queue)
```

## 🛠️ Desarrollo

### Servicios

Los servicios están desarrollados en Go y se pueden ejecutar localmente:

```bash
cd services

# Construir todas las imágenes
make build-all

# Para desarrollo con Kind/Minikube
make build-load-all

# Ver servicios disponibles
make help
```

### Infraestructura

La infraestructura se gestiona con Helm charts:

```bash
cd k8s

# Ver comandos disponibles
make help

# Desplegar componente específico
helm upgrade --install kafka ./kafka --namespace medisupply

# Limpiar todo
make clean
```

## 📊 Monitoreo y Observabilidad

### Métricas y Logs

- **Istio Service Mesh**: Métricas de tráfico y latencia
- **Prometheus**: Recolección de métricas
- **Jaeger**: Trazabilidad distribuida
- **Grafana**: Dashboards de monitoreo
- **Kiali**: Visualización del service mesh

### Health Checks

Todos los servicios incluyen endpoints de health check:

```bash
# Verificar estado de servicios
kubectl get pods -A
kubectl get svc -A

# Logs específicos
kubectl logs -l app=mqtt-event-generator -f
kubectl logs -l app=kafka-replicator -f
```

## 🔧 Configuración

### Variables de Entorno

Los servicios se configuran mediante variables de entorno. Ver archivos `.env.example` en cada servicio.

### Helm Values

Cada chart tiene su archivo `values.yaml` personalizable. Configuraciones principales en `k8s/config/`.

## 🚨 Troubleshooting

### Problemas Comunes

1. **Pods en estado Pending**: Verificar recursos del cluster
2. **Conexiones MQTT fallidas**: Verificar configuración de EMQX
3. **Kafka no disponible**: Verificar bootstrap servers
4. **Port-forward falla**: Verificar que los pods estén running

### Comandos Útiles

```bash
# Estado general
make status

# Logs de componentes específicos
kubectl logs -l app.kubernetes.io/name=kafka -n medisupply
kubectl logs -l app=mqtt-event-generator -n medilogistic

# Reiniciar servicios
kubectl rollout restart deployment/kafka-ui -n medisupply
```

## 🤝 Contribución

1. Fork el repositorio
2. Crear feature branch (`git checkout -b feature/nueva-funcionalidad`)
3. Commit cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push al branch (`git push origin feature/nueva-funcionalidad`)
5. Crear Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo `LICENSE` para más detalles.

## 📞 Soporte

Para soporte y preguntas:
- Crear un issue en GitHub
- Revisar la documentación en `k8s/README.md` y `k8s/ARCHITECTURE.md`
- Consultar los logs de los servicios para debugging
