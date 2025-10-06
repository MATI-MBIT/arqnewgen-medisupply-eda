#!/bin/bash
# ============================================================================
# PRUEBA RÁPIDA DE KEDA - SCRIPT SIMPLE
# ============================================================================
# Script simple para verificar rápidamente si KEDA está funcionando
#
# USO:
# ./k8s/scripts/quick-test-keda.sh
# ============================================================================

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== VERIFICACIÓN RÁPIDA DE KEDA ===${NC}"
echo ""

# 1. Verificar que KEDA está instalado
echo -e "${BLUE}1. Verificando KEDA...${NC}"
if kubectl get crd scaledobjects.keda.sh &> /dev/null; then
    echo -e "${GREEN}✓ KEDA está instalado${NC}"
else
    echo -e "${RED}✗ KEDA no está instalado${NC}"
    exit 1
fi

# 2. Verificar ScaledObject
echo -e "\n${BLUE}2. Verificando ScaledObject...${NC}"
if kubectl get scaledobject order-management-scaler -n medisupply &> /dev/null; then
    echo -e "${GREEN}✓ ScaledObject existe${NC}"
    kubectl get scaledobject order-management-scaler -n medisupply -o wide
else
    echo -e "${RED}✗ ScaledObject no existe${NC}"
    echo "Instala con: ./k8s/config/install-keda-linux.sh"
    exit 1
fi

# 3. Verificar pods actuales
echo -e "\n${BLUE}3. Estado actual de pods...${NC}"
kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management

# 4. Verificar estado del ScaledObject
echo -e "\n${BLUE}4. Estado detallado del ScaledObject...${NC}"
kubectl describe scaledobject order-management-scaler -n medisupply | grep -E "(Status|Ready|Active|Replicas)"

# 5. Verificar conectividad a RabbitMQ
echo -e "\n${BLUE}5. Verificando conectividad a RabbitMQ...${NC}"
if kubectl get service rabbitmq -n mediorder &> /dev/null; then
    echo -e "${GREEN}✓ RabbitMQ accesible${NC}"
else
    echo -e "${RED}✗ RabbitMQ no accesible${NC}"
fi

echo ""
echo -e "${BLUE}=== RESUMEN ===${NC}"
echo "Para probar el escalado:"
echo "1. Envía más de 10 mensajes a la cola 'order-damage-queue'"
echo "2. Observa que los pods escalen de 1 a 3"
echo "3. Cuando la cola se vacíe, volverá a 1 pod"
echo ""
echo "Comandos de monitoreo:"
echo "  kubectl get scaledobject -n medisupply -w"
echo "  kubectl get pods -n medisupply -l app.kubernetes.io/name=order-management -w"
echo ""
echo "Para prueba completa: ./k8s/scripts/test-keda-scaling.sh"
