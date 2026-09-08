-- 额度/VIP → 积分制迁移
-- 1) quota 重命名为 points（若已有 points 列可跳过 CHANGE）
-- 2) 删除 vipTime
-- 3) 历史 vip 角色降为 user

SET NAMES utf8mb4;

-- 若仍存在 quota 列：改名为 points
-- 注意：若列已是 points，本句会报错，可忽略后继续执行后续语句
ALTER TABLE `user`
    CHANGE COLUMN `quota` `points` int NOT NULL DEFAULT 100 COMMENT '积分';

-- 删除 VIP 时间列（若不存在会报错，可忽略）
ALTER TABLE `user` DROP COLUMN `vipTime`;

-- 历史 VIP 统一为普通用户
UPDATE `user` SET `userRole` = 'user' WHERE `userRole` = 'vip' AND `isDelete` = 0;
