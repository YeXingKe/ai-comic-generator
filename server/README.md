# server/

Go 后端：Gin + GORM + MySQL + Redis Session。负责用户鉴权、自动化/自定义漫画流水线、生图与排版、可选公众号发布。

## 小白从这里读

完整的**产品介绍 + 技术方案**（状态机、目录分层、API、配置降级）：

→ [docs/server-overview.md](../docs/server-overview.md)

## 快速启动

```bash
# 1. 复制配置并填写 MySQL / Redis /（可选）DashScope
cp config.yaml.example config.yaml

# 2. 建库执行 sql/create_table.sql，按需执行 sql/increment_*.sql

# 3. 启动（也可在仓库根目录 npm run server）
go run ./cmd/server/main.go
```

健康检查：`GET http://localhost:8080/api/health`

## 目录速查

| 路径 | 职责 |
|------|------|
| `cmd/server/main.go` | 入口与路由 |
| `internal/app/` | 依赖组装 |
| `internal/handler/` | HTTP |
| `internal/service/` | 业务与流水线 |
| `internal/store/` | 数据库 |
| `internal/model/` | 实体与 DTO |
| `sql/` | 建表与增量脚本 |

更多仓库约定见根目录 [`AGENTS.md`](../AGENTS.md)。
