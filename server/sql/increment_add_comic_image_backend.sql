-- 增量脚本：为 comic 表添加 imageBackend 字段（用户按需选择生图后端）
-- 执行前提：comic 表已存在
-- 执行时机：数据库已部署但没有 imageBackend 字段时执行

USE ai_comic_generator;

ALTER TABLE `comic`
ADD COLUMN imageBackend varchar(50) NOT NULL DEFAULT 'hunyuan' COMMENT '生图后端：hunyuan/openai_image_1k/openai_image_4k' AFTER style;
