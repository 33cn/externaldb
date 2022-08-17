
# externaldb rpc 

 rpc 接口基本可以分为三类

 1. 指定ID， 可以直接取数据
 1. 通过条件搜索
 1. 需要在接口中带上部分业务逻辑.

一般情况下不产生第3类接口。为避免多次请求，可以将多个请求合并成一个请求，提供Gets/Searches 类似这样的接口。 

 1. Gets 同时指定多个ID
 1. Searches 同时指定多个查询条件

Search 参数尽量和 ES的参数一致， 减少代码量， 也方便开发rpc。

## 1. URL

 1. http{s}://Host:Port/Title

```
http://127.0.0.1:9991/bityuan
http://47.111.132.89:19992/user.p.sakurachain.
```

## 2. 接口的查询参数

由于查询的需求多种多样， 查询请求做的已经通用。 主要提供4个项

 1. page： 分页。 不提供的话， 会提供默认个
 1. sort： 是数组，可以提供多个排序条件。相对于 sql 的 order by。
    1.  会提供 key 为 height_index 通用的上链顺序
    1.  其他的业务相关的， 如价格(boardlot_price)等
 1. match：数组。 查询的筛选条件. 交易中可以提供
    1.  send： 提供地址， 发起交易的人 
    1.  asset_symbol: 交易的资产 (如果存在不同合约里有同名的资产， 可以同时提供 asset_exec 指定合约)
 1. match_one：数组。 查询的筛选条件. 满足一个条件就可以了
    1.  send： 提供地址， 发起交易的人 
    1.  asset_symbol: 交易的资产 (如果存在不同合约里有同名的资产， 可以同时提供 asset_exec 指定合约)
 1. range: 提供区间查询
    1. start, end 包含临界值
    1. gt, lt 不包含临界值(great than, less than)
 1. multi_match: 数据。 查询的筛选条件. 在一组域中其中一个匹配条件就可以。 
    1. key 是一个数组
    1. value 对应的匹配的值。 只需要key中有一个匹配
    1. 常见使用：to/from 地址中是 A地址，来筛选 A地址相关的交易。 match_one 在 value 一样的情况下的，简化使用

查询嵌套
 1. 在match, match_one 中可以有 query 子查询， 来构造更复杂的查询。

```
# sub query 例子
# (from = A Or to = A) And (is_para = true Or success = true)
{ 
         "match" : [
            {
               "query" : {
                  "match_one" : [
                     {"key" : "from", "value" : "1HyW3ZJ8DJVwmMuEjaiXg6D9Zz8ooZLRbJ"},
                     {"key" : "to", "value" : "1HyW3ZJ8DJVwmMuEjaiXg6D9Zz8ooZLRbJ"}
                  ]
               }
            },
            {
               "query" : {
                  "match_one" : [
                     {"key" : "is_para", "value" : true},
                     {"key" : "success", "value" : true}
                  ]
               }
            },
            {
               "value" : "multisig",
               "key" : "execer"
            }
         ],
         "sort" : [
            {
               "key" : "height",
               "ascending" : true
            }
         ],
         "page" : {
            "number" : 1,
            "size" : 100
         }
      }
```

```
# match_one 例子
       {
         "match_one" : [
            {
               "value" : "14uBEP6LSHKdFvy97pTYRPVPAqij6bteee",
               "key" : "to"
            },
            {
               "value" : "coins",
               "key" : "execer"
            }
         ],
         "sort" : [
            {
               "key" : "height",
               "ascending" : true
            }
         ],
         "page" : {
            "number" : 1,
            "size" : 100
         }
      }
```

## 3. 获得区块相关信息

 提供6个接口，返回类型如下。 6个接口分为两组，一个为原始值， 一组为累计值。

 1. Gets, Search, Searches: 原始值
 1. StatGets，StatSearch， StatSearches： 累计值

```
      {
         "height" : 1,        			# 当前高度   
         "time" : 1565577904,			# 区块时间
         "coins" : 317431002999900000,		# 有多少主币
         "mine" : 3000000000,			# 挖矿所得： 当前区块/从0开始到当前区块累计
         "fee" : 100000,			# 手续费：当前区块/从0开始到当前区块累计
         "tx_count" : 10			# 交易数：当前区块/从0开始到当前区块累计
      }
```

### 3.1 BlockStat.StatGets

```
curl -d@stat_1.json   http://127.0.0.1:9991/bityuan | json_pp
```


request :stat_1.json

```
{
	"id" : 1 ,
	"method" : "BlockStat.StatGets", 
	"params":[{ 
		"height" : [1, 10, 100]
	}]
}
```


response


```
{
   "result" : [
      {
         "height" : 1,
         "fee" : 100000,
         "mine" : 3000000000,
         "coins" : 317431002999900000,
         "tx_count" : 10,
         "time" : 1565577904
      },
      {
         "height" : 10,
         "fee" : 1000000,
         "mine" : 30000000000,
         "coins" : 317431029999000000,
         "tx_count" : 19,
         "time" : 1565577905
      },
      {
         "mine" : 300000000000,
         "height" : 100,
         "fee" : 11200000,
         "tx_count" : 121,
         "coins" : 317431299988800000,
         "time" : 1565577918
      }
   ],
   "error" : null,
   "id" : 1
}

```

### 3.2 BlockStat.Gets

```
curl -d@org_1.json   http://127.0.0.1:9991/bityuan | json_pp
```

request

```
#$ cat org_1.json 
{
	"id" : 1 ,
	"method" : "BlockStat.Gets", 
	"params":[{ 
		"height" : [1, 10, 100]
	}]
}


```

response

```
{
   "id" : 1,
   "error" : null,
   "result" : [
      {
         "time" : 1565577904,
         "mine" : 3000000000,
         "fee" : 100000,
         "coins" : 317431002999900000,
         "height" : 1,
         "tx_count" : 1
      },
      {
         "coins" : 317431029999000000,
         "fee" : 100000,
         "mine" : 3000000000,
         "time" : 1565577905,
         "tx_count" : 1,
         "height" : 10
      },
      {
         "mine" : 3000000000,
         "time" : 1565577918,
         "coins" : 317431299988800000,
         "fee" : 100000,
         "height" : 100,
         "tx_count" : 1
      }
   ]
}

```

### search 参数通用介绍

查询请求做的已经通用。 主要提供4个项

 1. page： 分页。 不提供的话， 会提供默认个
 1. sort： 是数组，可以提供多个排序条件。相对于 sql 的 order by。
    1.  会提供 key 为 height 通用的上链顺序
    1.  其他的业务相关的
 1. match：数组。 查询的筛选条件. 交易中可以提供
    1.  send： 提供地址， 发起交易的人
    1.  asset_symbol: 交易的资产 (如果存在不同合约里有同名的资产， 可以同时提供 asset_exec 指定合约)
 1. range: 提供区间查询
    1. start, end 包含临界值
    1. gt, lt 不包含临界值(great than, less than)

### 3.3 BlockStat.StatSearch

```
 curl -d@stat_2.json   http://127.0.0.1:9991/bityuan | json_pp
```

request

```

{
	"id" : 1 ,
	"method" : "BlockStat.StatSearch", 
        "params":[{ 
   		"page" : {
   		   "number" : 1,
   		   "size" : 2 
   		},
   		"sort" : [
   		   {
   		      "ascending" : true,
   		      "key" : "height"
   		   }
   		],
   		"range" : [
   		   {
   		      "start" : 1565577919,
   		      "key" : "time"
   		   }
   		]
                
        }]
}


```

response

```
{
   "id" : 1,
   "error" : null,
   "result" : [
      {
         "fee" : 11300000,
         "height" : 101,
         "mine" : 303000000000,
         "coins" : 317431302988700000,
         "time" : 1565577919,
         "tx_count" : 122
      },
      {
         "fee" : 11400000,
         "height" : 102,
         "mine" : 306000000000,
         "coins" : 317431305988600000,
         "tx_count" : 123,
         "time" : 1565577919
      }
   ]
}

```

### 3.4 BlockStat.Search

```
 curl -d@stat_2.json   http://127.0.0.1:9991/bityuan | json_pp
```

request

```
{
	"id" : 1 ,
	"method" : "BlockStat.Search", 
        "params":[{ 
   		"page" : {
   		   "number" : 1,
   		   "size" : 2 
   		},
   		"sort" : [
   		   {
   		      "ascending" : true,
   		      "key" : "height"
   		   }
   		],
   		"range" : [
   		   {
   		      "start" : 1565577919,
   		      "key" : "time"
   		   }
   		]
                
        }]
}


```

response

```

{
   "result" : [
      {
         "fee" : 100000,
         "tx_count" : 1,
         "mine" : 3000000000,
         "height" : 102,
         "coins" : 317431305988600000,
         "time" : 1565577919
      },
      {
         "mine" : 3000000000,
         "tx_count" : 1,
         "fee" : 100000,
         "coins" : 317431308988500000,
         "height" : 103,
         "time" : 1565577919
      }
   ],
   "id" : 1,
   "error" : null
}


```

### 3.5 BlockStat.StatSearches 

```
curl -d@stat_3.json   http://127.0.0.1:9991/bityuan | json_pp
```

request 

```
{
   "method" : "BlockStat.StatSearches",
   "params" : [
      [
         {
            "page" : {
               "size" : 2,
               "number" : 1
            },
            "range" : [
               {
                  "start" : 1565579919,
                  "key" : "time"
               }
            ],
            "sort" : [
               {
                  "key" : "height",
                  "ascending" : true
               }
            ]
         },
         {
            "sort" : [
               {
                  "key" : "height",
                  "ascending" : true
               }
            ],
            "range" : [
               {
                  "start" : 1565576919,
                  "key" : "time"
               }
            ],
            "page" : {
               "size" : 2,
               "number" : 1
            }
         }
      ]
   ],
   "id" : 1
}


```

response

```
{
   "result" : [
      [
         {
            "fee" : 103600000,
            "coins" : 317433606896400000,
            "mine" : 2607000000000,
            "tx_count" : 947,
            "height" : 869,
            "time" : 1565579919
         },
         {
            "height" : 870,
            "time" : 1565579920,
            "fee" : 103700000,
            "coins" : 317433609896300000,
            "tx_count" : 948,
            "mine" : 2610000000000
         }
      ],
      [
         {
            "height" : 1,
            "time" : 1565577904,
            "fee" : 100000,
            "coins" : 317431002999900000,
            "tx_count" : 10,
            "mine" : 3000000000
         },
         {
            "fee" : 200000,
            "coins" : 317431005999800000,
            "tx_count" : 11,
            "mine" : 6000000000,
            "height" : 2,
            "time" : 1565577904
         }
      ]
   ],
   "id" : 1,
   "error" : null
}

```


### 3.6 BlockStat.Searches

```
curl -d@org_3.json   http://127.0.0.1:9991/bityuan | json_pp
```

request

```
{
   "id" : 1,
   "method" : "BlockStat.Searches",
   "params" : [
      [
         {
            "page" : {
               "number" : 1,
               "size" : 2
            },
            "sort" : [
               {
                  "ascending" : true,
                  "key" : "height"
               }
            ],
            "range" : [
               {
                  "start" : 1565579919,
                  "key" : "time"
               }
            ]
         },
         {
            "page" : {
               "number" : 1,
               "size" : 2
            },
            "sort" : [
               {
                  "key" : "height",
                  "ascending" : true
               }
            ],
            "range" : [
               {
                  "start" : 1565576919,
                  "key" : "time"
               }
            ]
         }
      ]
   ]
}

```

response

```
{
   "id" : 1,
   "result" : [
      [
         {
            "height" : 870,
            "coins" : 317433609896300000,
            "time" : 1565579920,
            "mine" : 3000000000,
            "tx_count" : 1,
            "fee" : 100000
         },
         {
            "tx_count" : 1,
            "fee" : 100000,
            "mine" : 3000000000,
            "coins" : 317433612896200000,
            "time" : 1565579920,
            "height" : 871
         }
      ],
      [
         {
            "fee" : 100000,
            "tx_count" : 1,
            "time" : 1565577904,
            "coins" : 317431005999800000,
            "mine" : 3000000000,
            "height" : 2
         },
         {
            "tx_count" : 1,
            "fee" : 100000,
            "mine" : 3000000000,
            "coins" : 317431008999700000,
            "time" : 1565577904,
            "height" : 3
         }
      ]
   ],
   "error" : null
}

```

## 4 Account

### 4.1 Account.ListAsset 

根据条件搜索帐号信息


 * request: 通用查询参数

```

演示： 用地址和合约查询帐号信息

{
        "id" : 1 ,
        "method" : "Account.ListAsset", 
        "params":[{
   		"match" : [
   		   {
   		      "value" : "145HxiUdRSzt49kKA2WoDyjBcYAJB67M46",
   		      "key" : "address"
   		   },
   		   {
   		      "value" : "1BXvgjmBw1aBgmGn1hjfGyRkmN3krWpFP4",
   		      "key" : "exec"
   		   }
   		]
	}
		
        ]
}

```

 * response 

```
{
   "result" : [
      {
         "asset_symbol" : "bty",
         "frozen" : 0,
         "total" : 0,
         "balance" : 0,
         "height_index" : 490600002,
         "exec" : "1BXvgjmBw1aBgmGn1hjfGyRkmN3krWpFP4",
         "address" : "145HxiUdRSzt49kKA2WoDyjBcYAJB67M46",
         "type" : "contractInternal",
         "asset_exec" : "coins"
      },
      {
         "type" : "contractInternal",
         "asset_exec" : "token",
         "address" : "145HxiUdRSzt49kKA2WoDyjBcYAJB67M46",
         "height_index" : 206500002,
         "exec" : "1BXvgjmBw1aBgmGn1hjfGyRkmN3krWpFP4",
         "total" : 675000000,
         "balance" : 675000000,
         "asset_symbol" : "TEST",
         "frozen" : 0
      }
   ],
   "id" : 1,
   "error" : null
}


```

### 4.2 Account.Count 

根据条件统计帐号信息


 * request: 通用查询参数

```

演示： coins 有多少人持有

{
   "method" : "Account.Count",
   "params" : [
      {
         "match" : [
            {
               "value" : "coins",
               "key" : "asset_exec"
            }
         ]
      }
   ],
   "id" : 1
}


```

 * response 

```
{"id":1,"result":16,"error":null}
```


### 4.3 Account.Search 

根据条件搜索帐号信息


 * request: 通用查询参数

```

演示： 用地址和合约查询帐号信息

{
   "method" : "Account.Search",
   "params" : [
      {
         "match" : [

            {
               "value" : "coins",
               "key" : "asset_exec"
            }
         ],
	 "page" : {
		 "size" : 2,
		 "number" :3
	 }
      }
   ],
   "id" : 1
}


```

 * response 

```

{
   "result" : [
      {
         "balance" : 0,
         "frozen" : 10000000000000000,
         "total" : 10000000000000000,
         "asset_exec" : "coins",
         "address" : "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF",
         "type" : "contractInternal",
         "exec" : "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp",
         "asset_symbol" : "bty",
         "height_index" : 5
      },
      {
         "total" : 10000000000,
         "asset_exec" : "coins",
         "frozen" : 0,
         "balance" : 10000000000,
         "height_index" : 61400001,
         "asset_symbol" : "bty",
         "address" : "14uBEP6LSHKdFvy97pTYRPVPAqij6bteee",
         "type" : "contract",
         "exec" : ""
      }
   ],
   "id" : 1,
   "error" : null
}

```

### 4.4 Account.Searches

同时进行多个搜索


 * request: 通用查询参数

```

演示：  
{
   "id" : 1,
   "params" : [
      [
         {
            "page" : {
               "size" : 2,
               "number" : 3
            },
            "match" : [
               {
                  "key" : "asset_exec",
                  "value" : "coins"
               }
            ]
         },
         {
            "match" : [
               {
                  "value" : "coins",
                  "key" : "asset_exec"
               }
            ],
            "page" : {
               "number" : 2,
               "size" : 3
            }
         }
      ]
   ],
   "method" : "Account.Searches"
}


```

 * response 

```
{
   "result" : [
      [
         {
            "asset_symbol" : "bty",
            "frozen" : 10000000000000000,
            "height_index" : 5,
            "type" : "contractInternal",
            "total" : 10000000000000000,
            "address" : "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF",
            "exec" : "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp",
            "balance" : 0,
            "asset_exec" : "coins"
         },
         {
            "asset_exec" : "coins",
            "balance" : 10000000000,
            "height_index" : 61400001,
            "type" : "contract",
            "asset_symbol" : "bty",
            "frozen" : 0,
            "exec" : "",
            "total" : 10000000000,
            "address" : "14uBEP6LSHKdFvy97pTYRPVPAqij6bteee"
         }
      ],
      [
         {
            "balance" : 20013005000000000,
            "asset_exec" : "coins",
            "height_index" : 1000000000,
            "type" : "contract",
            "asset_symbol" : "bty",
            "frozen" : 0,
            "exec" : "",
            "total" : 20013005000000000,
            "address" : "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
         },
         {
            "balance" : 0,
            "asset_exec" : "coins",
            "total" : 10000000000000000,
            "address" : "1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF",
            "exec" : "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp",
            "asset_symbol" : "bty",
            "frozen" : 10000000000000000,
            "height_index" : 5,
            "type" : "contractInternal"
         },
         {
            "frozen" : 0,
            "asset_symbol" : "bty",
            "type" : "contract",
            "height_index" : 61400001,
            "address" : "14uBEP6LSHKdFvy97pTYRPVPAqij6bteee",
            "total" : 10000000000,
            "exec" : "",
            "asset_exec" : "coins",
            "balance" : 10000000000
         }
      ]
   ],
   "error" : null,
   "id" : 1
}


```

### 4.5 Account.TopAsset

 1. 功能：返回在coins和ticket合约中主币最多的地址列表(定制功能给浏览器用)
 1. 参数：分页 page.number, page.size
 1. 非实时数据，使用缓存，第一次调用会慢， 一小时更新一次

 输入
```
 # topAsset.p1.json
{
   "method" : "Account.TopAsset",
   "params" : [
      {
         "page" : 
            {
               "number" : 1,
               "size" : 5
            }
      }
   ],
   "id" : 1
}
 # topAsset.p2.json
{
   "method" : "Account.TopAsset",
   "params" : [
      {
         "page" :
            {
               "number" : 2,
               "size" : 5
            }
      }
   ],
   "id" : 1
}

$ curl http://47.100.234.232:20012/bityuan -d@topAsset.p1.json | json_pp
{
   "error" : null,
   "result" : [
      {
         "frozen" : 994580000000000,
         "asset_symbol" : "bty",
         "balance" : 106699200000,
         "address" : "1PSYYfCbtSeT1vJTvSKmQvhz8y6VhtddWi",
         "total" : 994686699200000,
         "asset_exec" : "coins"
      },
      {
         "asset_exec" : "coins",
         "total" : 993884199600000,
         "balance" : 1118199600000,
         "address" : "1NtVQebPQ5o5G8MrTqDiTctXm45V5c2E4y",
         "asset_symbol" : "bty",
         "frozen" : 992766000000000
      },
      {
         "balance" : 302799200000,
         "address" : "1BG9ZoKtgU5bhKLpcsrncZ6xdzFCgjrZud",
         "total" : 965781299200000,
         "asset_exec" : "coins",
         "frozen" : 965478500000000,
         "asset_symbol" : "bty"
      },
      {
         "frozen" : 928861500000000,
         "asset_symbol" : "bty",
         "address" : "1AFKj1SpvLinKAFK1kKK9rbPD4iJQbgfJH",
         "balance" : 299699300000,
         "total" : 929161199300000,
         "asset_exec" : "coins"
      },
      {
         "total" : 921825995000000,
         "asset_exec" : "coins",
         "address" : "1FiDC6XWHLe7fDMhof8wJ3dty24f6aKKjK",
         "balance" : 158495000000,
         "asset_symbol" : "bty",
         "frozen" : 921667500000000
      }
   ],
   "id" : 1
}
$ curl http://47.100.234.232:20012/bityuan -d@topAsset.p2.json | json_pp
{
   "result" : [
      {
         "total" : 918831598900000,
         "frozen" : 918663000000000,
         "balance" : 168598900000,
         "address" : "1AH9HRd4WBJ824h9PP1jYpvRZ4BSA4oN6Y",
         "asset_symbol" : "bty",
         "asset_exec" : "coins"
      },
      {
         "total" : 909315998900000,
         "address" : "1Lw6QLShKVbKM6QvMaCQwTh5Uhmy4644CG",
         "balance" : 247498900000,
         "frozen" : 909068500000000,
         "asset_exec" : "coins",
         "asset_symbol" : "bty"
      },
      {
         "asset_exec" : "coins",
         "asset_symbol" : "bty",
         "total" : 902193998800000,
         "address" : "1KNGHukhbBnbWWnMYxu1C7YMoCj45Z3amm",
         "balance" : 30498800000,
         "frozen" : 902163500000000
      },
      {
         "asset_exec" : "coins",
         "asset_symbol" : "bty",
         "total" : 900754089658700,
         "frozen" : 900660000000000,
         "balance" : 94089658700,
         "address" : "12gRhkP2BBe9FM2EDEPYu2PqeXWS1x4zyW"
      },
      {
         "total" : 900363998500000,
         "frozen" : 780318000000000,
         "balance" : 120045998500000,
         "address" : "1FB8L3DykVF7Y78bRfUrRcMZwesKue7CyR",
         "asset_symbol" : "bty",
         "asset_exec" : "coins"
      }
   ],
   "id" : 1,
   "error" : null
}

```

### 4.6 Account.TopAssetCount

 1. 功能：返回在coins和ticket合约中主币最多的地址列表数量(定制功能给浏览器用)
 1. 参数：无
 1. 非实时数据，使用缓存，第一次调用会慢， 一小时更新一次

 说明
```
# topAssetCount.json 
{
   "method" : "Account.TopAssetCount",
   "params" : [
      { }
   ],
   "id" : 1
}

$ curl http://47.100.234.232:20016/bityuan -d@topAssetCount.json 
{"id":1,"result":21272,"error":null}

```
## 5 Token

### 5.2 Token.TxList
 推荐使用 Tx.TxList
	
1. 查询token交易列表
    * request:
    ```
    {
        "method":"Token.TxList",
        "params":[
            {
                "page":{
                    "number":1,
                    "size":10
                },
                "sort":[
                    {
                        "key":"height",
                        "ascending":true
                    }
                ]
            }
        ]
    }
    ```
    * response:
    ```
    {
        "id":null,
        "result":[
            {
                "height_index":14000003,,
                "height":0,
                "block_time":1514533394,
                "block_hash":"0x67c58d6ba9175313f0468ae4e0ddec946549af7748037c2fdd5d54298afd20b6",
                "success":true,
                "index":3,
                "hash":"0xd29b6a7a7167100f73cbf25b8967fac4e5a9e316d20a65a4065df236849219a2",
                "from":"1HT7xU2Ngenf7D4yocz2SAcnNLW7rK8d4E",
                "to":"1PUiGcbsccfxW3zuvHXZBJfznziph5miAo",
                "execer":"token",
                "amount":1000000000000,
                "fee":0,
                "action_name":"transfer",
                "group_count":0,
                "is_withdraw":false,
                "options":null,
                "assets":[
    
                ]
            },
            {
                "height_index":14000006,
                "height":0,
                "block_time":1514533394,
                "block_hash":"0x67c58d6ba9175313f0468ae4e0ddec946549af7748037c2fdd5d54298afd20b6",
                "success":true,
                "index":6,
                "hash":"0x9282a2cbc1549b76ce0f09ca9b9758bff1b32f888553b861de976d4546c8310e",
                "from":"1HT7xU2Ngenf7D4yocz2SAcnNLW7rK8d4E",
                "to":"1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX",
                "execer":"token",
                "amount":1000000000000,
                "fee":0,
                "action_name":"transfer",
                "group_count":0,
                "is_withdraw":false,
                "options":null,
                "assets":[
    
                ]
            }
        ],
        "error":null
    }
    ```
### 5.2 Token.TxCount

推荐使用 Tx.TxCount

查询token交易数量  

    * request:

    ```
    {
        "method":"Token.TxCount",
        "params":[
    
        ]
    }
    ```
    * response:
    ```
    {
        "id":null,
        "result":18134,
        "error":null
    }
    ```
### 5.3 Token.AddrsInfo

3. 代币总量、代币持有人列表及余额  
    * request：
    ```
    {
        "method":"Token.AddrsInfo",
        "params":[
            {
                "page":{
                    "number":1,
                    "size":10
                },
                "tokens":[
                    "ABCDE",
                    "TEST"
                ]
            }
        ]
    }
    ```
    * response：
    ```
    {
        "id":null,
        "result":[
            {
                "symbol":"ABCDE",
                "total":100000000000,
                "addrs":[
                    {
                        "addr":"1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK",
                        "balance":10000000000
                    },
                    {
                        "addr":"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR",
                        "balance":20000000000
                    }
                ]
            }
        ],
        "error":null
    }
    ```
### 5.4 Token.ListToken

4. 列出所有token及其发行量
    * request:
    ```
    {
        "method":"Token.ListToken",
        "params":[
    
        ]
    }
    ```
    * response:
    ```
    {
        "id":null,
        "result":{
            "ABCDE":100000000000,
            "TEST":100000000000
        },
        "error":null
    }
    ```
### 5.5 Token.TokenCount

5. 已发行的token数
    * request:
    ```
    {
        "method":"Token.TokenCount",
        "params":[
    
        ]
    }
    ```
    * response:
    ```
    {
        "id":null,
        "result":2,
        "error":null
    }
    ```
### 5.6 Token.TokenAddrCount

6. token相关的地址数量  
    * request:
    ```
{
   "id" : 1,
   "method" : "Token.TokenAddrCount",
   "params" : [
      [
         {
            "match" : [
               {
                  "value" : "token",
                  "key" : "asset_exec"
               },
               {
                  "key" : "asset_symbol",
                  "value" : "ABCDE"
               },
               {
                  "key" : "type",
                  "value" : "personage"
               }
            ]
         },
         {
            "match" : [
               {
                  "value" : "token",
                  "key" : "asset_exec"
               },
               {
                  "value" : "TEST",
                  "key" : "asset_symbol"
               },
               {
                  "key" : "type",
                  "value" : "personage"
               }
            ]
         }
      ]
   ]
}


    ```
    * response:
    ```
    {
        "id":null,
        "result":[
            12,
            132
        ],
        "error":null
    }
    ```

## 6 Tx 

根据各种条件查找交易列表或统计数据

### 6.1 Tx.TxCount

 * request

```
{
   "id" : 1,
   "method" : "Tx.TxCount",
   "params" : [
      {
         "match" : [
            {
               "value" : "coins",
               "key" : "execer"
            }
         ]
      }
   ]
}


```

 response

```
{
   "result" : 224,
   "error" : null,
   "id" : 1
}

```

### 6.2 Tx.TxList 

request

```
{
   "id" : 1,
   "method" : "Tx.TxList",
   "params" : [
      {
         "match" : [
            {
               "value" : "coins",
               "key" : "execer"
            }
         ],
         "sort" : [
            {
               "key" : "height",
               "ascending" : true
            }
         ],
         "page" : {
            "number" : 3,
            "size" : 2
         }
      }
   ]
}

```

response

```
{
   "error" : null,
   "result" : [
      {
         "success" : true,
         "block_time" : 1514533394,
         "action_name" : "genesis",
         "options" : null,
         "is_withdraw" : false,
         "height_index" : 7,
         "amount" : 10000000000000000,
         "index" : 7,
         "fee" : 0,
         "from" : "1HT7xU2Ngenf7D4yocz2SAcnNLW7rK8d4E",
         "group_count" : 0,
         "height" : 0,
         "to" : "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp",
         "block_hash" : "0x67c58d6ba9175313f0468ae4e0ddec946549af7748037c2fdd5d54298afd20b6",
         "assets" : [],
         "hash" : "0x2bf133bddd389b89575af330225011101e25ab58d0be26504eebd6661c7aec9d",
         "execer" : "coins"
      },
      {
         "block_time" : 1514533394,
         "action_name" : "genesis",
         "options" : null,
         "is_withdraw" : false,
         "success" : true,
         "fee" : 0,
         "from" : "1HT7xU2Ngenf7D4yocz2SAcnNLW7rK8d4E",
         "group_count" : 0,
         "height_index" : 0,
         "amount" : 1000000000000,
         "index" : 0,
         "block_hash" : "0x67c58d6ba9175313f0468ae4e0ddec946549af7748037c2fdd5d54298afd20b6",
         "height" : 0,
         "to" : "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",
         "hash" : "0x397809683b447f6355ce437117de737a389b9708fe668b8f09e6c41c42923dc0",
         "execer" : "coins",
         "assets" : []
      }
   ],
   "id" : 1
}

```

### 6.3 Tx.ExecerInfo

过期

合约统计交易数: 用 Tx.TxCount 参数 execer ， 可以有同样的功能

request 

```
{
	"id" : 1 ,
	"method" : "Tx.ExecerInfo", 
        "params":[
		"coins"
        ]
}


```

response

```
{
   "result" : {
      "Intro" : "coins",
      "name" : "coins",
      "tx_count" : 224
   },
   "id" : 1,
   "error" : null
}


```

### 6.4 Tx.BlockList

获得区块列表： 完整的区块列表或交易列表请用区块链节点接口

内部调试用


request

 1. 用page 限制大小
 1. 用range height参数进行翻页
 1. 没有其他信息

```
{
	"id" : 1 ,
	"method" : "Tx.BlockList", 
        "params":[

			{ 
			}
        ]
}


```

response

```
{
   "id" : 1,
   "result" : [
      {
         "block_hash" : "0xe4bba9c8d55d6e1b2609c5442244595d4cecb07097a27fd2a3ea823edb527bd3",
         "block_time" : 1565578606,
         "height" : 618
      },
      {
         "block_hash" : "0x630cda1443e4e4e468cc2e0880e437da8277be09a3ae63409a675e60d2e8ee94",
         "block_time" : 1565578626,
         "height" : 619
      },
      {
         "block_time" : 1565578758,
         "height" : 639,
         "block_hash" : "0xca7d80ea61f1129eff2eb4c16bb5a04d533b4161779208f5dacc3f858ea2594d"
      },
      {
         "block_hash" : "0x89951603d51400865c4b1fa09119e80cbd620c0d280a968453a2b34aa7ae59db",
         "height" : 642,
         "block_time" : 1565578767
      },
      {
         "block_hash" : "0x1bf3d085121438b7e9c4954f6db20ae39a3ea66d5efb3b70b9ea1866fb116356",
         "height" : 646,
         "block_time" : 1565578777
      },
      {
         "block_hash" : "0x6551ccc75988ba8ed96d5b0fe7c372cf17c6ce2128ea44c1b5a0e74d866caeaa",
         "block_time" : 1565578781,
         "height" : 648
      },
      {
         "height" : 650,
         "block_time" : 1565578809,
         "block_hash" : "0x1b6327a99722aeef681d77343c714a17c0e7a604fc47501979117c6aae5fc986"
      },
      {
         "block_hash" : "0x3c5413189735f12d0395708db85e2fe0bc3cf5ff285445ff8a27495d5e0ad371",
         "block_time" : 1565578869,
         "height" : 659
      },
      {
         "block_hash" : "0x196dea46b3b6d7175dc892704e34db24ee551e6608edded035c208c14bea79d0",
         "height" : 667,
         "block_time" : 1565578927
      },
      {
         "height" : 668,
         "block_time" : 1565578927,
         "block_hash" : "0xe6381a546f6212d2d7776439f2de645f4e31941a85cc72a58d87cf1873a5eccc"
      }
   ],
   "error" : null
}

```


## 7 Trade 交易的查询接口

### 7.1  Trade.ListTx  7.2 Trade.ListOrder

 1. Trade.ListTx    查询交易
 1. Trade.ListOrder 查询订单

request : 查询通用参数

```

{
        "id" : 1 ,
        "method" : "Trade.ListOrder", 
        "params":[{ 
   		"page" : {
   		   "number" : 1,
   		   "size" : 20
   		},
   		"sort" : [
   		   {
   		      "ascending" : true,
   		      "key" : "height_index"
   		   }
   		],
   		"match" : [
   		   {
   		      "value" : "145HxiUdRSzt49kKA2WoDyjBcYAJB67M46",
   		      "key" : "send"
   		   }
   		]
                
        }]
}

不同的params 
 
{
   "page" : {
      "number" : 1,
      "size" : 20
   },
   "sort" : [
      {
         "ascending" : true,
         "key" : "key"
      }
   ],
   "match" : [
      {
         "value" : "1DPKUEVwcqHYeEa94fCizrMYjzy9x1owe8",
         "key" : "send"
      }
   ]
}

{
   "page" : {
      "number" : 1,
      "size" : 20
   },
   "sort" : [
      {
         "ascending" : true,
         "key" : "key"
      }
   ],
   "match" : [
      {
         "value" : "TOKEN1",
         "key" : "asset_symbol"
      },
      {
         "value" : true,
         "key" : "is_sell"
      }
   ],
   "range" : [
      {
         "key" : "boardlot_price",
         "start" : 1,
         "end" : 10
      },
      {
         "key" : "height",
         "start" : 1,
         "end" : 10000
      }
   ]
}

```

  返回是交易或订单的列表

 订单

```

{
   "traded_boardlot" : 30,
   "is_finished" : false,
   "txHash" : "Hash",
   "buy_id" : "buy-id",
   "boardlot_price" : 2,
   "sell_id" : "SellID",
   "asset_symbol" : "Test",
   "height" : 1,
   "height_index" : 102222,
   "boardlot_amount" : 1,
   "min_boardlot" : 30,
   "asset_exec" : "token",
   "hash" : "common.ToHex",
   "owner" : "linjing-address",
   "is_sell" : true,
   "ts" : 2222,
   "total_boardlot" : 100,
   "index" : 2,
   "send" : "tx.From",
   "status" : "created",
   "price_exec" : "coins",
   "price_symbol" : "bty"
}
```

 交易

```
{
   "height" : 1,
   "send" : "tx.From",
   "ts" : 2222,
   "txHash" : "Hash",
   "height_index" : 100002,
   "hash" : "common.ToHex",
   "index" : 2,
   "tx_type" : "sell_limit",
   "sell_limit" : null,
   "success" : true
}
```
### 7.3 Trade.ListAsset

 获得trade资产

 * request 

```
{
        "id" : 1 ,
        "method" : "Trade.ListAsset", 
        "params":[{ 
        }]
}

```
 
 * responce

```
{
   "error" : null,
   "result" : [
      {
         "asset_symbol" : "TEST",
         "asset_exec" : "token",
         "price_exec" : "coins",
         "price_symbol" : "bty"
      }
   ],
   "id" : 1
}

```

### 7.4  Trade.ListLastPrice

获得最新价格

 * request 

```
{
        "id" : 1 ,
        "method" : "Trade.ListLastPrice", 
        "params":[ 
		[{"asset_exec":"token", "asset_symbol":"TEST"}]
        ]
}

默认 price_exec = "coins",  price_symbol = 主币名称, 参数中可以增加指定price信息
```

 * response

```
{
   "result" : [
      {
         "asset_exec" : "token",
         "boardlot_amount" : 1000000,
         "height" : 58322,
         "boardlot_price" : 10000,
         "asset_symbol" : "TEST",
         "index" : 1,
         "price_exec" : "coins",
         "price_symbol" : "bty"
      }
   ],
   "error" : null,
   "id" : 1
}

```

## 8 MultiSig


### 8.1 MultiSig.Search

 按owner搜索多重签名地址
 
 1. request:

```
{
	"id" : 1 ,
	"method" : "MultiSig.Search", 
	"params":[{ 
		"match" : [
			{"key": "address", "value" : "1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj"},
			{"key": "type", "value" : "owner"}
		]		
	}]
}

```

 1. response： 返回多重签名地址， 详细信息可以调用Gets 获得

```
{
   "error" : null,
   "id" : 1,
   "result" : [
      "31s9k1GFBkrYCSUSsDderRB3BvxocbeDKW",
      "38QFFgjSUozg6cGjis6cxACUJvkpfTD7ZS"
   ]
}
```

### 8.2 MultiSig.Searches

多个查询

 1. request: 按owner， 按create_address 

```
{
   "method" : "MultiSig.Searches",
   "params" : [
      [
         {
            "match" : [
               {
                  "key" : "address",
                  "value" : "1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj"
               },
               {
                  "key" : "type",
                  "value" : "owner"
               }
            ]
         },
         {
            "match" : [
               {
                  "value" : "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",
                  "key" : "create_address"
               },
               {
                  "value" : "account",
                  "key" : "type"
               }
            ]
         }
      ]
   ],
   "id" : 1
}

```

```
{
   "result" : [
      [
         "38QFFgjSUozg6cGjis6cxACUJvkpfTD7ZS",
         "31s9k1GFBkrYCSUSsDderRB3BvxocbeDKW"
      ],
      [
         "31s9k1GFBkrYCSUSsDderRB3BvxocbeDKW",
         "38QFFgjSUozg6cGjis6cxACUJvkpfTD7ZS"
      ]
   ],
   "id" : 1,
   "error" : null
}

```


### 8.3 MultiSig.Gets

 * requset: 入参是多重签名的地址

```
{
   "id" : 1,
   "method" : "MultiSig.Gets",
   "params" : [
      {
         "address" : [
            "38QFFgjSUozg6cGjis6cxACUJvkpfTD7ZS",
            "31s9k1GFBkrYCSUSsDderRB3BvxocbeDKW"
         ]
      }
   ]
}

```

 * response

```
{
   "id" : 1,
   "result" : [
      {
         "tx_count" : 0,
         "owners" : [
            {
               "weight" : 20,
               "address" : "1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK",
               "type" : "owner",
               "multi_signature_address" : "31s9k1GFBkrYCSUSsDderRB3BvxocbeDKW"
            },
            {
               "address" : "1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj",
               "weight" : 10,
               "multi_signature_address" : "31s9k1GFBkrYCSUSsDderRB3BvxocbeDKW",
               "type" : "owner"
            },
            {
               "weight" : 30,
               "address" : "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",
               "type" : "owner",
               "multi_signature_address" : "31s9k1GFBkrYCSUSsDderRB3BvxocbeDKW"
            }
         ],
         "required_weight" : 15,
         "create_address" : "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",
         "multi_signature_address" : "31s9k1GFBkrYCSUSsDderRB3BvxocbeDKW",
         "limits" : [
            {
               "execer" : "coins",
               "multi_signature_address" : "31s9k1GFBkrYCSUSsDderRB3BvxocbeDKW",
               "daily_limit" : 1000000000,
               "last_day" : 1565578589,
               "type" : "limit",
               "symbol" : "BTY",
               "spent_today" : 0
            }
         ],
         "type" : "account"
      },
      {
         "owners" : [
            {
               "type" : "owner",
               "multi_signature_address" : "38QFFgjSUozg6cGjis6cxACUJvkpfTD7ZS",
               "weight" : 30,
               "address" : "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
            },
            {
               "weight" : 30,
               "address" : "1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK",
               "type" : "owner",
               "multi_signature_address" : "38QFFgjSUozg6cGjis6cxACUJvkpfTD7ZS"
            },
            {
               "address" : "1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj",
               "weight" : 10,
               "multi_signature_address" : "38QFFgjSUozg6cGjis6cxACUJvkpfTD7ZS",
               "type" : "owner"
            }
         ],
         "tx_count" : 12,
         "type" : "account",
         "limits" : [
            {
               "symbol" : "BTY",
               "spent_today" : 0,
               "execer" : "coins",
               "multi_signature_address" : "38QFFgjSUozg6cGjis6cxACUJvkpfTD7ZS",
               "last_day" : 1565578152,
               "daily_limit" : 1200000000,
               "type" : "limit"
            }
         ],
         "required_weight" : 15,
         "multi_signature_address" : "38QFFgjSUozg6cGjis6cxACUJvkpfTD7ZS",
         "create_address" : "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
      }
   ],
   "error" : null
}

```
 
## 20 获得挖矿帐号的余额 

给挖矿监控用

### 挖矿地址的余额

 1. url:  http://47.111.132.89:19991/bityuan
 1. request body: 在addresses 中指定需要地址列表， 地址列表为空的情况下， 会获得默认的一组信息

```
{
	"id" : 1 ,
	"method" : "MinerAccount.ShowAccounts", 
	"params":[{ 
		"addresses" : ["1bgg6HwQretMiVcSWvayPRvVtwjyKfz1J"]
	}]
}

```


 1. 默认的地址列表， 来自挖矿机器：addrNewYi.txt， 具体在附录中
 1. response: 每个地址将获得在coins， ticket 里的余额信息

```
{
   "error" : null,
   "id" : 1,
   "result" : [
      {
         "type" : "personage",
         "asset_exec" : "coins",
         "frozen" : 0,
         "total" : 137924147,
         "asset_symbol" : "bty",
         "balance" : 137924147,
         "height_index" : 276936400001,
         "exec" : "",
         "address" : "1bgg6HwQretMiVcSWvayPRvVtwjyKfz1J"
      },
      {
         "asset_symbol" : "bty",
         "total" : 969926200000,
         "frozen" : 900000000000,
         "asset_exec" : "coins",
         "type" : "contractInternal",
         "address" : "1bgg6HwQretMiVcSWvayPRvVtwjyKfz1J",
         "exec" : "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp",
         "height_index" : 276834200001,
         "balance" : 69926200000
      }
   ]
}

``` 

### 获得最近的挖矿信息


 1. url:  http://47.111.132.89:19991/bityuan
 1. request body: 在addresses 中指定需要地址列表， 地址列表为空的情况下， 会获得默认的一组信息

```
{
	"id" : 1 ,
	"method" : "MinerAccount.ShowMinerStatus", 
	"params":[{ 
		"addresses" : ["1bgg6HwQretMiVcSWvayPRvVtwjyKfz1J", "1LszcYi5sCkrvrHHyZFxkC14NF2563xNXF"]
	}]
}
```

 1. 默认的地址列表， 来自挖矿机器：addrNewYi.txt， 具体在附录中
 1. response: 每个地址将获得在近期的挖矿信息. 最近1小时， 1天的挖矿次数 和最后一次挖矿的高度和时间

```
{
   "result" : [
      {
         "mine_count_last_day" : 0,
         "mine_count_last_hour" : 0,
         "address" : "1bgg6HwQretMiVcSWvayPRvVtwjyKfz1J",
         "last_mine_ts" : 1563691542,
         "last_mine_height" : 2781691
      },
      {
         "last_mine_ts" : 1563786139,
         "last_mine_height" : 2800429,
         "mine_count_last_day" : 292,
         "mine_count_last_hour" : 13,
         "address" : "1LszcYi5sCkrvrHHyZFxkC14NF2563xNXF"
      }
   ],
   "error" : null,
   "id" : 1
}


``` 
 

## 默认的地址列表， 来自挖矿机器：addrNewYi.txt， 需要更新联系我

```
   	"1Q5QcUaDXET3RJ3UBurMZzF3gGHyjnFQEa",
	"1M8E2TnHFgxRisCuMXuYHMfELbVk4AkrCh",
	"1PXhFDUTiR69vn9EdbJ9AyviR6wzjsVJuL",
	"19zQrJrtZTkLHDMogRo4vaw3QzNoqTn8aX",
	"1Gv6QVMz24yGMg6UYQ2TXSr64sng48Wj8W",
	"1GR1Ubkt4u2owmSHEWarsYkA1HgCvJ3rFq",
	"1EqgQMSsHxkDHPVZvr8jMQkBAaeC1Ujkn5",
	"1Ff2bGhKS5YFEbNBWrHwzbChizrMZhyi8e",
	"19MaTsSzsZBTpFcubANJ1DMVXKLfaA92Zv",
	"13SnepxL5RELnLwxGTWFkMJMgMJqJhP4FV",
	"1K5aR21sU7xDEwcQ7G1fPXbCJW6R5kLxjZ",
	"1Laq81McJfNwFB2uqdJTosogoaBVm8x6Z2",
	"121qqR7Kb7CcwuHfx9GYkSZhx6L6w8GmzH",
	"1EVGErxMR45eWu28moAQank4A8dQf5b698",
	"19QPCDcrPPPnYDAKeMPCaBFnra5vdjcUoR",
	"1Hy8oXArAPSWFu1P3hHvUaxCNjacnwLKeb",
	"16vAAd4WeqWMCJjeTciUTvJHVj9GuPAyWQ",
	"18AiuYrj2UKUBeFKrCw5mLCC9xyyWxiu4z",
	"137esbqhb3fv8gQMWZEZAqxmSktzXA9jSr",
	"1M4ns1eGHdHak3SNc2UTQB75vnXyJQd91s",
	"1C9M1RCv2e9b4GThN9ddBgyxAphqMgh5zq",
	"1HFUhgxarjC7JLru1FLEY6aJbQvCSL58CB",
	"1EwkKd9iU1pL2ZwmRAC5RrBoqFD1aMrQ2",
	"19ozyoUGPAQ9spsFiz9CJfnUCFeszpaFuF",
	"1NbLdeZYqdaYHFyjBu9evsAmBzfRpiQAZP",
	"1MoEnCDhXZ6Qv5fNDGYoW6MVEBTBK62HP2",
	"12T8QfKbCRBhQdRfnAfFbUwdnH7TDTm4vx",
	"1bgg6HwQretMiVcSWvayPRvVtwjyKfz1J",
	"12pLjJkEAzJCnX9o1enC1yF5MpAffexaKi",
	"1E4UF3HW8LEwdbJjF9FzKhNFeqhp5zNm8",

```
