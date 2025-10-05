# KEDA Autoscaling para Order Management

## 📋 Resumen de Cambios

Esta implementación de KEDA **NO MODIFICA NINGÚN ARCHIVO EXISTENTE** del proyecto. Solo añade nuevos archivos para escalado automático basado en cola de mensajes de RabbitMQ.

## 🎯 Objetivo

Escalar automáticamente el servicio `order_management` de 1 a máximo 3 pods cuando la cola `order-damage-queue` en RabbitMQ tenga más de 10 mensajes.

## 📁 Archivos Nuevos (No modifica existentes)

### 1. `k8s/config/keda-order-management-minimal.yaml`
- **TriggerAuthentication**: Configuración de autenticación para RabbitMQ
- **ScaledObject**: Configuración de escalado automático
- **Completamente independiente** del proyecto existente

### 2. `k8s/scripts/install-keda-scaling.sh` (Linux)
- Script de instalación para sistemas Linux
- Verificaciones previas y rollback automático
- Uso: `./k8s/scripts/install-keda-scaling.sh`

### 3. `k8s/scripts/install-keda-scaling.ps1` (Windows)
- Script de instalación para desarrollo en Windows
- Misma funcionalidad que el script de Linux
- Uso: `.\k8s\scripts\install-keda-scaling.ps1`

## 🚀 Instalación en Linux

```bash
# 1. Hacer ejecutable el script
chmod +x k8s/scripts/install-keda-scaling.sh

# 2. Ejecutar instalación
./k8s/scripts/install-keda-scaling.sh

# 3. Verificar instalación
kubectl get scaledobject -n medisupply
kubectl describe scaledobject order-management-scaler -n medisupply
```

## 🔄 Rollback (Si hay problemas)

```bash
# Rollback completo
./k8s/scripts/install-keda-scaling.sh --rollback

# O manualmente
kubectl delete -f k8s/config/keda-order-management-minimal.yaml
```

## 📊 Configuración de Escalado

| Parámetro | Valor | Descripción |
|-----------|-------|-------------|
| **Min Replicas** | 1 | Mínimo de pods siempre activos |
| **Max Replicas** | 3 | Máximo de pods permitidos |
| **Umbral** | 10 mensajes | Escalar cuando la cola tenga >10 mensajes |
| **Polling** | 60 segundos | Frecuencia de verificación |
| **Cooldown** | 10 minutos | Tiempo antes de reducir réplicas |

## 🔍 Monitoreo

### Ver estado del escalado
```bash
kubectl get scaledobject -n medisupply
kubectl describe scaledobject order-management-scaler -n medisupply
```

### Ver pods escalados
```bash
kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management
```

### Ver logs de KEDA
```bash
kubectl logs -n keda -l app=keda-operator
```

## 🛡️ Seguridad

- **No modifica** ningún deployment existente
- **Usa la configuración actual** de RabbitMQ
- **Rollback completo** disponible en cualquier momento
- **Verificaciones previas** antes de instalar

## 🔧 Configuración Técnica

### RabbitMQ Connection
- **Host**: `rabbitmq.mediorder.svc.cluster.local`
- **Queue**: `order-damage-queue`
- **VHost**: `/`
- **Protocol**: `amqp`
- **Auth**: Usa secret existente `rabbitmq`

### KEDA Configuration
- **Trigger Type**: `rabbitmq`
- **Mode**: `QueueLength`
- **Value**: `10` (mensajes en cola)
- **Authentication**: `TriggerAuthentication` con secret

## 📈 Comportamiento Esperado

1. **Estado Normal**: 1 pod activo
2. **Alta Carga**: Cuando `order-damage-queue` > 10 mensajes → escalar a 2-3 pods
3. **Baja Carga**: Después de 10 minutos sin mensajes → volver a 1 pod

## 🚨 Troubleshooting

### KEDA no escala
```bash
# Verificar que KEDA está funcionando
kubectl get pods -n keda

# Verificar configuración del ScaledObject
kubectl describe scaledobject order-management-scaler -n medisupply

# Verificar conectividad a RabbitMQ
kubectl exec -n medisupply deployment/order-management-order -- curl rabbitmq.mediorder.svc.cluster.local:5672
```

### Error de autenticación
```bash
# Verificar que el secret existe
kubectl get secret rabbitmq -n medisupply

# Verificar TriggerAuthentication
kubectl describe triggerauthentication rabbitmq-order-auth -n medisupply
```

## 📝 Notas Importantes

- ✅ **Compatible** con la configuración actual del proyecto
- ✅ **No invasivo** - no modifica archivos existentes
- ✅ **Reversible** - rollback completo disponible
- ✅ **Probado** con Kind, Istio, Kiali
- ✅ **Funciona** en Linux (entorno de producción)

## 🔗 Referencias

- [KEDA Documentation](https://keda.sh/docs/)
- [KEDA RabbitMQ Scaler](https://keda.sh/docs/2.17/scalers/rabbitmq-queue/)
- [KEDA ScaledObject](https://keda.sh/docs/2.17/concepts/scaling-deployments/)
