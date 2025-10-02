# EMQX Chart - MediSupply EDA

Broker MQTT de alto rendimiento para la arquitectura Event-Driven de MediSupply. EMQX maneja la comunicación entre sensores IoT, generadores de eventos y clientes de procesamiento.

## 🎯 Propósito en MediSupply EDA

EMQX actúa como el broker MQTT central en el flujo de eventos:

```
mqtt-event-generator → EMQX → mqtt-order-event-client → EMQX → mqtt-kafka-bridge
```

### Funciones Principales

- **Recepción de eventos IoT**: Sensores de temperatura, humedad, daños
- **Distribución de eventos**: Routing a múltiples consumidores
- **Gestión de conexiones**: Manejo de clientes MQTT concurrentes
- **Dashboard de monitoreo**: Interfaz web para administración

## 🚀 Instalación

### Instalación en MediSupply

```bash
# Desde el directorio k8s
helm install emqx ./mqtt/emqx --namespace medilogistic --create-namespace

# O usando el Makefile
make deploy  # Incluye EMQX en el despliegue completo
```

### Verificación

```bash
# Verificar pods
kubectl get pods -l app.kubernetes.io/name=emqx -n medilogistic

# Verificar servicios
kubectl get svc -l app.kubernetes.io/name=emqx -n medilogistic

# Acceder al dashboard
kubectl port-forward svc/emqx 18083:18083 -n medilogistic
```

## ⚙️ Configuración MediSupply

### Topics MQTT Utilizados

| Topic | Propósito | Publisher | Subscriber |
|-------|-----------|-----------|------------|
| `events/sensor` | Eventos de sensores IoT | mqtt-event-generator | mqtt-order-event-client |
| `orders/events` | Eventos de pedidos | mqtt-order-event-client | mqtt-kafka-bridge |
| `damage/alerts` | Alertas de daños | Sensores | Sistema de alertas |

### Configuración de Producción

```yaml
replicaCount: 3

emqxConfig:
  EMQX_CLUSTER__DISCOVERY_STRATEGY: "k8s"
  EMQX_CLUSTER__K8S__SERVICE_NAME: "emqx-headless"
  EMQX_CLUSTER__K8S__NAMESPACE: "medilogistic"
  
  # Configuración MQTT
  EMQX_MQTT__MAX_PACKET_SIZE: "1MB"
  EMQX_MQTT__MAX_CLIENTID_LEN: 65535
  EMQX_MQTT__MAX_TOPIC_LEVELS: 128
  
  # Configuración de listeners
  EMQX_LISTENERS__TCP__DEFAULT__MAX_CONNECTIONS: 1024000
  EMQX_LISTENERS__WS__DEFAULT__MAX_CONNECTIONS: 102400

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi
```

## 🌐 Acceso y Puertos

### Puertos Estándar

| Puerto | Protocolo | Descripción |
|--------|-----------|-------------|
| 1883 | MQTT | Conexiones MQTT estándar |
| 8883 | MQTTS | MQTT sobre SSL/TLS |
| 8083 | WebSocket | MQTT sobre WebSocket |
| 8084 | WSS | MQTT sobre WebSocket Secure |
| 18083 | HTTP | Dashboard y API REST |

### Acceso al Dashboard

```bash
# Port-forward para acceso local
kubectl port-forward svc/emqx 18083:18083 -n medilogistic

# Acceder en: http://localhost:18083
# Usuario: admin
# Contraseña: public (por defecto)
```

## 🔧 Configuración Avanzada

### Autenticación y Autorización

```yaml
emqxConfig:
  # Autenticación por base de datos interna
  EMQX_AUTH__MNESIA__PASSWORD_HASH: "sha256"
  
  # Configuración de ACL
  EMQX_AUTHORIZATION__SOURCES: |
    [
      {
        "type": "built_in_database",
        "enable": true
      }
    ]
```

### Persistencia

```yaml
persistence:
  enabled: true
  storageClass: "standard"
  size: 10Gi
  accessMode: ReadWriteOnce
```

### Clustering

```yaml
emqxConfig:
  EMQX_CLUSTER__DISCOVERY_STRATEGY: "k8s"
  EMQX_CLUSTER__K8S__SERVICE_NAME: "emqx-headless"
  EMQX_CLUSTER__K8S__NAMESPACE: "medilogistic"
  EMQX_CLUSTER__K8S__ADDRESS_TYPE: "hostname"
  EMQX_CLUSTER__K8S__SUFFIX: "svc.cluster.local"
```

## 📊 Monitoreo

### Métricas Prometheus

```yaml
metrics:
  enabled: true
  type: "prometheus"

service:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "18083"
    prometheus.io/path: "/api/v5/prometheus/stats"
```

### Health Checks

```yaml
livenessProbe:
  httpGet:
    path: /status
    port: 18083
  initialDelaySeconds: 60
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /status
    port: 18083
  initialDelaySeconds: 10
  periodSeconds: 5
```

## 🔒 Seguridad

### SSL/TLS

```yaml
ssl:
  enabled: true
  useExisting: false
  commonName: "emqx.medilogistic.local"
  dnsnames:
    - "emqx.medilogistic.local"
    - "*.emqx.medilogistic.local"

emqxConfig:
  EMQX_LISTENERS__SSL__DEFAULT__SSL_OPTIONS__CERTFILE: "/tmp/ssl/tls.crt"
  EMQX_LISTENERS__SSL__DEFAULT__SSL_OPTIONS__KEYFILE: "/tmp/ssl/tls.key"
```

### Configuración de Usuarios

```bash
# Crear usuario via API
curl -X POST http://localhost:18083/api/v5/authentication/password_based:built_in_database/users \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "medisupply_client",
    "password": "secure_password"
  }'
```

## 🧪 Testing

### Conexión MQTT

```bash
# Instalar cliente MQTT
apt-get install mosquitto-clients

# Publicar mensaje de prueba
mosquitto_pub -h localhost -p 1883 -t "events/sensor" -m '{"temperature": 25.5, "humidity": 60}'

# Suscribirse a topic
mosquitto_sub -h localhost -p 1883 -t "events/sensor"
```

### Testing con Servicios MediSupply

```bash
# Verificar que mqtt-event-generator está publicando
kubectl logs -l app=mqtt-event-generator -n medilogistic

# Verificar que mqtt-order-event-client está recibiendo
kubectl logs -l app=mqtt-order-event-client -n medilogistic
```

## 🚨 Troubleshooting

### Problemas Comunes

1. **Pods no se inician**:
   ```bash
   kubectl describe pod -l app.kubernetes.io/name=emqx -n medilogistic
   kubectl logs -l app.kubernetes.io/name=emqx -n medilogistic
   ```

2. **Cluster no se forma**:
   ```bash
   # Verificar servicio headless
   kubectl get svc emqx-headless -n medilogistic
   
   # Verificar DNS interno
   kubectl exec -it emqx-0 -n medilogistic -- nslookup emqx-headless.medilogistic.svc.cluster.local
   ```

3. **Conexiones MQTT fallan**:
   ```bash
   # Verificar puertos
   kubectl port-forward svc/emqx 1883:1883 -n medilogistic
   
   # Test de conectividad
   telnet localhost 1883
   ```

### Logs y Debug

```bash
# Logs detallados
kubectl logs emqx-0 -n medilogistic -f

# Logs de todos los pods
kubectl logs -l app.kubernetes.io/name=emqx -n medilogistic --all-containers=true

# Configuración actual
kubectl exec emqx-0 -n medilogistic -- emqx ctl cluster status
```

## 📋 Configuración por Defecto

```yaml
replicaCount: 1

image:
  repository: emqx/emqx
  tag: "5.0"
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  mqtt: 1883
  mqttssl: 8883
  ws: 8083
  wss: 8084
  dashboard: 18083

emqxConfig:
  EMQX_CLUSTER__DISCOVERY_STRATEGY: "manual"
  EMQX_DASHBOARD__DEFAULT_USERNAME: "admin"
  EMQX_DASHBOARD__DEFAULT_PASSWORD: "public"

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

## 🔄 Integración con MediSupply

### Flujo de Datos

1. **mqtt-event-generator** → publica eventos a `events/sensor`
2. **mqtt-order-event-client** → suscribe a `events/sensor`, procesa y publica a `orders/events`
3. **mqtt-kafka-bridge** → suscribe a `orders/events` y envía a Kafka

### Configuración de Clientes

Los servicios MediSupply se configuran para conectar a EMQX:

```yaml
# En mqtt-event-generator
env:
  - name: MQTT_BROKER
    value: "tcp://emqx:1883"
  - name: MQTT_TOPIC
    value: "events/sensor"

# En mqtt-order-event-client  
env:
  - name: MQTT_BROKER
    value: "tcp://emqx:1883"
  - name: MQTT_TOPIC
    value: "events/sensor"
```

# Configuration

The following table lists the configurable parameters of the emqx chart and their default values.

| Parameter                            | Description                                                                                                                                                  | Default Value                                           |
|--------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------|
| `replicaCount`                       | It is recommended to have odd number of nodes in a cluster, otherwise the emqx cluster cannot be automatically healed in case of net-split.                  | 3                                                       |
| `image.repository`                   | EMQX Image name                                                                                                                                              | emqx/emqx                                               |
| `image.pullPolicy`                   | The image pull policy                                                                                                                                        | IfNotPresent                                            |
| `image.pullSecrets `                 | The image pull secrets                                                                                                                                       | `[]` (does not add image pull secrets to deployed pods) |
| `serviceAccount.create`              | If `true`, create a new service account                                                                                                                      | `true`                                                  |
| `serviceAccount.name`                | Service account to be used. If not set and `serviceAccount.create` is `true`, a name is generated using the full-name template                               |                                                         |
| `serviceAccount.annotations`         | Annotations to add to the service account                                                                                                                    |                                                         |
| `envFromSecret`                      | The name pull a secret in the same Kubernetes namespace which contains values that will be added to the environment                                          | nil                                                     |
| `recreatePods`                       | Forces the recreation of pods during upgrades, which can be useful to always apply the most recent configuration.                                            | false                                                   |
| `podAnnotations `                    | Annotations for pod                                                                                                                                          | `{}`                                                    |
| `podManagementPolicy`                | To redeploy a chart with existing PVC(s), the value must be set to Parallel to avoid deadlock                                                                | `Parallel`                                              |
| `persistence.enabled`                | Enable EMQX persistence using PVC                                                                                                                            | false                                                   |
| `persistence.storageClass`           | Storage class of backing PVC                                                                                                                                 | `nil` (uses alpha storage class annotation)             |
| `persistence.existingClaim`          | EMQX data Persistent Volume existing claim name, evaluated as a template                                                                                     | ""                                                      |
| `persistence.accessMode`             | PVC Access Mode for EMQX volume                                                                                                                              | ReadWriteOnce                                           |
| `persistence.size`                   | PVC Storage Request for EMQX volume                                                                                                                          | 20Mi                                                    |
| `initContainers`                     | Containers that run before the creation of EMQX containers. They can contain utilities or setup scripts.                                                     | `{}`                                                    |
| `resources`                          | CPU/Memory resource requests/limits                                                                                                                          | {}                                                      |
| `extraVolumeMounts`                  | Additional volumeMounts to the default backend container.                                                                                                    | []                                                      |
| `extraVolumes`                       | Additional volumes to the default backend pod.                                                                                                               | []                                                      |
| `nodeSelector`                       | Node labels for pod assignment                                                                                                                               | `{}`                                                    |
| `tolerations`                        | Toleration labels for pod assignment                                                                                                                         | `[]`                                                    |
| `affinity`                           | Map of node/pod affinities                                                                                                                                   | `{}`                                                    |
| `topologySpreadConstraints`          | List of topology spread constraints without labelSelector                                                                                                    | `[]`                                                    |
| `service.type`                       | Kubernetes Service type.                                                                                                                                     | ClusterIP                                               |
| `service.mqtt`                       | Port for MQTT.                                                                                                                                               | 1883                                                    |
| `service.mqttssl`                    | Port for MQTT(SSL).                                                                                                                                          | 8883                                                    |
| `service.ws`                         | Port for WebSocket/HTTP.                                                                                                                                     | 8083                                                    |
| `service.wss`                        | Port for WSS/HTTPS.                                                                                                                                          | 8084                                                    |
| `service.dashboard`                  | Port for dashboard and API.                                                                                                                                  | 18083                                                   |
| `service.nodePorts.mqtt`             | Kubernetes node port for MQTT.                                                                                                                               | nil                                                     |
| `service.nodePorts.mqttssl`          | Kubernetes node port for MQTT(SSL).                                                                                                                          | nil                                                     |
| `service.nodePorts.ws`               | Kubernetes node port for WebSocket/HTTP.                                                                                                                     | nil                                                     |
| `service.nodePorts.wss`              | Kubernetes node port for WSS/HTTPS.                                                                                                                          | nil                                                     |
| `service.nodePorts.dashboard`        | Kubernetes node port for dashboard.                                                                                                                          | nil                                                     |
| `service.loadBalancerClass`          | The load balancer implementation this Service belongs to                                                                                                     |                                                         |
| `service.loadBalancerIP`             | loadBalancerIP for Service                                                                                                                                   | nil                                                     |
| `service.loadBalancerSourceRanges`   | Address(es) that are allowed when service is LoadBalancer                                                                                                    | []                                                      |
| `service.externalIPs`                | ExternalIPs for the service                                                                                                                                  | []                                                      |
| `service.externalTrafficPolicy`      | External Traffic Policy for the service                                                                                                                      | `Cluster`                                               |
| `service.annotations`                | Service/ServiceMonitor annotations                                                                                                                           | {}(evaluated as a template)                             |
| `service.labels`                     | Service/ServiceMonitor labels                                                                                                                                | {}(evaluated as a template)                             |
| `ingress.dashboard.enabled`          | Enable ingress for EMQX Dashboard                                                                                                                            | false                                                   |
| `ingress.dashboard.ingressClassName` | Set the ingress class for EMQX Dashboard                                                                                                                     |                                                         |
| `ingress.dashboard.path`             | Ingress path for EMQX Dashboard                                                                                                                              | /                                                       |
| `ingress.dashboard.pathType`         | Ingress pathType for EMQX Dashboard                                                                                                                          | `ImplementationSpecific`                                |
| `ingress.dashboard.hosts`            | Ingress hosts for EMQX Dashboard                                                                                                                             | dashboard.emqx.local                                    |
| `ingress.dashboard.tls`              | Ingress tls for EMQX Dashboard                                                                                                                               | []                                                      |
| `ingress.dashboard.annotations`      | Ingress annotations for EMQX Dashboard                                                                                                                       | {}                                                      |
| `ingress.dashboard.ingressClassName` | Set the ingress class for EMQX Dashboard                                                                                                                     |                                                         |
| `ingress.mqtt.enabled`               | Enable ingress for MQTT                                                                                                                                      | false                                                   |
| `ingress.mqtt.ingressClassName`      | Set the ingress class for MQTT                                                                                                                               |                                                         |
| `ingress.mqtt.path`                  | Ingress path for MQTT                                                                                                                                        | /                                                       |
| `ingress.mqtt.pathType`              | Ingress pathType for MQTT                                                                                                                                    | `ImplementationSpecific`                                |
| `ingress.mqtt.hosts`                 | Ingress hosts for MQTT                                                                                                                                       | mqtt.emqx.local                                         |
| `ingress.mqtt.tls`                   | Ingress tls for MQTT                                                                                                                                         | []                                                      |
| `ingress.mqtt.annotations`           | Ingress annotations for MQTT                                                                                                                                 | {}                                                      |
| `ingress.mqtt.ingressClassName`      | Set the ingress class for MQTT                                                                                                                               |                                                         |
| `metrics.enable`                     | If set to true, [prometheus-operator](https://github.com/prometheus-operator/prometheus-operator) needs to be installed, and emqx_prometheus needs to enable | false                                                   |
| `metrics.type`                       | Now we only supported "prometheus"                                                                                                                           | "prometheus"                                            |
| `ssl.enabled`                        | Enable SSL support                                                                                                                                           | false                                                   |
| `ssl.useExisting`                    | Use existing certificate or let cert-manager generate one                                                                                                    | false                                                   |
| `ssl.existingName`                   | Name of existing certificate                                                                                                                                 | emqx-tls                                                |
| `ssl.commonName`                     | Common name for or certificate to be generated                                                                                                               |                                                         |
| `ssl.dnsnames`                       | DNS name(s) for certificate to be generated                                                                                                                  | {}                                                      |
| `ssl.issuer.name`                    | Issuer name for certificate generation                                                                                                                       | letsencrypt-dns                                         |
| `ssl.issuer.kind`                    | Issuer kind for certificate generation                                                                                                                       | ClusterIssuer                                           |

## EMQX specific settings

The following table lists the configurable [EMQX](https://www.emqx.io/)-specific parameters of the chart and their
default values.
| Parameter                                                                                                                                                              | Description                                                                   | Default Value |
|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|---------------|
| `emqxConfig`                                                                                                                                                           | Map of [configuration](https://www.emqx.io/docs/en/v5.0/admin/cfg.html) items |               |
| expressed as [environment variables](https://www.emqx.io/docs/en/v5.0/admin/cfg.html#environment-variables) (prefix `EMQX_` can be omitted) or using the configuration |                                                                               |               |
| files [namespaced dotted notation](https://www.emqx.io/docs/en/v5.0/admin/cfg.html#syntax)                                                                             | `nil`                                                                         |               |
| `emqxLicenseSecretName`                                                                                                                                                | Name of the secret that holds the license information                         | `nil`         |

## SSL settings
`cert-manager` generates secrets with certificate data using the keys `tls.crt` and `tls.key`. The helm chart always mounts those keys as files to `/tmp/ssl/`
which needs to explicitly configured by either changing the emqx config file or by passing the following environment variables:

```
  EMQX_LISTENERS__SSL__DEFAULT__SSL_OPTIONS__CERTFILE: /tmp/ssl/tls.crt
  EMQX_LISTENERS__SSL__DEFAULT__SSL_OPTIONS__KEYFILE: /tmp/ssl/tls.key
```

If you chose to use an existing certificate, make sure, you update the filenames accordingly.

## Tips
Enable the Proxy Protocol V1/2 if the EMQX cluster is deployed behind HAProxy or Nginx.
In order to preserve the original client's IP address, you could change the emqx config by passing the following environment variable:

```
EMQX_LISTENERS__TCP__DEFAULT__PROXY_PROTOCOL: "true"
```

With HAProxy you'd also need the following ingress annotation:

```
haproxy-ingress.github.io/proxy-protocol: "v2"
```
