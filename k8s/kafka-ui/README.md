# Kafka UI Chart

Interfaz web para gestión y monitoreo de clusters Apache Kafka. Proporciona una interfaz gráfica intuitiva para administrar topics, consumer groups, mensajes y configuraciones de Kafka.

## 🎯 Características

- **Gestión de Topics**: Crear, eliminar y configurar topics
- **Monitoreo de Consumer Groups**: Ver lag, offsets y estado de consumidores
- **Explorador de Mensajes**: Producir y consumir mensajes en tiempo real
- **Métricas del Cluster**: Estadísticas de brokers, particiones y throughput
- **Configuración de Brokers**: Ver y modificar configuraciones
- **Soporte Multi-Cluster**: Gestionar múltiples clusters desde una interfaz

## 🚀 Instalación

### Instalación Básica

```bash
helm install kafka-ui ./kafka-ui --namespace medisupply --create-namespace
```

### Con Configuración Personalizada

```bash
helm install kafka-ui ./kafka-ui \
  --namespace medisupply \
  --set kafka.clusters[0].name="Mi Cluster" \
  --set kafka.clusters[0].bootstrapServers="kafka:9092"
```

## ⚙️ Configuración

### Configuración de Clusters Kafka

```yaml
kafka:
  clusters:
    - name: "Kafka Principal"
      bootstrapServers: "kafka:9092"
      properties: {}
    - name: "Kafka Warehouse"
      bootstrapServers: "kafka-warehouse:9092"
      properties: {}
```

### Configuración de Servicio

```yaml
service:
  type: ClusterIP
  port: 8080
  annotations: {}

ingress:
  enabled: false
  className: ""
  annotations: {}
  hosts:
    - host: kafka-ui.local
      paths:
        - path: /
          pathType: Prefix
  tls: []
```

### Recursos y Escalabilidad

```yaml
replicaCount: 1

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 256Mi

autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 3
  targetCPUUtilizationPercentage: 80
```

## 🔧 Configuración Avanzada

### Autenticación y Seguridad

```yaml
kafka:
  clusters:
    - name: "Kafka Seguro"
      bootstrapServers: "kafka:9092"
      properties:
        security.protocol: SASL_SSL
        sasl.mechanism: PLAIN
        sasl.jaas.config: 'org.apache.kafka.common.security.plain.PlainLoginModule required username="user" password="password";'
```

### Variables de Entorno

```yaml
env:
  - name: KAFKA_CLUSTERS_0_NAME
    value: "local"
  - name: KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS
    value: "kafka:9092"
  - name: LOGGING_LEVEL_COM_PROVECTUS
    value: "DEBUG"
```

### Configuración de JVM

```yaml
jvm:
  options: "-Xms256m -Xmx512m"
```

## 🌐 Acceso

### Port Forward Local

```bash
kubectl port-forward svc/kafka-ui 9090:8080 -n medisupply
```

Acceder en: http://localhost:9090

### Ingress

```yaml
ingress:
  enabled: true
  className: "nginx"
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
  hosts:
    - host: kafka-ui.example.com
      paths:
        - path: /
          pathType: Prefix
```

## 📊 Funcionalidades

### Gestión de Topics

- **Crear Topics**: Configurar particiones, factor de replicación y configuraciones
- **Explorar Mensajes**: Ver contenido de mensajes con filtros y búsqueda
- **Producir Mensajes**: Enviar mensajes de prueba con headers personalizados
- **Configuración**: Modificar configuraciones de topics en tiempo real

### Monitoreo de Consumer Groups

- **Estado de Grupos**: Ver todos los consumer groups activos
- **Lag Monitoring**: Monitorear retraso de consumidores por partición
- **Reset Offsets**: Reiniciar offsets de consumer groups
- **Miembros del Grupo**: Ver consumidores activos y sus asignaciones

### Administración de Brokers

- **Estado del Cluster**: Ver salud y métricas de brokers
- **Configuraciones**: Explorar configuraciones de brokers
- **Logs**: Acceder a logs de Kafka (si está configurado)

## 🔍 Monitoreo

### Health Checks

```yaml
livenessProbe:
  httpGet:
    path: /actuator/health
    port: http
  initialDelaySeconds: 30
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /actuator/health
    port: http
  initialDelaySeconds: 5
  periodSeconds: 10
```

### Métricas

Kafka UI expone métricas en formato Prometheus:

```yaml
serviceMonitor:
  enabled: true
  interval: 30s
  path: /actuator/prometheus
```

## 🚨 Troubleshooting

### Problemas Comunes

1. **No se conecta a Kafka**:
   ```bash
   # Verificar configuración de bootstrap servers
   kubectl logs deployment/kafka-ui -n medisupply
   
   # Verificar conectividad de red
   kubectl exec -it deployment/kafka-ui -n medisupply -- nc -zv kafka 9092
   ```

2. **Interfaz no carga**:
   ```bash
   # Verificar estado del pod
   kubectl get pods -l app.kubernetes.io/name=kafka-ui -n medisupply
   
   # Verificar logs
   kubectl logs -l app.kubernetes.io/name=kafka-ui -n medisupply
   ```

3. **Problemas de memoria**:
   ```yaml
   # Aumentar recursos
   resources:
     limits:
       memory: 1Gi
     requests:
       memory: 512Mi
   ```

### Debugging

```bash
# Habilitar logs debug
helm upgrade kafka-ui ./kafka-ui \
  --set env[0].name=LOGGING_LEVEL_ROOT \
  --set env[0].value=DEBUG \
  --namespace medisupply
```

## 🔧 Desarrollo

### Configuración Local

Para desarrollo local con múltiples clusters:

```yaml
kafka:
  clusters:
    - name: "Local Kafka"
      bootstrapServers: "localhost:9092"
    - name: "Docker Kafka"
      bootstrapServers: "kafka:9092"
    - name: "Warehouse Kafka"
      bootstrapServers: "kafka-warehouse:9092"
```

### Testing

```bash
# Verificar conectividad
kubectl run kafka-test --image=confluentinc/cp-kafka:latest --rm -it -- bash
kafka-topics --bootstrap-server kafka:9092 --list
```

## 📋 Valores por Defecto

```yaml
replicaCount: 1

image:
  repository: provectuslabs/kafka-ui
  pullPolicy: IfNotPresent
  tag: "latest"

service:
  type: ClusterIP
  port: 8080

kafka:
  clusters:
    - name: "kafka"
      bootstrapServers: "kafka:9092"

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 256Mi
```

## 🤝 Contribución

1. Modificar templates en `templates/`
2. Actualizar `values.yaml` con nuevas configuraciones
3. Probar cambios con `helm template`
4. Documentar cambios en este README

## 📄 Licencia

Este chart está bajo la Licencia MIT.