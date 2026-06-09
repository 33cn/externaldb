// scannerfix 是一个数据修复工具，从指定起始高度扫描到指定结束高度。
// 与 scanner 不同的是：
//   - -s (start block) 和 -e (end block) 是必填参数，不填退出
//   - 不读/写 scan_progress 表（方案 D），不影响线上 scanner 的断点续扫
//   - 处理进度仅通过日志输出，默认日志文件 scannerfix_app.log
//   - 到达 end block 后自动退出
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/33cn/externaldb/erc20Scaner/config"
	"github.com/33cn/externaldb/erc20Scaner/logger"
	"github.com/33cn/externaldb/erc20Scaner/scanner/engine"
	"github.com/33cn/externaldb/escli"
)

func main() {
	// 自定义 flags：-s 和 -e 必填
	configFile := flag.String("c", "config.yaml", "config file path")
	startBlock := flag.Int64("s", 0, "start block (REQUIRED)")
	endBlock := flag.Int64("e", 0, "end block (REQUIRED)")
	nodeURL := flag.String("u", "", "node url (overrides config)")
	dbEnabled := flag.Bool("db", false, "enable database (overrides config)")
	dbDSN := flag.String("dsn", "", "database DSN (overrides config)")
	esEnabled := flag.Bool("es", false, "enable ES mode (overrides config)")
	esHost := flag.String("es-host", "", "ES host (overrides config)")
	esPrefix := flag.String("es-prefix", "", "ES prefix (overrides config)")
	esVersion := flag.Int("es-version", 0, "ES version (6 or 7, overrides config)")
	esUser := flag.String("es-user", "", "ES username (overrides config)")
	esPwd := flag.String("es-pwd", "", "ES password (overrides config)")
	chainGRPC := flag.String("chain-grpc", "", "Chain33 gRPC host (overrides config)")
	chainSymbol := flag.String("chain-symbol", "", "Chain33 symbol (overrides config)")
	skipInlineBalanceUpdate := flag.Bool("skip-inline-balance", false, "skip inline balanceOf (overrides config)")
	flag.Parse()

	// 验证必填参数
	if *startBlock <= 0 {
		fmt.Fprintln(os.Stderr, "ERROR: -s (start block) is required and must be > 0")
		fmt.Fprintln(os.Stderr, "Usage: scannerfix -s <start_block> -e <end_block> [-c config.yaml] [-u node_url] [-db -dsn dsn] [-es ...]")
		os.Exit(1)
	}
	if *endBlock <= 0 {
		fmt.Fprintln(os.Stderr, "ERROR: -e (end block) is required and must be > 0")
		fmt.Fprintln(os.Stderr, "Usage: scannerfix -s <start_block> -e <end_block> [-c config.yaml] [-u node_url] [-db -dsn dsn] [-es ...]")
		os.Exit(1)
	}
	if *startBlock >= *endBlock {
		fmt.Fprintf(os.Stderr, "ERROR: -s (%d) must be less than -e (%d)\n", *startBlock, *endBlock)
		os.Exit(1)
	}

	// 加载配置文件（如果存在）
	var cfg *config.Config
	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		cfg = config.GetDefaultConfig()
	} else {
		var loadErr error
		cfg, loadErr = config.LoadConfig(*configFile)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load config file: %v, using defaults\n", loadErr)
			cfg = config.GetDefaultConfig()
		}
	}

	// 覆盖配置（命令行 > 配置文件）
	if *nodeURL != "" {
		cfg.Node.URL = *nodeURL
	}
	if *dbEnabled {
		cfg.Database.Enabled = *dbEnabled
	}
	if *dbDSN != "" {
		cfg.Database.DSN = *dbDSN
	}
	if *esEnabled {
		cfg.ES.Enabled = *esEnabled
	}
	if *esHost != "" {
		cfg.ES.Host = *esHost
	}
	if *esPrefix != "" {
		cfg.ES.Prefix = *esPrefix
	}
	if *esVersion != 0 {
		cfg.ES.Version = int32(*esVersion)
	}
	if *esUser != "" {
		cfg.ES.User = *esUser
	}
	if *esPwd != "" {
		cfg.ES.Password = *esPwd
	}
	if *chainGRPC != "" {
		cfg.Node.GRPC = *chainGRPC
	}
	if *chainSymbol != "" {
		cfg.Node.Symbol = *chainSymbol
	}
	if *skipInlineBalanceUpdate {
		cfg.Scanner.SkipInlineBalanceUpdate = *skipInlineBalanceUpdate
	}

	// 使用 scannerfix 专用日志文件名
	cfg.Log.File = "./logs/scannerfix_app.log"

	// 初始化日志
	log, err := logger.InitLogger(cfg.Log, "scannerfix")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// 设置 engine 包的 logger
	engine.SetLogger(log)

	log.Info("=== ScannerFix Started ===",
		"startBlock", *startBlock,
		"endBlock", *endBlock,
		"nodeURL", cfg.Node.URL,
		"dbEnabled", cfg.Database.Enabled,
		"esEnabled", cfg.ES.Enabled,
		"mode", "fix (no scan_progress read/write)")

	// 创建 Process 进入修复模式
	p := new(engine.Process)
	p.StartPoint = uint64(*startBlock)
	p.EndPoint = uint64(*endBlock)
	p.EnableDB = cfg.Database.Enabled
	p.DBDSN = cfg.Database.DSN
	p.NodeURL = cfg.Node.URL
	p.SkipInlineBalanceUpdate = cfg.Scanner.SkipInlineBalanceUpdate
	p.NoProgress = true // 方案 D：不读/写 scan_progress

	if cfg.ES.Enabled {
		log.Info("ES mode enabled", "host", cfg.ES.Host, "prefix", cfg.ES.Prefix)
		esClient, err := escli.NewESLongConnect(cfg.ES.Host, cfg.ES.Prefix, cfg.ES.Version, cfg.ES.User, cfg.ES.Password)
		if err != nil {
			log.Error("Failed to connect to ES", "err", err, "host", cfg.ES.Host)
			os.Exit(1)
		}
		log.Info("ES connection established successfully")
		p.Init()
		if cfg.Database.Enabled && cfg.BalanceRefresher.Enabled {
			go p.RunBalanceRefresher(context.Background(), cfg.BalanceRefresher)
		}
		defer p.Close()
		p.StartWithEsClient(esClient)
	} else {
		log.Info("Node mode enabled", "url", cfg.Node.URL)
		p.Init()
		if cfg.Database.Enabled && cfg.BalanceRefresher.Enabled {
			go p.RunBalanceRefresher(context.Background(), cfg.BalanceRefresher)
		}
		defer p.Close()
		p.Start()
	}

	log.Info("=== ScannerFix Exited ===")
}
