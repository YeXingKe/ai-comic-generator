-- 电脑网站支付（page）跳转 URL 含签名，远超 512
-- 已有 pay_order 表时执行

USE ai_comic_generator;

ALTER TABLE `pay_order`
    MODIFY COLUMN `codeUrl` text NULL COMMENT '扫码 qr_code 或电脑网站支付跳转 URL';
