---
name: add-api
description: >-
  新增 API 功能的全栈 checklist（Go 后端 + React 前端）。
  在添加 REST 接口、表单/表格对接 API、或全栈 CRUD 时使用。
disable-model-invocation: true
---

# 新增 API（全栈 checklist）

自上而下：契约 → 后端 → 前端 → 联调验证。

## 1. 设计契约

- [ ] URL 在 `/api/...` 下（见 `server/config.yaml` 的 `context_path`）
- [ ] 方法 + 路径（当前项目用户 API 多为 POST，保持一致）
- [ ] 请求/响应字段命名一致（TS camelCase，Go json tag）
- [ ] 鉴权：公开 / 登录 / 管理员
- [ ] 错误码：`40000` 参数、`40001` 业务、`50000` 服务端

## 2. 后端（`server/`）

顺序参考 `.cursor/rules/server.mdc` 与 `handler.mdc`：

1. [ ] `internal/model/` — 实体 + `*Request`，json + binding tag
2. [ ] `internal/store/` — GORM 方法
3. [ ] `internal/service/` — 校验、鉴权、sentinel 业务错误
4. [ ] `internal/handler/` — 绑定 JSON、调 service、`common.Success` / `common.Error`
5. [ ] `cmd/server/main.go` — 注册路由 + 中间件（session / admin）
6. [ ] 手动测试：curl 或 `/api/health` + 新接口（带 cookie session）

```bash
cd server && go vet ./...
```

## 3. 前端（`web/`）

1. [ ] `src/types/api.ts` — 与 Go 对齐（必填/可选字段）
2. [ ] `src/api/<domain>.ts` — `request.post/get` + `unwrap()`
3. [ ] 页面放在正确目录：
   - 用户功能 → `pages/user/...`
   - 管理功能 → `pages/admin/...`
4. [ ] UI：Ant Design `Table` / `Form`；需登录/管理员时加路由守卫
5. [ ] `res.code !== 0` 时用 `message.error(res.message)`

```bash
npm run lint --prefix web
npm run build --prefix web
```

## 4. 联调验证

- [ ] `npm run dev` — 无 proxy ECONNREFUSED
- [ ] 需 Session 时走登录流程（cookie + `withCredentials`）
- [ ] 管理员接口：非 admin 应被重定向或 API 报错
- [ ] 类型与网络 payload 字段一致

## 5. 勿忘

- 仅在有新的跨端约定时更新 `.cursor/rules`（可选）
- 勿提交 `config.yaml` 或生成的 `.d.ts`
- 用户中心（`/user/center`）与用户管理（`/admin/users`）保持分离
