-- 积分充值订单与积分流水
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `pay_order` (
  `id`            bigint       NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `orderNo`       varchar(64)  NOT NULL COMMENT '商户订单号（out_trade_no）',
  `userId`        bigint       NOT NULL COMMENT '下单用户 ID',
  `channel`       varchar(16)  NOT NULL COMMENT '支付渠道：alipay / mock',
  `packageCode`   varchar(32)  NOT NULL COMMENT '充值方案编码（pay_package.code）',
  `amountFen`     int          NOT NULL COMMENT '应付金额（分）',
  `points`        int          NOT NULL COMMENT '到账积分',
  `status`        varchar(16)  NOT NULL DEFAULT 'PENDING' COMMENT '订单状态：PENDING 待支付 / PAID 已支付 / CLOSED 已关闭',
  `channelTxnId`  varchar(64)  DEFAULT NULL COMMENT '渠道交易号（如支付宝 trade_no）',
  `codeUrl`       varchar(512) DEFAULT NULL COMMENT '扫码支付链接（支付宝 qr_code）',
  `expireAt`      datetime     NOT NULL COMMENT '订单过期时间',
  `paidAt`        datetime     DEFAULT NULL COMMENT '支付成功时间',
  `createTime`    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updateTime`    datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_orderNo` (`orderNo`),
  KEY `idx_userId_createTime` (`userId`, `createTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分充值订单';

CREATE TABLE IF NOT EXISTS `point_log` (
  `id`           bigint       NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `userId`       bigint       NOT NULL COMMENT '用户 ID',
  `delta`        int          NOT NULL COMMENT '积分变动（正数增加，负数扣减）',
  `balanceAfter` int          NOT NULL COMMENT '变动后积分余额',
  `bizType`      varchar(32)  NOT NULL COMMENT '业务类型（如 recharge 充值入账）',
  `bizId`        varchar(64)  DEFAULT NULL COMMENT '业务关联 ID（如订单号 orderNo）',
  `remark`       varchar(255) DEFAULT NULL COMMENT '备注说明',
  `createTime`   datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_biz` (`bizType`, `bizId`),
  KEY `idx_userId` (`userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分流水';
