package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/33cn/externaldb/erc20Scaner/config"
)

var (
	configFile = flag.String("c", "config.yaml", "config file path")
	rawUrl     = flag.String("u", "", "node url (overrides config)")
	startPoint = flag.Int64("s", 0, "start point (overrides config)")
	endPoint   = flag.Int64("e", 0, "end point (overrides config, use -1 for unlimited)")
	enableDB   = flag.Bool("db", false, "enable database write (overrides config)")
	dbDSN      = flag.String("dsn", "", "database DSN (overrides config)")
	// 42241940-42241949 34562327- deploy contract
	// 42399545-42399546 token transfer
	//
)

func main() {
	flag.Parse()

	// 加载配置文件
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Printf("Warning: Failed to load config file %s: %v, using defaults", *configFile, err)
		cfg = config.GetDefaultConfig()
	} else {
		log.Printf("Config loaded from %s", *configFile)
	}

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

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "s":
			startBlock = startPoint
		case "e":
			endBlock = endPoint
		case "db":
			dbEnabled = enableDB
		}
	})

	if *dbDSN != "" {
		dbDSNStr = dbDSN
	}

	cfg.MergeWithFlags(nodeURL, startBlock, endBlock, dbEnabled, dbDSNStr)

	// 打印最终配置
	fmt.Println("=== Configuration ===")
	fmt.Printf("Node URL: %s\n", cfg.Node.URL)
	fmt.Printf("Start Block: %d\n", cfg.Scanner.StartBlock)
	if cfg.Scanner.EndBlock > 0 {
		fmt.Printf("End Block: %d\n", cfg.Scanner.EndBlock)
	} else {
		fmt.Printf("End Block: unlimited\n")
	}
	fmt.Printf("Database Enabled: %v\n", cfg.Database.Enabled)
	if cfg.Database.Enabled {
		fmt.Printf("Database DSN: %s\n", maskDSN(cfg.Database.DSN))
	}
	fmt.Println("===================")

	// 初始化并启动
	p := new(Process)
	p.startPoint = uint64(cfg.Scanner.StartBlock)
	p.endPoint = uint64(cfg.Scanner.EndBlock)
	p.enableDB = cfg.Database.Enabled
	p.dbDSN = cfg.Database.DSN
	p.nodeURL = cfg.Node.URL
	p.Init()
	defer p.Close()
	p.Start()
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
