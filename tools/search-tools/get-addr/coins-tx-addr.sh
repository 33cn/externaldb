#!/usr/bin/env bash
set -x
req="{ \"size\" : 1, \"query\" : { \"match\": { \"to\": \"$1\" } } }"
curl -X GET -d "$req" --header "Content-Type:application/json" http://localhost:9200/coins-tx/coins-tx/_search | json_pp | grep to -w | cut -d '"' -f 4
