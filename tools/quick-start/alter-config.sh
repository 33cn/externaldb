#!/bin/bash
# shellcheck disable=SC2086
CONF=./etc/externaldb.toml

. .env

set_key_value() {
    local key value row current
    key=$1
    value=$2
    row=$3
    if [ -n "$value" ] && [ "$value" != \"\" ]; then
        current=$(sed -n -e "${row}s/^\($key = \)\([^ ']*\)\(.*\)$/\2/p" $CONF) # value带单引号
        echo "current: $key = $current"
        if [ -n "$current" ]; then
            echo "setting $CONF : $key = $value $row"
            sed -i "${row}s|${key} = ${current}|${key} = ${value}|" ${CONF}
        fi
    fi
}

combine_es_host() {
    local key value row current
    key=$1
    value=$2
    row=$3
    if [ -n "$value" ] && [ "$value" != \"\" ]; then
        current=$(sed -n -e "${row}s/^\($key = \)\([^ ']*\)\(.*\)$/\2/p" $CONF) # value带单引号
        echo "current: $key = $current"
        if [ -n "$current" ]; then
            local es_host="\"http://localhost:$value/\""
            echo "setting $CONF : $key = $es_host $row"
            sed -i "${row}s|${key} = ${current}|${key} = ${es_host}|" ${CONF}
        fi
    fi
}

deal_browser_config() {
    line1=$(sed -n "/======start======/=" $CONF)
    line2=$(sed -n "/======end======/=" $CONF)

    start=$(("$line1" + 1))
    end=$(("$line2" - 1))

    if [ "$EX_SERVICE_TYPE" = "proof" ]; then
        sed -i "${start},${end}s/^/#/g" $CONF
        set_key_value "openAccessControl" true
        set_key_value "saveBlockInfo" false
    else
        sed -i "${start},${end}s/^#*\s*//g" $CONF
        set_key_value "openAccessControl" false
        set_key_value "saveBlockInfo" true
    fi
}

# 如果在配置文件中key值不是唯一的，则要加上行号进行定位。===>第一个参数为key值，第二个参数为value值，第三个参数为行号
# chain相关
set_key_value "title" \"${CHAIN_TITLE}\"
set_key_value "host" \"${CHAIN_HOST}\" 10
set_key_value "grpcHost" \"${CHAIN_GRPC_HOST}\"
set_key_value "symbol" \"${CHAIN_SYMBOL}\"
# es相关
set_key_value "esVersion" ${ES_VERSION}

combine_es_host "host" ${ES_PORT} 26
set_key_value "prefix" \"${SYNC_PREFIX}\" 27
set_key_value "pwd" \"${ES_PASSWORD}\" 31

combine_es_host "host" ${ES_PORT} 35
set_key_value "prefix" \"${CONVERT_PREFIX}\" 36
set_key_value "pwd" \"${ES_PASSWORD}\" 38
# sync相关
set_key_value "pushBind" \"${PUSH_BIND}\"
set_key_value "pushHost" \"${PUSH_HOST}\"
set_key_value "pushName" \"${PUSH_NAME}\"
set_key_value "pushFormat" \"${PUSH_FORMAT}\"
set_key_value "startSeq" ${START_SEQ} 58
set_key_value "startHeight" ${START_HEIGHT}
# rpc相关
set_key_value "host" \"${RPC_HOST}\" 67
set_key_value "jrpcHost" \"${JRPC_HOST}\"
set_key_value "swaggerHost" \"${SWAGGER_HOST}\"
# convert相关
set_key_value "addressDriver" \"${ADDR_DRIVER}\"
set_key_value "dealOtherChain" ${DEAL_OTHER_CHAIN}

deal_browser_config
