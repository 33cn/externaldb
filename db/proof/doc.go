// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proof

/*
 1.proof存证数据在es数据库中的格式，json字符串

 1.1 对key值添加了一个proof_的前缀，在通过key值查询时，对前端输入的原始key值增加前缀
 1.2 查询结果返回时，需要将key值的前缀去掉。
{
	"proof_data": "[{\"key\":\"user-id-0\",\"data\":{\"value\":\"0\",\"type\":\"number\",\"format\":\"int\"}},{\"key\": \"user-id-1\",\"data\": [{\"value\": \"0\",\"type\": \"number\",\"format\": \"int32\"}, {\"value\": \"2\",\"type\": \"number\",\"format\": \"int64\"}]},{\"key\": \"user-id-2\",\"data\": {\"key\": \"user-id-3\",\"data\": {\"value\": \"1\",\"type\": \"number\",\"format\": \"int32\"}}},{\"key\": \"user-id-4\",\"data\": [{\"key\": \"user-id-5\",\"data\": {\"value\": \"1\",\"type\": \"number\",\"format\": \"int32\"}}, {\"key\": \"user-id-6\",\"data\": {\"value\": \"2\",\"type\": \"number\",\"format\": \"int32\"}}]},{\"key\": \"ref_hashes\",\"data\": [{\"value\": \"0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f\",\"type\": \"file\",\"format\": \"hash\"}, {\"value\": \"0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56\",\"type\": \"file\",\"format\": \"hash\"}]}]",
	"proof_height_index": 100000,
	"proof_organization": "fuzamei",
	"proof_ref_hashes": ["0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f", "0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56"],
	"proof_sender": "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",
	"proof_tx_hash": "0xc9e1a33446f140f12548496708c804b6cedc31b9025b11925997f590e1e2ec7d",
	"proof_user-id-0": 0,
	"proof_user-id-1": [0, 2],
	"proof_user-id-2": {
		"user-id-3": 1
	},
	"proof_user-id-4": [{
		"user-id-5": 1
	}, {
		"user-id-6": 2
	}]
}

2. 目前支持如下四种json格式

2.1 一个k对应一个value值
{
	"key": "user-id-0",
	"data": {
		"value": "0",
		"type": "number",
		"format": "int32"
	}
}

2.2 一个key对应多个value值
{
	"key": "user-id-0",
	"data": [{
		"value": "0",
		"type": "number",
		"format": "int32"
	}, {
		"value": "2",
		"type": "number",
		"format": "int64"
	}]
}

2.3 一个key对应一个k:v对obj
{
	"key": "user-id-0",
	"data": {
		"key": "user-id-1",
		"data": {
			"value": "1",
			"type": "number",
			"format": "int32"
		}
	}
}

2.4 一个key嵌套多个k:v对obj
{
	"key": "user-id-0",
	"data": [{
		"key": "user-id-1",
		"data": {
			"value": "1",
			"type": "number",
			"format": "int32"
		}
	}, {
		"key": "user-id-2",
		"data": {
			"value": "2",
			"type": "number",
			"format": "int32"
		}
	}]
}

3. 目前支持如下查询接口
3.1 List Proof
3.1.1  根据上传者地址获取存证列表，分页显示 List Proof
{
   "id" : 1,
   "method" : "Proof.List",
   "params" : [
      {
         "match" : [
            {
               "key" : "proof_sender",
               "value" : "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
            }
         ],
         "page" : {
            "number" : 1,
            "size" : 10
         }
      }
   ]
}

3.1.2  根据组织者名称获取存证列表，分页显示 List Proof
{
   "id" : 1,
   "method" : "Proof.List",
   "params" : [
      {
         "match" : [
            {
               "key" : "proof_organization",
               "value" : "fuzamei"
            }
         ],
         "page" : {
            "number" : 1,
            "size" : 10
         }
      }
   ]
}

3.2 Count Proof
3.2.1  通过上传者账户获取已经上传的存证数量
{
   "method" : "Proof.Count",
   "params" : [
      {
         "match" : [
            {
               "key" : "proof_organization",
               "value" : "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
            }
         ]
      }
   ],
   "id" : 1
}

3.2.2  通过组织者名称获取已经上传的存证数量
{
   "method" : "Proof.Count",
   "params" : [
      {
         "match" : [
            {
               "key" : "proof_sender",
               "value" : "fuzamei"
            }
         ]
      }
   ],
   "id" : 1
}

3.3 Show Proof
3.3.1  通过交易hash获取存证信息
{
   "method" : "Proof.Show",
   "params" : [
      {
         "match" : [
            {
               "key" : "proof_tx_hash",
               "value" : "0xc9e1a33446f140f12548496708c804b6cedc31b9025b11925997f590e1e2ec7d"
            }
         ]
      }
   ],
   "id" : 1
}

3.3 Gets 通过存证的交易hash列表获取多个存证信息
{
	"id" : 1 ,
	"method" : "Proof.Gets",
	"params":[
		{"hashs" :["0xa7904af41d9886c01873c7cfdda471e9d839c42eb6e4e1223c6e118b33b94322","0x25c86923535ee1a209725450ba270110e5446c3afbdeb8112990f86160d81eb0"]}
	]
}

4. jrcp接口测试使用json文件
curl  http://localhost:9992/v1/proof/Count  -d@proof.count.organization.json | json_pp
curl  http://localhost:9992/v1/proof/Count  -d@proof.count.sender.json | json_pp

curl  http://localhost:9992/v1/proof/List  -d@proof.list.organization.json | json_pp
curl  http://localhost:9992/v1/proof/List  -d@proof.list.sender.json | json_pp

curl  http://localhost:9992/v1/proof/Show  -d@proof.show.json | json_pp

curl  http://localhost:9992/v1/proof/Gets  -d@proof.gets.json | json_pp

5.一条proof的es存储数据
var jsonstr = `{"proof_data": "[{\"key\":\"user-id-0\",\"data\":{\"value\":\"0\",\"type\":\"number\",\"format\":\"int\"}},{\"key\": \"user-id-1\",\"data\": [{\"value\": \"0\",\"type\": \"number\",\"format\": \"int32\"}, {\"value\": \"2\",\"type\": \"number\",\"format\": \"int64\"}]},{\"key\": \"user-id-2\",\"data\": {\"key\": \"user-id-3\",\"data\": {\"value\": \"1\",\"type\": \"number\",\"format\": \"int32\"}}},{\"key\": \"user-id-4\",\"data\": [{\"key\": \"user-id-5\",\"data\": {\"value\": \"1\",\"type\": \"number\",\"format\": \"int32\"}}, {\"key\": \"user-id-6\",\"data\": {\"value\": \"2\",\"type\": \"number\",\"format\": \"int32\"}}]},{\"key\": \"ref_hashes\",\"data\": [{\"value\": \"0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f\",\"type\": \"file\",\"format\": \"hash\"}, {\"value\": \"0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56\",\"type\": \"file\",\"format\": \"hash\"}]}]","proof_height_index": 100000,"proof_organization": "fuzamei","proof_ref_hashes": ["0xa5f9d70546c60b264dc62de3a94561b0c93317294d0a56cf5d759b1e7076468f", "0x29d9edcec9e8b4265040429474cc846311c382f1038af7e9d3a00c1d30139b56"],"proof_sender": "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv","proof_tx_hash": "0xc9e1a33446f140f12548496708c804b6cedc31b9025b11925997f590e1e2ec7d","proof_user-id-0": 0,"proof_user-id-1": [0, 2],"proof_user-id-2": {"user-id-3": 1},"proof_user-id-4": [{"user-id-5": 1}, {"user-id-6": 2}]}`


6.目前支持将存证信息解析成嵌套模式的kv和非嵌套模式的kv
6.1 嵌套模式：
"产品参数": [
        {
          "产品名称": "碧根果"
        },
        {
          "生产商": "浙江杭州临安锦北街道"
        }
      ],

6.2 非嵌套模式：只解析最底层的kv对

"产品名称": "碧根果"，
"生产商": "浙江杭州临安锦北街道"

*/
