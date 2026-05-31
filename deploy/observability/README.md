# 可观测性

## Compose 环境

启动后访问：

| 组件 | 地址 | 说明 |
|------|------|------|
| Prometheus | http://localhost:9090 | 指标采集 + 查询 |
| Grafana | http://localhost:3000 | 指标可视化 (admin/admin) |
| Jaeger | http://localhost:16686 | 链路追踪查询 |

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
