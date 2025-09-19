# MediSupply EDA - Kubernetes Infrastructure

Infraestructura Kubernetes para el sistema MediSupply con arquitectura Event-Driven usando Istio Service Mesh, Apache Kafka y KEDA.

## 🏗️ Arquitectura

- **Istio Service Mesh**: Comunicación segura entre servicios
- **Apache Kafka**: Sistema de mensajería para eventos
- **KEDA**: Autoescalado basado en eventos
- **Kafka UI**: Interfaz gráfica para gestión de Kafka

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

# Desplegar Kafka y Kafka UI
make deploy
```

### Comandos Disponibles

```bash
make help                    # Mostrar ayuda
make init                    # Crear cluster con Kind (default)
make init PROVIDER=minikube  # Crear cluster con Minikube
make deploy                  # Desplegar Kafka y Kafka UI
make kafka-ui                # Abrir Kafka UI (http://localhost:9090)
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
| Kafka UI | http://localhost:9090 | Interfaz de gestión de Kafka |
| Kiali | http://localhost:20001 | Observabilidad de Istio |

## 📁 Estructura del Proyecto

```
k8s/
├── istio/           # Charts de Istio
│   ├── base/        # Componentes base
│   ├── istiod/      # Plano de control
│   └── gateway/     # Gateway de entrada
├── kafka/           # Chart de Apache Kafka
├── kafka-ui/        # Chart de Kafka UI
├── keda/            # Chart de KEDA
├── config/          # Configuraciones
│   ├── kind-config.yaml
│   ├── minikube-config.yaml
│   └── kafka-values.yaml
├── Makefile         # Comandos de gestión
└── README.md        # Este archivo
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
