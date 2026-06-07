# Phase 8: Loki Log Search and Trace ID Correlation

## Goal

Phase 8 adds a lightweight log search system and connects logs with traces:

- Use Loki as the log storage and LogQL query backend.
- Use Grafana Explore as the log search UI.
- Use Grafana Alloy to collect Kubernetes Pod logs and send them to Loki.
- Add trace id visibility to HTTP responses and business logs.
- Keep the learning environment disposable: logs stay inside containers or Kubernetes temporary storage, not on the host filesystem.

The desired learning flow is:

```text
send API request
  -> get X-Trace-Id from HTTP response
  -> open Jaeger and inspect the trace
  -> open Grafana Explore
  -> query Loki by trace_id and service labels
```

## Branch Requirement

Claude Code must implement this phase on a new branch, not directly on `main`.

Suggested branch:

```bash
git checkout main
git pull
git checkout -b feature/phase8-log-search-trace-id
```

If local `main` has uncommitted changes, Claude Code must stop and report them before creating the branch.

## Design Decisions

1. **Use Loki for logs.**
   Loki stores log streams and executes LogQL queries. Grafana is only the query and visualization UI.

2. **Use Alloy, not Promtail.**
   Promtail should not be introduced for new work. Alloy is the preferred collector for this project because it can collect logs and also fits the broader OpenTelemetry collector model.

3. **Do not use host-mounted log storage.**
   Compose and Kubernetes log storage should be disposable, like the current PostgreSQL learning setup.

4. **Do not use trace id as a Loki label.**
   `trace_id` has high cardinality. Keep it inside JSON log fields and filter it with LogQL pipeline stages.

5. **Do not replace Jaeger.**
   Phase 8 adds log search and trace/log correlation. Jaeger remains the trace UI for now.

## Architecture

Compose path:

```text
business service stdout
  -> Docker container logs
  -> Loki Docker logging driver or Alloy-compatible collection path
  -> Loki
  -> Grafana datasource
```

Kubernetes path:

```text
business service stdout
  -> node container log files
  -> Alloy DaemonSet
  -> Loki
  -> Grafana datasource
```

Request correlation path:

```text
gateway-api receives request
  -> OpenTelemetry creates trace/span
  -> gateway-api returns X-Trace-Id
  -> logx writes JSON logs with trace_id
  -> Loki query filters by trace_id
```

## Scope

Must implement:

- Loki in Docker Compose.
- Loki in Kubernetes observability manifests.
- Grafana datasource provisioning for Loki in Compose and Kubernetes.
- Alloy DaemonSet in Kubernetes to collect Pod logs.
- trace id extraction helper for Go contexts.
- HTTP middleware in `gateway-api` that returns `X-Trace-Id`.
- Business logs enriched with `trace_id` where context is available.
- README / observability docs / Makefile updates.
- Verification scripts for Compose and Kubernetes log search.

Should consider:

- Add `span_id` to logs when easily available.
- Add a small reusable package such as `pkg/observability/traceid`.
- Keep existing go-zero `logx.WithContext(ctx)` style.

Do not implement:

- ELK, OpenSearch, Tempo, Promtail, Helm, Argo CD, Flux, service mesh.
- Production-grade Loki retention, object storage, compactor, ruler, or multi-tenant auth.
- Host-mounted Loki data directory.
- New business features.

## Storage Policy

Logs must stay in the runtime environment:

- Docker Compose: no host directory bind mount for Loki data.
- Kubernetes: use `emptyDir` or container temporary storage for Loki data.
- `make compose-down` / `docker compose down -v` may delete logs.
- `make k8s-down` may delete logs by deleting the `go-svc` namespace.

This is intentional for the learning project. Durable logs can be a later phase.

## Compose Requirements

Update `deploy/docker-compose/docker-compose.yml`:

- Add `loki`.
- Add Loki datasource provisioning to Grafana.
- Keep Grafana on `3000`.
- Expose Loki to the Compose network on `3100`.
- Host port `3100:3100` is optional for learning/debugging.
- Do not mount a host path for Loki data.

Recommended Loki config location:

```text
deploy/observability/loki/loki.yml
```

Recommended Grafana datasource provisioning:

```text
deploy/observability/grafana/provisioning/datasources/loki.yml
```

If collecting Docker Compose container logs directly becomes too complex for this phase, it is acceptable to:

- Ensure Loki and Grafana datasource are up in Compose.
- Keep full automatic log collection for Kubernetes through Alloy.
- Document Compose log search as a known follow-up.

## Kubernetes Requirements

Update `deploy/k8s/observability.yaml` or split into clear files under `deploy/k8s/`:

- Add Loki ConfigMap, Deployment, and Service.
- Add Alloy ConfigMap, ServiceAccount, RBAC, and DaemonSet.
- Alloy should collect Pod logs from the node log paths and attach Kubernetes metadata.
- Grafana datasource provisioning should include Loki:

```text
http://loki:3100
```

Labels for Loki log streams should include low-cardinality Kubernetes metadata:

- `namespace`
- `app`
- `pod`
- `container`

Do not add `trace_id`, `user_id`, `order_id`, or request ids as Loki labels.

## Go Logging and Trace ID Requirements

Add a small shared helper, for example:

```text
pkg/observability/traceid
```

Responsibilities:

- Extract current trace id from `context.Context`.
- Extract current span id when available.
- Return empty strings when the context has no valid span.

`gateway-api` must add an HTTP middleware:

- Read the current trace id after go-zero/OpenTelemetry has attached span context.
- Set response header:

```text
X-Trace-Id: <trace-id>
```

Business logs should include trace fields when context is available:

```text
trace_id
span_id
service
```

Keep existing useful business fields:

```text
user_id
product_id
order_id
quantity
```

Sensitive data must not be logged:

- password
- token
- authorization header
- database password

## Example Queries

Find logs for gateway-api:

```logql
{namespace="go-svc", app="gateway-api"}
```

Find logs by trace id:

```logql
{namespace="go-svc"} | json | trace_id="abc123"
```

Find order-rpc errors:

```logql
{namespace="go-svc", app="order-rpc"} | json | level="error"
```

Find logs around order creation:

```logql
{namespace="go-svc", app="order-rpc"} |= "create order"
```

## Makefile Requirements

Add or update targets:

```makefile
obs-compose-urls
obs-k8s-port-forward
verify-observability-compose
verify-observability-k8s
```

`obs-compose-urls` should include:

```text
Grafana: http://localhost:3000
Jaeger:  http://localhost:16686
Loki:    http://localhost:3100
```

`obs-k8s-port-forward` should include or print:

```bash
kubectl port-forward -n go-svc svc/grafana 3000:3000
kubectl port-forward -n go-svc svc/jaeger 16686:16686
kubectl port-forward -n go-svc svc/loki 3100:3100
```

Verification scripts must clean up any port-forward process they start.

## Verification

Base checks:

```bash
make fmt
make test
git diff --check
make gen
docker compose -f deploy/docker-compose/docker-compose.yml config
kubectl apply --dry-run=client -f deploy/k8s/
```

Compose checks:

```bash
make compose-up
make e2e-compose
make verify-observability-compose
make compose-down
```

Kubernetes checks:

```bash
make k8s-up
make e2e-k8s
make verify-observability-k8s
make k8s-down
```

Manual trace/log check:

```bash
curl -i http://localhost:8080/healthz
```

Expected:

- HTTP response includes `X-Trace-Id` when tracing is active.
- Jaeger can find a trace for the request.
- Grafana Explore can query Loki logs for the same trace id after a business request.

If the current local environment cannot run Docker or Kubernetes, Claude Code must report which checks were skipped and why.

After any runtime validation, Claude Code must stop Compose/Kubernetes/port-forward processes it started and check relevant ports:

```text
8080, 9000, 9001, 9002, 9003, 5432, 9090, 3000, 16686, 3100
```

Do not kill unrelated user processes. If a port is occupied by an unrelated process, report it.

## Claude Code Prompt

```text
请在新分支实现 docs/phase-8-log-search-trace-id.md。

要求：
- 先确认当前分支和工作区状态。如果不在新分支，请基于 main 创建 feature/phase8-log-search-trace-id。
- 严格遵循 Phase 8 文档。
- 引入 Loki + Grafana datasource + Kubernetes Alloy 日志采集。
- 不引入 Promtail、ELK/OpenSearch、Tempo、Helm、Argo CD、Flux 或 service mesh。
- Loki 日志数据不挂载到宿主机；Compose/K8s 都使用可清理的临时存储。
- 保持 PostgreSQL 当前“不本地持久化，通过 migration 重建”的学习环境策略。
- gateway-api 返回 X-Trace-Id。
- 业务关键日志补充 trace_id，但不要把 trace_id 作为 Prometheus label 或 Loki label。
- 不记录密码、token、Authorization header、数据库密码等敏感信息。
- 更新 README、deploy/observability/README.md、Makefile 和验证脚本。
- 保持 Phase 4/5/6/7 原有能力不回退。
- 实现后运行 make fmt、make test、git diff --check、make gen。
- 如果 Docker 可用，运行 docker compose -f deploy/docker-compose/docker-compose.yml config；如可启动 Compose，运行 make compose-up、make e2e-compose、make verify-observability-compose、make compose-down。
- 如果 kubectl 可用，运行 kubectl apply --dry-run=client -f deploy/k8s/；如 K8s 可用，运行 make k8s-up、make e2e-k8s、make verify-observability-k8s、make k8s-down。
- 验证结束后必须清理 Compose/K8s/port-forward 进程，并复查 8080、9000、9001、9002、9003、5432、9090、3000、16686、3100 无异常监听。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```

## Review Focus

- 是否在新分支实现，没有直接污染 main。
- Loki 是否可被 Grafana 查询。
- Kubernetes Alloy 是否能采集 Pod stdout 日志并带上 namespace/app/pod/container 标签。
- Loki 是否没有使用宿主机目录持久化数据。
- Grafana datasource 是否同时包含 Prometheus 和 Loki。
- `X-Trace-Id` 是否在 HTTP 响应中可见。
- 日志是否包含 `trace_id`，且没有泄露敏感信息。
- `trace_id` 是否没有成为 Prometheus label 或 Loki label。
- 验证脚本是否有 trap 清理 port-forward。
- Compose/K8s/e2e/CI 原有命令是否没有回退。
