.PHONY: quickStart run logs logs-convert logs-rpc logs-rpc stop restart rm down clean initConfig initES restartES downES
include .env

quickStart:
	@echo "================start initES================"
	@bash es-init.sh
	@echo "================waiting  10s================"
	@sleep 10s
	@echo "==============start initConfig=============="
	@bash alter-config.sh
	@echo "============start run externaldb============"
	@chmod +x ./bin/*
	@docker-compose -p ${PROJECT_NAME} up -d
	@echo "==================complete=================="

run:
	@chmod +x ./bin/*
	@docker-compose -p ${PROJECT_NAME} up -d

logs:
	@docker-compose -p ${PROJECT_NAME} logs

logs-sync:
	@docker-compose -p ${PROJECT_NAME} logs sync

logs-rpc:
	@docker-compose -p ${PROJECT_NAME} logs rpc

stop:
	@docker-compose -p ${PROJECT_NAME} stop

restart:
	@docker-compose -p ${PROJECT_NAME} restart

rm:
	@docker-compose -p ${PROJECT_NAME} rm

down:
	@docker-compose -p ${PROJECT_NAME} down

clean:
	@rm -rf ./bin/logs

initConfig:
	@bash alter-config.sh

initES:
	@bash es-init.sh

restartES:
	@if [ "${ES_VERSION}" = 6 ]; then \
	    docker-compose -f ./env/es6/docker-compose.yaml restart; \
    else \
    	docker-compose -f ./env/es7/docker-compose.yaml restart; \
	fi

downES:
	@if [ "${ES_VERSION}" = 6 ]; then \
	    docker-compose -f ./env/es6/docker-compose.yaml down; \
    else \
    	docker-compose -f ./env/es7/docker-compose.yaml down; \
	fi

es-indices:
	curl -s --location --request GET http://localhost:${ES_PORT_HTTP}/_cat/indices?v -u elastic:${ES_PASSWORD} | grep "${CONVERT_PREFIX}\|docs.count"

search-nft:
	curl -s --location --request GET http://localhost:${ES_PORT_HTTP}/${CONVERT_PREFIX}nft/_search?pretty -u elastic:${ES_PASSWORD}

search-contract:
	curl -s --location --request GET http://localhost:${ES_PORT_HTTP}/${CONVERT_PREFIX}contract/_search?pretty -u elastic:${ES_PASSWORD}

clean-convert-indices:
	curl -s --location --request DELETE http://localhost:${ES_PORT_HTTP}/${CONVERT_PREFIX}* -u elastic:${ES_PASSWORD}

search-evm-%:
	curl -s --location --request GET http://localhost:${ES_PORT_HTTP}/${CONVERT_PREFIX}/_doc/evm-${*} -u elastic:${ES_PASSWORD}

clean-file-cache:
	curl -s --location --request DELETE http://localhost:${ES_PORT_HTTP}/${CONVERT_PREFIX}file_part -u elastic:${ES_PASSWORD}

# 数据需要重新同步时可采用此方法在后台同步，同步完成后更新
.PHONY: init-update run-update stop-update switch-update
init-update:
	@cp ./etc/externaldb.toml ./etc/externaldb-bak.toml
	@make initConfig
	@cp ./etc/externaldb.toml ./etc/externaldb-update.toml
	@cp ./etc/externaldb-bak.toml ./etc/externaldb.toml
	@mkdir -p ./bin-update/
	@cp ./bin/sync ./bin-update/

run-update:
	@docker-compose -f docker-compose-update.yaml -p ${PROJECT_NAME} up -d

stop-update:
	@docker-compose -f docker-compose-update.yaml -p ${PROJECT_NAME} down

switch-update: stop-update stop
	@cp ./etc/externaldb-update.toml ./etc/externaldb.toml
	@make run