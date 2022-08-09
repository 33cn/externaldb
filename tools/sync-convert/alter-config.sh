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

set_proof_manager() {
    local key value row
    key=$1
    value=$2
    row=$3
    sed -i "${row}d" ${CONF}
    sed -i "${row}i${key} = ${value}" ${CONF}
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

# 如果在配置文件中key值不是唯一的，则要加上行号进行定位。===>第一个参数为key值，第二个参数为value值，第三个参数为行号
# chain相关
set_key_value "title" \"${CHAIN_TITLE}\"
set_key_value "host" \"${CHAIN_HOST}\" 12
set_key_value "symbol" \"${CHAIN_SYMBOL}\"
# es相关
set_key_value "esVersion" ${ES_VERSION}
set_proof_manager "managerAddress" ${PROOF_MANAGER} 6

combine_es_host "host" ${ES_PORT} 23
set_key_value "prefix" \"${CONVERT_PREFIX}\" 24
set_key_value "pwd" \"${ES_PASSWORD}\" 28
# sync相关
set_key_value "pushBind" \"${PUSH_BIND}\"
set_key_value "pushHost" \"${PUSH_HOST}\"
set_key_value "pushName" \"${PUSH_NAME}\"
# rpc相关
set_key_value "host" \"${RPC_HOST}\" 57
set_key_value "host" \"${JRPC_HOST}\" 58
set_key_value "swaggerHost" \"${SWAGGER_HOST}\"
