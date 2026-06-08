# evm-v1.4 → v1.5 升级说明

## 变更概要

v1.5 新增 **ERC20 授权额度跟踪** 功能：扫描器解析 Approval 事件，支持按地址查询所有授权记录。

| 变更 | 说明 |
|------|------|
| 新表 `token_allowances` | 存储 `(owner, spender, contract, amount)` 授权额度 |
| 扫描器增强 | 解析 Approval 事件，写入 events 表和 token_allowances 表 |
| 新接口 | `GET /evmapi/accounts/{address}/approvals` |
| 余额刷新器 | `balance_refresher` 配置项（可选，token_allowances 不需要） |

## 升级步骤

### 1. 停止服务

```bash
docker-compose down
```

### 2. 备份数据库

```bash
docker-compose exec mysql mysqldump -uroot -p${MYSQL_ROOT_PASSWORD} token_scanner > backup_v1.4_$(date +%Y%m%d).sql
```

### 3. 执行数据库升级

```bash
# 进入 MySQL 容器执行升级脚本
docker-compose exec mysql mysql -uroot -p${MYSQL_ROOT_PASSWORD} token_scanner < database/migration_add_token_allowances.sql
```

或者手动连接数据库执行 `database/migration_add_token_allowances.sql`。

### 4. 替换二进制文件

```bash
# 备份旧版本
cp bin/scanner bin/scanner.v1.4
cp bin/rpc-server bin/rpc-server.v1.4

# 替换为新版本
cp /path/to/new/bin/scanner bin/scanner
cp /path/to/new/bin/rpc-server bin/rpc-server
chmod +x bin/scanner bin/rpc-server
```

### 5. 启动服务

```bash
docker-compose up -d
```

### 6. 验证

```bash
# 检查服务状态
docker-compose ps

# 测试健康检查
curl http://localhost:8080/health

# 测试新接口（替换为实际地址）
curl "http://localhost:8080/evmapi/accounts/0x.../approvals?page=1&size=10"
```

## 注意事项

- **历史数据不回填**：升级前的 Approval 事件不会补采，token_allowances 表从升级后扫描到的第一个 Approval 事件开始有数据
- **amount=0 即删除**：取消授权（approve(spender, 0)）会自动删除 token_allowances 中的记录
- **无需额外配置**：新功能随扫描器自动启用，不需要修改配置文件
- **回滚**：如遇问题，停止服务后替换回 v1.4 二进制文件即可；token_allowances 表不会影响旧版本运行
