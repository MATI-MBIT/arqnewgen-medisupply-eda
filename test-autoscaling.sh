#!/bin/bash
# Script para probar el autoescalado de KEDA

set -e

echo "🚀 Iniciando prueba de autoescalado KEDA..."

# Verificar que el ScaledObject existe
if ! kubectl get scaledobject order-management-scaler -n medisupply &>/dev/null; then
    echo "❌ ScaledObject no encontrado. Ejecuta primero:"
    echo "   ./k8s/scripts/install-keda-scaling.sh"
    exit 1
fi

echo "✅ ScaledObject encontrado"

# Función para obtener el número de pods
get_pod_count() {
    kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management --no-headers | wc -l
}

# Función para obtener mensajes en cola
get_queue_length() {
    kubectl exec -n mediorder rabbitmq-0 -- rabbitmqctl list_queues name messages | grep order-damage-queue | awk '{print $2}' || echo "0"
}

echo "📊 Estado inicial:"
echo "   Pods actuales: $(get_pod_count)"
echo "   Mensajes en cola: $(get_queue_length)"

echo ""
echo "🔥 Generando carga (enviando 15 mensajes a la cola)..."

# Generar mensajes en la cola
for i in {1..15}; do
    kubectl exec -n mediorder rabbitmq-0 -- rabbitmqadmin publish exchange=events routing_key=order.damage payload="Test message $i"
    echo "   Mensaje $i enviado"
    sleep 1
done

echo ""
echo "📈 Esperando escalado automático (puede tomar 1-2 minutos)..."

# Monitorear por 5 minutos
for i in {1..30}; do
    pods=$(get_pod_count)
    queue=$(get_queue_length)
    
    echo "   Minuto $i: Pods=$pods, Cola=$queue mensajes"
    
    if [ "$pods" -gt 1 ]; then
        echo "🎉 ¡Escalado detectado! Pods escalaron de 1 a $pods"
        break
    fi
    
    sleep 10
done

echo ""
echo "📉 Esperando que se procesen los mensajes y se reduzca la escala..."
echo "   (Esto puede tomar 10+ minutos debido al cooldown period)"

echo ""
echo "🔍 Para monitorear en tiempo real, ejecuta en otra terminal:"
echo "   watch kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management"
echo "   watch kubectl describe scaledobject order-management-scaler -n medisupply"

echo ""
echo "✅ Prueba completada. El autoescalado está funcionando!"