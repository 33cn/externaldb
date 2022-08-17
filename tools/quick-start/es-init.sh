#!/bin/bash

. .env

exist_container=$(docker ps | grep -w "${ES_NAME}")
exist_port=$(netstat -tunlp | grep ":${ES_PORT_HTTP}")

function confirm() {
    while true; do
        read -r -p "是否连接已有es [y/n]：" input
        case $input in
            [yY][eE][sS] | [yY])
                return 0
                ;;
            [nN][oO] | [nN])
                return 1
                ;;
            *)
                echo "Invalid input, 请重新输入"
                ;;
        esac
    done
}

function verdict() {
    if [ -n "${exist_container}" ] && [ -n "${exist_port}" ]; then
        echo "es exist"
        confirm
    elif [ -n "${exist_container}" ] && [ -z "${exist_port}" ]; then
        echo "es container name already exist, please to reset container name"
        echo "退出运行"
        return 1
    elif [ -z "${exist_container}" ] && [ -n "${exist_port}" ]; then
        echo "es port already exist, please to reset port"
        echo "退出运行"
        return 1
    else
        echo "es don't exist, to create es"
        mkdir -p "${ES_DIR}"
        chmod -R 777 "${ES_DIR}"
        if [ "${ES_VERSION}" = 6 ]; then
            echo "creating es6"
            cp -f .env ./env/es6/
            docker-compose -f ./env/es6/docker-compose.yaml up -d
        else
            echo "creating es7"
            cp -f .env ./env/es7/
            docker-compose -f ./env/es7/docker-compose.yaml up -d
        fi
    fi
}

verdict
