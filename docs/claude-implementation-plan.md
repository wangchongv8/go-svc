# Claude Code 实施计划

本文档是交给 Claude Code 的执行入口。每个阶段都应小步实现、可运行、可验证。

## 全局要求

Claude Code 执行任何阶段时都必须遵守：

- 严格参考 `docs/architecture.md` 和 `docs/requirements.md`。
- 不实现当前阶段之外的大块功能。
- 新增命令必须写入 README 或对应文档。
- 生成代码后要运行格式化和可用的验证命令。
- 如果命令因为本地环境缺失无法运行，需要在结果中说明。
- 不引入未在文档中确认的新框架或中间件。

## Phase 0: 文档和协作脚手架

状态：已由 Codex 初始化。

目标：

- 建立 README。
- 建立需求、架构、实施计划和 review 清单。
- 为后续需求讨论预留位置。

验收：

- `docs/workflow.md` 存在。
- `docs/requirements.md` 存在。
- `docs/architecture.md` 存在。
- `docs/claude-implementation-plan.md` 存在。
- `docs/review-checklist.md` 存在。

## Phase 1: 最小 Go 工程

状态：已完成。

目标：

- 初始化 Go module。
- 创建一个最小 HTTP 服务，作为后续 `gateway-api` 的前身。
- 增加 Makefile。
- 增加基础健康检查接口。
- 增加 README 启动说明。
- 暂不引入 go-zero、gRPC、PostgreSQL、Docker 或 Kubernetes。

建议目录：

```text
.
├── cmd/
│   └── gateway/
│       └── main.go
├── internal/
│   └── gateway/
│       ├── handler.go
│       └── handler_test.go
├── docs/
├── Makefile
├── go.mod
└── README.md
```

功能要求：

- `GET /healthz` 返回 HTTP 200。
- 响应 JSON 至少包含 `status` 字段，例如 `{"status":"ok"}`。
- HTTP 监听端口默认 `8080`。
- 允许通过环境变量 `PORT` 覆盖端口。
- 代码结构要为后续迁移到 `apps/gateway-api` 或 go-zero 生成结构留下空间，但不要提前实现 go-zero。

Makefile 至少包含：

- `make fmt`
- `make test`
- `make run`

建议验收命令：

```bash
make fmt
go test ./...
make run
curl http://localhost:8080/healthz
```

Claude Code 执行提示词：

```text
请在当前仓库实现 docs/claude-implementation-plan.md 中的 Phase 1。

要求：
- 严格遵循 docs/architecture.md 和 docs/requirements.md。
- 本阶段只实现最小 Go HTTP 工程，不要引入 go-zero、gRPC、PostgreSQL、Docker、Kubernetes。
- 初始化 Go module，模块名可使用 go-svc。
- 创建 GET /healthz，返回 HTTP 200 和 JSON: {"status":"ok"}。
- 默认监听 8080，并支持 PORT 环境变量覆盖。
- 增加 Makefile，至少包含 fmt、test、run。
- 增加或更新 README 的本地运行说明。
- 增加最小测试覆盖 healthz handler。
- 实现后运行 make fmt、make test；如果运行 make run，需要说明如何验证 curl。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```

## Phase 2: go-zero API + RPC 基础

状态：已完成。

目标：

- 引入 go-zero API。
- 创建 `gateway-api`。
- 创建至少一个纯 gRPC 服务，例如 `user-rpc`。
- API 调用 RPC 返回结果。
- 暂不接 PostgreSQL。

验收：

- `make fmt`
- `make test`
- `make gen`
- 启动 `user-rpc` 和 `gateway-api` 后，验证注册、登录、查询用户和错误映射。

## Phase 3: 多服务业务闭环

状态：已完成。

目标：

- 完成核心业务服务拆分。
- 实现一个端到端业务流程，例如创建订单。
- 明确服务间调用链。
- 接入 PostgreSQL。

验收：

- `product-rpc`、`inventory-rpc`、`order-rpc` 已实现。
- PostgreSQL schema 已新增到 `deploy/sql/001_phase3_schema.sql`。
- 库存扣减使用条件更新防超卖。
- 创建订单链路为 `gateway-api -> order-rpc -> product-rpc/inventory-rpc -> PostgreSQL`。
- 单元测试使用 fake repository / fake RPC client。
- PostgreSQL 集成验证留到 Phase 4 Docker Compose。

## Phase 4: Docker Compose 本地集群

状态：已完成。

目标：

- 为服务增加 Dockerfile。
- 增加 docker-compose。
- 启动 PostgreSQL 和业务服务。
- 在 Docker Compose 环境中执行数据库迁移。
- 用脚本完成端到端集成验证。

本阶段暂不实现：

- Redis。
- etcd 服务发现。
- Kubernetes。
- payment-rpc。
- 支付、订单取消、库存回滚。
- Prometheus、Grafana、Jaeger。

说明：

- Phase 4 的核心目标是让本地一条命令拉起可用集群，并把 Phase 3 延后的 PostgreSQL 集成验证补上。
- etcd 和 Redis 可以在后续阶段按学习目标加入；本阶段不要为了“看起来完整”提前引入。

### Phase 4 目录建议

```text
deploy/
├── docker-compose/
│   ├── docker-compose.yml
│   └── README.md
├── docker/
│   └── service.Dockerfile
└── sql/
    └── 001_phase3_schema.sql

scripts/
├── compose-up.sh
├── compose-down.sh
└── e2e-phase4.sh
```

也可以不加 shell 脚本，改用 Makefile 命令直接封装；但 README 必须给出清晰命令。

### Phase 4 Dockerfile 要求

可以使用一个通用多阶段 Dockerfile，例如：

```text
deploy/docker/service.Dockerfile
```

要求：

- 使用 multi-stage build。
- builder 阶段编译目标服务。
- runtime 阶段只保留二进制、配置文件和必要运行时文件。
- 支持通过 build args 指定服务入口，例如：
  - `SERVICE_PATH=apps/gateway-api`
  - `SERVICE_BIN=gateway-api`
  - `SERVICE_MAIN=apps/gateway-api/gateway.go`
- 不把 Go build cache、调试二进制、`.git` 放进镜像。

也可以每个服务一个 Dockerfile，但要避免大量复制粘贴。

### Phase 4 Compose 服务

Compose 至少包含：

| 服务 | 端口 | 说明 |
| --- | --- | --- |
| postgres | 5432 | PostgreSQL 数据库 |
| db-migrate | 无 | 执行 `deploy/sql/001_phase3_schema.sql` 后退出 |
| user-rpc | 9000 | 用户 RPC |
| product-rpc | 9001 | 商品 RPC |
| inventory-rpc | 9002 | 库存 RPC |
| order-rpc | 9003 | 订单 RPC |
| gateway-api | 8080 | HTTP API |

要求：

- 本地 Compose 环境每次清空数据，通过 `db-migrate` 回放 SQL 重建表结构。不使用 volume 持久化 PostgreSQL 数据。
- `postgres` 必须有 healthcheck。
- `db-migrate` 依赖 `postgres` healthy。
- `product-rpc`、`inventory-rpc`、`order-rpc` 必须依赖 `db-migrate` 成功完成。
- `order-rpc` 必须使用 Compose 服务名访问 `product-rpc` 和 `inventory-rpc`，例如 `product-rpc:9001`。
- `gateway-api` 必须使用 Compose 服务名访问各 RPC 服务，例如 `user-rpc:9000`。
- 暴露给宿主机的端口至少包括 `8080`，RPC 端口可以暴露也可以只在 Compose 网络内使用；为了学习调试，可以保留 9000-9003 映射。

### Phase 4 配置要求

当前服务配置文件可以新增 Compose 专用配置，例如：

```text
apps/gateway-api/etc/gateway-api.compose.yaml
apps/product-rpc/etc/product.compose.yaml
apps/inventory-rpc/etc/inventory.compose.yaml
apps/order-rpc/etc/order.compose.yaml
apps/user-rpc/etc/user.compose.yaml
```

要求：

- 本地配置继续保留 `127.0.0.1`。
- Compose 配置使用服务名。
- PostgreSQL DSN 使用 Compose 网络地址：

```text
postgres://postgres:postgres@postgres:5432/go_svc?sslmode=disable
```

### Phase 4 Makefile

至少增加：

- `make compose-up`
- `make compose-down`
- `make compose-logs`
- `make compose-ps`
- `make e2e-compose`

建议命令：

```bash
make compose-up
make e2e-compose
make compose-down
```

`compose-down` 统一使用 `docker compose down -v` 清理容器和数据。

### Phase 4 集成验证

新增脚本或 Makefile 目标执行以下流程：

```bash
curl -f http://localhost:8080/healthz

curl -f -X POST http://localhost:8080/api/v1/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"123456"}'

curl -f -X POST http://localhost:8080/api/v1/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Keyboard","description":"Mechanical keyboard","price_cents":19900}'

curl -f -X PUT http://localhost:8080/api/v1/inventories/1 \
  -H 'Content-Type: application/json' \
  -d '{"stock":10}'

curl -f -X POST http://localhost:8080/api/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"product_id":1,"quantity":2}'

curl -f http://localhost:8080/api/v1/orders/1
curl -f http://localhost:8080/api/v1/inventories/1
curl -f http://localhost:8080/api/v1/users/1/orders
```

验证要求：

- 订单创建成功。
- 订单 `total_price_cents` 为 `39800`。
- 库存从 `10` 扣减到 `8`。
- 用户订单列表能查到刚创建的订单。
- 库存不足时返回非 500 错误。

脚本应尽量使用 `set -euo pipefail`，失败时退出非 0。

### Phase 4 验收命令

基础验证：

```bash
make fmt
make test
git diff --check
make gen
```

Docker Compose 验证：

```bash
make compose-up
make compose-ps
make e2e-compose
make compose-down
```

如果本地 Docker 不可用，Claude Code 需要说明未运行的命令、原因和手动验证步骤。

### Phase 4 Claude Code 执行提示词

```text
请在当前仓库实现 docs/claude-implementation-plan.md 中的 Phase 4。

要求：
- 严格遵循 Phase 4 说明。
- 本阶段只实现 Dockerfile、Docker Compose、本地集群启动、数据库迁移、端到端集成验证。
- 不实现 Kubernetes、Redis、etcd、payment-rpc、支付、订单取消、库存回滚、Prometheus、Grafana、Jaeger。
- 保留本地 127.0.0.1 配置，同时新增 Compose 专用配置，Compose 内服务互调用服务名。
- PostgreSQL 使用 postgres:5432，默认账号密码 postgres/postgres，数据库 go_svc。
- 增加 db-migrate 服务或等价机制，确保业务服务启动前 schema 已创建。
- 增加 Makefile 目标：compose-up、compose-down、compose-logs、compose-ps、e2e-compose。
- 增加端到端验证脚本或 Makefile 目标，覆盖注册用户、创建商品、设置库存、创建订单、查询订单、查询库存、查询用户订单列表。
- 不提交调试二进制、构建产物、数据库数据目录、缓存目录。
- 实现后运行 make fmt、make test、git diff --check、make gen。
- 如果 Docker 可用，运行 make compose-up、make e2e-compose、make compose-down。
- 验证结束后必须停止 Compose 服务，并复查 8080、9000、9001、9002、9003、5432 无异常监听。若 5432 是用户已有本地 PostgreSQL，请说明不要误杀。
- 如果 Docker 不可用，说明未运行的命令、原因和手动验证步骤。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```

### Phase 4 Review 重点

- Compose 是否能一键启动。
- 业务服务是否使用 Compose 服务名互相访问。
- 数据库迁移是否在业务服务启动前完成。
- PostgreSQL healthcheck 和服务依赖是否合理。
- `e2e-compose` 是否验证真实跨服务 + PostgreSQL 链路。
- Docker 镜像是否避免携带源码外的构建垃圾。
- README 是否包含启动、验证、停止、清理命令。
- 验证后是否停止服务并复查端口。

## Phase 5: Kubernetes 部署

状态：已完成。

目标：

- 将 Phase 4 的 Docker Compose 本地集群迁移到 Kubernetes manifests。
- 使用 Kubernetes Service + DNS 做服务发现，不让业务服务直接依赖 etcd。
- 覆盖 Namespace、Deployment、Service、ConfigMap、Secret、Job、Ingress。
- 支持两种部署路径：本地 kind 验证，以及通过 GHCR 将镜像发布到另一台电脑上的 Kubernetes 环境。
- 如果本机没有 kind/Kubernetes，也要能做 YAML 静态校验。
- 保持本地学习环境可重复：PostgreSQL 不做持久化，每次重建后通过 migration Job 初始化 schema。
- 复用 Phase 4 的业务端到端验证链路。

不做：

- 不引入 etcd、Redis、消息队列。
- 不引入 Prometheus、Grafana、Jaeger。
- 不实现 payment-rpc、支付、取消订单、库存回滚等业务新功能。
- 不引入 Helm；本阶段使用普通 YAML、少量脚本和 Makefile 即可。
- 不追求生产级安全、资源配额和高可用 PostgreSQL。

### Phase 5 推荐目录

```text
deploy/
├── k8s/
│   ├── README.md
│   ├── namespace.yaml
│   ├── postgres.yaml
│   ├── db-migrate-job.yaml
│   ├── gateway-api.yaml
│   ├── user-rpc.yaml
│   ├── product-rpc.yaml
│   ├── inventory-rpc.yaml
│   ├── order-rpc.yaml
│   └── ingress.yaml
scripts/
├── e2e-compose.sh
└── e2e-k8s.sh
```

可以拆得更细，但不要引入过度复杂的 kustomize/Helm 结构。

### Phase 5 Kubernetes 资源要求

Namespace：

- 使用 `go-svc` namespace。
- 所有 Phase 5 资源都放在该 namespace。

PostgreSQL：

- 使用 `postgres:16-alpine`。
- 使用 Secret 保存 `POSTGRES_USER`、`POSTGRES_PASSWORD`、`POSTGRES_DB`。
- 使用 `emptyDir` 或容器临时存储即可，不做 PVC。
- 提供 `ClusterIP` Service，服务名为 `postgres`，端口 `5432`。
- 配置 readiness/liveness probe，至少使用 `pg_isready`。

数据库迁移：

- 使用 Kubernetes Job 执行 `deploy/sql/001_phase3_schema.sql`。
- Job 需要等待 PostgreSQL ready 后再执行 `psql`。
- SQL 可以通过 ConfigMap 挂载。若复制 SQL 到 ConfigMap，需在 README 中说明它来自 `deploy/sql/001_phase3_schema.sql`，避免后续维护时忘记同步。
- Job 成功后业务服务应能正常处理请求。

业务服务：

- 为以下服务各创建 Deployment + Service：
  - `gateway-api`，HTTP `8080`
  - `user-rpc`，gRPC `9000`
  - `product-rpc`，gRPC `9001`
  - `inventory-rpc`，gRPC `9002`
  - `order-rpc`，gRPC `9003`
- 镜像命名建议：
  - `ghcr.io/wangchongv8/go-svc-gateway-api:phase5`
  - `ghcr.io/wangchongv8/go-svc-user-rpc:phase5`
  - `ghcr.io/wangchongv8/go-svc-product-rpc:phase5`
  - `ghcr.io/wangchongv8/go-svc-inventory-rpc:phase5`
  - `ghcr.io/wangchongv8/go-svc-order-rpc:phase5`
- Deployment 默认 `replicas: 1`。
- `imagePullPolicy: IfNotPresent`，方便 kind 加载本地镜像。
- 资源 requests/limits 可以设置较小默认值，避免本地集群压力过大。
- gateway-api 使用 HTTP readiness/liveness probe：`GET /healthz`。
- RPC 服务可以使用 `tcpSocket` readiness/liveness probe。
- Service 使用 `ClusterIP`，只暴露集群内部端口；外部验证通过 port-forward 或 Ingress。

镜像发布：

- 默认镜像仓库使用 GHCR：`ghcr.io/wangchongv8`。
- 默认镜像 tag 使用 `phase5`。
- Phase 5 先假设 GHCR Package 可以设为 public，目标机器不需要 `imagePullSecret` 即可拉取镜像。
- 如果镜像设置为 private，`deploy/k8s/README.md` 必须补充 `imagePullSecret` 创建方式，并在 Deployment 中说明如何启用。
- 不要求在另一台电脑上自建 registry。
- 不要求实现 `docker save/load` tar 包分发；可以在 README 中作为轻量备选方案说明。

配置：

- 新增 K8s 专用配置，推荐命名：

```text
apps/gateway-api/etc/gateway-api.k8s.yaml
apps/product-rpc/etc/product.k8s.yaml
apps/inventory-rpc/etc/inventory.k8s.yaml
apps/order-rpc/etc/order.k8s.yaml
apps/user-rpc/etc/user.k8s.yaml
```

- K8s 配置必须使用 Kubernetes Service 名称：

```text
user-rpc:9000
product-rpc:9001
inventory-rpc:9002
order-rpc:9003
postgres:5432
```

- 优先通过 ConfigMap 挂载配置文件，再让容器命令使用 `-f /app/etc/<service>.k8s.yaml`。
- PostgreSQL 账号密码应来自 Secret。若当前 go-zero 配置暂不支持从 Secret 拼接 DSN，可以先采用学习环境默认 DSN，但需要在 README 中说明这是 Phase 5 的简化点，并在 review 中作为后续改进项关注。

Ingress：

- 增加 `ingress.yaml`，路由到 `gateway-api` Service。
- 建议 host 使用 `go-svc.local`。
- 注明需要本地集群安装 ingress-nginx 或等价 Ingress Controller。
- Phase 5 主要验收可以通过 `kubectl port-forward svc/gateway-api 8080:8080` 完成，不强制要求本机已经安装 Ingress Controller。

### Phase 5 Makefile

至少增加：

- `make k8s-build`
- `make k8s-push`
- `make k8s-kind-load`
- `make k8s-up`
- `make k8s-ps`
- `make k8s-logs`
- `make k8s-port-forward`
- `make e2e-k8s`
- `make k8s-down`

建议变量：

```makefile
K8S_NAMESPACE ?= go-svc
KIND_CLUSTER ?= go-svc
IMAGE_REGISTRY ?= ghcr.io/wangchongv8
IMAGE_TAG ?= phase5
```

远端 Kubernetes 部署建议命令：

开发机：

```bash
make k8s-build
make k8s-push
```

目标机器：

```bash
make k8s-up
make k8s-ps
make e2e-k8s
make k8s-down
```

本地 kind 验证建议命令：

```bash
make k8s-build
make k8s-kind-load
make k8s-up
make k8s-ps
make e2e-k8s
make k8s-down
```

说明：

- `k8s-build` 使用 `deploy/docker/service.Dockerfile` 构建 5 个业务镜像，镜像名使用 `$(IMAGE_REGISTRY)/go-svc-<service>:$(IMAGE_TAG)`。
- `k8s-push` 将 5 个业务镜像推送到 `$(IMAGE_REGISTRY)`，默认是 `ghcr.io/wangchongv8`。如果未登录 GHCR，应给出清晰错误提示。
- `k8s-kind-load` 使用 `kind load docker-image` 把镜像加载进 kind 集群；如果 kind 不存在，应给出清晰错误。该命令是本地 kind 模式专用，不是远端部署必需步骤。
- `k8s-up` 应创建 namespace、应用 Secret/ConfigMap/PostgreSQL、等待 PostgreSQL ready、执行 db-migrate Job、应用业务服务和 Ingress。
- `k8s-ps` 输出 pods、services、jobs。
- `k8s-logs` 至少能查看 namespace 下最近日志，或提示用户指定服务。
- `k8s-port-forward` 将 `gateway-api` Service 转发到本地 `8080`。
- `e2e-k8s` 可复用 Phase 4 的 e2e 逻辑，但必须支持 `BASE_URL` 或自行启动临时 port-forward。
- `k8s-down` 删除 `go-svc` namespace，清理所有 K8s 资源。

### Phase 5 验证脚本

推荐调整 `scripts/e2e-compose.sh`：

- 支持 `BASE_URL` 环境变量，默认仍为 `http://localhost:8080`。
- 保持 Compose 验证不回退。

新增 `scripts/e2e-k8s.sh`：

- 检查 `kubectl` 是否可用。
- 检查 `go-svc` namespace 下 gateway-api 是否 ready。
- 启动临时 `kubectl port-forward -n go-svc svc/gateway-api 8080:8080`。
- 复用同一套 API 断言：健康检查、注册、登录、创建商品、设置库存、创建订单、查询订单、查询库存、查询用户订单列表、库存不足返回非 500。
- 脚本退出时必须清理 port-forward 进程。

### Phase 5 验收命令

基础验证：

```bash
make fmt
make test
git diff --check
make gen
docker compose -f deploy/docker-compose/docker-compose.yml config
kubectl apply --dry-run=client -f deploy/k8s/
```

kind 验证：

```bash
make k8s-build
make k8s-kind-load
make k8s-up
make k8s-ps
make e2e-k8s
make k8s-down
```

远端 Kubernetes 验证：

开发机：

```bash
docker login ghcr.io
make k8s-build
make k8s-push
```

目标机器：

```bash
git pull
kubectl apply --dry-run=client -f deploy/k8s/
make k8s-up
make k8s-ps
make e2e-k8s
make k8s-down
```

如果镜像是 private，目标机器还需要在 `go-svc` namespace 创建 GHCR `imagePullSecret`。

如果本地没有 Kubernetes 集群或 kind：

- 必须说明未运行的命令、原因和手动验证步骤。
- 至少运行 `kubectl apply --dry-run=client -f deploy/k8s/`，如果 kubectl 也不可用则说明原因。

### Phase 5 Claude Code 执行提示词

```text
请在当前仓库实现 docs/claude-implementation-plan.md 中的 Phase 5。

要求：
- 严格遵循 Phase 5 说明。
- 本阶段只实现 Kubernetes manifests、本地 kind 验证、GHCR 镜像发布、远端 Kubernetes 部署说明和相关脚本/命令。
- 不实现 Phase 6 可观测性。
- 不引入 etcd、Redis、消息队列、Helm、payment-rpc、支付、订单取消、库存回滚。
- Kubernetes 服务发现使用 Service + DNS，不让业务服务直接依赖 etcd。
- PostgreSQL 在本地 K8s 学习环境不持久化数据，通过 db-migrate Job 回放 deploy/sql/001_phase3_schema.sql。
- 新增或更新 K8s 专用配置，服务间地址使用 Kubernetes Service 名称。
- 默认镜像仓库使用 ghcr.io/wangchongv8，默认 tag 使用 phase5。
- 新增 deploy/k8s/README.md，说明 GHCR 登录、镜像构建、镜像推送、远端机器部署、kind 环境、镜像加载、启动、验证、清理、Ingress 使用条件。
- 新增 Makefile 目标：k8s-build、k8s-push、k8s-kind-load、k8s-up、k8s-ps、k8s-logs、k8s-port-forward、e2e-k8s、k8s-down。
- e2e-k8s 必须验证真实 Kubernetes 链路，结束后清理 port-forward 进程。
- 保持 Phase 4 Compose 能力不回退。
- 实现后运行 make fmt、make test、git diff --check、make gen。
- 如果 kubectl 可用，运行 kubectl apply --dry-run=client -f deploy/k8s/。
- 如果 Docker/GHCR 登录可用，运行 make k8s-build、make k8s-push。
- 如果 kind/Kubernetes 可用，运行 make k8s-build、make k8s-kind-load、make k8s-up、make k8s-ps、make e2e-k8s、make k8s-down。
- 验证结束后必须清理 Kubernetes 资源和 port-forward 进程，并复查 8080 无异常监听。不要误杀用户已有进程。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```

### Phase 5 Review 重点

- K8s manifests 是否能 dry-run 通过。
- Service 名称、端口、容器端口是否和应用配置一致。
- gateway-api、order-rpc 的服务间调用是否使用 K8s Service DNS。
- PostgreSQL Secret、ConfigMap、Job 是否职责清晰。
- db-migrate Job 是否能在业务验证前完成。
- readiness/liveness probe 是否合理。
- Ingress 是否清楚标注依赖 Ingress Controller。
- GHCR 镜像名是否统一使用 `ghcr.io/wangchongv8/go-svc-<service>:phase5` 或 Makefile 变量生成的等价结果。
- 远端 Kubernetes 部署说明是否清楚区分开发机 push 和目标机器 deploy。
- private GHCR 镜像的 `imagePullSecret` 说明是否完整。
- Makefile 命令是否可重复执行并能清理资源。
- `e2e-k8s` 是否验证真实 Kubernetes 链路，并清理 port-forward。
- Phase 4 Docker Compose 验证是否没有被破坏。
- 验证后是否清理 K8s 资源并复查本地端口。

## Phase 6: 可观测性

状态：已完成。

目标：

- 为 5 个业务服务和 gateway-api 增加最小可观测性闭环。
- 使用 go-zero 原生 `Log`、`DevServer`、`Telemetry` 配置，避免手写一套观测框架。
- 暴露 Prometheus `/metrics`。
- 接入 Jaeger/OpenTelemetry trace。
- 在 Docker Compose 和 Kubernetes 两种部署形态中都能验证指标和链路追踪。
- 提供 Prometheus、Grafana、Jaeger 的本地访问说明。

不做：

- 不新增业务功能。
- 不引入 ELK/Loki/Tempo/OpenSearch 等日志平台。
- 不做生产级告警规则、SLO、Grafana 完整 dashboard。
- 不改造为 service mesh。
- 不引入 Helm。
- 不要求目标机器必须长期运行观测组件；学习环境可随 `compose-down` / `k8s-down` 一起清理。

### Phase 6 设计原则

- 优先使用 go-zero 已有能力：
  - `Log`：结构化 JSON 日志。
  - `DevServer`：暴露 `/metrics`、`/healthz`、pprof。
  - `Telemetry`：OpenTelemetry trace exporter。
- 观测组件以学习可见性为主，资源占用要小。
- Compose 作为最容易完整验证的环境。
- K8s 作为部署形态验证：能部署 Prometheus/Jaeger，能 port-forward 查看。
- 日志本阶段只要求输出结构化 JSON 到 stdout，由 Docker/K8s 收集，不上日志平台。

### Phase 6 推荐目录

```text
deploy/
├── observability/
│   ├── README.md
│   ├── prometheus/
│   │   └── prometheus.yml
│   └── grafana/
│       ├── provisioning/
│       │   └── datasources/
│       │       └── prometheus.yml
│       └── dashboards/
│           └── go-svc-overview.json   # 可选，简单即可
├── k8s/
│   ├── observability.yaml
│   └── ...
scripts/
├── verify-observability-compose.sh
└── verify-observability-k8s.sh
```

如果 Grafana dashboard 工作量过大，可以只提供 datasource provisioning 和 README 查询说明，不强制做复杂 dashboard。

### Phase 6 服务配置要求

所有服务配置文件都要补充可观测性配置：

```yaml
Log:
  Mode: console
  Encoding: json
  Level: info
  Stat: true

DevServer:
  Enabled: true
  Host: 0.0.0.0
  Port: 6060
  MetricsPath: /metrics
  HealthPath: /healthz
  EnableMetrics: true
  EnablePprof: true

Telemetry:
  Name: <service-name>
  Endpoint: <trace-endpoint>
  Sampler: 1.0
  Batcher: otlpgrpc
```

说明：

- 本地 `*.yaml` 可以不配置 `Telemetry.Endpoint`，避免未启动 Jaeger 时本地服务启动失败或刷错误。
- Compose 配置使用 `Endpoint: jaeger:4317`。
- K8s 配置使用 `Endpoint: jaeger:4317` 或同 namespace 下的 Jaeger Service 名。
- 每个 Pod 内部都可以使用 `DevServer.Port: 6060`，K8s Service 通过 named port 暴露 metrics。
- Docker Compose 内部 Prometheus 通过服务名抓取 `service:6060`，不要求把所有 6060 映射到宿主机。

需要更新的配置文件：

```text
apps/gateway-api/etc/gateway-api.yaml
apps/gateway-api/etc/gateway-api.compose.yaml
apps/gateway-api/etc/gateway-api.k8s.yaml
apps/user-rpc/etc/user.yaml
apps/user-rpc/etc/user.compose.yaml
apps/user-rpc/etc/user.k8s.yaml
apps/product-rpc/etc/product.yaml
apps/product-rpc/etc/product.compose.yaml
apps/product-rpc/etc/product.k8s.yaml
apps/inventory-rpc/etc/inventory.yaml
apps/inventory-rpc/etc/inventory.compose.yaml
apps/inventory-rpc/etc/inventory.k8s.yaml
apps/order-rpc/etc/order.yaml
apps/order-rpc/etc/order.compose.yaml
apps/order-rpc/etc/order.k8s.yaml
```

### Phase 6 Docker Compose 要求

更新 `deploy/docker-compose/docker-compose.yml`：

- 增加 `prometheus` 服务。
- 增加 `grafana` 服务。
- 增加 `jaeger` 服务。
- Prometheus 使用 `deploy/observability/prometheus/prometheus.yml`。
- Grafana 默认暴露 `3000:3000`。
- Prometheus 默认暴露 `9090:9090`。
- Jaeger UI 默认暴露 `16686:16686`。
- Jaeger OTLP gRPC 暴露给 Compose 网络内服务使用 `4317`。

Prometheus scrape targets 至少包含：

```yaml
- targets:
    - gateway-api:6060
    - user-rpc:6060
    - product-rpc:6060
    - inventory-rpc:6060
    - order-rpc:6060
```

建议访问地址：

```text
Prometheus: http://localhost:9090
Grafana:    http://localhost:3000
Jaeger:     http://localhost:16686
```

### Phase 6 Kubernetes 要求

更新 K8s manifests：

- 每个业务 Service 增加 metrics port：

```yaml
- name: metrics
  port: 6060
  targetPort: 6060
```

- 每个业务 Deployment 暴露 `containerPort: 6060`。
- 新增 `deploy/k8s/observability.yaml`，包含：
  - Prometheus ConfigMap。
  - Prometheus Deployment + Service。
  - Grafana Deployment + Service。
  - Jaeger Deployment + Service。
- `make k8s-up` 应 apply `observability.yaml`，或提供 `make k8s-observability-up`。建议默认 apply，便于学习闭环。
- 资源 requests/limits 要小，适合 minikube。

K8s 访问方式：

```bash
kubectl port-forward -n go-svc svc/prometheus 9090:9090
kubectl port-forward -n go-svc svc/grafana 3000:3000
kubectl port-forward -n go-svc svc/jaeger 16686:16686
```

如果用户不想长期运行观测组件，`make k8s-down` 删除 namespace 即可清理。

### Phase 6 日志要求

- 所有服务 stdout 输出 JSON 日志。
- 业务关键路径补充少量有意义日志，不要刷屏：
  - 用户注册/登录成功或失败。
  - 商品创建。
  - 设置库存。
  - 创建订单成功、库存不足。
- 日志字段应包含必要业务 ID，例如 `user_id`、`product_id`、`order_id`。
- 不记录明文密码、token 或敏感信息。
- 使用 go-zero `logx.WithContext(ctx)` / `logx.Infow` / `logx.Errorw` 等本地风格，不引入新的日志库。

### Phase 6 指标要求

最低要求：

- 每个服务 `/metrics` 可访问。
- Prometheus target 页面能看到 5 个业务服务 + gateway-api 为 UP。
- API 请求后 metrics 中能看到 HTTP/RPC 请求计数或延迟相关指标。

可选增强：

- 增加少量业务指标，例如：
  - `go_svc_order_created_total`
  - `go_svc_order_stock_insufficient_total`

如果实现业务指标，必须放在独立包中，例如 `pkg/observability/metrics`，避免散落在 logic 文件里。

### Phase 6 Trace 要求

最低要求：

- gateway-api 收到 HTTP 请求后能生成 trace。
- gateway-api 调用 user/product/inventory/order RPC 时 trace 能向下传播。
- order-rpc 调 product-rpc / inventory-rpc 时 trace 能继续传播。
- Jaeger UI 能查询到至少一个跨服务 trace。

实现建议：

- 优先使用 go-zero 的 `Telemetry` 配置。
- Compose/K8s 使用 OTLP gRPC exporter，Endpoint 指向 `jaeger:4317`。
- 不手写 OpenTelemetry SDK 初始化，除非 go-zero 原生配置无法满足。

### Phase 6 Makefile

至少新增：

- `make obs-compose-urls`
- `make obs-k8s-port-forward`
- `make verify-observability-compose`
- `make verify-observability-k8s`

说明：

- `obs-compose-urls` 打印 Prometheus/Grafana/Jaeger 本地访问地址。
- `obs-k8s-port-forward` 可以同时或分别提示 port-forward 命令，不要求复杂进程管理。
- `verify-observability-compose` 检查：
  - Prometheus `/api/v1/targets` 可访问。
  - Jaeger UI 可访问。
  - gateway-api `/healthz` 可访问。
  - 至少一个 metrics endpoint 可访问。
- `verify-observability-k8s` 检查：
  - Prometheus/Grafana/Jaeger Service 存在。
  - 业务服务 metrics port 存在。
  - 如启动 port-forward，必须用 trap 清理。

### Phase 6 验收命令

基础验证：

```bash
make fmt
make test
git diff --check
make gen
docker compose -f deploy/docker-compose/docker-compose.yml config
kubectl apply --dry-run=client -f deploy/k8s/
```

Compose 验证：

```bash
make compose-up
make e2e-compose
make verify-observability-compose
make compose-down
```

K8s 验证：

```bash
make k8s-up
make e2e-k8s
make verify-observability-k8s
make k8s-down
```

如果本机没有可用 K8s API Server：

- 说明 `kubectl dry-run` / `make k8s-up` 未执行或失败原因。
- 至少运行 YAML 语法解析检查。

### Phase 6 Claude Code 执行提示词

```text
请在当前仓库实现 docs/claude-implementation-plan.md 中的 Phase 6。

要求：
- 严格遵循 Phase 6 说明。
- 本阶段只实现可观测性：结构化日志、Prometheus 指标、Jaeger/OpenTelemetry trace、相关 Compose/K8s 部署和文档。
- 不实现新业务功能，不引入 Redis、etcd、消息队列、Helm、service mesh。
- 优先使用 go-zero 原生 Log、DevServer、Telemetry 配置，不要重新发明观测框架。
- 所有服务配置补充 Log、DevServer；Compose/K8s 配置补充 Telemetry endpoint。
- Docker Compose 增加 prometheus、grafana、jaeger，并提供 Prometheus scrape 配置。
- Kubernetes 增加 observability.yaml，并让业务 Service 暴露 metrics port 6060。
- 日志只输出到 stdout，使用 JSON 编码，不引入日志平台。
- 如增加业务指标，集中放到 pkg/observability/metrics，不要散落。
- 新增或更新 Makefile 目标：obs-compose-urls、obs-k8s-port-forward、verify-observability-compose、verify-observability-k8s。
- 更新 README 和 deploy/observability/README.md，说明如何访问 Prometheus/Grafana/Jaeger，如何验证 metrics 和 trace。
- 保持 Phase 4 Compose、Phase 5 K8s 现有验证能力不回退。
- 实现后运行 make fmt、make test、git diff --check、make gen。
- 如果 Docker 可用，运行 docker compose -f deploy/docker-compose/docker-compose.yml config。
- 如果可启动 Compose，运行 make compose-up、make e2e-compose、make verify-observability-compose、make compose-down。
- 如果 kubectl 可用，运行 kubectl apply --dry-run=client -f deploy/k8s/；如果没有可用集群导致失败，说明原因。
- 如果 K8s 可用，运行 make k8s-up、make e2e-k8s、make verify-observability-k8s、make k8s-down。
- 验证结束后必须清理 Compose/K8s/port-forward 进程，并复查 8080、9000、9001、9002、9003、5432、9090、3000、16686 无异常监听。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```

### Phase 6 Review 重点

- 是否使用 go-zero 原生配置而不是引入重复框架。
- 所有服务是否都有 JSON 日志、DevServer `/metrics`、Compose/K8s trace endpoint。
- Prometheus scrape targets 是否覆盖 gateway-api 和 4 个 RPC 服务。
- Compose 是否能一键启动观测组件。
- K8s `observability.yaml` 是否资源清晰、端口一致、适合 minikube。
- `verify-observability-*` 是否能验证 metrics/trace 基础可用性。
- port-forward 或后台进程是否有 trap 清理。
- 日志是否避免记录密码等敏感信息。
- Phase 4/5 原有命令是否没有回退。
- 验证后是否清理进程并复查端口。

## Phase 7: GitHub Actions CI + Kuboard 发布管理

详细方案见 [docs/phase-7-ci-kuboard.md](phase-7-ci-kuboard.md)。

### Phase 7 目标

建立 CI 和远端发布管理的基础能力：

- GitHub Actions 自动运行格式化、测试、脚本语法、Compose 配置和 YAML 静态校验。
- GitHub Actions 构建 5 个业务服务镜像，并推送到 GHCR。
- 远端 Kubernetes 从 GHCR 拉取镜像。
- 远端安装 Kuboard，用于查看、更新和回滚 Kubernetes 工作负载。

本阶段采用半自动发布：

```text
git push
  -> GitHub Actions CI
  -> GitHub Actions build/push images
  -> 远端机器或 Kuboard 更新 image tag
  -> kubectl/Kuboard 观察 rollout
```

### Phase 7 范围

必须实现：

- `.github/workflows/ci.yml`
- `.github/workflows/images.yml`
- `make ci-check`
- `make k8s-set-images`
- `make k8s-rollout-status`
- README / K8s / Phase 7 文档更新

建议调整：

- `IMAGE_TAG ?= phase6` 改为 `IMAGE_TAG ?= local` 或其他非阶段绑定默认值。
- `images.yml` 使用 matrix 显式描述每个服务：
  - `service`
  - `main`
  - `conf_dir`
  - `image`

不做：

- 不让 GitHub Actions 直接部署远端机器。
- 不把 kubeconfig、SSH private key 或集群管理员凭证放到 GitHub Secrets。
- 不引入 Helm、Kustomize、Argo CD、Flux。
- 不改造业务功能。

### Phase 7 CI Workflow

`ci.yml` 触发：

```yaml
on:
  push:
  pull_request:
```

检查项：

```bash
go fmt ./...
go test ./...
git diff --check
bash -n scripts/*.sh
docker compose -f deploy/docker-compose/docker-compose.yml config
ruby -e 'require "yaml"; Dir["deploy/k8s/*.yaml", "apps/*/etc/*.yaml"].each { |f| YAML.load_stream(File.read(f)); puts f }'
```

说明：

- CI 不启动 Compose 集群。
- CI 不启动 Kubernetes 集群。
- e2e 仍然保留给本地或远端环境执行。

### Phase 7 Images Workflow

`images.yml` 触发：

```yaml
on:
  push:
    branches: [main]
  workflow_dispatch:
```

权限：

```yaml
permissions:
  contents: read
  packages: write
```

镜像：

```text
ghcr.io/wangchongv8/go-svc-user-rpc:<git-sha>
ghcr.io/wangchongv8/go-svc-product-rpc:<git-sha>
ghcr.io/wangchongv8/go-svc-inventory-rpc:<git-sha>
ghcr.io/wangchongv8/go-svc-order-rpc:<git-sha>
ghcr.io/wangchongv8/go-svc-gateway-api:<git-sha>
```

同时可以推送 `main` tag，方便测试；实际远端发布优先使用 git sha tag。

### Phase 7 Kuboard

远端机器安装 Kuboard v3，推荐 Docker 独立部署：

```text
remote machine
  -> Docker container: Kuboard
  -> Kubernetes cluster: go-svc namespace
```

Kuboard 用途：

- 查看 Deployment、Service、Pod、ConfigMap、Secret、Event。
- 查看 Pod 日志和状态。
- 修改 Deployment image tag，触发滚动更新。
- 观察 rollout。
- 手动回滚。

Kuboard 不负责构建镜像；镜像由 GitHub Actions 构建并推送到 GHCR。

### Phase 7 验收命令

本地静态验证：

```bash
make ci-check
git diff --check
```

远端验证：

```bash
git pull
make k8s-up
make e2e-k8s
make verify-observability-k8s
```

发布指定 GitHub Actions 构建出的镜像：

```bash
IMAGE_TAG=<git-sha> make k8s-set-images
make k8s-rollout-status
make e2e-k8s
make verify-observability-k8s
```

GitHub Actions 验证：

- push 到 GitHub 后确认 `CI` workflow 通过。
- main 分支或手动触发 `Build Images` 后确认 5 个 GHCR 镜像均存在。

### Phase 7 Claude Code 执行提示词

```text
请在当前仓库实现 docs/phase-7-ci-kuboard.md。

要求：
- 新增 GitHub Actions workflow：.github/workflows/ci.yml 和 .github/workflows/images.yml。
- 新增或更新 Makefile：ci-check、k8s-set-images、k8s-rollout-status，并把 IMAGE_TAG 默认值改成 local 或可覆盖值，不要继续使用 phase 固定默认值。
- images workflow 使用 GHCR：ghcr.io/wangchongv8/go-svc-<service>，推送 git sha tag 和 main tag。
- images workflow 使用 matrix 明确配置每个服务的 SERVICE_MAIN、SERVICE_CONF_DIR、image 名称。
- 不实现自动部署远端机器，不引入 SSH/kubeconfig secrets。
- 更新 README、deploy/k8s/README.md 或新增 docs，说明 GitHub Actions、GHCR、远端机器、Kuboard 的使用流程。
- 保持 Phase 4/5/6 原有命令可用。
- 实现后运行 make ci-check；如果 GitHub Actions 无法本地运行，说明需要 push 后在 GitHub UI 验证。
- 不启动长期运行的 Compose/K8s 服务；如为了验证启动了，必须清理并检查端口。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```

### Phase 7 Review 重点

- Workflow 是否能在 GitHub-hosted runner 上运行。
- GHCR login 是否使用 `GITHUB_TOKEN` 和 `packages: write`，没有硬编码 token。
- Matrix 是否覆盖 5 个服务，且 main/conf_dir/image 对应正确。
- Docker build args 是否和 `deploy/docker/service.Dockerfile` 匹配。
- `make ci-check` 是否能本地复现大部分 CI 检查。
- `IMAGE_TAG` 是否不再绑定历史 phase。
- 远端发布是否仍可通过 `kubectl set image` 或 Kuboard 手动完成。
- 文档是否清楚说明 public/private GHCR、imagePullSecret、Kuboard 安装和发布流程。
- 是否没有引入 GitHub Actions 自动部署远端机器。
- Phase 4/5/6 命令是否没有回退。

## Phase 8: Loki 日志检索 + trace_id 关联

状态：已完成。

详细方案见 [docs/phase-8-log-search-trace-id.md](phase-8-log-search-trace-id.md)。

### Phase 8 目标

在已有 Prometheus/Grafana/Jaeger 基础上补齐日志检索：

- 引入 Loki 作为日志存储和 LogQL 查询后端。
- 引入 Grafana Alloy 采集 Kubernetes Pod stdout 日志。
- Grafana 增加 Loki datasource，作为日志查询入口。
- gateway-api 在 HTTP 响应中返回 `X-Trace-Id`。
- 业务关键日志写入 `trace_id`，支持从 Jaeger trace 跳到 Loki 日志检索。

### Phase 8 约束

- Claude Code 必须在新分支实现，建议 `feature/phase8-log-search-trace-id`。
- 不引入 Promtail；新项目直接使用 Alloy。
- 不引入 ELK/OpenSearch、Tempo、Helm、Argo CD、Flux 或 service mesh。
- Loki 日志数据不挂载宿主机目录，Compose/K8s 都使用可清理的临时存储。
- 保持 PostgreSQL 当前不本地持久化、通过 migration 回放重建的策略。
- `trace_id` 不作为 Prometheus label 或 Loki label，只作为 JSON 日志字段查询。

### Phase 8 验收

```bash
make fmt
make test
git diff --check
make gen
docker compose -f deploy/docker-compose/docker-compose.yml config
kubectl apply --dry-run=client -f deploy/k8s/
```

如环境可用，还应执行：

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

验证结束后必须清理 Compose/K8s/port-forward 进程，并复查：

```text
8080, 9000, 9001, 9002, 9003, 5432, 9090, 3000, 16686, 3100
```

## Phase 9: Ory Kratos 身份认证接入

状态：规划中。

详细方案见 [docs/phase-9-ory-kratos-auth.md](phase-9-ory-kratos-auth.md)。

### Phase 9 目标

把注册、登录和 session 验证从业务用户服务中拆出，接入 Ory Kratos：

- Kratos 负责 identity、密码凭证、注册登录 flow 和 session。
- gateway-api 对外提供项目自己的 auth API。
- gateway-api 通过 Kratos `/sessions/whoami` 校验 `Authorization: Bearer <session_token>`。
- user-rpc 不再为新 auth 接口保存或校验密码，只维护本地业务用户 profile。
- 本地业务用户通过 `kratos_identity_id` 映射到 Kratos identity。
- 至少一个核心业务接口改成依赖认证上下文，优先选择创建订单。

### Phase 9 约束

- Claude Code 必须在新分支实现，建议 `feature/phase9-ory-kratos-auth`。
- 第一版只做 API flow + session token，不做浏览器登录页。
- 不引入 OAuth/OIDC 社交登录、MFA、邮箱验证、找回密码、Hydra、Keto、Oathkeeper。
- Kratos 使用 PostgreSQL，但不能破坏现有业务表。
- Kratos admin API 不暴露给外部客户端或 Ingress。
- `session_token`、密码、Authorization header、Cookie 不得写入日志。
- `kratos_identity_id`、`user_id`、`session_token` 不得作为 Prometheus label。
- 旧 `/api/v1/register`、`/api/v1/login` 可以暂时保留，但必须标记为 legacy。

### Phase 9 实现范围

基础设施：

- Docker Compose 增加 `kratos` 和 `kratos-migrate`。
- Kubernetes 增加 Kratos ConfigMap、Secret、Service、Deployment、migration Job。
- 新增 Kratos 配置和 identity schema，建议路径：

```text
deploy/kratos/kratos.yml
deploy/kratos/identity.schema.json
```

gateway-api：

- 新增：

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/auth/me
```

- 注册/登录通过 Kratos API flow 完成。
- 登录成功返回 `session_token` 和本地用户 profile。
- 受保护接口从 `Authorization: Bearer <session_token>` 读取 token。
- 调 Kratos `/sessions/whoami` 验证 session。
- 把 `kratos_identity_id` 和本地 `user_id` 放入请求上下文。
- auth 必须通过 middleware 或 auth-aware handler 实现，不允许在 `gateway.go` 重复注册同一个业务路由。
- `/api/v1/auth/register` 必须在 Kratos 创建 identity 后调用 user-rpc 创建或获取本地 user profile。
- `/api/v1/auth/login` 和 `/api/v1/auth/me` 必须返回真实本地 `user.id`，不能返回 placeholder 或 `id=0`。

user-rpc：

- 新增或调整本地用户 profile 映射能力：

```text
id
kratos_identity_id
username
created_at
```

- 推荐 RPC：

```text
GetOrCreateByKratosIdentity(kratos_identity_id, username)
GetByKratosIdentity(kratos_identity_id)
GetUser(id)
```

- 这些必须是 protobuf/zrpc 暴露的真实 RPC 方法，不只是 `UserStore` 内部 helper。
- 修改 proto 后必须重新生成 pb、grpc、zrpc、server 代码。
- gateway-api 必须通过生成的 `userrpc` client 调用这些方法。

业务接口：

- 优先改造 `POST /api/v1/orders`：
  - 必须登录。
  - 未带 token 返回 401。
  - 请求体不再信任客户端传入的 `user_id`。
  - `user_id` 从认证上下文得到。
  - 只能注册一条 `POST /api/v1/orders` 路由。
  - 不允许硬编码 `userID := int64(1)`。

### Phase 9 当前实现修复要求

如果当前实现存在以下临时方案，Claude Code 必须删除并替换为正式实现：

- 在 `gateway.go` 里重复注册 `POST /api/v1/orders` 来包 auth。
- auth middleware 中调用 legacy `Register(username, "kratos-managed")` 来模拟用户映射。
- auth register/login/me 只返回 username，没有真实 local user id。
- `CreateOrderLogic` 使用固定 `userID := int64(1)`。
- user-rpc 只在 `UserStore` 增加 Kratos helper，但没有暴露为 protobuf/zrpc 方法。

正确实现链路：

```text
register
  -> Kratos registration flow
  -> identity.id
  -> user-rpc.GetOrCreateByKratosIdentity
  -> return session_token + local user

login
  -> Kratos login flow
  -> identity.id
  -> user-rpc.GetOrCreateByKratosIdentity or GetByKratosIdentity
  -> return session_token + local user

me
  -> Kratos /sessions/whoami
  -> identity.id
  -> user-rpc.GetByKratosIdentity
  -> return local user

protected order
  -> AuthMiddleware
  -> Kratos /sessions/whoami
  -> user-rpc identity mapping
  -> context local user_id
  -> CreateOrderLogic uses context user_id
```

### Phase 9 验收

必须执行：

```bash
make fmt
go test ./...
go vet ./...
docker compose -f deploy/docker-compose/docker-compose.yml config -q
bash -n scripts/*.sh
git diff --check
```

如果 Docker 可用，还应执行：

```bash
make compose-up
make e2e-compose
make compose-down
```

Compose e2e 至少覆盖：

- `/api/v1/auth/register` 注册。
- `/api/v1/auth/login` 登录并获取 `session_token`。
- `/api/v1/auth/me` 带 token 返回 200。
- 未带 token 创建订单返回 401。
- 带 token 创建订单成功。

如果 Kubernetes 可用，还应执行：

```bash
kubectl apply --dry-run=client -f deploy/k8s/
IMAGE_TAG=<tag> make k8s-up
make e2e-k8s
make k8s-down
```

验证结束后必须清理 Compose/K8s/port-forward 进程，并复查相关端口。

### Phase 9 Review 重点

- Kratos API flow 是否按官方模型实现，没有绕过 flow 或直接操作 Kratos 内部表。
- session token 是否没有进入日志、metrics、trace attributes。
- gateway 是否通过 `/sessions/whoami` 做鉴权。
- 业务接口是否不再信任客户端传入的 `user_id`。
- 本地 user profile 和 Kratos identity 映射是否唯一且幂等。
- Compose/K8s Kratos 配置是否一致。
- Kratos admin API 是否没有暴露给外部入口。
- 旧 register/login 与新 auth 接口边界是否清楚。
- e2e 是否覆盖 401 和带 token 成功路径。

## Claude Code Prompt 模板

```text
请在当前仓库实现 docs/claude-implementation-plan.md 中的 Phase <编号>。

要求：
- 严格遵循 docs/architecture.md 和 docs/requirements.md。
- 不实现 Phase <编号> 之外的内容。
- 新增或修改的命令必须写入 README 或相关 docs。
- 实现后运行可用的格式化、测试和启动验证命令。
- 在回复中列出修改文件、验证命令、结果和未完成事项。
- 如果设计文档存在冲突，先说明冲突点，不要自行扩大范围。
```
