# Kubernetes 部署

使用 kind 本地集群运行完整微服务（PostgreSQL + 5 个业务服务）。

## 前置条件

```bash
brew install kind kubectl
kind create cluster --name go-svc
```

## 部署

```bash
make k8s-build        # 构建 5 个 Docker 镜像
make k8s-kind-load    # 加载镜像到 kind 节点
make k8s-up           # 部署全部 K8s 资源
make k8s-ps           # 查看 Pod 和 Service
```

## 验证

```bash
make e2e-k8s          # 端到端验证（自动 port-forward）
```

## 手动验证

```bash
make k8s-port-forward # 在另一个终端
curl http://localhost:8080/healthz
```

## 停止

```bash
make k8s-down         # 删除 namespace（清理全部资源）
```

## 服务访问

K8s Service DNS 自动解析：

| 服务 | 集群内地址 |
|------|-----------|
| gateway-api | gateway-api:8080 |
| user-rpc | user-rpc:9000 |
| product-rpc | product-rpc:9001 |
| inventory-rpc | inventory-rpc:9002 |
| order-rpc | order-rpc:9003 |
| postgres | postgres:5432 |

## 资源清单

| 资源 | 文件 | 说明 |
|------|------|------|
| Namespace | `namespace.yaml` | go-svc |
| Secret | `postgres.yaml` | PG 账号密码 |
| Deployment + Service | `postgres.yaml` | PostgreSQL (emptyDir, 无持久化) |
| ConfigMap + Job | `db-migrate-job.yaml` | 等待 PG ready 后执行建表 SQL |
| ConfigMap + Deployment + Service | `user-rpc.yaml` 等 | 5 个业务服务 |
| Ingress | `ingress.yaml` | go-svc.local → gateway-api |

## 已知简化点

- DSN 在 ConfigMap 中直写，密码和 Secret 一致（学习环境）
- PostgreSQL 使用 emptyDir，Pod 重启/删除后数据丢失，通过 migration Job 重建
- Ingress 需要 Ingress Controller（kind 默认未安装），建议用 `port-forward` 验证
- 镜像推送到 GHCR 需要先 `docker login ghcr.io`

## 远端 K8s 部署

镜像默认使用 `ghcr.io/wangchongv8/go-svc-<service>:phase5`。

```bash
# 开发机：构建 + 推送
make k8s-build          # 构建并打上 GHCR tag
make k8s-push           # 推送到 ghcr.io/wangchongv8

# 目标机器：部署 + 验证
make k8s-up
make e2e-k8s
make k8s-down
```

### 前置条件

- 目标机器能访问 `ghcr.io`（网络可达）
- GHCR Package 需设为 **public**（否则需要 imagePullSecret）
- 开发机已 `docker login ghcr.io`（本地推送时）
- 目标机器已安装 `kubectl` 并配置好 kubeconfig

### Private GHCR 镜像

如果 Package 设为 private，目标机器需要创建 imagePullSecret：

```bash
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=<github-username> \
  --docker-password=<github-token> \
  -n go-svc
```

然后在 Deployment 的 `spec.template.spec` 下增加：

```yaml
      imagePullSecrets:
        - name: ghcr-secret
```

也可以直接 `kubectl patch serviceaccount default -p '{"imagePullSecrets":[{"name":"ghcr-secret"}]}' -n go-svc` 让所有 Pod 自动使用。

### SQL 同步说明

`db-migrate-job.yaml` 内嵌的 SQL 源自 `deploy/sql/001_phase3_schema.sql`。
修改 schema 时需要同时更新两个文件，避免漂移。
