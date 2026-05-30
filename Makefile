.PHONY: fmt test run run-user-rpc run-gateway-api gen

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

# Legacy alias: runs gateway-api
run: run-gateway-api

# Regenerate go-zero code from .api and .proto files
gen:
	goctl rpc protoc apps/user-rpc/user.proto --go_out=apps/user-rpc --go-grpc_out=apps/user-rpc --zrpc_out=apps/user-rpc
	goctl api go -api apps/gateway-api/gateway.api -dir apps/gateway-api
