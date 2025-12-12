package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Node     NodeConfig     `yaml:"node"`
	Scanner  ScannerConfig  `yaml:"scanner"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
}

// NodeConfig 节点配置
type NodeConfig struct {
	URL string `yaml:"url"`
}

// ScannerConfig 扫描器配置
type ScannerConfig struct {
	StartBlock int64 `yaml:"start_block"`
	EndBlock   int64 `yaml:"end_block"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Enabled bool   `yaml:"enabled"`
	DSN     string `yaml:"dsn"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
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
			URL: "https://mainnet.bityuan.com/eth",
		},
		Scanner: ScannerConfig{
			StartBlock: 42399544,
			EndBlock:   -1,
		},
		Database: DatabaseConfig{
			Enabled: false,
			DSN:     "root:password@tcp(localhost:3306)/token_scanner?charset=utf8mb4&parseTime=True&loc=Local",
		},
		Log: LogConfig{
			Level: "info",
			File:  "",
		},
	}
}

// MergeWithFlags 合并命令行参数（命令行参数优先级更高）
func (c *Config) MergeWithFlags(nodeURL *string, startBlock, endBlock *int64, enableDB *bool, dbDSN *string) {
	if nodeURL != nil && *nodeURL != "" {
		c.Node.URL = *nodeURL
	}
	if startBlock != nil {
		c.Scanner.StartBlock = *startBlock
	}
	if endBlock != nil {
		c.Scanner.EndBlock = *endBlock
	}
	if enableDB != nil {
		c.Database.Enabled = *enableDB
	}
	if dbDSN != nil && *dbDSN != "" {
		c.Database.DSN = *dbDSN
	}
}
