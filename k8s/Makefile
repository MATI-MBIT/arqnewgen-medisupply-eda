# Makefile para gestión de cluster local con Kind/Minikube y charts de Helm
# Requiere: kind/minikube, kubectl, helm, docker

CLUSTER_NAME := local-k8s
KUBECONFIG_PATH := ~/.kube/config
PROVIDER := $(or $(PROVIDER),kind)

.PHONY: help init deploy clean destroy kiali

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@echo "  init PROVIDER=kind     - Crear cluster local con Kind (default)"
	@echo "  init PROVIDER=minikube - Crear cluster local con Minikube"
	@echo "  deploy                 - Desplegar charts istio, gateway y keda"
	@echo "  kiali                  - Abrir Kiali dashboard (requiere Istio)"
	@echo "  clean                  - Eliminar charts del cluster"
	@echo "  destroy                - Eliminar cluster completamente"
	@echo ""
	@echo "Ejemplos:"
	@echo "  make init              # Usa Kind por defecto"
	@echo "  make init PROVIDER=kind"
	@echo "  make init PROVIDER=minikube"
	@echo "  make kiali             # Abrir Kiali dashboard"

init: ## Crear cluster local con Kind o Minikube
	@echo "🚀 Creando cluster local con $(PROVIDER)..."
ifeq ($(PROVIDER),kind)
	@if kind get clusters | grep -q $(CLUSTER_NAME); then \
		echo "⚠️  El cluster $(CLUSTER_NAME) ya existe"; \
	else \
		kind create cluster --name $(CLUSTER_NAME) --config=config/kind-config.yaml; \
	fi
	@echo "✅ Configurando kubectl context..."
	@kubectl config use-context kind-$(CLUSTER_NAME)
else ifeq ($(PROVIDER),minikube)
	@if minikube status -p $(CLUSTER_NAME) >/dev/null 2>&1; then \
		echo "⚠️  El cluster $(CLUSTER_NAME) ya existe"; \
	else \
		minikube start -p $(CLUSTER_NAME) \
			--driver=docker \
			--cpus=2 \
			--memory=6144 \
			--disk-size=20g \
			--kubernetes-version=stable \
			--addons=ingress,dashboard,metrics-server; \
	fi
	@echo "✅ Configurando kubectl context..."
	@kubectl config use-context $(CLUSTER_NAME)
else
	@echo "❌ PROVIDER debe ser 'kind' o 'minikube'"
	@exit 1
endif
	@echo "🎉 Cluster $(CLUSTER_NAME) con $(PROVIDER) listo!"

deploy: ## Desplegar charts istio (base, istiod, gateway) y keda
	@echo "📦 Desplegando charts en el cluster..."
	@echo "🔧 Verificando que el cluster esté activo..."
ifeq ($(PROVIDER),kind)
	@kubectl cluster-info --context kind-$(CLUSTER_NAME) > /dev/null
else
	@kubectl cluster-info --context $(CLUSTER_NAME) > /dev/null
endif
	
	@echo "📋 Agregando repositorios de Helm..."
	@helm repo add istio https://istio-release.storage.googleapis.com/charts || true
	@helm repo add kedacore https://kedacore.github.io/charts || true
	@helm repo update
	
	@echo "🚀 Desplegando Istio Base..."
	@helm upgrade --install istio-base ./istio/base \
		--namespace istio-system \
		--set defaultRevision=default \
		--create-namespace \
		--wait
	
	@echo "🔧 Desplegando Istiod..."
	@helm upgrade --install istiod ./istio/istiod \
		--namespace istio-system \
		--wait

	@echo "📦 AddOns no está instalado. Instalando AddOns..."; \
		kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.27/samples/addons/prometheus.yaml; \
		kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.27/samples/addons/jaeger.yaml; \
		kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.27/samples/addons/grafana.yaml; \
		kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.27/samples/addons/kiali.yaml; \
		echo "⏳ Esperando a que AddOns esté listo..."; \
		kubectl wait --for=condition=available --timeout=300s deployment/prometheus -n istio-system; \
		kubectl wait --for=condition=available --timeout=300s deployment/jaeger -n istio-system; \
		kubectl wait --for=condition=available --timeout=300s deployment/grafana -n istio-system; \
		kubectl wait --for=condition=available --timeout=300s deployment/kiali -n istio-system; \

	@echo "🌐 Desplegando Istio Gateway..."
	@helm upgrade --install istio-gateway ./istio/gateway --debug \
		--namespace istio-system \
		--set profile=demo \
		--wait
	
	@echo "⚡ Desplegando KEDA..."
	@helm upgrade --install keda ./keda \
		--namespace keda-system \
		--create-namespace \
		--wait
	
	@echo "✅ Todos los charts desplegados exitosamente!"

clean: ## Eliminar charts del cluster
	@echo "🧹 Eliminando charts del cluster..."
	@echo "❌ Eliminando KEDA..."
	@helm uninstall keda --namespace keda-system || true
	@kubectl delete namespace keda-system --ignore-not-found=true
	
	@echo "❌ Eliminando Istio Gateway..."
	@helm uninstall istio-gateway --namespace istio-system || true
	
	@echo "❌ Eliminando Istiod..."
	@helm uninstall istiod --namespace istio-system || true
	
	@echo "❌ Eliminando Istio Base..."
	@helm uninstall istio-base --namespace istio-system || true
	@kubectl delete namespace istio-system --ignore-not-found=true
	@kubectl delete namespace medisupply --ignore-not-found=true
	
	@echo "✅ Charts eliminados del cluster!"

destroy: ## Eliminar cluster completamente
	@echo "💥 Eliminando cluster $(CLUSTER_NAME)..."
ifeq ($(PROVIDER),kind)
	@if kind get clusters | grep -q $(CLUSTER_NAME); then \
		kind delete cluster --name $(CLUSTER_NAME); \
		echo "✅ Cluster $(CLUSTER_NAME) eliminado completamente!"; \
	else \
		echo "⚠️  El cluster $(CLUSTER_NAME) no existe"; \
	fi
else ifeq ($(PROVIDER),minikube)
	@if minikube status -p $(CLUSTER_NAME) >/dev/null 2>&1; then \
		minikube delete -p $(CLUSTER_NAME); \
		echo "✅ Cluster $(CLUSTER_NAME) eliminado completamente!"; \
	else \
		echo "⚠️  El cluster $(CLUSTER_NAME) no existe"; \
	fi
else
	@echo "❌ PROVIDER debe ser 'kind' o 'minikube'"
	@exit 1
endif