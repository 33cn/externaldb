package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFlagSet_NoFlags(t *testing.T) {
	// 重置 flag 包
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	info := DetectFlagSet()

	if info.NodeURL {
		t.Error("NodeURL should not be set")
	}
	if info.DBEnabled {
		t.Error("DBEnabled should not be set")
	}
	if info.StartBlock {
		t.Error("StartBlock should not be set")
	}
	if info.Port {
		t.Error("Port should not be set")
	}
}

func TestDetectFlagSet_SomeFlags(t *testing.T) {
	// 重置 flag 包
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// 定义一些 flags
	_ = flag.String("u", "", "node url")
	_ = flag.Int64("s", 0, "start block")
	_ = flag.String("port", "8080", "port")

	// 模拟命令行参数
	testArgs := []string{"-u", "https://test.com", "-s", "1000000", "-port", "9090"}
	flag.CommandLine.Parse(testArgs)

	info := DetectFlagSet()

	if !info.NodeURL {
		t.Error("NodeURL should be set")
	}
	if !info.StartBlock {
		t.Error("StartBlock should be set")
	}
	if !info.Port {
		t.Error("Port should be set")
	}
}

func TestMergeConfigWithFlagsFromStruct(t *testing.T) {
	cfg := GetDefaultConfig()
	originalGRPC := cfg.Node.GRPC

	// 创建 flags
	nodeURL := "https://flag-override.com"
	dbEnabled := true
	flags := &CommonFlags{
		NodeURL:   &nodeURL,
		DBEnabled: &dbEnabled,
	}

	// 创建 flag info
	flagInfo := &FlagSetInfo{
		NodeURL:   true,
		DBEnabled: true,
	}

	MergeConfigWithFlagsFromStruct(cfg, flags, flagInfo)

	// 验证被覆盖的字段
	if cfg.Node.URL != nodeURL {
		t.Errorf("Expected Node.URL to be '%s', got '%s'", nodeURL, cfg.Node.URL)
	}
	if cfg.Database.Enabled != dbEnabled {
		t.Errorf("Expected Database.Enabled to be %v, got %v", dbEnabled, cfg.Database.Enabled)
	}

	// 验证未设置的字段保持不变
	if cfg.Node.GRPC != originalGRPC {
		t.Errorf("Expected Node.GRPC to remain '%s', got '%s'", originalGRPC, cfg.Node.GRPC)
	}
}

func TestMergeConfigWithFlagsFromStruct_FlagNotSet(t *testing.T) {
	cfg := GetDefaultConfig()
	originalURL := cfg.Node.URL

	// 创建 flags（但 flag 未被设置）
	nodeURL := "https://flag-override.com"
	flags := &CommonFlags{
		NodeURL: &nodeURL,
	}

	// flag 未被设置
	flagInfo := &FlagSetInfo{
		NodeURL: false,
	}

	MergeConfigWithFlagsFromStruct(cfg, flags, flagInfo)

	// 验证配置未被覆盖
	if cfg.Node.URL != originalURL {
		t.Errorf("Expected Node.URL to remain '%s', got '%s'", originalURL, cfg.Node.URL)
	}
}

func TestGetFinalValue(t *testing.T) {
	tests := []struct {
		name         string
		flagValue    *string
		flagSet      bool
		configValue  string
		defaultValue string
		expected     string
	}{
		{
			name:         "flag set and not empty",
			flagValue:    stringPtr("flag-value"),
			flagSet:      true,
			configValue:  "config-value",
			defaultValue: "default-value",
			expected:     "flag-value",
		},
		{
			name:         "flag not set, use config",
			flagValue:    stringPtr("flag-value"),
			flagSet:      false,
			configValue:  "config-value",
			defaultValue: "default-value",
			expected:     "config-value",
		},
		{
			name:         "flag not set, config empty, use default",
			flagValue:    stringPtr("flag-value"),
			flagSet:      false,
			configValue:  "",
			defaultValue: "default-value",
			expected:     "default-value",
		},
		{
			name:         "flag set but empty, use config",
			flagValue:    stringPtr(""),
			flagSet:      true,
			configValue:  "config-value",
			defaultValue: "default-value",
			expected:     "config-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFinalValue(tt.flagValue, tt.flagSet, tt.configValue, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetFinalIntValue(t *testing.T) {
	tests := []struct {
		name         string
		flagValue    *int
		flagSet      bool
		configValue  int32
		defaultValue int32
		expected     int32
	}{
		{
			name:         "flag set and not zero",
			flagValue:    intPtr(100),
			flagSet:      true,
			configValue:  50,
			defaultValue: 10,
			expected:     100,
		},
		{
			name:         "flag not set, use config",
			flagValue:    intPtr(100),
			flagSet:      false,
			configValue:  50,
			defaultValue: 10,
			expected:     50,
		},
		{
			name:         "flag not set, config zero, use default",
			flagValue:    intPtr(100),
			flagSet:      false,
			configValue:  0,
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "flag set but zero, use config",
			flagValue:    intPtr(0),
			flagSet:      true,
			configValue:  50,
			defaultValue: 10,
			expected:     50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFinalIntValue(tt.flagValue, tt.flagSet, tt.configValue, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetFinalBoolValue(t *testing.T) {
	tests := []struct {
		name         string
		flagValue    *bool
		flagSet      bool
		configValue  bool
		expected     bool
	}{
		{
			name:        "flag set to true",
			flagValue:   boolPtr(true),
			flagSet:     true,
			configValue: false,
			expected:    true,
		},
		{
			name:        "flag set to false",
			flagValue:   boolPtr(false),
			flagSet:     true,
			configValue: true,
			expected:    false,
		},
		{
			name:        "flag not set, use config",
			flagValue:   boolPtr(true),
			flagSet:     false,
			configValue: false,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFinalBoolValue(tt.flagValue, tt.flagSet, tt.configValue)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLoadAndMergeForScanner_WithConfigFile(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "scanner_config.yaml")

	yamlContent := `
node:
  url: "https://config.example.com/eth"
  grpc: "config:8802"
  symbol: "config"

scanner:
  start_block: 1000000
  end_block: 2000000

database:
  enabled: true
  dsn: "config:dsn"

es:
  enabled: true
  host: "http://config-es:9200/"
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(configFile)

	// 重置 flag
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flags := DefineScannerFlags()

	// 设置 config file flag
	*flags.ConfigFile = configFile

	cfg, err := LoadAndMergeForScanner(flags)
	if err != nil {
		t.Fatalf("LoadAndMergeForScanner failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadAndMergeForScanner returned nil config")
	}

	// 验证从配置文件加载的值
	if cfg.Node.URL != "https://config.example.com/eth" {
		t.Errorf("Expected Node.URL from config, got '%s'", cfg.Node.URL)
	}
	if cfg.Scanner.StartBlock != 1000000 {
		t.Errorf("Expected Scanner.StartBlock from config, got %d", cfg.Scanner.StartBlock)
	}
}

func TestLoadAndMergeForScanner_WithFlags(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "scanner_config.yaml")

	yamlContent := `
node:
  url: "https://config.example.com/eth"

scanner:
  start_block: 1000000
  end_block: 2000000
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(configFile)

	// 重置 flag
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flags := DefineScannerFlags()

	// 设置 config file 和 flags
	*flags.ConfigFile = configFile
	*flags.NodeURL = "https://flag.example.com/eth"
	*flags.StartBlock = 5000000
	*flags.EndBlock = 6000000

	// 模拟 flag.Parse()
	flag.CommandLine.Parse([]string{
		"-c", configFile,
		"-u", "https://flag.example.com/eth",
		"-s", "5000000",
		"-e", "6000000",
	})

	cfg, err := LoadAndMergeForScanner(flags)
	if err != nil {
		t.Fatalf("LoadAndMergeForScanner failed: %v", err)
	}

	// 验证 flag 值覆盖了配置值
	if cfg.Node.URL != "https://flag.example.com/eth" {
		t.Errorf("Expected Node.URL from flag, got '%s'", cfg.Node.URL)
	}
	if cfg.Scanner.StartBlock != 5000000 {
		t.Errorf("Expected Scanner.StartBlock from flag, got %d", cfg.Scanner.StartBlock)
	}
	if cfg.Scanner.EndBlock != 6000000 {
		t.Errorf("Expected Scanner.EndBlock from flag, got %d", cfg.Scanner.EndBlock)
	}
}

func TestLoadAndMergeForRPC_WithConfigFile(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rpc_config.yaml")

	yamlContent := `
rpc:
  port: "9090"

node:
  grpc: "rpc-config:8802"
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(configFile)

	// 重置 flag
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flags := DefineRPCFlags()

	*flags.ConfigFile = configFile

	cfg, err := LoadAndMergeForRPC(flags)
	if err != nil {
		t.Fatalf("LoadAndMergeForRPC failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadAndMergeForRPC returned nil config")
	}

	// 验证从配置文件加载的值
	if cfg.RPC.Port != "9090" {
		t.Errorf("Expected RPC.Port from config, got '%s'", cfg.RPC.Port)
	}
}

func TestLoadAndMergeForRPC_WithFlags(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rpc_config.yaml")

	yamlContent := `
rpc:
  port: "9090"
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(configFile)

	// 重置 flag
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flags := DefineRPCFlags()

	*flags.ConfigFile = configFile
	*flags.Port = "8080"

	// 模拟 flag.Parse()
	flag.CommandLine.Parse([]string{
		"-c", configFile,
		"-port", "8080",
	})

	cfg, err := LoadAndMergeForRPC(flags)
	if err != nil {
		t.Fatalf("LoadAndMergeForRPC failed: %v", err)
	}

	// 验证 flag 值覆盖了配置值
	if cfg.RPC.Port != "8080" {
		t.Errorf("Expected RPC.Port from flag, got '%s'", cfg.RPC.Port)
	}
}

func TestLoadAndMergeForRPC_DefaultPort(t *testing.T) {
	// 创建临时配置文件（不设置 port）
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rpc_config.yaml")

	yamlContent := `
node:
  url: "https://test.com/eth"
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(configFile)

	// 重置 flag
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flags := DefineRPCFlags()

	*flags.ConfigFile = configFile

	cfg, err := LoadAndMergeForRPC(flags)
	if err != nil {
		t.Fatalf("LoadAndMergeForRPC failed: %v", err)
	}

	// 验证使用默认端口
	if cfg.RPC.Port != "8080" {
		t.Errorf("Expected RPC.Port to be default '8080', got '%s'", cfg.RPC.Port)
	}
}

func TestMergeConfigWithFlagsFromStruct_AllFields(t *testing.T) {
	cfg := GetDefaultConfig()

	// 创建所有 flags
	nodeURL := "https://flag-node.com"
	chainGRPC := "flag-grpc:8802"
	chainSymbol := "flag-symbol"
	dbEnabled := true
	dbDSN := "flag:dsn"
	esEnabled := true
	esHost := "http://flag-es:9200/"
	esPrefix := "flag_"
	esVersion := 6
	esUser := "flaguser"
	esPassword := "flagpass"

	flags := &CommonFlags{
		NodeURL:     &nodeURL,
		ChainGRPC:   &chainGRPC,
		ChainSymbol: &chainSymbol,
		DBEnabled:   &dbEnabled,
		DBDSN:       &dbDSN,
		ESEnabled:   &esEnabled,
		ESHost:      &esHost,
		ESPrefix:    &esPrefix,
		ESVersion:   &esVersion,
		ESUser:      &esUser,
		ESPassword:  &esPassword,
	}

	flagInfo := &FlagSetInfo{
		NodeURL:     true,
		ChainGRPC:   true,
		ChainSymbol: true,
		DBEnabled:   true,
		DBDSN:       true,
		ESEnabled:   true,
		ESHost:      true,
		ESPrefix:    true,
		ESVersion:   true,
		ESUser:      true,
		ESPassword:  true,
	}

	MergeConfigWithFlagsFromStruct(cfg, flags, flagInfo)

	// 验证所有字段都被覆盖
	if cfg.Node.URL != nodeURL {
		t.Errorf("Expected Node.URL to be '%s', got '%s'", nodeURL, cfg.Node.URL)
	}
	if cfg.Node.GRPC != chainGRPC {
		t.Errorf("Expected Node.GRPC to be '%s', got '%s'", chainGRPC, cfg.Node.GRPC)
	}
	if cfg.Node.Symbol != chainSymbol {
		t.Errorf("Expected Node.Symbol to be '%s', got '%s'", chainSymbol, cfg.Node.Symbol)
	}
	if cfg.Database.Enabled != dbEnabled {
		t.Errorf("Expected Database.Enabled to be %v, got %v", dbEnabled, cfg.Database.Enabled)
	}
	if cfg.Database.DSN != dbDSN {
		t.Errorf("Expected Database.DSN to be '%s', got '%s'", dbDSN, cfg.Database.DSN)
	}
	if cfg.ES.Enabled != esEnabled {
		t.Errorf("Expected ES.Enabled to be %v, got %v", esEnabled, cfg.ES.Enabled)
	}
	if cfg.ES.Host != esHost {
		t.Errorf("Expected ES.Host to be '%s', got '%s'", esHost, cfg.ES.Host)
	}
	if cfg.ES.Prefix != esPrefix {
		t.Errorf("Expected ES.Prefix to be '%s', got '%s'", esPrefix, cfg.ES.Prefix)
	}
	if cfg.ES.Version != int32(esVersion) {
		t.Errorf("Expected ES.Version to be %d, got %d", esVersion, cfg.ES.Version)
	}
	if cfg.ES.User != esUser {
		t.Errorf("Expected ES.User to be '%s', got '%s'", esUser, cfg.ES.User)
	}
	if cfg.ES.Password != esPassword {
		t.Errorf("Expected ES.Password to be '%s', got '%s'", esPassword, cfg.ES.Password)
	}
}

func TestLoadAndMergeForScanner_ConfigFileError(t *testing.T) {
	// 测试配置文件读取错误的情况（使用无效路径）
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flags := DefineScannerFlags()

	// 使用一个会导致读取错误的文件（目录而不是文件）
	tmpDir := t.TempDir()
	*flags.ConfigFile = tmpDir // 这是一个目录，不是文件

	cfg, err := LoadAndMergeForScanner(flags)
	// 应该返回错误，但使用默认配置
	if err != nil {
		t.Logf("Expected error for invalid config file: %v", err)
	}

	// 应该返回默认配置
	if cfg == nil {
		t.Fatal("LoadAndMergeForScanner should return default config on error")
	}

	// 验证是默认配置
	if cfg.Node.URL != "https://mainnet.bityuan.com/eth" {
		t.Errorf("Expected default Node.URL, got '%s'", cfg.Node.URL)
	}
}

func TestDetectFlagSet_AllFlags(t *testing.T) {
	// 测试所有 flag 的检测
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// 定义所有 flags
	_ = flag.String("c", "", "config file")
	_ = flag.String("u", "", "node url")
	_ = flag.Bool("db", false, "enable db")
	_ = flag.String("dsn", "", "database dsn")
	_ = flag.Bool("es", false, "enable es")
	_ = flag.String("es-host", "", "es host")
	_ = flag.String("es-prefix", "", "es prefix")
	_ = flag.Int("es-version", 0, "es version")
	_ = flag.String("es-user", "", "es user")
	_ = flag.String("es-pwd", "", "es password")
	_ = flag.String("chain-grpc", "", "chain grpc")
	_ = flag.String("chain-symbol", "", "chain symbol")
	_ = flag.Int64("s", 0, "start block")
	_ = flag.Int64("e", 0, "end block")
	_ = flag.String("port", "", "port")

	// 解析所有 flags
	testArgs := []string{
		"-c", "test.yaml",
		"-u", "https://test.com",
		"-db",
		"-dsn", "test:dsn",
		"-es",
		"-es-host", "http://test:9200/",
		"-es-prefix", "test_",
		"-es-version", "7",
		"-es-user", "user",
		"-es-pwd", "pass",
		"-chain-grpc", "grpc:8802",
		"-chain-symbol", "symbol",
		"-s", "1000",
		"-e", "2000",
		"-port", "9090",
	}
	flag.CommandLine.Parse(testArgs)

	info := DetectFlagSet()

	// 验证所有 flags 都被检测到
	if !info.ConfigFile {
		t.Error("ConfigFile should be set")
	}
	if !info.NodeURL {
		t.Error("NodeURL should be set")
	}
	if !info.DBEnabled {
		t.Error("DBEnabled should be set")
	}
	if !info.DBDSN {
		t.Error("DBDSN should be set")
	}
	if !info.ESEnabled {
		t.Error("ESEnabled should be set")
	}
	if !info.ESHost {
		t.Error("ESHost should be set")
	}
	if !info.ESPrefix {
		t.Error("ESPrefix should be set")
	}
	if !info.ESVersion {
		t.Error("ESVersion should be set")
	}
	if !info.ESUser {
		t.Error("ESUser should be set")
	}
	if !info.ESPassword {
		t.Error("ESPassword should be set")
	}
	if !info.ChainGRPC {
		t.Error("ChainGRPC should be set")
	}
	if !info.ChainSymbol {
		t.Error("ChainSymbol should be set")
	}
	if !info.StartBlock {
		t.Error("StartBlock should be set")
	}
	if !info.EndBlock {
		t.Error("EndBlock should be set")
	}
	if !info.Port {
		t.Error("Port should be set")
	}
}

func TestDetectFlagSet_AlternativeFlagNames(t *testing.T) {
	// 测试替代 flag 名称
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	_ = flag.String("es-password", "", "es password")
	_ = flag.String("chain_grpc", "", "chain grpc")
	_ = flag.String("chain_symbol", "", "chain symbol")

	testArgs := []string{
		"-es-password", "pass",
		"-chain_grpc", "grpc:8802",
		"-chain_symbol", "symbol",
	}
	flag.CommandLine.Parse(testArgs)

	info := DetectFlagSet()

	if !info.ESPassword {
		t.Error("ESPassword should be set (via es-password)")
	}
	if !info.ChainGRPC {
		t.Error("ChainGRPC should be set (via chain_grpc)")
	}
	if !info.ChainSymbol {
		t.Error("ChainSymbol should be set (via chain_symbol)")
	}
}

func TestLoadAndMergeForRPC_EmptyPortInConfig(t *testing.T) {
	// 测试配置文件中 port 为空的情况
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "rpc_config.yaml")

	yamlContent := `
rpc:
  port: ""
`

	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(configFile)

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flags := DefineRPCFlags()

	*flags.ConfigFile = configFile

	cfg, err := LoadAndMergeForRPC(flags)
	if err != nil {
		t.Fatalf("LoadAndMergeForRPC failed: %v", err)
	}

	// 应该使用默认端口
	if cfg.RPC.Port != "8080" {
		t.Errorf("Expected RPC.Port to be default '8080', got '%s'", cfg.RPC.Port)
	}
}

func TestMergeConfigWithFlagsFromStruct_EmptyStringFlags(t *testing.T) {
	// 测试空字符串 flag 的处理
	cfg := GetDefaultConfig()
	originalURL := cfg.Node.URL

	emptyURL := ""
	flags := &CommonFlags{
		NodeURL: &emptyURL,
	}

	flagInfo := &FlagSetInfo{
		NodeURL: true,
	}

	MergeConfigWithFlagsFromStruct(cfg, flags, flagInfo)

	// 空字符串不应该覆盖配置
	if cfg.Node.URL != originalURL {
		t.Errorf("Expected Node.URL to remain '%s' when flag is empty, got '%s'", originalURL, cfg.Node.URL)
	}
}

func TestMergeConfigWithFlagsFromStruct_NilFlags(t *testing.T) {
	// 测试 nil flags
	cfg := GetDefaultConfig()
	originalURL := cfg.Node.URL

	flags := &CommonFlags{
		NodeURL: nil,
	}

	flagInfo := &FlagSetInfo{
		NodeURL: true,
	}

	MergeConfigWithFlagsFromStruct(cfg, flags, flagInfo)

	// nil flag 不应该覆盖配置
	if cfg.Node.URL != originalURL {
		t.Errorf("Expected Node.URL to remain '%s' when flag is nil, got '%s'", originalURL, cfg.Node.URL)
	}
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

