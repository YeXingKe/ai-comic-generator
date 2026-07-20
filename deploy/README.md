# 宝塔面板部署指南（新手向）

把 ai-comic-generator 前后端部署到宝塔服务器，实现「提交代码 → 服务器跑一条命令 → 重新上线」。

**这份文档假设你是第一次做服务器部署**，所以步骤写得比较细，每一步都会说明「做完之后应该看到什么」，方便你判断是不是走对了。如果你已经很熟悉宝塔/Nginx/systemd，可以直接跳到「二、日常部署」。

---

## 0. 先搞懂几个名词（跳过也不影响操作，卡住时回来看）

| 名词 | 一句话解释 |
|------|-----------|
| **宝塔面板** | 一个网页版的服务器管理工具，装好之后不用敲太多命令，点点鼠标就能装 Nginx/MySQL、管网站、开终端 |
| **Nginx** | 挂在最外层接收所有请求的网关。它做两件事：把 `web/dist` 里的前端文件直接发给浏览器；把 `/api` 开头的请求转发（反向代理）给 Go 后端 |
| **反向代理** | 浏览器只认识 Nginx 的域名，并不知道背后还有个跑在 `127.0.0.1:8080` 的 Go 程序。Nginx 负责把请求转给它，再把结果转回浏览器 |
| **systemd 服务** | Linux 用来「常驻运行一个程序」的机制。把 Go 后端注册成 systemd 服务后，它就能开机自启、崩溃自动重启，不用你手动 `nohup` 挂着 |
| **反代 / 静态资源** | 前端页面（HTML/JS/CSS）叫静态资源，直接读文件返回；`/api` 请求要经过后端程序处理，叫反代 |

架构图（文字版）：

```
浏览器
  │ https://你的域名
  ▼
Nginx（宝塔管理，含 SSL）
  ├─ 静态文件 → web/dist/（前端产物）
  └─ /api、/static/comics → 反代 127.0.0.1:8080（Go 后端，systemd 常驻）
                                  │
                                  ├─ MySQL（宝塔自带，本机 3306）
                                  └─ Redis（宝塔自带，本机 6379）
```

本目录文件：

| 文件 | 作用 |
|------|------|
| `deploy.sh` | 一键部署脚本：git pull → 构建前端 → 构建后端 → 重启服务 → 健康检查 |
| `ai-comic-server.service` | 后端 systemd 服务单元（崩溃自动重启、开机自启） |
| `nginx.conf.example` | Nginx 站点配置示例（前端静态 + `/api` 反代 + `/static/comics`） |

## 部署全流程一览（先建立心理预期）

首次部署大致要走这几步，一次性的，做完之后基本不用再碰：

1. **装环境**：宝塔软件商店装 Nginx / MySQL / Redis，终端里装 Go + Node（10~20 分钟）
2. **拉代码 + 配置**：`git clone` 项目，复制并填写 `config.yaml`（5 分钟）
3. **注册 systemd 服务 + 首次构建**：跑一次 `deploy.sh`，后端跑起来（几分钟，看服务器配置和网速）
4. **配置 Nginx 站点**：绑定域名、加反代规则、申请 SSL（5~10 分钟）
5. **验证**：浏览器打开域名，能看到首页、能注册登录（几分钟）

之后每次改了代码，只用重复「二、日常部署」里的一条命令。

---

## 一、服务器环境准备（一次性）

### 1. 宝塔软件商店安装

登录宝塔面板 → 左侧【软件商店】，搜索并安装：

- **Nginx**（网站服务，选默认稳定版即可）
- **MySQL 5.7+ / 8.0**（装好后在【数据库】里新建库 `ai_comic_generator`，字符集选 `utf8mb4`，记下你设的用户名/密码）
- **Redis**（装好默认就是 `127.0.0.1:6379`，不用额外配置）

装完之后到【软件商店 → 已安装】确认三个都是「运行中」状态。

> 服务器要能访问外网，因为后端要调用通义千问 / 混元的 API。如果你的服务器在内网或者有出站白名单限制，先确认能 `curl https://dashscope.aliyuncs.com` 通。

### 2. 安装 Go 与 Node（宝塔终端里执行）

宝塔左侧【终端】（或用 SSH 工具登录服务器）：

```bash
# Go 1.24（版本号可去 https://go.dev/dl/ 核对最新补丁版）
wget https://mirrors.aliyun.com/golang/go1.24.4.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile
go version   # 应输出 go version go1.24.4 linux/amd64
go env -w GOPROXY=https://goproxy.cn,direct   # 国内环境必设，否则 go build 下载依赖大概率超时

# Node 18+（也可以在宝塔【软件商店 → Node 版本管理器】里装，图形化更省心）
node -v && npm -v
```

> 踩坑提醒：`echo ... >> /etc/profile` 只对新打开的终端生效。如果你是在宝塔网页终端里连续操作，执行完 `source /etc/profile` 后 `go version` 还是报 `command not found`，就关掉终端窗口重新打开一个再试。

### 3. 拉取代码

```bash
cd /www/wwwroot
git clone <你的仓库地址> ai-comic-generator
cd ai-comic-generator
```

> 私有仓库拉不下来，通常是两种情况：没配 SSH deploy key，或者 HTTPS 地址里没带访问 token。二选一配好即可，跟部署本身无关，属于 Git 平台（GitHub/GitLab）账号权限设置。

### 4. 配置后端

```bash
cp server/config.yaml.example server/config.yaml
vi server/config.yaml   # 不熟悉 vi 也可以在宝塔【文件】里直接图形化编辑这个文件
```

必改项：

| 配置项 | 填什么 | 不填会怎样 |
|--------|--------|-----------|
| `database.host/user/password/name` | 第 1 步在宝塔建的库信息，host 填 `127.0.0.1` | 后端启动直接报错退出 |
| `redis.host` | `127.0.0.1` | 同上 |
| `session.secret` | 换成一段随机字符串（比如 `openssl rand -hex 32` 生成） | 用默认值也能跑，但登录态加密强度低，生产环境务必换 |
| `ai.dashscope.api_key` | 你的通义千问 key | **不填则整个漫画生成模块被禁用**（用户注册/登录等基础功能仍可用，但生成不了漫画） |
| `ai.hunyuan.enabled` / `wechat.enabled` | 需要生图/发公众号时改 `true` 并填对应密钥 | 保持 `false` 时走占位图/草稿降级，不影响其余流程 |

> `config.yaml` 已经在 `.gitignore` 里，不会被提交到仓库，放心填真实密钥。也可以不改文件，改用环境变量覆盖敏感项（对应字段见 `server/internal/config/config.go` 的 `applyEnvOverrides`：`DB_HOST`、`DB_PORT`、`DB_NAME`、`DB_USER`、`DB_PASSWORD`、`REDIS_HOST`、`REDIS_PORT`、`REDIS_PASSWORD`、`DASHSCOPE_API_KEY`）。环境变量的用法见下面 systemd 部分的 `EnvironmentFile`。

填完之后可以自查一遍：`cat server/config.yaml`（或用宝塔文件管理打开看看），确认密码、api_key 那几行不是空的、没有多余引号错位。

### 5. 注册 systemd 服务

```bash
# 按实际路径改 .service 里的 WorkingDirectory / ExecStart / User（如果就是按本文档的路径部署，通常不用改）
cp deploy/ai-comic-server.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable ai-comic-server      # 开机自启，执行后不会立刻启动服务，只是登记
```

首次部署还没有编译出后端二进制，直接跑一次部署脚本，它会帮你构建 + 启动：

```bash
bash deploy/deploy.sh
```

**这一步做完之后应该看到什么**：终端最后一行打印 `[deploy] 部署成功，后端已就绪 ✅`。如果卡住或报错，别急着往下走，先看本文档「三、运维排查」定位问题，后端没跑起来的话后面配 Nginx 也访问不通。

想再确认一次服务真的在跑：

```bash
systemctl status ai-comic-server
```

看到 `Active: active (running)` 就是正常的；如果是 `failed` 或 `inactive`，跳到排查章节。

### 6. 配置 Nginx 站点

1. 宝塔【网站 → 添加站点】，绑定你的域名，**不勾选创建数据库、PHP 版本选「纯静态」**（因为静态文件由 Nginx 直接发，动态逻辑全在 Go 后端，不需要 PHP）
2. 站点列表点进去 →【设置 → 配置文件】，把 `deploy/nginx.conf.example` 里的 `location /api/`、`location /static/comics/`、SPA 回退这几段，合并进宝塔已经生成的 `server { ... }` 块里；同时把 `root` 改成指向 `/www/wwwroot/ai-comic-generator/web/dist`
3. 【设置 → SSL】申请 Let's Encrypt 免费证书，勾选「强制 HTTPS」
4. 点保存，宝塔会自动帮你 reload Nginx（不需要额外敲命令）

前端在生产环境走的是同域名，`web/.env.production` 里已经配好 `VITE_API_BASE_URL=/api`（相对路径），构建时会打进产物里，浏览器请求 `/api/xxx` 时自然会被同域名下的 Nginx 反代到后端，天然避开了跨域问题，不需要额外配 CORS 白名单。

### 7. 验证部署成功

三个检查，任何一步失败都说明有环节没配对，回上面对应小节排查：

```bash
# 1. 后端自身健康检查（服务器本机执行）
curl http://127.0.0.1:8080/api/health
# 期望：返回 {"code":0,"data":{"status":"ok"},"message":"ok"}，不是连接拒绝/超时

# 2. 经过 Nginx 反代之后是否也通（服务器本机执行，替换成你的域名）
curl https://your-domain.com/api/health
# 期望：和上面结果一致

# 3. 浏览器打开 https://your-domain.com
# 期望：能看到首页；试着注册/登录一个账号，走一遍基础流程
```

三步都通过，首次部署就算完成了。

---

## 二、日常部署（每次提交代码后）

首次部署做完之后，以后每次改完代码，流程就简单了：

1. 本地开发机：正常 `git add / commit / push`
2. 服务器：SSH 或宝塔终端登录，执行：

```bash
cd /www/wwwroot/ai-comic-generator
bash deploy/deploy.sh
```

脚本会自动做这几件事，不需要你手动分步执行：

1. `git fetch` + `git reset --hard origin/main`，把服务器代码强制对齐远端最新分支（**注意**：这一步会丢弃服务器本地任何改动，所以线上代码只应该通过 git push 更新，别直接在服务器上改文件）
2. `npm ci && npm run build` 构建前端产物到 `web/dist/`
3. `go build` 编译后端到 `server/bin/server`
4. `systemctl restart ai-comic-server` 重启后端进程
5. 反复 curl `/api/health`（最多等 10 秒）确认新版本已经跑起来

看到 `[deploy] 部署成功，后端已就绪 ✅` 说明这次更新上线成功。如果中途某一步报错（比如前端 `npm run build` 类型检查没过），脚本会在那一步直接停下并退出，不会继续往后执行——此时线上还在运行上一次成功部署的旧版本，不会因为这次失败变得不可用，你可以慢慢排查再重跑。

常用参数（只改了一端时用，能省点构建时间）：

```bash
bash deploy/deploy.sh --skip-web      # 只改了后端，跳过前端构建
bash deploy/deploy.sh --skip-server   # 只改了前端，跳过后端构建 + 重启
bash deploy/deploy.sh --no-pull       # 跳过 git pull，用服务器上现有代码构建（调试脚本本身才会用到，日常基本用不到）
```

### 免 SSH：宝塔计划任务一键部署

如果不想每次都开终端敲命令，可以让宝塔帮你保存这条命令：

宝塔【计划任务 → 添加任务】：

- 任务类型：Shell 脚本
- 执行周期：选「手动执行」（不需要定时，代码更新是不定期的事）
- 脚本内容：

```bash
cd /www/wwwroot/ai-comic-generator && bash deploy/deploy.sh >> /www/wwwlogs/ai-comic-deploy.log 2>&1
```

保存后，以后每次部署只需要在宝塔面板【计划任务】列表里点一下「执行」，不用登录终端。执行日志会追加写到 `/www/wwwlogs/ai-comic-deploy.log`，想确认这次有没有成功，打开这个文件看最后几行就行。

---

## 三、运维排查

先记住这四条命令，绝大部分问题靠它们就能定位：

```bash
systemctl status ai-comic-server                  # 服务当前状态（运行中/已停止/崩溃循环）
journalctl -u ai-comic-server -n 100 --no-pager    # 看最近 100 行日志，报错信息基本都在这里
journalctl -u ai-comic-server -f                   # 实时跟踪日志（Ctrl+C 退出），适合边操作边看报错
curl http://127.0.0.1:8080/api/health              # 后端本机健康检查，最基础的连通性验证
```

排查思路：先 `curl /api/health`，通就说明后端本身没问题，问题在 Nginx 那层；不通就看 `journalctl` 里的报错信息，日志里一般会直接写出是数据库连不上、配置读取失败还是端口被占用。

常见问题：

- **健康检查失败（curl 本机 8080 都不通）**：先看 `journalctl -u ai-comic-server -n 100`，多数是 `config.yaml` 里数据库/Redis 连不上——确认宝塔里 MySQL、Redis 服务已经是运行中状态，密码填对了，`host` 用 `127.0.0.1` 而不是外网 IP 或域名。
- **报错 `no route to host` 或连 MySQL/Redis 超时**：一般是防火墙/安全组拦截了本机端口，去宝塔【安全】和云厂商的安全组里放行 3306（MySQL）、6379（Redis），或者确认这两个服务确实 bind 在 `127.0.0.1` 上而不是被限制访问。
- **浏览器打开域名是空白/404，刷新非首页路径直接 404**：Nginx 缺 SPA 回退配置，检查站点配置文件里有没有 `location / { try_files $uri $uri/ /index.html; }` 这一段（对照 `deploy/nginx.conf.example`）。
- **能登录，但生成漫画一直失败或没反应**：先确认 `server/config.yaml` 里 `ai.dashscope.api_key` 填了且没过期/欠费；再看 `journalctl` 里有没有调用通义千问接口报错的日志，常见是 key 无效或服务器出不了公网。
- **漫画图片显示不出来（一直转圈或裂图）**：检查 Nginx 的 `/static/comics/` 反代规则是不是也配了；再确认 `storage.base_path`（默认 `server/data/comics`）这个目录存在，并且对运行后端的用户（`ai-comic-server.service` 里配的 `User=www`）有读写权限，权限不对可以 `chown -R www:www server/data`。
- **部署脚本卡在 `npm ci` 或 `go mod download` 很久没反应**：大概率是网络问题。前端确认 npm 源（可换成淘宝镜像 `npm config set registry https://registry.npmmirror.com`），后端确认第 2 步的 `GOPROXY` 已经设成 `https://goproxy.cn,direct`。
- **`systemctl restart` 之后过几秒服务又变回 `failed`**：说明程序启动后很快崩溃，通常还是配置问题（比如 `config.yaml` 格式错误、字段缩进错了）。看 journalctl 日志里进程退出前打的最后几行，一般会直接指出是哪个配置项解析失败。

### 上线前最后检查一遍（安全相关）

正式对外提供服务前，建议再确认一遍：

- `server/config.yaml` 里的 `session.secret` 已经换成随机字符串，不是示例里的默认值
- 默认管理员账号 `admin` / `admin123456`（首次启动自动创建）已经登录后台改成强密码
- `server/config.yaml` 没有被误提交进 git（`git status` 确认它不在待提交列表里）
- Nginx 已经强制 HTTPS，避免账号密码用明文 HTTP 传输
- 如果暂时不打算开放某个功能（比如公众号发布），保持对应 `enabled: false`，别为了「以防以后要用」提前开启并填入真实密钥

---

## 四、后续进阶（可选）

起步阶段用「手动跑一条命令」或「宝塔计划任务点一下」完全够用，不用一开始就追求全自动。等流程跑顺了、部署频率上来了，可以考虑：

- **Git Webhook**：宝塔【软件商店】装 Git webhook 插件，或者自己写一个极简的 hook 服务监听仓库的 push 事件，收到通知后自动调用 `deploy.sh`，实现「git push 后全自动上线」。注意要校验 webhook 签名、限制来源 IP，避免被人伪造请求触发部署。
- **CI 构建**：在 GitHub Actions 里完成前后端构建，把产物用 rsync/scp 传到服务器，服务器只负责 `systemctl restart`，好处是服务器上不用再装 Go/Node 工具链，构建也不占用服务器资源。

这两项都不是必需的，是团队规模变大、部署更频繁之后才值得投入的优化，个人项目或者小团队没必要一上来就折腾。
