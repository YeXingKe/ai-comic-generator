---
name: init
description: >-
  初始化 ai-comic-generator 本地开发环境。在用户执行 /init、首次搭建项目、克隆仓库后 onboarding、
  或 /api 出现 ECONNREFUSED 时使用。
disable-model-invocation: true
---

# 项目初始化（ai-comic-generator）

按流程端到端执行，勿跳过验证步骤。

## 环境要求

- Node.js 20+
- Go 1.22+
- MySQL 8+（本地已启动）
- Redis：按 `server/config.yaml` 配置（Session 依赖时可必需）

## 初始化清单

复制并勾选进度：

```text
- [ ] MySQL 数据库已创建
- [ ] server/config.yaml 已复制并编辑
- [ ] 根目录与 web 依赖已安装
- [ ] Go modules 已 tidy
- [ ] 开发服务启动无报错
- [ ] 前端可访问且 API 代理正常
```

## 步骤 1：数据库

```sql
CREATE DATABASE ai_comic_generator CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 步骤 2：服务端配置

在仓库根目录：

```bash
cp server/config.yaml.example server/config.yaml
```

编辑 `server/config.yaml`：

- 填写 `database.password`
- 确认 `database.host`、`database.port`、`database.name`、`database.user`
- 非本地环境请设置强 `session.secret`

## 步骤 3：安装依赖

在仓库根目录：

```bash
npm install
npm run install:web
cd server && go mod tidy
```

## 步骤 4：启动开发

在仓库根目录（前后端同时）：

```bash
npm run dev
```

或分别启动：

```bash
npm run web      # 仅前端 → http://localhost:5173
npm run server   # 仅后端 → http://localhost:8080/api
```

首次构建会生成 `web/src/components.d.ts` 与 `web/src/auto-imports.d.ts`（已 gitignore，由 Vite 插件生成）。

## 步骤 5：验证

| 检查项 | 期望结果 |
|--------|----------|
| http://localhost:5173 | 首页正常显示 |
| http://localhost:8080/api/health | 健康检查通过 |
| 管理员登录 | `admin` / `admin123456`（首次启动 server 自动创建） |

若 Vite 出现 `http proxy error: /api/... ECONNREFUSED`，说明 Go 未启动或 `config.yaml` 数据库连接失败。

## 页面目录

```text
web/src/pages/
├── common/     # 公开：home、auth
├── user/       # 登录用户：center、create、history、article/detail
├── admin/      # 管理员：users、data
└── _shared/    # 共享样式与表格工具
```

## 关键路由

| 路径 | 权限 |
|------|------|
| `/` | 公开 |
| `/user/login` | 公开 |
| `/user/center` | 需登录 |
| `/history`、`/create` | 公开（当前多为 mock/静态） |
| `/admin/users` | 仅管理员 |
| `/admin/data` | 仅管理员 |

兼容重定向：`/data` → `/admin/data`，`/admin/userManage` → `/admin/users`。

## 前端约定

- antd 组件/图标：`unplugin-react-components` 自动引入，多数文件勿手写 `import from 'antd'`
- `message`、`Form.useForm()` 等：`unplugin-auto-import` 自动引入
- 类型（`MenuProps` 等）：仍用 `import type from 'antd'`
- 主题：`classic` / `immersive`（`web/src/stores/theme.ts`）

## 构建（可选验证）

```bash
npm run build:web
npm run build:server
```

## 禁止

- 提交含真实密码的 `server/config.yaml`
- 提交 `web/src/components.d.ts` 或 `web/src/auto-imports.d.ts`（自动生成）
