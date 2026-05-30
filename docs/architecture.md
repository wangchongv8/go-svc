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

## 通信模型

- 外部客户端调用 `gateway-api` 的 HTTP 接口。
- `gateway-api` 调用内部 RPC 服务。
- RPC 服务之间按业务需要调用，例如 `order-rpc` 调用 `inventory-rpc` 和 `payment-rpc`。
- Docker Compose 阶段可以使用 etcd 完成 go-zero/gRPC 服务注册与发现。
- Kubernetes 阶段优先使用 Kubernetes Service + DNS，例如 `user-rpc:9001`，业务服务不直接访问 Kubernetes 内部 etcd。
- 配置通过本地 YAML、Docker Compose env 和 Kubernetes ConfigMap 分层管理。

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
