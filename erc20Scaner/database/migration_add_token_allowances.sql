-- 数据库升级脚本：新增 token_allowances 授权额度表
-- 原因：支持 ERC20 Approval 事件跟踪和授权额度查询
-- 执行时间：2026-06-08
-- 相关功能：GET /evmapi/accounts/{address}/approvals

-- 创建授权额度表
CREATE TABLE IF NOT EXISTS token_allowances (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    owner VARCHAR(42) NOT NULL COMMENT '授权方地址',
    spender VARCHAR(42) NOT NULL COMMENT '被授权方地址',
    contract_address VARCHAR(42) NOT NULL COMMENT 'ERC20合约地址(外键关联contracts表)',
    amount DECIMAL(65,0) NOT NULL DEFAULT '0' COMMENT '授权额度（原始值）',
    last_tx_hash VARCHAR(66) DEFAULT NULL COMMENT '最近一笔Approval事件tx hash',
    last_block_number BIGINT UNSIGNED DEFAULT NULL COMMENT '最近一笔Approval事件区块号',
    last_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_owner_spender_contract (owner, spender, contract_address),
    KEY idx_owner_amount (owner, amount),
    KEY idx_spender (spender),
    KEY idx_contract_address (contract_address),
    CONSTRAINT fk_allowance_contract FOREIGN KEY (contract_address) REFERENCES contracts(contract_address) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ERC20授权额度表';

-- 注意：
-- 1. 此表仅新增，不影响现有表和数据
-- 2. 已有安装执行此脚本即可，Go 代码中 InitTokenAllowancesTable() 也会自动建表
-- 3. 历史 Approval 事件不会回填，仅从升级后扫描的区块开始采集
-- 4. amount=0 的记录会被扫描器自动删除（表示授权已撤销）
