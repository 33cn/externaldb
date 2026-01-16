# Logger 模块

统一的日志初始化模块，用于 scanner 和 rpc 程序。

## 功能

- 统一的日志初始化接口
- 支持日志级别配置（debug, info, warn, error）
- 自动日志轮转（基于 lumberjack）
- JSON 格式输出
- 自动创建日志目录
- DSN 密码隐藏工具函数

## 使用方法

### 基本使用

```go
package main

import (
    "github.com/33cn/externaldb/erc20Scaner/config"
    "github.com/33cn/externaldb/erc20Scaner/logger"
    "log/slog"
)

var log *slog.Logger

func main() {
    // 加载配置
    cfg, err := config.LoadAndMergeForScanner(flags)
    if err != nil {
        // 处理错误
        return
    }

    // 初始化日志（如果失败则退出程序）
    log, err = logger.InitLogger(cfg.Log)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
        return
    }

    // 使用日志
    log.Info("Application started", "version", "1.0.0")
    log.Debug("Debug information", "key", "value")
    log.Warn("Warning message", "key", "value")
    log.Error("Error occurred", "err", err)
}
```

### 配置

日志配置在 `config.yaml` 中：

```yaml
log:
  level: "info"              # 日志级别: debug, info, warn, error
  file: "./logs/app.log"     # 日志文件路径（可选，默认为 ./logs/app.log）
  max_size: 100              # 单个日志文件最大大小（MB），默认 100
  max_backups: 5             # 保留旧日志文件的最大数量，默认 5
  max_age: 30                # 日志文件保留天数，默认 30
```

### 日志级别

- `debug`: 调试信息，最详细
- `info`: 一般信息，默认级别
- `warn`: 警告信息
- `error`: 错误信息

### 日志轮转

日志文件会自动轮转，可通过配置文件自定义：

- **max_size**: 单个日志文件最大大小（MB），默认 100
- **max_backups**: 保留的旧日志文件数量，默认 5
- **max_age**: 日志文件保留天数，默认 30
- **Compress**: 自动压缩旧日志文件（固定为 true）

### 工具函数

#### MaskDSN

隐藏数据库连接字符串中的密码：

```go
dsn := "user:password@tcp(localhost:3306)/db"
masked := logger.MaskDSN(dsn)
// 输出: "user:***@tcp(localhost:3306)/db"
```

## 实现细节

### 日志格式

使用 JSON 格式输出，便于日志收集和分析：

```json
{
  "time": "2026-01-16T12:00:00Z",
  "level": "INFO",
  "msg": "Application started",
  "version": "1.0.0"
}
```

### 目录创建

如果日志文件指定的目录不存在，会自动创建（权限 0755）。

### 错误处理

- 如果日志目录创建失败，会回退到默认路径 `./logs/app.log`
- 如果日志级别配置无效，默认使用 `info` 级别

## 示例

### Scanner 中使用

```go
// scanner/main.go
var log *slog.Logger

func main() {
    cfg, err := config.LoadAndMergeForScanner(flags)
    if err != nil {
        return
    }
    
    log, err = logger.InitLogger(cfg.Log)
    if err != nil {
        fmt.Printf("Failed to initialize logger: %v\n", err)
        return
    }
    
    log.Info("Scanner started", 
        "startBlock", cfg.Scanner.StartBlock,
        "endBlock", cfg.Scanner.EndBlock)
}
```

### RPC 中使用

```go
// rpc/main.go
var slogger *slog.Logger

func main() {
    cfg, err := config.LoadAndMergeForRPC(flags)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    slogger, err = logger.InitLogger(cfg.Log)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    
    slogger.Info("RPC server started", "port", port)
}
```

## 依赖

- `log/slog`: Go 1.21+ 标准库
- `gopkg.in/natefinch/lumberjack.v2`: 日志轮转库

## 注意事项

1. 日志初始化应该在配置加载之后进行
2. 如果日志初始化失败，程序应该退出（使用 `log.Fatalf` 或返回错误）
3. 日志初始化成功后，不需要检查 logger 是否为 nil，可以直接使用
4. 如果需要在日志初始化之前输出错误，使用标准 `log` 包或 `fmt` 包
5. 日志文件路径使用相对路径时，相对于程序运行目录
6. 确保程序对日志目录有写权限
7. 日志轮转配置如果未设置或设置为 0，会使用默认值

