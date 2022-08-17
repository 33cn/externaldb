#!/usr/bin/env bash

set -x

# run in ci dir

sedfix=""
if [ "$(uname)" == "Darwin" ]; then
    sedfix=".bak"
fi

appConfigDir=$(pwd)/configs
appDataDir=$(pwd)/datas
mkdir "${appConfigDir}" -p
mkdir "${appDataDir}" -p
chmod 777 "${appDataDir}"

dockerfile="docker-compose.yaml"
sed -i $sedfix "s|APP_CONFIG_DIR|$appConfigDir|" ${dockerfile}
sed -i $sedfix "s|APP_DATA_DIR|$appDataDir|" ${dockerfile}

sed -i $sedfix 's|para=""|para="user.p.testproof."|' configs/convert.toml
