# Token 持币排行

## 背景

需要知道每个 ERC20 代币的持币地址分布：谁持有多少，按余额排序。这是链上数据查询的基础需求，可用于榜单展示、大户监控、持仓分析等场景。

## 解决方案概览

整体分三步：**发现 → 扫描 → 查询接口**。

| 步骤 | 做了什么 | 关键代码 |
|------|---------|---------|
| **1. 发现** | 扫块解析 Transfer 事件，自动识别 ERC20 合约并入库 | `parseERC20Transfer()` |
| **2. 扫描** | 调用链上 `balanceOf` 获取余额，写入 `token_balances` 表；支持内联更新和后台批量刷新两种模式 | `getBalanceFromContract()` / `runBalanceRefresher()` |
| **3. 接口** | `GET /evmapi/contracts/{address}/holders` 查询持币排行，`min_balance` 过滤 + 分页 + 余额降序 | `handleContractHolders()` |

```
扫块发现 Transfer 事件 → 调用 balanceOf 获取余额 → 写入 token_balances 表
                                                              │
                                     GET /evmapi/contracts/{address}/holders  ← 查询接口
```

---

## 1. 发现：扫描区块中的 Transfer 事件

### 触发路径

- `Process.Start()` / `Process.StartWithEsClient()` → 逐块扫描
- `Process.ParaseBlock()` / `Process.parseBlockFromES()` → 遍历 EVM 交易
- `Process.processTransactionWithReceipt()` → 获取 receipt，判断交易类型
- `Process.parseERC20Transfer()` → 解析 receipt.Logs 中的 Transfer 事件

### Transfer 事件提取

在 [process.go:771-823](erc20Scaner/scanner/process.go#L771-L823) 中：

1. 遍历 `receipt.Logs`，匹配 `Transfer(address,address,uint256)` 事件签名
2. 从 `log.Topics[1]` 提取 `from` 地址，`log.Topics[2]` 提取 `to` 地址
3. 从 `log.Data[:32]` 提取 `value`（转账数量）
4. 调用 `getTokenInfo()` 获取代币的 name / symbol / decimals

### 合约发现

对于 Transfer 事件中遇到的新合约地址，自动检测是否为 ERC20：
- 检查字节码中的 ERC20 函数选择器（6 个核心 + 3 个可选）
- 是 ERC20 则调用 `name()`/`symbol()`/`decimals()`/`totalSupply()` 写入 `contracts` 表
- 不是则标记为 `UNKNOWN` 类型

---

## 2. 采集余额：两种模式

|  | 模式 A：扫块内联更新（默认） | 模式 B：后台批量刷新 |
|---|---|---|
| **配置** | `skipInlineBalanceUpdate = false` | `skipInlineBalanceUpdate = true` |
| **触发** | 每个 Transfer 事件后立即 `balanceOf` | Transfer 事件只插占位行(balance=0)，`balance_refresher` 定时批量刷 |
| **余额实时性** | 实时准确 | 有延迟（取决于刷新间隔） |
| **RPC 消耗** | 高（每个 Transfer 2 次 balanceOf） | 低（合并批量，减少重复调用） |
| **扫块速度** | 受节点响应限制 | 快，不阻塞 |
| **适用场景** | 数据量小、追求实时 | 数据量大、可接受延迟 |

### 模式 A：扫块内联更新（默认）

`skipInlineBalanceUpdate = false`

每处理一个 Transfer 事件后，立即调用链上 `balanceOf()` 获取发送方和接收方的最新余额，写入 `token_balances` 表。

```
Transfer 事件 → balanceOf(from) → UPDATE token_balances SET balance = ?
               → balanceOf(to)   → UPDATE token_balances SET balance = ?
```

**优点**：余额实时准确。
**缺点**：每个 Transfer 多 2 次链上 RPC 调用，扫块速度受限于节点响应。

### 模式 B：后台批量刷新

`skipInlineBalanceUpdate = true`

Transfer 事件仅插入占位行（balance=0），由 `balance_refresher` 定时任务批量从链上 `balanceOf` 刷新真实余额。

```
Transfer 事件 → INSERT token_balances (balance=0)   // 只记元数据
balance_refresher 定时触发 → 批量 balanceOf → UPDATE token_balances SET balance = ?
```

**优点**：扫块快，balanceOf 调用合并批量，减少 RPC 压力。
**缺点**：余额有延迟（取决于刷新间隔）。

### balance_refresher 配置

```yaml
balance_refresher:
  enabled: true
  interval: "300s"        # 刷新间隔
  batch_size: 50          # 每批取多少行
  concurrency: 5          # 并发 balanceOf 调用数
  min_age: "60s"          # 只刷新 last_updated_at 超过此时间的行
```

---

## 3. 数据模型

### token_balances 表

| 字段 | 类型 | 说明 |
|------|------|------|
| `address` | VARCHAR | 持有者地址（小写，联合主键） |
| `contract_address` | VARCHAR | ERC20 合约地址（小写，联合主键） |
| `balance` | DECIMAL(65,0) | 持有数量（原始值，未除 decimals） |
| `last_tx_hash` | VARCHAR | 最近一笔相关交易 hash |
| `last_tx_block_number` | BIGINT | 最近一笔相关交易区块号 |
| `last_updated_at` | TIMESTAMP | 余额最后更新时间 |
| `created_at` | TIMESTAMP | 记录创建时间 |

唯一键：`(address, contract_address)`

### 关联表

- `contracts`：提供代币的 decimals / symbol / name，查询时 JOIN 用于余额格式化

---

## 4. 查询接口

### GET /evmapi/contracts/{address}/holders

按代币合约查询持币排行（按余额降序）。

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `address` | path | 是 | ERC20 合约地址（0x 开头，42 字符） |
| `page` | query | 否 | 页码，默认 1 |
| `size` | query | 否 | 每页条数，默认 20，最大 100 |
| `min_balance` | query | 否 | 最低余额过滤（原始值字符串） |

**响应格式**：

```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "holders": [
      {
        "address": "0x...",
        "balance": "1000000000000000000",
        "balance_formatted": "1.0 USDT",
        "last_tx_hash": "0x...",
        "last_tx_block": 12345678,
        "last_updated": "2026-01-01T00:00:00Z"
      }
    ],
    "page": 1,
    "size": 20,
    "total": 1523
  }
}
```

**实现位置**：[contract.go:420-515](erc20Scaner/rpc/contract.go#L420-L515)

**SQL 查询**：

```sql
SELECT tb.address, tb.balance, tb.last_tx_hash,
       tb.last_tx_block_number, tb.last_updated_at, c.decimals
FROM token_balances tb
LEFT JOIN contracts c ON tb.contract_address = c.contract_address
WHERE tb.contract_address = ?
  AND tb.balance >= ?          -- 可选 min_balance 过滤
ORDER BY tb.balance DESC
LIMIT ? OFFSET ?
```

### 补充接口：按地址查持仓

**GET /evmapi/accounts/{address}/erc20-balances** — 反过来，查某个地址持有的所有 ERC20 及余额。

---

## 5. 非功能需求

- **地址规范化**：所有地址统一 lowercase 存储，避免大小写不一致导致重复行
- **分页上限**：单页最多 100 条，防止大查询拖垮 DB
- **余额精度**：`DECIMAL(65,0)` 支持超大整数（uint256 最大值），不丢精度
- **排序性能**：`token_balances` 表需建立 `(contract_address, balance)` 联合索引
