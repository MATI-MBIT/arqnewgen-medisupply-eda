# 🧪 Guía de Pruebas para KEDA Autoescalado

## 📋 Métodos para Probar el Autoescalado

### 🚀 **Método 1: Script Automatizado (Recomendado)**

```bash
# Hacer ejecutable
chmod +x k8s/scripts/test-keda-scaling.sh

# Ejecutar prueba completa
./k8s/scripts/test-keda-scaling.sh
```

Este script incluye:
- ✅ Verificación de prerrequisitos
- ✅ Simulación de carga en RabbitMQ
- ✅ Monitoreo en tiempo real
- ✅ Limpieza automática

### ⚡ **Método 2: Verificación Rápida**

```bash
# Verificación básica
chmod +x k8s/scripts/quick-test-keda.sh
./k8s/scripts/quick-test-keda.sh
```

### 🔧 **Método 3: Prueba Manual Paso a Paso**

#### Paso 1: Verificar Estado Inicial
```bash
# Ver pods actuales
kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management

# Ver ScaledObject
kubectl get scaledobject -n medisupply
kubectl describe scaledobject order-management-scaler -n medisupply
```

#### Paso 2: Simular Carga en RabbitMQ

**Opción A: Usando rabbitmqadmin (si está disponible)**
```bash
# Conectar a RabbitMQ y enviar mensajes
kubectl exec -n mediorder deployment/rabbitmq -- rabbitmqadmin \
  publish exchange=events routing_key=order.damage payload='{"test":"load","timestamp":"'$(date)'"}'
```

**Opción B: Usando curl con HTTP API**
```bash
# Obtener credenciales
RABBITMQ_USER="user"
RABBITMQ_PASS=$(kubectl get secret rabbitmq -n medisupply -o jsonpath='{.data.rabbitmq-password}' | base64 -d)

# Enviar mensajes (repetir varias veces)
for i in {1..15}; do
  curl -u "$RABBITMQ_USER:$RABBITMQ_PASS" \
    -H "content-type:application/json" \
    -X POST \
    -d '{"properties":{},"routing_key":"order.damage","payload":"{\"test\":\"load'$i'\",\"timestamp\":\"'$(date)'\"}","payload_encoding":"string"}' \
    http://rabbitmq.mediorder.svc.cluster.local:15672/api/exchanges/%2F/events/publish
  sleep 1
done
```

**Opción C: Crear pod temporal para generar carga**
```bash
# Crear pod temporal
kubectl run load-generator --image=busybox --rm -it --restart=Never -n medisupply -- sh

# Dentro del pod, instalar herramientas y enviar mensajes
apk add --no-cache curl
for i in $(seq 1 20); do
  curl -u "user:public" -X POST \
    -H "Content-Type: application/json" \
    -d '{"properties":{},"routing_key":"order.damage","payload":"{\"test\":\"load'$i'\",\"timestamp\":\"'$(date)'\"}","payload_encoding":"string"}' \
    http://rabbitmq.mediorder.svc.cluster.local:15672/api/exchanges/%2F/events/publish
  sleep 0.5
done
```

#### Paso 3: Monitorear el Escalado

**Monitoreo en tiempo real:**
```bash
# Ver pods escalando
kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management -w

# Ver ScaledObject en tiempo real
kubectl get scaledobject -n medisupply -w

# Ver logs de KEDA
kubectl logs -n keda -l app=keda-operator -f
```

**Verificar métricas:**
```bash
# Ver estado del ScaledObject
kubectl describe scaledobject order-management-scaler -n medisupply

# Ver métricas de RabbitMQ
kubectl exec -n mediorder deployment/rabbitmq -- rabbitmqctl list_queues name messages
```

#### Paso 4: Verificar que Vuelve a Estado Normal

```bash
# Esperar a que se procesen los mensajes (puede tomar 5-10 minutos)
# Monitorear que los pods vuelvan a 1
kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management -w
```

## 📊 **Qué Esperar Durante la Prueba**

### ✅ **Comportamiento Normal**

1. **Estado Inicial**: 1 pod activo
2. **Al Enviar >10 mensajes**: 
   - KEDA detecta la cola con muchos mensajes
   - Escala a 2-3 pods en 1-2 minutos
3. **Al Procesar Mensajes**:
   - Los pods procesan los mensajes
   - La cola se vacía gradualmente
4. **Estado Final**: Vuelve a 1 pod después de 10 minutos (cooldown)

### ⚠️ **Posibles Problemas**

| Problema | Causa | Solución |
|----------|-------|----------|
| No escala | Cola no tiene >10 mensajes | Enviar más mensajes |
| No escala | KEDA no puede conectar a RabbitMQ | Verificar conectividad |
| Escala pero no reduce | Cooldown muy largo | Esperar 10 minutos |
| Error de autenticación | Secret incorrecto | Verificar secret rabbitmq |

## 🔍 **Comandos de Diagnóstico**

### Verificar Estado de KEDA
```bash
# Ver pods de KEDA
kubectl get pods -n keda

# Ver logs de KEDA
kubectl logs -n keda -l app=keda-operator

# Ver métricas de KEDA
kubectl top pods -n keda
```

### Verificar Conectividad a RabbitMQ
```bash
# Test de conectividad básica
kubectl exec -n medisupply deployment/order-management-order -- nc -zv rabbitmq.mediorder.svc.cluster.local 5672

# Verificar credenciales
kubectl get secret rabbitmq -n medisupply -o yaml
```

### Verificar ScaledObject
```bash
# Estado detallado
kubectl describe scaledobject order-management-scaler -n medisupply

# Ver eventos relacionados
kubectl get events -n medisupply --field-selector involvedObject.name=order-management-scaler
```

## 🚨 **Troubleshooting Rápido**

### Si KEDA no escala:
```bash
# 1. Verificar que el ScaledObject está Ready
kubectl get scaledobject order-management-scaler -n medisupply

# 2. Ver logs de KEDA
kubectl logs -n keda -l app=keda-operator | grep "order-management"

# 3. Verificar conectividad a RabbitMQ
kubectl exec -n medisupply deployment/order-management-order -- curl rabbitmq.mediorder.svc.cluster.local:15672
```

### Si escala pero no reduce:
```bash
# Verificar cooldown period
kubectl get scaledobject order-management-scaler -n medisupply -o yaml | grep cooldownPeriod

# Forzar escala down (NO recomendado en producción)
kubectl scale deployment order-management-order --replicas=1 -n medisupply
```

## 📈 **Métricas a Monitorear**

```bash
# Número de pods activos
kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management --no-headers | wc -l

# Mensajes en cola RabbitMQ
kubectl exec -n mediorder deployment/rabbitmq -- rabbitmqctl list_queues name messages | grep order-damage

# Estado del ScaledObject
kubectl get scaledobject order-management-scaler -n medisupply -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
```

## 🎯 **Resultado Esperado**

Después de una prueba exitosa deberías ver:

1. **Inicial**: 1 pod activo
2. **Durante carga**: 2-3 pods activos
3. **Final**: 1 pod activo (después del cooldown)

¡El autoescalado está funcionando correctamente! 🎉
