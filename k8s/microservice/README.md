# Microservice Helm Chart

Un chart genérico de Helm para desplegar microservicios siguiendo el patrón DRY (Don't Repeat Yourself).

## Descripción

Este chart está diseñado para ser reutilizable y configurable, permitiendo el despliegue de cualquier microservicio con configuraciones específicas a través de valores personalizados.

## Características

- ✅ Deployment configurable
- ✅ Service con múltiples tipos
- ✅ ServiceAccount opcional
- ✅ HorizontalPodAutoscaler (HPA)
- ✅ Ingress configurable
- ✅ ConfigMap y Secret opcionales
- ✅ Health checks configurables
- ✅ Soporte para MQTT (opcional)
- ✅ Variables de entorno flexibles
- ✅ Soporte para Istio

## Instalación

```bash
helm install my-microservice ./k8s/microservice
```

## Configuración

### Valores básicos

```yaml
# values.yaml
image:
  repository: "my-microservice"
  tag: "v1.0.0"

service:
  port: 8080
  targetPort: 8080

replicaCount: 2
```

### Ejemplo para MQTT Event Generator

```yaml
# values-mqtt-event-generator.yaml
image:
  repository: mqtt-event-generator
  tag: latest

mqtt:
  enabled: true
  broker: "tcp://emqx.medilogistic.svc.cluster.local:1883"
  clientId: "event-generator-k8s"
  topic: "events/sensor"
  username: "admin"
  password: "public"

app:
  config:
    EVENT_INTERVAL_SECONDS: "30"

livenessProbe:
  initialDelaySeconds: 30
  periodSeconds: 30

readinessProbe:
  initialDelaySeconds: 5
  periodSeconds: 10
```

### Ejemplo para MQTT Order Event Client

```yaml
# values-mqtt-order-event-client.yaml
image:
  repository: mqtt-order-event-client
  tag: latest

mqtt:
  enabled: true
  broker: "tcp://emqx.medilogistic.svc.cluster.local:1883"
  clientId: "order-event-client"
  topic: "events/sensor"
  username: "admin"
  password: "public"

livenessProbe:
  initialDelaySeconds: 10
  periodSeconds: 15

readinessProbe:
  initialDelaySeconds: 5
  periodSeconds: 10
```

## Uso con diferentes microservicios

### 1. MQTT Event Generator

```bash
helm install mqtt-event-generator ./k8s/microservice -f values-mqtt-event-generator.yaml
```

### 2. MQTT Order Event Client

```bash
helm install mqtt-order-event-client ./k8s/microservice -f values-mqtt-order-event-client.yaml
```

### 3. API REST genérico

```yaml
# values-api-service.yaml
image:
  repository: my-api-service
  tag: v1.0.0

service:
  port: 3000
  targetPort: 3000

ingress:
  enabled: true
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix

env:
  custom:
    - name: DATABASE_URL
      valueFrom:
        secretKeyRef:
          name: db-secret
          key: url
    - name: API_KEY
      value: "my-api-key"
```

## Configuración avanzada

### Variables de entorno

```yaml
env:
  # Variables comunes (siempre incluidas)
  common:
    - name: POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
  
  # Variables personalizadas
  custom:
    - name: CUSTOM_VAR
      value: "custom-value"
    - name: SECRET_VAR
      valueFrom:
        secretKeyRef:
          name: my-secret
          key: secret-key

# Configuración de aplicación (se convierte en variables de entorno)
app:
  config:
    HTTP_PORT: "8080"
    LOG_LEVEL: "info"
```

### ConfigMap y Secrets

```yaml
configMap:
  enabled: true
  data:
    config.yaml: |
      server:
        port: 8080
      logging:
        level: info

secret:
  enabled: true
  data:
    password: bXlwYXNzd29yZA==  # base64 encoded
```

### Autoscaling

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

## Parámetros de configuración

| Parámetro | Descripción | Valor por defecto |
|-----------|-------------|-------------------|
| `image.repository` | Repositorio de la imagen | `""` |
| `image.tag` | Tag de la imagen | `"latest"` |
| `image.pullPolicy` | Política de pull de imagen | `IfNotPresent` |
| `replicaCount` | Número de réplicas | `1` |
| `service.type` | Tipo de servicio | `ClusterIP` |
| `service.port` | Puerto del servicio | `8080` |
| `service.targetPort` | Puerto del contenedor | `8080` |
| `mqtt.enabled` | Habilitar configuración MQTT | `false` |
| `mqtt.broker` | URL del broker MQTT | `""` |
| `autoscaling.enabled` | Habilitar HPA | `false` |
| `ingress.enabled` | Habilitar Ingress | `false` |
| `configMap.enabled` | Crear ConfigMap | `false` |
| `secret.enabled` | Crear Secret | `false` |

## Migración desde charts específicos

Para migrar desde los charts `mqtt-event-generator` o `mqtt-order-event-client`:

1. Crea un archivo de valores específico basado en los ejemplos anteriores
2. Desinstala el chart anterior: `helm uninstall <release-name>`
3. Instala usando el nuevo chart: `helm install <release-name> ./k8s/microservice -f <values-file>`

## Contribución

Para añadir nuevas características al chart genérico, asegúrate de:

1. Mantener la compatibilidad hacia atrás
2. Hacer las nuevas características opcionales
3. Documentar los nuevos parámetros
4. Probar con diferentes tipos de microservicios