# 架构设计草案

本文档记录目标架构。当前版本对应第一版轻量电商下单系统。

## 架构目标

- 使用 Go 构建多个可独立运行的服务。
- HTTP 入口使用 go-zero API。
- 内部服务通过 gRPC 通信。
- 使用 PostgreSQL 持久化业务数据。
- 本地开发可以用 Docker Compose 一键启动。
- 后续部署到 Kubernetes 时使用 Service + DNS 做服务发现。
- 每个服务都具备清晰的配置、日志、健康检查和文档。

## 初始服务拆分

第一版服务：

| 服务 | 类型 | 职责 |
| --- | --- | --- |
| gateway-api | HTTP API | 对外统一入口，聚合 RPC 服务能力 |
| user-rpc | gRPC | 用户注册、登录、用户资料 |
| product-rpc | gRPC | 商品查询、商品管理 |
| inventory-rpc | gRPC | 库存查询、库存扣减、库存回滚 |
| order-rpc | gRPC | 创建订单、查询订单、取消订单、订单状态 |
| payment-rpc | gRPC | 模拟支付、支付状态、支付幂等 |

## 基础组件

第一版组件：

| 组件 | 用途 |
| --- | --- |
| PostgreSQL | 持久化业务数据 |
| Redis | 缓存、幂等或锁，后续可选 |
| etcd | Docker Compose 阶段学习 go-zero 服务注册与发现，后续可选 |
| Docker Compose | 本地微服务集群 |
| Kubernetes | 容器编排练习 |
| Prometheus | 指标采集，后续可选 |
| Grafana | 指标看板，后续可选 |
| Jaeger | 链路追踪，后续可选 |
| Loki | 日志存储和 LogQL 查询，后续可选 |
| Grafana Alloy | 采集 Kubernetes Pod 日志并发送到 Loki，后续可选 |
| Ory Kratos | 注册、登录、身份和 session 验证，Phase 9 引入 |

## 通信模型

- 外部客户端调用 `gateway-api` 的 HTTP 接口。
- `gateway-api` 调用内部 RPC 服务。
- RPC 服务之间按业务需要调用，例如 `order-rpc` 调用 `inventory-rpc` 和 `payment-rpc`。
- Docker Compose 阶段可以使用 etcd 完成 go-zero/gRPC 服务注册与发现。
- Kubernetes 阶段优先使用 Kubernetes Service + DNS，例如 `user-rpc:9001`，业务服务不直接访问 Kubernetes 内部 etcd。
- 配置通过本地 YAML、Docker Compose env 和 Kubernetes ConfigMap 分层管理。

## 身份认证模型

Phase 9 规划引入 Ory Kratos，把身份认证从业务用户服务中拆出：

```text
client
  -> gateway-api
     -> Kratos public API: registration, login, sessions/whoami
     -> user-rpc: local business user profile
     -> other RPC services
```

边界原则：

- Kratos 负责 identity、密码凭证、注册登录 flow 和 session。
- `gateway-api` 负责对外暴露项目自己的 `/api/v1/auth/*` 接口，并调用 Kratos public API。
- `user-rpc` 不再为新认证接口保存或校验密码，只保留业务用户资料。
- 本地业务用户通过 `kratos_identity_id` 映射到 Kratos identity。
- 业务接口不应信任客户端传入的 `user_id`；应从已验证的 session 上下文中解析当前用户。

注册和登录的业务边界：

```text
Kratos 管：identity、密码、session。
user-rpc 管：local user_id、业务用户资料、kratos_identity_id 映射。
gateway-api 管：把 Kratos session 转换成本地 user_id 并写入 context。
```

因此 `/api/v1/auth/register` 不能只创建 Kratos identity。它必须在 Kratos 注册成功后调用 `user-rpc.GetOrCreateByKratosIdentity`，创建或返回本地业务用户，再把真实本地 `user.id` 返回给客户端。

第一版只做 API flow + session token，不做浏览器登录页、OAuth、MFA、邮箱验证、找回密码、Hydra、Keto 或 Oathkeeper。

## 可观测性模型

项目的可观测性按三类数据拆分：

| 类型 | 组件 | 用途 |
| --- | --- | --- |
| Metrics | Prometheus + Grafana | 查看服务健康、请求量、延迟和错误趋势 |
| Traces | Jaeger + OpenTelemetry | 查看一次请求跨 gateway 和 RPC 服务的调用链 |
| Logs | Loki + Grafana + Alloy | 检索容器日志，并通过 `trace_id` 和 trace 关联 |

日志链路：

```text
Go 服务 stdout JSON 日志
  -> Docker/Kubernetes 容器日志
  -> Alloy 采集 Pod 日志并补充 namespace/app/pod/container 标签
  -> Loki 存储日志流并执行 LogQL 查询
  -> Grafana Explore 作为查询入口
```

trace 和日志关联原则：

- HTTP 请求和 gRPC 调用链由 OpenTelemetry 生成 trace。
- gateway-api 应把当前请求的 trace id 返回到 HTTP 响应头，例如 `X-Trace-Id`。
- 业务关键日志应包含 `trace_id`、`span_id`、`service`、`user_id`、`product_id`、`order_id` 等必要字段。
- `trace_id` 不作为 Prometheus label，也不建议作为 Loki label；它应保留在 JSON 日志字段中，通过 LogQL 管道解析过滤，避免高基数索引问题。

学习环境存储策略：

- PostgreSQL 不挂载宿主机目录，通过 migration 重建数据。
- Loki 也不挂载宿主机目录，Compose 使用容器内临时数据，Kubernetes 使用 `emptyDir` 或容器临时存储。
- 删除 Compose/Kubernetes 环境后，日志和数据库数据都可以被清理。

## Kubernetes 服务发现边界

Kubernetes 集群内部确实使用 etcd 保存 Pod、Service、EndpointSlice 等对象状态，但业务服务不直接读写 etcd。

Kubernetes 阶段的服务访问模型：

```text
gateway-api Pod
  -> DNS 解析 user-rpc.default.svc.cluster.local
  -> 得到 user-rpc Service ClusterIP
  -> kube-proxy 或 CNI 根据 EndpointSlice 转发到某个 user-rpc Pod
```

外部 HTTP 流量入口模型：

```text
client
  -> public DNS
  -> cloud LoadBalancer
  -> Ingress Controller, for example nginx
  -> gateway-api Service
  -> gateway-api Pod
```

因此，etcd 在本项目里的定位是：

- Compose 阶段：作为可显式学习的服务注册发现组件。
- K8s 阶段：作为 Kubernetes 控制面内部存储，不作为业务代码依赖。

## 推荐目录结构

```text
.
├── README.md
├── apps/
│   ├── gateway-api/
│   ├── user-rpc/
│   ├── product-rpc/
│   ├── inventory-rpc/
│   ├── order-rpc/
│   └── payment-rpc/
├── pkg/
│   ├── errors/
│   ├── middleware/
│   └── observability/
├── deploy/
│   ├── docker-compose/
│   └── k8s/
├── docs/
├── scripts/
├── tests/
├── Makefile
├── go.mod
└── go.work
```

实际目录需要结合 go-zero 生成结构再确认。

## 关键设计原则

- 先保证能跑，再逐步增加治理能力。
- 每个阶段都保留明确的启动命令和验证命令。
- 服务之间不要共享数据库表访问逻辑。
- API 层不直接访问业务数据库，尽量通过 RPC 聚合。
- RPC 接口优先稳定清晰，避免过早复杂化。
- 配置按环境拆分，但先避免过度抽象。
- PostgreSQL 表结构按服务边界划分，避免跨服务直接读写对方数据。
- 订单状态使用有限状态机思路，避免状态随意跳转。

## 待决策项

| 决策 | 状态 | 备注 |
| --- | --- | --- |
| 业务主题 | 已确认 | 轻量电商下单系统 |
| HTTP 框架 | 已确认 | go-zero API |
| 内部 RPC | 已确认 | gRPC |
| 数据库 | 已确认 | PostgreSQL |
| 本地编排 | 待确认 | 推荐 Docker Compose |
| Compose 服务发现 | 倾向确认 | 使用 etcd 学习服务注册与发现 |
| K8s 服务发现 | 已确认 | Kubernetes Service + DNS |
| K8s 环境 | 待确认 | kind、minikube 或只产出 manifests |
| 消息队列 | 待确认 | 可放到后续阶段 |
| 可观测性 | 待确认 | 可放到后续阶段 |
