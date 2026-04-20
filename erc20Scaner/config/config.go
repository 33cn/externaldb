package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	// 依赖的外部服务配置
	Node     NodeConfig     `yaml:"node"`
	Database DatabaseConfig `yaml:"database"`
	ES       ESConfig       `yaml:"es"`

	// 应用程序配置
	Scanner          ScannerConfig          `yaml:"scanner"`
	RPC              RPCConfig              `yaml:"rpc"` // RPC 服务配置
	BalanceRefresher BalanceRefresherConfig `yaml:"balance_refresher"`

	// 日志配置
	Log LogConfig `yaml:"log"`
}

// NodeConfig Chain33 节点配置
type NodeConfig struct {
	URL    string `yaml:"url"`
	GRPC   string `yaml:"grpc"`
	Symbol string `yaml:"symbol"`
}

// ScannerConfig 扫描器配置
type ScannerConfig struct {
	StartBlock int64 `yaml:"start_block"`
	EndBlock   int64 `yaml:"end_block"`
	// SkipInlineBalanceUpdate 为 true 时扫块不写链上 balanceOf，仅维护占位行并由 balance_refresher 刷新
	SkipInlineBalanceUpdate bool `yaml:"skip_inline_balance_update"`
}

// BalanceRefresherConfig 地址代币余额后台刷新（链上 balanceOf）
type BalanceRefresherConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Interval    string `yaml:"interval"`     // 如 1h、30m；空则默认 1h
	BatchSize   int    `yaml:"batch_size"`   // 每轮最多行数，<=0 则默认 50
	Concurrency int    `yaml:"concurrency"`  // 并发 RPC 上限，<=0 则默认 5
	MinAge      string `yaml:"min_age"`      // 只刷新超过该时间未更新的行，如 30s；空表示不限制
}

// ParseBalanceRefresherDurations 解析 Interval 与 MinAge，供扫描器使用
func (b BalanceRefresherConfig) ParseBalanceRefresherDurations() (interval, minAge time.Duration, err error) {
	intervalStr := b.Interval
	if intervalStr == "" {
		intervalStr = "1h"
	}
	interval, err = time.ParseDuration(intervalStr)
	if err != nil {
		return 0, 0, fmt.Errorf("balance_refresher.interval: %w", err)
	}
	if b.MinAge == "" {
		return interval, 0, nil
	}
	minAge, err = time.ParseDuration(b.MinAge)
	if err != nil {
		return 0, 0, fmt.Errorf("balance_refresher.min_age: %w", err)
	}
	return interval, minAge, nil
}

// RPCConfig RPC 服务配置
type RPCConfig struct {
	Port string `yaml:"port"` // HTTP 服务端口
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Enabled bool   `yaml:"enabled"`
	DSN     string `yaml:"dsn"`
}

// ESConfig ES配置
type ESConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Prefix   string `yaml:"prefix"`
	Version  int32  `yaml:"version"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`       // 日志级别: debug, info, warn, error
	File       string `yaml:"file"`        // 日志文件路径
	MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小（MB），默认 100
	MaxBackups int    `yaml:"max_backups"` // 保留旧日志文件的最大数量，默认 5
	MaxAge     int    `yaml:"max_age"`     // 日志文件保留天数，默认 30
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return GetDefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		Node: NodeConfig{
			URL:    "https://mainnet.bityuan.com/eth",
			GRPC:   "localhost:8802",
			Symbol: "bty",
		},
		Scanner: ScannerConfig{
			StartBlock:              42399544,
			EndBlock:                -1,
			SkipInlineBalanceUpdate: false,
		},
		BalanceRefresher: BalanceRefresherConfig{
			Enabled:     false,
			Interval:    "1h",
			BatchSize:   50,
			Concurrency: 5,
			MinAge:      "",
		},
		RPC: RPCConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			Enabled: false,
			DSN:     "root:password@tcp(localhost:3306)/token_scanner?charset=utf8mb4&parseTime=True&loc=Local",
		},
		ES: ESConfig{
			Enabled:  false,
			Host:     "http://localhost:9200/",
			Prefix:   "seq01_",
			Version:  7,
			User:     "",
			Password: "",
		},
		Log: LogConfig{
			Level:      "info",
			File:       "./logs/app.log",
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     30,
		},
	}
}
