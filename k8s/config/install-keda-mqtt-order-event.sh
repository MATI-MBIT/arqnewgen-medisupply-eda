#!/bin/bash
# ============================================================================
# INSTALADOR KEDA PARA MQTT ORDER EVENT CLIENT
# ============================================================================
# Script para instalar KEDA scaling en mqtt-order-event-client
#
# USO:
# ./k8s/config/install-keda-mqtt-order-event.sh
#
# ROLLBACK:
# ./k8s/config/install-keda-mqtt-order-event.sh --rollback
# ============================================================================

set -e

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }
warning() { echo -e "${YELLOW}⚠${NC} $1"; }

# Función de rollback
rollback() {
    log "Ejecutando rollback..."
    kubectl delete -f k8s/config/keda-mqtt-order-event-client.yaml 2>/dev/null || true
    success "Rollback completado. KEDA eliminado."
    exit 0
}

# Verificar prerrequisitos básicos
check_basic() {
    log "Verificando prerrequisitos..."
    
    # Verificar KEDA
    if ! kubectl get crd scaledobjects.keda.sh &> /dev/null; then
        error "KEDA no está instalado. Instala con:"
        echo "helm repo add kedacore https://kedacore.github.io/charts"
        echo "helm install keda kedacore/keda --namespace keda --create-namespace"
        exit 1
    fi
    
    # Verificar deployment
    if ! kubectl get deployment mqtt-order-event-client -n medisupply &> /dev/null; then
        error "Deployment mqtt-order-event-client no encontrado en namespace medisupply"
        exit 1
    fi
    
    # Verificar EMQX
    if ! kubectl get service emqx -n medisupply &> /dev/null; then
        error "Servicio EMQX no encontrado en namespace medisupply"
        exit 1
    fi
    
    success "Prerrequisitos verificados"
}

# Instalación principal
install() {
    log "Instalando KEDA scaling para mqtt-order-event-client..."
    
    # Aplicar configuración
    kubectl apply -f k8s/config/keda-mqtt-order-event-client.yaml
    
    # Esperar a que esté listo
    kubectl wait --for=condition=Ready scaledobject/mqtt-order-event-client-scaler -n medisupply --timeout=60s
    
    success "KEDA scaling instalado correctamente"
    
    # Mostrar estado
    log "Estado del ScaledObject:"
    kubectl get scaledobject mqtt-order-event-client-scaler -n medisupply
}

# Función principal
main() {
    echo "============================================================================"
    echo "INSTALADOR KEDA - MQTT ORDER EVENT CLIENT SCALING"
    echo "============================================================================"
    
    if [ "$1" = "--rollback" ]; then
        rollback
    fi
    
    check_basic
    install
    
    echo ""
    echo "============================================================================"
    success "INSTALACIÓN COMPLETADA"
    echo "============================================================================"
    echo ""
    echo "El servicio mqtt-order-event-client ahora escala automáticamente:"
    echo "- Mínimo: 1 pod"
    echo "- Máximo: 3 pods"
    echo "- Umbral: >10 mensajes en topic events/sensor"
    echo ""
    echo "Comandos útiles:"
    echo "  kubectl get scaledobject -n medisupply"
    echo "  kubectl get pods -n medisupply -l app.kubernetes.io/name=mqtt-order-event-client"
    echo ""
    echo "Para probar el escalado:"
    echo "  ./k8s/scripts/test-mqtt-order-event-scaling.sh"
    echo ""
    echo "Rollback: ./k8s/config/install-keda-mqtt-order-event.sh --rollback"
    echo ""
}

main "$@"
