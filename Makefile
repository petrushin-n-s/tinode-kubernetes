KUBECTL = kubectl
NAMESPACE = default
DEPLOYMENTS = tinode-deployment.yaml go-app.yaml
POD_NAME = mongo # pod label
MONGO = mongo-statefulset.yaml
PVC = mongo-data-mongodb-0

.PHONY: all
all: deploy


.PHONY: deploy
deploy: mongo-init build
	@echo "Deploying applications..."
	@for deploy in $(DEPLOYMENTS); do \
		$(KUBECTL) apply -f $$deploy -n $(NAMESPACE); \
	done

# Шаг деплоя
.PHONY: build
build:
	@echo "Building applications..."
	@eval $(minikube -p minikube docker-env)
	@docker build -t go-app:latest .


.PHONY: mongo
mongo:
	@echo "Deploying mongo..."
	@$(KUBECTL) apply -f $(MONGO) -n $(NAMESPACE);

# Шаг ожидания пода перед выполнением команд
.PHONY: wait-for-pod
wait-for-pod: mongo
	@echo "Waiting for pod to be ready..."
	@-$(KUBECTL) wait --for=condition=ready pod -l app=$(POD_NAME) -n $(NAMESPACE) --timeout=120s

# Шаг для входа в контейнер и выполнения команд
.PHONY: mongo-init
mongo-init: wait-for-pod
	@echo "Executing commands inside the container..."
	POD_IP=$(shell kubectl get pod -l app=$(POD_NAME) -n $(NAMESPACE) -o jsonpath='{.items[0].status.podIP}'); \
	$(KUBECTL) exec -it $$(kubectl get pod -l app=$(POD_NAME) -o jsonpath='{.items[0].metadata.name}') -n $(NAMESPACE) -- mongosh --eval "rs.initiate( {\"_id\": \"rs0\", \"members\": [ {\"_id\": 0, \"host\": \"$(POD_IP):27017\"} ]} )"


.PHONY: clean-pvc
clean-pvc: clean-deploy
	@$(KUBECTL) delete pvc mongo-data-mongodb-0


.PHONY: clean-deploy
clean-deploy:
	@-$(KUBECTL) delete -f $(MONGO) -n $(NAMESPACE);
	@-for deploy in $(DEPLOYMENTS); do \
		$(KUBECTL) delete -f $$deploy -n $(NAMESPACE); \
	done

# Очистка ресурсов
.PHONY: clean
clean: clean-pvc
	@echo "Cleaning up..."
	@eval $(minikube -p minikube docker-env -u)


some:
	POD_IP=$(shell kubectl get pod -l app=$(POD_NAME) -n $(NAMESPACE) -o jsonpath='{.items[0].status.podIP}'); \
	echo $$POD_IP;