# Phase 8 Review: Loki Log Search and Trace ID Correlation

Review date: 2026-06-07

## Findings

No code-level findings from static review.

## Resolved Since Previous Review

- README trace/log example now uses `POST /api/v1/register` and the `user-rpc` Loki query, matching the verification script.
- `scripts/verify-observability-k8s.sh` manages a gateway-api port-forward with `OBS_GATEWAY_PORT` and cleans it up through the existing trap.
- `scripts/verify-observability-k8s.sh` uses `POST /api/v1/register` to create a business log in `user-rpc`, captures `X-Trace-Id`, and queries Loki with `{namespace="go-svc", app="user-rpc"} | json | trace_id="..."`.
- `scripts/verify-observability-k8s.sh` includes a general LogQL check for `{namespace="go-svc"}`.
- Alloy includes `stage.cri {}` before `stage.json`, so CRI container log lines can be unwrapped before JSON field parsing.
- `scripts/verify-observability-compose.sh` uses a GET request while collecting response headers, not `curl -I`.
- K8s Loki ConfigMap key matches `-config.file=/etc/loki/local-config.yaml`.
- `make k8s-up` applies `deploy/k8s/alloy.yaml`.
- Alloy defines `__path__` for `/var/log/pods/.../*.log`.
- Alloy maps low-cardinality labels: `namespace`, `pod`, `container`, `app`.
- Business logs include `trace_id`/`span_id` in register, login, create product, create order, and stock deduction failure paths.
- `obs-k8s-port-forward` uses one shell with trap cleanup.
- Compose Loki limitation is documented in `deploy/observability/README.md`.

## Verification Run

Passed:

```bash
go test ./...
go vet ./...
make gen
docker compose -f deploy/docker-compose/docker-compose.yml config -q
bash -n scripts/verify-observability-compose.sh scripts/verify-observability-k8s.sh scripts/e2e-compose.sh scripts/e2e-k8s.sh
git diff --check
ruby YAML parse for deploy/k8s/*.yaml and apps/*/etc/*.yaml
```

Not completed on this machine:

```bash
kubectl apply --dry-run=client -f deploy/k8s/
make k8s-up
make e2e-k8s
make verify-observability-k8s
```

Reason: this machine has no current kubectl context (`current-context is not set`), so it cannot connect to a Kubernetes API Server for OpenAPI/schema validation or runtime verification.

Runtime Compose/K8s validation was not started during review, so there were no services or port-forwards to clean up.

## Summary

Static review passes. Before merging, run the runtime K8s checks on the remote minikube machine:

```bash
make k8s-up
make e2e-k8s
make verify-observability-k8s
```

If those pass, Phase 8 is ready to merge.
