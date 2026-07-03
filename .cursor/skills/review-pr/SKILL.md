---
name: review-pr
description: >-
  审查本 monorepo 的 Pull Request：正确性、鉴权边界、API 契约与改动范围。
  在用户要求 review PR、看分支 diff、或评估是否可合并时使用。
disable-model-invocation: true
---

# PR 审查（ai-comic-generator）

对照 `main`（或用户指定基线）审查分支改动。优先使用 `gh pr diff` / `git diff base...HEAD`。

## 检查清单

### 范围与结构

- [ ] 改动符合 `pages/` 划分（`common` / `user` / `admin`）
- [ ] 无密钥（`server/config.yaml`、`.env`、凭证）
- [ ] 未提交生成文件（`web/src/components.d.ts`、`auto-imports.d.ts`）
- [ ] diff 尽量小，无无关重构

### 后端（`server/`）

- [ ] Handler 薄；逻辑在 service；SQL 在 store
- [ ] JSON 使用 `common.Success` / `common.Error`，HTTP 200
- [ ] 管理端接口校验角色
- [ ] 新路由已在 `cmd/server/main.go` 注册
- [ ] `go vet ./...` 通过

### 前端（`web/`）

- [ ] `types/api.ts` 与 Go model 一致
- [ ] API 使用 `request` + `unwrap`，`withCredentials: true`
- [ ] antd/图标勿手动 import（类型与 locale 除外）
- [ ] 受保护页面使用路由守卫（`RequireAuth`、`RequireAdmin`）
- [ ] 列表 UI 优先 `Table`
- [ ] `npm run lint --prefix web` 通过

### 鉴权与路由

- [ ] 个人资料 → `/user/center`（非管理列表）
- [ ] 用户管理 → `/admin/users`
- [ ] 若改 URL，保留兼容重定向

## 输出格式

```markdown
## 摘要
[1–3 句话]

## 必须修复
- [file:line] 问题 — 原因 — 建议改法

## 建议
- ...

## 测试计划
- [ ] ...
```

**必须修复** 阻塞合并；**建议** 可选。
