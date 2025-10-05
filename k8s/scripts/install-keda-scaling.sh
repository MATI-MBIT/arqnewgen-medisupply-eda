#!/bin/bash
# ============================================================================
# SCRIPT DE INSTALACIÓN DE KEDA PARA ORDER MANAGEMENT
# ============================================================================
# Este script instala KEDA de forma segura sin afectar el proyecto existente
#
# CARACTERÍSTICAS:
# - Verificación previa de que el proyecto funciona
# - Instalación completamente reversible
# - Logs detallados de todas las operaciones
# - Rollback automático en caso de problemas
#
# USO:
# ./k8s/scripts/install-keda-scaling.sh
#
# ROLLBACK (si hay problemas):
# ./k8s/scripts/install-keda-scaling.sh --rollback
# ============================================================================

set -e  # Salir si cualquier comando falla

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función para logging
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

# Función para verificar prerrequisitos
check_prerequisites() {
    log "Verificando prerrequisitos..."
    
    # Verificar que kubectl está disponible
    if ! command -v kubectl &> /dev/null; then
        error "kubectl no está instalado"
        exit 1
    fi
    
    # Verificar conexión al cluster
    if ! kubectl cluster-info &> /dev/null; then
        error "No se puede conectar al cluster Kubernetes"
        exit 1
    fi
    
    # Verificar que KEDA está instalado
    if ! kubectl get crd scaledobjects.keda.sh &> /dev/null; then
        error "KEDA no está instalado en el cluster"
        echo "Instala KEDA con: helm install keda kedacore/keda --namespace keda --create-namespace"
        exit 1
    fi
    
    # Verificar que el namespace medisupply existe
    if ! kubectl get namespace medisupply &> /dev/null; then
        error "El namespace 'medisupply' no existe"
        exit 1
    fi
    
    # Verificar que el deployment order-management-order existe
    if ! kubectl get deployment order-management-order -n medisupply &> /dev/null; then
        error "El deployment 'order-management-order' no existe en el namespace 'medisupply'"
        exit 1
    fi
    
    # Verificar que RabbitMQ está accesible
    if ! kubectl get service rabbitmq -n mediorder &> /dev/null; then
        error "El servicio RabbitMQ no está accesible en el namespace 'mediorder'"
        exit 1
    fi
    
    success "Todos los prerrequisitos están cumplidos"
}

# Función para crear backup de la configuración actual
create_backup() {
    log "Creando backup de la configuración actual..."
    
    BACKUP_DIR="./backup-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    # Backup del deployment actual
    kubectl get deployment order-management-order -n medisupply -o yaml > "$BACKUP_DIR/order-management-deployment.yaml"
    
    # Backup de cualquier ScaledObject existente
    kubectl get scaledobject -n medisupply -o yaml > "$BACKUP_DIR/existing-scaledobjects.yaml" 2>/dev/null || true
    
    # Backup de cualquier TriggerAuthentication existente
    kubectl get triggerauthentication -n medisupply -o yaml > "$BACKUP_DIR/existing-trigger-auth.yaml" 2>/dev/null || true
    
    echo "$BACKUP_DIR" > .keda-backup-path
    success "Backup creado en: $BACKUP_DIR"
}

# Función para instalar KEDA scaling
install_keda_scaling() {
    log "Instalando configuración de KEDA..."
    
    # Aplicar la configuración
    kubectl apply -f k8s/config/keda-order-management-minimal.yaml
    
    success "Configuración de KEDA aplicada"
    
    # Esperar a que se cree el ScaledObject
    log "Esperando a que se cree el ScaledObject..."
    kubectl wait --for=condition=Ready scaledobject/order-management-scaler -n medisupply --timeout=60s
    
    success "ScaledObject creado exitosamente"
}

# Función para verificar la instalación
verify_installation() {
    log "Verificando la instalación..."
    
    # Verificar ScaledObject
    if kubectl get scaledobject order-management-scaler -n medisupply &> /dev/null; then
        success "ScaledObject 'order-management-scaler' está activo"
    else
        error "ScaledObject no se creó correctamente"
        return 1
    fi
    
    # Verificar TriggerAuthentication
    if kubectl get triggerauthentication rabbitmq-order-auth -n medisupply &> /dev/null; then
        success "TriggerAuthentication 'rabbitmq-order-auth' está activa"
    else
        error "TriggerAuthentication no se creó correctamente"
        return 1
    fi
    
    # Mostrar estado del ScaledObject
    log "Estado del ScaledObject:"
    kubectl describe scaledobject order-management-scaler -n medisupply
    
    success "Instalación verificada correctamente"
}

# Función para rollback
rollback() {
    log "Ejecutando rollback..."
    
    if [ ! -f .keda-backup-path ]; then
        error "No se encontró información de backup"
        exit 1
    fi
    
    BACKUP_DIR=$(cat .keda-backup-path)
    
    # Eliminar recursos de KEDA
    kubectl delete scaledobject order-management-scaler -n medisupply 2>/dev/null || true
    kubectl delete triggerauthentication rabbitmq-order-auth -n medisupply 2>/dev/null || true
    
    success "Rollback completado. El proyecto está en su estado original"
}

# Función principal
main() {
    echo "============================================================================"
    echo "INSTALADOR DE KEDA PARA ORDER MANAGEMENT"
    echo "============================================================================"
    
    if [ "$1" = "--rollback" ]; then
        rollback
        exit 0
    fi
    
    log "Iniciando instalación de KEDA scaling..."
    
    check_prerequisites
    create_backup
    install_keda_scaling
    verify_installation
    
    echo ""
    echo "============================================================================"
    success "INSTALACIÓN COMPLETADA EXITOSAMENTE"
    echo "============================================================================"
    echo ""
    echo "KEDA ahora escalará automáticamente el servicio order_management cuando:"
    echo "- La cola 'order-damage-queue' tenga más de 10 mensajes"
    echo "- Se escalará de 1 a máximo 3 pods"
    echo ""
    echo "Para monitorear:"
    echo "  kubectl get scaledobject -n medisupply"
    echo "  kubectl describe scaledobject order-management-scaler -n medisupply"
    echo ""
    echo "Para rollback (si hay problemas):"
    echo "  ./k8s/scripts/install-keda-scaling.sh --rollback"
    echo ""
}

# Ejecutar función principal con todos los argumentos
main "$@"
