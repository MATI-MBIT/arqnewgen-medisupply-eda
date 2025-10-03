# Makefile para el chart genérico microservice

# Variables
CHART_NAME = microservice
NAMESPACE = default

# Comandos de ayuda
.PHONY: help
help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Validación y testing
.PHONY: lint
lint: ## Validar sintaxis del chart
	helm lint .

.PHONY: template
template: ## Generar templates sin instalar
	helm template test-release . --debug

.PHONY: template-mqtt-event-generator
template-mqtt-event-generator: ## Generar templates para MQTT Event Generator
	helm template mqtt-event-generator . -f examples/values-mqtt-event-generator.yaml --debug

.PHONY: template-mqtt-order-event-client
template-mqtt-order-event-client: ## Generar templates para MQTT Order Event Client
	helm template mqtt-order-event-client . -f examples/values-mqtt-order-event-client.yaml --debug

.PHONY: template-api-service
template-api-service: ## Generar templates para API Service
	helm template api-service . -f examples/values-api-service.yaml --debug

# Instalación
.PHONY: install-mqtt-event-generator
install-mqtt-event-generator: ## Instalar MQTT Event Generator (usando configuración validada)
	helm install mqtt-event-generator . -f config/mqtt-event-generator-values.yaml --namespace $(NAMESPACE) --create-namespace

.PHONY: install-mqtt-event-generator-example
install-mqtt-event-generator-example: ## Instalar MQTT Event Generator (usando ejemplo)
	helm install mqtt-event-generator . -f examples/values-mqtt-event-generator.yaml --namespace $(NAMESPACE) --create-namespace

.PHONY: install-mqtt-order-event-client
install-mqtt-order-event-client: ## Instalar MQTT Order Event Client (usando configuración validada)
	helm install mqtt-order-event-client . -f config/mqtt-order-event-client-values.yaml --namespace $(NAMESPACE) --create-namespace

.PHONY: install-mqtt-order-event-client-example
install-mqtt-order-event-client-example: ## Instalar MQTT Order Event Client (usando ejemplo)
	helm install mqtt-order-event-client . -f examples/values-mqtt-order-event-client.yaml --namespace $(NAMESPACE) --create-namespace

.PHONY: install-api-service
install-api-service: ## Instalar API Service
	helm install api-service . -f examples/values-api-service.yaml --namespace $(NAMESPACE) --create-namespace

# Actualización
.PHONY: upgrade-mqtt-event-generator
upgrade-mqtt-event-generator: ## Actualizar MQTT Event Generator (usando configuración validada)
	helm upgrade mqtt-event-generator . -f config/mqtt-event-generator-values.yaml --namespace $(NAMESPACE)

.PHONY: upgrade-mqtt-event-generator-example
upgrade-mqtt-event-generator-example: ## Actualizar MQTT Event Generator (usando ejemplo)
	helm upgrade mqtt-event-generator . -f examples/values-mqtt-event-generator.yaml --namespace $(NAMESPACE)

.PHONY: upgrade-mqtt-order-event-client
upgrade-mqtt-order-event-client: ## Actualizar MQTT Order Event Client (usando configuración validada)
	helm upgrade mqtt-order-event-client . -f config/mqtt-order-event-client-values.yaml --namespace $(NAMESPACE)

.PHONY: upgrade-mqtt-order-event-client-example
upgrade-mqtt-order-event-client-example: ## Actualizar MQTT Order Event Client (usando ejemplo)
	helm upgrade mqtt-order-event-client . -f examples/values-mqtt-order-event-client.yaml --namespace $(NAMESPACE)

.PHONY: upgrade-api-service
upgrade-api-service: ## Actualizar API Service
	helm upgrade api-service . -f examples/values-api-service.yaml --namespace $(NAMESPACE)

# Desinstalación
.PHONY: uninstall-mqtt-event-generator
uninstall-mqtt-event-generator: ## Desinstalar MQTT Event Generator
	helm uninstall mqtt-event-generator --namespace $(NAMESPACE)

.PHONY: uninstall-mqtt-order-event-client
uninstall-mqtt-order-event-client: ## Desinstalar MQTT Order Event Client
	helm uninstall mqtt-order-event-client --namespace $(NAMESPACE)

.PHONY: uninstall-api-service
uninstall-api-service: ## Desinstalar API Service
	helm uninstall api-service --namespace $(NAMESPACE)

# Utilidades
.PHONY: list
list: ## Listar releases instalados
	helm list --namespace $(NAMESPACE)

.PHONY: status-mqtt-event-generator
status-mqtt-event-generator: ## Ver estado de MQTT Event Generator
	helm status mqtt-event-generator --namespace $(NAMESPACE)

.PHONY: status-mqtt-order-event-client
status-mqtt-order-event-client: ## Ver estado de MQTT Order Event Client
	helm status mqtt-order-event-client --namespace $(NAMESPACE)

.PHONY: status-api-service
status-api-service: ## Ver estado de API Service
	helm status api-service --namespace $(NAMESPACE)

.PHONY: clean
clean: ## Limpiar todos los releases
	@echo "¿Estás seguro de que quieres desinstalar todos los releases? [y/N]" && read ans && [ $${ans:-N} = y ]
	helm uninstall mqtt-event-generator --namespace $(NAMESPACE) --ignore-not-found
	helm uninstall mqtt-order-event-client --namespace $(NAMESPACE) --ignore-not-found
	helm uninstall api-service --namespace $(NAMESPACE) --ignore-not-found

# Desarrollo
.PHONY: dry-run-mqtt-event-generator
dry-run-mqtt-event-generator: ## Dry run para MQTT Event Generator (configuración validada)
	helm install mqtt-event-generator . -f config/mqtt-event-generator-values.yaml --dry-run --debug

.PHONY: validate-mqtt-event-generator
validate-mqtt-event-generator: ## Validar configuración de MQTT Event Generator
	@echo "=== Generando template con configuración validada ==="
	helm template mqtt-event-generator . -f config/mqtt-event-generator-values.yaml | grep -A 10 "env:"
	@echo ""
	@echo "=== Validación completada ==="

.PHONY: dry-run-mqtt-order-event-client
dry-run-mqtt-order-event-client: ## Dry run para MQTT Order Event Client (configuración validada)
	helm install mqtt-order-event-client . -f config/mqtt-order-event-client-values.yaml --dry-run --debug

.PHONY: validate-mqtt-order-event-client
validate-mqtt-order-event-client: ## Validar configuración de MQTT Order Event Client
	@echo "=== Generando template con configuración validada ==="
	helm template mqtt-order-event-client . -f config/mqtt-order-event-client-values.yaml | grep -A 10 "env:"
	@echo ""
	@echo "=== Validación completada ==="

.PHONY: validate-all
validate-all: ## Validar todas las configuraciones
	@echo "============================================================================"
	@echo "VALIDACIÓN COMPLETA - CHART GENÉRICO MICROSERVICE"
	@echo "============================================================================"
	@echo ""
	@echo "=== MQTT EVENT GENERATOR ==="
	@$(MAKE) validate-mqtt-event-generator
	@echo ""
	@echo "=== MQTT ORDER EVENT CLIENT ==="
	@$(MAKE) validate-mqtt-order-event-client
	@echo ""
	@echo "============================================================================"
	@echo "✅ VALIDACIÓN COMPLETA - AMBOS MICROSERVICIOS VALIDADOS"
	@echo "============================================================================"

.PHONY: dry-run-api-service
dry-run-api-service: ## Dry run para API Service
	helm install api-service . -f examples/values-api-service.yaml --dry-run --debug