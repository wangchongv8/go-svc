.PHONY: fmt test run run-user-rpc run-product-rpc run-inventory-rpc run-order-rpc run-gateway-api gen db-migrate compose-up compose-down compose-logs compose-ps e2e-compose verify-observability-compose k8s-build k8s-push k8s-kind-load k8s-up k8s-ps k8s-logs k8s-port-forward e2e-k8s verify-observability-k8s k8s-down ci-check k8s-set-images k8s-rollout-status

fmt:
	go fmt ./...

test:
	go test ./...

# Run gateway-api (HTTP server on :8080)
run-gateway-api:
	go run ./apps/gateway-api/gateway.go -f apps/gateway-api/etc/gateway-api.yaml

# Run user-rpc (gRPC server on :9000)
run-user-rpc:
	go run ./apps/user-rpc/user.go -f apps/user-rpc/etc/user.yaml

# Run product-rpc (gRPC server on :9001)
run-product-rpc:
	go run ./apps/product-rpc/product.go -f apps/product-rpc/etc/product.yaml

# Run inventory-rpc (gRPC server on :9002)
run-inventory-rpc:
	go run ./apps/inventory-rpc/inventory.go -f apps/inventory-rpc/etc/inventory.yaml

# Run order-rpc (gRPC server on :9003)
run-order-rpc:
	go run ./apps/order-rpc/order.go -f apps/order-rpc/etc/order.yaml

# Legacy alias: runs gateway-api
run: run-gateway-api

# Regenerate go-zero code from .api and .proto files
gen:
	goctl rpc protoc apps/user-rpc/user.proto --go_out=apps/user-rpc --go-grpc_out=apps/user-rpc --zrpc_out=apps/user-rpc
	goctl rpc protoc apps/product-rpc/product.proto --go_out=apps/product-rpc --go-grpc_out=apps/product-rpc --zrpc_out=apps/product-rpc
	goctl rpc protoc apps/inventory-rpc/inventory.proto --go_out=apps/inventory-rpc --go-grpc_out=apps/inventory-rpc --zrpc_out=apps/inventory-rpc
	goctl rpc protoc apps/order-rpc/order.proto --go_out=apps/order-rpc --go-grpc_out=apps/order-rpc --zrpc_out=apps/order-rpc
	goctl api go -api apps/gateway-api/gateway.api -dir apps/gateway-api

# Run database migrations (requires PostgreSQL running locally)
db-migrate:
	psql -h localhost -U postgres -d go_svc -f deploy/sql/001_phase3_schema.sql

# Docker Compose commands
COMPOSE_FILE := deploy/docker-compose/docker-compose.yml

compose-up:
	docker compose -f $(COMPOSE_FILE) up --build -d

compose-down:
	docker compose -f $(COMPOSE_FILE) down -v

compose-logs:
	docker compose -f $(COMPOSE_FILE) logs -f

compose-ps:
	docker compose -f $(COMPOSE_FILE) ps

e2e-compose:
	./scripts/e2e-compose.sh

verify-observability-compose:
	./scripts/verify-observability-compose.sh

# Kubernetes commands
K8S_NAMESPACE ?= go-svc
KIND_CLUSTER ?= go-svc
IMAGE_REGISTRY ?= ghcr.io/wangchongv8
IMAGE_TAG ?= local

k8s-build:
	docker build --platform linux/amd64 --build-arg SERVICE_MAIN=apps/user-rpc/user.go --build-arg SERVICE_CONF_DIR=apps/user-rpc/etc -t $(IMAGE_REGISTRY)/go-svc-user-rpc:$(IMAGE_TAG) -f deploy/docker/service.Dockerfile .
	docker build --platform linux/amd64 --build-arg SERVICE_MAIN=apps/product-rpc/product.go --build-arg SERVICE_CONF_DIR=apps/product-rpc/etc -t $(IMAGE_REGISTRY)/go-svc-product-rpc:$(IMAGE_TAG) -f deploy/docker/service.Dockerfile .
	docker build --platform linux/amd64 --build-arg SERVICE_MAIN=apps/inventory-rpc/inventory.go --build-arg SERVICE_CONF_DIR=apps/inventory-rpc/etc -t $(IMAGE_REGISTRY)/go-svc-inventory-rpc:$(IMAGE_TAG) -f deploy/docker/service.Dockerfile .
	docker build --platform linux/amd64 --build-arg SERVICE_MAIN=apps/order-rpc/order.go --build-arg SERVICE_CONF_DIR=apps/order-rpc/etc -t $(IMAGE_REGISTRY)/go-svc-order-rpc:$(IMAGE_TAG) -f deploy/docker/service.Dockerfile .
	docker build --platform linux/amd64 --build-arg SERVICE_MAIN=apps/gateway-api/gateway.go --build-arg SERVICE_CONF_DIR=apps/gateway-api/etc -t $(IMAGE_REGISTRY)/go-svc-gateway-api:$(IMAGE_TAG) -f deploy/docker/service.Dockerfile .

k8s-push:
	@for svc in user-rpc product-rpc inventory-rpc order-rpc gateway-api; do \
		img=$(IMAGE_REGISTRY)/go-svc-$$svc:$(IMAGE_TAG); \
		echo "Pushing $$img"; \
		docker push $$img; \
	done

k8s-kind-load:
	@for svc in user-rpc product-rpc inventory-rpc order-rpc gateway-api; do \
		img=$(IMAGE_REGISTRY)/go-svc-$$svc:$(IMAGE_TAG); \
		echo "Loading $$img"; \
		kind load docker-image $$img --name $(KIND_CLUSTER); \
	done

k8s-up:
	kubectl apply -f deploy/k8s/namespace.yaml
	kubectl apply -f deploy/k8s/postgres.yaml
	kubectl wait --for=condition=ready pod -l app=postgres -n $(K8S_NAMESPACE) --timeout=60s
	kubectl apply -f deploy/k8s/db-migrate-job.yaml
	kubectl wait --for=condition=complete job/db-migrate -n $(K8S_NAMESPACE) --timeout=60s
	kubectl apply -f deploy/k8s/user-rpc.yaml
	kubectl apply -f deploy/k8s/product-rpc.yaml
	kubectl apply -f deploy/k8s/inventory-rpc.yaml
	kubectl apply -f deploy/k8s/order-rpc.yaml
	kubectl apply -f deploy/k8s/gateway-api.yaml
	kubectl apply -f deploy/k8s/ingress.yaml
	kubectl apply -f deploy/k8s/observability.yaml
	@echo "Waiting for pods..."
	@for app in user-rpc product-rpc inventory-rpc order-rpc gateway-api; do \
		kubectl wait --for=condition=ready pod -l app=$$app -n $(K8S_NAMESPACE) --timeout=60s; \
	done
	@echo "All pods ready."

k8s-ps:
	kubectl get pods,svc -n $(K8S_NAMESPACE)

k8s-logs:
	kubectl logs -l app=gateway-api -n $(K8S_NAMESPACE) -f

k8s-port-forward:
	kubectl port-forward svc/gateway-api 8080:8080 -n $(K8S_NAMESPACE)

e2e-k8s:
	./scripts/e2e-k8s.sh

verify-observability-k8s:
	./scripts/verify-observability-k8s.sh

k8s-down:
	kubectl delete namespace $(K8S_NAMESPACE)

# CI checks (same as GitHub Actions ci.yml)
ci-check:
	go fmt ./... && git diff --exit-code
	go vet ./...
	go test ./...
	for f in scripts/*.sh; do bash -n "$$f"; done
	docker compose -f deploy/docker-compose/docker-compose.yml config -q
	ruby -e 'require "yaml"; Dir.glob("deploy/k8s/*.yaml").each{|f| YAML.load_stream(File.read(f)); puts "#{f}: OK"}; Dir.glob("apps/*/etc/*.yaml").each{|f| YAML.load_stream(File.read(f)); puts "#{f}: OK"}'
	make gen && git diff --exit-code

# Update K8s Deployment image tags
k8s-set-images:
	kubectl set image deployment/user-rpc user-rpc=$(IMAGE_REGISTRY)/go-svc-user-rpc:$(IMAGE_TAG) -n $(K8S_NAMESPACE)
	kubectl set image deployment/product-rpc product-rpc=$(IMAGE_REGISTRY)/go-svc-product-rpc:$(IMAGE_TAG) -n $(K8S_NAMESPACE)
	kubectl set image deployment/inventory-rpc inventory-rpc=$(IMAGE_REGISTRY)/go-svc-inventory-rpc:$(IMAGE_TAG) -n $(K8S_NAMESPACE)
	kubectl set image deployment/order-rpc order-rpc=$(IMAGE_REGISTRY)/go-svc-order-rpc:$(IMAGE_TAG) -n $(K8S_NAMESPACE)
	kubectl set image deployment/gateway-api gateway-api=$(IMAGE_REGISTRY)/go-svc-gateway-api:$(IMAGE_TAG) -n $(K8S_NAMESPACE)

# Watch rollout status
k8s-rollout-status:
	@for app in user-rpc product-rpc inventory-rpc order-rpc gateway-api; do \
		echo "=== $$app ==="; \
		kubectl rollout status deployment/$$app -n $(K8S_NAMESPACE) --timeout=120s; \
	done
