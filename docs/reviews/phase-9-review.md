# Phase 9 Review: Ory Kratos Authentication

## Findings

No blocking findings in this pass.

The previously reported trace-log correlation issue is fixed: auth register, login, middleware identity mapping, and `/auth/me` user lookup now propagate trace metadata to `user-rpc` with `traceid.WithOutgoingMetadata(...)`.

The previously reported middleware test concern is also addressed enough for this phase: there is now a real production middleware test in `apps/gateway-api/internal/middleware/authmiddleware_test.go`. The helper-chain test in `apps/gateway-api/internal/authctx/authctx_middleware_test.go` still exists, but it is now explicitly scoped as an `authctx` helper test and points readers to the production middleware test.

## Confirmed Fixed

- `scripts/verify-observability-k8s.sh` calls `/api/v1/auth/register`, not removed `/api/v1/register`.
- README documents `/api/v1/auth/register`, `/api/v1/auth/login`, and `/api/v1/auth/me`.
- README no longer says `order-rpc` uses fixed `user_id = 1`.
- README no longer claims old `/api/v1/register` and `/api/v1/login` remain as legacy endpoints.
- Auth endpoints are declared in `gateway.api`.
- `gateway.go` no longer manually registers `/api/v1/auth/*`.
- `AuthMiddleware` is generated into routes via `@server(middleware: AuthMiddleware)` and `rest.WithMiddlewares`.
- `GET /api/v1/auth/me` is protected by `AuthMiddleware` and reads authenticated `user_id` from request context.
- `CreateOrderHandler` no longer manually performs authentication.
- `scripts/e2e-compose.sh` and `scripts/e2e-k8s.sh` pass Authorization to protected order reads.
- Kratos K8s resources are split into config, create-db Job, migrate Job, and deployment manifests.
- `make k8s-up` applies Kratos config, waits for create-db, waits for migrate, then applies the Kratos Deployment.
- `CREATE DATABASE kratos` is idempotent in both Compose and K8s.
- Auth-to-`user-rpc` calls propagate trace metadata:
  - `apps/gateway-api/internal/logic/auth/authregisterlogic.go`
  - `apps/gateway-api/internal/logic/auth/authloginlogic.go`
  - `apps/gateway-api/internal/authctx/authctx.go`
  - `apps/gateway-api/internal/logic/auth/authmelogic.go`
- Production `AuthMiddleware` tests cover:
  - missing token returns 401 and does not call `next`
  - invalid token returns 401 and does not call `next`
  - valid token injects `user_id` and calls `next`
  - identity mapping failure returns 401 and does not call `next`

## Verification Run

Passed:

```bash
go test ./...
go vet ./...
docker compose -f deploy/docker-compose/docker-compose.yml config -q
for f in scripts/*.sh; do bash -n "$f"; done
ruby -e 'require "yaml"; Dir.glob("deploy/k8s/*.yaml").each{|f| YAML.load_stream(File.read(f)); puts "#{f}: OK"}; Dir.glob("apps/*/etc/*.yaml").each{|f| YAML.load_stream(File.read(f)); puts "#{f}: OK"}'
git diff --check
```

Attempted but not completed:

```bash
kubectl apply --dry-run=client --validate=false -f deploy/k8s/
```

Result: the command reached the local kube context but failed with `localhost:8080: connect: connection refused`. This was treated as an environment limitation because the local kube-apiserver is not running, not as a manifest failure.

Not run:

```bash
make compose-up
make e2e-compose
make verify-observability-compose
make compose-down
make k8s-up
make e2e-k8s
make verify-observability-k8s
make k8s-down
```

No long-running services or port-forwards were started in this review pass. A common-port scan found no listeners on the usual project verification ports.

## Summary

This Phase 9 implementation is ready to approve from the code-review pass I can run locally. The remaining confidence gap is runtime validation in a real Docker/K8s environment, especially `make e2e-k8s` and `make verify-observability-k8s`, because this local environment currently has no reachable kube-apiserver.
