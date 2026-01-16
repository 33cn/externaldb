package config

import (
	"flag"
	"log"
)

// FlagSetInfo 记录哪些 flag 被显式设置
type FlagSetInfo struct {
	ConfigFile  bool
	NodeURL     bool
	DBEnabled   bool
	DBDSN       bool
	ESEnabled   bool
	ESHost      bool
	ESPrefix    bool
	ESVersion   bool
	ESUser      bool
	ESPassword  bool
	ChainGRPC   bool
	ChainSymbol bool
	StartBlock  bool
	EndBlock    bool
	Port        bool
}

// DetectFlagSet 检测哪些 flag 被显式设置
func DetectFlagSet() *FlagSetInfo {
	info := &FlagSetInfo{}
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "c":
			info.ConfigFile = true
		case "u":
			info.NodeURL = true
		case "db":
			info.DBEnabled = true
		case "dsn":
			info.DBDSN = true
		case "es":
			info.ESEnabled = true
		case "es-host":
			info.ESHost = true
		case "es-prefix":
			info.ESPrefix = true
		case "es-version":
			info.ESVersion = true
		case "es-user":
			info.ESUser = true
		case "es-pwd", "es-password":
			info.ESPassword = true
		case "chain-grpc", "chain_grpc":
			info.ChainGRPC = true
		case "chain-symbol", "chain_symbol":
			info.ChainSymbol = true
		case "s":
			info.StartBlock = true
		case "e":
			info.EndBlock = true
		case "port":
			info.Port = true
		}
	})
	return info
}

// MergeConfigWithFlagsFromStruct 从结构体合并配置（更灵活的方式）
// 如果 flag 被显式设置，则使用 flag 值；否则使用配置文件的值
func MergeConfigWithFlagsFromStruct(cfg *Config, flags *CommonFlags, flagInfo *FlagSetInfo) {
	if flagInfo.NodeURL && flags.NodeURL != nil && *flags.NodeURL != "" {
		cfg.Node.URL = *flags.NodeURL
	}
	if flagInfo.DBEnabled && flags.DBEnabled != nil {
		cfg.Database.Enabled = *flags.DBEnabled
	}
	if flagInfo.DBDSN && flags.DBDSN != nil && *flags.DBDSN != "" {
		cfg.Database.DSN = *flags.DBDSN
	}
	if flagInfo.ESEnabled && flags.ESEnabled != nil {
		cfg.ES.Enabled = *flags.ESEnabled
	}
	if flagInfo.ESHost && flags.ESHost != nil && *flags.ESHost != "" {
		cfg.ES.Host = *flags.ESHost
	}
	if flagInfo.ESPrefix && flags.ESPrefix != nil && *flags.ESPrefix != "" {
		cfg.ES.Prefix = *flags.ESPrefix
	}
	if flagInfo.ESVersion && flags.ESVersion != nil && *flags.ESVersion != 0 {
		cfg.ES.Version = int32(*flags.ESVersion)
	}
	if flagInfo.ESUser && flags.ESUser != nil && *flags.ESUser != "" {
		cfg.ES.User = *flags.ESUser
	}
	if flagInfo.ESPassword && flags.ESPassword != nil && *flags.ESPassword != "" {
		cfg.ES.Password = *flags.ESPassword
	}
	if flagInfo.ChainGRPC && flags.ChainGRPC != nil && *flags.ChainGRPC != "" {
		cfg.Node.GRPC = *flags.ChainGRPC
	}
	if flagInfo.ChainSymbol && flags.ChainSymbol != nil && *flags.ChainSymbol != "" {
		cfg.Node.Symbol = *flags.ChainSymbol
	}
}

// LoadAndMergeForScanner Scanner 专用的配置加载
func LoadAndMergeForScanner(flags *ScannerFlags) (*Config, error) {
	flagInfo := DetectFlagSet()
	
	configFile := "config.yaml"
	if flags.ConfigFile != nil && *flags.ConfigFile != "" {
		configFile = *flags.ConfigFile
	}
	
	cfg, err := LoadConfig(configFile)
	if err != nil {
		log.Printf("Warning: Failed to load config file: %v, using defaults", err)
		cfg = GetDefaultConfig()
	}

	// 合并通用配置
	MergeConfigWithFlagsFromStruct(cfg, &flags.CommonFlags, flagInfo)

	// 合并 Scanner 专用配置
	if flagInfo.StartBlock && flags.StartBlock != nil {
		cfg.Scanner.StartBlock = *flags.StartBlock
	}
	if flagInfo.EndBlock && flags.EndBlock != nil {
		cfg.Scanner.EndBlock = *flags.EndBlock
	}

	return cfg, nil
}

// LoadAndMergeForRPC RPC 专用的配置加载
func LoadAndMergeForRPC(flags *RPCFlags) (*Config, error) {
	flagInfo := DetectFlagSet()
	
	configFile := "config.yaml"
	if flags.ConfigFile != nil && *flags.ConfigFile != "" {
		configFile = *flags.ConfigFile
	}
	
	cfg, err := LoadConfig(configFile)
	if err != nil {
		log.Printf("Warning: Failed to load config file: %v, using defaults", err)
		cfg = GetDefaultConfig()
	}

	// 合并通用配置
	MergeConfigWithFlagsFromStruct(cfg, &flags.CommonFlags, flagInfo)

	// 合并 RPC 专用配置
	if flagInfo.Port && flags.Port != nil && *flags.Port != "" {
		cfg.RPC.Port = *flags.Port
	} else if cfg.RPC.Port == "" {
		// 如果配置文件中没有设置，使用默认值
		cfg.RPC.Port = "8080"
	}

	return cfg, nil
}

// GetFinalValue 获取最终配置值（flag 优先，否则使用配置文件）
// 这是一个通用的辅助函数，用于获取字符串类型的配置值
func GetFinalValue(flagValue *string, flagSet bool, configValue string, defaultValue string) string {
	if flagSet && flagValue != nil && *flagValue != "" {
		return *flagValue
	}
	if configValue != "" {
		return configValue
	}
	return defaultValue
}

// GetFinalIntValue 获取最终配置值（flag 优先，否则使用配置文件）
// 用于整数类型的配置值
func GetFinalIntValue(flagValue *int, flagSet bool, configValue int32, defaultValue int32) int32 {
	if flagSet && flagValue != nil && *flagValue != 0 {
		return int32(*flagValue)
	}
	if configValue != 0 {
		return configValue
	}
	return defaultValue
}

// GetFinalBoolValue 获取最终配置值（flag 优先，否则使用配置文件）
// 用于布尔类型的配置值
func GetFinalBoolValue(flagValue *bool, flagSet bool, configValue bool) bool {
	if flagSet && flagValue != nil {
		return *flagValue
	}
	return configValue
}

