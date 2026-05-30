# Phase 3 Review

Review 时间：2026-05-30

## 结论

需要修改后再进入 Phase 4。

Phase 3 主体方向正确：

- 已新增 `product-rpc`、`inventory-rpc`、`order-rpc`。
- 已新增 PostgreSQL schema。
- 已使用 `database/sql` + `pgx`，未引入 ORM。
- 已实现 `gateway-api -> order-rpc -> product-rpc/inventory-rpc` 的下单链路。
- 库存扣减使用条件更新，具备防超卖基础。
- 单元测试使用 fake repository / fake RPC client，符合本阶段测试策略。

但当前还有几个会影响业务闭环和 API 语义的问题，需要先修复。

## 高优先级问题

### 1. 创建订单接口丢失 `user_id`，实际写入订单的用户固定为 `0`

文件：

- `apps/gateway-api/gateway.api:82`
- `apps/gateway-api/internal/logic/order/createorderlogic.go:28`
- `apps/gateway-api/internal/logic/order/createorderlogic.go:31`
- `README.md:212`
- `README.md:224`

问题：

- Phase 3 设计要求 `CreateOrder(user_id, product_id, quantity) -> order`。
- 文档也写了 `user_id` 手动传入。
- 但 `CreateOrderReq` 只有 `product_id` 和 `quantity`。
- gateway 创建订单时把 `UserId` 固定传成 `0`。

影响：

- `POST /api/v1/orders` 创建出的订单 `user_id` 会是 `0`。
- `GET /api/v1/users/:user_id/orders` 无法查到用户刚创建的订单。
- README 的手动验证步骤和实际 API 行为不一致。

建议：

- 在 `gateway.api` 的 `CreateOrderReq` 中加入 `UserID int64 json:"user_id"`。
- gateway 创建订单时传递 `req.UserID`。
- order-rpc 校验 `user_id > 0`，非法时返回 `InvalidArgument`。
- README curl 示例改为包含 `user_id`。
- 增加测试覆盖 gateway/order 逻辑不再写死 `0`。

### 2. 新增 handler 把请求解析错误映射为 HTTP 500

文件：

- `apps/gateway-api/internal/handler/product/createproducthandler.go:19`
- `apps/gateway-api/internal/handler/inventory/setstockhandler.go:19`
- `apps/gateway-api/internal/handler/order/createorderhandler.go:19`
- `apps/gateway-api/internal/handler/order/listuserordershandler.go:19`

问题：

- 这些 handler 在 `httpx.Parse` 失败时调用 `httperr.Write`。
- `httperr.Write` 用 `status.Convert(err)` 处理普通 Go error，会得到 `codes.Unknown`。
- `codes.Unknown` 当前映射为 HTTP 500。

影响：

- JSON 格式错误、路径参数错误、请求参数解析错误会变成 `500 Internal Server Error`。
- 这类错误应该是客户端错误，通常应返回 `400 Bad Request`。

建议：

- 对 `httpx.Parse` 错误使用 `httpx.ErrorCtx`，或扩展 `httperr.Write` 让非 gRPC status error 映射为 `400`。
- 保持业务/RPC 错误继续走 gRPC code 到 HTTP JSON 的映射。
- 补充至少一个 handler 或 httperr 单元测试，覆盖普通 parse error 不返回 500。

## 中优先级问题

### 1. PostgreSQL 初始化错误被吞掉，服务可能带着无效 DB 继续启动

文件：

- `apps/product-rpc/internal/svc/servicecontext.go:20`
- `apps/product-rpc/internal/svc/servicecontext.go:21`
- `apps/inventory-rpc/internal/svc/servicecontext.go:20`
- `apps/inventory-rpc/internal/svc/servicecontext.go:21`
- `apps/order-rpc/internal/svc/servicecontext.go:25`
- `apps/order-rpc/internal/svc/servicecontext.go:26`

问题：

- `sql.Open` 的错误被忽略。
- 没有 `PingContext` 做启动期连通性检查。
- 如果 DSN 或数据库环境有问题，服务可能启动成功，但请求时才失败。

建议：

- 初始化 PostgreSQL repository 时不要吞掉错误。
- 至少在启动时 `PingContext`，失败则明确退出。
- 如果为了单元测试需要 fake repository，继续通过空 DSN 或测试构造函数注入 fake，不要在生产配置中静默降级。

### 2. README 的 Makefile 命令表没有列出 Phase 3 新命令

文件：

- `README.md:229`

问题：

- 表格只列出 `run-user-rpc`、`run-gateway-api`、`gen`。
- Phase 3 新增的 `run-product-rpc`、`run-inventory-rpc`、`run-order-rpc`、`db-migrate` 没有列出。

建议：

- 更新 README 命令表，保持和 Makefile 一致。

## 低优先级问题

### 1. Phase 3 状态被提前标记为完成

文件：

- `docs/claude-implementation-plan.md`
- `README.md`

问题：

- 当前 review 尚未通过，但文档中已经标记 Phase 3 完成。

建议：

- 修复 review 问题后再保留完成状态。
- 或在修复前标记为 `review 修复中`。

## 验证结果

已执行并通过：

```bash
make fmt
make test
git diff --check
make gen
```

`make test` 结果：

```text
ok go-svc/apps/product-rpc/internal/logic
ok go-svc/apps/inventory-rpc/internal/logic
ok go-svc/apps/order-rpc/internal/logic
ok go-svc/apps/user-rpc/internal/logic
```

未执行数据库端到端验证：

- 当前环境没有 `psql`、`createdb`、`pg_isready` 命令。
- 本阶段不强制 PostgreSQL 集成测试，因此未启动业务服务做端到端 curl。

进程和端口：

- 本次 review 未启动服务。
- 已复查 `8080`、`9000`、`9001`、`9002`、`9003`，均无监听进程。

## 建议同步给 Claude Code 的修复 Prompt

```text
请根据 docs/reviews/phase-3-review.md 修复 Phase 3 review 问题。

要求：
- 不进入 Phase 4。
- 不引入 Docker Compose、Kubernetes、Redis、etcd、消息队列、payment-rpc。
- 修复创建订单的 user_id：
  - gateway.api 的 CreateOrderReq 增加 user_id。
  - gateway-api 调 order-rpc 时传递 req.UserID，不允许写死 0。
  - order-rpc 校验 user_id > 0，非法返回 InvalidArgument。
  - README 的下单 curl 示例包含 user_id。
  - 增加或更新测试覆盖 user_id 不再丢失。
- 修复请求解析错误映射：
  - httpx.Parse 失败不能返回 500。
  - 可以对 parse error 使用 httpx.ErrorCtx，或让 httperr 对非 gRPC status error 返回 400 JSON。
  - 补测试覆盖普通请求解析错误不会返回 500。
- 修复 PostgreSQL 初始化：
  - 不要吞掉 sql.Open 错误。
  - 有 DSN 时启动期执行 PingContext，失败要明确返回/退出。
  - 单元测试继续通过 fake repository/fake RPC client，不要引入 SQLite。
- README 的 Makefile 命令表补充 run-product-rpc、run-inventory-rpc、run-order-rpc、db-migrate。
- 根据修复状态调整 README 和 docs/claude-implementation-plan.md 中 Phase 3 状态。
- 运行 make fmt、make test、git diff --check、make gen。
- 如果本地 PostgreSQL 不可用，说明未运行数据库验证命令、原因和手动验证步骤。
- 如果启动服务做验证，结束后必须停止进程，并复查 8080、9000、9001、9002、9003 无监听。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```

## 修复后复查

复查时间：2026-05-30

结论：

- 仍需小修改后再进入 Phase 4。

已确认修复：

- `CreateOrderReq` 已增加 `user_id`。
- `gateway-api` 创建订单时已传递 `req.UserID`，不再写死 `0`。
- `order-rpc` 已校验 `user_id > 0`，非法返回 `InvalidArgument`。
- README 下单 curl 示例已包含 `user_id`。
- 请求解析错误已通过 `httperr` 映射为 `400 Bad Request`。
- 已补充 `httperr` 测试覆盖非 gRPC error。
- README Makefile 命令表已补充 Phase 3 新命令。

仍需修复：

### PostgreSQL Ping 失败仍然只是 warning

文件：

- `apps/product-rpc/internal/svc/servicecontext.go`
- `apps/inventory-rpc/internal/svc/servicecontext.go`
- `apps/order-rpc/internal/svc/servicecontext.go`

问题：

- 有 DSN 时已执行 `PingContext`，但失败后只是打印 `WARNING`。
- 服务仍会继续启动，后续请求才会暴露数据库不可用。
- 这仍不满足上一轮 review 中“失败要明确返回/退出”的要求。

建议：

- 有 DSN 时，`PingContext` 失败应直接 `log.Fatalf` 或让初始化函数返回 error 并由 main 退出。
- 当前项目结构下，最小修复可以用 `log.Fatalf("failed to ping database: %v", err)`。
- fake repository 仍通过空 DSN 或测试构造函数使用，不要引入 SQLite。

复查命令：

```bash
make fmt
make test
git diff --check
make gen
```

复查结果：

```text
make fmt: passed
make test: passed
git diff --check: passed
make gen: passed
```

未执行数据库端到端验证：

- 当前环境没有 `psql`、`createdb`、`pg_isready` 命令。
- 本阶段不强制 PostgreSQL 集成测试。

进程和端口：

- 本次复查未启动服务。
- 已复查 `8080`、`9000`、`9001`、`9002`、`9003`，均无监听进程。

## 建议同步给 Claude Code 的二次修复 Prompt

```text
请根据 docs/reviews/phase-3-review.md 的“修复后复查”部分做 Phase 3 二次修复。

只修一个问题：
- product-rpc、inventory-rpc、order-rpc 在配置了 Dsn 时，如果 db.PingContext 失败，必须明确退出服务启动流程。
- 当前可以直接使用 log.Fatalf("failed to ping database: %v", err)。

要求：
- 不进入 Phase 4。
- 不引入 Docker Compose、Kubernetes、Redis、etcd、消息队列、payment-rpc。
- 不引入 SQLite。
- 保持单元测试继续使用 fake repository/fake RPC client。
- 运行 make fmt、make test、git diff --check。
- 如运行 make gen，也应确认没有意外生成漂移。
- 如果启动服务做验证，结束后必须停止进程，并复查 8080、9000、9001、9002、9003 无监听。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```

## 二次修复后复查

复查时间：2026-05-30

结论：

- 仍需修复，同一个数据库启动失败处理问题尚未解决。

复查发现：

- `apps/product-rpc/internal/svc/servicecontext.go` 仍然在 `PingContext` 失败时打印 `WARNING` 并继续启动。
- `apps/inventory-rpc/internal/svc/servicecontext.go` 仍然在 `PingContext` 失败时打印 `WARNING` 并继续启动。
- `apps/order-rpc/internal/svc/servicecontext.go` 仍然在 `PingContext` 失败时打印 `WARNING` 并继续启动。

当前代码片段仍是：

```go
if err := db.PingContext(ctx); err != nil {
    log.Printf("WARNING: database ping failed: %v", err)
}
```

需要改为明确退出，例如：

```go
if err := db.PingContext(ctx); err != nil {
    log.Fatalf("failed to ping database: %v", err)
}
```

复查命令：

```bash
make fmt
make test
git diff --check
rg -n "WARNING: database ping failed|failed to ping database|PingContext" \
  apps/product-rpc/internal/svc \
  apps/inventory-rpc/internal/svc \
  apps/order-rpc/internal/svc
```

复查结果：

```text
make fmt: passed
make test: passed
git diff --check: passed
PingContext failure handling: not fixed
```

进程和端口：

- 本次复查未启动服务。
- 已复查 `8080`、`9000`、`9001`、`9002`、`9003`，均无监听进程。

## 建议同步给 Claude Code 的三次修复 Prompt

```text
请继续修复 docs/reviews/phase-3-review.md 中“二次修复后复查”指出的问题。

只做这个改动：
- 将 product-rpc、inventory-rpc、order-rpc 的 db.PingContext 失败处理从 log.Printf WARNING 改为明确退出。
- 可以使用：
  log.Fatalf("failed to ping database: %v", err)

必须修改这三个文件：
- apps/product-rpc/internal/svc/servicecontext.go
- apps/inventory-rpc/internal/svc/servicecontext.go
- apps/order-rpc/internal/svc/servicecontext.go

不要做其他范围：
- 不进入 Phase 4。
- 不引入 Docker Compose、Kubernetes、Redis、etd、消息队列、payment-rpc。
- 不引入 SQLite。

验证：
- 运行 make fmt。
- 运行 make test。
- 运行 git diff --check。
- 运行 rg -n "WARNING: database ping failed" apps/product-rpc/internal/svc apps/inventory-rpc/internal/svc apps/order-rpc/internal/svc，应该没有结果。
- 如果启动服务做验证，结束后必须停止进程，并复查 8080、9000、9001、9002、9003 无监听。
```

## 三次修复后复查

复查时间：2026-05-30

结论：

- 修复后已通过，可以进入 Phase 4。

已确认：

- `apps/product-rpc/internal/svc/servicecontext.go` 中 `db.PingContext` 失败会 `log.Fatalf`。
- `apps/inventory-rpc/internal/svc/servicecontext.go` 中 `db.PingContext` 失败会 `log.Fatalf`。
- `apps/order-rpc/internal/svc/servicecontext.go` 中 `db.PingContext` 失败会 `log.Fatalf`。
- 已无 `WARNING: database ping failed`。

复查命令：

```bash
make fmt
make test
git diff --check
make gen
rg -n "WARNING: database ping failed" \
  apps/product-rpc/internal/svc \
  apps/inventory-rpc/internal/svc \
  apps/order-rpc/internal/svc
```

复查结果：

```text
make fmt: passed
make test: passed
git diff --check: passed
make gen: passed
WARNING search: no matches
```

进程和端口：

- 本次复查未启动服务。
- 已复查 `8080`、`9000`、`9001`、`9002`、`9003`，均无监听进程。
