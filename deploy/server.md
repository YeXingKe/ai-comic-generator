# 服务器部署指南

## 用户规划

| 用户      | 用途                       | 登录方式 |
| --------- | -------------------------- | -------- |
| `deploy`  | CI/CD 上传二进制、重启服务 | SSH 密钥 |
| `appuser` | 运行应用进程               | 不可登录 |
| `root`    | 系统维护                   | 禁用 SSH |

---

## 一、创建用户

```bash
# 创建 deploy 用户（负责部署操作，可 SSH 登录）
sudo useradd -m -s /bin/bash deploy

# 创建 appuser（负责运行应用，无法登录）
sudo useradd -r -s /sbin/nologin appuser
```

---

## 二、配置 SSH 密钥登录

### 本地生成密钥对

```powershell
# PowerShell
ssh-keygen -t ed25519 -C "deploy@comic.wszhu.top" -f C:\Users\Administrator\.ssh\deploy_key

# Git Bash
ssh-keygen -t ed25519 -C "deploy@comic.wszhu.top" -f /c/Users/Administrator/.ssh/deploy_key
```

生成两个文件：

- `deploy_key` — 私钥，只留在本地，不能泄露
- `deploy_key.pub` — 公钥，上传到服务器

### 查看公钥内容

```powershell
# PowerShell
type C:\Users\Administrator\.ssh\deploy_key.pub
```

### 将公钥上传到服务器

在服务器上执行（用 root 或已有权限的用户登录）：

```bash
sudo mkdir -p /home/deploy/.ssh
sudo chmod 700 /home/deploy/.ssh
echo "ssh-ed25519 AAAA...（公钥内容）" | sudo tee /home/deploy/.ssh/authorized_keys
sudo chmod 600 /home/deploy/.ssh/authorized_keys
sudo chown -R deploy:deploy /home/deploy/.ssh
```

### 本地测试 SSH 登录

```powershell
# PowerShell
ssh -i C:\Users\Administrator\.ssh\deploy_key deploy@comic.wszhu.top

# Git Bash
ssh -i /c/Users/Administrator/.ssh/deploy_key deploy@comic.wszhu.top
```

---

## 三、加固 SSH 配置

```bash
sudo vi /etc/ssh/sshd_config
```

确认以下配置：

```
PermitRootLogin no          # 禁止 root SSH 登录
PasswordAuthentication no   # 禁止密码登录，只允许密钥
PubkeyAuthentication yes
```

```bash
sudo systemctl reload sshd
```

---

## 四、配置 sudo 权限（最小化）

```bash
sudo visudo -f /etc/sudoers.d/deploy
```

```
# 只允许 deploy 用户执行服务管理命令
deploy ALL=(ALL) NOPASSWD: /bin/systemctl restart ai-comic-server, /bin/systemctl stop ai-comic-server, /bin/systemctl start ai-comic-server
```

```bash
sudo chmod 440 /etc/sudoers.d/deploy
```

---

## 五、配置应用目录权限

```bash
sudo mkdir -p /opt/ai-comic/{bin,logs}

# bin 和 logs 归 deploy（部署时需要写入）
sudo chown -R deploy:deploy /opt/ai-comic/bin
sudo chown -R deploy:deploy /opt/ai-comic/logs

# 配置文件由 root 拥有，appuser 只读（保护 API Key）
sudo chown root:appuser /opt/ai-comic/config.yaml
sudo chmod 640 /opt/ai-comic/config.yaml
```

---

## 六、配置 systemd 服务

创建服务文件：

```bash
sudo vi /etc/systemd/system/ai-comic-server.service
```

```ini
[Unit]
Description=AI Comic Generator
After=network.target

[Service]
User=appuser
Group=appuser
WorkingDirectory=/opt/ai-comic
ExecStart=/opt/ai-comic/bin/server
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

启用并启动服务：

```bash
sudo systemctl daemon-reload       # 加载新配置
sudo systemctl enable ai-comic     # 设置开机自启
sudo systemctl start ai-comic      # 启动服务
sudo systemctl status ai-comic     # 查看状态
```

常用管理命令：

```bash
sudo systemctl stop ai-comic       # 停止
sudo systemctl restart ai-comic    # 重启
sudo journalctl -u ai-comic -f     # 实时查看日志
sudo journalctl -u ai-comic -n 100 # 查看最近 100 行日志
```

---

## 七、部署流程（日常更新）

```bash
# 1. 本地构建
npm run build:server

# 2. 上传二进制到服务器
scp -i C:\Users\Administrator\.ssh\deploy_key bin/server deploy@comic.wszhu.top:/opt/ai-comic/bin/server

# 3. SSH 登录后重启服务
ssh -i C:\Users\Administrator\.ssh\deploy_key deploy@comic.wszhu.top
sudo systemctl restart ai-comic
sudo systemctl status ai-comic
```
