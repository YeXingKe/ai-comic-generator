# 技术方案：登录鉴权增强

> 版本：v1.2  
> 日期：2026-08-14（v1.2：2026-09-22 增加「改进项跟踪」状态列；v1.1 单 Session 方案）  
> 关联产品文档：[`prd-auth.md`](./prd-auth.md)  
> 实现方式：**自行按本文档落地**（不依赖 Agent 直接改业务代码）  
> 建议顺序：先完成本文档 **P0**，再实现支付（[`tech-recharge.md`](./tech-recharge.md)）

---

## 1. 总览

### 1.1 结论

- **继续使用** Redis Session + Cookie + `AuthCheck`  
- **不建议**本期改为纯 JWT  
- 重点是：密码、Cookie、禁用校验、CORS、前端未登录处理  

### 1.2 原则

1. 最小改动，贴合现有 `handler → service → store`  
2. 存量用户无感迁移（登录时升级哈希）  
3. 配置与密钥不进仓库  

---

## 2. 改进项跟踪（登录 / 注册 / Session）

**状态说明**：✅ 已实现 · 🔶 部分实现 · ⬜ 未实现 · 📋 运维/产品（非代码）  
**审查基准**：仓库代码（2026-09-22）；实现后请同步更新本表。

| 优先级 | 状态 | 项 | 说明 | 代码锚点 |
|--------|------|-----|------|----------|
| P0 | ✅ | 新注册 bcrypt | 注册写入 bcrypt 哈希 | `UserService.Register` → `hashPassword` |
| P0 | 🔶 | 旧 MD5 登录迁移 | 登录校验 MD5 成功后升级 bcrypt；盐常量仍保留供迁移 | `verifyPassword`、`common.PasswordSalt` |
| P0 | ✅ | 移除公开加密接口 | 路由未注册 `POST /user/encrypt/password` | `internal/router/router.go` |
| P0 | ✅ | Session Cookie 可配置 | `Secure` / `SameSite` / `MaxAge` / `HttpOnly` | `middleware/session.go`、`config.Session` |
| P0 | ✅ | CORS 白名单 | 按 `cors.allow_origins` 回写 Origin | `middleware/cors.go` |
| P0 | 🔶 | 禁用账号即时失效 | `GetLoginUser` 清 Session 并带禁用文案；**受保护接口经 `AuthCheck` 时文案被抹成通用未登录** | `UserService.GetLoginUser`、`middleware/auth.go` |
| P0 | ⬜ | 禁用专用错误码 `40102` | 与「未登录」区分，便于前端提示 | `common/error.go`、`auth.go`、`request.ts` |
| P0 | ✅ | 前端未登录统一跳转 | HTTP 401 与 `code===40100`（`/user/info` 除外） | `web/src/utils/request.ts` |
| P0 | 🔶 | 鉴权 HTTP 形态统一 | 业务接口多 200+code；`AuthCheck` 仍 **401** + JSON | `middleware/auth.go` vs 其它 handler |
| P0 | 📋 | 生产 Session `secret` | 须部署换长随机密钥，勿用 example 默认值 | `config.yaml`（不进库） |
| P1 | ⬜ | 单 Session / 异地挤出 | Redis `login:ver` + Session `userLoginVer`，见 §4.1 | `UserService.Login` / `GetLoginUser` |
| P1 | ⬜ | 被挤出错误码 `40103` | 前端 `message` 后再跳登录 | §4.1.5、`request.ts` |
| P1 | ⬜ | 改密后踢全员 Session | 改密后 `INCR login:ver` 或 `Clear` Session | `UserService.UpdatePassword` |
| P1 | ⬜ | 登出彻底清 Session | 建议 `session.Clear()` + `Save()` | `UserService.Logout` |
| P1 | ⬜ | 登录失败限流 | Redis 计数 IP/账号，超限 `42900` | 待增 |
| P1 | ⬜ | 登录 Session 固定防护 | 登录成功后轮换 Session ID（若库支持） | `UserService.Login` |
| P1 | ⬜ | 前端区分 40102/40103 | 禁用、被挤下线专用文案 | `web/src/utils/request.ts` |
| P1 | 🔶 | 注册开放策略一致 | UI 关闭注册 Tab + Alert，**`POST /user/register` 仍开放** | `pages/common/auth/index.tsx`、`router` |
| P1 | ⬜ | 服务端注册开关 | 如 `security.allow_register` | `config` + `UserService.Register` |
| P2 | ⬜ | 找回密码 / 验证码 / MFA | 产品另开 | — |
| P2 | ⬜ | 密码复杂度 / 弱口令库 | 现仅 ≥8 位 | `Register` / `UpdatePassword` |
| P2 | ⬜ | 安全审计日志 | 登录成败、改密、禁用等 | 待增 |
| P2 | ⬜ | CSRF Token | 现依赖 SameSite；跨站 `None` 时需加强 | — |
| — | 🔶 | 代码卫生 | `UpdatePassword` 末尾重复 `return nil`（死代码） | `user_service.go` |

### 2.1 初版问题对照（已并入上表）

原「MD5 弱哈希、加密接口、CORS 写死、禁用仅 Login」等条目，见上表 P0 行；**路由注册**已迁至 `internal/router/router.go` → `Register`。

---

## 3. P0 改造说明

### 3.1 移除或锁定 `/user/encrypt/password`

**推荐**：删除路由注册；若 handler/service 仅被该路由使用，一并删除对外暴露。

**备选**：`security.allow_encrypt_api: false`（默认），且仅管理员 + 非生产可开。

### 3.2 密码哈希：bcrypt + 旧 MD5 迁移

#### 存储

- 字段仍为 `user.userPassword`  
- 建议列长度 ≥ **255**（bcrypt 串约 60，留余量）  
- 新注册、管理员创建默认密码、用户改密：一律写 **bcrypt**（cost ≥ 10）

#### 校验伪代码

```text
func verifyAndMaybeUpgrade(user, plain):
  stored := user.UserPassword
  if strings.HasPrefix(stored, "$2a$") || HasPrefix(stored, "$2b$"):
    return bcrypt.Compare(stored, plain)
  // 旧逻辑
  if stored == md5(plain + PasswordSalt):
    newHash := bcrypt.Generate(plain)
    _ = store.UpdatePassword(user.ID, newHash)  // 升级；失败应记日志
    return true
  return false
```

#### 涉及调用点（自行排查替换）

- `Register`  
- `Login`  
- `UpdatePassword`（校验旧密码 + 写新密码）  
- `Create`（管理员创建用户默认密码）  
- `EncryptPassword`（若删除接口则删除）  

依赖：Go 标准库周边常用 `golang.org/x/crypto/bcrypt`。

### 3.3 Session Cookie 选项

修改 `SetupSession` 中 `sessions.Options`，并从配置读取：

| 配置项 | 开发建议 | 生产建议 |
|--------|----------|----------|
| `HttpOnly` | true | true |
| `Secure` | false | true（HTTPS） |
| `SameSite` | `Lax` | `Lax` |
| `MaxAge` | 可沿用或改为 7 天 | ≤ 7 天更稳妥 |
| `Path` | `/` | `/` |
| `secret` | 本地随机 | 长随机，仅 yaml |

`config.yaml.example` 示例：

```yaml
session:
  secret: "change-me-to-a-long-random-string"
  max_age: 604800
  secure: false
  same_site: "lax"
```

注意：`gin-contrib/sessions` 的 `SameSite` 需按该库 API 设置（`http.SameSiteLaxMode` 等）。

### 3.4 禁用账号即时失效

在 `UserService.GetLoginUser` 中，查到用户后：

```text
if user.Status == 0:
  session.Delete(UserLoginState)  # 或 Clear
  session.Save()
  return 未登录 或 专用「账号已禁用」错误
```

因 `AuthCheck` 依赖 `GetLoginUser`，所有受保护接口将一致生效。

可选：为禁用单独业务码（如 `40102`），前端区分文案；P0 也可用现有 `40100` + message。

### 3.5 CORS 配置化

```yaml
cors:
  allow_origins:
    - "http://localhost:5173"
```

中间件逻辑：

1. 读请求 `Origin`  
2. 若在白名单内，回写 `Access-Control-Allow-Origin: <该 Origin>`  
3. 保持 `Allow-Credentials: true`  
4. **禁止**在 Credentials 模式下使用 `*`  

`main.go`：`CORS()` 改为注入 `cfg`（或闭包捕获配置）。

### 3.6 前端未登录统一处理

文件：`web/src/utils/request.ts`

建议响应拦截：

```text
if HTTP status === 401:
  清 loginUser → 跳转 /user/login? 带 from
if body.code === 40100:   # 即便 HTTP 200
  同上（注意避免登录页自身请求死循环）
```

实现注意：

- 跳转前可调用 store 的重置，避免循环依赖可用动态 import 或轻量事件  
- `/user/login`、`/user/register` 相关请求不要触发「再跳登录」  
- 长期：统一后端鉴权失败形态（见 P1 A9）

当前 `AuthCheck` 使用 `http.StatusUnauthorized`（401）+ JSON body；其它业务错误多为 HTTP 200。拦截器需 **两种都兼容**。

---

## 4. P1 改造说明（可选，紧随 P0）

| 项 | 实现要点 |
|----|----------|
| 登录限流 | Redis：`login:fail:{account}` / `login:fail:ip:{ip}`；超限返回 `42900` |
| Logout | `session.Clear()` + `Save()`，确保 Redis 中会话删除 |
| 改密后踢下线 | 改密成功后 `INCR login:ver:{userId}`（见 4.1）或 Clear Session，前端提示重新登录 |
| 响应统一 | `AuthCheck` 改为 HTTP 200 + `code=40100/40101`，与项目其它接口一致 |
| **单 Session（异地挤出）** | **后登录使先登录 Session 失效**；详见 **4.1**（**待迭代，本文仅方案**） |

### 4.1 单 Session：同一账号异地登录挤出（待实现）

#### 4.1.1 产品行为

| 规则 | 说明 |
|------|------|
| 策略 | **后登录生效**：B 地登录成功后，A 地旧 Session 在**下一次请求**鉴权失败 |
| 同浏览器多 Tab | 共用同一 Cookie / Session ID → **不互踢** |
| 实时推送 | **不做**（无 WebSocket）；被挤用户感知依赖下一次 API 调用 |
| 管理员 | 默认与普通用户相同规则；是否豁免可配置 |
| 与禁用 | 禁用仍优先于版本校验（现有 `GetLoginUser` 逻辑） |

#### 4.1.2 技术结论（保持 Redis Session）

- **不改为 JWT**；在现有 `gin-contrib/sessions/redis` 上增加 **登录世代（session version）** 即可。  
- **不推荐**仅维护 `userId → sessionId` 并手动删 Redis Session 键（依赖 store 内部 key 格式，易碎）；世代校验足够让旧 Cookie 失效。

#### 4.1.3 数据与 Session 键

**Redis（推荐，免改表）**

| Key | 含义 |
|-----|------|
| `login:ver:{userId}` | 当前有效登录世代（整数），TTL ≈ `session.max_age` |

**Session（Cookie 对应 Redis Session 值）**

| 键 | 含义 |
|----|------|
| 现有 `userLoginState` | 用户 ID（`common.UserLoginState`） |
| 新增 `userLoginVer` | 登录时写入的世代，与 Redis 一致 |

**配置（`config.yaml.example`）**

```yaml
security:
  single_session: true   # true：后登录挤掉前登录；false：保持现有多 Session 并存
```

#### 4.1.4 流程伪代码

**Login（`UserService.Login`，在 `session.Save` 前）**

```text
if !cfg.Security.SingleSession:
  session.Set(userLoginState, user.ID)
  session.Save()
  return

ver := Redis INCR login:ver:{user.ID}   // 或 SET 带 TTL
session.Set(userLoginState, user.ID)
session.Set(userLoginVer, ver)
session.Save()
```

**GetLoginUser（所有 `AuthCheck` 经此校验）**

```text
user := load user by session userLoginState
if user disabled: ...现有逻辑...

if cfg.Security.SingleSession:
  verSession := session.Get(userLoginVer)
  verCurrent := Redis GET login:ver:{user.ID}
  if verSession != verCurrent:
    session.Delete(userLoginState)
    session.Delete(userLoginVer)
    session.Save()
    return ErrSessionReplaced  // 建议 code 40103，message「账号已在其他设备登录，请重新登录」

return user
```

**Logout**

- 默认：**不** `INCR login:ver`（仅清当前 Session）。  
- 若产品要求「本机登出也踢掉其它端」，登出时对 `userId` 再 `INCR`（需产品确认）。

**UpdatePassword（改密成功）**

- 与 P1「改密后踢下线」合并：`INCR login:ver:{userId}`，使**全部**旧 Session 失效；当前浏览器若需保持登录，改密接口可在同一请求内重新 `Login` 写 Session + 新 ver（或强制前端重新登录，实现更简单）。

#### 4.1.5 错误码（与 P0 区分）

| code | 场景 | HTTP（现状可暂保留 401） |
|------|------|---------------------------|
| `40100` | 未登录 / Session 无效 | 401 或 200 |
| `40101` | 无权限 | 403 或 200 |
| `40102` | 账号已禁用（可选，P0 3.4） | 同左 |
| **`40103`** | **被其它登录挤出** | 同左 |

前端 `web/src/utils/request.ts`：对 `40103` 先 `message.warning` 再跳转登录；`40100` 保持现有文案。

#### 4.1.6 代码落点（实现时自检）

| 层 | 路径 |
|----|------|
| 配置 | `internal/config/config.go` → `SecurityConfig.SingleSession` |
| Redis | 复用 `app.RedisClient` 或薄封装 `internal/store/session_login.go` |
| 业务 | `internal/service/user_service.go` → `Login` / `GetLoginUser` / `UpdatePassword` |
| 常量 | `internal/common/constants.go` → `UserLoginVer`；`error.go` → `ErrSessionReplaced` |
| 前端 | `web/src/utils/request.ts`、`types/api.ts`（可选常量） |
| 文档 | `prd-auth.md` 增加场景「异地登录」 |

**一般不需要改** `middleware/auth.go`、`internal/router/router.go`（仍调 `GetLoginUser`）。

#### 4.1.7 测试要点（实现后）

| 用例 | 期望 |
|------|------|
| A、B 两浏览器同账号登录 | B 登录后，A 下一次受保护接口返回 `40103`（或 401 + 对应 code） |
| A 两 Tab | 同一 Session，均正常 |
| `single_session: false` | 两浏览器可同时在线（回归现状） |
| 改密 | 其它端 Session 失效（若采用 INCR ver） |
| 禁用 | 仍即时失效，不受 ver 影响 |

#### 4.1.8 实施顺序建议

1. 完成 **P0**（密码、禁用、CORS、前端 401）  
2. 配置 + Redis ver + `Login` / `GetLoginUser` + `40103`  
3. 前端拦截文案  
4. 可选：改密 INCR、登录限流  

---


## 5. 建议改动文件清单（自检用）

### 后端

```text
server/internal/router/router.go          # 删加密路由（若仍存在）
server/cmd/server/main.go                 # CORS 注入 cfg
server/internal/middleware/session.go     # Cookie 选项
server/internal/middleware/cors.go        # 白名单
server/internal/middleware/auth.go        # 一般可不动（靠 GetLoginUser）
server/internal/service/user_service.go   # 哈希、禁用、Logout/改密、单 Session（P1）
server/internal/common/constants.go       # 盐仅迁移期保留；UserLoginVer（P1）
server/internal/config/config.go          # Session/CORS/Security 字段
server/config.yaml.example                # 示例配置
server/sql/...                            # 如需加长 userPassword 列
```

### 前端

```text
web/src/utils/request.ts                  # 401 / 40100（待扩展 40102、40103）
web/src/stores/loginUser.ts               # 如需导出 clear / 配合拦截器
```

---

## 6. 配置完整示意

```yaml
cors:
  allow_origins:
    - "http://localhost:5173"

session:
  secret: "change-me-to-a-long-random-string"
  max_age: 604800
  secure: false
  same_site: "lax"

# 可选
security:
  allow_encrypt_api: false
  single_session: true   # P1：后登录挤掉前登录，见 4.1
```

---

## 7. 测试清单（自行验收）

| 类型 | 用例 |
|------|------|
| 注册 | 新用户库中密码为 bcrypt 形态 |
| 登录 | 旧 MD5 用户能登录；再次登录后库中已是 bcrypt |
| 改密 | 新密码 bcrypt；旧密码不能再登 |
| 禁用 | 禁用后带旧 Cookie 调 `/comic/create` 或任意 Auth 接口失败 |
| 登出 | 登出后再调需登录接口失败 |
| 前端 | 手动清 Cookie 或调需登录接口 → 跳转登录页 |
| CORS | 配置第二 Origin，确认仅白名单生效 |
| 加密接口 | `POST /user/encrypt/password` 应 404 或无权限 |
| 单 Session（P1） | 双浏览器同账号：后登者有效，先登者下一请求 `40103` |

本地命令提示：

```bash
cd server && go vet ./...
cd server && go test ./...
npm run build --prefix web
```

---

## 8. 明确不做（或不在本期）

- 整站 JWT  
- 微信 OAuth  
- **Session 设备列表 / 「踢掉指定设备」管理 UI**（P1 单 Session 仅后登录挤前登录，见 §4.1）  
- 被挤下线时的 **WebSocket / SSE 实时通知**（依赖下一次 API）  

---

## 9. 与支付的关系

鉴权 P0 是支付上线的前置条件（Session 安全、禁用生效、前端未登录处理）。  
支付产品/技术方案见：

- [`prd-recharge.md`](./prd-recharge.md)  
- [`tech-recharge.md`](./tech-recharge.md)  
