#!/usr/bin/env bash

# es启动后，等待一段时间才能可用，通常为几秒，需要等待es可用后，再启动程序
while true; do
    start=$(curl 127.0.0.1:9200/_cat/health 2>/dev/null | grep -c -v red)
    if [ "$start" == 1 ]; then break; else sleep 1; fi
done

/root/sync -f /root/config/externaldb.toml
