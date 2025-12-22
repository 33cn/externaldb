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

var (
	configFile = flag.String("c", "config.yaml", "config file path")
	rawUrl     = flag.String("u", "", "node url (overrides config)")
	startPoint = flag.Int64("s", 0, "start point (overrides config)")
	endPoint   = flag.Int64("e", 0, "end point (overrides config, use -1 for unlimited)")
	enableDB   = flag.Bool("db", false, "enable database write (overrides config)")
	dbDSN      = flag.String("dsn", "", "database DSN (overrides config)")
	// ES相关参数
	esEnabled  = flag.Bool("es", false, "enable ES mode (read blocks from ES, overrides config)")
	esHost     = flag.String("es-host", "", "ES host (overrides config)")
	esPrefix   = flag.String("es-prefix", "", "ES prefix (overrides config)")
	esVersion  = flag.Int("es-version", 0, "ES version (6 or 7, overrides config)")
	esUser     = flag.String("es-user", "", "ES username (overrides config)")
	esPassword = flag.String("es-pwd", "", "ES password (overrides config)")
	// 42241940-42241949 34562327- deploy contract
	// 42399545-42399546 token transfer
	//
)

var log *slog.Logger

func main() {
	flag.Parse()

	// 加载配置文件
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		// 先使用标准log输出，因为log15还未初始化
		fmt.Printf("Warning: Failed to load config file %s: %v, using defaults\n", *configFile, err)
		cfg = config.GetDefaultConfig()
	} else {
		fmt.Printf("Config loaded from %s\n", *configFile)
	}

	// 初始化日志系统
	logLevel := slog.LevelDebug
	if cfg.Log.Level != "" {
		if cfg.Log.Level == "debug" {
			logLevel = slog.LevelDebug
		} else if cfg.Log.Level == "info" {
			logLevel = slog.LevelInfo
		} else if cfg.Log.Level == "warn" {
			logLevel = slog.LevelWarn
		} else if cfg.Log.Level == "error" {
			logLevel = slog.LevelError
		}
	}

	rotateWriter := lumberjack.Logger{

		// 指定日志文件前缀和路径。备份文件将保留在同一目录。
		// 例如：./logs/app.log, ./logs/app-2025-12-22T08-30-00.log.gz
		Filename:   "./logs/app.log", // 日志文件的基础名，即前缀
		MaxSize:    100,              // 单位：MB。日志文件达到此大小后轮转
		MaxBackups: 5,                // 保留旧日志文件的最大数量
		MaxAge:     30,               // 单位：天。根据文件名中的时间戳删除旧文件
		Compress:   true,             // 是否压缩轮转后的旧日志文件
		// 更多可选配置...
	}
	defer rotateWriter.Close()
	handler := slog.NewJSONHandler(&rotateWriter, &slog.HandlerOptions{
		Level: logLevel,
	})

	log = slog.New(handler)
	// 记得在应用退出前关闭处理器，确保日志写入完成

	// 合并命令行参数（命令行参数优先级更高）
	// 使用flag.Visit检查参数是否被显式设置
	var nodeURL *string
	if *rawUrl != "" {
		nodeURL = rawUrl
	}

	var startBlock *int64
	var endBlock *int64
	var dbEnabled *bool
	var dbDSNStr *string
	var esEnabledFlag *bool
	var esHostFlag *string
	var esPrefixFlag *string
	var esVersionFlag *int32
	var esUserFlag *string
	var esPasswordFlag *string

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "s":
			startBlock = startPoint
		case "e":
			endBlock = endPoint
		case "db":
			dbEnabled = enableDB
		case "es":
			esEnabledFlag = esEnabled
		case "es-host":
			esHostFlag = esHost
		case "es-prefix":
			esPrefixFlag = esPrefix
		case "es-version":
			v := int32(*esVersion)
			esVersionFlag = &v
		case "es-user":
			esUserFlag = esUser
		case "es-pwd":
			esPasswordFlag = esPassword
		}
	})

	if *dbDSN != "" {
		dbDSNStr = dbDSN
	}

	cfg.MergeWithFlags(nodeURL, startBlock, endBlock, dbEnabled, dbDSNStr,
		esEnabledFlag, esHostFlag, esPrefixFlag, esVersionFlag, esUserFlag, esPasswordFlag)

	// 打印最终配置（使用结构化日志）
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

	// 初始化并启动
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
