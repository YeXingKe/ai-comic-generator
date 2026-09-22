# Go 后端小白指南：产品介绍与技术方案

> 版本：v1.0  
> 日期：2026-08-25  
> 范围：仓库 `server/`（Go + Gin + GORM）  
> 读者：第一次接触本项目后端的同学（产品 + 技术一体说明）  
> 相关：环境搭建见 [`AGENTS.md`](../AGENTS.md)；鉴权增强见 [`prd-auth.md`](./prd-auth.md) / [`tech-auth.md`](./tech-auth.md)；充值见 [`prd-recharge.md`](./prd-recharge.md)

---

## 0. 先用一句话说清楚

**ai-comic-generator 的后端**负责：用户登录、额度与权限，以及把「一个主题」通过多智能体流水线变成「可预览、可发布的漫画长图」。

前端 `web/` 只负责页面与轮询；真正跑 LLM、生图、拼图、落库的，是本目录的 Go 服务。

```text
用户在网页点「创作」
    → 浏览器请求 /api/...
    → Go 服务（本仓库 server/）
    → MySQL 存任务与产物
    → Redis 存登录 Session
    → 通义千问（写故事/角色/分镜）
    → 混元或 OpenAI 兼容生图（画格子）
    → 本地/COS 存图片
    → （可选）微信公众号发草稿
```

---

## 1. 产品介绍（给小白）

### 1.1 这个产品解决什么问题？

普通人想发公众号/小红书漫画，但不会画、不会写分镜。本产品提供两条创作路径：

| 模式 | 用户做什么 | 系统做什么 | 典型用途 |
|------|------------|------------|----------|
| **自动化创作** | 填主题、风格、格数；确认标题与分镜 | 故事 → 角色 → 分镜 → 生图 → 拼成长图 → 可选发公众号 | 公众号条漫 |
| **自定义创作** | 给提示词，指定格数 | 拆成多格并直接生图，可下载 ZIP | 小红书等快速出图 |

### 1.2 用户能感知到的完整旅程（自动化）

下面按「人能看见的步骤」说明，括号里是系统内部阶段名。

```text
① 创建任务
   用户输入：主题、风格、生图后端、文案模式、格数(4/6/8)
   系统：立刻返回 taskId，后台生成多个标题候选
   状态：PROCESSING → AWAITING_CONFIRM（等选标题）

② 确认标题
   用户：从推荐里选一个，或自己改标题
   系统：写入标题，进入 TITLE_CONFIRMED（还不自动开跑）

③ 开始生成
   用户：点「开始生成漫画」
   系统异步跑：
     故事构思 (STORY_IDEATION)
     角色设定 + 定妆照 (CHARACTER_DESIGN)
     分镜脚本 (STORYBOARD_SCRIPT)
   然后停住：AWAITING_STORYBOARD（等人确认分镜）

④ 确认分镜
   用户：可改某一格的场景/台词/Prompt，再确认
   系统异步跑：
     逐格生图 (IMAGE_GENERATION)
     竖向拼长图 (LAYOUT_COMPOSE)
   完成：COMPLETED

⑤ （可选）发布
   用户：在历史里点发布
   系统：调微信公众号接口；未配置时仅记草稿结果
```

失败时状态为 `FAILED`，可调用「从失败阶段重试」；某一格不满意可「单格重绘」后再合成。

### 1.3 角色与权限（产品视角）

| 角色 | 能做什么 |
|------|----------|
| 游客 | 注册、登录 |
| 普通用户 / VIP | 创作、查自己的任务、改资料；VIP/管理员设计上可不扣额度 |
| 管理员 | 用户 CRUD、改额度、看数据看板 `/stat/dashboard` |

积分字段是 `points`。管理端可改；支付充值见独立 PRD（尚未必接）。角色仅 `user` / `admin`，无 VIP。

### 1.4 核心产品能力一览

1. **多智能体流水线**：标题 / 故事 / 角色 / 分镜 各是一个 LLM Agent。  
2. **人机卡点**：标题确认、分镜确认——避免一次跑完无法干预。  
3. **角色一致性**：`visualAnchor` 短锚点 + 定妆照 `avatarUrl`；生图时强制把锚点塞进 Prompt 头部。  
4. **生图可切换**：混元 / OpenAI 兼容 1K / 4K；都关则用占位图，流水线仍可跑通。  
5. **文案叠加**：`none` / `top` 顶部字幕 / `bubble` 气泡。  
6. **降级策略**：LLM 未配置则禁用自动化 `/comic`；公众号/COS/混元可关，不挡主流程。

### 1.5 产品边界（当前不做或另文档）

- 完整支付收银台：见 `prd-recharge.md`  
- 鉴权加固（密码强度、禁用即时失效等）：见 `prd-auth.md`  
- 端到端消息队列：当前用进程内 `go func()` 异步，进程挂了进行中任务会丢

---

## 2. 技术总览（给小白）

### 2.1 技术栈一句话

| 层 | 技术 | 干什么 |
|----|------|--------|
| HTTP | Gin | 路由、中间件、JSON |
| 业务 | 自写 service | 校验、编排、扣权限 |
| 数据 | GORM + MySQL | 用户、comic、custom_comic |
| 登录 | Cookie Session + Redis | 记住谁登录了 |
| LLM | langchaingo（OpenAI 兼容）→ 通义千问 | 写文案与结构化 JSON |
| 生图 | 腾讯混元 / OpenAI Image 兼容 | 出格子图 |
| 拼图 | imaging + gg | 竖向长图、标题、字幕 |
| 对象存储 | 本地目录 或 腾讯云 COS | 图片 URL |
| 发布 | 微信公众号 API | 可选 |

Go 版本见 `server/go.mod`（当前为 1.25.x 一带）。

### 2.2 目录地图（打开 `server/` 就认路）

```text
server/
├── cmd/server/main.go      # 程序入口：读配置、装中间件、启动
├── config.yaml.example     # 配置模板（复制为 config.yaml，勿提交密钥）
├── sql/                    # 建表与增量迁移 SQL
└── internal/               # 业务代码（外部包不能 import internal）
    ├── router/             # HTTP 路由 Register(r, cfg, app)
    ├── app/                # 组装：DB、Redis、各 Service/Handler 注入
    ├── config/             # 解析 YAML
    ├── handler/            # HTTP 薄层：绑 JSON → 调 service → 统一响应
    ├── service/            # 业务逻辑 + 流水线编排 + 生图/合成/发布
    ├── store/              # 只访问数据库（GORM）
    ├── model/              # 表实体、请求体、状态常量、内存态 ComicState
    ├── middleware/         # CORS、Session、AuthCheck
    ├── common/             # 响应、错误码、Prompt、锚点工具
    ├── agent/              # Agent 接口 + title/story/character/script
    ├── client/             # 外部 SDK：hunyuan / gpt / cos / wechat
    ├── storage/            # 本地路径与 PublicURL
    └── pkg/llmjson/        # 从 LLM 脏输出里抠 JSON
```

**记忆口诀**：请求先进 `handler`，规矩在 `service`，SQL 只在 `store`，长相在 `model`，启动拼装在 `app`，对外喊人在 `client`。

### 2.3 一次请求怎么走？

以「创建漫画」为例：

```text
POST /api/comic/create  (+ Cookie Session)
  → middleware.AuthCheck   确认已登录，把 LoginUser 放进 Context
  → ComicHandler.Create    绑定 CreateComicRequest
  → ComicService.Create    生成 taskId，写 comic 表，go func 跑标题
  → common.Success({ taskId })
  → 前端拿到 taskId，轮询 GET /api/comic/get
```

统一响应形状（几乎总是 HTTP 200，看 body 里的 `code`）：

```json
{ "code": 0, "data": { "...": "..." }, "message": "ok" }
```

常见业务码：

| code | 含义 |
|------|------|
| 0 | 成功 |
| 40000 | 参数错误 |
| 40100 | 未登录 |
| 40101 / 40300 | 无权限 |
| 40400 | 数据不存在 |
| 50000 / 50001 | 系统/操作失败 |

> 注意：鉴权中间件对未登录有时会直接返回 HTTP 401，与部分业务错误的「HTTP 200 + code」并存。前端 `request.ts` 需两边都能处理。

---

## 3. 核心技术方案：自动化漫画流水线

### 3.1 为什么要「编排器」？

步骤多、耗时长（几分钟级），不能阻塞 HTTP。设计拆成：

1. **接口立刻返回**（或只做状态校验）  
2. **`ComicOrchestrator` 在后台 `go func()` 里跑步骤**  
3. **每步：改 phase → 执行 → `SyncState` 把内存态写入 DB JSON 列**  
4. **失败：`MarkFailed`，前端可 retry**

编排器代码：`internal/service/comic_orchestrator.go`。

### 3.2 两段式 + 两个人机卡点

| 编排方法 | 做什么 | 结束后状态 |
|----------|--------|------------|
| `RunTitles` | 标题 Agent | `AWAITING_CONFIRM` |
| `RunFromStory` | 故事 → 角色(+定妆照) → 分镜 | `AWAITING_STORYBOARD` |
| `RunFromImages` | 生图 → 合成 | `COMPLETED` |
| `RetryFromPhase` | 从失败 phase 续跑对应段 | 同上 |

人机卡点对应 API：

- `POST /comic/confirm-title`  
- `POST /comic/start`（启动 `RunFromStory`）  
- `POST /comic/confirm-storyboard`（启动 `RunFromImages`）  
- `POST /comic/retry`  
- `POST /comic/regenerate-panel`

### 3.3 内存态 `ComicState` vs 数据库 `Comic`

| | Comic（表） | ComicState（内存） |
|--|-------------|-------------------|
| 用途 | 持久化、列表、前端查询 | 步骤之间传递 |
| JSON 列 | 存字符串 | 已解析成结构体 |
| 谁维护 | store `SyncState` / `UpdatePhase` | Agent / Image / Compose 读写 |

读接口时：`Comic.ToComicInfo()` 把 JSON 列解析成前端友好结构。

### 3.4 各 Agent 做什么？

都实现同一接口：

```go
type Agent interface {
    Execute(ctx context.Context, state *model.ComicState) error
}
```

| Agent | 阶段 | 产出写入 state |
|-------|------|----------------|
| TitleAgent | TITLE_GENERATION | 多个标题选项 |
| StoryAgent | STORY_IDEATION | 梗概、冲突、亮点、标题 |
| CharacterAgent | CHARACTER_DESIGN | 角色列表；并规范化 `visualAnchor` |
| ScriptAgent | STORYBOARD_SCRIPT | 分镜格（含英文 imagePrompt） |

Prompt 集中在 `internal/common/prompt.go`（中文）与 `prompt_en.go`（英文），由 `PromptBuilder` 按 `ai.prompt_lang` 选择。LLM 返回的 JSON 用 `pkg/llmjson` 容错解析。

### 3.5 角色一致性（Week3 能力）

1. LLM 必须产出 `visualAnchor`（短外貌关键词）；没有则从 `appearance` 压缩。  
2. 角色步骤后调用 `ImageService.GenerateCharacterAvatars` 生成定妆照。  
3. 分镜 Prompt 要求复用锚点；最终生图前 `ForceInjectCharacterAnchors`，截断时**优先保留锚点**。  

相关：`common/character_anchor.go`、`service/image_service.go`。

### 3.6 生图与合成

**ImageService**

- 按任务的 `imageBackend` 从注册表选生成器（混元 / GPT 1K / GPT 4K）。  
- 未启用则画占位图，方法标记 `PLACEHOLDER` 或等价降级。  
- 可按 `captionTextMode` 叠字幕/气泡（`panel_caption.go`）。  
- 成功后可选上传 COS，URL 写回 `panelImages`。

**ComposeService**

- 固定单格约 960×540（16:9），再按模板拼成品图。  
- **现状**：一律单列竖拼 + 顶栏标题（观感偏弱）。  
- **需求变更（待做）**：1～3 格单列竖拼；4/6/8 格两列网格；统一黑框与 gutter。详见 [`prd-compose-layout.md`](./prd-compose-layout.md)。  
- 输出 `composed.png`（及预览 URL），写入 `composedLayout`。

### 3.7 自定义创作（旁路）

路径前缀：`/api/comic/custom`。

- **不跑**完整故事/分镜确认状态机。  
- 用户给一段 prompt + 格数，服务拆格生图，可 ZIP 下载。  
- **不依赖** DashScope 必配（无 LLM 时有兜底拆分）；与自动化表分离（`custom_comic`）。

---

## 4. 用户、登录与权限

### 4.1 Session 方案

1. 登录成功：在 Redis 里写入 Session，浏览器带 Cookie（默认名 `session`）。  
2. 后续请求：`AuthCheck` 读 Session → `UserService.GetLoginUser`。  
3. 管理员接口：`AuthCheck(..., AdminRole)`，再校验角色。  

配置项：`session.secret` / `max_age` / `secure` / `same_site`；CORS 允许前端源（如 `http://localhost:5173`）。

### 4.2 用户模型要点

- 角色：`user` / `admin` / `vip`  
- `status`：启用/禁用  
- `points`：用户积分（store 已有 `DecrementPoints` / `AddPoints`；创作扣积分建议在 create/start 打通）  
- 启动时可 `EnsureAdmin` 保证默认管理员存在（文档与 AGENTS：`admin` / `admin123456`，以实际实现为准）

### 4.3 用户相关 API

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/user/register` | 无 |
| POST | `/api/user/login` | 无 |
| GET | `/api/user/info` | Session（实现以代码为准） |
| POST | `/api/user/logout` | Session |
| POST | `/api/user/profile/update` | 登录 |
| POST | `/api/user/password/update` | 登录 |
| POST | `/api/user/page/vo` 等 | 管理员 |

---

## 5. 数据与存储

### 5.1 主要表

| 表 | 作用 |
|----|------|
| `user` | 账号、角色、额度、VIP |
| `comic` | 自动化任务：状态机 + 各步 JSON 产物 |
| `custom_comic` | 自定义多格任务 |

建表与增量脚本在 `server/sql/`（如 `create_table.sql`、`increment_add_comic_panel_count.sql` 等）。**改表结构请走增量 SQL，勿只改 GORM 模型。**

### 5.2 `comic` 表设计思路

- 一行 = 一次创作任务，用 `taskId`（UUID）对外。  
- 过程产物放 JSON 列，避免过早拆很多子表，适合流水线迭代。  
- `status` 管总态，`phase` 管当前步，前端进度条靠这两个字段。

### 5.3 图片存哪？

```text
storage.base_path  (默认 ./comics)
  └── {taskId}/
        panel_1.png
        avatar_1.png
        composed.png

公开访问：config 里 public_url（默认 /comics）
Gin：r.Static(public_url, base_path)

若 cos.enabled：上传后优先返回 COS URL
```

---

## 6. API 清单（自动化漫画）

前缀均为 `config.server.context_path`（默认 `/api`），下列均需登录（除特殊说明）。

| 方法 | 路径 | 作用 |
|------|------|------|
| POST | `/comic/create` | 创建任务，异步标题推荐 |
| POST | `/comic/confirm-title` | 确认/编辑标题 |
| POST | `/comic/start` | 启动故事→角色→分镜 |
| POST | `/comic/confirm-storyboard` | 确认分镜，启动生图→合成 |
| POST | `/comic/retry` | 失败后从阶段重试 |
| POST | `/comic/regenerate-panel` | 单格重绘并再合成 |
| POST | `/comic/publish` | 公众号发布 |
| GET | `/comic/get` | 查单个任务详情 |
| POST | `/comic/page` | 分页列表 |
| POST | `/comic/custom/create` | 自定义创作 |
| GET | `/comic/custom/get` | 自定义详情 |
| POST | `/comic/custom/page` | 自定义列表 |
| GET | `/comic/custom/download` | 下载 ZIP |
| GET | `/health` | 健康检查 |
| GET | `/stat/dashboard` | 管理端统计 |

新增接口推荐顺序（与仓库 skill `add-api` 一致）：

`model` → `store` → `service` → `handler` → `internal/router/router.go` 的 `Register` → 前端 `types/api.ts` + `api/`。

---

## 7. 配置与降级矩阵

复制 `config.yaml.example` → `config.yaml`。

| 配置块 | 关掉或未配时的行为 |
|--------|-------------------|
| `ai.dashscope.api_key` | **自动化漫画模块整体不注册**（`ComicHandler == nil`） |
| `ai.hunyuan.enabled` | 该后端不可用；可改用其它 imageBackend 或占位图 |
| `openai_image_*` | 同上 |
| `cos.enabled` | 只用本地 `/comics` URL |
| `wechat.enabled` | 发布走草稿/降级，不阻断创作完成 |

开发建议：先只配 DashScope，混元关掉，用占位图把状态机和前端联调跑通，再开真生图。

端口：`server.port` 默认 8080。注意前端 Vite 代理有时指向 `2026`，本地需两边一致，否则会出现代理连不上。

---

## 8. 启动与自检（小白清单）

1. 安装 Go、MySQL、Redis。  
2. 建库执行 `sql/create_table.sql`，并按需跑 `sql/increment_*.sql`。  
3. `cp config.yaml.example config.yaml`，填数据库密码与（可选）DashScope Key。  
4. 仓库根目录：`npm run server` 或 `cd server && go run ./cmd/server/main.go`。  
5. 浏览器或 curl：`GET http://localhost:8080/api/health`。  
6. API 文档（Swaggo）：`http://localhost:8080/swagger/index.html`；改 handler 注解后执行 `npm run swagger` 重新生成 `server/docs/`。  
6. 登录后走一遍：create → get 轮询 → confirm-title → start → …  

静态检查：`cd server && go vet ./...`；有单测的包：`go test ./internal/common/ ...`。

---

## 9. 架构图（总览）

```text
                 ┌──────────────┐
                 │  web (React) │
                 └──────┬───────┘
                        │ Cookie + /api
                 ┌──────▼───────┐
                 │  Gin main.go │
                 │  CORS/Session│
                 └──────┬───────┘
         ┌──────────────┼──────────────┐
         ▼              ▼              ▼
    UserHandler   ComicHandler   Custom/Stat
         │              │
         ▼              ▼
    UserService   ComicService ── go func ──► ComicOrchestrator
         │              │                      │
         ▼              ▼                      ├ Title/Story/Character/Script Agents
    UserStore      ComicStore                  ├ ImageService ──► hunyuan/gpt
         │              │                      └ ComposeService ──► gg/imaging
         └──────┬───────┘
                ▼
         MySQL / Redis
                │
         Local FS / COS / WeChat
```

---

## 10. 设计取舍（读代码时别踩坑）

1. **进程内异步，不是消息队列**  
   实现简单，适合个人项目；多副本/重启会丢任务。以后要稳，可再引入队列，但接口与状态机可先不动。

2. **JSON 大列存产物**  
   迭代 Agent 输出快；缺点是列表查询不宜扫巨大 JSON（列表接口应只取摘要字段——以 store 实现为准）。

3. **Handler 禁止写 SQL**  
   保持可测、可换存储。新功能不要图省事在 handler 里 `db.Create`。

4. **角色校验在服务端**  
   不信任前端传来的 `userRole`；管理员能力以 Session 用户为准。

5. **50000 不泄露内部细节**  
   日志可以详，返回给前端要克制（尤其密钥、堆栈）。

---

## 11. 推荐阅读顺序（小白学习路径）

1. 本文第 1～2 章（产品 + 目录）  
2. `internal/router/router.go`（有哪些路由）  
3. `internal/app/app.go`（依赖怎么拼起来）  
4. `handler/comic_handler.go` + `service/comic_service.go`（人机卡点）  
5. `service/comic_orchestrator.go`（状态机心脏）  
6. `agent/agents/*.go` + `common/prompt.go`（AI 在写什么）  
7. `service/image_service.go` + `compose_service.go`（图怎么来）  
8. `middleware/auth.go` + `service/user_service.go`（登录）  
9. 需要加固/支付时再读 `docs/tech-auth.md`、`docs/tech-recharge.md`

---

## 12. 术语表

| 术语 | 含义 |
|------|------|
| taskId | 一次创作任务的 UUID |
| phase | 流水线当前步骤名 |
| status | 任务总状态（等确认/进行中/完成/失败） |
| Agent | 单步 LLM 智能体 |
| Orchestrator | 按顺序调用 Agent/生图/合成的编排器 |
| visualAnchor | 角色外貌短锚点，保证多格长得像 |
| imageBackend | 用户选择的生图引擎 |
| captionTextMode | 是否在图上叠字幕/气泡 |
| Session | 存在 Redis 的登录凭证，靠 Cookie 带回 |

---

## 13. 文档维护

- 路由以 `internal/router/router.go` 为准；本文若滞后，以代码为准。  
- 表结构以 `server/sql/` 为准。  
- 产品向增强（鉴权、充值）不写进本文细节，见同目录其它 PRD/Tech。
