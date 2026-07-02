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
    editTime     datetime     default CURRENT_TIMESTAMP not null comment '编辑时间',
    createTime   datetime     default CURRENT_TIMESTAMP not null comment '创建时间',
    updateTime   datetime     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    isDelete     tinyint      default 0                 not null comment '是否删除',
    UNIQUE KEY uk_userAccount (userAccount),
    INDEX idx_userName (userName)
) comment '用户' collate = utf8mb4_unicode_ci;

-- 初始化数据
-- 密码是 12345678（MD5 加密 + 盐值 mason）
INSERT INTO user (id, userAccount, userPassword, userName, userAvatar, userProfile, userRole) VALUES
(1, 'admin', 'cf0bada5c2fcee97fc65b9f2534ac461', '管理员', 'https://www.codefather.cn/logo.png', '系统管理员', 'admin'),
(2, 'user', 'cf0bada5c2fcee97fc65b9f2534ac461', '普通用户', 'https://www.codefather.cn/logo.png', '我是一个普通用户', 'user'),
(3, 'test', 'cf0bada5c2fcee97fc65b9f2534ac461', '测试账号', 'https://www.codefather.cn/logo.png', '这是一个测试账号', 'user');
