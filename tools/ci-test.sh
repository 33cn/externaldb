#!/usr/bin/env bash

set -x

esseq_prefix="seq_"
esdb_prefix="db_"

function start() {
    docker-compose up &
}

function stop() {
    docker-compose down
}

# 看数据库, seq 是否已经大于某个值
function wait_process() {
    while true; do
        seq=$(curl 127.0.0.1:9200/${esdb_prefix}last_seq/last_seq/convert_bty3 2>/dev/null | jq '._source.sync_seq')
        if [ "$seq" == "" ] || [ "$seq" == "null" ]; then
            echo "wait for sync/convert"
            sleep 1
            continue
        fi

        if [ "$seq" -gt "4" ]; then
            break
        else
            echo "wait for sync/convert"
            sleep 1
        fi
    done
}

function debug_function() {
    set -x
    "$@"
    set +x
}

function test_seq() {
    hash=$(curl 127.0.0.1:9200/${esseq_prefix}seq/seq/2 | jq ._source.hash)
    if [ "$hash" == '"0x6d61696e686173682d30303031"' ]; then
        echo "== Pass test_seq"
    else
        echo "== Failed test_seq"
        TEST_ERR=${TEST_ERR}+". Failed test_seq"
    fi
}

function test_proof_config() {
    role=$(curl '127.0.0.1:9200/'${esdb_prefix}'proof_config/proof_config/member-1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k' | jq '._source.role')
    if [ "$role" == '"member"' ]; then
        echo "== Pass test_proof_config"
    else
        echo "== Failed test_proof_config"
        TEST_ERR=${TEST_ERR}+". Failed test_proof_config"
    fi
}

# 填写测试
TEST_ERR=""
function tests() {
    debug_function test_seq
    debug_function test_proof_config
    echo "TODO more test for contract data and rpc"
}

start
wait_process
tests
stop
if [ "$TEST_ERR" != "" ]; then
    echo "Test failed"
    exit 1
else
    echo "Test success"
fi
