.PHONY: quickStart run logs logs-convert logs-rpc logs-rpc stop restart rm down clean initConfig initES restartES downES
include .env

quickStart:
	@echo "================start initES================"
	@bash es-init.sh
	@echo "==============waiting es start=============="
	@until [ `curl -o /dev/null -s -w %{http_code} http://localhost:${ES_PORT_HTTP} -u elastic:${ES_PASSWORD}` = 200 ]; do \
        		sleep 1s; \
    done
	@echo "==============start initConfig=============="
	@bash alter-config.sh
	@if [ "${ES_VERSION}" = 6 ]; then \
  		cd ./env/es6 && chmod +x proof-init.bash && ./proof-init.bash && cd ../..; \
	else \
    	cd ./env/es7 && chmod +x proof-init.bash && ./proof-init.bash && cd ../..; \
    fi
	@echo "============start run externaldb============"
	@cd ./bin && chmod +x sync convert rpc jrpc
	@if [ "${EX_SERVICE_TYPE}" = "proof" ]; then \
		docker-compose -p ${PROJECT_NAME} up -d sync convert rpc; \
	elif [ "${EX_SERVICE_TYPE}" = "browser" ]; then \
		docker-compose -p ${PROJECT_NAME} up -d sync convert jrpc; \
	else \
		docker-compose -p ${PROJECT_NAME} up -d; \
	fi
	@echo "==================complete=================="

run:
	@cd ./bin && chmod +x sync convert rpc jrpc
	@if [ "${EX_SERVICE_TYPE}" = "proof" ]; then \
		docker-compose -p ${PROJECT_NAME} up -d sync convert rpc; \
	elif [ "${EX_SERVICE_TYPE}" = "browser" ]; then \
		docker-compose -p ${PROJECT_NAME} up -d sync convert jrpc; \
	else \
		docker-compose -p ${PROJECT_NAME} up -d; \
	fi

log-sync:
	@tail -200f ./bin/logs/sync.log
log-sync-err:
	@tail -200f ./bin/logs/sync.error.log
log-convert:
	@tail -200f ./bin/logs/convert.log
log-convert-err:
	@tail -200f ./bin/logs/convert.error.log

logs:
	@docker-compose -p ${PROJECT_NAME} logs

logs-sync:
	@docker-compose -p ${PROJECT_NAME} logs sync

logs-convert:
	@docker-compose -p ${PROJECT_NAME} logs convert

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
	@if [ "${ES_VERSION}" = 6 ]; then \
        cd ./env/es6 && chmod +x proof-init.bash && ./proof-init.bash && cd ../..; \
    else \
    	cd ./env/es7 && chmod +x proof-init.bash && ./proof-init.bash && cd ../..; \
    fi

alterExConfig:
	@bash alter-config.sh

alterEsInit:
	@if [ "${ES_VERSION}" = 6 ]; then \
        cd ./env/es6 && ./proof-init.bash && cd ../..; \
    else \
    	cd ./env/es7 && ./proof-init.bash && cd ../..; \
    fi

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
	curl -s --location --request GET http://localhost:${ES_PORT_HTTP}/_cat/indices?v -u elastic:${ES_PASSWORD} | grep "${CONVERT_PREFIX}\|${SYNC_PREFIX}\|docs.count"

search-nft:
	curl -s --location --request GET http://localhost:${ES_PORT_HTTP}/${CONVERT_PREFIX}nft/_search?pretty -u elastic:${ES_PASSWORD}

search-contract:
	curl -s --location --request GET http://localhost:${ES_PORT_HTTP}/${CONVERT_PREFIX}contract/_search?pretty -u elastic:${ES_PASSWORD}

reset-convert:stop clean-convert-indices initConfig run
	@echo done

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
	@cp ./bin/convert ./bin-update/

run-update:
	@docker-compose -f docker-compose-update.yaml -p ${PROJECT_NAME} up -d

stop-update:
	@docker-compose -f docker-compose-update.yaml -p ${PROJECT_NAME} down

switch-update: stop-update stop
	@cp ./etc/externaldb-update.toml ./etc/externaldb.toml
	@make run