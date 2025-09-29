# Makefile para gestión de cluster local con Kind/Minikube y charts de Helm
# Requiere: kind/minikube, kubectl, helm, docker

CLUSTER_NAME := local-k8s
KUBECONFIG_PATH := ~/.kube/config
PROVIDER := $(or $(PROVIDER),kind)

.PHONY: help init deploy clean destroy kiali kafka-ui mqtt rabbitmq

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@echo "  init PROVIDER=kind     - Crear cluster local con Kind (default)"
	@echo "  init PROVIDER=minikube - Crear cluster local con Minikube"
	@echo "  deploy                 - Desplegar Kafka, Kafka UI, EMQX, RabbitMQ y MQTT Event Generator"
	@echo "  kiali                  - Abrir Kiali dashboard (requiere Istio)"
	@echo "  kafka-ui               - Abrir Kafka UI dashboard"
	@echo "  mqtt                   - Abrir EMQX dashboard"
	@echo "  rabbitmq               - Abrir RabbitMQ Management UI"
	@echo "  clean                  - Eliminar charts del cluster"
	@echo "  destroy                - Eliminar cluster completamente"
	@echo ""
	@echo "Ejemplos:"
	@echo "  make init              # Usa Kind por defecto"
	@echo "  make init PROVIDER=kind"
	@echo "  make init PROVIDER=minikube"
	@echo "  make kiali             # Abrir Kiali dashboard"
	@echo "  make kafka-ui          # Abrir Kafka UI dashboard"
	@echo "  make mqtt              # Abrir EMQX dashboard"
	@echo "  make rabbitmq          # Abrir RabbitMQ Management UI"

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

	@echo "📦 Desplegando charts en el cluster..."
	@echo "🔧 Verificando que el cluster esté activo..."
ifeq ($(PROVIDER),kind)
	@kubectl cluster-info --context kind-$(CLUSTER_NAME) > /dev/null
else
	@kubectl cluster-info --context $(CLUSTER_NAME) > /dev/null
endif
	
	@echo "📋 Agregando repositorios de Helm..."
#	@helm repo add istio https://istio-release.storage.googleapis.com/charts || true
#	@helm repo add kedacore https://kedacore.github.io/charts || true
#	@helm repo update
	
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
	@helm upgrade --install istio-gateway ./istio/gateway \
		--namespace istio-system \
		--set profile=demo \
		--wait
	
	@echo "⚡ Desplegando KEDA..."
	@helm upgrade --install keda ./keda \
		--namespace keda-system \
		--create-namespace \
		--wait
	
	@echo "✅ Todos los charts desplegados exitosamente!"

deploy:
	@kubectl create namespace medisupply --dry-run=client -o yaml | kubectl apply -f - || echo "✅ Namespace medisupply ya existe o fue creado"
	@kubectl label namespace medisupply istio-injection=enabled
	@echo "⚡ Desplegando Kafka..."
	@helm upgrade --install kafka ./kafka \
		--namespace medisupply \
		--values ./config/kafka-values.yaml \
		--wait

	@echo "🖥️  Desplegando Kafka UI..."
	@helm upgrade --install kafka-ui ./kafka-ui \
		--namespace medisupply \
		--wait

	@echo "📡 Desplegando MQTT (EMQX)..."
	@helm upgrade --install emqx ./mqtt/emqx \
		--namespace medisupply \
		--wait

	@echo "🐰 Desplegando RabbitMQ..."
	@helm upgrade --install rabbitmq ./rabbitmq \
		--namespace medisupply \
		--wait

	@echo "📊 Desplegando MQTT Event Generator..."
	@helm upgrade --install mqtt-event-generator ./mqtt-event-generator \
		--namespace medisupply \
	
	@echo "📊 Desplegando MQTT Order Event Client..."
	@helm upgrade --install mqtt-order-event-client ./mqtt-order-event-client \
		--namespace medisupply \

	@echo "🔗 Desplegando MQTT-Kafka Bridge..."
	@helm upgrade --install mqtt-kafka-bridge ./mqtt-kafka-bridge \
		--namespace medisupply \

	@echo ""
	@echo "🎉 ¡Despliegue completado exitosamente!"
	@echo "📋 Servicios desplegados:"
	@echo "  ✅ Kafka + Kafka UI"
	@echo "  ✅ EMQX (MQTT Broker)"
	@echo "  ✅ RabbitMQ"
	@echo "  ✅ MQTT Event Generator"
	@echo "  ✅ MQTT-Kafka Bridge"
	@echo ""
	@echo "💡 Para acceder a los dashboards:"
	@echo "  make kafka-ui    # Kafka UI en http://localhost:9090"
	@echo "  make mqtt        # EMQX en http://localhost:18083"
	@echo "  make rabbitmq    # RabbitMQ en http://localhost:15672"

clean: ## Eliminar charts del cluster
	@echo "🧹 Eliminando charts del cluster..."
	@echo ""
	@echo "📊 Eliminando MQTT Event Generator..."
	@helm uninstall mqtt-event-generator --namespace medisupply || true
	@echo "🔗 Eliminando MQTT-Kafka Bridge..."
	@helm uninstall mqtt-kafka-bridge --namespace medisupply || true
	@echo "🐰 Eliminando RabbitMQ..."
	@helm uninstall rabbitmq --namespace medisupply || true
	@echo "📡 Eliminando EMQX..."
	@helm uninstall emqx --namespace medisupply || true
	@echo "🖥️  Eliminando Kafka UI..."
	@helm uninstall kafka-ui --namespace medisupply || true
	@echo "⚡ Eliminando Kafka..."
	@helm uninstall kafka --namespace medisupply || true
	@echo ""
	@echo "🗑️  Eliminando namespace medisupply..."
	@kubectl delete namespace medisupply --ignore-not-found=true
	@echo ""
	@echo "✅ Todos los charts eliminados del cluster!"
	@echo "📋 Servicios eliminados:"
	@echo "  ❌ MQTT Event Generator"
	@echo "  ❌ Kafka Connect"
	@echo "  ❌ RabbitMQ"
	@echo "  ❌ EMQX (MQTT Broker)"
	@echo "  ❌ Kafka UI"
	@echo "  ❌ Kafka"
	@echo "  ❌ Namespace medisupply"

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

rabbitmq: ## Abrir RabbitMQ Management UI
	@echo "🐰 Abriendo RabbitMQ Management UI..."
	@echo "📡 Configurando port-forward para RabbitMQ..."
	@kubectl port-forward -n medisupply svc/rabbitmq 15672:15672 &
	@sleep 3
	@echo "🌐 RabbitMQ Management UI disponible en: http://localhost:15672"
	@echo "💡 Usuario: user, Contraseña: $(shell kubectl get secret --namespace medisupply rabbitmq -o jsonpath="{.data.rabbitmq-password}" | base64 -d)"
	@echo "💡 Presiona Ctrl+C para detener el port-forward"

mqtt: ## Abrir EMQX dashboard
	@echo "📡 Abriendo EMQX dashboard..."
	@echo "📡 Configurando port-forward para EMQX..."
	@kubectl port-forward -n medisupply svc/emqx 18083:18083 &
	@sleep 3
	@echo "🌐 EMQX disponible en: http://localhost:18083"
	@echo "💡 Usuario: admin, Contraseña: public"
	@echo "💡 Presiona Ctrl+C para detener el port-forward"

kafka-ui: ## Abrir Kafka UI dashboard
	@echo "🖥️  Abriendo Kafka UI dashboard..."
	@echo "📡 Configurando port-forward para Kafka UI..."
	@kubectl port-forward -n medisupply svc/kafka-ui 9090:8080 &
	@sleep 3
	@echo "🌐 Kafka UI disponible en: http://localhost:9090"
	@echo "💡 Presiona Ctrl+C para detener el port-forward"

kiali: ## Abrir Kiali dashboard
	@echo "🔍 Abriendo Kiali dashboard..."
	@echo "📡 Configurando port-forward para Kiali..."
	@kubectl port-forward -n istio-system svc/kiali 20001:20001 &
	@sleep 3
	@echo "🌐 Kiali disponible en: http://localhost:20001"
	@echo "💡 Presiona Ctrl+C para detener el port-forward"
