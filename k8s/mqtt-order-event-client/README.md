# MQTT Order Event Client Helm Chart

Este chart despliega el servicio MQTT Order Event Client en Kubernetes.

## Descripción

El MQTT Order Event Client es un servicio que se suscribe a eventos publicados por `mqtt-event-generator` a través de un broker MQTT (EMQX) y expone una API HTTP para consultar los eventos recibidos y estadísticas básicas.

## Instalación

```bash
# Desde la carpeta k8s
helm upgrade --install mqtt-order-event-client ./mqtt-order-event-client \
  --namespace medisupply \
  --create-namespace
```

## Configuración

### Valores principales

| Parámetro | Descripción | Valor por defecto |
|-----------|-------------|-------------------|
| `replicaCount` | Número de réplicas | `1` |
| `image.repository` | Repositorio de la imagen | `mqtt-order-event-client` |
| `image.tag` | Tag de la imagen | `latest` |
| `mqtt.broker` | URL del broker MQTT | `tcp://emqx.medisupply.svc.cluster.local:1883` |
| `mqtt.clientId` | ID del cliente MQTT | `order-event-client` |
| `mqtt.topic` | Topic a suscribir | `events/sensor` |
| `mqtt.username` | Usuario MQTT | `admin` |
| `mqtt.password` | Contraseña MQTT | `public` |
| `service.port` | Puerto HTTP expuesto por el servicio | `8080` |

### Ejemplo de configuración personalizada

```yaml
# custom-values.yaml
replicaCount: 2

mqtt:
  broker: "tcp://my-mqtt-broker:1883"
  topic: "sensors/temperature"
  username: "myuser"
  password: "mypassword"

service:
  port: 9090

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

```bash
helm upgrade --install mqtt-order-event-client ./mqtt-order-event-client \
  --namespace medisupply \
  --values custom-values.yaml
```

## Monitoreo

### Ver logs
```bash
kubectl logs -f deployment/mqtt-order-event-client -n medisupply
```

### Ver estado
```bash
kubectl get pods -l app.kubernetes.io/name=mqtt-order-event-client -n medisupply
```

### Ver API
```bash
kubectl port-forward svc/mqtt-order-event-client 8080:8080 -n medisupply
curl http://localhost:8080/health
curl http://localhost:8080/events
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

- **EMQX**: Debe estar desplegado en el mismo namespace (`medisupply`)
- **Namespace**: Con inyección de Istio habilitada (opcional)
