-- 充值方案配置（管理员可 CRUD，用户端展示 enabled=1）
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `pay_package` (
  `id`          bigint       NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `code`        varchar(32)  NOT NULL COMMENT '套餐编码（唯一，下单时 packageCode）',
  `name`        varchar(64)  NOT NULL COMMENT '展示名称',
  `amountFen`   int          NOT NULL COMMENT '售价（分）',
  `points`      int          NOT NULL COMMENT '到账积分',
  `sortOrder`   int          NOT NULL DEFAULT 0 COMMENT '排序权重，越小越靠前',
  `enabled`     tinyint      NOT NULL DEFAULT 1 COMMENT '上架状态：1 上架，0 下架',
  `createTime`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updateTime`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='充值方案';

INSERT INTO `pay_package` (`code`, `name`, `amountFen`, `points`, `sortOrder`, `enabled`)
SELECT 'p60', '入门包', 600, 60, 10, 1 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `pay_package` WHERE `code` = 'p60');

INSERT INTO `pay_package` (`code`, `name`, `amountFen`, `points`, `sortOrder`, `enabled`)
SELECT 'p180', '常用包', 1800, 200, 20, 1 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `pay_package` WHERE `code` = 'p180');

INSERT INTO `pay_package` (`code`, `name`, `amountFen`, `points`, `sortOrder`, `enabled`)
SELECT 'p680', '创作包', 6800, 800, 30, 1 FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `pay_package` WHERE `code` = 'p680');
