# Configuraciones MediSupply EDA

Directorio de configuraciones personalizadas para los charts de Helm en la arquitectura Event-Driven de MediSupply. Contiene archivos de valores específicos para diferentes entornos y componentes.

## 📁 Estructura de Configuraciones

```
k8s/config/
├── services/                    # Configuraciones de microservicios
│   ├── mqtt-event-generator-values.yaml
│   └── mqtt-order-event-client-values.yaml
├── kafka-values.yaml           # Cluster Kafka principal
├── kafka-warehouse-values.yaml # Cluster Kafka warehouse
├── kind-config.yaml           # Configuración Kind
└── minikube-config.yaml       # Configuración Minikube
```

## 📁 Archivos de Configuración

### Clusters Locales

| Archivo | Descripción | Uso |
|---------|-------------|-----|
| `kind-config.yaml` | Configuración para cluster Kind | Desarrollo local con Docker |
| `minikube-config.yaml` | Configuración para cluster Minikube | Desarrollo local con VM |

### Kafka Clusters

| Archivo | Descripción | Namespace | Puerto |
|---------|-------------|-----------|--------|
| `kafka-values.yaml` | Cluster Kafka principal | `medisupply` | 9092 |
| `kafka-warehouse-values.yaml` | Cluster Kafka warehouse | `mediwarehouse` | 9092 |

### Servicios

Los archivos de configuración de servicios están organizados en el subdirectorio `services/`:

| Archivo | Descripción | Componente |
|---------|-------------|------------|
| `services/mqtt-event-generator-values.yaml` | Configuración del generador MQTT | mqtt-event-generator |
| `services/mqtt-order-event-client-values.yaml` | Configuración del cliente MQTT de órdenes | mqtt-order-event-client |

## 🚀 Uso de Configuraciones

### Clusters Locales

```bash
# Crear cluster con Kind
make init PROVIDER=kind
# Usa automáticamente kind-config.yaml

# Crear cluster con Minikube  
make init PROVIDER=minikube
# Usa automáticamente minikube-config.yaml
```

### Despliegue con Configuraciones Específicas

```bash
# Kafka principal
helm install kafka ./kafka \
  --namespace medisupply \
  --values ./config/kafka-values.yaml

# Kafka warehouse
helm install kafka-warehouse ./kafka \
  --namespace mediwarehouse \
  --values ./config/kafka-warehouse-values.yaml

# Microservicios (usando el chart genérico)
helm install mqtt-event-generator ./microservice \
  --namespace medilogistic \
  --values ./config/services/mqtt-event-generator-values.yaml

helm install mqtt-order-event-client ./microservice \
  --namespace medilogistic \
  --values ./config/services/mqtt-order-event-client-values.yaml
```

### Usando el Makefile (Recomendado)

```bash
# Desplegar todos los servicios MQTT
make deploy-mqtt-services

# O individualmente
make deploy-mqtt-event-generator
make deploy-mqtt-order-client
```

## ⚙️ Configuraciones Detalladas

### Kind Configuration

```yaml
# kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
```

**Características:**
- Expone puertos 80/443 para ingress
- Configuración optimizada para desarrollo
- Soporte para LoadBalancer con MetalLB

### Minikube Configuration

```yaml
# minikube-config.yaml
# Configuración aplicada via comandos en Makefile
cpus: 2
memory: 6144
disk-size: 20g
driver: docker
addons:
  - ingress
  - dashboard
  - metrics-server
```

**Características:**
- 2 CPUs, 6GB RAM, 20GB disco
- Addons preinstalados
- Driver Docker por defecto

### Kafka Principal

```yaml
# kafka-values.yaml (extracto)
controller:
  replicaCount: 1
  
listeners:
  client:
    protocol: PLAINTEXT
  controller:
    protocol: PLAINTEXT
  interbroker:
    protocol: PLAINTEXT

service:
  ports:
    client: 9092

persistence:
  enabled: false  # Para desarrollo

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi
```

**Características:**
- Configuración para desarrollo (sin persistencia)
- Protocolo PLAINTEXT (sin autenticación)
- Factor de replicación 1
- Recursos optimizados para clusters locales

### Kafka Warehouse

```yaml
# kafka-warehouse-values.yaml (extracto)
controller:
  replicaCount: 1

service:
  ports:
    client: 9092  # Mismo puerto, diferente namespace

# Configuración específica para warehouse
extraConfig: |
  auto.create.topics.enable=true
  default.replication.factor=1
  min.insync.replicas=1
```

**Características:**
- Configuración similar al principal
- Namespace separado (`mediwarehouse`)
- Auto-creación de topics habilitada
- Optimizado para replicación

## 🔧 Personalización

### Crear Configuración Personalizada

```bash
# Copiar configuración base
cp config/kafka-values.yaml config/kafka-production.yaml

# Editar para producción
vim config/kafka-production.yaml
```

### Configuración de Producción

```yaml
# kafka-production.yaml
controller:
  replicaCount: 3

persistence:
  enabled: true
  size: 100Gi
  storageClass: "fast-ssd"

listeners:
  client:
    protocol: SASL_SSL
  
auth:
  sasl:
    enabled: true
    mechanism: SCRAM-SHA-512

resources:
  limits:
    cpu: 2000m
    memory: 4Gi
  requests:
    cpu: 1000m
    memory: 2Gi
```

### Configuración por Entorno

```bash
# Desarrollo
helm install kafka ./kafka -f config/kafka-values.yaml

# Staging
helm install kafka ./kafka -f config/kafka-staging.yaml

# Producción
helm install kafka ./kafka -f config/kafka-production.yaml
```

## 📊 Configuraciones Recomendadas

### Desarrollo Local

```yaml
# Recursos mínimos
resources:
  requests:
    cpu: 100m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi

# Sin persistencia
persistence:
  enabled: false

# Configuración simple
auth:
  enabled: false
```

### Staging

```yaml
# Recursos medios
resources:
  requests:
    cpu: 500m
    memory: 1Gi
  limits:
    cpu: 1000m
    memory: 2Gi

# Persistencia temporal
persistence:
  enabled: true
  size: 20Gi

# Autenticación básica
auth:
  enabled: true
```

### Producción

```yaml
# Recursos completos
resources:
  requests:
    cpu: 1000m
    memory: 2Gi
  limits:
    cpu: 2000m
    memory: 4Gi

# Persistencia completa
persistence:
  enabled: true
  size: 100Gi
  storageClass: "fast-ssd"

# Seguridad completa
auth:
  enabled: true
  tls:
    enabled: true
```

## 🧹 Archivos Obsoletos

Los siguientes archivos pueden estar obsoletos y deberían revisarse para eliminación:

- `mqtt-generator-values.yaml` (raíz) - Reemplazado por `services/mqtt-event-generator-values.yaml`

**Verificar antes de eliminar:**
```bash
# Buscar referencias al archivo obsoleto
grep -r "mqtt-generator-values.yaml" .
```

## 🚨 Troubleshooting

### Problemas Comunes

1. **Configuración no se aplica**:
   ```bash
   # Verificar que el archivo existe
   ls -la config/
   
   # Verificar sintaxis YAML
   yamllint config/kafka-values.yaml
   ```

2. **Recursos insuficientes**:
   ```bash
   # Verificar recursos del cluster
   kubectl top nodes
   kubectl describe nodes
   
   # Ajustar recursos en configuración
   vim config/kafka-values.yaml
   ```

3. **Conflictos de puertos**:
   ```bash
   # Verificar puertos en uso
   kubectl get svc -A
   
   # Cambiar puertos en configuración
   vim config/kafka-warehouse-values.yaml
   ```

## 📋 Plantillas de Configuración

### Nueva Configuración de Servicio

```yaml
# template-service-values.yaml
replicaCount: 1

image:
  repository: my-service
  tag: latest
  pullPolicy: IfNotPresent

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

env:
  - name: SERVICE_ENV
    value: "development"

nodeSelector: {}
tolerations: []
affinity: {}
```

### Configuración de Cluster

```yaml
# template-cluster-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
  - containerPort: 443
    hostPort: 443
- role: worker
- role: worker
```

## 🤝 Contribución

1. Crear nueva configuración basada en plantilla
2. Probar configuración en entorno local
3. Documentar cambios específicos
4. Actualizar este README si es necesario
5. Crear PR con los cambios

## 📄 Licencia

Configuraciones bajo la misma licencia del proyecto principal.