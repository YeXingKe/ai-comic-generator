-- 增量脚本：为 comic 表添加 panelCount 字段（自动化分镜格数）
-- 执行前提：comic 表已存在

USE ai_comic_generator;

ALTER TABLE `comic`
ADD COLUMN panelCount int NOT NULL DEFAULT 4 COMMENT '分镜格数 4/6/8' AFTER captionTextMode;
