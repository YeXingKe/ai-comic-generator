# 技术方案：登录鉴权增强

> 版本：v1.0  
> 日期：2026-08-14  
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

## 2. 现状问题与代码锚点

| 问题 | 位置 | 风险 |
|------|------|------|
| MD5 + 固定盐 `mason` | `server/internal/service/user_service.go` → `encryptPassword` | 密码易被撞 |
| 盐常量 | `server/internal/common/constants.go` → `PasswordSalt` | 与 SQL 种子绑定 |
| 公开加密接口 | `server/cmd/server/main.go` → `POST /user/encrypt/password` | 泄露算法与盐 |
| Cookie 缺 Secure/SameSite | `server/internal/middleware/session.go` | CSRF / 传输风险 |
| 禁用仅 Login 检查 | `UserService.Login` vs `GetLoginUser` | 禁用后旧 Session 仍可用 |
| CORS 写死本地 | `server/internal/middleware/cors.go` | 生产跨域失败 |
| 前端无统一 401 | `web/src/utils/request.ts` | 半登录态 |
| AuthCheck | `server/internal/middleware/auth.go` | 角色校验逻辑保留 |
| 登录态 store | `web/src/stores/loginUser.ts` | 需配合清态 |

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
| 改密后踢下线 | 改密成功后 Clear Session，前端提示重新登录 |
| 响应统一 | `AuthCheck` 改为 HTTP 200 + `code=40100/40101`，与项目其它接口一致 |

---

## 5. 建议改动文件清单（自检用）

### 后端

```text
server/cmd/server/main.go                 # 删加密路由；CORS 注入 cfg
server/internal/middleware/session.go     # Cookie 选项
server/internal/middleware/cors.go        # 白名单
server/internal/middleware/auth.go        # 一般可不动（靠 GetLoginUser）
server/internal/service/user_service.go   # 哈希、禁用、Logout/改密
server/internal/common/constants.go       # 盐仅迁移期保留
server/internal/config/config.go          # Session/CORS 字段
server/config.yaml.example                # 示例配置
server/sql/...                            # 如需加长 userPassword 列
```

### 前端

```text
web/src/utils/request.ts                  # 401 / 40100
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

本地命令提示：

```bash
cd server && go vet ./...
cd server && go test ./...
npm run build --prefix web
```

---

## 8. 明确不做

- 整站 JWT  
- 微信 OAuth  
- 完整多端 Session 管理 UI（「踢掉其它设备」可后续）  

---

## 9. 与支付的关系

鉴权 P0 是支付上线的前置条件（Session 安全、禁用生效、前端未登录处理）。  
支付产品/技术方案见：

- [`prd-recharge.md`](./prd-recharge.md)  
- [`tech-recharge.md`](./tech-recharge.md)  
