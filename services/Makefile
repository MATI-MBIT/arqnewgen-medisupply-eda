# Makefile para servicios - Construcción y publicación de imágenes Docker
# Requiere: docker

# Configuración
K8S_PROVIDER := $(or $(K8S_PROVIDER),kind)
CLUSTER_NAME := local-k8s
DOCKER_REGISTRY := $(or $(DOCKER_REGISTRY),)
IMAGE_TAG := $(or $(IMAGE_TAG),latest)

# Servicios disponibles
SERVICES := mqtt-event-generator mqtt-order-event-client warehouse/batch

.PHONY: help build push build-all push-all build-push build-push-all list-images clean-images setup-local-registry stop-local-registry load-to-k8s load-all-to-k8s

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@echo "  build SERVICE=<name>   - Construir imagen Docker para un servicio"
	@echo "  push SERVICE=<name>    - Publicar imagen Docker al registry"
	@echo "  build-all              - Construir todas las imágenes"
	@echo "  push-all               - Publicar todas las imágenes"
	@echo "  build-push SERVICE=<name> - Construir y publicar imagen"
	@echo "  build-push-all         - Construir y publicar todas las imágenes"
	@echo "  load-to-k8s SERVICE=<name> - Cargar imagen a cluster K8s local"
	@echo "  load-all-to-k8s        - Cargar todas las imágenes a K8s local"
	@echo "  build-load SERVICE=<name> - Construir y cargar a K8s local"
	@echo "  build-load-all         - Construir y cargar todas a K8s local"
	@echo "  list-images            - Listar imágenes construidas"
	@echo "  clean-images           - Eliminar imágenes Docker locales"
	@echo "  setup-local-registry   - Configurar registry local para desarrollo"
	@echo "  stop-local-registry    - Detener registry local"
	@echo ""
	@echo "Variables de entorno:"
	@echo "  K8S_PROVIDER           - Proveedor K8s (kind/minikube, default: kind)"
	@echo "  CLUSTER_NAME           - Nombre del cluster (default: local-k8s)"
	@echo "  DOCKER_REGISTRY        - Registry Docker (opcional para push)"
	@echo "  IMAGE_TAG              - Tag de la imagen (default: latest)"
	@echo ""
	@echo "Servicios disponibles: $(SERVICES)"
	@echo ""
	@echo "Ejemplos para desarrollo local:"
	@echo "  make build-load SERVICE=mqtt-event-generator  # Para Kind/Minikube"
	@echo "  make build-load-all                           # Todas las imágenes"
	@echo ""
	@echo "Ejemplos para registry externo:"
	@echo "  DOCKER_REGISTRY=my-registry.com make build-push SERVICE=mqtt-event-generator"

build: ## Construir imagen Docker para un servicio específico
	@if [ -z "$(SERVICE)" ]; then \
		echo "❌ Error: Debes especificar SERVICE=<nombre>"; \
		echo "Servicios disponibles: $(SERVICES)"; \
		exit 1; \
	fi
	@if [ ! -d "$(SERVICE)" ]; then \
		echo "❌ Error: El servicio '$(SERVICE)' no existe"; \
		exit 1; \
	fi
	@echo "🔨 Construyendo imagen para $(SERVICE)..."
	@if [ -n "$(DOCKER_REGISTRY)" ]; then \
		cd $(SERVICE) && docker build -t $(DOCKER_REGISTRY)/$(SERVICE):$(IMAGE_TAG) .; \
		echo "✅ Imagen construida: $(DOCKER_REGISTRY)/$(SERVICE):$(IMAGE_TAG)"; \
	else \
		cd $(SERVICE) && docker build -t $(SERVICE):$(IMAGE_TAG) .; \
		echo "✅ Imagen construida: $(SERVICE):$(IMAGE_TAG)"; \
	fi

push: ## Publicar imagen Docker al registry
	@if [ -z "$(SERVICE)" ]; then \
		echo "❌ Error: Debes especificar SERVICE=<nombre>"; \
		exit 1; \
	fi
	@echo "📤 Publicando imagen $(SERVICE) al registry..."
	@docker push $(DOCKER_REGISTRY)/$(SERVICE):$(IMAGE_TAG)
	@echo "✅ Imagen publicada: $(DOCKER_REGISTRY)/$(SERVICE):$(IMAGE_TAG)"

build-push: ## Construir y publicar imagen Docker
	@if [ -z "$(SERVICE)" ]; then \
		echo "❌ Error: Debes especificar SERVICE=<nombre>"; \
		exit 1; \
	fi
	@$(MAKE) build SERVICE=$(SERVICE)
	@$(MAKE) push SERVICE=$(SERVICE)
	@echo "✅ Imagen $(SERVICE) construida y publicada!"

build-all: ## Construir todas las imágenes
	@echo "🔨 Construyendo todas las imágenes..."
	@for service in $(SERVICES); do \
		echo "📦 Construyendo $$service..."; \
		$(MAKE) build SERVICE=$$service; \
	done
	@echo "✅ Todas las imágenes construidas!"

push-all: ## Publicar todas las imágenes
	@echo "📤 Publicando todas las imágenes..."
	@for service in $(SERVICES); do \
		echo "📤 Publicando $$service..."; \
		$(MAKE) push SERVICE=$$service; \
	done
	@echo "✅ Todas las imágenes publicadas!"

build-push-all: ## Construir y publicar todas las imágenes
	@echo "🔨📤 Construyendo y publicando todas las imágenes..."
	@for service in $(SERVICES); do \
		echo "🔨📤 Procesando $$service..."; \
		$(MAKE) build-push SERVICE=$$service; \
	done
	@echo "✅ Todas las imágenes construidas y publicadas!"

list-images: ## Listar imágenes construidas
	@echo "📋 Imágenes Docker construidas:"
	@for service in $(SERVICES); do \
		if docker images $(DOCKER_REGISTRY)/$$service --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" | grep -v REPOSITORY; then \
			echo ""; \
		fi; \
	done

clean-images: ## Eliminar imágenes Docker locales
	@echo "🧹 Eliminando imágenes Docker locales..."
	@for service in $(SERVICES); do \
		echo "🧹 Eliminando imágenes de $$service..."; \
		docker rmi $(DOCKER_REGISTRY)/$$service:$(IMAGE_TAG) 2>/dev/null || true; \
		docker rmi $$service:$(IMAGE_TAG) 2>/dev/null || true; \
	done
	@echo "✅ Imágenes locales eliminadas!"

# Registry local para desarrollo
setup-local-registry: ## Configurar registry local para desarrollo
	@echo "🐳 Configurando registry local..."
	@if ! docker ps | grep -q registry; then \
		docker run -d -p 5000:5000 --name registry registry:2; \
		echo "✅ Registry local iniciado en localhost:5000"; \
	else \
		echo "⚠️  Registry local ya está ejecutándose"; \
	fi

stop-local-registry: ## Detener registry local
	@echo "🛑 Deteniendo registry local..."
	@docker stop registry && docker rm registry || true
	@echo "✅ Registry local detenido"

# Cargar imágenes directamente a clusters K8s locales
load-to-k8s: ## Cargar imagen a cluster K8s local (Kind/Minikube)
	@if [ -z "$(SERVICE)" ]; then \
		echo "❌ Error: Debes especificar SERVICE=<nombre>"; \
		exit 1; \
	fi
	@echo "📦 Cargando imagen $(SERVICE) a cluster $(K8S_PROVIDER)..."
ifeq ($(K8S_PROVIDER),kind)
	@kind load docker-image $(SERVICE):$(IMAGE_TAG) --name $(CLUSTER_NAME)
	@echo "✅ Imagen $(SERVICE):$(IMAGE_TAG) cargada a Kind cluster $(CLUSTER_NAME)"
else ifeq ($(K8S_PROVIDER),minikube)
	@minikube image load $(SERVICE):$(IMAGE_TAG) -p $(CLUSTER_NAME)
	@echo "✅ Imagen $(SERVICE):$(IMAGE_TAG) cargada a Minikube cluster $(CLUSTER_NAME)"
else
	@echo "❌ K8S_PROVIDER debe ser 'kind' o 'minikube'"
	@exit 1
endif

load-all-to-k8s: ## Cargar todas las imágenes a K8s local
	@echo "📦 Cargando todas las imágenes a cluster $(K8S_PROVIDER)..."
	@for service in $(SERVICES); do \
		echo "📦 Cargando $$service..."; \
		$(MAKE) load-to-k8s SERVICE=$$service; \
	done
	@echo "✅ Todas las imágenes cargadas a $(K8S_PROVIDER)!"

build-load: ## Construir y cargar imagen a K8s local
	@if [ -z "$(SERVICE)" ]; then \
		echo "❌ Error: Debes especificar SERVICE=<nombre>"; \
		exit 1; \
	fi
	@$(MAKE) build SERVICE=$(SERVICE)
	@$(MAKE) load-to-k8s SERVICE=$(SERVICE)
	@echo "✅ Imagen $(SERVICE) construida y cargada a K8s!"

build-load-all: ## Construir y cargar todas las imágenes a K8s local
	@echo "🔨📦 Construyendo y cargando todas las imágenes a K8s..."
	@for service in $(SERVICES); do \
		echo "🔨📦 Procesando $$service..."; \
		$(MAKE) build-load SERVICE=$$service; \
	done
	@echo "✅ Todas las imágenes construidas y cargadas a K8s!"