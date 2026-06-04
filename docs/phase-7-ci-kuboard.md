# Phase 7: GitHub Actions CI and Kuboard Release Management

## Goal

Phase 7 introduces CI and a lightweight remote release workflow:

- GitHub Actions runs static checks, Go tests, and manifest validation.
- GitHub Actions builds service images and pushes them to GHCR.
- The remote Kubernetes machine pulls GHCR images.
- Kuboard is installed on the remote machine for visual release management and troubleshooting.

The first version should be semi-automatic:

```text
git push
  -> GitHub Actions validates code
  -> GitHub Actions builds and pushes images
  -> remote Kubernetes pulls images
  -> Kuboard is used to inspect/update/rollback workloads
```

Do not implement direct GitHub Actions deployment to the remote machine in this phase. Keep kubeconfig, SSH keys, and cluster-admin credentials out of GitHub Actions for now.

## Design Decisions

1. **CI and release management stay separate.**
   GitHub Actions produces verified artifacts. Kuboard manages the remote cluster.

2. **Use GHCR as the image registry.**
   Image names keep the existing convention:

   ```text
   ghcr.io/wangchongv8/go-svc-user-rpc:<tag>
   ghcr.io/wangchongv8/go-svc-product-rpc:<tag>
   ghcr.io/wangchongv8/go-svc-inventory-rpc:<tag>
   ghcr.io/wangchongv8/go-svc-order-rpc:<tag>
   ghcr.io/wangchongv8/go-svc-gateway-api:<tag>
   ```

3. **Use immutable image tags for real releases.**
   The primary tag should be the Git commit SHA. A moving `main` tag can be pushed for convenience, but Kubernetes releases should prefer the SHA tag.

4. **Do not introduce Helm, Argo CD, Flux, or GitOps yet.**
   Those can be later phases. This phase focuses on understanding CI, image artifacts, Kubernetes rollout, and Kuboard.

5. **Kuboard runs outside the cluster.**
   Kuboard v3 is recommended as a standalone Docker container on the remote machine, then it imports/manages the Kubernetes cluster.

## GitHub Actions Workflows

Create:

```text
.github/workflows/ci.yml
.github/workflows/images.yml
```

### ci.yml

Runs on:

- `push`
- `pull_request`

Required checks:

```bash
go fmt ./...
go test ./...
git diff --check
bash -n scripts/*.sh
docker compose -f deploy/docker-compose/docker-compose.yml config
```

Kubernetes manifests should also be checked without requiring a live cluster. Use a lightweight YAML parser or schema-aware tool if available. A Ruby YAML parse is acceptable for this learning phase:

```bash
ruby -e 'require "yaml"; Dir["deploy/k8s/*.yaml", "apps/*/etc/*.yaml"].each { |f| YAML.load_stream(File.read(f)); puts f }'
```

Do not run `make compose-up`, `make k8s-up`, or e2e tests in GitHub-hosted CI in this phase.

### images.yml

Runs on:

- `push` to `main`
- `workflow_dispatch`

Permissions:

```yaml
permissions:
  contents: read
  packages: write
```

Build matrix:

| service | main | conf_dir | image |
|---|---|---|---|
| user-rpc | `apps/user-rpc/user.go` | `apps/user-rpc/etc` | `go-svc-user-rpc` |
| product-rpc | `apps/product-rpc/product.go` | `apps/product-rpc/etc` | `go-svc-product-rpc` |
| inventory-rpc | `apps/inventory-rpc/inventory.go` | `apps/inventory-rpc/etc` | `go-svc-inventory-rpc` |
| order-rpc | `apps/order-rpc/order.go` | `apps/order-rpc/etc` | `go-svc-order-rpc` |
| gateway-api | `apps/gateway-api/gateway.go` | `apps/gateway-api/etc` | `go-svc-gateway-api` |

Each image should be pushed with:

```text
ghcr.io/wangchongv8/<image>:<git-sha>
ghcr.io/wangchongv8/<image>:main
```

Use Docker official GitHub Actions:

- `docker/login-action`
- `docker/setup-buildx-action`
- `docker/build-push-action`

Use `GITHUB_TOKEN` for GHCR login. No personal access token should be needed for publishing packages from this repository.

## Makefile Changes

Add or update:

```makefile
IMAGE_TAG ?= local

ci-check:
	go fmt ./...
	go test ./...
	git diff --check
	bash -n scripts/*.sh
	docker compose -f deploy/docker-compose/docker-compose.yml config
	ruby -e 'require "yaml"; Dir["deploy/k8s/*.yaml", "apps/*/etc/*.yaml"].each { |f| YAML.load_stream(File.read(f)); puts f }'

k8s-set-images:
	kubectl set image deployment/user-rpc user-rpc=$(IMAGE_REGISTRY)/go-svc-user-rpc:$(IMAGE_TAG) -n $(K8S_NAMESPACE)
	kubectl set image deployment/product-rpc product-rpc=$(IMAGE_REGISTRY)/go-svc-product-rpc:$(IMAGE_TAG) -n $(K8S_NAMESPACE)
	kubectl set image deployment/inventory-rpc inventory-rpc=$(IMAGE_REGISTRY)/go-svc-inventory-rpc:$(IMAGE_TAG) -n $(K8S_NAMESPACE)
	kubectl set image deployment/order-rpc order-rpc=$(IMAGE_REGISTRY)/go-svc-order-rpc:$(IMAGE_TAG) -n $(K8S_NAMESPACE)
	kubectl set image deployment/gateway-api gateway-api=$(IMAGE_REGISTRY)/go-svc-gateway-api:$(IMAGE_TAG) -n $(K8S_NAMESPACE)

k8s-rollout-status:
	@for app in user-rpc product-rpc inventory-rpc order-rpc gateway-api; do \
		kubectl rollout status deployment/$$app -n $(K8S_NAMESPACE) --timeout=120s; \
	done
```

`k8s-set-images` is optional but useful. It lets the remote machine update images to a specific CI-produced tag without editing YAML by hand.

## Kubernetes Manifest Strategy

Current Kubernetes YAML files hardcode image tags such as `phase6`. For Phase 7, keep the YAML simple and do not introduce Helm.

Recommended approach:

- Keep a stable default tag in YAML for local readability.
- Use `kubectl set image` or Kuboard UI to update the image tag on the remote cluster.
- Document the exact SHA tag deployed.

Later, if tag drift becomes annoying, add Kustomize overlays or a tiny render script.

## GHCR and Image Pull

If GHCR packages are public, Kubernetes can pull without an image pull secret.

If GHCR packages are private, create an image pull secret on the remote cluster:

```bash
kubectl create namespace go-svc --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=wangchongv8 \
  --docker-password=<github-token> \
  -n go-svc

kubectl patch serviceaccount default \
  -n go-svc \
  -p '{"imagePullSecrets":[{"name":"ghcr-secret"}]}'
```

Use a GitHub token with package read permission. Do not commit tokens.

## Remote Machine Workflow

On the remote machine:

```bash
git pull
kubectl apply -f deploy/k8s/namespace.yaml
make k8s-up
make e2e-k8s
make verify-observability-k8s
```

For a newly built image tag:

```bash
IMAGE_TAG=<git-sha> make k8s-set-images
make k8s-rollout-status
make e2e-k8s
make verify-observability-k8s
```

Kuboard can do the same image update visually:

```text
Kuboard
  -> cluster
  -> namespace go-svc
  -> Deployment
  -> edit image tag
  -> save
  -> watch rollout / pod events / logs
```

## Kuboard Installation Plan

Use Kuboard v3 with built-in user storage for the learning environment.

Remote machine requirements:

- Docker running.
- The remote Kubernetes cluster is reachable from the remote machine.
- Ports for Kuboard are available and firewalled appropriately.
- Kuboard data directory is persistent.

Recommended deployment shape:

```text
remote machine
  -> Docker
     -> Kuboard container
  -> Kubernetes cluster
     -> go-svc namespace
```

Kuboard setup steps:

1. Install Kuboard v3 with Docker according to the official Kuboard v3 built-in-user installation guide.
2. Persist Kuboard data to a host directory.
3. Access Kuboard from your browser.
4. Import the local Kubernetes cluster using kubeconfig or Kuboard agent.
5. Open namespace `go-svc`.
6. Verify that Deployments, Services, Pods, ConfigMaps, Secrets, and Events are visible.

For learning, keep authentication simple:

- Start with built-in users.
- Later, optionally configure GitHub OAuth/OIDC or LDAP/SSO.

Security notes:

- Do not expose Kuboard directly to the public internet without firewall/VPN/auth hardening.
- Do not put kubeconfig into GitHub.
- Do not store GHCR tokens in repository files.

## Recommended Human Workflow

### Normal Development

```text
local code change
  -> make ci-check
  -> git commit
  -> git push
  -> GitHub Actions CI passes
```

### Build Images

```text
merge/push to main
  -> images workflow builds 5 images
  -> images pushed to GHCR with git SHA tag
```

### Remote Release

```text
remote machine / Kuboard
  -> choose SHA tag
  -> update 5 Deployments
  -> watch rollout
  -> run e2e and observability checks
```

### Rollback

```text
Kuboard
  -> Deployment
  -> Rollback / edit image tag to previous SHA
  -> watch rollout
```

Or:

```bash
kubectl rollout undo deployment/gateway-api -n go-svc
```

## Acceptance Criteria

- `.github/workflows/ci.yml` exists and runs checks on push and pull request.
- `.github/workflows/images.yml` exists and can build/push all 5 service images to GHCR.
- `make ci-check` exists and matches the CI checks as closely as practical.
- `IMAGE_TAG` default is no longer phase-specific.
- Docs explain how to:
  - view GitHub Actions results,
  - find GHCR image tags,
  - create `imagePullSecret` if GHCR is private,
  - install Kuboard,
  - import the remote cluster,
  - update image tags through Kuboard,
  - verify rollout with CLI commands.
- Existing Phase 4/5/6 workflows do not regress.

## Out of Scope

- Automatic deployment from GitHub Actions to the remote cluster.
- Storing kubeconfig, SSH private keys, or cluster-admin credentials in GitHub.
- Helm/Kustomize/Argo CD/Flux.
- Production-grade RBAC, network policy, TLS, and SSO hardening.

## Claude Code Prompt

```text
请在当前仓库实现 docs/phase-7-ci-kuboard.md。

要求：
- 新增 GitHub Actions workflow：.github/workflows/ci.yml 和 .github/workflows/images.yml。
- 新增或更新 Makefile：ci-check、k8s-set-images、k8s-rollout-status，并把 IMAGE_TAG 默认值改成 local 或可覆盖值，不要继续使用 phase 固定默认值。
- images workflow 使用 GHCR：ghcr.io/wangchongv8/go-svc-<service>，推送 git sha tag 和 main tag。
- images workflow 使用 matrix 明确配置每个服务的 SERVICE_MAIN、SERVICE_CONF_DIR、image 名称。
- 不实现自动部署远端机器，不引入 SSH/kubeconfig secrets。
- 更新 README、deploy/k8s/README.md 或新增 docs，说明 GitHub Actions、GHCR、远端机器、Kuboard 的使用流程。
- 保持 Phase 4/5/6 原有命令可用。
- 实现后运行 make ci-check；如果 GitHub Actions 无法本地运行，说明需要 push 后在 GitHub UI 验证。
- 不启动长期运行的 Compose/K8s 服务；如为了验证启动了，必须清理并检查端口。
- 回复中列出修改文件、验证命令、结果和未完成事项。
```

