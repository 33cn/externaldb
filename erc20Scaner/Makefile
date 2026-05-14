# Makefile for ERC20 Scanner
# Static builds match mkpkg.sh: CGO_ENABLED=0 + stripped/packed ldflags.

.PHONY: help build up down logs restart clean txinfo blockparse scanner rpc-server bins static

# Same defaults as mkpkg.sh (linux/amd64 fully static binary when CGO_ENABLED=0).
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED := 0
GO_LDFLAGS := -s -w

GO_BUILD := CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(GO_LDFLAGS)"

help:
	@echo "Available commands:"
	@echo "  make bins - Static-build scanner, rpc-server, txinfo, blockparse -> ./bin/"
	@echo "  make static     - Same as bins (alias)"
	@echo "  make scanner    - Static-build ./bin/scanner"
	@echo "  make rpc-server - Static-build ./bin/rpc-server"
	@echo "  make txinfo     - Static-build ./bin/txinfo"
	@echo "  make blockparse - Static-build ./bin/blockparse (parse one block by height)"
	@echo "  make build      - Build Docker images"
	@echo "  make up         - Start all services"
	@echo "  make down       - Stop all services"
	@echo "  make logs       - View logs"
	@echo "  make restart    - Restart all services"
	@echo "  make clean      - Remove containers and volumes"
	@echo "  make db-shell   - Connect to MySQL"
	@echo "  make test-api   - Test API endpoints"
	@echo ""
	@echo "Override cross-compile: make bins GOOS=darwin GOARCH=arm64"

build:
	docker-compose build

bin:
	@mkdir -p bin

scanner: bin
	$(GO_BUILD) -o bin/scanner ./scanner/*.go

rpc-server: bin
	$(GO_BUILD) -o bin/rpc-server ./rpc/*.go

txinfo: bin
	$(GO_BUILD) -o bin/txinfo ./cmd/txinfo/

blockparse: bin
	$(GO_BUILD) -o bin/blockparse ./cmd/blockparse/

bins: scanner rpc-server txinfo blockparse
	@echo "Built (static, CGO_ENABLED=0): bin/scanner bin/rpc-server bin/txinfo bin/blockparse"

static: bins

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

restart:
	docker-compose restart

clean:
	docker-compose down -v
	docker system prune -f

db-shell:
	docker-compose exec mysql mysql -uroot -p$${MYSQL_ROOT_PASSWORD:-rootpassword} token_scanner

test-api:
	@echo "Testing API endpoints..."
	@curl -s http://localhost:8080/health | jq .
	@echo "\nTesting contracts list..."
	@curl -s "http://localhost:8080/api/contracts?page=1&size=5" | jq .

status:
	docker-compose ps
