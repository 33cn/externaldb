
## 获得挖矿帐号的余额

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

## 获得最近的挖矿信息


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
