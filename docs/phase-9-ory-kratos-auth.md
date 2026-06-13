# Phase 9: Ory Kratos Authentication

## Goal

Phase 9 introduces Ory Kratos as the identity and authentication system for this learning microservice cluster.

The goal is to move password, login, and session verification out of `user-rpc` and into a dedicated identity system:

```text
client
  -> gateway-api
     -> Kratos public API for registration, login, and session verification
     -> user-rpc for local business user profile
     -> business RPC services
```

This phase focuses on backend API clients. Do not build login pages yet.

## Branch Requirement

Claude Code must implement this phase on a new branch, not directly on `main`.

Suggested branch:

```bash
git checkout main
git pull
git checkout -b feature/phase9-ory-kratos-auth
```

If local `main` has uncommitted changes, Claude Code must stop and report them before creating the branch.

## References

- Ory Kratos quickstart: https://www.ory.com/docs/kratos/quickstart
- Ory Kratos login flow: https://www.ory.com/docs/kratos/self-service/flows/user-login
- Ory Kratos session overview: https://www.ory.com/docs/kratos/session-management/overview

Design assumptions from Ory documentation:

- Kratos is API-first and headless. The project should expose its own API through `gateway-api`.
- Kratos has public and admin APIs. Business traffic should use the public API.
- API login flows can return an Ory session token.
- `gateway-api` can verify a session by calling Kratos `/sessions/whoami`.

## Design Decisions

1. **Use Kratos for identity and authentication.**
   Kratos owns identities, password credentials, registration flows, login flows, and sessions.

2. **Keep business users in `user-rpc`.**
   `user-rpc` should keep the local business user profile and map it to Kratos identity id.

3. **Use API flows first.**
   Use Kratos API flows and return `session_token` to the client. Browser flows, cookies, CSRF behavior, and custom login pages can be a later phase.

4. **Do not introduce the rest of the Ory platform yet.**
   Do not add Hydra, Keto, Oathkeeper, OAuth/OIDC social login, MFA, email verification, or password recovery in this phase.

5. **Do not log secrets.**
   Never log `session_token`, password, `Authorization`, `Cookie`, Kratos DSN, or Kratos secrets.

6. **Do not create high-cardinality metrics labels.**
   `kratos_identity_id`, `user_id`, and `session_token` must not become Prometheus labels. `session_token` must not appear in logs either.

7. **Auth is middleware, not duplicate route registration.**
   `POST /api/v1/orders` must be registered once. Do not add a second route in `gateway.go` to wrap auth. The existing order route must run through auth middleware or an auth-aware handler path.

8. **Kratos registration must provision a local user.**
   Registration is incomplete if it only creates a Kratos identity. After Kratos returns `identity.id`, `gateway-api` must call `user-rpc` to create or fetch the local business user mapped by `kratos_identity_id`.

## Target Architecture

Registration:

```text
POST /api/v1/auth/register
  -> gateway-api initializes Kratos registration API flow
  -> gateway-api submits username/password to Kratos
  -> Kratos creates identity
  -> gateway-api creates or upserts local user profile through user-rpc
  -> gateway-api returns local user profile and optional session_token
```

Required registration behavior:

- Kratos stores identity and password credentials.
- `user-rpc` stores the business user profile.
- The response must include the real local `user.id`.
- Returning only `username` or `id: 0` is not acceptable.

Login:

```text
POST /api/v1/auth/login
  -> gateway-api initializes Kratos login API flow
  -> gateway-api submits identifier/password to Kratos
  -> Kratos returns session and session_token
  -> gateway-api maps kratos_identity_id to local user profile
  -> gateway-api returns session_token and user profile
```

Required login behavior:

- Login must resolve `identity.id` to a local business user.
- Login must not create a password-backed legacy user.
- If old data has a Kratos identity without a local profile, gateway may call `GetOrCreateByKratosIdentity` to self-heal.

Authenticated request:

```text
client sends Authorization: Bearer <session_token>
  -> gateway-api calls Kratos /sessions/whoami
  -> gateway-api extracts identity.id
  -> gateway-api maps identity.id to local user id
  -> gateway-api puts user information into request context
  -> business logic calls RPC services
```

Required middleware behavior:

- Validate `Authorization: Bearer <session_token>`.
- Call Kratos `/sessions/whoami`.
- Resolve `identity.id` to local `users.id` through `user-rpc`.
- Put local `user_id` into request context.
- Reject mapping failures instead of continuing with `user_id=0`.

Create order after this phase:

```text
POST /api/v1/orders
Authorization: Bearer <session_token>

{
  "product_id": 1,
  "quantity": 2
}
```

`user_id` must come from the authenticated context, not from the client request body.

## Scope

Must implement:

- Docker Compose Kratos deployment.
- Docker Compose Kratos migration container or one-shot migration command.
- Kubernetes Kratos ConfigMap, Secret, Service, Deployment, and migration Job.
- Kratos identity schema for password login.
- Gateway auth endpoints:
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `GET /api/v1/auth/me`
- Gateway helper or middleware for reading `Authorization: Bearer <session_token>`.
- Gateway session verification through Kratos `/sessions/whoami`.
- Gateway auth middleware that maps Kratos identity to local user id and writes local user id into request context.
- Local user profile mapping:
  - `id`
  - `kratos_identity_id`
  - `username`
  - `created_at`
- At least one authenticated business flow. Prefer `POST /api/v1/orders`.
- README and architecture docs updates.
- Compose e2e coverage for the new auth flow.

Should consider:

- Keep old `/api/v1/register` and `/api/v1/login` temporarily, but mark them as legacy.
- Add a small package such as `pkg/auth/kratos` or `pkg/identity/kratos` for the Kratos client.
- Reuse current HTTP error mapping style so Kratos failures become stable gateway errors.
- Keep trace id logging on gateway auth operations.
- Log `kratos_identity_id` and local `user_id` only after successful authentication.

Must not do:

- Do not duplicate-register `POST /api/v1/orders` in `gateway.go`.
- Do not hardcode `userID := int64(1)`.
- Do not use legacy `Register(username, "kratos-managed")` as a substitute for Kratos identity mapping.
- Do not return auth user profiles with `id: 0`.

Do not implement:

- Frontend pages.
- Browser login flow and cookie-based sessions.
- OAuth/OIDC social login.
- Ory Hydra.
- Ory Keto.
- Ory Oathkeeper.
- MFA, WebAuthn, passkeys, email verification, password recovery.
- Production-grade email delivery.
- Production-grade secret management.

## Data Model

Kratos should use PostgreSQL, but it must not pollute or break existing business tables.

Preferred learning setup:

```text
PostgreSQL
  -> go_svc database/schema for business tables
  -> kratos database/schema for Kratos tables
```

The exact split can be either:

- separate database: `go_svc` and `kratos`
- separate schemas in the same database: `public` and `kratos`

Choose the simpler option for Compose and K8s consistency.

Local business user profile should be stored outside Kratos:

```sql
users
  id BIGSERIAL PRIMARY KEY
  kratos_identity_id UUID UNIQUE NOT NULL
  username TEXT UNIQUE NOT NULL
  created_at TIMESTAMPTZ NOT NULL
```

Existing password columns should not be used for Kratos-backed users. If keeping legacy auth temporarily, document the boundary clearly.

## Kratos Configuration

Add a Kratos identity schema for the first version.

Recommended first version:

```text
identifier: username
credential: password
```

Using `username` keeps the current learning API simple. Email can be introduced later together with verification and recovery.

Kratos config should include:

- DSN from environment or Kubernetes Secret.
- Public API listen address, usually `0.0.0.0:4433`.
- Admin API listen address, usually `0.0.0.0:4434`.
- Identity schema location.
- Password strategy enabled.
- Development-friendly URLs for Compose and K8s.

Do not expose admin API through Ingress.

## Compose Requirements

Update `deploy/docker-compose/docker-compose.yml`:

- Add `kratos`.
- Add `kratos-migrate`.
- Add Kratos config and identity schema mounts.
- Use PostgreSQL as Kratos storage.
- Expose Kratos public API for local debugging if useful.
- Do not expose Kratos admin API to external clients unless explicitly documented as local-only.

Recommended config paths:

```text
deploy/kratos/kratos.yml
deploy/kratos/identity.schema.json
```

The normal project path should be:

```text
client -> gateway-api -> kratos
```

Clients should not need to call Kratos directly for ordinary project APIs.

## Kubernetes Requirements

Add Kubernetes resources under `deploy/k8s/`:

- Kratos ConfigMap.
- Kratos Secret for DSN and secrets.
- Kratos Service.
- Kratos Deployment.
- Kratos migration Job.

The Service should expose:

- public API for gateway internal access.
- admin API only inside the namespace if required for migration or debugging.

Do not expose Kratos admin API through Ingress.

`make k8s-up` should apply Kratos resources in the correct order:

```text
namespace
postgres
kratos config/secret
kratos migration job
kratos deployment/service
business migrations
business services
```

## Gateway API Requirements

Add these endpoints:

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/auth/me
```

Request shape:

```json
{
  "username": "alice",
  "password": "123456"
}
```

Login response:

```json
{
  "session_token": "...",
  "user": {
    "id": 1,
    "username": "alice"
  }
}
```

`session_token` can be returned to the client but must not be logged.

Protected requests should use:

```text
Authorization: Bearer <session_token>
```

Auth middleware must follow this chain:

```text
request
  -> AuthMiddleware
     -> read Bearer token
     -> Kratos /sessions/whoami
     -> extract identity.id
     -> user-rpc.GetOrCreateByKratosIdentity or GetByKratosIdentity
     -> context local user_id
  -> business handler
```

`gateway.go` should stay a thin bootstrap file:

```text
load config
create server
register global middleware
register routes
start
```

It should not manually re-register generated business routes with the same method/path.

## user-rpc Requirements

Refactor `user-rpc` toward profile ownership:

- Add or update profile methods for Kratos identity mapping.
- Keep local `users.id` as the business user id.
- Make `kratos_identity_id` unique.
- Do not hash or verify passwords for Kratos-backed endpoints.
- Existing legacy register/login may remain temporarily but must be marked as legacy in docs.

Recommended RPC methods:

```text
GetOrCreateByKratosIdentity(kratos_identity_id, username)
GetByKratosIdentity(kratos_identity_id)
GetUser(id)
```

Exact protobuf naming can follow existing project style.

These methods must be real RPC methods, not only in-memory helper methods:

- Update `apps/user-rpc/user.proto`.
- Regenerate protobuf, gRPC, zrpc client, and server code.
- Implement server logic using `UserStore.GetOrCreateByKratosIdentity` and `UserStore.FindByKratosIdentity`.
- Use the generated `userrpc` client from `gateway-api`.

`GetOrCreateByKratosIdentity` must be idempotent:

```text
same kratos_identity_id + same username
  -> always returns the same local user
```

## Order API Requirement

Change one meaningful business endpoint to prove authentication is real.

Preferred endpoint:

```text
POST /api/v1/orders
```

After Phase 9:

- It must require `Authorization: Bearer <session_token>`.
- It must get `user_id` from authenticated context.
- It should reject unauthenticated requests with HTTP 401.
- Request body should not accept or trust client-provided `user_id`.
- There must be exactly one `POST /api/v1/orders` route registration.
- `CreateOrderLogic` must not hardcode user id.

If changing the existing request shape is too disruptive, keep a compatibility path temporarily but document it as legacy and make the new authenticated path the primary one.

## E2E Requirements

Compose e2e must cover:

1. Register through `/api/v1/auth/register`.
2. Login through `/api/v1/auth/login`.
3. Capture `session_token`.
4. Call `/api/v1/auth/me` with token and expect 200.
5. Call protected order creation without token and expect 401.
6. Call protected order creation with token and expect success.
7. Verify no response or log exposes password or session token beyond the login response body.

Kubernetes e2e is optional for this phase if local K8s is not available, but YAML must be syntactically valid.

## Observability Requirements

Keep Phase 8 behavior:

- HTTP responses should still include `X-Trace-Id`.
- Auth logs should include `trace_id` where available.
- Logs may include `kratos_identity_id` and local `user_id` after authentication.
- Logs must not include password, session token, cookie, or authorization header.
- Do not add auth identifiers as Prometheus labels.

Recommended logs:

```text
auth register success
auth login success
auth whoami success
auth unauthorized
```

## Validation

Required local checks:

```bash
make fmt
go test ./...
go vet ./...
docker compose -f deploy/docker-compose/docker-compose.yml config -q
bash -n scripts/*.sh
git diff --check
```

If Docker is available:

```bash
make compose-up
make e2e-compose
make compose-down
```

If Kubernetes is available:

```bash
kubectl apply --dry-run=client -f deploy/k8s/
IMAGE_TAG=<tag> make k8s-up
make e2e-k8s
make k8s-down
```

Validation cleanup rule:

- Stop Compose/K8s resources after runtime verification.
- Stop any port-forward process started for verification.
- Recheck relevant ports if long-running services were started.

## Claude Code Prompt

```text
请在当前仓库新建分支 feature/phase9-ory-kratos-auth，实现 Phase 9：接入 Ory Kratos 作为注册、登录和 session 验证系统。

要求：
- 严格遵循 docs/phase-9-ory-kratos-auth.md。
- 不引入前端页面、OAuth、MFA、邮箱验证、找回密码、Hydra、Keto、Oathkeeper。
- Kratos 使用 PostgreSQL，且不能破坏现有业务表。
- gateway-api 新增 /api/v1/auth/register、/api/v1/auth/login、/api/v1/auth/me。
- 注册/登录通过 Kratos API flow 完成。
- 登录成功返回 session_token，但 session_token 不能写入日志。
- gateway-api 通过 Kratos /sessions/whoami 校验 Authorization: Bearer <session_token>。
- user-rpc 不再负责新 auth 接口的密码校验，只维护本地业务用户 profile 和 kratos_identity_id 映射。
- 至少让创建订单支持认证上下文，不再信任客户端传 user_id。
- Compose e2e 跑通 register、login、me、未认证 401、带 token 创建订单。
- K8s YAML 至少通过 dry-run 或说明无法执行原因。
- 更新 README、docs/architecture.md、docs/claude-implementation-plan.md 和相关部署文档。
- 实现后运行可用的格式化、测试和验证命令。
- 验证结束后清理 Compose/K8s/port-forward 进程。

当前实现如存在以下问题，必须修掉：
- /api/v1/auth/register 只创建 Kratos identity，但没有调用 user-rpc 创建本地 user profile。
- /api/v1/auth/login 和 /api/v1/auth/me 没有返回真实 local user id。
- /api/v1/orders 通过 gateway.go 重复注册路由来包 auth。
- auth middleware 使用 legacy Register(username, "kratos-managed") 代替 kratos_identity_id 映射。
- CreateOrderLogic 硬编码 userID := int64(1)。

正确流程：
- register: Kratos 创建 identity -> gateway 调 user-rpc.GetOrCreateByKratosIdentity -> 返回 session_token + local user。
- login: Kratos 登录 -> gateway 调 user-rpc.GetOrCreateByKratosIdentity 或 GetByKratosIdentity -> 返回 session_token + local user。
- me: Kratos whoami -> gateway 调 user-rpc.GetByKratosIdentity -> 返回 local user。
- order: AuthMiddleware -> context local user_id -> CreateOrderLogic 使用 context user_id。

回复中列出：
- 修改文件
- 验证命令和结果
- 未完成事项或明确 deferred 的内容
```

## Review Focus

Codex review should focus on:

- Kratos API flow是否按官方模型实现，没有硬拼内部数据库或绕过 flow。
- session token是否只在响应中返回，不进入日志、metrics、trace attributes。
- gateway是否通过 `/sessions/whoami` 做鉴权。
- 业务接口是否不再信任客户端传入的 `user_id`。
- 本地 user profile和 Kratos identity映射是否唯一且幂等。
- Compose/K8s Kratos配置是否一致。
- Kratos admin API是否没有暴露给外部入口。
- 旧 register/login 与新 auth接口边界是否清楚。
- e2e是否覆盖 401 和带 token 成功路径。
