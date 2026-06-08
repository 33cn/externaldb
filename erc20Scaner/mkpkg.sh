#!/bin/bash

# ERC20 Scanner 打包脚本
# 功能：编译应用并创建可部署的测试包
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="erc20-scanner"
VERSION=$(date +%Y%m%d_%H%M%S)
BUILD_DIR="${SCRIPT_DIR}/_build"
PKG_DIR="${BUILD_DIR}/${PROJECT_NAME}-${VERSION}"
PKG_NAME="${PROJECT_NAME}-${VERSION}.tgz"

# 清理函数（只清理临时文件，保留打包好的文件）
cleanup() {
    if [ -d "${PKG_DIR}" ] && [ -f "${BUILD_DIR}/${PKG_NAME}" ]; then
        # 打包成功，只清理临时目录，保留打包文件
        echo -e "${YELLOW}清理临时文件...${NC}"
        rm -rf "${PKG_DIR}"
    elif [ -d "${BUILD_DIR}" ]; then
        # 打包失败，清理所有临时文件
        echo -e "${YELLOW}清理临时文件...${NC}"
        rm -rf "${BUILD_DIR}"
    fi
}

# 错误处理（只在错误时清理）
trap cleanup ERR

echo "=========================================="
echo "ERC20 Scanner 打包脚本"
echo "=========================================="
echo ""

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未找到 Go 环境，请先安装 Go${NC}"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}Go 版本: ${GO_VERSION}${NC}"

# 创建构建目录
echo ""
echo -e "${YELLOW}创建构建目录...${NC}"
mkdir -p "${PKG_DIR}"/{bin,config,database,logs,env}

# 编译应用
echo ""
echo -e "${YELLOW}编译应用...${NC}"

# 编译 Scanner
echo "  编译 scanner..."
cd "${SCRIPT_DIR}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${PKG_DIR}/bin/scanner" ./scanner/*.go
if [ $? -eq 0 ]; then
    echo -e "  ${GREEN}✓ scanner 编译成功${NC}"
else
    echo -e "  ${RED}✗ scanner 编译失败${NC}"
    exit 1
fi

# 编译 RPC
echo "  编译 rpc-server..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${PKG_DIR}/bin/rpc-server" ./rpc/*.go
if [ $? -eq 0 ]; then
    echo -e "  ${GREEN}✓ rpc-server 编译成功${NC}"
else
    echo -e "  ${RED}✗ rpc-server 编译失败${NC}"
    exit 1
fi

# 显示编译结果
echo ""
echo "编译结果:"
ls -lh "${PKG_DIR}/bin/"

# 复制配置文件
echo ""
echo -e "${YELLOW}复制配置文件...${NC}"

# Docker Compose 文件（从 deploy/ 目录复制）
if [ -f "${SCRIPT_DIR}/deploy/docker-compose.yml" ]; then
    cp "${SCRIPT_DIR}/deploy/docker-compose.yml" "${PKG_DIR}/docker-compose.yml"
    echo -e "  ${GREEN}✓ docker-compose.yml${NC}"
else
    echo -e "  ${RED}✗ deploy/docker-compose.yml 不存在${NC}"
    exit 1
fi

# MySQL Docker Compose 文件 → env/docker-compose.yml
if [ -f "${SCRIPT_DIR}/deploy/mysql-docker-compose.yml" ]; then
    cp "${SCRIPT_DIR}/deploy/mysql-docker-compose.yml" "${PKG_DIR}/env/docker-compose.yml"
    echo -e "  ${GREEN}✓ env/docker-compose.yml${NC}"
else
    echo -e "  ${RED}✗ deploy/mysql-docker-compose.yml 不存在${NC}"
    exit 1
fi

# 复制其他配置文件
if [ -f "${SCRIPT_DIR}/config/config.yaml.example" ]; then
    cp "${SCRIPT_DIR}/config/config.yaml.example" "${PKG_DIR}/config/"
    echo -e "  ${GREEN}✓ config.yaml.example${NC}"
fi

if [ -f "${SCRIPT_DIR}/env.example" ]; then
    cp "${SCRIPT_DIR}/env.example" "${PKG_DIR}/"
    echo -e "  ${GREEN}✓ env.example${NC}"
fi

if [ -f "${SCRIPT_DIR}/mysql.cnf" ]; then
    cp "${SCRIPT_DIR}/mysql.cnf" "${PKG_DIR}/database/"
    echo -e "  ${GREEN}✓ database/mysql.cnf${NC}"
fi

# 复制数据库文件
if [ -f "${SCRIPT_DIR}/database/schema.sql" ]; then
    cp "${SCRIPT_DIR}/database/schema.sql" "${PKG_DIR}/database/"
    echo -e "  ${GREEN}✓ database/schema.sql${NC}"
fi

if [ -f "${SCRIPT_DIR}/database/migration_add_token_allowances.sql" ]; then
    cp "${SCRIPT_DIR}/database/migration_add_token_allowances.sql" "${PKG_DIR}/database/"
    echo -e "  ${GREEN}✓ database/migration_add_token_allowances.sql${NC}"
fi

# 创建说明文档
echo ""
echo -e "${YELLOW}创建说明文档...${NC}"

cat > "${PKG_DIR}/README.md" << 'EOF'
# ERC20 Scanner 部署包

## 快速开始

### 1. 解压包

```bash
tar -xzf erc20-scanner-*.tgz
cd erc20-scanner-*
```

### 2. 配置环境变量（可选）

```bash
# 复制环境变量示例
cp env.example .env

# 编辑环境变量
vim .env
```

### 3. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f scanner
docker-compose logs -f rpc
```

### 4. 测试服务

```bash
# 测试 RPC API
curl http://localhost:8080/health

# 查询合约列表
curl http://localhost:8080/api/contracts?page=1&size=5
```

### 5. 停止服务

```bash
# 停止服务
docker-compose down

# 停止并删除数据（谨慎使用）
docker-compose down -v
```

## 目录结构

```
erc20-scanner-*/
├── bin/                  # 编译好的二进制文件
│   ├── scanner          # 扫描器
│   └── rpc-server       # RPC服务
├── config/              # 配置文件
│   └── config.yaml.example
├── database/            # 数据库文件
│   ├── schema.sql       # 数据库初始化脚本
│   └── mysql.cnf        # MySQL 配置
├── logs/                # 日志目录（自动创建）
├── docker-compose.yml   # Docker Compose 配置（scanner + rpc）
├── env/                 # 环境配置
│   └── docker-compose.yml   # MySQL Docker Compose 配置
├── env.example          # 环境变量示例
└── README.md            # 本文件
```

## 环境变量配置

在 `.env` 文件中配置以下变量：

```bash
# MySQL配置
MYSQL_ROOT_PASSWORD=rootpassword
MYSQL_DATABASE=token_scanner
MYSQL_USER=scanner
MYSQL_PASSWORD=scannerpass
MYSQL_PORT=3306

# 扫描器配置
NODE_URL=https://mainnet.bityuan.com/eth
START_BLOCK=42399544
END_BLOCK=-1

# RPC服务配置
RPC_PORT=8080
```

## 服务说明

### MySQL服务
- **端口**: 3306
- **数据持久化**: Docker卷 `mysql_data`
- **自动初始化**: 启动时自动执行 `database/schema.sql`

### Scanner服务
- **功能**: 扫描区块链并写入数据库
- **日志**: `logs/` 目录

### RPC服务
- **端口**: 8080
- **功能**: 提供HTTP API接口查询数据
- **健康检查**: `http://localhost:8080/health`

## API接口

### 健康检查
```bash
GET /health
```

### 查询合约详情
```bash
GET /api/contract/{address}
```

### 查询合约列表
```bash
GET /api/contracts?page=1&size=20&symbol=USDT
```

### 查询Transfer记录
```bash
GET /api/transfers/{address}?page=1&size=20
```

### 查询交易记录
```bash
GET /api/transactions/{address}?page=1&size=20
```

### 查询Holder信息
```bash
GET /api/holders/{address}?page=1&size=20
```

详细API文档请参考 `rpc/README.md`

## 故障排查

### 1. 服务无法启动

```bash
# 查看详细日志
docker-compose logs [service_name]

# 检查服务状态
docker-compose ps
```

### 2. 数据库连接失败

```bash
# 检查MySQL是否运行
docker-compose ps mysql

# 检查数据库
docker-compose exec mysql mysql -uroot -p${MYSQL_ROOT_PASSWORD} -e "SHOW DATABASES;"
```

### 3. 端口冲突

修改 `.env` 文件中的端口配置。

### 4. 权限问题

确保二进制文件有执行权限：

```bash
chmod +x bin/scanner bin/rpc-server
```

## 系统要求

- Docker 20.10+
- Docker Compose 2.0+
- Linux x86_64

## 支持

如有问题，请查看日志文件或联系技术支持。
EOF
echo -e "  ${GREEN}✓ README.md${NC}"

# 创建快速启动脚本
cat > "${PKG_DIR}/start.sh" << 'EOF'
#!/bin/bash

# ERC20 Scanner 快速启动脚本

set -e

echo "=========================================="
echo "ERC20 Scanner 启动脚本"
echo "=========================================="

# 检查docker-compose
if ! command -v docker-compose &> /dev/null && ! command -v docker &> /dev/null; then
    echo "错误: 未找到 docker-compose，请先安装"
    exit 1
fi

# 检查Docker
if ! docker info &> /dev/null; then
    echo "错误: Docker未运行，请先启动Docker"
    exit 1
fi

# 检查二进制文件
if [ ! -f "bin/scanner" ] || [ ! -f "bin/rpc-server" ]; then
    echo "错误: 未找到二进制文件"
    exit 1
fi

# 设置执行权限
chmod +x bin/scanner bin/rpc-server

# 启动服务
echo ""
echo "启动服务..."
if command -v docker-compose &> /dev/null; then
    docker-compose up -d
else
    docker compose up -d
fi

# 等待服务启动
echo ""
echo "等待服务启动..."
sleep 5

# 显示状态
echo ""
echo "服务状态:"
if command -v docker-compose &> /dev/null; then
    docker-compose ps
else
    docker compose ps
fi

echo ""
echo "=========================================="
echo "服务已启动！"
echo "=========================================="
echo ""
echo "MySQL: localhost:${MYSQL_PORT:-3306}"
echo "RPC API: http://localhost:${RPC_PORT:-8080}"
echo ""
echo "测试: curl http://localhost:${RPC_PORT:-8080}/health"
echo ""
EOF
chmod +x "${PKG_DIR}/start.sh"
echo -e "  ${GREEN}✓ start.sh${NC}"

# 创建 .gitignore（如果需要在包中保留）
cat > "${PKG_DIR}/.gitignore" << 'EOF'
logs/
*.log
.env
EOF

# 打包
echo ""
echo -e "${YELLOW}打包文件...${NC}"
cd "${BUILD_DIR}"
tar -czf "${PKG_NAME}" "${PROJECT_NAME}-${VERSION}"

# 显示结果
PKG_SIZE=$(du -h "${PKG_NAME}" | cut -f1)
echo ""
echo "=========================================="
echo -e "${GREEN}打包完成！${NC}"
echo "=========================================="
echo ""
echo "包名: ${PKG_NAME}"
echo "大小: ${PKG_SIZE}"
echo "位置: ${BUILD_DIR}/${PKG_NAME}"
echo ""
echo "包内容:"
echo "  - bin/scanner (扫描器)"
echo "  - bin/rpc-server (RPC服务)"
echo "  - docker-compose.yml (scanner + rpc 服务)"
echo "  - env/docker-compose.yml (MySQL 服务)"
echo "  - database/schema.sql (数据库初始化)"
echo "  - config/ (配置文件)"
echo "  - README.md (说明文档)"
echo ""
echo "部署步骤:"
echo "  1. 复制 ${PKG_NAME} 到测试机器"
echo "  2. 解压: tar -xzf ${PKG_NAME}"
echo "  3. 进入目录: cd ${PROJECT_NAME}-${VERSION}"
echo "  4. 启动 MySQL: docker-compose -f env/docker-compose.yml up -d"
echo "  5. 启动服务: docker-compose up -d"
echo ""

# 清理临时目录（保留打包文件）
if [ -d "${PKG_DIR}" ]; then
    rm -rf "${PKG_DIR}"
    echo -e "${GREEN}已清理临时目录${NC}"
fi

