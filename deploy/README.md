# 宝塔面板部署指南

把 ai-comic-generator 前后端部署到宝塔服务器，之后「提交代码 → 服务器跑一条命令 → 重新上线」。

本目录文件：

| 文件 | 作用 |
|------|------|
| `deploy.sh` | 一键部署脚本：git pull → 构建前端 → 构建后端 → 重启服务 → 健康检查 |
| `ai-comic-server.service` | 后端 systemd 服务单元（崩溃自动重启、开机自启） |
| `nginx.conf.example` | Nginx 站点配置示例（前端静态 + `/api` 反代 + `/static/comics`） |

架构：Nginx（宝塔管理，含 SSL）→ 前端静态产物 `web/dist/` + 反代 `/api` 到 Go 后端（`127.0.0.1:8080`，由 systemd 常驻）。MySQL、Redis 用宝塔自带。

---

## 一、服务器环境准备（一次性）

### 1. 宝塔软件商店安装

- **Nginx**（网站服务）
- **MySQL 5.7+ / 8.0**，建库 `ai_comic_generator`（字符集 `utf8mb4`）
- **Redis**
- 保证服务器能访问外网（调用通义千问 / 混元 API）

### 2. 安装 Go 与 Node（宝塔终端里执行）

宝塔【终端】或 SSH：

```bash
# Go 1.24（按官网最新补丁版）
wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile
go version
go env -w GOPROXY=https://goproxy.cn,direct   # 国内加速

# Node 18+（宝塔【软件商店 → Node 版本管理器】装也行）
node -v && npm -v
```

### 3. 拉取代码

```bash
cd /www/wwwroot
git clone <你的仓库地址> ai-comic-generator
cd ai-comic-generator
```

> 私有仓库需先在服务器配置 SSH deploy key 或用带 token 的 HTTPS 地址。

### 4. 配置后端

```bash
cp server/config.yaml.example server/config.yaml
vi server/config.yaml
```

必改项：

- `database`：host `127.0.0.1`、宝塔建的库名/账号/密码
- `redis`：host `127.0.0.1`
- `session.secret`：换成随机长串
- `ai.dashscope.api_key`：填通义千问 key（**不填则整个漫画模块被禁用**，仅用户功能可用）
- 需要时开 `ai.hunyuan.enabled` / `wechat.enabled` 并填密钥

> `config.yaml` 已在 `.gitignore`，不会入库。也可用环境变量覆盖敏感项（见 systemd 里的 `EnvironmentFile`，字段见 `config.go` 的 `applyEnvOverrides`：`DB_PASSWORD`、`DASHSCOPE_API_KEY` 等）。

### 5. 注册 systemd 服务

```bash
# 按实际路径改 .service 里的 WorkingDirectory / ExecStart / User
cp deploy/ai-comic-server.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable ai-comic-server      # 开机自启
```

首次先手动构建一次二进制再启动（或直接跑下面的 deploy.sh）：

```bash
bash deploy/deploy.sh
```

### 6. 配置 Nginx 站点

1. 宝塔【网站 → 添加站点】，绑定域名，**不建数据库、PHP 选纯静态**
2. 【设置 → 配置文件】，参照 `deploy/nginx.conf.example` 把 `location /api/`、`/static/comics/`、SPA 回退段合并进去，`root` 指向 `web/dist`
3. 【SSL】申请 Let's Encrypt 证书，开启强制 HTTPS
4. 保存后宝塔自动 reload Nginx

前端生产环境走同域名，`web/.env.production` 里 `VITE_API_BASE_URL=/api`（相对路径，无需改），由 Nginx 反代到后端，天然规避跨域。

---

## 二、日常部署（每次提交代码后）

本地正常提交推送后，在服务器项目根目录执行：

```bash
cd /www/wwwroot/ai-comic-generator
bash deploy/deploy.sh
```

脚本会：拉最新代码 → `npm ci && npm run build` 出前端 → `go build` 出后端 → `systemctl restart` → 健康检查 `/api/health`。成功打印 ✅，失败提示查日志。

常用参数：

```bash
bash deploy/deploy.sh --skip-web      # 只动了后端
bash deploy/deploy.sh --skip-server   # 只动了前端
bash deploy/deploy.sh --no-pull       # 用服务器现有代码构建
```

### 免 SSH：宝塔计划任务一键部署

宝塔【计划任务 → 添加任务】：

- 类型：Shell 脚本
- 执行周期：不定时（或按需手动执行）
- 脚本内容：

```bash
cd /www/wwwroot/ai-comic-generator && bash deploy/deploy.sh >> /www/wwwlogs/ai-comic-deploy.log 2>&1
```

之后每次部署在宝塔面板点一下「执行」即可，日志落到 `/www/wwwlogs/ai-comic-deploy.log`。

---

## 三、运维排查

```bash
systemctl status ai-comic-server              # 服务状态
journalctl -u ai-comic-server -n 100 --no-pager   # 最近 100 行日志
journalctl -u ai-comic-server -f              # 实时日志
curl http://127.0.0.1:8080/api/health         # 后端健康检查
```

常见问题：

- **健康检查失败**：先看 journalctl，多为 `config.yaml` 数据库/Redis 连不上（宝塔里确认服务已启动、密码正确、host 用 `127.0.0.1`）。
- **`no route to host` 连 MySQL**：宝塔防火墙/安全组拦截，放行本机 3306/6379 或确认 bind 到 `127.0.0.1`。
- **前端 404 / 刷新白屏**：Nginx 缺 SPA 回退 `try_files ... /index.html`。
- **漫画图片不显示**：检查 Nginx `/static/comics/` 反代，以及 `storage.base_path` 目录（`server/data/comics`）对运行用户可写。
- **构建慢/超时**：确认 `GOPROXY` 已设国内源；前端 `npm ci` 慢可换 npm 镜像。

---

## 四、后续进阶（可选）

起步用手动 / 计划任务足够。想做到「git push 后全自动上线」，可加：

- **Git Webhook**：宝塔【软件商店】装 Git webhook 插件，或写个极简 hook 服务监听 push 事件后调用 `deploy.sh`。注意校验签名、限制来源 IP。
- **CI 构建**：在 GitHub Actions 里构建产物，用 rsync/scp 传到服务器，服务器只负责 `systemctl restart`，省去服务器装 Go/Node 工具链。
