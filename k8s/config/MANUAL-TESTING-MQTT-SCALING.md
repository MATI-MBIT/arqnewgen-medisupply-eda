# 🧪 Pruebas Manuales de Escalado - MQTT Order Event Client

## 📋 Pruebas Sin Scripts - Solo Comandos Manuales

Esta guía te permite probar el escalado del `mqtt-order-event-client` usando únicamente comandos kubectl y herramientas básicas, sin instalar ningún script.

## 🎯 **Objetivo**
Verificar que el servicio `mqtt-order-event-client` escala automáticamente de 1 a 3 pods cuando hay más de 10 mensajes MQTT en el topic `events/sensor`.

---

## 🔧 **Paso 1: Instalar KEDA Scaling (Solo una vez)**

```bash
# Aplicar configuración de KEDA
kubectl apply -f k8s/config/keda-mqtt-order-event-client.yaml

# Verificar que se instaló correctamente
kubectl get scaledobject -n medisupply
kubectl describe scaledobject mqtt-order-event-client-scaler -n medisupply
```

---

## 📊 **Paso 2: Verificar Estado Inicial**

```bash
# Ver pods actuales del mqtt-order-event-client
kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client

# Ver estado del ScaledObject
kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply

# Ver que EMQX está funcionando
kubectl get pods -n medisupply -l app.kubernetes.io/name=emqx

# Ver servicios
kubectl get services -n medisupply | grep emqx
```

**Resultado esperado:** 1 pod activo del mqtt-order-event-client

---

## 🚀 **Paso 3: Crear Cliente MQTT Temporal**

```bash
# Crear pod temporal para enviar mensajes MQTT
kubectl run mqtt-test-client \
  --image=eclipse-mosquitto:2.0 \
  --namespace=medisupply \
  --restart=Never \
  --command -- /bin/sh -c "apk add --no-cache mosquitto-clients && sleep 3600"

# Esperar a que esté listo
kubectl wait --for=condition=Ready pod/mqtt-test-client -n medisupply --timeout=60s

# Verificar que el pod está funcionando
kubectl get pod mqtt-test-client -n medisupply
```

---

## 📤 **Paso 4: Enviar Mensajes MQTT para Activar Escalado**

### **Método A: Envío Individual (Paso a Paso)**

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

# Verificar que el mensaje se envió
echo "Mensaje 1 enviado ✓"
```

### **Método B: Envío Masivo (Activar Escalado)**

```bash
# Enviar 15 mensajes para activar escalado (>10)
for i in {1..15}; do
  kubectl exec -n medisupply mqtt-test-client -- mosquitto_pub \
    -h emqx.medilogistic.svc.cluster.local \
    -p 1883 \
    -t "events/sensor" \
    -m "{\"id\":\"test-$i\",\"timestamp\":\"$(date -Iseconds)\",\"sensor\":\"temperature\",\"value\":$((20 + RANDOM % 10)),\"location\":\"warehouse-a\"}" \
    -u admin \
    -P public \
    -q 1
  echo "Mensaje $i enviado ✓"
  sleep 0.5
done

echo "✅ 15 mensajes enviados - Esto debería activar el escalado"
```

### **Método C: Envío Continuo (Carga Alta)**

```bash
# Enviar mensajes continuamente (en una terminal separada)
while true; do
  kubectl exec -n medisupply mqtt-test-client -- mosquitto_pub \
    -h emqx.medilogistic.svc.cluster.local \
    -p 1883 \
    -t "events/sensor" \
    -m "{\"id\":\"continuous-$(date +%s)\",\"timestamp\":\"$(date -Iseconds)\",\"sensor\":\"temperature\",\"value\":$((20 + RANDOM % 10)),\"location\":\"warehouse-a\"}" \
    -u admin \
    -P public \
    -q 1
  sleep 1
done
```

---

## 👀 **Paso 5: Monitorear el Escalado**

### **Abrir Terminal 1: Monitorear Pods**
```bash
# Ver pods escalando en tiempo real
kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client -w
```

### **Abrir Terminal 2: Monitorear ScaledObject**
```bash
# Ver ScaledObject en tiempo real
kubectl get scaledobject -n medisupply -w
```

### **Abrir Terminal 3: Ver Logs de KEDA**
```bash
# Ver logs de KEDA para entender el escalado
kubectl logs -n keda -l app=keda-operator -f | grep mqtt-order-event
```

### **Verificar Estado Cada 30 Segundos**
```bash
# Comando para ejecutar periódicamente
watch -n 30 'echo "=== PODS ==="; kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client; echo ""; echo "=== SCALEDOBJECT ==="; kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply'
```

---

## 📈 **Paso 6: Verificar Procesamiento de Mensajes**

```bash
# Ver logs del mqtt-order-event-client para confirmar que está procesando
kubectl logs -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client -f

# O ver logs de un pod específico
kubectl logs -n medisupply deployment/mqtt-order-event-client -f
```

---

## ⏱️ **Paso 7: Esperar y Verificar Scale Down**

```bash
# Después de 5-10 minutos, los mensajes deberían procesarse
# y el escalado debería volver a 1 pod

# Verificar estado final
kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client
kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply
```

---

## 🧹 **Paso 8: Limpiar Recursos de Prueba**

```bash
# Eliminar pod de prueba
kubectl delete pod mqtt-test-client -n medisupply

# Verificar que se eliminó
kubectl get pods -n medisupply | grep mqtt-test
```

---

## 📊 **Qué Esperar Durante la Prueba**

### ✅ **Comportamiento Normal:**

| Tiempo | Estado | Pods | Descripción |
|--------|--------|------|-------------|
| 0 min | Inicial | 1 | Estado normal |
| 1-2 min | Escalando | 2-3 | KEDA detecta mensajes pendientes |
| 5-10 min | Procesando | 2-3 | Pods procesan mensajes |
| 10-15 min | Scale Down | 1 | Vuelve a estado normal |

### 🔍 **Comandos de Verificación Rápida:**

```bash
# Ver estado actual completo
echo "=== PODS ==="
kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client

echo "=== SCALEDOBJECT ==="
kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply

echo "=== RECENT EVENTS ==="
kubectl get events -n medisupply --field-selector involvedObject.name=mqtt-order-event-client-scaler --sort-by='.lastTimestamp' | tail -5
```

---

## 🚨 **Troubleshooting Manual**

### **Si no escala:**

```bash
# 1. Verificar ScaledObject
kubectl describe scaledobject mqtt-order-event-client-scaler -n medisupply

# 2. Verificar conectividad a EMQX
kubectl exec -n medisupply deployment/mqtt-order-event-client -- nc -zv emqx.medilogistic.svc.cluster.local 1883

# 3. Verificar que EMQX está funcionando
kubectl exec -n medisupply deployment/mqtt-order-event-client -- curl -s emqx.medilogistic.svc.cluster.local:18083

# 4. Ver logs de KEDA
kubectl logs -n keda -l app=keda-operator | grep mqtt
```

### **Si escala pero no reduce:**

```bash
# Verificar cooldown period
kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply -o yaml | grep cooldownPeriod

# Ver si hay mensajes pendientes
kubectl logs -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client | tail -20
```

---

## 🎯 **Resultado Esperado**

Después de una prueba exitosa deberías ver:

1. **Inicial**: 1 pod activo
2. **Durante carga MQTT**: 2-3 pods activos  
3. **Final**: 1 pod activo (después del cooldown)

---

## 🔄 **Rollback Manual (Si hay problemas)**

```bash
# Eliminar KEDA scaling
kubectl delete -f k8s/config/keda-mqtt-order-event-client.yaml

# Verificar que se eliminó
kubectl get scaledobject -n medisupply | grep mqtt-order-event
```

---

## 📝 **Notas Importantes**

- ✅ **Sin scripts**: Solo comandos kubectl básicos
- ✅ **Completamente manual**: Control total sobre cada paso
- ✅ **Fácil rollback**: Un solo comando para deshacer
- ✅ **Monitoreo en tiempo real**: Múltiples terminales para observar
- ✅ **Pruebas graduales**: Puedes enviar mensajes uno por uno

**¡Ahora puedes probar el escalado completamente de forma manual!** 🎉
