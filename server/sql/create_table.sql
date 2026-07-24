-- 设置字符集（解决中文乱码问题）
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 创建库
create database if not exists ai_comic_generator CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 切换库
use ai_comic_generator;

-- 用户表（基础字段，quota 和 vipTime 由增量脚本添加）
create table if not exists user
(
    id           bigint auto_increment comment 'id' primary key,
    userAccount  varchar(256)                           not null comment '账号',
    userPassword varchar(512)                           not null comment '密码',
    userName     varchar(256)                           null comment '用户昵称',
    userAvatar   varchar(1024)                          null comment '用户头像',
    userProfile  varchar(512)                           null comment '用户简介',
    userRole     varchar(256) default 'user'            not null comment '用户角色：user/admin',
    status       tinyint      default 1                 not null comment '用户状态：1 启用，0 禁用',
    editTime     datetime     default CURRENT_TIMESTAMP not null comment '编辑时间',
    createTime   datetime     default CURRENT_TIMESTAMP not null comment '创建时间',
    updateTime   datetime     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    isDelete     tinyint      default 0                 not null comment '是否删除',
    UNIQUE KEY uk_userAccount (userAccount),
    INDEX idx_userName (userName),
    INDEX idx_status (status)
) comment '用户' collate = utf8mb4_unicode_ci;

-- 初始化数据
-- 密码是 12345678（MD5 加密 + 盐值 mason）
INSERT INTO user (id, userAccount, userPassword, userName, userAvatar, userProfile, userRole) VALUES
(1, 'admin', 'cf0bada5c2fcee97fc65b9f2534ac461', '管理员', 'https://www.codefather.cn/logo.png', '系统管理员', 'admin'),
(2, 'user', 'cf0bada5c2fcee97fc65b9f2534ac461', '普通用户', 'https://www.codefather.cn/logo.png', '我是一个普通用户', 'user'),
(3, 'test', 'cf0bada5c2fcee97fc65b9f2534ac461', '测试账号', 'https://www.codefather.cn/logo.png', '这是一个测试账号', 'user');

-- 漫画生成任务表（对应 model.Comic，表名 comic）
create table if not exists comic
(
    id              bigint auto_increment comment 'id' primary key,  -- 主键
    taskId          varchar(64)                           not null comment '任务 ID（UUID），唯一标识一次漫画生成',
    userId          bigint                                not null comment '所属用户 ID，关联 user.id',
    topic           varchar(500)                          not null comment '创作主题/关键词',
    userDescription TEXT                                  null comment '用户补充描述',
    title           varchar(200)                          null comment '漫画标题（故事构思后写入）',
    coverImage      varchar(1024)                         null comment '封面图 URL（排版合成后写入）',
    style           varchar(50)  default 'cartoon'        not null comment '漫画风格：cartoon/realistic/chibi',
    titleOptions    JSON                                  null comment '第0步：标题推荐列表',
    storyIdeation   JSON                                  null comment '第1步：故事构思结果',
    characters      JSON                                  null comment '第2步：角色设定列表',
    storyboard      JSON                                  null comment '第3步：分镜脚本',
    panelImages     JSON                                  null comment '第4步：分镜格图片列表',
    composedLayout  JSON                                  null comment '第5步：排版合成结果',
    publishResult   JSON                                  null comment '第6步：公众号发布结果',
    status          varchar(20)  default 'PENDING'        not null comment '任务状态：PENDING/PROCESSING/AWAITING_CONFIRM/TITLE_CONFIRMED/COMPLETED/FAILED',
    phase           varchar(50)  default 'PENDING'        not null comment '当前阶段：PENDING/TITLE_GENERATION/TITLE_SELECTING/STORY_IDEATION/...',
    errorMessage    TEXT                                  null comment '失败时的错误信息',
    createTime      datetime     default CURRENT_TIMESTAMP not null comment '创建时间',
    completedTime   datetime                              null comment '完成时间',
    updateTime      datetime     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    isDelete        tinyint      default 0                 not null comment '软删除：0 正常，1 已删除',
    -- 用来加快按某些字段查询的速度
    UNIQUE KEY uk_taskId (taskId),   -- 唯一索引 UNIQUE
    INDEX idx_userId (userId),   -- 普通索引 INDEX
    INDEX idx_status (status),
    INDEX idx_createTime (createTime),
    INDEX idx_userId_status (userId, status)  -- 联合索引
) comment '漫画生成任务表' COLLATE = utf8mb4_unicode_ci;

-- 智能体执行日志表
create table if not exists agent_log
(
    id              bigint auto_increment comment 'id' primary key,
    taskId          varchar(64)                        not null comment '任务ID',
    agentName       varchar(50)                        not null comment '智能体名称',
    startTime       datetime                           not null comment '开始时间',
    endTime         datetime                           null comment '结束时间',
    durationMs      int                                null comment '耗时（毫秒）',
    status          varchar(20)                        not null comment '状态：SUCCESS/FAILED',
    errorMessage    text                               null comment '错误信息',
    prompt          text                               null comment '使用的Prompt',
    inputData       json                               null comment '输入数据（JSON格式）',
    outputData      json                               null comment '输出数据（JSON格式）',
    createTime      datetime    default CURRENT_TIMESTAMP not null comment '创建时间',
    updateTime      datetime    default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    isDelete        tinyint     default 0              not null comment '是否删除',
    INDEX idx_taskId (taskId),
    INDEX idx_agentName (agentName),
    INDEX idx_status (status),
    INDEX idx_createTime (createTime)
) comment '智能体执行日志表' collate = utf8mb4_unicode_ci;