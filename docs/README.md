# docs

项目产品与技术方案文档。**请自行按文档实现；鉴权优先于支付。**

## 入门：Go 后端总览（小白）

| 文档 | 说明 |
|------|------|
| [server-overview.md](./server-overview.md) | 产品介绍 + 技术实现方案（流水线、分层、API、配置降级） |
| [../server/README.md](../server/README.md) | `server/` 目录速查与启动 |

## 需求变更：合成版式

| 文档 | 说明 |
|------|------|
| [prd-compose-layout.md](./prd-compose-layout.md) | 分格 16:9 合成：1～3 竖排、4/6/8 两列、黑框模板（待开发） |

## 优先：鉴权

| 文档 | 说明 |
|------|------|
| [prd-auth.md](./prd-auth.md) | 产品：登录鉴权增强（目标、场景、P0/P1 验收） |
| [tech-auth.md](./tech-auth.md) | 技术方案：密码/Session/CORS/前端 401、改动清单与自测 |

## 随后：支付 / 充值

| 文档 | 说明 |
|------|------|
| [prd-recharge.md](./prd-recharge.md) | 产品：用户充值、管理员收款记录 |
| [tech-recharge.md](./tech-recharge.md) | 技术方案：表结构、API、前后端落点 |

## 建议顺序

```text
1. 按 tech-auth.md 完成鉴权 P0 并自测验收
2. （可选）鉴权 P1
3. 按 prd-compose-layout.md 改 Compose 版式（可与鉴权并行）
4. 按 tech-recharge.md 做 mock 支付 → 前端页面 → 真渠道
```

环境搭建与仓库约定见 [`AGENTS.md`](../AGENTS.md)。
