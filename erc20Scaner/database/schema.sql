-- 代币扫描器数据库表结构设计（支持ERC20/ERC721/ERC1155等）
-- 创建数据库
CREATE DATABASE IF NOT EXISTS token_scanner DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE token_scanner;

-- 0. 函数签名表 (function_signatures)
-- 存储常见的函数签名和对应的函数信息
CREATE TABLE IF NOT EXISTS function_signatures (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    selector VARCHAR(10) NOT NULL COMMENT '函数选择器(4字节hex，如a9059cbb)',
    function_name VARCHAR(100) NOT NULL COMMENT '函数名称(如transfer)',
    function_signature VARCHAR(255) NOT NULL COMMENT '完整函数签名(如transfer(address,uint256))',
    standard_type VARCHAR(50) DEFAULT NULL COMMENT '标准类型(ERC20/ERC721/ERC1155等)',
    function_type VARCHAR(50) DEFAULT NULL COMMENT '函数类型(view/pure/payable/nonpayable)',
    description TEXT DEFAULT NULL COMMENT '函数描述',
    is_standard BOOLEAN DEFAULT TRUE COMMENT '是否为标准函数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_selector (selector),
    KEY idx_function_name (function_name),
    KEY idx_standard_type (standard_type),
    KEY idx_function_type (function_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='函数签名表';

-- 1. 合约表 (contracts)
-- 存储代币合约的基本信息（支持ERC20/ERC721/ERC1155等）
CREATE TABLE IF NOT EXISTS contracts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    contract_address VARCHAR(42) NOT NULL COMMENT '合约地址(0x开头)',
    contract_name VARCHAR(255) DEFAULT NULL COMMENT '代币名称',
    contract_symbol VARCHAR(50) DEFAULT NULL COMMENT '代币符号',
    contract_type VARCHAR(50) DEFAULT NULL COMMENT '合约类型(ERC20/ERC721/ERC1155/UNKNOWN等)',
    decimals TINYINT UNSIGNED DEFAULT 18 COMMENT '小数位数(ERC20使用)',
    total_supply DECIMAL(65,0) DEFAULT NULL COMMENT '总供应量(原始值，不考虑decimals)',
    deploy_tx_hash VARCHAR(66) DEFAULT NULL COMMENT '部署交易哈希',
    deploy_block_number BIGINT UNSIGNED DEFAULT NULL COMMENT '部署区块号',
    deploy_block_time DATETIME DEFAULT NULL COMMENT '部署时间',
    deployer_address VARCHAR(42) DEFAULT NULL COMMENT '部署者地址',
    verification_status TINYINT DEFAULT 0 COMMENT '验证状态: 0-未验证, 1-已验证, 2-验证失败',
    verified_functions TEXT DEFAULT NULL COMMENT '已验证的函数选择器列表(JSON格式)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_contract_address (contract_address),
    KEY idx_deploy_block_number (deploy_block_number),
    KEY idx_deploy_time (deploy_block_time),
    KEY idx_symbol (contract_symbol),
    KEY idx_contract_type (contract_type),
    KEY idx_verification_status (verification_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='合约表';

-- 2. 交易表 (transactions)
-- 存储与合约相关的交易信息
CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    tx_hash VARCHAR(66) NOT NULL COMMENT '交易哈希',
    block_number BIGINT UNSIGNED NOT NULL COMMENT '区块号',
    block_hash VARCHAR(66) DEFAULT NULL COMMENT '区块哈希',
    block_time DATETIME NOT NULL COMMENT '区块时间',
    tx_index INT UNSIGNED DEFAULT NULL COMMENT '交易在区块中的索引',
    from_address VARCHAR(42) DEFAULT NULL COMMENT '发送者地址',
    to_address VARCHAR(42) DEFAULT NULL COMMENT '接收者地址(合约地址)',
    contract_address VARCHAR(42) NOT NULL COMMENT '合约地址(外键关联contracts表)',
    func_selector VARCHAR(50) DEFAULT NULL COMMENT '函数选择器(4字节hex或特殊标记如nested_call，外键关联function_signatures表)',
    func_name VARCHAR(50) DEFAULT NULL COMMENT '函数名称(transfer/transferFrom/approve等)',
    value DECIMAL(65,0) DEFAULT NULL COMMENT '交易金额(原始值，不考虑decimals)',
    gas_limit BIGINT UNSIGNED DEFAULT NULL COMMENT 'Gas限制',
    gas_used BIGINT UNSIGNED DEFAULT NULL COMMENT '实际使用的Gas',
    gas_price DECIMAL(30,0) DEFAULT NULL COMMENT 'Gas价格',
    tx_fee DECIMAL(30,0) DEFAULT NULL COMMENT '交易手续费',
    status TINYINT DEFAULT 1 COMMENT '交易状态: 0-失败, 1-成功',
    tx_data TEXT DEFAULT NULL COMMENT '交易输入数据(hex)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_tx_hash (tx_hash),
    KEY idx_block_number (block_number),
    KEY idx_block_time (block_time),
    KEY idx_contract_address (contract_address),
    KEY idx_from_address (from_address),
    KEY idx_to_address (to_address),
    KEY idx_func_selector (func_selector),
    KEY idx_status (status),
    CONSTRAINT fk_tx_contract FOREIGN KEY (contract_address) REFERENCES contracts(contract_address) ON DELETE RESTRICT ON UPDATE CASCADE
    -- 注意：func_selector 外键约束已移除，因为需要支持特殊标记（如 "nested_call"）
    -- 这些特殊标记不在 function_signatures 表中
    -- CONSTRAINT fk_tx_func_selector FOREIGN KEY (func_selector) REFERENCES function_signatures(selector) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易表';

-- 3. 地址持有代币余额表 (token_balances)
-- 存储地址的代币余额信息
CREATE TABLE IF NOT EXISTS token_balances (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    address VARCHAR(42) NOT NULL COMMENT '地址',
    contract_address VARCHAR(42) NOT NULL COMMENT '合约地址(外键关联contracts表)',
    balance DECIMAL(65,0) NOT NULL DEFAULT 0 COMMENT '余额(原始值，不考虑decimals)',
    balance_formatted DECIMAL(30,18) DEFAULT NULL COMMENT '格式化后的余额(考虑decimals)',
    last_tx_hash VARCHAR(66) DEFAULT NULL COMMENT '最后一次交易哈希',
    last_tx_block_number BIGINT UNSIGNED DEFAULT NULL COMMENT '最后一次交易区块号',
    last_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_address_contract (address, contract_address),
    KEY idx_address (address),
    KEY idx_contract_address (contract_address),
    KEY idx_balance (balance),
    KEY idx_last_tx_block (last_tx_block_number),
    CONSTRAINT fk_balance_contract FOREIGN KEY (contract_address) REFERENCES contracts(contract_address) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='地址代币余额表';

-- 4. 事件表 (events)
-- 存储Transfer等事件信息
CREATE TABLE IF NOT EXISTS events (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    tx_hash VARCHAR(66) NOT NULL COMMENT '交易哈希(外键关联transactions表)',
    block_number BIGINT UNSIGNED NOT NULL COMMENT '区块号',
    block_time DATETIME NOT NULL COMMENT '区块时间',
    log_index INT UNSIGNED NOT NULL COMMENT '日志索引(在交易中的位置)',
    contract_address VARCHAR(42) NOT NULL COMMENT '合约地址(外键关联contracts表)',
    event_name VARCHAR(50) NOT NULL COMMENT '事件名称(Transfer/Approval等)',
    event_signature VARCHAR(66) DEFAULT NULL COMMENT '事件签名hash',
    from_address VARCHAR(42) DEFAULT NULL COMMENT 'From地址(Transfer事件)',
    to_address VARCHAR(42) DEFAULT NULL COMMENT 'To地址(Transfer事件)',
    value DECIMAL(65,0) DEFAULT NULL COMMENT '金额(原始值，不考虑decimals)',
    value_formatted DECIMAL(30,18) DEFAULT NULL COMMENT '格式化后的金额(考虑decimals)',
    owner_address VARCHAR(42) DEFAULT NULL COMMENT 'Owner地址(Approval事件)',
    spender_address VARCHAR(42) DEFAULT NULL COMMENT 'Spender地址(Approval事件)',
    amount DECIMAL(65,0) DEFAULT NULL COMMENT '授权金额(Approval事件)',
    topic0 VARCHAR(66) DEFAULT NULL COMMENT 'Topic0(事件签名)',
    topic1 VARCHAR(66) DEFAULT NULL COMMENT 'Topic1(通常是from或owner)',
    topic2 VARCHAR(66) DEFAULT NULL COMMENT 'Topic2(通常是to或spender)',
    topic3 VARCHAR(66) DEFAULT NULL COMMENT 'Topic3(如果有)',
    data TEXT DEFAULT NULL COMMENT '事件数据(hex)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_tx_log_index (tx_hash, log_index),
    KEY idx_block_number (block_number),
    KEY idx_block_time (block_time),
    KEY idx_contract_address (contract_address),
    KEY idx_event_name (event_name),
    KEY idx_from_address (from_address),
    KEY idx_to_address (to_address),
    KEY idx_owner_address (owner_address),
    KEY idx_spender_address (spender_address),
    KEY idx_topic0 (topic0),
    CONSTRAINT fk_event_tx FOREIGN KEY (tx_hash) REFERENCES transactions(tx_hash) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_event_contract FOREIGN KEY (contract_address) REFERENCES contracts(contract_address) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='事件表';

-- 创建视图：代币余额视图(格式化显示)
CREATE OR REPLACE VIEW v_token_balances AS
SELECT 
    tb.id,
    tb.address,
    tb.contract_address,
    c.contract_name,
    c.contract_symbol,
    c.decimals,
    tb.balance,
    tb.balance / POW(10, c.decimals) AS balance_formatted,
    tb.last_tx_hash,
    tb.last_tx_block_number,
    tb.last_updated_at,
    tb.created_at
FROM token_balances tb
INNER JOIN contracts c ON tb.contract_address = c.contract_address;

-- 创建视图：交易详情视图
CREATE OR REPLACE VIEW v_transaction_details AS
SELECT 
    t.id,
    t.tx_hash,
    t.block_number,
    t.block_time,
    t.from_address,
    t.to_address,
    t.contract_address,
    c.contract_name,
    c.contract_symbol,
    c.decimals,
    t.func_name,
    t.value,
    CASE 
        WHEN t.value IS NOT NULL AND c.decimals IS NOT NULL 
        THEN t.value / POW(10, c.decimals)
        ELSE NULL
    END AS value_formatted,
    t.gas_used,
    t.gas_price,
    t.tx_fee,
    t.status,
    t.created_at
FROM transactions t
LEFT JOIN contracts c ON t.contract_address = c.contract_address;

-- 创建视图：事件详情视图
CREATE OR REPLACE VIEW v_event_details AS
SELECT 
    e.id,
    e.tx_hash,
    e.block_number,
    e.block_time,
    e.contract_address,
    c.contract_name,
    c.contract_symbol,
    c.decimals,
    e.event_name,
    e.from_address,
    e.to_address,
    e.value,
    CASE 
        WHEN e.value IS NOT NULL AND c.decimals IS NOT NULL 
        THEN e.value / POW(10, c.decimals)
        ELSE NULL
    END AS value_formatted,
    e.owner_address,
    e.spender_address,
    e.amount,
    CASE 
        WHEN e.amount IS NOT NULL AND c.decimals IS NOT NULL 
        THEN e.amount / POW(10, c.decimals)
        ELSE NULL
    END AS amount_formatted,
    e.created_at
FROM events e
LEFT JOIN contracts c ON e.contract_address = c.contract_address;

-- 5. 扫描进度表 (scan_progress)
-- 记录扫描器处理进度，用于断点续传
CREATE TABLE IF NOT EXISTS scan_progress (
    id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '主键ID(固定为1，确保只有一条记录)',
    last_block_number BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最后处理的区块号',
    last_block_hash VARCHAR(66) DEFAULT NULL COMMENT '最后处理的区块哈希',
    last_block_time DATETIME DEFAULT NULL COMMENT '最后处理的区块时间',
    processed_tx_count BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '已处理的交易总数',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='扫描进度表';

-- 插入初始记录（如果不存在）
-- INSERT IGNORE INTO scan_progress (id, last_block_number, processed_tx_count) 
-- VALUES (1, 0, 0);

-- 插入常见的ERC20函数签名
INSERT INTO function_signatures (selector, function_name, function_signature, standard_type, function_type, description, is_standard) VALUES
-- ERC20 核心函数
('18160ddd', 'totalSupply', 'totalSupply()', 'ERC20', 'view', '返回代币总供应量', TRUE),
('70a08231', 'balanceOf', 'balanceOf(address)', 'ERC20', 'view', '查询地址余额', TRUE),
('a9059cbb', 'transfer', 'transfer(address,uint256)', 'ERC20', 'nonpayable', '转账代币', TRUE),
('23b872dd', 'transferFrom', 'transferFrom(address,address,uint256)', 'ERC20', 'nonpayable', '授权转账', TRUE),
('095ea7b3', 'approve', 'approve(address,uint256)', 'ERC20', 'nonpayable', '授权额度', TRUE),
('dd62ed3e', 'allowance', 'allowance(address,address)', 'ERC20', 'view', '查询授权额度', TRUE),
-- ERC20 可选函数
('06fdde03', 'name', 'name()', 'ERC20', 'view', '代币名称', TRUE),
('95d89b41', 'symbol', 'symbol()', 'ERC20', 'view', '代币符号', TRUE),
('313ce567', 'decimals', 'decimals()', 'ERC20', 'view', '小数位数', TRUE),
-- ERC20 扩展函数
('39509351', 'increaseAllowance', 'increaseAllowance(address,uint256)', 'ERC20', 'nonpayable', '增加授权额度', TRUE),
('a457c2d7', 'decreaseAllowance', 'decreaseAllowance(address,uint256)', 'ERC20', 'nonpayable', '减少授权额度', TRUE),
-- ERC721 核心函数
('6352211e', 'ownerOf', 'ownerOf(uint256)', 'ERC721', 'view', '查询NFT所有者', TRUE),
('42842e0e', 'safeTransferFrom', 'safeTransferFrom(address,address,uint256)', 'ERC721', 'nonpayable', '安全转账NFT', TRUE),
('b88d4fde', 'safeTransferFrom', 'safeTransferFrom(address,address,uint256,bytes)', 'ERC721', 'nonpayable', '安全转账NFT(带数据)', TRUE),
('42966c68', 'burn', 'burn(uint256)', 'ERC721', 'nonpayable', '销毁NFT', FALSE),
-- ERC1155 核心函数
('00fdd58e', 'balanceOf', 'balanceOf(address,uint256)', 'ERC1155', 'view', '查询地址的特定代币余额', TRUE),
('4e1273f4', 'balanceOfBatch', 'balanceOfBatch(address[],uint256[])', 'ERC1155', 'view', '批量查询余额', TRUE),
('f242432a', 'safeTransferFrom', 'safeTransferFrom(address,address,uint256,uint256,bytes)', 'ERC1155', 'nonpayable', '安全转账', TRUE),
('2eb2c2d6', 'safeBatchTransferFrom', 'safeBatchTransferFrom(address,address,uint256[],uint256[],bytes)', 'ERC1155', 'nonpayable', '批量安全转账', TRUE),
('e985e9c5', 'isApprovedForAll', 'isApprovedForAll(address,address)', 'ERC1155', 'view', '查询是否全部授权', TRUE),
('a22cb465', 'setApprovalForAll', 'setApprovalForAll(address,bool)', 'ERC1155', 'nonpayable', '设置全部授权', TRUE),
('0e89341c', 'uri', 'uri(uint256)', 'ERC1155', 'view', '获取代币URI', TRUE)
ON DUPLICATE KEY UPDATE
    function_name = VALUES(function_name),
    function_signature = VALUES(function_signature),
    standard_type = VALUES(standard_type),
    updated_at = CURRENT_TIMESTAMP;

