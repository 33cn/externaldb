#!/bin/bash

# config
host="localhost:9202"
prefix="es1_bty2_"

# 修改交易1万条上限
curl -X PUT http://${host}/${prefix}transaction/_settings -d '{ "index" : { "max_result_window" : 20000}}' -H "Content-Type: application/json"
