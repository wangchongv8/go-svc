# Phase 1 Review

Review 时间：2026-05-30

## 结论

修复后已通过，可以进入 Phase 2。

Phase 1 主体符合预期：

- 已初始化 Go module。
- 已创建最小 HTTP 服务。
- 已实现 `GET /healthz` 的健康检查能力。
- 已增加 `Makefile`。
- 已增加测试。
- 未提前引入 go-zero、gRPC、PostgreSQL、Docker 或 Kubernetes。

## 高优先级问题

无。

## 中优先级问题

### 1. 根目录存在编译产物

状态：已修复。

文件：

- `/gateway`

问题：

- `gateway` 是本机编译产物，检测结果为 `Mach-O 64-bit executable arm64`。
- 编译产物不应该保留在项目根目录，也不应该进入版本控制。

建议：

- 删除根目录下的 `gateway` 文件。
- 增加 `.gitignore`，至少忽略：

```gitignore
/gateway
bin/
*.test
```

## 低优先级问题

### 1. `/healthz` 未限制 HTTP method

状态：已修复。

文件：

- `internal/gateway/handler.go`
- `internal/gateway/handler_test.go`

问题：

- 文档和注释都说明接口是 `GET /healthz`。
- 当前实现没有限制 method，`POST /healthz` 也会返回 `200 OK`。

建议：

- 非 `GET` 请求返回 `405 Method Not Allowed`。
- 在测试中补充非 GET 请求用例。

## 验证结果

已执行并通过：

```bash
make fmt
make test
```

`make test` 结果：

```text
?   	go-svc/cmd/gateway	[no test files]
ok  	go-svc/internal/gateway
```

运行验证：

- 默认 `8080` 端口已被占用。
- 使用 `PORT=38080 make run` 启动成功。
- 使用 `curl -i http://localhost:38080/healthz` 验证成功。

响应结果：

```text
HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ok"}
```

## 修复后复查

复查时间：2026-05-30

已确认：

- 当前目录已初始化为 git 仓库。
- 根目录 `gateway` 编译产物已删除。
- 已新增 `.gitignore`，覆盖 `/gateway`、`bin/`、`*.test` 等编译产物。
- `GET /healthz` 返回 `200 OK` 和 `{"status":"ok"}`。
- 非 `GET /healthz` 返回 `405 Method Not Allowed`。
- 已补充非 GET 请求测试。

复查命令：

```bash
make fmt
make test
PORT=38081 make run
curl -i http://localhost:38081/healthz
curl -i -X POST http://localhost:38081/healthz
```

复查结果：

```text
make fmt: passed
make test: passed
GET /healthz: 200 OK
POST /healthz: 405 Method Not Allowed
```

## 建议同步给 Claude Code 的修复 Prompt

```text
请根据 docs/reviews/phase-1-review.md 修复 Phase 1 review 问题。

要求：
- 不进入 Phase 2。
- 不引入 go-zero、gRPC、PostgreSQL、Docker、Kubernetes。
- 删除根目录下的 gateway 编译产物。
- 增加 .gitignore，至少忽略 /gateway、bin/、*.test。
- 修改 /healthz handler：只允许 GET；非 GET 返回 405 Method Not Allowed。
- 为非 GET /healthz 补充测试。
- 运行 make fmt 和 make test。
- 如果当前目录还不是 git 仓库，请初始化 git 仓库，但不要创建提交，除非用户明确要求。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```
