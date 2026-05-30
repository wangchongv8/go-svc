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

## 阶段记录

| 阶段 | 状态 | 内容 |
|------|------|------|
| Phase 0 | ✅ 完成 | 项目脚手架和文档体系 |
| Phase 1 | ✅ 完成 | 最小 Go HTTP 服务 + Makefile + 测试 |
| Phase 2 | ✅ 完成 | go-zero API + gRPC (gateway-api ↔ user-rpc) |
| Phase 3 | 📋 待开始 | 多服务业务闭环 + PostgreSQL |

## 当前决策

- 业务主题：轻量电商下单系统。
- HTTP 服务：使用 go-zero API。
- 内部 RPC：使用 gRPC。
- 数据库：PostgreSQL。
- 本地服务发现：Docker Compose 阶段可以用 etcd 学习 go-zero 服务注册与发现。
- Kubernetes 服务发现：优先使用 Kubernetes Service + DNS，不让业务服务直接依赖 etcd。

## Phase 2: go-zero API + RPC 基础（已完成）

已引入 go-zero 框架，创建 gateway-api（HTTP）和 user-rpc（gRPC），实现 API 调用 RPC 的完整链路。

### 目录结构

```
.
├── apps/
│   ├── gateway-api/              # go-zero HTTP API 服务
│   │   ├── gateway.api           # API DSL 定义
│   │   ├── gateway.go            # 入口 (goctl 生成)
│   │   ├── etc/gateway-api.yaml  # 运行时配置
│   │   └── internal/
│   │       ├── config/config.go  # 配置结构体
│   │       ├── handler/          # HTTP handler
│   │       ├── logic/            # 业务逻辑（调用 RPC）
│   │       ├── svc/              # ServiceContext (DI)
│   │       └── types/            # 请求/响应类型
│   └── user-rpc/                 # go-zero gRPC 服务
│       ├── user.proto            # Protobuf 定义
│       ├── user.go               # 入口 (goctl 生成)
│       ├── etc/user.yaml         # 运行时配置
│       ├── user/                 # protoc 生成的 pb 文件
│       ├── userrpc/              # gRPC 客户端 stub
│       ├── model/userstore.go    # 内存用户存储
│       └── internal/
│           ├── config/config.go
│           ├── logic/            # RPC 方法实现
│           ├── server/           # gRPC server 注册
│           └── svc/              # ServiceContext (DI)
├── docs/
├── Makefile
├── go.mod
└── README.md
```

### 服务关系

```
curl → gateway-api (:8080, HTTP)
         └── user-rpc (:9000, gRPC)
                └── UserStore (in-memory)
```

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/healthz` | 健康检查 |
| POST | `/api/v1/register` | 用户注册 |
| POST | `/api/v1/login` | 用户登录 |
| GET | `/api/v1/users/:id` | 获取用户信息 |

### 快速开始

```bash
# 格式化代码
make fmt

# 运行测试
make test

# 终1：启动 user-rpc (gRPC :9000)
make run-user-rpc

# 终端 2：启动 gateway-api (HTTP :8080)
make run-gateway-api

# 终端 3：验证
curl http://localhost:8080/healthz
# → {"status":"ok"}

curl -X POST http://localhost:8080/api/v1/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"123456"}'
# → {"id":1,"username":"alice"}

curl -X POST http://localhost:8080/api/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"123456"}'
# → {"id":1}

curl http://localhost:8080/api/v1/users/1
# → {"id":1,"username":"alice"}
```

### 错误响应

业务错误统一返回 JSON，HTTP 状态码由 gRPC status code 映射：

| gRPC Code | HTTP Status | 示例 |
|-----------|-------------|------|
| InvalidArgument | 400 | `{"error":"username must not be empty"}` |
| Unauthenticated | 401 | `{"error":"invalid password"}` |
| NotFound | 404 | `{"error":"user not found"}` |
| AlreadyExists | 409 | `{"error":"username already exists"}` |
| Internal | 500 | `{"error":"internal error"}` |

### Makefile 命令

| 命令 | 说明 |
| --- | --- |
| `make fmt` | 格式化所有 Go 代码 |
| `make test` | 运行所有测试 |
| `make run-user-rpc` | 启动 user-rpc (gRPC :9000) |
| `make run-gateway-api` | 启动 gateway-api (HTTP :8080) |
| `make gen` | 从 .api 和 .proto 重新生成代码 |
