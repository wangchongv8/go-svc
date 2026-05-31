# Phase 6 Review

## Final Re-review Findings

No blocking findings remain in this pass.

## Resolved Since Previous Review

- `Makefile:1,62-63,123-124` now exposes `verify-observability-compose` and `verify-observability-k8s`.
- `scripts/verify-observability-k8s.sh` was added.
- `deploy/docker-compose/docker-compose.yml:93-95` now publishes `gateway-api` metrics on host `6060`, matching the Compose docs/script.
- Host-local configs no longer enable a shared `DevServer.Port: 6060`; Compose/K8s configs keep `6060` inside isolated network namespaces.
- K8s Grafana datasource is provisioned via `grafana-datasource` ConfigMap and volume mount in `deploy/k8s/observability.yaml:70-133`.
- Prometheus/Grafana/Jaeger images are pinned in Compose and K8s.
- `scripts/verify-observability-compose.sh:39` now checks for a real Prometheus healthy response instead of an empty expected string.
- `scripts/verify-observability-k8s.sh:5-34` now uses a non-default Prometheus local port and verifies port-forward PIDs after startup.
- `scripts/verify-observability-compose.sh:65-73` now parses the registered user ID and uses it for the trace order request.
- `scripts/verify-observability-compose.sh:42-50` and `scripts/verify-observability-k8s.sh:37-45` now use Prometheus query API with per-instance `up{instance="svc:6060"}` checks instead of parsing `/api/v1/targets` with grep context.
- `scripts/verify-observability-k8s.sh:49-51` now normalizes the Jaeger count before numeric comparison.

## Verification

- `git diff --check` passed.
- `bash -n scripts/verify-observability-compose.sh scripts/verify-observability-k8s.sh scripts/e2e-compose.sh scripts/e2e-k8s.sh` passed.
- `docker compose -f deploy/docker-compose/docker-compose.yml config` passed.
- YAML parsing for `deploy/k8s/*.yaml` and `apps/*/etc/*.yaml` passed via Ruby YAML loader.
- `make test` passed.
- No Compose/K8s services were started during this review. Port check for `3000,4317,5432,6060,8080,9000-9003,9090,16686,19090` returned no listeners.
