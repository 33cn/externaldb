#!/usr/bin/env bash

set -x
# 等待es启动时没用的, 需要等待es可用
while true; do
    start=$(curl 127.0.0.1:9200/_cat/health 2>/dev/null | grep -c -v red)
    if [ "$start" == 1 ]; then break; else sleep 1; fi
done

# init proof_config super manage
prefix="db_"
curl -X PUT http://localhost:9200/${prefix}proof_config/proof_config/member-1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY/_create -d'{"address":"1E5saiXVb9mW8wcWUUZjsHJPZs5GmdzuSY","role": "manager","organization": "system", "note" : "system-init-manager"}' -H "Content-Type: application/json"
curl -X PUT http://localhost:9200/${prefix}proof_config_org/proof_config_org/org-system/_create -d'{"count": 1,"organization": "system", "note" : "system-init"}' -H "Content-Type: application/json"

/root/convert -f /root/config/externaldb.toml -c /root/config/chain33.toml
