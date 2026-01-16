package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/33cn/externaldb/erc20Scaner/config"
	"github.com/33cn/externaldb/erc20Scaner/database"
	"github.com/33cn/externaldb/erc20Scaner/logger"
)

var slogger *slog.Logger

var (
	// 全局配置变量，供 tx_parse.go 使用
	globalChainGRPC   string
	globalChainSymbol string
	globalESHost      string
	globalESPrefix    string
	globalESVersion   int
	globalESUser      string
	globalESPassword  string
)

var db *database.DB

func main() {
	// 定义命令行参数
	flags := config.DefineRPCFlags()
	flag.Parse()

	// 加载并合并配置
	cfg, err := config.LoadAndMergeForRPC(flags)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化结构化日志（如果失败则退出程序）
	slogger, err = logger.InitLogger(cfg.Log, "rpc")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 从已合并的配置中获取值（LoadAndMergeForRPC 已经处理了 flag 优先、配置文件、默认值的逻辑）
	dsn := cfg.Database.DSN
	port := cfg.RPC.Port

	// 设置全局变量供 tx_parse.go 使用
	globalChainGRPC = cfg.Node.GRPC
	globalChainSymbol = cfg.Node.Symbol
	globalESHost = cfg.ES.Host
	globalESPrefix = cfg.ES.Prefix
	globalESVersion = int(cfg.ES.Version)
	globalESUser = cfg.ES.User
	globalESPassword = cfg.ES.Password

	// 连接数据库
	db, err = database.NewDB(dsn)
	if err != nil {
		slogger.Error("Failed to connect to database", "err", err, "dsn", logger.MaskDSN(dsn))
		return
	}
	defer db.Close()

	slogger.Info("Database connected successfully")
	slogger.Info("Starting HTTP server", "port", port)

	// 注册路由（注意：handleContractAddressTransactions 和 handleContractAddressTransfers 会先检查路径，如果不是匹配的格式会调用 handleContractDetail）
	http.HandleFunc("/api/contract/", handleContractAddressTransfers)
	http.HandleFunc("/api/contracts", handleContractList)
	http.HandleFunc("/api/token/", handleTokenDetail)
	http.HandleFunc("/api/tokens", handleTokenList)
	http.HandleFunc("/api/transfers/", handleTransfers)
	http.HandleFunc("/api/transactions/", handleTransactions)
	http.HandleFunc("/api/holders/", handleHolders)
	http.HandleFunc("/api/parse_tx", handleParseTx)
	http.HandleFunc("/health", handleHealth)

	// 启动服务器
	addr := fmt.Sprintf(":%s", port)
	slogger.Info("Server listening", "addr", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		slogger.Error("Failed to start server", "err", err)
		log.Fatalf("Failed to start server: %v", err)
	}
}
