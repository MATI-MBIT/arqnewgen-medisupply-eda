#!/bin/bash
# ============================================================================
# SCRIPT DE PRUEBA PARA MQTT ORDER EVENT CLIENT SCALING
# ============================================================================
# Este script prueba el autoescalado del mqtt-order-event-client
# enviando mensajes MQTT al topic events/sensor
#
# USO:
# ./k8s/scripts/test-mqtt-order-event-scaling.sh
# ============================================================================

set -e

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log() { echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }
warning() { echo -e "${YELLOW}⚠${NC} $1"; }
info() { echo -e "${CYAN}ℹ${NC} $1"; }

# Configuración MQTT
MQTT_BROKER="emqx.medilogistic.svc.cluster.local:1883"
MQTT_TOPIC="events/sensor"
MQTT_USERNAME="admin"
MQTT_PASSWORD="public"

# Función para mostrar estado actual
show_status() {
    echo ""
    log "=== ESTADO ACTUAL ==="
    
    # Mostrar ScaledObject
    echo -e "\n${CYAN}ScaledObject mqtt-order-event-client:${NC}"
    kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply -o wide 2>/dev/null || warning "ScaledObject no encontrado"
    
    # Mostrar pods actuales
    echo -e "\n${CYAN}Pods del mqtt-order-event-client:${NC}"
    kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client -o wide
    
    # Mostrar estado de EMQX
    echo -e "\n${CYAN}Estado de EMQX:${NC}"
    kubectl get pods -n medisupply -l app.kubernetes.io/name=emqx -o wide 2>/dev/null || warning "EMQX no encontrado"
    
    echo ""
}

# Función para verificar prerrequisitos
check_prerequisites() {
    log "Verificando prerrequisitos para la prueba..."
    
    # Verificar que KEDA está instalado
    if ! kubectl get crd scaledobjects.keda.sh &> /dev/null; then
        error "KEDA no está instalado"
        exit 1
    fi
    
    # Verificar que el ScaledObject existe
    if ! kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply &> /dev/null; then
        error "ScaledObject no está instalado. Instala primero con:"
        echo "kubectl apply -f k8s/config/keda-mqtt-order-event-client.yaml"
        exit 1
    fi
    
    # Verificar que EMQX está accesible
    if ! kubectl get service emqx -n medisupply &> /dev/null; then
        error "EMQX no está accesible en el namespace medisupply"
        exit 1
    fi
    
    # Verificar que el deployment existe
    if ! kubectl get deployment mqtt-order-event-client -n medisupply &> /dev/null; then
        error "Deployment mqtt-order-event-client no encontrado en namespace medisupply"
        exit 1
    fi
    
    success "Prerrequisitos verificados"
}

# Función para crear pod temporal de MQTT client
create_mqtt_client() {
    log "Creando pod temporal para envío de mensajes MQTT..."
    
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
    args: ["-c", "sleep 3600"]
    env:
    - name: MQTT_BROKER
      value: "$MQTT_BROKER"
    - name: MQTT_TOPIC
      value: "$MQTT_TOPIC"
    - name: MQTT_USERNAME
      value: "$MQTT_USERNAME"
    - name: MQTT_PASSWORD
      value: "$MQTT_PASSWORD"
  restartPolicy: Never
EOF

    # Esperar a que el pod esté listo
    kubectl wait --for=condition=Ready pod/mqtt-test-client -n medisupply --timeout=60s
    
    success "Pod MQTT client creado"
}

# Función para enviar mensajes MQTT
send_mqtt_messages() {
    local count=${1:-15}
    
    log "Enviando $count mensajes MQTT al topic $MQTT_TOPIC..."
    
    # Crear script temporal para enviar mensajes
    cat <<'EOF' > /tmp/send_mqtt_messages.sh
#!/bin/bash
COUNT=$1
MQTT_BROKER=$2
MQTT_TOPIC=$3
MQTT_USERNAME=$4
MQTT_PASSWORD=$5

for i in $(seq 1 $COUNT); do
    MESSAGE="{\"id\":\"test-$i\",\"timestamp\":\"$(date -Iseconds)\",\"sensor\":\"temperature\",\"value\":$((20 + RANDOM % 10)),\"location\":\"warehouse-a\"}"
    
    mosquitto_pub \
        -h $MQTT_BROKER \
        -p 1883 \
        -t "$MQTT_TOPIC" \
        -m "$MESSAGE" \
        -u "$MQTT_USERNAME" \
        -P "$MQTT_PASSWORD" \
        -q 1
    
    echo "Mensaje $i enviado"
    sleep 0.5
done
EOF

    chmod +x /tmp/send_mqtt_messages.sh
    
    # Copiar script al pod y ejecutarlo
    kubectl cp /tmp/send_mqtt_messages.sh medisupply/mqtt-test-client:/tmp/send_mqtt_messages.sh
    
    # Instalar mosquitto client en el pod
    kubectl exec -n medisupply mqtt-test-client -- apk add --no-cache mosquitto-clients
    
    # Ejecutar el script
    kubectl exec -n medisupply mqtt-test-client -- /tmp/send_mqtt_messages.sh $count $MQTT_BROKER $MQTT_TOPIC $MQTT_USERNAME $MQTT_PASSWORD
    
    success "$count mensajes MQTT enviados"
}

# Función para monitorear escalado
monitor_scaling() {
    local duration=${1:-300}  # 5 minutos por defecto
    local interval=10         # Verificar cada 10 segundos
    
    log "Monitoreando escalado por $duration segundos (intervalo: ${interval}s)..."
    
    local start_time=$(date +%s)
    local end_time=$((start_time + duration))
    
    while [ $(date +%s) -lt $end_time ]; do
        local current_time=$(date +%H:%M:%S)
        local pod_count=$(kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client --no-headers | wc -l)
        
        echo -e "${CYAN}[$current_time]${NC} Pods activos: ${GREEN}$pod_count${NC}"
        
        # Mostrar estado del ScaledObject
        local scaledobject_status=$(kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null || echo "Unknown")
        echo "  ScaledObject: $scaledobject_status"
        
        sleep $interval
    done
}

# Función para limpiar recursos de prueba
cleanup_test() {
    log "Limpiando recursos de prueba..."
    
    # Eliminar pod de prueba
    kubectl delete pod mqtt-test-client -n medisupply 2>/dev/null || true
    
    # Eliminar script temporal
    rm -f /tmp/send_mqtt_messages.sh
    
    # Esperar un poco para que se procesen mensajes
    log "Esperando a que se procesen los mensajes restantes..."
    sleep 30
    
    success "Limpieza completada"
}

# Función para prueba manual completa
manual_test() {
    log "=== PRUEBA MANUAL COMPLETA ==="
    
    echo ""
    info "Esta prueba:"
    echo "1. Verifica el estado inicial"
    echo "2. Envía mensajes MQTT para activar escalado"
    echo "3. Monitorea el escalado en tiempo real"
    echo "4. Verifica que vuelve a estado normal"
    echo ""
    
    read -p "¿Continuar con la prueba? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log "Prueba cancelada"
        exit 0
    fi
    
    # Paso 1: Estado inicial
    log "Paso 1: Estado inicial"
    show_status
    
    # Paso 2: Crear cliente MQTT
    log "Paso 2: Creando cliente MQTT temporal"
    create_mqtt_client
    
    # Paso 3: Enviar mensajes
    log "Paso 3: Enviando mensajes MQTT"
    send_mqtt_messages 15
    
    # Paso 4: Monitorear escalado
    log "Paso 4: Monitoreando escalado (5 minutos)"
    monitor_scaling 300
    
    # Paso 5: Estado después del escalado
    log "Paso 5: Estado después del escalado"
    show_status
    
    # Paso 6: Limpiar y verificar estado final
    log "Paso 6: Limpiando y verificando estado final"
    cleanup_test
    sleep 30
    show_status
    
    success "Prueba manual completada"
}

# Función para prueba rápida
quick_test() {
    log "=== PRUEBA RÁPIDA ==="
    
    # Estado inicial
    show_status
    
    # Crear cliente y enviar mensajes
    create_mqtt_client
    send_mqtt_messages 12
    
    # Esperar un poco y verificar escalado
    sleep 60
    
    local pod_count=$(kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client --no-headers | wc -l)
    if [ $pod_count -gt 1 ]; then
        success "Escalado detectado: $pod_count pods activos"
    else
        warning "No se detectó escalado: $pod_count pods activos"
        info "Esto puede ser normal si KEDA aún está procesando"
    fi
    
    # Limpiar
    cleanup_test
    
    success "Prueba rápida completada"
}

# Función principal
main() {
    echo "============================================================================"
    echo "PROBADOR DE MQTT ORDER EVENT CLIENT SCALING"
    echo "============================================================================"
    
    check_prerequisites
    
    echo ""
    echo "Opciones de prueba:"
    echo "1) Prueba manual completa (recomendada)"
    echo "2) Prueba rápida automática"
    echo "3) Solo mostrar estado actual"
    echo "4) Enviar mensajes MQTT manualmente"
    echo ""
    
    read -p "Selecciona una opción (1-4): " -n 1 -r
    echo
    
    case $REPLY in
        1)
            manual_test
            ;;
        2)
            quick_test
            ;;
        3)
            show_status
            ;;
        4)
            create_mqtt_client
            echo "Pod MQTT client creado. Puedes enviar mensajes manualmente:"
            echo "kubectl exec -n medisupply mqtt-test-client -- mosquitto_pub -h $MQTT_BROKER -t '$MQTT_TOPIC' -m 'test message' -u $MQTT_USERNAME -P $MQTT_PASSWORD"
            ;;
        *)
            error "Opción inválida"
            exit 1
            ;;
    esac
    
    echo ""
    echo "============================================================================"
    success "PRUEBA COMPLETADA"
    echo "============================================================================"
    echo ""
    echo "Comandos útiles para monitoreo:"
    echo "  kubectl get scaledobject -n medisupply -w"
    echo "  kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client -w"
    echo "  kubectl describe scaledobject mqtt-order-event-client-scaler -n medisupply"
    echo ""
}

main "$@"
