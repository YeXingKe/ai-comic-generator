-- 增量脚本：为 comic 表添加 captionTextMode 字段（文案展示模式）
-- 执行前提：comic 表已存在
-- 执行时机：数据库已部署但没有 captionTextMode 字段时执行

USE ai_comic_generator;

ALTER TABLE `comic`
ADD COLUMN captionTextMode varchar(20) NOT NULL DEFAULT 'top' COMMENT '文案展示模式：none/top/bubble' AFTER imageBackend;
