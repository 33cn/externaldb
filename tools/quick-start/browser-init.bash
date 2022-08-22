#!/usr/bin/env bash

. .env

curl -X PUT http://localhost:"$ES_PORT"/"$CONVERT_PREFIX"block_info/_settings -d '{ "index" : { "max_result_window" : 100000}}' -H "Content-Type: application/json" -u elastic:"$ES_PASSWORD"
echo ""
curl -X PUT http://localhost:"$ES_PORT"/"$CONVERT_PREFIX"transaction/_settings -d '{ "index" : { "max_result_window" : 100000}}' -H "Content-Type: application/json" -u elastic:"$ES_PASSWORD"
echo ""

# 浏览器从0开始同步数据时，为了提升写入性能，使用此配置
curl -XPUT http://localhost:"$ES_PORT"/"$CONVERT_PREFIX"*/_settings?preserve_existing=true -u elastic:"$ES_PASSWORD" -H "Content-Type: application/json"  -d '{
    "index.translog.durability" : "async",
    "index.translog.flush_threshold_size" : "512mb",
    "index.translog.sync_interval" : "3s"
}'

curl -XPUT http://localhost:"$ES_PORT"/"$CONVERT_PREFIX"account,"$CONVERT_PREFIX"address/_settings -u elastic:"$ES_PASSWORD" -H 'Content-Type: application/json' -d '{
    "index": {
        "refresh_interval":"60s"
   }
}'

# 浏览器数据同步到最新时，恢复默认配置，保证account数据的实时性
#curl -XPUT http://localhost:"$ES_PORT"/"$CONVERT_PREFIX"*/_settings?preserve_existing=true -u elastic:"$ES_PASSWORD" -H "Content-Type: application/json"  -d '{
#    "index.translog.durability" : "fsync",
#    "index.translog.flush_threshold_size" : "512mb",
#}'
#
#curl -XPUT http://localhost:"$ES_PORT"/"$CONVERT_PREFIX"account,"$CONVERT_PREFIX"address/_settings -u elastic:"$ES_PASSWORD" -H 'Content-Type: application/json' -d '{
#    "index": {
#        "refresh_interval":"1s"
#   }
#}'