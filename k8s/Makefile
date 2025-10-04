# Makefile para gestión de cluster local con Kind/Minikube y charts de Helm
# Requiere: kind/minikube, kubectl, helm, docker

CLUSTER_NAME := local-k8s
KUBECONFIG_PATH := ~/.kube/config
PROVIDER := $(or $(PROVIDER),kind)

.PHONY: help init deploy clean destroy kiali kafka-ui mqtt rabbitmq kafka-rabbitmq-replicator kafka-rabbitmq-replicator-logs kafka-rabbitmq-replicator-status create-kafka-topics wait-for-kafka create-topic-if-not-exists create-rabbitmq-resources wait-for-rabbitmq

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@echo "  init PROVIDER=kind     - Crear cluster local con Kind (default)"
	@echo "  init PROVIDER=minikube - Crear cluster local con Minikube"
	@echo "  deploy                 - Desplegar toda la arquitectura EDA completa"
	@echo "  status                 - Mostrar estado de todos los componentes"
	@echo "  kiali                  - Abrir Kiali dashboard (requiere Istio)"
	@echo "  kafka-ui               - Abrir Kafka UI dashboard"
	@echo ""
	@echo "  mqtt                   - Abrir EMQX dashboard"
	@echo "  rabbitmq               - Abrir RabbitMQ Management UI"
	@echo "  order-gateway          - Configurar Istio Gateway para Order Service"
	@echo "  clean                  - Eliminar charts del cluster"
	@echo "  destroy                - Eliminar cluster completamente"
	@echo ""
	@echo "🔄 Arquitectura desplegada:"
	@echo "  Flujo Principal: MQTT → Kafka (Principal)"
	@echo "  Cluster Adicional: Kafka Warehouse"
	@echo "  Replicación Bidireccional:"
	@echo "    • damage, events-sensor → Kafka Principal → Kafka Warehouse"
	@echo "    • warehouse-events, inventory-updates ← Kafka Principal ← Kafka Warehouse"
	@echo ""
	@echo "Ejemplos:"
	@echo "  make init              # Usa Kind por defecto"
	@echo "  make deploy            # Despliega arquitectura completa"
	@echo "  make status            # Ver estado de componentes"
	@echo "  make kafka-ui          # Abrir Kafka UI dashboard"
	@echo "  make mqtt              # Abrir EMQX dashboard"
	@echo "  make rabbitmq          # Abrir RabbitMQ Management UI"
	@echo "  make create-kafka-topics # Crear topics necesarios para replicación"
	@echo "  make create-rabbitmq-resources # Crear queues y exchanges de RabbitMQ"
	@echo "  kafka-rabbitmq-replicator-logs   # Ver logs del replicador Kafka-RabbitMQ"
	@echo "  kafka-rabbitmq-replicator-status # Estado del replicador Kafka-RabbitMQ"

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

ifeq ($(PROVIDER),kind)
	@echo "🔨 Construyendo y cargando imágenes Docker en Kind..."
	@$(MAKE) -C ../services build-load-all
endif

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

	@kubectl create namespace medilogistic --dry-run=client -o yaml | kubectl apply -f - || echo "✅ Namespace medilogistic ya existe o fue creado"
	@kubectl label namespace medilogistic istio-injection=enabled

	@kubectl create namespace mediorder --dry-run=client -o yaml | kubectl apply -f - || echo "✅ Namespace mediorder ya existe o fue creado"
	@kubectl label namespace mediorder istio-injection=enabled

	@kubectl create namespace mediwarehouse --dry-run=client -o yaml | kubectl apply -f - || echo "✅ Namespace mediwarehouse ya existe o fue creado"
	@kubectl label namespace mediwarehouse istio-injection=enabled

    # Medisupply - Main
	@echo "⚡ Desplegando Kafka (cluster principal)..."
	@helm upgrade --install kafka ./kafka \
		--namespace medisupply \
		--values ./config/kafka-values.yaml \
		--wait

	@echo "🖥️  Desplegando Kafka UI (ambos clusters)..."
	@helm upgrade --install kafka-ui ./kafka-ui \
		--namespace medisupply \
		--wait
	
	@echo "📡 Desplegando MQTT (EMQX)..."
	@helm upgrade --install emqx ./mqtt/emqx \
		--namespace medilogistic \
		--wait

	@echo "🔗 Desplegando MQTT-Kafka Bridge..."
	@helm upgrade --install mqtt-kafka-bridge ./mqtt-kafka-bridge \
		--namespace medisupply \
		--wait

	@echo "⚡ Desplegando Kafka Warehouse (cluster secundario)..."
	@helm upgrade --install kafka-warehouse ./kafka \
		--namespace mediwarehouse \
		--values ./config/kafka-warehouse-values.yaml \
		--wait

	@echo "� Creando topics necesarios para replicación..."
	@$(MAKE) create-kafka-topics

	@echo "🔄 Desplegando Kafka-Kafka Replicator (replicación bidireccional)..."
	@helm upgrade --install kafka-replicator ./kafka-replicator \
		--namespace medisupply \
		--wait --timeout 10m

	@echo "🐰 Desplegando RabbitMQ..."
	@helm upgrade --install rabbitmq ./rabbitmq \
		--namespace mediorder \
		--wait

	@echo "🐰 Copiar secreto de mediorder a medisupply"
	@if kubectl get secret rabbitmq -n medisupply >/dev/null 2>&1; then \
		echo "🔄 Secreto rabbitmq ya existe en medisupply, eliminando y recreando..."; \
		kubectl delete secret rabbitmq -n medisupply; \
		kubectl get secret rabbitmq -n mediorder -o yaml | \
		sed 's/namespace: mediorder/namespace: medisupply/' | \
		kubectl create -f -; \
	else \
		echo "📋 Creando secreto rabbitmq en medisupply..."; \
		kubectl get secret rabbitmq -n mediorder -o yaml | \
		sed 's/namespace: mediorder/namespace: medisupply/' | \
		kubectl create -f -; \
	fi

	@echo "� Cereando recursos de RabbitMQ necesarios para replicación..."
	@$(MAKE) create-rabbitmq-resources
	
	@echo "🔄 Desplegando Kafka-RabbitMQ Replicator (replicación bidireccional)..."
	@helm upgrade --install kafka-rabbitmq-replicator ./kafka-rabbitmq-replicator \
		--namespace medisupply \
		--wait

    # Logistic
	@echo "📊 Desplegando MQTT Event Generator..."
	@helm upgrade --install mqtt-event-generator ./microservice \
		--namespace medilogistic \
		--values ./config/services/mqtt-event-generator-values.yaml \
		--wait

	@echo "📊 Desplegando MQTT Order Event Client..."
	@helm upgrade --install mqtt-order-event-client ./microservice \
		--namespace medilogistic \
		--values ./config/services/mqtt-order-event-client-values.yaml \
		--wait

    # Order
	@echo "📦 Desplegando Order Management Service..."
	@helm upgrade --install order-service ./microservice \
		--namespace mediorder \
		--values ./config/services/order_management/order/order-service-values.yaml \
		--wait

    # Warehouse
	@echo "📦 Desplegando Warehouse Batch..."
	@helm upgrade --install warehouse-batch ./microservice \
		--namespace mediwarehouse \
		--values ./config/services/warehouse/batch/batch-service-values.yaml \
		--wait

    
	@echo ""
	@echo "🎉 ¡Despliegue completado exitosamente!"
	@echo "📋 Servicios desplegados:"
	@echo "  ✅ Kafka (Principal) + Kafka Warehouse + Kafka UI"
	@echo "  ✅ Topics de Kafka creados automáticamente"
	@echo "  ✅ EMQX (MQTT Broker)"
	@echo "  ✅ RabbitMQ + Queues y Exchanges creados automáticamente"
	@echo "  ✅ Order Management Service"
	@echo "  ✅ MQTT Event Generator"
	@echo "  ✅ MQTT Order Event Client"
	@echo "  ✅ Warehouse Batch"
	@echo "  ✅ MQTT-Kafka Bridge"
	@echo "  ✅ Kafka Replicator (Replicación Bidireccional)"
	@echo "  ✅ Kafka-RabbitMQ Replicator (Replicación Bidireccional)"
	@echo "  ✅ Strimzi Operator"
	@echo "  ✅ Kafka Pedidos (Cluster secundario)"
	@echo "  ✅ Kafka Connect + RabbitMQ Connector"
	@echo "  ✅ Kafka Replicator (Python-based)"
	@echo ""
	@echo ""
	@echo "🔄 Flujos de datos configurados:"
	@echo "  📡 MQTT → Kafka:"
	@echo "     mqtt-event-generator → EMQX → mqtt-order-event-client → EMQX → mqtt-kafka-bridge → Kafka"
	@echo "  🔄 Replicación Bidireccional (Kafka Replicator):"
	@echo "     Kafka ⟷ Kafka Warehouse"
	@echo "     • events-order-damage → warehouse-order-damage"
	@echo "     • events-sensor → warehouse-sensor"
	@echo "     • order-events → order-events"
	@echo "  🔄 Replicación Bidireccional (Kafka-RabbitMQ Replicator):"
	@echo "     Kafka ⟷ RabbitMQ"
	@echo "     • events-order-damage → events exchange → order-damage-queue"
	@echo "     • order-events-queue → order-events topic"
	@echo ""
	@echo "💡 Para acceder a los dashboards:"
	@echo "  make kafka-ui    # Kafka UI (ambos clusters) en http://localhost:9090"
	@echo "  make mqtt        # EMQX en http://localhost:18083"
	@echo "  make rabbitmq    # RabbitMQ en http://localhost:15672"
	@echo ""

clean: ## Eliminar charts del cluster
	@echo "🧹 Eliminando charts del cluster..."
	@echo ""
    # Logistic
	@echo "📊 Eliminando MQTT Event Generator..."
	@helm uninstall mqtt-event-generator --namespace medilogistic || true
	@echo "📊 Eliminando MQTT Order Event Client..."
	@helm uninstall mqtt-order-event-client --namespace medilogistic || true
	@echo "📡 Eliminando EMQX..."
	@helm uninstall emqx --namespace medilogistic || true

    # Order
	@echo "� Eliminnando Order Management Service..."
	@helm uninstall order-service --namespace mediorder || true
	@echo "🐰 Eliminando RabbitMQ..."
	@helm uninstall rabbitmq --namespace mediorder || true

    # Warehouse
	@helm uninstall kafka-warehouse --namespace mediwarehouse || true
	@echo "⚡ Eliminando Kafka..."
	@echo "📊 Eliminando Warehouse Batch..."
	@helm uninstall warehouse-batch --namespace mediwarehouse || true

    # MediSupply
	@echo "🔄 Eliminando Kafka Replicator..."
	@helm uninstall kafka-replicator --namespace medisupply || true
	@echo "🔄 Eliminando Kafka-RabbitMQ Replicator..."
	@helm uninstall kafka-rabbitmq-replicator --namespace medisupply || true
	@echo "🔗 Eliminando MQTT-Kafka Bridge..."
	@helm uninstall mqtt-kafka-bridge --namespace medisupply || true
	@echo "🖥️  Eliminando Kafka UI..."
	@helm uninstall kafka-ui --namespace medisupply || true
	@echo "⚡ Eliminando Kafka Warehouse..."
	@helm uninstall kafka --namespace medisupply || true
	@echo ""
	@echo "🗑️  Eliminando namespace medisupply..."
	@kubectl delete namespace medisupply --ignore-not-found=true
	@kubectl delete namespace medilogistic --ignore-not-found=true
	@kubectl delete namespace mediorder --ignore-not-found=true
	@kubectl delete namespace mediwarehouse --ignore-not-found=true
	@echo ""
	@echo "✅ Todos los charts eliminados del cluster!"
	@echo "📋 Servicios eliminados:"
	@echo "  ❌ Warehouse Batch"
	@echo "  ❌ MQTT Event Generator"
	@echo "  ❌ MQTT Order Event Client"
	@echo "  ❌ MQTT-Kafka Bridge"
	@echo "  ❌ Order Management Service"
	@echo "  ❌ RabbitMQ"
	@echo "  ❌ EMQX (MQTT Broker)"
	@echo "  ❌ Kafka UI"
	@echo "  ❌ Kafka"
	@echo "  ❌ Kafka Pedidos"
	@echo "  ❌ Kafka Connect + Conectores"
	@echo "  ❌ Kafka Replicator"
	@echo "  ❌ Kafka-RabbitMQ Replicator"
	@echo "  ❌ Strimzi Operator"
	@echo "  ❌ Namespace medisupply"
	@echo "  ❌ Namespace medilogistic"
	@echo "  ❌ Namespace mediorder"
	@echo "  ❌ Namespace mediwarehouse"

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
	@kubectl port-forward -n mediorder svc/rabbitmq 15672:15672 &
	@sleep 3
	@echo "🌐 RabbitMQ Management UI disponible en: http://localhost:15672"
	@echo "💡 Usuario: user, Contraseña: $(shell kubectl get secret --namespace medisupply rabbitmq -o jsonpath="{.data.rabbitmq-password}" | base64 -d)"
	@echo "💡 Presiona Ctrl+C para detener el port-forward"

mqtt: ## Abrir EMQX dashboard
	@echo "📡 Abriendo EMQX dashboard..."
	@echo "📡 Configurando port-forward para EMQX..."
	@kubectl port-forward -n medilogistic svc/emqx 18083:18083 &
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

status: ## Mostrar estado de todos los componentes
	@echo "📊 Estado de los componentes del sistema:"
	@echo ""
	@echo "🔍 Pods en namespace medisupply:"
	@kubectl get pods -n medisupply -o wide || echo "❌ Error obteniendo pods"
	@echo ""
	@echo "🔍 Pods en namespace medilogictic:"
	@kubectl get pods -n medilogistic -o wide || echo "❌ Error obteniendo pods"
	@echo ""
	@echo "🔍 Pods en namespace mediorder:"
	@kubectl get pods -n mediorder -o wide || echo "❌ Error obteniendo pods"
	@echo ""
	@echo "🔍 Pods en namespace mediwarehouse:"
	@kubectl get pods -n mediwarehouse -o wide || echo "❌ Error obteniendo pods"
	@echo ""
	@echo "🔍 Servicios medisupply:"
	@kubectl get svc -n medisupply || echo "❌ Error obteniendo servicios"
	@echo ""
	@echo "🔍 Servicios medilogistic:"
	@kubectl get svc -n medilogistic || echo "❌ Error obteniendo servicios"
	@echo ""
	@echo "🔍 Servicios mediorder:"
	@kubectl get svc -n mediorder || echo "❌ Error obteniendo servicios"
	@echo ""
	@echo "🔍 Servicios mediwarehouse:"
	@kubectl get svc -n mediwarehouse || echo "❌ Error obteniendo servicios"

wait-for-kafka: ## Esperar a que Kafka esté disponible
	@echo "⏳ Esperando a que Kafka Principal esté listo..."
	@for i in $$(seq 1 30); do \
		if kubectl exec -n medisupply kafka-controller-0 -- kafka-topics.sh --bootstrap-server localhost:9092 --list >/dev/null 2>&1; then \
			echo "✅ Kafka Principal está listo"; \
			break; \
		else \
			echo "⏳ Esperando Kafka Principal (intento $$i/30)..."; \
			sleep 10; \
		fi; \
		if [ $$i -eq 30 ]; then \
			echo "❌ Timeout esperando Kafka Principal"; \
			exit 1; \
		fi; \
	done
	
	@echo "⏳ Esperando a que Kafka Warehouse esté listo..."
	@for i in $$(seq 1 30); do \
		if kubectl exec -n mediwarehouse kafka-warehouse-controller-0 -- kafka-topics.sh --bootstrap-server localhost:9092 --list >/dev/null 2>&1; then \
			echo "✅ Kafka Warehouse está listo"; \
			break; \
		else \
			echo "⏳ Esperando Kafka Warehouse (intento $$i/30)..."; \
			sleep 10; \
		fi; \
		if [ $$i -eq 30 ]; then \
			echo "❌ Timeout esperando Kafka Warehouse"; \
			exit 1; \
		fi; \
	done

create-topic-if-not-exists: ## Crear topic si no existe (requiere KAFKA_POD, NAMESPACE, TOPIC_NAME)
	@echo "📋 Verificando topic $(TOPIC_NAME) en $(KAFKA_POD)..."
	@if kubectl exec -n $(NAMESPACE) $(KAFKA_POD) -- kafka-topics.sh \
		--bootstrap-server localhost:9092 \
		--list | grep -q "^$(TOPIC_NAME)$$"; then \
		echo "✅ Topic $(TOPIC_NAME) ya existe"; \
	else \
		echo "📝 Creando topic $(TOPIC_NAME)..."; \
		kubectl exec -n $(NAMESPACE) $(KAFKA_POD) -- kafka-topics.sh \
			--bootstrap-server localhost:9092 \
			--create \
			--topic $(TOPIC_NAME) \
			--partitions 1 \
			--replication-factor 1 \
			--config retention.ms=604800000 \
			--config cleanup.policy=delete; \
		echo "✅ Topic $(TOPIC_NAME) creado exitosamente"; \
	fi

create-kafka-topics: wait-for-kafka ## Crear todos los topics necesarios para replicación
	@echo "📋 Creando topics necesarios para replicación..."
	@echo ""
	
	@echo "🔹 Creando topics en Kafka Principal (medisupply)..."
	@$(MAKE) create-topic-if-not-exists KAFKA_POD=kafka-controller-0 NAMESPACE=medisupply TOPIC_NAME=events-order-damage
	@$(MAKE) create-topic-if-not-exists KAFKA_POD=kafka-controller-0 NAMESPACE=medisupply TOPIC_NAME=events-sensor
	@$(MAKE) create-topic-if-not-exists KAFKA_POD=kafka-controller-0 NAMESPACE=medisupply TOPIC_NAME=order-events
	
	@echo ""
	@echo "🔹 Creando topics en Kafka Warehouse (mediwarehouse)..."
	@$(MAKE) create-topic-if-not-exists KAFKA_POD=kafka-warehouse-controller-0 NAMESPACE=mediwarehouse TOPIC_NAME=warehouse-order-damage
	@$(MAKE) create-topic-if-not-exists KAFKA_POD=kafka-warehouse-controller-0 NAMESPACE=mediwarehouse TOPIC_NAME=warehouse-order-events
	
	@echo ""
	@echo "✅ Todos los topics creados exitosamente!"
	@echo "📋 Topics creados:"
	@echo "  🔸 Kafka Principal:"
	@echo "    • events-order-damage"
	@echo "    • events-sensor" 
	@echo "    • order-events"
	@echo "  🔸 Kafka Warehouse:"
	@echo "    • warehouse-order-damage"
	@echo "    • warehouse-sensor"
	@echo "    • order-events"

wait-for-rabbitmq: ## Esperar a que RabbitMQ esté disponible
	@echo "⏳ Esperando a que RabbitMQ esté listo..."
	@for i in $$(seq 1 30); do \
		if kubectl exec -n mediorder rabbitmq-0 -- rabbitmqctl status >/dev/null 2>&1; then \
			echo "✅ RabbitMQ está listo"; \
			break; \
		else \
			echo "⏳ Esperando RabbitMQ (intento $$i/30)..."; \
			sleep 10; \
		fi; \
		if [ $$i -eq 30 ]; then \
			echo "❌ Timeout esperando RabbitMQ"; \
			exit 1; \
		fi; \
	done

create-rabbitmq-resources: wait-for-rabbitmq ## Crear exchanges, queues y bindings necesarios para replicación
	@echo "🐰 Creando recursos de RabbitMQ necesarios para replicación..."
	@echo ""
	
	@echo "🔹 Obteniendo credenciales de RabbitMQ..."
	@RABBITMQ_PASSWORD=$$(kubectl get secret --namespace mediorder rabbitmq -o jsonpath="{.data.rabbitmq-password}" | base64 -d); \
	echo "🔹 Creando exchange 'events'..."; \
	kubectl exec -n mediorder rabbitmq-0 -- rabbitmqadmin \
		--host=localhost \
		--port=15672 \
		--username=user \
		--password=$$RABBITMQ_PASSWORD \
		declare exchange \
		name=events \
		type=topic \
		durable=true || echo "⚠️  Exchange 'events' ya existe o error al crear"; \
	\
	echo "🔹 Creando queue 'order-damage-queue'..."; \
	kubectl exec -n mediorder rabbitmq-0 -- rabbitmqadmin \
		--host=localhost \
		--port=15672 \
		--username=user \
		--password=$$RABBITMQ_PASSWORD \
		declare queue \
		name=order-damage-queue \
		durable=true || echo "⚠️  Queue 'order-damage-queue' ya existe o error al crear"; \
	\
	echo "🔹 Creando queue 'order-events-queue'..."; \
	kubectl exec -n mediorder rabbitmq-0 -- rabbitmqadmin \
		--host=localhost \
		--port=15672 \
		--username=user \
		--password=$$RABBITMQ_PASSWORD \
		declare queue \
		name=order-events-queue \
		durable=true || echo "⚠️  Queue 'order-events-queue' ya existe o error al crear"; \
	\
	echo "🔹 Creando binding 'order-damage-queue' → 'events' exchange..."; \
	kubectl exec -n mediorder rabbitmq-0 -- rabbitmqadmin \
		--host=localhost \
		--port=15672 \
		--username=user \
		--password=$$RABBITMQ_PASSWORD \
		declare binding \
		source=events \
		destination=order-damage-queue \
		routing_key=order.damage || echo "⚠️  Binding ya existe o error al crear"; \
	\
	echo "🔹 Creando binding 'order-events-queue' → 'events' exchange..."; \
	kubectl exec -n mediorder rabbitmq-0 -- rabbitmqadmin \
		--host=localhost \
		--port=15672 \
		--username=user \
		--password=$$RABBITMQ_PASSWORD \
		declare binding \
		source=events \
		destination=order-events-queue \
		routing_key=order.events || echo "⚠️  Binding ya existe o error al crear"; \
	\
	echo ""; \
	echo "✅ Recursos de RabbitMQ creados exitosamente!"; \
	echo "📋 Recursos creados:"; \
	echo "  🔸 Exchange: events (topic)"; \
	echo "  🔸 Queues:"; \
	echo "    • order-damage-queue (durable)"; \
	echo "    • order-events-queue (durable)"; \
	echo "  🔸 Bindings:"; \
	echo "    • order-damage-queue ← events [order.damage]"
