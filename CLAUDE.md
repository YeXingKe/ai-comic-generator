# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概览

**ai-comic-generator** — AI 漫画创作平台，前后端同仓：`web/`（React 19 + Vite + Ant Design 6）+ `server/`（Go 1.24 + Gin + GORM + MySQL + Redis Session）。核心是一条把「主题」变成公众号漫画文章的多智能体流水线。

补充文档：`AGENTS.md`（面向 AI/Agent 的总览）、`.cursor/rules/*.mdc`（分端约定，`project.mdc` 始终生效）。

## 常用命令（仓库根目录）

```bash
npm run dev              # 同时启动 web(:5173) + server
npm run web              # 仅前端
npm run server           # 仅后端（cd server && go run ./cmd/server/main.go）
npm run build:web        # vite build && tsc -b，产物 web/dist/
npm run build:server     # 产物 bin/server

npm run lint --prefix web   # oxlint（前端唯一 lint）
cd server && go vet ./...   # 后端静态检查（无 golangci-lint）
```

后端无测试文件；前端无测试框架。改动后至少跑 `go vet ./...` / `go build` 与 `npm run build --prefix web`（build 内含 `tsc -b` 类型检查）。

首次搭建：`cp server/config.yaml.example server/config.yaml`，建 MySQL 库 `ai_comic_generator`，起 Redis。默认管理员 `admin` / `admin123456`（`UserService.EnsureAdmin` 自动创建）。

## 后端架构（server/）

模块路径 `github.com/ai-comic-generator/server/...`，严格分层，依赖在 `internal/app/app.go` 统一组装注入：

```
handler/   HTTP 薄适配层：绑定 JSON → 调 service → 返回 common 响应，不写 SQL
service/   业务逻辑（校验、编排）
store/     GORM 数据访问，JSON 列在此序列化/反序列化
model/     实体 + 请求/响应 DTO + 状态常量
agent/     漫画流水线单步智能体（LLM 步骤）
client/    外部服务：hunyuan（腾讯混元生图）、wechat（公众号）
```

新增接口顺序：`model/` → `store/` → `service/` → `handler/` → `cmd/server/main.go` 注册路由与中间件 → 前端同步 `web/src/types/api.ts` + `web/src/api/`。

### 响应与鉴权约定

- JSON 统一 `{ code, data?, message }`，一律 HTTP 200，用 `common.Success` / `common.Error`；`code === 0` 为成功（与前端 `BaseResponse` 一致）。
- Session 存 Redis + Cookie，中间件 `middleware.AuthCheck(userService, role)`；管理端传 `common.AdminRole`。角色校验在 handler/service 做，勿信客户端传来的 role。`50000` 响应不得泄露内部错误细节。

### 漫画流水线（核心）

`ComicService` → `ComicOrchestrator`，分两阶段、都以 `go func()` 异步跑，状态机落库到 `comic` 表：

1. `RunTitles` — 标题推荐（`TitleAgent`），完成后置 `AWAITING_CONFIRM` / `TITLE_SELECTING`，等用户 `POST /comic/confirm-title`。
2. `RunFromStory`（用户 `POST /comic/start` 触发）— 六步顺序执行：故事构思 → 角色设定 → 分镜脚本（三步走 LLM Agent）→ 图片生成（`ImageService` + 混元）→ 排版合成（`ComposeService`）→ 公众号发布（`PublishService`）。每步先 `UpdatePhase`，再 `SyncState` 落库，失败 `MarkFailed`。

状态/阶段常量见 `model/comic.go`（`ComicStatus*` / `ComicPhase*`）。内存态用 `ComicState` 在各步间传递；数据库六步产物以 JSON 列存储，`Comic.ToComicInfo()` 读时解析为结构体。新增 LLM 步骤：实现 `agent.Agent` 接口（`Execute(ctx, *ComicState) error`），在 orchestrator 的 steps 切片挂入。

### LLM 与外部依赖降级

- LLM 走 `langchaingo` OpenAI 兼容模式连通义千问（qwen-plus），初始化在 `service.NewLLM`。**若 DashScope api_key 未配置，整个漫画模块被禁用**（`app.go` 里 `comicHandler` 为 nil，`main.go` 跳过 `/comic` 路由注册），用户功能仍可用。
- 混元生图（`hunyuan.enabled`）、公众号发布（`wechat.enabled`）未开启时走占位/草稿降级，不阻断流水线。
- 提示词集中在 `common/prompt.go`（`BuildXxxPrompt`）；LLM 返回的 JSON 用 `pkg/llmjson.Unmarshal` 容错解析。

## 前端架构（web/）

React 19 + React Router 7 + Zustand + Axios + Sass，路径别名 `@/` → `src/`。

页面按受众分目录（勿混用）：`pages/common/`（公开：home、auth）、`pages/user/`（登录用户：create、history、info、article/detail）、`pages/admin/`（管理员：Users、StaticPage）。路由守卫 `router/guards.tsx`（`RequireAuth` / `RequireAdmin`）。个人资料归 `user/`，用户 CRUD 归 `admin/`。

- API 层：`api/*.ts` 用 `request` + `unwrap()`（`utils/request.ts`，`withCredentials: true`）；类型 `types/api.ts` 与 Go `model` 对齐。
- 主题：`stores/theme.ts`（`classic` | `immersive`），SCSS 用 CSS 变量（`--app-page-bg` 等）；共享布局 `@import '@/pages/_shared/pageShell.scss'`。
- 列表页用 Ant Design `Table`，管理端列抽到同目录 `*TableColumns.tsx`。

### Ant Design 自动引入（重要）

`vite.config.ts` 配置了 `unplugin-react-components` + `unplugin-auto-import`，组件与图标自动引入，**TSX 中勿手写 `import ... from 'antd'`**。生成的 `components.d.ts` / `auto-imports.d.ts` 勿提交。仍需手写的：

- `import type { ... } from 'antd'`（如 `MenuProps`、`TablePaginationConfig`）
- `import zhCN from 'antd/locale/zh_CN'`
- 静态 API `message`、`notification`、`Modal`、`Form`（含 `Form.useForm()`）已自动引入
- 勿把 `main.tsx` 根组件改名为 `App`，否则要同步改 vite resolver 的排除项

> 注意：`vite.config.ts` 的 dev 代理目标为 `http://localhost:2026`，而 `config.yaml.example` 默认 `server.port: 8080`。本地起后端时请让端口与代理一致（或改其一）。

## 提交约定

Conventional Commits，由 husky `commit-msg` + commitlint 校验（`npm run prepare` 注册 hook），`.cursor/hooks.json` 在 `git commit` 前跑 oxlint + `go vet` + 提交信息校验。

```
<type>(<scope>): <subject>
# type: feat|fix|docs|style|refactor|perf|test|chore|ci|build|revert
# 例：feat(web): 添加用户中心表格   fix(server): 修复 session 过期
```

默认不做（除非明确要求）：force push、amend 已推送提交、跳过 hooks、无关的大规模重构、未经要求的 `git commit`；勿提交 `config.yaml`、`.env` 等含密钥文件。
