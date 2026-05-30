# go-svc

学习型 Go 微服务集群项目，业务主题为轻量电商下单系统。

这个仓库会用来练习和串联 Go、go-zero、gRPC、PostgreSQL、Docker、Kubernetes、服务发现、配置管理、可观测性等微服务开发技术。

当前阶段不是直接写业务代码，而是先建立一套清晰的协作流程：

- Codex 负责需求澄清、架构方案、实现计划、验收标准和最终 review。
- Claude Code 负责按文档分阶段实现代码和基础设施文件。
- 每个阶段完成后，通过可运行命令、测试结果和 review 清单确认质量。

## 文档入口

- [项目协作流程](docs/workflow.md)
- [需求工作台](docs/requirements.md)
- [架构设计草案](docs/architecture.md)
- [Claude Code 实施计划](docs/claude-implementation-plan.md)
- [Review 检查清单](docs/review-checklist.md)

## 推荐推进节奏

1. 先在 `docs/requirements.md` 对齐学习目标、服务边界和技术范围。
2. 再更新 `docs/architecture.md`，确定服务拆分、通信方式、数据组件和部署形态。
3. 然后把任务拆进 `docs/claude-implementation-plan.md`，交给 Claude Code 小步实现。
4. 每完成一个阶段，使用 `docs/review-checklist.md` 做 review。

## 当前决策

- 业务主题：轻量电商下单系统。
- HTTP 服务：使用 go-zero API。
- 内部 RPC：使用 gRPC。
- 数据库：PostgreSQL。
- 本地服务发现：Docker Compose 阶段可以用 etcd 学习 go-zero 服务注册与发现。
- Kubernetes 服务发现：优先使用 Kubernetes Service + DNS，不让业务服务直接依赖 etcd。

## Phase 1: 最小 Go HTTP 服务（已完成）

当前阶段已实现一个最小 Go HTTP 服务，作为后续 `gateway-api` 的前身。

### 目录结构

```
.
├── cmd/gateway/main.go          # 服务入口
├── internal/gateway/
│   ├── handler.go               # HealthzHandler
│   └── handler_test.go          # 测试
├── docs/
├── Makefile
├── go.mod
└── README.md
```

### 快速开始

```bash
# 格式化代码
make fmt

# 运行测试
make test

# 启动服务（默认监听 8080）
make run

# 自定义端口
PORT=9090 make run
```

### 验证

```bash
# 启动服务后，在另一个终端执行：
curl http://localhost:8080/healthz
# 预期输出: {"status":"ok"}
```

### Makefile 命令

| 命令 | 说明 |
| --- | --- |
| `make fmt` | 格式化所有 Go 代码 |
| `make test` | 运行所有测试 |
| `make run` | 启动 gateway 服务 |
