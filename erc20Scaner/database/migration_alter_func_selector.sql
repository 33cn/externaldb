-- 数据库升级脚本：修改 func_selector 字段长度
-- 原因：需要支持特殊标记如 "nested_call" (11个字符)，原 VARCHAR(10) 不够
-- 执行时间：2026-01-21
-- 注意：MySQL 5.7+ 不支持 IF EXISTS，需要先检查外键是否存在

-- 1. 删除外键约束（因为特殊标记不在 function_signatures 表中）
-- 注意：如果外键不存在，会报错，可以忽略
-- 对于 MySQL 8.0+，可以使用以下方式：
-- SET @dbname = DATABASE();
-- SET @tablename = "transactions";
-- SET @constraintname = "fk_tx_func_selector";
-- SET @preparedStatement = (SELECT IF(
--   (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS 
--    WHERE TABLE_SCHEMA = @dbname 
--    AND TABLE_NAME = @tablename 
--    AND CONSTRAINT_NAME = @constraintname) > 0,
--   CONCAT("ALTER TABLE ", @tablename, " DROP FOREIGN KEY ", @constraintname),
--   "SELECT 1"
-- ));
-- PREPARE stmt FROM @preparedStatement;
-- EXECUTE stmt;
-- DEALLOCATE PREPARE stmt;

-- 简单方式：直接删除，如果不存在会报错，可以忽略
ALTER TABLE transactions DROP FOREIGN KEY fk_tx_func_selector;

-- 2. 修改 func_selector 字段长度
ALTER TABLE transactions MODIFY COLUMN func_selector VARCHAR(50) DEFAULT NULL COMMENT '函数选择器(4字节hex或特殊标记如nested_call)';

-- 3. 不重新创建外键约束
-- 原因：由于 "nested_call" 等特殊标记不在 function_signatures 表中，
-- 如果保留外键约束，这些特殊标记的 func_selector 必须为 NULL
-- 为了支持特殊标记，不创建外键约束

