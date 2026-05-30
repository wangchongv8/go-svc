# 项目协作流程

本文档定义 Codex、Claude Code 和用户之间的协作方式。目标是让这个学习型微服务项目既能逐步落地，又能保留足够清晰的设计脉络，方便复盘和扩展。

## 角色分工

### 用户

- 提出学习目标和业务偏好。
- 决定技术取舍，例如是否使用 go-zero、是否加入 Kubernetes、是否引入消息队列。
- 把 Codex 产出的实施文档交给 Claude Code。
- 把 Claude Code 的实现结果交回 Codex review。

### Codex

- 帮助澄清需求和学习目标。
- 产出架构设计、实施计划、验收标准和 review 清单。
- 在 Claude Code 实现后进行工程 review。
- 关注架构一致性、可运行性、学习价值、错误处理、部署体验和文档完整性。

### Claude Code

- 严格按实施计划小步实现。
- 每个阶段尽量保持可运行、可验证。
- 不随意扩大技术范围。
- 实现后更新 README、运行命令、配置说明和阶段状态。

## 推荐循环

每一轮协作都按下面节奏推进：

1. Codex 和用户讨论需求。
2. Codex 更新设计文档和实施计划。
3. 用户把对应阶段的 Claude Code prompt 交给 Claude Code。
4. Claude Code 实现并提交变更。
5. Codex review 代码和文档。
6. 用户确认是否进入下一阶段。

## 阶段边界

为了避免一次实现过大，项目按阶段交付：

- Phase 0: 项目脚手架和文档体系。
- Phase 1: 单体可运行的最小 Go 服务和基础工程命令。
- Phase 2: 引入 go-zero，拆分 API 服务和 RPC 服务。
- Phase 3: 引入数据库、Redis、etcd 和服务发现。
- Phase 4: 使用 Docker Compose 启动本地微服务集群。
- Phase 5: 增加 Kubernetes manifests。
- Phase 6: 增加日志、指标、链路追踪和压测入口。
- Phase 7: 增加 CI、测试覆盖和故障演练。

具体阶段可以根据学习重点裁剪。

## Claude Code 执行约束

交给 Claude Code 的任务应该满足：

- 一次只做一个 phase 或一个明确子任务。
- 明确列出允许修改的文件范围。
- 明确列出验收命令。
- 明确要求更新文档。
- 明确禁止引入未讨论的新框架。

推荐 prompt 格式：

```text
请在当前仓库实现 docs/claude-implementation-plan.md 中的 Phase X。

要求：
- 严格遵循 docs/architecture.md。
- 不实现 Phase X 之外的内容。
- 实现后更新 README 和相关 docs。
- 给出你运行过的验证命令和结果。
- 如果发现设计文档有冲突，先说明问题，不要自行大改架构。
```

## Codex Review 输入

Claude Code 实现后，建议提供：

- 变更摘要。
- `git diff` 或提交记录。
- 运行过的命令和结果。
- 未完成事项。
- 遇到的设计冲突。

Codex review 时优先检查：

- 是否符合设计文档。
- 是否能按 README 一键运行。
- 是否有过度实现或偏离范围。
- 是否有明显的稳定性、安全性或部署问题。

