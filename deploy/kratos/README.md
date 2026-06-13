# Kratos 部署

## Compose

Kratos 作为独立服务运行，使用 PostgreSQL `kratos` 数据库。

```bash
# 端口：
# - 4433: Kratos Public API (gateway 调用)
# - 4434: Kratos Admin API (仅本地)

# 验证
curl http://localhost:4433/health/alive
```

## K8s

```bash
kubectl apply -f deploy/k8s/kratos.yaml
kubectl get pods -n go-svc -l app=kratos
```

## 配置

- `kratos.yml`: Kratos 服务配置
- `identity.schema.json`: 身份 schema (username + password)
- 数据库：`kratos` database in PostgreSQL
- Secret：默认密钥 `dev-secret-not-for-production`（学习环境）
