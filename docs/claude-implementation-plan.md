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

状态：可交给 Claude Code 实施。

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

建议验收命令待补充。

## Phase 3: 多服务业务闭环

状态：待 Phase 2 review 后细化。

目标：

- 完成核心业务服务拆分。
- 实现一个端到端业务流程，例如创建订单。
- 明确服务间调用链。
- 接入 PostgreSQL。

建议验收命令待补充。

## Phase 4: Docker Compose 本地集群

状态：待需求确认后细化。

目标：

- 为服务增加 Dockerfile。
- 增加 docker-compose。
- 启动 PostgreSQL、Redis、etcd 和业务服务。

建议验收命令待补充。

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
