-- 增量：自定义创作增加参考图片 JSON 列
-- 已有 custom_comic 表时执行

USE ai_comic_generator;

ALTER TABLE custom_comic
    ADD COLUMN referenceImages json null comment '角色参考图列表（角色设定/立绘）' AFTER panelImages;
