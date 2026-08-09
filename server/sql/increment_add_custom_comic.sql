-- 增量脚本：自定义创作任务表
-- 执行前提：库 ai_comic_generator 已存在
-- 执行时机：需要自定义创作功能时执行

USE ai_comic_generator;

CREATE TABLE IF NOT EXISTS custom_comic
(
    id           bigint auto_increment comment 'id' primary key,
    taskId       varchar(64)                        not null comment '任务 ID（UUID）',
    userId       bigint                             not null comment '所属用户 ID',
    prompt       text                               not null comment '用户提示词',
    aspectRatio  varchar(16)  default '16:9'        not null comment '画幅：1:1 / 16:9 / 9:16 / 2:3',
    imageBackend varchar(50)  default 'hunyuan'     not null comment '生图后端',
    panelCount   int          default 4             not null comment '分镜格数 2-8',
    panelImages  json                               null comment '分镜图片列表',
    status       varchar(20)  default 'PENDING'     not null comment 'PENDING/PROCESSING/COMPLETED/FAILED',
    errorMessage text                               null comment '失败错误信息',
    createTime   datetime     default CURRENT_TIMESTAMP not null comment '创建时间',
    updateTime   datetime     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    isDelete     tinyint      default 0             not null comment '软删除',
    UNIQUE KEY uk_taskId (taskId),
    INDEX idx_userId (userId),
    INDEX idx_status (status),
    INDEX idx_createTime (createTime)
) comment '自定义创作任务表' collate = utf8mb4_unicode_ci;
