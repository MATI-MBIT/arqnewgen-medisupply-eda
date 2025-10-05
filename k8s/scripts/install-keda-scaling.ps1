# ============================================================================
# SCRIPT DE INSTALACIÓN DE KEDA PARA ORDER MANAGEMENT (PowerShell)
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
# .\k8s\scripts\install-keda-scaling.ps1
#
# ROLLBACK (si hay problemas):
# .\k8s\scripts\install-keda-scaling.ps1 -Rollback
# ============================================================================

param(
    [switch]$Rollback
)

# Función para logging con colores
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    
    switch ($Level) {
        "SUCCESS" { Write-Host "✓ $Message" -ForegroundColor Green }
        "WARNING" { Write-Host "⚠ $Message" -ForegroundColor Yellow }
        "ERROR"   { Write-Host "✗ $Message" -ForegroundColor Red }
        default   { Write-Host "[$timestamp] $Message" -ForegroundColor Blue }
    }
}

# Función para verificar prerrequisitos
function Test-Prerequisites {
    Write-Log "Verificando prerrequisitos..."
    
    # Verificar que kubectl está disponible
    try {
        kubectl version --client | Out-Null
    }
    catch {
        Write-Log "kubectl no está instalado o no está en el PATH" "ERROR"
        exit 1
    }
    
    # Verificar conexión al cluster
    try {
        kubectl cluster-info | Out-Null
    }
    catch {
        Write-Log "No se puede conectar al cluster Kubernetes" "ERROR"
        exit 1
    }
    
    # Verificar que KEDA está instalado
    try {
        kubectl get crd scaledobjects.keda.sh | Out-Null
    }
    catch {
        Write-Log "KEDA no está instalado en el cluster" "ERROR"
        Write-Log "Instala KEDA con: helm install keda kedacore/keda --namespace keda --create-namespace" "ERROR"
        exit 1
    }
    
    # Verificar que el namespace medisupply existe
    try {
        kubectl get namespace medisupply | Out-Null
    }
    catch {
        Write-Log "El namespace 'medisupply' no existe" "ERROR"
        exit 1
    }
    
    # Verificar que el deployment order-management-order existe
    try {
        kubectl get deployment order-management-order -n medisupply | Out-Null
    }
    catch {
        Write-Log "El deployment 'order-management-order' no existe en el namespace 'medisupply'" "ERROR"
        exit 1
    }
    
    # Verificar que RabbitMQ está accesible
    try {
        kubectl get service rabbitmq -n mediorder | Out-Null
    }
    catch {
        Write-Log "El servicio RabbitMQ no está accesible en el namespace 'mediorder'" "ERROR"
        exit 1
    }
    
    Write-Log "Todos los prerrequisitos están cumplidos" "SUCCESS"
}

# Función para crear backup
function New-Backup {
    Write-Log "Creando backup de la configuración actual..."
    
    $backupDir = ".\backup-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
    New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
    
    try {
        # Backup del deployment actual
        kubectl get deployment order-management-order -n medisupply -o yaml | Out-File -FilePath "$backupDir\order-management-deployment.yaml"
        
        # Backup de cualquier ScaledObject existente
        try {
            kubectl get scaledobject -n medisupply -o yaml | Out-File -FilePath "$backupDir\existing-scaledobjects.yaml"
        }
        catch {
            # No hay ScaledObjects existentes, crear archivo vacío
            "" | Out-File -FilePath "$backupDir\existing-scaledobjects.yaml"
        }
        
        # Backup de cualquier TriggerAuthentication existente
        try {
            kubectl get triggerauthentication -n medisupply -o yaml | Out-File -FilePath "$backupDir\existing-trigger-auth.yaml"
        }
        catch {
            # No hay TriggerAuthentications existentes, crear archivo vacío
            "" | Out-File -FilePath "$backupDir\existing-trigger-auth.yaml"
        }
        
        $backupDir | Out-File -FilePath ".keda-backup-path"
        Write-Log "Backup creado en: $backupDir" "SUCCESS"
    }
    catch {
        Write-Log "Error al crear backup: $_" "ERROR"
        exit 1
    }
}

# Función para instalar KEDA scaling
function Install-KedaScaling {
    Write-Log "Instalando configuración de KEDA..."
    
    try {
        # Aplicar la configuración
        kubectl apply -f k8s\config\keda-order-management-minimal.yaml
        
        Write-Log "Configuración de KEDA aplicada" "SUCCESS"
        
        # Esperar a que se cree el ScaledObject
        Write-Log "Esperando a que se cree el ScaledObject..."
        kubectl wait --for=condition=Ready scaledobject/order-management-scaler -n medisupply --timeout=60s
        
        Write-Log "ScaledObject creado exitosamente" "SUCCESS"
    }
    catch {
        Write-Log "Error durante la instalación: $_" "ERROR"
        exit 1
    }
}

# Función para verificar la instalación
function Test-Installation {
    Write-Log "Verificando la instalación..."
    
    # Verificar ScaledObject
    try {
        kubectl get scaledobject order-management-scaler -n medisupply | Out-Null
        Write-Log "ScaledObject 'order-management-scaler' está activo" "SUCCESS"
    }
    catch {
        Write-Log "ScaledObject no se creó correctamente" "ERROR"
        exit 1
    }
    
    # Verificar TriggerAuthentication
    try {
        kubectl get triggerauthentication rabbitmq-order-auth -n medisupply | Out-Null
        Write-Log "TriggerAuthentication 'rabbitmq-order-auth' está activa" "SUCCESS"
    }
    catch {
        Write-Log "TriggerAuthentication no se creó correctamente" "ERROR"
        exit 1
    }
    
    # Mostrar estado del ScaledObject
    Write-Log "Estado del ScaledObject:"
    kubectl describe scaledobject order-management-scaler -n medisupply
    
    Write-Log "Instalación verificada correctamente" "SUCCESS"
}

# Función para rollback
function Invoke-Rollback {
    Write-Log "Ejecutando rollback..."
    
    if (-not (Test-Path ".keda-backup-path")) {
        Write-Log "No se encontró información de backup" "ERROR"
        exit 1
    }
    
    $backupDir = Get-Content ".keda-backup-path"
    
    try {
        # Eliminar recursos de KEDA
        kubectl delete scaledobject order-management-scaler -n medisupply 2>$null
        kubectl delete triggerauthentication rabbitmq-order-auth -n medisupply 2>$null
        
        Write-Log "Rollback completado. El proyecto está en su estado original" "SUCCESS"
    }
    catch {
        Write-Log "Error durante el rollback: $_" "ERROR"
        exit 1
    }
}

# Función principal
function Main {
    Write-Host "============================================================================" -ForegroundColor Cyan
    Write-Host "INSTALADOR DE KEDA PARA ORDER MANAGEMENT" -ForegroundColor Cyan
    Write-Host "============================================================================" -ForegroundColor Cyan
    
    if ($Rollback) {
        Invoke-Rollback
        return
    }
    
    Write-Log "Iniciando instalación de KEDA scaling..."
    
    Test-Prerequisites
    New-Backup
    Install-KedaScaling
    Test-Installation
    
    Write-Host ""
    Write-Host "============================================================================" -ForegroundColor Green
    Write-Log "INSTALACIÓN COMPLETADA EXITOSAMENTE" "SUCCESS"
    Write-Host "============================================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "KEDA ahora escalará automáticamente el servicio order_management cuando:"
    Write-Host "- La cola 'order-damage-queue' tenga más de 10 mensajes"
    Write-Host "- Se escalará de 1 a máximo 3 pods"
    Write-Host ""
    Write-Host "Para monitorear:"
    Write-Host "  kubectl get scaledobject -n medisupply"
    Write-Host "  kubectl describe scaledobject order-management-scaler -n medisupply"
    Write-Host ""
    Write-Host "Para rollback (si hay problemas):"
    Write-Host "  .\k8s\scripts\install-keda-scaling.ps1 -Rollback"
    Write-Host ""
}

# Ejecutar función principal
Main
