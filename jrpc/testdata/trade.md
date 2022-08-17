
## 交易的查询接口

 1. Trade.ListTx    查询交易
 1. Trade.ListOrder 查询订单

## 交易的查询参数

由于交易查询的需求多种多样， 查询请求做的已经通用。 主要提供4个项

 1. page： 分页。 不提供的话， 会提供默认个
 1. sort： 是数组，可以提供多个排序条件。相对于 sql 的 order by。
    1.  会提供 key 为 height_index 通用的上链顺序
    1.  其他的业务相关的， 如价格(boardlot_price)等
 1. match：数组。 查询的筛选条件. 交易中可以提供
    1.  send： 提供地址， 发起交易的人 
    1.  asset_symbol: 交易的资产 (如果存在不同合约里有同名的资产， 可以同时提供 asset_exec 指定合约)
 1. range: 提供区间查询
    1. start, end 包含临界值
    1. gt, lt 不包含临界值(great than, less than)

``` 
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

## 返回是交易或订单的列表

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
   "status" : "created"
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
## 获得trade资产

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
         "asset_exec" : "token"
      }
   ],
   "id" : 1
}

```

## 获得最新价格

 * request 

```
{
        "id" : 1 ,
        "method" : "Trade.ListLastPrice", 
        "params":[ 
		[{"asset_exec":"token", "asset_symbol":"TEST"}]
        ]
}
```

 * response

```
{
   "result" : [
      {
         "exec_asset" : "token",
         "boardlot_amount" : 1000000,
         "height" : 58322,
         "boardlot_price" : 10000,
         "asset_symbol" : "TEST",
         "index" : 1
      }
   ],
   "error" : null,
   "id" : 1
}

```
