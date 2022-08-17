#!/bin/bash
file_path=$(
    cd "$(dirname "$0")" || exit
    pwd
)/../..

swag2md_name="swag2md"

port="9992"

echo "start generating api.md"
"${file_path}/${swag2md_name}" \
    -t "存证展开服务RPC接口文档

> 👉 [Swagger 文档](http://172.16.101.87:${port}/swagger/index.html)" \
    -s "${file_path}/rpc/docs/swagger.json" \
    -o "${file_path}/rpc-doc.md"
echo "generating rpc-doc.md success"
