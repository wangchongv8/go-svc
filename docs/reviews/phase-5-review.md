# Phase 5 Review

Review 时间：2026-05-31

## 结论

需要修改后再通过 Phase 5。

Phase 5 主体方向正确：

- 已新增 `deploy/k8s/` Kubernetes manifests。
- 已覆盖 Namespace、PostgreSQL、db-migrate Job、5 个业务服务、Ingress。
- 已新增 K8s 专用配置，服务间地址使用 Kubernetes Service 名称。
- 已新增 `k8s-build`、`k8s-push`、`k8s-kind-load`、`k8s-up`、`e2e-k8s` 等 Makefile 目标。
- 已保留 Kubernetes Service + DNS 的服务发现策略，没有引入 etcd/Redis/Helm/可观测性。

但当前还有两个会影响 Phase 5 核心验收的问题：

- 远端 Kubernetes 部署不会拉取 GHCR 镜像。
- `e2e-k8s` 失败时可能遗留 `kubectl port-forward` 进程。

## 高优先级问题

### 1. 远端部署路径推送了 GHCR 镜像，但 K8s YAML 仍使用本地短镜像名

文件：

- `Makefile:68`
- `Makefile:69`
- `Makefile:75`
- `Makefile:77`
- `deploy/k8s/gateway-api.yaml:51`
- `deploy/k8s/user-rpc.yaml:42`
- `deploy/k8s/product-rpc.yaml:43`
- `deploy/k8s/inventory-rpc.yaml:43`
- `deploy/k8s/order-rpc.yaml:47`
- `deploy/k8s/README.md:71`

问题：

- Phase 5 设计要求支持远端 Kubernetes：开发机 build/push 到 `ghcr.io/wangchongv8`，目标机器从 GHCR 拉镜像。
- `make k8s-push` 会把 `go-svc-<service>:phase5` tag 成 `ghcr.io/wangchongv8/go-svc-<service>:phase5` 并推送。
- 但所有 Deployment YAML 仍写的是短镜像名，例如 `go-svc-gateway-api:phase5`。
- 目标机器执行 `make k8s-up` 时会尝试拉 `docker.io/library/go-svc-gateway-api:phase5` 或查找节点本地镜像，而不是拉 GHCR。

影响：

- 本地 kind 模式可能通过，因为 `kind load docker-image go-svc-<service>:phase5` 加载的是同一个短镜像名。
- 远端机器部署会出现 `ErrImagePull` / `ImagePullBackOff`，这正好破坏了本阶段新增的远端部署目标。

建议：

- 让 K8s YAML 默认使用 GHCR 镜像名，例如 `ghcr.io/wangchongv8/go-svc-gateway-api:phase5`。
- 或提供明确的镜像替换机制，`make k8s-up` 应基于 `IMAGE_REGISTRY` / `IMAGE_TAG` 渲染后再 apply。
- `k8s-build` 最好直接构建 GHCR tag，同时可选打本地短 tag 供 kind 使用。
- `k8s-kind-load` 应加载与 YAML 一致的镜像名，避免本地和远端两套名字漂移。

### 2. `e2e-k8s` 没有 trap，失败时会遗留 port-forward 进程

文件：

- `scripts/e2e-k8s.sh:2`
- `scripts/e2e-k8s.sh:34`
- `scripts/e2e-k8s.sh:35`
- `scripts/e2e-k8s.sh:40`
- `scripts/e2e-k8s.sh:52`

问题：

- 脚本使用 `set -euo pipefail`。
- `kubectl port-forward` 后只在脚本正常跑到末尾时执行 `kill $PF_PID`。
- 如果任意 `curl` 失败、断言失败前的命令异常、或 port-forward 未准备好，脚本会提前退出，`kill` 不会执行。
- 启动 port-forward 后只 `sleep 2`，没有等待端口真正 ready，也没有检查 port-forward 进程是否已经退出。

影响：

- 失败场景会遗留监听 `8080` 的 port-forward 进程。
- 下一次验证可能因为端口占用失败。
- 这违反了我们已经约定的“验证后必须清理进程和端口”。

建议：

- 启动 port-forward 后立刻注册 `trap`：

```bash
cleanup() {
  if [ -n "${PF_PID:-}" ]; then
    kill "$PF_PID" 2>/dev/null || true
    wait "$PF_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT
```

- 使用循环等待 `http://localhost:8080/healthz` ready，而不是固定 `sleep 2`。
- 如果支持 `BASE_URL`，则当用户传入非默认 `BASE_URL` 时可以不启动 port-forward。
- 验证负向用例时应明确要求 HTTP 非 500，而不是只检查响应体包含 `error`。

## 中优先级问题

### 1. `make k8s-up` 没有应用 `ingress.yaml`

文件：

- `Makefile:89`
- `Makefile:99`
- `deploy/k8s/ingress.yaml:1`
- `deploy/k8s/README.md:62`

问题：

- Phase 5 要求覆盖 Ingress。
- 仓库新增了 `deploy/k8s/ingress.yaml`。
- README 资源清单也列出了 Ingress。
- 但 `make k8s-up` 没有 apply `deploy/k8s/ingress.yaml`。

影响：

- 执行 `make k8s-up` 后，实际集群里不会有 Ingress 资源。
- README 中“部署全部 K8s 资源”的描述和实际行为不一致。

建议：

- 在 `k8s-up` 中 apply `deploy/k8s/ingress.yaml`。
- 如果担心用户没有 Ingress Controller，也可以照常 apply Ingress，并在 README 中说明没有 Controller 时不会产生外部入口，主验证仍使用 port-forward。

### 2. Phase 5 文档与命令表对远端部署说明还不完整

文件：

- `README.md:37`
- `README.md:208`
- `README.md:258`
- `deploy/k8s/README.md:3`
- `deploy/k8s/README.md:71`

问题：

- README 阶段记录已经标记 Phase 5 完成，但当前 review 尚未通过。
- README 命令表缺少 `k8s-push`、`k8s-logs`、`k8s-port-forward`。
- `deploy/k8s/README.md` 开头仍以 kind 本地集群为主，远端部署只给了最短命令，没有说明 GHCR 镜像名、public/private、`imagePullSecret`、目标机器需要能拉 GHCR。

建议：

- 修复 review 问题前，把 Phase 5 状态改为 `review 修复中` 或类似状态。
- README 命令表补齐 Phase 5 新命令。
- `deploy/k8s/README.md` 明确区分：
  - 开发机：`docker login ghcr.io`、`make k8s-build`、`make k8s-push`
  - 目标机器：`kubectl apply --dry-run=client -f deploy/k8s/`、`make k8s-up`、`make e2e-k8s`、`make k8s-down`
  - private GHCR 镜像：如何创建 `imagePullSecret`。

### 3. `scripts/e2e-compose.sh` 仍不支持 `BASE_URL`

文件：

- `scripts/e2e-compose.sh:4`
- `docs/claude-implementation-plan.md:566`

问题：

- Phase 5 设计建议 `e2e-compose` 支持 `BASE_URL`，方便复用同一套断言。
- 当前仍写死 `BASE="http://localhost:8080"`。

建议：

- 改为 `BASE="${BASE_URL:-http://localhost:8080}"`。
- `e2e-k8s` 可以设置 `BASE_URL` 后复用同一套脚本，或至少保持两个脚本行为一致。

## 低优先级问题

### 1. SQL 被复制到 K8s ConfigMap，后续容易和源 SQL 漂移

文件：

- `deploy/k8s/db-migrate-job.yaml:8`
- `deploy/sql/001_phase3_schema.sql:1`

问题：

- Phase 5 允许 SQL 复制到 ConfigMap，但要求 README 说明它来自 `deploy/sql/001_phase3_schema.sql`。
- 当前 README 说明不够明确。

建议：

- 在 `deploy/k8s/README.md` 中说明：`db-migrate-job.yaml` 内嵌 SQL 源自 `deploy/sql/001_phase3_schema.sql`，修改 schema 时需要同步。
- 或新增生成/校验脚本，避免人工复制漂移。

## 验证结果

已执行并通过：

```bash
make fmt
make test
git diff --check
make gen
docker compose -f deploy/docker-compose/docker-compose.yml config
kubectl version --client
kind version
```

未执行通过：

```bash
kubectl apply --dry-run=client -f deploy/k8s/
kubectl apply --dry-run=client --validate=false -f deploy/k8s/
```

原因：

- 当前 kube context 指向 `localhost:8080`。
- 本机没有可用 Kubernetes API Server。
- kubectl 尝试访问 `http://localhost:8080/api` / `openapi/v2`，连接被当前环境拒绝。
- 因此本轮只能做文件级 review，未能完成 Kubernetes dry-run。

未执行：

- `make k8s-build`
- `make k8s-push`
- `make k8s-kind-load`
- `make k8s-up`
- `make e2e-k8s`
- `make k8s-down`

原因：

- 本轮 review 已发现无需启动集群即可确认的阻断问题。
- 避免在未修复 `e2e-k8s` 清理逻辑前启动 port-forward。

进程和端口：

- 本次 review 未启动业务服务、Docker Compose、K8s Pod 或 port-forward。
- 已复查 `8080`、`9000`、`9001`、`9002`、`9003`、`5432`，均无监听进程。

## 建议同步给 Claude Code 的修复 Prompt

```text
请根据 docs/reviews/phase-5-review.md 修复 Phase 5 review 问题。

要求：
- 不进入 Phase 6。
- 保持 Kubernetes Service + DNS 服务发现策略，不引入 etcd、Redis、Helm、可观测性或新业务功能。
- 修复远端部署镜像问题：
  - K8s Deployment 必须能在目标机器上从 GHCR 拉镜像。
  - 默认镜像仓库为 ghcr.io/wangchongv8，默认 tag 为 phase5。
  - k8s-build、k8s-push、k8s-kind-load、k8s-up 的镜像名必须一致，不能出现 push 到 GHCR 但 YAML 使用本地短镜像名的情况。
  - 可以让 YAML 默认写 GHCR 镜像名，也可以实现简单渲染机制，但 README 必须写清楚。
- 修复 scripts/e2e-k8s.sh：
  - 必须用 trap 清理 kubectl port-forward。
  - 失败、Ctrl-C、curl 失败、断言失败时都不能遗留 8080 监听。
  - 用循环等待 healthz ready，不要只 sleep 固定时间。
  - 负向库存不足验证应确认 HTTP 非 500。
  - 支持 BASE_URL 更好；如果传入 BASE_URL，可以不启动 port-forward。
- make k8s-up 应 apply deploy/k8s/ingress.yaml，或文档和命令名必须明确说明默认不 apply Ingress。建议直接 apply。
- README 命令表补齐 k8s-push、k8s-logs、k8s-port-forward。
- deploy/k8s/README.md 补充 GHCR public/private、imagePullSecret、远端机器部署流程、内嵌 SQL 与 deploy/sql/001_phase3_schema.sql 的同步关系。
- Phase 5 修复完成前不要在 README 标记为最终完成；修复后可恢复完成状态。
- 运行 make fmt、make test、git diff --check、make gen。
- 如果 kubectl 可用，运行 kubectl apply --dry-run=client -f deploy/k8s/；如果没有可用集群导致失败，说明原因。
- 如果 Docker/GHCR 登录可用，运行 make k8s-build、make k8s-push；否则说明未运行原因。
- 如果 Kubernetes/kind 可用，运行 make k8s-up、make k8s-ps、make e2e-k8s、make k8s-down；否则说明未运行原因。
- 验证结束后必须清理 Kubernetes 资源和 port-forward 进程，并复查 8080、9000、9001、9002、9003、5432 无异常监听。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```

## 修复后复查

复查时间：2026-05-31

结论：

- Phase 5 可以通过。

已确认修复：

- K8s Deployment 镜像已改为 `ghcr.io/wangchongv8/go-svc-<service>:phase5`。
- `make k8s-build`、`make k8s-push`、`make k8s-kind-load` 已统一使用 `IMAGE_REGISTRY` / `IMAGE_TAG`。
- `make k8s-up` 已 apply `deploy/k8s/ingress.yaml`。
- `scripts/e2e-k8s.sh` 已增加 `trap cleanup EXIT`，失败路径会清理 port-forward。
- `scripts/e2e-k8s.sh` 已支持 `BASE_URL`，并改为循环等待 `healthz` ready。
- 库存不足负向验证已改为检查 HTTP 非 `500` / 非 `000`。
- `scripts/e2e-compose.sh` 已支持 `BASE_URL`。
- README 命令表已补充 `k8s-push`、`k8s-logs`、`k8s-port-forward`。
- `deploy/k8s/README.md` 已补充 GHCR public/private、远端部署流程和 SQL 同步说明。

剩余非阻断建议：

- `deploy/k8s/README.md` 中写到 private 镜像时“各 service yaml 中取消注释对应行即可”，但当前 YAML 没有预留 `imagePullSecrets` 注释块。建议后续补一个示例片段，或在各 Deployment 中预留注释，避免用户按文档找不到位置。

复查验证：

已执行并通过：

```bash
make fmt
make test
git diff --check
bash -n scripts/e2e-compose.sh scripts/e2e-k8s.sh
make gen
docker compose -f deploy/docker-compose/docker-compose.yml config
kubectl version --client
kind version
ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_stream(File.read(f)); puts "ok #{f}" }' deploy/k8s/*.yaml
```

未执行通过：

```bash
kubectl apply --dry-run=client -f deploy/k8s/
kubectl create --dry-run=client --validate=false -f deploy/k8s/
```

原因：

- 当前 kube context 指向 `localhost:8080`。
- 本机没有可用 Kubernetes API Server。
- kubectl 尝试访问 `http://localhost:8080/api` / `openapi/v2`，连接被当前环境拒绝。

未执行：

- `make k8s-build`
- `make k8s-push`
- `make k8s-kind-load`
- `make k8s-up`
- `make e2e-k8s`
- `make k8s-down`

原因：

- 本机没有可用 Kubernetes API Server。
- 未确认本机 GHCR 登录状态，不在 review 中推送镜像。

进程和端口：

- 本次复查未启动业务服务、Docker Compose、K8s Pod 或 port-forward。
- 已复查 `8080`、`9000`、`9001`、`9002`、`9003`、`5432`，均无监听进程。
