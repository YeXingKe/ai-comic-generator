# 技术方案：用户充值与收款记录

> 版本：v1.0  
> 日期：2026-08-14  
> 关联产品文档：[`prd-recharge.md`](./prd-recharge.md)  
> 前置：建议先完成 [`tech-auth.md`](./tech-auth.md) **P0**  
> 实现方式：自行按文档落地

---

## 1. 总览

### 1.1 架构

```text
web: /user/recharge, /admin/payments
  → PaymentHandler：下单 / 查单 / 我的订单 / 管理端分页
  → 回调 notify：验签 → 幂等发货（加 quota / 升 VIP）
```

### 1.2 原则

1. 发货只信回调或服务端查单  
2. 发货幂等（`PENDING → PAID` 影响行 = 1 才加额度）  
3. 复用现有 `quota` / `UpgradeToVIP` / `DecrementQuota`  
4. 先 `mock` 后真渠道  

---

## 2. 数据模型

### 2.1 `payment_product`

| 字段 | 说明 |
|------|------|
| name / sku / type | `quota` \| `vip` |
| quota_amount / vip_days | 权益 |
| price | 分 |
| status / sort | 上下架与排序 |

### 2.2 `payment_order`

| 字段 | 说明 |
|------|------|
| order_no / user_id / product_id | |
| 商品快照字段 | name/type/quota/vip_days/amount |
| channel | `mock` / `wechat` / `alipay` |
| status | `PENDING`/`PAID`/`CLOSED`/`REFUNDED` |
| transaction_id / pay_time | 渠道信息 |

可选：`payment_notify_log` 存回调原文。

---

## 3. API

前缀 `/api/payment`，响应 `{ code, data, message }`。

### 3.1 用户侧（需登录）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/payment/products` | 上架商品 |
| POST | `/payment/orders` | 下单 |
| GET | `/payment/orders/:orderNo` | 查自己的单 |
| GET | `/payment/orders/mine` | 我的订单分页 |

下单请求：`{ "productId": 1, "channel": "mock" }`。

### 3.2 回调（无 Session，验签）

`POST /payment/notify/wechat`、`/payment/notify/alipay`。

### 3.3 管理端

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/payment/admin/orders/page` | 全站分页 |
| GET | `/payment/admin/summary` | 可选汇总 |

---

## 4. 代码落点

```text
server/internal/model/payment*.go
server/internal/store/payment_*.go
server/internal/service/payment_service.go
server/internal/handler/payment_handler.go
server/internal/client/wechatpay/   # 真支付阶段
server/sql/payment.sql
server/cmd/server/main.go
server/config.yaml.example         # payment.*
```

发货与订单状态更新同一事务。

---

## 5. 前端

| 路由 | 文件 |
|------|------|
| `/user/recharge` | `pages/user/recharge/` |
| `/admin/payments` | `pages/admin/Payments/` + `paymentTableColumns.tsx` |

另：`api/payment.ts`、`types/api.ts`、`nav.ts`、创作页「去充值」。

可复用组件：`components/Payment/`。

交互：下单 → mock 直接成功或二维码 + 轮询查单 → `fetchLoginUser()`。

---

## 6. 配置示意

```yaml
payment:
  enabled: true
  mock_enabled: true    # 生产 false
  notify_base_url: "https://api.example.com"
  wechat:
    enabled: false
  alipay:
    enabled: false
```

---

## 7. 安全清单

| 项 | 要求 |
|----|------|
| 回调验签 | 必须 |
| 金额一致 | 必须 |
| 幂等发货 | 必须 |
| 查单越权 | `order.user_id` 校验 |
| mock | 生产关闭 |

---

## 8. 实施顺序（支付）

1. SQL + model/store/service/handler + mock  
2. 前端充值页 + 收款记录  
3. 真微信/支付宝  

---

## 9. 测试要点

发货幂等、非本人查单失败、非管理员访问 admin 失败、mock 到账后创作扣次、VIP 角色变化。

---

## 10. 现有锚点

| 能力 | 路径 |
|------|------|
| DecrementQuota / UpgradeToVIP | `server/internal/store/user.go` |
| 创作页额度 | `web/src/pages/user/create/index.tsx` |
| 路由注册 | `server/cmd/server/main.go` |
