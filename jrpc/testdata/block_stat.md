
## 4 获得区块相关信息

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

### StatGets

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

### Gets

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

### StatSearch

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

### Search

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

### StatSearches 

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


### Searches

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
