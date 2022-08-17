#!/bin/bash

version=1.2
suffix="t"
docker_repo_host="172.16.100.249:5000"

function tag_push() {
    module=$1
    docker tag chain33/"${module}":${version} ${docker_repo_host}/"${module}":${version}${suffix}
    docker push ${docker_repo_host}/"${module}":${version}${suffix}

}
#docker tag chain33/externaldb/sync:1.2 172.16.100.249:5000/externaldb/sync:1.2t

tag_push externaldb/sync
tag_push externaldb/convert
tag_push externaldb/rpc
tag_push externaldb/jrpc
