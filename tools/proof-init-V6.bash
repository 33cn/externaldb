#!/usr/bin/env bash

prefix="db_"

namagerCount=2
namager1="1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY"
namager2="1E89P2n6RsAE1K4HfFNF71cGdqjoRzuxD2"

# 用管理员地址替换
curl -X PUT http://localhost:9200/${prefix}proof_config/proof_config/member-${namager1}/_create -d'{"address":"'${namager1}'","role": "manager","organization": "system", "note" : "system-init-manager"}' -H "Content-Type: application/json"
echo ""
curl -X PUT http://localhost:9200/${prefix}proof_config/proof_config/member-${namager2}/_create -d'{"address":"'${namager2}'","role": "manager","organization": "system", "note" : "system-init-manager"}' -H "Content-Type: application/json"
echo ""
curl -X PUT http://localhost:9200/${prefix}proof_config_org/proof_config_org/org-system/_create -d'{"count": '${namagerCount}',"organization": "system", "note" : "system-init"}' -H "Content-Type: application/json"
echo ""
curl -X PUT http://localhost:9200/${prefix}proof/proof/proof-create-proof-index/_create -d '{"init": 1}' -H "Content-Type: application/json"
echo ""
curl -X PUT http://localhost:9200/${prefix}proof/proof/_mapping?pretty -d '{"proof": {"properties": {"捐赠数量": {"type": "double"}}}}' -H "Content-Type: application/json" 
echo ""
curl -X PUT http://localhost:9200/${prefix}proof/proof/_mapping?pretty -d '{"proof": {"properties": {"志愿者数量": {"type": "long"}}}}' -H "Content-Type: application/json"
echo ""
curl -X PUT http://localhost:9200/${prefix}proof/_settings -d '{ "index" : { "max_result_window" : 20000}}' -H "Content-Type: application/json"
echo ""
curl -X PUT  http://localhost:9200/${prefix}proof/_settings -d '{"index.mapping.total_fields.limit":1000000}' -H "Content-Type: application/json"
echo ""
curl -X DELETE http://localhost:9200/${prefix}proof/proof/proof-create-proof-index -H "Content-Type: application/json"
