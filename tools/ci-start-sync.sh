#!/usr/bin/env bash

# 等待es启动时没用的, 需要等待es可用
while true; do
    start=$(curl 127.0.0.1:9200/_cat/health 2>/dev/null | grep -c -v red)
    if [ "$start" == 1 ]; then break; else sleep 1; fi
done

/root/sync -f /root/config/externaldb.toml
