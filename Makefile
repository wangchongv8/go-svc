.PHONY: fmt test run run-user-rpc run-product-rpc run-inventory-rpc run-order-rpc run-gateway-api gen db-migrate

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
