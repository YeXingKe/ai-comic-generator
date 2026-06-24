# AI Comic Generator

AI 漫画内容创作平台 — 前后端同仓、分目录、接口分离、可独立部署。

## 项目结构

```text
ai-comic-generator/
├── web/          # React + Ant Design + Vite 前端
├── server/       # Go + Gin + GORM 后端
└── package.json  # 根目录一键启动前后端
```

## 技术栈

| 模块 | 技术 |
|------|------|
| 前端 | React 19、Ant Design、dayjs、React Router、Zustand、Axios、Vite |
| 后端 | Go、Gin、GORM、MySQL、Session |
| 主题 | 紫色（参考 [ai-content-creator](https://github.com/YeXingKe/ai-content-creator) 布局，颜色改为紫色） |

## 已实现页面

- 首页 `/`
- 文章列表 `/article/list`（静态示例数据）
- 文章详情 `/article/:taskId`（静态示例数据）
- 用户登录 `/user/login`
- 用户注册 `/user/register`
- 用户管理 `/admin/userManage`（需管理员登录）
- 创作编辑页 `/create` — **暂未实现**

## 快速开始

### 1. 准备数据库

创建 MySQL 数据库：

```sql
CREATE DATABASE ai_comic_generator CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

复制配置文件：

```bash
cp server/config.yaml.example server/config.yaml
```

按本地环境修改 `server/config.yaml` 中的数据库账号密码。

### 2. 安装依赖

```bash
npm run install:web
cd server && go mod tidy
```

### 3. 启动开发环境（根目录同时启动前后端）

```bash
npm run dev
```

- 前端：http://localhost:5173
- 后端：http://localhost:8080/api
- 健康检查：http://localhost:8080/api/health

### 4. 默认管理员

首次启动会自动创建管理员账号：

- 账号：`admin`
- 密码：`admin123456`

## 独立部署

```bash
# 前端
npm run build:web
# 产物在 web/dist/

# 后端
npm run build:server
# 产物在 bin/server
```

生产环境前端通过 `web/.env.production` 配置 `VITE_API_BASE_URL` 指向后端 API 地址。

## 开发命令

| 命令 | 说明 |
|------|------|
| `npm run dev` | 同时启动 web + server |
| `npm run dev:web` | 仅启动前端 |
| `npm run dev:server` | 仅启动后端 |
| `npm run build:web` | 构建前端 |
| `npm run build:server` | 构建后端 |
