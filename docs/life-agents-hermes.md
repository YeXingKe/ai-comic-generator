# 生活参谋：多 Agent + Hermes 微信助手

> 版本：v1.0  
> 日期：2026-09-09  
> 状态：方案文档（待独立仓库落地，**非本漫画仓库实现范围**）  
> 定位：热点 / 餐饮 / 出行 / 娱乐 / 旅行计划；微信发一条指令即可  
> 入口：NousResearch **Hermes Agent**（Weixin / iLink Bot 私聊）

---

## 1. 背景与目标

### 1.1 要解决什么

用户不想打开多个 App 查热搜、找餐厅、排周末路线。希望在 **微信私聊** 里发一句指令，由多智能体协作给出可读、可执行的推荐。

### 1.2 产品目标

| 目标 | 说明 |
|------|------|
| 一条指令可用 | `/热点` `/吃饭` `/出行` `/娱乐` `/计划` 或自然语言「帮我…」 |
| 推荐可追溯 | 条目来自 Tool / 知识库，禁止模型凭空编店名、景点 |
| 微信可读 | 短卡片：摘要 + 最多 3 条 + 追问；详情回 `1/2/3` |
| 可演进 | Hermes 只做通道；业务多 Agent 独立服务，可测可换 |

### 1.3 非目标（MVP 不做）

- 普通微信群 @ 唤醒（iLink 对群消息支持受限）
- 自动下单 / 支付 / 美团一键预订
- 完整旅游 SaaS、多城市运营后台
- 与 `ai-comic-generator` 漫画流水线耦合

### 1.4 与漫画项目的关系

本仓库仅 **存档方案**，便于学习多 Agent / 编排。落地建议新建独立仓（如 `life-hermes-agents`）。可复用思想：`handler → service → agent`、统一 JSON、编排器、微信侧会话。

---

## 2. 产品方案

### 2.1 能力矩阵

| Agent | 用户意图 | 典型输入 | 输出 |
|-------|----------|----------|------|
| 热点 | 看热搜/话题 | `/热点 AI` | 热榜条目 + 一句话点评 |
| 餐饮 | 找吃的 | `/吃饭 徐汇 火锅 人均100` | 3 家店 + 理由 + 区域/价位 |
| 出行 | 短途玩 | `/出行 上海周边 一天` | 3 个玩法 + 交通提示 |
| 娱乐 | 今晚看啥 | `/娱乐 电影` | 片单/展览/演出精选 |
| 旅行计划 | 多日行程 | `/计划 杭州 2日 轻松` | 分日上午/下午/晚上表 |
| Router | 分流 | `帮我周六别太远喝咖啡` | 选中子 Agent + 填槽 |

### 2.2 微信指令约定

```text
/热点 [关键词] [可选条数]
/吃饭 [区域] [品类] [预算] [人数]
/出行 [城市] [时长] [偏好]
/娱乐 [类型] [时间]
/计划 [目的地] [天数] [预算] [同行]
帮我 <自然语言>
```

示例：

```text
/吃饭 静安 日料 150
/计划 苏州 2日 情侣
帮我安排周六：上午咖啡下午散步，徐汇附近
```

多轮补槽：缺预算/城市时先问一句，再出结果。详情：用户回复 `1` / `2` / `3`。

### 2.3 回复版式（微信）

```text
【餐饮】徐汇附近 3 家人均≈150 日料

1. 店名A — 一句话理由
   静安寺 · ¥140 · 寿司
2. …
3. …

回复 1/2/3 看详情；或说「换成火锅」
```

旅行计划用分日短表，控制总字数；Hermes 侧可分段发送（注意平台单条长度限制）。

---

## 3. 总体技术架构

### 3.1 选型结论

采用 **方案 B：Hermes 做微信网关，自建 life-agents 做多 Agent 业务**。

| 方案 | 描述 | 结论 |
|------|------|------|
| A. Hermes 内堆全部 Agent | 提示词 + Hermes tools 全写在网关里 | 出活快，难测、难演进 |
| **B. Hermes 通道 + 自建服务** | Hermes Tool 调 `POST /api/chat` | **推荐** |

### 3.2 逻辑架构图

```text
┌─────────────────────────────────────────────────────────┐
│  微信（个人号私聊 → iLink Bot 身份）                      │
└───────────────────────────┬─────────────────────────────┘
                            │ long-poll（无需公网 Webhook）
                            ▼
┌─────────────────────────────────────────────────────────┐
│  Hermes Gateway                                         │
│  · Weixin adapter（QR 登录 / account_id + token）         │
│  · allowlist（仅授权用户）                                │
│  · session / 分段回复 / 可选 cron 推送                    │
│  · Tool: life_assist(message, user_id)                  │
└───────────────────────────┬─────────────────────────────┘
                            │ HTTP JSON
                            ▼
┌─────────────────────────────────────────────────────────┐
│  life-agents 服务                                        │
│                                                         │
│  ┌─────────┐   ┌──────────────┐   ┌──────────────────┐ │
│  │ handler │ → │ Router Agent │ → │ Expert Agents    │ │
│  │ /api/*  │   │ 意图+槽位     │   │ 热点餐饮出行…    │ │
│  └─────────┘   └──────────────┘   └────────┬─────────┘ │
│                                            │            │
│                    ┌───────────────────────┼──────┐     │
│                    ▼           ▼           ▼      ▼     │
│                 Tools       RAG         Cache   LLM     │
│               (地图天气     (精选库)    (Redis) (通义等) │
│                热搜POI)                                  │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
                 RecommendationSchema → 微信短文案
```

### 3.3 分层（对齐漫画后端习惯）

```text
life-hermes-agents/
├── docs/                      # 本方案可迁至此
├── hermes/                    # 网关配置说明、tool 描述
│   └── TOOLS.md
├── server/
│   ├── cmd/ / main            # 启动入口
│   ├── internal/
│   │   ├── handler/           # HTTP：/api/chat /health
│   │   ├── service/           # 会话、编排入口
│   │   ├── agent/
│   │   │   ├── router/
│   │   │   ├── hot/
│   │   │   ├── dining/
│   │   │   ├── trip/
│   │   │   ├── fun/
│   │   │   └── plan/          # 编排其它 agent/tools
│   │   ├── tools/             # POI、天气、热搜、距离
│   │   ├── rag/               # 精选库检索
│   │   ├── model/             # Schema、Session、槽位
│   │   ├── store/             # Redis / 可选 MySQL
│   │   └── format/            # Schema → 微信文本
│   └── config.yaml
└── data/seed/                 # 餐厅/路线种子 JSON
```

语言建议：

- **Python（FastAPI）**：接 LLM / RAG / LangGraph 更快，适合本项目 MVP  
- **Go**：若希望与漫画仓技术栈统一，同样可行  

以下接口与模块名与语言无关。

---

## 4. Hermes 接入设计

### 4.1 角色

| 组件 | 做什么 |
|------|--------|
| Hermes Weixin | 扫码绑定 iLink；收私聊；按 allowlist 过滤 |
| Hermes Tool `life_assist` | 把用户原文 + 平台 user_id 转发给 life-agents |
| Hermes 系统提示（薄） | 「生活类问题调用 life_assist；不要自己编造店铺」 |
| Hermes Cron（P2） | 每天定时调热点接口，推送到绑定用户 |

### 4.2 安全

1. `WEIXIN` / DM **allowlist** 只加自己（及测试号）  
2. life-agents 校验 `X-Api-Key` 或 mTLS，只接受 Hermes 所在机器调用  
3. 不把高危 shell / 任意 URL 抓取工具暴露给微信通道  
4. 日志打 `user_id` + 意图，不落完整手机号等敏感信息  

### 4.3 已知平台限制（实现前必读）

- 适配器面向 **个人微信 iLink Bot**，与企业微信 WeCom 不同  
- **私聊为主**；普通群消息 / @ 个人号往往到不了 Hermes  
- long-poll：**同一 token 不要多实例**，否则可能丢消息  
- 官方文档： [Hermes Weixin](https://hermes-agent.nousresearch.com/docs/user-guide/messaging/weixin)

### 4.4 Hermes ↔ 服务契约

Hermes Tool 入参示例：

```json
{
  "user_id": "wx_ilink_xxx",
  "message": "/吃饭 徐汇 火锅 100",
  "channel": "weixin"
}
```

life-agents 返回：

```json
{
  "reply": "【餐饮】…（已格式化的微信文本）",
  "meta": { "intent": "dining", "item_count": 3 }
}
```

Hermes 将 `reply` 原样发回微信（注意长度分段）。

---

## 5. 核心实现

### 5.1 统一输出 Schema

所有专家 Agent 最终收敛为：

```json
{
  "type": "hot | dining | trip | fun | plan",
  "summary": "一句话摘要",
  "items": [
    {
      "id": "1",
      "title": "名称",
      "reason": "推荐理由（≤40字）",
      "meta": { "area": "", "price": 0, "tags": [], "url": "" },
      "action": "可选：导航文案或链接"
    }
  ],
  "followups": ["要不要换成…？"],
  "sources": ["tool:hotlist", "rag:dining_seed"]
}
```

**硬规则**：`items[].title` 必须能在本次 Tool/RAG 命中集合中找到；否则丢弃该条并降级「暂时没有合适结果」。

### 5.2 Router

```text
message
  → 若匹配 ^/(热点|吃饭|出行|娱乐|计划)
       则规则路由 + 正则/LLM 抽槽
  → 否则 LLM 分类到意图枚举 + 抽槽
  → 写入 Session.slots
  → 槽位不足则返回追问（不调专家）
  → 槽位足够则调对应 Agent
```

意图枚举：`hot | dining | trip | fun | plan | smalltalk | unknown`。

### 5.3 专家 Agent 职责

| Agent | 步骤 |
|-------|------|
| Hot | `get_hotlist(keyword)` → 截取 TopN → LLM 写一句点评（不改标题） |
| Dining | 槽位 → `search_poi` 或 `rag_search(dining)` → 过滤预算 → 排序 → 填 Schema |
| Trip | 天气 + POI/周边 → 3 套玩法 |
| Fun | 精选库按类型/时间过滤 |
| Plan | 编排：天气 → 候选点 → `route_distance` 校验 → 生成 D1/D2 表 → Schema.type=plan |

Agent 接口建议：

```text
Execute(ctx, state: SessionState) -> RecommendationSchema
```

与漫画项目 `agent.Agent` 同构，便于对照学习。

### 5.4 Tools 清单

| Tool | 输入 | 输出 | MVP |
|------|------|------|-----|
| `get_hotlist` | keyword, limit | 标题/热度/链接 | P0（可先 mock） |
| `rag_search` | collection, query, top_k | 文档块 | P0 餐饮种子 |
| `search_poi` | city, keyword, radius | 名称/地址/均价 | P1 |
| `get_weather` | city, date | 天气摘要 | P1 |
| `route_distance` | points[] | 路程/时长 | P2 计划用 |

### 5.5 RAG（餐饮/娱乐先用这个压幻觉）

```text
data/seed/dining_shanghai.json
  → 切块/向量（或先 BM25 关键词）
  → rag_search("徐汇 火锅 100")
  → 仅在命中集里让 LLM 挑选 3 条并写 reason
```

种子字段示例：`name, area, category, price_avg, tags, one_liner, url`。

### 5.6 Session / 多轮

Redis key：`life:sess:{user_id}`

```json
{
  "intent": "dining",
  "slots": { "area": "徐汇", "category": "火锅", "budget": null },
  "last_items": [ {"id":"1","title":"..."}, ... ],
  "updated_at": "..."
}
```

用户发 `2` → 若存在 `last_items`，返回第 2 条详情，不重新路由。

### 5.7 HTTP API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/health` | 健康检查 |
| POST | `/api/chat` | Hermes 主入口 |
| POST | `/api/recommend` | 可选：按 intent 直调（调试用） |
| POST | `/api/push/daily-hot` | P2：定时热点（Hermes cron 或系统 cron） |

`POST /api/chat` 请求：

```json
{
  "user_id": "string",
  "message": "string",
  "channel": "weixin"
}
```

响应：

```json
{
  "code": 0,
  "data": {
    "reply": "string",
    "intent": "dining",
    "need_followup": false
  },
  "message": "ok"
}
```

错误码可与漫画项目对齐：`0 / 40000 / 50000`。

### 5.8 配置项（示例）

```yaml
server:
  port: 8090
  api_key: "changeme"

llm:
  api_key: ""
  model: "qwen-plus"
  base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"

redis:
  addr: "localhost:6379"

rag:
  dining_seed: "./data/seed/dining_shanghai.json"

tools:
  hotlist:
    enabled: false   # false 时用 mock
  amap:
    enabled: false
    key: ""
```

---

## 6. 端到端时序

### 6.1 前缀指令（吃饭）

```text
用户微信: /吃饭 徐汇 火锅 100
  → Hermes 收消息（allowlist 通过）
  → Tool life_assist
  → POST /api/chat
  → Router: intent=dining, slots 齐全
  → DiningAgent: rag_search → 选 3 → Schema
  → format.Wechat(Schema) → reply
  → Hermes 发回微信
```

### 6.2 自然语言 + 补槽

```text
用户: 帮我找个地方吃饭
  → Router: dining, area/budget 空
  → reply: 「在哪个区域？预算大概多少？」
用户: 徐汇 100
  → 合并 slots → DiningAgent → 3 条推荐
用户: 1
  → 详情（地址/标签/一句话）
```

### 6.3 旅行计划（P2）

```text
/计划 杭州 2日 轻松
  → PlanAgent
      get_weather(杭州, D1/D2)
      rag/poi 候选
      route_distance 过滤过远组合
      生成日程表 Schema
  → 微信分日短文案
```

---

## 7. 落地里程碑

### P0（先跑通「微信一条指令」）

- [ ] Hermes Weixin QR 登录 + allowlist  
- [ ] life-agents：`/health` + `/api/chat`  
- [ ] Router 前缀：`/热点` `/吃饭`  
- [ ] Hot mock + Dining 种子 RAG  
- [ ] Schema → 微信文案  
- [ ] Hermes 注册 `life_assist` 并联调  

**验收**：私聊发送 `/吃饭 徐汇 火锅 100`，收到 3 条结构化推荐。

### P1

- [ ] `/出行` `/娱乐`  
- [ ] 多轮槽位 + Redis Session + 回 `1/2/3`  
- [ ] 自然语言「帮我」分类  
- [ ] 热搜/天气真接口（可配置降级 mock）  

### P2

- [ ] `/计划` 编排 + 路程校验  
- [ ] 每日热点推送  
- [ ] 简单反馈（有用/没用）日志，供后续排序  

---

## 8. 质量与风控

| 风险 | 对策 |
|------|------|
| 餐饮幻觉 | 只推荐 RAG/POI 命中集；`sources` 必填 |
| Prompt 过长 | 槽位短、种子摘要短；理由限字 |
| 微信刷接口 | allowlist + api_key + 限流 |
| 外部 API 挂 | tools.enabled=false 走 mock/种子，保证指令有响应 |
| 隐私 | 日志脱敏；种子数据勿含未授权爬取内容 |

---

## 9. 测试建议

| 类型 | 内容 |
|------|------|
| 单测 | Router 前缀解析、槽位合并、Schema 过滤「无来源条目」 |
| 契约测 | `/api/chat` 固定用例黄金文案快照 |
| 联调 | Hermes 真机私聊 10 条指令清单 |
| 降级 | 关 LLM 时是否仍能用纯检索出 3 家（可选） |

---

## 10. 决策摘要

| 项 | 选择 |
|----|------|
| 入口 | Hermes + 微信私聊（iLink） |
| 业务 | 独立 life-agents 多 Agent 服务 |
| MVP | 热点 + 餐饮 |
| 指令 | `/` 前缀 + `帮我` |
| 防幻觉 | Tool/RAG 白名单 + 统一 Schema |
| 仓库 | 新建 `life-hermes-agents`（本文件先存于漫画仓 docs） |

---

## 11. 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.0 | 2026-09-09 | 初稿：产品指令、Hermes 通道架构、模块/API/Tools、P0–P2 |
