# 可观测性

## Compose 环境

启动后访问：

| 组件 | 地址 | 说明 |
|------|------|------|
| Prometheus | http://localhost:9090 | 指标采集 + 查询 |
| Grafana | http://localhost:3000 | 指标可视化 (admin/admin) |
| Jaeger | http://localhost:16686 | 链路追踪查询 |
| Loki | 通过 Grafana Explore 访问 | 日志存储 + LogQL 查询，Phase 8 引入 |

### 验证

```bash
# 检查所有 scrape target 为 UP
curl -s http://localhost:9090/api/v1/targets | grep '"job":"go-svc"' -A2

# gateway-api metrics
curl -s http://localhost:6060/metrics | head -10
```

## K8s 环境

```bash
kubectl port-forward -n go-svc svc/prometheus 9090:9090
kubectl port-forward -n go-svc svc/grafana 3000:3000
kubectl port-forward -n go-svc svc/jaeger 16686:16686
```

## 配置说明

- `Log.Mode: console` + `Encoding: json` → stdout 输出结构化 JSON
- `DevServer.EnableMetrics: true` → 暴露 `/metrics` 端点
- `Telemetry.Endpoint: jaeger:4317` → OTLP gRPC trace 上报
- 本地 `*.yaml` 不配置 Telemetry，避免无 Jaeger 时启动报错

## Phase 8 日志检索规划

Phase 8 将补充 Loki + Alloy：

- Loki 负责存储和查询日志。
- Grafana 作为日志检索入口。
- Alloy 作为 Kubernetes 日志采集器，读取 Pod 容器日志并发送到 Loki。
- 不新增 Promtail。Promtail 已进入淘汰路径，新项目直接使用 Alloy。
- Loki 数据不挂载到宿主机，和 PostgreSQL 学习环境一致，允许随环境重建清理。

示例 LogQL：

```logql
{namespace="go-svc", app="gateway-api"} | json | trace_id="abc123"
```

```logql
{namespace="go-svc", app="order-rpc"} |= "create order"
```

## 已知限制

- Compose 环境 Loki 已部署且 Grafana 已配置 datasource，但 Compose 容器日志**不会自动采集到 Loki**。Compose 的 Loki 验证仅检查服务可用性。
- K8s 环境通过 Alloy DaemonSet 自动采集 Pod stdout 日志到 Loki。
- Compose 日志自动采集是后续 Phase 的待办事项。
