-- 增量脚本：为用户表添加 status 字段
-- 执行前提：user 表已存在
-- 执行时机：数据库已部署但没有 status 字段时执行

USE ai_comic_generator;

-- 添加 status 字段
ALTER TABLE `user`
ADD COLUMN status tinyint NOT NULL DEFAULT 1 COMMENT '用户状态：1 启用，0 禁用' AFTER userRole;

-- 添加索引
ALTER TABLE `user`
ADD INDEX idx_status (status);
