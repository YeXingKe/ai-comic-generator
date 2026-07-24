#!/usr/bin/env bash
# ai-comic-generator 一键部署脚本（在宝塔服务器上执行）
#
# 用途：git pull 最新代码 → 构建前端 → 构建后端 → 重启 systemd 服务 → 健康检查
# 前置：服务器已装 Go 1.24+、Node 18+/npm，已 clone 本仓库，已放好 server/config.yaml，
#       已 systemctl enable 好 deploy/ai-comic-server.service（见 deploy/README.md）
#
# 用法：
#   cd /www/wwwroot/ai-comic-generator      # 你的项目根目录
#   bash deploy/deploy.sh                    # 完整部署
#   bash deploy/deploy.sh --skip-web         # 只重建后端
#   bash deploy/deploy.sh --skip-server      # 只重建前端
#   bash deploy/deploy.sh --no-pull          # 跳过 git pull（用本地已有代码）


# ⭐⭐⭐ 暂不使用当前脚本部署 ⭐⭐⭐

set -euo pipefail

# ---- 可按需覆盖的变量（也可通过环境变量传入） ----
SERVICE_NAME="${SERVICE_NAME:-ai-comic-server}"
GIT_BRANCH="${GIT_BRANCH:-main}"
HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:8080/api/health}"

# 定位项目根目录（脚本在 deploy/ 下）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT_DIR}"

SKIP_WEB=0
SKIP_SERVER=0
NO_PULL=0
for arg in "$@"; do
  case "$arg" in
    --skip-web) SKIP_WEB=1 ;;
    --skip-server) SKIP_SERVER=1 ;;
    --no-pull) NO_PULL=1 ;;
    *) echo "未知参数: $arg" >&2; exit 1 ;;
  esac
done

log() { printf '\033[36m[deploy]\033[0m %s\n' "$*"; }
die() { printf '\033[31m[deploy][error]\033[0m %s\n' "$*" >&2; exit 1; }

# ---- 0. 前置检查 ----
[ -f server/config.yaml ] || die "缺少 server/config.yaml，请先从 config.yaml.example 复制并填写"
command -v go >/dev/null || die "未找到 go，请先安装 Go 1.24+"
command -v npm >/dev/null || die "未找到 npm，请先安装 Node.js"

# ---- 1. 拉取代码 ----
if [ "$NO_PULL" -eq 0 ]; then
  log "拉取分支 ${GIT_BRANCH} 最新代码"
  git fetch --prune origin
  git checkout "${GIT_BRANCH}"
  git reset --hard "origin/${GIT_BRANCH}"   # 服务器不做本地改动，硬对齐远端
  git log -1 --oneline
fi

# ---- 2. 构建前端 ----
if [ "$SKIP_WEB" -eq 0 ]; then
  log "构建前端 (web/)"
  cd "${ROOT_DIR}/web"
  if [ -f package-lock.json ]; then
    npm ci
  else
    npm install
  fi
  npm run build            # 内含 vite build + tsc -b，产物 web/dist/
  cd "${ROOT_DIR}"
  log "前端产物已更新：web/dist/"
fi

# ---- 3. 构建后端 ----
if [ "$SKIP_SERVER" -eq 0 ]; then
  log "构建后端 (server/)"
  cd "${ROOT_DIR}/server"
  go mod download
  # 输出到 server/bin/server，systemd 的 WorkingDirectory 指向 server/，
  # 保证运行时能读到同目录的 config.yaml 与写入 ./data/comics
  # 服务器内存紧张，限制编译并行度以降低构建期内存峰值，避免被 OOM killer 杀掉
  CGO_ENABLED=0 GOFLAGS="-p=1" go build -trimpath -ldflags "-s -w" -o bin/server ./cmd/server/main.go
  cd "${ROOT_DIR}"
  log "后端产物已更新：server/bin/server"
fi

# ---- 4. 重启服务 ----
if [ "$SKIP_SERVER" -eq 0 ]; then
  log "重启 systemd 服务 ${SERVICE_NAME}"
  sudo systemctl restart "${SERVICE_NAME}"
  sleep 2
  sudo systemctl --no-pager --lines 0 status "${SERVICE_NAME}" || true
fi

# ---- 5. 健康检查 ----
log "健康检查 ${HEALTH_URL}"
for i in 1 2 3 4 5; do
  if curl -fsS "${HEALTH_URL}" >/dev/null 2>&1; then
    log "部署成功，后端已就绪 ✅"
    exit 0
  fi
  sleep 2
done
die "健康检查失败，请查看日志：sudo journalctl -u ${SERVICE_NAME} -n 100 --no-pager"
