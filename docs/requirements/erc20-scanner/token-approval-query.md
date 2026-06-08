# Token 授权查询

## 背景

用户需要知道自己的地址对各个 ERC20 合约授予了哪些授权（approve 了哪些 spender、额度多少），以便及时取消不必要的授权，降低安全风险。

取消授权或修改额度直接在链上操作（调用 `approve(spender, newAmount)`），本功能只负责**数据采集与查询展示**。

## 解决方案概览

| 步骤 | 做了什么 | 对应 |
|------|---------|------|
| **1. 发现** | Scanner 解析 Approval 事件（当前只解析 Transfer，Approval 被跳过） | 改造 `parseERC20Transfer()` |
| **2. 存储** | 新增 `token_allowances` 表，记录每条授权的当前状态 | 新增 DB 表和写入逻辑 |
| **3. 查询** | `GET /evmapi/accounts/{address}/approvals` 按 owner 查所有授权 | 新增 RPC 接口 |

```
扫块发现 Approval 事件 → 写入/更新 token_allowances 表（amount=0 则删除记录）
                                                              │
                            GET /evmapi/accounts/{address}/approvals  ← 查询接口
```

---

## 1. Approval 事件解析

### ERC20 Approval 事件

```
event Approval(address indexed owner, address indexed spender, uint256 value)
```

- `owner`：授权方（token 持有者）
- `spender`：被授权方（可以动用 token 的地址）
- `value`：授权额度

### 取消授权

**取消授权就是发一笔 `approve(spender, 0)` 交易，产生 `Approval(owner, spender, 0)` 事件。**

`increaseAllowance` / `decreaseAllowance` 同样 emit Approval 事件（value 为加减后的新额度），无需特殊处理。

> 因此在事件处理时：amount=0 → 删除 `token_allowances` 记录（授权已撤销）；amount>0 → upsert 记录。

### 扫描器改造

在 `parseERC20Transfer()` 中，当前只检查 Transfer 事件：

```go
// 现状：Approval 事件被跳过
if log.Topics[0] != transferEvent.ID {
    continue
}
```

改造方案：

1. 同时匹配 `Approval` 事件签名
2. 从 `log.Topics[1]` 提取 `owner`，`log.Topics[2]` 提取 `spender`
3. 从 `log.Data[:32]` 提取 `value`
4. 保存到 events 表 + 更新 `token_allowances` 表

```go
approvalEvent := parsedAbi.Events["Approval"]
// ...
// 匹配 Approval 事件（有 3 个 topics: [event_hash, owner, spender]）
if len(log.Topics) == 3 && log.Topics[0] == approvalEvent.ID {
    owner  := common.BytesToAddress(log.Topics[1].Bytes())
    spender := common.BytesToAddress(log.Topics[2].Bytes())
    value  := new(big.Int).SetBytes(log.Data[:32])
    // 保存事件 + 更新 token_allowances
}
```

---

## 2. 数据模型

### 新增表：token_allowances

| 字段 | 类型 | 说明 |
|------|------|------|
| `owner` | VARCHAR(42) | 授权方地址（小写，联合主键） |
| `spender` | VARCHAR(42) | 被授权方地址（小写，联合主键） |
| `contract_address` | VARCHAR(42) | ERC20 合约地址（小写，联合主键） |
| `amount` | DECIMAL(65,0) | 授权额度（原始值） |
| `last_tx_hash` | VARCHAR(66) | 最近一笔 Approval 事件 tx hash |
| `last_block_number` | BIGINT | 最近一笔 Approval 事件区块号 |
| `last_updated_at` | TIMESTAMP | 更新时间 |
| `created_at` | TIMESTAMP | 创建时间 |

唯一键：`(owner, spender, contract_address)`

### 写入逻辑

```
Approval 事件
  ├── amount > 0 → INSERT ... ON DUPLICATE KEY UPDATE amount, last_tx_hash, ...
  └── amount = 0 → DELETE WHERE owner=? AND spender=? AND contract_address=?
```

### 为什么不需要后台定期刷新

与 `token_balances` 不同，授权额度**不会因外部事件而变化**：

| | token_balances | token_allowances |
|---|---|---|
| 变化来源 | Transfer 事件（from/to 双方余额同时变） | Approval 事件（仅 owner 主动操作） |
| 差额累积 | 有（余额 = 初始 ± 累计 transfer） | 无（额度 = 最后一次 approve 的值） |
| 丢失事件后果 | 余额永久不准 | 只错过最后一次 approve，下次事件会修正 |
| 是否需要定期刷新 | 可选（有 skipInlineBalanceUpdate 模式） | **不需要** |

### 查询时按需刷新（可选增强）

如果担心遗漏事件导致数据不准，可以在 RPC 查询时：

```
1. 从 DB 读取 token_allowances（可能有延迟/遗漏）
2. 可选：对最近更新的记录，异步调用链上 allowance(owner, spender) 校验并修正
3. 返回结果
```

这样不增加扫块负担，且链上数据始终是最终真相来源。

---

## 3. 查询接口

### GET /evmapi/accounts/{address}/approvals

查询某地址的所有授权记录（该地址作为 owner 授权了哪些 spender）。

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `address` | path | 是 | owner 地址（0x 开头，42 字符） |
| `page` | query | 否 | 页码，默认 1 |
| `size` | query | 否 | 每页条数，默认 20，最大 100 |
| `min_amount` | query | 否 | 最低授权额度过滤（原始值字符串） |
| `contract_address` | query | 否 | 按 ERC20 合约地址过滤 |

**响应格式**：

```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "approvals": [
      {
        "contract_address": "0x...",
        "contract_name": "Tether USD",
        "contract_symbol": "USDT",
        "decimals": 6,
        "spender": "0x...",
        "amount": "115792089237316195423570985008687907853269984665640564039457584007913129639935",
        "amount_formatted": "unlimited",
        "last_tx_hash": "0x...",
        "last_block_number": 12345678,
        "last_updated": "2026-01-01T00:00:00Z"
      }
    ],
    "page": 1,
    "size": 20,
    "total": 5
  }
}
```

**SQL 查询**：

```sql
SELECT ta.owner, ta.spender, ta.contract_address, ta.amount,
       ta.last_tx_hash, ta.last_block_number, ta.last_updated_at,
       c.decimals, c.contract_name, c.contract_symbol
FROM token_allowances ta
LEFT JOIN contracts c ON ta.contract_address = c.contract_address
WHERE ta.owner = ?
  AND ta.amount >= ?            -- 可选 min_amount
  AND ta.contract_address = ?   -- 可选 contract_address
ORDER BY ta.amount DESC
LIMIT ? OFFSET ?
```

### 补充接口：按 spender 反查

**GET /evmapi/accounts/{address}/approved-by** — 查询哪些地址授权给了某个 spender（反向视角），按需增加。

---

## 4. 实现要点

### 4.1 扫描器改造清单

| 改动 | 文件 | 说明 |
|------|------|------|
| Approval 事件匹配 | `scanner/process.go` | `parseERC20Transfer()` 增加 Approval 分支 |
| 保存 Approval 事件 | `scanner/process.go` | 已有 `saveEventToDB()`，扩展支持 Approval |
| 写入 token_allowances | `scanner/process.go` | 新增 `updateAllowanceInDB()`；amount=0 时删除 |
| 数据库方法 | `database/models.go` | `UpsertAllowance()` + `DeleteAllowance()` |
| 按 owner 查授权 | `database/models.go` | `ListAllowancesByOwner()` |

### 4.2 地址规范化

与现有逻辑一致：所有地址统一 lowercase 存储。

### 4.3 无限授权

`amount = 2^256 - 1`（即 `uint256` 最大值）表示"无限授权"，前端展示时可格式化为 `unlimited`。

---

## 5. 非功能需求

- **地址规范化**：所有地址统一 lowercase，与现有 contracts/token_balances 保持一致
- **amount=0 即删除**：减少无效数据，保持表干净
- **无后台刷新**：与 token_balances 不同，不需要 balance_refresher 机制
- **查询性能**：`token_allowances` 表建立 `(owner, amount)` 联合索引
