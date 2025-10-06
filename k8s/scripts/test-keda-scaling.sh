#!/bin/bash
# ============================================================================
# SCRIPT DE PRUEBA PARA KEDA AUTOESCALADO
# ============================================================================
# Este script prueba el autoescalado de KEDA de diferentes maneras:
# 1. Verifica que KEDA esté funcionando
# 2. Simula carga en RabbitMQ para activar escalado
# 3. Monitorea el comportamiento del escalado
# 4. Verifica que vuelva a estado normal
#
# USO:
# ./k8s/scripts/test-keda-scaling.sh
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

# Función para mostrar estado actual
show_status() {
    echo ""
    log "=== ESTADO ACTUAL ==="
    
    # Mostrar ScaledObject
    echo -e "\n${CYAN}ScaledObject:${NC}"
    kubectl get scaledobject order-management-scaler -n medisupply -o wide 2>/dev/null || error "ScaledObject no encontrado"
    
    # Mostrar pods actuales
    echo -e "\n${CYAN}Pods del order_management:${NC}"
    kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management -o wide
    
    # Mostrar métricas de RabbitMQ (si están disponibles)
    echo -e "\n${CYAN}Cola RabbitMQ (order-damage-queue):${NC}"
    # Intentar obtener métricas de la cola
    kubectl exec -n mediorder deployment/rabbitmq -- rabbitmqctl list_queues name messages 2>/dev/null | grep order-damage-queue || warning "No se pudo obtener métricas de RabbitMQ"
    
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
    if ! kubectl get scaledobject order-management-scaler -n medisupply &> /dev/null; then
        error "ScaledObject no está instalado. Instala primero con:"
        echo "./k8s/config/install-keda-linux.sh"
        exit 1
    fi
    
    # Verificar que RabbitMQ está accesible
    if ! kubectl get service rabbitmq -n mediorder &> /dev/null; then
        error "RabbitMQ no está accesible"
        exit 1
    fi
    
    success "Prerrequisitos verificados"
}

# Función para monitorear escalado en tiempo real
monitor_scaling() {
    local duration=${1:-300}  # 5 minutos por defecto
    local interval=10         # Verificar cada 10 segundos
    
    log "Monitoreando escalado por $duration segundos (intervalo: ${interval}s)..."
    
    local start_time=$(date +%s)
    local end_time=$((start_time + duration))
    
    while [ $(date +%s) -lt $end_time ]; do
        local current_time=$(date +%H:%M:%S)
        local pod_count=$(kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management --no-headers | wc -l)
        
        echo -e "${CYAN}[$current_time]${NC} Pods activos: ${GREEN}$pod_count${NC}"
        
        # Mostrar estado del ScaledObject
        kubectl get scaledobject order-management-scaler -n medisupply -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}' 2>/dev/null | grep -q "True" && echo "  ScaledObject: Ready" || echo "  ScaledObject: Not Ready"
        
        sleep $interval
    done
}

# Función para simular carga en RabbitMQ
simulate_load() {
    log "Simulando carga en RabbitMQ..."
    
    # Crear un pod temporal para enviar mensajes a RabbitMQ
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: rabbitmq-load-generator
  namespace: medisupply
spec:
  containers:
  - name: rabbitmq-producer
    image: rabbitmq:3-management
    command: ["/bin/bash"]
    args: ["-c", "while true; do rabbitmqadmin -H rabbitmq.mediorder.svc.cluster.local -u user -p \$(cat /var/run/secrets/kubernetes.io/serviceaccount/..data/rabbitmq-password 2>/dev/null || echo 'public') publish exchange=events routing_key=order.damage payload='{\"test\":\"load\",\"timestamp\":\"'$(date)'\"}'; sleep 0.1; done"]
  restartPolicy: Never
EOF

    info "Pod de carga creado. Enviando mensajes a la cola..."
    sleep 5
    
    # Verificar que el pod está funcionando
    if kubectl get pod rabbitmq-load-generator -n medisupply &> /dev/null; then
        success "Generador de carga activo"
        echo "El pod enviará mensajes cada 0.1 segundos para simular carga alta"
        echo "Esto debería activar el escalado cuando supere 10 mensajes"
    else
        error "No se pudo crear el generador de carga"
        return 1
    fi
}

# Función para limpiar recursos de prueba
cleanup_test() {
    log "Limpiando recursos de prueba..."
    
    # Eliminar pod de carga
    kubectl delete pod rabbitmq-load-generator -n medisupply 2>/dev/null || true
    
    # Esperar a que los mensajes se procesen
    log "Esperando a que se procesen los mensajes restantes..."
    sleep 30
    
    success "Limpieza completada"
}

# Función para prueba manual de escalado
manual_test() {
    log "=== PRUEBA MANUAL DE ESCALADO ==="
    
    echo ""
    info "Esta prueba te permitirá:"
    echo "1. Ver el estado inicial"
    echo "2. Simular carga para activar escalado"
    echo "3. Monitorear el escalado en tiempo real"
    echo "4. Verificar que vuelve a estado normal"
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
    
    # Paso 2: Simular carga
    log "Paso 2: Simulando carga en RabbitMQ"
    simulate_load
    
    # Paso 3: Monitorear escalado
    log "Paso 3: Monitoreando escalado (5 minutos)"
    monitor_scaling 300
    
    # Paso 4: Estado después del escalado
    log "Paso 4: Estado después del escalado"
    show_status
    
    # Paso 5: Limpiar y verificar estado final
    log "Paso 5: Limpiando y verificando estado final"
    cleanup_test
    sleep 30
    show_status
    
    success "Prueba manual completada"
}

# Función para prueba automática rápida
quick_test() {
    log "=== PRUEBA RÁPIDA AUTOMÁTICA ==="
    
    # Estado inicial
    show_status
    
    # Simular carga breve
    log "Simulando carga breve..."
    simulate_load
    sleep 60  # Esperar 1 minuto
    
    # Verificar escalado
    local pod_count=$(kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management --no-headers | wc -l)
    if [ $pod_count -gt 1 ]; then
        success "Escalado detectado: $pod_count pods activos"
    else
        warning "No se detectó escalado: $pod_count pods activos"
        info "Esto puede ser normal si la cola no tiene suficientes mensajes"
    fi
    
    # Limpiar
    cleanup_test
    
    success "Prueba rápida completada"
}

# Función principal
main() {
    echo "============================================================================"
    echo "PROBADOR DE KEDA AUTOESCALADO - ORDER MANAGEMENT"
    echo "============================================================================"
    
    check_prerequisites
    
    echo ""
    echo "Opciones de prueba:"
    echo "1) Prueba manual completa (recomendada)"
    echo "2) Prueba rápida automática"
    echo "3) Solo mostrar estado actual"
    echo "4) Monitorear escalado en tiempo real"
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
            log "Monitoreando escalado en tiempo real (Ctrl+C para salir)..."
            monitor_scaling 600
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
    echo "Comandos útiles para monitoreo continuo:"
    echo "  kubectl get scaledobject -n medisupply -w"
    echo "  kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management -w"
    echo "  kubectl describe scaledobject order-management-scaler -n medisupply"
    echo ""
}

main "$@"
