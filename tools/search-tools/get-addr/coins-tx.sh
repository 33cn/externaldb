#!/usr/bin/env bash
curl -X GET -d "@coins-tx-from.json" --header "Content-Type:application/json" http://localhost:9200/coins-tx/coins-tx/_search | json_pp
