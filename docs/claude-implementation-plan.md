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

状态：待需求确认后细化。

目标：

- 增加 K8s manifests。
- 覆盖 Deployment、Service、ConfigMap、Secret。
- 提供本地 kind 或 minikube 验证说明。

建议验收命令待补充。

## Phase 6: 可观测性

状态：待需求确认后细化。

目标：

- 增加结构化日志。
- 增加 Prometheus 指标。
- 增加链路追踪。
- 提供 Grafana 或 Jaeger 使用说明。

建议验收命令待补充。

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
