#!/usr/bin/env bash

prefix="db_"

managerCount=2
manager1="1E89P2n6RsAE1K4HfFNF71cGdqjoRzuxD2"
manager2="1868J6fNJLmk5r3Hb6vR3HcUwrkMkVBQkX"

numberOfShards=1
numberOfReplicas=1

# 用管理员地址替换
curl -X PUT http://localhost:9200/${prefix}proof_config -d '{"settings":{"number_of_shards":'${numberOfShards}',"number_of_replicas":'${numberOfReplicas}'}}' -H "Content-Type: application/json"
echo ""
curl -X PUT http://localhost:9200/${prefix}proof_config/_doc/member-${manager1}/_create  -d'{"address":"'${manager1}'","role": "manager","organization": "system", "note" : "system-init-manager"}' -H "Content-Type: application/json"

echo ""
curl -X PUT http://localhost:9200/${prefix}proof_config/_doc/member-${manager2}/_create  -d'{"address":"'${manager2}'","role": "manager","organization": "system", "note" : "system-init-manager"}' -H "Content-Type: application/json"
echo ""

curl -X PUT http://localhost:9200/${prefix}proof_config_org -d'{"settings":{"number_of_shards":'${numberOfShards}',"number_of_replicas":'${numberOfReplicas}'}}' -H "Content-Type: application/json"
echo ""
curl -X PUT http://localhost:9200/${prefix}proof_config_org/_doc/org-system/_create  -d'{"count": '${managerCount}',"organization": "system", "note" : "system-init"}' -H "Content-Type: application/json"

echo ""
curl -X PUT  http://localhost:9200/${prefix}proof -d '{"settings":{"number_of_shards":'${numberOfShards}',"number_of_replicas":'${numberOfReplicas}'}}' -H "Content-Type: application/json"
echo ""
curl -X PUT http://localhost:9200/${prefix}proof/_doc/proof-create-proof-index/_create -d '{"init": 1}' -H "Content-Type: application/json"
echo ""
curl -X PUT http://localhost:9200/${prefix}proof/_mapping?pretty -d '{"properties": {"捐赠数量": {"type": "double"}}}' -H "Content-Type: application/json"
echo ""
curl -X PUT http://localhost:9200/${prefix}proof/_mapping?pretty -d '{"properties": {"志愿者数量": {"type": "long"}}}' -H "Content-Type: application/json"

echo ""
curl -X PUT http://localhost:9200/${prefix}proof/_settings -d '{"index":{"max_result_window":20000}}' -H "Content-Type: application/json"

echo ""
curl -X PUT  http://localhost:9200/${prefix}proof/_settings -d '{"index.mapping.total_fields.limit":1000000}' -H "Content-Type: application/json"
echo ""
curl -X DELETE http://localhost:9200/${prefix}proof/_doc/proof-create-proof-index -H "Content-Type: application/json"
