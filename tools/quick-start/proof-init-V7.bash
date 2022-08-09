#!/usr/bin/env bash

. ../../.env

prefix=${CONVERT_PREFIX}
port=${ES_PORT}
manager=${PROOF_MANAGER}

ES_USER=elastic

OLD_IFS="$IFS"
IFS=","
proof_manager=($manager)
IFS="$OLD_IFS"

managerCount=${#proof_manager[*]}

numberOfShards=5
numberOfReplicas=1

# 用管理员地址替换
curl -X PUT http://localhost:${port}/${prefix}proof_config -d '{"settings":{"number_of_shards":'${numberOfShards}',"number_of_replicas":'${numberOfReplicas}'}}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""
for i in ${proof_manager[*]}; do
  curl -X PUT http://localhost:${port}/${prefix}proof_config/_doc/member-${i}/_create -d'{"address":"'${i}'","role": "manager","organization": "system", "note" : "system-init-manager"}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
  echo ""
done

curl -X PUT http://localhost:${port}/${prefix}proof_config_org -d'{"settings":{"number_of_shards":'${numberOfShards}',"number_of_replicas":'${numberOfReplicas}'}}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""
curl -X PUT http://localhost:${port}/${prefix}proof_config_org/_doc/org-system/_create -d'{"count": '${managerCount}',"organization": "system", "note" : "system-init"}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""

curl -X PUT http://localhost:${port}/${prefix}proof -d '{"settings":{"number_of_shards":'${numberOfShards}',"number_of_replicas":'${numberOfReplicas}'}}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""
curl -X PUT http://localhost:${port}/${prefix}proof/_doc/proof-create-proof-index/_create -d '{"init": 1}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""

curl -X PUT http://localhost:${port}/${prefix}proof/_mapping?pretty -d '{"properties": {"捐赠数量": {"type": "double"}}}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""
curl -X PUT http://localhost:${port}/${prefix}proof/_mapping?pretty -d '{"properties": {"志愿者数量": {"type": "long"}}}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""

curl -X PUT http://localhost:${port}/${prefix}proof/_settings -d '{"index":{"max_result_window":20000}}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""
curl -X PUT http://localhost:${port}/${prefix}proof/_settings -d '{"index.mapping.total_fields.limit":1000000}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""

curl -X DELETE http://localhost:${port}/${prefix}proof/_doc/proof-create-proof-index -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""

curl -X PUT http://localhost:${port}/${prefix}block_info/_settings -d '{ "index" : { "max_result_window" : 100000}}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""
curl -X PUT http://localhost:${port}/${prefix}transaction/_settings -d '{ "index" : { "max_result_window" : 100000}}' -H "Content-Type: application/json" -u $ES_USER:$ES_PASSWORD
echo ""