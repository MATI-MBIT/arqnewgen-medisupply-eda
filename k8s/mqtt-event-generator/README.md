# MQTT Event Generator Helm Chart

Este chart despliega el servicio MQTT Event Generator en Kubernetes.

## Descripción

El MQTT Event Generator es un servicio que genera eventos simulados de sensores y los publica a un broker MQTT (EMQX) cada cierto intervalo de tiempo configurable.

## Instalación

```bash
# Desde la carpeta k8s
helm upgrade --install mqtt-event-generator ./mqtt-event-generator \
  --namespace medisupply \
  --create-namespace
```

## Configuración

### Valores principales

| Parámetro | Descripción | Valor por defecto |
|-----------|-------------|-------------------|
| `replicaCount` | Número de réplicas | `1` |
| `image.repository` | Repositorio de la imagen | `mqtt-event-generator` |
| `image.tag` | Tag de la imagen | `latest` |
| `mqtt.broker` | URL del broker MQTT | `tcp://emqx:1883` |
| `mqtt.clientId` | ID del cliente MQTT | `event-generator-k8s` |
| `mqtt.topic` | Topic donde publicar | `events/sensor` |
| `mqtt.username` | Usuario MQTT | `admin` |
| `mqtt.password` | Contraseña MQTT | `public` |
| `eventGenerator.intervalSeconds` | Frecuencia de eventos (segundos) | `30` |

### Ejemplo de configuración personalizada

```yaml
# custom-values.yaml
replicaCount: 2

mqtt:
  broker: "tcp://my-mqtt-broker:1883"
  topic: "sensors/temperature"
  username: "myuser"
  password: "mypassword"

eventGenerator:
  intervalSeconds: 10

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

```bash
helm upgrade --install mqtt-event-generator ./mqtt-event-generator \
  --namespace medisupply \
  --values custom-values.yaml
```

## Monitoreo

### Ver logs
```bash
kubectl logs -f deployment/mqtt-event-generator -n medisupply
```

### Ver estado
```bash
kubectl get pods -l app.kubernetes.io/name=mqtt-event-generator -n medisupply
```

### Ver eventos generados
Conectarse al broker MQTT y suscribirse al topic configurado para ver los eventos.

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

- **EMQX**: Debe estar desplegado en el mismo namespace (`medisupply`)
- **Namespace**: Con inyección de Istio habilitada (opcional)

### Verificar conectividad con EMQX

```bash
# Verificar que EMQX esté ejecutándose
kubectl get pods -l app.kubernetes.io/name=emqx -n medisupply

# Verificar el servicio EMQX
kubectl get svc emqx -n medisupply

# Probar conectividad desde el pod del event generator
kubectl exec -it deployment/mqtt-event-generator -n medisupply -- \
  sh -c 'nc -zv emqx.medisupply.svc.cluster.local 1883'
```

### Configuración de EMQX

El chart asume que EMQX está configurado con:
- **Usuario**: `admin`
- **Contraseña**: `public`
- **Puerto MQTT**: `1883`
- **Servicio**: `emqx.medisupply.svc.cluster.local`