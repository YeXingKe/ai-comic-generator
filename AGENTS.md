# AGENTS.md

面向 Cursor Cloud Agent、子 Agent 及在本仓库使用 AI 的开发者。

## 项目简介

**ai-comic-generator** — 前后端同仓：`web/`（React + Vite + Ant Design）+ `server/`（Go + Gin + GORM + MySQL + Redis Session）。

## 从这里开始

| 需求 | 阅读 / 执行 |
|------|-------------|
| 首次搭建环境 | Skill：`/init` → `.cursor/skills/init/SKILL.md` |
| 项目通用约定 | `.cursor/rules/project.mdc`（始终生效） |
| 前端开发 | `.cursor/rules/web.mdc` + `pages.mdc` |
| 后端开发 | `.cursor/rules/server.mdc` + `handler.mdc` |
| 新增全栈 API | Skill：`add-api` |
| PR 审查 | Skill：`review-pr` |

## 常用命令（仓库根目录）

```bash
npm run dev      # web :5173 + server :8080
npm run web      # 仅前端
npm run server   # 仅后端
npm run lint --prefix web
cd server && go vet ./...
```

配置：复制 `server/config.yaml.example` → `server/config.yaml`。默认管理员：`admin` / `admin123456`。

## 架构速览

```
web/src/pages/common/   → 公开页（home、auth）
web/src/pages/user/     → 普通用户（center、create、history…）
web/src/pages/admin/    → 管理员（users、data）
server/internal/        → handler → service → store → model
```

路由：`/user/center`（需登录）、`/admin/users` 与 `/admin/data`（仅管理员）。API 前缀：`/api`。

## 自动化

### Git Hooks（husky + commitlint）

- **`commit-msg`**：校验 Conventional Commits（见 `commitlint.config.mjs`）
- 安装依赖后执行 `npm run prepare` 注册 hooks

### Cursor Hooks（`.cursor/hooks.json`）

- **beforeShellExecution**（`git commit`）：web oxlint + `go vet` + 提交信息规范（含 `-m` 时）

## 提交信息格式

```text
<type>(<scope>): <subject>
```

type：`feat` `fix` `docs` `style` `refactor` `perf` `test` `chore` `ci` `build` `revert`

示例：`feat(web): 添加用户中心表格`、`fix(server): 修复 session 过期`

## Agent 行为约定

- 改动尽量小，遵循现有模式
- 勿提交密钥或生成的 `components.d.ts` / `auto-imports.d.ts`
- antd 组件/图标自动引入，TSX 中勿手写 `import from 'antd'`
- 用户中心与用户管理分离，勿合并

## 默认不做（除非用户明确要求）

- force push、amend 已推送提交、跳过 hooks
- 与任务无关的大规模重构
- 未经要求自动 git commit
