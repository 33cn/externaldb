# golang1.12 or latest
# 1. make help
# 2. make dep
# 3. make build
# ...
SRC := github.com/33cn/externaldb
SRC_CLI := ${SRC}/cmd/
SRC_RPC := ${SRC}/jrpc
PRC := build/jrpc
CLI_SYNC := build/sync
CLI_CONVERT := build/convert

AUTOTEST := build/autotest/autotest
SRC_AUTOTEST := github.com/33cn/chain33/cmd/autotest

PKG_LIST := `go list ./... | grep -v "mocks"`
PKG_LIST_VET := `go list ./... | grep -v "common/log/log15"`
PKG_LIST_INEFFASSIGN= `go list -f {{.Dir}} ./...  grep -v "common/log/log15"`
PKG_LIST_Q := `go list ./... | grep -v "mocks"`
PKG_LIST_GOSEC := `go list -f "{{.Dir}}" ./... | grep -v "mocks" | grep -v "cmd" | grep -v "types" | grep -v "commands" | grep -v "log15"`

LDFLAGS := -ldflags "-w -s"
BUILD_FLAGS = -ldflags "-X github.com/33cn/externaldb/version.GitCommit=`git rev-parse --short=8 HEAD` -X github.com/33cn/externaldb/version.ReleaseDate=`date +%Y%m%d`"
MKPATH=$(abspath $(lastword $(MAKEFILE_LIST)))
MKDIR=$(dir $(MKPATH))
DAPP := ""
PROJ := "build"

.PHONY: default dep all build release linter race test fmt vet bench msan coverage coverhtml docker docker-compose protobuf clean help autotest

default: build
include .env

dep: ## Get the dependencies
	@go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.40.1
	@go get -u golang.org/x/tools/cmd/goimports
	@go get -u github.com/mitchellh/gox
	@go get -u github.com/vektra/mockery/.../
	@go get -u mvdan.cc/sh/cmd/shfmt
	@go get -u mvdan.cc/sh/cmd/gosh
	@git checkout go.mod go.sum
	@apt install clang-format || echo run \"apt install clang-format\" by root
	@apt install shellcheck || echo run \"apt install shellcheck\" by root
	@go get -u github.com/swaggo/swag/cmd/swag@v1.7.8

.PHONY: proto
proto: ## Generate protbuf file of types package
	@cd ./proto && ./create_protobuf.sh && cd ..

.PHONY: swag
swag: ## gen rpc doc with swag
	@cd ./rpc && swag init && cd ..

help: ## Display this help screen
	@printf "Help doc:\nUsage: make [command]\n"
	@printf "[command]\n"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

fmt: fmt_go fmt_proto fmt_shell ## fmt .go .proto .sh

.PHONY: fmt_go fmt_proto fmt_shell
fmt_go: ## fmt go
	find . -name '*.go' -not -path "./vendor/*" | xargs gofmt -s -w
	find . -name '*.go' -not -path "./vendor/*" | xargs goimports -l -w

fmt_proto: ## go fmt protobuf file
	@find . -name '*.proto' -not -path "./vendor/*" | xargs clang-format -i

fmt_shell: ## check shell file
	find . -name '*.sh' -not -path "./vendor/*" | xargs shfmt -w -s -i 4 -ci -bn

shcheck: ## check shell
	shellcheck tools/*sh

.PHONY: checkgofmt
checkgofmt: ## get all go files and run go fmt on them
	@files=$$(find . -name '*.go' -not -path "./vendor/*" | xargs gofmt -l -s); if [ -n "$$files" ]; then \
		  echo "Error: 'make fmt' needs to be run on:"; \
		  find . -name '*.go' -not -path "./vendor/*" | xargs gofmt -l -s ;\
		  exit 1; \
		  fi;
	@files=$$(find . -name '*.go' -not -path "./vendor/*" | xargs goimports -l -w); if [ -n "$$files" ]; then \
		  echo "Error: 'make fmt' needs to be run on:"; \
		  find . -name '*.go' -not -path "./vendor/*" | xargs goimports -l -w ;\
		  exit 1; \
		  fi;

.PHONY: doc
doc: swag
	@go build $(BUILD_FLAGS) -v -i -o swag2md ${SRC}/tools/swag2md
	@bash ./tools/doc/doc.sh

.PHONY: code_gen
code_gen: swag

build: code_gen ## Build the binary file
	@go build $(BUILD_FLAGS) -v -i -o  $(PRC) $(SRC_RPC)
	@go build $(BUILD_FLAGS) -v -i -o  $(CLI_SYNC) $(SRC_CLI)/sync
	@go build $(BUILD_FLAGS) -v -i -o  $(CLI_CONVERT) $(SRC_CLI)/convert
	@go build $(BUILD_FLAGS) -v -i -o  build/rpc $(SRC)/rpc/
	@go build $(BUILD_FLAGS) -v -i -o  build/dummy_node $(SRC_CLI)/dummy_node/
	@go build $(BUILD_FLAGS) -v -i -o  build/sync_convert $(SRC_CLI)/sync_convert
	@cp config/externaldb.toml build/externaldb.toml

.PHONY: build_convert
build_convert: proto
	@go build $(BUILD_FLAGS) -v -i -o  $(CLI_CONVERT) $(SRC_CLI)/convert

.PHONY: build_rpc
build_rpc: code_gen
	@go build $(BUILD_FLAGS) -v -i -o  build/rpc $(SRC)/rpc/

release: ## Build the binary file
	@go build $(LDFLAGS) -v -i -o  $(PRC) $(SRC_RPC)
	@go build $(LDFLAGS) -v -i -o  $(CLI_SYNC) $(SRC_CLI)/sync
	@go build $(LDFLAGS) -v -i -o  $(CLI_CONVERT) $(SRC_CLI)/convert
	@go build $(BUILD_FLAGS) -v -i -o  build/rpc $(SRC)/rpc/
	@cp config/externaldb.toml build/externaldb.toml

clean: ## Remove previous build
	@rm -rf build/*
	@go clean ./...

test: ## Run unittests
	@go test -race $(PKG_LIST)

testq: ## Run unittests
	@go test $(PKG_LIST_Q)

bench: ## Run benchmark of all
	@go test ./... -v -bench=.

autotest: test ## build autotest binary

autotest_ci: test ## autotest jerkins ci

build_ci: dep build ## Build the binary file for CI

linter: vet ineffassign gosec ## Use gometalinter check code, ignore some unserious warning
	@./tools/golinter.sh "filter"
	@find . -name '*.sh' -not -path "./vendor/*" | xargs shellcheck -x

linter_test: ## Use gometalinter check code, for local test
	@./tools/golinter.sh "test" "${p}"
	@find . -name '*.sh' -not -path "./vendor/*" | xargs shellcheck -x

gosec:
	@golangci-lint  run --no-config --issues-exit-code=1  --deadline=2m --disable-all --enable=gosec ${PKG_LIST_GOSEC}

race: ## Run data race detector
	@go test -race -short $(PKG_LIST)

vet:
	@go vet ${PKG_LIST_VET}

ineffassign:
	@golangci-lint  run --no-config --issues-exit-code=1  --deadline=2m --disable-all   --enable=ineffassign   -n ${PKG_LIST_INEFFASSIGN}


msan: ## Run memory sanitizer
	@go test -msan -short $(PKG_LIST)

coverage: ## Generate global code coverage report
	@./build/tools/coverage.sh

coverhtml: ## Generate global code coverage report in HTML
	@./build/tools/coverage.sh html



.PHONY: auto_ci
auto_fmt := find . -name '*.go' -not -path './vendor/*' | xargs goimports -l -w
auto_ci: clean fmt_proto fmt_shell protobuf mock swag
	@-find . -name '*.go' -not -path './vendor/*' | xargs gofmt -l -w -s
	@-${auto_fmt}
	@-find . -name '*.go' -not -path './vendor/*' | xargs gofmt -l -w -s
	@${auto_fmt}
	@git status
	@files=$$(git status -suno);if [ -n "$$files" ]; then \
		  git add *.go *.sh *.proto; \
		  git status; \
		  git commit -a -m "auto ci"; \
		  git push origin HEAD:$(branch); \
		  exit 1; \
		  fi;

.PHONY: update_chain33
update_chain33: ## make update_chain33 version=tag/branch/commit-hash
	if [ "${version}" = "" ]; then \
		go mod edit -require github.com/33cn/chain33@master; \
	else \
		go mod edit -require github.com/33cn/chain33@${version}; \
	fi; \
	go mod tidy

update_plugin: ## make update_plugin version=tag/branch/commit-hash
	if [ "${version}" = "" ]; then \
		go mod edit -require github.com/33cn/plugin@master; \
	else \
		go mod edit -require github.com/33cn/plugin@${version}; \
	fi; \
	go mod tidy

docker: build ## build docker image for externaldb with sync convert jrpc
	cp tools/*Dockerfile build
	cp tools/ci-*.sh build
	cp tools/sync-convert/sync-convert.Dockerfile build
	if [ "${version}" = "" ]; then \
		cd build && \
		docker build . -f sync.Dockerfile   -t chain33/externaldb/sync:lastest; \
		docker build . -f convert.Dockerfile -t chain33/externaldb/convert:lastest; \
		docker build . -f jrpc.Dockerfile -t chain33/externaldb/jrpc:lastest; \
		docker build . -f rpc.Dockerfile -t chain33/externaldb/rpc:lastest; \
		docker build . -f dummy-node.Dockerfile -t chain33/externaldb/dummy-node:lastest; \
		docker build . -f sync-convert.Dockerfile -t chain33/externaldb/sync-convert:lastest; \
		cd ..; \
	else \
		cd build && \
		docker build . -f sync.Dockerfile   -t chain33/externaldb/sync:${version}; \
		docker build . -f convert.Dockerfile -t chain33/externaldb/convert:${version}; \
		docker build . -f jrpc.Dockerfile -t chain33/externaldb/jrpc:${version}; \
		docker build . -f dummy-node.Dockerfile -t chain33/externaldb/dummy-node:${version}; \
		docker build . -f rpc.Dockerfile -t chain33/externaldb/rpc:${version}; \
		docker build . -f sync-convert.Dockerfile -t chain33/externaldb/sync-convert:${version}; \
		cd ..; \
	fi; 

pkgProof:
	rm -rf externaldb-proof/*
	mkdir -p externaldb-proof/ex/bin externaldb-proof/ex/etc externaldb-proof/es externaldb-proof/mq externaldb-proof/ex-merge/bin externaldb-proof/ex-merge/etc
	cp build/rpc build/convert build/sync tools/proof-init-*.bash externaldb-proof/ex/bin
	cp config/externaldb.toml externaldb-proof/ex/etc/
	cp tools/proof-es.yaml externaldb-proof/es/docker-compose.yaml
	cp tools/externaldb.docker-compose.yaml externaldb-proof/ex/docker-compose.yaml
	cp mq/kafka/docker-compose.yml externaldb-proof/mq
	cp build/rpc build/sync_convert externaldb-proof/ex-merge/bin
	cp config/externaldb_merge.toml externaldb-proof/ex-merge/etc/externaldb.toml
	cp tools/sync-convert/docker-compose.yaml externaldb-proof/ex-merge/docker-compose.yaml
	cp tools/sync-convert/README.md externaldb-proof/ex-merge/README.md
	tar zcvf externaldb-proof.tgz externaldb-proof

pkg-quickStart:
	rm externaldb-proof/* -rf
	mkdir -p externaldb-proof/bin externaldb-proof/etc externaldb-proof/ex-merge/bin externaldb-proof/ex-merge/etc
	mkdir -p externaldb-proof/env/es6 externaldb-proof/env/es7 externaldb-proof/env/mq
	cp build/rpc build/convert build/sync build/jrpc externaldb-proof/bin/
	cp config/externaldb.toml externaldb-proof/etc/
	cp mq/kafka/docker-compose.yml externaldb-proof/env/mq/
	cp tools/quick-start/proof-init-V6.bash externaldb-proof/env/es6/proof-init.bash
	cp tools/quick-start/proof-init-V7.bash externaldb-proof/env/es7/proof-init.bash
	cp tools/quick-start/proof-es6.yaml externaldb-proof/env/es6/docker-compose.yaml
	cp tools/quick-start/proof-es7.yaml externaldb-proof/env/es7/docker-compose.yaml
	cp tools/mime.types externaldb-proof/etc/mime.types
	cp tools/quick-start/docker-compose.yaml externaldb-proof/
	cp tools/quick-start/docker-compose-update.yaml externaldb-proof/
	cp tools/quick-start/Makefile externaldb-proof/
	cp tools/quick-start/README.md externaldb-proof/
	cp tools/quick-start/alter-config.sh externaldb-proof/
	cp tools/quick-start/es-init.sh externaldb-proof/
	cp .env externaldb-proof/
	cp build/rpc build/sync_convert externaldb-proof/ex-merge/bin
	cp config/externaldb_merge.toml externaldb-proof/ex-merge/etc/externaldb.toml
	cp tools/sync-convert/docker-compose.yaml externaldb-proof/ex-merge/docker-compose.yaml
	cp tools/sync-convert/README.md externaldb-proof/ex-merge/README.md
	tar zcvf externaldb-proof.tgz externaldb-proof

pkg-mergeStart:
	rm externaldb-proof/* -rf
	mkdir -p externaldb-proof/bin externaldb-proof/etc
	mkdir -p externaldb-proof/env/es6 externaldb-proof/env/es7 externaldb-proof/env/mq
	cp mq/kafka/docker-compose.yml externaldb-proof/env/mq/
	cp tools/quick-start/proof-es6.yaml externaldb-proof/env/es6/docker-compose.yaml
	cp tools/quick-start/proof-es7.yaml externaldb-proof/env/es7/docker-compose.yaml
	cp tools/mime.types externaldb-proof/etc/mime.types
	cp tools/sync-convert/alter-config.sh externaldb-proof/
	cp tools/sync-convert/es-init.sh externaldb-proof/
	cp tools/sync-convert/Makefile externaldb-proof/
	cp tools/sync-convert/.env externaldb-proof/
	cp build/rpc build/sync_convert externaldb-proof/bin
	cp config/externaldb_merge.toml externaldb-proof/etc/externaldb.toml
	cp tools/sync-convert/docker-compose.yaml externaldb-proof/docker-compose.yaml
	cp tools/sync-convert/README.md externaldb-proof/README.md
	tar zcvf externaldb-proof.tgz externaldb-proof

pkg:
	rm externaldb/* -rf
	mkdir -p externaldb/ex/bin externaldb/ex/etc externaldb/es externaldb/mq
	cp build/jrpc build/convert  build/sync tools/browser-init.bash externaldb/ex/bin
	cp config/externaldb.toml externaldb/ex/etc
	cp tools/proof-es.yaml externaldb/es/docker-compose.yaml
	cp tools/externaldb.docker-compose.yaml externaldb/ex/docker-compose.yaml
	cp mq/kafka/docker-compose.yml externaldb/mq
	cp ./tools/chain33.para.6.4.toml externaldb/ex/etc
	tar zcvf externaldb.tgz externaldb

docker-compose: ## run docker-compose as jenkins-ci
	@cd build && \
		if ! [ -d ci ]; then \
			make -C .. docker; \
			mkdir ci/configs -p; \
		fi; \
		cp ../tools/docker-compose-ci.yaml ./ci/docker-compose.yaml; \
		cp ../tools/ci-*sh ./ci/; \
		cp ../tools/chain33.para.6.3.toml ./ci/configs/chain33.toml; \
		cp *toml ./ci/configs; \
		cd ci && \
		./ci-setup-config.sh; \
		./ci-test.sh \
		cd ..; \
	cd ..

docker-compose-down: ## stop ci 
	@cd build && if [ -d ci ]; then \
	 	cp ../tools/docker-compose-ci.yaml ./ci/docker-compose.yaml; \
	 	cd ci/ && \
		 ./ci-setup-config.sh; \
		 docker-compose down && \
	  	cd .. ; \
		rm ci -rf; \
	 fi; \
	 cd ..
