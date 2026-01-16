package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/33cn/externaldb/erc20Scaner/config"
	"github.com/33cn/externaldb/escli"

	"log/slog"

	"gopkg.in/natefinch/lumberjack.v2"
)

var log *slog.Logger

func main() {
	// 定义命令行参数
	flags := config.DefineScannerFlags()
	flag.Parse()

	// 加载并合并配置
	cfg, err := config.LoadAndMergeForScanner(flags)
	if err != nil {
		// 使用 fmt 输出，因为 log 还未初始化
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// 初始化日志
	log = initLogger(cfg)

	// 打印配置信息
	logConfig(cfg, log)

	// 初始化并启动
	initAndStart(cfg)
}

func initAndStart(cfg *config.Config) {
	p := new(Process)
	p.startPoint = uint64(cfg.Scanner.StartBlock)
	p.endPoint = uint64(cfg.Scanner.EndBlock)
	p.enableDB = cfg.Database.Enabled
	p.dbDSN = cfg.Database.DSN
	p.nodeURL = cfg.Node.URL

	// 如果启用了ES模式，优先使用ES读取区块
	if cfg.ES.Enabled {
		log.Info("ES mode enabled", "host", cfg.ES.Host, "prefix", cfg.ES.Prefix)
		esClient, err := escli.NewESLongConnect(cfg.ES.Host, cfg.ES.Prefix, cfg.ES.Version, cfg.ES.User, cfg.ES.Password)
		if err != nil {
			log.Error("Failed to connect to ES", "err", err, "host", cfg.ES.Host)
			return
		}
		log.Info("ES connection established successfully")
		p.Init()
		defer p.Close()
		p.StartWithEsClient(esClient)
	} else {
		// 使用节点模式
		log.Info("Node mode enabled", "url", cfg.Node.URL)
		p.Init()
		defer p.Close()
		p.Start()
	}
}

func initLogger(cfg *config.Config) *slog.Logger {
	logLevel := slog.LevelDebug
	switch cfg.Log.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	rotateWriter := lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
	defer rotateWriter.Close()

	handler := slog.NewJSONHandler(&rotateWriter, &slog.HandlerOptions{
		Level: logLevel,
	})
	return slog.New(handler)
}

func logConfig(cfg *config.Config, log *slog.Logger) {
	log.Info("=== Configuration ===",
		"nodeURL", cfg.Node.URL,
		"startBlock", cfg.Scanner.StartBlock,
		"endBlock", cfg.Scanner.EndBlock,
		"dbEnabled", cfg.Database.Enabled,
		"esEnabled", cfg.ES.Enabled)

	if cfg.Database.Enabled {
		log.Info("Database configuration", "dsn", maskDSN(cfg.Database.DSN))
	}
	if cfg.ES.Enabled {
		log.Info("ES configuration",
			"host", cfg.ES.Host,
			"prefix", cfg.ES.Prefix,
			"version", cfg.ES.Version,
			"user", cfg.ES.User)
	}
}

// maskDSN 隐藏DSN中的密码
func maskDSN(dsn string) string {
	// 简单的密码隐藏：user:password@ -> user:***@
	parts := strings.Split(dsn, "@")
	if len(parts) > 0 {
		userPass := strings.Split(parts[0], ":")
		if len(userPass) == 2 {
			return userPass[0] + ":***@" + strings.Join(parts[1:], "@")
		}
	}
	return dsn
}
