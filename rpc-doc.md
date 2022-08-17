# 存证展开服务RPC接口文档

> 👉 [Swagger 文档](http://172.16.101.87:9992/swagger/index.html)

## 接口概览（总计31个）

### Comm

| **路径** | **功能** | **请求方式** |
|---------|---------|-------------|
| [/v1/LastSeq](#获取当前最新同步以及解析的区块序列号) | 获取当前最新同步以及解析的区块序列号 | POST |
| [/v1/health](#获取服务运行状态和版本) | 获取服务运行状态和版本 | POST |
| [/v1/status](#获取服务详细状态信息) | 获取服务详细状态信息 | POST |

### EVM

| **路径** | **功能** | **请求方式** |
|---------|---------|-------------|
| [/v1/evm/Count](#查询通证数量) | 查询通证数量 | POST |
| [/v1/evm/List](#查询通证列表) | 查询通证列表 | POST |
| [/v1/evm/nft/account/Count](#查询账户信息数量) | 查询账户信息数量 | POST |
| [/v1/evm/nft/account/List](#查询账户信息列表) | 查询账户信息列表 | POST |
| [/v1/evm/nft/transfer/Count](#查询转账数量) | 查询转账数量 | POST |
| [/v1/evm/nft/transfer/List](#查询转账列表) | 查询转账列表 | POST |

### File

| **路径** | **功能** | **请求方式** |
|---------|---------|-------------|
| [/v1/file-clean-cache](#清理文件缓存) | 清理文件缓存 | GET |
| [/v1/file/{hash}](#获取上链文件) | 获取上链文件 | GET |

### Proof

| **路径** | **功能** | **请求方式** |
|---------|---------|-------------|
| [/v1/proof/Count](#获取存证数量) | 获取存证数量 | POST |
| [/v1/proof/CountByTime](#根据年/月/日对存证的数量进行统计) | 根据年/月/日对存证的数量进行统计 | POST |
| [/v1/proof/DonationStats](#获取捐款排名信息) | 获取捐款排名信息 | POST |
| [/v1/proof/FetchSource](#获取满足条件的数据的指定字段的值) | 获取满足条件的数据的指定字段的值 | POST |
| [/v1/proof/GetProofs](#获取多个指定hash的存证信息) | 获取多个指定hash的存证信息 | POST |
| [/v1/proof/GetTemplates](#获取多个指定hash的存证模板) | 获取多个指定hash的存证模板 | POST |
| [/v1/proof/Gets](#获取多个指定hash的存证信息) | 获取多个指定hash的存证信息 | POST |
| [/v1/proof/List](#获取存证列表) | 获取存证列表 | POST |
| [/v1/proof/ListUpdateProof](#获取最新存证列表) | 获取最新存证列表 | POST |
| [/v1/proof/ListUpdateRecord](#获取存证更新记录的列表) | 获取存证更新记录的列表 | POST |
| [/v1/proof/QueryStatsInfo](#获取统计项信息) | 获取统计项信息 | POST |
| [/v1/proof/Show](#获得指定hash的存证信息) | 获得指定hash的存证信息 | POST |
| [/v1/proof/TotalStats](#获取满足条件的数据的指定字段的总值) | 获取满足条件的数据的指定字段的总值 | POST |
| [/v1/proof/VolunteerStats](#获取志愿者的分布图按照省/单位) | 获取志愿者的分布图按照省/单位 | POST |

### proofmember

| **路径** | **功能** | **请求方式** |
|---------|---------|-------------|
| [/v1/proofmember/Count](#获得指定范围的用户的数量) | 获得指定范围的用户的数量 | POST |
| [/v1/proofmember/Gets](#获得指定地址的用户) | 获得指定地址的用户 | POST |
| [/v1/proofmember/List](#分页列出指定范围的用户) | 分页列出指定范围的用户 | POST |

### prooforganization

| **路径** | **功能** | **请求方式** |
|---------|---------|-------------|
| [/v1/prooforganization/Count](#获得指定范围的组织的数量) | 获得指定范围的组织的数量 | POST |
| [/v1/prooforganization/Gets](#获得指定的组织的信息) | 获得指定的组织的信息 | POST |
| [/v1/prooforganization/List](#分页列出指定范围的组织) | 分页列出指定范围的组织 | POST |

## 接口详情

### Comm

### 获取当前最新同步以及解析的区块序列号

[返回概览](#Comm)

POST /v1/LastSeq  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method"
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [rpcutils.RepLastSeq](#rpcutilsRepLastSeq) |  |
| &emsp; lastConvertSeq | 最新解析区块高度 | integer |  |
| &emsp; lastSyncSeq | 最新同步区块高度 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": {
    "lastConvertSeq": 1,
    "lastSyncSeq": 1
  }
}
```

### 获取服务运行状态和版本

[返回概览](#Comm)

POST /v1/health  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method"
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Health](#swaggerHealth) |  |
| &emsp; status | 状态 | string |  |
| &emsp; version | 版本 | string |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": {
    "status": "status",
    "version": "version"
  }
}
```

### 获取服务详细状态信息

[返回概览](#Comm)

POST /v1/status  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method"
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Status](#swaggerStatus) |  |
| &emsp; chain | 链状态 | [swagger.ChainStatus](#swaggerChainStatus) |  |
| &emsp; es | ElasticSearch状态 | [swagger.EsStatus](#swaggerEsStatus) |  |
| &emsp; server | 服务状态 | [swagger.ServerStatus](#swaggerServerStatus) |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": {
    "chain": {
      "coin": "coin",
      "push_seq": 1,
      "status": "status",
      "version": {
        "app": "app",
        "chain33": "chain33",
        "localDb": "localDb",
        "title": "title"
      }
    },
    "es": {
      "_nodes": {
        "failed": 1,
        "successful": 1,
        "total": 1
      },
      "cluster_name": "cluster_name",
      "nodes": null,
      "status": "status"
    },
    "server": {
      "coin": "coin",
      "conv_seq": 1,
      "sync_seq": 1,
      "title": "title",
      "version": "version"
    }
  }
}
```

### EVM

### 查询通证数量

[返回概览](#EVM)

POST /v1/evm/Count  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": 1
}
```

### 查询通证列表

[返回概览](#EVM)

POST /v1/evm/List  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.EVMToken](#swaggerEVMToken) array |  |
| &emsp; amount | 金额 | integer |  |
| &emsp; call_func_name | 调用方法名称 | string |  |
| &emsp; contract_addr | 合约地址 | string |  |
| &emsp; contract_used_gas | 消耗gas | integer |  |
| &emsp; evm_block_hash | 区块hash | string |  |
| &emsp; evm_block_time | 上链时间 | integer |  |
| &emsp; evm_events | evm事件 | string |  |
| &emsp; evm_height | 区块高度 | integer |  |
| &emsp; evm_height_index | 高度索引 | integer |  |
| &emsp; evm_note | 备注信息 | string |  |
| &emsp; evm_param | evm调用参数 | string |  |
| &emsp; evm_tx_hash | 交易hash | string |  |
| &emsp; goods_id | 物品唯一标识 | integer |  |
| &emsp; goods_type | 物品类型 | integer |  |
| &emsp; label_id | 物品标签id | string |  |
| &emsp; name | 物品名称 | string |  |
| &emsp; owner | 拥有者 | string |  |
| &emsp; publish_time | 发布时间 | integer |  |
| &emsp; publisher | 发布者 | string |  |
| &emsp; remark | 备注 | string |  |
| &emsp; source_hash | 关联交易hash | string array |  |
| &emsp; trace_hash | 关联溯源hash | string array |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "amount": 1,
      "call_func_name": "call_func_name",
      "contract_addr": "contract_addr",
      "contract_used_gas": 1,
      "evm_block_hash": "evm_block_hash",
      "evm_block_time": 1,
      "evm_events": "evm_events",
      "evm_height": 1,
      "evm_height_index": 1,
      "evm_note": "evm_note",
      "evm_param": "evm_param",
      "evm_tx_hash": "evm_tx_hash",
      "goods_id": 1,
      "goods_type": 1,
      "label_id": "label_id",
      "name": "name",
      "owner": "owner",
      "publish_time": 1,
      "publisher": "publisher",
      "remark": "remark",
      "source_hash": [
        "source_hash"
      ],
      "trace_hash": [
        "trace_hash"
      ]
    }
  ]
}
```

### 查询账户信息数量

[返回概览](#EVM)

POST /v1/evm/nft/account/Count  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": 1
}
```

### 查询账户信息列表

[返回概览](#EVM)

POST /v1/evm/nft/account/List  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.EVMToken](#swaggerEVMToken) array |  |
| &emsp; amount | 金额 | integer |  |
| &emsp; call_func_name | 调用方法名称 | string |  |
| &emsp; contract_addr | 合约地址 | string |  |
| &emsp; contract_used_gas | 消耗gas | integer |  |
| &emsp; evm_block_hash | 区块hash | string |  |
| &emsp; evm_block_time | 上链时间 | integer |  |
| &emsp; evm_events | evm事件 | string |  |
| &emsp; evm_height | 区块高度 | integer |  |
| &emsp; evm_height_index | 高度索引 | integer |  |
| &emsp; evm_note | 备注信息 | string |  |
| &emsp; evm_param | evm调用参数 | string |  |
| &emsp; evm_tx_hash | 交易hash | string |  |
| &emsp; goods_id | 物品唯一标识 | integer |  |
| &emsp; goods_type | 物品类型 | integer |  |
| &emsp; label_id | 物品标签id | string |  |
| &emsp; name | 物品名称 | string |  |
| &emsp; owner | 拥有者 | string |  |
| &emsp; publish_time | 发布时间 | integer |  |
| &emsp; publisher | 发布者 | string |  |
| &emsp; remark | 备注 | string |  |
| &emsp; source_hash | 关联交易hash | string array |  |
| &emsp; trace_hash | 关联溯源hash | string array |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "amount": 1,
      "call_func_name": "call_func_name",
      "contract_addr": "contract_addr",
      "contract_used_gas": 1,
      "evm_block_hash": "evm_block_hash",
      "evm_block_time": 1,
      "evm_events": "evm_events",
      "evm_height": 1,
      "evm_height_index": 1,
      "evm_note": "evm_note",
      "evm_param": "evm_param",
      "evm_tx_hash": "evm_tx_hash",
      "goods_id": 1,
      "goods_type": 1,
      "label_id": "label_id",
      "name": "name",
      "owner": "owner",
      "publish_time": 1,
      "publisher": "publisher",
      "remark": "remark",
      "source_hash": [
        "source_hash"
      ],
      "trace_hash": [
        "trace_hash"
      ]
    }
  ]
}
```

### 查询转账数量

[返回概览](#EVM)

POST /v1/evm/nft/transfer/Count  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": 1
}
```

### 查询转账列表

[返回概览](#EVM)

POST /v1/evm/nft/transfer/List  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.EVMToken](#swaggerEVMToken) array |  |
| &emsp; amount | 金额 | integer |  |
| &emsp; call_func_name | 调用方法名称 | string |  |
| &emsp; contract_addr | 合约地址 | string |  |
| &emsp; contract_used_gas | 消耗gas | integer |  |
| &emsp; evm_block_hash | 区块hash | string |  |
| &emsp; evm_block_time | 上链时间 | integer |  |
| &emsp; evm_events | evm事件 | string |  |
| &emsp; evm_height | 区块高度 | integer |  |
| &emsp; evm_height_index | 高度索引 | integer |  |
| &emsp; evm_note | 备注信息 | string |  |
| &emsp; evm_param | evm调用参数 | string |  |
| &emsp; evm_tx_hash | 交易hash | string |  |
| &emsp; goods_id | 物品唯一标识 | integer |  |
| &emsp; goods_type | 物品类型 | integer |  |
| &emsp; label_id | 物品标签id | string |  |
| &emsp; name | 物品名称 | string |  |
| &emsp; owner | 拥有者 | string |  |
| &emsp; publish_time | 发布时间 | integer |  |
| &emsp; publisher | 发布者 | string |  |
| &emsp; remark | 备注 | string |  |
| &emsp; source_hash | 关联交易hash | string array |  |
| &emsp; trace_hash | 关联溯源hash | string array |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "amount": 1,
      "call_func_name": "call_func_name",
      "contract_addr": "contract_addr",
      "contract_used_gas": 1,
      "evm_block_hash": "evm_block_hash",
      "evm_block_time": 1,
      "evm_events": "evm_events",
      "evm_height": 1,
      "evm_height_index": 1,
      "evm_note": "evm_note",
      "evm_param": "evm_param",
      "evm_tx_hash": "evm_tx_hash",
      "goods_id": 1,
      "goods_type": 1,
      "label_id": "label_id",
      "name": "name",
      "owner": "owner",
      "publish_time": 1,
      "publisher": "publisher",
      "remark": "remark",
      "source_hash": [
        "source_hash"
      ],
      "trace_hash": [
        "trace_hash"
      ]
    }
  ]
}
```

### File

### 清理文件缓存

[返回概览](#File)

GET /v1/file-clean-cache

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|

### 获取上链文件

[返回概览](#File)

GET /v1/file/{hash}

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| query | hash | 文件哈希 | string | 必填 |  |
| path | name | 文件名称 | string | 必填 |  |

请求示例：

```
Query:
/v1/file/{hash}?hash=hash
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|

### Proof

### 获取存证数量

[返回概览](#Proof)

POST /v1/proof/Count  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": 1
}
```

### 根据年/月/日对存证的数量进行统计

[返回概览](#Proof)

POST /v1/proof/CountByTime  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [rpcutils.CountByTime](#rpcutilsCountByTime) array | 非必填 |  |
| body | &emsp; match | 匹配条件 | [rpcutils.QMatch](#rpcutilsQMatch) array | 非必填 |  |
| body | &emsp; ranges | 范围 | [rpcutils.QRanges](#rpcutilsQRanges) | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "ranges": {
        "key": "key",
        "ranges": [
          {
            "end": null,
            "start": null
          }
        ]
      }
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | object |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": null
}
```

### 获取捐款排名信息

[返回概览](#Proof)

POST /v1/proof/DonationStats  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [rpcutils.DonationStats](#rpcutilsDonationStats) array | 非必填 |  |
| body | &emsp; match | 匹配条件 | [rpcutils.QMatch](#rpcutilsQMatch) array | 非必填 |  |
| body | &emsp; subSumAgg | 子聚合字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) | 非必填 |  |
| body | &emsp; termsAgg | 聚合字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "subSumAgg": {
        "key": "key"
      },
      "termsAgg": {
        "key": "key"
      }
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.DonationStats](#swaggerDonationStats) |  |
| &emsp; itemes | 列表 | [swagger.DonationStatItem](#swaggerDonationStatItem) array |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": {
    "itemes": [
      {
        "count": 1,
        "name": "name",
        "total": 1
      }
    ]
  }
}
```

### 获取满足条件的数据的指定字段的值

[返回概览](#Proof)

POST /v1/proof/FetchSource  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [rpcutils.SpecifiedFields](#rpcutilsSpecifiedFields) array | 非必填 |  |
| body | &emsp; count | 总量 | integer | 非必填 |  |
| body | &emsp; fields | 字段列表 | string array | 非必填 |  |
| body | &emsp; match | 匹配条件 | [rpcutils.QMatch](#rpcutilsQMatch) array | 非必填 |  |
| body | &emsp; sort | 排序 | [rpcutils.QSort](#rpcutilsQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "count": 1,
      "fields": [
        "fields"
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | string array |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    "result"
  ]
}
```

### 获取多个指定hash的存证信息

[返回概览](#Proof)

POST /v1/proof/GetProofs  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [rpcutils.Hashes](#rpcutilsHashes) array | 非必填 |  |
| body | &emsp; hash | 哈希列表 | string array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "hash": [
        "hash"
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Proof](#swaggerProof) array |  |
| &emsp; basehash | 增量存证依赖的主hash | string |  |
| &emsp; evidenceName | 存证名称 | string |  |
| &emsp; prehash | 增量存证前一个hash | string |  |
| &emsp; proof_block_hash | 区块hash | string |  |
| &emsp; proof_block_time | 上链时间 | integer |  |
| &emsp; proof_data | 存证数据 | string |  |
| &emsp; proof_deleted | 删除存证交易hash | string |  |
| &emsp; proof_deleted_flag | 删除标志 | boolean |  |
| &emsp; proof_deleted_note | 删除备注 | string |  |
| &emsp; proof_height | 存证高度 | integer |  |
| &emsp; proof_height_index | 存证高度索引 | integer |  |
| &emsp; proof_id | 存证id | string |  |
| &emsp; proof_note | 存证备注 | string |  |
| &emsp; proof_organization | 组织 | string |  |
| &emsp; proof_original | 来源 | string |  |
| &emsp; proof_sender | 存证发起者 | string |  |
| &emsp; proof_tx_hash | 交易哈希 | string |  |
| &emsp; source_hash | 依赖交易哈希 | object |  |
| &emsp; update_hash | 更新依赖主哈希 | string |  |
| &emsp; update_version | 更新版本 | integer |  |
| &emsp; user_auth_type | 用户认证类型 | integer |  |
| &emsp; user_email | 用户邮箱 | string |  |
| &emsp; user_enterprise_name | 用户企业名称 | string |  |
| &emsp; user_icon | 用户头像链接地址 | string |  |
| &emsp; user_name | 用户名 | string |  |
| &emsp; user_phone | 用户手机号 | string |  |
| &emsp; user_real_name | 用户真是名称 | string |  |
| &emsp; version | 存证版本 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "basehash": "basehash",
      "evidenceName": "evidenceName",
      "prehash": "prehash",
      "proof_block_hash": "proof_block_hash",
      "proof_block_time": 1,
      "proof_data": "proof_data",
      "proof_deleted": "proof_deleted",
      "proof_deleted_flag": false,
      "proof_deleted_note": "proof_deleted_note",
      "proof_height": 1,
      "proof_height_index": 1,
      "proof_id": "proof_id",
      "proof_note": "proof_note",
      "proof_organization": "proof_organization",
      "proof_original": "proof_original",
      "proof_sender": "proof_sender",
      "proof_tx_hash": "proof_tx_hash",
      "source_hash": null,
      "update_hash": "update_hash",
      "update_version": 1,
      "user_auth_type": 1,
      "user_email": "user_email",
      "user_enterprise_name": "user_enterprise_name",
      "user_icon": "user_icon",
      "user_name": "user_name",
      "user_phone": "user_phone",
      "user_real_name": "user_real_name",
      "version": 1
    }
  ]
}
```

### 获取多个指定hash的存证模板

[返回概览](#Proof)

POST /v1/proof/GetTemplates  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [rpcutils.Hashes](#rpcutilsHashes) array | 非必填 |  |
| body | &emsp; hash | 哈希列表 | string array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "hash": [
        "hash"
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Template](#swaggerTemplate) array |  |
| &emsp; template_block_hash | 区块哈希 | string |  |
| &emsp; template_block_time | 上链时间 | integer |  |
| &emsp; template_data | 模板数据 | string |  |
| &emsp; template_deleted | 删除交易哈希 | string |  |
| &emsp; template_deleted_flag | 删除标志 | boolean |  |
| &emsp; template_deleted_note | 删除备注 | string |  |
| &emsp; template_height | 高度 | integer |  |
| &emsp; template_height_index | 高度索引 | integer |  |
| &emsp; template_id | 模板id | string |  |
| &emsp; template_name | 模板名称 | string |  |
| &emsp; template_organization | 组织 | string |  |
| &emsp; template_sender | 交易发送人 | string |  |
| &emsp; template_tx_hash | 交易哈希 | string |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "template_block_hash": "template_block_hash",
      "template_block_time": 1,
      "template_data": "template_data",
      "template_deleted": "template_deleted",
      "template_deleted_flag": false,
      "template_deleted_note": "template_deleted_note",
      "template_height": 1,
      "template_height_index": 1,
      "template_id": "template_id",
      "template_name": "template_name",
      "template_organization": "template_organization",
      "template_sender": "template_sender",
      "template_tx_hash": "template_tx_hash"
    }
  ]
}
```

### 获取多个指定hash的存证信息

[返回概览](#Proof)

POST /v1/proof/Gets  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [rpcutils.Hashes](#rpcutilsHashes) array | 非必填 |  |
| body | &emsp; hash | 哈希列表 | string array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "hash": [
        "hash"
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Proof](#swaggerProof) array |  |
| &emsp; basehash | 增量存证依赖的主hash | string |  |
| &emsp; evidenceName | 存证名称 | string |  |
| &emsp; prehash | 增量存证前一个hash | string |  |
| &emsp; proof_block_hash | 区块hash | string |  |
| &emsp; proof_block_time | 上链时间 | integer |  |
| &emsp; proof_data | 存证数据 | string |  |
| &emsp; proof_deleted | 删除存证交易hash | string |  |
| &emsp; proof_deleted_flag | 删除标志 | boolean |  |
| &emsp; proof_deleted_note | 删除备注 | string |  |
| &emsp; proof_height | 存证高度 | integer |  |
| &emsp; proof_height_index | 存证高度索引 | integer |  |
| &emsp; proof_id | 存证id | string |  |
| &emsp; proof_note | 存证备注 | string |  |
| &emsp; proof_organization | 组织 | string |  |
| &emsp; proof_original | 来源 | string |  |
| &emsp; proof_sender | 存证发起者 | string |  |
| &emsp; proof_tx_hash | 交易哈希 | string |  |
| &emsp; source_hash | 依赖交易哈希 | object |  |
| &emsp; update_hash | 更新依赖主哈希 | string |  |
| &emsp; update_version | 更新版本 | integer |  |
| &emsp; user_auth_type | 用户认证类型 | integer |  |
| &emsp; user_email | 用户邮箱 | string |  |
| &emsp; user_enterprise_name | 用户企业名称 | string |  |
| &emsp; user_icon | 用户头像链接地址 | string |  |
| &emsp; user_name | 用户名 | string |  |
| &emsp; user_phone | 用户手机号 | string |  |
| &emsp; user_real_name | 用户真是名称 | string |  |
| &emsp; version | 存证版本 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "basehash": "basehash",
      "evidenceName": "evidenceName",
      "prehash": "prehash",
      "proof_block_hash": "proof_block_hash",
      "proof_block_time": 1,
      "proof_data": "proof_data",
      "proof_deleted": "proof_deleted",
      "proof_deleted_flag": false,
      "proof_deleted_note": "proof_deleted_note",
      "proof_height": 1,
      "proof_height_index": 1,
      "proof_id": "proof_id",
      "proof_note": "proof_note",
      "proof_organization": "proof_organization",
      "proof_original": "proof_original",
      "proof_sender": "proof_sender",
      "proof_tx_hash": "proof_tx_hash",
      "source_hash": null,
      "update_hash": "update_hash",
      "update_version": 1,
      "user_auth_type": 1,
      "user_email": "user_email",
      "user_enterprise_name": "user_enterprise_name",
      "user_icon": "user_icon",
      "user_name": "user_name",
      "user_phone": "user_phone",
      "user_real_name": "user_real_name",
      "version": 1
    }
  ]
}
```

### 获取存证列表

[返回概览](#Proof)

POST /v1/proof/List  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Proof](#swaggerProof) array |  |
| &emsp; basehash | 增量存证依赖的主hash | string |  |
| &emsp; evidenceName | 存证名称 | string |  |
| &emsp; prehash | 增量存证前一个hash | string |  |
| &emsp; proof_block_hash | 区块hash | string |  |
| &emsp; proof_block_time | 上链时间 | integer |  |
| &emsp; proof_data | 存证数据 | string |  |
| &emsp; proof_deleted | 删除存证交易hash | string |  |
| &emsp; proof_deleted_flag | 删除标志 | boolean |  |
| &emsp; proof_deleted_note | 删除备注 | string |  |
| &emsp; proof_height | 存证高度 | integer |  |
| &emsp; proof_height_index | 存证高度索引 | integer |  |
| &emsp; proof_id | 存证id | string |  |
| &emsp; proof_note | 存证备注 | string |  |
| &emsp; proof_organization | 组织 | string |  |
| &emsp; proof_original | 来源 | string |  |
| &emsp; proof_sender | 存证发起者 | string |  |
| &emsp; proof_tx_hash | 交易哈希 | string |  |
| &emsp; source_hash | 依赖交易哈希 | object |  |
| &emsp; update_hash | 更新依赖主哈希 | string |  |
| &emsp; update_version | 更新版本 | integer |  |
| &emsp; user_auth_type | 用户认证类型 | integer |  |
| &emsp; user_email | 用户邮箱 | string |  |
| &emsp; user_enterprise_name | 用户企业名称 | string |  |
| &emsp; user_icon | 用户头像链接地址 | string |  |
| &emsp; user_name | 用户名 | string |  |
| &emsp; user_phone | 用户手机号 | string |  |
| &emsp; user_real_name | 用户真是名称 | string |  |
| &emsp; version | 存证版本 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "basehash": "basehash",
      "evidenceName": "evidenceName",
      "prehash": "prehash",
      "proof_block_hash": "proof_block_hash",
      "proof_block_time": 1,
      "proof_data": "proof_data",
      "proof_deleted": "proof_deleted",
      "proof_deleted_flag": false,
      "proof_deleted_note": "proof_deleted_note",
      "proof_height": 1,
      "proof_height_index": 1,
      "proof_id": "proof_id",
      "proof_note": "proof_note",
      "proof_organization": "proof_organization",
      "proof_original": "proof_original",
      "proof_sender": "proof_sender",
      "proof_tx_hash": "proof_tx_hash",
      "source_hash": null,
      "update_hash": "update_hash",
      "update_version": 1,
      "user_auth_type": 1,
      "user_email": "user_email",
      "user_enterprise_name": "user_enterprise_name",
      "user_icon": "user_icon",
      "user_name": "user_name",
      "user_phone": "user_phone",
      "user_real_name": "user_real_name",
      "version": 1
    }
  ]
}
```

### 获取最新存证列表

[返回概览](#Proof)

POST /v1/proof/ListUpdateProof  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Proof](#swaggerProof) array |  |
| &emsp; basehash | 增量存证依赖的主hash | string |  |
| &emsp; evidenceName | 存证名称 | string |  |
| &emsp; prehash | 增量存证前一个hash | string |  |
| &emsp; proof_block_hash | 区块hash | string |  |
| &emsp; proof_block_time | 上链时间 | integer |  |
| &emsp; proof_data | 存证数据 | string |  |
| &emsp; proof_deleted | 删除存证交易hash | string |  |
| &emsp; proof_deleted_flag | 删除标志 | boolean |  |
| &emsp; proof_deleted_note | 删除备注 | string |  |
| &emsp; proof_height | 存证高度 | integer |  |
| &emsp; proof_height_index | 存证高度索引 | integer |  |
| &emsp; proof_id | 存证id | string |  |
| &emsp; proof_note | 存证备注 | string |  |
| &emsp; proof_organization | 组织 | string |  |
| &emsp; proof_original | 来源 | string |  |
| &emsp; proof_sender | 存证发起者 | string |  |
| &emsp; proof_tx_hash | 交易哈希 | string |  |
| &emsp; source_hash | 依赖交易哈希 | object |  |
| &emsp; update_hash | 更新依赖主哈希 | string |  |
| &emsp; update_version | 更新版本 | integer |  |
| &emsp; user_auth_type | 用户认证类型 | integer |  |
| &emsp; user_email | 用户邮箱 | string |  |
| &emsp; user_enterprise_name | 用户企业名称 | string |  |
| &emsp; user_icon | 用户头像链接地址 | string |  |
| &emsp; user_name | 用户名 | string |  |
| &emsp; user_phone | 用户手机号 | string |  |
| &emsp; user_real_name | 用户真是名称 | string |  |
| &emsp; version | 存证版本 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "basehash": "basehash",
      "evidenceName": "evidenceName",
      "prehash": "prehash",
      "proof_block_hash": "proof_block_hash",
      "proof_block_time": 1,
      "proof_data": "proof_data",
      "proof_deleted": "proof_deleted",
      "proof_deleted_flag": false,
      "proof_deleted_note": "proof_deleted_note",
      "proof_height": 1,
      "proof_height_index": 1,
      "proof_id": "proof_id",
      "proof_note": "proof_note",
      "proof_organization": "proof_organization",
      "proof_original": "proof_original",
      "proof_sender": "proof_sender",
      "proof_tx_hash": "proof_tx_hash",
      "source_hash": null,
      "update_hash": "update_hash",
      "update_version": 1,
      "user_auth_type": 1,
      "user_email": "user_email",
      "user_enterprise_name": "user_enterprise_name",
      "user_icon": "user_icon",
      "user_name": "user_name",
      "user_phone": "user_phone",
      "user_real_name": "user_real_name",
      "version": 1
    }
  ]
}
```

### 获取存证更新记录的列表

[返回概览](#Proof)

POST /v1/proof/ListUpdateRecord  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Proof](#swaggerProof) array |  |
| &emsp; basehash | 增量存证依赖的主hash | string |  |
| &emsp; evidenceName | 存证名称 | string |  |
| &emsp; prehash | 增量存证前一个hash | string |  |
| &emsp; proof_block_hash | 区块hash | string |  |
| &emsp; proof_block_time | 上链时间 | integer |  |
| &emsp; proof_data | 存证数据 | string |  |
| &emsp; proof_deleted | 删除存证交易hash | string |  |
| &emsp; proof_deleted_flag | 删除标志 | boolean |  |
| &emsp; proof_deleted_note | 删除备注 | string |  |
| &emsp; proof_height | 存证高度 | integer |  |
| &emsp; proof_height_index | 存证高度索引 | integer |  |
| &emsp; proof_id | 存证id | string |  |
| &emsp; proof_note | 存证备注 | string |  |
| &emsp; proof_organization | 组织 | string |  |
| &emsp; proof_original | 来源 | string |  |
| &emsp; proof_sender | 存证发起者 | string |  |
| &emsp; proof_tx_hash | 交易哈希 | string |  |
| &emsp; source_hash | 依赖交易哈希 | object |  |
| &emsp; update_hash | 更新依赖主哈希 | string |  |
| &emsp; update_version | 更新版本 | integer |  |
| &emsp; user_auth_type | 用户认证类型 | integer |  |
| &emsp; user_email | 用户邮箱 | string |  |
| &emsp; user_enterprise_name | 用户企业名称 | string |  |
| &emsp; user_icon | 用户头像链接地址 | string |  |
| &emsp; user_name | 用户名 | string |  |
| &emsp; user_phone | 用户手机号 | string |  |
| &emsp; user_real_name | 用户真是名称 | string |  |
| &emsp; version | 存证版本 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "basehash": "basehash",
      "evidenceName": "evidenceName",
      "prehash": "prehash",
      "proof_block_hash": "proof_block_hash",
      "proof_block_time": 1,
      "proof_data": "proof_data",
      "proof_deleted": "proof_deleted",
      "proof_deleted_flag": false,
      "proof_deleted_note": "proof_deleted_note",
      "proof_height": 1,
      "proof_height_index": 1,
      "proof_id": "proof_id",
      "proof_note": "proof_note",
      "proof_organization": "proof_organization",
      "proof_original": "proof_original",
      "proof_sender": "proof_sender",
      "proof_tx_hash": "proof_tx_hash",
      "source_hash": null,
      "update_hash": "update_hash",
      "update_version": 1,
      "user_auth_type": 1,
      "user_email": "user_email",
      "user_enterprise_name": "user_enterprise_name",
      "user_icon": "user_icon",
      "user_name": "user_name",
      "user_phone": "user_phone",
      "user_real_name": "user_real_name",
      "version": 1
    }
  ]
}
```

### 获取统计项信息

[返回概览](#Proof)

POST /v1/proof/QueryStatsInfo  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method"
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | string array |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    "result"
  ]
}
```

### 获得指定hash的存证信息

[返回概览](#Proof)

POST /v1/proof/Show  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Proof](#swaggerProof) array |  |
| &emsp; basehash | 增量存证依赖的主hash | string |  |
| &emsp; evidenceName | 存证名称 | string |  |
| &emsp; prehash | 增量存证前一个hash | string |  |
| &emsp; proof_block_hash | 区块hash | string |  |
| &emsp; proof_block_time | 上链时间 | integer |  |
| &emsp; proof_data | 存证数据 | string |  |
| &emsp; proof_deleted | 删除存证交易hash | string |  |
| &emsp; proof_deleted_flag | 删除标志 | boolean |  |
| &emsp; proof_deleted_note | 删除备注 | string |  |
| &emsp; proof_height | 存证高度 | integer |  |
| &emsp; proof_height_index | 存证高度索引 | integer |  |
| &emsp; proof_id | 存证id | string |  |
| &emsp; proof_note | 存证备注 | string |  |
| &emsp; proof_organization | 组织 | string |  |
| &emsp; proof_original | 来源 | string |  |
| &emsp; proof_sender | 存证发起者 | string |  |
| &emsp; proof_tx_hash | 交易哈希 | string |  |
| &emsp; source_hash | 依赖交易哈希 | object |  |
| &emsp; update_hash | 更新依赖主哈希 | string |  |
| &emsp; update_version | 更新版本 | integer |  |
| &emsp; user_auth_type | 用户认证类型 | integer |  |
| &emsp; user_email | 用户邮箱 | string |  |
| &emsp; user_enterprise_name | 用户企业名称 | string |  |
| &emsp; user_icon | 用户头像链接地址 | string |  |
| &emsp; user_name | 用户名 | string |  |
| &emsp; user_phone | 用户手机号 | string |  |
| &emsp; user_real_name | 用户真是名称 | string |  |
| &emsp; version | 存证版本 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "basehash": "basehash",
      "evidenceName": "evidenceName",
      "prehash": "prehash",
      "proof_block_hash": "proof_block_hash",
      "proof_block_time": 1,
      "proof_data": "proof_data",
      "proof_deleted": "proof_deleted",
      "proof_deleted_flag": false,
      "proof_deleted_note": "proof_deleted_note",
      "proof_height": 1,
      "proof_height_index": 1,
      "proof_id": "proof_id",
      "proof_note": "proof_note",
      "proof_organization": "proof_organization",
      "proof_original": "proof_original",
      "proof_sender": "proof_sender",
      "proof_tx_hash": "proof_tx_hash",
      "source_hash": null,
      "update_hash": "update_hash",
      "update_version": 1,
      "user_auth_type": 1,
      "user_email": "user_email",
      "user_enterprise_name": "user_enterprise_name",
      "user_icon": "user_icon",
      "user_name": "user_name",
      "user_phone": "user_phone",
      "user_real_name": "user_real_name",
      "version": 1
    }
  ]
}
```

### 获取满足条件的数据的指定字段的总值

[返回概览](#Proof)

POST /v1/proof/TotalStats  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [rpcutils.TotalStats](#rpcutilsTotalStats) array | 非必填 |  |
| body | &emsp; match | 匹配条件 | [rpcutils.QMatch](#rpcutilsQMatch) array | 非必填 |  |
| body | &emsp; sumAgg | 聚合字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "sumAgg": {
        "key": "key"
      }
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | number |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": 1
}
```

### 获取志愿者的分布图按照省/单位

[返回概览](#Proof)

POST /v1/proof/VolunteerStats  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [rpcutils.VolunteerStats](#rpcutilsVolunteerStats) array | 非必填 |  |
| body | &emsp; match | 匹配条件 | [rpcutils.QMatch](#rpcutilsQMatch) array | 非必填 |  |
| body | &emsp; subSumAgg | 子统计字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) | 非必填 |  |
| body | &emsp; subTermsAgg | 子组合字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) | 非必填 |  |
| body | &emsp; termsAgg | 聚合字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "subSumAgg": {
        "key": "key"
      },
      "subTermsAgg": {
        "key": "key"
      },
      "termsAgg": {
        "key": "key"
      }
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.VolunteerStats](#swaggerVolunteerStats) |  |
| &emsp; count | 总数 | integer |  |
| &emsp; termsAgges | 聚合 | [swagger.TermsAgg](#swaggerTermsAgg) array |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": {
    "count": 1,
    "termsAgges": [
      {
        "count": 1,
        "subTermsAgges": [
          {
            "count": 1,
            "subTermsAggKey": "subTermsAggKey"
          }
        ],
        "termsAggKey": "termsAggKey"
      }
    ]
  }
}
```

### proofmember

### 获得指定范围的用户的数量

[返回概览](#proofmember)

POST /v1/proofmember/Count  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": 1
}
```

### 获得指定地址的用户

[返回概览](#proofmember)

POST /v1/proofmember/Gets  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [rpcutils.Addresses](#rpcutilsAddresses) array | 非必填 |  |
| body | &emsp; address | 地址列表 | string array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "address": [
        "address"
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Member](#swaggerMember) array |  |
| &emsp; address | 地址 | string |  |
| &emsp; auth_type | 认证类型 | integer |  |
| &emsp; block_hash | 区块哈希 | string |  |
| &emsp; email | 邮箱 | string |  |
| &emsp; enterprise_name | 企业名称 | string |  |
| &emsp; height | 高度 | integer |  |
| &emsp; height_index | 高度索引 | integer |  |
| &emsp; index | 交易索引号 | integer |  |
| &emsp; note | 备注 | string |  |
| &emsp; organization | 组织 | string |  |
| &emsp; phone | 手机号 | string |  |
| &emsp; real_name | 真实姓名 | string |  |
| &emsp; role | 角色 | string |  |
| &emsp; send | 交易发起人 | string |  |
| &emsp; ts | 上链时间 | integer |  |
| &emsp; tx_hash | 交易hash | string |  |
| &emsp; user_icon | 头像地址链接 | string |  |
| &emsp; user_name | 用户名 | string |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "address": "address",
      "auth_type": 1,
      "block_hash": "block_hash",
      "email": "email",
      "enterprise_name": "enterprise_name",
      "height": 1,
      "height_index": 1,
      "index": 1,
      "note": "note",
      "organization": "organization",
      "phone": "phone",
      "real_name": "real_name",
      "role": "role",
      "send": "send",
      "ts": 1,
      "tx_hash": "tx_hash",
      "user_icon": "user_icon",
      "user_name": "user_name"
    }
  ]
}
```

### 分页列出指定范围的用户

[返回概览](#proofmember)

POST /v1/proofmember/List  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Member](#swaggerMember) array |  |
| &emsp; address | 地址 | string |  |
| &emsp; auth_type | 认证类型 | integer |  |
| &emsp; block_hash | 区块哈希 | string |  |
| &emsp; email | 邮箱 | string |  |
| &emsp; enterprise_name | 企业名称 | string |  |
| &emsp; height | 高度 | integer |  |
| &emsp; height_index | 高度索引 | integer |  |
| &emsp; index | 交易索引号 | integer |  |
| &emsp; note | 备注 | string |  |
| &emsp; organization | 组织 | string |  |
| &emsp; phone | 手机号 | string |  |
| &emsp; real_name | 真实姓名 | string |  |
| &emsp; role | 角色 | string |  |
| &emsp; send | 交易发起人 | string |  |
| &emsp; ts | 上链时间 | integer |  |
| &emsp; tx_hash | 交易hash | string |  |
| &emsp; user_icon | 头像地址链接 | string |  |
| &emsp; user_name | 用户名 | string |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "address": "address",
      "auth_type": 1,
      "block_hash": "block_hash",
      "email": "email",
      "enterprise_name": "enterprise_name",
      "height": 1,
      "height_index": 1,
      "index": 1,
      "note": "note",
      "organization": "organization",
      "phone": "phone",
      "real_name": "real_name",
      "role": "role",
      "send": "send",
      "ts": 1,
      "tx_hash": "tx_hash",
      "user_icon": "user_icon",
      "user_name": "user_name"
    }
  ]
}
```

### prooforganization

### 获得指定范围的组织的数量

[返回概览](#prooforganization)

POST /v1/prooforganization/Count  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | integer |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": 1
}
```

### 获得指定的组织的信息

[返回概览](#prooforganization)

POST /v1/prooforganization/Gets  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [rpcutils.Organizations](#rpcutilsOrganizations) array | 非必填 |  |
| body | &emsp; organization | 组织列表 | string array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "organization": [
        "organization"
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Organization](#swaggerOrganization) array |  |
| &emsp; count | 数量 | integer |  |
| &emsp; note | 备注 | string |  |
| &emsp; organization | 组织名 | string |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "count": 1,
      "note": "note",
      "organization": "organization"
    }
  ]
}
```

### 分页列出指定范围的组织

[返回概览](#prooforganization)

POST /v1/prooforganization/List  
Content-Type: application/json

请求参数：

| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |
|----------|----------|----------|----------|----------|----------|
| body | id | 请求标识 | integer | 非必填 |  |
| body | method | 方法 | string | 非必填 |  |
| body | params | 参数 | [swagger.Query](#swaggerQuery) array | 非必填 |  |
| body | &emsp; fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) | 非必填 |  |
| body | &emsp; filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array | 非必填 |  |
| body | &emsp; not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array | 非必填 |  |
| body | &emsp; page | 分页 | [swagger.QPage](#swaggerQPage) | 非必填 |  |
| body | &emsp; range | 范围 | [swagger.QRange](#swaggerQRange) array | 非必填 |  |
| body | &emsp; size | 大小 | [swagger.QSize](#swaggerQSize) | 非必填 |  |
| body | &emsp; sort | 排序 | [swagger.QSort](#swaggerQSort) array | 非必填 |  |

请求示例：

```json
{
  "id": 1,
  "method": "method",
  "params": [
    {
      "fetch": {
        "fetch_source": false,
        "keys": [
          "keys"
        ]
      },
      "filter": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match": [
        {
          "key": "key",
          "value": null
        }
      ],
      "match_one": [
        {
          "key": "key",
          "value": null
        }
      ],
      "multi_match": [
        {
          "keys": [
            "keys"
          ],
          "value": null
        }
      ],
      "not": [
        {
          "key": "key",
          "value": null
        }
      ],
      "page": {
        "number": 1,
        "size": 1
      },
      "range": [
        {
          "end": null,
          "gt": null,
          "key": "key",
          "lt": null,
          "start": null
        }
      ],
      "size": {
        "size": 1
      },
      "sort": [
        {
          "ascending": false,
          "key": "key"
        }
      ]
    }
  ]
}
```

响应参数：

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Organization](#swaggerOrganization) array |  |
| &emsp; count | 数量 | integer |  |
| &emsp; note | 备注 | string |  |
| &emsp; organization | 组织名 | string |  |

响应示例：

```json
{
  "error": null,
  "id": 1,
  "result": [
    {
      "count": 1,
      "note": "note",
      "organization": "organization"
    }
  ]
}
```

## 类型定义

### rpcutils.Addresses

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| address | 地址列表 | string array |  |

### rpcutils.CountByTime

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| match | 匹配条件 | [rpcutils.QMatch](#rpcutilsQMatch) array |  |
| ranges | 范围 | [rpcutils.QRanges](#rpcutilsQRanges) |  |

### rpcutils.DonationStats

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| match | 匹配条件 | [rpcutils.QMatch](#rpcutilsQMatch) array |  |
| subSumAgg | 子聚合字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) |  |
| termsAgg | 聚合字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) |  |

### rpcutils.Hashes

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| hash | 哈希列表 | string array |  |

### rpcutils.Organizations

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| organization | 组织列表 | string array |  |

### rpcutils.QMatch

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| key | 字段名 | string |  |
| value | 值 | object |  |

### rpcutils.QMatchKey

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| key | 字段名 | string |  |

### rpcutils.QRanges

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| key | 字段名 | string |  |
| ranges | 范围 | [rpcutils.Range](#rpcutilsRange) array |  |

### rpcutils.QSort

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| ascending | 是否升序 | boolean |  |
| key | 字段名 | string |  |

### rpcutils.Range

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| end | 结束位置 | object | 小于等于 |
| start | 开始位置 | object | 大于等于 |

### rpcutils.RepLastSeq

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| lastConvertSeq | 最新解析区块高度 | integer |  |
| lastSyncSeq | 最新同步区块高度 | integer |  |

### rpcutils.ServerResponse

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | object |  |

### rpcutils.SpecifiedFields

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count | 总量 | integer |  |
| fields | 字段列表 | string array |  |
| match | 匹配条件 | [rpcutils.QMatch](#rpcutilsQMatch) array |  |
| sort | 排序 | [rpcutils.QSort](#rpcutilsQSort) array |  |

### rpcutils.TotalStats

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| match | 匹配条件 | [rpcutils.QMatch](#rpcutilsQMatch) array |  |
| sumAgg | 聚合字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) |  |

### rpcutils.VolunteerStats

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| match | 匹配条件 | [rpcutils.QMatch](#rpcutilsQMatch) array |  |
| subSumAgg | 子统计字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) |  |
| subTermsAgg | 子组合字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) |  |
| termsAgg | 聚合字段 | [rpcutils.QMatchKey](#rpcutilsQMatchKey) |  |

### swagger.Attributes

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| ml.enabled |  | string |  |
| ml.machine_memory |  | string |  |
| ml.max_open_jobs |  | string |  |

### swagger.BufferPools

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| direct |  | [swagger.Direct](#swaggerDirect) |  |
| mapped |  | [swagger.Mapped](#swaggerMapped) |  |

### swagger.CPU

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| load_average |  | [swagger.LoadAverage](#swaggerLoadAverage) |  |
| percent |  | integer |  |

### swagger.ChainStatus

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| coin | 主代币信息 | string |  |
| push_seq | 推送高度 | integer |  |
| status | 状态 | string |  |
| version | 版本 | [swagger.ChainVersionInfo](#swaggerChainVersionInfo) |  |

### swagger.ChainVersionInfo

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| app | 应用 app 的版本 | string |  |
| chain33 | 版本信息，版本号-GitCommit | string | 前八个字符 |
| localDb | localdb 版本号 | string |  |
| title | 区块链名，该节点 chain33.toml 中配置的 title 值 | string |  |

### swagger.Classes

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| current_loaded_count |  | integer |  |
| total_loaded_count |  | integer |  |
| total_unloaded_count |  | integer |  |

### swagger.ClientRequest

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| id | 请求标识 | integer |  |
| method | 方法 | string |  |
| params | 参数 | object |  |

### swagger.ClientRequestNil

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| id | 请求标识 | integer |  |
| method | 方法 | string |  |

### swagger.Collectors

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| old |  | [swagger.GcOld](#swaggerGcOld) |  |
| young |  | [swagger.GcYoung](#swaggerGcYoung) |  |

### swagger.Completion

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| size_in_bytes |  | integer |  |

### swagger.Direct

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count |  | integer |  |
| total_capacity_in_bytes |  | integer |  |
| used_in_bytes |  | integer |  |

### swagger.Docs

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count |  | integer |  |
| deleted |  | integer |  |

### swagger.DonationStatItem

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count | 数量 | integer |  |
| name | 名称 | string |  |
| total | 总合 | integer |  |

### swagger.DonationStats

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| itemes | 列表 | [swagger.DonationStatItem](#swaggerDonationStatItem) array |  |

### swagger.EVMToken

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| amount | 金额 | integer |  |
| call_func_name | 调用方法名称 | string |  |
| contract_addr | 合约地址 | string |  |
| contract_used_gas | 消耗gas | integer |  |
| evm_block_hash | 区块hash | string |  |
| evm_block_time | 上链时间 | integer |  |
| evm_events | evm事件 | string |  |
| evm_height | 区块高度 | integer |  |
| evm_height_index | 高度索引 | integer |  |
| evm_note | 备注信息 | string |  |
| evm_param | evm调用参数 | string |  |
| evm_tx_hash | 交易hash | string |  |
| goods_id | 物品唯一标识 | integer |  |
| goods_type | 物品类型 | integer |  |
| label_id | 物品标签id | string |  |
| name | 物品名称 | string |  |
| owner | 拥有者 | string |  |
| publish_time | 发布时间 | integer |  |
| publisher | 发布者 | string |  |
| remark | 备注 | string |  |
| source_hash | 关联交易hash | string array |  |
| trace_hash | 关联溯源hash | string array |  |

### swagger.EsStatus

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| _nodes | 节点数量信息 | [swagger.NodesCount](#swaggerNodesCount) |  |
| cluster_name | 集群名 | string |  |
| nodes | 节点信息 | map\[string\] [swagger.Node](#swaggerNode) |  |
| status | 状态 | string |  |

### swagger.Fielddata

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| evictions |  | integer |  |
| memory_size_in_bytes |  | integer |  |

### swagger.FileSizes

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|

### swagger.Flush

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| total |  | integer |  |
| total_time_in_millis |  | integer |  |

### swagger.Gc

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| collectors |  | [swagger.Collectors](#swaggerCollectors) |  |

### swagger.GcOld

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| collection_count |  | integer |  |
| collection_time_in_millis |  | integer |  |

### swagger.GcYoung

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| collection_count |  | integer |  |
| collection_time_in_millis |  | integer |  |

### swagger.Get

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| current |  | integer |  |
| exists_time_in_millis |  | integer |  |
| exists_total |  | integer |  |
| missing_time_in_millis |  | integer |  |
| missing_total |  | integer |  |
| time_in_millis |  | integer |  |
| total |  | integer |  |

### swagger.HTTP

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| current_open |  | integer |  |
| total_opened |  | integer |  |

### swagger.Health

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| status | 状态 | string |  |
| version | 版本 | string |  |

### swagger.ISearch

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| fetch_current |  | integer |  |
| fetch_time_in_millis |  | integer |  |
| fetch_total |  | integer |  |
| open_contexts |  | integer |  |
| query_current |  | integer |  |
| query_time_in_millis |  | integer |  |
| query_total |  | integer |  |
| scroll_current |  | integer |  |
| scroll_time_in_millis |  | integer |  |
| scroll_total |  | integer |  |
| suggest_current |  | integer |  |
| suggest_time_in_millis |  | integer |  |
| suggest_total |  | integer |  |

### swagger.Indexing

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| delete_current |  | integer |  |
| delete_time_in_millis |  | integer |  |
| delete_total |  | integer |  |
| index_current |  | integer |  |
| index_failed |  | integer |  |
| index_time_in_millis |  | integer |  |
| index_total |  | integer |  |
| is_throttled |  | boolean |  |
| noop_update_total |  | integer |  |
| throttle_time_in_millis |  | integer |  |

### swagger.Indices

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| completion |  | [swagger.Completion](#swaggerCompletion) |  |
| docs |  | [swagger.Docs](#swaggerDocs) |  |
| fielddata |  | [swagger.Fielddata](#swaggerFielddata) |  |
| flush |  | [swagger.Flush](#swaggerFlush) |  |
| get |  | [swagger.Get](#swaggerGet) |  |
| indexing |  | [swagger.Indexing](#swaggerIndexing) |  |
| merges |  | [swagger.Merges](#swaggerMerges) |  |
| query_cache |  | [swagger.QueryCache](#swaggerQueryCache) |  |
| recovery |  | [swagger.Recovery](#swaggerRecovery) |  |
| refresh |  | [swagger.Refresh](#swaggerRefresh) |  |
| request_cache |  | [swagger.RequestCache](#swaggerRequestCache) |  |
| search |  | [swagger.ISearch](#swaggerISearch) |  |
| segments |  | [swagger.Segments](#swaggerSegments) |  |
| store |  | [swagger.Store](#swaggerStore) |  |
| translog |  | [swagger.Translog](#swaggerTranslog) |  |
| warmer |  | [swagger.Warmer](#swaggerWarmer) |  |

### swagger.Jvm

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| buffer_pools |  | [swagger.BufferPools](#swaggerBufferPools) |  |
| classes |  | [swagger.Classes](#swaggerClasses) |  |
| gc |  | [swagger.Gc](#swaggerGc) |  |
| mem |  | [swagger.JvmMem](#swaggerJvmMem) |  |
| threads |  | [swagger.Threads](#swaggerThreads) |  |
| timestamp |  | integer |  |
| uptime_in_millis |  | integer |  |

### swagger.JvmMem

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| heap_committed_in_bytes |  | integer |  |
| heap_max_in_bytes |  | integer |  |
| heap_used_in_bytes |  | integer |  |
| heap_used_percent |  | integer |  |
| non_heap_committed_in_bytes |  | integer |  |
| non_heap_used_in_bytes |  | integer |  |
| pools |  | [swagger.Pools](#swaggerPools) |  |

### swagger.ListEVMResult

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.EVMToken](#swaggerEVMToken) array |  |

### swagger.ListProofResult

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.Proof](#swaggerProof) array |  |

### swagger.LoadAverage

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| 15m |  | number |  |
| 1m |  | number |  |
| 5m |  | number |  |

### swagger.Mapped

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count |  | integer |  |
| total_capacity_in_bytes |  | integer |  |
| used_in_bytes |  | integer |  |

### swagger.Member

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| address | 地址 | string |  |
| auth_type | 认证类型 | integer |  |
| block_hash | 区块哈希 | string |  |
| email | 邮箱 | string |  |
| enterprise_name | 企业名称 | string |  |
| height | 高度 | integer |  |
| height_index | 高度索引 | integer |  |
| index | 交易索引号 | integer |  |
| note | 备注 | string |  |
| organization | 组织 | string |  |
| phone | 手机号 | string |  |
| real_name | 真实姓名 | string |  |
| role | 角色 | string |  |
| send | 交易发起人 | string |  |
| ts | 上链时间 | integer |  |
| tx_hash | 交易hash | string |  |
| user_icon | 头像地址链接 | string |  |
| user_name | 用户名 | string |  |

### swagger.Merges

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| current |  | integer |  |
| current_docs |  | integer |  |
| current_size_in_bytes |  | integer |  |
| total |  | integer |  |
| total_auto_throttle_in_bytes |  | integer |  |
| total_docs |  | integer |  |
| total_size_in_bytes |  | integer |  |
| total_stopped_time_in_millis |  | integer |  |
| total_throttled_time_in_millis |  | integer |  |
| total_time_in_millis |  | integer |  |

### swagger.Node

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| attributes |  | [swagger.Attributes](#swaggerAttributes) |  |
| host |  | string |  |
| http |  | [swagger.HTTP](#swaggerHTTP) |  |
| indices | 索引 | [swagger.Indices](#swaggerIndices) |  |
| ip |  | string |  |
| jvm | java虚拟机 | [swagger.Jvm](#swaggerJvm) |  |
| name |  | string |  |
| os | 系统 | [swagger.Os](#swaggerOs) |  |
| roles |  | string array |  |
| timestamp |  | integer |  |
| transport_address |  | string |  |

### swagger.NodesCount

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| failed | 不正常数量 | integer |  |
| successful | 正常数量 | integer |  |
| total | 总计 | integer |  |

### swagger.Organization

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count | 数量 | integer |  |
| note | 备注 | string |  |
| organization | 组织名 | string |  |

### swagger.Os

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| cpu |  | [swagger.CPU](#swaggerCPU) |  |
| mem |  | [swagger.OsMem](#swaggerOsMem) |  |
| swap |  | [swagger.Swap](#swaggerSwap) |  |
| timestamp |  | integer |  |

### swagger.OsMem

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| free_in_bytes |  | integer |  |
| free_percent |  | integer |  |
| total_in_bytes |  | integer |  |
| used_in_bytes |  | integer |  |
| used_percent |  | integer |  |

### swagger.Pools

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| old |  | [swagger.PoolsOld](#swaggerPoolsOld) |  |
| survivor |  | [swagger.Survivor](#swaggerSurvivor) |  |
| young |  | [swagger.PoolsYoung](#swaggerPoolsYoung) |  |

### swagger.PoolsOld

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| max_in_bytes |  | integer |  |
| peak_max_in_bytes |  | integer |  |
| peak_used_in_bytes |  | integer |  |
| used_in_bytes |  | integer |  |

### swagger.PoolsYoung

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| max_in_bytes |  | integer |  |
| peak_max_in_bytes |  | integer |  |
| peak_used_in_bytes |  | integer |  |
| used_in_bytes |  | integer |  |

### swagger.Proof

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| basehash | 增量存证依赖的主hash | string |  |
| evidenceName | 存证名称 | string |  |
| prehash | 增量存证前一个hash | string |  |
| proof_block_hash | 区块hash | string |  |
| proof_block_time | 上链时间 | integer |  |
| proof_data | 存证数据 | string |  |
| proof_deleted | 删除存证交易hash | string |  |
| proof_deleted_flag | 删除标志 | boolean |  |
| proof_deleted_note | 删除备注 | string |  |
| proof_height | 存证高度 | integer |  |
| proof_height_index | 存证高度索引 | integer |  |
| proof_id | 存证id | string |  |
| proof_note | 存证备注 | string |  |
| proof_organization | 组织 | string |  |
| proof_original | 来源 | string |  |
| proof_sender | 存证发起者 | string |  |
| proof_tx_hash | 交易哈希 | string |  |
| source_hash | 依赖交易哈希 | object |  |
| update_hash | 更新依赖主哈希 | string |  |
| update_version | 更新版本 | integer |  |
| user_auth_type | 用户认证类型 | integer |  |
| user_email | 用户邮箱 | string |  |
| user_enterprise_name | 用户企业名称 | string |  |
| user_icon | 用户头像链接地址 | string |  |
| user_name | 用户名 | string |  |
| user_phone | 用户手机号 | string |  |
| user_real_name | 用户真是名称 | string |  |
| version | 存证版本 | integer |  |

### swagger.QFetch

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| fetch_source | 是否获取 | boolean |  |
| keys | 字段名列表 | string array |  |

### swagger.QMatch

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| key | 字段名 | string |  |
| value | 值 | object |  |

### swagger.QMultiMatch

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| keys | 字段名列表 | string array |  |
| value | 值 | object |  |

### swagger.QPage

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| number | 当前页数 | integer |  |
| size | 大小 | integer |  |

### swagger.QRange

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| end | 小于等于 | object |  |
| gt | 大于 | object |  |
| key | 字段名 | string |  |
| lt | 小于 | object |  |
| start | 大于等于 | object |  |

### swagger.QSize

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| size | 大小 | integer |  |

### swagger.QSort

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| ascending | 是否递增 | boolean |  |
| key | 字段名 | string |  |

### swagger.Query

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| fetch | 获取字段 | [swagger.QFetch](#swaggerQFetch) |  |
| filter | 过滤 | [swagger.QMatch](#swaggerQMatch) array |  |
| match | 且匹配 | [swagger.QMatch](#swaggerQMatch) array |  |
| match_one | 或匹配 | [swagger.QMatch](#swaggerQMatch) array |  |
| multi_match | 多字段匹配 | [swagger.QMultiMatch](#swaggerQMultiMatch) array |  |
| not | 非匹配 | [swagger.QMatch](#swaggerQMatch) array |  |
| page | 分页 | [swagger.QPage](#swaggerQPage) |  |
| range | 范围 | [swagger.QRange](#swaggerQRange) array |  |
| size | 大小 | [swagger.QSize](#swaggerQSize) |  |
| sort | 排序 | [swagger.QSort](#swaggerQSort) array |  |

### swagger.QueryCache

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| cache_count |  | integer |  |
| cache_size |  | integer |  |
| evictions |  | integer |  |
| hit_count |  | integer |  |
| memory_size_in_bytes |  | integer |  |
| miss_count |  | integer |  |
| total_count |  | integer |  |

### swagger.Recovery

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| current_as_source |  | integer |  |
| current_as_target |  | integer |  |
| throttle_time_in_millis |  | integer |  |

### swagger.Refresh

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| listeners |  | integer |  |
| total |  | integer |  |
| total_time_in_millis |  | integer |  |

### swagger.RequestCache

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| evictions |  | integer |  |
| hit_count |  | integer |  |
| memory_size_in_bytes |  | integer |  |
| miss_count |  | integer |  |

### swagger.Segments

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count |  | integer |  |
| doc_values_memory_in_bytes |  | integer |  |
| file_sizes |  | [swagger.FileSizes](#swaggerFileSizes) |  |
| fixed_bit_set_memory_in_bytes |  | integer |  |
| index_writer_memory_in_bytes |  | integer |  |
| max_unsafe_auto_id_timestamp |  | integer |  |
| memory_in_bytes |  | integer |  |
| norms_memory_in_bytes |  | integer |  |
| points_memory_in_bytes |  | integer |  |
| stored_fields_memory_in_bytes |  | integer |  |
| term_vectors_memory_in_bytes |  | integer |  |
| terms_memory_in_bytes |  | integer |  |
| version_map_memory_in_bytes |  | integer |  |

### swagger.ServerResponse

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | object |  |

### swagger.ServerStatus

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| coin | 币名称 | string |  |
| conv_seq | 转换序列高度 | integer |  |
| sync_seq | 同步序列高度 | integer |  |
| title | 标题 | string |  |
| version | 版本 | string |  |

### swagger.Status

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| chain | 链状态 | [swagger.ChainStatus](#swaggerChainStatus) |  |
| es | ElasticSearch状态 | [swagger.EsStatus](#swaggerEsStatus) |  |
| server | 服务状态 | [swagger.ServerStatus](#swaggerServerStatus) |  |

### swagger.Store

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| size_in_bytes |  | integer |  |

### swagger.SubTermsAgges

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count | 聚合数量 | integer |  |
| subTermsAggKey | 子聚合键值 | string |  |

### swagger.Survivor

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| max_in_bytes |  | integer |  |
| peak_max_in_bytes |  | integer |  |
| peak_used_in_bytes |  | integer |  |
| used_in_bytes |  | integer |  |

### swagger.Swap

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| free_in_bytes |  | integer |  |
| total_in_bytes |  | integer |  |
| used_in_bytes |  | integer |  |

### swagger.Template

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| template_block_hash | 区块哈希 | string |  |
| template_block_time | 上链时间 | integer |  |
| template_data | 模板数据 | string |  |
| template_deleted | 删除交易哈希 | string |  |
| template_deleted_flag | 删除标志 | boolean |  |
| template_deleted_note | 删除备注 | string |  |
| template_height | 高度 | integer |  |
| template_height_index | 高度索引 | integer |  |
| template_id | 模板id | string |  |
| template_name | 模板名称 | string |  |
| template_organization | 组织 | string |  |
| template_sender | 交易发送人 | string |  |
| template_tx_hash | 交易哈希 | string |  |

### swagger.TermsAgg

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count | 数量 | integer |  |
| subTermsAgges | 子聚合 | [swagger.SubTermsAgges](#swaggerSubTermsAgges) array |  |
| termsAggKey | 聚合键值 | string |  |

### swagger.Threads

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count |  | integer |  |
| peak_count |  | integer |  |

### swagger.Translog

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| operations |  | integer |  |
| size_in_bytes |  | integer |  |
| uncommitted_operations |  | integer |  |
| uncommitted_size_in_bytes |  | integer |  |

### swagger.VolunteerStats

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| count | 总数 | integer |  |
| termsAgges | 聚合 | [swagger.TermsAgg](#swaggerTermsAgg) array |  |

### swagger.VolunteerStatsResult

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| error | 错误描述 | object |  |
| id | 请求标识 | integer |  |
| result | 返回结果 | [swagger.VolunteerStats](#swaggerVolunteerStats) |  |

### swagger.Warmer

| **参数** | **描述** | **类型** | **说明** |
|----------|----------|----------|----------|
| current |  | integer |  |
| total |  | integer |  |
| total_time_in_millis |  | integer |  |

