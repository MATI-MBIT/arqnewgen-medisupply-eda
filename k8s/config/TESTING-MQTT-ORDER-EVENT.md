# 🧪 Guía de Pruebas Manuales para MQTT Order Event Client Scaling

## 📋 Resumen de la Configuración

El `mqtt-order-event-client` se escala automáticamente basado en mensajes MQTT en el topic `events/sensor` usando KEDA.

### 🔧 Configuración KEDA:
- **Trigger**: MQTT (EMQX broker)
- **Topic**: `events/sensor`
- **Umbral**: >10 mensajes pendientes
- **Escalado**: 1-3 pods
- **Broker**: `emqx.medilogistic.svc.cluster.local:1883`

## 🚀 **Método 1: Script Automatizado (Recomendado)**

### Instalar KEDA Scaling:
```bash
# Hacer ejecutable e instalar
chmod +x k8s/config/install-keda-mqtt-order-event.sh
./k8s/config/install-keda-mqtt-order-event.sh
```

### Probar el escalado:
```bash
# Hacer ejecutable y ejecutar prueba
chmod +x k8s/scripts/test-mqtt-order-event-scaling.sh
./k8s/scripts/test-mqtt-order-event-scaling.sh
```

## 🔧 **Método 2: Pruebas Manuales Paso a Paso**

### Paso 1: Instalar KEDA Scaling
```bash
# Aplicar configuración de KEDA
kubectl apply -f k8s/config/keda-mqtt-order-event-client.yaml

# Verificar instalación
kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply
kubectl describe scaledobject mqtt-order-event-client-scaler -n medisupply
```

### Paso 2: Verificar Estado Inicial
```bash
# Ver pods actuales
kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client

# Ver estado de EMQX
kubectl get pods -n medisupply -l app.kubernetes.io/name=emqx

# Ver ScaledObject
kubectl get scaledobject -n medisupply
```

### Paso 3: Crear Cliente MQTT Temporal
```bash
# Crear pod temporal para enviar mensajes MQTT
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: mqtt-test-client
  namespace: medisupply
spec:
  containers:
  - name: mqtt-publisher
    image: eclipse-mosquitto:2.0
    command: ["/bin/sh"]
    args: ["-c", "apk add --no-cache mosquitto-clients && sleep 3600"]
  restartPolicy: Never
EOF

# Esperar a que esté listo
kubectl wait --for=condition=Ready pod/mqtt-test-client -n medisupply --timeout=60s
```

### Paso 4: Enviar Mensajes MQTT para Activar Escalado

#### Opción A: Enviar mensajes individuales
```bash
# Enviar un mensaje de prueba
kubectl exec -n medisupply mqtt-test-client -- mosquitto_pub \
  -h emqx.medilogistic.svc.cluster.local \
  -p 1883 \
  -t "events/sensor" \
  -m '{"id":"test-1","timestamp":"2024-01-01T12:00:00Z","sensor":"temperature","value":25,"location":"warehouse-a"}' \
  -u admin \
  -P public \
  -q 1

# Enviar múltiples mensajes (repetir para activar escalado)
for i in {1..15}; do
  kubectl exec -n medisupply mqtt-test-client -- mosquitto_pub \
    -h emqx.medilogistic.svc.cluster.local \
    -p 1883 \
    -t "events/sensor" \
    -m "{\"id\":\"test-$i\",\"timestamp\":\"$(date -Iseconds)\",\"sensor\":\"temperature\",\"value\":$((20 + RANDOM % 10)),\"location\":\"warehouse-a\"}" \
    -u admin \
    -P public \
    -q 1
  echo "Mensaje $i enviado"
  sleep 0.5
done
```

#### Opción B: Script automatizado de envío
```bash
# Crear script de envío masivo
cat <<'EOF' > /tmp/send_messages.sh
#!/bin/bash
for i in {1..20}; do
  kubectl exec -n medisupply mqtt-test-client -- mosquitto_pub \
    -h emqx.medilogistic.svc.cluster.local \
    -p 1883 \
    -t "events/sensor" \
    -m "{\"id\":\"load-test-$i\",\"timestamp\":\"$(date -Iseconds)\",\"sensor\":\"temperature\",\"value\":$((20 + RANDOM % 10)),\"location\":\"warehouse-a\"}" \
    -u admin \
    -P public \
    -q 1
  echo "Mensaje $i enviado"
  sleep 0.2
done
EOF

chmod +x /tmp/send_messages.sh
/tmp/send_messages.sh
```

### Paso 5: Monitorear el Escalado

#### Monitoreo en tiempo real:
```bash
# Ver pods escalando
kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client -w

# Ver ScaledObject en tiempo real
kubectl get scaledobject -n medisupply -w

# Ver logs de KEDA
kubectl logs -n keda -l app=keda-operator -f | grep mqtt-order-event
```

#### Verificar métricas:
```bash
# Ver estado del ScaledObject
kubectl describe scaledobject mqtt-order-event-client-scaler -n medisupply

# Ver eventos relacionados
kubectl get events -n medisupply --field-selector involvedObject.name=mqtt-order-event-client-scaler
```

### Paso 6: Verificar Procesamiento de Mensajes

```bash
# Ver logs del mqtt-order-event-client
kubectl logs -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client -f

# Verificar que los mensajes se están procesando
kubectl logs -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client | grep "Processing message"
```

### Paso 7: Limpiar Recursos de Prueba

```bash
# Eliminar pod de prueba
kubectl delete pod mqtt-test-client -n medisupply

# Eliminar script temporal
rm -f /tmp/send_messages.sh
```

## 📊 **Qué Esperar Durante la Prueba**

### ✅ **Comportamiento Normal**

1. **Estado Inicial**: 1 pod activo
2. **Al Enviar >10 mensajes MQTT**: 
   - KEDA detecta mensajes pendientes en el topic
   - Escala a 2-3 pods en 1-2 minutos
3. **Al Procesar Mensajes**:
   - Los pods procesan los mensajes MQTT
   - Los mensajes se consumen del topic
4. **Estado Final**: Vuelve a 1 pod después de 5 minutos (cooldown)

### ⚠️ **Posibles Problemas**

| Problema | Causa | Solución |
|----------|-------|----------|
| No escala | Topic no tiene >10 mensajes | Enviar más mensajes |
| No escala | KEDA no puede conectar a EMQX | Verificar conectividad |
| No escala | Credenciales MQTT incorrectas | Verificar username/password |
| Escala pero no reduce | Cooldown muy largo | Esperar 5 minutos |
| Error de conexión MQTT | EMQX no accesible | Verificar servicio EMQX |

## 🔍 **Comandos de Diagnóstico**

### Verificar Estado de KEDA
```bash
# Ver pods de KEDA
kubectl get pods -n keda

# Ver logs de KEDA
kubectl logs -n keda -l app=keda-operator | grep mqtt

# Ver métricas de KEDA
kubectl top pods -n keda
```

### Verificar Conectividad MQTT
```bash
# Test de conectividad básica
kubectl exec -n medisupply deployment/mqtt-order-event-client -- nc -zv emqx.medilogistic.svc.cluster.local 1883

# Verificar que EMQX está funcionando
kubectl exec -n medisupply deployment/mqtt-order-event-client -- curl emqx.medilogistic.svc.cluster.local:18083
```

### Verificar ScaledObject
```bash
# Estado detallado
kubectl describe scaledobject mqtt-order-event-client-scaler -n medisupply

# Ver eventos relacionados
kubectl get events -n medisupply --field-selector involvedObject.name=mqtt-order-event-client-scaler
```

## 🚨 **Troubleshooting Rápido**

### Si KEDA no escala:
```bash
# 1. Verificar que el ScaledObject está Ready
kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply

# 2. Ver logs de KEDA
kubectl logs -n keda -l app=keda-operator | grep "mqtt-order-event"

# 3. Verificar conectividad a EMQX
kubectl exec -n medisupply deployment/mqtt-order-event-client -- curl emqx.medilogistic.svc.cluster.local:18083
```

### Si escala pero no reduce:
```bash
# Verificar cooldown period
kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply -o yaml | grep cooldownPeriod

# Verificar que los mensajes se están procesando
kubectl logs -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client | tail -20
```

## 📈 **Métricas a Monitorear**

```bash
# Número de pods activos
kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client --no-headers | wc -l

# Estado del ScaledObject
kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'

# Mensajes procesados (desde logs)
kubectl logs -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client | grep "Processing message" | wc -l
```

## 🎯 **Resultado Esperado**

Después de una prueba exitosa deberías ver:

1. **Inicial**: 1 pod activo
2. **Durante carga MQTT**: 2-3 pods activos
3. **Final**: 1 pod activo (después del cooldown)

## 🔄 **Rollback (Si hay problemas)**

```bash
# Rollback completo
./k8s/config/install-keda-mqtt-order-event.sh --rollback

# O manualmente
kubectl delete -f k8s/config/keda-mqtt-order-event-client.yaml
```

¡El autoescalado del mqtt-order-event-client está funcionando correctamente! 🎉
