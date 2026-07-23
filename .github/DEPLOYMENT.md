# GitHub Actions 自动化部署指南

本文档说明如何配置 GitHub Actions 实现自动化部署，使得每次 `git push` 到 `main` 分支时，都会自动在云端构建并部署到服务器。

## 概述

### 工作流程

```
git push main
    ↓
GitHub Actions 触发
    ├─ 拉取代码
    ├─ 安装 Node.js + Go 环境
    ├─ 构建前端 (npm build → web/dist/)
    ├─ 构建后端 (go build → server/bin/server)
    ├─ 上传前端产物到服务器
    ├─ 上传后端二进制到服务器
    ├─ SSH 连接服务器重启 systemd 服务
    └─ 健康检查 (/api/health)
```

### 优势

- **无需本地工具链** — 开发机只需要 Git，无需装 Go、Node
- **无需服务器工具链** — 服务器只需要 systemd 和 rsync，无需装 Go、Node、npm
- **快速部署** — 云端并行构建，上传只需秒级文件传输，比「在服务器上构建」快 10 倍
- **隔离构建** — 每次构建都是干净环境，避免本地/服务器依赖版本差异导致的问题
- **可审计** — 每次部署的日志都在 GitHub Actions 里永久保存

---

## 一、前置准备（一次性）

### 1.1 生成 SSH 密钥对

**在本地开发机执行**（不要在服务器上执行）：

```bash
# 生成一对 RSA 密钥，用于 GitHub Actions 认证服务器
# -f 指定保存路径，建议单独给部署流程用一个密钥，与个人密钥分开
ssh-keygen -t rsa -b 4096 -f ~/.ssh/ai-comic-deploy

# 提示 "Enter passphrase" 时，直接按 Enter（不输入密码）
# 因为 GitHub Actions 需要无密码的密钥才能自动认证
```

执行后会生成两个文件：
- `~/.ssh/ai-comic-deploy` — **私钥**（放到 GitHub Secrets，严格保密）
- `~/.ssh/ai-comic-deploy.pub` — **公钥**（放到服务器，可公开）

### 1.2 授权服务器

**登录服务器执行**（用你常用的账号，比如宝塔终端）：

```bash
# 1. 查看本机的公钥
cat ~/.ssh/ai-comic-deploy.pub

# 2. 在服务器的 www 用户下授权这个公钥
# 方式 A：直接追加到 www 用户的 authorized_keys（推荐）
echo "<粘贴上面复制的公钥内容>" | sudo tee -a /home/www/.ssh/authorized_keys

# 方式 B：如果 www 用户没有 .ssh 目录，先创建
sudo -u www mkdir -p /home/www/.ssh
sudo -u www touch /home/www/.ssh/authorized_keys
sudo chmod 700 /home/www/.ssh
sudo chmod 600 /home/www/.ssh/authorized_keys

# 3. 验证权限
sudo ls -la /home/www/.ssh/authorized_keys
# 应该输出：-rw------- 1 www www ...
```

**验证 SSH 连接**（在本地执行，确保能免密登录）：

```bash
ssh -i ~/.ssh/ai-comic-deploy www@your-server-ip "echo 'SSH 连接成功'"
# 如果输出 "SSH 连接成功"，说明配置正确
```

### 1.3 在 GitHub 配置 Secrets

打开你的 GitHub 仓库，进入 **Settings → Secrets and variables → Actions**，添加以下三个 Secrets：

#### Secret 1: `DEPLOY_SSH_KEY`

- **值**：`~/.ssh/ai-comic-deploy` 文件的完整内容
- **获取方式**：
  ```bash
  cat ~/.ssh/ai-comic-deploy
  # 复制整个输出，包括 "-----BEGIN OPENSSH PRIVATE KEY-----" 和 "-----END OPENSSH PRIVATE KEY-----"
  ```

#### Secret 2: `DEPLOY_SERVER_HOST`

- **值**：你的服务器 IP 或域名
- **例子**：`123.45.67.89` 或 `server.example.com`

#### Secret 3: `DEPLOY_USER`

- **值**：`www`（你部署用的用户）

---

## 二、工作流程详解

### workflow 文件位置

`.github/workflows/deploy.yml`

### 各步骤说明

| 步骤 | 作用 | 失败原因 |
|------|------|--------|
| **Checkout code** | 拉取 GitHub 上的最新代码 | 网络问题 |
| **Setup Node.js** | 安装 Node 18 环境 | 官方源问题 |
| **Setup Go** | 安装 Go 1.24 环境 | 官方源问题 |
| **Build web** | 前端编译：`npm ci && npm run build` | TypeScript 类型检查失败、npm 依赖下载失败 |
| **Build server** | 后端编译：`go build` | Go 编译失败、依赖下载失败 |
| **Setup SSH** | 配置 SSH 密钥和已知主机 | 密钥不匹配 |
| **Upload web artifacts** | rsync 上传前端产物 | 目标路径权限不足、网络连接问题 |
| **Upload server binary** | rsync 上传后端二进制 | 同上 |
| **Restart service** | SSH 重启 systemd 服务 | 服务不存在、权限不足、systemd 异常 |
| **Health check** | curl `/api/health` 验证后端 | 后端启动失败、配置错误 |

### 构建环境

- **操作系统**：Ubuntu 22.04 LTS（官方 runner）
- **Node.js 版本**：18（与项目要求一致）
- **Go 版本**：1.24（与项目要求一致）

---

## 三、日常部署流程

### 正常部署

1. **本地开发机**：正常开发、commit、push

   ```bash
   git add .
   git commit -m "feat: add new feature"
   git push origin main
   ```

2. **GitHub Actions 自动部署**：
   - 打开 GitHub 仓库 → **Actions** 标签
   - 看到 "Build & Deploy" workflow 在运行
   - 等待完成（通常 3~5 分钟）
   - 看到 ✅ 绿色对勾说明部署成功

3. **验证上线**：
   ```bash
   # 在本地开发机或任何有网络的地方验证
   curl https://your-domain.com/api/health
   ```

### 快速查看部署日志

1. GitHub 仓库 → **Actions** → 点进最新的 workflow 运行
2. 左侧列表选择具体的 job（"build-and-deploy"）
3. 右侧看每个 step 的输出日志

### 部署失败怎么办

**第一步**：看 GitHub Actions 日志定位是哪个 step 失败

- **Build web 失败** — 通常是前端 TypeScript 编译错误，修复代码后重新 push
- **Build server 失败** — 通常是后端代码编译错误或 Go 依赖问题，修复后重新 push
- **Upload 失败** — 检查服务器上的文件权限，或者 rsync 网络问题（重新 push 重试）
- **Restart service 失败** — SSH 连接成功但服务重启失败，登录服务器手动排查

**第二步**：登录服务器确认当前状态

```bash
ssh www@your-server-ip

# 检查服务状态
systemctl status ai-comic-server

# 看最近的错误日志
journalctl -u ai-comic-server -n 50 --no-pager

# 手动重启尝试
sudo systemctl restart ai-comic-server
```

---

## 四、服务器端配置清单

部署前请确认服务器上已经准备好这些：

### 必需

- [ ] 已安装 systemd 服务文件 `deploy/ai-comic-server.service`
  ```bash
  sudo cp deploy/ai-comic-server.service /etc/systemd/system/
  sudo systemctl daemon-reload
  sudo systemctl enable ai-comic-server
  ```

- [ ] 已配置 `server/config.yaml`
  ```bash
  cp server/config.yaml.example server/config.yaml
  # 填写数据库、Redis、API 密钥等配置
  ```

- [ ] `www` 用户能写入这些目录（部署时需要覆盖文件）：
  ```bash
  ls -la /www/wwwroot/ai-comic-generator/web/
  ls -la /www/wwwroot/ai-comic-generator/server/bin/
  # 应该输出 www www 作为所有者
  ```

- [ ] Nginx 配置指向 `web/dist` 目录并反代 `/api` 到 `127.0.0.1:8080`

### 可选

- [ ] 配置防火墙允许 SSH 连接（云厂商安全组）
- [ ] 监控告警：服务异常时收到通知

---

## 五、故障排查

### 常见问题

#### Q: "Upload web artifacts" 失败，提示 "Permission denied"

**原因**：服务器上 `/www/wwwroot/ai-comic-generator/web/dist/` 目录所有者不是 `www`

**解决**：
```bash
sudo chown -R www:www /www/wwwroot/ai-comic-generator/web/
```

---

#### Q: "Restart service" 失败，提示 "sudo: command not found"

**原因**：`www` 用户没有 sudo 权限

**解决**：改 systemd service 文件，用 `root` 用户跑，或者给 `www` 配置 sudo 免密

```bash
sudo visudo
# 在末尾加一行
www ALL=(ALL) NOPASSWD: /bin/systemctl
```

---

#### Q: SSH 连接失败，提示 "Permission denied (publickey)"

**原因**：服务器上的公钥没有正确授权

**解决**：
```bash
# 服务器上重新授权
sudo -u www mkdir -p /home/www/.ssh
cat ~/.ssh/ai-comic-deploy.pub | sudo tee -a /home/www/.ssh/authorized_keys
sudo chown -R www:www /home/www/.ssh
sudo chmod 700 /home/www/.ssh
sudo chmod 600 /home/www/.ssh/authorized_keys
```

然后在本地重新验证：
```bash
ssh -i ~/.ssh/ai-comic-deploy www@your-server-ip "echo test"
```

---

#### Q: "Health check" 失败，后端起不来

**原因**：多数是 `config.yaml` 配置错误或数据库连不上

**解决**：登录服务器看日志
```bash
ssh www@your-server-ip
sudo journalctl -u ai-comic-server -n 100 --no-pager
# 日志里通常会直接指出是哪个配置项有问题
```

---

#### Q: 怎么回滚到上个版本？

**选项 1**：本地 git 回滚后再 push
```bash
git revert HEAD           # 创建一个新的回滚 commit
git push origin main      # GitHub Actions 自动部署回滚版本
```

**选项 2**：手动在服务器上回滚（不推荐，容易出问题）
```bash
ssh www@your-server-ip
cd /www/wwwroot/ai-comic-generator
# 手工恢复前个版本的 web/dist/ 和 server/bin/server
sudo systemctl restart ai-comic-server
```

---

## 六、性能优化

### 如果部署速度太慢

可以在 workflow 里添加依赖缓存，加快构建：

```yaml
# 在 "Build web" 之前加上
- name: Cache Node dependencies
  uses: actions/cache@v3
  with:
    path: web/node_modules
    key: node-${{ hashFiles('web/package-lock.json') }}
    restore-keys: node-

# 在 "Build server" 之前加上
- name: Cache Go modules
  uses: actions/cache@v3
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: go-${{ hashFiles('server/go.sum') }}
    restore-keys: go-
```

添加后，第一次构建要下载依赖，后续构建会直接从缓存取，快 50~70%。

---

## 七、安全建议

- ✅ SSH 私钥只在 GitHub Secrets 里存放，本地不上传
- ✅ `config.yaml` 继续保持在 `.gitignore` 里，不提交密钥到 Git
- ✅ 定期轮换 SSH 密钥（比如每 90 天）
- ✅ 如果密钥泄露，立即在服务器上删除对应公钥，GitHub 上重新生成
- ✅ 限制 workflow 只在 `main` 分支运行，避免开发分支误触发部署

---

## 八、进阶用法

### 只部署前端（跳过后端构建）

如果只改了前端代码，可以加参数跳过后端构建加快速度。修改 workflow：

```yaml
# Build server 步骤改成
- name: Build server
  if: false  # 暂时禁用
  run: ...
```

### 手动触发部署

改 workflow 的 `on:` 部分：

```yaml
on:
  push:
    branches:
      - main
  workflow_dispatch:  # 添加这一行，允许在 GitHub UI 里手动触发
```

然后在 **Actions** 标签里可以手动点「Run workflow」。

### 定时部署

```yaml
on:
  schedule:
    - cron: '0 2 * * *'  # 每天凌晨 2 点自动部署
```

---

## 九、反馈与迭代

如果部署过程中发现问题或想优化工作流，可以：

1. **修改 workflow 文件**：直接编辑 `.github/workflows/deploy.yml`，改完 push 到 `main` 就会用新版本
2. **查看日志**：GitHub Actions 的日志永久保存，便于事后分析
3. **记录 Secrets 变更**：如果更新了服务器 IP、SSH 密钥等，记得同步更新 GitHub Secrets

---

**如有问题，先看「五、故障排查」，通常能找到答案。**
