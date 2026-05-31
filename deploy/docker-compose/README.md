# Docker Compose 本地集群

一条命令拉起完整微服务环境。

## 启动

```bash
make compose-up
```

服务启动顺序：postgres → db-migrate → product-rpc / inventory-rpc → order-rpc → gateway-api

## 端口

| 服务 | 端口 |
|------|------|
| gateway-api | 8080 |
| user-rpc | 9000 |
| product-rpc | 9001 |
| inventory-rpc | 9002 |
| order-rpc | 9003 |
| postgres | 5432 |

## 验证

```bash
make e2e-compose
```

## 停止

```bash
make compose-down
```

每次启动都是全新环境，数据库通过 `db-migrate` 回放 SQL 脚本初始化。
